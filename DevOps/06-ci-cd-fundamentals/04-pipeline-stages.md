# Pipeline Stages

> **A CI/CD pipeline is a series of automated stages that transform code into a running application. Each stage has a job, and the pipeline stops at the first failure.**

---

## 🟢 Anatomy of a Pipeline

### The Stages

```
┌─────────┐   ┌──────────┐   ┌───────┐   ┌───────┐   ┌──────┐   ┌────────┐
│Checkout │ → │  Deps    │ → │ Lint  │ → │ Test  │ → │Build │ → │ Push   │
│  Code   │   │ Install  │   │       │   │       │   │Image │   │Registry│
└─────────┘   └──────────┘   └───────┘   └───────┘   └──────┘   └────────┘
                                                                      │
                                        ┌──────────┐   ┌──────────┐   │
                                        │  Smoke   │ ← │ Deploy   │ ←─┘
                                        │  Tests   │   │ Staging  │
                                        └──────────┘   └──────────┘
                                              │
                                        ┌──────────┐   ┌──────────┐
                                        │  Smoke   │ ← │ Deploy   │
                                        │  Tests   │   │   Prod   │
                                        └──────────┘   └──────────┘
```

**If any stage fails, the entire pipeline stops.** No broken code reaches production.

---

## 🟢 Stage 1: Checkout Code

### What Happens

The CI server clones your repository at the exact commit that triggered the pipeline.

**GitHub Actions:**

```yaml
steps:
  - name: Checkout code
    uses: actions/checkout@v4
    with:
      fetch-depth: 0    # Full history (needed for git describe, tags)
```

**Why `fetch-depth: 0`:**

```bash
# Shallow clone (default: fetch-depth: 1)
git describe --tags    # fatal: No names found

# Full clone (fetch-depth: 0)
git describe --tags    # v1.2.3-5-gabc1234  ← Useful!
```

### What Can Go Wrong

| Problem | Cause | Fix |
|---------|-------|-----|
| Checkout times out | Large repo with full history | Use `fetch-depth: 1` (shallow clone) |
| Submodules missing | Repo has git submodules | Add `submodules: recursive` |
| Wrong branch | PR builds checkout merge commit | Usually correct by default |

---

## 🟢 Stage 2: Install Dependencies

### Go

```yaml
- name: Setup Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.22'
    cache: true           # Cache ~/go/pkg/mod

- name: Download dependencies
  run: go mod download

- name: Verify dependencies
  run: go mod verify      # Check nothing was tampered with
```

### Node.js / TypeScript

```yaml
- name: Setup Node.js
  uses: actions/setup-node@v4
  with:
    node-version: '20'
    cache: 'npm'          # Cache ~/.npm

- name: Install dependencies
  run: npm ci             # Deterministic! Not npm install.
```

### Caching Dependencies

**Without cache:**
```
Install dependencies: 2 minutes (downloads everything from internet)
```

**With cache:**
```
Install dependencies: 10 seconds (restores from cache)
```

**GitHub Actions caching:**

```yaml
- name: Cache Go modules
  uses: actions/cache@v4
  with:
    path: ~/go/pkg/mod
    key: go-${{ hashFiles('go.sum') }}
    restore-keys: go-

- name: Cache node_modules
  uses: actions/cache@v4
  with:
    path: ~/.npm
    key: npm-${{ hashFiles('package-lock.json') }}
    restore-keys: npm-
```

**Cache invalidation:** When `go.sum` or `package-lock.json` changes, a new cache is created. The old cache is still used as a fallback (`restore-keys`).

---

## 🟢 Stage 3: Lint and Static Analysis

### Why Lint Before Testing

```
Linting: 10 seconds
Testing: 3 minutes

If code has a style issue:
  Without lint stage: Wait 3 minutes for tests → then notice lint error
  With lint stage: Fail in 10 seconds → fix immediately
```

**Fail fast. Run the cheapest checks first.**

### Go Linting

