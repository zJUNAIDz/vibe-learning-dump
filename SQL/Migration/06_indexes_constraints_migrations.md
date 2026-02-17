# Indexes and Constraints: The Silent Production Killers

## Why Indexes and Constraints Are Special

Most developers think of migrations as "schema changes." But indexes and constraints are where things get truly dangerous.

**The problem**:
- Creating an index locks the table
- Adding constraints validates all existing data
- Both can take minutes or hours on large tables
- During this time, your application might be down

These are the migrations that cause 3 AM pages.

## Creating Indexes: The Lock Problem

### Standard Index Creation (Dangerous)

```sql
CREATE INDEX idx_users_email ON users(email);
```

**What happens**:
1. PostgreSQL acquires **SHARE lock** on table
2. Reads are OK
3. **Writes are blocked**
4. Builds the index (can take minutes)
5. Releases lock

On a production table with writes happening:
```sql
-- Meanwhile...
INSERT INTO users (email) VALUES ('new@example.com');
-- ⏸️ Blocked, waiting for index creation
-- ⏸️ Still waiting...
-- ⏸️ Still waiting...
-- 💥 Eventually times out
```

**Result**: Production writes fail. Application breaks.

### Concurrent Index Creation (Safe)

```sql
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);
```

**What happens**:
1. Builds index in background
2. Does **NOT** block writes
3. Takes longer than standard CREATE INDEX
4. Safe for production

**This is the only way to add indexes to production tables.**

### Limitations of CONCURRENTLY

1. **Cannot run inside a transaction**

```sql
BEGIN;
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);
-- ERROR: CREATE INDEX CONCURRENTLY cannot run inside a transaction block
```

Your migration script needs to handle this:

```sql
-- Don't wrap in transaction
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON users(email);
```

2. **Can fail and leave invalid index**

If creation fails (crash, timeout, etc.):

```sql
\di+ idx_users_email
-- shows INVALID in status
```

You must drop and retry:

```sql
DROP INDEX CONCURRENTLY idx_users_email;
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);
```

3. **Takes longer**

Because it doesn't lock, it must work around concurrent changes. Budget 2-3x the time of regular CREATE INDEX.

4. **Requires two sequential scans**

More I/O intensive. On very large tables, this matters.

### Best Practices for Index Creation

```sql
-- Always use CONCURRENTLY in production
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON users(email);

-- For composite indexes
CREATE INDEX CONCURRENTLY idx_users_email_status 
  ON users(email, status);

-- For partial indexes (smaller, faster)
CREATE INDEX CONCURRENTLY idx_users_active_email 
  ON users(email) 
  WHERE status = 'active';

-- For expression indexes
CREATE INDEX CONCURRENTLY idx_users_lower_email 
  ON users(LOWER(email));
```

**Migration template**:

```sql
-- Migration: add_index_users_email
-- Note: Uses CONCURRENTLY, cannot be in transaction

-- Check if index already exists (idempotent)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_indexes 
    WHERE tablename = 'users' 
    AND indexname = 'idx_users_email'
  ) THEN
    -- Create index (note: this whole block is NOT in a transaction)
    CREATE INDEX CONCURRENTLY idx_users_email ON users(email);
  END IF;
END $$;
```

Actually, better:

```sql
-- Just use IF NOT EXISTS (simpler)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email 
  ON users(email);
```

## Dropping Indexes Safely

### Standard Drop (Usually Safe)

```sql
DROP INDEX idx_users_email;
```

Dropping an index is usually fast and safe. But it does acquire a lock briefly.

### Concurrent Drop (Safest)

```sql
DROP INDEX CONCURRENTLY idx_users_email;
```

Even safer, no lock held.

**Best practice**: Use CONCURRENTLY for drops too.

```sql
-- Idempotent drop
DROP INDEX CONCURRENTLY IF EXISTS idx_users_email;
```

## Adding NOT NULL Constraints

### The Dangerous Way

```sql
ALTER TABLE users ALTER COLUMN email SET NOT NULL;
```

