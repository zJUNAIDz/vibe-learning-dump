# Common Rookie Mistakes: What Not to Do

## Mistake #1: Long-Running Transactions

### The Mistake

```sql
BEGIN;

-- Update millions of rows in one transaction
UPDATE users SET status = 'migrated' WHERE created_at < '2020-01-01';
-- This affects 10M rows

-- Add some more changes
UPDATE orders SET migrated = TRUE WHERE user_id IN (
  SELECT id FROM users WHERE status = 'migrated'
);
-- This affects 50M rows

COMMIT;
-- 15 minutes later...
```

### Why It's Bad

**During those 15 minutes**:
- Transaction holds locks on millions of rows
- WAL (Write-Ahead Log) grows massively
- Replication lag increases
- Autovacuum can't clean dead tuples
- Other queries wait
- Connection pool exhausts
- Application times out

**If it fails**:
- Rollback takes as long as the forward operation
- Database is frozen during rollback
- Complete disaster

### The Fix

**Batch updates in small transactions**:

```sql
DO $$
DECLARE
  batch_size INT := 1000;
  affected INT;
BEGIN
  LOOP
    UPDATE users 
    SET status = 'migrated' 
    WHERE id IN (
      SELECT id FROM users 
      WHERE status != 'migrated' 
        AND created_at < '2020-01-01'
      LIMIT batch_size
    );
    
    GET DIAGNOSTICS affected = ROW_COUNT;
    EXIT WHEN affected = 0;
    
    COMMIT;  -- Small transaction
    PERFORM pg_sleep(0.05);  -- Breathe
  END LOOP;
END $$;
```

**Rule**: Never update more than a few thousand rows in one transaction in production.

## Mistake #2: Adding NOT NULL Without Default

### The Mistake

```sql
-- On a table with millions of rows
ALTER TABLE users ADD COLUMN phone VARCHAR(20) NOT NULL;
```

### Why It's Bad

**PostgreSQL < 11**:
- Requires full table rewrite
- Acquires ACCESS EXCLUSIVE lock
- Takes minutes on large tables
- Blocks all reads and writes

**PostgreSQL 11+**:
- Still problematic if table has existing rows
- Validation takes time
- Risk of failure if any NULLs exist

**Old app code**:
```typescript
// Old code doesn't set phone
await db.query('INSERT INTO users (email) VALUES ($1)', [email]);
// 💥 ERROR: null value in column "phone" violates not-null constraint
```

### The Fix

**Multi-step approach**:

```sql
-- Step 1: Add nullable column (fast)
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- Step 2: Set default for new rows
ALTER TABLE users ALTER COLUMN phone SET DEFAULT '';

-- Step 3: Deploy app code that sets phone

-- Step 4: Backfill existing rows
UPDATE users SET phone = '' WHERE phone IS NULL;

-- Step 5: Add NOT NULL (now safe)
ALTER TABLE users ALTER COLUMN phone SET NOT NULL;

-- Step 6: Remove default if you want
ALTER TABLE users ALTER COLUMN phone DROP DEFAULT;
```

**Rule**: Never add NOT NULL in a single step on existing tables.

## Mistake #3: Blocking Writes with Index Creation

### The Mistake

```sql
-- Creating index on large table
CREATE INDEX idx_events_user_id ON events(user_id);
```

Without `CONCURRENTLY`, this acquires a SHARE lock:
- Reads: ✅ OK
- Writes: ❌ BLOCKED

### Why It's Bad

On a table with writes:
```sql
-- Meanwhile...
INSERT INTO events (user_id, type) VALUES (123, 'click');
-- ⏸️ Waiting for index creation...
-- ⏸️ Still waiting...
-- ⏸️ Still waiting...
-- 💥 Timeout
```

Result: Application can't write, errors spike, production down.

### The Fix

```sql
-- Always use CONCURRENTLY in production
CREATE INDEX CONCURRENTLY idx_events_user_id ON events(user_id);
```

**Yes, it takes longer. That's fine.**

**Rule**: Never create an index without CONCURRENTLY on a production table (PostgreSQL).

## Mistake #4: Forgetting About Old App Versions

### The Mistake

