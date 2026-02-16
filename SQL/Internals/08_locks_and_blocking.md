# Locks and Blocking

## Introduction

```sql
-- Terminal 1:
BEGIN;
UPDATE users SET email = 'alice@new.com' WHERE id = 1;
-- (Don't commit yet)

-- Terminal 2:
UPDATE users SET name = 'Alice Smith' WHERE id = 1;
-- (Hangs...)
```

**Why is Terminal 2 blocked?**

The answer: **locks**.

Locks are the mechanism databases use to prevent conflicting concurrent access. They're essential for correctness, but they can also cause:
- Blocking (transactions waiting for locks)
- Deadlocks (transactions waiting for each other)
- Performance degradation

This chapter explains how locks work, when they're acquired, and how to avoid lock-related disasters.

---

## What Are Locks?

A lock is a mechanism to control concurrent access to a resource (row, table, index, etc.).

**Purpose**: Ensure data integrity when multiple transactions access the same data.

### Types of Locks

1. **Row-level locks**: Lock individual rows
2. **Table-level locks**: Lock entire tables
3. **Advisory locks**: Application-defined locks (not tied to data)

---

## Row-Level Locks

### Implicit Row Locks

PostgreSQL automatically acquires row locks when you:
- **UPDATE** a row
- **DELETE** a row
- **SELECT FOR UPDATE** a row

```sql
BEGIN;
UPDATE users SET email = 'alice@new.com' WHERE id = 1;
-- Acquires an exclusive row lock on row id=1
```

**Any other transaction** trying to update the same row must wait.

### Lock Modes

PostgreSQL has several row-level lock modes:

| Lock Mode           | Command                   | Blocks                     |
|---------------------|---------------------------|----------------------------|
| `FOR UPDATE`        | `SELECT ... FOR UPDATE`   | Other `FOR UPDATE`, `UPDATE`, `DELETE` |
| `FOR NO KEY UPDATE` | `SELECT ... FOR NO KEY UPDATE` | Same as `FOR UPDATE` except doesn't block FK checks |
| `FOR SHARE`         | `SELECT ... FOR SHARE`    | `UPDATE`, `DELETE`, `FOR UPDATE` |
| `FOR KEY SHARE`     | `SELECT ... FOR KEY SHARE` | `UPDATE` (only on key columns) |

### Example: SELECT FOR UPDATE

**Use case**: Prevent lost updates.

```sql
-- Transaction A:
BEGIN;
SELECT balance FROM accounts WHERE id = 1 FOR UPDATE;
-- Result: 100 (and acquires lock)
-- No other transaction can update this row until we commit
UPDATE accounts SET balance = balance - 50 WHERE id = 1;
COMMIT;
```

**Without `FOR UPDATE`**:
```sql
-- Transaction A:
BEGIN;
SELECT balance FROM accounts WHERE id = 1;  -- balance = 100
-- (Another transaction updates the row here)
UPDATE accounts SET balance = 50 WHERE id = 1;  -- Overwrites the other update!
COMMIT;
```

**Lost update** (race condition).

**With `FOR UPDATE`**, the other transaction waits until Transaction A commits.

---

## Table-Level Locks

PostgreSQL uses table-level locks for DDL and some DML operations.

### Lock Modes

