# Migration Basics: How Migrations Actually Work

## What Do Migrations Track?

At their core, migration systems track **which schema changes have been applied** to a database.

Think of it like Git for your schema:
- Git tracks which commits have been merged
- Migrations track which schema changes have been applied

### The Migration Table

Every migration system creates a metadata table in your database:

```sql
-- PostgreSQL example (names vary by tool)
CREATE TABLE schema_migrations (
  version VARCHAR(255) PRIMARY KEY,
  applied_at TIMESTAMP DEFAULT NOW()
);

-- Or sometimes more detailed
CREATE TABLE migrations (
  id SERIAL PRIMARY KEY,
  version VARCHAR(255) UNIQUE NOT NULL,
  name VARCHAR(255),
  applied_at TIMESTAMP DEFAULT NOW(),
  execution_time_ms INTEGER
);
```

When you run migrations, the tool:

1. Reads all migration files in your project
2. Checks the metadata table to see which have been applied
3. Runs unapplied migrations in order
4. Records each successful migration in the metadata table

**Critical insight**: This metadata table is the source of truth. If it gets out of sync with reality, you're in trouble.

## Versioning and Ordering

Migrations must run in a specific order. There are two common versioning strategies:

### Sequential Numbering

```
001_create_users_table.sql
002_add_email_to_users.sql
003_create_posts_table.sql
```

**Pros**:
- Dead simple
- Obvious order
- Easy to understand

**Cons**:
- Merge conflicts in teams (two people create `003_...`)
- Have to renumber to insert migrations

**Used by**: Rails (old style), some simple tools

### Timestamp-Based

```
20240215100530_create_users_table.sql
20240215143022_add_email_to_users.sql
20240216091145_create_posts_table.sql
```

Format: `YYYYMMDDHHmmss_description.sql`

**Pros**:
- No merge conflicts (timestamps are unique)
- Order is chronological
- No renumbering needed

**Cons**:
- Can have unexpected order if clocks are wrong
- Slightly less human-readable

**Used by**: Rails (modern), Django, most modern tools

**Best practice**: Use timestamps. The merge conflict problem alone is worth it.

## Up vs Down Migrations

### Style 1: Separate Files

```
migrations/
  20240215100530_add_phone_to_users.up.sql
  20240215100530_add_phone_to_users.down.sql
```

**Up migration** (`*.up.sql`):
```sql
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
```

**Down migration** (`*.down.sql`):
```sql
ALTER TABLE users DROP COLUMN phone;
```

**Used by**: go-migrate, golang-migrate

### Style 2: Single File with Markers

```sql
-- Migration: add_phone_to_users
-- Up
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- Down
ALTER TABLE users DROP COLUMN phone;
```

**Used by**: Some Node.js tools

### Style 3: Programmatic (TypeScript/JavaScript)

```typescript
// 20240215100530_add_phone_to_users.ts
import { Migration } from 'migration-tool';

export default class AddPhoneToUsers implements Migration {
  async up(db: Database) {
    await db.schema.alterTable('users', (table) => {
      table.string('phone', 20);
    });
  }

  async down(db: Database) {
    await db.schema.alterTable('users', (table) => {
      table.dropColumn('phone');
    });
  }
}
```

**Used by**: Drizzle, TypeORM, Knex.js

### Which Style Should You Use?

**For critical production systems**: Raw SQL files (Style 1 or 2)
- You see exactly what runs
- No abstraction surprises
- Easy to review
- Portable across tools

**For rapid development**: Programmatic (Style 3)
- Faster to write
- Type-safe
- Can use programming logic
- But: harder to review, abstraction leaks

**My opinion**: Start with raw SQL. Graduate to programmatic only when you're comfortable and understand the generated SQL.

## Idempotency: Migrations That Can Run Twice

An idempotent migration can be run multiple times safely.

### Non-Idempotent (Dangerous)

```sql
-- This fails the second time
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
-- ERROR: column "phone" already exists
```

```sql
-- This creates duplicates
INSERT INTO roles (name) VALUES ('admin');
-- Now you have multiple 'admin' roles
```

