# Terraform Modules

> **Modules are the packages of Terraform. Write once, use everywhere — across environments, teams, and projects.**

---

## 🟢 What Is a Module?

A module is a **directory of `.tf` files** that's used as a unit.

```
Every Terraform project is a module:
  - The root directory = "root module"
  - Any directory you call with `module` block = "child module"
```

```
project/
├── main.tf              ← Root module
├── variables.tf
├── outputs.tf
└── modules/
    ├── vpc/             ← Child module
    │   ├── main.tf
    │   ├── variables.tf
    │   └── outputs.tf
    ├── compute/         ← Child module
    │   ├── main.tf
    │   ├── variables.tf
    │   └── outputs.tf
    └── database/        ← Child module
        ├── main.tf
        ├── variables.tf
        └── outputs.tf
```

---

## 🟢 Creating a Module

### VPC Module Example

```hcl
# modules/vpc/variables.tf
variable "name" {
  description = "Name prefix for VPC resources"
  type        = string
}

variable "cidr_block" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of AZs to use"
  type        = list(string)
}

variable "environment" {
  description = "Environment name"
  type        = string
}
```

```hcl
# modules/vpc/main.tf
resource "aws_vpc" "this" {
  cidr_block           = var.cidr_block
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "${var.name}-vpc"
    Environment = var.environment
  }
}

resource "aws_subnet" "public" {
  count             = length(var.availability_zones)
  vpc_id            = aws_vpc.this.id
  cidr_block        = cidrsubnet(var.cidr_block, 8, count.index)
  availability_zone = var.availability_zones[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.name}-public-${var.availability_zones[count.index]}"
  }
}

resource "aws_subnet" "private" {
  count             = length(var.availability_zones)
  vpc_id            = aws_vpc.this.id
  cidr_block        = cidrsubnet(var.cidr_block, 8, count.index + length(var.availability_zones))
  availability_zone = var.availability_zones[count.index]

  tags = {
    Name = "${var.name}-private-${var.availability_zones[count.index]}"
  }
}

resource "aws_internet_gateway" "this" {
  vpc_id = aws_vpc.this.id

  tags = {
    Name = "${var.name}-igw"
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.this.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.this.id
  }

  tags = {
    Name = "${var.name}-public-rt"
  }
}

resource "aws_route_table_association" "public" {
  count          = length(var.availability_zones)
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}
```

```hcl
# modules/vpc/outputs.tf
output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.this.id
}

output "public_subnet_ids" {
  description = "IDs of public subnets"
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "IDs of private subnets"
  value       = aws_subnet.private[*].id
}

output "cidr_block" {
  description = "CIDR block of the VPC"
  value       = aws_vpc.this.cidr_block
}
```

---

## 🟢 Using a Module

```hcl
# main.tf (root module)

module "vpc" {
  source = "./modules/vpc"       # Local path

  name               = "production"
  cidr_block         = "10.0.0.0/16"
  availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
  environment        = "production"
}

module "web_servers" {
  source = "./modules/compute"

  name          = "web"
  vpc_id        = module.vpc.vpc_id              # ← Reference module output
  subnet_ids    = module.vpc.public_subnet_ids   # ← Reference module output
  instance_type = "t3.small"
  instance_count = 3
}

module "database" {
  source = "./modules/database"

  name             = "main-db"
  vpc_id           = module.vpc.vpc_id
  subnet_ids       = module.vpc.private_subnet_ids
  instance_class   = "db.r6g.large"
  engine           = "postgresql"
  engine_version   = "15.4"
}
```

### Module Sources

```hcl
# Local path
module "vpc" {
  source = "./modules/vpc"
}

# Git repository
module "vpc" {
  source = "git::https://github.com/myorg/terraform-modules.git//vpc?ref=v1.2.0"
}

# Terraform Registry (public modules)
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.4.0"
}

# S3 bucket
module "vpc" {
  source = "s3::https://s3-eu-west-1.amazonaws.com/my-modules/vpc.zip"
}
```

---

## 🟡 Module Best Practices

### 1. Module Interface Design

