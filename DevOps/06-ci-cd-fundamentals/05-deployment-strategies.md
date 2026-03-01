# Deployment Strategies

> **How you update production matters as much as what you deploy. The right strategy is the difference between zero-downtime and "the site is down."**

---

## 🟢 Why Deployment Strategy Matters

### The Naive Approach (Don't Do This)

```bash
# "Just replace it"
kubectl delete deployment myapp
kubectl apply -f deployment.yaml
```

**What happens:**
1. All old pods killed instantly
2. New pods starting... (takes 30-60 seconds)
3. **Users see errors for 30-60 seconds**
4. If new version is broken, **no old version to fall back to**

**This is a Big Bang deployment.** All or nothing. No safety net.

### What We Actually Want

```
✅ Zero downtime — Users never see an error during deployment
✅ Gradual rollout — Test with a few users before everyone
✅ Instant rollback — Go back to the old version in seconds
✅ Observability — Know if the new version is healthy
```

---

## 🟢 Strategy 1: Rolling Update (Kubernetes Default)

### How It Works

```
Time 0: All pods running v1
┌─────┐ ┌─────┐ ┌─────┐
│ v1  │ │ v1  │ │ v1  │
└─────┘ └─────┘ └─────┘
Traffic: ──→ v1, v1, v1

Time 1: Start replacing one pod
┌─────┐ ┌─────┐ ┌─────┐
│ v2  │ │ v1  │ │ v1  │
└─────┘ └─────┘ └─────┘
Traffic: ──→ v2, v1, v1

Time 2: Replace another
┌─────┐ ┌─────┐ ┌─────┐
│ v2  │ │ v2  │ │ v1  │
└─────┘ └─────┘ └─────┘
Traffic: ──→ v2, v2, v1

Time 3: All replaced
┌─────┐ ┌─────┐ ┌─────┐
│ v2  │ │ v2  │ │ v2  │
└─────┘ └─────┘ └─────┘
Traffic: ──→ v2, v2, v2
```

**Key:** Old pods are only killed after new pods are ready. At no point are zero pods serving traffic.

### Kubernetes Configuration

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1          # Create 1 extra pod during update
      maxUnavailable: 0     # Never have fewer than 3 ready pods
  template:
    spec:
      containers:
        - name: myapp
          image: myapp:v2
          readinessProbe:    # CRITICAL for rolling updates
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
```

### Understanding maxSurge and maxUnavailable

```
replicas: 3

maxSurge: 1, maxUnavailable: 0
  → At most 4 pods (3 + 1 surge), at least 3 ready
  → Safest, slowest

maxSurge: 1, maxUnavailable: 1
  → At most 4 pods, at least 2 ready
  → Faster, but brief capacity drop

maxSurge: 3, maxUnavailable: 0
  → At most 6 pods, at least 3 ready
  → Fastest, but needs double the resources
```

### Why Readiness Probes Are Critical

```
Without readiness probe:
  1. New pod starts
  2. Kubernetes says "it's running!" (container started)
  3. Kubernetes sends traffic to it
  4. App is still initializing (loading config, warming cache)
  5. Users get 503 errors

With readiness probe:
  1. New pod starts
  2. Kubernetes checks /health every 5 seconds
  3. /health returns 503 (app still starting)
  4. Kubernetes does NOT send traffic yet
  5. After 15 seconds, /health returns 200
  6. NOW Kubernetes sends traffic
  7. Users see zero errors
```

### Rollback

```bash
# Something wrong? Roll back instantly:
kubectl rollout undo deployment/myapp

# Check rollout history:
kubectl rollout history deployment/myapp

# Roll back to specific revision:
kubectl rollout undo deployment/myapp --to-revision=3

