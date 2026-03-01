# Namespaces and Resource Quotas

> **Multi-tenancy and resource isolation in Kubernetes — preventing one team from starving another**

---

## 🟢 The Mental Model: Namespaces are Virtual Clusters

**Imagine a shared Kubernetes cluster with 10 teams:**

```
Without namespaces:
┌─────────────────────────────────────────────┐
│  All resources in one bucket                │
│  - dev-api, prod-api, staging-api (messy!)  │
│  - Name conflicts                           │
│  - Hard to apply policies                   │
│  - One team can starve others               │
└─────────────────────────────────────────────┘

With namespaces:
┌──────────────┬──────────────┬──────────────┐
│ dev          │ staging      │ production   │
│ - api        │ - api        │ - api        │
│ - db         │ - db         │ - db         │
│ (isolated)   │ (isolated)   │ (isolated)   │
└──────────────┴──────────────┴──────────────┘
```

**Key Concept:** Namespaces provide **logical isolation**, not security boundaries.

---

## 🟢 Default Namespaces

```bash
kubectl get namespaces
```

**Output:**
```
NAME              STATUS   AGE
default           Active   30d    # Your workloads go here (if no namespace specified)
kube-system       Active   30d    # Kubernetes system components
kube-public       Active   30d    # Public resources (readable by all)
kube-node-lease   Active   30d    # Node heartbeat data
```

**Rule of thumb:** Don't use `default` in production. Create explicit namespaces.

---

## 🟢 Creating Namespaces

### Method 1: Imperative

```bash
kubectl create namespace dev
kubectl create namespace staging
kubectl create namespace production
```

### Method 2: Declarative

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: dev
  labels:
    environment: development
---
apiVersion: v1
kind: Namespace
metadata:
  name: staging
  labels:
    environment: staging
---
apiVersion: v1
kind: Namespace
metadata:
  name: production
  labels:
    environment: production
```

```bash
kubectl apply -f namespaces.yaml
```

---

## 🟢 Working with Namespaces

### Deploying to a Namespace

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: dev  # Explicit namespace
spec:
  replicas: 2
  selector:
    matchLabels:
      app: api
  template:
    metadata:
      labels:
        app: api
    spec:
      containers:
      - name: api
        image: myapi:1.0
```

### Viewing Resources in a Namespace

```bash
# List resources in specific namespace
kubectl get pods -n dev
kubectl get services -n staging
kubectl get deployments -n production

# List resources in all namespaces
kubectl get pods --all-namespaces
kubectl get pods -A  # Short form
```

### Setting Default Namespace

```bash
# Set default namespace for current context
kubectl config set-context --current --namespace=dev

# Now all commands use 'dev' by default
kubectl get pods  # Same as: kubectl get pods -n dev
```

### Deleting a Namespace (WARNING: Destructive)

```bash
# This deletes the namespace AND all resources inside it!
kubectl delete namespace dev

# Everything in 'dev' namespace is gone:
# - Pods, Deployments, Services, ConfigMaps, Secrets, etc.
```

---

## 🟡 DNS and Cross-Namespace Communication

### Service DNS Format

```
<service-name>.<namespace>.svc.cluster.local
```

**Example services:**

```yaml
# In 'dev' namespace
apiVersion: v1
kind: Service
metadata:
  name: api
  namespace: dev
# DNS: api.dev.svc.cluster.local
---
# In 'staging' namespace
apiVersion: v1
kind: Service
metadata:
  name: database
  namespace: staging
# DNS: database.staging.svc.cluster.local
```

### Same Namespace Communication

```yaml
# Pod in 'dev' namespace
apiVersion: v1
kind: Pod
metadata:
  name: web
  namespace: dev
spec:
  containers:
  - name: app
    image: myapp:1.0
    env:
    - name: API_URL
      value: "http://api"  # Short form (same namespace)
      # Full form: http://api.dev.svc.cluster.local
```

### Cross-Namespace Communication

