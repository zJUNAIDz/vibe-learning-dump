# Failures and Recovery

## Introduction

Your database server crashes. Power outage. Kernel panic. Disk failure. Network partition.

**What happens to your data?**

If you committed a transaction:
```sql
BEGIN;
INSERT INTO orders (user_id, total) VALUES (1, 100);
COMMIT;
```

**Will this order survive the crash?**

The answer depends on how databases implement **durability** (the "D" in ACID).

This chapter explains:
- How databases ensure data survives crashes
- What happens during recovery
- Common mistakes that cause data loss

---

## The Durability Problem

When you run:
```sql
INSERT INTO orders (user_id, total) VALUES (1, 100);
COMMIT;
```

**Where is the data physically stored?**

1. **In memory** (buffer cache): Fast but volatile (lost on crash)
2. **On disk**: Slow but durable (survives crash)

**The challenge**: Writing to disk is slow (~0.1ms on SSD). If every `COMMIT` waited for disk writes, throughput would be terrible.

**The solution**: Write-Ahead Logging (WAL).

---

## Write-Ahead Logging (WAL)

**WAL** is a sequential log of all changes to the database.

### How It Works

**Before modifying data pages**, the database:
1. Writes a description of the change to the WAL (a sequential log file)
2. Calls `fsync` to flush the WAL to disk
3. Returns "commit successful" to the client

**Later** (asynchronously):
4. Applies the change to the actual data pages
5. Writes the modified data pages to disk

```
┌───────────────────────────────────────┐
│          Buffer Cache                 │
│  (data pages in memory)               │
└───────────────┬───────────────────────┘
                │ Async writes
                ▼
┌───────────────────────────────────────┐
│         Data Files (heap)             │
│  (on disk, updated lazily)            │
└───────────────────────────────────────┘

┌───────────────────────────────────────┐
│              WAL                      │
│  (on disk, updated immediately)       │
└───────────────────────────────────────┘
```

**Key insight**: The WAL is **sequential** (fast to write). Data pages are **random** (slow to write).

**Result**: Commits are fast (only need to flush WAL), but durability is guaranteed (WAL contains all changes).

---

## WAL Entries

Each change is logged as a WAL entry:

```
WAL Entry 1: XACT_START (transaction 1234)
WAL Entry 2: INSERT INTO orders (id, user_id, total) VALUES (1, 1, 100)
WAL Entry 3: XACT_COMMIT (transaction 1234)
```

**Properties**:
- WAL entries are **append-only** (fast writes)
- Each entry has a **sequence number** (LSN: Log Sequence Number)
- WAL is flushed to disk on `COMMIT`

---

## Crash Recovery

**Scenario**: The database crashes mid-operation.

**What happens on restart?**

### Recovery Process

1. **Start from the last checkpoint**
   - A checkpoint is a known-good state where all dirty pages were flushed to disk
   - The WAL records the LSN of the last checkpoint

2. **Replay WAL entries**
   - Read WAL from the checkpoint to the end
   - Reapply all committed changes
   - Roll back any uncommitted transactions

3. **Bring the database online**
   - Once WAL replay is complete, the database is consistent

```
Timeline:

10:00 AM - Checkpoint (LSN 1000)
10:01 AM - INSERT ... (LSN 1010)
10:02 AM - UPDATE ... (LSN 1020)
10:03 AM - COMMIT (LSN 1030)
10:04 AM - INSERT ... (LSN 1040)  ← Transaction not committed
10:05 AM - CRASH!

Recovery:
- Start from checkpoint (LSN 1000)
- Replay LSN 1010, 1020, 1030 (committed)
- Discard LSN 1040 (not committed)
- Database is consistent at LSN 1030
```

**Guarantee**: All **committed** transactions survive. Uncommitted transactions are lost.

---

## fsync and Durability Costs

`fsync` ensures data is physically written to disk (not just cached by the OS).

### Without fsync

```sql
COMMIT;
-- Data is written to the OS page cache (but not disk)
-- If the server crashes, data is lost
```

### With fsync

```sql
COMMIT;
-- Data is written to the OS page cache
-- fsync() is called to flush to disk
-- Only then is "commit successful" returned
```

**Cost**: 
- SSD: ~0.1 ms per fsync
- HDD: ~10 ms per fsync

**Throughput limit**:
- SSD: ~10,000 commits/sec
- HDD: ~100 commits/sec

### Group Commit

Modern databases batch multiple commits into a single `fsync`:

```
Transaction A: COMMIT (waits for fsync)
Transaction B: COMMIT (waits for fsync)
Transaction C: COMMIT (waits for fsync)

Database: fsync() once (flushes all three)
```

**Result**: 3 commits in the time of 1 fsync.

**Throughput**: ~100,000 commits/sec on modern hardware.

---

## What Happens on a Crash

### Scenario 1: Crash Before COMMIT