**What happens**:
1. PostgreSQL scans **entire table** to verify no NULLs exist
2. Holds **ACCESS EXCLUSIVE lock** during scan
3. On a large table, this can take minutes
4. Everything is blocked

**This is a production killer.**

### The Safe Way (PostgreSQL 12+)

**Step 1**: Add constraint as NOT VALID

```sql
-- Fast: doesn't validate existing data
ALTER TABLE users 
  ADD CONSTRAINT users_email_not_null 
  CHECK (email IS NOT NULL) 
  NOT VALID;
```

This adds the constraint but doesn't verify existing rows. New rows must satisfy it, but old rows aren't checked.

**Step 2**: Validate the constraint

```sql
-- Validates without blocking writes
ALTER TABLE users 
  VALIDATE CONSTRAINT users_email_not_null;
```

This scans the table but only takes SHARE UPDATE EXCLUSIVE lock (allows reads and writes).

**Step 3**: Convert to true NOT NULL

```sql
-- Now safe: constraint is validated
ALTER TABLE users ALTER COLUMN email SET NOT NULL;

-- Clean up the check constraint
ALTER TABLE users DROP CONSTRAINT users_email_not_null;
```

**Why this works**:
- Step 1: Fast, minimal lock
- Step 2: Slow, but doesn't block writes (much)
- Step 3: Fast, just metadata change

### Before PostgreSQL 12

Use the CHECK constraint approach but skip step 3:

```sql
-- Add constraint
ALTER TABLE users 
  ADD CONSTRAINT users_email_not_null 
  CHECK (email IS NOT NULL) 
  NOT VALID;

-- Validate
ALTER TABLE users 
  VALIDATE CONSTRAINT users_email_not_null;

-- Leave the CHECK constraint in place (acts like NOT NULL)
```

### Ensuring No NULLs First

Before adding NOT NULL, make sure no NULLs exist:

```sql
-- Check for NULLs
SELECT COUNT(*) FROM users WHERE email IS NULL;

-- If any exist, fix them first
UPDATE users SET email = 'unknown@example.com' WHERE email IS NULL;

-- Or fail explicitly
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM users WHERE email IS NULL) THEN
    RAISE EXCEPTION 'Cannot add NOT NULL: % rows have NULL email', 
      (SELECT COUNT(*) FROM users WHERE email IS NULL);
  END IF;
END $$;
```

## Adding Foreign Keys

### The Dangerous Way

```sql
ALTER TABLE orders 
  ADD CONSTRAINT fk_orders_user_id 
  FOREIGN KEY (user_id) 
  REFERENCES users(id);
```

**What happens**:
1. Validates all existing `user_id` values
2. Scans entire `orders` table
3. Holds locks on both tables
4. Can take a very long time

**This blocks writes to both tables.**

### The Safe Way

**Step 1**: Add constraint as NOT VALID

```sql
ALTER TABLE orders 
  ADD CONSTRAINT fk_orders_user_id 
  FOREIGN KEY (user_id) 
  REFERENCES users(id)
  NOT VALID;
```

Fast. New rows must satisfy constraint, but existing rows aren't checked.

**Step 2**: Validate the constraint

```sql
ALTER TABLE orders 
  VALIDATE CONSTRAINT fk_orders_user_id;
```

Scans table but doesn't block writes (as much).

### Before Adding Foreign Key

Ensure referential integrity:

```sql
-- Find orphaned rows
SELECT COUNT(*) 
FROM orders 
WHERE user_id IS NOT NULL 
  AND user_id NOT IN (SELECT id FROM users);

-- Fix them
DELETE FROM orders 
WHERE user_id NOT IN (SELECT id FROM users);

-- Or update to a placeholder user
UPDATE orders 
SET user_id = (SELECT id FROM users WHERE email = 'deleted@example.com')
WHERE user_id NOT IN (SELECT id FROM users);
```

## Adding CHECK Constraints

### The Pattern

```sql
-- Step 1: Add as NOT VALID
ALTER TABLE products 
  ADD CONSTRAINT check_price_positive 
  CHECK (price > 0) 
  NOT VALID;

-- Step 2: Validate
ALTER TABLE products 
  VALIDATE CONSTRAINT check_price_positive;
```

