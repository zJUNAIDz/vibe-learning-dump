# Data Migrations: Transforming Production Data Safely

## Schema Migration vs Data Migration

**Schema migration**: Changes the structure
```sql
ALTER TABLE users ADD COLUMN full_name VARCHAR(255);
```

**Data migration**: Transforms the content
```sql
UPDATE users SET full_name = first_name || ' ' || last_name;
```

Schema migrations are scary because of locks. Data migrations are scary because of:
- **Scale**: Millions of rows to update
- **Transactions**: Long-running transactions can block other queries
- **Data integrity**: Risk of corrupting data
- **Reversibility**: Hard to undo transformations

## The Naive Approach (Don't Do This)

```sql
-- Update 10 million rows in one transaction
UPDATE users 
SET preferences = '{"theme": "light", "notifications": true}'::jsonb
WHERE preferences IS NULL;
```

**What happens**:

1. **Transaction starts**
2. PostgreSQL begins updating rows
3. Acquires row locks on millions of rows
4. Transaction log (WAL) grows enormously
5. Takes 10+ minutes
6. During this time:
   - Locks accumulate
   - Replication lags
   - Backups may fail
   - Autovacuum can't clean dead tuples
   - Disk fills with transaction logs

7. **If it fails** (timeout, connection drop, etc.):
   - PostgreSQL must rollback
   - Rollback takes as long or longer than the forward operation
   - Your database is effectively frozen

**Never update millions of rows in one transaction.**

## Batch Updates: The Foundation

Update data in small, manageable chunks:

```sql
-- Update in batches of 1000
DO $$
DECLARE
  rows_affected INT;
BEGIN
  LOOP
    UPDATE users 
    SET preferences = '{"theme": "light"}'::jsonb
    WHERE id IN (
      SELECT id FROM users 
      WHERE preferences IS NULL 
      LIMIT 1000
    );
    
    GET DIAGNOSTICS rows_affected = ROW_COUNT;
    EXIT WHEN rows_affected = 0;
    
    COMMIT;  -- Release locks after each batch
    PERFORM pg_sleep(0.1);  -- Don't overwhelm the database
  END LOOP;
END $$;
```

**Why this is better**:
- ✓ Each batch is a separate transaction
- ✓ Locks are released between batches
- ✓ Other queries can proceed
- ✓ Can monitor progress
- ✓ Can pause or kill without massive rollback
- ✓ Transaction log stays manageable

## Backfilling Data Safely

Common scenario: You added a column, now need to populate it.

### Step-by-Step Safe Backfill

```sql
-- 1. Create helper function
CREATE OR REPLACE FUNCTION backfill_user_slugs(batch_size INT DEFAULT 1000)
RETURNS TABLE(batch_num INT, rows_updated INT, elapsed_ms INT) AS $$
DECLARE
  batch INT := 0;
  affected INT;
  start_time TIMESTAMP;
BEGIN
  LOOP
    start_time := clock_timestamp();
    
    -- Update one batch
    WITH to_update AS (
      SELECT id FROM users 
      WHERE slug IS NULL 
      LIMIT batch_size
      FOR UPDATE SKIP LOCKED  -- Don't wait on locked rows
    )
    UPDATE users 
    SET slug = LOWER(REGEXP_REPLACE(username, '[^a-zA-Z0-9]', '-', 'g'))
    WHERE id IN (SELECT id FROM to_update);
    
    GET DIAGNOSTICS affected = ROW_COUNT;
    EXIT WHEN affected = 0;
    
    batch := batch + 1;
    
    RETURN QUERY SELECT 
      batch,
      affected,
      EXTRACT(MILLISECONDS FROM (clock_timestamp() - start_time))::INT;
    
    COMMIT;
    PERFORM pg_sleep(0.05);  -- 50ms breather
  END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 2. Run the backfill with monitoring
SELECT * FROM backfill_user_slugs(1000);

-- Output:
-- batch_num | rows_updated | elapsed_ms
-- ----------+--------------+-----------
--         1 |         1000 |         45
--         2 |         1000 |         43
--         3 |         1000 |         47
--       ...
```

**Benefits**:
- Progress visibility
- Performance metrics
- Can stop anytime
- Resumable (WHERE slug IS NULL)

### Backfill in Application Code

Sometimes it's easier to backfill from application code:

