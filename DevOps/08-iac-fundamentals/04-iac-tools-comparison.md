# IaC Tools Comparison

> **There are many IaC tools. Each has strengths and weaknesses. Choose based on your actual needs, not hype.**

---

## 🟢 The Landscape

```
                         PROVISIONING                    CONFIGURATION
                    (Create infrastructure)           (Configure servers)
                    
  Declarative    ┌─────────────────────────┐    ┌─────────────────────────┐
                 │  Terraform              │    │  Ansible                │
                 │  CloudFormation         │    │  Puppet                 │
                 │  Pulumi                 │    │  Chef                   │
                 │  CDK                    │    │  SaltStack              │
                 └─────────────────────────┘    └─────────────────────────┘
                 
  Imperative     ┌─────────────────────────┐    ┌─────────────────────────┐
                 │  AWS CLI scripts         │    │  Bash scripts           │
                 │  Boto3 (Python)          │    │  Fabric                 │
                 │  gcloud CLI scripts      │    │  SSH + commands         │
                 └─────────────────────────┘    └─────────────────────────┘
```

---

## 🟢 Terraform

```hcl
# Multi-cloud, declarative, HCL language
provider "aws" {
  region = "us-east-1"
}

resource "aws_instance" "web" {
  ami           = "ami-12345678"
  instance_type = "t3.small"
  
  tags = {
    Name = "web-server"
  }
}

resource "aws_s3_bucket" "data" {
  bucket = "my-app-data-bucket"
}
```

| Aspect | Details |
|--------|---------|
| **Language** | HCL (HashiCorp Configuration Language) |
| **Cloud support** | AWS, GCP, Azure, Kubernetes, + 3000 providers |
| **State** | Local or remote (S3, GCS, Terraform Cloud) |
| **License** | BSL (Business Source License) since Aug 2023 |
| **Community** | Massive — most popular IaC tool |
| **Learning curve** | Moderate |

**Strengths:**
- Multi-cloud (same tool for AWS, GCP, Azure)
- Massive provider ecosystem (3000+ providers)
- Excellent plan/apply workflow (preview before change)
- Strong community and documentation
- Modules for code reuse

**Weaknesses:**
- HCL is limited (no real loops, limited conditionals)
- State management is complex and dangerous
- BSL license restricts competing products
- No built-in rollback (must re-apply previous config)

---

## 🟢 AWS CloudFormation

```yaml
# AWS-only, declarative, YAML/JSON
AWSTemplateFormatVersion: '2010-09-09'
Description: Web server stack

Resources:
  WebInstance:
    Type: AWS::EC2::Instance
    Properties:
      ImageId: ami-12345678
      InstanceType: t3.small
      Tags:
        - Key: Name
          Value: web-server
  
  DataBucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: my-app-data-bucket
```

| Aspect | Details |
|--------|---------|
| **Language** | YAML or JSON |
| **Cloud support** | AWS only |
| **State** | Managed by AWS (no state file to manage!) |
| **License** | Proprietary (free to use with AWS) |
| **Community** | Large (AWS ecosystem) |
| **Learning curve** | Moderate to high |

**Strengths:**
- No state file management (AWS handles it)
- Built-in rollback on failure
- Deep AWS integration (supports all AWS services immediately)
- Drift detection built-in
- Free (you pay for resources, not the tool)

**Weaknesses:**
- AWS only (vendor lock-in)
- Verbose YAML (can be thousands of lines)
- Slow (creation/updates can take minutes)
- Error messages are cryptic
- Limited programming constructs

---

## 🟡 Pulumi

```typescript
// Code-based, declarative behavior, real programming language
import * as aws from "@pulumi/aws";

const vpc = new aws.ec2.Vpc("main", {
    cidrBlock: "10.0.0.0/16",
});

// Real loops, conditionals, functions — it's just TypeScript!
const servers = [];
for (let i = 0; i < 3; i++) {
    servers.push(new aws.ec2.Instance(`server-${i}`, {
        ami: "ami-12345678",
        instanceType: "t3.small",
        subnetId: vpc.id,
    }));
}

// Export outputs
export const serverIps = servers.map(s => s.publicIp);
```

| Aspect | Details |
|--------|---------|
| **Language** | TypeScript, Python, Go, C#, Java |
| **Cloud support** | AWS, GCP, Azure, Kubernetes, + more |
| **State** | Pulumi Cloud or self-managed backends |
| **License** | Apache 2.0 (open source) |
| **Learning curve** | Low (if you know the programming language) |