| Lock Mode                | Acquired By               | Blocks                     |
|--------------------------|---------------------------|----------------------------|
| `ACCESS SHARE`           | `SELECT`                  | `ACCESS EXCLUSIVE`         |
| `ROW SHARE`              | `SELECT FOR UPDATE`       | `EXCLUSIVE`, `ACCESS EXCLUSIVE` |
| `ROW EXCLUSIVE`          | `UPDATE`, `DELETE`, `INSERT` | `SHARE`, `SHARE ROW EXCLUSIVE`, `EXCLUSIVE`, `ACCESS EXCLUSIVE` |
| `SHARE UPDATE EXCLUSIVE` | `VACUUM`, `CREATE INDEX CONCURRENTLY` | `SHARE UPDATE EXCLUSIVE`, `SHARE`, `SHARE ROW EXCLUSIVE`, `EXCLUSIVE`, `ACCESS EXCLUSIVE` |
| `SHARE`                  | `CREATE INDEX`            | `ROW EXCLUSIVE`, `SHARE UPDATE EXCLUSIVE`, `SHARE ROW EXCLUSIVE`, `EXCLUSIVE`, `ACCESS EXCLUSIVE` |
| `SHARE ROW EXCLUSIVE`    | Rare                      | `ROW EXCLUSIVE`, `SHARE UPDATE EXCLUSIVE`, `SHARE`, `SHARE ROW EXCLUSIVE`, `EXCLUSIVE`, `ACCESS EXCLUSIVE` |
| `EXCLUSIVE`              | `REFRESH MATERIALIZED VIEW` | All except `ACCESS SHARE` |
| `ACCESS EXCLUSIVE`       | `DROP TABLE`, `TRUNCATE`, `ALTER TABLE` | All other locks            |

**Key takeaway**: Most queries acquire weak locks (`ACCESS SHARE`, `ROW EXCLUSIVE`) that don't conflict much. DDL operations acquire strong locks (`ACCESS EXCLUSIVE`) that block everything.

### Example: ALTER TABLE Blocks Everything

```sql
-- Terminal 1:
ALTER TABLE users ADD COLUMN last_login TIMESTAMP;
-- Acquires ACCESS EXCLUSIVE lock (blocks all queries)

-- Terminal 2:
SELECT * FROM users WHERE id = 1;
-- Blocked until ALTER TABLE finishes
```

**Problem**: Adding a column locks the table (in older PostgreSQL versions). On a large table, this can take minutes.

**Fix** (PostgreSQL 11+):
- Adding a column with a `DEFAULT` value locks the table (rewrites all rows)
- Adding a column **without** a default is instant (metadata-only change)

```sql
-- Slow (locks table and rewrites):
ALTER TABLE users ADD COLUMN last_login TIMESTAMP DEFAULT NOW();

-- Fast (metadata-only):
ALTER TABLE users ADD COLUMN last_login TIMESTAMP;
UPDATE users SET last_login = NOW() WHERE last_login IS NULL;  -- Backfill later
```

---

## Explicit Locking

You can explicitly acquire locks:

### LOCK TABLE

```sql
BEGIN;
LOCK TABLE users IN ACCESS EXCLUSIVE MODE;
-- No other transaction can read or write to `users`
-- Do some critical operations
COMMIT;
```

**When to use**: Rarely. Only for maintenance operations where you need exclusive access.

### SELECT FOR UPDATE

```sql
BEGIN;
SELECT * FROM users WHERE id = 1 FOR UPDATE;
-- Acquires row lock
COMMIT;
```

**When to use**: Prevent lost updates.

### SELECT FOR SHARE

```sql
BEGIN;
SELECT * FROM users WHERE id = 1 FOR SHARE;
-- Acquires shared row lock (other transactions can also read, but not update)
COMMIT;
```

**When to use**: Rarely. Useful if you want to ensure a row doesn't change while you do some computation, but don't need to update it yourself.

---

## Deadlocks

**Deadlock**: Two (or more) transactions wait for each other indefinitely.

### Example

```sql
-- Transaction A:
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;  -- Locks row 1
-- (Try to lock row 2)
UPDATE accounts SET balance = balance + 100 WHERE id = 2;  -- Waits...

-- Transaction B (concurrent):
BEGIN;
UPDATE accounts SET balance = balance - 50 WHERE id = 2;   -- Locks row 2
-- (Try to lock row 1)
UPDATE accounts SET balance = balance + 50 WHERE id = 1;   -- Waits...
```

**Both transactions are waiting for each other.**

**PostgreSQL detects the deadlock after 1 second** (configurable via `deadlock_timeout`):
```
ERROR:  deadlock detected
DETAIL:  Process 1234 waits for ShareLock on transaction 5678; blocked by process 5678.
Process 5678 waits for ShareLock on transaction 1234; blocked by process 1234.
HINT:  See server log for query details.
```

