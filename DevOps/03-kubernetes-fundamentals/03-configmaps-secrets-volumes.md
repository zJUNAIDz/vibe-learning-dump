# ConfigMaps, Secrets, and Volumes

> **Configuration and storage in Kubernetes — understanding the difference between ephemeral and persistent data**

---

## 🟢 The Mental Model: Separation of Concerns

**Key Principle:** Separate **code** from **configuration** from **state**.

```
┌──────────────────────────────────────────────┐
│  Container Image (Code)                      │
│  - Your application binary                   │
│  - Language runtime                          │
│  - Dependencies                              │
│  - NEVER contains secrets or config          │
└──────────────────────────────────────────────┘
                    ↕
┌──────────────────────────────────────────────┐
│  ConfigMaps & Secrets (Configuration)        │
│  - Database URLs                             │
│  - Feature flags                             │
│  - API keys (Secrets only!)                  │
└──────────────────────────────────────────────┘
                    ↕
┌──────────────────────────────────────────────┐
│  Volumes (State/Data)                        │
│  - Database files                            │
│  - Uploaded images                           │
│  - Logs (sometimes)                          │
└──────────────────────────────────────────────┘
```

**Why this matters:**
- Same image works in dev, staging, production
- Change config without rebuilding
- Rollback code without losing data

---

## 🟢 ConfigMaps: Non-Sensitive Configuration

### What Is a ConfigMap?

A ConfigMap is a Kubernetes object that stores **key-value pairs**.

**Think of it as:** Environment variables + config files in the cluster.

### Creating a ConfigMap

#### Method 1: From Literal Values
```bash
kubectl create configmap app-config \
  --from-literal=DATABASE_HOST=postgres.default.svc.cluster.local \
  --from-literal=LOG_LEVEL=info \
  --from-literal=MAX_CONNECTIONS=100
```

#### Method 2: From a File
```bash
# Create config file
cat > app.properties <<EOF
database.host=postgres.default.svc.cluster.local
log.level=info
max.connections=100
EOF

# Create ConfigMap from file
kubectl create configmap app-config --from-file=app.properties
```

#### Method 3: YAML Manifest (Declarative)
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
data:
  DATABASE_HOST: "postgres.default.svc.cluster.local"
  LOG_LEVEL: "info"
  MAX_CONNECTIONS: "100"
  app.properties: |
    database.host=postgres.default.svc.cluster.local
    log.level=info
    max.connections=100
```

### Using ConfigMaps in Pods

#### Option 1: Environment Variables
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp
spec:
  containers:
  - name: app
    image: myapp:1.0
    env:
    - name: DATABASE_HOST
      valueFrom:
        configMapKeyRef:
          name: app-config
          key: DATABASE_HOST
    - name: LOG_LEVEL
      valueFrom:
        configMapKeyRef:
          name: app-config
          key: LOG_LEVEL
```

#### Option 2: Mount as Volume (File)
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp
spec:
  containers:
  - name: app
    image: myapp:1.0
    volumeMounts:
    - name: config-volume
      mountPath: /etc/config
      readOnly: true
  volumes:
  - name: config-volume
    configMap:
      name: app-config
```

Now your app can read `/etc/config/app.properties`.

#### Option 3: All Keys as Env Vars (Bulk)
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp
spec:
  containers:
  - name: app
    image: myapp:1.0
    envFrom:
    - configMapRef:
        name: app-config
```

### 🔴 ConfigMap Reality Check

**ConfigMaps are NOT automatically reloaded!**

```yaml
# You change the ConfigMap
kubectl edit configmap app-config

# Your pod doesn't see the change until you restart it!
kubectl rollout restart deployment myapp
```

**Why?** Environment variables are set at pod startup.

**Solution:** Use file-based ConfigMaps and watch for changes in your app code.

---

## 🟡 Secrets: Sensitive Configuration

### What Is a Secret?

A Secret is like a ConfigMap but for **sensitive data**.

**CRITICAL:** Secrets in Kubernetes are **base64-encoded, NOT encrypted**.

