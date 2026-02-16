# Storage Engine Basics

## Introduction

When you run:
```sql
SELECT * FROM users WHERE id = 42;
```

What actually happens on disk?

Most developers think: "The database finds row 42 and returns it." But the reality is far more nuanced—and understanding it is the key to understanding database performance.

This chapter explains how databases physically store and retrieve data.

---

## The Fundamental Unit: Pages (Blocks)

Databases don't read or write individual rows. They read and write **pages** (also called **blocks**).

### What Is a Page?

A page is a fixed-size chunk of data:
- **PostgreSQL**: 8 KB per page
- **MySQL (InnoDB)**: 16 KB per page
- **Oracle**: 8 KB per block (configurable)

Each page contains:
- **Header** (metadata: page type, free space pointer, checksum)
- **Row data** (actual tuples)
- **Special space** (indexes, btree pointers, etc.)

```
┌────────────────────────────────────────┐
│         Page Header (24 bytes)         │
├────────────────────────────────────────┤
│          Row 1 (120 bytes)             │
├────────────────────────────────────────┤
│          Row 2 (95 bytes)              │
├────────────────────────────────────────┤
│          Row 3 (200 bytes)             │
├────────────────────────────────────────┤
│              ...                       │
├────────────────────────────────────────┤
│         Free Space (512 bytes)         │
├────────────────────────────────────────┤
│       Special Space (64 bytes)         │
└────────────────────────────────────────┘
            8 KB Page
```

### Why Pages?

**Operating system I/O is page-based.**

When you ask the OS to read 1 byte from a file, it actually reads an entire 4 KB OS page (or 8 KB, depending on the system).

Databases align their page size with OS pages to minimize overhead.

**Key insight**: If your query needs to read 1 row, the database actually reads at least 1 entire page (8 KB in PostgreSQL).

---

## Rows vs Tuples

In PostgreSQL terminology:
- **Row**: The logical concept (what you insert with `INSERT`)
- **Tuple**: The physical storage of a row on disk

Why the distinction?

Because PostgreSQL uses **MVCC** (Multi-Version Concurrency Control), a single logical row can have multiple physical tuples on disk:

```sql
-- Transaction 1:
UPDATE users SET name = 'Alice' WHERE id = 1;

-- Transaction 2 (concurrent):
SELECT * FROM users WHERE id = 1;
```

Transaction 2 sees the **old tuple** (before the update).
Transaction 1 sees the **new tuple** (after the update).

Both tuples exist on disk simultaneously.

**Terminology**:
- PostgreSQL: "tuple"
- MySQL: "row"
- Oracle: "row"

We'll use "tuple" when discussing PostgreSQL-specific behavior, "row" otherwise.

---

## Heap Storage (PostgreSQL)

PostgreSQL uses **heap storage** for tables.

### What Is Heap Storage?

A heap is an unordered collection of pages. Rows are inserted wherever there's free space—no particular order.

```
┌─────┐  ┌─────┐  ┌─────┐  ┌─────┐
│ P0  │  │ P1  │  │ P2  │  │ P3  │
│ Row1│  │ Row4│  │ Row2│  │ Row9│
│ Row3│  │ Row5│  │ Row8│  │ Free│
└─────┘  └─────┘  └─────┘  └─────┘
         Table File (Heap)
```

**Characteristics**:
- No inherent ordering
- Fast inserts (just append to the end or fill free space)
- Slow full table scans (must read every page)

### Contrast: MySQL InnoDB (Clustered Index)

MySQL InnoDB does **not** use heap storage. It uses a **clustered index** on the primary key.

Rows are physically stored **in primary key order** inside a B-Tree.

```
         ┌─────────────┐
         │   Root Node │
         └──────┬──────┘
       ┌────────┴────────┐
       ▼                 ▼
  ┌─────────┐      ┌─────────┐
  │  id=1-5 │      │ id=6-10 │
  │  Row1   │      │  Row6   │
  │  Row2   │      │  Row7   │
  │  ...    │      │  ...    │
  └─────────┘      └─────────┘
  B-Tree with rows stored in leaf nodes
```

**Implications**:
- Fast primary key lookups (O(log N))
- Fast range scans on primary key (`WHERE id BETWEEN 1 AND 100`)
- **Slow** inserts on random primary keys (causes page splits)
- Secondary indexes store the primary key, not a row pointer (extra indirection)

**PostgreSQL vs MySQL**:
| Feature                     | PostgreSQL (Heap)        | MySQL InnoDB (Clustered) |
|-----------------------------|--------------------------|--------------------------|
| Primary key lookup          | O(log N) via index       | O(log N) directly        |
| Full table scan             | Read entire heap         | Read entire B-Tree       |
| Insert random PK            | Fast (heap append)       | Slow (page split)        |
| Insert sequential PK        | Fast                     | Fast                     |
| Secondary index size        | Stores row TID (6 bytes) | Stores PK (variable)     |

---

## Why Databases Read in Pages, Not Rows