```sql
-- Migration drops column
ALTER TABLE users DROP COLUMN username;

-- Deploy new app code immediately after
```

**Timeline**:
```
14:00:00 - Migration runs, drops username
14:00:05 - New app code starts deploying
14:00:30 - Old app still running (rolling deploy)
14:00:31 - Old app tries to query username
           💥 ERROR: column "username" does not exist
```

### Why It's Bad

In a rolling deployment:
- Old and new app versions run simultaneously
- Old version expects old schema
- New schema breaks old version
- Partial outage during deploy

### The Fix

**Backward compatibility**:

```sql
-- Phase 1: Deploy code that doesn't use username
// Remove all references to username

-- Phase 2: Wait for complete rollout

-- Phase 3: Drop column (safe now)
ALTER TABLE users DROP COLUMN username;
```

Or use expand-migrate-contract pattern.

**Rule**: Schema changes must be compatible with both old and new app versions during deploy.

## Mistake #5: Assuming "Small Table" Means Safe

### The Mistake

```sql
-- "Settings table only has 1 row, this is fine"
ALTER TABLE settings ADD CONSTRAINT check_theme 
  CHECK (theme IN ('light', 'dark'));
```

### Why It's Bad

Table size doesn't matter. Query patterns matter.

If `settings` is queried on **every request**:
```typescript
// Every request reads settings
const settings = await db.query('SELECT * FROM settings LIMIT 1');
```

Even a 100ms lock means:
- 1000+ requests wait (if you get 10K req/sec)
- Connection pool exhausts
- Health checks fail
- Application down

### The Fix

**Consider query patterns**:

```sql
-- Use NOT VALID pattern even for small tables
ALTER TABLE settings 
  ADD CONSTRAINT check_theme 
  CHECK (theme IN ('light', 'dark')) 
  NOT VALID;

ALTER TABLE settings 
  VALIDATE CONSTRAINT check_theme;
```

Or deploy during low-traffic window.

**Rule**: Small table ≠ safe. Hot table = careful.

## Mistake #6: Trusting ORM Defaults Blindly

### The Mistake

```typescript
// TypeORM migration
await queryRunner.query(`
  ALTER TABLE users ADD COLUMN phone VARCHAR(20) NOT NULL
`);
```

**You didn't write this. TypeORM generated it.**

### Why It's Bad

ORMs generate SQL that might:
- Add NOT NULL immediately (unsafe)
- Create indexes without CONCURRENTLY (unsafe)
- Use suboptimal data types
- Make assumptions about your data

You don't know what's running until production breaks.

### The Fix

**Always review generated SQL**:

```bash
# Generate migration
$ npx typeorm migration:generate -n AddPhone

# Review the SQL
$ cat src/migrations/1234567890123-AddPhone.ts
```

**Edit if needed**:

```typescript
// Generated (unsafe)
await queryRunner.query(`ALTER TABLE users ADD phone VARCHAR(20) NOT NULL`);

// Edited (safe)
await queryRunner.query(`ALTER TABLE users ADD phone VARCHAR(20)`);
await queryRunner.query(`ALTER TABLE users ALTER COLUMN phone SET DEFAULT ''`);
// Backfill and add NOT NULL in separate migrations
```

**Rule**: Review generated SQL. Treat it as a starting point, not gospel.

## Mistake #7: No Testing on Production-Sized Data

### The Mistake

```bash
# Test on empty local database
$ docker run -d postgres
$ psql -c "CREATE TABLE users (id SERIAL, email TEXT);"
$ psql -c "INSERT INTO users (email) VALUES ('test@example.com');"

# Run migration
$ time psql -f migration.sql
# Time: 0.005 seconds

# Ship it! 🚢
```

**Then in production**:
```bash
$ time psql production -f migration.sql
# Time: 45 minutes
# 💥 Production down
```

### Why It's Bad

Empty tables:
- No locks contend
- No index overhead
- No disk I/O constraints
- Instant operations

Production tables:
- Millions of rows
- Concurrent queries
- Index maintenance
- Disk I/O limits
- Real locks

**They're completely different.**

### The Fix

**Test on production-sized data**:

