# Terraform Fundamentals

> **Providers, resources, data sources, variables, outputs — these are the five building blocks of every Terraform project.**

---

## 🟢 Providers

A **provider** is a plugin that lets Terraform talk to a specific API (AWS, GCP, Kubernetes, GitHub, etc.)

```hcl
# Providers tell Terraform WHERE to create resources

# AWS Provider
provider "aws" {
  region  = "us-east-1"
  profile = "production"       # AWS CLI profile
}

# Multiple AWS providers (multi-region)
provider "aws" {
  alias  = "us_west"
  region = "us-west-2"
}

# Kubernetes Provider
provider "kubernetes" {
  config_path = "~/.kube/config"
  context     = "production-cluster"
}

# Docker Provider
provider "docker" {
  host = "unix:///var/run/docker.sock"
}
```

### Provider Versioning

```hcl
terraform {
  required_version = ">= 1.7.0"    # Terraform itself

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"            # >= 5.0, < 6.0
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.25.0"
    }
  }
}
```

**Version constraints:**
```
"= 5.0.0"     # Exactly 5.0.0
">= 5.0.0"    # 5.0.0 or newer
"~> 5.0"       # >= 5.0, < 6.0 (recommended — minor updates only)
"~> 5.0.0"     # >= 5.0.0, < 5.1.0 (patch updates only)
">= 5.0, < 6.0" # Same as ~> 5.0
```

---

## 🟢 Resources

Resources are the **things Terraform creates** in the real world.

```hcl
# Syntax: resource "TYPE" "NAME" { ... }

# Create a VPC
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "production-vpc"
    Environment = "production"
  }
}

# Create a subnet (references the VPC above)
resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id        # Reference another resource
  cidr_block              = "10.0.1.0/24"
  availability_zone       = "us-east-1a"
  map_public_ip_on_launch = true

  tags = {
    Name = "public-subnet-1a"
  }
}

# Create an EC2 instance
resource "aws_instance" "web" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t3.small"
  subnet_id     = aws_subnet.public.id

  root_block_device {
    volume_size = 20
    volume_type = "gp3"
    encrypted   = true
  }

  tags = {
    Name = "web-server"
  }
}
```

### Resource Dependencies

```hcl
# Terraform automatically detects IMPLICIT dependencies
resource "aws_subnet" "public" {
  vpc_id = aws_vpc.main.id          # ← Terraform knows: create VPC first
}

# Sometimes you need EXPLICIT dependencies
resource "aws_instance" "web" {
  ami           = "ami-12345678"
  instance_type = "t3.small"

  depends_on = [aws_iam_role_policy.web_policy]
  # ↑ Even though there's no direct reference,
  #   the instance needs the IAM policy to exist first
}
```

### Resource Lifecycle

```hcl
resource "aws_instance" "web" {
  ami           = "ami-12345678"
  instance_type = "t3.small"

  lifecycle {
    create_before_destroy = true    # Create new instance before destroying old one
    prevent_destroy       = true    # Prevent accidental deletion
    ignore_changes        = [tags]  # Don't update if only tags changed
    
    # Replace when any of these change
    replace_triggered_by = [
      aws_security_group.web.id
    ]
  }
}
```

---

## 🟢 Data Sources

Data sources let you **read** information from existing infrastructure (without creating anything).

```hcl
# "What's the latest Amazon Linux 2 AMI?"
data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}

# Use it in a resource
resource "aws_instance" "web" {
  ami           = data.aws_ami.amazon_linux.id    # ← From data source
  instance_type = "t3.small"
}
```

```hcl
# "What are the available AZs in the current region?"
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_subnet" "public" {
  count             = length(data.aws_availability_zones.available.names)
  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet("10.0.0.0/16", 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
}
```

```hcl
# "What's in the Terraform state of another project?"
data "terraform_remote_state" "vpc" {
  backend = "s3"
  config = {
    bucket = "my-terraform-state"
    key    = "networking/terraform.tfstate"
    region = "us-east-1"
  }
}

resource "aws_instance" "web" {
  subnet_id = data.terraform_remote_state.vpc.outputs.public_subnet_id
}
```

---

## 🟢 Variables

### Input Variables

