# Shell Mastery: Think in Shell, Don't Memorize Commands

> **Deep, practical curriculum for Bash + Zsh on Fedora Linux.** Not a cheat sheet. Not for beginners. For people who copy-paste commands and want to finally *read* them.

## Overview

This curriculum treats the shell as what it is: a **programming language glued to the operating system**. It teaches you the mental models that make arbitrary commands readable, not a list of commands to memorize.

After this curriculum, when you see:

```bash
find /var/log -name '*.log' -mtime +30 -exec gzip {} \; 2>/dev/null &
```

You won't feel dread. You'll read it like a sentence.

---

## Who This Is For

✅ **Take this if you:**
- Have used Linux daily for 1+ year
- Are comfortable navigating the filesystem
- Frequently copy-paste commands without fully understanding them
- Want to think in shell, not memorize commands
- Use Fedora (dnf, systemd, SELinux, GNU coreutils)

❌ **Skip this if you:**
- Need a "what is the terminal" introduction
- Want a cheat sheet of useful commands
- Want to learn shell for macOS/BSD (this assumes GNU + Fedora)

---

## Philosophy

### Core Principles

1. **The shell is a programming language** — It has variables, functions, control flow, and I/O. Treat it as one.
2. **Commands are just programs** — Everything is an executable or a builtin. Nothing is magic.
3. **Mental models over command lists** — Understand *how the shell parses your input* and you can decode any command.
4. **Fedora is the ground truth** — systemd, SELinux enforcing, dnf, GNU coreutils. No abstract "Linux" hand-waving.
5. **Reading > memorizing** — The goal is literacy, not a bigger clipboard.

### What This Curriculum Fixes

| Confusion | Level That Fixes It |
|-----------|-------------------|
| "Why did my script break on spaces?" | Level 2: Quoting & Expansion |
| "I don't know what `2>&1` means" | Level 3: Redirection & FDs |
| "My script ran but nothing happened" | Level 4: Exit Codes |
| "Why doesn't my variable work in the subshell?" | Level 5: Variables & Scope |
| "I can't read awk one-liners" | Level 7: Text Processing |
| "Why does this fail even as root?" | Level 9: Fedora Reality |
| "My script works locally but not in CI" | Level 11: Debugging |

---

## Curriculum Structure

### Learning Path

```
┌─────────────────────────────────────────────────────────────┐
│  Level 0: What a Shell Actually Is                           │
│  Shell vs terminal vs TTY. Login vs non-login.               │
│  Why your dotfiles exist.                                    │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 1: Commands Are Just Programs                         │
│  $PATH, builtins vs binaries, sudo, command resolution.      │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 2: Arguments, Quoting & Expansion                     │
│  Where 70% of shell bugs live. Word splitting, globbing,     │
│  "$@", brace expansion. THE critical level.                  │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 3: Pipes, Redirection & File Descriptors              │
│  stdin/stdout/stderr as streams. Composing Unix tools.       │
│  Why pipes aren't temp files.                                │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 4: Exit Codes & Failure Handling                       │
│  $?, set -euo pipefail, && vs ||.                            │
│  Why scripts "succeed" when they shouldn't.                  │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 5: Variables, Scope & Environment                     │
│  Shell vars vs env vars. export. Subshells.                  │
│  Why VAR=value cmd works. Fedora service environments.       │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 6: Bash Scripting as a Real Language                   │
│  Functions, conditionals, [ vs [[, loops, script structure.  │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 7: Text Processing Philosophy                         │
│  grep, sed, awk as tools-with-purpose. jq for JSON.         │
│  Real pipelines on real logs.                                │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 8: Zsh Power (Without Magic)                          │
│  Zsh vs Bash differences. Completions. Why Oh-My-Zsh         │
│  can hurt understanding.                                     │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 9: Fedora-Specific Reality                            │
│  dnf, systemd, SELinux, firewalld. Why things fail           │
│  "even as root."                                             │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 10: Writing Scripts You Can Trust                      │
│  Idempotency, locking, safe temp files, dry-runs, logging.   │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 11: Debugging Shell Scripts                           │
│  set -x, trap, reproducing failures, "works on my machine." │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 12: Shell Literacy Checklist                          │
│  Final self-assessment. Can you read any command?            │
└─────────────────────────────────────────────────────────────┘
```

---

## Module Breakdown

### Level 0: What a Shell Actually Is
- [Shell vs Terminal vs TTY](00-what-a-shell-is/00-shell-vs-terminal-vs-tty.md)
- [Bash vs Zsh vs POSIX sh](00-what-a-shell-is/01-bash-vs-zsh-vs-posix.md)
- [Login vs Non-Login Shells & Dotfiles](00-what-a-shell-is/02-login-shells-and-dotfiles.md)

### Level 1: Commands Are Just Programs
- [$PATH and Command Resolution](01-commands-are-programs/00-path-and-resolution.md)
- [Builtins vs External Binaries](01-commands-are-programs/01-builtins-vs-externals.md)
- [How sudo Actually Works](01-commands-are-programs/02-how-sudo-works.md)

