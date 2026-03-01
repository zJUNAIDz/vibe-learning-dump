# Workspaces and Environments

> **Same code, different environments. The goal: production and staging should differ only in scale, never in structure.**

---

## 🟢 Strategy 1: Directory-Based Environments (Recommended)

```
terraform/
├── modules/                    # Shared modules
│   ├── vpc/
│   ├── compute/
│   └── database/
├── environments/
│   ├── dev/
│   │   ├── main.tf            # Uses modules with dev settings
│   │   ├── backend.tf         # key = "dev/terraform.tfstate"
│   │   ├── variables.tf
│   │   └── terraform.tfvars   # dev values
│   ├── staging/
│   │   ├── main.tf            # Same modules, staging settings
│   │   ├── backend.tf         # key = "staging/terraform.tfstate"
│   │   ├── variables.tf
│   │   └── terraform.tfvars   # staging values
│   └── production/
│       ├── main.tf            # Same modules, production settings
│       ├── backend.tf         # key = "production/terraform.tfstate"
│       ├── variables.tf
│       └── terraform.tfvars   # production values
```

```hcl
# environments/dev/terraform.tfvars
environment    = "dev"
instance_count = 1
instance_type  = "t3.small"
db_instance    = "db.t3.micro"
multi_az       = false

# environments/production/terraform.tfvars
environment    = "production"
instance_count = 5
instance_type  = "t3.large"
db_instance    = "db.r6g.xlarge"
multi_az       = true
```

```bash
# Deploy to dev
cd environments/dev
terraform init
terraform apply

# Deploy to production
cd environments/production
terraform init
terraform apply
```

**Benefits:**
- Complete isolation between environments (separate state files)
- Can have different configurations (production has WAF, dev doesn't)
- Changes to dev can't accidentally affect production
- Easy to understand and navigate

---

## 🟢 Strategy 2: Variable Files (Simpler projects)

```
terraform/
├── main.tf
├── variables.tf
├── outputs.tf
├── backend.tf
└── environments/
    ├── dev.tfvars
    ├── staging.tfvars
    └── production.tfvars
```

```bash
# Same code, different variable files
terraform plan -var-file=environments/dev.tfvars
terraform plan -var-file=environments/production.tfvars
```

**Warning:** This shares the SAME state file unless you change the backend key!

```hcl
# backend.tf — must parameterize the key
terraform {
  backend "s3" {
    bucket = "my-state-bucket"
    key    = "myapp/${terraform.workspace}/terraform.tfstate"  # ← Uses workspace name
    region = "us-east-1"
  }
}
```

---

## 🟡 Strategy 3: Terraform Workspaces

```bash
# Workspaces = named state files for the same code

# Create workspaces
terraform workspace new dev
terraform workspace new staging
terraform workspace new production

# List workspaces
terraform workspace list
#   default
# * dev
#   staging
#   production

# Switch workspace
terraform workspace select production

# Apply (uses the workspace-specific state)
terraform apply -var-file=environments/${terraform.workspace}.tfvars
```

```hcl
# Using workspace name in code
resource "aws_instance" "web" {
  instance_type = terraform.workspace == "production" ? "t3.large" : "t3.small"
  
  tags = {
    Environment = terraform.workspace
  }
}

locals {
  instance_count = {
    dev        = 1
    staging    = 2
    production = 5
  }
}

resource "aws_instance" "web" {
  count         = local.instance_count[terraform.workspace]
  instance_type = terraform.workspace == "production" ? "t3.large" : "t3.small"
}
```

### When Workspaces Are Useful

```
✅ Good for:
  - Same infrastructure, different scale
  - Temporary environments (feature branches)
  - Simple projects with identical structure across environments

❌ Bad for:
  - Environments with different resources (prod has WAF, dev doesn't)
  - Different providers per environment
  - Large teams (everyone shares the same code)
  - Production safety (too easy to be in wrong workspace)
```

### The Workspace Danger

```bash
# The Nightmare Scenario:
$ terraform workspace select production   # I THOUGHT I was in dev
$ terraform destroy                        # 💥 Destroys production!

# Prevention:
# 1. Always check: terraform workspace show
# 2. Use directory-based environments for production
# 3. Add workspace name to your shell prompt
# 4. Use CI/CD (not laptops) for production changes
```

---

## 🟡 Environment Parity

### What Should Be the Same

```
                  Dev          Staging       Production
                  ───          ───────       ──────────
Architecture:     Same ✅       Same ✅        Same ✅
Modules used:     Same ✅       Same ✅        Same ✅
Network topology: Same ✅       Same ✅        Same ✅
Security groups:  Same ✅       Same ✅        Same ✅
Monitoring:       Same ✅       Same ✅        Same ✅
```

### What Should Differ

```
                  Dev          Staging       Production
                  ───          ───────       ──────────
Instance count:   1            2             10
Instance type:    t3.small     t3.medium     t3.large
DB size:          db.t3.micro  db.t3.small   db.r6g.xlarge
Multi-AZ DB:     No           No             Yes
Auto-scaling:    No           Yes            Yes
CDN:             No           No             Yes
WAF:             No           No             Yes
```

```hcl
# Use conditionals for environment-specific features
resource "aws_waf_web_acl" "main" {
  count = var.environment == "production" ? 1 : 0
  # Only exists in production
}

resource "aws_cloudfront_distribution" "main" {
  count = var.enable_cdn ? 1 : 0
  # Variable controls whether CDN exists
}
```

---

## 🟡 CI/CD for Terraform

```yaml
# .github/workflows/terraform.yml
name: Terraform

on:
  pull_request:
    paths: ['terraform/**']
  push:
    branches: [main]
    paths: ['terraform/**']

jobs:
  plan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: 1.7.0
      
      - name: Terraform Init
        run: terraform init
        working-directory: terraform/environments/production
      
      - name: Terraform Format Check
        run: terraform fmt -check -recursive
        working-directory: terraform
      
      - name: Terraform Validate
        run: terraform validate
        working-directory: terraform/environments/production
      
      - name: Terraform Plan
        run: terraform plan -no-color -out=tfplan
        working-directory: terraform/environments/production
      
      # Post plan output as PR comment
      - uses: actions/github-script@v7
        if: github.event_name == 'pull_request'
        with:
          script: |
            const plan = require('fs').readFileSync('terraform/environments/production/tfplan.txt', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '```\n' + plan + '\n```'
            });

  apply:
    needs: plan
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: production        # Requires manual approval
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
      
      - name: Terraform Apply
        run: |
          terraform init
          terraform apply -auto-approve
        working-directory: terraform/environments/production
```

---

## 🔴 Common Environment Pitfalls

### Pitfall 1: Snowflake Environments

```
❌ Dev was set up 2 years ago with Terraform 0.12
   Staging was set up 1 year ago with Terraform 1.0
   Production was set up last month with Terraform 1.7
   
   All three use different module versions
   All three have different resource structures
   "Staging passed but production failed"

✅ Same Terraform version, same modules, same structure
   Only variables differ
```

### Pitfall 2: Testing in Production

```
❌ "We don't have a staging environment, so we test in production"
   → One bad change → outage
   
✅ Always have at least dev + production
   → Test in dev, promote to production
   → Cheaper to run a small dev environment than to have a production outage
```

---

**Previous:** [03. Modules](./03-modules.md)  
**Next:** [05. Best Practices](./05-best-practices.md)
