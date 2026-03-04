# Lesson 1 — Real Interview Questions

These questions come from actual senior/staff-level Node.js interviews. Each answer demonstrates the depth expected.

---

## Category 1: Runtime & Event Loop

### Q1: "Explain what happens when you run `node app.ts` with Node 25."

**Great Answer:**

Node 25 supports running TypeScript files directly via **type stripping** (enabled by default since Node 22.6+):

1. **Process startup**: The OS loads the Node binary (V8 + libuv + Node APIs). V8 initializes its heap, compiles built-in JavaScript modules.

2. **TypeScript handling**: Node detects the `.ts` extension and runs the file through its built-in TypeScript transformer (based on `swc`/`amaro`). This strips type annotations — it does NOT type-check. The output is valid JavaScript.

3. **Module resolution**: Node treats the file as ESM or CJS based on `package.json` `"type"` field or file extension (`.mts`/`.cts`). It resolves imports, loads dependencies.

4. **V8 compilation**: The JS source goes through V8's pipeline:
   - **Ignition** (interpreter): Generates bytecode. Runs immediately — no wait for compilation.
   - **Maglev** (mid-tier JIT): Compiles warm functions to optimized machine code.
   - **TurboFan** (top-tier JIT): Profiles hot functions, applies speculative optimizations (inline caching, hidden class transitions).

5. **Event loop**: After executing top-level code, Node enters the event loop (`uv_run()`). If there are no pending handles (timers, I/O, etc.), the process exits.

**Key detail to emphasize**: Type stripping means you can use TypeScript syntax but NOT features that require transformation (like `enum`, `namespace`, or parameter properties in some configurations). It's type erasure, not compilation.

---

### Q2: "What are the phases of the Node.js event loop? What runs between phases?"

**Great Answer:**

The event loop has **6 phases** executed in order:

```
   ┌───────────────────────────┐
┌─>│        timers              │  setTimeout, setInterval callbacks
│  └───────────┬───────────────┘
│  ┌───────────┴───────────────┐
│  │     pending callbacks      │  System-level callbacks (TCP errors)
│  └───────────┬───────────────┘
│  ┌───────────┴───────────────┐
│  │       idle, prepare        │  Internal use only
│  └───────────┬───────────────┘
│  ┌───────────┴───────────────┐
│  │         poll               │  I/O callbacks (file, network)
│  └───────────┬───────────────┘
│  ┌───────────┴───────────────┐
│  │         check              │  setImmediate callbacks
│  └───────────┬───────────────┘
│  ┌───────────┴───────────────┐
│  │     close callbacks        │  socket.on('close'), etc.
│  └───────────┴───────────────┘
│              │
└──────────────┘
```

**Between every phase**, Node drains two microtask queues:
1. **`process.nextTick()` queue** — runs first, can starve other microtasks
2. **Promise microtask queue** — `.then()`, `await` continuations

**Critical detail**: `process.nextTick()` is NOT part of the event loop. It runs between every phase transition and can starve the event loop if called recursively.

---

### Q3: "What's the difference between `setImmediate()` and `setTimeout(fn, 0)`?"

**Great Answer:**

They run in different **phases** of the event loop:

- `setTimeout(fn, 0)` → **timers phase** (beginning of the loop)
- `setImmediate(fn)` → **check phase** (after the poll phase)

**Key nuance**: When called from top-level code, the order is **non-deterministic** because it depends on whether the 1ms timer threshold has elapsed by the time the event loop starts.

But inside an I/O callback, `setImmediate()` **always** runs before `setTimeout(fn, 0)`:

```typescript
import { readFile } from "node:fs";

readFile("/dev/null", () => {
  setTimeout(() => console.log("timeout"), 0);
  setImmediate(() => console.log("immediate"));
});

// Output: ALWAYS "immediate" then "timeout"
// Because: I/O callback runs in poll phase.
// Next phase is check (setImmediate), then loop restarts at timers.
```

---

### Q4: "How does Node.js handle 10,000 concurrent connections with a single thread?"

**Great Answer:**

Node doesn't handle them "with a single thread" — it uses **kernel-level async I/O** via libuv:

1. **No thread per connection**: Unlike Java's traditional model, Node doesn't spawn a thread per connection. A single thread runs JavaScript code.

