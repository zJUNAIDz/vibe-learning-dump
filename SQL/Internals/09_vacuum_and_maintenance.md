# VACUUM and Maintenance

## Introduction

You've deployed your PostgreSQL app. It's been running great for months. Then suddenly:
- Queries that took 10ms now take 1 second
- Disk usage has ballooned from 10 GB to 100 GB
- `COUNT(*)` is painfully slow

**What happened?**

**Table bloat.**

PostgreSQL's MVCC creates dead tuples with every `UPDATE` and `DELETE`. Those dead tuples accumulate over time, wasting disk space and slowing down queries.

**The solution: VACUUM.**

This chapter explains why VACUUM exists, how it works, and how to tune it for production systems.

---

## Why VACUUM Exists (PostgreSQL-Specific)

### The MVCC Problem

Every `UPDATE` in PostgreSQL creates a new tuple. The old tuple is marked "dead" but remains on disk.

```sql
UPDATE users SET balance = 200 WHERE id = 1;
-- Creates a new tuple (balance = 200)
-- Old tuple (balance = 100) is marked dead
```

After 1 million updates, you have:
- 1 live tuple
- 999,999 dead tuples

**Dead tuples**:
- Waste disk space
- Slow down sequential scans (must read and skip them)
- Slow down index scans (indexes point to dead tuples too)

**VACUUM's job**: Remove dead tuples and reclaim space.

---

## What VACUUM Does

```sql
VACUUM users;
```

**Actions**:
1. **Scans the table** to find dead tuples
2. **Marks pages with dead tuples as reusable** (updates the free space map)
3. **Updates the visibility map** (marks pages where all tuples are visible)
4. **Truncates trailing empty pages** (returns disk space to the OS, if possible)

