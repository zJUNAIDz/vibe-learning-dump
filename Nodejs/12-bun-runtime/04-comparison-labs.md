# Lesson 04 — Performance Comparison Labs

## Lab 1: HTTP Server Throughput

```typescript
// http-benchmark.ts
// Run this file with both: node and bun, then compare

import http from "node:http";

let requestCount = 0;

const server = http.createServer((req, res) => {
  requestCount++;
  
  const body = JSON.stringify({
    message: "Hello, World!",
    count: requestCount,
    timestamp: Date.now(),
  });
  
  res.writeHead(200, {
    "Content-Type": "application/json",
    "Content-Length": Buffer.byteLength(body),
  });
  res.end(body);
});

server.listen(3000, () => {
  console.log("Server on :3000 — benchmark with:");
  console.log("  wrk -t4 -c100 -d10s http://localhost:3000/");
  console.log("  # or: autocannon -c100 -d10 http://localhost:3000/");
});

// Typical results:
// Node.js:  ~40,000-60,000 req/sec
// Bun:      ~90,000-150,000 req/sec
// Bun with Bun.serve: ~200,000-300,000 req/sec
```

---

## Lab 2: File I/O Benchmark

```typescript
// file-io-benchmark.ts
import { writeFileSync, unlinkSync, mkdirSync, rmSync } from "node:fs";
import { readFile, writeFile } from "node:fs/promises";

const BENCH_DIR = "/tmp/bench-files";
const FILE_COUNT = 500;
const FILE_SIZE_KB = 10;

// Setup
try { mkdirSync(BENCH_DIR, { recursive: true }); } catch {}
for (let i = 0; i < FILE_COUNT; i++) {
  writeFileSync(`${BENCH_DIR}/file-${i}.txt`, "x".repeat(FILE_SIZE_KB * 1024));
}

// Benchmark: Sequential reads
console.log(`\n--- Sequential Read (${FILE_COUNT} × ${FILE_SIZE_KB}KB) ---`);
{
  const start = performance.now();
  for (let i = 0; i < FILE_COUNT; i++) {
    await readFile(`${BENCH_DIR}/file-${i}.txt`);
  }
  const elapsed = performance.now() - start;
  console.log(`Time: ${elapsed.toFixed(1)}ms`);
  console.log(`Rate: ${(FILE_COUNT / (elapsed / 1000)).toFixed(0)} files/sec`);
}

// Benchmark: Concurrent reads
console.log(`\n--- Concurrent Read (${FILE_COUNT} × ${FILE_SIZE_KB}KB) ---`);
{
  const start = performance.now();
  await Promise.all(
    Array.from({ length: FILE_COUNT }, (_, i) =>
      readFile(`${BENCH_DIR}/file-${i}.txt`)
    )
  );
  const elapsed = performance.now() - start;
  console.log(`Time: ${elapsed.toFixed(1)}ms`);
  console.log(`Rate: ${(FILE_COUNT / (elapsed / 1000)).toFixed(0)} files/sec`);
}

// Benchmark: Sequential writes
console.log(`\n--- Sequential Write (${FILE_COUNT} × ${FILE_SIZE_KB}KB) ---`);
{
  const data = Buffer.alloc(FILE_SIZE_KB * 1024, "y");
  const start = performance.now();
  for (let i = 0; i < FILE_COUNT; i++) {
    await writeFile(`${BENCH_DIR}/out-${i}.txt`, data);
  }
  const elapsed = performance.now() - start;
  console.log(`Time: ${elapsed.toFixed(1)}ms`);
  console.log(`Rate: ${(FILE_COUNT / (elapsed / 1000)).toFixed(0)} files/sec`);
}

// Benchmark: Concurrent writes
console.log(`\n--- Concurrent Write (${FILE_COUNT} × ${FILE_SIZE_KB}KB) ---`);
{
  const data = Buffer.alloc(FILE_SIZE_KB * 1024, "z");
  const start = performance.now();
  await Promise.all(
    Array.from({ length: FILE_COUNT }, (_, i) =>
      writeFile(`${BENCH_DIR}/cout-${i}.txt`, data)
    )
  );
  const elapsed = performance.now() - start;
  console.log(`Time: ${elapsed.toFixed(1)}ms`);
  console.log(`Rate: ${(FILE_COUNT / (elapsed / 1000)).toFixed(0)} files/sec`);
}

// Cleanup
rmSync(BENCH_DIR, { recursive: true, force: true });

// Expected:
// Concurrent reads — Bun ~3-5x faster (io_uring vs libuv thread pool)
// Sequential reads — Bun ~1.2-2x faster
// Writes — Similar, OS buffer cache dominates
```

