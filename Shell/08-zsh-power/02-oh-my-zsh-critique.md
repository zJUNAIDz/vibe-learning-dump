# Oh My Zsh — A Critique

## What Problem This Solves

Oh My Zsh (OMZ) is the most popular Zsh framework. It's how most people first configure Zsh. But it comes with costs that most users don't understand. This isn't about hating OMZ — it's about understanding what it does so you can make an informed choice.

## What OMZ Actually Does

When you install OMZ, it:

1. Replaces your `~/.zshrc` with its own
2. Sources `~/.oh-my-zsh/oh-my-zsh.sh` which:
   - Sets dozens of shell options
   - Loads a theme (prompt)
   - Loads plugins (each sourcing more files)
   - Initializes completions
3. Adds 50+ files to source during startup

## The Problems

### 1. Startup Time

```bash
# Measure your shell startup:
time zsh -i -c exit

# Typical results:
# Clean Zsh:     ~30ms
# OMZ (minimal): ~200ms
# OMZ (typical): ~400-800ms
# OMZ (heavy):   ~1-3 seconds

# Every time you open a terminal, split a pane, or run a subshell.
```

**Profile what's slow:**

```bash
# Add to top of ~/.zshrc:
zmodload zsh/zprof

# Add to bottom of ~/.zshrc:
zprof

# Now open a new terminal and see what takes the longest.
```

### 2. You Don't Understand Your Shell

OMZ sets dozens of options and aliases you didn't ask for:

```bash
# Check what OMZ aliased:
alias | wc -l    # Often 100+ aliases

# Some surprising defaults:
alias | grep -E '^(l|g|d|k)'
# l → ls -lah
# la → ls -lAh
# ll → ls -lh
# g → git
# ga → git add
# gc → git commit
# ... dozens of git aliases

# Are these YOUR choices? Do you know what each does?
```

### 3. Plugin Quality Varies

```bash
# OMZ has 300+ plugins. Most just add aliases:
cat ~/.oh-my-zsh/plugins/docker/docker.plugin.zsh | head -20
# It's mostly alias definitions

# Some plugins conflict with each other.
# Some plugins are unmaintained.
# Some add functions you'll never use.
# Each one adds to startup time.
```

### 4. Updates Can Break Things

OMZ auto-updates by default. A random update can change your environment:
- Aliases change meaning
- Options get added/removed
- Plugin behavior changes
- Themes change prompt format

## What to Do Instead

### Option A: Use OMZ Consciously

If you keep OMZ, be deliberate:

```bash
# In .zshrc, MINIMIZE plugins:
plugins=(
    git             # Only if you USE these aliases
    docker          # Only if you USE these aliases
    # That's it. Not 15 plugins.
)

# Disable auto-update if stability matters:
DISABLE_AUTO_UPDATE="true"

# Know what your plugins do:
cat ~/.oh-my-zsh/plugins/git/git.plugin.zsh | head -50
```

### Option B: Build Your Own Config

Replace OMZ with ~50 lines in `.zshrc`:

```bash
# ~/.zshrc — Everything you need, nothing you don't

# ── History ─────────────────────────────────────
HISTFILE=~/.zsh_history
HISTSIZE=50000
SAVEHIST=50000
setopt SHARE_HISTORY HIST_IGNORE_ALL_DUPS HIST_REDUCE_BLANKS

# ── Navigation ──────────────────────────────────
setopt AUTO_CD AUTO_PUSHD PUSHD_IGNORE_DUPS
setopt CORRECT
setopt GLOB_DOTS     # Include dotfiles in globs

# ── Completion ──────────────────────────────────
autoload -Uz compinit
if [[ -n ~/.zcompdump(#qN.mh+24) ]]; then
    compinit
else
    compinit -C
fi
zstyle ':completion:*' matcher-list 'm:{a-z}={A-Z}'
zstyle ':completion:*' menu select
zstyle ':completion:*' list-colors ${(s.:.)LS_COLORS}

# ── Prompt ──────────────────────────────────────
autoload -Uz vcs_info
precmd() { vcs_info }
zstyle ':vcs_info:git:*' formats '%F{green}%b%f'
setopt PROMPT_SUBST
PROMPT='%F{blue}%~%f ${vcs_info_msg_0_} %F{yellow}$%f '

# ── Key Bindings ────────────────────────────────
bindkey -e                                          # Emacs mode
bindkey '^[[A' history-beginning-search-backward    # Up
bindkey '^[[B' history-beginning-search-forward     # Down

# ── Aliases (YOUR choices) ──────────────────────
alias ls='ls --color=auto'
alias ll='ls -lah'
alias grep='grep --color=auto'
alias ..='cd ..'
alias ...='cd ../..'

# ── PATH ────────────────────────────────────────
typeset -U PATH
path=(~/.local/bin $path)
```

