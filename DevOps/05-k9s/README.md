# Module 05: K9s - Terminal UI for Kubernetes

> **K9s is to kubectl what htop is to ps — a better way to interact with Kubernetes**

---

## What Is K9s?

K9s is a terminal-based UI to manage Kubernetes clusters. It's **drastically faster** than typing kubectl commands repeatedly.

```
kubectl get pods -n production
kubectl describe pod api-7d8f9c-xyz -n production  
kubectl logs api-7d8f9c-xyz -n production -f
kubectl delete pod api-7d8f9c-xyz -n production

vs.

k9s
:pods
/api
<enter>
<shift+f>  # Follow logs
<ctrl+k>   # Delete
```

---

## Topics Covered

### ✅ [01. K9s Basics](./01-k9s-basics.md)
- Installation (Fedora: `sudo dnf install k9s`)
- Navigation keys
- Filtering and searching
- Context switching

### ✅ [02. Resource Management](./02-resource-management.md)
- Viewing pods, deployments, services
- Editing resources live
- Deleting resources safely
- Scaling deployments

### ✅ [03. Debugging Workflows](./03-debugging-workflows.md)
- Viewing logs (live tail)
- Describing resources
- Executing into pods
- Port forwarding
- Viewing events

### ✅ [04. Advanced Features](./04-advanced-features.md)
- Custom views (skins)
- Aliases
- Benchmarking
- RBAC visualization

---

## Why K9s Matters

**Scenario:** Your API is down in production.

**With kubectl:**
```bash
kubectl get pods -n production | grep api
kubectl describe pod api-xyz -n production
kubectl logs api-xyz -n production
# Copy pod name, paste, typo, retry
# 2 minutes to see logs
```

**With K9s:**
```
k9s
:pods production
/api
<enter>
<shift+l>
# 10 seconds to see logs
```

**Speed matters when production is on fire.**

---

## Installation

```bash
# Fedora
sudo dnf install k9s

# Or via Go
go install github.com/derailed/k9s@latest

# Verify
k9s version
```

---

## Quick Reference

| Key | Action |
|-----|--------|
| `:pods` | View pods |
| `:deploy` | View deployments |
| `:svc` | View services |
| `/` | Filter |
| `<enter>` | Describe |
| `<shift+l>` | View logs |
| `<shift+f>` | Follow logs |
| `<ctrl+k>` | Delete |
| `<e>` | Edit YAML |
| `<s>` | Shell into pod |
| `<shift+p>` | Port forward |
| `<ctrl+d>` | Delete |
| `:ctx` | Switch context |
| `:ns` | Switch namespace |
| `?` | Help |

---

**Previous:** [04. Kubernetes for Developers](../04-kubernetes-for-developers/)  
**Next:** [06. CI/CD Fundamentals](../06-ci-cd-fundamentals/)