**One transaction is aborted** (randomly chosen). The other proceeds.

### How to Avoid Deadlocks

1. **Lock rows in a consistent order**

```sql
-- Always lock rows in ascending ID order
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
UPDATE accounts SET balance = balance + 100 WHERE id = 2;
COMMIT;
```

If all transactions lock rows in the same order, deadlocks can't occur.

2. **Use shorter transactions**

Deadlocks are more likely in long transactions (more time for conflicts).

3. **Retry on deadlock**

```javascript
while (true) {
  try {
    await db.query('BEGIN');
    // ... your queries
    await db.query('COMMIT');
    break;
  } catch (err) {
    if (err.code === '40P01') {  // Deadlock detected
      continue;  // Retry
    }
    throw err;
  }
}
```

---

## Lock Contention and Hot Rows

**Hot row**: A row that many transactions try to update simultaneously.

### Example: Global Counter

```sql
-- 100 concurrent connections all do:
UPDATE counters SET value = value + 1 WHERE id = 1;
```

**Problem**: All 100 transactions try to lock the same row. They serialize (one at a time).

**Throughput**: Limited by lock acquisition speed (~10,000 lock/unlocks per second).

### Solutions

#### 1. Sharding

Split the counter into multiple rows:
```sql
CREATE TABLE counters (
  id SERIAL PRIMARY KEY,
  shard INT,
  value BIGINT
);

-- Insert 10 shards
INSERT INTO counters (shard, value) SELECT i, 0 FROM generate_series(1, 10) i;

-- Increment a random shard
UPDATE counters SET value = value + 1 WHERE shard = floor(random() * 10) + 1;

-- Get total count
SELECT SUM(value) FROM counters;
```

**Result**: Lock contention is spread across 10 rows. 10× higher throughput.

#### 2. Application-Level Aggregation

Don't update the database on every increment. Instead:
- Accumulate increments in memory (or Redis)
- Flush to the database periodically (e.g., every 10 seconds)

#### 3. Use Specialized Counters

PostgreSQL `pg_stat_statements` and similar extensions use lock-free counters for high throughput.

---

## Lock Queues and Queue Jumping

When a lock is held, other transactions **queue** for it.

**FIFO (First In, First Out)**: The first transaction to request the lock gets it next.

**But**: Some lock modes can "jump the queue."

### Example

```sql
-- Transaction A:
BEGIN;
UPDATE users SET email = 'alice@new.com' WHERE id = 1;  -- Holds exclusive lock
-- (Don't commit)

-- Transaction B:
SELECT * FROM users WHERE id = 1;  -- Waits for lock

-- Transaction C:
ALTER TABLE users ADD COLUMN last_login TIMESTAMP;  -- Waits for ACCESS EXCLUSIVE

-- Transaction A commits:
COMMIT;
```

**Who gets the lock next?**

Answer: **Transaction C** (ACCESS EXCLUSIVE jumps the queue and blocks everything).

**Result**: Transaction B waits even longer.

**Implication**: DDL operations can block queries even if they're not running yet.

---

## Monitoring Locks

### pg_locks (PostgreSQL)

```sql
SELECT 
  locktype, 
  relation::regclass, 
  mode, 
  granted, 
  pid
FROM pg_locks
WHERE NOT granted;
```

**Fields**:
- `locktype`: `relation` (table), `tuple` (row), etc.
- `relation`: Which table
- `mode`: Lock mode (e.g., `RowExclusiveLock`, `AccessExclusiveLock`)
- `granted`: `false` if waiting
- `pid`: Process ID

### Finding Blocking Queries

