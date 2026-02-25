# Subnets, Routing, and Gateways

> If VPCs are your private network, subnets are the rooms inside it. Routing tables are the hallway signs that tell traffic where to go. And gateways are the doors that connect your private network to the outside world. Get any of these wrong, and traffic either can't get in, can't get out, or goes somewhere it shouldn't.

---

## Table of Contents

1. [Subnets In Depth](#subnets)
2. [Public vs Private Subnets](#public-private)
3. [Route Tables — How Traffic Decisions Are Made](#route-tables)
4. [Internet Gateway (IGW)](#igw)
5. [NAT Gateway](#nat-gateway)
6. [Elastic IPs and Public IPs](#elastic-ips)
7. [VPC Endpoints — Private Access to Cloud Services](#vpc-endpoints)
8. [DNS Inside a VPC](#vpc-dns)
9. [VPN and Direct Connect](#vpn-direct-connect)
10. [Complete Traffic Flow Examples](#traffic-flows)
11. [Common Misconfigurations](#misconfigurations)
12. [Key Takeaways](#key-takeaways)

---

## Subnets In Depth

### What a subnet really is

```
A subnet is a range of IP addresses within your VPC, 
tied to a specific Availability Zone.

VPC: 10.0.0.0/16 (65,536 IPs)
  ├── Subnet-A: 10.0.1.0/24 (256 IPs) → AZ us-east-1a
  ├── Subnet-B: 10.0.2.0/24 (256 IPs) → AZ us-east-1b
  ├── Subnet-C: 10.0.10.0/24 (256 IPs) → AZ us-east-1a
  └── Subnet-D: 10.0.20.0/24 (256 IPs) → AZ us-east-1b

Key rules:
  1. A subnet exists in exactly ONE AZ
  2. A subnet has exactly ONE route table
  3. Subnets in the same VPC CAN communicate (by default)
  4. Reserved IPs per subnet (AWS):
     .0 = network, .1 = router, .2 = DNS, .3 = reserved, .255 = broadcast
     → /24 gives 251 usable IPs, not 254
```

### Subnet sizing

```
CIDR     IPs    Usable (AWS)   Use case
────     ───    ────────────   ────────
/28      16     11             Tiny: NAT Gateway, single resources
/26      64     59             Small: load balancers
/24      256    251            Standard: most workloads
/22      1024   1019           Large: EKS nodes with many pods
/20      4096   4091           Very large: high pod density
/19      8192   8187           Maximum: Kubernetes + service mesh

Common mistake: Making subnets too small.
  EKS with VPC CNI needs an IP per pod.
  100 nodes × 30 pods = 3,000 IPs.
  A /24 (251 IPs) won't work. Need at least /20.

Rules of thumb:
  - Use /24 for most workloads
  - Use /20 or /19 for Kubernetes with VPC CNI
  - Leave room to add more subnets later
  - Don't use the entire VPC CIDR for subnets
```

---

## Public vs Private Subnets

### The ONLY difference

```
A subnet is "public" or "private" based SOLELY on its route table.

Public subnet route table:
  ┌────────────────────────────────────────┐
  │ Destination     Target                 │
  │ ──────────────  ──────                 │
  │ 10.0.0.0/16     local         ← VPC   │
  │ 0.0.0.0/0       igw-abc123   ← IGW!  │
  └────────────────────────────────────────┘
  
  0.0.0.0/0 → Internet Gateway = PUBLIC
  Instances CAN have public IPs.
  Traffic to the internet goes directly through IGW.

Private subnet route table:
  ┌────────────────────────────────────────┐
  │ Destination     Target                 │
  │ ──────────────  ──────                 │
  │ 10.0.0.0/16     local         ← VPC   │
  │ 0.0.0.0/0       nat-xyz789   ← NAT!  │
  └────────────────────────────────────────┘
  
  0.0.0.0/0 → NAT Gateway = PRIVATE
  Instances have private IPs only.
  Outbound traffic goes through NAT.
  NO inbound traffic from internet.

Isolated subnet (no internet at all):
  ┌────────────────────────────────────────┐
  │ Destination     Target                 │
  │ ──────────────  ──────                 │
  │ 10.0.0.0/16     local         ← VPC   │
  │ (no default route)                     │
  └────────────────────────────────────────┘
  
  No path to internet. Completely isolated.
  Can still reach other subnets in the VPC.
```

---

## Route Tables — How Traffic Decisions Are Made

### Route table mechanics

```
Every subnet has exactly one route table.
Multiple subnets can share a route table.

Route evaluation:
  Most specific route wins (longest prefix match).
  
  Example route table:
    10.0.0.0/16     local
    10.0.1.0/24     pcx-abc (peering)
    10.1.0.0/16     tgw-xyz (transit gateway)
    0.0.0.0/0       igw-123 (internet gateway)
  
  Packet to 10.0.1.50:
    Matches 10.0.0.0/16 (/16) AND 10.0.1.0/24 (/24)
    /24 is more specific → goes to peering connection ✓
  
  Packet to 10.1.5.10:
    Matches 10.1.0.0/16 → goes to transit gateway ✓
  
  Packet to 8.8.8.8:
    Matches only 0.0.0.0/0 → goes to internet gateway ✓
  
  Packet to 10.0.2.50:
    Matches 10.0.0.0/16 → local (within VPC) ✓
```

### The "local" route

```
Every route table automatically has a "local" route:

  10.0.0.0/16     local

This route:
  - Cannot be removed or modified
  - Handles all intra-VPC traffic
  - Means any instance in any subnet can reach any other instance
    in the VPC (subject to security groups)
  
This is why "subnets in the same VPC can always talk to each other."
The local route ensures it.

To PREVENT two subnets from communicating:
  You can't remove the local route.
  Instead, use NACLs or Security Groups to deny traffic.
```

---

## Internet Gateway (IGW)

### How the IGW works

```
The Internet Gateway is the door between your VPC and the internet.

Properties:
  - Horizontally scaled, redundant, and highly available
  - AWS manages it — no bandwidth limits, no failover to configure
  - No single point of failure
  - You attach ONE IGW per VPC

What happens when an instance sends traffic to the internet:

  1. Instance (10.0.1.50) sends packet to 8.8.8.8
  2. Route table: 0.0.0.0/0 → igw
  3. IGW performs 1:1 NAT:
     Src IP: 10.0.1.50 → 54.23.45.67 (public/Elastic IP)
  4. Packet goes to internet with public IP as source
  
  For return traffic:
  5. Response from 8.8.8.8 arrives at IGW
  6. IGW reverses NAT: Dst 54.23.45.67 → 10.0.1.50
  7. Routes to the correct subnet and instance

  IMPORTANT: IGW does 1:1 NAT (unlike home router's many:1 NAT)
  Each instance needs its OWN public IP to use the IGW directly.

  Instance without public IP + IGW route?
  → Outbound traffic CAN'T go through IGW
  → Need NAT Gateway instead
```

---

## NAT Gateway

### How NAT Gateway works

```
NAT Gateway allows private instances to reach the internet 
for outbound traffic, without being reachable from the internet.

Architecture:
  ┌────────────────────────────────────────────────┐
  │ VPC                                             │
  │                                                 │
  │ ┌─────────────┐        ┌──────────────────┐    │
  │ │ Private      │        │ Public           │    │
  │ │ Subnet       │        │ Subnet           │    │
  │ │              │        │                  │    │
  │ │ ┌──────────┐ │        │ ┌──────────────┐ │    │
  │ │ │ App      │ │───────>│ │ NAT Gateway  │ │────>│── IGW ── Internet
  │ │ │ Instance │ │        │ │ (Elastic IP) │ │    │
  │ │ │ 10.0.2.x │ │        │ │ 10.0.1.x     │ │    │
  │ │ └──────────┘ │        │ └──────────────┘ │    │
  │ │              │        │                  │    │
  │ │ Route:       │        │ Route:           │    │
  │ │ 0.0.0.0/0   │        │ 0.0.0.0/0       │    │
  │ │ → NAT-GW    │        │ → IGW            │    │
  │ └─────────────┘        └──────────────────┘    │
  └────────────────────────────────────────────────┘

Flow:
  1. App instance (10.0.2.50) sends to pypi.org (151.101.0.223)
  2. Private route table: 0.0.0.0/0 → nat-gw
  3. NAT Gateway:
     - Rewrites Src IP: 10.0.2.50 → NAT GW's Elastic IP
     - Tracks connection in state table
  4. Packet goes through IGW to internet
  5. Response comes back to NAT GW's Elastic IP
  6. NAT GW reverses: Dst → 10.0.2.50
  7. Response delivered to app instance

Key facts:
  - NAT GW lives in a PUBLIC subnet (needs IGW access)
  - NAT GW has an Elastic IP (static public IP)
  - Many:1 NAT (like home router)
  - Supports up to 55,000 simultaneous connections per destination
  - ~$32/month + $0.045/GB processed (it's expensive!)
  - For HA: deploy NAT GW in each AZ
```

### NAT Gateway gotchas

```
1. Cost: $32/month + data processing charges
   This adds up fast with large data transfers.
   Consider VPC endpoints for AWS service access instead.

2. Connection limits: 55,000 per destination IP
   If you're making many connections to the same IP:
   "ErrorPortAllocation" in CloudWatch
   Solution: Use multiple NAT Gateways

3. Single AZ: NAT GW is in one AZ
   If that AZ goes down, all private subnets lose internet.
   Solution: NAT GW per AZ + per-AZ route tables

   ┌────────────┐    ┌────────────┐
   │ AZ-a       │    │ AZ-b       │
   │            │    │            │
   │ Private-a  │    │ Private-b  │
   │ ↓          │    │ ↓          │
   │ NAT-GW-a   │    │ NAT-GW-b   │
   │ ↓          │    │ ↓          │
   │ IGW        │    │ IGW        │
   └────────────┘    └────────────┘
   
   Each AZ has its own NAT GW and route table.
   AZ-a failure doesn't affect AZ-b traffic.

4. Timeout: idle TCP timeout is 350 seconds
   Long-idle connections get dropped.
   Your app may see mysterious connection resets.
```

---

## Elastic IPs and Public IPs

```
Two types of public IPs:

Auto-assigned public IP:
  - Assigned when instance launches (if subnet setting enables it)
  - Changes if instance is stopped and started
  - Released when instance is terminated
  - Free while instance is running

Elastic IP (EIP):
  - Static public IP you allocate
  - Persists until you release it
  - Can move between instances
  - $0.005/hour when NOT attached (waste penalty)
  - Used for: NAT Gateways, instances that need stable IPs

The IGW handles the mapping:
  ┌────────────────────────────────────────────┐
  │ IGW NAT Table                               │
  │                                             │
  │ Private IP       Public/Elastic IP          │
  │ ──────────       ─────────────────          │
  │ 10.0.1.50        54.23.45.67 (auto)         │
  │ 10.0.1.51        3.221.10.100 (EIP)         │
  │ 10.0.1.52        (none) → can't use IGW     │
  └────────────────────────────────────────────┘

The instance NEVER knows its public IP.
  Inside the VM: ip addr → only shows 10.0.1.50
  The IGW translates transparently.
  
  This is different from a "real" public IP on a server.
  It's always 1:1 NAT at the IGW level.
```

---

## VPC Endpoints — Private Access to Cloud Services

### The problem

```
Your private instance needs to access S3, DynamoDB, or SQS.
Without a VPC endpoint:

  Private Instance → NAT GW → IGW → Internet → S3

  Problems:
    1. Traffic goes through the public internet
    2. NAT GW costs money for data processing
    3. Higher latency
    4. Security concern (data crosses public network)
```

### Gateway Endpoints (S3 and DynamoDB)

```
Free. Uses route table entries.

  ┌──────────────────────────────────────────────┐
  │ VPC                                           │
  │                                               │
  │ ┌─────────┐      ┌──────────────────┐        │
  │ │ Private  │──────│ Gateway Endpoint │──── S3 │
  │ │ Instance │      │ (route table     │        │
  │ │          │      │  prefix list)    │        │
  │ └─────────┘      └──────────────────┘        │
  └──────────────────────────────────────────────┘

Route table gets:
  pl-63a5400a (S3 prefix list)  →  vpce-abc123

Traffic to S3 stays within the AWS network.
Free. No NAT GW charges. Lower latency.
```

### Interface Endpoints (everything else)

```
Creates an ENI (Elastic Network Interface) in your subnet.

  ┌──────────────────────────────────────────────┐
  │ VPC                                           │
  │                                               │
  │ ┌─────────┐      ┌──────────────────┐        │
  │ │ Private  │──────│ Interface        │──── SQS│
  │ │ Instance │      │ Endpoint         │        │
  │ │          │      │ (ENI with        │        │
  │ │          │      │  private IP)     │        │
  │ └─────────┘      └──────────────────┘        │
  └──────────────────────────────────────────────┘

  Uses PrivateLink technology.
  Creates: vpce-xyz.sqs.us-east-1.vpce.amazonaws.com
  Private DNS can map: sqs.us-east-1.amazonaws.com → private IP
  
  Cost: ~$7.20/month per AZ + $0.01/GB
  Still cheaper than NAT GW for high-traffic services.
```

---

## DNS Inside a VPC

### VPC DNS resolver

```
Every VPC has a built-in DNS resolver at:
  VPC CIDR base + 2
  
  Example: VPC 10.0.0.0/16 → DNS at 10.0.2 (10.0.0.2)

This resolver:
  - Resolves public DNS names normally
  - Resolves private hosted zone names
  - Resolves VPC endpoint DNS names
  - Resolves instance private DNS names

Instance DNS names (if enabled):
  ip-10-0-1-50.ec2.internal  →  10.0.1.50

Private Hosted Zones:
  You can create: internal.mycompany.com
  Only resolvable WITHIN the VPC.
  
  api.internal.mycompany.com → 10.0.2.30
  db.internal.mycompany.com  → 10.0.20.10
  
  Not resolvable from the public internet.
```

### DNS resolution flow

```
Instance queries: api.example.com

  1. Instance → VPC DNS (10.0.0.2)
  2. VPC DNS checks:
     a. Private Hosted Zone for this VPC? → No
     b. VPC Endpoint private DNS? → No
     c. Forward to public DNS → Recursive resolution
  3. Returns: 52.20.30.40

Instance queries: db.internal.mycompany.com

  1. Instance → VPC DNS (10.0.0.2)
  2. VPC DNS checks:
     a. Private Hosted Zone for this VPC? → YES
     b. Returns: 10.0.20.10
     (never goes to public DNS)
```

---

## VPN and Direct Connect

### Site-to-Site VPN

```
Connect your on-premises network to your VPC over the internet.

  ┌──────────┐     IPsec tunnel     ┌────────────────┐
  │ On-Prem   │◄═══════════════════►│ VGW / TGW       │
  │ Router    │     (encrypted)      │ (Virtual/Transit│
  │ (CGW)     │                      │  Gateway)       │
  └──────────┘                      └────────────────┘
  
  Two IPsec tunnels for redundancy.
  Each tunnel: ~1.25 Gbps max.
  Latency: depends on internet path (variable)
  Cost: ~$35/month per VPN connection

Route propagation:
  VGW can propagate BGP routes to VPC route tables.
  Your on-prem routes appear automatically:
    192.168.0.0/24    vgw-abc123 (propagated)
```

### Direct Connect

```
Dedicated physical connection to AWS.

  ┌──────────┐    Fiber    ┌────────┐    ┌──────┐
  │ On-Prem   │────────────│ DX     │────│ AWS  │
  │ DC        │            │ Location│   │ VPC  │
  └──────────┘            └────────┘    └──────┘
  
  Bandwidth: 1 Gbps, 10 Gbps, or 100 Gbps
  Latency: consistent (private link, not internet)
  Cost: port fee + data transfer (expensive but predictable)
  
  Use when:
    - Need consistent low latency
    - Transferring large amounts of data
    - Compliance requires private connection
    - ISP internet quality is unreliable
```

---

## Complete Traffic Flow Examples

### Example 1: Internet to public instance

```
User (internet) → EC2 instance in public subnet

  1. User sends request to 54.23.45.67 (public IP)
  2. Traffic arrives at AWS edge + Internet Gateway
  3. IGW: NAT 54.23.45.67 → 10.0.1.50 (private IP)
  4. VPC router: 10.0.1.0/24 → local → deliver to subnet
  5. Security Group evaluated:
     Inbound rule: allow TCP 443 from 0.0.0.0/0? → YES
  6. Packet delivered to instance

  Key: Instance in public subnet + has public IP + 
       IGW route + Security Group allows traffic
```

### Example 2: Private instance to internet

```
Private EC2 → apt-get update (needs internet)

  1. Instance (10.0.2.50) sends to archive.ubuntu.com (91.189.88.152)
  2. Route table: 0.0.0.0/0 → nat-gw
  3. NAT GW (10.0.1.200): 
     Src 10.0.2.50:43210 → 3.220.50.100:61234 (EIP + port)
  4. NAT GW's route: 0.0.0.0/0 → igw
  5. IGW: Already public IP (EIP), forward to internet
  6. Response returns to EIP → NAT GW → reverse NAT → instance

  Key: Private instance → NAT GW (in public subnet) → IGW → internet
```

### Example 3: Instance to S3 via endpoint

```
Private EC2 → S3 (via Gateway Endpoint)

  1. Instance (10.0.2.50) sends to s3.amazonaws.com
  2. DNS resolves to S3's IP (e.g., 52.217.x.x)
  3. Route table:
     52.217.0.0/16 matches S3 prefix list → vpce-abc (gateway endpoint)
  4. Traffic goes directly to S3 via AWS internal network
  5. No NAT GW, no IGW, no internet

  Key: FREE, lower latency, stays in AWS network
```

### Example 4: Cross-VPC via peering

```
Instance in VPC-A → Instance in VPC-B

  VPC-A: 10.0.0.0/16     VPC-B: 10.1.0.0/16

  1. Instance (10.0.1.50) sends to 10.1.2.30
  2. VPC-A route table: 10.1.0.0/16 → pcx-abc (peering)
  3. Traffic crosses peering connection
  4. VPC-B route table: 10.0.0.0/16 → pcx-abc (reverse route)
  5. VPC-B Security Group evaluated for destination instance
  6. Packet delivered to 10.1.2.30

  Both routing tables MUST have entries.
  Both Security Groups MUST allow the traffic.
  CIDRs MUST NOT overlap.
```

---

## Common Misconfigurations

```
1. "Instance can't reach the internet"
   Checklist:
   □ Is instance in a subnet with route to IGW (public) or NAT GW (private)?
   □ Does instance have a public IP (if public subnet)?
   □ Does Security Group allow outbound traffic?
   □ Does NACL allow outbound AND inbound (return traffic)?
   □ Is the IGW attached to the VPC?

2. "Can't reach instance from the internet"
   Checklist:
   □ Instance in public subnet?
   □ Instance has public/Elastic IP?
   □ Route table has IGW route?
   □ Security Group allows inbound on the port?
   □ NACL allows inbound on the port?
   □ OS firewall (iptables) allows the port?

3. "VPC peering doesn't work"
   Checklist:
   □ Peering connection accepted by both sides?
   □ Route table in VPC-A points to peering for VPC-B's CIDR?
   □ Route table in VPC-B points to peering for VPC-A's CIDR?
   □ Security Groups reference the peering VPC's CIDR?
   □ CIDRs don't overlap?
   □ DNS resolution enabled on peering connection?

4. "Private subnets can't reach AWS services"
   □ Option A: NAT Gateway (costs money)
   □ Option B: VPC Endpoint (preferred for S3, DynamoDB)
   □ Option C: Interface Endpoint (for other services)

5. "NAT Gateway costs too much"
   Common cause: Instances pulling from S3 through NAT GW.
   Fix: Add S3 Gateway Endpoint (free).
   This alone can save hundreds of dollars per month.
```

---

## Key Takeaways

1. **Public vs private subnet is ONLY about the route table** — public subnets route 0.0.0.0/0 to an Internet Gateway. Private subnets route to a NAT Gateway (or nowhere). There's no magic "public" checkbox
2. **The most specific route wins (longest prefix match)** — a /24 route beats a /16 route for addresses in the /24 range. The 0.0.0.0/0 default route is always the last resort
3. **The Internet Gateway performs 1:1 NAT** — it translates between your instance's private IP and its public/Elastic IP. This is different from NAT Gateway's many:1 NAT
4. **NAT Gateway is expensive and a single point of failure per AZ** — deploy one per AZ for high availability. Use VPC endpoints to avoid sending AWS service traffic through NAT
5. **VPC Gateway Endpoints for S3/DynamoDB are free** — they keep traffic on the AWS backbone and avoid NAT Gateway charges. There's almost no reason NOT to use them
6. **Every VPC gets a DNS resolver at CIDR+2** — it handles public DNS, private hosted zones, and VPC endpoint DNS. Private hosted zones let you create internal-only domain names
7. **The "local" route is permanent and handles all intra-VPC traffic** — you can't remove it. To restrict subnet-to-subnet traffic, use NACLs or Security Groups
8. **VPN connections give you encrypted tunnels over the internet; Direct Connect gives you a dedicated physical line** — VPN is cheaper but has variable latency. Direct Connect is expensive but consistent
9. **Most connectivity issues come from missing route table entries or Security Group rules** — always check both. Route tables control WHERE traffic goes; Security Groups control WHETHER it's allowed
10. **Design subnets large enough for growth** — Kubernetes with VPC CNI can consume IPs rapidly. A /24 (251 IPs) fills fast when every pod needs an IP. Use /20 or larger for K8s workloads

---

## Next

→ [Security Groups, NACLs, and Load Balancers](./03-security-groups-nacls-load-balancers.md) — Controlling and distributing traffic in the cloud