Imagine a query:
```sql
SELECT * FROM users WHERE age > 30;
```

Even if only 10% of rows match, the database must read pages containing all rows to check the `age` column (unless there's an index on `age`).

### Example: Sequential Scan

Table: 1 million rows, avg 100 bytes per row.
Total size: 100 MB.
PostgreSQL page size: 8 KB.
Number of pages: ~12,500.

**Query cost**:
- Must read all 12,500 pages
- Even if only 100,000 rows match

**Why?** Because the database doesn't know which pages contain matching rows without reading them.

**Disk I/O cost**:
- SSD: 12,500 pages / 10,000 IOPS = ~1.25 seconds
- HDD: 12,500 pages / 100 IOPS = ~125 seconds (ouch!)

### With an Index

If there's an index on `age`, the database can:
1. Scan the index to find matching rows
2. Use the row pointers to fetch only relevant pages

**I/O cost**:
- Index scan: ~10 pages (B-Tree depth of 3-4)
- Heap fetch: ~100 pages (one per matching row, assuming scattered)

Total: ~110 pages / 10,000 IOPS = ~0.01 seconds.

**100× faster** with an index!

---

## How This Affects Query Performance

### 1. **Wide Rows = Fewer Rows per Page**

```sql
CREATE TABLE logs (
  id SERIAL PRIMARY KEY,
  timestamp TIMESTAMP,
  message TEXT  -- Can be 10 KB!
);
```

If `message` is large, each page might fit only 1-2 rows.

**Result**: A query that scans 1,000 rows must read 1,000 pages (instead of ~10 pages if rows were small).

**Fix**: Store large columns in separate tables (normalization) or use PostgreSQL's **TOAST** (The Oversized-Attribute Storage Technique).

### 2. **Small Rows = More Rows per Page**

```sql
CREATE TABLE events (
  id BIGINT PRIMARY KEY,
  type SMALLINT,
  user_id BIGINT,
  created_at TIMESTAMP
);
```

Each row is ~24 bytes. Each page fits ~300 rows.

**Result**: Scanning 1 million rows requires only ~3,300 pages (~26 MB). Fast!

### 3. **Dead Tuples Waste Space**

PostgreSQL's MVCC creates dead tuples:
```sql
UPDATE users SET last_login = NOW() WHERE id = 1;
```

The old tuple is marked dead but remains in the page until `VACUUM` removes it.

If you do this 1,000 times, you've created 1,000 dead tuples.

**Result**: The page is full of garbage. The database must read and skip dead tuples during scans.

**Fix**: Run `VACUUM` regularly (autovacuum does this automatically, but can fall behind under heavy write load).

### 4. **Partial Page Reads Are Impossible**

You can't read half a page. If your query needs 1 row in page 42, the database reads all 8 KB.

**Implication**: Storing rarely-accessed columns in the same table as frequently-accessed columns wastes I/O.

**Fix**: Vertical partitioning. Move rarely-accessed columns to a separate table.

---

## Row Storage Format

How are individual rows laid out in a page?

### PostgreSQL Tuple Format

Each tuple has:
- **Heap tuple header** (23 bytes): transaction IDs, flags, null bitmap
- **Column data**: actual values

```
┌───────────────────────────────────────┐
│         Heap Tuple Header             │
│  • t_xmin (4B): insert transaction ID │
│  • t_xmax (4B): delete transaction ID │
│  • t_cid (4B): command ID             │
│  • t_ctid (6B): current tuple ID      │
│  • null bitmap                        │
├───────────────────────────────────────┤
│  id (4 bytes)                         │
├───────────────────────────────────────┤
│  name (variable length)               │
├───────────────────────────────────────┤
│  email (variable length)              │
└───────────────────────────────────────┘
```

**Key point**: The tuple header alone is 23 bytes. If your row data is only 10 bytes, you're wasting 70% space on metadata!

**Implication**: Storing billions of tiny rows is inefficient in PostgreSQL. Consider batching or compression.

### MySQL InnoDB Row Format

InnoDB uses a more compact row format (DYNAMIC or COMPACT):
- **Row header**: ~5-6 bytes
- **Hidden columns**: transaction ID, rollback pointer (7 bytes)
- **Column data**

InnoDB is slightly more space-efficient for small rows.

---

## Table Files Are Just Files

PostgreSQL stores each table as a file in the data directory:

```
$ ls -lh /var/lib/postgresql/data/base/16384/
-rw------- 1 postgres postgres 8.0M  users
-rw------- 1 postgres postgres 64K   users_pkey
-rw------- 1 postgres postgres 1.2G  orders
```

Each file is divided into 8 KB pages.

**File size growth**:
- Inserting 1 million rows (100 bytes each) grows the file by ~100 MB
- Deleting 1 million rows does **not** shrink the file (PostgreSQL marks pages as reusable but doesn't return them to the OS)

**Implication**: Disk space doesn't automatically get reclaimed. You need `VACUUM FULL` (locks the table) or `pg_repack` (online).

---

## How Databases Know Where Rows Are

### PostgreSQL: TID (Tuple Identifier)

Every tuple has a **TID** (tuple identifier):
- Format: `(page_number, offset)`
- Example: `(42, 3)` means page 42, offset 3

Indexes store TIDs, not row pointers:
```
Index Entry:
  Key: "Alice"
  Value: TID (42, 3)

Heap Tuple:
  Page 42, Offset 3: (id=1, name="Alice", ...)
```

**Implication**: If you update a row and it moves to a different page (e.g., due to page split), the index still points to the old TID. PostgreSQL uses **HOT (Heap-Only Tuples)** to avoid updating indexes in some cases.

### MySQL InnoDB: Primary Key

InnoDB stores the **primary key** in secondary indexes, not a TID:

```
Secondary Index (on name):
  Key: "Alice"
  Value: Primary Key (id=1)

Clustered Index (primary key):
  Key: id=1
  Value: Entire row (id=1, name="Alice", ...)
```

**Implication**: 
- Secondary index lookups require **two B-Tree traversals** (index + clustered index)
- Large primary keys (e.g., UUID) make secondary indexes huge

---

## Sequential vs Random I/O

### Sequential Scan

Reading pages 0, 1, 2, 3, 4, ... in order.

**SSD performance**: ~500 MB/s
**HDD performance**: ~100 MB/s

Sequential scans are **fast** on SSDs because of read-ahead caching and parallelism.

### Random Access

Reading pages 42, 7, 192, 8, 305, ... in random order.

**SSD performance**: ~10,000 IOPS (can fetch 10,000 random pages per second)
**HDD performance**: ~100 IOPS (limited by disk seek time)

**Key insight**: On SSDs, random access is only ~10× slower than sequential. On HDDs, it's ~100× slower.

**Implication**: Index lookups are practical on SSDs but painful on HDDs.

---

## Postgres vs MySQL Storage Differences (Summary)

| Feature                | PostgreSQL (Heap)            | MySQL InnoDB (Clustered) |
|------------------------|------------------------------|--------------------------|
| Row storage            | Unordered heap               | Ordered by primary key   |
| Primary key lookup     | Index → TID → heap fetch     | Direct B-Tree lookup     |
| Secondary index lookup | Index → TID → heap fetch     | Index → PK → B-Tree      |
| UPDATE overhead        | Creates new tuple (MVCC)     | In-place or undo log     |
| Space reclamation      | VACUUM needed                | Purge thread (automatic) |
| Best insert pattern    | Any (heap append)            | Sequential primary key   |
| Worst insert pattern   | Random (causes bloat)        | Random primary key       |

**Rule of thumb**:
- PostgreSQL: Better for read-heavy workloads with complex queries
- MySQL: Better for write-heavy workloads with simple key-value lookups

---

## Practical Takeaways

### 1. **Keep Rows Small**

Wide rows mean fewer rows per page. Scanning N rows requires reading more pages.

**Fix**: Normalize large text columns into separate tables.

### 2. **Understand Heap vs Clustered Storage**

- PostgreSQL: Primary key lookups require an index scan + heap fetch
- MySQL: Primary key lookups are a single B-Tree traversal

**Implication**: In PostgreSQL, always create a primary key index. In MySQL, the primary key **is** the table.

### 3. **Sequential Primary Keys Are Good (MySQL)**

In MySQL, inserting rows with `id = 1, 2, 3, ...` appends to the B-Tree. Fast!

Inserting rows with `id = UUID()` causes random B-Tree insertions. Slow + fragmentation.

**Fix**: Use `AUTO_INCREMENT` or `SERIAL` for primary keys.

### 4. **Watch for Table Bloat (PostgreSQL)**

Heavy updates create dead tuples. If autovacuum can't keep up, your table grows 10× larger than necessary.

**Monitoring**:
```sql
SELECT 
  schemaname, 
  tablename, 
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
  n_dead_tup,
  n_live_tup
FROM pg_stat_user_tables
ORDER BY n_dead_tup DESC;
```

**Fix**: Tune autovacuum or run manual `VACUUM`.

### 5. **SSDs Change the Game**

On HDDs, random I/O is so slow that sequential scans were often faster than index scans.

On SSDs, index scans are almost always faster (except for very small tables or when reading >20% of rows).

**Implication**: Don't avoid indexes because of "old wisdom" about random I/O being slow. That was true for HDDs, not SSDs.

---

## Summary

- Databases read and write data in **pages** (8 KB in PostgreSQL, 16 KB in MySQL)
- A query that reads 1 row actually reads at least 1 page (8 KB)
- PostgreSQL uses **heap storage** (unordered pages)
- MySQL InnoDB uses **clustered indexes** (rows stored in primary key order)
- Rows have significant overhead (23 bytes per tuple in PostgreSQL)
- Dead tuples waste space and slow down scans
- Sequential I/O is fast; random I/O is slower (but much less so on SSDs)

Understanding these fundamentals lets you predict query performance before running `EXPLAIN`.
