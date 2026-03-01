# Eventual vs Strong Consistency — What the Words Actually Mean

---

## The Confusion

"Eventual consistency" is the most misunderstood term in distributed databases. People hear "eventual" and think "unreliable." That's wrong — but the reality is nuanced.

---

## Strong Consistency

**Definition**: After a write completes, every subsequent read — from any client, on any node, anywhere — returns that write or a later one. Always. No exceptions.

```mermaid
sequenceDiagram
    participant A as Client A
    participant DB as Database
    participant B as Client B
    
    A->>DB: WRITE name = "Alice"
    DB-->>A: OK ✅
    
    Note over DB: From this moment,<br/>ALL reads see "Alice"
    
    B->>DB: READ name
    DB-->>B: "Alice" ✅
    
    Note right of B: Always "Alice".<br/>Never stale.
```

PostgreSQL, MySQL (single-machine), CockroachDB, and Spanner provide this. The write isn't acknowledged until all replicas (or a quorum in distributed systems) have the data.

**Cost**: Latency. Every write waits for replication. Every read might wait for coordination. In a multi-region setup, strong consistency adds 50-200ms per operation.

---

## Eventual Consistency

**Definition**: After a write completes, replicas **will converge** to the same value — but not immediately. For some window of time, different clients may read different values.

```mermaid
sequenceDiagram
    participant A as Client A (US East)
    participant N1 as Node 1 (US East)
    participant N2 as Node 2 (EU West)
    participant B as Client B (EU West)
    
    A->>N1: WRITE name = "Alice"
    N1-->>A: OK ✅
    
    Note over N1,N2: Replication in progress...<br/>~100ms across Atlantic
    
    B->>N2: READ name
    N2-->>B: "Bob" (old value! ⚠️)
    
    Note over N1,N2: Replication completes
    
    B->>N2: READ name
    N2-->>B: "Alice" ✅
```

For a window (milliseconds to seconds), Client B sees stale data. After convergence, everyone sees the same value.

"Eventual" means:
- **Not "never"** — convergence happens, typically within milliseconds
- **Not "random"** — there are formal guarantees about convergence
- **Not "wrong"** — the data was correct at the time it was written; it just hasn't propagated everywhere yet

---

## The Spectrum Between Them

Consistency isn't binary. It's a spectrum:

```mermaid
graph LR
    subgraph "Consistency Spectrum"
        direction LR
        EV["Eventual<br/>Consistency"]
        MW["Monotonic<br/>Writes"]
        MR["Monotonic<br/>Reads"]
        RYW["Read Your<br/>Own Writes"]
        CS["Causal<br/>Consistency"]
        SEQ["Sequential<br/>Consistency"]
        LIN["Linearizability<br/>(Strong)"]
    end
    
    EV --- MW --- MR --- RYW --- CS --- SEQ --- LIN
    
    style EV fill:#ff6b6b,color:#fff
    style RYW fill:#f39c12,color:#fff
    style CS fill:#e67e22,color:#fff
    style LIN fill:#4ecdc4,color:#fff
```

### Key Levels

**Read Your Own Writes**: After you write, YOUR reads always see that write. Other clients might not (yet). Most applications need at least this.

**Monotonic Reads**: If you read version 5, you'll never subsequently read version 4. Time doesn't go backward for you.

**Causal Consistency**: If operation A caused operation B (A happened-before B), everyone sees A before B. Example: if I post a message, then you reply, nobody sees the reply without the original message.

**Linearizability (Strong)**: All operations appear to execute atomically, in a single total order. The gold standard — and the most expensive.

---

## How Each Database Handles It

| Database | Default Consistency | Strongest Available | How to Get Strong |
|----------|-------------------|--------------------|--------------------|
| PostgreSQL | Strong (linearizable) | Linearizable | It's the default |
| MongoDB | Read-your-own-writes | Linearizable | `readConcern: "linearizable"` |
| Cassandra | Tunable (per-query) | Linearizable (single partition) | Read + Write at QUORUM; or LWT |
| DynamoDB | Eventual | Strong (per-table) | `ConsistentRead: true` |
| Redis | Eventual (replicas) | Strong (single node) | Don't use replicas for reads |
| CockroachDB | Serializable | Serializable | It's the default |

---

## When Eventual Consistency Causes Real Problems

### Problem 1: The Double-Submit

```mermaid
sequenceDiagram
    participant User
    participant LB as Load Balancer
    participant N1 as Node 1 (Primary)
    participant N2 as Node 2 (Replica)
    
    User->>LB: Submit payment
    LB->>N1: Process payment
    N1-->>LB: Success ✅
    LB-->>User: Payment confirmed
    
    Note over User: User refreshes page
    
    User->>LB: GET /payments
    LB->>N2: Read payments (routed to replica)
    N2-->>LB: No payment found ⚠️
    LB-->>User: "No recent payments"
    
    Note right of User: User panics,<br/>submits payment again
```

