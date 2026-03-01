# K9s Resource Management

> **Viewing, editing, scaling, and deleting Kubernetes resources — all from a single terminal window**

---

## 🟢 Viewing Resources

### Pods View

```
:pods         # or :po
```

**What you see:**

```
NAMESPACE  NAME              READY  STATUS    RESTARTS  AGE     IP            NODE
default    api-7d8f9c-abc    1/1    Running   0         5m      10.244.1.5    node-1
default    api-7d8f9c-def    1/1    Running   2         5m      10.244.1.6    node-1
default    web-4b6e3a-ghi    0/1    Error     5         10m     10.244.2.3    node-2
```

**Toggle wide view:** `Ctrl+W` (shows more columns like IP, Node)

**Toggle all namespaces:** Press `0` (zero)

### Deployments View

```
:deploy       # or :dp
```

```
NAMESPACE  NAME    READY  UP-TO-DATE  AVAILABLE  AGE
default    api     2/2    2           2          10m
default    web     1/3    3           1          10m    ← Something wrong!
```

**From here you can:**
- `Enter` → Describe deployment
- `e` → Edit YAML
- `s` → Scale replicas
- `r` → Restart rolling update

### Services View

```
:svc
```

```
NAMESPACE  NAME         TYPE           CLUSTER-IP   EXTERNAL-IP    PORT(S)
default    api-svc      ClusterIP      10.96.0.10   <none>         80/TCP
default    web-svc      LoadBalancer   10.96.0.20   34.123.45.67   80:31234/TCP
default    kubernetes   ClusterIP      10.96.0.1    <none>         443/TCP
```

### Other Common Views

```
:cm          # ConfigMaps
:sec         # Secrets
:ing         # Ingress
:pvc         # PersistentVolumeClaims
:no          # Nodes
:ev          # Events
:hpa         # HorizontalPodAutoscalers
:jobs        # Jobs
:cj          # CronJobs
:rs          # ReplicaSets
:ds          # DaemonSets
:sts         # StatefulSets
:ep          # Endpoints
:np          # NetworkPolicies
```

---

## 🟡 Editing Resources Live

### Edit a Deployment

```
:deploy
# Select deployment
e            # Opens YAML in $EDITOR (default: vim)
```

**What happens:**
1. K9s opens the resource YAML in your editor
2. You make changes (e.g., update image tag, change env vars)
3. Save and quit (`:wq` in vim)
4. K9s applies the changes to the cluster

**Example: Change replica count**

```yaml
# In the editor, change:
spec:
  replicas: 3    # ← Change from 2 to 3
```

Save → K9s applies → Deployment scales to 3 pods.

### Edit a ConfigMap

```
:cm
# Select configmap
e
```

**Change values, save, and the ConfigMap is updated.**

**Remember:** Pods using this ConfigMap need a restart to see changes!

### Edit a Secret

```
:sec
# Select secret
e
```

**K9s decodes base64 automatically!** You edit the plaintext values.

---

## 🟡 Scaling Deployments

### Method 1: Scale Dialog

```
:deploy
# Select deployment
s            # Opens scale dialog
```

```
┌─ Scale ────────────────────┐
│                            │
│  Current Replicas: 2       │
│  New Replicas: [5]         │ ← Type desired count
│                            │
│  [OK]  [Cancel]            │
└────────────────────────────┘
```

### Method 2: Edit YAML

```
:deploy
# Select deployment
e
# Change spec.replicas in editor
# Save and quit
```

**Scaling down to 0:**

```
:deploy
# Select deployment
s
# Set replicas to 0
```

**Use case:** Temporarily stop a service without deleting it.

---

## 🟡 Deleting Resources

### Delete a Pod

```
:pods
# Select pod
Ctrl+K       # Kill/delete
```

**Confirmation dialog:**

```
┌─ Confirm ────────────────────────┐
│                                  │
│  Delete pod api-7d8f9c-abc?      │
│                                  │
│  [OK]  [Cancel]                  │
└──────────────────────────────────┘
```

