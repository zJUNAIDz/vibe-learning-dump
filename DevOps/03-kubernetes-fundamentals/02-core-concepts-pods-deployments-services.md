# Core Concepts: Pods, Deployments, Services

🟢 **Fundamentals** → 🟡 **Intermediate**

---

## The Building Blocks

Kubernetes has many resource types, but these three are **fundamental**:

```
Pod         → Smallest deployable unit (one or more containers)
Deployment  → Manages pods (replicas, updates, rollbacks)
Service     → Network access to pods (load balancing, DNS)
```

---

## Pods: The Atomic Unit

### What Is a Pod?

A **pod** is a group of one or more containers that:
- Share the same network namespace (same IP, localhost works)
- Share the same storage volumes
- Are scheduled together on the same node
- Have a shared lifecycle (created/destroyed together)

**Most common:** 1 container per pod.

---

### Why Pods, Not Just Containers?

**Scenario:** You have an app container and a log-shipping sidecar.

Without pods:
```
Container A (app) → IP: 172.17.0.2
Container B (logs) → IP: 172.17.0.3

Problem: How does B access A's logs? Complicated networking.
```

With pods:
```
Pod (A + B) → IP: 172.17.0.2

Container A writes logs to /var/log/app
Container B reads from /var/log/app (shared volume)
Both containers can talk via localhost
```

---

### Pod Manifest

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-pod
  labels:
    app: nginx
spec:
  containers:
  - name: nginx
    image: nginx:1.25-alpine
    ports:
    - containerPort: 80
```

**Deploy:**
```bash
kubectl apply -f pod.yaml
kubectl get pods
kubectl describe pod nginx-pod
```

---

### Multi-Container Pod (Sidecar Pattern)

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-with-sidecar
spec:
  containers:
  # Main application
  - name: app
    image: myapp:latest
    volumeMounts:
    - name: logs
      mountPath: /var/log/app
  
  # Sidecar (log shipper)
  - name: log-shipper
    image: fluentd:latest
    volumeMounts:
    - name: logs
      mountPath: /var/log/app
      readOnly: true
  
  volumes:
  - name: logs
    emptyDir: {}
```

**Both containers:**
- Share the same IP
- Can talk via `localhost`
- Share the `/var/log/app` directory

---

### Pod Lifecycle

```
Pending    → Scheduled but not yet running (pulling image, etc.)
Running    → At least one container is running
Succeeded  → All containers exited successfully (for batch jobs)
Failed     → All containers exited, at least one failed
Unknown    → Can't get pod status (node unreachable)
```

**Check status:**
```bash
kubectl get pods
kubectl describe pod <name>
kubectl logs <name>
```

---

### Why You Never Deploy Pods Directly

**Problem:** Pods are **ephemeral** (temporary).

```
Pod crashes → Gone forever
Node dies → Pod lost
You delete it → No replacement
```

**Solution:** Use **controllers** (Deployments, StatefulSets, etc.) that manage pods for you.

---

## ReplicaSets: Maintaining Pod Count

A **ReplicaSet** ensures a specified number of pod replicas are running.

```yaml
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: nginx-rs
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.25-alpine
```

**What it does:**
- Watches for pods with label `app=nginx`
- If count < 3 → creates more
- If count > 3 → deletes extras

**You almost never create ReplicaSets directly.** Use Deployments instead.

---

## Deployments: The Right Way to Run Apps

A **Deployment** manages ReplicaSets, which manage Pods.

```
Deployment → ReplicaSet → Pods
```

**Why Deployments > ReplicaSets:**
- Rolling updates
- Rollbacks
- Version history
- Declarative updates

---

### Deployment Manifest

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.25-alpine
        ports:
        - containerPort: 80
```

**Deploy:**
```bash
kubectl apply -f deployment.yaml
kubectl get deployments
kubectl get rs  # ReplicaSets
kubectl get pods
```

---

### Rolling Updates

Change the image version:
```yaml
spec:
  containers:
  - name: nginx
    image: nginx:1.26-alpine  # Changed from 1.25
```

```bash
kubectl apply -f deployment.yaml

# Watch rollout
kubectl rollout status deployment/nginx-deployment

# What happened:
# 1. Create new ReplicaSet (nginx:1.26)
# 2. Scale new RS up (1 pod)
# 3. Scale old RS down (1 pod)
# 4. Repeat until all pods updated
```

**Zero downtime!**

---

### Rollback

```bash
# Oh no, new version is broken!
kubectl rollout undo deployment/nginx-deployment

# Go back to specific revision
kubectl rollout history deployment/nginx-deployment
kubectl rollout undo deployment/nginx-deployment --to-revision=2
```

---

### Scaling

```bash
# Imperatively (quick test)
kubectl scale deployment/nginx-deployment --replicas=10

# Declaratively (production)
# Edit deployment.yaml: replicas: 10
kubectl apply -f deployment.yaml
```

---

### Update Strategies

#### RollingUpdate (Default)
```yaml
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1  # Max pods unavailable during update
      maxSurge: 1        # Max extra pods created during update
```

**Example:** 3 replicas
```
Old pods: [A][B][C]
Step 1:   [A][B][C][D]  (Create 1 new, maxSurge=1)
Step 2:   [B][C][D]     (Delete 1 old, maxUnavailable=1)
Step 3:   [B][C][D][E]
Step 4:   [C][D][E]
Step 5:   [C][D][E][F]
Step 6:   [D][E][F]
```

#### Recreate
```yaml
spec:
  strategy:
    type: Recreate
