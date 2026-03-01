# Phase 2 — Partitioning & Scale

## The Problem We're Solving

In Phase 1, we have one topic with one partition. One consumer reads everything sequentially. This works for low throughput, but what happens when we need to process 10,000 orders per second?

A single consumer can't keep up. We need parallelism.

But we also need **ordering guarantees**: all events for the same order must be processed in sequence. Order-created before order-paid before order-shipped.

These two requirements are in tension. Parallelism breaks ordering. Partitioning is the mechanism that resolves this tension.

## Kafka Concepts Introduced

### Partitions

A topic is split into one or more partitions. Each partition is an independent, ordered, append-only log.

```mermaid
graph TB
    subgraph "orders topic"
        subgraph "Partition 0"
            P0O0["offset 0<br/>user-1"] --> P0O1["offset 1<br/>user-1"] --> P0O2["offset 2<br/>user-1"]
        end
        subgraph "Partition 1"
            P1O0["offset 0<br/>user-2"] --> P1O1["offset 1<br/>user-2"]
        end
        subgraph "Partition 2"
            P2O0["offset 0<br/>user-3"] --> P2O1["offset 1<br/>user-3"] --> P2O2["offset 2<br/>user-3"]
        end
    end
```

Key properties:
- **Ordering is guaranteed within a partition**, not across partitions
- **Each partition can be consumed by a different consumer** — this is how you parallelize
- **Partitions are the unit of parallelism** — you can't have more active consumers than partitions

### Message Keys

How does Kafka decide which partition a message goes to?

- **No key:** Round-robin across partitions (no ordering guarantee per entity)
- **With key:** `hash(key) % numPartitions` → same key always goes to the same partition

This is the critical insight: **the key determines which partition a message lands in**.

If you use `userId` as the key, all events for `user-1` go to the same partition. This guarantees ordering per user.

```mermaid
graph LR
    P[Producer] -->|"key=user-1"| Part0[Partition 0]
    P -->|"key=user-2"| Part1[Partition 1]
    P -->|"key=user-3"| Part2[Partition 2]
    P -->|"key=user-1"| Part0
    P -->|"key=user-2"| Part1
    
    Part0 -->|consume| C0[Consumer 0]
    Part1 -->|consume| C1[Consumer 1]
    Part2 -->|consume| C2[Consumer 2]
```

### Why Not Just Add More Partitions?

- Partitions are fixed at topic creation (can increase but not decrease)
- More partitions = more open file handles on brokers
- More partitions = longer leader election during failures
- More partitions = more memory for consumer group coordination

**Rule of thumb**: start with `max(expected_throughput / consumer_throughput, num_consumers)`. You can always add more. You can never remove them.

## Key-Based Ordering In Our System

For our order pipeline, what's the right key?

| Key Choice | Effect | Good For |
|-----------|--------|----------|
| `orderId` | All events for one order go to one partition | Order lifecycle events (created → paid → shipped) |
| `userId` | All events for one user go to one partition | Per-user ordering, user activity streams |
| No key | Round-robin | Maximum throughput, no ordering needed |

We'll use `orderId` as the key. This means:
- `order-created`, `order-paid`, and `order-shipped` for the same order always land on the same partition
- They're always processed in order by the same consumer

## What Breaks If You Get This Wrong

```mermaid
graph TB
    subgraph "No Key (Round-Robin)"
        P1["order-1 CREATED<br/>→ Partition 0"] 
        P2["order-1 PAID<br/>→ Partition 1"]
        P3["order-1 SHIPPED<br/>→ Partition 2"]
    end
    
    C0[Consumer 0] -->|"processes CREATED"| R0["✅"]
    C2[Consumer 2] -->|"processes SHIPPED before PAID!"| R2["❌ Bug"]
    C1[Consumer 1] -->|"processes PAID after SHIPPED"| R1["❌ Bug"]
    
    P1 -.-> C0
    P2 -.-> C1
    P3 -.-> C2
```

Without a key, events for the same order can land on different partitions and be consumed out of order. A payment might be processed before the order is created. An order might be shipped before it's paid.

## Code

- [TypeScript Implementation](ts-implementation.md)
- [Go Implementation](go-implementation.md)

## What's Next

We have partitioned producers and multiple consumers. But right now, we started each consumer manually and told it which partition to read. That doesn't scale operationally.

In [Phase 3](../phase-03-consumer-groups/README.md), Kafka assigns partitions to consumers automatically through **consumer groups**.
