# Build vs Test vs Deploy

> **Every CI/CD pipeline has three core phases. Understanding each one deeply prevents 90% of pipeline problems.**

---

## 🟢 The Three Phases

```
┌─────────┐      ┌──────────┐      ┌──────────┐
│  BUILD  │ ───→ │   TEST   │ ───→ │  DEPLOY  │
│         │      │          │      │          │
│ Compile │      │ Verify   │      │ Ship     │
│ Bundle  │      │ Quality  │      │ Release  │
│ Package │      │ Safety   │      │ Deliver  │
└─────────┘      └──────────┘      └──────────┘
```

**If BUILD fails** → Code doesn't compile. Nothing else matters.  
**If TEST fails** → Code compiles but has bugs. Don't deploy.  
**If DEPLOY fails** → Code is good but can't reach users. Fix infrastructure.

---

## 🟢 Phase 1: Build

### What "Build" Means

**Build = Transform source code into something runnable.**

The output depends on your language:

| Language | Build Input | Build Output |
|----------|-------------|-------------|
| Go | `.go` source files | Single binary (`myapp`) |
| TypeScript | `.ts` source files | Transpiled `.js` files |
| Java | `.java` source files | `.jar` or `.war` file |
| Rust | `.rs` source files | Single binary |
| Docker | Dockerfile + source | Container image |

### Go Build

```bash
# Simple build
go build -o myapp ./cmd/server

# Production build (stripped, smaller binary)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" -o myapp ./cmd/server

# Cross-compile (build Linux binary on Mac)
GOOS=linux GOARCH=amd64 go build -o myapp-linux ./cmd/server
GOOS=darwin GOARCH=arm64 go build -o myapp-mac ./cmd/server
```

**What happens during `go build`:**
1. Parse source files
2. Type check
3. Compile to machine code
4. Link dependencies
5. Output single static binary

**If it fails:** Syntax errors, type errors, missing dependencies.

### TypeScript/Node.js Build

```bash
# Install dependencies
npm ci                    # Deterministic install (not npm install!)

# Type check
npx tsc --noEmit

# Build
npm run build             # Usually runs tsc + bundler
```

**What happens during a TypeScript build:**
1. Install exact dependencies from `package-lock.json`
2. Type check all `.ts` files
3. Transpile TypeScript → JavaScript
4. Bundle (webpack/esbuild/vite)
5. Output optimized `.js` files

**Why `npm ci` not `npm install`:**

```bash
npm install   # Reads package.json, may update lock file
              # Different runs may install different versions
              # NON-DETERMINISTIC

npm ci        # Reads package-lock.json exactly
              # Deletes node_modules, fresh install
              # Same input = same output always
              # DETERMINISTIC ← What you want in CI
```

### Docker Build (The Universal Build)

```dockerfile
# Multi-stage build for a Go app
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /myapp ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /myapp /myapp
EXPOSE 8080
ENTRYPOINT ["/myapp"]
```

```dockerfile
# Multi-stage build for a TypeScript app
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./
EXPOSE 3000
CMD ["node", "dist/index.js"]
```

**Why Docker for builds:**
- Same build on any machine (laptop, CI server, Mac, Linux)
- Dependencies included in the image
- Multi-stage = small final image
- The image IS the artifact

### Build Best Practices

**1. Deterministic builds:**

```bash
# Bad: Different results on different runs
npm install
pip install -r requirements.txt

# Good: Exact same results every time
npm ci
pip install -r requirements.txt --no-deps
go mod download
```

**2. Cache dependencies:**

```dockerfile
# Bad: Reinstalls dependencies on every code change
COPY . .
RUN npm ci

# Good: Only reinstalls when package.json changes
COPY package*.json ./
RUN npm ci
COPY . .
```

**3. Build metadata:**

```bash
# Embed version info into the binary
VERSION=$(git describe --tags --always)
COMMIT=$(git rev-parse --short HEAD)
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

go build -ldflags="-X main.version=$VERSION -X main.commit=$COMMIT -X main.buildTime=$BUILD_TIME" -o myapp
```

```go
// Access build metadata in your app
var (
    version   = "dev"
    commit    = "none"
    buildTime = "unknown"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{
        "version":   version,
        "commit":    commit,
        "buildTime": buildTime,
    })
}
```

---

## 🟢 Phase 2: Test

### The Testing Pyramid

```
        /  E2E   \           Slow, expensive, fewer
       / ________ \
      / Integration \        Medium speed, some
     / ______________ \
    /    Unit Tests     \    Fast, cheap, many
   /____________________\
```