### Idempotent (Safe)

```sql
-- PostgreSQL
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20);
```

```sql
-- Insert only if doesn't exist
INSERT INTO roles (name) 
SELECT 'admin' 
WHERE NOT EXISTS (
  SELECT 1 FROM roles WHERE name = 'admin'
);
```

```sql
-- Drop safely
DROP INDEX IF EXISTS idx_users_email;
```

### Why Idempotency Matters

Imagine this scenario:

1. Migration starts
2. Network hiccup
3. Migration tool crashes
4. Did the migration complete? You don't know
5. You run migrations again
6. Non-idempotent migration fails

With idempotent migrations:
- You can safely retry
- Partial failures are recoverable
- Less panic during incidents

**Best practice**: Make migrations idempotent when possible, especially for DDL changes.

## Why Manual Database Changes Are Dangerous

You're debugging production. You connect directly to the database:

```sql
-- "I'll just fix this quickly"
ALTER TABLE users ADD COLUMN debug_flag BOOLEAN DEFAULT false;
```

What just happened?

1. ✓ The column exists in production
2. ✗ No migration file exists in your repo
3. ✗ The migration metadata table doesn't know about this
4. ✗ Staging doesn't have this column
5. ✗ Local dev doesn't have this column
6. ✗ Other developers don't know about this
7. ✗ Next time you rebuild the database, it's missing

**This is called "schema drift"** and it's a nightmare.

Now your environments are out of sync:

```
Production:  users [ id, email, username, debug_flag ]
Staging:     users [ id, email, username ]
Local Dev:   users [ id, email, username ]
```

Code that works in production fails in staging. Tests pass locally but fail in production. Debugging becomes impossible.

### The Exception: Incident Response

Sometimes you need to make manual changes during an incident:

```sql
-- Production is down, users table is locked
-- You kill the blocking query
SELECT pg_terminate_backend(pid) 
FROM pg_stat_activity 
WHERE state = 'active' AND query LIKE '%users%';
```

This is fine. But afterward:

1. Document what you did
2. Create a migration file to match reality
3. Apply the migration to other environments
4. Mark it as already applied in production

**The rule**: Every schema change must have a migration file, even if it was applied manually first.

## Local Dev vs Shared Environments

### Local Development

```
Your DB → Your migrations → Your code
```

- Fast iteration
- Can drop and recreate anytime
- Safe to experiment

**Common workflow**:
```bash
# Oops, bad migration
$ migrate down
$ vim migration.sql
$ migrate up
$ # Try again
```

You can "undo" freely because it's just your local data.

### Staging / Production

```
Shared DB → Migrations (one-way) → Multiple app instances
```

- Changes are permanent
- Can't easily "undo"
- Affects everyone
- Migrations run once and are recorded

**Common workflow**:
```bash
# Deploy to staging
$ git push origin main
$ # CI/CD runs migrations automatically
$ # Monitor for issues
$ # If OK, promote to production
```

You can't undo because:
1. Other people's work depends on the new schema
2. Data has been transformed
3. Rolling back might lose data

## Migration States

A migration can be in several states:

### 1. Pending

Migration file exists, but hasn't been applied:

```bash
$ migrate status
Pending migrations:
  20240215100530_add_phone_to_users.sql
  20240215143022_add_verified_to_users.sql
```

### 2. Applied

Migration has been successfully run:

```bash
$ migrate status
Applied migrations:
  20240215100530_add_phone_to_users.sql (applied 2 days ago)
```

### 3. Failed

Migration started but didn't complete:

```bash
$ migrate up
Running: 20240215100530_add_phone_to_users.sql
ERROR: syntax error at line 3
Migration failed!
```

**Danger zone**: The database might be partially modified. You need to:
1. Check what actually got applied
2. Fix the migration
3. Manually clean up if needed
4. Retry

### 4. Dirty

The metadata table says a migration is "in progress" but nothing is running:

```sql
SELECT * FROM schema_migrations WHERE dirty = true;
```

This happens when:
- Migration tool crashes
- Database connection drops
- Server restarts during migration

