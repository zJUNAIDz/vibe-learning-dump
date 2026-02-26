# Shell Literacy Checklist

## How to Use This

This is your self-assessment. Go through each item honestly. If you can't explain it to someone else without looking it up, you don't know it yet — go back to the linked module.

Don't rush. This curriculum exists because shallow knowledge of shell is worse than no knowledge — it creates false confidence that leads to `rm -rf` disasters and "works on my machine" mysteries.

---

## Level 0: What a Shell Is

- [ ] I can explain the difference between a terminal emulator, a shell, and a TTY
- [ ] I know what `$TERM`, `$SHELL`, and `$0` contain and why they might differ
- [ ] I can list 3 concrete differences between Bash and Zsh
- [ ] I know that Bash arrays start at 0 and Zsh arrays start at 1
- [ ] I know which dotfiles load for login shells vs non-login shells (for both Bash and Zsh)
- [ ] I can explain why Fedora's `.bash_profile` sources `.bashrc`

**If not → [Level 0](../00-what-a-shell-is/)**

---

## Level 1: Commands Are Programs

- [ ] I can explain the command resolution order: aliases → functions → builtins → hash → `$PATH`
- [ ] I can use `type`, `which`, `command -v`, and `hash` — and I know which to trust
- [ ] I can name at least 5 shell builtins and explain why they *must* be builtins
- [ ] I understand that `sudo` runs a *new process* and why `sudo echo "x" > /protected/file` fails
- [ ] I know the `sudo tee` pattern and when to use `sudo bash -c`
- [ ] I can read a sudoers rule and understand NOPASSWD, command restrictions, and user specs

**If not → [Level 1](../01-commands-are-programs/)**

---

## Level 2: Arguments, Quoting, and Expansion

- [ ] I can explain what word splitting is and why unquoted `$var` is dangerous
- [ ] I know exactly what happens in each quoting context: none, `"double"`, `'single'`, `$'ansi'`
- [ ] I know the difference between globs (`*.txt`) and regex (`.*\.txt`) and never confuse them
- [ ] I can write the 9-step expansion order from memory (or reconstruct it from understanding)
- [ ] I understand `$()`, `<()`, `>()`, and `{}` brace expansion
- [ ] I can explain why `"$@"` exists and when to use it instead of `$*`

**If not → [Level 2](../02-arguments-quoting-expansion/)**

---

## Level 3: Pipes, Redirection, and File Descriptors

- [ ] I know that every process has fd 0 (stdin), fd 1 (stdout), fd 2 (stderr) by default
- [ ] I can explain why `2>&1` order matters and what `cmd > file 2>&1` vs `cmd 2>&1 > file` does
- [ ] I understand `noclobber` and the difference between `>`, `>>`, and `>|`
- [ ] I can explain that pipe stages run *concurrently* in *separate subshells*
- [ ] I know what `PIPESTATUS` / `pipestatus` is and why `set -o pipefail` matters
- [ ] I understand the pipe subshell problem (variable changes in pipe stages are lost)

**If not → [Level 3](../03-pipes-redirection-fds/)**

---

## Level 4: Exit Codes and Failure

- [ ] I know that exit codes are 0-255, where 0 = success and anything else = failure
- [ ] I can explain what signals 128+n exit codes mean (e.g., 137 = killed by SIGKILL)
- [ ] I understand `set -e` (errexit) and can name at least 3 situations where it *doesn't* work
- [ ] I know that `local var=$(failing_cmd)` hides the error and can explain why
- [ ] I can write a `trap` cleanup handler for EXIT, INT, and TERM
- [ ] I understand idempotency and why scripts should be safe to run twice

**If not → [Level 4](../04-exit-codes-failure/)**

---

## Level 5: Variables, Scope, and Environment

- [ ] I can explain the difference between a shell variable and an environment variable
- [ ] I know what `export` does and why child processes can't modify parent variables
- [ ] I can explain what a subshell is and list 5 things that create one
- [ ] I know the difference between `( )` (subshell) and `{ }` (current shell)
- [ ] I understand that systemd environment files are NOT shell scripts
- [ ] I can configure environment variables for a systemd service using `Environment=` and `EnvironmentFile=`

**If not → [Level 5](../05-variables-scope-env/)**

---

## Level 6: Bash as a Programming Language

- [ ] I can write functions that return data via stdout (not global variables)
- [ ] I know that `local` has its own exit code and the pattern `local x; x=$(cmd)` is safer
- [ ] I understand the three conditional systems: `[ ]` (test), `[[ ]]` (Bash), `(( ))` (arithmetic)
- [ ] I can write `while IFS= read -r line` loops and explain every part
- [ ] I know when to use process substitution `<()` instead of a pipe with `while read`
- [ ] I can write a complete script with: shebang, strict mode, logging, cleanup, arg parsing, and `main()`

