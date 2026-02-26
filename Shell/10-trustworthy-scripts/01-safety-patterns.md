# Safety Patterns

## What Problem This Solves

Shell scripts can destroy data in a single typo. `rm -rf $UNSET_VAR/` becomes `rm -rf /`. A missing quote splits a filename and deletes the wrong file. This file collects the defensive patterns that prevent disasters.

## The Catastrophic Mistakes and How to Prevent Them

### 1. The Empty Variable Delete

```bash
# The nightmare:
DEPLOY_DIR=""
rm -rf "$DEPLOY_DIR/"    # Becomes: rm -rf /
```

**Prevention:**

```bash
# Use set -u (catches unset variables):
set -u
rm -rf "$DEPLOY_DIR/"    # Error: DEPLOY_DIR: unbound variable

# Or use ${var:?} for critical variables:
rm -rf "${DEPLOY_DIR:?DEPLOY_DIR is not set}/"
# Exits with error instead of deleting /

# Or validate before use:
[[ -n "$DEPLOY_DIR" ]] || die "DEPLOY_DIR is empty"
[[ "$DEPLOY_DIR" == /* ]] || die "DEPLOY_DIR must be an absolute path"
[[ "$DEPLOY_DIR" != "/" ]] || die "DEPLOY_DIR cannot be /"
```

### 2. Path Injection

```bash
# User provides a path:
target="$1"
rm -rf "/opt/myapp/$target"

# If target is "../../etc":
# rm -rf "/opt/myapp/../../etc" → rm -rf /etc!

# Prevention — validate and canonicalize:
target="$1"
resolved=$(realpath -m "/opt/myapp/$target")
[[ "$resolved" == /opt/myapp/* ]] || die "Path escape detected: $target"
rm -rf "$resolved"
```

### 3. Unquoted Variables in rm

```bash
# Filename with spaces:
file="important document.txt"
rm $file        # Deletes "important" AND "document.txt"
rm "$file"      # Deletes "important document.txt"

# ALWAYS quote variables in destructive commands.
```

## Safe Deletion Patterns

### Delete with Safeguards

```bash
safe_rm() {
    local target="$1"

    # Must be set:
    [[ -n "$target" ]] || { error "safe_rm: empty path"; return 1; }

    # Must exist:
    [[ -e "$target" ]] || { warn "safe_rm: does not exist: $target"; return 0; }

    # Must not be a critical path:
    case "$target" in
        /|/etc|/usr|/var|/home|/boot|/bin|/sbin|/lib|/lib64)
            die "safe_rm: refusing to delete critical path: $target"
            ;;
    esac

    # Must be under an expected parent:
    local allowed_parent="/opt/myapp"
    local resolved
    resolved=$(realpath "$target")
    [[ "$resolved" == "$allowed_parent"/* ]] || \
        die "safe_rm: $target resolves to $resolved which is outside $allowed_parent"

    rm -rf "$target"
}
```

### Use Trash Instead of rm

```bash
# Move to trash instead of deleting:
trash() {
    local trash_dir="${HOME}/.local/share/Trash/files"
    mkdir -p "$trash_dir"
    for f in "$@"; do
        mv "$f" "$trash_dir/$(basename "$f").$(date +%s)"
    done
}

# Or use the trash-cli package:
sudo dnf install trash-cli
trash-put file.txt        # Move to trash
trash-list                # List trashed files
trash-restore             # Restore from trash
```

## Temporary File Safety

### mktemp — The Only Safe Way

```bash
# WRONG — predictable, race condition:
tmp="/tmp/myapp_temp"          # Anyone can predict this
echo "data" > "$tmp"           # Symlink attack possible

# RIGHT — unpredictable, secure:
tmp=$(mktemp)                  # /tmp/tmp.xKz4j2mN1p
echo "data" > "$tmp"

# Temp directory:
tmpdir=$(mktemp -d)            # /tmp/tmp.ABCdef123

# With a template:
tmp=$(mktemp /tmp/myapp.XXXXXX)

# ALWAYS clean up:
cleanup() {
    rm -f "$tmp"
    rm -rf "$tmpdir"
}
trap cleanup EXIT
```

### mktemp in Pipelines

```bash
# Avoid reading and writing the same file:
# WRONG — truncates the file before sed reads it:
sed 's/old/new/' file > file

# RIGHT:
tmp=$(mktemp)
sed 's/old/new/' file > "$tmp" && mv "$tmp" file

# Or use sed -i (which does this internally):
sed -i 's/old/new/' file
```

## Safe File Writing

### Atomic Write Pattern

```bash
safe_write() {
    local target="$1"
    local tmp
    tmp=$(mktemp "${target}.XXXXXX")

    # Write to temp file (caller pipes or redirects into this):
    cat > "$tmp"

    # Set permissions before moving:
    chmod 644 "$tmp"

    # Atomic move:
    mv "$tmp" "$target"
}

# Usage:
echo "new config content" | safe_write /etc/myapp/config.toml

# Or with a function:
generate_config | safe_write /etc/myapp/config.toml
```

### Backup Before Modify

```bash
safe_modify() {
    local file="$1"
    local backup="${file}.bak.$(date +%s)"

    [[ -f "$file" ]] || die "File not found: $file"

    cp "$file" "$backup"
    log INFO "Backup saved: $backup"

    # Return backup path for potential rollback:
    echo "$backup"
}

# Usage:
backup=$(safe_modify /etc/nginx/nginx.conf)
# ... make changes ...
if ! nginx -t; then
    cp "$backup" /etc/nginx/nginx.conf
    die "Config validation failed, restored backup"
fi
```

