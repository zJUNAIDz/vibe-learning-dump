# Building Real Pipelines

## What Problem This Solves

You know grep, sed, awk, and jq individually. But real work means **combining** them. This file teaches you to think in data flow, build pipelines incrementally, and avoid the common traps.

## Pipeline Thinking

A pipeline is a data factory. Each stage takes input, transforms it, and passes it to the next stage:

```
Raw Data → Filter → Transform → Aggregate → Format → Output
```

**Build pipelines ONE STAGE AT A TIME.** Run each stage, inspect the output, then add the next.

```bash
# DON'T write this all at once:
cat access.log | grep "POST" | awk '{print $1}' | sort | uniq -c | sort -rn | head

# DO build incrementally:
head access.log                                    # Step 1: see the data
grep "POST" access.log | head                      # Step 2: filter
grep "POST" access.log | awk '{print $1}' | head   # Step 3: extract
# ... add more stages once each works
```

## The Core Unix Text Tools

Beyond grep, sed, awk, you should know these:

### sort

```bash
sort file                    # Alphabetical sort
sort -n file                 # Numeric sort
sort -r file                 # Reverse
sort -k2 file                # Sort by 2nd column
sort -k2,2n -k1,1 file      # By 2nd column (numeric), then 1st
sort -t: -k3 -n /etc/passwd  # Sort by UID
sort -u file                 # Sort and deduplicate
sort -h file                 # Human-readable numbers (1K, 2M, 3G)
```

### uniq (requires sorted input!)

```bash
sort file | uniq             # Remove adjacent duplicates
sort file | uniq -c          # Count occurrences
sort file | uniq -d          # Show only duplicates
sort file | uniq -u          # Show only unique lines
sort file | uniq -c | sort -rn  # Top frequencies
```

### cut

```bash
cut -d: -f1 /etc/passwd     # First field (colon-separated)
cut -d: -f1,7 /etc/passwd   # Fields 1 and 7
cut -c1-10 file              # First 10 characters per line
cut -d' ' -f3- file          # Fields 3 onwards
```

### tr (translate/delete characters)

```bash
tr 'a-z' 'A-Z' < file       # Uppercase
tr -d '[:digit:]' < file     # Delete all digits
tr -s ' ' < file             # Squeeze multiple spaces to one
tr '\n' ' ' < file           # Join all lines with spaces
echo "hello world" | tr ' ' '\n'  # Split into lines
```

### paste

```bash
paste file1 file2            # Merge files side by side (tab-separated)
paste -d, file1 file2       # Merge with comma
paste -s file                # Sequential — all lines into one row
```

### tee

```bash
# Save intermediate output while continuing the pipeline:
command | tee /tmp/debug.txt | next_command

# Append:
command | tee -a logfile | next_command

# Write to multiple files:
command | tee file1 file2 | next_command
```

### head / tail

```bash
head -20 file                # First 20 lines
tail -20 file                # Last 20 lines
tail -f /var/log/messages    # Follow — show new lines as they appear
tail -f -n 0 logfile         # Follow, starting from NOW (no historical lines)

# All but the first line (remove header):
tail -n +2 file

# All but the last 5 lines:
head -n -5 file
```

### wc

```bash
wc -l file                   # Count lines
wc -w file                   # Count words
wc -c file                   # Count bytes
wc -m file                   # Count characters
```

### xargs

```bash
# Run a command for each line of input:
find /tmp -name "*.log" | xargs rm -f

# With -print0 for safety:
find /tmp -name "*.log" -print0 | xargs -0 rm -f

# One at a time:
find . -name "*.sh" | xargs -n 1 shellcheck

# Parallel:
find . -name "*.jpg" | xargs -P 4 -n 1 convert_image

# Use as argument in specific position:
find . -name "*.bak" | xargs -I{} mv {} /backup/
```

## Real Pipeline Examples

### 1. Web Server Log Analysis

```bash
# Top 10 IPs by request count:
awk '{print $1}' access.log | sort | uniq -c | sort -rn | head -10

# Requests per hour:
awk '{print substr($4, 2, 14)}' access.log | \
  cut -d: -f1-2 | sort | uniq -c

# 404 errors with their URLs:
awk '$9 == 404 {print $7}' access.log | sort | uniq -c | sort -rn | head -20

# Bandwidth by URL (sum bytes):
awk '{urls[$7] += $10} END {
    for (url in urls) printf "%10d %s\n", urls[url], url
}' access.log | sort -rn | head -20

# Slowest endpoints:
awk '{print $NF, $7}' access.log | sort -rn | head -10
```

### 2. System Investigation

```bash
# Which processes are using the most memory:
ps aux --sort=-%mem | head -11 | awk 'NR>1 {printf "%-8s %6.1f%% %s\n", $1, $4, $11}'

# Open files by process:
ls -l /proc/*/fd 2>/dev/null | grep -v "total\|^l" | \
  awk '{print $NF}' | sort | uniq -c | sort -rn | head

# Failed SSH logins:
journalctl -u sshd --since "1 hour ago" --no-pager | \
  grep "Failed password" | \
  awk '{print $(NF-3)}' | sort | uniq -c | sort -rn

# Disk usage — largest directories:
du -h --max-depth=2 /var 2>/dev/null | sort -rh | head -20

# Find recently modified config files:
find /etc -name "*.conf" -mtime -7 -ls 2>/dev/null | \
  awk '{print $NF}' | sort
```