**If not → [Level 6](../06-bash-scripting-language/)**

---

## Level 7: Text Processing

- [ ] I know when to use `grep` vs `sed` vs `awk` and don't use one where another fits better
- [ ] I can write `grep -E` (extended regex), `grep -P` (PCRE), and `grep -o` patterns
- [ ] I can write `sed` substitutions with capture groups: `sed 's/\(pattern\)/\1/'`
- [ ] I can write an awk program with `BEGIN`/`END` blocks, field processing, and associative arrays
- [ ] I can use `jq` to filter, transform, and construct JSON from shell scripts
- [ ] I can build multi-stage pipelines using `sort`, `uniq`, `cut`, `tr`, `paste`, `xargs`

**If not → [Level 7](../07-text-processing/)**

---

## Level 8: Zsh Power Features

- [ ] I know the key behavioral differences: word splitting, globbing, array indexing
- [ ] I can configure the Zsh completion system with `compinit` and `zstyle`
- [ ] I have a considered opinion on Oh My Zsh and can articulate its trade-offs
- [ ] I can write a functional `.zshrc` in under 50 lines that doesn't rely on a framework
- [ ] I know which plugins are actually worth using (autosuggestions, syntax-highlighting, completions)

**If not → [Level 8](../08-zsh-power/)**

---

## Level 9: Fedora System Administration

- [ ] I can use `dnf` for install, provides, history/undo, groups, and repo management
- [ ] I can use `systemctl` to manage services, read logs, create overrides, and set up timers
- [ ] I can use `journalctl` to filter by unit, priority, time range, and boot
- [ ] I understand SELinux contexts and can troubleshoot AVCs with `ausearch` and `restorecon`
- [ ] I know the 5 access layers (filesystem → DAC → capabilities → SELinux → firewall) and can debug each
- [ ] I can use `getfacl`/`setfacl`, `getcap`/`setcap`, and `firewall-cmd`

**If not → [Level 9](../09-fedora-reality/)**

---

## Level 10: Production-Quality Scripts

- [ ] My scripts validate all inputs before acting
- [ ] My scripts use `mktemp` for temporary files with `trap` cleanup
- [ ] My scripts use lock files (`flock` or `mkdir`) when they shouldn't run concurrently
- [ ] My scripts handle signals gracefully and clean up on exit
- [ ] My scripts support `--dry-run` for destructive operations
- [ ] I run ShellCheck on every script and fix all warnings
- [ ] I know the safe deletion pattern and never use `rm -rf "$unvalidated_var/"`
- [ ] I understand atomic operations (`mv` is atomic, multi-step copies are not)

**If not → [Level 10](../10-trustworthy-scripts/)**

---

## Level 11: Debugging

- [ ] I can use `bash -x` and custom `PS4` to trace script execution
- [ ] I can use `strace` to see what system calls a command makes
- [ ] I can systematically debug "works interactively, fails in script" problems
- [ ] I understand the environment differences between: interactive shell, scripts, cron, systemd, Docker, CI
- [ ] I can write scripts that work in all environments by normalizing PATH, locale, and assumptions
- [ ] I know the debug checklist: error message → exit code → trace → env → permissions → logs → simplify

**If not → [Level 11](../11-debugging/)**

---

## The Meta-Checklist

Beyond specific knowledge, you should be able to:

- [ ] **Read any shell script** and understand what it does, line by line
- [ ] **Write scripts that fail safely** — errors stop execution, cleanup always runs
- [ ] **Debug without Stack Overflow** — you have systematic tools, not just "try things"
- [ ] **Explain your commands** — not just type them
- [ ] **Choose the right tool** — shell for glue/automation, a real language for complexity
- [ ] **Recognize when shell is wrong** — and reach for Python, Go, or another language

---

## When to Stop Using Shell

Shell is the right tool for:
- Gluing programs together
- Automation scripts under ~200 lines
- System administration tasks
- Quick data processing with Unix tools
- CI/CD pipelines and build scripts

Shell is the **wrong** tool for:
- Complex data structures (use Python, Go)
- Error handling that needs to be robust (use a real language)
- Anything over ~300 lines (use a real language)
- HTTP APIs, database access, complex parsing (use a real language)
- Anything that needs to be maintainable by a team (probably use a real language)

The mark of shell literacy isn't writing everything in shell — it's knowing exactly when shell is the right choice and when it isn't.

---

## What to Do Next

1. **Practice daily**: Use your terminal intentionally. When you copy a command, stop and understand it first.
2. **Read other people's scripts**: `/etc/profile`, `/etc/bashrc`, systemd unit files, CI configs.
3. **Write real scripts**: Automate something you do manually at least once a week.
4. **Break things safely**: Use VMs or containers to experiment with destructive commands.
5. **Teach someone**: Explaining forces understanding. Write notes, pair program, review scripts.

---

*You've reached the end of the curriculum. Go build things.*