```hcl
# ❌ Bad: Too many required variables, user must know internal details
module "vpc" {
  source                  = "./modules/vpc"
  vpc_cidr                = "10.0.0.0/16"
  public_subnet_cidrs     = ["10.0.1.0/24", "10.0.2.0/24"]
  private_subnet_cidrs    = ["10.0.101.0/24", "10.0.102.0/24"]
  nat_gateway_count       = 2
  enable_dns_hostnames    = true
  enable_dns_support      = true
  route_table_count       = 4
  igw_route_table_id      = "..."
  # ... 20 more variables
}

# ✅ Good: Sensible defaults, user provides only what matters
module "vpc" {
  source = "./modules/vpc"
  
  name               = "production"
  environment        = "production"
  availability_zones = ["us-east-1a", "us-east-1b"]
  # Everything else has sensible defaults
}
```

### 2. README for Every Module

```markdown
# VPC Module

Creates a VPC with public and private subnets across multiple AZs.

## Usage

```hcl
module "vpc" {
  source = "./modules/vpc"
  
  name               = "production"
  availability_zones = ["us-east-1a", "us-east-1b"]
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| name | Name prefix | string | - | yes |
| cidr_block | VPC CIDR | string | "10.0.0.0/16" | no |

## Outputs

| Name | Description |
|------|-------------|
| vpc_id | ID of the VPC |
| public_subnet_ids | IDs of public subnets |
```

### 3. Version Your Modules

```hcl
# ❌ Using latest (might break when module changes)
module "vpc" {
  source = "git::https://github.com/myorg/modules.git//vpc"
}

# ✅ Pin to a version
module "vpc" {
  source = "git::https://github.com/myorg/modules.git//vpc?ref=v1.2.0"
}

# ✅ Pin registry modules
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"    # >= 5.0, < 6.0
}
```

---

## 🟡 for_each and count in Modules

### Using count

```hcl
# Create multiple instances of a module
module "web_server" {
  count  = var.environment == "production" ? 3 : 1
  source = "./modules/compute"

  name          = "web-${count.index + 1}"
  instance_type = "t3.small"
  subnet_id     = module.vpc.public_subnet_ids[count.index % length(module.vpc.public_subnet_ids)]
}
```

### Using for_each

```hcl
# More powerful — use a map
variable "services" {
  default = {
    api = {
      instance_type = "t3.medium"
      port          = 8080
    }
    worker = {
      instance_type = "t3.small"
      port          = 9090
    }
    scheduler = {
      instance_type = "t3.micro"
      port          = 9091
    }
  }
}

module "service" {
  for_each = var.services
  source   = "./modules/compute"

  name          = each.key                      # "api", "worker", "scheduler"
  instance_type = each.value.instance_type
  port          = each.value.port
  subnet_id     = module.vpc.private_subnet_ids[0]
}

# Access: module.service["api"].instance_id
# Access: module.service["worker"].private_ip
```

---

## 🔴 Module Anti-Patterns

### Anti-Pattern 1: God Module

```hcl
# ❌ One module that creates EVERYTHING
module "infrastructure" {
  source = "./modules/everything"
  # Creates VPC, subnets, instances, databases, load balancers,
  # DNS, certificates, monitoring, alerting...
  # 2000 lines of code, impossible to understand
}

# ✅ Small, focused modules
module "networking"  { source = "./modules/vpc" }
module "compute"     { source = "./modules/compute" }
module "database"    { source = "./modules/database" }
module "monitoring"  { source = "./modules/monitoring" }
```

### Anti-Pattern 2: Deep Nesting

```
# ❌ Modules calling modules calling modules
module "app"
  └── module "infra"
       └── module "network"
            └── module "vpc"
                 └── module "subnets"
                      └── module "route_tables"

# Debugging this is a nightmare
# State paths: module.app.module.infra.module.network.module.vpc.aws_vpc.this

# ✅ Maximum 2 levels deep
module "vpc"      (calls no other modules)
module "compute"  (calls no other modules)
module "app"      (calls vpc and compute modules — 1 level deep)
```

### Anti-Pattern 3: Expose All Attributes

```hcl
# ❌ Exposing internal details
output "vpc_dhcp_options_id" { value = aws_vpc.this.dhcp_options_id }
output "vpc_main_route_table_id" { value = aws_vpc.this.main_route_table_id }
output "vpc_default_network_acl_id" { value = aws_vpc.this.default_network_acl_id }
# ... 20 more internal attributes nobody needs

# ✅ Expose only what consumers need
output "vpc_id" { value = aws_vpc.this.id }
output "public_subnet_ids" { value = aws_subnet.public[*].id }
output "private_subnet_ids" { value = aws_subnet.private[*].id }
```

---

**Previous:** [02. State Deep Dive](./02-state-deep-dive.md)  
**Next:** [04. Workspaces and Environments](./04-workspaces-and-environments.md)
