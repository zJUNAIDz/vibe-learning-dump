# Environment Differences

## What Problem This Solves

Your script works perfectly on your laptop. Then it fails in cron. Then it fails in CI. Then it fails in Docker. Then it fails when someone SSHes in. Every environment strips away things you took for granted, and if you don't understand what's different, you'll keep getting bitten.

## The Core Problem

Your interactive terminal session has accumulated state: dotfiles loaded, aliases defined, PATH populated, locale set, terminal type configured. Scripts in other environments have *none of this* by default.

### The Environment Ladder

From most to least configured:

```
┌────────────────────────────────────────────────┐
│  1. Interactive login shell                     │
│     Has: EVERYTHING - dotfiles, aliases, PATH   │
│     Where: You open a terminal, SSH login       │
├────────────────────────────────────────────────┤
│  2. Interactive non-login shell                 │
│     Has: .bashrc/.zshrc but not login files     │
│     Where: New terminal tab (Bash), subshell    │
├────────────────────────────────────────────────┤
│  3. Non-interactive shell (scripts)             │
│     Has: Inherited env only                     │
│     Where: bash script.sh, ./script.sh          │
├────────────────────────────────────────────────┤
│  4. Cron                                        │
│     Has: Minimal env (HOME, LOGNAME, SHELL)     │
│     Where: Scheduled tasks                      │
├────────────────────────────────────────────────┤
│  5. systemd service                             │
│     Has: Almost nothing by default              │
│     Where: systemd units                        │
├────────────────────────────────────────────────┤
│  6. Docker container                            │
│     Has: Only what you explicitly set           │
│     Where: Containerized deployments            │
└────────────────────────────────────────────────┘
```

Each level down strips away more assumptions.

## Interactive vs Non-Interactive

### How to Check

```bash
# Check if the current shell is interactive:
case "$-" in
    *i*) echo "Interactive" ;;
    *)   echo "Non-interactive" ;;
esac

# Bash-specific:
[[ $- == *i* ]] && echo "Interactive"

# Shell also checks PS1:
[[ -n "${PS1}" ]] && echo "Probably interactive"
```

### What Changes

```bash
# Things available interactively but NOT in scripts:
# 1. Aliases
alias ll='ls -la'
# In a script, `ll` → command not found

# 2. Job control (bg, fg, Ctrl+Z)
# Scripts don't have a controlling terminal

# 3. History
# Scripts don't load or record history

# 4. Prompt (PS1, PS2)
# No prompt in non-interactive mode

# 5. .bashrc is NOT loaded (by default)
# Functions defined in .bashrc aren't available

# 6. Completion
# Tab completion isn't loaded
```

### When .bashrc IS Loaded for Scripts

```bash
# Bash checks if BASH_ENV is set for non-interactive shells:
export BASH_ENV=~/.bashrc
bash script.sh    # Now .bashrc IS loaded

# But this is almost never what you want.
# Instead, source what you need explicitly:

#!/usr/bin/env bash
source /path/to/shared-functions.sh
# Now use the functions
```

## Cron's Minimal Environment

### What Cron Provides

```bash
# Cron's default environment (approximately):
SHELL=/bin/sh         # NOT bash!
HOME=/home/username
LOGNAME=username
PATH=/usr/bin:/bin    # Very limited!
# That's it. No USER, no DISPLAY, no TERM, no locale.
```

### Common Cron Failures

```bash
# Failure 1: "command not found"
# Your PATH in terminal: /usr/local/bin:/usr/bin:/bin:/home/you/.local/bin
# Cron's PATH: /usr/bin:/bin
# Fix: Use absolute paths

# BAD cron entry:
* * * * * backup-script.sh

# GOOD cron entry:
* * * * * /usr/local/bin/backup-script.sh

# Failure 2: Script relies on shell features
# Cron uses /bin/sh by default (may be dash, not bash)
# Fix: Specify shell in crontab or use shebang
SHELL=/bin/bash
* * * * * /home/user/backup.sh

# Failure 3: Script works but you can't see errors
# Fix: Redirect output
* * * * * /home/user/backup.sh >> /var/log/backup.log 2>&1

# Failure 4: Environment variables missing
# Fix: Set them in the crontab or source a file
PATH=/usr/local/bin:/usr/bin:/bin
MAILTO=user@example.com
* * * * * /home/user/backup.sh
```

### The Cron-Proof Script Pattern

```bash
#!/usr/bin/env bash
set -euo pipefail

# Don't assume PATH — set it explicitly
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

# Don't assume working directory
cd "$(dirname "$0")" || exit 1

# Don't assume locale
export LANG=en_US.UTF-8
export LC_ALL=en_US.UTF-8

# Don't assume HOME (for root cron)
export HOME="${HOME:-/root}"

# Log everything
exec >> /var/log/myapp/backup.log 2>&1
echo "=== $(date '+%Y-%m-%d %H:%M:%S') ==="

# Now your actual logic
# ...
```

## systemd Service Environment

### What systemd Provides

