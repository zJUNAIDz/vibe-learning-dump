# Immutable Artifacts

> **Build once, deploy everywhere. The artifact that passed your tests is the exact artifact that runs in production.**

---

## 🟢 What Is an Immutable Artifact?

### The Problem with Mutable Deployments

**How things used to work (bad):**

```
Build #1 for staging
  → Download dependencies
  → Compile code
  → Deploy to staging
  → Tests pass ✓

Build #2 for production (different build!)
  → Download dependencies    ← Might get different versions!
  → Compile code             ← Might compile differently!
  → Deploy to production
  → Something breaks 🔥
```

**What went wrong?** The staging build and the production build were different. Even though the source code was the same, the build environment changed:

- A dependency released a patch version between builds
- The build server had a different Go/Node version
- An environment variable changed
- Network issues corrupted a download

**You tested Artifact A but deployed Artifact B.**

### The Solution: Build Once, Deploy Many

```
Build #1 → myapp:v1.2.3  (one image, one time)
  ↓           ↓               ↓
Staging    Production    Disaster Recovery
  ✓            ✓               ✓

Same. Exact. Bytes.
```

**An immutable artifact is built once and never modified.** The same artifact moves through environments untouched.

### Mental Model: Factory Sealed Product

Think of a phone from a factory:

```
Factory → Sealed Box → Store A → Customer unpacks
                     → Store B → Customer unpacks
                     → Online  → Customer unpacks
```

The phone is the same regardless of which store sells it. The store doesn't open the box and modify the phone. The artifact (phone) is **immutable**.

---

## 🟢 Docker Images as Artifacts

### Why Docker Images Are Perfect Artifacts

```
Docker Image = Code + Dependencies + Runtime + Configuration Template
```

**Everything is frozen inside the image:**

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download            # Dependencies frozen at build time
COPY . .
RUN CGO_ENABLED=0 go build -o /myapp ./cmd/server  # Binary frozen

FROM alpine:3.19
COPY --from=builder /myapp /myapp
EXPOSE 8080
ENTRYPOINT ["/myapp"]
```

Once built, this image contains:
- Exact Go version used to compile
- Exact dependency versions (from `go.sum`)
- Exact compiled binary
- Exact base OS (Alpine 3.19)

**None of this changes between staging and production.**

### Tagging Images Properly

```bash
# BAD: Mutable tag
docker build -t myapp:latest .     # ← "latest" keeps changing!
docker push myapp:latest

# GOOD: Immutable tags
docker build -t myapp:v1.2.3 .     # Semantic version
docker build -t myapp:abc1234 .     # Git commit SHA
docker push myapp:v1.2.3
docker push myapp:abc1234
```

**Never use `:latest` in production.** It's mutable — it points to whatever was pushed last.

```yaml
# BAD: Kubernetes deployment with mutable tag
spec:
  containers:
    - name: myapp
      image: myapp:latest          # Which "latest"? Nobody knows.

# GOOD: Kubernetes deployment with immutable tag
spec:
  containers:
    - name: myapp
      image: myapp:v1.2.3         # Exact version, auditable
```

### Image Digest (The Ultimate Immutability)

Even tags can be moved. Digests cannot.

```bash
# Push and get the digest
docker push myapp:v1.2.3
# Output: sha256:a3ed95caeb02ffe68cdd9fd8... 

# Use digest in production for maximum safety
spec:
  containers:
    - name: myapp
      image: myapp@sha256:a3ed95caeb02ffe68cdd9fd8...
```

**Digest = content hash of the image layers.** If any bit changes, the digest changes. It's like a fingerprint.

---

## 🟡 Semantic Versioning (SemVer)

### The Format

```
MAJOR.MINOR.PATCH
  v1  .  2  .  3
```

| Component | When to bump | Example |
|-----------|-------------|---------|
| **MAJOR** | Breaking changes | API endpoint removed, schema migration required |
| **MINOR** | New features (backward compatible) | New endpoint added, new optional field |
| **PATCH** | Bug fixes (backward compatible) | Fixed null pointer, corrected calculation |

### Real Examples

```
v1.0.0 → v1.0.1    # Fixed login timeout bug
v1.0.1 → v1.1.0    # Added user avatar feature
v1.1.0 → v1.1.1    # Fixed avatar upload crash
v1.1.1 → v2.0.0    # Rewrote auth system (all tokens invalidated)
```

### Pre-Release Versions

```
v1.2.3-alpha.1      # Early testing, unstable
v1.2.3-beta.1       # Feature complete, may have bugs
v1.2.3-rc.1         # Release candidate, likely final
v1.2.3              # Stable release
```

### Automating Version Bumps

**Using git tags:**

```bash
# Tag the current commit
git tag -a v1.2.3 -m "Release v1.2.3: Fixed login timeout"
git push origin v1.2.3

