# Migration Philosophy: The Mindset Shift

## What Is a Database Migration, Really?

A migration is not just a SQL script that changes your database schema. That's the mechanical view, and it's dangerously incomplete.

**A migration is a contract evolution event.**

Your database schema is a contract between:
- Your application code
- Other services that depend on your data
- Background jobs and workers
- BI tools and analytics pipelines
- Sometimes, other teams' applications

When you change that schema, you're **breaking or modifying a contract** that running code depends on. The code doesn't pause while you deploy. Users don't stop clicking. Cron jobs don't wait for your migration to finish.

This is why migrations fail in production. Not because the SQL is wrong, but because **the contract change wasn't coordinated with the contract users**.

## Schema as an API

Think of your database schema like a public API:

```
// Your schema is like this API contract
interface User {
  id: number;
  email: string;
  username: string;
}
```

If you just delete `username` one day, every client breaks. Same with your database:

```sql
-- This is like deleting a field from your API
ALTER TABLE users DROP COLUMN username;
```

Except worse, because:
1. Your API has multiple versions running simultaneously (old app servers during deploy)
2. There's no compile-time check
3. The failure happens in production, under load, with real users

**Golden Rule**: Treat schema changes with the same caution you'd treat breaking API changes.

## Why Migrations Are Not "Just DDL Scripts"

Junior thinking:
> "I need to add a column, so I'll write `ALTER TABLE users ADD COLUMN phone VARCHAR(20);` and run it."

Senior thinking:
> "I need to add a column. Let me think through:
> - Will this lock the table?
> - How long will it take on a table with 50M rows?
> - Can old application code handle this column existing?
> - Can new application code handle this column being NULL initially?
> - What happens if the migration fails halfway?
> - How do I roll back if needed?
> - Do I need to backfill data?
> - Is this change reversible?"

The SQL is 10% of the problem. The other 90% is coordination, timing, and failure handling.

## Forward-Only vs Reversible Migrations

### Forward-Only Migrations

Some teams only write "up" migrations and never look back:

```sql
-- migrations/20240215_add_phone.sql
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
```

**Philosophy**: "We don't revert database changes. If there's a problem, we write a new migration to fix it."

**Pros**:
- Honest about what's actually possible
- No false confidence in "down" migrations that were never tested
- Forces you to think about compatibility

**Cons**:
- Can't easily test rollback locally
- Requires more discipline in design

### Reversible Migrations

Other teams write explicit "down" migrations:

```sql
-- Up
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- Down
ALTER TABLE users DROP COLUMN phone;
```

**Philosophy**: "Every change should be reversible so we can rollback if needed."

**The Lie**: This works great locally. In production, it's mostly theater.

Why? Because:

1. **Data Loss**: Dropping a column destroys data. Your "down" migration might succeed technically, but you just lost production data.

2. **Time Travel Doesn't Exist**: If your migration backfilled 10M rows with computed data, the "down" can't uncompute it.

3. **External State**: If your migration triggered external effects (sent emails, called APIs, updated caches), you can't undo those.

**Reality Check**: Down migrations are useful for local development and testing, but in production, you usually "fix forward" with a new migration instead.

## Why Migrations Fail in Real Life

### Mistake #1: Forgetting About Running Code

```sql
-- Migration
ALTER TABLE orders ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'pending';
```

Looks safe? It's not. Here's what happens:

1. Migration starts at 10:00:00
2. Takes 30 seconds on a large table (locks acquired)
3. Old app code is still running, trying to INSERT into orders
4. Old code doesn't know about the `status` column
5. In PostgreSQL: OK (uses default)
6. In MySQL: Might fail depending on version and settings

But even if it succeeds:

```typescript
// Old app code still running
const order = await db.query(`
  SELECT id, user_id, amount FROM orders WHERE id = $1
`);

// New code expects status to exist
if (order.status === 'pending') {  // Boom! Property doesn't exist
  // ...
}
```