2. **Kernel event notification**: libuv uses `epoll` (Linux), `kqueue` (macOS), or `IOCP` (Windows). These kernel mechanisms tell Node which sockets have data ready — without polling.

3. **Non-blocking I/O**: When a request arrives, the callback reads data, processes it, and sends a response. If it needs to do database I/O, it issues an async call and moves on — the thread is free for the next connection.

4. **Thread pool**: For operations where the OS doesn't provide async APIs (DNS resolution, file system on Linux pre-io_uring), libuv uses a thread pool (default 4 threads, max 1024 via `UV_THREADPOOL_SIZE`).

5. **Memory per connection**: Each connection uses ~10-20KB (socket buffer + small JS object). 10,000 connections ≈ 200MB. A Java server with thread-per-connection would use ~10GB (1MB stack per thread).

**The bottleneck is CPU, not I/O**: If each request does 1ms of CPU work, one thread handles ~1000 req/sec. For 10K req/sec, you need cluster mode (multi-process) or worker threads.

---

## Category 2: Memory & V8

### Q5: "How does garbage collection work in V8?"

**Great Answer:**

V8 uses **generational GC** based on the observation that most objects die young:

**Young Generation (Semi-space/Scavenger)**:
- Two equal-sized spaces: "from" and "to"
- New objects allocated in "from" space
- When full (~1-8MB), Scavenger runs:
  - Copies live objects to "to" space
  - Dead objects are simply not copied (free)
  - Swap "from" and "to"
- Objects surviving 2 scavenges are **promoted** to old generation
- Very fast: ~1-5ms, doesn't scan old generation

**Old Generation (Mark-Sweep-Compact)**:
- Larger (~512MB-4GB depending on heap limit)
- Three phases:
  1. **Mark**: Traverse from roots (global, stack), mark reachable objects
  2. **Sweep**: Reclaim unmarked memory
  3. **Compact**: Move objects to reduce fragmentation (optional, expensive)
- Uses **incremental marking**: breaks mark phase into small steps, interleaved with JS execution (~5ms chunks)
- **Concurrent marking**: Marking happens on helper threads while JS runs

