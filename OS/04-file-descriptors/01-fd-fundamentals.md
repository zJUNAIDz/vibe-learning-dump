# File Descriptors Fundamentals

**The Interface to Everything**

🟢 **Fundamentals**

---

## What is a File Descriptor?

**A file descriptor (fd) is a non-negative integer that represents an open file or I/O resource.**

When your code opens a file, creates a socket, or creates a pipe, the kernel gives you back an integer:

```typescript
import * as fs from 'fs';

const fd = fs.openSync('/tmp/file.txt', 'w');
console.log(fd); // Prints: 3 (or 4, or 5, ...)
```

**That integer is your handle for all future operations on that resource.**

```typescript
fs.writeSync(fd, 'Hello World');  // Write using fd
fs.closeSync(fd);                 // Close using fd
```

---

## File Descriptors are NOT Files

**Common confusion:**

❌ File descriptor is the file  
✅ File descriptor is a **reference** to a kernel object

```
Your Process                    Kernel
┌─────────────────┐            ┌──────────────────────────┐
│ fd 3            │----------->│ Open File Description    │
│ fd 4            │----------->│ Open File Description    │
│ fd 5            │----------->│ Open File Description    │
└─────────────────┘            └──────────────────────────┘
                                           ↓
                               ┌──────────────────────────┐
                               │ Inode (actual file)      │
                               └──────────────────────────┘
                                           ↓
                               ┌──────────────────────────┐
                               │ Data blocks (content)    │
                               └──────────────────────────┘
```

**What the kernel tracks:**
- Which process owns which fds
- What each fd points to (file, socket, pipe)
- Current position in file (for reading/writing)
- Flags (read-only, write-only, append, etc.)

---

## The Standard Three

**Every process starts with three file descriptors already open:**

| fd | Name | Purpose | Typical Destination |
|----|------|---------|---------------------|
| 0 | stdin | Standard input | Keyboard / pipe |
| 1 | stdout | Standard output | Terminal / pipe |
| 2 | stderr | Standard error | Terminal |

```typescript
// These are equivalent:
console.log('Hello');                    // Writes to stdout (fd 1)
fs.writeSync(1, 'Hello\n');              // Same thing

// These are equivalent:
console.error('Error!');                 // Writes to stderr (fd 2)
fs.writeSync(2, 'Error!\n');             // Same thing
```

**Why this matters:**

```bash
# Redirect stdout to file
$ node script.js > output.txt
# fd 1 now points to output.txt instead of terminal

# Redirect stderr to file
$ node script.js 2> errors.txt
# fd 2 now points to errors.txt

# Redirect both
$ node script.js > output.txt 2>&1
# fd 1 to output.txt, fd 2 follows fd 1
```

---

## Opening Files

### open() System Call

**Under the hood of every file operation:**

```c
// C example
int fd = open("/tmp/file.txt", O_WRONLY | O_CREAT, 0644);
if (fd == -1) {
    perror("open failed");
    return;
}
```

**Flags:**
- `O_RDONLY` — Read only
- `O_WRONLY` — Write only
- `O_RDWR` — Read and write
- `O_CREAT` — Create if doesn't exist
- `O_APPEND` — Append to end
- `O_TRUNC` — Truncate to zero length
- `O_NONBLOCK` — Non-blocking mode

### Node.js Examples

```typescript
import * as fs from 'fs';

// Open for writing, create if needed
const fd = fs.openSync('/tmp/file.txt', 'w');

// Open for reading
const fd = fs.openSync('/tmp/file.txt', 'r');

// Open for appending
const fd = fs.openSync('/tmp/file.txt', 'a');

// Open for reading and writing
const fd = fs.openSync('/tmp/file.txt', 'r+');
```

**High-level APIs use open() internally:**

```typescript
// This:
const data = fs.readFileSync('/tmp/file.txt');

// Is roughly equivalent to:
const fd = fs.openSync('/tmp/file.txt', 'r');
const buffer = Buffer.alloc(1024);
fs.readSync(fd, buffer, 0, 1024, 0);
fs.closeSync(fd);
```

