# Users, Permissions & Security

**Understanding Linux Access Control**

🟢 **Beginner-Friendly** | 🟡 **Intermediate**

---

## Introduction

Every file, process, and resource in Linux has an **owner** and **permissions**.

Understanding this is critical for:
- Running services securely
- Debugging "permission denied" errors
- Containerized applications
- DevOps workflows

---

## Users and Groups

### Users

Every process runs as a **user**.

```bash
$ whoami
zjunaidz

$ id
uid=1000(zjunaidz) gid=1000(zjunaidz) groups=1000(zjunaidz),10(wheel),998(docker)
```

**Two types of users:**

1. **Regular users** — UID ≥ 1000
   - Created with `useradd` or `adduser`
   - Can log in
   - Have home directory in `/home/`

2. **System users** — UID < 1000
   - For services (nginx, postgres, etc.)
   - Usually can't log in
   - No password

**Special user:**
- **root** — UID 0, superuser, can do anything

### Groups

Users belong to **groups** for shared access.

```bash
$ groups
zjunaidz wheel docker

$ id -nG
zjunaidz wheel docker
```

**Common groups:**

| Group | Purpose |
|-------|---------|
| `wheel` | Can use `sudo` (Fedora/RHEL) |
| `sudo` | Can use `sudo` (Debian/Ubuntu) |
| `docker` | Can run Docker commands |
| `audio`, `video` | Access audio/video devices |
| `systemd-journal` | Read system logs |

### User Database

**`/etc/passwd`** — User information

```bash
$ cat /etc/passwd | grep zjunaidz
zjunaidz:x:1000:1000:Zjunaidz User:/home/zjunaidz:/bin/bash
#        ^ ^    ^    ^           ^              ^
#        | |    |    |           |              Login shell
#        | |    |    |           Home directory
#        | |    |    Description
#        | |    Primary group (GID)
#        | User ID (UID)
#        Password (x = in /etc/shadow)
```

**`/etc/shadow`** — Password hashes (only root can read)

```bash
$ sudo cat /etc/shadow | grep zjunaidz
zjunaidz:$6$rounds=5000$...:19400:0:99999:7:::
#        ^^^^^^^^^^^^^^^^^^^^^
#        Hashed password
```

**`/etc/group`** — Group information

```bash
$ cat /etc/group | grep docker
docker:x:998:zjunaidz
#      ^ ^   ^
#      | |   Members
#      | GID
#      Password (rarely used)
```

---

## File Permissions

Every file has:
- **Owner** (user)
- **Group**
- **Permissions** (read, write, execute)

```bash
$ ls -l /etc/passwd
-rw-r--r-- 1 root root 2847 Feb 20 10:00 /etc/passwd
│││││││││  │ │    │    │
│││││││││  │ │    │    Size
│││││││││  │ │    Group
│││││││││  │ Owner
│││││││││  Links
│││││││││
││││││││└─ Others permissions: r-- (read only)
│││││└──── Group permissions:  r-- (read only)
││││└───── Owner permissions:  rw- (read, write)
│││└────── Special bits
││└─────── Type: - (regular file)
│└──────── (ignored)
└───────── (ignored)
```

### Permission Bits

| Bit | Letter | Numeric | File | Directory |
|-----|--------|---------|------|-----------|
| Read | `r` | 4 | Read contents | List files (`ls`) |
| Write | `w` | 2 | Modify file | Create/delete files |
| Execute | `x` | 1 | Run as program | Enter directory (`cd`) |

**Numeric representation:**

```
rwx = 4 + 2 + 1 = 7
rw- = 4 + 2 + 0 = 6
r-x = 4 + 0 + 1 = 5
r-- = 4 + 0 + 0 = 4
```

**Example:**

```bash
$ ls -l script.sh
-rwxr-xr-x 1 zjunaidz zjunaidz 1234 Feb 20 10:00 script.sh
 ^^^ ^^^ ^^^
 755

Owner: rwx (7) = read, write, execute
Group: r-x (5) = read, execute
Others: r-x (5) = read, execute
```

### Changing Permissions

**`chmod` (change mode):**

```bash
# Symbolic mode
$ chmod u+x script.sh     # Owner: add execute
$ chmod g-w file.txt      # Group: remove write
$ chmod o=r file.txt      # Others: set to read-only
$ chmod a+x script.sh     # All: add execute

# Numeric mode
$ chmod 755 script.sh     # rwxr-xr-x
$ chmod 644 file.txt      # rw-r--r--
$ chmod 600 private.key   # rw-------
$ chmod 700 ~/bin         # rwx------
```

**Common patterns:**

```bash
$ chmod 755 executable    # Script/binary
$ chmod 644 file.txt      # Regular file (read by all)
$ chmod 600 ~/.ssh/id_rsa # Private key (owner only)
$ chmod 666 /dev/null     # World-writable device
```

### Changing Ownership

**`chown` (change owner):**

```bash
$ sudo chown zjunaidz file.txt           # Change owner
$ sudo chown zjunaidz:developers file.txt # Change owner and group
$ sudo chown :developers file.txt        # Change group only

# Recursive
$ sudo chown -R www-data:www-data /var/www/html
```

