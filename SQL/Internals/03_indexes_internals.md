# Index Internals

## Introduction

Indexes are the difference between a query that takes 0.1ms and one that takes 10 seconds.

But indexes aren't magic—they're data structures with specific characteristics, tradeoffs, and failure modes. Understanding how they work lets you:
- Predict which queries will be fast
- Design better schemas
- Debug slow queries without trial and error

This chapter demystifies indexes from a practical, application-developer perspective.

---

## What Is an Index?

An index is a **sorted copy** of some table columns, stored separately, with pointers back to the actual rows.

Think of it like a book's index:
- The book itself is the **table** (unordered pages)
- The index at the back is the **B-Tree index** (sorted topic names with page numbers)

```
Table (Heap):
┌────────────────────────────┐
│ id | name    | age | city  │
├────────────────────────────┤
│ 3  | Charlie | 35  | NYC   │
│ 1  | Alice   | 30  | SF    │
│ 2  | Bob     | 25  | LA    │
└────────────────────────────┘
        Unordered

Index on name:
┌─────────────────┐
│ Alice    → (1)  │  ← Points to row id=1
│ Bob      → (2)  │
│ Charlie  → (3)  │
└─────────────────┘
        Sorted
```

When you query:
```sql
SELECT * FROM users WHERE name = 'Bob';
```

**Without index**: Scan the entire table (sequential scan).
**With index**: Look up "Bob" in the index (binary search), get row pointer, fetch the row.

---

## B-Tree Structure (The Only Index Type That Matters)

99% of indexes are **B-Trees** (B-Tree or B+Tree, depending on the database).

### What Is a B-Tree?

A B-Tree is a self-balancing tree data structure where:
- Each node contains multiple keys (not just one, like a binary tree)
- Nodes can have many children (not just two)
- All leaf nodes are at the same depth (perfectly balanced)

```
                 ┌──────────────┐
                 │  [50, 100]   │  ← Root (Level 1)
                 └──────┬───────┘
          ┌─────────────┼─────────────┐
          │             │             │
     ┌────▼────┐   ┌────▼────┐   ┌───▼─────┐
     │ [10, 25]│   │ [60, 75]│   │[110,150]│ ← Internal (Level 2)
     └────┬────┘   └────┬────┘   └────┬────┘
    ┌─────┼─────┐  ┌────┼────┐   ┌────┼─────┐
    ▼     ▼     ▼  ▼    ▼    ▼   ▼    ▼     ▼
  [1-9][10-24][25-49][50-59][60-74]...   ← Leaves (Level 3)
   │     │     │     │      │
  TIDs TIDs  TIDs  TIDs   TIDs  ← Pointers to table rows
```

**Key properties**:
- **Shallow**: A B-Tree with 1 billion entries typically has depth ~3-4
- **Sorted**: Leaf nodes are in sorted order (enables range scans)
- **Self-balancing**: Insertions and deletions maintain balance automatically

### Why B-Trees?

**1. Logarithmic lookups**

Finding a key requires O(log N) comparisons:
- 1,000 rows → 3 levels
- 1,000,000 rows → 4 levels
- 1,000,000,000 rows → 5 levels

Each level = one page read. **5 page reads to find any row in a billion-row table.**

**2. Range scans are efficient**

```sql
SELECT * FROM users WHERE age BETWEEN 25 AND 35;
```

The B-Tree:
1. Searches to find `age = 25` (log N)
2. Scans leaf nodes sequentially until `age > 35`

The leaf nodes are **linked**, so scanning is fast.

**3. Multiple columns supported**

```sql
CREATE INDEX idx_age_city ON users(age, city);
```

The B-Tree is sorted first by `age`, then by `city`:
```
(25, "LA")
(25, "NYC")
(25, "SF")
(30, "LA")
(30, "NYC")
```

This enables:
- `WHERE age = 30` (first column)
- `WHERE age = 30 AND city = 'NYC'` (both columns)

But **not** efficiently:
- `WHERE city = 'NYC'` (second column only—must scan entire index)

**Rule**: Composite indexes only help if the query uses a **left prefix** of the columns.

---

## Why B-Trees Are Shallow

This is the most important property of B-Trees.

### Page-Based Storage

A B-Tree node is stored in a **database page** (8 KB in PostgreSQL).

Each page can hold hundreds of keys:
```
8 KB page ÷ ~20 bytes per entry = ~400 entries per page
```

**Fanout**: Each internal node has ~400 children.

**Tree depth**:
- 1 level: 400 rows
- 2 levels: 400 × 400 = 160,000 rows
- 3 levels: 400 × 400 × 400 = 64 million rows
- 4 levels: ~25 billion rows