```typescript
async function backfillUserSlugs() {
  const batchSize = 1000;
  let processed = 0;
  
  while (true) {
    // Fetch batch of users needing backfill
    const users = await db.query(`
      SELECT id, username 
      FROM users 
      WHERE slug IS NULL 
      LIMIT $1
    `, [batchSize]);
    
    if (users.length === 0) {
      console.log(`Backfill complete! Processed ${processed} users`);
      break;
    }
    
    // Process batch
    for (const user of users) {
      const slug = generateSlug(user.username);
      await db.query(
        'UPDATE users SET slug = $1 WHERE id = $2',
        [slug, user.id]
      );
    }
    
    processed += users.length;
    console.log(`Processed ${processed} users...`);
    
    // Small delay to avoid overwhelming DB
    await new Promise(resolve => setTimeout(resolve, 100));
  }
}
```

**Pros**:
- Can use application logic (complex transformations)
- Easy to add error handling
- Can call external APIs if needed
- Familiar programming environment

**Cons**:
- Slower than pure SQL (network roundtrips)
- More complex coordination

**Use SQL when**: Simple transformations, database-resident data

**Use application code when**: Complex logic, external dependencies

## Monitoring Progress

### Track Remaining Work

```sql
-- How many rows left?
SELECT COUNT(*) FROM users WHERE slug IS NULL;

-- What percentage complete?
SELECT 
  COUNT(*) FILTER (WHERE slug IS NOT NULL) * 100.0 / COUNT(*) as percent_complete
FROM users;
```

### Monitor Performance Impact

```sql
-- Watch active queries
SELECT 
  pid,
  state,
  query_start,
  NOW() - query_start AS duration,
  query
FROM pg_stat_activity
WHERE state = 'active'
ORDER BY duration DESC;

-- Watch for lock contention
SELECT 
  COUNT(*) FILTER (WHERE wait_event_type = 'Lock') as waiting_on_locks,
  COUNT(*) as total_connections
FROM pg_stat_activity;

-- Check replication lag (if applicable)
SELECT 
  client_addr,
  state,
  sync_state,
  pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn) AS lag_bytes
FROM pg_stat_replication;
```

## Avoiding Long Transactions

### The Problem

```sql
BEGIN;

-- This looks like it should be fine
UPDATE posts SET status = 'archived' WHERE created_at < '2020-01-01';
-- But if it updates 5M rows, trouble...

-- Meanwhile, other queries wait
-- Transaction log grows
-- Connections time out

COMMIT;  -- Or worse, rollback...
```

Long transactions:
- Hold locks
- Block vacuum
- Cause replication lag
- Exhaust connection pools
- Make rollback expensive

### The Solution: Explicit Batching

```sql
-- Process in controlled batches
DO $$
DECLARE
  batch_size INT := 5000;
  total_updated INT := 0;
  batch_updated INT;
BEGIN
  LOOP
    -- Each iteration is a separate transaction
    UPDATE posts 
    SET status = 'archived'
    WHERE id IN (
      SELECT id FROM posts
      WHERE status != 'archived'
        AND created_at < '2020-01-01'
      LIMIT batch_size
    );
    
    GET DIAGNOSTICS batch_updated = ROW_COUNT;
    total_updated := total_updated + batch_updated;
    
    EXIT WHEN batch_updated = 0;
    
    RAISE NOTICE 'Updated % rows (% total)', batch_updated, total_updated;
    COMMIT;
    PERFORM pg_sleep(0.1);
  END LOOP;
  
  RAISE NOTICE 'Complete! Updated % total rows', total_updated;
END $$;
```

## Online Migrations: Keep the App Running

**Online migration**: Data transformation that happens while the application continues running.

### Challenge: Dual Writes

You're migrating from one schema to another. During the transition:
- Old code writes to old schema
- New code writes to new schema
- Both must stay in sync

```sql
-- Old schema: separate first_name, last_name columns
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  first_name VARCHAR(100),
  last_name VARCHAR(100)
);

-- New schema: single full_name column
ALTER TABLE users ADD COLUMN full_name VARCHAR(255);
```

**Phase 1**: Deploy code that writes to both

```typescript
// Old code
await db.query(
  'UPDATE users SET first_name = $1, last_name = $2 WHERE id = $3',
  [firstName, lastName, userId]
);

// New code (dual write)
await db.query(`
  UPDATE users 
  SET first_name = $1, 
      last_name = $2,
      full_name = $1 || ' ' || $2
  WHERE id = $3
`, [firstName, lastName, userId]);
```

**Phase 2**: Backfill existing data

```sql
UPDATE users 
SET full_name = first_name || ' ' || last_name
WHERE full_name IS NULL;
```

**Phase 3**: Deploy code that reads from new column

```typescript
// Old code read from old columns
const user = await db.query('SELECT first_name, last_name FROM users WHERE id = $1');
const fullName = `${user.first_name} ${user.last_name}`;

// New code reads from new column
const user = await db.query('SELECT full_name FROM users WHERE id = $1');
const fullName = user.full_name;
```

**Phase 4**: Stop writing to old columns

**Phase 5**: Drop old columns

