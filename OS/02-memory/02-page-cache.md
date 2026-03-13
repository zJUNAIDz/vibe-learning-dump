# The Page Cache

**Linux's Secret Performance Weapon**

🟡 **Intermediate**

---

## Introduction

Have you ever noticed that the second time you read a massive file or run `grep` across a huge codebase, it runs like 100x faster? 

That's not your SSD being built different. That is the **Page Cache** stepping in. The Linux kernel's motto is: "Unused RAM is wasted RAM." Cap. Seriously, why leave RAM empty when it could be caching disk reads?

---

## What is the Page Cache?

When you read a file from the disk, the kernel doesn't just pass the data directly to your application. It:
1. Grabs the block from the disk.
2. Copies it into an empty chunk of RAM (a "page").
3. Hands it to your application.

If you read that file again a minute later, the kernel says: "Hold up, I already have this in RAM." It skips the hard drive entirely and serves the file straight from memory. Lightning fast. ⚡

### Reading `free -h` the Right Way

Usually, people run `free -h`, see that `free` is `100Mi`, and panic. **Don't panic.** Look at `buff/cache`.

```bash
$ free -h
              total        used        free      shared  buff/cache   available
Mem:           16Gi       4.0Gi       500Mi       100Mi        11Gi        11Gi
Swap:         2.0Gi          0B       2.0Gi
```

- **Used**: RAM used by your actual apps (Node, Docker, Slack, etc.).
- **Free**: Literally empty RAM doing absolutely nothing (useless).
- **Buff/Cache**: RAM the kernel hijacked to cache disk files.
- **Available**: How much RAM is *actually* available if your apps need it. The kernel will instantly drop the page cache if an app needs the memory.

**TL;DR:** Your system is perfectly healthy. The kernel is just holding onto 11GB of disk files temporarily to make your workflow faster. Based.

---

## Write-Behind Caching (Dirty Pages)

It works for writing, too! When your app writes to a file, it doesn't instantly write to the actual spinning physical disk. That would be literal trash for performance.

Instead:
1. Your app calls `write()`.
2. The kernel writes the data to the Page Cache in RAM.
3. The kernel tells your app: "Yep, wrote it! You're good to go!" (Lies).
4. The page in RAM is marked as **"Dirty"** (meaning: it differs from the physical disk).
5. Later, in the background, a kernel thread (`pdflush` or `kworker`) silently flushes the dirty pages to the physical disk.

### The Danger 🛑

What happens if you pull the power plug before the dirty pages are flushed? 
**Data loss.** 

If you are writing a database (like Postgres or SQLite) and you need a 100% guarantee that the data is physically on the disk, you must call a special syscall: `fsync()`.
`fsync()` forces the kernel to flush the file's dirty pages to disk immediately and blocks until it's done.

---

## Direct I/O (Bypassing the Cache)

Some apps, like massive databases (Oracle, Postgres) say: "Listen kernel, I have my own optimized caching mechanisms. Your Page Cache is getting in my way, let me talk to the disk directly."

They open files with the `O_DIRECT` flag. This completely bypasses the Page Cache.

---

## To Drop or Not to Drop

Sometimes, during benchmarking, you want to clear the page cache to see how slow a disk read *actually* is.

```bash
# Drop the page cache cleanly (Run as root)
$ sync; echo 1 > /proc/sys/vm/drop_caches
```
*Note: Do not do this in production just because you saw RAM usage get high. Let the kernel cook.*

## Key Takeaways
1. Linux uses all "free" RAM to cache disk I/O.
2. High memory usage in `buff/cache` is a good thing. 
3. Writing to disk actually just writes to RAM initially (Dirty Pages).
4. Databases use `fsync()` to force dirty pages to disk so they don't lose data on a power outage.

---
**Next:** [Memory: Swap and the OOM Killer](03-swap-and-oom.md)