**Implication**: Even massive tables have B-Trees with depth ≤ 5.

**I/O cost**: Finding a row requires at most 5 page reads. If those pages are cached, that's microseconds.

---

## How Range Scans Work

```sql
SELECT * FROM orders 
WHERE created_at BETWEEN '2024-01-01' AND '2024-01-31';
```

Assume there's an index on `created_at`.

**Steps**:
1. **Search** for `'2024-01-01'` in the B-Tree (3-4 page reads)
2. **Scan** leaf nodes sequentially, collecting TIDs
3. **Fetch** rows from the heap using TIDs

```
B-Tree leaf nodes:
[...][2024-01-01→TID1, 2024-01-02→TID2, ...][2024-01-15→TID50, ...][2024-02-01, ...]
           ▲─────────── Read these ─────────▲  Stop here
```

**Cost**:
- B-Tree traversal: 4 page reads
- Leaf scan: ~10 pages (if date range is 1 month out of 12)
- Heap fetches: ~1,000 page reads (one per matching row)

**Total**: ~1,014 page reads.

Compare to a **sequential scan** of the entire table:
- 1 million rows, 100 bytes each = 100 MB
- 100 MB ÷ 8 KB per page = ~12,500 pages

**Index is 12× faster** (1,014 vs 12,500 pages).

---

## Why Random Inserts Hurt

When you insert a row:
```sql
INSERT INTO users (id, name, age) VALUES (42, 'Dave', 28);
```

If `id` is indexed, the database must:
1. Find the correct position in the B-Tree for `id = 42`
2. Insert the key into the leaf node

If the leaf node is full, it must **split**:

```
Before:
┌───────────────────────────────────┐
│ 10, 20, 30, 40, ..., 80, 90, 100 │  ← Full leaf node
└───────────────────────────────────┘

Insert 42:
┌─────────────────────┐  ┌───────────────────────┐
│ 10, 20, 30, 40, 42  │  │ ..., 80, 90, 100      │
└─────────────────────┘  └───────────────────────┘
       New node                Old node (split)
```

**Cost**:
- Allocate a new page
- Update parent node to point to both pages
- Write both pages to disk

**If inserts are random** (e.g., `id = UUID()`), every page eventually splits. The B-Tree becomes fragmented.

**If inserts are sequential** (e.g., `id = AUTO_INCREMENT`), splits only happen at the end. No fragmentation.

### Sequential vs Random Inserts (Benchmark)

PostgreSQL, 1 million rows:

| Primary Key Type     | Insert Time | Index Size | Fragmentation |
|----------------------|-------------|------------|---------------|
| SERIAL (sequential)  | 10 seconds  | 25 MB      | Low           |
| UUID (random)        | 30 seconds  | 50 MB      | High          |

**Why?**
- Sequential inserts append to the last page (fast, compact)
- Random inserts cause page splits throughout the tree (slow, bloated)

**Rule of thumb**: Use sequential primary keys (`SERIAL`, `BIGSERIAL`, `AUTO_INCREMENT`) unless you have a specific reason not to.

---

## Index-Only Scans (Covering Indexes)

Normally, querying with an index requires:
1. Scan the index
2. Fetch rows from the heap

But if the index contains **all columns** needed by the query, the database can skip step 2.

### Example

```sql
CREATE INDEX idx_age ON users(age);
SELECT age FROM users WHERE age > 30;
```

The index contains `age` and row pointers. **The query only needs `age`.**

The database can scan the index and return values **without touching the table**.

**I/O cost**:
- Index scan: ~10 pages
- Heap fetch: **0 pages** (skipped!)

This is called an **index-only scan** (PostgreSQL) or **covering index** (MySQL).

### Composite Covering Index

```sql
CREATE INDEX idx_age_city ON users(age, city);
SELECT age, city FROM users WHERE age > 30;
```

The index contains both `age` and `city`. The query doesn't need any other columns.

**Result**: Index-only scan.

### When Index-Only Scans Fail (PostgreSQL-Specific)

PostgreSQL has a visibility problem: the index doesn't know if a tuple is visible to your transaction.

So even with a covering index, PostgreSQL must check the **visibility map** or visit the heap to verify visibility.

**Workaround**: If the table is frequently vacuumed, the visibility map marks pages as "all tuples visible," allowing true index-only scans.

