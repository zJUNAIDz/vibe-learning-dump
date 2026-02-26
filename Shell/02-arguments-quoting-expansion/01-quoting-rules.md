# Quoting Rules: Single, Double, None

## What Problem This Solves

You see commands with a mix of quotes — sometimes single, sometimes double, sometimes backticks, sometimes none. It feels arbitrary. It's not. Each quoting style controls exactly which expansions the shell performs.

## How People Misunderstand It

1. **"Single and double quotes are interchangeable"** — They are not. `'$HOME'` is the literal string `$HOME`. `"$HOME"` is `/home/user`.
2. **"Quotes are optional if there are no spaces"** — Even without spaces, unquoted variables undergo glob expansion. `file="*.txt"; echo $file` will expand globs.
3. **"I'll just use double quotes everywhere"** — Good default, but you need single quotes when you want to prevent ALL expansion (regex patterns, `awk` programs, `find` arguments).

## The Mental Model

There are exactly **four quoting contexts**:

```
No quotes    →  Everything expands. Word splitting happens. Globs expand.
"double"     →  Variables and $() expand. NO word splitting. NO globbing.
'single'     →  NOTHING expands. Completely literal.
$'ansi'      →  ANSI escape sequences expand (\n, \t, \\). No variable expansion.
```

### Decision Matrix

| I want... | Use |
|-----------|-----|
| A variable's value as one word | `"$var"` |
| A literal string, no expansion | `'literal'` |
| A string with both literal and variable parts | `"literal $var more"` |
| A regex or awk program | `'pattern'` (single quotes prevent shell interpretation) |
| A string with escape chars like tab/newline | `$'line1\nline2'` |
| Intentional word splitting (rare) | `$var` (unquoted) |

## Double Quotes: The Default

Double quotes should be your default. They protect against word splitting and globbing while still allowing useful expansions:

```bash
name="Alice"
dir="/home/alice/my projects"

# These expand inside double quotes:
echo "$name"                    # Alice
echo "Home: $HOME"             # Home: /home/alice
echo "Today: $(date +%Y-%m-%d)"  # Today: 2026-02-26
echo "Sum: $((2 + 3))"         # Sum: 5

# These do NOT expand inside double quotes:
echo "Price: $5"               # Hmm — $5 is a positional parameter (usually empty)
echo "Glob: *.txt"             # Glob: *.txt  (literal, no expansion)
echo "No split: $dir"          # No split: /home/alice/my projects (one word)
```

### Embedding Quotes

```bash
# Double quotes inside single quotes:
echo 'He said "hello"'         # He said "hello"

# Single quotes inside double quotes:
echo "It's working"            # It's working

# Double quotes inside double quotes:
echo "He said \"hello\""       # He said "hello"  (escape with \)

# Dollar signs in double quotes:
echo "Price: \$5"              # Price: $5  (escape with \)
echo 'Price: $5'               # Price: $5  (single quotes → no expansion)
```

## Single Quotes: The Lockdown

Single quotes make everything literal. Nothing expands. Nothing is special except the single quote itself.

```bash
echo '$HOME is where the heart is'    # $HOME is where the heart is
echo '$(rm -rf /)'                    # $(rm -rf /) — safe! Not executed!
echo 'Backslash: \'                   # Error! You can't escape inside single quotes

# How to include a single quote in single-quoted string:
echo 'It'\''s working'               # End quote, escaped quote, start quote
echo "It's working"                   # Or just use double quotes
echo $'It\'s working'                 # Or use ANSI quotes
```

### When Single Quotes Are Essential

```bash
# awk programs — $1, $2 are awk variables, not shell variables:
awk '{print $1, $3}' file.txt         # ✅ Single quotes
awk "{print $1, $3}" file.txt         # ❌ Shell expands $1 and $3 first!

# grep with regex — special chars are for grep, not the shell:
grep 'error\|warning' /var/log/messages    # ✅ Single quotes
grep "error\|warning" /var/log/messages    # ⚠️ Works in Bash (\ not special in "")
                                            #     but semantics are less clear

# find with -name:
find / -name '*.conf'           # ✅ Single quotes prevent shell glob expansion
find / -name *.conf             # ❌ Shell expands *.conf BEFORE find sees it

# cron expressions in scripts:
echo '0 * * * * /usr/bin/backup.sh'   # ✅ Literal
echo "0 * * * * /usr/bin/backup.sh"   # Works here too, but single is clearer
```

