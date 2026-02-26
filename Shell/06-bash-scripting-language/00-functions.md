# Functions

## What Problem This Solves

You have the same 10 lines scattered across 5 scripts. Or your script is 300 lines with no organization. Functions let you name blocks of logic, reuse them, and make scripts readable.

## The Mental Model

Bash functions are **not like functions in other languages.** They're closer to "named command groups." They:

- Run in the **current shell** (not a subshell)
- Receive arguments via `$1`, `$2`, `$@` — NOT named parameters
- Return an **exit code** (0-255), not a value
- "Return" data by printing to stdout
- Have **no type system** — everything is strings
- Variables are **global by default**

```
┌─── What You Think ───────────┐   ┌─── What Actually Happens ──────┐
│ result = add(3, 5)           │   │ result=$(add 3 5)              │
│ function returns value       │   │ function prints to stdout      │
│ named parameters             │   │ positional: $1=3, $2=5         │
│ local scope by default       │   │ global scope by default        │
└──────────────────────────────┘   └────────────────────────────────┘
```

## Function Syntax

```bash
# Preferred syntax (POSIX compatible):
my_function() {
    echo "Hello from function"
}

# Alternative syntax (Bash-only, adds nothing):
function my_function {
    echo "Hello from function"
}

# Don't use this (confusing mix):
function my_function() {
    echo "Avoid this style"
}
```

**Always use the `name() { }` form.** It's portable and unambiguous.

## Arguments Are Positional

```bash
greet() {
    local name="$1"
    local greeting="${2:-Hello}"    # Default value
    echo "$greeting, $name!"
}

greet "Alice"              # Hello, Alice!
greet "Bob" "Hey"          # Hey, Bob!
greet                      # Hello, !  (oops — no validation)

# Better — validate inputs:
greet() {
    if [[ $# -lt 1 ]]; then
        echo "Usage: greet NAME [GREETING]" >&2
        return 1
    fi
    local name="$1"
    local greeting="${2:-Hello}"
    echo "$greeting, $name!"
}
```

### All the Positional Variables

```bash
show_args() {
    echo "\$0 = $0"            # Script name (not function name!)
    echo "\$# = $#"            # Number of arguments
    echo "\$1 = $1"            # First argument
    echo "\$2 = $2"            # Second argument
    echo "\$@ = $@"            # All arguments (as separate words)
    echo "\$* = $*"            # All arguments (as single string)
    echo "FUNCNAME = ${FUNCNAME[0]}"  # Current function name (Bash-only)
}
```

## Returning Values

### Exit Codes (the real "return")

```bash
is_running() {
    local service="$1"
    systemctl is-active --quiet "$service"
    # return implicitly uses the exit code of the last command
}

if is_running "sshd"; then
    echo "SSH is running"
fi

# Explicit return:
validate_port() {
    local port="$1"
    if [[ "$port" -lt 1 || "$port" -gt 65535 ]] 2>/dev/null; then
        return 1
    fi
    return 0
}
```

### Returning Data via stdout

```bash
# The Bash way to "return" a value — print it:
get_ip() {
    hostname -I | awk '{print $1}'
}

# Capture with command substitution:
my_ip=$(get_ip)
echo "My IP is $my_ip"

# DANGER: Don't mix output and "return values":
bad_function() {
    echo "Starting process..."    # Debug message goes to stdout!
    echo "192.168.1.1"            # "Return value" also goes to stdout!
}

ip=$(bad_function)
echo "$ip"    # "Starting process...\n192.168.1.1" — broken!

# FIX: Send debug output to stderr:
good_function() {
    echo "Starting process..." >&2    # Debug → stderr
    echo "192.168.1.1"                # Return value → stdout
}

ip=$(good_function)
echo "$ip"    # "192.168.1.1"
```

### Returning Data via Global Variable

```bash
# Sometimes used for complex return values:
parse_config() {
    local file="$1"
    # Set global variables as "return values":
    CONFIG_HOST=$(grep '^host=' "$file" | cut -d= -f2)
    CONFIG_PORT=$(grep '^port=' "$file" | cut -d= -f2)
}

parse_config /etc/myapp/config
echo "Connecting to $CONFIG_HOST:$CONFIG_PORT"

# This works because functions run in the current shell.
# Downside: caller must know the variable names. Not very clean.
```

