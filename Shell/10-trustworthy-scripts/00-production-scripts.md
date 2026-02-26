# Production-Quality Scripts

## What Problem This Solves

Your script works on your machine but fails in CI, on another server, or after someone else edits it. Production scripts need to be **predictable**, **debuggable**, and **safe to run multiple times.** This file shows the patterns that separate "works for me" from "works everywhere."

## The Non-Negotiable Checklist

Every production script should have:

```bash
#!/usr/bin/env bash        # 1. Explicit interpreter
set -euo pipefail          # 2. Strict mode
                           # 3. Descriptive header comment
                           # 4. Input validation
                           # 5. Cleanup on exit (trap)
                           # 6. Logging to stderr
                           # 7. Meaningful exit codes
                           # 8. No hardcoded paths
                           # 9. ShellCheck clean
                           # 10. Idempotent operations
```

## 1. Input Validation

```bash
# Validate everything before doing anything:
main() {
    # Required arguments:
    local target="${1:?Usage: $0 <target-directory>}"

    # Type checks:
    [[ -d "$target" ]] || die "Not a directory: $target"
    [[ -w "$target" ]] || die "Not writable: $target"

    # Required commands:
    require_command curl
    require_command jq
    require_command rsync

    # Required environment:
    [[ -n "${API_KEY:-}" ]] || die "API_KEY environment variable is required"
    [[ -n "${DEPLOY_HOST:-}" ]] || die "DEPLOY_HOST environment variable is required"

    # Sanity checks:
    local disk_free
    disk_free=$(df --output=avail -B1 "$target" | tail -1)
    (( disk_free > 1073741824 )) || die "Less than 1GB free on $target"
}

require_command() {
    command -v "$1" &>/dev/null || die "Required command not found: $1"
}
```

## 2. Idempotency

A script is **idempotent** if running it twice produces the same result as running it once. This is critical for deployment scripts and automation.

```bash
# NON-IDEMPOTENT (breaks on second run):
mkdir /opt/myapp                 # Fails if exists
echo "config" >> /etc/myapp.conf # Appends duplicate
useradd myapp                    # Fails if user exists

# IDEMPOTENT:
mkdir -p /opt/myapp              # No-op if exists
echo "config" > /etc/myapp.conf  # Overwrites (or use a check)
id myapp &>/dev/null || useradd myapp  # Only creates if missing

# Pattern — "ensure" functions:
ensure_directory() {
    local dir="$1"
    local owner="${2:-}"
    if [[ ! -d "$dir" ]]; then
        mkdir -p "$dir"
        log INFO "Created directory: $dir"
    fi
    if [[ -n "$owner" ]]; then
        chown "$owner" "$dir"
    fi
}

ensure_package() {
    local pkg="$1"
    if ! rpm -q "$pkg" &>/dev/null; then
        sudo dnf install -y "$pkg"
        log INFO "Installed package: $pkg"
    else
        log DEBUG "Package already installed: $pkg"
    fi
}

ensure_service() {
    local svc="$1"
    if ! systemctl is-active --quiet "$svc"; then
        sudo systemctl start "$svc"
        log INFO "Started service: $svc"
    fi
    if ! systemctl is-enabled --quiet "$svc"; then
        sudo systemctl enable "$svc"
        log INFO "Enabled service: $svc"
    fi
}
```

## 3. Atomic Operations

Don't leave partial results if something fails:

```bash
# WRONG — partial file on failure:
curl -f "https://example.com/data.json" > /opt/myapp/config.json
# If curl fails halfway, config.json is corrupted

# RIGHT — write to temp, then move:
tmp=$(mktemp)
curl -f "https://example.com/data.json" -o "$tmp"
mv "$tmp" /opt/myapp/config.json
# mv is atomic on the same filesystem

# Pattern for config updates:
update_config() {
    local target="$1"
    local tmp
    tmp=$(mktemp "${target}.XXXXXX")

    # Write new config:
    generate_config > "$tmp"

    # Validate before replacing:
    validate_config "$tmp" || { rm "$tmp"; die "Invalid config generated"; }

    # Atomic replace:
    mv "$tmp" "$target"
    log INFO "Updated config: $target"
}

# Pattern for database migrations:
deploy() {
    local backup
    backup="/backups/pre-deploy-$(date +%Y%m%d-%H%M%S).sql"

    # Backup first:
    pg_dump mydb > "$backup" || die "Backup failed"
    log INFO "Backup saved: $backup"

    # Deploy:
    if ! run_migrations; then
        log ERROR "Migration failed, restoring backup"
        psql mydb < "$backup"
        die "Deploy failed and was rolled back"
    fi

    log INFO "Deploy successful"
}
```