# Watch rollback progress:
kubectl rollout status deployment/myapp
```

### Pros and Cons

| Pros | Cons |
|------|------|
| Zero-downtime | Both versions run simultaneously (mixed responses) |
| Built into Kubernetes | Rollback takes time (another rolling update) |
| No extra infrastructure | Database schema changes are tricky |
| Simple to configure | No traffic control (can't send 1% to v2) |

---

## 🟡 Strategy 2: Blue/Green Deployment

### How It Works

```
Blue (current production):
┌──────────────────────┐
│  v1  v1  v1          │  ← Receiving all traffic
└──────────────────────┘
         ↑
    Load Balancer
         ↓
┌──────────────────────┐
│  (empty)             │  ← Green environment idle
└──────────────────────┘


Step 1: Deploy v2 to Green:
┌──────────────────────┐
│  v1  v1  v1          │  ← Still receiving all traffic
└──────────────────────┘
         ↑
    Load Balancer
         ↓
┌──────────────────────┐
│  v2  v2  v2          │  ← Ready, tested, no traffic yet
└──────────────────────┘


Step 2: Switch traffic to Green:
┌──────────────────────┐
│  v1  v1  v1          │  ← Idle (kept for rollback)
└──────────────────────┘

    Load Balancer ───── switched!
         ↓
┌──────────────────────┐
│  v2  v2  v2          │  ← Receiving all traffic
└──────────────────────┘
```

### Kubernetes Implementation

```yaml
# Blue deployment (current production)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-blue
  labels:
    app: myapp
    version: blue
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
      version: blue
  template:
    metadata:
      labels:
        app: myapp
        version: blue
    spec:
      containers:
        - name: myapp
          image: myapp:v1.0.0

---
# Green deployment (new version)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-green
  labels:
    app: myapp
    version: green
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
      version: green
  template:
    metadata:
      labels:
        app: myapp
        version: green
    spec:
      containers:
        - name: myapp
          image: myapp:v2.0.0

---
# Service points to blue (production)
apiVersion: v1
kind: Service
metadata:
  name: myapp
spec:
  selector:
    app: myapp
    version: blue     # ← Switch this to "green" to swap
  ports:
    - port: 80
      targetPort: 8080
```

### Switching Traffic

```bash
# Test green directly (before switching)
kubectl port-forward deployment/myapp-green 8081:8080
curl http://localhost:8081/health    # Verify v2 works

# Switch traffic from blue to green
kubectl patch service myapp -p '{"spec":{"selector":{"version":"green"}}}'

# If something is wrong, switch back instantly
kubectl patch service myapp -p '{"spec":{"selector":{"version":"blue"}}}'

# After confirming green is stable, clean up blue
kubectl delete deployment myapp-blue
```

### Pros and Cons

| Pros | Cons |
|------|------|
| Instant rollback (switch selector) | Requires double the resources |
| Test new version before switching | All-or-nothing traffic switch |
| No mixed versions serving traffic | Database migrations still tricky |
| Simple to understand | More manual orchestration |

---

## 🟡 Strategy 3: Canary Deployment

### How It Works

```
Step 1: Deploy v2 to small subset
┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐
│ v1  │ │ v1  │ │ v1  │ │ v2  │
└─────┘ └─────┘ └─────┘ └─────┘
  25%      25%     25%     25%    ← Traffic split
                            ↑
                      "Canary" pod

Step 2: Monitor canary (error rates, latency)
  v2 error rate: 0.1% ✓
  v2 latency: 45ms ✓ (same as v1)
  → Canary is healthy!

Step 3: Increase canary traffic
┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐
│ v1  │ │ v2  │ │ v2  │ │ v2  │
└─────┘ └─────┘ └─────┘ └─────┘
  25%      25%     25%     25%
                            
Step 4: Full rollout
┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐
│ v2  │ │ v2  │ │ v2  │ │ v2  │
└─────┘ └─────┘ └─────┘ └─────┘
 100% v2
