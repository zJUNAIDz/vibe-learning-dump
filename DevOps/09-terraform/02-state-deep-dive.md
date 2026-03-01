# Terraform State Deep Dive

> **State is where Terraform keeps its memory. Lose it, corrupt it, or mishanage it, and you're in for a very bad day.**

---

## 🟢 What's in the State File?

State maps every resource in your code to a real-world object.

```json
// terraform.tfstate (simplified)
{
  "version": 4,
  "terraform_version": "1.7.0",
  "serial": 15,              // Incremented on every change
  "lineage": "abc123...",    // Unique ID — identifies this state "lineage"
  "outputs": {
    "vpc_id": {
      "value": "vpc-abc123def",
      "type": "string"
    }
  },
  "resources": [
    {
      "mode": "managed",              // managed = Terraform created it
      "type": "aws_instance",         // AWS resource type
      "name": "web",                  // Name in your .tf file
      "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
      "instances": [
        {
          "schema_version": 1,
          "attributes": {
            "id": "i-0abc123def456789",     // Real AWS resource ID
            "ami": "ami-12345678",
            "instance_type": "t3.small",
            "private_ip": "10.0.1.15",
            "public_ip": "54.123.45.67",
            "tags": {
              "Name": "web-server"
            }
            // ... many more attributes
          },
          "private": "base64encodeddata...",
          "dependencies": [
            "aws_subnet.public",
            "aws_security_group.web"
          ]
        }
      ]
    }
  ]
}
```

### Why State Contains Sensitive Data

```
State includes EVERY attribute of EVERY resource:
  ├── Database password (plaintext)
  ├── API keys
  ├── Private IP addresses
  ├── TLS certificate private keys
  ├── SSH key contents
  └── Any secret passed as a variable
  
This is NOT a design flaw — Terraform NEEDS this info
to compare desired state with actual state.

But it means: PROTECT YOUR STATE FILE LIKE A SECRET.
```

---

## 🟢 Local vs Remote State

### Local State

```bash
$ ls -la
-rw-r--r-- main.tf
-rw-r--r-- terraform.tfstate         # 📂 On your laptop
-rw-r--r-- terraform.tfstate.backup  # 📂 Previous version

# Problems:
# 1. Not shared — only you have it
# 2. Not backed up — laptop dies = state lost
# 3. Not locked — two people can modify simultaneously
# 4. Secrets in plaintext on disk
```

### Remote State — S3 Backend

```hcl
terraform {
  backend "s3" {
    bucket         = "mycompany-terraform-state"
    key            = "production/web/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true                         # SSE-S3 encryption
    dynamodb_table = "terraform-state-locks"      # Locking
    
    # Optional: use a specific AWS profile
    # profile = "terraform"
  }
}
```

```bash
# Setting up the S3 backend (bootstrap — run ONCE manually)

# Create S3 bucket
aws s3api create-bucket \
    --bucket mycompany-terraform-state \
    --region us-east-1

# Enable versioning (for state recovery)
aws s3api put-bucket-versioning \
    --bucket mycompany-terraform-state \
    --versioning-configuration Status=Enabled

# Enable encryption
aws s3api put-bucket-encryption \
    --bucket mycompany-terraform-state \
    --server-side-encryption-configuration \
    '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"aws:kms"}}]}'

# Block public access
aws s3api put-public-access-block \
    --bucket mycompany-terraform-state \
    --public-access-block-configuration \
    "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true"

# Create DynamoDB table for locking
aws dynamodb create-table \
    --table-name terraform-state-locks \
    --attribute-definitions AttributeName=LockID,AttributeType=S \
    --key-schema AttributeName=LockID,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST
```

### Remote State — GCS Backend

```hcl
terraform {
  backend "gcs" {
    bucket = "mycompany-terraform-state"
    prefix = "production/web"
  }
}
```

### Remote State — Terraform Cloud

```hcl
terraform {
  cloud {
    organization = "mycompany"
    workspaces {
      name = "production-web"
    }
  }
}
```

---

## 🟢 State Locking Deep Dive

```
Without locking:

  Developer A                    Developer B
  ────────────                   ────────────
  terraform plan                 terraform plan
  (reads state: serial 10)       (reads state: serial 10)
       │                              │
  terraform apply                     │
  (writes state: serial 11)           │
  VPC created ✅                       │
       │                         terraform apply
       │                         (writes state: serial 11)
       │                         ← OVERWRITES state!
       │                         State no longer knows about VPC!
       │                         Resources orphaned = leaked resources + billing
```

```
With DynamoDB locking:

  Developer A                    Developer B
  ────────────                   ────────────
  terraform apply
  → Writes lock to DynamoDB ✅
  → Reads state
  → Makes changes                terraform apply
  → Writes updated state         → Tries to acquire lock
  → Releases lock ✅              → LOCK DENIED ❌
                                  → "Error acquiring the state lock"
                                  → "Lock Info: ID=abc123, Who=dev-a"
                                  → Waits...
  
                                  (After Dev A finishes)
                                  terraform apply
                                  → Acquires lock ✅
                                  → Reads UPDATED state
                                  → Makes changes safely
```

