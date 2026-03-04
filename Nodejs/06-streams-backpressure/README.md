# Module 06 — Streams & Backpressure

## Overview

Streams are Node.js's most powerful abstraction for handling data that's too large to fit in memory, or data that arrives over time. They enable processing gigabytes of data with constant memory usage. But streams are also the hardest Node.js concept to master — backpressure bugs cause memory leaks, data loss, and silent failures.

---

## Stream Taxonomy

```mermaid
graph TB
    subgraph "Stream Types"
        R[Readable] -->|"pipe"| W[Writable]
        R -->|"pipe"| T[Transform]
        T -->|"pipe"| W
        D[Duplex] -->|"both directions"| D
    end
    
    subgraph "Examples"
        R1[fs.createReadStream]
        R2[http.IncomingMessage]
        R3[process.stdin]
        
        W1[fs.createWriteStream]
        W2[http.ServerResponse]
        W3[process.stdout]
        
        T1[zlib.createGzip]
        T2[crypto.createCipher]
        
        D1[net.Socket]
        D2[WebSocket]
    end
```

---

## Lessons

| # | Lesson | What You'll Learn |
|---|--------|-------------------|
| 01 | [Readable Streams](01-readable-streams.md) | Internal buffering, flowing vs paused mode, async iteration |
| 02 | [Writable Streams](02-writable-streams.md) | Write buffering, drain event, cork/uncork |
| 03 | [Backpressure](03-backpressure.md) | Why pipe() exists, manual backpressure, pipeline() |
| 04 | [Transform & Pipeline Labs](04-transform-labs.md) | Build real streaming data processors |

---

## Key Takeaways

- Every stream has an internal buffer limited by `highWaterMark`
- Readable streams have two modes: flowing (data events) and paused (manual read)
- `write()` returns `false` when the internal buffer is full — you MUST respect this
- `pipeline()` is the only safe way to connect streams (handles errors and cleanup)
- Never use `.pipe()` in production — it doesn't handle errors properly
- Backpressure propagates upstream: if the consumer is slow, the producer pauses