**`chgrp` (change group):**

```bash
$ sudo chgrp developers project/
$ sudo chgrp -R docker /var/lib/docker
```

---

## Special Permission Bits

### Setuid (Set User ID) — `u+s`

**When executable has setuid bit, it runs as the file owner (not the user who runs it).**

**Example: `passwd` command**

```bash
$ ls -l /usr/bin/passwd
-rwsr-xr-x 1 root root 68208 ... /usr/bin/passwd
   ^
   s = setuid bit
```

- Owned by `root`
- Has setuid bit
- Any user can run it
- **Runs as root** (can modify `/etc/shadow`)

**Set setuid:**

```bash
$ chmod u+s program
$ chmod 4755 program  # 4 = setuid
```

**⚠️ Security risk if misused!**

### Setgid (Set Group ID) — `g+s`

**On file:** Runs with file's group instead of user's group

**On directory:** New files inherit directory's group (not user's primary group)

```bash
$ mkdir shared
$ sudo chgrp developers shared
$ chmod g+s shared

$ ls -ld shared
drwxrwsr-x 2 root developers 4096 ... shared
      ^
      s = setgid bit

# Files created in shared/ automatically get "developers" group
$ touch shared/newfile
$ ls -l shared/newfile
-rw-r--r-- 1 zjunaidz developers 0 ... newfile
#                    ^^^^^^^^^^
#                    Inherited from directory
```

### Sticky Bit — `o+t`

**On directory:** Users can only delete files they own (even if directory is world-writable).

**Example: `/tmp`**

```bash
$ ls -ld /tmp
drwxrwxrwt 20 root root 4096 ... /tmp
         ^
         t = sticky bit

# Anyone can create files in /tmp
# But you can only delete your own files
```

**Set sticky bit:**

```bash
$ chmod +t directory
$ chmod 1777 directory  # 1 = sticky bit
```

---

## sudo: Temporary Root Access

**`sudo`** — Run commands as another user (usually root).

### Configuration

**`/etc/sudoers`** (edit with `visudo` only!)

```bash
$ sudo visudo
```

**Basic syntax:**

```
user  host=(runas) command
```

**Examples:**

```bash
# Grant wheel group full sudo access
%wheel ALL=(ALL) ALL

# User can run all commands as root
zjunaidz ALL=(ALL) ALL

# User can restart nginx without password
zjunaidz ALL=(ALL) NOPASSWD: /bin/systemctl restart nginx

# User can run any docker command
zjunaidz ALL=(ALL) NOPASSWD: /usr/bin/docker
```

**Check sudo access:**

```bash
$ sudo -l
User zjunaidz may run the following commands on hostname:
    (ALL) ALL
```

### Common sudo Commands

```bash
$ sudo command              # Run as root
$ sudo -u postgres psql     # Run as "postgres" user
$ sudo -i                   # Login shell as root
$ sudo -s                   # Shell as root (keep environment)
$ sudo !!                   # Run previous command with sudo
```

---

## Capabilities: Fine-Grained Permissions

**Problem:** Programs often need **one root privilege**, but `setuid` gives **all privileges**.

**Solution:** **Capabilities** — break root into 40+ individual privileges.

### Common Capabilities

| Capability | Purpose |
|-----------|---------|
| `CAP_NET_BIND_SERVICE` | Bind to ports < 1024 |
| `CAP_NET_RAW` | Use raw sockets (ping) |
| `CAP_SYS_ADMIN` | Mount filesystems |
| `CAP_SYS_TIME` | Set system time |
| `CAP_KILL` | Send signals to any process |
| `CAP_CHOWN` | Change file ownership |

### Viewing Capabilities

```bash
# Check file capabilities
$ getcap /usr/bin/ping
/usr/bin/ping cap_net_raw=ep

# Check process capabilities
$ getpcaps $$
Capabilities for '12345': cap_chown,cap_dac_override,...
```

### Setting Capabilities

```bash
# Instead of setuid, give specific capability
$ sudo setcap cap_net_bind_service=+ep /usr/bin/node

# Now Node.js can bind to port 80 without root
$ node server.js  # Listens on port 80, runs as regular user
```

**In Docker/Kubernetes:**

```yaml
# Give container only necessary capabilities
securityContext:
  capabilities:
    add:
      - NET_ADMIN
      - SYS_TIME
    drop:
      - ALL  # Drop all others
```

---

## AppArmor / SELinux: Mandatory Access Control

**Beyond user/group permissions:** MAC (Mandatory Access Control)

### SELinux (Security-Enhanced Linux)

**Used in:** Fedora, RHEL, CentOS

**Concept:** Every file and process has a **security context**.

```bash
$ ls -Z /var/www/html/index.html
-rw-r--r--. root root system_u:object_r:httpd_sys_content_t:s0 index.html
#                    ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
#                    SELinux context
```

**Context format:** `user:role:type:level`

