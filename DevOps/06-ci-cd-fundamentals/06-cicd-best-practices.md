# CI/CD Best Practices

> **The difference between a pipeline that helps your team and one that everyone hates comes down to these principles.**

---

## 🟢 Practice 1: Keep Builds Fast (< 10 Minutes)

### Why Speed Matters

```
3-minute pipeline:
  Developer pushes → gets feedback → fixes → pushes again
  Total: 10 minutes to iterate
  Developer stays in flow state ✓

30-minute pipeline:
  Developer pushes → opens Twitter → forgets what they were doing
  Gets notification → context switch → re-reads code
  Total: 45+ minutes to iterate
  Developer hates CI/CD ✗
```

**Research shows:** If CI takes more than 10 minutes, developers stop waiting for it and merge without checking results.

### How to Speed Up Pipelines

**1. Cache dependencies aggressively:**

```yaml
# GitHub Actions: Cache Go modules
- uses: actions/setup-go@v5
  with:
    go-version: '1.22'
    cache: true                # Built-in caching

# GitHub Actions: Cache npm
- uses: actions/setup-node@v4
  with:
    node-version: '20'
    cache: 'npm'              # Built-in caching
```

**Impact:** Dependency install from 2 minutes → 10 seconds.

**2. Parallelize independent jobs:**

```yaml
# BAD: Sequential
jobs:
  lint:
    ...
  test:
    needs: lint          # Waits for lint!
    ...

# GOOD: Parallel
jobs:
  lint:
    ...
  test:
    ...                  # Runs at the same time as lint!
  build:
    needs: [lint, test]  # Only build waits
```

**Impact:** Lint (1 min) + Test (3 min) = 4 min sequential → 3 min parallel.

**3. Use fast runners:**

```yaml
# Larger runners = faster builds
runs-on: ubuntu-latest          # 2 CPU, 7 GB RAM (free)
runs-on: ubuntu-latest-4-core   # 4 CPU, 16 GB RAM (paid)
runs-on: ubuntu-latest-8-core   # 8 CPU, 32 GB RAM (paid)
```

**4. Cache Docker layers:**

```yaml
- uses: docker/build-push-action@v5
  with:
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

**Impact:** Docker build from 5 minutes → 30 seconds (when only code changed).

**5. Only run what changed:**

```yaml
# Only run Go tests when Go files change
test-go:
  if: |
    contains(github.event.head_commit.modified, '.go') ||
    contains(github.event.head_commit.modified, 'go.mod')
```

### Time Budget

| Stage | Budget | If Over Budget |
|-------|--------|----------------|
| Checkout | <10s | Shallow clone |
| Dependencies | <30s | Cache |
| Lint | <1m | Parallelize linters |
| Unit tests | <2m | Parallelize, faster tests |
| Integration tests | <3m | Containers, test isolation |
| Build image | <2m | Layer cache, multi-stage |
| Push | <30s | Only push changed layers |
| **Total** | **<10m** | **Something needs fixing** |

---

## 🟢 Practice 2: Fail Fast

### Run the Cheapest Checks First

```yaml
jobs:
  # Stage 1: Instant checks (< 30 seconds)
  quick-checks:
    steps:
      - run: gofmt -l .              # 2 seconds
      - run: go vet ./...            # 5 seconds
      - run: npx prettier --check .  # 3 seconds

  # Stage 2: Medium checks (< 2 minutes)
  tests:
    needs: quick-checks     # Don't bother if formatting is wrong
    steps:
      - run: go test ./internal/... -short  # Unit tests only

  # Stage 3: Expensive checks (< 5 minutes)
  integration:
    needs: tests            # Don't bother if units fail
    steps:
      - run: go test ./tests/...    # Integration with DB

  # Stage 4: Most expensive (< 3 minutes)
  build:
    needs: integration      # Don't build if broken
    steps:
      - run: docker build .
```

**The principle:** A formatting error should fail in 2 seconds, not after a 5-minute Docker build.

### Fail Early, Fail Loud

```yaml
# GOOD: Fail immediately when a test fails
- run: go test ./... -failfast    # Stop at first failure

# GOOD: Fail on any lint warning
- run: npx eslint . --max-warnings 0

# GOOD: Fail on security issues
- run: npm audit --audit-level=high
```

---

## 🟢 Practice 3: Idempotent Pipelines

### What Idempotent Means

**Idempotent:** Running the pipeline multiple times produces the same result.

```bash
# IDEMPOTENT: Running twice gives same result
kubectl apply -f deployment.yaml   # Creates or updates
kubectl apply -f deployment.yaml   # Same state, no change

