# Read Repair and Anti-Entropy — Self-Healing Distributed Systems

---

## The Reality: Replicas Drift

Even in a healthy cluster, data diverges across replicas:
- Network dropped a packet during replication
- A node restarted and missed some writes
- A disk had a silent corruption (bit flip)
- Hinted handoff didn't fully catch up after a node outage

Distributed databases address this with two mechanisms: **read repair** (on-demand) and **anti-entropy repair** (scheduled).

---

## Read Repair

When you read from multiple replicas and they disagree, the coordinator fixes the stale replicas in the background.

```mermaid
sequenceDiagram
    participant C as Client
    participant Coord as Coordinator
    participant R1 as Replica 1
    participant R2 as Replica 2
    participant R3 as Replica 3
    
    C->>Coord: READ user_id=42
    
    Note over Coord: Read at QUORUM (2 of 3)
    Coord->>R1: Full read request
    Coord->>R2: Digest read request
    Coord->>R3: Digest read request
    
    R1-->>Coord: Full data (version 5)
    R2-->>Coord: Digest hash (matches v5) ✅
    
    Coord-->>C: Return version 5
    
    Note over Coord: R3's digest arrives late
    R3-->>Coord: Digest hash (matches v3 — STALE!)
    
    Note over Coord,R3: Background read repair
    Coord->>R3: Here's version 5, update yourself
    R3-->>Coord: Repaired ✅
```

### How Cassandra Read Repair Works

Cassandra sends the full data request to one replica and a **digest request** (hash of the data) to the others.

- If all digests match → data is consistent, return immediately
- If digests differ → fetch full data from all disagreeing replicas, determine the latest version (highest timestamp), return it to the client, and send corrections to stale replicas

### Types of Read Repair in Cassandra

| Type | When | Performance Impact |
|------|------|-------------------|
| Blocking read repair | Digest mismatch during a QUORUM read | Adds ~1-5ms to the read (must fetch full data from disagreeing replica) |
| Background read repair | After returning to client, based on `read_repair_chance` | None to client — happens asynchronously |

```sql
-- Configure read repair probability
ALTER TABLE users WITH read_repair_chance = 0.1;
-- 10% of reads trigger background read repair
-- (In Cassandra 4.0+, this is deprecated in favor of full repair)
```

### Read Repair Limitations

1. **Only repairs data that's read** — Data nobody reads stays inconsistent
2. **Per-partition only** — Won't fix inconsistencies in partitions that aren't queried
3. **Doesn't handle all failure modes** — Can't fix data that's been lost on all replicas

---

## Anti-Entropy Repair

A scheduled, full-cluster consistency check that repairs ALL data, not just data that's read.

### How It Works: Merkle Trees

Cassandra builds a **Merkle tree** (hash tree) for each table on each node. Nodes exchange tree roots and compare:

```mermaid
graph TD
    subgraph "Merkle Tree - Node 1"
        NH1["Root Hash: a1b2c3"]
        NH1 --> L1["Left: x7y8"]
        NH1 --> R1["Right: p9q0"]
        L1 --> LL1["Partition A-M<br/>Hash: abcd"]
        L1 --> LR1["Partition N-Z<br/>Hash: efgh"]
    end
    
    subgraph "Merkle Tree - Node 2"
        NH2["Root Hash: a1b2c3"]
        NH2 --> L2["Left: x7y8"]
        NH2 --> R2["Right: DIFFERENT!"]
        L2 --> LL2["Partition A-M<br/>Hash: abcd ✅"]
        L2 --> LR2["Partition N-Z<br/>Hash: efgh ✅"]
    end
    
    R1 -.->|"Mismatch!<br/>Drill down to find<br/>specific partitions"| R2
    
    style R2 fill:#ff6b6b,color:#fff
```

1. **Build**: Each node hashes its data into a Merkle tree
2. **Compare**: Nodes exchange root hashes. If roots match → data is consistent.
3. **Drill down**: If roots differ, compare child hashes to identify exactly which partitions diverge.
4. **Stream**: Send only the divergent data to the node with the stale version.

This is efficient — instead of comparing every row, you compare hashes. A single root hash comparison tells you if billions of rows are consistent.

### Running Repairs

```bash
# Full repair — compares all data with replicas
# Run this weekly per node (not all nodes simultaneously)
nodetool repair --full

# Incremental repair — only repairs data written since last repair
# Faster, can run more frequently
nodetool repair

# Repair a specific keyspace and table
nodetool repair myapp users
```

### Repair Schedule

| Repair Type | Frequency | Duration | Impact |
|-------------|-----------|----------|--------|
| Full repair | Every `gc_grace_seconds` (default 10 days) | Hours on large datasets | High I/O, high network |
| Incremental repair | Daily or every few days | Minutes to an hour | Moderate I/O |
| Subrange repair | Rotate through token ranges | Varies | Low per-run impact |

