# Subshells, Export, and Scope

## What Problem This Solves

Your variable changes inside a loop but is empty after the loop! Or you source a script and it changes your PATH but it shouldn't. Understanding when the shell creates new processes vs runs in the current process eliminates these mysteries.

## The Mental Model

**Subshell** = a child copy of the current shell. It inherits everything (all variables, functions, options), but any changes it makes are **thrown away** when it exits.

**Current shell execution** = code runs in YOUR shell process. Changes (variables, cd, exported values) persist.

```
┌── Subshell Creators ──────────────────────────────────────┐
│ ( commands )          ← Explicit subshell                  │
│ cmd | cmd             ← Each side of pipe (Bash default)   │
│ $(cmd)                ← Command substitution               │
│ bash script.sh        ← New shell process                  │
│ cmd &                 ← Background process                 │
└───────────────────────────────────────────────────────────┘

┌── Current Shell Execution ────────────────────────────────┐
│ source script.sh      ← Runs in YOUR shell                │
│ . script.sh           ← Same as source                    │
│ { commands; }         ← Group commands (NOT a subshell!)   │
│ function_call         ← Functions run in current shell     │
│ eval "code"           ← Runs in current shell              │
└───────────────────────────────────────────────────────────┘
```

## Subshell Deep Dive

### Explicit Subshell: `( ... )`

```bash
x=1
echo "Before: $x"     # 1

(
    x=99
    echo "Inside: $x"  # 99
    cd /tmp
    echo "Dir: $PWD"   # /tmp
)

echo "After: $x"       # 1  ← Change was lost!
echo "Dir: $PWD"        # (original directory — cd was lost too!)
```

**Use cases for explicit subshells:**
```bash
# Temporarily change directory without affecting the parent:
(cd /var/log && tar czf /tmp/logs.tar.gz *.log)
# Back in original directory — no need to cd back

# Temporarily change IFS:
(IFS=:; echo $PATH)
# IFS is unchanged after

# Isolate dangerous operations:
(set +e; dangerous_command; echo "survived")
# set -e is restored in parent
```

### Pipe Subshells (The Classic Trap)

```bash
count=0
echo -e "a\nb\nc" | while read -r line; do
    ((count++))
done
echo "Count: $count"    # 0!!! Not 3!
```

**Why?** In Bash, each side of a pipe runs in a subshell. The `while` loop modifies `count` in a child process. The parent never sees the change.

**Fixes:**

```bash
# Fix 1: Redirect instead of pipe
count=0
while read -r line; do
    ((count++))
done < <(echo -e "a\nb\nc")    # Process substitution — while runs in current shell
echo "Count: $count"    # 3

# Fix 2: Use lastpipe (Bash 4.2+)
set +m          # Disable job control (required for lastpipe)
shopt -s lastpipe
count=0
echo -e "a\nb\nc" | while read -r line; do
    ((count++))
done
echo "Count: $count"    # 3 (last command in pipe runs in current shell)

# Fix 3: Restructure to avoid the problem
count=$(echo -e "a\nb\nc" | wc -l)

# Zsh note: In Zsh, the LAST pipe segment runs in the current shell by default!
```

### Command Substitution: `$(...)` subshell

```bash
# $() runs in a subshell:
x=1
result=$(x=99; echo $x)
echo "$result"    # 99 (the subshell's echo output)
echo "$x"         # 1  (parent unchanged)
```

## source vs bash — The Critical Difference

```bash
# bash script.sh  →  Runs in a NEW process (subshell)
# source script.sh  →  Runs in the CURRENT shell
# . script.sh  →  Same as source (POSIX syntax)
```

### When to use each:

```bash
# source: When you WANT changes to affect the current shell
source ~/.bashrc          # Reload config
. /opt/myapp/env.sh       # Load environment variables
source ./venv/bin/activate # Python virtualenv

# bash: When you want isolation
bash deploy.sh            # Script can't mess up your shell
bash -x debug.sh          # Run with debugging
```

### Danger of source

```bash
# If env.sh contains:
# cd /tmp
# export PATH="/dangerous:$PATH"
# alias ls='rm'

source env.sh
# Now YOUR shell is in /tmp, has a modified PATH, and ls will delete files!

# This is why you should NEVER source untrusted files.
# bash env.sh would be safe — changes isolated to the child process.
```

## { } vs ( ): Group vs Subshell

```bash
# Curly braces — runs in CURRENT shell:
{
    x=99
    echo "Inside braces: $x"
}
echo "After braces: $x"    # 99 — change persisted!

# Parentheses — runs in SUBSHELL:
(
    y=99
    echo "Inside parens: $y"
)
echo "After parens: $y"    # Empty — change lost!

# Syntax note: { } require spaces and semicolons:
{ echo hello; echo world; }    # Correct
{echo hello}                    # WRONG — syntax error
```

### Practical Use of Braces

```bash
# Redirect a block of commands:
{
    echo "=== System Info ==="
    date
    hostname
    uname -r
    df -h
} > /tmp/sysinfo.txt
# All output goes to the file. Variables set inside persist.

# Compare with subshell:
(
    echo "=== System Info ==="
    date
) > /tmp/sysinfo.txt
# Output also goes to file. But variables set inside DON'T persist.
```

## Variable Scope in Functions

Bash functions run in the current shell (not a subshell), but their variables are **global by default**:

```bash
my_func() {
    result="from function"    # Global variable!
    local temp="temporary"    # Local variable — only visible in this function
}

my_func
echo "$result"    # "from function" — visible because it's global
echo "$temp"      # Empty — local was scoped to the function
```

### The `local` Keyword

```bash
# Without local — variable "leaks" out:
outer() {
    x="outer"
    inner
    echo "After inner: $x"
}
inner() {
    x="inner"    # Modifies the same $x!
}
outer
# After inner: inner  ← outer's x was clobbered!

# With local — variable is scoped:
outer() {
    local x="outer"
    inner
    echo "After inner: $x"
}
inner() {
    local x="inner"
}
outer
# After inner: outer  ← outer's x is untouched
```

## declare and readonly

```bash
# readonly — can't be changed or unset:
readonly PI=3.14159
PI=3.0    # Error: PI: readonly variable

# declare -r — same as readonly:
declare -r CONST="immutable"

# declare -i — integer variable:
declare -i num=5
num="hello"     # Sets to 0! Bash interprets non-numeric as 0
num=10+5        # Sets to 15! Arithmetic evaluation

# declare -a — indexed array:
declare -a fruits=("apple" "banana" "cherry")

# declare -A — associative array (Bash 4+):
declare -A colors
colors[sky]="blue"
colors[grass]="green"
```

## Exercise

1. Prove that `( ... )` creates a subshell: set a variable inside and show it's not visible outside.

2. Demonstrate the pipe subshell problem: count lines in a pipe's while loop, show the count is wrong, then fix it with process substitution.

3. Create a file `setup-env.sh` that sets 3 environment variables. Show that `bash setup-env.sh` doesn't affect your shell but `source setup-env.sh` does.

4. Write two functions where one clobbers the other's variable (without `local`), then fix it with `local`.

---

Next: [Fedora Service Environments](02-fedora-service-env.md)
