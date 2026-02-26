# Shell Variables vs Environment Variables

## What Problem This Solves

You set a variable in your terminal, then try to use it in a script or a program you call — and it's not there. Or you `export` everything and wonder why. The distinction between shell variables and environment variables is fundamental to how Unix processes communicate.

## How People Misunderstand It

1. **"Shell variables and environment variables are the same thing"** — Shell variables are local to the shell process. Environment variables are inherited by child processes. `export` bridges the gap.
2. **"export makes a variable permanent"** — It only lasts for the current session. It's inherited by children, but it's gone when you close the terminal.
3. **"Programs can access all my shell variables"** — Programs only see environment variables, never shell-local variables.

## The Mental Model

Every process on Unix has two variable spaces:

```
Shell Process
┌──────────────────────────────────────┐
│  Shell Variables (local)              │
│  ┌──────────────────────────────┐    │
│  │  foo="hello"                  │    │  ← Only this shell can see this
│  │  count=42                     │    │
│  └──────────────────────────────┘    │
│                                      │
│  Environment Variables (exported)    │
│  ┌──────────────────────────────┐    │
│  │  HOME=/home/user              │    │  ← Inherited by child processes
│  │  PATH=/usr/bin:/usr/local/bin  │    │
│  │  LANG=en_US.UTF-8             │    │
│  │  MY_VAR=exported_value        │    │
│  └──────────────────────────────┘    │
│                                      │
│  When shell runs a command (fork+exec)│
│  → child gets a COPY of the env vars │
│  → child does NOT get shell variables │
└──────────────────────────────────────┘
       │
       ├── fork+exec: grep "pattern" file
       │   └── grep receives: HOME, PATH, LANG, MY_VAR
       │       grep does NOT receive: foo, count
       │
       └── fork+exec: bash script.sh
           └── script.sh receives: HOME, PATH, LANG, MY_VAR
               script.sh does NOT receive: foo, count
```

## Seeing the Difference

```bash
# Set a shell variable (NOT exported):
greeting="hello"

# Set and export an environment variable:
export name="Alice"

# See all environment variables:
env | head -10         # Or: printenv

# See all shell variables (including non-exported):
set | head -20         # Shows EVERYTHING: vars, functions, etc.

# Check if a specific variable is exported:
declare -p greeting    # declare -- greeting="hello"  (no -x = not exported)
declare -p name        # declare -x name="Alice"       (-x = exported)

# Test: does a child process see it?
echo "$greeting"       # hello  (current shell sees it)
bash -c 'echo "$greeting"'   # (empty — child doesn't see shell variables)

echo "$name"           # Alice  (current shell sees it)
bash -c 'echo "$name"'       # Alice  (child sees exported variables)
```

## export Explained Properly

`export` doesn't "send" a variable somewhere. It marks the variable so that when the kernel creates a child process (fork+exec), the variable is included in the child's environment.

```bash
# These are equivalent:
export MY_VAR="value"

# Same as:
MY_VAR="value"
export MY_VAR

# Unexport (remove from environment, keep as shell variable):
export -n MY_VAR    # Now it's just a shell variable again
```

### Key Properties

```bash
# 1. export is ONE-WAY (parent → child, never child → parent):
export MY_VAR="original"
bash -c 'MY_VAR="changed"; echo "Child: $MY_VAR"'
echo "Parent: $MY_VAR"
# Child: changed
# Parent: original  ← Parent is unaffected!

# 2. Children get a COPY, not a reference:
export COUNTER=0
bash -c 'COUNTER=100'
echo "$COUNTER"    # Still 0

# 3. Changes to exported vars propagate to FUTURE children:
export COLOR="red"
bash -c 'echo $COLOR'    # red
COLOR="blue"              # Change it (it stays exported)
bash -c 'echo $COLOR'    # blue
```

## The VAR=value command Pattern

You can set a variable for just one command without affecting the current shell:

```bash
# Set LANG only for this one sort command:
LANG=C sort file.txt

# After this line, LANG is unchanged in your shell
echo "$LANG"    # Still en_US.UTF-8 (or whatever it was)

# Multiple variables:
CC=gcc CFLAGS="-O2 -Wall" make

# This is commonly used for:
TZ=UTC date                          # Show UTC time
EDITOR=nano crontab -e               # Use nano for crontab
PAGER=cat man ls                     # No paging for man page
LC_ALL=C grep '[A-Z]' file.txt       # Use C locale for predictable sorting
```

### How It Works

The shell parses `VAR=value command` and:
1. Forks a child process
2. Sets `VAR=value` in the child's environment
3. Execs the command in the child
4. The parent's environment is untouched

```bash
# This is roughly equivalent to:
(export VAR=value; exec command)

# But it's NOT the same as:
export VAR=value; command    # This modifies the parent shell!
```

## Important Environment Variables

```bash
# System:
HOME        # User's home directory
USER        # Current username
SHELL       # Default login shell
PATH        # Command search path
LANG        # Locale (language/encoding)
TERM        # Terminal type
PWD         # Current working directory
OLDPWD      # Previous working directory

# Shell-specific:
BASH_VERSION   # Bash version
PS1         # Primary prompt
PS2         # Continuation prompt (for multi-line commands)
HISTFILE    # History file location
HISTSIZE    # Number of commands in memory
IFS         # Input Field Separator

# Commonly used:
EDITOR      # Default text editor
VISUAL      # Visual editor (for terminal editors)
PAGER       # Default pager (less, more)
TZ          # Timezone
http_proxy  # HTTP proxy URL
no_proxy    # Hosts to bypass proxy
```

## Common Footguns

**Footgun 1: Expecting child processes to change parent variables**
```bash
# This is physically impossible:
get_result() {
    result="42"    # This sets it in the FUNCTION's scope (same shell = ok for functions)
}
get_result
echo "$result"    # 42 — works because functions run in the current shell

# But this does nothing:
bash -c 'result=42'    # Sets it in the child. Parent never sees it.
echo "$result"          # Empty
```

**Footgun 2: export in a subshell**
```bash
# Subshell (parentheses) = child process:
(export MY_VAR="set in subshell")
echo "$MY_VAR"    # Empty — subshell's changes are lost

# This is also a subshell:
echo "hello" | while read -r word; do
    MY_VAR="$word"    # Set in pipe subshell — lost!
done
echo "$MY_VAR"    # Empty
```

**Footgun 3: .env files don't auto-load**
```bash
# .env files are NOT magic. They're just text files.
# This does NOT work:
cat .env          # DATABASE_URL=postgres://localhost/mydb

# The shell doesn't read .env files automatically. You must source them:
source .env       # Or: . .env
# Now $DATABASE_URL is set in the current shell

# But be careful: sourcing untrusted .env files is dangerous!
# They can contain: rm -rf /; DATABASE_URL=x
# Only source .env files you trust.

# For docker-compose and similar tools, .env is read by the TOOL, not the shell.
```

## Exercise

1. Set a shell variable and an exported variable. Run `bash -c 'echo shell=$myshellvar env=$myenvvar'` to prove which one the child process sees.

2. Use `VAR=value cmd` to run `date` in UTC timezone without changing your shell's timezone.

3. Source a `.env` file and verify the variables are set:
   ```bash
   echo 'DB_HOST=localhost' > /tmp/test.env
   echo 'DB_PORT=5432' >> /tmp/test.env
   source /tmp/test.env
   echo "$DB_HOST:$DB_PORT"
   ```

4. Demonstrate that `export` is one-way: export a variable, change it in a child process, show the parent still has the original value.

---

Next: [Subshells, Export, and Scope](01-subshells-and-scope.md)
