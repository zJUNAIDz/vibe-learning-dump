# Write Concern & Read Concern — MongoDB's Consistency Knobs

---

## The Problem

You write data to MongoDB. Is it safe? That depends on what "safe" means to you:

- "The primary acknowledged it" — but it could crash before replicating
- "A majority of replicas have it" — now it survives individual node failures
- "It's on disk" — now it survives power loss

MongoDB doesn't make this choice for you. You configure it per-operation via **write concern** and **read concern**.

---

## Write Concern

Write concern controls **how many replicas must acknowledge a write** before MongoDB tells your application "OK."

```mermaid
graph TD
    subgraph "Replica Set (3 nodes)"
        P[Primary]
        S1[Secondary 1]
        S2[Secondary 2]
    end
    
    W1["w: 1<br/>Primary only"] --> P
    WM["w: 'majority'<br/>2 of 3 nodes"] --> P
    WM --> S1
    WA["w: 3<br/>All nodes"] --> P
    WA --> S1
    WA --> S2
    
    style W1 fill:#4ecdc4,color:#fff
    style WM fill:#f39c12,color:#fff
    style WA fill:#ff6b6b,color:#fff
```

| Write Concern | Meaning | Latency | Safety |
|--------------|---------|---------|--------|
| `w: 0` | Fire and forget — no acknowledgment | ~0ms | ❌ May be lost |
| `w: 1` | Primary acknowledged | ~2ms | ⚠️ Lost if primary crashes before replication |
| `w: "majority"` | Majority of replicas | ~5-15ms | ✅ Survives any single node failure |
| `w: 3` (all) | All replicas | ~10-30ms | ✅✅ Maximum durability |
| `j: true` | Written to journal (on-disk WAL) | +2-5ms | Survives power loss |

```typescript
// Social media post — speed matters, brief loss is acceptable
await db.collection('posts').insertOne(post, {
  writeConcern: { w: 1 }
});

// Financial transaction — must survive node failure
await db.collection('transactions').insertOne(txn, {
  writeConcern: { w: 'majority', j: true }
});

// Analytics event — fire and forget
await db.collection('events').insertOne(event, {
  writeConcern: { w: 0 }
});
```

### What happens with `w: 1` and a primary crash?

```mermaid
sequenceDiagram
    participant App
    participant Primary
    participant Secondary
    
    App->>Primary: INSERT (w: 1)
    Primary-->>App: OK ✅
    Note over Primary: Write is in memory,<br/>not yet replicated
    
    Primary-xPrimary: CRASH 💥
    
    Note over Secondary: Elected as new primary<br/>Write is LOST
    
    App->>Secondary: READ
    Secondary-->>App: Document not found ❌
```

This is why `w: "majority"` exists — it ensures the write reaches enough replicas that no single failure loses it.

---

## Read Concern

Read concern controls **what data a read operation can see**.

| Read Concern | Meaning | Use Case |
|-------------|---------|----------|
| `"local"` | Returns the most recent data on this node (may be uncommitted on secondaries) | Default. Fast. |
| `"available"` | Like local, but works during sharding migrations (may return orphaned data) | Sharded collections |
| `"majority"` | Returns only data committed on a majority of replicas | Strong reads |
| `"linearizable"` | Strongest — ensures you see the most recent majority-committed write | Financial, leader election |
| `"snapshot"` | Consistent point-in-time snapshot (used with transactions) | Multi-statement transactions |

```typescript
// Default read — fast, may read uncommitted data
const user = await db.collection('users').findOne(
  { _id: userId },
  { readConcern: { level: 'local' } }
);

// Strong read — only see committed data
const balance = await db.collection('accounts').findOne(
  { _id: accountId },
  { readConcern: { level: 'majority' } }
);

// Causal consistency (session-level)
const session = client.startSession({ causalConsistency: true });
await db.collection('users').updateOne(
  { _id: userId },
  { $set: { name: 'New Name' } },
  { session, writeConcern: { w: 'majority' } }
);
// This read WILL see the write above (causal ordering)
const updated = await db.collection('users').findOne(
  { _id: userId },
  { session, readConcern: { level: 'majority' } }
);
```

