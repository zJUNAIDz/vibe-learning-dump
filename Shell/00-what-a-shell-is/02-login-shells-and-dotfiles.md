# Login vs Non-Login Shells & Dotfiles

## What Problem This Solves

You add `export PATH="$HOME/.local/bin:$PATH"` to `.bashrc` but it doesn't work when you SSH in. Or you add an alias to `.bash_profile` but it doesn't work in new terminal tabs. The dotfile system seems arbitrary until you understand the login vs non-login distinction.

## How People Misunderstand It

1. **"Just put everything in .bashrc"** — Works for some setups but breaks others (SSH sessions, display managers).
2. **"I edited .bash_profile but nothing changed in my terminal"** — GNOME Terminal opens non-login shells by default. It reads `.bashrc`, not `.bash_profile`.
3. **"Why are there so many config files?"** — Historical accident + real architectural reasons. It's messy, but there is logic.

## The Mental Model

The shell asks itself **two questions** when it starts:

1. **Am I a login shell?** (Is this the first shell for this user session?)
2. **Am I interactive?** (Is a human typing commands?)

The answers determine which files get read:

### Bash

```
Login + Interactive (SSH session, console login):
  1. /etc/profile
  2. ~/.bash_profile  (OR ~/.bash_login OR ~/.profile — first one found)
  3. On exit: ~/.bash_logout

Non-Login + Interactive (new terminal tab in GUI):
  1. /etc/bash.bashrc  (Fedora: /etc/bashrc)
  2. ~/.bashrc

Non-Interactive (script execution):
  1. File in $BASH_ENV (if set)
  2. Nothing else
```

### The Fedora Trick

Fedora's default `~/.bash_profile` contains this:

```bash
# .bash_profile

# Get the aliases and functions
if [ -f ~/.bashrc ]; then
    . ~/.bashrc
fi

# User specific environment and startup programs
```

This is the crucial line: **`.bash_profile` sources `.bashrc`**. This means on Fedora, `.bashrc` runs in BOTH login and non-login shells. So on Fedora specifically, putting everything in `.bashrc` usually works. But this is Fedora convention, not universal behavior.

### Zsh

```
Login + Interactive:
  1. /etc/zsh/zprofile  (or /etc/zprofile)
  2. ~/.zprofile
  3. /etc/zsh/zshrc  (or /etc/zshrc)
  4. ~/.zshrc
  5. /etc/zsh/zlogin
  6. ~/.zlogin
  On exit: ~/.zlogout

Non-Login + Interactive:
  1. /etc/zsh/zshrc
  2. ~/.zshrc

Always (login and non-login):
  1. /etc/zsh/zshenv
  2. ~/.zshenv
```

Zsh is more structured: `.zshenv` ALWAYS runs (even for scripts), `.zshrc` runs for interactive shells, `.zprofile`/`.zlogin` only for login shells.

### Visual Decision Tree

```
Shell starts
  │
  ├─ Is it a login shell?
  │   ├─ Yes → Read profile files (.bash_profile / .zprofile)
  │   │        Then, if interactive → Read rc files (.bashrc / .zshrc)
  │   │        (Fedora: .bash_profile sources .bashrc, so both run)
  │   │
  │   └─ No → Is it interactive?
  │       ├─ Yes → Read rc files (.bashrc / .zshrc) only
  │       └─ No  → Read almost nothing (scripts)
```

## What Goes Where

| Config | Put In | Because |
|--------|--------|---------|
| `$PATH` modifications | `.bash_profile` / `.zshenv` | Only need to set once per session, inherited by child processes |
| Aliases | `.bashrc` / `.zshrc` | Aliases aren't inherited; need to be set in every interactive shell |
| Functions | `.bashrc` / `.zshrc` | Same reason as aliases |
| Prompt (`PS1`) | `.bashrc` / `.zshrc` | Only matters in interactive shells |
| `export` variables | `.bash_profile` / `.zshenv` | Set once, inherited |
| Shell options (`shopt`, `setopt`) | `.bashrc` / `.zshrc` | Per-shell settings |
| `ssh-agent` startup | `.bash_profile` / `.zprofile` | Once per session |

## Real Fedora Examples

### How Fedora's Default Setup Works

