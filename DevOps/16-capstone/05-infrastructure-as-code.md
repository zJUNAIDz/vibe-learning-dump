# Capstone Phase 05: Infrastructure as Code

> **If your infrastructure isn't in Git, it doesn't exist. Terraform defines every cloud resource your task-service needs — VPC, EKS cluster, RDS database, IAM roles — version controlled, reviewable, repeatable.**

---

## 🟢 Directory Structure

```
terraform/
├── modules/
│   ├── vpc/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── eks/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   └── rds/
│       ├── main.tf
│       ├── variables.tf
│       └── outputs.tf
├── environments/
│   ├── staging/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── terraform.tfvars
│   │   └── backend.tf
│   └── production/
│       ├── main.tf
│       ├── variables.tf
│       ├── terraform.tfvars
│       └── backend.tf
└── README.md
```

---

## 🟢 Backend Configuration (Remote State)

```hcl
# terraform/environments/staging/backend.tf
terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "myorg-terraform-state"
    key            = "task-service/staging/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-locks"  # State locking
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "task-service"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}
```

---

## 🟢 VPC Module

```hcl
# terraform/modules/vpc/main.tf
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "${var.project}-${var.environment}-vpc"
  }
}

# Public subnets (for load balancers)
resource "aws_subnet" "public" {
  count = length(var.availability_zones)

  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(var.vpc_cidr, 4, count.index)
  availability_zone       = var.availability_zones[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name                     = "${var.project}-${var.environment}-public-${count.index}"
    "kubernetes.io/role/elb" = "1"
  }
}

# Private subnets (for EKS nodes and RDS)
resource "aws_subnet" "private" {
  count = length(var.availability_zones)

  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 4, count.index + length(var.availability_zones))
  availability_zone = var.availability_zones[count.index]

  tags = {
    Name                              = "${var.project}-${var.environment}-private-${count.index}"
    "kubernetes.io/role/internal-elb" = "1"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
  tags   = { Name = "${var.project}-${var.environment}-igw" }
}

# NAT Gateway (private subnets need internet access for pulling images)
resource "aws_eip" "nat" {
  domain = "vpc"
  tags   = { Name = "${var.project}-${var.environment}-nat-eip" }
}

resource "aws_nat_gateway" "main" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public[0].id
  tags          = { Name = "${var.project}-${var.environment}-nat" }
}

# Route tables
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = { Name = "${var.project}-${var.environment}-public-rt" }
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.main.id
  }

  tags = { Name = "${var.project}-${var.environment}-private-rt" }
}

resource "aws_route_table_association" "public" {
  count          = length(aws_subnet.public)
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "private" {
  count          = length(aws_subnet.private)
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private.id
}
```

```hcl
# terraform/modules/vpc/variables.tf
variable "project" {
  type = string
}

variable "environment" {
  type = string
}

variable "vpc_cidr" {
  type    = string
  default = "10.0.0.0/16"
}

variable "availability_zones" {
  type = list(string)
}
```

```hcl
# terraform/modules/vpc/outputs.tf
output "vpc_id" {
  value = aws_vpc.main.id
}

output "public_subnet_ids" {
  value = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  value = aws_subnet.private[*].id
}
```

---

## 🟡 EKS Module

