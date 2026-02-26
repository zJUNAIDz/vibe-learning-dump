# How sudo Actually Works

## What Problem This Solves

People type `sudo` before commands like a magic "please work" prefix. When `sudo` doesn't fix their problem, they're stuck. Understanding what `sudo` actually does — and what it doesn't do — makes permission issues debuggable.

## How People Misunderstand It

1. **"sudo = run as root"** — `sudo` can run as ANY user, not just root. `sudo -u postgres psql` runs as the postgres user.
2. **"sudo gives the whole command root power"** — `sudo` only elevates the *command* it runs, not the shell. Redirections (`>`, `|`) are still processed by YOUR shell.
3. **"If sudo doesn't work, nothing will"** — SELinux, capabilities, mount options, and other mechanisms can block even root.

## The Mental Model

`sudo` is a **setuid program** (`/usr/bin/sudo`). When you run it:

```
1. sudo reads /etc/sudoers to check if YOU are allowed to run this command
2. It authenticates you (password prompt) unless recently authenticated
3. It forks a new process with the target user's UID/GID (default: root)
4. It execs the target command in that new process
5. Your shell waits for it to finish
```

The critical insight: **sudo runs a NEW process as root. Your shell stays as you.**

```
Your Shell (uid=1000)
  └── sudo (setuid, runs as root momentarily)
       └── the-command (runs as root, uid=0)
```

### Why Redirections Bypass sudo

```bash
sudo echo "nameserver 8.8.8.8" > /etc/resolv.conf
```

The shell parses this BEFORE executing sudo:

```
Step 1: Shell sees ">"  →  Opens /etc/resolv.conf for writing AS YOU (uid=1000)
Step 2: Shell runs: sudo echo "nameserver 8.8.8.8"  →  stdout goes to the file
Step 3: Step 1 fails because YOU can't write to /etc/resolv.conf
```

Your shell handles `>`, not sudo. Sudo only runs `echo`.

**Solutions:**
```bash
# Use tee (tee runs as root, receives from pipe):
echo "nameserver 8.8.8.8" | sudo tee /etc/resolv.conf

# Use tee -a to append:
echo "nameserver 8.8.8.8" | sudo tee -a /etc/resolv.conf

# Run the entire thing in a root shell:
sudo bash -c 'echo "nameserver 8.8.8.8" > /etc/resolv.conf'
```

## Real Fedora Examples

### The sudoers File

```bash
# View sudoers safely (NEVER edit with a regular editor):
sudo visudo

# Default Fedora entry for the wheel group:
# %wheel  ALL=(ALL)  ALL
# Meaning: users in group "wheel" can run any command as any user on any host

# Check if you're in wheel:
groups
# user wheel ...

# Check what you're allowed to do:
sudo -l
```

### sudo Environment

```bash
# sudo resets most environment variables for security:
echo $PATH           # /home/user/.local/bin:/usr/bin:...
sudo env | grep PATH # PATH=/sbin:/bin:/usr/sbin:/usr/bin

# This is controlled by secure_path in /etc/sudoers:
# Defaults    secure_path = /sbin:/bin:/usr/sbin:/usr/bin

# To pass your PATH through sudo:
sudo env "PATH=$PATH" mycommand

# To preserve specific env vars:
sudo --preserve-env=HOME,EDITOR mycommand

# To get a root shell with root's full environment:
sudo -i          # Login shell as root
sudo -s          # Shell as root, keeps your env
sudo su -        # Old-school, less preferred
```

### sudo vs su

```bash
# sudo: run ONE command as root, authenticated with YOUR password
sudo systemctl restart httpd

# su: switch to another user entirely, authenticated with THEIR password
su - root         # Become root (needs root's password — often disabled)
su - postgres     # Become postgres user

# su - (with dash): login shell, fresh environment
# su (no dash): just switch user, keep current environment

# Modern practice: use sudo, not su
# Fedora disables root password by default — su to root doesn't work
# sudo uses YOUR password and logs who did what
```

### sudo and systemd

```bash
# systemctl commands that need root:
sudo systemctl restart httpd     # Works — sudo runs systemctl as root
systemctl status httpd           # Works without sudo — status is read-only

# But for user services, DON'T use sudo:
systemctl --user start myservice   # Runs as YOU, not root
# sudo systemctl --user start myservice  ← WRONG — this runs as root's --user
```

## Common Footguns

**Footgun 1: `sudo` with pipes**
```bash
sudo grep "secret" /etc/shadow | wc -l
# Works fine — sudo runs grep as root, then YOUR shell pipes to wc

cat /etc/shadow | sudo grep "secret"
# Fails — cat runs as YOU and can't read /etc/shadow
# The sudo on grep doesn't help because cat already failed
```

**Footgun 2: `sudo` doesn't carry aliases**
```bash
alias dnf='dnf -y'   # Your alias
sudo dnf install vim  # Runs /usr/bin/dnf, not your alias!

# sudo runs a new process — it doesn't know about your aliases
# Fix: alias sudo='sudo ' (trailing space makes sudo expand the next alias)
alias sudo='sudo '    # Note the space — Bash then checks if the next word is an alias
```

**Footgun 3: File ownership after sudo**
```bash
sudo vim /home/user/config.txt
# vim runs as root → creates swap files as root → leaves root-owned files
# Now you might have root-owned files in your home directory

# Better:
sudoedit /home/user/config.txt    # Copies file, you edit as yourself, copies back
# Or:
sudo -e /home/user/config.txt     # Same thing
```

**Footgun 4: sudo caches credentials (security risk in shared sessions)**
```bash
# After entering your password, sudo doesn't ask again for ~5 minutes (default)
# Anyone with access to your terminal in those 5 minutes can run sudo

# Reset the timer:
sudo -k

# Require password every time (in /etc/sudoers):
# Defaults timestamp_timeout=0
```

## Why This Matters in Real Systems

- **Audit trail**: `sudo` logs every command to `/var/log/secure` (Fedora) with the user's name. Using `su -` to root loses accountability.
- **Principle of least privilege**: `sudoers` can restrict which commands a user can run as root, not just "all or nothing."
- **CI/CD**: In containers and CI, you might be root already (no sudo needed) or not have sudo at all. Scripts that hardcode `sudo` break.
- **Security**: `sudo -i` gives a full root shell — the blast radius of a mistake is maximized. Prefer `sudo specificcommand` when possible.

```bash
# Check sudo audit log:
sudo journalctl _COMM=sudo --since today

# See who ran what:
sudo grep COMMAND /var/log/secure | tail -20
```

## Exercise

1. Run `sudo echo hello > /tmp/test.txt` — does it work? Now run `echo hello | sudo tee /tmp/test.txt`. Compare.
2. Run `env | wc -l` and `sudo env | wc -l`. Count how many environment variables sudo strips.
3. Run `sudo -l` to see what commands you're allowed to run.
4. Identify 3 commands you regularly `sudo` that you could run without it.

---

Next: [Level 2: Word Splitting — Why Spaces Break Everything](../02-arguments-quoting-expansion/00-word-splitting.md)