## 4. Lock Files (Prevent Concurrent Execution)

```bash
# Prevent two instances from running simultaneously:
LOCK_FILE="/var/lock/myapp-deploy.lock"

acquire_lock() {
    if ! mkdir "$LOCK_FILE" 2>/dev/null; then
        local pid
        pid=$(cat "$LOCK_FILE/pid" 2>/dev/null || echo "unknown")
        die "Another instance is running (PID: $pid, lock: $LOCK_FILE)"
    fi
    echo $$ > "$LOCK_FILE/pid"
    log DEBUG "Acquired lock: $LOCK_FILE"
}

release_lock() {
    rm -rf "$LOCK_FILE"
    log DEBUG "Released lock: $LOCK_FILE"
}

# Use with trap:
cleanup() {
    release_lock
}
trap cleanup EXIT

main() {
    acquire_lock
    # ... do work ...
}
```

Why `mkdir` instead of a regular file? `mkdir` is **atomic** — it either succeeds or fails, no race condition.

### flock Alternative (Better for systemd)

```bash
# Using flock (file-based locking):
exec 200>/var/lock/myapp.lock
flock -n 200 || die "Another instance is already running"

# Or wrap the entire script:
# In the shebang area:
[ "${FLOCKER:-}" != "$0" ] && exec env FLOCKER="$0" flock -en "$0" "$0" "$@" || :
```

## 5. Signal Handling

```bash
# Handle interrupts gracefully:
interrupted=false

handle_signal() {
    local signal="$1"
    log WARN "Received signal: $signal"
    interrupted=true
    # Don't exit immediately — let the current operation finish
}

trap 'handle_signal INT' INT
trap 'handle_signal TERM' TERM

# In your main loop:
for item in "${items[@]}"; do
    if [[ "$interrupted" = true ]]; then
        log WARN "Interrupted — stopping after current item"
        break
    fi
    process_item "$item"
done

# Cleanup always runs:
trap cleanup EXIT
```

## 6. Timeouts

```bash
# Command timeout:
timeout 30 curl -sf "https://api.example.com/health" || die "Health check timed out"

# Custom timeout for a block:
with_timeout() {
    local seconds="$1"
    shift
    timeout "$seconds" bash -c "$*"
}

with_timeout 60 "rsync -avz /src/ remote:/dest/"
```

## 7. Dry-Run Mode

```bash
DRY_RUN="${DRY_RUN:-false}"

run() {
    if [[ "$DRY_RUN" = true ]]; then
        log INFO "[DRY RUN] $*"
    else
        log DEBUG "Running: $*"
        "$@"
    fi
}

# Usage:
run sudo systemctl restart nginx
run rsync -avz "$src/" "$dest/"
run rm -rf "$tmp_dir/"

# Invoke:
DRY_RUN=true ./deploy.sh      # See what would happen
./deploy.sh                    # Actually do it
```

## 8. Configuration from Environment (12-Factor)

```bash
# All config comes from environment with sensible defaults:
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-myapp}"
LOG_LEVEL="${LOG_LEVEL:-info}"
DEPLOY_ENV="${DEPLOY_ENV:-staging}"

# Validate required vars:
required_vars=(API_KEY DEPLOY_HOST SLACK_WEBHOOK)
for var in "${required_vars[@]}"; do
    [[ -n "${!var:-}" ]] || die "Required environment variable not set: $var"
done

# Never hardcode secrets:
# WRONG: API_KEY="sk-abc123"
# RIGHT: API_KEY from environment, vault, or file
if [[ -f /run/secrets/api_key ]]; then
    API_KEY=$(< /run/secrets/api_key)
fi
```

## 9. ShellCheck

ShellCheck is a static analysis tool that catches bugs:

```bash
# Install:
sudo dnf install ShellCheck

# Run on a script:
shellcheck deploy.sh

# Common things it catches:
# SC2086: Double quote to prevent globbing and word splitting
# SC2046: Quote this to prevent word splitting
# SC2034: Variable appears unused
# SC2155: Declare and assign separately to avoid masking return values
# SC2164: Use cd ... || exit in case cd fails

# Integrate with CI:
shellcheck --severity=warning scripts/*.sh

# Ignore specific rules (when you know what you're doing):
# shellcheck disable=SC2086
echo $unquoted_intentionally

# Editor integration:
# VS Code: "ShellCheck" extension
# Vim: ALE or Syntastic
# Neovim: null-ls
```

