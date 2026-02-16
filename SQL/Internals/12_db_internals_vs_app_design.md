# Database Internals vs Application Design

## Introduction

You've learned how databases work internally:
- Pages, tuples, and storage engines
- Indexes and B-Trees
- Buffer cache and I/O
- Query planner and optimization
- MVCC and visibility
- Transactions and isolation
- Locks and blocking
- VACUUM and maintenance
- Replication and scaling
- Failures and recovery

**Now the crucial question: How does this knowledge change the way you design applications?**

This chapter connects database internals to practical application design decisions.

---

## How Internals Influence Schema Design

### 1. **Normalize to Avoid Page Bloat**

**Problem**: Wide rows mean fewer rows per page. Scanning N rows requires reading more pages.

```sql
-- Bad: Wide table
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  name TEXT,
  email TEXT,
  bio TEXT,  -- 10 KB on average
  preferences JSONB,  -- 5 KB on average
  last_login TIMESTAMP
);
```

**Each row is ~15 KB. Each 8 KB page fits 0-1 rows.**

**Query cost**:
```sql
SELECT id, name FROM users WHERE last_login > NOW() - INTERVAL '1 day';
-- Must read 1 page per row (even though we only need 2 small columns)
```

**Fix**: Vertical partitioning (split wide columns into a separate table):

```sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  name TEXT,
  email TEXT,
  last_login TIMESTAMP
);

CREATE TABLE user_profiles (
  user_id INT PRIMARY KEY REFERENCES users(id),
  bio TEXT,
  preferences JSONB
);
```

**Each row in `users` is ~100 bytes. Each 8 KB page fits ~80 rows.**

**Query cost**:
```sql
SELECT id, name FROM users WHERE last_login > NOW() - INTERVAL '1 day';
-- Reads 1 page per 80 rows (80× fewer pages!)
```

### 2. **Avoid Indexing Frequently-Updated Columns**

**Problem**: Every `UPDATE` to an indexed column creates a new index entry. Hot columns cause index bloat and prevent HOT updates.

```sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  name TEXT,
  email TEXT,
  last_login TIMESTAMP  -- Updated on every login
);

-- Bad: Index on frequently-updated column
CREATE INDEX idx_last_login ON users(last_login);
```

**Impact**:
- Every login updates the row (creates new tuple)
- Index must be updated (creates new index entry)
- Old index entries become dead tuples
- Index bloats over time

**Fix**: Store `last_login` in a separate table:

```sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  name TEXT,
  email TEXT
);

CREATE TABLE user_sessions (
  user_id INT REFERENCES users(id),
  last_login TIMESTAMP
);

-- Update last_login without touching users table
UPDATE user_sessions SET last_login = NOW() WHERE user_id = 1;
```

**Or**: Don't index `last_login` (use other columns for filtering).

### 3. **Use Sequential Primary Keys**

**Problem**: Random primary keys (UUIDs) cause page splits in indexes, leading to fragmentation.

```sql
-- Bad: Random UUIDs
CREATE TABLE orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id INT,
  total DECIMAL
);
```

**Impact** (on MySQL InnoDB):
- Clustered index (the table) is ordered by primary key
- Random inserts cause page splits throughout the tree
- Fragmentation, slower inserts

**Impact** (on PostgreSQL):
- Primary key index becomes fragmented
- Slower lookups

**Fix**: Use sequential IDs:

```sql
CREATE TABLE orders (
  id SERIAL PRIMARY KEY,  -- Sequential
  user_id INT,
  total DECIMAL
);
```

**Or**: Use ULIDs (time-ordered UUIDs):
```sql
CREATE TABLE orders (
  id UUID PRIMARY KEY DEFAULT generate_ulid(),  -- Time-ordered UUID
  user_id INT,
  total DECIMAL
);
```

### 4. **Partition Large Tables**

**Problem**: Vacuuming a 1 TB table takes hours. Querying it is slow.

**Fix**: Partition by time or range:

```sql
CREATE TABLE logs (
  id BIGSERIAL,
  created_at TIMESTAMP,
  message TEXT
) PARTITION BY RANGE (created_at);

CREATE TABLE logs_2024_01 PARTITION OF logs FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE logs_2024_02 PARTITION OF logs FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
...
```

**Benefits**:
- Queries filter by `created_at` only scan relevant partitions (partition pruning)
- VACUUM runs on smaller partitions (faster)
- Old partitions can be dropped instantly (no VACUUM needed)