```yaml
# Pod in 'dev' namespace calling service in 'staging' namespace
apiVersion: v1
kind: Pod
metadata:
  name: web
  namespace: dev
spec:
  containers:
  - name: app
    image: myapp:1.0
    env:
    - name: DATABASE_URL
      value: "postgres://database.staging.svc.cluster.local:5432/mydb"
      # Must use full DNS name (with namespace)
```

**Key Point:** Services can call across namespaces by default (use NetworkPolicies to restrict).

---

## 🟡 Resource Quotas: Preventing Resource Starvation

### The Problem

Team A deploys 100 pods with no resource limits → uses all cluster CPU/memory → Team B's pods can't start.

**Solution:** Resource Quotas.

### Creating a ResourceQuota

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: dev-quota
  namespace: dev
spec:
  hard:
    # Compute resources
    requests.cpu: "10"        # Max 10 CPU cores requested
    requests.memory: "20Gi"   # Max 20GB memory requested
    limits.cpu: "20"          # Max 20 CPU cores limit
    limits.memory: "40Gi"     # Max 40GB memory limit
    
    # Object counts
    pods: "50"                # Max 50 pods
    services: "20"            # Max 20 services
    persistentvolumeclaims: "10"  # Max 10 PVCs
    configmaps: "50"          # Max 50 ConfigMaps
    secrets: "50"             # Max 50 Secrets
```

### Apply Quota

```bash
kubectl apply -f resource-quota.yaml
```

### Check Quota Usage

```bash
kubectl get resourcequota -n dev
kubectl describe resourcequota dev-quota -n dev
```

**Output:**
```
Name:                   dev-quota
Namespace:              dev
Resource                Used   Hard
--------                ----   ----
limits.cpu              8      20
limits.memory           16Gi   40Gi
persistentvolumeclaims  3      10
pods                    12     50
requests.cpu            4      10
requests.memory         8Gi    20Gi
services                5      20
```

---

## 🟡 LimitRange: Default Resource Limits

**Problem:** Developers forget to set resource requests/limits → pods use unlimited resources.

**Solution:** LimitRange sets default values.

### Creating a LimitRange

```yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: dev-limits
  namespace: dev
spec:
  limits:
  # Container defaults
  - type: Container
    default:  # Default limits (if not specified)
      cpu: "500m"
      memory: "512Mi"
    defaultRequest:  # Default requests (if not specified)
      cpu: "250m"
      memory: "256Mi"
    max:  # Maximum allowed
      cpu: "2"
      memory: "2Gi"
    min:  # Minimum allowed
      cpu: "100m"
      memory: "128Mi"
  
  # Pod limits
  - type: Pod
    max:
      cpu: "4"
      memory: "8Gi"
  
  # PersistentVolumeClaim limits
  - type: PersistentVolumeClaim
    max:
      storage: "10Gi"
    min:
      storage: "1Gi"
```

### Apply LimitRange

```bash
kubectl apply -f limit-range.yaml
```

### How It Works

**Before LimitRange:**
```yaml
# Developer creates pod without resource specs
apiVersion: v1
kind: Pod
metadata:
  name: myapp
  namespace: dev
spec:
  containers:
  - name: app
    image: myapp:1.0
    # No resources specified!
```

**After LimitRange:**
```bash
kubectl get pod myapp -n dev -o yaml | grep -A 4 resources
```

**Output (defaults applied automatically):**
```yaml
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi
```

---

## 🔴 Resource Requests vs Limits

### Mental Model

```
requests = "I need at least this much"
limits   = "I must not use more than this"
```

**Example:**
```yaml
resources:
  requests:
    cpu: "250m"     # 0.25 CPU cores
    memory: "256Mi"
  limits:
    cpu: "500m"     # 0.5 CPU cores
    memory: "512Mi"
```

**What happens:**

```
┌────────────────────────────────────────────────┐
│  Node has 4 CPU cores, 16GB memory             │
│                                                │
│  Scheduler uses REQUESTS to decide placement   │
│  - Finds node with at least 250m CPU available │
│  - Finds node with at least 256Mi RAM available│
│                                                │
│  Once pod is running:                          │
│  - Can use between 250m-500m CPU (bursting)    │
│  - Can use between 256Mi-512Mi memory          │
│  - If exceeds CPU limit: throttled             │
│  - If exceeds memory limit: OOMKilled 💥       │
└────────────────────────────────────────────────┘
```

### CPU Throttling

```yaml
limits:
  cpu: "1"  # 1 core