# Get current version in CI
VERSION=$(git describe --tags --always)
echo $VERSION  # v1.2.3 or v1.2.3-5-gabc1234 (if commits after tag)
```

**Using conventional commits:**

```bash
# Commit messages drive version bumps
git commit -m "fix: resolve login timeout"          # → Patch bump
git commit -m "feat: add user avatar uploads"       # → Minor bump
git commit -m "feat!: rewrite auth system"          # → Major bump (!)

# Tools like semantic-release automate this:
# fix: → v1.2.3 → v1.2.4
# feat: → v1.2.4 → v1.3.0
# feat!: → v1.3.0 → v2.0.0
```

---

## 🟡 The Artifact Lifecycle

### From Code to Production

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  Source   │ ──→ │  Build   │ ──→ │ Registry │ ──→ │  Deploy  │
│   Code   │     │ Artifact │     │  (Store) │     │          │
│          │     │          │     │          │     │          │
│ main.go  │     │ myapp:   │     │ Docker   │     │ Staging  │
│ go.mod   │     │ v1.2.3   │     │ Hub /    │     │ Prod     │
│ ...      │     │          │     │ ECR /    │     │ DR       │
│          │     │          │     │ GCR      │     │          │
└──────────┘     └──────────┘     └──────────┘     └──────────┘
     ↑                                                  │
     │                                                  │
     └──────────── Rollback = Deploy old version ───────┘
```

### Step 1: Build the Artifact

```bash
# In CI (e.g., GitHub Actions)
VERSION=$(git describe --tags --always)
COMMIT=$(git rev-parse --short HEAD)

docker build \
  --build-arg VERSION=$VERSION \
  --build-arg COMMIT=$COMMIT \
  -t myapp:$VERSION \
  -t myapp:$COMMIT \
  .
```

### Step 2: Store in Registry

```bash
# Push to container registry
docker tag myapp:$VERSION registry.example.com/myapp:$VERSION
docker push registry.example.com/myapp:$VERSION
```

**Popular registries:**

| Registry | Best For |
|----------|----------|
| Docker Hub | Open source, public images |
| AWS ECR | AWS workloads |
| Google GCR/Artifact Registry | GCP workloads |
| GitHub Container Registry | GitHub repos |
| Harbor | Self-hosted, enterprise |

### Step 3: Deploy the Same Artifact Everywhere

```bash
# Deploy to staging
kubectl set image deployment/myapp myapp=registry.example.com/myapp:v1.2.3 -n staging

# Run tests against staging
./scripts/smoke-test.sh https://staging.example.com

# Deploy SAME image to production
kubectl set image deployment/myapp myapp=registry.example.com/myapp:v1.2.3 -n production
```

**Same image. Same bytes. Same behavior.**

### Step 4: Rollback = Deploy an Older Version

```bash
# Something went wrong in production?
# Just deploy the previous version:
kubectl set image deployment/myapp myapp=registry.example.com/myapp:v1.2.2 -n production

# v1.2.2 is still in the registry, unchanged
# It's the exact artifact that was running before
# Rollback is instant and safe
```

---

## 🟡 Configuration vs Artifact

### What Goes IN the Artifact

```
✅ Application code (compiled binary or bundled JS)
✅ Dependencies (libraries, packages)
✅ Runtime (Go binary is self-contained, or Node.js + node_modules)
✅ Default configuration templates
✅ Static assets (HTML, CSS, images)
```

### What Stays OUTSIDE the Artifact

```
❌ Database URLs
❌ API keys and secrets
❌ Feature flags
❌ Log levels
❌ Replica counts
❌ Resource limits
❌ Anything environment-specific
```

### How to Inject Configuration

**Environment variables (most common):**

```yaml
# Kubernetes deployment
spec:
  containers:
    - name: myapp
      image: myapp:v1.2.3        # ← Immutable artifact
      env:
        - name: DATABASE_URL      # ← Environment-specific config
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: url
        - name: LOG_LEVEL
          value: "info"            # ← "debug" in staging, "info" in prod
```

```go
// App reads config from environment
func main() {
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        log.Fatal("DATABASE_URL not set")
    }

    logLevel := os.Getenv("LOG_LEVEL")
    if logLevel == "" {
        logLevel = "info"  // sensible default
    }
}
```

**ConfigMaps and Secrets in Kubernetes:**

```yaml
# Different config per environment, same image
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-config
  namespace: staging
data:
  LOG_LEVEL: "debug"
  CACHE_TTL: "60"
  FEATURE_NEW_UI: "true"

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-config
  namespace: production
data:
  LOG_LEVEL: "warn"
  CACHE_TTL: "3600"
  FEATURE_NEW_UI: "false"     # Not yet in production
```

---

## 🔴 War Story: The Artifact That Wasn't Immutable

**What happened:** A team built their app for staging, ran tests, all passed. Then they ran another build for production.

```
Monday:    Build for staging → npm ci → lodash@4.17.20 → Tests pass ✓
Wednesday: Build for production → npm ci → lodash@4.17.21 → Something breaks 🔥
```

**Between Monday and Wednesday**, lodash released 4.17.21 which had a subtle behavior change. The staging build used 4.17.20, the production build got 4.17.21.

