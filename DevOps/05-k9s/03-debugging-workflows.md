# K9s Debugging Workflows

> **When production is on fire, speed matters — debugging Kubernetes issues at the speed of thought**

---

## 🟢 The Debugging Mindset with K9s

**Traditional debugging (slow):**
```bash
kubectl get pods -n production | grep api
kubectl describe pod api-7d8f9c-abc -n production
kubectl logs api-7d8f9c-abc -n production --tail=100
kubectl logs api-7d8f9c-abc -n production --previous
kubectl exec -it api-7d8f9c-abc -n production -- /bin/sh
```

**K9s debugging (fast):**
```
k9s -n production
/api
Enter          → Describe
l              → Logs
p              → Previous logs
s              → Shell
```

**Time saved: 90 seconds → 10 seconds.**

---

## 🟢 Workflow 1: Pod Not Starting

**Scenario:** Deployment has 3 replicas, but pods are in Error/Pending/CrashLoop.

### Step 1: Find Problematic Pods

```
:pods
0                    # Show all namespaces
```

**Look for red/yellow status:**
```
NAMESPACE  NAME              READY  STATUS             RESTARTS  AGE
default    api-7d8f9c-abc    0/1    CrashLoopBackOff   5         10m  ← Problem!
default    api-7d8f9c-def    0/1    ImagePullBackOff   0         10m  ← Problem!
default    web-4b6e3a-ghi    1/1    Running            0         10m  ← OK
```

### Step 2: Describe the Pod

```
# Select problematic pod
Enter        # or 'd'
```

**K9s shows the describe output — scroll to Events section:**

```
Events:
  Type     Reason     Age    From               Message
  ----     ------     ----   ----               -------
  Warning  Failed     2m     kubelet            Failed to pull image "myapp:v999"
  Warning  BackOff    1m     kubelet            Back-off pulling image "myapp:v999"
```

**Diagnosis:** Wrong image tag `v999`.

### Step 3: Fix It

```
Esc          # Go back to pod list
:deploy      # Switch to deployments
# Select the deployment
e            # Edit YAML
```

**Change the image tag and save.**

### Step 4: Watch Recovery

```
:pods
/api
# Watch pods transition: Terminating → ContainerCreating → Running
```

---

## 🟢 Workflow 2: Viewing Logs

### Live Logs

```
:pods
# Select pod
l            # View logs
```

**Log view controls:**

| Key | Action |
|-----|--------|
| `0` | Toggle all containers |
| `w` | Toggle line wrap |
| `t` | Toggle timestamps |
| `s` | Toggle auto-scroll |
| `/` | Search/filter in logs |
| `Esc` | Exit log view |

### Previous Container Logs

When a pod has crashed and restarted:

```
:pods
# Select pod
p            # View previous container logs
```

**This is the equivalent of `kubectl logs POD --previous`.**

**Shows logs from the LAST crashed container — crucial for debugging CrashLoopBackOff.**

### Multi-Container Pod Logs

If a pod has multiple containers (sidecars, init containers):

```
:pods
# Select multi-container pod
l
```

**K9s prompts which container to view:**

```
┌─ Select Container ─────────┐
│                             │
│  > app                      │
│    sidecar-proxy            │
│    log-agent                │
│                             │
└─────────────────────────────┘
```

**Toggle all containers:** Press `0` in log view to see all containers interleaved.

### Filtering Logs

While in log view:

```
/error        # Show only lines with "error"
/panic        # Show only lines with "panic"
/timeout      # Show only lines with "timeout"
Esc           # Clear filter
```

**Regex supported:**

```
/error|warn|panic           # Multiple patterns
/\d{3}                      # Status codes
/connection.*refused        # Connection refused errors
```

---

## 🟡 Workflow 3: Shell into a Container

### Getting a Shell

```
:pods
# Select pod
s            # Shell
```

**K9s prompts for container (if multiple) and shell:**

```
┌─ Shell ──────────────────────┐
│                              │
│  Container: app              │
│  Shell: /bin/sh              │
│                              │
└──────────────────────────────┘
```

**You're now inside the container!**

