# Word Splitting: Why Spaces Break Everything

## What Problem This Solves

This is the single most common source of shell bugs. You write `rm $file` and it works fine until `$file` contains a space. Then it deletes the wrong things. Understanding word splitting transforms you from someone who gets bitten to someone who prevents it.

## How People Misunderstand It

1. **"Variables just get replaced by their value"** — Not exactly. After substitution, the result gets *split into words* and *glob-expanded*. The value goes through further processing.
2. **"It works without quotes on my machine"** — It works until a filename has a space, a glob character, or your locale changes.
3. **"I'll just avoid spaces in filenames"** — You can't control other people's filenames. Logs, downloads, mounted drives, and user-created files regularly have spaces.

## The Mental Model

When the shell sees an **unquoted** variable, it performs two extra steps:

```
Step 1: Parameter expansion
  $file  →  "my document.txt"

Step 2: Word splitting (ONLY on unquoted results)
  "my document.txt"  →  two words: "my" and "document.txt"

Step 3: Pathname expansion (globbing) (ONLY on unquoted results)
  If either word matches a glob pattern, expand it
```

**Double-quoting prevents steps 2 and 3.** That's the entire reason `"$file"` works and `$file` doesn't.

### What "Words" Mean to the Shell

The shell doesn't see characters — it sees **words**. A word is a unit that becomes one argument to a command. The shell splits input into words using `$IFS` (Internal Field Separator), which defaults to: space, tab, newline.

```bash
file="my document.txt"

# Without quotes — shell splits on spaces:
ls $file
# ls receives TWO arguments: "my" and "document.txt"
# Equivalent to: ls my document.txt
# Result: ls: cannot access 'my': No such file or directory
#         ls: cannot access 'document.txt': No such file or directory

# With quotes — no splitting:
ls "$file"
# ls receives ONE argument: "my document.txt"
# Result: shows the file correctly
```

### Visualizing the Problem

```bash
file="my document.txt"

# What you THINK happens:
ls $file  →  ls my document.txt    # One argument with spaces

# What ACTUALLY happens:
ls $file  →  ls "my" "document.txt"  # Two separate arguments

# The fix:
ls "$file"  →  ls "my document.txt"  # One argument, preserved
```

## The IFS Variable

Word splitting uses `$IFS` to decide where to split:

```bash
# Default IFS: space, tab, newline
echo -n "$IFS" | xxd
# Outputs: 20 09 0a  (space=0x20, tab=0x09, newline=0x0a)

# You can change IFS:
IFS=':'
line="one:two:three"
for word in $line; do   # Now splits on colons
    echo "[$word]"
done
# [one]
# [two]
# [three]

# ALWAYS restore IFS or use a subshell:
(
    IFS=':'
    # Splitting on colons here
)
# IFS is back to normal here
```

### Reading /etc/passwd with IFS

```bash
while IFS=: read -r user _ uid gid _ home shell; do
    echo "$user (uid=$uid) home=$home shell=$shell"
done < /etc/passwd
```

Here `IFS=:` is set only for the `read` command (using the `VAR=value cmd` syntax), so it splits each line on colons.

## Word Splitting in Practice

### The mv Disaster

```bash
# Imagine these files exist:
# "backup 2024.tar.gz"
# "notes.txt"

for f in $(ls); do
    mv "$f" /backup/
done

# $(ls) output: "backup 2024.tar.gz\nnotes.txt"
# Word splitting produces: "backup" "2024.tar.gz" "notes.txt"
# mv tries to move THREE files instead of TWO
# "backup" and "2024.tar.gz" don't exist individually → errors
```

### The rm Catastrophe

```bash
dir="/home/user/my projects"

# DANGER:
rm -rf $dir
# Word splits into: rm -rf /home/user/my projects
# rm deletes /home/user/my (if it exists!) and then "projects" in the current dir
# This is how people accidentally delete directories

# SAFE:
rm -rf "$dir"
# rm deletes: /home/user/my projects (as intended)
```

### Empty Variables and Word Splitting

```bash
file=""

# Without quotes — empty variable disappears entirely:
ls $file
# Becomes: ls (no arguments — lists current directory!)

# With quotes — empty string passed as argument:
ls "$file"
# Becomes: ls "" — ls tries to open a file named "" (error, but predictable)
```

This is why `[ $var = "test" ]` breaks when `$var` is empty:
```bash
var=""
[ $var = "test" ]
# Becomes: [ = "test" ] — syntax error!

[ "$var" = "test" ]
# Becomes: [ "" = "test" ] — works correctly, evaluates to false
```

## Common Footguns

**Footgun 1: Looping over `$(ls)` or `$(find)`**
```bash
# BROKEN — word-splits on spaces in filenames:
for f in $(find . -name '*.txt'); do
    echo "$f"
done

# CORRECT — use process substitution or find -exec:
find . -name '*.txt' -print0 | while IFS= read -r -d '' f; do
    echo "$f"
done

# Or simply:
find . -name '*.txt' -exec echo {} \;
```

**Footgun 2: Unquoted command substitution**
```bash
# BROKEN — word splitting on spaces in results:
contents=$(cat file.txt)
echo $contents       # Loses formatting, collapses whitespace

echo "$contents"     # Preserves formatting
```

**Footgun 3: Glob characters in variables**
```bash
file="*.txt"
echo $file          # Expands to matching files! Not the literal string!
echo "$file"        # Prints: *.txt

# Even worse:
var="Price is $5 for 3*2 items"
echo $var           # $5 → empty (no $5 variable), 3*2 might glob-expand
echo "$var"         # Exact string (well, $5 still expands — need single quotes for that)
```

## The Rule

**Quote every variable expansion unless you have a specific reason not to.**

```bash
"$variable"          # Always quote
"${variable}"        # Always quote
"$(command)"         # Always quote
"${array[@]}"        # Always quote

# The ONLY time you intentionally skip quotes:
# When you WANT word splitting (rare):
flags="-l -a -h"
ls $flags            # Intentionally split into: ls -l -a -h
```

## Exercise

1. Create a file named `my file.txt` with some content. Try `cat $f` (where `f="my file.txt"`) and observe the error. Then try `cat "$f"`.
2. Set `var="*.md"` and run `echo $var` vs `echo "$var"`. In a directory with markdown files, observe the difference.
3. Set `empty=""` and run `[ $empty = "test" ]` vs `[ "$empty" = "test" ]`. Observe the error.
4. Write a script that reads filenames from `find` output and safely handles spaces (use `-print0` and `read -d ''`).

---

Next: [Quoting Rules: Single, Double, None](01-quoting-rules.md)
