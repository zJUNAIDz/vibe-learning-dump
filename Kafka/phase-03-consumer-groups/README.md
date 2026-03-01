# Phase 3 — Consumer Groups & Coordination

## The Problem We're Solving

In Phase 2, we produced to 3 partitions. But we ran a single consumer that reads from all 3. That's not scaling — it's the same throughput with extra complexity.

To scale, we want **multiple consumer instances**, each processing a subset of partitions. But who decides which consumer gets which partition? What happens when a consumer crashes? What happens when we add a new consumer?

This is where **consumer groups** come in.

## Kafka Concepts Introduced

### Consumer Groups

A consumer group is a set of consumers that cooperate to consume a topic. Kafka ensures:

1. **Each partition is assigned to exactly one consumer** in the group
2. **A consumer can handle multiple partitions** (if there are fewer consumers than partitions)
3. **Adding/removing consumers triggers a rebalance** — partitions are reassigned

```mermaid
graph TB
    subgraph "orders topic (3 partitions)"
        P0[Partition 0]
        P1[Partition 1]
        P2[Partition 2]
    end
    
    subgraph "payment-group (3 consumers)"
        C0[Consumer 0]
        C1[Consumer 1]
        C2[Consumer 2]
    end
    
    P0 --> C0
    P1 --> C1
    P2 --> C2
    
    subgraph "analytics-group (1 consumer)"
        A0[Analytics Consumer]
    end
    
    P0 --> A0
    P1 --> A0
    P2 --> A0
```

Key insight: **different consumer groups are completely independent**. The `payment-group` and `analytics-group` both receive *all* messages. Within each group, messages are divided among the consumers.

### Rebalancing

When consumers join or leave a group, Kafka **rebalances** — it reassigns partitions across the remaining consumers.

```mermaid
sequenceDiagram
    participant K as Kafka Coordinator
    participant C1 as Consumer 1
    participant C2 as Consumer 2
    participant C3 as Consumer 3
    
    Note over K,C2: Initial State: 2 consumers
    K->>C1: Assign P0, P1
    K->>C2: Assign P2
    
    Note over K,C3: Consumer 3 joins
    K->>C1: REBALANCE: Stop processing
    K->>C2: REBALANCE: Stop processing
    Note over K: Reassigning partitions...
    K->>C1: Assign P0
    K->>C2: Assign P1
    K->>C3: Assign P2
    
    Note over K,C3: Consumer 2 crashes
    K->>C1: REBALANCE: Stop processing
    K->>C3: REBALANCE: Stop processing
    Note over K: Reassigning partitions...
    K->>C1: Assign P0, P1
    K->>C3: Assign P2
```

### Why Rebalances Hurt

During a rebalance:
1. **All consumers in the group stop processing** (even healthy ones)
2. **Partitions are redistributed** — a consumer might get a different partition than before
3. **In-flight messages may be reprocessed** — if offsets weren't committed before the rebalance

Rebalances are the #1 operational pain point with Kafka consumer groups. They're necessary, but they cause pauses and potential duplicate processing.

### Offset Commits

Consumers track their progress by **committing offsets** — telling Kafka "I've processed everything up to offset N."

Two strategies:

| Strategy | How It Works | Risk |
|----------|-------------|------|
| **Auto-commit** | Kafka commits offsets periodically (every 5s by default) | Crash between commit and processing → lost messages |
| **Manual commit** | You commit after processing | More code, but you control exactly when |

```mermaid
graph LR
    subgraph "Auto-Commit Problem"
        direction TB
        A1["Read offset 5"] --> A2["Auto-commit offset 6"]
        A2 --> A3["💥 Crash during processing"]
        A3 --> A4["Restart: offset 6 committed"]
        A4 --> A5["❌ Message 5 was never processed<br/>but Kafka thinks it was"]
    end
    
    subgraph "Manual Commit (Safe)"
        direction TB
        B1["Read offset 5"] --> B2["Process message"]
        B2 --> B3["Commit offset 6"]
        B3 --> B4["💥 Crash? No problem."]
        B4 --> B5["Restart: still at offset 5"]
        B5 --> B6["✅ Message 5 re-processed"]
    end
    
    style A5 fill:#ff6b6b,color:#fff
    style B6 fill:#51cf66,color:#fff
```

Manual commits give you **at-least-once semantics**: you might process a message twice (if you crash after processing but before committing), but you'll never miss one.

## Code

- [TypeScript Implementation](ts-implementation.md)
- [Go Implementation](go-implementation.md)

## What Breaks If Misused

| Mistake | What Happens |
|---------|-------------|
| More consumers than partitions | Extra consumers sit idle, wasting resources |
| Auto-commit with slow processing | Messages "committed" before actually processed. Crash = data loss. |
| Long processing time per message | Consumer heartbeat times out → Kafka thinks it's dead → rebalance storm |
| Not handling rebalances | Consumer gets new partition but starts from wrong offset → duplicates or gaps |
| Multiple consumer groups for same service | Each group gets all messages → duplicate processing |

## What's Next

We now have scaled consumers with proper offset management. But what happens when processing *fails*? In [Phase 4](../phase-04-failure-retries/README.md), we deal with retries, idempotency, and the reality that messages sometimes can't be processed.
