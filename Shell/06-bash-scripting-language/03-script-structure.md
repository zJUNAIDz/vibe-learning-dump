# Script Structure

## What Problem This Solves

Your script is 500 lines of sequential commands with no organization. When it fails, you don't know where. When you come back in 3 months, you can't figure out what it does. Good structure makes scripts maintainable, debuggable, and safe.

## The Template

Every non-trivial script should follow this structure:

```bash
#!/usr/bin/env bash
#
# script-name.sh — One-line description of what this script does.
#
# Usage:
#   ./script-name.sh [OPTIONS] <required-arg>
#
# Options:
#   -v, --verbose    Enable verbose output
#   -n, --dry-run    Show what would be done without doing it
#   -h, --help       Show this help message
#
# Examples:
#   ./script-name.sh /var/log
#   ./script-name.sh --verbose --dry-run /tmp
#

set -euo pipefail

# ── Constants ──────────────────────────────────────────────
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"

# ── Configuration (overridable via environment) ────────────
VERBOSE="${VERBOSE:-false}"
DRY_RUN="${DRY_RUN:-false}"

# ── Logging ────────────────────────────────────────────────
log()   { printf '[%s] %s\n' "$(date '+%H:%M:%S')" "$*" >&2; }
info()  { log "INFO  $*"; }
warn()  { log "WARN  $*"; }
error() { log "ERROR $*"; }
debug() { [[ "$VERBOSE" = true ]] && log "DEBUG $*"; return 0; }
die()   { error "$@"; exit 1; }

# ── Cleanup ────────────────────────────────────────────────
TMP_DIR=""
cleanup() {
    local exit_code=$?
    if [[ -n "$TMP_DIR" && -d "$TMP_DIR" ]]; then
        rm -rf "$TMP_DIR"
        debug "Cleaned up temp dir: $TMP_DIR"
    fi
    exit "$exit_code"
}
trap cleanup EXIT

# ── Argument Parsing ───────────────────────────────────────
usage() {
    # Print the header comment block as usage text:
    sed -n '/^# Usage:/,/^[^#]/{ /^#/s/^# \{0,1\}//p }' "$0"
    # Or just hardcode it:
    cat <<EOF
Usage: $SCRIPT_NAME [OPTIONS] <target-dir>

Options:
  -v, --verbose    Enable verbose output
  -n, --dry-run    Show what would be done
  -h, --help       Show this help message
EOF
}

parse_args() {
    local positional=()

    while [[ $# -gt 0 ]]; do
        case "$1" in
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -n|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            --)
                shift
                positional+=("$@")
                break
                ;;
            -*)
                die "Unknown option: $1 (use --help for usage)"
                ;;
            *)
                positional+=("$1")
                shift
                ;;
        esac
    done

    # Assign positional arguments:
    TARGET_DIR="${positional[0]:-}"
    [[ -n "$TARGET_DIR" ]] || die "Missing required argument: target-dir"
    [[ -d "$TARGET_DIR" ]] || die "Not a directory: $TARGET_DIR"
}

# ── Helper Functions ───────────────────────────────────────
run_or_dry() {
    if [[ "$DRY_RUN" = true ]]; then
        info "[DRY RUN] Would run: $*"
    else
        debug "Running: $*"
        "$@"
    fi
}

require_command() {
    local cmd="$1"
    command -v "$cmd" &>/dev/null || die "Required command not found: $cmd (install with: sudo dnf install $cmd)"
}

# ── Core Logic ─────────────────────────────────────────────
do_the_work() {
    local target="$1"
    info "Processing: $target"

    TMP_DIR=$(mktemp -d)
    debug "Created temp dir: $TMP_DIR"

    # ... actual work here ...
    info "Done processing $target"
}

# ── Main ───────────────────────────────────────────────────
main() {
    parse_args "$@"

    require_command "curl"
    require_command "jq"

    info "Starting $SCRIPT_NAME"
    debug "Target: $TARGET_DIR"
    debug "Verbose: $VERBOSE"
    debug "Dry run: $DRY_RUN"

    do_the_work "$TARGET_DIR"

    info "Completed successfully"
}

main "$@"
```

## Why This Structure Works

### `set -euo pipefail` at the Top

```bash
set -e          # Exit on error
set -u          # Error on undefined variables
set -o pipefail # Pipe fails if any command fails
```

This catches the majority of bugs. See Level 4 for details.

### `readonly` for Constants

```bash
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
```

- `${BASH_SOURCE[0]}` — the path to THIS script (works even when sourced)
- `dirname` — gets the directory part
- `cd ... && pwd` — resolves to an absolute path
- `readonly` — prevents accidental modification

### Logging to stderr

```bash
log() { printf '[%s] %s\n' "$(date '+%H:%M:%S')" "$*" >&2; }
```

All log messages go to stderr (`>&2`). This means stdout is clean for actual output, enabling piping:

```bash
./process.sh /data > results.txt    # Log goes to terminal, results to file
```

