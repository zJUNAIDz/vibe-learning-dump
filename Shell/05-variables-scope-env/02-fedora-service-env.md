# Fedora Service Environments

## What Problem This Solves

You write a script, it works perfectly in your terminal. You put it in a systemd service or cron job, and it fails. Why? Because services don't run in your shell. They get a **completely different environment** — no dotfiles, no PATH you're used to, no variables you "set."

## The Mental Model

```
Your Terminal:
  Login → /etc/profile → ~/.bash_profile → ~/.bashrc
  → You get: full PATH, aliases, functions, custom vars, $HOME, $USER

systemd Service:
  systemd → ExecStart=/path/to/binary
  → You get: minimal PATH, no dotfiles, no aliases, no $HOME (maybe)

Cron Job:
  crond → /bin/sh -c "your command"
  → You get: minimal PATH, SHELL=/bin/sh, almost nothing
```

## systemd Environment

### What a Service Actually Gets

Check what environment a service sees:

```bash
# Create a test service:
sudo tee /etc/systemd/system/env-test.service << 'EOF'
[Unit]
Description=Environment Test

[Service]
Type=oneshot
ExecStart=/usr/bin/env
StandardOutput=journal
EOF

sudo systemctl daemon-reload
sudo systemctl start env-test
sudo journalctl -u env-test --no-pager -n 50
```

You'll see something like:

```
LANG=en_US.UTF-8
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin
INVOCATION_ID=...
JOURNAL_STREAM=...
SYSTEMD_EXEC_PID=...
```

**That's it.** No $HOME, no custom PATH additions, no $GOPATH, no $NVM_DIR. Nothing from your dotfiles.

### Setting Environment for Services

```ini
# Method 1: Inline in the unit file
[Service]
Environment="NODE_ENV=production"
Environment="PORT=3000"
Environment="DB_HOST=localhost"

# Method 2: Environment file
[Service]
EnvironmentFile=/etc/myapp/config.env

# Method 3: Multiple environment files (later overrides earlier)
[Service]
EnvironmentFile=/etc/myapp/defaults.env
EnvironmentFile=/etc/myapp/overrides.env

# Method 4: Optional environment file (- prefix means don't fail if missing)
[Service]
EnvironmentFile=-/etc/myapp/optional.env
```

### Environment File Format

```bash
# /etc/myapp/config.env
# This is NOT a shell script. The format is simpler:

# Correct format:
NODE_ENV=production
PORT=3000
DATABASE_URL=postgres://localhost/mydb

# Quotes are included literally (different from shell!):
# WRONG — value will be "production" including the quotes:
NODE_ENV="production"

# No variable expansion:
# WRONG — $HOME won't expand:
LOG_DIR=$HOME/logs

# No command substitution:
# WRONG — $(hostname) won't execute:
SERVER_NAME=$(hostname)

# Comments work with #
# Blank lines are ignored
```

**Critical difference:** systemd environment files are NOT sourced as shell scripts. They're parsed with simpler rules. No expansion, no command substitution. If you need those features, use a wrapper script.

### Fixing PATH for Services

```ini
# If your binary is in /usr/local/go/bin:
[Service]
Environment="PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin"
ExecStart=/usr/local/go/bin/myapp

# Or just use the full path (preferred):
[Service]
ExecStart=/usr/local/go/bin/myapp
```

### The User= Directive

```ini
# Run as a specific user:
[Service]
User=myapp
Group=myapp

# This sets $USER and $HOME but NOT the user's dotfiles!
# The service does NOT run a login shell.
```

### Checking a Service's Environment

```bash
# See the effective environment of a running service:
sudo systemctl show myapp --property=Environment

# See all environment variables of a running service's process:
sudo cat /proc/$(systemctl show myapp --property=MainPID --value)/environ | tr '\0' '\n'

# Override environment for testing:
sudo systemctl edit myapp
# Adds an override file. Add:
# [Service]
# Environment="DEBUG=true"
```

## Cron Environment

Cron is even more minimal:

```bash
# What cron gives you:
SHELL=/bin/sh          # Not bash! Not zsh!
PATH=/usr/bin:/bin     # Bare minimum
HOME=/home/user        # At least this
LOGNAME=user

# Your ~/.bashrc is NOT loaded.
# Your aliases don't exist.
# Your custom PATH entries are gone.
```

### Making Cron Work

```bash
# Bad — relies on PATH and environment:
# */5 * * * * backup.sh

# Good — full paths, explicit environment:
# */5 * * * * /home/me/scripts/backup.sh

# Better — set PATH in crontab:
# PATH=/usr/local/bin:/usr/bin:/bin:/home/me/.local/bin
# */5 * * * * /home/me/scripts/backup.sh

# Or source your profile (but this is fragile):
# */5 * * * * bash -lc '/home/me/scripts/backup.sh'
```

### Systemd Timers (The Modern Alternative)

```ini
# /etc/systemd/system/backup.timer
[Unit]
Description=Daily backup

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
```

```ini
# /etc/systemd/system/backup.service
[Unit]
Description=Run backup script

[Service]
Type=oneshot
User=myapp
ExecStart=/home/myapp/scripts/backup.sh
EnvironmentFile=/etc/myapp/backup.env
```

```bash
sudo systemctl enable --now backup.timer
systemctl list-timers          # See all timers
```

Timers give you: better logging (journalctl), proper environment control, dependency management, and no surprises.

## The Wrapper Script Pattern

When a service needs complex environment setup:

```bash
#!/usr/bin/env bash
# /opt/myapp/run.sh — Wrapper script for systemd service

# Now we're in a real shell, so we can do real things:
export PATH="/opt/myapp/bin:$PATH"
export APP_CONFIG="/etc/myapp/config.toml"
export LOG_LEVEL="${LOG_LEVEL:-info}"

# Load secrets from a file:
if [[ -f /run/secrets/db_password ]]; then
    export DB_PASSWORD
    DB_PASSWORD=$(< /run/secrets/db_password)
fi

# Dynamic values:
export HOSTNAME
HOSTNAME=$(hostname -f)

exec /opt/myapp/bin/myapp "$@"
# exec replaces this shell with the app process — no extra shell process hanging around
```

```ini
[Service]
ExecStart=/opt/myapp/run.sh
```

**Key insight:** `exec` replaces the shell process with the app. systemd then monitors the app directly, not the wrapper shell.

## Common Footguns

### 1. "But It Works When I Run It Manually!"

```bash
# You test:
sudo bash /opt/myapp/start.sh     # Works!
# But:
sudo systemctl start myapp        # Fails!

# Because sudo preserves more environment than systemd gives.
# Test like systemd would:
sudo -i env -i PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin bash /opt/myapp/start.sh
```

### 2. SELinux Context

```bash
# Files you create in your terminal have context:
ls -Z /home/you/script.sh
# unconfined_u:object_r:user_home_t:s0

# systemd services run with different SELinux contexts:
# Your script may not be able to access files it could before.
# Check the audit log:
sudo ausearch -m avc -ts recent

# Fix — set the right context:
sudo semanage fcontext -a -t bin_t "/opt/myapp/bin(/.*)?"
sudo restorecon -Rv /opt/myapp/bin/
```

### 3. /tmp Is Different

```bash
# systemd can give services private /tmp:
[Service]
PrivateTmp=true    # Many Fedora services have this by default!

# Your service's /tmp is NOT the same /tmp you see in your terminal.
# Files written to /tmp won't be where you expect.
# Check: /tmp/systemd-private-*-myapp.service-*/tmp/
```

## Exercise

1. Create a systemd oneshot service that dumps its full environment to a file. Compare that with your shell's `env` output.

2. Write a script that works in your terminal but fails as a cron job because of PATH. Fix it three different ways:
   - Full path in the script
   - PATH in the crontab
   - Wrapper script that sets up the environment

3. Create a systemd service with `EnvironmentFile=`. Put variables in the env file and verify the service sees them.

4. Use the wrapper script pattern: write a wrapper that sets up environment, then `exec`s a simple program (like a Python HTTP server).

---

Next: [Level 6 — Bash as a Scripting Language](../06-bash-scripting-language/00-functions.md)
