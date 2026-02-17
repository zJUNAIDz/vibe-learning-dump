# Rollback and Failure Recovery: When Things Go Wrong

## The Uncomfortable Truth About Rollbacks

**Most rollbacks are lies.**

Your migration tool says you can "rollback" by running the "down" migration. This works great locally. In production, it's often impossible or dangerous.

Why?

```sql
-- Up migration
ALTER TABLE users DROP COLUMN deprecated_field;

-- Down migration
ALTER TABLE users ADD COLUMN deprecated_field VARCHAR(255);
```

**Problem**: The data is gone. You can add the column back, but it's empty. You lost production data.

## When Rollbacks Are Actually Possible

### True Rollbacks (Safe to Revert)

✓ **Adding a nullable column**
```sql
-- Up: Add column
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- Down: Remove column (no data loss if column unused)
ALTER TABLE users DROP COLUMN phone;
```

✓ **Adding an index**
```sql
-- Up
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);

-- Down
DROP INDEX CONCURRENTLY idx_users_email;
```

✓ **Adding a table**
```sql
-- Up
CREATE TABLE new_feature (...);

-- Down
DROP TABLE new_feature;
```

**Condition**: No production data has been written yet.

✓ **Adding a constraint (if you can relax it)**
```sql
-- Up
ALTER TABLE products ADD CONSTRAINT check_price CHECK (price > 0);

-- Down
ALTER TABLE products DROP CONSTRAINT check_price;
```

### Irreversible Changes (Cannot Rollback)

✗ **Dropping a column**
```sql
-- Up: Drop column (data deleted!)
ALTER TABLE users DROP COLUMN phone;

-- Down: Can't recover the data
ALTER TABLE users ADD COLUMN phone VARCHAR(20);  -- Empty!
```

✗ **Data transformations**
```sql
-- Up: Destructive change
UPDATE users SET email = LOWER(email);

-- Down: Can't un-lowercase
UPDATE users SET email = ???  -- Original case is lost
```

✗ **Dropping a table**
```sql
-- Up
DROP TABLE old_feature;

-- Down
CREATE TABLE old_feature (...);  -- Empty!
```

✗ **Changing column type (with data loss)**
```sql
-- Up: Truncates data
ALTER TABLE logs ALTER COLUMN message TYPE VARCHAR(100);

-- Down: Can't restore truncated data
ALTER TABLE logs ALTER COLUMN message TYPE TEXT;
```

## The Real Rollback: Fix Forward

In production, you usually don't rollback. You **fix forward** with a new migration.

### Example: Bad Migration

```sql
-- Migration 001: Add NOT NULL too early
ALTER TABLE users ADD COLUMN phone VARCHAR(20) NOT NULL;
-- ERROR: column "phone" contains null values
```

Migration failed partway. Now what?

### Wrong Response: Try to Rollback

```sql
-- Try to undo
ALTER TABLE users DROP COLUMN phone;
-- But the column might be partially added
-- Or might have scattered NULL values
-- Chaos
```

### Right Response: Fix Forward

```sql
-- Migration 002: Fix the mistake
-- Remove the NOT NULL constraint (if column was added)
ALTER TABLE users ALTER COLUMN phone DROP NOT NULL;

-- Or if column wasn't added, add it correctly
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20);
```

Then in your code:
```typescript
// Handle both cases defensively
const phone = user.phone || null;
```

## Designing Reversible Migrations

Even if true rollback isn't always possible, design for reversibility:

### Store Original Values Before Transformation

```sql
-- Bad: Destructive
UPDATE users SET email = LOWER(email);

-- Better: Keep original
ALTER TABLE users ADD COLUMN email_original TEXT;
UPDATE users SET email_original = email;
UPDATE users SET email = LOWER(email);

-- Now can rollback
UPDATE users SET email = email_original;
ALTER TABLE users DROP COLUMN email_original;
```

### Use Feature Flags for Data Migrations

```typescript
// Don't hard-code the migration logic
async function getUser(id: number) {
  const user = await db.query('SELECT * FROM users WHERE id = $1', [id]);
  
  if (featureFlags.isEnabled('use-lowercase-email')) {
    return { ...user, email: user.email.toLowerCase() };
  }
  return user;
}
```

Now you can "rollback" by toggling the flag, without touching the database.

### Dual-Write During Risky Migrations

Keep both old and new representations:

```sql
-- Migration adds new column
ALTER TABLE users ADD COLUMN preferences_jsonb JSONB;

-- App writes to both
UPDATE users 
SET preferences_text = $1,
    preferences_jsonb = $2
WHERE id = $3;

-- If JSONB causes issues, fall back to text
SELECT COALESCE(preferences_jsonb::text, preferences_text) FROM users;
```

## Disaster Recovery Scenarios

### Scenario 1: Migration Hangs Indefinitely

**What happened**:
```sql
ALTER TABLE users ADD COLUMN status VARCHAR(20) DEFAULT 'active';
-- Running for 30 minutes... still going...
```