```hcl
# terraform/modules/eks/main.tf
resource "aws_eks_cluster" "main" {
  name     = "${var.project}-${var.environment}"
  role_arn = aws_iam_role.cluster.arn
  version  = var.kubernetes_version

  vpc_config {
    subnet_ids              = var.private_subnet_ids
    endpoint_private_access = true
    endpoint_public_access  = true
    security_group_ids      = [aws_security_group.cluster.id]
  }

  # Enable logging
  enabled_cluster_log_types = [
    "api",
    "audit",
    "authenticator",
    "controllerManager",
    "scheduler",
  ]

  depends_on = [
    aws_iam_role_policy_attachment.cluster_policy,
    aws_iam_role_policy_attachment.service_policy,
  ]
}

# Managed node group
resource "aws_eks_node_group" "main" {
  cluster_name    = aws_eks_cluster.main.name
  node_group_name = "${var.project}-${var.environment}-nodes"
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = var.private_subnet_ids

  instance_types = var.node_instance_types
  capacity_type  = var.environment == "production" ? "ON_DEMAND" : "SPOT"

  scaling_config {
    desired_size = var.node_desired_size
    min_size     = var.node_min_size
    max_size     = var.node_max_size
  }

  update_config {
    max_unavailable = 1
  }

  labels = {
    environment = var.environment
    project     = var.project
  }

  depends_on = [
    aws_iam_role_policy_attachment.node_policy,
    aws_iam_role_policy_attachment.cni_policy,
    aws_iam_role_policy_attachment.ecr_policy,
  ]
}

# ─── IAM Roles ───────────────────────────────

resource "aws_iam_role" "cluster" {
  name = "${var.project}-${var.environment}-eks-cluster"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "eks.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "cluster_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_iam_role_policy_attachment" "service_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSServicePolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_iam_role" "node" {
  name = "${var.project}-${var.environment}-eks-node"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "node_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.node.name
}

resource "aws_iam_role_policy_attachment" "cni_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.node.name
}

resource "aws_iam_role_policy_attachment" "ecr_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.node.name
}

# Security group
resource "aws_security_group" "cluster" {
  name_prefix = "${var.project}-${var.environment}-eks-"
  vpc_id      = var.vpc_id

  ingress {
    from_port = 443
    to_port   = 443
    protocol  = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  lifecycle {
    create_before_destroy = true
  }
}

# IRSA — OIDC provider for service account IAM roles
data "tls_certificate" "cluster" {
  url = aws_eks_cluster.main.identity[0].oidc[0].issuer
}

resource "aws_iam_openid_connect_provider" "cluster" {
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.cluster.certificates[0].sha1_fingerprint]
  url             = aws_eks_cluster.main.identity[0].oidc[0].issuer
}
```

```hcl
# terraform/modules/eks/variables.tf
variable "project" { type = string }
variable "environment" { type = string }
variable "vpc_id" { type = string }
variable "private_subnet_ids" { type = list(string) }
variable "kubernetes_version" {
  type    = string
  default = "1.29"
}
variable "node_instance_types" {
  type    = list(string)
  default = ["t3.medium"]
}
variable "node_desired_size" {
  type    = number
  default = 2
}
variable "node_min_size" {
  type    = number
  default = 1
}
variable "node_max_size" {
  type    = number
  default = 5
}
```

```hcl
# terraform/modules/eks/outputs.tf
output "cluster_endpoint" {
  value = aws_eks_cluster.main.endpoint
}

output "cluster_name" {
  value = aws_eks_cluster.main.name
}

output "cluster_ca_certificate" {
  value = aws_eks_cluster.main.certificate_authority[0].data
}

output "oidc_provider_arn" {
  value = aws_iam_openid_connect_provider.cluster.arn
}
```

---

## 🟡 RDS Module

```hcl
# terraform/modules/rds/main.tf
resource "aws_db_subnet_group" "main" {
  name       = "${var.project}-${var.environment}"
  subnet_ids = var.private_subnet_ids

  tags = { Name = "${var.project}-${var.environment}-db-subnet" }
}

resource "aws_security_group" "rds" {
  name_prefix = "${var.project}-${var.environment}-rds-"
  vpc_id      = var.vpc_id

  # Only allow traffic from EKS nodes
  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [var.eks_security_group_id]
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_db_instance" "main" {
  identifier = "${var.project}-${var.environment}"

  engine         = "postgres"
  engine_version = "16.1"
  instance_class = var.instance_class

  allocated_storage     = var.allocated_storage
  max_allocated_storage = var.max_allocated_storage
  storage_encrypted     = true

  db_name  = "tasks"
  username = "taskadmin"
  password = var.db_password  # Pass via tfvars or Secrets Manager

  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]

  multi_az            = var.environment == "production"
  skip_final_snapshot = var.environment != "production"

  backup_retention_period = var.environment == "production" ? 7 : 1
  deletion_protection     = var.environment == "production"

  performance_insights_enabled = true

  tags = { Name = "${var.project}-${var.environment}-db" }
}
```

```hcl
# terraform/modules/rds/variables.tf
variable "project" { type = string }
variable "environment" { type = string }
variable "vpc_id" { type = string }
variable "private_subnet_ids" { type = list(string) }
variable "eks_security_group_id" { type = string }
variable "db_password" {
  type      = string
  sensitive = true
}
variable "instance_class" {
  type    = string
  default = "db.t3.micro"
}
variable "allocated_storage" {
  type    = number
  default = 20
}
variable "max_allocated_storage" {
  type    = number
  default = 100
}
```

