# grep, sed, and awk — The Big Three

## What Problem This Solves

You need to find, filter, transform, or extract data from text. These three tools handle 95% of text processing you'll ever need from the shell. The problem is knowing WHICH one to use WHEN.

## When to Use Which

```
┌───────────┬────────────────────────┬─────────────────────────────┐
│ Tool      │ Primary Purpose        │ Use When You Need To...     │
├───────────┼────────────────────────┼─────────────────────────────┤
│ grep      │ Find matching lines    │ Search, filter, check if    │
│           │                        │ a pattern exists            │
├───────────┼────────────────────────┼─────────────────────────────┤
│ sed       │ Transform text         │ Find & replace, delete      │
│           │ (stream editor)        │ lines, simple transforms    │
├───────────┼────────────────────────┼─────────────────────────────┤
│ awk       │ Process structured     │ Extract columns, compute,   │
│           │ text (mini language)   │ conditional formatting      │
└───────────┴────────────────────────┴─────────────────────────────┘
```

Quick decision:
- "Does this line contain X?" → **grep**
- "Replace X with Y" → **sed**
- "Give me the 3rd column" → **awk**
- "Compute a sum" → **awk**
- "Reshape data" → **awk**

---

## grep — Search and Filter

### Basic Usage

```bash
# Find lines containing a string:
grep "error" /var/log/messages

# Case-insensitive:
grep -i "error" /var/log/messages

# Show line numbers:
grep -n "error" /var/log/messages

# Count matches:
grep -c "error" /var/log/messages

# Invert — show lines that DON'T match:
grep -v "debug" /var/log/messages

# Recursive search in a directory:
grep -r "TODO" src/

# Only filenames (not content):
grep -rl "TODO" src/

# Show context (lines before/after match):
grep -B 2 -A 5 "panic" /var/log/messages    # 2 before, 5 after
grep -C 3 "panic" /var/log/messages          # 3 before AND after
```

### Extended Regex (`-E` or `egrep`)

```bash
# Without -E, you must escape special characters:
grep 'error\|warning\|critical' logfile    # Yikes

# With -E (extended regex) — much cleaner:
grep -E 'error|warning|critical' logfile

# Common regex patterns:
grep -E '^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' access.log  # IP addresses
grep -E '^[[:space:]]*$' file.txt    # Blank lines
grep -E '^\s*#' /etc/ssh/sshd_config    # Comment lines
```

### Perl-Compatible Regex (`-P`)

```bash
# Lookahead, lookbehind, non-greedy matching:
grep -P '(?<=password=)\S+' config       # Extract value after password=
grep -P 'error(?!.*ignored)' log         # "error" not followed by "ignored"
grep -Po '"name":\s*"\K[^"]+' data.json  # Extract JSON values (hacky but works)
```

### Useful grep Patterns

```bash
# Fixed string (not regex — faster for literal text):
grep -F "192.168.1.1" access.log

# Whole word match:
grep -w "error" log    # Matches "error" but not "errors" or "myerror"

# Binary files — suppress "Binary file matches":
grep -a "text" binary_file     # Treat binary as text
grep -I "text" *               # Skip binary files

# Quiet mode — just check if match exists:
if grep -q "error" logfile; then
    echo "Errors found!"
fi

# Multiple patterns:
grep -e "error" -e "warning" logfile
# or from a file:
grep -f patterns.txt logfile
```

---

## sed — Stream Editor

### Mental Model

sed processes text line by line. For each line, it applies your commands and prints the result. It does NOT modify the file unless you use `-i`.

```
Input → [sed command] → Output
  ↓         ↓              ↓
line 1 → transform → modified line 1
line 2 → transform → modified line 2
...
```

### Find and Replace

```bash
# Basic substitution (first match per line):
sed 's/old/new/' file

# Global substitution (all matches per line):
sed 's/old/new/g' file

# Case-insensitive:
sed 's/error/WARNING/gi' file    # GNU sed only

# Replace only on specific lines:
sed '3s/old/new/' file           # Only line 3
sed '1,5s/old/new/g' file       # Lines 1 through 5
sed '/pattern/s/old/new/g' file  # Only lines matching pattern

# Delete lines:
sed '/^#/d' file                 # Delete comment lines
sed '/^$/d' file                 # Delete blank lines
sed '1d' file                    # Delete first line

# Print specific lines:
sed -n '5p' file                 # Print only line 5
sed -n '10,20p' file             # Print lines 10-20
sed -n '/start/,/end/p' file    # Print between patterns
```

### In-Place Editing

```bash
# Modify file directly:
sed -i 's/old/new/g' file

# With backup (SAFER):
sed -i.bak 's/old/new/g' file    # Creates file.bak

# Fedora/GNU sed supports -i without arg.
# macOS/BSD sed requires -i '' (empty backup suffix).
# For portability: sed -i.bak '...' file && rm file.bak
```

### Delimiter Tricks

```bash
# Default delimiter is /, but you can use anything:
sed 's/path\/to\/file/new\/path/g' file    # Ugly with /
sed 's|path/to/file|new/path|g' file       # Clean with |
sed 's#/usr/local/bin#/opt/bin#g' file     # Clean with #
```

### Multi-Command and Advanced sed