**Run in this order:** Unit → Integration → E2E

**Why this order:** Fail fast. Unit tests take seconds. E2E tests take minutes. Don't wait 10 minutes for an E2E test when a unit test would have caught the bug in 3 seconds.

### Unit Tests

**What:** Test individual functions in isolation.  
**Speed:** Milliseconds per test.  
**Quantity:** Hundreds to thousands.

```go
// Go unit test
func TestCalculateDiscount(t *testing.T) {
    tests := []struct {
        name     string
        price    float64
        code     string
        expected float64
    }{
        {"no discount", 100.0, "", 100.0},
        {"10% off", 100.0, "SAVE10", 90.0},
        {"free shipping", 50.0, "FREESHIP", 50.0},
        {"invalid code", 100.0, "FAKE", 100.0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculateDiscount(tt.price, tt.code)
            if result != tt.expected {
                t.Errorf("expected %.2f, got %.2f", tt.expected, result)
            }
        })
    }
}
```

```typescript
// TypeScript unit test (Jest/Vitest)
describe('calculateDiscount', () => {
  it('returns original price with no code', () => {
    expect(calculateDiscount(100, '')).toBe(100);
  });

  it('applies 10% discount', () => {
    expect(calculateDiscount(100, 'SAVE10')).toBe(90);
  });

  it('ignores invalid code', () => {
    expect(calculateDiscount(100, 'FAKE')).toBe(100);
  });
});
```

### Integration Tests

**What:** Test how components work together (e.g., app + database).  
**Speed:** Seconds per test.  
**Quantity:** Dozens to hundreds.

```go
// Go integration test
func TestUserRepository_Create(t *testing.T) {
    // Real database needed
    db := setupTestDB(t)
    defer db.Close()

    repo := NewUserRepository(db)

    user, err := repo.Create(context.Background(), User{
        Email: "test@example.com",
        Name:  "Test User",
    })

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Verify it was actually saved
    found, err := repo.FindByID(context.Background(), user.ID)
    if err != nil {
        t.Fatalf("could not find user: %v", err)
    }
    if found.Email != "test@example.com" {
        t.Errorf("expected email test@example.com, got %s", found.Email)
    }
}
```

```typescript
// TypeScript integration test
describe('UserRepository', () => {
  let db: Database;
  let repo: UserRepository;

  beforeAll(async () => {
    db = await createTestDatabase();
    repo = new UserRepository(db);
  });

  afterAll(async () => {
    await db.close();
  });

  it('creates and retrieves a user', async () => {
    const created = await repo.create({
      email: 'test@example.com',
      name: 'Test User',
    });

    const found = await repo.findById(created.id);
    expect(found?.email).toBe('test@example.com');
  });
});
```

**Integration tests in CI often use Docker containers:**

```yaml
# Docker Compose for test dependencies
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: testdb
      POSTGRES_PASSWORD: testpass
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

### End-to-End (E2E) Tests

**What:** Test the entire system from the user's perspective.  
**Speed:** Seconds to minutes per test.  
**Quantity:** Tens (keep this small!).

```typescript
// Playwright E2E test
import { test, expect } from '@playwright/test';

test('user can sign up and log in', async ({ page }) => {
  // Sign up
  await page.goto('/signup');
  await page.fill('[name="email"]', 'newuser@test.com');
  await page.fill('[name="password"]', 'SecurePass123!');
  await page.click('button[type="submit"]');
  await expect(page).toHaveURL('/dashboard');

  // Log out
  await page.click('[data-testid="logout"]');
  await expect(page).toHaveURL('/login');

  // Log in
  await page.fill('[name="email"]', 'newuser@test.com');
  await page.fill('[name="password"]', 'SecurePass123!');
  await page.click('button[type="submit"]');
  await expect(page).toHaveURL('/dashboard');
  await expect(page.locator('[data-testid="welcome"]')).toContainText('newuser');
});
```

### Other Test Types in CI

**Linting (code quality):**

```bash
# Go
golangci-lint run ./...

# TypeScript
npx eslint . --ext .ts,.tsx
npx prettier --check .
```

**Security scanning:**

```bash
# Dependency vulnerabilities
npm audit
go list -json -deps ./... | nancy sleuth

# Container scanning
trivy image myapp:latest
```

**Type checking:**

```bash
# TypeScript (without emitting)
npx tsc --noEmit
```

### Test Configuration in CI

**Run tests in parallel where possible:**

```bash
# Go tests (parallel by default per package)
go test ./... -count=1 -race -timeout 5m

