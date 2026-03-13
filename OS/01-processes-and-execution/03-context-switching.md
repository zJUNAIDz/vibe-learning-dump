# Context Switching

**The Kernel's Magic Sleight of Hand**

🟢 **Beginner-Friendly** | 🟡 **Intermediate**

---

## Introduction

If you have 4 CPU cores, you can physically only run 4 things at the exact same millisecond. 
But if you run `top`, you'll see hundreds of processes "running." How is this possible? 
It's not magic, it's just the kernel being incredibly fast at swapping things out.

This illusion of infinite multitasking is powered by **Context Switching** and the OS **Scheduler**. It’s giving main character energy to every single app on your laptop simultaneously. 🌟

---

## What is a Context Switch?

Imagine you are reading a book (doing work). Suddenly, your phone rings (an interrupt).
To answer the phone, you must:
1. Put a bookmark in the book exactly where you stopped.
2. Put the book down.
3. Pick up the phone and talk.
4. Put the phone down.
5. Pick up the book, open to the bookmark, and resume reading.

That process of "saving state," doing something else, and "restoring state" is exactly what a **Context Switch** is for a CPU.

When the OS switches from `Process A` to `Process B`, it:
1. Pauses `Process A`.
2. Saves all the CPU registers, instruction pointers (the "bookmark"), and state of `A` into RAM (the Process Control Block).
3. Loads the saved state for `Process B` from RAM into the CPU registers.
4. Resumes `Process B`.

### The Catch: It's expensive!
Context switching takes time (microseconds). If you switch too often, the CPU spends more time swapping books than actually reading them. This is called **Thrashing**. Total vibes killer.

---

## Voluntary vs Involuntary Switches

How does a process lose the CPU? Two ways:

### 1. Voluntary (It yielded)
The process says, "I'm waiting on something slow (like disk I/O, a network request, or `sleep(1)`). You can have the CPU back until I get my data."
- Blocked on a read/write.
- Waiting on a mutex/lock.
- Very polite. Very mindful. 🧘

### 2. Involuntary (Preemption)
The process is stuck in an infinite `while(true)` loop crunching numbers. The kernel's **scheduler** sets an interrupt timer. When the timer pops (the "time slice" or "quantum" is up), a hardware interrupt fires. The kernel kicks the door open, freezes the process mid-calculation, and hands the CPU to the next process in line. 
- "Your turn is over, bro." Preempted!

---

## How to observe it?

You don't have to guess. The OS tracks this.

```bash
# Check the context switches of a specific process using pidstat
$ pidstat -w -p <PID> 1

10:00:01      UID       PID   cswch/s nvcswch/s  Command
10:00:02     1000      1234    150.00     10.00  node
```
- `cswch/s`: Voluntary context switches per second.
- `nvcswch/s`: Non-voluntary (involuntary) context switches per second.

If `nvcswch/s` is super high, your app is CPU bound and fighting for time slices.
If `cswch/s` is super high, your app is doing a lot of I/O (network/disk) or locking.

---

## Why does this matter to developers?

1. **Async I/O (Node.js, Go)**: Node.js has one thread. It tries to MINIMIZE context switches by doing non-blocking I/O using `epoll` (we cover this later). Instead of spawning 10,000 threads for 10,000 requests (which would cause massive context switching overhead and crash), it handles them all on one thread efficiently. Big brain moves. 🧠
2. **Spinlocks vs Mutexes**: If you hold a lock, should a waiting thread just spin in a `while` loop (burn CPU, save the context switch) or sleep (save CPU, but pay the context switch cost)? It depends on how long the lock is held!
3. **Database Tuning**: High context switching on a DB server usually means too many concurrent connections fighting over limited CPU cores. Use a connection pool so you stop thrashing your CPU cache.

## Key Takeaways
1. A **context switch** is saving the state of one process to load another.
2. It gives the illusion of parallelism but consumes CPU time.
3. Too many threads = too much context switching = bad performance. Don't be that guy.

---
**Next:** [Module 02: Memory Fundamentals](../02-memory/01-memory-fundamentals.md)