```

### Mental Model: The Canary in the Coal Mine

Coal miners brought canaries into mines. If toxic gas was present, the canary would die first, warning the miners.

**Your canary pod is the same.** It encounters production traffic first. If it dies (errors, crashes, high latency), you know v2 is bad before it reaches all users.

### Simple Kubernetes Canary (Replica-Based)

```yaml
# Stable deployment (v1) — most of the traffic
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-stable
spec:
  replicas: 9              # 90% of traffic
  template:
    metadata:
      labels:
        app: myapp          # Same label as canary
    spec:
      containers:
        - name: myapp
          image: myapp:v1.0.0

---
# Canary deployment (v2) — small slice of traffic
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-canary
spec:
  replicas: 1              # 10% of traffic
  template:
    metadata:
      labels:
        app: myapp          # Same label as stable
    spec:
      containers:
        - name: myapp
          image: myapp:v2.0.0

---
# Service sends traffic to BOTH (same label selector)
apiVersion: v1
kind: Service
metadata:
  name: myapp
spec:
  selector:
    app: myapp              # Matches both stable and canary
  ports:
    - port: 80
      targetPort: 8080
```

**Traffic split:** 9 stable pods + 1 canary pod = ~10% traffic to canary.

### Progressive Canary Rollout

```bash
# Start with 1 canary (10%)
kubectl scale deployment myapp-canary --replicas=1

# Monitor for 10 minutes...
# Check error rates, latency, CPU/memory

# Increase to 30%
kubectl scale deployment myapp-stable --replicas=7
kubectl scale deployment myapp-canary --replicas=3

# Monitor for 10 minutes...

# Increase to 50%
kubectl scale deployment myapp-stable --replicas=5
kubectl scale deployment myapp-canary --replicas=5

# Monitor for 10 minutes...

# Full rollout
kubectl scale deployment myapp-stable --replicas=0
kubectl scale deployment myapp-canary --replicas=10

# Clean up: rename canary to stable
kubectl set image deployment/myapp-stable myapp=myapp:v2.0.0
kubectl scale deployment myapp-stable --replicas=10
kubectl delete deployment myapp-canary
```

### Advanced Canary with Istio/Nginx

For precise traffic control (not just replica-based):

```yaml
# Istio VirtualService — traffic split by percentage
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: myapp
spec:
  hosts:
    - myapp
  http:
    - route:
        - destination:
            host: myapp
            subset: stable
          weight: 95        # 95% to stable
        - destination:
            host: myapp
            subset: canary
          weight: 5         # 5% to canary

---
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: myapp
spec:
  host: myapp
  subsets:
    - name: stable
      labels:
        version: v1
    - name: canary
      labels:
        version: v2
```

**Advantage:** Exact traffic percentages, independent of replica count.

### Automated Canary with Flagger

[Flagger](https://flagger.app/) automates the canary process:

```yaml
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: myapp
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: myapp
  progressDeadlineSeconds: 60
  analysis:
    interval: 1m              # Check every minute
    threshold: 5               # Max 5 failed checks
    maxWeight: 50              # Max 50% traffic to canary
    stepWeight: 10             # Increase by 10% each step
    metrics:
      - name: request-success-rate
        thresholdRange:
          min: 99              # Need 99% success rate
        interval: 1m
      - name: request-duration
        thresholdRange:
          max: 500             # Max 500ms latency
        interval: 1m
```

**Flagger automatically:**
1. Deploys canary with 10% traffic
2. Checks success rate (>99%) and latency (<500ms)
3. If healthy: increase to 20%, then 30%, ... up to 50%
4. If unhealthy: automatically rolls back
5. No human intervention needed

### Pros and Cons

| Pros | Cons |
|------|------|
| Gradual risk exposure | More complex to set up |
| Real production testing | Two versions run simultaneously |
| Automated rollback possible | Session affinity issues (user sees v1 then v2) |
| Metrics-driven decisions | Needs observability (monitoring) |

---

## 🟡 Strategy 4: Recreate (Big Bang)

### How It Works

```
Step 1: Kill all old pods
┌─────┐ ┌─────┐ ┌─────┐
│ v1  │ │ v1  │ │ v1  │  ← All terminated
└─────┘ └─────┘ └─────┘
         DOWNTIME! ☠️

