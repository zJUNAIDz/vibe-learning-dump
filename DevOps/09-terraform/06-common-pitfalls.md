# Common Terraform Pitfalls

> **Every Terraform user hits these. Some are annoying, some are catastrophic. Know them before they know you.**

---

## 🟢 Pitfall 1: State File Corruption

### How It Happens

```
1. Two people run terraform apply simultaneously (no locking)
2. Terraform crashes mid-apply (network issue, OOM, Ctrl+C)
3. Someone manually edits the state file
4. S3 eventual consistency (rare but possible)
```

### Symptoms

```bash
$ terraform plan
Error: Failed to load state: ...
# or
Error: Unsupported state file format: ...
# or
Plan shows resources being created that already exist
```

### Recovery

```bash
# Step 1: Check for backup
ls -la terraform.tfstate.backup

# Step 2: If remote, check S3 versioning
aws s3api list-object-versions \
    --bucket my-state-bucket \
    --prefix "production/terraform.tfstate" \
    --max-items 5

# Restore previous version
aws s3api get-object \
    --bucket my-state-bucket \
    --key "production/terraform.tfstate" \
    --version-id "previous-version-id" \
    recovered.tfstate

# Step 3: Verify
terraform plan  # Should show current state accurately

# Step 4: If no backup → import everything from scratch 😱
terraform import aws_vpc.main vpc-abc123
terraform import aws_instance.web i-0abc123def
# ... for EVERY resource
```

### Prevention

```
✅ Always use remote state with versioning
✅ Always use state locking (DynamoDB)
✅ Never run terraform from laptops (use CI/CD)
✅ Never manually edit state files
✅ Set up monitoring for state file changes
```

---

## 🟢 Pitfall 2: Drift Detection

### The Scenario

```
You wrote Terraform for a security group:

resource "aws_security_group" "web" {
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

A panicking engineer opens port 22 via AWS console at 2 AM.
Your Terraform code doesn't know about this.

Next terraform apply:
  ⚠️ Will REMOVE the port 22 rule (reverting to code)
  
  Or worse: nobody runs terraform apply for weeks
  Port 22 stays open to the internet
  Attacker gets in
```

### Detection

```bash
# terraform plan detects drift
$ terraform plan
  # aws_security_group.web will be updated in-place
  ~ resource "aws_security_group" "web" {
      ~ ingress {
        - {from_port=22, to_port=22, cidr=["0.0.0.0/0"]}  # ← REMOVED
      }
    }

# Refresh state to match reality (without changing anything)
$ terraform apply -refresh-only
```

### Continuous Drift Detection

```yaml
# Run terraform plan on a schedule to detect drift
# .github/workflows/drift-detection.yml
name: Terraform Drift Detection

on:
  schedule:
    - cron: '0 6 * * *'   # Daily at 6 AM

jobs:
  detect-drift:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
      
      - name: Terraform Plan
        run: |
          terraform init
          terraform plan -detailed-exitcode
        working-directory: terraform/environments/production
        # Exit code 0 = no changes
        # Exit code 2 = changes detected (DRIFT!)
      
      - name: Alert on Drift
        if: failure()
        run: |
          curl -X POST "$SLACK_WEBHOOK" \
            -d '{"text":"🚨 Terraform drift detected in production!"}'
```

---

## 🟢 Pitfall 3: Resource Dependencies

### Implicit vs Explicit Dependencies

```hcl
# IMPLICIT: Terraform figures it out from references
resource "aws_subnet" "main" {
  vpc_id = aws_vpc.main.id     # ← Terraform knows: create VPC first
}

# EXPLICIT: You tell Terraform about hidden dependencies
resource "aws_instance" "web" {
  ami           = "ami-12345678"
  instance_type = "t3.small"
  
  depends_on = [aws_iam_role_policy_attachment.web]
  # ↑ The instance needs this IAM policy, but there's no direct reference
  #   Without depends_on, the instance might start before the policy is attached
}
```

### Circular Dependencies

```hcl
# ❌ This will fail
resource "aws_security_group" "a" {
  ingress {
    security_groups = [aws_security_group.b.id]   # A references B
  }
}

resource "aws_security_group" "b" {
  ingress {
    security_groups = [aws_security_group.a.id]   # B references A → CIRCULAR!
  }
}

# ✅ Fix: Use separate security group rules
resource "aws_security_group" "a" { }
resource "aws_security_group" "b" { }

resource "aws_security_group_rule" "a_from_b" {
  type                     = "ingress"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.a.id
  source_security_group_id = aws_security_group.b.id
}

resource "aws_security_group_rule" "b_from_a" {
  type                     = "ingress"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.b.id
  source_security_group_id = aws_security_group.a.id
}
```

---

## 🟡 Pitfall 4: Blast Radius

### What Is Blast Radius?

