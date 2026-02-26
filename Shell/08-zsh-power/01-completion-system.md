# The Zsh Completion System

## What Problem This Solves

Tab completion is the #1 interactive productivity feature. Zsh's completion system is absurdly powerful — it can complete command flags, remote hostnames, git branches, docker containers, kubernetes resources, and more. But most people use the default or Oh My Zsh's config without understanding what's happening.

## The Basics

```bash
# Enable the completion system:
autoload -Uz compinit && compinit
```

`autoload -Uz compinit` loads the function lazily. `compinit` initializes the completion system, scanning for completion definitions.

```bash
# What this enables:
ls -<TAB>           # Shows all ls flags with descriptions
git ch<TAB>         # Completes to checkout/cherry-pick/etc
ssh <TAB>           # Completes hostnames from ~/.ssh/config
kill <TAB>          # Completes process IDs with names
cd <TAB>            # Shows directories with previews
```

## Configuring with zstyle

Completions are configured with `zstyle`, which uses a pattern-based configuration system:

```bash
zstyle ':completion:CONTEXT:COMPLETER:COMMAND:ARGUMENT:TAG' style value
```

You usually use wildcards (`*`) for most parts:

```bash
# Case-insensitive matching:
zstyle ':completion:*' matcher-list 'm:{a-zA-Z}={A-Za-z}'

# Allow partial completion:
zstyle ':completion:*' matcher-list '' 'm:{a-zA-Z}={A-Za-z}' '+l:|=* r:|=*'
# This means: exact → case-insensitive → substring match

# Arrow-key menu selection:
zstyle ':completion:*' menu select

# Colored completion list (uses LS_COLORS):
zstyle ':completion:*' list-colors ${(s.:.)LS_COLORS}

# Group completions by type:
zstyle ':completion:*' group-name ''
zstyle ':completion:*:descriptions' format '%F{yellow}── %d ──%f'
zstyle ':completion:*:messages' format '%F{purple}── %d ──%f'
zstyle ':completion:*:warnings' format '%F{red}── no matches ──%f'

# Cache completions (speeds up expensive completions like apt/dnf):
zstyle ':completion:*' use-cache yes
zstyle ':completion:*' cache-path "$HOME/.zsh/cache"
```

### Practical Recipes

```bash
# Complete PIDs with process names:
zstyle ':completion:*:*:kill:*' menu yes select
zstyle ':completion:*:kill:*' force-list always
zstyle ':completion:*:*:kill:*' command 'ps -u $USER -o pid,%cpu,tty,args'

# SSH/SCP completion from ~/.ssh/config:
zstyle ':completion:*:(ssh|scp|rsync):*' hosts ${${(f)"$(cat ~/.ssh/config 2>/dev/null | grep '^Host ' | grep -v '[*?]')"}#Host }

# Docker container completion:
# (Usually auto-detected if docker completion is installed)

# Ignore certain directories in completion:
zstyle ':completion:*' ignored-patterns '*.pyc' '__pycache__' 'node_modules'

# Fuzzy completion (allow 1 error):
zstyle ':completion:*' completer _complete _correct _approximate
zstyle ':completion:*:approximate:*' max-errors 1 numeric
```

## How Completion Functions Work

Every command's completion is defined by a function. When you press TAB after `git`, Zsh calls `_git`. These functions use a DSL for describing what to complete:

```bash
# See what completion function handles a command:
whence -w _git     # _git: function

# See all loaded completions:
echo ${(k)_comps}

# Check which function handles 'docker':
echo $_comps[docker]    # _docker
```

### Writing a Simple Completion

```bash
# For a command 'myapp' with subcommands:
_myapp() {
    local -a subcmds
    subcmds=(
        'start:Start the application'
        'stop:Stop the application'
        'status:Show current status'
        'deploy:Deploy to production'
    )
    _describe 'command' subcmds
}
compdef _myapp myapp

# Now: myapp <TAB> shows:
# start   -- Start the application
# stop    -- Stop the application
# status  -- Show current status
# deploy  -- Deploy to production
```

### Completion with Flags

```bash
_myapp() {
    _arguments \
        '-v[Enable verbose output]' \
        '-p[Port number]:port:_ports' \
        '-f[Config file]:file:_files -g "*.conf"' \
        '--dry-run[Show what would happen]' \
        '1:command:(start stop status deploy)' \
        '*:file:_files'
}
compdef _myapp myapp
```

The `_arguments` function is the workhorse — it handles flags, positional args, and their completions.

## Useful Built-in Completers

```bash
_files              # File completion
_directories        # Directory completion
_users              # Username completion
_hosts              # Hostname completion
_ports              # Port completion
_pids               # Process ID completion
_services           # Systemd service completion
_parameters         # Shell variable completion
_command_names      # Command name completion
```

## Debugging Completions

```bash
# See what's happening during completion:
# Press Ctrl+X then h after a partial command to see the completion context

# Or enable verbose completion debugging:
zstyle ':completion:*' verbose yes

# See which function handles completion:
echo $_comps[git]

# Reload completions after editing:
unfunction _myapp 2>/dev/null; compinit

# Or just:
exec zsh
```

## Speed Optimization

The completion system can feel slow if not configured well:

```bash
# 1. Use caching:
zstyle ':completion:*' use-cache yes
zstyle ':completion:*' cache-path "$HOME/.zsh/cache"
mkdir -p "$HOME/.zsh/cache"

# 2. Only rebuild completion dump daily:
autoload -Uz compinit
if [[ -n ~/.zcompdump(#qN.mh+24) ]]; then
    compinit
else
    compinit -C    # Skip security check (faster)
fi

# 3. Lazy-load completions for slow commands:
# (Some frameworks do this automatically)
```

## Completion vs Other Tab Features

```bash
# Tab for completion:
git ch<TAB>          # Completion

# Tab-Tab for listing all possibilities:
git <TAB><TAB>       # Show all git subcommands

# Ctrl+X h for completion debugging

# Ctrl+X ? for show completion info

# In menu mode (with menu select):
# TAB/arrows to navigate
# Enter to accept
# Ctrl+G to cancel
```

## Exercise

1. Add case-insensitive completion and menu selection to your `.zshrc`. Test with a command like `ls /USR/L<TAB>`.

2. Write a completion function for a script that has `start`, `stop`, and `restart` subcommands, each with a `--force` flag.

3. Configure SSH completion to work from your `~/.ssh/config` file. Verify with `ssh <TAB>`.

4. Profile your completion startup time: time `exec zsh` with and without `compinit -C`.

---

Next: [Oh My Zsh — A Critique](02-oh-my-zsh-critique.md)
