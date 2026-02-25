# Security Groups, NACLs, and Load Balancers

> Security in the cloud is about layers. Security Groups protect individual resources. NACLs protect subnets. Load balancers distribute traffic. Together, they form the traffic control system for your cloud network. Get them right, and your infrastructure is rock solid. Get them wrong, and you're either wide open or completely locked out.

---

## Table of Contents

1. [Security Groups — Instance-Level Firewalls](#security-groups)
2. [Security Group Rules Deep Dive](#sg-rules)
3. [Security Group Patterns](#sg-patterns)
4. [Network ACLs — Subnet-Level Firewalls](#nacls)
5. [Security Groups vs NACLs](#sg-vs-nacl)
6. [Load Balancers — Traffic Distribution](#load-balancers)
7. [Application Load Balancer (ALB)](#alb)
8. [Network Load Balancer (NLB)](#nlb)
9. [ALB vs NLB — When to Use Which](#alb-vs-nlb)
10. [Load Balancer + Security Group Architecture](#lb-sg-architecture)
11. [End-to-End Cloud Architecture Example](#full-architecture)
12. [Key Takeaways](#key-takeaways)

---

## Security Groups — Instance-Level Firewalls

### What Security Groups are

```
A Security Group is a STATEFUL firewall attached to 
a network interface (ENI) — not to an instance, not to a subnet.

  ┌─────────────────────────────────┐
  │ EC2 Instance                     │
  │         ┌──────────┐            │
  │         │ App      │            │
  │         │ :8080    │            │
  │         └────┬─────┘            │
  │              │                   │
  │    ┌─────────┴──────────┐       │
  │    │ Security Group      │ ← Evaluated HERE  │
  │    │ • Allow TCP 443 in  │   (at the virtual  │
  │    │ • Allow all out     │    NIC level)       │
  │    └─────────┬──────────┘       │
  │              │                   │
  │         ┌────┴─────┐            │
  │         │ ENI      │            │
  │         │ eth0     │            │
  │         └──────────┘            │
  └─────────────────────────────────┘

Key properties:
  1. STATEFUL: If you allow inbound, return traffic automatically allowed
  2. ALLOW only: You can only write allow rules. No deny rules.
  3. Default deny: Everything not explicitly allowed is denied
  4. Applied at the interface: Enforced at the hypervisor/Nitro level
  5. Multiple SGs per ENI: Rules are combined (union)
  6. An SG can reference another SG as source/destination
```

### Stateful — what it means

```
STATEFUL means the firewall tracks connections:

  Inbound rule: Allow TCP 443 from 0.0.0.0/0
  
  1. Client sends SYN to :443 → ALLOWED (matches rule)
  2. Server sends SYN-ACK back → ALLOWED (tracked connection)
  3. Client sends ACK → ALLOWED (tracked)
  4. Application data flows → ALLOWED (tracked)
  
  You did NOT need an outbound rule for the response.
  The SG remembers: "I let this connection in, 
  so return traffic is automatically allowed."

  Similarly:
  Outbound rule: Allow all (default)
  
  1. Instance sends request to 8.8.8.8:443 → ALLOWED (outbound rule)
  2. Response comes back → ALLOWED (tracked outbound connection)
  
  You did NOT need an inbound rule for the response.

Compare to NACLs (stateless): you'd need rules for BOTH directions.
```

---

## Security Group Rules Deep Dive

### Rule anatomy

```
Each rule has:
  - Type: TCP, UDP, ICMP, or All Traffic
  - Protocol: TCP(6), UDP(17), ICMP(1), All(-1)
  - Port range: 443, 8080-8090, or All
  - Source/Destination: CIDR, Security Group ID, or Prefix List
  - Description: (optional but strongly recommended)

Example rules (inbound):
  ┌───────────┬──────────┬───────────────────────┬───────────────────┐
  │ Type      │ Port     │ Source                │ Description       │
  ├───────────┼──────────┼───────────────────────┼───────────────────┤
  │ HTTPS     │ 443      │ 0.0.0.0/0             │ Public web traffic│
  │ SSH       │ 22       │ 10.0.0.0/16           │ VPC-only SSH      │
  │ Custom TCP│ 8080     │ sg-alb-12345          │ From ALB only     │
  │ PostgreSQL│ 5432     │ sg-app-67890          │ From app tier     │
  │ All ICMP  │ All      │ 10.0.0.0/16           │ VPC ping          │
  └───────────┴──────────┴───────────────────────┴───────────────────┘
```

### Referencing other Security Groups

```
This is the MOST POWERFUL feature of Security Groups.

Instead of:  Allow port 8080 from 10.0.1.0/24  (CIDR — brittle)
Use:         Allow port 8080 from sg-alb-12345  (SG reference — dynamic)

Why SG references are better:
  - CIDRs change as you scale (new subnets, IPs)
  - SG references automatically include ANY instance in that SG
  - If ALB moves to a new subnet → still works
  - If you add 10 more ALB instances → still works

Example architecture:
  ┌───────────────┐     ┌───────────────┐     ┌───────────────┐
  │ ALB            │     │ App Instances  │     │ Database       │
  │ sg-alb         │────>│ sg-app         │────>│ sg-db          │
  │                │     │                │     │                │
  │ Inbound:       │     │ Inbound:       │     │ Inbound:       │
  │ 443 from       │     │ 8080 from      │     │ 5432 from      │
  │ 0.0.0.0/0      │     │ sg-alb ✓       │     │ sg-app ✓       │
  └───────────────┘     └───────────────┘     └───────────────┘

  Traffic can ONLY flow: Internet → ALB → App → DB
  App can't be accessed directly from internet.
  DB can't be accessed except from app instances.
  
  No CIDRs. No IP addresses. Just SG references.
  Scales automatically. This is the correct pattern.
```

---

## Security Group Patterns

### Pattern 1: Web application tier

```
sg-alb (Load Balancer):
  Inbound:  TCP 443 from 0.0.0.0/0
  Inbound:  TCP 80  from 0.0.0.0/0 (redirect to HTTPS)
  Outbound: All traffic

sg-app (Application):
  Inbound:  TCP 8080 from sg-alb
  Inbound:  TCP 22   from sg-bastion (SSH for debugging)
  Outbound: All traffic

sg-db (Database):
  Inbound:  TCP 5432 from sg-app
  Outbound: All traffic (or restrict to sg-app only)

sg-bastion (Jump box):
  Inbound:  TCP 22 from <your office IP>/32
  Outbound: All traffic
```

### Pattern 2: Kubernetes workers

```
sg-k8s-nodes:
  Inbound:  TCP 443   from sg-alb              (ingress)
  Inbound:  TCP 10250 from sg-k8s-nodes        (kubelet)
  Inbound:  TCP 30000-32767 from sg-alb        (NodePort)
  Inbound:  All from sg-k8s-nodes              (pod-to-pod)
  Inbound:  TCP 443   from sg-k8s-control      (API server)
  Outbound: All traffic

  NOTE: VPC CNI uses the node's SG for pods.
  With security groups for pods (SGP), individual pods 
  can have their own Security Groups.
```

### Pattern 3: Microservices

```
Instead of one big SG, create per-service SGs:

  sg-user-service:
    Inbound: TCP 8080 from sg-api-gateway
    Inbound: TCP 8080 from sg-order-service

  sg-order-service:
    Inbound: TCP 8080 from sg-api-gateway
    
  sg-payment-service:
    Inbound: TCP 8080 from sg-order-service
    Inbound: TCP 8080 from sg-billing-service

  Each microservice has its own SG.
  Dependencies are explicitly encoded in SG rules.
  You can read the architecture from the SG graph.
```

---

## Network ACLs — Subnet-Level Firewalls

### What NACLs are

```
NACL = Network Access Control List.
A STATELESS firewall applied to an entire subnet.

  ┌────────────────────────────────┐
  │ Subnet 10.0.1.0/24            │
  │                                │
  │ ┌──────────┐ ┌──────────┐     │
  │ │ Instance │ │ Instance │     │
  │ │ A         │ │ B         │   │
  │ └──────────┘ └──────────┘     │
  │                                │
  └────────────┬───────────────────┘
               │
  ┌────────────┴───────────────────┐
  │ NACL                            │ ← Traffic filtered HERE
  │                                │   (at the subnet boundary)
  │ Rule#  Type   Port  Source  Act│
  │ 100    TCP    443   0/0     ALW│
  │ 200    TCP    22    10/16   ALW│
  │ *      All    All   0/0     DNY│  ← default deny
  └────────────────────────────────┘

Key properties:
  1. STATELESS: Must explicitly allow BOTH inbound AND outbound
  2. Has ALLOW and DENY rules
  3. Rules are numbered: evaluated in order, first match wins
  4. One NACL per subnet (but one NACL can serve multiple subnets)
  5. Default NACL: allows all traffic (most people leave it)
```

### NACL rules — stateless means pain

```
Want to allow HTTPS in? You need TWO rules:

Inbound:
  Rule 100: Allow TCP 443 from 0.0.0.0/0    ← request comes in

Outbound:
  Rule 100: Allow TCP 1024-65535 to 0.0.0.0/0  ← response goes out
                                                  (ephemeral ports!)

Why ephemeral ports?
  Client connects FROM a random high port (e.g., 52341).
  Server responds TO that port.
  Since NACL is stateless, it doesn't know about the connection.
  You must allow the full ephemeral port range for responses.
  
  Linux ephemeral range: 32768-60999
  Windows ephemeral range: 49152-65535
  Safe NACL range: 1024-65535 (covers all OS variants)

This is annoying. This is why most people use the default 
(allow-all) NACL and rely on Security Groups for filtering.
```

---

## Security Groups vs NACLs

```
Feature              Security Group          NACL
──────────           ──────────────          ────
Level                Instance (ENI)          Subnet
State                Stateful                Stateless
Rules                Allow only              Allow AND Deny
Default              Deny all inbound        Allow all (default NACL)
Rule order           All rules evaluated     Numbered, first match wins
Return traffic       Automatic               Must be explicitly allowed
Can reference SGs    Yes                     No (CIDRs only)

When to use what:

  Security Groups: ALWAYS. Primary firewall. Use SG references.
  
  NACLs: Rarely. Use for:
    - Blocking specific IPs (deny rules)
    - Subnet-level blocking (e.g., block a compromised CIDR)
    - Compliance requirements ("defense in depth")
    - Emergency: block an attacking IP range

Real-world practice:
  95% of traffic control → Security Groups
  5% of traffic control → NACLs (edge cases, deny rules)
  
  Most teams leave NACLs at default (allow all) 
  and manage everything with Security Groups.
```

### Defense in depth illustration

```
Traffic from internet to your app:

  ┌──── Internet ────┐
  │                   │
  │ ┌───────────────┐ │
  │ │ NACL (subnet) │ │  ← Layer 1: Subnet boundary
  │ │ Allow 443 in  │ │     Deny known bad CIDRs
  │ │ Allow ephem out│ │     Block entire IP ranges
  │ └───────┬───────┘ │
  │         │         │
  │ ┌───────┴───────┐ │
  │ │ Security Group│ │  ← Layer 2: Instance boundary
  │ │ Allow 443 from│ │     Fine-grained: SG references
  │ │ sg-alb        │ │     Stateful: auto return traffic
  │ └───────┬───────┘ │
  │         │         │
  │ ┌───────┴───────┐ │
  │ │ OS Firewall   │ │  ← Layer 3: Inside the instance
  │ │ (iptables)    │ │     Last line of defense
  │ └───────────────┘ │
  └───────────────────┘

Three layers. Any one can save you if the others fail.
```

---

## Load Balancers — Traffic Distribution

### Why load balancers in the cloud

```
Without LB:
  DNS round-robin to 3 instances:
  api.example.com → 10.0.1.10, 10.0.1.11, 10.0.1.12
  
  Problems:
    - DNS caches: can't remove unhealthy instance fast
    - No health checks: sends traffic to dead instances
    - No session affinity
    - DNS TTL delays
    - Client gets ALL IPs (security concern)

With LB:
  api.example.com → ALB (single endpoint)
  ALB → healthy instances only
  
  Benefits:
    - Health checks: automatically removes unhealthy targets
    - Even distribution: weighted, round-robin, least connections
    - TLS termination: offload crypto from app servers
    - Single endpoint: clients don't know about backends
    - Scaling: add/remove targets dynamically
```

---

## Application Load Balancer (ALB)

### How ALB works

```
ALB operates at Layer 7 (HTTP/HTTPS).

  ┌────────────────────────────────────────────┐
  │ ALB                                         │
  │                                             │
  │ Listener: HTTPS 443                          │
  │   ├── Rule: Host = api.example.com           │
  │   │   └── Forward to: tg-api (Target Group) │
  │   ├── Rule: Host = web.example.com           │
  │   │   └── Forward to: tg-web                │
  │   ├── Rule: Path = /health                   │
  │   │   └── Return 200 (fixed response)        │
  │   └── Default: Return 404                    │
  │                                             │
  │ Target Groups:                               │
  │   tg-api: [10.0.2.10:8080, 10.0.2.11:8080]  │
  │   tg-web: [10.0.2.20:3000, 10.0.2.21:3000]  │
  └────────────────────────────────────────────┘

ALB components:
  Listener → Rules → Target Group → Targets (instances/pods/IPs)
```

### ALB routing capabilities

```
Route based on:
  - Host header:   api.example.com vs web.example.com
  - Path:          /api/* vs /static/* vs /health
  - HTTP method:   GET vs POST
  - Query string:  ?version=2
  - HTTP headers:  Custom-Header: value
  - Source IP:     From specific CIDR

This enables:
  Single ALB, multiple microservices.
  
  /api/users/*    → user-service
  /api/orders/*   → order-service
  /api/payments/* → payment-service
  /health         → fixed 200 response
  /*              → web-frontend
```

### ALB health checks

```
ALB continuously checks backend health:

  Health check configuration:
    Protocol: HTTP
    Path: /health
    Port: 8080
    Interval: 30 seconds
    Timeout: 5 seconds
    Healthy threshold: 3 consecutive successes
    Unhealthy threshold: 2 consecutive failures
    Success codes: 200-299

  ┌──────┐      /health       ┌──────────┐
  │ ALB  │───────────────────>│ Target   │
  │      │<─── 200 OK ────────│ (healthy)│
  │      │                    └──────────┘
  │      │      /health       ┌──────────┐
  │      │───────────────────>│ Target   │
  │      │<─── timeout ───────│ (failing)│
  │      │ Remove from pool   └──────────┘

  Unhealthy targets stop receiving traffic.
  When they recover, they're automatically added back.
  
  This is why your app MUST have a /health endpoint.
  It's not optional. The LB depends on it.
```

### ALB connection behavior

```
ALB is a FULL PROXY (Layer 7):

  Client ←── TCP+TLS ──→ ALB ←── TCP ──→ Backend

  Two separate TCP connections.
  ALB terminates the client's TLS.
  ALB opens a new connection to the backend.
  (Backend can optionally use TLS too: end-to-end encryption)

Headers added by ALB:
  X-Forwarded-For: 203.0.113.50     (client's real IP)
  X-Forwarded-Proto: https           (original protocol)
  X-Forwarded-Port: 443              (original port)

  Your app MUST read X-Forwarded-For for the real client IP.
  Otherwise you'll see the ALB's private IP as the "client."

Idle timeout:
  Default: 60 seconds
  If no data for 60s → ALB closes the connection
  
  CRITICAL: Backend keep-alive timeout MUST be > ALB idle timeout.
  If backend closes at 55s and ALB tries at 58s → 502 error.
  See Module 18 (Real-World Failures) for this failure pattern.
```

---

## Network Load Balancer (NLB)

### How NLB works

```
NLB operates at Layer 4 (TCP/UDP).

  ┌──────────────────────────────────┐
  │ NLB                               │
  │                                   │
  │ Does NOT inspect HTTP headers     │
  │ Does NOT terminate TLS*           │
  │ Does NOT read Host/Path           │
  │                                   │
  │ Just forwards TCP connections     │
  │ to healthy targets.               │
  │                                   │
  │ * TLS termination optional (TLS   │
  │   listener, added later)          │
  └──────────────────────────────────┘

NLB is a Layer 4 pass-through:
  Client ←──── TCP ────→ NLB ←──── TCP ────→ Backend
  
  NLB preserves the client's source IP (by default).
  Backend sees the real client IP directly.
  No X-Forwarded-For needed.
  
  NLB has static IPs (or Elastic IPs).
  ALB has dynamic IPs (DNS name only).
```

### NLB performance

```
NLB is designed for extreme performance:
  - Millions of connections per second
  - Ultra-low latency (~100 microseconds added)
  - No processing overhead (no HTTP parsing)
  - Static IPs for whitelisting
  
  ALB:     ~5ms added latency (HTTP parsing + routing)
  NLB:     ~0.1ms added latency (TCP forwarding only)

Use NLB for:
  - gRPC/HTTP2 where you handle TLS at the app
  - TCP protocols (databases, MQTT, custom)
  - Extreme throughput requirements
  - When clients need static IPs to whitelist
  - UDP workloads (DNS, gaming, IoT)
```

---

## ALB vs NLB — When to Use Which

```
Feature              ALB              NLB
──────────           ───              ───
OSI Layer            7 (HTTP)         4 (TCP/UDP)
Routing              Host/Path/Header IP:Port only
TLS termination      Always           Optional
Client IP to backend X-Forwarded-For  Preserved directly
Static IP            No (DNS only)    Yes (Elastic IPs)
Performance          Good (~5ms)      Extreme (~0.1ms)
WebSockets           Yes              Yes (passthrough)
gRPC                 Yes              Yes (passthrough)
Cost                 ~$16/mo + LCU    ~$16/mo + NLCU
Health checks        HTTP/HTTPS       TCP/HTTP/HTTPS

Decision tree:
  Need host/path routing?        → ALB
  Need to inspect HTTP headers?  → ALB
  Need static IP?                → NLB
  Need non-HTTP protocol?        → NLB (TCP/UDP)
  Need ultra-low latency?        → NLB
  General web application?       → ALB (most common choice)
  
Most web applications: ALB.
Most infrastructure services: NLB.
```

---

## Load Balancer + Security Group Architecture

### Complete setup

```
┌──────────────────────────────────────────────────────────────┐
│ VPC 10.0.0.0/16                                              │
│                                                              │
│ Public Subnets (10.0.1.0/24, 10.0.3.0/24)                   │
│ ┌──────────────────────────────────────────┐                 │
│ │                                          │                 │
│ │         ┌──────────────┐                 │                 │
│ │  IGW ───│ ALB          │                 │                 │
│ │         │ sg-alb:      │                 │                 │
│ │         │  in: 443/0.0 │                 │                 │
│ │         │  out: all     │                │                 │
│ │         └──────┬───────┘                 │                 │
│ └────────────────┼─────────────────────────┘                 │
│                  │                                           │
│ Private Subnets (10.0.2.0/24, 10.0.4.0/24)                  │
│ ┌────────────────┼─────────────────────────┐                 │
│ │         ┌──────┴───────┐                 │                 │
│ │         │ App Instances │                 │                 │
│ │         │ sg-app:       │                 │                 │
│ │         │  in: 8080     │                 │                 │
│ │         │  from sg-alb  │                 │                 │
│ │         │  out: all     │                 │                 │
│ │         └──────┬───────┘                 │                 │
│ │                │                          │                 │
│ │         ┌──────┴───────┐                 │                 │
│ │         │ RDS Database  │                 │                 │
│ │         │ sg-db:        │                 │                 │
│ │         │  in: 5432     │                 │                 │
│ │         │  from sg-app  │                 │                 │
│ │         │  out: all     │                 │                 │
│ │         └──────────────┘                 │                 │
│ └──────────────────────────────────────────┘                 │
│                                                              │
│ NAT GW in public subnet for outbound from private            │
│ S3 Gateway Endpoint for S3 access (free)                     │
│ Interface Endpoint for other AWS services                    │
└──────────────────────────────────────────────────────────────┘

Traffic flow:
  Internet → IGW → ALB (public subnet, sg-alb allows 443)
  ALB → App (private subnet, sg-app allows 8080 from sg-alb)
  App → DB (same VPC, sg-db allows 5432 from sg-app)
  App → S3 (Gateway Endpoint, no NAT GW needed)
  App → External API (NAT GW → IGW → Internet)
```

---

## End-to-End Cloud Architecture Example

### Production-grade setup

```
Region: us-east-1
VPC: 10.0.0.0/16

AZ: us-east-1a                    AZ: us-east-1b
┌─────────────────────┐           ┌─────────────────────┐
│ Public 10.0.1.0/24  │           │ Public 10.0.3.0/24  │
│  • ALB nodes        │           │  • ALB nodes        │
│  • NAT Gateway A    │           │  • NAT Gateway B    │
├─────────────────────┤           ├─────────────────────┤
│ App 10.0.10.0/20    │           │ App 10.0.16.0/20    │
│  • EKS worker nodes │           │  • EKS worker nodes │
│  • (pods consume IPs)│          │  • (pods consume IPs)│
├─────────────────────┤           ├─────────────────────┤
│ Data 10.0.20.0/24   │           │ Data 10.0.22.0/24   │
│  • RDS primary       │           │  • RDS standby       │
│  • ElastiCache       │           │  • ElastiCache       │
└─────────────────────┘           └─────────────────────┘

Connectivity:
  • Internet: IGW + ALB (public subnets)
  • Outbound: NAT GW per AZ (HA)
  • S3: Gateway Endpoint (free)
  • ECR/SQS: Interface Endpoints
  • On-prem: Transit Gateway + VPN

Security:
  • ALB SG: 443 from internet
  • Worker SG: 8080 from ALB SG, all from self (pod mesh)
  • DB SG: 5432 from worker SG only
  • NACLs: default (allow all), deny known bad CIDRs
  • Subnet isolation: DB only reachable from app tier

DNS:
  • Public: api.example.com → ALB DNS name (CNAME/Alias)
  • Private: *.internal.example.com → private IPs
  • Pod DNS: CoreDNS → VPC DNS resolver

Monitoring:
  • VPC Flow Logs → CloudWatch/S3
  • ALB access logs → S3
  • Security Group changes → CloudTrail
```

### Terraform sketch for the Security Groups

```hcl
# ALB Security Group
resource "aws_security_group" "alb" {
  name_prefix = "alb-"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTPS from internet"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# Application Security Group
resource "aws_security_group" "app" {
  name_prefix = "app-"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
    description     = "HTTP from ALB only"
  }

  # Allow pod-to-pod communication
  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
    description = "All traffic within app tier"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# Database Security Group
resource "aws_security_group" "db" {
  name_prefix = "db-"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.app.id]
    description     = "PostgreSQL from app tier only"
  }

  # No egress to internet needed
  egress {
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    security_groups = [aws_security_group.app.id]
    description     = "Responses to app tier"
  }
}
```

---

## Key Takeaways

1. **Security Groups are stateful, allow-only, instance-level firewalls** — they track connections automatically, so you don't need explicit rules for return traffic. Everything not explicitly allowed is denied
2. **Reference Security Groups instead of CIDRs** — `allow port 8080 from sg-alb` is dynamic, self-healing, and scales automatically. CIDRs are static and break when infrastructure changes
3. **NACLs are stateless, subnet-level firewalls with allow AND deny rules** — you must explicitly allow BOTH directions including ephemeral ports for return traffic. Most teams leave them at default (allow all)
4. **Use Security Groups for 95% of access control, NACLs for deny rules** — need to block a specific IP range? NACL. Everything else? Security Group
5. **ALB operates at Layer 7 (HTTP) with host/path routing** — it terminates TLS, inspects HTTP headers, and can route to different target groups based on URL patterns. Most web apps use ALB
6. **NLB operates at Layer 4 (TCP/UDP) with extreme performance** — it preserves client IPs, supports static IPs, and adds minimal latency. Use for non-HTTP protocols or when you need static IPs
7. **ALB is a full proxy — two separate TCP connections** — client→ALB and ALB→backend. The backend sees the ALB's IP unless it reads X-Forwarded-For. Keep-alive timeouts must be longer on the backend than on the ALB
8. **Health checks are not optional** — load balancers continuously check /health on your backends. Unhealthy targets are automatically removed. Your app MUST implement a health check endpoint
9. **The three-tier security pattern is: SG-ALB → SG-App → SG-DB** — each tier only accepts traffic from the tier above it, using SG references. This encodes your architecture in firewall rules
10. **VPC Flow Logs capture all traffic metadata** — they record every accepted and rejected connection at the ENI level. Essential for debugging connectivity issues and security auditing

---

## Course Complete

You've followed a packet from the user's keyboard through every layer of the networking stack, across the internet, through cloud infrastructure, and back. You understand DNS, TCP, TLS, HTTP, routing, NAT, firewalls, load balancers, containers, Kubernetes networking, and cloud networking.

The rest is practice. Debug real problems. Read packet captures. Break things and fix them.

→ Return to [Course Overview](../README.md)
