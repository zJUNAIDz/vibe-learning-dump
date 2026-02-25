# CIDR & Subnetting — Dividing Networks with Precision

> Subnetting is the single most tested skill in networking interviews and certifications. More importantly, it's what you do every time you design a VPC, size a Kubernetes pod network, or troubleshoot "why can't server A reach server B." This chapter makes you fluent.

---

## Table of Contents

1. [Why Subnet?](#why-subnet)
2. [CIDR: Classless Inter-Domain Routing](#cidr-classless-inter-domain-routing)
3. [Subnetting Step by Step](#subnetting-step-by-step)
4. [Variable-Length Subnet Masks (VLSM)](#variable-length-subnet-masks-vlsm)
5. [Supernetting and Route Aggregation](#supernetting-and-route-aggregation)
6. [Subnetting Practice Problems](#subnetting-practice-problems)
7. [Real-World Subnet Design](#real-world-subnet-design)
8. [Linux: Subnet Operations](#linux-subnet-operations)

---

## Why Subnet?

### Problem 1: Broadcast containment

Without subnetting, all devices on a single network form one broadcast domain. An ARP request from any device reaches ALL devices. With 1,000 devices, every host processes 1,000× more broadcasts than necessary.

Subnetting breaks a large network into smaller ones, each with its own broadcast domain.

### Problem 2: Address allocation efficiency

An organization with 500 hosts used to get a Class B (65,534 addresses) because Class C (254) was too small. Subnetting lets them use exactly the right-sized blocks.

### Problem 3: Security and isolation

Different departments (engineering, finance, HR) should be on different subnets. Traffic between subnets passes through a router/firewall where access control can be applied.

### Problem 4: Routing efficiency

Routers work with network prefixes, not individual hosts. More specifically-sized subnets enable more efficient routing tables.

---

## CIDR: Classless Inter-Domain Routing

CIDR (RFC 4632) allows arbitrary prefix lengths — not just /8, /16, /24.

### CIDR notation

```
10.0.0.0/24    → Network: 10.0.0.0,     Hosts: 10.0.0.1 – 10.0.0.254
10.0.0.0/25    → Network: 10.0.0.0,     Hosts: 10.0.0.1 – 10.0.0.126
10.0.0.128/25  → Network: 10.0.0.128,   Hosts: 10.0.0.129 – 10.0.0.254
```

### How CIDR works

The number after the slash tells routers how many bits are the "network" prefix:
- **/24** = first 24 bits are network → 8 bits for hosts → $2^8 - 2 = 254$ usable
- **/25** = first 25 bits are network → 7 bits for hosts → $2^7 - 2 = 126$ usable
- **/20** = first 20 bits are network → 12 bits for hosts → $2^{12} - 2 = 4094$ usable

The formula:

$$\text{Usable hosts} = 2^{(32 - \text{prefix length})} - 2$$

The $-2$ accounts for the network address and broadcast address.

### CIDR reference table

| CIDR | Subnet Mask | Total IPs | Usable Hosts |
|------|-------------|-----------|-------------|
| /32 | 255.255.255.255 | 1 | 1 (single host) |
| /31 | 255.255.255.254 | 2 | 2 (point-to-point) |
| /30 | 255.255.255.252 | 4 | 2 |
| /29 | 255.255.255.248 | 8 | 6 |
| /28 | 255.255.255.240 | 16 | 14 |
| /27 | 255.255.255.224 | 32 | 30 |
| /26 | 255.255.255.192 | 64 | 62 |
| /25 | 255.255.255.128 | 128 | 126 |
| /24 | 255.255.255.0 | 256 | 254 |
| /23 | 255.255.254.0 | 512 | 510 |
| /22 | 255.255.252.0 | 1,024 | 1,022 |
| /21 | 255.255.248.0 | 2,048 | 2,046 |
| /20 | 255.255.240.0 | 4,096 | 4,094 |
| /19 | 255.255.224.0 | 8,192 | 8,190 |
| /18 | 255.255.192.0 | 16,384 | 16,382 |
| /17 | 255.255.128.0 | 32,768 | 32,766 |
| /16 | 255.255.0.0 | 65,536 | 65,534 |

---

## Subnetting Step by Step

### The method: splitting a network

Given: 192.168.1.0/24. Split into 4 equal subnets.

**Step 1**: How many subnets? 4. How many bits needed? $2^n \geq 4$, so $n = 2$

**Step 2**: Borrow 2 bits from the host portion:
- Original: /24 (24 network bits, 8 host bits)
- New: /26 (24 + 2 = 26 network bits, 6 host bits)

**Step 3**: Calculate each subnet:

```
Original: 192.168.1.00|000000   (/24, 8 host bits)
                      ↑
              Borrow 2 bits here → /26

Subnet 0:  192.168.1.00|000000 = 192.168.1.0/26     (hosts: .1 – .62,    broadcast: .63)
Subnet 1:  192.168.1.01|000000 = 192.168.1.64/26    (hosts: .65 – .126,  broadcast: .127)
Subnet 2:  192.168.1.10|000000 = 192.168.1.128/26   (hosts: .129 – .190, broadcast: .191)
Subnet 3:  192.168.1.11|000000 = 192.168.1.192/26   (hosts: .193 – .254, broadcast: .255)
```

**Step 4**: Verify — each subnet has $2^6 - 2 = 62$ usable hosts. Total: 4 × 62 = 248 usable (vs 254 in the original /24 — you "lose" 6 addresses to the extra network and broadcast addresses).

### Block size trick

A faster way to figure out subnet boundaries: the **block size** is $2^{\text{host bits}}$.

For /26: host bits = 6, block size = $2^6 = 64$

Subnets start at multiples of 64: 0, 64, 128, 192.

For /27: host bits = 5, block size = $2^5 = 32$

Subnets start at: 0, 32, 64, 96, 128, 160, 192, 224.

### Given an IP, find its subnet

**Problem**: What subnet does 10.50.100.200/21 belong to?

**Step 1**: /21 means the mask is 255.255.248.0. The "interesting" octet is the third one (248).

**Step 2**: Block size in the third octet: 256 - 248 = 8. Subnets start at multiples of 8 in the third octet.

**Step 3**: 100 ÷ 8 = 12.5 → floor = 12 → 12 × 8 = **96**

**Answer**: 10.50.96.0/21 (range: 10.50.96.0 – 10.50.103.255)

**Verify**: 
```bash
ipcalc 10.50.100.200/21
# Network:   10.50.96.0/21
# Broadcast: 10.50.103.255
# HostMin:   10.50.96.1
# HostMax:   10.50.103.254
```

---

## Variable-Length Subnet Masks (VLSM)

### The problem with equal-sized subnets

Real networks have departments of different sizes:
- Server subnet: 10 hosts
- Engineering: 100 hosts
- Management: 20 hosts
- Point-to-point links: 2 hosts each

Using /26 for everything wastes addresses in small subnets and may not be enough for large ones.

### VLSM solution

VLSM lets you use **different prefix lengths** for different subnets within the same address block.

**Example**: Given 192.168.1.0/24, create subnets for:
- Engineering: 100 hosts
- Management: 25 hosts
- Servers: 10 hosts
- Link A: 2 hosts
- Link B: 2 hosts

**Strategy**: Allocate the largest subnet first, then progressively smaller.

**Engineering (100 hosts)**: Need $2^n - 2 \geq 100$ → $n = 7$ (128 - 2 = 126 hosts) → **/25**
- 192.168.1.0/25 (hosts: .1 – .126, broadcast: .127)

**Management (25 hosts)**: Need $2^n - 2 \geq 25$ → $n = 5$ (32 - 2 = 30 hosts) → **/27**
- 192.168.1.128/27 (hosts: .129 – .158, broadcast: .159)

**Servers (10 hosts)**: Need $2^n - 2 \geq 10$ → $n = 4$ (16 - 2 = 14 hosts) → **/28**
- 192.168.1.160/28 (hosts: .161 – .174, broadcast: .175)

**Link A (2 hosts)**: /30
- 192.168.1.176/30 (hosts: .177, .178, broadcast: .179)

**Link B (2 hosts)**: /30
- 192.168.1.180/30 (hosts: .181, .182, broadcast: .183)

**Remaining space**: 192.168.1.184/24 through 192.168.1.255 — available for future use.

---

## Supernetting and Route Aggregation

### Combining routes: the opposite of subnetting

If a router has these routes:
```
10.0.0.0/24
10.0.1.0/24
10.0.2.0/24
10.0.3.0/24
```

It can **aggregate** them into a single route: **10.0.0.0/22**

Why? The first 22 bits of all four prefixes are identical:
```
10.0.0.x  = 00001010.00000000.000000|00.xxxxxxxx
10.0.1.x  = 00001010.00000000.000000|01.xxxxxxxx
10.0.2.x  = 00001010.00000000.000000|10.xxxxxxxx
10.0.3.x  = 00001010.00000000.000000|11.xxxxxxxx
                                      ↑
                               22 bits are the same
```

**Benefits**:
- Fewer routing table entries (4 → 1)
- Faster route lookups
- Smaller routing updates in protocols like BGP and OSPF

This is why ISPs and cloud providers allocate contiguous address blocks — to enable aggregation.

---

## Subnetting Practice Problems

### Problem 1
**Given**: 172.16.0.0/16. Create 8 equal subnets. What is the new prefix length? What are the subnets?

**Solution**: 
- 8 subnets → borrow 3 bits ($2^3 = 8$)
- New prefix: /16 + 3 = **/19**
- Block size in third octet: $2^{(24-19)} = 2^5 = 32$
- Subnets: 172.16.0.0/19, 172.16.32.0/19, 172.16.64.0/19, 172.16.96.0/19, 172.16.128.0/19, 172.16.160.0/19, 172.16.192.0/19, 172.16.224.0/19
- Hosts per subnet: $2^{13} - 2 = 8190$

### Problem 2
**Given**: IP address 10.128.200.77/20. What network is this in? What's the broadcast address?

**Solution**:
- /20 → mask 255.255.240.0
- Block size in third octet: 256 - 240 = 16
- 200 ÷ 16 = 12.5 → floor = 12, 12 × 16 = 192
- **Network**: 10.128.192.0/20
- **Broadcast**: 10.128.207.255 (192 + 16 - 1 = 207)
- Host range: 10.128.192.1 – 10.128.207.254

### Problem 3
**Given**: You need exactly 500 hosts. What is the smallest subnet that fits?

**Solution**:
- $2^n - 2 \geq 500$ → $2^9 - 2 = 510$ → n = 9 host bits
- Prefix: $32 - 9 =$ **/23**
- A /23 gives 510 usable hosts

```bash
# Verify all problems with ipcalc
ipcalc 172.16.0.0/19
ipcalc 10.128.200.77/20
ipcalc 10.0.0.0/23
```

---

## Real-World Subnet Design

### AWS VPC subnet design

A typical AWS VPC design:

```
VPC: 10.0.0.0/16 (65,534 hosts — room to grow)

Public subnets (for load balancers, bastion hosts):
  10.0.0.0/24   (AZ-a, 254 hosts)
  10.0.1.0/24   (AZ-b, 254 hosts)
  10.0.2.0/24   (AZ-c, 254 hosts)

Private subnets (for application servers):
  10.0.10.0/24  (AZ-a)
  10.0.11.0/24  (AZ-b)
  10.0.12.0/24  (AZ-c)

Database subnets (isolated):
  10.0.20.0/24  (AZ-a)
  10.0.21.0/24  (AZ-b)

Room for growth:
  10.0.100.0/22 (future expansion — 1022 hosts)
```

**AWS-specific**: AWS reserves 5 IPs per subnet (first 4 + last 1), not just the usual 2. So a /24 gives you only 251 usable addresses in AWS.

### Kubernetes pod networking

Kubernetes needs IP addresses for every pod. A cluster with 100 nodes, each running 30 pods, needs 3,000 pod IPs:

```
Pod CIDR: 10.244.0.0/16  (65,534 addresses)
Each node gets a /24:     10.244.0.0/24, 10.244.1.0/24, ...
Pods on node 0:           10.244.0.1 – 10.244.0.254
Pods on node 1:           10.244.1.1 – 10.244.1.254
```

### Home network

```
Router LAN: 192.168.1.0/24
DHCP range: 192.168.1.100 – 192.168.1.254 (for phones, laptops)
Static IPs:  192.168.1.1 (router), .2 (NAS), .3 (printer)
```

---

## Linux: Subnet Operations

### Working with routes and subnets

```bash
# Show routing table (includes subnet information)
ip route show
# 192.168.1.0/24 dev eth0 proto kernel scope link src 192.168.1.100
# 10.0.0.0/8 via 192.168.1.1 dev eth0
# default via 192.168.1.1 dev eth0

# Add a route to a specific subnet
sudo ip route add 10.0.0.0/24 via 192.168.1.1 dev eth0

# Add a route to a directly connected subnet
sudo ip route add 172.16.0.0/24 dev eth1

# Delete a route
sudo ip route del 10.0.0.0/24

# Which route does a specific IP match?
ip route get 10.0.0.50
# Shows which routing table entry will be used
```

### Testing subnet reachability

```bash
# Ping all hosts in a /24 (broadcast ping — often blocked)
ping -b 192.168.1.255

# Scan a subnet with nmap
nmap -sn 192.168.1.0/24
# -sn: ping scan only (no port scan)
# Shows all responding hosts in the subnet

# ARP scan a subnet (faster, Layer 2)
sudo apt install arp-scan
sudo arp-scan --interface=eth0 192.168.1.0/24
```

### ipcalc and sipcalc

```bash
# Install subnet calculators
sudo apt install ipcalc sipcalc

# ipcalc: basic subnet information
ipcalc 192.168.1.100/26
# Address:   192.168.1.100
# Network:   192.168.1.64/26
# Netmask:   255.255.255.192
# Broadcast: 192.168.1.127
# HostMin:   192.168.1.65
# HostMax:   192.168.1.126
# Hosts/Net: 62

# sipcalc: more detailed, supports IPv6
sipcalc 10.0.0.0/20
```

---

## Key Takeaways

1. **CIDR replaces classful addressing** — use exactly the prefix length you need
2. **Subnetting = borrowing host bits** to create more networks
3. **Block size = $2^{\text{host bits}}$** — subnets start at multiples of the block size
4. **VLSM** allows different subnet sizes within the same allocation — essential for real networks
5. **Route aggregation (supernetting)** combines contiguous routes into one — reduces routing table size
6. **Always allocate largest subnets first** in VLSM design
7. **Real-world design**: Leave room for growth, separate by function (public/private/database), plan for multi-AZ

---

## Next

→ [03-ip-header-fragmentation.md](03-ip-header-fragmentation.md) — What's inside an IP packet, and what happens when it's too big