Step 2: Start new pods
┌─────┐ ┌─────┐ ┌─────┐
│ v2  │ │ v2  │ │ v2  │  ← Starting...
└─────┘ └─────┘ └─────┘
         STILL DOWN! ☠️

Step 3: New pods ready
┌─────┐ ┌─────┐ ┌─────┐
│ v2  │ │ v2  │ │ v2  │  ← Ready ✓
└─────┘ └─────┘ └─────┘
```

```yaml
spec:
  strategy:
    type: Recreate    # Kill all, then create all
```

### When to Use Recreate

- **Development/staging environments** where downtime is acceptable
- **Applications that can't run two versions simultaneously** (e.g., database schema changes)
- **Single-instance applications** (some legacy apps)
- **GPU workloads** where you can't afford double resources

### Don't Use Recreate for Production Web Applications

---

## 🔴 Rollback vs Roll-Forward

### Rollback: Deploy the Previous Version

```bash
# "Oh no, v2 is broken! Go back to v1!"
kubectl rollout undo deployment/myapp
```

**When to rollback:**
- v2 has a critical bug affecting all users
- v2 crashes on startup
- v2 has a security vulnerability
- You need to stop the bleeding NOW

### Roll-Forward: Fix v2 and Deploy v3

```bash
# "v2 has a bug. Let's fix it and deploy v3."
git commit -m "fix: resolve null pointer in auth flow"
git push   # Triggers CI/CD → new image → v3 deployed
```

**When to roll forward:**
- The bug is small and understood
- A fix is quick (< 30 minutes)
- Rolling back has its own risks (data migration issues)
- v2 also contained critical security fixes you don't want to lose

### Decision Framework

```
Is production on fire?
  ├── YES → ROLLBACK immediately, investigate later
  └── NO → 
        Is the fix known and quick?
          ├── YES → ROLL FORWARD
          └── NO → ROLLBACK, investigate calmly
```

---

## 🟡 Database Migrations and Deployments

### The Hard Problem

```
Database schema (current):
  users: id, email, name

v2 adds a new column:
  users: id, email, name, avatar_url

If v1 pods and v2 pods run simultaneously (rolling update):
  v1 pod: SELECT id, email, name FROM users        ← Works ✓
  v2 pod: SELECT id, email, name, avatar_url FROM users  ← Works ✓
  
This is fine! v1 ignores the new column.
```

**But what if v2 REMOVES a column?**

```
v2 removes 'name', adds 'first_name' and 'last_name':
  Users: id, email, first_name, last_name

