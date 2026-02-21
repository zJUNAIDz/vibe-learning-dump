# Real Failure Stories

**Production Incidents Explained at the OS Level**

🟡 **Intermediate** | 🔴 **Advanced**

---

## Introduction: Learning from Failures

These are realistic scenarios based on common production incidents. Each explains:
1. **Symptoms** — What users/operators saw
2. **Investigation** — How to debug it
3. **Root cause** — OS-level explanation
4. **Fix** — How to resolve and prevent

---

## Scenario 1: Node.js App Freezes Under Load

### Symptoms

```
Production web service (Node.js)
- Normal traffic: 100 req/s, everything fine
- Spike to 1000 req/s
- Response times spike to 30+ seconds
- Some requests timeout
- Logs show "Event loop blocked" warnings
```

### Investigation

```bash
# Check process state
$ ps aux | grep node
user  1234 99.0 2.0 ... Rl  ... node server.js
#            ^^^         ^^
#            99% CPU    Running (not sleeping)

# Check what the process is doing
$ strace -p 1234
futex(0x7f8a..., FUTEX_WAIT_PRIVATE, 0, NULL) = 0
futex(0x7f8a..., FUTEX_WAIT_PRIVATE, 0, NULL) = 0
# Hmm, waiting on futex (lock)

# Get JavaScript stack trace
$ kill -USR1 1234  # Node.js writes heap dump on USR1
$ node --inspect server.js
# Attach Chrome DevTools

# See stack trace:
calculatePrimes() at worker.js:45
  for (let i = 2; i < n; i++) {  // CPU-intensive loop
    ...
  }
```

### Root Cause

**CPU-bound work on the event loop thread.**

```typescript
// server.ts
app.get('/calculate', (req, res) => {
  const n = parseInt(req.query.n);
  
  // BUG: Blocks event loop
  const primes = calculatePrimes(n);  // Takes 5 seconds for large n
  
  res.json({ primes });
});

// While calculating:
// - Event loop blocked
// - Can't process other requests
// - Can't handle I/O
// - Appears frozen
```

**Why this happens:**
- Node.js is single-threaded (event loop)
- Synchronous CPU work blocks everything
- Other requests queue up
- Timeouts start firing

### Fix

**Option 1: Move to worker thread**

```typescript
import { Worker } from 'worker_threads';

app.get('/calculate', (req, res) => {
  const n = parseInt(req.query.n);
  
  const worker = new Worker('./prime-worker.js', {
    workerData: { n }
  });
  
  worker.on('message', (primes) => {
    res.json({ primes });
  });
  
  worker.on('error', (err) => {
    res.status(500).json({ error: err.message });
  });
});

// prime-worker.js
const { workerData, parentPort } = require('worker_threads');
const primes = calculatePrimes(workerData.n);
parentPort.postMessage(primes);
```

**Option 2: Use a job queue**

```typescript
// Offload to background job processor
app.get('/calculate', async (req, res) => {
  const n = parseInt(req.query.n);
  
  const job = await queue.add('calculate-primes', { n });
  
  res.json({ jobId: job.id, status: 'pending' });
});

// Separate worker process handles CPU-intensive work
```

---

## Scenario 2: Container Gets OOMKilled

### Symptoms

```
Kubernetes pod keeps restarting
Event: "OOMKilled"
Exit code: 137 (128 + 9 = SIGKILL)

$ kubectl get pods
NAME            READY   STATUS      RESTARTS
api-server      0/1     OOMKilled   5

$ kubectl describe pod api-server
...
Last State:   Terminated
Reason:       OOMKilled
Exit Code:    137
```

### Investigation

```bash
# Check memory limit
$ kubectl get pod api-server -o yaml | grep -A 5 resources
resources:
  limits:
    memory: 512Mi
  requests:
    memory: 256Mi

# Check actual memory usage before crash
$ kubectl logs api-server --previous | grep -i memory
# (application logs if any)

# On node where pod ran
$ journalctl -u kubelet | grep -i oom
Feb 21 12:34:56 node1 kernel: Memory cgroup out of memory: Killed process 12345 (node)
```

**Check if memory limit too low:**