```sql
BEGIN;
INSERT INTO orders (user_id, total) VALUES (1, 100);
-- CRASH! (before COMMIT)
```

**Result**: The transaction is lost. The WAL has no `XACT_COMMIT` entry, so recovery discards it.

**Application impact**: The client sees an error ("connection lost"). The app must retry.

### Scenario 2: Crash After COMMIT (WAL Flushed)

```sql
BEGIN;
INSERT INTO orders (user_id, total) VALUES (1, 100);
COMMIT;
-- CRASH! (after COMMIT)
```

**Result**: The transaction is durable. During recovery:
1. WAL is replayed
2. The `INSERT` is reapplied
3. The order exists after recovery

**Application impact**: None. The data is safe.

### Scenario 3: Crash After COMMIT (WAL Not Flushed)

```sql
BEGIN;
INSERT INTO orders (user_id, total) VALUES (1, 100);
COMMIT;
-- CRASH! (WAL was written to OS cache but not flushed to disk)
```

**Result**: The transaction is **lost** if the OS cache is lost (e.g., power failure).

**But**: PostgreSQL calls `fsync` before returning "commit successful," so this scenario only happens if:
- `fsync = off` (dangerous setting!)
- Disk lies about fsync (some SSDs cache writes)

---

## Checkpoints

A **checkpoint** is a point where all dirty pages in the buffer cache are flushed to disk.

**Purpose**: Limit the amount of WAL that needs to be replayed during recovery.

**Without checkpoints**: Recovery must replay the entire WAL (could be gigabytes, taking minutes).

**With checkpoints**: Recovery only replays WAL since the last checkpoint.

### Checkpoint Frequency

PostgreSQL triggers checkpoints based on:
- **Time**: Every `checkpoint_timeout` seconds (default: 5 minutes)
- **WAL size**: After `max_wal_size` bytes (default: 1 GB)

**Configuration** (`postgresql.conf`):
```conf
checkpoint_timeout = 5min
max_wal_size = 1GB
```

**Trade-off**:
- Frequent checkpoints → faster recovery, but more I/O overhead
- Infrequent checkpoints → slower recovery, but less I/O overhead

---

## Recovery Time

**How long does recovery take?**

**Factors**:
1. **Amount of WAL**: More WAL = longer recovery
2. **Disk speed**: Faster disk = faster WAL replay
3. **Checkpoint frequency**: More checkpoints = less WAL to replay

**Typical recovery time**: Seconds to minutes (depending on WAL size).

**Example**:
- 1 GB of WAL
- SSD (500 MB/s sequential read)
- Recovery time: ~2 seconds

**If recovery takes too long**: Increase checkpoint frequency:
```conf
checkpoint_timeout = 1min
max_wal_size = 512MB
```

---

## Data Loss Scenarios (and How to Avoid Them)

### 1. **Disk Corruption**

**Scenario**: The disk develops bad sectors.

**Impact**: Data pages or WAL entries may be corrupted.

**Detection**: PostgreSQL checksums pages (enabled by default in PostgreSQL 12+).

**Recovery**: Restore from backups.

**Prevention**:
- Use RAID (redundancy)
- Use replicas (high availability)
- Take regular backups

### 2. **Operator Error**

**Scenario**: You accidentally run:
```sql
DROP TABLE orders;
COMMIT;
```

**Impact**: All orders are gone.

**Recovery**: Restore from backups (no other option).

**Prevention**:
- Use `BEGIN` before dangerous operations
- Require `WHERE` clauses on `DELETE`
- Use audit logs

### 3. **Disk Full**

**Scenario**: WAL fills up the disk.

**Impact**: Database stops accepting writes:
```
ERROR: could not extend file "pg_wal/000000010000000000000042": No space left on device
```

**Recovery**: Delete old WAL files or add disk space.

**Prevention**: Monitor disk usage. Set up alerts at 80% full.

### 4. **fsync = off**

**Scenario**: You disable `fsync` for "performance":
```conf
fsync = off
```

**Impact**: Commits don't force WAL to disk. If the OS crashes, **committed transactions are lost**.

**Recovery**: Restore from backups (no WAL to replay).

**Prevention**: Never set `fsync = off` in production. Use group commit instead.

### 5. **SSD Lies About fsync**

**Scenario**: Some SSDs have volatile write caches and ignore `fsync`.

**Impact**: Even with `fsync = on`, data can be lost on power failure.

**Detection**: Check SSD specifications (look for "power-loss protection").

**Prevention**: Use enterprise SSDs with power-loss protection.

---

## WAL Archiving and Point-in-Time Recovery (PITR)

**WAL archiving**: Continuously copying WAL files to a backup location.

**Use case**: Restore the database to any point in time (e.g., 5 minutes before the `DROP TABLE` accident).

### Setup (PostgreSQL)

**1. Enable WAL archiving** (`postgresql.conf`):
```conf
wal_level = replica
archive_mode = on
archive_command = 'cp %p /mnt/backup/wal/%f'
```