**Critical rule**: You MUST run repair at least once within `gc_grace_seconds` (10 days default). Otherwise, deleted data (tombstones) that have been garbage-collected on some nodes may **resurrect** on others — zombie data.

---

## MongoDB's Approach: Oplog-Based Replication

MongoDB doesn't need read repair or anti-entropy because it uses a fundamentally different architecture:

```mermaid
graph TD
    subgraph "MongoDB Replica Set"
        P[Primary<br/>All writes here]
        S1[Secondary 1<br/>Tails oplog]
        S2[Secondary 2<br/>Tails oplog]
    end
    
    P -->|"Oplog stream"| S1
    P -->|"Oplog stream"| S2
    
    style P fill:#3498db,color:#fff
```

**Single writer (primary)** means no write conflicts. Secondaries **tail the oplog** — an ordered log of all operations on the primary. If a secondary falls behind, it replays missed oplog entries.

If a secondary falls too far behind (oplog has been overwritten), it does an **initial sync** — copies the entire dataset from the primary.

### MongoDB's Repair Equivalent

```javascript
// Check replica lag
rs.status().members.forEach(m => {
  if (m.stateStr === 'SECONDARY') {
    console.log(`${m.name}: lag = ${m.optimeDate - rs.status().date}ms`);
  }
});

// If a secondary is severely out of sync, force initial sync:
// 1. Stop the mongod on the secondary
// 2. Delete its data directory
// 3. Restart — it will initial sync from the primary
```

---

## DynamoDB's Approach: Managed Anti-Entropy

DynamoDB handles all repair internally. As a managed service, you don't run repair tools. Amazon's infrastructure:
- Detects inconsistent replicas automatically
- Repairs them in the background
- Provides no visibility into this process (it's fully abstracted)

This is the trade-off of managed services: less control, less operational burden.

---

## Failure Recovery Flow

```mermaid
graph TD
    FAIL[Node failure<br/>or data loss]
    FAIL --> DURATION{How long<br/>was it down?}
    
    DURATION -->|"Seconds-minutes"| HH["Hinted handoff<br/>catches up automatically"]
    DURATION -->|"Minutes-hours"| RR["Read repair<br/>fixes on demand +<br/>anti-entropy for rest"]
    DURATION -->|"Hours-days"| REPAIR["Full repair required<br/>nodetool repair --full"]
    DURATION -->|"Longer than<br/>gc_grace_seconds"| REBUILD["Full rebuild<br/>from other replicas<br/>(or risk zombie data)"]
    
    style HH fill:#4ecdc4,color:#fff
    style RR fill:#f39c12,color:#fff
    style REPAIR fill:#e67e22,color:#fff
    style REBUILD fill:#ff6b6b,color:#fff
```

---

## Split-Brain Recovery

The hardest recovery scenario: a network partition splits the cluster and both sides accepted writes.

### Cassandra Split-Brain Recovery

Cassandra is leaderless — both sides can accept writes at CL=ONE during a partition. When the partition heals:

1. Read repair and anti-entropy detect divergent data
2. LWW (last-write-wins by timestamp) resolves conflicts
3. **Data loss is possible** — if both sides wrote to the same row, the lower-timestamp write is silently discarded

### MongoDB Split-Brain Recovery

MongoDB's primary-based architecture prevents true split-brain:

1. Primary is on one side of the partition
2. If the primary is on the minority side, it steps down
3. Majority side elects a new primary
4. When partition heals, old primary rolls back any writes that weren't replicated to the majority
5. Rolled-back writes are saved to a `rollback/` directory for manual recovery

```mermaid
sequenceDiagram
    participant P1 as Old Primary (minority)
    participant S1 as Secondary (majority)
    participant S2 as Secondary (majority)
    
    Note over P1,S2: Network partition splits cluster
    
    Note over P1: Still thinks it's primary<br/>Accepts 5 writes
    
    Note over S1,S2: Elect new primary (S1)
    S1->>S1: Becomes new primary
    
    Note over P1,S2: Partition heals
    
    P1->>S1: Discovers S1 is primary
    P1->>P1: Rolls back 5 writes<br/>Saves to rollback dir
    P1->>P1: Becomes secondary<br/>Syncs from S1
```

---

## Practical Monitoring

### What to Monitor

| Metric | Healthy | Warning | Critical |
|--------|---------|---------|----------|
| Read repair rate | < 1% of reads | 1-5% | > 5% (data diverging frequently) |
| Pending repairs | 0 | 1-5 | > 10 (repair falling behind) |
| Hinted handoff queue | < 100 hints | 100-1000 | > 1000 (nodes not recovering) |
| Replica lag (MongoDB) | < 1 second | 1-10 seconds | > 10 seconds |

---

## Next

→ [05-split-brain-and-partition-tolerance.md](./05-split-brain-and-partition-tolerance.md) — A deeper look at network partitions: what actually happens when your cluster splits, and how different databases handle the impossible choices.
