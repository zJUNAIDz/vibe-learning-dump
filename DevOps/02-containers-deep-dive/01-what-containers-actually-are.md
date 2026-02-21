# What Containers Actually Are

🟡 **Intermediate**

---

## The Complete Picture

You've learned that containers use **cgroups** and **namespaces**. Now let's add the third piece:

```
Container = cgroups + namespaces + layered filesystem

cgroups           → Resource limits
namespaces        → Isolation  
layered filesystem → Images, sharing, efficiency
```

---

## Layered Filesystems (Union FS)

### The Problem

Without layers:
```
Image A (Node.js app):     1 GB
Image B (Another Node.js): 1 GB
Image C (Another Node.js): 1 GB

Total disk: 3 GB (lots of duplication!)
```

With layers:
```
Base layer (Ubuntu):       100 MB  ← Shared
Node.js layer:             200 MB  ← Shared
App A layer:               10 MB
App B layer:               15 MB
App C layer:               12 MB

Total disk: 100 + 200 + 10 + 15 + 12 = 337 MB
```

---

## How Image Layers Work

An image is a **stack of read-only layers**.

```
┌─────────────────────┐
│  App files          │ ← Layer 3 (your app)
├─────────────────────┤
│  npm install        │ ← Layer 2 (dependencies)
├─────────────────────┤
│  Node.js, npm       │ ← Layer 1 (runtime)
├─────────────────────┤
│  Ubuntu base        │ ← Layer 0 (OS)
└─────────────────────┘
```

**Key insight:** 
- Layers are **immutable** (read-only)
- Docker uses **copy-on-write**: if a container modifies a file, it copies it to a new layer

---

## Copy-on-Write (CoW)

1. Container starts from image (read-only layers)
2. Docker adds a **writable layer** on top
3. Container reads files normally (from lower layers)
4. If container **writes** to a file:
   - File is copied from lower layer to writable layer
   - Changes happen in writable layer only
   - Original file unchanged

```
Container view:
┌─────────────────────┐
│ Writable layer      │ ← Changes go here
├─────────────────────┤
│ Read-only layers    │ ← Original image
└─────────────────────┘
```

---

## Union Filesystems in Docker

Docker supports multiple filesystems:

| Filesystem | Status | Notes |
|------------|--------|-------|
| **overlay2** | Default (modern) | Fast, efficient |
| **aufs** | Old (deprecated) | Slower |
| **btrfs** | Specialized | CoW built into filesystem |
| **zfs** | Specialized | CoW built into filesystem |

**Check yours:**
```bash
docker info | grep "Storage Driver"
```

---

## Anatomy of a Container

```
┌────────────────────────────────────────┐
│          Container                     │
│                                        │
│  ┌──────────────────────────────────┐ │
│  │  Process tree (PID namespace)    │ │
│  │  - PID 1: /app/server            │ │
│  │  - PID 2: /app/worker            │ │
│  └──────────────────────────────────┘ │
│                                        │
│  ┌──────────────────────────────────┐ │
│  │  Network (NET namespace)         │ │
│  │  - eth0: 172.17.0.2              │ │
│  │  - lo: 127.0.0.1                 │ │
│  └──────────────────────────────────┘ │
│                                        │
│  ┌──────────────────────────────────┐ │
│  │  Filesystem (MNT namespace)      │ │
│  │  - Layered image + writable top  │ │
│  └──────────────────────────────────┘ │
│                                        │
│  ┌──────────────────────────────────┐ │
│  │  Hostname (UTS namespace)        │ │
│  │  - f3a8b2c1e234                  │ │
│  └──────────────────────────────────┘ │
│                                        │
│  ┌──────────────────────────────────┐ │
│  │  Resource limits (cgroups)       │ │
│  │  - CPU: 1 core                   │ │
│  │  - Memory: 512 MB                │ │
│  └──────────────────────────────────┘ │
└────────────────────────────────────────┘
```

---

## Container Lifecycle

```
1. CREATE
   docker create <image>
   → Container created (not running)
   → Filesystem layers allocated
   → Namespaces NOT yet created

2. START
   docker start <container>
   → Namespaces created
   → cgroups applied
   → Process (PID 1) started

3. RUNNING
   → Process executing
   → Consuming resources

4. STOP
   docker stop <container>
   → Sends SIGTERM to PID 1
   → Waits 10 seconds (default)
   → Sends SIGKILL if still alive

5. REMOVE
   docker rm <container>
   → Namespaces destroyed
   → Writable layer deleted
   → Base image layers remain (shared)
```

---

## Images vs Containers

| Image | Container |
|-------|-----------|
| Read-only template | Running instance |
| Stored on disk | Exists in memory |
| Built once | Created many times |
| Defined by Dockerfile | Defined by `docker run` args |
| Immutable | Has writable layer |

**Analogy:**
```
Image = Class
Container = Object instance
```

---

## Docker Architecture

```
┌──────────────────────────────────────────────────┐
│                    Docker CLI                    │
│                   (docker ...)                   │
└────────────────────┬─────────────────────────────┘
                     │
                     │ REST API
                     ↓
┌──────────────────────────────────────────────────┐
│               Docker Daemon (dockerd)            │
│                                                  │
│  ┌──────────────┐  ┌──────────────┐            │
│  │ Image Mgmt   │  │ Container    │            │
│  │              │  │ Lifecycle    │            │
│  └──────────────┘  └──────────────┘            │
│                                                  │
│  ┌──────────────┐  ┌──────────────┐            │
│  │ Networking   │  │ Volume Mgmt  │            │
│  └──────────────┘  └──────────────┘            │
└────────────────────┬─────────────────────────────┘
                     │
                     ↓
┌──────────────────────────────────────────────────┐
│              containerd (runtime)                │
│  - Manages container lifecycle                  │
│  - Pulls images                                 │
│  - Low-level container operations               │
└────────────────────┬─────────────────────────────┘
                     │
                     ↓
┌──────────────────────────────────────────────────┐
│                  runc (OCI runtime)              │
│  - Creates namespaces                           │
│  - Sets up cgroups                              │
│  - Executes container process                   │
└──────────────────────────────────────────────────┘
```

