# Terraform Best Practices

> **Terraform is easy to start and hard to manage at scale. These practices prevent the pain that hits at month 6.**

---

## 🟢 Directory Structure

```
# Recommended structure for a real project

terraform/
├── modules/                           # Reusable modules
│   ├── vpc/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── README.md
│   ├── ecs-service/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── README.md
│   └── rds/
│       ├── main.tf
│       ├── variables.tf
│       ├── outputs.tf
│       └── README.md
├── environments/
│   ├── dev/
│   │   ├── main.tf                    # Calls modules
│   │   ├── backend.tf                 # S3 backend config
│   │   ├── providers.tf               # Provider config
│   │   ├── variables.tf               # Variable declarations
│   │   ├── terraform.tfvars           # Variable values
│   │   └── outputs.tf
│   ├── staging/
│   │   └── ... (same structure)
│   └── production/
│       └── ... (same structure)
├── .terraform-version                 # Pin Terraform version (tfenv)
├── .tflint.hcl                        # Linting config
└── Makefile                           # Common commands
```

---

## 🟢 Naming Conventions

```hcl
# Resources: lowercase, underscores, descriptive
resource "aws_instance" "web_server" { }          # ✅
resource "aws_instance" "WebServer" { }           # ❌ No PascalCase
resource "aws_instance" "web-server" { }          # ❌ No hyphens
resource "aws_instance" "instance1" { }           # ❌ Not descriptive

# Variables: lowercase, underscores
variable "instance_type" { }                      # ✅
variable "instanceType" { }                       # ❌ No camelCase

# Outputs: lowercase, underscores
output "vpc_id" { }                               # ✅

# Modules: lowercase, hyphens in directory names
module "web_server" {                             # ✅
  source = "./modules/web-server"
}

# Tags: consistent across all resources
locals {
  common_tags = {
    Environment = var.environment
    Project     = var.project_name
    Team        = var.team
    ManagedBy   = "terraform"
  }
}
```

---

## 🟢 Code Organization

### One Responsibility per File

```
# ✅ Organized by purpose
main.tf           # Primary resources
networking.tf     # VPC, subnets, security groups
compute.tf        # EC2, ECS, Lambda
database.tf       # RDS, DynamoDB
monitoring.tf     # CloudWatch, alerts
iam.tf            # IAM roles, policies
variables.tf      # ALL variable declarations
outputs.tf        # ALL outputs
providers.tf      # Provider configuration
backend.tf        # Backend configuration
locals.tf         # Local values

# ❌ One giant main.tf with 2000 lines
# ❌ Random file names (stuff.tf, misc.tf, new.tf)
```

### Keep It Simple

```hcl
# ❌ Over-engineered: dynamic blocks, complex expressions
resource "aws_security_group" "this" {
  dynamic "ingress" {
    for_each = { for k, v in var.ingress_rules : k => v if lookup(v, "enabled", true) }
    content {
      from_port   = ingress.value.from_port
      to_port     = ingress.value.to_port
      protocol    = lookup(ingress.value, "protocol", "tcp")
      cidr_blocks = lookup(ingress.value, "cidr_blocks", [])
    }
  }
}

# ✅ Simple and readable
resource "aws_security_group" "web" {
  name   = "${var.name}-web-sg"
  vpc_id = var.vpc_id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

---

## 🟢 When NOT to Use Terraform

```
❌ Don't use Terraform for:

1. Application deployment
   Use: kubectl, Helm, ArgoCD
   Why: Terraform is for infrastructure, not app config

2. Kubernetes resource management (usually)
   Use: kubectl + YAML, Helm, Kustomize
   Why: K8s has its own declarative model that handles state better

3. One-time scripts
   Use: AWS CLI, bash scripts
   Why: No need for state management for throwaway tasks

4. Frequently changing configuration
   Use: Ansible, Consul, environment variables
   Why: Terraform plan/apply is slow for frequent changes

5. Manual experimentation
   Use: AWS Console, CLI
   Why: Terraform is for production, not learning
```

### Terraform + Kubernetes: When It Makes Sense

```hcl
# ✅ YES: Create the Kubernetes CLUSTER with Terraform
resource "aws_eks_cluster" "main" {
  name     = "production"
  role_arn = aws_iam_role.cluster.arn
  version  = "1.29"

  vpc_config {
    subnet_ids = module.vpc.private_subnet_ids
  }
}

