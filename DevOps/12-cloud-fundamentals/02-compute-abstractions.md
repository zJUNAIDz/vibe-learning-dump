# Compute Abstractions

> **VMs, containers, serverless, Kubernetes — each is a layer of abstraction that trades control for convenience. Know when to use which.**

---

## 🟢 The Compute Spectrum

```
More Control ←─────────────────────────────────────→ Less Control
More Work    ←─────────────────────────────────────→ Less Work

  Bare Metal    VMs         Containers    Serverless
  ──────────    ───         ──────────    ──────────
  Physical      Virtual     Docker        Functions
  hardware      machines    containers    triggered by
                                          events
  
  You manage:   You manage:  You manage:  You manage:
  Everything    OS + app     App + deps   Just code
  
  Example:      EC2          ECS/EKS      Lambda
                GCE          GKE          Cloud Functions
                Azure VMs    AKS          Azure Functions
```

---

## 🟢 Virtual Machines (IaaS)

A VM is a software-defined computer with its own OS.

```
Physical Server
├── Hypervisor (splits the hardware)
│   ├── VM 1: Ubuntu 22.04 — 4 vCPUs, 16 GB RAM
│   │   ├── nginx
│   │   ├── Node.js app
│   │   └── PostgreSQL client
│   ├── VM 2: Amazon Linux — 2 vCPUs, 8 GB RAM
│   │   ├── Go API server
│   │   └── Redis
│   └── VM 3: Windows Server — 8 vCPUs, 32 GB RAM
│       └── .NET application
```

### AWS EC2 (Elastic Compute Cloud)

```bash
# Launch an EC2 instance
aws ec2 run-instances \
  --image-id ami-0c55b159cbfafe1f0 \       # Ubuntu 22.04
  --instance-type t3.medium \               # 2 vCPUs, 4 GB RAM
  --key-name my-ssh-key \
  --security-group-ids sg-12345 \
  --subnet-id subnet-12345

# Instance types:
# t3.micro   — 2 vCPU, 1 GB   ($0.0104/hr) — dev, testing
# t3.medium  — 2 vCPU, 4 GB   ($0.0416/hr) — small apps
# m5.large   — 2 vCPU, 8 GB   ($0.096/hr)  — general purpose
# c5.xlarge  — 4 vCPU, 8 GB   ($0.17/hr)   — CPU intensive
# r5.large   — 2 vCPU, 16 GB  ($0.126/hr)  — memory intensive
```

### Terraform Example

```hcl
resource "aws_instance" "web" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t3.medium"
  key_name      = "my-ssh-key"
  
  vpc_security_group_ids = [aws_security_group.web.id]
  subnet_id              = aws_subnet.public.id
  
  user_data = <<-EOF
    #!/bin/bash
    apt-get update
    apt-get install -y nginx
    systemctl start nginx
  EOF
  
  tags = {
    Name = "web-server"
  }
}
```

**When to use VMs:**
```
✅ Need full OS control
✅ Running legacy applications
✅ Specific OS/kernel requirements
✅ Long-running, predictable workloads
✅ Compliance requirements (dedicated tenancy)
```

---

## 🟢 Containers (CaaS)

Containers share the host OS kernel but isolate the application.

```
VM approach:
┌─────────────┐ ┌─────────────┐
│  App A       │ │  App B       │
│  Libraries   │ │  Libraries   │
│  Guest OS    │ │  Guest OS    │  ← Each VM has full OS (GB)
│  (Ubuntu)    │ │  (Alpine)    │
├─────────────┴─┴──────────────┤
│          Hypervisor           │
├──────────────────────────────┤
│        Host OS + Hardware     │
└──────────────────────────────┘

Container approach:
┌─────────┐ ┌─────────┐ ┌─────────┐
│  App A   │ │  App B   │ │  App C   │
│  Libs    │ │  Libs    │ │  Libs    │  ← Containers share OS (MB)
├─────────┴─┴─────────┴─┴─────────┤
│        Container Runtime          │
│        (Docker / containerd)      │
├──────────────────────────────────┤
│        Host OS + Hardware         │
└──────────────────────────────────┘

VM:        Minutes to start, GBs of disk, heavy
Container: Seconds to start, MBs of disk, light
```

### Cloud Container Services

```
AWS:
  ECS (Elastic Container Service) — AWS's container orchestrator
  EKS (Elastic Kubernetes Service) — Managed Kubernetes
  Fargate — Serverless containers (no servers to manage)
  
GCP:
  Cloud Run — Serverless containers (simplest option)
  GKE (Google Kubernetes Engine) — Managed Kubernetes
  
Azure:
  ACI (Azure Container Instances) — Simple container hosting
  AKS (Azure Kubernetes Service) — Managed Kubernetes

Common pattern:
  Small project → Cloud Run / Fargate (zero server management)
  Complex project → EKS / GKE / AKS (full Kubernetes)
```

### ECS Task Definition

```json
{
  "family": "my-api",
  "containerDefinitions": [
    {
      "name": "api",
      "image": "123456789.dkr.ecr.us-east-1.amazonaws.com/my-api:v1.2.3",
      "portMappings": [
        {
          "containerPort": 3000,
          "protocol": "tcp"
        }
      ],
      "environment": [
        { "name": "NODE_ENV", "value": "production" }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/my-api",
          "awslogs-region": "us-east-1"
        }
      },
      "memory": 512,
      "cpu": 256
    }
  ]
}
```

