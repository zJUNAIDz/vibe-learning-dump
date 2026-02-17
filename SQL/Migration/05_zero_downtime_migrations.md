# Zero-Downtime Migrations: The Expand-Migrate-Contract Pattern

## The Problem with Traditional Migrations

Traditional approach:
1. Stop the application
2. Run migration
3. Deploy new code
4. Start the application

**Result**: Downtime. Users can't access your service.

In modern production systems, downtime is unacceptable:
- 24/7 global users
- SLA requirements (99.9% = only 43 minutes downtime per month)
- Revenue loss
- User trust

We need **zero-downtime migrations**: change the schema while the application keeps running.

## The Core Challenge

```
Old code expects old schema
New code expects new schema
Both need to run simultaneously during deployment
```

**Example**:

```typescript
// Old code
const user = await db.query('SELECT id, username FROM users');
console.log(user.username);

// New code
const user = await db.query('SELECT id, email FROM users');
console.log(user.email);
```

```sql
-- Migration renames column
ALTER TABLE users RENAME COLUMN username TO email;
```

**What happens during deployment**:
- Old code still running → queries for `username` → column doesn't exist → ERROR
- System breaks

**Solution**: Make schema changes that are compatible with both old and new code.

## The Expand-Migrate-Contract Pattern

This is the gold standard for zero-downtime migrations:

```
1. EXPAND: Add new schema elements (backward compatible)
2. MIGRATE: Update application to use new schema
3. CONTRACT: Remove old schema elements
```

Let's break this down with a real example.

## Example: Renaming `username` to `email`

### Phase 1: Expand (Add Without Breaking)

```sql
-- Migration 1: Add new column
ALTER TABLE users ADD COLUMN email VARCHAR(255);
```

Schema now:
```
users: [id, username, email]
       [1,  "alice",  NULL]
```

Old code continues working (still reads/writes `username`).

**Deploy**: Migration only, no code changes yet.

### Phase 2: Dual Write (Application Layer)

```typescript
// Update application to write to both columns
async function updateUser(id: number, value: string) {
  await db.query(`
    UPDATE users 
    SET username = $1, 
        email = $1  -- Write to both
    WHERE id = $2
  `, [value, id]);
}
```

**Deploy**: Code that dual-writes.

Now new data goes to both columns:
```
users: [id, username,  email]
       [1,  "alice",   NULL]       ← Old row
       [2,  "bob",     "bob"]      ← New row
```

### Phase 3: Migrate Data (Backfill)

```sql
-- Migration 2: Copy existing data
UPDATE users 
SET email = username 
WHERE email IS NULL;
```

Schema now:
```
users: [id, username,  email]
       [1,  "alice",   "alice"]
       [2,  "bob",     "bob"]
```

**Deploy**: Migration only.

### Phase 4: Read from New Column

```typescript
// Update application to read from new column
async function getUser(id: number) {
  const result = await db.query(`
    SELECT id, email FROM users WHERE id = $1
  `, [id]);
  return result.email;  // Use new column
}

// But still write to both (for safety)
async function updateUser(id: number, value: string) {
  await db.query(`
    UPDATE users 
    SET username = $1, email = $1 
    WHERE id = $2
  `, [value, id]);
}
```

**Deploy**: Code that reads from `email`.

### Phase 5: Contract (Remove Old Schema)

```typescript
// Stop writing to username
async function updateUser(id: number, value: string) {
  await db.query(`
    UPDATE users SET email = $1 WHERE id = $2
  `, [value, id]);
}
```

**Deploy**: Code that only uses `email`.

Wait for deployment to complete, then:

```sql
-- Migration 3: Drop old column
ALTER TABLE users DROP COLUMN username;
```

**Deploy**: Final migration.

### Timeline

```
Week 1, Monday:    Deploy migration + code (add email column, dual write)
Week 1, Tuesday:   Run backfill migration
Week 1, Wednesday: Deploy code (read from email)
Week 2, Monday:    Deploy code (stop writing to username)
Week 2, Tuesday:   Deploy migration (drop username column)
```

