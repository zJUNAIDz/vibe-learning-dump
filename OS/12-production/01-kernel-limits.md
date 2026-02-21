# Production Concerns & Kernel Limits

**Tuning Linux for Production Workloads**

🟡 **Intermediate** | 🔴 **Advanced**

---

## Introduction

Production systems need tuning beyond defaults. This module covers:
- Kernel limits (`ulimit`, `sysctl`)
- Common bottlenecks
- Performance tuning
- Monitoring

---

## ulimit: Per-Process Limits

**`ulimit`** — Limits for shell and its child processes

```bash
# View all limits
$ ulimit -a
core file size          (blocks, -c) 0
data seg size           (kbytes, -d) unlimited
scheduling priority             (-e) 0
file size               (blocks, -f) unlimited
pending signals                 (-i) 15408
max locked memory       (kbytes, -l) 65536
max memory size         (kbytes, -m) unlimited
open files                      (-n) 1024  # ← Important!
pipe size            (512 bytes, -p) 8
POSIX message queues     (bytes, -q) 819200
real-time priority              (-r) 0
stack size              (kbytes, -s) 8192
cpu time               (seconds, -t) unlimited
max user processes              (-u) 15408
virtual memory          (kbytes, -v) unlimited
file locks                      (-x) unlimited
```

### Open Files Limit

**Most common limit issue:**

```bash
$ ulimit -n
1024  # Default: too low for servers

# Increase for current shell
$ ulimit -n 65536

# Check process limit
$ cat /proc/<pid>/limits | grep "open files"
Max open files      1024      4096      files
#                   soft      hard
```

**Make permanent:**

`/etc/security/limits.conf`:

```bash
# User limits
myuser  soft  nofile  65536
myuser  hard  nofile  65536

# All users
*       soft  nofile  65536
*       hard  nofile  65536

# Root
root    soft  nofile  65536
root    hard  nofile  65536
```

**For systemd services:**

```ini
[Service]
LimitNOFILE=65536
```

**For Docker containers:**

```bash
$ docker run --ulimit nofile=65536:65536 myimage
```

```yaml
# docker-compose.yml
services:
  app:
    ulimits:
      nofile:
        soft: 65536
        hard: 65536
```

### Other Important Limits

**Max processes:**

```bash
$ ulimit -u
15408

# Increase
$ ulimit -u 30000
```

`/etc/security/limits.conf`:

```bash
*  soft  nproc  30000
*  hard  nproc  30000
```

**Stack size:**

```bash
$ ulimit -s
8192  # 8 MB

# Increase (for deep recursion)
$ ulimit -s 16384
```

---

## sysctl: Kernel Parameters

**`sysctl`** — Runtime kernel tuning

```bash
# View all parameters
$ sysctl -a

# View specific parameter
$ sysctl net.ipv4.tcp_fin_timeout
net.ipv4.tcp_fin_timeout = 60

# Set temporarily
$ sudo sysctl -w net.ipv4.tcp_fin_timeout=30

# Make permanent
$ echo "net.ipv4.tcp_fin_timeout = 30" | sudo tee -a /etc/sysctl.conf
$ sudo sysctl -p  # Reload
```

### File System Limits

**Max open files (system-wide):**

```bash
$ sysctl fs.file-max
fs.file-max = 9223372036854775807

# Increase
$ sudo sysctl -w fs.file-max=2000000
```

`/etc/sysctl.conf`:

```bash
fs.file-max = 2000000
```

**Check current usage:**

```bash
$ cat /proc/sys/fs/file-nr
1234    0    2000000
# open  free  max
```

**Inotify limits (file watchers):**

```bash
# Max watches per user
$ sysctl fs.inotify.max_user_watches
fs.inotify.max_user_watches = 8192

# Increase (for IDEs, build tools)
$ sudo sysctl -w fs.inotify.max_user_watches=524288
```

`/etc/sysctl.conf`:

```bash
fs.inotify.max_user_watches = 524288
fs.inotify.max_user_instances = 512
```

### Network Tuning

**TCP connection backlog:**

```bash
# Max SYN backlog (pending connections)
$ sysctl net.core.somaxconn
net.core.somaxconn = 128  # Default: too low!

# Increase
$ sudo sysctl -w net.core.somaxconn=4096
```

**Port range:**

```bash
# Local port range for outgoing connections
$ sysctl net.ipv4.ip_local_port_range
net.ipv4.ip_local_port_range = 32768    60999

# Increase range (more connections)
$ sudo sysctl -w net.ipv4.ip_local_port_range="10000 65535"
```

**TIME_WAIT reuse:**

```bash
# Reuse sockets in TIME_WAIT state
$ sudo sysctl -w net.ipv4.tcp_tw_reuse=1
```

**TCP keepalive:**

