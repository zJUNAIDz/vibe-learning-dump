# Terraform State Corruption

> **Terraform state is the source of truth for your infrastructure. If it's corrupted, Terraform doesn't know what exists. It will try to recreate everything — including your production database.**

---

## 🟢 What Is Terraform State

```
Terraform state maps your .tf files to real resources:

Resource in code         State file              Real world
──────────────          ──────────              ──────────
aws_instance.web  →  id: i-abc123  →  EC2 instance i-abc123
aws_db.main       →  id: mydb      →  RDS instance mydb
aws_s3.bucket     →  id: my-files  →  S3 bucket my-files

If state is lost or corrupted:
  → Terraform doesn't know i-abc123 exists
  → Terraform says "I need to CREATE aws_instance.web"
  → Now you have TWO EC2 instances
  → Or worse: Terraform tries to create RDS, name conflict,
    gives up, and your pipeline fails at 3 AM

If state says a resource exists but it was manually deleted:
  → Terraform tries to update a resource that doesn't exist
  → Error, confusion, manual intervention required
```

---

## 🟢 Concurrent Applies

### How It Happens

```
Scenario: Two developers run terraform apply at the same time.

Developer A:                        Developer B:
  terraform plan                      terraform plan
  → "Add new EC2 instance"           → "Change security group"
  terraform apply                     terraform apply
       │                                    │
       │  Read state file                   │  Read state file
       │  (same version)                    │  (same version)
       │                                    │
       ▼                                    ▼
  Write state (v2)                    Write state (v2)
  Added EC2 instance                  Changed security group
       │                                    │
       └─── B's write OVERWRITES A's ──────┘
       
Result: State says security group changed but NO EC2 instance.
  → EC2 instance exists in AWS but NOT in state
  → Terraform doesn't know about it → "orphaned resource"
  → Next terraform plan wants to create it AGAIN
```

### Prevention: State Locking

```hcl
# S3 backend with DynamoDB locking
terraform {
  backend "s3" {
    bucket         = "my-terraform-state"
    key            = "production/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-locks"  # THIS prevents concurrent applies
  }
}

# What happens now:
# Developer A: terraform apply → acquires lock ✓
# Developer B: terraform apply → "Error: state locked by A"
# Developer A: apply completes → releases lock
# Developer B: retries → acquires lock ✓ → sees A's changes
```

```hcl
# Create the locking table
resource "aws_dynamodb_table" "terraform_locks" {
  name         = "terraform-locks"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"
  
  attribute {
    name = "LockID"
    type = "S"
  }
}
```

---

## 🟡 Manual Changes (State Drift)

### The Worst Thing You Can Do

```
Scenario: Someone changes infrastructure through the AWS console.

Before:
  Terraform state: security group allows port 443
  AWS reality: security group allows port 443
  → In sync ✓

Someone opens port 22 in AWS console "just for debugging":
  Terraform state: security group allows port 443
  AWS reality: security group allows port 443 AND 22
  → DRIFT ✗

Next terraform apply:
  Option 1: Terraform removes port 22 (reverts manual change)
    → Developer SSH session drops mid-debug
    → "WHO DELETED MY SSH ACCESS?!"
    
  Option 2: Terraform sees the change and is confused
    → Plan shows unexpected changes
    → Nobody knows if they're safe to apply

Worse scenario: Someone deletes an RDS instance manually
  Terraform state: RDS instance exists
  AWS reality: RDS instance is GONE
  
  terraform plan: "No changes" (reads state, not reality)
  terraform apply: tries to modify → Error: instance not found
  → Now you need to manually fix state
```

### Detecting and Fixing Drift

```bash
# Detect drift
terraform plan -refresh-only
# Shows: AWS has port 22 open, state doesn't know about it

# Option 1: Accept the drift (import into state)
terraform apply -refresh-only
# Updates state to match reality — port 22 now in state
# Then update your .tf files to match

# Option 2: Revert the drift (apply your code)
terraform apply
# Terraform removes port 22, back to what code says

# Option 3: Import manually created resources
terraform import aws_security_group.main sg-abc123
```

### Prevention

```
1. NEVER change infrastructure manually
   → Not in the console
   → Not with AWS CLI
   → Not with kubectl (for Terraform-managed resources)
   
2. Use read-only console access for most people
   → Only CI/CD pipeline has write access
   → Developers can look but not touch

3. Run drift detection regularly
   → Scheduled terraform plan -refresh-only
   → Alert if drift detected

4. Tag Terraform-managed resources
   → ManagedBy = "terraform"
   → Anyone seeing this tag knows: don't touch manually
```

