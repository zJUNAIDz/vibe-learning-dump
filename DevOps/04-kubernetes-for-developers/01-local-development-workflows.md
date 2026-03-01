# Local Development Workflows with Kubernetes

> **Making Kubernetes development fast and painless — because waiting 5 minutes per code change is unacceptable**

---

## 🟢 The Problem: Kubernetes is Slow for Development

**Traditional workflow (painful):**

```
1. Write code
2. Build Docker image (2-5 minutes)
3. Push to registry (1-3 minutes)
4. Update Kubernetes deployment
5. Wait for pod to start (30 seconds)
6. Find a bug
7. Go to step 1

Total: 5-10 minutes per iteration 😭
```

**What we want:**

```
1. Write code
2. See changes in < 5 seconds ⚡
```

---

## 🟢 Local Kubernetes Options

### Option 1: minikube (Most Popular)

**What it is:** Single-node Kubernetes cluster in a VM.

```bash
# Install (Fedora)
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# Start cluster
minikube start

# Use local Docker images (no push to registry!)
eval $(minikube docker-env)
docker build -t myapp:dev .
kubectl run myapp --image=myapp:dev --image-pull-policy=Never
```

**Pros:**
- Easy to install
- Works on Linux, Mac, Windows
- Supports LoadBalancer services
- Addons (ingress, metrics, dashboard)

**Cons:**
- Runs in VM (extra resource overhead)
- Slower on non-Linux systems

### Option 2: kind (Kubernetes in Docker)

**What it is:** Kubernetes cluster runs in Docker containers.

```bash
# Install
go install sigs.k8s.io/kind@latest

# Create cluster
kind create cluster

# Load local image into cluster (no registry!)
docker build -t myapp:dev .
kind load docker-image myapp:dev

kubectl run myapp --image=myapp:dev --image-pull-policy=Never
```

**Pros:**
- Faster than minikube (no VM)
- Multi-node clusters (good for testing)
- Perfect for CI/CD pipelines

**Cons:**
- Slightly harder to set up
- No LoadBalancer service (needs workarounds)

### Option 3: k3d (Lightweight k3s in Docker)

**What it is:** Rancher's k3s (lightweight K8s) in Docker.

```bash
# Install
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# Create cluster
k3d cluster create dev

# Load image
docker build -t myapp:dev .
k3d image import myapp:dev

kubectl run myapp --image=myapp:dev --image-pull-policy=Never
```

**Pros:**
- Fastest startup
- Smallest resource footprint
- Built-in LoadBalancer

**Cons:**
- k3s is slightly different from "real" Kubernetes

### Option 4: Docker Desktop Kubernetes

**What it is:** Built-in Kubernetes in Docker Desktop.

**Pros:**
- One-click enable
- Shares Docker daemon (no image pushing)

**Cons:**
- Mac/Windows only
- Resource-heavy
- Single-node only

---

## 🟡 Fast Iteration: Hot Reload Solutions

### Problem with Kubernetes Development

```
Change code → Rebuild image → Redeploy → Wait 2-5 minutes
```

**We need hot reload like we have for local development.**

---

## 🟡 Solution 1: Skaffold (Recommended)

**What it is:** Automates the build-push-deploy cycle.

### Install

```bash
# Fedora/Linux
curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64
sudo install skaffold /usr/local/bin/

skaffold version
```

### Simple Example: Node.js App

**Project structure:**
```
my-app/
├── Dockerfile
├── skaffold.yaml
├── k8s/
│   └── deployment.yaml
└── src/
    └── server.js
```

**Dockerfile:**
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY src/ ./src/
CMD ["node", "src/server.js"]
```

**k8s/deployment.yaml:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 1
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
        image: myapp  # Skaffold replaces this
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: myapp
spec:
  type: LoadBalancer
  selector:
    app: myapp
  ports:
  - port: 80
    targetPort: 8080
```

**skaffold.yaml:**
```yaml
apiVersion: skaffold/v4beta6
kind: Config
metadata:
  name: myapp
build:
  artifacts:
  - image: myapp
    docker:
      dockerfile: Dockerfile
    sync:
      manual:
      - src: "src/**/*.js"
        dest: /app/src
deploy:
  kubectl:
    manifests:
    - k8s/*.yaml
portForward:
- resourceType: service
  resourceName: myapp
  port: 80
  localPort: 8080
```

### Development Workflow

```bash
# Start development mode (watches for changes)
skaffold dev

# What Skaffold does:
# 1. Builds Docker image
# 2. Pushes to registry (or loads locally)
# 3. Deploys to Kubernetes
# 4. Streams logs
# 5. Watches for file changes
# 6. On change: syncs files OR rebuilds (depending on config)
# 7. Port-forwards service to localhost
```