```bash
# Time before sending keepalive probes (default 7200 = 2 hours)
$ sysctl net.ipv4.tcp_keepalive_time
net.ipv4.tcp_keepalive_time = 7200

# Reduce (detect dead connections faster)
$ sudo sysctl -w net.ipv4.tcp_keepalive_time=600   # 10 minutes
$ sudo sysctl -w net.ipv4.tcp_keepalive_intvl=60   # Probe every 60s
$ sudo sysctl -w net.ipv4.tcp_keepalive_probes=3   # 3 probes before giving up
```

**TCP buffer sizes:**

```bash
# Read buffer sizes (min, default, max)
$ sysctl net.ipv4.tcp_rmem
net.ipv4.tcp_rmem = 4096    87380   6291456

# Write buffer sizes
$ sysctl net.ipv4.tcp_wmem
net.ipv4.tcp_wmem = 4096    16384   4194304

# Increase for high-bandwidth networks
$ sudo sysctl -w net.ipv4.tcp_rmem="4096 87380 16777216"
$ sudo sysctl -w net.ipv4.tcp_wmem="4096 87380 16777216"
```

**SYN cookies (DDoS protection):**

```bash
# Enable SYN cookies (prevent SYN flood)
$ sudo sysctl -w net.ipv4.tcp_syncookies=1
```

### Memory and VM Tuning

**Swappiness:**

```bash
# How aggressively to use swap (0-100)
$ sysctl vm.swappiness
vm.swappiness = 60  # Default

# Reduce swap usage (prefer RAM)
$ sudo sysctl -w vm.swappiness=10
```

**`vm.swappiness` values:**
- `0`: Avoid swap, only use when necessary
- `10`: Minimal swap (good for databases)
- `60`: Default (balanced)
- `100`: Aggressive swap

**Overcommit memory:**

```bash
$ sysctl vm.overcommit_memory
vm.overcommit_memory = 0

# Values:
# 0 = heuristic (default)
# 1 = always allow overcommit (no OOM killer)
# 2 = never overcommit (strict accounting)

# For databases: use 2 or 0
$ sudo sysctl -w vm.overcommit_memory=2
```

**Dirty pages (disk write buffering):**

```bash
# Max % of memory that can be dirty pages
$ sysctl vm.dirty_ratio
vm.dirty_ratio = 20  # 20% of RAM

# Background flush threshold
$ sysctl vm.dirty_background_ratio
vm.dirty_background_ratio = 10

# Reduce for databases (flush more frequently)
$ sudo sysctl -w vm.dirty_ratio=10
$ sudo sysctl -w vm.dirty_background_ratio=5
```

---

## Common Production Configurations

### High-Traffic Web Server

`/etc/sysctl.conf`:

```bash
# Network
net.core.somaxconn = 65535
net.core.netdev_max_backlog = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.ip_local_port_range = 10000 65535
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 30

# TCP buffers
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216

# File handles
fs.file-max = 2000000

# Swap
vm.swappiness = 10
```

`/etc/security/limits.conf`:

```bash
*  soft  nofile  65536
*  hard  nofile  65536
*  soft  nproc   30000
*  hard  nproc   30000
```

### Database Server

`/etc/sysctl.conf`:

```bash
# Memory
vm.swappiness = 1  # Avoid swap
vm.dirty_ratio = 10
vm.dirty_background_ratio = 5
vm.overcommit_memory = 2

# Huge pages (for large databases)
vm.nr_hugepages = 1024  # 2GB (if using huge pages)

# Shared memory (for PostgreSQL, Oracle)
kernel.shmmax = 68719476736  # 64GB
kernel.shmall = 4294967296

# Network (if database accessed over network)
net.core.somaxconn = 4096
net.ipv4.tcp_keepalive_time = 600
net.ipv4.tcp_keepalive_intvl = 60
net.ipv4.tcp_keepalive_probes = 3
```

### Container Host

`/etc/sysctl.conf`:

```bash
# Network (many containers)
net.ipv4.ip_forward = 1  # Required for Docker
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-ip6tables = 1

# Connection tracking (for NAT)
net.netfilter.nf_conntrack_max = 1048576

# File handles (many open files)
fs.file-max = 2000000
fs.inotify.max_user_watches = 1048576
fs.inotify.max_user_instances = 1024

# ARP cache (many IPs)
net.ipv4.neigh.default.gc_thresh1 = 80000
net.ipv4.neigh.default.gc_thresh2 = 90000
net.ipv4.neigh.default.gc_thresh3 = 100000
```

---

## Monitoring Kernel Limits

### File Descriptor Usage

**System-wide:**

```bash
$ cat /proc/sys/fs/file-nr
12345   0   2000000
# Current usage: 12345 / 2000000

# Percentage
$ awk '{printf "%.2f%%\n", ($1/$3)*100}' /proc/sys/fs/file-nr
0.62%
```