---

## 🟡 Force Unlocking

### When the Lock Gets Stuck

```
Scenario:
  1. Developer runs terraform apply
  2. Laptop crashes mid-apply
  3. Lock is stuck in DynamoDB
  4. Nobody else can run terraform
  5. "terraform apply" → Error: state locked

The WRONG fix:
  terraform force-unlock <lock-id>
  
Why it's dangerous:
  → Maybe the original apply is STILL RUNNING somewhere
  → CI/CD pipeline crashed but the apply finished
  → Force unlock + new apply = concurrent modification
```

### Safe Force Unlock Process

```bash
# Step 1: VERIFY the lock is actually stuck
# Check who holds the lock
aws dynamodb get-item \
  --table-name terraform-locks \
  --key '{"LockID":{"S":"my-terraform-state/production/terraform.tfstate-md5"}}' \
  --output json | jq '.Item'
# Shows: who locked it, when, from which machine

# Step 2: Confirm the operation is NOT still running
# → Check the CI/CD pipeline
# → Check with the person named in the lock
# → Wait at least 30 minutes for long applies

# Step 3: Force unlock (only when CERTAIN it's stuck)
terraform force-unlock <lock-id>
# WARNING: This will do "dangerous" things if someone is still running!

# Step 4: Run terraform plan BEFORE apply
terraform plan
# Verify the state is consistent before making changes
```

---

## 🔴 State File Corruption

### How State Gets Corrupted

```
1. Partial write (crash during terraform apply)
   → State file half-written
   → JSON is invalid
   → Terraform can't read it

2. Merge conflict (shouldn't happen with remote state, but...)
   → Two branches modify state
   → Git merge produces invalid JSON
   → State file is garbage

3. Manual editing (someone "fixes" state by hand)
   → Forgets a comma
   → Removes a resource mapping
   → Changes an ID incorrectly
```

### Recovery from Backup

```bash
# S3 versioning — your lifeline
# Enable versioning on state bucket (do this NOW)

# List state file versions
aws s3api list-object-versions \
  --bucket my-terraform-state \
  --prefix "production/terraform.tfstate" \
  --max-items 10

# Download previous version
aws s3api get-object \
  --bucket my-terraform-state \
  --key "production/terraform.tfstate" \
  --version-id "abc123" \
  terraform.tfstate.backup

# Verify backup
terraform show terraform.tfstate.backup

# Restore by uploading
aws s3 cp terraform.tfstate.backup \
  s3://my-terraform-state/production/terraform.tfstate

# Then immediately:
terraform plan -refresh-only
# Reconcile state with actual infrastructure
```

### Last Resort: Rebuild State

```bash
# If state is completely lost:
# 1. Create empty state
rm -f terraform.tfstate

# 2. Import each resource manually
terraform import aws_vpc.main vpc-abc123
terraform import aws_subnet.public subnet-def456
terraform import aws_instance.web i-ghi789
terraform import aws_db_instance.main mydb

# This is PAINFUL for large infrastructure.
# Can take hours for 100+ resources.
# This is why you ALWAYS enable S3 versioning on state buckets.
```

---

## 🔴 Prevention Checklist

```
□ Remote state backend (S3, GCS, Azure Blob)
□ State locking (DynamoDB, GCS native, Azure native)
□ State file encryption at rest
□ S3 versioning enabled on state bucket
□ Read-only access for developers (CI/CD applies)
□ terraform plan in CI, terraform apply only on main merge
□ Drift detection (scheduled plan -refresh-only)
□ ManagedBy=terraform tags on all resources
□ No manual infrastructure changes
□ State backup tested (can you actually restore?)
□ Separate state files per environment (dev/staging/prod)
□ Separate state files per service/component
```

```hcl
# Minimum viable state configuration
terraform {
  backend "s3" {
    bucket         = "mycompany-terraform-state"
    key            = "production/networking/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-locks"
    
    # Separate state per component:
    # production/networking/terraform.tfstate
    # production/database/terraform.tfstate
    # production/application/terraform.tfstate
    # staging/networking/terraform.tfstate
  }
}
```

---

**Previous:** [04. Kubernetes Outages](./04-kubernetes-outages.md)  
**Next:** [06. Observability Failures](./06-observability-failures.md)