**Diagnosis**:
```sql
-- Check what's running
SELECT pid, query, state, state_change 
FROM pg_stat_activity 
WHERE state = 'active';

-- Check what's blocked
SELECT * FROM pg_locks WHERE NOT granted;
```

**Response**:

Option A: Let it finish (if close to done)
```sql
-- Monitor progress
SELECT * FROM pg_stat_progress_create_index;  -- If creating index
```

Option B: Kill it
```sql
-- Terminate the migration
SELECT pg_terminate_backend(pid) 
FROM pg_stat_activity 
WHERE query LIKE '%ALTER TABLE users%';
```

**After killing**:

PostgreSQL (transactional DDL):
```sql
-- Check if change was applied
\d users
-- If not, you're back to square one. Safe.
```

MySQL (non-transactional DDL):
```sql
-- Check what state the table is in
DESCRIBE users;
-- Might be partially modified. Ugh.
```

**Fix**:
```sql
-- Determine current state
\d users  -- or DESCRIBE users

-- Write a new migration to reach desired state
-- Either complete the change or undo partial changes
```

### Scenario 2: Migration Fails Partway

**What happened**:
```sql
BEGIN;
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
UPDATE users SET phone = '';  -- Syntax error or timeout
COMMIT;
-- ERROR: Something went wrong
```

**In PostgreSQL**:
- Transaction rolls back automatically
- Table is unchanged
- Safe to retry after fixing

**In MySQL**:
- ALTER TABLE already committed
- Column exists
- UPDATE didn't run
- Now you have a column full of NULLs

**Fix**:
```sql
-- Check current state
SELECT COUNT(*) FROM users WHERE phone IS NULL;

-- Write migration to complete the job
UPDATE users SET phone = '' WHERE phone IS NULL;
```

### Scenario 3: Migration Succeeds, App Breaks

**What happened**:
```sql
-- Migration dropped a column
ALTER TABLE users DROP COLUMN username;
```

```typescript
// Old code still running
const user = await db.query('SELECT id, username FROM users WHERE id = $1');
// ERROR: column "username" does not exist
```

**Immediate response**:

Option A: Rollback the app (if code deploy was recent)
```bash
# Revert to previous deployment
kubectl rollout undo deployment/api
```

Option B: Emergency migration to restore column
```sql
-- Recreate column (but data is gone!)
ALTER TABLE users ADD COLUMN username VARCHAR(255);

-- Try to recover from another source
UPDATE users SET username = email;  -- Temporary fix
```

Option C: Hot-patch the app
```typescript
// Emergency code push
const user = await db.query('SELECT id, email FROM users WHERE id = $1');
const username = user.email;  // Use email as username temporarily
```

**Long-term fix**:
- Restore data from backup
- Implement proper column deprecation (expand-migrate-contract)
- Add monitoring to catch this earlier

### Scenario 4: Data Corruption After Migration

**What happened**:
```sql
-- Migration transformed data incorrectly
UPDATE products SET price_cents = price_dollars * 100;
-- But some prices were already in cents!
```

**Detection**:
```sql
-- Prices look wrong
SELECT * FROM products WHERE price_cents > 1000000;
-- $10,000 items? Suspicious.
```

**Response**:

1. **Stop the bleeding**: Prevent further damage
```sql
-- Add validation constraint
ALTER TABLE products ADD CONSTRAINT check_price_sane 
  CHECK (price_cents < 100000)  -- Max $1000
  NOT VALID;
```

2. **Quantify the damage**:
```sql
-- How many rows affected?
SELECT COUNT(*) FROM products WHERE price_cents > 10000;
```

3. **Restore from backup**:
```bash
# Restore just the products table from backup
pg_restore -t products -d production backup.dump
```

4. **Or surgically fix**:
```sql
-- If pattern is identifiable
UPDATE products 
SET price_cents = price_cents / 100 
WHERE price_cents > 10000;
```

### Scenario 5: Migration Succeeds, Performance Tanks

**What happened**:
```sql
-- Added index that's actually slowing things down
CREATE INDEX idx_events_all_columns ON events(id, user_id, type, created_at, data);
```

**Detection**:
```sql
-- Queries are slower
EXPLAIN ANALYZE SELECT * FROM events WHERE user_id = 123;
-- Using the wrong index!
```

**Response**:

Immediate: Drop the index
```sql
DROP INDEX CONCURRENTLY idx_events_all_columns;
```

Investigate:
```sql
-- Check what indexes exist
\di+ events*

-- Check query plans
EXPLAIN SELECT * FROM events WHERE user_id = 123;
```

Fix:
```sql
-- Create a better index
CREATE INDEX CONCURRENTLY idx_events_user_id ON events(user_id);
```

### Scenario 6: "Dirty" Migration State

**What happened**:
Migration tool crashed mid-migration. Metadata table says migration is "in progress" but nothing is running.

```sql
SELECT * FROM schema_migrations WHERE dirty = true;
-- version | dirty
-- --------|------
-- 20240215| true
```

**Response**:

1. **Check actual database state**:
```sql
\d users  -- Does the change exist?
```

