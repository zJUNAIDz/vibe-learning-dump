# Replication and Scaling

## Introduction

Your application is growing. A single database server can't handle the load.

**Common scaling strategies**:
1. **Vertical scaling**: Bigger machine (more CPU, RAM, disk)
2. **Read replicas**: Route reads to replica servers
3. **Connection pooling**: Reduce per-connection overhead
4. **Sharding**: Split data across multiple databases

But replication isn't magic. It introduces:
- **Replication lag**: Replicas are behind the primary
- **Stale reads**: Reading old data from replicas
- **Operational complexity**: Monitoring, failover, consistency

This chapter explains how replication works, when to use it, and how to avoid common pitfalls.

---

## Primary-Replica Model

The most common replication setup:

```
┌─────────────────┐
│   Primary DB    │
│  (read/write)   │
└────────┬────────┘
         │ WAL stream
         ├────────────────────────┐
         ▼                        ▼
┌────────────────┐       ┌────────────────┐
│   Replica 1     │       │   Replica 2     │
│  (read-only)    │       │  (read-only)    │
└────────────────┘       └────────────────┘
```

**Primary**:
- Accepts writes (`INSERT`, `UPDATE`, `DELETE`)
- Accepts reads

**Replicas**:
- Read-only
- Continuously apply changes from the primary
- **Lag** behind the primary (milliseconds to seconds)

---

## How Replication Works (PostgreSQL)

### Write-Ahead Log (WAL)

Every write operation is logged to the **WAL** (a sequential log file) before being applied to the data files.

**WAL entries** describe changes:
```
WAL Entry 1: INSERT INTO users (id, name) VALUES (1, 'Alice')
WAL Entry 2: UPDATE users SET balance = 100 WHERE id = 1
WAL Entry 3: DELETE FROM users WHERE id = 2
```

### Streaming Replication

The primary **streams** WAL entries to replicas in real-time.

**Steps**:
1. Primary writes changes to WAL
2. Primary streams WAL to replicas
3. Replicas apply WAL entries to their local data files

**Result**: Replicas eventually have the same data as the primary.

### Synchronous vs Asynchronous Replication

#### Asynchronous (Default)

