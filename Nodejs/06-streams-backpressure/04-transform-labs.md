# Lesson 04 — Transform & Pipeline Labs

## Lab 1: Streaming JSON Line Processor

Process newline-delimited JSON (NDJSON/JSONL) files of any size with constant memory.

```typescript
// ndjson-processor.ts
import { Transform, Writable } from "node:stream";
import { pipeline } from "node:stream/promises";
import { createReadStream, createWriteStream, writeFileSync } from "node:fs";

// Generate test NDJSON
const records = Array.from({ length: 50_000 }, (_, i) => 
  JSON.stringify({
    id: i,
    name: `User ${i}`,
    score: Math.random() * 100,
    active: Math.random() > 0.3,
    timestamp: Date.now() - Math.floor(Math.random() * 86400000),
  })
).join("\n");
writeFileSync("/tmp/data.ndjson", records);

// --- Transform: Parse NDJSON ---
class NDJSONParser extends Transform {
  private remainder = "";

  constructor() {
    super({ readableObjectMode: true });
  }

  _transform(chunk: Buffer, encoding: string, callback: Function) {
    const data = this.remainder + chunk.toString();
    const lines = data.split("\n");
    this.remainder = lines.pop() ?? "";

    for (const line of lines) {
      if (line.trim()) {
        try {
          this.push(JSON.parse(line));
        } catch {
          // Skip malformed lines
        }
      }
    }
    callback();
  }

  _flush(callback: Function) {
    if (this.remainder.trim()) {
      try { this.push(JSON.parse(this.remainder)); } catch {}
    }
    callback();
  }
}

// --- Transform: Filter active users with high scores ---
class Filter extends Transform {
  constructor(private predicate: (obj: any) => boolean) {
    super({ objectMode: true });
  }

  _transform(obj: any, encoding: string, callback: Function) {
    if (this.predicate(obj)) this.push(obj);
    callback();
  }
}

// --- Transform: Aggregate stats ---
class Aggregator extends Writable {
  public count = 0;
  public sum = 0;
  public min = Infinity;
  public max = -Infinity;

  constructor(private field: string) {
    super({ objectMode: true });
  }

  _write(obj: any, encoding: string, callback: Function) {
    const value = obj[this.field];
    if (typeof value === "number") {
      this.count++;
      this.sum += value;
      if (value < this.min) this.min = value;
      if (value > this.max) this.max = value;
    }
    callback();
  }

  get average() { return this.sum / this.count; }
}

// Pipeline it all
const aggregator = new Aggregator("score");

await pipeline(
  createReadStream("/tmp/data.ndjson"),
  new NDJSONParser(),
  new Filter((obj) => obj.active && obj.score > 50),
  aggregator,
);

console.log("Active users with score > 50:");
console.log(`  Count: ${aggregator.count}`);
console.log(`  Avg score: ${aggregator.average.toFixed(1)}`);
console.log(`  Min: ${aggregator.min.toFixed(1)}, Max: ${aggregator.max.toFixed(1)}`);
```

---

## Lab 2: Streaming File Compressor