## 10. Putting It All Together

```bash
#!/usr/bin/env bash
#
# deploy.sh — Deploy application to target server.
#
# Prerequisites:
#   - rsync, curl, jq installed
#   - DEPLOY_HOST, API_KEY set in environment
#
# Usage:
#   DEPLOY_HOST=prod.internal API_KEY=xxx ./deploy.sh [--dry-run] <version>
#

set -euo pipefail

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"
readonly LOCK_FILE="/var/lock/${SCRIPT_NAME%.sh}.lock"

# ── Config ─────────────────────────────────────────────
DEPLOY_HOST="${DEPLOY_HOST:?DEPLOY_HOST is required}"
API_KEY="${API_KEY:?API_KEY is required}"
DRY_RUN=false
VERSION=""

# ── Logging ────────────────────────────────────────────
log()   { printf '[%s] [%-5s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$1" "${*:2}" >&2; }
info()  { log INFO "$@"; }
warn()  { log WARN "$@"; }
error() { log ERROR "$@"; }
die()   { error "$@"; exit 1; }

# ── Cleanup ────────────────────────────────────────────
cleanup() {
    local exit_code=$?
    rm -rf "$LOCK_FILE" 2>/dev/null || true
    if (( exit_code != 0 )); then
        error "Deploy failed with exit code $exit_code"
        # Notify on failure:
        notify_slack "Deploy $VERSION to $DEPLOY_HOST FAILED" || true
    fi
    exit "$exit_code"
}
trap cleanup EXIT

# ── Helpers ────────────────────────────────────────────
run() {
    if [[ "$DRY_RUN" = true ]]; then
        info "[DRY RUN] $*"
    else
        "$@"
    fi
}

notify_slack() {
    local msg="$1"
    if [[ -n "${SLACK_WEBHOOK:-}" ]]; then
        curl -sf -X POST "$SLACK_WEBHOOK" \
            -H 'Content-type: application/json' \
            -d "$(jq -n --arg text "$msg" '{text: $text}')" \
            || warn "Slack notification failed"
    fi
}

# ── Main ───────────────────────────────────────────────
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --dry-run) DRY_RUN=true; shift ;;
            -h|--help) usage; exit 0 ;;
            -*) die "Unknown option: $1" ;;
            *) VERSION="$1"; shift ;;
        esac
    done
    [[ -n "$VERSION" ]] || die "Version argument required"
}

preflight_checks() {
    for cmd in rsync curl jq; do
        command -v "$cmd" &>/dev/null || die "Required: $cmd"
    done

    # Acquire lock:
    mkdir "$LOCK_FILE" 2>/dev/null || die "Already running (lock: $LOCK_FILE)"
    echo $$ > "$LOCK_FILE/pid"

    # Test connectivity:
    timeout 5 ssh -o ConnectTimeout=3 "$DEPLOY_HOST" true \
        || die "Cannot reach $DEPLOY_HOST"
}

deploy() {
    info "Deploying version $VERSION to $DEPLOY_HOST"

    # Build:
    run make build VERSION="$VERSION"

    # Upload:
    run rsync -avz --delete ./dist/ "$DEPLOY_HOST:/opt/myapp/releases/$VERSION/"

    # Switch:
    run ssh "$DEPLOY_HOST" "ln -sfn /opt/myapp/releases/$VERSION /opt/myapp/current"

    # Restart:
    run ssh "$DEPLOY_HOST" "sudo systemctl restart myapp"

    # Verify:
    info "Waiting for health check..."
    sleep 3
    if ! timeout 30 bash -c "until curl -sf http://$DEPLOY_HOST/health; do sleep 2; done"; then
        die "Health check failed after deploy"
    fi

    info "Deploy successful: $VERSION → $DEPLOY_HOST"
    notify_slack "Deployed $VERSION to $DEPLOY_HOST ✓"
}

main() {
    parse_args "$@"
    preflight_checks
    deploy
}

main "$@"
```

## Exercise

1. Take a script you've written and add: input validation, a cleanup trap, logging, and ShellCheck compliance.

2. Write an idempotent setup script that installs packages, creates directories, creates a system user, and configures a systemd service. Each step should be safe to run repeatedly.

3. Implement the dry-run pattern in a script that modifies files. Verify that `--dry-run` shows what would happen without making changes.

4. Add lock file protection to a script. Test it by running two instances simultaneously.

---

Next: [Safety Patterns](01-safety-patterns.md)