```bash
# systemd units get almost nothing:
# - No PATH (or a very minimal one)
# - No HOME
# - No USER
# - No shell dotfiles
# - No TTY
# - Different /tmp (PrivateTmp=yes)
# - Different filesystem view (ProtectSystem, ProtectHome)
# - SELinux context may differ

# Check what a service sees:
sudo systemctl show myservice --property=Environment
systemctl cat myservice    # See the unit file
```

### Setting Environment for Services

```bash
# Method 1: In the unit file
[Service]
Environment="NODE_ENV=production"
Environment="PORT=3000"
# Each variable needs its own line or all in one:
Environment="NODE_ENV=production" "PORT=3000"

# Method 2: From a file
[Service]
EnvironmentFile=/etc/myapp/config
# File format (NOT shell syntax!):
# PORT=3000
# NODE_ENV=production
# No export, no quotes around values (unless you want literal quotes)

# Method 3: Override just the environment
sudo systemctl edit myservice
# Opens an override file where you can add:
[Service]
Environment="DEBUG=true"
```

### Common systemd Pitfalls

```bash
# Pitfall 1: "Works with sudo, fails as service"
# Because sudo preserves more environment than systemd
# Debug: add this to the service:
ExecStartPre=/usr/bin/env > /tmp/service-env-dump.txt

# Pitfall 2: "Can't find my files"
# WorkingDirectory isn't set — defaults to /
[Service]
WorkingDirectory=/opt/myapp

# Pitfall 3: PrivateTmp
# Service sees /tmp/systemd-private-XXX/tmp instead of /tmp
# Files other processes put in /tmp aren't visible
[Service]
PrivateTmp=false    # Only if you actually need shared /tmp

# Pitfall 4: PATH doesn't include /usr/local/bin
[Service]
Environment="PATH=/usr/local/bin:/usr/bin:/bin"
```

## Docker Environment

### What's Different in Containers

```bash
# 1. No systemd (usually)
# Can't use systemctl inside most containers

# 2. Minimal tools
# No vim, no less, no man, often no bash
docker exec -it container sh    # May need sh, not bash

# 3. Different user
# May be root or a custom user
whoami
id

# 4. Ephemeral filesystem
# Changes disappear when container restarts

# 5. Different hostname, IPs, DNS
hostname
cat /etc/resolv.conf

# 6. Environment from docker run
docker run -e "DB_HOST=postgres" -e "DB_PORT=5432" myapp
```

### Scripts That Run in Docker

```bash
#!/bin/sh
# Use /bin/sh, not /bin/bash — bash may not be installed!

# Check for required tools:
for cmd in curl jq; do
    command -v "$cmd" >/dev/null 2>&1 || {
        echo "ERROR: $cmd is required but not installed" >&2
        exit 1
    }
done

# Check for required environment variables:
: "${DB_HOST:?DB_HOST must be set}"
: "${DB_PORT:?DB_PORT must be set}"

# Use POSIX-compatible syntax:
# No [[ ]], no arrays, no process substitution
if [ -z "$OPTIONAL_VAR" ]; then
    OPTIONAL_VAR="default"
fi
```

## CI/CD Environment (GitHub Actions, Jenkins, etc.)

### What's Different

```bash
# 1. Different OS/distro (often Ubuntu, not Fedora)
cat /etc/os-release

# 2. Different shell (might be dash as /bin/sh)
readlink -f /bin/sh

# 3. Environment variables injected by CI
# GITHUB_ACTIONS=true, CI=true, JENKINS_URL=...

# 4. Non-interactive (no TTY)
tty    # Returns "not a tty"

# 5. Different user (often root or a CI-specific user)
id

# 6. Ephemeral workspace (clean checkout each run)
# Can't rely on cached data from previous runs
```

### CI-Proof Scripts

```bash
#!/usr/bin/env bash
set -euo pipefail

# Detect CI environment:
is_ci() {
    [[ -n "${CI:-}" ]] || [[ -n "${GITHUB_ACTIONS:-}" ]] || [[ -n "${JENKINS_URL:-}" ]]
}

# Adjust behavior for CI:
if is_ci; then
    # No color in CI (unless terminal supports it)
    NO_COLOR=1
    # Don't prompt for input
    INTERACTIVE=false
    # Use CI-specific tmp
    TMPDIR="${RUNNER_TEMP:-/tmp}"
else
    NO_COLOR=0
    INTERACTIVE=true
    TMPDIR="/tmp"
fi

# Don't assume tools are installed:
ensure_tool() {
    local tool="$1"
    if ! command -v "$tool" &>/dev/null; then
        echo "Installing $tool..." >&2
        if command -v dnf &>/dev/null; then
            sudo dnf install -y "$tool"
        elif command -v apt-get &>/dev/null; then
            sudo apt-get install -y "$tool"
        else
            echo "ERROR: Cannot install $tool — unknown package manager" >&2
            exit 1
        fi
    fi
}
```

## SSH Environment Gotchas

