# Conditionals and Tests

## What Problem This Solves

You write `if [ $x = "yes" ]` and it breaks when `$x` is empty. Or you use `-a` and `-o` and get bizarre results. Bash has **three different conditional systems** and most people use the wrong one for the job.

## The Three Systems

```
┌─────────────┬──────────────┬────────────────────────────┐
│ Syntax      │ Type         │ When to Use                │
├─────────────┼──────────────┼────────────────────────────┤
│ [ ... ]     │ Command      │ POSIX scripts (sh)         │
│ [[ ... ]]   │ Bash keyword │ Bash/Zsh scripts (prefer!) │
│ (( ... ))   │ Arithmetic   │ Number comparisons         │
└─────────────┴──────────────┴────────────────────────────┘
```

## `[ ... ]` — The Old Way (Actually a Command)

`[` is literally a command — it's `/usr/bin/[` (or a shell builtin). This means:

```bash
# Every token must be a SEPARATE word:
[ "$x" = "yes" ]     # 4 arguments to the [ command: "$x", "=", "yes", "]"

# Forget the spaces and things break:
["$x" = "yes"]       # Tries to run command named [hello, which doesn't exist

# Empty variable breaks it:
x=""
[ $x = "yes" ]       # Becomes: [ = "yes" ] → syntax error!
# Fix: always quote
[ "$x" = "yes" ]     # Becomes: [ "" = "yes" ] → works

# -a and -o are ambiguous:
[ -n "$a" -a -n "$b" ]    # Can break in edge cases
# Use separate tests instead:
[ -n "$a" ] && [ -n "$b" ]
```

**Rule: Use `[[ ]]` in Bash/Zsh scripts. Only use `[ ]` in POSIX sh scripts.**

## `[[ ... ]]` — The Right Way in Bash/Zsh

`[[ ]]` is a **shell keyword**, not a command. The shell parses it specially:

```bash
# No word splitting inside [[ ]]:
x=""
[[ $x = "yes" ]]     # Works fine! No need to quote (but quoting is still good practice)

# Pattern matching with ==:
[[ "$name" == j* ]]       # True if name starts with j (glob, not regex)
[[ "$name" == "j*" ]]     # True only if name is literally "j*" (quoted = literal)

# Regex matching with =~:
[[ "$email" =~ ^[a-zA-Z]+@[a-zA-Z]+\.[a-z]{2,}$ ]]
if [[ $? -eq 0 ]]; then
    echo "Valid email"
fi
# Captured groups in BASH_REMATCH:
[[ "2024-01-15" =~ ^([0-9]{4})-([0-9]{2})-([0-9]{2})$ ]]
echo "${BASH_REMATCH[1]}"    # 2024
echo "${BASH_REMATCH[2]}"    # 01

# Logical operators use && and || (not -a and -o):
[[ -f "$file" && -r "$file" ]]
[[ "$x" = "a" || "$x" = "b" ]]

# No need to escape < and >:
[[ "$a" < "$b" ]]    # String comparison — works!
[ "$a" \< "$b" ]     # Must escape in [ ] or it becomes redirection
```

## String Tests

```bash
# String equality:
[[ "$str" = "value" ]]      # Equal (single = is fine in [[ ]])
[[ "$str" == "value" ]]     # Equal (== also works in [[ ]])
[[ "$str" != "value" ]]     # Not equal

# String emptiness:
[[ -z "$str" ]]      # True if string is empty (zero-length)
[[ -n "$str" ]]      # True if string is non-empty
[[ "$str" ]]          # Same as -n (non-empty)  — but less clear

# String ordering (lexicographic):
[[ "$a" < "$b" ]]    # a sorts before b
[[ "$a" > "$b" ]]    # a sorts after b
```

## Number Comparisons

```bash
# Inside [[ ]] — use the old operators:
[[ "$a" -eq "$b" ]]    # Equal
[[ "$a" -ne "$b" ]]    # Not equal
[[ "$a" -lt "$b" ]]    # Less than
[[ "$a" -le "$b" ]]    # Less than or equal
[[ "$a" -gt "$b" ]]    # Greater than
[[ "$a" -ge "$b" ]]    # Greater than or equal

# Inside (( )) — use normal math operators (PREFERRED for numbers):
(( a == b ))    # Equal
(( a != b ))    # Not equal
(( a < b ))     # Less than
(( a <= b ))    # Less than or equal
(( a > b ))     # Greater than
(( a >= b ))    # Greater than or equal

# (( )) doesn't need $ for variables:
count=5
if (( count > 3 )); then
    echo "Count is more than 3"
fi

# DANGER: Don't use = or > for numbers inside [[ ]]:
[[ 9 > 10 ]]    # TRUE! Because "9" > "10" lexicographically!
(( 9 > 10 ))    # FALSE — correct numeric comparison
```

## File Tests

