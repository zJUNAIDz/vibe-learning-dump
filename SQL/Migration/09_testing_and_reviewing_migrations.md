# Testing and Reviewing Migrations: Catching Issues Before Production

## The Problem

Most migration disasters happen because:
1. Migration wasn't tested on production-sized data
2. Migration wasn't reviewed by someone who understands the implications
3. Team didn't anticipate failure modes

**Prevention is cheaper than recovery.**

## Testing Locally: The Bare Minimum

### Don't Test on Empty Databases

```bash
# Wrong: Empty database
$ docker run -d postgres
$ psql -c "CREATE DATABASE test;"
$ migrate up
# ✓ Works fine

# Then in production:
$ migrate up
# 💥 Takes 10 minutes, locks table, brings down app
```

**Why**: Empty tables behave nothing like production tables.

### Test on Production-Sized Data

```bash
# 1. Get production data volume estimate
$ psql production -c "SELECT COUNT(*) FROM users;"
# 5,234,234 rows

# 2. Create test database with synthetic data
$ pgbench -i -s 1000 test_db  # Generates ~100M rows

# Or restore a production snapshot
$ pg_dump production | psql test_db

# 3. Test migration
$ time psql test_db -f migrations/001_add_phone.sql
# Time: 45.234 seconds

# Now you know: will take ~45 seconds in production
```

### Seed Realistic Data

```sql
-- seed_data.sql
-- Create realistic data volumes

-- 5M users
INSERT INTO users (email, username, created_at)
SELECT 
  'user' || i || '@example.com',
  'user' || i,
  NOW() - (RANDOM() * INTERVAL '2 years')
FROM generate_series(1, 5000000) AS i;

-- 50M events
INSERT INTO events (user_id, type, data, created_at)
SELECT 
  (RANDOM() * 5000000)::INT,
  CASE (RANDOM() * 3)::INT
    WHEN 0 THEN 'login'
    WHEN 1 THEN 'click'
    ELSE 'view'
  END,
  '{}'::jsonb,
  NOW() - (RANDOM() * INTERVAL '1 year')
FROM generate_series(1, 50000000) AS i;

-- Add indexes to match production
CREATE INDEX idx_events_user_id ON events(user_id);
CREATE INDEX idx_events_created_at ON events(created_at);
```

```bash
$ psql test_db -f seed_data.sql
$ # Now test your migration
```

## Measuring Migration Impact

### Time the Migration

```sql
\timing on

ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- Time: 1234.567 ms
```

**Interpretation**:
- < 100ms: Probably fine
- 100ms - 1s: Monitor closely
- 1s - 10s: Risky, needs planning
- > 10s: Very risky, use advanced techniques

### Monitor Lock Duration

```sql
-- Terminal 1: Start migration
BEGIN;
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
-- Don't commit yet

-- Terminal 2: Check what's locked
SELECT 
  locktype,
  relation::regclass,
  mode,
  granted
FROM pg_locks
WHERE relation = 'users'::regclass;

-- Output:
--  locktype | relation |        mode         | granted 
-- ----------+----------+---------------------+---------
--  relation | users    | AccessExclusiveLock | t
```

```sql
-- Terminal 2: Try a read
SELECT * FROM users LIMIT 1;
-- ⏸️ Blocked, waiting for lock
```

Now you know: reads will be blocked during this migration.

### Measure Write Impact

```bash
# Terminal 1: Run migration
$ time psql -c "ALTER TABLE users ADD COLUMN phone VARCHAR(20);"

# Terminal 2: Try writes during migration
$ while true; do
  psql -c "INSERT INTO users (email) VALUES ('test@example.com');" && echo "✓" || echo "✗"
  sleep 0.1
done

# Output:
# ✗ (blocked)
# ✗ (blocked)
# ✗ (blocked)
# ✓ (migration finished)
```

## Testing Rollback

Don't just test the "up" migration. Test rollback too.

```bash
# Apply migration
$ migrate up

# Verify it worked
$ psql -c "\d users"
# phone column exists

# Rollback
$ migrate down 1

# Verify rollback worked
$ psql -c "\d users"
# phone column gone

# Re-apply (test idempotency)
$ migrate up

# Verify again
$ psql -c "\d users"
# phone column exists
```

If rollback fails, you'll discover it locally, not in production.

## Testing Data Transformations

