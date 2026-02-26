# Why "$@" Exists and Why People Mess It Up

## What Problem This Solves

You write a wrapper script that passes arguments through to another command. Filenames with spaces break it. Or all your arguments collapse into one string. `"$@"` is the solution, but only if you use it exactly right.

## How People Misunderstand It

1. **"`$*` and `$@` are the same thing"** — They're not. `$*` joins all arguments into one string. `$@` keeps them separate. The difference only matters when quoted.
2. **"I quoted my variables, so I'm fine"** — `"$*"` is quoted but wrong (collapses arguments). `"$@"` is quoted and correct. The `@` matters.
3. **"I'll just use `$1 $2 $3`"** — This breaks on arguments with spaces and limits you to a fixed number of arguments.

## The Mental Model

Think of it as:

```
$*    →  Unquoted: each arg is a separate word (then word-split again)
$@    →  Unquoted: each arg is a separate word (then word-split again)
"$*"  →  ONE string: "arg1 arg2 arg3" (joined by first char of IFS)
"$@"  →  SEPARATE strings: "arg1" "arg2" "arg3" (each arg preserved)
```

`"$@"` is almost always what you want. It preserves each argument as a separate word, including arguments that contain spaces.

### Visual Example

Given a script called with: `./script.sh "hello world" foo "bar baz"`

```
$1 = "hello world"
$2 = "foo"
$3 = "bar baz"

"$@" expands to:  "hello world" "foo" "bar baz"    ← 3 separate args ✅
"$*" expands to:  "hello world foo bar baz"          ← 1 combined string ❌
$@   expands to:  hello world foo bar baz             ← 5 words (split!) ❌
$*   expands to:  hello world foo bar baz             ← 5 words (split!) ❌
```

## Real Examples

### Wrapper Script

The most common use case — wrapping another command:

```bash
#!/usr/bin/env bash
# safe-rm.sh - a wrapper around rm that confirms before deleting

echo "About to delete: $@"    # Shows args, but could miscount with spaces
echo "About to delete: $*"    # Same problem

# The RIGHT way to pass arguments through:
rm -i "$@"

# What happens without "$@":
# ./safe-rm.sh "my file.txt"   → rm -i my file.txt  ← BROKEN: two args
# With "$@":
# ./safe-rm.sh "my file.txt"   → rm -i "my file.txt"  ← CORRECT: one arg
```

### Logging Wrapper

```bash
#!/usr/bin/env bash
# logged-cmd.sh - run a command and log it

log_file="/var/log/commands.log"
echo "[$(date -Iseconds)] Running: $*" >> "$log_file"

# Execute the command with arguments preserved:
"$@"
exit_code=$?

echo "[$(date -Iseconds)] Exit code: $exit_code" >> "$log_file"
exit $exit_code
```

Usage: `./logged-cmd.sh rsync -avz "my folder/" remote:/backup/`

### Function Arguments

Functions use the same `$@` mechanism:

```bash
retry() {
    local max_attempts=3
    local attempt=1
    
    while (( attempt <= max_attempts )); do
        echo "Attempt $attempt..."
        "$@" && return 0    # "$@" runs the command with all its arguments
        ((attempt++))
        sleep 1
    done
    
    echo "Failed after $max_attempts attempts"
    return 1
}

# Usage:
retry curl -s -f "https://example.com/api"
retry rsync -avz "source dir/" "dest dir/"
```

### Iterating Over Arguments

```bash
#!/usr/bin/env bash
# process-files.sh - do something with each argument

# WRONG — breaks on filenames with spaces:
for file in $@; do
    echo "Processing: $file"
done

# RIGHT — each argument preserved:
for file in "$@"; do
    echo "Processing: $file"
done

# Also correct (implicit "$@" in for loops):
for file; do
    echo "Processing: $file"
done
```

## $# — The Argument Count

