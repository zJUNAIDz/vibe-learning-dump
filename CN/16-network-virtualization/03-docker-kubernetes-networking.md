# Docker and Kubernetes Networking

> Docker and Kubernetes don't invent new networking. They orchestrate the Linux primitives you already know: namespaces, veth pairs, bridges, iptables, and overlay networks. This file connects those building blocks to how containers and pods actually communicate in production.

---

## Table of Contents

1. [Docker Networking Model](#docker-model)
2. [Docker Bridge Network (default)](#docker-bridge)
3. [Docker Host Network](#docker-host)
4. [Docker None Network](#docker-none)
5. [Docker Overlay Network (Swarm)](#docker-overlay)
6. [Docker DNS and Service Discovery](#docker-dns)
7. [Kubernetes Networking Model](#k8s-model)
8. [Pod Networking — The Pause Container](#pause)
9. [CNI — Container Network Interface](#cni)
10. [Service Networking and kube-proxy](#services)
11. [DNS in Kubernetes](#k8s-dns)
12. [Network Policies](#network-policies)
13. [Debugging Container Networking](#debugging)
14. [Key Takeaways](#key-takeaways)

---

## Docker Networking Model

```
Docker provides 4 network drivers:

  ┌──────────────────────────────────────────────────────┐
  │  Driver     Description                              │
  │  ──────     ───────────                              │
  │  bridge     Default. Containers on a virtual bridge. │
  │  host       Container uses host's network stack.     │
  │  none       No networking at all.                    │
  │  overlay    Multi-host via VXLAN (Docker Swarm).     │
  └──────────────────────────────────────────────────────┘

Most applications use bridge or host mode.
```

---

## Docker Bridge Network (default)

### What happens when you run a container

```bash
docker run -d --name nginx -p 8080:80 nginx
```

```
Docker creates:

  1. Network namespace for the container
  2. veth pair: one end in container (eth0), one on docker0 bridge
  3. IP assigned from docker0 subnet (172.17.0.0/16)
  4. Default route inside container points to docker0 (172.17.0.1)
  5. iptables DNAT rule for -p 8080:80

  ┌──────────────────────────────────────────────────────┐
  │  HOST                                                │
  │                                                      │
  │  eth0: 192.168.1.100 (host IP)                       │
  │    │                                                 │
  │  iptables:                                           │
  │    PREROUTING -p tcp --dport 8080                    │
  │      -j DNAT --to-destination 172.17.0.2:80          │
  │    POSTROUTING -s 172.17.0.0/16 -j MASQUERADE       │
  │                                                      │
  │  docker0 bridge: 172.17.0.1/16                       │
  │  ├── veth_abc ══════ eth0 (nginx container)          │
  │  │                   172.17.0.2/16                   │
  │  │                   default via 172.17.0.1          │
  │  ├── veth_def ══════ eth0 (redis container)          │
  │  │                   172.17.0.3/16                   │
  └──────────────────────────────────────────────────────┘
```

### Verify it yourself

```bash
# See the docker0 bridge
ip link show docker0
ip addr show docker0
# 172.17.0.1/16

# See veth pairs
ip link show type veth
# veth1234567@if8: ... master docker0

# See the iptables rules Docker creates
sudo iptables -t nat -L -n -v
# Chain DOCKER
#   DNAT tcp dpt:8080 to:172.17.0.2:80

# Enter the container's network namespace
PID=$(docker inspect nginx --format '{{.State.Pid}}')
sudo nsenter --target $PID --net ip addr
# 1: lo: <LOOPBACK,UP> ...
# 8: eth0@if9: <BROADCAST,MULTICAST,UP> ... inet 172.17.0.2/16

sudo nsenter --target $PID --net ip route
# default via 172.17.0.1 dev eth0
# 172.17.0.0/16 dev eth0 proto kernel scope link src 172.17.0.2
```

### Container-to-container communication (same host)

```
Container A (172.17.0.2) → Container B (172.17.0.3):

  1. A sends packet → eth0 (172.17.0.3) 
  2. Packet traverses veth to docker0 bridge
  3. Bridge looks up MAC for 172.17.0.3
  4. Bridge forwards through B's veth
  5. Arrives at B's eth0

  → Direct Layer 2 switching through the bridge.
  → No NAT, no iptables (same subnet).
  → But containers must know each other's IP.
```

### Custom bridge networks

```bash
# Create a custom bridge network
docker network create --driver bridge mynet

# Docker creates a NEW bridge (br-xxxx), separate from docker0
ip link show type bridge
# docker0
# br-a1b2c3d4e5

# Run containers on custom network
docker run -d --name web --network mynet nginx
docker run -d --name api --network mynet node-app

# Key difference: custom networks get DNS!
docker exec api ping web
# PING web (172.18.0.2): ...
# ← Name "web" resolves to container IP!

# Containers on different networks can't communicate by default
# (they're on different bridges with different subnets)
```

---

## Docker Host Network

```bash
docker run -d --network host nginx
```

```
Container uses the HOST's network namespace directly:
  - No network isolation
  - No NAT or port mapping needed
  - Container sees host's eth0, host's IP tables
  - Performance: identical to running directly on host

  ┌──────────────────────────────────────────┐
  │  HOST                                    │
  │                                          │
  │  eth0: 192.168.1.100                     │
  │  ← nginx binds to port 80 on THIS eth0  │
  │                                          │
  │  No bridge, no veth, no NAT              │
  └──────────────────────────────────────────┘

When to use:
  - Maximum network performance (no veth overhead)
  - When container needs to see real host network
  - Port 80 in container = port 80 on host (no -p mapping)
  
Trade-off:
  - No port isolation (can't run two containers on port 80)
  - Container can see/modify host's network config
```

---

## Docker None Network

```bash
docker run -d --network none busybox sleep 3600
```

```
Container gets NO network at all:
  - Only loopback interface (lo)
  - Cannot communicate with anything
  - Complete network isolation

Use cases:
  - Security-sensitive workloads (no network attack surface)
  - Batch processing that doesn't need network
  - Manual networking setup (add your own interfaces later)
```

---

## Docker Overlay Network (Swarm)

```
For multi-host container communication:

  Host 1                              Host 2
  ┌────────────────────────┐          ┌────────────────────────┐
  │  Container A           │          │  Container B           │
  │  10.0.0.2              │          │  10.0.0.3              │
  │     │                  │          │     │                  │
  │  ┌──┴────┐             │          │  ┌──┴────┐             │
  │  │br-ovl │             │          │  │br-ovl │             │
  │  └──┬────┘             │          │  └──┬────┘             │
  │  ┌──┴──────┐           │          │  ┌──┴──────┐           │
  │  │ vxlan0  │           │          │  │ vxlan0  │           │
  │  │ VNI=256 │           │          │  │ VNI=256 │           │
  │  └──┬──────┘           │          │  └──┬──────┘           │
  │  eth0: 192.168.1.10    │          │  eth0: 192.168.1.20    │
  └────────┬───────────────┘          └────────┬───────────────┘
           │                                   │
           └──── Physical Network ─────────────┘

Docker Swarm uses VXLAN with a built-in control plane:
  - Gossip protocol spreads MAC→VTEP mappings
  - No multicast needed
  - Encrypted option available (IPsec)

Create: docker network create --driver overlay --attachable myoverlay
```

---

## Docker DNS and Service Discovery

```
Docker's embedded DNS server (127.0.0.11):

  ┌─────────────┐        ┌──────────────┐
  │ Container A │──DNS──>│ Docker DNS   │
  │             │        │ 127.0.0.11   │
  │ nslookup B  │        │              │
  │ → 172.18.0.3│<───────│ B=172.18.0.3 │
  └─────────────┘        └──────────────┘

Rules:
  - Only works on CUSTOM bridge networks (not default docker0!)
  - Container name = DNS name
  - Service name = DNS name with round-robin for replicas:
    docker service create --name api --replicas 3 node-app
    nslookup api → returns IPs of all 3 replicas

Container's /etc/resolv.conf:
  nameserver 127.0.0.11
  options ndots:0

Docker intercepts DNS queries on 127.0.0.11:
  - Container name → resolve to container IP
  - Unknown names → forward to host's DNS
```

---

## Kubernetes Networking Model

### The four requirements

```
Kubernetes networking has 4 fundamental rules:

  1. Every pod gets its own IP address
  2. Pods on the same node can communicate without NAT
  3. Pods on different nodes can communicate without NAT
  4. The IP a pod sees for itself is the same IP others see
  
  → Flat network: any pod can reach any pod by IP.
  → No NAT between pods (except for Services).
  → This is NOT how Docker bridge works (Docker uses NAT).
```

### Kubernetes network topology

```
  ┌──────────────────────────────────────────────────────┐
  │  Cluster                                             │
  │                                                      │
  │  Node 1 (10.0.1.0/24)           Node 2 (10.0.2.0/24)│
  │  ┌────────────────────┐         ┌────────────────────┐
  │  │ Pod A: 10.0.1.5    │         │ Pod C: 10.0.2.8    │
  │  │ Pod B: 10.0.1.12   │         │ Pod D: 10.0.2.15   │
  │  │                    │         │                    │
  │  │ cbr0 bridge        │         │ cbr0 bridge        │
  │  └────────┬───────────┘         └────────┬───────────┘
  │           │                              │
  │           └──── CNI (VXLAN/BGP/...) ─────┘
  │                                                      │
  │  Each node gets a unique pod subnet.                 │
  │  CNI plugin ensures cross-node reachability.         │
  └──────────────────────────────────────────────────────┘

  Pod A (10.0.1.5) → Pod C (10.0.2.8):
    Works directly! CNI handles the routing/tunneling.
    Pod A sees destination as 10.0.2.8 (no NAT).
    Pod C sees source as 10.0.1.5 (no NAT).
```

---

## Pod Networking — The Pause Container

### What the pause container does

```
Every pod has a hidden "pause" container:

  Pod "my-app":
  ┌──────────────────────────────────────────┐
  │                                          │
  │  ┌───────────┐  ┌───────────┐            │
  │  │ pause     │  │ my-app    │            │
  │  │ (infra)   │  │ container │            │
  │  │           │  │           │            │
  │  │ Creates & │  │ Joins     │            │
  │  │ HOLDS the │  │ pause's   │            │
  │  │ network   │  │ network   │            │
  │  │ namespace │  │ namespace │            │
  │  └───────────┘  └───────────┘            │
  │                                          │
  │  Network namespace (shared by all        │
  │  containers in the pod):                 │
  │    eth0: 10.0.1.5/24                     │
  │    lo: 127.0.0.1                         │
  │                                          │
  └──────────────────────────────────────────┘

Why pause container?
  1. Creates the network namespace (and keeps it alive)
  2. App containers join this namespace
  3. If app container crashes and restarts, network stays
  4. Multiple containers in same pod share localhost

Result:
  - Container A and Container B in same pod share an IP
  - They can talk via localhost (127.0.0.1)
  - They share the port space (can't both use port 80)
  - To outside: the pod has ONE IP address
```

### Multi-container pod networking

```
Pod with nginx + app:
  ┌──────────────────────────────────────────┐
  │  Pod (10.0.1.5)                          │
  │                                          │
  │  ┌───────────┐  ┌───────────┐            │
  │  │  nginx    │  │  app      │            │
  │  │  :80      │  │  :3000    │            │
  │  └───────────┘  └───────────┘            │
  │                                          │
  │  Both share eth0 (10.0.1.5)              │
  │  nginx reaches app at localhost:3000     │
  │  External reaches nginx at 10.0.1.5:80   │
  └──────────────────────────────────────────┘

This is why sidecars work:
  - Envoy sidecar proxy intercepts traffic on localhost
  - Log collector reads from shared volume
  - They're in the same network context
```

---

## CNI — Container Network Interface

### What CNI does

```
CNI = standard interface between Kubernetes and network plugins.

Kubernetes says: "I need networking for this pod."
CNI plugin says: "Done. Here's the IP."

Kubernetes does NOT implement networking.
The CNI plugin does all the work:
  1. Create veth pair
  2. Put one end in pod's namespace
  3. Connect other end to the node's network
  4. Assign IP address
  5. Set up routes
  6. Configure any overlay/tunnel needed
```

### Popular CNI plugins

```
┌───────────┬────────────────────────────────────────────────┐
│ Plugin    │ How it works                                   │
├───────────┼────────────────────────────────────────────────┤
│ Flannel   │ Simple overlay (VXLAN) or host-gw (routing)   │
│           │ Good for: simple clusters, getting started     │
├───────────┼────────────────────────────────────────────────┤
│ Calico    │ BGP routing (no overlay) or VXLAN/IPIP        │
│           │ Good for: performance, network policies        │
├───────────┼────────────────────────────────────────────────┤
│ Cilium    │ eBPF-based (bypasses iptables entirely)       │
│           │ Good for: performance, observability, security│
├───────────┼────────────────────────────────────────────────┤
│ Weave     │ VXLAN with mesh routing                       │
│           │ Good for: simplicity, encryption              │
├───────────┼────────────────────────────────────────────────┤
│ AWS VPC   │ Uses ENI (native AWS networking)              │
│           │ Good for: EKS, pod-to-AWS-service routing     │
└───────────┴────────────────────────────────────────────────┘

How Flannel VXLAN works:
  Node 1 (pod subnet: 10.244.0.0/24)
    → Flannel creates cni0 bridge + flannel.1 VXLAN interface
    → Pods get 10.244.0.x IPs
    → Cross-node traffic: encapsulate in VXLAN to destination node

How Calico BGP works:
  Node 1 (pod subnet: 10.244.0.0/24)
    → Calico announces 10.244.0.0/24 via BGP to all nodes
    → No overlay, no encapsulation
    → Physical routers know how to reach pod subnets
    → Best performance, but requires BGP-capable network
```

---

## Service Networking and kube-proxy

### The problem Services solve

```
Pods are ephemeral — IPs change constantly.
  Pod dies → new pod → new IP.
  Can't hardcode pod IPs.

Service: stable virtual IP (ClusterIP) that load balances to pods.

  ┌───────────────────────────────────────────────┐
  │  Service: my-api                              │
  │  ClusterIP: 10.96.0.50:80                     │
  │                                               │
  │  Endpoints:                                   │
  │  ├── 10.244.1.5:3000  (pod on Node 1)        │
  │  ├── 10.244.2.8:3000  (pod on Node 2)        │
  │  └── 10.244.1.12:3000 (pod on Node 1)        │
  └───────────────────────────────────────────────┘

  Any pod → 10.96.0.50:80
    → Intercepted by kube-proxy
    → DNAT to one of the endpoint pods
    → Random or round-robin selection
```

### How kube-proxy implements Services

```
Mode 1: iptables (default)

  kube-proxy programs iptables rules on every node:
  
  -A KUBE-SERVICES -d 10.96.0.50/32 -p tcp --dport 80 
    -j KUBE-SVC-XXXXX
  
  -A KUBE-SVC-XXXXX -m statistic --mode random --probability 0.333
    -j KUBE-SEP-AAAA    (→ DNAT to 10.244.1.5:3000)
  -A KUBE-SVC-XXXXX -m statistic --mode random --probability 0.500
    -j KUBE-SEP-BBBB    (→ DNAT to 10.244.2.8:3000)
  -A KUBE-SVC-XXXXX
    -j KUBE-SEP-CCCC    (→ DNAT to 10.244.1.12:3000)
  
  Probability math ensures equal distribution.
  Works but doesn't scale well (thousands of rules).

Mode 2: IPVS
  Uses Linux IPVS (IP Virtual Server) for load balancing.
  Better performance at scale (hash table vs. iptables chain).
  
Mode 3: eBPF (Cilium)
  Replaces kube-proxy entirely.
  eBPF programs handle service routing in kernel.
  Best performance, most features.
```

### NodePort and LoadBalancer

```
ClusterIP: Internal only (pods within cluster).

NodePort: Opens a port on EVERY node:
  Service type: NodePort
  Port 30080 on every node → forwards to Service → Pod
  
  External → Node1:30080 ─┐
  External → Node2:30080 ─┼→ Service → Pod
  External → Node3:30080 ─┘

LoadBalancer: Cloud provider creates external LB:
  External → Cloud LB:80 → NodePort:30080 → Service → Pod
  
  AWS: Creates an ELB/ALB/NLB
  GCP: Creates a Network/HTTP Load Balancer
  Azure: Creates an Azure Load Balancer
```

---

## DNS in Kubernetes

### CoreDNS

```
Kubernetes runs CoreDNS for service discovery:

  Every pod gets /etc/resolv.conf:
    nameserver 10.96.0.10  (CoreDNS ClusterIP)
    search default.svc.cluster.local svc.cluster.local cluster.local
    options ndots:5

DNS resolution:
  "my-api"
    → my-api.default.svc.cluster.local
    → 10.96.0.50 (Service ClusterIP)

  "my-api.other-namespace"
    → my-api.other-namespace.svc.cluster.local
    → 10.96.1.30

Pod-to-pod DNS (headless service):
  "my-pod.my-headless-svc.default.svc.cluster.local"
    → Direct pod IP (used by StatefulSets)
```

### ndots:5 — The performance gotcha

```
ndots:5 means: if a name has fewer than 5 dots, try search domains first.

  curl http://api.example.com
  
  DNS queries generated (in order):
  1. api.example.com.default.svc.cluster.local  ← NXDOMAIN
  2. api.example.com.svc.cluster.local          ← NXDOMAIN
  3. api.example.com.cluster.local              ← NXDOMAIN
  4. api.example.com.                           ← SUCCESS!

  4 DNS queries instead of 1!
  Multiply by every HTTP request = huge DNS load.

Fix: Use trailing dot for external names:
  curl http://api.example.com.   ← note the trailing dot
  → Only 1 DNS query!

Or set ndots:1 in pod spec (trade-off: breaks short service names).
```

---

## Network Policies

### Default: everything can talk to everything

```
Without network policies:
  
  Pod A → Pod B: ✓
  Pod A → Pod C: ✓
  Any pod → any pod: ✓  (flat network, no restrictions)

This is a security problem.
Network Policies add firewall rules at the pod level.
```

### Example: Restrict database access

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: db-allow-api-only
  namespace: default
spec:
  podSelector:
    matchLabels:
      app: database        # Apply to pods labeled app=database
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: api          # Only allow from pods labeled app=api
    ports:
    - protocol: TCP
      port: 5432            # Only PostgreSQL port
```

```
Result:
  api pod → database:5432   ✓ (allowed by policy)
  web pod → database:5432   ✗ (blocked!)
  api pod → database:6379   ✗ (wrong port!)

Important: Network policies are enforced by the CNI plugin.
  Calico: ✓       Cilium: ✓
  Flannel: ✗ (no network policy support!)
  WeaveNet: ✓
```

---

## Debugging Container Networking

### Docker debugging

```bash
# See all Docker networks
docker network ls

# Inspect a network (see containers, subnets)
docker network inspect bridge

# See container's IP and network settings
docker inspect <container> --format '{{json .NetworkSettings}}'

# Check iptables rules Docker creates
sudo iptables -t nat -L DOCKER -n -v

# Test from inside container
docker exec -it <container> sh
  > ip addr          # What IP do I have?
  > ip route         # Where is my default gateway?
  > cat /etc/resolv.conf   # What DNS am I using?
  > ping 172.17.0.1  # Can I reach the gateway?
  > nslookup other-container  # Does DNS work?

# Capture traffic on docker0 bridge
sudo tcpdump -i docker0 -nn

# Capture traffic on specific veth (to isolate one container)
sudo tcpdump -i veth1234abc -nn
```

### Kubernetes debugging

```bash
# Pod networking
kubectl exec -it <pod> -- sh
  > ip addr          # Pod IP and interfaces
  > ip route         # Default route (to node)
  > cat /etc/resolv.conf   # CoreDNS settings
  > nslookup kubernetes    # Test cluster DNS
  > wget -qO- http://my-service   # Test service discovery

# Node-level debugging
# Enter pod's network namespace from the node
CONTAINER_ID=$(crictl ps | grep <pod-name> | head -1 | awk '{print $1}')
PID=$(crictl inspect $CONTAINER_ID | jq .info.pid)
nsenter --target $PID --net -- ip addr
nsenter --target $PID --net -- ss -tlnp

# Check CNI configuration
ls /etc/cni/net.d/
cat /etc/cni/net.d/10-flannel.conflist

# Check kube-proxy rules
sudo iptables -t nat -L KUBE-SERVICES -n | head -20

# Check IPVS if using IPVS mode
sudo ipvsadm -Ln

# See service endpoints
kubectl get endpoints <service-name>

# DNS debugging pod
kubectl run dnsutils --image=gcr.io/kubernetes-e2e-test-images/dnsutils \
  --restart=Never -- sleep 3600
kubectl exec -it dnsutils -- nslookup kubernetes
kubectl exec -it dnsutils -- nslookup my-service.my-namespace
```

### Common issues

```
1. Pod can't reach other pods
   → Check CNI is running: kubectl get pods -n kube-system
   → Check node routes: ip route (should see pod subnet routes)
   → Check VXLAN: ip -d link show flannel.1
   
2. Pod can't reach Service
   → Check service exists: kubectl get svc
   → Check endpoints: kubectl get endpoints <svc> (should have pod IPs)
   → Check kube-proxy: iptables -t nat -L KUBE-SERVICES | grep <svc-ip>
   → Check CoreDNS: kubectl logs -n kube-system <coredns-pod>

3. External traffic can't reach NodePort
   → Check firewall on node: iptables -L INPUT
   → Check NodePort is open: ss -tlnp | grep <nodeport>
   → Check cloud security group allows the port

4. DNS resolution slow
   → Check ndots setting (probably 5, causing extra queries)
   → Add trailing dot to external domains
   → Check CoreDNS pods are healthy
```

---

## Key Takeaways

1. **Docker bridge networking = namespace + veth + bridge + iptables NAT** — understanding the primitives lets you debug any Docker networking issue
2. **Docker host networking removes all isolation** — container shares host's network stack, best performance but no port isolation
3. **Custom Docker bridge networks get DNS resolution** — container names resolve to IPs. Default docker0 bridge does NOT get this
4. **Kubernetes requires a flat pod network: any pod can reach any pod without NAT** — this is fundamentally different from Docker's default NAT-based model
5. **The pause container creates and holds the pod's network namespace** — app containers join it, so they share an IP and can communicate via localhost
6. **CNI plugins implement the actual networking** — Kubernetes doesn't do networking itself, it delegates to Flannel/Calico/Cilium
7. **kube-proxy turns Service ClusterIPs into iptables DNAT rules** — every node can route to any service. eBPF (Cilium) is the modern replacement
8. **ndots:5 causes 4 extra DNS queries for external domains** — use trailing dots or reduce ndots for external service calls
9. **Network Policies default to allow-all** — you must explicitly create policies to restrict traffic. Requires a CNI that supports them (not Flannel!)
10. **Debug with nsenter, not just kubectl exec** — nsenter into the pod's namespace from the node gives you full Linux networking tools

---

## Next Module

→ [Module 17: Systematic Network Debugging](../17-debugging-methodology/01-systematic-network-debugging.md) — A structured approach to diagnosing any network problem