```sql
ALTER TABLE users DROP COLUMN first_name;
ALTER TABLE users DROP COLUMN last_name;
```

**This is tedious**. It requires multiple deploys. But it's **zero downtime**.

## Dealing with Data Integrity

### Validation During Migration

```sql
-- Backfill with validation
DO $$
DECLARE
  invalid_count INT;
BEGIN
  -- Perform update
  UPDATE products 
  SET price_cents = (price_dollars * 100)::INT
  WHERE price_cents IS NULL;
  
  -- Validate result
  SELECT COUNT(*) INTO invalid_count
  FROM products 
  WHERE price_cents IS NOT NULL 
    AND price_dollars IS NOT NULL
    AND price_cents != (price_dollars * 100)::INT;
  
  IF invalid_count > 0 THEN
    RAISE EXCEPTION 'Data integrity violation: % products have mismatched prices', invalid_count;
  END IF;
  
  RAISE NOTICE 'Validation passed';
END $$;
```

### Checksums and Reconciliation

For critical migrations, validate before and after:

```sql
-- Before migration: capture state
CREATE TABLE migration_checksum AS
SELECT 
  COUNT(*) as total_rows,
  SUM(id) as sum_ids,
  MD5(STRING_AGG(id::TEXT || email, ',' ORDER BY id)) as data_hash
FROM users;

-- Perform migration
UPDATE users SET ...;

-- After migration: verify
DO $$
DECLARE
  original_total INT;
  current_total INT;
BEGIN
  SELECT total_rows INTO original_total FROM migration_checksum;
  SELECT COUNT(*) INTO current_total FROM users;
  
  IF original_total != current_total THEN
    RAISE EXCEPTION 'Row count mismatch: expected %, got %', original_total, current_total;
  END IF;
  
  RAISE NOTICE 'Row count verified';
END $$;

-- Clean up
DROP TABLE migration_checksum;
```

## Idempotent Data Migrations

Make your data migrations rerunnable:

```sql
-- Not idempotent
UPDATE users SET email = LOWER(email);
-- Run twice: no harm, but wastes time


-- Idempotent: only update what needs updating
UPDATE users 
SET email = LOWER(email)
WHERE email != LOWER(email);
-- Run twice: second run updates 0 rows
```

```sql
-- Not idempotent
UPDATE products SET price = price * 1.1;
-- Run twice: price increases by 21%! (1.1 * 1.1 = 1.21)

-- Idempotent: track what's been updated
ALTER TABLE products ADD COLUMN price_updated_v2 BOOLEAN DEFAULT FALSE;

UPDATE products 
SET price = price * 1.1,
    price_updated_v2 = TRUE
WHERE price_updated_v2 = FALSE;
-- Run twice: second run updates 0 rows
```

## Background Jobs vs Migrations

Sometimes the data transformation doesn't belong in a migration.

### Use Migration When:
- ✓ Simple transformation
- ✓ Required for schema change to make sense
- ✓ Reasonable to complete during deploy
- ✓ Can batch efficiently

### Use Background Job When:
- ✗ Complex business logic
- ✗ Calls external APIs
- ✗ Might take hours/days
- ✗ Failure is acceptable (can retry)
- ✗ Needs monitoring/alerting

**Example**: Generating image thumbnails

```sql
-- Migration just adds the column
ALTER TABLE uploads ADD COLUMN thumbnail_url TEXT;

-- Background job processes the queue
-- (runs after deploy, not during)
```

```typescript
// Celery / Bull / background worker
async function generateThumbnails() {
  while (true) {
    const upload = await db.query(`
      SELECT id, image_url 
      FROM uploads 
      WHERE thumbnail_url IS NULL 
      LIMIT 1
    `);
    
    if (!upload) break;
    
    // Expensive operation
    const thumbnail = await imageProcessor.generateThumbnail(upload.image_url);
    const url = await s3.upload(thumbnail);
    
    await db.query(
      'UPDATE uploads SET thumbnail_url = $1 WHERE id = $2',
      [url, upload.id]
    );
  }
}
```

**Guideline**: If the transformation takes more than a few minutes total, consider a background job.

## Handling Failures Gracefully

### Savepoints for Partial Rollback

```sql
BEGIN;

-- Create savepoint before risky operation
SAVEPOINT before_price_update;

UPDATE products SET price = new_calculated_price WHERE category = 'electronics';

-- Check if it looks right
DO $$
DECLARE
  avg_price NUMERIC;
BEGIN
  SELECT AVG(price) INTO avg_price FROM products WHERE category = 'electronics';
  
  IF avg_price > 10000 THEN
    RAISE EXCEPTION 'Average price suspiciously high: %', avg_price;
  END IF;
END $$;

-- If exception raised, rollback to savepoint
EXCEPTION WHEN OTHERS THEN
  ROLLBACK TO SAVEPOINT before_price_update;
  RAISE NOTICE 'Rolled back price update';
  RAISE;

COMMIT;
```