**What VACUUM does NOT do**:
- Rewrite the table
- Lock the table (queries can run concurrently)
- Immediately free disk space (space is reusable but the file size doesn't shrink)

---

## Dead Tuples

### How Dead Tuples Accumulate

```sql
-- Initial state: 1 row
INSERT INTO users (id, name, balance) VALUES (1, 'Alice', 100);
-- Table size: 8 KB (1 page)

-- Update 1 million times
DO $$
BEGIN
  FOR i IN 1..1000000 LOOP
    UPDATE users SET balance = balance + 1 WHERE id = 1;
  END LOOP;
END $$;
-- Table size: ~100 MB (1 live tuple + 999,999 dead tuples)
```

**Disk usage is proportional to the number of tuples (dead + live), not the number of live rows.**

---

## Checking for Bloat

### Dead Tuple Count

```sql
SELECT 
  schemaname, 
  tablename, 
  n_live_tup AS live_tuples,
  n_dead_tup AS dead_tuples,
  round(n_dead_tup * 100.0 / NULLIF(n_live_tup + n_dead_tup, 0), 2) AS dead_ratio
FROM pg_stat_user_tables
ORDER BY n_dead_tup DESC;
```

**Thresholds**:
- `dead_ratio < 10%`: Healthy
- `dead_ratio 10-20%`: Autovacuum should kick in soon
- `dead_ratio > 20%`: Heavy bloat. Investigate autovacuum.

### Table Size

```sql
SELECT 
  schemaname, 
  tablename, 
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
  pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) AS indexes_size
FROM pg_tables
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

**Look for**:
- Tables that are much larger than expected
- Tables growing over time even though row count is stable

---

## Autovacuum

PostgreSQL has **autovacuum**, a background process that runs `VACUUM` automatically.

### How Autovacuum Decides When to Run

Autovacuum triggers when:
```
dead_tuples > autovacuum_vacuum_threshold + autovacuum_vacuum_scale_factor * n_live_tuples
```

**Default settings**:
- `autovacuum_vacuum_threshold`: 50
- `autovacuum_vacuum_scale_factor`: 0.2

**Example**:
- Table: 1 million rows
- Threshold: 50 + 0.2 × 1,000,000 = 200,050 dead tuples

**Autovacuum runs after 200,000 dead tuples accumulate** (~20% of the table).

**Problem**: For large tables, 20% bloat is huge. You might want to vacuum more frequently.

### Tuning Autovacuum

**Per-table tuning**:
```sql
ALTER TABLE users SET (autovacuum_vacuum_scale_factor = 0.05);
-- Now autovacuum runs after 5% bloat instead of 20%
```

**Global tuning** (`postgresql.conf`):
```conf
autovacuum_vacuum_scale_factor = 0.1
autovacuum_vacuum_threshold = 50
autovacuum_max_workers = 3  # Number of parallel autovacuum workers
autovacuum_naptime = 10s    # How often autovacuum checks for work
```

---

## When Autovacuum Falls Behind

### Heavy Write Workload

If you're updating 1 million rows per second, autovacuum can't keep up.

**Symptom**: `n_dead_tup` keeps increasing.

**Fix**:
1. Increase autovacuum workers:
   ```conf
   autovacuum_max_workers = 10
   ```
2. Decrease the scale factor (vacuum more often):
   ```sql
   ALTER TABLE users SET (autovacuum_vacuum_scale_factor = 0.01);
   ```
3. Run manual VACUUM during low-traffic periods:
   ```sql
   VACUUM users;
   ```

### Long-Running Transactions

Autovacuum **cannot** remove dead tuples if a long-running transaction might still need them.

**Example**:
```sql
-- Transaction A (started at 10:00 AM):
BEGIN;
SELECT * FROM users WHERE id = 1;
-- (Leave transaction open for 1 hour)

-- At 10:30 AM: You update 1 million rows
UPDATE users SET balance = balance + 1;

-- At 10:30 AM: Autovacuum tries to clean up
VACUUM users;
-- Result: Cannot clean up the 1 million dead tuples because Transaction A might still need them
```

**Fix**: Kill idle or long-running transactions:
```sql
SELECT pg_terminate_backend(pid) 
FROM pg_stat_activity 
WHERE state = 'idle in transaction' AND now() - xact_start > interval '5 minutes';
```

---

## Manual VACUUM

Sometimes you need to run VACUUM manually.

### VACUUM (Standard)

```sql
VACUUM users;
```

**Does**:
- Marks dead tuples as reusable
- Updates visibility map
- Doesn't block queries

**Doesn't**:
- Shrink the table file (space is reusable but not returned to the OS)

### VACUUM FULL (Aggressive)

```sql
VACUUM FULL users;
```

**Does**:
- Rewrites the entire table (removes dead tuples)
- Shrinks the table file (returns space to the OS)
- **Acquires ACCESS EXCLUSIVE lock** (blocks all queries!)

**Use case**: Only if you have severe bloat (>50%) and can afford downtime.

**Alternative**: Use `pg_repack` (online table rewrite, no locking).

---

## VACUUM vs VACUUM FULL

| Feature                | VACUUM                          | VACUUM FULL                    |
|------------------------|---------------------------------|--------------------------------|
| Removes dead tuples    | Yes                             | Yes                            |
| Returns space to OS    | No (marks as reusable)          | Yes (shrinks file)             |
| Locks table            | No (concurrent queries allowed) | Yes (ACCESS EXCLUSIVE)         |
| Time to complete       | Fast (depends on dead tuples)   | Slow (rewrites entire table)   |
| Use in production      | Yes (safe)                      | Avoid (causes downtime)        |

---

## The Visibility Map

The **visibility map** is a bitmap tracking which pages have "all tuples visible" (no dead tuples, no uncommitted transactions).

**Purpose**: Optimize queries and vacuums.

### How It Helps

**1. Index-Only Scans**

If all tuples on a page are visible, PostgreSQL can skip fetching the heap page during an index-only scan.

```sql
CREATE INDEX idx_age ON users(age);
SELECT age FROM users WHERE age > 30;
-- If the visibility map marks all pages as visible, this is truly index-only
```

**2. VACUUM Optimization**

VACUUM can skip pages marked as "all visible" (no dead tuples to clean up).

**When the visibility map is updated**: After VACUUM confirms all tuples are visible.

---

## ANALYZE

`ANALYZE` is different from `VACUUM`. It updates **statistics** for the query planner.

```sql
ANALYZE users;
```

**Does**:
- Samples rows from the table
- Updates statistics (distinct values, most common values, histograms)
- Helps the query planner choose better plans

**Doesn't**:
- Remove dead tuples
- Affect disk space

**Auto-analyze**: Runs automatically (similar to autovacuum).

**When to run manually**:
- After bulk inserts/updates
- If query plans are bad (check with `EXPLAIN ANALYZE`)

---

## Bloat in Indexes

Indexes also bloat.

### Cause

Every `UPDATE` that changes an indexed column creates a new index entry. The old entry is marked dead but remains in the index.

### Detecting Index Bloat

```sql
SELECT 
  schemaname, 
  tablename, 
  indexname, 
  pg_size_pretty(pg_relation_size(indexrelid)) AS index_size,
  idx_scan AS index_scans
FROM pg_stat_user_indexes
ORDER BY pg_relation_size(indexrelid) DESC;
```

**Look for**:
- Indexes larger than expected
- Indexes with low `idx_scan` (unused indexes)

### Fixing Index Bloat

**Option 1: REINDEX**
```sql
REINDEX INDEX idx_users_email;
-- Or reindex the entire table:
REINDEX TABLE users;
```

**Problem**: Acquires ACCESS EXCLUSIVE lock (blocks queries).

**Option 2: REINDEX CONCURRENTLY** (PostgreSQL 12+)
```sql
REINDEX INDEX CONCURRENTLY idx_users_email;
```

**Does**: Rebuilds the index without locking.

**Downside**: Takes longer (builds a new index in parallel, then swaps).

---

## Table Bloat: When VACUUM Isn't Enough

Even with regular VACUUM, tables can accumulate free space faster than autovacuum can reclaim it.

### Causes

1. **Heavy update workload**: Updating 10% of rows per minute creates 10% dead tuples per minute. Autovacuum can't keep up.
2. **Long-running transactions**: Block autovacuum from cleaning dead tuples.
3. **High `autovacuum_vacuum_scale_factor`**: Autovacuum only runs after 20% bloat.

### Solutions

#### 1. **Tune Autovacuum**

```sql
ALTER TABLE users SET (autovacuum_vacuum_scale_factor = 0.05);
```

#### 2. **Run VACUUM More Frequently**

Add a cron job:
```bash
*/10 * * * * psql -c "VACUUM users;"
```

#### 3. **Use pg_repack**

Rewrites the table online (no locking):
```bash
pg_repack -d mydb -t users
```

#### 4. **Partition the Table**

If the table is huge (>100 GB), consider partitioning. Vacuum smaller partitions independently.

---

## VACUUM and Performance

### When VACUUM Is Slow

**Cause 1: Large table with many dead tuples**

VACUUM must scan the entire table to find dead tuples.

**Fix**: Tune autovacuum to run more often (prevent bloat from accumulating).

**Cause 2: I/O bottleneck**

VACUUM reads many pages from disk.

**Fix**: Run VACUUM during low-traffic periods.

### VACUUM's Impact on Production

**Normally**: VACUUM is lightweight. It reads pages, marks dead tuples, and updates metadata.

**But**: On a heavily-updated table with millions of dead tuples, VACUUM can:
- Read gigabytes of data from disk
- Increase I/O load
- Compete with queries for buffer cache

**Mitigation**: Tune `autovacuum_vacuum_cost_limit` to throttle VACUUM:
```conf
autovacuum_vacuum_cost_limit = 200  # Default: -1 (use vacuum_cost_limit)
vacuum_cost_limit = 200             # Throttle VACUUM I/O
```

Lower values = slower VACUUM = less impact on queries.

---

## Monitoring Autovacuum

### Check Last Autovacuum

```sql
SELECT 
  schemaname, 
  tablename, 
  last_vacuum, 
  last_autovacuum, 
  n_dead_tup
FROM pg_stat_user_tables
ORDER BY n_dead_tup DESC;
```

**Look for**:
- `last_autovacuum IS NULL` or very old → autovacuum not running
- High `n_dead_tup` → bloat accumulating

### Check Autovacuum Activity

```sql
SELECT 
  pid, 
  now() - xact_start AS duration, 
  query
FROM pg_stat_activity
WHERE query LIKE '%autovacuum%';
```

**Look for**:
- Long-running autovacuum (>30 minutes) → large bloat

---

## Why DELETE Doesn't Free Space Immediately

```sql
DELETE FROM logs WHERE created_at < '2020-01-01';
-- Deleted 10 GB of rows
SELECT pg_size_pretty(pg_total_relation_size('logs'));
-- Still 100 GB (no space freed!)
```

**Why?**

`DELETE` marks tuples as dead but doesn't remove them physically.

**VACUUM reclaims the space**, but:
- Space is marked reusable (available for new rows)
- But the file size doesn't shrink (space not returned to the OS)

**To actually shrink the file**:
```sql
VACUUM FULL logs;  -- Shrinks the file (but locks the table!)
```

**Or**:
```bash
pg_repack -d mydb -t logs  # Shrinks the file (no locking)
```

---

## PostgreSQL vs MySQL Maintenance

| Feature                | PostgreSQL (VACUUM)             | MySQL InnoDB                   |
|------------------------|---------------------------------|--------------------------------|
| Dead tuple storage     | In table file (bloat)           | In undo log (no bloat)         |
| Maintenance required   | Yes (VACUUM)                    | No (purge thread automatic)    |
| Space reclamation      | Marks reusable (doesn't shrink) | Automatic                      |
| Long transaction impact | Blocks VACUUM                   | Grows undo log                 |
| Table rewrite required | VACUUM FULL (locks table)       | OPTIMIZE TABLE (locks table)   |

**Takeaway**: MySQL's undo log avoids table bloat. PostgreSQL's MVCC requires VACUUM maintenance.

---

## Practical Takeaways

### 1. **Monitor Dead Tuples**

```sql
SELECT tablename, n_live_tup, n_dead_tup FROM pg_stat_user_tables ORDER BY n_dead_tup DESC;
```

If `n_dead_tup` is >20% of `n_live_tup`, investigate.

### 2. **Tune Autovacuum for Hot Tables**

```sql
ALTER TABLE users SET (autovacuum_vacuum_scale_factor = 0.05);
```

### 3. **Avoid Long-Running Transactions**

They block VACUUM from cleaning dead tuples.

**Monitor**:
```sql
SELECT pid, state, now() - xact_start FROM pg_stat_activity WHERE state = 'idle in transaction';
```

**Kill**:
```sql
SELECT pg_terminate_backend(pid);
```

### 4. **Run VACUUM Manually During Maintenance Windows**

If autovacuum can't keep up:
```sql
VACUUM ANALYZE users;
```

### 5. **Use pg_repack for Severely Bloated Tables**

Don't use `VACUUM FULL` in production (it locks the table).

### 6. **Partition Large Tables**

Vacuuming a 1 TB table is slow. Vacuum 10 × 100 GB partitions instead.

### 7. **Avoid Indexing Frequently-Updated Columns**

Every update to an indexed column creates dead index entries.

**Alternative**: Store frequently-updated columns in a separate table.

---

## Summary

- VACUUM removes dead tuples created by MVCC
- Dead tuples waste disk space and slow down queries
- Autovacuum runs automatically but can fall behind
- Long-running transactions block VACUUM
- VACUUM marks space reusable but doesn't shrink the file
- VACUUM FULL shrinks the file but locks the table (avoid in production)
- pg_repack is a better alternative (online rewrite)
- Monitor dead tuples with `pg_stat_user_tables`
- Tune autovacuum for hot tables
- Kill idle transactions to allow VACUUM to proceed

Understanding VACUUM is essential for maintaining PostgreSQL performance in production.
