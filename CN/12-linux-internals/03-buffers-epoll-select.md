# Buffers, epoll, and select — I/O Multiplexing and Event-Driven Networking

> Modern web servers handle tens of thousands of connections on a single machine. They don't create a thread per connection — that can't scale. Instead, they use **I/O multiplexing**: a single thread monitors many sockets simultaneously, handling whichever ones are ready. This is the foundation of nginx, Node.js, Redis, Go's runtime, and every high-performance network application. Understanding how it works is essential.

---

## Table of Contents

1. [The C10K Problem](#c10k)
2. [Socket Buffers](#socket-buffers)
3. [Blocking I/O and Its Limits](#blocking)
4. [select()](#select)
5. [poll()](#poll)
6. [epoll()](#epoll)
7. [Edge-Triggered vs Level-Triggered](#et-vs-lt)
8. [Event-Driven Architecture](#event-driven)
9. [io_uring — The Future](#io-uring)
10. [Buffer Tuning](#tuning)
11. [Debugging Buffer Issues](#debugging)
12. [Key Takeaways](#key-takeaways)

---

## The C10K Problem

### History

```
Year 2000: How do you handle 10,000 concurrent connections?

Thread-per-connection model:
  10,000 connections = 10,000 threads
  Each thread: ~1 MB stack = 10 GB RAM just for stacks
  Context switching: 10,000 threads = constant thrashing
  Result: server falls over

The solutions:
  1. I/O multiplexing (select → poll → epoll)
  2. Event-driven architectures (single thread, many connections)
  3. Non-blocking I/O

Today: C10M (10 million connections) is the frontier.
epoll handles this with ease on modern hardware.
```

### Scaling models

```
1. Thread-per-connection:
   ┌──────────┐   ┌──────────┐   ┌──────────┐
   │ Thread 1 │   │ Thread 2 │   │ Thread 3 │   ... Thread N
   │ Client 1 │   │ Client 2 │   │ Client 3 │
   └──────────┘   └──────────┘   └──────────┘
   Simple but doesn't scale past ~1000 connections

2. Event-driven (epoll):
   ┌───────────────────────────────────────┐
   │              Event Loop               │
   │                                       │
   │  epoll_wait() → [fd4, fd7, fd102]     │
   │  handle(fd4)  → read, process, write  │
   │  handle(fd7)  → read, process, write  │
   │  handle(fd102)→ accept new connection │
   │                                       │
   │  Monitors 50,000 fds on 1 thread      │
   └───────────────────────────────────────┘
   Scales to millions of connections

3. Hybrid (thread pool + epoll):
   Event loop accepts and reads, dispatches CPU-heavy work to thread pool
   Used by: Go runtime, Rust Tokio, Java NIO
```

---

## Socket Buffers

### Every socket has two buffers

```
                    Kernel Space
┌──────────────────────────────────────────────┐
│                                              │
│    Send Buffer              Receive Buffer   │
│  ┌──────────────┐        ┌──────────────┐    │
│  │ ████████░░░░ │        │ ██████░░░░░░ │    │
│  │ data waiting │        │ data waiting │    │
│  │ to be sent   │        │ to be read   │    │
│  └──────┬───────┘        └──────┬───────┘    │
│         │                       │            │
│         ▼                       ▲            │
│    TCP sends when              TCP places    │
│    window allows               received data │
│                                              │
└──────────────────────────────────────────────┘
         │                       │
    Application                Application
    write(fd, data)            read(fd, buf)
```

### Buffer sizes

```bash
# TCP buffer sizes (autotuning)
cat /proc/sys/net/ipv4/tcp_rmem
# 4096  131072  6291456
#  min  default   max
# Kernel auto-adjusts between min and max based on memory pressure

cat /proc/sys/net/ipv4/tcp_wmem
# 4096  16384  4194304

# Per-socket buffer sizes (viewable with ss)
ss -tnmi
# ESTAB  0  0  10.0.0.1:80  10.0.0.2:54321
#   skmem:(r0,rb131072,t0,tb87040,f0,w0,o0,bl0,d0)
#   r=0: bytes in receive buffer
#   rb=131072: receive buffer size limit
#   t=0: bytes in send buffer
#   tb=87040: send buffer size limit

# Global maximum (setsockopt can't exceed this)
cat /proc/sys/net/core/rmem_max   # max receive buffer
cat /proc/sys/net/core/wmem_max   # max send buffer
```

### What happens when buffers fill

```
Receive buffer full:
  → TCP: Receiver advertises window = 0 (zero window)
  → Sender stops sending (flow control)
  → Application must call recv() to drain buffer
  → If application is too slow → connection stalls

Send buffer full:
  → write()/send() BLOCKS (blocking mode)
  → write()/send() returns EAGAIN (non-blocking mode)
  → Application must wait until space available

UDP receive buffer full:
  → Kernel DROPS incoming datagrams!
  → No flow control in UDP
  → Check: cat /proc/net/snmp | grep Udp
  →   Udp: ... InErrors=1234  ← UDP drops
```

---

## Blocking I/O and Its Limits

### How blocking I/O works

```c
// Blocking read — thread sleeps until data arrives
char buf[4096];
ssize_t n = read(fd, buf, sizeof(buf));
// Thread is suspended here. Could be microseconds or minutes.
// During this time, thread cannot do ANYTHING else.

// Blocking accept — thread sleeps until client connects
int client_fd = accept(listen_fd, NULL, NULL);
// Thread waits here for next connection.
```

### Why it doesn't scale

```
Server with 10,000 clients using blocking I/O:

Thread 1: read(fd_1) ← sleeping, waiting for client 1
Thread 2: read(fd_2) ← sleeping, waiting for client 2
Thread 3: read(fd_3) ← sleeping, waiting for client 3
...
Thread 10000: read(fd_10000) ← sleeping

Problems:
  - 10,000 threads × 8 MB stack = 80 GB virtual memory
  - Context switch overhead: kernel schedules 10,000 threads
  - Most threads idle most of the time (waste of resources)
  - Cache thrashing: each context switch pollutes L1/L2 cache

Solution: ONE thread monitoring ALL sockets
  → I/O multiplexing (select, poll, epoll)
```

---

## select()

### How it works

```c
fd_set read_fds;
FD_ZERO(&read_fds);
FD_SET(fd1, &read_fds);   // monitor fd1
FD_SET(fd2, &read_fds);   // monitor fd2
FD_SET(fd3, &read_fds);   // monitor fd3

int max_fd = fd3;  // must pass the highest fd number

// Block until at least one fd is ready
int ready = select(max_fd + 1, &read_fds, NULL, NULL, NULL);

// Check which fds are ready
if (FD_ISSET(fd1, &read_fds)) { /* fd1 has data */ }
if (FD_ISSET(fd2, &read_fds)) { /* fd2 has data */ }
if (FD_ISSET(fd3, &read_fds)) { /* fd3 has data */ }
```

### Why select() is slow

```
Problem 1: fd_set is a fixed-size bitmask (FD_SETSIZE = 1024)
  → Maximum 1024 file descriptors. Period.
  → Can't monitor fd > 1023

Problem 2: O(n) kernel scan
  → Kernel scans ALL FD_SETSIZE bits every call (not just the ones you set)
  → Even if you have 3 fds, kernel scans 1024 bits

Problem 3: fd_set is modified in place
  → Must rebuild fd_set before every call
  → Copying overhead

Problem 4: Copy overhead
  → Entire fd_set copied from userspace → kernel → userspace each call

Verdict: Usable for < 100 fds. Avoid for anything serious.
```

---

## poll()

### How it works

```c
struct pollfd fds[3];
fds[0] = (struct pollfd){.fd = fd1, .events = POLLIN};
fds[1] = (struct pollfd){.fd = fd2, .events = POLLIN};
fds[2] = (struct pollfd){.fd = fd3, .events = POLLIN};

int ready = poll(fds, 3, -1);  // -1 = block forever

for (int i = 0; i < 3; i++) {
    if (fds[i].revents & POLLIN) {
        // fds[i].fd has data ready
    }
}
```

### Better than select, but still O(n)

```
Improvements over select:
  ✓ No fd limit (array, not bitmask)
  ✓ Events and return events are separate fields (no rebuild needed)
  ✓ Cleaner API

Still slow:
  ✗ O(n) kernel scan — kernel checks ALL fds in array every call
  ✗ Entire array copied user↔kernel each call

  With 50,000 fds:
    - poll() copies 50,000 × 8 bytes = 400 KB per call
    - Kernel checks all 50,000 fds
    - Even if only 3 are ready
    
Verdict: Fine for < 1000 fds. Too slow for 10K+.
```

---

## epoll()

### The solution — O(1) per ready fd

```
Key insight: Registration is separate from waiting.
  - Register fds ONCE (epoll_ctl)
  - Wait returns ONLY ready fds (epoll_wait)

                    ┌──────────────────────────┐
                    │     epoll instance        │
                    │                           │
  epoll_ctl(ADD) ──→│  Internal data structure  │
  epoll_ctl(DEL) ──→│  (red-black tree +        │
  epoll_ctl(MOD) ──→│   ready list)             │
                    │                           │
                    │  ┌─ Ready list ──────────┐│
                    │  │ fd7: EPOLLIN          ││
  epoll_wait() ←────│  │ fd102: EPOLLIN        ││
  Returns ONLY ←────│  │ fd4: EPOLLOUT         ││
  ready fds         │  └──────────────────────┘│
                    └──────────────────────────┘
```

### API

```c
// 1. Create epoll instance
int epfd = epoll_create1(0);

// 2. Register fds (done ONCE per fd)
struct epoll_event ev;
ev.events = EPOLLIN;      // interested in read readiness
ev.data.fd = listen_fd;
epoll_ctl(epfd, EPOLL_CTL_ADD, listen_fd, &ev);

// Also add client fds as they connect...
ev.events = EPOLLIN;
ev.data.fd = client_fd;
epoll_ctl(epfd, EPOLL_CTL_ADD, client_fd, &ev);

// 3. Event loop
struct epoll_event events[MAX_EVENTS];
while (1) {
    int n = epoll_wait(epfd, events, MAX_EVENTS, -1);
    // n = number of READY fds (not total fds!)
    
    for (int i = 0; i < n; i++) {
        if (events[i].data.fd == listen_fd) {
            // New connection → accept()
            int new_fd = accept(listen_fd, NULL, NULL);
            // Add new_fd to epoll...
        } else {
            // Client data ready → read()
            read(events[i].data.fd, buf, sizeof(buf));
            // Process and respond...
        }
    }
}
```

### Why epoll is fast

```
1. Register once, wait many times
   epoll_ctl(ADD) = one-time cost
   No need to pass fd list on each wait

2. Kernel maintains ready list via callbacks
   When data arrives on a socket, kernel callback adds fd to ready list
   No scanning needed

3. epoll_wait returns ONLY ready fds
   50,000 fds monitored, 3 ready → returns 3 events
   select/poll would scan all 50,000

4. No user↔kernel copy of fd list
   Kernel maintains the interest set in kernel memory
   Only ready events are copied out

Performance:
  50,000 fds, 100 ready:
    select:  ~10 ms (scan 50K fds)
    poll:    ~10 ms (scan 50K fds)
    epoll:   ~0.01 ms (return 100 ready fds)
```

### epoll scalability

```
              Time per call (microseconds)
              ┌──────────────────────────────────────────┐
              │                                          │
     10,000   │  select ──────────────/                  │
              │                     /                    │
              │  poll ─────────────/                     │
      1,000   │                   /                      │
              │                  /                       │
              │                /                         │
        100   │              /                           │
              │            /                             │
              │          /                               │
         10   │        /                                 │
              │      /                                   │
              │    /                                     │
          1   │  epoll ──────────────────────────────    │
              │                                          │
              └──────────────────────────────────────────┘
              10     100    1K    10K   50K   100K
                     Number of file descriptors
```

---

## Edge-Triggered vs Level-Triggered

### Two notification modes

```
Level-Triggered (LT) — DEFAULT:
  epoll_wait() returns if fd IS ready
  Like a full inbox: keeps notifying you until you empty it
  
  If you read SOME data but not all:
    → Next epoll_wait() will return this fd AGAIN
    → Safe: you won't miss data
    → But: may cause extra wakeups

Edge-Triggered (ET):
  epoll_wait() returns only when fd BECOMES ready (state change)
  Like a doorbell: rings once, not again until new mail
  
  If you read SOME data but not all:
    → epoll_wait() will NOT return this fd again
    → Until NEW data arrives
    → MUST read ALL available data in a loop until EAGAIN
    → If you don't: data sits in buffer, never processed → bug!

  ev.events = EPOLLIN | EPOLLET;  // Edge-triggered
```

### Edge-triggered gotcha

```c
// WRONG with ET:
if (events[i].events & EPOLLIN) {
    read(fd, buf, sizeof(buf));  // read ONCE
    // If there was more data, it's stuck → never notified again!
}

// CORRECT with ET:
if (events[i].events & EPOLLIN) {
    while (1) {
        ssize_t n = read(fd, buf, sizeof(buf));
        if (n == -1) {
            if (errno == EAGAIN) break;  // all data consumed
            // handle real error
        }
        if (n == 0) {
            // Connection closed
            close(fd);
            break;
        }
        process(buf, n);
    }
}
```

### When to use which

```
Level-triggered (LT):
  - Simpler to use
  - Harder to have bugs
  - Slightly more epoll_wait() calls
  - Good default choice

Edge-triggered (ET):
  - Fewer epoll_wait() calls
  - Must drain fd completely (read until EAGAIN)
  - Easy to introduce bugs
  - Used by: nginx (performance-critical)

Most applications: Use LT. Switch to ET only if benchmarking shows benefit.
```

---

## Event-Driven Architecture

### The event loop pattern

```
                    ┌──────────────────────────────────┐
                    │           Event Loop              │
                    │                                   │
                    │  while (true) {                   │
                    │    events = epoll_wait(epfd)      │
                    │    for event in events {          │
                    │      if (event == new_connection) │
                    │        accept() + register        │
                    │      if (event == data_ready)     │
                    │        read() + process + write() │
                    │      if (event == write_ready)    │
                    │        flush pending writes       │
                    │      if (event == close)          │
                    │        cleanup + deregister       │
                    │    }                              │
                    │  }                                │
                    └──────────────────────────────────┘

CRITICAL RULE: Never block inside the event loop!
  - No sleep(), no synchronous file I/O, no CPU-heavy computation
  - If you block → ALL connections stall
  - Offload blocking work to thread pool
```

### Real-world implementations

```
nginx:
  - Master process + N worker processes
  - Each worker: single-threaded event loop with epoll
  - SO_REUSEPORT: kernel distributes connections across workers
  - Can handle 100K+ concurrent connections

Node.js:
  - Single-threaded event loop (libuv, uses epoll on Linux)
  - All I/O is non-blocking
  - CPU-heavy work → worker_threads or external service
  - "Don't block the event loop" is the #1 Node.js rule

Redis:
  - Single-threaded event loop (ae library, uses epoll)
  - All operations are in-memory → fast → no blocking
  - I/O threads added in Redis 6 for parsing/writing

Go:
  - Runtime uses epoll internally (netpoller)
  - Goroutines appear blocking but kernel threads use epoll underneath
  - Runtime multiplexes goroutines onto OS threads
  - Developer writes simple blocking code, runtime handles non-blocking

Rust (Tokio):
  - Async runtime using epoll/io_uring
  - Work-stealing thread pool
  - async/await syntax for non-blocking I/O
```

---

## io_uring — The Future

### What it is

```
io_uring (Linux 5.1+) is the next evolution in async I/O:

Traditional I/O:
  1. Application makes system call (read/write/send/recv)
  2. Context switch to kernel
  3. Kernel does I/O
  4. Context switch back to application
  → Each I/O = 2 context switches minimum

io_uring:
  1. Application writes request to submission queue (shared memory)
  2. Kernel picks up request (no system call!)
  3. Kernel writes result to completion queue (shared memory)
  4. Application reads result (no system call!)
  → Zero system calls for I/O!

Performance gain:
  - Eliminates system call overhead
  - Batching: submit multiple I/O requests at once
  - Zero-copy possible
  - 2-10× throughput improvement over epoll in benchmarks
```

### io_uring architecture

```
  ┌──────────────────────────────────────────┐
  │              Application                  │
  │                                           │
  │  Submission Queue (SQ)  Completion Queue  │
  │  ┌──┬──┬──┬──┐        ┌──┬──┬──┬──┐     │
  │  │R1│R2│R3│  │        │C1│C2│  │  │     │
  │  └──┴──┴──┴──┘        └──┴──┴──┴──┘     │
  │      │   shared memory    ▲              │
  ├──────┼───────────────────┼──────────────┤
  │      ▼   Kernel           │              │
  │  Process requests → Do I/O → Post results│
  └──────────────────────────────────────────┘
  
  No system calls for submission or completion!
  (except io_uring_enter() when SQ needs a kick)
```

---

## Buffer Tuning

### TCP buffer sizing

```bash
# TCP autotuning (enabled by default)
cat /proc/sys/net/ipv4/tcp_moderate_rcvbuf
# 1  (autotuning enabled)

# Buffer ranges
cat /proc/sys/net/ipv4/tcp_rmem
# 4096  131072  6291456   (min=4K, default=128K, max=6M)

# For high-bandwidth, high-latency links:
# BDP = Bandwidth × Delay (Bandwidth-Delay Product)
# Example: 1 Gbps link with 50ms RTT
# BDP = 1,000,000,000 bits/sec × 0.050 sec = 50,000,000 bits = 6.25 MB
# Buffer should be >= BDP for full utilization

# Increase for high-BDP links:
sysctl -w net.ipv4.tcp_rmem="4096 131072 16777216"   # max 16 MB
sysctl -w net.ipv4.tcp_wmem="4096 131072 16777216"
sysctl -w net.core.rmem_max=16777216
sysctl -w net.core.wmem_max=16777216
```

### UDP buffer sizing

```bash
# UDP has no flow control — if buffer fills, packets drop!
# Default UDP receive buffer is often too small for bursty traffic

# Check default
sysctl net.core.rmem_default    # 212992 (208 KB)
sysctl net.core.rmem_max        # 212992

# Increase for UDP-heavy applications (DNS, metrics, logging)
sysctl -w net.core.rmem_max=26214400        # 25 MB
sysctl -w net.core.rmem_default=26214400

# Application can also set per-socket:
# setsockopt(fd, SOL_SOCKET, SO_RCVBUF, &size, sizeof(size));
```

---

## Debugging Buffer Issues

### Socket buffer status

```bash
# See buffer usage for all TCP sockets
ss -tnm
# ESTAB  0  36  10.0.0.1:80  10.0.0.2:54321
#        │   │
#        │   └── Send-Q: bytes in send buffer (waiting to be sent)
#        └── Recv-Q: bytes in recv buffer (waiting to be read by application)

# If Recv-Q is large and growing:
#   → Application is reading too slowly
#   → Will eventually cause zero window (TCP flow control)

# If Send-Q is large and growing:
#   → Network or peer is slow
#   → Window limited or congestion

# See buffer sizes per socket
ss -tnmi
#   skmem:(r0,rb131072,t0,tb87040,f0,w0,o0,bl0,d0)
#   r: receive buffer used | rb: receive buffer limit
#   t: transmit buffer used | tb: transmit buffer limit
```

### Detecting drops

```bash
# UDP drops (receive buffer overflow)
cat /proc/net/snmp | grep Udp
# Udp: InDatagrams NoPorts InErrors OutDatagrams ...
#   InErrors = number of UDP datagrams dropped

# Watch UDP drops over time
watch -d "cat /proc/net/snmp | grep Udp"

# TCP listen queue overflow
nstat -az | grep Listen
# TcpExtListenOverflows = accept queue overflowed
# TcpExtListenDrops = connections dropped due to listen overflow

# Check specific listen socket
ss -tlnp | grep :80
# LISTEN  0  128  0.0.0.0:80  *:*
#         │   │
#         │   └── Send-Q = backlog (max queue size)
#         └── Recv-Q = current queue size (connections waiting for accept)
# If Recv-Q approaches Send-Q → application too slow at accept()ing
```

### epoll debugging

```bash
# Count epoll fds per process
ls /proc/$(pidof nginx)/fd/ | wc -l

# See what epoll is monitoring  
cat /proc/$(pidof nginx)/fdinfo/5
# pos: 0
# flags: 02
# mnt_id: 15
# tfd: 6 events: 2019 data: 600000006  ← fd 6 monitored for events 0x2019

# strace to see epoll calls
strace -e trace=epoll_wait,epoll_ctl -p $(pidof nginx) 2>&1 | head
# epoll_wait(8, [{EPOLLIN, {u32=6, u64=6}}], 512, -1) = 1
# epoll_wait returned 1 ready fd (fd 6 is readable)
```

---

## Key Takeaways

1. **Thread-per-connection doesn't scale** — 10,000 threads = too much memory + context switching overhead
2. **I/O multiplexing** monitors many sockets with few threads: select (old) → poll (better) → epoll (modern)
3. **epoll is O(1) per ready fd** — register once, wait returns only ready fds. select/poll are O(n) per ALL fds
4. **Socket buffers control flow** — receive buffer full → TCP zero window → sender stops. UDP has no flow control → drops
5. **Level-triggered = safe default**, edge-triggered = faster but must drain completely (read until EAGAIN)
6. **Never block in an event loop** — one blocking call stalls ALL connections. Offload to thread pool
7. **BDP determines buffer size** — for high-latency links, buffer must be ≥ bandwidth × delay
8. **UDP drops are silent** — check `/proc/net/snmp` InErrors. Increase `rmem_max` for UDP-heavy workloads
9. **io_uring eliminates syscall overhead** — shared-memory submission/completion queues, 2-10× over epoll
10. **Debug with**: `ss -tnm` (buffer usage), `nstat` (listen overflows), `strace -e epoll_wait` (event loop behavior)

---

## Next Module

→ [Module 13: Networking Tools](../13-networking-tools/01-ip-and-ss.md) — Master the essential Linux networking commands
