# Mapping to AWS / GCP / Azure

> **Different clouds, same concepts. Learn the concept once, translate to any provider. Here's the Rosetta Stone.**

---

## 🟢 Service Mapping Table

### Compute

| Concept | AWS | GCP | Azure |
|---------|-----|-----|-------|
| Virtual Machines | EC2 | Compute Engine | Virtual Machines |
| Serverless Functions | Lambda | Cloud Functions | Azure Functions |
| Serverless Containers | Fargate | Cloud Run | Container Instances |
| Managed Kubernetes | EKS | GKE | AKS |
| Container Orchestration | ECS | — | — |
| Auto Scaling | Auto Scaling Groups | Managed Instance Groups | VM Scale Sets |
| Spot/Preemptible | Spot Instances | Preemptible VMs | Spot VMs |

### Storage

| Concept | AWS | GCP | Azure |
|---------|-----|-----|-------|
| Object Storage | S3 | Cloud Storage (GCS) | Blob Storage |
| Block Storage | EBS | Persistent Disks | Managed Disks |
| File Storage | EFS | Filestore | Azure Files |
| Archive Storage | S3 Glacier | Archive Storage | Archive Storage |
| Local SSD | Instance Store | Local SSD | Temp Disk |

### Networking

| Concept | AWS | GCP | Azure |
|---------|-----|-----|-------|
| Virtual Network | VPC | VPC | VNet |
| Firewall Rules | Security Groups | Firewall Rules | NSG |
| Load Balancer (L7) | ALB | Cloud Load Balancing | Application Gateway |
| Load Balancer (L4) | NLB | Network LB | Azure LB |
| CDN | CloudFront | Cloud CDN | Front Door / CDN |
| DNS | Route 53 | Cloud DNS | Azure DNS |
| Private Connectivity | VPC Peering / Transit GW | VPC Peering | VNet Peering |
| VPN | Site-to-Site VPN | Cloud VPN | VPN Gateway |

### Database

| Concept | AWS | GCP | Azure |
|---------|-----|-----|-------|
| Managed PostgreSQL/MySQL | RDS | Cloud SQL | Azure Database |
| Managed NoSQL | DynamoDB | Firestore / Bigtable | Cosmos DB |
| Managed Redis | ElastiCache | Memorystore | Azure Cache |
| Data Warehouse | Redshift | BigQuery | Synapse Analytics |
| Managed MongoDB | DocumentDB | — | Cosmos DB (Mongo API) |

### Messaging & Queues

| Concept | AWS | GCP | Azure |
|---------|-----|-----|-------|
| Message Queue | SQS | Cloud Tasks | Azure Queue |
| Pub/Sub | SNS | Pub/Sub | Service Bus |
| Event Streaming | Kinesis | Pub/Sub | Event Hubs |
| Event Bus | EventBridge | Eventarc | Event Grid |

### Identity & Security

| Concept | AWS | GCP | Azure |
|---------|-----|-----|-------|
| Identity (IAM) | IAM | IAM | Azure AD / Entra ID |
| Secrets Manager | Secrets Manager | Secret Manager | Key Vault |
| Certificate Manager | ACM | Certificate Manager | — |
| Key Management | KMS | Cloud KMS | Key Vault |

### DevOps & CI/CD

| Concept | AWS | GCP | Azure |
|---------|-----|-----|-------|
| Container Registry | ECR | Artifact Registry | ACR |
| CI/CD | CodePipeline | Cloud Build | Azure DevOps |
| IaC | CloudFormation | Deployment Manager | ARM / Bicep |
| Monitoring | CloudWatch | Cloud Monitoring | Azure Monitor |
| Logging | CloudWatch Logs | Cloud Logging | Log Analytics |

### Observability

| Concept | AWS | GCP | Azure |
|---------|-----|-----|-------|
| Metrics | CloudWatch Metrics | Cloud Monitoring | Azure Monitor Metrics |
| Logs | CloudWatch Logs | Cloud Logging | Log Analytics |
| Tracing | X-Ray | Cloud Trace | Application Insights |
| Dashboards | CloudWatch Dashboards | Cloud Monitoring Dashboards | Azure Dashboards |

---

## 🟡 Key Differences

### Networking Model

