# Query Planner and Optimizer

## Introduction

You write SQL. The database decides *how* to execute it.

```sql
SELECT u.name, COUNT(o.id)
FROM users u
JOIN orders o ON u.id = o.user_id
WHERE u.created_at > '2024-01-01'
GROUP BY u.name;
```

This query could be executed in dozens of different ways:
- Scan `users` first or `orders` first?
- Use an index on `created_at` or scan the table?
- Hash join or nested loop join?
- Sort in memory or on disk?

The **query planner** (or **query optimizer**) decides. And its decisions determine whether your query takes 10 milliseconds or 10 seconds.

This chapter explains how the planner works, why it sometimes makes bad decisions, and how to guide it toward better ones.

---

## What the Query Planner Does

The planner has one job: **Find the fastest way to execute your query.**

### The Process

1. **Parse the SQL** → Build an abstract syntax tree (AST)
2. **Rewrite the query** → Apply transformations (e.g., remove redundant conditions)
3. **Generate candidate plans** → Enumerate possible execution strategies
4. **Estimate costs** → For each plan, estimate I/O, CPU, and memory costs
5. **Choose the cheapest plan** → Execute it

```
SQL Query
   ↓
Parser
   ↓
Rewriter
   ↓
Planner (generates plans)
   ↓
Cost Estimator (estimates costs)
   ↓
Best Plan
   ↓
Executor (runs the plan)
   ↓
Results
```

### Example

```sql
SELECT * FROM users WHERE age = 30;
```

**Candidate plans**:
1. **Sequential scan**: Scan the entire table, filter rows where `age = 30`
2. **Index scan**: Look up `age = 30` in the index, fetch matching rows

**Cost estimation**:
- Sequential scan: Read all pages (~1000 pages)
- Index scan: Read index (~5 pages) + fetch rows (~100 pages) = 105 pages

**Decision**: Index scan is cheaper → choose index scan.

---

## Cost-Based Optimization

The planner doesn't just guess. It estimates the **cost** of each plan.

### Cost Units

PostgreSQL uses arbitrary "cost units" (not real-world time):
- **seq_page_cost**: Cost to read one page sequentially (default: 1.0)
- **random_page_cost**: Cost to read one random page (default: 4.0)
- **cpu_tuple_cost**: Cost to process one row in memory (default: 0.01)
- **cpu_operator_cost**: Cost to execute a comparison operator (default: 0.0025)

**Total cost** = I/O cost + CPU cost.

### Example: Sequential Scan

Table: 1000 pages, 50,000 rows.

**Cost**:
- I/O: 1000 pages × `seq_page_cost` (1.0) = 1000
- CPU: 50,000 rows × `cpu_tuple_cost` (0.01) = 500
- **Total: 1500**

### Example: Index Scan

Index: 5 pages.
Matching rows: 1000 (on 100 different heap pages).

**Cost**:
- Index I/O: 5 pages × `random_page_cost` (4.0) = 20
- Heap I/O: 100 pages × `random_page_cost` (4.0) = 400
- CPU: 1000 rows × `cpu_tuple_cost` (0.01) = 10
- **Total: 430**

**Decision**: Index scan (430) is cheaper than sequential scan (1500).

---

## Statistics and Histograms

The planner relies on **statistics** to estimate costs.

### What Statistics Are Collected

For each column:
- **Number of distinct values** (`n_distinct`)
- **Most common values** (MCVs) and their frequencies
- **Histogram** (distribution of values)
- **Null fraction**

For each table:
- **Number of rows** (`reltuples`)
- **Number of pages** (`relpages`)

### How Statistics Are Updated

**PostgreSQL**:
```sql
ANALYZE users;
```

This samples rows from the table and updates `pg_stats`.

**When to run**:
- After bulk inserts
- After major updates
- If query plans seem wrong

**Automatic**: PostgreSQL's **autovacuum** runs `ANALYZE` automatically, but can fall behind under heavy load.

### Viewing Statistics

```sql
SELECT 
  tablename, 
  attname AS column, 
  n_distinct, 
  most_common_vals, 
  most_common_freqs
FROM pg_stats
WHERE tablename = 'users';
```

**Example output**:
```
tablename | column | n_distinct | most_common_vals   | most_common_freqs
----------|--------|------------|--------------------|-------------------
users     | age    | 50         | {25,30,35}         | {0.2,0.3,0.15}
```

**Interpretation**: 
- There are ~50 distinct ages
- 30% of users have `age = 30`
- 20% have `age = 25`

The planner uses this to estimate how many rows match `WHERE age = 30`.