# NOT IDEMPOTENT: Running twice gives different result
kubectl create -f deployment.yaml  # Creates
kubectl create -f deployment.yaml  # ERROR: already exists!
```

### In Practice

**Deployment scripts:**

```bash
# BAD: Not idempotent
kubectl create namespace staging           # Fails if exists
kubectl create configmap app-config ...    # Fails if exists

# GOOD: Idempotent
kubectl create namespace staging --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f configmap.yaml
```

**Database migrations:**

```bash
# BAD: Not idempotent
psql -c "ALTER TABLE users ADD COLUMN avatar TEXT"
# Running twice: ERROR: column "avatar" already exists

# GOOD: Idempotent
psql -c "ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar TEXT"
```

**Pipeline scripts:**

```bash
# BAD: Not idempotent
mkdir /tmp/build                    # Fails if exists
echo "version=1" >> version.txt    # Appends every run

# GOOD: Idempotent
mkdir -p /tmp/build                 # Creates or ignores
echo "version=1" > version.txt     # Overwrites
```

### Why Idempotency Matters

```
Pipeline run #1: Fails at "deploy" step (network timeout)
Pipeline run #2: Re-run... but "build" creates duplicate artifacts?
                 "create namespace" fails?
                 "migrate database" runs twice?

With idempotent pipeline:
Pipeline run #2: Re-run works perfectly ✓
                 Same result as if run #1 had succeeded
```

---

## 🟡 Practice 4: Secrets Management

### Never Do This

```yaml
# NEVER commit secrets to code
env:
  DATABASE_URL: postgres://admin:r3alP@ssw0rd@db.example.com/prod  # 🚨
  API_KEY: sk-abc123def456                                          # 🚨

# NEVER echo secrets in CI
- run: echo "Using key: $API_KEY"  # Visible in CI logs!  # 🚨
```

### Correct Approach

**GitHub Actions Secrets:**

```yaml
# Secrets configured in GitHub Settings → Secrets
- name: Deploy
  env:
    DATABASE_URL: ${{ secrets.DATABASE_URL }}
    API_KEY: ${{ secrets.API_KEY }}
  run: ./deploy.sh
```

**GitHub automatically masks secrets in logs:**

```
# If accidentally echoed:
echo $DATABASE_URL
# Output: ***
```

### Secrets Hierarchy

```
Level 1: CI/CD platform secrets (GitHub Secrets, Jenkins Credentials)
  → Good for most cases
  → Encrypted at rest
  → Access controlled per environment

Level 2: External secrets manager (AWS Secrets Manager, HashiCorp Vault)
  → Better for many services sharing secrets
  → Automatic rotation
  → Audit trail

Level 3: Kubernetes Secrets + External Secrets Operator
  → Best for K8s-native workflows
  → Syncs secrets from external manager to K8s
```

### Environment-Specific Secrets

```yaml
# Different secrets per environment
deploy-staging:
  environment: staging     # Uses staging secrets
  steps:
    - run: echo "DB is ${{ secrets.DATABASE_URL }}"
    # Uses staging DATABASE_URL

deploy-production:
  environment: production  # Uses production secrets
  steps:
    - run: echo "DB is ${{ secrets.DATABASE_URL }}"
    # Uses production DATABASE_URL (different value!)
```

### Rotating Secrets

```bash
# 1. Create new secret
aws secretsmanager put-secret-value --secret-id db-password --secret-string "newpassword"

# 2. Update CI/CD secret (GitHub CLI)
gh secret set DATABASE_URL --body "postgres://admin:newpassword@db.example.com/prod"

# 3. Re-deploy (picks up new secret)
# Trigger pipeline or:
kubectl rollout restart deployment/myapp
```

---

## 🟡 Practice 5: Branch Strategies

### Trunk-Based Development (Recommended)

```
main  ──●──●──●──●──●──●──●──●──●→
           \   /  \  /      \  /
            ●─●    ●●        ●●
           feature  fix      feature
          (< 1 day) (hours)  (< 1 day)
```

**Rules:**
1. `main` is always deployable
2. Feature branches live < 1 day
3. Code reviews happen on small PRs (< 200 lines)
4. Feature flags hide incomplete features

```yaml
# CI pipeline for trunk-based
on:
  push:
    branches: [main]        # Every push to main triggers CI/CD
  pull_request:
    branches: [main]        # PRs get tested before merge
