# dnf, systemd, and journalctl

## What Problem This Solves

These three tools are how you actually manage a Fedora system. You need to install software, manage services, and read logs — and you need to do it from the shell, not just clicking buttons. Most people know `sudo dnf install` and `systemctl restart` but don't go deeper.

## dnf — Package Management

### The Essentials

```bash
# Install:
sudo dnf install nginx

# Install multiple packages:
sudo dnf install nginx postgresql redis

# Remove:
sudo dnf remove nginx

# Update everything:
sudo dnf upgrade

# Update a specific package:
sudo dnf upgrade nginx

# Search for a package:
dnf search "web server"

# Info about a package (installed or available):
dnf info nginx
```

### Beyond the Basics

```bash
# What provides a specific file?
dnf provides /usr/bin/dig
# → bind-utils

# What provides a command you need?
dnf provides '*/jq'

# List installed packages:
dnf list installed
dnf list installed | grep python

# List available updates:
dnf check-update

# History — see what you've installed/removed:
dnf history
dnf history info 42       # Details of transaction 42
dnf history undo 42       # Undo transaction 42

# Groups of packages:
dnf group list
dnf group install "Development Tools"

# Which packages depend on this one:
dnf repoquery --whatrequires python3

# Files in a package:
rpm -ql nginx              # List files in installed package
dnf repoquery -l nginx     # List files (even if not installed)

# Which package owns this file:
rpm -qf /usr/bin/curl      # → curl
```

### Repositories

```bash
# List enabled repos:
dnf repolist

# Add a repo (like RPM Fusion):
sudo dnf install \
  https://mirrors.rpmfusion.org/free/fedora/rpmfusion-free-release-$(rpm -E %fedora).noarch.rpm

# Enable a COPR repo:
sudo dnf copr enable user/project

# Disable a repo temporarily:
sudo dnf install --disablerepo=rpmfusion-free-updates package

# Clean cache:
sudo dnf clean all
```

### dnf in Scripts

```bash
# Non-interactive install (don't prompt for y/n):
sudo dnf install -y nginx

# Check if package is installed:
if rpm -q nginx &>/dev/null; then
    echo "nginx is installed"
else
    echo "nginx is NOT installed"
fi

# Install only if not present:
rpm -q nginx &>/dev/null || sudo dnf install -y nginx
```

---

## systemd — Service Management

### Service Lifecycle

```bash
# Start/stop/restart:
sudo systemctl start nginx
sudo systemctl stop nginx
sudo systemctl restart nginx    # Stop then start
sudo systemctl reload nginx     # Reload config without stopping (if supported)

# Enable (start on boot) / Disable:
sudo systemctl enable nginx     # Start on boot
sudo systemctl disable nginx    # Don't start on boot
sudo systemctl enable --now nginx   # Enable AND start immediately

# Status:
systemctl status nginx
# Shows: loaded, active/inactive, PID, recent logs, memory usage
```

### Inspecting Services

```bash
# Is a service running?
systemctl is-active nginx        # "active" or "inactive"
systemctl is-active --quiet nginx && echo "running"

# Is it enabled (starts on boot)?
systemctl is-enabled nginx

# Show the unit file:
systemctl cat nginx

# Show all settings (including defaults):
systemctl show nginx

# Show specific property:
systemctl show nginx --property=MainPID
systemctl show nginx --property=MemoryCurrent

# List all services:
systemctl list-units --type=service

# List all services (including non-running):
systemctl list-units --type=service --all

# List failed services:
systemctl --failed

# What's starting on boot:
systemctl list-unit-files --type=service --state=enabled
```

### Overriding Service Configuration

```bash
# Edit an override (don't modify the package file!):
sudo systemctl edit nginx
# Creates /etc/systemd/system/nginx.service.d/override.conf

# Example override content:
[Service]
Environment="NGINX_PORT=8080"
LimitNOFILE=65535

# To see the combined (full) unit:
sudo systemctl edit --full nginx
# This copies the entire unit file for editing

# After editing:
sudo systemctl daemon-reload
sudo systemctl restart nginx
```

### Useful systemd Commands

```bash
# List timers (cron replacements):
systemctl list-timers

# See the boot process:
systemd-analyze                    # Total boot time
systemd-analyze blame              # Time per service
systemd-analyze critical-chain     # Dependency chain

# See what depends on what:
systemctl list-dependencies nginx
systemctl list-dependencies --reverse nginx    # What depends on nginx

# Mask a service (prevent starting even manually):
sudo systemctl mask bluetooth
sudo systemctl unmask bluetooth
```

