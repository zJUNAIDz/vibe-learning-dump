# $PATH and Command Resolution

## What Problem This Solves

You type `python` and get Python 3.9. Your coworker types `python` and gets Python 3.12. Same command, different results. Or you install a tool and the shell can't find it. Understanding `$PATH` and how the shell finds commands makes these problems obvious.

## How People Misunderstand It

1. **"Commands are built into the system"** — No. Commands are files on disk. `grep` is `/usr/bin/grep`. The shell just knows where to look.
2. **"I installed it, so it should work"** — The shell doesn't scan the entire filesystem. It only checks the directories listed in `$PATH`, in order.
3. **"I'll just use the full path"** — Works but misses the point. `$PATH` is the lookup mechanism that lets you type `grep` instead of `/usr/bin/grep`.

## The Mental Model

When you type a command name (not a path), the shell does this:

```
You type: grep "error" /var/log/messages

Shell asks:
  1. Is "grep" a shell builtin? → No
  2. Is "grep" a function I know? → No
  3. Is "grep" an alias? → No (well, maybe — check)
  4. Have I cached "grep" in my hash table? → Maybe
  5. Search $PATH directories, left to right:
     /home/user/.local/bin/grep  → not found
     /usr/local/bin/grep         → not found
     /usr/bin/grep               → FOUND! Use this one.
```

### The Full Resolution Order

```
1. Aliases         →  Only in interactive shells
2. Functions       →  Defined with name() { }
3. Builtins        →  Built into the shell (cd, echo, type, etc.)
4. Hash table      →  Cached PATH lookups from previous runs
5. $PATH search    →  Left-to-right scan of directories
```

This order is why an alias can "override" a command, why a function can shadow a builtin, and why the first match in `$PATH` wins.

### $PATH Is Just a String

```bash
echo "$PATH"
# /home/user/.local/bin:/usr/local/bin:/usr/bin:/usr/sbin

# It's a colon-separated list of directories
# Make it readable:
echo "$PATH" | tr ':' '\n'
```

The shell takes each directory, in order, and checks if an executable file with the given name exists there. First match wins.

## Real Fedora Examples

### Viewing Your PATH

```bash
# Human-readable PATH
echo "$PATH" | tr ':' '\n'

# Typical Fedora PATH for a regular user:
# /home/user/.local/bin
# /usr/local/bin
# /usr/bin
# /usr/sbin

# Root's PATH often includes:
# /usr/local/sbin
# /usr/local/bin
# /usr/sbin
# /usr/bin
```

### Adding to PATH

```bash
# Temporarily (current shell only):
export PATH="$HOME/my-tools:$PATH"

# Permanently (add to ~/.bash_profile):
export PATH="$HOME/.local/bin:$PATH"

# Prepending vs appending:
export PATH="$HOME/mybin:$PATH"    # mybin checked FIRST (override system tools)
export PATH="$PATH:$HOME/mybin"    # mybin checked LAST (fallback)
```

### Finding Where Commands Live

```bash
# Where is grep?
which grep              # /usr/bin/grep  (but 'which' lies sometimes — see next file)
type grep               # grep is /usr/bin/grep  (better)
type -a grep            # Shows ALL matches across PATH and aliases

# What package provides a command?
rpm -qf /usr/bin/grep   # grep-3.x-x.fc39.x86_64
dnf provides '*/bin/jq' # What package provides jq?

# Find all executables in PATH with a name:
type -a python           # python is /usr/bin/python
                         # python is /usr/local/bin/python  (if multiple)
```

### The Hash Table

Bash caches command locations in a hash table for speed:

```bash
# Run grep once — Bash caches the path
grep --version > /dev/null

# See the cache:
hash
# hits    command
#    1    /usr/bin/grep

# If you install a new version of grep elsewhere in PATH,
# Bash will still use the cached location!
# Clear the cache:
hash -r          # Clear all
hash -d grep     # Clear just grep
```

This is why "I installed a new version but the old one still runs" happens. The shell cached the old path.

## Common Footguns

**Footgun 1: Installing to a directory not in PATH**
```bash
# You install a tool to /opt/mytool/bin
# Then type the command and get "command not found"
# Because /opt/mytool/bin isn't in your PATH

# Fix: Add it to PATH in ~/.bash_profile
export PATH="/opt/mytool/bin:$PATH"
```

**Footgun 2: PATH order surprise**
```bash
# You have Python 3.9 at /usr/bin/python3 and 3.12 at /usr/local/bin/python3
# Which one runs depends on which directory comes first in PATH

echo "$PATH" | tr ':' '\n' | head -5
# If /usr/local/bin comes before /usr/bin → Python 3.12
# If /usr/bin comes before /usr/local/bin → Python 3.9
```

**Footgun 3: Empty PATH component means current directory**
```bash
# This is dangerous:
export PATH=":/usr/bin"     # Leading colon = empty first entry = current directory
export PATH="/usr/bin:"     # Trailing colon = same thing
export PATH="/usr/bin::/usr/local/bin"  # Double colon = same thing

# Why dangerous? If someone puts a malicious "ls" in the current directory,
# you'll run it instead of /usr/bin/ls.
```

## Why This Matters in Real Systems

- **sudo has its own PATH** — `sudo` resets PATH to a secure default. Your `~/.local/bin` tools won't be found via `sudo`. Use `sudo env PATH="$PATH" mycommand` or configure `secure_path` in `/etc/sudoers`.
- **Cron has a minimal PATH** — Usually just `/usr/bin:/bin`. Scripts that work interactively break in cron because tools in `/usr/local/bin` or `~/.local/bin` aren't found. Always use full paths in cron scripts or set PATH explicitly.
- **systemd services don't use your PATH** — They have their own `PATH` defined in the unit file or systemd defaults.

```bash
# See what PATH cron uses:
echo '* * * * * echo $PATH > /tmp/cron-path' | crontab -
# Wait a minute, then:
cat /tmp/cron-path
# Usually: /usr/bin:/bin

# See what PATH sudo uses:
sudo env | grep PATH
```

## Exercise

1. Run `echo "$PATH" | tr ':' '\n' | nl` to see your PATH directories numbered.
2. Run `type -a python3` (or `python`) — are there multiple? Which one wins?
3. Create a script called `hello` in `/tmp`, make it executable, and try to run it by just typing `hello`. It won't work. Now add `/tmp` to your PATH temporarily and try again.
4. Run a command, then check `hash` to see the cached path. Install a tool to a different location and observe that the old cached path persists until you `hash -r`.

---

Next: [Builtins vs External Binaries](01-builtins-vs-externals.md)