```typescript
// app.js - Memory leak?
const cache = new Map();

app.get('/user/:id', async (req, res) => {
  const user = await db.users.findById(req.params.id);
  
  // BUG: Cache grows forever
  cache.set(req.params.id, user);
  
  res.json(user);
});

// After 100k requests, cache has 100k users in memory
// Each user = ~5 KB → 500 MB+
// Exceeds container limit (512 MB) → OOMKilled
```

### Root Cause

**Memory usage exceeded cgroup limit.**

**What happens:**
1. Container configured with 512 MB limit (`memory.max` in cgroup)
2. Application allocates memory (cache grows)
3. Memory usage reaches 512 MB
4. Next allocation triggers page fault
5. Kernel can't allocate (cgroup limit reached)
6. Kernel OOM killer selects process
7. Sends SIGKILL (signal 9) to process
8. Pod terminated, exit code 137

**View cgroup events on node:**

```bash
$ cat /sys/fs/cgroup/kubepods/pod<pod-uid>/memory.events
...
oom_kill 3  # Killed 3 times
```

### Fix

**Option 1: Fix memory leak**

```typescript
// Use LRU cache with max size
import LRU from 'lru-cache';

const cache = new LRU({
  max: 1000,      // Max 1000 entries
  maxAge: 1000 * 60 * 10  // 10 minutes TTL
});

app.get('/user/:id', async (req, res) => {
  let user = cache.get(req.params.id);
  
  if (!user) {
    user = await db.users.findById(req.params.id);
    cache.set(req.params.id, user);
  }
  
  res.json(user);
});
```

**Option 2: Increase memory limit**

```yaml
# pod.yaml
resources:
  limits:
    memory: 2Gi  # Increased from 512Mi
  requests:
    memory: 1Gi
```

**Option 3: Monitor and alert**

```typescript
// Expose metrics
app.get('/metrics', (req, res) => {
  const used = process.memoryUsage();
  res.json({
    heapUsed: used.heapUsed,
    heapTotal: used.heapTotal,
    external: used.external,
    rss: used.rss
  });
});

// Prometheus alert when memory > 80% of limit
```

---

## Scenario 3: "Address Already in Use" After Restart

### Symptoms

```bash
$ docker stop api-server
$ docker run --name api-server -p 8080:8080 api-image

Error: Cannot start service api: address already in use
```

### Investigation

```bash
# Check what's using port 8080
$ sudo lsof -i :8080
COMMAND   PID USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
node     1234 user   10u  IPv4  12345      0t0  TCP *:8080 (LISTEN)

# Old process still running
$ ps -p 1234
  PID TTY      STAT   TIME COMMAND
 1234 ?        Sl     0:05 node server.js

$ kill 1234

# Try again - still fails!
$ sudo lsof -i :8080
# No output now

$ ss -tln | grep :8080
LISTEN 0  128  *:8080  *:*
State: TIME_WAIT
```

### Root Cause

**TCP socket in TIME_WAIT state.**

**TCP connection close sequence:**

```
Client                    Server
  |                         |
  |-------- FIN ----------->|  (Server closes)
  |<------- ACK ------------|
  |<------- FIN ------------|  (Client closes)
  |-------- ACK ----------->|
  |                         |
  Server enters TIME_WAIT (waits 2*MSL = 60-120 seconds)
```

**Why TIME_WAIT exists:**
- Ensures all packets from old connection are gone
- Prevents packets from old connection arriving after new connection starts on same port
- Typically lasts 60 seconds

**During TIME_WAIT:**
- Port appears "in use"
- Can't bind new socket to that port by default

### Fix

**Option 1: Use SO_REUSEADDR (typical solution)**

```typescript
import * as net from 'net';

const server = net.createServer();

// Allow reusing address in TIME_WAIT
server.listen({
  port: 8080,
  host: '0.0.0.0',
  exclusive: false  // Implicitly sets SO_REUSEADDR
});
```

**C equivalent:**

