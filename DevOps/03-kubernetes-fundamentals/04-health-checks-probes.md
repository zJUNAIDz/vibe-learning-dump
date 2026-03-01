# Health Checks and Probes

> **Teaching Kubernetes when your app is ready, alive, and healthy — the difference between crashing and recovering**

---

## 🟢 The Mental Model: Kubernetes Doesn't Know If Your App Works

**The Problem:**

```bash
# Your pod starts
kubectl get pods
# NAME        READY   STATUS    RESTARTS
# myapp-xyz   1/1     Running   0

# But is your app actually working?
# - Database connection failed?
# - Still loading config files?
# - Deadlocked?
# - Out of memory but not crashed yet?
```

**Kubernetes sees "process is running" ≠ "app is healthy"**

**Solution:** Health checks (probes).

---

## 🟢 Three Types of Probes

```
┌─────────────────────────────────────────────────┐
│  Startup Probe (🟢)                             │
│  "Is my app done starting?"                     │
│  - Runs ONCE at startup (until success)         │
│  - Gives slow-starting apps time                │
│  - If fails: restart pod                        │
└─────────────────────────────────────────────────┘
                    ↓ (once startup succeeds)
┌─────────────────────────────────────────────────┐
│  Liveness Probe (🔴)                            │
│  "Is my app alive?"                             │
│  - Runs throughout pod's life                   │
│  - Detects deadlocks, hangs, crashes            │
│  - If fails: restart container                  │
└─────────────────────────────────────────────────┘
                    +
┌─────────────────────────────────────────────────┐
│  Readiness Probe (🟡)                           │
│  "Is my app ready to receive traffic?"          │
│  - Runs throughout pod's life                   │
│  - Temporary unavailability (DB down, etc.)     │
│  - If fails: remove from Service endpoints      │
│  - Does NOT restart pod                         │
└─────────────────────────────────────────────────┘
```

---

## 🟢 Liveness Probe: "Is My App Alive?"

**Purpose:** Detect and recover from unrecoverable failures (deadlocks, infinite loops).

### Example: HTTP Liveness Probe

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: web-app
spec:
  containers:
  - name: app
    image: myapp:1.0
    ports:
    - containerPort: 8080
    livenessProbe:
      httpGet:
        path: /healthz
        port: 8080
      initialDelaySeconds: 10  # Wait 10s after container starts
      periodSeconds: 5         # Check every 5 seconds
      timeoutSeconds: 2        # Timeout after 2 seconds
      failureThreshold: 3      # Restart after 3 consecutive failures
```

**What happens:**
1. Container starts
2. Kubernetes waits 10 seconds (`initialDelaySeconds`)
3. Every 5 seconds, makes HTTP GET to `http://localhost:8080/healthz`
4. If response is 200-399: healthy
5. If fails 3 times in a row: restart container

### Example Health Endpoint (Go)

```go
package main

import (
    "fmt"
    "net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
    // Simple health check
    w.WriteHeader(http.StatusOK)
    fmt.Fprint(w, "OK")
}

func livenessHandler(w http.ResponseWriter, r *http.Request) {
    // Check if app is alive (not deadlocked)
    // For example, check if goroutines are responding
    
    // If app is healthy:
    w.WriteHeader(http.StatusOK)
    fmt.Fprint(w, "alive")
    
    // If app is deadlocked/unhealthy:
    // w.WriteHeader(http.StatusServiceUnavailable)
}

func main() {
    http.HandleFunc("/healthz", livenessHandler)
    http.HandleFunc("/ready", healthHandler)
    http.ListenAndServe(":8080", nil)
}
```

### Example Health Endpoint (TypeScript/Node)

```typescript
import express from 'express';

const app = express();
let isReady = false;

// Liveness: Is the process alive?
app.get('/healthz', (req, res) => {
  // If app is deadlocked, this won't respond
  res.status(200).send('alive');
});

// Readiness: Is the app ready to serve traffic?
app.get('/ready', (req, res) => {
  if (isReady) {
    res.status(200).send('ready');
  } else {
    res.status(503).send('not ready yet');
  }
});

// Simulate slow startup
setTimeout(() => {
  isReady = true;
  console.log('App is now ready!');
}, 10000); // 10 seconds

app.listen(8080, () => {
  console.log('Server started on :8080');
});
```