**Strengths:**
- Use your existing programming language (TypeScript!)
- Full programming power (loops, conditionals, functions, classes)
- Type safety and IDE support
- Multi-cloud
- Can import Terraform providers

**Weaknesses:**
- Smaller community than Terraform
- State managed by Pulumi Cloud (or self-host)
- Complexity — full programming language can be over-engineered
- Testing infrastructure code is harder than it seems

---

## 🟡 AWS CDK (Cloud Development Kit)

```typescript
// Code-based, generates CloudFormation
import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as ecs from 'aws-cdk-lib/aws-ecs';

export class MyStack extends cdk.Stack {
    constructor(scope: cdk.App, id: string) {
        super(scope, id);
        
        const vpc = new ec2.Vpc(this, 'VPC', {
            maxAzs: 2,
        });
        
        const cluster = new ecs.Cluster(this, 'Cluster', {
            vpc,
        });
        
        // High-level construct: creates ALB + ECS Service + Target Group + everything
        new ecs.patterns.ApplicationLoadBalancedFargateService(this, 'Service', {
            cluster,
            taskImageOptions: {
                image: ecs.ContainerImage.fromRegistry('my-app:latest'),
            },
        });
    }
}
```

| Aspect | Details |
|--------|---------|
| **Language** | TypeScript, Python, Java, C#, Go |
| **Cloud support** | AWS only |
| **State** | Managed by CloudFormation |
| **License** | Apache 2.0 |
| **Learning curve** | Moderate |

**Strengths:**
- High-level constructs (L2/L3) abstract away boilerplate
- No state file management (CloudFormation handles it)
- TypeScript type safety
- Generates CloudFormation (you can inspect the output)
- Built-in rollback

**Weaknesses:**
- AWS only
- Generates CloudFormation (debugging means reading generated YAML)
- Large dependency tree
- Breaking changes between CDK versions

---

## 🟡 When to Use What

```
Decision Tree:

"What clouds do I use?"
├── AWS only → "Do I want code or YAML?"
│   ├── Code → AWS CDK
│   └── YAML → CloudFormation
│
├── Multi-cloud → "Do I want HCL or real code?"
│   ├── HCL → Terraform
│   └── Code → Pulumi
│
└── Kubernetes only → "What am I managing?"
    ├── K8s resources → kubectl + YAML (or Helm)
    └── K8s + Cloud → Terraform (or Pulumi)
```

### Comparison Matrix

| Feature | Terraform | CloudFormation | Pulumi | CDK |
|---------|-----------|---------------|--------|-----|
| Multi-cloud | ✅ | ❌ | ✅ | ❌ |
| Language | HCL | YAML/JSON | TS/Python/Go | TS/Python/Go |
| State management | You manage | AWS manages | Pulumi Cloud | AWS manages |
| Rollback | Manual | Automatic | Manual | Automatic |
| Community size | Huge | Large | Growing | Growing |
| IDE support | Good | Poor | Excellent | Excellent |
| Testing | Limited | Limited | Native | Native |
| Learning curve | Moderate | Moderate | Low* | Moderate |
| Best for | Multi-cloud, large orgs | AWS-only shops | Dev-friendly IaC | AWS with code |

*Low if you already know TypeScript/Python

---

## 🔴 Common Mistakes When Choosing Tools

### Mistake 1: Multi-Cloud When You Don't Need It

```
"We chose Terraform because it's multi-cloud"
"We only use AWS"
"We've been using it for 3 years"
"We still only use AWS"

→ CloudFormation or CDK would have been simpler
→ No state file management needed
→ Native AWS integration

Lesson: Don't optimize for a future that never comes
```

### Mistake 2: Too Many Tools

```
❌ Team uses:
  - Terraform for EC2 and S3
  - CloudFormation for Lambda and API Gateway
  - Ansible for server configuration
  - Manual clicks for RDS (because "it's easier")
  - Bash scripts for DNS

Everyone is confused. Nobody knows where to look.

✅ Pick ONE provisioning tool and stick with it
  - Terraform for ALL infrastructure
  - Ansible only for server configuration (if needed)
```

### Mistake 3: Treating IaC Like Application Code

```
❌ Over-engineering IaC:
  - Abstract factory pattern for Terraform modules
  - 12 levels of module nesting
  - DRY taken to the extreme (unreadable)
  - Generic "create anything" module

✅ Keep IaC simple:
  - Flat structure when possible
  - Some duplication is OK (clarity > DRY)
  - Maximum 2-3 levels of module nesting
  - Specific modules for specific purposes
```

---

**Previous:** [03. State Management](./03-state-management.md)