````c
int optval = 1;
setsockopt(sockfd, SOL_SOCKET, SO_REUSEADDR, &optval, sizeof(optval));
bind(sockfd, ...);
```

**Option 2: Wait for TIME_WAIT to expire**

```bash
$ # Wait 60-120 seconds
$ docker run ... # Now works
```

**Option 3: Change TIME_WAIT duration (not recommended)**

```bash
# Reduce TIME_WAIT duration (affects all connections!)
$ sudo sysctl -w net.ipv4.tcp_fin_timeout=15
# Default is 60 seconds
```

---

## Scenario 4: "Too Many Open Files"

### Symptoms

```typescript
// Node.js application
const server = http.createServer(async (req, res) => {
  const data = await fs.promises.readFile('/data/file.txt');
  res.end(data);
});

// After running for a while:
Error: EMFILE: too many open files, open '/data/file.txt'
```

### Investigation

```bash
# Check process limits
$ cat /proc/1234/limits
Limit                     Soft Limit           Hard Limit
Max open files            1024                 4096

# Check currently open files
$ ls /proc/1234/fd/ | wc -l
1023  # At the limit!

# See what files are open
$ ls -l /proc/1234/fd/
lrwx------ 1 user user 64 0 -> /dev/pts/0
lrwx------ 1 user user 64 1 -> /dev/pts/0
lrwx------ 1 user user 64 2 -> /dev/pts/0
lrwx------ 1 user user 64 3 -> socket:[12345]
lrwx------ 1 user user 64 4 -> socket:[12346]
...
lrwx------ 1 user user 64 1023 -> socket:[99999]

# Mostly sockets - connection leak?
$ ls -l /proc/1234/fd/ | grep socket | wc -l
1020  # 1020 open sockets!
```

### Root Cause

**File descriptor leak — connections never closed.**

```typescript
// Bug example
const http = require('http');

function fetchData(url) {
  return new Promise((resolve, reject) => {
    http.get(url, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => resolve(data));
      // BUG: No error handling, connection stays open on error
    });
  });
}

// If remote server times out/errors:
// - Socket stays open
// - File descriptor never closed
// - After 1024 requests with errors: EMFILE
```

### Fix

**Option 1: Properly close connections**

```typescript
function fetchData(url) {
  return new Promise((resolve, reject) => {
    const req = http.get(url, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => resolve(data));
      res.on('error', (err) => {
        req.destroy();  // Close connection
        reject(err);
      });
    });
    
    req.on('error', (err) => {
      req.destroy();
      reject(err);
    });
    
    req.setTimeout(5000, () => {
      req.destroy();
      reject(new Error('Timeout'));
    });
  });
}
```

**Option 2: Use library with built-in connection pooling**

```typescript
import axios from 'axios';

// axios handles connection pooling and cleanup automatically
const response = await axios.get(url, { timeout: 5000 });
```

**Option 3: Increase ulimit (temporary workaround)**

```bash
# For current shell
$ ulimit -n 65536

# System-wide (/etc/security/limits.conf)
* soft nofile 65536
* hard nofile 65536

# For Docker container
$ docker run --ulimit nofile=65536:65536 ...
```

**Monitoring:**

```typescript
// Add metrics
setInterval(() => {
  const fds = fs.readdirSync('/proc/self/fd').length;
  console.log(`Open file descriptors: ${fds}`);
  if (fds > 900) {
    console.error('WARNING: Approaching fd limit!');
  }
}, 60000);
```

---

## Scenario 5: DNS Works on Host, Fails in Container

### Symptoms

```bash
# On host
$ ping example.com
PING example.com (93.184.216.34) 56(84) bytes of data.
64 bytes from 93.184.216.34: icmp_seq=1 ttl=56 time=10 ms
# Works fine

# In container
$ docker run ubuntu ping example.com
ping: example.com: Name or service not known
# Fails!
```

### Investigation

```bash
# Check container's /etc/resolv.conf
$ docker run ubuntu cat /etc/resolv.conf
nameserver 192.168.1.1  # Local router

# Can the container reach the nameserver?
$ docker run ubuntu ping 192.168.1.1
PING 192.168.1.1: 56 data bytes
Request timeout
# Can't reach it!

# Why? Check Docker network
$ docker network inspect bridge
"Config": [
  {
    "Subnet": "172.17.0.0/16",
    "Gateway": "172.17.0.1"
  }
]