---

## Lab 3: JSON Serialization

```typescript
// json-benchmark.ts

const ITERATIONS = 100_000;

// Small object
const small = { id: 1, name: "Alice", active: true };

// Medium object (100 properties)
const medium: Record<string, any> = {};
for (let i = 0; i < 100; i++) {
  medium[`field_${i}`] = {
    value: Math.random(),
    label: `Label ${i}`,
    tags: [`tag_${i}`, `tag_${i + 1}`],
  };
}

// Large object (array of 1000 items)
const large = Array.from({ length: 1000 }, (_, i) => ({
  id: i,
  name: `User ${i}`,
  email: `user${i}@example.com`,
  scores: Array.from({ length: 10 }, () => Math.random() * 100),
}));

function benchJsonOps(name: string, obj: any, iterations: number) {
  // Stringify
  const start1 = performance.now();
  let str = "";
  for (let i = 0; i < iterations; i++) {
    str = JSON.stringify(obj);
  }
  const stringifyTime = performance.now() - start1;
  
  // Parse
  const start2 = performance.now();
  for (let i = 0; i < iterations; i++) {
    JSON.parse(str);
  }
  const parseTime = performance.now() - start2;
  
  console.log(`${name}:`);
  console.log(`  stringify: ${stringifyTime.toFixed(1)}ms (${(iterations / (stringifyTime / 1000)).toFixed(0)} ops/sec)`);
  console.log(`  parse:     ${parseTime.toFixed(1)}ms (${(iterations / (parseTime / 1000)).toFixed(0)} ops/sec)`);
  console.log(`  payload:   ${(Buffer.byteLength(str) / 1024).toFixed(1)}KB`);
}

benchJsonOps("Small object", small, ITERATIONS);
benchJsonOps("Medium object (100 props)", medium, ITERATIONS / 10);
benchJsonOps("Large array (1000 items)", large, ITERATIONS / 100);

// Bun uses SIMD-optimized JSON parsing (simdjson concepts in Zig)
// Typically ~1.5-3x faster for parsing large JSON
// Stringify speed is more comparable
```

---

## Lab 4: Startup Time

```bash
#!/bin/bash
# startup-benchmark.sh

echo "=== Startup Time Benchmark ==="
echo ""

# Empty script
echo 'console.log("ok")' > /tmp/bench-startup.ts

echo "--- Empty script ---"
echo -n "Node: "
time node --experimental-strip-types /tmp/bench-startup.ts 2>/dev/null

echo -n "Bun:  "
time bun /tmp/bench-startup.ts 2>/dev/null

# Script with imports
cat > /tmp/bench-imports.ts << 'EOF'
import http from "node:http";
import { readFileSync } from "node:fs";
import { createHash } from "node:crypto";
import path from "node:path";
import { EventEmitter } from "node:events";
console.log("imported 5 modules");
EOF

echo ""
echo "--- 5 module imports ---"
echo -n "Node: "
time node --experimental-strip-types /tmp/bench-imports.ts 2>/dev/null

echo -n "Bun:  "
time bun /tmp/bench-imports.ts 2>/dev/null

# Typical results:
# Empty:        Node ~40ms,  Bun ~10ms  (4x)
# 5 imports:    Node ~65ms,  Bun ~15ms  (4x)
# 20 imports:   Node ~120ms, Bun ~25ms  (5x)
```

---

