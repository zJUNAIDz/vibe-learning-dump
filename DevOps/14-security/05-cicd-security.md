# CI/CD Security

> **Your CI/CD pipeline has access to production credentials, container registries, and cloud accounts. If an attacker compromises your pipeline, they own everything you deploy to.**

---

## 🟢 The Attack Surface

```
A typical CI/CD pipeline has access to:
  → Source code (read/write)
  → Container registry (push images)
  → Kubernetes cluster (deploy pods)
  → Cloud credentials (AWS, GCP, Azure)
  → Secrets (database passwords, API keys)
  → Artifact storage (S3, GCS)

If compromised:
  → Attacker pushes malicious code (backdoor in image)
  → Attacker steals all secrets (database credentials)
  → Attacker deploys to production (crypto miner in cluster)
  → Attacker modifies infrastructure (opens firewall rules)

This is not theoretical. SolarWinds, Codecov, and dozens of
less-publicized incidents exploited CI/CD pipelines.
```

---

## 🟢 Secrets in Pipelines

### Never Do This

```yaml
# BAD — secrets in pipeline config (visible in Git)
env:
  DATABASE_URL: "postgres://admin:password@db.prod.internal:5432/app"
  AWS_ACCESS_KEY_ID: "AKIAIOSFODNN7EXAMPLE"
  AWS_SECRET_ACCESS_KEY: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

# BAD — echoing secrets in logs
steps:
  - run: echo "Deploying with key $AWS_SECRET_ACCESS_KEY"

# BAD — secrets in build args (visible in image layers)
docker build --build-arg DB_PASSWORD=secret -t my-app .
```

### Use CI/CD Secret Management

```yaml
# GitHub Actions — secrets stored encrypted, never logged
jobs:
  deploy:
    steps:
      - name: Configure AWS
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: arn:aws:iam::123456789:role/github-deploy
          aws-region: us-east-1
          # No static keys! Uses OIDC federation

      - name: Deploy
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
        run: |
          # GitHub automatically masks secret values in logs
          kubectl apply -f k8s/
```

```groovy
// Jenkins — use Credentials plugin
pipeline {
    environment {
        DB_CREDS = credentials('production-db')  // From Jenkins credential store
    }
    stages {
        stage('Deploy') {
            steps {
                sh '''
                    # Jenkins masks $DB_CREDS in logs
                    export DATABASE_URL="postgres://${DB_CREDS_USR}:${DB_CREDS_PSW}@db:5432/app"
                    ./deploy.sh
                '''
            }
        }
    }
}
```

---

## 🟡 OIDC Federation (No Static Keys)

```
Problem: CI/CD needs AWS/GCP credentials to deploy
Old way: Store static access keys as CI secrets
  → Keys never expire
  → Keys can be stolen
  → Hard to rotate

Better way: OIDC federation
  → CI/CD proves its identity to cloud provider
  → Cloud provides short-lived credentials (15 min)
  → No static keys anywhere
  → Credentials auto-expire
```

### GitHub Actions + AWS OIDC

```yaml
# 1. Terraform: Create OIDC provider and role in AWS
resource "aws_iam_openid_connect_provider" "github" {
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = ["6938fd4d98bab03faadb97b34396831e3780aea1"]
}

resource "aws_iam_role" "github_deploy" {
  name = "github-deploy"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Federated = aws_iam_openid_connect_provider.github.arn
      }
      Action = "sts:AssumeRoleWithWebIdentity"
      Condition = {
        StringEquals = {
          "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
        }
        StringLike = {
          # Only allow from your repo's main branch
          "token.actions.githubusercontent.com:sub" = "repo:myorg/myrepo:ref:refs/heads/main"
        }
      }
    }]
  })
}

# 2. GitHub Actions workflow
jobs:
  deploy:
    permissions:
      id-token: write    # Required for OIDC
      contents: read
    steps:
      - uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: arn:aws:iam::123456789:role/github-deploy
          aws-region: us-east-1
          # No secrets needed! GitHub proves identity via OIDC
```

---

## 🟡 Build Isolation

### Preventing Supply Chain Attacks in CI

```
Threats during build:
  → Malicious dependency installed by npm install
  → Compromised build tool (webpack, esbuild)
  → Modified source code (PR from attacker)
  → Build cache poisoning

Mitigations:

1. PIN EVERYTHING
   → npm ci (not npm install)
   → Pin actions versions: uses: actions/checkout@v4.1.1
   → Pin image digests: FROM node:20@sha256:abc123...
   → Pin tool versions

2. ISOLATE BUILDS
   → Each build in a fresh container (default in GitHub Actions)
   → No shared state between builds
   → No persistent cache that could be poisoned

3. LIMIT PERMISSIONS
   → Build jobs shouldn't have deploy credentials
   → Separate build job from deploy job
   → Deploy only from main branch
```

