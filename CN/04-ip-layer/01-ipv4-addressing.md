# IPv4 Addressing — Every Bit Matters

> An IP address is not just a number. It's a carefully structured identifier that encodes both "which network" and "which host" — and understanding that structure is the foundation of everything in networking above Layer 2.

---

## Table of Contents

1. [Why We Need Layer 3 Addressing](#why-we-need-layer-3-addressing)
2. [IPv4 Address Structure](#ipv4-address-structure)
3. [Binary Representation — You Must Be Fluent](#binary-representation--you-must-be-fluent)
4. [Network and Host Portions](#network-and-host-portions)
5. [Subnet Masks Explained](#subnet-masks-explained)
6. [Classful Addressing (Historical but Essential)](#classful-addressing)
7. [Special Addresses](#special-addresses)
8. [Private vs Public Addresses (RFC 1918)](#private-vs-public-addresses-rfc-1918)
9. [How a Host Decides: Local or Remote?](#how-a-host-decides-local-or-remote)
10. [IPv4 Address Exhaustion](#ipv4-address-exhaustion)
11. [IPv6 Overview](#ipv6-overview)
12. [Linux: Working with IP Addresses](#linux-working-with-ip-addresses)

---

## Why We Need Layer 3 Addressing

MAC addresses work within a single network segment (broadcast domain). But the internet connects millions of segments. MAC addresses are flat — they have no structure that tells you where a device is located in the network topology.

IP addresses are **hierarchical**. They have two parts:
1. **Network portion**: Identifies which network the host is on
2. **Host portion**: Identifies which specific host within that network

This structure enables **routing** — the ability to forward packets across networks without knowing every individual host, just the networks.

**Analogy**: A phone number like +1-212-555-1234 has country code (1), area code (212), and subscriber number (555-1234). You don't need to know every phone in Manhattan to route a call there — you just need to know that 212 goes to Manhattan. IP works the same way.

---

## IPv4 Address Structure

An IPv4 address is a **32-bit number**, written as four decimal octets separated by dots:

```
Dotted decimal:    192.168.1.100
Binary:            11000000.10101000.00000001.01100100
Hex:               C0.A8.01.64
```

Each octet is 8 bits, so each ranges from 0 to 255.

$32 \text{ bits} = 2^{32} = 4,294,967,296 \text{ possible addresses}$

That's ~4.3 billion addresses — which seemed infinite in 1981 when IPv4 was designed, but we ran out decades ago.

---

## Binary Representation — You Must Be Fluent

Networking professionals MUST be comfortable with binary. Subnet masks, CIDR notation, and routing all work at the bit level.

### Converting decimal to binary

For each octet:
```
128  64  32  16   8   4   2   1
 2⁷  2⁶  2⁵  2⁴  2³  2²  2¹  2⁰
```

Example: Convert 192 to binary:
- 192 ≥ 128? Yes → 1, remainder = 192 - 128 = 64
- 64 ≥ 64? Yes → 1, remainder = 64 - 64 = 0
- 0 ≥ 32? No → 0
- 0 ≥ 16? No → 0
- 0 ≥ 8? No → 0
- 0 ≥ 4? No → 0
- 0 ≥ 2? No → 0
- 0 ≥ 1? No → 0

Result: 192 = **11000000**

### Common values you should memorize

| Decimal | Binary | Notes |
|---------|--------|-------|
| 0 | 00000000 | All zeros |
| 128 | 10000000 | Highest bit only |
| 192 | 11000000 | Top 2 bits |
| 224 | 11100000 | Top 3 bits |
| 240 | 11110000 | Top 4 bits |
| 248 | 11111000 | Top 5 bits |
| 252 | 11111100 | Top 6 bits |
| 254 | 11111110 | Top 7 bits |
| 255 | 11111111 | All ones |

These are the only values that appear in subnet masks. Memorize them.

---

## Network and Host Portions

Every IPv4 address is split into two parts:

```
IP Address:  192.168.1.100
             |← Network →|← Host →|
             
The exact split depends on the subnet mask.
```

### The subnet mask defines the split

A subnet mask is also a 32-bit number. The **1 bits** mark the network portion. The **0 bits** mark the host portion.

```
IP:          11000000.10101000.00000001.01100100  (192.168.1.100)
Mask:        11111111.11111111.11111111.00000000  (255.255.255.0)
             |←──── Network (24 bits) ─────→|←H→|  Host (8 bits)
```

**AND operation** to get the network address:
```
IP:          11000000.10101000.00000001.01100100
Mask:        11111111.11111111.11111111.00000000
Network:     11000000.10101000.00000001.00000000  = 192.168.1.0
```

This "bitwise AND" is exactly what your computer does to determine the network address. It's the most fundamental operation in IP networking.

### What each portion tells you

| Portion | Purpose |
|---------|---------|
| **Network** | "Which street you live on" — routers use this to forward packets toward the right network |
| **Host** | "Which house on that street" — used within the local network to identify the specific device |

### Two reserved addresses per network

In every subnet, two addresses are reserved:
1. **Network address**: All host bits = 0 (e.g., 192.168.1.0/24) — identifies the network itself
2. **Broadcast address**: All host bits = 1 (e.g., 192.168.1.255/24) — sends to all hosts on the network

So a /24 network has $2^8 - 2 = 254$ usable host addresses (1-254).

In general: a network with $h$ host bits has $2^h - 2$ usable addresses.

---

## Subnet Masks Explained

### Dotted decimal notation

```
255.255.255.0    = 11111111.11111111.11111111.00000000 → /24
255.255.0.0      = 11111111.11111111.00000000.00000000 → /16
255.0.0.0        = 11111111.00000000.00000000.00000000 → /8
255.255.255.128  = 11111111.11111111.11111111.10000000 → /25
255.255.255.192  = 11111111.11111111.11111111.11000000 → /26
```

### CIDR notation

CIDR (Classless Inter-Domain Routing) notation is simpler: just write the number of network bits after a slash.

192.168.1.0 **/24** means "the first 24 bits are the network portion."

### All valid subnet masks

A valid subnet mask is a sequence of 1s followed by a sequence of 0s. Never mixed. This means only certain values are valid per octet:

| CIDR | Mask (last varying octet) | # Hosts | # Usable |
|------|---------------------------|---------|----------|
| /24 | 255.255.255.0 | 256 | 254 |
| /25 | 255.255.255.128 | 128 | 126 |
| /26 | 255.255.255.192 | 64 | 62 |
| /27 | 255.255.255.224 | 32 | 30 |
| /28 | 255.255.255.240 | 16 | 14 |
| /29 | 255.255.255.248 | 8 | 6 |
| /30 | 255.255.255.252 | 4 | 2 |
| /31 | 255.255.255.254 | 2 | 2 (special, point-to-point) |
| /32 | 255.255.255.255 | 1 | 1 (single host) |

**Memorize /24 through /30**. They appear constantly in real networking.

---

## Classful Addressing

Before CIDR (pre-1993), IPv4 addresses were divided into **classes**:

| Class | First bits | Range | Default mask | Networks | Hosts/network |
|-------|-----------|-------|-------------|----------|---------------|
| A | 0xxx | 0.0.0.0 – 127.255.255.255 | /8 | 128 | 16,777,214 |
| B | 10xx | 128.0.0.0 – 191.255.255.255 | /16 | 16,384 | 65,534 |
| C | 110x | 192.0.0.0 – 223.255.255.255 | /24 | 2,097,152 | 254 |
| D | 1110 | 224.0.0.0 – 239.255.255.255 | (multicast) | — | — |
| E | 1111 | 240.0.0.0 – 255.255.255.255 | (reserved) | — | — |

### Why classful addressing failed

The problem was the enormous gap between class sizes:
- A Class C gave you 254 hosts — too few for many organizations
- A Class B gave you 65,534 hosts — too many for most organizations (wasted addresses)
- A Class A gave you 16.7 million hosts — only the largest organizations needed this

Companies that needed 1,000 hosts were given a Class B (65,534 addresses), wasting 64,534 addresses. This waste accelerated IPv4 exhaustion.

**CIDR (1993)** eliminated classes. Now you can have a /22 (1,022 hosts) or a /27 (30 hosts) — any size that fits your needs.

**But**: The class A/B/C terminology persists in informal usage. "Private Class A" (10.0.0.0/8) is technically incorrect CIDR-wise but universally understood.

---

## Special Addresses

| Address(es) | Purpose |
|------------|---------|
| 0.0.0.0/8 | "This network" — used in routing (default route: 0.0.0.0/0) |
| 10.0.0.0/8 | Private (RFC 1918) |
| 100.64.0.0/10 | Carrier-Grade NAT (RFC 6598) |
| 127.0.0.0/8 | Loopback — 127.0.0.1 is "localhost" |
| 169.254.0.0/16 | Link-local — auto-assigned when DHCP fails |
| 172.16.0.0/12 | Private (RFC 1918) |
| 192.168.0.0/16 | Private (RFC 1918) |
| 224.0.0.0/4 | Multicast |
| 255.255.255.255 | Limited broadcast (all hosts on local network) |

### 127.0.0.1 — Loopback

Traffic to 127.0.0.1 never leaves the machine. The kernel short-circuits it back to the sending process. It's used for:
- Testing services locally (curl http://127.0.0.1:8080)
- Inter-process communication
- Health checks

The entire 127.0.0.0/8 range is loopback — not just 127.0.0.1.

### 169.254.x.x — Link-Local

When a device is configured for DHCP but no DHCP server responds, it auto-assigns itself an address in the 169.254.0.0/16 range. This is called APIPA (Automatic Private IP Addressing).

Seeing 169.254.x.x on your machine usually means: **DHCP is broken**.

```bash
# Check if you got a link-local address (bad sign)
ip addr show eth0 | grep "169.254"
# If you see this, investigate DHCP
```

---

## Private vs Public Addresses (RFC 1918)

RFC 1918 defines three ranges that are **never routed on the public internet**:

| Range | CIDR | Addresses | Typical use |
|-------|------|-----------|-------------|
| 10.0.0.0 – 10.255.255.255 | 10.0.0.0/8 | 16,777,216 | Large enterprises, cloud VPCs |
| 172.16.0.0 – 172.31.255.255 | 172.16.0.0/12 | 1,048,576 | Medium networks |
| 192.168.0.0 – 192.168.255.255 | 192.168.0.0/16 | 65,536 | Home networks, small offices |

### Why private addresses exist

With only ~4.3 billion IPv4 addresses and billions of devices, there aren't enough public IPs for everyone. Private addresses let millions of organizations use the same address ranges internally, with **NAT** translating to public addresses at the network boundary.

**At your home**: Your router has ONE public IP (e.g., 203.0.113.5). All your devices (phone, laptop, tablet) have private IPs (192.168.1.x). NAT on the router translates between them.

**In AWS/GCP/Azure**: Your VPC uses private IPs (10.0.0.0/16). Traffic to the internet goes through a NAT Gateway or Internet Gateway.

### How to check your public vs private IP

```bash
# Your private IP (local network)
ip addr show eth0 | grep "inet "
# e.g., 192.168.1.100/24

# Your public IP (as seen by the internet)
curl -s ifconfig.me
# e.g., 203.0.113.5
```

---

## How a Host Decides: Local or Remote?

When your machine wants to send to an IP, it must decide: is this IP on my local network (same subnet) or on a remote network?

**Algorithm**:
1. Take the destination IP
2. Bitwise AND it with your subnet mask
3. Compare the result with your own network address (your IP AND your mask)
4. If they match → **local** (send directly, ARP for the destination)
5. If they don't match → **remote** (send to the default gateway, ARP for the gateway)

**Example**:
```
My IP:        192.168.1.100
My mask:      255.255.255.0 (/24)
My network:   192.168.1.0

Destination:  192.168.1.50
Dest AND Mask = 192.168.1.0
My network =    192.168.1.0
Match → LOCAL → ARP for 192.168.1.50

Destination:  10.0.0.1
Dest AND Mask = 10.0.0.0  
My network =    192.168.1.0
No match → REMOTE → send to default gateway
```

```bash
# See your routing table to understand local vs remote decisions
ip route show
# 192.168.1.0/24 dev eth0 proto kernel scope link src 192.168.1.100
# default via 192.168.1.1 dev eth0
#
# Line 1: 192.168.1.0/24 is directly connected (local)
# Line 2: Everything else goes to 192.168.1.1 (default gateway)
```

---

## IPv4 Address Exhaustion

### Timeline

| Year | Event |
|------|-------|
| 1981 | IPv4 defined (RFC 791) — 4.3 billion addresses seemed plenty |
| 1993 | CIDR introduced — slowed exhaustion by eliminating wasteful classful allocation |
| 1996 | NAT deployed widely — multiplied address utility |
| 1998 | IPv6 specified (RFC 2460) — the "real" solution |
| 2011 | IANA allocated the last /8 blocks to Regional Internet Registries |
| 2015 | ARIN (North America) ran out |
| 2019 | RIPE (Europe) ran out |
| 2024 | IPv4 addresses are traded on markets for $30-60 per address |

### Why we haven't fully switched to IPv6

1. **NAT works** — private addresses + NAT give virtually unlimited internal addresses
2. **IPv4-only devices** — legacy hardware and software
3. **Cost of transition** — requires updating every device, router, firewall, monitoring tool
4. **Dual-stack complexity** — running IPv4 and IPv6 simultaneously is complex
5. **No perceived urgency** — NAT "solved" the problem for most organizations

### Current state

As of 2024, ~40% of Google's traffic comes over IPv6. Mobile networks heavily use IPv6 (cellular carriers were among the first to run out of IPv4). Most enterprise networks still use IPv4 internally with NAT.

---

## IPv6 Overview

IPv6 addresses are **128 bits** — written as 8 groups of 4 hex digits:

```
Full:        2001:0db8:85a3:0000:0000:8a2e:0370:7334
Shortened:   2001:db8:85a3::8a2e:370:7334
             (leading zeros dropped, one :: for longest run of zeros)
```

$2^{128} = 3.4 \times 10^{38}$ addresses — enough for every atom on Earth to have billions of addresses.

### Key differences from IPv4

| Feature | IPv4 | IPv6 |
|---------|------|------|
| Address size | 32 bits | 128 bits |
| Notation | Dotted decimal | Hex with colons |
| Header size | 20-60 bytes (variable) | 40 bytes (fixed) |
| Fragmentation | Routers can fragment | Only sender can fragment |
| Broadcast | Yes | No (replaced by multicast) |
| ARP | Yes | Replaced by NDP (ICMPv6) |
| DHCP | Required or manual | SLAAC can auto-configure |
| NAT | Common | Unnecessary (enough addresses) |
| IPsec | Optional | Mandatory in spec (not in practice) |

We don't do a full IPv6 deep dive in this curriculum — it would double the length. The concepts (subnetting, routing, etc.) are the same; only the address format and some protocol details differ.

---

## Linux: Working with IP Addresses

### Viewing addresses

```bash
# Show all interfaces with addresses
ip addr show
# Key fields:
# inet = IPv4 address
# inet6 = IPv6 address
# brd = broadcast address
# scope = link (link-local), global (routable), host (loopback)

# Show just one interface
ip addr show dev eth0

# Show only IPv4
ip -4 addr show

# Show only running interfaces
ip addr show up
```

### Adding and removing addresses

```bash
# Add an IP address
sudo ip addr add 10.0.0.10/24 dev eth0
# A Linux interface can have MULTIPLE IPs!

# Add a secondary IP
sudo ip addr add 10.0.0.11/24 dev eth0

# Remove an IP
sudo ip addr del 10.0.0.11/24 dev eth0

# Flush all IPs from an interface
sudo ip addr flush dev eth0
```

### Verifying connectivity

```bash
# Ping a local address
ping -c 3 192.168.1.1

# Ping with specific source IP (useful when you have multiple IPs)
ping -I 10.0.0.10 -c 3 10.0.0.1

# Check which source IP is used for a destination
ip route get 8.8.8.8
# Shows: 8.8.8.8 via 192.168.1.1 dev eth0 src 192.168.1.100
# The "src" tells you which of your IPs will be used
```

### Checking subnet membership

```bash
# Does this IP match my subnet?
# Your address: 192.168.1.100/24
# Calculate manually or use ipcalc:
sudo apt install ipcalc
ipcalc 192.168.1.100/24
# Network:   192.168.1.0/24
# Broadcast: 192.168.1.255
# HostMin:   192.168.1.1
# HostMax:   192.168.1.254
# Hosts/Net: 254

ipcalc 10.0.0.50/28
# Network:   10.0.0.48/28
# Broadcast: 10.0.0.63
# HostMin:   10.0.0.49
# HostMax:   10.0.0.62
# Hosts/Net: 14
```

---

## Key Takeaways

1. **IPv4 addresses are 32-bit numbers** split into network and host portions by the subnet mask
2. **The subnet mask determines the split** — all 1s = network bits, all 0s = host bits
3. **CIDR notation** (/24, /16, etc.) replaced wasteful classful addressing
4. **Private addresses** (10.x, 172.16-31.x, 192.168.x.x) are used internally with NAT
5. **Two addresses per subnet are reserved**: network address (all host bits 0) and broadcast (all host bits 1)
6. **Local vs remote decision**: bitwise AND with mask, compare to your network — determines whether to ARP directly or send to gateway
7. **IPv4 is exhausted** — NAT and IPv6 are the responses
8. **Binary fluency is non-negotiable** for subnetting and routing

---

## Next

→ [02-cidr-subnetting.md](02-cidr-subnetting.md) — Subnetting in practice: how to divide networks
