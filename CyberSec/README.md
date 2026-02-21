# Cybersecurity for Developers

> **A practical, defensive-first security curriculum for fullstack developers**

This is not hacker cosplay. This is not bug bounty training. This is **how to not ship vulnerabilities** in modern web applications.

---

## 🎯 Who This Is For

- **Fullstack/backend developers** building web applications
- Already comfortable with Linux, OS fundamentals, and networking
- Working with TypeScript/Node.js (and maybe Go)
- Using Docker, Kubernetes, CI/CD pipelines
- Want to understand **why** security bugs happen and **how** to prevent them

---

## 🧠 Core Philosophy

Security is not about memorizing CVEs or running automated scanners. It's about:

1. **Understanding systems** — how browsers, servers, networks, and OSes interact
2. **Reasoning about trust** — what can attackers control? What do you assume?
3. **Defensive design** — building systems that resist attacks by default
4. **Realistic threat modeling** — who would attack you and why?

Most breaches are **boring**:
- Leaked credentials
- Misconfigured S3 buckets
- Unpatched dependencies
- Logic flaws in authorization

This curriculum teaches you to **think like a defender** while understanding how attackers reason.

---

## 📁 Curriculum Structure

```
cybersecurity-for-developers/
├── 00-orientation/
├── 01-threat-modeling/
├── 02-web-architecture/
├── 03-http-security/
├── 04-authentication-authorization/
├── 05-owasp-top-10/
├── 06-linux-os-security/
├── 07-networking-attacks/
├── 08-burp-suite/
├── 09-wireshark/
├── 10-linux-cli-tools/
├── 11-typescript-security/
├── 12-secure-api-design/
├── 13-cicd-supply-chain/
├── 14-logging-monitoring/
├── 15-real-world-failures/
└── 16-capstone/
```

---

## 🚦 Difficulty Levels

Throughout the curriculum, modules are tagged:

- 🟢 **Fundamentals** — Core concepts every developer must know
- 🟡 **Intermediate** — Deeper understanding, practical application
- 🔴 **Advanced** — Systems-level details, edge cases, complex attacks

---

## 🛠️ Prerequisites

You should be comfortable with:

- **Linux command line** (Fedora or similar)
- **OS fundamentals** (processes, memory, filesystems)
- **Networking basics** (TCP/IP, DNS, HTTP)
- **Web development** (APIs, authentication, databases)
- **TypeScript/Node.js** (primary language for examples)

---

## 🗺️ How to Use This Curriculum

### Option 1: Linear (Recommended)
Work through modules 00-16 in order. Each builds on previous concepts.

### Option 2: Topic-Focused
Jump to specific topics:
- **Web app security?** → 02, 03, 04, 05
- **Infrastructure security?** → 06, 07, 13
- **Tooling?** → 08, 09, 10
- **Language-specific?** → 11, 12

### Option 3: Capstone-First
Read module 16 first to see the end goal, then work backward through foundational topics.

---

## ⚠️ What This Is NOT

- ❌ Certification prep (OSCP, CEH, etc.)
- ❌ Bug bounty training
- ❌ Penetration testing course
- ❌ Tool memorization
- ❌ "Elite hacker" cosplay

---

## ✅ What This IS

- ✅ **Defensive development** practices
- ✅ **Root cause understanding** of vulnerabilities
- ✅ **Systems thinking** about security
- ✅ **Practical tooling** to validate security assumptions
- ✅ **Real-world examples** from production systems

---

## 🚀 Quick Start

1. Start with [00-orientation](./00-orientation/00-how-web-apps-get-hacked.md)
2. Read [START_HERE.md](./START_HERE.md) for setup instructions
3. Follow the [GETTING_STARTED.md](./GETTING_STARTED.md) guide
4. Bookmark [QUICK_REFERENCE.md](./QUICK_REFERENCE.md) for tools and commands

---

## 📚 Additional Resources

Each module includes:
- **Concepts** — Theory and mental models
- **Examples** — Real code (TypeScript/Go)
- **Tools** — Practical usage
- **Exercises** — Hands-on practice
- **Common Mistakes** — What to avoid

---

## 🤝 Contributing

Found an error? Have a suggestion? This curriculum is a living document.

---

## 📖 License

This curriculum is open for learning. Use it, share it, improve it.

---

**Remember:** The goal is not to become a security expert. The goal is to **ship code that doesn't get pwned**.

Let's begin. → [Start Here](./START_HERE.md)