```bash
echo "mypassword" | base64
# bXlwYXNzd29yZAo=

echo "bXlwYXNzd29yZAo=" | base64 -d
# mypassword
```

**Base64 is NOT encryption.** Anyone with cluster access can decode secrets.

### Creating Secrets

#### Method 1: From Literal Values
```bash
kubectl create secret generic db-credentials \
  --from-literal=username=admin \
  --from-literal=password=supersecret123
```

#### Method 2: From Files
```bash
echo -n 'admin' > ./username.txt
echo -n 'supersecret123' > ./password.txt

kubectl create secret generic db-credentials \
  --from-file=username=./username.txt \
  --from-file=password=./password.txt

# Clean up (don't leave secrets on disk!)
shred -u username.txt password.txt
```

#### Method 3: YAML Manifest
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-credentials
type: Opaque
data:
  username: YWRtaW4=          # base64 of "admin"
  password: c3VwZXJzZWNyZXQxMjM=  # base64 of "supersecret123"
```

**WARNING:** If you check this YAML into Git, you've leaked your secrets!

### Using Secrets in Pods

#### Option 1: Environment Variables
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp
spec:
  containers:
  - name: app
    image: myapp:1.0
    env:
    - name: DB_USERNAME
      valueFrom:
        secretKeyRef:
          name: db-credentials
          key: username
    - name: DB_PASSWORD
      valueFrom:
        secretKeyRef:
          name: db-credentials
          key: password
```

**Your app code (Go):**
```go
package main

import (
    "os"
    "fmt"
)

func main() {
    dbUser := os.Getenv("DB_USERNAME")
    dbPass := os.Getenv("DB_PASSWORD")
    
    // DO NOT LOG SECRETS!
    fmt.Println("Connecting to database as", dbUser)
}
```

#### Option 2: Mount as Volume (More Secure)
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp
spec:
  containers:
  - name: app
    image: myapp:1.0
    volumeMounts:
    - name: secrets-volume
      mountPath: /etc/secrets
      readOnly: true
  volumes:
  - name: secrets-volume
    secret:
      secretName: db-credentials
```

Now your app reads:
- `/etc/secrets/username` → `admin`
- `/etc/secrets/password` → `supersecret123`

**Why volumes are better:**
- Not visible in `kubectl describe pod`
- Not visible in container inspect
- Can be rotated without restart (with proper app code)

---

## 🔴 Secret Management: The Real World

### Problem: Kubernetes Secrets Are Not Encrypted at Rest (by default)

**What "at rest" means:** The etcd database where Kubernetes stores secrets.

**Check if encryption is enabled:**
```bash
# This requires cluster admin access
kubectl get secret -n kube-system -o yaml
# If you see "data:", it's base64 (NOT encrypted)
# If you see "encryptedData:", encryption is enabled
```

**Solution Options:**

1. **Enable etcd encryption** (cluster admin task)
2. **Use external secret managers:**
   - AWS Secrets Manager
   - HashiCorp Vault
   - Azure Key Vault
   - Google Secret Manager

3. **Use Sealed Secrets (Bitnami)**
   - Encrypt secrets before checking into Git
   - Kubernetes operator decrypts them

4. **Use External Secrets Operator**
   - Syncs secrets from external sources

### War Story: The Leaked Database Password

> *"We had all our secrets in Git, base64-encoded. A developer pushed code to GitHub. GitHub's secret scanner flagged it. We had to rotate 47 database passwords across 12 services in production. Took 6 hours. The incident report was 14 pages."*

**Lessons:**
- Never commit secrets to Git (even base64-encoded)
- Use `.gitignore` for any secret-containing files
- Use secret managers for production
- Rotate secrets regularly

---

## 🟢 Volumes: Ephemeral vs Persistent Storage

### The Problem: Containers Are Ephemeral

```bash
# Pod writes to /tmp/data.txt
kubectl exec myapp -- touch /tmp/data.txt