### Level 2: Arguments, Quoting & Expansion
- [Word Splitting: Why Spaces Break Everything](02-arguments-quoting-expansion/00-word-splitting.md)
- [Quoting Rules: Single, Double, None](02-arguments-quoting-expansion/01-quoting-rules.md)
- [Globbing vs Regex](02-arguments-quoting-expansion/02-globbing-vs-regex.md)
- [Command Substitution & Brace Expansion](02-arguments-quoting-expansion/03-substitution-and-expansion.md)
- [Why "$@" Exists](02-arguments-quoting-expansion/04-why-dollar-at-exists.md)

### Level 3: Pipes, Redirection & File Descriptors
- [Streams: stdin, stdout, stderr](03-pipes-redirection-fds/00-streams.md)
- [Redirection Deep Dive](03-pipes-redirection-fds/01-redirection-deep-dive.md)
- [Pipes and Unix Composition](03-pipes-redirection-fds/02-pipes-and-composition.md)

### Level 4: Exit Codes & Failure Handling
- [Exit Codes and $?](04-exit-codes-failure/00-exit-codes.md)
- [set -euo pipefail Explained](04-exit-codes-failure/01-set-flags.md)
- [Writing Scripts That Fail Safely](04-exit-codes-failure/02-failing-safely.md)

### Level 5: Variables, Scope & Environment
- [Shell Variables vs Environment Variables](05-variables-scope-env/00-shell-vs-env-vars.md)
- [Subshells, Export, and Scope](05-variables-scope-env/01-subshells-and-scope.md)
- [Fedora Service Environments](05-variables-scope-env/02-fedora-service-env.md)

### Level 6: Bash Scripting as a Real Language
- [Functions and Their Weird Scoping](06-bash-scripting-language/00-functions.md)
- [Conditionals: [ vs [[ and test](06-bash-scripting-language/01-conditionals.md)
- [Loops That Don't Explode](06-bash-scripting-language/02-loops.md)
- [Script Structure That Doesn't Rot](06-bash-scripting-language/03-script-structure.md)

### Level 7: Text Processing Philosophy
- [grep, sed, awk: Mental Models](07-text-processing/00-grep-sed-awk.md)
- [awk Is a Programming Language](07-text-processing/01-awk-deep-dive.md)
- [JSON in Shell: jq](07-text-processing/02-jq-json-processing.md)
- [Real Pipelines on Real Logs](07-text-processing/03-real-pipelines.md)

### Level 8: Zsh Power (Without Magic)
- [Zsh vs Bash: Real Differences](08-zsh-power/00-zsh-vs-bash.md)
- [The Completion System](08-zsh-power/01-completion-system.md)
- [Why Oh-My-Zsh Can Hurt You](08-zsh-power/02-oh-my-zsh-critique.md)

### Level 9: Fedora-Specific Reality
- [dnf, systemd, journalctl](09-fedora-reality/00-dnf-systemd-journalctl.md)
- [SELinux from the Shell](09-fedora-reality/01-selinux-from-shell.md)
- [When Root Isn't Enough](09-fedora-reality/02-when-root-isnt-enough.md)

### Level 10: Writing Scripts You Can Trust
- [Idempotent, Safe, Production Scripts](10-trustworthy-scripts/00-production-scripts.md)
- [Temp Files, Locking, Logging, Dry-Runs](10-trustworthy-scripts/01-safety-patterns.md)

### Level 11: Debugging Shell Scripts
- [set -x, trap, and Reproducing Failures](11-debugging/00-debugging-tools.md)
- [Why "Works on My Machine" Happens](11-debugging/01-environment-differences.md)

### Level 12: Shell Literacy Checklist
- [Final Self-Assessment](12-shell-literacy-checklist/00-checklist.md)

---

## How to Use This Curriculum

### Sequential (Recommended)

Go in order. Each level builds on the previous. **4-6 weeks** at 5-8 hours/week.

### Fix-a-Confusion Track

Jump to the level that matches your current pain point (see table above), but read Level 0 and Level 2 regardless — they're foundational.

### Reference Mode

After completing the curriculum, use the [QUICK_REFERENCE.md](QUICK_REFERENCE.md) as a lookup table. But finish the curriculum first — you need the mental models to make the reference useful.

---

## Conventions Used

- `$` prefix means "type this at a shell prompt"
- `#` prefix means "type this as root" (or with sudo)
- **Footgun** = a common mistake that blows up in production
- **Mental Model** = the way to think about something so it makes sense forever
- All examples assume Fedora with Bash 5.x or Zsh 5.9+
- File paths assume GNU coreutils behavior

---

## Prerequisites

- Fedora Linux installed (or any RHEL-family distro)
- 1+ year of daily Linux usage
- Comfort with `cd`, `ls`, `cp`, `mv`, `rm`, `cat`
- A text editor you're comfortable with
- Willingness to type commands and observe what happens

---

## What You Will NOT Learn Here

- Fish shell (incompatible syntax, different philosophy)
- macOS/BSD shell differences (different `sed`, `find`, etc.)
- PowerShell
- Comprehensive sysadmin training (this is shell mastery, not RHCSA prep)