### trap EXIT for Cleanup

```bash
trap cleanup EXIT
```

The `cleanup` function runs when the script exits — whether normally, via `die`, via `set -e`, or via signal. It preserves the original exit code.

### main() Pattern

```bash
main() {
    # all logic here
}
main "$@"
```

Why a `main()` function?
1. **Prevents accidental global variables** — use `local` inside functions
2. **Makes the script sourceable** — if someone runs `source script.sh`, main doesn't auto-execute (you could add a guard)
3. **Clear entry point** — easy to find where execution starts

### Sourceable Guard (Optional)

```bash
# Only run main if script is executed directly, not sourced:
if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    main "$@"
fi
```

This lets other scripts `source` yours to use its functions without triggering `main`.

## Argument Parsing Patterns

### Simple: Positional Only

```bash
main() {
    local src="${1:?Usage: $0 <source> <dest>}"
    local dest="${2:?Usage: $0 <source> <dest>}"
    # ${var:?message} exits with an error if var is unset or empty
}
```

### Medium: Flags + Positional

The `while/case/shift` pattern from the template above. This is the most common and most flexible approach.

### Advanced: getopts (for short options)

```bash
while getopts ":vnh" opt; do
    case "$opt" in
        v) VERBOSE=true ;;
        n) DRY_RUN=true ;;
        h) usage; exit 0 ;;
        :) die "Option -$OPTARG requires an argument" ;;
        \?) die "Unknown option: -$OPTARG" ;;
    esac
done
shift $((OPTIND - 1))    # Remove parsed options, leaving positional args
```

**Limitations of getopts:** No long options (`--verbose`). Only short flags. For complex CLIs, the while/case pattern or a proper argument parsing library is better.

## File Organization

```
project/
├── scripts/
│   ├── deploy.sh          # Main script
│   ├── lib/
│   │   ├── logging.sh     # Shared logging functions
│   │   └── validation.sh  # Shared validation functions
│   └── config/
│       ├── defaults.env   # Default configuration
│       └── production.env # Production overrides
```

### Sourcing Libraries

```bash
#!/usr/bin/env bash
set -euo pipefail

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source shared libraries:
source "$SCRIPT_DIR/lib/logging.sh"
source "$SCRIPT_DIR/lib/validation.sh"

main() {
    # Now you can use functions from the libraries
    log INFO "Starting deployment"
    validate_environment
}

main "$@"
```

## Configuration Patterns

### Environment Variables with Defaults

```bash
# In the script:
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
LOG_LEVEL="${LOG_LEVEL:-info}"

# Caller can override:
DB_HOST=prod-db.internal LOG_LEVEL=debug ./deploy.sh
```

### Config File Loading

```bash
load_config() {
    local config_file="$1"
    [[ -f "$config_file" ]] || die "Config not found: $config_file"

    # Source it (runs as shell code — only trust your own files!):
    source "$config_file"

    # Or parse it safely (key=value only, no execution):
    while IFS='=' read -r key value; do
        [[ "$key" =~ ^[A-Z_]+$ ]] || continue    # Only accept uppercase var names
        declare -g "$key=$value"
    done < "$config_file"
}
```

## Naming Conventions

```bash
# Variables: UPPER_SNAKE for globals/constants, lower_snake for locals
readonly MAX_RETRIES=5
local current_count=0

# Functions: lower_snake_case, verb-first
validate_input()
process_file()
send_notification()

# Script files: lowercase, hyphens
deploy-app.sh
run-migrations.sh
check-health.sh

# Flag-like variables: use true/false strings
VERBOSE=false
DRY_RUN=false
if [[ "$VERBOSE" = true ]]; then ...
```

## Common Footguns

### 1. Script Doesn't Work When Called from a Different Directory

```bash
# WRONG — assumes CWD is the script's directory:
source lib/utils.sh
config_file="config/app.conf"

# RIGHT — use SCRIPT_DIR:
source "$SCRIPT_DIR/lib/utils.sh"
config_file="$SCRIPT_DIR/config/app.conf"
```

### 2. No Error Messages on Failure

```bash
# WRONG:
cd "$target" || exit 1        # Exits silently

# RIGHT:
cd "$target" || die "Cannot cd to $target"
```

### 3. Hardcoded Paths

```bash
# WRONG:
log_file="/home/alice/logs/app.log"

# RIGHT:
log_file="${LOG_DIR:-/var/log/myapp}/app.log"
```

## Exercise

1. Take any messy script you've written and restructure it using the template above. Add argument parsing, logging, and cleanup.

2. Create a script with `--help`, `--verbose`, `--dry-run` flags and one positional argument. Demonstrate all three modes.

3. Split a script into a main file and a library file (`lib/helpers.sh`). Source the library using `SCRIPT_DIR`.

4. Write a script that works correctly regardless of which directory it's called from — test by calling it from `/tmp`, from `~`, and from the script's own directory.

---

Next: [Level 7 — Text Processing](../07-text-processing/00-grep-sed-awk.md)
