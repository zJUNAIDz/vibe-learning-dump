# Redirection Deep Dive

## What Problem This Solves

You see `2>&1`, `&>`, `<<-EOF`, and `> >(tee)` in scripts and they look like line noise. Each one is a specific instruction to the shell about where file descriptor streams should point. Once you understand the syntax, they're all readable.

## The Mental Model

Redirection is the shell **reconnecting file descriptors** before launching a command. The command itself doesn't know — it just reads from fd 0 and writes to fd 1 and fd 2 as normal. The shell decided beforehand where those descriptors point.

```
Before redirection:        After: cmd > file.txt
fd 0 → keyboard            fd 0 → keyboard  (unchanged)
fd 1 → terminal            fd 1 → file.txt  (redirected!)
fd 2 → terminal            fd 2 → terminal  (unchanged)
```

## Every Redirection Operator

### Output Redirection

```bash
cmd > file       # fd 1 → file (overwrite). Create if doesn't exist.
cmd >> file      # fd 1 → file (append). Create if doesn't exist.
cmd 2> file      # fd 2 → file (overwrite)
cmd 2>> file     # fd 2 → file (append)
```

### Input Redirection

```bash
cmd < file       # fd 0 ← file. Command reads from file instead of keyboard.
```

### Combining Streams

```bash
cmd > file 2>&1  # fd 1 → file, then fd 2 → wherever fd 1 goes (also file)
cmd &> file      # Bash shorthand for above (stdout + stderr → file)
cmd &>> file     # Bash shorthand: append both stdout and stderr
```

### Here Documents

```bash
cmd << 'EOF'     # Feed literal text as stdin (no expansion with quotes)
text here
EOF

cmd << EOF       # Feed text as stdin (variables expand)
Hello $USER
EOF

cmd <<- EOF      # Same, but leading TABS are stripped (not spaces!)
	indented text
	EOF
```

### Here Strings

```bash
cmd <<< "string"   # Feed a single string as stdin
grep "pattern" <<< "search in this text"
```

## Order Matters: The `2>&1` Rule

**This is the most confusing part of shell redirection and the source of most mistakes.**

`2>&1` means "make fd 2 point to wherever fd 1 *currently* points." It's a **copy**, not "merge them forever."

```bash
# RIGHT: Both stdout and stderr go to file
cmd > file.txt 2>&1
# Step 1: fd 1 → file.txt
# Step 2: fd 2 → wherever fd 1 is → file.txt
# Result: Both in file.txt ✅

# WRONG: Only stdout goes to file, stderr goes to terminal
cmd 2>&1 > file.txt
# Step 1: fd 2 → wherever fd 1 is → terminal (at this moment!)
# Step 2: fd 1 → file.txt
# Result: stderr on terminal, stdout in file ❌ (probably not what you wanted)
```

Think of redirections as executing **left to right**. Each one modifies the state, and `2>&1` copies the *current* state of fd 1.

### The Reading Trick

Read redirections left to right, and for each one, ask "what is this fd currently pointing at?"

```bash
# Example: cmd 3>&1 1>&2 2>&3 3>&-
# This SWAPS stdout and stderr:
# Start:  fd 1 → terminal-stdout, fd 2 → terminal-stderr
# 3>&1:   fd 3 → terminal-stdout (save stdout in fd 3)
# 1>&2:   fd 1 → terminal-stderr (stdout now goes where stderr was)
# 2>&3:   fd 2 → terminal-stdout (stderr now goes where stdout was)
# 3>&-:   fd 3 closed (cleanup)

# After: stdout and stderr are swapped!
```

## Practical Redirection Patterns

### Discard Everything

```bash
cmd > /dev/null 2>&1     # Discard both stdout and stderr
cmd &> /dev/null          # Bash shorthand
```

### Log stdout and stderr separately

```bash
cmd > stdout.log 2> stderr.log
```

### Log everything AND see it on screen

```bash
cmd 2>&1 | tee output.log       # See output AND save it
cmd |& tee output.log            # Bash shorthand
```

### Append to log while monitoring