2. **If migration completed**:
```sql
-- Mark as complete
UPDATE schema_migrations SET dirty = false WHERE version = '20240215';
```

3. **If migration didn't complete**:
```sql
-- Mark as not applied
DELETE FROM schema_migrations WHERE version = '20240215';

-- Then retry
```

4. **Manual cleanup**:
```sql
-- Undo partial changes if necessary
ALTER TABLE users DROP COLUMN IF EXISTS partial_column;

-- Mark clean
UPDATE schema_migrations SET dirty = false WHERE version = '20240215';
```

## The Migration Postmortem

After a failed migration, write it up:

### Template

```markdown
# Migration Incident Postmortem

## What Happened
- Migration 20240215_add_phone_to_users started at 14:32 UTC
- Acquired lock on users table
- Blocked all writes for 23 minutes
- Application errors spiked to 100%
- Terminated migration at 14:55 UTC

## Root Cause
- Migration added NOT NULL column with DEFAULT on 50M row table
- Did not test on production-sized data
- Caused full table rewrite
- Did not use multi-step approach

## Impact
- 23 minutes of failed writes
- ~5000 affected users
- No data loss (PostgreSQL rolled back)

## Resolution
- Terminated migration
- Deployed new multi-step migration:
  1. Add nullable column
  2. Backfill in batches
  3. Add NOT NULL constraint
- Completed successfully in 3 hours (off-hours)

## Prevention
- [ ] Always test migrations on production-sized staging data
- [ ] Use NOT NULL in multiple steps
- [ ] Monitor lock duration
- [ ] Have kill procedure ready
- [ ] Better team review process
```

Share this with the team. Learn from it.

## Backup Strategies for Migrations

### Before Risky Migrations

```bash
# Full backup
pg_dump -Fc production > backup_before_migration.dump

# Or just the affected tables
pg_dump -Fc -t users -t orders production > backup_tables.dump
```

**Restore if needed**:
```bash
pg_restore -d production -c backup_before_migration.dump
```

### Point-in-Time Recovery

If you have WAL archiving enabled:

```bash
# Restore to just before the migration
pg_restore --target-time '2024-02-15 14:30:00'
```

This is the gold standard, but requires setup.

### Logical Replication

For critical migrations, use a replica:

1. Run migration on replica
2. Test thoroughly
3. Promote replica to primary
4. Old primary becomes standby

Zero downtime, full rollback capability.

## The "Oh Shit" Plan

Every migration should have an "oh shit" plan:

```markdown
# Migration: Add email_verified column

## Execution Plan
1. Add nullable column (fast)
2. Deploy app code
3. Backfill data
4. Add NOT NULL

## Oh Shit Plan

### If step 1 fails:
- Retry (idempotent)

### If step 2 fails (app deploy issues):
- Rollback app
- Column exists but unused (safe)

### If step 3 fails (backfill):
- Pause backfill
- Partial data is OK (app handles NULL)
- Resume later

### If step 4 fails (NOT NULL):
- Check for remaining NULLs
- Fix data issue
- Retry

### If app breaks after deploy:
- Feature flag: disable email_verified feature
- Or rollback app
- Column exists but unused (safe)

## Rollback
- Cannot truly rollback (data exists)
- Can disable feature in app
- Can drop column if no production data written
```

## Practicing Failure

In dev/staging:

1. Run migration
2. Kill it midway
3. Practice recovery
4. Document the procedure

**You want to have done this before production.**

```bash
# Staging drill
$ migrate up &
$ PID=$!
$ sleep 5
$ kill $PID
# Now what? Practice fixing it.
```

## When to Use Database Replicas

For extremely risky migrations:

1. **Test on replica first**
```bash
# Restore backup to test database
pg_restore -d test_db prod_backup.dump

# Run migration
psql test_db < migration.sql

# If it works, proceed to production
```

2. **Run on standby, then promote**
```bash
# On replica
psql replica -c "ALTER TABLE users ADD COLUMN phone VARCHAR(20);"

# Promote replica to primary
# Old primary becomes standby (automatic rollback path)
```

This is the safest approach for mission-critical systems.

## Summary: Rollback Readiness Checklist

Before running a migration:

- [ ] Do I have a recent backup?
- [ ] Have I tested rollback locally?
- [ ] Is rollback actually possible? Or is this irreversible?
- [ ] What's my "oh shit" plan?
- [ ] Can I fix forward if rollback isn't possible?
- [ ] Do I have feature flags to disable new behavior?
- [ ] Have I practiced this failure scenario?
- [ ] Is someone on-call who can help if things break?
- [ ] Do I know how to check if the migration partially applied?
- [ ] Can I kill the migration safely if it hangs?

**The truth**: You probably won't rollback. You'll fix forward. But having a plan reduces panic.

**The goal**: Not to prevent all failures. But to recover quickly and learn from them.

---

**Next**: [Tooling: Drizzle and go-migrate](./08_tooling_drizzle_go_migrate.md) - Practical tools for real migrations
