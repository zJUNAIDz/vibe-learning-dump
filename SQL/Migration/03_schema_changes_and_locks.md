# Schema Changes and Locks: Why ALTER TABLE Is Scary

## The Lock Problem

When you run `ALTER TABLE`, PostgreSQL must acquire a lock. This prevents data corruption but also **blocks other queries**.

The fundamental tension:
- **Safety**: Locks prevent corruption
- **Availability**: Locks block queries
- **Production**: Both matter simultaneously

Understanding locks is the difference between safe migrations and production outages.

## Lock Types: PostgreSQL

PostgreSQL has a sophisticated lock hierarchy. For migrations, know these:

### ACCESS SHARE (Least Restrictive)

Acquired by: `SELECT`

Conflicts with: `ACCESS EXCLUSIVE`

```sql
SELECT * FROM users;  -- Acquires ACCESS SHARE
```

Multiple `SELECT` queries can run simultaneously. This is fine.

### ROW EXCLUSIVE

Acquired by: `INSERT`, `UPDATE`, `DELETE`

Conflicts with: `SHARE`, `EXCLUSIVE`, `ACCESS EXCLUSIVE`

```sql
UPDATE users SET last_login = NOW() WHERE id = 123;  -- Acquires ROW EXCLUSIVE
```

Multiple writes can happen simultaneously (on different rows). Still fine.

### ACCESS EXCLUSIVE (Most Restrictive)

Acquired by: `ALTER TABLE`, `DROP TABLE`, `TRUNCATE`, `VACUUM FULL`

Conflicts with: **EVERYTHING**

```sql
ALTER TABLE users ADD COLUMN phone VARCHAR(20);  -- Acquires ACCESS EXCLUSIVE
```

**This blocks**:
- All `SELECT` queries
- All `INSERT/UPDATE/DELETE` queries
- Everything trying to touch the table

**This is why migrations are scary**.

### Visualizing Lock Conflicts

```
Query Type              Lock Acquired          Blocks What?
─────────────────────────────────────────────────────────────
SELECT                  ACCESS SHARE           (only ACCESS EXCLUSIVE)
INSERT/UPDATE/DELETE    ROW EXCLUSIVE          (only table-level changes)
ALTER TABLE             ACCESS EXCLUSIVE       EVERYTHING
```

## The Typical Migration Disaster

Let's walk through a real scenario:

### 10:00:00 - You Start Migration

```sql
ALTER TABLE users ADD COLUMN preferences JSONB;
```

PostgreSQL:
1. Acquires ACCESS EXCLUSIVE lock
2. Begins rewriting the table (adding column)
3. This will take 45 seconds on 5M rows

### 10:00:01 - Application Queries Arrive

```sql
-- Every app request tries to do this
SELECT * FROM users WHERE id = 123;
```

These queries **wait** for the lock to be released.

### 10:00:05 - Connection Pool Exhausts

Your connection pool (say, 20 connections) fills up:
- All connections waiting on the `users` table
- No connections available for new requests
- Application starts timing out

### 10:00:10 - Cascading Failure

