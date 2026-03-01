# Networking Abstractions

> **In the cloud, you build your own private network from scratch. VPCs, subnets, route tables, security groups — it's like designing a building's floor plan before moving in.**

---

## 🟢 VPC (Virtual Private Cloud)

A VPC is your own isolated network in the cloud. Nothing gets in or out unless you allow it.

```
AWS Region (us-east-1)
┌──────────────────────────────────────────────────┐
│                    VPC                            │
│              10.0.0.0/16                          │
│    (65,536 IP addresses available)               │
│                                                   │
│  ┌─────────────────┐  ┌─────────────────┐        │
│  │  Public Subnet   │  │  Public Subnet   │        │
│  │  10.0.1.0/24     │  │  10.0.2.0/24     │        │
│  │  (AZ: us-east-1a)│  │  (AZ: us-east-1b)│        │
│  │                   │  │                   │        │
│  │  ┌──────┐        │  │  ┌──────┐        │        │
│  │  │ ALB  │        │  │  │ NAT  │        │        │
│  │  │      │        │  │  │ GW   │        │        │
│  │  └──────┘        │  │  └──────┘        │        │
│  └─────────────────┘  └─────────────────┘        │
│                                                   │
│  ┌─────────────────┐  ┌─────────────────┐        │
│  │  Private Subnet  │  │  Private Subnet  │        │
│  │  10.0.10.0/24    │  │  10.0.20.0/24    │        │
│  │  (AZ: us-east-1a)│  │  (AZ: us-east-1b)│        │
│  │                   │  │                   │        │
│  │  ┌──────┐ ┌────┐ │  │  ┌──────┐ ┌────┐ │        │
│  │  │ App  │ │ App│ │  │  │ App  │ │ DB │ │        │
│  │  │ #1   │ │ #2 │ │  │  │ #3   │ │    │ │        │
│  │  └──────┘ └────┘ │  │  └──────┘ └────┘ │        │
│  └─────────────────┘  └─────────────────┘        │
│                                                   │
└──────────────────────────────────────────────────┘
```

---

## 🟢 Subnets and CIDR Blocks

### CIDR Notation

```
10.0.0.0/16 → First 16 bits are the network → 65,536 addresses
10.0.1.0/24 → First 24 bits are the network → 256 addresses
10.0.1.0/28 → First 28 bits are the network → 16 addresses

Common VPC design:
  VPC: 10.0.0.0/16 (65,536 IPs — "the whole building")
  
  Public subnets (internet-facing):
    10.0.1.0/24  (AZ-a, 256 IPs) — load balancers, bastion hosts
    10.0.2.0/24  (AZ-b, 256 IPs)
    
  Private subnets (no direct internet):
    10.0.10.0/24 (AZ-a, 256 IPs) — application servers
    10.0.20.0/24 (AZ-b, 256 IPs)
    
  Database subnets (most restricted):
    10.0.100.0/24 (AZ-a, 256 IPs) — databases
    10.0.200.0/24 (AZ-b, 256 IPs)
```

### Public vs Private Subnets

```
Public subnet:
  → Has a route to the Internet Gateway (IGW)
  → Instances CAN have public IPs
  → Accessible from the internet (if security group allows)
  → Put here: Load balancers, bastion hosts, NAT gateways

Private subnet:
  → NO route to the Internet Gateway
  → Instances CANNOT have public IPs
  → NOT accessible from the internet
  → Can access internet OUTBOUND through NAT Gateway
  → Put here: Application servers, databases, workers

Why private subnets?
  → Database MUST NOT be accessible from the internet
  → App servers only receive traffic from the load balancer
  → Defense in depth: even if security group misconfigured,
    no internet route = no access
```

---

## 🟢 Security Groups and NACLs

### Security Groups (Stateful Firewall)

```
Security groups are attached to instances/ENIs.
They are STATEFUL: if you allow inbound, the response is automatically allowed.

SG: web-server
┌──────────────────────────────────────────────┐
│ Inbound Rules:                                │
│   Port 80   from 0.0.0.0/0      (HTTP)       │
│   Port 443  from 0.0.0.0/0      (HTTPS)      │
│   Port 22   from 10.0.0.0/16    (SSH — VPC)  │
│                                                │
│ Outbound Rules:                                │
│   All traffic to 0.0.0.0/0      (any)         │
└──────────────────────────────────────────────┘

SG: database
┌──────────────────────────────────────────────┐
│ Inbound Rules:                                │
│   Port 5432 from sg-web-server   (PostgreSQL) │
│   ← Only web servers can connect!             │
│                                                │
│ Outbound Rules:                                │
│   All traffic to 0.0.0.0/0                    │
└──────────────────────────────────────────────┘
```

```hcl
# Terraform
resource "aws_security_group" "web" {
  name        = "web-server"
  description = "Allow HTTP/HTTPS from internet"
  vpc_id      = aws_vpc.main.id
  
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "database" {
  name        = "database"
  description = "Allow PostgreSQL from web servers only"
  vpc_id      = aws_vpc.main.id
  
  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.web.id]  # Reference by SG!
  }
}
```

---

## 🟢 Load Balancers

Distribute traffic across multiple servers.

