# Capstone Phase 04: CI/CD Pipeline

> **Every push to `main` should automatically test, build, scan, and deploy your application. No manual steps. No "works on my machine." The pipeline IS the deployment process.**

---

## 🟢 Pipeline Overview

```
┌─────────────────────────────────────────────────────────┐
│                    GitHub Actions                        │
│                                                          │
│  Push to main                                            │
│       │                                                  │
│       ▼                                                  │
│  ┌─────────┐    ┌─────────┐    ┌──────────┐             │
│  │  Test    │───►│  Build  │───►│  Scan    │             │
│  │  (Jest)  │    │ (Docker)│    │ (Trivy)  │             │
│  └─────────┘    └─────────┘    └──────────┘             │
│                                     │                    │
│                                     ▼                    │
│                              ┌──────────┐                │
│                              │  Push    │                │
│                              │ (GHCR)  │                │
│                              └──────────┘                │
│                                     │                    │
│                    ┌────────────────┼────────────┐       │
│                    ▼                ▼             │       │
│              ┌──────────┐    ┌──────────┐        │       │
│              │  Deploy  │    │  Deploy  │        │       │
│              │ Staging  │    │   Prod   │◄──┐    │       │
│              └──────────┘    └──────────┘   │    │       │
│                    │                        │    │       │
│                    ▼                  ┌─────┴─┐  │       │
│              ┌──────────┐            │Manual │  │       │
│              │  Smoke   │            │Approve│  │       │
│              │  Test    │────────────►│       │  │       │
│              └──────────┘            └───────┘  │       │
│                                                  │       │
└─────────────────────────────────────────────────────────┘
```

---

## 🟢 The Complete Workflow

```yaml
# .github/workflows/deploy.yaml
name: Build and Deploy

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

# Cancel in-progress runs for the same branch
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}/task-service

permissions:
  contents: read
  packages: write          # Push to GHCR
  id-token: write          # OIDC for cloud auth
  security-events: write   # Upload scan results

jobs:
  # ────────────────────────────────────
  # Stage 1: Test
  # ────────────────────────────────────
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: "20"
          cache: "npm"

      - name: Install dependencies
        run: npm ci

      - name: Lint
        run: npm run lint

      - name: Unit tests
        run: npm test -- --coverage

      - name: Upload coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage
          path: coverage/

  # ────────────────────────────────────
  # Stage 2: Build and Push Image
  # ────────────────────────────────────
  build:
    name: Build Image
    runs-on: ubuntu-latest
    needs: test
    # Only build on main branch pushes, not PRs
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    outputs:
      image-tag: ${{ steps.meta.outputs.tags }}
      image-digest: ${{ steps.build.outputs.digest }}
    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to GHCR
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=sha,prefix=
            type=raw,value=latest

      - name: Build and push
        id: build
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  # ────────────────────────────────────
  # Stage 3: Security Scan
  # ────────────────────────────────────
  scan:
    name: Security Scan
    runs-on: ubuntu-latest
    needs: build
    steps:
      - uses: actions/checkout@v4

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${{ needs.build.outputs.image-digest }}
          format: "sarif"
          output: "trivy-results.sarif"
          severity: "CRITICAL,HIGH"
          exit-code: "1"   # Fail pipeline on critical/high vulns

      - name: Upload scan results
        uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: "trivy-results.sarif"

  # ────────────────────────────────────
  # Stage 4: Deploy to Staging
  # ────────────────────────────────────
  deploy-staging:
    name: Deploy to Staging
    runs-on: ubuntu-latest
    needs: [build, scan]
    environment:
      name: staging
      url: https://staging.tasks.example.com
    steps:
      - uses: actions/checkout@v4

      - name: Configure kubectl
        uses: azure/setup-kubectl@v3

      - name: Set kubeconfig
        run: |
          mkdir -p ~/.kube
          echo "${{ secrets.KUBE_CONFIG_STAGING }}" | base64 -d > ~/.kube/config

      - name: Update image tag
        run: |
          cd k8s
          # Use kustomize or sed to update image
          kubectl set image deployment/task-service \
            task-service=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${{ needs.build.outputs.image-digest }} \
            -n task-service \
            --record

      - name: Wait for rollout
        run: |
          kubectl rollout status deployment/task-service \
            -n task-service \
            --timeout=120s

      - name: Verify deployment
        run: |
          # Wait for pods to be ready
          kubectl wait --for=condition=ready pod \
            -l app=task-service \
            -n task-service \
            --timeout=60s
          
          echo "Deployment successful"

  # ────────────────────────────────────
  # Stage 5: Smoke Tests
  # ────────────────────────────────────
  smoke-test:
    name: Smoke Tests
    runs-on: ubuntu-latest
    needs: deploy-staging
    steps:
      - uses: actions/checkout@v4

      - name: Health check
        run: |
          for i in {1..10}; do
            STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
              https://staging.tasks.example.com/health/ready)
            if [ "$STATUS" = "200" ]; then
              echo "Health check passed"
              break
            fi
            echo "Attempt $i: got $STATUS, retrying..."
            sleep 5
          done
          
          if [ "$STATUS" != "200" ]; then
            echo "Health check failed"
            exit 1
          fi

      - name: API smoke tests
        run: |
          BASE_URL="https://staging.tasks.example.com"
          
          # Create a task
          TASK=$(curl -s -X POST "$BASE_URL/api/tasks" \
            -H "Content-Type: application/json" \
            -d '{"title": "Smoke test task", "description": "CI smoke test"}')
          
          TASK_ID=$(echo "$TASK" | jq -r '.id')
          echo "Created task: $TASK_ID"
          
          # Get the task
          curl -sf "$BASE_URL/api/tasks/$TASK_ID" | jq .
          
          # Delete the task
          curl -sf -X DELETE "$BASE_URL/api/tasks/$TASK_ID"
          
          echo "Smoke tests passed!"

      - name: Check metrics endpoint
        run: |
          curl -sf https://staging.tasks.example.com/metrics | head -20
          echo "Metrics endpoint OK"

  # ────────────────────────────────────
  # Stage 6: Deploy to Production
  # ────────────────────────────────────
  deploy-production:
    name: Deploy to Production
    runs-on: ubuntu-latest
    needs: [build, smoke-test]
    environment:
      name: production
      url: https://tasks.example.com
    steps:
      - uses: actions/checkout@v4

      - name: Configure kubectl
        uses: azure/setup-kubectl@v3

      - name: Set kubeconfig
        run: |
          mkdir -p ~/.kube
          echo "${{ secrets.KUBE_CONFIG_PRODUCTION }}" | base64 -d > ~/.kube/config

      - name: Deploy with image digest
        run: |
          kubectl set image deployment/task-service \
            task-service=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${{ needs.build.outputs.image-digest }} \
            -n task-service \
            --record

      - name: Wait for rollout
        run: |
          kubectl rollout status deployment/task-service \
            -n task-service \
            --timeout=180s

      - name: Verify production
        run: |
          kubectl wait --for=condition=ready pod \
            -l app=task-service \
            -n task-service \
            --timeout=60s
          echo "Production deployment successful"
```