```bash
# Interactive SSH login: gets login shell treatment
ssh user@host          # Loads .bash_profile, .bashrc

# SSH with a command: NON-LOGIN, NON-INTERACTIVE
ssh user@host 'echo $PATH'    # Minimal environment!
# .bash_profile NOT loaded
# .bashrc may or may not load (depends on distro)

# Force loading environment:
ssh user@host 'source ~/.bash_profile && echo $PATH'

# Or use login shell:
ssh user@host 'bash -l -c "echo \$PATH"'

# SSH environment file:
# ~/.ssh/environment (if PermitUserEnvironment is enabled)
# Usually disabled for security
```

## Locale and Character Encoding

```bash
# Check current locale:
locale

# What changes:
# LANG=C                → ASCII only, fast sorting
# LANG=en_US.UTF-8      → UTF-8, locale-aware sorting
# LC_ALL=C              → Overrides everything to C locale

# Why it matters in scripts:
# 1. Sort order changes:
echo -e "a\nB\nc" | LANG=C sort           # B a c  (ASCII order)
echo -e "a\nB\nc" | LANG=en_US.UTF-8 sort # a B c  (locale order)

# 2. Character classes change:
echo "Hello" | LANG=C grep '[a-z]'           # Matches 'ello'
echo "Hello" | LANG=en_US.UTF-8 grep '[a-z]' # May match 'Hello' (locale-dependent!)

# 3. Number formatting:
# Some locales use comma as decimal separator
printf "%'.2f\n" 1234567.89   # Locale-dependent output

# Best practice for scripts:
export LC_ALL=C    # Predictable behavior for text processing
```

## Terminal and TTY

```bash
# Check if stdout is a terminal:
if [[ -t 1 ]]; then
    echo "stdout is a terminal — use color"
else
    echo "stdout is a pipe or file — no color"
fi

# Check for stdin:
if [[ -t 0 ]]; then
    echo "Can read from keyboard"
else
    echo "Reading from pipe or file"
fi

# Programs that behave differently:
ls       # Columns when terminal, one-per-line when piped
grep     # Color when terminal, no color when piped
git      # Pager when terminal, plain when piped

# Force behavior:
ls --color=always | less -R    # Force color into pipe
git --no-pager log             # Suppress pager
```

## The Universal Script Template

A script that works everywhere:

```bash
#!/usr/bin/env bash
# ^ env finds bash regardless of path

set -euo pipefail

# === Environment Normalization ===
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH"
export LC_ALL=C
export LANG=C

# Resolve script location regardless of how it was called:
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Minimal required Bash version:
if (( BASH_VERSINFO[0] < 4 )); then
    echo "ERROR: Bash 4+ required (have $BASH_VERSION)" >&2
    exit 1
fi

# Required commands:
for cmd in grep sed awk; do
    command -v "$cmd" >/dev/null 2>&1 || {
        echo "ERROR: Required command '$cmd' not found" >&2
        exit 1
    }
done

# Required environment variables:
: "${REQUIRED_VAR:?ERROR: REQUIRED_VAR must be set}"

# Optional with defaults:
LOG_LEVEL="${LOG_LEVEL:-info}"
DRY_RUN="${DRY_RUN:-false}"

# === Logging ===
log() {
    printf '%s [%s] %s\n' "$(date -Iseconds)" "${1^^}" "$2" >&2
}

# === Main Logic ===
main() {
    log info "Starting from $SCRIPT_DIR"
    log info "Running as $(id -un) on $(hostname)"
    # ...
}

main "$@"
```

## Common Footguns

```bash
# 1. Relying on ~/.bashrc functions in scripts
# Fix: source a shared library file explicitly

# 2. Using $HOME without checking it's set
# In systemd: HOME might not be set
# Fix: "${HOME:-$(getent passwd "$(id -un)" | cut -d: -f6)}"

# 3. Assuming /tmp is shared
# systemd PrivateTmp creates isolated /tmp
# Fix: use mktemp, don't hardcode /tmp paths for IPC

# 4. Assuming bash is at /bin/bash
# macOS: /bin/bash is 3.2, newer bash at /usr/local/bin/bash
# Docker: might not have bash at all
# Fix: #!/usr/bin/env bash

# 5. Using bash-specific syntax in #!/bin/sh scripts
# On Debian/Ubuntu, /bin/sh is dash (not bash)
# Fix: Use POSIX syntax or change shebang to bash
```

## Exercise

1. Create a script that prints its full environment context: interactive/non-interactive, login/non-login, shell version, PATH, locale, whether it has a TTY, and whether it's in a container. Run it in your terminal, via cron, and via `bash script.sh`.

2. Write a cron job that fails because of a missing PATH entry. Use `journalctl` or mail output to diagnose. Fix it using the cron-proof pattern.

3. Create a systemd service unit that runs a script. Compare the environment it gets (using `env > /tmp/env.txt`) with your interactive environment. Count the differences.

4. Write a script that uses `bash -x` to trace itself into a separate file while keeping stdout/stderr clean for the user.

---

Next: [Shell Literacy Checklist](../12-shell-literacy-checklist/00-checklist.md)