**The Fix**: Deploy code and schema changes in a compatible order (we'll cover this in zero-downtime migrations).

### Mistake #2: Underestimating Lock Duration

```sql
-- "This is a small table, it'll be quick"
ALTER TABLE users ADD COLUMN preferences JSONB;
```

Reality:
- Table has 5M rows
- ALTER TABLE acquires an ACCESS EXCLUSIVE lock
- PostgreSQL needs to rewrite the table (pre-11, or if adding column with DEFAULT)
- Takes 45 seconds
- During those 45 seconds: **ALL queries to users table are blocked**
- Your API starts timing out
- Health checks fail
- Load balancer removes your instances
- Cascading failure

**The Fix**: Understand locking behavior. Add columns in multiple steps. Test on production-sized data.

### Mistake #3: Assuming Small Table = Safe

```sql
-- "Settings table only has one row, this is safe"
ALTER TABLE settings ADD CONSTRAINT check_valid_theme 
  CHECK (theme IN ('light', 'dark'));
```

Problems:
1. Even a small table locks
2. Every query waits during the lock
3. If settings is queried on every request (common!), your entire app freezes
4. Constraint validation can be slow if there's expensive logic

**The Fix**: Understand your query patterns, not just table size.

### Mistake #4: Trusting the ORM

```typescript
// Your ORM migration
await schema.table('users', (table) => {
  table.string('phone').notNullable();
});
```

What SQL does this generate? You don't know. The ORM might:
- Add NOT NULL immediately (unsafe on large tables)
- Not use CONCURRENTLY for indexes (causes locks)
- Generate different SQL on PostgreSQL vs MySQL
- Not handle default values correctly

**The Fix**: Always review the generated SQL. For critical migrations, write raw SQL.

## The Mindset Shift: Dev to Prod

### Development Mindset

- Fast iteration
- Break things
- Database is empty or has seed data
- Can drop and recreate anytime
- Migrations are quick
- Rollback is easy (just revert the migration file)

### Production Mindset

- Stability first
- Change is risk
- Database has years of production data
- Tables have millions/billions of rows
- Migrations can take minutes or hours
- Active traffic during migration
- Multiple app versions running simultaneously
- No downtime tolerance
- Rollback is complex or impossible

**The Bridge**: You need to think with a production mindset even in development. Test migrations on production-sized data. Consider failure modes before running.

## The Real Migration Lifecycle

```
┌─────────────────────────────────────────────────────┐
│ 1. Design the Change                                │
│    - What needs to change and why?                  │
│    - What's the safest approach?                    │
└─────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────┐
│ 2. Consider Backward Compatibility                  │
│    - Can old code work with new schema?             │
│    - Can new code work with old schema?             │
└─────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────┐
│ 3. Write the Migration                              │
│    - Explicit, reviewable SQL                       │
│    - Comments explaining tradeoffs                  │
└─────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────┐
│ 4. Test Locally (Production-Like Data)             │
│    - Seed database with realistic row counts        │
│    - Measure lock duration                          │
│    - Test with app code running                     │
└─────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────┐
│ 5. Review with Team                                 │
│    - Explain the approach                           │
│    - Discuss failure modes                          │
│    - Get sign-off from senior engineers             │
└─────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────┐
│ 6. Test in Staging (Real Data Volume)              │
│    - Production snapshot preferred                  │
│    - Measure actual duration                        │
│    - Monitor locks and performance                  │
└─────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────┐
│ 7. Plan the Deployment                              │
│    - Coordinate with team                           │
│    - Choose low-traffic window if needed            │
│    - Prepare monitoring                             │
│    - Have rollback plan ready                       │
└─────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────┐
│ 8. Execute in Production                            │
│    - Monitor actively                               │
│    - Watch for lock waits                           │
│    - Check application metrics                      │
│    - Be ready to intervene                          │
└─────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────┐
│ 9. Verify                                           │
│    - Check data integrity                           │
│    - Confirm application works                      │
│    - Monitor error rates                            │
└─────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────┐
│ 10. Document                                        │
│     - What was changed                              │
│     - Why this approach was chosen                  │
│     - Any gotchas for future                        │
└─────────────────────────────────────────────────────┘
```

Notice how "write the SQL" is step 3 out of 10? That's the point.

## Core Principles

### 1. Migrations Are One-Way Causality

You can't un-pour milk. Once a migration runs in production:
- Data is transformed
- Constraints are validated
- Indexes are built
- The world has moved forward

Your "down" migration is usually a new transformation, not a reversal.

### 2. Compatibility Is Your Responsibility

The database doesn't know about your application code. It's your job to ensure:
- Schema changes are compatible with running code
- Deploy order is safe
- Multiple versions can coexist

### 3. Boring Is Good

In production migrations:
- Clever is bad
- Simple is good
- Boring is great
- Manual steps are fine if they're safer

Don't optimize for elegance. Optimize for **not breaking production**.

### 4. Measure Twice, Cut Once

You get one shot in production. You can't undo easily. So:
- Test thoroughly
- Measure actual duration
- Understand failure modes
- Have a rollback plan
- Document your reasoning

### 5. Failure Recovery > Perfect Execution

Your migration will eventually fail or cause problems. Plan for it:
- How do you detect failure?
- How do you stop the damage?
- How do you recover?
- How do you prevent data loss?

## The Confidence Curve

```
Low Confidence → Medium Confidence → High Confidence
     ↓                    ↓                    ↓
 "Migrations       "I understand         "I can design
  are magic"        how they work"        production-safe
                                          migrations"
     ↓                    ↓                    ↓
 Copy/paste        Write basic           Consider locks,
 examples         migrations             compatibility,
                                         failure modes
     ↓                    ↓                    ↓
 Panic when        Understand            Calmly execute
 things break      what broke            complex migrations
```

**This repository's goal**: Move you from left to right.

You'll never eliminate risk entirely. But you can:
- Predict risk accurately
- Mitigate risk systematically
- Respond to failure calmly

## The Senior Engineer's Migration Checklist

Before running ANY migration in production, ask:

- [ ] What locks will this acquire?
- [ ] How long will it take? (tested on production-sized data)
- [ ] Can old app code handle the new schema?
- [ ] Can new app code handle the old schema?
- [ ] What happens if this migration fails partway through?
- [ ] Can I rollback? What does rollback mean here?
- [ ] Is there data loss risk?
- [ ] Have I tested this on a production snapshot?
- [ ] Does anyone else need to coordinate? (other teams, on-call, etc.)
- [ ] What will I monitor during execution?
- [ ] What's my "oh shit" plan?

If you can't answer these questions, **don't run the migration**.

## Final Thoughts

Migrations are not a chore. They're not "just database stuff." They're **production surgery**.

You're modifying the heart of your system while it's beating.

Treat it with respect. Think through the consequences. Test thoroughly. Deploy carefully.

The difference between a junior and senior engineer isn't the SQL they write. It's the questions they ask **before** writing the SQL.

Let's make you that senior engineer.

---

**Next**: [Migration Basics](./02_migration_basics.md) - The mechanics of how migrations work
