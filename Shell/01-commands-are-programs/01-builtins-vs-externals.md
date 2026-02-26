# Builtins vs External Binaries

## What Problem This Solves

You type `echo` and it works. You type `cd` and it works. But `echo` could be `/usr/bin/echo` OR a shell builtin. `cd` is *always* a builtin — it literally cannot be an external program. Understanding why some commands are built into the shell and others are standalone programs clarifies a lot of confusing behavior.

## How People Misunderstand It

1. **"All commands are programs on disk"** — Some commands (`cd`, `export`, `source`) are built into the shell. They MUST be, because they modify the shell's own state.
2. **"`which` tells me where a command is"** — `which` only searches `$PATH`. It knows nothing about builtins, functions, or aliases. It will tell you `which cd` is "not found" even though `cd` works fine.
3. **"It doesn't matter if it's a builtin or not"** — It matters for performance (builtins don't fork), for `sudo` (you can't `sudo` a builtin), and for debugging (different error messages).

## The Mental Model

**Builtins** are code compiled into the shell binary itself. When you run a builtin, **no new process is created**. The shell handles it directly.

**External commands** are separate executable files. When you run one, the shell forks a child process, execs the program, and waits for it to finish.

### Why Some Commands MUST Be Builtins

```
cd        → Changes the current directory of the shell process.
            An external program can't change its PARENT's directory.
            
export    → Modifies the shell's own environment.
            An external program can't modify its parent's environment.

source (.) → Reads and executes commands in the current shell.
             An external program runs in its own process — can't 
             define variables or functions in the parent.

exec      → Replaces the shell process with another program.
            Must be done from inside the shell.

exit      → Terminates the shell process.
            Must be done from inside the shell.
```

These are structurally impossible to implement as external programs. They change the shell's own state.

### Commands That Are BOTH

Some commands exist as both a builtin AND an external binary:

```bash
type echo
# echo is a shell builtin

type -a echo
# echo is a shell builtin
# echo is /usr/bin/echo

type printf
# printf is a shell builtin

type -a printf
# printf is a shell builtin
# printf is /usr/bin/printf
```

The builtin version wins (runs by default). The external version exists for programs that aren't shells (like `xargs` or `find -exec`).

## The Right Tool: `type`

```bash
# type tells you what something REALLY is:
type cd          # cd is a shell builtin
type ls          # ls is aliased to 'ls --color=auto'  (on Fedora)
type grep        # grep is /usr/bin/grep
type type        # type is a shell builtin (meta!)

# type -a shows ALL matches:
type -a echo     
# echo is a shell builtin
# echo is /usr/bin/echo

# type -t shows JUST the type (useful in scripts):
type -t cd       # builtin
type -t ls       # alias
type -t grep     # file
type -t myfunc   # function
```

### Why `which` Lies

```bash
# which only searches $PATH:
which cd         # "no cd in (/usr/local/bin:/usr/bin:...)"  — misleading!
which echo       # /usr/bin/echo  — exists, but the builtin runs first!

# which also can't see aliases or functions
alias ll='ls -la'
which ll         # "no ll in ..." — but ll works!

# which itself may be an alias on some systems:
type which       # On some distros: "which is aliased to ..."
```

**Use `type` for humans, `command -v` for scripts:**

```bash
# In scripts, use command -v (POSIX portable):
if command -v jq &>/dev/null; then
    echo "jq is available"
else
    echo "jq is not installed"
fi

# Don't use 'which' in scripts — it's unreliable
```

## Real Fedora Examples

### Fedora's Default Aliases

Fedora sets up aliases in `/etc/profile.d/` that affect `type` output:

```bash
type ls
# ls is aliased to 'ls --color=auto'

type grep
# grep is aliased to 'grep --color=auto'

# To bypass the alias and run the raw command:
\ls                  # Backslash bypasses alias
command ls           # 'command' bypasses aliases AND functions
/usr/bin/ls          # Full path bypasses everything

# To see the underlying command after bypassing aliases:
type -a ls
# ls is aliased to 'ls --color=auto'
# ls is /usr/bin/ls
```

### Builtin vs External Behavior Differences

```bash
# External echo and builtin echo differ:
echo -e "hello\tworld"        # Builtin: interpretation depends on Bash version/config
/usr/bin/echo -e "hello\tworld"   # External: always interprets \t

# Timing difference:
time for i in $(seq 1 1000); do echo hi > /dev/null; done   # Fast (builtin)
time for i in $(seq 1 1000); do /usr/bin/echo hi > /dev/null; done  # Slow (fork+exec each time)
```

### The `enable` Builtin

You can disable builtins to force external command usage:

```bash
enable -n echo     # Disable the echo builtin
type echo          # echo is /usr/bin/echo (now uses external)
enable echo        # Re-enable the builtin
```

## Common Footguns

**Footgun 1: `sudo cd`**
```bash
sudo cd /root/secret
# Error: sudo: cd: command not found
# cd is a builtin — sudo can't run it because sudo runs external programs.

# What you actually want:
sudo -i            # Get a root shell, then cd
sudo bash -c 'cd /root/secret && ls'  # Run commands in a root subshell
```

**Footgun 2: `sudo echo` into protected files**
```bash
sudo echo "data" > /etc/myconfig
# FAILS. The redirection is done by YOUR shell (not root).
# sudo only elevates the 'echo' command.

# Solutions:
echo "data" | sudo tee /etc/myconfig         # tee runs as root
sudo bash -c 'echo "data" > /etc/myconfig'   # Entire command as root
```

**Footgun 3: Assuming `command not found` means "not installed"**
```bash
$ mycmd
bash: mycmd: command not found

# This could mean:
# 1. mycmd is genuinely not installed
# 2. mycmd is installed but not in $PATH
# 3. mycmd exists but isn't executable (chmod +x missing)
# 4. The file exists but has a bad shebang

# Diagnose:
find / -name mycmd 2>/dev/null     # Is it somewhere on disk?
file /path/to/mycmd                # What type of file is it?
ls -la /path/to/mycmd              # Is it executable?
head -1 /path/to/mycmd             # What's the shebang?
```

## Why This Matters in Real Systems

- `type` is essential for debugging "why is this running the wrong version" problems, especially with Python, Node.js, and other tools that have version managers.
- Understanding builtins explains why `sudo cd` fails, why `sudo export` fails, and why you can't `strace cd`.
- In performance-critical scripts (tight loops), knowing that builtins don't fork means preferring `echo` (builtin) over calling an external command 10,000 times.
- When writing portable scripts, `command -v` is POSIX; `which` is not.

## Exercise

1. Run `type -a echo`, `type -a printf`, `type -a test`, `type -a [`. Which are both builtins AND external?
2. Run `compgen -b` (Bash) to list ALL builtins. How many are there? How many did you already know?
3. Try `sudo cd /root`. Observe the error. Then try `sudo ls /root`. Explain why one works and the other doesn't.
4. Run `type -t` on 10 commands you use daily. Categorize each as builtin, alias, function, or file.

---

Next: [How sudo Actually Works](02-how-sudo-works.md)