- Health checks fail (can't connect to DB)
- Load balancer removes instances
- Traffic shifts to remaining instances
- They overload immediately
- Everything melts down

### 10:00:45 - Migration Completes

Lock releases. But:
- Your app is already down
- Users are gone
- Incident started

**Cause**: 45 seconds of an ACCESS EXCLUSIVE lock during production traffic.

## Why ALTER TABLE Takes Time

### Fast Operations (Metadata Only)

PostgreSQL 11+ can do these instantly:

```sql
-- Instant: just updates metadata
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- Instant: metadata-only
ALTER TABLE users ADD COLUMN preferences JSONB DEFAULT NULL;
```

**Why fast**: No table rewrite needed. Just updates the system catalog.

### Slow Operations (Table Rewrite)

These rewrite the entire table:

```sql
-- Requires table rewrite on older Postgres, or with non-null default
ALTER TABLE users ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active';

-- Always requires rewrite
ALTER TABLE users ALTER COLUMN email TYPE TEXT;

-- Requires rewrite
ALTER TABLE users ADD COLUMN id_new SERIAL PRIMARY KEY;
```

**Why slow**: PostgreSQL must:
1. Create new table structure
2. Copy all rows (transforming data)
3. Rebuild indexes
4. Swap old and new tables

On a 5M row table, this takes real time (seconds to minutes).

## The Lock Duration Problem

The ACCESS EXCLUSIVE lock is held for the **entire operation**.

```sql
-- This takes 45 seconds
-- Lock is held for ALL 45 seconds
ALTER TABLE users ADD COLUMN preferences JSONB NOT NULL DEFAULT '{}';
```

During those 45 seconds, the table is completely inaccessible.

### Lock Queues and Starvation

Worse: locks queue. If there's a long-running query when you try to alter:

```sql
-- Transaction 1 (started at 10:00:00)
BEGIN;
SELECT COUNT(*) FROM users;  -- Takes 2 minutes (big table)
-- Still running...

-- Transaction 2 (you, at 10:00:30)
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
-- Waits for Transaction 1

-- Transaction 3 (user request, at 10:00:31)
SELECT * FROM users WHERE id = 123;
-- Waits for Transaction 2
-- Which waits for Transaction 1

-- Everything is blocked
```

Even a fast `ALTER TABLE` gets stuck behind a slow `SELECT`.

## Adding Columns Safely

### Safe: NULL Column (PostgreSQL 11+)

```sql
-- No table rewrite, no lock held long
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
```

✓ Fast (metadata only)
✓ Short lock duration
✓ Safe in production

### Unsafe: NOT NULL with DEFAULT

```sql
-- Rewrites entire table!
ALTER TABLE users ADD COLUMN phone VARCHAR(20) NOT NULL DEFAULT '';
```

✗ Locks table during entire rewrite
✗ Dangerous on large tables

### Safe Approach: Multi-Step

```sql
-- Step 1: Add as nullable (fast)
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- Step 2: Backfill in batches (no lock)
UPDATE users SET phone = '' WHERE phone IS NULL AND id >= 0 AND id < 10000;
UPDATE users SET phone = '' WHERE phone IS NULL AND id >= 10000 AND id < 20000;
-- ... continue in batches

-- Step 3: Add NOT NULL (fast, since no NULLs exist)
ALTER TABLE users ALTER COLUMN phone SET NOT NULL;
```

✓ Each step is fast
✓ Locks are held briefly
✓ Can monitor progress

## Dropping Columns Safely

### The Problem

```sql
-- Old code is still running
ALTER TABLE users DROP COLUMN username;
-- New queries fail: "column username does not exist"
```

### Safe Approach: Multi-Phase

**Phase 1**: Deploy code that doesn't use the column
```typescript
// Old code
const user = await db.query('SELECT id, email, username FROM users');

// New code (deployed first)
const user = await db.query('SELECT id, email FROM users');
```

**Phase 2**: Wait for old deployments to finish

**Phase 3**: Drop the column
```sql
ALTER TABLE users DROP COLUMN username;
```

### Gotcha: Column Drop Is Fast But...

In PostgreSQL, `DROP COLUMN` is metadata-only (fast), but:

```sql
-- This takes a long lock briefly
ALTER TABLE users DROP COLUMN username;
```

Still acquires ACCESS EXCLUSIVE, so:
- Do it during low traffic
- Prepare for brief unavailability
- Monitor for lock waits

## Renaming Columns and Tables

### Renaming a Column

```sql
-- Instant metadata change, but breaks everything
ALTER TABLE users RENAME COLUMN email TO email_address;
```

**Problem**: Every query now fails:
```sql
SELECT email FROM users;
-- ERROR: column "email" does not exist
```

**Safe Approach**: Don't rename. Add new column, backfill, deprecate old:

```sql
-- 1. Add new column
ALTER TABLE users ADD COLUMN email_address VARCHAR(255);

-- 2. Backfill
UPDATE users SET email_address = email WHERE email_address IS NULL;

-- 3. Deploy code that writes to both (dual writes)
-- 4. Deploy code that reads from email_address
-- 5. Eventually drop email column
```

It's tedious, but safe. Renaming is a breaking change—treat it as such.

### Renaming a Table

```sql
-- Instant, but breaks everything
ALTER TABLE users RENAME TO accounts;
```

Same problem, same solution: Create new table, dual write, migrate, deprecate.

**Or**: Use a database view as an alias:

```sql
-- Keep both names working
CREATE VIEW users AS SELECT * FROM accounts;
```

This gives you breathing room to migrate code gradually.

## PostgreSQL vs MySQL Locking Differences

### PostgreSQL

- **DDL is transactional**: Can rollback failed `ALTER TABLE`
- **Explicit lock control**: `LOCK TABLE users IN ACCESS EXCLUSIVE MODE`
- **Concurrent index creation**: `CREATE INDEX CONCURRENTLY`
- **Fast column add** (v11+): No rewrite for nullable columns

### MySQL

- **DDL is NOT transactional**: Failed `ALTER TABLE` leaves partial state
- **Implicit locking**: Less control
- **No concurrent operations**: Most DDLs block fully
- **InnoDB online DDL**: Some operations support `ALGORITHM=INPLACE` (5.6+)

Example in MySQL:

```sql
-- MySQL 5.6+ can do this without full table lock
ALTER TABLE users 
  ADD COLUMN phone VARCHAR(20),
  ALGORITHM=INPLACE, 
  LOCK=NONE;
```

But many operations still require `ALGORITHM=COPY` (full table rewrite).

**Best practice for MySQL**: Test extensively, use `pt-online-schema-change` for large tables.

## Measuring Lock Duration

Before running a migration in production, measure it:

### Test on Production-Sized Data

```bash
# Restore production backup to staging
pg_restore -d staging_db prod_backup.dump

# Time the migration
\timing on
ALTER TABLE users ADD COLUMN preferences JSONB;
-- Time: 45231.456 ms (45 seconds)
```

Now you know: "This will lock the table for ~45 seconds in production."

### Monitor Active Locks

While migration runs:

```sql
-- See what's blocking
SELECT 
  pid,
  usename,
  pg_blocking_pids(pid) as blocked_by,
  query,
  state,
  wait_event_type,
  wait_event
FROM pg_stat_activity
WHERE wait_event_type IS NOT NULL;
```

```sql
-- See lock modes
SELECT 
  t.relname,
  l.locktype,
  l.mode,
  l.granted,
  a.query
FROM pg_locks l
JOIN pg_stat_activity a ON l.pid = a.pid
JOIN pg_class t ON l.relation = t.oid
WHERE t.relname = 'users';
```

## Strategies to Minimize Lock Time

### 1. Lock Timeout

Prevent indefinite waiting:

```sql
-- Fail if can't acquire lock in 5 seconds
SET lock_timeout = '5s';
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
```

If it fails, your migration doesn't hang forever blocking everything.

### 2. Statement Timeout

Prevent long-running migrations:

```sql
-- Fail if migration takes >1 minute
SET statement_timeout = '60s';
ALTER TABLE users ADD COLUMN preferences JSONB DEFAULT '{}';
```

### 3. Retry Logic

```bash
#!/bin/bash
# Retry migration up to 10 times
for i in {1..10}; do
  psql -c "
    SET lock_timeout = '2s';
    ALTER TABLE users ADD COLUMN phone VARCHAR(20);
  " && break
  
  echo "Failed to acquire lock, retrying in 5s..."
  sleep 5
done
```

This waits for a gap in traffic to acquire the lock.

### 4. Low-Traffic Windows

The boring solution:
- Run migrations at 3 AM
- Or during planned maintenance windows

Not sexy, but effective.

### 5. Break Into Smaller Operations

Instead of:

```sql
-- One big lock
ALTER TABLE users 
  ADD COLUMN phone VARCHAR(20),
  ADD COLUMN verified BOOLEAN,
  ADD COLUMN preferences JSONB;
```

Do:

```sql
-- Three smaller locks
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
ALTER TABLE users ADD COLUMN verified BOOLEAN;
ALTER TABLE users ADD COLUMN preferences JSONB;
```

Each lock is held for less time. Queries can squeeze in between.

## PostgreSQL-Specific: Adding Columns with Defaults

### Before PostgreSQL 11

```sql
-- Rewrites entire table
ALTER TABLE users ADD COLUMN status VARCHAR(20) DEFAULT 'active';
```

### PostgreSQL 11+

```sql
-- Instant! No rewrite
ALTER TABLE users ADD COLUMN status VARCHAR(20) DEFAULT 'active';
```

**Why**: PostgreSQL 11 stores default values in the system catalog. Existing rows logically have the default, but physically it's not stored until row is updated.

**Caveat**: NOT NULL still requires validation:

```sql
-- Instant metadata change, but...
ALTER TABLE users ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active';
-- Still acquires ACCESS EXCLUSIVE briefly
```

## Dealing with Large Tables

### What's "Large"?

In production:
- **Small**: < 100K rows (ALTER TABLE usually OK)
- **Medium**: 100K - 10M rows (test timing, proceed with caution)
- **Large**: 10M+ rows (assume ALTER TABLE is dangerous)

But also consider:
- Table bloat (actual disk size)
- Row width (wide rows are slower)
- Index count (more indexes = longer operations)

### Tools for Large Tables

#### PostgreSQL: pg_repack

Rebuild tables without locking:

```bash
pg_repack -t users -d mydb
```

How it works:
1. Creates shadow table
2. Copies data incrementally
3. Applies ongoing changes via triggers
4. Swaps tables with brief lock

#### MySQL: pt-online-schema-change

Percona Toolkit:

```bash
pt-online-schema-change \
  --alter "ADD COLUMN phone VARCHAR(20)" \
  D=mydb,t=users \
  --execute
```

Similar approach: shadow table + triggers + swap.

### Manual Approach: Shadow Tables

```sql
-- 1. Create new table with desired structure
CREATE TABLE users_new (
  id SERIAL PRIMARY KEY,
  email VARCHAR(255),
  phone VARCHAR(20)  -- New column
);

-- 2. Copy data in batches
INSERT INTO users_new (id, email, phone)
SELECT id, email, NULL 
FROM users
WHERE id >= 0 AND id < 10000;

-- Continue in batches...

-- 3. Set up triggers to keep new table updated
CREATE TRIGGER users_sync AFTER INSERT OR UPDATE OR DELETE ON users
FOR EACH ROW EXECUTE FUNCTION sync_to_users_new();

-- 4. Once caught up, swap tables (brief lock)
BEGIN;
ALTER TABLE users RENAME TO users_old;
ALTER TABLE users_new RENAME TO users;
COMMIT;

-- 5. Clean up old table
DROP TABLE users_old;
```

Tedious, but gives full control.

## Real-World Example: Adding a Column to a Hot Table

**Scenario**: Add `last_active_at` to `users` table with 50M rows.

### Naive Approach (Don't Do This)

```sql
ALTER TABLE users ADD COLUMN last_active_at TIMESTAMP DEFAULT NOW();
```

**Result**: 
- Table locks for 3+ minutes
- Entire application down
- Users can't log in
- Payments fail
- Disaster

### Safe Approach

```sql
-- Step 1: Add column as nullable (instant in PG 11+)
ALTER TABLE users ADD COLUMN last_active_at TIMESTAMP;

-- Step 2: Deploy code that starts writing to the column
-- (New logins update last_active_at)

-- Step 3: Backfill existing rows in batches
DO $$
DECLARE
  batch_size INT := 10000;
  min_id INT;
  max_id INT;
BEGIN
  SELECT MIN(id), MAX(id) INTO min_id, max_id FROM users;
  
  FOR i IN min_id..max_id BY batch_size LOOP
    UPDATE users 
    SET last_active_at = created_at  -- Use sensible default
    WHERE id >= i AND id < i + batch_size
      AND last_active_at IS NULL;
    
    COMMIT;  -- Release lock between batches
    PERFORM pg_sleep(0.1);  -- Breathe
  END LOOP;
END $$;

-- Step 4: Once backfilled, add NOT NULL if needed
ALTER TABLE users ALTER COLUMN last_active_at SET NOT NULL;

-- Step 5: Maybe add an index
CREATE INDEX CONCURRENTLY idx_users_last_active 
  ON users(last_active_at);
```

**Result**:
- No downtime
- Small, manageable locks
- Can monitor progress
- Can pause/resume
- Safe

## Summary: Lock Avoidance Checklist

Before running a migration:

- [ ] Do I know what lock this will acquire?
- [ ] How long will the lock be held? (tested on production-sized data)
- [ ] Can I break this into smaller operations?
- [ ] Can I add the column as nullable first?
- [ ] Am I doing this during low traffic?
- [ ] Do I have a lock timeout set?
- [ ] Am I monitoring for lock waits?
- [ ] What's my abort plan if this takes too long?

**Remember**: The goal isn't to avoid locks entirely (impossible). The goal is to **minimize lock duration** to a level your application can tolerate.

In practice, "tolerable" is usually:
- **< 100ms**: Probably fine
- **100ms - 1s**: Noticeable, acceptable
- **1s - 10s**: Risky, requires planning
- **> 10s**: Dangerous, use advanced techniques

Know your SLA and stay within it.

---

**Next**: [Data Migrations](./04_data_migrations.md) - Transforming data safely in production
