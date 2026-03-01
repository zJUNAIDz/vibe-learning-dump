# Secrets Management

> **The #1 security mistake in DevOps: secrets in source code. If your database password is in Git, assume it's compromised. It doesn't matter if the repo is "private."**

---

## 🟢 What NOT to Do

### Secrets in Code

```python
# BAD — password in source code
DB_PASSWORD = "P@ssw0rd123!"
stripe_key = "sk_live_4eC39HqLyjWDarjtT1zdp7dc"

# This code will be:
#   → Pushed to Git (forever in history, even if you delete it)
#   → Visible to everyone with repo access
#   → Visible in CI/CD logs
#   → Copied to developer laptops
#   → Possibly leaked in error messages
```

### Secrets in Config Files

```yaml
# BAD — credentials in docker-compose.yaml
services:
  postgres:
    environment:
      POSTGRES_PASSWORD: "SuperSecret123"

# BAD — credentials in Terraform
resource "aws_db_instance" "main" {
  password = "my-db-password"  # Stored in state file too!
}

# BAD — credentials in Kubernetes manifests
apiVersion: v1
kind: Pod
spec:
  containers:
    - env:
        - name: DATABASE_URL
          value: "postgres://admin:password@db:5432/app"
```

### What Happens When Secrets Leak

```
1. Someone pushes AWS keys to public GitHub
2. Automated bots find it in < 30 seconds
3. They spin up crypto miners on your account
4. Your AWS bill: $50,000 in 2 hours
5. AWS might flag it, might not

This happens EVERY DAY. GitHub scans for secrets
and notifies providers, but you can't rely on that.
```

---

## 🟢 Environment Variables (Basic)

```bash
# Pass secrets via environment variables
# Never hardcode them

# Running locally
export DATABASE_URL="postgres://admin:secret@localhost:5432/app"
export STRIPE_KEY="sk_test_xxx"
node server.js

# Docker
docker run \
  -e DATABASE_URL="postgres://admin:secret@db:5432/app" \
  -e STRIPE_KEY="sk_test_xxx" \
  my-app

# docker-compose (using .env file)
# .env (NEVER commit this file — add to .gitignore)
DATABASE_URL=postgres://admin:secret@db:5432/app
STRIPE_KEY=sk_test_xxx

# docker-compose.yaml
services:
  app:
    env_file: .env
```

### The .env Problem

```
.env files are better than hardcoding, but:
  → Easy to accidentally commit (.gitignore miss)
  → Stored in plaintext on disk
  → Shared via Slack/email (insecure)
  → No rotation mechanism
  → No audit trail (who accessed what?)

.env files are fine for LOCAL development.
For staging/production, use proper secrets management.
```

---

## 🟡 Kubernetes Secrets

### Creating Secrets

```bash
# Create from literal values
kubectl create secret generic db-credentials \
  --from-literal=username=admin \
  --from-literal=password='S3cretP@ss!'

# Create from files
kubectl create secret generic tls-cert \
  --from-file=cert.pem \
  --from-file=key.pem

# Create from .env file
kubectl create secret generic app-secrets \
  --from-env-file=.env
```

### Using Secrets in Pods

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  template:
    spec:
      containers:
        - name: app
          image: my-app:latest
          
          # Option 1: As environment variables
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: db-credentials
                  key: url
            - name: STRIPE_KEY
              valueFrom:
                secretKeyRef:
                  name: stripe-secrets
                  key: api-key
          
          # Option 2: As mounted files (more secure)
          volumeMounts:
            - name: secrets
              mountPath: /etc/secrets
              readOnly: true
      
      volumes:
        - name: secrets
          secret:
            secretName: db-credentials
```

### The Kubernetes Secrets Problem

```
Kubernetes Secrets are NOT encrypted by default!
  → They're base64 encoded (NOT encryption)
  → Anyone with kubectl access can read them
  → They're stored in etcd (often unencrypted)
  → They appear in pod spec (kubectl get pod -o yaml)

$ echo "S3cretP@ss!" | base64
UzNjcmV0UEBzcyE=

$ echo "UzNjcmV0UEBzcyE=" | base64 -d
S3cretP@ss!

Mitigations:
  → Enable etcd encryption at rest
  → Use RBAC to restrict Secret access
  → Use External Secrets Operator (pulls from Vault/AWS)
  → Use Sealed Secrets (encrypts secrets in Git)
```

---

## 🟡 Sealed Secrets (GitOps-Friendly)

```
Problem: You want secrets in Git (GitOps) but can't store plaintext.
Solution: Sealed Secrets encrypts secrets with a cluster-specific key.

Flow:
  1. Developer creates a regular Secret
  2. kubeseal encrypts it → SealedSecret (safe for Git)
  3. SealedSecret controller in cluster decrypts → creates Secret
  4. Pods use the Secret normally

Only the cluster can decrypt. Even if someone reads your Git repo,
they can't decrypt the SealedSecrets without the cluster's private key.
```

```bash
# Install sealed-secrets controller
helm install sealed-secrets sealed-secrets/sealed-secrets \
  -n kube-system

# Encrypt a secret
kubectl create secret generic db-creds \
  --from-literal=password='S3cret!' \
  --dry-run=client -o yaml | \
  kubeseal --format yaml > sealed-db-creds.yaml

