# Ingress and Load Balancing

> **Exposing services to the outside world — understanding how traffic reaches your pods**

---

## 🟢 The Problem: How Do Users Reach Your App?

```
User's Browser
    ↓
    ??? (How does traffic get into the cluster?)
    ↓
Your Pod
```

**Kubernetes provides 3 solutions:**

1. **NodePort** — Expose service on every node's IP
2. **LoadBalancer** — Cloud provider creates a load balancer
3. **Ingress** — HTTP/HTTPS routing (most common)

---

## 🟢 Service Types Recap

### ClusterIP (Default - Internal Only)

```yaml
apiVersion: v1
kind: Service
metadata:
  name: api-service
spec:
  type: ClusterIP
  selector:
    app: api
  ports:
  - port: 80
    targetPort: 8080
```

**Access:** Only inside the cluster.

**Use case:** Internal services (databases, caches, microservices).

### NodePort (External Access via Node IP)

```yaml
apiVersion: v1
kind: Service
metadata:
  name: api-service
spec:
  type: NodePort
  selector:
    app: api
  ports:
  - port: 80
    targetPort: 8080
    nodePort: 30080  # Exposed on all nodes
```

**Access:** `http://<node-ip>:30080`

**Port range:** 30000-32767 (by default)

**Problems:**
- Need to know node IPs (they change!)
- No SSL termination
- No path-based routing
- Port collision across services

**Use case:** Development, testing, or when you have no other option.

### LoadBalancer (Cloud Provider Integration)

```yaml
apiVersion: v1
kind: Service
metadata:
  name: api-service
spec:
  type: LoadBalancer
  selector:
    app: api
  ports:
  - port: 80
    targetPort: 8080
```

**What happens (on AWS/GCP/Azure):**
1. Kubernetes asks cloud provider for a load balancer
2. Cloud provider creates ELB/ALB (AWS), Cloud Load Balancer (GCP), Azure LB
3. Load balancer gets a public IP
4. Traffic → LB → NodePort → Pod

**Check external IP:**
```bash
kubectl get service api-service
# NAME          TYPE           EXTERNAL-IP       PORT(S)
# api-service   LoadBalancer   34.123.45.67      80:31234/TCP
```

**Access:** `http://34.123.45.67`