**Result**: Zero downtime. Users never noticed.

**Cost**: 5 deployments instead of 1. But worth it for zero downtime.

## Feature Flags + Migrations

Feature flags give you even more control:

```typescript
// Dual write, but read based on feature flag
async function getUser(id: number) {
  const user = await db.query(`
    SELECT id, username, email FROM users WHERE id = $1
  `, [id]);
  
  if (featureFlags.isEnabled('use-email-column')) {
    return user.email;
  } else {
    return user.username;
  }
}
```

If the new column has issues, flip the flag back without redeploying:

```typescript
featureFlags.disable('use-email-column');
```

This is defense in depth: migrations + feature flags.

## Backward Compatibility Rules

For zero-downtime, your schema changes must be **backward compatible**.

### Safe Changes (Backward Compatible)

✓ **Add Nullable Column**
```sql
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
```
Old code: Ignores the column
New code: Can use it

✓ **Add Table**
```sql
CREATE TABLE notifications (...);
```
Old code: Doesn't know about it
New code: Can use it

✓ **Add Index**
```sql
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);
```
Old code: Benefits from it automatically
New code: Benefits from it

✓ **Add Constraint (with NOTNOT VALID + validate later)**
```sql
ALTER TABLE users ADD CONSTRAINT check_email 
  CHECK (email LIKE '%@%') NOT VALID;
-- Later: ALTER TABLE users VALIDATE CONSTRAINT check_email;
```

### Unsafe Changes (Breaking Compatibility)

✗ **Drop Column**
```sql
ALTER TABLE users DROP COLUMN username;
```
Old code: Queries fail immediately

✗ **Rename Column**
```sql
ALTER TABLE users RENAME COLUMN username TO email;
```
Old code: Can't find `username`

✗ **Change Column Type**
```sql
ALTER TABLE users ALTER COLUMN email TYPE TEXT;
```
Might break old code depending on the change

✗ **Add NOT NULL to Existing Column**
```sql
ALTER TABLE users ALTER COLUMN email SET NOT NULL;
```
Old code might insert NULLs → constraint violation

✗ **Drop Table**
```sql
DROP TABLE old_table;
```
Old code: Queries fail

**Rule**: Additive changes are safe. Subtractive/mutative changes require multi-phase approach.

## Pattern: Adding a NOT NULL Column

### Wrong Way (Causes Downtime)

```sql
ALTER TABLE users ADD COLUMN phone VARCHAR(20) NOT NULL DEFAULT '';
```

Problems:
- In PostgreSQL: might lock table during rewrite
- Old code doesn't set `phone` on INSERT → might fail depending on version

### Right Way: Multi-Phase

**Phase 1**: Add as nullable

```sql
-- Migration 1
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
```

**Phase 2**: Update application to set the column

```typescript
// Old code
await db.query('INSERT INTO users (email) VALUES ($1)', [email]);

// New code
await db.query('INSERT INTO users (email, phone) VALUES ($1, $2)', [email, phone || '']);
```

**Phase 3**: Backfill existing rows

```sql
-- Migration 2
UPDATE users SET phone = '' WHERE phone IS NULL;
```

**Phase 4**: Add NOT NULL constraint

```sql
-- Migration 3
ALTER TABLE users ALTER COLUMN phone SET NOT NULL;
```

Now the constraint is safe because:
- No NULL values exist
- Application always sets the value

## Pattern: Dropping a Column

### Wrong Way

```sql
ALTER TABLE users DROP COLUMN deprecated_field;
```

Old code still queries it → instant errors.

### Right Way

**Phase 1**: Stop writing to the column

```typescript
// Old code
await db.query('UPDATE users SET deprecated_field = $1 WHERE id = $2', [value, id]);

// New code (remove the write)
// (no update to deprecated_field)
```

**Phase 2**: Stop reading from the column

```typescript
// Old code
const user = await db.query('SELECT id, email, deprecated_field FROM users WHERE id = $1');

// New code (remove from SELECT)
const user = await db.query('SELECT id, email FROM users WHERE id = $1');
```

