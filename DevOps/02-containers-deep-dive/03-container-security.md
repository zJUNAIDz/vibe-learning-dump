# Container Security

🔴 **Advanced**

---

## Why Container Security Matters

**Reality:** Containers share the host kernel. A compromised container **can attack the host**.

```
┌──────────────────────────────┐
│        Host Kernel           │ ← Single point of failure
├────────┬─────────┬───────────┤
│ App A  │  App B  │  App C    │ ← All share kernel
└────────┴─────────┴───────────┘
```

**If App B gets compromised and breaks out, it can:**
- Attack host
- Attack other containers
- Steal secrets
- Mine crypto
- Join a botnet

---

## Attack Surface

### 1. **The Image**
- Vulnerable base images
- Malicious dependencies
- Embedded secrets

### 2. **The Runtime**
- Running as root
- Excessive capabilities
- Writable filesystem

### 3. **The Host**
- Docker socket exposed
- Kubelet API accessible
- Shared volumes with sensitive data

---

## Defense #1: Use Trusted Base Images

### Bad
```dockerfile
FROM random-guy/ubuntu:latest
# Who is random-guy? What's in this image?
```

### Good
```dockerfile
# Official images
FROM node:18-alpine
FROM golang:1.21-alpine
FROM postgres:14-alpine

# Or build from scratch
FROM scratch
```

**Verify images:**
```bash
# Check image provenance
docker trust inspect node:18-alpine

# Scan for known vulnerabilities
trivy image node:18-alpine
```

---

## Defense #2: Minimize Attack Surface

### Smaller Image = Fewer Vulnerabilities

```
ubuntu:latest     → 77 MB  → 100+ packages → many CVEs
alpine:latest     → 5 MB   → 14 packages   → fewer CVEs
scratch           → 0 MB   → 0 packages    → zero CVEs
```

**Example: Comparing CVEs**
```bash
trivy image ubuntu:latest | grep CVE | wc -l
# 200+ CVEs

trivy image alpine:latest | grep CVE | wc -l
# 10-20 CVEs

trivy image myapp-from-scratch | grep CVE | wc -l
# 0 CVEs (just your binary)
```

---

## Defense #3: Never Run as Root

**Why running as root is dangerous:**
```
Container process → Runs as root (UID 0)
                 → Escapes container (kernel exploit)
                 → Now root on host
                 → Game over
```

### Bad
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY . .
CMD ["node", "server.js"]
# Runs as root by default!
```

### Good
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY . .

# Create non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001

# Change ownership
RUN chown -R nodejs:nodejs /app

# Switch to non-root
USER nodejs

CMD ["node", "server.js"]
```

**Or use built-in user:**
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY . .
USER node  # node user already exists
CMD ["node", "server.js"]
```

---

## Defense #4: Read-Only Filesystem

Make the container filesystem **read-only** to prevent tampering.

```bash
docker run --read-only --tmpfs /tmp myapp
```

**Why:**
- Attacker can't write malware to disk
- Can't modify binaries
- Can't persist backdoors

**Caveat:** App needs `/tmp` for temporary files (hence `--tmpfs /tmp`).

---

## Defense #5: Drop Capabilities

Linux **capabilities** are fine-grained permissions (less than root, more than user).

**Default capabilities Docker grants:**
```
CAP_CHOWN, CAP_DAC_OVERRIDE, CAP_FOWNER, CAP_FSETID,
CAP_KILL, CAP_SETGID, CAP_SETUID, CAP_SETPCAP,
CAP_NET_BIND_SERVICE, CAP_NET_RAW, CAP_SYS_CHROOT,
CAP_MKNOD, CAP_AUDIT_WRITE, CAP_SETFCAP
```

**Most apps don't need any of these.**

### Drop All Capabilities
```bash
docker run --cap-drop=ALL myapp
```

### Add Only What's Needed
```bash
docker run --cap-drop=ALL --cap-add=NET_BIND_SERVICE myapp
# Only allows binding to ports < 1024
```

---

## Defense #6: Use seccomp Profiles

**seccomp** restricts system calls a container can make.

**Default profile:** Docker blocks ~44 dangerous syscalls (e.g., reboot, module loading).

**Custom profile:**
```json
{
  "defaultAction": "SCMP_ACT_ERRNO",
  "architectures": ["SCMP_ARCH_X86_64"],
  "syscalls": [
    { "names": ["read", "write", "open", "close", "stat"], "action": "SCMP_ACT_ALLOW" },
    { "names": ["listen", "bind", "accept"], "action": "SCMP_ACT_ALLOW" }
  ]
}
```

**Usage:**
```bash
docker run --security-opt seccomp=profile.json myapp
```

---

## Defense #7: AppArmor / SELinux

**AppArmor** (Ubuntu/Debian) and **SELinux** (RHEL/Fedora) provide **mandatory access control (MAC)**.

### SELinux (Fedora)
```bash
# Check SELinux status
sestatus

# Run container with SELinux label
docker run --security-opt label=type:container_t myapp