During rolling update:
  v1 pod: SELECT id, email, name FROM users    ← CRASH (column doesn't exist!)
```

### The Solution: Expand-Contract Pattern

```
Step 1: EXPAND (v2) — Add new columns, keep old ones
  users: id, email, name, first_name, last_name
  v1 uses: name        ← Still works
  v2 uses: first_name, last_name  ← New columns

Step 2: MIGRATE — Copy data from 'name' to first/last
  UPDATE users SET first_name = split_part(name, ' ', 1),
                   last_name = split_part(name, ' ', 2);

Step 3: CONTRACT (v3) — Remove old column
  users: id, email, first_name, last_name
  Only v3 pods are running (v2 fully rolled out)
```

**Three deployments instead of one, but zero downtime.**

---

## 🎯 Strategy Comparison

| Strategy | Downtime | Risk | Complexity | Resource Cost | Best For |
|----------|----------|------|-----------|---------------|----------|
| Rolling Update | Zero | Medium | Low | 1x + surge | Most apps |
| Blue/Green | Zero | Low | Medium | 2x | Critical apps |
| Canary | Zero | Very Low | High | 1x + small | High-traffic apps |
| Recreate | Yes | High | Very Low | 1x | Dev/staging |

### Which Strategy Should You Use?

```
Start here:
  → Rolling Update (Kubernetes default)
  → It handles 90% of cases

If you need instant rollback:
  → Blue/Green

If you need gradual rollout with metrics:
  → Canary

If downtime is acceptable:
  → Recreate (only for non-production)
```

---

## ✅ Hands-On Exercise

### Simulate Deployment Strategies with Kubernetes

**Prerequisites:** A running Kubernetes cluster (minikube, kind, or similar).

**1. Rolling Update (Default):**

```bash
# Deploy v1
kubectl create deployment myapp --image=nginx:1.24 --replicas=3

# Watch pods
kubectl get pods -w &

# Update to v2 (rolling update happens automatically)
kubectl set image deployment/myapp nginx=nginx:1.25

# Watch the rolling update:
# Old pods terminate one by one
# New pods start one by one
# Traffic never drops to zero

# Rollback
kubectl rollout undo deployment/myapp

# Clean up
kubectl delete deployment myapp
```

**2. Blue/Green Simulation:**

```bash
# Deploy blue (v1)
kubectl create deployment myapp-blue --image=nginx:1.24 --replicas=3
kubectl expose deployment myapp-blue --port=80 --target-port=80 --name=myapp-service

# Deploy green (v2) — no traffic yet
kubectl create deployment myapp-green --image=nginx:1.25 --replicas=3

# Test green directly
kubectl port-forward deployment/myapp-green 8081:80 &
curl http://localhost:8081    # Verify v2 works

# Switch traffic to green
kubectl patch service myapp-service \
  -p '{"spec":{"selector":{"app":"myapp-green"}}}'

# Rollback to blue
kubectl patch service myapp-service \
  -p '{"spec":{"selector":{"app":"myapp-blue"}}}'

# Clean up
kubectl delete deployment myapp-blue myapp-green
kubectl delete service myapp-service
```

**3. Canary Simulation:**

```bash
# Deploy stable (v1)
kubectl create deployment myapp-stable --image=nginx:1.24 --replicas=9
kubectl label deployment myapp-stable role=stable

# Deploy canary (v2) — small replica count
kubectl create deployment myapp-canary --image=nginx:1.25 --replicas=1
kubectl label deployment myapp-canary role=canary

# Create service pointing to both
kubectl expose deployment myapp-stable --port=80 --name=myapp \
  --selector=app  # This is simplified

# Traffic split: 9 stable + 1 canary ≈ 90/10 split

# If canary is healthy, scale up
kubectl scale deployment myapp-canary --replicas=5
kubectl scale deployment myapp-stable --replicas=5

# Full rollout
kubectl scale deployment myapp-canary --replicas=10
kubectl scale deployment myapp-stable --replicas=0

# Clean up
kubectl delete deployment myapp-stable myapp-canary
kubectl delete service myapp
```

---

## 📚 Summary

| Concept | Key Takeaway |
|---------|-------------|
| **Rolling Update** | Kubernetes default, zero downtime, gradual replacement |
| **Blue/Green** | Two environments, instant switch, double resources |
| **Canary** | Gradual rollout, metrics-driven, lowest risk |
| **Recreate** | Kill all → start all, has downtime, simplest |
| **Rollback** | Go to previous version, for emergencies |
| **Roll-forward** | Fix and deploy new version, for small bugs |
| **Readiness Probes** | Essential for zero-downtime in all strategies |
| **Expand-Contract** | Safe database migrations during rolling deployments |

**Start with Rolling Update. Add Blue/Green or Canary when your app's criticality demands it.**

---

**Previous:** [04. Pipeline Stages](./04-pipeline-stages.md)  
**Next:** [06. CI/CD Best Practices](./06-cicd-best-practices.md)  
**Module:** [06. CI/CD Fundamentals](./README.md)