```sql
SELECT 
  blocked.pid AS blocked_pid,
  blocked.query AS blocked_query,
  blocking.pid AS blocking_pid,
  blocking.query AS blocking_query
FROM pg_stat_activity AS blocked
JOIN pg_locks AS blocked_locks ON blocked.pid = blocked_locks.pid
JOIN pg_locks AS blocking_locks ON blocked_locks.locktype = blocking_locks.locktype
  AND blocked_locks.database IS NOT DISTINCT FROM blocking_locks.database
  AND blocked_locks.relation IS NOT DISTINCT FROM blocking_locks.relation
  AND blocked_locks.page IS NOT DISTINCT FROM blocking_locks.page
  AND blocked_locks.tuple IS NOT DISTINCT FROM blocking_locks.tuple
  AND blocked_locks.virtualxid IS NOT DISTINCT FROM blocking_locks.virtualxid
  AND blocked_locks.transactionid IS NOT DISTINCT FROM blocking_locks.transactionid
  AND blocked_locks.classid IS NOT DISTINCT FROM blocking_locks.classid
  AND blocked_locks.objid IS NOT DISTINCT FROM blocking_locks.objid
  AND blocked_locks.objsubid IS NOT DISTINCT FROM blocking_locks.objsubid
  AND blocked_locks.pid != blocking_locks.pid
JOIN pg_stat_activity AS blocking ON blocking_locks.pid = blocking.pid
WHERE NOT blocked_locks.granted;
```

**Simpler heuristic**:
```sql
SELECT pid, query, state, now() - query_start AS duration
FROM pg_stat_activity
WHERE state = 'active' AND now() - query_start > interval '1 minute';
```

Any query running for >1 minute is suspicious.

---

## Why "Simple" Updates Block Everything

```sql
UPDATE users SET last_login = NOW() WHERE id = 1;
```

This looks innocent, but:
1. **Acquires a `ROW EXCLUSIVE` lock on the table** (prevents DDL)
2. **Acquires an exclusive lock on row `id = 1`** (blocks other updates to that row)

**If this update is slow** (e.g., in a long transaction), other queries are blocked.

### Common Causes

#### 1. **Long Transaction**

```sql
BEGIN;
UPDATE users SET last_login = NOW() WHERE id = 1;
-- (Do some application logic that takes 10 seconds)
COMMIT;
```

The lock is held for 10 seconds.

**Fix**: Keep transactions short. Commit frequently.

#### 2. **Large Batch Update**

```sql
UPDATE users SET verified = true WHERE created_at < '2020-01-01';
-- Updates 1 million rows (takes 10 seconds)
```

This locks 1 million rows simultaneously.

**Fix**: Batch the update:
```sql
DO $$
BEGIN
  LOOP
    UPDATE users SET verified = true 
    WHERE id IN (SELECT id FROM users WHERE created_at < '2020-01-01' AND NOT verified LIMIT 1000);
    EXIT WHEN NOT FOUND;
    COMMIT;
  END LOOP;
END $$;
```

#### 3. **Accidental Full-Table Lock**

```sql
ALTER TABLE users ADD COLUMN last_login TIMESTAMP DEFAULT NOW();
```

This rewrites the entire table (locks everything).

**Fix**: Add the column without a default:
```sql
ALTER TABLE users ADD COLUMN last_login TIMESTAMP;
```

---

## ORM-Induced Locking Disasters

### Problem 1: N+1 Locks

**ORM code**:
```javascript
const users = await User.findAll();
for (const user of users) {
  await user.update({ last_login: new Date() });
}
```

**Generated SQL**:
```sql
SELECT * FROM users;
UPDATE users SET last_login = '...' WHERE id = 1;
UPDATE users SET last_login = '...' WHERE id = 2;
...
UPDATE users SET last_login = '...' WHERE id = 1000;
```

**Problem**: 1000 separate UPDATE statements. Each acquires a lock, waits for a round-trip, etc.

**Fix**: Batch update:
```sql
UPDATE users SET last_login = NOW();
```

### Problem 2: Implicit Transactions

**ORM code**:
```javascript
await sequelize.transaction(async (t) => {
  const user = await User.findOne({ where: { id: 1 }, lock: true });
  // Do some slow computation (10 seconds)
  await user.update({ balance: user.balance - 100 });
});
```

