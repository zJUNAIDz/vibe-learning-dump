# Streams: stdin, stdout, stderr

## What Problem This Solves

You run a command and see output, but you can't control where it goes. Error messages interleave with data. You can't redirect just the errors. Understanding that every process has three separate streams — and what each one is — makes I/O control straightforward.

## How People Misunderstand It

1. **"A program just prints text"** — A program sends text to specific numbered streams. What you see on screen is *two* streams (stdout + stderr) merged by the terminal.
2. **"Errors go to the same place as output"** — They go to the SAME terminal by default, but they're separate streams. You can redirect them independently.
3. **"Pipes capture everything"** — Pipes only connect stdout. stderr still goes to the terminal unless you explicitly redirect it.

## The Mental Model

Every Unix process starts with three open file descriptors:

```
┌──────────────────────────────┐
│         Your Process         │
│                              │
│  fd 0 (stdin)  ← Input      │  ← keyboard, file, pipe, /dev/null
│  fd 1 (stdout) → Output     │  → terminal, file, pipe
│  fd 2 (stderr) → Errors     │  → terminal, file, pipe
│                              │
│  fd 3, 4, 5... → Other      │  → files, sockets, etc.
│                              │
└──────────────────────────────┘
```

- **fd 0 (stdin)**: Where the process reads input from. By default, the keyboard (terminal).
- **fd 1 (stdout)**: Where the process writes normal output. By default, the terminal.
- **fd 2 (stderr)**: Where the process writes error/diagnostic output. By default, the terminal.

These are **file descriptors** — just integer handles to open files, devices, pipes, or sockets. The kernel manages the mapping. The process just reads from fd 0 and writes to fd 1 and fd 2.

### Why Two Output Streams?

The separation exists so you can **process data while still seeing errors**:

```bash
# Without separation: errors mixed into data pipeline
sort /etc/passwd /nonexistent
# Error about /nonexistent PLUS sorted passwd data — mixed garbage

# With separation: errors on screen, data processed
sort /etc/passwd /nonexistent 2>/dev/null
# Only the sorted passwd data — errors discarded

# Or: save data to file, see errors on screen
sort /etc/passwd /nonexistent > sorted.txt
# Terminal shows error. File has only data.
```

## Seeing the Streams in Action

```bash
# A command that writes to BOTH stdout and stderr:
ls /etc/hostname /nonexistent
# stdout: /etc/hostname
# stderr: ls: cannot access '/nonexistent': No such file or directory
# Both appear on your terminal, interleaved

# Prove they're separate by redirecting only stdout:
ls /etc/hostname /nonexistent > /dev/null
# Only stderr appears (the error message)

# Redirect only stderr:
ls /etc/hostname /nonexistent 2> /dev/null
# Only stdout appears (/etc/hostname)

# Redirect both to different files:
ls /etc/hostname /nonexistent > stdout.txt 2> stderr.txt
cat stdout.txt    # /etc/hostname
cat stderr.txt    # ls: cannot access '/nonexistent': ...
```

### Programs That USE the Difference

Well-written programs send data to stdout and diagnostics to stderr:

```bash
# curl sends data to stdout, progress to stderr:
curl https://example.com > page.html
# Progress bar appears (stderr) while HTML goes to file (stdout)

curl https://example.com 2>/dev/null > page.html
# Silent download — progress bar suppressed

# tar sends extracted file info to stderr, data to stdout:
tar czf - /etc/hostname 2>/dev/null | wc -c
# Counts bytes of the tar archive (stdout) without listing files (stderr)
```

### Programs that Read from stdin

```bash
# These commands read from stdin if no file argument is given:
cat              # Type text, Ctrl+D to end
sort             # Type lines, Ctrl+D to sort them
grep "pattern"   # Type lines, matches are printed
wc               # Type text, Ctrl+D for counts

# stdin can come from:
cat < file.txt           # From a file
echo "hello" | cat       # From a pipe  
cat <<< "hello"          # From a here-string
cat << EOF               # From a here-document
hello world
EOF
```

## File Descriptors Are Just Numbers

The shell lets you open more file descriptors:

