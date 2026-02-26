# Getting Started: Environment Setup

## Verify Your Shell Environment

Run these commands and check the output matches expectations:

```bash
# What shell am I running right now?
echo $SHELL          # Your default login shell
echo $0              # The shell running THIS command (may differ)
ps -p $$             # Process info for current shell

# What Bash version?
bash --version       # Need 5.x for this curriculum

# What Zsh version?
zsh --version        # Need 5.9+ (install with: sudo dnf install zsh)

# What distro?
cat /etc/fedora-release   # Should show Fedora XX

# Where do programs live?
echo $PATH | tr ':' '\n'  # Your command search path, one per line
```

---

## Install Useful Tools

These aren't strictly required but are referenced throughout the curriculum:

```bash
# ShellCheck — static analysis for shell scripts
sudo dnf install ShellCheck

# jq — JSON processor for the command line  
sudo dnf install jq

# tree — directory visualization
sudo dnf install tree

# bat — cat with syntax highlighting (useful for reading scripts)
sudo dnf install bat

# fd-find — user-friendly find alternative (for comparison)
sudo dnf install fd-find

# ripgrep — fast grep alternative (for comparison)
sudo dnf install ripgrep
```

---

## Editor Setup

### VS Code

Install the **ShellCheck** extension and the **Bash IDE** extension. They'll catch quoting errors and common mistakes as you write.

### Neovim / Vim

Add `shellcheck` to your linter setup (via ALE, null-ls, or nvim-lint).

### Any Editor

The curriculum is text files. Any editor works. The important thing is having a terminal open next to your editor.

---

## Create a Practice Directory

```bash
mkdir -p ~/shell-practice
cd ~/shell-practice

# Create a test script to verify your setup
cat > test-setup.sh << 'EOF'
#!/usr/bin/env bash
set -euo pipefail

echo "Shell: $BASH_VERSION"
echo "User: $(whoami)"
echo "Host: $(hostname)"
echo "Distro: $(cat /etc/fedora-release 2>/dev/null || echo 'Not Fedora')"
echo "Kernel: $(uname -r)"
echo "PATH entries: $(echo "$PATH" | tr ':' '\n' | wc -l)"
echo ""
echo "Setup looks good. You're ready."
EOF

chmod +x test-setup.sh
./test-setup.sh
```

If that prints your info without errors, you're ready.

---

## File Conventions in This Curriculum

- All scripts assume `#!/usr/bin/env bash` unless stated otherwise
- `$` prompt = regular user
- `#` prompt = root (or use `sudo`)
- Examples that modify the system will say so explicitly
- Exercises create files in `~/shell-practice/` — won't touch your real system

---

## Next Step

→ [START_HERE.md](START_HERE.md) for the curriculum overview  
→ [Level 0: What a Shell Actually Is](00-what-a-shell-is/00-shell-vs-terminal-vs-tty.md) to begin