---

## journalctl — System Logs

### The Basics

```bash
# All logs:
journalctl

# Recent logs:
journalctl -n 50               # Last 50 lines
journalctl --since "1 hour ago"
journalctl --since "2024-01-15 10:00" --until "2024-01-15 12:00"
journalctl --since today

# Follow (like tail -f):
journalctl -f

# For a specific service:
journalctl -u nginx
journalctl -u nginx --since "1 hour ago"
journalctl -u nginx -f          # Follow nginx logs

# By priority:
journalctl -p err               # Errors and above
journalctl -p warning           # Warnings and above
# Priorities: emerg, alert, crit, err, warning, notice, info, debug

# Kernel messages:
journalctl -k
journalctl -k --since "1 hour ago"
```

### Power Features

```bash
# By PID:
journalctl _PID=1234

# By executable:
journalctl _COMM=sshd

# By boot:
journalctl -b                   # Current boot
journalctl -b -1                # Previous boot
journalctl --list-boots         # All recorded boots

# JSON output (for processing):
journalctl -u nginx -o json | jq '.'
journalctl -u nginx -o json-pretty | head -50

# Specific fields:
journalctl -u nginx -o json | jq -r '.MESSAGE'

# Disk usage:
journalctl --disk-usage

# Rotate/clean:
sudo journalctl --vacuum-time=30d    # Keep only 30 days
sudo journalctl --vacuum-size=500M   # Keep only 500MB
```

### journalctl in Scripts

```bash
# Check for errors in the last hour:
if journalctl -u myapp --since "1 hour ago" -p err --quiet; then
    echo "No errors"
else
    echo "Errors found!"
    journalctl -u myapp --since "1 hour ago" -p err --no-pager
fi

# Count errors:
error_count=$(journalctl -u myapp --since "1 hour ago" -p err --no-pager | wc -l)

# Extract structured data:
journalctl -u nginx --since today -o json | \
  jq -r 'select(.PRIORITY == "3") | .MESSAGE' | \
  sort | uniq -c | sort -rn | head

# IMPORTANT: Always use --no-pager in scripts!
journalctl -u nginx --no-pager -n 100
```

## Putting Them Together

### Service Health Check Script

```bash
#!/usr/bin/env bash
set -euo pipefail

check_service() {
    local service="$1"
    local status
    status=$(systemctl is-active "$service" 2>/dev/null || true)

    case "$status" in
        active)
            printf "%-20s ✓ running\n" "$service"
            ;;
        *)
            printf "%-20s ✗ %s\n" "$service" "$status"
            # Show recent errors:
            journalctl -u "$service" --since "10 min ago" -p err --no-pager -n 5 2>/dev/null || true
            ;;
    esac
}

echo "=== Service Health Check ==="
for svc in nginx postgresql redis sshd firewalld; do
    check_service "$svc"
done
```

### Quick Diagnostic

```bash
# "Why did this service fail?" workflow:
sudo systemctl status myapp                    # Quick overview
journalctl -u myapp -n 50 --no-pager          # Recent logs
journalctl -u myapp -p err --since today       # Today's errors
systemctl cat myapp                            # Check the config
systemctl show myapp --property=ExecStart      # What command does it run?
```

## Common Footguns

### 1. Editing Package Unit Files Directly

```bash
# WRONG — your changes get overwritten on package update:
sudo vim /usr/lib/systemd/system/nginx.service

# RIGHT — use override:
sudo systemctl edit nginx
```

### 2. Forgetting daemon-reload

```bash
# After editing any unit file:
sudo systemctl daemon-reload    # REQUIRED!
sudo systemctl restart nginx
```

### 3. journalctl Uses a Pager

```bash
# In scripts, the pager blocks execution:
journalctl -u nginx    # Hangs waiting for you to press 'q'

# Fix:
journalctl -u nginx --no-pager
```

## Exercise

1. Find which package provides the `htop` command without installing it. Install it, verify it works, then check `dnf history` for the transaction.

2. Create a simple systemd service that runs a script, enable it, check its status, and read its logs with `journalctl`.

3. Write a "system report" script that uses `systemctl` and `journalctl` to check 5 critical services and report any with errors in the last 24 hours.

4. Use `journalctl -o json` with `jq` to extract and count error messages from a service, grouped by message content.

---

Next: [SELinux from the Shell](01-selinux-from-shell.md)