# Audit denials
sudo ausearch -m avc -ts recent
```

**Kubernetes automatically applies SELinux labels on RHEL/Fedora.**

---

## Defense #8: Don't Expose Docker Socket

**The Docker socket (`/var/run/docker.sock`) is root on the host.**

### Extremely Dangerous
```bash
docker run -v /var/run/docker.sock:/var/run/docker.sock myapp
# Container can now control Docker (create/delete containers, run commands as root)
```

**If compromised:**
```bash
# Inside container
docker run --rm -it -v /:/host alpine chroot /host /bin/bash
# Now you're root on the HOST, not the container
```

**Rule:** **NEVER mount the Docker socket unless absolutely necessary.**

---

## Defense #9: Use Distroless Images (Google)

**Distroless** images contain only the app and its runtime dependencies (no shell, no package manager).

```dockerfile
# Go example
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o myapp .

FROM gcr.io/distroless/static-debian11
COPY --from=builder /app/myapp /myapp
CMD ["/myapp"]
```

**Why distroless:**
- No shell → can't `docker exec` or get a shell if compromised
- Minimal attack surface
- Smaller image size

**Available variants:**
- `gcr.io/distroless/static-debian11` → For static binaries (Go)
- `gcr.io/distroless/base-debian11` → Minimal libc, tzdata
- `gcr.io/distroless/nodejs-debian11` → Node.js runtime

---

## Defense #10: Scan Images Continuously

**Vulnerabilities are discovered constantly.** An image that was safe yesterday may not be today.

### Trivy (Open Source)
```bash
# Scan local image
trivy image myapp:latest

# Scan and fail CI if HIGH/CRITICAL found
trivy image --exit-code 1 --severity HIGH,CRITICAL myapp:latest

# Scan filesystem (for Dockerfile, etc.)
trivy fs .
```

### Docker Scout (Built-in)
```bash
docker scout cves myapp:latest
docker scout recommendations myapp:latest
```

### Grype (Anchore)
```bash
grype myapp:latest
```

---

## Defense #11: Sign and Verify Images

**Docker Content Trust** uses The Update Framework (TUF) to sign images.

### Enable
```bash
export DOCKER_CONTENT_TRUST=1
docker push myapp:latest
# Prompts for signing key

docker pull myapp:latest
# Verifies signature before pulling
```

**Why:**
- Guarantees image authenticity
- Prevents man-in-the-middle attacks

---

## Defense #12: Limit Resources

Prevent resource exhaustion attacks (DoS).

```bash
docker run \
  --memory=512m \
  --memory-swap=512m \
  --cpus=1 \
  --pids-limit=100 \
  myapp
```

**In Kubernetes:**
```yaml
resources:
  limits:
    memory: "512Mi"
    cpu: "1"
  requests:
    memory: "256Mi"
    cpu: "500m"
```

---

## Docker Bench Security

Automated security audit for Docker.

```bash
git clone https://github.com/docker/docker-bench-security.git
cd docker-bench-security
sudo sh docker-bench-security.sh
```

**Checks:**
- Host configuration
- Docker daemon config
- Image and container best practices
- Secrets management

---

## Kubernetes Security Context

In Kubernetes, security policies are defined in `securityContext`.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: secure-pod
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    fsGroup: 2000
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: app
    image: myapp:latest
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
          - ALL
    volumeMounts:
    - name: tmp
      mountPath: /tmp
  volumes:
  - name: tmp
    emptyDir: {}
```

---

## War Story: The Crypto Miner

A team deployed a Node.js app to Kubernetes. After a week, they noticed:
- AWS bill increased 10x
- Nodes running at 100% CPU

**Investigation:**
```bash
kubectl top nodes
# All nodes maxed out

kubectl top pods
# One pod using 15 cores (limit was 2)
```

**Root cause:**
- App had an RCE (remote code execution) vulnerability
- Attacker injected crypto mining script
- No resource limits enforced (Kubernetes bug: limits not applied)

**Fixes:**
1. Patched RCE vulnerability
2. Set resource limits (and verified they worked)
3. Added monitoring alerts for CPU spikes
4. Enabled Pod Security Standards

**Lesson:** Defense in depth. One layer will fail.

---

## Key Takeaways

1. **Containers share kernel** → weaker isolation than VMs
2. **Never run as root** → use `USER` directive
3. **Drop all capabilities** → add only what's needed
4. **Read-only filesystem** → prevents tampering
5. **Never expose Docker socket** → it's root on host
6. **Scan images continuously** → Trivy, Docker Scout, Grype
7. **Use distroless or alpine** → minimal attack surface
8. **Sign images** → prevent supply chain attacks
9. **Enforce resource limits** → prevent DoS
10. **Defense in depth** → no single layer is perfect

---

## Exercises

1. **Run as non-root:**
   - Create a Dockerfile that runs as `USER 1000`
   - Try to write to `/` (should fail)

2. **Drop all capabilities:**
   ```bash
   docker run --cap-drop=ALL alpine sh
   # Try to bind to port 80 (should fail)
   ```

3. **Scan your images:**
   ```bash
   trivy image myapp:latest
   # Review CVEs, fix what you can
   ```

4. **Kubernetes security context:**
   - Deploy a pod with `readOnlyRootFilesystem: true`
   - Mount `/tmp` as `emptyDir`
   - Verify app works

---

**Module 02 Complete!** 🎉

You now understand:
- ✅ What containers actually are (cgroups, namespaces, layers)
- ✅ Dockerfile optimization (multi-stage, caching, alpine)
- ✅ Container security (non-root, capabilities, scanning)

---

**Next Module:** [03. Kubernetes from First Principles →](../03-kubernetes-fundamentals/01-why-kubernetes-exists.md)
