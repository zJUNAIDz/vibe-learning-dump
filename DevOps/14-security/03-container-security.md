# Container Security

> **A container running as root with a writable filesystem and no resource limits is not a container. It's a liability. Lock it down BEFORE it reaches production.**

---

## 🟢 Non-Root Users

### Why Root in Containers Is Dangerous

```
By default, most containers run as root (UID 0).

Why this is bad:
  → If an attacker exploits your app, they have ROOT inside the container
  → Container escapes (rare but real) give them root on the HOST
  → Root can modify any file in the container
  → Root can install tools (wget, curl, nc for data exfiltration)
  → Root can access mounted secrets

Reality:
  Most applications need ZERO root capabilities.
  Your Node.js app doesn't need to be root.
  Your Go binary doesn't need to be root.
```

### Running as Non-Root

```dockerfile
# GOOD Dockerfile — non-root user
FROM node:20-slim AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build

FROM node:20-slim
WORKDIR /app

# Create non-root user
RUN groupadd -r appuser && useradd -r -g appuser -s /bin/false appuser

# Copy built app
COPY --from=builder --chown=appuser:appuser /app/dist ./dist
COPY --from=builder --chown=appuser:appuser /app/node_modules ./node_modules
COPY --from=builder --chown=appuser:appuser /app/package.json ./

# Switch to non-root user
USER appuser

# Cannot bind to ports < 1024 (that's fine, use 3000+)
EXPOSE 3000
CMD ["node", "dist/server.js"]
```

```dockerfile
# Go binary — even simpler
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /server

FROM scratch
COPY --from=builder /server /server
# scratch has no users, but UID works
USER 65534
ENTRYPOINT ["/server"]
```

### Kubernetes Security Context

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  template:
    spec:
      securityContext:
        runAsNonRoot: true        # Reject if image tries to run as root
        runAsUser: 1000           # UID to run as
        runAsGroup: 1000          # GID to run as
        fsGroup: 1000             # Files created by volume mounts
        seccompProfile:
          type: RuntimeDefault    # Enable seccomp
      containers:
        - name: app
          image: my-app:latest
          securityContext:
            allowPrivilegeEscalation: false  # Can't sudo
            readOnlyRootFilesystem: true     # Can't write to disk
            capabilities:
              drop: ["ALL"]       # Remove all Linux capabilities
          # If app needs to write temp files:
          volumeMounts:
            - name: tmp
              mountPath: /tmp
      volumes:
        - name: tmp
          emptyDir: {}
```

---

## 🟡 Image Scanning

### Why Scan Images

```
Your container image includes:
  → Base OS packages (Ubuntu, Alpine, etc.)
  → Language runtime (Node.js, Python, Go, Java)
  → Your dependencies (npm packages, pip packages)
  → Your application code

Any of these can have known vulnerabilities (CVEs).

Example: Log4Shell (CVE-2021-44228)
  → Affected: Any image with Log4j 2.0-2.14.1
  → Impact: Remote code execution
  → If you scanned your images, you'd know in minutes
  → If you didn't, you'd find out from the news
```

### Scanning with Trivy

```bash
# Trivy is free, fast, and widely used
brew install trivy

# Scan a local image
trivy image my-app:latest

# Scan and fail on high/critical vulnerabilities
trivy image --severity HIGH,CRITICAL --exit-code 1 my-app:latest

# Scan before pushing in CI
# .github/workflows/build.yaml
jobs:
  build:
    steps:
      - name: Build image
        run: docker build -t my-app:${{ github.sha }} .
      
      - name: Scan image
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: my-app:${{ github.sha }}
          severity: 'CRITICAL,HIGH'
          exit-code: '1'
          format: 'table'
```

### Scan Output Example

```
my-app:latest (ubuntu 22.04)
=============================
Total: 12 (HIGH: 8, CRITICAL: 4)

┌──────────────┬────────────────┬──────────┬───────────┬──────────────────┐
│   Library    │ Vulnerability  │ Severity │  Version  │  Fixed Version   │
├──────────────┼────────────────┼──────────┼───────────┼──────────────────┤
│ openssl      │ CVE-2024-XXXX  │ CRITICAL │ 3.0.2     │ 3.0.13           │
│ curl         │ CVE-2024-YYYY  │ HIGH     │ 7.81.0    │ 7.81.0-1ubuntu1.15│
│ express      │ CVE-2024-ZZZZ  │ HIGH     │ 4.18.1    │ 4.18.3           │
└──────────────┴────────────────┴──────────┴───────────┴──────────────────┘