## No Quotes: The Danger Zone

Unquoted text undergoes full expansion: variable substitution, word splitting, AND glob expansion.

```bash
# Sometimes intentional:
files="file1.txt file2.txt file3.txt"
ls $files     # Intentionally split into 3 arguments

# Usually a bug:
path="/home/user/my documents"
ls $path      # BROKEN — splits into two arguments
```

**Rule: Use no quotes only when you consciously want word splitting or glob expansion.**

## ANSI Quoting: `$'...'`

For when you need literal special characters:

```bash
# Tab-separated values:
echo $'column1\tcolumn2\tcolumn3'

# Newlines:
echo $'line1\nline2\nline3'

# Single quote inside a string:
echo $'It\'s alive'

# Null byte (for xargs -0, etc.):
printf '%s\0' "$@"     # printf approach
echo $'hello\x00'      # ANSI approach (but echo may not handle \0)
```

## Nesting and Combining Quotes

```bash
# Variable inside a single-quoted-looking context:
echo "Files in '$dir' directory"     # Double quotes outer, single quotes are literal chars

# Building complex strings:
cmd="grep"
pattern='error|warning'
echo "Running: $cmd '$pattern'"      # Running: grep 'error|warning'

# Mixing quote styles:
echo 'Part one '"$variable"' part three'
# 'Part one '   →  literal
# "$variable"   →  expanded
# ' part three' →  literal
```

## Here Documents and Here Strings

```bash
# Here document — multiline input:
cat << 'EOF'
No $expansion here.
Everything is literal.
EOF

cat << EOF
Variables $expand here.
Command $(date) substitution too.
EOF

# Indented here-doc (uses <<- and tabs):
if true; then
    cat <<- EOF
	Indented content (leading TABS removed)
	Variable: $HOME
	EOF
fi

# Here string — single-line input:
grep "error" <<< "some error message"
tr '[:lower:]' '[:upper:]' <<< "hello"    # HELLO
```

## Common Footguns

**Footgun 1: Quoting around array expansion**
```bash
files=("file one.txt" "file two.txt" "file three.txt")

# WRONG — collapses array into single string:
echo "${files[*]}"    # "file one.txt file two.txt file three.txt"

# RIGHT — preserves each element:
for f in "${files[@]}"; do
    echo "File: $f"
done
# File: file one.txt
# File: file two.txt
# File: file three.txt
```

**Footgun 2: Quoting the wrong part of a command**
```bash
# WRONG — quotes include the command name:
"grep pattern file"    # Tries to run a program literally named "grep pattern file"

# RIGHT — quotes only around arguments that need them:
grep "pattern with spaces" "$file"
```

**Footgun 3: Nested command substitution quoting**
```bash
# This is fine — Bash handles nested quotes in $() correctly:
echo "Today is $(date "+%Y-%m-%d")"

# But old-style backticks DON'T handle nested quotes well:
echo "Today is `date "+%Y-%m-%d"`"   # May break depending on shell version
# Always prefer $() over backticks
```

## Exercise

1. Predict the output for each, then verify:
   ```bash
   var="hello world"
   echo $var
   echo "$var"
   echo '$var'
   echo "'$var'"
   echo '"$var"'
   ```

2. Explain why this `awk` command fails:
   ```bash
   column=3
   awk "{print $$column}" file.txt
   ```
   What does the shell do with `$$` and `$column` before awk sees it?

3. Create a file whose name contains a single quote (e.g., `it's_a_test.txt`). Write a script that deletes it safely.

4. Write a command that echoes a literal: `Price: $5 for "large" (or 'small')`

---

Next: [Globbing vs Regex](02-globbing-vs-regex.md)