```bash
# Check what Fedora ships:
cat /etc/profile         # System-wide login shell config
cat /etc/bashrc          # System-wide interactive Bash config (Fedora-specific path)
cat ~/.bash_profile      # Your login shell config
cat ~/.bashrc            # Your interactive Bash config
```

### Detecting Login vs Non-Login

```bash
# In Bash:
shopt -q login_shell && echo "login shell" || echo "not login shell"

# In Zsh:
[[ -o login ]] && echo "login shell" || echo "not login shell"

# Works in both:
echo $0
# "-bash" or "-zsh" (leading dash) = login shell
# "bash" or "zsh" (no dash) = non-login shell
```

### When Each Type Happens

```
Login shell triggered by:
  - SSH session
  - Console login (Ctrl+Alt+F2)
  - `su - username` (the dash matters!)
  - `bash --login`

Non-login interactive shell triggered by:
  - Opening a new tab in GNOME Terminal
  - Running `bash` or `zsh` manually
  - `su username` (no dash!)
  - tmux new pane/window (usually)

Non-interactive shell:
  - Running a script: `bash script.sh`
  - Command substitution: $(cmd)
  - Cron jobs
```

## Common Footguns

**Footgun 1: `su` vs `su -`**
```bash
su root          # Non-login shell. Keeps YOUR environment. May not have root's PATH.
su - root        # Login shell. Gets root's full environment. This is usually what you want.
```

**Footgun 2: PATH set in .bashrc gets duplicated**
Every new subshell re-runs `.bashrc`, appending to PATH again:
```bash
# BAD (in .bashrc):
export PATH="$HOME/.local/bin:$PATH"
# After opening 5 nested shells, PATH has 5 copies of ~/.local/bin

# BETTER (in .bashrc):
case ":$PATH:" in
  *":$HOME/.local/bin:"*) ;;  # Already there
  *) export PATH="$HOME/.local/bin:$PATH" ;;
esac

# BEST: Put PATH in .bash_profile, not .bashrc
```

**Footgun 3: Aliases don't work in scripts**
```bash
# In .bashrc:
alias ll='ls -la'

# In script.sh:
#!/bin/bash
ll /tmp    # ERROR: ll: command not found
# Scripts don't read .bashrc (non-interactive), so aliases aren't defined.
# Solution: Use functions instead of aliases if you need them in scripts.
```

**Footgun 4: SSH + non-interactive commands skip .bashrc**
```bash
# This reads .bash_profile (login shell, interactive):
ssh server

# This reads NOTHING on most systems:
ssh server 'echo $PATH'
# The shell is non-interactive AND non-login — no dotfiles loaded
# On Fedora with default sshd config, .bashrc IS sourced for this case
# But DON'T rely on it — other distros won't.
```

## Why This Matters in Real Systems

- **CI/CD pipelines** run scripts in non-interactive, non-login shells. Your PATH, aliases, and fancy functions don't exist. Scripts must be self-contained.
- **Cron jobs** run with a minimal environment. If your script depends on PATH being set up by `.bashrc`, it will fail in cron.
- **systemd services** have their own environment mechanism (`Environment=` / `EnvironmentFile=`). They don't read any shell dotfiles.
- **Docker containers** typically run with `bash -c "command"` (non-interactive) or `bash` (non-login interactive). Your dotfiles may or may not be present.

## Exercise

1. Run `shopt -q login_shell && echo LOGIN || echo NOT-LOGIN` in:
   - A GNOME Terminal tab
   - An SSH session to localhost: `ssh localhost`
   - After running `bash --login`
   
   Note the differences.

2. Add `echo "PROFILE LOADED"` to the end of `~/.bash_profile` and `echo "BASHRC LOADED"` to the end of `~/.bashrc`. Open a new terminal tab and note which messages appear. Then SSH to localhost and note which messages appear.

3. Run `env` in an interactive shell and in a cron job (`* * * * * env > /tmp/cron-env.txt`). Compare the output. Count how many variables are missing in the cron environment.

4. (Clean up: Remove the echo lines from your dotfiles when done.)

---

Next: [Level 1: $PATH and Command Resolution](../01-commands-are-programs/00-path-and-resolution.md)
