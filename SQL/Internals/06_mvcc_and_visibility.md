# MVCC and Visibility

## Introduction

You've probably seen this behavior:

```sql
-- Terminal 1:
BEGIN;
UPDATE users SET balance = balance - 100 WHERE id = 1;
-- (Don't commit yet)

-- Terminal 2:
SELECT balance FROM users WHERE id = 1;
-- Returns the OLD value (before the update)
```

**How did Terminal 2 see the old value if Terminal 1 already updated the row?**

The answer: **MVCC** (Multi-Version Concurrency Control).

MVCC is PostgreSQL's core concurrency mechanism. It's why reads don't block writes, writes don't block reads, and multiple transactions can access the same data without locking.

But MVCC comes with tradeoffs: table bloat, VACUUM overhead, and surprising visibility rules.

This chapter demystifies MVCC from a practical, application-developer perspective.

---

## What Is MVCC?

**Multi-Version Concurrency Control** means:
- The database keeps **multiple versions** of the same row
- Each transaction sees a **consistent snapshot** of the data
- Reads don't block writes; writes don't block reads

Contrast with **locking-based concurrency** (e.g., MySQL MyISAM, old Oracle):
- Reads acquire read locks
- Writes acquire write locks
- Reads and writes block each other

### Why MVCC?

**Concurrency**: Thousands of readers and writers can operate simultaneously without blocking.

**Consistency**: Each transaction sees a consistent snapshot (no "torn reads" where you see partial updates).

**Simplicity**: Application code doesn't need to manage locks explicitly.

---

## How MVCC Works in PostgreSQL

### Tuple Versions

When you update a row, PostgreSQL doesn't modify it in place. Instead:
1. The old tuple is marked as "dead"
2. A new tuple is inserted

```sql
-- Initial state:
Row 1: (id=1, name='Alice', balance=100)

-- UPDATE:
UPDATE users SET balance = 200 WHERE id = 1;

-- After update:
Row 1 (old): (id=1, name='Alice', balance=100)  ← Dead
Row 2 (new): (id=1, name='Alice', balance=200)  ← Alive
```

Both tuples exist physically on disk.

**How does the database know which version to show?**

### Transaction IDs (xmin and xmax)

Every tuple has two transaction IDs:
- **xmin**: The transaction that created this tuple
- **xmax**: The transaction that deleted this tuple (0 if still alive)

```
┌─────────────────────────────────────┐
│ Tuple 1                             │
│   xmin = 100  (created by txn 100)  │
│   xmax = 150  (deleted by txn 150)  │
│   data: (id=1, name='Alice', ...)   │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ Tuple 2                             │
│   xmin = 150  (created by txn 150)  │
│   xmax = 0    (still alive)         │
│   data: (id=1, name='Alice', ...)   │
└─────────────────────────────────────┘
```

### Visibility Rules

When your transaction (txn 200) queries a row, it checks each tuple:

**Is this tuple visible to me?**
1. Was it created (`xmin`) by a committed transaction before my snapshot?
2. Was it deleted (`xmax`) by a transaction after my snapshot or not at all?

**If yes to both → visible. Otherwise → skip.**

### Example

```sql
-- Transaction 100:
BEGIN;  -- txn_id = 100
INSERT INTO users (id, name, balance) VALUES (1, 'Alice', 100);
COMMIT;

-- Transaction 150:
BEGIN;  -- txn_id = 150
UPDATE users SET balance = 200 WHERE id = 1;
-- (Don't commit yet)

-- Transaction 200 (concurrent):
BEGIN;  -- txn_id = 200, snapshot = 149 (before txn 150 committed)
SELECT balance FROM users WHERE id = 1;
-- Result: 100 (sees Tuple 1, not Tuple 2, because txn 150 hasn't committed)
```

**Visibility check for Tuple 1**:
- `xmin = 100` → created by committed transaction (✓)
- `xmax = 150` → deleted by uncommitted transaction (✓ still visible)

**Visibility check for Tuple 2**:
- `xmin = 150` → created by uncommitted transaction (✗ not visible)

**Result**: Transaction 200 sees the old value (100).

---

## Why Reads Don't Block Writes in PostgreSQL