**When to use containers:**
```
✅ Microservices architecture
✅ Need consistent dev/staging/prod environments
✅ Fast scaling (seconds, not minutes)
✅ CI/CD pipelines (build once, deploy anywhere)
✅ Modern applications (12-factor apps)
```

---

## 🟡 Serverless (FaaS)

"No servers to manage" — you write a function, the cloud runs it when triggered.

```
Traditional:
  Server running 24/7, waiting for requests
  Paying even when idle ($$$)

Serverless:
  Function deployed, not running
  Request comes → cold start → function runs → function stops
  Paying only when executing ($)
```

### AWS Lambda

```typescript
// handler.ts — a Lambda function
import { APIGatewayEvent, APIGatewayProxyResult } from 'aws-lambda';

export const handler = async (event: APIGatewayEvent): Promise<APIGatewayProxyResult> => {
  const name = event.queryStringParameters?.name || 'World';
  
  return {
    statusCode: 200,
    body: JSON.stringify({
      message: `Hello, ${name}!`,
      timestamp: new Date().toISOString()
    })
  };
};
```

```hcl
# Lambda in Terraform
resource "aws_lambda_function" "api" {
  function_name = "my-api"
  runtime       = "nodejs18.x"
  handler       = "handler.handler"
  filename      = "lambda.zip"
  
  memory_size = 256     # MB
  timeout     = 30      # seconds
  
  environment {
    variables = {
      NODE_ENV = "production"
      DB_HOST  = aws_rds_cluster.main.endpoint
    }
  }
}

# API Gateway trigger
resource "aws_apigatewayv2_api" "api" {
  name          = "my-api"
  protocol_type = "HTTP"
}
```

### Serverless Pros/Cons

```
Pros:
  ✅ True pay-per-use (free at low traffic)
  ✅ Auto-scales to zero (no idle costs)
  ✅ Auto-scales to thousands (no capacity planning)
  ✅ No servers to patch or manage
  ✅ Great for event-driven workloads

Cons:
  ❌ Cold starts (first request is slow: 100ms-10s)
  ❌ Execution time limits (Lambda: 15 min max)
  ❌ Memory limits (Lambda: 10 GB max)
  ❌ Vendor lock-in (Lambda code ≠ portable)
  ❌ Hard to debug and test locally
  ❌ Complex architectures become hard to reason about
```

**When to use serverless:**
```
✅ API endpoints with variable traffic
✅ Event processing (S3 upload → resize image)
✅ Scheduled tasks (cron jobs)
✅ Webhooks
✅ Low-traffic applications

❌ Long-running processes (video encoding)
❌ WebSocket servers (persistent connections)
❌ High-frequency, low-latency workloads
❌ Applications needing local state
```

---

## 🟡 Managed Kubernetes

Kubernetes in the cloud without managing the control plane.

```
Self-managed K8s:              Managed K8s (EKS/GKE/AKS):
  You manage:                    Cloud manages:
  ├── Control plane              ├── Control plane ✅
  │   ├── etcd                   │   ├── etcd
  │   ├── API server             │   ├── API server
  │   ├── Scheduler              │   ├── Scheduler
  │   └── Controller manager     │   └── Controller manager
  └── Worker nodes               You manage:
      ├── kubelet                └── Worker nodes
      ├── kube-proxy                 ├── Node pools
      └── Container runtime         ├── Your pods/deployments
                                     └── Your services
```

```hcl
# EKS cluster in Terraform
resource "aws_eks_cluster" "main" {
  name     = "my-cluster"
  role_arn = aws_iam_role.eks_cluster.arn
  version  = "1.28"
  
  vpc_config {
    subnet_ids = aws_subnet.private[*].id
  }
}

resource "aws_eks_node_group" "workers" {
  cluster_name    = aws_eks_cluster.main.name
  node_group_name = "workers"
  node_role_arn   = aws_iam_role.eks_node.arn
  subnet_ids      = aws_subnet.private[*].id
  
  scaling_config {
    desired_size = 3
    max_size     = 10
    min_size     = 1
  }
  
  instance_types = ["t3.medium"]
}
```

**When to use managed Kubernetes:**
```
✅ Multiple services (microservices)
✅ Need advanced scheduling / auto-scaling
✅ Team already knows Kubernetes
✅ Want portability across clouds
✅ Complex deployment strategies (canary, blue-green)

❌ Single simple application (use Fargate/Cloud Run)
❌ Small team with no K8s experience
❌ Cost-sensitive (K8s overhead is significant)
```

---

## 🟡 Decision Tree

```
"How should I run my application?"

Is it a simple API or web app?
  ├── YES → How much traffic?
  │         ├── Unpredictable → Serverless (Lambda / Cloud Run)
  │         └── Steady        → Container (Fargate / Cloud Run)
  │
  └── NO →  How many services?
            ├── 1-3 services → Containers (ECS / Cloud Run)
            └── 4+ services  → Kubernetes (EKS / GKE / AKS)

Need full OS control?
  → VM (EC2)

Need GPU?
  → VM with GPU (p3/g4 instances)

Need to run for < 15 minutes?
  → Lambda / Cloud Functions

Need persistent connections (WebSocket)?
  → Container or VM (not serverless)
```

---

**Previous:** [01. What Cloud Is](./01-what-cloud-is.md)  
**Next:** [03. Storage Abstractions](./03-storage-abstractions.md)