---

## How Internals Influence API Design

### 1. **Batch Reads to Reduce Round-Trips**

**Problem**: N+1 queries.

```javascript
// Bad: N+1 queries
const users = await db.query('SELECT id, name FROM users');
for (const user of users) {
  const orders = await db.query('SELECT * FROM orders WHERE user_id = $1', [user.id]);
}
```

**Cost**: 1 + N database round-trips (slow if N is large).

**Fix**: Batch with a JOIN or IN clause:

```javascript
// Good: 1 query
const result = await db.query(`
  SELECT u.id, u.name, o.id AS order_id, o.total
  FROM users u
  LEFT JOIN orders o ON u.id = o.user_id
`);
```

**Or use DataLoader** (batching library):

```javascript
const userLoader = new DataLoader(async (userIds) => {
  const orders = await db.query('SELECT * FROM orders WHERE user_id = ANY($1)', [userIds]);
  return orders;
});

for (const user of users) {
  const orders = await userLoader.load(user.id);  // Batched automatically
}
```

### 2. **Paginate Large Result Sets**

**Problem**: Returning 1 million rows.

```javascript
// Bad: Returns all rows
const users = await db.query('SELECT * FROM users');
// 1 million rows × 100 bytes = 100 MB of data
```

**Cost**:
- Database serializes 1 million rows
- Network transfer (slow)
- Application deserializes 1 million rows (slow, high memory)

**Fix**: Paginate:

```javascript
// Good: Returns 100 rows at a time
const users = await db.query('SELECT * FROM users ORDER BY id LIMIT 100 OFFSET 0');
```

**Better**: Use cursor-based pagination (no OFFSET overhead):

```javascript
const users = await db.query('SELECT * FROM users WHERE id > $1 ORDER BY id LIMIT 100', [lastSeenId]);
```

### 3. **Cache Expensive Queries**

**Problem**: Running an expensive aggregation on every request.

```javascript
// Bad: Runs on every request
app.get('/stats', async (req, res) => {
  const stats = await db.query('SELECT COUNT(*) AS total FROM users');
  res.json(stats);
});
```

**Cost**: If `users` has 10 million rows, `COUNT(*)` takes ~1 second (must scan entire table).

**Fix**: Cache the result:

```javascript
// Good: Cache for 5 minutes
const cache = new NodeCache({ stdTTL: 300 });

app.get('/stats', async (req, res) => {
  let stats = cache.get('user_count');
  if (!stats) {
    stats = await db.query('SELECT COUNT(*) AS total FROM users');
    cache.set('user_count', stats);
  }
  res.json(stats);
});
```

**Or**: Use a materialized view:

```sql
CREATE MATERIALIZED VIEW user_stats AS
SELECT COUNT(*) AS total FROM users;

-- Refresh periodically
REFRESH MATERIALIZED VIEW user_stats;
```

---

## Batch vs Chatty Queries

### Chatty (Bad)

```javascript
for (let i = 0; i < 1000; i++) {
  await db.query('INSERT INTO logs (message) VALUES ($1)', [`Message ${i}`]);
}
```

**Cost**: 1000 round-trips. If each round-trip is 1ms, total time = 1 second.

### Batched (Good)

```javascript
const values = [];
for (let i = 0; i < 1000; i++) {
  values.push(`('Message ${i}')`);
}
await db.query(`INSERT INTO logs (message) VALUES ${values.join(',')}`);
```

**Cost**: 1 round-trip. Total time = 10ms.

**Or use multi-row INSERT**:

```javascript
await db.query('INSERT INTO logs (message) VALUES ' + Array(1000).fill('(?)').join(','), messages);
```

**Or use COPY** (fastest):

```javascript
const stream = db.query(copyFrom('COPY logs (message) FROM STDIN'));
for (const message of messages) {
  stream.write(message + '\n');
}
stream.end();
```

---

## Long Transactions and Why They're Evil

### Problem

```javascript
// Bad: Long transaction
await db.query('BEGIN');
const user = await db.query('SELECT * FROM users WHERE id = 1');
// Do some slow computation or API call (10 seconds)
const result = await externalAPI.call(user);
await db.query('UPDATE users SET data = $1 WHERE id = 1', [result]);
await db.query('COMMIT');
```

**Impact**:
- Transaction holds locks for 10 seconds (blocks other transactions)
- Prevents VACUUM from cleaning dead tuples (causes bloat)
- Increases risk of deadlocks

### Fix