## Safe sudo Usage

```bash
# WRONG — redirects happen BEFORE sudo:
sudo echo "data" > /etc/config    # Fails! Redirect runs as your user

# RIGHT — use tee:
echo "data" | sudo tee /etc/config > /dev/null

# RIGHT — use a heredoc:
sudo tee /etc/config > /dev/null <<'EOF'
data line 1
data line 2
EOF

# Append:
echo "new line" | sudo tee -a /etc/config > /dev/null
```

## Defensive Coding Patterns

### Guard Clauses

```bash
# Instead of deeply nested ifs:
process_file() {
    local file="$1"
    [[ -n "$file" ]] || return 1
    [[ -f "$file" ]] || return 1
    [[ -r "$file" ]] || return 1
    [[ -s "$file" ]] || return 0    # Empty file is OK, just skip

    # Now the actual logic — no indentation nightmare:
    local content
    content=$(< "$file")
    process "$content"
}
```

### Readonly Variables

```bash
# Prevent accidental modification of critical values:
readonly PROD_DB="prod-db.internal"
readonly BACKUP_DIR="/var/backups"
readonly MAX_RETRIES=5

# If someone accidentally tries to change these:
PROD_DB="localhost"    # Error: PROD_DB: readonly variable
```

### Explicit Failure Points

```bash
# Mark places where failure is expected/handled:
# This suppresses set -e for this command:
if ! result=$(curl -sf "$url" 2>/dev/null); then
    warn "Could not reach $url, using cached data"
    result=$(< "$cache_file")
fi

# Or use || true for commands that may fail non-critically:
rm -f /tmp/optional_file || true
```

### Input Sanitization

```bash
# For user-provided input used in commands:
sanitize_name() {
    local input="$1"
    # Only allow alphanumeric, dash, underscore, dot:
    if [[ ! "$input" =~ ^[a-zA-Z0-9._-]+$ ]]; then
        die "Invalid name: $input (only alphanumeric, dash, underscore, dot allowed)"
    fi
    echo "$input"
}

# Usage:
service_name=$(sanitize_name "$1")
systemctl status "$service_name"

# NEVER put user input directly into eval or bash -c:
# EXTREMELY DANGEROUS:
eval "echo $user_input"        # Code injection!
bash -c "process $user_input"  # Code injection!

# Safer alternatives:
echo "$user_input"             # Just echo it
"$command" "$user_input"       # Pass as argument, not code
```

## Error Recovery Patterns

### Retry with Backoff

```bash
retry_with_backoff() {
    local max_attempts="$1"
    local initial_delay="$2"
    shift 2
    local cmd=("$@")

    local attempt=1
    local delay="$initial_delay"

    while (( attempt <= max_attempts )); do
        if "${cmd[@]}"; then
            return 0
        fi

        if (( attempt == max_attempts )); then
            error "All $max_attempts attempts failed for: ${cmd[*]}"
            return 1
        fi

        warn "Attempt $attempt failed, retrying in ${delay}s..."
        sleep "$delay"
        delay=$(( delay * 2 ))    # Exponential backoff
        (( attempt++ ))
    done
}

# Usage:
retry_with_backoff 5 2 curl -sf "https://api.example.com/health"
```

### Rollback on Failure

```bash
deploy_with_rollback() {
    local version="$1"
    local previous
    previous=$(readlink /opt/myapp/current || echo "")

    # Deploy new version:
    info "Deploying $version (previous: $previous)"
    ln -sfn "/opt/myapp/releases/$version" /opt/myapp/current
    systemctl restart myapp

    # Verify:
    if ! timeout 30 bash -c 'until curl -sf http://localhost/health; do sleep 1; done'; then
        error "Health check failed after deploying $version"

        if [[ -n "$previous" ]]; then
            warn "Rolling back to $previous"
            ln -sfn "$previous" /opt/myapp/current
            systemctl restart myapp
        fi

        die "Deploy failed, rolled back to $previous"
    fi

    info "Deploy successful: $version"
}
```

## Testing Shell Scripts

```bash
# ShellCheck (static analysis):
shellcheck script.sh

# Dry-run mode (built into your script):
DRY_RUN=true ./deploy.sh staging

# Test with a container (isolated environment):
podman run --rm -v "$PWD:/scripts:Z" fedora:latest bash /scripts/test.sh

# Test with set -x (see what's happening):
bash -x script.sh

# Test specific functions by sourcing:
source ./lib/helpers.sh
# Now call functions directly in your terminal
```

## The "Am I About to Destroy Something?" Checklist

Before running a command that modifies or deletes:

1. **Is the variable set?** `echo "$DEPLOY_DIR"` before using it
2. **Does the path make sense?** `ls "$target"` before `rm -rf "$target"`
3. **Am I on the right machine?** `hostname` before destructive ops
4. **Is this idempotent?** Can I safely run it again?
5. **Do I have a backup?** Before modifying production data
6. **What happens if this fails halfway?** Plan for partial failure

## Exercise

1. Write a `safe_rm` function with all the safeguards (non-empty, not critical path, under allowed parent). Test it with dangerous inputs.

2. Implement the atomic write pattern: write new content to a temp file, validate it, then atomically replace the original.

3. Create a deploy script with rollback capability. Simulate a failure (health check fails) and verify the rollback works.

4. Run ShellCheck on 3 of your existing scripts. Fix all warnings. Note the most common issues you had.

---

Next: [Level 11 — Debugging](../11-debugging/00-debugging-tools.md)
