# Real-World Migration Scenarios: Battle-Tested Patterns

## Scenario 1: Renaming a Column Used by Production Traffic

**Context**: You need to rename `users.username` to `users.handle`. Table has 10M rows. Application gets 1000+ req/sec.

### The Wrong Way (Causes Outage)

```sql
-- Single migration
ALTER TABLE users RENAME COLUMN username TO handle;
```

**Result**:
- ✅ Rename succeeds (fast, metadata-only)
- 💥 Old application code breaks immediately
- 💥 All queries fail: "column username does not exist"
- 💥 Production down

### The Right Way: Multi-Phase

**Phase 1: Expand (Week 1, Monday)**

```sql
-- Migration 1: Add new column
ALTER TABLE users ADD COLUMN handle VARCHAR(100);

-- Create index on new column
CREATE INDEX CONCURRENTLY idx_users_handle ON users(handle);
```

```typescript
// Deploy: Dual-write application code
async function updateUser(id: number, value: string) {
  await db.query(`
    UPDATE users 
    SET username = $1, handle = $1  -- Write to both
    WHERE id = $2
  `, [value, id]);
}

// Still read from old column
async function getUser(id: number) {
  const user = await db.query('SELECT id, username FROM users WHERE id = $1');
  return user.username;
}
```

**Phase 2: Backfill (Week 1, Tuesday)**

```sql
-- Migration 2: Copy data
UPDATE users 
SET handle = username 
WHERE handle IS NULL;
```

Run in batches if large:

```sql
DO $$
BEGIN
  LOOP
    UPDATE users 
    SET handle = username 
    WHERE id IN (
      SELECT id FROM users WHERE handle IS NULL LIMIT 5000
    );
    EXIT WHEN NOT FOUND;
    COMMIT;
  END LOOP;
END $$;
```

**Phase 3: Switch Reads (Week 1, Thursday)**

```typescript
// Deploy: Read from new column
async function getUser(id: number) {
  const user = await db.query('SELECT id, handle FROM users WHERE id = $1');
  return user.handle;  // Use new column
}

// Still dual-write (for safety)
async function updateUser(id: number, value: string) {
  await db.query(`
    UPDATE users 
    SET username = $1, handle = $1 
    WHERE id = $2
  `, [value, id]);
}
```

Monitor for issues. If problems arise, roll back to reading from `username`.

**Phase 4: Stop Dual Writing (Week 2, Monday)**

```typescript
// Deploy: Only write to new column
async function updateUser(id: number, value: string) {
  await db.query(`
    UPDATE users 
    SET handle = $1 
    WHERE id = $2
  `, [value, id]);
}
```

**Phase 5: Contract (Week 2, Thursday)**

```sql
-- Migration 3: Drop old column
ALTER TABLE users DROP COLUMN username;

-- Rename index if desired
ALTER INDEX idx_users_handle RENAME TO idx_users_username;
```

### Timeline

```
Week 1, Mon:    Add column + dual write
Week 1, Tue:    Backfill data
Week 1, Thu:    Switch reads
Week 2, Mon:    Stop dual writes
Week 2, Thu:    Drop old column

Total: ~10 days
```

**Cost**: 5 deployments, 10 days

**Benefit**: Zero downtime, zero data loss, safe rollback at each step

## Scenario 2: Splitting a Table

**Context**: `users` table has become too large. Split into `users` and `user_profiles`.

### Current State

```sql
users:
  id, email, password_hash, first_name, last_name, bio, avatar_url, settings
```

### Target State

```sql
users:
  id, email, password_hash

user_profiles:
  user_id, first_name, last_name, bio, avatar_url, settings
```

### Multi-Phase Approach

**Phase 1: Create New Table**

```sql
-- Migration 1
CREATE TABLE user_profiles (
  user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  first_name VARCHAR(100),
  last_name VARCHAR(100),
  bio TEXT,
  avatar_url TEXT,
  settings JSONB
);

CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);
```

**Phase 2: Dual Write**

```typescript
// Deploy: Write to both tables
async function updateUser(userId: number, data: UserData) {
  await db.transaction(async (trx) => {
    // Write auth data to users
    await trx.query(`
      UPDATE users 
      SET email = $1, password_hash = $2 
      WHERE id = $3
    `, [data.email, data.passwordHash, userId]);
    
    // Write profile data to user_profiles
    await trx.query(`
      INSERT INTO user_profiles (user_id, first_name, last_name, bio, avatar_url, settings)
      VALUES ($1, $2, $3, $4, $5, $6)
      ON CONFLICT (user_id) DO UPDATE SET
        first_name = EXCLUDED.first_name,
        last_name = EXCLUDED.last_name,
        bio = EXCLUDED.bio,
        avatar_url = EXCLUDED.avatar_url,
        settings = EXCLUDED.settings
    `, [userId, data.firstName, data.lastName, data.bio, data.avatarUrl, data.settings]);
  });
}

// Still read from old table
async function getUser(userId: number) {
  return await db.query('SELECT * FROM users WHERE id = $1', [userId]);
}
```