```

**What happens:** If pod tries to use >1 core, it's throttled (slows down).  
**Effect:** Pod runs slower, but doesn't crash.

### Memory OOMKill

```yaml
limits:
  memory: "512Mi"
```

**What happens:** If pod uses >512Mi, Linux kernel kills it.  
**Effect:** Pod crashes and restarts.

```bash
kubectl get pods
# NAME        READY   STATUS      RESTARTS
# myapp-xyz   0/1     OOMKilled   3
```

**How to diagnose:**
```bash
kubectl describe pod myapp-xyz
# Last State:     Terminated
#   Reason:       OOMKilled
#   Exit Code:    137
```

---

## 🔴 War Story: The Runaway Pod

> *"A developer deployed a pod with no resource limits. It had a memory leak. Over 2 hours, it consumed 60GB of memory, starving all other pods on that node. 15 services went down. We had to cordon the node and manually kill the pod. Afterward, we enforced LimitRanges on all namespaces."*

**Lesson:** Always set resource limits. Use LimitRanges to enforce defaults.

---

## 🟡 Network Policies: Namespace-Level Firewalls

**Default behavior:** All pods can talk to all pods (even across namespaces).

**Problem:** Production pods shouldn't talk to dev pods.

**Solution:** NetworkPolicies.

### Example: Deny All Traffic by Default

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all
  namespace: production
spec:
  podSelector: {}  # Apply to all pods in namespace
  policyTypes:
  - Ingress
  - Egress
```

**Effect:** All pods in `production` namespace can't receive or send traffic (locked down).

### Example: Allow Only Within Namespace

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-same-namespace
  namespace: production
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector: {}  # Allow from any pod in same namespace
```

**Effect:** Pods in `production` can only talk to other pods in `production`.

### Example: Allow Specific Cross-Namespace Communication

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-monitoring
  namespace: production
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    ports:
    - protocol: TCP
      port: 9090
```

**Effect:** Pods in `production` allow traffic from `monitoring` namespace on port 9090.

**Note:** NetworkPolicies require a network plugin that supports them (Calico, Cilium, Weave).

---

## 🎯 Real-World Example: Multi-Environment Setup

```yaml
# namespaces.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: dev
  labels:
    environment: development
---
apiVersion: v1
kind: Namespace
metadata:
  name: staging
  labels:
    environment: staging
---
apiVersion: v1
kind: Namespace
metadata:
  name: production
  labels:
    environment: production
---
# resource-quota-dev.yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: dev-quota
  namespace: dev
spec:
  hard:
    requests.cpu: "5"
    requests.memory: "10Gi"
    limits.cpu: "10"
    limits.memory: "20Gi"
    pods: "30"
---
# resource-quota-production.yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: prod-quota
  namespace: production
spec:
  hard:
    requests.cpu: "50"
    requests.memory: "100Gi"
    limits.cpu: "100"
    limits.memory: "200Gi"
    pods: "200"
---
# limit-range.yaml (apply to all namespaces)
apiVersion: v1
kind: LimitRange
metadata:
  name: default-limits
  namespace: dev
spec:
  limits:
  - type: Container
    default:
      cpu: "500m"
      memory: "512Mi"
    defaultRequest:
      cpu: "100m"
      memory: "128Mi"
    max:
      cpu: "2"
      memory: "4Gi"
```

**Deploy:**
```bash
kubectl apply -f namespaces.yaml
kubectl apply -f resource-quota-dev.yaml
kubectl apply -f resource-quota-production.yaml

# Apply LimitRange to each namespace
for ns in dev staging production; do
  kubectl apply -f limit-range.yaml -n $ns
done
```

---

## ✅ Hands-On Exercise

### Task: Create Multi-Tenant Cluster Simulation

**1. Create namespaces:**