Time: ~30ms startup. You understand every line.

### Option C: Lightweight Frameworks

If you want a framework but lighter than OMZ:

| Framework | Startup | Philosophy |
|-----------|---------|------------|
| Oh My Zsh | 200-800ms | Batteries included |
| Prezto | 100-200ms | Curated configuration |
| zinit | 50-100ms | Plugin manager with lazy loading |
| zplug | 100-200ms | Plugin manager |
| None | 20-40ms | Full control |

**zinit example** (lazy-loading plugins):

```bash
# Install zinit:
bash -c "$(curl --fail --show-error --silent --location https://raw.githubusercontent.com/zdharma-continuum/zinit/HEAD/scripts/install.sh)"

# In .zshrc — plugins load ASYNCHRONOUSLY:
zinit light zsh-users/zsh-autosuggestions
zinit light zsh-users/zsh-syntax-highlighting
zinit light zsh-users/zsh-completions
```

## Essential Plugins (If You Use Any)

Only 3 plugins are genuinely worth the overhead:

### 1. zsh-autosuggestions

Shows a ghost suggestion based on history as you type. Accept with → arrow.

```bash
# Install without framework:
git clone https://github.com/zsh-users/zsh-autosuggestions ~/.zsh/zsh-autosuggestions
source ~/.zsh/zsh-autosuggestions/zsh-autosuggestions.zsh
```

### 2. zsh-syntax-highlighting

Colors your command as you type — red for errors, green for valid commands.

```bash
git clone https://github.com/zsh-users/zsh-syntax-highlighting ~/.zsh/zsh-syntax-highlighting
source ~/.zsh/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh
# MUST be last thing sourced in .zshrc
```

### 3. zsh-completions

Extra completion definitions for many commands.

```bash
git clone https://github.com/zsh-users/zsh-completions ~/.zsh/zsh-completions
fpath=(~/.zsh/zsh-completions/src $fpath)
```

That's it. Three plugins. Everything else is aliases you can write yourself.

## Prompt: Theme vs DIY

### OMZ Themes

```bash
# OMZ themes are just PROMPT variable with colors:
ZSH_THEME="robbyrussell"    # The default

# Most themes add:
# - Current directory
# - Git branch/status
# - Username/hostname
# - Time
# All things you can do in 5 lines
```

### Starship (Recommended Alternative)

```bash
# Install on Fedora:
sudo dnf install starship

# Add to end of .zshrc:
eval "$(starship init zsh)"

# Configure in ~/.config/starship.toml:
[character]
success_symbol = "[❯](green)"
error_symbol = "[❯](red)"

[directory]
truncation_length = 3

[git_branch]
format = "[$symbol$branch]($style) "
```

Starship is fast (written in Rust), cross-shell (works in Bash, Zsh, Fish), and highly configurable.

### DIY Prompt (Simplest)

```bash
# Git branch in prompt — no framework needed:
autoload -Uz vcs_info
precmd() { vcs_info }
zstyle ':vcs_info:git:*' formats '(%b)'
setopt PROMPT_SUBST

# Simple and informative:
PROMPT='%F{blue}%~%f %F{green}${vcs_info_msg_0_}%f
$ '
```

## The Decision Framework

```
Do you understand what OMZ does?
├── No → Read this file. Decide again.
├── Yes, and I use <5 plugins → Keep OMZ, it's fine
├── Yes, and my shell is slow → Switch to manual config or zinit
└── Yes, and I want full control → Manual config (Option B)

Are you setting up a new machine?
├── Priority: get started fast → OMZ with 2-3 plugins
└── Priority: understand your tools → Manual config
```

## Exercise

1. Time your current shell startup: `time zsh -i -c exit`. If it's over 200ms, profile it with `zprof`.

2. Run `alias | wc -l` to see how many aliases you have. For any you don't recognize, look them up and decide if you want them.

3. Try the "Option B" config: backup your `.zshrc`, replace it with the ~50 line version, and use it for a day. Note what you miss (if anything).

4. If you keep OMZ, reduce your plugins to only those you actually use daily. Measure the startup time before and after.

---

Next: [Level 9 — Fedora Reality](../09-fedora-reality/00-dnf-systemd-journalctl.md)