**Phase 3: Backfill New Table**

```sql
-- Migration 2: Copy existing data
INSERT INTO user_profiles (user_id, first_name, last_name, bio, avatar_url, settings)
SELECT id, first_name, last_name, bio, avatar_url, settings
FROM users
ON CONFLICT (user_id) DO NOTHING;
```

Batch if large:

```sql
-- Batch backfill
INSERT INTO user_profiles (user_id, first_name, last_name, bio, avatar_url, settings)
SELECT id, first_name, last_name, bio, avatar_url, settings
FROM users
WHERE id >= $1 AND id < $2
ON CONFLICT (user_id) DO NOTHING;
```

**Phase 4: Switch Reads**

```typescript
// Deploy: Read from new table structure
async function getUser(userId: number) {
  const [user, profile] = await Promise.all([
    db.query('SELECT id, email FROM users WHERE id = $1', [userId]),
    db.query('SELECT * FROM user_profiles WHERE user_id = $1', [userId]),
  ]);
  
  return { ...user, ...profile };
}

// Still dual write
```

**Phase 5: Drop Old Columns**

```sql
-- Migration 3: Remove old columns
ALTER TABLE users DROP COLUMN first_name;
ALTER TABLE users DROP COLUMN last_name;
ALTER TABLE users DROP COLUMN bio;
ALTER TABLE users DROP COLUMN avatar_url;
ALTER TABLE users DROP COLUMN settings;
```

**Phase 6: Stop Dual Writing**

```typescript
// Deploy: Only write to new structure
async function updateUser(userId: number, data: UserData) {
  await db.transaction(async (trx) => {
    await trx.query('UPDATE users SET email = $1 WHERE id = $2', [data.email, userId]);
    await trx.query('UPDATE user_profiles SET first_name = $1 WHERE user_id = $2', [data.firstName, userId]);
  });
}
```

## Scenario 3: Merging Tables

**Context**: Separate `email_preferences` and `notification_preferences` tables. Merge into single `user_preferences`.

### Current State

```sql
email_preferences:
  user_id, newsletter, marketing, digest

notification_preferences:
  user_id, push_enabled, sms_enabled, in_app
```

### Target State

```sql
user_preferences:
  user_id, newsletter, marketing, digest, push_enabled, sms_enabled, in_app
```

### Approach

**Phase 1: Create Unified Table**

```sql
CREATE TABLE user_preferences (
  user_id INTEGER PRIMARY KEY REFERENCES users(id),
  newsletter BOOLEAN DEFAULT TRUE,
  marketing BOOLEAN DEFAULT FALSE,
  digest BOOLEAN DEFAULT TRUE,
  push_enabled BOOLEAN DEFAULT TRUE,
  sms_enabled BOOLEAN DEFAULT FALSE,
  in_app BOOLEAN DEFAULT TRUE
);
```

**Phase 2: Backfill from Both Tables**

```sql
INSERT INTO user_preferences (
  user_id, newsletter, marketing, digest, push_enabled, sms_enabled, in_app
)
SELECT 
  u.id,
  COALESCE(ep.newsletter, TRUE),
  COALESCE(ep.marketing, FALSE),
  COALESCE(ep.digest, TRUE),
  COALESCE(np.push_enabled, TRUE),
  COALESCE(np.sms_enabled, FALSE),
  COALESCE(np.in_app, TRUE)
FROM users u
LEFT JOIN email_preferences ep ON ep.user_id = u.id
LEFT JOIN notification_preferences np ON np.user_id = u.id
ON CONFLICT (user_id) DO NOTHING;
```

**Phase 3: Dual Write**

```typescript
// Write to all three tables
async function updatePreferences(userId: number, prefs: Preferences) {
  await db.transaction(async (trx) => {
    // Old tables
    await trx.query('UPDATE email_preferences SET newsletter = $1 WHERE user_id = $2', [prefs.newsletter, userId]);
    await trx.query('UPDATE notification_preferences SET push_enabled = $1 WHERE user_id = $2', [prefs.pushEnabled, userId]);
    
    // New table
    await trx.query('UPDATE user_preferences SET newsletter = $1, push_enabled = $2 WHERE user_id = $3', [prefs.newsletter, prefs.pushEnabled, userId]);
  });
}
```

**Phase 4: Switch Reads**

```typescript
// Read from new table
async function getPreferences(userId: number) {
  return await db.query('SELECT * FROM user_preferences WHERE user_id = $1', [userId]);
}
```

