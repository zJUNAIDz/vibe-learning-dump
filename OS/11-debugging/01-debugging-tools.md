# Performance & Debugging Tools

**Essential Tools for Understanding What's Happening**

🟡 **Intermediate** | 🔴 **Advanced**

---

## Introduction

When something goes wrong in production, you need tools to **observe** what the OS and processes are actually doing.

This module covers the essential debugging tools every developer should know.

---

## The Debugging Mindset

**Don't guess. Observe.**

```
Problem: "App is slow"

❌ Bad approach:
   "Maybe we need more RAM?"
   "Let's add a cache?"
   "Should we upgrade the server?"

✅ Good approach:
   1. Measure current resource usage
   2. Identify the bottleneck (CPU, memory, disk, network)
   3. Understand why that resource is constrained
   4. Fix the root cause
```

---

## Process Inspection

### `ps` — Process Status

**Basic usage:**

```bash
# All processes, full details
$ ps aux

# Specific process
$ ps -p 1234

# Process tree
$ ps auxf  # or: pstree

# Threads
$ ps -eLf
```

**Useful columns:**

```bash
$ ps aux
USER  PID  %CPU %MEM    VSZ   RSS TTY STAT START   TIME COMMAND
root    1   0.0  0.1 168944 12345 ?   Ss   10:00   0:03 /lib/systemd/systemd
```

| Column | Meaning |
|--------|---------|
| `PID` | Process ID |
| `%CPU` | CPU usage |
| `%MEM` | Memory usage (% of total RAM) |
| `VSZ` | Virtual memory size (KB) |
| `RSS` | Resident set size (physical RAM, KB) |
| `STAT` | Process state (R=running, S=sleeping, D=disk wait, Z=zombie) |
| `TIME` | Total CPU time |

**Process states:**

```
R  Running or runnable
S  Sleeping (waiting for event)
D  Uninterruptible sleep (usually disk I/O)
Z  Zombie (terminated, waiting for parent)
T  Stopped (by signal)
I  Idle kernel thread
```

**Find processes:**

```bash
# By name
$ ps aux | grep nginx

# Better: use pgrep
$ pgrep nginx
1234
1235

# With details
$ pgrep -a nginx
1234 nginx: master process /usr/sbin/nginx
1235 nginx: worker process
```

### `top` / `htop` — Live Process Monitor

**`top`:**

```bash
$ top

# Sorted by memory
$ top -o %MEM

# Show specific user
$ top -u zjunaidz

# Batch mode (for logging)
$ top -b -n 1 > snapshot.txt
```

**Interactive commands in `top`:**

```
P  Sort by CPU
M  Sort by memory
k  Kill process
r  Renice (change priority)
1  Show individual CPU cores
h  Help
q  Quit
```

**`htop` (better UI):**

```bash
$ htop

# Navigate with arrow keys
# F5: Tree view
# F6: Sort by column
# F9: Kill process
```

---

## CPU Profiling

### `perf` — Performance Profiler

**Record CPU profile:**

```bash
# Record for 10 seconds
$ sudo perf record -F 99 -a -g -- sleep 10

# Record specific process
$ sudo perf record -F 99 -p 1234 -g -- sleep 10
```

**View results:**

```bash
$ sudo perf report
```

**Flame graph (visual CPU profile):**

```bash
$ git clone https://github.com/brendangregg/FlameGraph
$ sudo perf record -F 99 -a -g -- sleep 30
$ sudo perf script | ./FlameGraph/stackcollapse-perf.pl | ./FlameGraph/flamegraph.pl > flame.svg
$ firefox flame.svg
```

### Application-Specific Profilers

**Node.js:**

```bash
# Built-in profiler
$ node --prof server.js
# Run some load
# Ctrl+C
$ node --prof-process isolate-*-v8.log > profile.txt

# Chrome DevTools
$ node --inspect server.js
# Open chrome://inspect in Chrome
# Click "inspect" → Profiler tab
```

**Go:**

```go
import (
    "net/http"
    _ "net/http/pprof"
)

func main() {
    go func() {
        http.ListenAndServe("localhost:6060", nil)
    }()
    
    // Your application...
}
```

```bash
# CPU profile
$ curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
$ go tool pprof cpu.prof

# Heap profile
$ curl http://localhost:6060/debug/pprof/heap > heap.prof
$ go tool pprof heap.prof
```

**Python:**

```bash
# cProfile
$ python -m cProfile -o profile.stats app.py
$ python -m pstats profile.stats

# py-spy (sampling profiler, no code changes)
$ pip install py-spy
$ sudo py-spy record -o profile.svg --pid 1234
```

---

## Memory Analysis

### Memory Usage Overview