```bash
cmd 2>&1 | tee -a /var/log/myapp.log
```

### Redirect inside a block

```bash
{
    echo "Starting backup..."
    rsync -av /data /backup
    echo "Backup complete."
} > /var/log/backup.log 2>&1
# All output from the entire block goes to the log
```

### Input from a file with heredoc for multiline

```bash
# Feed a multi-line SQL query:
psql -d mydb << 'SQL'
SELECT u.name, count(o.id) as orders
FROM users u
LEFT JOIN orders o ON o.user_id = u.id
GROUP BY u.name
ORDER BY orders DESC;
SQL
```

### noclobber: Prevent Accidental Overwrites

```bash
set -o noclobber       # Enable
echo "data" > file.txt  # FAILS if file.txt already exists
echo "data" >| file.txt # Override noclobber for this one redirect

set +o noclobber       # Disable
```

## Advanced: exec Redirection

`exec` without a command changes redirections for the **current shell**:

```bash
# Redirect all subsequent stdout to a file:
exec > /tmp/all-output.log
echo "This goes to the file"
echo "So does this"

# Redirect all stderr to a file:
exec 2> /tmp/all-errors.log

# Save original stdout, redirect, then restore:
exec 3>&1               # Save fd 1 in fd 3
exec > /tmp/output.log   # Redirect stdout to file
echo "This goes to file"
exec 1>&3                # Restore stdout from fd 3
exec 3>&-                # Close fd 3
echo "This goes to terminal again"
```

### Script Logging Pattern

```bash
#!/usr/bin/env bash
# Log everything this script does:

LOG_FILE="/var/log/myscript.log"
exec > >(tee -a "$LOG_FILE") 2>&1
# >(tee -a "$LOG_FILE") is process substitution — 
# stdout goes to tee, which writes to both terminal AND log file
# 2>&1 merges stderr into stdout

echo "Starting at $(date)"
# This appears on terminal AND in the log file
```

## Common Footguns

**Footgun 1: Redirecting in the wrong order**
```bash
# You want to capture stderr but not stdout:
cmd 2>&1 >/dev/null
# This sends STDERR to the terminal (where stdout WAS)
# and sends STDOUT to /dev/null
# You probably wanted stderr captured — but it's not

# What you probably want:
cmd 2>&1 >/dev/null | grep error
# stderr is merged into the pre-redirect stdout (terminal/pipe)
# stdout goes to /dev/null
# grep only sees the original stderr
```

**Footgun 2: Writing to a file you're reading from**
```bash
# DANGER — this truncates the file first, then reads nothing:
sort < file.txt > file.txt
# The shell opens file.txt for writing (truncating it) BEFORE sort reads it!

# Fix — use sponge (from moreutils) or a temp file:
sort < file.txt | sponge file.txt
# Or:
sort < file.txt > /tmp/sorted && mv /tmp/sorted file.txt
```

**Footgun 3: `>>` in loops**
```bash
# Opening and closing a file 1000 times:
for i in $(seq 1 1000); do
    echo "$i" >> output.txt    # Each iteration opens, appends, closes
done

# More efficient — redirect the whole loop:
for i in $(seq 1 1000); do
    echo "$i"
done > output.txt    # One open, many writes, one close
```

**Footgun 4: Heredoc indentation with spaces vs tabs**
```bash
# <<- strips leading TABS, not spaces:
if true; then
    cat <<- EOF
    this is indented with spaces — NOT stripped!
	this is indented with a tab — stripped!
	EOF
fi
# Use tabs if you want <<- to work
```

## Exercise

1. Run `ls /etc/hostname /nonexistent > out.txt 2> err.txt`. Check both files.
2. Demonstrate the order problem: run `ls /etc/hostname /nonexistent > out.txt 2>&1` and `ls /etc/hostname /nonexistent 2>&1 > out.txt`. Compare what's in the file vs what's on screen.
3. Use `exec` to redirect all script output to a log file while keeping it on screen (the `tee` pattern).
4. Write a script with a heredoc that creates an nginx config file.

---

Next: [Pipes and Unix Composition](02-pipes-and-composition.md)