Same pattern: add fast, validate slowly without blocking writes.

## Adding UNIQUE Constraints

### The Problem

```sql
ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);
```

This:
1. Creates a unique index (without CONCURRENTLY)
2. Blocks writes

### The Solution

Create the unique index separately, then add the constraint:

```sql
-- Step 1: Create unique index concurrently
CREATE UNIQUE INDEX CONCURRENTLY idx_users_email_unique 
  ON users(email);

-- Step 2: Add constraint using existing index
ALTER TABLE users 
  ADD CONSTRAINT users_email_unique 
  UNIQUE USING INDEX idx_users_email_unique;
```

**This is safe**: Index creation doesn't block, constraint addition just uses existing index.

## Dropping Constraints

Usually safe, but acquires brief lock:

```sql
ALTER TABLE users DROP CONSTRAINT users_email_unique;
```

For foreign keys, can use CASCADE:

```sql
-- Drop FK and any dependent objects
ALTER TABLE orders 
  DROP CONSTRAINT fk_orders_user_id CASCADE;
```

## Index Building Progress

For long-running index creation, monitor progress:

```sql
-- PostgreSQL 12+
SELECT 
  phase,
  round(100.0 * blocks_done / nullif(blocks_total, 0), 1) AS "% complete",
  blocks_done,
  blocks_total,
  tuples_done,
  tuples_total
FROM pg_stat_progress_create_index;
```

Example output:
```
phase          | % complete | blocks_done | blocks_total | tuples_done | tuples_total
---------------|------------|-------------|--------------|-------------|--------------
building index |       45.2 |      123456 |       273045 |     5234234 |     12000000
```

## Real-World Example: Adding Index to Large Table

**Scenario**: `events` table has 500M rows. Need to add index on `user_id`.

### Wrong Approach

```sql
CREATE INDEX idx_events_user_id ON events(user_id);
```

**Result**:
- Blocks writes for ~45 minutes
- Production down
- Disaster

### Right Approach

**Step 1**: Test in staging

```bash
# Restore production data to staging
pg_restore -d staging prod_backup.dump

# Time the index creation
\timing on
CREATE INDEX CONCURRENTLY idx_events_user_id ON events(user_id);
-- Time: 2891234.567 ms (48 minutes)
```

Now you know: ~48 minutes in production.

**Step 2**: Schedule the migration

```
Subject: Index creation on events table - Monday 2 AM UTC

We'll be adding an index to the events table to improve query performance.
This will be done concurrently and should not affect application availability.
Expected duration: 45-50 minutes.
```

**Step 3**: Run with monitoring

```bash
# Terminal 1: Run the migration
psql -c "CREATE INDEX CONCURRENTLY idx_events_user_id ON events(user_id);"

# Terminal 2: Monitor progress
watch -n 10 "psql -c \"SELECT * FROM pg_stat_progress_create_index;\""

# Terminal 3: Monitor application
watch -n 5 "curl https://api.example.com/health"

# Terminal 4: Monitor locks
watch -n 5 "psql -c \"SELECT COUNT(*) FROM pg_stat_activity WHERE wait_event_type = 'Lock';\""
```

**Step 4**: Verify

```sql
-- Index created?
\di+ idx_events_user_id

-- Is it being used?
EXPLAIN SELECT * FROM events WHERE user_id = 123;
-- Should show "Index Scan using idx_events_user_id"

-- Performance improvement?
EXPLAIN ANALYZE SELECT * FROM events WHERE user_id = 123;
```

**Step 5**: Monitor application metrics

Watch for:
- Query latency improvements
- Error rate (should not change)
- Throughput improvements

## Partial Indexes: Smaller and Faster

Instead of indexing the entire table:

```sql
-- Indexes all 100M rows
CREATE INDEX CONCURRENTLY idx_events_user_id 
  ON events(user_id);
```

Index only what you query:

