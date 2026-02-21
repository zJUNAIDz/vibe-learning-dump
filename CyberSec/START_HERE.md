# Start Here 🚀

Welcome to **Cybersecurity for Developers** — a practical, no-BS guide to not shipping vulnerabilities.

---

## 📌 What You're About to Learn

This curriculum teaches you to:

1. **Think defensively** while understanding attacker reasoning
2. **Trace vulnerabilities to root causes** (OS, network, architecture)
3. **Use security tools** as observation instruments, not magic
4. **Design systems** that resist common attacks
5. **Debug security issues** like you debug application bugs

---

## 🎯 Your Current Skills (Assumed)

You already know:

- ✅ Linux CLI (basic to intermediate)
- ✅ Process model, filesystems, memory
- ✅ Networking (TCP/IP, DNS, HTTP)
- ✅ Web development (Node.js, TypeScript, REST APIs)
- ✅ Docker and containerization basics

If you're rusty on any of these, **that's fine** — we'll review OS/network concepts through a security lens.

---

## 🛠️ Required Setup

### 1. Operating System
- **Fedora Linux** (or similar: Ubuntu, Debian, Arch)
- **WSL2** works if you're on Windows
- **macOS** works for most things (some Linux-specific tools differ)

### 2. Core Tools (Install Now)

```bash
# Update system
sudo dnf update -y

# Core development
sudo dnf install -y git curl wget vim tmux

# Node.js and TypeScript
sudo dnf install -y nodejs npm
npm install -g typescript ts-node

# Go (optional, for later modules)
sudo dnf install -y golang

# Docker and Podman
sudo dnf install -y docker podman podman-docker
sudo systemctl enable --now docker

# Security tools (we'll learn these)
sudo dnf install -y nmap netcat wireshark-cli tcpdump
sudo dnf install -y strace ltrace lsof

# HTTP tools
sudo dnf install -y httpie jq
```

### 3. Burp Suite Community Edition

```bash
# Download from: https://portswigger.net/burp/communitydownload
# Install and configure (we'll cover this in module 08)
```

### 4. Optional: Firefox + FoxyProxy
- Firefox is better for security testing than Chrome
- Install FoxyProxy extension for easy proxy switching

---

## 📂 How This Curriculum is Organized

```
CyberSec/
├── README.md              ← You've read this
├── START_HERE.md          ← You are here
├── GETTING_STARTED.md     ← Detailed setup guide
├── QUICK_REFERENCE.md     ← Command cheat sheet
│
├── 00-orientation/        🟢 What cybersecurity actually is
├── 01-threat-modeling/    🟢 Thinking like an attacker
├── 02-web-architecture/   🟢 Attack surface analysis
├── 03-http-security/      🟢 HTTP, cookies, CORS, CSRF
├── 04-authentication-authorization/ 🟢 Auth done right
├── 05-owasp-top-10/       🟡 Common vulnerabilities
├── 06-linux-os-security/  🟡 OS-level security
├── 07-networking-attacks/ 🟡 Network-layer threats
├── 08-burp-suite/         🟡 Proxying and testing
├── 09-wireshark/          🟡 Packet analysis
├── 10-linux-cli-tools/    🟡 Security debugging
├── 11-typescript-security/ 🟡 Language-specific issues
├── 12-secure-api-design/  🟡 API security patterns
├── 13-cicd-supply-chain/  🔴 Build and supply chain
├── 14-logging-monitoring/ 🔴 Detection and response
├── 15-real-world-failures/ 🔴 Case studies
└── 16-capstone/           🔴 Break and defend
```

---

## 🗺️ Choose Your Path

### Path A: Full Linear Journey (Recommended)
**Time: 6-8 weeks (2-3 hours/day)**

Work through modules 00-16 in order. Best for systematic learning.

### Path B: Web Security Focus (Fastest)
**Time: 3-4 weeks**

1. Module 00 (orientation)
2. Module 01 (threat modeling)
3. Module 02 (web architecture)
4. Module 03 (HTTP security)
5. Module 04 (authentication)
6. Module 05 (OWASP Top 10)
7. Module 11 (TypeScript security)
8. Module 12 (API design)
9. Module 16 (capstone)

### Path C: Infrastructure Security Focus
**Time: 3-4 weeks**

1. Module 00 (orientation)
2. Module 06 (Linux/OS security)
3. Module 07 (networking attacks)
4. Module 10 (CLI tools)
5. Module 13 (CI/CD security)
6. Module 14 (logging/monitoring)
7. Module 16 (capstone)

### Path D: Capstone-Driven (Exploratory)
**Time: Self-paced**

1. Start with Module 16 (capstone)
2. Work backward through topics you don't understand
3. Build practical intuition first, theory second

---

## 🧭 Learning Approach

### 1. Read Actively
- Take notes
- Question assumptions
- Try to break examples mentally

### 2. Run Every Command
- Type commands yourself
- Modify them
- See what breaks

### 3. Build Mental Models
Security is not a checklist. It's a way of reasoning about:
- **Trust boundaries** — what's hostile vs friendly?
- **Attack surfaces** — what can attackers touch?
- **Assumptions** — what did you assume was safe?

### 4. Fail Forward
- Make mistakes in your local environment
- Break things intentionally
- Fix them yourself

---

## ⚠️ Safety and Ethics

### What You Can Do
- ✅ Test **your own applications**
- ✅ Use deliberately vulnerable apps (DVWA, WebGoat, Juice Shop)
- ✅ Experiment in **isolated local environments**
- ✅ Learn for **defensive purposes**

### What You CANNOT Do
- ❌ Attack systems you don't own
- ❌ Test production apps without authorization
- ❌ Use these techniques illegally
- ❌ Bypass authentication on real services "for learning"

**Legal and ethical behavior is non-negotiable.**

---

## 📚 Supplementary Resources

These are optional but valuable:

- **Books:**
  - *The Web Application Hacker's Handbook* (Stuttard & Pinto)
  - *Security Engineering* (Ross Anderson)
  - *The Tangled Web* (Michal Zalewski)

- **Deliberately Vulnerable Apps:**
  - OWASP Juice Shop (Node.js)
  - DVWA (PHP)
  - WebGoat (Java)

- **Blogs/Papers:**
  - Troy Hunt's blog
  - Krebs on Security
  - Google Project Zero

---

## ✅ Quick Checklist

Before starting Module 00, make sure you have:

- [ ] Linux environment ready
- [ ] Core CLI tools installed (`git`, `curl`, `vim`)
- [ ] Node.js and TypeScript installed
- [ ] Docker/Podman running
- [ ] Basic security tools installed (`nmap`, `wireshark-cli`, `tcpdump`)
- [ ] Text editor or IDE configured
- [ ] Noted the ethical guidelines above

---

## 🚀 Ready?

You're all set. Let's start with the fundamentals:

→ **[Module 00: Orientation](./00-orientation/00-how-web-apps-get-hacked.md)**

---

Remember: **Security is not a destination. It's a way of thinking.**
