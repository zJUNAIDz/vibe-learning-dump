# Split-Brain and Partition Tolerance — The Impossible Choices

---

## What a Network Partition Really Is

A network partition isn't "the network is down." It's **"the network is partially down."** Nodes can talk to some nodes but not others. The cluster splits into disconnected groups, each believing the other side has failed.

```mermaid
graph TD
    subgraph "Normal Cluster"
        N1[Node 1] --- N2[Node 2]
        N2 --- N3[Node 3]
        N3 --- N4[Node 4]
        N4 --- N5[Node 5]
        N1 --- N3
        N2 --- N4
    end
    
    subgraph "Partitioned Cluster"
        subgraph "Side A"
            PA1[Node 1] --- PA2[Node 2]
            PA1 --- PA3[Node 3]
            PA2 --- PA3
        end
        
        subgraph "Side B"
            PB1[Node 4] --- PB2[Node 5]
        end
    end
    
    PA3 -.-x|"PARTITION<br/>No communication"| PB1
    
    style PA3 fill:#ff6b6b,color:#fff
    style PB1 fill:#ff6b6b,color:#fff
```

Partitions happen because of:
- Switch failure between racks
- Firewall rule change
- Cloud provider network issues
- DNS failures
- GC pauses so long that heartbeats time out (Cassandra JVM)
- Overloaded network links dropping packets

Partitions are **not rare**. Any production system running long enough will experience them.

---

## The CAP Theorem Choice — In Practice

During a partition, a database must choose:

1. **Stay consistent (CP)**: Refuse writes on the minority side. Some clients can't write. Availability drops.
2. **Stay available (AP)**: Accept writes on both sides. Data diverges. Consistency drops.

```mermaid
graph TD
    PART[Network Partition<br/>Occurs]
    PART --> CHOICE{What does the<br/>database do?}
    
    CHOICE -->|"CP: Consistency"| REFUSE["Minority side refuses<br/>reads/writes ❌<br/>Clients get errors"]
    CHOICE -->|"AP: Availability"| ACCEPT["Both sides accept<br/>reads/writes ✅<br/>Data may diverge"]
    
    REFUSE --> HEAL1["Partition heals →<br/>No conflict to resolve ✅"]
    ACCEPT --> HEAL2["Partition heals →<br/>Must merge conflicting data ⚠️"]
    
    style REFUSE fill:#3498db,color:#fff
    style ACCEPT fill:#e67e22,color:#fff
```

### What Each Database Chooses

| Database | During Partition | Choice | What Happens |
|----------|-----------------|--------|--------------|
| PostgreSQL (single) | N/A — no partition possible | CP | One machine, always consistent |
| MongoDB | Primary on minority side steps down | CP | Minority can't write; majority elects new primary |
| Cassandra (CL=QUORUM) | Minority can't reach quorum | CP for that query | Write fails, client retries |
| Cassandra (CL=ONE) | Both sides accept writes | AP | Conflicts resolved by LWW after partition heals |
| DynamoDB | AWS handles internally | Depends on configuration | Strong reads may fail; eventual reads continue |
| CockroachDB | Minority ranges become unavailable | CP | Majority side continues; minority halts |
| Redis Sentinel | Old primary steps down eventually | CP (with brief confusion) | Possible data loss during failover |

---

## Scenario: MongoDB During a Partition

```mermaid
sequenceDiagram
    participant C1 as Client (US)
    participant P as Primary (US)
    participant S1 as Secondary (US)
    participant S2 as Secondary (EU)
    
    Note over P,S2: Partition: US can't reach EU
    
    C1->>P: Write order
    P->>S1: Replicate ✅
    P->>S2: Replicate ❌ (partitioned)
    
    Note over P: 2 of 3 nodes reachable
    Note over P: Primary has majority → stays primary ✅
    P-->>C1: Write acknowledged ✅
    
    Note over S2: Can't reach primary<br/>Knows it's in minority<br/>Does NOT attempt election
    
    Note over P,S2: Partition heals
    S2->>P: Catch up via oplog ✅
```

**If the primary were on the minority side**:

```mermaid
sequenceDiagram
    participant C1 as Client (EU)
    participant P as Primary (EU)
    participant S1 as Secondary (US)
    participant S2 as Secondary (US)
    
    Note over P,S2: Partition: EU can't reach US
    
    Note over P: Primary has minority (1 of 3)<br/>Steps down after election timeout
    
    Note over S1,S2: US side has majority (2 of 3)<br/>Hold election → S1 becomes primary
    
    C1->>P: Write order
    P-->>C1: Error: not primary ❌
    
    Note over C1: Client must find<br/>new primary (retry)
    
    Note over P,S2: Partition heals
    P->>S1: Discovers S1 is primary
    P->>P: Roll back uncommitted writes
    P->>P: Become secondary, sync from S1
```

**Key trade-off**: During the partition, clients near the old primary can't write. But no data conflicts occur.

---

## Scenario: Cassandra During a Partition