**Example policy:** `httpd` (Apache) can only access files with type `httpd_sys_content_t`.

**Check SELinux status:**

```bash
$ getenforce
Enforcing  # Active and blocking violations

$ sestatus
SELinux status:                 enabled
Current mode:                   enforcing
```

**Common SELinux commands:**

```bash
# View context
$ ls -Z file

# Change context
$ chcon -t httpd_sys_content_t /var/www/html/newfile

# Fix contexts automatically
$ restorecon -R /var/www/html

# Temporarily disable (for debugging)
$ sudo setenforce 0  # Permissive mode

# Check denials
$ sudo ausearch -m avc -ts recent
```

### AppArmor

**Used in:** Ubuntu, Debian, openSUSE

**Concept:** Profiles define what programs can do.

```bash
# Check status
$ sudo aa-status

# Example profile locations
$ ls /etc/apparmor.d/
usr.bin.firefox
usr.sbin.nginx
```

**Disable profile (for debugging):**

```bash
$ sudo aa-complain /usr/sbin/nginx  # Complain mode (log but don't block)
$ sudo aa-enforce /usr/sbin/nginx   # Enforce mode
```

---

## User Namespaces (Containers)

**Containers remap UIDs** so root inside container isn't root on host.

```bash
# On host
$ ps aux | grep nginx
root     1234  ...  nginx

# In container
$ docker exec -it container ps aux
USER   PID  COMMAND
root     1  nginx
#       ^^^ PID 1 in container
#           But PID 1234 on host
```

**User namespace mapping:**

```
Container UID 0 (root) → Host UID 231072
Container UID 1        → Host UID 231073
...
```

**Check on host:**

```bash
$ ps aux | grep nginx
231072   1234  ...  nginx
# Root in container = UID 231072 on host (unprivileged!)
```

---

## Production Security Best Practices

### 1. Run Services as Non-Root

**Bad:**

```dockerfile
FROM node:18
COPY . /app
CMD ["node", "server.js"]
# Runs as root (UID 0) in container!
```

**Good:**

```dockerfile
FROM node:18
RUN useradd -m -u 1001 appuser
USER appuser
COPY --chown=appuser:appuser . /app
CMD ["node", "server.js"]
# Runs as UID 1001
```

### 2. Minimal Permissions

```bash
# Application files: read-only for everyone except owner
$ chmod 644 *.js *.json

# Secrets: owner read-only
$ chmod 600 .env secrets.yaml

# Scripts: executable for owner
$ chmod 700 scripts/*.sh

# Logs directory: writable by app
$ chown appuser:appuser /var/log/myapp
$ chmod 755 /var/log/myapp
```

### 3. Drop Capabilities in Containers

```yaml
# docker-compose.yml
services:
  app:
    image: myapp
    cap_drop:
      - ALL
    cap_add:
      - NET_BIND_SERVICE  # Only if binding to port < 1024
    security_opt:
      - no-new-privileges:true
```

### 4. Read-Only Root Filesystem

```yaml
# docker-compose.yml
services:
  app:
    image: myapp
    read_only: true
    tmpfs:
      - /tmp  # If app needs to write temp files
```

### 5. Use Security Scanning

```bash
# Scan Docker images
$ docker scan myapp:latest

# Check for vulnerabilities
$ trivy image myapp:latest
```

---

## Common Permission Errors

### "Permission denied" writing to file

```bash
$ echo "data" > /var/log/app.log
bash: /var/log/app.log: Permission denied

# Check permissions
$ ls -l /var/log/app.log
-rw-r--r-- 1 root root 0 ... app.log

# Fix: Make writable by app user
$ sudo chown appuser:appuser /var/log/app.log
$ sudo chmod 644 /var/log/app.log
```

### "Permission denied" executing script

```bash
$ ./script.sh
bash: ./script.sh: Permission denied

# Add execute permission
$ chmod +x script.sh
$ ./script.sh  # Works now
```

### "Cannot bind to port 80"

```bash
$ node server.js
Error: bind EACCES 0.0.0.0:80

# Ports < 1024 require root or CAP_NET_BIND_SERVICE
# Option 1: Use higher port
$ PORT=8080 node server.js

# Option 2: Give capability
$ sudo setcap cap_net_bind_service=+ep $(which node)

# Option 3: Run as root (not recommended)
$ sudo node server.js
```

---

## Key Takeaways

1. **Every process runs as a user — check with `ps aux` or `id`**
2. **File permissions: owner, group, others — use `ls -l` and `chmod`**
3. **sudo grants temporary root access — configured in `/etc/sudoers`**
4. **Capabilities split root privileges into granular permissions**
5. **Run production services as non-root users for security**
6. **Containers can isolate users with user namespaces**
7. **SELinux/AppArmor provide additional mandatory access control**

---

## What's Next

- [Module 11: Performance & Debugging](../11-debugging/) — Tools for diagnosing issues
- [Module 12: Production Concerns](../12-production/) — Kernel limits, file descriptor exhaustion

---

**Next:** [Module 11: Performance & Debugging](../11-debugging/01-debugging-tools.md)