**Flow**:
1. Client sends `INSERT`
2. Primary writes to WAL
3. Primary **commits immediately** (doesn't wait for replicas)
4. Primary streams WAL to replicas (in the background)

**Pros**: Fast (low latency)
**Cons**: Replicas lag behind. If primary crashes before WAL reaches replicas, data loss is possible.

#### Synchronous

**Flow**:
1. Client sends `INSERT`
2. Primary writes to WAL
3. Primary waits for at least one replica to confirm receiving the WAL
4. Primary commits

**Pros**: No data loss (replicas are up-to-date)
**Cons**: Higher latency (each commit waits for network round-trip)

**Configuration** (`postgresql.conf`):
```conf
synchronous_commit = on
synchronous_standby_names = 'replica1'
```

---

## Replication Lag

**Replication lag**: The time delay between a write on the primary and its visibility on a replica.

### Causes

1. **Network latency**: WAL takes time to transfer
2. **Replica is slow**: Disk I/O, CPU, or query load on the replica
3. **Long transactions**: Large transactions take time to apply
4. **Hot standby feedback**: Replica delays applying changes to avoid killing queries

### Measuring Lag

**Primary**:
```sql
SELECT 
  client_addr, 
  state, 
  sync_state,
  pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn) AS lag_bytes,
  EXTRACT(EPOCH FROM (now() - replay_time)) AS lag_seconds
FROM pg_stat_replication;
```

**Output**:
```
client_addr | state     | sync_state | lag_bytes | lag_seconds
------------|-----------|------------|-----------|-------------
10.0.1.5    | streaming | async      | 512000    | 2.3
```

**Interpretation**: Replica is 512 KB behind (2.3 seconds of lag).

**Acceptable lag**: <1 second for most applications. >10 seconds indicates a problem.

---

## Stale Reads

**Stale read**: Reading old data from a replica.

### Example

```sql
-- On primary:
INSERT INTO orders (user_id, total) VALUES (1, 100);
COMMIT;

-- Immediately on replica:
SELECT * FROM orders WHERE user_id = 1;
-- Returns: 0 rows (replication lag!)
```

**Problem**: User places an order, refreshes the page, and doesn't see their order.

### Solutions

#### 1. **Read Your Own Writes** (Session Stickiness)

After a write, route subsequent reads from the same user to the primary (not replicas).

**Implementation**:
- Use a session cookie or header to mark "this session just wrote data"
- Route those requests to the primary for the next N seconds

#### 2. **Check Replication Lag Before Reading**

Before reading from a replica, check if it's up-to-date:

**Primary** (after write):
```sql
SELECT pg_current_wal_lsn();  -- Returns: 0/3000000
```

**Replica** (before read):
```sql
SELECT pg_last_wal_replay_lsn();  -- Returns: 0/2FFFFFF (behind!)
-- Wait until it catches up
```

**Application logic**:
```javascript
// After write on primary
const lsn = await primary.query('SELECT pg_current_wal_lsn()');

// Before read on replica
while (true) {
  const replicaLsn = await replica.query('SELECT pg_last_wal_replay_lsn()');
  if (replicaLsn >= lsn) break;
  await sleep(10);  // Wait 10ms
}
const result = await replica.query('SELECT * FROM orders WHERE user_id = 1');
```

#### 3. **Use Synchronous Replication**

If you can afford the latency, use synchronous replication:
```conf
synchronous_commit = on
```

Writes are only committed after at least one replica is up-to-date.

---

## Logical vs Physical Replication

### Physical Replication (Binary)

Replicas receive **WAL** (low-level binary data).

**Characteristics**:
- Entire database is replicated (can't replicate a subset)
- Replicas must be the same PostgreSQL version
- Replicas are **read-only**

**Use case**: Disaster recovery, read scaling.

### Logical Replication (Row-Level)

Replicas receive **logical changes** (e.g., "insert row X into table Y").

**Characteristics**:
- Can replicate a subset of tables
- Replicas can be different PostgreSQL versions
- Replicas can be **writable** (for different data)

**Use case**: 
- Cross-datacenter replication
- Upgrading PostgreSQL (replicate from v13 to v14)
- Selective replication (replicate only some tables)

**Configuration**:
```sql
-- On primary:
CREATE PUBLICATION my_pub FOR TABLE users, orders;

-- On replica:
CREATE SUBSCRIPTION my_sub CONNECTION 'host=primary dbname=mydb' PUBLICATION my_pub;
```

---

## Why "Read Scaling" Is Not Free

**Myth**: "Add more replicas = handle more load."

**Reality**: Replicas help with **read-heavy** workloads. But:

### 1. **Writes Don't Scale**

All writes go to the primary. Adding replicas doesn't help.

**If your bottleneck is writes**, replicas won't help. You need sharding or a different architecture.

### 2. **Replication Overhead**

The primary must:
- Generate WAL
- Stream WAL to all replicas

**More replicas = more network bandwidth, more CPU on primary.**

**Practical limit**: ~10 replicas per primary (depends on hardware).

### 3. **Complexity**

With replicas, you must:
- Route queries to the correct server (primary vs replica)
- Handle replication lag
- Monitor replica health
- Handle failover (if primary dies)

---

## Connection Pooling

**Problem**: Each PostgreSQL connection consumes ~10 MB of RAM. With 1000 connections, that's 10 GB of RAM just for connections.

**Solution**: Connection pooling.

### What Is a Connection Pool?

A middleman that maintains a **pool** of database connections and reuses them.

```
┌──────────────────────────────┐
│   1000 Application Threads   │
└──────────────┬───────────────┘
               │
               ▼
┌──────────────────────────────┐
│    Connection Pool (PgBouncer, │
│     PgPool, etc.)              │
│  → Maintains 100 connections   │
└──────────────┬───────────────┘
               │
               ▼
┌──────────────────────────────┐
│       PostgreSQL             │
│   100 active connections     │
└──────────────────────────────┘
```

**Result**: 1000 app threads share 100 database connections.

### PgBouncer

The most popular PostgreSQL connection pooler.

**Modes**:
1. **Session mode**: Assigns a connection for the duration of a session
2. **Transaction mode**: Assigns a connection for the duration of a transaction
3. **Statement mode**: Assigns a connection for a single query

**Best for most use cases**: Transaction mode.

**Configuration** (`pgbouncer.ini`):
```ini
[databases]
mydb = host=localhost dbname=mydb

[pgbouncer]
pool_mode = transaction
max_client_conn = 1000
default_pool_size = 20
```

**Result**: 1000 clients share 20 connections.

---

## Failover and High Availability

**Failover**: Promoting a replica to primary when the primary fails.

### Manual Failover

1. Primary crashes
2. DBA promotes a replica:
   ```bash
   pg_ctl promote -D /var/lib/postgresql/data
   ```
3. Application reconnects to the new primary

**Downtime**: Minutes (manual intervention required).

### Automatic Failover (Patroni, Repmgr)

Tools like **Patroni** monitor the primary and automatically promote a replica if the primary fails.

**Flow**:
1. Primary crashes
2. Patroni detects failure (health check)
3. Patroni promotes the most up-to-date replica
4. Patroni updates DNS or virtual IP to point to the new primary

**Downtime**: Seconds (automatic).

**Caveat**: If replication was asynchronous, some data may be lost (the last few seconds of writes).

---

## Postgres vs MySQL Replication

| Feature                    | PostgreSQL                    | MySQL                        |
|----------------------------|-------------------------------|------------------------------|
| Replication type           | WAL-based (physical)          | Binlog-based (logical)       |
| Replica lag monitoring     | `pg_stat_replication`         | `SHOW SLAVE STATUS`          |
| Synchronous replication    | Yes                           | Yes (semi-sync)              |
| Logical replication        | Yes (PostgreSQL 10+)          | Yes (always)                 |
| Cascading replication      | Yes                           | Yes                          |
| Read-only replicas         | Yes                           | Yes                          |
| Replica can lag behind     | Yes                           | Yes                          |
| Automatic failover         | Requires tool (Patroni, etc.) | Requires tool (Orchestrator, MHA) |

**Key difference**: MySQL uses **binlog** (logical changes), PostgreSQL uses **WAL** (physical changes). But functionally they're similar.

---

## Sharding (Manual)

**Sharding**: Splitting data across multiple databases.

### Example: User Sharding

**Shard by user ID**:
- Users 0-999 → Database 1
- Users 1000-1999 → Database 2
- Users 2000-2999 → Database 3

**Application logic**:
```javascript
function getDbForUser(userId) {
  const shardId = Math.floor(userId / 1000);
  return databases[shardId];
}

const db = getDbForUser(1234);  // Database 2
await db.query('SELECT * FROM users WHERE id = 1234');
```

### Challenges

1. **Cross-shard queries**: Can't join data across shards
2. **Rebalancing**: If a shard grows too large, you must migrate data
3. **Consistency**: Transactions can't span shards

**When to shard**: Only if a single database can't handle the load (rare for OLTP workloads).

---

## Postgres Scaling Strategies (Summary)

| Strategy               | Scales Reads? | Scales Writes? | Complexity | Use Case                    |
|------------------------|---------------|----------------|------------|------------------------------|
| Vertical scaling       | Yes           | Yes            | Low        | Default (upgrade hardware)   |
| Read replicas          | Yes           | No             | Medium     | Read-heavy workloads         |
| Connection pooling     | Yes           | Yes            | Low        | High connection count        |
| Caching (Redis, etc.)  | Yes           | No             | Medium     | Frequently-read data         |
| Sharding               | Yes           | Yes            | High       | Massive scale (last resort)  |

**General advice**: Start with vertical scaling + connection pooling. Add replicas if read-heavy. Shard only if absolutely necessary.

---

## Practical Takeaways

### 1. **Use Asynchronous Replication by Default**

Synchronous replication adds latency. Only use it if you can't tolerate any data loss.

### 2. **Monitor Replication Lag**

```sql
SELECT client_addr, lag_bytes, lag_seconds FROM pg_stat_replication;
```

Alert if lag > 10 seconds.

### 3. **Handle Stale Reads**

Options:
- Route "read your own writes" to the primary
- Check replica lag before reading
- Use synchronous replication

### 4. **Use Connection Pooling**

PgBouncer reduces connection overhead by 10×.

### 5. **Don't Over-Replicate**

Each replica adds overhead to the primary. Start with 1-2 replicas, add more only if needed.

### 6. **Plan for Failover**

Use tools like Patroni for automatic failover. Test failover regularly.

### 7. **Shard Only as a Last Resort**

Sharding is operationally complex. Most applications don't need it.

---

## Summary

- Replication copies data from a primary to replicas
- Replicas are **read-only** and lag behind the primary (milliseconds to seconds)
- Asynchronous replication is fast but risks data loss; synchronous is safe but slower
- Stale reads happen when reading from lagging replicas
- Connection pooling reduces overhead for high-connection workloads
- Read replicas scale **reads**, not writes
- Sharding scales both reads and writes but is operationally complex
- Use tools like Patroni for automatic failover
- Monitor replication lag with `pg_stat_replication`

Understanding replication helps you:
- Scale your application effectively
- Avoid stale read bugs
- Plan for disaster recovery
- Design a resilient architecture