**2. Take a base backup**:
```bash
pg_basebackup -D /mnt/backup/base -Fp -Xs -P
```

**3. Restore to a specific point in time**:
```bash
# Copy base backup
cp -r /mnt/backup/base /var/lib/postgresql/data

# Create recovery.conf
echo "restore_command = 'cp /mnt/backup/wal/%f %p'" > recovery.conf
echo "recovery_target_time = '2024-01-15 14:30:00'" >> recovery.conf

# Start PostgreSQL (it will replay WAL until the target time)
pg_ctl start
```

**Result**: The database is restored to exactly 14:30:00 (before the accident).

---

## Backup Strategies

### 1. **Logical Backup (pg_dump)**

**How**: Exports data as SQL statements or CSV.

```bash
pg_dump mydb > backup.sql
```

**Pros**:
- Portable (can restore to different PostgreSQL versions)
- Easy to understand

**Cons**:
- Slow (reads entire database)
- Not suitable for large databases (>100 GB)

### 2. **Physical Backup (pg_basebackup)**

**How**: Copies the entire data directory.

```bash
pg_basebackup -D /mnt/backup -Fp -Xs -P
```

**Pros**:
- Fast (copies files directly)
- Suitable for large databases

**Cons**:
- Must restore to the same PostgreSQL version

### 3. **Continuous Archiving (PITR)**

**How**: Combines base backup + WAL archiving.

**Pros**:
- Can restore to any point in time
- Minimal data loss (only the last few seconds)

**Cons**:
- Complex setup

### 4. **Replicas as Backups**

**How**: Maintain a read replica. If primary fails, promote the replica.

**Pros**:
- Near-zero downtime
- No data loss (if synchronous replication)

**Cons**:
- Doesn't protect against operator error (e.g., `DROP TABLE` replicates to replica!)

**Best practice**: Combine replicas + PITR backups.

---

## Postgres vs MySQL Recovery

| Feature                | PostgreSQL (WAL)              | MySQL InnoDB (Redo Log)      |
|------------------------|-------------------------------|------------------------------|
| Durability mechanism   | Write-Ahead Logging (WAL)     | Redo log + binlog            |
| Crash recovery         | Replay WAL                    | Replay redo log              |
| Point-in-time recovery | Yes (WAL archiving)           | Yes (binlog)                 |
| Checkpoints            | Yes                           | Yes                          |
| fsync enforcement      | Yes (fsync = on)              | Yes (innodb_flush_log_at_trx_commit = 1) |

**Key difference**: PostgreSQL uses WAL for both recovery and replication. MySQL uses redo log for recovery and binlog for replication.

---

## Monitoring and Alerting

### 1. **WAL Size**

```sql
SELECT pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), '0/0'));
```

**Alert** if WAL size > 2 GB (checkpoints may be too infrequent).

### 2. **Checkpoint Stats**

```sql
SELECT * FROM pg_stat_bgwriter;
```

**Look for**:
- `checkpoints_req`: Forced checkpoints (too many = increase `max_wal_size`)
- `checkpoints_timed`: Scheduled checkpoints

### 3. **Disk Space**

```bash
df -h /var/lib/postgresql
```

**Alert** if disk usage > 80%.

### 4. **Backup Age**

Check when the last backup was taken. Alert if > 24 hours old.

---

## Practical Takeaways

### 1. **Never Disable fsync in Production**

```conf
fsync = on  # Always!
```

Disabling fsync risks data loss on crashes.

### 2. **Take Regular Backups**

Automate backups (daily or hourly):
```bash
0 2 * * * pg_dump mydb | gzip > /mnt/backup/mydb-$(date +\%Y\%m\%d).sql.gz
```

Test restoring from backups regularly.

### 3. **Use WAL Archiving for PITR**

Enables recovery to any point in time.

### 4. **Monitor Disk Space**

WAL can fill up the disk quickly under heavy write load.

### 5. **Tune Checkpoints for Your Workload**

- Write-heavy: Increase `max_wal_size` (reduce checkpoint overhead)
- Recovery-critical: Decrease `checkpoint_timeout` (faster recovery)

### 6. **Use Replicas for High Availability**

Replicas provide near-zero downtime failover.

### 7. **Test Disaster Recovery**

Simulate a crash and practice restoring from backups. Don't wait until production fails.

---

## Summary

- WAL (Write-Ahead Logging) ensures durability
- On COMMIT, WAL is flushed to disk (fsync)
- Crash recovery replays WAL to restore committed transactions
- Checkpoints flush dirty pages to disk (limit recovery time)
- Backups protect against disk failure and operator error
- WAL archiving enables point-in-time recovery
- Never disable fsync in production
- Monitor WAL size and disk space
- Test disaster recovery regularly

Understanding failures and recovery helps you:
- Design resilient systems
- Minimize data loss
- Recover from disasters confidently
- Avoid rookie mistakes that cause data loss