**Recovery**:
1. Check the database state manually
2. Determine if migration completed
3. Update metadata table manually
4. Fix the migration if needed

## Anatomy of a Good Migration File

```sql
-- Migration: add_phone_to_users
-- Created: 2024-02-15
-- Author: @yourname
-- 
-- Why: Users need to optionally provide phone numbers for 2FA
-- Risk: Low - adds nullable column, no locks expected
-- Rollback: Can drop column if needed within 24h (before data exists)
--
-- Tested on staging: 2024-02-14
-- Duration on 5M rows: 1.2 seconds

-- Start transaction if DDL is transactional (PostgreSQL yes, MySQL no)
BEGIN;

-- Make the change
ALTER TABLE users 
  ADD COLUMN IF NOT EXISTS phone VARCHAR(20);

-- Verify it worked
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns 
    WHERE table_name = 'users' AND column_name = 'phone'
  ) THEN
    RAISE EXCEPTION 'Migration failed: phone column not added';
  END IF;
END $$;

COMMIT;
```

Notice:
- Comments explaining context
- Risk assessment
- Test results
- Timing information
- Idempotent (IF NOT EXISTS)
- Verification step
- Transaction boundary

Compare to a bad migration:

```sql
alter table users add column phone varchar(20);
```

Both do the same thing technically. But the good one:
- Documents why
- Assesses risk
- Shows it was tested
- Can be reviewed effectively
- Can be understood 6 months later

## Migration File Naming Conventions

### Good Names

```
20240215100530_add_phone_to_users.sql
20240215143022_create_posts_table.sql
20240216091145_add_index_on_users_email.sql
20240216114520_backfill_user_roles.sql
```

Characteristics:
- Timestamp prefix
- Snake_case
- Verb + object
- Describes what it does
- Under 50 characters

### Bad Names

```
migration.sql           # Not unique
fix.sql                 # Not descriptive
update_20240215.sql     # Timestamp in wrong place
AddPhoneToUsers.sql     # CamelCase (harder to read)
new_migration_v2_final_really_final.sql  # Chaos
```

**Best practice**: Most tools auto-generate good names. Use the generators.

```bash
# go-migrate
$ migrate create -ext sql -dir migrations add_phone_to_users

# Drizzle
$ npm run drizzle-kit generate:pg

# Custom script
$ ./scripts/create_migration.sh "add_phone_to_users"
```

## Transaction Behavior

### PostgreSQL: DDL is Transactional

```sql
BEGIN;
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
-- Something goes wrong
ROLLBACK;  -- Reverts the ALTER TABLE
```

This is amazing! If your migration fails, Postgres can roll back the schema changes.

**But**: Some operations can't be in transactions:
- `CREATE INDEX CONCURRENTLY`
- `VACUUM`
- `CREATE DATABASE`

### MySQL: DDL is NOT Transactional

```sql
START TRANSACTION;
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
-- Something goes wrong
ROLLBACK;  -- Does nothing! ALTER TABLE already committed
```

In MySQL, each DDL statement implicitly commits.

**This means**: Failed migrations in MySQL leave the database in a partial state. You can't undo.

**Best practice for MySQL**:
- Test migrations exhaustively before production
- Have manual rollback plans
- Consider using tools like `pt-online-schema-change`

## The Migration Workflow: Step by Step

### 1. Create Migration File

```bash
$ migrate create -ext sql -dir migrations add_verified_to_users
Created: migrations/20240215100530_add_verified_to_users.up.sql
Created: migrations/20240215100530_add_verified_to_users.down.sql
```

### 2. Write the Migration

```sql
-- up
ALTER TABLE users ADD COLUMN verified BOOLEAN DEFAULT false NOT NULL;

-- down
ALTER TABLE users DROP COLUMN verified;
```

### 3. Test Locally

```bash
$ migrate up
Running: 20240215100530_add_verified_to_users.up.sql
Success!

$ # Test with your app
$ npm run dev

$ # Looks good? Test rollback
$ migrate down
Running: 20240215100530_add_verified_to_users.down.sql
Success!

$ # Apply again
$ migrate up
```

