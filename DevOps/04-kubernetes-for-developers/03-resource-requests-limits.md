# Resource Requests and Limits — Avoiding OOMKills

> **Understanding CPU throttling and memory limits — the difference between your app running smoothly and mysteriously crashing**

---

## 🟢 The Mental Model: Resource Guarantees vs Caps

```
requests = "Guaranteed minimum" (what scheduler uses)
limits   = "Maximum allowed" (enforced by cgroup)
```

**Analogy:** Hotel room reservation
- **Request** = "I need a room with at least 1 bed" (guaranteed)
- **Limit** = "Fire code max capacity: 4 people" (hard cap)

---

## 🟢 CPU vs Memory: Different Behaviors

| Resource | Compressible? | What Happens When Exceeded? |
|----------|---------------|----------------------------|
| **CPU** | ✅ Yes | **Throttled** (slows down, doesn't crash) |
| **Memory** | ❌ No | **OOMKilled** (process terminated) |

**Key difference:** You can "squeeze" CPU (slow down), but you can't "squeeze" memory.

---

## 🟢 Resource Requests: Scheduler Decisions

```yaml
resources:
  requests:
    cpu: "500m"      # 0.5 CPU cores
    memory: "512Mi"  # 512 megabytes
```

**What the scheduler does:**

1. Finds nodes with **at least** 500m CPU available
2. Finds nodes with **at least** 512Mi memory available
3. Schedules pod to a node that meets both

**Units:**
- **CPU**: `1` = 1 core, `500m` = 0.5 cores (m = millicores)
- **Memory**: `Ki`, `Mi`, `Gi` (1024-based) or `K`, `M`, `G` (1000-based)

### Example: Cluster Scheduling

```
Node 1: 4 CPU cores, 16Gi memory
├─ Pod A: requests 1 CPU, 2Gi    (running)
├─ Pod B: requests 2 CPU, 4Gi    (running)
└─ Available: 1 CPU, 10Gi

New Pod C: requests 2 CPU, 8Gi
❌ Can't schedule (needs 2 CPU, node only has 1 available)

Node 2: 4 CPU cores, 16Gi memory
├─ No pods running
└─ Available: 4 CPU, 16Gi

New Pod C: requests 2 CPU, 8Gi
✅ Can schedule here!
```

**Pod C goes to Node 2.**

---

## 🟡 Resource Limits: Runtime Enforcement

```yaml
resources:
  limits:
    cpu: "1"        # Maximum 1 CPU core
    memory: "1Gi"   # Maximum 1GB memory
```

### CPU Limits: Throttling

**What happens:**

```yaml
resources:
  requests:
    cpu: "500m"
  limits:
    cpu: "1"
```

```
Normal load:    Uses 500m CPU  → No throttling
Traffic spike:  Tries to use 1.5 CPU → Throttled to 1 CPU max
                (Process slows down, but doesn't crash)
```

**How to detect throttling:**

```bash
kubectl top pod myapp-xyz
# NAME        CPU(cores)   MEMORY(bytes)   
# myapp-xyz   1000m        512Mi

# If CPU is at exactly the limit for extended periods, you're being throttled
```

**Check throttling metrics:**

```bash
# Get pod cgroup
kubectl get pod myapp-xyz -o jsonpath='{.status.containerStatuses[0].containerID}'

# On the node (requires node access)
cat /sys/fs/cgroup/cpu/kubepods/.../cpu.stat
# nr_throttled: 1234      ← Number of times throttled
# throttled_time: 5000000000  ← Nanoseconds spent throttled
```

### Memory Limits: OOMKilled

**What happens:**

```yaml
resources:
  limits:
    memory: "1Gi"
```

```
Normal: Uses 500Mi → Fine
Growth: Uses 1.2Gi → ❌ KILLED BY KERNEL (exit code 137)
```

**Symptoms:**

```bash
kubectl get pods
# NAME        READY   STATUS      RESTARTS
# myapp-xyz   0/1     OOMKilled   3

kubectl describe pod myapp-xyz
# Last State:     Terminated
#   Reason:       OOMKilled
#   Exit Code:    137
```

**Why it's killed:** Linux kernel's OOM (Out Of Memory) killer terminates the process.

---

## 🔴 Common Mistake: No Limits = Noisy Neighbor Problem

```yaml
# BAD: No limits
spec:
  containers:
  - name: app
    image: myapp:1.0
    # No resources specified!
```

**What happens:**

```
Node has 4 CPUs, 16Gi memory

Pod A (your app, no limits): Uses 3.5 CPUs, 14Gi memory
Pod B (someone else's app): Can't get resources, starved
Pod C (database): Slowed down due to resource contention

❌ Your pod is a "noisy neighbor"
```

**Fix:** Always set limits.

---

## 🟡 The QoS Classes (Quality of Service)

Kubernetes assigns QoS classes based on requests/limits:

### 1. Guaranteed (Highest Priority)

```yaml
resources:
  requests:
    cpu: "1"
    memory: "1Gi"
  limits:
    cpu: "1"        # Same as request
    memory: "1Gi"   # Same as request
```

**Characteristics:**
- Requests == Limits for ALL containers
- Highest priority (last to be evicted)
- Use for critical workloads (databases, stateful apps)

### 2. Burstable (Medium Priority)

```yaml
resources:
  requests:
    cpu: "500m"
    memory: "512Mi"
  limits:
    cpu: "2"        # Higher than request
    memory: "2Gi"   # Higher than request
```

**Characteristics:**
- Requests < Limits
- Can burst above requests (up to limits)
- Medium priority (evicted after BestEffort)
- Use for most applications

### 3. BestEffort (Lowest Priority)

```yaml
# No resources specified at all
spec:
  containers:
  - name: app
    image: myapp:1.0
    # No requests or limits
```

**Characteristics:**
- No requests or limits
- Lowest priority (first to be evicted when node runs out of resources)
- Use for batch jobs, non-critical workloads

### Eviction Order (Node Under Pressure)

```
1. BestEffort pods killed first
2. Burstable pods using more than requests
3. Burstable pods using less than requests
4. Guaranteed pods (only if absolutely necessary)
```

---

## 🔴 Right-Sizing: How Much to Request?

### Method 1: Start Small, Monitor, Adjust

```yaml
# Week 1: Conservative guess
resources:
  requests:
    cpu: "100m"
    memory: "128Mi"
  limits:
    cpu: "500m"
    memory: "512Mi"
```

**Monitor:**

```bash
# Check actual usage
kubectl top pod myapp-xyz

# Install metrics-server first if not available
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

**Week 2: Adjust based on data**

```yaml
# If using 300m CPU and 256Mi memory consistently
resources:
  requests:
    cpu: "400m"      # 300m + 33% buffer
    memory: "384Mi"  # 256Mi + 50% buffer
  limits:
    cpu: "1"         # Allow bursting
    memory: "768Mi"  # 2x request
```

### Method 2: Load Testing

```bash
# Run load test
hey -n 10000 -c 100 http://myapp-service

# Monitor during load
kubectl top pod -l app=myapp --watch
```

**Record peak usage, add buffer:**
- Peak CPU: 800m → Request: 1000m, Limit: 2000m
- Peak memory: 600Mi → Request: 768Mi, Limit: 1Gi

### Method 3: Use VPA (Vertical Pod Autoscaler)

**VPA automatically recommends resource values.**

```bash
# Install VPA
git clone https://github.com/kubernetes/autoscaler.git
cd autoscaler/vertical-pod-autoscaler
./hack/vpa-up.sh

# Create VPA for your deployment
cat <<EOF | kubectl apply -f -
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: myapp-vpa
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: myapp
  updateMode: "Off"  # Just recommend, don't auto-update
EOF

# Check recommendations
kubectl describe vpa myapp-vpa
```

**Output:**
```
Recommendation:
  Container Recommendations:
    Container Name: app
    Lower Bound:
      Cpu:     100m
      Memory:  128Mi
    Target:
      Cpu:     500m     ← Use this for requests
      Memory:  512Mi
    Upper Bound:
      Cpu:     2
      Memory:  2Gi      ← Use this for limits
```

---

## 🔴 War Story: The Memory Leak OOMKill Loop

> *"Our Node.js API kept restarting every 2 hours. kubectl describe showed OOMKilled. We increased memory limit from 512Mi → 1Gi → 2Gi. Still crashed after 4 hours. Turns out: memory leak in a dependency. We were treating symptoms, not the cause. Fixed the leak, went back to 512Mi limit."*

**Lessons:**
1. OOMKilled doesn't always mean "increase limit"
2. Use profiling tools (pprof for Go, heapdump for Node.js)
3. Monitor memory growth over time

---

## 🟡 CPU Requests: The Hidden Cost

### Problem: Over-Requesting CPU Wastes Money

```yaml
# Your app only uses 100m CPU
# But you request 1000m "to be safe"
resources:
  requests:
    cpu: "1000m"
```

**Cluster impact:**
- Scheduler thinks node has less capacity
- More nodes needed (costs $$$$)
- But actual CPU usage is low (wasted money)

**Example:**

```
10 nodes × 4 CPUs = 40 CPUs total

100 pods × 1 CPU request = 100 CPUs requested
❌ Need 25 nodes!

Reality: 100 pods × 0.1 CPU actual = 10 CPUs used
✅ Could fit on 3 nodes if requests were accurate
```

**Fix:** Request what you actually need (with small buffer).

---

## 🎯 Best Practices

### ✅ DO:

1. **Always set requests and limits**
   ```yaml
   resources:
     requests:
       cpu: "500m"
       memory: "512Mi"
     limits:
       cpu: "1"
       memory: "1Gi"
   ```

2. **Requests = typical usage + 20-30% buffer**
   - Actual: 400m CPU → Request: 500m

3. **Limits = 2-3x requests (allow bursting)**
   - Request: 500m CPU → Limit: 1-1.5 CPU

4. **Memory limits close to requests** (memory leaks are bugs, not features)
   - Request: 512Mi → Limit: 768Mi or 1Gi

5. **Use QoS Guaranteed for stateful workloads**
   ```yaml
   # Databases, caches
   resources:
     requests:
       cpu: "2"
       memory: "4Gi"
     limits:
       cpu: "2"
       memory: "4Gi"
   ```

### ❌ DON'T:

1. **Don't omit resources** (creates BestEffort pods)

2. **Don't over-request** (wastes cluster capacity)
   ```yaml
   # BAD: App uses 100m, requesting 4 CPUs "just in case"
   requests:
     cpu: "4"
   ```

3. **Don't set CPU limits too low** (causes throttling)
   ```yaml
   # BAD: limits < typical usage
   limits:
     cpu: "100m"  # But app needs 500m normally
   ```

4. **Don't set memory limits too high for leaky apps**
   ```yaml
   # BAD: Hiding memory leaks with huge limits
   limits:
     memory: "32Gi"  # For an API that should use 512Mi
   ```

---

## ✅ Hands-On Exercise

### Task: Right-Size a Deployment

**1. Deploy app with conservative resources:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: demo
  template:
    metadata:
      labels:
        app: demo
    spec:
      containers:
      - name: app
        image: nginx
        resources:
          requests:
            cpu: "100m"
            memory: "64Mi"
          limits:
            cpu: "200m"
            memory: "128Mi"
```

```bash
kubectl apply -f demo-app.yaml
```

**2. Monitor usage:**

```bash
# Check current usage
kubectl top pod -l app=demo

# Output:
# NAME                CPU(cores)   MEMORY(bytes)
# demo-app-abc       2m           8Mi
# demo-app-def       2m           8Mi
# demo-app-ghi       2m           8Mi

# Nginx uses very little resources!
```

**3. Generate load:**

```bash
# Install hey (HTTP load generator)
go install github.com/rakyll/hey@latest

# Get service IP
kubectl expose deployment demo-app --port=80

# Generate load
hey -n 10000 -c 50 http://demo-app
```

**4. Monitor under load:**

```bash
kubectl top pod -l app=demo --watch

# Peak usage:
# demo-app-abc    50m    32Mi
```

**5. Right-size based on data:**

```yaml
resources:
  requests:
    cpu: "75m"       # Peak 50m + 50% buffer
    memory: "64Mi"   # Peak 32Mi + 100% buffer
  limits:
    cpu: "200m"      # Allow bursting
    memory: "128Mi"  # 2x request
```

---

## 📚 Summary

| Concept | Purpose | Effect |
|---------|---------|--------|
| **Requests** | Scheduling decision | Guaranteed minimum |
| **Limits** | Runtime enforcement | Hard cap |
| **CPU limit exceeded** | Throttling | App slows down |
| **Memory limit exceeded** | OOMKill | App killed (exit 137) |

**QoS Classes:**
- **Guaranteed**: requests == limits (critical workloads)
- **Burstable**: requests < limits (most apps)
- **BestEffort**: no resources (batch jobs)

**Right-Sizing Formula:**
```
Request = Typical usage × 1.3
Limit = Request × 2 (CPU), Request × 1.5-2 (memory)
```

**Monitoring:**
```bash
kubectl top pod POD_NAME
kubectl describe pod POD_NAME | grep -A 5 "Limits\|Requests"
```

---

**Previous:** [02. Debugging Pods and Containers](./02-debugging-pods.md)  
**Next:** [04. Application Configuration](./04-application-configuration.md)  
**Module:** [04. Kubernetes for Developers](./README.md)