```yaml
- name: Run linter
  uses: golangci/golangci-lint-action@v4
  with:
    version: 'v1.57'
    args: --timeout=5m

- name: Check formatting
  run: |
    if [ -n "$(gofmt -l .)" ]; then
      echo "Code is not formatted. Run 'gofmt -w .'"
      gofmt -l .
      exit 1
    fi
```

### TypeScript Linting

```yaml
- name: Lint
  run: npx eslint . --ext .ts,.tsx --max-warnings 0

- name: Check formatting
  run: npx prettier --check .

- name: Type check
  run: npx tsc --noEmit
```

### Security Scanning

```yaml
- name: Check for vulnerabilities
  run: |
    # Go
    go list -json -deps ./... | nancy sleuth
    
    # Node.js
    npm audit --audit-level=high
```

---

## 🟢 Stage 4: Run Tests

### Ordering Tests by Speed

```yaml
# Run fast tests first
- name: Unit tests
  run: go test ./internal/... -race -count=1 -timeout 2m

# Then slower tests
- name: Integration tests
  run: go test ./tests/integration/... -race -count=1 -timeout 5m
  env:
    DATABASE_URL: postgres://postgres:testpass@localhost:5432/testdb

# Then slowest tests
- name: E2E tests
  run: npx playwright test
  if: github.ref == 'refs/heads/main'    # Only on main branch
```

### Running Tests with Databases

**GitHub Actions services (sidecars):**

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: testdb
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: testpass
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4
      
      - name: Run tests
        run: go test ./... -race -count=1
        env:
          DATABASE_URL: postgres://postgres:testpass@localhost:5432/testdb?sslmode=disable
          REDIS_URL: redis://localhost:6379
```

### Test Artifacts

```yaml
- name: Run tests with coverage
  run: go test ./... -coverprofile=coverage.out

- name: Upload coverage
  uses: actions/upload-artifact@v4
  with:
    name: coverage-report
    path: coverage.out

# Or send to Codecov
- name: Upload to Codecov
  uses: codecov/codecov-action@v4
  with:
    file: coverage.out
```

### Parallel Test Execution

```yaml
# Run different test suites in parallel jobs
jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - run: go test ./internal/...

  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres: ...
    steps:
      - run: go test ./tests/integration/...

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - run: npx playwright test

  # Only proceed if ALL test jobs pass
  deploy:
    needs: [unit-tests, integration-tests, e2e-tests]
    runs-on: ubuntu-latest
    steps: ...
```

---

## 🟡 Stage 5: Build Docker Image

### Building With Metadata

```yaml
- name: Set build variables
  id: vars
  run: |
    echo "VERSION=$(git describe --tags --always)" >> $GITHUB_OUTPUT
    echo "COMMIT=$(git rev-parse --short HEAD)" >> $GITHUB_OUTPUT
    echo "BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")" >> $GITHUB_OUTPUT

- name: Build Docker image
  run: |
    docker build \
      --build-arg VERSION=${{ steps.vars.outputs.VERSION }} \
      --build-arg COMMIT=${{ steps.vars.outputs.COMMIT }} \
      --build-arg BUILD_TIME=${{ steps.vars.outputs.BUILD_TIME }} \
      -t myapp:${{ steps.vars.outputs.VERSION }} \
      -t myapp:${{ steps.vars.outputs.COMMIT }} \
      .
```

### Dockerfile for the Build

```dockerfile
FROM golang:1.22-alpine AS builder

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_TIME=unknown

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o /myapp ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder /myapp /myapp
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://localhost:8080/health || exit 1
ENTRYPOINT ["/myapp"]
```

### Docker Layer Caching in CI

```yaml
- name: Set up Docker Buildx
  uses: docker/setup-buildx-action@v3

- name: Build with cache
  uses: docker/build-push-action@v5
  with:
    context: .
    push: false
    tags: myapp:${{ steps.vars.outputs.VERSION }}
    cache-from: type=gha        # GitHub Actions cache
    cache-to: type=gha,mode=max
