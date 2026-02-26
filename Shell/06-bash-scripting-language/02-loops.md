# Loops and Iteration

## What Problem This Solves

You need to process every file in a directory, retry a command 5 times, or read lines from a file. But you're doing `for f in $(ls *.txt)` (broken!) or `cat file | while read line` (broken for different reasons). Here's how loops actually work.

## for Loops

### Iterate Over a List

```bash
# Literal list:
for fruit in apple banana cherry; do
    echo "$fruit"
done

# Variable (careful with quoting!):
files="one.txt two.txt three.txt"
for f in $files; do    # Word splitting happens here — intentional but fragile
    echo "$f"
done

# Array (THE correct way for lists with spaces):
files=("my file.txt" "your file.txt" "data.csv")
for f in "${files[@]}"; do
    echo "Processing: $f"
done
```

### Iterate Over Files (Globbing)

```bash
# CORRECT — glob expansion happens safely:
for f in *.txt; do
    # Handle case where no files match:
    [[ -e "$f" ]] || continue
    echo "Processing $f"
done

# WRONG — never parse ls output:
for f in $(ls *.txt); do    # Breaks on spaces, newlines, special chars
    echo "$f"
done

# Recursive — use find or ** glob:
# Bash (requires shopt -s globstar):
shopt -s globstar
for f in **/*.log; do
    [[ -e "$f" ]] || continue
    echo "$f"
done

# Or use find (more portable, more powerful):
while IFS= read -r -d '' f; do
    echo "Processing: $f"
done < <(find /var/log -name "*.log" -print0)
```

### C-Style for Loop

```bash
for ((i = 0; i < 10; i++)); do
    echo "Iteration $i"
done

# With step:
for ((i = 0; i <= 100; i += 10)); do
    echo "$i"
done

# Count down:
for ((i = 10; i > 0; i--)); do
    echo "$i..."
done
echo "Launch!"
```

### Brace Expansion in for

```bash
# Range:
for i in {1..10}; do
    echo "$i"
done

# Range with step:
for i in {0..100..5}; do
    echo "$i"
done

# GOTCHA: Brace expansion doesn't work with variables!
n=10
for i in {1..$n}; do     # WRONG — prints literally "{1..10}"
    echo "$i"
done

# Fix — use seq or C-style loop:
for i in $(seq 1 "$n"); do
    echo "$i"
done

for ((i = 1; i <= n; i++)); do
    echo "$i"
done
```

## while Loops

### Basic while

```bash
count=0
while (( count < 5 )); do
    echo "Count: $count"
    ((count++))
done

# Infinite loop:
while true; do
    echo "Press Ctrl+C to stop"
    sleep 1
done

# Until — the opposite of while:
until ping -c 1 -W 1 google.com &>/dev/null; do
    echo "Waiting for network..."
    sleep 2
done
echo "Network is up!"
```

### Reading Lines from a File

```bash
# CORRECT:
while IFS= read -r line; do
    echo "Line: $line"
done < /etc/hostname

# Breaking down the parts:
# IFS=         → Don't strip leading/trailing whitespace
# read -r      → Don't interpret backslashes
# line         → Variable name to store each line
# < file       → Redirect file into the while loop's stdin

# WRONG — useless use of cat:
cat /etc/hostname | while read line; do    # Also a pipe subshell problem!
    echo "Line: $line"
done
# Problems: 1) cat is unnecessary, 2) pipe creates subshell,
# 3) missing IFS=, 4) missing -r

# Reading with a custom delimiter:
while IFS= read -r -d ',' field; do
    echo "Field: $field"
done <<< "one,two,three,"

# Reading colon-separated fields (like /etc/passwd):
while IFS=: read -r user _ uid gid _ home shell; do
    echo "$user (UID=$uid) → $shell"
done < /etc/passwd
```

### Processing Command Output

```bash
# Use process substitution to avoid subshell:
while IFS= read -r line; do
    echo "Process: $line"
done < <(ps aux --no-headers)

# Process null-delimited output (safe for any filename):
while IFS= read -r -d '' file; do
    echo "Found: $file"
done < <(find /tmp -name "*.log" -print0)
```

## select (Interactive Menus)

```bash
echo "Choose a color:"
select color in red green blue quit; do
    case "$color" in
        red|green|blue)
            echo "You chose $color"
            ;;
        quit)
            break
            ;;
        *)
            echo "Invalid option"
            ;;
    esac
done
```

