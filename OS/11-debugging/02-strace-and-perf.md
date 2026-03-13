# System Calls and Profiling: strace & perf

**Staring Directly Into the Matrix**

🟡 **Intermediate** | 🔴 **Advanced**

---

## Introduction

So `top` says your app is using 100% CPU. Okay, cool. **WHY?**
Is it stuck in an infinite `while` loop? Is it waiting for a lock? Is it reading a file 10,000 times a second?

If you just restart the app, you're literally ghosting the problem. Let's act like engineers and look under the hood. To do that, we use the two strongest observability tools in the Linux ecosystem: `strace` and `perf`.

---

## `strace`: The System Call Tracer

Remember Module 00? An application cannot touch hardware without asking the kernel permission via a **System Call**.
`strace` safely attaches to any running process and prints out *every single system call* it makes in real time. It is the ultimate spyware for your own apps.

### The Problem: "My app is hung. It's just frozen."

```bash
# 1. Find the PID
$ ps aux | grep my_frozen_app
user      1234  ...  my_frozen_app

# 2. Attach strace to it
$ sudo strace -p 1234
```

**What you might see:**
```text
futex(0x7ffd0b..., FUTEX_WAIT_PRIVATE, 0, NULL
```
**Translation:** Your app is deadlocked. `futex` is a fast-userspace-mutex (a lock). It is permanently waiting for another thread to release a lock.

Or maybe you see this:
```text
connect(3, {sa_family=AF_INET, sin_port=htons(5432), ...})
... (frozen)
```
**Translation:** Your app is trying to connect to a database on port 5432, but the DB isn't responding, and you forgot to set a connection timeout. Big oof.

### Finding "File Not Found" Bugs instantly

App failing to start because it can't find a config file, but the error message is unhelpful trash?

```bash
# Run the app under strace, filter specifically for `open` or `openat` calls
$ strace -e trace=openat node server.js

openat(AT_FDCWD, "/app/config.json", O_RDONLY) = -1 ENOENT (No such file or directory)
openat(AT_FDCWD, "/etc/app/config.json", O_RDONLY) = -1 ENOENT (No such file or directory)
```
Boom. You can see exactly where it's looking for `config.json` and failing (`ENOENT`). You don't even have to read the source code.

---

## `perf`: The Profiler of the Gods

`strace` is great for debugging slow I/O and crashes. But what if your app is just burning 100% CPU on math/logic? `strace` won't show anything, because math doesn't require System Calls!

For CPU-bound issues, we use `perf`.
`perf` interrupts the CPU 99 times a second and takes a snapshot of *exactly which function* the CPU is currently executing. Then it aggregates the results.

### How to Profile a CPU Spike

```bash
# 1. Record the CPU for 10 seconds on your spicy process
$ sudo perf record -F 99 -p 1234 -g -- sleep 10

[ perf record: Woken up 1 times to write data ]
[ perf record: Captured and wrote 0.052 MB perf.data (123 samples) ]

# 2. View the report
$ sudo perf report
```

The output will look like this:
```text
  45.31%  node     [.] calculatePrimeNumbers
  22.10%  node     [.] parseLargeJsonSubroutine
  10.05%  kernel   [k] copy_page_range
```
Wait... `calculatePrimeNumbers` is eating 45% of your CPU? 
Congratulations, you just found the exact line of code causing the lag spike without putting a single `console.log()` in your app. Huge W.

### Flamegraphs 🔥

Reading `perf report` in the terminal is cool, but real chads use **Flamegraphs**.
Created by performance legend Brendan Gregg, a Flamegraph takes `perf.data` and turns it into an interactive SVG chart you can open in your browser. 
The wider the bar, the more CPU time that function consumed. The vertical axis shows the stack trace. 

*(If you ever want to look like a senior engineer in an incident response Zoom call, just say: "Let me pull a flamegraph of the CPU profile." They will hand you the keys to the company.)*

---

## The Catch: Performance Overhead

1. **strace**: Tracing system calls forces the kernel to pause your app, write the log, and resume it for *every single call*. If you run `strace` on a production database doing 10,000 queries a second, the DB will slow down to a crawl. Use with extreme caution in prod.
2. **perf**: Uses statistical sampling. It has extremely low overhead (like 1%). It is generally completely safe to run in production.

## Key Takeaways
1. Use `strace` for debugging **interactions with the OS**: File reads, network hang-ups, missing configs, locks, permissions.
2. Use `perf` for debugging **CPU usage**: Infinite loops, slow algorithms, garbage collection spikes.
3. Don't guess why something is broken. The OS knows exactly what the process is doing. Just ask it.

---
**Next:** [Module 12: Production Limitations](../12-production/01-kernel-limits.md)
