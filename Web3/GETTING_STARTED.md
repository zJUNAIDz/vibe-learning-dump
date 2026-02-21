# Getting Started

Quick setup guide to begin learning Web3.

---

## Environment Setup

### 1. Install Node.js (v18+)

```bash
# Check current version
node --version
npm --version

# Fedora/RHEL/CentOS
sudo dnf install nodejs npm

# Ubuntu/Debian
sudo apt update
sudo apt install nodejs npm

# Or use nvm (recommended for version management)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
source ~/.bashrc
nvm install 18
nvm use 18
```

---

### 2. Install Git

```bash
# Fedora/RHEL
sudo dnf install git

# Ubuntu/Debian
sudo apt install git

# Verify
git --version
```

---

### 3. Install VS Code (or your preferred editor)

```bash
# Fedora/RHEL
sudo rpm --import https://packages.microsoft.com/keys/microsoft.asc
sudo sh -c 'echo -e "[code]\nname=Visual Studio Code\nbaseurl=https://packages.microsoft.com/yumrepos/vscode\nenabled=1\ngpgcheck=1\ngpgkey=https://packages.microsoft.com/keys/microsoft.asc" > /etc/yum.repos.d/vscode.repo'
sudo dnf install code

# Ubuntu/Debian
wget -qO- https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > packages.microsoft.gpg
sudo install -o root -g root -m 644 packages.microsoft.gpg /usr/share/keyrings/
sudo sh -c 'echo "deb [arch=amd64 signed-by=/usr/share/keyrings/packages.microsoft.gpg] https://packages.microsoft.com/repos/vscode stable main" > /etc/apt/sources.list.d/vscode.list'
sudo apt update
sudo apt install code
```

---

### 4. Create Your Learning Directory

```bash
mkdir -p ~/web3-learning
cd ~/web3-learning
```

---

## Optional Tooling (Install Later)

### Docker (for running local blockchain nodes)

```bash
# Fedora/RHEL
sudo dnf install docker
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER

# Ubuntu/Debian
sudo apt install docker.io
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER

# Log out and back in for group changes to take effect
```

---

### Hardhat (Ethereum development environment)

You'll install this in **Module 10**.

```bash
npm install --save-dev hardhat
```

---

### Foundry (Alternative Ethereum development toolkit)

Also covered in **Module 10**.

```bash
curl -L https://foundry.paradigm.xyz | bash
foundryup
```

---

## Verify Your Setup

Run these commands to confirm everything works:

```bash
node --version    # Should show v18.x.x or higher
npm --version     # Should show 9.x.x or higher
git --version     # Should show git version
code --version    # Should show VS Code version
```

---

## Create Your First Test Project

Let's verify Node.js and npm work correctly:

```bash
mkdir ~/web3-learning/test-project
cd ~/web3-learning/test-project
npm init -y
npm install typescript @types/node --save-dev
npx tsc --init
```

Create a test file:

```bash
echo 'console.log("Setup complete!");' > index.ts
npx ts-node index.ts
```

If you see "Setup complete!", you're ready to start.

---

## What's Next?

1. Read [START_HERE.md](START_HERE.md) for curriculum overview
2. Begin with [00-orientation/](00-orientation/)

---

## Troubleshooting

### "command not found: node"
- Ensure Node.js is installed
- Check your PATH: `echo $PATH`
- If using nvm, run: `nvm use 18`

### "permission denied: docker"
- Add yourself to docker group: `sudo usermod -aG docker $USER`
- Log out and back in

### npm install fails
- Try clearing npm cache: `npm cache clean --force`
- Delete `node_modules` and `package-lock.json`, then reinstall

---

You're ready to learn Web3. [→ Start Here](START_HERE.md)