**Key components:**
- **Docker CLI** → What you type
- **dockerd** → Docker daemon (background service)
- **containerd** → High-level runtime (pulls images, manages containers)
- **runc** → Low-level runtime (creates namespaces, starts process)

---

## OCI (Open Container Initiative)

**Problem:** Docker was the only game in town (vendor lock-in).

**Solution:** Open Container Initiative standardized:
1. **Image spec** — How to build images
2. **Runtime spec** — How to run containers

**Result:**
- Multiple runtimes: runc, crun, kata (VMs), gVisor (sandboxed)
- Multiple tools: Docker, Podman, containerd, CRI-O
- Interoperability

**Key insight:** Docker is just **one implementation** of OCI standards.

---

## Docker vs Podman

| Feature | Docker | Podman |
|---------|--------|--------|
| Daemon | Yes (dockerd) | No (daemonless) |
| Root required | Yes (by default) | No (rootless mode) |
| CLI compatibility | Docker CLI | Docker-compatible CLI |
| Kubernetes native | No | Yes (generates K8s YAML) |
| Default on Fedora | No | Yes |

**Why Podman exists:**
- No daemon = better security
- Rootless = run as normal user
- Direct systemd integration

**On Fedora:**
```bash
# Podman is preinstalled
podman run hello-world

# Alias for Docker compatibility
alias docker=podman
```

---

## Container Registries

A **registry** stores and distributes images.

```
┌────────────────────────────────────┐
│  Registry (e.g., Docker Hub)       │
│                                    │
│  user/myapp:latest                 │
│  user/myapp:v1.0                   │
│  nginx:alpine                      │
│  postgres:14                       │
└────────────────────────────────────┘
         ↑              ↓
     docker push    docker pull
         ↑              ↓
┌────────────────────────────────────┐
│       Local Docker Daemon          │
└────────────────────────────────────┘
```

**Major registries:**
- **Docker Hub** — Public, free tier limited
- **GitHub Container Registry (ghcr.io)** — Free for public repos
- **AWS ECR** — Private, AWS-integrated
- **Google Artifact Registry** — Private, GCP-integrated
- **Quay.io** — Public/private, by Red Hat

---

## Image Tags

```
nginx:latest
│     │
│     └────── Tag (version)
└──────────── Repository (image name)

Full name:
registry.example.com/user/nginx:v1.0
│                    │    │     │
│                    │    │     └─ Tag
│                    │    └─────── Image name
│                    └──────────── Namespace/user
└───────────────────────────────── Registry
```

**Special tags:**
- `latest` → **NOT** "newest version", just a default tag
- `alpine` → Minimal base (5 MB vs 100 MB)
- `slim` → Smaller than full, bigger than alpine

**Best practice:**
```bash
# ❌ Bad (unpredictable)
docker pull nginx:latest

# ✅ Good (pinned version)
docker pull nginx:1.25.3-alpine
```

---

## Image Digests (The Real Immutable Identifier)

Tags are **mutable** (can be overwritten). Digests are **immutable**.

```bash
# Pull by tag
docker pull nginx:latest

# See digest
docker images --digests
# nginx  latest  sha256:abc123...  5 MB

# Pull by digest (guarantees exact image)
docker pull nginx@sha256:abc123...
```

**Why this matters:**
- Security: Ensure you're running the exact image you audited
- Reproducibility: CI/CD should use digests, not `latest`

---

## War Story: The Disappearing Bug

A team deployed `myapp:latest` to production. Tests passed, everything worked.

Next day, the same deployment **started failing**.

**What happened:**
1. Another team pushed a new `myapp:latest` (with a breaking change)
2. Kubernetes pulled the new `latest` on a different node
3. Same tag, different image

**The fix:**
```yaml
# Before
image: myapp:latest

# After
image: myapp:v1.2.3
# Or even better:
image: myapp@sha256:abc123...
```

**Lesson:** **Never use `latest` in production.**

---

## Key Takeaways

1. **Containers = cgroups + namespaces + layers** — complete mental model
2. **Layers are read-only, shared, and stacked** — efficiency through deduplication
3. **Copy-on-write** — containers modify files by copying to writable layer
4. **Images are templates, containers are instances** — like classes and objects
5. **Docker is just one implementation** — OCI standardizes the ecosystem
6. **Registries store images, digest ensures immutability** — tags can change
7. **Never use `latest` in production** — pin versions or use digests

---

## Exercises

1. **Inspect image layers:**
   ```bash
   docker history nginx:alpine
   # See each layer and its size
   ```

2. **Compare storage usage:**
   ```bash
   docker system df
   # See how much space images/containers/volumes use
   ```

3. **Pull image by digest:**
   ```bash
   docker pull nginx:alpine
   docker images --digests | grep nginx
   # Copy digest, then pull by digest
   ```

4. **Run Podman (Fedora):**
   ```bash
   podman run --rm -it alpine sh
   # Notice no daemon needed
   ```

---

**Next:** [02. Dockerfile Optimization →](./02-dockerfile-optimization.md)
