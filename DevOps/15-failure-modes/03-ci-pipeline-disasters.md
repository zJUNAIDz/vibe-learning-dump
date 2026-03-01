# CI Pipeline Disasters

> **A CI pipeline that runs unattended with production credentials is the most dangerous tool in your organization. When it goes wrong, it goes SPECTACULARLY wrong.**

---

## 🟢 Infinite Loops

### How It Happens

```
Scenario 1: CI triggers itself

  1. CI pipeline runs on push to main
  2. Pipeline auto-formats code and commits
  3. Commit triggers another pipeline run
  4. Pipeline auto-formats again and commits
  5. Repeat forever
  6. 2000 pipeline runs, $500 in compute costs

Scenario 2: Webhook loop

  1. GitHub webhook triggers Jenkins
  2. Jenkins updates GitHub status
  3. Status update triggers webhook
  4. Webhook triggers Jenkins again
  5. 10,000 builds queued in 30 minutes

Scenario 3: Self-deploying deploy

  1. CD pipeline deploys on every commit to main
  2. Deploy script modifies a ConfigMap
  3. ArgoCD detects ConfigMap change
  4. ArgoCD commits "sync" to Git
  5. Commit triggers CD pipeline again
```

### Prevention

```yaml
# GitHub Actions — skip CI for automated commits
jobs:
  format:
    # Don't run if the commit was made by the CI bot
    if: github.actor != 'github-actions[bot]'
    steps:
      - run: npm run format
      - run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add -A
          git diff --staged --quiet || git commit -m "chore: auto-format [skip ci]"
          # [skip ci] in commit message prevents re-triggering
          git push

# Jenkins — filter webhook events
pipeline {
    triggers {
        githubPush()
    }
    stages {
        stage('Check') {
            when {
                not {
                    changelog '.*\\[skip ci\\].*'
                }
            }
            steps {
                sh 'make build'
            }
        }
    }
}
```

---

## 🟢 Disk Space Exhaustion

### How It Happens

```
CI agents have limited disk space (50-100GB typically).

Causes:
  → Docker images never cleaned up
  → Build artifacts accumulated over months
  → npm/pip cache growing unbounded
  → Log files filling up
  → Test artifacts (screenshots, videos) not cleaned

Symptoms:
  → "No space left on device"
  → Docker build fails silently
  → npm install fails with cryptic errors
  → Tests fail with write errors

Real scenario:
  Monday: Jenkins agent starts with 50GB free
  Tuesday: 45GB free (Docker images from builds)
  Wednesday: 35GB free (more images, caches)
  Thursday: 10GB free (builds start failing intermittently)
  Friday: 0GB free (NOTHING works, all pipelines blocked)
```

### Prevention

```bash
# Docker cleanup (run daily on CI agents)
docker system prune -af --filter "until=24h"
docker volume prune -f

# In Kubernetes-based CI (ephemeral pods = no buildup)
# But still clean up Docker layer cache:
docker builder prune -af

# GitHub Actions — automatically clean
# Self-hosted runners need manual cleanup:
# /etc/cron.daily/docker-cleanup
#!/bin/bash
docker system prune -af --filter "until=48h"
docker volume prune -f
```

```yaml
# Kubernetes CI agents — use ephemeral storage limits
containers:
  - name: builder
    resources:
      limits:
        ephemeral-storage: "10Gi"
    # Pod evicted if it exceeds 10Gi — prevents disk fill
```

---

## 🟡 Credential Leaks in CI

### How Secrets Appear in Logs

```bash
# BAD — verbose mode exposes secrets
npm install --verbose
# output includes: //registry.npmjs.org/:_authToken=npm_XXXX

# BAD — debugging that logs environment
env | sort
# DATABASE_URL=postgres://admin:P@ssword@db:5432
# AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI...

# BAD — terraform plan shows sensitive values
terraform plan
# ~ password = "OldP@ss" -> "NewP@ss"

# BAD — docker build args in image history
docker build --build-arg NPM_TOKEN=xxx .
docker history my-app
# Shows: ARG NPM_TOKEN=xxx
```

### Prevention

```yaml
# GitHub Actions — secrets are auto-masked
# But can still leak through indirect exposure:

steps:
  - name: Deploy
    env:
      DB_URL: ${{ secrets.DATABASE_URL }}
    run: |
      # BAD — base64 encoding bypasses masking
      echo $DB_URL | base64
      # GitHub doesn't know to mask the base64 version!
      
      # BAD — character-by-character output
      echo $DB_URL | fold -w1
      # Masking doesn't work on split strings
      
      # GOOD — never echo secrets, even indirectly
      ./deploy.sh  # Script uses $DB_URL internally
```