```bash
# Multiple commands:
sed -e 's/foo/bar/g' -e 's/baz/qux/g' file

# Or with semicolons:
sed 's/foo/bar/g; s/baz/qux/g' file

# Insert/append:
sed '1i\# This is a header' file           # Insert before line 1
sed '$a\# This is a footer' file           # Append after last line

# Capture groups:
sed -E 's/([0-9]+)-([0-9]+)/\2-\1/g' file    # Swap: 123-456 → 456-123

# Transform (transliterate, like tr):
sed 'y/abc/ABC/' file    # a→A, b→B, c→C
```

---

## awk — The Mini Language

### Mental Model

awk splits each line into **fields** and lets you process them. It's a pattern-action language:

```
awk 'PATTERN { ACTION }' file
```

For each line: if PATTERN matches, execute ACTION.

```
Input line: "alice 25 engineering"
            $0 = entire line
            $1 = "alice"
            $2 = "25"  
            $3 = "engineering"
            NF = 3 (number of fields)
            NR = line number
```

### Basic Column Extraction

```bash
# Print specific columns:
awk '{print $1}' /etc/passwd              # First field (wrong — : separated!)
awk -F: '{print $1}' /etc/passwd          # First field with correct separator
awk -F: '{print $1, $7}' /etc/passwd      # Username and shell

# Print last column:
awk '{print $NF}' file

# Print second-to-last:
awk '{print $(NF-1)}' file
```

### Patterns and Conditions

```bash
# Only print if condition is true:
awk '$3 > 1000' /etc/passwd              # Lines where field 3 > 1000 (wrong separator)
awk -F: '$3 > 1000 {print $1, $3}' /etc/passwd    # Users with UID > 1000

# Pattern matching:
awk '/error/ {print $0}' logfile         # Lines containing "error"
awk '!/comment/ {print}' file            # Lines NOT containing "comment"

# Range pattern:
awk '/START/,/END/' file                 # Print lines between START and END

# BEGIN and END blocks:
awk 'BEGIN {print "Name,UID"} -F: {print $1","$3} END {print "Done"}' /etc/passwd
```

### Built-in Variables

```bash
# NR — current line number:
awk '{print NR": "$0}' file              # Number all lines

# NF — number of fields on current line:
awk 'NF > 3' file                        # Lines with more than 3 fields

# FS — field separator (same as -F):
awk 'BEGIN {FS=":"} {print $1}' /etc/passwd

# OFS — output field separator:
awk -F: 'BEGIN {OFS=","} {print $1, $3, $7}' /etc/passwd    # CSV output

# RS — record separator (default: newline)
# ORS — output record separator (default: newline)
```

### Computation

```bash
# Sum a column:
awk '{sum += $1} END {print sum}' numbers.txt

# Average:
awk '{sum += $1; count++} END {print sum/count}' numbers.txt

# Max:
awk 'BEGIN {max=0} $1 > max {max=$1} END {print max}' numbers.txt

# Count occurrences:
awk '{count[$1]++} END {for (k in count) print k, count[k]}' access.log

# Sum disk usage by directory:
du -sh /var/*/ 2>/dev/null | awk '{sum += $1} END {printf "Total: %.1fG\n", sum}'
```

### Formatting

```bash
# printf for formatted output:
awk -F: '{printf "%-20s UID=%-6s Shell=%s\n", $1, $3, $7}' /etc/passwd

# Conditional formatting:
awk -F: '{
    if ($3 == 0) 
        printf "\033[31m%-20s ROOT\033[0m\n", $1
    else if ($3 < 1000) 
        printf "%-20s system\n", $1
    else 
        printf "%-20s user\n", $1
}' /etc/passwd
```

---

## Putting Them Together

### Pipeline Patterns

```bash
# Find errors, extract timestamp and message:
grep "ERROR" app.log | awk '{print $1, $2, $NF}'

# Count HTTP status codes:
awk '{print $9}' access.log | sort | uniq -c | sort -rn | head

# Replace in files found by grep:
grep -rl "old_function" src/ | xargs sed -i 's/old_function/new_function/g'

# Extract unique IPs from log:
grep -oE '[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+' access.log | sort -u

# Find large files, format nicely:
find /var -size +100M -exec ls -lh {} \; 2>/dev/null | awk '{print $5, $9}' | sort -rh
```

### Common Anti-Patterns

```bash
# WRONG — grep + awk when awk alone works:
grep "error" file | awk '{print $3}'
# RIGHT:
awk '/error/ {print $3}' file

# WRONG — cat + grep:
cat file | grep "pattern"
# RIGHT:
grep "pattern" file

# WRONG — grep + wc when grep -c works:
grep "error" file | wc -l
# RIGHT:
grep -c "error" file

# WRONG — complex sed when awk is clearer:
sed 's/.*name="\([^"]*\)".*/\1/' file
# RIGHT (arguably clearer):
awk -F'"' '/name=/ {print $2}' file
```

## Exercise

1. From `/etc/passwd`, extract all usernames with UID ≥ 1000, sorted alphabetically.

2. Use sed to remove all comments and blank lines from a config file (like `/etc/ssh/sshd_config`), showing only active configuration lines.

3. Write an awk one-liner that reads a CSV and computes the sum and average of a numeric column.

4. Build a pipeline that finds the top 10 most frequent words in a text file (hint: tr, sort, uniq, head).

---

Next: [awk Deep Dive](01-awk-deep-dive.md)