---

## File Descriptor Table

**Each process has a file descriptor table:**

```
Process PID 1234:
┌─────────────────────────────┐
│ File Descriptor Table       │
├────┬────────────────────────┤
│ 0  │ → /dev/pts/0 (stdin)   │
│ 1  │ → /dev/pts/0 (stdout)  │
│ 2  │ → /dev/pts/0 (stderr)  │
│ 3  │ → /tmp/file.txt        │
│ 4  │ → socket:[12345]       │
│ 5  │ → pipe:[67890]         │
│ 6  │ → /var/log/app.log     │
│... │                        │
└────┴────────────────────────┘
```

**Kernel assigns lowest available number:**

```typescript
const fd1 = fs.openSync('/tmp/a.txt', 'r'); // Probably 3
const fd2 = fs.openSync('/tmp/b.txt', 'r'); // Probably 4
fs.closeSync(fd1);
const fd3 = fs.openSync('/tmp/c.txt', 'r'); // Reuses 3!
```

**View open file descriptors:**

```bash
$ ls -l /proc/self/fd/
lrwx------ 1 user user 64 Feb 21 12:00 0 -> /dev/pts/0
lrwx------ 1 user user 64 Feb 21 12:00 1 -> /dev/pts/0
lrwx------ 1 user user 64 Feb 21 12:00 2 -> /dev/pts/0
lrwx------ 1 user user 64 Feb 21 12:00 3 -> /tmp/file.txt
```

---

## Everything Uses File Descriptors

### Sockets are File Descriptors

```typescript
import * as net from 'net';

const server = net.createServer((socket) => {
  // socket is a file descriptor under the hood
  socket.write('Hello\n'); // Writes to fd
});

server.listen(3000);
```

**Kernel perspective:**

```
Process PID 1234:
├── fd 3 → socket:[tcp:*:3000] (listening socket)
├── fd 4 → socket:[tcp:127.0.0.1:3000 -> 127.0.0.1:54321] (connection 1)
├── fd 5 → socket:[tcp:127.0.0.1:3000 -> 127.0.0.1:54322] (connection 2)
```

### Pipes are File Descriptors

```typescript
import { spawn} from 'child_process';

const child = spawn('ls', ['-l']);

// child.stdout is a file descriptor (pipe)
child.stdout.on('data', (data) => {
  console.log(data.toString());
});
```

**Kernel perspective:**

```
Parent process:
├── fd 3 → pipe:[read end]

Child process:
├── fd 1 → pipe:[write end]  (stdout redirected to pipe)
```

---

## Reading and Writing

### read() and write()

```c
// C examples
char buffer[1024];

// Read up to 1024 bytes
ssize_t bytes_read = read(fd, buffer, 1024);

// Write bytes
ssize_t bytes_written = write(fd, "Hello", 5);
```

### Node.js Examples

```typescript
import * as fs from 'fs';

const fd = fs.openSync('/tmp/file.txt', 'r');
const buffer = Buffer.alloc(1024);

// Read into buffer
const bytesRead = fs.readSync(fd, buffer, 0, 1024, null);
console.log(`Read ${bytesRead} bytes`);

// Write from buffer
const writeBuffer = Buffer.from('Hello World');
const bytesWritten = fs.writeSync(fd, writeBuffer);
console.log(`Wrote ${bytesWritten} bytes`);

fs.closeSync(fd);
```

### File Position

**Each open file description has a current position:**

```typescript
const fd = fs.openSync('/tmp/file.txt', 'r+');

fs.writeSync(fd, 'ABCDE');  // Writes at position 0
// Position now at 5

fs.writeSync(fd, 'FGH');    // Writes at position 5
// Position now at 8

// Seek to beginning
fs.readSync(fd, buffer, 0, 3, 0); // Read 3 bytes from position 0
// buffer contains: 'ABC'

fs.closeSync(fd);
```

