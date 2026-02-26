# Pipes and Unix Composition

## What Problem This Solves

You see `command1 | command2 | command3` and wonder: Is this running sequentially? Is it creating temp files? How do I debug when a pipeline gives wrong results? Understanding how pipes actually work makes complex pipelines composable and debuggable.

## How People Misunderstand It

1. **"Pipes run commands one after another"** — No. All commands in a pipeline start **simultaneously**. The pipe connects stdout of one to stdin of the next. They run concurrently.
2. **"A pipe is like writing to a temp file"** — Pipes are in-memory kernel buffers (typically 64KB on Linux). No files touch the disk.
3. **"The exit code of a pipeline is whether it 'worked'"** — By default, `$?` is the exit code of the **last** command. If an earlier command fails, you won't know unless you use `pipefail`.

## The Mental Model

A pipe connects the stdout (fd 1) of the left command to the stdin (fd 0) of the right command:

```
cmd1 | cmd2 | cmd3

   cmd1                 cmd2                 cmd3
┌──────────┐         ┌──────────┐         ┌──────────┐
│ stdin  ← │ keyboard│ stdin  ← │─────────│ stdin  ← │─────────
│ stdout → │─────────│ stdout → │─────────│ stdout → │ terminal
│ stderr → │ terminal│ stderr → │ terminal│ stderr → │ terminal
└──────────┘         └──────────┘         └──────────┘
     │                     │                     │
     └── pipe 1 ───────────┘                     │
                           └── pipe 2 ───────────┘
```

Key facts:
- **All three commands start at the same time** (the kernel forks all of them)
- **stderr is NOT piped** — it goes directly to the terminal from every command
- **Pipes buffer** — if cmd2 can't read fast enough, cmd1 blocks when the pipe buffer is full
- **No temp files** — data flows through a kernel buffer in memory

### Pipe Internals

```bash
# The kernel creates a pipe:
# - It's a 64KB buffer (on Linux, check with: cat /proc/sys/fs/pipe-max-size)
# - Writer writes bytes into it
# - Reader reads bytes from it
# - If the buffer is full, the writer blocks (backpressure)
# - If the buffer is empty, the reader blocks (waiting for data)
# - When the writer closes its end, the reader gets EOF

# See the pipe buffer size:
cat /proc/sys/fs/pipe-max-size
# 1048576 (1MB max, 65536 = 64KB default per pipe)
```

## The Unix Philosophy in Action

The pipe is what makes the Unix philosophy work: small tools that do one thing, connected through standard I/O.

```bash
# "How many unique IP addresses accessed my web server today?"
cat /var/log/httpd/access_log | awk '{print $1}' | sort | uniq | wc -l

# Let's read this left to right:
# cat     → streams the file content
# awk     → extracts the first field (IP address) from each line
# sort    → sorts alphabetically (required for uniq)
# uniq    → removes adjacent duplicates
# wc -l   → counts remaining lines
```

This is more efficient than you'd think. All five commands run simultaneously, processing data as it flows through. The file doesn't need to fit in memory — it streams.

### Useless Use of Cat (UUOC)

```bash
# THIS:
cat file | grep pattern

# IS THE SAME AS:
grep pattern file
# Or:
grep pattern < file

# cat is unnecessary when the command can read a file directly.
# Not a huge deal, but it adds an extra process for no reason.

# However, cat IS useful when you need to concatenate multiple files:
cat file1 file2 file3 | sort | uniq
```

## Building Real Pipelines

### Step-by-Step Pipeline Construction

Don't build long pipelines in one shot. Build them incrementally:

```bash
# Step 1: See the raw data
journalctl -u sshd --since today --no-pager | head -20

# Step 2: Extract what you want (inspect output)
journalctl -u sshd --since today --no-pager | grep "Failed" | head -10

# Step 3: Refine extraction
journalctl -u sshd --since today --no-pager | grep "Failed" | awk '{print $11}' | head -10

# Step 4: Aggregate
journalctl -u sshd --since today --no-pager | grep "Failed" | awk '{print $11}' | sort | uniq -c | sort -rn

# Final result: IP addresses with failed SSH attempts, ranked by count
```

### Common Pipeline Patterns

```bash
# Top 10 most common log entries:
journalctl --since today --no-pager | awk '{for(i=6;i<=NF;i++) printf $i" "; print ""}' | sort | uniq -c | sort -rn | head -10

# Disk usage by directory, sorted:
du -sh /var/* 2>/dev/null | sort -rh | head -10

# All listening ports with process names:
ss -tlnp | awk 'NR>1 {print $4, $6}' | sort

# Find large files modified recently:
find /var -type f -mtime -1 -size +10M -ls 2>/dev/null | sort -k7 -rn

# Parse JSON API response:
curl -s "https://api.example.com/data" | jq '.results[] | {name: .name, count: .count}' | head -20
```

