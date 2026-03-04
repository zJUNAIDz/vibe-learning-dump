# Lesson 01 — Benchmarking

## Why Most Benchmarks Are Wrong

Common mistakes:
1. **Measuring in dev mode** — V8 optimizes differently under load vs cold start
2. **Not warming up** — first 1000 calls trigger JIT compilation, skewing results
3. **Ignoring GC** — garbage collection pauses hide in averages
4. **Using `Date.now()`** — only millisecond precision; use `performance.now()` (microsecond)
5. **Comparing single runs** — statistical noise requires multiple iterations

---

## Proper Microbenchmarking

```typescript
// benchmark.ts
interface BenchmarkResult {
  name: string;
  iterations: number;
  totalMs: number;
  avgNs: number;
  opsPerSec: number;
  minNs: number;
  maxNs: number;
  p99Ns: number;
}

async function benchmark(
  name: string,
  fn: () => void | Promise<void>,
  options: { warmup?: number; iterations?: number } = {}
): Promise<BenchmarkResult> {
  const { warmup = 1000, iterations = 10_000 } = options;
  
  // Warmup — let V8 JIT compile
  for (let i = 0; i < warmup; i++) {
    await fn();
  }
  
  // Force GC before measurement (run with --expose-gc)
  if (globalThis.gc) globalThis.gc();
  
  const times: number[] = [];
  const totalStart = performance.now();
  
  for (let i = 0; i < iterations; i++) {
    const start = performance.now();
    await fn();
    const elapsed = (performance.now() - start) * 1_000_000; // Convert to ns
    times.push(elapsed);
  }
  
  const totalMs = performance.now() - totalStart;
  
  times.sort((a, b) => a - b);
  
  const sum = times.reduce((a, b) => a + b, 0);
  const avgNs = sum / times.length;
  
  return {
    name,
    iterations,
    totalMs,
    avgNs,
    opsPerSec: Math.round(1_000_000_000 / avgNs),
    minNs: times[0],
    maxNs: times[times.length - 1],
    p99Ns: times[Math.floor(times.length * 0.99)],
  };
}

function printResult(result: BenchmarkResult) {
  console.log(`\n${result.name}`);
  console.log(`  ${result.opsPerSec.toLocaleString()} ops/sec`);
  console.log(`  avg: ${(result.avgNs / 1000).toFixed(2)}μs`);
  console.log(`  min: ${(result.minNs / 1000).toFixed(2)}μs`);
  console.log(`  max: ${(result.maxNs / 1000).toFixed(2)}μs`);
  console.log(`  p99: ${(result.p99Ns / 1000).toFixed(2)}μs`);
}

// --- Compare approaches ---

// Map vs Object for key-value lookup
const map = new Map<string, number>();
const obj: Record<string, number> = {};
for (let i = 0; i < 1000; i++) {
  const key = `key_${i}`;
  map.set(key, i);
  obj[key] = i;
}

printResult(await benchmark("Map.get()", () => {
  map.get("key_500");
}));

printResult(await benchmark("Object property access", () => {
  obj["key_500"];
}));

// JSON.parse vs structured clone
const data = { users: Array.from({ length: 100 }, (_, i) => ({ id: i, name: `user_${i}` })) };
const json = JSON.stringify(data);

printResult(await benchmark("JSON.parse(JSON.stringify())", () => {
  JSON.parse(JSON.stringify(data));
}));

printResult(await benchmark("structuredClone()", () => {
  structuredClone(data);
}));

// String concatenation vs template literals vs array join
const parts = ["hello", "world", "foo", "bar", "baz"];

printResult(await benchmark("String concatenation (+)", () => {
  let s = "";
  for (const p of parts) s += p;
}));

printResult(await benchmark("Template literal", () => {
  `${parts[0]}${parts[1]}${parts[2]}${parts[3]}${parts[4]}`;
}));

printResult(await benchmark("Array.join()", () => {
  parts.join("");
}));
```

---

## HTTP Load Testing with autocannon

```typescript
// load-test.ts
// npm install autocannon
import autocannon from "autocannon";
import http from "node:http";

// Target server
const server = http.createServer((req, res) => {
  // Simulate DB query
  setTimeout(() => {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ status: "ok", timestamp: Date.now() }));
  }, 10);
});

await new Promise<void>((resolve) => server.listen(3001, resolve));

// Run load test
const result = await autocannon({
  url: "http://localhost:3001",
  connections: 100,     // Concurrent connections
  duration: 10,         // Seconds
  pipelining: 1,        // HTTP pipelining factor
});

console.log("\n=== Load Test Results ===");
console.log(`Requests:    ${result.requests.total}`);
console.log(`Throughput:  ${result.throughput.average} bytes/sec`);
console.log(`Latency:`);
console.log(`  avg:       ${result.latency.average}ms`);
console.log(`  p50:       ${result.latency.p50}ms`);
console.log(`  p99:       ${result.latency.p99}ms`);
console.log(`  max:       ${result.latency.max}ms`);
console.log(`Errors:      ${result.errors}`);
console.log(`Timeouts:    ${result.timeouts}`);

server.close();
```

---

## Built-in performance.mark / performance.measure

```typescript
// perf-marks.ts
import { performance, PerformanceObserver } from "node:perf_hooks";

// Set up observer to collect measurements
const observer = new PerformanceObserver((list) => {
  for (const entry of list.getEntries()) {
    console.log(`${entry.name}: ${entry.duration.toFixed(2)}ms`);
  }
});
observer.observe({ entryTypes: ["measure"] });

// Mark start/end of operations
performance.mark("db-query-start");
await new Promise((r) => setTimeout(r, 50)); // Simulate DB
performance.mark("db-query-end");
performance.measure("DB Query", "db-query-start", "db-query-end");

performance.mark("render-start");
JSON.stringify(Array.from({ length: 10_000 }, (_, i) => ({ id: i })));
performance.mark("render-end");
performance.measure("JSON Render", "render-start", "render-end");

// Built-in Node.js timing
performance.mark("import-start");
await import("node:crypto");
performance.mark("import-end");
performance.measure("Import crypto", "import-start", "import-end");

observer.disconnect();
```

---

## Interview Questions

### Q1: "How do you benchmark Node.js code properly?"

**Answer**: Five requirements for valid benchmarks:
1. **Warmup phase** (1000+ iterations) — let V8's JIT optimizer (TurboFan) compile hot functions before measuring
2. **Force GC before measurement** — run with `--expose-gc` and call `gc()` to eliminate GC noise
3. **Use `performance.now()`** — microsecond precision vs `Date.now()`'s millisecond precision
4. **Measure percentiles**, not just averages — p99 reveals GC pauses and tail latency that averages hide
5. **Multiple runs** with statistical analysis — compute standard deviation, discard outliers

### Q2: "What's the difference between microbenchmarks and load tests?"

**Answer**:
- **Microbenchmark**: Measures a single operation in isolation (e.g., "how fast is `JSON.parse`?"). Useful for comparing alternatives but doesn't reflect real-world behavior (no I/O, no concurrency, no memory pressure).
- **Load test**: Sends concurrent HTTP requests to measure end-to-end system behavior under realistic conditions. Reveals: connection handling overhead, event loop saturation, memory leaks under load, GC pressure, OS-level limits (file descriptors, TCP backlog).

Always optimize based on load test results, not microbenchmarks. A function that's 10x faster in isolation might not matter if it only accounts for 1% of request time.
