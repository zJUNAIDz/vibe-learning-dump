# Shell vs Terminal vs TTY vs Emulator

## What Problem This Solves

People use "shell," "terminal," "command line," and "console" interchangeably. This creates confusion when you need to configure one of them specifically — you don't know which thing you're actually configuring.

## How People Misunderstand It

The most common confusion: "I opened my terminal and typed a command." Which part handled what? Was it the terminal that ran the command? Was it the shell? Where does the kernel fit in?

People also confuse terminal *settings* (colors, fonts, keybindings) with shell *settings* (aliases, prompts, PATH). This leads to editing the wrong config file.

## The Mental Model

Think of it as **layers**:

```
┌──────────────────────────────────────────┐
│  Terminal Emulator (GNOME Terminal, Kitty)│ ← Draws pixels on screen
│  ┌────────────────────────────────────┐  │
│  │  PTY (Pseudo-Terminal)             │  │ ← Kernel device that pretends to be hardware
│  │  ┌──────────────────────────────┐  │  │
│  │  │  Shell (Bash, Zsh)           │  │  │ ← Reads input, runs programs, returns output
│  │  │  ┌────────────────────────┐  │  │  │
│  │  │  │  Programs (ls, grep)   │  │  │  │ ← Execute and produce output
│  │  │  └────────────────────────┘  │  │  │
│  │  └──────────────────────────────┘  │  │
│  └────────────────────────────────────┘  │
└──────────────────────────────────────────┘
```

### What Each Layer Does

**TTY (TeleTYpewriter)**
Originally a physical device — a keyboard + printer connected to a computer via serial line. The kernel still uses this abstraction. When you press `Ctrl+C`, it's the TTY layer (specifically the kernel's terminal line discipline) that sends `SIGINT` to the foreground process, not the shell. The TTY layer handles:
- Line editing (backspace, Ctrl+U to clear line) in "cooked" mode
- Signal generation (Ctrl+C, Ctrl+Z, Ctrl+\)
- Echo (showing what you type)

```bash
# See your TTY device
tty
# Output: /dev/pts/0  (pts = pseudo-terminal slave)

# See all TTYs with logged-in users
who
```

**Pseudo-Terminal (PTY)**
Since you don't have a physical serial terminal, the kernel creates a *virtual* one. A PTY has two ends:
- **Master side**: owned by the terminal emulator
- **Slave side**: the shell reads from and writes to this (`/dev/pts/N`)

The terminal emulator captures your keystrokes, writes them to the master side. The kernel's TTY layer processes them and makes them available on the slave side, where the shell reads them.

**Terminal Emulator**
The GUI application you see on screen. GNOME Terminal, Kitty, Alacritty, Konsole, xterm. Its job:
- Render text output as pixels
- Capture keyboard and mouse input
- Interpret ANSI escape codes (colors, cursor movement)
- Manage fonts, scrollback, window size

It knows nothing about `ls` or `grep`. It just displays whatever text the shell sends through the PTY.

**Shell**
The actual command interpreter. It:
1. Reads a line of input from stdin (which is the PTY slave)
2. Parses the line (expansion, splitting, quoting)
3. Looks up the command
4. Forks a child process and `exec`s the program
5. Waits for the program to exit
6. Prints the prompt and loops back to step 1

The shell is just a program. It's `/usr/bin/bash` or `/usr/bin/zsh`. You can run a shell inside a shell:

```bash
# You're in Bash currently
bash         # Now you're in a NEW Bash process
exit         # Back to the original
```

## Real Fedora Examples

```bash
# What terminal emulator am I using?
# There's no reliable universal command for this, but:
echo $TERM             # Terminal type (e.g., xterm-256color)
echo $TERM_PROGRAM     # Set by some emulators (Kitty, iTerm2)

# What shell is running?
echo $0                # The shell that's running right now
echo $SHELL            # Your DEFAULT shell (may differ from current)
ps -p $$               # Process info for current shell

# List available shells on the system
cat /etc/shells

# Change your default shell
chsh -s /usr/bin/zsh   # Sets default shell to Zsh
# You need to log out and back in for this to take effect
```

## Common Footguns

**Footgun 1: Confusing $SHELL with $0**
```bash
$ echo $SHELL    # /bin/bash  (your default)
$ zsh            # Start a Zsh subprocess
$ echo $SHELL    # /bin/bash  ← STILL shows Bash!
$ echo $0        # zsh        ← THIS shows the current shell
```
`$SHELL` is set at login and never changes during the session. It's your *default* shell, not your *current* shell.

**Footgun 2: "My terminal doesn't support colors"**
It's almost never the terminal. It's usually:
- The program not detecting color support (check `$TERM`)
- The shell not configured to use colors in its prompt
- The program being piped (programs disable color when stdout isn't a terminal)

```bash
# Is stdout a terminal?
[[ -t 1 ]] && echo "stdout is a terminal" || echo "stdout is a pipe/file"

# Force color even in pipes (if the program supports it)
grep --color=always "pattern" file | less -R
```

**Footgun 3: Editing terminal emulator settings to fix shell behavior**
If your command history isn't working, don't look at GNOME Terminal settings. Look at your shell config (`.bashrc` / `.zshrc`). The terminal emulator doesn't know what "command history" is.

## Why This Matters in Real Systems

- When you SSH into a server, you get a new PTY. Your terminal emulator is still local. Understanding this helps you debug why colors or key bindings behave differently remotely.
- When a cron job or systemd service runs your script, there's **no terminal at all**. The script's stdin is `/dev/null` or a pipe. If your script assumes it's running in a terminal (e.g., prompting for input), it will hang or crash.
- Docker containers often have no PTY. That's why `docker exec` needs `-it` (interactive + TTY) to give you an interactive shell.

```bash
# Run with a TTY
docker exec -it mycontainer bash

# Run without (for scripts)
docker exec mycontainer bash -c 'echo hello'
```

## Exercise

1. Run `tty` to see your PTY device. Open another terminal tab and run `tty` again — notice it's a different number.
2. Run `ps -ef | grep pts` to see all processes attached to PTYs.
3. Run `echo hello > /dev/pts/N` (using the other tab's PTY number) — you just sent text directly to another terminal, bypassing the shell entirely.
4. Run `stty -a` to see the TTY settings. Find which key is mapped to "intr" (interrupt = Ctrl+C).

---

Next: [Bash vs Zsh vs POSIX sh](01-bash-vs-zsh-vs-posix.md)