**Move computation outside the transaction**:

```javascript
// Good: Short transaction
const user = await db.query('SELECT * FROM users WHERE id = 1');
// Do computation outside transaction
const result = await externalAPI.call(user);
// Short transaction for write
await db.query('BEGIN');
await db.query('UPDATE users SET data = $1 WHERE id = 1', [result]);
await db.query('COMMIT');
```

**Rule**: Keep transactions **short** (<100ms if possible).

---

## Why "Just Cache It" Often Backfires

**Scenario**: Your app is slow. Someone suggests: "Just cache everything in Redis!"

### Problems

**1. Cache Invalidation**

When do you invalidate the cache?

```javascript
// Write to database
await db.query('UPDATE users SET name = $1 WHERE id = $2', ['Alice', 1]);
// Invalidate cache
cache.del('user:1');

// But what if another service also updates users?
// Your cache is now stale.
```

**Cache invalidation is hard.** You must:
- Invalidate on every write
- Handle race conditions (write and invalidate are not atomic)
- Invalidate in all services that cache the data

**Alternative**: Use short TTLs (e.g., 1 minute) and accept some staleness.

**2. Cache Stampede**

When the cache expires, many requests simultaneously try to recompute it:

```javascript
// 1000 requests hit at the same time
const data = cache.get('expensive_query');
if (!data) {
  // All 1000 requests run this query simultaneously!
  data = await db.query('SELECT COUNT(*) FROM users');
  cache.set('expensive_query', data);
}
```

**Result**: 1000 concurrent queries (overwhelms the database).

**Fix**: Use a lock or "cache warming":

```javascript
const lock = await acquireLock('expensive_query_lock');
if (lock) {
  const data = await db.query('SELECT COUNT(*) FROM users');
  cache.set('expensive_query', data);
  releaseLock('expensive_query_lock');
} else {
  // Wait for another request to populate the cache
  await sleep(100);
}
```

**3. Memory Overhead**

Caching everything in Redis consumes RAM. If your dataset is 100 GB and you cache it all, you need 100 GB of Redis RAM.

**Alternative**: Cache only hot data (the 10% accessed frequently).

**Rule**: Cache **selectively**. Don't cache everything reflexively.

---

## The N+1 Query Problem (ORMs)

ORMs generate inefficient queries if you're not careful.

### Example (Sequelize)

```javascript
// Fetch users
const users = await User.findAll();

// Fetch orders for each user
for (const user of users) {
  const orders = await user.getOrders();  // N+1 queries!
}
```

**Generated SQL**:
```sql
SELECT * FROM users;
SELECT * FROM orders WHERE user_id = 1;
SELECT * FROM orders WHERE user_id = 2;
...
SELECT * FROM orders WHERE user_id = 1000;
```

**Cost**: 1 + 1000 queries.

### Fix: Eager Loading

```javascript
const users = await User.findAll({
  include: [{ model: Order }]
});
```

**Generated SQL**:
```sql
SELECT * FROM users;
SELECT * FROM orders WHERE user_id IN (1, 2, 3, ..., 1000);
```

**Cost**: 2 queries (much better).

---

## Database-Driven Design Principles

### 1. **Don't Treat the Database as a Dumb Store**

**Bad**:
```javascript
const users = await db.query('SELECT * FROM users');
const active = users.filter(u => u.status === 'active');
```

**Why bad**: Fetches all users, filters in application (slow, high memory).

**Good**:
```javascript
const active = await db.query('SELECT * FROM users WHERE status = $1', ['active']);
```

