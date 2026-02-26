# Start Here: Shell Mastery Orientation

## What Is This?

A curriculum that teaches you **how the shell thinks** so you can read any command, write scripts that don't break, and stop copy-pasting blindly.

This is not a reference manual. It's a structured path from "I can use Linux" to "I think in shell."

---

## The Problem This Solves

You've used Linux for a year. You can navigate directories, install packages, edit files. But:

- You copy long commands from StackOverflow and pray they work
- You don't know why quotes matter (until something breaks)
- `2>&1` looks like line noise
- You've never written a script longer than 10 lines that you trusted
- `set -euo pipefail` is something you paste at the top without understanding

This curriculum fixes all of that.

---

## How the Levels Work

Each level fixes a **specific class of confusion**:

| Level | What You'll Stop Being Confused About |
|-------|--------------------------------------|
| 0 | What the shell even is (vs terminal, vs TTY) |
| 1 | Why `type` is better than `which`, how `sudo` works |
| 2 | Why spaces break scripts, what quotes actually do |
| 3 | What `2>&1` means, how pipes work internally |
| 4 | Why scripts "succeed" when they fail |
| 5 | Why your variable vanished in the subshell |
| 6 | How to write Bash like a real programming language |
| 7 | How to read `awk` / `sed` / `grep` one-liners |
| 8 | What Zsh gives you that Bash doesn't |
| 9 | Why SELinux denies things even as root |
| 10 | How to write scripts you'd let run in production |
| 11 | How to actually debug a script |
| 12 | Final checklist — can you read any command? |

---

## Prerequisites

Before you start, verify you can do all of these without looking anything up:

```bash
cd /var/log
ls -la
cat /etc/hostname
cp file.txt /tmp/
rm -rf /tmp/junk/    # You know what -rf does
sudo dnf install htop
grep "error" /var/log/messages
```

If that's comfortable, you're ready. If not, use Linux daily for a few more months first.

---

## What You Need

- **Fedora Linux** (37+ recommended, any RHEL-family works)
- **Bash 5.x** (check: `bash --version`)
- **Zsh 5.9+** (install: `sudo dnf install zsh`)
- A terminal emulator (GNOME Terminal, Kitty, Alacritty — doesn't matter)
- A text editor (VS Code, Neovim, Helix — doesn't matter)

---

## How to Study

### The Right Way

1. **Read the explanation** — understand the mental model
2. **Type the examples yourself** — don't copy-paste (ironic, yes)
3. **Break things intentionally** — remove quotes, change redirections, see what happens
4. **Do the exercises** — they're small but real
5. **Re-read after a week** — you'll understand more the second time

### The Wrong Way

- Skimming for "useful commands" to save somewhere
- Reading without a terminal open
- Skipping Level 2 (quoting) because "I already know quotes"

---

## Time Estimate

| Approach | Time |
|----------|------|
| Sequential, thorough | 4-6 weeks (5-8 hrs/week) |
| Targeted (fix specific confusion) | 1-2 weeks per level |
| Reference after completion | Ongoing |

---

## Start Now

→ [Level 0: What a Shell Actually Is](00-what-a-shell-is/00-shell-vs-terminal-vs-tty.md)
