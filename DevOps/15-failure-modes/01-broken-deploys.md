# Broken Deploys

> **The fastest way to take down production is a bad deploy. The second fastest is a bad deploy without rollback capabilities. Every disaster here has happened to real companies.**

---

## 🟢 The `latest` Tag Disaster

### What Happens

```
Timeline:
  Monday 10:00 — Team A pushes my-app:latest (v1.2.0, working)
  Monday 14:00 — Team B pushes my-app:latest (v1.3.0, broken)
  Monday 14:05 — Kubernetes restarts a pod (node maintenance)
  Monday 14:06 — Pod pulls my-app:latest → gets v1.3.0 (broken!)
  Monday 14:07 — Another pod restarts → also gets v1.3.0
  Monday 14:30 — Half the pods run v1.2.0, half run v1.3.0
  Monday 14:35 — Errors spike, nobody knows why
  Monday 15:00 — "Wait, what version are we running?!"

Root cause:
  → latest is a MUTABLE tag — it points to different images over time
  → Pods started at different times run DIFFERENT code
  → kubectl describe pod shows "image: my-app:latest" for ALL pods
  → No way to know which version each pod actually runs
```

### The Fix

```yaml
# BAD
image: my-app:latest

# GOOD — use Git SHA or semantic version
image: my-app:a1b2c3d4
image: my-app:v1.2.0

# BEST — use image digest (immutable, can't be overwritten)
image: my-app@sha256:abc123def456...

# In CI/CD pipeline
docker build -t my-app:${{ github.sha }} .
docker push my-app:${{ github.sha }}
kubectl set image deployment/app app=my-app:${{ github.sha }}
```

```yaml
# Also set imagePullPolicy correctly
containers:
  - name: app
    image: my-app:v1.2.0
    imagePullPolicy: Always  # Always check registry
    # IfNotPresent with :latest = recipe for inconsistency
```

---

## 🟢 Missing Health Checks

### What Happens Without Health Checks

```
Scenario: App starts but takes 30 seconds to connect to database

Without health checks:
  1. Pod starts
  2. Kubernetes immediately sends traffic to it
  3. App receives requests but can't reach DB yet
  4. Users get 500 errors for 30 seconds
  5. During rolling deploy: ALL users hit errors as each new pod starts

Without liveness probe:
  1. App gets into deadlocked state
  2. Process is running, port is open
  3. But no requests are being processed
  4. Kubernetes thinks pod is healthy
  5. Pod stays running forever, serving nothing
  6. Users report "the app is slow" (it's actually dead)
```

### Proper Health Check Configuration

```yaml
containers:
  - name: app
    image: my-app:v1.2.0
    ports:
      - containerPort: 3000
    
    # READINESS: "Can this pod handle traffic?"
    # Fails → pod removed from Service endpoints (no traffic)
    readinessProbe:
      httpGet:
        path: /health/ready
        port: 3000
      initialDelaySeconds: 5    # Wait before first check
      periodSeconds: 10         # Check every 10s
      failureThreshold: 3       # 3 failures → unready
      successThreshold: 1       # 1 success → ready again
    
    # LIVENESS: "Is this pod alive?"
    # Fails → pod KILLED and restarted
    livenessProbe:
      httpGet:
        path: /health/live
        port: 3000
      initialDelaySeconds: 15   # Wait longer than readiness
      periodSeconds: 20         # Check every 20s
      failureThreshold: 3       # 3 failures → restart pod
    
    # STARTUP: "Has this pod finished starting?"
    # While failing → readiness and liveness probes are disabled
    startupProbe:
      httpGet:
        path: /health/live
        port: 3000
      periodSeconds: 5
      failureThreshold: 30      # 30 × 5s = 150s max startup time
```

### Health Check Endpoints

```typescript
// Good health check implementation
app.get('/health/live', (req, res) => {
  // Am I alive? Can I respond at all?
  // DON'T check dependencies here — if DB is down,
  // restarting THIS pod won't fix it
  res.status(200).json({ status: 'alive' });
});

app.get('/health/ready', async (req, res) => {
  // Am I ready to serve traffic?
  // CHECK dependencies here
  try {
    await db.query('SELECT 1');
    await redis.ping();
    res.status(200).json({ status: 'ready' });
  } catch (error) {
    // Not ready — remove from load balancer, but don't restart
    res.status(503).json({ status: 'not ready', error: error.message });
  }
});
```

