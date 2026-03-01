# Why Infrastructure as Code Exists

> **If you can't reproduce your infrastructure from scratch in 30 minutes, you don't have infrastructure — you have a ticking time bomb.**

---

## 🟢 The World Before IaC

### Manual Infrastructure (The Dark Ages)

```
Step 1: Submit a ticket to the infrastructure team
        "Please create a new server for our API"
        Wait: 2 weeks

Step 2: Get server access
        "Here's your server: 10.0.1.47"
        SSH in, start installing things manually

Step 3: Install software
        $ sudo apt-get update
        $ sudo apt-get install nginx nodejs postgresql
        $ sudo vim /etc/nginx/sites-available/myapp    ← good luck remembering this

Step 4: Configure networking
        Ask the network team for a load balancer
        Wait: 1 week
        Ask for DNS entry
        Wait: 3 days

Step 5: Repeat for staging environment
        "It's slightly different because I forgot one step"

Step 6: Something breaks
        "What did we install on this server?"
        "Nobody knows, the person who set it up left the company"
```

### The Problems with Manual Infrastructure

| Problem | What Happens |
|---------|-------------|
| **Snowflake servers** | Every server is unique — "configured by hand, understood by nobody" |
| **Configuration drift** | Dev and prod slowly diverge until nothing matches |
| **No audit trail** | Who changed that firewall rule? When? Why? |
| **Disaster recovery** | Data center burns down → months to rebuild |
| **Knowledge loss** | Senior engineer leaves → tribal knowledge gone |
| **Slow provisioning** | Weeks to get a new environment |
| **Human error** | Typo in iptables rule → production outage |

---

## 🟢 What Infrastructure as Code Means

**IaC = Describe your infrastructure in text files, version them in Git, and let tools create/update the actual infrastructure.**

```
Before IaC:                          With IaC:
┌──────────────────┐                 ┌──────────────────┐
│ Human            │                 │ Code (Git)       │
│ clicks buttons   │                 │ main.tf          │
│ types commands   │                 │ variables.tf     │
│ follows runbook  │                 │ outputs.tf       │
│ (maybe)          │                 │                  │
└────────┬─────────┘                 └────────┬─────────┘
         │                                    │
         ▼                                    ▼
┌──────────────────┐                 ┌──────────────────┐
│ Infrastructure   │                 │ IaC Tool         │
│ (snowflake,      │                 │ (Terraform,      │
│  undocumented,   │                 │  CloudFormation)  │
│  unreproducible) │                 │                  │
└──────────────────┘                 └────────┬─────────┘
                                              │
                                              ▼
                                     ┌──────────────────┐
                                     │ Infrastructure   │
                                     │ (identical,      │
                                     │  documented,     │
                                     │  reproducible)   │
                                     └──────────────────┘
```

### The Core Principle

```
Infrastructure code should be treated EXACTLY like application code:

✅ Version controlled (Git)
✅ Code reviewed (Pull Requests)
✅ Tested (plan before apply)
✅ Automated (CI/CD pipeline)
✅ Documented (the code IS the documentation)
✅ Idempotent (run it twice, same result)
```

---

## 🟢 Infrastructure Drift

### What Is Drift?

```
Day 0: Terraform creates 3 servers with identical config
       Server A: nginx 1.24, Node 20, Ubuntu 22.04
       Server B: nginx 1.24, Node 20, Ubuntu 22.04
       Server C: nginx 1.24, Node 20, Ubuntu 22.04

Day 30: Someone SSHs into Server B to "fix something quickly"
       Server A: nginx 1.24, Node 20, Ubuntu 22.04
       Server B: nginx 1.25, Node 18, Ubuntu 22.04  ← DRIFTED
       Server C: nginx 1.24, Node 20, Ubuntu 22.04

Day 90: Someone updates Server C through the AWS console
       Server A: nginx 1.24, Node 20, Ubuntu 22.04
       Server B: nginx 1.25, Node 18, Ubuntu 22.04  ← DRIFTED
       Server C: nginx 1.24, Node 20, Ubuntu 22.04, extra security group ← DRIFTED

Day 180: "Why does this work in staging but not production?"
        "Because production server B has a different Node version"
        "WHO CHANGED IT?"
        "Nobody knows"
```