**Rule**: Push filtering, sorting, and aggregation to the database (it's optimized for this).

### 2. **Leverage Database Constraints**

**Bad**: Enforce uniqueness in application code:
```javascript
const existing = await db.query('SELECT id FROM users WHERE email = $1', [email]);
if (existing.length > 0) {
  throw new Error('Email already exists');
}
await db.query('INSERT INTO users (email) VALUES ($1)', [email]);
```

**Problem**: Race condition. Two concurrent requests can both pass the check and insert duplicate emails.

**Good**: Use a unique constraint:
```sql
CREATE UNIQUE INDEX idx_users_email ON users(email);
```

**Database enforces uniqueness atomically.** No race conditions.

### 3. **Use Transactions for Multi-Step Operations**

**Bad**: No transaction (race condition):
```javascript
const balance = await db.query('SELECT balance FROM accounts WHERE id = $1', [1]);
await db.query('UPDATE accounts SET balance = $1 WHERE id = $2', [balance - 100, 1]);
```

**Problem**: Another transaction can modify the balance between the SELECT and UPDATE (lost update).

**Good**: Use a transaction:
```javascript
await db.query('BEGIN');
await db.query('UPDATE accounts SET balance = balance - 100 WHERE id = $1', [1]);
await db.query('COMMIT');
```

**Or use SELECT FOR UPDATE**:
```javascript
await db.query('BEGIN');
const balance = await db.query('SELECT balance FROM accounts WHERE id = $1 FOR UPDATE', [1]);
await db.query('UPDATE accounts SET balance = $1 WHERE id = $2', [balance - 100, 1]);
await db.query('COMMIT');
```

### 4. **Design for the 80/20 Rule**

**80% of queries access 20% of data** (hot data).

**Implication**:
- Ensure hot data fits in buffer cache (tune `shared_buffers`)
- Index columns used in hot queries
- Cache hot data in Redis (but not all data!)

### 5. **Avoid Premature Optimization**

**Premature optimization**: Adding 10 indexes, partitioning tables, caching everything—before measuring.

**Better**:
1. Deploy with minimal indexes (primary key, foreign keys)
2. Measure slow queries in production (`pg_stat_statements`)
3. Add indexes for the top 10 slowest queries
4. Repeat

**Rule**: Optimize based on real data, not guesses.

---

## Debugging Production Issues

### Symptom 1: Queries Suddenly Slow

**Possible causes**:
1. **Stale statistics**: Run `ANALYZE`
2. **Cache cold**: Restart warmed the cache; wait for it to warm up
3. **Lock contention**: Check `pg_locks`
4. **Table bloat**: Check `pg_stat_user_tables`

### Symptom 2: Disk Full

**Possible causes**:
1. **WAL growth**: Check WAL size; increase checkpoint frequency
2. **Table bloat**: Run `VACUUM`
3. **Logs**: Old log files filling disk

### Symptom 3: High CPU

**Possible causes**:
1. **Expensive queries**: Check `pg_stat_statements` for high `total_exec_time`
2. **Too many connections**: Use connection pooling
3. **VACUUM overhead**: Tune autovacuum

### Symptom 4: Data Loss After Crash

**Possible causes**:
1. **fsync = off**: Check `postgresql.conf`
2. **Disk failure**: Check hardware
3. **No backups**: Implement WAL archiving

---

## Final Checklist for Production

### Schema Design
- [ ] Use sequential primary keys (avoid UUIDs)
- [ ] Normalize wide tables (split large columns)
- [ ] Avoid indexing frequently-updated columns
- [ ] Partition very large tables (>100 GB)

### Query Patterns
- [ ] Batch reads (avoid N+1 queries)
- [ ] Paginate large result sets
- [ ] Push filtering/sorting to database
- [ ] Use indexes for frequent filters

### Transactions
- [ ] Keep transactions short (<100ms)
- [ ] Use `SELECT FOR UPDATE` to prevent lost updates
- [ ] Handle serialization failures (retry)

### Maintenance
- [ ] Tune autovacuum for hot tables
- [ ] Monitor dead tuples (`pg_stat_user_tables`)
- [ ] Kill long-running idle transactions
- [ ] Monitor disk space (alert at 80%)

### Durability
- [ ] Enable `fsync` (never disable in production)
- [ ] Take daily backups (test restores!)
- [ ] Set up WAL archiving (PITR)
- [ ] Use replicas for high availability

### Monitoring
- [ ] Monitor slow queries (`pg_stat_statements`)
- [ ] Monitor replication lag (`pg_stat_replication`)
- [ ] Monitor cache hit ratio (>99%)
- [ ] Set up alerts for critical metrics

---

## Summary

Understanding database internals influences:
- **Schema design** (avoid wide rows, use sequential IDs, partition large tables)
- **API design** (batch queries, paginate, cache selectively)
- **Transaction patterns** (keep short, use locks correctly)
- **Query optimization** (push logic to DB, use indexes)
- **Operational practices** (monitor, backup, tune autovacuum)

**The goal**: Design systems that work **with** the database, not against it.

By understanding how the database works internally, you can:
- Predict performance before deploying
- Debug production issues confidently
- Design schemas that scale
- Write queries that are fast by default
- Avoid common pitfalls (bloat, N+1 queries, long transactions)

**You're no longer treating the database as a black box.** You understand what happens under the hood—and that makes you a better engineer.