## Lab 5: Summary Comparison Table

Run all benchmarks and fill in results:

```typescript
// summary.ts

interface BenchResult {
  test: string;
  nodeResult: string;
  bunResult: string;
  winner: "Node" | "Bun" | "Tie";
  factor: string;
}

const results: BenchResult[] = [
  {
    test: "HTTP req/sec (http module)",
    nodeResult: "~50,000",
    bunResult: "~120,000",
    winner: "Bun",
    factor: "~2.4x",
  },
  {
    test: "HTTP req/sec (Bun.serve)",
    nodeResult: "N/A",
    bunResult: "~250,000",
    winner: "Bun",
    factor: "~5x vs Node http",
  },
  {
    test: "Concurrent file reads (500)",
    nodeResult: "~45ms",
    bunResult: "~12ms",
    winner: "Bun",
    factor: "~3.7x",
  },
  {
    test: "JSON parse (large)",
    nodeResult: "~8ms",
    bunResult: "~3ms",
    winner: "Bun",
    factor: "~2.7x",
  },
  {
    test: "Startup time (5 imports)",
    nodeResult: "~65ms",
    bunResult: "~15ms",
    winner: "Bun",
    factor: "~4.3x",
  },
  {
    test: "CPU-bound (fibonacci)",
    nodeResult: "~850ms",
    bunResult: "~870ms",
    winner: "Tie",
    factor: "~1x",
  },
  {
    test: "Ecosystem compatibility",
    nodeResult: "100%",
    bunResult: "~95%",
    winner: "Node",
    factor: "N/A",
  },
  {
    test: "Production battle-testing",
    nodeResult: "15+ years",
    bunResult: "~2 years",
    winner: "Node",
    factor: "N/A",
  },
];

console.log("| Test | Node.js | Bun | Winner | Factor |");
console.log("|------|---------|-----|--------|--------|");
for (const r of results) {
  console.log(`| ${r.test} | ${r.nodeResult} | ${r.bunResult} | ${r.winner} | ${r.factor} |`);
}
```

---

## Interview Questions

### Q1: "If Bun is faster, why hasn't everyone switched?"

**Answer**: Performance isn't the only factor:
1. **Ecosystem compatibility**: ~5% of npm packages have issues in Bun (native addons, V8-specific code, edge cases in stream behavior)
2. **Production track record**: Node.js has 15+ years of production battle-testing. Known failure modes, known workarounds.
3. **Tooling**: Node.js has better debugging (Chrome DevTools, `--inspect`, V8 profiler, heap snapshots). Bun's debugging story is catching up.
4. **Team expertise**: Most teams know Node.js internals. Bun's internals (JSC, Zig, io_uring) are less familiar.
5. **CPU-bound parity**: For compute, V8 and JSC perform similarly. Bun's advantages are in I/O and startup.

### Q2: "Where does Bun NOT outperform Node.js?"

**Answer**:
- **CPU-bound computation**: V8's TurboFan and JSC's FTL produce comparable machine code. Fibonacci, sorting, regex — results are within 10%.
- **Long-running servers**: Startup time advantage is amortized to zero for servers running for hours/days.
- **Native addon workloads**: Code that spends most time in C++ addons (image processing, ML inference) — runtime doesn't matter.
- **macOS I/O**: Without io_uring, Bun falls back to kqueue, which is similar to Node's approach.

### Q3: "How would you benchmark Node.js vs Bun fairly?"

**Answer**:
1. **Same code**: Use `node:` module imports that both runtimes support
2. **Same machine**: Run benchmarks on the same hardware, same OS, same conditions
3. **Warmup**: Run each benchmark 3+ times before measuring (JIT warmup)
4. **Multiple runs**: Take median of 10+ runs, report standard deviation
5. **Realistic workload**: Don't benchmark `1 + 1`. Test your actual application patterns (HTTP handling, DB queries, JSON serialization)
6. **Pin versions**: Specify exact runtime versions — both are actively improving
7. **Disable turbo boost**: CPU frequency scaling adds noise to benchmarks
