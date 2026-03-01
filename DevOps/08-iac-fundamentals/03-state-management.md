# State Management in IaC

> **State is the scariest part of IaC. It's the single source of truth about what exists in the real world — and if it's wrong, everything breaks.**

---

## 🟢 What Is State?

State is a **mapping** between your code and the real-world infrastructure.

```
Your Code (main.tf):              State File:                 Real World (AWS):
┌─────────────────┐          ┌──────────────────┐          ┌──────────────────┐
│ resource "aws_  │          │ {                │          │                  │
│   instance"     │  ←map→   │   "aws_instance" │  ←map→   │ EC2 Instance     │
│   "server" {    │          │   "server":      │          │ i-0abc123def     │
│   ami = "..."   │          │   id: "i-0abc123"│          │ running          │
│   type = "t3"   │          │   ami: "..."     │          │ t3.small         │
│ }               │          │   type: "t3.small│          │                  │
└─────────────────┘          │ }                │          └──────────────────┘
                             └──────────────────┘
```

### Why State Exists

```
Without state, Terraform would have NO IDEA:
  - "Does this resource already exist?"
  - "What's its current configuration?"
  - "What needs to change to match the code?"
  
Without state:
  terraform apply → "Let me create everything from scratch" ← DUPLICATE RESOURCES!
  
With state:
  terraform apply → "Let me compare desired state (code) with actual state (state file)"
                   → "Oh, the instance already exists. Only the instance type changed."
                   → "I'll just update the instance type."
```

---

## 🟢 What's Inside a State File?

```json
{
  "version": 4,
  "terraform_version": "1.7.0",
  "serial": 42,
  "lineage": "abc123-def456",
  "outputs": {
    "server_ip": {
      "value": "54.123.45.67",
      "type": "string"
    }
  },
  "resources": [
    {
      "mode": "managed",
      "type": "aws_instance",
      "name": "server",
      "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
      "instances": [
        {
          "schema_version": 1,
          "attributes": {
            "id": "i-0abc123def456789",
            "ami": "ami-12345678",
            "instance_type": "t3.small",
            "private_ip": "10.0.1.15",
            "public_ip": "54.123.45.67",
            "subnet_id": "subnet-abc123",
            "tags": {
              "Name": "my-server"
            }
          }
        }
      ]
    }
  ]
}
```

### State Contains EVERYTHING

```
⚠️ State includes SENSITIVE DATA:
  - Database passwords (in plaintext!)
  - API keys
  - Private IPs
  - Certificate private keys
  
This is why:
  ❌ NEVER commit terraform.tfstate to Git
  ❌ NEVER share state files over email/Slack
  ✅ Store state in encrypted remote backend
  ✅ Restrict access to state files
```

---

## 🟢 Local vs Remote State

### Local State (Default — Don't Use for Teams)

```
$ terraform apply
# Creates: terraform.tfstate in current directory

project/
├── main.tf
├── variables.tf
├── terraform.tfstate          ← State on YOUR laptop
└── terraform.tfstate.backup   ← Previous state
```

**Problems with local state:**
```
Developer A runs terraform apply on their laptop
Developer B runs terraform apply on their laptop
Both have DIFFERENT state files!

Developer A: "I created server-1"
Developer B: "I also created server-1"  ← DUPLICATE! Or worse...

Developer A: "I deleted server-2"
Developer B's state still shows server-2 exists → DRIFT
```

### Remote State (Always Use This)

```hcl
# backend.tf — Store state in S3
terraform {
  backend "s3" {
    bucket         = "my-company-terraform-state"
    key            = "production/infrastructure/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true                    # Encrypt at rest
    dynamodb_table = "terraform-state-lock"  # State locking
  }
}
```

```
Remote State Flow:
                                    ┌──────────────┐
Developer A ──terraform apply──→   │   S3 Bucket   │
                                    │  (state file) │
Developer B ──terraform apply──→   │              │
                                    └──────────────┘
                                         │
Both developers read/write the          │
SAME state file. Single source          │
of truth.                               │
                                    ┌──────────────┐
                                    │  DynamoDB     │
                                    │  (lock table) │
                                    └──────────────┘
                                    Prevents concurrent
                                    modifications
```

---

## 🟢 State Locking

### Why Locking Matters

```
Without locking:

10:00:00  Dev A: terraform plan     → Reads state
10:00:01  Dev B: terraform plan     → Reads SAME state
10:00:05  Dev A: terraform apply    → Writes updated state
10:00:06  Dev B: terraform apply    → Overwrites A's state! 💥
                                      ← B's state doesn't include A's changes
                                      ← Infrastructure might be corrupted
```

### How Locking Works

