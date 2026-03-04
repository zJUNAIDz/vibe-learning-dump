# Lesson 04 — Filesystem Labs

## Lab 1: Build a File Watcher with Debouncing

File watchers are used in dev tools (hot reload), build systems, and log processors. Native `fs.watch()` fires multiple events per save — your watcher must debounce.

```typescript
// file-watcher.ts
import { watch, statSync, readFileSync } from "node:fs";
import { EventEmitter } from "node:events";

interface FileChange {
  path: string;
  type: "created" | "modified" | "deleted";
  size: number;
  timestamp: number;
}

class FileWatcher extends EventEmitter {
  private watchers = new Map<string, ReturnType<typeof watch>>();
  private debounceTimers = new Map<string, ReturnType<typeof setTimeout>>();
  private lastSizes = new Map<string, number>();
  private debounceMs: number;

  constructor(debounceMs = 100) {
    super();
    this.debounceMs = debounceMs;
  }

  watch(filePath: string): void {
    if (this.watchers.has(filePath)) return;

    try {
      this.lastSizes.set(filePath, statSync(filePath).size);
    } catch {
      this.lastSizes.set(filePath, 0);
    }

    const watcher = watch(filePath, (eventType) => {
      // Debounce: editors save in multiple steps
      const existing = this.debounceTimers.get(filePath);
      if (existing) clearTimeout(existing);

      this.debounceTimers.set(
        filePath,
        setTimeout(() => {
          this.handleChange(filePath);
          this.debounceTimers.delete(filePath);
        }, this.debounceMs)
      );
    });

    watcher.on("error", (err) => {
      this.emit("error", { path: filePath, error: err });
    });

    this.watchers.set(filePath, watcher);
    console.log(`Watching: ${filePath}`);
  }

  private handleChange(filePath: string): void {
    try {
      const stat = statSync(filePath);
      const previousSize = this.lastSizes.get(filePath) ?? 0;

      const change: FileChange = {
        path: filePath,
        type: stat.size === 0 && previousSize > 0 ? "created" : "modified",
        size: stat.size,
        timestamp: Date.now(),
      };

      this.lastSizes.set(filePath, stat.size);
      this.emit("change", change);
    } catch {
      const change: FileChange = {
        path: filePath,
        type: "deleted",
        size: 0,
        timestamp: Date.now(),
      };
      this.emit("change", change);
    }
  }

  unwatch(filePath: string): void {
    const watcher = this.watchers.get(filePath);
    if (watcher) {
      watcher.close();
      this.watchers.delete(filePath);
      this.debounceTimers.get(filePath) &&
        clearTimeout(this.debounceTimers.get(filePath)!);
      this.debounceTimers.delete(filePath);
      this.lastSizes.delete(filePath);
    }
  }

  close(): void {
    for (const [path] of this.watchers) {
      this.unwatch(path);
    }
  }
}

// Usage
import { writeFileSync, unlinkSync } from "node:fs";

const watcher = new FileWatcher(150);

watcher.on("change", (change: FileChange) => {
  console.log(`[${change.type.toUpperCase()}] ${change.path} (${change.size} bytes)`);
});

const testFile = "/tmp/watched-file.txt";
writeFileSync(testFile, "initial");
watcher.watch(testFile);

// Simulate changes
setTimeout(() => writeFileSync(testFile, "updated content"), 500);
setTimeout(() => writeFileSync(testFile, "another update with more data"), 1000);
setTimeout(() => {
  watcher.close();
  console.log("Watcher closed");
  unlinkSync(testFile);
}, 2000);
```

---

## Lab 2: Streaming CSV Parser

Process a CSV file of any size with constant memory. Uses backpressure properly.