```bash
# Restore production backup
$ pg_restore -d test_db prod_backup.dump

# Time the migration
$ time psql test_db -f migration.sql
# Time: 2 minutes 34 seconds

# Now you know what to expect
```

Or use synthetic data:

```sql
-- Generate 10M rows
INSERT INTO users (email, created_at)
SELECT 
  'user' || i || '@example.com',
  NOW() - (RANDOM() * INTERVAL '2 years')
FROM generate_series(1, 10000000) AS i;
```

**Rule**: Always test on production-sized data before deploying.

## Mistake #8: Not Checking for Existing Data

### The Mistake

```sql
-- Add foreign key constraint
ALTER TABLE orders 
  ADD CONSTRAINT fk_orders_user 
  FOREIGN KEY (user_id) REFERENCES users(id);
```

**Boom**:
```
ERROR: insert or update on table "orders" violates foreign key constraint "fk_orders_user"
DETAIL: Key (user_id)=(12345) is not present in table "users".
```

### Why It's Bad

Your production data is dirty:
- Orphaned records (users deleted, orders remain)
- Test data that doesn't match constraints
- Legacy data from before constraints existed

Blindly adding constraints fails.

### The Fix

**Check first**:

```sql
-- Find orphaned orders
SELECT COUNT(*) 
FROM orders 
WHERE user_id NOT IN (SELECT id FROM users);

-- 1,234 orphaned orders found
```

**Clean up**:

```sql
-- Option 1: Delete orphaned records
DELETE FROM orders 
WHERE user_id NOT IN (SELECT id FROM users);

-- Option 2: Assign to a placeholder user
UPDATE orders 
SET user_id = (SELECT id FROM users WHERE email = 'deleted@example.com')
WHERE user_id NOT IN (SELECT id FROM users);
```

**Then add constraint**:

```sql
ALTER TABLE orders 
  ADD CONSTRAINT fk_orders_user 
  FOREIGN KEY (user_id) REFERENCES users(id)
  NOT VALID;

ALTER TABLE orders VALIDATE CONSTRAINT fk_orders_user;
```

**Rule**: Always check data integrity before adding constraints.

## Mistake #9: Forgetting Indexes After Column Type Change

### The Mistake

```sql
-- Change column type
ALTER TABLE users ALTER COLUMN email TYPE TEXT;

-- Forget about indexes...
```

**Problem**: Indexes on `email` might not work optimally or at all after type change.

### Why It's Bad

```sql
-- Before: email was VARCHAR(255), had index
-- After: email is TEXT

-- This query might not use the index anymore
SELECT * FROM users WHERE email = 'alice@example.com';
-- Seq Scan (slow!)
```

### The Fix

**Recreate indexes after type changes**:

```sql
-- Change type
ALTER TABLE users ALTER COLUMN email TYPE TEXT;

-- Drop old index
DROP INDEX CONCURRENTLY idx_users_email;

-- Recreate index
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);

-- Or rebuild in place
REINDEX INDEX CONCURRENTLY idx_users_email;
```

**Rule**: After changing column types, reconsider indexes.

## Mistake #10: Manual Database Changes (Schema Drift)

### The Mistake

```bash
# SSH into production
$ psql production

# "Quick fix"
production=# ALTER TABLE users ADD COLUMN debug_mode BOOLEAN DEFAULT FALSE;
```

**What you just did**:
- ❌ No migration file in git
- ❌ No record in migration history
- ❌ Staging doesn't have this column
- ❌ Local dev doesn't have this column
- ❌ Other engineers don't know about it
- ❌ Next database restore loses it

### Why It's Bad

**Schema drift**: Environments are out of sync.

```
Production:  users [id, email, username, debug_mode]
Staging:     users [id, email, username]
Local:       users [id, email, username]
```

Code works in production, fails in staging. Confusion. Bugs. Wasted time.

### The Fix

**Always use migrations**:

```bash
# Create migration
$ migrate create add_debug_mode

# Write SQL
-- migrations/001_add_debug_mode.sql
ALTER TABLE users ADD COLUMN debug_mode BOOLEAN DEFAULT FALSE;

# Apply to production
$ migrate -database "$PROD_URL" up

# Apply to other environments
$ migrate -database "$STAGING_URL" up

# Commit to git
$ git add migrations/001_add_debug_mode.sql
$ git commit -m "Add debug_mode column"
```

