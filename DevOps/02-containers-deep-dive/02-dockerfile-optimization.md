# Dockerfile Optimization

🟡 **Intermediate** → 🔴 **Advanced**

---

## Why Optimization Matters

**Bad Dockerfile:**
- 2 GB image size
- 10 minute builds
- Rebuilds everything on tiny code change
- Slow CI/CD pipeline

**Good Dockerfile:**
- 50 MB image size
- 30 second builds (with cache)
- Only rebuilds what changed
- Fast deployments

---

## Dockerfile Basics (Quick Refresh)

```dockerfile
# Base image
FROM node:18-alpine

# Working directory
WORKDIR /app

# Copy files
COPY package*.json ./
RUN npm install

COPY . .

# Command to run
CMD ["node", "server.js"]
```

**Each instruction creates a layer.**

---

## Optimization #1: Layer Caching

Docker caches each layer. If a layer hasn't changed, Docker reuses it.

### Bad (Cache-Busting)

```dockerfile
FROM node:18-alpine
WORKDIR /app

# ❌ Copies everything first
COPY . .

# Every code change invalidates cache here
RUN npm install

CMD ["node", "server.js"]
```

**Problem:** Every code change → re-runs `npm install` (slow!)

---

### Good (Cache-Friendly)

```dockerfile
FROM node:18-alpine
WORKDIR /app

# ✅ Copy dependency files first
COPY package*.json ./
RUN npm install

# Copy code last
COPY . .

CMD ["node", "server.js"]
```

**Why this works:**
- `package.json` rarely changes → `npm install` layer cached
- Code changes often → only last `COPY` layer rebuilds

---

## Optimization #2: Multi-Stage Builds

**Problem:** Build tools bloat the image.

### Bad (Single Stage)

```dockerfile
FROM node:18
WORKDIR /app

COPY package*.json ./
RUN npm install  # Includes devDependencies

COPY . .
RUN npm run build  # TypeScript compilation

CMD ["node", "dist/server.js"]

# Final image: ~1 GB (includes TypeScript, build tools, etc.)
```

---

### Good (Multi-Stage)

```dockerfile
# Stage 1: Build
FROM node:18 AS builder
WORKDIR /app

COPY package*.json ./
RUN npm install  # All dependencies

COPY . .
RUN npm run build

# Stage 2: Production
FROM node:18-alpine
WORKDIR /app

COPY package*.json ./
RUN npm install --production  # Only production deps

COPY --from=builder /app/dist ./dist

CMD ["node", "dist/server.js"]

# Final image: ~150 MB (no dev dependencies, no source code)
```

**Key insight:** Only the **last stage** ends up in the final image.

---

## Multi-Stage Build for Go

```dockerfile
# Stage 1: Build
FROM golang:1.21 AS builder
WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp

# Stage 2: Production
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/myapp .

CMD ["./myapp"]

# Final image: ~10 MB (Go binary is static!)
```

---

## Optimization #3: Minimize Layers

Each `RUN`, `COPY`, `ADD` creates a layer.

### Bad (Many Layers)

```dockerfile
RUN apt-get update
RUN apt-get install -y curl
RUN apt-get install -y git
RUN apt-get install -y vim
```

**Problem:** 4 layers, larger image size.

---

### Good (Combined Layers)

```dockerfile
RUN apt-get update && apt-get install -y \
    curl \
    git \
    vim \
    && rm -rf /var/lib/apt/lists/*
```

**Why this is better:**
- 1 layer instead of 4
- Cleans up apt cache in same layer (smaller size)

---

## Optimization #4: Use `.dockerignore`

Like `.gitignore`, but for Docker.

**`.dockerignore`:**
```
node_modules/
dist/
.git/
.env
*.log
*.md
Dockerfile
.dockerignore
```

**Why:**
- Speeds up `COPY . .` (less data transferred)
- Smaller context size
- Avoids accidentally copying secrets

---

## Optimization #5: Order Instructions by Change Frequency

```dockerfile
# Least likely to change → Most likely to change

FROM node:18-alpine                  # Rarely changes
RUN apk add --no-cache python3       # Rarely changes
WORKDIR /app                         # Never changes
COPY package*.json ./                # Changes sometimes
RUN npm install                      # Changes when deps change
COPY . .                             # Changes often (code)
CMD ["node", "server.js"]            # Rarely changes
```

---

## Optimization #6: Use Specific Base Images

| Base | Size | Use Case |
|------|------|----------|
| `node:18` | ~1 GB | Full tooling (debugging, dev) |
| `node:18-slim` | ~200 MB | Smaller, still has package manager |
| `node:18-alpine` | ~150 MB | Minimal (based on Alpine Linux) |
| `scratch` | 0 MB | Empty (only for static binaries) |

### Alpine Linux

**Pros:**
- Very small (~5 MB base)
- Security-focused

**Cons:**
- Uses `musl libc` instead of `glibc` (can cause compatibility issues)
- Fewer packages available

**Best practice:** Use `alpine` for production, full image for debugging.

---

## Optimization #7: Static Binaries (Go)

Go can compile **static binaries** (no dependencies).

```dockerfile
FROM scratch
COPY myapp /
CMD ["/myapp"]

# Final image: 5-10 MB (just the binary!)
```

**Why this works:**
- Go binary is self-contained
- `scratch` is an empty image (0 MB)

**Caveat:** No shell, no utilities. Hard to debug.

**Compromise:**
```dockerfile
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY myapp /
CMD ["/myapp"]

# Final image: ~15 MB (alpine + binary)
```

---

## Optimization #8: Build Arguments

Pass values at build time.

