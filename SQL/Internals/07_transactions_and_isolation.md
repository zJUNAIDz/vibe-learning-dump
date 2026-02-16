# Transactions and Isolation

## Introduction

Transactions are the promise that your database operations are **atomic**, **consistent**, **isolated**, and **durable** (ACID).

But what does "isolated" really mean?

```sql
-- Transaction A:
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
COMMIT;

-- Transaction B (concurrent):
BEGIN;
SELECT balance FROM accounts WHERE id = 1;
COMMIT;
```

**Question**: Does Transaction B see the updated balance or the old balance?

**Answer**: *It depends on the isolation level.*

This chapter explains how transactions work, what isolation levels mean in practice, and how they affect application behavior.

---

## What Is a Transaction?

A transaction is a **unit of work** that either:
- Completes fully (all changes are committed), or
- Fails entirely (all changes are rolled back)

**No in-between state.**

### Example: Bank Transfer

```sql
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;  -- Debit
UPDATE accounts SET balance = balance + 100 WHERE id = 2;  -- Credit
COMMIT;
```

**Guarantee**: Either both updates happen or neither happens.

**Without transactions**:
- If the system crashes after the debit but before the credit, you've lost $100.

---

## ACID Properties

### Atomicity

All operations in a transaction succeed or all fail.

**Example**:
```sql
BEGIN;
INSERT INTO orders (user_id, total) VALUES (1, 100);
INSERT INTO order_items (order_id, product_id, qty) VALUES (1, 42, 2);
COMMIT;
```

If the second `INSERT` fails (e.g., foreign key violation), the first `INSERT` is rolled back.

**Implementation**: Write-Ahead Log (WAL). More on this in the Failures and Recovery chapter.

### Consistency

The database transitions from one valid state to another.

**Example**:
```sql
-- Constraint: balance >= 0
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
-- If this violates balance >= 0, the transaction is aborted
COMMIT;
```

The database enforces constraints (foreign keys, check constraints, unique indexes).

### Isolation

Concurrent transactions don't interfere with each other (to varying degrees, depending on isolation level).

**Example**:
```sql
-- Transaction A:
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
-- (Not committed yet)

-- Transaction B:
BEGIN;
SELECT balance FROM accounts WHERE id = 1;
-- Does this see the updated balance?
```

**Answer**: Depends on isolation level (explained below).

### Durability

Once committed, changes survive crashes.

**Example**:
```sql
BEGIN;
INSERT INTO logs (message) VALUES ('Critical event');
COMMIT;
-- Even if the database crashes immediately after COMMIT, the row will exist after recovery
```

**Implementation**: Write-Ahead Log (WAL) + fsync. Covered in the Failures chapter.

---

## Isolation Levels (The Standard)

The SQL standard defines four isolation levels:

| Isolation Level      | Dirty Read | Non-Repeatable Read | Phantom Read | Serialization Anomalies |
|----------------------|------------|---------------------|--------------|-------------------------|
| Read Uncommitted     | Possible   | Possible            | Possible     | Possible                |
| Read Committed       | Not possible | Possible          | Possible     | Possible                |
| Repeatable Read      | Not possible | Not possible      | Possible (in theory) | Possible         |
| Serializable         | Not possible | Not possible      | Not possible | Not possible            |

**Higher isolation = fewer anomalies = slower performance (more locking/retries).**

---

## Dirty Reads

**Dirty read**: Reading uncommitted changes from another transaction.

### Example

```sql
-- Transaction A:
BEGIN;
UPDATE accounts SET balance = 1000 WHERE id = 1;
-- (Not committed yet)

-- Transaction B (Read Uncommitted):
BEGIN TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
SELECT balance FROM accounts WHERE id = 1;
-- Result: 1000 (dirty read!)
COMMIT;

-- Transaction A (decides to roll back):
ROLLBACK;
-- The balance is still 100, but Transaction B saw 1000
```

**Problem**: Transaction B saw data that **never existed** in any committed state.

**PostgreSQL note**: PostgreSQL doesn't support `READ UNCOMMITTED`. It's silently upgraded to `READ COMMITTED`.

**MySQL**: `READ UNCOMMITTED` is supported (but rarely used in production).

