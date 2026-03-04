# Lesson 2 — Trick Questions & Gotchas

These questions are specifically designed to catch developers who have surface-level knowledge. Each exposes a common misconception.

---

## Gotcha 1: "Is Node.js single-threaded?"

### The Trap

Most developers say "yes." That's wrong.

### Correct Answer

**JavaScript execution** is single-threaded. **Node.js** is not.

Node.js uses multiple threads:
- **Main thread**: Runs JavaScript, the event loop
- **libuv thread pool**: 4 threads by default for file I/O, DNS lookup, crypto
- **V8 helper threads**: Compile JavaScript (TurboFan), garbage collect (concurrent marking)
- **Worker threads**: Additional JS threads you create explicitly

A typical Node process has **7+ threads** even without Worker Threads:

```bash
# Check yourself:
node -e "setTimeout(() => {}, 60000)" &
# Find the PID
ps aux | grep node
# Count threads
ls /proc/<PID>/task | wc -l
# You'll see 7-11 threads
```

**Why the misconception persists**: "Single-threaded" refers to the JS execution model — only one piece of JS runs at a time. You don't need mutexes for JS objects. But the runtime itself is heavily multi-threaded.

---

## Gotcha 2: "What does `setTimeout(fn, 0)` do?"

### The Trap

"It runs immediately" or "It runs after 0 milliseconds."

### Correct Answer

It runs **after the current phase of the event loop completes, on the next timers phase**, with a minimum delay of **1 millisecond** (per the HTML spec that Node follows, the actual minimum is clamped).

```typescript
console.log("A");

setTimeout(() => console.log("B"), 0);

Promise.resolve().then(() => console.log("C"));

process.nextTick(() => console.log("D"));

console.log("E");

// Output: A, E, D, C, B
// 
// Why:
// A, E — synchronous, runs first
// D — process.nextTick, runs before microtasks
// C — Promise microtask, runs after nextTick
// B — setTimeout, runs on next timers phase (after microtasks)
```

**The real gotcha**: `setTimeout(fn, 0)` does NOT mean "run as soon as possible." `process.nextTick()` and `queueMicrotask()` both run sooner.

---

## Gotcha 3: "Does `async/await` make code concurrent?"

### The Trap

"Yes, `await` runs things in parallel."

### Correct Answer

No. `async/await` is **syntax for sequential Promises**. Each `await` pauses the function until the awaited value resolves:

```typescript
// Sequential — one at a time
async function badExample() {
  const user = await fetchUser(1);      // 100ms
  const posts = await fetchPosts(1);    // 100ms
  const comments = await fetchComments(1); // 100ms
  // Total: ~300ms
}

// Parallel — all at once
async function goodExample() {
  const [user, posts, comments] = await Promise.all([
    fetchUser(1),      // 100ms ┐
    fetchPosts(1),     // 100ms ├─ overlap
    fetchComments(1),  // 100ms ┘
  ]);
  // Total: ~100ms
}
```

**`await` is a suspension point**, not a parallelism mechanism. To run things concurrently, start the Promises (by calling the functions) and then await them together.

---

## Gotcha 4: "What's wrong with this code?"

```typescript
const server = require("http").createServer((req, res) => {
  const data = JSON.parse(req.body);
  res.end("OK");
});
```

### The Trap

Developers who've used Express assume `req.body` is available and parsed.

### Correct Answer

**Three problems**:

1. **`req.body` doesn't exist** in raw Node.js. `IncomingMessage` is a `Readable` stream. You must consume it:

```typescript
const chunks: Buffer[] = [];
req.on("data", (chunk) => chunks.push(chunk));
req.on("end", () => {
  const body = Buffer.concat(chunks).toString();
  const data = JSON.parse(body);
});
```

2. **No error handling** on `JSON.parse()`. Malformed JSON crashes the server.

3. **No Content-Type check**. The client might send XML, form data, or nothing.

Corrected:
```typescript
import { createServer, IncomingMessage } from "node:http";

createServer(async (req, res) => {
  if (req.headers["content-type"] !== "application/json") {
    res.writeHead(415);
    res.end("Unsupported Media Type");
    return;
  }

  const chunks: Buffer[] = [];
  for await (const chunk of req) {
    chunks.push(chunk as Buffer);
  }
  
  try {
    const data = JSON.parse(Buffer.concat(chunks).toString());
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ received: true }));
  } catch {
    res.writeHead(400);
    res.end("Invalid JSON");
  }
}).listen(3000);
```

---

## Gotcha 5: "What's the output?"

```typescript
for (var i = 0; i < 3; i++) {
  setTimeout(() => console.log(i), 0);
}
```

### The Trap

Many say "0, 1, 2."

### Correct Answer

**3, 3, 3**

`var` is function-scoped, not block-scoped. By the time the timeouts fire, the loop has completed and `i` is 3.