---

## Read Preference

Separate from read concern, **read preference** controls **which node** serves reads:

```mermaid
graph TD
    subgraph "Replica Set"
        P[Primary<br/>Read/Write]
        S1[Secondary 1<br/>Read only]
        S2[Secondary 2<br/>Read only]
    end
    
    RP1["primary<br/>All reads go to primary"] --> P
    RP2["secondary<br/>Reads go to secondaries"] --> S1
    RP2 --> S2
    RP3["nearest<br/>Reads go to lowest latency node"] --> P
    RP3 --> S1
    RP3 --> S2
```

| Read Preference | When to Use | Warning |
|----------------|-------------|---------|
| `primary` (default) | Consistency required | All read load on primary |
| `primaryPreferred` | Prefer primary, fallback to secondary | Brief staleness during failover |
| `secondary` | Offload read traffic | Data may be stale by replication lag |
| `secondaryPreferred` | Read scaling, tolerate staleness | Usually milliseconds stale |
| `nearest` | Minimize latency (multi-region) | Any node, any staleness |

---

## Combining Write Concern + Read Concern for Guarantees

```mermaid
graph TD
    subgraph "Guarantee Levels"
        L1["w: 1, readConcern: local<br/>Fastest, weakest<br/>Default MongoDB behavior"]
        L2["w: majority, readConcern: majority<br/>Strong consistency<br/>Reads only committed data"]
        L3["w: majority, j: true<br/>readConcern: linearizable<br/>Strongest possible<br/>Highest latency"]
    end
    
    L1 --> U1[Social media, analytics,<br/>non-critical data]
    L2 --> U2[E-commerce, user accounts,<br/>important business data]
    L3 --> U3[Financial transactions,<br/>leader election]
    
    style L1 fill:#4ecdc4,color:#fff
    style L2 fill:#f39c12,color:#fff
    style L3 fill:#ff6b6b,color:#fff
```

### The "Read Your Own Writes" Problem

```typescript
// Without causal consistency:
await db.collection('users').updateOne(
  { _id: userId },
  { $set: { name: 'New Name' } },
  { writeConcern: { w: 1 } }
);

// If this read goes to a secondary (readPreference: secondary)...
const user = await db.collection('users').findOne({ _id: userId });
// ...it might return the OLD name (secondary hasn't replicated yet!)
```

**Fix: Use causal sessions**

```typescript
const session = client.startSession({ causalConsistency: true });

await db.collection('users').updateOne(
  { _id: userId },
  { $set: { name: 'New Name' } },
  { session, writeConcern: { w: 'majority' } }
);

// Guaranteed to see the write above, even from a secondary
const user = await db.collection('users').findOne(
  { _id: userId },
  { session, readConcern: { level: 'majority' } }
);
await session.endSession();
```

---

## Per-Operation Configuration

The most important insight: **you don't choose one consistency level for your entire application.** Different operations need different guarantees.

```typescript
// Configuration at the collection level (defaults)
const postsCollection = db.collection('posts', {
  writeConcern: { w: 1 },
  readConcern: { level: 'local' }
});

const accountsCollection = db.collection('accounts', {
  writeConcern: { w: 'majority', j: true },
  readConcern: { level: 'majority' }
});

// Or override per-operation
await postsCollection.insertOne(post);  // Uses w: 1 (default)
await accountsCollection.updateOne(     // Uses w: majority (default)
  { _id: accountId },
  { $inc: { balance: -amount } }
);
```

---

## Summary

| Question | Answer With |
|----------|------------|
| Can this write be lost if one node crashes? | Write concern (`w: "majority"`) |
| Can this write be lost on power failure? | Journal (`j: true`) |
| Can this read return uncommitted data? | Read concern (`"majority"`) |
| Can this read see stale data? | Read preference (`primary`) |
| Must I see my own writes? | Causal sessions |

**Default MongoDB is `w: 1`, `readConcern: "local"`, `readPreference: primary`.** This is fast but provides weaker guarantees than SQL's ACID defaults. Tune accordingly.

---

## Next

→ [10-performance-debugging.md](./10-performance-debugging.md) — How to find and fix slow queries in production.