---

## Non-Repeatable Reads

**Non-repeatable read**: Reading the same row twice within a transaction and getting different values.

### Example

```sql
-- Transaction A (Read Committed):
BEGIN;
SELECT balance FROM accounts WHERE id = 1;
-- Result: 100

-- Transaction B:
BEGIN;
UPDATE accounts SET balance = 200 WHERE id = 1;
COMMIT;

-- Back to Transaction A:
SELECT balance FROM accounts WHERE id = 1;
-- Result: 200 (different from the first read!)
COMMIT;
```

**Problem**: The row changed during the transaction. This can cause application logic bugs.

**Use case where this is acceptable**: Displaying data that doesn't need to be consistent (e.g., dashboards).

**Use case where this is problematic**: Financial calculations, reports.

---

## Phantom Reads

**Phantom read**: A query returns different sets of rows on successive reads within a transaction.

### Example

```sql
-- Transaction A (Repeatable Read):
BEGIN TRANSACTION ISOLATION LEVEL REPEATABLE READ;
SELECT COUNT(*) FROM orders WHERE user_id = 1;
-- Result: 10

-- Transaction B:
BEGIN;
INSERT INTO orders (user_id, total) VALUES (1, 50);
COMMIT;

-- Back to Transaction A:
SELECT COUNT(*) FROM orders WHERE user_id = 1;
-- Result: 11 (phantom row!)
COMMIT;
```

**Problem**: New rows appeared during the transaction.

**PostgreSQL note**: PostgreSQL's `REPEATABLE READ` prevents phantom reads (it's actually snapshot isolation). MySQL's `REPEATABLE READ` also prevents phantom reads in most cases.

---

## Read Committed (PostgreSQL Default)

**Guarantee**: You only see committed data.

**But**: Each statement gets a **new snapshot**. Different statements in the same transaction can see different data.

### Example

```sql
-- Transaction A:
BEGIN;  -- Read Committed (default)
SELECT balance FROM accounts WHERE id = 1;
-- Result: 100

-- Transaction B:
UPDATE accounts SET balance = 200 WHERE id = 1;
COMMIT;

-- Back to Transaction A:
SELECT balance FROM accounts WHERE id = 1;
-- Result: 200 (sees the new committed value)
COMMIT;
```

**When to use**:
- Simple queries
- Applications that don't need consistent snapshots
- High-traffic web applications (default for a reason!)

**When NOT to use**:
- Financial calculations across multiple queries
- Reports that need consistency

---

## Repeatable Read (Snapshot Isolation)

**Guarantee**: All queries in the transaction see the **same snapshot** of the database.

**PostgreSQL**: Implements **snapshot isolation**, which is stronger than the SQL standard's `REPEATABLE READ`.

### Example

```sql
-- Transaction A:
BEGIN TRANSACTION ISOLATION LEVEL REPEATABLE READ;
SELECT balance FROM accounts WHERE id = 1;
-- Result: 100

-- Transaction B:
UPDATE accounts SET balance = 200 WHERE id = 1;
COMMIT;

-- Back to Transaction A:
SELECT balance FROM accounts WHERE id = 1;
-- Result: 100 (still sees the old snapshot)
COMMIT;
```

**When to use**:
- Reports
- Multi-step business logic where consistency matters

**Trade-off**: Higher chance of serialization failures (see below).

---

## Serializable (Strictest Isolation)

**Guarantee**: Transactions execute as if they ran **serially** (one after another), even though they run concurrently.

**No anomalies possible.**

### How It Works

PostgreSQL uses **Serializable Snapshot Isolation (SSI)**:
- Runs transactions concurrently (like `REPEATABLE READ`)
- Detects serialization conflicts
- Aborts transactions if a conflict is detected

### Example: Write Skew

```sql
-- Constraint: The sum of accounts 1 and 2 must be >= 100

-- Transaction A:
BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;
SELECT balance FROM accounts WHERE id IN (1, 2);
-- Result: account 1 = 80, account 2 = 30 (total = 110)
UPDATE accounts SET balance = balance - 20 WHERE id = 1;
-- (total would be 90, but Transaction A doesn't know about Transaction B yet)

-- Transaction B (concurrent):
BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;
SELECT balance FROM accounts WHERE id IN (1, 2);
-- Result: account 1 = 80, account 2 = 30 (total = 110)
UPDATE accounts SET balance = balance - 20 WHERE id = 2;
-- (total would be 90)

-- Both transactions commit:
COMMIT;  -- Transaction A
COMMIT;  -- Transaction B: ERROR! Serialization failure
```

