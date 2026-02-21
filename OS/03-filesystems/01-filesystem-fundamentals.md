# Filesystems and Storage

**"Everything is a File" — What That Really Means**

🟢 **Fundamentals** | 🟡 **Intermediate**

---

## What "Everything is a File" Means

In Linux, the phrase "everything is a file" doesn't mean everything is stored on disk. It means:

**"Everything uses the file interface."**

The same operations work on:
- Regular files on disk
- Devices (`/dev/sda`, `/dev/null`)
- Pipes
- Sockets
- Processes (`/proc/[pid]/`)
- System information (`/sys/`)

```c
// Same API for all:
int fd = open("/dev/urandom", O_RDONLY);  // Device
int fd = open("/tmp/file.txt", O_RDONLY); // Regular file
int fd = open("/proc/self/stat", O_RDONLY); // Process info

// All use:
read(fd, buffer, size);
close(fd);
```

**Why this matters:**
- Uniform interface
- Tools work on diverse resources (`cat /proc/cpuinfo`)
- Redirection just works (`program < input.txt > output.txt`)

---

## The Filesystem Hierarchy

Linux has a **single unified filesystem tree**, not drive letters.

```
/                           Root of everything
├── bin/                    Essential binaries (ls, cat, bash)
├── boot/                   Kernel, bootloader files
├── dev/                    Device files
├── etc/                    System configuration
├── home/                   User directories
│   └── user/               Your files
├── lib/                    Shared libraries
├── mnt/                    Temporary mount points
├── opt/                    Optional software
├── proc/                   Process information (virtual)
├── root/                   Root user's home
├── run/                    Runtime data (PIDs, sockets)
├── sbin/                   System binaries (root)
├── srv/                    Service data
├── sys/                    Hardware/kernel info (virtual)
├── tmp/                    Temporary files (cleared on boot)
├── usr/                    User programs and data
│   ├── bin/                User binaries
│   ├── lib/                Libraries
│   └── local/              Locally installed software
└── var/                    Variable data
    ├── log/                Log files
    ├── tmp/                Persistent temporary files
    └── run/                Runtime files
```

**Key principles:**
1. Everything starts at `/` (root)
2. No `C:\`, `D:\` — all storage unified
3. External drives mounted into tree (e.g., `/mnt/usb`)
4. Network shares mounted into tree (e.g., `/mnt/nfs/share`)

---

## Inodes: What Files Really Are

**A file is not its name. A file is an inode.**

### What's in an Inode?

```
Inode 12345:
┌─────────────────────────────┐
│ File metadata:              │
│  - Type (file/dir/link)     │
│  - Permissions (rwxr-xr-x)  │
│  - Owner (UID: 1000)        │
│  - Group (GID: 1000)        │
│  - Size (bytes)             │
│  - Timestamps:              │
│    • atime (access)         │
│    • mtime (modification)   │
│    • ctime (change)         │
│  - Link count               │
│  - Pointers to data blocks  │
└─────────────────────────────┘
```

**NOT in the inode:**
- Filename (stored in directory entries)
- File content (stored in data blocks)

### Filenames Are Just Pointers

```
Directory entries point to inodes:

/home/user/
├── file.txt  → inode 12345
├── backup.txt → inode 12345  (same inode!)
└── other.txt  → inode 67890

Inode 12345 → Data blocks: [4000, 4001, 4002]
Inode 67890 → Data blocks: [5000]
```

**Multiple names can point to the same inode (hard links).**

---

## Hard Links vs Symbolic Links

### Hard Links

**Multiple directory entries pointing to the same inode:**

```bash
$ echo "Hello" > original.txt
$ ln original.txt hardlink.txt
$ ls -li
12345 -rw-r--r-- 2 user user 6 Feb 21 12:00 original.txt
12345 -rw-r--r-- 2 user user 6 Feb 21 12:00 hardlink.txt
#^^^^                     ^
#Same inode              Link count = 2
```

**Characteristics:**
- Share the same inode
- Same permissions, size, content
- Modifying one modifies both (they're the same file)
- Deleting one doesn't delete the file (link count decremented)
- File only deleted when link count reaches 0
- **Cannot cross filesystem boundaries** (inode numbers are per-filesystem)
- **Cannot link directories** (would create cycles)

```bash
$ echo "Modified" > original.txt
$ cat hardlink.txt
Modified  # Same file!

