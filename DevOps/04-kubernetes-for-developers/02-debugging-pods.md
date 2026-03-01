# Debugging Pods and Containers

> **When your pod won't start, keeps crashing, or behaves mysteriously — a systematic debugging guide**

---

## 🟢 The Debugging Mental Model

```
Pod Status → Events → Logs → Exec into container → Check resources
     ↓
Is it starting?  Is it crashing?  Is it running but broken?
```

**Debugging is about narrowing down the problem systematically.**

---

## 🟢 Pod Lifecycle States

```bash
kubectl get pods
```

**Common statuses:**

| Status | Meaning | Common Causes |
|--------|---------|---------------|
| `Pending` | Pod accepted but not running yet | No suitable nodes, resource constraints |
| `ContainerCreating` | Pulling image, mounting volumes | Image pull issues, storage problems |
| `Running` | At least one container is running | (Good state, but check if app works) |
| `CrashLoopBackOff` | Container keeps crashing | App crashes on startup, bad command |
| `ImagePullBackOff` | Can't pull container image | Wrong image name, auth issues, network |
| `Error` | Pod terminated with error | Container exited with non-zero code |
| `Completed` | Pod finished successfully | Jobs, one-time tasks (normal)|
| `Unknown` | Can't communicate with node | Node down, network issues |
| `Terminating` | Pod is being deleted | Stuck if pod won't terminate gracefully |

---

## 🔴 Problem 1: ImagePullBackOff

### Symptom

```bash
kubectl get pods
# NAME        READY   STATUS             RESTARTS
# myapp-xyz   0/1     ImagePullBackOff   0
```

### Diagnosis

```bash
kubectl describe pod myapp-xyz
```

**Look for:**
```
Events:
  Warning  Failed   5s   kubelet   Failed to pull image "myapp:v1.0.0": rpc error: code = Unknown desc = Error response from daemon: pull access denied for myapp, repository does not exist or may require 'docker login'
```

### Common Causes

#### Cause 1: Typo in Image Name

```yaml
containers:
- name: app
  image: myap:1.0  # Typo! Should be myapp:1.0
```

**Fix:** Correct the image name.

#### Cause 2: Image Doesn't Exist

```yaml
image: myrepo/myapp:v1.0.0  # This tag doesn't exist
```

**Check if image exists:**
```bash
# Docker Hub
curl -s https://hub.docker.com/v2/repositories/myrepo/myapp/tags | jq '.results[].name'

# Or try pulling manually
docker pull myrepo/myapp:v1.0.0
```

#### Cause 3: Private Registry Needs Authentication

```yaml
image: mycompany.azurecr.io/myapp:1.0
# Kubernetes can't pull from private registry without credentials
```

**Fix: Create image pull secret**

```bash
kubectl create secret docker-registry acr-secret \
  --docker-server=mycompany.azurecr.io \
  --docker-username=myusername \
  --docker-password=mypassword \
  --docker-email=me@example.com
```

**Use secret in pod spec:**
```yaml
spec:
  imagePullSecrets:
  - name: acr-secret
  containers:
  - name: app
    image: mycompany.azurecr.io/myapp:1.0
```

#### Cause 4: Using `:latest` with imagePullPolicy: IfNotPresent

```yaml
image: myapp:latest
imagePullPolicy: IfNotPresent  # Doesn't pull if cached
```

**Fix:**
```yaml
image: myapp:latest
imagePullPolicy: Always  # Always pull latest
```

---

## 🔴 Problem 2: CrashLoopBackOff

### Symptom

```bash
kubectl get pods
# NAME        READY   STATUS             RESTARTS
# myapp-xyz   0/1     CrashLoopBackOff   5
```

**"CrashLoopBackOff" = Container starts → crashes → Kubernetes waits → tries again → crashes again → repeat**

### Diagnosis

```bash
# Check logs
kubectl logs myapp-xyz

# Check previous crashed container logs
kubectl logs myapp-xyz --previous

# Check events
kubectl describe pod myapp-xyz
```

### Common Causes

#### Cause 1: App Immediately Exits

**Example: Missing command**

```yaml
# Dockerfile
FROM busybox
# No CMD!
```