```typescript
// csv-parser.ts
import { createReadStream, writeFileSync } from "node:fs";
import { Transform, Writable } from "node:stream";
import { pipeline } from "node:stream/promises";

class CSVParser extends Transform {
  private remainder = "";
  private headers: string[] = [];
  private lineNumber = 0;

  constructor() {
    super({ objectMode: true }); // Output objects, not buffers
  }

  _transform(chunk: Buffer, encoding: string, callback: Function) {
    const data = this.remainder + chunk.toString("utf8");
    const lines = data.split("\n");

    // Last element might be incomplete — save for next chunk
    this.remainder = lines.pop() ?? "";

    for (const line of lines) {
      if (line.trim() === "") continue;

      this.lineNumber++;

      if (this.lineNumber === 1) {
        // First line = headers
        this.headers = this.parseLine(line);
        continue;
      }

      const values = this.parseLine(line);
      const record: Record<string, string> = {};

      for (let i = 0; i < this.headers.length; i++) {
        record[this.headers[i]] = values[i] ?? "";
      }

      this.push(record);
    }

    callback();
  }

  _flush(callback: Function) {
    // Handle any remaining data
    if (this.remainder.trim()) {
      const values = this.parseLine(this.remainder);
      const record: Record<string, string> = {};
      for (let i = 0; i < this.headers.length; i++) {
        record[this.headers[i]] = values[i] ?? "";
      }
      this.push(record);
    }
    callback();
  }

  private parseLine(line: string): string[] {
    // Simple CSV parse (handles quoted fields)
    const result: string[] = [];
    let current = "";
    let inQuotes = false;

    for (const char of line) {
      if (char === '"') {
        inQuotes = !inQuotes;
      } else if (char === "," && !inQuotes) {
        result.push(current.trim());
        current = "";
      } else {
        current += char;
      }
    }
    result.push(current.trim());
    return result;
  }
}

// Generate test CSV
const headers = "id,name,email,score";
const rows = Array.from(
  { length: 50_000 },
  (_, i) => `${i},"User ${i}",user${i}@example.com,${Math.floor(Math.random() * 100)}`
);
writeFileSync("/tmp/data.csv", [headers, ...rows].join("\n"));

// Process it
let count = 0;
let totalScore = 0;

const aggregator = new Writable({
  objectMode: true,
  write(record: Record<string, string>, encoding, callback) {
    count++;
    totalScore += parseInt(record.score, 10);
    callback();
  },
});

const start = performance.now();

await pipeline(
  createReadStream("/tmp/data.csv"),
  new CSVParser(),
  aggregator
);

const elapsed = performance.now() - start;

console.log(`Processed ${count} records in ${elapsed.toFixed(1)}ms`);
console.log(`Average score: ${(totalScore / count).toFixed(1)}`);
console.log(`Memory: ${(process.memoryUsage().heapUsed / 1024 / 1024).toFixed(1)}MB`);
```

---

## Lab 3: Build a Write-Ahead Log (WAL)

Production databases use WALs for crash recovery. Build a simplified version.

```typescript
// wal.ts
import {
  openSync, writeSync, readFileSync, closeSync,
  existsSync, renameSync, unlinkSync, appendFileSync, fsyncSync,
} from "node:fs";

interface WALEntry {
  seq: number;
  timestamp: number;
  op: "SET" | "DELETE";
  key: string;
  value?: string;
}

class WriteAheadLog {
  private fd: number;
  private sequence = 0;
  private filePath: string;
  private store = new Map<string, string>();

  constructor(filePath: string) {
    this.filePath = filePath;

    // Recover from existing WAL
    if (existsSync(filePath)) {
      this.recover();
    }

    // Open for appending
    this.fd = openSync(filePath, "a");
  }

  private recover(): void {
    console.log("Recovering from WAL...");
    const content = readFileSync(this.filePath, "utf8");
    const lines = content.split("\n").filter((l) => l.trim());

    let recovered = 0;
    for (const line of lines) {
      try {
        const entry: WALEntry = JSON.parse(line);
        this.applyEntry(entry);
        this.sequence = Math.max(this.sequence, entry.seq);
        recovered++;
      } catch {
        console.warn(`Corrupt WAL entry skipped: ${line.slice(0, 50)}`);
      }
    }
    console.log(`Recovered ${recovered} entries, last seq: ${this.sequence}`);
  }

  private applyEntry(entry: WALEntry): void {
    if (entry.op === "SET") {
      this.store.set(entry.key, entry.value!);
    } else if (entry.op === "DELETE") {
      this.store.delete(entry.key);
    }
  }

  set(key: string, value: string): number {
    const entry: WALEntry = {
      seq: ++this.sequence,
      timestamp: Date.now(),
      op: "SET",
      key,
      value,
    };

    // Write to WAL FIRST (write-ahead)
    const line = JSON.stringify(entry) + "\n";
    writeSync(this.fd, line);
    fsyncSync(this.fd); // Force to disk — crash safe

    // Then apply to in-memory store
    this.applyEntry(entry);

    return entry.seq;
  }

  delete(key: string): number {
    const entry: WALEntry = {
      seq: ++this.sequence,
      timestamp: Date.now(),
      op: "DELETE",
      key,
    };

    const line = JSON.stringify(entry) + "\n";
    writeSync(this.fd, line);
    fsyncSync(this.fd);

    this.applyEntry(entry);
    return entry.seq;
  }

  get(key: string): string | undefined {
    return this.store.get(key);
  }

  // Compact WAL: write current state as a snapshot
  compact(): void {
    const snapshotPath = this.filePath + ".snapshot";

    // Write current state as fresh WAL
    let newSeq = 0;
    const lines: string[] = [];
    for (const [key, value] of this.store) {
      const entry: WALEntry = {
        seq: ++newSeq,
        timestamp: Date.now(),
        op: "SET",
        key,
        value,
      };
      lines.push(JSON.stringify(entry));
    }

    // Write snapshot atomically
    appendFileSync(snapshotPath, lines.join("\n") + "\n");

    // Close old WAL, replace with snapshot
    closeSync(this.fd);
    renameSync(snapshotPath, this.filePath); // Atomic on POSIX
    this.fd = openSync(this.filePath, "a");
    this.sequence = newSeq;

    console.log(`Compacted WAL: ${this.store.size} entries`);
  }

  close(): void {
    closeSync(this.fd);
  }

  get size(): number {
    return this.store.size;
  }
}

// Demo
const wal = new WriteAheadLog("/tmp/test.wal");

// Write some data
wal.set("user:1", JSON.stringify({ name: "Alice", role: "admin" }));
wal.set("user:2", JSON.stringify({ name: "Bob", role: "user" }));
wal.set("config:theme", "dark");
wal.delete("user:2");
wal.set("user:1", JSON.stringify({ name: "Alice", role: "superadmin" }));

console.log(`Store size: ${wal.size}`);
console.log(`user:1 = ${wal.get("user:1")}`);
console.log(`user:2 = ${wal.get("user:2")}`); // undefined (deleted)

wal.compact();
wal.close();

// Simulate crash recovery
console.log("\n--- Simulating restart ---");
const wal2 = new WriteAheadLog("/tmp/test.wal");
console.log(`user:1 after recovery = ${wal2.get("user:1")}`);
console.log(`Store size after recovery: ${wal2.size}`);
wal2.close();

// Cleanup
unlinkSync("/tmp/test.wal");
```