**Phase 3**: Drop the column

```sql
ALTER TABLE users DROP COLUMN deprecated_field;
```

Timeline:
1. Deploy code that doesn't use the column
2. Wait for all old deployments to stop
3. Run migration to drop column

## Pattern: Changing Column Type

### Example: `price` from INTEGER to DECIMAL

Current:
```sql
price INTEGER  -- Stored in cents
```

Desired:
```sql
price DECIMAL(10, 2)  -- Stored as dollars.cents
```

### Multi-Phase Approach

**Phase 1**: Add new column

```sql
-- Migration 1
ALTER TABLE products ADD COLUMN price_decimal DECIMAL(10, 2);
```

**Phase 2**: Dual write

```typescript
// Write to both
await db.query(`
  UPDATE products 
  SET price = $1,           -- cents
      price_decimal = $1 / 100.0  -- dollars
  WHERE id = $2
`, [priceInCents, productId]);
```

**Phase 3**: Backfill

```sql
-- Migration 2
UPDATE products 
SET price_decimal = price / 100.0 
WHERE price_decimal IS NULL;
```

**Phase 4**: Switch reads

```typescript
// Old code
const price = product.price;  // cents

// New code
const price = product.price_decimal * 100;  // convert back to cents for now
```

**Phase 5**: Update application logic to work in dollars

```typescript
// Now use price_decimal directly
const price = product.price_decimal;  // dollars
```

**Phase 6**: Drop old column

```sql
-- Migration 3
ALTER TABLE products DROP COLUMN price;
ALTER TABLE products RENAME COLUMN price_decimal TO price;
```

Yes, this is tedious. But it's zero downtime.

## Dual Writes: Common Pitfalls

### Pitfall 1: Forgetting to Write to Old Column

```typescript
// Bug! Only writes to new column
await db.query('UPDATE users SET email = $1 WHERE id = $2', [value, id]);
// Old code reads username → sees stale data
```

**Solution**: During dual-write phase, **always** write to both.

### Pitfall 2: Transaction Consistency

```typescript
// Bug! Two separate transactions
await db.query('UPDATE users SET username = $1 WHERE id = $2', [value, id]);
await db.query('UPDATE users SET email = $1 WHERE id = $2', [value, id]);
// If second fails, data is inconsistent
```

**Solution**: Single transaction

```typescript
await db.query(`
  UPDATE users 
  SET username = $1, email = $1 
  WHERE id = $2
`, [value, id]);
```

### Pitfall 3: Forgetting About Background Jobs

```typescript
// Main app is updated to dual-write
// But background worker still only writes username
async function processQueue() {
  await db.query('UPDATE users SET username = $1 WHERE id = $2', [value, id]);
  // email is now stale!
}
```

**Solution**: Update **all** code paths that touch the column.

## Reading Old and New Schemas: Defensive Coding

```typescript
// Defensive: handle both schemas
async function getUser(id: number) {
  const user = await db.query(`
    SELECT id, username, email FROM users WHERE id = $1
  `, [id]);
  
  // Use new column if available, fall back to old
  return user.email || user.username;
}
```

This lets your code work during the transition:
- Before migration: `email` is NULL, uses `username`
- After migration: `email` is populated, uses `email`
- After cleanup: only `email` exists

## Real-World Example: Adding `users.email_verified`

**Goal**: Add `email_verified` boolean column.

### Problem

```sql
-- This is unsafe
ALTER TABLE users ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE;
```

Why unsafe?
- Old code doesn't know about it
- If old code does `INSERT INTO users (email) ...` it might fail (depending on DB version/settings)

### Safe Approach

**Deploy 1**: Add nullable column

```sql
ALTER TABLE users ADD COLUMN email_verified BOOLEAN;
```

```typescript
// Update app to write the column
await db.query(`
  INSERT INTO users (email, email_verified) 
  VALUES ($1, $2)
`, [email, false]);
```

**Deploy 2**: Backfill

```sql
UPDATE users SET email_verified = FALSE WHERE email_verified IS NULL;
```

