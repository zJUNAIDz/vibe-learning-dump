# Getting Started Guide 🛠️

This guide walks you through setting up your environment for hands-on security learning.

---

## 🖥️ System Requirements

### Minimum
- **OS:** Linux (Fedora, Ubuntu, Debian) or WSL2
- **RAM:** 8 GB
- **Disk:** 20 GB free
- **Network:** Unrestricted internet access

### Recommended
- **RAM:** 16 GB (for running VMs/containers)
- **Disk:** 50 GB free
- **Network:** Ability to run local servers

---

## 📦 Installation Steps

### 1. Update Your System

**Fedora:**
```bash
sudo dnf update -y
sudo dnf install -y dnf-plugins-core
```

**Ubuntu/Debian:**
```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y software-properties-common
```

---

### 2. Development Tools

```bash
# Fedora
sudo dnf groupinstall -y "Development Tools"
sudo dnf install -y git curl wget vim tmux htop

# Ubuntu/Debian
sudo apt install -y build-essential git curl wget vim tmux htop
```

---

### 3. Node.js and TypeScript

**Recommended: Use nvm (Node Version Manager)**

```bash
# Install nvm
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.5/install.sh | bash

# Reload shell
source ~/.bashrc  # or ~/.zshrc

# Install Node.js LTS
nvm install --lts
nvm use --lts

# Install TypeScript globally
npm install -g typescript ts-node @types/node

# Verify
node --version
npm --version
tsc --version
```

---

### 4. Go (Optional)

**Fedora:**
```bash
sudo dnf install -y golang
```

**Ubuntu:**
```bash
sudo snap install go --classic
```

**Verify:**
```bash
go version
```

---

### 5. Docker and Container Tools

**Fedora:**
```bash
# Install Docker
sudo dnf install -y docker
sudo systemctl enable --now docker
sudo usermod -aG docker $USER

# Or use Podman (native Fedora)
sudo dnf install -y podman podman-docker
```

**Ubuntu:**
```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
```

**Verify:**
```bash
# Log out and back in for group changes
docker run hello-world
```

---

### 6. Security Tools

#### Network Analysis
```bash
# Fedora
sudo dnf install -y nmap netcat wireshark-cli tcpdump

# Ubuntu
sudo apt install -y nmap netcat wireshark-cli tcpdump
```

#### System Debugging
```bash
# Fedora
sudo dnf install -y strace ltrace lsof sysstat

# Ubuntu
sudo apt install -y strace ltrace lsof sysstat
```

#### HTTP Testing
```bash
# HTTPie (better than curl for testing)
# Fedora
sudo dnf install -y httpie jq

# Ubuntu
sudo apt install -y httpie jq
```

---

### 7. Burp Suite Community Edition

Burp Suite is a web application security testing toolkit.

**Installation:**

1. Download from: https://portswigger.net/burp/communitydownload
2. Choose "Linux (64-bit)"
3. Install:

```bash
# Make installer executable
chmod +x burpsuite_community_linux_*.sh

# Run installer
./burpsuite_community_linux_*.sh

# Follow GUI prompts
```

4. Launch:
```bash
# Should be in your applications menu, or run:
BurpSuiteCommunity
```

---

### 8. Firefox + Security Extensions

**Firefox is better for security testing than Chrome** (easier to configure proxies, better dev tools for security).

**Install Firefox:**
```bash
# Fedora
sudo dnf install -y firefox

# Ubuntu
sudo snap install firefox
```

**Recommended Extensions:**
1. **FoxyProxy** — Easy proxy switching for Burp
2. **Cookie-Editor** — View/modify cookies
3. **Wappalyzer** — Identify technologies

---

### 9. Python (Optional, for some tools)

Most distros have Python 3 pre-installed. Verify:

```bash
python3 --version
pip3 --version
```

If not installed:
```bash
# Fedora
sudo dnf install -y python3 python3-pip

# Ubuntu
sudo apt install -y python3 python3-pip
```

---

## 🧪 Test Your Setup

Run these commands to verify everything works:

### Basic Tools
```bash
# Should all return version info
node --version
npm --version
tsc --version
docker --version
git --version
```

### Security Tools
```bash
# Network tools
nmap --version
nc -h
tcpdump --version

# System tools
strace --version
lsof -v

# HTTP tools
http --version
jq --version
```

### Docker Test
```bash
docker run --rm alpine:latest echo "Docker works!"
```

---

## 🗂️ Workspace Setup

Create a dedicated directory for practice:

```bash
mkdir -p ~/security-lab
cd ~/security-lab

# Create subdirectories
mkdir -p {web-apps,tools,notes,vulnerable-apps}
```

---

## 🐛 Deliberately Vulnerable Applications

These are safe, legal targets for practice.

### OWASP Juice Shop (Recommended)
```bash
docker pull bkimminich/juice-shop
docker run -d -p 3000:3000 bkimminich/juice-shop

# Access at http://localhost:3000
```

### DVWA (Damn Vulnerable Web Application)
```bash
docker pull vulnerables/web-dvwa
docker run -d -p 8080:80 vulnerables/web-dvwa

# Access at http://localhost:8080
# Default creds: admin/password
```

### WebGoat
```bash
docker pull webgoat/goatandwolf
docker run -d -p 8081:8080 webgoat/goatandwolf

# Access at http://localhost:8081/WebGoat
```

---

## 🔧 Environment Configuration

### Shell Configuration

Add useful aliases to `~/.bashrc` or `~/.zshrc`:

```bash
# Security aliases
alias http-server='python3 -m http.server'
alias myip='curl -s https://ifconfig.me'
alias listening='ss -tunapl | grep LISTEN'
alias ports='sudo netstat -tulanp'

# Docker shortcuts
alias dps='docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"'
alias dlog='docker logs -f'

# Reload shell
source ~/.bashrc  # or ~/.zshrc
```

---

## 📝 Note-Taking Setup (Optional)

Keep organized notes as you learn:

**Option 1: Markdown + Git**
```bash
mkdir ~/security-notes
cd ~/security-notes
git init
echo "# Security Learning Notes" > README.md
```

**Option 2: Obsidian or Notion**
- Obsidian (offline, markdown-based)
- Notion (online, feature-rich)

---

## ✅ Final Checklist

Before starting Module 00:

- [ ] Linux environment ready
- [ ] Node.js + TypeScript working
- [ ] Docker running
- [ ] Basic security tools installed
- [ ] At least one vulnerable app running (Juice Shop recommended)
- [ ] Firefox installed (with FoxyProxy)
- [ ] Burp Suite installed
- [ ] Workspace directory created

---

## 🚨 Troubleshooting

### Docker Permission Errors
```bash
sudo usermod -aG docker $USER
# Log out and log back in
```

### Burp Suite Won't Launch
```bash
# Check Java is installed
java --version

# If not:
sudo dnf install -y java-17-openjdk  # Fedora
sudo apt install -y openjdk-17-jdk   # Ubuntu
```

### Wireshark Permission Denied
```bash
# Add user to wireshark group
sudo usermod -aG wireshark $USER
# Log out and back in
```

---

## 🎓 You're Ready!

Everything set up? Great! Head to:

→ **[Module 00: Orientation](./00-orientation/00-how-web-apps-get-hacked.md)**

---

Questions or issues? Review [START_HERE.md](./START_HERE.md) or check the [QUICK_REFERENCE.md](./QUICK_REFERENCE.md).
