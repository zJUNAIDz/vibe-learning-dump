# Latency Deep Dive — Understanding Every Component of Delay

> "Latency is the tax you pay for distance, for queuing, for processing, and for protocol overhead. Most engineers can name the first one but forget the other three."

---

## Table of Contents

1. [What Latency Actually Is](#what-latency-actually-is)
2. [The Four Components of Latency](#the-four-components-of-latency)
3. [Propagation Delay — The Speed of Light Tax](#propagation-delay--the-speed-of-light-tax)
4. [Transmission Delay — Pushing Bits Onto the Wire](#transmission-delay--pushing-bits-onto-the-wire)
5. [Processing Delay — What Routers Do](#processing-delay--what-routers-do)
6. [Queuing Delay — The Wildcard](#queuing-delay--the-wildcard)
7. [End-to-End Latency: Putting It All Together](#end-to-end-latency-putting-it-all-together)
8. [RTT vs One-Way Delay](#rtt-vs-one-way-delay)
9. [Why Latency Matters More Than Bandwidth (Often)](#why-latency-matters-more-than-bandwidth-often)
10. [Measuring Latency on Linux](#measuring-latency-on-linux)
11. [Common Latency Misinterpretations](#common-latency-misinterpretations)
12. [Protocol-Induced Latency](#protocol-induced-latency)

---

## What Latency Actually Is

Latency is the **time it takes for data to get from point A to point B**.

That sounds simple, but the simplicity is deceptive. There are multiple types of latency, multiple ways to measure it, and multiple reasons it varies. Let's be precise.

### Definitions that matter

**One-way latency**: The time from when a sender transmits the first bit of a packet to when the receiver receives the last bit of that packet.

**Round-trip time (RTT)**: The time from when a sender transmits a packet to when it receives the response. RTT ≈ 2 × one-way latency, but not exactly, because the forward and return paths may be different.

**Application-perceived latency**: The time from when an application sends a request to when it receives a usable response. This includes:
- Network latency (RTT)
- Server processing time
- Protocol overhead (TCP handshake, TLS handshake)
- Serialization/deserialization time

When someone says "the latency is 50ms," you need to ask: "One-way or round-trip? Network latency or total latency? Average, median, P99, or maximum?"

---

## The Four Components of Latency

Every packet's latency is the sum of four components at every hop:

$$\text{Total Latency} = \sum_{\text{each hop}} (d_{\text{prop}} + d_{\text{trans}} + d_{\text{proc}} + d_{\text{queue}})$$

Let's understand each one deeply.

---

## Propagation Delay — The Speed of Light Tax

### What it is

Propagation delay is the time it takes for a signal to physically travel from one end of a link to the other.

$$d_{\text{prop}} = \frac{\text{distance}}{\text{speed of signal in medium}}$$

### The hard physics

Light in a vacuum travels at $c = 3 \times 10^8$ m/s. But signals don't travel in vacuum:
- In **fiber optic cable**: ~$2 \times 10^8$ m/s (about 2/3 the speed of light, because glass has a refractive index of ~1.5)
- In **copper wire**: ~$2 \times 10^8$ m/s (similar to fiber, depends on cable type)
- In **air/radio waves**: ~$3 \times 10^8$ m/s (nearly speed of light)

### Real-world examples

| Route | Distance | Propagation Delay (one-way) |
|-------|----------|---------------------------|
| Same data center rack | ~2 meters | ~10 nanoseconds |
| Across a data center | ~500 meters | ~2.5 microseconds |
| New York to San Francisco | ~4,000 km | ~20 milliseconds |
| New York to London | ~5,500 km | ~27 milliseconds |
| New York to Tokyo | ~10,800 km | ~54 milliseconds |
| New York to Sydney | ~16,000 km | ~80 milliseconds |

**Key insight**: Propagation delay is **fixed by physics**. No amount of money, engineering, or optimization can reduce it below the speed of light limit. If you need sub-20ms one-way latency between New York and London, you must move the server closer.

This is why:
- CDNs exist (bring content closer to users)
- Financial firms co-locate servers at exchanges
- Cloud providers have regions distributed globally

### The speed of light is slow at internet scale

At human scale, light seems instant. At internet scale, it's the bottleneck:
- A round trip from the US East Coast to a server in Asia and back is ~200ms just from propagation
- A human perceives delay above ~100ms
- Interactive applications (gaming, video calls) need <150ms RTT to feel responsive
- High-frequency trading firms pay millions for microseconds less propagation delay

### Measuring propagation delay

You can't directly measure propagation delay on the internet because it's mixed with other delays. But you can estimate it:

```bash
# Ping a distant server
ping -c 10 example.com

# The MINIMUM RTT you observe is closest to pure propagation delay
# (because it had minimal queuing delay)
# Divide by 2 for one-way estimate
```

---

## Transmission Delay — Pushing Bits Onto the Wire

### What it is

Transmission delay (also called serialization delay) is the time it takes to push all the bits of a packet onto the link. Think of it like squeezing toothpaste out of a tube — you can only push bits out at the rate the link allows.

$$d_{\text{trans}} = \frac{\text{packet size (bits)}}{\text{link bandwidth (bits/sec)}}$$

### Examples

| Packet Size | Link Speed | Transmission Delay |
|-------------|-----------|-------------------|
| 1500 bytes (12,000 bits) | 100 Mbps | 120 microseconds |
| 1500 bytes | 1 Gbps | 12 microseconds |
| 1500 bytes | 10 Gbps | 1.2 microseconds |
| 1500 bytes | 100 Gbps | 0.12 microseconds |

### When transmission delay matters

On modern high-speed links (1 Gbps+), transmission delay for standard-sized packets (1500 bytes) is negligible — a few microseconds.

But it matters in specific scenarios:
1. **Slow links**: On a 1 Mbps link, a 1500-byte packet takes 12 ms to transmit. That's significant.
2. **Large packets/frames**: A 9000-byte jumbo frame on a 1 Gbps link takes 72 microseconds.
3. **Serialization at every hop**: If there are 15 hops, transmission delay at each hop adds up.
4. **Store-and-forward switches**: A switch must receive the entire frame before forwarding it, so transmission delay is paid at every hop (except with cut-through switching).

### Transmission vs propagation — the pithy distinction

**Propagation delay**: How long does it take for the first bit to reach the other end? (Distance / speed)
**Transmission delay**: How long does it take to push the entire packet onto the wire? (Packet size / bandwidth)

These are independent. You can have:
- High propagation, low transmission: Fiber link across the ocean (long distance, high speed)
- Low propagation, high transmission: Short copper cable but slow link speed

```bash
# See your link speed
ethtool eth0 | grep Speed
# Example output: Speed: 1000Mb/s

# Calculate transmission delay for a 1500-byte packet:
# 1500 × 8 / 1,000,000,000 = 0.000012 seconds = 12 microseconds
```

---

## Processing Delay — What Routers Do

### What it is

Processing delay is the time a router (or switch, or host) takes to:
1. Receive the packet header
2. Check for bit errors
3. Determine the output link (routing table lookup)
4. Apply access control / firewall rules
5. Update header fields (decrement TTL, recalculate checksum)

### How long does this take?

On modern hardware:
- A simple routing table lookup: **nanoseconds to microseconds**
- A firewall rule check with a small rule set: **microseconds**
- Deep packet inspection (reading application data): **tens of microseconds to milliseconds**
- Software-based routers: **tens of microseconds** (CPU processes each packet)
- Hardware-based routers (ASICs): **nanoseconds** (dedicated silicon)

For most internet paths, processing delay at each router is **negligible** (microseconds) compared to propagation delay (milliseconds). But:
- If there are 15+ hops, microseconds add up
- Complex firewall rules can add significant delay
- Overloaded routers may have increased processing delay

### Where processing delay becomes significant

1. **Firewalls with thousands of rules**: Each packet must be checked against every rule in order. Some firewalls with 10,000+ rules can add milliseconds.
2. **NAT devices**: NAT must modify headers and maintain connection state tables. Under heavy load, NAT processing can become a bottleneck.
3. **Software routers** (e.g., Linux machine acting as a router): CPU-based forwarding is much slower than hardware forwarding.
4. **DPI (Deep Packet Inspection)**: Inspecting application-layer content is orders of magnitude slower than reading IP headers.

---

## Queuing Delay — The Wildcard

### What it is

Queuing delay is the time a packet spends **waiting in a buffer** at a router or switch because the outgoing link is busy transmitting other packets.

This is the most **variable** and **unpredictable** component of latency. It can range from zero to hundreds of milliseconds, depending on network load.

### Why queuing happens

When packets arrive at a router faster than the router can forward them, the excess packets are stored in a buffer (queue). They wait their turn.

```
Imagine a highway on-ramp (merge point):

                   ┌────────────────┐
Incoming Link 1 ──→│                │
                   │  Router Queue  │──→ Outgoing Link (congested)
Incoming Link 2 ──→│  ████████░░░░  │
                   │                │
Incoming Link 3 ──→│                │
                   └────────────────┘

Three 1 Gbps incoming links feeding one 1 Gbps outgoing link.
If all three are at full capacity: 3 Gbps in, 1 Gbps out.
2 Gbps worth of packets queue up... and eventually get dropped.
```

### The queuing delay spectrum

```
Network Load    Queuing Delay    What You Observe
──────────────────────────────────────────────────
0-30%           Near zero        Consistent low latency
30-60%          Small, stable    Slightly higher latency
60-80%          Growing          Noticeable latency increase
80-95%          Large, variable  Latency spikes, jitter
95-100%         Massive          Timeouts, packet drops
100%+           Buffer overflow  Packets DROPPED
```

### How queuing delay relates to jitter

**Jitter** is the variation in latency over time. When queuing delay is low and stable, jitter is low. When queuing delay is high and variable, jitter increases.

Applications that are sensitive to jitter (VoIP, video conferencing, gaming) suffer dramatically when queuing delay is variable, even if the average delay is acceptable.

### Bufferbloat: too much queuing

We'll cover this deeply in Module 07, but here's the preview:

**Bufferbloat** occurs when routers have oversized buffers. Instead of dropping packets when congested (which signals TCP to slow down), they queue thousands of packets. The result:
- Latency increases by hundreds of milliseconds
- TCP can't detect congestion (no packet loss)
- Interactive traffic (web browsing, SSH) becomes sluggish
- The problem is invisible to bandwidth speed tests

```bash
# Test for bufferbloat
# Run a speed test while simultaneously pinging:
ping -c 100 8.8.8.8  # Watch RTT increase during heavy traffic
```

---

## End-to-End Latency: Putting It All Together

### Trace of a real request

Let's estimate the latency for an HTTP request from New York to a server in London:

```
Hop 1: Your PC → Your Router (local)
  Propagation: ~0.01 ms (10m cable)
  Transmission: ~0.01 ms (1 Gbps)
  Processing: ~0.01 ms
  Queuing: ~0.1 ms (home router, moderate load)
  Subtotal: ~0.13 ms

Hops 2-4: Your Router → ISP → regional backbone
  Propagation: ~1 ms (100 km total)
  Processing + Queuing: ~1 ms (3 hops, light load)
  Subtotal: ~2 ms

Hops 5-8: US backbone (New York → undersea cable landing)
  Propagation: ~2 ms (200 km)
  Processing + Queuing: ~1 ms (4 hops)
  Subtotal: ~3 ms

Hop 9: Undersea cable (New York → London)
  Propagation: ~27 ms (5,500 km of fiber)
  Processing: negligible
  Subtotal: ~27 ms

Hops 10-13: London backbone → data center
  Propagation: ~1 ms
  Processing + Queuing: ~1 ms
  Subtotal: ~2 ms

TOTAL ONE-WAY: ~34 ms
RTT: ~68 ms (assuming symmetric return path)
```

But wait — that's just the network latency. For an HTTPS request:

```
Network RTT: ~68 ms

TCP handshake: 1 RTT = 68 ms
TLS 1.2 handshake: 2 RTTs = 136 ms
HTTP request + response: 1 RTT = 68 ms
Server processing: ~20 ms

TOTAL time to first byte: 68 + 136 + 68 + 20 = ~292 ms

With TLS 1.3: TCP handshake (68ms) + TLS handshake 1 RTT (68ms) + HTTP (68ms) + processing (20ms) = ~224 ms
```

This is why **reducing RTT is so powerful** — it reduces the cost of every handshake and round trip.

---

## RTT vs One-Way Delay

### Why we usually measure RTT

**One-way delay** requires synchronized clocks on both machines. Even with NTP (Network Time Protocol), clock synchronization is only accurate to a few milliseconds, which introduces measurement error that may be larger than the latency you're trying to measure.

**RTT** only requires one clock. You timestamp when you send and when you receive the reply. The difference is RTT. No clock synchronization needed.

### When RTT ≠ 2 × one-way delay

Asymmetric routing: The forward path (A→B) might go through different routers than the return path (B→A). Different congestion levels on each path can create asymmetric delays.

```
A → B: 30ms (5 hops, all fast)
B → A: 50ms (8 hops, one congested)
RTT: 80ms
Average one-way: 40ms
But actual one-way delays: 30ms and 50ms — not equal
```

This asymmetry is common on the internet. When you ping a server and get 80ms RTT, you don't know if it's 40ms each way or 20ms + 60ms.

### Measuring on Linux

```bash
# Basic RTT measurement
ping -c 20 example.com
# Shows: min/avg/max/mdev
# mdev = standard deviation → indicator of jitter

# Detailed per-packet timing
ping -c 10 -D example.com
# -D adds timestamps to each packet

# Use mtr for per-hop latency
mtr --report -c 100 example.com
# Shows latency at EACH hop, not just end-to-end
# Helps identify WHERE the delay is

# TCP-level RTT (from established connections)
ss -ti
# Shows "rtt:X/Y" where X is smoothed RTT, Y is RTT variance
```

---

## Why Latency Matters More Than Bandwidth (Often)

### The bandwidth-delay product

Bandwidth tells you how many bits per second can flow through a pipe. Latency tells you how long the pipe is. The **bandwidth-delay product** (BDP) tells you how many bits are "in flight" — inside the pipe at any moment.

$$\text{BDP} = \text{bandwidth} \times \text{RTT}$$

Example:
- 100 Mbps link, 10ms RTT → BDP = 100 × 0.01 = 1 Mbit = 125 KB
- 100 Mbps link, 200ms RTT → BDP = 100 × 0.2 = 20 Mbit = 2.5 MB

The BDP determines how much data TCP must keep "in flight" to fully utilize the link. With high latency, TCP needs larger buffers and more time to ramp up.

### Why doubling bandwidth doesn't halve page load time

Web pages require many small resources (HTML, CSS, JS, images). Loading each resource involves:
1. DNS lookup (1 RTT if not cached)
2. TCP connection (1 RTT)
3. TLS handshake (1-2 RTTs)
4. HTTP request/response (1+ RTTs)

If RTT is 100ms, loading one resource takes ~400ms minimum, regardless of bandwidth. With 50 resources, even with parallel connections, latency dominates.

Going from 10 Mbps to 100 Mbps might not noticeably improve page load time if the bottleneck is latency (RTTs). Going from 100ms RTT to 20ms RTT will dramatically improve it.

This is why:
- CDNs reduce latency by serving from nearby locations
- HTTP/2 multiplexes requests over one connection (saves handshake RTTs)
- HTTP/3 (QUIC) combines TCP + TLS handshakes (saves 1 RTT)
- `dns-prefetch`, `preconnect` hints reduce effective RTTs

```bash
# Measure the effect of latency on page load
# Time a curl request (shows connection vs transfer time)
curl -o /dev/null -s -w "DNS: %{time_namelookup}s\nConnect: %{time_connect}s\nTLS: %{time_appconnect}s\nFirst byte: %{time_starttransfer}s\nTotal: %{time_total}s\n" https://example.com
```

This output separates DNS lookup time, TCP connection time, TLS handshake time, server processing time, and total transfer time. You'll often find that most of the time is spent on handshakes and waiting, not data transfer.

---

## Measuring Latency on Linux

### Tool 1: ping

```bash
# Basic latency measurement
ping -c 20 8.8.8.8

# Output explanation:
# 64 bytes from 8.8.8.8: icmp_seq=1 ttl=118 time=12.3 ms
#                                                ^^^^^^^^
#                                                This is RTT

# Summary line:
# rtt min/avg/max/mdev = 11.5/12.8/15.2/1.1 ms
# min   = best case (minimal queuing)
# avg   = average
# max   = worst case (most queuing)
# mdev  = standard deviation (jitter indicator)
```

**Gotcha**: ICMP (ping) packets may be rate-limited or deprioritized by routers. The latency you see with ping might not match what TCP traffic experiences. Some routers treat ICMP as low priority.

### Tool 2: mtr (My Traceroute)

```bash
# Per-hop latency analysis
mtr --report -c 50 example.com

# Output shows each router hop with:
# - Loss% at that hop
# - Last, Avg, Best, Worst latency TO that hop
# - Standard deviation

# Interactive mode (live updates)
mtr example.com
```

**How to read mtr**: 
- If hop 5 shows 50ms avg and hop 6 shows 120ms avg, the link between hops 5 and 6 adds ~70ms
- If a hop shows high loss but subsequent hops don't, the router is probably deprioritizing ICMP (not actually dropping traffic)
- Look for sudden jumps in latency — that's where the delay is

### Tool 3: ss (for TCP RTT)

```bash
# See smoothed RTT for established TCP connections
ss -ti state established

# Output includes:
# rtt:12.5/6.25 → smoothed RTT: 12.5ms, RTT variance: 6.25ms
# This is what TCP actually uses for timeout calculations
```

### Tool 4: curl timing

```bash
# Detailed timing breakdown
curl -w "\
DNS Lookup:    %{time_namelookup}s\n\
TCP Connect:   %{time_connect}s\n\
TLS Handshake: %{time_appconnect}s\n\
First Byte:    %{time_starttransfer}s\n\
Total Time:    %{time_total}s\n\
" -o /dev/null -s https://example.com
```

---

## Common Latency Misinterpretations

### Trap 1: "Average latency is fine, so there's no problem"

Average latency can hide terrible tail latency. If 99% of requests take 10ms but 1% take 500ms, the average is ~15ms — looks fine! But 1 in 100 users waits half a second. This is why P95, P99, and P99.9 percentiles matter more than averages.

### Trap 2: "Our latency suddenly increased from 10ms to 50ms"

Before panicking:
- Is it RTT or one-way?
- Is a congested link adding queuing delay? (Check different times of day)
- Did a routing change add more hops?
- Is the server under more load? (Server processing time increased)
- Did a DNS change point to a more distant server?

### Trap 3: "Ping shows 5ms but the application takes 200ms"

Ping measures ICMP RTT. Application latency includes:
- TCP handshake (1 RTT minimum)
- TLS handshake (1-2 RTT)
- Server processing time
- Data transfer time
- Application serialization/deserialization

5ms ping + all the overhead easily becomes 200ms application latency.

### Trap 4: "The latency is high because we need more bandwidth"

As explained above, for small requests (web pages, API calls), latency is usually bounded by RTT and protocol handshakes, not bandwidth. More bandwidth helps large file transfers, not interactive web applications.

### Trap 5: "This traceroute hop shows 100ms — there's a problem at hop 7"

Some routers deprioritize ICMP (the protocol used by traceroute). They respond to traceroute packets slowly but forward data packets normally. A high-latency hop in traceroute that doesn't cause high end-to-end latency is a false alarm.

Look at the **end-to-end** latency. If the final hop shows 30ms RTT but an intermediate hop shows 100ms, the intermediate hop is likely just slow at responding to ICMP, not slow at forwarding your traffic.

---

## Protocol-Induced Latency

Beyond physical network latency, protocols add their own delays:

### TCP three-way handshake: 1 RTT

```
Client → Server: SYN
Server → Client: SYN-ACK       ← 1 RTT
Client → Server: ACK + data
```

Before any application data flows, TCP consumes 1 full RTT for the handshake. With an 80ms RTT, that's 80ms of pure waiting.

### TLS 1.2 handshake: 2 additional RTTs

```
Client → Server: ClientHello
Server → Client: ServerHello, Certificate    ← 1 RTT
Client → Server: Key Exchange, Finished
Server → Client: Finished                    ← 2 RTTs total
```

### TLS 1.3 handshake: 1 additional RTT

```
Client → Server: ClientHello + Key Share
Server → Client: ServerHello + Key Share + encrypted data  ← 1 RTT
```

TLS 1.3 saves a full RTT over TLS 1.2 by combining messages. This is one of the biggest practical improvements in TLS 1.3.

### HTTP/1.1 head-of-line blocking

HTTP/1.1 is request-response: send a request, wait for the response, then send the next request. Even with pipelining (sending multiple requests without waiting), the responses must arrive in order.

If the first response is slow (large file), subsequent small responses are blocked behind it.

### DNS resolution: 0-4 RTTs depending on caching

```
Cache hit (local):    0 RTT
Cache hit (resolver): 1 RTT
Full resolution:      2-4 RTTs (recursive: root → TLD → authoritative)
```

### Total first-request latency (cold start)

```
DNS resolution:     ~50ms (assuming 1 RTT to resolver)
TCP handshake:      ~80ms (1 RTT)
TLS 1.3 handshake:  ~80ms (1 RTT)  
HTTP request:       ~80ms (1 RTT) + server processing
─────────────────────────────────
TOTAL:              ~290ms + server processing

That's ~4 RTTs before the user sees any content.
On a 200ms RTT connection (mobile, satellite): ~800ms
On a 400ms RTT connection (global, congested):  ~1600ms
```

This is why every protocol optimization that saves even one RTT has a significant impact on user experience.

---

## Key Takeaways

1. **Latency has four components**: propagation, transmission, processing, queuing — and they have radically different characteristics
2. **Propagation delay is bounded by physics**: speed of light, can't be optimized, can only be reduced by moving endpoints closer
3. **Queuing delay is the wildcard**: varies from zero to hundreds of ms, causes jitter, and creates the most debugging headaches
4. **RTT is the practical measurement**: one-way delay requires synchronized clocks; RTT doesn't
5. **Latency often matters more than bandwidth** for interactive applications — it determines how many round trips are needed
6. **Protocol overhead multiplies latency impact**: each handshake costs 1+ RTTs, so high RTT environments pay an enormous tax
7. **Don't trust averages**: use percentiles (P95, P99) to understand tail latency
8. **Don't trust ping alone**: ICMP may be treated differently than real traffic

---

## Next

→ [02-bandwidth-throughput.md](02-bandwidth-throughput.md) — Why "fast internet" is ambiguous and bandwidth ≠ throughput