**Phase 5: Drop Old Tables**

```sql
DROP TABLE email_preferences;
DROP TABLE notification_preferences;
```

## Scenario 4: Changing a Primary Key

**Context**: `orders` table uses sequential INTEGER id. Need to switch to UUID for distributed system.

### The Challenge

Primary keys are referenced everywhere:
- Foreign keys from other tables
- Application code
- External systems
- Cached data

**This is one of the hardest migrations.**

### Approach

**Phase 1: Add New UUID Column**

```sql
-- Migration 1
ALTER TABLE orders ADD COLUMN uuid UUID DEFAULT gen_random_uuid() NOT NULL;

CREATE UNIQUE INDEX CONCURRENTLY idx_orders_uuid ON orders(uuid);
```

**Phase 2: Backfill UUIDs**

```sql
-- Ensure all rows have UUIDs
UPDATE orders SET uuid = gen_random_uuid() WHERE uuid IS NULL;
```

**Phase 3: Add UUID to Related Tables**

```sql
-- Migration 2
ALTER TABLE order_items ADD COLUMN order_uuid UUID;

-- Backfill
UPDATE order_items 
SET order_uuid = orders.uuid 
FROM orders 
WHERE order_items.order_id = orders.id;
```

**Phase 4: Dual-Key Lookups**

```typescript
// Application supports both id and uuid
async function getOrder(identifier: number | string) {
  if (typeof identifier === 'string') {
    return await db.query('SELECT * FROM orders WHERE uuid = $1', [identifier]);
  } else {
    return await db.query('SELECT * FROM orders WHERE id = $1', [identifier]);
  }
}

// New records use uuid as reference
async function createOrderItem(orderUuid: string, item: Item) {
  await db.query(`
    INSERT INTO order_items (order_uuid, product_id, quantity)
    VALUES ($1, $2, $3)
  `, [orderUuid, item.productId, item.quantity]);
}
```

**Phase 5: Switch Primary Key**

```sql
-- Migration 3 (requires downtime or very careful coordination)
BEGIN;

-- Remove old primary key constraint
ALTER TABLE orders DROP CONSTRAINT orders_pkey;

-- Remove old foreign keys
ALTER TABLE order_items DROP CONSTRAINT order_items_order_id_fkey;

-- Add new primary key
ALTER TABLE orders ADD PRIMARY KEY (uuid);

-- Add new foreign key using uuid
ALTER TABLE order_items 
  ADD CONSTRAINT order_items_order_uuid_fkey 
  FOREIGN KEY (order_uuid) REFERENCES orders(uuid);

COMMIT;
```

**Phase 6: Clean Up**

```sql
-- Migration 4
ALTER TABLE orders DROP COLUMN id;
ALTER TABLE order_items DROP COLUMN order_id;
```

**Note**: This migration often requires a brief maintenance window. True zero-downtime is extremely complex.

## Scenario 5: Large Backfill on Hot Table

**Context**: Add `last_active_at` to `users` table with 50M rows. Table gets constant writes.

### The Challenge

- Can't update 50M rows in one transaction (too slow, locks table)
- New writes happening constantly
- Need to backfill without impacting production

### Approach

**Phase 1: Add Column**

```sql
-- Migration 1: Add nullable column
ALTER TABLE users ADD COLUMN last_active_at TIMESTAMP;
```

**Phase 2: Deploy App to Populate New Rows**

```typescript
// New writes set last_active_at
async function recordActivity(userId: number) {
  await db.query(`
    UPDATE users 
    SET last_active_at = NOW() 
    WHERE id = $1
  `, [userId]);
}
```

**Phase 3: Backfill Strategically**

Don't backfill all at once. Prioritize:

```sql
-- Strategy: Backfill active users first

-- 1. Recent users (most important)
UPDATE users 
SET last_active_at = last_login 
WHERE last_active_at IS NULL 
  AND last_login > NOW() - INTERVAL '7 days'
  AND id IN (
    SELECT id FROM users 
    WHERE last_active_at IS NULL 
      AND last_login > NOW() - INTERVAL '7 days'
    LIMIT 10000
  );

-- 2. Then less recent users (batch by batch)
DO $$
BEGIN
  FOR batch IN 0..4999 LOOP
    UPDATE users 
    SET last_active_at = COALESCE(last_login, created_at)
    WHERE id >= batch * 10000 
      AND id < (batch + 1) * 10000
      AND last_active_at IS NULL;
    
    COMMIT;
    PERFORM pg_sleep(0.1);  -- Don't overwhelm DB
    
    IF batch % 100 = 0 THEN
      RAISE NOTICE 'Processed % rows', batch * 10000;
    END IF;
  END LOOP;
END $$;
```

