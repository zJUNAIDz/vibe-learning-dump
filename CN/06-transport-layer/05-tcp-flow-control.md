# TCP Flow Control — Preventing Receiver Overload

> The sender can transmit faster than the receiver can process. Without flow control, the receiver's buffer overflows, packets are dropped, and retransmissions waste bandwidth. TCP's flow control mechanism lets the receiver say "slow down" or "stop" — and the sender obeys.

---

## Table of Contents

1. [The Problem: Speed Mismatch](#the-problem)
2. [Receive Window (rwnd)](#receive-window)
3. [How Flow Control Works](#how-it-works)
4. [Window Scaling](#window-scaling)
5. [Zero Window and Window Probes](#zero-window)
6. [Silly Window Syndrome](#silly-window-syndrome)
7. [Flow Control vs Congestion Control](#flow-vs-congestion)
8. [Send Buffer and Receive Buffer](#buffers)
9. [Linux: Flow Control in Practice](#linux-practice)
10. [Key Takeaways](#key-takeaways)

---

## The Problem

Consider: A 10 Gbps server sends data to a Raspberry Pi with 100 Mbps ethernet and limited RAM. Without flow control:

```
Server:  Sends 1 GB/sec
Pi:      Processes 12 MB/sec
Buffer:  Fills in 0.02 seconds
Result:  Packets dropped, retransmitted, dropped again → collapse
```

Or more commonly: A fast server sends data to a mobile app that's busy doing UI rendering and can't read from the socket fast enough.

**Flow control exists so the receiver can tell the sender exactly how much data it can handle.**

---

## Receive Window

The **receive window (rwnd)** is a 16-bit field in every TCP header that tells the sender: "I have this many bytes of buffer space available. Don't send more than this."

```
TCP Header:
  ...
  Window Size: 16384    ← "I can accept 16,384 more bytes"
  ...
```

Every ACK includes the current window size. The sender uses this to limit outstanding (unacknowledged) data:

$$\text{Data in flight} \leq \min(\text{rwnd}, \text{cwnd})$$

Where `cwnd` is the congestion window (covered in Module 07).

### How the receiver maintains the window

The receiver has a buffer (the TCP receive buffer). As data arrives, it fills the buffer. As the application reads data, it drains the buffer. The window advertised = buffer space remaining.

```
Receive buffer (64 KB):

Before any data:
  [                    64 KB free                    ]
  Window advertised: 65535

After receiving 20 KB (app hasn't read yet):
  [===== 20 KB data =====|        44 KB free        ]
  Window advertised: 45056

After app reads 15 KB:
  [= 5 KB |              59 KB free                  ]
  Window advertised: 60416

After receiving 50 KB more:
  [======== 55 KB data ========|   9 KB free         ]
  Window advertised: 9216
```

---

## How It Works

### Step by step flow control

```
Sender                              Receiver (buffer=32KB)
  |                                   |
  |  seq=0, len=16KB                 | Buffer: [16KB data | 16KB free]
  |--------------------------------->| ack=16384, win=16384
  |                                   |
  |  seq=16384, len=16KB             | Buffer: [32KB data | 0 free]
  |--------------------------------->| ack=32768, win=0       ← STOP!
  |                                   |
  |  (sender stops, can't send)      |
  |                                   |
  |  (app reads 8KB from buffer)     | Buffer: [24KB data | 8KB free]
  |          ack=32768, win=8192     |
  |<---------------------------------| ← "You can send 8KB more"
  |                                   |
  |  seq=32768, len=8KB              | Buffer: [32KB data | 0 free]
  |--------------------------------->| ack=40960, win=0
  |                                   |
```

### The sliding window

The "window" slides forward as data is acknowledged and buffer space is freed:

```
Bytes: 0    10K   20K   30K   40K   50K   60K   70K
       |     |     |     |     |     |     |     |
       [sent+ACKed][sent,not ACKed][ can send ][can't send]
       [          ][    in flight  ][ window  ][          ]
                   ^               ^          ^
                   send_una        send_nxt   send_una + window
```

Three categories:
1. **Sent and acknowledged**: Done, buffers can be freed
2. **Sent but not acknowledged**: In flight, waiting for ACK
3. **Can send**: Within the window, but not yet sent
4. **Can't send**: Beyond the window, must wait

---

## Window Scaling

### The problem with 16 bits

The Window field in the TCP header is 16 bits → max window = 65,535 bytes (64 KB).

The **bandwidth-delay product (BDP)** formula shows why this is inadequate:

$$BDP = \text{Bandwidth} \times \text{RTT}$$

For a 1 Gbps link with 50ms RTT:

$$BDP = 1,000,000,000 \text{ bits/s} \times 0.05 \text{ s} = 50,000,000 \text{ bits} = 6.25 \text{ MB}$$

To fully utilize this link, you need 6.25 MB in flight. A 64 KB window uses only 1% of the link capacity!

### The solution: Window Scale option (RFC 1323)

During the SYN handshake, both sides can negotiate a **window scale factor** (0-14). The actual window = header window × 2^scale.

```
SYN: Window Scale = 7

Actual window = Header Window × 2^7 = Header Window × 128

If header says window = 512:
  Actual window = 512 × 128 = 65,536 bytes

If header says window = 65535:
  Actual window = 65,535 × 128 = 8,388,480 bytes (~8 MB)
```

Maximum possible window:

$$65535 \times 2^{14} = 1,073,725,440 \text{ bytes} \approx 1 \text{ GB}$$

### Linux window scaling

```bash
# Check if window scaling is enabled (it is by default)
cat /proc/sys/net/ipv4/tcp_window_scaling
# 1

# See active window scale on connections
ss -ti
# Shows: wscale:7,7 (sender scale, receiver scale)

# The kernel automatically chooses the scale factor based on buffer sizes
```

---

## Zero Window and Window Probes

### Zero window

When the receiver's buffer is completely full, it advertises `window = 0`. The sender MUST stop.

```
Receiver: ack=50000, win=0
  → "My buffer is full. Do not send anything."

Sender: stops transmitting.
```

### The problem: window update loss

What if the receiver frees buffer space and sends a window update (`win=8192`), but that packet is lost?

```
Receiver: ack=50000, win=0      → Sender stops
Receiver: (app reads data)
Receiver: ack=50000, win=8192   → LOST! Sender never gets it
...
Sender waiting for window update
Receiver waiting for data
DEADLOCK
```

### Window probes (persist timer)

To prevent deadlock, the sender periodically sends **window probe** packets — tiny segments (1 byte or even zero-length) that trigger the receiver to respond with its current window size.

```
Sender                              Receiver
  |  (zero window received)         |
  |  ... waiting ...                 |
  |  Window Probe (1 byte)          |
  |-------------------------------->|
  |                                  |
  |  ack=50001, win=16384           |
  |<--------------------------------| "Buffer freed! Send more!"
  |                                  |
  |  Resumes sending                |
```

The probe interval starts at RTO and doubles each time (exponential backoff), up to 60 seconds.

```bash
# Monitor zero window events
watch -n 1 'nstat -az | grep -i "ZeroWindow\|Persist\|Prune"'

# In ss output, "persist" timer indicates zero-window probe mode
ss -ti | grep persist
```

---

## Silly Window Syndrome

### The problem

If the receiver advertises tiny windows (like 1 byte), the sender sends tiny segments. This creates extreme overhead (40 bytes of headers for 1 byte of data):

```
Receiver reads 1 byte → advertises win=1
Sender sends 1 byte (41 byte packet for 1 byte payload)
Receiver reads 1 byte → advertises win=1
...repeat forever
```

### Receiver-side solution: Clark's algorithm

Don't advertise a small window. Wait until the window is at least:
- MSS (typically 1460 bytes), OR
- Half of the buffer size

Whichever is less. Until then, advertise `window = 0`.

```
Instead of:
  win=1 → win=2 → win=5 → win=1 → ...  (tiny windows)

Clark's algorithm:
  win=0 → win=0 → win=0 → win=1460    (wait until meaningful)
```

### Sender-side solution: Nagle's algorithm

Don't send small segments when there's unacknowledged data outstanding (already covered in the retransmission chapter).

### Both sides cooperate

```
Without SWS prevention:
  Receiver: "I have 1 byte free"       Sender: sends 1 byte
  Receiver: "I have 1 byte free"       Sender: sends 1 byte
  → 40:1 overhead ratio, terrible

With SWS prevention:
  Receiver: "window = 0" (only 1 byte free, not worth advertising)
  Receiver: "window = 0" (10 bytes free, still not worth it)
  Receiver: "window = 1460" (MSS bytes free, now it's worth it!)
  Sender: sends 1460 bytes
  → 40:1460 overhead ratio, excellent
```

---

## Flow Control vs Congestion Control

These are different mechanisms solving different problems:

| | Flow Control | Congestion Control |
|---|---|---|
| **Problem** | Receiver too slow | Network too slow |
| **Who signals** | Receiver (via window field) | Network (via packet loss / ECN) |
| **Mechanism** | Advertised window (rwnd) | Congestion window (cwnd) |
| **Scope** | End-to-end (receiver's buffer) | Network path (router queues) |
| **Speed limit** | Receiver's processing speed | Available network bandwidth |

TCP uses both simultaneously. The effective window is:

$$\text{Effective Window} = \min(\text{rwnd}, \text{cwnd})$$

```
Example:
  rwnd = 64 KB  (receiver can handle 64 KB)
  cwnd = 16 KB  (network can handle 16 KB without congestion)
  → Sender limited to 16 KB in flight (congestion is the bottleneck)

Another scenario:
  rwnd = 4 KB   (slow receiver)
  cwnd = 200 KB (fast network with no congestion)
  → Sender limited to 4 KB in flight (receiver is the bottleneck)
```

---

## Buffers

### Receive buffer

The OS allocates a receive buffer for each TCP connection. Incoming data sits here until the application reads it.

```bash
# Default receive buffer size
cat /proc/sys/net/core/rmem_default
# 212992 (208 KB)

# Maximum receive buffer size
cat /proc/sys/net/core/rmem_max
# 212992

# TCP-specific auto-tuning range: min, default, max
cat /proc/sys/net/ipv4/tcp_rmem
# 4096  131072  6291456
# min   default  max
# 4KB   128KB    6MB
```

### Send buffer

The sender also has a buffer. Data from `write()` goes into the send buffer. TCP sends from the buffer according to the window.

```bash
# Send buffer settings
cat /proc/sys/net/ipv4/tcp_wmem
# 4096  16384  4194304
# min   default  max

# If the send buffer is full, write() blocks (or returns EAGAIN for non-blocking)
```

### Auto-tuning

Linux auto-tunes buffer sizes based on memory pressure and connection BDP:

```bash
# Enable auto-tuning (enabled by default)
cat /proc/sys/net/ipv4/tcp_moderate_rcvbuf
# 1

# The kernel dynamically adjusts receive buffer between tcp_rmem min and max
# based on:
# - Available system memory
# - Measured BDP of the connection  
# - Memory pressure from other connections
```

### Memory pressure

```bash
# TCP memory limits (in pages, 1 page = 4096 bytes)
cat /proc/sys/net/ipv4/tcp_mem
# 382926  510568  765852
# low      pressure  high

# Below 'low': no memory pressure, buffers grow freely
# Above 'pressure': start constraining new allocations
# Above 'high': drop packets, reject connections

# Current TCP memory usage (in pages)
cat /proc/net/sockstat | grep TCP
# TCP: inuse 45 orphan 0 tw 120 alloc 60 mem 15
#                                         ^^^
#                                         Pages allocated
```

---

## Linux Practice

### Viewing flow control state per connection

```bash
# ss -ti shows detailed TCP info
ss -ti dst google.com

# Key fields:
# rcv_space:     Current receive window
# snd_wnd:       Current send window (from peer's rwnd + scaling)
# rcv_ssthresh:  Receive slow-start threshold (auto-tuning)
# wscale:7,7:    Window scale factors (local, remote)

# Example:
# ESTAB  
#   skmem:(r0,rb131072,t0,tb87040,f0,w0,o0,bl0,d0)
#   ts sack cubic wscale:7,7 rto:204 rtt:1.5/0.75
#   rcv_space:29200 rcv_ssthresh:29200 snd_wnd:65536
```

### Decoding skmem

```bash
# skmem:(r0,rb131072,t0,tb87040,f0,w0,o0,bl0,d0)
#
# r:  receive queue bytes (data waiting for application to read)
# rb: receive buffer size (allocated)
# t:  transmit queue bytes (data waiting to be sent/ACKed)
# tb: transmit buffer size (allocated)
# f:  forward allocated memory
# w:  write queue bytes (scheduled but not sent)
# o:  option memory (TCP header options)
# bl: backlog queue bytes
# d:  dropped packets
```

### Detecting flow control problems

```bash
# If 'r' in skmem is consistently near 'rb', the application is reading too slowly
# The receiver is advertising small/zero window → flow control is throttling the sender

# Check for zero-window events
nstat -az | grep -i "zero"
# TCPAbortOnMemory   0
# TCPZeroWindowDrop  0  ← Zero-window drops

# Check if application is reading fast enough
# If RECV-Q in ss is consistently high, the app is slow:
ss -tn | awk '$2 > 0 {print $0}'
# This shows connections where receive queue has unread data
```

### Tuning for high-throughput transfers

```bash
# For a 10 Gbps link with 20ms RTT:
# BDP = 10 Gbps × 20ms = 25 MB
# Need buffers ≥ 25 MB

# Increase maximum buffer sizes
sudo sysctl -w net.core.rmem_max=67108864      # 64 MB
sudo sysctl -w net.core.wmem_max=67108864      # 64 MB
sudo sysctl -w net.ipv4.tcp_rmem="4096 131072 67108864"
sudo sysctl -w net.ipv4.tcp_wmem="4096 131072 67108864"

# Verify auto-tuning is on
sudo sysctl -w net.ipv4.tcp_moderate_rcvbuf=1
```

### Testing flow control behavior

```bash
# Demonstrate flow control with iperf3 and artificial bottleneck:

# Terminal 1: Start iperf3 server with small receive buffer
iperf3 -s --rcv-space-set 16384   # 16 KB receive buffer

# Terminal 2: Client sends rapidly
iperf3 -c localhost -t 10

# Terminal 3: Watch the connection
watch -n 0.5 "ss -ti sport = :5201 | grep -E 'rcv_space|snd_wnd|wscale'"
# You'll see snd_wnd oscillate as the small buffer fills and drains
```

---

## Key Takeaways

1. **Flow control protects the receiver** from being overwhelmed by a fast sender
2. **The receive window (rwnd)** in every ACK tells the sender how much buffer space remains
3. **Window scaling** extends the 16-bit window to ~1 GB using a scale factor negotiated during handshake
4. **Zero window** means "stop sending" — the persist timer prevents deadlock
5. **Silly window syndrome** wastes bandwidth with tiny segments — solved by Clark's algorithm (receiver) and Nagle's algorithm (sender)
6. **Flow control ≠ congestion control**: rwnd protects the receiver; cwnd protects the network
7. **Effective window = min(rwnd, cwnd)** — the tighter constraint wins
8. **Linux auto-tunes** buffer sizes based on BDP and memory pressure
9. **Monitor with `ss -ti`** — check `rcv_space`, `snd_wnd`, and `skmem` for flow control issues

---

## Next

→ [../07-congestion-control/01-congestion-collapse.md](../07-congestion-control/01-congestion-collapse.md) — What happens when TCP doesn't control congestion
