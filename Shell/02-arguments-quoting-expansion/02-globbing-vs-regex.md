# Globbing vs Regex

## What Problem This Solves

People use `*` in `ls *.txt` and `*` in `grep '.*error'` and think they're the same thing. They're not. Globbing (shell patterns) and regex (regular expressions) are completely different languages that happen to share some characters. Conflating them causes subtle, hard-to-find bugs.

## How People Misunderstand It

1. **"* means the same thing everywhere"** — In globs, `*` means "any characters." In regex, `*` means "zero or more of the previous thing." These are completely different.
2. **"regex is the shell's pattern matching"** — The shell uses globs, not regex. `grep` and `sed` use regex. You must know which tool uses which.
3. **"I'll just use regex everywhere"** — Regex is overkill for simple filename matching. Globs exist for a reason.

## The Mental Model

**Globs** are the shell's built-in pattern language for matching **filenames**.
**Regex** is a separate pattern language used by **programs** (grep, sed, awk) for matching **text**.

The shell expands globs *before* the command runs. Regex is passed to the command, which interprets it itself.

```
ls *.txt
   ↑ Shell expands this glob, then passes resulting filenames to ls

grep '.*\.txt' file
      ↑ Shell passes this STRING to grep. Grep interprets the regex.
```

### Side-by-Side Comparison

| I want to match... | Glob | Regex |
|-------------------|------|-------|
| Any characters | `*` | `.*` |
| Any single character | `?` | `.` |
| One of specific chars | `[abc]` | `[abc]` |
| Literal dot | `.` (it's already literal) | `\.` (must escape) |
| Zero or more of X | N/A | `X*` |
| One or more of X | N/A | `X+` |
| Start of string | N/A (globs match whole name) | `^` |
| End of string | N/A (globs match whole name) | `$` |

### Critical Difference: `*`

```bash
# GLOB: * matches any sequence of characters (in filenames)
ls *.log           # All files ending in .log

# REGEX: * means "zero or more of the previous character"
grep 'ab*c' file   # Matches: ac, abc, abbc, abbbc, ...
                    # The * applies to 'b', not "anything"

# REGEX: .* means "zero or more of any character" (this is the glob * equivalent)
grep '.*error' file  # Matches: "error", "some error", "123error", etc.
```

## Globs in Detail

### Basic Globs

```bash
*           # Match any characters (including none), except / in pathnames
?           # Match exactly one character
[abc]       # Match one of: a, b, or c
[a-z]       # Match one character in range a-z
[!abc]      # Match one character NOT a, b, or c (also [^abc] in some contexts)
```

### How Globs Work

```bash
echo *.txt
# 1. Shell scans the CURRENT DIRECTORY for files matching *.txt
# 2. Replaces *.txt with the list of matching filenames
# 3. echo receives those filenames as separate arguments

# If no files match:
#   Bash: passes the literal string "*.txt" (what?!)
#   Zsh: error "no matches found" (safer)

# Bash nullglob option changes this:
shopt -s nullglob  # No match → empty (the glob disappears)
echo *.xyz         # If no .xyz files: prints nothing
```

### Extended Globs (Bash)

```bash
# Enable extended globs in Bash:
shopt -s extglob

# Now you can use:
?(pattern)     # Match zero or one occurrence
*(pattern)     # Match zero or more occurrences
+(pattern)     # Match one or more occurrences
@(pattern)     # Match exactly one occurrence
!(pattern)     # Match anything EXCEPT the pattern

# Examples:
ls !(*.log)           # All files EXCEPT .log files
ls @(*.txt|*.md)      # .txt or .md files
ls *(.[ch])           # .c or .h files (zero or more matches of .[ch])
```

### Zsh Glob Qualifiers

Zsh has much more powerful globbing:

```zsh
# Recursive glob:
echo **/*.txt         # All .txt files in all subdirectories

# Glob qualifiers (Zsh only):
ls -d *(/)            # Only directories
ls *(.)               # Only regular files
ls *(.x)              # Only executable files
ls *(mh-1)            # Files modified in the last hour
ls *(Lm+10)           # Files larger than 10MB
```

## When to Use Globs vs Regex

```
Use GLOBS when:
  → Matching filenames on the command line
  → Shell patterns in case statements
  → find -name 'pattern' (find uses globs, NOT regex)
  → rsync include/exclude patterns

Use REGEX when:
  → Searching text content (grep, sed, awk)
  → Pattern matching in [[ $string =~ regex ]] (Bash)
  → Complex text transformation (sed 's/regex/replacement/')
  → Programming languages (Python, JS, Go, etc.)
```

## Common Footguns

**Footgun 1: Using regex syntax in `find -name`**
```bash
# WRONG — find uses globs, not regex:
find . -name '.*\.txt'    # This looks for files starting with dot, any char, .txt
                           # The .* is a glob, not regex!

# RIGHT — glob syntax:
find . -name '*.txt'       # All .txt files

# If you really want regex in find:
find . -regex '.*\.txt$'   # Use -regex flag explicitly
```

**Footgun 2: Forgetting to quote globs passed to commands**
```bash
# BROKEN — shell expands the glob BEFORE grep sees it:
grep *.txt /var/log/messages
# Shell expands *.txt to matching filenames in current directory
# grep gets weird arguments

# CORRECT — quote it so grep gets the literal pattern:
grep '*.txt' /var/log/messages    # But this is a bad regex too...
grep '.*\.txt' /var/log/messages  # Correct regex for matching .txt
```

**Footgun 3: `*` doesn't match hidden files**
```bash
ls *           # Does NOT show .bashrc, .config, etc.
ls .*          # Shows ONLY hidden files (plus . and ..)
ls .* *        # Shows everything (but includes . and ..)

# Bash option to include hidden files in *:
shopt -s dotglob
ls *           # Now includes hidden files
```

**Footgun 4: Glob in a variable**
```bash
pattern="*.txt"
ls $pattern      # Works (shell expands the glob after variable expansion)
ls "$pattern"    # Literal "*.txt" — no glob expansion inside quotes
```

## Exercise

1. List all `.conf` files in `/etc/` using a glob: `ls /etc/*.conf`
2. Try `find /etc -name '*.conf'` (glob) vs `find /etc -regex '.*\.conf$'` (regex). Both work but for different reasons.
3. In a directory with files, compare:
   ```bash
   echo *
   echo '.*'
   echo ".*"
   ```
   Predict which ones expand, then verify.
4. Use `[[ "hello123" =~ ^[a-z]+[0-9]+$ ]] && echo match` — this is regex in Bash. Now try `[[ "hello123" == hello* ]] && echo match` — this is a glob. Same result, different pattern languages.

---

Next: [Command Substitution & Brace Expansion](03-substitution-and-expansion.md)
