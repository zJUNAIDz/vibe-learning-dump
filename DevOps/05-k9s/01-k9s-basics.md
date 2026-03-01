# K9s Basics

> **Installation, navigation, and core concepts — becoming 10x faster at Kubernetes operations**

---

## 🟢 What Is K9s?

K9s is a **terminal-based UI** for Kubernetes. Think of it as `htop` for your cluster.

**Why it exists:**

```bash
# Without K9s (kubectl)
kubectl get pods -n production                              # List pods
kubectl describe pod api-7d8f9c-xyz -n production           # Describe
kubectl logs api-7d8f9c-xyz -n production -f                # Logs
kubectl exec -it api-7d8f9c-xyz -n production -- /bin/sh    # Shell
kubectl delete pod api-7d8f9c-xyz -n production             # Delete

# With K9s
# :pods → /api → Enter → l → s → ctrl-k
# 5 keystrokes vs 5 long commands
```

---

## 🟢 Installation

### Fedora

```bash
sudo dnf install k9s
```

### Other Linux (Binary)

```bash
curl -sS https://webi.sh/k9s | sh
```

### Go Install

```bash
go install github.com/derailed/k9s@latest
```

### Verify

```bash
k9s version
```

### First Launch

```bash
# Uses your current kubectl context
k9s

# Specify context
k9s --context my-cluster

# Specify namespace
k9s -n production

# Read-only mode (safe for production!)
k9s --readonly
```

---

## 🟢 The K9s Interface

When you launch K9s, you see:

```
┌──────────────────────────────────────────────────┐
│  K9s - [context: minikube] [namespace: default]  │ ← Header
├──────────────────────────────────────────────────┤
│  NAMESPACE  NAME           READY  STATUS  AGE    │ ← Resource list
│  default    api-abc123     1/1    Running 5m     │
│  default    web-def456     1/1    Running 5m     │
│  default    db-ghi789      1/1    Running 5m     │
│                                                  │
│                                                  │
├──────────────────────────────────────────────────┤
│  <pod> Filter: ...                               │ ← Command bar
└──────────────────────────────────────────────────┘
```

---

## 🟢 Navigation: The Core Keybindings

### Moving Around

| Key | Action |
|-----|--------|
| `↑` / `↓` / `j` / `k` | Navigate resources |
| `Enter` | View/describe selected resource |
| `Esc` | Go back / cancel |
| `q` | Quit K9s |
| `?` | Help (shows all keybindings) |

### Switching Resources (Command Mode)

Press `:` (colon) to enter command mode:

```
:pods        → View pods (shortcut: :po)
:deploy      → View deployments (shortcut: :dp)
:svc         → View services
:ns          → View/switch namespaces
:ctx         → View/switch cluster contexts
:nodes       → View nodes (shortcut: :no)
:configmaps  → View configmaps (shortcut: :cm)
:secrets     → View secrets (shortcut: :sec)
:ingress     → View ingress resources (shortcut: :ing)
:events      → View events (shortcut: :ev)
:pvc         → View persistent volume claims
:hpa         → View horizontal pod autoscalers
:jobs        → View jobs
:cronjobs    → View cronjobs (shortcut: :cj)
:sa          → View service accounts
:rb          → View role bindings
:crb         → View cluster role bindings
```

**Pro tip:** Type `:` then start typing — K9s has autocomplete!

### Filtering

| Key | Action |
|-----|--------|
| `/` | Start filter (regex supported) |
| `/api` | Show only resources matching "api" |
| `/!running` | Show resources NOT matching "running" |
| `Esc` | Clear filter |

**Example:**
```
:pods
/crash        → Shows only pods with "crash" in name/status
/!Running     → Shows pods that are NOT Running
```

### Namespace Switching

```
:ns           → Shows all namespaces
Enter         → Switch to selected namespace

# Or use shortcut
:pods -n production    → Jump to pods in production namespace
```

