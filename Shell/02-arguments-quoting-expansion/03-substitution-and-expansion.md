# Command Substitution & Brace Expansion

## What Problem This Solves

You see `$(...)` in commands and wonder how it works. You see `{1..10}` and think it's related to variables. These are expansion mechanisms — ways the shell generates text before running the command. Understanding them means you can compose commands dynamically instead of typing repetitive things.

## Command Substitution: `$(command)`

### The Mental Model

`$(command)` runs a command and replaces itself with the command's stdout output. Think of it as "inline execution."

```bash
echo "Today is $(date +%A)"
# 1. Shell runs: date +%A             → "Wednesday"
# 2. Substitution: "Today is Wednesday"
# 3. Shell runs: echo "Today is Wednesday"
```

The old syntax is backticks: `` `command` ``. Avoid it. `$()` is clearer, nestable, and less error-prone.

### Practical Examples

```bash
# Store command output in a variable:
kernel=$(uname -r)
echo "Kernel: $kernel"

# Use in-line:
echo "Host: $(hostname), User: $(whoami), PID: $$"

# Use in file paths:
logfile="/var/log/myapp-$(date +%Y%m%d).log"

# Use in conditionals:
if [[ $(whoami) != "root" ]]; then
    echo "Must run as root"
    exit 1
fi

# Nesting (try this with backticks — it's painful):
echo "Kernel version: $(uname -r | cut -d. -f1-2)"
```

### Command Substitution and Quoting

```bash
# ALWAYS quote command substitutions in assignments and arguments:
files="$(ls /tmp)"       # Preserves newlines in the variable
echo "$files"            # Prints with newlines

echo $(ls /tmp)          # Unquoted: newlines become spaces, word splitting happens
echo "$(ls /tmp)"        # Quoted: preserves formatting

# Command substitution strips trailing newlines:
output="$(printf 'hello\n\n\n')"
echo "$output"           # "hello" — trailing newlines gone!
```

### Process Substitution: `<(command)`

Not the same as command substitution! Process substitution creates a temporary file descriptor:

```bash
# Compare two directory listings:
diff <(ls /usr/bin) <(ls /usr/sbin)

# <(command) creates a /dev/fd/N pseudo-file that the command reads from
# This is NOT $() — it produces a file path, not text

# Use when a command requires a filename, not stdin:
paste <(cut -d: -f1 /etc/passwd) <(cut -d: -f7 /etc/passwd)
# Side-by-side: username and shell
```

## Brace Expansion: `{...}`

### The Mental Model

Brace expansion generates a list of strings. It happens **before** variable expansion — it's purely textual.

```bash
echo {a,b,c}
# Output: a b c

echo file.{txt,md,log}
# Output: file.txt file.md file.log

echo {1..5}
# Output: 1 2 3 4 5
```

### Two Types

**Comma-separated list:**
```bash
{a,b,c}           # a b c
{cat,dog,fish}     # cat dog fish
pre{a,b,c}suf      # preasuf prebsuf precsuf
```

**Sequence:**
```bash
{1..10}            # 1 2 3 4 5 6 7 8 9 10
{01..10}           # 01 02 03 04 05 06 07 08 09 10  (zero-padded!)
{a..z}             # a b c d ... x y z
{1..20..3}         # 1 4 7 10 13 16 19  (step of 3)
```

### Real Uses

```bash
# Create multiple directories at once:
mkdir -p project/{src,tests,docs,build}
# Creates: project/src  project/tests  project/docs  project/build

# Backup a file (classic pattern):
cp config.yaml{,.bak}
# Expands to: cp config.yaml config.yaml.bak

# Rename file extension:
mv report.{txt,md}
# Expands to: mv report.txt report.md

# Generate multiple paths:
echo /var/log/{messages,secure,cron}
# /var/log/messages /var/log/secure /var/log/cron

# Nested braces:
echo {a,b}{1,2}
# a1 a2 b1 b2 (cartesian product)

# Create a date-stamped directory tree:
mkdir -p logs/2024/{01..12}/{01..31}
# Creates logs/2024/01/01 through logs/2024/12/31
```