```
With locking (DynamoDB for S3 backend):

10:00:00  Dev A: terraform apply
          → Acquires lock in DynamoDB ✅
          → Reads state from S3
          → Makes changes
          
10:00:01  Dev B: terraform apply
          → Tries to acquire lock
          → Lock exists! ❌
          → "Error: Error acquiring the state lock"
          → Waits or retries later

10:00:30  Dev A: terraform apply finishes
          → Writes updated state to S3
          → Releases lock ✅

10:00:31  Dev B: terraform apply
          → Acquires lock ✅
          → Reads UPDATED state (includes A's changes)
          → Makes changes safely
```

```hcl
# Setting up state locking with S3 + DynamoDB
# Step 1: Create the DynamoDB table (do this once, manually or with a bootstrap script)

resource "aws_dynamodb_table" "terraform_lock" {
  name           = "terraform-state-lock"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }
}

# Step 2: Configure backend with locking
terraform {
  backend "s3" {
    bucket         = "my-terraform-state"
    key            = "prod/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"  # ← This enables locking
  }
}
```

---

## 🟡 State Operations

### Viewing State

```bash
# List all resources in state
terraform state list
# aws_instance.server
# aws_vpc.main
# aws_subnet.main
# aws_security_group.web

# Show details of a specific resource
terraform state show aws_instance.server
# id          = "i-0abc123def456789"
# ami         = "ami-12345678"
# instance_type = "t3.small"
# ...
```

### Moving Resources in State

```bash
# Rename a resource (when you refactor your code)
# Code change: resource "aws_instance" "server" → resource "aws_instance" "web_server"
terraform state mv aws_instance.server aws_instance.web_server
# Without this: Terraform would DELETE the old one and CREATE a new one!

# Move to a module
terraform state mv aws_instance.server module.compute.aws_instance.server
```

### Removing from State (Without Destroying)

```bash
# "I want Terraform to forget about this resource, but DON'T delete it"
terraform state rm aws_instance.server
# Resource still exists in AWS, but Terraform no longer manages it

# Use case: Moving resource management to a different Terraform project
```

### Importing Existing Resources

```bash
# "This resource already exists in AWS but Terraform doesn't know about it"
# Step 1: Write the resource block in your .tf file
# Step 2: Import the real resource into state
terraform import aws_instance.server i-0abc123def456789
# Now Terraform manages this existing instance
```

---

## 🟡 State File Organization

### One State File per Environment

```
terraform-infrastructure/
├── environments/
│   ├── dev/
│   │   ├── main.tf
│   │   ├── backend.tf        # key = "dev/terraform.tfstate"
│   │   └── terraform.tfvars
│   ├── staging/
│   │   ├── main.tf
│   │   ├── backend.tf        # key = "staging/terraform.tfstate"
│   │   └── terraform.tfvars
│   └── production/
│       ├── main.tf
│       ├── backend.tf        # key = "production/terraform.tfstate"
│       └── terraform.tfvars
└── modules/
    ├── vpc/
    ├── compute/
    └── database/
```

**Why separate state files?**
```
1. Blast radius: mistake in dev doesn't affect production
2. Locking: dev operations don't block production deploys
3. Permissions: different IAM roles per environment
4. Speed: smaller state files = faster plan/apply
```

---

## 🔴 State Disasters and Recovery

### Disaster 1: State File Deleted

```
$ rm terraform.tfstate  # Oops

Terraform: "I have no state. I don't know about any resources."
         "terraform plan shows: + create ALL resources"
         
But the resources EXIST in AWS!
If you terraform apply now → DUPLICATE EVERYTHING

Recovery:
  Option 1: Restore from backup (terraform.tfstate.backup)
  Option 2: Restore from S3 versioning (if remote state)
  Option 3: terraform import every resource manually 😱
```

### Disaster 2: State File Corrupted

```
$ terraform plan
Error: Failed to read state file

Recovery:
  1. Check terraform.tfstate.backup
  2. Check S3 versioning for previous state
  3. If no backup → import everything from scratch
```

### Disaster 3: State Locked and Can't Unlock

```
$ terraform apply
Error: Error acquiring the state lock

This happens when:
  - Previous terraform apply crashed
  - Someone's terminal was killed during apply
  - CI/CD pipeline timed out during apply

Recovery:
  # Check who holds the lock
  terraform force-unlock LOCK_ID
  # ⚠️ Only do this if you're SURE no one is running terraform
```

### Prevention

```bash
# 1. Always use remote state with versioning
aws s3api put-bucket-versioning \
    --bucket my-terraform-state \
    --versioning-configuration Status=Enabled

# 2. Enable S3 bucket lifecycle (keep 30 versions)
# 3. Use DynamoDB locking
# 4. Run terraform in CI/CD (not laptops) to prevent concurrent runs
# 5. Never manually edit the state file
```

---

**Previous:** [02. Declarative vs Imperative](./02-declarative-vs-imperative.md)  
**Next:** [04. IaC Tools Comparison](./04-iac-tools-comparison.md)
