# Linux Operating Systems: A Developer's Deep Dive

**A kernel-aware, production-ready curriculum for software engineers who want to understand what the OS is actually doing.**

---

## 🎯 Who This Is For

- **Linux daily users** (Fedora/systemd-based distros)
- **Backend & systems developers** (TypeScript, Go, Python)
- **DevOps engineers** working with Docker, Kubernetes, CI/CD
- **Curious developers** tired of "it just works" — you want to know **why**

This is NOT a certification prep course. This is NOT for Windows users. This is NOT a command memorization guide.

This **IS** a deep, reality-based exploration of how Linux actually works, from the developer's perspective.

---

## 🧠 Teaching Philosophy

### 1. OS as a Machine, Not Magic
Every abstraction is explained downward:
```
Your Application
    ↓ (library call)
Process/Thread
    ↓ (syscall)
Kernel
    ↓ (hardware interaction)
CPU/Memory/Disk
```

### 2. Developer-First Perspective
Every concept answers: **"Why should I, a developer, care?"**

Ties to:
- Performance characteristics
- Production bugs and incidents
- Container behavior
- Kubernetes quirks
- Database performance

### 3. Reality-Based Linux
- systemd exists → we explain why
- `/proc` exists → we explain why
- Signals exist → we explain why
- "POSIX says so" is never sufficient

### 4. Networking is Part of the OS
We don't treat networking as separate theory. Sockets are file descriptors. Network namespaces are containers. Taught where they naturally belong.

### 5. Incremental Depth
- 🟢 **Fundamentals** — Essential mental models
- 🟡 **Intermediate** — Production-relevant depth
- 🔴 **Advanced** — Kernel-level understanding

---

## 📚 Curriculum Structure

### [Module 00: Orientation](00-orientation/)
**How Linux Actually Thinks**
- What an OS fundamentally does
- Kernel space vs user space (and why it matters)
- Monolithic kernel architecture
- How applications interact with the kernel
- Why Linux is not "just Unix"

