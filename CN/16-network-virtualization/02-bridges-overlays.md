# Bridges, VLANs, and Overlay Networks

> Single-host networking is easy: connect namespaces with a bridge. Multi-host networking is the real challenge. How does Container A on Host 1 talk to Container B on Host 2 when they're on completely different networks? The answer is overlay networks — tunnels that create the illusion of a flat Layer 2 network spanning multiple hosts.

---

## Table of Contents

1. [Linux Bridges Deep Dive](#bridges)
2. [Bridge Forwarding and MAC Learning](#mac-learning)
3. [VLANs — Virtual LANs](#vlans)
4. [The Multi-Host Problem](#multi-host)
5. [VXLAN — The Overlay Standard](#vxlan)
6. [VXLAN Internals](#vxlan-internals)
7. [Building a VXLAN Overlay Manually](#vxlan-lab)
8. [GRE Tunnels](#gre)
9. [IPsec Tunnels](#ipsec)
10. [WireGuard](#wireguard)
11. [Overlay Network Comparison](#comparison)
12. [Performance Implications](#performance)
13. [Key Takeaways](#key-takeaways)

---

## Linux Bridges Deep Dive

### What a bridge does

```
A Linux bridge is a virtual Layer 2 switch.

Physical switch:
  Port 1 ─┐
  Port 2 ─┤── Switch ──→ Forwards frames by MAC address
  Port 3 ─┘

Linux bridge:
  veth1 ─┐
  veth2 ─┤── br0 ──→ Forwards frames by MAC address
  eth0  ─┘

  It's the same thing, in software.
```

### Bridge operations

```bash
# Create a bridge
sudo ip link add br0 type bridge

# Add ports to the bridge
sudo ip link set veth1 master br0
sudo ip link set veth2 master br0

# Optionally add a physical interface
# (bridges the host's network with virtual network)
sudo ip link set eth0 master br0

# Bring everything up
sudo ip link set br0 up
sudo ip link set veth1 up
sudo ip link set veth2 up

# Give the bridge an IP (makes hosts accessible through the bridge)
sudo ip addr add 10.0.0.1/24 dev br0

# Show bridge configuration
bridge link show
# veth1: <BROADCAST,MULTICAST,UP> mtu 1500 master br0 state forwarding
# veth2: <BROADCAST,MULTICAST,UP> mtu 1500 master br0 state forwarding

# Show MAC address table (forwarding database)
bridge fdb show br br0
# aa:bb:cc:dd:ee:01 dev veth1 master br0
# aa:bb:cc:dd:ee:02 dev veth2 master br0
```

---

## Bridge Forwarding and MAC Learning

```
How the bridge learns which MAC is on which port:

Step 1: Frame arrives on port veth1
  Source MAC = AA:BB:CC:DD:EE:01
  Bridge records: "AA:BB:CC:DD:EE:01 is on port veth1"

Step 2: Destination MAC lookup
  Dest MAC = AA:BB:CC:DD:EE:02
  Bridge checks FDB: "AA:BB:CC:DD:EE:02 is on port veth2"
  → Forward frame to veth2 ONLY

Unknown destination:
  If bridge doesn't know the destination MAC yet:
  → FLOOD: send frame to ALL ports (except the source port)
  → When destination replies, bridge learns its port
  
This is identical to how physical Ethernet switches work.

  ┌──────────────────────────────────────────────┐
  │  Bridge FDB (Forwarding Database)            │
  │                                              │
  │  MAC Address          Port     Age           │
  │  ──────────────────   ─────    ────          │
  │  aa:bb:cc:dd:ee:01    veth1    30s           │
  │  aa:bb:cc:dd:ee:02    veth2    15s           │
  │  aa:bb:cc:dd:ee:03    veth3    5s            │
  └──────────────────────────────────────────────┘
```

### STP (Spanning Tree Protocol)

```bash
# Bridges support STP to prevent loops
# (same as physical switches)

# Check STP status
bridge link show
# Look for "state forwarding" or "state blocking"

# Disable STP (for simple setups)
sudo ip link set br0 type bridge stp_state 0

# When STP is ON:
# - New ports start in "listening" state
# - Takes ~30 seconds to transition to "forwarding"
# - This causes delays when containers start!

# Docker disables STP on docker0 bridge (no loops expected)
# Kubernetes CNI plugins typically disable STP too
```

---

## VLANs — Virtual LANs

### What VLANs do

```
VLANs partition a switch into virtual switches:

Physical switch with VLANs:
  ┌─────────────────────────────────────┐
  │  Switch                             │
  │                                     │
  │  VLAN 10 (Engineering):             │
  │  ├── Port 1                         │
  │  ├── Port 2                         │
  │                                     │
  │  VLAN 20 (Marketing):              │
  │  ├── Port 3                         │
  │  ├── Port 4                         │
  │                                     │
  │  Port 1 and Port 3 CANNOT talk      │
  │  (different VLANs = different        │
  │   broadcast domains)                │
  └─────────────────────────────────────┘

Frames get a VLAN tag (802.1Q):
  ┌────────┬──────────┬────────┬─────────┬─────────┐
  │ Dst MAC│ Src MAC  │802.1Q  │ EtherType│ Payload │
  │        │          │Tag     │         │         │
  └────────┴──────────┼────────┼─────────┴─────────┘
                      │TPID:   │
                      │0x8100  │
                      │VLAN ID:│
                      │10      │
                      └────────┘
```

### Linux VLAN interfaces

```bash
# Create a VLAN interface on eth0
sudo ip link add link eth0 name eth0.10 type vlan id 10
sudo ip addr add 10.10.0.1/24 dev eth0.10
sudo ip link set eth0.10 up

# Create another VLAN
sudo ip link add link eth0 name eth0.20 type vlan id 20
sudo ip addr add 10.20.0.1/24 dev eth0.20
sudo ip link set eth0.20 up

# Traffic on eth0.10 is tagged with VLAN 10
# Traffic on eth0.20 is tagged with VLAN 20
# They're isolated at Layer 2

# Check VLAN config
cat /proc/net/vlan/eth0.10
```

---

## The Multi-Host Problem

```
Single host: easy — use a bridge.

  Host 1:
  ┌────────────────────────┐
  │  Container A           │
  │  10.0.0.2  ═══ br0 ═══ Container B    
  │                10.0.0.3│
  └────────────────────────┘

Multi-host: how does Container A talk to Container C?

  Host 1:                          Host 2:
  ┌───────────────┐                ┌───────────────┐
  │ Container A   │    ???         │ Container C   │
  │ 10.0.0.2      │════════════════│ 10.0.0.4      │
  │ br0           │  Physical net  │ br0           │
  └───────────────┘  192.168.1.x   └───────────────┘

Problems:
  1. 10.0.0.0/24 is on BOTH hosts (not routable between them)
  2. MAC addresses are local to each bridge (can't switch across)
  3. Physical network only knows about 192.168.1.x

Solutions:
  Option A: Route container traffic (each host gets unique subnet)
     Host 1: 10.0.1.0/24, Host 2: 10.0.2.0/24
     Add routes on the physical network
     → Simple but requires infrastructure changes

  Option B: Overlay network (tunnel container traffic)
     Encapsulate container frames in host-to-host packets
     → Works on any network, no infrastructure changes
     → This is what VXLAN does
```

---

## VXLAN — The Overlay Standard

### How VXLAN works

```
VXLAN = Virtual Extensible LAN

The idea: Encapsulate Layer 2 frames inside Layer 3 (UDP) packets.

Original container frame:
  ┌─────────┬─────────┬─────────────┐
  │ Cont.   │ Cont.   │  Payload    │
  │ Dst MAC │ Src MAC │             │
  └─────────┴─────────┴─────────────┘

Encapsulated in VXLAN:
  ┌─────────┬─────────┬──────┬──────┬────────┬─────────┬─────────┬─────────────┐
  │ Host    │ Host    │ IP   │ UDP  │ VXLAN  │ Cont.   │ Cont.   │  Payload    │
  │ Dst MAC │ Src MAC │ Hdr  │ 4789 │ Header │ Dst MAC │ Src MAC │             │
  └─────────┴─────────┴──────┴──────┴────────┴─────────┴─────────┴─────────────┘
  ←── Outer headers (host network) ──→←── Inner frame (container network) ──→

The physical network sees a normal UDP packet!
  Source: Host 1 (192.168.1.10)
  Dest:   Host 2 (192.168.1.20)
  Port:   UDP 4789

Inside that UDP packet is the ORIGINAL container frame.
The physical network doesn't know or care about container IPs/MACs.
```

### VXLAN Network Identifier (VNI)

```
VXLAN header includes a 24-bit VNI (like a VLAN ID):

  VNI range: 0 – 16,777,215 (vs VLAN's 0-4095)
  
  VNI 100 = Container network "red"
  VNI 200 = Container network "blue"
  
  Containers in VNI 100 can talk to each other.
  Containers in VNI 200 can talk to each other.
  VNI 100 and VNI 200 are isolated.

This is how Kubernetes network policies and 
Docker overlay networks achieve multi-tenancy.
```

---

## VXLAN Internals

### VTEP — VXLAN Tunnel Endpoint

```
Each host runs a VTEP (VXLAN Tunnel Endpoint):

  Host 1 (VTEP: 192.168.1.10)        Host 2 (VTEP: 192.168.1.20)
  ┌────────────────────┐              ┌────────────────────┐
  │ Container A        │              │ Container C        │
  │ 10.0.0.2           │              │ 10.0.0.4           │
  │     │               │              │     │               │
  │  ┌──┴───┐           │              │  ┌──┴───┐           │
  │  │ br0  │           │              │  │ br0  │           │
  │  └──┬───┘           │              │  └──┬───┘           │
  │  ┌──┴──────┐        │              │  ┌──┴──────┐        │
  │  │ vxlan0  │← VTEP  │              │  │ vxlan0  │← VTEP  │
  │  │ VNI=100 │        │              │  │ VNI=100 │        │
  │  └──┬──────┘        │              │  └──┬──────┘        │
  │     │  eth0          │              │     │  eth0          │
  │  192.168.1.10       │              │  192.168.1.20       │
  └────────┬────────────┘              └────────┬────────────┘
           │                                    │
           └────── Physical Network ────────────┘
           
Packet flow: Container A → Container C
  1. Container A sends frame to br0
  2. br0 forwards to vxlan0 (VTEP)
  3. VTEP encapsulates: wrap in UDP with outer IP headers
     Outer: 192.168.1.10 → 192.168.1.20, UDP 4789
     Inner: Container A MAC → Container C MAC
  4. Physical network delivers to Host 2
  5. Host 2's VTEP decapsulates
  6. Inner frame delivered to br0 → Container C
```

### MAC learning in VXLAN

```
How does VTEP on Host 1 know Container C is on Host 2?

Option 1: Multicast learning (manual/classic)
  - Unknown destination → VTEP floods to all VTEPs via multicast
  - VTEPs learn MAC→VTEP mappings from return traffic
  - Like bridge MAC learning, but across hosts

Option 2: Static FDB entries (explicit)
  - Admin/orchestrator programs MAC→VTEP mappings
  bridge fdb append 00:00:00:00:00:00 dev vxlan0 dst 192.168.1.20
  - Used by simpler setups

Option 3: Control plane (Kubernetes/Docker)
  - Central controller tells each VTEP about all MACs
  - Flannel, Calico, Weave, Cilium do this
  - Most scalable, no flooding
```

---

## Building a VXLAN Overlay Manually

### Prerequisites

```
Two Linux hosts that can reach each other:
  Host 1: 192.168.1.10
  Host 2: 192.168.1.20

We'll create an overlay network: 10.0.0.0/24
  Container on Host 1: 10.0.0.2
  Container on Host 2: 10.0.0.3
```

### On Host 1

```bash
# Create VXLAN interface
sudo ip link add vxlan0 type vxlan \
  id 100 \
  dstport 4789 \
  remote 192.168.1.20 \
  local 192.168.1.10 \
  dev eth0

# Create bridge
sudo ip link add br0 type bridge
sudo ip link set vxlan0 master br0

# Create namespace (container)
sudo ip netns add c1
sudo ip link add veth-c1 type veth peer name veth-c1-br
sudo ip link set veth-c1 netns c1
sudo ip link set veth-c1-br master br0

# Configure
sudo ip netns exec c1 ip addr add 10.0.0.2/24 dev veth-c1
sudo ip netns exec c1 ip link set veth-c1 up
sudo ip netns exec c1 ip link set lo up

# Bring everything up
sudo ip link set vxlan0 up
sudo ip link set br0 up
sudo ip link set veth-c1-br up
```

### On Host 2

```bash
# Same but with different addresses
sudo ip link add vxlan0 type vxlan \
  id 100 \
  dstport 4789 \
  remote 192.168.1.10 \
  local 192.168.1.20 \
  dev eth0

sudo ip link add br0 type bridge
sudo ip link set vxlan0 master br0

sudo ip netns add c2
sudo ip link add veth-c2 type veth peer name veth-c2-br
sudo ip link set veth-c2 netns c2
sudo ip link set veth-c2-br master br0

sudo ip netns exec c2 ip addr add 10.0.0.3/24 dev veth-c2
sudo ip netns exec c2 ip link set veth-c2 up
sudo ip netns exec c2 ip link set lo up

sudo ip link set vxlan0 up
sudo ip link set br0 up
sudo ip link set veth-c2-br up
```

### Test it

```bash
# From Host 1's container
sudo ip netns exec c1 ping 10.0.0.3
# PING 10.0.0.3 (10.0.0.3) 56(84) bytes of data.
# 64 bytes from 10.0.0.3: icmp_seq=1 ttl=64 time=0.543 ms
# ← SUCCESS! Container-to-container across hosts!

# Verify on the wire (capture on Host 1's eth0)
sudo tcpdump -i eth0 udp port 4789 -nn
# 192.168.1.10.44556 > 192.168.1.20.4789: VXLAN, flags [I] (0x08),
#   vni 100: ICMP echo request

# You can see:
#   Outer: 192.168.1.10 → 192.168.1.20  (host IPs)
#   Protocol: UDP port 4789 (VXLAN)
#   VNI: 100
#   Inner: ICMP echo request (container ping)
```

---

## GRE Tunnels

### Generic Routing Encapsulation

```
GRE is simpler than VXLAN — just wraps packets in IP:

  Original packet:
  ┌─────────┬──────────┐
  │ Inner IP│ Payload  │
  │  Header │          │
  └─────────┴──────────┘

  GRE encapsulated:
  ┌──────────┬─────────┬─────────┬──────────┐
  │ Outer IP │ GRE     │ Inner IP│ Payload  │
  │ Header   │ Header  │ Header  │          │
  └──────────┴─────────┴─────────┴──────────┘

  GRE = Layer 3 overlay (IP over IP)
  VXLAN = Layer 2 overlay (Ethernet over UDP/IP)
```

### GRE setup

```bash
# On Host 1 (192.168.1.10)
sudo ip tunnel add gre1 mode gre \
  remote 192.168.1.20 \
  local 192.168.1.10 \
  ttl 255
sudo ip addr add 10.0.0.1/30 dev gre1
sudo ip link set gre1 up

# On Host 2 (192.168.1.20)
sudo ip tunnel add gre1 mode gre \
  remote 192.168.1.10 \
  local 192.168.1.20 \
  ttl 255
sudo ip addr add 10.0.0.2/30 dev gre1
sudo ip link set gre1 up

# Test
ping 10.0.0.2  # from Host 1
# Works! Traffic tunneled through GRE

# GRE is point-to-point (not multipoint like VXLAN)
# Good for site-to-site connections
# No encryption (unlike IPsec or WireGuard)
```

---

## IPsec Tunnels

```
IPsec adds encryption to tunnels:

Modes:
  Transport Mode: Encrypts payload, keeps original IP header
  Tunnel Mode: Encrypts entire original packet, adds new IP header

  Tunnel mode:
  ┌──────────┬──────────┬──────────────────────────┐
  │ New IP   │ ESP      │ Encrypted(Original IP +  │
  │ Header   │ Header   │ Original Payload)        │
  └──────────┴──────────┴──────────────────────────┘

  ESP = Encapsulating Security Payload
  
Components:
  IKE: Internet Key Exchange (negotiate keys)
  ESP: Encrypt and authenticate packets
  SA:  Security Association (agreed parameters)

IPsec is used for:
  - VPN connections (site-to-site)
  - Securing GRE tunnels (GRE over IPsec)
  - Some Kubernetes CNIs (Calico with IPsec)

Downside:
  - Complex configuration
  - Kernel-space crypto = CPU overhead
  - Key management (IKE is complicated)
```

---

## WireGuard

```
WireGuard: Modern, simple, fast VPN tunnel.

Why WireGuard over IPsec:
  - ~4,000 lines of code vs 400,000+ for IPsec
  - Faster (in-kernel, optimized crypto)
  - Simpler configuration
  - Uses modern cryptography only (Curve25519, ChaCha20, BLAKE2)

Setup:
  # On Host 1
  sudo ip link add wg0 type wireguard
  sudo ip addr add 10.0.0.1/24 dev wg0
  wg genkey | tee privatekey | wg pubkey > publickey
  sudo wg set wg0 \
    listen-port 51820 \
    private-key ./privatekey \
    peer <HOST2_PUBLIC_KEY> \
    allowed-ips 10.0.0.2/32 \
    endpoint 192.168.1.20:51820
  sudo ip link set wg0 up

  # On Host 2 (mirror configuration)
  sudo ip link add wg0 type wireguard
  sudo ip addr add 10.0.0.2/24 dev wg0
  wg genkey | tee privatekey | wg pubkey > publickey
  sudo wg set wg0 \
    listen-port 51820 \
    private-key ./privatekey \
    peer <HOST1_PUBLIC_KEY> \
    allowed-ips 10.0.0.1/32 \
    endpoint 192.168.1.10:51820
  sudo ip link set wg0 up

WireGuard in container networking:
  - Cilium supports WireGuard for pod-to-pod encryption
  - Calico supports WireGuard for cross-node encryption
  - Encrypts overlay traffic between nodes
```

---

## Overlay Network Comparison

```
Technology   Layer   Encryption   Overhead   Use Case
──────────   ─────   ──────────   ────────   ────────
VXLAN        L2      No           50 bytes   Container networking (Flannel, Docker)
GRE          L3      No           24 bytes   Site-to-site tunnels
GRE+IPsec   L3      Yes          50+ bytes  Secure site-to-site
GENEVE       L2      No           Variable   Next-gen VXLAN (OVN, OpenStack)
WireGuard    L3      Yes          60 bytes   Encrypted tunnels, VPN
IPsec        L3/L4   Yes          50+ bytes  Legacy VPN, compliance

Container networking choices:
  ┌────────────────────────────────────────────────────┐
  │  CNI Plugin        Overlay Technology              │
  │  ──────────────    ────────────────────            │
  │  Flannel (VXLAN)   VXLAN                          │
  │  Flannel (host-gw) No overlay (direct routing)    │
  │  Calico (IPIP)     IP-in-IP tunnel                │
  │  Calico (VXLAN)    VXLAN                          │
  │  Calico (BGP)      No overlay (native routing)    │
  │  Cilium (VXLAN)    VXLAN                          │
  │  Cilium (native)   No overlay (direct routing)    │
  │  Weave             VXLAN + sleeve (fallback)      │
  └────────────────────────────────────────────────────┘
```

---

## Performance Implications

### Overlay overhead

```
VXLAN encapsulation cost:
  50 bytes per packet (outer Ethernet + IP + UDP + VXLAN headers)
  
  Original MTU: 1500 bytes
  Effective MTU: 1450 bytes (50 bytes consumed by encapsulation)
  
  If inner applications use MTU 1500:
    → Fragmentation! → Performance degradation
    
  Fix: 
    - Set inner MTU to 1450 (most CNI plugins do this)
    - Or use jumbo frames on physical network (MTU 9000)
      → Inner MTU can be 8950, effectively 1500+ usable

CPU cost:
  - VXLAN encap/decap: minimal (kernel handles efficiently)
  - With encryption (IPsec/WireGuard): 10-30% throughput reduction
  - Hardware offload available for VXLAN on modern NICs
```

### When to avoid overlays

```
No overlay (direct routing) is better when:
  - You control the physical network (can add routes)
  - Maximum performance required (bare-metal, HPC)
  - Simpler debugging (no encapsulation to unwrap)

Calico BGP mode: No overlay
  - Announces pod subnets via BGP to physical routers
  - Pods are directly routable
  - Best performance, but requires BGP-capable infrastructure

Overlay is necessary when:
  - You DON'T control the physical network (cloud VPC)
  - Hosts are across different L2 domains
  - You need L2 adjacency across L3 boundaries
  - Multi-tenant isolation required
```

---

## Key Takeaways

1. **A Linux bridge is a software switch** — it forwards frames by MAC address, learns MACs automatically, and supports STP just like physical switches
2. **VLANs tag frames with an ID to create isolated Layer 2 domains** — but limited to 4095 VLANs. VXLAN supports 16 million VNIs
3. **The multi-host problem: containers on different hosts are on different Layer 2 networks** — overlay networks solve this by tunneling
4. **VXLAN encapsulates Layer 2 frames in UDP packets** — the physical network just sees normal UDP traffic. Inner container traffic is invisible to infrastructure
5. **VTEP (VXLAN Tunnel Endpoint) handles encap/decap on each host** — bridges connect local containers, VXLAN connects bridges across hosts
6. **VXLAN adds 50 bytes overhead** — inner MTU must be reduced to 1450, or use jumbo frames on the underlay. This is the #1 source of mysteriously broken connections
7. **GRE is simple IP-in-IP tunneling without encryption** — good for point-to-point, but no security
8. **WireGuard is the modern choice for encrypted tunnels** — simpler and faster than IPsec, used by Cilium and Calico for pod traffic encryption
9. **Avoid overlays when you can** — direct routing (Calico BGP, Flannel host-gw) gives better performance and simpler debugging
10. **Every container networking solution uses some combination of these primitives** — understanding bridges, VXLAN, and routing lets you debug any CNI plugin

---

## Next

→ [Docker and Kubernetes Networking](./03-docker-kubernetes-networking.md) — How Docker and Kubernetes build on these primitives