```mermaid
graph TD
    subgraph "Partitioned Cassandra RF=3"
        subgraph "Side A (2 nodes)"
            CA1[Node 1<br/>Has replica]
            CA2[Node 2<br/>Has replica]
        end
        
        subgraph "Side B (1 node)"
            CB1[Node 3<br/>Has replica]
        end
    end
    
    CQ["Client writes at CL=QUORUM"]
    CO["Client writes at CL=ONE"]
    
    CQ -->|"Side A: 2 ≥ 2 ✅"| CA1
    CQ -->|"Side B: 1 < 2 ❌"| CB1
    
    CO -->|"Side A: 1 ≥ 1 ✅"| CA1
    CO -->|"Side B: 1 ≥ 1 ✅"| CB1
    
    NOTE1["At QUORUM: Side B<br/>is unavailable (CP)"]
    NOTE2["At ONE: Both sides<br/>accept writes (AP)"]
    
    style NOTE1 fill:#3498db,color:#fff
    style NOTE2 fill:#e67e22,color:#fff
```

**At CL=QUORUM**: Cassandra behaves CP. The minority side rejects writes because it can't reach quorum. No conflicts.

**At CL=ONE**: Cassandra behaves AP. Both sides accept writes. When the partition heals, conflicting writes are resolved by LWW (last-write-wins). Data can be silently lost.

This is why Cassandra's consistency is "tunable" — you choose CP or AP per query.

---

## The Split-Brain Problem

True split-brain: both sides of a partitioned cluster believe they are the rightful authority and accept writes.

### When Split-Brain Causes Real Damage

```
User has $100 in their account.

Side A: User withdraws $80 → balance = $20
Side B: User withdraws $60 → balance = $40

Partition heals → LWW picks one:
Result: balance = $40 (Side B wins by timestamp)

Problem: User withdrew $80 + $60 = $140 from $100 account
The $80 withdrawal is lost. Bank is out $80.
```

This is why financial systems don't use AP databases (or if they do, they add external coordination).

### Preventing Split-Brain

| Technique | How It Works | Used By |
|-----------|-------------|---------|
| Quorum writes | Minority can't reach quorum → writes fail | Cassandra (QUORUM), MongoDB |
| Leader election | Only leader accepts writes; minority has no leader | MongoDB, etcd, CockroachDB |
| Fencing tokens | Old leader's writes are rejected by storage | ZooKeeper, etcd |
| STONITH | "Shoot The Other Node In The Head" — power off the misbehaving node | Pacemaker, VMware HA |

---

## What "Partition Tolerance" Means

Every distributed system is partition-tolerant — it's not optional. Networks WILL partition. The question is what the system does when it happens.

"Partition-tolerant" just means "the system has a defined behavior during partitions." It might refuse writes (CP). It might accept conflicting writes (AP). But it doesn't crash or corrupt data silently.

```mermaid
graph TD
    subgraph "The Real CAP Question"
        Q["During a partition,<br/>which do you sacrifice?"]
        Q --> C["Consistency<br/>Accept conflicting writes<br/>(AP systems)"]
        Q --> A["Availability<br/>Refuse writes from minority<br/>(CP systems)"]
        
        Note["You ALWAYS have partition<br/>tolerance. P is not a choice.<br/>The choice is between C and A."]
    end
    
    style C fill:#e67e22,color:#fff
    style A fill:#3498db,color:#fff
```

---

## Multi-Region Partition Strategies

In multi-DC deployments, inter-DC partitions are the most common (undersea cables, WAN issues):

```mermaid
graph LR
    subgraph "US East"
        US1[Node 1]
        US2[Node 2]
        US3[Node 3]
    end
    
    subgraph "EU West"
        EU1[Node 4]
        EU2[Node 5]
        EU3[Node 6]
    end
    
    US3 -.-x|"WAN partition"| EU1
```

### Strategy 1: Active-Passive (MongoDB, PostgreSQL)

One DC is primary. Other DC has read replicas only. During partition:
- Primary DC continues normally
- Passive DC serves stale reads
- No write conflicts possible

### Strategy 2: Active-Active with LOCAL_QUORUM (Cassandra)

Both DCs accept writes, but only require local quorum:
- Each DC operates independently during partition
- Cross-DC replication catches up when partition heals
- Same-key conflicts resolved by LWW

### Strategy 3: Global Consensus (CockroachDB, Spanner)

Every write goes through consensus across DCs:
- Strong consistency globally
- High latency (50-200ms per write for cross-DC consensus)
- During partition, ranges on the minority side become unavailable

| Strategy | Consistency | Latency | Partition Behavior |
|----------|------------|---------|-------------------|
| Active-Passive | Strong | Low (local reads) | Passive can't write |
| Active-Active LOCAL_QUORUM | Eventual (cross-DC) | Low | Both DCs write, LWW merge |
| Global Consensus | Strong | High | Minority unavailable |

---

## The Practical Decision

For most applications:

1. **Single region**: Use MongoDB or CockroachDB. Partitions within a region are rare and brief. CP behavior during partitions is acceptable.
2. **Multi-region, read-heavy**: Active-passive. Route reads to nearest DC. Writes go to primary DC.
3. **Multi-region, write-heavy**: Cassandra with LOCAL_QUORUM. Accept eventual consistency across DCs. Design data model to avoid cross-DC conflicts.
4. **Multi-region, strong consistency**: CockroachDB or Spanner. Pay the latency cost.

---

## Next Phase

→ [../05-data-modeling-patterns/01-embedding-patterns.md](../05-data-modeling-patterns/01-embedding-patterns.md) — Practical data modeling patterns for NoSQL databases: embedding, fan-out, bucketing, and more.