**What happens when you delete a pod managed by a Deployment:**
- Pod is deleted
- ReplicaSet notices desired count != actual count
- New pod is created automatically
- **This is the standard way to "restart" a pod!**

### Force Delete

```
Ctrl+K
# If pod is stuck in Terminating, select Force Delete
```

Equivalent to: `kubectl delete pod NAME --grace-period=0 --force`

### Delete a Deployment

```
:deploy
# Select deployment
Ctrl+K
```

**Warning:** This deletes the Deployment AND all its pods!

### Drain Delete (Bulk)

Select multiple pods with `Space`, then `Ctrl+K` to delete all selected.

---

## 🟡 Viewing YAML

```
# Select any resource
y            # View full YAML definition
```

**Useful for:**
- Seeing the complete resource definition
- Understanding what's configured
- Copying YAML for use elsewhere

**Save YAML to file:** `Ctrl+S` → saves to `/tmp/k9s-xxx.yml`

---

## 🔴 Resource Sorting and Columns

### Sort by Column

Press `Shift+` followed by the column number:

```
:pods
Shift+1      # Sort by namespace
Shift+2      # Sort by name
Shift+5      # Sort by status
Shift+6      # Sort by restarts (find crashy pods!)
Shift+7      # Sort by age
```

### Custom Columns

Toggle wide view with `Ctrl+W` to see additional columns:
- IP address
- Node name
- Container images
- Resource requests/limits

---

## 🟡 Benchmarking Pods

K9s has a built-in HTTP benchmarking tool:

```
:pods
# Select a pod with an HTTP service
b            # Benchmark
```

```
┌─ Bench ─────────────────────────┐
│                                 │
│  URL: http://localhost:8080     │
│  Concurrency: 2                │
│  Requests: 200                  │
│  Method: GET                    │
│                                 │
│  [OK]  [Cancel]                 │
└─────────────────────────────────┘
```

**Output:**
```
Total:      2.5s
Slowest:    0.15s
Fastest:    0.01s
Average:    0.05s
Requests/sec: 80

Status code distribution:
  [200] 200 responses
```

---

## ✅ Hands-On Exercise

### Task: Manage Resources with K9s

**1. Start K9s and deploy a sample app:**

```bash
# In another terminal
kubectl create deployment nginx --image=nginx --replicas=3
kubectl expose deployment nginx --port=80
```

**2. In K9s:**

```
# View pods
:pods
# Filter for nginx
/nginx

# Describe a pod
Enter

# Go back
Esc

# Scale deployment
:deploy
# Select nginx
s
# Change to 5 replicas

# View pods again
:pods
/nginx
# Should see 5 pods

# Delete one pod
# Select a pod
Ctrl+K
# Confirm
# Watch new pod auto-created by ReplicaSet

# Edit deployment
:deploy
# Select nginx
e
# Change nginx image tag to nginx:alpine
# Save and quit
# Watch rolling update in :pods view

# View events
:events
# See scaling and update events

# Clean up
:deploy
# Select nginx
Ctrl+K
```

---

## 📚 Summary

| Operation | Key | Notes |
|-----------|-----|-------|
| View YAML | `y` | Full resource definition |
| Edit | `e` | Opens in $EDITOR |
| Scale | `s` | Deployments only |
| Delete | `Ctrl+K` | With confirmation |
| Sort | `Shift+N` | By column number |
| Wide view | `Ctrl+W` | More columns |
| Save YAML | `Ctrl+S` | To /tmp/ |
| Select multiple | `Space` | For bulk operations |
| Benchmark | `b` | HTTP bench |

**Key workflow:** `:resource` → `/filter` → select → action key

---

**Previous:** [01. K9s Basics](./01-k9s-basics.md)  
**Next:** [03. Debugging Workflows](./03-debugging-workflows.md)  
**Module:** [05. K9s](./README.md)