```bash
# Check if app is listening
netstat -tlnp
ss -tlnp
curl localhost:8080/healthz

# Check environment variables
env | grep DATABASE
env | grep API

# Check filesystem
ls -la /app
cat /etc/config/app.properties

# Check DNS resolution
nslookup database-service
nslookup api-service.production.svc.cluster.local

# Check connectivity to other services
wget -O- http://database-service:5432
curl http://api-service/health

# Check processes
ps aux

# Check resource usage
top

# Exit shell
exit
```

**If container has no shell:**

Some minimal images (distroless, scratch) don't have `/bin/sh`.

**Solution:** Use ephemeral debug container (from kubectl):

```bash
kubectl debug POD_NAME -it --image=nicolaka/netshoot
```

---

## 🟡 Workflow 4: Port Forwarding

### Forward a Pod Port

```
:pods
# Select pod
Shift+F      # Port forward
```

```
┌─ Port Forward ────────────────┐
│                               │
│  Container Port: 8080         │
│  Local Port: [8080]           │ ← Enter local port
│  Address: localhost            │
│                               │
│  [OK]  [Cancel]               │
└───────────────────────────────┘
```

**Now access the app:** `http://localhost:8080`

### Forward a Service Port

```
:svc
# Select service
Shift+F
```

### View Active Port Forwards

```
:pf          # List active port forwards
```

```
NAMESPACE  NAME         CONTAINER  PORTS           AGE     STATUS
default    api-abc123   app        8080:8080       5m      Running
default    web-svc      web        3000:80         10m     Running
```

**Stop port forward:** Select and `Ctrl+K`.

---

## 🟡 Workflow 5: Viewing Events

### Cluster Events

```
:events      # or :ev
```

```
NAMESPACE  TYPE     REASON              OBJECT                 MESSAGE                              AGE
default    Normal   Scheduled           pod/api-abc            Successfully assigned to node-1       5m
default    Normal   Pulling             pod/api-abc            Pulling image "myapi:v1"              5m
default    Normal   Pulled              pod/api-abc            Successfully pulled image             4m
default    Normal   Created             pod/api-abc            Created container app                 4m
default    Normal   Started             pod/api-abc            Started container app                 4m
default    Warning  FailedScheduling    pod/web-def            0/3 nodes are available: 3 Insufficient cpu  2m
default    Warning  Unhealthy           pod/db-ghi             Readiness probe failed                1m
```

**Filter events:**

```
/warning          # Show only warnings
/failed           # Show failures
/oom              # OOM events
```

**Sort by time:**

```
Shift+7           # Sort by age (newest first)
```

---

## 🔴 Workflow 6: Full Incident Response

**Scenario:** "API is returning 502 errors in production."

### Minute 0: Assess the Situation

```bash
k9s -n production --readonly    # Read-only to avoid accidents!
```

```
:pulse           # Cluster health overview
```

**Look for:**
- Pods not ready
- High restart counts
- Warning events

### Minute 1: Find Broken Pods

```
:pods
/api             # Filter for API pods
```

**Check STATUS column:**
- `Running` but `0/1` READY → readiness probe failing
- `CrashLoopBackOff` → app is crashing
- `OOMKilled` → out of memory

### Minute 2: Check Logs

```
# Select broken pod
l                # Live logs
/error           # Filter for errors
```

**Common findings:**
```
ERROR: Connection refused to database:5432
ERROR: Out of memory
PANIC: runtime error: invalid memory address
ERROR: context deadline exceeded
```

### Minute 3: Check Dependencies

```
Esc              # Back to pods
:svc             # Check services
/database        # Find database service

:pods
/postgres        # Check database pods
```

**Is the database running?**

### Minute 4: Check Events

```
:events
/warning         # Show warnings only
```

**Look for:**
- OOM kills
- Readiness probe failures
- Node pressure events
- Scheduling failures

### Minute 5: Check Resources

```
:pods
# Select pod
d                # Describe
```

**Scroll to:**
- Containers section → resource requests/limits
- Conditions section → Ready, Initialized
- Events section → recent events

### Minute 6: Fix or Rollback