---

## 🟡 Readiness Probe: "Can I Receive Traffic?"

**Purpose:** Temporarily remove pod from load balancer if it's not ready.

**Key Difference from Liveness:**
- **Liveness failure** → restart container (problem is unrecoverable)
- **Readiness failure** → stop sending traffic (problem is temporary)

### Example: Readiness Probe

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: api-server
spec:
  containers:
  - name: app
    image: api:1.0
    ports:
    - containerPort: 8080
    readinessProbe:
      httpGet:
        path: /ready
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 3
      failureThreshold: 2
```

**What happens:**
1. Pod starts and joins Service
2. Readiness probe checks `/ready` every 3 seconds
3. If `/ready` returns non-200 for 2 consecutive checks:
   - Pod stays running
   - Pod is removed from Service endpoints
   - No traffic is sent to this pod
4. Once `/ready` returns 200 again:
   - Pod is added back to Service endpoints

### Example: Readiness with Database Check (Go)

```go
package main

import (
    "database/sql"
    "fmt"
    "net/http"
    _ "github.com/lib/pq"
)

var db *sql.DB

func readinessHandler(w http.ResponseWriter, r *http.Request) {
    // Check database connection
    err := db.Ping()
    if err != nil {
        // Database is down, don't send traffic to this pod
        w.WriteHeader(http.StatusServiceUnavailable)
        fmt.Fprintf(w, "Database unreachable: %v", err)
        return
    }
    
    // All dependencies are healthy
    w.WriteHeader(http.StatusOK)
    fmt.Fprint(w, "ready")
}

func main() {
    var err error
    db, err = sql.Open("postgres", "postgres://user:pass@db:5432/mydb")
    if err != nil {
        panic(err)
    }
    
    http.HandleFunc("/ready", readinessHandler)
    http.ListenAndServe(":8080", nil)
}
```

**Why this matters:**
- During database maintenance, pods mark themselves as "not ready"
- Traffic stops flowing to them
- They stay running (not restarted)
- Once DB is back, they're automatically added back

---

## 🟡 Startup Probe: "Have I Finished Starting?"

**Problem:** Some apps take a long time to start (legacy Java apps, ML model loading).

**Without startup probe:**
```yaml
livenessProbe:
  initialDelaySeconds: 120  # Have to guess startup time
  # If you guess too low, pod restarts before it's ready
  # If you guess too high, you wait unnecessarily
```

**With startup probe:**
```yaml
startupProbe:
  httpGet:
    path: /healthz
    port: 8080
  periodSeconds: 5
  failureThreshold: 30  # 30 * 5 = 150 seconds max startup time

livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  periodSeconds: 5
  failureThreshold: 3
```

**How it works:**
1. Startup probe runs every 5 seconds (up to 30 times = 150 seconds)
2. Once startup probe succeeds, it stops running
3. Liveness probe takes over
4. If startup probe fails 30 times, pod is restarted

**Use case:** Apps with unpredictable startup times.

---

## 🟢 Probe Types: HTTP, TCP, Exec

### 1. HTTP Probe (Most Common)

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
    httpHeaders:
    - name: X-Health-Check
      value: "true"
  periodSeconds: 10
```

**When to use:** Web apps, APIs.

### 2. TCP Probe

```yaml
livenessProbe:
  tcpSocket:
    port: 5432
  periodSeconds: 10
```

**When to use:** Databases, TCP services (no HTTP endpoint).

**What it does:** Tries to open a TCP connection. If successful, probe passes.

### 3. Exec Probe (Command Execution)

```yaml
livenessProbe:
  exec:
    command:
    - cat
    - /tmp/healthy
  periodSeconds: 10
```

**When to use:** Custom health checks, non-HTTP services.

**What it does:** Runs command inside container. Exit code 0 = success.

#### Example: PostgreSQL Health Check

```yaml
livenessProbe:
  exec:
    command:
    - /bin/sh
    - -c
    - pg_isready -U $POSTGRES_USER
  periodSeconds: 10
```

---

## 🔴 Common Mistakes

### ❌ Mistake 1: No Health Checks at All