### 4. Commit to Git

```bash
$ git add migrations/20240215100530_add_verified_to_users.*
$ git commit -m "Add verified column to users"
$ git push
```

### 5. Deploy to Staging

```bash
$ # CI/CD automatically runs migrations
$ # Or manually:
$ ssh staging
$ cd /app
$ migrate up
```

### 6. Deploy to Production

```bash
$ # During deployment, migrations run first
$ # Then new app code deploys
$ # App code is compatible with both old and new schema
```

## Common Migration Patterns

### Adding a Column

```sql
-- Safe: nullable column
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
```

```sql
-- Risky: NOT NULL without DEFAULT in one step
-- (Locks table while rewriting)
ALTER TABLE users ADD COLUMN phone VARCHAR(20) NOT NULL DEFAULT '';
```

```sql
-- Safe: Add nullable first, backfill later
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
-- Later migration:
UPDATE users SET phone = '' WHERE phone IS NULL;
ALTER TABLE users ALTER COLUMN phone SET NOT NULL;
```

### Removing a Column

```sql
-- Step 1: Deploy code that doesn't use the column
-- Wait for deployment to complete
-- Step 2: Drop the column
ALTER TABLE users DROP COLUMN old_field;
```

Never drop a column while code is still using it!

### Renaming a Column

```sql
-- Don't do this:
ALTER TABLE users RENAME COLUMN email TO email_address;
-- Old code immediately breaks!
```

```sql
-- Instead (multi-step):
-- Step 1: Add new column
ALTER TABLE users ADD COLUMN email_address VARCHAR(255);

-- Step 2: Backfill
UPDATE users SET email_address = email WHERE email_address IS NULL;

-- Step 3: Deploy code that writes to both columns

-- Step 4: Deploy code that reads from new column

-- Step 5: Drop old column
ALTER TABLE users DROP COLUMN email;
```

## Migration Hygiene

### Do:
- ✓ One logical change per migration
- ✓ Test on production-sized data
- ✓ Add comments explaining context
- ✓ Make idempotent when possible
- ✓ Commit migration files to git
- ✓ Review migrations in PRs

### Don't:
- ✗ Edit applied migrations
- ✗ Delete old migration files
- ✗ Make manual schema changes
- ✗ Skip testing
- ✗ Combine unrelated changes
- ✗ Leave dirty state unresolved

## When Migrations Aren't Enough

Migrations are great for:
- Schema changes
- Small data changes
- Constraint additions/removals

Migrations are awkward for:
- Large data backfills (better as background jobs)
- Data cleanup (might need multiple passes)
- Complex transformations (might need application logic)

Sometimes you need a hybrid:
```sql
-- Migration creates structure
CREATE TABLE processed_orders (
  id SERIAL PRIMARY KEY,
  order_id INTEGER REFERENCES orders(id),
  processed_at TIMESTAMP
);
```

```typescript
// Background job backfills data
async function backfillProcessedOrders() {
  const batchSize = 1000;
  let offset = 0;
  
  while (true) {
    const orders = await db.query(`
      SELECT id FROM orders 
      WHERE status = 'processed'
      AND id NOT IN (SELECT order_id FROM processed_orders)
      LIMIT ${batchSize} OFFSET ${offset}
    `);
    
    if (orders.length === 0) break;
    
    // Process batch...
    offset += batchSize;
    await sleep(100); // Don't overwhelm DB
  }
}
```

**Best practice**: Migration creates the structure, background job fills it.

## Summary: The Mental Model

```
Migration File → Migration Tool → Database → Metadata Table
      ↓               ↓              ↓             ↓
  "What to do"   "Executor"    "Reality"    "What was done"
```

1. **Migration files** are the plan
2. **Migration tool** executes the plan
3. **Database** is the current state
4. **Metadata table** tracks what's been applied

All four must stay in sync. When they drift, you have problems.

The rest of this guide teaches you how to keep them in sync safely, even under production constraints.

---

**Next**: [Schema Changes and Locks](./03_schema_changes_and_locks.md) - Understanding what makes migrations scary
