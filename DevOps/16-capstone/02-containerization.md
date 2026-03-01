# Capstone Phase 02: Containerization

> **Take your application and package it in a Docker image that's small, secure, and production-ready. Not just "it works in Docker" — actually good.**

---

## 🟢 Multi-Stage Dockerfile (TypeScript)

```dockerfile
# ──────────────────────────────────────────────
# Stage 1: Install dependencies
# ──────────────────────────────────────────────
FROM node:20-slim AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --only=production && \
    cp -R node_modules /prod_modules && \
    npm ci
# /prod_modules has production deps only
# node_modules has ALL deps (including devDependencies for build)

# ──────────────────────────────────────────────
# Stage 2: Build TypeScript
# ──────────────────────────────────────────────
FROM node:20-slim AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build
# Output: /app/dist/

# ──────────────────────────────────────────────
# Stage 3: Production image
# ──────────────────────────────────────────────
FROM node:20-slim AS production
WORKDIR /app

# Create non-root user
RUN groupadd -r appuser && useradd -r -g appuser -s /bin/false appuser

# Copy only what's needed
COPY --from=deps --chown=appuser:appuser /prod_modules ./node_modules
COPY --from=builder --chown=appuser:appuser /app/dist ./dist
COPY --chown=appuser:appuser package.json ./

# Security: non-root user
USER appuser

# Security: read-only filesystem (app writes to /tmp only)
# Set in K8s securityContext, not here

EXPOSE 3000

# Use node directly, not npm (npm adds unnecessary process)
CMD ["node", "dist/server.js"]
```

### What This Achieves

```
Stage 1 (deps):     Installs all dependencies
Stage 2 (builder):  Compiles TypeScript to JavaScript
Stage 3 (final):    Only compiled JS + production deps

What's NOT in the final image:
  → TypeScript source code
  → TypeScript compiler
  → devDependencies (jest, eslint, etc.)
  → package-lock.json
  → tests/
  → .git/
  → node_modules dev packages

Result:
  node:20           → ~1.1 GB
  node:20-slim      → ~240 MB
  Our final image   → ~80 MB ← (slim base + minimal deps)
```

---

## 🟢 Multi-Stage Dockerfile (Go)

```dockerfile
# ──────────────────────────────────────────────
# Stage 1: Build
# ──────────────────────────────────────────────
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags='-w -s' -o /server ./cmd/server

# ──────────────────────────────────────────────
# Stage 2: Production
# ──────────────────────────────────────────────
FROM scratch
COPY --from=builder /server /server
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 65534:65534
EXPOSE 3000
ENTRYPOINT ["/server"]

# Final image size: ~10-15 MB (just the binary + CA certs)
```

---

## 🟢 .dockerignore

```dockerignore
# .dockerignore — CRITICAL for build speed and security
node_modules
npm-debug.log
.git
.gitignore
.env
.env.*
*.md
tests/
coverage/
.github/
k8s/
terraform/
docker-compose*.yml
Makefile
.vscode/
.idea/
```

---

## 🟡 Image Optimization

### Layer Caching

```dockerfile
# BAD — cache busted on every code change
COPY . .
RUN npm ci

# GOOD — dependencies cached until package.json changes
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
# npm ci only runs when package*.json changes
# Source code changes DON'T reinstall dependencies
```

### Measuring Image Size

```bash
# Check image size
docker images task-service
# REPOSITORY    TAG      SIZE
# task-service  latest   78MB

# See what's in each layer
docker history task-service:latest
# IMAGE         CREATED BY                      SIZE
# abc123        CMD ["node" "dist/server.js"]    0B
# def456        COPY --from=builder ...          2.3MB
# ghi789        COPY --from=deps ...             45MB
# jkl012        RUN groupadd ...                 1.2KB

# Deep dive with dive tool
brew install dive
dive task-service:latest
# Interactive layer-by-layer exploration
# Shows wasted space, efficiency score
```

---

## 🟡 Security Scanning

```bash
# Scan with Trivy before pushing
trivy image task-service:latest

# Fail on critical/high vulnerabilities
trivy image --severity HIGH,CRITICAL --exit-code 1 task-service:latest

# Scan for secrets accidentally baked in
trivy image --scanners secret task-service:latest

# Example CI step
# Build → Scan → Push (only if clean)
docker build -t task-service:$SHA .
trivy image --exit-code 1 --severity HIGH,CRITICAL task-service:$SHA
docker push task-service:$SHA
```

---

## 🟡 Docker Compose for Local Development

```yaml
# docker-compose.yaml
version: '3.8'

services:
  app:
    build:
      context: .
      target: production  # Use production stage
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
      - PORT=3000
      - LOG_LEVEL=info
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:3000/health/live"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 10s

  # Local development with hot reload
  app-dev:
    build:
      context: .
      target: deps  # Use deps stage (has all devDependencies)
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=development
      - LOG_LEVEL=debug
    volumes:
      - ./src:/app/src  # Hot reload
    command: npx tsx watch src/server.ts
```

---

## 🔴 Complete Build and Verify

```bash
# Full build, scan, and test workflow

# 1. Build
docker build -t task-service:local .

# 2. Check size
docker images task-service:local --format "{{.Size}}"
# Should be < 100MB

# 3. Verify non-root
docker run --rm task-service:local id
# uid=1000(appuser) gid=1000(appuser)

# 4. Scan for vulnerabilities
trivy image --severity HIGH,CRITICAL task-service:local

# 5. Test the image
docker run -d --name test-app -p 3000:3000 task-service:local
sleep 3

# Health check
curl http://localhost:3000/health/live
# {"status":"alive"}

# Metrics
curl http://localhost:3000/metrics | head -5
# # HELP http_requests_total Total HTTP requests

# API test
curl -X POST http://localhost:3000/api/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Test from Docker"}'
# {"id":"...","title":"Test from Docker","completed":false}

# Cleanup
docker stop test-app && docker rm test-app
```

---

## 🔴 Checklist

```
□ Multi-stage build (builder + production stages)
□ Final image < 100MB
□ Running as non-root user (USER directive)
□ .dockerignore excludes sensitive files
□ npm ci used (not npm install)
□ node used directly (not npm start)
□ No secrets in image layers
□ Trivy scan passes (no HIGH/CRITICAL)
□ Health check endpoint works in container
□ Graceful shutdown handles SIGTERM
```

---

**Previous:** [01. Application](./01-application.md)  
**Next:** [03. Kubernetes Manifests](./03-kubernetes-manifests.md)