**Pod YAML:**
```yaml
spec:
  containers:
  - name: app
    image: myapp:1.0
    # No command specified
```

**Container starts → has nothing to do → exits → CrashLoopBackOff**

**Fix:**
```yaml
spec:
  containers:
  - name: app
    image: myapp:1.0
    command: ["sh", "-c", "while true; do echo hello; sleep 10; done"]
```

#### Cause 2: App Crashes on Startup

**Check logs:**
```bash
kubectl logs myapp-xyz --previous
```

**Example output:**
```
Error: Cannot connect to database at postgres:5432
Connection refused
```

**Problem:** Database isn't ready when app starts.

**Fix: Use init containers or retry logic**

```yaml
spec:
  initContainers:
  - name: wait-for-db
    image: busybox
    command: ['sh', '-c', 'until nc -z postgres 5432; do echo waiting for db; sleep 2; done']
  containers:
  - name: app
    image: myapp:1.0
```

#### Cause 3: Wrong Command/Args

```yaml
spec:
  containers:
  - name: app
    image: node:18
    command: ["npm"]
    args: ["start"]  # But package.json doesn't have "start" script
```

**Logs:**
```
Missing script: "start"
```

**Fix:** Verify command works locally first.

#### Cause 4: OOMKilled (Out of Memory)

```bash
kubectl describe pod myapp-xyz
```

**Look for:**
```
Last State:     Terminated
  Reason:       OOMKilled
  Exit Code:    137
```

**Problem:** App used more memory than its limit.

**Fix:** Increase memory limit (or fix memory leak).

```yaml
resources:
  limits:
    memory: "2Gi"  # Increased from 512Mi
```

---

## 🟡 Problem 3: Pending Pods

### Symptom

```bash
kubectl get pods
# NAME        READY   STATUS    RESTARTS
# myapp-xyz   0/1     Pending   0
```

**Pod is accepted but not scheduled to any node.**

### Diagnosis

```bash
kubectl describe pod myapp-xyz
```

**Look for:**
```
Events:
  Warning  FailedScheduling  3s  default-scheduler  0/3 nodes are available: 3 Insufficient cpu.
```

### Common Causes

#### Cause 1: Not Enough Resources

```yaml
resources:
  requests:
    cpu: "10"      # Requesting 10 CPUs
    memory: "32Gi"  # Requesting 32GB memory
# But cluster nodes only have 4 CPUs and 16GB each
```

**Fix:** Reduce resource requests or add more nodes.

#### Cause 2: Node Selector Doesn't Match

```yaml
spec:
  nodeSelector:
    disktype: ssd  # No nodes have this label!
```

**Check node labels:**
```bash
kubectl get nodes --show-labels
```

**Fix:** Use correct labels or remove node selector.

#### Cause 3: Taints and Tolerations

**Node has taint:**
```bash
kubectl describe node node1 | grep Taints
# Taints: dedicated=gpu:NoSchedule
```

**Pod doesn't tolerate it:**
```yaml
# Pod has no tolerations, so it can't be scheduled on tainted nodes
```

**Fix: Add toleration**
```yaml
spec:
  tolerations:
  - key: "dedicated"
    operator: "Equal"
    value: "gpu"
    effect: "NoSchedule"
```

#### Cause 4: PersistentVolumeClaim Not Bound

```yaml
volumes:
- name: data
  persistentVolumeClaim:
    claimName: my-pvc
```

**Check PVC:**
```bash
kubectl get pvc my-pvc
# NAME     STATUS   VOLUME   CAPACITY
# my-pvc   Pending  ...      ...
```

**Problem:** No PersistentVolume available.

**Fix:** Create PV or use dynamic provisioning.

---

## 🟡 Problem 4: Running but Not Working

### Symptom

```bash
kubectl get pods
# NAME        READY   STATUS    RESTARTS
# myapp-xyz   1/1     Running   0

# But app doesn't respond
curl http://myapp-service
# Connection refused
```

### Diagnosis Steps

#### Step 1: Check Logs

```bash
kubectl logs myapp-xyz

# Follow logs (live tail)
kubectl logs myapp-xyz -f

# Check last 100 lines
kubectl logs myapp-xyz --tail=100

# Check all containers in pod
kubectl logs myapp-xyz --all-containers
```