```bash
kubectl create namespace team-a
kubectl create namespace team-b
```

**2. Create ResourceQuotas:**

```yaml
# quota-team-a.yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: team-a-quota
  namespace: team-a
spec:
  hard:
    requests.cpu: "2"
    requests.memory: "4Gi"
    limits.cpu: "4"
    limits.memory: "8Gi"
    pods: "10"
```

```yaml
# quota-team-b.yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: team-b-quota
  namespace: team-b
spec:
  hard:
    requests.cpu: "1"
    requests.memory: "2Gi"
    limits.cpu: "2"
    limits.memory: "4Gi"
    pods: "5"
```

```bash
kubectl apply -f quota-team-a.yaml
kubectl apply -f quota-team-b.yaml
```

**3. Try to exceed quota:**

```yaml
# deployment-team-a.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: team-a
spec:
  replicas: 15  # Exceeds quota (max 10 pods)
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: app
        image: nginx
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
```

```bash
kubectl apply -f deployment-team-a.yaml

# Check status
kubectl get deployments -n team-a
# DESIRED   CURRENT   READY
# 15        10        10      # Only 10 pods created (quota limit)

kubectl describe resourcequota -n team-a
# pods: 10/10 (quota reached!)
```

**4. Test cross-namespace communication:**

```bash
# Deploy service in team-a
kubectl run web -n team-a --image=nginx --port=80
kubectl expose pod web -n team-a --port=80

# Try to access from team-b
kubectl run test -n team-b --image=curlimages/curl -it --rm -- \
  curl web.team-a.svc.cluster.local

# Should work (cross-namespace calls allowed by default)
```

---

## 🧩 Common Mistakes

### ❌ Mistake 1: No Resource Limits in Production

**Problem:** One pod consumes all node resources.

**Fix:** Always use ResourceQuotas and LimitRanges.

### ❌ Mistake 2: Using `default` Namespace

**Problem:** Hard to organize, no isolation, can't apply policies.

**Fix:** Create explicit namespaces.

### ❌ Mistake 3: Forgetting to Specify Namespace

```bash
# You think you're working in 'production'
kubectl delete deployment api

# You just deleted 'api' in 'default' namespace!
```

**Fix:** Always use `-n namespace` or set default context.

### ❌ Mistake 4: Setting Requests == Limits

```yaml
resources:
  requests:
    cpu: "1"
  limits:
    cpu: "1"  # No bursting allowed
```

**Problem:** Pod can't burst during traffic spikes.

**Better:**
```yaml
resources:
  requests:
    cpu: "500m"   # Guaranteed
  limits:
    cpu: "2"      # Can burst up to 2 cores
```

---

## 📚 Summary

| Concept | Purpose | Effect |
|---------|---------|--------|
| **Namespace** | Logical isolation | Organize resources, apply policies |
| **ResourceQuota** | Limit resource usage per namespace | Prevent one team from starving others |
| **LimitRange** | Default resource limits | Enforce good practices |
| **NetworkPolicy** | Namespace-level firewall | Control traffic between namespaces |

**Best Practices:**
1. ✅ Use explicit namespaces (dev, staging, production)
2. ✅ Apply ResourceQuotas to all namespaces
3. ✅ Use LimitRanges to set default resource limits
4. ✅ Set requests < limits (allow bursting)
5. ✅ Use NetworkPolicies for production namespaces
6. ✅ Label namespaces for easier management

**Namespace Workflow:**
```bash
# Create namespace
kubectl create namespace my-app

# Set as default
kubectl config set-context --current --namespace=my-app

# Apply policies
kubectl apply -f resource-quota.yaml -n my-app
kubectl apply -f limit-range.yaml -n my-app

# Deploy app
kubectl apply -f app.yaml  # Uses 'my-app' namespace
```

---

**Previous:** [05. Ingress and Load Balancing](./05-ingress-load-balancing.md)  
**Next:** [Module 04: Kubernetes for Developers](../04-kubernetes-for-developers/README.md)  
**Module:** [03. Kubernetes Fundamentals](./README.md)