$ rm original.txt
$ cat hardlink.txt
Modified  # Still exists! Inode link count now 1

$ rm hardlink.txt
# Now link count = 0, inode freed, data blocks freed
```

### Symbolic Links (Symlinks)

**A special file containing a path to another file:**

```bash
$ ln -s /home/user/original.txt symlink.txt
$ ls -li
12345 -rw-r--r-- 1 user user  6 Feb 21 12:00 original.txt
67890 lrwxrwxrwx 1 user user 24 Feb 21 12:01 symlink.txt -> /home/user/original.txt
#^^^^                     ^
#Different inode         'l' = symbolic link
```

**Characteristics:**
- Has its own inode (different from target)
- Contains path as data
- Can cross filesystem boundaries
- Can link directories
- Can point to non-existent files (broken link)
- Deleting target breaks the symlink

```bash
$ rm original.txt
$ cat symlink.txt
cat: symlink.txt: No such file or directory
# Symlink now broken

$ ls -l symlink.txt
lrwxrwxrwx 1 user user 24 Feb 21 12:01 symlink.txt -> /home/user/original.txt
# Symlink still exists, but target is gone
```

**When to use what:**

| Use Case | Hard Link | Symlink |
|----------|-----------|---------|
| Backup without duplication | ✅ | ❌ |
| Link across filesystems | ❌ | ✅ |
| Link directories | ❌ | ✅ |
| Point to non-existent path | ❌ | ✅ |
| Transparent to applications | ✅ | ❌ (may follow or not) |

---

## File Permissions Beyond rwx

### Standard Permissions

```bash
$ ls -l file.txt
-rw-r--r-- 1 user group 1024 Feb 21 12:00 file.txt
^^^^^^^^^
│││││││││
││││││││└─ Others execute
│││││││└── Others write
││││││└─── Others read
│││││└──── Group execute
││││└───── Group write
│││└────── Group read
││└─────── Owner execute
│└──────── Owner write
└───────── Owner read
```

**First character:**
- `-` = regular file
- `d` = directory
- `l` = symbolic link
- `c` = character device
- `b` = block device
- `s` = socket
- `p` = pipe (FIFO)

### Special Permissions

**Setuid (s):**
```bash
$ ls -l /usr/bin/passwd
-rwsr-xr-x 1 root root 68208 May 28  2020 /usr/bin/passwd
   ^
   Setuid bit
```

- Runs with file owner's permissions, not caller's
- `passwd` runs as root so it can modify `/etc/shadow`
- Security risk if misused

**Setgid (s):**
- On file: Runs with file group's permissions
- On directory: New files inherit directory's group

**Sticky bit (t):**
```bash
$ ls -ld /tmp
drwxrwxrwt 20 root root 4096 Feb 21 12:00 /tmp
        ^
        Sticky bit
```

- On directory: Only owner can delete their files
- `/tmp` is world-writable but you can't delete others' files

---

## Mount Points and the Unified Tree

**Filesystems are mounted into a single tree:**

```bash
$ df -h
Filesystem      Size  Used Avail Use% Mounted on
/dev/sda1        50G   20G   28G  42% /
/dev/sda2       100G   60G   35G  64% /home
/dev/sdb1       500G  200G  275G  42% /mnt/data
tmpfs           8.0G  1.2G  6.8G  15% /tmp
```

```
/ (root filesystem on /dev/sda1)
├── home/ (separate filesystem on /dev/sda2)
│   └── user/
│       └── documents/
├── mnt/
│   └── data/ (separate filesystem on /dev/sdb1)
│       └── backups/
└── tmp/ (tmpfs - in RAM!)
    └── temp_files