#### Step 2: Exec into Container

```bash
# Get a shell inside container
kubectl exec -it myapp-xyz -- /bin/sh

# Try to access the app from inside
wget -O- http://localhost:8080
# or
curl http://localhost:8080
```

**If it works from inside:** Problem is with Service or Ingress.  
**If it doesn't work from inside:** Problem is with the app.

#### Step 3: Check Service Endpoints

```bash
kubectl get svc myapp-service
kubectl describe svc myapp-service

# Check if service has endpoints
kubectl get endpoints myapp-service
```

**Expected output:**
```
NAME            ENDPOINTS
myapp-service   10.244.1.5:8080,10.244.1.6:8080
```

**If no endpoints:** Pod labels don't match service selector.

```yaml
# Service
spec:
  selector:
    app: myapp  # Looking for pods with label app=myapp

# Pod
metadata:
  labels:
    app: webapp  # MISMATCH! Should be app=myapp
```

#### Step 4: Test Service DNS

```bash
# From another pod
kubectl run test --image=curlimages/curl -it --rm -- sh

# Inside test pod
curl http://myapp-service:80
nslookup myapp-service
```

#### Step 5: Check Port Mapping

```yaml
# Common mistake: Port mismatch
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: app
    image: myapp:1.0
    ports:
    - containerPort: 8080  # App listens on 8080
---
apiVersion: v1
kind: Service
spec:
  selector:
    app: myapp
  ports:
  - port: 80
    targetPort: 3000  # WRONG! Should be 8080
```

---

## 🟢 Essential Debugging Commands

### Viewing Logs

```bash
# Current container logs
kubectl logs POD_NAME

# Previous crashed container
kubectl logs POD_NAME --previous

# Specific container in multi-container pod
kubectl logs POD_NAME -c CONTAINER_NAME

# Follow logs (live)
kubectl logs POD_NAME -f

# Last N lines
kubectl logs POD_NAME --tail=50

# Since timestamp
kubectl logs POD_NAME --since=1h

# All containers in pod
kubectl logs POD_NAME --all-containers=true
```

### Executing Commands

```bash
# Interactive shell
kubectl exec -it POD_NAME -- /bin/sh
kubectl exec -it POD_NAME -- /bin/bash

# Run single command
kubectl exec POD_NAME -- ls -la /app
kubectl exec POD_NAME -- cat /etc/config/app.conf

# Specific container
kubectl exec -it POD_NAME -c CONTAINER_NAME -- sh
```

### Copying Files

```bash
# Copy from pod to local
kubectl cp POD_NAME:/app/logs/error.log ./error.log

# Copy from local to pod
kubectl cp ./config.json POD_NAME:/app/config.json

# Specify container
kubectl cp POD_NAME:/logs/error.log ./error.log -c CONTAINER_NAME
```

### Port Forwarding (Debug without Service)

```bash
# Forward pod port to localhost
kubectl port-forward POD_NAME 8080:8080

# Now access: http://localhost:8080

# Forward service
kubectl port-forward svc/myapp-service 8080:80

# Forward deployment
kubectl port-forward deployment/myapp 8080:8080
```

### Getting Events

```bash
# All events in namespace
kubectl get events --sort-by='.lastTimestamp'

# Events for specific pod
kubectl describe pod POD_NAME | grep -A 10 Events

# Watch events
kubectl get events -w
```

### Describing Resources

```bash
# Detailed pod info
kubectl describe pod POD_NAME

# Detailed deployment info
kubectl describe deployment DEPLOYMENT_NAME

# Detailed service info
kubectl describe service SERVICE_NAME
```

---

## 🔴 Advanced Debugging: Ephemeral Containers

**Problem:** Production pod doesn't have debugging tools (no `curl`, `netstat`, `strace`).

**Solution:** Ephemeral debug containers (Kubernetes 1.23+).

```bash
# Add debug container to running pod
kubectl debug POD_NAME -it --image=nicolaka/netshoot

# Now you have a fully-loaded debug container in the same namespace/network as the target pod
# Comes with: curl, dig, netstat, tcpdump, iperf, nmap, etc.
```

