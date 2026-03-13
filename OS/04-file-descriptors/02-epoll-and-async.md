# epoll(), select(), and Async I/O

**How Node.js & Nginx Serve 10k Connections Without Sweating**

🟡 **Intermediate** | 🔴 **Advanced**

---

## Introduction

So you wrote a web server. It listens on port 8080. When a user connects, what happens?

In the old days (we're talking 1990s), servers worked like this:
1. `accept()` a new connection.
2. Fork a brand new process (or spawn a thread) to handle that connection.
3. The new thread calls `read()` and **BLOCKS (sleeps)** until the user sends data.
4. When the user disconnects, kill the thread.

### The C10k Problem

This model is fine for 10 users. But what if 10,000 users connect at once? 
You'd need 10,000 threads. As we learned in Module 01, threads are heavy! 
10,000 threads means:
- 10,000 stack memory allocations (GBs of RAM wasted).
- Continuous, chaotic Context Switching. CPU thrashing.
- The server flatlines. Down completely. Huge L.

This was known as the **C10k Problem** (Handling 10,000 Concurrent Connections).

---

## The Big Brain Solution: Asynchronous, Non-Blocking I/O

What if 1 thread could handle all 10,000 connections?

To do this, we need to mark our file descriptors (sockets) as **Non-Blocking**.
When a socket is non-blocking, calling `read()` on it won't put your thread to sleep. If there's no data, the kernel instantly replies: "EAGAIN: Try again later, bro. Empty inbox."

Okay, but how does the 1 thread know *which* of the 10,000 sockets actually has data ready to read?
Does it just loop through all 10,000 very fast, asking "You got data? You got data?"
No. That's a busy-wait loop. It would burn 100% of your CPU doing nothing.

We need the kernel to tap us on the shoulder and say: *"Hey, out of those 10,000 sockets, sockets #42, #1003, and #9999 just received data."*

---

## The Evolution of OS Tapping

Unix engineers came up with several tools for this over the decades.

### 1. `select()` (The Boomer)
- **Vibe:** "Kernel, here is a list of 10,000 FDs. Tell me which ones have data."
- **Flaw:** Every single time you call `select()`, you have to pass the *entire array* of 10,000 FDs into the kernel. The kernel scans them all, updates the array, and passes it back. It is O(N). If you have 10,000 connections, it's slow as molasses.

### 2. `poll()` (The Gen X)
- Basically exactly the same as `select()` but tweaked to avoid arbitrary array size limits. Still O(N). Still incredibly slow with thousands of connections.

### 3. `epoll()` (The Gen Z Savior)
Linux introduced `epoll` in kernel 2.5.44. It is the absolute GOAT. 👑

Instead of passing the list of 10,000 FDs back and forth every millisecond, `epoll` works like a subscription service inside the kernel:
1. **`epoll_create()`**: You tell the kernel, "Create an event registry for me."
2. **`epoll_ctl()`**: You tell the kernel, "Subscribe my server to FDs 1 through 10,000. Keep track of them."
3. **`epoll_wait()`**: Your thread goes to sleep. It says, "Kernel, wake me up when ANY of my subscribed FDs have data."

Because the kernel is already tracking the network card, when a packet arrives for FD #42, the kernel instantly puts #42 into an "active list" and wakes up your thread, handing it ONLY the active FDs. It is **O(1)**.

10 connections or 100,000 connections? `epoll` literally doesn't care. It takes the same amount of time to wake your thread up.

---

## How This Powers Modern Tech

If you've ever used Node.js, Nginx, or Redis, you've used `epoll` (or its equivalents like `kqueue` on Mac/BSD, or `IOCP` on Windows).

### The Node.js Event Loop
When people say Node.js is "Event-Driven and Non-Blocking," they are literally just describing an infinite `while` loop calling `epoll_wait()`.

Under the hood, Node's C++ engine (libuv) does this:
```c
// Pseudocode for the Node.js Event Loop
while (true) {
    // 1. Sleep until the OS says SOME socket has data/events (epoll_wait)
    events = waitForEvents();
    
    // 2. Loop through ONLY the active events
    for (event in events) {
        if (event.type == "HTTP_REQUEST") {
            // 3. Call the JavaScript callback you wrote like app.get('/', ...)
            runJavaScriptCallback(event.request);
        }
    }
}
// 1 Thread. Zero blocking. Maximum throughput.
```

### Golang Goroutines
Go takes a different approach. Go lets you write blocking code (which is easier to read), like spawning 10,000 goroutines! 
Wait, didn't we say 10,000 threads is bad?
Yes! But goroutines aren't OS threads. They are "Virtual Threads."

Under the hood, the Go Runtime takes your 10,000 goroutines and multiplexes them onto just ~4 real OS threads. When your goroutine tries to `read()` from the network, the Go runtime intercepts the call, parks your goroutine, adds the socket to `epoll`, and switches to a different goroutine instantly.
To you, the developer, it looks like a blocking thread. To the Linux kernel, it's just one highly efficient `epoll` loop. Absolutely unhinged levels of optimization.

---

## Key Takeaways
1. Creating a new OS thread for every network connection doesn't scale (The C10k problem).
2. **Non-blocking I/O + `epoll`** allows a single thread to monitor thousands of File Descriptors efficiently via kernel "subscriptions."
3. This exact OS system call is the engine behind Node.js, Nginx, Redis, and Go's insane concurrency performance.

---
**Next:** [Module 05: Networking Fundamentals](../05-networking-fundamentals/01-networking-basics.md)