Fix:
  → Update base image: FROM node:20-slim (rebuilt weekly with patches)
  → Update dependencies: npm audit fix
  → Rebuild and rescan
```

### Minimal Base Images

```
Image size = attack surface. Smaller = safer.

Image                  Size      Packages    Risk
───────────────────    ────      ────────    ────
ubuntu:22.04           77MB      412         HIGH
node:20                1.1GB     1000+       HIGH
node:20-slim           240MB     ~200        MEDIUM
node:20-alpine         180MB     ~50         LOW
distroless/nodejs20    ~130MB    ~20         VERY LOW
scratch                0MB       0           MINIMAL

Recommendation:
  → Use Alpine or slim variants for most apps
  → Use distroless for production (no shell = attackers can't exec in)
  → Use scratch for static Go/Rust binaries
```

```dockerfile
# Distroless — no shell, no package manager, nothing extra
FROM node:20-slim AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build

FROM gcr.io/distroless/nodejs20-debian12
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
CMD ["dist/server.js"]
# No shell to exec into, no apt-get, no curl
```

---

## 🟡 Seccomp, AppArmor, SELinux

### Seccomp (Secure Computing)

```
Seccomp restricts which SYSTEM CALLS a container can make.

Linux has ~450 system calls. Most apps use ~50.
Seccomp blocks the other ~400.

Default Docker/K8s profile blocks dangerous calls:
  → mount (can't mount host filesystem)
  → reboot (can't reboot the host)
  → sethostname (can't change hostname)
  → etc.
```

```yaml
# Enable default seccomp profile in Kubernetes
securityContext:
  seccompProfile:
    type: RuntimeDefault  # Docker/containerd default profile

# Custom seccomp profile (advanced)
securityContext:
  seccompProfile:
    type: Localhost
    localhostProfile: profiles/my-app.json
```

### AppArmor

```
AppArmor restricts what FILES and CAPABILITIES a container can access.
Common on Ubuntu/Debian systems.

Default Docker AppArmor profile:
  → Blocks writing to /proc, /sys
  → Blocks mounting filesystems
  → Blocks changing network config
```

```yaml
# Kubernetes annotation for AppArmor
metadata:
  annotations:
    container.apparmor.security.beta.kubernetes.io/app: runtime/default
```

### Summary

```
Layer          What it restricts              When to use
─────          ──────────────────             ───────────
seccomp        System calls                   Always (RuntimeDefault)
AppArmor       File access, capabilities      Ubuntu/Debian systems
SELinux        File access (label-based)      RHEL/CentOS systems

Start with RuntimeDefault seccomp. That covers 90% of security needs.
```

---

## 🔴 Pod Security Standards

```
Kubernetes Pod Security Standards define three levels:

1. PRIVILEGED (no restrictions)
   → For system pods (CNI, logging, monitoring)
   → Almost never for application pods

2. BASELINE (prevents known privilege escalations)
   → No hostNetwork, hostPID, hostIPC
   → No privileged containers
   → No dangerous capabilities
   → Good default for most apps

3. RESTRICTED (heavily restricted)
   → Must run as non-root
   → Must drop ALL capabilities
   → Must use readOnlyRootFilesystem
   → Must use seccomp RuntimeDefault
   → Best for production workloads
```

```yaml
# Enforce restricted standard on a namespace
apiVersion: v1
kind: Namespace
metadata:
  name: production
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/warn: restricted
    pod-security.kubernetes.io/audit: restricted

# Any pod violating "restricted" in this namespace
# will be REJECTED by the API server
```

---

## 🔴 Anti-Patterns

```
❌ Running as root "because it works"
   → Fix: Add USER to Dockerfile, set runAsNonRoot in K8s

❌ Using latest tag
   → FROM node:latest ← Which version? Changes without warning!
   → Fix: Pin versions — FROM node:20.11.1-slim

❌ Not scanning images in CI
   → "We only scan quarterly"
   → Fix: Scan on every build, fail pipeline on CRITICAL

❌ Fat images with dev tools
   → Production image has gcc, make, vim, curl
   → Fix: Multi-stage builds, distroless/alpine

❌ Mounting Docker socket
   → volumes: ["/var/run/docker.sock:/var/run/docker.sock"]
   → This gives container FULL control of the host
   → Only for CI/CD build agents with extreme caution

❌ No resource limits
   → A compromised container can consume ALL host CPU/memory
   → Fix: Always set resources.limits in K8s
```

---

**Previous:** [02. IAM Concepts](./02-iam-concepts.md)  
**Next:** [04. Supply Chain Security](./04-supply-chain-security.md)