```dockerfile
ARG NODE_VERSION=18
FROM node:${NODE_VERSION}-alpine

ARG BUILD_ENV=production
ENV NODE_ENV=${BUILD_ENV}

WORKDIR /app
COPY . .
RUN npm install

CMD ["node", "server.js"]
```

**Usage:**
```bash
# Default
docker build -t myapp .

# Custom
docker build --build-arg NODE_VERSION=20 --build-arg BUILD_ENV=development -t myapp .
```

---

## Optimization #9: BuildKit (Faster Builds)

**BuildKit** is Docker's modern build engine (enabled by default in recent versions).

**Features:**
- Parallel builds
- Better caching
- Secret mounts (never in image)

**Enable:**
```bash
export DOCKER_BUILDKIT=1
docker build -t myapp .
```

**Secrets (never persisted in layers):**
```dockerfile
# Mount secret at build time (not in final image)
RUN --mount=type=secret,id=npm_token \
    echo "//registry.npmjs.org/:_authToken=$(cat /run/secrets/npm_token)" > ~/.npmrc && \
    npm install && \
    rm ~/.npmrc
```

**Build:**
```bash
docker build --secret id=npm_token,src=~/.npmrc -t myapp .
```

---

## TypeScript App Example (Full Optimization)

```dockerfile
# syntax=docker/dockerfile:1

# Stage 1: Dependencies
FROM node:18-alpine AS deps
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

# Stage 2: Build
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Stage 3: Production
FROM node:18-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
USER node

COPY --from=deps --chown=node:node /app/node_modules ./node_modules
COPY --from=builder --chown=node:node /app/dist ./dist
COPY --chown=node:node package*.json ./

EXPOSE 8080
CMD ["node", "dist/server.js"]
```

**Key techniques:**
- ✅ Multi-stage build
- ✅ Separate dependencies stage
- ✅ Production-only dependencies
- ✅ Alpine base
- ✅ Non-root user
- ✅ Layer caching optimized

---

## Go App Example (Full Optimization)

```dockerfile
# syntax=docker/dockerfile:1

# Stage 1: Build
FROM golang:1.21-alpine AS builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Download dependencies first (caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o myapp .

# Stage 2: Production
FROM scratch
WORKDIR /

# Copy timezone data, CA certs from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy binary
COPY --from=builder /app/myapp /myapp

# Run as non-root
USER nobody:nobody

EXPOSE 8080
ENTRYPOINT ["/myapp"]
```

**Key techniques:**
- ✅ Multi-stage build
- ✅ Static binary
- ✅ `scratch` base (minimal)
- ✅ Non-root user
- ✅ Strip debug symbols (`-ldflags="-s -w"`)

---

## Optimization #10: Health Checks

Add health checks to the Dockerfile.

```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY . .
RUN npm install

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD node healthcheck.js || exit 1

CMD ["node", "server.js"]
```

**Why:**
- Docker/Kubernetes can detect unhealthy containers
- Automatic restarts

---

## Security Best Practices

### 1. Don't Run as Root

```dockerfile
# Bad
FROM node:18-alpine
WORKDIR /app
COPY . .
CMD ["node", "server.js"]
# Runs as root (UID 0) by default

# Good
FROM node:18-alpine
WORKDIR /app
COPY . .
USER node  # Run as non-root user
CMD ["node", "server.js"]
```

---

### 2. Don't Embed Secrets

```dockerfile
# ❌ BAD (secret in layer forever)
RUN echo "db_password=secret123" > /app/config

# ✅ GOOD (use env vars or mounted secrets)
# Pass at runtime:
docker run -e DB_PASSWORD=secret123 myapp
```

---

### 3. Scan Images for Vulnerabilities

```bash
# Docker Scout (built-in)
docker scout cves myapp:latest

# Trivy (popular)
trivy image myapp:latest

# Grype
grype myapp:latest
```

---

## War Story: The 5 GB Image

A team's Node.js image was **5 GB**. Deploys took 20 minutes.

**Investigation:**
```bash
docker history myapp:latest
```

**Findings:**
1. Used `node:18` (1 GB) instead of `node:18-alpine` (150 MB)
2. Installed `npm install` (including devDependencies)
3. Copied entire repo (including `.git`, `node_modules`)
4. Didn't use multi-stage build

**Fixes:**
- Switched to `node:18-alpine`
- Multi-stage build (builder + runtime)
- Added `.dockerignore`
- Used `npm ci --only=production`

**Result:** 5 GB → **80 MB** (62x smaller!)

---

## Key Takeaways

1. **Layer caching is critical** — order instructions by change frequency
2. **Multi-stage builds** — separate build from runtime
3. **Alpine base images** — small, secure, but watch for libc compatibility
4. **Go static binaries** — can run from `scratch` (~10 MB)
5. **`.dockerignore`** — prevent bloat, speed up builds
6. **Never use `latest`** — pin versions
7. **Never run as root** — use `USER` directive
8. **Scan for vulnerabilities** — use Trivy, Docker Scout, etc.
9. **BuildKit secrets** — keep secrets out of layers

---

## Exercises

1. **Optimize a TypeScript app:**
   - Start with a basic Dockerfile
   - Add multi-stage build
   - Switch to alpine
   - Compare sizes before/after

2. **Compare build times:**
   - Time a build with poor caching
   - Time a build with optimized layer order
   - Observe cache hits

3. **Go static binary:**
   - Build a Go app with `scratch` base
   - Verify it runs
   - Try to shell into it (spoiler: you can't)

4. **Scan an image:**
   ```bash
   trivy image node:18
   trivy image node:18-alpine
   # Compare vulnerabilities
   ```

---

**Next:** [03. Container Security →](./03-container-security.md)
