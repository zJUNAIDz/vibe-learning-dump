# Leaking Secrets

> **Once a secret is in Git history, it's compromised. You can't un-push it. Bots scan GitHub in real time. AWS keys last about 30 seconds before someone finds them.**

---

## 🟢 Committed to Git

### How It Happens

```
Developer workflow:
  1. Create .env file with database password
  2. Forget to add .env to .gitignore
  3. git add .
  4. git commit -m "add config"
  5. git push

Timeline after push to public repo:
  0:00 — Push to GitHub
  0:02 — Bot scrapes GitHub events API
  0:05 — Bot finds AWS_SECRET_ACCESS_KEY in your code
  0:10 — Bot launches 200 EC2 instances for crypto mining
  0:30 — Your AWS bill is already $500
  2:00 — Your AWS bill is $5,000
  Next morning — You wake up to a $50,000+ bill

This happens DAILY. GitHub has reported millions of
secret exposures per year across public repositories.
```

### The "I Deleted It" Myth

```bash
# "I removed the secret and pushed again, we're fine"
# NO. We are NOT fine.

git log --all --full-history -- .env
# commit abc123: "add config" ← SECRET IS HERE
# commit def456: "remove .env" ← deletion

# Anyone can still see it:
git show abc123:.env
# DATABASE_URL=postgres://admin:P@ssw0rd@prod-db.internal:5432/app

# Even after force push, GitHub may cache commits
# Forks have copies
# CI build logs have copies
# Developer laptops have copies

# The ONLY fix: rotate the secret immediately
```

### Prevention

```bash
# 1. .gitignore (most basic defense)
# .gitignore
.env
.env.*
*.pem
*.key
id_rsa
credentials.json
service-account.json

# 2. Pre-commit hooks (catches before commit)
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/gitleaks/gitleaks
    rev: v8.18.2
    hooks:
      - id: gitleaks

# Install
pip install pre-commit
pre-commit install

# Now trying to commit a secret:
# $ git commit -m "oops"
# gitleaks...Failed
# Finding: AWS Secret Access Key
# File: config.js:3

# 3. GitHub Secret Scanning (automatic for public repos)
# GitHub Advanced Security for private repos (paid)
# Notifies you AND the cloud provider
```

### If a Secret Is Already in Git

```bash
# Step 1: ROTATE THE SECRET IMMEDIATELY
# Change the password/key/token RIGHT NOW
# Do NOT wait until you've cleaned Git history

# Step 2: Clean Git history (optional — secret is already compromised)
# Using git-filter-repo (recommended over BFG)
pip install git-filter-repo
git filter-repo --path .env --invert-paths

# Or using BFG Repo-Cleaner
bfg --delete-files .env
git reflog expire --expire=now --all
git gc --prune=now --aggressive
git push --force

# Step 3: Verify
gitleaks detect --source . --verbose

# Step 4: Notify
# → Tell your team
# → Check for unauthorized access using the exposed credential
# → Check cloud provider logs (AWS CloudTrail, etc.)
```

---

## 🟡 Exposed in Logs

### How It Happens

```typescript
// BAD — logging the full request
app.use((req, res, next) => {
  console.log('Request:', JSON.stringify(req.headers));
  // Logs: { "authorization": "Bearer eyJhbGciOi..." }
  // Now anyone with log access has the user's auth token
  next();
});

// BAD — logging errors with full context
try {
  await connectDB(connectionString);
} catch (error) {
  console.error('DB connection failed:', connectionString);
  // Logs: DB connection failed: postgres://admin:P@ssw0rd@db:5432/app
  // Password is now in logs, log aggregation system, log backups...
}

// BAD — debug logging that never gets removed
console.log('Payment payload:', JSON.stringify(paymentData));
// Logs: { "card_number": "4111111111111111", "cvv": "123" }
// PCI violation. Audit failure. Possible fine.
```

### The Fix

```typescript
// GOOD — structured logging with field filtering
import pino from 'pino';

const logger = pino({
  redact: {
    paths: [
      'req.headers.authorization',
      'req.headers.cookie',
      'password',
      'creditCard',
      'ssn',
      '*.token',
      '*.secret',
    ],
    censor: '[REDACTED]',
  },
});

// Logs: { "req": { "headers": { "authorization": "[REDACTED]" } } }

// GOOD — log what you need, not everything
logger.info({
  action: 'db_connect',
  host: dbConfig.host,
  database: dbConfig.database,
  // NOT the password or full connection string
}, 'Connecting to database');

// GOOD — error logging without sensitive data
try {
  await processPayment(order);
} catch (error) {
  logger.error({
    action: 'payment_failed',
    orderId: order.id,
    errorCode: error.code,
    // NOT the card number, CVV, or full error object
  }, 'Payment processing failed');
}
```

