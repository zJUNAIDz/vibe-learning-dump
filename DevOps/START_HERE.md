# 👋 START HERE

Welcome to **DevOps for Developers** — a production-grade curriculum that will take you from "I can run Docker" to "I can deploy and debug production systems confidently."

---

## 🎯 Who This Is For

You are:
- A **software developer** (fullstack, backend, frontend)
- Writing primarily in **TypeScript** and/or **Go**
- On **Linux** (Fedora or similar)
- Already know **basic Docker** and some **AWS**
- Want to be **DevOps fluent**, not just survive it

This curriculum is **NOT**:
- For pure infrastructure/SRE roles
- Vendor-specific (AWS-only, etc.)
- A collection of "just run this" tutorials

This curriculum **IS**:
- First principles explanations
- Platform-agnostic knowledge
- Developer-centric
- Real-world focused (failures, tradeoffs, war stories)

---

## 📖 Essential Reading (Read First!)

1. **[GETTING_STARTED.md](./GETTING_STARTED.md)** — Setup, prerequisites, learning tips
2. **[README.md](./README.md)** — Curriculum overview and structure
3. **[QUICK_REFERENCE.md](./QUICK_REFERENCE.md)** — Bookmark this for commands

---

## 🚀 Your Learning Path

### Week 1-2: Foundations
**Goal:** Understand what's under the hood

- ✅ [00. DevOps Mindset](./00-devops-mindset/) (1-2 hours)
  - What DevOps actually means
  - "You build it, you run it" explained
  - How developers cause outages

- ✅ [01. Linux & Systems](./01-linux-and-systems/) (8-10 hours)
  - Processes, memory, CPU
  - Networking (TCP/UDP, DNS, ports)
  - cgroups and namespaces (bridge to containers)
  - systemd

---

### Week 3-4: Containers
**Goal:** Master containerization

- ✅ [02. Containers Deep Dive](./02-containers-deep-dive/) (4-6 hours)
  - What containers actually are
  - Dockerfile optimization
  - Multi-stage builds
  - Security best practices

---

### Week 5-7: Kubernetes
**Goal:** Understand orchestration

- ✅ [03. Kubernetes Fundamentals](./03-kubernetes-fundamentals/) (10-12 hours)
  - Why Kubernetes exists
  - Control plane architecture
  - Pods, Deployments, Services
  - ConfigMaps, Secrets, Volumes

- ✅ [04. Kubernetes for Developers](./04-kubernetes-for-developers/) (6-8 hours)
  - Debugging pods
  - Resource limits (avoiding OOMKills)
  - Local development workflows
  - Advanced patterns

- ✅ [05. K9s](./05-k9s/) (2-3 hours)
  - Terminal UI for Kubernetes
  - Faster debugging workflows

---

### Week 8-10: CI/CD & Automation
**Goal:** Automate everything

- ✅ [06. CI/CD Fundamentals](./06-ci-cd-fundamentals/) (4-6 hours)
  - Pipelines from first principles
  - Build, test, deploy
  - Deployment strategies

- ✅ [07. Jenkins Deep Dive](./07-jenkins/) (6-8 hours)
  - Pipeline as code
  - Building TypeScript/Go apps
  - Kubernetes integration

---

### Week 11-13: Infrastructure as Code
**Goal:** Version your infrastructure

- ✅ [08. IaC Fundamentals](./08-iac-fundamentals/) (2-3 hours)
  - Why IaC exists
  - Declarative vs imperative
  - State management

- ✅ [09. Terraform](./09-terraform/) (6-8 hours)
  - Providers, resources, modules
  - State deep dive
  - Best practices

- ✅ [10. Ansible](./10-ansible/) (4-6 hours)
  - Configuration management
  - Playbooks and roles
  - Ansible + Terraform workflows

- ✅ [11. Makefile](./11-makefile/) (2-3 hours)
  - Automating common tasks
  - Make beyond C folklore