**Problem**: The lock is held for 10 seconds (during the computation).

**Fix**: Keep transactions short. Move computation outside the transaction:
```javascript
const user = await User.findOne({ where: { id: 1 } });
const newBalance = await computeNewBalance(user);  // Outside transaction
await sequelize.transaction(async (t) => {
  await User.update({ balance: newBalance }, { where: { id: 1 } });
});
```

---

## Lock Timeout

Prevent transactions from waiting forever:

```sql
SET lock_timeout = '5s';
BEGIN;
UPDATE users SET email = 'alice@new.com' WHERE id = 1;
-- If the lock can't be acquired within 5 seconds, the transaction is aborted
COMMIT;
```

**Global setting**:
```sql
ALTER DATABASE mydb SET lock_timeout = '10s';
```

**Use case**: Detect deadlocks and lock contention early.

---

## MySQL vs PostgreSQL Locking Differences

| Feature                     | PostgreSQL                | MySQL InnoDB             |
|-----------------------------|---------------------------|--------------------------|
| Row locking                 | Yes                       | Yes                      |
| Table locking               | Yes                       | Yes                      |
| MVCC                        | Yes                       | Yes                      |
| Reads block writes          | No (MVCC)                 | No (MVCC)                |
| Writes block writes         | Yes                       | Yes                      |
| Deadlock detection          | Yes (1 second)            | Yes (immediate)          |
| Gap locks                   | No                        | Yes (for REPEATABLE READ) |
| Explicit `FOR UPDATE`       | Yes                       | Yes                      |
| DDL locks entire table      | Yes (most operations)     | Yes (most operations)    |

**Key difference**: MySQL uses **gap locks** in `REPEATABLE READ` to prevent phantom reads. PostgreSQL uses snapshot isolation instead.

---

## Practical Takeaways

### 1. **Use SELECT FOR UPDATE to Prevent Lost Updates**

```sql
BEGIN;
SELECT balance FROM accounts WHERE id = 1 FOR UPDATE;
UPDATE accounts SET balance = balance - 50 WHERE id = 1;
COMMIT;
```

### 2. **Avoid Long Transactions**

- Keep transactions short (< 100ms if possible)
- Don't do I/O (network calls, file reads) inside transactions

### 3. **Lock Rows in Consistent Order**

Prevents deadlocks:
```sql
-- Always lock id=1 before id=2
UPDATE accounts SET balance = ... WHERE id IN (1, 2) ORDER BY id;
```

### 4. **Batch Updates to Reduce Lock Contention**

Instead of 1000 separate UPDATEs, use one:
```sql
UPDATE users SET last_login = NOW() WHERE id = ANY(ARRAY[1,2,...,1000]);
```

### 5. **Monitor Lock Waits**

```sql
SELECT * FROM pg_locks WHERE NOT granted;
```

If you see locks waiting for >1 second, investigate.

### 6. **Use Lock Timeouts**

Don't let transactions wait forever:
```sql
SET lock_timeout = '5s';
```

### 7. **Avoid DDL on Hot Tables During Peak Hours**

`ALTER TABLE` locks the entire table. Run during low-traffic periods.

---

## Summary

- Locks ensure data integrity during concurrent access
- Row-level locks are acquired by `UPDATE`, `DELETE`, `SELECT FOR UPDATE`
- Table-level locks are acquired by DDL (`ALTER TABLE`, `TRUNCATE`)
- Deadlocks occur when transactions wait for each other
- To avoid deadlocks: lock in consistent order, keep transactions short
- Lock contention on hot rows limits throughput
- Use `SELECT FOR UPDATE` to prevent lost updates
- Monitor locks with `pg_locks` and `pg_stat_activity`
- Keep transactions short to minimize lock hold time

Understanding locks helps you:
- Avoid deadlocks
- Prevent lost updates
- Debug blocking issues in production
- Design schemas that minimize contention
