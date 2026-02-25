# Bufferbloat & Beyond — Modern Congestion Problems

> Classic congestion control assumed small buffers: when they overflow, packets drop, senders slow down. Modern networks have enormous buffers. Packets don't drop — they queue. Latency skyrockets. This is bufferbloat, and it's everywhere.

---

## Table of Contents

1. [What Is Bufferbloat?](#what-is-bufferbloat)
2. [Why Buffers Grew](#why-buffers-grew)
3. [The Symptoms](#symptoms)
4. [Measuring Bufferbloat](#measuring)
5. [Active Queue Management (AQM)](#aqm)
6. [RED: Random Early Detection](#red)
7. [CoDel: Controlled Delay](#codel)
8. [fq_codel: Fair Queuing + CoDel](#fq-codel)
9. [CAKE: Common Applications Kept Enhanced](#cake)
10. [ECN: Explicit Congestion Notification](#ecn)
11. [Linux: Diagnosing and Fixing](#linux-fixing)
12. [Key Takeaways](#key-takeaways)

---

## What Is Bufferbloat?

Bufferbloat is **excessive latency caused by oversized network buffers**.

Routers and switches have buffers to absorb short traffic bursts. This is good — without buffers, any momentary spike drops packets. But when buffers are too large:

```
Small buffer router:
  Burst arrives → buffer absorbs some → drops excess → sender detects loss → slows down
  Latency: low (buffer drains quickly)
  
Huge buffer router:
  Burst arrives → buffer absorbs ALL of it → no drops!
  More traffic arrives → buffer keeps growing → still no drops
  Sender thinks everything is fine → keeps sending at full speed
  Latency: catastrophic (packets wait in queue for seconds)
```

### A concrete example

Home router with a 1 Mbps upload link and 128 KB buffer:

```
Buffer capacity at 1 Mbps:
  128 KB × 8 / 1,000,000 bps = 1.024 seconds of buffering

Someone starts uploading a large file (saturates the link):
  Packets queue up → buffer holds 1 second of data
  
Now you try to load a webpage:
  Your DNS query enters the buffer... behind 1 second of upload data
  Your SYN packet enters the buffer... behind 1+ second of data
  Every packet waits 1+ second in the queue
  
  Perceived latency: 1000ms+ (instead of normal 20ms)
  Gaming: unplayable (1000ms ping)
  VoIP: broken (1000ms delay makes conversation impossible)
```

---

## Why Buffers Grew

### Memory got cheap

In 1988 (when TCP congestion control was invented), router RAM was expensive. Buffers were small. Drops happened quickly, and TCP's loss-based congestion control worked perfectly.

By 2010, RAM was cheap. Equipment vendors added huge buffers because "dropped packets are bad" — a reasonable-sounding but ultimately harmful design choice.

### The BDP rule (misapplied)

The conventional wisdom was:

$$\text{Buffer size} = \text{Bandwidth} \times \text{RTT}$$

For a 1 Gbps link with 100ms RTT: buffer = 12.5 MB. This is an enormous buffer that can add 100ms of queuing delay when full.

Research by Appenzeller et al. (2004) showed that with many flows, you only need:

$$\text{Buffer size} = \frac{\text{Bandwidth} \times \text{RTT}}{\sqrt{N}}$$

Where $N$ is the number of flows. With 10,000 flows: buffer = 125 KB instead of 12.5 MB.

### Home routers are the worst

Consumer routers often have buffers measured in seconds, not milliseconds:

```
Cable modem with 20 Mbps upload, 10 MB buffer:
  Buffer delay = 10 MB × 8 / 20 Mbps = 4 seconds!
  
WiFi AP with 50 Mbps, 5 MB buffer:
  Buffer delay = 5 MB × 8 / 50 Mbps = 0.8 seconds
```

---

## Symptoms

### How to recognize bufferbloat

1. **High latency under load**: Ping goes from 20ms to 500ms+ when someone starts a download
2. **VoIP degrades during transfers**: Voice breaks up when uploading/downloading
3. **Gaming lag during downloads**: Ping spikes make games unplayable
4. **Web pages feel slow during transfers**: Even on "fast" connections

### The paradox

Bufferbloat is counterintuitive: your connection has plenty of bandwidth, but latency is terrible. Users complain "the internet is slow" when really it's "the internet has high latency because packets are stuck in a queue."

```
Without bufferbloat:
  Speed test: 100 Mbps ✓    Latency: 15ms ✓    Quality: Excellent

With bufferbloat:
  Speed test: 100 Mbps ✓    Latency: 800ms ✗    Quality: Terrible
  
  The speed test looks fine because it measures throughput (bytes/second).
  But interactive traffic is destroyed by the huge queuing delay.
```

---

## Measuring

### Quick test

```bash
# Baseline latency (unloaded)
ping -c 10 8.8.8.8
# Average: 15ms

# Now saturate the link while pinging
# Terminal 1: saturate upload
iperf3 -c speedtest.server.com -t 30 -R  # or just start a large upload

# Terminal 2: measure loaded latency
ping -c 30 8.8.8.8
# Average: 350ms  ← bufferbloat!

# The difference is the bufferbloat:
# 350ms - 15ms = 335ms of queuing delay
```

### Using flent (recommended)

```bash
# Install flent (FLExible Network Tester)
sudo apt install flent

# RRUL test: simultaneously uploads, downloads, and measures latency
flent rrul -p all_scaled -l 60 -H speedtest.server.com -o bufferbloat_test.png

# This produces a graph showing:
# - Download throughput
# - Upload throughput  
# - Latency under load (the bufferbloat indicator)
```

### DSLReports / Waveform Speed Test

Waveform's speed test (https://www.waveform.com/tools/bufferbloat) specifically tests for bufferbloat with grades A-F.

---

## AQM

**Active Queue Management** is the solution to bufferbloat. Instead of waiting for the buffer to overflow (tail drop), AQM algorithms **proactively** manage queue size.

### The principle

```
Tail drop (no AQM):
  Queue:  [pkt][pkt][pkt][pkt][pkt][pkt][pkt]FULL → drop new packets
  Problem: Buffer fills completely before any signal → high latency
  
AQM:
  Queue:  [pkt][pkt][pkt][pkt] → "queue is getting large, start dropping/marking"
  Result: Senders detect loss early → slow down before buffer is full → low latency
```

---

## RED

**Random Early Detection** (1993) was the first widely deployed AQM.

### How it works

```
              min_th    max_th    buffer_max
                │         │         │
Queue: [  low   │ medium  │  high   │ overflow ]
                │         │         │
Action:  accept  │ random   │ drop   │ drop all
         all    │ drop     │ all    │
                │ (increasing│       │
                │  probability)     │
```

1. Queue < min_threshold → accept all packets
2. Queue between min and max threshold → drop packets randomly with increasing probability
3. Queue > max threshold → drop all new packets

### Why random?

If you drop deterministically (e.g., every Nth packet), all flows synchronize their retransmissions, causing oscillation. Random dropping ensures different flows back off at different times →global synchronization avoided.

### RED's problems

RED has 5 tuning parameters (min_th, max_th, max_p, weight, gentle mode). Getting them right is hard. Wrong parameters make RED worse than tail drop. RED was rarely tuned properly in practice.

---

## CoDel

**CoDel (Controlled Delay)** — pronounced "coddle" — was designed by Kathleen Nichols and Van Jacobson (2012) to be a **parameter-free** AQM.

### Key insight

CoDel doesn't look at queue LENGTH. It looks at **how long packets spend in the queue** (sojourn time).

Why? Queue length is meaningless without knowing the link speed. 100 packets in a 10 Gbps queue drain in microseconds. 100 packets in a 1 Mbps queue take nearly a second.

### How CoDel works

```
TARGET = 5ms          # acceptable queuing delay
INTERVAL = 100ms      # measurement window

Every INTERVAL:
  If min(sojourn_time in last INTERVAL) > TARGET:
    Start dropping packets
    Drop interval decreases: 100ms, 70ms, 57ms, 50ms, ...
    (1/√n progression — starts gentle, gets aggressive)
  
  If min(sojourn_time in last INTERVAL) < TARGET:
    Stop dropping
```

### Why minimum sojourn time?

Using the minimum (not average) filters out bursts:

```
A short burst arrives → queue spikes → sojourn time spikes
But the minimum sojourn time over the interval stays low
CoDel doesn't drop → correct! Burst is temporary.

Persistent overload → queue stays long → minimum sojourn time stays high
CoDel starts dropping → correct! This is real congestion.
```

### CoDel on Linux

```bash
# Apply CoDel to an interface
sudo tc qdisc replace dev eth0 root codel

# With parameters (rarely needed — defaults are good)
sudo tc qdisc replace dev eth0 root codel target 5ms interval 100ms

# View statistics
tc -s qdisc show dev eth0
# Shows: dropped, overlimit, requeues, delay, count
```

---

## fq_codel

**fq_codel = Fair Queuing + CoDel**. This is the **recommended AQM** and the default qdisc on many Linux distributions since kernel 3.6.

### Why combine FQ and CoDel?

CoDel alone treats all traffic the same. A single bulk download fills the queue and increases latency for interactive flows (SSH, gaming, VoIP).

Fair Queuing creates **separate queues per flow** (identified by 5-tuple hash). Each queue gets equal service. CoDel manages each queue independently.

```
Without FQ:
  [bulk download][bulk][DNS][bulk][SSH keystroke][bulk][bulk]
  DNS and SSH wait behind bulk download packets → high latency

With FQ (1024 queues, flows hashed to queues):
  Queue 1 (download): [bulk][bulk][bulk][bulk][bulk][bulk]
  Queue 2 (DNS):      [DNS query]
  Queue 3 (SSH):      [keystroke]
  Queue 4 (VoIP):     [audio packet]
  
  Round-robin between non-empty queues:
  DNS → SSH → VoIP → bulk → DNS → SSH → VoIP → bulk → ...
  
  Interactive flows get served immediately → low latency
  Bulk flow gets most of the bandwidth (queue is longer)
```

### fq_codel on Linux

```bash
# fq_codel is often the default qdisc
tc qdisc show dev eth0
# qdisc fq_codel 0: root refcnt 2 limit 10240p flows 1024 ...

# Explicitly set it
sudo tc qdisc replace dev eth0 root fq_codel

# View statistics
tc -s qdisc show dev eth0
# Shows: packets sent, dropped, overlimit, new_flows, old_flows

# Key parameters:
# flows 1024     ← number of flow queues  
# target 5ms     ← CoDel target delay
# interval 100ms ← CoDel measurement interval
# quantum 1514   ← bytes served per queue per round
```

---

## CAKE

**CAKE (Common Applications Kept Enhanced)** is the most advanced qdisc, combining fair queuing, AQM, traffic shaping, and prioritization.

### What CAKE adds over fq_codel

1. **Built-in shaping**: Rate-limit to your actual link speed (essential for fixing bufferbloat at the ISP modem)
2. **Per-host fairness**: Fair sharing between hosts, not just flows (so one host with 100 flows doesn't get 100x more than a host with 1 flow)
3. **DSCP-based prioritization**: Marks different traffic classes (voice, video, bulk)
4. **Overhead compensation**: Accounts for link-layer overhead (ATM, PPPoE)

### CAKE for home network bufferbloat fix

```bash
# Install CAKE (if not in kernel)
sudo apt install linux-modules-extra-$(uname -r)
sudo modprobe sch_cake

# Apply CAKE with shaping to your upload speed
# (set to ~85-90% of your actual upload speed)
sudo tc qdisc replace dev eth0 root cake bandwidth 18mbit

# For DOCSIS cable (adds per-host fairness, overhead compensation)
sudo tc qdisc replace dev eth0 root cake bandwidth 18mbit docsis nat

# For PPPoE (DSL)
sudo tc qdisc replace dev eth0 root cake bandwidth 18mbit overhead 22 pppoe-ptm

# View statistics
tc -s qdisc show dev eth0
```

### Why shape below actual link speed?

If you don't shape, the bottleneck queue is in your ISP's modem — which you can't control. By shaping your traffic to 85-90% of link speed, the bottleneck moves to YOUR router where CAKE manages it:

```
Without shaping:
  Your router → 20 Mbps → ISP modem (bloated buffer) → ISP
  Bottleneck: ISP modem (uncontrolled) → bufferbloat!

With shaping at 18 Mbps:
  Your router (CAKE at 18 Mbps) → ISP modem → ISP
  Bottleneck: Your router (CAKE manages queue) → no bufferbloat!
```

---

## ECN

**Explicit Congestion Notification** allows routers to signal congestion WITHOUT dropping packets.

### How ECN works

Two bits in the IP header (ECN field):

| ECN bits | Meaning |
|----------|---------|
| 00 | Not ECN-capable |
| 01 or 10 | ECN-capable transport |
| 11 | **Congestion Experienced (CE)** — router marks this when queue is growing |

```
Normal (without ECN):
  Router queue growing → drops packet → sender retransmits → wasted bandwidth

With ECN:
  Router queue growing → marks packet with CE bit → packet delivered normally
  Receiver sees CE → sets ECE flag in TCP ACK
  Sender sees ECE → reduces cwnd (as if loss occurred) + sets CWR flag
  Result: congestion controlled WITHOUT packet loss
```

### ECN on Linux

```bash
# Check ECN status
cat /proc/sys/net/ipv4/tcp_ecn
# 0 = disabled
# 1 = enabled (request ECN in outgoing connections)
# 2 = server-side only (respond to ECN but don't request)

# Enable ECN
sudo sysctl -w net.ipv4.tcp_ecn=1

# Verify ECN negotiation on connections
ss -ti | grep ecn
# Should show "ecn" for connections that negotiated it
```

### ECN limitations

- Both endpoints AND all routers on path must support it
- Some middleboxes (firewalls) strip ECN bits → negotiation fails
- Not universally deployed (improving but not ubiquitous)
- Still needs AQM at routers to decide when to mark (ECN without AQM is useless)

---

## Linux: Diagnosing and Fixing

### Check current qdisc

```bash
# View qdisc (queuing discipline) on all interfaces
tc qdisc show

# Detailed statistics
tc -s qdisc show dev eth0

# Common output:
# qdisc fq_codel 0: root refcnt 2 limit 10240p flows 1024
#  Sent 12345678 bytes 9876 pkt (dropped 42, overlimits 0 requeues 5)
#  backlog 0b 0p requeues 5
```

### Test for bufferbloat on your connection

```bash
# Method 1: ping under load
# Terminal 1:
ping -c 60 1.1.1.1 > idle_ping.txt

# Terminal 2: generate load
iperf3 -c speedtest.server.com -t 30

# Terminal 1 (simultaneously):
ping -c 60 1.1.1.1 > loaded_ping.txt

# Compare:
echo "Idle:" && grep "avg" idle_ping.txt
echo "Loaded:" && grep "avg" loaded_ping.txt
# If loaded avg >> idle avg → bufferbloat
```

### Fix bufferbloat on your Linux router

```bash
# Option 1: fq_codel (good, no configuration needed)
sudo tc qdisc replace dev eth0 root fq_codel

# Option 2: CAKE (best, needs shaping)
sudo tc qdisc replace dev eth0 root cake bandwidth 90mbit

# For upload direction (egress):
sudo tc qdisc replace dev wan0 root cake bandwidth 18mbit wash nat

# For download direction (ingress — requires ifb device):
sudo ip link add name ifb0 type ifb
sudo ip link set ifb0 up
sudo tc qdisc replace dev wan0 handle ffff: ingress
sudo tc filter add dev wan0 parent ffff: protocol all u32 match u32 0 0 action mirred egress redirect dev ifb0
sudo tc qdisc replace dev ifb0 root cake bandwidth 90mbit wash nat

# Persist across reboots by adding to /etc/network/interfaces or a script
```

### Monitor queue health

```bash
# Watch queue statistics in real time
watch -n 1 'tc -s qdisc show dev eth0'

# Key metrics:
# dropped:    packets dropped by AQM (some drops are intentional → good)
# overlimits: times rate exceeded target
# backlog:    current queue depth in bytes and packets
#             High backlog = potential bufferbloat

# Check network interface queue length
ip link show eth0 | grep qlen
# txqueuelen 1000 (default, rarely needs changing with AQM)
```

---

## Key Takeaways

1. **Bufferbloat** = excessive latency from oversized buffers. Not a capacity problem — a latency problem.
2. **You can have high bandwidth AND terrible latency** if buffers are too large
3. **Active Queue Management (AQM)** proactively manages queues instead of waiting for overflow
4. **CoDel** measures sojourn time, not queue length — parameter-free and effective
5. **fq_codel** (Fair Queuing + CoDel) is the recommended default — gives interactive traffic low latency
6. **CAKE** is the most complete solution — with built-in shaping, per-host fairness, and prioritization
7. **Shape to ~85-90%** of link speed to move the bottleneck queue to a device you control
8. **ECN** signals congestion without dropping packets — better but requires universal support
9. **Test with ping under load** to detect bufferbloat on your connection
10. **BBR + fq qdisc** together are the best server-side configuration for avoiding bufferbloat

---

## Next

→ [../08-dns/01-why-dns-exists.md](../08-dns/01-why-dns-exists.md) — The naming system of the Internet
