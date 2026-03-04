# Lesson 3 — Whiteboard Explanations

## How to Explain Runtime Concepts Visually

In interviews, you're often asked to explain a concept on a whiteboard (or virtual equivalent). This lesson provides step-by-step drawings for the most common topics.

---

## 1. The Event Loop

### Drawing Steps

```
Step 1: Draw the loop as a circle with labeled phases

          ┌─────────────┐
    ┌────>│   timers     │─────┐
    │     └─────────────┘     │
    │     ┌─────────────┐     │
    │     │   pending    │<────┘
    │     └──────┬──────┘
    │     ┌──────┴──────┐
    │     │    poll      │
    │     └──────┬──────┘
    │     ┌──────┴──────┐
    └─────│    check     │
          └─────────────┘

Step 2: Add microtask queues BETWEEN phases

          ┌─────────────┐
    ┌────>│   timers     │─── [nextTick] [microtasks] ──┐
    │     └─────────────┘                               │
    │     ┌─────────────┐                               │
    │     │   pending    │<─────────────────────────────┘
    │     └──────┬──────┘
    │   [nextTick] [microtasks]
    │     ┌──────┴──────┐
    │     │    poll      │     (I/O callbacks)
    │     └──────┬──────┘
    │   [nextTick] [microtasks]
    │     ┌──────┴──────┐
    └─────│    check     │     (setImmediate)
          └─────────────┘

Step 3: Annotate what goes where:
  - timers:   setTimeout, setInterval
  - poll:     fs callbacks, network callbacks
  - check:    setImmediate
  - between:  process.nextTick, Promise.then
```

### Talking Points

1. "The event loop is a cycle of phases. Each phase has a FIFO queue of callbacks."
2. "Between every phase, Node drains ALL microtasks — nextTick first, then Promise callbacks."
3. "The poll phase is where most I/O happens. If nothing is pending, Node will block here waiting for I/O."
4. "setTimeout(fn, 0) goes to timers. setImmediate goes to check. Inside an I/O callback, setImmediate always runs first because poll → check → next loop → timers."

---

## 2. V8 Memory Layout

### Drawing Steps

```
Step 1: Draw the heap as two main areas

┌────────────────────────────────────────────────┐
│                    V8 HEAP                      │
│                                                │
│  ┌────────────────┐  ┌─────────────────────┐  │
│  │ Young Generation│  │   Old Generation     │  │
│  │  (1-8 MB)      │  │   (up to 4 GB)      │  │
│  │                │  │                     │  │
│  │  ┌────┐┌────┐  │  │                     │  │
│  │  │From││ To │  │  │                     │  │
│  │  └────┘└────┘  │  │                     │  │
│  └────────────────┘  └─────────────────────┘  │
└────────────────────────────────────────────────┘

Step 2: Show the Scavenger GC flow

  From-Space            To-Space
  ┌──────────┐         ┌──────────┐
  │ A (live) │────────>│ A (copy) │
  │ B (dead) │   X     │          │
  │ C (live) │────────>│ C (copy) │
  │ D (dead) │   X     │          │
  └──────────┘         └──────────┘
  
  Then swap: To becomes From.
  Dead objects aren't copied = free memory.
  Objects surviving 2 copies → promoted to Old.

Step 3: Show the Mark-Sweep for Old Generation

  Before:
  ┌──────────────────────────────┐
  │ A  ██  B  ░░  C  ██  D  ░░  │
  │ ██=live   ░░=dead            │
  └──────────────────────────────┘

  After Sweep:
  ┌──────────────────────────────┐
  │ A  ██     C  ██              │
  │       free       free        │
  └──────────────────────────────┘

  After Compact (optional):
  ┌──────────────────────────────┐
  │ A  ██  C  ██                 │
  │              all free        │
  └──────────────────────────────┘
```

### Talking Points

1. "V8 divides the heap into young and old generations. Most objects die young."
2. "Young gen uses semi-space: copy live objects, dead ones are just left behind. Very fast — 1-5ms."
3. "Old gen uses mark-sweep: walk from roots, mark what's reachable, sweep the rest."
4. "Objects that survive 2 young gen collections get promoted to old gen."
5. "Compact is expensive — moves objects to reduce fragmentation. Only done when needed."

---

## 3. How `require()` / `import` Works

### Drawing Steps