# ✅ YES: Create CLUSTER-LEVEL resources (namespaces, RBAC)
resource "kubernetes_namespace" "app" {
  metadata {
    name = "my-app"
  }
}

# ❌ NO: Don't manage individual deployments/pods with Terraform
# Use Helm or ArgoCD instead
resource "kubernetes_deployment" "app" {
  # This is technically possible but painful
  # Every code change = terraform apply
  # No rolling updates, no canary, no rollback
}
```

---

## 🟡 Terraform with Makefile

```makefile
# Makefile — Common Terraform commands

ENV ?= dev

.PHONY: init plan apply destroy fmt validate lint

init:
	cd environments/$(ENV) && terraform init

plan:
	cd environments/$(ENV) && terraform plan -out=tfplan

apply:
	cd environments/$(ENV) && terraform apply tfplan

destroy:
	cd environments/$(ENV) && terraform destroy

fmt:
	terraform fmt -recursive

validate:
	cd environments/$(ENV) && terraform validate

lint:
	tflint --recursive

# Safety checks
plan-production:
	@echo "⚠️  Planning PRODUCTION changes"
	@echo "Press Ctrl+C to cancel..."
	@sleep 5
	cd environments/production && terraform plan -out=tfplan

apply-production:
	@echo "🚨 Applying to PRODUCTION 🚨"
	@echo "Press Ctrl+C to cancel..."
	@sleep 10
	cd environments/production && terraform apply tfplan
```

```bash
# Usage
make plan ENV=dev
make apply ENV=dev
make plan-production     # Extra safety for prod
```

---

## 🟡 Error Prevention

### terraform fmt

```bash
# Auto-format ALL .tf files
terraform fmt -recursive

# Check formatting (CI — fail if not formatted)
terraform fmt -check -recursive
# Exit code 0 = all formatted
# Exit code 3 = files need formatting
```

### terraform validate

```bash
# Check syntax and configuration validity
terraform validate
# Success! The configuration is valid.
```

### tflint

```bash
# Install tflint
brew install tflint   # Mac
# or: curl -s https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh | bash

# Run linter
tflint --recursive
```

```hcl
# .tflint.hcl
plugin "aws" {
  enabled = true
  version = "0.29.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

rule "terraform_naming_convention" {
  enabled = true
}

rule "terraform_documented_variables" {
  enabled = true
}
```

### Pre-commit Hooks

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/antonbabenko/pre-commit-terraform
    rev: v1.86.0
    hooks:
      - id: terraform_fmt
      - id: terraform_validate
      - id: terraform_tflint
      - id: terraform_docs
```

---

## 🔴 Common Pitfalls

### Pitfall 1: Destroying Everything Accidentally

```bash
# The "Oops" command:
terraform destroy -auto-approve
# Gone. Everything. In seconds.

# Prevention:
# 1. Use lifecycle { prevent_destroy = true } on critical resources
resource "aws_rds_cluster" "main" {
  lifecycle {
    prevent_destroy = true
  }
}

# 2. Never use -auto-approve for production
# 3. Always review the plan
# 4. Use CI/CD with approval gates
```

### Pitfall 2: Massive Blast Radius

```
❌ One state file for ALL infrastructure:
   "I changed a variable and it's destroying 47 resources"
   
✅ Split into smaller, independent projects:
   terraform/networking/    → VPC, subnets (rarely changes)
   terraform/database/      → RDS (critical, rarely changes)
   terraform/compute/       → ECS services (changes frequently)
   terraform/monitoring/    → CloudWatch (changes moderately)
   
   Each with its OWN state file
   Blast radius: limited to one component
```

### Pitfall 3: Manual Changes

```
Someone changes a security group in the AWS console.
terraform plan: "I see drift — let me revert your change."

Prevention:
1. Lock down console access (read-only for most people)
2. Run terraform plan regularly to detect drift
3. Culture: "If it's not in Terraform, it doesn't exist"
```

---

**Previous:** [04. Workspaces and Environments](./04-workspaces-and-environments.md)  
**Next:** [06. Common Pitfalls](./06-common-pitfalls.md)