**Exception**: During active incident, manual fixes are OK. But afterward:
1. Create migration file to match
2. Mark as already applied in production
3. Apply to other environments

**Rule**: Every schema change must have a migration file, even if applied manually first.

## Mistake #11: Dropping Columns Too Quickly

### The Mistake

```sql
-- Old code stopped using username
-- Immediately drop it
ALTER TABLE users DROP COLUMN username;
```

**Then**:
```
💥 Old deployment still running in one datacenter
💥 Background worker still uses username
💥 Analytics queries break
```

### Why It's Bad

Code removal and schema removal are async:
- Rolling deploys take time
- Background workers might not be updated
- Analytics dashboards might query it
- Other teams' services might depend on it

### The Fix

**Wait before dropping**:

```sql
-- Week 1: Stop writing to username (code deploy)
-- Week 2: Stop reading from username (code deploy)
-- Week 3: Monitor for any queries using username
-- Week 4: Drop column (safe now)
ALTER TABLE users DROP COLUMN username;
```

Or mark as deprecated:

```sql
-- Add comment
COMMENT ON COLUMN users.username IS 'DEPRECATED: Use email instead. Will be removed 2024-03-01.';
```

**Rule**: Wait at least 1-2 weeks after code stops using a column before dropping it.

## Mistake #12: Not Monitoring During Migration

### The Mistake

```bash
# Run migration
$ migrate up

# Walk away, grab coffee ☕
```

**Meanwhile**:
- Migration hangs
- Locks pile up
- Application starts failing
- No one notices for 10 minutes
- Disaster

### Why It's Bad

Migrations can fail in unexpected ways:
- Lock timeout
- Disk space full
- Out of memory
- Network hiccup
- Conflicting transaction

If you're not watching, you won't catch it in time.

### The Fix

**Active monitoring**:

```bash
# Terminal 1: Run migration
$ migrate up

# Terminal 2: Watch locks
$ watch -n 2 'psql -c "SELECT COUNT(*) FROM pg_stat_activity WHERE wait_event_type = '\''Lock'\'';"'

# Terminal 3: Watch application
$ watch -n 5 'curl -s https://api.example.com/health | jq .status'

# Terminal 4: Watch error rates
$ watch -n 5 'curl -s https://api.example.com/metrics | grep error_rate'
```

Set alerts:

```sql
-- Alert if locks queue up
SELECT pg_sleep(1);
RAISE NOTICE 'Checking for locks...';

SELECT COUNT(*) FROM pg_stat_activity 
WHERE wait_event_type = 'Lock';
-- If > 10, something is wrong
```

**Rule**: Never walk away from a running migration in production.

## Summary: The Rookie Mistakes Checklist

Before deploying a migration, ask:

- [ ] Am I updating millions of rows in one transaction?
- [ ] Am I adding NOT NULL without preparation?
- [ ] Am I creating an index without CONCURRENTLY?
- [ ] Will old app versions break with this schema?
- [ ] Did I test on production-sized data?
- [ ] Did I check for data integrity issues first?
- [ ] Did I review ORM-generated SQL?
- [ ] Am I dropping a column immediately after code change?
- [ ] Will I be monitoring during execution?
- [ ] Do I have a migration file in git?
- [ ] Is this a "small table" that's actually hot?
- [ ] Did I consider failure modes?

**If you answer "yes" to any of the first questions, stop and rethink your approach.**

## The Pattern of Mistakes

Notice a theme?

**Rookie mistakes come from**:
1. **Impatience**: Doing things in one step instead of multiple
2. **Optimism**: "It'll be fine" without testing
3. **Ignorance**: Not understanding locking behavior
4. **Assumptions**: "Small table = safe"

**Senior engineers**:
1. **Patient**: Multi-phase deployments
2. **Paranoid**: Test everything, assume the worst
3. **Knowledgeable**: Understand implications
4. **Empirical**: Measure, don't guess

**The shift**: From "make it work" to "make it safe."

That's the journey this guide is about.

---

**Next**: [Migrations as Team Practice](./12_migrations_as_team_practice.md) - How teams coordinate safely
