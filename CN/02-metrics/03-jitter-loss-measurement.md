# Jitter, Packet Loss, and Network Measurement

> These two metrics — jitter and packet loss — are responsible for more real-world application failures than low bandwidth ever is. Let's understand them deeply.

---

## Table of Contents

1. [Jitter: Why Consistency Matters More Than Speed](#jitter-why-consistency-matters-more-than-speed)
2. [What Causes Jitter](#what-causes-jitter)
3. [How Jitter Breaks Applications](#how-jitter-breaks-applications)
4. [Measuring Jitter on Linux](#measuring-jitter-on-linux)
5. [Packet Loss: The Silent Killer](#packet-loss-the-silent-killer)
6. [What Causes Packet Loss](#what-causes-packet-loss)
7. [How Packet Loss Breaks Things](#how-packet-loss-breaks-things)
8. [Measuring Packet Loss on Linux](#measuring-packet-loss-on-linux)
9. [Simulating Bad Networks (For Testing)](#simulating-bad-networks-for-testing)
10. [The Measurement Meta-Problem](#the-measurement-meta-problem)
11. [Building a Network Health Dashboard](#building-a-network-health-dashboard)

---

## Jitter: Why Consistency Matters More Than Speed

### What jitter is

**Jitter** is the variation in latency over time. If every packet arrives with exactly 50ms delay, jitter is zero. If packets arrive with delays of 40ms, 60ms, 45ms, 80ms, 35ms — that variation is jitter.

Formally, jitter is often measured as the **standard deviation** of latency, or as the **inter-packet delay variation** (the difference in delay between consecutive packets).

### Why jitter matters

Consider two scenarios, both with 50ms average latency:

**Scenario A (low jitter):**
```
Packet 1: 48ms
Packet 2: 51ms
Packet 3: 49ms
Packet 4: 52ms
Packet 5: 50ms
Average: 50ms, Jitter: ~1.5ms
```

**Scenario B (high jitter):**
```
Packet 1: 20ms
Packet 2: 110ms
Packet 3: 30ms
Packet 4: 90ms
Packet 5: 50ms
Average: 50ms, Jitter: ~37ms
```

Same average latency. Completely different user experience.

For a VoIP call in scenario A: smooth, natural conversation.
For a VoIP call in scenario B: choppy, robotic audio with awkward pauses.

### The jitter buffer

Applications that need steady data flow (audio, video) use a **jitter buffer** — a small buffer that holds incoming packets and plays them out at a steady rate.

```
Without jitter buffer:
  Packets arrive:   ──●──────●●──────────●──●──
  Audio plays:      choppy, gaps, bunching

With jitter buffer (50ms):
  Packets arrive:   ──●──────●●──────────●──●──
  Buffer absorbs:   [====================]
  Audio plays:      ──●──●──●──●──●──●── (smooth, but 50ms delayed)
```

The trade-off: a larger jitter buffer absorbs more variation but adds more delay. Too much delay breaks interactivity (you hear the other person's response late).

Typical jitter buffers:
- VoIP: 20-60ms
- Video conferencing: 50-200ms  
- Live streaming: 1-5 seconds
- On-demand video: 5-30 seconds (not really jitter buffer, but similar concept)

---

## What Causes Jitter

### 1. Variable queuing delay

The most common cause. As traffic at routers fluctuates, packets experience different queuing delays. During a burst of traffic, a packet waits longer. During a lull, the next packet sails through.

### 2. Route changes

If the network path changes (due to a link failure or routing protocol convergence), packets suddenly take a different path with a different latency. Until the path stabilizes, consecutive packets may take different routes.

### 3. Resource contention

On shared links (Wi-Fi, cable internet), your packets compete with other users' traffic. When another user starts a large download, your packets may be delayed.

### 4. Software scheduling

Operating systems are not real-time. The kernel's packet processing can be delayed by:
- CPU load (another process is running)
- Interrupt coalescing (the NIC batches interrupts for efficiency, adding delay)
- Power saving states (CPU may be in low-power mode, taking microseconds to wake up)

### 5. Wireless effects

Wi-Fi is particularly jittery because:
- Only one device can transmit at a time (CSMA/CA)
- Interference causes retransmissions
- Signal quality varies with distance and obstacles
- Channel sharing with neighbors

---

## How Jitter Breaks Applications

### Real-time communication (VoIP, video conferencing)

High jitter → packets arrive out of cadence → jitter buffer fills/empties unpredictably → audio/video glitches

If jitter exceeds the jitter buffer size, packets are considered "too late" and discarded. This appears as packet loss to the application, even though the packets eventually arrived.

### Online gaming

Games predict near-future state based on current data. High jitter means predictions are wrong more often, causing "rubber-banding" (objects snapping to new positions) and "hit registration" issues.

### Stock trading / financial systems

Jitter means unpredictable order execution times. In high-frequency trading, microseconds of jitter can mean millions of dollars of difference.

### TCP performance

TCP's retransmission timer (RTO) is based on measured RTT. High jitter means high RTT variance, which makes TCP set conservative (large) RTO values. Large RTO means TCP waits a long time before retransmitting, reducing throughput.

```
With stable RTT (50ms ± 2ms):
  SRTT = 50ms, RTTvar = 2ms
  RTO = 50 + 4×2 = 58ms → fast retransmission

With jittery RTT (50ms ± 40ms):
  SRTT = 50ms, RTTvar = 40ms
  RTO = 50 + 4×40 = 210ms → slow retransmission
```

---

## Measuring Jitter on Linux

### Method 1: ping statistics

```bash
# Send 100 pings, 200ms interval
ping -c 100 -i 0.2 8.8.8.8

# Look at the summary line:
# rtt min/avg/max/mdev = 11.5/12.8/15.2/1.1 ms
#                                           ^^^
#                                           mdev = standard deviation ≈ jitter

# Low mdev (<2ms): good
# Medium mdev (2-10ms): noticeable for real-time apps
# High mdev (>10ms): problematic for VoIP/gaming
```

### Method 2: mtr jitter column

```bash
# mtr shows jitter per hop
mtr --report -c 100 example.com

# StDev column shows jitter at each hop
# Find where jitter increases to identify the source
```

### Method 3: iperf3 UDP jitter

```bash
# Server
iperf3 -s

# Client: UDP test at 1 Mbps
iperf3 -c <server> -u -b 1M -t 30

# Output includes:
# Jitter: 2.345 ms  ← inter-packet delay variation
# Lost/Total: 5/30000 (0.017%)
```

### Method 4: Continuous monitoring

```bash
# Script to monitor jitter over time
while true; do
  ping -c 10 -q 8.8.8.8 | grep "mdev" | awk -F'/' '{print strftime("%H:%M:%S"), "avg="$5, "jitter="$7}'
  sleep 5
done

# Output:
# 14:30:15 avg=12.5 jitter=1.2
# 14:30:25 avg=13.1 jitter=2.8
# 14:30:35 avg=45.3 jitter=22.1  ← SPIKE! Something changed
```

---

## Packet Loss: The Silent Killer

### What packet loss means

**Packet loss** occurs when packets sent from A don't arrive at B. They're gone. Silently. No error message (at the IP layer).

### Types of loss

1. **Random loss**: Individual packets dropped unpredictably. Usually caused by link errors (electromagnetic interference, CRC failures) or router buffer overflow.

2. **Burst loss**: Multiple consecutive packets lost together. Usually caused by buffer overflow during traffic spikes, or interface flaps (link goes down momentarily).

3. **Correlated loss**: Loss rate increases during certain conditions (time of day, traffic patterns). Usually congestion-related.

4. **Tail drop**: When a router buffer fills completely, all new packets are dropped until buffer space frees up. This tends to drop bursts of packets from the same flow.

### What does "1% packet loss" actually mean?

It means that out of every 100 packets, one is dropped. That sounds small. But consider TCP:

- TCP must detect the loss (via timeout or duplicate ACKs — takes 1-3 RTTs)
- TCP must retransmit the lost packet (1 RTT minimum)
- TCP interprets loss as congestion → cuts its sending rate in half
- TCP then slowly ramps back up

For interactive traffic:
- 0% loss: everything works as expected
- 0.01-0.1%: barely noticeable, TCP handles it efficiently
- 0.1-1%: noticeable throughput reduction, occasional minor delays
- 1-5%: significant throughput reduction, TCP spends much time recovering
- 5-10%: very poor TCP performance, real-time apps (VoIP) struggle badly
- 10%+: practically unusable for TCP, will be very slow with many retransmissions

For UDP (real-time):
- VoIP can tolerate up to ~1% loss with concealment algorithms
- Video streaming up to ~2-5% with error resilience
- Gaming up to ~1-2% with prediction/interpolation
- Above these thresholds: noticeable quality degradation

---

## What Causes Packet Loss

### 1. Congestion (buffer overflow)

The most common cause. When a router receives more traffic than it can forward, its buffer fills up. Once full, incoming packets are dropped.

```
Traffic in: 2 Gbps (across multiple links)
Traffic out: 1 Gbps (one link)
Buffer: 10 MB

Buffer fills in: 10 MB / 1 Gbps excess = ~80ms
After 80ms: packets start dropping
```

This is "normal" — it's how the internet signals congestion. TCP interprets loss as "slow down." The problem isn't that loss happens — it's when you have SO MUCH buffering that loss doesn't happen soon enough (bufferbloat — covered in Module 07).

### 2. Link errors

Physical medium issues:
- Damaged cables
- Electromagnetic interference (EMI)
- Poor fiber connectors
- Wireless interference
- Signal attenuation over long cables

These cause bit errors. The data link layer detects them (CRC check fails) and drops the corrupted frame. TCP eventually retransmits.

```bash
# Check for link errors
ip -s link show eth0
# Look at:
#   RX errors, dropped, overruns, frame
#   TX errors, dropped, carrier, collisions
# Non-zero values indicate physical or driver issues

# More detailed
ethtool -S eth0
# Shows per-driver statistics including CRC errors, alignment errors, etc.
```

### 3. Hardware/software limits

- **NIC ring buffer overflow**: Packets arrive faster than the kernel can process them
- **CPU overload**: The kernel can't process packets fast enough
- **Memory pressure**: Not enough memory for socket buffers

```bash
# Check if the NIC is dropping packets
ethtool -S eth0 | grep -i drop
# rx_dropped, tx_dropped, rx_missed_errors

# Check kernel drop statistics
cat /proc/net/softnet_stat
# Third column: number of times the backlog was full (packets dropped)

# Check for OOM-related drops
dmesg | grep -i "drop\|oom\|memory"
```

### 4. Firewall / ACL drops

Packets intentionally dropped by firewall rules. These aren't errors — they're policy.

```bash
# Check iptables drop counters
sudo iptables -L -v -n
# The "pkts" column shows how many packets matched each rule
# DROP rules show intentionally discarded packets

# Check nftables (modern Linux)
sudo nft list ruleset
```

### 5. MTU mismatches

If a packet is too large for a link and the "Don't Fragment" (DF) flag is set, the router SHOULD send back an ICMP "Fragmentation Needed" message. But if ICMP is blocked (many firewalls do this), the packet is silently dropped. This creates an "MTU black hole."

```bash
# Test path MTU
ping -M do -s 1472 destination
# -M do: set Don't Fragment
# -s 1472: payload size (1472 + 28 = 1500 total)
# If it works, try higher values until it fails

# Check interface MTU
ip link show eth0 | grep mtu
```

### 6. Rate limiting

Routers/firewalls may intentionally rate-limit certain traffic (like ICMP), dropping packets that exceed the rate.

---

## How Packet Loss Breaks Things

### TCP: Loss means slowdown

TCP interprets loss as congestion. On every loss:
1. TCP halves its congestion window (sending rate drops by ~50%)
2. TCP retransmits the lost segment (uses bandwidth without new progress)
3. TCP slowly ramps back up (takes many RTTs to recover)

Net effect: throughput drops significantly, and recovery is slow.

### UDP applications: Loss means missing data

Since UDP doesn't retransmit:
- VoIP: Brief silence or distorted audio during loss
- Video: Artifacts, freezes, blocky video
- DNS: Query timeout (falls back to retry after ~2 seconds)
- Gaming: Character teleportation, missed actions

### Connection failures

Heavy loss can cause:
- **TCP SYN losses**: Connection can't be established (must retry after 1 second, then 2, then 4...)
- **TCP data losses**: If too many retransmissions fail, TCP gives up (~2-9 minutes depending on OS)
- **Keep-alive failures**: TCP connections declared dead when keep-alive probes are lost

---

## Measuring Packet Loss on Linux

### Method 1: ping

```bash
# Send 100 pings to detect loss
ping -c 100 8.8.8.8

# Look at:
# 100 packets transmitted, 97 received, 3% packet loss
#                                        ^^^^^^^^^^
# 0% is ideal
# <0.1% is acceptable for most uses
# >1% is problematic
```

**Important caveat**: ICMP loss doesn't necessarily equal TCP loss. Some routers prioritize TCP over ICMP, or rate-limit ICMP. But if you see loss on ping, there's almost certainly a problem.

### Method 2: mtr (per-hop loss)

```bash
mtr --report -c 200 example.com

# Host                     Loss%   Snt   Last   Avg  Best  Wrst StDev
# 1. 192.168.1.1            0.0%   200    1.0   1.2   0.5   3.0   0.5
# 2. 10.0.0.1               0.0%   200    5.0   5.5   4.0   8.0   1.0
# 3. isp-core.example.com   2.5%   200   15.0  18.0  12.0  45.0   8.0  ← LOSS HERE
# 4. peer.example.com       0.0%   200   20.0  22.0  18.0  30.0   3.0
# 5. destination.com        0.0%   200   25.0  26.0  22.0  35.0   3.0
```

**How to read mtr loss**:
- If hop 3 shows 2.5% loss but hops 4 and 5 show 0% — hop 3 is just rate-limiting ICMP (not actually dropping traffic). This is a false alarm.
- If hop 3 shows 2.5% loss AND hops 4 and 5 ALSO show ~2.5% loss — there's real loss at hop 3.

This "loss at intermediate hops but not at destination" pattern is VERY common and misleads many people.

### Method 3: iperf3

```bash
# UDP test with loss measurement
iperf3 -c <server> -u -b 50M -t 30

# Output shows:
# [  5]   0.00-30.00  sec  178 MBytes  49.9 Mbits/sec  0.123 ms  25/22803 (0.11%)  sender
#                                                                 ^^^^^^^^^^^^^^^^
#                                                                 Lost/Total (Loss%)
```

### Method 4: TCP retransmission monitoring

```bash
# Watch retransmissions in real-time
watch -n 1 'cat /proc/net/snmp | grep Tcp | tail -1'
# InSegs and OutSegs grow steadily
# RetransSegs growing → loss is occurring

# Or use ss for per-connection retransmissions
ss -ti | grep retrans
# Shows retransmission count per connection

# Or nstat for counters
nstat -a | grep -i retrans
# TcpRetransSegs: total retransmitted segments
```

### Method 5: tcpdump for loss analysis

```bash
# Capture traffic and look for retransmissions
sudo tcpdump -i any -nn 'tcp[tcpflags] & (tcp-syn) != 0' -c 100
# Count SYN retransmissions → connection establishment failures

# Capture all TCP traffic and analyze later
sudo tcpdump -i any -nn -w /tmp/capture.pcap tcp
# Open in Wireshark → Analyze → Expert Information → shows retransmissions
```

---

## Simulating Bad Networks (For Testing)

### Why simulate

You can't test how your application handles packet loss and jitter on a perfect local network. Linux's `tc` (traffic control) tool lets you simulate real-world network conditions.

### tc netem: Network emulation

```bash
# Add 100ms delay to all traffic on eth0
sudo tc qdisc add dev eth0 root netem delay 100ms

# Add 100ms delay with 20ms jitter (random variation)
sudo tc qdisc add dev eth0 root netem delay 100ms 20ms

# Add 100ms delay with 20ms jitter, correlated 25%
# (consecutive packets tend to have similar delays)
sudo tc qdisc add dev eth0 root netem delay 100ms 20ms 25%

# Add 5% random packet loss
sudo tc qdisc add dev eth0 root netem loss 5%

# Add 5% loss with 25% correlation (bursty loss)
sudo tc qdisc add dev eth0 root netem loss 5% 25%

# Combine: 50ms delay + 10ms jitter + 1% loss
sudo tc qdisc add dev eth0 root netem delay 50ms 10ms loss 1%

# Add packet reordering: 25% of packets reordered, 50% correlated
sudo tc qdisc add dev eth0 root netem delay 50ms reorder 25% 50%

# Add packet duplication: 1% of packets duplicated
sudo tc qdisc add dev eth0 root netem duplicate 1%

# Limit bandwidth as well (1 Mbps)
sudo tc qdisc add dev eth0 root tbf rate 1mbit burst 32kbit latency 400ms

# Remove all tc rules (restore normal)
sudo tc qdisc del dev eth0 root

# Check current rules
tc qdisc show dev eth0
```

### Testing your application under bad conditions

```bash
# Step 1: Add impairment
sudo tc qdisc add dev lo root netem delay 50ms 10ms loss 2%

# Step 2: Test your service
curl -v http://localhost:8080/api/test
# Observe: increased response time, occasional failures

# Step 3: Remove impairment
sudo tc qdisc del dev lo root

# Use on loopback (lo) to test local services
# Use on eth0/wlan0 to test remote services
```

### Warning about tc on production machines

**Never apply tc rules on production network interfaces** unless you know exactly what you're doing. Applying 5% loss to a production server's eth0 will cause real outages. Always test on:
- Loopback interface (lo)
- Network namespaces (isolated)
- VMs or containers
- Dedicated test environments

---

## The Measurement Meta-Problem

### You measure the tool, not just the network

Every measurement tool has its own overhead and behavior:
- `ping` uses ICMP, which may be treated differently than TCP/UDP
- `traceroute` uses TTL-expired ICMP or UDP, which routers may rate-limit
- `iperf3` tests bulk throughput, which behaves differently from interactive traffic
- `curl` timing includes application processing, not just network delay

### Heisenberg problem: measuring changes the system

- Running a bandwidth test consumes bandwidth, potentially causing congestion
- Multiple ping processes increase ICMP traffic, potentially triggering rate limits
- Capturing packets with tcpdump adds CPU load, potentially increasing processing delay

### Time-of-day effects

Networks behave differently at different times:
- Business hours: more traffic, more congestion, higher loss
- Evenings: residential streaming traffic increases
- Weekends: different traffic patterns
- Maintenance windows: routing changes, brief outages

Always measure at multiple times and report the distribution, not a single point.

### Geographic effects

- Peering points between ISPs can be congested
- Traffic to different regions takes different paths
- CDN behavior varies by location
- International traffic goes through undersea cables (higher latency)

---

## Building a Network Health Dashboard

### What to track continuously

If you're responsible for a service, track these metrics:

```
1. Latency
   - Median RTT to key endpoints (P50)
   - Tail latency (P95, P99)
   - Latency over time (24h trend)

2. Jitter
   - RTT standard deviation
   - Max - Min RTT gap

3. Packet Loss
   - TCP retransmission rate (from /proc/net/snmp)
   - Connection establishment failure rate
   - DNS query timeout rate

4. Throughput
   - Interface utilization (% of capacity)
   - Application-level goodput
   - Per-connection throughput distribution

5. Errors
   - Interface error counters (ip -s link show)
   - TCP reset (RST) count
   - Connection timeout count
```

### Simple monitoring script

```bash
#!/bin/bash
# network-health.sh — Run every minute via cron

LOG="/var/log/network-health.log"
TARGET="8.8.8.8"
TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")

# Latency and loss
PING_RESULT=$(ping -c 10 -q $TARGET 2>/dev/null)
LOSS=$(echo "$PING_RESULT" | grep -oP '\d+(?=% packet loss)')
RTT_LINE=$(echo "$PING_RESULT" | grep 'rtt')
AVG_RTT=$(echo "$RTT_LINE" | awk -F'/' '{print $5}')
JITTER=$(echo "$RTT_LINE" | awk -F'/' '{print $7}' | tr -d ' ms')

# TCP retransmissions
RETRANS=$(nstat -a 2>/dev/null | grep TcpRetransSegs | awk '{print $2}')

# Interface errors
ERRORS=$(ip -s link show eth0 2>/dev/null | grep errors | head -1 | awk '{print $2}')

echo "$TIMESTAMP loss=$LOSS% avg_rtt=${AVG_RTT}ms jitter=${JITTER}ms retrans=$RETRANS iface_errors=$ERRORS" >> $LOG
```

```bash
# Add to crontab
# crontab -e
# * * * * * /path/to/network-health.sh
```

---

## Key Takeaways

1. **Jitter is variation in latency**, not latency itself — constant 100ms is better than variable 50ms for real-time apps
2. **Jitter is caused primarily by variable queuing** — traffic bursts at routers create inconsistent delay
3. **Packet loss of even 0.1% measurably impacts TCP throughput** — TCP's congestion response is aggressive
4. **Loss at intermediate hops in mtr may be a false alarm** — routers rate-limit ICMP but forward data normally
5. **Use tc netem to simulate bad networks** — critical for testing application resilience
6. **Measurements are imperfect** — they can interfere with the system and vary by time/location
7. **Monitor continuously, not just when things break** — baselines let you detect anomalies

---

## Next Module

→ [../03-physical-datalink/01-signals-and-encoding.md](../03-physical-datalink/01-signals-and-encoding.md) — Physical and Data Link layers: Signals, Ethernet, MAC, and ARP