```hcl
# terraform/modules/rds/outputs.tf
output "endpoint" {
  value = aws_db_instance.main.endpoint
}

output "database_name" {
  value = aws_db_instance.main.db_name
}
```

---

## 🟡 Environment Composition

```hcl
# terraform/environments/staging/main.tf
module "vpc" {
  source = "../../modules/vpc"

  project            = var.project
  environment        = var.environment
  vpc_cidr           = "10.0.0.0/16"
  availability_zones = var.availability_zones
}

module "eks" {
  source = "../../modules/eks"

  project             = var.project
  environment         = var.environment
  vpc_id              = module.vpc.vpc_id
  private_subnet_ids  = module.vpc.private_subnet_ids
  kubernetes_version  = "1.29"
  node_instance_types = ["t3.medium"]
  node_desired_size   = 2
  node_min_size       = 1
  node_max_size       = 4
}

module "rds" {
  source = "../../modules/rds"

  project               = var.project
  environment           = var.environment
  vpc_id                = module.vpc.vpc_id
  private_subnet_ids    = module.vpc.private_subnet_ids
  eks_security_group_id = module.eks.cluster_security_group_id
  db_password           = var.db_password
  instance_class        = "db.t3.micro"
}
```

```hcl
# terraform/environments/staging/variables.tf
variable "project" {
  type    = string
  default = "task-service"
}

variable "environment" {
  type    = string
  default = "staging"
}

variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "availability_zones" {
  type    = list(string)
  default = ["us-east-1a", "us-east-1b"]
}

variable "db_password" {
  type      = string
  sensitive = true
}
```

```hcl
# terraform/environments/staging/terraform.tfvars
project            = "task-service"
environment        = "staging"
aws_region         = "us-east-1"
availability_zones = ["us-east-1a", "us-east-1b"]
# db_password provided via TF_VAR_db_password or -var flag
```

---

## 🔴 Production vs Staging Differences

```hcl
# terraform/environments/production/main.tf
# Same module calls, different parameters:

module "eks" {
  source = "../../modules/eks"
  # ...
  node_instance_types = ["t3.large"]     # Bigger nodes
  node_desired_size   = 3                 # More nodes
  node_min_size       = 3                 # Never fewer than 3
  node_max_size       = 10                # Scale higher
}

module "rds" {
  source = "../../modules/rds"
  # ...
  instance_class        = "db.r6g.large" # Production-grade
  allocated_storage     = 100
  max_allocated_storage = 500
  # multi_az and deletion_protection auto-enabled
  # via environment == "production" checks in module
}
```

| Resource | Staging | Production |
|---|---|---|
| EKS nodes | t3.medium, 2 nodes, SPOT | t3.large, 3 nodes, ON_DEMAND |
| RDS | db.t3.micro, single-AZ, no delete protection | db.r6g.large, multi-AZ, delete protection |
| Backups | 1 day retention | 7 day retention |
| VPC | 2 AZs | 3 AZs |

---

## 🔴 Workflow

```bash
# Initialize
cd terraform/environments/staging
terraform init

# Plan — always review before applying
terraform plan -var="db_password=$DB_PASSWORD" -out=plan.tfplan

# Apply
terraform apply plan.tfplan

# Check outputs
terraform output cluster_endpoint
terraform output rds_endpoint

# Connect kubectl to the new cluster
aws eks update-kubeconfig \
  --name task-service-staging \
  --region us-east-1

# Verify
kubectl get nodes
```

---

## 🔴 Checklist

```
□ Remote state in S3 with DynamoDB locking
□ State encryption enabled
□ Modular structure (vpc, eks, rds as separate modules)
□ Environments separated (staging/ and production/ directories)
□ Sensitive values marked as sensitive
□ Database password NOT in terraform.tfvars (use env var or secrets manager)
□ Production has multi-AZ, delete protection, longer backups
□ Staging uses spot instances to save costs
□ All resources tagged with Project, Environment, ManagedBy
□ OIDC provider configured for IRSA
□ terraform plan reviewed before every apply
□ .terraform/ and *.tfstate in .gitignore
```

---

**Previous:** [04. CI/CD Pipeline](./04-cicd-pipeline.md)  
**Next:** [06. Observability](./06-observability.md)