Output:
```
Choose a color:
1) red
2) green
3) blue
4) quit
#? 
```

## Loop Control

```bash
# break — exit the loop entirely:
for i in {1..100}; do
    if (( i > 5 )); then
        break
    fi
    echo "$i"
done

# continue — skip to next iteration:
for f in *.txt; do
    [[ -e "$f" ]] || continue    # Skip if glob didn't match
    [[ -s "$f" ]] || continue    # Skip empty files
    process "$f"
done

# break N — break out of N nested loops:
for i in {1..3}; do
    for j in {1..3}; do
        if (( i == 2 && j == 2 )); then
            break 2    # Breaks out of BOTH loops
        fi
        echo "$i,$j"
    done
done
```

## Common Patterns

### Retry Loop

```bash
max_attempts=5
attempt=1
while (( attempt <= max_attempts )); do
    if curl -sf "https://api.example.com/health" > /dev/null; then
        echo "Success on attempt $attempt"
        break
    fi
    echo "Attempt $attempt failed, retrying in $((attempt * 2))s..." >&2
    sleep $((attempt * 2))    # Exponential-ish backoff
    ((attempt++))
done

if (( attempt > max_attempts )); then
    echo "All $max_attempts attempts failed" >&2
    exit 1
fi
```

### Parallel Processing (Simple)

```bash
# Run up to N background jobs:
max_jobs=4
for file in *.csv; do
    process_file "$file" &     # Run in background

    # Wait if we have too many background jobs:
    while (( $(jobs -r | wc -l) >= max_jobs )); do
        sleep 0.1
    done
done
wait    # Wait for all remaining background jobs
```

### Accumulating Results

```bash
# Build a comma-separated list:
result=""
for item in alpha beta gamma; do
    result+="${result:+,}$item"    # ${result:+,} adds comma only if result is non-empty
done
echo "$result"    # alpha,beta,gamma

# Collect into an array:
matches=()
for f in /var/log/*.log; do
    if grep -q "ERROR" "$f" 2>/dev/null; then
        matches+=("$f")
    fi
done
echo "Found ${#matches[@]} files with errors"
printf '%s\n' "${matches[@]}"
```

## Common Footguns

### 1. Modifying a Collection While Iterating

```bash
# WRONG — don't delete files while globbing over them in complex ways:
for f in /tmp/session-*; do
    if is_expired "$f"; then
        rm "$f"    # This is actually fine for simple globs
    fi
done
# Simple globs are pre-expanded, so this works. But be careful with find.
```

### 2. Loop Variable Leaking

```bash
for i in 1 2 3; do
    echo "$i"
done
echo "After loop: $i"    # 3 — the variable persists!
# Use a function with local if this matters.
```

### 3. Empty Glob Produces Literal String

```bash
# If no .xyz files exist:
for f in *.xyz; do
    echo "$f"        # Prints literally "*.xyz"
done

# Fix (Bash):
shopt -s nullglob    # Unmatched globs expand to nothing
for f in *.xyz; do
    echo "$f"        # Loop body never executes
done

# Alternate fix:
for f in *.xyz; do
    [[ -e "$f" ]] || continue
    echo "$f"
done
```

## When NOT to Loop

Many loops can be replaced with more efficient operations:

```bash
# SLOW — loop and grep each file:
for f in *.log; do
    grep "ERROR" "$f"
done

# FAST — grep handles multiple files natively:
grep "ERROR" *.log

# SLOW — loop and rename:
for f in *.txt; do
    mv "$f" "${f%.txt}.md"
done

# FAST-ISH — use rename (Fedora has perl-rename):
rename 's/\.txt$/.md/' *.txt

# SLOW — loop and process:
for f in *.csv; do
    wc -l "$f"
done

# FAST — wc handles multiple files:
wc -l *.csv
```

## Exercise

1. Write a loop that reads `/etc/passwd` and prints only users with `/bin/bash` as their shell.

2. Write a retry loop that waits for a file to appear (check every 2 seconds, give up after 30 seconds).

3. Use `find -print0` with a while loop to safely process all `.conf` files under `/etc/` that are readable. Count them.

4. Demonstrate the nullglob problem: show what happens with and without `shopt -s nullglob`.

---

Next: [Script Structure](03-script-structure.md)
