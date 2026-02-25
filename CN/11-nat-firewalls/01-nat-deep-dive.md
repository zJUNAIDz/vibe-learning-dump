# NAT Deep Dive — Network Address Translation

> NAT is the reason the internet didn't run out of IPv4 addresses in the 1990s. It's also the source of countless networking headaches — broken peer-to-peer connections, port forwarding nightmares, and protocol incompatibilities. Understanding NAT is essential for any engineer working with networks.

---

## Table of Contents

1. [Why NAT Exists](#why-nat)
2. [How NAT Works](#how-nat-works)
3. [Types of NAT](#types)
4. [NAT Traversal](#traversal)
5. [CGNAT — Carrier-Grade NAT](#cgnat)
6. [NAT and IPv6](#ipv6)
7. [Linux: NAT with iptables/nftables](#linux-nat)
8. [Debugging NAT Issues](#debugging)
9. [Key Takeaways](#key-takeaways)

---

## Why NAT Exists

### The IPv4 address problem

```
IPv4 addresses: 32 bits → 2^32 = 4,294,967,296 addresses
World population: ~8 billion people
Devices per person: 3-10 (phone, laptop, tablet, IoT, etc.)
Required addresses: 24-80 billion
Available: 4.3 billion

The math doesn't work.
```

### Before NAT (early 1990s)

```
Every device had a public IP:
  Desktop:    198.51.100.10
  Printer:    198.51.100.11
  Server:     198.51.100.12
  
  → Company with 1000 devices needed 1000 public IPs
  → IPv4 exhaustion predicted by mid-1990s
```

### NAT's solution (RFC 1631, 1994)

```
Replace ONE public IP for MANY private IPs:

Private network (10.0.0.0/8):
  PC 1:     10.0.0.2
  PC 2:     10.0.0.3
  Phone:    10.0.0.4
  Printer:  10.0.0.5
  ...100s of devices

NAT Router: 10.0.0.1 (internal) ↔ 198.51.100.1 (public)

All devices share the ONE public IP: 198.51.100.1
The NAT router rewrites source/destination addresses.
```

### Private address ranges (RFC 1918)

```
10.0.0.0/8      → 16,777,216 addresses (10.0.0.0 – 10.255.255.255)
172.16.0.0/12   → 1,048,576 addresses  (172.16.0.0 – 172.31.255.255)
192.168.0.0/16  → 65,536 addresses     (192.168.0.0 – 192.168.255.255)

These addresses are NEVER routed on the public internet.
Any organization can use them internally.
```

---

## How NAT Works

### Outbound (SNAT — Source NAT)

```
Internal PC (10.0.0.2) wants to reach web server (93.184.216.34):

1. PC sends packet:
   Src: 10.0.0.2:54321 → Dst: 93.184.216.34:443

2. NAT router rewrites source:
   Src: 198.51.100.1:40000 → Dst: 93.184.216.34:443
   Records mapping: 10.0.0.2:54321 ↔ 198.51.100.1:40000

3. Server sees request from 198.51.100.1:40000
   Server responds:
   Src: 93.184.216.34:443 → Dst: 198.51.100.1:40000

4. NAT router looks up mapping, rewrites destination:
   Src: 93.184.216.34:443 → Dst: 10.0.0.2:54321

5. PC receives response. It doesn't know NAT happened.
```

### The NAT translation table

```
┌──────────────────┬──────────────────────┬───────────────────────┐
│ Internal (LAN)   │ External (WAN)       │ Destination           │
├──────────────────┼──────────────────────┼───────────────────────┤
│ 10.0.0.2:54321   │ 198.51.100.1:40000   │ 93.184.216.34:443    │
│ 10.0.0.3:55555   │ 198.51.100.1:40001   │ 142.250.80.100:80    │
│ 10.0.0.2:54322   │ 198.51.100.1:40002   │ 93.184.216.34:443    │
│ 10.0.0.4:12345   │ 198.51.100.1:40003   │ 1.1.1.1:53           │
└──────────────────┴──────────────────────┴───────────────────────┘

Each outbound connection gets a unique external port.
Entry created on first outbound packet, removed after timeout.
```

### Port Forwarding (DNAT — Destination NAT)

```
Problem: External users can't initiate connections to internal devices.
  → No mapping exists until internal device starts a connection.

Solution: Port forwarding — static mappings:

  "Forward port 8080 on public IP to 10.0.0.5:80"
  
  External request:
    Src: anywhere → Dst: 198.51.100.1:8080
  
  NAT rewrites:
    Src: anywhere → Dst: 10.0.0.5:80
  
  Use cases: Hosting servers, game servers, SSH access behind NAT
```

---

## Types of NAT

### Full Cone NAT (1-to-1)

```
Once internal→external mapping exists, ANY external host can send to that external port.

10.0.0.2:54321 → 198.51.100.1:40000
  Created by: 10.0.0.2 → server A
  After: server B, server C, ANYONE can send to 198.51.100.1:40000
         → reaches 10.0.0.2:54321

Most permissive. Easiest for P2P. Rare in practice.
```

### Restricted Cone NAT

```
External host can only send to mapped port IF the internal host previously sent to that host's IP.

10.0.0.2:54321 → 198.51.100.1:40000
  10.0.0.2 sent to: 142.250.80.100
  → 142.250.80.100:ANY_PORT can send to 198.51.100.1:40000 ✓
  → 93.184.216.34:ANY_PORT CANNOT (never contacted) ✗
```

### Port-Restricted Cone NAT

```
Like restricted cone, but also checks port:

10.0.0.2:54321 → 198.51.100.1:40000
  10.0.0.2 sent to: 142.250.80.100:443
  → 142.250.80.100:443 can send to 198.51.100.1:40000 ✓
  → 142.250.80.100:80 CANNOT (wrong port) ✗
  → 93.184.216.34:443 CANNOT (never contacted) ✗
  
Most common NAT type for home routers.
```

### Symmetric NAT

```
Different mapping for each destination:

10.0.0.2:54321 → server A: uses 198.51.100.1:40000
10.0.0.2:54321 → server B: uses 198.51.100.1:40001  (DIFFERENT port!)

Only the specific dest can reply to each mapping.
Most restrictive. Breaks many P2P protocols.
Common in enterprise/carrier-grade NAT.
```

---

## NAT Traversal

### The fundamental problem

```
Two devices behind different NATs can't connect directly:

Alice (10.0.0.2) behind NAT-A (198.51.100.1)
Bob   (10.0.0.3) behind NAT-B (203.0.113.1)

Alice → Bob: Alice doesn't know Bob's NAT external address
             Even if she did, NAT-B would DROP the packet
             (no mapping exists — Bob didn't initiate)

Bob → Alice: Same problem in reverse.

Neither side can initiate!
```

### STUN (Session Traversal Utilities for NAT)

```
STUN server (public IP) helps devices discover their external address:

1. Alice → STUN server: "What's my external IP:port?"
   STUN server sees packet from 198.51.100.1:40000
   STUN server → Alice: "You appear as 198.51.100.1:40000"

2. Alice shares this address with Bob (via signaling server)
3. Bob does the same

4. Both try to send to each other's external addresses simultaneously
   → "UDP hole punching"

Works with: Full Cone, Restricted Cone, Port-Restricted Cone NATs
Fails with: Symmetric NAT (different port per destination)
```

### TURN (Traversal Using Relays around NAT)

```
When direct connection fails (symmetric NAT), relay through a server:

Alice → TURN server → Bob

TURN server relays ALL traffic between Alice and Bob.

Downsides:
  - Higher latency (extra hop)
  - Higher bandwidth cost (server relays everything)
  - TURN server becomes single point of failure

Used by: WebRTC as last resort fallback
```

### ICE (Interactive Connectivity Establishment)

```
ICE combines STUN + TURN to find the best path:

1. Gather candidates:
   a) Host candidates: direct LAN addresses
   b) Server reflexive: external address (via STUN)
   c) Relay: TURN server address (fallback)

2. Connectivity checks: Try all candidate pairs
   - Direct connection? → use it (best)
   - STUN hole punch works? → use it (good)
   - Only TURN works? → use relay (acceptable)

Used by: WebRTC, VoIP (SIP/ICE)
```

### UDP hole punching

```
How both NATs get "opened":

1. Alice → Bob's external: Packet reaches NAT-B, DROPPED (no mapping)
   BUT: NAT-A now has mapping: Alice:port → Bob's external

2. Bob → Alice's external: Packet reaches NAT-A, ACCEPTED!
   (NAT-A has mapping from step 1)
   NAT-B now has mapping: Bob:port → Alice's external

3. Alice → Bob: Now NAT-B has mapping → ACCEPTED!

Both NATs are "punched" — bidirectional communication established.
Timing matters — both sides must send nearly simultaneously.
```

---

## CGNAT

### ISPs run out of public IPs too

```
Carrier-Grade NAT (CGNAT, RFC 6598): ISP puts customers behind NAT.

Without CGNAT:
  Customer → public IP 198.51.100.x → Internet

With CGNAT:
  Customer router (private) → Customer NAT → 
    100.64.x.x (shared address space) → ISP CGNAT → 
      public IP → Internet

= Double NAT!
```

### CGNAT address range

```
100.64.0.0/10  (RFC 6598 "Shared Address Space")
= 4,194,304 addresses for ISP ↔ customer link
NOT routable on the public internet
Different from RFC 1918 private addresses
```

### CGNAT problems

```
1. Port exhaustion:
   Each public IP has 65,535 ports
   Shared among hundreds of customers
   Heavy user (torrents, many tabs) exhausts port allocation

2. Hosting impossible:
   Can't port-forward through ISP's NAT
   → Self-hosting, game servers, SSH access broken

3. IP reputation:
   Hundreds of customers share one IP
   One bad actor → IP blacklisted → all customers affected

4. Geolocation:
   Public IP geolocates to ISP's NAT, not customer
   → Wrong location for location-based services

5. Logging:
   ISP must log ALL NAT translations for legal compliance
   → Massive storage and processing requirements
```

---

## NAT and IPv6

### IPv6 eliminates the need for NAT

```
IPv6: 128 bits → 2^128 = 340 undecillion addresses
      = 340,282,366,920,938,463,463,374,607,431,768,211,456

Every device gets a globally unique public address.
No NAT needed! Direct end-to-end connectivity restored.

But: Some organizations STILL use NAT66 (IPv6 NAT)
  → Misguided attempt to get "security" from NAT
  → NAT is NOT a security feature (use firewalls instead)
```

### Why NAT is not security

```
NAT hides internal IPs — but this is NOT security:
  - Any outbound connection creates a NAT mapping
  - That mapping allows inbound traffic (the response)
  - NAT doesn't inspect traffic content
  - NAT doesn't authenticate
  - NAT doesn't filter malware

Use firewalls for security, not NAT.
IPv6 + stateful firewall = correct approach.
```

---

## Linux: NAT with iptables/nftables

### iptables NAT (legacy, still widely used)

```bash
# Enable IP forwarding (required for routing)
echo 1 > /proc/sys/net/ipv4/ip_forward
# Or permanently:
echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.conf && sysctl -p

# SNAT: Masquerade outbound traffic (dynamic public IP)
iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE

# SNAT: Fixed public IP
iptables -t nat -A POSTROUTING -o eth0 -j SNAT --to-source 198.51.100.1

# DNAT: Port forwarding (external:8080 → internal:80)
iptables -t nat -A PREROUTING -p tcp --dport 8080 -j DNAT --to-destination 10.0.0.5:80
# Also need to allow forwarding:
iptables -A FORWARD -p tcp -d 10.0.0.5 --dport 80 -j ACCEPT

# View NAT rules
iptables -t nat -L -n -v

# View active NAT translations (conntrack)
conntrack -L
# Or: cat /proc/net/nf_conntrack
```

### nftables NAT (modern replacement)

```bash
# Create NAT table
nft add table nat
nft add chain nat postrouting { type nat hook postrouting priority 100 \; }
nft add chain nat prerouting { type nat hook prerouting priority -100 \; }

# SNAT (masquerade)
nft add rule nat postrouting oifname "eth0" masquerade

# DNAT (port forward)
nft add rule nat prerouting tcp dport 8080 dnat to 10.0.0.5:80

# View rules
nft list ruleset
```

### Connection tracking

```bash
# View all tracked connections
conntrack -L

# Example output:
# tcp  6 117 TIME_WAIT src=10.0.0.2 dst=93.184.216.34 sport=54321 dport=443
# src=93.184.216.34 dst=198.51.100.1 sport=443 dport=40000 mark=0 use=1

# Count connections
conntrack -C

# Monitor new connections in real-time
conntrack -E

# Clear all entries
conntrack -F

# Check conntrack table size
sysctl net.netfilter.nf_conntrack_max
# Default: 65536 — increase for busy NAT gateways!

# Check current usage
cat /proc/sys/net/netfilter/nf_conntrack_count
```

---

## Debugging

### Common NAT issues

```
Symptom: "Can't reach server from outside"
  → Check port forwarding rules
  → Check firewall allows forwarded traffic
  → NAT hairpin: internal client accessing public IP → needs hairpin NAT rule

Symptom: "Intermittent connection drops"
  → NAT table full (check conntrack_max)
  → NAT timeout (UDP default 30s, TCP established 5 days)
  → Adjust: sysctl net.netfilter.nf_conntrack_tcp_timeout_established

Symptom: "Application doesn't work through NAT"
  → ALG (Application Layer Gateway) needed for protocols 
    that embed IP addresses in payload (SIP, FTP active mode)
  → Check: lsmod | grep nf_nat_sip
```

### Useful commands

```bash
# What's my external IP?
curl -s https://ifconfig.me
curl -s https://api.ipify.org

# Trace packet through NAT
sudo tcpdump -i eth0 -n host 93.184.216.34

# Check if you're behind CGNAT
traceroute 8.8.8.8
# If you see 100.64.x.x addresses → you're behind CGNAT

# Test port forwarding from outside
# On external machine:
nc -zv 198.51.100.1 8080
# Or use: https://portchecker.co
```

---

## Key Takeaways

1. **NAT maps private IPs to public IPs** — lets many devices share one public address
2. **SNAT** rewrites source (outbound); **DNAT** rewrites destination (inbound/port forwarding)
3. **NAT table** tracks mappings using 5-tuple (src IP, src port, dst IP, dst port, protocol)
4. **NAT types** (Full Cone → Symmetric) vary in restrictiveness — affects P2P connectivity
5. **NAT traversal**: STUN discovers external address, hole punching opens NATs, TURN relays as fallback
6. **CGNAT** = ISP-level NAT — double NAT causes port exhaustion, hosting impossible, shared IP reputation
7. **NAT is NOT security** — use firewalls for security, not NAT
8. **IPv6 eliminates need for NAT** — every device gets globally unique address
9. **Linux NAT**: `iptables -t nat` (legacy) or `nft` (modern), conntrack tracks translations
10. **conntrack table size** is critical for busy NAT gateways — default 65536 may be too small

---

## Next

→ [02-iptables-nftables.md](02-iptables-nftables.md) — Linux firewall in depth