## Local Variables

```bash
# WITHOUT local — variable leaks:
process() {
    temp="sensitive data"    # Global!
}
process
echo "$temp"    # "sensitive data" — leaked!

# WITH local — variable is scoped:
process() {
    local temp="sensitive data"
}
process
echo "$temp"    # Empty — properly scoped

# RULE: Always use local for variables that shouldn't escape the function.
```

### local Gotchas

```bash
# Gotcha 1: local clobbers $?
my_func() {
    local result
    result=$(failing_command)
    local status=$?         # WRONG — $? is now the exit code of `local`, which is 0!
}

# Fix: separate declaration and assignment
my_func() {
    local result
    result=$(failing_command)
    local status=$?         # Actually, this case IS fine — it's only wrong when combined:
}

my_func() {
    local result=$(failing_command)
    echo $?                 # 0!!! Because `local` succeeded, hiding the failure
}

# Fix:
my_func() {
    local result
    result=$(failing_command)    # Now $? reflects failing_command
    echo $?                      # Correct exit code
}

# Gotcha 2: local variables are visible in called functions (dynamic scoping):
outer() {
    local x="outer"
    inner
}
inner() {
    echo "$x"    # "outer" — inner can see outer's local!
}
# This is dynamic scoping, not lexical scoping. It's confusing.
```

## Function Patterns

### Die Function

```bash
die() {
    echo "ERROR: $*" >&2
    exit 1
}

# Usage:
[[ -f "$config_file" ]] || die "Config file not found: $config_file"
```

### Log Function

```bash
log() {
    local level="$1"
    shift
    printf '[%s] [%-5s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$level" "$*" >&2
}

log INFO "Starting backup"
log ERROR "Connection failed"
# [2024-01-15 14:30:22] [INFO ] Starting backup
# [2024-01-15 14:30:22] [ERROR] Connection failed
```

### Retry Function

```bash
retry() {
    local max_attempts="$1"
    local delay="$2"
    shift 2
    local cmd=("$@")

    local attempt=1
    while (( attempt <= max_attempts )); do
        if "${cmd[@]}"; then
            return 0
        fi
        log WARN "Attempt $attempt/$max_attempts failed. Retrying in ${delay}s..."
        sleep "$delay"
        ((attempt++))
    done

    log ERROR "All $max_attempts attempts failed: ${cmd[*]}"
    return 1
}

# Usage:
retry 3 5 curl -sf "https://api.example.com/health"
```

### Cleanup with trap

```bash
cleanup() {
    local exit_code=$?
    rm -f "$tmp_file"
    log INFO "Cleanup complete (exit code: $exit_code)"
    exit "$exit_code"    # Preserve the original exit code
}
trap cleanup EXIT

tmp_file=$(mktemp)
```

## Common Footguns

### 1. Forgetting That `return` Exits the Function, Not the Script

```bash
check() {
    [[ -f "$1" ]] || return 1    # Returns from check(), not from script
}

check "/nonexistent"
echo "This still runs"    # Yes, it does
```

### 2. Using `exit` Inside a Function (Sometimes Wrong)

```bash
# exit inside a function exits the ENTIRE script:
validate() {
    [[ -n "$1" ]] || exit 1    # Kills the whole script!
}

# Usually you want return:
validate() {
    [[ -n "$1" ]] || return 1    # Returns to caller
}
```

### 3. Function Name Collides with a Command

```bash
# DANGEROUS — overrides the real test command:
test() {
    echo "Running tests..."
}

# Now [[ and test behave differently than expected!
# NEVER name functions after builtins or common commands.
```

## Exercise

1. Write a `confirm()` function that asks "Are you sure? [y/N]" and returns 0 for yes, 1 for no. Use it before a destructive operation.

2. Write a `require_command()` function that checks if a command exists (with `command -v`) and dies with a helpful message if it doesn't.

3. Write a function that "returns" a value by printing to stdout. Demonstrate the stderr-for-debug-messages pattern.

4. Create a script with a `main()` function that calls helper functions. Use `local` everywhere. Trap EXIT for cleanup.

---

Next: [Conditionals and Tests](01-conditionals.md)