**Problem:** Kubernetes thinks your app is healthy because the process is running.

```yaml
# BAD: No probes
spec:
  containers:
  - name: app
    image: myapp:1.0
```

**Reality:**
- App deadlocks → pod stays "Running" forever
- Database connection lost → traffic still sent to pod

**Fix:** Always add at least liveness and readiness probes.

### ❌ Mistake 2: Using Liveness Probe for Temporary Failures

```yaml
# BAD: Will restart pod if database is temporarily down
livenessProbe:
  httpGet:
    path: /health
  periodSeconds: 5
```

**Your `/health` endpoint:**
```go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    // BAD: Checking external dependency
    if err := db.Ping(); err != nil {
        w.WriteHeader(500)  // Liveness probe fails → restart pod
        return
    }
    w.WriteHeader(200)
}
```

**Problem:** Database glitch causes all pods to restart → cascading failure.

**Fix:** Use readiness probe for dependency checks.

```yaml
livenessProbe:
  httpGet:
    path: /alive
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /ready  # This checks DB
  periodSeconds: 5
```

### ❌ Mistake 3: Probe Timeouts Too Aggressive

```yaml
# BAD: Only 1 second timeout
livenessProbe:
  httpGet:
    path: /healthz
  periodSeconds: 5
  timeoutSeconds: 1
  failureThreshold: 1  # BAD: Restart after 1 failure
```

**Problem:** Under load, health check might take >1 second → pod restarts → more load on other pods → cascading failures.

**Fix:** Be generous with timeouts and failure thresholds.

```yaml
livenessProbe:
  httpGet:
    path: /healthz
  periodSeconds: 10
  timeoutSeconds: 3
  failureThreshold: 3  # Tolerate 3 failures = 30 seconds before restart
```

### ❌ Mistake 4: Health Endpoint Does Expensive Work

```go
// BAD: Expensive health check
func healthHandler(w http.ResponseWriter, r *http.Request) {
    // This runs every 5 seconds!
    rows, err := db.Query("SELECT COUNT(*) FROM users")  // Expensive!
    if err != nil {
        w.WriteHeader(500)
        return
    }
    // ...
}
```

**Problem:** Health checks run frequently. Expensive checks add load.

**Fix:** Keep health checks lightweight.

```go
// GOOD: Fast health check
func healthHandler(w http.ResponseWriter, r *http.Request) {
    // Just check if DB connection pool is alive
    if err := db.Ping(); err != nil {
        w.WriteHeader(500)
        return
    }
    w.WriteHeader(200)
}
```

---

## 🔴 War Story: The Cascading Restart

> *"We had liveness probes checking the database. One night, our cloud provider had a 10-second network blip. All 50 pods failed their liveness checks simultaneously and restarted. During restart, they couldn't connect to DB (still recovering), so they failed liveness again and restarted again. Restart loop lasted 45 minutes. We fixed it by changing DB checks to readiness probes only."*

**Lesson:**
- Liveness = "Is my code broken?" (internal)
- Readiness = "Are my dependencies broken?" (external)

---

## 🎯 Decision Tree: Which Probe to Use?

```
Is this an internal failure (deadlock, OOM, infinite loop)?
  ├─ YES → Liveness Probe
  └─ NO → Is this a temporary dependency failure (DB, cache, API)?
            ├─ YES → Readiness Probe
            └─ NO → Does your app take >30s to start?
                      ├─ YES → Startup Probe + Liveness + Readiness
                      └─ NO → Liveness + Readiness
```

---

## 🎯 Best Practices

### ✅ DO:
1. **Always implement both liveness and readiness probes**
2. **Use separate endpoints** (`/healthz` for liveness, `/ready` for readiness)
3. **Liveness checks internal health** (not dependencies)
4. **Readiness checks dependencies** (DB, cache, APIs)
5. **Keep probes lightweight** (milliseconds, not seconds)
6. **Use generous timeouts** (3-5 seconds)
7. **Tolerate multiple failures** (`failureThreshold: 3`)

### ❌ DON'T:
1. **Don't use liveness probes for dependency checks** (use readiness)
2. **Don't make health checks expensive** (no complex queries)
3. **Don't use `failureThreshold: 1`** (too aggressive)
4. **Don't skip probes** ("it works on my laptop" ≠ production)