---

## Blocking vs Non-Blocking I/O

### Blocking I/O (Default)

```typescript
// This blocks the thread until data is available
const data = fs.readFileSync('/tmp/file.txt');
console.log('Read complete');
```

**What happens:**
1. Thread calls `read()`
2. Kernel checks if data available
3. If not, kernel **suspends the thread** (state: `S`)
4. When data arrives, kernel wakes thread
5. Thread resumes, `read()` returns

**Problem: Thread is blocked, can't do anything else.**

### Non-Blocking I/O

```c
// Set fd to non-blocking mode
int flags = fcntl(fd, F_GETFL, 0);
fcntl(fd, F_SETFL, flags | O_NONBLOCK);

// Read returns immediately
ssize_t bytes = read(fd, buffer, 1024);
if (bytes == -1 && errno == EAGAIN) {
    // No data available right now, try again later
}
```

**Node.js handles this automatically with async APIs:**

```typescript
// Non-blocking (using libuv thread pool)
fs.readFile('/tmp/file.txt', (err, data) => {
  console.log('Read complete');
});

console.log('Not blocked!');  // Executes immediately
```

---

## select, poll, epoll: Scalable I/O

**Problem: How to wait for activity on thousands of file descriptors?**

### Old Way: select()

```c
fd_set readfds;
FD_ZERO(&readfds);
FD_SET(fd1, &readfds);
FD_SET(fd2, &readfds);
FD_SET(fd3, &readfds);

// Block until any fd has data
select(max_fd + 1, &readfds, NULL, NULL, NULL);

if (FD_ISSET(fd1, &readfds)) {
    // fd1 has data
}
```

**Limitations:**
- Limited to 1024 file descriptors
- Must iterate through all fds to check readiness
- O(n) performance

### Better: poll()

```c
struct pollfd fds[3];
fds[0].fd = fd1;
fds[0].events = POLLIN;  // Wait for readable

fds[1].fd = fd2;
fds[1].events = POLLIN;

fds[2].fd = fd3;
fds[2].events = POLLIN;

// Block until any fd ready
poll(fds, 3, -1);

// Check which fds are ready
for (int i = 0; i < 3; i++) {
    if (fds[i].revents & POLLIN) {
        // fds[i].fd is ready
    }
}
```

**Better but:**
- Still O(n) to check readiness
- Still passes entire array to kernel each call

### Best: epoll() (Linux)

```c
// Create epoll instance
int epfd = epoll_create1(0);

// Add fds to monitor
struct epoll_event ev;
ev.events = EPOLLIN;
ev.data.fd = fd1;
epoll_ctl(epfd, EPOLL_CTL_ADD, fd1, &ev);

epoll_ctl(epfd, EPOLL_CTL_ADD, fd2, &ev);
epoll_ctl(epfd, EPOLL_CTL_ADD, fd3, &ev);

// Wait for events
struct epoll_event events[10];
int nfds = epoll_wait(epfd, events, 10, -1);

// Only ready fds returned!
for (int i = 0; i < nfds; i++) {
    int ready_fd = events[i].data.fd;
    // Handle ready_fd
}
```

**Why epoll is fast:**
- O(1) to add/remove fds
- O(1) to wait (returns only ready fds)
- Scales to millions of file descriptors

**This is how Node.js and nginx scale to handle 10,000+ connections on one thread.**

---

## "Too Many Open Files" Error

**Every process has a limit on open file descriptors:**

```bash
$ ulimit -n
1024  # Default soft limit
```

**What happens when you exceed it:**

```typescript
// Open 2000 files
for (let i = 0; i < 2000; i++) {
  try {
    const fd = fs.openSync(`/tmp/file${i}.txt`, 'w');
  } catch (err) {
    console.error(err.message);
    // EMFILE: too many open files
    break;
  }
}
```

### Common Causes