### Why Drift Is Dangerous

```
1. Inconsistency → "Works in staging, broken in prod"
2. Security gaps → "That firewall rule was opened temporarily... 6 months ago"
3. Untracked changes → Terraform thinks state is X, reality is Y
4. Recovery failure → "We rebuilt the server but it doesn't work like the old one"
```

### How IaC Prevents Drift

```
Option 1: Detect and Alert
  terraform plan
  # Shows: "Server B has been modified outside of Terraform"
  # You fix it manually or run terraform apply to overwrite

Option 2: Prevent Manual Changes
  - Lock down console/SSH access
  - ALL changes must go through Terraform
  - Policy: "If it's not in Git, it doesn't exist"

Option 3: Immutable Infrastructure
  - Never modify servers
  - Need a change? Build a NEW server, destroy the old one
  - "Treat servers like cattle, not pets"
```

---

## 🟡 Disaster Recovery with IaC

### Without IaC

```
Scenario: AWS region goes down (us-east-1 outage — this happens!)

Without IaC:
  1. "What resources do we have in us-east-1?"
     → Nobody has a complete list
  2. "Can we recreate them in us-west-2?"
     → Some things, maybe, if we can remember the configs
  3. "How long will recovery take?"
     → Days to weeks
  4. "Will the new environment be identical?"
     → No. We can't even remember the original.

Result: Days of downtime, partial recovery, lost data
```

### With IaC

```
Scenario: Same AWS region outage

With IaC:
  1. "What resources do we have?"
     → Everything is in main.tf, networking.tf, database.tf...
  2. "Can we recreate in us-west-2?"
     → Change region variable, terraform apply
  3. "How long?"
     → 30 minutes to create infrastructure
     → Plus data restore from backups
  4. "Will it be identical?"
     → Yes. Same code, same infrastructure.

Result: Hours of downtime, full recovery, complete confidence
```

```hcl
# Change ONE variable to recover in a different region
variable "region" {
  default = "us-east-1"    # Change to "us-west-2" for disaster recovery
}

provider "aws" {
  region = var.region
}

# ALL resources are defined in code
# terraform apply → identical infrastructure in the new region
```

---

## 🟡 Multi-Environment Management

### The Problem

```
Most apps need multiple environments:
  ┌─────────┐    ┌──────────┐    ┌────────────┐
  │   Dev   │ →  │ Staging  │ →  │ Production │
  │         │    │          │    │            │
  │ 1 server│    │ 2 servers│    │ 10 servers │
  │ small DB│    │ medium DB│    │ large DB   │
  │ no CDN  │    │ no CDN   │    │ CDN + WAF  │
  └─────────┘    └──────────┘    └────────────┘

Without IaC:
  Each environment was set up by different people
  At different times, with different configurations
  They're "supposed to be the same" but never are

With IaC:
  SAME code, DIFFERENT variables
```

```hcl
# environments/dev.tfvars
instance_count = 1
instance_type  = "t3.small"
db_instance    = "db.t3.micro"
enable_cdn     = false

# environments/staging.tfvars
instance_count = 2
instance_type  = "t3.medium"
db_instance    = "db.t3.small"
enable_cdn     = false

# environments/production.tfvars
instance_count = 10
instance_type  = "t3.large"
db_instance    = "db.r6g.xlarge"
enable_cdn     = true
```

```bash
# Deploy to dev
terraform apply -var-file=environments/dev.tfvars

# Deploy to staging
terraform apply -var-file=environments/staging.tfvars

# Deploy to production
terraform apply -var-file=environments/production.tfvars
```

---

## 🟡 Benefits Summary

| Benefit | Manual | IaC |
|---------|--------|-----|
| Time to create new environment | Days/weeks | Minutes |
| Consistency across environments | Low | High |
| Disaster recovery | Hours/days | Minutes |
| Audit trail | None | Git history |
| Knowledge transfer | Tribal | Documented |
| Scaling | Manual, slow | Automated |
| Testing | "We'll see in prod" | Plan + apply in staging first |
| Cost control | Unknown | Visible in code |

---

**Previous:** [README](./README.md)  
**Next:** [02. Declarative vs Imperative](./02-declarative-vs-imperative.md)