```

**Crossing mount points:**

```bash
$ cd /home/user
$ pwd
/home/user
# This is on /dev/sda2

$ cd /
$ pwd
/
# This is on /dev/sda1

# Transparent to applications!
```

**Why this matters:**
- Performance varies by filesystem
- Disk space isolated per mount
- Can unmount busy filesystems

---

## Special Filesystems

### /proc — Process Information

**Not a real filesystem. Generated on-the-fly by kernel.**

```bash
$ ls /proc
1/  2/  3/  ... cpuinfo  meminfo  version
```

**What's available:**

```bash
# CPU information
$ cat /proc/cpuinfo

# Memory information
$ cat /proc/meminfo

# Kernel command line
$ cat /proc/cmdline

# Process information
$ cat /proc/1234/cmdline  # Command that started process
$ ls /proc/1234/fd/       # Open file descriptors
$ cat /proc/1234/status   # Process status
$ cat /proc/1234/maps     # Memory mappings
```

**Why developers care:**
- Debugging tools read /proc (ps, top, htop)
- You can read it directly for monitoring
- Container systems use /proc for isolation

### /sys — Hardware/Kernel Info

**Interface to kernel and device information:**

```bash
$ ls /sys
block/  bus/  class/  dev/  devices/  firmware/  fs/  kernel/  module/  power/

# Network interface information
$ cat /sys/class/net/eth0/address
00:0c:29:6a:4e:11

# Block device information
$ cat /sys/block/sda/size
976773168  # Size in 512-byte sectors

# CPU frequency
$ cat /sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq
2400000  # 2.4 GHz
```

### /dev — Devices

**Device files allow programs to communicate with hardware:**

```bash
$ ls -l /dev/
crw-rw-rw- 1 root root 1, 3 Feb 21 09:00 null      # Discard all writes
crw-rw-rw- 1 root root 1, 8 Feb 21 09:00 random    # Random number generator
crw-rw-rw- 1 root root 1, 9 Feb 21 09:00 urandom   # Better random
brw-rw---- 1 root disk 8, 0 Feb 21 09:00 sda       # First SATA disk
brw-rw---- 1 root disk 8, 1 Feb 21 09:00 sda1      # First partition
```

- `c` = character device (byte-by-byte access)
- `b` = block device (block access, e.g., disks)

**Common device files:**

```bash
# Null device (discard)
$ echo "ignored" > /dev/null

# Random number generator
$ head -c 16 /dev/urandom | base64
qK8hJxQY5vZnP8Fw==

# Disk access (requires root)
$ sudo dd if=/dev/sda of=disk.img bs=1M count=100
```

---

## OverlayFS and Container Layers

**How Docker creates efficient container images:**

```
Docker Image Layers (read-only):
┌─────────────────────────────┐
│ Layer 3: Application code   │
├─────────────────────────────┤
│ Layer 2: Dependencies       │
├─────────────────────────────┤
│ Layer 1: Base OS            │
└─────────────────────────────┘
         ↓
    OverlayFS
         ↓
┌─────────────────────────────┐
│ Container view (unified)    │
│ /app/server.js              │
│ /usr/bin/node               │
│ /lib/libc.so.6              │
└─────────────────────────────┘
         ↑
