# systemd and Service Management

🟢 **Fundamentals** → 🟡 **Intermediate**

---

## What Is systemd?

**systemd** is the **init system** and **service manager** for Linux.

Translation: It's the first process that starts (PID 1) and manages all other services.

```
Boot sequence:
1. Kernel loads
2. Kernel starts PID 1 (systemd)
3. systemd reads configuration
4. systemd starts services (network, SSH, Docker, etc.)
5. System ready
```

---

## Why systemd Exists

**Old init systems (SysVinit):**
- Sequential startup (slow)
- Shell scripts (`/etc/init.d/`)
- No dependency management
- No auto-restart
- Lots of manual work

**systemd:**
- Parallel startup (fast)
- Declarative config files
- Dependency management
- Auto-restart on failure
- Logs managed centrally (journald)

---

## Why You Care as a Developer

1. **Your app runs as a systemd service** (on VMs, bare metal)
2. **Containers use systemd concepts** (though PID 1 is your app, not systemd)
3. **Debugging servers** — check service status, logs, crashes
4. **Auto-restart policies** — same concept in Docker/Kubernetes

---

## Core Concepts

### Unit
A **unit** is anything systemd manages.

**Types of units:**
- `.service` → Background services (nginx, docker, your app)
- `.socket` → Network/IPC sockets
- `.timer` → Scheduled tasks (like cron)
- `.target` → Group of units (like "multi-user.target")
- `.mount` → Filesystem mounts
- `.device` → Hardware devices

**Most common:** `.service` units.

---

## Working with Services

### Check status
```bash
# Is Docker running?
systemctl status docker

# Output:
● docker.service - Docker Application Container Engine
   Loaded: loaded (/usr/lib/systemd/system/docker.service; enabled)
   Active: active (running) since Mon 2026-01-15 10:00:00
   ...
```

**Key fields:**
- `Loaded` → Is the unit file loaded? Is it enabled (auto-start on boot)?
- `Active` → Current state (running, stopped, failed)

---

### Start/stop services
```bash
# Start a service
sudo systemctl start docker

# Stop a service
sudo systemctl stop docker

# Restart (stop then start)
sudo systemctl restart docker

# Reload config without restarting
sudo systemctl reload docker
```

---

### Enable/disable auto-start
```bash
# Enable (auto-start on boot)
sudo systemctl enable docker

# Disable (don't auto-start on boot)
sudo systemctl disable docker

# Enable AND start immediately
sudo systemctl enable --now docker
```

---

### View logs
```bash
# Show logs for a service
sudo journalctl -u docker

# Follow logs (like tail -f)
sudo journalctl -u docker -f

# Show last 50 lines
sudo journalctl -u docker -n 50

# Show logs since boot
sudo journalctl -u docker -b

# Show logs for last 1 hour
sudo journalctl -u docker --since "1 hour ago"
```

---

## Creating a Service for Your App

Let's say you have a Go service: `/home/user/myapp`

### 1. Create a systemd unit file

```bash
sudo vi /etc/systemd/system/myapp.service
```

**Content:**
```ini
[Unit]
Description=My Go Application
After=network.target
# Wait for network to be ready before starting

[Service]
Type=simple
User=myuser
Group=myuser
WorkingDirectory=/home/myuser/myapp
ExecStart=/home/myuser/myapp/bin/myapp
Restart=on-failure
RestartSec=5s

# Environment variables
Environment="PORT=8080"
Environment="LOG_LEVEL=info"

# Limits
LimitNOFILE=4096

# Logging
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

---

### 2. Reload systemd
```bash
# Tell systemd about the new unit
sudo systemctl daemon-reload
```

---

### 3. Start your service
```bash
# Start
sudo systemctl start myapp

# Check status
sudo systemctl status myapp

# Enable auto-start
sudo systemctl enable myapp
```

---

### 4. View logs
```bash
sudo journalctl -u myapp -f
```

---

## Service Types

| Type | Behavior | When to Use |
|------|----------|-------------|
| `simple` | systemd considers it started immediately | Most apps (HTTP servers, APIs) |
| `forking` | Process forks and parent exits | Old-style daemons (nginx, Apache) |
| `oneshot` | Runs once and exits | Setup scripts, one-time tasks |
| `notify` | Process sends "ready" signal to systemd | Apps using sd_notify (Go: systemd library) |

**Most common:** `simple`

---

## Restart Policies

```ini
# Never restart
Restart=no

# Restart only on failure (exit code ≠ 0)
Restart=on-failure

# Always restart (even on success)
Restart=always

# Restart unless explicitly stopped
Restart=on-abnormal
```

**Delay between restarts:**
```ini
RestartSec=5s    # Wait 5 seconds before restarting
```

**This is the same concept as:**
- Docker: `--restart=always`
- Kubernetes: `restartPolicy: Always`

---

## Dependencies

### `After=` vs `Requires=` vs `Wants=`

**After:**
```ini
After=network.target
# Start AFTER network is ready (but don't fail if network fails)
```

**Requires:**
```ini
Requires=postgresql.service
# Start AFTER postgresql, and FAIL if postgresql fails
```

**Wants:**
```ini
Wants=redis.service
# Start AFTER redis (if available), but DON'T fail if redis fails
```

**Common pattern:**
```ini
After=network.target
Wants=postgresql.service
# Start after network, try to start after postgresql (but tolerate failure)
```

---

## Environment Variables

### Option 1: Inline
```ini
Environment="PORT=8080"
Environment="DB_HOST=localhost"
```

### Option 2: Environment file
```ini
EnvironmentFile=/etc/myapp/env
```

**`/etc/myapp/env`:**
```
PORT=8080
DB_HOST=localhost
DB_PASS=secret
```

**Security tip:** Set permissions so only root/service user can read:
```bash
sudo chmod 600 /etc/myapp/env
```

---

## Resource Limits (cgroups Integration!)

systemd uses **cgroups** to enforce limits.

```ini
[Service]
# Memory limit
MemoryLimit=512M