# Pod crashes and restarts
# /tmp/data.txt is GONE
```

**Why?** Each container restart gets a fresh filesystem from the image.

### Volume Types (Simplified)

```
Ephemeral (dies with pod)         Persistent (outlives pod)
━━━━━━━━━━━━━━━━━━━━              ━━━━━━━━━━━━━━━━━━━━━━
emptyDir                          PersistentVolumeClaim
hostPath (dangerous!)             Cloud volumes (EBS, Azure Disk)
ConfigMaps/Secrets                NFS
                                  Ceph, GlusterFS
```

---

## 🟢 emptyDir: Shared Scratch Space

**Use case:** Sharing data between containers in the same pod.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: multi-container
spec:
  containers:
  - name: writer
    image: busybox
    command: ['sh', '-c', 'while true; do date >> /data/log.txt; sleep 5; done']
    volumeMounts:
    - name: shared-data
      mountPath: /data
  
  - name: reader
    image: busybox
    command: ['sh', '-c', 'tail -f /data/log.txt']
    volumeMounts:
    - name: shared-data
      mountPath: /data
  
  volumes:
  - name: shared-data
    emptyDir: {}
```

**Key points:**
- Created when pod starts
- Deleted when pod is deleted
- Shared between all containers in the pod
- Stored on the node's disk

---

## 🟡 hostPath: Mounting Node Filesystem (Dangerous!)

**Use case:** Very rare — usually only for system workloads.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: host-access
spec:
  containers:
  - name: app
    image: myapp:1.0
    volumeMounts:
    - name: host-logs
      mountPath: /mnt/logs
  volumes:
  - name: host-logs
    hostPath:
      path: /var/log
      type: Directory
```

**Why it's dangerous:**
- Pod can read/write host filesystem
- Security risk
- Pod only works on specific node
- Not portable across clusters

**When to use:** DaemonSets that need node-level access (log collectors, monitoring agents).

---

## 🟡 PersistentVolumes and PersistentVolumeClaims

### Mental Model: Storage as a Resource

```
PersistentVolume (admin creates)
    ↕
PersistentVolumeClaim (developer requests)
    ↕
Pod (uses claim)
```

**Think of it like:**
- PersistentVolume = Actual hard drive
- PersistentVolumeClaim = "I need 10GB of storage"
- StorageClass = "Give me fast SSD" or "cheap HDD"

### Example: PostgreSQL with Persistent Storage

#### 1. Create PersistentVolumeClaim
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pvc
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: standard  # or gp2 (AWS), pd-ssd (GCP)
```

#### 2. Use in Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15
        env:
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: password
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
```

#### 3. Verify
```bash
# Check PVC status
kubectl get pvc postgres-pvc
# NAME           STATUS   VOLUME                                     CAPACITY   ACCESS
# postgres-pvc   Bound    pvc-abc123-456def-789ghi                  10Gi       RWO

# Check PV (auto-created by cloud provider)
kubectl get pv
```

---

## 🟡 Access Modes Explained

```yaml
accessModes:
- ReadWriteOnce   # RWO - One node can mount read-write
- ReadOnlyMany    # ROX - Many nodes can mount read-only
- ReadWriteMany   # RWX - Many nodes can mount read-write
```

**Reality check:**
- **RWO** is most common (EBS, Azure Disk)
- **RWX** is expensive/slow (NFS, EFS, Azure Files)
- Most cloud block storage is RWO only

**Gotcha:** If you scale to 5 replicas with RWO, only 1 pod will work!

---

## 🔴 War Story: The OOMKilled Database

> *"We deployed PostgreSQL to Kubernetes. After a week, the pod kept restarting. Logs showed corruption. We forgot to mount persistent storage. PostgreSQL was writing to the ephemeral container filesystem. Every restart lost data. We had backups, but lost 4 hours of transactions."*

**Lesson:** Stateful workloads (databases) MUST use PersistentVolumes.

---

## 🎯 Commands Cheat Sheet

### ConfigMaps
```bash
# Create
kubectl create configmap NAME --from-literal=KEY=VALUE
kubectl create configmap NAME --from-file=FILE

# View
kubectl get configmap
kubectl describe configmap NAME
kubectl get configmap NAME -o yaml

# Edit
kubectl edit configmap NAME

