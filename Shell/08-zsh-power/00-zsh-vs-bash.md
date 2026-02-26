# Zsh vs Bash — The Real Differences

## What Problem This Solves

You use Zsh as your interactive shell but write scripts in Bash. Or you switched to Zsh because someone told you to but you don't know what you gained. This file covers the **actual differences** that affect your daily work — not cosmetic things.

## The Key Differences That Matter

### 1. Array Indexing

```bash
# Bash: 0-indexed
arr=(a b c)
echo "${arr[0]}"    # a

# Zsh: 1-indexed
arr=(a b c)
echo "${arr[1]}"    # a
echo "${arr[0]}"    # empty! (nothing at index 0)
```

**This is the #1 source of bugs when moving between Bash and Zsh.** Always be aware of which shell your script targets.

### 2. Word Splitting

```bash
# Bash: Unquoted variables undergo word splitting
x="one two three"
for word in $x; do echo "$word"; done
# one
# two
# three

# Zsh: NO word splitting by default
x="one two three"
for word in $x; do echo "$word"; done
# one two three   (single item!)

# Zsh: To get splitting, use ${=var}:
for word in ${=x}; do echo "$word"; done
# one
# two
# three
```

This means **many Bash gotchas don't apply in Zsh**, but it also means scripts that rely on word splitting break.

### 3. Glob Behavior

```bash
# Bash: Unmatched glob stays literal
echo *.xyz        # *.xyz (if no .xyz files exist)

# Zsh: Unmatched glob is an ERROR by default
echo *.xyz        # zsh: no matches found: *.xyz

# Zsh: To get Bash behavior:
setopt NULL_GLOB    # Unmatched glob → nothing (empty)
setopt NO_NOMATCH   # Unmatched glob → literal (like Bash)
```

### 4. Zsh Extended Globs (Built-in)

Zsh has powerful globs without needing `shopt -s extglob`:

```bash
# Recursive glob (both shells, but native to Zsh):
ls **/*.txt

# Glob qualifiers (ZSH ONLY — extremely powerful):
ls *(.)            # Regular files only
ls *(/)            # Directories only
ls *(@)            # Symlinks only
ls *(x)            # Executable files only
ls *(m-7)          # Modified in last 7 days
ls *(Lk+100)       # Files larger than 100KB
ls *(om[1,5])      # 5 most recently modified files
ls *(On)           # Sort by name (reverse)
ls *(oL)           # Sort by size (ascending)

# Combine qualifiers:
ls *.log(m-1Lk+10)  # .log files modified today AND larger than 10KB
```

### 5. Parameter Expansion

Zsh has additional expansion flags:

```bash
# Convert case:
x="Hello World"
echo ${x:l}          # hello world (lowercase) — Zsh
echo ${x:u}          # HELLO WORLD (uppercase) — Zsh
echo ${x,,}          # hello world — Bash 4+
echo ${x^^}          # HELLO WORLD — Bash 4+

# Split on a character:
path="/usr/local/bin"
echo ${(s:/:)path}   # usr local bin — Zsh

# Join array elements:
arr=(a b c)
echo ${(j:,:)arr}    # a,b,c — Zsh

# Unique array elements:
arr=(a b a c b)
echo ${(u)arr}       # a b c — Zsh

# Sort:
arr=(cherry apple banana)
echo ${(o)arr}       # apple banana cherry — Zsh (ascending)
echo ${(O)arr}       # cherry banana apple — Zsh (descending)
```

### 6. Pipe Behavior

```bash
# Bash: ALL pipe segments run in subshells (by default)
echo "hello" | read word
echo "$word"         # Empty! (read was in subshell)

# Zsh: LAST pipe segment runs in current shell
echo "hello" | read word
echo "$word"         # "hello" (works!)
```

This means the classic "pipe subshell problem" from Level 5 **doesn't exist in Zsh**.

## Zsh Interactive Power

These features make Zsh's interactive experience superior:

### Right Prompt

```bash
# Left prompt:
PROMPT='%n@%m %~ $ '

# Right prompt (shows on the right side of the terminal):
RPROMPT='%T'         # Shows current time
RPROMPT='%?'         # Shows last exit code
```

### Better History

```bash
# History substring search (with arrow keys):
# Type part of a command, press Up — only matches containing that text appear

# Share history across all terminal sessions:
setopt SHARE_HISTORY

# Don't store duplicates:
setopt HIST_IGNORE_ALL_DUPS

# Don't store commands starting with a space:
setopt HIST_IGNORE_SPACE
```

### Auto-cd

```bash
# In Zsh, you can type a directory name without cd:
setopt AUTO_CD
/tmp               # Same as: cd /tmp
..                 # Same as: cd ..
```

### Spelling Correction