```

---

## 🟡 Stage 6: Push to Registry

### Docker Hub

```yaml
- name: Login to Docker Hub
  uses: docker/login-action@v3
  with:
    username: ${{ secrets.DOCKERHUB_USERNAME }}
    password: ${{ secrets.DOCKERHUB_TOKEN }}

- name: Push image
  run: |
    docker tag myapp:${{ steps.vars.outputs.VERSION }} \
      username/myapp:${{ steps.vars.outputs.VERSION }}
    docker push username/myapp:${{ steps.vars.outputs.VERSION }}
```

### AWS ECR

```yaml
- name: Configure AWS credentials
  uses: aws-actions/configure-aws-credentials@v4
  with:
    aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
    aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
    aws-region: us-east-1

- name: Login to ECR
  id: ecr-login
  uses: aws-actions/amazon-ecr-login@v2

- name: Push to ECR
  run: |
    ECR_REGISTRY=${{ steps.ecr-login.outputs.registry }}
    docker tag myapp:${{ steps.vars.outputs.VERSION }} \
      $ECR_REGISTRY/myapp:${{ steps.vars.outputs.VERSION }}
    docker push $ECR_REGISTRY/myapp:${{ steps.vars.outputs.VERSION }}
```

### GitHub Container Registry

```yaml
- name: Login to GHCR
  uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}

- name: Push to GHCR
  run: |
    docker tag myapp:${{ steps.vars.outputs.VERSION }} \
      ghcr.io/${{ github.repository }}/myapp:${{ steps.vars.outputs.VERSION }}
    docker push ghcr.io/${{ github.repository }}/myapp:${{ steps.vars.outputs.VERSION }}
```

---

## 🟡 Stage 7: Deploy to Staging

### Kubernetes Deployment

```yaml
- name: Deploy to staging
  run: |
    # Update kubeconfig
    aws eks update-kubeconfig --name my-cluster --region us-east-1

    # Set the new image
    kubectl set image deployment/myapp \
      myapp=$ECR_REGISTRY/myapp:${{ steps.vars.outputs.VERSION }} \
      -n staging

    # Wait for rollout to complete
    kubectl rollout status deployment/myapp -n staging --timeout=300s
```

### Helm Deployment

```yaml
- name: Deploy with Helm
  run: |
    helm upgrade --install myapp ./chart \
      --namespace staging \
      --set image.tag=${{ steps.vars.outputs.VERSION }} \
      --set image.repository=$ECR_REGISTRY/myapp \
      --wait \
      --timeout 5m
```

---

## 🟡 Stage 8: Smoke Tests

### Post-Deployment Verification

```yaml
- name: Run smoke tests
  run: |
    STAGING_URL="https://staging.example.com"
    
    echo "Waiting for deployment to be ready..."
    for i in $(seq 1 30); do
      STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$STAGING_URL/health" || echo "000")
      if [ "$STATUS" = "200" ]; then
        echo "Health check passed!"
        break
      fi
      echo "Attempt $i: Status $STATUS, retrying..."
      sleep 5
    done

    if [ "$STATUS" != "200" ]; then
      echo "FAIL: Health check failed after 30 attempts"
      exit 1
    fi

    # Verify version
    DEPLOYED_VERSION=$(curl -s "$STAGING_URL/health" | jq -r '.version')
    EXPECTED_VERSION="${{ steps.vars.outputs.VERSION }}"
    
    if [ "$DEPLOYED_VERSION" != "$EXPECTED_VERSION" ]; then
      echo "FAIL: Expected version $EXPECTED_VERSION, got $DEPLOYED_VERSION"
      exit 1
    fi
    
    echo "Deployed version confirmed: $DEPLOYED_VERSION"
```

### What Smoke Tests Should Check

```bash
# 1. App is running
curl -f https://staging.example.com/health

# 2. Correct version deployed
curl -s https://staging.example.com/health | jq '.version'

# 3. Database connection works
curl -f https://staging.example.com/api/v1/status

# 4. Critical endpoints respond
curl -f https://staging.example.com/api/v1/users?limit=1
curl -f https://staging.example.com/api/v1/products?limit=1

