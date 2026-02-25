# Bandwidth vs Throughput — Why "Fast Internet" Is Ambiguous

> "I have 1 Gbps internet." Great — but what you actually get is throughput, not bandwidth. They're different. Let's understand why.

---

## Table of Contents

1. [Bandwidth: The Theoretical Maximum](#bandwidth-the-theoretical-maximum)
2. [Throughput: What You Actually Get](#throughput-what-you-actually-get)
3. [Why Throughput Is Always Less Than Bandwidth](#why-throughput-is-always-less-than-bandwidth)
4. [Goodput: The Only Number Users Care About](#goodput-the-only-number-users-care-about)
5. [The Bandwidth-Delay Product (BDP)](#the-bandwidth-delay-product-bdp)
6. [Bottleneck Links](#bottleneck-links)
7. [Measuring Bandwidth and Throughput on Linux](#measuring-bandwidth-and-throughput-on-linux)
8. [Common Myths and Misconceptions](#common-myths-and-misconceptions)
9. [Capacity Planning Basics](#capacity-planning-basics)

---

## Bandwidth: The Theoretical Maximum

### Definition

**Bandwidth** is the maximum rate at which data CAN be transmitted over a link.

It's measured in bits per second (bps), or more commonly:
- Kbps (kilobits per second) — $10^3$ bps
- Mbps (megabits per second) — $10^6$ bps
- Gbps (gigabits per second) — $10^9$ bps

**Critical unit confusion**: Bandwidth is in **bits**. File sizes are in **bytes**. 1 byte = 8 bits.

A "100 Mbps" connection can theoretically transfer:
$$100 \text{ Mbps} \div 8 = 12.5 \text{ MB/s}$$

This catches people off guard. "I have 100 Megabit internet but my download speed is only 12 megabytes per second!" — that's correct. That's the conversion.

### What determines bandwidth

**For wired links**: The bandwidth is determined by the physical standard and hardware:
- Fast Ethernet: 100 Mbps
- Gigabit Ethernet: 1 Gbps
- 10 Gigabit Ethernet: 10 Gbps
- 25/40/100/400 Gigabit Ethernet: Used in data centers

**For wireless links**: Bandwidth depends on:
- Available radio spectrum (frequency band and channel width)
- Modulation scheme (how much data per radio signal)
- Number of spatial streams (MIMO antennas)
- Signal quality (distance, interference)
- Shared with other users on the same access point

**For ISP connections**: Your "plan speed" is usually the bandwidth of your last-mile connection (your home to the ISP's equipment). The pipes inside the ISP and on the internet backbone have their own (usually much higher) bandwidth.

```bash
# Check your interface's bandwidth (link speed)
ethtool eth0 | grep Speed
# Output: Speed: 1000Mb/s → this is 1 Gbps Ethernet

# Check wireless link speed
iw dev wlan0 link
# Shows TX/RX bitrate
```

---

## Throughput: What You Actually Get

### Definition

**Throughput** is the actual rate at which data is successfully transferred over a period of time.

Throughput ≤ Bandwidth. Always. There is no scenario where throughput exceeds bandwidth.

### The pipeline analogy

Think of bandwidth as the **diameter** of a water pipe, and throughput as the **actual flow** of water through it.

- A 4-inch pipe (bandwidth) could carry 100 gallons/minute (theoretical max)
- But if there's sediment buildup, partial blockages, or low water pressure, actual flow (throughput) might be 60 gallons/minute
- And the water pressure is like window size / congestion window in TCP — it determines how hard you push

---

## Why Throughput Is Always Less Than Bandwidth

Multiple factors reduce throughput below the theoretical bandwidth:

### 1. Protocol overhead

Every packet carries headers that consume bandwidth but don't carry user data:

```
For a 1500-byte Ethernet frame:
  Ethernet header:  14 bytes
  IP header:        20 bytes
  TCP header:       32 bytes (with typical options)
  ═══════════════════════════
  Total overhead:   66 bytes
  Payload:          1434 bytes
  
  Efficiency: 1434/1500 = 95.6%
```

But there's more overhead you don't see:
- Ethernet **preamble** (8 bytes) and **inter-frame gap** (12 bytes): 20 bytes per frame
- Layer 1 overhead: 1500 + 14 + 4 (FCS) + 8 + 12 = 1538 bytes total on the wire
- Effective data per 1538 wire bytes: 1434 = 93.2% efficiency

For small packets (e.g., TCP ACKs at 54 bytes), overhead is massive — ACKs carry 0 payload bytes but consume wire time.

### 2. TCP congestion control

TCP doesn't blast data at full bandwidth. It:
1. **Starts slowly**: Slow start begins at a small congestion window (often 10 segments)
2. **Probes carefully**: Increases sending rate until it detects congestion
3. **Backs off on loss**: Cuts sending rate when packets are lost

For short transfers (small files, API calls), TCP may never reach full bandwidth because the transfer completes before slow start ramps up.

For long transfers, TCP eventually stabilizes near the available bandwidth, but congestion control introduces ongoing variation.

### 3. Packet loss forces retransmission

When a packet is lost, TCP must retransmit it. This wastes time and bandwidth:
- The lost packet's original bandwidth was wasted
- A retransmission consumes additional bandwidth
- Pausing for retransmission (and the timeout/fast retransmit delay) reduces effective throughput

Even 0.1% packet loss significantly impacts TCP throughput. The Mathis equation approximates TCP throughput:

$$\text{Throughput} \approx \frac{\text{MSS}}{RTT \times \sqrt{p}}$$

Where MSS is the maximum segment size, RTT is round-trip time, and $p$ is the loss rate.

With 1460-byte MSS, 50ms RTT, and 1% loss:
$$\text{Throughput} \approx \frac{1460 \times 8}{0.050 \times \sqrt{0.01}} = \frac{11680}{0.005} = 2.3 \text{ Mbps}$$

Even on a 1 Gbps link, 1% packet loss can reduce TCP throughput to just 2.3 Mbps. This is why packet loss is so devastating.

### 4. Receiver window limits

TCP flow control uses a **receive window** — the maximum amount of unacknowledged data the receiver will accept. If the receive window is smaller than the BDP (bandwidth-delay product), the sender must wait for acknowledgments before sending more, leaving the pipe partially empty.

### 5. Shared links

Your 1 Gbps link is shared with other traffic:
- Other devices on your network
- Other streams from the same device
- Background OS traffic (updates, cloud sync)
- Router control traffic

The bandwidth you see is **total capacity**, not **reserved for you**.

### 6. Queuing and buffering

Intermediate routers may queue your packets, introducing delay but not reducing bandwidth per se. However, the perceived throughput (bytes received per wall-clock second) decreases when packets are delayed.

---

## Goodput: The Only Number Users Care About

### Definition

**Goodput** is the rate at which **useful** application data is delivered.

```
Bandwidth:  Raw capacity of the link
Throughput: Actual data rate (including retransmissions, headers)
Goodput:    Application-level data rate (excluding all overhead)
```

Example:
- Link bandwidth: 100 Mbps
- Throughput (including headers): 90 Mbps
- Minus retransmissions: 88 Mbps
- Minus protocol headers: ~82 Mbps
- Minus TLS encryption overhead: ~80 Mbps
- **Goodput: ~80 Mbps** (the rate at which your application receives usable data)

When a user says "my download speed is 80 Mbps," they're describing goodput. When a network engineer says "the link capacity is 100 Mbps," they're describing bandwidth.

---

## The Bandwidth-Delay Product (BDP)

### What it is

The bandwidth-delay product is the amount of data that can be "in flight" (in the network) at any moment:

$$\text{BDP} = \text{Bandwidth} \times \text{RTT}$$

Think of it as the volume of the pipe:
- Bandwidth is the pipe's cross-section (how wide)
- RTT is the pipe's length (how long — round trip)
- BDP is the total volume of water (data) the pipe holds

### Why BDP matters

To fully utilize a high-bandwidth, high-latency link, TCP must keep BDP bytes "in flight." If TCP's congestion window is smaller than the BDP, the link is underutilized.

```
Example 1: Data center (high bandwidth, low latency)
  Bandwidth: 10 Gbps
  RTT: 0.2 ms
  BDP: 10,000,000,000 × 0.0002 = 2,000,000 bits = 250 KB
  
  TCP needs only 250 KB in flight → easy

Example 2: Transcontinental (high bandwidth, high latency)
  Bandwidth: 1 Gbps
  RTT: 100 ms
  BDP: 1,000,000,000 × 0.1 = 100,000,000 bits = 12.5 MB
  
  TCP needs 12.5 MB in flight → requires very large buffers
  
  With default 64 KB TCP window (no scaling):
  Throughput = 64 KB / 0.1 s = 640 KB/s = 5.12 Mbps
  On a 1 Gbps link! → 0.5% utilization!
```

This is why **TCP window scaling** (RFC 7323) exists — it allows windows up to 1 GB, enabling high-throughput transfers over high-latency links.

```bash
# Check if TCP window scaling is enabled
sysctl net.ipv4.tcp_window_scaling
# Should be 1 (enabled by default on modern Linux)

# Check current TCP buffer sizes
sysctl net.ipv4.tcp_rmem
sysctl net.ipv4.tcp_wmem
# Format: min default max (in bytes)
# Default max is often 4-6 MB — may need tuning for long fat pipes
```

### Long fat networks (LFNs)

A network with high bandwidth AND high latency is called a "long fat network" (LFN, pronounced "elephant"). LFNs are challenging because:
- BDP is very large → TCP needs huge windows
- Many packets are in flight → one lost packet means lots of data must wait
- Slow start takes many RTTs to ramp up to full speed
- Congestion control algorithms may oscillate

Satellite links (500ms+ RTT) and transcontinental fiber (100ms+ RTT) are classic LFNs.

---

## Bottleneck Links

### Every path has a bottleneck

When data flows from A to B through multiple links, each link has its own bandwidth. The maximum throughput of the entire path is limited by the **slowest link** — the bottleneck.

```mermaid
graph LR
    A["Your PC"] -->|"1 Gbps"| B["Router"]
    B -->|"100 Mbps"| C["ISP"]
    C -->|"10 Gbps"| D["Internet"]
    D -->|"10 Gbps"| E["Server DC"]
    
    style C fill:#ffcccc
```

In this path, the ISP link (100 Mbps) is the bottleneck. Even though your LAN and the backbone are much faster, your throughput cannot exceed 100 Mbps.

### Finding the bottleneck

```bash
# Method 1: iperf3 to various points
# Run iperf3 server on a remote machine
iperf3 -s  # on the server

# Test from your machine
iperf3 -c <server-ip> -t 10
# Shows: sender and receiver bandwidth
# If sender >> receiver, there's a bottleneck somewhere

# Method 2: traceroute + mtr to see where latency jumps
mtr --report -c 50 <destination>
# Large latency jumps between hops might indicate congestion

# Method 3: pathchar / pathrate (specialized tools)
# These estimate per-link bandwidth along a path
```

### When the bottleneck isn't the link

Sometimes the bottleneck isn't link bandwidth:
- **Server CPU**: The server can't generate data fast enough
- **Disk I/O**: The server can't read data from disk fast enough
- **Application logic**: The application processes requests slowly
- **TCP window**: The receiver window is too small for the BDP
- **Congestion window**: TCP hasn't ramped up enough

Always check: is the bottleneck in the network, or in the endpoints?

```bash
# Check if the sender's CPU is maxed out
top  # or htop

# Check disk I/O
iostat -x 1

# Check TCP socket buffer usage
ss -tm
# Shows: skmem (socket memory) — if send buffer is full, sender is waiting
```

---

## Measuring Bandwidth and Throughput on Linux

### Tool 1: iperf3 (the gold standard for throughput testing)

```bash
# Server side
iperf3 -s

# Client side — basic TCP test
iperf3 -c <server-ip>
# Default: 10 seconds, 1 stream

# Multiple parallel streams (better for high-BDP links)
iperf3 -c <server-ip> -P 4

# UDP test (measures bandwidth without TCP overhead)
iperf3 -c <server-ip> -u -b 100M
# -b 100M sets target bandwidth to 100 Mbps

# Reverse mode (server sends to client — tests download)
iperf3 -c <server-ip> -R

# Output explanation:
# [ ID] Interval       Transfer     Bitrate        Retr  Cwnd
# [  5]  0.00-10.00 sec   112 MBytes  94.0 Mbits/sec   3  254 KBytes
#
# Transfer: total data transferred
# Bitrate: average throughput
# Retr: TCP retransmissions (indicator of loss)
# Cwnd: TCP congestion window (how much TCP is willing to send)
```

### Tool 2: curl (application-level throughput)

```bash
# Download and measure speed
curl -o /dev/null -s -w "Speed: %{speed_download} bytes/sec\nSize: %{size_download} bytes\nTime: %{time_total}s\n" http://speedtest.example.com/large-file

# This measures GOODPUT (application-level data rate)
```

### Tool 3: ethtool (link speed)

```bash
# Check physical link capabilities
ethtool eth0
# Speed: 1000Mb/s  ← this is BANDWIDTH (link speed)
# This says nothing about actual throughput
```

### Tool 4: nload (real-time bandwidth monitor)

```bash
# Install: sudo apt install nload
nload eth0
# Shows real-time incoming/outgoing bandwidth usage
# Useful for seeing how much bandwidth is currently consumed
```

### Tool 5: bmon (bandwidth monitor)

```bash
# Install: sudo apt install bmon
bmon
# Shows per-interface bandwidth usage with graphs
```

---

## Common Myths and Misconceptions

### Myth 1: "Bandwidth = speed"

Bandwidth is **capacity**, not speed. An 8-lane highway (high bandwidth) can move more cars, but each car doesn't go faster. Speed in networking is latency. A 1 Gbps connection is not "faster" than a 100 Mbps connection if they have the same latency — it just has more capacity.

### Myth 2: "More bandwidth always makes things faster"

For downloads of large files: yes, more bandwidth helps.
For web browsing: often no — latency (RTTs) is the bottleneck.
For API calls: almost never — the data is small, latency dominates.

### Myth 3: "My speed test shows 500 Mbps, so I should get 500 Mbps to any server"

Speed tests measure throughput to a specific server, usually nearby and designed for high performance. Throughput to other servers depends on:
- The bottleneck link on the path to THAT server
- Server capacity
- Congestion along THAT specific route
- TCP performance over THAT specific RTT

### Myth 4: "If I use 10% of my bandwidth, I have 90% spare capacity"

Traffic is bursty. Average utilization of 10% might mean peaks of 80% occurring every few seconds. Those peaks cause queuing and latency spikes. 70% average utilization is generally the maximum recommended for a link before quality degrades.

### Myth 5: "Mbps and MB/s are the same"

- Mbps = megaBITS per second (network speed)
- MB/s = megaBYTES per second (file size per second)
- 1 MB/s = 8 Mbps

ISPs advertise in Mbps because the number is 8× larger and looks more impressive.

---

## Capacity Planning Basics

### How to estimate bandwidth needs

```
Per user/device:
  Web browsing:      1-5 Mbps
  Video streaming:   5-25 Mbps (depends on quality)
  Video conferencing: 2-8 Mbps (up + down)
  File transfers:    Depends on file size and frequency
  
For N concurrent users:
  Required bandwidth ≈ N × average_per_user × oversubscription_ratio
  
  Oversubscription ratio: 3:1 to 10:1
  (Not everyone uses max bandwidth simultaneously)
```

### When to worry about bandwidth

- Average link utilization consistently above 70%
- Peak utilization hitting 90%+ regularly
- Increasing packet loss correlating with high utilization
- Latency spikes during peak hours

```bash
# Monitor utilization over time
sar -n DEV 1 60
# Shows per-interface bytes/sec every 1 second for 60 samples

# Or use vnstat for historical data
vnstat -d  # daily summary
vnstat -h  # hourly summary
```

---

## Key Takeaways

1. **Bandwidth is capacity; throughput is actual rate; goodput is useful data rate** — three different things
2. **Throughput < bandwidth** due to overhead, congestion control, loss, and sharing
3. **BDP determines how much data TCP must keep in flight** — high-latency links need large TCP windows
4. **The bottleneck link limits entire path throughput** — upgrading non-bottleneck links doesn't help
5. **Packet loss dramatically reduces throughput** — even 0.1% loss can crush TCP performance
6. **Use iperf3 for throughput testing**, not speed test websites
7. **Bits vs bytes**: ISPs advertise bits (8× bigger number); divide by 8 for bytes
8. **More bandwidth doesn't always improve user experience** — often latency is the bottleneck

---

## Next

→ [03-jitter-loss-measurement.md](03-jitter-loss-measurement.md) — Jitter, packet loss, and how to measure what actually matters