```hcl
# variables.tf

variable "environment" {
  description = "Deployment environment (dev, staging, production)"
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "staging", "production"], var.environment)
    error_message = "Environment must be dev, staging, or production."
  }
}

variable "instance_count" {
  description = "Number of EC2 instances to create"
  type        = number
  default     = 1
}

variable "enable_monitoring" {
  description = "Enable detailed monitoring"
  type        = bool
  default     = false
}

variable "allowed_cidrs" {
  description = "CIDR blocks allowed to access the service"
  type        = list(string)
  default     = ["10.0.0.0/8"]
}

variable "instance_config" {
  description = "Instance configuration"
  type = object({
    instance_type = string
    volume_size   = number
    volume_type   = string
  })
  default = {
    instance_type = "t3.small"
    volume_size   = 20
    volume_type   = "gp3"
  }
}

variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}

# Sensitive variable (value hidden in logs)
variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}
```

### Setting Variable Values

```bash
# 1. Command line
terraform apply -var="environment=production" -var="instance_count=5"

# 2. Variable file (recommended)
terraform apply -var-file="environments/production.tfvars"

# 3. Environment variables
export TF_VAR_environment="production"
export TF_VAR_db_password="secret123"
terraform apply

# 4. Auto-loaded files (terraform.tfvars or *.auto.tfvars)
# These are loaded automatically without -var-file flag
```

```hcl
# environments/production.tfvars
environment       = "production"
instance_count    = 5
enable_monitoring = true
allowed_cidrs     = ["10.0.0.0/8", "172.16.0.0/12"]

instance_config = {
  instance_type = "t3.large"
  volume_size   = 100
  volume_type   = "gp3"
}

tags = {
  Environment = "production"
  Team        = "platform"
  CostCenter  = "12345"
}
```

---

## 🟢 Outputs

Outputs **expose values** from your Terraform configuration.

```hcl
# outputs.tf

output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.main.id
}

output "public_subnet_ids" {
  description = "IDs of public subnets"
  value       = aws_subnet.public[*].id
}

output "web_server_ip" {
  description = "Public IP of the web server"
  value       = aws_instance.web.public_ip
}

output "database_endpoint" {
  description = "Database connection endpoint"
  value       = aws_db_instance.main.endpoint
  sensitive   = true          # Hidden in console output
}
```

```bash
# View outputs after apply
terraform output
# vpc_id          = "vpc-abc123"
# public_subnets  = ["subnet-111", "subnet-222"]
# web_server_ip   = "54.123.45.67"

# Get a specific output (useful in scripts)
terraform output -raw web_server_ip
# 54.123.45.67

# JSON format (useful for automation)
terraform output -json
```

---

## 🟡 The Terraform Workflow

```
1. terraform init
   │  Download providers
   │  Initialize backend
   │  Download modules
   ▼
2. terraform plan
   │  Read current state
   │  Compare with desired state
   │  Show what will change
   │  NO CHANGES MADE YET
   ▼
3. terraform apply
   │  Show plan again
   │  Ask for confirmation
   │  Make the changes
   │  Update state file
   ▼
4. terraform destroy
      Show what will be destroyed
      Ask for confirmation
      Delete all resources
      Clear state
```

```bash
# Initialize (run once, or when providers/modules change)
terraform init

# Preview changes (ALWAYS do this before apply)
terraform plan
# + aws_instance.web will be created
# ~ aws_security_group.web will be updated in-place
# - aws_instance.old will be destroyed

# Apply changes
terraform apply
# Do you want to perform these actions? yes

# Save plan to file (recommended for CI/CD)
terraform plan -out=tfplan
terraform apply tfplan       # No confirmation needed — plan was already reviewed

# Destroy everything
terraform destroy
```

---

## 🟡 Locals (Computed Values)

```hcl
locals {
  # Computed from variables
  name_prefix = "${var.environment}-${var.app_name}"
  
  # Common tags applied to all resources
  common_tags = merge(var.tags, {
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = var.app_name
  })
  
  # Conditional values
  instance_type = var.environment == "production" ? "t3.large" : "t3.small"
  
  # Computed from data sources
  az_count = length(data.aws_availability_zones.available.names)
}

resource "aws_instance" "web" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = local.instance_type        # ← Using local value
  
  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-web"
  })
}
```

---

**Previous:** [README](./README.md)  
**Next:** [02. State Deep Dive](./02-state-deep-dive.md)
