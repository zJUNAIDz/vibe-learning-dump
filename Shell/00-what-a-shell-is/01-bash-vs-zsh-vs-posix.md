# Bash vs Zsh vs POSIX sh

## What Problem This Solves

You write a script that works on your machine (Zsh) and it breaks on a server (Bash). Or you use `[[` in a script with `#!/bin/sh` and it fails on Debian. Understanding which shell is which — and what POSIX sh actually means — prevents this entire class of bugs.

## How People Misunderstand It

1. **"Bash and Zsh are basically the same"** — They share heritage but differ in important ways (word splitting defaults, array indexing, globbing behavior).
2. **"sh is Bash"** — On Fedora, `/bin/sh` is symlinked to Bash, which runs in POSIX-compatibility mode. On Debian/Ubuntu, `/bin/sh` is `dash`, a minimal shell. Scripts that assume `sh = bash` break across distros.
3. **"I should write scripts in Zsh because that's what I use"** — Interactive shell choice and script shell choice are independent decisions.

## The Mental Model

Think of three layers:

```
POSIX sh specification
   └── What any compliant shell MUST support
   └── The "portable" subset: no arrays, no [[, no $()$() nesting in old versions, etc.

Bash (Bourne Again Shell)
   └── POSIX sh + arrays, [[ ]], process substitution, $'...' quoting,
       brace expansion, better parameter expansion, etc.
   └── The default scripting shell on most Linux systems
   └── When invoked as "sh", restricts itself to POSIX mode

Zsh (Z Shell)
   └── POSIX-ish + massive interactive features
   └── Different defaults (no word splitting on unquoted expansion!)
   └── Superior completion, globbing, prompt system
   └── NOT Bash-compatible — many Bash scripts break in Zsh
```

### Key Decision Framework

| Situation | Use |
|-----------|-----|
| Interactive daily use | Zsh (better UX) or Bash (if you prefer simplicity) |
| System scripts, automation | Bash (available everywhere, well-understood) |
| Maximum portability | POSIX sh (but you lose features) |
| One-off personal scripts | Whatever you want |

## Real Differences That Bite You

### Array Indexing

```bash
# Bash: arrays are 0-indexed
arr=(a b c)
echo "${arr[0]}"    # a

# Zsh: arrays are 1-indexed (like humans count)
arr=(a b c)
echo "${arr[1]}"    # a
echo "${arr[0]}"    # empty!
```

### Word Splitting on Unquoted Variables

```bash
# Bash: unquoted variables undergo word splitting
file="my file.txt"
ls $file            # Bash: ls tries "my" and "file.txt" separately — BROKEN

# Zsh: unquoted variables do NOT word-split by default
file="my file.txt"
ls $file            # Zsh: ls tries "my file.txt" as one arg — WORKS
```

This means **bad Bash habits (forgetting quotes) work in Zsh**, which hides bugs. When you move that script to Bash, it breaks.

### Glob Behavior

```bash
# Zsh: error if glob matches nothing (by default)
ls *.xyz            # Zsh: "no matches found: *.xyz" — error!

# Bash: passes the literal string "*.xyz" to ls (by default)
ls *.xyz            # Bash: ls tries to open a file literally named "*.xyz"
```

Zsh's behavior is actually *safer* — it tells you nothing matched instead of silently doing something weird.

### `[[` vs `[`

Both Bash and Zsh support `[[`, but POSIX sh does not:

```sh
#!/bin/sh
# This WORKS on Fedora (sh is Bash in POSIX mode... but [[ still works there)
# This BREAKS on Debian (sh is dash, which doesn't have [[)
[[ -f /etc/hostname ]] && echo "exists"
```

**Rule**: In `#!/bin/sh` scripts, use `[` (single bracket). In `#!/bin/bash` or `#!/bin/zsh` scripts, use `[[` (double bracket, which is better in every way).

## Real Fedora Context

```bash
# What is /bin/sh on Fedora?
ls -la /bin/sh
# lrwxrwxrwx. 1 root root 4 ... /bin/sh -> bash

# Bash runs in POSIX mode when invoked as "sh"
# You can test this:
sh -c 'echo $BASHOPTS' | tr ':' '\n' | grep posix
# Should not show "posix" in regular bash, but will in sh mode

# What shell does root use?
grep root /etc/passwd
# root:x:0:0:root:/root:/bin/bash

# What shells are installed?
cat /etc/shells
```

## When to Use What

### Your Interactive Shell: Zsh

Zsh has objectively better interactive features:
- Tab completion that actually understands command arguments
- Spelling correction
- Better glob patterns (`**/*.txt` recursive glob)
- Right-side prompt
- Better history search

### Your Script Shebang: Bash

```bash
#!/usr/bin/env bash
```

Why Bash, not Zsh, for scripts:
- Bash is installed on every Linux system by default
- Zsh may not be installed on servers, containers, CI
- Bash's behavior is better documented for scripting
- Most StackOverflow answers assume Bash
- Zsh's "helpful" defaults (no word splitting) mask bugs

Why `#!/usr/bin/env bash` instead of `#!/bin/bash`:
- On some systems, Bash lives at `/usr/local/bin/bash` (FreeBSD, NixOS)
- `env` searches `$PATH` for the binary — more portable

### When POSIX sh Matters

Almost never, unless:
- You're writing a package's init script
- You're targeting embedded systems with `busybox`
- You need to run on very old or minimal systems

For everyday scripts, Bash is fine. Don't handicap yourself with POSIX sh unless you have a real portability reason.

## Common Footguns

**Footgun 1: Writing Bash syntax in a `#!/bin/sh` script**
```sh
#!/bin/sh
# This array syntax is Bash-only. On Debian, this script BREAKS.
files=(*.txt)
```

**Footgun 2: Testing scripts in Zsh, deploying in Bash**
```zsh
# Works in Zsh (no word splitting on $file):
for file in $(find . -name '*.log'); do rm $file; done

# Breaks in Bash on filenames with spaces:
# "my access.log" becomes two arguments: "my" and "access.log"
```

**Footgun 3: Assuming `echo` behaves the same everywhere**
```bash
echo -e "hello\tworld"    # Bash: prints tab. POSIX sh: prints "-e hello\tworld"
printf "hello\tworld\n"   # Works everywhere. Prefer printf.
```

## Exercise

1. Open two terminals. Run `bash` in one, `zsh` in the other.
2. In both, run:
   ```bash
   arr=(one two three)
   echo ${arr[0]}
   echo ${arr[1]}
   ```
   Notice the indexing difference.

3. In both, run:
   ```bash
   var="hello world"
   echo $var | wc -w
   ```
   Bash says 2 words (word splitting). Zsh says... check yourself.

4. Create a file `test.sh` with `#!/bin/sh` and put `[[ -f /etc/hostname ]] && echo yes` inside. Run it on Fedora (works — sh is Bash). Then try `dash -c '[[ -f /etc/hostname ]] && echo yes'` if dash is installed (`sudo dnf install dash`).

---

Next: [Login vs Non-Login Shells & Dotfiles](02-login-shells-and-dotfiles.md)