---

## Lab 4: Memory-Mapped File Performance Test

Compare different file access strategies and measure their performance characteristics.

```typescript
// file-perf-benchmark.ts
import {
  openSync, readSync, closeSync, writeFileSync, readFileSync,
  createReadStream, statSync,
} from "node:fs";
import { readFile } from "node:fs/promises";

const FILE_PATH = "/tmp/bench-file.bin";
const FILE_SIZE = 50 * 1024 * 1024; // 50MB

// Setup
writeFileSync(FILE_PATH, Buffer.alloc(FILE_SIZE, 0x42));

async function benchReadFile(): Promise<number> {
  const start = performance.now();
  const data = await readFile(FILE_PATH);
  const elapsed = performance.now() - start;
  
  // Touch every byte to ensure it's actually read
  let sum = 0;
  for (let i = 0; i < data.length; i += 4096) sum += data[i];
  
  return elapsed;
}

async function benchReadStream(chunkSize: number): Promise<number> {
  return new Promise((resolve) => {
    const start = performance.now();
    let sum = 0;
    
    const stream = createReadStream(FILE_PATH, {
      highWaterMark: chunkSize,
    });
    
    stream.on("data", (chunk: Buffer) => {
      for (let i = 0; i < chunk.length; i += 4096) sum += chunk[i];
    });
    
    stream.on("end", () => {
      resolve(performance.now() - start);
    });
  });
}

function benchRandomRead(reads: number): number {
  const fd = openSync(FILE_PATH, "r");
  const buf = Buffer.alloc(4096);
  const fileSize = statSync(FILE_PATH).size;
  
  const start = performance.now();
  
  for (let i = 0; i < reads; i++) {
    const offset = Math.floor(Math.random() * (fileSize - 4096));
    readSync(fd, buf, 0, 4096, offset);
  }
  
  closeSync(fd);
  return performance.now() - start;
}

function benchSequentialRead(): number {
  const fd = openSync(FILE_PATH, "r");
  const buf = Buffer.alloc(64 * 1024);
  const fileSize = statSync(FILE_PATH).size;
  
  const start = performance.now();
  let offset = 0;
  
  while (offset < fileSize) {
    const bytesRead = readSync(fd, buf, 0, buf.length, offset);
    offset += bytesRead;
    if (bytesRead === 0) break;
  }
  
  closeSync(fd);
  return performance.now() - start;
}

// Run benchmarks
console.log(`File size: ${FILE_SIZE / 1024 / 1024}MB\n`);

console.log("--- Sequential Full Read ---");
const readFileTime = await benchReadFile();
console.log(`  readFile():          ${readFileTime.toFixed(1)}ms`);

const seqTime = benchSequentialRead();
console.log(`  Sequential read():   ${seqTime.toFixed(1)}ms`);

const stream16 = await benchReadStream(16 * 1024);
console.log(`  ReadStream 16KB:     ${stream16.toFixed(1)}ms`);

const stream64 = await benchReadStream(64 * 1024);
console.log(`  ReadStream 64KB:     ${stream64.toFixed(1)}ms`);

const stream256 = await benchReadStream(256 * 1024);
console.log(`  ReadStream 256KB:    ${stream256.toFixed(1)}ms`);

console.log("\n--- Random Access (1000 reads × 4KB) ---");
const randomTime = benchRandomRead(1000);
console.log(`  Random read():       ${randomTime.toFixed(1)}ms`);
console.log(`  Avg per read:        ${(randomTime / 1000).toFixed(3)}ms`);

// Cleanup
import { unlinkSync } from "node:fs";
unlinkSync(FILE_PATH);
```

---

## Challenges

1. **Extend the WAL**: Add a `snapshot()` method that writes the current state as a binary file for faster recovery, falling back to the WAL for entries after the snapshot.

2. **Build a file differ**: Write a stream-based tool that compares two files line-by-line, outputting the differences — handling files too large for memory.

3. **Implement a ring buffer log**: Write a fixed-size log file that wraps around when full (like a circular buffer), always keeping the most recent N entries.
