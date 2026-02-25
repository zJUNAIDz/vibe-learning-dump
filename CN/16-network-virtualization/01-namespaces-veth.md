# Linux Network Namespaces and veth Pairs

> Network namespaces are the foundation of container networking. Before Docker or Kubernetes does anything, it's just Linux namespaces and veth pairs under the hood. Understanding these two primitives lets you understand — and debug — any container networking setup.

---

## Table of Contents

1. [What Are Network Namespaces?](#namespaces)
2. [Creating and Using Network Namespaces](#creating)
3. [What is a veth Pair?](#veth)
4. [Connecting Two Namespaces](#connecting)
5. [Namespace Isolation Deep Dive](#isolation)
6. [Namespace Network Configuration](#configuration)
7. [Routing Between Namespaces](#routing)
8. [Namespace Inspection and Debugging](#debugging)
9. [Practical Lab: Build Container Networking From Scratch](#lab)
10. [How Docker Uses Namespaces](#docker)
11. [Key Takeaways](#key-takeaways)

---

## What Are Network Namespaces?

### The problem they solve

```
Without namespaces:
  All processes share ONE network stack:
    - Same interfaces (eth0, lo)
    - Same routing table
    - Same iptables rules
    - Same port space (only one process on port 80)

  Process A wants port 80 ─┐
                            ├→ CONFLICT!
  Process B wants port 80 ─┘

With namespaces:
  Each namespace gets its OWN network stack:
  
  ┌─────────────────────┐    ┌─────────────────────┐
  │  Namespace "app1"   │    │  Namespace "app2"    │
  │                     │    │                      │
  │  eth0: 10.0.1.2     │    │  eth0: 10.0.2.2      │
  │  lo: 127.0.0.1      │    │  lo: 127.0.0.1       │
  │  Routes: ...        │    │  Routes: ...         │
  │  iptables: ...      │    │  iptables: ...       │
  │  Port 80: app1 ✓    │    │  Port 80: app2 ✓     │
  └─────────────────────┘    └──────────────────────┘
  
  Both can use port 80! They're isolated.
```

### What a namespace contains

```
Each network namespace has its own:
  ✓ Network interfaces (eth0, lo, etc.)
  ✓ IP addresses
  ✓ Routing table
  ✓ ARP table
  ✓ iptables/nftables rules
  ✓ /proc/net and /sys/class/net
  ✓ Socket port space
  ✓ Netfilter conntrack table

Things NOT namespaced (shared with host):
  ✗ Filesystem (unless also using mount namespace)
  ✗ Hostname (unless also using UTS namespace)
  ✗ PIDs (unless also using PID namespace)
  
  → Containers use ALL namespace types together
```

---

## Creating and Using Network Namespaces

### Basic operations

```bash
# Create a network namespace
sudo ip netns add red
sudo ip netns add blue

# List all network namespaces
ip netns list
# red
# blue

# Execute a command INSIDE a namespace
sudo ip netns exec red ip addr
# 1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN
#     link/loopback 00:00:00:00:00:00

# Notice: Only loopback, and it's DOWN
# Each new namespace starts with nothing but a down loopback

# Bring up loopback inside namespace
sudo ip netns exec red ip link set lo up

# Run a shell inside the namespace
sudo ip netns exec red bash
# Now every command runs in the "red" namespace
ip addr   # shows red's interfaces
exit      # back to host namespace
```

### The default namespace

```
The host runs in the "default" (init) network namespace.

It has NO name — it just exists:
  ip addr                    ← shows host interfaces
  ip netns exec red ip addr  ← shows red's interfaces

Processes inherit their parent's namespace by default.
  Only root (or CAP_NET_ADMIN) can create/switch namespaces.
```

---

## What is a veth Pair?

### Virtual Ethernet pairs

```
A veth pair is a virtual cable with two ends:

  ┌─────────┐                    ┌─────────┐
  │  veth0  │════════════════════│  veth1  │
  └─────────┘                    └─────────┘
  
  Whatever goes IN one end comes OUT the other.
  Like a virtual crossover cable.
  
  Each end is a full network interface:
    - Has its own MAC address
    - Can have IP addresses
    - Can be placed in different namespaces
    - Shows up in `ip link`

Key property:
  The two ends can be in DIFFERENT namespaces.
  This is how you connect namespaces together!
```

### Creating a veth pair

```bash
# Create a veth pair
sudo ip link add veth-red type veth peer name veth-blue

# Both ends exist in the host namespace initially
ip link show type veth
# veth-red@veth-blue: ...
# veth-blue@veth-red: ...

# The @ notation shows what they're paired with
```

---

## Connecting Two Namespaces

### Step by step

```bash
# Step 1: Create two namespaces
sudo ip netns add red
sudo ip netns add blue

# Step 2: Create a veth pair
sudo ip link add veth-red type veth peer name veth-blue

# Step 3: Move each end into its namespace
sudo ip link set veth-red netns red
sudo ip link set veth-blue netns blue

# After this, veth-red and veth-blue disappear from host!
ip link show type veth
# (nothing — they're in their namespaces now)

# Step 4: Assign IP addresses
sudo ip netns exec red ip addr add 10.0.0.1/24 dev veth-red
sudo ip netns exec blue ip addr add 10.0.0.2/24 dev veth-blue

# Step 5: Bring interfaces up
sudo ip netns exec red ip link set veth-red up
sudo ip netns exec blue ip link set veth-blue up
sudo ip netns exec red ip link set lo up
sudo ip netns exec blue ip link set lo up

# Step 6: TEST!
sudo ip netns exec red ping 10.0.0.2
# PING 10.0.0.2 (10.0.0.2) 56(84) bytes of data.
# 64 bytes from 10.0.0.2: icmp_seq=1 ttl=64 time=0.042 ms
# ← SUCCESS! Red can reach Blue!
```

### What we built

```
  ┌─────────────────┐              ┌─────────────────┐
  │  Namespace: red │              │ Namespace: blue  │
  │                 │              │                  │
  │  veth-red       │              │  veth-blue       │
  │  10.0.0.1/24    │══════════════│  10.0.0.2/24     │
  │                 │  veth pair   │                  │
  │  lo: 127.0.0.1  │              │  lo: 127.0.0.1   │
  └─────────────────┘              └──────────────────┘
  
  This is the simplest container-to-container network.
  Each namespace = one "container."
```

---

## Namespace Isolation Deep Dive

### Complete isolation

```bash
# From the HOST, can you reach the namespace?
ping 10.0.0.1
# FAIL — the host's routing table doesn't know about 10.0.0.0/24

# From red, can you reach the internet?
sudo ip netns exec red ping 8.8.8.8
# FAIL — red has no default route, no path to the internet

# From red, can you reach the host?
sudo ip netns exec red ping <host-ip>
# FAIL — no route to host's network

# This is REAL isolation:
# - Can't reach in
# - Can't reach out
# - Only the veth pair connects red ↔ blue

# Check red's routing table
sudo ip netns exec red ip route
# 10.0.0.0/24 dev veth-red proto kernel scope link src 10.0.0.1
# ← Only one route: the directly connected veth network
```

### Process isolation

```bash
# Start a server in the "red" namespace
sudo ip netns exec red python3 -m http.server 80 &

# From blue, access it
sudo ip netns exec blue curl http://10.0.0.1:80
# ← Works!

# From the host
curl http://10.0.0.1:80
# ← FAILS (host can't reach namespace)

# Port 80 in "red" is completely separate from port 80 on host
# You can run ANOTHER server on port 80 on the host — no conflict!
python3 -m http.server 80 &
# ← Works fine, separate port space
```

---

## Namespace Network Configuration

### Each namespace has its own everything

```bash
# Routing table per namespace
sudo ip netns exec red ip route show
# 10.0.0.0/24 dev veth-red ...

# ARP table per namespace
sudo ip netns exec red ip neigh show
# 10.0.0.2 dev veth-red lladdr xx:xx:xx:xx:xx:xx

# iptables per namespace
sudo ip netns exec red iptables -L
# Chain INPUT (policy ACCEPT)
# Chain FORWARD (policy ACCEPT)
# Chain OUTPUT (policy ACCEPT)
# ← Clean slate! No host rules here.

# Add firewall rules to namespace
sudo ip netns exec red iptables -A INPUT -p tcp --dport 22 -j DROP
# Only affects "red" namespace, not host or "blue"

# Sockets per namespace
sudo ip netns exec red ss -tlnp
# Shows only listeners in the "red" namespace
```

---

## Routing Between Namespaces

### Giving a namespace internet access

```bash
# Create a namespace with veth pair to host
sudo ip netns add container
sudo ip link add veth-host type veth peer name veth-cont

# Move one end into namespace
sudo ip link set veth-cont netns container

# Configure host side
sudo ip addr add 10.200.0.1/24 dev veth-host
sudo ip link set veth-host up

# Configure namespace side
sudo ip netns exec container ip addr add 10.200.0.2/24 dev veth-cont
sudo ip netns exec container ip link set veth-cont up
sudo ip netns exec container ip link set lo up

# Set default route in namespace (via host)
sudo ip netns exec container ip route add default via 10.200.0.1

# Now namespace can reach host at 10.200.0.1
sudo ip netns exec container ping 10.200.0.1
# ← Works!

# But still can't reach internet... why?
sudo ip netns exec container ping 8.8.8.8
# ← FAILS! Packets reach host but host doesn't forward/NAT them.

# Enable IP forwarding on host
echo 1 | sudo tee /proc/sys/net/ipv4/ip_forward

# Add NAT (masquerade) for namespace traffic
sudo iptables -t nat -A POSTROUTING -s 10.200.0.0/24 -j MASQUERADE

# NOW namespace can reach the internet!
sudo ip netns exec container ping 8.8.8.8
# ← Works!
```

### What we built

```
                    Internet
                       │
                   ┌───┴────┐
                   │ Router │
                   └───┬────┘
                       │
  ┌────────────────────┼──────────────────────────┐
  │  HOST              │                          │
  │                 eth0: 192.168.1.100           │
  │                    │                          │
  │          ┌─────────┴──────────┐               │
  │          │    IP Forwarding    │               │
  │          │    + NAT (iptables) │               │
  │          └─────────┬──────────┘               │
  │                    │                          │
  │           veth-host: 10.200.0.1               │
  │                    ║ veth pair                 │
  │  ┌─────────────────╨──────────────┐           │
  │  │  Namespace: container          │           │
  │  │  veth-cont: 10.200.0.2        │           │
  │  │  default route via 10.200.0.1  │           │
  │  └────────────────────────────────┘           │
  └───────────────────────────────────────────────┘

  This is EXACTLY what Docker does (with a bridge in the middle).
```

---

## Namespace Inspection and Debugging

### Finding namespace for a process

```bash
# What namespace is a process in?
ls -la /proc/<PID>/ns/net
# lrwxrwxrwx 1 root root 0 ... /proc/1234/ns/net -> 'net:[4026531992]'

# The number is the namespace inode
# Same number = same namespace

# Compare two processes:
readlink /proc/1/ns/net      # PID 1 (init) — host namespace
# net:[4026531992]
readlink /proc/1234/ns/net   # Container process
# net:[4026532400]            ← Different! In a different namespace.

# Enter a process's namespace
sudo nsenter --target <PID> --net
# Now you're in that process's network namespace
ip addr    # See its interfaces
ss -tlnp   # See its listeners
exit
```

### Finding Docker container namespaces

```bash
# Get the PID of a Docker container
docker inspect <container> --format '{{.State.Pid}}'
# 12345

# Enter the container's network namespace
sudo nsenter --target 12345 --net
ip addr       # See container's interfaces
ip route      # See container's routing
iptables -L   # See container's firewall rules

# OR: use ip netns with a symlink trick
# Docker doesn't create named namespaces in /var/run/netns/
# But you can link them:
PID=$(docker inspect <container> --format '{{.State.Pid}}')
sudo ln -s /proc/$PID/ns/net /var/run/netns/mycontainer
sudo ip netns exec mycontainer ip addr
```

---

## Practical Lab: Build Container Networking From Scratch

### What we'll build

```
         ┌──────────────────────────────────────────┐
         │  HOST                                    │
         │                                          │
         │  eth0: attached to internet              │
         │                                          │
         │  ┌────────────────────────┐              │
         │  │  Bridge: br0           │              │
         │  │  10.100.0.1/24         │              │
         │  └──┬──────────────┬──────┘              │
         │     │              │                     │
         │     │veth-c1-br    │veth-c2-br           │
         │     ║              ║                     │
         │     ║veth-c1       ║veth-c2              │
         │  ┌──╨────────┐ ┌──╨────────┐            │
         │  │ NS: c1    │ │ NS: c2    │            │
         │  │ 10.100.0.2│ │ 10.100.0.3│            │
         │  └───────────┘ └───────────┘            │
         └──────────────────────────────────────────┘

This is basically what Docker bridge networking does!
```

### Build it

```bash
#!/bin/bash
# build-container-network.sh — Build Docker-style networking from scratch

set -e

# --- Create bridge ---
sudo ip link add br0 type bridge
sudo ip addr add 10.100.0.1/24 dev br0
sudo ip link set br0 up

# --- Container 1 ---
sudo ip netns add c1
sudo ip link add veth-c1 type veth peer name veth-c1-br
sudo ip link set veth-c1 netns c1
sudo ip link set veth-c1-br master br0
sudo ip link set veth-c1-br up
sudo ip netns exec c1 ip addr add 10.100.0.2/24 dev veth-c1
sudo ip netns exec c1 ip link set veth-c1 up
sudo ip netns exec c1 ip link set lo up
sudo ip netns exec c1 ip route add default via 10.100.0.1

# --- Container 2 ---
sudo ip netns add c2
sudo ip link add veth-c2 type veth peer name veth-c2-br
sudo ip link set veth-c2 netns c2
sudo ip link set veth-c2-br master br0
sudo ip link set veth-c2-br up
sudo ip netns exec c2 ip addr add 10.100.0.3/24 dev veth-c2
sudo ip netns exec c2 ip link set veth-c2 up
sudo ip netns exec c2 ip link set lo up
sudo ip netns exec c2 ip route add default via 10.100.0.1

# --- Enable forwarding + NAT ---
sudo sysctl -w net.ipv4.ip_forward=1
sudo iptables -t nat -A POSTROUTING -s 10.100.0.0/24 ! -o br0 -j MASQUERADE

echo "Container network ready!"
echo "Test: sudo ip netns exec c1 ping 10.100.0.3"
echo "Test: sudo ip netns exec c1 ping 8.8.8.8"
```

### Test it

```bash
# Container 1 → Container 2 (via bridge)
sudo ip netns exec c1 ping -c 2 10.100.0.3
# ← Works! (through br0)

# Container 1 → Internet (via NAT)
sudo ip netns exec c1 ping -c 2 8.8.8.8
# ← Works! (NAT through host)

# Container 2 → Container 1
sudo ip netns exec c2 ping -c 2 10.100.0.2
# ← Works!

# Run a web server in c1
sudo ip netns exec c1 python3 -m http.server 80 &

# Access from c2
sudo ip netns exec c2 curl http://10.100.0.2
# ← Works!

# Access from host
curl http://10.100.0.2
# ← Works (host has route via br0)
```

### Clean up

```bash
#!/bin/bash
# cleanup.sh
sudo ip netns del c1
sudo ip netns del c2
sudo ip link del br0
# Deleting a namespace also deletes its veth ends
# Deleting bridge removes it from the datapath
sudo iptables -t nat -D POSTROUTING -s 10.100.0.0/24 ! -o br0 -j MASQUERADE
```

---

## How Docker Uses Namespaces

```
When you run: docker run -p 8080:80 nginx

Docker does essentially:
  1. Create a network namespace (the container)
  2. Create a veth pair
  3. Put one end in container, other end on docker0 bridge
  4. Assign IP from docker0 subnet (172.17.0.x)
  5. Set default route to docker0 (172.17.0.1)
  6. Add iptables NAT for -p 8080:80:
     -A DOCKER -p tcp --dport 8080 -j DNAT --to 172.17.0.2:80

Result:
  ┌───────────────────────────────────────────────┐
  │  HOST                                         │
  │                                               │
  │  docker0 bridge: 172.17.0.1                   │
  │  ├── veth_abc123 ══ eth0 (container: nginx)   │
  │  │                   172.17.0.2               │
  │  ├── veth_def456 ══ eth0 (container: redis)   │
  │  │                   172.17.0.3               │
  │                                               │
  │  iptables:                                    │
  │    DNAT :8080 → 172.17.0.2:80                 │
  │    MASQUERADE outbound from 172.17.0.0/16     │
  └───────────────────────────────────────────────┘

It's just namespaces + veth + bridge + iptables.
The same things we just built manually.
```

---

## Key Takeaways

1. **A network namespace is a complete, isolated copy of the network stack** — separate interfaces, IPs, routes, iptables, port space. Processes in different namespaces can bind the same port
2. **veth pairs are virtual cables** — whatever goes in one end comes out the other. The two ends can be in different namespaces, connecting them
3. **New namespaces start with nothing** — just a DOWN loopback. You must create interfaces, assign IPs, add routes
4. **IP forwarding + NAT gives namespaces internet access** — identical to how a home router works: enable forwarding, masquerade outbound traffic
5. **A Linux bridge connects multiple veth pairs** — like a virtual switch. This is how multiple containers talk to each other on the same host
6. **Docker networking IS namespace networking** — docker0 is a bridge, each container is a namespace, connected by veth pairs with iptables NAT
7. **`ip netns exec` runs commands in a namespace** — essential for debugging container networking at the Linux level
8. **`nsenter` lets you enter any process's namespace** — even Docker containers that don't have named namespaces in `/var/run/netns/`
9. **Each namespace has its own iptables** — firewall rules in one namespace don't affect another. This is how containers have independent security policies
10. **Understanding namespaces = understanding container networking** — every container orchestrator (Docker, K8s, Podman) builds on these same Linux primitives

---

## Next

→ [Bridges and Overlay Networks](./02-bridges-overlays.md) — How to connect namespaces across hosts