Fixes:
```typescript
// Fix 1: Use let (block-scoped)
for (let i = 0; i < 3; i++) {
  setTimeout(() => console.log(i), 0);
} // 0, 1, 2

// Fix 2: IIFE (old-school)
for (var i = 0; i < 3; i++) {
  ((j) => setTimeout(() => console.log(j), 0))(i);
} // 0, 1, 2
```

**Modern Node**: This shouldn't appear in production TypeScript, but it tests understanding of closures and scope.

---

## Gotcha 6: "Can you catch this error?"

```typescript
try {
  setTimeout(() => {
    throw new Error("boom");
  }, 0);
} catch (err) {
  console.log("Caught:", err);
}
```

### The Trap

"Yes — the try/catch wraps it."

### Correct Answer

**No.** The `try/catch` only covers synchronous code. By the time the timeout callback runs, the `try/catch` is long gone — we're in a different stack frame.

The error becomes an **uncaught exception** that crashes the process.

```typescript
// Fix 1: Put try/catch INSIDE the callback
setTimeout(() => {
  try {
    throw new Error("boom");
  } catch (err) {
    console.log("Caught:", err);
  }
}, 0);

// Fix 2: Global handler (last resort)
process.on("uncaughtException", (err) => {
  console.error("Uncaught:", err);
  process.exit(1); // Always exit after uncaughtException
});

// Fix 3: Use async/await with Promises (try/catch works)
try {
  await new Promise((_, reject) => {
    setTimeout(() => reject(new Error("boom")), 0);
  });
} catch (err) {
  console.log("Caught:", err); // This works!
}
```

---

## Gotcha 7: "What does `require()` return the second time you call it?"

```typescript
// module-a.ts
console.log("Module A loaded");
export const value = Math.random();

// main.ts
const a1 = require("./module-a");
const a2 = require("./module-a");
console.log(a1.value === a2.value); // ?
```

### The Trap

Some think it re-executes the module.

### Correct Answer

**`true`** — `require()` caches the module after the first load.

`console.log("Module A loaded")` runs only **once**. The second `require()` returns the same cached object from `require.cache`.

This is called the **module singleton pattern** and has implications:

```typescript
// counter.ts
let count = 0;
export function increment() { return ++count; }

// a.ts
import { increment } from "./counter";
increment(); // 1

// b.ts  
import { increment } from "./counter";
increment(); // 2 — same module instance!
```

**Gotchas with caching**:
- Circular dependencies: partially loaded modules can be returned
- `delete require.cache[require.resolve("./module")]` forces re-load
- ESM (`import`) also caches, but the cache is different and not programmatically clearable

---

## Gotcha 8: "Is `process.nextTick()` part of the event loop?"

### The Trap

"Yes, it's in one of the phases."

### Correct Answer

**No.** `process.nextTick()` runs **between** event loop phases, not during any phase.

It's its own queue that drains completely before the event loop moves to the next phase. This means recursive `process.nextTick()` can **starve the event loop**:

```typescript
// This blocks EVERYTHING — I/O, timers, all of it
function recursive() {
  process.nextTick(recursive);
}
recursive(); // Event loop never advances

// This does NOT block — setImmediate is an event loop phase
function safeRecursive() {
  setImmediate(safeRecursive);
}
safeRecursive(); // I/O and timers still run between iterations
```

The official Node.js docs recommend `queueMicrotask()` over `process.nextTick()` for most cases. `process.nextTick()` exists for historical reasons and has a higher priority than Promise microtasks, which can cause surprising ordering.

---

## Gotcha 9: "What happens when you read a file with `await readFile()`?"

```typescript
import { readFile } from "node:fs/promises";

const data = await readFile("large-file.txt");
```

### The Trap

"It reads the file asynchronously, so it's efficient."

### Correct Answer

Yes, the I/O is async, but **the entire file is loaded into memory**. For a 2GB file, this allocates a 2GB Buffer — likely crashing the process.

`readFile` is async (non-blocking) but it's **not streaming**.

Use streams for large files:
```typescript
import { createReadStream } from "node:fs";

// Constant memory regardless of file size
const stream = createReadStream("large-file.txt", {
  highWaterMark: 64 * 1024, // Read 64KB chunks
});

for await (const chunk of stream) {
  // Process chunk — only 64KB in memory at a time
}
```

**Rule of thumb**: Use `readFile` for files < 10MB, streams for anything larger.

---

## Gotcha 10: "Is this code safe in production?"

```typescript
app.get("/image/:id", async (req, res) => {
  const buffer = await sharp(req.params.id)
    .resize(800, 600)
    .toBuffer();
  
  res.end(buffer);
});
```

### The Trap

"Yes — it's async."

### Correct Answer

**No.** Multiple problems:

1. **Unbounded concurrency**: If 1000 requests arrive simultaneously, 1000 sharp operations run at once. Each allocates ~5-50MB for image processing. Combined: potentially 50GB memory usage.

2. **CPU blocking**: `sharp` (libvips) does CPU-intensive work. While the Promise is async, the work happens on the libuv thread pool. With 4 default threads and 1000 pending operations, the thread pool becomes a massive bottleneck — file I/O and DNS resolution stop working.

3. **Path traversal**: `req.params.id` is used directly — an attacker can request `../../etc/passwd`.

Fix:
```typescript
import { Semaphore } from "./semaphore"; // Custom concurrency limiter

const imageSemaphore = new Semaphore(4); // Max 4 concurrent resizes

app.get("/image/:id", async (req, res) => {
  const id = req.params.id.replace(/[^a-zA-Z0-9-]/g, ""); // Sanitize
  
  const release = await imageSemaphore.acquire();
  try {
    const buffer = await sharp(`/uploads/${id}.jpg`)
      .resize(800, 600)
      .toBuffer();
    
    res.setHeader("Content-Type", "image/jpeg");
    res.setHeader("Content-Length", buffer.length);
    res.setHeader("Cache-Control", "public, max-age=86400");
    res.end(buffer);
  } finally {
    release();
  }
});
```

---

## Gotcha 11: "What's wrong with this error handling?"

```typescript
process.on("uncaughtException", (err) => {
  console.error("Error:", err);
  // Keep running
});
```

### The Trap

"It catches all errors and prevents crashes — it's a safety net."

### Correct Answer

**This is dangerous.** After an uncaught exception, the application state is **unreliable**:

- Event handlers may have been partially executed
- Database connections may be in an inconsistent state
- In-memory data structures may be corrupted
- Resource cleanup may not have happened

The Node.js docs explicitly say: *"The correct use of 'uncaughtException' is to perform synchronous cleanup of allocated resources, then exit."*

```typescript
// Correct pattern:
process.on("uncaughtException", (err) => {
  console.error("Uncaught exception — shutting down:", err);
  
  // Synchronous cleanup only
  // Flush logs, close files
  
  process.exit(1); // Always exit
});

// For unhandled Promise rejections (Node 15+: crashes by default)
process.on("unhandledRejection", (reason) => {
  console.error("Unhandled rejection:", reason);
  process.exit(1);
});
```

The real safety net is a **process manager** (PM2, systemd, Kubernetes) that restarts the process.

---

## Gotcha 12: "Does `Promise.all()` run promises in parallel?"

### The Trap

"Yes — it runs all promises at the same time."

### Correct Answer

**`Promise.all()` doesn't run anything.** Promises start executing the moment they're created:

```typescript
// The work starts HERE — when the function is called
const p1 = fetchUser(1);   // Started
const p2 = fetchPosts(1);  // Started
const p3 = fetchComments(1); // Started

// Promise.all just waits for all of them
const results = await Promise.all([p1, p2, p3]);

// This is equivalent:
const results2 = await Promise.all([
  fetchUser(1),      // Created + started
  fetchPosts(1),     // Created + started
  fetchComments(1),  // Created + started
]);
```

`Promise.all()` is a **synchronization** mechanism, not an execution mechanism. It takes Promises that are already running and waits for all to complete. The actual parallelism comes from the fact that the Promises were created (and started their async operations) before `await`.

**Important**: `Promise.all()` rejects on the first failure and does NOT cancel other promises. Use `Promise.allSettled()` when you need all results regardless of individual failures.

---

## Gotcha 13: "What's the difference between `==` and `===` in this case?"

```typescript
null == undefined   // ?
null === undefined  // ?
NaN === NaN         // ?
```

### The Trap

This tests awareness of JavaScript's type coercion rules.

### Correct Answer

```typescript
null == undefined   // true  — spec defines them as loosely equal
null === undefined  // false — different types (object vs undefined)
NaN === NaN         // false — NaN is not equal to anything, including itself
```

Use `Number.isNaN()` instead of `=== NaN`:
```typescript
Number.isNaN(NaN)      // true
Number.isNaN(undefined) // false (unlike global isNaN which returns true!)
```

---

## The Meta-Gotcha: Depth vs. Breadth

In interviews, the most common failing pattern isn't getting specific answers wrong — it's giving **shallow answers that sound rehearsed**.

When asked "How does the event loop work?", don't recite the phase list. Instead:

1. Start with the high-level mental model
2. Pick one detail and go deep (e.g., what happens between phases)
3. Connect it to a real problem you've seen
4. Show awareness of what you don't know

Example: "The event loop processes I/O callbacks, timers, and microtasks in a specific order. What's interesting is that `process.nextTick()` runs between phases and can starve the loop — I found this when a recursive nextTick prevented my health check endpoint from responding."

**This shows understanding, not memorization.**