### Atomic Batch Updates with CTEs

```sql
-- Update with validation in one statement
WITH to_update AS (
  SELECT id, email, LOWER(email) AS email_lower
  FROM users
  WHERE email != LOWER(email)
  LIMIT 1000
),
validated AS (
  SELECT * FROM to_update
  WHERE email_lower LIKE '%@%'  -- Basic validation
)
UPDATE users 
SET email = validated.email_lower
FROM validated
WHERE users.id = validated.id;
```

## Performance Optimization

### Use EXPLAIN ANALYZE

Before running a large update, understand the query plan:

```sql
EXPLAIN ANALYZE
UPDATE users 
SET last_login = NOW()
WHERE status = 'active';
```

Look for:
- **Seq Scan**: Scanning entire table (slow on large tables)
- **Index Scan**: Using an index (faster)
- **Execution time**: Multiply by number of batches

### Add Indexes for Migration

If your backfill uses a WHERE clause that scans the entire table:

```sql
-- Slow: Seq Scan
UPDATE users 
SET migrated = TRUE
WHERE subscription_id IS NOT NULL 
  AND migrated = FALSE;
```

Add a temporary index:

```sql
-- Speed up the migration
CREATE INDEX CONCURRENTLY idx_users_migration 
ON users(subscription_id) 
WHERE migrated = FALSE;

-- Run migration (now uses index)
UPDATE users 
SET migrated = TRUE
WHERE subscription_id IS NOT NULL 
  AND migrated = FALSE;

-- Clean up
DROP INDEX idx_users_migration;
```

### Parallel Processing

For very large datasets, parallelize:

```sql
-- Split by ID range
-- Terminal 1:
UPDATE users SET migrated = TRUE WHERE id >= 0 AND id < 1000000;

-- Terminal 2:
UPDATE users SET migrated = TRUE WHERE id >= 1000000 AND id < 2000000;

-- Terminal 3:
UPDATE users SET migrated = TRUE WHERE id >= 2000000 AND id < 3000000;
```

Or in application code, spawn multiple workers processing different chunks.

**Caution**: Don't overwhelm the database. Monitor load.

## Real-World Example: Normalizing Dates

**Scenario**: Dates were stored in multiple timezone formats. Normalize to UTC.

### Step 1: Add New Column

```sql
ALTER TABLE events ADD COLUMN occurred_at_utc TIMESTAMP;
```

### Step 2: Backfill in Batches

```sql
DO $$
DECLARE
  batch_size INT := 5000;
  offset_val INT := 0;
  rows_updated INT;
BEGIN
  LOOP
    UPDATE events
    SET occurred_at_utc = (occurred_at AT TIME ZONE timezone) AT TIME ZONE 'UTC'
    WHERE id IN (
      SELECT id FROM events
      WHERE occurred_at_utc IS NULL
      LIMIT batch_size
    );
    
    GET DIAGNOSTICS rows_updated = ROW_COUNT;
    EXIT WHEN rows_updated = 0;
    
    RAISE NOTICE 'Updated % rows', rows_updated;
    COMMIT;
    PERFORM pg_sleep(0.1);
  END LOOP;
END $$;
```

### Step 3: Validate

```sql
-- Any rows not converted?
SELECT COUNT(*) FROM events WHERE occurred_at_utc IS NULL;

-- Spot check: dates should be similar
SELECT 
  occurred_at,
  occurred_at_utc,
  ABS(EXTRACT(EPOCH FROM (occurred_at - occurred_at_utc))) as diff_seconds
FROM events
LIMIT 100;
```

### Step 4: Deploy Code Using New Column

### Step 5: Drop Old Column

```sql
ALTER TABLE events DROP COLUMN occurred_at;
ALTER TABLE events RENAME COLUMN occurred_at_utc TO occurred_at;
```

## Summary: Data Migration Checklist

Before a data migration:

- [ ] Have I tested on production-sized data?
- [ ] Am I using batches (not one huge transaction)?
- [ ] Is there progress monitoring?
- [ ] Can I pause/resume if needed?
- [ ] Is it idempotent?
- [ ] Have I validated the transformation logic?
- [ ] Do I have a rollback plan?
- [ ] Am I monitoring database load during execution?
- [ ] If this fails partway, what's the recovery procedure?

**Golden Rule**: Data migrations should be boring. Slow and steady. No surprises.

---

**Next**: [Zero-Downtime Migrations](./05_zero_downtime_migrations.md) - The expand-migrate-contract pattern
