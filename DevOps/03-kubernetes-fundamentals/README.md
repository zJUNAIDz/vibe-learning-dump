# Module 03: Kubernetes Fundamentals

> **Understanding Kubernetes from first principles — no magic, just control loops and reconciliation**

---

## What You'll Learn

This module teaches you **how Kubernetes actually works**, not just how to use it.

By the end:
- You'll understand the control plane architecture
- You'll know why Kubernetes makes certain decisions
- You'll be able to debug issues at the cluster level
- You'll think in "desired state" vs "actual state"

---

## Module Structure

### ✅ [01. Why Kubernetes Exists](./01-why-kubernetes-exists.md)
- The problems Kubernetes solves
- Desired state vs actual state
- Control loops and reconciliation
- Architecture overview (control plane vs data plane)

### ✅ [02. Core Concepts: Pods, Deployments, Services](./02-core-concepts-pods-deployments-services.md)
- Pods: the atomic unit
- ReplicaSets and Deployments
- Services: stable networking
- Rolling updates and rollbacks
- Labels and selectors

### 📝 03. ConfigMaps, Secrets, and Volumes
- Configuration management
- Secret handling (and why base64 isn't encryption)
- Persistent storage
- Volume types (emptyDir, hostPath, PVC)

### 📝 04. Health Checks and Probes
- Liveness vs readiness vs startup probes
- Why health checks matter
- Common patterns and anti-patterns

### 📝 05. Ingress and Load Balancing
- How Ingress controllers work
- NGINX Ingress controller
- TLS termination
- Path-based routing

### 📝 06. Namespaces and Resource Quotas
- Isolating resources
- Resource limits and requests
- LimitRanges
- Network policies

---

## Prerequisites

- Completed Module 01 (Linux & Systems)
- Completed Module 02 (Containers)
- Access to a Kubernetes cluster (minikube, kind, or cloud)

---

## Hands-On Setup

### Local Kubernetes (Choose One)

**minikube (recommended for learning):**
```bash
# Fedora
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
minikube start
```

**kind (Kubernetes in Docker):**
```bash
# Install
go install sigs.k8s.io/kind@latest

# Create cluster
kind create cluster
```

**kubectl:**
```bash
# Fedora
sudo dnf install kubernetes-client
kubectl version
```

---

## What This Module Does NOT Cover

- Cloud-specific Kubernetes (EKS, GKE, AKS) → covered in Module 12
- Helm → Module 04
- Advanced scheduling (taints, tolerations, affinity) → Module 04
- Operators and CRDs → Module 04
- Service meshes (Istio, Linkerd) → Beyond scope

---

## Learning Path

1. **Read the concepts** in order
2. **Deploy examples** to your local cluster
3. **Break things** (delete pods, nodes, see what happens)
4. **Watch reconciliation** in action
5. **Debug failures** (use `describe`, `logs`, `events`)

---

## Key Takeaways

After this module, you should be able to:
- Explain how Kubernetes maintains desired state
- Deploy stateless applications
- Configure and debug services
- Understand when and why pods restart
- Read YAML manifests and understand what they do

---

**Next Module:** [04. Kubernetes for Developers →](../04-kubernetes-for-developers/)