```bash
# Existence:
[[ -e "$path" ]]    # Exists (any type)
[[ -f "$path" ]]    # Is a regular file
[[ -d "$path" ]]    # Is a directory
[[ -L "$path" ]]    # Is a symbolic link
[[ -S "$path" ]]    # Is a socket
[[ -p "$path" ]]    # Is a named pipe (FIFO)

# Permissions:
[[ -r "$path" ]]    # Is readable
[[ -w "$path" ]]    # Is writable
[[ -x "$path" ]]    # Is executable

# Size:
[[ -s "$path" ]]    # Exists and has size > 0

# Comparison:
[[ "$file1" -nt "$file2" ]]    # file1 is newer than file2
[[ "$file1" -ot "$file2" ]]    # file1 is older than file2
[[ "$file1" -ef "$file2" ]]    # Same file (hard link or same inode)
```

## if/elif/else

```bash
if [[ "$status" = "active" ]]; then
    echo "Service is running"
elif [[ "$status" = "inactive" ]]; then
    echo "Service is stopped"
else
    echo "Unknown status: $status"
fi

# Remember: if tests the EXIT CODE of any command, not just [[ ]]:
if grep -q "error" /var/log/messages; then
    echo "Errors found!"
fi

if systemctl is-active --quiet sshd; then
    echo "SSH is running"
fi

if command -v docker &>/dev/null; then
    echo "Docker is installed"
fi
```

## case Statements

For multiple string matching, `case` is cleaner than nested if/elif:

```bash
case "$1" in
    start)
        do_start
        ;;
    stop)
        do_stop
        ;;
    restart|reload)        # Multiple patterns with |
        do_stop
        do_start
        ;;
    status)
        do_status
        ;;
    *)                     # Default (catch-all)
        echo "Usage: $0 {start|stop|restart|status}" >&2
        exit 1
        ;;
esac

# Pattern matching in case:
case "$input" in
    [0-9]*)       echo "Starts with a digit" ;;
    [a-zA-Z]*)    echo "Starts with a letter" ;;
    -*)           echo "Starts with a dash (flag?)" ;;
    "")           echo "Empty string" ;;
    *)            echo "Something else" ;;
esac

# Case with ;& (fall-through, Bash 4+):
case "$level" in
    error)
        send_alert
        ;&                 # Fall through to the next case
    warn)
        log_message
        ;&
    info)
        update_counter
        ;;
esac
```

## Short-Circuit Patterns

```bash
# && as a simple "if true, do this":
[[ -f "$config" ]] && source "$config"

# || as a simple "if false, do this":
[[ -d "$dir" ]] || mkdir -p "$dir"

# Combined — careful with this:
[[ -f "$file" ]] && echo "exists" || echo "missing"
# This is NOT the same as if/else!
# If the && command fails, the || also runs:
[[ -f "$file" ]] && false || echo "this always prints if file exists!"

# Rule: Use && and || for simple one-liners. Use if/else for anything complex.
```

## The test Command and Its Quirks

```bash
# These are all equivalent:
test -f "$file"
[ -f "$file" ]

# test with no arguments returns false:
test; echo $?      # 1

# test with one argument returns true if non-empty:
test "hello"; echo $?    # 0
test ""; echo $?         # 1

# This creates bizarre edge cases:
x="="
[ "$x" ]        # Is this "test if $x is non-empty" or "test = ???"
# Answer: it's valid because [ has 1 argument ("="), which is non-empty → true
```

## Common Footguns

### 1. Forgetting Semicolons

```bash
# WRONG:
if [[ "$x" = "yes" ]]
then                      # Works only because then is on the next line

# Both of these are correct:
if [[ "$x" = "yes" ]]; then
    echo "yes"
fi

if [[ "$x" = "yes" ]]
then
    echo "yes"
fi
```

### 2. = vs == vs -eq

```bash
[[ "5" = "5" ]]     # String comparison — true
[[ "5" == "5" ]]    # String comparison — true (same as = in [[ ]])
[[ 5 -eq 5 ]]       # Numeric comparison — true
(( 5 == 5 ))         # Numeric comparison — true

[[ "05" = "5" ]]    # String: FALSE (different strings)
[[ 05 -eq 5 ]]       # Numeric: TRUE (same number)
```

### 3. Testing Command Success vs Output

```bash
# WRONG — tests if the OUTPUT is non-empty, not if the command succeeded:
if [[ $(some_command) ]]; then ...

# RIGHT — tests the exit code:
if some_command; then ...

# If you need both:
output=$(some_command)
if [[ $? -eq 0 && -n "$output" ]]; then
    echo "Command succeeded and produced output: $output"
fi
```

## Exercise

1. Write a script that takes a filepath as `$1` and reports: whether it exists, its type (file/directory/symlink/other), its permissions (readable/writable/executable), and its size status (empty or non-empty).

2. Write a `case` statement that categorizes HTTP status codes: 2xx → success, 3xx → redirect, 4xx → client error, 5xx → server error.

3. Demonstrate the difference between `=` and `-eq` by comparing "05" with "5" both ways.

4. Show why `[[ "9" > "10" ]]` is true (string comparison) while `(( 9 > 10 ))` is false (numeric).

---

Next: [Loops and Iteration](02-loops.md)
