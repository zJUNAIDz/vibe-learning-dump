# Static vs Dynamic Routing — When to Use What

> Static routing is simple but fragile. Dynamic routing is complex but resilient. Real networks use both.

---

## Table of Contents

1. [Static Routing](#static-routing)
2. [Why Static Isn't Enough](#why-static-isnt-enough)
3. [Dynamic Routing Fundamentals](#dynamic-routing-fundamentals)
4. [Distance Vector vs Link State](#distance-vector-vs-link-state)
5. [When to Use Static vs Dynamic](#when-to-use-static-vs-dynamic)
6. [Linux Static Routing Lab](#linux-static-routing-lab)

---

## Static Routing

A static route is manually configured by an administrator:

```bash
sudo ip route add 10.0.0.0/8 via 192.168.1.1
```

### Advantages

1. **Simple**: No routing protocol to configure, no CPU overhead
2. **Predictable**: Traffic always takes the exact path you specified
3. **Secure**: No routing protocol messages to intercept or spoof
4. **Zero overhead**: No bandwidth consumed by routing updates

### Disadvantages

1. **Doesn't adapt**: If a link goes down, the static route still sends traffic down the dead link until you manually change it
2. **Doesn't scale**: 100 routers × 50 subnets = thousands of manual entries
3. **Error-prone**: Typos cause routing black holes
4. **No load balancing intelligence**: Can't shift traffic based on congestion

### When static routes make sense

- **Stub networks** (only one exit): Your home network has one router to the ISP. There's only one possible path.
- **Default routes**: Even in dynamic routing environments, the default route is often static
- **Policy overrides**: "Always send traffic to X through this specific path"
- **Testing/debugging**: Temporarily force traffic down a specific path

---

## Why Static Isn't Enough

### The convergence problem

```
Before failure:
  A ──── B ──── C
  
  Static routes on A: 10.0.0.0/8 via B
  Traffic flows: A → B → C ✓

Link B-C fails:
  A ──── B    ✗  C
  
  A still sends to B (static route unchanged)
  B can't reach C → packets dropped
  NETWORK IS DOWN until admin manually reconfigures
```

With dynamic routing:
1. B detects C is unreachable (link down notification)
2. B tells A: "I can't reach C anymore"
3. If an alternate path exists (A → D → C), routers converge on it automatically
4. **Convergence time**: Seconds (OSPF/ISIS) to minutes (BGP)

### Scale problem

| Network size | Static routes needed | Practical? |
|-------------|---------------------|-----------|
| 3 routers | ~10 | Yes |
| 10 routers | ~50 | Manageable |
| 100 routers | ~5,000+ | No |
| Internet (~80,000 ASes) | ~1,000,000 | Absolutely not |

---

## Dynamic Routing Fundamentals

Dynamic routing protocols automate two tasks:
1. **Discovery**: Learning about networks reachable through neighboring routers
2. **Selection**: Choosing the best path when multiple routes exist

### How dynamic routing works (conceptually)

1. Routers exchange information with their neighbors
2. Each router builds a picture of the network topology
3. Each router computes the best path to every known destination
4. The results are installed in the routing table
5. When the topology changes (link up/down), steps 1-4 repeat

### Key metrics for choosing routes

| Metric | What it measures | Used by |
|--------|-----------------|---------|
| Hop count | Number of routers to cross | RIP |
| Cost/Bandwidth | Link speed (inversely proportional) | OSPF |
| Delay | Measured latency | EIGRP |
| AS path length | Number of autonomous systems | BGP |
| Composite | Multiple factors weighted | EIGRP, BGP |

---

## Distance Vector vs Link State

Two fundamental approaches to distributed routing:

### Distance Vector (e.g., RIP)

**Concept**: Each router tells its neighbors: "Here are the networks I can reach and how far they are."

**Algorithm**: Bellman-Ford distributed

```
Router A's table:
  Network     Distance    Via
  10.0.1.0/24    0        directly connected
  10.0.2.0/24    1        Router B
  10.0.3.0/24    2        Router B  (B reaches it in 1 hop)
```

**How it works**:
1. Each router periodically broadcasts its entire routing table to neighbors
2. When a router receives a neighbor's table, it checks: "Can I reach any network cheaper through this neighbor?"
3. If yes → update my table with the new route (distance = neighbor's distance + 1)

**Problems**:
- **Slow convergence**: Changes propagate hop-by-hop. Large networks can take minutes.
- **Count-to-infinity**: If a network becomes unreachable, routers keep passing increasingly wrong distances until they hit a maximum (typically 16).
- **Routing loops during convergence**: Before all routers agree on the new topology, packets can loop.

**Mitigation techniques**: Split horizon, route poisoning, hold-down timers — all workarounds for fundamental design limitations.

**RIP** (Routing Information Protocol) is the classic distance-vector protocol. Max hop count: 15 (16 = unreachable). It's obsolete for serious use.

### Link State (e.g., OSPF, IS-IS)

**Concept**: Each router tells ALL routers: "Here are the links I have and their costs."

**Algorithm**: Each router runs Dijkstra's shortest path algorithm independently.

```
Router A announces: "I have links to B (cost 10) and C (cost 5)"
Router B announces: "I have links to A (cost 10) and D (cost 20)"
Router C announces: "I have links to A (cost 5) and D (cost 15)"
Router D announces: "I have links to B (cost 20) and C (cost 15)"

Every router now has the COMPLETE MAP. Each runs Dijkstra independently.
```

**How it works**:
1. Each router discovers its neighbors (Hello protocol)
2. Each router creates a Link State Advertisement (LSA) listing its links and costs
3. LSAs are flooded to ALL routers in the network (reliable flooding)
4. Each router has an identical Link State Database (LSDB) — the complete topology
5. Each router runs Dijkstra's algorithm to compute shortest paths
6. Results are installed in the routing table

**Advantages over distance vector**:
- **Fast convergence**: All routers have the complete topology, so they can compute new paths immediately when a link changes
- **No count-to-infinity**: Routers have complete information, not just neighbor distances
- **Loop-free**: Dijkstra guarantees loop-free shortest paths

**Disadvantages**:
- **More complex**: LSA flooding, database synchronization, Dijkstra computation
- **More memory**: Each router stores the entire topology (not just routes)
- **More CPU**: Running Dijkstra on large topologies is compute-intensive

### Comparison

| Property | Distance Vector | Link State |
|----------|---------------|-----------|
| Knowledge | Partial (neighbor distances) | Complete (full topology) |
| Algorithm | Bellman-Ford | Dijkstra |
| Updates sent | Entire routing table (periodic) | Only changed links (event-triggered) |
| Convergence | Slow (minutes for RIP) | Fast (seconds for OSPF) |
| Scalability | Poor | Good (with areas/hierarchy) |
| Memory | Low | High (stores topology) |
| CPU | Low | Higher (Dijkstra computation) |

---

## When to Use Static vs Dynamic

| Scenario | Static | Dynamic | Why |
|----------|--------|---------|-----|
| Home network | ✓ | – | One path, one router |
| Small office (5 devices) | ✓ | – | Simple enough to manage manually |
| Enterprise campus | – | ✓ (OSPF) | Too many paths to manage statically |
| Between ISPs | – | ✓ (BGP) | Policy-based routing, thousands of routes |
| Data center | – | ✓ (BGP/OSPF) | Redundancy and scale |
| Default route | ✓ | – | Even in dynamic environments |
| Cloud VPC | ✓ | ✓ | Static for simple setups, dynamic for VPN/hybrid |

---

## Linux Static Routing Lab

### Setup: Two networks connected by a "router"

```bash
# Create three network namespaces
sudo ip netns add host-a
sudo ip netns add router
sudo ip netns add host-b

# Create veth pairs
sudo ip link add veth-a type veth peer name veth-ar
sudo ip link add veth-b type veth peer name veth-br

# Connect host-a to router
sudo ip link set veth-a netns host-a
sudo ip link set veth-ar netns router

# Connect host-b to router
sudo ip link set veth-b netns host-b
sudo ip link set veth-br netns router

# Configure IPs
sudo ip netns exec host-a ip addr add 10.0.1.10/24 dev veth-a
sudo ip netns exec router ip addr add 10.0.1.1/24 dev veth-ar
sudo ip netns exec router ip addr add 10.0.2.1/24 dev veth-br
sudo ip netns exec host-b ip addr add 10.0.2.10/24 dev veth-b

# Bring interfaces up
sudo ip netns exec host-a ip link set veth-a up
sudo ip netns exec host-a ip link set lo up
sudo ip netns exec router ip link set veth-ar up
sudo ip netns exec router ip link set veth-br up
sudo ip netns exec router ip link set lo up
sudo ip netns exec host-b ip link set veth-b up
sudo ip netns exec host-b ip link set lo up

# Enable forwarding on the router
sudo ip netns exec router sysctl -w net.ipv4.ip_forward=1

# Add default routes on hosts (point to router)
sudo ip netns exec host-a ip route add default via 10.0.1.1
sudo ip netns exec host-b ip route add default via 10.0.2.1

# Test
sudo ip netns exec host-a ping -c 3 10.0.2.10
# Should work! Traffic flows: host-a → router → host-b

# See routing tables
sudo ip netns exec host-a ip route show
sudo ip netns exec router ip route show
sudo ip netns exec host-b ip route show

# Cleanup
sudo ip netns del host-a
sudo ip netns del router
sudo ip netns del host-b
```

---

## Key Takeaways

1. **Static routing is simple but doesn't adapt** — good for stub networks and default routes
2. **Dynamic routing adapts to failures** — essential for anything beyond trivial networks
3. **Distance vector**: routers share distances with neighbors (simple, slow convergence)
4. **Link state**: routers share topology with everyone (complex, fast convergence)
5. **OSPF (link state)** dominates enterprise interior routing
6. **BGP (path vector)** dominates internet inter-domain routing
7. **Linux network namespaces** let you build multi-router labs on a single machine

---

## Next

→ [03-ospf-conceptual.md](03-ospf-conceptual.md) — How OSPF discovers and computes routes
