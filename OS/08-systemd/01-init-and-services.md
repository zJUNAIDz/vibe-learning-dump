# systemd: The Modern Init System

**Understanding Linux's Service Manager**

🟢 **Beginner-Friendly** | 🟡 **Intermediate**

---

## Introduction

If you've ever run:

```bash
$ sudo systemctl restart nginx
$ journalctl -u docker
$ systemctl status postgresql
```

You're using **systemd** — the modern init system that starts, stops, and manages services on most Linux distributions.

---

## What is systemd?

**systemd is:**
- **Init system** — First process (PID 1) that starts when Linux boots
- **Service manager** — Starts and monitors background services
- **Logger** — Collects all system and service logs
- **Many other things** — Manages devices, mounts, timers, sockets, etc.

**What it replaced:**
- SysV init scripts (`/etc/init.d/`)
- Upstart
- Traditional syslog

---

## systemd as PID 1

When Linux kernel finishes booting:

```
Kernel
  ↓
Executes /sbin/init (which is actually /lib/systemd/systemd)
  ↓
systemd becomes PID 1
  ↓
Reads configuration
  ↓
Starts all enabled services
```

**Check:**

```bash
$ ps -p 1
  PID TTY      STAT   TIME COMMAND
    1 ?        Ss     0:03 /lib/systemd/systemd --system

$ ls -l /sbin/init
lrwxrwxrwx 1 root root 20 ... /sbin/init -> /lib/systemd/systemd
```

**Responsibilities of PID 1:**
- Start all services
- Reap orphaned processes (adopt zombies)
- Handle system shutdown
- Never exit (if PID 1 dies, kernel panics)

---

## Units: systemd's Building Blocks

Everything in systemd is a **unit**. Types:

| Unit Type | Purpose | Example |
|-----------|---------|---------|
| `.service` | Background service | `nginx.service`, `docker.service` |
| `.socket` | Network/IPC socket | `docker.socket`, `sshd.socket` |
| `.timer` | Scheduled task (like cron) | `backup.timer` |
| `.mount` | Filesystem mount | `home.mount` |
| `.device` | Hardware device | `dev-sda.device` |
| `.target` | Group of units | `multi-user.target` |

**Most common: `.service` units**

---

## Service Units

### Example: Nginx Service

```bash
$ systemctl cat nginx
# /lib/systemd/system/nginx.service
[Unit]
Description=A high performance web server
Documentation=man:nginx(8)
After=network.target

[Service]
Type=forking
PIDFile=/run/nginx.pid
ExecStartPre=/usr/sbin/nginx -t -q -g 'daemon on; master_process on;'
ExecStart=/usr/sbin/nginx -g 'daemon on; master_process on;'
ExecReload=/bin/kill -s HUP $MAINPID
ExecStop=/bin/kill -s QUIT $MAINPID
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

### Sections

**[Unit]**
- `Description=` — Human-readable description
- `After=` — Start after these units
- `Before=` — Start before these units
- `Requires=` — Hard dependency (if it fails, this fails)
- `Wants=` — Soft dependency (if it fails, this continues)

**[Service]**
- `Type=` — Service type (see below)
- `ExecStart=` — Command to start service
- `ExecStop=` — Command to stop service
- `ExecReload=` — Command to reload config
- `Restart=` — When to restart (on-failure, always, no)
- `RestartSec=` — Wait before restarting
- `User=`, `Group=` — Run as this user/group
- `Environment=` — Set environment variables
- `WorkingDirectory=` — Start in this directory

**[Install]**
- `WantedBy=` — Which target wants this service
- `RequiredBy=` — Which target requires this service

---

## Service Types

**`Type=simple` (default)**
```bash
[Service]
Type=simple
ExecStart=/usr/bin/my-server
```
- systemd starts the process
- Considers it "started" immediately
- Process runs in foreground

**Use for:** Simple servers that don't fork

**`Type=forking`**
```bash
[Service]
Type=forking
PIDFile=/run/myapp.pid
ExecStart=/usr/sbin/nginx
```
- Process forks into background
- systemd waits for parent to exit
- Tracks child via PID file

**Use for:** Traditional daemons (nginx, Apache)

**`Type=notify`**
```bash
[Service]
Type=notify
ExecStart=/usr/bin/my-server
```
- Process notifies systemd when ready
- Uses `sd_notify()` function

**Use for:** Services that take time to initialize

**Example in Go:**

```go
import "github.com/coreos/go-systemd/daemon"

