# TCP Reliability & Retransmission — Guaranteeing Delivery Over an Unreliable Network

> IP loses packets. Links corrupt data. Routers drop packets when busy. TCP's job is to create the **illusion of a perfectly reliable byte stream** on top of all this chaos. This file explains exactly how.

---

## Table of Contents

1. [The Reliability Problem](#the-problem)
2. [Sequence Numbers](#sequence-numbers)
3. [Acknowledgments (ACKs)](#acks)
4. [Retransmission Timer (RTO)](#rto)
5. [Fast Retransmit](#fast-retransmit)
6. [Selective Acknowledgment (SACK)](#sack)
7. [Duplicate ACKs and Why They Matter](#duplicate-acks)
8. [TCP Segmentation](#tcp-segmentation)
9. [Nagle's Algorithm](#nagles-algorithm)
10. [Delayed ACKs](#delayed-acks)
11. [Nagle + Delayed ACKs: The Deadly Interaction](#nagle-delayed-ack-interaction)
12. [Linux: Retransmission Debugging](#linux-debugging)
13. [Key Takeaways](#key-takeaways)

---

## The Problem

IP provides "best effort" delivery. Packets can be:
- **Lost**: Router queue overflow, wireless interference, cable unplugged
- **Duplicated**: Network path changes cause a packet to traverse two paths
- **Reordered**: Different packets take different routes with different delays
- **Corrupted**: Bit errors from noise (caught by checksums)

TCP must handle ALL of these and present a clean, ordered, complete byte stream to the application.

The mechanisms:
1. **Sequence numbers** → detect reordering and duplicates
2. **ACKs** → confirm receipt
3. **Checksums** → detect corruption
4. **Retransmission timers** → recover from loss
5. **SACK** → efficiently identify what's missing

---

## Sequence Numbers

Every byte transmitted over TCP has a sequence number. The initial sequence number (ISN) is exchanged during the handshake.

### How sequence numbers work

```
ISN = 1000 (from handshake)

Send "Hello" (5 bytes):
  Segment: seq=1000, data="Hello"
  Bytes 1000, 1001, 1002, 1003, 1004

Send "World" (5 bytes):
  Segment: seq=1005, data="World"
  Bytes 1005, 1006, 1007, 1008, 1009

Next expected sequence number: 1010
```

### Sequence numbers are byte-based, not packet-based

TCP numbers **bytes**, not packets. If you send a 1000-byte segment starting at sequence 5000, the next segment starts at sequence 6000.

```
Packet 1: seq=5000, len=1000  → covers bytes 5000-5999
Packet 2: seq=6000, len=1000  → covers bytes 6000-6999
Packet 3: seq=7000, len=500   → covers bytes 7000-7499
```

### ISN randomization

The ISN is **not** 0. It's randomized for security reasons:

- Predictable ISNs allow TCP sequence prediction attacks (IP spoofing + guessing the next sequence number → inject data into someone else's connection)
- Linux uses a cryptographic hash: `ISN = hash(src_ip, dst_ip, src_port, dst_port, secret_key, time)`

### 32-bit sequence space

Sequence numbers are 32 bits (0 to 4,294,967,295). On a fast link, they can wrap around:

| Link Speed | Time to Wrap |
|------------|-------------|
| 10 Mbps | ~57 minutes |
| 100 Mbps | ~5.7 minutes |
| 1 Gbps | ~34 seconds |
| 10 Gbps | ~3.4 seconds |

TCP timestamps solve this (PAWS — Protection Against Wrapped Sequences). A segment with a timestamp older than recently received segments is rejected, even if the sequence number happens to be valid after wrapping.

---

## ACKs

### Cumulative ACKs

TCP ACKs are **cumulative**: ACK number N means "I have received all bytes up to N-1, send byte N next."

```
Sender                              Receiver
  |  seq=1000, len=500              |
  |-------------------------------->| Received bytes 1000-1499
  |                                 |
  |  seq=1500, len=500              |
  |-------------------------------->| Received bytes 1500-1999
  |                                 |
  |          ack=2000               |
  |<--------------------------------| "I have everything up to 1999"
```

### The problem with cumulative ACKs

If segments arrive out of order or with gaps, cumulative ACKs can only report the highest contiguous byte received:

```
Sender                              Receiver
  |  seq=1000, len=500  (received)  |
  |-------------------------------->| Got 1000-1499 ✓
  |                                 |
  |  seq=1500, len=500  (LOST!)     |
  |           X                     |
  |                                 |
  |  seq=2000, len=500  (received)  |
  |-------------------------------->| Got 2000-2499 ✓
  |                                 | But missing 1500-1999!
  |          ack=1500               |
  |<--------------------------------| Can only ACK up to 1500
  |                                 | (can't say "I have 2000-2499 but not 1500-1999")
```

The receiver has bytes 1000-1499 and 2000-2499, but the cumulative ACK can only say "I have up to 1500." The sender doesn't know that 2000-2499 arrived. Without SACK (see below), it might retransmit both segments.

---

## RTO

### How does the sender know a packet was lost?

It waits. If no ACK arrives within a timeout, the packet is presumed lost and retransmitted. This timeout is the **Retransmission Timeout (RTO)**.

### Computing RTO

The RTO must adapt to network conditions. A fixed timeout would be too short for satellite links (600ms RTT) and too long for datacenter links (0.1ms RTT).

TCP continuously measures the **Smoothed RTT (SRTT)** and **RTT Variance (RTTVAR)**:

$$SRTT = (1 - \alpha) \times SRTT + \alpha \times \text{measured RTT}$$

$$RTTVAR = (1 - \beta) \times RTTVAR + \beta \times |SRTT - \text{measured RTT}|$$

$$RTO = SRTT + 4 \times RTTVAR$$

Where $\alpha = 1/8$ and $\beta = 1/4$ (RFC 6298).

**Intuition**: RTO = average RTT + safety margin based on how much the RTT varies.

### RTO boundaries

```
Minimum RTO: 200ms (Linux default, set by TCP_RTO_MIN)
Maximum RTO: 120 seconds (TCP_RTO_MAX)
Initial RTO: 1 second (before any measurement)
```

### Exponential backoff

When a retransmission also times out, the RTO doubles:

```
1st attempt:  RTO = 300ms     → timeout
2nd attempt:  RTO = 600ms     → timeout
3rd attempt:  RTO = 1200ms    → timeout
4th attempt:  RTO = 2400ms    → timeout
...
Until max retries (tcp_retries2 = 15, taking ~13-30 minutes)
```

```bash
# Check max retries
cat /proc/sys/net/ipv4/tcp_retries2
# 15 (default)
# Total time before giving up: ~13-30 minutes

cat /proc/sys/net/ipv4/tcp_retries1
# 3 (soft error threshold — after 3 retries, report to IP layer)
```

### Karn's algorithm

**Problem**: If you retransmit a segment and then receive an ACK, was that ACK for the original or the retransmission? If it was for the original (just delayed), your RTT measurement would be wrong.

**Solution (Karn's algorithm)**: Don't update SRTT/RTTVAR based on retransmitted segments. Only measure RTT from segments that are ACKed without retransmission.

TCP timestamps solve this problem more elegantly — the timestamp in the ACK tells you exactly which send it corresponds to.

---

## Fast Retransmit

Waiting for the RTO timer is slow. If the RTO is 300ms, you waste 300ms before retransmitting. Fast retransmit detects loss sooner.

### How it works

When a receiver gets an out-of-order segment, it immediately sends a **duplicate ACK** — an ACK for the same byte it already ACKed.

```
Sender                              Receiver
  |  seq=1000, len=500              |
  |-------------------------------->| ack=1500 (normal ACK)
  |                                 |
  |  seq=1500, len=500  (LOST!)     |
  |           X                     |
  |                                 |
  |  seq=2000, len=500              |
  |-------------------------------->| ack=1500 (duplicate ACK #1 — "I still want 1500")
  |                                 |
  |  seq=2500, len=500              |
  |-------------------------------->| ack=1500 (duplicate ACK #2 — "I still want 1500!")
  |                                 |
  |  seq=3000, len=500              |
  |-------------------------------->| ack=1500 (duplicate ACK #3 — "I STILL want 1500!")
  |                                 |
  Sender: 3 duplicate ACKs → retransmit seq=1500 immediately!
  |  seq=1500, len=500 (RETRANSMIT) |
  |-------------------------------->| ack=3500 (got everything now!)
```

**Rule**: After receiving **3 duplicate ACKs** (4 total ACKs for the same byte), the sender retransmits the missing segment immediately, without waiting for the RTO timer.

### Why 3 and not 1?

Packets can be reordered in the network. If the threshold were 1 duplicate ACK, minor reordering would trigger unnecessary retransmissions. Three duplicate ACKs strongly indicate loss, not reordering.

```bash
# Linux has RACK (Recent ACKnowledgment) which is smarter than the 3-dup threshold
# It uses time-based detection instead of counting duplicates
cat /proc/sys/net/ipv4/tcp_recovery
# 1 = RACK enabled (default on modern kernels)
```

---

## SACK

Selective Acknowledgment (RFC 2018) solves the inefficiency of cumulative ACKs.

### The problem SACK solves

With only cumulative ACKs, if bytes 3000-3499 are lost but 3500-5999 arrive:

```
Without SACK:
  Receiver: ack=3000 (all I can say)
  Sender retransmits seq=3000. But what about 3500-5999?
  Maybe retransmit everything? (wasteful)
  Maybe retransmit only 3000-3499? (risky — what if more is missing?)

With SACK:
  Receiver: ack=3000, SACK=[3500-6000]
  Sender: "I need to retransmit only 3000-3499 — everything else arrived"
```

### SACK in the TCP header

SACK is a TCP option (up to 4 SACK blocks):

```
ACK = 3000
SACK blocks:
  [3500-6000]    ← "I have bytes 3500 through 5999"
  [7000-8000]    ← "I also have bytes 7000 through 7999"

Missing: 3000-3499 and 6000-6999
```

### SACK on Linux

```bash
# Check if SACK is enabled
cat /proc/sys/net/ipv4/tcp_sack
# 1 = enabled (default)

# SACK is negotiated during handshake (SYN options)
# Both sides must support it

# D-SACK (Duplicate SACK, RFC 2883) — reports duplicate segments
cat /proc/sys/net/ipv4/tcp_dsack
# 1 = enabled (default)
# Tells the sender "you retransmitted something I already had"
# Helps the sender understand that loss wasn't as bad as it thought
```

---

## Duplicate ACKs

Duplicate ACKs serve multiple purposes:

### 1. Signal packet loss

Three duplicate ACKs trigger fast retransmit (explained above).

### 2. Signal reordering

If only 1-2 duplicate ACKs arrive, the sender knows packets might be reordered but doesn't retransmit yet.

### 3. Keep the pipe full

During fast recovery, each duplicate ACK tells the sender that one more packet has left the network. The sender can send one new packet per duplicate ACK received.

```
Fast Recovery (simplified):
  1. 3 duplicate ACKs → retransmit lost segment, halve cwnd
  2. Each additional dup ACK → send one new segment (inflate cwnd temporarily)
  3. Non-duplicate ACK arrives → deflate cwnd back to halved value
     (This means the receiver got the retransmission and caught up)
```

---

## TCP Segmentation

TCP is a byte-stream protocol. Applications write bytes without caring about segment boundaries. TCP decides how to break the stream into segments.

### Maximum Segment Size (MSS)

MSS = Maximum amount of data in one TCP segment (payload only, not headers).

```
Ethernet MTU:         1500 bytes
IP header:              20 bytes
TCP header:             20 bytes (minimum)
                      ──────────
MSS:                  1460 bytes
```

MSS is negotiated during the handshake and is NOT negotiable upward afterwards.

```bash
# Check MSS on a connection
ss -ti dst google.com
# Shows mss:1460 (or lower for VPN/tunnels)

# Force a specific MSS (useful for debugging)
iptables -A FORWARD -p tcp --tcp-flags SYN,RST SYN -j TCPMSS --set-mss 1360
```

### Segmentation offload

Modern NICs do segmentation in hardware (TSO — TCP Segmentation Offload):

```bash
# Check offload settings
ethtool -k eth0 | grep segmentation
# tcp-segmentation-offload: on
# generic-segmentation-offload: on

# The kernel sends a large chunk to the NIC.
# The NIC splits it into MSS-sized segments and adds headers.
# Much faster than doing it in software.
```

---

## Nagle's Algorithm

### The problem

A telnet session sends one keystroke at a time. Each keystroke = 1 byte of data, but you need 40 bytes of headers (20 IP + 20 TCP). That's 41 bytes to send 1 byte — 97.5% overhead.

### The solution

RFC 896 (Nagle's Algorithm): Buffer small writes. Only send if:
1. There's enough data to fill a segment (MSS), OR
2. All previously sent data has been acknowledged

```
Without Nagle:
  Keystroke 'H' → send immediately (41 bytes for 1 byte of data)
  Keystroke 'e' → send immediately (41 bytes for 1 byte of data)
  Keystroke 'l' → send immediately (41 bytes for 1 byte of data)
  3 packets × 41 bytes = 123 bytes for 3 bytes of data

With Nagle:
  Keystroke 'H' → send immediately (no unACKed data)
  Keystroke 'e' → buffer (waiting for ACK of 'H')
  Keystroke 'l' → buffer (still waiting)
  ACK arrives → send "el" (1 packet)
  2 packets × ~42 bytes = ~84 bytes for 3 bytes of data
```

### When Nagle hurts

For request-response protocols where you write a small request and immediately want to send it:

```python
# This can cause Nagle delays:
sock.send(b"GET ")
sock.send(b"/index.html ")
sock.send(b"HTTP/1.1\r\n")
sock.send(b"\r\n")
# Nagle might buffer the small writes

# Solution 1: Combine into one write
sock.send(b"GET /index.html HTTP/1.1\r\n\r\n")

# Solution 2: Disable Nagle
sock.setsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY, 1)
```

```bash
# In Linux, TCP_NODELAY disables Nagle's algorithm
# Real-time applications (games, trading) always disable it
```

---

## Delayed ACKs

### The idea

Instead of ACKing every segment immediately, wait up to 200ms hoping to piggyback the ACK on a data response.

```
Without delayed ACK:
  Receive data → send ACK immediately (pure ACK, no data, waste a packet)

With delayed ACK:
  Receive data → wait up to 200ms
  If response data to send → piggyback ACK on data packet (saves a packet!)
  If 200ms passes → send ACK anyway
  If 2nd segment arrives → ACK immediately (ACK every other segment)
```

```bash
# Delayed ACK is enabled by default on Linux
# Disable with TCP_QUICKACK:
setsockopt(fd, IPPROTO_TCP, TCP_QUICKACK, &one, sizeof(one));
# Note: TCP_QUICKACK is not sticky — it only affects the next ACK
```

---

## Nagle + Delayed ACK Interaction

This is a classic networking performance bug.

### The deadly scenario

```
Client (Nagle ON) sends two small writes:

Write 1 ("GET "):
  → Sent immediately (no unACKed data)
  → Server receives, enables delayed ACK timer (wait up to 200ms)

Write 2 ("/index.html"):
  → Nagle buffers it (waiting for ACK of Write 1)
  → Server is waiting 200ms before ACKing...
  → Client is waiting for ACK before sending...
  → DEADLOCK for up to 200ms!

After 200ms: Server's delayed ACK timer fires → sends ACK → Client sends Write 2
Result: 200ms of unnecessary latency on every request
```

### Solutions

1. **Combine writes into one** (best solution)
2. **Disable Nagle** (`TCP_NODELAY = 1`) for latency-sensitive protocols
3. **Use writev() / sendmsg()**: Scatter-gather I/O sends multiple buffers as one TCP segment

```bash
# Check if TCP_NODELAY is set on a connection
ss -ti | grep nodelay
# If you see "nodelay" in the output, Nagle is disabled for that connection
```

---

## Linux: Retransmission Debugging

### Monitoring retransmissions

```bash
# TCP retransmission statistics
cat /proc/net/snmp | grep Tcp
# Key fields:
# RetransSegs: Total retransmitted segments (cumulative since boot)
# OutSegs:     Total segments sent

# Retransmission rate
awk '/Tcp:/ {if(NR==8) print "Retransmit rate:", $13/$11 * 100 "%"}' /proc/net/snmp

# Watch retransmissions in real-time
watch -n 1 'nstat -az TcpRetransSegs'

# Per-connection retransmission info
ss -ti
# Shows: rto, rtt, retrans (count), lost, etc.
# Example output:
# rto:204 rtt:1.5/0.5 ... retrans:0/3 ... lost:0
#                           ^current/total retransmit count
```

### Capturing retransmissions with tcpdump

```bash
# Capture retransmitted packets (packets with the same sequence number as a previous packet)
sudo tcpdump -nn -c 100 'tcp[tcpflags] & tcp-syn == 0' -w /tmp/capture.pcap

# Analyze in Wireshark: filter "tcp.analysis.retransmission"
# Or with tshark:
tshark -r /tmp/capture.pcap -Y "tcp.analysis.retransmission" -T fields -e frame.time -e ip.src -e ip.dst -e tcp.seq
```

### Common retransmission scenarios

```bash
# High retransmission rate + high RTT variance:
#   → Network path quality issues (packet loss on intermediate links)

# Retransmissions followed by RST:
#   → Firewall dropping packets (connection can't complete)

# Retransmissions only for specific destination:
#   → Problem on the path to that destination, not local

# Spurious retransmissions (sender retransmits, then receives ACK for original):
#   → RTO too aggressive, or delayed ACKs causing late ACKs
```

### Tuning retransmission behavior

```bash
# Number of retries before giving up
cat /proc/sys/net/ipv4/tcp_retries2
# 15 (default, ~13-30 minutes total)

# Enable thin-stream retransmission (for SSH-like low-bandwidth connections)
cat /proc/sys/net/ipv4/tcp_thin_linear_timeouts
# 0 (default, set to 1 for interactive connections)

# Early retransmit — allows fast retransmit with fewer dup ACKs
# when the window is small
cat /proc/sys/net/ipv4/tcp_early_retrans
# 3 (default)

# RACK loss detection
cat /proc/sys/net/ipv4/tcp_recovery
# 1 (default, RACK enabled)
```

---

## Key Takeaways

1. **Sequence numbers identify every byte** — they're byte-based, not packet-based
2. **Cumulative ACKs** report the highest contiguous byte received
3. **RTO adapts to network conditions** using smoothed RTT and variance; doubles on each retry
4. **Fast retransmit** triggers after 3 duplicate ACKs — much faster than waiting for RTO
5. **SACK** lets the receiver tell the sender exactly what's missing
6. **Nagle reduces overhead** by batching small writes, but adds latency
7. **Delayed ACKs** reduce packet count but interact badly with Nagle
8. **TCP_NODELAY** disables Nagle — essential for latency-sensitive applications
9. **Monitor `TcpRetransSegs`** and per-connection `retrans` count via `ss -ti`

---

## Next

→ [05-tcp-flow-control.md](05-tcp-flow-control.md) — How TCP prevents overwhelming the receiver