---

## 🟡 Docker Image Layers

### How Secrets Leak Through Layers

```dockerfile
# BAD — secret is baked into image layer
FROM node:20-slim
WORKDIR /app
COPY . .

# This creates a layer with the .env file
# Even if you delete it later, the layer still exists
ENV DATABASE_URL=postgres://admin:secret@db:5432/app
RUN npm ci

# "Deleting" the secret doesn't help — previous layer still has it
RUN rm .env
```

```bash
# Attacker can extract secrets from any layer:
docker save my-app:latest | tar -x
# Each layer/ directory contains a layer.tar
# Extract and search:
find . -name "layer.tar" -exec tar -tf {} \; | grep -i env
# Found it! Now extract and read the .env file
```

### The Fix

```dockerfile
# GOOD — multi-stage build, secrets never in final image
FROM node:20-slim AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Final image — only built artifacts, no secrets
FROM node:20-slim
WORKDIR /app
RUN groupadd -r app && useradd -r -g app app
COPY --from=builder --chown=app:app /app/dist ./dist
COPY --from=builder --chown=app:app /app/node_modules ./node_modules
USER app
CMD ["node", "dist/server.js"]
# No .env, no source code, no build tools
```

```dockerfile
# GOOD — use Docker BuildKit secrets for build-time secrets
# syntax=docker/dockerfile:1
FROM node:20-slim
WORKDIR /app
COPY package*.json ./

# Secret is mounted temporarily during RUN, never saved to layer
RUN --mount=type=secret,id=npmrc,target=/root/.npmrc \
    npm ci

COPY . .
RUN npm run build
```

```bash
# Build with secret
docker build --secret id=npmrc,src=$HOME/.npmrc -t my-app .
```

### Scanning Images for Secrets

```bash
# Use Trivy to scan for secrets in images
trivy image --scanners secret my-app:latest

# Output:
# SECRET
# ══════
# my-app:latest (secrets)
# 
# app/config.js (secrets)
# ┌──────────┬───────────────────────────────┐
# │ Category │ Description                   │
# ├──────────┼───────────────────────────────┤
# │ AWS      │ AWS Access Key ID             │
# └──────────┴───────────────────────────────┘
```

---

## 🔴 Environment Variable Exposure

```
Environment variables aren't as secret as you think:

1. docker inspect shows ALL env vars
   docker inspect <container> | grep -A 50 Env

2. /proc/1/environ on Linux shows process env
   kubectl exec pod -- cat /proc/1/environ

3. Crash dumps may include env vars

4. Child processes inherit all parent env vars
   → Your app spawns a subprocess
   → Subprocess has all your secrets
   → If subprocess is compromised, secrets are exposed

5. Kubernetes: env vars visible in pod spec
   kubectl get pod app -o yaml | grep -A 5 env
   
Better alternative: mount secrets as files
  → Only the process that reads the file has the secret
  → File permissions can be restricted
  → Not visible in process listing or pod spec
```

---

## 🔴 Incident Response Checklist

```
When you discover a leaked secret:

IMMEDIATE (within 5 minutes):
  □ Rotate the secret (change password/regenerate key)
  □ Update all services using the old secret
  □ Verify new secret works

NEXT HOUR:
  □ Check cloud provider audit logs
    → AWS: CloudTrail
    → GCP: Audit Logs
    → Azure: Activity Log
  □ Check for unauthorized usage
    → Unexpected EC2 instances
    → Unexpected API calls
    → Unexpected data access
  □ Revoke old secret (don't just rotate — explicitly revoke)

NEXT DAY:
  □ Clean Git history (if applicable)
  □ Add pre-commit hooks to prevent recurrence
  □ Update .gitignore
  □ Post-mortem: how did it happen? How to prevent?
  □ Notify security team
  □ If user data exposed: legal/compliance notification
```

---

**Previous:** [01. Broken Deploys](./01-broken-deploys.md)  
**Next:** [03. CI Pipeline Disasters](./03-ci-pipeline-disasters.md)