```groovy
// Jenkins — use credentials binding
pipeline {
    stages {
        stage('Deploy') {
            steps {
                withCredentials([
                    usernamePassword(
                        credentialsId: 'db-creds',
                        usernameVariable: 'DB_USER',
                        passwordVariable: 'DB_PASS'
                    )
                ]) {
                    // Jenkins masks $DB_USER and $DB_PASS in logs
                    sh './deploy.sh'
                }
                // Credentials NOT available outside this block
            }
        }
    }
}
```

```hcl
# Terraform — mark outputs as sensitive
output "database_password" {
  value     = aws_db_instance.main.password
  sensitive = true  # Won't appear in plan/apply output
}

# Also in variables
variable "db_password" {
  type      = string
  sensitive = true
}
```

---

## 🟡 Flaky Tests Blocking Deploys

```
Scenario:
  → Test passes 95% of the time
  → 20 flaky tests × 95% pass rate = 36% chance all pass
  → Pipeline fails more often than it succeeds
  → Developers start ignoring test failures
  → ACTUAL bugs slip through because "it's probably just flaky"

The trap:
  → Team adds retry logic to all tests
  → Now tests pass but take 3x longer
  → Nobody fixes the actual flakiness
  → Eventually: 45-minute CI pipeline that's still unreliable
```

### Handling Flaky Tests

```yaml
# Strategy 1: Quarantine flaky tests
# Run them separately, track flakiness rate

# .github/workflows/ci.yaml
jobs:
  tests:
    steps:
      - run: npm test -- --exclude-pattern="**/*.flaky.test.ts"
  
  flaky-tests:
    continue-on-error: true  # Don't block pipeline
    steps:
      - run: npm test -- --pattern="**/*.flaky.test.ts"
      - name: Report flakiness
        if: failure()
        run: |
          # Track in metrics: which tests are flaky and how often
          curl -X POST $METRICS_URL -d "flaky_test_failed=1"
```

```bash
# Strategy 2: Retry only known-flaky tests (not all tests)
# Jest
npx jest --bail --forceExit

# If a specific test is flaky, fix it or mark it:
describe.skip('flaky: payment webhook test', () => {
  // TODO: Fix race condition in mock server
  // Tracking issue: #1234
});
```

---

## 🔴 Pipeline as Single Point of Failure

```
What happens when CI/CD goes down:

Scenario: Jenkins master crashes
  → No builds can run
  → No deployments can happen
  → Critical hotfix needed but can't build or deploy
  → Engineers try to deploy manually
  → Manual deploy has a typo
  → Now you have TWO outages

Scenario: GitHub Actions rate limited
  → Repository uses 100% of Actions minutes
  → All pipelines queued, none running
  → Deploy to production blocked for 3 hours
  → Meanwhile, production has a bug

Prevention:
  → Have a manual deploy runbook (tested monthly)
  → Multi-runner setup (not single Jenkins master)
  → Self-hosted runners as backup to cloud CI
  → Pipeline status monitoring with alerts
```

---

## 🔴 Long-Running Pipelines

```
Symptoms:
  → Pipeline takes 45+ minutes
  → Developers push code and context-switch
  → By the time pipeline finishes, they forgot what they changed
  → Merge conflicts multiply because branches live longer
  → Pipeline queue grows, feedback loop breaks

Root causes:
  → Running ALL tests (unit + integration + e2e) in sequence
  → Rebuilding everything from scratch (no caching)
  → Large Docker images (downloading 1GB base images)
  → Sequential jobs that could be parallel

Fixes:
  → Parallelize: run unit tests, lint, security scan simultaneously
  → Cache: npm cache, Docker layer cache, build artifacts
  → Split: fast checks first (lint + unit in 2 min), slow later
  → Optimize: smaller Docker images, faster test frameworks
```

```yaml
# Fast pipeline structure
jobs:
  # Phase 1: Fast checks (< 2 minutes)
  lint:
    steps: [checkout, lint]  # 30 seconds
  
  type-check:
    steps: [checkout, tsc]    # 45 seconds
  
  unit-test:
    steps: [checkout, jest]   # 90 seconds
  
  # Phase 2: Medium checks (< 5 minutes) — only if Phase 1 passes
  integration-test:
    needs: [lint, type-check, unit-test]
    steps: [checkout, docker-compose, test]
  
  security-scan:
    needs: [lint, type-check, unit-test]
    steps: [checkout, build, trivy]
  
  # Phase 3: Deploy (< 3 minutes) — only if Phase 2 passes
  deploy:
    needs: [integration-test, security-scan]
    if: github.ref == 'refs/heads/main'
    steps: [deploy-to-k8s]
```

---

**Previous:** [02. Leaking Secrets](./02-leaking-secrets.md)  
**Next:** [04. Kubernetes Outages](./04-kubernetes-outages.md)