```sql
-- Only indexes active events (maybe 10M rows)
CREATE INDEX CONCURRENTLY idx_events_active_user_id 
  ON events(user_id) 
  WHERE status = 'active';
```

**Benefits**:
- Smaller index (faster builds, less disk space)
- Faster queries on active events
- Less maintenance overhead

**Use when**: You mostly query a subset of data.

## Expression Indexes

For queries that transform data:

```sql
-- Query always uses LOWER()
SELECT * FROM users WHERE LOWER(email) = 'alice@example.com';
```

Index the expression:

```sql
CREATE INDEX CONCURRENTLY idx_users_lower_email 
  ON users(LOWER(email));
```

Now the query uses the index.

## Multi-Column Indexes: Order Matters

```sql
-- Index on (status, created_at)
CREATE INDEX CONCURRENTLY idx_events_status_created 
  ON events(status, created_at);

-- This query uses the index
SELECT * FROM events WHERE status = 'active' AND created_at > '2024-01-01';

-- This query uses the index (prefix match)
SELECT * FROM events WHERE status = 'active';

-- This query does NOT use the index (no prefix match)
SELECT * FROM events WHERE created_at > '2024-01-01';
```

**Rule**: Index columns in order of selectivity (most selective first).

## Index Maintenance

Over time, indexes can bloat:

```sql
-- Check index bloat
SELECT 
  schemaname,
  tablename,
  indexname,
  pg_size_pretty(pg_relation_size(indexrelid)) AS index_size,
  idx_scan,
  idx_tup_read,
  idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY pg_relation_size(indexrelid) DESC;
```

If index is bloated, consider rebuilding:

```sql
-- Rebuild index
REINDEX INDEX CONCURRENTLY idx_users_email;
```

Or drop and recreate:

```sql
DROP INDEX CONCURRENTLY idx_users_email;
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);
```

## When Indexes Hurt Performance

Indexes aren't free:

1. **Write overhead**: Every INSERT/UPDATE/DELETE must update indexes
2. **Disk space**: Indexes consume storage
3. **Planning overhead**: Query planner must consider all indexes

Too many indexes slow down writes:

```sql
-- Table with 10 indexes
-- Every INSERT updates 10 indexes
-- Writes become slow
```

**Guideline**: Only index what you actually query.

### Finding Unused Indexes

```sql
-- Indexes that are never used
SELECT 
  schemaname,
  tablename,
  indexname,
  idx_scan,
  pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexrelname NOT LIKE 'pg_toast%'
ORDER BY pg_relation_size(indexrelid) DESC;
```

Drop unused indexes:

```sql
DROP INDEX CONCURRENTLY idx_never_used;
```

## PostgreSQL vs MySQL Differences

### PostgreSQL
- `CREATE INDEX CONCURRENTLY` supported
- `NOT VALID` for constraints
- Online index operations

### MySQL
- No `CONCURRENTLY` keyword
- Online DDL supported (5.6+)
- `ALGORITHM=INPLACE` and `LOCK=NONE`

```sql
-- MySQL online index
CREATE INDEX idx_users_email 
  ON users(email) 
  ALGORITHM=INPLACE, LOCK=NONE;
```

But support varies by index type and MySQL version.

## Summary: Index and Constraint Checklist

Creating an index/constraint in production:

- [ ] Am I using `CONCURRENTLY`? (PostgreSQL)
- [ ] Have I tested on production-sized data?
- [ ] Do I know how long it will take?
- [ ] Am I monitoring progress during creation?
- [ ] For constraints: Am I using `NOT VALID` + `VALIDATE`?
- [ ] Have I ensured data validity before adding constraints?
- [ ] Is this index actually needed? (checked query patterns)
- [ ] What's my rollback plan if this causes issues?

**Golden rules**:
1. Always use `CREATE INDEX CONCURRENTLY` in production
2. Always use `NOT VALID` + `VALIDATE` for constraints
3. Test on production-sized data first
4. Monitor during execution
5. Don't add indexes you don't need

---

**Next**: [Rollback and Failure Recovery](./07_rollback_and_failure_recovery.md) - When things go wrong
