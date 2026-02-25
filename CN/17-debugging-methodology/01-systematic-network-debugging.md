# Systematic Network Debugging

> Most people debug networks by randomly trying things. Experienced engineers follow a systematic method: start at the bottom of the stack, verify each layer, and isolate the problem. This file gives you that method — a repeatable process for diagnosing any network issue, from "the internet is down" to "why is this API call taking 3 seconds?"

---

## Table of Contents

1. [The Debugging Mindset](#mindset)
2. [The Layer-by-Layer Method](#layer-method)
3. [Layer 1: Physical / Link](#layer-1)
4. [Layer 2: Data Link / MAC](#layer-2)
5. [Layer 3: Network / IP](#layer-3)
6. [Layer 4: Transport / TCP-UDP](#layer-4)
7. [Layer 7: Application](#layer-7)
8. [DNS — The Special Case](#dns)
9. [The Halving Method](#halving)
10. [Performance Debugging](#performance)
11. [Intermittent Issues](#intermittent)
12. [Container / Kubernetes Debugging](#containers)
13. [The Debugging Toolkit](#toolkit)
14. [Debugging Flowcharts](#flowcharts)
15. [Key Takeaways](#key-takeaways)

---

## The Debugging Mindset

### Rules for network debugging

```
RULE 1: Reproduce first.
  Can you reproduce the problem?
  If yes → you can test fixes.
  If no → gather data for when it happens again.

RULE 2: Change ONE thing at a time.
  Changed 3 things and it works? You don't know which fixed it.
  Changed 1 thing and it works? You know the cause.

RULE 3: Verify each layer bottom-up.
  Don't assume Layer 3 works. Prove it.
  Don't assume DNS works. Test it.

RULE 4: Isolate the problem.
  Is it the client, the network, or the server?
  Test from different locations.
  Capture on both sides.

RULE 5: Check what changed.
  "It was working yesterday" → what changed since yesterday?
  Deployments, config changes, infrastructure updates, certificates.

RULE 6: Trust data, not assumptions.
  "The firewall should allow it" → verify with tcpdump.
  "DNS should resolve" → test with dig.
  "The route should be there" → check with ip route.
```

### The first 60 seconds

```
When someone reports "the network is broken":

  1. WHAT exactly isn't working? (be specific)
     "Can't connect to production API"
     Not: "the internet is down"

  2. WHEN did it start?
     "Started 20 minutes ago"
     → Check: what changed 20 minutes ago?

  3. WHO is affected?
     "Just me" → probably client-side
     "My whole team" → probably service-side or DNS
     "Everyone in the company" → probably infrastructure

  4. WHAT has changed?
     Deployments? DNS changes? Certificate renewals?
     Firewall rule updates? Cloud config changes?

  5. QUICK TESTS (do these immediately):
     ping <target>          → basic connectivity
     curl -v <url>          → HTTP-level details
     dig <domain>           → DNS resolution
     traceroute <target>    → path analysis
```

---

## The Layer-by-Layer Method

```
Start at the bottom. Work your way up.

  ┌─────────────────────────────────────────┐
  │ Layer 7: Application                     │  curl, browser  
  ├─────────────────────────────────────────┤
  │ Layer 4: Transport (TCP/UDP)             │  ss, netstat, tcpdump
  ├─────────────────────────────────────────┤
  │ Layer 3: Network (IP)                    │  ping, traceroute, ip route
  ├─────────────────────────────────────────┤
  │ Layer 2: Data Link (Ethernet)            │  ip link, arp, bridge
  ├─────────────────────────────────────────┤
  │ Layer 1: Physical                        │  cables, LEDs, ethtool
  └─────────────────────────────────────────┘
                ↑ START HERE

If Layer 3 doesn't work, no point checking Layer 7.
Fix the lowest broken layer first.
```

---

## Layer 1: Physical / Link

### Checks

```bash
# Is the interface UP?
ip link show eth0
# Look for: <...UP...> (not DOWN)
# Look for: state UP (not state DOWN)

# Is there a cable connected? (physical servers)
ethtool eth0
# Link detected: yes   ← cable connected
# Link detected: no    ← no cable or bad cable
# Speed: 1000Mb/s      ← negotiated speed
# Duplex: Full

# Check for errors
ethtool -S eth0 | grep -i error
# rx_errors: 0         ← should be 0
# tx_errors: 0
# rx_crc_errors: 25    ← BAD cable or NIC issue!

# Check for drops
ethtool -S eth0 | grep -i drop
# rx_dropped: 0        ← should be near 0
# rx_missed_errors: 15 ← ring buffer overflow (need tuning)

# Check interface statistics
ip -s link show eth0
# RX: bytes packets errors dropped
# TX: bytes packets errors dropped

# For VMs / cloud: Is the interface attached?
# AWS: Check ENI attachment in console
# GCP: Check network interface in console
```

### Common Layer 1 problems

```
Problem                          Fix
───────                          ───
Interface DOWN                   ip link set eth0 up
No link detected                 Check cable, switch port
CRC errors                       Replace cable
Lots of RX drops                 Increase ring buffer (ethtool -G)
Speed negotiation wrong          Force speed (ethtool -s eth0 speed 1000)
MTU mismatch                     Set consistent MTU (ip link set mtu)
```

---

## Layer 2: Data Link / MAC

### Checks

```bash
# Does the interface have a MAC address?
ip link show eth0
# link/ether 00:1a:2b:3c:4d:5e

# Can we resolve neighbor MAC addresses?
ip neigh show
# 10.0.0.1 dev eth0 lladdr 00:11:22:33:44:55 REACHABLE  ← good
# 10.0.0.1 dev eth0             INCOMPLETE                ← bad (ARP failing)
# 10.0.0.1 dev eth0 lladdr 00:11:22:33:44:55 STALE       ← ok (will re-ARP)

# ARP request for gateway
arping -I eth0 10.0.0.1
# If no reply → gateway down, VLAN mismatch, or wrong subnet

# Check for duplicate IPs (ARP conflicts)
arping -D -I eth0 10.0.0.5
# If reply → someone else has this IP!

# Check bridge FDB (if using bridges)
bridge fdb show

# VLAN check
cat /proc/net/vlan/config
```

### Common Layer 2 problems

```
Problem                          Diagnosis
───────                          ─────────
ARP INCOMPLETE                   Gateway down, wrong VLAN, or firewall blocking ARP
Duplicate IP (ARP conflict)      Two devices with same IP on subnet
Wrong VLAN                       Interface on wrong VLAN, no trunk to switch
MAC flapping on switch           Loop or VM migration
No ARP replies from gateway      Gateway down, VLAN mismatch, or cable on wrong port
```

---

## Layer 3: Network / IP

### Checks

```bash
# Does the interface have an IP?
ip addr show eth0
# inet 10.0.0.5/24 ← has IP

# Can I reach the default gateway?
ip route show default
# default via 10.0.0.1 dev eth0
ping -c 3 10.0.0.1
# ← If this fails, don't bother testing anything beyond

# Can I reach a known IP? (skip DNS)
ping -c 3 8.8.8.8
# ← Tests connectivity beyond the gateway

# Trace the path
traceroute -n 8.8.8.8
# Look for: where does it stop? Where do asterisks start?
# 1  10.0.0.1 (gateway)
# 2  * * *  ← problem is at hop 2
# 3  * * *

# Check routing table
ip route show
# Is there a route for the destination?
ip route get 10.5.0.100
# 10.5.0.100 via 10.0.0.1 dev eth0 src 10.0.0.5
# ← Shows exactly how Linux would route this packet

# Check for asymmetric routing
traceroute -n <dest>       # outbound path
# On remote: traceroute -n <my-ip>  # return path
# If different → firewall may drop (conntrack expects symmetric)
```

### Common Layer 3 problems

```
Problem                          Diagnosis / Fix
───────                          ───────────────
No default route                 ip route add default via <gateway>
Wrong subnet mask                ip addr check (e.g., /32 instead of /24)
Can ping gateway, not beyond     Gateway not routing/NAT, or firewall
Asymmetric routing               Return path different → conntrack drops
Black hole route                 Route exists but next hop is unreachable
Source IP wrong                  Wrong interface selected, check src in ip route get
ICMP blocked                     ping fails but TCP works → ICMP filtered (not broken)
```

---

## Layer 4: Transport / TCP-UDP

### Checks

```bash
# Is the port open? (TCP)
# From client:
nc -zv <server> <port>
# Connection to 10.0.0.5 443 port [tcp/https] succeeded!
# or: Connection refused → port not listening
# or: timeout → firewall dropping SYN

# From server: Is something listening?
ss -tlnp | grep :<port>
# LISTEN  0  128  *:443  *:*  users:(("nginx",pid=1234,fd=6))
# ← nginx is listening on 443

# Is the service bound to the right address?
ss -tlnp | grep 443
# LISTEN  0  128  127.0.0.1:443     ← only localhost!
# vs
# LISTEN  0  128  0.0.0.0:443       ← all interfaces ✓

# Check connection state
ss -tn | grep <ip>
# ESTAB     0      0      10.0.0.5:443    10.0.0.100:52394
# ← Connection established

# Check for connection errors
ss -tn | grep -c ESTAB    # established connections
ss -tn | grep -c TIME-WAIT  # closing connections
ss -tn | grep -c CLOSE-WAIT # stuck connections (application bug!)
ss -tn | grep -c SYN-SENT   # outbound connections waiting (firewall?)

# Detailed TCP info
ss -tnei | grep -A 2 <ip>
# Shows: retransmits, RTT, send/recv buffer, congestion window

# Capture and analyze
sudo tcpdump -i eth0 host <target> -nn
# Look for: SYN without SYN-ACK → connection problem
# Look for: retransmissions → packet loss
# Look for: RST → connection refused/killed
```

### Common Layer 4 problems

```
Problem                          Evidence
───────                          ────────
Connection refused               RST in response to SYN → nothing listening
Connection timeout               SYN sent, no response → firewall
Connection reset                 RST after ESTABLISHED → app crash, firewall
CLOSE_WAIT accumulation          Application not calling close() → app bug
TIME_WAIT exhaustion             Too many short-lived connections → port exhaustion
Slow connection                  Retransmissions → packet loss
Service bound to 127.0.0.1       Can connect locally but not remotely
SYN backlog full                 SYN cookies activated (nstat | grep SyncookiesSent)
```

---

## Layer 7: Application

### Checks

```bash
# HTTP testing
curl -v http://server:port/endpoint
# Look for:
#   TCP connection time
#   TLS handshake time
#   HTTP status code
#   Response headers

# HTTPS certificate check
curl -vvI https://example.com 2>&1 | grep -E 'SSL|expire|issuer|subject'
# Or:
openssl s_client -connect example.com:443 -servername example.com </dev/null 2>/dev/null \
  | openssl x509 -noout -dates -subject
# notBefore=...
# notAfter=...  ← is it expired?
# subject=CN=example.com

# HTTP timing breakdown
curl -o /dev/null -w "\
  DNS:        %{time_namelookup}s\n\
  Connect:    %{time_connect}s\n\
  TLS:        %{time_appconnect}s\n\
  TTFB:       %{time_starttransfer}s\n\
  Total:      %{time_total}s\n" \
  https://api.example.com/health

# Check TLS version and cipher
openssl s_client -connect example.com:443 </dev/null 2>/dev/null | grep -E 'Protocol|Cipher'
# Protocol  : TLSv1.3
# Cipher    : TLS_AES_256_GCM_SHA384

# Test specific HTTP method
curl -X POST -H "Content-Type: application/json" \
  -d '{"test": true}' \
  http://api.example.com/endpoint

# Check response body for error messages
curl -s http://api.example.com/health | jq .
```

### Common Layer 7 problems

```
Problem                          Evidence / Fix
───────                          ──────────────
Certificate expired              openssl shows notAfter in the past
Certificate name mismatch        CN/SAN doesn't match the hostname
HTTP 502 Bad Gateway             Reverse proxy can't reach backend
HTTP 504 Gateway Timeout         Backend too slow, proxy times out
HTTP 503 Service Unavailable     Backend overloaded, health check failing
TLS version mismatch             Client requires TLS 1.3, server only has 1.2
Connection works, no response    Application running but not processing requests
Slow response                    curl timing shows which phase is slow
```

---

## DNS — The Special Case

### DNS fails = everything fails

```bash
# Basic resolution test
dig example.com
# Look for: ANSWER section with IP addresses
# Look for: status: NOERROR (vs NXDOMAIN, SERVFAIL, REFUSED)

# Specific record types
dig A example.com        # IPv4
dig AAAA example.com     # IPv6
dig MX example.com       # Mail
dig CNAME api.example.com  # Alias

# Test against specific DNS server
dig @8.8.8.8 example.com
# If this works but default dig doesn't → local DNS issue

# Check what DNS server you're using
cat /etc/resolv.conf

# Trace the resolution path
dig +trace example.com
# Shows: root → TLD → authoritative → answer

# Check for DNS cache issues
# Flush systemd-resolved cache
sudo systemd-resolve --flush-caches

# Time DNS resolution
time dig example.com

# Check for DNSSEC issues
dig +dnssec example.com
```

### DNS debugging checklist

```
□ Can you resolve at all? → dig example.com
□ Is the DNS server reachable? → ping <dns-server-ip>
□ Does it work with a public DNS? → dig @8.8.8.8 example.com
□ Is it a specific record? → dig CNAME / dig A / dig AAAA
□ Is it cached (stale)? → dig example.com (check TTL)
□ Is /etc/resolv.conf correct? → cat /etc/resolv.conf
□ Is systemd-resolved working? → resolvectl status
□ In Kubernetes: is CoreDNS running? → kubectl get pods -n kube-system
```

---

## The Halving Method

### When you can't find the problem: split the problem in half

```
Client ──── Switch ──── Router ──── Firewall ──── Server

Can client reach the router? (middle of path)
  YES → problem is between router and server
  NO  → problem is between client and router

Now test the middle of the remaining half:
  Can client reach the switch?
  Can router reach the firewall?

Keep halving until you find the exact point of failure.

This is binary search applied to network debugging.
O(log n) instead of O(n) hops to check.
```

### Apply it to application debugging too

```
Client → DNS → TCP connect → TLS → HTTP request → Server process → Response

Which phase is slow?

  curl timing breakdown tells you:
    DNS: 3.015s       ← SLOW! DNS is the problem.
    Connect: 0.020s   
    TLS: 0.045s
    TTFB: 0.100s
    Total: 3.180s

  OR:
    DNS: 0.015s
    Connect: 0.020s
    TLS: 0.045s
    TTFB: 2.500s     ← SLOW! Server processing is the problem.
    Total: 2.580s
```

---

## Performance Debugging

### Bandwidth issues

```bash
# Quick bandwidth test (between two hosts you control)
# Server:
iperf3 -s
# Client:
iperf3 -c <server-ip>
# Result: [  4]  0.00-10.00 sec  1.10 GBytes  941 Mbits/sec

# UDP test (for jitter and loss)
iperf3 -c <server-ip> -u -b 100M
# Result: Shows jitter and packet loss percentage

# Check interface speed/duplex
ethtool eth0 | grep -E 'Speed|Duplex'
# Speed: 10000Mb/s
# Duplex: Full

# Check for interface saturation
sar -n DEV 1 5
# Shows bytes/s and packets/s per interface over time
# Compare to interface speed
```

### Latency issues

```bash
# Baseline latency
ping -c 100 <target>
# rtt min/avg/max/mdev = 0.5/0.8/3.2/0.4 ms
# Low mdev = consistent → good
# High mdev = jitter → possible congestion

# Detailed per-hop latency
mtr -n --report <target>
# Shows latency and loss at each hop

# TCP connection time (isolates network latency from app)
curl -o /dev/null -w "TCP connect: %{time_connect}s\n" http://<target>

# Compare TCP connect time to ping RTT:
# If similar → network latency is baseline
# If TCP >> ping → SYN queue delay (server overloaded)
```

### Packet loss detection

```bash
# ICMP loss
ping -c 1000 -i 0.01 <target>
# 1000 packets transmitted, 998 received, 0.2% packet loss

# Better: use mtr for per-hop loss
mtr -n --report -c 100 <target>
# Loss at specific hop → problem at that hop
# Loss at LAST hop only → could be ICMP rate limiting (not real loss)

# TCP retransmission rate (real loss indicator)
nstat -az | grep -i retrans
# TcpRetransSegs    150    → retransmitted segments
# TcpInSegs         100000 → total segments received
# Retrans rate: 150/100000 = 0.15% → some loss

# In tcpdump
tcpdump -r capture.pcap 'tcp[13] & 0x04 != 0' | wc -l  # RSTs
tshark -r capture.pcap -Y tcp.analysis.retransmission | wc -l
```

---

## Intermittent Issues

### How to catch problems that come and go

```bash
# Continuous ping with timestamps
ping <target> | while read line; do echo "$(date '+%H:%M:%S') $line"; done
# Save to file:
ping <target> | while read line; do echo "$(date '+%H:%M:%S') $line"; done | tee ping.log

# Monitor for connection failures
while true; do
  result=$(curl -s -o /dev/null -w "%{http_code} %{time_total}s" \
    http://api.example.com/health 2>&1)
  echo "$(date '+%H:%M:%S.%N') $result"
  sleep 1
done | tee monitor.log

# Capture only when things go wrong
# Run long capture, filter interesting events later:
sudo tcpdump -i eth0 -w /tmp/long-capture.pcap \
  host <target> -G 3600 -W 24 \
  # -G: rotate every 3600 seconds (1 hour)
  # -W: keep max 24 files

# Look for retransmissions in the capture
tshark -r /tmp/long-capture.pcap -Y tcp.analysis.retransmission \
  -T fields -e frame.time -e ip.src -e ip.dst -e tcp.analysis.retransmission

# Monitor for slow responses
while true; do
  t=$(curl -o /dev/null -s -w "%{time_total}" http://api.example.com/health)
  ts=$(date '+%H:%M:%S')
  if (( $(echo "$t > 1.0" | bc -l) )); then
    echo "SLOW: $ts ${t}s" >> /tmp/slow.log
  fi
  sleep 0.5
done
```

---

## Container / Kubernetes Debugging

### The layered approach for containers

```
Extra layers to check:

  ┌────────────────────────────────────────────┐
  │ App Layer: HTTP, gRPC, WebSocket           │
  ├────────────────────────────────────────────┤
  │ Service Layer: kube-proxy, iptables DNAT   │ ← K8s Services
  ├────────────────────────────────────────────┤
  │ Network Policy: CNI enforcement            │ ← Pod firewall
  ├────────────────────────────────────────────┤
  │ Pod Network: veth, bridge, CNI             │ ← Pod-to-pod
  ├────────────────────────────────────────────┤
  │ Overlay: VXLAN, IPIP, WireGuard            │ ← Cross-node
  ├────────────────────────────────────────────┤
  │ Node Network: eth0, routing                │ ← Host network
  └────────────────────────────────────────────┘
```

### Step-by-step container debugging

```bash
# 1. Can the pod reach its own service?
kubectl exec -it <pod> -- wget -qO- http://localhost:<port>/health

# 2. Can the pod resolve DNS?
kubectl exec -it <pod> -- nslookup kubernetes

# 3. Can the pod reach another pod directly (by IP)?
kubectl exec -it <pod> -- wget -qO- http://<other-pod-ip>:<port>/health

# 4. Can the pod reach a Service (by name)?
kubectl exec -it <pod> -- wget -qO- http://<service-name>:<port>/health

# 5. Can the pod reach external internet?
kubectl exec -it <pod> -- wget -qO- http://httpbin.org/ip

# If step 1 fails → app issue
# If step 2 fails → CoreDNS issue
# If step 3 fails → CNI / overlay issue
# If step 4 fails → kube-proxy / Service issue
# If step 5 fails → NAT / egress issue

# 6. Check from the node
# Get node of the pod
kubectl get pod <pod> -o wide
# SSH to that node
# Capture traffic:
sudo tcpdump -i any host <pod-ip> -nn

# Check CNI
kubectl get pods -n kube-system | grep -E 'calico|flannel|cilium|weave'
# Are CNI pods running on this node?

# Check iptables (kube-proxy rules)
sudo iptables -t nat -L KUBE-SERVICES -n | grep <svc-cluster-ip>
```

---

## The Debugging Toolkit

### Quick reference

```
Layer    Tool              What it tells you
─────    ────              ─────────────────
L1       ethtool           Link status, speed, errors
         ip link           Interface state, MTU
         
L2       ip neigh          ARP table (MAC resolution)
         arping            ARP-level reachability
         bridge fdb        Bridge forwarding table
         
L3       ping              Basic IP reachability
         traceroute/mtr    Path and per-hop latency
         ip route          Routing table
         ip route get      How a specific dest would be routed

L4       ss                Socket state (listening, established)
         nc                Test TCP/UDP connectivity
         tcpdump           Packet-level analysis
         nstat             TCP statistics (retransmits, errors)

DNS      dig               DNS resolution and diagnostics
         nslookup          Simple DNS lookup
         resolvectl        systemd-resolved status

L7       curl -v           HTTP request/response details
         curl -w           Timing breakdown
         openssl s_client  TLS certificate and handshake
         
Perf     iperf3            Bandwidth test
         mtr               Continuous path analysis
         sar               Interface utilization over time
         
K8s      kubectl exec      Test from inside pods
         nsenter           Enter pod namespace from node
         crictl            Container runtime debugging
```

---

## Debugging Flowcharts

### "I can't connect to X"

```
Can't connect to X
│
├── Can you ping X's IP?
│   ├── YES → Layer 3 works
│   │   ├── Can you nc -zv X <port>?
│   │   │   ├── YES → TCP works. Check app layer (curl -v)
│   │   │   ├── "Connection refused" → Nothing listening on that port
│   │   │   └── "Timeout" → Firewall blocking that port
│   │   └── Are you using a hostname?
│   │       ├── dig <hostname> → works? IP correct?
│   │       └── Try by IP instead (skip DNS)
│   │
│   └── NO → Layer 3 broken
│       ├── Can you ping default gateway?
│       │   ├── YES → Gateway works. Check routes + remote firewalls
│       │   └── NO → Can't reach gateway
│       │       ├── Is interface UP? (ip link)
│       │       ├── Is IP correct? (ip addr)
│       │       └── Is gateway correct? (ip route)
│       └── Check traceroute: where does it stop?
```

### "It's slow"

```
X is slow
│
├── Where is the slowness?
│   curl -w "DNS: %{time_namelookup}\nTCP: %{time_connect}\n
│            TLS: %{time_appconnect}\nTTFB: %{time_starttransfer}\n"
│
├── DNS slow? (time_namelookup > 1s)
│   → dig @8.8.8.8 → faster? → local DNS problem
│   → K8s: ndots:5 issue? → use trailing dot
│
├── TCP connect slow? (time_connect - time_namelookup > expected RTT)
│   → Server SYN queue full? → check for SYN cookies
│   → High latency? → traceroute to check path
│
├── TLS slow? (time_appconnect - time_connect > 100ms)
│   → Certificate chain too long?
│   → OCSP stapling not enabled?
│   → TLS 1.2 instead of 1.3?
│
├── TTFB slow? (time_starttransfer - time_appconnect > 200ms)
│   → Server processing time
│   → Check server logs, database queries, external calls
│
└── Total slow but all phases fast?
    → Large response body + low bandwidth
    → Check bandwidth: iperf3
```

---

## Key Takeaways

1. **Start at the bottom: verify Layer 1 before Layer 7** — if ping doesn't work, don't debug HTTP. Fix connectivity first
2. **Change one thing at a time** — making multiple changes simultaneously means you won't know what fixed (or broke) things
3. **The first question is always: what changed?** — most network problems are caused by a recent change: deploy, config, certificate, firewall rule
4. **Use `curl -w` timing to pinpoint which phase is slow** — DNS, TCP connect, TLS, or server processing. This immediately narrows your search
5. **`ip route get <dest>` shows you EXACTLY how Linux will route a packet** — more useful than `ip route show` for debugging specific flows
6. **"Connection refused" and "connection timeout" mean very different things** — refused = reached server, port closed. Timeout = blocked or unreachable
7. **DNS failures masquerade as every other kind of failure** — always test DNS explicitly with `dig` before assuming the problem is elsewhere
8. **CLOSE_WAIT means application bug, not network bug** — the remote end closed the connection but the local app never called `close()`
9. **For intermittent issues, instrument and wait** — continuous ping with timestamps, curl monitoring loops, and rotating packet captures
10. **In Kubernetes, debug layer by layer: pod → DNS → pod IP → Service → external** — each step isolates a different component (app, CoreDNS, CNI, kube-proxy, NAT)

---

## Next Module

→ [Module 18: Real-World Failure Scenarios](../18-failure-scenarios/01-real-world-failures.md) — Production failures and how to diagnose them