**They tested one artifact but deployed a different one.**

**Fix:**

```bash
# Build ONCE
docker build -t myapp:abc1234 .

# Use that SAME image everywhere
docker push registry.example.com/myapp:abc1234

# Staging
kubectl set image deployment/myapp myapp=registry.example.com/myapp:abc1234 -n staging
# Test ✓

# Production (SAME image!)
kubectl set image deployment/myapp myapp=registry.example.com/myapp:abc1234 -n production
```

**Now they'd never see this problem again.** The image that passed tests is exactly what runs in production.

---

## 🟡 Artifact Retention and Cleanup

### How Long to Keep Artifacts

```bash
# Retention policy example:
# - Keep last 10 versions always
# - Keep last 30 days of images
# - Keep any version that's currently deployed
# - Delete everything else

# AWS ECR lifecycle policy
aws ecr put-lifecycle-policy \
  --repository-name myapp \
  --lifecycle-policy-text '{
    "rules": [
      {
        "rulePriority": 1,
        "description": "Keep last 10 images",
        "selection": {
          "tagStatus": "tagged",
          "tagPrefixList": ["v"],
          "countType": "imageCountMoreThan",
          "countNumber": 10
        },
        "action": { "type": "expire" }
      },
      {
        "rulePriority": 2,
        "description": "Delete untagged images older than 7 days",
        "selection": {
          "tagStatus": "untagged",
          "countType": "sinceImagePushed",
          "countUnit": "days",
          "countNumber": 7
        },
        "action": { "type": "expire" }
      }
    ]
  }'
```

### Why Retention Matters

```
Without cleanup:
  Week 1:  10 images  ×  500 MB = 5 GB
  Month 1: 40 images  ×  500 MB = 20 GB
  Year 1:  480 images ×  500 MB = 240 GB   ← $$$
```

---

## ✅ Hands-On Exercise

### Build an Immutable Artifact Pipeline (Local)

**1. Create an app with version metadata:**

```bash
mkdir -p ~/artifact-demo && cd ~/artifact-demo

cat > main.go << 'EOF'
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

var (
    version   = "dev"
    commit    = "none"
    buildTime = "unknown"
)

type BuildInfo struct {
    Version   string `json:"version"`
    Commit    string `json:"commit"`
    BuildTime string `json:"buildTime"`
    Env       string `json:"environment"`
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "development"
    }
    json.NewEncoder(w).Encode(BuildInfo{
        Version:   version,
        Commit:    commit,
        BuildTime: buildTime,
        Env:       env,
    })
}

func main() {
    http.HandleFunc("/info", infoHandler)
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    fmt.Printf("Starting %s (commit: %s, built: %s)\n", version, commit, buildTime)
    http.ListenAndServe(":"+port, nil)
}
EOF

go mod init artifact-demo
```

**2. Build with version metadata:**

```bash
VERSION="v1.0.0"
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "local")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

go build \
  -ldflags="-X main.version=$VERSION -X main.commit=$COMMIT -X main.buildTime=$BUILD_TIME" \
  -o myapp .

echo "Built artifact: myapp (version: $VERSION, commit: $COMMIT)"
```

**3. Deploy the SAME binary to different "environments":**

```bash
# "Staging"
APP_ENV=staging PORT=8081 ./myapp &
STAGING_PID=$!
sleep 1
echo "Staging:"
curl -s http://localhost:8081/info | jq .

# "Production"
APP_ENV=production PORT=8082 ./myapp &
PROD_PID=$!
sleep 1
echo "Production:"
curl -s http://localhost:8082/info | jq .

# Notice: version, commit, buildTime are identical
# Only "environment" changes (from APP_ENV)

kill $STAGING_PID $PROD_PID
```

**4. Verify immutability:**

```bash
# Check the binary hash
sha256sum myapp
# abc123...  myapp

# Same artifact, same hash, every time
# If you rebuild without changing code, you get a different hash
# (because buildTime changes) — that's a new artifact
```

---

## 📚 Summary

| Concept | What It Means |
|---------|---------------|
| **Immutable Artifact** | Built once, never modified, deployed everywhere |
| **Mutable Tag** | `:latest` — changes over time, unpredictable |
| **Immutable Tag** | `:v1.2.3` or `:abc1234` — always points to the same image |
| **Digest** | `@sha256:...` — cryptographic proof of exact content |
| **SemVer** | `MAJOR.MINOR.PATCH` — communicates change impact |
| **Config outside artifact** | Env vars, ConfigMaps, Secrets — not baked into image |
| **Artifact Registry** | Storage for versioned, immutable images (ECR, GCR, Docker Hub) |

**The golden rule:** The artifact that passed your tests is the exact artifact that runs in production. If you rebuild, it's a new artifact and needs new tests.

---

**Previous:** [02. Build vs Test vs Deploy](./02-build-test-deploy.md)  
**Next:** [04. Pipeline Stages](./04-pipeline-stages.md)  
**Module:** [06. CI/CD Fundamentals](./README.md)