**Phase 4: Background Worker**

```typescript
// Background job processes backfill slowly
async function backfillLastActiveAt() {
  const batchSize = 5000;
  
  while (true) {
    const result = await db.query(`
      WITH to_update AS (
        SELECT id FROM users 
        WHERE last_active_at IS NULL 
        LIMIT $1
      )
      UPDATE users 
      SET last_active_at = COALESCE(last_login, created_at)
      WHERE id IN (SELECT id FROM to_update)
    `, [batchSize]);
    
    if (result.rowCount === 0) break;
    
    console.log(`Backfilled ${result.rowCount} rows`);
    await sleep(1000);  // 1 second between batches
  }
  
  console.log('Backfill complete!');
}
```

Run this as a background task, not during migration.

**Phase 5: Add NOT NULL (Optional)**

Once 100% backfilled:

```sql
-- Migration 2: Add constraint
ALTER TABLE users ALTER COLUMN last_active_at SET NOT NULL;
```

## Scenario 6: Migration Gone Wrong - Postmortem

### The Incident

**Date**: 2024-01-15, 14:30 UTC

**What Happened**:
- Engineer deployed migration to add index on `events` table
- Used `CREATE INDEX` (not `CONCURRENTLY`)
- Table has 500M rows
- Index creation took 35 minutes
- All writes to `events` blocked
- Application errors spiked to 100%
- Cascade failure across services

### Timeline

```
14:30:00 - Migration deployed, starts
14:30:05 - First write attempts, blocked
14:31:00 - Connection pool exhausted
14:32:00 - Health checks failing
14:33:00 - Alerts fire
14:35:00 - On-call engineer paged
14:40:00 - Investigation starts
14:45:00 - Decision to kill migration
14:46:00 - Migration terminated via pg_terminate_backend()
14:46:30 - Partial index remains, marked INVALID
14:50:00 - Services recovering
15:05:00 - Full recovery
15:30:00 - Post-incident cleanup (drop invalid index)
```

### Root Cause

1. Migration did not use `CONCURRENTLY`
2. Not tested on production-sized data
3. No staging validation
4. Deployed during peak traffic hours

### What We Fixed

**Immediate**:
```sql
-- Terminated the migration
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE query LIKE '%CREATE INDEX%';

-- Cleaned up invalid index
DROP INDEX events_created_at_idx;
```

**Correct Migration** (deployed next day, off-hours):
```sql
-- Migration v2 (corrected)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_events_created_at 
  ON events(created_at);
```

Took 42 minutes but no production impact (zero writes blocked).

### Prevention Measures

1. **Added linting**:
```typescript
if (sql.includes('CREATE INDEX') && !sql.includes('CONCURRENTLY')) {
  throw new Error('CREATE INDEX must use CONCURRENTLY');
}
```

2. **Mandatory staging validation**:
```yaml
# CI/CD now requires staging test with production-sized data
- name: Test migration in staging
  run: restore_prod_snapshot && run_migration && validate
```

3. **Migration review checklist**:
- [ ] Uses CONCURRENTLY for indexes?
- [ ] Tested on production-sized data?
- [ ] Lock duration measured?
- [ ] Deployed during low-traffic window?

4. **Improved monitoring**:
```sql
-- Alert if lock wait time > 5 seconds
SELECT * FROM pg_stat_activity 
WHERE wait_event_type = 'Lock' 
  AND state_change < NOW() - INTERVAL '5 seconds';
```

### Lessons Learned

- Testing on empty tables gives false confidence
- Always use `CONCURRENTLY` for production indexes
- Timing matters (off-hours are safer)
- Have a kill-switch ready
- Document rollback procedures before deploying

## Summary: Migration Patterns Cheat Sheet

| Scenario | Pattern | Phases | Downtime |
|----------|---------|--------|----------|
| **Add column** | Add nullable → backfill → add constraint | 3 | None |
| **Drop column** | Stop writes → stop reads → drop | 3 | None |
| **Rename column** | Add new → dual write → backfill → switch reads → drop old | 5 | None |
| **Split table** | Create new → dual write → backfill → switch reads → drop old | 5 | None |
| **Merge tables** | Create unified → backfill → dual write → switch reads → drop old | 5 | None |
| **Change PK** | Add new → backfill → switch app → swap constraints | 4 | Brief window |
| **Large backfill** | Add column → app writes new → background backfill | 3 | None |
| **Add index** | CREATE CONCURRENTLY | 1 | None |
| **Add FK** | Add NOT VALID → validate | 2 | None |

**The pattern**: Expand → Migrate → Contract

**The principle**: Make changes backward-compatible, deploy in phases, never break running code.

---

**Next**: [Common Rookie Mistakes](./11_common_rookie_mistakes.md) - What not to do