```
AWS:
  → VPCs are regional (span all AZs in a region)
  → Subnets are per-AZ
  → Security Groups attached to instances
  → Must explicitly create Internet Gateway, NAT Gateway
  → Cross-AZ traffic costs money

GCP:
  → VPCs are GLOBAL (span all regions!)
  → Subnets are regional
  → Firewall rules are VPC-wide
  → Simpler networking model
  → Cross-AZ traffic within region is FREE

Azure:
  → VNets are regional
  → Subnets span AZs within a region
  → NSGs can be attached to subnets or NICs
  → Similar to AWS but different terminology
```

### IAM Model

```
AWS:
  → Users, Groups, Roles, Policies
  → Policies = JSON documents
  → Roles for cross-account access
  → Roles for service-to-service
  
GCP:
  → Members (users, service accounts)
  → Roles (predefined or custom)
  → Bound at project/folder/org level
  → Simpler but less granular

Azure:
  → Azure AD (Entra ID) for users
  → RBAC for resource access
  → Managed Identities for services
  → Tied to Active Directory
```

### Kubernetes

```
EKS (AWS):
  → Most popular managed K8s
  → Add-ons as separate installs
  → AWS IAM integration (IRSA)
  → Slightly more setup required
  
GKE (GCP):
  → Google invented Kubernetes
  → Best managed K8s experience
  → Autopilot mode (fully managed nodes)
  → Fastest cluster creation

AKS (Azure):
  → Good integration with Azure AD
  → Free control plane
  → Azure Policy for governance
  → Most enterprise-friendly
```

---

## 🟡 Cloud-Agnostic Strategies

### When to Stay Cloud-Agnostic

```
✅ Use Kubernetes (runs on any cloud)
✅ Use Terraform (manages any cloud)
✅ Use standard protocols (PostgreSQL, Redis, HTTP)
✅ Use containers (portable by definition)
✅ Avoid proprietary services when generic alternatives exist

☁️ BUT: Using cloud-native services is usually better
   → RDS > self-managed PostgreSQL on EC2
   → S3 > self-managed MinIO
   → SQS > self-managed RabbitMQ
   
Pragmatic approach:
  Use managed services WHERE THEY MATTER
  + Kubernetes for compute (portable)
  + Terraform for infrastructure (portable config)
  + Standard databases (PostgreSQL, Redis)
```

### Multi-Cloud Terraform

```hcl
# Same Terraform workspace, different providers
provider "aws" {
  region = "us-east-1"
}

provider "google" {
  project = "my-project"
  region  = "us-central1"
}

# DNS on AWS
resource "aws_route53_record" "api" {
  name    = "api.example.com"
  type    = "A"
  zone_id = aws_route53_zone.main.zone_id
  # ...
}

# Kubernetes on GCP
resource "google_container_cluster" "main" {
  name     = "my-cluster"
  location = "us-central1"
  # ...
}
```

---

## 🟡 Decision Framework

```
Choosing a Cloud Provider:

"My team knows AWS already"
  → AWS (don't switch for no reason)

"I need the best Kubernetes experience"
  → GCP (Google made K8s)

"I'm a Microsoft shop (Active Directory, .NET)"
  → Azure

"I want the most services and ecosystem"
  → AWS (broadest service catalog)

"I want simplest pricing"
  → GCP (per-second billing, sustained discounts)

"I need to comply with European data laws"
  → Any of them (all have EU regions)

"I want the cheapest option"
  → Depends on workload. All are comparable.
  → Reserved instances / committed use discounts matter more
     than which provider you choose.

"I'm a startup with no preference"
  → AWS (most documentation, most Stack Overflow answers,
     most job candidates know it)
```

---

## 🔴 Vendor Lock-In: Real Talk

```
High lock-in (hard to migrate):
  ❌ Lambda functions → completely AWS-specific
  ❌ DynamoDB → proprietary API
  ❌ CloudFormation → AWS-only IaC
  ❌ SQS/SNS → AWS-specific messaging
  ❌ Azure Functions → Azure-specific
  
Low lock-in (easy to migrate):
  ✅ Kubernetes → runs anywhere
  ✅ PostgreSQL on RDS → migrate to Cloud SQL/Azure DB
  ✅ Terraform → change provider, same concepts
  ✅ S3 (standard API) → every cloud has object storage
  ✅ Docker containers → run anywhere

Honest assessment:
  Lock-in risk is usually OVERESTIMATED.
  Migration cost is usually UNDERESTIMATED.
  
  Most companies will never migrate clouds.
  Optimize for productivity, not theoretical portability.
```

---

**Previous:** [05. Regions and AZs](./05-regions-and-azs.md)  
**Next:** [Module 13: Observability](../13-observability/README.md)