```bash
# Open fd 3 for writing:
exec 3> /tmp/mylog.txt
echo "Log entry 1" >&3     # Write to fd 3
echo "Normal output"        # Goes to stdout (fd 1) as usual
echo "Log entry 2" >&3
exec 3>&-                   # Close fd 3

# Open fd 4 for reading:
exec 4< /etc/hostname
read -r hostname <&4        # Read from fd 4
exec 4<&-                   # Close fd 4
echo "Hostname: $hostname"
```

### /dev/null, /dev/zero, /dev/urandom

```bash
/dev/null     # Write: data disappears. Read: immediate EOF.
/dev/zero     # Read: infinite stream of zero bytes. Write: discarded.
/dev/urandom  # Read: infinite stream of random bytes.

# Discard all output:
cmd > /dev/null 2>&1

# Generate 16 random hex chars:
head -c 8 /dev/urandom | xxd -p

# Create a 1GB file of zeros:
dd if=/dev/zero of=bigfile bs=1M count=1024
```

### /dev/stdin, /dev/stdout, /dev/stderr

These are symlinks to your actual file descriptors:

```bash
ls -la /dev/stdin /dev/stdout /dev/stderr
# /dev/stdin -> /proc/self/fd/0
# /dev/stdout -> /proc/self/fd/1
# /dev/stderr -> /proc/self/fd/2

# Useful when a program needs a filename but you want to use a stream:
curl https://example.com | diff /dev/stdin local-copy.html
# Compares downloaded content (via stdin) with a local file
```

## Checking If stdin/stdout Is a Terminal

Programs often change behavior based on whether they're connected to a terminal:

```bash
# In Bash:
[[ -t 0 ]] && echo "stdin is a terminal" || echo "stdin is a pipe/file"
[[ -t 1 ]] && echo "stdout is a terminal" || echo "stdout is a pipe/file"

# This is why:
ls           # Shows colors (stdout is terminal)
ls | cat     # No colors (stdout is a pipe)
ls | less    # No colors (stdout is a pipe, but less is interactive)

# Force color in pipes:
ls --color=always | less -R
```

## Common Footguns

**Footgun 1: Ignoring stderr**
```bash
result=$(command_that_might_fail)
# $result captures stdout ONLY. Errors go to the terminal.
# If you want to capture stderr too:
result=$(command_that_might_fail 2>&1)
# Now both stdout and stderr are in $result

# Capture them separately:
result=$(command 2>/tmp/error.log)
errors=$(cat /tmp/error.log)
```

**Footgun 2: Pipe only captures stdout**
```bash
# This only pipes stdout:
ls /real /fake | grep "real"
# stderr ("ls: cannot access '/fake'...") still on the terminal

# To pipe stderr too:
ls /real /fake 2>&1 | grep "real"   # Merge stderr into stdout, then pipe
ls /real /fake |& grep "real"        # Bash shorthand for the same thing
```

**Footgun 3: Assuming sequential output**
```bash
# stdout and stderr may not interleave in the order you expect
# because they can be independently buffered:
echo "stdout line"     # Line-buffered when stdout is a terminal
echo "stderr line" >&2 # Usually unbuffered

# In a pipeline, stdout becomes fully buffered (block-buffered)
# which means output appears in large chunks, not line by line
# Force line buffering with:
stdbuf -oL command | grep pattern
```

## Exercise

1. Run `ls /etc/hostname /nonexistent 2>/dev/null` and `ls /etc/hostname /nonexistent >/dev/null`. Observe which stream you suppressed in each.
2. Write to a custom file descriptor:
   ```bash
   exec 3> /tmp/myfd.txt
   echo "hello fd 3" >&3
   exec 3>&-
   cat /tmp/myfd.txt
   ```
3. Run `[[ -t 1 ]] && echo terminal || echo pipe` directly, and then `echo test | [[ -t 0 ]] && echo terminal || echo pipe` (use a subshell for this). Observe the difference.
4. Capture both stdout and stderr into separate variables from a single command.

---

Next: [Redirection Deep Dive](01-redirection-deep-dive.md)