**Why this matters for production**:
- Avoid promoting objects unnecessarily (short-lived allocations are fine — they're collected cheaply by Scavenger)
- Large old generation + lots of live objects = long GC pauses
- Memory leaks fill old generation → GC runs more frequently → performance degrades

---

### Q6: "How do you find and fix a memory leak in production?"

**Great Answer:**

**Step 1: Detect** — Monitor `process.memoryUsage().heapUsed` over time. A monotonically increasing heap that doesn't plateau after GC is a leak.

**Step 2: Reproduce** — Send realistic traffic with a load testing tool. Watch heap grow.

**Step 3: Heap snapshot** — Take two snapshots 5 minutes apart:
```typescript
import { writeHeapSnapshot } from "node:v8";
// Expose via HTTP endpoint (protected)
writeHeapSnapshot(); // Produces .heapsnapshot file
```

**Step 4: Compare** — In Chrome DevTools, load both snapshots. Use "Comparison" view. Sort by "Size Delta" — the objects growing most are likely the leak.

**Step 5: Common patterns**:
1. **Unbounded caches**: `Map`/`Set` used as cache without eviction. Fix: LRU with max size.
2. **Event listener accumulation**: Adding listeners in a loop without removing. Fix: `removeListener()` or `AbortSignal`.
3. **Closures capturing scope**: Callbacks holding references to large objects. Fix: Drop references after use (`obj = null`).
4. **Global state**: Module-level arrays/maps that grow with each request.
5. **Timers**: `setInterval()` without `clearInterval()` — the callback's closure keeps referenced objects alive.

**Production tool**: Use `--diagnostic-report-on-signal` or `process.report.writeReport()` to capture a diagnostic report that includes heap statistics and active handles.

---

### Q7: "What are hidden classes and why do they matter for performance?"

**Great Answer:**

V8 creates **hidden classes** (internally called "Maps") to track the shape of objects — which properties exist and their offsets in memory.

```typescript
// V8 creates hidden class transitions:
const obj = {};           // HiddenClass H0 (empty)
obj.x = 1;               // HiddenClass H1 (x at offset 0)
obj.y = 2;               // HiddenClass H2 (x at offset 0, y at offset 4)
```

**Why they matter**: V8 uses **inline caches** to speed up property access. When a function accesses `obj.x`, V8 caches the hidden class and offset. Next call with the same hidden class → direct memory read (no lookup).

```typescript
// Monomorphic (fast) — same hidden class every time
function getX(point) { return point.x; }
getX({ x: 1, y: 2 }); // All objects have same shape → IC stays monomorphic

// Megamorphic (slow) — different hidden classes
getX({ x: 1 });            // Shape 1
getX({ x: 1, y: 2 });      // Shape 2
getX({ x: 1, y: 2, z: 3 }); // Shape 3
getX({ a: 0, x: 1 });      // Shape 4 (different property order!)
// IC becomes megamorphic → falls back to hash table lookup (5-10x slower)
```

**Rules**:
- Always create objects with the same properties in the same order
- Use classes instead of object literals in hot paths
- Never `delete` properties — marks the object as "dictionary mode"
- Initialize all properties in the constructor

---

## Category 3: Streams & I/O

### Q8: "What is backpressure and how does Node.js handle it?"

**Great Answer:**

Backpressure is the mechanism that prevents a fast producer from overwhelming a slow consumer.

In Node.js streams:
- Each writable stream has an internal buffer with a `highWaterMark` (default 16KB for byte streams, 16 objects for object mode)
- `writable.write(chunk)` returns `false` when the buffer exceeds `highWaterMark`
- The readable side should stop pushing data and wait for the `'drain'` event
- When the buffer drains below the watermark, `'drain'` fires, and reading resumes

**`pipeline()` handles this automatically**. With raw `.pipe()` or manual reading, you must handle it yourself:

```typescript
// Manual backpressure handling:
const writable = getSlowWritable();
const readable = getFastReadable();

readable.on("data", (chunk) => {
  const canContinue = writable.write(chunk);
  if (!canContinue) {
    readable.pause();
    writable.once("drain", () => readable.resume());
  }
});
```

**Real-world impact**: Without backpressure handling, a 100MB/s file read piped to a 10MB/s network upload accumulates 90MB/s in memory. After 60 seconds: ~5.4GB buffered → OOM crash.

---

### Q9: "Explain the difference between `Buffer.alloc()`, `Buffer.allocUnsafe()`, and `Buffer.from()`."

**Great Answer:**

- **`Buffer.alloc(size)`**: Allocates `size` bytes, **zero-filled**. Safe but slower because it writes zeros to every byte.

- **`Buffer.allocUnsafe(size)`**: Allocates `size` bytes from the **pre-allocated pool**, does NOT zero-fill. May contain old data from previous allocations. Faster because it skips zeroing. Use only when you'll immediately overwrite all bytes.

- **`Buffer.from(data)`**: Creates a buffer from existing data (string, array, ArrayBuffer). Always a copy — doesn't share memory.

**The pool**: `Buffer.allocUnsafe()` uses an 8KB pre-allocated slab managed by the `Buffer` pooling mechanism. Small allocations (< 4KB) share the same slab, reducing GC pressure.

**Security implication**: `allocUnsafe()` can leak data if you don't overwrite all bytes:

```typescript
// BAD: Leaks old memory contents
const buf = Buffer.allocUnsafe(100);
buf.write("short");  // Only wrote 5 bytes
sendToClient(buf);   // Sends 100 bytes — including 95 bytes of old data!

// SAFE: Zero-filled
const buf2 = Buffer.alloc(100);
buf2.write("short");
sendToClient(buf2);  // Remaining bytes are zeros
```

---

## Category 4: Concurrency

### Q10: "Worker Threads vs. Cluster — when do you use each?"

**Great Answer:**

| Aspect | Worker Threads | Cluster |
|--------|---------------|---------|
| **Model** | Threads in same process | Separate processes (fork) |
| **Memory** | Shared via SharedArrayBuffer | Fully isolated |
| **Communication** | postMessage + SharedArrayBuffer | IPC (serialized messages) |
| **Use case** | CPU-intensive computation | Scaling HTTP servers |
| **Crash impact** | Can crash the whole process | Only the crashed worker dies |
| **Overhead** | ~2MB per thread | ~30MB per process |
| **V8 instance** | One per thread | One per process |
| **Max useful** | CPU core count | CPU core count |

**Use Worker Threads when**:
- You need to do CPU-heavy work (image processing, crypto, parsing) without blocking the event loop
- You want to share memory between workers (SharedArrayBuffer for zero-copy data transfer)
- The overhead of separate processes is too high

**Use Cluster when**:
- You want to scale an HTTP server across all CPU cores
- You need fault isolation (one worker crash shouldn't kill others)
- You're doing primarily I/O work (each process handles its own connections)

**Key insight**: For a typical web API, use **cluster** for horizontal scaling and **worker threads** within each cluster worker for CPU offloading.

---

### Q11: "What are Atomics and when would you use them?"

**Great Answer:**

`Atomics` provides **lock-free, thread-safe** operations on `SharedArrayBuffer` data. They're needed because worker threads can read and write the same memory simultaneously.

Without Atomics:
```typescript
// Thread A:                    Thread B:
// Read value: 5                Read value: 5
// Add 1: 6                    Add 1: 6
// Write 6                     Write 6
// Expected: 7, Got: 6 — race condition!
```

With Atomics:
```typescript
const shared = new SharedArrayBuffer(4);
const view = new Int32Array(shared);

// Thread-safe increment
Atomics.add(view, 0, 1);  // Atomic read-modify-write
```

**Key operations**:
- `Atomics.add/sub/and/or/xor` — atomic read-modify-write
- `Atomics.compareExchange` — CAS (compare-and-swap) for building locks/mutexes
- `Atomics.wait/notify` — thread blocking/signaling (like condition variables)
- `Atomics.load/store` — atomic read/write with memory ordering guarantees

**Use case**: Worker pool with shared counters, implementing mutexes, lock-free queues between threads.

---

## Category 5: Production

### Q12: "How do you implement graceful shutdown in Node.js?"

**Great Answer:**

1. **Listen for SIGTERM** (from Kubernetes/orchestrator) and SIGINT (Ctrl+C)
2. **Mark unhealthy** — health endpoint returns 503 so load balancer stops sending traffic
3. **Wait for LB drain** — 5-10 seconds for the load balancer to detect the unhealthy status
4. **Call `server.close()`** — stops accepting new TCP connections, existing requests continue
5. **Set force timeout** — after 30 seconds, force exit (respect `terminationGracePeriodSeconds`)
6. **Close resources** — drain database connection pools, close Redis, flush logs
7. **Exit 0** — clean exit code

```typescript
process.on("SIGTERM", async () => {
  isHealthy = false;
  await new Promise(r => setTimeout(r, 5000));  // LB drain
  server.close(() => {
    pool.close().finally(() => process.exit(0));
  });
  setTimeout(() => process.exit(1), 30000).unref();
});
```

**The `.unref()` is critical** — without it, the force timeout timer itself keeps the process alive even if all other work is done.

---

### Q13: "What's the maximum size of the Node.js thread pool and when would you change it?"

**Great Answer:**

Default: **4 threads** (`UV_THREADPOOL_SIZE`). Maximum: **1024**.

The thread pool handles:
- **File system operations** (on Linux/macOS, most fs calls)
- **DNS resolution** (`dns.lookup()` — the libc-based one, not `dns.resolve()`)
- **Crypto**: `pbkdf2`, `scrypt`, `randomBytes`
- **Zlib** compression

**When to increase**:
- Heavy file I/O applications where all 4 threads are busy and you see high event loop latency
- Lots of concurrent `dns.lookup()` calls (each blocks a thread until resolved)
- Measure first: track event loop lag and thread pool utilization

**When NOT to increase blindly**:
- More threads ≠ more throughput if the bottleneck is disk I/O speed
- Each thread uses ~1MB of stack memory
- Context-switch overhead increases with too many threads
- Better fix: use `dns.resolve()` (pure network call, no thread pool) instead of `dns.lookup()`

Set it: `UV_THREADPOOL_SIZE=16 node app.ts` (must be set before process starts).

---

## Category 6: Advanced

### Q14: "What is `AsyncLocalStorage` and why would you use it?"

**Great Answer:**

`AsyncLocalStorage` provides a way to store data that follows the async execution context — like thread-local storage but for async operations.

**Use case**: Request context propagation. Without it, you'd pass `requestId` through every function call:

```typescript
// Without ALS — context threading pollution
async function handleRequest(req) {
  const requestId = generateId();
  const user = await getUser(req.userId, requestId);
  const posts = await getPosts(user.id, requestId);
  await logAction(requestId, "fetched posts");
}

// With ALS — context is implicit
const store = new AsyncLocalStorage<{ requestId: string }>();

async function handleRequest(req) {
  store.run({ requestId: generateId() }, async () => {
    const user = await getUser(req.userId);
    const posts = await getPosts(user.id);
    // Any function, anywhere in the call tree, can access requestId:
    // store.getStore().requestId
  });
}
```

**How it works internally**: V8's async hooks track Promise creation. When a Promise is created inside an `ALS.run()` context, the context is propagated to all `.then()` callbacks and `await` continuations. The context "flows" through the entire async call chain.

**Performance**: < 5% overhead in typical applications. The per-request `run()` cost is negligible compared to I/O.

---

### Q15: "Explain the difference between `dns.lookup()` and `dns.resolve()`."

**Great Answer:**

They use completely different mechanisms:

- **`dns.lookup()`**: Calls the operating system's `getaddrinfo()` function via the libuv thread pool. Uses the system's resolver configuration (`/etc/resolv.conf`, `/etc/hosts`, NSS). **Blocks a thread pool thread** until resolution completes.

- **`dns.resolve()`**: Uses c-ares library to perform DNS queries directly over the network. Does NOT use the thread pool — purely async I/O. Goes directly to DNS servers.

**Implications**:
1. `dns.lookup()` respects `/etc/hosts` and system DNS caching. `dns.resolve()` does not.
2. With the default thread pool size of 4, just 4 slow DNS lookups block all file system operations.
3. HTTP/HTTPS modules use `dns.lookup()` by default — this is why you can exhaust the thread pool with many concurrent outbound HTTP requests to different hostnames.

**Fix for high-traffic**: Use a DNS caching layer or configure the HTTP agent to use `dns.resolve()`:
```typescript
import { Agent } from "node:http";
import { Resolver } from "node:dns/promises";

const resolver = new Resolver();
// Custom lookup that uses dns.resolve() instead of dns.lookup()
```

---

### Q16: "What happens if you `await` inside a `forEach()`?"

**Great Answer:**

**Nothing useful** — `forEach()` ignores the return value of the callback, including Promises:

```typescript
const ids = [1, 2, 3];

// ❌ This does NOT wait — all 3 fetch calls fire simultaneously
// and the code after forEach runs before any of them complete
ids.forEach(async (id) => {
  await fetch(`/api/users/${id}`);  // Runs, but forEach doesn't await this
  console.log(`Fetched ${id}`);
});
console.log("Done"); // Prints BEFORE any "Fetched" message

// ✅ Sequential: use for...of
for (const id of ids) {
  await fetch(`/api/users/${id}`);
  console.log(`Fetched ${id}`);
}

// ✅ Parallel: use Promise.all + map
await Promise.all(ids.map(async (id) => {
  await fetch(`/api/users/${id}`);
  console.log(`Fetched ${id}`);
}));
```

Why: `Array.prototype.forEach` calls `callback(element)` but doesn't do anything with the return value. When `callback` is `async`, it returns a Promise — which `forEach` ignores. The Promises float in the void, unhandled.

---

### Q17: "How would you profile a Node.js application in production?"

**Great Answer:**

**CPU Profile** (safe for production):
```bash
# V8 sampling profiler — minimal overhead (1-5%)
node --prof app.ts
# Generates v8.log, process with:
node --prof-process v8.log > profile.txt
```

**Programmatic** (triggered via HTTP endpoint):
```typescript
import { Session } from "node:inspector/promises";

async function captureProfile(durationMs: number) {
  const session = new Session();
  session.connect();
  await session.post("Profiler.enable");
  await session.post("Profiler.start");
  await new Promise(r => setTimeout(r, durationMs));
  const { profile } = await session.post("Profiler.stop");
  session.disconnect();
  return profile; // Load in Chrome DevTools
}
```

**Heap Snapshot** (causes a pause — use carefully):
```typescript
import { writeHeapSnapshot } from "node:v8";
writeHeapSnapshot(); // Freezes process for seconds on large heaps
```

**Diagnostic Report** (safe, instant):
```typescript
process.report.writeReport(); // JSON file with stack, handles, heap stats
```

**Production guidelines**:
- CPU profiles: safe to run for 10-30 seconds
- Heap snapshots: dangerous on large heaps (causes pause), take on a drained instance
- Always expose behind an authenticated HTTP endpoint, never in public routes
- Use `performance.mark()` / `performance.measure()` for custom metrics — zero overhead when not observed

---

### Q18: "What is `io_uring` and why does Bun use it?"

**Great Answer:**

`io_uring` is a Linux kernel interface (5.1+) for async I/O that avoids the overhead of both blocking calls and the traditional `epoll` + thread pool model.

**How it works**:
- Two ring buffers in memory shared between kernel and userspace
- **Submission Queue (SQ)**: Application writes I/O requests
- **Completion Queue (CQ)**: Kernel writes results
- No syscall needed per I/O — just write to shared memory
- Kernel processes requests in batches
- Supports: file read/write, network, fsync, mkdir, etc.

**Why Bun uses it**:
- Node.js uses libuv's thread pool for file I/O on Linux. Default 4 threads → bottleneck under concurrent file operations.
- Bun submits file operations directly to io_uring — no thread pool, no context switching, no thread exhaustion.
- Result: 3-5x faster concurrent file I/O in benchmarks.

**Limitations**:
- Linux only (5.1+ kernel)
- Security concerns (large kernel attack surface — disabled in some container runtimes)
- Node.js chose not to adopt it to maintain cross-platform consistency

---

### Q19: "Explain the V8 JIT compilation pipeline."

**Great Answer:**

V8 has a **tiered compilation** system — code gets progressively more optimized as it runs:

```
Source Code
    ↓
[Parser] → AST
    ↓
[Ignition] → Bytecode
    ↓ (function is "warm")
[Maglev] → Semi-optimized machine code
    ↓ (function is "hot")
[TurboFan] → Fully optimized machine code
    ↓ (type assumption violated)
[Deoptimize] → Back to Ignition bytecode
```

1. **Ignition** (interpreter): Generates bytecode from AST. Executes immediately. Collects **type feedback** — tracks what types each operation receives.

2. **Maglev** (mid-tier JIT, added in V8 11.1): Compiles warm functions using type feedback. Faster than Ignition, compiles faster than TurboFan. Fills the gap between interpreter and full optimization.

3. **TurboFan** (top-tier JIT): Uses type feedback for **speculative optimization**. Assumes types won't change (e.g., "this function always receives Numbers"). Generates highly optimized machine code with type guards.

4. **Deoptimization**: If a type assumption is wrong (e.g., function suddenly receives a String), TurboFan's code is discarded and execution falls back to Ignition bytecode. This is expensive — avoid type inconsistency in hot functions.

**Key insight**: V8 optimizes for **predictability**. Functions that always receive the same types run 10-100x faster than functions with varying types.

---

### Q20: "Design a system to handle 10,000 webhook deliveries per minute."

**Great Answer:**

Architecture:
1. **API receives webhook registration**: Store URL, events, secret in database
2. **Event occurs**: Enqueue a delivery job in Redis (BullMQ or similar)
3. **Worker pool**: 4 cluster workers, each processing 5 concurrent deliveries = 20 in parallel
4. **Delivery**: POST to webhook URL with HMAC signature, timeout 10s
5. **Retry**: Exponential backoff with jitter (1s, 2s, 4s, 8s, 16s), max 5 attempts
6. **Dead letter**: After 5 failures, stop attempting, notify user via email

Key decisions:
- **Rate limiting per URL**: Some endpoints can't handle high throughput. Use token bucket (10 req/sec default per URL).
- **Circuit breaker per domain**: If a destination is down, stop trying for 5 minutes — don't waste retry attempts.
- **Idempotency key**: Include a unique delivery ID in the header. Receivers can deduplicate.
- **Signature**: HMAC-SHA256 of body with webhook secret. Include timestamp to prevent replay attacks.
- **Timeout**: 10 seconds, hard. If they don't respond in 10s, it counts as a failure.

Throughput math: 10,000/min = ~167/sec. At 20 concurrent with average 200ms response = 100/sec. Need ~34 concurrent. Scale to 2 servers with 4 workers × 5 concurrency each = 40 concurrent capacity.
