# Shell Quick Reference

> Use this **after** completing the curriculum. Without the mental models, this is just another cheat sheet.

---

## Command Resolution Order

```
1. Aliases         →  alias ll='ls -la'
2. Functions       →  my_func() { ... }
3. Builtins        →  cd, echo, type, export
4. Hash table      →  cached $PATH lookups
5. $PATH search    →  /usr/bin/grep, /usr/local/bin/jq
```

Check what a command resolves to:
```bash
type -a grep        # Shows ALL matches (alias, function, builtin, path)
command -v grep     # POSIX-portable, scripts should use this
which grep          # Lies sometimes — avoid in scripts
```

---

## Quoting Rules

| Syntax | Expansion? | Use When |
|--------|-----------|----------|
| `"double"` | Variables and `$()` expand | Almost always — default choice |
| `'single'` | Nothing expands | Literal strings, regex patterns |
| `$'ansi'` | Escape sequences (`\n`, `\t`) | Need literal special chars |
| No quotes | Word splitting + globbing | Almost never in scripts |

**Rule of thumb: When in doubt, double-quote.**

```bash
file="my file.txt"
cat "$file"          # ✅ Correct — treats as one argument
cat $file            # ❌ Broken — splits into "my" and "file.txt"
```

---

## Expansion Order

The shell expands in this order (matters for debugging):

```
1. Brace expansion       →  {a,b,c}  →  a b c
2. Tilde expansion        →  ~  →  /home/user
3. Parameter expansion    →  $var, ${var:-default}
4. Command substitution   →  $(cmd)
5. Arithmetic expansion   →  $((1+2))
6. Word splitting         →  (on unquoted results of 3-5)
7. Pathname expansion     →  *.txt  →  file1.txt file2.txt
8. Quote removal          →  quotes stripped from final result
```

---

## Redirection Cheat Sheet

```bash
cmd > file          # stdout → file (overwrite)
cmd >> file         # stdout → file (append)
cmd 2> file         # stderr → file
cmd &> file         # stdout + stderr → file (Bash shorthand)
cmd > file 2>&1     # stdout + stderr → file (POSIX)
cmd 2>&1 | less     # stderr merged into stdout, then piped
cmd < file          # stdin ← file
cmd <<< "string"    # here-string (stdin from string)
cmd << 'EOF'        # here-doc (stdin from block, no expansion with quotes)
text here
EOF

cmd > /dev/null 2>&1   # Discard all output
```

**Order matters:** `2>&1 > file` ≠ `> file 2>&1`

---

## Exit Codes

```bash
$?                  # Exit code of last command (0 = success)
cmd && next         # Run 'next' only if 'cmd' succeeds
cmd || fallback     # Run 'fallback' only if 'cmd' fails
cmd1 | cmd2         # $? is exit code of cmd2 (last in pipe)
${PIPESTATUS[@]}    # Array of all exit codes in last pipeline (Bash)
```

### Script Safety Header

```bash
#!/usr/bin/env bash
set -euo pipefail   # Exit on error, undefined vars, pipe failures
```

| Flag | Effect |
|------|--------|
| `set -e` | Exit on any command failure |
| `set -u` | Error on undefined variables |
| `set -o pipefail` | Pipe fails if ANY command in pipe fails |
| `set -x` | Print every command before executing (debug) |

---

## Variable Operations

```bash
# Parameter expansion
${var:-default}     # Use default if var is unset/empty
${var:=default}     # Set var to default if unset/empty
${var:+alternate}   # Use alternate if var IS set
${var:?error msg}   # Exit with error if var is unset/empty

# String manipulation
${var#pattern}      # Remove shortest prefix match
${var##pattern}     # Remove longest prefix match
${var%pattern}      # Remove shortest suffix match
${var%%pattern}     # Remove longest suffix match
${var/old/new}      # Replace first match
${var//old/new}     # Replace all matches
${#var}             # String length
${var:offset:len}   # Substring
```

---

## Tests and Conditionals

```bash
# String tests
[[ -z "$var" ]]     # True if empty
[[ -n "$var" ]]     # True if not empty
[[ "$a" == "$b" ]]  # String equality
[[ "$a" =~ regex ]] # Regex match (Bash only)

# File tests
[[ -f "$path" ]]    # Is regular file
[[ -d "$path" ]]    # Is directory
[[ -e "$path" ]]    # Exists
[[ -r "$path" ]]    # Is readable
[[ -w "$path" ]]    # Is writable
[[ -x "$path" ]]    # Is executable
[[ -s "$path" ]]    # Exists and not empty

# Numeric comparison
[[ "$a" -eq "$b" ]] # Equal
[[ "$a" -lt "$b" ]] # Less than
[[ "$a" -gt "$b" ]] # Greater than
(( a > b ))         # Arithmetic context (cleaner)
```

---

## Common Patterns

```bash
# Process substitution (Bash/Zsh)
diff <(cmd1) <(cmd2)          # Compare output of two commands

# Safe temp file
tmpfile=$(mktemp)
trap 'rm -f "$tmpfile"' EXIT  # Auto-cleanup on exit

# Read file line by line
while IFS= read -r line; do
  echo "$line"
done < file.txt

# Loop over find results safely
find . -name '*.log' -print0 | while IFS= read -r -d '' file; do
  echo "$file"
done

# Default values
name="${1:-anonymous}"

# Check if command exists
if command -v jq &>/dev/null; then
  echo "jq is available"
fi
```

---

## Fedora-Specific Commands

```bash
# Package management
sudo dnf install <pkg>
sudo dnf remove <pkg>
dnf search <term>
dnf info <pkg>
dnf provides '*/bin/dig'      # What package provides this file?

# Services
systemctl status <service>
systemctl start/stop/restart <service>
systemctl enable/disable <service>
journalctl -u <service> -f    # Follow logs

# SELinux
getenforce                     # Check mode
ls -Z /path                    # Show SELinux context
ausearch -m avc -ts recent     # Recent denials
restorecon -Rv /path           # Fix contexts
setsebool -P <bool> on         # Set boolean permanently

# Firewall
firewall-cmd --list-all
firewall-cmd --add-service=http --permanent
firewall-cmd --reload
```