**Problems:**
- Costs $$$ (one LB per service!)
- No HTTP routing (can't do `/api` → service1, `/web` → service2)
- Still need to configure SSL manually

**Use case:** Simple deployments, databases that need external access.

---

## 🟡 Ingress: HTTP(S) Routing

### Mental Model: Ingress is a Smart Reverse Proxy

```
                    Ingress Controller
                          ↓
User → ingress.example.com/api     → api-service     → api pods
User → ingress.example.com/web     → web-service     → web pods
User → ingress.example.com/admin   → admin-service   → admin pods
```

**Key Concepts:**
- **Ingress Resource** — Defines routing rules (YAML)
- **Ingress Controller** — Runs the actual proxy (NGINX, Traefik, etc.)

### Installing NGINX Ingress Controller

```bash
# Install NGINX Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/cloud/deploy.yaml

# Check installation
kubectl get pods -n ingress-nginx

# Wait for LoadBalancer IP
kubectl get service -n ingress-nginx ingress-nginx-controller
```

**On cloud providers (AWS/GCP/Azure):** This creates a LoadBalancer that forwards traffic to the Ingress Controller.

**On minikube:**
```bash
minikube addons enable ingress
```

---

## 🟡 Creating an Ingress Resource

### Example 1: Simple Path-Based Routing

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  ingressClassName: nginx
  rules:
  - host: myapp.example.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: api-service
            port:
              number: 80
      - path: /web
        pathType: Prefix
        backend:
          service:
            name: web-service
            port:
              number: 80
```

**What this does:**
- `myapp.example.com/api/*` → routes to `api-service`
- `myapp.example.com/web/*` → routes to `web-service`

**Deploy:**
```bash
kubectl apply -f ingress.yaml
kubectl get ingress app-ingress
```

**Test:**
```bash
# Get Ingress IP
INGRESS_IP=$(kubectl get ingress app-ingress -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

# Test (with Host header)
curl -H "Host: myapp.example.com" http://$INGRESS_IP/api
curl -H "Host: myapp.example.com" http://$INGRESS_IP/web
```

---

### Example 2: Host-Based Routing

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: multi-host-ingress
spec:
  ingressClassName: nginx
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-service
            port:
              number: 80
  - host: web.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: web-service
            port:
              number: 80
```

**What this does:**
- `api.example.com/*` → `api-service`
- `web.example.com/*` → `web-service`

---

## 🟡 SSL/TLS Termination

### Step 1: Create SSL Certificate

**Option A: Using cert-manager (automatic Let's Encrypt)**

```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Create ClusterIssuer for Let's Encrypt
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

**Option B: Using existing certificate**

```bash
# Create TLS secret from certificate files
kubectl create secret tls myapp-tls \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key
```

### Step 2: Configure Ingress with TLS

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: tls-ingress
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"  # For automatic certs
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - myapp.example.com
    secretName: myapp-tls  # TLS certificate stored here
  rules:
  - host: myapp.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: web-service
            port:
              number: 80
```

**What happens:**
1. User visits `https://myapp.example.com`
2. Ingress Controller terminates SSL
3. Backend traffic is HTTP (inside cluster)

**Test:**
```bash
curl https://myapp.example.com
```

---

## 🟢 PathType: Prefix vs Exact

```yaml
paths:
- path: /api
  pathType: Prefix  # Matches /api, /api/, /api/users, etc.
  
- path: /exact-path
  pathType: Exact   # Only matches /exact-path (not /exact-path/)
```

**Use Prefix for most cases.**

---

## 🟡 Common Annotations

### NGINX Ingress Annotations

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: advanced-ingress
  annotations:
    # Redirect HTTP to HTTPS
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    
    # Rewrite paths
    nginx.ingress.kubernetes.io/rewrite-target: /$2
    
    # CORS
    nginx.ingress.kubernetes.io/enable-cors: "true"
    nginx.ingress.kubernetes.io/cors-allow-origin: "*"
    
    # Rate limiting
    nginx.ingress.kubernetes.io/limit-rps: "10"
    
    # Connection limits
    nginx.ingress.kubernetes.io/limit-connections: "100"
    
    # Custom timeouts
    nginx.ingress.kubernetes.io/proxy-connect-timeout: "600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "600"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "600"
    
    # Whitelist IPs
    nginx.ingress.kubernetes.io/whitelist-source-range: "10.0.0.0/8,192.168.0.0/16"
spec:
  # ... rest of ingress config
```

### Rewrite Example

```yaml
# URL pattern: /api/v1/users → backend receives /users
annotations:
  nginx.ingress.kubernetes.io/rewrite-target: /$2
  
rules:
- host: api.example.com
  http:
    paths:
    - path: /api/v1(/|$)(.*)
      pathType: Prefix
      backend:
        service:
          name: api-service
          port:
            number: 80
```

---

## 🔴 Load Balancing Algorithms

### NGINX Ingress: Round Robin (Default)

```
Request 1 → Pod A
Request 2 → Pod B
Request 3 → Pod C
Request 4 → Pod A
...
```

### Sticky Sessions (Session Affinity)

**Problem:** User's session data is on Pod A, but next request goes to Pod B.

**Solution:** Sticky sessions (route same client to same pod).

```yaml
annotations:
  nginx.ingress.kubernetes.io/affinity: "cookie"
  nginx.ingress.kubernetes.io/session-cookie-name: "route"
  nginx.ingress.kubernetes.io/session-cookie-expires: "172800"
  nginx.ingress.kubernetes.io/session-cookie-max-age: "172800"
```

**How it works:** Sets a cookie; subsequent requests with that cookie go to the same pod.

**Better solution:** Use stateless services with external session storage (Redis).

---

## 🔴 War Story: The Missing Ingress Controller

> *"We deployed 50 Ingress resources. None of them worked. Spent 3 hours debugging. Turns out we never installed an Ingress Controller. Kubernetes accepted the Ingress resources but nothing was processing them. Always check: `kubectl get pods -n ingress-nginx`"*

**Lesson:** Ingress resources do nothing without an Ingress Controller!

---

## 🟡 Debugging Ingress

### 1. Check Ingress Status

```bash
kubectl get ingress
kubectl describe ingress myapp-ingress

# Look for:
# - ADDRESS column (should have IP or hostname)
# - Events (errors will show here)
```

### 2. Check Ingress Controller Logs

```bash
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller

# Look for errors related to your ingress
```

### 3. Check Backend Service

```bash
# Make sure service exists
kubectl get service api-service

# Make sure pods are running
kubectl get pods -l app=api

# Test service directly (from inside cluster)
kubectl run test --image=curlimages/curl -it --rm -- curl http://api-service
```

### 4. Test with Host Header

```bash
# Get Ingress IP
INGRESS_IP=$(kubectl get ingress -o jsonpath='{.items[0].status.loadBalancer.ingress[0].ip}')

# Test with curl
curl -H "Host: myapp.example.com" http://$INGRESS_IP
```

### 5. Check DNS

```bash
# Make sure domain points to Ingress IP
dig myapp.example.com
nslookup myapp.example.com
```

---

## 🎯 Real-World Example: Full Stack App

```yaml
# Backend Service
apiVersion: v1
kind: Service
metadata:
  name: api
spec:
  selector:
    app: api
  ports:
  - port: 80
    targetPort: 8080
---
# Frontend Service
apiVersion: v1
kind: Service
metadata:
  name: web
spec:
  selector:
    app: web
  ports:
  - port: 80
    targetPort: 3000
---
# Ingress
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: fullstack-ingress
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - myapp.com
    secretName: myapp-tls
  rules:
  - host: myapp.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: api
            port:
              number: 80
      - path: /
        pathType: Prefix
        backend:
          service:
            name: web
            port:
              number: 80
```

**Traffic flow:**
- `https://myapp.com/` → web service → React app
- `https://myapp.com/api/users` → api service → Go backend

---

## ✅ Hands-On Exercise

### Task: Deploy Two Services with Ingress

**1. Deploy two simple services:**

```yaml
# service-a.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-a
spec:
  replicas: 2
  selector:
    matchLabels:
      app: app-a
  template:
    metadata:
      labels:
        app: app-a
    spec:
      containers:
      - name: app
        image: hashicorp/http-echo
        args:
        - "-text=Hello from App A"
        ports:
        - containerPort: 5678
---
apiVersion: v1
kind: Service
metadata:
  name: service-a
spec:
  selector:
    app: app-a
  ports:
  - port: 80
    targetPort: 5678
```

```yaml
# service-b.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-b
spec:
  replicas: 2
  selector:
    matchLabels:
      app: app-b
  template:
    metadata:
      labels:
        app: app-b
    spec:
      containers:
      - name: app
        image: hashicorp/http-echo
        args:
        - "-text=Hello from App B"
        ports:
        - containerPort: 5678
---
apiVersion: v1
kind: Service
metadata:
  name: service-b
spec:
  selector:
    app: app-b
  ports:
  - port: 80
    targetPort: 5678
```

**2. Create Ingress:**

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  ingressClassName: nginx
  rules:
  - http:
      paths:
      - path: /a
        pathType: Prefix
        backend:
          service:
            name: service-a
            port:
              number: 80
      - path: /b
        pathType: Prefix
        backend:
          service:
            name: service-b
            port:
              number: 80
```

**3. Deploy:**

```bash
kubectl apply -f service-a.yaml
kubectl apply -f service-b.yaml
kubectl apply -f ingress.yaml
```

**4. Test:**

```bash
# Get Ingress IP
INGRESS_IP=$(kubectl get ingress test-ingress -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

# Test Service A
curl http://$INGRESS_IP/a
# Hello from App A

# Test Service B
curl http://$INGRESS_IP/b
# Hello from App B
```

---

## 📚 Summary

| Method | Use Case | Pros | Cons |
|--------|----------|------|------|
| **ClusterIP** | Internal services | Simple, secure | Not externally accessible |
| **NodePort** | Development | Simple, no dependencies | Requires knowing node IPs, limited ports |
| **LoadBalancer** | Single service | Simple, cloud-native | Expensive (one LB per service) |
| **Ingress** | Multiple services | Cost-effective, routing, SSL | Requires ingress controller |

**Decision Tree:**

```
Need external access?
├─ NO → ClusterIP
└─ YES → Multiple services or HTTP routing?
           ├─ NO → LoadBalancer (simple)
           └─ YES → Ingress (path/host routing, SSL)
```

**Key Takeaways:**
1. Ingress is the standard way to expose HTTP services
2. Ingress Controller must be installed separately
3. One LoadBalancer → Many services via Ingress (cost-effective)
4. Use cert-manager for automatic SSL certificates
5. Path-based routing: `/api` → backend, `/` → frontend

---

**Previous:** [04. Health Checks and Probes](./04-health-checks-probes.md)  
**Next:** [06. Namespaces and Resource Quotas](./06-namespaces-resource-quotas.md)  
**Module:** [03. Kubernetes Fundamentals](./README.md)