---

## Why Bad Estimates Happen

### 1. **Stale Statistics**

You inserted 1 million rows, but `ANALYZE` hasn't run yet.

**Result**: The planner thinks the table has 1000 rows. It chooses a nested loop join (optimal for small tables) instead of a hash join (optimal for large tables).

**Fix**: Run `ANALYZE` manually.

### 2. **Correlated Columns**

```sql
SELECT * FROM orders 
WHERE country = 'USA' AND state = 'CA';
```

The planner assumes `country` and `state` are **independent**:
- `country = 'USA'`: 50% of rows
- `state = 'CA'`: 10% of rows
- **Estimate**: 50% × 10% = 5% of rows

**Reality**: If `country = 'USA'`, then `state = 'CA'` is 12% (not 10%). The estimate is wrong.

**Fix**: Create an index on `(country, state)` or use extended statistics:
```sql
CREATE STATISTICS stats_country_state (dependencies) ON country, state FROM orders;
ANALYZE orders;
```

### 3. **Skewed Data**

```sql
SELECT * FROM users WHERE email = 'alice@example.com';
```

The planner assumes all values are equally common.

**Reality**: Some emails appear once, some appear 1000 times (bots, test accounts).

**Result**: The planner underestimates the number of rows and chooses a slow plan.

**Fix**: Run `ANALYZE` to capture the most common values (MCVs).

### 4. **Complex Expressions**

```sql
SELECT * FROM users WHERE UPPER(email) = 'ALICE@EXAMPLE.COM';
```

The planner has no statistics for `UPPER(email)`.

**Result**: It guesses wildly (often assumes 0.5% of rows match).

**Fix**: Create an index on the expression:
```sql
CREATE INDEX idx_upper_email ON users(UPPER(email));
ANALYZE users;
```

Now the planner has statistics for `UPPER(email)`.

---

## When the Planner Chooses a Terrible Plan

### Example: Nested Loop Join on Large Tables

```sql
SELECT u.name, o.total
FROM users u
JOIN orders o ON u.id = o.user_id;
```

**Bad plan** (chosen if statistics are stale):
```
Nested Loop
  -> Seq Scan on users u
  -> Index Scan on orders o (user_id = u.id)
```

**Cost**: For each of 1 million users, look up orders in the index.
**Total**: 1 million index lookups × 5 pages each = 5 million page reads.

**Good plan**:
```
Hash Join
  -> Seq Scan on users u
  -> Seq Scan on orders o
```

**Cost**: Read both tables sequentially (~100,000 pages total).

**Why the bad plan was chosen**: The planner thought `users` had only 1000 rows (stale statistics), so nested loop seemed optimal.

**Fix**: `ANALYZE` both tables.

---

## How to Read EXPLAIN Output

`EXPLAIN` shows the query plan.

```sql
EXPLAIN SELECT * FROM users WHERE age > 30;
```

**Output**:
```
Seq Scan on users  (cost=0.00..1500.00 rows=25000 width=50)
  Filter: (age > 30)
```

### Key Fields

- **Seq Scan**: Operation type (sequential scan)
- **cost=0.00..1500.00**: 
  - `0.00` = startup cost (cost before first row)
  - `1500.00` = total cost
- **rows=25000**: Estimated number of rows returned
- **width=50**: Estimated average row size (bytes)

### Common Operations

| Operation             | Description                              |
|-----------------------|------------------------------------------|
| `Seq Scan`            | Full table scan                          |
| `Index Scan`          | Read index + fetch heap tuples           |
| `Index Only Scan`     | Read index only (no heap fetch)          |
| `Bitmap Index Scan`   | Build a bitmap of matching TIDs          |
| `Nested Loop`         | Nested loop join                         |
| `Hash Join`           | Hash join (build hash table, then probe)|
| `Merge Join`          | Merge join (both inputs sorted)          |
| `Sort`                | Sort rows                                |
| `Aggregate`           | GROUP BY or aggregation                  |

---

## EXPLAIN ANALYZE (Actual Execution)

`EXPLAIN` shows **estimates**. `EXPLAIN ANALYZE` shows **actual execution**.

```sql
EXPLAIN ANALYZE SELECT * FROM users WHERE age > 30;
```

**Output**:
```
Seq Scan on users  (cost=0.00..1500.00 rows=25000 width=50) 
                   (actual time=0.123..456.789 rows=30000 loops=1)
  Filter: (age > 30)
  Rows Removed by Filter: 20000
Planning Time: 1.234 ms
Execution Time: 500.000 ms
```

### Key Fields

