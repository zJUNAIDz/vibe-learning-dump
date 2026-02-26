# When Root Isn't Enough

## What Problem This Solves

You ran `sudo` but still can't do something. Or you're root but a file won't delete. Or your process has root but SELinux blocks it. This file covers all the access control layers on Fedora and what each one can block.

## The Layers of Access Control

```
┌───────────────────────────────────────────┐
│ Layer 5: firewalld / nftables             │  ← Network access
├───────────────────────────────────────────┤
│ Layer 4: SELinux (MAC)                    │  ← Process/file labels
├───────────────────────────────────────────┤
│ Layer 3: Capabilities                     │  ← Fine-grained root powers
├───────────────────────────────────────────┤
│ Layer 2: DAC (Unix permissions + ACLs)    │  ← rwx, user/group
├───────────────────────────────────────────┤
│ Layer 1: Filesystem flags (immutable)     │  ← chattr, mount flags
└───────────────────────────────────────────┘
```

Each layer is checked independently. Being root bypasses Layer 2 but NOT layers 1, 3, 4, or 5.

## Layer 1: Filesystem Attributes

### Immutable Files

```bash
# Make a file immutable — even root can't modify or delete it:
sudo chattr +i /etc/resolv.conf

# Now try:
sudo rm /etc/resolv.conf         # Operation not permitted!
sudo echo "x" >> /etc/resolv.conf  # Permission denied!

# Check attributes:
lsattr /etc/resolv.conf
# ----i--------e-- /etc/resolv.conf

# Remove the immutable flag:
sudo chattr -i /etc/resolv.conf
```

**When you encounter this:** Something marked a file immutable to prevent modification (common for DNS config, system files). `lsattr` reveals it.

### Append-Only

```bash
# Can only add data, not modify existing data:
sudo chattr +a /var/log/security.log
echo "new entry" >> /var/log/security.log    # Works
echo "overwrite" > /var/log/security.log     # Fails!
```

### Read-Only Mount

```bash
# If a filesystem is mounted read-only:
mount | grep "on / "
# /dev/sda1 on / type ext4 (ro,relatime)  ← READ-ONLY!

# Remount read-write:
sudo mount -o remount,rw /

# Or the filesystem might be in an error state and remounted ro automatically
```

## Layer 2: Unix Permissions and ACLs

### Beyond rwx: ACLs

Regular permissions (user/group/other) are sometimes not enough. ACLs provide fine-grained control:

```bash
# View ACLs:
getfacl /path/to/file

# Grant read access to a specific user:
setfacl -m u:alice:r /path/to/file

# Grant access to a specific group:
setfacl -m g:devteam:rw /path/to/file

# Default ACLs (apply to new files in a directory):
setfacl -d -m g:devteam:rw /shared/project/

# Remove all ACLs:
setfacl -b /path/to/file

# A + in ls -l indicates ACLs exist:
ls -l /path/to/file
# -rw-r--r--+ 1 user user 100 Jan 15 10:00 file
#           ^ ACL present
```

### Sticky Bit, SUID, SGID

```bash
# Sticky bit (t) — on directories, prevents deletion by non-owners:
ls -ld /tmp
# drwxrwxrwt — the 't' means sticky bit
# Anyone can create files, only owners can delete their own

# SUID (s) — runs with the FILE OWNER's privileges:
ls -l /usr/bin/passwd
# -rwsr-xr-x — the 's' means SUID
# passwd runs as root (to modify /etc/shadow)

# SGID (s) on directories — new files inherit the group:
ls -ld /shared/project
# drwxrwsr-x — new files get this directory's group
```

## Layer 3: Linux Capabilities

Root is actually a collection of ~40 separate capabilities. You can grant specific capabilities to programs without giving full root:

```bash
# View capabilities of a binary:
getcap /usr/bin/ping
# /usr/bin/ping cap_net_raw=ep

# This means ping can create raw network sockets without being root or SUID.

# Grant a capability:
sudo setcap 'cap_net_bind_service=ep' /opt/myapp/server
# Now the server can bind to ports < 1024 without root

# Remove capabilities:
sudo setcap -r /opt/myapp/server

# View capabilities of a running process:
cat /proc/$(pgrep nginx)/status | grep Cap
# CapPrm: 00000000a80435fb
# CapEff: 00000000a80435fb

# Decode capability hex:
capsh --decode=00000000a80435fb
```

### Common Capabilities

```bash
cap_net_bind_service   # Bind to ports < 1024
cap_net_raw            # Raw sockets (ping, tcpdump)
cap_sys_admin          # Various admin operations
cap_dac_override       # Bypass file permission checks
cap_chown              # Change file ownership
cap_setuid             # Change UID
cap_sys_ptrace         # Trace/debug other processes
```