# CPU limit (50% of one core)
CPUQuota=50%

# Max file descriptors
LimitNOFILE=4096

# Max processes
TasksMax=10
```

**This is the same as:**
- Docker: `--memory=512m --cpus=0.5`
- Kubernetes: `resources.limits.memory`

---

## Timers (Replacing cron)

**systemd timers** are like cron, but more powerful.

### Example: Run a backup script daily

**1. Create the service:**
`/etc/systemd/system/backup.service`
```ini
[Unit]
Description=Backup Script

[Service]
Type=oneshot
ExecStart=/usr/local/bin/backup.sh
```

**2. Create the timer:**
`/etc/systemd/system/backup.timer`
```ini
[Unit]
Description=Run backup daily

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
```

**3. Enable and start:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable backup.timer
sudo systemctl start backup.timer

# Check when it'll run next
systemctl list-timers
```

---

## Targets (Like Runlevels)

A **target** groups units together.

| Target | Purpose | Similar to SysVinit |
|--------|---------|---------------------|
| `multi-user.target` | Multi-user mode, no GUI | Runlevel 3 |
| `graphical.target` | Multi-user + GUI | Runlevel 5 |
| `rescue.target` | Single-user mode | Runlevel 1 |

**Check current target:**
```bash
systemctl get-default
```

**Change default:**
```bash
sudo systemctl set-default multi-user.target
```

---

## journald (Logs)

systemd includes **journald**, a centralized logging daemon.

**Advantages:**
- Structured logs (JSON-like)
- Indexed (fast searching)
- Automatic rotation

**Viewing logs:**
```bash
# All logs
journalctl

# Last 100 lines
journalctl -n 100

# Follow (tail -f)
journalctl -f

# Specific service
journalctl -u myapp

# Since boot
journalctl -b

# Time range
journalctl --since "2026-01-15 10:00:00" --until "2026-01-15 11:00:00"

# By priority (emerg, alert, crit, err, warning, notice, info, debug)
journalctl -p err
```

---

## War Story: The Service That Wouldn't Start

A developer deployed a Node.js app as a systemd service. It kept failing.

```bash
sudo systemctl status myapp
● myapp.service - My Node App
   Active: failed (Result: exit-code)
```

**Checking logs:**
```bash
sudo journalctl -u myapp -n 20
```

**Output:**
```
Error: Cannot find module 'express'
```

**The issue:**
The service was running as user `myapp`, but `node_modules` were installed as `root`.

**Fix 1 (quick):**
```bash
sudo chown -R myapp:myapp /home/myapp/node_modules
```

**Fix 2 (proper):**
Update the unit file to run `npm install` before starting:
```ini
ExecStartPre=/usr/bin/npm install --production
ExecStart=/usr/bin/node server.js
```

---

## Key Takeaways

1. **systemd is PID 1** — manages all services
2. **Units are declarative** — describe desired state, systemd handles it
3. **Service types matter** — `simple` for most, `forking` for old daemons
4. **Restart policies** — same concept as Docker/Kubernetes
5. **Dependencies** — `After`, `Requires`, `Wants` control startup order
6. **Resource limits = cgroups** — systemd uses cgroups under the hood
7. **journald centralizes logs** — `journalctl` is your friend
8. **Timers replace cron** — more reliable, easier debugging

---

## Exercises

1. **Create a simple service:**
   - Write a Go/TypeScript app that prints "Hello" every 5 seconds
   - Create a systemd service for it
   - Enable auto-start
   - Reboot and verify it starts

2. **Test restart policy:**
   - Create a service that exits after 10 seconds
   - Set `Restart=on-failure`
   - Verify it restarts automatically

3. **Explore journald:**
   ```bash
   # Show all errors from boot
   journalctl -b -p err
   
   # Show logs from last hour
   journalctl --since "1 hour ago"
   ```

4. **Create a timer:**
   - Write a script that logs the date to `/tmp/timestamp.log`
   - Create a timer to run it every minute
   - Verify with `systemctl list-timers`

---

## Module 01 Complete! 🎉

You now understand:
- ✅ Processes, signals, zombies
- ✅ Memory (stack, heap, virtual memory)
- ✅ CPU scheduling, threads, context switching
- ✅ Networking (TCP/UDP, ports, DNS)
- ✅ cgroups and namespaces (container primitives)
- ✅ systemd (service management)

**This is the foundation for everything else.** Containers, Kubernetes, CI/CD — they all build on these concepts.

---

**Next Module:** [02. Containers Deep Dive →](../02-containers-deep-dive/01-what-containers-actually-are.md)