# 5. External dependencies reachable
curl -f https://staging.example.com/api/v1/health/dependencies
```

---

## 🟡 Stage 9: Deploy to Production

### Manual Approval Gate

```yaml
deploy-production:
  needs: [smoke-tests-staging]
  environment:
    name: production
    url: https://api.example.com
  runs-on: ubuntu-latest
  steps:
    - name: Deploy to production
      run: |
        kubectl set image deployment/myapp \
          myapp=$ECR_REGISTRY/myapp:${{ steps.vars.outputs.VERSION }} \
          -n production
        kubectl rollout status deployment/myapp -n production --timeout=300s

    - name: Verify production
      run: ./scripts/smoke-test.sh https://api.example.com
```

**The `environment: production` setting in GitHub Actions creates an approval gate.** Someone must click "Approve" before this job runs.

---

## 🔴 Complete Pipeline: GitHub Actions

### Full Working Example

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}/myapp

jobs:
  # ── Stage 1-3: Lint ──────────────────────
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - uses: golangci/golangci-lint-action@v4
      - run: |
          if [ -n "$(gofmt -l .)" ]; then
            echo "Run gofmt"; exit 1
          fi

  # ── Stage 4: Test ────────────────────────
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: testdb
          POSTGRES_PASSWORD: testpass
        ports: ['5432:5432']
        options: --health-cmd pg_isready --health-interval 10s --health-timeout 5s --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go test ./... -race -count=1 -coverprofile=coverage.out
        env:
          DATABASE_URL: postgres://postgres:testpass@localhost:5432/testdb?sslmode=disable
      - uses: actions/upload-artifact@v4
        with:
          name: coverage
          path: coverage.out

  # ── Stage 5-6: Build & Push ──────────────
  build-push:
    needs: [lint, test]
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    outputs:
      version: ${{ steps.vars.outputs.VERSION }}
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - id: vars
        run: |
          echo "VERSION=$(git describe --tags --always)" >> $GITHUB_OUTPUT

      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - uses: docker/setup-buildx-action@v3

      - uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: |
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ steps.vars.outputs.VERSION }}
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max

  # ── Stage 7-8: Deploy Staging ────────────
  deploy-staging:
    needs: [build-push]
    runs-on: ubuntu-latest
    environment:
      name: staging
      url: https://staging.example.com
    steps:
      - uses: actions/checkout@v4

      - name: Deploy to staging
        run: |
          echo "Deploying ${{ needs.build-push.outputs.version }} to staging"
          # kubectl set image deployment/myapp ...

      - name: Smoke test staging
        run: |
          echo "Running smoke tests against staging"
          # ./scripts/smoke-test.sh https://staging.example.com

  # ── Stage 9: Deploy Production ───────────
  deploy-production:
    needs: [deploy-staging, build-push]
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://api.example.com
    steps:
      - uses: actions/checkout@v4

      - name: Deploy to production
        run: |
          echo "Deploying ${{ needs.build-push.outputs.version }} to production"
          # kubectl set image deployment/myapp ...

      - name: Smoke test production
        run: |
          echo "Running smoke tests against production"
          # ./scripts/smoke-test.sh https://api.example.com

      - name: Notify
        if: success()
        run: |
          echo "Successfully deployed ${{ needs.build-push.outputs.version }} to production"
          # Send Slack notification, etc.
```

---

## 🟡 Pipeline Visualization

### How GitHub Actions Visualizes This

```
┌──────┐     ┌──────┐
│ lint │     │ test │     Runs in parallel
└──┬───┘     └──┬───┘
   │            │
   └─────┬──────┘
         │
   ┌─────▼─────┐
   │ build-push │     Only runs if both pass
   └─────┬─────┘
         │
   ┌─────▼──────────┐
   │ deploy-staging  │     Manual approval (environment)
   └─────┬──────────┘
         │
   ┌─────▼───────────────┐
   │ deploy-production    │     Manual approval (environment)
   └─────────────────────┘
```

### Pipeline Duration Target

