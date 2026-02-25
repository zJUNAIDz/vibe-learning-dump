# Sockets and File Descriptors — How Programs Talk to the Network

> Every network connection your program makes — every HTTP request, every database query, every DNS lookup — goes through a socket. A socket is just a file descriptor. Understanding sockets means understanding how the Linux kernel exposes networking to userspace programs. This is the foundation for everything: web servers, containers, load balancers.

---

## Table of Contents

1. [Everything Is a File](#everything-is-a-file)
2. [What Is a Socket?](#what-is-a-socket)
3. [File Descriptors](#file-descriptors)
4. [Socket Types](#socket-types)
5. [The Socket API — System Calls](#socket-api)
6. [TCP Socket Lifecycle](#tcp-lifecycle)
7. [UDP Sockets](#udp-sockets)
8. [Socket Addresses and Binding](#binding)
9. [Blocking vs Non-Blocking](#blocking)
10. [Socket Options](#socket-options)
11. [Unix Domain Sockets](#unix-sockets)
12. [Observing Sockets on a Live System](#observing)
13. [Key Takeaways](#key-takeaways)

---

## Everything Is a File

Linux's design philosophy: **everything is a file descriptor (fd)**.

```
Regular file     →  fd → read()/write()
Directory        →  fd → getdents()
Terminal         →  fd → read()/write()
Pipe             →  fd → read()/write()
Network socket   →  fd → read()/write()/send()/recv()
Timer            →  fd → timerfd_create()
Signal           →  fd → signalfd()
Event            →  fd → eventfd()
```

### Why this matters

```
Because sockets are file descriptors:
  1. You can use read()/write() on network connections (not just send/recv)
  2. You can use select()/poll()/epoll() to monitor sockets AND files together
  3. You can pass sockets between processes (via Unix domain sockets)
  4. You can redirect network I/O to files and vice versa
  5. Standard tools (strace, lsof) work on sockets

This unification is NOT just academic — it's why Linux is excellent at networking.
```

### File descriptor table

```
Every process has a file descriptor table:

Process (PID 1234)
┌─────────┬──────────────────────────────┐
│ FD      │ Points to                     │
├─────────┼──────────────────────────────┤
│ 0       │ stdin  (terminal /dev/pts/0)  │
│ 1       │ stdout (terminal /dev/pts/0)  │
│ 2       │ stderr (terminal /dev/pts/0)  │
│ 3       │ /var/log/app.log (file)       │
│ 4       │ TCP socket → 10.0.0.1:5432   │  ← PostgreSQL connection
│ 5       │ TCP socket → 10.0.0.2:6379   │  ← Redis connection
│ 6       │ TCP listening socket :8080    │  ← HTTP server
│ 7       │ UDP socket → DNS resolver     │  ← DNS query
│ 8       │ Unix socket → /var/run/app.sock│  ← IPC
└─────────┴──────────────────────────────┘

System-wide limit: cat /proc/sys/fs/file-max     (e.g., 9223372036854775807)
Per-process limit: ulimit -n                      (e.g., 1024 default!)
```

---

## What Is a Socket?

A socket is an **endpoint for communication**. It's the interface between application code and the kernel's networking stack.

```
                Application
                    │
                    │ socket API (system calls)
                    ▼
            ┌───────────────┐
            │    Socket     │  ← Kernel data structure
            │               │
            │  Protocol     │  TCP / UDP / RAW
            │  Local addr   │  IP:port
            │  Remote addr  │  IP:port (for connected sockets)
            │  State        │  LISTEN / ESTABLISHED / CLOSE_WAIT / ...
            │  Buffers      │  Send buffer, receive buffer
            │  Options      │  SO_REUSEADDR, TCP_NODELAY, ...
            └───────┬───────┘
                    │
                    ▼
            Kernel networking stack
            (IP → routing → device driver → NIC)
```

### Socket = 5-tuple

A connection is identified by five values:

```
{protocol, local_ip, local_port, remote_ip, remote_port}

Example:
{TCP, 10.0.0.5, 54321, 93.184.216.34, 443}

This 5-tuple is unique across the system.
Multiple connections to the same server use different local ports.
```

---

## File Descriptors

### FD lifecycle

```c
// 1. Create a socket → returns file descriptor
int fd = socket(AF_INET, SOCK_STREAM, 0);  // fd = 3 (next available)

// 2. Use the socket
connect(fd, ...);
write(fd, "GET / HTTP/1.1\r\n...", len);
read(fd, buffer, sizeof(buffer));

// 3. Close the socket → frees the fd number for reuse
close(fd);
```

### FD limits — a critical production issue

```bash
# Default per-process limit
ulimit -n
# 1024  ← WAY TOO LOW for servers!

# A web server handling 10,000 connections needs 10,000+ fds
# Each connection = 1 socket fd
# Plus: log files, database connections, inotify watches, etc.

# Increase soft limit
ulimit -n 65536

# Permanent: /etc/security/limits.conf
*    soft    nofile    65536
*    hard    nofile    131072

# Systemd services: /etc/systemd/system/myapp.service
[Service]
LimitNOFILE=65536

# System-wide: /etc/sysctl.conf
fs.file-max = 2097152
```

### Viewing file descriptors

```bash
# See all fds for a process
ls -la /proc/$(pidof nginx)/fd/
# lrwx------ 1 root root 64 ... 0 -> /dev/null
# lrwx------ 1 root root 64 ... 1 -> /var/log/nginx/access.log
# lrwx------ 1 root root 64 ... 6 -> socket:[12345]
# lrwx------ 1 root root 64 ... 7 -> socket:[12346]

# Count open fds
ls /proc/$(pidof nginx)/fd/ | wc -l

# lsof shows socket details
lsof -p $(pidof nginx)
# COMMAND  PID USER   FD   TYPE   DEVICE SIZE/OFF NODE NAME
# nginx   1234 root    6u  IPv4   12345  0t0      TCP *:80 (LISTEN)
# nginx   1234 root    7u  IPv4   12346  0t0      TCP 10.0.0.1:80->10.0.0.2:54321 (ESTABLISHED)
```

---

## Socket Types

### SOCK_STREAM (TCP)

```
  - Connection-oriented
  - Reliable, ordered byte stream
  - Flow control and congestion control
  - Used by: HTTP, HTTPS, SSH, SMTP, PostgreSQL, MySQL

  Analogy: Phone call
    → You dial (connect), talk (send/recv), hang up (close)
    → Guaranteed delivery, in order
```

### SOCK_DGRAM (UDP)

```
  - Connectionless
  - Unreliable datagrams (may be lost, reordered, duplicated)
  - No flow/congestion control
  - Low overhead, low latency
  - Used by: DNS, DHCP, NTP, video streaming, gaming, QUIC

  Analogy: Postal mail
    → You send letters (datagrams). Some might get lost.
    → No confirmation of delivery.
```

### SOCK_RAW

```
  - Raw access to IP layer
  - Application constructs headers itself
  - Requires root (CAP_NET_RAW)
  - Used by: ping (ICMP), traceroute, packet crafting tools (scapy)
```

### SOCK_SEQPACKET

```
  - Like STREAM but preserves message boundaries
  - Reliable, ordered, but each send() = one message
  - Rarely used (SCTP, some Unix domain sockets)
```

### Address families

```
AF_INET    → IPv4
AF_INET6   → IPv6
AF_UNIX    → Unix domain sockets (local IPC, no network)
AF_PACKET  → Raw Ethernet frames (like tcpdump uses)
AF_NETLINK → Kernel ↔ userspace communication
```

---

## The Socket API — System Calls

Every network program, regardless of language, ultimately calls these kernel system calls:

### Core system calls

```
socket()      Create a socket → returns fd
bind()        Assign local address (IP:port) to socket
listen()      Mark socket as passive (server: ready to accept connections)
accept()      Wait for and accept incoming connection → returns NEW fd
connect()     Initiate connection to remote (client)
send()/write() Send data
recv()/read()  Receive data
close()       Close socket
```

### Server flow (TCP)

```
Server:                              Client:
────────                             ────────
fd = socket()                        fd = socket()
bind(fd, port 80)                         │
listen(fd, backlog=128)                   │
         │                                │
         │        ←── SYN ───────── connect(fd, server:80)
         │        ─── SYN-ACK ──→         │
         │        ←── ACK ──────          │
new_fd = accept(fd)                       │
         │                                │
    recv(new_fd, ...) ←── data ─── send(fd, "GET /")
    send(new_fd, ...) ─── data ──→ recv(fd, ...)
         │                                │
    close(new_fd)                    close(fd)
```

### Key: accept() returns a NEW fd

```
Listening socket (fd=6): stays open, accepts more connections
Connection socket (fd=7): specific client connection

Server with 3 clients:
  fd 6 → listening on :80  (never gets data — only accepts)
  fd 7 → client 1 (10.0.0.2:54321)
  fd 8 → client 2 (10.0.0.3:54322)
  fd 9 → client 3 (10.0.0.4:54323)
```

### The backlog parameter

```c
listen(fd, 128);  // backlog = 128
```

```
The backlog controls TWO queues:

1. SYN queue (half-open connections): SYN received, SYN-ACK sent, waiting for ACK
2. Accept queue (completed connections): 3-way handshake done, waiting for accept()

If accept queue is full → kernel drops/rejects new connections
  Symptom: "connection refused" or SYN retransmissions

Check queue sizes:
  ss -ltn
  State  Recv-Q  Send-Q  Local Address:Port
  LISTEN    0      128    0.0.0.0:80
           │       │
           │       └── backlog (max pending connections)
           └── currently pending connections

If Recv-Q approaches Send-Q → application isn't calling accept() fast enough!
```

### Checking with strace

```bash
# Watch a program's socket system calls
strace -f -e trace=network curl https://example.com 2>&1 | head -30
# socket(AF_INET6, SOCK_STREAM, IPPROTO_TCP) = 5
# connect(5, {sa_family=AF_INET, sin_port=htons(443), sin_addr=inet_addr("93.184.216.34")}, 16) = -1
# ... (TLS happens via read/write)
# sendto(5, "GET / HTTP/1.1\r\nHost: example.com\r\n...", 78, MSG_NOSIGNAL, NULL, 0) = 78
# recvfrom(5, "HTTP/1.1 200 OK\r\n...", 16384, 0, NULL, NULL) = 1256
```

---

## TCP Socket Lifecycle

### Complete state diagram (socket perspective)

```
                    socket()
                       │
              ┌────────┴────────┐
              │                 │
           bind()            connect()
           listen()             │
              │              SYN_SENT
              │                 │
           LISTEN          SYN-ACK rcvd
              │                 │
         accept()          ESTABLISHED
              │                 │
         ESTABLISHED        send()/recv()
              │                 │
         send()/recv()     close()
              │                 │
           close()          FIN_WAIT_1
              │                 │
          FIN_WAIT_1        FIN_WAIT_2 or CLOSING
          or CLOSE_WAIT        │
              │             TIME_WAIT
              │                 │
           CLOSED            CLOSED
```

### TIME_WAIT — the misunderstood state

```
After active closer sends final ACK:
  Socket enters TIME_WAIT for 2×MSL (usually 60 seconds)

Purpose:
  1. Ensure final ACK was received (retransmit if needed)
  2. Prevent old delayed packets from being confused with new connection

Problem:
  High-traffic server closing many connections → thousands of TIME_WAIT sockets
  Each TIME_WAIT holds a 5-tuple → can't reuse that exact combination

Check:
  ss -s
  # TCP:   15230 (estab 412, closed 8, orphaned 3, timewait 14812)
  #                                                 ^^^^^^^^^^
  # 14,812 TIME_WAIT sockets!

Solutions:
  1. SO_REUSEADDR: Allow binding to port even with TIME_WAIT sockets
  2. TCP_NODELAY: Not directly related but reduces small-packet overhead
  3. Connection pooling: Reuse connections instead of creating new ones
  4. net.ipv4.tcp_tw_reuse = 1: Allow reusing TIME_WAIT for outbound connections
```

---

## UDP Sockets

### Simpler lifecycle

```
UDP has no connection state:

  fd = socket(AF_INET, SOCK_DGRAM, 0)
  bind(fd, :53)                     ← Optional (server typically binds)
  
  sendto(fd, data, dst_addr)        ← Send datagram to any address
  recvfrom(fd, buf, &src_addr)      ← Receive datagram, learn sender

  close(fd)

No connect, no accept, no listen.
(You CAN call connect() on UDP — it just sets a default destination.)
```

### Connected UDP sockets

```c
// Unconnected: must specify destination each time
sendto(fd, data, len, 0, &dest_addr, sizeof(dest_addr));

// Connected: set default destination
connect(fd, &dest_addr, sizeof(dest_addr));
send(fd, data, len, 0);  // no need to specify dest

// Benefits of connected UDP:
//   1. Kernel can deliver ICMP errors (port unreachable) to the socket
//   2. Slightly faster (kernel caches route lookup)
//   3. Only receives from connected peer (some filtering)
```

---

## Socket Addresses and Binding

### Binding to addresses

```
bind() assigns a local address to a socket:

bind(fd, {0.0.0.0, 8080})    → Listen on ALL interfaces, port 8080
bind(fd, {127.0.0.1, 8080})  → Listen on loopback ONLY (localhost)
bind(fd, {10.0.0.5, 8080})   → Listen on specific interface only
bind(fd, {::, 8080})         → Listen on ALL IPv6 (and IPv4 if dual-stack)

Check bindings:
  ss -tlnp
  LISTEN  0  128  0.0.0.0:8080  *:*  users:(("nginx",pid=1234,fd=6))
  LISTEN  0  128  127.0.0.1:5432 *:*  users:(("postgres",pid=5678,fd=3))
```

### Port allocation

```
Well-known ports:    0-1023    (require root or CAP_NET_BIND_SERVICE)
Registered ports:    1024-49151
Ephemeral ports:     49152-65535 (kernel assigns for outgoing connections)

Check ephemeral range:
  cat /proc/sys/net/ipv4/ip_local_port_range
  # 32768   60999

Client connects → kernel picks next available ephemeral port
If all ephemeral ports used → "Cannot assign requested address" error
```

### SO_REUSEADDR and SO_REUSEPORT

```
SO_REUSEADDR:
  Without it: bind() fails if port has TIME_WAIT sockets
  With it:    bind() succeeds even with TIME_WAIT sockets
  ALWAYS set this for servers. Not setting it = restart fails for 60 seconds.

SO_REUSEPORT:
  Allows MULTIPLE sockets to bind to the SAME port
  Kernel load-balances incoming connections across them
  Used by: nginx (multiple worker processes), high-performance servers
  
  Without SO_REUSEPORT:
    1 listening socket → accept() → distribute to workers via IPC
    (thundering herd problem)
  
  With SO_REUSEPORT:
    N listening sockets (one per worker) → kernel distributes
    Better cache locality, no thundering herd
```

---

## Blocking vs Non-Blocking

### Blocking I/O (default)

```
By default, socket calls BLOCK:

  recv(fd, buf, len, 0);
  // Thread sleeps here until data arrives (could be seconds/minutes/forever)

  accept(fd, ...);
  // Thread sleeps here until a client connects

  connect(fd, ...);
  // Thread sleeps until connection established (or timeout)

Problem: One thread per connection
  10,000 connections = 10,000 threads = massive memory + context switching
  This is the C10K problem (handling 10K concurrent connections)
```

### Non-blocking I/O

```c
// Make socket non-blocking
fcntl(fd, F_SETFL, O_NONBLOCK);

// Now recv() returns immediately
ssize_t n = recv(fd, buf, len, 0);
if (n == -1 && errno == EAGAIN) {
    // No data available right now — try again later
}
// No blocking! But now you need to know WHEN to try again → use epoll
```

### The evolution of I/O multiplexing

```
Problem: How to monitor thousands of fds efficiently?

select() [1983]:
  - Pass a bitmask of fds to monitor
  - Kernel scans ALL fds every time
  - O(n) per call, limited to 1024 fds (FD_SETSIZE)
  - Old, slow, avoid

poll() [1986]:
  - Pass array of fd structs
  - No fd limit
  - Still O(n) — kernel scans all fds every call
  - Better than select but still slow at scale

epoll() [2002, Linux-specific]:
  - Register fds once with epoll instance
  - epoll_wait() returns ONLY the ready fds
  - O(1) for each ready fd (not O(n) for all fds)
  - Scales to millions of connections
  - Used by nginx, Node.js, Go runtime, Redis
```

---

## Socket Options

### Setting options

```c
int opt = 1;
setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));
```

### Important socket options

```
SO_REUSEADDR:
  Allow binding to port with existing TIME_WAIT sockets.
  Essential for servers that restart.

SO_REUSEPORT:
  Allow multiple sockets on the same port.
  Kernel load-balances connections.

SO_KEEPALIVE:
  Send TCP keepalive probes on idle connections.
  Detects dead peers.
  Default interval: 2 hours (too long; tune with TCP_KEEPIDLE).

TCP_NODELAY:
  Disable Nagle's algorithm (which buffers small writes).
  Reduces latency for interactive/real-time protocols.
  ALWAYS enable for latency-sensitive apps (gRPC, game servers, SSH).

SO_RCVBUF / SO_SNDBUF:
  Set receive/send buffer sizes.
  Larger buffers = more throughput on high-latency links.
  Kernel auto-tunes these (usually fine).

SO_LINGER:
  Control behavior on close().
  Default: close() returns immediately, kernel sends FIN.
  With linger: close() blocks until data sent or timeout.

TCP_QUICKACK:
  Disable delayed ACKs. Send ACK immediately.
  Useful for request-response protocols.
```

### Viewing socket options

```bash
# See buffer sizes for all TCP sockets
ss -tnmi
# ... cubic wscale:7,7 rto:204 rtt:1.5/0.5 ato:40 mss:1448
#     cwnd:10 bytes_sent:1234 bytes_acked:1234 bytes_received:5678
#     send 77.1Mbps rcv_space:14600 rcv_ssthresh:64076

# See kernel TCP tuning parameters
sysctl net.ipv4.tcp_rmem  # min default max receive buffer
sysctl net.ipv4.tcp_wmem  # min default max send buffer
# net.ipv4.tcp_rmem = 4096  131072  6291456
```

---

## Unix Domain Sockets

### What they are

```
Unix domain sockets: IPC (inter-process communication) on the SAME machine.
  - No network involved — data never hits a NIC
  - Uses filesystem path instead of IP:port
  - Faster than TCP loopback (no TCP overhead, no routing)

Used by:
  - Docker daemon (/var/run/docker.sock)
  - PostgreSQL local connections (/var/run/postgresql/.s.PGSQL.5432)
  - X11 display server
  - systemd journal
  - Nginx → PHP-FPM / uWSGI communication
```

### Types

```
SOCK_STREAM:  Like TCP — reliable byte stream (most common)
SOCK_DGRAM:   Like UDP — datagrams (but reliable on Unix sockets!)
SOCK_SEQPACKET: Like STREAM but preserves message boundaries
```

### Why faster than TCP localhost

```
TCP loopback (127.0.0.1):
  Application → socket → TCP state machine → IP routing →
  loopback device → IP routing → TCP state machine → socket → Application
  
  Full TCP overhead: checksums, window management, Nagle, delayed ACKs

Unix domain socket:
  Application → socket → kernel buffer → socket → Application
  
  Skip: TCP, IP, routing, checksums
  Result: ~2× throughput, ~30% lower latency
```

### Observing Unix sockets

```bash
# List Unix domain sockets
ss -xlp
# Netid State  Recv-Q Send-Q  Local Address:Port  Peer Address:Port
# u_str LISTEN  0      128     /var/run/docker.sock 12345  * 0
#                               users:(("dockerd",pid=1234,fd=3))

# With lsof
lsof -U
# COMMAND  PID USER  FD   TYPE   DEVICE SIZE/OFF NODE NAME
# dockerd 1234 root   3u  unix 0x... 0t0  12345  /var/run/docker.sock type=STREAM
```

---

## Observing Sockets on a Live System

### ss — the modern socket tool

```bash
# All TCP sockets
ss -tnap
# State   Recv-Q  Send-Q   Local Address:Port   Peer Address:Port  Process
# LISTEN  0       128      0.0.0.0:80           0.0.0.0:*          users:(("nginx",...))
# ESTAB   0       0        10.0.0.1:80          10.0.0.2:54321     users:(("nginx",...))

# Flags:
#   -t = TCP    -u = UDP    -x = Unix
#   -l = listening only     -a = all (including listening)
#   -n = numeric (no DNS)   -p = show process
#   -m = memory usage       -i = TCP internal info

# Filter by state
ss -tn state established
ss -tn state time-wait
ss -tn state close-wait    # ← Leak detection!

# Filter by port
ss -tn 'sport == :80'
ss -tn 'dport == :443'

# Count connections by state
ss -tn | awk '{print $1}' | sort | uniq -c | sort -rn
```

### lsof — sockets as files

```bash
# All network connections for a process
lsof -i -P -n -p $(pidof nginx)

# All connections on port 80
lsof -i :80

# All connections to a specific IP
lsof -i @93.184.216.34
```

### /proc filesystem

```bash
# Raw socket info
cat /proc/net/tcp      # TCP sockets (hex encoded)
cat /proc/net/udp      # UDP sockets
cat /proc/net/unix     # Unix domain sockets

# Per-process fds
ls -la /proc/$(pidof nginx)/fd/
# Each symlink → socket:[inode] or file

# Socket detail
cat /proc/$(pidof nginx)/net/tcp
```

### Common debugging scenarios

```
Scenario: "Address already in use"
  → Another process is using the port
  → ss -tlnp | grep :8080
  → Kill the process or use SO_REUSEADDR

Scenario: Too many CLOSE_WAIT sockets
  → Application is receiving FIN but not calling close()
  → Memory/fd leak
  → ss -tn state close-wait | wc -l
  → Fix the application code

Scenario: Too many TIME_WAIT sockets
  → Normal for high-traffic servers
  → Enable tcp_tw_reuse if needed
  → Use connection pooling

Scenario: "Too many open files"
  → FD limit reached
  → ulimit -n (check limit)
  → ls /proc/PID/fd/ | wc -l (check usage)
  → Increase limit or fix fd leaks
```

---

## Key Takeaways

1. **Sockets are file descriptors** — Linux's "everything is a file" philosophy means sockets use the same API as files
2. **A connection = 5-tuple** — {protocol, src_ip, src_port, dst_ip, dst_port} uniquely identifies every connection
3. **accept() returns a NEW fd** — the listening socket stays open; each client gets its own fd
4. **FD limits matter** — default 1024 is too low for servers. Always increase `ulimit -n` in production
5. **Blocking I/O can't scale** — one thread per connection fails at thousands of connections. Use epoll for high concurrency
6. **SO_REUSEADDR is mandatory** for servers — without it, restart fails for 60 seconds due to TIME_WAIT
7. **Unix domain sockets** bypass TCP/IP entirely — use them for same-machine communication (Docker, databases)
8. **ss is your best friend** — `ss -tnap` shows every socket, state, and owning process
9. **TIME_WAIT is normal**, CLOSE_WAIT is a bug — too many CLOSE_WAIT = application not closing sockets
10. **strace -e trace=network** reveals exactly which system calls a program makes

---

## Next

→ [Kernel Networking Stack](./02-kernel-networking-stack.md) — How packets flow through the Linux kernel from NIC to application