# Delete
kubectl delete configmap NAME
```

### Secrets
```bash
# Create
kubectl create secret generic NAME --from-literal=KEY=VALUE
kubectl create secret generic NAME --from-file=KEY=PATH

# View (decoded)
kubectl get secret NAME -o jsonpath='{.data.KEY}' | base64 -d

# Delete
kubectl delete secret NAME
```

### Volumes
```bash
# List PVCs
kubectl get pvc

# Describe PVC
kubectl describe pvc NAME

# Delete PVC (WARNING: Deletes data!)
kubectl delete pvc NAME

# List PVs
kubectl get pv
```

---

## ✅ Hands-On Exercise

### Task: Deploy a Web App with Config and Secrets

1. **Create ConfigMap:**
```bash
kubectl create configmap web-config \
  --from-literal=APP_ENV=production \
  --from-literal=LOG_LEVEL=info
```

2. **Create Secret:**
```bash
kubectl create secret generic api-key \
  --from-literal=key=super-secret-api-key-12345
```

3. **Create Deployment:**
```yaml
# Save as app-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
spec:
  replicas: 2
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
        image: hashicorp/http-echo:latest
        args:
        - "-text=Hello from $(APP_ENV) with log level $(LOG_LEVEL)"
        env:
        - name: APP_ENV
          valueFrom:
            configMapKeyRef:
              name: web-config
              key: APP_ENV
        - name: LOG_LEVEL
          valueFrom:
            configMapKeyRef:
              name: web-config
              key: LOG_LEVEL
        - name: API_KEY
          valueFrom:
            secretKeyRef:
              name: api-key
              key: key
```

4. **Deploy:**
```bash
kubectl apply -f app-deployment.yaml
```

5. **Test:**
```bash
kubectl get pods -l app=web
kubectl logs deployment/web-app
kubectl exec deployment/web-app -- printenv | grep -E 'APP_ENV|LOG_LEVEL|API_KEY'
```

6. **Change Config (and see it doesn't update automatically):**
```bash
kubectl create configmap web-config \
  --from-literal=APP_ENV=production \
  --from-literal=LOG_LEVEL=debug \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart deployment to pick up changes
kubectl rollout restart deployment/web-app
```

---

## 🧩 Common Mistakes

### ❌ Mistake 1: Committing Secrets to Git
```yaml
# BAD: Never do this!
apiVersion: v1
kind: Secret
data:
  password: cGFzc3dvcmQxMjM=  # Anyone can decode this
```

**Fix:** Use external secret managers or Sealed Secrets.

### ❌ Mistake 2: Using Secrets as Env Vars in Prod
**Problem:** Visible in `kubectl describe pod` and container inspections.

**Fix:** Mount secrets as volumes.

### ❌ Mistake 3: Not Using Persistent Storage for Databases
**Problem:** Data loss on pod restart.

**Fix:** Always use PersistentVolumeClaims for stateful workloads.

### ❌ Mistake 4: Expecting ConfigMap Changes to Auto-Reload
**Problem:** Pods don't see ConfigMap changes until restart.

**Fix:** Use `kubectl rollout restart` or implement file watching.

---

## 📚 Summary

| Concept | Use Case | Survives Pod Restart? |
|---------|----------|----------------------|
| ConfigMap | Non-sensitive config | Yes (separate object) |
| Secret | Sensitive config | Yes (separate object) |
| emptyDir | Temp storage, shared between containers | No |
| hostPath | Node filesystem access | Yes (but dangerous) |
| PersistentVolumeClaim | Databases, uploads, logs | Yes |

**Key Takeaways:**
1. Separate code, config, and state
2. Secrets are base64, NOT encrypted (use external managers in prod)
3. ConfigMaps/Secrets don't auto-reload
4. Stateful workloads need PersistentVolumes
5. Never commit secrets to Git

---

**Previous:** [02. Core Concepts: Pods, Deployments, Services](./02-core-concepts-pods-deployments-services.md)  
**Next:** [04. Health Checks and Probes](./04-health-checks-probes.md)  
**Module:** [03. Kubernetes Fundamentals](./README.md)