```bash
$ free -h
              total        used        free      shared  buff/cache   available
Mem:           15Gi       3.2Gi       8.1Gi       200Mi       4.0Gi        11Gi
Swap:         8.0Gi          0B       8.0Gi
```

**Columns:**
- `total`: Total RAM
- `used`: Used by processes
- `free`: Completely unused
- `buff/cache`: Used for disk cache (can be freed if needed)
- `available`: Actually available for new programs

**Key insight:** `available` is what matters, not `free`.

### Per-Process Memory

```bash
# RSS (Resident Set Size) = actual RAM used
$ ps aux --sort=-%mem | head

# Detailed memory map
$ pmap -x 1234

# Or from /proc
$ cat /proc/1234/status | grep -i vm
VmSize:     500000 kB  # Virtual memory
VmRSS:      250000 kB  # Physical RAM
VmData:     100000 kB  # Heap
VmStk:         132 kB  # Stack
```

### Memory Leaks

**Node.js:**

```bash
# Take heap snapshots
$ kill -USR2 <pid>
# Saves to heapsnapshot-*.heapsnapshot

# Analyze in Chrome DevTools:
# Open chrome://inspect → Memory → Load heapsnapshot file
```

**Go:**

```bash
# Heap profile
$ curl http://localhost:6060/debug/pprof/heap > heap1.prof
# Wait a bit
$ curl http://localhost:6060/debug/pprof/heap > heap2.prof

# Compare
$ go tool pprof -base=heap1.prof heap2.prof
(pprof) top
(pprof) list functionName
```

---

## File Descriptor Inspection

### `lsof` — List Open Files

**All open files:**

```bash
$ sudo lsof | wc -l
123456  # Total open files on system
```

**By process:**

```bash
$ lsof -p 1234

# Count
$ lsof -p 1234 | wc -l
```

**By type:**

```bash
# Regular files
$ lsof -p 1234 | grep REG

# Sockets
$ lsof -p 1234 | grep -i tcp
$ lsof -p 1234 | grep -i udp

# Pipes
$ lsof -p 1234 | grep FIFO
```

**Network connections:**

```bash
# All network connections
$ sudo lsof -i

# Specific port
$ sudo lsof -i :8080

# By protocol
$ sudo lsof -i tcp
$ sudo lsof -i udp

# Established connections
$ sudo lsof -i -sTCP:ESTABLISHED
```

**Find what's using a file:**

```bash
$ lsof /var/log/syslog
COMMAND  PID     USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
rsyslogd 567     syslog  7w   REG  8,1    123456  789  /var/log/syslog
```

### File Descriptor Limits

```bash
# System-wide limit
$ cat /proc/sys/fs/file-max
1000000

# Per-process limit
$ ulimit -n
1024

# Current usage
$ cat /proc/sys/fs/file-nr
1234    0       1000000
# open   free    max
```

**Increase limit:**

```bash
# For current shell
$ ulimit -n 65536

# Permanently (/etc/security/limits.conf)
* soft nofile 65536
* hard nofile 65536

# For systemd service
[Service]
LimitNOFILE=65536
```

---

## Network Debugging

### `ss` — Socket Statistics (modern `netstat`)

**All listening sockets:**

```bash
$ ss -tlnp
State   Recv-Q Send-Q Local Address:Port  Peer Address:Port
LISTEN  0      128    0.0.0.0:22          0.0.0.0:*       users:(("sshd",pid=1234,fd=3))
LISTEN  0      128    0.0.0.0:80          0.0.0.0:*       users:(("nginx",pid=2345,fd=6))
```