```typescript
// streaming-compressor.ts
import { createReadStream, createWriteStream, writeFileSync, statSync } from "node:fs";
import { pipeline } from "node:stream/promises";
import { createGzip, createBrotliCompress, constants } from "node:zlib";
import { Transform } from "node:stream";

// Create a test file with repetitive data (compresses well)
const testData = Array.from({ length: 100_000 }, (_, i) =>
  `[${new Date().toISOString()}] INFO request_id=${i} method=GET path=/api/users status=200 duration=${Math.floor(Math.random() * 500)}ms`
).join("\n");
writeFileSync("/tmp/compress-test.log", testData);

// Progress reporter transform
class ProgressReporter extends Transform {
  private bytesProcessed = 0;
  private totalBytes: number;
  private lastReport = 0;

  constructor(totalBytes: number) {
    super();
    this.totalBytes = totalBytes;
  }

  _transform(chunk: Buffer, encoding: string, callback: Function) {
    this.bytesProcessed += chunk.length;
    const percent = (this.bytesProcessed / this.totalBytes * 100);
    
    if (percent - this.lastReport >= 20) {
      this.lastReport = Math.floor(percent);
      process.stdout.write(`\r  Progress: ${percent.toFixed(0)}%`);
    }
    
    this.push(chunk);
    callback();
  }

  _flush(callback: Function) {
    process.stdout.write(`\r  Progress: 100%\n`);
    callback();
  }
}

const inputSize = statSync("/tmp/compress-test.log").size;
console.log(`Original size: ${(inputSize / 1024).toFixed(0)}KB\n`);

// Gzip compression
console.log("Gzip:");
const gzipStart = performance.now();
await pipeline(
  createReadStream("/tmp/compress-test.log"),
  new ProgressReporter(inputSize),
  createGzip({ level: 6 }),
  createWriteStream("/tmp/compress-test.log.gz"),
);
const gzipSize = statSync("/tmp/compress-test.log.gz").size;
console.log(`  Compressed: ${(gzipSize / 1024).toFixed(0)}KB (${(gzipSize / inputSize * 100).toFixed(1)}%)`);
console.log(`  Time: ${(performance.now() - gzipStart).toFixed(0)}ms`);

// Brotli compression (better ratio, slower)
console.log("\nBrotli:");
const brotliStart = performance.now();
await pipeline(
  createReadStream("/tmp/compress-test.log"),
  new ProgressReporter(inputSize),
  createBrotliCompress({
    params: { [constants.BROTLI_PARAM_QUALITY]: 6 },
  }),
  createWriteStream("/tmp/compress-test.log.br"),
);
const brotliSize = statSync("/tmp/compress-test.log.br").size;
console.log(`  Compressed: ${(brotliSize / 1024).toFixed(0)}KB (${(brotliSize / inputSize * 100).toFixed(1)}%)`);
console.log(`  Time: ${(performance.now() - brotliStart).toFixed(0)}ms`);
```

---

## Lab 3: HTTP Streaming Proxy

```typescript
// streaming-proxy.ts
import { createServer, request as httpRequest, IncomingMessage } from "node:http";
import { pipeline } from "node:stream/promises";
import { Transform } from "node:stream";

// A streaming proxy that transforms responses in-flight
// Memory stays constant regardless of response size

class ResponseTransform extends Transform {
  _transform(chunk: Buffer, encoding: string, callback: Function) {
    // Example: uppercase all text responses
    const text = chunk.toString();
    this.push(text.toUpperCase());
    callback();
  }
}

const proxy = createServer(async (clientReq, clientRes) => {
  const targetUrl = new URL(clientReq.url ?? "/", "http://localhost:3001");
  
  console.log(`Proxying: ${clientReq.method} ${targetUrl.pathname}`);
  
  const proxyReq = httpRequest(
    {
      hostname: targetUrl.hostname,
      port: targetUrl.port,
      path: targetUrl.pathname + targetUrl.search,
      method: clientReq.method,
      headers: { ...clientReq.headers, host: targetUrl.host },
    },
    async (proxyRes: IncomingMessage) => {
      clientRes.writeHead(proxyRes.statusCode ?? 200, proxyRes.headers);
      
      const isText = proxyRes.headers["content-type"]?.includes("text");
      
      try {
        if (isText) {
          await pipeline(proxyRes, new ResponseTransform(), clientRes);
        } else {
          await pipeline(proxyRes, clientRes);
        }
      } catch {
        // Client disconnected
      }
    }
  );
  
  // Stream request body to target (for POST/PUT)
  try {
    await pipeline(clientReq, proxyReq);
  } catch {
    clientRes.writeHead(502);
    clientRes.end("Bad Gateway");
  }
});

// Backend server
const backend = createServer((req, res) => {
  res.writeHead(200, { "Content-Type": "text/plain" });
  res.end(`Hello from backend! Path: ${req.url}\n`);
});

backend.listen(3001, () => {
  proxy.listen(3000, () => {
    console.log("Proxy on :3000 → Backend on :3001");
    console.log("Test: curl http://localhost:3000/test");
  });
});
```

---

## Lab 4: Build a Streaming ETL Pipeline