```yaml
# Good: Separate build and deploy with branch restrictions
jobs:
  build:
    # Runs on all branches and PRs
    # Has NO secrets except container registry token
    steps:
      - uses: actions/checkout@v4
      - run: npm ci
      - run: npm test
      - run: npm run build
      - run: docker build -t my-app:${{ github.sha }} .
      - run: docker push my-app:${{ github.sha }}
  
  deploy:
    needs: build
    # ONLY runs on main branch
    if: github.ref == 'refs/heads/main'
    environment: production  # Requires approval
    permissions:
      id-token: write
    steps:
      - uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.DEPLOY_ROLE_ARN }}
      - run: kubectl set image deployment/app app=my-app:${{ github.sha }}
```

### Pin Action Versions

```yaml
# BAD — mutable tag, could be changed by attacker
- uses: actions/checkout@v4

# BETTER — pin to exact commit SHA
- uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1

# Why?
# → If an attacker compromises the actions/checkout repo
# → They can push malicious code to the v4 tag
# → SHA pins are immutable — can't be changed after creation
```

---

## 🟡 Artifact Signing and Verification

```
Sign your build artifacts so you can verify:
  → This artifact was built by OUR CI pipeline
  → It hasn't been tampered with since building
  → It came from THIS specific commit

Without signing:
  → Someone pushes a malicious image to your registry
  → Kubernetes pulls and runs it
  → You're running attacker's code in production

With signing:
  → CI builds and signs image
  → Kubernetes verifies signature before running
  → Unsigned/tampered images are REJECTED
```

```yaml
# Full pipeline: build, scan, sign, deploy
jobs:
  build:
    steps:
      - uses: actions/checkout@v4
      - run: docker build -t my-registry/app:${{ github.sha }} .
      - run: docker push my-registry/app:${{ github.sha }}
  
  scan:
    needs: build
    steps:
      - name: Scan for vulnerabilities
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: my-registry/app:${{ github.sha }}
          exit-code: '1'
          severity: 'CRITICAL'
  
  sign:
    needs: scan
    permissions:
      id-token: write
    steps:
      - uses: sigstore/cosign-installer@v3
      - run: cosign sign my-registry/app:${{ github.sha }}
  
  deploy:
    needs: sign
    if: github.ref == 'refs/heads/main'
    environment: production
    steps:
      # Verify signature before deploying
      - run: cosign verify my-registry/app:${{ github.sha }}
      - run: kubectl set image deployment/app app=my-registry/app:${{ github.sha }}
```

---

## 🔴 Branch Protection and Review

```
Your CI/CD pipeline trusts what's in the main branch.
If anyone can push to main, anyone can deploy anything.

Required protections:
  → Require pull request reviews (at least 1 reviewer)
  → Require status checks to pass (tests, linting, scanning)
  → No direct pushes to main
  → No force pushes to main
  → Require signed commits (optional but good)
  → Require review from code owners for critical paths
```

```yaml
# CODEOWNERS — require specific reviewers for sensitive files
# .github/CODEOWNERS

# Infrastructure changes require platform team review
/terraform/           @myorg/platform-team
/k8s/                 @myorg/platform-team
.github/workflows/    @myorg/platform-team

# Security-sensitive files
/src/auth/            @myorg/security-team
Dockerfile            @myorg/platform-team
```

---

## 🔴 Anti-Patterns

```
❌ Static AWS keys in CI secrets
   → Use OIDC federation for cloud providers
   → Short-lived credentials > long-lived keys

❌ Pipeline can deploy from any branch
   → Only deploy from main/release branches
   → Require PR review before merge to main

❌ No separation between build and deploy
   → Build job has production credentials
   → Compromised build = compromised production

❌ Mutable action/image references
   → uses: actions/checkout@v4 (tag can be moved)
   → FROM node:20 (rebuilt daily)
   → Pin to SHA digests for security-critical pipelines

❌ No audit trail
   → Who triggered the deploy? What changed?
   → Use environments with required approvals
   → Enable audit logs on CI platform

❌ Secrets accessible to PR builds from forks
   → Attacker forks repo, modifies workflow, gets your secrets
   → GitHub: Secrets are NOT available to fork PRs by default
   → Never change this setting
```

---

**Previous:** [04. Supply Chain Security](./04-supply-chain-security.md)  
**Up:** [README](./README.md)