For data migrations, verify correctness:

```sql
-- Migration: Normalize emails to lowercase
UPDATE users SET email = LOWER(email);

-- Test: Spot check
SELECT email FROM users LIMIT 10;
-- All lowercase? ✓

-- Test: Verify no data loss
SELECT COUNT(*) FROM users WHERE email IS NULL;
-- 0 rows? ✓

-- Test: Check for duplicates (if email is unique)
SELECT email, COUNT(*) 
FROM users 
GROUP BY email 
HAVING COUNT(*) > 1;
-- 0 rows? ✓
```

## Staging Environment Testing

### Staging Should Mirror Production

**Staging checklist**:
- [ ] Same PostgreSQL version
- [ ] Similar data volume (or production restore)
- [ ] Similar hardware specs
- [ ] Similar application load
- [ ] Same migration tool version

**Don't trust staging if**:
- Empty database
- Different PostgreSQL version
- Tiny dataset
- No concurrent load

### Run Migration in Staging First

```bash
# 1. Restore production backup to staging
$ pg_dump production > backup.sql
$ psql staging < backup.sql

# 2. Run migration
$ time migrate -database "$STAGING_DATABASE_URL" up

# 3. Monitor application
$ curl https://staging.example.com/health
# Check: Does app still work?

# 4. Check for errors
$ psql staging -c "SELECT * FROM pg_stat_activity WHERE state = 'idle in transaction';"

# 5. Verify data integrity
$ psql staging -c "SELECT COUNT(*) FROM users WHERE phone IS NOT NULL;"
```

If staging works, production *probably* will too. But test carefully.

## Migration Review Checklist

Before merging a migration PR:

### Schema Changes

- [ ] Does this use `CONCURRENTLY` for indexes?
- [ ] Does this use `NOT VALID` + `VALIDATE` for constraints?
- [ ] Is the migration idempotent? (Can run twice safely)
- [ ] Does it add NOT NULL with proper phasing?
- [ ] Will this acquire an ACCESS EXCLUSIVE lock?
- [ ] How long will the lock be held? (tested on production-sized data)

### Data Safety

- [ ] Is there risk of data loss?
- [ ] Does dropping a column lose production data?
- [ ] Are data transformations reversible?
- [ ] Is there validation logic to ensure correctness?

### Compatibility

- [ ] Can old application code handle the new schema?
- [ ] Is this a breaking change requiring code deploy first?
- [ ] Are we using expand-migrate-contract pattern?
- [ ] Do background jobs / cron tasks need updating?

### Performance

- [ ] Will this slow down queries?
- [ ] Does this create an index that might not be used?
- [ ] Are we adding too many indexes? (write performance impact)
- [ ] Will this cause replication lag?

### Testing

- [ ] Has this been tested on production-sized data?
- [ ] Has rollback been tested?
- [ ] Has this been run in staging?
- [ ] Do we have timing measurements?

### Documentation

- [ ] Is there a comment explaining why this change is needed?
- [ ] Are lock implications documented?
- [ ] Is there a rollback plan?
- [ ] Is there an "oh shit" plan?

## Code Review: What to Look For

### Red Flags 🚩

```sql
-- 🚩 Dangerous: Adds NOT NULL without preparation
ALTER TABLE users ADD COLUMN phone VARCHAR(20) NOT NULL;
```

```sql
-- 🚩 Dangerous: Drops column still in use
ALTER TABLE users DROP COLUMN username;
```

```sql
-- 🚩 Dangerous: Renames column (breaking change)
ALTER TABLE users RENAME COLUMN email TO email_address;
```

```sql
-- 🚩 Dangerous: Index without CONCURRENTLY
CREATE INDEX idx_users_email ON users(email);
```

```sql
-- 🚩 Dangerous: Long transaction
BEGIN;
UPDATE users SET status = 'active' WHERE status IS NULL;  -- Millions of rows
COMMIT;
```

```sql
-- 🚩 Dangerous: Foreign key without NOT VALID
ALTER TABLE orders ADD CONSTRAINT fk_user 
  FOREIGN KEY (user_id) REFERENCES users(id);
```

### Green Flags ✅

```sql
-- ✅ Good: Nullable column
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
```

```sql
-- ✅ Good: Index with CONCURRENTLY
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);
```

