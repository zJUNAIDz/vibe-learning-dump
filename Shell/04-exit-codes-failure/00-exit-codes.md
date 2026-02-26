# Exit Codes and $?

## What Problem This Solves

You run a command. It prints nothing. Did it work? Did it fail? You have no idea. Every Unix command returns an **exit code** — a number from 0 to 255 that tells you whether it succeeded. Scripts that don't check exit codes are scripts that fail silently.

## How People Misunderstand It

1. **"If there's no error message, it worked"** — Many commands fail silently. `grep` returns 1 when it finds no match. `diff` returns 1 when files differ. No error message, but non-zero exit code.
2. **"0 and 1 mean false and true"** — Opposite! In shell, **0 = success (true)**, **non-zero = failure (false)**. This is backwards from every other programming language.
3. **"I just need to check if the command ran"** — You need to check if it *succeeded*. A command can run (process starts) and still fail (non-zero exit).

## The Mental Model

Every command, when it finishes, returns a single integer:

```
0       = Success (the command did what you asked)
1-125   = Failure (command-specific error codes)
126     = Command found but not executable
127     = Command not found
128+N   = Killed by signal N (e.g., 128+9=137 → killed by SIGKILL)
```

### $? — The Exit Code of the Last Command

```bash
ls /etc/hostname
echo $?    # 0 (success)

ls /nonexistent
echo $?    # 2 (ls: error code for "file not found")

grep "nonexistent" /etc/passwd
echo $?    # 1 (grep: "no match found" — not an error, but non-zero!)

grep "root" /etc/passwd
echo $?    # 0 (match found)
```

### Why 0 = Success

Every other language: `0 = false`, `1 = true`. Shell is the opposite because:
- There's **one way to succeed** and **many ways to fail**
- Success is the common case, so it gets the simple value (0)
- Failure codes (1-255) encode *what went wrong*

```bash
# In if statements, the shell checks if the exit code is 0:
if grep -q "root" /etc/passwd; then
    echo "root user exists"     # grep returned 0 → true
fi

if grep -q "nonexistent" /etc/passwd; then
    echo "found"
else
    echo "not found"            # grep returned 1 → false
fi
```

## Common Exit Codes

```
0    Success
1    General error / catchall
2    Misuse of shell builtins (bash), or command-specific
126  Permission denied (file exists but not executable)
127  Command not found
128  Invalid argument to exit
130  Script killed by Ctrl+C (128 + 2, where 2 = SIGINT)
137  Process killed by kill -9 (128 + 9 = SIGKILL)
143  Process killed by kill (128 + 15 = SIGTERM)
```

### Signal-Based Exit Codes

When a process is killed by a signal, its exit code is 128 + signal number:

```bash
# Start a background process:
sleep 1000 &
pid=$!

# Kill it normally (SIGTERM = 15):
kill $pid
wait $pid
echo $?    # 143 (128 + 15)

# Kill it forcefully (SIGKILL = 9):
sleep 1000 &
pid=$!
kill -9 $pid
wait $pid
echo $?    # 137 (128 + 9)

# Ctrl+C (SIGINT = 2):
# If you Ctrl+C a command: exit code = 130 (128 + 2)
```

## && and || — Conditional Execution

These operators use exit codes to decide whether to run the next command:

```bash
# && — run next command ONLY if previous succeeded (exit 0):
mkdir /tmp/mydir && echo "Directory created"
# If mkdir fails, echo doesn't run

# || — run next command ONLY if previous failed (exit non-zero):
mkdir /tmp/mydir || echo "Failed to create directory"
# If mkdir succeeds, echo doesn't run

# Combined (poor man's if-else):
command && echo "ok" || echo "failed"
# WARNING: This is NOT a true if-else!
# If "echo ok" fails (unlikely but possible), "echo failed" also runs.
```

### Chaining with &&

```bash
# Classic safe deployment pattern:
cd /var/www && git pull && systemctl restart myapp && echo "Deployed!"
# If ANY step fails, everything after it is skipped

# Without &&, a failed cd means git pull runs in the wrong directory:
cd /var/www
git pull            # If cd failed, this runs in your home directory!
systemctl restart myapp
```

## Setting Your Own Exit Codes

```bash
#!/usr/bin/env bash

check_disk() {
    local usage
    usage=$(df / --output=pcent | tail -1 | tr -d ' %')
    
    if (( usage > 90 )); then
        echo "CRITICAL: Disk usage at ${usage}%"
        return 2     # Return specific error code
    elif (( usage > 70 )); then
        echo "WARNING: Disk usage at ${usage}%"
        return 1
    else
        echo "OK: Disk usage at ${usage}%"
        return 0
    fi
}

check_disk
exit_code=$?

case $exit_code in
    0) echo "All good" ;;
    1) echo "Needs attention" ;;
    2) echo "Needs immediate action" ;;
esac

exit $exit_code    # Pass the code to the caller
```

### exit vs return

```bash
exit N     # Terminate the SCRIPT (or current shell) with code N
return N   # Return from a FUNCTION with code N

# In a script: use exit
# In a function: use return
# exit inside a function still kills the entire script!
```

## Common Footguns

**Footgun 1: Check $? too late**
```bash
some_command
echo "Command finished"
if [[ $? -ne 0 ]]; then
    echo "Command failed"    # WRONG — $? is the exit code of echo, not some_command!
fi

# Fix — capture immediately:
some_command
result=$?
echo "Command finished"
if [[ $result -ne 0 ]]; then
    echo "Command failed"
fi

# Or just use if directly:
if ! some_command; then
    echo "Command failed"
fi
```

**Footgun 2: grep returning 1 doesn't mean an error**
```bash
# grep returns 1 for "no matches" — not an error
grep "pattern" file.txt
echo $?
# 0 = pattern found
# 1 = pattern not found
# 2 = actual error (file doesn't exist, bad regex, etc.)

# In scripts with set -e, grep returning 1 kills the script!
set -e
count=$(grep -c "pattern" file.txt)   # Script exits if no matches!

# Fix:
count=$(grep -c "pattern" file.txt || true)
```

**Footgun 3: diff returns 1 as "normal" result**
```bash
diff file1 file2
echo $?
# 0 = files are identical
# 1 = files differ (not an error!)
# 2 = error (file not found, etc.)
```

**Footgun 4: Command substitution hides exit codes**
```bash
result=$(failing_command)
echo $?    # 0?! No — this IS the exit code of failing_command
           # BUT $result is the stdout, and the command may have printed
           # error messages to stderr that you didn't capture

# More subtle:
local result=$(failing_command)    # In a function with 'local'
echo $?    # ALWAYS 0! 'local' itself succeeds and clobbers $?

# Fix:
local result
result=$(failing_command)
echo $?    # Now shows failing_command's exit code
```

## Exercise

1. Check exit codes of these commands and explain each:
   ```bash
   true; echo $?
   false; echo $?
   ls /nonexistent 2>/dev/null; echo $?
   grep -q "root" /etc/passwd; echo $?
   grep -q "zzzzz" /etc/passwd; echo $?
   ```

2. Write a script that takes a filename argument, checks if the file exists, and exits with 0 (exists), 1 (doesn't exist), or 2 (no argument provided).

3. Chain three commands with `&&` and observe that a failure in the middle prevents the rest from running.

4. Demonstrate the `local` clobbering problem: write a function where `local var=$(false)` and show `$?` is 0.

---

Next: [set -euo pipefail Explained](01-set-flags.md)
