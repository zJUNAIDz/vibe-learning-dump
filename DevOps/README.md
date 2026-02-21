# DevOps for Developers

> **A production-grade DevOps curriculum for software developers who want to understand, build, deploy, and maintain systems in production.**

---

## 🎯 Who This Is For

You are a **fullstack developer** who:
- Writes primarily in **TypeScript** or **Go**
- Uses **Linux** (Fedora or similar)
- Already knows Docker basics and some AWS
- Wants to be **fluent in DevOps**, not just survive it
- Needs to design, debug, deploy, and reason about production systems
- Wants platform-agnostic knowledge, not vendor memorization

This is **NOT** a pure infrastructure/SRE track. It's for developers who want strong DevOps fluency.

---

## 🧠 Teaching Philosophy

This curriculum follows these principles:

1. **First principles first** — Understand *why* before *how*
2. **Mental models over magic** — No "it just works" handwaving
3. **Developer-centric** — How infra decisions affect your code
4. **Reality over fantasy** — Tradeoffs, failures, footguns included
5. **Incremental depth** — Beginner → Intermediate → Advanced

**Difficulty Markers:**
- 🟢 **Fundamentals** — Core concepts everyone must know
- 🟡 **Intermediate** — Real-world complexity emerges
- 🔴 **Advanced** — Deep dives, edge cases, production war stories

---

## 📚 Curriculum Structure

### **[00. DevOps Mindset for Developers](./00-devops-mindset/)**
What DevOps actually means, "you build it you run it", and how developers cause outages.

### **[01. Linux, OS & Systems Fundamentals](./01-linux-and-systems/)**
Processes, memory, filesystems, networking, cgroups, namespaces — enough to understand containers and Kubernetes.

### **[02. Containers Deep Dive](./02-containers-deep-dive/)**
What containers actually are (Linux primitives), image optimization, multi-stage builds, security issues.

### **[03. Kubernetes from First Principles](./03-kubernetes-fundamentals/)**
Control plane, data plane, desired vs actual state, etcd, scheduler, controllers — the real mental model.

### **[04. Kubernetes as a Developer](./04-kubernetes-for-developers/)**
How to think in Kubernetes, debug pods, manage resources, avoid OOMKills.

### **[05. K9s (Terminal UX for Kubernetes)](./05-k9s/)**
Real debugging workflows with K9s, mapping to kubectl concepts.

### **[06. CI/CD from First Principles](./06-ci-cd-fundamentals/)**
What pipelines really are, build vs deploy, versioning, rollbacks.

### **[07. Jenkins Deep Dive](./07-jenkins/)**
Jenkins architecture, agents, pipelines as code, CI for TypeScript/Go apps.

### **[08. Infrastructure as Code (IaC)](./08-iac-fundamentals/)**
Why IaC exists, declarative vs imperative, state management.

### **[09. Terraform](./09-terraform/)**
Terraform mental model, providers, state, modules, environment separation.

### **[10. Ansible](./10-ansible/)**
Configuration management, agentless model, playbooks, idempotency.

### **[11. Makefile for Developers](./11-makefile/)**
Make beyond C folklore, dependency graphs, using Make for dev tooling.

### **[12. Cloud from First Principles](./12-cloud-fundamentals/)**
Platform-agnostic cloud concepts: compute, storage, networking, scaling.

### **[13. Observability](./13-observability/)**
Logs vs metrics vs traces, RED/USE metrics, alerts vs signals.

### **[14. Security for Developers](./14-security/)**
Secrets management, IAM, least privilege, container security, supply chain attacks.

### **[15. Real-World Scenarios & Failure Modes](./15-failure-modes/)**
Broken deploys, bad rollouts, leaking secrets, CI disasters, Kubernetes outages.

### **[16. Capstone: From Laptop to Production](./16-capstone/)**
Full walkthrough: TypeScript/Go service → containerized → CI → K8s → IaC → observability.

---

## 🚀 How to Use This Curriculum

1. **Follow the order** — Each module builds on previous ones
2. **Do the exercises** — Reading alone won't make you fluent
3. **Break things** — Use VMs/containers, not production
4. **Question everything** — If something sounds like magic, dig deeper

---

## 🛠️ Prerequisites

- Comfortable with command line
- Know at least one programming language well (TypeScript or Go preferred)
- Basic Docker knowledge (can write and run a Dockerfile)
- Willing to break things and debug

---

## 📖 How to Read

- Each module is self-contained
- Code examples are real and tested
- "War Stories" sections are real production failures
- "Common Traps" sections highlight rookie mistakes

---

## 🔗 Quick Navigation

- **New to DevOps?** → Start at [00-devops-mindset](./00-devops-mindset/)
- **Know containers, need Kubernetes?** → Jump to [03-kubernetes-fundamentals](./03-kubernetes-fundamentals/)
- **Need CI/CD now?** → Start at [06-ci-cd-fundamentals](./06-ci-cd-fundamentals/)
- **Production firefighting?** → See [15-failure-modes](./15-failure-modes/)

---

## 🎓 After This Curriculum

You will be able to:
- Design and deploy production-grade services
- Debug infrastructure and CI/CD issues
- Collaborate effectively with DevOps/SRE teams
- Make informed decisions about architecture and tooling
- Respond to outages without panicking
- Reason about tradeoffs in distributed systems

---

## 📝 Contributing

This is a living curriculum. If you find errors, unclear explanations, or want to add real-world examples, contributions are welcome.

---

**Start here:** [00. DevOps Mindset for Developers →](./00-devops-mindset/00-what-is-devops-for-developers.md)