# Container is on 172.17.x.x network
# Trying to reach 192.168.1.1 (host network)
# No route!
```

### Root Cause

**Container in isolated network namespace, can't reach host's DNS server.**

**Docker copies host `/etc/resolv.conf` to container**, but:
- Host uses local DNS (192.168.1.1)
- Container on different network (172.17.0.0/16)
- No route from container to 192.168.1.1

### Fix

**Option 1: Use public DNS servers**

```bash
$ docker run --dns=8.8.8.8 --dns=1.1.1.1 ubuntu ping example.com
# Works!
```

**Or in docker-compose.yml:**

```yaml
services:
  app:
    image: ubuntu
    dns:
      - 8.8.8.8
      - 1.1.1.1
```

**Option 2: Use host network (bypasses network namespace)**

```bash
$ docker run --network=host ubuntu ping example.com
# Works, but container shares host network (less isolation)
```

**Option 3: Configure Docker daemon**

```json
// /etc/docker/daemon.json
{
  "dns": ["8.8.8.8", "1.1.1.1"]
}
```

```bash
$ sudo systemctl restart docker
```

---

## Scenario 6: Kubernetes Pod CrashLoopBackOff

### Symptoms

```bash
$ kubectl get pods
NAME                READY   STATUS              RESTARTS
api-deployment      0/1     CrashLoopBackOff    10

$ kubectl logs api-deployment
Error: Cannot find module 'express'
```

### Investigation

```bash
# Check why it's crashing
$ kubectl describe pod api-deployment
...
State:          Waiting
  Reason:       CrashLoopBackOff
Last State:     Terminated
  Reason:       Error
  Exit Code:    1

# Get logs
$ kubectl logs api-deployment
Error: Cannot find module 'express'
    at Function.Module._resolveFilename (internal/modules/cjs/loader.js:...)

# Check Dockerfile
$ cat Dockerfile
FROM node:18
WORKDIR /app
COPY package.json .
# BUG: Forgot to run npm install!
COPY . .
CMD ["node", "server.js"]
```

### Root Cause

**Image missing dependencies because `npm install` not run during build.**

**Why CrashLoopBackOff:**
1. Process exits immediately (exit code 1)
2. Kubernetes tries to restart
3. Process crashes again
4. Kubernetes backs off (wait longer between restarts)
5. Repeat: Waiting → Running → CrashLoopBackOff

### Fix

**Fix Dockerfile:**

```dockerfile
FROM node:18
WORKDIR /app

# Copy package files first
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy application code
COPY . .

CMD ["node", "server.js"]
````

**Rebuild and redeploy:**

```bash
$ docker build -t api-image:v2 .
$ kubectl set image deployment/api-deployment api=api-image:v2
$ kubectl get pods
NAME                READY   STATUS    RESTARTS
api-deployment      1/1     Running   0
```

---

## Scenario 7: System Becomes Unresponsive Under I/O Load

### Symptoms

```
Database server under heavy load
- System extremely slow
- SSH sessions freeze
- Commands take minutes to respond
- Top shows low CPU usage (10-20%)
```

### Investigation

```bash
# Check load average
$ uptime
 12:34:56 up 5 days, 3:21,  1 user,  load average: 45.23, 38.12, 32.45
#                                                   ^^^^^ Very high!

# But CPU is idle?
$ top
%Cpu(s): 12.5 us,  5.5 sy,  0.0 ni, 82.0 id,  0.0 wa,  0.0 hi,  0.0 si,  0.0 st
#                                           ^^^^ Only 82% idle (should be OK)

# Check process states
$ ps aux | awk '{print $8}' | sort | uniq -c
  5 R   # Running
 89 S   # Sleeping
 25 D   # Uninterruptible sleep (DISK I/O!)
#  ^^

# Processes stuck in D state
$ ps aux | grep ' D '
postgres 1234  0.0  1.0 ... D ... postgres: writer process
postgres 1235  0.0  1.2 ... D ... postgres: checkpointer
...

# Check I/O wait
$ iostat -x 1
Device   rrqm/s wrqm/s  r/s   w/s   rkB/s wkB/s util
sda        0.0  5000.0  0.0 2000.0    0.0 80000.0 100.0%
#                                                  ^^^^^ Disk saturated!
```