**0 (zero)** toggles between "current namespace" and "all namespaces".

---

## 🟢 Actions on Resources

When you have a resource selected:

### Pod Actions

| Key | Action |
|-----|--------|
| `Enter` / `d` | Describe (detailed info) |
| `l` | View logs |
| `Shift+L` | View logs (previous container) |
| `s` | Shell into container |
| `Shift+F` | Port-forward |
| `e` | Edit YAML |
| `Ctrl+K` | Kill/delete |
| `y` | View YAML |

### Deployment Actions

| Key | Action |
|-----|--------|
| `Enter` / `d` | Describe |
| `e` | Edit YAML |
| `s` | Scale (change replicas) |
| `r` | Rollback / restart |
| `Ctrl+K` | Delete |

### General Actions (Work Everywhere)

| Key | Action |
|-----|--------|
| `c` | Copy resource name |
| `Ctrl+S` | Save YAML to file |
| `Ctrl+W` | Toggle wide view |

---

## 🟡 Context and Cluster Switching

If you manage multiple clusters:

```
:ctx              → List all contexts
Enter             → Switch to selected context
```

**Your kubeconfig contexts:**
```bash
# K9s reads from ~/.kube/config
kubectl config get-contexts
```

**Example workflow:**
```
:ctx
# Select "production-cluster"
# Now all views show production resources
:pods
# These are production pods
```

---

## 🟡 The Pulse View

Press `Ctrl+P` or type `:pulse` to see cluster health at a glance:

```
┌─ Pulse ──────────────────────────────────────┐
│                                              │
│  Cluster:  minikube                          │
│  K8s:      v1.28.0                          │
│  CPU:      ████████░░░░░  62%               │
│  MEM:      ██████░░░░░░░  45%               │
│                                              │
│  Pods:     Running: 12  Pending: 0  Failed: 1│
│  Deploy:   Ready: 5  NotReady: 0            │
│  Services: 8                                 │
│  Events:   Warnings: 2                      │
└──────────────────────────────────────────────┘
```

**Great for:** Quick health check when you first connect.

---

## 🟡 XRay View

Type `:xray deploy` to see hierarchical view:

```
Deployment: api
├── ReplicaSet: api-7d8f9c
│   ├── Pod: api-7d8f9c-abc12
│   │   └── Container: app (Running)
│   └── Pod: api-7d8f9c-def34
│       └── Container: app (Running)
├── Service: api-service
└── HPA: api-hpa
```

**Great for:** Understanding how Kubernetes resources relate to each other.

---

## 🟢 Hands-On Exercise

### Task: Navigate a Cluster with K9s

```bash
# 1. Start K9s
k9s

# 2. View all pods
:pods

# 3. Filter to specific app
/nginx

# 4. Describe a pod (select one, then press Enter)

# 5. View logs (press 'l')

# 6. Go back (Esc)

# 7. Switch to deployments
:deploy

# 8. View services
:svc

# 9. Check events
:events

# 10. View cluster health
:pulse

# 11. Quit
q
```

---

## 📚 Summary

| Task | kubectl | K9s |
|------|---------|-----|
| List pods | `kubectl get pods -n ns` | `:pods` |
| Describe pod | `kubectl describe pod NAME` | Select + `Enter` |
| View logs | `kubectl logs NAME -f` | Select + `l` |
| Shell into pod | `kubectl exec -it NAME -- sh` | Select + `s` |
| Delete pod | `kubectl delete pod NAME` | Select + `Ctrl+K` |
| Switch namespace | `kubectl config set-context --current --namespace=ns` | `:ns` → select |
| Port forward | `kubectl port-forward NAME 8080:80` | Select + `Shift+F` |

**K9s is consistently 5-10x faster than kubectl for interactive operations.**

---

**Next:** [02. Resource Management](./02-resource-management.md)  
**Module:** [05. K9s](./README.md)