func main() {
    // Start server
    go http.ListenAndServe(":8080", nil)
    
    // Tell systemd we're ready
    daemon.SdNotify(false, daemon.SdNotifyReady)
}
```

**`Type=oneshot`**
```bash
[Service]
Type=oneshot
ExecStart=/usr/local/bin/setup-script.sh
RemainAfterExit=yes
```
- Runs once and exits
- systemd waits for completion
- `RemainAfterExit=yes` keeps unit "active" after exit

**Use for:** One-time setup tasks

---

## Creating a Custom Service

### Example: Node.js API Server

**1. Create the service file**

```bash
$ sudo nano /etc/systemd/system/api-server.service
```

```ini
[Unit]
Description=API Server (Node.js)
After=network.target
Wants=postgresql.service

[Service]
Type=simple
User=apiuser
Group=apiuser
WorkingDirectory=/opt/api-server
Environment="NODE_ENV=production"
Environment="PORT=3000"
ExecStart=/usr/bin/node server.js
Restart=on-failure
RestartSec=10s

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=api-server

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/api-server/logs

[Install]
WantedBy=multi-user.target
```

**2. Reload systemd (to see new file)**

```bash
$ sudo systemctl daemon-reload
```

**3. Enable (start on boot)**

```bash
$ sudo systemctl enable api-server
Created symlink /etc/systemd/system/multi-user.target.wants/api-server.service → /etc/systemd/system/api-server.service.
```

**4. Start the service**

```bash
$ sudo systemctl start api-server
```

**5. Check status**

```bash
$ systemctl status api-server
● api-server.service - API Server (Node.js)
     Loaded: loaded (/etc/systemd/system/api-server.service; enabled; vendor preset: enabled)
     Active: active (running) since Tue 2024-02-20 10:15:23 UTC; 5min ago
   Main PID: 12345 (node)
      Tasks: 11 (limit: 4915)
     Memory: 45.2M
        CPU: 1.234s
     CGroup: /system.slice/api-server.service
             └─12345 /usr/bin/node server.js

Feb 20 10:15:23 server systemd[1]: Started API Server (Node.js).
Feb 20 10:15:24 server api-server[12345]: Server listening on port 3000
```

---

## Common Commands

**Start/Stop/Restart:**

```bash
$ sudo systemctl start nginx       # Start service
$ sudo systemctl stop nginx        # Stop service
$ sudo systemctl restart nginx     # Stop then start
$ sudo systemctl reload nginx      # Reload config (if supported)
$ sudo systemctl try-restart nginx # Restart only if already running
```

**Enable/Disable (boot time):**

```bash
$ sudo systemctl enable nginx      # Start on boot
$ sudo systemctl disable nginx     # Don't start on boot
$ sudo systemctl enable --now nginx # Enable + start now
```

**Status:**

```bash
$ systemctl status nginx           # Detailed status
$ systemctl is-active nginx        # Active or inactive
$ systemctl is-enabled nginx       # Enabled or disabled
$ systemctl is-failed nginx        # Failed or not
```

**List:**

```bash
$ systemctl list-units --type=service         # All services
$ systemctl list-units --type=service --state=running  # Running services
$ systemctl list-units --type=service --state=failed   # Failed services
$ systemctl list-unit-files --type=service    # All service files
```

**Dependencies:**

```bash
$ systemctl list-dependencies nginx   # What nginx depends on
$ systemctl list-dependencies --reverse nginx  # What depends on nginx
```

---

## journald: systemd's Logging System

**All logs go to journald** (instead of `/var/log/` text files).

**View logs:**

```bash
# All logs
$ journalctl

# Since last boot
$ journalctl -b

# Last 100 lines
$ journalctl -n 100