### [Module 01: Processes, Threads & Execution](01-processes-and-execution/)
**The Foundation of Running Code**
- What a process truly is (beyond "running program")
- Process lifecycle: fork, exec, wait
- Threads vs processes (and when to use which)
- Context switching (what gets saved, what doesn't)
- Process tree, PID namespaces
- Signals and process control
- Zombie and orphan processes
- Why Node.js behaves the way it does

### [Module 02: Memory Management](02-memory/)
**Critical for Every Developer**
- Virtual memory architecture
- Address spaces (do processes really have "their own" memory?)
- Stack vs heap (in reality, not theory)
- Memory mapping (mmap)
- The page cache (Linux's secret performance weapon)
- Swapping and OOM killer
- Why containers get OOMKilled
- Why "free memory" being low is often good

### [Module 03: Filesystems & Storage](03-filesystems/)
**"Everything is a File" — What That Really Means**
- Filesystem fundamentals
- Inodes (what files "really" are)
- Hard links vs symbolic links
- File permissions beyond rwx
- Mount points and the unified filesystem tree
- Special filesystems: `/dev`, `/sys`, `/proc`
- OverlayFS and container layers
- How this affects Docker and Kubernetes

### [Module 04: File Descriptors & I/O](04-file-descriptors/)
**The Interface to Everything**
- File descriptors as integers referencing kernel objects
- stdin, stdout, stderr (0, 1, 2)
- Blocking vs non-blocking I/O
- select, poll, epoll (how scalable servers work)
- Why Node.js and Go scale well
- Pipes and redirection
- "Too many open files" errors explained

### [Module 05: Networking Inside Linux](05-networking-fundamentals/)
**Networking as an OS Subsystem**
- Sockets are file descriptors
- TCP vs UDP from the kernel's perspective
- Ports and binding
- Loopback interface (127.0.0.1)
- Routing tables
- ARP (how IP maps to MAC)
- DNS resolution path in Linux
- Essential tools: ss, ip, ping, traceroute

### [Module 06: Linux Networking Deep Dive](06-networking-deep/)
**How Containers Network**
- Network namespaces (network isolation)
- veth pairs (virtual ethernet)
- Linux bridges
- NAT (Network Address Translation)
- iptables/nftables (packet filtering)
- How Docker networking works
- Kubernetes CNI basics
- "Why my service works locally but not remotely"

### [Module 07: Namespaces & cgroups](07-containers/)
**Containers Are Just Linux Processes**
- What namespaces isolate (pid, mount, net, user, etc.)
- What cgroups limit (CPU, memory, I/O)
- cgroups v2 architecture
- How Docker uses these primitives
- How Kubernetes builds on top
- Why containers are NOT VMs

### [Module 08: systemd](08-systemd/)
**The Init System (Without the Politics)**
- Why systemd exists
- What it replaced (and why)
- Units, services, targets
- Timers (cron replacement)
- Journald (logging)
- Socket activation
- How servers actually boot
- Debugging service failures

### [Module 09: Boot Process](09-boot/)
**From Power Button to Shell Prompt**
- Firmware (UEFI/BIOS)
- Bootloader (GRUB)
- Kernel loading and initialization
- initramfs (initial ramdisk)
- Init system handoff
- Kernel parameters
- Why boot failures happen

### [Module 10: Users, Permissions & Security](10-security/)
**Multi-User OS Fundamentals**
- UID and GID (what they really are)
- Permission model (user, group, other)
- Capabilities (fine-grained privileges)
- setuid and setgid bits
- sudo (how it works)
- Why containers shouldn't run as root
- Security namespaces

### [Module 11: Performance, Debugging & Failure Analysis](11-debugging/)
**Tools to Understand What's Actually Happening**
- strace (syscall tracing)
- lsof (list open files)
- vmstat (virtual memory statistics)
- iostat (I/O statistics)
- perf (performance analysis)
- How to debug a hung process
- How to debug high CPU usage
- How to debug memory leaks

### [Module 12: Linux for Developers in Production](12-production/)
**What Changes When Code Runs in Production**
- Kernel limits (ulimit, sysctl)
- File descriptor exhaustion
- Clock synchronization (NTP)
- Time zones and UTC
- Signal handling in production
- Core dumps
- Production debugging without disrupting service

### [Module 13: Real Failure Stories](13-failure-stories/)
**Learn from Production Incidents**
Realistic scenarios with OS-level explanations:
- Node.js app freezes under load
- Go service memory leak
- Port conflicts and "Address already in use"
- DNS works on host but fails in container
- Kubernetes pod CrashLoopBackOff
- System becomes unresponsive under I/O pressure
- Zombie process accumulation

### [Module 14: Capstone: Think Like the Kernel](14-capstone/)
**Mental Models for Linux Reasoning**
- Debugging framework: symptom → syscall → subsystem
- The kernel's perspective
- When to use which tool
- Building intuition
- Going deeper (resources)

---

## 🚀 How to Use This Curriculum

### Recommended Path
1. Start with Module 00 (Orientation) — get the mental model
2. Work through Modules 01-04 in order — these are foundational
3. Modules 05-06 (Networking) can be taken together
4. Modules 07-09 build on everything prior
5. Modules 10-12 are more specialized but critical for production
6. Module 13 ties everything together with real scenarios
7. Module 14 is your graduation — a framework for continued learning

### Hands-On Practice
This curriculum emphasizes **understanding over memorization**. For each module:
- Read the concepts
- Run the examples on your Linux system
- Experiment and break things (safely)
- Apply to real problems you face

### Prerequisites
- Daily Linux usage (terminal comfort)
- Basic programming experience
- Willingness to dig deep

---

## 🛠️ Environment

**Recommended:**
- Fedora Linux (or RHEL-based)
- systemd-based distribution
- Modern kernel (5.x+)

**Also works with:**
- Ubuntu/Debian (systemd)
- Arch Linux
- Any modern Linux distribution

**Does NOT cover:**
- Windows (WSL is Linux, but we focus on native Linux)
- macOS (BSD-based, different enough to be out of scope)

---

## 📖 Reading Approach

Each module file is:
- **Standalone** — can be read independently
- **Complete** — no "TODO" or "coming soon"
- **Practical** — tied to real developer concerns
- **Deep** — goes beyond surface-level commands

Look for:
- 🟢 🟡 🔴 depth markers
- `Code examples`
- Mermaid diagrams
- "Why this matters to developers" sections
- "Common pitfalls" warnings

---

## 🎓 After This Curriculum

You will be able to:
- Understand what your application is actually doing at the OS level
- Debug production issues using OS-level tools
- Make informed performance decisions
- Understand container and Kubernetes behavior
- Read kernel documentation and understand it
- Reason about any system problem from first principles

You will NOT be:
- A kernel developer (unless you want to continue in that direction)
- Certified in anything (we don't care about certs)
- Able to fix Windows (not our problem)

---

## 📜 License

This curriculum is educational content. Use it to learn, teach, and build better systems.

---

## 🏁 Ready?

Start with [Module 00: How Linux Actually Works](00-orientation/00-how-linux-actually-works.md)

**Remember:** There is no magic. Just layers of abstraction you can understand.

Welcome to Linux.