**Options:**
- `-t`: TCP
- `-u`: UDP
- `-l`: Listening
- `-n`: Numeric (don't resolve hostnames)
- `-p`: Show process

**Established connections:**

```bash
$ ss -tnp
State   Recv-Q Send-Q Local Address:Port   Peer Address:Port
ESTAB   0      0      192.168.1.10:45678   93.184.216.34:443   users:(("firefox",pid=5678,fd=42))
```

**Summary:**

```bash
$ ss -s
Total: 500 (kernel 512)
TCP:   200 (estab 150, closed 40, orphaned 0, synrecv 0, timewait 40/0)
```

### `tcpdump` — Packet Capture

**Capture on interface:**

```bash
# All traffic
$ sudo tcpdump -i eth0

# Specific host
$ sudo tcpdump -i eth0 host example.com

# Specific port
$ sudo tcpdump -i eth0 port 80

# Save to file
$ sudo tcpdump -i eth0 -w capture.pcap
# Analyze with Wireshark
```

**HTTP traffic (readable):**

```bash
$ sudo tcpdump -i eth0 -A -s 0 port 80
```

**Common filters:**

```bash
# TCP SYN packets (connection attempts)
$ sudo tcpdump -i eth0 'tcp[tcpflags] & tcp-syn != 0'

# DNS queries
$ sudo tcpdump -i eth0 port 53

# Traffic to/from specific IP
$ sudo tcpdump -i eth0 host 192.168.1.10
```

### `ping` / `traceroute`

**Check connectivity:**

```bash
$ ping example.com
PING example.com (93.184.216.34) 56(84) bytes of data.
64 bytes from 93.184.216.34: icmp_seq=1 ttl=56 time=10.2 ms
```

**Trace route:**

```bash
$ traceroute example.com
 1  router.local (192.168.1.1)  1.234 ms
 2  isp-gateway (10.0.0.1)  5.678 ms
 3  ...
```

**Better traceroute:**

```bash
$ mtr example.com  # Continuous traceroute with statistics
```

---

## Disk I/O Analysis

### `iostat` — I/O Statistics

```bash
$ iostat -x 1
Device  rrqm/s wrqm/s  r/s   w/s  rkB/s wkB/s await r_await w_await util
sda       0.00   5.00  0.00 10.00   0.00 40.00  2.50    0.00    2.50 10.0%
```

**Key metrics:**
- `r/s`, `w/s`: Reads/writes per second
- `rkB/s`, `wkB/s`: KB read/written per second
- `await`: Average wait time (ms)
- `util`: Percentage utilization (100% = saturated)

### `iotop` — Top for Disk I/O

```bash
$ sudo iotop -o  # Only show processes doing I/O

Total DISK READ: 10.00 M/s | Total DISK WRITE: 50.00 M/s
  TID  PRIO  USER   DISK READ  DISK WRITE  COMMAND
 1234  be/4  root       0.00 B   48.00 M/s  postgres: writer
 2345  be/4  user      10.00 M/s   0.00 B   python backup.py
```

### `df` / `du` — Disk Usage

**Filesystem usage:**

```bash
$ df -h
Filesystem      Size  Used Avail Use% Mounted on
/dev/sda1       100G   45G   50G  48% /
/dev/sdb1       1.0T  800G  200G  80% /data
```

**Directory usage:**

```bash
# Current directory
$ du -sh .

# Top-level directories
$ du -sh /*

# Find largest directories
$ du -h /var | sort -h | tail -10
```

---

## System Calls Tracing

### `strace` — Trace System Calls

**Trace a command:**

```bash
$ strace ls
execve("/usr/bin/ls", ["ls"], ...) = 0
openat(AT_FDCWD, "/etc/ld.so.cache", O_RDONLY|O_CLOEXEC) = 3
read(3, "...", 832) = 832
close(3) = 0
...
```

**Attach to running process:**

```bash
$ sudo strace -p 1234

# Summary of syscalls
$ sudo strace -c -p 1234
% time     seconds  usecs/call     calls    errors syscall
------ ----------- ----------- --------- --------- ----------------
 45.67    0.000234          10        23           read
 34.12    0.000175          12        15           write
 20.21    0.000104           8        13           epoll_wait
```

**Filter specific syscalls:**

```bash
# Only file operations
$ strace -e trace=open,read,write,close ls

# Only network
$ strace -e trace=socket,connect,send,recv curl example.com

# Timing
$ strace -T ls  # Show time per syscall
```

**Common use cases:**

```bash
# Why is this slow?
$ strace -T ./slow-program

# What files does it access?
$ strace -e trace=open,openat,stat ./program

# What's it waiting on?
$ strace -p 1234
# If you see: futex(...) = <unfinished>
# Process is blocked on a lock
```

### `ltrace` — Library Call Trace

```bash
$ ltrace ls
__libc_start_main(0x4024d0, 1, 0x7ffc..., 0x4112e0 <no return ...>
...
```

---

## Logging and Journals

### `journalctl` — systemd Logs

**Recent logs:**

```bash
$ journalctl -n 100        # Last 100 lines
$ journalctl -f            # Follow (like tail -f)
$ journalctl --since today
$ journalctl --since "1 hour ago"
```

**By service:**

```bash
$ journalctl -u nginx
$ journalctl -u nginx --since today
```

**By priority:**

```bash
$ journalctl -p err        # Errors only
$ journalctl -p warning    # Warnings and above
```

**Kernel messages:**

```bash
$ journalctl -k            # Same as dmesg
```

**Boot logs:**

```bash
$ journalctl -b            # This boot
$ journalctl -b -1         # Previous boot
$ journalctl --list-boots  # All boots
```

### `dmesg` — Kernel Ring Buffer

```bash
$ dmesg | tail

# With human-readable timestamps
$ dmesg -T

# Follow
$ dmesg -w

# Errors only
$ dmesg --level=err
```

---

## Performance Monitoring

### Load Average

```bash
$ uptime
 12:34:56 up 5 days, 3:21,  1 user,  load average: 1.23, 0.98, 0.76
#                                                   1min  5min  15min
```

**Interpretation:**

```
For 4-core system:
Load < 4.0:  Good (CPUs not fully utilized)
Load ≈ 4.0:  Optimal (all CPUs busy)
Load > 4.0:  Overloaded (processes waiting for CPU)

Rule of thumb: Load average should be < number of CPU cores
```

### System-Wide Metrics

```bash
# CPU, memory, disk, network in one
$ vmstat 1
procs -----------memory---------- ---swap-- -----io---- -system-- ------cpu-----
 r  b   swpd   free   buff  cache   si   so    bi    bo   in   cs us sy id wa st
 2  0      0  8123M  145M  3987M    0    0    10   150  500 1000  5  2 93  0  0
```

**Columns:**
- `r`: Processes waiting for CPU
- `b`: Processes blocked on I/O
- `si`/`so`: Swap in/out
- `bi`/`bo`: Blocks in/out (disk)
- `us`: User CPU time
- `sy`: System (kernel) CPU time
- `id`: Idle
- `wa`: I/O wait

```bash
# Continuous network stats
$ sar -n DEV 1
```

---

## Container Debugging

### Docker

**Container logs:**

```bash
$ docker logs container-id
$ docker logs -f container-id        # Follow
$ docker logs --tail 100 container-id
```

**Execute command in container:**

```bash
$ docker exec -it container-id bash
$ docker exec container-id ps aux
$ docker exec container-id cat /proc/1/status
```

**Container stats:**

```bash
$ docker stats
CONTAINER ID   NAME      CPU %   MEM USAGE / LIMIT   MEM %   NET I/O        BLOCK I/O
abc123         web       0.5%    234MiB / 512MiB     45%     1.2MB / 3.4MB  10MB / 15MB
```

**Inspect container:**

```bash
$ docker inspect container-id

# Get specific field
$ docker inspect -f '{{.State.Pid}}' container-id
1234  # Host PID of container's PID 1

# See what PID 1 is on host
$ ps -p 1234
```

### Kubernetes

**Pod logs:**

```bash
$ kubectl logs pod-name
$ kubectl logs pod-name -c container-name  # Specific container
$ kubectl logs pod-name --previous         # Previous crash
```

**Execute in pod:**

```bash
$ kubectl exec -it pod-name -- bash
$ kubectl exec pod-name -- ps aux
```

**Pod events:**

```bash
$ kubectl describe pod pod-name
Events:
  Type     Reason     Message
  ----     ------     -------
  Normal   Scheduled  Successfully assigned default/pod-name to node1
  Normal   Pulling    Pulling image "nginx:1.21"
  Normal   Pulled     Successfully pulled image
  Normal   Created    Created container
  Normal   Started    Started container
```

**Resource usage:**

```bash
$ kubectl top pods
NAME        CPU(cores)   MEMORY(bytes)
pod-1       100m         234Mi
pod-2       250m         512Mi

$ kubectl top nodes
NAME     CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-1   2000m        50%    8Gi             50%
```

---

## Debugging Checklist

When investigating an issue:

1. **Check process status**
   ```bash
   $ ps aux | grep myapp
   $ systemctl status myapp
   ```

2. **Check logs**
   ```bash
   $ journalctl -u myapp -n 100
   $ docker logs myapp
   ```

3. **Check resource usage**
   ```bash
   $ top
   $ free -h
   $ df -h
   $ iostat -x 1
   ```

4. **Check network**
   ```bash
   $ ss -tlnp
   $ ping example.com
   $ curl -v http://localhost:8080
   ```

5. **Check open files/connections**
   ```bash
   $ lsof -p <pid>
   $ ss -tp
   ```

6. **Trace system calls**
   ```bash
   $ strace -p <pid>
   ```

7. **Check kernel messages**
   ```bash
   $ dmesg | tail
   $ journalctl -k
   ```

---

## Key Takeaways

1. **`ps aux` shows all processes — check PID, CPU, memory, state**
2. **`top`/`htop` for live monitoring**
3. **`perf` for CPU profiling, flame graphs for visualization**
4. **`lsof` shows open files and network connections**
5. **`ss` shows sockets (better than netstat)**
6. **`strace` traces system calls — see what a process is doing**
7. **`iostat` shows disk I/O, `iotop` shows per-process I/O**
8. **`journalctl` for all logs, `dmesg` for kernel messages**

---

## What's Next

- [Module 12: Production Concerns](../12-production/) — Kernel limits, tuning
- [Module 13: Real Failure Stories](../13-failure-stories/) — Apply these tools to real scenarios
- [Module 14: Capstone](../14-capstone/) — Complete mental model

---

**Next:** [Module 12: Production Concerns](../12-production/01-kernel-limits.md)
