# ip and ss — Modern Linux Network Commands

> `ip` and `ss` replaced the legacy `ifconfig`, `route`, `arp`, and `netstat` commands. They're faster, more powerful, and actually maintained. If you're still using `ifconfig`, stop. This file covers everything you need from these two essential tools.

---

## Table of Contents

1. [Why ip and ss?](#why)
2. [ip link — Interfaces](#ip-link)
3. [ip addr — Addresses](#ip-addr)
4. [ip route — Routing Table](#ip-route)
5. [ip neigh — ARP/Neighbor Table](#ip-neigh)
6. [ip netns — Network Namespaces](#ip-netns)
7. [ip rule and ip route — Policy Routing](#policy-routing)
8. [ss — Socket Statistics](#ss)
9. [ss Filtering and Examples](#ss-examples)
10. [Practical Scenarios](#scenarios)
11. [Key Takeaways](#key-takeaways)

---

## Why ip and ss?

### Legacy vs Modern

```
Legacy (net-tools)          Modern (iproute2)         Why modern is better
────────────────────────────────────────────────────────────────────────
ifconfig                    ip link / ip addr         Supports namespaces,
                                                      VLANs, bridges, bonds
route                       ip route                  Policy routing, tables
arp                         ip neigh                  NUD states, proxy ARP
netstat                     ss                        10-100× faster on busy
                                                      systems, filter syntax
brctl                       ip link (bridge)          Unified tool
tunctl                      ip tuntap                 Unified tool
```

```bash
# net-tools package: deprecated, may not be installed
# iproute2: always available on modern Linux

# Check what you have
which ip ss     # should exist
which ifconfig  # might not exist (and that's fine)
```

---

## ip link — Interfaces

### Viewing interfaces

```bash
# List all interfaces
ip link show
# 1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN
#     link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
# 2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UP
#     link/ether 52:54:00:12:34:56 brd ff:ff:ff:ff:ff:ff
# 3: docker0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 ...
#     link/ether 02:42:ac:11:00:01 brd ff:ff:ff:ff:ff:ff

# Short output
ip -br link show
# lo      UNKNOWN  00:00:00:00:00:00 <LOOPBACK,UP,LOWER_UP>
# eth0    UP       52:54:00:12:34:56 <BROADCAST,MULTICAST,UP,LOWER_UP>

# Single interface
ip link show eth0

# With statistics
ip -s link show eth0
# RX:  bytes  packets errors dropped  missed  mcast
#      12345  100     0      0        0       0
# TX:  bytes  packets errors dropped  carrier collisions
#      54321  200     0      0        0       0
```

### Interface flags

```
UP:          Interface is administratively up
LOWER_UP:    Physical link is up (cable connected)
BROADCAST:   Supports broadcast
MULTICAST:   Supports multicast
PROMISC:     Promiscuous mode (captures all frames)
NO-CARRIER:  No physical link detected

UP but no LOWER_UP = cable unplugged or link down
```

### Managing interfaces

```bash
# Bring interface up/down
ip link set eth0 up
ip link set eth0 down

# Change MTU
ip link set eth0 mtu 9000    # jumbo frames

# Change MAC address
ip link set eth0 down
ip link set eth0 address aa:bb:cc:dd:ee:ff
ip link set eth0 up

# Enable promiscuous mode (for packet capture)
ip link set eth0 promisc on

# Create virtual interfaces
ip link add veth0 type veth peer name veth1    # veth pair
ip link add br0 type bridge                     # bridge
ip link add bond0 type bond mode 802.3ad        # LACP bond
ip link add vlan100 link eth0 type vlan id 100  # VLAN
```

---

## ip addr — Addresses

### Viewing addresses

```bash
# All addresses
ip addr show
# 1: lo: <LOOPBACK,UP,LOWER_UP> ...
#     inet 127.0.0.1/8 scope host lo
#     inet6 ::1/128 scope host
# 2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> ...
#     inet 10.0.0.5/24 brd 10.0.0.255 scope global eth0
#     inet6 fe80::5054:ff:fe12:3456/64 scope link

# Brief format (very useful)
ip -br addr show
# lo      UNKNOWN  127.0.0.1/8 ::1/128
# eth0    UP       10.0.0.5/24 fe80::5054:ff:fe12:3456/64

# Only IPv4
ip -4 addr show

# Only IPv6
ip -6 addr show
```

### Managing addresses

```bash
# Add address
ip addr add 10.0.0.10/24 dev eth0

# Add secondary address (multiple IPs on one interface)
ip addr add 10.0.0.11/24 dev eth0

# Remove address
ip addr del 10.0.0.10/24 dev eth0

# Flush all addresses on interface
ip addr flush dev eth0
```

### Scope

```
global:   Routable everywhere (normal addresses)
link:     Only valid on this link (fe80:: link-local, 169.254.x.x)
host:     Only valid on this host (127.0.0.1 loopback)
```

---

## ip route — Routing Table

### Viewing routes

```bash
# Show routing table
ip route show
# default via 10.0.0.1 dev eth0 proto dhcp metric 100
# 10.0.0.0/24 dev eth0 proto kernel scope link src 10.0.0.5 metric 100
# 172.17.0.0/16 dev docker0 proto kernel scope link src 172.17.0.1

# Breakdown:
# default via 10.0.0.1   → gateway for everything not matched below
# 10.0.0.0/24 dev eth0   → this subnet is directly connected
# 172.17.0.0/16 dev docker0 → Docker network

# Show route to a specific destination
ip route get 8.8.8.8
# 8.8.8.8 via 10.0.0.1 dev eth0 src 10.0.0.5 uid 1000
#   Shows: next hop, output device, source IP selected

ip route get 10.0.0.100
# 10.0.0.100 dev eth0 src 10.0.0.5
#   No "via" = directly connected (same subnet), ARP to resolve
```

### Managing routes

```bash
# Add route
ip route add 192.168.1.0/24 via 10.0.0.1 dev eth0

# Add default gateway
ip route add default via 10.0.0.1

# Delete route
ip route del 192.168.1.0/24

# Replace route (add or update)
ip route replace 192.168.1.0/24 via 10.0.0.2

# Add blackhole route (silently drop)
ip route add blackhole 10.10.0.0/16

# Add unreachable route (ICMP unreachable)
ip route add unreachable 10.20.0.0/16
```

### Route table fields

```
Destination        via Gateway     dev Interface    proto Origin     metric Priority
────────────────────────────────────────────────────────────────────────────────
default            via 10.0.0.1   dev eth0         proto dhcp        100
10.0.0.0/24                       dev eth0         proto kernel      100
192.168.1.0/24     via 10.0.0.2   dev eth0         proto static      50

Proto:
  kernel   = auto-created when address added
  dhcp     = learned from DHCP
  static   = manually added
  bird/bgp = from routing daemon
```

---

## ip neigh — ARP/Neighbor Table

### Viewing ARP cache

```bash
# Show ARP table
ip neigh show
# 10.0.0.1 dev eth0 lladdr 52:54:00:ab:cd:ef REACHABLE
# 10.0.0.2 dev eth0 lladdr 52:54:00:11:22:33 STALE
# 10.0.0.3 dev eth0  FAILED

# NUD (Neighbor Unreachability Detection) states:
# REACHABLE   → Recently confirmed (just heard from them)
# STALE       → Haven't heard recently (will reverify on next use)
# DELAY       → Waiting a short time before probing
# PROBE       → Actively sending ARP/NDP probes
# FAILED      → Could not reach (no ARP reply)
# INCOMPLETE  → ARP request sent, waiting for reply
# PERMANENT   → Manually configured (static ARP)
```

### Managing ARP entries

```bash
# Add static ARP entry
ip neigh add 10.0.0.100 lladdr aa:bb:cc:dd:ee:ff dev eth0

# Delete entry
ip neigh del 10.0.0.100 dev eth0

# Flush all entries (force re-ARP)
ip neigh flush all

# Change entry state
ip neigh change 10.0.0.100 lladdr aa:bb:cc:dd:ee:ff dev eth0 nud permanent
```

---

## ip netns — Network Namespaces

### What namespaces are

```
Network namespaces provide isolated network stacks:
  - Each namespace has its own interfaces, routing table, firewall rules
  - Containers (Docker, Kubernetes) use namespaces
  - Each namespace is like a separate machine

Default namespace:
  Everything runs in the default namespace unless explicitly created
```

### Working with namespaces

```bash
# Create namespace
ip netns add test-ns

# List namespaces
ip netns list
# test-ns

# Run command inside namespace
ip netns exec test-ns ip link show
# 1: lo: <LOOPBACK> mtu 65536 ...   (only loopback — isolated!)

# Connect namespaces with veth pair
ip link add veth0 type veth peer name veth1
ip link set veth1 netns test-ns

# Configure addresses
ip addr add 10.200.0.1/24 dev veth0
ip link set veth0 up
ip netns exec test-ns ip addr add 10.200.0.2/24 dev veth1
ip netns exec test-ns ip link set veth1 up
ip netns exec test-ns ip link set lo up

# Test connectivity
ip netns exec test-ns ping 10.200.0.1

# Delete namespace
ip netns del test-ns
```

---

## Policy Routing

### Multiple routing tables

```bash
# Linux has 256 routing tables (0-255)
# Main table = 254 (what ip route show shows by default)

# View all tables
ip rule show
# 0:    from all lookup local      ← loopback, broadcast
# 32766: from all lookup main      ← default table
# 32767: from all lookup default   ← fallback

# Add rule: traffic from 10.0.1.0/24 uses table 100
ip rule add from 10.0.1.0/24 table 100

# Add route to table 100
ip route add default via 10.0.2.1 table 100

# Result: traffic from 10.0.1.0/24 exits via 10.0.2.1 gateway
# All other traffic uses the main table's default gateway

# Use case: Dual-WAN (two ISPs)
ip rule add from 10.0.0.0/24 table 10    # LAN1 → ISP1
ip rule add from 10.0.1.0/24 table 20    # LAN2 → ISP2
ip route add default via 1.1.1.1 table 10
ip route add default via 2.2.2.2 table 20
```

---

## ss — Socket Statistics

### Basic usage

```bash
# All TCP connections (ESTABLISHED)
ss -tn
# State  Recv-Q  Send-Q  Local Address:Port  Peer Address:Port
# ESTAB  0       0       10.0.0.5:80         10.0.0.2:54321
# ESTAB  0       36      10.0.0.5:80         10.0.0.3:54322

# All listening sockets
ss -tlnp
# State  Recv-Q  Send-Q  Local Address:Port  Peer Address:Port  Process
# LISTEN  0      128     0.0.0.0:80          0.0.0.0:*          users:(("nginx",pid=1234,fd=6))
# LISTEN  0      128     0.0.0.0:443         0.0.0.0:*          users:(("nginx",pid=1234,fd=7))
# LISTEN  0      128     127.0.0.1:5432      0.0.0.0:*          users:(("postgres",pid=5678,fd=3))

# UDP sockets
ss -ulnp

# Unix domain sockets
ss -xlp

# All sockets (everything)
ss -anp
```

### Flag reference

```
-t  TCP
-u  UDP
-x  Unix
-w  Raw
-l  Listening only
-a  All (listening + non-listening)
-n  Numeric (no DNS resolution — MUCH faster)
-p  Show process (requires root for other users' processes)
-m  Show memory usage
-i  Show TCP internal info (cwnd, rtt, etc.)
-e  Extended info (uid, inode)
-s  Summary statistics
-o  Show timers
```

### Understanding ss output

```
For LISTEN sockets:
  Recv-Q = connections in accept queue (waiting for accept())
  Send-Q = backlog (maximum accept queue size)
  
  If Recv-Q → Send-Q: application too slow at accepting!

For ESTABLISHED sockets:
  Recv-Q = bytes in receive buffer (not yet read by app)
  Send-Q = bytes in send buffer (not yet ACKed by peer)
  
  Large Recv-Q: application reading too slowly
  Large Send-Q: network slow, peer slow, or congestion
```

---

## ss Filtering and Examples

### Filter syntax

```bash
# By state
ss -tn state established
ss -tn state time-wait
ss -tn state close-wait
ss -tn state listening
ss -tn state syn-sent

# By port
ss -tn 'sport == :80'          # source port 80
ss -tn 'dport == :443'         # destination port 443
ss -tn '( sport == :80 or sport == :443 )'

# By address
ss -tn 'dst 10.0.0.0/24'      # destination in subnet
ss -tn 'src 10.0.0.5'          # source IP

# Combined
ss -tn 'sport == :80 and dst 10.0.0.2'

# By port range
ss -tn 'sport >= :8000 and sport <= :9000'
```

### Practical commands

```bash
# Count connections by state
ss -tn | awk '{print $1}' | sort | uniq -c | sort -rn
#  5234 ESTAB
#   832 TIME-WAIT
#    12 CLOSE-WAIT
#     4 SYN-SENT

# Top talkers (most connections by remote IP)
ss -tn | awk '{print $5}' | cut -d: -f1 | sort | uniq -c | sort -rn | head
#   234 10.0.0.2
#   156 10.0.0.3
#    89 10.0.0.4

# Connections per port
ss -tn | awk '{print $4}' | rev | cut -d: -f1 | rev | sort -n | uniq -c | sort -rn | head
#  5000 80
#   234 443

# Watch connections in real time
watch -n 1 "ss -tn | wc -l"

# Memory usage per socket
ss -tnm | head -20

# TCP internals (cwnd, rtt, retransmits)
ss -tni state established | head -20
# cubic wscale:9,9 rto:204 rtt:1.234/0.567 ato:40 mss:1448
# cwnd:10 bytes_sent:12345 bytes_received:67890
# send 93.9Mbps pacing_rate 187.8Mbps delivery_rate 89.0Mbps
```

---

## Practical Scenarios

### Scenario: Find what's listening on port 8080

```bash
ss -tlnp | grep :8080
# LISTEN  0  128  *:8080  *:*  users:(("java",pid=12345,fd=42))

# Or with lsof
lsof -i :8080
```

### Scenario: Detect connection leak (CLOSE_WAIT)

```bash
# CLOSE_WAIT means: remote side closed, but LOCAL app didn't close()
ss -tn state close-wait | wc -l
# 5000  ← Application bug! Not closing sockets.

# Find which process
ss -tnp state close-wait | awk '{print $6}' | sort | uniq -c | sort -rn
#  4800 users:(("java",pid=12345,fd=...))  ← This process is leaking
```

### Scenario: Network unreachable?

```bash
# 1. Check interface
ip -br link show
# eth0  UP  52:54:00:12:34:56

# 2. Check address
ip -br addr show
# eth0  UP  10.0.0.5/24

# 3. Check default route
ip route show default
# default via 10.0.0.1 dev eth0

# 4. Check gateway reachability
ip neigh show 10.0.0.1
# 10.0.0.1 dev eth0 lladdr 52:54:00:ab:cd:ef REACHABLE

# 5. Check DNS
ss -ulnp | grep :53
```

### Scenario: Change IP without downtime

```bash
# Add new IP first
ip addr add 10.0.0.100/24 dev eth0

# Verify both IPs work
ip addr show eth0
# inet 10.0.0.5/24 ...
# inet 10.0.0.100/24 ...

# Remove old IP
ip addr del 10.0.0.5/24 dev eth0
```

---

## Key Takeaways

1. **Use `ip` and `ss`**, not `ifconfig` and `netstat` — they're faster, more capable, and actively maintained
2. **`ip -br` gives clean output** — `ip -br addr show` and `ip -br link show` are your daily drivers
3. **`ip route get <dst>`** shows exactly how a packet would be routed — invaluable for debugging
4. **`ss -tlnp`** answers "what's listening on what port?" instantly
5. **Always use `-n` flag** with `ss` — skipping DNS resolution makes it 100× faster
6. **Watch Recv-Q on LISTEN sockets** — if it grows toward Send-Q, your application can't accept() fast enough
7. **CLOSE_WAIT = application bug** — remote closed, local app didn't. It's always a code issue
8. **Network namespaces** provide full network isolation — each namespace is a virtual machine's worth of network stack
9. **Policy routing** lets you route different traffic through different gateways based on source IP, marking, etc.
10. **`ip neigh` reveals ARP issues** — FAILED state means can't reach neighbor, STALE means cached but old

---

## Next

→ [ping and traceroute](./02-ping-traceroute.md) — Connectivity testing and path discovery