```
Blast radius = How much damage can one wrong terraform apply do?

❌ Single state for everything:
   One mistake → ALL infrastructure affected
   Blast radius: ENTIRE COMPANY
   
   Resources: VPC + 50 EC2 + 10 RDS + 5 S3 + CloudFront + Route53 + ...
   terraform destroy → everything is gone

✅ Split by component:
   terraform/networking/    → VPC, subnets
   terraform/database/      → RDS instances  
   terraform/compute/       → ECS/EC2
   terraform/dns/           → Route53
   
   Mistake in compute? Only compute is affected.
   Database is in a different state file → untouched.
```

### Reducing Blast Radius

```
1. Separate state per environment AND component
   networking/dev/          networking/production/
   compute/dev/             compute/production/
   database/dev/            database/production/

2. Use prevent_destroy on critical resources
   resource "aws_rds_cluster" "main" {
     lifecycle { prevent_destroy = true }
   }

3. Review plan output carefully
   terraform plan | grep -E "destroy|replace"
   # If you see "destroy" on a database → STOP AND THINK

4. Use targeted applies for risky changes
   terraform apply -target=aws_instance.web
   # Only affects the specified resource
```

---

## 🟡 Pitfall 5: Secrets in Terraform

### The Problem

```hcl
# ❌ Secrets in .tf files (COMMITTED TO GIT!)
resource "aws_db_instance" "main" {
  username = "admin"
  password = "SuperSecretPassword123!"   # ← IN GIT HISTORY FOREVER
}

# ❌ Secrets in terraform.tfvars (often committed to Git)
db_password = "SuperSecretPassword123!"

# ❌ Secrets in state file (always — by design)
# Even if you use variables, the password ends up in state
```

### Solutions

```hcl
# ✅ Option 1: Environment variables
variable "db_password" {
  type      = string
  sensitive = true     # Hides from plan output
}

# Set before running terraform:
# export TF_VAR_db_password="secret"

# ✅ Option 2: AWS Secrets Manager / SSM Parameter Store
data "aws_secretsmanager_secret_version" "db_password" {
  secret_id = "production/database/password"
}

resource "aws_db_instance" "main" {
  password = data.aws_secretsmanager_secret_version.db_password.secret_string
}

# ✅ Option 3: Generate random password
resource "random_password" "db" {
  length  = 32
  special = true
}

resource "aws_db_instance" "main" {
  password = random_password.db.result
}

# Store in Secrets Manager for apps to retrieve
resource "aws_secretsmanager_secret_version" "db" {
  secret_id     = aws_secretsmanager_secret.db.id
  secret_string = random_password.db.result
}
```

**Remember: The password is STILL in the state file!**
**→ Encrypt state file, restrict access, use remote backend.**

---

## 🔴 Pitfall 6: Replacing Instead of Updating

### The Destroy-Then-Create Surprise

```bash
$ terraform plan

# aws_instance.web must be replaced
  -/+ resource "aws_instance" "web" {
      ~ ami = "ami-old" -> "ami-new"   # (forces replacement)
        ...
      }

# "forces replacement" = DESTROY old instance, CREATE new one
# This means DOWNTIME!
```

### Resources That Force Replacement

```
Changes that force resource REPLACEMENT (not update-in-place):
  ├── EC2: ami change
  ├── EC2: instance_type change (sometimes)
  ├── RDS: engine change
  ├── S3: bucket name change
  ├── VPC: cidr_block change
  └── Basically: any attribute that can't be changed on a live resource

Changes that update IN-PLACE (no downtime):
  ├── Tags
  ├── Security group associations
  ├── IAM policies
  └── Most "soft" configuration changes
```

### Prevention

```hcl
# create_before_destroy: new resource created BEFORE old one is destroyed
resource "aws_instance" "web" {
  lifecycle {
    create_before_destroy = true
  }
}

# Always check terraform plan for "must be replaced" before applying!
terraform plan | grep "must be replaced"
```

---

## 🔴 Pitfall 7: Provider Version Upgrades

```bash
# You update the AWS provider from 4.x to 5.x
# terraform plan now shows:
# 47 resources will be changed
# 12 resources will be replaced
# 3 resources will be destroyed

# What happened?
# Provider 5.x changed default values, deprecated arguments,
# or changed how resources behave.
```

### Prevention

```hcl
# Pin provider versions with ~> (allow only patch updates)
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.30"     # >= 5.30, < 5.31 (patch only)
    }
  }
}

# Use .terraform.lock.hcl (auto-generated, commit to Git)
# This locks the EXACT provider version across all team members
```

```bash
# When upgrading providers:
# 1. Read the changelog
# 2. Upgrade in dev first
# 3. Run terraform plan and review carefully
# 4. Apply to dev, then staging, then production
# 5. Never upgrade multiple providers at once
```

---

**Previous:** [05. Best Practices](./05-best-practices.md)