- **actual time=0.123..456.789**: Actual time (in ms) to:
  - `0.123` = return first row
  - `456.789` = return all rows
- **actual rows=30000**: Actual number of rows returned
- **estimated rows=25000**: Planner's estimate

**Mismatch**: Estimated 25,000 rows, actually 30,000 rows. The estimate is off by 20%.

**If the mismatch is >10×**, the planner's statistics are stale or wrong.

---

## Nested Loop vs Hash Join vs Merge Join

### Nested Loop Join

**How it works**:
```
For each row in outer table:
  For each row in inner table:
    If join condition matches:
      Return joined row
```

**Cost**: O(outer_rows × inner_rows) (without index) or O(outer_rows × log(inner_rows)) (with index).

**Best for**:
- Small outer table
- Inner table has an index on the join key

**Worst for**:
- Large outer table
- No index on inner table

### Hash Join

**How it works**:
1. Build a hash table from the smaller table
2. Scan the larger table and probe the hash table

**Cost**: O(outer_rows + inner_rows).

**Best for**:
- Large tables
- No index on join key
- Equi-joins (`=`)

**Worst for**:
- Non-equi-joins (`<`, `>`)

### Merge Join

**How it works**:
1. Sort both tables by the join key
2. Merge the sorted lists

**Cost**: O(outer_rows × log(outer_rows) + inner_rows × log(inner_rows)).

**Best for**:
- Both tables already sorted (e.g., by an index)
- Large tables

**Worst for**:
- Unsorted tables (sorting is expensive)

---

## How ORMs Trigger Bad Plans

### Problem 1: N+1 Queries

**ORM code**:
```javascript
const users = await User.findAll();
for (const user of users) {
  const orders = await Order.findAll({ where: { userId: user.id } });
}
```

**Result**: 1 query for users + 1000 queries for orders = **1001 queries**.

**Each query has overhead** (parsing, planning, execution, network round-trip).

**Fix**: Use a join or batch fetch:
```javascript
const users = await User.findAll({ include: [Order] });
```

**Generated SQL**:
```sql
SELECT * FROM users u LEFT JOIN orders o ON u.id = o.user_id;
```

**Result**: 1 query.

### Problem 2: SELECT *

**ORM code**:
```javascript
const user = await User.findOne({ where: { id: 42 } });
```

**Generated SQL**:
```sql
SELECT * FROM users WHERE id = 42;
```

If `users` has 50 columns (including large text fields), this reads a lot of unnecessary data.

**Fix**: Select only needed columns:
```javascript
const user = await User.findOne({ where: { id: 42 }, attributes: ['id', 'name'] });
```

**Generated SQL**:
```sql
SELECT id, name FROM users WHERE id = 42;
```

### Problem 3: No Index Usage

**ORM code**:
```javascript
const users = await User.findAll({ where: { email: { [Op.like]: '%@example.com' } } });
```

**Generated SQL**:
```sql
SELECT * FROM users WHERE email LIKE '%@example.com';
```

**Problem**: `LIKE '%...'` can't use a B-Tree index (no left prefix).

**Result**: Full table scan.

**Fix**: Redesign the query or use a specialized index (e.g., trigram index in PostgreSQL).

---

## Hints and Plan Forcing (Use Sparingly)

Sometimes the planner is wrong and you want to override it.

### PostgreSQL: limited hint support

PostgreSQL has no native hints, but you can use `pg_hint_plan` extension:

```sql
/*+ SeqScan(users) */
SELECT * FROM users WHERE age > 30;
```

**Warning**: Hints are brittle. If your data grows, the hint may become suboptimal.

**Better**: Fix the root cause (update statistics, tune configuration).

### MySQL: Hints

MySQL supports hints natively:

```sql
SELECT /*+ INDEX(users idx_age) */ * FROM users WHERE age > 30;
```

### Oracle: Extensive hint support

Oracle has the most powerful hint system:

```sql
SELECT /*+ INDEX(users idx_age) */ * FROM users WHERE age > 30;
SELECT /*+ FULL(users) */ * FROM users WHERE age > 30;  -- Force full scan
SELECT /*+ USE_HASH(users orders) */ ...  -- Force hash join
```

---

## Query Rewriting Tricks

Sometimes you can rewrite a query to get a better plan.

### 1. **Replace OR with UNION**

**Slow**:
```sql
SELECT * FROM users WHERE age = 25 OR age = 30;
```

The planner might not use an index efficiently.

**Fast**:
```sql
SELECT * FROM users WHERE age = 25
UNION
SELECT * FROM users WHERE age = 30;
```

