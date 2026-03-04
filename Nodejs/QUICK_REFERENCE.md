# Quick Reference — Node.js Runtime Internals

## Event Loop Phase Order

```
┌───────────────────────────┐
│         timers             │  ← setTimeout, setInterval
├───────────────────────────┤
│    pending callbacks       │  ← I/O callbacks deferred
├───────────────────────────┤
│      idle, prepare         │  ← internal use only
├───────────────────────────┤
│          poll              │  ← I/O events, incoming data
├───────────────────────────┤
│         check              │  ← setImmediate
├───────────────────────────┤
│    close callbacks         │  ← socket.on('close')
└───────────────────────────┘

Between EVERY phase:
  → process.nextTick queue (drained completely)
  → Promise microtask queue (drained completely)
```

## Execution Priority (highest → lowest)

1. `process.nextTick()` — before any I/O or timers
2. `Promise.then()` / `queueMicrotask()` — microtask queue
3. `setTimeout(fn, 0)` — timers phase
4. `setImmediate(fn)` — check phase
5. I/O callbacks — poll phase

## libuv Thread Pool

| Operation | Uses Thread Pool? | Kernel Async? |
|-----------|------------------|---------------|
| `fs.*` | Yes (4 threads default) | No |
| `crypto.*` | Yes | No |
| `dns.lookup()` | Yes | No |
| `dns.resolve()` | No | Yes (c-ares) |
| `net.*` (TCP) | No | Yes (epoll/kqueue) |
| `http.*` | No | Yes (epoll/kqueue) |
| `child_process` | No | Yes (signals) |
| `zlib.*` | Yes | No |

## V8 Heap Structure

```
┌─────────────────────────────────┐
│           V8 Heap               │
├────────────┬────────────────────┤
│  New Space │    Old Space       │
│  (young)   │    (tenured)       │
│  1-8 MB    │    up to --max-old │
│            │                    │
│  Scavenge  │   Mark-Sweep-     │
│  GC (fast) │   Compact (slow)  │
└────────────┴────────────────────┘
```

## Key V8 Flags

```bash
--max-old-space-size=4096    # Max old space in MB
--max-semi-space-size=64     # Max new space semi-space in MB
--expose-gc                  # Expose global.gc()
--trace-gc                   # Log GC events
--prof                       # CPU profiling
--heap-prof                  # Heap profiling
```

## Key Environment Variables

```bash
UV_THREADPOOL_SIZE=8         # libuv thread pool (default 4, max 1024)
NODE_OPTIONS="--max-old-space-size=4096"
NODE_DEBUG=net,http,fs       # Debug specific modules
```

## Stream Types Cheat Sheet

| Type | Purpose | Key Method |
|------|---------|------------|
| Readable | Source of data | `.read()`, `.pipe()` |
| Writable | Destination for data | `.write()`, `.end()` |
| Duplex | Both read and write | Both above |
| Transform | Modify data in transit | `._transform()` |

## Worker Thread Communication

```
Main Thread              Worker Thread
    │                        │
    │──── postMessage() ────→│
    │                        │
    │←── parentPort.         │
    │    postMessage() ──────│
    │                        │
    │  SharedArrayBuffer     │
    │◄══════════════════════►│
    │  (zero-copy shared)    │
```

## Cluster Architecture

```
            ┌─────────────┐
            │   Master    │
            │  (no HTTP)  │
            └──────┬──────┘
         ┌─────────┼─────────┐
         │         │         │
    ┌────▼──┐ ┌───▼───┐ ┌──▼────┐
    │Worker1│ │Worker2│ │Worker3│
    │ :3000 │ │ :3000 │ │ :3000 │
    └───────┘ └───────┘ └───────┘
    
    All workers share the same port
    via SO_REUSEPORT or IPC fd passing
```
