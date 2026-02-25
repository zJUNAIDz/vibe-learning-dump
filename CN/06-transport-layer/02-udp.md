# UDP — User Datagram Protocol

> UDP is the "send and forget" protocol. No handshake, no acknowledgments, no ordering, no retransmission. It adds ports to IP and gets out of the way. Understanding UDP is understanding what TCP adds and why.

---

## Table of Contents

1. [Why UDP Exists](#why-udp-exists)
2. [UDP Header (8 Bytes)](#udp-header)
3. [UDP Checksum](#udp-checksum)
4. [UDP vs TCP: The Trade-offs](#udp-vs-tcp)
5. [When UDP Is the Right Choice](#when-udp-is-right)
6. [Protocols Built on UDP](#protocols-built-on-udp)
7. [UDP in Practice on Linux](#udp-on-linux)
8. [QUIC: UDP's Modern Evolution](#quic)
9. [Common Misconceptions](#misconceptions)
10. [Key Takeaways](#key-takeaways)

---

## Why UDP Exists

IP delivers packets between hosts. But IP has no concept of ports — it can't deliver to a specific application. The minimum you need on top of IP is:

1. Source port
2. Destination port

UDP adds exactly this, plus a length field and an optional checksum. Nothing more.

$$\text{UDP} = \text{IP} + \text{Ports} + \text{Length} + \text{Checksum}$$

**Why not just always use TCP?** Because TCP's reliability mechanisms have costs:

- **Latency**: TCP requires a 3-way handshake before data (1.5 RTTs)
- **Head-of-line blocking**: One lost packet blocks all subsequent packets
- **Overhead**: TCP header is 20-60 bytes; UDP header is 8 bytes
- **Complexity**: TCP maintains per-connection state

For many applications, these costs exceed the benefits.

---

## UDP Header

UDP has the **simplest possible transport header**: 8 bytes, 4 fields.

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|          Source Port          |       Destination Port        |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|            Length             |           Checksum            |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                          Data ...                            |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

| Field | Size | Description |
|-------|------|-------------|
| Source Port | 16 bits | Sender's port (optional — can be 0 if no reply needed) |
| Destination Port | 16 bits | Receiver's port |
| Length | 16 bits | Total datagram length (header + data), minimum 8 |
| Checksum | 16 bits | Error detection (optional in IPv4, mandatory in IPv6) |

**Maximum UDP datagram size**: The Length field is 16 bits → max 65,535 bytes. Minus the 8-byte header → max 65,527 bytes of data. But IP fragmentation limits are the real constraint.

**Practical max size**: Most applications keep UDP datagrams below the MTU (1500 bytes - 20 IP header - 8 UDP header = **1472 bytes**) to avoid IP fragmentation.

### Compare with TCP

```
UDP Header:  8 bytes,  4 fields, no state
TCP Header: 20-60 bytes, 10+ fields, lots of state
```

---

## UDP Checksum

The UDP checksum covers:
1. A **pseudo-header** (source IP, dest IP, protocol number, UDP length)
2. The UDP header
3. The UDP data

```
Pseudo-header (for checksum computation only):

+--------+--------+--------+--------+
|          Source Address             |
+--------+--------+--------+--------+
|       Destination Address          |
+--------+--------+--------+--------+
|  zero  |Protocol|   UDP Length     |
+--------+--------+--------+--------+
```

### Why include the pseudo-header?

The pseudo-header includes IP addresses to detect misdelivered packets. Without it, a packet corrupted in transit could arrive at the wrong host, pass the checksum (which only covers the UDP portion), and be accepted by the wrong application.

### IPv4 vs IPv6

- **IPv4**: UDP checksum is optional. Setting it to 0 means "not computed."
- **IPv6**: UDP checksum is mandatory. IPv6 removed the IP header checksum, so upper layers must protect themselves.

---

## UDP vs TCP

| Property | UDP | TCP |
|----------|-----|-----|
| Connection setup | None | 3-way handshake (1.5 RTT) |
| Reliability | None — packets can be lost | Guaranteed delivery with retransmission |
| Ordering | None — packets arrive in any order | Strict byte-stream ordering |
| Flow control | None | Receiver window |
| Congestion control | None | AIMD / slow start |
| Header size | 8 bytes | 20-60 bytes |
| Communication model | Datagram (message-based) | Stream (byte-based) |
| State | Stateless | Stateful |
| Overhead | Minimal | Significant |

### The key difference: message boundaries

**UDP preserves message boundaries**. If you send 3 messages of 100 bytes each, the receiver gets exactly 3 messages of 100 bytes.

**TCP is a byte stream**. If you send 3 messages of 100 bytes, the receiver might get them as one 300-byte chunk, or five 60-byte chunks, or any other combination. TCP doesn't know or care about "messages."

```
UDP send:  [100 bytes] [100 bytes] [100 bytes]
UDP recv:  [100 bytes] [100 bytes] [100 bytes]  ← preserved

TCP send:  [100 bytes] [100 bytes] [100 bytes]
TCP recv:  [200 bytes] [100 bytes]               ← merged
```

This is why application-layer protocols on TCP need message framing (HTTP uses Content-Length or chunked encoding).

---

## When UDP Is the Right Choice

### 1. Real-time audio/video (VoIP, video calls, live streaming)

**Why**: A 250ms-old audio sample is worthless. Retransmitting it would arrive even later. It's better to play silence and move on.

```
TCP for voice call:
  Packet 5 lost → TCP retransmits → packets 6,7,8 wait (head-of-line blocking)
  → 200ms pause → plays packet 5 (now stale) → gap in audio

UDP for voice call:
  Packet 5 lost → play packets 6,7,8 normally
  → tiny glitch → audio continues smoothly
```

### 2. DNS queries

**Why**: A DNS query is typically a single small packet. The TCP handshake (3 packets) would triple the overhead. If the UDP response is lost, the client simply retries.

```
DNS over UDP: Query → Response (2 packets, 1 RTT)
DNS over TCP: SYN → SYN-ACK → ACK → Query → Response → FIN... (7+ packets, 3+ RTTs)
```

Note: DNS uses TCP for zone transfers and responses > 512 bytes (or > ~1232 bytes with EDNS0).

### 3. Gaming (fast-paced multiplayer)

**Why**: Player positions update 60+ times per second. If position update #47 is lost, retransmitting it when you already have updates #48-52 is pointless.

### 4. DHCP

**Why**: The client doesn't have an IP address yet, so it can't establish a TCP connection. UDP's stateless nature works perfectly for bootstrap protocols.

### 5. Metrics/telemetry collection

**Why**: Sending 10,000 metrics per second. If 0.01% are lost, the statistical impact is negligible. TCP's reliability overhead would be wasteful.

### 6. When you want to implement your own reliability

**Why**: Sometimes TCP's reliability model doesn't match your needs. QUIC, Google's replacement for TCP+TLS, is built on UDP with custom reliability, ordering, and encryption.

---

## Protocols Built on UDP

| Protocol | Port | Why UDP? |
|----------|------|----------|
| DNS | 53 | Small queries, faster without handshake |
| DHCP | 67/68 | Client has no IP yet |
| TFTP | 69 | Simplicity (adds its own reliability) |
| SNMP | 161/162 | Network monitoring, fault tolerance |
| NTP | 123 | Time sync — accuracy matters, not guaranteed delivery |
| RTP | dynamic | Real-time media — latency matters more than reliability |
| QUIC | 443 | Modern HTTP/3 — implements reliability above UDP |
| WireGuard | 51820 | VPN tunnel — handles reliability at tunnel level |
| mDNS | 5353 | Local service discovery |
| syslog | 514 | Log shipping — some loss acceptable |

---

## UDP on Linux

### Sending UDP packets

```bash
# Send a UDP packet with nc (netcat)
echo "Hello UDP" | nc -u -w1 192.168.1.100 8080

# Send DNS query (UDP port 53)
dig @8.8.8.8 google.com

# Listen for UDP on port 8080
nc -ul 8080
# In another terminal:
echo "test" | nc -u -w1 localhost 8080
```

### Monitoring UDP traffic

```bash
# Show UDP listening sockets
ss -ulnp

# Show UDP statistics
cat /proc/net/udp

# UDP protocol statistics
cat /proc/net/snmp | grep Udp
# Udp: InDatagrams NoPorts InErrors OutDatagrams RcvbufErrors SndbufErrors
# Udp: 12345        678     0        9876         0             0

# Key counters:
# InDatagrams:  Successfully received
# NoPorts:      Received for ports with no listener (generates ICMP Port Unreachable)
# InErrors:     Failed to deliver (buffer full, checksum error)
# RcvbufErrors: Dropped because receive buffer was full
```

### UDP buffer tuning

```bash
# Default UDP receive buffer
cat /proc/sys/net/core/rmem_default
# 212992 (208 KB)

# Maximum UDP receive buffer
cat /proc/sys/net/core/rmem_max
# 212992

# Increase for high-throughput UDP
sudo sysctl -w net.core.rmem_max=26214400     # 25 MB
sudo sysctl -w net.core.rmem_default=26214400  # 25 MB

# Per-socket (in application code):
# setsockopt(fd, SOL_SOCKET, SO_RCVBUF, &size, sizeof(size));
```

### Detecting UDP packet loss

```bash
# Watch for receive buffer errors (packets dropped because buffer was full)
watch -n 1 'cat /proc/net/snmp | grep Udp'

# If RcvbufErrors keeps increasing, your application isn't reading fast enough
# or the buffer is too small

# Per-socket drops
ss -uamp
# Shows drop count per socket
```

---

## QUIC

QUIC (Quick UDP Internet Connections) deserves mention because it challenges the "use TCP for reliability" paradigm.

### Why build on UDP instead of creating a new protocol?

**Middlebox ossification**: Firewalls, NATs, and routers understand TCP and UDP. They drop unknown protocols. Deploying a new transport protocol on the Internet is effectively impossible. UDP passes through everything.

### What QUIC adds over UDP

```
Traditional:  TCP + TLS + HTTP/2
QUIC:         UDP + (QUIC handles reliability + encryption + multiplexing)
```

- **0-RTT connection establishment** (vs TCP+TLS = 3 RTT)
- **Independent streams** within one connection (no head-of-line blocking)
- **Connection migration** (survives IP address changes — mobile roaming)
- **Built-in encryption** (always encrypted, unlike TCP where TLS is optional)
- **Userspace implementation** (faster iteration than kernel TCP)

```
TCP + TLS handshake (3 round trips for new connection):
  Client                    Server
    |--- SYN ------------------>|
    |<-- SYN-ACK ---------------|
    |--- ACK ------------------>|  (TCP done: 1.5 RTT)
    |--- ClientHello ---------->|
    |<-- ServerHello ------------|  (TLS done: 1 more RTT)
    |--- Data ----------------->|
    Total: 2-3 RTTs before data

QUIC (1 round trip for new, 0 for repeat):
  Client                    Server
    |--- Initial (crypto+data)->|
    |<-- Handshake + Data ------|
    Total: 1 RTT (or 0 RTT for repeat connections)
```

---

## Misconceptions

### "UDP is unreliable, so it's bad"

UDP is unreliable **by design**. Reliability is a choice, not a requirement. For many applications, unreliability is the correct choice.

### "UDP is faster than TCP"

Not inherently. Both use IP. The speed difference comes from:
- No handshake latency (fewer RTTs to start)
- No head-of-line blocking (lost packets don't delay others)
- No congestion control (can blast as fast as it wants — this is a double-edged sword)

### "UDP packets are guaranteed to arrive intact or not at all"

Mostly true thanks to checksums, but in IPv4, the UDP checksum is optional. Without it, a corrupted packet could be delivered. Always enable checksums.

### "You can send unlimited data over UDP"

You can send without congestion control, but:
- Your ISP might rate-limit or drop
- Intermediate routers will drop when buffers overflow
- You can congest your own link

---

## Key Takeaways

1. **UDP = IP + Ports** — it's the thinnest possible transport layer
2. **8-byte header** — source port, dest port, length, checksum
3. **No connection, no state, no guarantees** — send and forget
4. **Preserves message boundaries** — unlike TCP's byte-stream model
5. **Use when**: real-time, latency-sensitive, loss-tolerant, small/simple queries, or implementing custom reliability
6. **QUIC builds on UDP** to get TCP's reliability without its limitations
7. **Monitor `RcvbufErrors`** in `/proc/net/snmp` for UDP packet drops on Linux

---

## Next

→ [03-tcp-handshake-teardown.md](03-tcp-handshake-teardown.md) — TCP's connection lifecycle
