# SELinux from the Shell

## What Problem This Solves

You deploy an app on Fedora and it can't read its config file even though permissions are 644. Or Nginx can't bind to port 8080 even though nothing else is using it. You Google the error, find "just disable SELinux," and run `setenforce 0`. **Don't do that.** SELinux is your friend once you understand the 5 commands you actually need.

## The Mental Model

SELinux adds a **second layer of access control** on top of regular Unix permissions (rwx).

```
Traditional Access Check (DAC):
  Does this USER have PERMISSION to this FILE?
  user=nginx, file=config.conf, permission=read → Allowed? Yes (644)

SELinux Access Check (MAC):
  Does this CONTEXT have permission to access this CONTEXT for this ACTION?
  process_context=httpd_t, file_context=user_home_t, action=read → DENIED!
```

Even if Unix permissions say "yes," SELinux can say "no."

### Context = Label

Everything on the system has a SELinux **context** (label):

```bash
# See file contexts:
ls -Z /var/www/html/
# system_u:object_r:httpd_sys_content_t:s0 index.html
#                   ^^^^^^^^^^^^^^^^^ This is the type — it's what matters

# See process contexts:
ps auxZ | grep nginx
# system_u:system_r:httpd_t:s0  nginx: master process
#                   ^^^^^^^ This is the type

# Key insight: SELinux checks if process TYPE can access file TYPE
# httpd_t → CAN read httpd_sys_content_t
# httpd_t → CANNOT read user_home_t
```

## The 5 Commands You Need

### 1. getenforce / setenforce — Check and Set Mode

```bash
# Check current mode:
getenforce
# Enforcing  → SELinux is blocking violations
# Permissive → SELinux logs but doesn't block (debugging mode)
# Disabled   → SELinux is off (bad!)

# Temporarily switch to permissive (for debugging):
sudo setenforce 0

# Switch back to enforcing:
sudo setenforce 1

# NEVER leave permissive in production!
# NEVER put SELINUX=disabled in /etc/selinux/config!
```

### 2. ls -Z / ps auxZ — See Contexts

```bash
# File contexts:
ls -Z /var/www/html/
ls -Z /home/user/myapp/
ls -Zd /var/log/nginx/

# Process contexts:
ps auxZ | grep nginx
ps auxZ | grep sshd

# You're looking for the TYPE field (third part):
# user_u:role_r:TYPE_t:level
```

### 3. ausearch — Find What's Being Blocked

```bash
# See recent SELinux denials:
sudo ausearch -m avc -ts recent

# Denials for a specific command:
sudo ausearch -m avc -c nginx

# Today's denials:
sudo ausearch -m avc -ts today

# The output looks like:
# type=AVC msg=audit(...): avc:  denied  { read } for
#   pid=1234 comm="nginx" name="config.toml"
#   scontext=system_u:system_r:httpd_t:s0
#   tcontext=unconfined_u:object_r:user_home_t:s0
#   tclass=file

# Translation:
# Process (httpd_t) tried to read a file (user_home_t) → DENIED
# The file has the wrong context!
```

### 4. restorecon — Fix File Contexts

```bash
# Restore correct context based on the system's rules:
sudo restorecon -Rv /var/www/html/
# -R: recursive
# -v: verbose (show what changed)

# This is the FIX for 80% of SELinux problems:
# 1. You copied/moved files (contexts don't transfer correctly on move)
# 2. You created files in the wrong location
# 3. Something changed the context accidentally
```

### 5. semanage — Manage Policies

```bash
# Set file context rules (for custom paths):
# If your web content is in /opt/myapp/public instead of /var/www:
sudo semanage fcontext -a -t httpd_sys_content_t "/opt/myapp/public(/.*)?"
sudo restorecon -Rv /opt/myapp/public/

# Allow a non-standard port:
# Nginx wants to use port 8080:
sudo semanage port -a -t http_port_t -p tcp 8080

# List allowed ports for HTTP:
sudo semanage port -l | grep http_port_t

# Allow a boolean (pre-defined permission toggle):
# Let Nginx connect to the network (proxy/reverse proxy):
sudo setsebool -P httpd_can_network_connect on
# -P makes it persistent across reboots

# List all booleans:
getsebool -a | grep httpd
# Or:
sudo semanage boolean -l | grep httpd
```

