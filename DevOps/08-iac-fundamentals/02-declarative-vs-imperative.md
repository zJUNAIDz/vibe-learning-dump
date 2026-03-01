# Declarative vs Imperative IaC

> **Declarative: "I want 3 servers." Imperative: "Create server 1, create server 2, create server 3." The difference matters more than you think.**

---

## 🟢 The Two Approaches

### Imperative: "How" (Step-by-Step Instructions)

```bash
# Bash script — imperative infrastructure
#!/bin/bash

# Step 1: Create VPC
VPC_ID=$(aws ec2 create-vpc --cidr-block 10.0.0.0/16 --query 'Vpc.VpcId' --output text)

# Step 2: Create subnet
SUBNET_ID=$(aws ec2 create-subnet --vpc-id $VPC_ID --cidr-block 10.0.1.0/24 --query 'Subnet.SubnetId' --output text)

# Step 3: Create security group
SG_ID=$(aws ec2 create-security-group --group-name my-sg --description "My SG" --vpc-id $VPC_ID --query 'GroupId' --output text)

# Step 4: Add rule to security group
aws ec2 authorize-security-group-ingress --group-id $SG_ID --protocol tcp --port 80 --cidr 0.0.0.0/0

# Step 5: Launch instances
for i in 1 2 3; do
    aws ec2 run-instances \
        --image-id ami-12345678 \
        --instance-type t3.small \
        --subnet-id $SUBNET_ID \
        --security-group-ids $SG_ID \
        --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=server-$i}]"
done
```

**You tell the tool exactly WHAT TO DO, step by step.**

### Declarative: "What" (Desired State)

```hcl
# Terraform — declarative infrastructure
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "main" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_security_group" "web" {
  vpc_id = aws_vpc.main.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "server" {
  count         = 3
  ami           = "ami-12345678"
  instance_type = "t3.small"
  subnet_id     = aws_subnet.main.id
  vpc_security_group_ids = [aws_security_group.web.id]

  tags = {
    Name = "server-${count.index + 1}"
  }
}
```

**You tell the tool WHAT YOU WANT. It figures out how to make it happen.**

---

## 🟢 The Critical Difference: Running It Twice

### Imperative — Run Twice = Problem

```bash
# First run: Creates 3 servers ✅
./create-infrastructure.sh

# Second run: Creates 3 MORE servers (now you have 6!) ❌
./create-infrastructure.sh

# You have to add checks:
if ! aws ec2 describe-vpcs --filters "Name=tag:Name,Values=my-vpc" | grep -q VpcId; then
    aws ec2 create-vpc ...
fi
# This gets VERY complicated VERY fast
```

### Declarative — Run Twice = Same Result

```bash
# First run: Creates 3 servers ✅
terraform apply

# Second run: "No changes. Infrastructure is up-to-date." ✅
terraform apply

# Third run: Still no changes ✅
terraform apply
```

**This property is called IDEMPOTENCY — running the same operation multiple times produces the same result.**

---

## 🟢 Idempotency — The Most Important Concept in IaC

### What Idempotency Means

```
Idempotent operation:
  f(x) = f(f(x)) = f(f(f(x)))
  
  In plain English:
  "Doing it once is the same as doing it 100 times"
  
  terraform apply × 1 = terraform apply × 100
  The result is the same.
```

### Why Idempotency Matters

```
Scenario: Deployment script runs halfway and crashes

Imperative (non-idempotent):
  Step 1: Create VPC ✅ (done)
  Step 2: Create subnet ✅ (done)
  Step 3: Create security group 💥 (crashed here)
  Step 4: Create instances (never ran)
  
  Now what? You can't just re-run the script because
  Steps 1 and 2 will try to create DUPLICATE resources!
  
  Options:
    A) Manually clean up and re-run → error-prone
    B) Add complex "check if exists" logic → hundreds of lines
    C) Give up and do it manually → defeats the purpose

Declarative (idempotent):
  terraform apply 💥 (crashed partway)
  
  Fix the issue, then:
  terraform apply ← JUST RUN IT AGAIN
  
  Terraform compares desired state with actual state:
    VPC exists? ✅ Skip
    Subnet exists? ✅ Skip
    Security group exists? ❌ Create it
    Instances exist? ❌ Create them
  
  Clean, safe, predictable.
```