### 3. Data Processing

```bash
# CSV: Average of column 3, grouped by column 1:
awk -F, 'NR>1 {sum[$1]+=$3; count[$1]++} END {
    for (k in sum) printf "%s: %.2f\n", k, sum[k]/count[k]
}' data.csv | sort

# Find duplicate lines in a file:
sort file | uniq -d

# Diff two sorted lists (what's in A but not B):
comm -23 <(sort list_a.txt) <(sort list_b.txt)

# Merge CSV files (same headers):
head -1 file1.csv > merged.csv
tail -n +2 -q file*.csv >> merged.csv

# Word frequency analysis:
tr -s '[:space:][:punct:]' '\n' < document.txt | \
  tr 'A-Z' 'a-z' | \
  sort | uniq -c | sort -rn | head -20
```

### 4. DevOps Pipelines

```bash
# Git: files changed most often (hotspots):
git log --pretty=format: --name-only | \
  grep -v '^$' | sort | uniq -c | sort -rn | head -20

# Docker: images over 500MB:
docker images --format '{{.Repository}}:{{.Tag}}\t{{.Size}}' | \
  awk -F'\t' '{
    size = $2
    if (size ~ /GB/ || (size ~ /MB/ && substr(size,1,length(size)-2) > 500))
      print size "\t" $1
  }' | sort -rh

# Kubernetes: pods with restart counts:
kubectl get pods -o json | jq -r '
  .items[] | 
  .metadata.name as $pod |
  .status.containerStatuses[]? | 
  select(.restartCount > 0) |
  "\($pod)\t\(.name)\t\(.restartCount)"
' | sort -t$'\t' -k3 -rn | column -t

# Environment diff between two servers:
diff <(ssh server1 env | sort) <(ssh server2 env | sort)
```

## Pipeline Debugging

### Use tee to Inspect Stages

```bash
# Insert tee at any point to see intermediate data:
grep "ERROR" app.log | \
  tee /dev/stderr | \            # Shows filtered lines on terminal
  awk '{print $4}' | \
  tee /tmp/debug_after_awk.txt | \  # Saves to file
  sort | uniq -c | sort -rn
```

### Use head Early

```bash
# When developing, use head to limit data:
big_pipeline_first_stage | head -5 | second_stage | head -5
# Once it works, remove the heads
```

### Check Exit Codes

```bash
# After a pipeline:
echo "${PIPESTATUS[@]}"
# Shows exit code of each command in the pipeline

# Example:
false | true | false
echo "${PIPESTATUS[@]}"    # 1 0 1
```

## Performance Considerations

### When Pipelines Are Slow

```bash
# SLOW — spawning many processes:
for f in *.txt; do
    grep -l "pattern" "$f"    # One grep per file
done

# FAST — let grep handle multiple files:
grep -rl "pattern" *.txt

# SLOW — sort huge file then uniq:
sort gigantic.log | uniq -c

# FASTER — use awk to count (no sorting needed):
awk '{count[$0]++} END {for(k in count) print count[k], k}' gigantic.log

# SLOW — grep | grep | grep:
grep "ERROR" log | grep "database" | grep -v "timeout"

# FASTER — single grep with regex:
grep -E 'ERROR.*database' log | grep -v "timeout"
# FASTEST — single awk:
awk '/ERROR/ && /database/ && !/timeout/' log
```

### Right Tool for the Job

```bash
# For HUGE files — consider:
# - ripgrep (rg) instead of grep: 3-10x faster
# - GNU parallel instead of xargs for CPU-bound work
# - awk instead of complex pipe chains
# - jq --stream for huge JSON files

# ripgrep example:
rg "ERROR" /var/log/ --glob '*.log'
```

## Anti-Patterns to Avoid

```bash
# 1. Useless cat:
cat file | grep "x"         # WRONG
grep "x" file               # RIGHT

# 2. grep | awk (awk already filters):
grep "error" file | awk '{print $3}'    # Redundant
awk '/error/ {print $3}' file           # Single tool

# 3. Parsing ls:
for f in $(ls *.txt); do ...    # BROKEN on spaces
for f in *.txt; do ...          # Correct

# 4. echo | sed for string ops (use parameter expansion):
result=$(echo "$str" | sed 's/old/new/')    # Spawns a process
result="${str/old/new}"                      # Pure bash — instant

# 5. while read with echo:
echo "$data" | while read line; do ...    # Subshell problem
while read line; do ... done <<< "$data"  # Better
```

## Exercise

1. Build a pipeline that analyzes `journalctl` output: count log messages by unit (service name) for the last hour, showing the top 10 noisiest services.

2. Write a pipeline that finds all shell scripts (by shebang, not extension) under `/usr/bin` and groups them by which shell they use (bash, sh, python, etc.).

3. Create a "git diff summary" pipeline: for the current repo, show the number of insertions and deletions per file for the last 10 commits.

4. Build an incremental pipeline in front of someone: start with raw data, add one stage at a time, explaining what each does. Use tee to debug at least one stage.

---

Next: [Level 8 — Zsh Power Features](../08-zsh-power/00-zsh-vs-bash.md)