# Jest (parallel by default)
npx jest --ci --coverage --forceExit
```

**Set timeouts (don't let tests hang forever):**

```bash
# Go: 5 minute timeout
go test ./... -timeout 5m

# Jest: bail after first failure
npx jest --bail
```

---

## 🟢 Phase 3: Deploy

### What "Deploy" Means

**Deploy = Make the new version available to users.**

This is NOT just copying files. It involves:

1. **Package** the build output (Docker image, binary, bundle)
2. **Push** the package to a registry/store
3. **Update** the target environment to use the new package
4. **Verify** the new version is working
5. **Route traffic** to the new version

### Deployment Targets

**Docker/Kubernetes deployment:**

```bash
# 1. Tag the image
docker tag myapp:latest registry.example.com/myapp:v1.2.3

# 2. Push to registry
docker push registry.example.com/myapp:v1.2.3

# 3. Update Kubernetes
kubectl set image deployment/myapp myapp=registry.example.com/myapp:v1.2.3

# 4. Watch rollout
kubectl rollout status deployment/myapp

# 5. Verify
curl https://api.example.com/health
```

**Cloud platform deployment (e.g., AWS):**

```bash
# Push to ECR
aws ecr get-login-password | docker login --username AWS --password-stdin $ECR_URL
docker push $ECR_URL/myapp:v1.2.3

# Update ECS service
aws ecs update-service \
  --cluster production \
  --service myapp \
  --force-new-deployment

# Or trigger CDK/Terraform
cdk deploy --all --require-approval never
```

### Environments

**Typical environment progression:**

```
Development → Staging → Production
```

| Environment | Purpose | Who Uses It | Deploy Frequency |
|------------|---------|-------------|-----------------|
| Development | Developer testing | Developers | Every commit |
| Staging | Pre-production validation | QA, PM, Devs | Every merge to main |
| Production | Real users | Everyone | Manual approve or automated |

**Each environment should be as close to production as possible:**

```yaml
# staging/values.yaml
replicaCount: 2
resources:
  requests:
    cpu: 250m
    memory: 256Mi

# production/values.yaml
replicaCount: 5
resources:
  requests:
    cpu: 500m
    memory: 512Mi
```

**The only difference between staging and production should be scale, not configuration.**

### Post-Deployment Verification

**Smoke tests after deployment:**

```bash
#!/bin/bash
# smoke-test.sh

BASE_URL="${1:-https://staging.example.com}"

echo "Running smoke tests against $BASE_URL..."

# Health check
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
if [ "$STATUS" != "200" ]; then
  echo "FAIL: Health check returned $STATUS"
  exit 1
fi

# API responds
RESPONSE=$(curl -s "$BASE_URL/api/v1/status")
if ! echo "$RESPONSE" | jq -e '.status == "ok"' > /dev/null 2>&1; then
  echo "FAIL: API status not ok"
  exit 1
fi

echo "PASS: All smoke tests passed"
```

---

## 🟡 How the Phases Connect in a Real Pipeline

### GitHub Actions Example (Go + Docker + K8s)

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  # ── PHASE 1: BUILD ──────────────────────
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Download dependencies
        run: go mod download

      - name: Build binary
        run: |
          CGO_ENABLED=0 go build \
            -ldflags="-s -w -X main.version=${{ github.sha }}" \
            -o myapp ./cmd/server

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: myapp-binary
          path: myapp

  # ── PHASE 2: TEST ───────────────────────
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: testdb
          POSTGRES_PASSWORD: testpass
        ports:
          - 5432:5432
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run linter
        uses: golangci/golangci-lint-action@v4

      - name: Run unit tests
        run: go test ./internal/... -race -count=1

      - name: Run integration tests
        run: go test ./tests/integration/... -race -count=1
        env:
          DATABASE_URL: postgres://postgres:testpass@localhost:5432/testdb?sslmode=disable

  # ── PHASE 3: DEPLOY ─────────────────────
  deploy:
    needs: [build, test]       # Only deploy if build AND test pass
    if: github.ref == 'refs/heads/main'  # Only deploy from main
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build Docker image
        run: docker build -t myapp:${{ github.sha }} .

      - name: Push to registry
        run: |
          echo "${{ secrets.REGISTRY_PASSWORD }}" | docker login -u ${{ secrets.REGISTRY_USER }} --password-stdin
          docker tag myapp:${{ github.sha }} registry.example.com/myapp:${{ github.sha }}
          docker push registry.example.com/myapp:${{ github.sha }}

      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/myapp \
            myapp=registry.example.com/myapp:${{ github.sha }}
          kubectl rollout status deployment/myapp --timeout=300s

      - name: Run smoke tests
        run: ./scripts/smoke-test.sh https://api.example.com
```