---

## 🟡 GitHub Environments Setup

The `environment` field in the workflow creates **approval gates**:

```
GitHub Repo → Settings → Environments

Staging:
  - No required reviewers (auto-deploy)
  - Deployment branches: main only

Production:
  - Required reviewers: 1+ team leads
  - Wait timer: 5 minutes (cool-down period)
  - Deployment branches: main only
  - Secrets: KUBE_CONFIG_PRODUCTION
```

When the pipeline hits `deploy-production`, GitHub pauses and waits for manual approval in the Actions UI.

---

## 🟡 PR Validation Workflow

```yaml
# .github/workflows/pr-check.yaml
name: PR Check

on:
  pull_request:
    branches: [main]

jobs:
  validate:
    name: Validate
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: "20"
          cache: "npm"

      - run: npm ci
      - run: npm run lint
      - run: npm test -- --coverage

      - name: Build Docker image (no push)
        run: docker build -t task-service:test .

      - name: Scan image
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: "task-service:test"
          severity: "CRITICAL,HIGH"
          exit-code: "1"
```

---

## 🟡 Rollback Strategy

```yaml
# .github/workflows/rollback.yaml
name: Rollback

on:
  workflow_dispatch:
    inputs:
      environment:
        description: "Environment to rollback"
        required: true
        type: choice
        options:
          - staging
          - production
      reason:
        description: "Reason for rollback"
        required: true
        type: string

jobs:
  rollback:
    name: Rollback ${{ inputs.environment }}
    runs-on: ubuntu-latest
    environment: ${{ inputs.environment }}
    steps:
      - name: Set kubeconfig
        run: |
          mkdir -p ~/.kube
          if [ "${{ inputs.environment }}" = "production" ]; then
            echo "${{ secrets.KUBE_CONFIG_PRODUCTION }}" | base64 -d > ~/.kube/config
          else
            echo "${{ secrets.KUBE_CONFIG_STAGING }}" | base64 -d > ~/.kube/config
          fi

      - name: Rollback
        run: |
          echo "Rolling back ${{ inputs.environment }}"
          echo "Reason: ${{ inputs.reason }}"
          
          kubectl rollout undo deployment/task-service -n task-service
          kubectl rollout status deployment/task-service -n task-service --timeout=120s
          
          echo "Rollback complete"
```

---

## 🔴 Secrets You Need

```
GitHub Repo → Settings → Secrets and variables → Actions

Repository secrets:
  KUBE_CONFIG_STAGING      # base64-encoded kubeconfig for staging cluster
  KUBE_CONFIG_PRODUCTION   # base64-encoded kubeconfig for production cluster

# Better: Use OIDC federation instead of static kubeconfigs
# See Module 14's CI/CD Security for setup
```

---

## 🔴 Anti-Patterns vs Correct

| Anti-Pattern | Correct |
|---|---|
| Deploy on every commit to every branch | Deploy only from `main` |
| Skip tests to deploy faster | Tests are the gate — always run |
| Use `:latest` tag | Use SHA digest (`@sha256:...`) |
| Static credentials in secrets | OIDC federation (no static keys) |
| No approval gate for prod | Manual approval via GitHub Environments |
| No rollback plan | `workflow_dispatch` rollback workflow |
| No smoke tests after deploy | Automated health + API smoke tests |
| Same pipeline for PRs and main | Separate: PRs validate, main deploys |

---

## 🔴 Checklist

```
□ Tests run on every push and PR
□ Docker build uses multi-stage (from Phase 02)
□ Image pushed to GHCR with SHA tag
□ Trivy scan fails pipeline on CRITICAL/HIGH vulns
□ Staging deploys automatically after scan passes
□ Smoke tests verify staging health, API, metrics
□ Production requires manual approval
□ Rollback workflow exists (workflow_dispatch)
□ Concurrency control prevents duplicate runs
□ Secrets managed properly (not hardcoded)
□ Image referenced by digest (not mutable tag)
```

---

**Previous:** [03. Kubernetes Manifests](./03-kubernetes-manifests.md)  
**Next:** [05. Infrastructure as Code](./05-infrastructure-as-code.md)
