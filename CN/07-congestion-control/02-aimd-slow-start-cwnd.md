# AIMD, Slow Start & cwnd — The Algorithms That Prevent Collapse

> TCP's congestion control is one of the most successful distributed algorithms ever deployed. Billions of devices independently decide how fast to send, and the Internet works. This file explains exactly how.

---

## Table of Contents

1. [The Congestion Window (cwnd)](#cwnd)
2. [Slow Start](#slow-start)
3. [Congestion Avoidance (AIMD)](#aimd)
4. [ssthresh: The Mode Switch](#ssthresh)
5. [Fast Retransmit & Fast Recovery](#fast-retransmit-recovery)
6. [Putting It All Together](#putting-it-together)
7. [TCP Reno vs Tahoe](#reno-vs-tahoe)
8. [TCP CUBIC (Linux Default)](#cubic)
9. [TCP BBR (Google)](#bbr)
10. [Comparing Algorithms](#comparing)
11. [Linux: Choosing and Tuning](#linux-tuning)
12. [Key Takeaways](#key-takeaways)

---

## cwnd

The **congestion window (cwnd)** is a variable maintained by the sender that limits how much data can be unacknowledged (in flight) at any time.

$$\text{Bytes in flight} \leq \min(\text{cwnd}, \text{rwnd})$$

Unlike `rwnd` (advertised by the receiver), `cwnd` is **never sent on the wire**. It's purely internal to the sender's TCP stack. The receiver doesn't know it exists.

```
cwnd starts small and grows as the sender probes for available bandwidth:

  Time 0:   cwnd = 1 MSS  → can send 1 segment
  Time 1:   cwnd = 2 MSS  → can send 2 segments at once
  Time 2:   cwnd = 4 MSS  → can send 4 segments
  ...
```

### Initial window (IW)

Modern Linux kernels use an initial window of **10 MSS** (14,600 bytes for typical Ethernet):

```bash
# Check initial congestion window
ss -ti | grep initcwnd
# Or: ip route show | grep initcwnd

# Set custom initial window
sudo ip route change default via 192.168.1.1 initcwnd 10

# Why 10? RFC 6928 (2013) found that IW=10 reduces page load time
# significantly compared to the original IW=1 (from 1988)
```

---

## Slow Start

### The idea

Don't send at full speed immediately. Start with a small window and grow **exponentially** until you either find the capacity or detect congestion.

Despite its name, slow start is actually **fast** — it doubles cwnd every RTT.

### The algorithm

```
On each ACK received:
  cwnd = cwnd + 1 MSS
```

This looks linear, but it's exponential because each RTT produces cwnd ACKs:

```
RTT 1: cwnd = 1    → send 1, get 1 ACK   → cwnd = 2
RTT 2: cwnd = 2    → send 2, get 2 ACKs  → cwnd = 4
RTT 3: cwnd = 4    → send 4, get 4 ACKs  → cwnd = 8
RTT 4: cwnd = 8    → send 8, get 8 ACKs  → cwnd = 16
RTT 5: cwnd = 16   → send 16             → cwnd = 32
...

After 10 RTTs: cwnd = 1024 MSS ≈ 1.5 MB
```

### Visualization

```
cwnd
(MSS)
  │
32│                                         ●
  │                                       ╱
16│                                    ●
  │                                  ╱
 8│                              ●
  │                            ╱
 4│                        ●
  │                      ╱
 2│                  ●
  │                ╱
 1│            ●
  │
  └────┬────┬────┬────┬────┬────┬────►
       1    2    3    4    5    6    RTT
       
  Slow start: exponential growth (doubles every RTT)
```

### When slow start ends

Slow start continues until one of:
1. `cwnd >= ssthresh` → switch to congestion avoidance (linear growth)
2. Packet loss detected → reset cwnd, update ssthresh
3. `cwnd >= rwnd` → limited by receiver, stop growing

---

## AIMD

AIMD = **Additive Increase, Multiplicative Decrease**. This is the congestion avoidance algorithm.

### Additive Increase (AI)

Once `cwnd >= ssthresh`, switch from exponential to linear growth:

```
On each ACK received:
  cwnd = cwnd + MSS × (MSS / cwnd)

Intuition: increase cwnd by 1 MSS per RTT (not per ACK)
```

```
RTT 1: cwnd = 16   → send 16, get 16 ACKs → cwnd = 17
RTT 2: cwnd = 17   → send 17, get 17 ACKs → cwnd = 18
RTT 3: cwnd = 18   → send 18, get 18 ACKs → cwnd = 19
...

Linear: +1 MSS per RTT (much slower than slow start's doubling)
```

### Multiplicative Decrease (MD)

When congestion is detected (packet loss):

```
ssthresh = cwnd / 2
cwnd = 1 MSS          (TCP Tahoe) OR
cwnd = cwnd / 2       (TCP Reno, on 3 dup ACKs)
```

### Why multiplicative decrease and not additive?

AIMD converges to fairness AND efficiency. The math proves:
- **Additive increase** moves toward the efficiency line
- **Multiplicative decrease** moves toward the fairness point

```
              cwnd of Connection 2
                │
                │     ╱ Efficiency line
                │   ╱   (total = capacity)
                │ ╱
      Fair      ╱      
      share →●╱─ ─ ─ ─ Fairness line
              ╱│         (equal share)
            ╱  │
          ╱    │
        ╱      │
      ╱        │
    ╱──────────┼──────── cwnd of Connection 1
               │

AIMD moves toward the ● (fair and efficient):
  - AI: both increase by same amount → moves toward efficiency line
  - MD: both multiply by same factor → moves toward fairness line

Other strategies (AIAD, MIMD) don't converge to the fair point.
```

### The AIMD sawtooth

cwnd over time looks like a sawtooth wave:

```
cwnd
  │          ╱│        ╱│        ╱│
  │        ╱  │      ╱  │      ╱  │
  │      ╱    │    ╱    │    ╱    │
  │    ╱      │  ╱      │  ╱      │
  │  ╱        │╱        │╱        │
  │╱          ╳         ╳         ╳ ← loss events
  └───────────────────────────────────► Time
   slow    cong.    cong.    cong.
   start   avoid   avoid    avoid
```

Each drop halves the window, then linear increase resumes.

---

## ssthresh

**Slow-start threshold (ssthresh)** determines when to switch from slow start (exponential) to congestion avoidance (linear).

```
if cwnd < ssthresh:
    mode = slow_start           # exponential growth
elif cwnd >= ssthresh:
    mode = congestion_avoidance # linear growth (AIMD)
```

### How ssthresh is set

- **Initially**: ssthresh = rwnd (or very large, effectively infinite)
- **On packet loss**: ssthresh = cwnd / 2

```
Example timeline:

1. Connection starts: cwnd=1, ssthresh=65535
2. Slow start: cwnd doubles each RTT → 1, 2, 4, 8, 16, 32, 64
3. No loss, ssthresh far away → stay in slow start
4. Loss at cwnd=64: ssthresh = 32, cwnd = 1 (Tahoe) or 32 (Reno)
5. Slow start again: cwnd = 1, 2, 4, 8, 16, 32
6. cwnd = ssthresh = 32: switch to congestion avoidance
7. cwnd = 32, 33, 34, 35, 36...  (linear)
8. Loss at cwnd=40: ssthresh = 20, cwnd drops
```

---

## Fast Retransmit and Fast Recovery

### Fast retransmit (already covered)

3 duplicate ACKs → retransmit immediately, don't wait for RTO.

### Fast recovery (TCP Reno)

After fast retransmit, don't reset cwnd to 1 (like Tahoe). Instead:

```
1. ssthresh = cwnd / 2
2. cwnd = ssthresh + 3 MSS  (3 for the 3 dup ACKs — those segments left the network)
3. For each additional dup ACK: cwnd += 1 MSS
   (Each dup ACK means another segment left the network → room for one more)
4. When new ACK (non-duplicate) arrives:
   cwnd = ssthresh
   (Deflate back to normal — done recovering)
```

### Why fast recovery is better than reset

```
TCP Tahoe (on 3 dup ACKs):
  cwnd was 40 → cwnd = 1, ssthresh = 20
  Must slow-start back up: 1, 2, 4, 8, 16, 20, 21, 22...
  Takes many RTTs to recover

TCP Reno (on 3 dup ACKs):
  cwnd was 40 → cwnd = 23, ssthresh = 20
  Immediately at half capacity, linear increase: 23, 24, 25...
  Much faster recovery
```

---

## Putting It Together

```python
# Pseudocode: TCP Reno congestion control

cwnd = IW            # initial window (10 MSS)
ssthresh = infinity  # initial threshold

while sending:
    if cwnd < ssthresh:
        # SLOW START: exponential growth
        on_each_ack:
            cwnd += 1 MSS
    else:
        # CONGESTION AVOIDANCE: linear growth (AIMD)
        on_each_ack:
            cwnd += MSS * (MSS / cwnd)  # ≈ 1 MSS per RTT
    
    if timeout:
        # SEVERE CONGESTION: packet totally lost
        ssthresh = cwnd / 2
        cwnd = 1 MSS
        # Enter slow start
    
    if 3_duplicate_acks:
        # MODERATE CONGESTION: fast retransmit + recovery
        ssthresh = cwnd / 2
        cwnd = ssthresh + 3 MSS
        retransmit_lost_segment()
        # Enter fast recovery
        while receiving_dup_acks:
            cwnd += 1 MSS
        on_new_ack:
            cwnd = ssthresh
            # Enter congestion avoidance
```

### Complete lifecycle example

```
cwnd
(MSS)
  │
  │ Phase 1:    Phase 2:       Phase 3:     Phase 4:
  │ Slow Start  Cong. Avoid    Recovery     Cong. Avoid
  │
40│                     ╱──●
  │                   ╱    │ 3 dup ACKs
32│                 ╱      │ ssthresh=20
  │               ╱        ↓
20│          ●──────────────●──────────────────
  │        ╱ ssthresh       ↑               ╱
16│      ╱                  │ cwnd=ssthresh╱
  │    ╱                    │            ╱
 8│  ╱                      │          ╱
  │╱                        │        ╱
 1●                         recovery ●
  └──────────────────────────────────────────► Time (RTTs)
   
   slow start → cong. avoid → loss → recovery → cong. avoid
```

---

## Reno vs Tahoe

| | TCP Tahoe | TCP Reno |
|---|---|---|
| **Year** | 1988 | 1990 |
| **On timeout** | cwnd=1, ssthresh=cwnd/2 | Same |
| **On 3 dup ACKs** | cwnd=1, ssthresh=cwnd/2 | cwnd=cwnd/2, fast recovery |
| **Recovery speed** | Slow (restart from 1) | Fast (restart from half) |
| **Multiple losses** | Handles OK (restarts anyway) | Poor (exits fast recovery too early) |

TCP NewReno (1999) improved Reno by handling multiple losses in the same window.

---

## CUBIC

CUBIC is the default congestion control algorithm in Linux since kernel 2.6.19 (2006).

### Why CUBIC?

Reno's AIMD is too slow for high-bandwidth, high-RTT links. With cwnd increasing by 1 MSS per RTT:

```
10 Gbps link, 100ms RTT, MSS=1460 bytes:
BDP = 10 Gbps × 100ms = 125 MB → need cwnd ≈ 85,000 segments

After a loss event (cwnd halved to 42,500):
Recovery time with Reno = 42,500 RTTs × 100ms = 4,250 seconds ≈ 71 minutes!
```

71 minutes to recover from a single loss event. Unacceptable.

### How CUBIC works

CUBIC uses a **cubic function** of time since the last congestion event:

$$W(t) = C \times (t - K)^3 + W_{max}$$

Where:
- $W(t)$ = congestion window at time $t$ since last loss
- $C$ = constant (0.4)
- $K$ = time to reach $W_{max}$ (the window size at last loss)
- $W_{max}$ = cwnd when loss was detected

```
cwnd
  │            Wmax ──────●──────
  │                     ╱│ │╲
  │                   ╱  │ │  ╲
  │                 ╱    │ │    ╲
  │               ╱      │ │      ╲  ← Probing above Wmax
  │  Concave    ╱        │ │        ╲   (convex region)
  │  (fast)   ╱          │ │
  │         ╱            │ │
  │       ╱              │ │
  │     ╱ ← Starting     │ │
  │   ╱      slow near   │ │
  │  ╱       Wmax        │ │
  │ ╱                    │ │
  │╱                     │ │
  └──────────────────────┼─┼──────► Time
        loss at ← →     K
```

The key insight: **CUBIC is concave (fast growth) when far from $W_{max}$, and convex (slow probing) near and past $W_{max}$**. It quickly recovers to near the previous capacity, then cautiously probes above.

### CUBIC advantages

1. **RTT-independent**: Growth depends on time since loss, not RTT. Fairer for high-RTT flows.
2. **Fast recovery**: Rapidly returns to previous capacity
3. **Cautious probing**: Slows down near the capacity that caused the last loss

```bash
# CUBIC is the default on Linux
cat /proc/sys/net/ipv4/tcp_congestion_control
# cubic

# CUBIC parameters (kernel internal, viewable per-connection)
ss -ti | grep cubic
# Shows: cubic wscale:7,7 rto:204 rtt:1.5/0.5 cwnd:10 ssthresh:7
```

---

## BBR

BBR (Bottleneck Bandwidth and Round-trip propagation time) is Google's congestion control algorithm, fundamentally different from loss-based approaches.

### The problem with loss-based CC

Traditional algorithms (Reno, CUBIC) use **packet loss** as the congestion signal. But modern networks have deep buffers:

```
Deep buffer router:
  1. Sender increases cwnd
  2. Buffer absorbs excess → no drops
  3. Latency increases (packets queuing)
  4. Sender keeps increasing (no loss signal!)
  5. Buffer finally overflows → drops
  6. But latency has been high for a long time → bufferbloat!

Loss-based CC fills buffers BEFORE detecting congestion.
```

### BBR's approach

BBR doesn't wait for loss. Instead, it measures:
1. **Bottleneck bandwidth (BtlBw)**: Maximum delivery rate observed
2. **Round-trip propagation time (RTprop)**: Minimum RTT observed

And sets cwnd to match the **BDP** (bandwidth-delay product):

$$\text{cwnd}_{BBR} = \text{BtlBw} \times \text{RTprop}$$

```
BBR operates in 4 phases:

1. STARTUP:    Exponential growth (like slow start)
               Until bandwidth stops increasing
               
2. DRAIN:      Reduce cwnd to drain any queue built during startup
               Until inflight = BDP

3. PROBE_BW:   Steady state. Cycles through:
               - 1.25x BDP (probe for more bandwidth) ← 1 RTT
               - 0.75x BDP (drain)                    ← 1 RTT
               - 1.0x BDP (cruise)                    ← 6 RTTs

4. PROBE_RTT:  Periodically reduce cwnd to 4 segments
               to measure minimum RTT (every 10 seconds)
```

### BBR vs CUBIC comparison

```
Network with bufferbloat (deep buffers):

CUBIC:
  ├── Fills buffer ─── High latency ─── Eventually loses packet ─── Drops cwnd ──┤
  └─────── Repeat: latency stays high, throughput oscillates ─────────────────────┘

BBR:
  ├── Measures bandwidth and RTT ─── Sets cwnd = BDP ─── No buffer filling ──┤
  └─────── Low latency, stable throughput ───────────────────────────────────┘
```

### Enabling BBR

```bash
# Load BBR module
sudo modprobe tcp_bbr

# Set as default
sudo sysctl -w net.ipv4.tcp_congestion_control=bbr

# Verify
cat /proc/sys/net/ipv4/tcp_congestion_control
# bbr

# Make persistent across reboots
echo "net.ipv4.tcp_congestion_control=bbr" | sudo tee -a /etc/sysctl.conf
echo "tcp_bbr" | sudo tee -a /etc/modules-load.d/bbr.conf

# BBR also benefits from using fq (fair queue) qdisc
sudo tc qdisc replace dev eth0 root fq
```

### BBR caveats

1. **BBRv1 fairness issues**: Can be aggressive against CUBIC flows, taking disproportionate bandwidth
2. **BBRv2** (in development) addresses fairness and loss handling
3. **Not always better**: On networks with shallow buffers and minimal bufferbloat, CUBIC works fine
4. **RTT sensitivity**: BBR's PROBE_RTT phase can cause periodic throughput drops

---

## Comparing

| Algorithm | Year | Based On | Best For | Issue |
|-----------|------|----------|----------|-------|
| **Tahoe** | 1988 | Loss | Historical | Slow recovery |
| **Reno** | 1990 | Loss | General | Poor with multiple losses |
| **NewReno** | 1999 | Loss | General | Still slow on high-BDP links |
| **CUBIC** | 2006 | Loss | High BDP | Fills buffers (bufferbloat) |
| **BBR** | 2016 | Bandwidth/RTT | Cloud, CDN | Fairness concerns (v1) |
| **BBRv2** | 2019+ | Bandwidth/RTT | General | Still in development |

### Which to use?

```
Datacenter (low RTT, shallow buffers):    CUBIC is fine
Internet server (mixed clients):          CUBIC (safe default)
CDN / high-volume server:                 BBR (better throughput, lower latency)
Satellite / high-RTT:                     BBR or CUBIC (both handle well)
```

---

## Linux: Choosing and Tuning

### View and change congestion control

```bash
# Current algorithm
cat /proc/sys/net/ipv4/tcp_congestion_control

# Available algorithms
cat /proc/sys/net/ipv4/tcp_available_congestion_control

# Load all compiled algorithms
ls /lib/modules/$(uname -r)/kernel/net/ipv4/tcp_*.ko 2>/dev/null

# Change (runtime)
sudo sysctl -w net.ipv4.tcp_congestion_control=bbr

# Per-route (different algo for different destinations)
sudo ip route add 10.0.0.0/24 via 192.168.1.1 congctl bbr
```

### Monitor cwnd and ssthresh

```bash
# Per-connection details
ss -ti

# Watch cwnd during a transfer
watch -n 0.5 "ss -ti dst <server_ip> | grep -oP 'cwnd:\d+|ssthresh:\d+|rtt:[^ ]+'"

# Log cwnd over time (for graphing)
while true; do
  echo "$(date +%s) $(ss -tin dst <server_ip> | grep -oP 'cwnd:\d+')" >> /tmp/cwnd.log
  sleep 0.1
done
```

### Key sysctl parameters

```bash
# Initial congestion window (segments)
ip route show | grep initcwnd
# Change: sudo ip route change default via <gw> initcwnd 10

# Slow start after idle
cat /proc/sys/net/ipv4/tcp_slow_start_after_idle
# 1 = reset cwnd to IW after idle (default)
# 0 = keep cwnd after idle (better for persistent connections)

# For keep-alive connections (web servers), disable:
sudo sysctl -w net.ipv4.tcp_slow_start_after_idle=0

# Slow start behavior
cat /proc/sys/net/ipv4/tcp_no_metrics_save
# 0 = save metrics in route cache (default)
# When connecting to a previously-seen destination, start with saved ssthresh

# Enable TCP pacing (smoother sending, less burstiness)
# BBR requires this, CUBIC benefits from it
sudo tc qdisc replace dev eth0 root fq
```

---

## Key Takeaways

1. **cwnd** is the sender's internal limit on in-flight data — never sent on the wire
2. **Slow start**: Exponential growth (double every RTT) until ssthresh or loss
3. **Congestion avoidance (AIMD)**: Linear increase (+1 MSS/RTT), multiplicative decrease (halve on loss)
4. **ssthresh** = mode switch between slow start and congestion avoidance
5. **The AIMD sawtooth** is the fundamental TCP behavior: grow, drop, grow, drop
6. **CUBIC** (Linux default) uses a cubic function for faster recovery on high-BDP links
7. **BBR** measures bandwidth and RTT instead of reacting to loss — avoids bufferbloat
8. **Disable `tcp_slow_start_after_idle`** for servers with persistent connections
9. **Monitor with `ss -ti`**: cwnd, ssthresh, rtt, retrans are your diagnostic tools

---

## Next

→ [03-bufferbloat-and-beyond.md](03-bufferbloat-and-beyond.md) — Modern congestion problems