### What Happens When Each Phase Fails

| Phase | Failure Example | Impact | Action |
|-------|----------------|--------|--------|
| Build | `syntax error line 42` | Nothing deployed | Fix code, re-push |
| Test | `TestCreateUser FAILED` | Build works but has bugs | Fix test or code, re-push |
| Deploy | `ImagePullBackOff` | Code is good but can't reach users | Fix registry/creds/config |

---

## 🔴 War Story: The Missing Phase

**What happened:** A team had CI that ran builds and tests. No smoke tests after deployment.

```yaml
# Their pipeline
build → test → push image → deploy
                             ↑
                             No verification after this!
```

**One day:** A deployment succeeded (exit code 0) but the app was crashing on startup because a new environment variable wasn't set.

```
Pod: CrashLoopBackOff
Log: "FATAL: DATABASE_URL not set"
```

**Nobody noticed for 2 hours** because the pipeline said "success."

**Fix:** Added post-deployment smoke tests.

```yaml
deploy:
  steps:
    - deploy-to-k8s
    - wait-for-rollout
    - run-smoke-tests         # ← This caught it next time
    - notify-slack
```

**Lesson:** A successful deployment ≠ a working application. Always verify after deploying.

---

## ✅ Hands-On Exercise

### Build the Three Phases Locally

**1. Create a project with all three phases:**

```bash
mkdir -p ~/pipeline-demo && cd ~/pipeline-demo

# Create a simple Go app
cat > main.go << 'EOF'
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

var version = "dev"

func healthHandler(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "ok",
        "version": version,
    })
}

func main() {
    http.HandleFunc("/health", healthHandler)
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    fmt.Printf("Starting server on :%s (version: %s)\n", port, version)
    http.ListenAndServe(":"+port, nil)
}
EOF

cat > main_test.go << 'EOF'
package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHealthHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/health", nil)
    rec := httptest.NewRecorder()
    healthHandler(rec, req)

    if rec.Code != http.StatusOK {
        t.Errorf("expected 200, got %d", rec.Code)
    }
}
EOF

go mod init pipeline-demo
```

**2. Run all three phases:**

```bash
# PHASE 1: BUILD
echo "=== BUILD ==="
go build -o myapp . && echo "BUILD: PASS" || echo "BUILD: FAIL"

# PHASE 2: TEST
echo "=== TEST ==="
go test ./... -v && echo "TEST: PASS" || echo "TEST: FAIL"

# PHASE 3: DEPLOY (local simulation)
echo "=== DEPLOY ==="
./myapp &
APP_PID=$!
sleep 1

# Smoke test
STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)
if [ "$STATUS" = "200" ]; then
    echo "DEPLOY: PASS (health check returned 200)"
else
    echo "DEPLOY: FAIL (health check returned $STATUS)"
fi

kill $APP_PID
```

**3. Break each phase and observe:**

```bash
# Break BUILD: Add syntax error
echo "invalid go code" >> main.go
go build -o myapp .    # FAIL: compilation error

# Break TEST: Make test wrong
# Change expected status code in test

# Break DEPLOY: Use wrong port
PORT=9999 ./myapp &    # App starts on 9999
curl http://localhost:8080/health  # FAIL: connection refused
```

---

## 📚 Summary

| Phase | Input | Output | Fails When |
|-------|-------|--------|-----------|
| **Build** | Source code | Binary/Image | Syntax errors, missing deps |
| **Test** | Built code | Pass/Fail | Logic bugs, integration issues |
| **Deploy** | Tested artifact | Running app | Infra issues, config errors |

**Key takeaways:**
1. **Build deterministically** — `npm ci` not `npm install`, lock files matter
2. **Test in layers** — Unit (fast) → Integration (medium) → E2E (slow)
3. **Verify after deploy** — Smoke tests are non-negotiable
4. **Fail fast** — Run the fastest checks first

---

**Previous:** [01. Why CI/CD Exists](./01-why-cicd-exists.md)  
**Next:** [03. Immutable Artifacts](./03-immutable-artifacts.md)  
**Module:** [06. CI/CD Fundamentals](./README.md)
