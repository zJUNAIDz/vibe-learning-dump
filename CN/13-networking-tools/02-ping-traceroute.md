# ping and traceroute — Connectivity Testing and Path Discovery

> When something is "down," the first question is: **Can I reach it?** That's `ping`. The second question: **Where does the path break?** That's `traceroute`. These are the most basic and most important network debugging tools. Every network engineer uses them daily. Master the nuances.

---

## Table of Contents

1. [ping — Is It Reachable?](#ping)
2. [How ping Works (ICMP)](#icmp)
3. [Reading ping Output](#reading-ping)
4. [Advanced ping Usage](#advanced-ping)
5. [When ping Lies](#ping-lies)
6. [traceroute — Show Me the Path](#traceroute)
7. [How traceroute Works](#how-traceroute)
8. [Reading traceroute Output](#reading-traceroute)
9. [traceroute Variants](#variants)
10. [mtr — Best of Both Worlds](#mtr)
11. [Practical Debugging Scenarios](#scenarios)
12. [Key Takeaways](#key-takeaways)

---

## ping — Is It Reachable?

### Basic usage

```bash
# Ping a host
ping 8.8.8.8
# PING 8.8.8.8 (8.8.8.8) 56(84) bytes of data.
# 64 bytes from 8.8.8.8: icmp_seq=1 ttl=118 time=12.3 ms
# 64 bytes from 8.8.8.8: icmp_seq=2 ttl=118 time=11.8 ms
# 64 bytes from 8.8.8.8: icmp_seq=3 ttl=118 time=12.1 ms
# ^C
# --- 8.8.8.8 ping statistics ---
# 3 packets transmitted, 3 received, 0% packet loss, time 2003ms
# rtt min/avg/max/mdev = 11.800/12.067/12.300/0.205 ms

# Ping a hostname (tests DNS + connectivity)
ping google.com

# Ping with count (stop after N)
ping -c 5 8.8.8.8

# Ping with interval
ping -i 0.2 8.8.8.8    # every 200ms (default is 1 second)

# Flood ping (requires root, maximum rate)
sudo ping -f 10.0.0.1
# ....                 ← dots show sent without reply
# Useful for stress testing, NOT for production
```

---

## How ping Works (ICMP)

```
ICMP (Internet Control Message Protocol):
  - Layer 3 protocol (carried inside IP)
  - NOT TCP or UDP — it's its own protocol (IP protocol number 1)
  - Used for network diagnostics and error reporting

Ping uses two ICMP message types:
  Type 8: Echo Request  (sent by ping)
  Type 0: Echo Reply    (sent back by target)

Process:
  1. Your machine sends ICMP Echo Request to target
  2. Target receives, sends ICMP Echo Reply back
  3. Your machine measures round-trip time (RTT)

  ┌──────────┐                        ┌──────────┐
  │  Client   │ ── ICMP Echo Req ──→  │  Server   │
  │           │ ←── ICMP Echo Reply ── │           │
  └──────────┘                        └──────────┘
       △                                    △
       │                                    │
       └── measures time ──────────────────┘
              = Round Trip Time (RTT)
```

### ICMP packet structure

```
IP Header (20 bytes):
  Protocol: 1 (ICMP)
  Source IP: your IP
  Dest IP: target IP
  TTL: 64 (default Linux)

ICMP Header (8 bytes):
  Type: 8 (request) or 0 (reply)
  Code: 0
  Checksum
  Identifier: process ID (to match replies)
  Sequence: increments each ping

ICMP Data (56 bytes default):
  Timestamp (for RTT calculation)
  Padding

Total: 20 (IP) + 8 (ICMP) + 56 (data) = 84 bytes on wire
(Plus 14 bytes Ethernet = 98 bytes total)
```

---

## Reading ping Output

```bash
ping -c 5 8.8.8.8
# 64 bytes from 8.8.8.8: icmp_seq=1 ttl=118 time=12.3 ms
#  │                       │          │       │
#  │                       │          │       └── RTT (round-trip time)
#  │                       │          └── TTL remaining (started at 128, 
#  │                       │              10 hops = 128-10=118)
#  │                       └── Sequence number (detect loss/reorder)
#  └── Response size (56 data + 8 ICMP header = 64)

# Statistics section:
# 5 packets transmitted, 5 received, 0% packet loss, time 4006ms
# rtt min/avg/max/mdev = 11.5/12.1/12.8/0.424 ms
#      │    │    │    │
#      │    │    │    └── Standard deviation (jitter measurement)
#      │    │    └── Worst case
#      │    └── Average (the number people quote)
#      └── Best case
```

### What the numbers mean

```
RTT interpretation:
  < 1 ms:    Same LAN / localhost
  1-20 ms:   Same city / region
  20-80 ms:  Same continent
  80-150 ms: Cross-continent
  150-300 ms: Intercontinental (e.g., US ↔ Asia)
  > 300 ms:  Satellite or very distant

Packet loss:
  0%:   Perfect
  1-2%: Minor issue (WiFi, slightly overloaded link)
  5%+:  Serious problem affecting TCP performance
  100%: Unreachable OR ICMP blocked (not necessarily "down"!)

mdev (jitter):
  Low mdev:  Stable link
  High mdev: Inconsistent path (WiFi, congested link, multiple paths)
```

### Detecting issues from ping

```
Symptom: High RTT on first ping, normal after
  = ARP resolution (first ping triggers ARP lookup)

Symptom: Increasing RTT over time
  = Bufferbloat (router buffers filling up)

Symptom: Periodic spikes in RTT
  = Congestion at a specific time, or routing changes

Symptom: Packet loss but low RTT on successful pings
  = Likely random drops (congested link, bad cable)

Symptom: 100% loss
  = Host down, ICMP blocked, or routing problem
  → DON'T assume host is down! Many hosts block ICMP
```

---

## Advanced ping Usage

```bash
# Change packet size (test MTU / fragmentation)
ping -s 1472 -M do 8.8.8.8
# -s 1472: 1472 data + 8 ICMP + 20 IP = 1500 (standard MTU)
# -M do: Don't Fragment flag set
# If too big: "Frag needed and DF set (mtu = 1500)"
# Use to discover Path MTU

# Ping from specific source
ping -I eth0 8.8.8.8         # by interface
ping -I 10.0.0.5 8.8.8.8    # by IP

# Set TTL
ping -t 5 8.8.8.8
# Will be dropped after 5 hops → "Time exceeded"

# Record route (show path in reply — limited to 9 hops)
ping -R 8.8.8.8

# Timestamp
ping -T tsonly 8.8.8.8

# Quiet mode (only summary)
ping -q -c 100 8.8.8.8

# Adaptive ping (faster when replies arrive)
sudo ping -A 8.8.8.8

# Set deadline (total time) vs count
ping -w 10 8.8.8.8    # stop after 10 seconds regardless
ping -c 100 8.8.8.8   # stop after 100 packets
```

### IPv6 ping

```bash
# IPv6 uses ping6 or ping with -6
ping -6 google.com
ping6 ::1
```

---

## When ping Lies

```
Scenario: ping works but application doesn't
  - ICMP handled differently than TCP/UDP
  - Firewall may allow ICMP but block TCP port
  - Host may respond to ping but service may be down

Scenario: ping fails but application works
  - Host blocks ICMP (common on cloud: AWS SGs default-block ICMP)
  - Rate-limited ICMP on routers
  - Firewall drops ICMP specifically

Scenario: ping works but performance is terrible
  - ICMP is deprioritized by many routers
  - Ping RTT may not reflect actual TCP performance
  - TCP may take a different path than ICMP (ECMP)

Rule: NEVER rely solely on ping for diagnosis.
  Use ss, curl, telnet, or actual application testing.
```

---

## traceroute — Show Me the Path

### Basic usage

```bash
# Trace route to destination
traceroute 8.8.8.8
# traceroute to 8.8.8.8 (8.8.8.8), 30 hops max, 60 byte packets
#  1  gateway (10.0.0.1)  0.5 ms  0.4 ms  0.4 ms
#  2  isp-router (203.0.113.1)  5.2 ms  5.1 ms  5.3 ms
#  3  core-1.isp.net (198.51.100.1)  10.1 ms  10.2 ms  10.0 ms
#  4  * * *
#  5  dns.google (8.8.8.8)  12.3 ms  12.1 ms  12.2 ms

# Each line = one router hop
# Three times per hop (3 probes per TTL)
# * = no response (router doesn't reply or drops)
```

---

## How traceroute Works

### TTL trick

```
traceroute exploits TTL (Time To Live):

  TTL decremented by 1 at each router
  When TTL = 0 → router drops packet and sends ICMP "Time Exceeded"
  By sending packets with TTL = 1, 2, 3, ... → learn each router

TTL=1:
  ┌────────┐    TTL=0!    ┌──────┐
  │ Source  │ ──────────→  │ Hop 1│ ──→ ICMP Time Exceeded back to source
  └────────┘              └──────┘

TTL=2:
  ┌────────┐    TTL=1     ┌──────┐    TTL=0!    ┌──────┐
  │ Source  │ ──────────→  │ Hop 1│ ──────────→  │ Hop 2│ → ICMP Time Exceeded
  └────────┘              └──────┘              └──────┘

TTL=3:
  ┌────────┐    TTL=2     ┌──────┐    TTL=1     ┌──────┐    TTL=0!   ┌──────┐
  │ Source  │ ──────────→  │ Hop 1│ ──────────→  │ Hop 2│ ─────────→ │ Hop 3│
  └────────┘              └──────┘              └──────┘             └──────┘
                                                                    ICMP Time Exceeded

  Continue until destination reached or max hops (default 30)
```

### Probe methods

```
traceroute can send different types of probes:

UDP (default on Linux):
  Sends UDP to high port (33434+)
  Destination responds with ICMP "Port Unreachable" (how we know we arrived)
  Problem: some firewalls block UDP

ICMP (-I flag):
  Sends ICMP Echo Request (same as ping)
  More likely to get through firewalls
  sudo traceroute -I 8.8.8.8

TCP SYN (-T flag):
  Sends TCP SYN to port 80 or 443
  Most likely to get through firewalls (looks like web traffic)
  sudo traceroute -T -p 443 8.8.8.8
```

---

## Reading traceroute Output

```bash
traceroute 8.8.8.8
#  1  10.0.0.1 (10.0.0.1)         0.5 ms    0.4 ms    0.4 ms
#  2  203.0.113.1 (203.0.113.1)   5.2 ms    5.1 ms    5.3 ms
#  3  198.51.100.1 (198.51.100.1) 10.1 ms   10.2 ms   10.0 ms
#  4  * * *
#  5  216.239.49.227              25.3 ms   25.1 ms   25.4 ms
#  6  8.8.8.8 (8.8.8.8)          12.3 ms   12.1 ms   12.2 ms

# Hop 1: Your gateway (0.5 ms — local network)
# Hop 2: ISP's first router (5 ms — jump = ISP access network)
# Hop 3: ISP core (10 ms — through ISP backbone)
# Hop 4: * * * — Router doesn't respond (ICMP rate-limited or blocked)
# Hop 5: Google's edge (25 ms — but note it's HIGHER than final)
# Hop 6: Destination (12 ms)
```

### Common patterns

```
Pattern: Obvious latency jump
  3  198.51.100.1    10 ms
  4  72.14.209.81    95 ms    ← 85 ms jump!
  5  209.85.251.9    96 ms
  Conclusion: intercontinental link between hops 3 and 4

Pattern: * * * row(s) then continues
  4  * * *
  5  * * *
  6  72.14.209.81    25 ms
  Conclusion: routers 4-5 don't reply to traceroute probes (normal)
  The path is FINE — just silent routers

Pattern: All *'s after a certain point
  3  198.51.100.1    10 ms
  4  * * *
  5  * * *
  ...
  30 * * *
  Conclusion: Firewall blocking probes OR host unreachable

Pattern: RTT decreases at later hops
  3  198.51.100.1    50 ms
  4  72.14.209.81    12 ms   ← lower than hop 3?!
  Explanation: RTT per hop is NOT cumulative!
  Each measurement is independent round-trip to that router.
  Router at hop 3 might be slow to generate ICMP responses.
  
CRITICAL: Per-hop times are NOT latency BETWEEN hops!
  They are separate RTTs from source to each hop.
```

### Asymmetric routing caveat

```
traceroute shows the FORWARD path only.
The return path for ICMP responses may be completely different.

Forward:  You → A → B → C → Destination
Return:   Destination → X → Y → Z → You

This means:
  - RTT at hop 3 includes return path you don't see
  - A "slow" hop may actually be a slow return path
  - Don't blame a specific hop without confirming with the hop's admin
```

---

## traceroute Variants

### tracepath (no root needed)

```bash
# Discovers MTU along path
tracepath 8.8.8.8
#  1?: [LOCALHOST]     pmtu 1500
#  1:  gateway         0.5ms
#  2:  203.0.113.1     5.2ms pmtu 1480   ← MTU changed!
#  3:  198.51.100.1    10.1ms reached
#     Resume: pmtu 1480   ← Path MTU = 1480
```

### tcptraceroute

```bash
# Uses TCP SYN — better firewall traversal
sudo tcptraceroute 8.8.8.8 443
# Like traceroute -T -p 443 but more features
```

### Paris traceroute

```bash
# Fixes ECMP (load-balanced path) issues
# Standard traceroute may show different paths per hop (misleading)
# Paris traceroute keeps flow identifier constant

paris-traceroute 8.8.8.8
```

---

## mtr — Best of Both Worlds

### What mtr is

```
mtr = traceroute + ping combined
  - Continuously sends probes
  - Shows packet loss and latency statistics per hop
  - Updates in real-time
  - THE tool for diagnosing intermittent issues
```

### Using mtr

```bash
# Interactive mode (real-time updates)
mtr 8.8.8.8
#                          Host              Loss%  Snt   Last  Avg   Best  Wrst  StDev
#  1. gateway               0.0%   100    0.5   0.4   0.3   1.2   0.1
#  2. 203.0.113.1            0.0%   100    5.2   5.3   4.8   6.1   0.3
#  3. 198.51.100.1           2.0%   100   10.1  10.3   9.5  15.2   1.2
#  4. 72.14.209.81           0.0%   100   25.3  25.5  24.1  28.2   0.8
#  5. 8.8.8.8                0.0%   100   12.3  12.4  11.8  14.1   0.5

# Report mode (run for N cycles, then print report)
mtr -rw -c 100 8.8.8.8
# -r = report mode
# -w = wide output (show full hostnames)
# -c = count

# TCP mode
mtr -T -P 443 8.8.8.8

# JSON output (for scripting)
mtr -j 8.8.8.8
```

### Reading mtr output

```
Loss%: Packet loss at this hop
  0% everywhere:           Network is healthy
  Loss at intermediate hop but 0% at destination:
    → That router rate-limits ICMP responses (NORMAL, not a problem)
  Loss at hop AND all hops after it:
    → Actual packet loss at that hop (real problem)
  Loss only at destination:
    → Destination's problem (overloaded, rate-limiting)

Avg vs Last:
  Avg is more reliable (smoothed over many probes)
  Last can spike due to a single bad probe

StDev (Standard Deviation):
  Low:  Stable connection
  High: Jitter (inconsistent, bufferbloat, load balancing)
```

---

## Practical Debugging Scenarios

### Scenario: Can't reach a website

```bash
# Step 1: Is DNS working?
dig example.com
# If no answer → DNS issue (check /etc/resolv.conf, try 8.8.8.8)

# Step 2: Can I reach the IP?
ping -c 3 93.184.216.34
# If timeout → routing or firewall issue

# Step 3: Where does the path break?
mtr -rw -c 20 93.184.216.34
# Look for the hop where loss starts

# Step 4: Is the port open?
ss -tn state established | grep 93.184.216.34
# or
timeout 5 bash -c "echo > /dev/tcp/93.184.216.34/443" && echo "open" || echo "closed"
```

### Scenario: Intermittent connectivity

```bash
# mtr with extended count
mtr -rwz -c 1000 example.com
# -z = show AS numbers
# Let it run for 1000 cycles — catch intermittent loss

# Continuous ping with timestamps
ping -D 8.8.8.8 | tee /tmp/ping.log
# -D = print timestamps
# Review later for patterns (every 5 min? every hour?)
```

### Scenario: High latency — where's the bottleneck?

```bash
mtr -rw -c 100 slow-server.com
#  1. gateway         0.0%  1.0 ms
#  2. isp-edge        0.0%  5.0 ms     ← +4 ms (local → ISP)
#  3. isp-core        0.0%  10.0 ms    ← +5 ms (ISP internal)
#  4. ix-peer         0.0%  15.0 ms    ← +5 ms (peering)
#  5. remote-edge     0.0%  95.0 ms    ← +80 ms! Bottleneck!
#  6. destination     0.0%  96.0 ms    ← +1 ms (normal)

# The 80ms jump between hops 4-5 is likely an ocean crossing
# or a severely congested link
```

---

## Key Takeaways

1. **ping tests reachability** — it uses ICMP, measures RTT, and detects packet loss
2. **100% ping loss ≠ host down** — many hosts/firewalls block ICMP. Always verify with TCP
3. **RTT < 1ms = LAN**, 1-20ms = local, 20-80ms = same region, 80-300ms = intercontinental
4. **traceroute reveals the path** — uses TTL trick to discover each router hop
5. **Per-hop RTT is NOT hop-to-hop latency** — each is an independent round trip from source
6. **`* * *` hops are usually fine** — router just doesn't respond to traceroute probes
7. **Use TCP traceroute** (`-T`) through firewalls — it looks like web traffic
8. **mtr is the superior tool** — combines ping + traceroute with continuous monitoring
9. **Loss at intermediate hop only ≠ real loss** — if destination has 0% loss, intermediate hops are just rate-limiting ICMP
10. **ping -s 1472 -M do** discovers MTU issues — if it fails, path MTU < 1500

---

## Next

→ [tcpdump](./03-tcpdump.md) — Capture and analyze packets on the wire
