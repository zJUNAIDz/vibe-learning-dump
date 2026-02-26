# set -euo pipefail Explained

## What Problem This Solves

Scripts that silently continue after errors cause catastrophic damage. `set -euo pipefail` is the "strict mode" header that catches common failure patterns. But using it without understanding it creates its own class of bugs.

## How People Misunderstand It

1. **"Just paste `set -euo pipefail` and you're safe"** — Each flag has edge cases that can make your script exit unexpectedly.
2. **"set -e catches all errors"** — It doesn't. It ignores errors in `if` conditions, `||` chains, and several other contexts.
3. **"I don't need this — I check errors manually"** — You miss one check and `rm -rf "$dir/$subdir"` runs with an empty `$subdir`. `set -u` catches that.

## Each Flag Explained

### `set -e` (Exit on Error / errexit)

The shell exits immediately when a command returns non-zero.

```bash
set -e

echo "Step 1"
false              # Script exits here! No step 2.
echo "Step 2"      # Never reached
```

**What counts as an "error" for `set -e`:**
- Any command that returns non-zero
- EXCEPT in these contexts (it's ignored):
  - Part of an `if` condition: `if cmd; then ...`
  - Left side of `&&` or `||`: `cmd || handle_error`
  - Negated with `!`: `! cmd`
  - Inside a `while` or `until` condition

```bash
set -e

# These DON'T trigger exit:
if false; then echo yes; fi         # false is in 'if' condition
false || echo "recovered"           # false is left of ||
! false                             # Negated — success!
while false; do echo nope; done     # false is in 'while' condition

# These DO trigger exit:
false                               # Bare command → EXIT
$(exit 1)                           # Command substitution → EXIT
```

### `set -u` (Undefined Variables / nounset)

Error on any use of an undefined variable.

```bash
set -u

echo "$undefined_var"    # Error: unbound variable → EXIT
echo "$HOME"             # Fine — HOME is defined
```

**Why this matters:** Without `-u`, a typo in a variable name silently expands to empty:

```bash
# Without set -u, this is CATASTROPHIC:
rm -rf "${diretcory}/"   # Typo! $diretcory is empty → rm -rf /
                          # (Thankfully rm has safeguards now, but still)

# With set -u:
rm -rf "${diretcory}/"   # Error: diretcory: unbound variable → script stops
```

**Using default values with `set -u`:**
```bash
set -u

# $1 might not be set. Use a default:
name="${1:-anonymous}"     # If $1 unset, use "anonymous"

# Check if set without triggering error:
echo "${maybe_defined:-}"  # Empty string if undefined — no error

# ${var+x} trick to check existence:
if [[ -n "${maybe_defined+x}" ]]; then
    echo "Defined: $maybe_defined"
else
    echo "Not defined"
fi
```

### `set -o pipefail`

Without pipefail, a pipeline's exit code is the **last** command's exit code. With pipefail, it's the **first** failure.

```bash
# Without pipefail:
false | true
echo $?     # 0 (true's exit code — false's failure is hidden!)

# With pipefail:
set -o pipefail
false | true
echo $?     # 1 (false's failure is reported)

# Real example:
set -o pipefail
curl -sf "https://bad-url" | jq '.'
echo $?     # curl's failure (non-zero) is caught
```

## The Standard Header

```bash
#!/usr/bin/env bash
set -euo pipefail
```

This is the commonly recommended "strict mode" for scripts. Some people also add `IFS=$'\n\t'` to change the default word-splitting characters (removes space, keeping only newline and tab), though this is more controversial.

## When `set -e` Is Dangerous

`set -e` has well-documented gotchas that cause scripts to exit when you don't expect.

### Gotcha 1: Command in an if kills set -e for the whole chain

```bash
set -e

# If you test a function, set -e is DISABLED inside that function:
might_fail() {
    false        # This does NOT cause exit!
    echo "Still running"  # This executes
}

if might_fail; then
    echo "Succeeded"
fi
# "Still running" and "Succeeded" both print
# set -e was disabled inside might_fail because it was in an 'if' condition
```

### Gotcha 2: Arithmetic with result 0

```bash
set -e
x=0
((x++))     # Exit code 1! Because x was 0, and ((0)) is "false"
echo "never reached"

# Fix:
((x++)) || true
# Or:
((++x))     # Pre-increment — result is 1, which is "true"
```

### Gotcha 3: grep finding no matches

```bash
set -e
count=$(grep -c "pattern" file.txt)  # If no matches → exit code 1 → script dies!

# Fix:
count=$(grep -c "pattern" file.txt || true)
# Or:
if matches=$(grep "pattern" file.txt); then
    count=$(echo "$matches" | wc -l)
else
    count=0
fi
```

### Gotcha 4: Subshells inherit set -e inconsistently

```bash
set -e
(
    false          # Subshell exits here — parent script ALSO exits
)
echo "reached?"    # Not reached

# But in command substitution, behavior varies by Bash version:
result=$(false)     # Does this kill the script? It depends on context.
```

## A Pragmatic Approach

Instead of blindly using `set -e`, consider this approach:

```bash
#!/usr/bin/env bash
set -uo pipefail   # Use -u and pipefail, but NOT -e

# Handle errors explicitly:
cd /var/www || { echo "Failed to cd" >&2; exit 1; }

# Critical commands:
if ! rsync -av /data /backup; then
    echo "Backup failed" >&2
    exit 1
fi

# Commands where failure is acceptable:
grep -q "pattern" file.txt || echo "Pattern not found (that's OK)"

# Cleanup on exit:
trap 'echo "Script exiting with code $?"' EXIT
```

Or if you use `set -e`, know the escapes:

```bash
#!/usr/bin/env bash
set -euo pipefail

# Allow a command to fail:
might_fail || true        # Always succeeds, even if might_fail returns non-zero

# Allow a pipeline to fail:
cmd | grep "x" || true    # Pipeline failure won't kill script

# Allow and capture:
result=$(might_fail) || true   # $result has output, script continues
```

## Exercise

1. Write a script with `set -e` and demonstrate each way it can be bypassed (if, ||, !).
2. Write a script that demonstrates the `((x++))` gotcha with `set -e`.
3. Compare the behavior of `false | true; echo $?` with and without `set -o pipefail`.
4. Write a script that uses `set -u` and handles optional arguments with `${1:-default}`.

---

Next: [Writing Scripts That Fail Safely](02-failing-safely.md)