```
Step 1: Module Resolution

  require("express")
      │
      ├─ Is it a core module? (fs, http, path)
      │    └─ YES → return built-in module
      │
      ├─ Starts with "./" or "../"?
      │    └─ YES → resolve relative to current file
      │        ├─ Try: ./express.js, ./express.ts
      │        ├─ Try: ./express/index.js
      │        └─ Try: ./express/package.json → main
      │
      └─ Search node_modules
           ├─ ./node_modules/express
           ├─ ../node_modules/express
           ├─ ../../node_modules/express
           └─ ... up to root

Step 2: Module Loading & Caching

  require("./module-a")
      │
  ┌───┴────────────────────┐
  │ Check require.cache    │
  │  Key: absolute path    │
  │                        │
  │  Found? ──── YES ──── → Return cached exports
  │    │                   │
  │   NO                   │
  │    │                   │
  │  Create new Module obj │
  │  Add to cache FIRST    │  ← Before executing!
  │  Execute module code   │     (handles circular deps)
  │  Return module.exports │
  └────────────────────────┘
```

### Key Point to Explain

"The module is added to the cache BEFORE its code runs. This is how Node handles circular dependencies — module A requires B, B requires A, but A is already in the cache (partially loaded). B gets A's partially-initialized exports."

---

## 4. Stream Pipeline and Backpressure

### Drawing Steps

```
Step 1: Normal flow

  Readable ──chunk──> Transform ──chunk──> Writable
  (100MB/s)            (50MB/s)            (50MB/s)
  
  Everything flows smoothly.

Step 2: Backpressure kicks in

  Readable ──chunk──> Transform ──chunk──> Writable
  (100MB/s)            (50MB/s)            (10MB/s)
                                            │
                                   Buffer filling up!
                                   write() returns false
                                            │
  Transform ◄── "slow down" ◄── "drain" ───┘
  Readable  ◄── "pause"     ◄── Transform
  
  Result: Read pauses. Memory stays constant.

Step 3: Draw the buffer watermark

  Writable internal buffer:
  ┌──────────────────────────────┐
  │████████████████░░░░░░░░░░░░░│  < highWaterMark (16KB)
  └───────┬──────────────┬──────┘
     data in         data out
     (fast)          (slow)
  
  When █ fills past highWaterMark:
    write() returns false → upstream pauses
  When █ drains:
    'drain' event fires → upstream resumes
```

### Talking Points

1. "Streams process data in small chunks — constant memory regardless of total data size."
2. "Each stream has a highWaterMark — a buffer size limit."
3. "When the buffer fills, write() returns false. The upstream should stop sending data."
4. "When the buffer drains below the watermark, the 'drain' event fires and upstream resumes."
5. "pipeline() handles all of this automatically. .pipe() does backpressure but not error propagation."

---

## 5. Cluster Architecture

### Drawing Steps

```
Step 1: Basic cluster

  ┌─────────────────────────────────────┐
  │           Primary Process            │
  │                                     │
  │  fork()  fork()  fork()  fork()     │
  └───┬───────┬───────┬───────┬─────────┘
      │       │       │       │
  ┌───┴──┐┌───┴──┐┌───┴──┐┌───┴──┐
  │ W1   ││ W2   ││ W3   ││ W4   │
  │:3000 ││:3000 ││:3000 ││:3000 │
  └──────┘└──────┘└──────┘└──────┘
      ▲       ▲       ▲       ▲
      └───────┴───┬───┴───────┘
                  │
           Single port :3000
           OS distributes via
           round-robin (Linux)

Step 2: How port sharing works

  OS Kernel
  ┌──────────────────────────┐
  │ Listening socket :3000   │
  │ (owned by primary)       │
  │                          │
  │ Incoming connection ──────→ Primary accepts
  │                          │  │
  │                          │  ├─ Send to W1 (IPC)
  │                          │  ├─ Send to W2 (IPC)
  │                          │  └─ Round-robin or
  │                          │     OS-scheduled
  └──────────────────────────┘

Step 3: Graceful restart flow

  Time →
  W1: ████████████████  ← serving
  W2: ████████████████  ← serving
  
  Deploy:
  W1: ████████████ drain ▓▓ stop
  W1': ── start ─ ready ████████  ← new version
  W2: ████████████████████ drain ▓▓ stop
  W2': ────────── start ─ ready ████████
  
  At any point, at least one worker serves traffic.
```

---

## 6. libuv and Kernel I/O

### Drawing Steps

```
Step 1: The layers

  ┌─────────────────────┐
  │   JavaScript Code    │
  │   (your app)         │
  ├─────────────────────┤
  │   Node.js Bindings   │
  │   (C++ layer)        │
  ├─────────────────────┤
  │       libuv          │  ← This is the key layer
  │   (event loop +      │
  │    thread pool +     │
  │    kernel async)     │
  ├─────────────────────┤
  │   Kernel             │
  │   (epoll/kqueue/     │
  │    IOCP/io_uring)    │
  └─────────────────────┘

Step 2: Which I/O goes where

  JavaScript: fs.readFile()
      │
  libuv decides:
      │
      ├─ Network I/O → epoll/kqueue (kernel async)
      │    └─ TCP, UDP, DNS (c-ares) — no thread needed
      │
      ├─ File I/O → thread pool
      │    └─ Linux has no true async file I/O*
      │    └─ Thread does blocking read, notifies event loop
      │
      └─ DNS lookup → thread pool
           └─ getaddrinfo() is blocking

  * io_uring (Linux 5.1+) changes this — Bun uses it
```