# Follow (like tail -f)
$ journalctl -f
```

**Filter by service:**

```bash
$ journalctl -u nginx             # nginx logs
$ journalctl -u nginx -u docker   # nginx + docker logs
$ journalctl -u nginx --since today
$ journalctl -u nginx --since "2024-02-20 10:00:00"
$ journalctl -u nginx --since "1 hour ago"
```

**Filter by priority:**

```bash
$ journalctl -p err               # Errors only
$ journalctl -p warning           # Warnings and above
```

Priorities: `emerg`, `alert`, `crit`, `err`, `warning`, `notice`, `info`, `debug`

**Filter by time:**

```bash
$ journalctl --since "2024-02-20 10:00:00" --until "2024-02-20 11:00:00"
$ journalctl --since yesterday --until now
```

**Output format:**

```bash
$ journalctl -u nginx -o json-pretty   # JSON
$ journalctl -u nginx -o cat           # Just messages (no metadata)
$ journalctl -u nginx -o short-iso     # ISO timestamps
```

**Follow multiple services:**

```bash
$ journalctl -u nginx -u api-server -f
```

**Disk usage:**

```bash
$ journalctl --disk-usage
Archived and active journals take up 500M in the file system.

# Vacuum old logs
$ sudo journalctl --vacuum-time=7d      # Keep last 7 days
$ sudo journalctl --vacuum-size=500M    # Keep last 500MB
```

---

## systemd Timers (Better than Cron)

**Timers** are systemd's replacement for cron jobs.

### Example: Backup Timer

**1. Create service (what to run):**

```bash
$ sudo nano /etc/systemd/system/backup.service
```

```ini
[Unit]
Description=Daily Backup

[Service]
Type=oneshot
ExecStart=/usr/local/bin/backup.sh
User=backup
```

**2. Create timer (when to run it):**

```bash
$ sudo nano /etc/systemd/system/backup.timer
```

```ini
[Unit]
Description=Daily Backup Timer

[Timer]
OnCalendar=daily
OnCalendar=*-*-* 02:00:00
Persistent=true

[Install]
WantedBy=timers.target
```

**3. Enable and start timer:**

```bash
$ sudo systemctl daemon-reload
$ sudo systemctl enable backup.timer
$ sudo systemctl start backup.timer
```

**4. Check timer status:**

```bash
$ systemctl list-timers
NEXT                         LEFT        LAST                         PASSED  UNIT           ACTIVATES
Wed 2024-02-21 02:00:00 UTC  8h left     Tue 2024-02-20 02:00:00 UTC  16h ago backup.timer   backup.service

$ systemctl status backup.timer
● backup.timer - Daily Backup Timer
     Loaded: loaded (/etc/systemd/system/backup.timer; enabled)
     Active: active (waiting) since Tue 2024-02-20 10:00:00 UTC; 8h ago
    Trigger: Wed 2024-02-21 02:00:00 UTC; 8h left
```

### Timer Options

**`OnCalendar=`** (like cron)

```ini
OnCalendar=daily              # Every day at midnight (00:00)
OnCalendar=*-*-* 02:00:00     # Every day at 2am
OnCalendar=Mon *-*-* 00:00:00 # Monday at midnight
OnCalendar=*-*-01 00:00:00    # 1st of month at midnight
OnCalendar=hourly             # Every hour
OnCalendar=*:0/15             # Every 15 minutes
```

**`OnBootSec=`** (after boot)

```ini
OnBootSec=5min                # 5 minutes after boot
```

**`OnUnitActiveSec=`** (after last activation)

```ini
OnUnitActiveSec=1h            # 1 hour after last run
```

**Advantages over cron:**
- Better logging (`journalctl -u backup.service`)
- Dependency management (run after network is up)
- `Persistent=true` runs missed jobs after reboot
- Integration with systemd (status, dependencies, etc.)

---

## Socket Activation

**systemd can start services on-demand when someone connects.**

### Example: SSH

```bash
$ systemctl cat sshd.socket
[Unit]
Description=OpenSSH Server Socket

[Socket]
ListenStream=22
Accept=no

[Install]
WantedBy=sockets.target
```

**How it works:**

```
1. systemd listens on port 22
2. Client connects
3. systemd starts sshd.service
4. sshd handles connection
5. (Optional) sshd stops after idle timeout
```

**Benefits:**
- Services don't need to run until first connection
- Saves memory
- Faster boot (services start in parallel)

**Check socket status:**

```bash
$ systemctl list-sockets
LISTEN                    UNIT                   ACTIVATES
/run/systemd/journal/socket systemd-journald.socket systemd-journald.service
[::]:22                   sshd.socket            sshd.service
```

---

## systemd Targets (Runlevels)

**Targets** group units together (like "boot to GUI" vs "boot to terminal").

**Common targets:**

| Target | Old Runlevel | Purpose |
|--------|--------------|---------|
| `poweroff.target` | 0 | Shut down |
| `rescue.target` | 1 | Single-user mode |
| `multi-user.target` | 3 | Multi-user, no GUI |
| `graphical.target` | 5 | Multi-user with GUI |
| `reboot.target` | 6 | Reboot |

**Check current target:**

```bash
$ systemctl get-default
graphical.target
```

**Change default target:**

```bash
$ sudo systemctl set-default multi-user.target
```

**Boot to different target once:**

```bash
$ sudo systemctl isolate rescue.target
```

---

## Resource Limits with systemd

**Limit what a service can use** (uses cgroups under the hood).

```ini
[Service]
# Memory
MemoryMax=512M              # Hard limit (OOMKill if exceeded)
MemoryHigh=400M             # Soft limit (throttle if exceeded)

