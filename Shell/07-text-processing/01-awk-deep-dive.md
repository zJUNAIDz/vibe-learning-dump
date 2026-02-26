# awk Deep Dive

## Why awk Gets Its Own File

grep finds. sed transforms. awk **processes**. It's a full programming language with variables, arrays, control flow, and functions. When a pipeline of grep/sed/cut/tr gets unreadable, a single awk command usually replaces it all.

## awk Program Structure

```
awk 'BEGIN { setup } PATTERN { action } END { summary }' input
```

- **BEGIN** — runs once before any input is read
- **PATTERN { action }** — runs for each line that matches pattern
- **END** — runs once after all input is processed
- If no pattern → action runs for every line
- If no action → matching line is printed

```bash
awk '
BEGIN { 
    print "=== Report ===" 
}
/error/ { 
    errors++
    print NR": "$0 
}
END { 
    printf "Total errors: %d\n", errors 
}
' /var/log/messages
```

## Field Processing

### Custom Separators

```bash
# Input field separator:
awk -F: '{print $1}' /etc/passwd        # Colon-separated
awk -F'\t' '{print $2}' data.tsv        # Tab-separated
awk -F',' '{print $3}' data.csv         # CSV (naive — doesn't handle quoted commas)
awk -F'[ \t]+' '{print $1}' file        # Multiple spaces/tabs (regex separator)

# Output field separator:
awk -F: 'BEGIN{OFS=","} {print $1,$3,$7}' /etc/passwd
# Output: root,0,/bin/bash

# Change OFS — must force record rebuild:
awk -F: 'BEGIN{OFS="\t"} {$1=$1; print}' /etc/passwd    # $1=$1 triggers rebuild
```

### Modifying Fields

```bash
# Change a field:
awk -F: 'BEGIN{OFS=":"} $7=="/bin/bash" {$7="/bin/zsh"; print}' /etc/passwd

# Add a field:
awk -F: 'BEGIN{OFS=":"} {$(NF+1)="added"; print}' /etc/passwd

# Delete a field (shift remaining):
awk '{for(i=2;i<=NF;i++) printf "%s%s",$i,(i<NF?OFS:ORS)}' file
```

## Variables and Types

awk variables are untyped — they're strings or numbers depending on context:

```bash
awk 'BEGIN {
    x = "42"        # String
    y = x + 0        # Now numeric: 42
    z = x "hello"    # String concatenation: "42hello"
    
    # Uninitialized variables are "" (empty string) or 0 (numeric)
    print count + 1   # 1 (count was 0)
    print name ""     # "" (name was empty)
}'
```

### Important Built-in Variables

```bash
# Per-line:
# $0       — entire current line
# $1..$NF  — individual fields
# NF       — number of fields in current line
# NR       — current line number (total across all files)
# FNR      — line number in current file
# FILENAME — current input filename

# Global:
# FS       — input field separator
# OFS      — output field separator
# RS       — input record separator (default: newline)
# ORS      — output record separator (default: newline)

# Example: process multiple files with FNR:
awk 'FNR==1 {print "=== " FILENAME " ==="}; {print NR, FNR, $0}' file1 file2
```

## Arrays (Associative)

awk arrays are **always associative** (like hashmaps/dictionaries). There are no indexed arrays — even `a[0]` uses "0" as a string key.

```bash
# Count occurrences:
awk '{count[$1]++} END {for (k in count) print k, count[k]}' access.log

# Sum by category:
awk -F, '{total[$1] += $2} END {for (cat in total) printf "%s: %.2f\n", cat, total[cat]}' sales.csv

# Check existence:
awk '{
    if ($1 in seen) {
        print "DUPLICATE:", $0
    } else {
        seen[$1] = 1
    }
}' file

# Delete an element:
awk '{a[$1]++} END {delete a["unwanted"]; for(k in a) print k, a[k]}' file

# Multi-dimensional (simulated with SUBSEP):
awk '{
    data[$1,$2] += $3    # Key is "$1\034$2" internally
} END {
    for (key in data) {
        split(key, parts, SUBSEP)
        printf "%s, %s: %d\n", parts[1], parts[2], data[key]
    }
}' data.txt
```

## Control Flow

```bash
awk '{
    # if/else:
    if ($3 > 100) {
        print "HIGH:", $0
    } else if ($3 > 50) {
        print "MED:", $0
    } else {
        print "LOW:", $0
    }
    
    # Ternary:
    status = ($3 > 100) ? "HIGH" : "LOW"
    
    # for loop:
    for (i = 1; i <= NF; i++) {
        printf "Field %d: %s\n", i, $i
    }
    
    # while loop:
    i = 1
    while (i <= 5) {
        print i
        i++
    }
    
    # next — skip to next input line:
    if (/^#/) next    # Skip comments
    
    # exit — stop processing:
    if (NR > 1000) exit
}' file
```

## Functions

### Built-in String Functions