**PostgreSQL detects the conflict** and aborts one transaction.

**Your application must retry**:
```javascript
while (true) {
  try {
    await db.query('BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE');
    // ... your queries
    await db.query('COMMIT');
    break;
  } catch (err) {
    if (err.code === '40001') {  // Serialization failure
      continue;  // Retry
    }
    throw err;
  }
}
```

**When to use**:
- Complex business logic with subtle race conditions
- Financial systems where correctness is critical

**Trade-off**: Higher abort rate = more retries = lower throughput.

---

## Isolation Levels in PostgreSQL vs MySQL

### PostgreSQL

| Isolation Level      | Snapshot Behavior               | Prevents Phantom Reads? | Retries Needed? |
|----------------------|---------------------------------|-------------------------|-----------------|
| Read Committed       | New snapshot per statement      | No                      | No              |
| Repeatable Read      | One snapshot per transaction    | Yes                     | Sometimes       |
| Serializable         | One snapshot + conflict detection | Yes                   | Often           |

### MySQL InnoDB

| Isolation Level      | Snapshot Behavior               | Prevents Phantom Reads? | Uses Locks?     |
|----------------------|---------------------------------|-------------------------|-----------------|
| Read Committed       | New snapshot per statement      | No                      | Some            |
| Repeatable Read      | One snapshot per transaction    | Yes (mostly)            | Some            |
| Serializable         | Same as Repeatable Read + locks | Yes                     | Heavy locking   |

**Key difference**: MySQL's `SERIALIZABLE` uses **locking** (not SSI). This can cause deadlocks and blocking.

---

## When Isolation Breaks Expectations

### Lost Update

```sql
-- Transaction A:
BEGIN;
SELECT balance FROM accounts WHERE id = 1;  -- balance = 100
-- (app calculates new balance: 100 - 50 = 50)
UPDATE accounts SET balance = 50 WHERE id = 1;
COMMIT;

-- Transaction B (concurrent):
BEGIN;
SELECT balance FROM accounts WHERE id = 1;  -- balance = 100 (stale!)
-- (app calculates new balance: 100 - 30 = 70)
UPDATE accounts SET balance = 70 WHERE id = 1;  -- Overwrites Transaction A's change
COMMIT;
```

**Problem**: Transaction B's update **overwrites** Transaction A's update. The account should have balance = 20, but it's 70.

**Fix**: Use `SELECT FOR UPDATE` (next chapter) or `UPDATE ... WHERE balance = 100` (optimistic locking).

### Write Skew

(Already covered in Serializable section.)

**Example**: Two transactions each check a condition, both see it's satisfied, both update, and together they violate the condition.

**Only prevented by**: `SERIALIZABLE` isolation.

---

## Practical Patterns

### Pattern 1: Read Committed + Optimistic Locking

**Use case**: High-traffic web apps.

```sql
-- Read the current version
SELECT id, balance, version FROM accounts WHERE id = 1;
-- Result: id=1, balance=100, version=5

-- Update with version check
UPDATE accounts 
SET balance = balance - 50, version = version + 1
WHERE id = 1 AND version = 5;

-- If no rows updated, another transaction modified the row → retry
```

**Pros**: No locking, high concurrency.
**Cons**: Requires app-level retry logic.

### Pattern 2: Repeatable Read for Reports

**Use case**: Generating consistent reports.

```sql
BEGIN TRANSACTION ISOLATION LEVEL REPEATABLE READ;
-- Run multiple queries
SELECT SUM(balance) FROM accounts;  -- Total balance
SELECT COUNT(*) FROM accounts WHERE balance < 0;  -- Overdrawn accounts
-- All queries see the same snapshot
COMMIT;
```

**Pros**: Consistent data.
**Cons**: Can block VACUUM (if the transaction runs too long).