---

## 🟡 Declarative IaC in Kubernetes

Kubernetes is inherently declarative:

```yaml
# "I want 3 replicas of nginx running"
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  replicas: 3              # Desired state: 3 pods
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.25
        ports:
        - containerPort: 80
```

```bash
# Apply the desired state
kubectl apply -f deployment.yaml
# Kubernetes creates 3 pods

# One pod crashes
# Kubernetes: "Desired: 3, Actual: 2, need to create 1 more"
# Automatically creates a replacement pod

# Apply again (idempotent)
kubectl apply -f deployment.yaml
# "deployment.apps/nginx unchanged"

# Change replicas to 5
# Edit: replicas: 5
kubectl apply -f deployment.yaml
# Kubernetes: "Desired: 5, Actual: 3, need to create 2 more"
```

**Kubernetes constantly reconciles actual state with desired state.** This is the declarative model at its best.

---

## 🟡 Comparison Table

| Feature | Imperative | Declarative |
|---------|-----------|-------------|
| **Defines** | Steps to execute | Desired end state |
| **Idempotent** | No (must add checks) | Yes (built-in) |
| **Partial failure recovery** | Hard (manual cleanup) | Easy (re-run) |
| **Learning curve** | Lower (familiar scripting) | Higher (new paradigm) |
| **Complexity** | Grows fast (all edge cases) | Stays manageable |
| **State tracking** | Manual | Automatic (state file) |
| **Drift detection** | Manual | Built-in (plan/diff) |
| **Examples** | Bash, AWS CLI, Python scripts | Terraform, CloudFormation, K8s |

### When to Use Imperative

```
✅ One-time scripts (data migration, cleanup)
✅ Simple, sequential operations
✅ Prototyping (before committing to IaC)
✅ Operations that are inherently procedural
```

### When to Use Declarative

```
✅ Production infrastructure (always)
✅ Anything you need to maintain long-term
✅ Multi-environment management
✅ Team collaboration
✅ Audit and compliance requirements
```

---

## 🟡 Hybrid Approaches

Some tools blur the line:

### Pulumi (Code-Based Declarative)

```typescript
// Looks imperative (it's TypeScript), but behaves declaratively
import * as aws from "@pulumi/aws";

const vpc = new aws.ec2.Vpc("main", {
    cidrBlock: "10.0.0.0/16",
});

const subnet = new aws.ec2.Subnet("main", {
    vpcId: vpc.id,
    cidrBlock: "10.0.1.0/24",
});

// Pulumi tracks state like Terraform
// Running twice → no changes (idempotent)
// But you get TypeScript's type system, loops, conditionals
for (let i = 0; i < 3; i++) {
    new aws.ec2.Instance(`server-${i}`, {
        ami: "ami-12345678",
        instanceType: "t3.small",
        subnetId: subnet.id,
    });
}
```

### AWS CDK (Code-Based, Generates CloudFormation)

```typescript
import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';

const vpc = new ec2.Vpc(this, 'VPC', {
    maxAzs: 2,
    natGateways: 1,
});

// CDK synthesizes this to CloudFormation YAML
// Then CloudFormation manages it declaratively
// Best of both worlds: code ergonomics + declarative management
```

---

## 🔴 Anti-Pattern: "Declarative Wrapper Around Imperative Logic"

```hcl
# ❌ Using Terraform but thinking imperatively
resource "null_resource" "setup" {
  provisioner "local-exec" {
    command = <<-EOT
      aws ec2 create-vpc --cidr-block 10.0.0.0/16
      aws ec2 create-subnet --vpc-id $(aws ec2 describe-vpcs ...) --cidr-block 10.0.1.0/24
      for i in 1 2 3; do
        aws ec2 run-instances --image-id ami-12345 --instance-type t3.small
      done
    EOT
  }
}
# This is just bash in a Terraform wrapper
# It's NOT idempotent, NOT tracked in state, NOT declarative

# ✅ Use actual Terraform resources (as shown above)
```

---

**Previous:** [01. Why IaC Exists](./01-why-iac-exists.md)  
**Next:** [03. State Management](./03-state-management.md)