**Per-process:**

```bash
$ ls /proc/<pid>/fd | wc -l

# Compare to limit
$ cat /proc/<pid>/limits | grep "open files"
Max open files      1024      4096      files
```

**Alert when approaching limit:**

```bash
#!/bin/bash
PID=$1
LIMIT=$(cat /proc/$PID/limits | awk '/open files/ {print $4}')
CURRENT=$(ls /proc/$PID/fd | wc -l)
PERCENT=$((CURRENT * 100 / LIMIT))

if [ $PERCENT -gt 80 ]; then
    echo "WARNING: Process $PID using $PERCENT% of file descriptors"
fi
```

### Connection Tracking

**Check conntrack usage:**

```bash
$ cat /proc/sys/net/netfilter/nf_conntrack_count
12345

$ cat /proc/sys/net/netfilter/nf_conntrack_max
65536

# Percentage
$ echo "scale=2; $(cat /proc/sys/net/netfilter/nf_conntrack_count) * 100 / $(cat /proc/sys/net/netfilter/nf_conntrack_max)" | bc
18.86%
```

**Increase if approaching limit:**

```bash
$ sudo sysctl -w net.netfilter.nf_conntrack_max=1048576
```

---

## Transparent Huge Pages (THP)

**What are huge pages?**
- Normal pages: 4 KB
- Huge pages: 2 MB or 1 GB
- Reduces page table overhead for large memory allocations

**Check status:**

```bash
$ cat /sys/kernel/mm/transparent_hugepage/enabled
always [madvise] never
#      ^^^^^^^^
# Current setting
```

**Disable (recommended for databases):**

```bash
$ echo never | sudo tee /sys/kernel/mm/transparent_hugepage/enabled

# Permanent (/etc/rc.local or systemd service)
echo never > /sys/kernel/mm/transparent_hugepage/enabled
echo never > /sys/kernel/mm/transparent_hugepage/defrag
```

**Why disable for databases:**
- PostgreSQL, MongoDB, Redis recommend disabling THP
- Can cause latency spikes
- Database engines already optimize memory usage

---

## CPU Affinity and NUMA

### CPU Affinity

**Pin process to specific CPU cores:**

```bash
# Start process on CPUs 0-3
$ taskset -c 0-3 ./my-server

# Set affinity of running process
$ taskset -cp 0-3 1234
```

**Why:**
- Reduces CPU cache misses
- Predictable performance
- Useful for latency-sensitive applications

### NUMA (Non-Uniform Memory Access)

**Check NUMA topology:**

```bash
$ numactl --hardware
available: 2 nodes (0-1)
node 0 cpus: 0 1 2 3
node 0 size: 16384 MB
node 0 free: 8192 MB
node 1 cpus: 4 5 6 7
node 1 size: 16384 MB
node 1 free: 10240 MB
```

**Run process on specific NUMA node:**

```bash
# Run on node 0
$ numactl --cpunodebind=0 --membind=0 ./my-server
```

**Why:**
- Memory access to local node is faster
- Cross-node access has latency penalty
- Important for database servers

---

## Disk I/O Scheduling

**Check I/O scheduler:**

```bash
$ cat /sys/block/sda/queue/scheduler
[mq-deadline] kyber bfq none
```

**Available schedulers:**
- `mq-deadline`: Default for SSDs (good for most workloads)
- `kyber`: For fast multi-queue devices
- `bfq`: Fair queueing (good for desktops)
- `none`: No scheduling (for NVMe drives)

**Change scheduler:**

```bash
$ echo "none" | sudo tee /sys/block/sda/queue/scheduler

# Permanent (/etc/udev/rules.d/60-scheduler.rules)
ACTION=="add|change", KERNEL=="sd[a-z]", ATTR{queue/scheduler}="mq-deadline"
ACTION=="add|change", KERNEL=="nvme[0-9]n[0-9]", ATTR{queue/scheduler}="none"
```

---

## Common Performance Issues

### Issue 1: "Too Many Open Files"

**Symptoms:**

```
Error: EMFILE: too many open files
```

**Check:**

```bash
# Process limit
$ cat /proc/<pid>/limits | grep "open files"

# System limit
$ cat /proc/sys/fs/file-nr
```

**Fix:**

```bash
# Increase ulimit
$ ulimit -n 65536

# Or in /etc/security/limits.conf
* soft nofile 65536
* hard nofile 65536
```

### Issue 2: Port Exhaustion

**Symptoms:**

```
Cannot assign requested address
```

**Check:**

```bash
# Available ports
$ sysctl net.ipv4.ip_local_port_range
net.ipv4.ip_local_port_range = 32768 60999
# Only ~28k ports available

# Current connections
$ ss -tan | wc -l
```