### Force Unlock (Emergency Only)

```bash
# Lock is stuck (someone's terminal crashed during apply)
terraform force-unlock LOCK_ID

# ⚠️ DANGER: Only do this when you're CERTAIN:
# 1. Nobody is actually running terraform
# 2. The previous run crashed/was killed
# 3. You understand the risk of concurrent modification
```

---

## 🟡 State Security Best Practices

### Encryption

```hcl
# S3: Encrypt at rest with KMS
backend "s3" {
  bucket     = "my-state-bucket"
  key        = "terraform.tfstate"
  encrypt    = true
  kms_key_id = "arn:aws:kms:us-east-1:111111111111:key/abc123"
}
```

### Access Control

```json
// IAM policy: Only specific roles can access state
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject"
      ],
      "Resource": "arn:aws:s3:::my-state-bucket/production/*",
      "Condition": {
        "StringEquals": {
          "aws:PrincipalTag/team": "platform"
        }
      }
    },
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:DeleteItem"
      ],
      "Resource": "arn:aws:dynamodb:us-east-1:*:table/terraform-state-locks"
    }
  ]
}
```

### State Isolation

```
❌ One state file for everything:
   everything.tfstate → contains ALL infrastructure
   Blast radius: ENTIRE COMPANY

✅ Separate state per environment AND per component:
   networking/production/terraform.tfstate
   compute/production/terraform.tfstate
   database/production/terraform.tfstate
   networking/staging/terraform.tfstate
   compute/staging/terraform.tfstate
   
   Blast radius: One component in one environment
```

---

## 🟡 Common State Operations

### terraform state list

```bash
$ terraform state list
aws_vpc.main
aws_subnet.public[0]
aws_subnet.public[1]
aws_security_group.web
aws_instance.web
aws_db_instance.main
module.monitoring.aws_cloudwatch_metric_alarm.cpu
```

### terraform state show

```bash
$ terraform state show aws_instance.web
# aws_instance.web:
resource "aws_instance" "web" {
    ami                         = "ami-12345678"
    arn                         = "arn:aws:ec2:us-east-1:111111:instance/i-0abc123"
    id                          = "i-0abc123def456789"
    instance_type               = "t3.small"
    private_ip                  = "10.0.1.15"
    public_ip                   = "54.123.45.67"
    subnet_id                   = "subnet-abc123"
    tags                        = {
        "Name" = "web-server"
    }
}
```

### terraform state mv (Rename/Refactor)

```bash
# You renamed a resource in your code:
# resource "aws_instance" "server" → resource "aws_instance" "web_server"
# Without state mv: Terraform will DESTROY server and CREATE web_server
# With state mv: Just rename in state, no infrastructure change

terraform state mv aws_instance.server aws_instance.web_server
# Move "aws_instance.server" to "aws_instance.web_server"
# Successfully moved 1 object(s).

# Move into a module
terraform state mv aws_instance.web module.compute.aws_instance.web
```

### terraform import

```bash
# Import existing infrastructure into Terraform state
# Step 1: Write the resource block in your .tf file
# resource "aws_instance" "legacy" {
#   # Will be populated after import
# }

# Step 2: Import
terraform import aws_instance.legacy i-0abc123def456789

# Step 3: Run terraform plan to verify
# Plan should show no changes if resource matches code

# Step 4: Fill in the resource block to match reality
```

### terraform state rm (Dangerous)

```bash
# Remove a resource from state WITHOUT destroying it
terraform state rm aws_instance.old_server
# The EC2 instance still exists in AWS
# But Terraform no longer manages it
# It becomes an "unmanaged" resource

# Use case: Moving resource to a different Terraform project
```

---

## 🔴 State Disaster Recovery

### Scenario 1: Corrupted State

```bash
$ terraform plan
│ Error: Failed to read state file
│ The state file could not be parsed as JSON

# Recovery options:
# 1. Use terraform.tfstate.backup (local)
cp terraform.tfstate.backup terraform.tfstate

# 2. Use S3 versioning (remote)
aws s3api list-object-versions \
    --bucket my-state-bucket \
    --prefix production/terraform.tfstate

aws s3api get-object \
    --bucket my-state-bucket \
    --key production/terraform.tfstate \
    --version-id "version-id-from-above" \
    terraform.tfstate.recovered

# 3. Last resort: import everything
# This is painful but possible
```

### Scenario 2: State Drift (Real World ≠ State)

```bash
# Someone changed something in AWS console
$ terraform plan
# aws_security_group.web has been changed
# ~ ingress {
#   - from_port = 80
#   + from_port = 8080   ← Someone changed this manually!
# }

# Options:
# 1. Let Terraform fix it (apply → overwrite manual change)
terraform apply

# 2. Accept the manual change (update code to match)
# Edit security_group.tf to match the manual change
# terraform plan → No changes

# 3. Refresh state (update state to match reality)
terraform refresh   # Deprecated in favor of:
terraform apply -refresh-only
```

---

**Previous:** [01. Terraform Fundamentals](./01-terraform-fundamentals.md)  
**Next:** [03. Modules](./03-modules.md)
