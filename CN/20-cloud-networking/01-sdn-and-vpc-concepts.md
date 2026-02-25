# SDN and VPC Concepts

> Cloud networking exists because traditional networking doesn't scale. When you need to provision a network for 10,000 customers in the same data center, each isolated from the other, you can't plug in physical cables. You need software. That's SDN.

---

## Table of Contents

1. [Why Cloud Networking Exists](#why)
2. [Software-Defined Networking (SDN)](#sdn)
3. [The SDN Control Plane and Data Plane Split](#control-data-split)
4. [Virtual Private Cloud (VPC) — The Foundation](#vpc)
5. [VPC Architecture Deep Dive](#vpc-architecture)
6. [How Cloud Providers Implement VPCs](#vpc-implementation)
7. [Multi-VPC Architectures](#multi-vpc)
8. [VPC Peering](#vpc-peering)
9. [Transit Gateway](#transit-gateway)
10. [Comparing Cloud Provider Terminology](#cloud-comparison)
11. [Key Takeaways](#key-takeaways)

---

## Why Cloud Networking Exists

```
Traditional networking:
  Buy switch → Buy router → Cable them → Configure VLANs
  Time: days to weeks
  Scale: hundreds of hosts
  Isolation: VLANs (max 4094)
  Change: SSH into device, modify config

Cloud networking:
  API call → Network exists
  Time: seconds
  Scale: millions of hosts
  Isolation: per-customer virtual networks (millions)
  Change: API call, infrastructure-as-code

The problem cloud solves:
  Thousands of customers sharing the same physical infrastructure.
  Each customer needs:
    - Their own IP address space (even if it overlaps others)
    - Complete isolation from other customers
    - Flexible subnets, routing, firewalls
    - All without touching physical hardware

Traditional VLANs max out at 4094.
Cloud providers need millions of isolated networks.
→ VLANs don't work. We need something new.
→ Software-Defined Networking.
```

---

## Software-Defined Networking (SDN)

### The core idea

```
Traditional networking:
  ┌──────────────────┐
  │ Router / Switch  │  ← Control plane + Data plane
  │ (Cisco, Juniper) │     in the same box
  │                  │
  │ Brain + Muscle   │  ← Makes decisions AND forwards packets
  │ tightly coupled  │
  └──────────────────┘

SDN networking:
  ┌──────────────────┐
  │ SDN Controller   │  ← Control plane (centralized brain)
  │ (software)       │     Makes ALL routing/switching decisions
  └────────┬─────────┘
           │ OpenFlow / gRPC / custom protocol
  ┌────────┴─────────┐
  │ Dumb switches    │  ← Data plane (decentralized muscle)
  │ (hardware/       │     Just forward packets based on rules
  │  software)       │     from the controller
  └──────────────────┘

SDN = separate the brain from the body.
The controller programs the switches.
Switches just follow rules.
```

### Why this matters

```
Without SDN:
  Want to change a firewall rule?
  → SSH into 500 switches
  → Update each one's ACL
  → Hope you didn't make a typo on switch #347

With SDN:
  Want to change a firewall rule?
  → API call to controller
  → Controller pushes to all 500 switches atomically
  → Done in seconds

This is how cloud providers manage networks for millions of VMs:
  One controller, thousands of network nodes.
  Every network change is a software operation.
```

---

## The SDN Control Plane and Data Plane Split

```
┌─────────────────────────────────────────────────┐
│                SDN ARCHITECTURE                  │
│                                                  │
│  ┌─────────────┐                                │
│  │ Application  │ ← "I need a network with       │
│  │ Layer        │    3 subnets and a firewall"    │
│  │ (API/UI)     │                                │
│  └──────┬──────┘                                │
│         │ Northbound API (REST, gRPC)            │
│  ┌──────┴──────┐                                │
│  │ Control     │ ← Translates intent into        │
│  │ Plane       │    forwarding rules              │
│  │ (Controller)│    Maintains network state       │
│  │             │    Computes paths                 │
│  └──────┬──────┘                                │
│         │ Southbound API (OpenFlow, P4, custom)  │
│  ┌──────┴──────┐                                │
│  │ Data Plane  │ ← Forwards packets              │
│  │ (Switches,  │    Applies rules                 │
│  │  vSwitches) │    Encap/decap overlays          │
│  └─────────────┘                                │
└─────────────────────────────────────────────────┘

In cloud providers:
  Application layer = AWS Console / Terraform / API
  Control plane = AWS internal network controller
  Data plane = Custom hardware (AWS Nitro, Google Andromeda)
```

### Real-world SDN implementations

```
Open source:
  - Open vSwitch (OVS): Software switch, used in OpenStack
  - ONOS: SDN controller (Java-based)
  - OpenDaylight: SDN controller framework

Cloud provider:
  - AWS: Nitro network cards + custom controller
  - Google Cloud: Andromeda SDN stack
  - Azure: Azure SmartNIC (FPGA-based)

These are NOT academic exercises.
Every VM in AWS runs on top of an SDN stack.
When you create a VPC, the SDN controller programs
the data plane to isolate your traffic.
```

---

## Virtual Private Cloud (VPC) — The Foundation

### What a VPC actually is

```
A VPC is YOUR private network inside the cloud provider's data center.

You define:
  ┌──────────────────────────────────────────────┐
  │ VPC: 10.0.0.0/16                              │
  │ (65,536 IP addresses, all yours)              │
  │                                                │
  │ ┌──────────────┐  ┌──────────────┐            │
  │ │ Subnet A     │  │ Subnet B     │            │
  │ │ 10.0.1.0/24  │  │ 10.0.2.0/24  │            │
  │ │ (public)     │  │ (private)    │            │
  │ │ us-east-1a   │  │ us-east-1b   │            │
  │ │              │  │              │            │
  │ │ ┌──┐ ┌──┐   │  │ ┌──┐ ┌──┐   │            │
  │ │ │EC2│ │EC2│  │  │ │RDS│ │EC2│  │            │
  │ │ └──┘ └──┘   │  │ └──┘ └──┘   │            │
  │ └──────┬───────┘  └──────┬───────┘            │
  │        │                 │                     │
  │    Route Table       Route Table               │
  │    (public)          (private)                 │
  │        │                 │                     │
  │    Internet           NAT Gateway              │
  │    Gateway            (for outbound)           │
  └────────┬──────────────────────────────────────┘
           │
       Internet

Key properties:
  1. Your IP space: You choose the CIDR block
  2. Isolation: No traffic leaks to other VPCs (by default)
  3. Control: You define subnets, routing, firewalls
  4. It's YOUR network in the cloud
```

### VPC vs. traditional data center

```
Traditional DC                Cloud VPC
──────────────                ─────────
Physical router               Virtual router (implicit)
Physical switch               No visible switch
VLAN for isolation             VPC for isolation
Hardware firewall              Security Groups + NACLs
Buy more hardware to scale     API call to add capacity
Takes weeks to change          Takes seconds to change
You manage everything          Provider manages infra
```

---

## VPC Architecture Deep Dive

### CIDR planning

```
VPC CIDR: The IP range for your entire VPC.

Common choices:
  10.0.0.0/16    → 65,534 usable IPs  (most common)
  10.0.0.0/20    → 4,094 usable IPs   (smaller VPC)
  172.16.0.0/16  → 65,534 usable IPs  (avoids 10.x conflicts)

CRITICAL RULE: Plan for VPC peering.
  VPC-A: 10.0.0.0/16
  VPC-B: 10.0.0.0/16     ← CANNOT peer (overlap!)
  VPC-B: 10.1.0.0/16     ← CAN peer ✓

AWS reserves 5 IPs per subnet:
  10.0.1.0/24 has 256 IPs:
    10.0.1.0    → Network address
    10.0.1.1    → AWS VPC router
    10.0.1.2    → AWS DNS
    10.0.1.3    → Reserved for future
    10.0.1.255  → Broadcast
  Usable: 251 addresses (not 256 or 254)
```

### Subnet design

```
Two patterns:

Pattern 1: Public + Private (simple)
  ┌──────────────────────────────────────────┐
  │ VPC 10.0.0.0/16                           │
  │                                           │
  │ AZ-a              AZ-b                    │
  │ ┌──────────┐      ┌──────────┐            │
  │ │ Public   │      │ Public   │            │
  │ │ 10.0.1.0 │      │ 10.0.3.0 │            │
  │ │ /24      │      │ /24      │            │
  │ │ ALB, NAT │      │ ALB      │            │
  │ └──────────┘      └──────────┘            │
  │ ┌──────────┐      ┌──────────┐            │
  │ │ Private  │      │ Private  │            │
  │ │ 10.0.2.0 │      │ 10.0.4.0 │            │
  │ │ /24      │      │ /24      │            │
  │ │ App, DB  │      │ App, DB  │            │
  │ └──────────┘      └──────────┘            │
  └──────────────────────────────────────────┘

Pattern 2: Three-tier (enterprise)
  ┌──────────────────────────────────────────┐
  │ VPC 10.0.0.0/16                           │
  │                                           │
  │ ┌──────────┐  Public:  ALB, bastion       │
  │ │ 10.0.1.0 │                              │
  │ └──────────┘                              │
  │ ┌──────────┐  App:     ECS/EKS tasks      │
  │ │ 10.0.10.0│                              │
  │ └──────────┘                              │
  │ ┌──────────┐  Data:    RDS, ElastiCache   │
  │ │ 10.0.20.0│                              │
  │ └──────────┘                              │
  └──────────────────────────────────────────┘

Public subnet: Route table has route to Internet Gateway
Private subnet: No route to Internet Gateway (uses NAT for outbound)
```

---

## How Cloud Providers Implement VPCs

### Under the hood

```
You think:
  "I have a network 10.0.0.0/16 with routing and firewalls."

Reality:
  Your VMs run on physical hosts shared with other customers.
  There is NO physical router for your VPC.
  There are NO physical switches for your subnets.
  
  Everything is SOFTWARE:
  
  ┌───────────────────────────────────────────┐
  │ Physical Host                              │
  │                                            │
  │ ┌──────┐ ┌──────┐ ┌──────┐               │
  │ │ VM-A │ │ VM-B │ │ VM-C │               │
  │ │ cust1│ │ cust2│ │ cust1│               │
  │ └──┬───┘ └──┬───┘ └──┬───┘               │
  │    │        │        │                     │
  │ ┌──┴────────┴────────┴───┐                │
  │ │ Hypervisor / Nitro NIC  │ ← THIS is the │
  │ │ (SDN data plane)        │   virtual      │
  │ │                         │   switch +     │
  │ │ Enforces:               │   router +     │
  │ │ - VPC isolation         │   firewall     │
  │ │ - Security Groups       │                │
  │ │ - Routing               │                │
  │ │ - Encapsulation         │                │
  │ └─────────┬───────────────┘                │
  │           │                                │
  └───────────┼────────────────────────────────┘
              │ Physical network
              │ (uses encapsulation for isolation,
              │  similar to VXLAN but proprietary)
              │
  ┌───────────┼────────────────────────────────┐
  │ Physical Host (another)                    │
  ...

Key insights:
  1. VPC routing happens at the hypervisor/NIC level
  2. Security Groups are enforced BEFORE the VM sees the packet
  3. Cross-host traffic is encapsulated (like VXLAN)
  4. The "VPC router" is a distributed virtual router
     running on every host
```

### AWS Nitro

```
Traditional:
  Hypervisor (software) handles all networking
  → CPU overhead, security attack surface

AWS Nitro:
  Dedicated hardware card handles networking
  → No CPU overhead for networking
  → Hardware-enforced isolation
  → Network processing offloaded to custom ASIC

  ┌──────────┐      ┌──────────────┐
  │ EC2 VM   │──────│ Nitro Card   │──── Physical Network
  │ (your    │      │ (hardware)   │
  │  code)   │      │ VPC routing  │
  │          │      │ Security Grp │
  │          │      │ Encap/Decap  │
  └──────────┘      └──────────────┘

This is why AWS can guarantee network isolation:
  It's enforced in hardware you can't access.
```

---

## Multi-VPC Architectures

### Why multiple VPCs?

```
Single VPC is fine for small deployments.
But organizations need:

  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
  │ Production   │  │ Staging      │  │ Development  │
  │ VPC          │  │ VPC          │  │ VPC          │
  │ 10.0.0.0/16  │  │ 10.1.0.0/16  │  │ 10.2.0.0/16  │
  └──────────────┘  └──────────────┘  └──────────────┘
  
  Separate VPCs for:
    - Environment isolation (prod vs dev)
    - Team isolation (team A can't affect team B)
    - Compliance (PCI data in isolated VPC)
    - Blast radius (failure in one doesn't affect others)
    - Different regions/accounts

Now the question: How do these VPCs communicate?
```

---

## VPC Peering

```
VPC Peering: Direct connection between two VPCs.

  ┌──────────────┐     Peering     ┌──────────────┐
  │ VPC-A        │◄───Connection───►│ VPC-B        │
  │ 10.0.0.0/16  │                  │ 10.1.0.0/16  │
  └──────────────┘                  └──────────────┘

Requirements:
  - Non-overlapping CIDRs
  - Route table entries in both VPCs
  - Security groups must allow the traffic

Properties:
  ✓ Traffic stays on cloud provider backbone (never public internet)
  ✓ Low latency (like being on the same network)
  ✓ Can peer across regions (inter-region peering)
  ✓ Can peer across accounts
  
  ✗ NOT transitive: A↔B and B↔C does NOT mean A↔C
  ✗ Doesn't scale: N VPCs need N*(N-1)/2 peering connections

Problem with peering at scale:
  5 VPCs:  10 peering connections
  10 VPCs: 45 peering connections
  20 VPCs: 190 peering connections
  
  → Need Transit Gateway
```

---

## Transit Gateway

```
Transit Gateway: Hub-and-spoke for VPC connectivity.

  Without TGW (mesh):          With TGW (hub-spoke):
  
  VPC-A ──── VPC-B             VPC-A ─┐
    │  ╲  ╱    │                       │
    │   ╳     │               VPC-B ──┼── Transit
    │  ╱  ╲    │                       │   Gateway
  VPC-C ──── VPC-D             VPC-C ──┤
                               VPC-D ──┤
  6 connections                VPN ────┘
                               4 connections

Transit Gateway properties:
  ✓ Hub-and-spoke: Each VPC connects to TGW only
  ✓ Transitive: A→TGW→C works
  ✓ Scales to thousands of VPCs
  ✓ Supports VPN and Direct Connect attachments
  ✓ Route tables for controlling traffic flow
  ✓ Cross-region peering between TGWs

Route table control:
  TGW has its own route table.
  You can control which VPCs can reach which:
  
  ┌──────────────────────────────────────────┐
  │ TGW Route Table: Production              │
  │                                          │
  │ Destination      Target                  │
  │ 10.0.0.0/16      VPC-Prod attachment     │
  │ 10.3.0.0/16      VPC-Shared attachment   │
  │ 0.0.0.0/0        VPN attachment          │
  │                                          │
  │ NOTE: 10.1.0.0/16 (Staging) NOT listed   │
  │ → Prod cannot reach Staging ✓            │
  └──────────────────────────────────────────┘
```

---

## Comparing Cloud Provider Terminology

```
Concept              AWS                  GCP                  Azure
───────              ───                  ───                  ─────
Virtual network      VPC                  VPC                  VNet
Subdivision          Subnet               Subnet               Subnet
Internet access      Internet Gateway     Cloud Router         Internet Gateway
NAT for private      NAT Gateway          Cloud NAT            NAT Gateway
Firewall (instance)  Security Group       Firewall Rules       NSG
Firewall (subnet)    NACL                 (Firewall Rules)     NSG (on subnet)
VPC connect          VPC Peering          VPC Peering          VNet Peering
Hub-spoke            Transit Gateway      NCC                  Virtual WAN
Private DNS          Route 53 Private     Cloud DNS Private    Private DNS Zone
Load balancer        ALB/NLB              Cloud LB             Azure LB/AppGW
VPN                  Site-to-Site VPN     Cloud VPN            VPN Gateway
Dedicated line       Direct Connect       Cloud Interconnect   ExpressRoute

The concepts are identical.
The names are different.
Learn the concepts, adapt the names.
```

---

## Key Takeaways

1. **SDN separates the control plane from the data plane** — the controller (brain) programs the switches (muscle). This is how cloud providers manage millions of virtual networks with software
2. **A VPC is your private network in the cloud** — you choose the IP space, define subnets, control routing, and enforce firewall rules. It's logically isolated from every other customer
3. **There is no physical router for your VPC** — it's all software running on the hypervisor or smart NIC. The "VPC router" is a distributed virtual router running on every physical host
4. **Security Groups are enforced at the hypervisor/NIC level** — before a packet even reaches your VM. This is hardware-enforced isolation you cannot bypass from inside the VM
5. **Subnets map to Availability Zones** — a subnet exists in one AZ. For high availability, deploy across at least 2 AZs with subnets in each
6. **Public vs private subnet is just a routing difference** — public subnets have a route to the Internet Gateway. Private subnets don't (they use NAT Gateway for outbound)
7. **Plan CIDRs for the future** — VPC peering requires non-overlapping CIDRs. If you use 10.0.0.0/16 everywhere, you can never peer. Use a consistent scheme (10.0.0.0/16, 10.1.0.0/16, etc.)
8. **VPC Peering is NOT transitive** — A↔B and B↔C doesn't mean A↔C. For complex topologies, use Transit Gateway as a hub
9. **Transit Gateway enables hub-and-spoke with route control** — scales to thousands of VPCs, supports VPN/Direct Connect, and route tables control which VPCs can talk to which
10. **All cloud providers use the same concepts with different names** — learn VPC/subnet/route table/security group/NAT gateway conceptually, and you can work in any cloud

---

## Next

→ [Subnets, Routing, and Gateways](./02-subnets-routing-gateways.md) — How traffic flows inside and outside your VPC