```
Internet
    │
    ▼
┌────────────┐
│ Application│
│ Load       │  Receives all traffic on port 443
│ Balancer   │  Routes to healthy backend servers
│ (ALB)      │  Handles SSL/TLS termination
└──────┬─────┘
       │  Health checks every 30s
   ┌───┼───────────────┐
   ▼   ▼               ▼
┌─────┐ ┌─────┐  ┌─────┐
│App 1│ │App 2│  │App 3│  Target group
│ :3k │ │ :3k │  │ :3k │  Port 3000
│ ✅  │ │ ✅  │  │ 💀  │  ← Removed from rotation
└─────┘ └─────┘  └─────┘
```

### Types of Load Balancers (AWS)

```
ALB (Application Load Balancer) — Layer 7:
  → HTTP/HTTPS aware
  → Path-based routing (/api → service A, /web → service B)
  → Host-based routing (api.example.com → service A)
  → WebSocket support
  → Use for: Web applications, APIs

NLB (Network Load Balancer) — Layer 4:
  → TCP/UDP level
  → Ultra-high performance (millions of requests/sec)
  → Static IP / Elastic IP
  → Use for: TCP services, gaming, IoT

GLB (Gateway Load Balancer) — Layer 3:
  → For third-party virtual appliances
  → Use for: Firewalls, intrusion detection
```

```hcl
resource "aws_lb" "main" {
  name               = "myapp-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = aws_subnet.public[*].id
}

resource "aws_lb_target_group" "api" {
  name     = "myapp-api"
  port     = 3000
  protocol = "HTTP"
  vpc_id   = aws_vpc.main.id
  
  health_check {
    path                = "/health"
    healthy_threshold   = 2
    unhealthy_threshold = 3
    timeout             = 5
    interval            = 30
  }
}

resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.main.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = aws_acm_certificate.main.arn
  
  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.api.arn
  }
}
```

---

## 🟡 NAT Gateway

Allows private subnet resources to access the internet (for updates, API calls) WITHOUT being accessible FROM the internet.

```
                Internet
                    │
            ┌───────┴───────┐
            │ Internet GW   │
            └───────┬───────┘
                    │
    ┌───────────────┴───────────────┐
    │         Public Subnet         │
    │  ┌─────────────────────────┐  │
    │  │      NAT Gateway        │  │
    │  │    (has public IP)      │  │
    │  └────────────┬────────────┘  │
    └───────────────┼───────────────┘
                    │
    ┌───────────────┼───────────────┐
    │         Private Subnet        │
    │                               │
    │  ┌──────┐  ┌──────┐          │
    │  │ App  │  │ App  │          │
    │  │ (can │  │ (can │          │
    │  │ reach│  │ reach│          │
    │  │ out) │  │ out) │          │
    │  └──────┘  └──────┘          │
    └───────────────────────────────┘
    
    Private instance:
      → CAN do: apt-get update, call external APIs
      → CANNOT be reached from internet
```

```
⚠️ NAT Gateway costs:
  $0.045/hour × 730 hours = ~$32/month per NAT GW
  PLUS $0.045/GB of data processed
  
  Common surprise: NAT Gateway often costs more than the EC2 instances!
  
  For dev/staging: consider NAT instance (cheaper, less reliable)
  For production: use NAT Gateway (managed, highly available)
```

---

## 🟡 CDN (Content Delivery Network)

Cache content at edge locations close to users.

```
Without CDN:
  User in Tokyo → Request travels to us-east-1 → 200ms latency
  
With CDN (CloudFront):
  User in Tokyo → Request goes to Tokyo edge → 10ms latency
  First request: fetches from origin, caches at edge
  Subsequent: served from cache

                    ┌──────────────┐
                    │   Origin     │
                    │  (S3 / ALB)  │
                    └──────┬───────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
         ┌────────┐  ┌────────┐  ┌────────┐
         │ Edge   │  │ Edge   │  │ Edge   │
         │ Tokyo  │  │ London │  │ Sydney │
         └───┬────┘  └───┬────┘  └───┬────┘
             ▼           ▼           ▼
         Users in     Users in    Users in
         Asia         Europe      Australia
```

### Common CDN services:
```
AWS CloudFront
GCP Cloud CDN
Azure Front Door
Cloudflare (cloud-agnostic)
Fastly (developer-focused)
```

---

## 🟡 DNS (Route 53, Cloud DNS)

```
DNS maps domain names to IP addresses.

Cloud DNS adds:
  → Health-based routing (send traffic to healthy servers)
  → Latency-based routing (send to nearest region)
  → Weighted routing (90% to v1, 10% to v2 for canary)
  → Geo-routing (EU users → EU servers for compliance)
```

```hcl
resource "aws_route53_record" "api" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "api.example.com"
  type    = "A"
  
  alias {
    name                   = aws_lb.main.dns_name
    zone_id                = aws_lb.main.zone_id
    evaluate_target_health = true
  }
}
```

---

## 🔴 Common Networking Mistakes

```
❌ Database in public subnet
   → NEVER. Database should be in private subnet with 
     security group allowing only app servers.

❌ Security group allowing 0.0.0.0/0 on port 22
   → Entire internet can SSH. At minimum, restrict to 
     your IP or VPN CIDR.

❌ Single AZ deployment
   → AZ goes down = your app goes down.
     Always deploy across at least 2 AZs.

❌ Ignoring NAT Gateway costs
   → $32+/month per NAT GW. In dev, consider alternatives.

❌ Overly permissive security groups
   → "Allow all traffic" is never the answer.
     Specify exact ports and sources.
```

---

**Previous:** [03. Storage Abstractions](./03-storage-abstractions.md)  
**Next:** [05. Regions and Availability Zones](./05-regions-and-azs.md)