Each branch can use the index.

### 2. **Replace DISTINCT with GROUP BY**

**Slow** (sometimes):
```sql
SELECT DISTINCT user_id FROM orders;
```

**Fast**:
```sql
SELECT user_id FROM orders GROUP BY user_id;
```

`GROUP BY` can use a hash aggregate; `DISTINCT` sometimes forces a sort.

### 3. **Materialize Subqueries**

**Slow**:
```sql
SELECT * FROM users 
WHERE id IN (SELECT user_id FROM orders WHERE total > 1000);
```

If the subquery is complex, the planner might execute it repeatedly.

**Fast**:
```sql
WITH high_value_users AS (
  SELECT DISTINCT user_id FROM orders WHERE total > 1000
)
SELECT * FROM users 
WHERE id IN (SELECT user_id FROM high_value_users);
```

Or use a `JOIN`:
```sql
SELECT DISTINCT u.* 
FROM users u
JOIN orders o ON u.id = o.user_id
WHERE o.total > 1000;
```

---

## Configuration Tuning

The planner's behavior is controlled by configuration parameters.

### PostgreSQL: Key Parameters

**random_page_cost** (default: 4.0)
- Cost of a random page read vs sequential read
- On SSDs, set to `1.1` (random I/O is only slightly slower)

**effective_cache_size** (default: 4 GB)
- How much RAM is available for caching (shared_buffers + OS cache)
- Set to 50-75% of total RAM
- Helps the planner decide whether data fits in cache

**work_mem** (default: 4 MB)
- Memory per sort/hash operation
- If too low, sorts spill to disk (slow)
- If too high, risk of OOM
- Set to 64 MB - 256 MB for OLTP workloads

**Example**:
```sql
SET random_page_cost = 1.1;
SET effective_cache_size = '16GB';
SET work_mem = '64MB';
```

### MySQL: Key Parameters

**innodb_buffer_pool_size** (default: 128 MB)
- Set to 70-80% of RAM

**join_buffer_size** (default: 256 KB)
- Memory for joins without indexes
- Increase if you do large joins

---

## Monitoring Slow Queries

### PostgreSQL: pg_stat_statements

Enable the extension:
```sql
CREATE EXTENSION pg_stat_statements;
```

View slow queries:
```sql
SELECT 
  query, 
  calls, 
  total_exec_time, 
  mean_exec_time, 
  max_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 20;
```

**Look for**:
- High `mean_exec_time` → optimize this query
- High `calls` × `mean_exec_time` → optimize this query (big impact)

### MySQL: Slow Query Log

Enable logging:
```sql
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 1;  -- Log queries > 1 second
```

View log:
```bash
cat /var/log/mysql/slow-query.log
```

---

## Practical Takeaways

### 1. **Always Run EXPLAIN ANALYZE**

Before deploying a query to production:
```sql
EXPLAIN ANALYZE SELECT ...;
```

Check:
- Are indexes being used?
- Are estimated rows close to actual rows?
- Are there expensive operations (sorts, nested loops on large tables)?

### 2. **Keep Statistics Fresh**

Run `ANALYZE` after bulk changes:
```sql
ANALYZE;
```

Or tune autovacuum to run more frequently.

### 3. **Watch for Mismatches**

If `estimated rows` differs from `actual rows` by >10×, your statistics are wrong.

**Fix**:
- Run `ANALYZE`
- Increase `default_statistics_target` (PostgreSQL)

### 4. **Use Composite Indexes**

For queries like:
```sql
WHERE user_id = 42 AND status = 'active'
```

Create:
```sql
CREATE INDEX idx_user_id_status ON orders(user_id, status);
```

### 5. **Avoid SELECT ***

Only select columns you need. This:
- Reduces data transfer
- Enables index-only scans
- Reduces cache pollution

### 6. **Monitor and Iterate**

Use `pg_stat_statements` or slow query log to find slow queries.

Optimize the top 10 slowest queries—that's usually 80% of the performance gain.

---

## Summary

- The query planner chooses how to execute your query
- It uses cost-based optimization (estimates I/O + CPU costs)
- It relies on statistics (run `ANALYZE` regularly!)
- Bad plans usually result from stale statistics or correlated columns
- Use `EXPLAIN ANALYZE` to debug slow queries
- ORMs can generate inefficient queries (N+1, SELECT *)
- Tune `random_page_cost`, `effective_cache_size`, and `work_mem`
- Monitor slow queries with `pg_stat_statements` or slow query log

Understanding the query planner lets you predict performance and debug slow queries confidently.
