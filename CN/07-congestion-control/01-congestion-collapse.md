# Congestion Collapse — When the Network Eats Itself

> In 1986, the Internet nearly died. Throughput between UC Berkeley and Lawrence Berkeley Lab dropped from 32 Kbps to 40 bps — a 99.9% collapse. Van Jacobson diagnosed the problem and invented TCP congestion control. This is the story of why congestion control exists.

---

## Table of Contents

1. [The Problem](#the-problem)
2. [What Congestion Is](#what-congestion-is)
3. [The 1986 Collapse](#the-1986-collapse)
4. [Why Congestion Cascades](#congestion-cascades)
5. [Congestion Collapse Explained](#congestion-collapse-explained)
6. [The Tragedy of the Commons](#tragedy)
7. [Why Flow Control Doesn't Solve This](#flow-control-insufficient)
8. [The Solution: Congestion Control](#the-solution)
9. [Detecting Congestion](#detecting-congestion)
10. [Linux: Observing Congestion](#linux-observing)
11. [Key Takeaways](#key-takeaways)

---

## The Problem

Imagine a highway where everyone drives as fast as possible with no traffic rules:

```
Normal load:
  ═══════════════════════════════════════
  Cars flowing at 100 km/h, highway works fine

Heavy load:
  ══╗═╗╗═══╗═╗╗══╗════╗═╗═╗══╗═╗═╗═══
  More cars → some slowdown, but traffic flows

Overload:
  ╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗╗
  Gridlock. Nobody moves. Adding more cars makes it WORSE.
```

Networks have the same problem. Routers have finite buffer space. When too many senders transmit too fast, router queues fill up, packets are dropped, senders retransmit (adding MORE traffic), queues fill more, more drops, more retransmissions → **positive feedback loop**.

---

## What Congestion Is

**Congestion occurs when the arrival rate at a link/router exceeds its processing/forwarding rate.**

```
                    ┌─────────────┐
 100 Mbps ────────►│             │
 100 Mbps ────────►│   Router    │────────► 100 Mbps (output link)
 100 Mbps ────────►│             │
                    └─────────────┘

Input: 300 Mbps    Output: 100 Mbps    Excess: 200 Mbps

Where does the excess go?
1. Queue in the router's buffer (adds latency)
2. Once buffer is full → DROP packets
```

### Congestion is NOT the same as high utilization

```
Good:  High utilization, low delay        ← near capacity, no build up
Bad:   High utilization, increasing delay  ← queues building
Awful: Queue overflow, dropping, retransmitting ← congestion collapse
```

The relationship between offered load and throughput:

```
Throughput
    │
    │           ╱─────────── Ideal (throughput = load)
    │         ╱ 
    │       ╱  ╱──── Reality (with congestion control)
    │     ╱  ╱
    │   ╱ ╱
    │ ╱╱        ╲───── Without congestion control
    │╱             ╲        (congestion collapse!)
    ├─────────────────────── Offered Load
    0         Capacity
```

Without congestion control, throughput eventually drops to near zero as load increases past capacity. **The network does MORE work but delivers LESS useful data.**

---

## The 1986 Collapse

### What happened

- October 1986, UC Berkeley to Lawrence Berkeley Lab, ~400 meters apart
- Normal throughput: 32 Kbps
- During congestion: 40 bps (bits, not kilobits)
- **800x throughput reduction**

### The root cause

TCP had no congestion control. Every sender would:
1. Send as fast as the receiver window allowed
2. If packets were dropped (due to router congestion), retransmit immediately
3. Retransmissions added MORE traffic to the already congested link
4. More congestion → more drops → more retransmissions → more congestion

The network was full of retransmitted packets that were also being retransmitted. Useful throughput approached zero while total traffic (mostly useless retransmissions) was at maximum.

### Van Jacobson's insight

Van Jacobson realized that TCP needed to:
1. **Not immediately send at full speed** — start slow and probe
2. **Back off when congestion is detected** — reduce sending rate on packet loss
3. **Fairly share bandwidth** — each connection should converge to a fair share

He published the seminal paper "Congestion Avoidance and Control" (1988), which introduced:
- Slow start
- Congestion avoidance
- Fast retransmit
- Fast recovery

These algorithms saved the Internet.

---

## Congestion Cascades

### Why one drop leads to many

```
Step 1: Router R1's buffer fills up
  → Drops packet from Connection A

Step 2: Connection A doesn't receive ACK within timeout
  → Retransmits the dropped packet
  → This retransmission is now ADDITIONAL traffic on the already congested path

Step 3: The retransmission might also be dropped
  → Connection A retransmits AGAIN (with exponential backoff)
  → Meanwhile, the original packet might have been delivered but the ACK was lost

Step 4: Other connections (B, C, D...) are also experiencing drops
  → They're ALL retransmitting
  → Total traffic = original traffic + ALL retransmissions

Step 5: Router is now handling 3x the traffic (originals + retransmissions)
  → Drops even more aggressively
  → Cycle repeats
```

### The math of collapse

```
Let offered load = L (original data)
Let loss rate = p
Each lost packet generates 1 retransmission (on average)
Retransmission also has loss rate p

Total traffic = L + Lp + Lp² + Lp³ + ...
             = L / (1 - p)

If p = 50% (severe congestion):
  Total traffic = L / 0.5 = 2L
  But only L/(1-p) = L of useful data gets through... wait, no.
  
  Goodput (useful throughput) = L × (1-p) = 0.5L
  Total traffic (counting retransmissions) = L + 0.5L + 0.25L + ... ≈ 2L
  
  The network carries 2L traffic but only delivers 0.5L useful data.
  Efficiency = 25%. Getting worse...

If p = 90% (extreme congestion):
  Goodput = 0.1L
  Total traffic ≈ 10L
  Efficiency = 1%. This is congestion collapse.
```

---

## Congestion Collapse Explained

Congestion collapse occurs when the goodput (useful throughput) decreases as the offered load increases past capacity.

```
                  Congestion Collapse
                        ║
                        ▼
    │         Capacity──┐
    │                   │╲
    │                   │  ╲
  G │                  ╱│    ╲
  o │                ╱  │      ╲
  o │              ╱    │        ╲───── goodput falls
  d │            ╱      │              as load increases!
  p │          ╱        │
  u │        ╱          │
  t │      ╱            │
    │    ╱              │
    │  ╱                │
    │╱                  │
    └───────────────────┼──────────────►
                     Capacity    Offered Load
```

### Two types of congestion collapse

**1. Classic collapse (retransmission-based)**: Network full of retransmitted packets. Original packets already expired or will be retransmitted. 90%+ of bandwidth wasted on useless retransmissions.

**2. Undelivered packets**: Packets traverse multiple hops but are dropped at the last hop. All the bandwidth used for the first N-1 hops is wasted.

```
Packet: A → R1 → R2 → R3 → R4 → B
                              ↑
                          Dropped here!
                          
Bandwidth used on R1→R2 and R2→R3 links: WASTED
The packet consumed resources but delivered zero value
```

---

## Tragedy

Congestion is a **tragedy of the commons** problem.

Each individual TCP connection benefits from sending faster. But if ALL connections send faster, everyone suffers.

```
Game theory:
  - If I send fast and others send slow → I get high throughput ✓
  - If everyone sends fast → everyone gets near-zero throughput ✗
  - If everyone cooperates → everyone gets fair share ≈

Without enforcement (congestion control):
  → Nash equilibrium: everyone sends fast → everyone loses
  
With enforcement (TCP congestion control):
  → Each connection voluntarily reduces speed when congestion is detected
  → Shared resources are used efficiently
```

### Why UDP complicates this

UDP has no congestion control. A UDP sender can blast at maximum speed regardless of network conditions. If UDP traffic is significant:

```
TCP connections: detect congestion, back off politely
UDP streams: don't detect congestion, keep sending at full speed
Result: UDP pushes TCP out, TCP connections get almost nothing

This is why QoS (Quality of Service) mechanisms exist in routers:
  → Rate-limit UDP to prevent it from starving TCP
  → But this isn't universally deployed
```

QUIC (built on UDP) implements its own congestion control, so it behaves fairly alongside TCP.

---

## Flow Control Insufficient

A common misconception: "Flow control prevents congestion."

**No.** Flow control and congestion control solve different problems:

```
Flow control:  "Sender, I can't process data this fast"
               → Receiver tells sender to slow down
               → Protects the RECEIVER

Congestion control: "The NETWORK can't handle this much traffic"
                    → Network signals (via drops/ECN) that there's too much
                    → Protects the NETWORK

You can have:
  - Flow control fine + congestion problem: 
    Receiver has lots of buffer, but the network path is congested
  
  - Congestion fine + flow control problem:
    Network has plenty of capacity, but receiver can't keep up
```

Even if both endpoints have unlimited buffers, the network between them can still be congested.

---

## The Solution

Congestion control has three requirements:

1. **Detect congestion**: Know when the network is overloaded
2. **React to congestion**: Reduce sending rate when congestion is detected
3. **Probe for bandwidth**: Carefully increase sending rate when congestion clears

This is implemented through the **congestion window (cwnd)** — an internal variable the sender maintains, separate from the receiver's `rwnd`.

$$\text{Sending rate} \leq \frac{\min(\text{cwnd}, \text{rwnd})}{\text{RTT}}$$

Congestion control algorithms (covered in the next file):
- **Slow start**: Start small, grow exponentially
- **Congestion avoidance**: Grow linearly once near capacity
- **Fast recovery**: React intelligently to loss without starting over

---

## Detecting Congestion

TCP uses two signals to detect congestion:

### 1. Packet loss (timeout)

If no ACK arrives within the RTO, the packet is assumed lost. Loss strongly suggests congestion (router dropped the packet because its queue was full).

```
RTO timeout → severe congestion signal → drastic rate reduction
```

### 2. Duplicate ACKs (fast retransmit)

Three duplicate ACKs suggest a packet was lost but subsequent packets got through. This is a **milder** congestion signal — the network is still delivering packets, just some are being dropped.

```
3 duplicate ACKs → moderate congestion signal → moderate rate reduction
```

### 3. ECN (Explicit Congestion Notification)

Instead of dropping packets, routers can mark them with an ECN flag: "This packet made it through, but I'm getting congested — please slow down."

```bash
# ECN support on Linux
cat /proc/sys/net/ipv4/tcp_ecn
# 0 = disabled
# 1 = enabled (negotiate with peers)
# 2 = server-only (respond to ECN requests but don't initiate)

# Enable ECN
sudo sysctl -w net.ipv4.tcp_ecn=1
```

ECN is better than packet drops because no data is lost, but it requires support from both endpoints AND all routers on the path.

---

## Linux: Observing

### Check which congestion control algorithm is in use

```bash
# Current default
cat /proc/sys/net/ipv4/tcp_congestion_control
# cubic (or bbr, reno, etc.)

# Available algorithms
cat /proc/sys/net/ipv4/tcp_available_congestion_control
# reno cubic

# Change default
sudo sysctl -w net.ipv4.tcp_congestion_control=bbr

# Enable BBR (needs module)
sudo modprobe tcp_bbr
sudo sysctl -w net.ipv4.tcp_congestion_control=bbr
```

### Monitor congestion metrics per connection

```bash
# ss -ti shows congestion window and other metrics
ss -ti dst google.com

# Key fields:
# cwnd:10    ← congestion window (in segments, multiply by MSS for bytes)
# ssthresh:7 ← slow-start threshold (when to switch from slow start to avoidance)
# rtt:25/5   ← smoothed RTT / RTT variance (ms)
# retrans:0/0 ← current / total retransmissions
# lost:0     ← lost segments detected
# bytes_sent:1234  ← total bytes sent
# bytes_acked:1000 ← total bytes acknowledged
```

### Watching congestion in real time

```bash
# Watch cwnd changes during a transfer
# Terminal 1: Start iperf3 server
iperf3 -s

# Terminal 2: Run client  
iperf3 -c server_ip -t 60

# Terminal 3: Watch congestion window
watch -n 0.5 "ss -ti dst server_ip | grep -E 'cwnd|ssthresh|rtt|retrans'"

# You'll see cwnd grow during slow start, stabilize during congestion avoidance,
# and drop sharply when congestion (loss) is detected
```

### Network statistics

```bash
# TCP statistics including congestion events
nstat -az | grep -i "tcp" | grep -iE "loss|retrans|abort|overflow|prune|collapse"

# Key metrics:
# TcpRetransSegs      ← segments retransmitted (congestion indicator)
# TcpLossProbes       ← loss probes sent
# TCPLostRetransmit   ← retransmissions also lost (severe congestion)
# TCPFastRetrans      ← fast retransmissions (3 dup ACKs)
# TCPSlowStartRetrans ← retransmissions during slow start
# TCPSACKReneging     ← receiver withdrew previously SACKed data
```

---

## Key Takeaways

1. **Congestion collapse** = network carries maximum traffic but delivers near-zero useful data
2. **The 1986 incident** nearly killed the early Internet — Van Jacobson saved it with congestion control
3. **Root cause**: Retransmissions add traffic to an already congested network → positive feedback loop
4. **Flow control ≠ congestion control**: One protects the receiver, the other protects the network
5. **Tragedy of the commons**: Without congestion control, every sender benefits from sending faster, but everyone suffers
6. **Congestion signals**: Packet loss (RTO timeout, 3 dup ACKs) or ECN marking
7. **cwnd** is the sender's internal congestion window — limits how much data can be in flight
8. **`ss -ti`** shows cwnd, ssthresh, retrans, rtt — essential for diagnosing congestion

---

## Next

→ [02-aimd-slow-start-cwnd.md](02-aimd-slow-start-cwnd.md) — The algorithms that prevent congestion collapse