# sealed-db-creds.yaml is safe to commit to Git
# It looks like:
# apiVersion: bitnami.com/v1alpha1
# kind: SealedSecret
# spec:
#   encryptedData:
#     password: AgBy3i4OJSWK+PiTySYZZA9rO...  (encrypted)
```

---

## 🟡 External Secrets Operator

```
Syncs secrets from external providers into Kubernetes Secrets.

Supported backends:
  → AWS Secrets Manager
  → HashiCorp Vault
  → GCP Secret Manager
  → Azure Key Vault
```

```yaml
# Install
helm install external-secrets external-secrets/external-secrets \
  -n external-secrets --create-namespace

# SecretStore — connects to AWS Secrets Manager
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secrets
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
      auth:
        jwt:
          serviceAccountRef:
            name: external-secrets-sa

# ExternalSecret — pulls specific secret
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: db-credentials
spec:
  refreshInterval: 1h  # Sync every hour
  secretStoreRef:
    name: aws-secrets
    kind: SecretStore
  target:
    name: db-credentials  # K8s Secret name to create
  data:
    - secretKey: password
      remoteRef:
        key: production/database  # AWS secret name
        property: password        # JSON key in the secret
```

---

## 🔴 HashiCorp Vault

```
Vault is the gold standard for secrets management:
  → Centralized secret storage (encrypted)
  → Dynamic secrets (generates DB credentials on-demand)
  → Secret rotation (auto-rotate passwords)
  → Audit logging (who accessed what, when)
  → Fine-grained ACL (who can access which secrets)
  → Multiple auth methods (Kubernetes, AWS IAM, LDAP)

When to use Vault:
  → Multiple teams/services need secrets
  → Compliance requires audit trails
  → You need dynamic/rotating credentials
  → You're past the "3 engineers" stage
```

### Using Vault with Kubernetes

```bash
# Install Vault via Helm
helm install vault hashicorp/vault \
  -n vault --create-namespace \
  --set "server.dev.enabled=true"  # Dev mode only!

# Store a secret
vault kv put secret/myapp/database \
  username="admin" \
  password="S3cretP@ss!"

# Read a secret
vault kv get secret/myapp/database
```

```yaml
# Vault Agent Injector — auto-injects secrets into pods
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  template:
    metadata:
      annotations:
        vault.hashicorp.com/agent-inject: "true"
        vault.hashicorp.com/role: "myapp"
        vault.hashicorp.com/agent-inject-secret-db: "secret/myapp/database"
        vault.hashicorp.com/agent-inject-template-db: |
          {{- with secret "secret/myapp/database" -}}
          DATABASE_URL=postgres://{{ .Data.data.username }}:{{ .Data.data.password }}@db:5432/app
          {{- end }}
    spec:
      serviceAccountName: myapp
      containers:
        - name: app
          image: my-app:latest
          # Secret available at /vault/secrets/db
```

---

## 🔴 Secret Rotation

```
Static secrets are dangerous:
  → If leaked, valid forever (until manually rotated)
  → If shared, everyone has the same password
  → No audit trail of who used the secret

Best practice: Rotate secrets regularly.

Vault dynamic secrets:
  → App requests DB credentials from Vault
  → Vault creates a NEW user/password with limited TTL
  → Credentials auto-expire after 1 hour
  → Each app instance gets DIFFERENT credentials
  → Vault revokes credentials when pod dies

AWS Secrets Manager rotation:
  → Supports automatic rotation for RDS, Redshift, DocumentDB
  → Lambda function rotates the secret on schedule
  → Applications always get the latest version
```

---

## 🔴 Pre-Commit Secret Scanning

```bash
# Prevent secrets from ever entering Git

# gitleaks — popular open source scanner
brew install gitleaks

# Scan current repo
gitleaks detect --source . --verbose

# Pre-commit hook (catches secrets before commit)
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/gitleaks/gitleaks
    rev: v8.18.2
    hooks:
      - id: gitleaks

# Install hooks
pre-commit install

# Now if you try to commit a file with "sk_live_..." it blocks:
# $ git commit -m "add config"
# gitleaks...........................................................Failed
# - hook id: gitleaks
# - exit code: 1
# Finding: sk_live_4eC39HqLyjWDarjtT1zdp7dc
```

```bash
# trufflehog — another option, scans Git history
trufflehog git file://. --only-verified

# GitHub also runs secret scanning on public repos
# And GitHub Advanced Security scans private repos (paid)
```

---

## 🔴 Anti-Patterns

```
❌ "We'll rotate the secret later"
   → You won't. Automate rotation from day 1.

❌ Same password everywhere
   → Dev, staging, prod all use the same DB password
   → Dev laptop gets stolen → production compromised

❌ Secrets in CI/CD environment variables (visible in UI)
   → Use CI/CD secret management features
   → GitHub Actions: Settings → Secrets
   → Jenkins: Credentials plugin
   → Never echo secrets in build logs

❌ Sharing secrets via Slack/email
   → Use a secrets manager
   → If you MUST share: use a one-time link (e.g., One-Time Secret)
   → Delete after sharing

❌ Not revoking secrets when employee leaves
   → Rotate ALL secrets an employee had access to
   → This includes shared passwords, API keys, SSH keys
```

---

**Next:** [02. IAM Concepts](./02-iam-concepts.md)  
**Up:** [README](./README.md)