| Stage | Target Duration | If Slower |
|-------|----------------|-----------|
| Checkout | <10 seconds | Use shallow clone |
| Dependencies | <30 seconds | Use caching |
| Lint | <1 minute | Parallelize linters |
| Unit tests | <2 minutes | Speed up tests, parallelize |
| Integration tests | <5 minutes | Use containers, parallelize |
| Build image | <2 minutes | Use build cache, multi-stage |
| Push image | <1 minute | Push only changed layers |
| Deploy staging | <2 minutes | Use rolling updates |
| Smoke tests | <1 minute | Test only critical paths |

**Total target: <15 minutes from commit to staging.**

---

## ✅ Hands-On Exercise

### Create a Local Pipeline Script

```bash
#!/bin/bash
# pipeline.sh — Simulates a CI/CD pipeline locally
set -e  # Exit on first failure (just like CI!)

echo "=============================="
echo "  CI/CD Pipeline (Local)"
echo "=============================="

# Stage 1: Checkout (already local)
echo ""
echo "Stage 1: Checkout ✓ (local)"

# Stage 2: Dependencies
echo ""
echo "Stage 2: Installing dependencies..."
go mod download
echo "Stage 2: Dependencies ✓"

# Stage 3: Lint
echo ""
echo "Stage 3: Linting..."
UNFORMATTED=$(gofmt -l . 2>/dev/null || true)
if [ -n "$UNFORMATTED" ]; then
  echo "FAIL: Unformatted files: $UNFORMATTED"
  exit 1
fi
go vet ./...
echo "Stage 3: Lint ✓"

# Stage 4: Test
echo ""
echo "Stage 4: Running tests..."
go test ./... -race -count=1 -timeout 2m
echo "Stage 4: Test ✓"

# Stage 5: Build
echo ""
echo "Stage 5: Building..."
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "local")
go build -ldflags="-X main.version=$VERSION -X main.commit=$COMMIT" -o ./bin/myapp .
echo "Stage 5: Build ✓ (version: $VERSION)"

# Stage 6: Package (Docker)
echo ""
echo "Stage 6: Building Docker image..."
if command -v docker &> /dev/null; then
  docker build -t myapp:$VERSION .
  echo "Stage 6: Docker ✓ (myapp:$VERSION)"
else
  echo "Stage 6: Docker skipped (not installed)"
fi

echo ""
echo "=============================="
echo "  Pipeline Complete! ✓"
echo "  Version: $VERSION"
echo "  Commit: $COMMIT"
echo "=============================="
```

**Save and run:**

```bash
chmod +x pipeline.sh
./pipeline.sh
```

**Break each stage and see the pipeline stop:**

```bash
# Break lint: Add unformatted code
# Break test: Change expected output
# Break build: Add syntax error
# Each stops the pipeline at that stage
```

---

## 📚 Summary

| Stage | Purpose | Speed Target | Fails When |
|-------|---------|-------------|-----------|
| Checkout | Get source code | <10s | Missing permissions, large repo |
| Dependencies | Install libraries | <30s | Network issues, version conflicts |
| Lint | Code quality | <1m | Style violations, type errors |
| Test | Correctness | <5m | Logic bugs, integration failures |
| Build | Create artifact | <2m | Compilation errors |
| Push | Store artifact | <1m | Registry auth, network |
| Deploy Staging | Ship to staging | <2m | K8s config, resource issues |
| Smoke Test | Verify deployment | <1m | App crashes, config missing |
| Deploy Prod | Ship to users | <2m | Approval needed, same as staging |

**Key principles:**
1. **Fail fast** — Run cheap checks first (lint before test)
2. **Cache everything** — Dependencies, Docker layers, build artifacts
3. **Parallelize** — Run independent stages simultaneously
4. **Gate deployments** — Staging must pass before production

---

**Previous:** [03. Immutable Artifacts](./03-immutable-artifacts.md)  
**Next:** [05. Deployment Strategies](./05-deployment-strategies.md)  
**Module:** [06. CI/CD Fundamentals](./README.md)
