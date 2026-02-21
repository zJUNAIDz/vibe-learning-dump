# Module 04: Kubernetes for Developers

> **Practical Kubernetes from a developer's perspective — debugging, local dev, and avoiding footguns**

---

## Overview

Module 03 taught you **how** Kubernetes works. This module teaches you **how to actually use it** as a developer.

---

## Topics Covered

### 📝 01. Local Development Workflows
- Local cluster options (minikube, kind, Docker Desktop)
- Hot reload and fast feedback
- Telepresence and remote debugging
- Skaffold for inner loop development

### 📝 02. Debugging Pods and Containers
- Why your pod won't start
- CrashLoopBackOff explained
- ImagePullBackOff troubleshooting
- Reading logs effectively
- Using ephemeral containers

### 📝 03. Resource Requests and Limits
- What requests and limits actually do
- Why your pod keeps getting OOMKilled
- CPU throttling explained
- Right-sizing your containers

### 📝 04. Application Configuration
- Environment variables
- ConfigMaps vs Secrets
- Configuration best practices
- 12-factor app principles in Kubernetes

### 📝 05. Persistent Storage for Developers
- When you need persistent storage
- StatefulSets vs Deployments
- PersistentVolumeClaims
- Storage classes
- Database considerations

### 📝 06. Advanced Patterns
- Init containers
- Sidecar containers
- Jobs and CronJobs
- DaemonSets
- Helm charts (introduction)

---

## Learning Goals

After this module:
- You can deploy and debug your own applications
- You understand resource limits and avoid OOMKills
- You know when to use StatefulSets vs Deployments
- You can read someone else's YAML and understand it
- You know the common mistakes and how to avoid them

---

**Previous:** [03. Kubernetes Fundamentals](../03-kubernetes-fundamentals/)  
**Next:** [05. K9s Terminal UX](../05-k9s/)