# CPU
CPUQuota=50%                # Max 50% of one CPU core

# Tasks (processes/threads)
TasksMax=100                # Max 100 tasks

# Files
LimitNOFILE=65536           # Max open files

# Time
RuntimeMaxSec=3600          # Kill after 1 hour
```

**Example:**

```bash
$ sudo nano /etc/systemd/system/limited-service.service
```

```ini
[Unit]
Description=Resource-Limited Service

[Service]
Type=simple
ExecStart=/usr/bin/my-server
MemoryMax=512M
CPUQuota=50%
TasksMax=50
LimitNOFILE=1024
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

**View actual resource usage:**

```bash
$ systemctl status limited-service
     Memory: 234.5M (max: 512.0M)
        CPU: 450ms
      Tasks: 12 (limit: 50)
```

---

## Real-World Example: Production Web Service

```ini
[Unit]
Description=Production Web API
Documentation=https://github.com/mycompany/api
After=network-online.target postgresql.service
Wants=network-online.target
Requires=postgresql.service

[Service]
Type=notify
User=apiuser
Group=apiuser
WorkingDirectory=/opt/api

# Environment
Environment="NODE_ENV=production"
Environment="PORT=3000"
EnvironmentFile=/etc/api/environment

# Execution
ExecStartPre=/usr/bin/npm ci --only=production
ExecStart=/usr/bin/node --max-old-space-size=1024 server.js
ExecReload=/bin/kill -s HUP $MAINPID

# Restart policy
Restart=on-failure
RestartSec=10s
StartLimitInterval=5min
StartLimitBurst=5

# Resource limits
MemoryMax=2G
CPUQuota=200%
TasksMax=256
LimitNOFILE=65536

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=api

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/api/logs /opt/api/uploads
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
```

**Key features:**
- **Dependencies:** Requires PostgreSQL, waits for network
- **Resource limits:** 2GB RAM, 2 CPU cores max
- **Restart policy:** Auto-restart on failure, max 5 times in 5 minutes
- **Security:** Restricted filesystem access, no new privileges
- **Logging:** All output to journal

---

## Debugging systemd Services

**Service won't start:**

```bash
# Check status
$ systemctl status myservice

# Check logs
$ journalctl -u myservice -n 50

# Check if service file has errors
$ systemd-analyze verify /etc/systemd/system/myservice.service
```

**Service starts but crashes immediately:**

```bash
# View recent logs
$ journalctl -u myservice --since "1 minute ago"

# Check exit code
$ systemctl show myservice -p ExecMainStatus
ExecMainStatus=1  # Non-zero = error
```

**Check why service is slow to start:**

```bash
$ systemd-analyze blame
         5.234s postgresql.service
         2.123s nginx.service
         1.456s docker.service
```

**Visualize boot process:**

```bash
$ systemd-analyze plot > boot.svg
$ firefox boot.svg
```

---

## Key Takeaways

1. **systemd is PID 1 — it starts everything and manages services**
2. **Units are configuration files — `.service` for services, `.timer` for scheduled tasks**
3. **`systemctl` manages services — start, stop, enable, status**
4. **`journalctl` views logs — all logs in one place, not scattered in `/var/log/`**
5. **Timers replace cron — better logging and dependencies**
6. **Socket activation starts services on-demand**
7. **Resource limits use cgroups — `MemoryMax=`, `CPUQuota=`**

---

## What's Next

- [Module 09: Boot Process](../09-boot/) — What happens from power-on to login
- [Module 11: Performance & Debugging](../11-debugging/) — systemd debugging in depth

---

**Next:** [Module 09: The Boot Process](../09-boot/01-boot-process.md)