### Capabilities in systemd

```bash
# In a unit file — grant only specific capabilities:
[Service]
User=myapp
AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
# Now the service runs as non-root but can bind to port 80
```

## Layer 4: SELinux

(Covered in detail in the SELinux file. Quick summary here.)

```bash
# Root doesn't bypass SELinux:
sudo -u root cat /some/file     # CAN be denied by SELinux

# Check:
sudo ausearch -m avc -ts recent | head -20

# Temporarily disable for testing:
sudo setenforce 0
# (Test if the problem goes away)
sudo setenforce 1
```

## Layer 5: Firewall

### firewalld (Fedora's Default)

```bash
# Check status:
sudo firewall-cmd --state

# List current rules:
sudo firewall-cmd --list-all

# Open a port:
sudo firewall-cmd --add-port=8080/tcp
sudo firewall-cmd --add-port=8080/tcp --permanent  # Survives reboot

# Open a service:
sudo firewall-cmd --add-service=http --permanent

# Reload after permanent changes:
sudo firewall-cmd --reload

# List available services:
sudo firewall-cmd --get-services

# Zones (Fedora uses zones for network interfaces):
sudo firewall-cmd --get-active-zones
sudo firewall-cmd --zone=public --list-all

# Rich rules (complex rules):
sudo firewall-cmd --add-rich-rule='rule family="ipv4" source address="192.168.1.0/24" port protocol="tcp" port="3306" accept' --permanent
```

### Quick Diagnostics

```bash
# "Why can't I reach my service?"
# 1. Is the service running?
systemctl is-active myapp

# 2. Is it listening?
ss -tlnp | grep 8080

# 3. Is the firewall blocking it?
sudo firewall-cmd --list-all | grep 8080

# 4. Is SELinux blocking the port?
sudo semanage port -l | grep 8080

# 5. Can you reach it locally?
curl http://localhost:8080

# 6. Can you reach it from another machine?
curl http://server:8080
```

## The Troubleshooting Hierarchy

When something is denied despite being root:

```bash
# 1. Check filesystem attributes:
lsattr /path/to/file

# 2. Check permissions + ACLs:
ls -la /path/to/file
getfacl /path/to/file

# 3. Check SELinux:
getenforce
sudo ausearch -m avc -c mycommand -ts recent

# 4. Check capabilities (if running as non-root):
getcap /path/to/binary

# 5. Check firewall (for network issues):
sudo firewall-cmd --list-all

# 6. Check mount flags:
mount | grep "$(df /path/to/file | tail -1 | awk '{print $1}')"

# 7. Still stuck? strace to see the actual syscall failure:
sudo strace -f -e trace=open,read,write,connect command 2>&1 | grep -i denied
```

## Least Privilege Patterns

### Run as Non-Root with Capabilities

```bash
# Instead of running your app as root:
[Service]
User=myapp
Group=myapp
AmbientCapabilities=CAP_NET_BIND_SERVICE
NoNewPrivileges=true
ProtectSystem=full
ProtectHome=true
ReadWritePaths=/var/lib/myapp
```

### Dedicated Service Users

```bash
# Create a system user with no login:
sudo useradd -r -s /usr/sbin/nologin -d /opt/myapp myapp

# Set ownership:
sudo chown -R myapp:myapp /opt/myapp /var/lib/myapp /var/log/myapp
```

### systemd Sandboxing

```bash
# systemd can restrict what a service can do:
[Service]
ProtectSystem=strict       # Mount / as read-only
ProtectHome=true           # Hide /home
PrivateTmp=true            # Private /tmp
NoNewPrivileges=true       # Can't gain new privileges
ProtectKernelModules=true  # Can't load kernel modules
ProtectKernelTunables=true # Can't modify /proc/sys
RestrictNamespaces=true    # Can't create namespaces
ReadWritePaths=/var/lib/myapp  # Whitelist writable paths

# Check what restrictions a service has:
systemd-analyze security myapp
# Gives a score and lists all security features
```

## Exercise

1. Make a file immutable with `chattr +i`. Try to modify it as root. Remove the flag and verify.

2. Use `getcap` to check which binaries on your system have Linux capabilities. Focus on `/usr/bin/ping`.

3. Run `systemd-analyze security sshd` and read the output. What's the security exposure score? What could be tightened?

4. Create a test service that runs as a non-root user but can bind to port 80 using capabilities. Verify it works.

---

Next: [Level 10 — Trustworthy Scripts](../10-trustworthy-scripts/00-production-scripts.md)