---

## ✅ Hands-On Exercise

### Task: Deploy App with Proper Health Checks

**1. Create a simple web server (TypeScript):**

```typescript
// server.ts
import express from 'express';

const app = express();
let isHealthy = true;
let isReady = false;

// Simulate startup delay
setTimeout(() => {
  isReady = true;
  console.log('App is ready!');
}, 15000); // 15 seconds

// Liveness: Is the app alive?
app.get('/healthz', (req, res) => {
  if (isHealthy) {
    res.status(200).send('alive');
  } else {
    res.status(500).send('not alive');
  }
});

// Readiness: Is the app ready for traffic?
app.get('/ready', (req, res) => {
  if (isReady) {
    res.status(200).send('ready');
  } else {
    res.status(503).send('not ready');
  }
});

// Main endpoint
app.get('/', (req, res) => {
  res.send('Hello from healthy app!');
});

// Simulate failure (for testing)
app.post('/fail', (req, res) => {
  isHealthy = false;
  res.send('App will fail health checks now');
});

app.listen(8080, () => {
  console.log('Server running on :8080');
});
```

**2. Dockerfile:**

```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY server.ts ./
RUN npx tsc server.ts
CMD ["node", "server.js"]
```

**3. Kubernetes Deployment:**

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: app
        image: myapp:1.0
        ports:
        - containerPort: 8080
        startupProbe:
          httpGet:
            path: /ready
            port: 8080
          periodSeconds: 5
          failureThreshold: 6  # 30 seconds max
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
          timeoutSeconds: 3
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          periodSeconds: 5
          failureThreshold: 2
---
apiVersion: v1
kind: Service
metadata:
  name: web-service
spec:
  selector:
    app: web
  ports:
  - port: 80
    targetPort: 8080
```

**4. Deploy:**

```bash
kubectl apply -f deployment.yaml
```

**5. Watch pod startup:**

```bash
# Watch pods come up
kubectl get pods -l app=web -w

# Check events
kubectl describe pod -l app=web

# You'll see:
# - Startup probe runs every 5s
# - After 15s, startup succeeds
# - Liveness and readiness probes take over
```

**6. Test failure scenario:**

```bash
# Get a pod name
POD=$(kubectl get pod -l app=web -o jsonpath='{.items[0].metadata.name}')

# Trigger failure
kubectl exec $POD -- curl -X POST http://localhost:8080/fail

# Watch pod restart
kubectl get pods -l app=web -w

# After ~30 seconds (3 failures × 10s period), pod restarts
```

---

## 📊 Probe Configuration Reference

```yaml
livenessProbe:
  httpGet:
    path: /healthz           # Endpoint to check
    port: 8080               # Port
    httpHeaders:             # Optional headers
    - name: X-Custom-Header
      value: value
  
  initialDelaySeconds: 10    # Wait this long after container starts
  periodSeconds: 10          # Check every N seconds
  timeoutSeconds: 3          # Fail if no response in N seconds
  successThreshold: 1        # Consider healthy after N successes (always 1 for liveness)
  failureThreshold: 3        # Fail after N consecutive failures
```

---

## 📚 Summary

| Probe Type | Purpose | On Failure | Check |
|------------|---------|------------|-------|
| **Startup** | Finished starting? | Restart pod | Internal + dependencies |
| **Liveness** | Still alive? | Restart container | Internal only |
| **Readiness** | Ready for traffic? | Remove from Service | Dependencies |

**Decision Matrix:**

| Scenario | Liveness | Readiness |
|----------|----------|-----------|
| App deadlocked | ✅ Fail | ➖ |
| App out of memory | ✅ Fail | ➖ |
| Database temporarily down | ❌ Pass | ✅ Fail |
| Cache unreachable | ❌ Pass | ✅ Fail |
| App still loading config | ❌ Pass | ✅ Fail |

---

**Previous:** [03. ConfigMaps, Secrets, and Volumes](./03-configmaps-secrets-volumes.md)  
**Next:** [05. Ingress and Load Balancing](./05-ingress-load-balancing.md)  
**Module:** [03. Kubernetes Fundamentals](./README.md)