```bash
#!/usr/bin/env bash

echo "You gave me $# arguments"

if [[ $# -lt 1 ]]; then
    echo "Usage: $0 <filename>"
    exit 1
fi

# $0 is the script name (not counted in $#)
# $1, $2, ... are the arguments
# ${10} and above need braces (otherwise $10 = ${1}0)
```

## shift — Moving Through Arguments

```bash
#!/usr/bin/env bash
# Parsing arguments manually

while [[ $# -gt 0 ]]; do
    case "$1" in
        -v|--verbose)
            verbose=1
            shift       # Move to next argument
            ;;
        -o|--output)
            output="$2" # Next argument is the value
            shift 2     # Move past both flag and value
            ;;
        --)
            shift
            break       # Everything after -- is positional
            ;;
        -*)
            echo "Unknown option: $1"
            exit 1
            ;;
        *)
            args+=("$1")  # Collect positional arguments
            shift
            ;;
    esac
done

# Remaining args after --:
echo "Positional args: ${args[*]}"
```

## Arrays and "$@"

Arrays behave identically to `$@` when expanded with `"${array[@]}"`:

```bash
files=("file one.txt" "file two.txt" "file three.txt")

# WRONG — word splitting breaks filenames:
for f in ${files[@]}; do
    echo "$f"
done
# Outputs: file, one.txt, file, two.txt, file, three.txt (6 iterations!)

# RIGHT — each element preserved:
for f in "${files[@]}"; do
    echo "$f"
done
# Outputs: file one.txt, file two.txt, file three.txt (3 iterations)

# The parallel:
# "${files[@]}"  is to arrays  what  "$@"  is to positional params
# "${files[*]}"  is to arrays  what  "$*"  is to positional params
```

## Common Footguns

**Footgun 1: echo "$@" looks fine but hides problems**
```bash
# echo collapses multiple arguments visually:
set -- "hello world" "foo"
echo "$@"     # hello world foo  — looks like 3 words, but it's 2 args!
echo "$*"     # hello world foo  — looks identical but IS 1 string!

# To see the actual argument structure:
printf '[%s] ' "$@"; echo
# [hello world] [foo]

printf '[%s] ' "$*"; echo
# [hello world foo]
```

**Footgun 2: Assigning "$@" to a variable**
```bash
# You can't store "$@" in a simple variable:
all="$@"     # Collapses to a single string — you lost the boundaries

# Use an array instead:
all=("$@")   # Preserves each argument
"${all[@]}"  # Use like "$@"
```

**Footgun 3: Using "$@" when $# is 0**
```bash
# "$@" with no arguments expands to NOTHING (not an empty string):
set --          # Clear all args
echo "$@"       # Prints empty line (echo with no args)
printf '%s\n' "$@"  # Prints nothing at all
```

## Summary Table

| Expression | With args: `"a b" "c"` | Arguments produced |
|-----------|------------------------|-------------------|
| `$*` | `a b c` (word split) | 3 words: `a`, `b`, `c` |
| `$@` | `a b c` (word split) | 3 words: `a`, `b`, `c` |
| `"$*"` | `"a b c"` | 1 string: `a b c` |
| `"$@"` | `"a b" "c"` | 2 strings: `a b` and `c` |

## Exercise

1. Write a wrapper script `safe-delete.sh` that prints what it will delete, asks for confirmation, then passes all arguments to `rm`. Test with `./safe-delete.sh "my file.txt" "other file.txt"`.

2. Write a function `run_as` that takes a username as `$1` and runs the remaining arguments as that user:
   ```bash
   run_as postgres psql -c "SELECT 1"
   ```
   Hint: use `shift` and `"$@"`.

3. Write a script that counts its arguments, then prints each one with `printf '[%s]\n' "$@"`. Test with arguments containing spaces, quotes, and empty strings.

---

Next: [Level 3: Streams — stdin, stdout, stderr](../03-pipes-redirection-fds/00-streams.md)