```

**Why trunk-based works:**
- Small changes = small risk
- Fast feedback (minutes, not days)
- No merge hell
- Everyone sees the latest code

### Feature Flags (How Trunk-Based Handles WIP)

```go
// Feature flag — deploy code before it's ready for users
func handler(w http.ResponseWriter, r *http.Request) {
    if featureflags.IsEnabled("new-checkout-flow", r.Context()) {
        newCheckoutFlow(w, r)    // New code (deployed but hidden)
    } else {
        oldCheckoutFlow(w, r)    // Old code (what users see)
    }
}
```

```typescript
// TypeScript feature flag
const showNewDashboard = await featureFlags.isEnabled('new-dashboard', {
  userId: user.id,
  percentage: 10, // 10% of users see it
});

if (showNewDashboard) {
  return <NewDashboard />;
}
return <OldDashboard />;
```

### GitFlow (When You Need It)

```
main     ──●──────────────●──────→  (production)
            \              ↑
develop ──●──●──●──●──●──●──●───→  (integration)
             \  /  \      /
              ●●    ●──●──
            feature  release branch

Hotfix:
main ──●─────●──────→
        \   ↑
         ●──   (hotfix branch)
```

**When GitFlow makes sense:**
- Multiple versions in production simultaneously
- Long release cycles (monthly, quarterly)
- Regulatory requirements for release process
- Large teams with separate QA phases

**When GitFlow is overkill:**
- Small teams (< 10 developers)
- Web applications (single deployed version)
- Continuous deployment to cloud

### CI Pipeline per Strategy

**Trunk-based:**

```yaml
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  ci:
    steps: [checkout, lint, test, build]
  deploy:
    if: github.ref == 'refs/heads/main'
    needs: ci
    steps: [deploy-staging, smoke-test, deploy-prod]
```

**GitFlow:**

```yaml
on:
  push:
    branches: [main, develop, 'release/**']
  pull_request:
    branches: [main, develop]

jobs:
  ci:
    steps: [checkout, lint, test, build]
  
  deploy-staging:
    if: github.ref == 'refs/heads/develop'
    needs: ci
    steps: [deploy-to-staging]
  
  deploy-prod:
    if: github.ref == 'refs/heads/main'
    needs: ci
    steps: [deploy-to-production]
```

---

## 🟡 Practice 6: Pipeline as Code

### Version Control Your Pipeline

```
my-repo/
├── .github/
│   └── workflows/
│       ├── ci.yaml           # Main CI/CD pipeline
│       ├── release.yaml      # Release workflow
│       └── cleanup.yaml      # Cleanup old images
├── Dockerfile
├── Makefile
├── scripts/
│   ├── smoke-test.sh
│   ├── migrate.sh
│   └── deploy.sh
└── src/
```

**Why pipeline as code:**
- Pipeline changes go through code review
- Pipeline history is in git (who changed what)
- Same pipeline process across all environments
- Tested with the same rigor as application code

### Makefile as Pipeline Interface

```makefile
# Makefile — Every developer runs the same commands

.PHONY: lint test build deploy

# Dependencies
deps:
	go mod download

# Linting
lint:
	golangci-lint run ./...
	gofmt -l . | tee /dev/stderr | (! read)

# Testing
test: deps
	go test ./... -race -count=1 -timeout 5m

test-coverage: deps
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Building
build: deps
	CGO_ENABLED=0 go build -o bin/myapp ./cmd/server

docker-build:
	docker build -t myapp:$(shell git describe --tags --always) .

# Deploying
deploy-staging: docker-build
	docker push registry.example.com/myapp:$(shell git describe --tags --always)
	kubectl set image deployment/myapp myapp=registry.example.com/myapp:$(shell git describe --tags --always) -n staging

# CI uses the same commands as developers
ci: lint test build
```

**Now CI just runs `make ci`, and every developer can run the same thing locally.**

---

## 🟡 Practice 7: Notification and Observability

### Notify on Failure (Not on Every Success)

```yaml
# Notify on failure
- name: Notify Slack on failure
  if: failure()
  uses: slackapi/slack-github-action@v1
  with:
    payload: |
      {
        "text": "🔴 Pipeline failed for ${{ github.repository }}",
        "blocks": [{
          "type": "section",
          "text": {
            "type": "mrkdwn",
            "text": "*Pipeline Failed*\nRepo: ${{ github.repository }}\nBranch: ${{ github.ref_name }}\nAuthor: ${{ github.actor }}\nCommit: ${{ github.event.head_commit.message }}\n<${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}|View Run>"
          }
        }]
      }
  env:
    SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}

# Notify on deployment (always useful)
- name: Notify deployment
  if: success()
  run: |
    curl -X POST "${{ secrets.SLACK_WEBHOOK }}" \
      -H 'Content-Type: application/json' \
      -d "{\"text\": \"✅ Deployed $VERSION to production\"}"
