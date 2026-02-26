# Debugging Tools

## What Problem This Solves

Your script fails and you don't know why. Or a command works interactively but not in a script. Or something worked yesterday and is broken today. You need systematic tools to find the problem, not just staring at the code.

## set -x — Trace Execution

The most important debugging tool. Shows every command before it executes, with variables expanded:

```bash
# Enable tracing:
set -x

# Your code:
name="world"
echo "hello $name"

# Output:
# + name=world
# + echo 'hello world'
# hello world
```

### Using set -x Effectively

```bash
# Trace the entire script:
bash -x script.sh

# Trace only a section:
set -x
problematic_function
set +x    # Disable

# Trace to a file (keep stdout/stderr clean):
exec 3>&2 2>/tmp/trace.log
set -x
# ... your code ...
set +x
exec 2>&3 3>&-
# Now check /tmp/trace.log

# Custom trace prefix (show line numbers):
export PS4='+ ${BASH_SOURCE[0]}:${LINENO}: ${FUNCNAME[0]:+${FUNCNAME[0]}(): }'
set -x
# Output now shows: + script.sh:42: main(): command
```

### PS4 — Custom Trace Format

```bash
# Default PS4 is just "+ "
# Make it informative:

# Show line numbers:
PS4='+(${LINENO}): '

# Show function, file, and line:
PS4='+${BASH_SOURCE}:${LINENO}:${FUNCNAME[0]}() '

# Show timestamp:
PS4='+ $(date +%T.%N) ${BASH_SOURCE}:${LINENO}: '

# With color:
PS4=$'\033[33m+ ${BASH_SOURCE}:${LINENO}: \033[0m'
```

## printf Debugging

When set -x is too noisy:

```bash
# Strategic debug prints:
debug() {
    [[ "${DEBUG:-false}" = true ]] && printf 'DEBUG: %s\n' "$*" >&2
    return 0
}

main() {
    local config_file="$1"
    debug "config_file=$config_file"

    local content
    content=$(< "$config_file")
    debug "content length=${#content}"

    local parsed
    parsed=$(parse "$content")
    debug "parsed=$parsed"
}

# Enable with:
DEBUG=true ./script.sh
```

### Variable Inspection

```bash
# Print variable details:
inspect() {
    local var_name="$1"
    local var_value="${!var_name}"    # Indirect expansion
    printf 'INSPECT: %s=[%s] (length=%d)\n' "$var_name" "$var_value" "${#var_value}" >&2
}

inspect PATH
inspect HOME
inspect MY_VAR

# Dump all variables:
declare -p          # All variables with their attributes
declare -p MY_VAR   # Specific variable

# Show array contents:
declare -p my_array
# declare -a my_array=([0]="one" [1]="two" [2]="three")
```

## strace — System Call Tracing

When you need to see what a program does at the OS level:

```bash
# Trace all system calls:
strace ls /tmp

# Common filters:
strace -e trace=open,read,write cat /etc/hostname     # File I/O
strace -e trace=network curl http://example.com        # Network calls
strace -e trace=process bash -c 'echo hi'             # Process creation

# Follow child processes:
strace -f bash script.sh

# Save to file:
strace -o /tmp/trace.log -f bash script.sh

# Show timestamps:
strace -t command    # Wall clock time
strace -r command    # Relative time between calls

# Summary only (no individual calls):
strace -c command
# Shows: time, calls, errors per syscall
```

### What strace Reveals

```bash
# "Why can't my script find this file?"
strace -e openat command 2>&1 | grep "No such file"
# Shows EXACTLY which files the program tried to open

# "Why is this command slow?"
strace -c -f slow_script.sh
# Shows which syscalls take the most time

# "What files does this program read?"
strace -e openat,stat command 2>&1 | grep -v ENOENT
# Shows all files successfully opened

# "What environment does a command see?"
strace -e trace=read -f env 2>&1 | head
```

## ltrace — Library Call Tracing

Like strace but for library function calls:

```bash
sudo dnf install ltrace

# Trace library calls:
ltrace ls 2>&1 | head -20

# Shows calls like:
# strlen("hello") = 5
# malloc(1024) = 0x55a...
# printf("Hello %s\n", "world") = 12
```

## Debugging Specific Problems

### "It Works Interactively but Not in a Script"

```bash
# Compare environments:
env > /tmp/interactive_env.txt           # In your terminal
# In the script:
env > /tmp/script_env.txt

diff /tmp/interactive_env.txt /tmp/script_env.txt

# Common culprits:
# 1. PATH is different
echo "PATH=$PATH" >&2

# 2. Shell is different
echo "Shell: $0, BASH_VERSION=$BASH_VERSION" >&2

# 3. Aliases aren't available in scripts
type ll 2>/dev/null || echo "ll alias not found" >&2

# 4. Dotfiles not loaded (non-interactive shell)
echo "Interactive: $-" >&2    # Contains 'i' if interactive
```