**Fix:**

```bash
# Increase port range
$ sudo sysctl -w net.ipv4.ip_local_port_range="10000 65535"

# Enable TIME_WAIT reuse
$ sudo sysctl -w net.ipv4.tcp_tw_reuse=1
```

### Issue 3: Connection Queue Full

**Symptoms:**

```
Connection refused (under load)
SYN packets dropped
```

**Check:**

```bash
# Listen queue overflows
$ netstat -s | grep -i "listen queue"

# Current backlog
$ ss -lnt
State    Recv-Q Send-Q Local Address:Port
LISTEN   129    128    0.0.0.0:80
#        ^^^    ^^^
#        Current backlog size, Max backlog
```

**Fix:**

```bash
# Increase backlog
$ sudo sysctl -w net.core.somaxconn=4096
$ sudo sysctl -w net.ipv4.tcp_max_syn_backlog=4096

# In application code
# Node.js:
server.listen(3000, () => {}, 4096);  // backlog = 4096
```

### Issue 4: Swap Thrashing

**Symptoms:**

```
System slow
High I/O wait
Swap usage increasing
```

**Check:**

```bash
$ free -h
              total        used        free      shared  buff/cache   available
Mem:           15Gi        14Gi       100Mi       200Mi       1.0Gi       500Mi
Swap:         8.0Gi       3.0Gi       5.0Gi
#                         ^^^^ Using swap

$ vmstat 1
procs -----------memory---------- ---swap-- -----io---- -system-- ------cpu-----
 r  b   swpd   free   buff  cache   si   so    bi    bo   in   cs us sy id wa st
 2  5  3000M   100M   50M   1000M  100  200   500  1000  ... ...  5  2 10 83  0
#                                  ^^^  ^^^                         ^^ High I/O wait
#                                  Swap in/out
```

**Fix:**

```bash
# Reduce swappiness
$ sudo sysctl -w vm.swappiness=10

# Or add more RAM / reduce workload
```

---

## Automated Monitoring Script

```bash
#!/bin/bash
# /usr/local/bin/monitor-limits.sh

LOG="/var/log/limit-monitor.log"

check_limit() {
    local name=$1
    local current=$2
    local max=$3
    local warn_threshold=80
    
    if [ $max -eq 0 ]; then
        return
    fi
    
    local percent=$((current * 100 / max))
    
    if [ $percent -gt $warn_threshold ]; then
        echo "$(date): WARNING: $name at $percent% ($current/$max)" | tee -a $LOG
    fi
}

# File descriptors
FD_CURRENT=$(cat /proc/sys/fs/file-nr | awk '{print $1}')
FD_MAX=$(cat /proc/sys/fs/file-nr | awk '{print $3}')
check_limit "File descriptors" $FD_CURRENT $FD_MAX

# Connection tracking
if [ -f /proc/sys/net/netfilter/nf_conntrack_count ]; then
    CT_CURRENT=$(cat /proc/sys/net/netfilter/nf_conntrack_count)
    CT_MAX=$(cat /proc/sys/net/netfilter/nf_conntrack_max)
    check_limit "Connection tracking" $CT_CURRENT $CT_MAX
fi

# Memory
MEM_AVAILABLE=$(free -m | awk '/^Mem:/ {print $7}')
MEM_TOTAL=$(free -m | awk '/^Mem:/ {print $2}')
MEM_USED=$((MEM_TOTAL - MEM_AVAILABLE))
check_limit "Memory" $MEM_USED $MEM_TOTAL

# Disk
df -h | awk 'NR>1 && $5+0 > 80 {print $0}' | while read line; do
    echo "$(date): WARNING: Disk usage high: $line" | tee -a $LOG
done
```

**Run via cron:**

```bash
$ sudo crontab -e
*/5 * * * * /usr/local/bin/monitor-limits.sh
```

---

## Key Takeaways

1. **`ulimit -n` sets per-process open files limit — increase to 65536 for servers**
2. **`sysctl` tunes kernel parameters — make permanent in `/etc/sysctl.conf`**
3. **`fs.file-max` is system-wide file limit — set to 2000000+**
4. **`net.core.somaxconn` controls connection backlog — increase to 4096+**
5. **`vm.swappiness` controls swap usage — reduce to 10 for databases**
6. **Monitor limits before they're hit — alert at 80% usage**

---

## What's Next

- [Module 11: Performance & Debugging](../11-debugging/) — Tools for investigating issues
- [Module 13: Real Failure Stories](../13-failure-stories/) — See these limits hit in practice
- [Module 14: Capstone](../14-capstone/) — Complete mental model

---

**Next:** [Module 14: Capstone](../14-capstone/01-debugging-framework.md)
