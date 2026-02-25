# HTTP/2 and HTTP/3 — Modern HTTP

> HTTP/1.1 has one fatal flaw: head-of-line blocking. HTTP/2 fixes it with multiplexing over a single TCP connection. HTTP/3 goes further — it replaces TCP entirely with QUIC (UDP-based) to eliminate head-of-line blocking at the transport layer too.

---

## Table of Contents

1. [Why HTTP/1.1 Isn't Enough](#why-http11-isnt-enough)
2. [HTTP/2 Overview](#http2-overview)
3. [HTTP/2 Streams and Multiplexing](#streams-and-multiplexing)
4. [HTTP/2 Header Compression (HPACK)](#hpack)
5. [HTTP/2 Server Push](#server-push)
6. [HTTP/2 Flow Control and Prioritization](#flow-control)
7. [HTTP/2 in Practice](#http2-practice)
8. [Why HTTP/2 Still Has Problems](#http2-problems)
9. [HTTP/3 and QUIC](#http3-quic)
10. [QUIC Deep Dive](#quic-deep-dive)
11. [HTTP/3 vs HTTP/2 vs HTTP/1.1](#comparison)
12. [Linux: Testing and Debugging](#linux-testing)
13. [Key Takeaways](#key-takeaways)

---

## Why HTTP/1.1 Isn't Enough

### The head-of-line blocking problem

HTTP/1.1 requires responses in order on each connection:

```
Connection 1:  req A ──────────────────> resp A (large 2MB image)
                                         req B → resp B  (50 byte JSON, waits!)

B waits behind A even though it's ready in milliseconds.
```

### Browser workarounds (all have costs)

```
1. Multiple connections: 6 parallel TCP connections per hostname
   Cost: 6× TCP+TLS handshakes, 6× slow start, memory overhead

2. Domain sharding: Split assets across cdn1.example.com, cdn2.example.com
   Cost: More DNS lookups, more connections, invalidates connection reuse

3. Asset concatenation: Bundle all CSS into one file, all JS into one file
   Cost: Change one line → invalidate entire bundle cache

4. Image sprites: Combine many icons into one image
   Cost: Complex CSS positioning, change one icon → invalidate sprite cache
```

These are hacks. HTTP/2 solves the root cause.

---

## HTTP/2 Overview

HTTP/2 (RFC 7540, 2015) fundamentally changes how HTTP is transported while keeping the same semantics (same methods, headers, status codes):

| Feature | HTTP/1.1 | HTTP/2 |
|---------|----------|--------|
| Format | Text | Binary |
| Connections per host | 6+ | 1 |
| Multiplexing | No | Yes |
| Header format | Repeated text | Compressed (HPACK) |
| Server push | No | Yes |
| Stream priority | No | Yes |

### Binary framing layer

HTTP/2 introduces a binary framing layer between HTTP semantics and TCP:

```
HTTP/1.1:                           HTTP/2:
┌─────────────────────┐            ┌─────────────────────┐
│ GET /page HTTP/1.1  │            │ HTTP Semantics      │
│ Host: example.com   │            │ (same methods,      │
│ Accept: text/html   │            │  headers, status)   │
├─────────────────────┤            ├─────────────────────┤
│ TCP                 │            │ Binary Framing Layer │
│                     │            │ (streams, frames)    │
│                     │            ├─────────────────────┤
│                     │            │ TCP                 │
└─────────────────────┘            └─────────────────────┘
```

---

## Streams and Multiplexing

### Core concepts

**Frame**: Smallest unit of communication. Each frame has a type (HEADERS, DATA, etc.) and a stream ID.

**Stream**: A bidirectional sequence of frames exchanged between client and server. Each request-response pair is a separate stream.

**Connection**: One TCP connection carries multiple streams.

```
One TCP Connection (to example.com)
├── Stream 1: GET /page.html      → 200 OK + HTML
├── Stream 3: GET /style.css      → 200 OK + CSS
├── Stream 5: GET /app.js         → 200 OK + JS
├── Stream 7: GET /logo.png       → 200 OK + image
└── Stream 9: GET /api/data       → 200 OK + JSON

All streams interleaved on ONE connection!
Stream IDs: odd = client-initiated, even = server-initiated
```

### Frame format

```
+-----------------------------------------------+
|                 Length (24 bits)               |
+------------+----------------------------------+
| Type (8)   | Flags (8)                        |
+-+----------+----------------------------------+
|R|          Stream Identifier (31 bits)        |
+-+---------------------------------------------+
|              Frame Payload (variable)         |
+-----------------------------------------------+

Length:    Size of payload (max 16,384 bytes default, up to 16 MB)
Type:     DATA, HEADERS, PRIORITY, RST_STREAM, SETTINGS, 
          PUSH_PROMISE, PING, GOAWAY, WINDOW_UPDATE, CONTINUATION
Flags:    END_STREAM, END_HEADERS, PADDED, PRIORITY
Stream ID: Which stream this frame belongs to (0 = connection-level)
```

### How multiplexing eliminates HOL blocking

```
HTTP/1.1 (sequential):
Time →
Conn 1: [====A====][==B==][=C=]
         Response A   B     C     (B waits for A, C waits for B)

HTTP/2 (interleaved on one connection):
Time →
Stream 1: [=A=]    [=A=]    [=A=]    [=A=]
Stream 3:    [=B=]    [=B=]
Stream 5:       [C]            [C]
               ↑
               Frames from different streams interleaved
               B and C don't wait for A!
```

### Stream states

```
                       send HEADERS
              idle ──────────────→ open
               │                    │
               │                    ├── send END_STREAM → half-closed (local)
               │                    │                        │
               │                    ├── recv END_STREAM → half-closed (remote)
               │                    │                        │
               │                    │                     recv END_STREAM
               │                    │                        │
               │                    └── both END_STREAM → closed
               │
               └── recv PUSH_PROMISE → reserved (remote)
```

---

## HPACK

### The header redundancy problem

HTTP/1.1 headers are verbose and repeated for every request:

```
Request 1:                          Request 2 (same page, different resource):
GET /page HTTP/1.1                  GET /style.css HTTP/1.1
Host: example.com                   Host: example.com           ← repeated
User-Agent: Mozilla/5.0...          User-Agent: Mozilla/5.0...  ← repeated
Accept: text/html                   Accept: text/css
Accept-Encoding: gzip, br           Accept-Encoding: gzip, br  ← repeated
Accept-Language: en-US               Accept-Language: en-US      ← repeated
Cookie: session=abc123...           Cookie: session=abc123...   ← repeated

~800 bytes of headers per request, 90% repeated
100 requests = 80 KB of redundant headers!
```

### HPACK compression

HPACK uses three techniques:

**1. Static table**: 61 commonly used header name-value pairs pre-defined:

```
Index  Header Name          Header Value
1      :authority           (empty)
2      :method              GET
3      :method              POST
4      :path                /
5      :path                /index.html
6      :scheme              http
7      :scheme              https
8      :status              200
...
61     www-authenticate     (empty)
```

**2. Dynamic table**: Headers seen in this connection are added to a table. Subsequent references use just the index number:

```
Request 1: Send full header → add to dynamic table at index 62
Request 2: Send just "62" instead of full header (1 byte vs 100+ bytes)
```

**3. Huffman encoding**: Header values encoded with Huffman coding (common chars = fewer bits).

Result: Headers compressed by **85-90%** in typical scenarios.

---

## Server Push

The server can send resources the client hasn't requested yet:

```
Traditional:
1. Client: GET /page.html
2. Server: 200 OK (HTML)
3. Client: (parses HTML, discovers style.css is needed)
4. Client: GET /style.css
5. Server: 200 OK (CSS)
   → 2 round trips to get page + CSS

Server Push:
1. Client: GET /page.html
2. Server: PUSH_PROMISE (I'll send /style.css too)
3. Server: 200 OK (HTML), 200 OK (CSS pushed)
   → 1 round trip for both!
```

### Server push in practice

Server push **sounds great** but has significant problems:

1. **Cache invalidation**: Server pushes resources the client already has cached → wasted bandwidth
2. **Priority conflicts**: Pushed resources compete with requested resources
3. **Complexity**: Hard to know what to push without understanding client state
4. **No adoption**: Most CDNs and servers disabled it
5. **Chrome removed support** (2022): The performance gains didn't justify the complexity

**Status**: Effectively dead. Use `<link rel="preload">` or 103 Early Hints instead.

```html
<!-- Modern alternative: preload hints -->
<link rel="preload" href="/style.css" as="style">
<link rel="preload" href="/font.woff2" as="font" crossorigin>

<!-- Or 103 Early Hints (server response before final response) -->
HTTP/1.1 103 Early Hints
Link: </style.css>; rel=preload; as=style

HTTP/1.1 200 OK
Content-Type: text/html
...
```

---

## Flow Control

### Stream-level flow control

HTTP/2 has its own flow control **on top of** TCP's flow control:

```
TCP flow control:  Manages bytes on the connection
HTTP/2 flow control: Manages bytes per stream AND per connection

Why both? Because TCP doesn't know about streams.
Without HTTP/2 flow control, one stream downloading a huge file
could consume all TCP window, starving other streams.
```

Each stream has a flow control window (default: 65,535 bytes). Data frame sender must not exceed the receiver's window. WINDOW_UPDATE frames increase the window.

### Stream prioritization

HTTP/2 allows clients to express priority preferences:

```
Stream 1 (HTML):    weight=256 (highest)
Stream 3 (CSS):     weight=220 (high, blocks rendering)
Stream 5 (JS):      weight=183 (medium)
Stream 7 (images):  weight=110 (lower, not render-blocking)

Server SHOULD respect priorities but doesn't have to.
```

Priority has a dependency tree:

```
         Stream 0 (root)
         ├── Stream 1 (HTML, weight 256)
         │   ├── Stream 3 (CSS, weight 220)
         │   └── Stream 5 (JS, weight 183)
         └── Stream 7 (imgs, weight 110)
```

In practice, priority is poorly implemented across servers and CDNs. HTTP/3 replaces it with a simpler mechanism (Extensible Priorities, RFC 9218).

---

## HTTP/2 in Practice

### Connection setup

HTTP/2 requires TLS in practice (browsers only support HTTP/2 over TLS):

```
1. TCP handshake (SYN, SYN-ACK, ACK)
2. TLS handshake with ALPN extension:
   ClientHello: alpn=[h2, http/1.1]
   ServerHello: alpn=h2
3. HTTP/2 connection preface:
   Client sends: "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n" + SETTINGS frame
   Server sends: SETTINGS frame
4. Requests begin
```

**ALPN** (Application-Layer Protocol Negotiation): TLS extension that negotiates HTTP/2 during TLS handshake (no extra round trip).

### Settings exchange

```
Client SETTINGS:
  HEADER_TABLE_SIZE: 4096          (HPACK dynamic table size)
  ENABLE_PUSH: 0                   (disable server push)
  MAX_CONCURRENT_STREAMS: 100      (max parallel requests)
  INITIAL_WINDOW_SIZE: 65535       (flow control window)
  MAX_FRAME_SIZE: 16384            (max frame payload)
  MAX_HEADER_LIST_SIZE: 8192       (max header size)
```

### Migration from HTTP/1.1

When switching to HTTP/2:

```
UNDO these HTTP/1.1 hacks:
  ✗ Domain sharding → one domain, one connection
  ✗ Asset concatenation → ship individual files (stream each)
  ✗ Image sprites → individual images (concurrent streams)
  ✗ CSS/JS inlining → external files (cacheable, streamable)

KEEP these:
  ✓ Compression (gzip/brotli)
  ✓ Caching headers
  ✓ CDN usage
  ✓ Minimizing payload size
```

---

## Why HTTP/2 Still Has Problems

### TCP head-of-line blocking

HTTP/2 solves HTTP-level HOL blocking but introduces a new problem:

```
HTTP/2 multiplexes streams on ONE TCP connection.
TCP sees all streams as one byte stream.

If TCP segment 3 is lost:
  TCP: "I need to retransmit segment 3. Block ALL data until it arrives."
  
  Stream A: [seg1][seg2] ← waiting for seg3 (which belongs to Stream B!)
  Stream B: [seg3-LOST]  ← retransmitting
  Stream C: [seg4][seg5] ← already received but blocked by TCP!
  
ALL streams blocked because one packet was lost!
```

This is worse than HTTP/1.1's 6 connections in lossy networks:

```
HTTP/1.1: 6 connections, loss on conn 3 blocks only conn 3
HTTP/2:   1 connection, loss blocks ALL streams

On a lossy WiFi/mobile network, HTTP/2 can be SLOWER than HTTP/1.1.
```

### Other HTTP/2 limitations

```
1. TCP slow start: Single connection starts with small cwnd
   (HTTP/1.1's 6 connections = 6× initial cwnd)

2. TCP connection coalescing: Complicated rules for reusing connections
   across domains sharing certificates

3. Middlebox interference: Some firewalls/proxies don't handle HTTP/2 well

4. Server push failure: Good idea, poor real-world results
```

---

## HTTP/3 and QUIC

HTTP/3 (RFC 9114, 2022) replaces TCP with QUIC:

```
HTTP/1.1:  HTTP → TCP → IP
HTTP/2:    HTTP → binary framing → TCP → IP
HTTP/3:    HTTP → QUIC → UDP → IP

QUIC = UDP + reliability + multiplexing + encryption
(QUIC is essentially TLS 1.3 + TCP-like reliability over UDP)
```

### Why QUIC uses UDP

TCP is implemented in the OS kernel. Changing TCP behavior (adding multiplexing awareness) requires OS updates across billions of devices — impossible in practice.

UDP is a thin wrapper. QUIC implements everything in userspace on top of UDP:

```
Kernel updates needed:
  TCP change: Update every OS (Linux, Windows, macOS, Android, iOS, routers)
  → Takes 10+ years for global deployment

  QUIC over UDP: Update application/library only
  → Deploy immediately via app update
```

---

## QUIC Deep Dive

### 0-RTT connection establishment

TCP + TLS 1.3 require 2-3 round trips before data. QUIC can do it in 0:

```
TCP + TLS 1.3 (HTTP/2):
RTT 1: TCP SYN → SYN-ACK
RTT 2: TLS ClientHello → ServerHello + keys
RTT 3: First HTTP request
= 2-3 RTTs before data flows

QUIC (HTTP/3):
New connection:    1 RTT (crypto + transport handshake combined)
Resumed connection: 0 RTT (send data immediately with cached keys)
```

### No head-of-line blocking

QUIC multiplexes streams independently at the transport layer:

```
QUIC Connection
├── Stream 1: [pkt1][pkt2][pkt5]     ← can deliver independently
├── Stream 3: [pkt3-LOST]             ← retransmitting, only stream 3 blocked
└── Stream 5: [pkt4][pkt6]            ← delivered immediately, not blocked!

vs TCP:
Single byte stream: [pkt1][pkt2][pkt3-LOST][pkt4][pkt5][pkt6]
                     Everything after pkt3 blocked until retransmit
```

### Connection migration

TCP connections are identified by (src IP, src port, dst IP, dst port). If any changes (e.g., WiFi → cellular), the connection breaks.

QUIC connections are identified by a Connection ID:

```
TCP on mobile:
  WiFi (192.168.1.10:54321 → server) → walk outside
  → IP changes to cellular (10.0.0.5:?)
  → TCP connection dead → new TCP + TLS + slow start
  → 500ms+ interruption

QUIC on mobile:
  WiFi → Connection ID = abc123
  → Walk outside → cellular
  → Same Connection ID = abc123 → seamless!
  → No interruption, no new handshake
```

### Built-in encryption

QUIC mandates TLS 1.3. Even the transport headers are encrypted:

```
TCP:     Headers visible to middleboxes (ISPs, firewalls can inspect)
QUIC:    Almost everything encrypted
         - Only Connection ID and a few fields visible
         - Middleboxes can't inspect or modify transport headers
         - Prevents ossification (protocol can evolve)
```

### Congestion control

QUIC doesn't mandate a specific congestion control algorithm. Current implementations commonly use:

- **CUBIC**: Default in Linux TCP, also used in QUIC
- **BBR**: Google's bandwidth-based approach (used in Google's QUIC)
- **New algorithms**: Easier to deploy because QUIC is in userspace

---

## Comparison

| Feature | HTTP/1.1 | HTTP/2 | HTTP/3 |
|---------|----------|--------|--------|
| Year | 1997 | 2015 | 2022 |
| Transport | TCP | TCP | QUIC (UDP) |
| Encryption | Optional | Practically required | Mandatory (TLS 1.3) |
| Format | Text | Binary | Binary |
| Multiplexing | No | Yes (but TCP HOL) | Yes (no HOL) |
| Header compression | No | HPACK | QPACK |
| Server push | No | Yes (deprecated) | Yes (rarely used) |
| Connection setup | 3 RTT (TCP+TLS) | 2-3 RTT | 1 RTT (0 RTT resumed) |
| Connection migration | No | No | Yes |
| HOL blocking | Yes (HTTP) | Yes (TCP) | No |
| Connections per host | 6+ | 1 | 1 |

### When does HTTP/3 help most?

```
High latency (satellite, intercontinental):
  → 0-RTT saves 100-300ms

Lossy networks (WiFi, mobile):
  → No TCP HOL blocking = fewer stalls

Mobile users switching networks:
  → Connection migration = seamless transition

Frequent short connections (API calls, IoT):
  → 0-RTT = data on first packet
```

### When does HTTP/3 help least?

```
Low-latency wired networks:
  → RTT savings minimal (1ms vs 3ms)

Long-lived connections:
  → Handshake savings amortized

No packet loss:
  → HOL blocking doesn't trigger

Bulk downloads:
  → TCP performs equally well
```

---

## Linux: Testing and Debugging

### Check HTTP version in use

```bash
# curl: see negotiated protocol
curl -v https://example.com 2>&1 | grep "ALPN"
# output: ALPN: server accepted h2

# Force specific versions
curl --http1.1 https://example.com
curl --http2 https://example.com
curl --http3 https://example.com   # requires curl 8.x with nghttp3

# Check if server supports HTTP/2
curl -sI https://example.com -o /dev/null -w '%{http_version}\n'
# output: 2

# Check if server supports HTTP/3 via Alt-Svc header
curl -sI https://google.com | grep -i alt-svc
# alt-svc: h3=":443"; ma=2592000
```

### nghttp — HTTP/2 debugging tool

```bash
# Install
sudo apt install nghttp2-client

# Make HTTP/2 request (verbose)
nghttp -v https://example.com

# See frame-level details
nghttp -nv https://example.com
# Output shows individual HEADERS, DATA, SETTINGS, WINDOW_UPDATE frames

# Get timing
nghttp -s https://example.com
```

### OpenSSL — check ALPN support

```bash
# Check server's ALPN (protocol negotiation)
echo | openssl s_client -alpn h2 -connect example.com:443 2>/dev/null | grep "ALPN"
# ALPN protocol: h2
```

### tcpdump / Wireshark

```bash
# Capture QUIC (HTTP/3) traffic
sudo tcpdump -i eth0 'udp port 443' -w quic.pcap

# Capture HTTP/2 traffic (over TLS, you'll see encrypted frames)
sudo tcpdump -i eth0 'tcp port 443' -w http2.pcap

# In Wireshark:
# - Filter: http2 (after TLS decryption with SSLKEYLOGFILE)
# - Filter: quic (for HTTP/3 traffic)
```

### Testing with SSLKEYLOGFILE

```bash
# Log TLS keys for Wireshark decryption
export SSLKEYLOGFILE=/tmp/tls-keys.log
curl https://example.com
# or
SSLKEYLOGFILE=/tmp/tls-keys.log firefox

# In Wireshark:
# Edit → Preferences → Protocols → TLS → (Pre)-Master-Secret log filename
# → /tmp/tls-keys.log
# Now you can see decrypted HTTP/2 frames
```

---

## Key Takeaways

1. **HTTP/2 = multiplexing** — many requests/responses interleaved on one TCP connection
2. **Binary framing** replaces text — more efficient parsing, not human-readable on wire
3. **HPACK** compresses headers by 85-90% (static table + dynamic table + Huffman)
4. **Server push is dead** — use `<link rel="preload">` or 103 Early Hints instead
5. **HTTP/2's Achilles heel**: TCP head-of-line blocking — one lost packet blocks ALL streams
6. **HTTP/3 = HTTP over QUIC** — QUIC replaces TCP with independent stream multiplexing
7. **QUIC runs over UDP** — implements reliability in userspace, can evolve without OS updates
8. **0-RTT connection** for resumed connections — critical for mobile and high-latency
9. **Connection migration** — QUIC survives IP address changes (WiFi → cellular)
10. **HTTP/3 helps most** on lossy/high-latency networks; on low-latency wired networks, differences are small

---

## Next

→ [03-websockets-grpc.md](03-websockets-grpc.md) — Beyond request-response: WebSockets, gRPC, and streaming