1. **File descriptor leaks**
   ```typescript
   // BAD: Never closed
   function leak() {
     const fd = fs.openSync('/tmp/file.txt', 'r');
     // ... use fd ...
     // Forgot to close!
   }
   
   for (let i = 0; i < 2000; i++) {
     leak(); // Eventually hits limit
   }
   ```

2. **Too many open connections**
   ```typescript
   // Web server handling 10,000 connections
   // Each connection = 1 fd
   // Default limit (1024) exceeded
   ```

### Solutions

1. **Always close file descriptors**
   ```typescript
   const fd = fs.openSync('/tmp/file.txt', 'r');
   try {
     // Use fd
   } finally {
     fs.closeSync(fd);  // Always close
   }
   ```

2. **Increase limits**
   ```bash
   # Temporary (current shell)
   $ ulimit -n 65536
   
   # Permanent (/etc/security/limits.conf)
   * soft nofile 65536
   * hard nofile 65536
   ```

3. **Use connection pooling**
   ```typescript
   // Database connection pool
   const pool = new Pool({ max: 20 }); // Limit concurrent connections
   ```

### Debugging

**Find open file descriptors:**

```bash
# List all open fds for process
$ ls /proc/1234/fd/ | wc -l
523

# See what they point to
$ ls -l /proc/1234/fd/
lr-x------ 1 user user 64 Feb 21 12:00 0 -> /dev/null
l-wx------ 1 user user 64 Feb 21 12:00 1 -> /var/log/app.log
l-wx------ 1 user user 64 Feb 21 12:00 2 -> /var/log/app.log
lrwx------ 1 user user 64 Feb 21 12:00 3 -> socket:[12345]
lrwx------ 1 user user 64 Feb 21 12:00 4 -> socket:[12346]
...

# Count fd types
$ ls -l /proc/1234/fd/ | grep socket | wc -l
500  # 500 open sockets
```

**Use lsof:**

```bash
# List open files for process
$ lsof -p 1234

# Count open files
$ lsof -p 1234 | wc -l

# Find who has a file open
$ lsof /tmp/file.txt
```

---

## Pipes and Redirection

### Creating Pipes

```typescript
import { spawn } from 'child_process';

const child = spawn('grep', ['error'], {
  stdio: ['pipe', 'pipe', 'inherit']
  //      stdin   stdout  stderr
});

// Write to child's stdin (pipe)
child.stdin.write('Some error message\n');
child.stdin.write('Normal log\n');
child.stdin.end();

// Read from child's stdout (pipe)
child.stdout.on('data', (data) => {
  console.log(`Matched: ${data}`);
});
```

**Kernel creates two fds:**

```
Parent process:
├── fd X → pipe:[write end]   (child's stdin)
├── fd Y → pipe:[read end]    (child's stdout)

Child process (grep):
├── fd 0 → pipe:[read end]   (stdin from parent)
├── fd 1 → pipe:[write end]  (stdout to parent)
```

### Shell Pipe

```bash
$ ls -l | grep ".txt"
```

**What the shell does:**

1. Creates a pipe
2. Forks to create two processes
3. First process: `stdout` → pipe write end
4. Second process: `stdin` → pipe read end
5. Both processes execute

---

## Key Takeaways

1. **File descriptor is an integer referencing a kernel object**
2. **0, 1, 2 are stdin, stdout, stderr**
3. **Sockets, pipes, files all use file descriptors**
4. **Blocking I/O suspends thread; non-blocking returns immediately**
5. **epoll() is how Node.js/nginx handle thousands of connections**
6. **"Too many open files" means fd limit exceeded**
7. **Always close file descriptors to avoid leaks**

---

## What's Next

- [Module 05: Networking Inside Linux](../05-networking-fundamentals/)
- [Module 11: Performance & Debugging (lsof, strace)](../11-debugging/)

---

**Next:** [Module 05: Networking Fundamentals](../05-networking-fundamentals/01-networking-basics.md)
