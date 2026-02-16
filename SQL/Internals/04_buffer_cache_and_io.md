# Buffer Cache and I/O

## Introduction

Why is the first query slow and the second one fast?

```sql
-- First run: 2 seconds
SELECT * FROM users WHERE age > 30;

-- Second run (immediate): 20 milliseconds
SELECT * FROM users WHERE age > 30;
```

The answer: **the buffer cache**.

The buffer cache is the database's in-memory storage layer. It's the difference between reading from disk (milliseconds) and reading from RAM (microseconds)—a **1000× performance difference**.

This chapter explains how the buffer cache works and why understanding it is critical for predicting and optimizing query performance.

---

## What Is the Buffer Cache?

The **buffer cache** (also called **shared buffers** in PostgreSQL or **buffer pool** in MySQL) is an in-memory cache of database pages.

```
┌─────────────────────────────────────┐
│          Application                │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│        Query Executor               │
└──────────────┬──────────────────────┘
               │ Request page 42
               ▼
┌─────────────────────────────────────┐
│         Buffer Cache                │
│  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐  │
│  │ P1  │ │ P42 │ │ P99 │ │ ...│  │  ← Pages in RAM
│  └─────┘ └─────┘ └─────┘ └─────┘  │
└──────────────┬──────────────────────┘
               │ Cache miss → read from disk
               ▼
┌─────────────────────────────────────┐
│         Disk (SSD/HDD)              │
│  All pages stored persistently      │
└─────────────────────────────────────┘
```

### How It Works

1. **Query requests a page** (e.g., "give me page 42")
2. **Check the buffer cache**:
   - **Cache hit**: Page is in memory → return immediately (microseconds)
   - **Cache miss**: Page is not in memory → read from disk (milliseconds)
3. **Store the page in cache** for future requests
4. **Evict old pages** if the cache is full (LRU or clock sweep algorithm)

### Key Terminology

- **Cache hit**: Requested page is in memory
- **Cache miss**: Requested page must be read from disk
- **Cache hit ratio**: Percentage of requests satisfied from cache (target: >99%)
- **Eviction**: Removing a page from cache to make room for a new page

---

## Shared Buffers (PostgreSQL)

PostgreSQL's buffer cache is called **shared_buffers**.

### Configuration

```sql
SHOW shared_buffers;  -- Default: 128 MB (way too small!)
```

**Typical settings**:
- Development: 128 MB
- Production: 25% of total RAM (e.g., 8 GB on a 32 GB server)
- Large production: 16-32 GB

**Why not use all RAM?** PostgreSQL relies on the OS page cache too. Setting `shared_buffers` too high can reduce OS cache effectiveness.

### Monitoring Cache Hit Ratio

```sql
SELECT 
  sum(heap_blks_read) AS disk_reads,
  sum(heap_blks_hit) AS cache_hits,
  sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)) AS hit_ratio
FROM pg_statio_user_tables;
```

**Target**: >0.99 (99% cache hit ratio).

If your hit ratio is <0.90, you're doing too much disk I/O. Either:
- Increase `shared_buffers`
- Optimize queries to read less data
- Add indexes

---

## Buffer Pool (MySQL InnoDB)

MySQL's buffer pool is similar but configured differently.

### Configuration

```sql
SHOW VARIABLES LIKE 'innodb_buffer_pool_size';
```

**Typical settings**:
- Development: 128 MB
- Production: 70-80% of total RAM (e.g., 24 GB on a 32 GB server)

MySQL's buffer pool is **more aggressive** than PostgreSQL's shared buffers because InnoDB manages more of its own caching.

### Monitoring Cache Hit Ratio

```sql
SHOW STATUS LIKE 'Innodb_buffer_pool_%';
```

**Key metrics**:
- `Innodb_buffer_pool_read_requests`: Total read requests
- `Innodb_buffer_pool_reads`: Disk reads (cache misses)

**Hit ratio**:
```
hit_ratio = 1 - (Innodb_buffer_pool_reads / Innodb_buffer_pool_read_requests)
```

---

## Cache Hits vs Disk Reads