┌─────────────────────────────┐
│ Writable layer (per container)│
└─────────────────────────────┘
```

**How it works:**

1. Multiple read-only layers stacked
2. Writable layer on top (per container)
3. Reads: Check upper layers first, then lower layers
4. Writes: Write to writable layer (copy-on-write)
5. Deletes: Create "whiteout" file in upper layer

**Why this matters:**
- Containers share base layers (saves disk space)
- Starting containers is fast (no copying)
- Image updates are incremental
- Understanding this explains Docker behavior

```bash
# See Docker layers
$ docker history node:18
IMAGE          CREATED       SIZE
e8c92b4e8e21   2 weeks ago   10MB    CMD ["node"]
<missing>      2 weeks ago   50MB    COPY app/ /app
<missing>      3 weeks ago   200MB   RUN apt-get install
<missing>      1 month ago   100MB   FROM debian:bullseye
```

---

## Filesystem Types

### ext4 (Most Common)

```bash
$ sudo mkfs.ext4 /dev/sdb1
$ mount /dev/sdb1 /mnt/data
```

**Characteristics:**
- Journaling (crash recovery)
- Large file support (16 TB)
- Mature, stable
- Default on many Linux distros

### XFS (High Performance)

**Characteristics:**
- Excellent for large files
- Parallel I/O
- Used by Red Hat/Fedora by default
- Good for databases

### Btrfs (Modern)

**Characteristics:**
- Copy-on-write
- Snapshots
- Built-in compression
- Self-healing (checksums)
- Still maturing

### tmpfs (In-Memory)

```bash
$ mount -t tmpfs -o size=1G tmpfs /mnt/ramdisk
```

**Characteristics:**
- Stored in RAM, not disk
- Extremely fast
- Lost on reboot
- Used for `/tmp` on many systems

---

## I/O and Caching

### Write Behavior

**Write to file doesn't immediately hit disk:**

```typescript
import * as fs from 'fs';

fs.writeFileSync('/tmp/data.txt', 'Hello'); // Returns immediately
// Data in page cache, not on disk yet!

// Force write to disk:
const fd = fs.openSync('/tmp/data.txt', 'w');
fs.writeSync(fd, 'Hello');
fs.fsyncSync(fd);  // Now on disk
fs.closeSync(fd);
```

**Why buffering exists:**
- Disks are slow (milliseconds)
- RAM is fast (nanoseconds)
- Kernel batches writes for efficiency

### Reading

```typescript
import * as fs from 'fs';

// First read: from disk (slow)
const data1 = fs.readFileSync('/var/log/syslog');

// Second read: from page cache (fast)
const data2 = fs.readFileSync('/var/log/syslog');
```

**Page cache accelerates repeated reads.**

---

## Production Scenarios

### Scenario 1: "Disk Full" but Space Available

```bash
$ df -h
Filesystem      Size  Used Avail Use% Mounted on
/dev/sda1        50G   30G   20G  60% /

$ touch /tmp/newfile
touch: cannot touch '/tmp/newfile': No space left on device

$ # What? 20GB available!
```

**Cause: Inode exhaustion**

```bash
$ df -i
Filesystem      Inodes  IUsed  IFree IUse% Mounted on
/dev/sda1      3276800 3276800     0  100% /

$ # Out of inodes! Too many small files.
```

**Fix: Delete many files, or recreate filesystem with more inodes.**

### Scenario 2: Deleted File Still Uses Space

```bash
$ df -h /
Filesystem      Size  Used Avail Use% Mounted on
/dev/sda1        50G   45G    5G  90% /

$ rm /var/log/huge.log  # Delete 10GB file

$ df -h /
Filesystem      Size  Used Avail Use% Mounted on
/dev/sda1        50G   45G    5G  90% /

$ # Space not freed!
```

**Cause: File still open by a process**

```bash
$ lsof | grep deleted
process  1234 user  3w  REG  8,1  10G  12345 /var/log/huge.log (deleted)
```

**Fix: Restart the process (or `> /proc/1234/fd/3` to truncate).**

---

## Key Takeaways

1. **Inodes are files; filenames are just pointers**
2. **Hard links share inodes; symlinks are separate files**
3. **Linux has unified filesystem tree (no drive letters)**
4. **Special filesystems (/proc, /sys, /dev) are kernel interfaces**
5. **OverlayFS enables efficient container layers**
6. **Writes are buffered (page cache); fsync() forces disk write**
7. **Can run out of inodes even with disk space available**
8. **Deleted files hold space if still open**

---

## What's Next

- [Module 04: File Descriptors & I/O](../04-file-descriptors/)
- [Module 07: Namespaces & cgroups (Container Filesystems)](../07-containers/)

---

**Next:** [Module 04: File Descriptors & I/O](../04-file-descriptors/01-fd-fundamentals.md)