```sql
-- ✅ Good: Constraint with NOT VALID
ALTER TABLE users ADD CONSTRAINT check_email 
  CHECK (email LIKE '%@%') NOT VALID;

ALTER TABLE users VALIDATE CONSTRAINT check_email;
```

```sql
-- ✅ Good: Batched data update
DO $$
DECLARE
  rows_updated INT;
BEGIN
  LOOP
    UPDATE users SET status = 'active' 
    WHERE id IN (
      SELECT id FROM users WHERE status IS NULL LIMIT 1000
    );
    GET DIAGNOSTICS rows_updated = ROW_COUNT;
    EXIT WHEN rows_updated = 0;
    COMMIT;
  END LOOP;
END $$;
```

```sql
-- ✅ Good: Idempotent
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON users(email);
```

## Pull Request Template for Migrations

```markdown
## Migration: Add phone column to users

### What
Adds optional `phone` column to `users` table for 2FA feature.

### Why
Product requirement: users should be able to add phone for 2FA.

### Schema Changes
- Adds `users.phone` VARCHAR(20), nullable

### Lock Analysis
- Uses `ADD COLUMN` (nullable) - metadata-only in PG 11+
- Expected lock duration: < 100ms
- Tested on staging with 5M rows: 45ms

### Compatibility
- ✅ Old code: Ignores the column (safe)
- ✅ New code: Can use the column
- ✅ No breaking changes

### Testing
- [x] Tested locally on 5M row table
- [x] Tested in staging
- [x] Verified rollback works
- [x] Migration is idempotent

### Rollback Plan
Can drop column if needed within 24 hours (before production data written).

### Deployment Plan
1. Run migration (metadata-only, fast)
2. Deploy app code (reads/writes phone)

### Oh Shit Plan
If app breaks: 
- Feature flag to disable 2FA feature
- Or rollback app code
- Column exists but unused (safe)

### Measurements
| Environment | Row Count | Duration | Lock Type |
|-------------|-----------|----------|-----------|
| Local       | 1K        | 5ms      | ACCESS EXCLUSIVE |
| Staging     | 5M        | 45ms     | ACCESS EXCLUSIVE |
| Production  | 5.2M      | ~50ms    | ACCESS EXCLUSIVE |

### Reviewer Checklist
- [ ] Schema change is safe
- [ ] Lock duration is acceptable
- [ ] Compatibility verified
- [ ] Testing is sufficient
- [ ] Rollback plan exists
```

## Automated Migration Checks

### Linting Migrations

```typescript
// scripts/lint-migration.ts
import fs from 'fs';

function lintMigration(filepath: string) {
  const sql = fs.readFileSync(filepath, 'utf-8');
  const issues: string[] = [];

  // Check 1: No naked ALTER TABLE (should use CONCURRENTLY)
  if (sql.includes('CREATE INDEX') && !sql.includes('CONCURRENTLY')) {
    issues.push('❌ CREATE INDEX without CONCURRENTLY detected');
  }

  // Check 2: No NOT NULL additions without staging
  if (sql.match(/ADD COLUMN.*NOT NULL/i)) {
    issues.push('❌ Adding NOT NULL column detected (should be multi-phase)');
  }

  // Check 3: No DROP COLUMN (usually needs planning)
  if (sql.includes('DROP COLUMN')) {
    issues.push('⚠️  DROP COLUMN detected (ensure code is deployed first)');
  }

  // Check 4: No RENAME COLUMN (breaking change)
  if (sql.includes('RENAME COLUMN')) {
    issues.push('⚠️  RENAME COLUMN detected (breaking change)');
  }

  // Check 5: Should be idempotent
  if (sql.includes('CREATE TABLE') && !sql.includes('IF NOT EXISTS')) {
    issues.push('⚠️  CREATE TABLE without IF NOT EXISTS');
  }

  return issues;
}

// Run on all migrations
const issues = lintMigration('./migrations/001_add_phone.sql');
if (issues.length > 0) {
  console.error('Migration issues found:');
  issues.forEach(issue => console.error(issue));
  process.exit(1);
}
```

Run in CI:

```yaml
# .github/workflows/ci.yml
- name: Lint migrations
  run: tsx scripts/lint-migration.ts
```

## Schema Diff Tools

Compare expected vs actual schema:

```typescript
// scripts/verify-schema.ts
import { drizzle } from 'drizzle-orm/postgres-js';
import postgres from 'postgres';
import * as schema from './src/db/schema';

const sql = postgres(process.env.DATABASE_URL);
const db = drizzle(sql, { schema });

async function verifySchema() {
  // Get actual schema
  const tables = await db.execute(`
    SELECT table_name, column_name, data_type 
    FROM information_schema.columns 
    WHERE table_schema = 'public'
  `);

  // Compare with expected schema
  // (Simplified - real implementation would be more thorough)
  
  const expected = Object.keys(schema);
  const actual = [...new Set(tables.map(t => t.table_name))];
  
  const missing = expected.filter(t => !actual.includes(t));
  
  if (missing.length > 0) {
    console.error('Schema drift detected! Missing tables:', missing);
    process.exit(1);
  }
  
  console.log('✓ Schema matches expected definition');
}

verifySchema();
```

Run after migrations:

```bash
$ migrate up
$ tsx scripts/verify-schema.ts
```

## Stress Testing Migrations

### Concurrent Load Testing

```bash
# Terminal 1: Run migration
$ time psql -c "CREATE INDEX CONCURRENTLY idx_users_email ON users(email);"

# Terminal 2: Generate concurrent load
$ pgbench -c 10 -t 1000 -f stress-test.sql test_db
```

```sql
-- stress-test.sql
INSERT INTO users (email) VALUES ('test@example.com');
UPDATE users SET last_login = NOW() WHERE id = 1;
SELECT * FROM users WHERE email LIKE '%test%';
```

**Monitor**:
- Does the migration complete?
- Do queries fail?
- What's the error rate?

### Lock Timeout Testing

```sql
-- Set aggressive timeout
SET lock_timeout = '5s';

-- Try migration
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- If it fails:
-- ERROR: canceling statement due to lock timeout

-- This means: migration can't acquire lock in 5 seconds
-- In production: might block indefinitely
```

## Regression Testing

Create a test suite that runs migrations:

```typescript
// test/migrations.test.ts
import { describe, it, beforeEach, afterEach } from 'vitest';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

describe('Migration regression tests', () => {
  beforeEach(async () => {
    // Reset test database
    await execAsync('psql test_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"');
  });

  it('should apply all migrations without error', async () => {
    const { stdout, stderr } = await execAsync('migrate -database "postgres://localhost/test_db" up');
    expect(stderr).toBe('');
  });

  it('should be idempotent', async () => {
    await execAsync('migrate -database "postgres://localhost/test_db" up');
    // Run again
    await execAsync('migrate -database "postgres://localhost/test_db" up');
    // Should not error
  });

  it('should rollback successfully', async () => {
    await execAsync('migrate -database "postgres://localhost/test_db" up');
    await execAsync('migrate -database "postgres://localhost/test_db" down 1');
    // Should not error
  });

  it('should preserve data through up/down cycle', async () => {
    await execAsync('migrate -database "postgres://localhost/test_db" up');
    
    // Insert data
    await execAsync('psql test_db -c "INSERT INTO users (email) VALUES (\'test@example.com\');"');
    
    const { stdout: beforeCount } = await execAsync('psql test_db -t -c "SELECT COUNT(*) FROM users;"');
    
    // Rollback and reapply
    await execAsync('migrate -database "postgres://localhost/test_db" down 1');
    await execAsync('migrate -database "postgres://localhost/test_db" up');
    
    const { stdout: afterCount } = await execAsync('psql test_db -t -c "SELECT COUNT(*) FROM users;"');
    
    expect(afterCount.trim()).toBe(beforeCount.trim());
  });
});
```

## Summary: Testing Checklist

Before running migration in production:

- [ ] Tested on production-sized data (not empty tables)
- [ ] Measured lock duration
- [ ] Tested concurrent queries during migration
- [ ] Tested rollback locally
- [ ] Ran in staging environment
- [ ] Verified data integrity after migration
- [ ] Checked for lock contention
- [ ] Monitored application during staging migration
- [ ] Code reviewed by senior engineer
- [ ] All automated checks pass
- [ ] Documentation complete (what, why, risks, rollback)

**The rule**: If you haven't tested risky scenarios, they'll happen in production.

---

**Next**: [Real-World Migration Scenarios](./10_real_world_migration_scenarios.md) - Battle-tested patterns and postmortems