### Performance Difference

**RAM access**: ~100 nanoseconds (0.0001 ms)
**SSD access**: ~0.1 milliseconds (100 microseconds)
**HDD access**: ~10 milliseconds

**Cache hit is ~1000× faster than SSD, ~100,000× faster than HDD.**

### Real-World Example

Table: 1 million rows, 100 bytes each = 100 MB.
Query: `SELECT * FROM users WHERE age > 30;`

**Cold cache** (first run):
- Must read 100 MB from disk
- SSD: 100 MB / 500 MB/s = 0.2 seconds
- HDD: 100 MB / 100 MB/s = 1 second

**Warm cache** (after first run):
- All pages in RAM
- 100 MB in memory = instant (< 10 ms to scan)

**100× faster with a warm cache!**

---

## Cold Cache vs Warm Cache

### Cold Cache

The buffer cache is empty (or doesn't contain the pages you need).

**When this happens**:
- Database just started
- You're querying a table for the first time
- The table is huge and doesn't fit in cache
- Other queries have evicted your pages

**Symptom**: First query is slow.

### Warm Cache

The buffer cache contains the pages you need.

**Symptom**: Subsequent queries are fast.

### Why This Matters

**In production**: The cache is usually warm because applications repeatedly query the same data.

**In benchmarks**: Beware of misleading results. If you run a query 10 times in a row, the first run is cold, the rest are warm.

Always **test cold cache performance** by:
1. Restarting the database, or
2. Clearing the cache, or
3. Querying different data each time

**PostgreSQL: Clear cache (dangerous in production!)**
```sql
-- Restart PostgreSQL
pg_ctl restart
```

**Linux: Clear OS page cache (dangerous!)**
```bash
sync; echo 3 > /proc/sys/vm/drop_caches
```

---

## How Pages Are Evicted

When the buffer cache is full, the database must evict pages to make room for new ones.

### LRU (Least Recently Used)

The most common eviction policy:
- Keep frequently accessed pages in cache
- Evict pages that haven't been accessed in a while

### Clock Sweep (PostgreSQL)

PostgreSQL uses a variant called **clock sweep** (similar to LRU but more efficient):
- Each page has a "usage count"
- When a page is accessed, its usage count increases
- When the cache is full, the database scans pages in a circular order and decrements usage counts
- Pages with usage count = 0 are evicted

**Implication**: Frequently accessed pages stay in cache, even if the cache is much smaller than the database.

### MySQL's LRU with Midpoint Insertion

InnoDB uses a **two-segment LRU**:
- **New pages** are inserted at the "midpoint" (not the most recently used end)
- Only if accessed again do they move to the "hot" end
- This prevents full table scans from evicting hot pages

**Why?** A large sequential scan would evict hot pages if new pages were inserted at the hot end.

---

## Why Large Scans Hurt Performance

```sql
-- Query 1: Scan 10 GB table
SELECT COUNT(*) FROM logs;

-- Query 2: Frequently accessed query
SELECT * FROM users WHERE id = 42;
```

**Problem**: Query 1 reads 10 GB of data. If the buffer cache is only 8 GB, it evicts all existing pages.

**Result**: Query 2, which was fast (cached), becomes slow (cache miss).

**This is called "cache pollution."**

### Solutions

#### 1. **Use LIMIT for large scans**
```sql
SELECT * FROM logs ORDER BY created_at DESC LIMIT 1000;
```

Don't scan the entire table if you only need recent rows.

#### 2. **Use indexes**
```sql
SELECT COUNT(*) FROM logs WHERE created_at > '2024-01-01';
-- With an index on created_at, this reads only relevant pages
```

#### 3. **Run analytics queries on a read replica**

Don't run heavy scans on the primary database.

#### 4. **Increase buffer cache size**

If your "working set" (frequently accessed data) is 10 GB, but your cache is only 1 GB, you'll have constant cache misses.

---

## Working Set

The **working set** is the subset of your database that is actively accessed.

Example:
- Database size: 100 GB
- Working set: 5 GB (recent users, active orders, etc.)

If your buffer cache is ≥ 5 GB, your working set fits in memory. **All queries are fast.**

If your buffer cache is < 5 GB, pages are constantly evicted and reloaded. **Queries are slow.**

### How to Measure Your Working Set

**Heuristic**: Look at the cache hit ratio over time. If it's >99%, your working set fits in cache.

If it's <95%, your working set is larger than your cache.

**Query for table sizes**:
```sql
SELECT 
  schemaname, 
  tablename, 
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

**Question**: Are the largest tables frequently accessed? If yes, you need a bigger cache.

---

## fsync and Durability Costs

When you commit a transaction:
```sql
COMMIT;
```

The database must ensure that the changes are **durable** (survive a crash).

### Write-Ahead Logging (WAL)

PostgreSQL and MySQL use **WAL** (write-ahead logging):
1. Write changes to the WAL (a sequential log file)
2. Call `fsync` to flush the WAL to disk
3. Return success to the client
4. Periodically write changes from the WAL to the actual data files

**Durability guarantee**: Even if the database crashes after commit, the WAL contains the changes, so they can be replayed during recovery.

### The Cost of fsync

`fsync` is **expensive**:
- SSD: ~0.1 ms per fsync
- HDD: ~10 ms per fsync

**Every commit requires an fsync.**

### Group Commit

Modern databases batch multiple commits into a single fsync:
- Transaction 1 commits
- Transaction 2 commits (before fsync completes)
- Transaction 3 commits
- Database does **one fsync for all three transactions**

**Result**: High-throughput commit rates (thousands of commits per second).

**But**: If you commit transactions serially (one at a time), you're limited by fsync latency:
- SSD: ~10,000 commits/sec
- HDD: ~100 commits/sec

### Disabling fsync (Dangerous!)

For testing or bulk loads, you can disable fsync:

**PostgreSQL**:
```sql
SET fsync = off;
```

**Warning**: If the database crashes, you lose **all uncommitted data**, and potentially **committed data** too. Only use this for disposable test environments.

---

## Why "First Query Is Slow"

You've seen this pattern:

```sql
-- First run: 1 second
SELECT * FROM orders WHERE user_id = 42;

-- Second run: 10 milliseconds
SELECT * FROM orders WHERE user_id = 42;
```

**Explanation**:
1. First run: Cache miss → read from disk (slow)
2. Pages are loaded into cache
3. Second run: Cache hit → read from memory (fast)

**Implications for benchmarking**: Always run queries multiple times. The first run is misleading.

**Implications for production**: The "first query slow" problem is rare because caches warm up quickly after restart.

**Exception**: If your database is so large that the working set doesn't fit in cache, *every* query might have cache misses (this is bad—you need more RAM or better indexing).

---

## SSD vs HDD Assumptions

Modern databases are designed for **SSDs**, not HDDs.

### HDDs (Old School)

- **Sequential reads**: Fast (~100 MB/s)
- **Random reads**: Slow (~100 IOPS = 100 reads/sec)
- **Seek time**: ~10 ms

**Implication**: Random I/O is catastrophically slow. Index lookups are often slower than sequential scans.

### SSDs (Modern)

- **Sequential reads**: Fast (~500 MB/s)
- **Random reads**: Fast (~10,000 IOPS)
- **Latency**: ~0.1 ms

**Implication**: Random I/O is only ~10× slower than sequential. Index lookups are almost always faster than sequential scans.

### Query Planner Assumptions

The query planner's cost model was designed in the HDD era. It penalizes random I/O heavily.

On SSDs, the penalty is overstated. Sometimes the planner chooses a sequential scan when an index would be faster.

**You can tune this**:

**PostgreSQL**:
```sql
-- Default: random_page_cost = 4, seq_page_cost = 1
-- For SSDs:
SET random_page_cost = 1.1;
```

This tells the planner that random I/O is only slightly more expensive than sequential I/O.

---

## Read-Ahead and Prefetching

Databases try to predict which pages you'll need next and prefetch them.

### Sequential Read-Ahead

If you're scanning a table sequentially:
```sql
SELECT * FROM orders ORDER BY created_at;
```

The database reads pages 1, 2, 3, ... in order.

**Optimization**: The database prefetches pages 4, 5, 6, ... before you need them.

**Result**: Sequential scans are very fast (no I/O stalls).

### Index Read-Ahead (Limited)

If you're scanning an index:
```sql
SELECT * FROM orders WHERE user_id = 42;
```

The database can prefetch the next leaf nodes in the B-Tree.

But it **cannot** prefetch heap pages (they're in random order).

**Result**: Index scans are slower than sequential scans if the index returns many rows.

---

## Cache Coherency and the OS Page Cache

PostgreSQL uses **two layers of caching**:
1. **Shared buffers** (database's cache)
2. **OS page cache** (kernel's cache)

When PostgreSQL reads a page:
1. Check shared buffers
2. If miss, read from file (which may be in the OS page cache)
3. If OS cache miss, read from disk

**Double caching** means:
- Your effective cache is larger than `shared_buffers`
- But there's overhead (data is cached twice)

**MySQL InnoDB** manages its own cache more aggressively and bypasses the OS page cache for most operations (using `O_DIRECT`).

---

## Monitoring I/O Performance

### PostgreSQL: pg_statio_user_tables

```sql
SELECT 
  schemaname, 
  tablename, 
  heap_blks_read AS disk_reads,
  heap_blks_hit AS cache_hits,
  heap_blks_hit::float / (heap_blks_hit + heap_blks_read) AS hit_ratio
FROM pg_statio_user_tables
WHERE heap_blks_read + heap_blks_hit > 0
ORDER BY disk_reads DESC;
```

**Look for**:
- Tables with high `disk_reads` → candidates for caching or indexing
- Tables with low `hit_ratio` → either too large to cache or queried infrequently

### MySQL: SHOW ENGINE INNODB STATUS

```sql
SHOW ENGINE INNODB STATUS\G
```

**Look for**:
- `Buffer pool hit rate`: Should be >99%
- `Pending reads`: If high, you're I/O bound

### OS-Level Monitoring

**iostat** (Linux):
```bash
iostat -x 1
```

**Look for**:
- `%util`: If near 100%, your disk is saturated
- `await`: Average I/O latency (should be <1 ms on SSD)

---

## Practical Takeaways

### 1. **Set shared_buffers Appropriately**

**PostgreSQL**: 25% of RAM
**MySQL**: 70-80% of RAM

Don't leave it at the default (128 MB)!

### 2. **Monitor Cache Hit Ratio**

Target: >99%.

If your hit ratio is low:
- Increase cache size
- Optimize queries (use indexes)
- Reduce working set (archive old data)

### 3. **Beware of Cache Pollution**

Large scans can evict hot data.

**Solutions**:
- Use `LIMIT`
- Use indexes
- Run analytics on replicas

### 4. **Trust SSDs**

If you're on SSDs, adjust `random_page_cost`:
```sql
SET random_page_cost = 1.1;
```

This makes the planner more likely to use indexes.

### 5. **Warm Up Caches After Restart**

After restarting the database, run common queries to warm the cache:
```bash
psql -c "SELECT COUNT(*) FROM users;"
psql -c "SELECT COUNT(*) FROM orders;"
```

This prevents the "first query slow" problem in production.

### 6. **Understand Your Working Set**

If your working set is 10 GB and your cache is 1 GB, you'll have constant cache misses.

**Either**:
- Increase RAM
- Reduce working set (archive old data, add indexes)

---

## Summary

- The buffer cache is the database's in-memory storage layer
- Cache hits are **1000× faster** than disk reads
- Cold cache = first query slow; warm cache = fast queries
- Cache hit ratio should be >99%
- Large scans can pollute the cache and slow down other queries
- fsync is expensive (required for durability)
- SSDs are much faster than HDDs for random I/O
- PostgreSQL uses shared buffers + OS page cache (double caching)
- MySQL uses a large buffer pool and bypasses OS cache

Understanding the buffer cache lets you predict when queries will be fast or slow—and how to fix performance problems.