```typescript
// etl-pipeline.ts
import { Readable, Transform, Writable } from "node:stream";
import { pipeline } from "node:stream/promises";

// Extract → Transform → Load pipeline processing user events

interface RawEvent {
  userId: string;
  action: string;
  timestamp: number;
  metadata: Record<string, any>;
}

interface EnrichedEvent extends RawEvent {
  date: string;
  hour: number;
  isWeekend: boolean;
  actionCategory: string;
}

interface AggregatedStats {
  totalEvents: number;
  uniqueUsers: Set<string>;
  actionCounts: Map<string, number>;
  hourlyDistribution: number[];
}

// Extract: Generate synthetic events
function createEventSource(count: number): Readable {
  const actions = ["page_view", "click", "purchase", "signup", "logout", "search", "add_to_cart"];
  let i = 0;

  return new Readable({
    objectMode: true,
    read() {
      if (i >= count) { this.push(null); return; }
      
      const event: RawEvent = {
        userId: `user_${Math.floor(Math.random() * 1000)}`,
        action: actions[Math.floor(Math.random() * actions.length)],
        timestamp: Date.now() - Math.floor(Math.random() * 7 * 86400000),
        metadata: { source: "web", version: "2.1" },
      };
      
      i++;
      this.push(event);
    },
  });
}

// Transform: Enrich events
class EventEnricher extends Transform {
  private actionCategories: Record<string, string> = {
    page_view: "engagement",
    click: "engagement",
    search: "engagement",
    purchase: "conversion",
    add_to_cart: "conversion",
    signup: "acquisition",
    logout: "session",
  };

  constructor() {
    super({ objectMode: true });
  }

  _transform(event: RawEvent, encoding: string, callback: Function) {
    const date = new Date(event.timestamp);
    const enriched: EnrichedEvent = {
      ...event,
      date: date.toISOString().split("T")[0],
      hour: date.getHours(),
      isWeekend: date.getDay() === 0 || date.getDay() === 6,
      actionCategory: this.actionCategories[event.action] ?? "unknown",
    };
    this.push(enriched);
    callback();
  }
}

// Transform: Filter (only conversion events)
class ConversionFilter extends Transform {
  public filtered = 0;
  public passed = 0;

  constructor() {
    super({ objectMode: true });
  }

  _transform(event: EnrichedEvent, encoding: string, callback: Function) {
    if (event.actionCategory === "conversion") {
      this.passed++;
      this.push(event);
    } else {
      this.filtered++;
    }
    callback();
  }
}

// Load: Aggregate into stats
class StatsAggregator extends Writable {
  public stats: AggregatedStats = {
    totalEvents: 0,
    uniqueUsers: new Set(),
    actionCounts: new Map(),
    hourlyDistribution: new Array(24).fill(0),
  };

  constructor() {
    super({ objectMode: true });
  }

  _write(event: EnrichedEvent, encoding: string, callback: Function) {
    this.stats.totalEvents++;
    this.stats.uniqueUsers.add(event.userId);
    this.stats.actionCounts.set(
      event.action,
      (this.stats.actionCounts.get(event.action) ?? 0) + 1
    );
    this.stats.hourlyDistribution[event.hour]++;
    callback();
  }
}

// Run the ETL
const filter = new ConversionFilter();
const loader = new StatsAggregator();
const start = performance.now();

await pipeline(
  createEventSource(100_000),
  new EventEnricher(),
  filter,
  loader,
);

const elapsed = performance.now() - start;

console.log(`ETL Pipeline Results (${elapsed.toFixed(0)}ms):`);
console.log(`  Total events generated: ${filter.passed + filter.filtered}`);
console.log(`  Conversion events: ${filter.passed} (${(filter.passed / (filter.passed + filter.filtered) * 100).toFixed(1)}%)`);
console.log(`  Unique users: ${loader.stats.uniqueUsers.size}`);
console.log(`  Action breakdown:`);
for (const [action, count] of loader.stats.actionCounts) {
  console.log(`    ${action}: ${count}`);
}
console.log(`  Peak hour: ${loader.stats.hourlyDistribution.indexOf(Math.max(...loader.stats.hourlyDistribution))}:00`);
console.log(`  Memory: ${(process.memoryUsage().heapUsed / 1024 / 1024).toFixed(1)}MB`);
```

---

## Challenges

1. **Build a streaming diff tool**: Compare two files line-by-line using streams, outputting unified diff format. Files can be larger than memory.

2. **Implement rate limiting as a Transform**: Create a Transform that limits throughput to N bytes/second, useful for throttling downloads.

3. **Build a multiplexer**: A stream that takes multiple input streams and interleaves their output into a single output stream, with framing to identify the source.