**Example:**
```bash
kubectl debug myapp-xyz -it --image=nicolaka/netshoot

# Inside debug container
curl localhost:8080
netstat -tulpn
tcpdump -i any port 8080
```

---

## 🔴 Debugging Init Containers

```yaml
spec:
  initContainers:
  - name: init-db
    image: busybox
    command: ['sh', '-c', 'until nc -z postgres 5432; do echo waiting; sleep 2; done']
  containers:
  - name: app
    image: myapp:1.0
```

**If init container fails, main container never starts.**

```bash
# Check init container logs
kubectl logs POD_NAME -c init-db

# Check init container status
kubectl describe pod POD_NAME
```

**Look for:**
```
Init Containers:
  init-db:
    State:          Waiting
      Reason:       CrashLoopBackOff
```

---

## 🎯 Debugging Checklist

When a pod isn't working:

1. ✅ Check pod status: `kubectl get pods`
2. ✅ Check events: `kubectl describe pod POD_NAME`
3. ✅ Check logs: `kubectl logs POD_NAME`
4. ✅ Try previous logs: `kubectl logs POD_NAME --previous`
5. ✅ Exec into container: `kubectl exec -it POD_NAME -- sh`
6. ✅ Test connectivity: `curl localhost:8080` (from inside container)
7. ✅ Check service endpoints: `kubectl get endpoints SERVICE_NAME`
8. ✅ Test DNS: `nslookup SERVICE_NAME` (from another pod)
9. ✅ Port forward: `kubectl port-forward POD_NAME 8080:8080`
10. ✅ Check resource usage: `kubectl top pod POD_NAME`

---

## ✅ Hands-On Exercise

### Task: Debug a Broken Deployment

**1. Deploy broken app:**

```yaml
# broken-app.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: broken-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: broken
  template:
    metadata:
      labels:
        app: broken
    spec:
      containers:
      - name: app
        image: nginx:wrongtag  # Doesn't exist!
        ports:
        - containerPort: 80
```

```bash
kubectl apply -f broken-app.yaml
```

**2. Diagnose:**

```bash
kubectl get pods
# STATUS: ImagePullBackOff

kubectl describe pod broken-app-xyz | grep -A 5 Events
# Failed to pull image "nginx:wrongtag"
```

**3. Fix:**

```bash
kubectl set image deployment/broken-app app=nginx:latest
```

**4. Create another broken app:**

```yaml
# crashing-app.yaml
apiVersion: v1
kind: Pod
metadata:
  name: crashing
spec:
  containers:
  - name: app
    image: busybox
    command: ["sh", "-c", "echo Starting...; exit 1"]
```

```bash
kubectl apply -f crashing-app.yaml

# Wait 10 seconds
kubectl get pods
# STATUS: CrashLoopBackOff

kubectl logs crashing
# Starting...

kubectl logs crashing --previous
# Starting...

# Fix: Make it not exit
kubectl delete pod crashing
```

---

## 📚 Summary

**Common Pod Problems:**

| Status | Meaning | First Step |
|--------|---------|------------|
| `ImagePullBackOff` | Can't pull image | `kubectl describe pod` → check image name/auth |
| `CrashLoopBackOff` | App keeps crashing | `kubectl logs --previous` → check app logs |
| `Pending` | Can't be scheduled | `kubectl describe pod` → check resources/nodes |
| `Running` but broken | App not responding | `kubectl exec` → test from inside container |

**Debugging Flow:**
```
kubectl get pods → kubectl describe pod → kubectl logs → kubectl exec
```

**Key Commands:**
```bash
kubectl logs POD_NAME              # Check logs
kubectl logs POD_NAME --previous   # Check crashed container logs
kubectl describe pod POD_NAME      # Check events
kubectl exec -it POD_NAME -- sh    # Get shell inside
kubectl port-forward POD_NAME 8080:8080  # Test without Service
kubectl debug POD_NAME -it --image=nicolaka/netshoot  # Advanced debugging
```

---

**Previous:** [01. Local Development Workflows](./01-local-development-workflows.md)  
**Next:** [03. Resource Requests and Limits](./03-resource-requests-limits.md)  
**Module:** [04. Kubernetes for Developers](./README.md)