**Now change `src/server.js`:**
- Skaffold detects change
- Syncs file to running container (no rebuild!)
- See changes in ~3 seconds ⚡

### File Sync vs Rebuild

```yaml
sync:
  manual:
  - src: "src/**/*.js"
    dest: /app/src  # Just copy file (FAST)
  - src: "package.json"
    dest: /app/package.json  # This triggers rebuild (npm install needed)
```

**Strategy:**
- Source code changes → file sync (fast)
- Dependency changes → full rebuild (necessary)

---

## 🟡 Solution 2: Telepresence (Hybrid Development)

**Problem:** Your laptop can't run the entire microservice architecture (50 services, 200GB memory).

**Solution:** Run most services in the cluster, run YOUR service locally.

```
┌─────────────────────────────────────────┐
│  Kubernetes Cluster                     │
│                                         │
│  [Auth Service] → [API Gateway]         │
│          ↓                              │
│  [Database] ← [Your Service (local)]   │  ← Runs on your laptop!
│          ↓                              │
│  [Cache] → [Worker]                     │
└─────────────────────────────────────────┘
```

### Install Telepresence

```bash
# Install
sudo curl -fL https://app.getambassador.io/download/tel2/linux/amd64/latest/telepresence -o /usr/local/bin/telepresence
sudo chmod a+x /usr/local/bin/telepresence

telepresence version
```

### Example: Replace Remote Service with Local

```bash
# Connect to cluster
telepresence connect

# Run your service locally (replaces 'api' deployment)
telepresence intercept api --port 8080

# Now run your local server
go run main.go
# or
npm run dev

# Traffic to 'api' service in cluster is routed to your laptop!
```

**What happens:**
1. Other services call `http://api.default.svc.cluster.local`
2. Telepresence intercepts that traffic
3. Routes it to `localhost:8080` (your local dev server)
4. Your service can call other cluster services as if it's inside the cluster

**End intercept:**
```bash
telepresence leave api
```

---

## 🟢 Solution 3: Tilt (Similar to Skaffold)

**What it is:** Alternative to Skaffold with better UI.

### Install

```bash
curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash
```

### Tiltfile (Replaces skaffold.yaml)

```python
# Tiltfile
docker_build('myapp', '.')

k8s_yaml('k8s/deployment.yaml')

k8s_resource('myapp', port_forwards=8080)
```

### Run

```bash
tilt up
```

**Opens browser with:**
- Build status
- Logs
- Resource status
- Port forwards

---

## 🟡 Solution 4: Bridge to Kubernetes (VS Code Extension)

**For VS Code users:** Integrates with IDE.

```bash
# Install extension
code --install-extension mindaro.mindaro

# In VS Code:
# 1. Open Command Palette (Ctrl+Shift+P)
# 2. "Bridge to Kubernetes: Configure"
# 3. Select service to replace
# 4. Start debugging (F5)
```

**Result:** Your local debugger-attached code runs in place of cluster service.

---

## 🔴 Local Development Best Practices

### 1. Use Separate Namespaces

```bash
# Don't pollute default namespace
kubectl create namespace dev-yourname

# Set as default
kubectl config set-context --current --namespace=dev-yourname

# Now all your experiments are isolated
```

### 2. Hot Reload in Your App

**Node.js (with nodemon):**
```json
{
  "scripts": {
    "dev": "nodemon src/server.js"
  }
}
```

**Go (with air):**
```bash
go install github.com/cosmtrek/air@latest

# .air.toml
[build]
  cmd = "go build -o ./tmp/main ."
  bin = "tmp/main"
```

**Combine with Skaffold's file sync for instant updates!**

### 3. Use Local Registries (No Push to Cloud)

**With minikube:**
```bash
eval $(minikube docker-env)
docker build -t myapp:dev .
# No push needed!
```

**With kind:**
```bash
kind load docker-image myapp:dev
```

### 4. Mock External Services

**Problem:** Your app calls AWS S3, Stripe API, etc.

**Solution:** Use mocks/stubs in local dev.

```yaml
# k8s/dev/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  S3_ENDPOINT: "http://localstack:4566"  # Mock S3
  STRIPE_API_KEY: "sk_test_fakefakefake"  # Test key
  DATABASE_URL: "postgres://postgres:postgres@db:5432/dev"
```

**Run LocalStack for AWS mocking:**
```bash
docker run -d -p 4566:4566 localstack/localstack
```

### 5. Use Lightweight Images for Dev

```dockerfile
# Dockerfile.dev
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install  # Includes devDependencies
COPY . .
CMD ["npm", "run", "dev"]  # Hot reload

# Dockerfile.prod
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --production  # No devDependencies
COPY dist/ ./dist/
CMD ["node", "dist/server.js"]
```