**Option A: Rollback deployment**

```
:deploy
# Select api deployment
r                # Restart/rollback
```

**Option B: Scale up**

```
:deploy
# Select api deployment
s
# Increase replicas from 3 to 5
```

**Option C: Edit config**

```
:cm              # ConfigMaps
# Select relevant configmap
e                # Edit
# Fix configuration
# Save

:deploy
# Select deployment
r                # Restart to pick up new config
```

---

## 🟡 Workflow 7: Comparing Resource Usage

### Top View (Resource Usage)

```
:pods
# Look at CPU/MEM columns (if metrics-server is installed)
```

**Sort by CPU usage:**

```
Shift+CPU_COLUMN_NUMBER    # Sort to find hungry pods
```

### Node Resource Usage

```
:nodes           # or :no
```

```
NAME      STATUS  ROLES    CPU      MEM       CPU%   MEM%
node-1    Ready   worker   2000m    8Gi       50%    60%
node-2    Ready   worker   3500m    14Gi      87%    87%   ← Nearly full!
node-3    Ready   worker   500m     4Gi       12%    25%
```

**Node-2 is nearly full** — new pods might not schedule there.

---

## 🎯 Debugging Cheat Sheet

| Problem | K9s Workflow |
|---------|-------------|
| Pod won't start | `:pods` → select → `Enter` → check Events |
| Pod keeps crashing | `:pods` → select → `p` (previous logs) |
| App returns errors | `:pods` → select → `l` → `/error` |
| Can't connect to service | `:svc` → check endpoints → `:pods` → `s` → test connectivity |
| High resource usage | `:pods` → `Ctrl+W` → sort by CPU/MEM |
| Node issues | `:nodes` → select → `Enter` → check conditions |
| Config not updated | `:cm` → verify → `:deploy` → `r` (restart) |
| Recent changes | `:events` → `/warning` |
| Image pull errors | `:pods` → select → `Enter` → check image name |
| Need to test endpoint | `:pods` → select → `Shift+F` (port forward) |

---

## ✅ Hands-On Exercise

### Task: Debug a Broken Deployment

**1. Create a broken app:**

```bash
# In a separate terminal
kubectl create deployment broken --image=nginx:doesnotexist --replicas=2
kubectl create deployment healthy --image=nginx --replicas=2
```

**2. Open K9s:**

```bash
k9s
```

**3. Find the problem:**

```
:pods

# You should see:
# broken-xxx   0/1   ImagePullBackOff   ← Problem!
# broken-yyy   0/1   ImagePullBackOff   ← Problem!
# healthy-aaa  1/1   Running            ← OK
# healthy-bbb  1/1   Running            ← OK
```

**4. Diagnose:**

```
# Select a broken pod
Enter       # Describe
# Scroll to Events: "Failed to pull image nginx:doesnotexist"
```

**5. Fix:**

```
Esc
:deploy
# Select 'broken'
e           # Edit YAML
# Change image from nginx:doesnotexist to nginx:alpine
# Save and quit (:wq)
```

**6. Watch recovery:**

```
:pods
# Watch pods transition to Running
```

**7. Clean up:**

```
:deploy
Ctrl+K      # Delete 'broken'
Ctrl+K      # Delete 'healthy'
```

---

## 📚 Summary

**K9s Debugging Workflow:**

```
:pods → /filter → Enter (describe) → l (logs) → s (shell)
```

**Speed comparison:**

| Task | kubectl | K9s |
|------|---------|-----|
| Find crashing pods | 15 seconds | 3 seconds |
| View previous logs | 20 seconds | 2 keystrokes |
| Shell into container | 25 seconds | 1 keystroke |
| Port forward | 15 seconds | 3 keystrokes |
| Check events | 10 seconds | 2 keystrokes |
| Full diagnostics | 2-3 minutes | 30 seconds |

**When production is down, every second counts. K9s gives you those seconds back.**

---

**Previous:** [02. Resource Management](./02-resource-management.md)  
**Next:** [04. Advanced Features](./04-advanced-features.md)  
**Module:** [05. K9s](./README.md)
