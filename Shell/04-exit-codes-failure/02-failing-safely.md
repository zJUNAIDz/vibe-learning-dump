# Writing Scripts That Fail Safely

## What Problem This Solves

You write a script that works 99% of the time but silently corrupts data the 1% when something unexpected happens — a full disk, a missing directory, a permissions change, a network timeout. This file is about writing scripts that **fail loudly, cleanly, and without collateral damage**.

## The Safety Checklist

Every script that does anything important should address:

1. **Fail on the first error** — Don't continue blindly
2. **Validate preconditions** — Check before you act
3. **Clean up on exit** — Remove temp files, release locks
4. **Log what you're doing** — So you can debug later
5. **Be idempotent** — Safe to run twice

## Pattern 1: The Safe Script Template

```bash
#!/usr/bin/env bash
set -euo pipefail

# Constants
readonly SCRIPT_NAME="$(basename "$0")"
readonly SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Logging
log() { echo "[$(date -Iseconds)] $SCRIPT_NAME: $*" >&2; }
die() { log "FATAL: $*"; exit 1; }

# Cleanup
cleanup() {
    local exit_code=$?
    # Remove temp files, release locks, etc.
    log "Exiting with code $exit_code"
    # Perform cleanup here
    exit $exit_code
}
trap cleanup EXIT

# Precondition checks
[[ $# -ge 1 ]] || die "Usage: $SCRIPT_NAME <target_dir>"
[[ -d "$1" ]] || die "Directory does not exist: $1"
command -v rsync &>/dev/null || die "rsync is not installed"

# Main logic
main() {
    local target_dir="$1"
    log "Starting backup to $target_dir"
    
    rsync -av /data/ "$target_dir/" || die "rsync failed"
    
    log "Backup complete"
}

main "$@"
```

## Pattern 2: Precondition Validation

Check everything before doing anything destructive:

```bash
validate_environment() {
    local errors=0
    
    # Check required commands:
    for cmd in rsync jq curl; do
        if ! command -v "$cmd" &>/dev/null; then
            log "ERROR: Required command not found: $cmd"
            ((errors++))
        fi
    done
    
    # Check required files:
    for file in /etc/myapp/config.yaml "$HOME/.ssh/id_rsa"; do
        if [[ ! -f "$file" ]]; then
            log "ERROR: Required file missing: $file"
            ((errors++))
        fi
    done
    
    # Check permissions:
    if [[ ! -w "/var/log/myapp" ]]; then
        log "ERROR: Cannot write to /var/log/myapp"
        ((errors++))
    fi
    
    # Check disk space (at least 1GB free):
    local free_kb
    free_kb=$(df /var --output=avail | tail -1 | tr -d ' ')
    if (( free_kb < 1048576 )); then
        log "ERROR: Less than 1GB free on /var"
        ((errors++))
    fi
    
    if (( errors > 0 )); then
        die "$errors precondition(s) failed"
    fi
}
```

## Pattern 3: Safe Cleanup with trap

`trap` sets a command to run when the script receives a signal or exits:

```bash
#!/usr/bin/env bash
set -euo pipefail

# Create temp file
tmpfile=$(mktemp /tmp/myapp.XXXXXX)

# ALWAYS clean up temp files — even on error, Ctrl+C, or kill
trap 'rm -f "$tmpfile"' EXIT

# Now use $tmpfile safely
curl -sf "https://api.example.com/data" > "$tmpfile"
jq '.results' < "$tmpfile"

# When the script exits (for ANY reason), the temp file is deleted
```

### trap for Multiple Signals

```bash
cleanup() {
    local exit_code=$?
    log "Cleaning up (exit code: $exit_code)..."
    rm -f "$tmpfile" "$lockfile"
    
    # If killed by signal, re-raise it:
    if (( exit_code > 128 )); then
        trap - EXIT  # Prevent double cleanup
        kill -$((exit_code - 128)) $$
    fi
    
    exit $exit_code
}

trap cleanup EXIT      # On normal exit
trap cleanup INT       # On Ctrl+C
trap cleanup TERM      # On kill
```

## Pattern 4: Idempotent Operations

An idempotent script produces the same result whether you run it once or ten times:

```bash
# NOT idempotent — fails on second run:
mkdir /var/myapp            # Error: already exists

# Idempotent — safe to run repeatedly:
mkdir -p /var/myapp          # Creates if needed, no error if exists

# NOT idempotent — appends duplicate entries:
echo "export PATH=/opt/bin:$PATH" >> ~/.bashrc

# Idempotent — only add if not present:
grep -qxF 'export PATH=/opt/bin:$PATH' ~/.bashrc || \
    echo 'export PATH=/opt/bin:$PATH' >> ~/.bashrc

# NOT idempotent — creates duplicate users:
useradd myuser

# Idempotent — check first:
id myuser &>/dev/null || useradd myuser
```

### Idempotent File Operations

```bash
# Create a config file only if it doesn't exist:
if [[ ! -f /etc/myapp/config.yaml ]]; then
    cat > /etc/myapp/config.yaml << 'YAML'
port: 8080
log_level: info
YAML
    log "Created config file"
else
    log "Config file already exists, skipping"
fi

# Ensure a specific line is in a file:
ensure_line() {
    local line="$1" file="$2"
    grep -qxF "$line" "$file" 2>/dev/null || echo "$line" >> "$file"
}
ensure_line "net.ipv4.ip_forward = 1" /etc/sysctl.d/99-custom.conf
```

## Pattern 5: Safe Variable Usage

```bash
# Always use ${var:?message} for required variables:
db_host="${DB_HOST:?DB_HOST environment variable is required}"
db_name="${DB_NAME:?DB_NAME environment variable is required}"

# Default values for optional variables:
db_port="${DB_PORT:-5432}"
log_level="${LOG_LEVEL:-info}"

# Never use unquoted variables in dangerous commands:
rm -rf "${target_dir:?target_dir is empty}"/ 
# The :? prevents rm -rf / when target_dir is unset
```

## Pattern 6: Error Reporting

```bash
#!/usr/bin/env bash
set -euo pipefail

# Rich error context:
die() {
    echo "ERROR in ${BASH_SOURCE[1]:-$0}:${BASH_LINENO[0]}: $*" >&2
    exit 1
}

# Or with a full stack trace:
error_handler() {
    local exit_code=$?
    local line_number=$1
    echo "ERROR: Script failed at line $line_number with exit code $exit_code" >&2
    
    # Print stack trace:
    local i=0
    while caller $i; do
        ((i++))
    done >&2
    
    exit $exit_code
}
trap 'error_handler $LINENO' ERR
```

## Pattern 7: Confirmation and Dry-Run

```bash
#!/usr/bin/env bash
set -euo pipefail

DRY_RUN="${DRY_RUN:-false}"

run() {
    if [[ "$DRY_RUN" == "true" ]]; then
        echo "[DRY RUN] $*" >&2
    else
        echo "[EXECUTING] $*" >&2
        "$@"
    fi
}

# Usage:
run rm -rf /tmp/old-backups
run systemctl restart myapp
run rsync -av /data /backup

# Run for real:
# ./script.sh

# Just see what would happen:
# DRY_RUN=true ./script.sh
```

## Common Footguns

**Footgun 1: Cleanup code that can fail**
```bash
trap 'rm "$tmpfile"' EXIT
# If $tmpfile is unset (set -u) or the file doesn't exist, cleanup fails!

# Fix:
trap 'rm -f "${tmpfile:-}"' EXIT
```

**Footgun 2: Race conditions in file operations**
```bash
# RACE CONDITION:
if [[ ! -f "$lockfile" ]]; then
    touch "$lockfile"     # Another instance could create it between the check and this!
fi

# FIX — atomic lock with mkdir (mkdir is atomic on POSIX):
if mkdir "$lockdir" 2>/dev/null; then
    trap 'rmdir "$lockdir"' EXIT
    # Critical section
else
    die "Another instance is running (lock: $lockdir)"
fi
```

**Footgun 3: Forgetting to quote in rm/mv**
```bash
# These are all different levels of danger:
rm $file          # Word splitting + globbing! Catastrophic if $file="* important"
rm "$file"        # Safe from splitting, but empty $file = rm "" = harmless error
rm "${file:?}"    # Safest — errors if $file is empty
```

## Exercise

1. Write a backup script using the safe template. It should:
   - Take a source and destination directory as arguments
   - Validate both exist  
   - Use a temp file for an intermediate step
   - Clean up the temp file on any exit
   - Be idempotent

2. Add a dry-run mode to an existing script.

3. Write a script with a lock file (using `mkdir`) that prevents two copies from running simultaneously. Test by running it in two terminals.

4. Write a deployment script that checks all preconditions (commands exist, disk space available, service is running) before making any changes.

---

Next: [Level 5: Shell Variables vs Environment Variables](../05-variables-scope-env/00-shell-vs-env-vars.md)