---

### Week 14-15: Production Operations
**Goal:** Run systems reliably

- ✅ [12. Cloud Fundamentals](./12-cloud-fundamentals/) (4-6 hours)
  - Platform-agnostic cloud concepts
  - Compute, storage, networking
  - Mapping to AWS/GCP/Azure

- ✅ [13. Observability](./13-observability/) (6-8 hours)
  - Logs, metrics, traces
  - Prometheus, Grafana
  - Alerting strategies

- ✅ [14. Security](./14-security/) (4-6 hours)
  - Secrets management
  - Container security
  - Supply chain security

---

### Week 16: Learning from Failures
**Goal:** Learn from others' mistakes

- ✅ [15. Real-World Failure Modes](./15-failure-modes/) (2-3 hours)
  - Broken deploys
  - Leaked secrets
  - Kubernetes outages
  - Terraform disasters

---

### Week 17-18: Capstone Project
**Goal:** Prove your skills

- ✅ [16. Capstone: Laptop to Production](./16-capstone/) (12-16 hours)
  - Build a complete service
  - Docker + CI/CD + Kubernetes + IaC
  - Observability + Security
  - Portfolio-ready project

---

## 🛠️ Quick Setup (5 Minutes)

```bash
# 1. Install essential tools (Fedora)
sudo dnf install -y git kubectl podman k9s terraform ansible

# 2. Start local Kubernetes
sudo dnf install -y minikube
minikube start

# 3. Verify
kubectl get nodes
podman run --rm hello-world

# 4. Clone this curriculum (if not already)
# (You're already in it!)

# 5. Start learning!
cd 00-devops-mindset
cat 00-what-is-devops-for-developers.md
```

---

## 📚 Study Tips

### ✅ DO:
- Run every command yourself
- Break things intentionally (in test environments!)
- Do the exercises
- Keep a learning journal
- Ask questions in communities

### ❌ DON'T:
- Just read without practicing
- Skip the foundational modules
- Practice on production
- Use `latest` tag in production
- Commit secrets to Git

---

## 🆘 Stuck? Here's How to Get Unstuck

1. **Check the module README** — troubleshooting tips
2. **Search the Quick Reference** — common commands
3. **Google the error** — someone else hit it
4. **Ask in communities:**
   - r/kubernetes, r/devops (Reddit)
   - Kubernetes Slack/Discord
   - Stack Overflow

---

## 🎓 After You Finish

You'll be able to:
- ✅ Deploy services to production confidently
- ✅ Debug infrastructure issues
- ✅ Write CI/CD pipelines
- ✅ Provision infrastructure with Terraform
- ✅ Secure containers and pipelines
- ✅ Respond to outages without panicking
- ✅ Collaborate with DevOps/SRE teams effectively

**You'll be DevOps fluent.** 🎉

---

## 🚀 Ready? Let's Go!

**Start here:**
### → [00. DevOps Mindset for Developers](./00-devops-mindset/00-what-is-devops-for-developers.md)

---

## 📊 Track Your Progress

```
[ ] Module 00: DevOps Mindset
[ ] Module 01: Linux & Systems
[ ] Module 02: Containers
[ ] Module 03: Kubernetes Fundamentals
[ ] Module 04: Kubernetes for Developers
[ ] Module 05: K9s
[ ] Module 06: CI/CD Fundamentals
[ ] Module 07: Jenkins
[ ] Module 08: IaC Fundamentals
[ ] Module 09: Terraform
[ ] Module 10: Ansible
[ ] Module 11: Makefile
[ ] Module 12: Cloud Fundamentals
[ ] Module 13: Observability
[ ] Module 14: Security
[ ] Module 15: Failure Modes
[ ] Module 16: Capstone Project

Completion: _____ / 17 modules (___%)
```

---

**Good luck! You've got this.** 💪

Remember: Everyone struggles at first, especially with Kubernetes. That's normal. Keep going.