### Brace Expansion vs Glob Expansion

```bash
# Brace expansion generates text even if files don't exist:
echo {a,b,c}.txt       # a.txt b.txt c.txt  (always, regardless of files)

# Glob expansion only shows existing files:
echo *.txt             # Only .txt files that actually exist

# Brace expansion happens BEFORE glob expansion:
ls {*.txt,*.md}        # First: *.txt *.md  Then each is glob-expanded
```

## Arithmetic Expansion: `$((...))`

```bash
echo $((2 + 3))           # 5
echo $((10 / 3))          # 3 (integer division!)
echo $((10 % 3))          # 1 (modulo)
echo $((2 ** 10))         # 1024 (exponentiation)

# With variables (no $ needed inside):
x=5
echo $((x * 2))           # 10
echo $((x + 1))           # 6

# Increment:
count=0
((count++))                # count is now 1

# Conditional (ternary):
a=5; b=10
echo $(( a > b ? a : b )) # 10 (max of a and b)
```

## Tilde Expansion

```bash
~              # /home/yourusername
~root          # /root
~+             # $PWD (current directory)
~-             # $OLDPWD (previous directory)

echo ~         # /home/user
echo "~"       # ~ (literal — tilde doesn't expand in quotes!)
echo ~/docs    # /home/user/docs
```

## The Expansion Order (Complete)

The shell processes expansions in this exact order:

```
1. Brace expansion        {a,b,c}  →  a b c
2. Tilde expansion         ~  →  /home/user
3. Parameter expansion     $var  →  value
4. Command substitution    $(cmd)  →  output
5. Arithmetic expansion    $((1+2))  →  3
6. Process substitution    <(cmd)  →  /dev/fd/63
7. Word splitting          (on unquoted results of 3-5)
8. Pathname expansion      *.txt  →  file1.txt file2.txt
9. Quote removal           "quotes" → quotes
```

This order matters. Brace expansion happens first, so `{$a,$b}` expands the braces *before* the variables:

```bash
a=hello
b=world
echo {$a,$b}      # First brace: $a $b  Then parameter: hello world
echo ${a}_{b}     # This is parameter expansion, not brace expansion!
```

## Common Footguns

**Footgun 1: Braces inside quotes don't expand**
```bash
echo "{a,b,c}"    # {a,b,c}  — literal, no expansion
echo {a,b,c}      # a b c    — expanded
```

**Footgun 2: Spaces inside braces break them**
```bash
echo {a, b, c}    # {a, b, c}  — literal! Spaces broke the brace expansion
echo {a,b,c}      # a b c      — works
```

**Footgun 3: Variables in brace sequences don't work**
```bash
n=5
echo {1..$n}      # {1..5}  — literal! Brace expansion happens before variable expansion
# Workaround:
eval echo "{1..$n}"    # Works but eval is dangerous
seq 1 "$n"             # Better alternative
```

**Footgun 4: Unquoted command substitution eats newlines**
```bash
# Command output has newlines:
output=$(printf 'line1\nline2\nline3')

# Unquoted — newlines become spaces:
echo $output       # line1 line2 line3

# Quoted — newlines preserved:
echo "$output"
# line1
# line2
# line3
```

## Exercise

1. Create a project skeleton in one command:
   ```bash
   mkdir -p myproject/{src/{main,test},docs,build,config}
   tree myproject
   ```

2. Use `cp` with brace expansion to back up three config files.

3. Calculate disk usage percentage:
   ```bash
   used=$(df / --output=pcent | tail -1 | tr -d ' %')
   echo "Root partition: $used% used"
   echo "Remaining: $((100 - used))%"
   ```

4. Compare `/etc/passwd` to `/etc/group` using process substitution and `diff`.

---

Next: [Why "$@" Exists](04-why-dollar-at-exists.md)