## The Troubleshooting Workflow

```bash
# Step 1: Is SELinux blocking something?
sudo ausearch -m avc -ts recent

# Step 2: Read the denial — what type is the process? What type is the file?
# scontext=...:httpd_t:...  → Process is an HTTP server
# tcontext=...:user_home_t:... → File is labeled as user home content
# The problem: httpd_t can't read user_home_t

# Step 3: Is it a WRONG file context? (Most common)
ls -Z /path/to/file
# If the type is wrong:
sudo restorecon -Rv /path/to/file

# Step 4: Is the file in a NON-STANDARD location?
# Define the correct context for your custom path:
sudo semanage fcontext -a -t httpd_sys_content_t "/custom/path(/.*)?"
sudo restorecon -Rv /custom/path/

# Step 5: Is it a PERMISSION/PORT issue?
# Check if there's a boolean for it:
sudo ausearch -m avc -ts recent | audit2why
# The output often tells you exactly which boolean to set

# Step 6: Still stuck?
# Generate a custom policy (last resort):
sudo ausearch -m avc -ts recent | audit2allow -M mypolicy
sudo semodule -i mypolicy.pp
```

## sealert — The User-Friendly Version

```bash
# Install:
sudo dnf install setroubleshoot-server

# Check for SELinux alerts:
sudo sealert -a /var/log/audit/audit.log

# This gives you HUMAN-READABLE explanations and SPECIFIC fix commands:
# "SELinux is preventing nginx from read access on the file config.toml"
# "If you want to allow nginx to read config.toml, you can run:
#  sudo restorecon -v /opt/myapp/config.toml"
```

## Common Scenarios on Fedora

### Nginx Serving Custom Content

```bash
# Problem: Nginx can't read files in /opt/mysite/
# Check:
ls -Z /opt/mysite/
# → default_t (wrong!)

# Fix:
sudo semanage fcontext -a -t httpd_sys_content_t "/opt/mysite(/.*)?"
sudo restorecon -Rv /opt/mysite/
```

### Application Can't Connect to Network

```bash
# Problem: Web app can't call an API
sudo ausearch -m avc -c myapp | tail -5
# → denied { name_connect }

# Fix:
sudo setsebool -P httpd_can_network_connect on
```

### Custom Port

```bash
# Problem: App wants port 3000
sudo semanage port -l | grep 3000
# (nothing)

# Fix:
sudo semanage port -a -t http_port_t -p tcp 3000
```

### Home Directory Content

```bash
# Problem: serving files from /home/user/public_html
# Quick fix:
sudo setsebool -P httpd_enable_homedirs on

# Better fix: put content in /var/www or /srv
```

## Common Footguns

### 1. cp vs mv

```bash
# cp CREATES a new file → gets context from the destination directory
sudo cp ~/myfile /var/www/html/    # File gets httpd_sys_content_t ✓

# mv MOVES the inode → KEEPS the original context!
sudo mv ~/myfile /var/www/html/    # File keeps user_home_t ✗

# After mv, always:
sudo restorecon -Rv /var/www/html/
```

### 2. Disabling SELinux "Temporarily"

```bash
# "I'll re-enable it later" → You won't.
# Use permissive mode for debugging, not disabling:
sudo setenforce 0    # For debugging
# ... fix the problem ...
sudo setenforce 1    # Back to enforcing
```

### 3. Not Checking SELinux First

When something "mysteriously" doesn't work on Fedora, always check:
```bash
sudo ausearch -m avc -ts recent
getenforce
```

## Exercise

1. Check the SELinux context of files in your home directory vs `/var/www/html/`. Notice the difference in types.

2. Create a file in `/tmp/`, move it to `/var/www/html/`, and show that the context is wrong. Fix it with `restorecon`.

3. Use `semanage port -l` to find what ports are allowed for `http_port_t`. Add port 9090 and verify.

4. Follow the troubleshooting workflow: create a deliberate SELinux denial (serve a file with wrong context via Nginx), find the AVC denial in `ausearch`, and fix it.

---

Next: [When Root Isn't Enough](02-when-root-isnt-enough.md)