```

### Don't Spam the Team

```
BAD: "Pipeline succeeded" for every commit (50 per day)
     → Everyone mutes the channel

GOOD: Only notify on:
  → Pipeline failures (need attention)
  → Production deployments (team awareness)
  → Security scan findings (critical)
```

---

## 🔴 Practice 8: Handling Flaky Tests

### What Makes Tests Flaky

```
Test passes locally:     ✓
Test passes in CI:       ✓
Test passes in CI:       ✗  ← Same code, different result!
Test passes in CI:       ✓
```

**Common causes:**
- Race conditions in concurrent code
- Hardcoded ports that conflict
- Time-dependent tests (`time.Now()`)
- External service dependencies
- Test ordering dependencies
- Insufficient timeouts

### How to Handle Flaky Tests

**1. Retry tests (band-aid, not a fix):**

```yaml
- name: Run tests with retry
  uses: nick-fields/retry@v2
  with:
    timeout_minutes: 10
    max_attempts: 3
    command: go test ./... -race -count=1
```

**2. Quarantine flaky tests:**

```go
func TestFlakyIntegration(t *testing.T) {
    if os.Getenv("SKIP_FLAKY") == "true" {
        t.Skip("Skipping flaky test (tracked in ISSUE-123)")
    }
    // ... flaky test
}
```

**3. Fix the root cause:**

```go
// BAD: Time-dependent
func TestToken(t *testing.T) {
    token := GenerateToken(time.Now())
    time.Sleep(time.Millisecond) // Race: might expire!
    if !ValidateToken(token) {
        t.Fatal("token invalid")
    }
}

// GOOD: Inject time
func TestToken(t *testing.T) {
    now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    token := GenerateToken(now)
    if !ValidateToken(token, now.Add(time.Millisecond)) {
        t.Fatal("token invalid")
    }
}
```

```go
// BAD: Hardcoded port
func TestServer(t *testing.T) {
    srv := httptest.NewServer(handler)  // Random port ✓
    // NOT: listen on :8080 (conflicts with other tests!)
}
```

**4. Track flaky test metrics:**

```
Flaky test rate > 5%: Pipeline is unreliable
Flaky test rate > 1%: Keep an eye on it
Flaky test rate < 0.1%: Healthy
```

---

## 🔴 Practice 9: Monorepo Pipelines

### When Your Repo Has Multiple Services

```
monorepo/
├── services/
│   ├── api/
│   │   ├── main.go
│   │   ├── Dockerfile
│   │   └── go.mod
│   ├── worker/
│   │   ├── main.go
│   │   ├── Dockerfile
│   │   └── go.mod
│   └── web/
│       ├── src/
│       ├── Dockerfile
│       └── package.json
└── .github/
    └── workflows/
        └── ci.yaml
```

### Only Build What Changed

```yaml
name: Monorepo CI

on:
  push:
    branches: [main]

jobs:
  detect-changes:
    runs-on: ubuntu-latest
    outputs:
      api: ${{ steps.filter.outputs.api }}
      worker: ${{ steps.filter.outputs.worker }}
      web: ${{ steps.filter.outputs.web }}
    steps:
      - uses: actions/checkout@v4
      - uses: dorny/paths-filter@v3
        id: filter
        with:
          filters: |
            api:
              - 'services/api/**'
            worker:
              - 'services/worker/**'
            web:
              - 'services/web/**'

  build-api:
    needs: detect-changes
    if: needs.detect-changes.outputs.api == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: cd services/api && go test ./... && docker build .

  build-worker:
    needs: detect-changes
    if: needs.detect-changes.outputs.worker == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: cd services/worker && go test ./... && docker build .

  build-web:
    needs: detect-changes
    if: needs.detect-changes.outputs.web == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: cd services/web && npm ci && npm test && docker build .