**Deploy 3**: Add NOT NULL

```sql
ALTER TABLE users ALTER COLUMN email_verified SET NOT NULL;
```

**Timeline**: 3 deployments, but zero downtime.

## Blue-Green Deployments + Migrations

In a **blue-green deployment**, you run two environments:
- Blue: Current production
- Green: New version

**Challenge**: Both point to the same database.

```
Blue Environment (old code)  →  Database
Green Environment (new code) →  Database
```

For this to work:
1. Schema must be compatible with both code versions
2. Follow expand-migrate-contract
3. Run migrations **before** switching traffic

**Process**:
1. Deploy green environment (new code)
2. Run migrations (expand phase)
3. Green uses new schema, blue still works
4. Switch traffic from blue to green
5. Later: Contract phase (remove old schema)

## Rolling Deployments + Migrations

In **rolling deployments**, you gradually replace instances:

```
Instance 1: Old code
Instance 2: Old code  → Deploy → New code
Instance 3: Old code
Instance 4: Old code  → Deploy → New code
...
```

At any moment, old and new code run simultaneously.

**Critical**: Schema must be compatible with both versions.

**Strategy**:
1. Deploy backward-compatible schema change
2. Deploy new code (rolling)
3. Wait for rollout to complete
4. Deploy cleanup migration (remove old schema)

**Never**: Deploy a breaking schema change during a rolling deployment.

## Monitoring During Zero-Downtime Migrations

### Application Metrics

Watch:
- Error rates (should not spike)
- Latency (should not increase)
- Success rates (should stay constant)

```typescript
// Log when using old vs new schema
logger.info('Reading from new email column', { userId });

// Track feature flag usage
metrics.increment('user.read_from_email_column');
metrics.increment('user.read_from_username_column');
```

### Database Metrics

Watch:
- Lock contention
- Query performance
- Replication lag

```sql
-- Monitor for blocked queries
SELECT COUNT(*) 
FROM pg_stat_activity 
WHERE wait_event_type = 'Lock';
```

### Rollback Readiness

If new schema causes problems:

```typescript
// Feature flag to instantly switch back
if (featureFlags.isEnabled('use-email-column')) {
  return user.email;
} else {
  return user.username;  // Fallback
}
```

## Checklist: Zero-Downtime Migration

Before deploying:

- [ ] Is the schema change backward compatible?
- [ ] Can old code continue running with new schema?
- [ ] Am I using expand-migrate-contract pattern?
- [ ] Am I dual-writing to both old and new columns?
- [ ] Have I planned multiple deployment phases?
- [ ] Do I have feature flags for rollback?
- [ ] Am I monitoring application metrics?
- [ ] Have I tested with both old and new code running?
- [ ] Is there a clear rollback plan?

## The Mental Model

```
Traditional Migration:
Stop app → Migrate → Deploy new code → Start app
            ↑
        Downtime

Zero-Downtime Migration:
Expand → Deploy new code (still compatible) → Migrate data → Contract
  ↑              ↑                              ↑             ↑
Always       Never breaks                   Background    Cleanup
compatible
```

It's more work. It requires discipline. But it's the only way to maintain 24/7 availability.

## When Zero-Downtime Isn't Possible

Some migrations are inherently risky:

- Major refactoring (e.g., splitting tables)
- Data that can't be backfilled easily
- Third-party schema constraints

In these cases:
- Schedule maintenance window
- Communicate with users
- Do it quickly and safely
- Have rollback ready

**But**: 90% of migrations can be zero-downtime with expand-migrate-contract.

## Summary

**Zero-downtime migrations require**:
1. Backward-compatible schema changes
2. Multi-phase deployments
3. Dual writes during transition
4. Feature flags for safety
5. Monitoring and rollback plans

**The pattern**:
- **Expand**: Add new, keep old
- **Migrate**: Switch code gradually
- **Contract**: Remove old

It's slower. It's more deploys. But users never notice. And that's the point.

---

**Next**: [Indexes and Constraints](./06_indexes_constraints_migrations.md) - The silent production killers