### Root Cause

**Disk I/O saturation — processes blocked waiting for slow disk.**

**Why system seems frozen:**
- Many processes in D state (uninterruptible sleep)
- Waiting for disk I/O to complete
- Can't respond to signals (not even SIGKILL)
- Disk is bottleneck (100% util)

**Common causes:**
- Database writing too much (no transaction log tuning)
- Backup running during peak hours
- No RAID / slow disks
- Filesystem full (triggers excessive metadata operations)

### Fix

**Immediate:**

```bash
# Identify what's writing
$ iotop -o
Total DISK WRITE: 80.00 M/s
  PID  USER   DISK WRITE  COMMAND
 1234  postgres  76.00 M/s  postgres: writer process

# Reduce PostgreSQL aggressiveness
$ psql -c "ALTER SYSTEM SET checkpoint_timeout = '15min';"
$ systemctl reload postgresql

# Or kill offending process (if safe)
$ kill 1234
```

**Long-term:**

1. **Better hardware:**
   - Use SSDs instead of HDDs
   - RAID for performance
   - More RAM for page cache

2. **Tune database:**
   ```sql
   -- PostgreSQL example
   ALTER SYSTEM SET shared_buffers = '8GB';
   ALTER SYSTEM SET effective_cache_size = '24GB';
   ALTER SYSTEM SET checkpoint_completion_target = 0.9;
   ```

3. **Schedule heavy I/O:**
   - Run backups off-peak
   - Batch operations
   - Use throttling

4. **Monitor:**
   ```bash
   # Alert when I/O util > 80%
   $ while true; do
       util=$(iostat -x 1 2 | tail -1 | awk '{print $NF}' | cut -d. -f1)
       if [ $util -gt 80 ]; then
           echo "ALERT: Disk util ${util}%"
       fi
       sleep 10
     done
   ```

---

## Common Debugging Patterns

### Pattern 1: High CPU

```bash
1. Identify process:
   $ top

2. Check what it's doing:
   $ strace -p PID
   $ perf record -p PID
   $ perf report

3. Get stack trace:
   $ pstack PID  # For C/C++
   $ kill -USR1 PID  # For Node.js (writes to stderr)
   $ kill -SIGQUIT PID  # For Go (writes stack trace)
```

### Pattern 2: High Memory

```bash
1. Check process memory:
   $ ps aux --sort=-%mem | head

2. Inspect process:
   $ cat /proc/PID/status | grep -i vm
   $ pmap -x PID

3. Application-specific:
   # Node.js: Heap snapshot
   $ kill -USR2 PID
   # Analyze in Chrome DevTools

   # Go: pprof
   $ curl http://localhost:6060/debug/pprof/heap > heap.prof
   $ go tool pprof heap.prof
```

### Pattern 3: Network Issues

```bash
1. Verify connectivity:
   $ ping host
   $ traceroute host

2. Check port:
   $ telnet host port
   $ nc -zv host port

3. Check local sockets:
   $ ss -tlnp  # Listening
   $ ss -tnp   # Established

4. Trace packets:
   $ sudo tcpdump -i any port 3000
```

---

## Key Takeaways

1. **Event loop blocking causes Node.js to freeze — use workers for CPU work**
2. **OOMKilled = exceeded cgroup memory limit — check for leaks or increase limit**
3. **TIME_WAIT state can prevent immediate port reuse — use SO_REUSEADDR**
4. **File descriptor leaks cause "too many open files" — always close fds**
5. **Container DNS issues often due to network namespace isolation**
6. **CrashLoopBackOff usually means immediate crash — check logs and exit code**
7. **D state (uninterruptible sleep) indicates I/O bottleneck**

---

## What's Next

- [Module 11: Performance & Debugging](../11-debugging/) — Deep dive into debugging tools
- [Module 14: Capstone](../14-capstone/) — Framework for reasoning about any issue

---

**Next:** [Module 14: Capstone: Think Like the Kernel](../14-capstone/01-debugging-framework.md)