### Pattern 3: Serializable for Critical Operations

**Use case**: Financial ledgers, inventory management.

```sql
BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;
-- Complex logic with multiple reads and writes
-- If conflict, PostgreSQL aborts the transaction
COMMIT;
```

**Pros**: Guaranteed correctness.
**Cons**: Must handle retries.

---

## Performance vs Correctness Tradeoffs

| Isolation Level      | Performance | Correctness | Complexity |
|----------------------|-------------|-------------|------------|
| Read Committed       | High        | Medium      | Low        |
| Repeatable Read      | Medium      | High        | Medium     |
| Serializable         | Low         | Highest     | High       |

**General rule**: Use the weakest isolation level that satisfies your correctness requirements.

---

## Transaction Lifecycle

### Starting a Transaction

```sql
BEGIN;  -- or START TRANSACTION
```

PostgreSQL assigns:
- **Transaction ID** (xid)
- **Snapshot** (for visibility checks)

### Committing

```sql
COMMIT;
```

- Changes become visible to other transactions
- WAL is flushed to disk (durability)
- Locks are released

### Rolling Back

```sql
ROLLBACK;
```

- All changes are discarded
- Locks are released

---

## Savepoints (Partial Rollback)

You can rollback to a specific point within a transaction:

```sql
BEGIN;
INSERT INTO logs (message) VALUES ('Step 1');
SAVEPOINT step1;
INSERT INTO logs (message) VALUES ('Step 2');
-- Oops, step 2 failed
ROLLBACK TO SAVEPOINT step1;
-- Step 1 is still committed; step 2 is rolled back
INSERT INTO logs (message) VALUES ('Step 2 (fixed)');
COMMIT;
```

**Use case**: Complex multi-step operations where you want to retry individual steps without discarding the entire transaction.

---

## Transaction Timeout

Long-running transactions are dangerous (they block VACUUM, hold locks, etc.).

**Set a timeout**:
```sql
SET statement_timeout = '5s';
BEGIN;
-- Long query
-- If it takes > 5 seconds, PostgreSQL aborts the transaction
COMMIT;
```

**Or set globally**:
```sql
ALTER DATABASE mydb SET statement_timeout = '30s';
```

---

## Monitoring Active Transactions

```sql
SELECT 
  pid, 
  now() - xact_start AS duration, 
  state, 
  query
FROM pg_stat_activity
WHERE state = 'active' OR state = 'idle in transaction'
ORDER BY duration DESC;
```

**Look for**:
- `state = 'idle in transaction'` → Transaction is open but doing nothing (bad!)
- Long `duration` → Blocking VACUUM

**Kill a transaction**:
```sql
SELECT pg_terminate_backend(pid);
```

---

## Practical Takeaways

### 1. **Use the Right Isolation Level**

- **Read Committed**: Default. Good for most web apps.
- **Repeatable Read**: For reports or multi-step logic.
- **Serializable**: For critical financial logic.

### 2. **Avoid Long-Running Transactions**

- Set `statement_timeout`
- Split large operations into smaller transactions

### 3. **Handle Serialization Failures**

If using `SERIALIZABLE`, retry on error code `40001`.

### 4. **Use SELECT FOR UPDATE for Critical Reads**

Prevents lost updates:
```sql
BEGIN;
SELECT balance FROM accounts WHERE id = 1 FOR UPDATE;
UPDATE accounts SET balance = balance - 50 WHERE id = 1;
COMMIT;
```

### 5. **Monitor Idle Transactions**

Idle transactions hold locks and block VACUUM. Kill them aggressively.

---

## Summary

- Transactions provide atomicity, consistency, isolation, and durability (ACID)
- Isolation levels control how transactions see concurrent changes
- **Read Committed**: Each statement gets a new snapshot (default)
- **Repeatable Read**: One snapshot for the entire transaction
- **Serializable**: Guarantees serial execution (may abort transactions)
- Higher isolation = fewer anomalies but slower performance
- PostgreSQL uses MVCC + SSI for serializable isolation
- MySQL uses locking for serializable isolation
- Application code must handle serialization failures (retries)

Understanding isolation levels helps you choose the right tradeoff between performance and correctness.
