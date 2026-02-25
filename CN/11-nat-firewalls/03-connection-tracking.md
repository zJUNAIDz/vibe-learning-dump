# Connection Tracking (Conntrack) — The Stateful Engine

> Connection tracking is the invisible engine behind Linux stateful firewalling, NAT, and load balancing. Every packet that flows through a Linux machine is tracked by conntrack. When it works, you don't notice it. When it runs out of capacity or misbehaves, everything breaks.

---

## Table of Contents

1. [What Is Connection Tracking?](#what-is-conntrack)
2. [How Conntrack Works](#how-it-works)
3. [Conntrack Table](#conntrack-table)
4. [Conntrack States](#states)
5. [Timeouts](#timeouts)
6. [Conntrack and NAT](#conntrack-nat)
7. [Performance and Tuning](#performance)
8. [Conntrack in Containers and Kubernetes](#containers)
9. [Debugging Conntrack Issues](#debugging)
10. [Key Takeaways](#key-takeaways)

---

## What Is Connection Tracking?

Connection tracking (conntrack, nf_conntrack) is a Netfilter subsystem that tracks the state of network connections passing through a Linux machine.

### Why it exists

```
Without connection tracking (stateless firewall):
  → Allow TCP port 80 inbound: iptables -A INPUT -p tcp --dport 80 -j ACCEPT
  → Allow responses out: iptables -A OUTPUT -p tcp --sport 80 -j ACCEPT
  → But what about connections WE initiated outbound?
  → Must allow ALL inbound traffic to ephemeral ports (32768-60999)
  → Security hole!

With connection tracking (stateful firewall):
  → Allow inbound to port 80: -A INPUT -p tcp --dport 80 -j ACCEPT
  → Allow established/related: -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
  → Conntrack remembers connections → only matching replies allowed in
  → Secure!
```

### What conntrack tracks

```
Every connection through the machine gets an entry:
  - Protocol (TCP, UDP, ICMP, etc.)
  - Source IP and port
  - Destination IP and port  
  - Connection state (NEW, ESTABLISHED, etc.)
  - Timeout (when to expire this entry)
  - NAT translations (if NAT is applied)
  - Byte/packet counters
```

---

## How It Works

### Connection lifecycle

```
1. First packet arrives (SYN for TCP, any for UDP)
   → Conntrack creates NEW entry in table
   → Firewall rules can match: --ctstate NEW

2. Reply packet arrives (SYN-ACK for TCP)
   → Conntrack updates entry to ESTABLISHED
   → Firewall rules can match: --ctstate ESTABLISHED

3. Subsequent packets in both directions
   → Conntrack matches to existing entry (fast hash lookup)
   → Updates timeout counter (keeps entry alive)
   → If NAT applied, translates addresses

4. Connection ends (FIN/RST for TCP, timeout for UDP)
   → Conntrack marks entry for deletion
   → Entry removed after timeout
```

### Hash table lookup

```
Conntrack uses a hash table for O(1) lookups:

Hash key = hash(src_ip, src_port, dst_ip, dst_port, protocol)

Incoming packet:
  1. Compute hash from packet headers
  2. Look up in conntrack hash table
  3. If found → ESTABLISHED (match existing connection)
     If not found → NEW (create new entry)

Each bucket can have multiple entries (chaining for collisions).
Bucket count = net.netfilter.nf_conntrack_buckets
```

---

## Conntrack Table

### Viewing the table

```bash
# Show all tracked connections
conntrack -L
# or
cat /proc/net/nf_conntrack

# Example entries:
# tcp  6 431999 ESTABLISHED src=10.0.0.2 dst=93.184.216.34 sport=54321 dport=443 
#   src=93.184.216.34 dst=198.51.100.1 sport=443 dport=54321 [ASSURED] mark=0 use=1

# Breakdown:
#   tcp 6           → protocol name and number
#   431999          → timeout remaining (seconds)
#   ESTABLISHED     → connection state
#   First tuple     → original direction (client → server)
#   Second tuple    → reply direction (server → client, after NAT)
#   [ASSURED]       → connection has seen traffic in both directions
#   mark            → connection mark (for policy routing)
```

### Conntrack entry fields

| Field | Description |
|-------|-------------|
| Protocol | TCP (6), UDP (17), ICMP (1), etc. |
| Timeout | Seconds until entry expires |
| State | TCP state machine tracking |
| Original tuple | src/dst of first packet |
| Reply tuple | src/dst of expected replies (after NAT) |
| [ASSURED] | Bidirectional traffic seen → won't be evicted early |
| [UNREPLIED] | Only seen original direction (no reply yet) |
| mark | Packet mark (used by policy routing/QoS) |
| zone | Conntrack zone (for multi-tenant setups) |

### Filtering conntrack entries

```bash
# Only TCP connections
conntrack -L -p tcp

# Only connections to a specific IP
conntrack -L -d 93.184.216.34

# Only connections from a source
conntrack -L -s 10.0.0.2

# Only ESTABLISHED connections
conntrack -L --state ESTABLISHED

# Only connections on port 443
conntrack -L -p tcp --dport 443

# Count entries
conntrack -C

# Monitor new/destroyed connections in real-time
conntrack -E
# [NEW]    tcp  6 120 SYN_SENT src=10.0.0.2 dst=93.184.216.34 sport=54321 dport=443
# [UPDATE] tcp  6 60  SYN_RECV src=10.0.0.2 dst=93.184.216.34 sport=54321 dport=443
# [UPDATE] tcp  6 432000 ESTABLISHED src=10.0.0.2 ...
# [DESTROY] tcp 6 ... 

# Delete specific entry
conntrack -D -p tcp --dport 443 -d 93.184.216.34

# Flush all entries (caution!)
conntrack -F
```

---

## States

### Conntrack states (not the same as TCP states!)

```
NEW:
  First packet of a connection.
  TCP: SYN packet
  UDP: First packet in this 5-tuple
  ICMP: Echo request

ESTABLISHED:
  Reply packet seen. Connection is bidirectional.
  TCP: After SYN-ACK received
  UDP: After first reply packet
  ICMP: After echo reply

RELATED:
  New connection related to an existing ESTABLISHED connection.
  Examples:
    - ICMP "port unreachable" in response to UDP packet
    - ICMP "fragmentation needed" in response to TCP
    - FTP data connection related to FTP control connection
    - TFTP data transfer

INVALID:
  Packet doesn't match any known connection and isn't a valid NEW.
  Examples:
    - TCP packet with invalid flag combination
    - ICMP error for non-existent connection  
    - Packet that exceeds window (sequence number out of range)
  
  ALWAYS DROP INVALID packets:
    iptables -A INPUT -m conntrack --ctstate INVALID -j DROP

UNTRACKED:
  Explicitly excluded from tracking (using -j CT --notrack in raw table).
  Used for high-performance scenarios where stateful tracking is too expensive.
```

### TCP state tracking

Conntrack tracks the TCP state machine internally:

```
Conntrack state    TCP states seen
─────────────────────────────────────
NONE              No packets yet
SYN_SENT          SYN sent (client)
SYN_RECV          SYN-ACK sent (server)
ESTABLISHED       ACK received (3-way handshake done)
FIN_WAIT          FIN sent
CLOSE_WAIT        FIN received
LAST_ACK          FIN sent after receiving FIN
TIME_WAIT         Both FINs exchanged, waiting for timeout
CLOSE             Connection fully closed

These internal states are NOT the same as conntrack states (NEW, ESTABLISHED, etc.)
The ESTABLISHED conntrack state maps to multiple TCP states (ESTABLISHED, FIN_WAIT, etc.)
```

---

## Timeouts

### Default timeouts

```bash
# View all conntrack timeouts
sysctl -a | grep conntrack | grep timeout

# Key timeouts:
net.netfilter.nf_conntrack_tcp_timeout_established = 432000   # 5 days!
net.netfilter.nf_conntrack_tcp_timeout_syn_sent    = 120      # 2 min
net.netfilter.nf_conntrack_tcp_timeout_syn_recv    = 60       # 1 min
net.netfilter.nf_conntrack_tcp_timeout_fin_wait    = 120      # 2 min
net.netfilter.nf_conntrack_tcp_timeout_time_wait   = 120      # 2 min
net.netfilter.nf_conntrack_tcp_timeout_close        = 10      # 10 sec
net.netfilter.nf_conntrack_udp_timeout             = 30       # 30 sec
net.netfilter.nf_conntrack_udp_timeout_stream      = 120      # 2 min (bidirectional UDP)
net.netfilter.nf_conntrack_icmp_timeout            = 30       # 30 sec
```

### Why timeouts matter

```
TCP established timeout = 5 days:
  - Even idle TCP connections hold conntrack entries for 5 days!
  - Busy server with 100K connections → 100K entries for DAYS
  - If table fills up → new connections DROPPED!

Fix: Reduce timeouts for your workload:
  Web server: 300-600 seconds (5-10 min) is usually enough
  Load balancer: Depends on backend keepalive settings
  NAT gateway: Consider workload-specific tuning
```

```bash
# Tune timeouts
sysctl -w net.netfilter.nf_conntrack_tcp_timeout_established=600
sysctl -w net.netfilter.nf_conntrack_udp_timeout=30

# Make permanent
echo "net.netfilter.nf_conntrack_tcp_timeout_established = 600" >> /etc/sysctl.conf
sysctl -p
```

---

## Conntrack and NAT

### How NAT uses conntrack

```
Conntrack stores BOTH tuples — original and reply (NAT-translated):

Original:  src=10.0.0.2:54321 dst=93.184.216.34:443
Reply:     src=93.184.216.34:443 dst=198.51.100.1:54321
                                     ↑ NAT's public IP

Outbound packet:
  1. Conntrack lookup by original tuple → found
  2. Apply SNAT: rewrite src 10.0.0.2 → 198.51.100.1
  3. Send packet

Inbound reply:
  1. Conntrack lookup by reply tuple → found
  2. Apply reverse DNAT: rewrite dst 198.51.100.1 → 10.0.0.2
  3. Deliver to internal host
```

### NAT depends on conntrack

```
NAT cannot work without conntrack:
  - NAT needs to remember the mapping (which internal host)
  - Conntrack table IS the NAT translation table
  - If conntrack is disabled → NAT breaks

If you disable conntrack (raw table NOTRACK):
  - Those packets bypass NAT
  - Those packets bypass stateful firewall rules
  - Only use for specific high-performance scenarios
```

---

## Performance and Tuning

### Table size

```bash
# Current entries / maximum
cat /proc/sys/net/netfilter/nf_conntrack_count     # current entries
cat /proc/sys/net/netfilter/nf_conntrack_max       # maximum allowed

# Default max: 65536 (much too small for busy servers!)
# When full: new connections SILENTLY DROPPED
# Log message: "nf_conntrack: table full, dropping packet"

# Increase table size
sysctl -w net.netfilter.nf_conntrack_max=262144    # 256K entries
# Each entry ≈ 350 bytes → 256K entries ≈ 85 MB RAM

# Hash table size (should be ~max/4 for good performance)
# Can only be set at module load time:
echo "options nf_conntrack hashsize=65536" > /etc/modprobe.d/nf_conntrack.conf
# Or: echo 65536 > /sys/module/nf_conntrack/parameters/hashsize
```

### Sizing guidelines

```
Entries needed ≈ connections_per_second × average_connection_duration

Example: Web server
  1000 req/sec, 30 sec average connection duration
  = 30,000 entries + headroom = 65,536 (default may work)

Example: NAT gateway for 500 users
  500 users × 100 connections each × 300 sec avg = 150,000 entries
  → Set max=262144 with hashsize=65536

Example: Load balancer, 10,000 req/sec
  10,000 × 600 sec (keepalive) = 6,000,000 entries
  → Set max=8388608 with hashsize=2097152
  → ~2.8 GB RAM for conntrack table
```

### When conntrack table overflows

```
Symptoms:
  - Random connection failures
  - "table full, dropping packet" in dmesg/kern.log
  - New connections fail, established connections work
  - Appears as network failure but ping works (ICMP has separate conntrack)

Detection:
  dmesg | grep "table full"
  cat /proc/sys/net/netfilter/nf_conntrack_count  # compare to max
  
Emergency fix:
  sysctl -w net.netfilter.nf_conntrack_max=1048576
  
Proper fix:
  1. Increase max
  2. Reduce timeouts
  3. Consider disabling tracking for high-volume traffic (NOTRACK)
```

### Disabling conntrack for specific traffic

```bash
# Skip conntrack for high-throughput traffic (e.g., CDN, DNS)
# WARNING: These packets won't be statefully filtered or NAT'd

# iptables:
iptables -t raw -A PREROUTING -p udp --dport 53 -j CT --notrack
iptables -t raw -A OUTPUT -p udp --sport 53 -j CT --notrack

# nftables:
table inet raw {
    chain prerouting {
        type filter hook prerouting priority raw;
        udp dport 53 notrack
    }
}
```

---

## Conntrack in Containers and Kubernetes

### Docker

```
Docker creates iptables NAT rules for port mappings:
  docker run -p 8080:80 → DNAT from host:8080 to container:80
  
  Every container connection creates conntrack entries
  Container networking = heavy conntrack usage

Common issue:
  Many containers with many connections → conntrack table full
  → Random container connectivity failures
```

### Kubernetes

```
Kubernetes networking relies heavily on conntrack:

Services (ClusterIP):
  kube-proxy creates iptables DNAT rules
  Service IP → Pod IP via conntrack

NodePort:
  External:NodePort → DNAT → Pod:TargetPort
  Conntrack entry for each connection

LoadBalancer:
  External LB → NodePort → Pod
  Multiple layers of conntrack

Problem: Large Kubernetes clusters generate MASSIVE conntrack tables
  - Thousands of pods
  - Service mesh (Istio) = proxy per pod = 2× connections
  - Horizontal pod autoscaling = conn churn
```

### Kubernetes conntrack issues

```
Symptom: DNS resolution failures in pods
  → CoreDNS receives traffic via ClusterIP (iptables DNAT)
  → Conntrack race condition: UDP DNS request gets wrong DNAT
  → Known issue: conntrack entry collision on parallel DNS queries

Fix: 
  - Use TCP for DNS (resolv.conf: options use-vc)
  - Use node-local DNS cache
  - Upgrade to IPVS mode (kube-proxy --proxy-mode=ipvs)

Symptom: Service intermittently unreachable after pod restart
  → Old conntrack entries point to dead pod IP
  → New connections get NAT'd to non-existent pod
  
Fix:
  - conntrack -D (flush stale entries)
  - kube-proxy handles this but may have delays
```

### Tuning conntrack for Kubernetes

```bash
# On Kubernetes nodes, increase conntrack limits:
sysctl -w net.netfilter.nf_conntrack_max=1048576
sysctl -w net.netfilter.nf_conntrack_tcp_timeout_established=86400

# Add to /etc/sysctl.d/99-kubernetes.conf:
net.netfilter.nf_conntrack_max = 1048576
net.netfilter.nf_conntrack_tcp_timeout_established = 86400
net.netfilter.nf_conntrack_tcp_timeout_close_wait = 3600
```

---

## Debugging

### Essential commands

```bash
# Is conntrack loaded?
lsmod | grep nf_conntrack

# Table stats
cat /proc/sys/net/netfilter/nf_conntrack_count   # current
cat /proc/sys/net/netfilter/nf_conntrack_max     # max
conntrack -S  # detailed statistics per CPU

# conntrack -S output:
# cpu=0   found=12345 invalid=67 insert=0 insert_failed=0 drop=0 
#         early_drop=0 error=0 search_restart=45
#
# Key fields:
#   drop:           packets dropped due to full table
#   insert_failed:  couldn't insert (race condition)
#   early_drop:     UNREPLIED entries evicted to make room

# Watch table fill rate
watch -n 1 "cat /proc/sys/net/netfilter/nf_conntrack_count"

# Check for "table full" errors
dmesg | grep "table full"
journalctl -k | grep conntrack

# Distribution by state
conntrack -L 2>/dev/null | awk '{print $4}' | sort | uniq -c | sort -rn
# 45000 ESTABLISHED
#  3000 TIME_WAIT
#   500 SYN_SENT
#   200 FIN_WAIT
```

### Common scenarios

```
Scenario: "table full, dropping packet"
  → Increase nf_conntrack_max
  → Reduce timeouts
  → Check for connection leaks (app not closing connections)

Scenario: NAT stops working intermittently
  → Conntrack table full → NAT can't create mappings
  → Check count vs max

Scenario: Established connections work, new fail
  → Classic conntrack full symptom
  → ESTABLISHED entries not affected (already tracked)
  → NEW entries can't be created

Scenario: Connections reset after firewall rule change
  → iptables -F flushes rules but NOT conntrack entries
  → Old entries still allow traffic
  → Use conntrack -F to flush entries too (but drops all connections!)
```

---

## Key Takeaways

1. **Conntrack tracks every connection** through a Linux machine — it's the foundation for stateful firewalling and NAT
2. **Hash table lookup** makes conntrack fast (O(1) per packet), but table overflow is catastrophic
3. **States**: NEW (first packet), ESTABLISHED (reply seen), RELATED (associated connection), INVALID (always drop)
4. **TCP established timeout = 5 days** by default — far too long for most servers, tune it down
5. **Table full = silent packet drops** — the #1 conntrack issue. Monitor `nf_conntrack_count` vs `nf_conntrack_max`
6. **Default 65K entries is too small** for busy servers, NAT gateways, or Kubernetes nodes
7. **Each entry ≈ 350 bytes** — 1M entries ≈ 330 MB RAM. Size according to your connection rate × duration
8. **NAT requires conntrack** — disable conntrack = disable NAT for that traffic
9. **Kubernetes is conntrack-heavy** — service proxying, DNS, pod networking all create conntrack entries
10. **Debug with**: `conntrack -L` (view entries), `conntrack -S` (stats), `conntrack -E` (monitor), `dmesg` (errors)

---

## Next Module

→ [Module 12: Linux Network Internals](../12-linux-internals/01-socket-syscalls.md)