---

## 🟡 Bad Rollout Strategies

### The Big Bang Deploy

```
What happens:
  1. kubectl apply replaces ALL pods at once
  2. All old pods terminated immediately
  3. New pods starting up, not ready yet
  4. ZERO pods serving traffic for 30-60 seconds
  5. Complete outage during every deploy

Fix: Never use Recreate strategy for user-facing services
```

### Rolling Update Misconfiguration

```yaml
# BAD — too aggressive
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 50%   # Half your pods go down at once
      maxSurge: 0           # No extra pods = very slow rollout

# GOOD — conservative
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0     # Never have fewer pods than desired
      maxSurge: 25%         # Spin up 25% more, then drain old
  minReadySeconds: 30       # Wait 30s after ready before continuing
```

### No Rollback Plan

```bash
# Disaster: new version deployed, everything is on fire
# If you don't know how to rollback...

# Kubernetes rollback is built in:
kubectl rollout undo deployment/app

# But ONLY if you use declarative deployments:
kubectl rollout history deployment/app
# REVISION  CHANGE-CAUSE
# 1         Initial deploy
# 2         Update to v1.2.0
# 3         Update to v1.3.0 (broken!)

kubectl rollout undo deployment/app --to-revision=2
# Instantly back to v1.2.0

# This DOESN'T work if you:
#   → Delete and recreate deployments (no history)
#   → Use Helm without --atomic (partial state)
#   → Apply manifests that reference external resources that changed
```

---

## 🟡 Resource Limits Missing

```yaml
# BAD — no limits
containers:
  - name: app
    image: my-app:v1.2.0
    # Memory leak → pod uses ALL node memory
    # Other pods on same node get OOM killed
    # One bad deploy takes down 10 other services

# GOOD — proper limits
containers:
  - name: app
    image: my-app:v1.2.0
    resources:
      requests:
        memory: "256Mi"   # Scheduler uses this for placement
        cpu: "100m"
      limits:
        memory: "512Mi"   # OOM killed if exceeds this
        cpu: "500m"        # Throttled if exceeds this
```

---

## 🔴 The Cascading Failure

```
Scenario: A deploy with a subtle bug

Timeline:
  09:00 — Deploy v2.0 (has memory leak, uses 10MB more per request)
  09:15 — Pod memory growing slowly
  09:30 — First pod hits memory limit → OOM killed → restarted
  09:31 — Remaining pods take more traffic → leak faster
  09:33 — Second pod OOM killed
  09:35 — Third pod OOM killed → only 1 pod left
  09:36 — Last pod overwhelmed → OOM killed
  09:37 — All pods in CrashLoopBackOff
  09:38 — Complete outage

Why rollback didn't help:
  → No alerting on memory growth
  → No alerting on pod restarts
  → Nobody noticed until ALL pods were down
  → By then, CrashLoopBackOff backoff was 5 minutes per pod

Prevention:
  → Alert on pod restarts: increase(kube_pod_container_status_restarts_total[1h]) > 3
  → Alert on memory growth trends
  → Set maxUnavailable: 0 (never have fewer pods than desired)
  → Use PodDisruptionBudget
```

```yaml
# PodDisruptionBudget — prevent too many pods going down
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: app-pdb
spec:
  minAvailable: 2  # Always keep at least 2 pods running
  # OR: maxUnavailable: 1  # At most 1 pod down at a time
  selector:
    matchLabels:
      app: my-app
```

---

## 🔴 Deploy Checklist

```
Before every production deploy:

□ Image tagged with specific version (not :latest)
□ Health checks configured (readiness + liveness)
□ Resource limits set (memory + CPU)
□ Rolling update strategy with maxUnavailable: 0
□ PodDisruptionBudget in place
□ Rollback command tested and documented
□ Monitoring dashboard open during deploy
□ Alert on error rate spike configured
□ Database migrations backward-compatible
□ Feature flags for risky changes
```

---

**Next:** [02. Leaking Secrets](./02-leaking-secrets.md)  
**Up:** [README](./README.md)
