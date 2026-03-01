# Phase 1 — Kafka as a Log

## The Problem We're Solving

In Phase 0, we built a synchronous chain: Order → Payment → Notification → Inventory. Every service had to be up, fast, and compatible for the chain to work.

The core issue: **temporal coupling**. The Order Service needs the Payment Service to be alive *right now*.

## The Shift

What if the Order Service didn't call the Payment Service at all?

What if it just wrote down: *"An order happened"* — and the Payment Service picked it up whenever it was ready?

That's Kafka. It's not a message queue. It's an **append-only log**.

## Kafka Concepts Introduced

### The Append-Only Log

Kafka stores messages in an ordered, immutable, append-only log. Think of it like a file you can only append to, and anyone can read from any position.

```mermaid
graph LR
    subgraph "orders topic (single partition)"
        direction LR
        O0["offset 0<br/>order-1"] 
        O1["offset 1<br/>order-2"]
        O2["offset 2<br/>order-3"]
        O3["offset 3<br/>order-4"]
        O4["offset 4<br/>order-5"]
    end
    
    P[Producer] -->|append| O4
    C1[Consumer A<br/>offset: 2] -.->|reading| O2
    C2[Consumer B<br/>offset: 4] -.->|reading| O4
```

Key properties:
- **Messages are never deleted** (until retention expires)
- **Messages are ordered** within a partition
- **Each message gets an offset** — a sequential ID
- **Consumers track their own position** — they choose where to read from
- **Multiple consumers can read the same data** independently

### Topics

A topic is a named log. Think of it as a category of events.

```
orders     → all order events go here
payments   → all payment events go here
```

### Offsets

An offset is just a number. It's the position of a message in the log.

- **Offset 0** is the first message ever written
- **Offset 47** is the 48th message
- Each consumer tracks which offset it has processed

This is why Kafka enables **replay**: you can always set a consumer back to offset 0 and reprocess everything.

## New Architecture

```mermaid
graph LR
    API[Order API] -->|"1. produce"| KT["orders topic"]
    KT -->|"2. consume"| PS[Payment Service]
    KT -->|"2. consume"| AS[Analytics Service]
    
    style KT fill:#f5f5f5,stroke:#333,stroke-width:2px
```

The Order Service doesn't know (or care) who is listening. It writes to the log. Done.

The Payment Service reads from the log at its own pace. If it's down for 5 minutes, it catches up when it comes back.

## What Changed (Before vs After)

```mermaid
graph TB
    subgraph "Before (Phase 0)"
        direction LR
        A1[Order] -->|HTTP| B1[Payment]
        B1 -->|HTTP| C1[Notification]
        C1 -->|HTTP| D1[Inventory]
    end
    
    subgraph "After (Phase 1)"
        direction LR
        A2[Order] -->|produce| K[Kafka Log]
        K -->|consume| B2[Payment]
        K -->|consume| E2[Analytics]
    end
```

Notice:
- **No direct service-to-service calls** for the initial event
- **Adding a new consumer** doesn't require changing the producer
- **Services are decoupled in time** — they don't need to be alive simultaneously

## Code

- [TypeScript Implementation](ts-implementation.md)
- [Go Implementation](go-implementation.md)

## What Breaks If Misused

| Mistake | What Happens |
|---------|-------------|
| Treating Kafka like a queue | You delete messages after reading. Kafka doesn't work that way. Messages stay. |
| Producing without a key | Messages go to random partitions. Ordering per entity is lost. (We fix this in Phase 2.) |
| Ignoring offsets | Consumer restarts and re-reads everything from the beginning. Duplicate processing. |
| Assuming delivery order = production order | With multiple partitions, ordering is only guaranteed *within* a partition. (Phase 2.) |

## What's Next

In [Phase 2](../phase-02-partitions/README.md), we add partitions — because a single log doesn't scale.