```

**Why:** If only `services/api/` changed, don't rebuild `web` and `worker`. Saves time and money.

---

## 🎯 The CI/CD Checklist

### Pipeline Health Check

```
□ Pipeline runs under 10 minutes
□ Pipeline is triggered on every push and PR
□ Tests run with race detection enabled
□ Dependencies are cached
□ Docker layers are cached
□ Lint runs before tests (fail fast)
□ Secrets are in CI/CD platform, NOT in code
□ Deployments are idempotent
□ Post-deployment smoke tests exist
□ Notifications go to the right channels
□ Flaky test rate is under 1%
□ Pipeline definition is version-controlled
□ Developers can run the same checks locally
```

### Common Anti-Patterns

| Anti-Pattern | Why It's Bad | Fix |
|-------------|-------------|-----|
| `npm install` in CI | Non-deterministic | Use `npm ci` |
| No caching | Slow builds | Cache deps and layers |
| Tests after build | Wastes time on broken code | Test first |
| `latest` Docker tag | Can't reproduce | Use git SHA or semver |
| Manual deployment steps | Error-prone, not repeatable | Automate everything |
| Ignoring failing tests | Erodes trust in CI | Fix or quarantine flaky tests |
| No smoke tests | "Deployed" ≠ "Working" | Always verify after deploy |
| Secrets in code | Security breach waiting to happen | Use secrets manager |
| 30-minute pipelines | Developers stop checking CI | Optimize relentlessly |
| No rollback plan | YOLO deployments | Test rollback regularly |

---

## ✅ Hands-On Exercise

### Audit Your CI/CD Pipeline

If you have an existing project with CI/CD, run this audit:

```bash
#!/bin/bash
# cicd-audit.sh — Check your pipeline health

echo "CI/CD Pipeline Audit"
echo "===================="

# Check 1: Pipeline speed
echo ""
echo "1. How long does your pipeline take?"
echo "   Target: < 10 minutes"
echo "   Check your CI/CD dashboard for average duration."

# Check 2: Deterministic dependencies
echo ""
echo "2. Checking for lock files..."
if [ -f "go.sum" ]; then echo "   ✓ go.sum found"; else echo "   ✗ go.sum missing"; fi
if [ -f "package-lock.json" ]; then echo "   ✓ package-lock.json found"; else echo "   Check: package-lock.json or yarn.lock?"; fi

# Check 3: No secrets in code
echo ""
echo "3. Checking for potential secrets in code..."
SECRETS=$(grep -r "password\|api_key\|secret_key\|access_key" --include="*.go" --include="*.ts" --include="*.yaml" -l . 2>/dev/null | grep -v node_modules | grep -v ".git" | grep -v vendor | head -5)
if [ -n "$SECRETS" ]; then
  echo "   ⚠ Potential secrets found in:"
  echo "$SECRETS" | sed 's/^/   /'
else
  echo "   ✓ No obvious secrets in code"
fi

# Check 4: Dockerfile exists
echo ""
echo "4. Checking for Dockerfile..."
if [ -f "Dockerfile" ]; then
  echo "   ✓ Dockerfile found"
  # Check for multi-stage
  STAGES=$(grep -c "^FROM" Dockerfile)
  if [ "$STAGES" -gt 1 ]; then
    echo "   ✓ Multi-stage build ($STAGES stages)"
  else
    echo "   ⚠ Single-stage build (consider multi-stage)"
  fi
else
  echo "   ✗ No Dockerfile found"
fi

# Check 5: CI config exists
echo ""
echo "5. Checking for CI configuration..."
if [ -d ".github/workflows" ]; then echo "   ✓ GitHub Actions found"; fi
if [ -f "Jenkinsfile" ]; then echo "   ✓ Jenkinsfile found"; fi
if [ -f ".gitlab-ci.yml" ]; then echo "   ✓ GitLab CI found"; fi
if [ -f ".circleci/config.yml" ]; then echo "   ✓ CircleCI found"; fi

# Check 6: Tests exist
echo ""
echo "6. Checking for tests..."
GO_TESTS=$(find . -name "*_test.go" -not -path "./vendor/*" 2>/dev/null | wc -l)
JS_TESTS=$(find . -name "*.test.ts" -o -name "*.spec.ts" -not -path "./node_modules/*" 2>/dev/null | wc -l)
echo "   Go test files: $GO_TESTS"
echo "   TS test files: $JS_TESTS"

echo ""
echo "Audit complete!"
```

---

## 📚 Summary

| Practice | One-Line Summary |
|----------|-----------------|
| **Fast builds** | Under 10 minutes or developers ignore CI |
| **Fail fast** | Lint (2s) before tests (3min) before build (5min) |
| **Idempotent** | Running the pipeline twice gives the same result |
| **Secrets management** | Never in code, always in CI platform or vault |
| **Trunk-based dev** | Short branches, feature flags, frequent merges |
| **Pipeline as code** | Version-controlled, reviewed, same as app code |
| **Smart notifications** | Notify on failure and deploys, not on every success |
| **Fix flaky tests** | Track, quarantine, and fix — never ignore |
| **Monorepo awareness** | Only build what changed |

**The goal of CI/CD is confidence.** Confidence that your code works, confidence that your deployment is safe, confidence that you can fix anything quickly. These practices build that confidence.

---

**Previous:** [05. Deployment Strategies](./05-deployment-strategies.md)  
**Module:** [06. CI/CD Fundamentals](./README.md)