---

## 🎯 Recommendation Matrix

| Scenario | Tool | Why |
|----------|------|-----|
| Simple app, fast iteration | **Skaffold** | File sync, automatic rebuild |
| Large microservices system | **Telepresence** | Run only your service locally |
| Team prefers UI | **Tilt** | Beautiful web dashboard |
| VS Code user | **Bridge to Kubernetes** | IDE integration |
| Just starting | **minikube + kubectl** | Learn basics first |

---

## ✅ Hands-On Exercise: Skaffold Development

### Task: Build a Fast Development Loop

**1. Install prerequisites:**
```bash
# minikube
minikube start

# Skaffold
curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64
sudo install skaffold /usr/local/bin/
```

**2. Create Node.js app:**

```bash
mkdir fast-dev-demo && cd fast-dev-demo
npm init -y
npm install express
```

**src/server.js:**
```javascript
const express = require('express');
const app = express();

app.get('/', (req, res) => {
  res.json({ message: 'Hello from Kubernetes!' });
});

app.listen(8080, () => {
  console.log('Server running on :8080');
});
```

**3. Create Dockerfile:**

```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY src/ ./src/
CMD ["node", "src/server.js"]
```

**4. Create Kubernetes manifests:**

```bash
mkdir k8s
```

**k8s/deployment.yaml:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo
spec:
  replicas: 1
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
        image: demo-app
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: demo
spec:
  type: LoadBalancer
  selector:
    app: demo
  ports:
  - port: 80
    targetPort: 8080
```

**5. Create skaffold.yaml:**

```yaml
apiVersion: skaffold/v4beta6
kind: Config
metadata:
  name: demo
build:
  artifacts:
  - image: demo-app
    docker:
      dockerfile: Dockerfile
    sync:
      manual:
      - src: "src/**/*.js"
        dest: /app/src
deploy:
  kubectl:
    manifests:
    - k8s/*.yaml
portForward:
- resourceType: service
  resourceName: demo
  port: 80
  localPort: 3000
```

**6. Start development:**

```bash
skaffold dev
```

**7. Test:**

```bash
# In another terminal
curl http://localhost:3000
# {"message":"Hello from Kubernetes!"}
```

**8. Make a change:**

Edit `src/server.js`:
```javascript
res.json({ message: 'Hello from HOT RELOAD!' });
```

**Watch Skaffold output:**
```
File sync succeeded
Streaming logs from pod...
```

**Test again (within 3 seconds):**
```bash
curl http://localhost:3000
# {"message":"Hello from HOT RELOAD!"}
```

**No rebuild! Changes synced instantly! 🎉**

---

## 🧩 Common Mistakes

### ❌ Mistake 1: Using `:latest` Tag

```yaml
image: myapp:latest  # BAD: Kubernetes caches this
```

**Problem:** Kubernetes doesn't pull new image if tag is the same.

**Fix:**
```yaml
imagePullPolicy: Always  # For dev only!
# Or use unique tags: myapp:v1.2.3
```

### ❌ Mistake 2: Rebuilding on Every Change

**Without file sync:**
- Change one line → rebuild entire image → 2 minutes

**With file sync (Skaffold/Tilt):**
- Change one line → copy file → 2 seconds

### ❌ Mistake 3: Running Full Stack Locally

**Problem:** Laptop can't run 20 microservices.

**Fix:** Use Telepresence to run only your service locally.

### ❌ Mistake 4: Not Using Namespaces

**Problem:** Experiments pollute shared cluster.

**Fix:** Use personal namespace (`dev-yourname`).

---

## 📚 Summary

**Fast Kubernetes Development:**

1. **Local Cluster**: minikube or kind
2. **Hot Reload Tool**: Skaffold or Tilt
3. **Hybrid Mode**: Telepresence (for large systems)
4. **File Sync**: Avoid rebuilds for code changes
5. **Namespaces**: Isolate your experiments

**Development Speed:**
- ❌ Manual: 5-10 minutes per change
- ✅ Skaffold: 5-10 seconds per change
- ✅ Telepresence: Native local dev speed

**Recommended Workflow:**
```bash
# Day 1: Learn Kubernetes
minikube start
kubectl apply -f app.yaml

# Day 7: Fast iteration
skaffold dev

# Day 30: Complex systems
telepresence intercept my-service --port 8080
```

---

**Previous:** [Module 03: Kubernetes Fundamentals](../03-kubernetes-fundamentals/README.md)  
**Next:** [02. Debugging Pods and Containers](./02-debugging-pods.md)  
**Module:** [04. Kubernetes for Developers](./README.md)