```sql
-- Terminal 1:
BEGIN;
SELECT * FROM users WHERE id = 1;  -- Reads old tuple
-- (Keep transaction open)

-- Terminal 2:
UPDATE users SET balance = 200 WHERE id = 1;  -- Inserts new tuple
COMMIT;
```

**Terminal 2's update doesn't block because**:
- It doesn't modify the old tuple (Terminal 1 is reading)
- It inserts a new tuple (which Terminal 1 won't see until it starts a new transaction)

**This is the magic of MVCC**: Readers and writers don't interfere.

---

## Why UPDATE Creates New Rows (and Why It Matters)

Every `UPDATE` in PostgreSQL creates a new tuple.

### Example

```sql
UPDATE users SET last_login = NOW() WHERE id = 1;
-- Run this 1 million times
```

**Result**: 1 million dead tuples.

### The Problem: Table Bloat

Dead tuples still occupy space on disk.

**Query performance degrades** because:
- Sequential scans must read dead tuples (and skip them using visibility rules)
- Indexes point to both old and new tuples (index bloat)

**Disk space grows** even though the logical data size is constant.

### Example: Measuring Bloat

```sql
-- Insert 1 million rows
INSERT INTO users (name, balance) 
SELECT 'User' || i, 100 FROM generate_series(1, 1000000) i;

-- Table size: ~100 MB
SELECT pg_size_pretty(pg_total_relation_size('users'));

-- Update all rows
UPDATE users SET balance = balance + 1;

-- Table size: ~200 MB (doubled!)
SELECT pg_size_pretty(pg_total_relation_size('users'));
```

**Solution**: `VACUUM` (explained later).

---

## Snapshots and Isolation

Each transaction operates on a **snapshot** of the database.

### What Is a Snapshot?

A snapshot is a consistent view of the database at a specific point in time.

When you run:
```sql
BEGIN;
```

PostgreSQL assigns your transaction a **snapshot** based on the current transaction ID.

**Key property**: Your snapshot doesn't change during the transaction. You always see the same data, even if other transactions commit changes.

### Example

```sql
-- Transaction A:
BEGIN;  -- Snapshot at txn 100
SELECT balance FROM users WHERE id = 1;  -- Result: 100

-- (Meanwhile, Transaction B updates and commits)
-- Transaction B:
BEGIN;
UPDATE users SET balance = 200 WHERE id = 1;
COMMIT;

-- Back to Transaction A:
SELECT balance FROM users WHERE id = 1;  -- Still: 100 (same snapshot!)
COMMIT;

-- New transaction:
BEGIN;
SELECT balance FROM users WHERE id = 1;  -- Result: 200 (new snapshot!)
COMMIT;
```

**Why this matters**: Your application sees a **consistent view** of the database during a transaction, even if the database is being modified concurrently.

---

## Transaction Isolation Levels and MVCC

PostgreSQL supports multiple isolation levels (more in the next chapter), but MVCC is the foundation.

### Read Committed (Default)

Each **statement** gets a new snapshot.

```sql
BEGIN;
SELECT balance FROM users WHERE id = 1;  -- Snapshot 1: balance = 100

-- (Another transaction updates and commits: balance = 200)

SELECT balance FROM users WHERE id = 1;  -- Snapshot 2: balance = 200
COMMIT;
```

**Result**: You see the updated value in the second `SELECT`.

### Repeatable Read

The transaction gets **one snapshot** for all statements.

```sql
BEGIN TRANSACTION ISOLATION LEVEL REPEATABLE READ;
SELECT balance FROM users WHERE id = 1;  -- Snapshot 1: balance = 100

-- (Another transaction updates and commits: balance = 200)

SELECT balance FROM users WHERE id = 1;  -- Still Snapshot 1: balance = 100
COMMIT;
```

**Result**: You see the same value throughout the transaction.

---

## Contrast: MySQL InnoDB

MySQL InnoDB uses MVCC too, but with key differences.

### InnoDB's Undo Log

Instead of storing multiple tuple versions in the table, InnoDB:
- Stores the **latest version** in the clustered index
- Stores **old versions** in the **undo log** (a separate structure)

```
Clustered Index:
  id=1 → (name='Alice', balance=200)

Undo Log:
  rollback segment: balance was 100 (txn 100 → txn 150)
```

When an old transaction reads the row:
1. Read the latest version from the clustered index (200)
2. Check undo log: "Was this version created after my snapshot?"
3. If yes, reconstruct the old version (100) from undo log

**Pros**:
- No table bloat (old versions are in a separate log)
- No VACUUM needed

**Cons**:
- Undo log can grow large if long-running transactions exist
- Older versions require reconstruction (slightly slower reads)

---

## Why Visibility Checks Are Expensive

Every time PostgreSQL reads a tuple, it must check:
- Is `xmin` committed?
- Is `xmax` committed?
- Is this tuple visible to my snapshot?

**For a table with 1 million rows and 50% dead tuples**:
- A sequential scan must read 1 million tuples
- Check visibility for 1 million tuples
- Skip 500,000 dead tuples

**Cost**: High CPU usage for visibility checks.

### The Visibility Map

PostgreSQL optimizes this with a **visibility map**:
- A bitmap tracking which pages have "all tuples visible" (no dead tuples)
- If all tuples on a page are visible, skip visibility checks

**Result**: Sequential scans are faster (especially after `VACUUM`).

---

## How DELETE Works (It Doesn't Actually Delete)

```sql
DELETE FROM users WHERE id = 1;
```

**What PostgreSQL does**:
- Sets `xmax` to the current transaction ID (marks tuple as deleted)
- The tuple is still on disk (dead tuple)

**What PostgreSQL does NOT do**:
- Free disk space
- Remove the tuple physically

**Why?** Other transactions might still need to see the old tuple (due to MVCC snapshots).

**When is space reclaimed?** After `VACUUM` runs and confirms no transaction needs the dead tuple.

---

## HOT Updates (Heap-Only Tuples)

Normally, an `UPDATE` creates a new tuple, and all indexes must be updated to point to the new tuple.

**This is expensive** if you have many indexes.

### Optimization: HOT

If the updated columns are **not indexed**, PostgreSQL uses **HOT (Heap-Only Tuples)**:
- The new tuple is stored in the same page (or a nearby page)
- Indexes are **not updated** (they still point to the old tuple)
- A pointer links the old tuple to the new tuple

```
Index:
  name='Alice' → TID(page=42, offset=1)

Heap (Page 42):
  Offset 1 (old): (name='Alice', balance=100) → points to Offset 2
  Offset 2 (new): (name='Alice', balance=200)
```

**Result**: Updates are faster (no index updates).

**Requirement**: The new tuple must fit on the same page (or have free space nearby).

### When HOT Fails

If:
- The updated column is indexed, or
- The page is full (no free space)

Then HOT is disabled, and all indexes must be updated.

**Implication**: Avoid indexing columns that change frequently (e.g., `last_login`, `updated_at`).

---

## Long-Running Transactions Are Evil

Imagine:
```sql
-- Transaction A:
BEGIN;
SELECT * FROM users WHERE id = 1;
-- (Leave transaction open for 1 hour)
```

**Problem**: Transaction A's snapshot is 1 hour old. All tuples created after that snapshot cannot be vacuumed (because Transaction A might still need to see them).

**Result**:
- Dead tuples accumulate
- Table bloat increases
- Queries slow down

**Monitoring**:
```sql
SELECT 
  pid, 
  now() - xact_start AS duration, 
  state, 
  query
FROM pg_stat_activity
WHERE now() - xact_start > interval '5 minutes'
ORDER BY duration DESC;
```

**Fix:** Kill long-running idle transactions:
```sql
SELECT pg_terminate_backend(pid);
```

---

## Practical Example: Race Condition Without Proper Isolation

```sql
-- Transaction A:
BEGIN;
SELECT balance FROM accounts WHERE id = 1;  -- balance = 100
-- (Calculate new balance in app: 100 - 50 = 50)
UPDATE accounts SET balance = 50 WHERE id = 1;
COMMIT;

-- Transaction B (concurrent):
BEGIN;
SELECT balance FROM accounts WHERE id = 1;  -- balance = 100 (old snapshot!)
-- (Calculate new balance in app: 100 - 30 = 70)
UPDATE accounts SET balance = 70 WHERE id = 1;  -- Overwrites Transaction A's update!
COMMIT;
```

**Result**: Lost update. The account should have balance = 20, but it's 70.

**Fix**: Use `SELECT FOR UPDATE` (explained in the next chapter):
```sql
BEGIN;
SELECT balance FROM accounts WHERE id = 1 FOR UPDATE;  -- Acquires row lock
UPDATE accounts SET balance = balance - 50 WHERE id = 1;
COMMIT;
```

Or use `REPEATABLE READ` isolation with serialization checks.

---

## Visibility and Index-Only Scans

Index-only scans are faster because they don't fetch heap tuples.

**But**: The index doesn't store visibility information (`xmin`, `xmax`).

**Problem**: PostgreSQL must check the heap to determine visibility.

**Solution**: The **visibility map**.
- If the visibility map marks a page as "all visible," PostgreSQL can skip the heap check
- Index-only scans become truly index-only

**When the visibility map is updated**: After `VACUUM` confirms all tuples are visible.

**Implication**: Run `VACUUM` regularly to enable fast index-only scans.

---

## PostgreSQL vs MySQL: MVCC Comparison

| Feature                     | PostgreSQL                | MySQL InnoDB             |
|-----------------------------|---------------------------|--------------------------|
| Version storage             | Multiple tuples in heap   | Undo log                 |
| UPDATE overhead             | Creates new tuple         | Modifies in place + undo |
| Table bloat                 | Yes (needs VACUUM)        | No                       |
| Visibility checks           | Per tuple (expensive)     | Via undo log             |
| Index-only scan limitation  | Needs visibility map      | Works immediately        |
| Long transaction impact     | Prevents VACUUM           | Grows undo log           |
| DELETE behavior             | Marks tuple dead          | Marks row deleted + undo |

**Takeaway**: PostgreSQL's MVCC is simpler but requires VACUUM maintenance. InnoDB avoids bloat but has undo log overhead.

---

## Practical Takeaways

### 1. **Avoid Long-Running Transactions**

Idle or slow transactions block VACUUM and cause bloat.

**Monitor**:
```sql
SELECT pid, state, now() - xact_start AS duration FROM pg_stat_activity;
```

**Fix**: Set a statement timeout or application-level timeout.

### 2. **Run VACUUM Regularly**

Autovacuum should handle this, but tune it if:
- Tables are heavily updated
- Tables are growing unexpectedly

### 3. **Avoid Indexing Frequently-Updated Columns**

Columns like `last_login` or `updated_at` that change on every update:
- Prevent HOT updates
- Cause index bloat

**Alternative**: Store them in a separate table.

### 4. **Understand Snapshot Isolation**

Under `REPEATABLE READ`, your transaction sees a frozen snapshot. Other transactions' changes are invisible.

**Use case**: Generating reports (you want consistent data).

**Anti-pattern**: Long-running analytics on the primary DB (blocks VACUUM).

### 5. **Monitor Table Bloat**

```sql
SELECT 
  schemaname, 
  tablename, 
  n_dead_tup, 
  n_live_tup, 
  round(n_dead_tup * 100.0 / NULLIF(n_live_tup + n_dead_tup, 0), 2) AS dead_ratio
FROM pg_stat_user_tables
WHERE n_live_tup > 0
ORDER BY dead_ratio DESC;
```

If `dead_ratio` > 20%, your table is bloated. Run `VACUUM`.

---

## Summary

- MVCC allows reads and writes to proceed concurrently without locking
- Each transaction sees a consistent snapshot of the database
- PostgreSQL stores multiple tuple versions in the heap
- UPDATEs create new tuples; DELETEs mark tuples as dead
- Dead tuples cause table bloat and slow queries
- VACUUM reclaims space from dead tuples
- Long-running transactions prevent VACUUM and cause bloat
- HOT updates avoid index maintenance for non-indexed columns
- MySQL InnoDB uses undo logs instead of heap tuple versions

Understanding MVCC helps you:
- Predict transaction behavior
- Avoid bloat and performance degradation
- Debug weird concurrency bugs