## Named Pipes (FIFOs)

Named pipes are pipes that exist as files in the filesystem:

```bash
# Create a named pipe:
mkfifo /tmp/mypipe

# Terminal 1 — write to it (blocks until someone reads):
echo "hello from terminal 1" > /tmp/mypipe

# Terminal 2 — read from it:
cat /tmp/mypipe
# Output: hello from terminal 1

# Cleanup:
rm /tmp/mypipe
```

Named pipes are useful for connecting processes that aren't in the same pipeline:

```bash
# Long-running log processor:
mkfifo /tmp/logpipe
tail -f /var/log/messages > /tmp/logpipe &
grep --line-buffered "error" < /tmp/logpipe
```

## Debugging Pipelines

### Using `tee` to Inspect

`tee` copies stdin to a file AND stdout — like a T-junction in a pipe:

```bash
# See what's flowing through at each stage:
cat /var/log/messages \
    | tee /tmp/stage1.txt \
    | grep "error" \
    | tee /tmp/stage2.txt \
    | awk '{print $5}' \
    | tee /tmp/stage3.txt \
    | sort | uniq -c | sort -rn

# Now check /tmp/stage1.txt, stage2.txt, stage3.txt
# to see what each stage received and produced
```

### Using `pv` for Pipeline Progress

```bash
# Install: sudo dnf install pv
# pv shows progress in a pipeline:
dd if=/dev/zero bs=1M count=100 | pv | gzip > /dev/null
# Shows: speed, total transferred, estimated time
```

### Pipeline Exit Codes

```bash
# Default: $? = exit code of LAST command only
false | true
echo $?     # 0 (true succeeded, false's failure is ignored!)

# With pipefail: $? = exit code of first failure
set -o pipefail
false | true
echo $?     # 1 (false failed — pipeline reports failure)

# PIPESTATUS array (Bash-specific):
ls /etc/hostname /nonexistent 2>/dev/null | sort | wc -l
echo "${PIPESTATUS[0]} ${PIPESTATUS[1]} ${PIPESTATUS[2]}"
# Shows exit code of each command in the pipeline
```

## Common Footguns

**Footgun 1: Variable assignment in a pipe**
```bash
# BROKEN — the while loop runs in a subshell (because of the pipe):
count=0
cat file.txt | while read -r line; do
    ((count++))
done
echo "$count"    # Always 0! The subshell's count is lost!

# FIX — redirect instead of pipe:
count=0
while read -r line; do
    ((count++))
done < file.txt
echo "$count"    # Correct count

# FIX — use lastpipe (Bash 4.2+):
shopt -s lastpipe
count=0
cat file.txt | while read -r line; do
    ((count++))
done
echo "$count"    # Works now (last command runs in current shell)
```

**Footgun 2: grep in a pipeline returning "no match" kills the pipe**
```bash
set -o pipefail
cmd | grep "pattern" | wc -l
# If grep finds NO matches, it exits 1 → pipeline fails!
# Even though "zero matches" is a valid result, not an error

# Fix — accept grep's exit code:
cmd | { grep "pattern" || true; } | wc -l
```

**Footgun 3: Pipe buffering delays output**
```bash
# Commands buffer their output differently:
# - Line-buffered when stdout is a terminal (see output line by line)
# - Block-buffered when stdout is a pipe (see output in chunks)

# This means `tail -f access.log | grep "error"` might show
# nothing for a long time, then several lines at once.

# Fix: force line buffering:
tail -f access.log | grep --line-buffered "error"
# Or use stdbuf:
tail -f access.log | stdbuf -oL grep "error"
```

**Footgun 4: Pipe to commands that need a terminal**
```bash
# Some commands need a terminal (isatty check):
echo "password" | sudo -S command     # Works but insecure (password in process list)
echo "input" | vim                     # vim complains — not a terminal

# Use expect, or script -c, or heredoc for interactive commands
```

## Exercise

1. Build a pipeline step by step to find the 5 largest installed packages on your Fedora system:
   ```bash
   rpm -qa --queryformat '%{SIZE} %{NAME}\n' | sort -rn | head -5
   ```
   Add `tee` between each stage to inspect intermediate results.

2. Demonstrate the subshell pipe problem:
   ```bash
   x=0; echo -e "a\nb\nc" | while read -r line; do ((x++)); done; echo $x
   ```
   Then fix it using `< <(echo -e "a\nb\nc")` or a file redirect.

3. Create a named pipe, write to it from one terminal, and read from another.

4. Use `PIPESTATUS` to check every command's exit code in a 3-stage pipeline.

---

Next: [Level 4: Exit Codes and $?](../04-exit-codes-failure/00-exit-codes.md)