```bash
awk '{
    # Length:
    print length($0)           # Length of line
    print length($1)           # Length of first field
    
    # Substring:
    print substr($0, 1, 10)    # First 10 chars
    
    # Split:
    n = split($0, parts, ":")  # Split into array, returns count
    
    # Find:
    pos = index($0, "error")   # Position of "error" (0 if not found)
    
    # Replace:
    sub(/old/, "new")          # Replace first match in $0
    gsub(/old/, "new")         # Replace ALL matches in $0
    gsub(/old/, "new", $3)     # Replace in specific field
    
    # Case:
    print tolower($1)
    print toupper($1)
    
    # Match (regex):
    if (match($0, /[0-9]+/)) {
        print "Found number at position", RSTART, "length", RLENGTH
        print substr($0, RSTART, RLENGTH)    # Extract the match
    }
}' file
```

### Built-in Math Functions

```bash
awk 'BEGIN {
    print int(3.7)       # 3 (truncate)
    print sqrt(144)      # 12
    print log(1)         # 0 (natural log)
    print exp(1)         # 2.71828
    print sin(3.14159)   # ~0
    print rand()         # Random number 0-1
    srand()              # Seed random number generator
}'
```

### User-Defined Functions

```bash
awk '
function max(a, b) {
    return (a > b) ? a : b
}

function trim(s) {
    gsub(/^[[:space:]]+|[[:space:]]+$/, "", s)
    return s
}

function human_size(bytes,    suffix, i) {
    split("B,KB,MB,GB,TB", suffix, ",")
    i = 1
    while (bytes >= 1024 && i < 5) {
        bytes /= 1024
        i++
    }
    return sprintf("%.1f%s", bytes, suffix[i])
}

{
    print max($1, $2)
    print trim("  hello  ")
    print human_size(1073741824)    # "1.0GB"
}
' data.txt
```

## Real-World Examples

### Log Analysis

```bash
# Top 10 IPs in access log (Apache/Nginx):
awk '{print $1}' access.log | sort | uniq -c | sort -rn | head -10

# OR entirely in awk:
awk '{ip[$1]++} END {
    for (i in ip) print ip[i], i
}' access.log | sort -rn | head -10

# Requests per hour:
awk '{
    split($4, t, "[:/]")
    hour = t[4]
    hits[hour]++
} END {
    for (h in hits) printf "%s:00 → %d requests\n", h, hits[h]
}' access.log | sort

# Slow requests (response time > 1 second):
awk '$NF > 1.0 {print $7, $NF"s"}' access.log | sort -t's' -k2 -rn | head -20
```

### System Administration

```bash
# Disk usage by user (from du output):
du -sh /home/*/ 2>/dev/null | awk '{print $2, $1}' | sort -k2 -h

# Process memory usage summary:
ps aux --no-headers | awk '{mem[$1] += $6} END {
    for (user in mem) printf "%-15s %10.1f MB\n", user, mem[user]/1024
}' | sort -k2 -rn

# Network connections by state:
ss -tan | awk 'NR>1 {state[$1]++} END {for(s in state) print s, state[s]}' | sort -k2 -rn

# Find duplicate files by size (first pass):
find /path -type f -exec ls -l {} + | awk '{size[$5] = size[$5] " " $NF} END {
    for (s in size) {
        n = split(size[s], files, " ")
        if (n > 1) {
            printf "Size %s bytes: ", s
            for (i in files) printf "%s ", files[i]
            print ""
        }
    }
}'
```

### Data Transformation

```bash
# CSV to JSON (simple — no quoting):
awk -F, 'NR==1 {
    for(i=1;i<=NF;i++) headers[i]=$i; next
} {
    printf "{"
    for(i=1;i<=NF;i++) {
        printf "\"%s\":\"%s\"", headers[i], $i
        if (i<NF) printf ","
    }
    print "}"
}' data.csv

# Transpose rows to columns:
awk '{
    for(i=1;i<=NF;i++) a[i][NR]=$i
} END {
    for(i=1;i<=NF;i++) {
        for(j=1;j<=NR;j++) printf "%s%s", a[i][j], (j<NR?"\t":"\n")
    }
}' file
```

## awk vs Other Tools

| Task | Don't | Do |
|------|-------|-----|
| Extract a column | `cut -d: -f1` (fragile with multi-char delims) | `awk -F: '{print $1}'` |
| Count pattern | `grep -c "x"` | `grep -c "x"` (grep is fine here) |
| Simple replace | `awk '{gsub(...)}` | `sed 's/old/new/g'` (sed is simpler) |
| Compute sums | An entire Python script | `awk '{s+=$1} END {print s}'` |
| Complex logic | Chaining 5 tools in a pipe | Single awk program |

**Rule of thumb:** If your pipeline has more than 3 `|` operators, consider rewriting it as a single awk program.

## Exercise

1. Using awk, analyze `/etc/passwd`: group users by their shell and count how many use each shell.

2. Write an awk script that reads a CSV file with headers and prints it as a formatted, aligned table.

3. Parse `ss -tan` output and build a report showing: total connections, connections per state, and connections per local port.

4. Write a log analyzer in awk: read a log file in "timestamp level message" format, count by level, report the first and last timestamp, and print error messages.

---

Next: [jq and JSON Processing](02-jq-json-processing.md)