```bash
setopt CORRECT
# Now if you mistype:
$ gti status
# zsh: correct 'gti' to 'git' [nyae]?
```

### Directory Stack

```bash
setopt AUTO_PUSHD
setopt PUSHD_IGNORE_DUPS
# Every cd automatically pushes to the directory stack
cd /tmp
cd /var/log
cd ~/projects
dirs -v            # Show numbered stack
cd -2              # Jump to stack entry 2 (/tmp)
```

## Zsh Configuration

### Config File Order

```
~/.zshenv       → Always loaded (even for scripts). Put PATH here.
~/.zprofile     → Login shells only. Rarely needed.
~/.zshrc        → Interactive shells. Put your config here.
~/.zlogin       → Login shells, after .zshrc. Rarely needed.
~/.zlogout      → On logout.
```

### A Minimal .zshrc

```bash
# ~/.zshrc — No framework, just what you need

# ── History ─────────────────────────────────────
HISTFILE=~/.zsh_history
HISTSIZE=50000
SAVEHIST=50000
setopt SHARE_HISTORY
setopt HIST_IGNORE_ALL_DUPS
setopt HIST_REDUCE_BLANKS

# ── Options ─────────────────────────────────────
setopt AUTO_CD
setopt AUTO_PUSHD
setopt PUSHD_IGNORE_DUPS
setopt CORRECT
setopt NO_CASE_GLOB
setopt GLOB_DOTS              # Include dotfiles in glob matches
setopt EXTENDED_GLOB          # Enable # ~ ^ in globs

# ── Prompt ──────────────────────────────────────
autoload -Uz vcs_info
precmd() { vcs_info }
zstyle ':vcs_info:git:*' formats '%b'
setopt PROMPT_SUBST
PROMPT='%F{blue}%~%f %F{green}${vcs_info_msg_0_}%f
%F{yellow}$%f '

# ── Completion ──────────────────────────────────
autoload -Uz compinit && compinit
zstyle ':completion:*' matcher-list 'm:{a-z}={A-Z}'     # Case-insensitive
zstyle ':completion:*' menu select                        # Arrow-key menu
zstyle ':completion:*' list-colors ${(s.:.)LS_COLORS}   # Colored completions

# ── Key Bindings ────────────────────────────────
bindkey '^[[A' history-beginning-search-backward   # Up arrow
bindkey '^[[B' history-beginning-search-forward    # Down arrow
bindkey '^R' history-incremental-search-backward   # Ctrl+R

# ── Aliases ─────────────────────────────────────
alias ls='ls --color=auto'
alias ll='ls -lah'
alias grep='grep --color=auto'

# ── PATH ────────────────────────────────────────
typeset -U PATH    # Remove duplicates from PATH
path=(~/.local/bin $path)
```

## When to Write in Bash vs Zsh

| Situation | Use |
|-----------|-----|
| Interactive daily use | Zsh |
| Scripts shared with others | Bash (more portable) |
| Scripts for servers/CI | Bash or POSIX sh |
| Scripts only for your machine | Either (be consistent) |
| systemd services | Bash (or none — use exec) |

**Shebang rule:**
```bash
#!/usr/bin/env bash    # For Bash scripts
#!/usr/bin/env zsh     # For Zsh scripts (rare)
#!/bin/sh              # For POSIX scripts (maximum portability)
```

## Common Footguns

### 1. Script Works in Terminal but Not with `bash script.sh`

Your terminal is Zsh. Your script has `#!/bin/bash`. Zsh features don't work in Bash:

```bash
# This works in Zsh terminal but NOT in a Bash script:
arr=(a b c)
echo $arr[1]          # Zsh: "a"  |  Bash: "a[1]" (literally!)
```

### 2. `.bashrc` vs `.zshrc`

When you switch from Bash to Zsh, your `.bashrc` isn't loaded. You need to move your config to `.zshrc` (or source `.bashrc` from `.zshrc`, but most Bash-specific syntax won't work).

### 3. Completion Scripts

Bash completions don't work in Zsh. You need Zsh-specific completions or a compatibility layer (`bashcompinit`):

```bash
autoload -Uz bashcompinit && bashcompinit
source /path/to/bash_completion_script
```

## Exercise

1. Compare array behavior: create an array in both Bash and Zsh, access elements by index, and note the difference.

2. Demonstrate word splitting: create a variable with spaces, iterate over it in Bash (splits) vs Zsh (doesn't split), then use `${=var}` in Zsh to split.

3. Use Zsh glob qualifiers to find: all executable files in `/usr/bin` modified in the last 24 hours that are larger than 1MB.

4. Set up a minimal `.zshrc` without any framework — include history, prompt with git branch, and completion.

---

Next: [Zsh Completion System](01-completion-system.md)