```

**Behavior:**
1. Delete ALL old pods
2. Create ALL new pods

**Downtime:** Yes (all pods down between steps 1 and 2)

**Use case:** Rarely (only if rolling update causes issues, e.g., schema migration)

---

## Services: Stable Network Access

**Problem:** Pods have dynamic IPs. They come and go.

```
Pod 1 → 10.0.1.5  (crashes, new pod gets 10.0.1.8)
Pod 2 → 10.0.1.6
Pod 3 → 10.0.1.7

How do clients reach "the app"?
```

**Solution:** Kubernetes **Service**.

---

### What Is a Service?

A **Service** provides:
- **Stable IP address** (never changes)
- **DNS name** (`my-service.default.svc.cluster.local`)
- **Load balancing** across all matching pods

```
            ┌─────────────┐
            │  Service    │
            │  10.96.0.1  │
            └──────┬──────┘
                   │
         Load balances to:
        ┌──────┬──────┬──────┐
        │ Pod1 │ Pod2 │ Pod3 │
        └──────┴──────┴──────┘
```

---

### Service Types

| Type | Purpose | Example |
|------|---------|---------|
| **ClusterIP** | Internal only (default) | Backend API |
| **NodePort** | Exposes on each node's IP | Testing |
| **LoadBalancer** | Cloud load balancer | Production web app |
| **ExternalName** | Maps to external DNS | Legacy database |

---

### ClusterIP Service (Default)

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
spec:
  type: ClusterIP
  selector:
    app: nginx  # Targets pods with this label
  ports:
  - protocol: TCP
    port: 80         # Service port
    targetPort: 80   # Container port
```

**Access:**
```bash
# From another pod in the cluster
curl http://nginx-service.default.svc.cluster.local

# Or just
curl http://nginx-service
```

**Not accessible from outside cluster.**

---

### NodePort Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-nodeport
spec:
  type: NodePort
  selector:
    app: nginx
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
    nodePort: 30080  # Must be 30000-32767
```

**Access:**
```
http://<any-node-ip>:30080
```

**Use case:** Testing, dev environments. **Not for production** (random high port, no real load balancing).

---

### LoadBalancer Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-lb
spec:
  type: LoadBalancer
  selector:
    app: nginx
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
```

**What happens (on AWS, GCP, Azure):**
1. Kubernetes tells cloud provider: "Create a load balancer"
2. Cloud creates ELB/ALB/NLB (AWS), Load Balancer (GCP), etc.
3. Load balancer routes traffic to NodePorts
4. NodePorts route to pods

**Access:**
```bash
kubectl get svc nginx-lb
# EXTERNAL-IP: a1234.us-east-1.elb.amazonaws.com

curl http://a1234.us-east-1.elb.amazonaws.com
```

**Use case:** Production apps.

---

### Headless Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-headless
spec:
  clusterIP: None  # Headless
  selector:
    app: nginx
  ports:
  - port: 80
```

**What it does:**
- **No load balancing**
- DNS returns **all pod IPs** (not service IP)

**Use case:** StatefulSets, databases (you want to connect to specific pods).

---

## Selectors and Labels

**Labels** are key-value pairs attached to resources.

```yaml
metadata:
  labels:
    app: nginx
    env: production
    tier: frontend
```

**Selectors** match labels.

```yaml
# Service
spec:
  selector:
    app: nginx  # Targets ALL pods with label app=nginx
```

**Powerful feature:** Change labels → change which pods Service routes to (blue/green deployments).

---

## Putting It All Together

**Typical workflow:**

1. **Create Deployment** (manages pods, handles updates)
2. **Create Service** (exposes pods via stable IP/DNS)
3. **Deploy:** `kubectl apply -f deployment.yaml -f service.yaml`
4. **Access:** `curl http://service-name`

**Example:**

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
spec:
  replicas: 3
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
        image: myapi:v1.0
        ports:
        - containerPort: 8080
---
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: api
spec:
  type: ClusterIP
  selector:
    app: api
  ports:
  - port: 80
    targetPort: 8080
```

**Deploy:**
```bash
kubectl apply -f deployment.yaml -f service.yaml

# From another pod
curl http://api
```

---

## War Story: The Flapping Deployment

A team deployed a new version. Within 5 minutes, Kubernetes started **flapping** (constantly restarting pods).

**Investigation:**
```bash
kubectl get pods
# All pods in CrashLoopBackOff

kubectl logs <pod>
# Error: Cannot connect to database
```

**Root cause:**
The new version required a database migration, but the code tried to connect before migration completed.

**Lessons:**
1. **Health checks matter** (readiness probe should check DB connection)
2. **Migrations should run before deployment** (init containers or separate job)
3. **Rollback is your friend** (`kubectl rollout undo`)

---

## Key Takeaways

1. **Pods are ephemeral** — never deploy them directly
2. **Deployments manage pods** — use them for stateless apps
3. **Services provide stable networking** — pods come and go, Service IP doesn't
4. **ClusterIP for internal, LoadBalancer for external** — choose based on access pattern
5. **Labels and selectors** — glue that connects Deployments and Services
6. **Rolling updates by default** — zero-downtime deployments
7. **Rollback is trivial** — `kubectl rollout undo`

---

## Exercises

1. **Create a Deployment:**
   - Write YAML for a 3-replica nginx Deployment
   - Deploy it, watch pods spin up
   - Scale to 5 replicas

2. **Expose via Service:**
   - Create a ClusterIP Service for your Deployment
   - From another pod, `curl` the Service

3. **Rolling update:**
   - Change image version in Deployment
   - Watch rollout: `kubectl rollout status`
   - Rollback: `kubectl rollout undo`

4. **Label experiments:**
   - Add a label to a pod manually
   - Change Service selector to NOT match pods
   - Observe traffic stops

---

**Next:** [03. ConfigMaps, Secrets, and Volumes →](./03-configmaps-secrets-volumes.md)