### "It Worked Yesterday"

```bash
# Check what changed:
# 1. Recent package updates:
dnf history --reverse | tail -10

# 2. Recently modified files:
find /etc -mtime -1 -ls 2>/dev/null | head -20

# 3. Git history:
git log --oneline --since="yesterday"
git diff HEAD~5..HEAD

# 4. System changes:
journalctl --since yesterday -p warning --no-pager | head -50

# 5. Disk space:
df -h

# 6. Run the same command with explicit debug:
bash -x /path/to/script.sh 2>/tmp/debug.log
```

### "Command Not Found" But It's Installed

```bash
# 1. Check if it's installed:
which command_name      # Check PATH
type command_name       # Check everything (builtins, aliases, functions)
command -v command_name # Most reliable

# 2. Find the actual binary:
rpm -ql package-name | grep bin/

# 3. Check PATH:
echo "$PATH" | tr ':' '\n'

# 4. Maybe it's in a different name:
dnf provides '*/command_name'

# 5. Hash table might be stale:
hash -r    # Clear the hash table (Bash)
rehash     # Clear the hash table (Zsh)
```

### "Permission Denied" But Permissions Look Right

```bash
# 1. Check all path components:
namei -l /full/path/to/file
# Shows permissions for EVERY directory in the path

# 2. Check SELinux:
ls -Z /path/to/file
sudo ausearch -m avc -ts recent

# 3. Check ACLs:
getfacl /path/to/file

# 4. Check filesystem attributes:
lsattr /path/to/file

# 5. Check mount flags:
mount | grep "$(df /path/to/file | tail -1 | awk '{print $6}')"

# 6. Check capabilities:
getcap /path/to/binary
```

### Debugging Pipes

```bash
# See what each stage produces:
cmd1 | tee /tmp/after_cmd1.txt | cmd2 | tee /tmp/after_cmd2.txt | cmd3

# Check exit codes of each stage:
cmd1 | cmd2 | cmd3
echo "Exit codes: ${PIPESTATUS[*]}"

# Add stage-by-stage debugging:
set -o pipefail
cmd1 | { tee /dev/stderr; } | cmd2 | { tee /dev/stderr; } | cmd3
```

## Process Investigation

```bash
# What is this process doing?
# 1. What files does it have open:
ls -l /proc/$(pgrep myapp)/fd

# 2. What's its current directory:
readlink /proc/$(pgrep myapp)/cwd

# 3. What environment does it have:
cat /proc/$(pgrep myapp)/environ | tr '\0' '\n'

# 4. What command line was used:
cat /proc/$(pgrep myapp)/cmdline | tr '\0' ' '

# 5. Where is it in the filesystem:
readlink /proc/$(pgrep myapp)/exe

# 6. What's it doing right now:
sudo strace -p $(pgrep myapp) -e trace=read,write

# 7. CPU/memory snapshot:
ps -p $(pgrep myapp) -o pid,vsz,rss,%cpu,%mem,etime,args
```

## Network Debugging from Shell

```bash
# Is the port listening?
ss -tlnp | grep :8080

# Can I connect?
timeout 3 bash -c 'echo > /dev/tcp/localhost/8080' && echo "open" || echo "closed"

# What's the DNS resolution?
dig +short example.com
host example.com

# Trace the network path:
traceroute example.com
mtr example.com

# Test HTTP:
curl -v http://localhost:8080/health

# Watch connections in real time:
watch -n 1 'ss -tnp | grep 8080'
```

## The Debug Checklist

When something doesn't work:

```
1. READ THE ERROR MESSAGE (fully — don't skim)
2. Check the exit code: echo $?
3. Run with set -x: bash -x script.sh
4. Check environment: env, echo $PATH
5. Check permissions: ls -la, namei -l, getenforce
6. Check logs: journalctl -u service, /var/log/
7. Simplify: reduce to minimal reproduction
8. Compare: what's different from when it works?
9. strace as last resort: strace -f command
```

## Exercise

1. Take a working script and add `PS4` with line numbers. Run it with `bash -x` and follow the execution flow.

2. Use `strace -e openat` on a command to see exactly which files it opens. Try it on `bash` and `python3`.

3. Simulate the "works interactively, fails in script" problem: create a script that uses an alias. Debug why it fails and fix it.

4. Use `namei -l` to diagnose a "permission denied" error on a deeply nested file. Identify which directory in the path is blocking access.

---

Next: [Environment Differences](01-environment-differences.md)