(MySQL doesn't have this issue—it checks visibility via the index itself.)

---

## Why Too Many Indexes Kill Write Performance

Every index must be updated when you insert, update, or delete rows.

### Example

Table with 5 indexes:
```sql
CREATE INDEX idx_name ON users(name);
CREATE INDEX idx_email ON users(email);
CREATE INDEX idx_age ON users(age);
CREATE INDEX idx_city ON users(city);
CREATE INDEX idx_created_at ON users(created_at);
```

When you insert a row:
```sql
INSERT INTO users (name, email, age, city, created_at) VALUES (...);
```

The database must:
1. Insert the row into the table
2. Insert into `idx_name`
3. Insert into `idx_email`
4. Insert into `idx_age`
5. Insert into `idx_city`
6. Insert into `idx_created_at`

**An insert becomes 6 write operations.**

**Benchmark** (PostgreSQL, 1 million rows):
| Number of Indexes | Insert Time |
|-------------------|-------------|
| 0                 | 8 seconds   |
| 1                 | 10 seconds  |
| 3                 | 15 seconds  |
| 5                 | 25 seconds  |
| 10                | 50 seconds  |

**Rule**: Only create indexes you **actually use in queries**. Don't "pre-optimize" by indexing everything.

---

## Partial Indexes (Filtered Indexes)

You can index a **subset** of rows:

```sql
CREATE INDEX idx_active_users ON users(email) WHERE active = true;
```

This index only contains rows where `active = true`.

**Advantages**:
- Smaller index (faster scans, less disk space)
- Faster writes (inactive users don't update the index)

**Use case**:
```sql
SELECT * FROM users WHERE email = 'alice@example.com' AND active = true;
```

The partial index is perfect for this query.

---

## Unique Indexes

```sql
CREATE UNIQUE INDEX idx_email ON users(email);
```

The database enforces that no two rows have the same `email`.

**Implementation**: Before inserting, the database checks the index for duplicates. If found, the insert fails.

**Performance**: Unique indexes are slightly slower to insert (due to the duplicate check), but the overhead is minimal.

---

## Multi-Column Index Order Matters

```sql
CREATE INDEX idx_age_city ON users(age, city);
```

This index is sorted by `(age, city)` in that order.

**Queries that use the index**:
```sql
WHERE age = 30;                   -- ✓ Uses index (first column)
WHERE age = 30 AND city = 'NYC';  -- ✓ Uses index (both columns)
WHERE age BETWEEN 25 AND 35;      -- ✓ Uses index (first column range)
```

**Queries that DON'T use the index efficiently**:
```sql
WHERE city = 'NYC';                           -- ✗ Full index scan (second column only)
WHERE city = 'NYC' AND age = 30;              -- ✗ Full index scan (wrong order)
WHERE age > 25 AND city = 'NYC';              -- ⚠ Partial use (range on first column)
```

**Rule**: Put the most selective column first, or the column used in equality filters first.

### Best Practice: Analyze Query Patterns

Before creating a composite index, look at your actual queries:

```sql
-- Queries you run:
SELECT * FROM users WHERE age = 30 AND city = 'NYC';
SELECT * FROM users WHERE age BETWEEN 25 AND 35;
SELECT * FROM users WHERE city = 'NYC';  -- Rare query
```

**Best index**: `idx_age_city` (because most queries filter by `age` first).

Don't create `idx_city_age` unless you frequently query by `city` alone.

---

## Expression Indexes

You can index the result of an expression:

```sql
CREATE INDEX idx_lower_email ON users(LOWER(email));
```

Now this query uses the index:
```sql
SELECT * FROM users WHERE LOWER(email) = 'alice@example.com';
```

**Without the expression index**, this query would do a full table scan.

**Use case**: Case-insensitive searches.

---

## When the Database Doesn't Use Your Index

You've created an index. But `EXPLAIN` shows a sequential scan. Why?

### 1. **The Query Returns Too Many Rows**

```sql
SELECT * FROM users WHERE age > 10;  -- Returns 99% of rows
```

**Index cost**: 
- Scan index: 1,000 pages
- Fetch rows from heap: 990,000 pages (random I/O)

**Sequential scan cost**:
- Scan table: 100,000 pages (sequential I/O)

**The sequential scan is faster!**

The query planner estimates costs and chooses the sequential scan.

### 2. **Statistics Are Stale**

The database maintains statistics about column distributions:
- How many distinct values?
- What's the most common value?
- Are values correlated?

If you've inserted 1 million rows since the last `ANALYZE`, the planner's estimates are wrong.

**Fix**:
```sql
ANALYZE users;
```

### 3. **Wrong Index Column Order**

```sql
CREATE INDEX idx_city_age ON users(city, age);
SELECT * FROM users WHERE age = 30;  -- Doesn't use index efficiently
```

The index is sorted by `city` first. Filtering by `age` alone requires scanning the entire index.

**Fix**: Create `idx_age` or `idx_age_city`.

### 4. **Type Mismatch**

```sql
CREATE INDEX idx_id ON users(id);  -- id is INTEGER
SELECT * FROM users WHERE id = '42';  -- Query uses STRING
```

PostgreSQL can't use the index because it must cast `id` to text for comparison.

**Fix**: Use the correct type:
```sql
SELECT * FROM users WHERE id = 42;
```

### 5. **Function on Indexed Column**

```sql
SELECT * FROM users WHERE UPPER(name) = 'ALICE';
```

The index is on `name`, not `UPPER(name)`. Can't be used.

**Fix**: Create an expression index:
```sql
CREATE INDEX idx_upper_name ON users(UPPER(name));
```

---

## Index Maintenance Costs

Indexes aren't free. They consume:
- **Disk space**: Each index is a separate data structure
- **Write performance**: Every insert/update/delete modifies all indexes
- **Cache space**: Indexes compete with table data for buffer cache

### Monitoring Index Usage

PostgreSQL tracks index usage:
```sql
SELECT 
  schemaname, 
  tablename, 
  indexname, 
  idx_scan, 
  idx_tup_read, 
  idx_tup_fetch,
  pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes
ORDER BY idx_scan;
```

**Look for**:
- `idx_scan = 0`: Index is never used—drop it!
- Large indexes with low `idx_scan`: Rarely used—consider dropping

### Dropping Unused Indexes

```sql
DROP INDEX idx_unused;
```

**Impact**:
- Frees disk space
- Speeds up writes
- Reduces cache pollution

**Risk**: If the index was used by a rare but important query, that query becomes slow.

**Best practice**: Monitor for a week before dropping. Or use `DROP INDEX CONCURRENTLY` (doesn't lock the table).

---

## PostgreSQL vs MySQL Index Differences

| Feature                    | PostgreSQL                  | MySQL InnoDB              |
|----------------------------|-----------------------------|---------------------------|
| Primary key is             | Normal index                | Clustered index (table)   |
| Secondary indexes store    | TID (6 bytes)               | Primary key (variable)    |
| Index-only scans           | Requires visibility map     | Always possible           |
| Covering index overhead    | Must check visibility       | None                      |
| Partial indexes            | Yes                         | No (before 8.0)           |
| Expression indexes         | Yes                         | Yes (as generated column) |
| Concurrent index creation  | Yes (`CREATE INDEX CONCURRENTLY`) | Yes (InnoDB)       |

**Key difference**: In MySQL, secondary indexes store the **primary key**, not a tuple ID. This means:
- Large primary keys (e.g., UUID) make all secondary indexes larger
- Secondary index lookups require two B-Tree traversals (index → clustered index)

---

## Practical Takeaways

### 1. **Index Columns You Filter, Join, and Sort By**

```sql
SELECT * FROM orders 
WHERE user_id = 42 
ORDER BY created_at DESC;
```

**Indexes needed**:
- `idx_user_id_created_at` (composite index handles both `WHERE` and `ORDER BY`)

### 2. **Avoid Indexing Low-Cardinality Columns**

```sql
CREATE INDEX idx_gender ON users(gender);  -- Only 3 values: M, F, Other
```

The index is nearly useless. Most queries will still scan a huge fraction of the table.

**Exception**: Partial indexes on rare values:
```sql
CREATE INDEX idx_deleted ON users(deleted_at) WHERE deleted_at IS NOT NULL;
```

### 3. **Use Composite Indexes for Multi-Column Queries**

```sql
SELECT * FROM logs WHERE user_id = 42 AND event_type = 'click';
```

**Bad**:
```sql
CREATE INDEX idx_user_id ON logs(user_id);
CREATE INDEX idx_event_type ON logs(event_type);
```

The database can only use one index. It must filter the other column by scanning.

**Good**:
```sql
CREATE INDEX idx_user_id_event_type ON logs(user_id, event_type);
```

### 4. **Monitor Index Bloat (PostgreSQL)**

Heavy updates cause index bloat (dead tuples in the index).

**Check bloat**:
```sql
SELECT 
  schemaname, 
  tablename, 
  indexname,
  pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes;
```

**Fix**: `REINDEX` or `VACUUM`.

### 5. **Don't Over-Index**

Start with these indexes:
- Primary key
- Foreign keys
- Columns in frequent `WHERE` clauses

Add more indexes **only after measuring slow queries in production**.

---

## Summary

- B-Trees are shallow (depth ~3-5 even for billion-row tables)
- Index lookups are O(log N), extremely fast
- Range scans are efficient because leaf nodes are linked
- Random inserts cause page splits and fragmentation
- Index-only scans avoid heap fetches (huge performance win)
- Too many indexes slow down writes
- Multi-column index order matters (left-prefix rule)
- The query planner sometimes skips indexes if the cost is higher than a sequential scan

Understanding B-Trees lets you predict query performance and design optimal schemas.