The user sees "no payment" because the replica hasn't caught up. They submit again → duplicate charge.

**Fix**: Read-your-own-writes guarantee. After writing, read from the primary (or use a session token to ensure routing).

### Problem 2: The Counter Race

```
User A reads counter: 100
User B reads counter: 100
User A writes counter: 101
User B writes counter: 101  ← Should be 102!
```

With eventual consistency, concurrent increments can lose updates.

**Fix**: Use actual counters (Cassandra `COUNTER` type), atomic increment operations (Redis `INCR`), or transactions.

### Problem 3: Referential Integrity Violation

```mermaid
sequenceDiagram
    participant App
    participant N1 as Node 1
    participant N2 as Node 2
    
    App->>N1: Create user (id=42)
    N1-->>App: OK
    
    App->>N2: Create order (user_id=42)
    Note over N2: User 42 hasn't<br/>replicated here yet!
    N2-->>App: OK (no foreign key check)
    
    Note over N1,N2: Order exists on N2<br/>but user doesn't exist on N2 yet<br/>Inconsistent window
```

NoSQL databases generally don't enforce foreign keys. In eventually consistent systems, referencing another entity can fail even in the brief inconsistency window.

**Fix**: Application-level validation, or accept that NoSQL databases don't provide referential integrity.

---

## When Eventual Consistency Is Fine

The majority of operations don't need strong consistency. Ask: "If the user sees data that's 2 seconds stale, what happens?"

```mermaid
graph TD
    subgraph "Strong Consistency Needed"
        S1["Password changes"]
        S2["Account balances"]
        S3["Inventory counts<br/>(when stock is low)"]
        S4["Permission changes"]
    end
    
    subgraph "Eventual Consistency Fine"
        E1["Activity feeds"]
        E2["Like counts"]
        E3["Product recommendations"]
        E4["Search results"]
        E5["Analytics dashboards"]
        E6["User profile views"]
    end
    
    style S1 fill:#ff6b6b,color:#fff
    style S2 fill:#ff6b6b,color:#fff
    style S3 fill:#ff6b6b,color:#fff
    style S4 fill:#ff6b6b,color:#fff
    style E1 fill:#4ecdc4,color:#fff
    style E2 fill:#4ecdc4,color:#fff
    style E3 fill:#4ecdc4,color:#fff
    style E4 fill:#4ecdc4,color:#fff
    style E5 fill:#4ecdc4,color:#fff
    style E6 fill:#4ecdc4,color:#fff
```

For most applications, fewer than 20% of operations need strong consistency. The other 80% can tolerate seconds of staleness with no user-visible impact.

---

## Per-Operation Consistency in Practice

### TypeScript — Choosing Consistency Per Query

```typescript
import { MongoClient, ReadConcern, WriteConcern, ReadPreference } from 'mongodb';

const client = new MongoClient('mongodb://localhost:27017', {
  replicaSet: 'rs0',
});

const db = client.db('myapp');

// Strong: Password update — must be immediately visible
async function updatePassword(userId: string, newHash: string): Promise<void> {
  await db.collection('users').updateOne(
    { _id: userId },
    { $set: { passwordHash: newHash } },
    { 
      writeConcern: new WriteConcern('majority', 5000, true), // majority + journal
    }
  );
}

// Strong: Read password for authentication
async function getPasswordHash(userId: string): Promise<string | null> {
  const user = await db.collection('users').findOne(
    { _id: userId },
    { 
      readConcern: new ReadConcern('majority'),
      readPreference: ReadPreference.primary, // always read from primary
    }
  );
  return user?.passwordHash ?? null;
}

// Eventual: Activity feed — stale by seconds is fine
async function getActivityFeed(userId: string): Promise<any[]> {
  return db.collection('activity_feed')
    .find(
      { userId },
      {
        readConcern: new ReadConcern('local'), // fast, might be slightly stale
        readPreference: ReadPreference.secondaryPreferred, // read from replicas
      }
    )
    .sort({ timestamp: -1 })
    .limit(50)
    .toArray();
}
```

---

## The Consistency Tax

Strong consistency has real costs:

| Metric | Eventual (ONE/local) | Strong (QUORUM/majority) | Difference |
|--------|---------------------|-------------------------|------------|
| Write latency | 1-3ms | 5-15ms | 3-5x slower |
| Read latency | 1-3ms | 3-10ms | 2-3x slower |
| Availability during failure | High | Reduced | Can't write if majority down |
| Throughput | Higher | Lower | ~50% fewer ops/sec |

Every operation you mark as "must be strong" costs performance. Don't make everything strong out of fear — make it strong only when staleness causes actual harm.

---

## Next

→ [02-quorum-and-consensus.md](./02-quorum-and-consensus.md) — How distributed databases actually agree on the state of data: quorum voting, Paxos, and Raft.