### Talking Points

1. "Network I/O uses kernel mechanisms (epoll) — truly async, no threads."
2. "File I/O goes through the thread pool because Linux doesn't have good async file I/O APIs."
3. "The thread pool default size is 4. This is why 4 concurrent fs operations or DNS lookups can block everything."
4. "Bun avoids the thread pool for files by using io_uring — direct kernel I/O without threads."

---

## 7. Worker Threads vs. Cluster vs. child_process

### Drawing Steps

```
           Process 1 (main)
  ┌──────────────────────────────┐
  │  ┌──────┐  ┌──────┐        │
  │  │Thread│  │Thread│  ...    │  Worker Threads
  │  │(JS)  │  │(JS)  │        │  - Share process memory
  │  └──────┘  └──────┘        │  - Can share ArrayBuffer
  │                              │  - One crash = all die
  │  One V8 per thread           │
  │  Shared: process, heap*      │
  └──────────────────────────────┘

           vs.

  ┌──────────────┐  ┌──────────────┐
  │  Process 1   │  │  Process 2   │
  │  (primary)   │  │  (worker)    │  Cluster / child_process
  │              │  │              │  - Separate memory
  │  V8 heap     │  │  V8 heap     │  - Communicate via IPC
  │  30MB+       │  │  30MB+       │  - One crash = only that one
  └──────┬───────┘  └──────┬───────┘
         │    IPC (pipe)    │
         └─────────────────┘

  * SharedArrayBuffer is explicitly shared, not implicit
```

### Decision Framework

```
"Do you need CPU parallelism or I/O scaling?"

  CPU-heavy work              I/O scaling (HTTP server)
  (image processing,          (handle more connections)
   crypto, parsing)
       │                            │
       ▼                            ▼
  Worker Threads                Cluster Mode
  - Lower overhead (~2MB)      - Fault isolation
  - Can share memory           - ~30MB per worker
  - Good for offloading        - OS handles load balancing
    compute from main thread   - Good for multi-core HTTP
```

---

## 8. How to Structure a Whiteboard Explanation

### The 4-Step Framework

```
1. OVERVIEW (10 seconds)
   "At a high level, [concept] does X because Y."
   
2. DIAGRAM (30 seconds)
   Draw the main components and their connections.
   Label arrows with data flow direction.
   
3. WALKTHROUGH (60 seconds)
   "When [trigger] happens, first A does X,
    then B receives it and does Y,
    finally C handles the result."
   
4. NUANCE (30 seconds)
   "The interesting part is [non-obvious detail].
    This matters because [production impact]."
```

### Example: "Explain how a request is handled in Node.js"

**Step 1 — Overview**: "Node uses an event-driven model. A single thread handles all JavaScript execution, but I/O is delegated to the kernel."

**Step 2 — Diagram**:
```
Client → TCP → OS Kernel → libuv → Event Loop → JS Callback
                                                      │
                                                  DB query
                                                      │
                                                  (async)
                                                      │
                                               ← Response ← 
```

**Step 3 — Walkthrough**: "A TCP connection comes in. The kernel notifies libuv via epoll. libuv queues the callback. On the poll phase, the event loop runs the callback — your route handler. If the handler does a DB query, that's an async operation. The thread is free to handle other requests. When the DB responds, the callback is queued again. The response is sent."

**Step 4 — Nuance**: "The key insight is that JavaScript never waits. Between the DB request and response, the thread handles hundreds of other requests. The bottleneck is CPU time per request, not connection count."

---

## Practice Exercises

### Exercise 1: Explain Memory Leaks
Draw a diagram showing how a closed-over variable in an event listener prevents garbage collection. Show the reference chain from GC root → closure → leaked object.

### Exercise 2: Explain Backpressure to a Non-Engineer
"Imagine a factory conveyor belt. The cutting machine processes 100 items/minute. The painting machine handles 10 items/minute. Without a feedback mechanism, 90 unpainted items pile up every minute. Backpressure is the sensor that tells the cutting machine to slow down when items pile up."

### Exercise 3: Draw the JIT Pipeline
V8's Ignition → Maglev → TurboFan pipeline with deoptimization arrows going backward. Label what triggers promotions (hot functions) and demotions (type changes).

### Exercise 4: Explain async/await to Someone Who Knows Callbacks
"Instead of nesting callbacks, you write code that LOOKS synchronous. Each `await` is a suspension point — the function pauses, the event loop handles other work, and when the async operation completes, the function resumes exactly where it left off."
