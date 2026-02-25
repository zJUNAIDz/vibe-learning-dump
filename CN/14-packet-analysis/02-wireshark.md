# Wireshark — Visual Packet Analysis

> Wireshark is the world's most popular network protocol analyzer. It provides a graphical interface for deep packet inspection, protocol dissection, stream reassembly, and statistical analysis. While tcpdump captures and tshark analyzes from the CLI, Wireshark lets you visually explore packet captures with unmatched depth.

---

## Table of Contents

1. [When to Use Wireshark vs tcpdump](#when)
2. [The Wireshark Interface](#interface)
3. [Capture and Display Filters](#filters)
4. [Protocol Dissection](#dissection)
5. [Following Streams](#streams)
6. [TCP Analysis Features](#tcp-analysis)
7. [TLS Decryption](#tls)
8. [DNS Analysis](#dns)
9. [Statistics and Graphs](#statistics)
10. [Expert Information](#expert)
11. [Coloring Rules](#coloring)
12. [Practical Analysis Workflows](#workflows)
13. [Key Takeaways](#key-takeaways)

---

## When to Use Wireshark vs tcpdump

```
tcpdump:
  ✓ Available on servers (no GUI needed)
  ✓ Lightweight, fast capture
  ✓ Scriptable
  ✓ Capture in production
  ✗ Limited protocol dissection
  ✗ Hard to analyze complex flows

Wireshark:
  ✓ Deep protocol dissection (3000+ protocols)
  ✓ Visual flow analysis  
  ✓ Automatic problem detection (expert info)
  ✓ TCP stream reassembly
  ✓ TLS decryption with key files
  ✓ Statistics, graphs, conversations
  ✗ Requires GUI (desktop)
  ✗ Not for production capture (too heavy)

Typical workflow:
  1. Capture on server: tcpdump -i eth0 -w capture.pcap
  2. Transfer to desktop: scp server:capture.pcap .
  3. Analyze in Wireshark: wireshark capture.pcap
```

---

## The Wireshark Interface

### Main components

```
┌──────────────────────────────────────────────────────────────┐
│  Menu Bar: File, Edit, View, Go, Capture, Analyze, Statistics│
├──────────────────────────────────────────────────────────────┤
│  Filter Bar: [tcp.port == 443 and !tcp.analysis.retransm...] │ ← Display filter
├──────────────────────────────────────────────────────────────┤
│  Packet List Pane (top):                                      │
│  No. | Time    | Source      | Destination  | Protocol | Info  │
│  1   | 0.000   | 10.0.0.2    | 93.184.216.34| TCP      | SYN  │
│  2   | 0.012   | 93.184.216.34| 10.0.0.2   | TCP      | SYN-ACK│
│  3   | 0.013   | 10.0.0.2    | 93.184.216.34| TCP      | ACK  │
│  4   | 0.014   | 10.0.0.2    | 93.184.216.34| TLS      | ClientHello│
├──────────────────────────────────────────────────────────────┤
│  Packet Details Pane (middle):                                │
│  ▶ Frame 4: 283 bytes on wire (2264 bits)                    │
│  ▶ Ethernet II, Src: 52:54:00:12:34:56, Dst: 52:54:00:ab:cd│
│  ▶ Internet Protocol Version 4, Src: 10.0.0.2, Dst: 93.184.│
│  ▼ Transmission Control Protocol, Src Port: 54321, Dst Port:│
│    Source Port: 54321                                         │
│    Destination Port: 443                                      │
│    Sequence Number: 1 (relative)                              │
│    [TCP Segment Len: 229]                                     │
│  ▼ Transport Layer Security                                   │
│    ▼ TLS Record Layer: Handshake Protocol: Client Hello       │
│      Content Type: Handshake (22)                             │
│      Version: TLS 1.0 (0x0301)                               │
├──────────────────────────────────────────────────────────────┤
│  Packet Bytes Pane (bottom):                                  │
│  0000  52 54 00 ab cd ef 52 54 00 12 34 56 08 00 45 00  RT....│
│  0010  01 0f 4a 5b 40 00 40 06 ...                            │
└──────────────────────────────────────────────────────────────┘
```

### Navigation shortcuts

```
Ctrl+G:        Go to packet number
Ctrl+F:        Find packet (by display filter, string, or hex)
Ctrl+Shift+N:  Next packet in same conversation
Ctrl+Right:    Next protocol field
Space:         Toggle packet details section
Right-click → Follow → TCP Stream:  Reassemble full conversation

Time display:
  View → Time Display Format → Seconds Since First Packet
  (Better for analysis than wall clock time)
```

---

## Capture and Display Filters

### Capture filters (BPF — same as tcpdump)

```
Applied DURING capture. Packets not matching are NEVER captured.
Use these to limit capture size.

host 10.0.0.5
port 443
tcp
not port 22
host 10.0.0.5 and port 80
```

### Display filters (Wireshark-specific)

```
Applied AFTER capture. All packets captured, display filtered.
Much more expressive than BPF.

# By protocol
http
dns
tls
arp
icmp
tcp
udp

# By field
ip.addr == 10.0.0.5
ip.src == 10.0.0.5
ip.dst == 93.184.216.34
tcp.port == 443
tcp.srcport == 54321
udp.port == 53

# Comparisons
tcp.len > 0                          # TCP segments with data
frame.len > 1000                     # Large frames
ip.ttl < 10                         # Low TTL (dying packets)

# String matching
http.host contains "api"
http.request.uri contains "/login"
dns.qry.name == "example.com"

# TCP analysis (automatic detection)
tcp.analysis.retransmission
tcp.analysis.fast_retransmission
tcp.analysis.duplicate_ack
tcp.analysis.zero_window
tcp.analysis.out_of_order
tcp.analysis.lost_segment
tcp.analysis.ack_rtt > 0.5          # Slow ACKs (> 500ms)

# Logical operators
ip.addr == 10.0.0.5 and tcp.port == 443
http.request or http.response
!(arp or dns)                        # Exclude ARP and DNS
tcp.flags.syn == 1 and tcp.flags.ack == 0   # SYN only (new connections)

# Membership
tcp.port in {80, 443, 8080, 8443}
http.response.code in {400..499}     # All 4xx errors
```

### Filter tips

```
Green filter bar:  Valid filter
Red filter bar:    Invalid filter syntax

Right-click any field → "Apply as Filter" → automatically creates filter
Type partial filter → auto-complete suggests options

Useful compound filters:
  tcp.analysis.retransmission or tcp.analysis.duplicate_ack
    → All retransmission-related events

  http.response.code >= 400
    → All HTTP errors

  dns.flags.rcode != 0
    → All DNS failures

  tcp.flags.reset == 1
    → All connection resets

  tls.handshake.type == 1
    → All TLS ClientHellos (shows SNI)
```

---

## Protocol Dissection

### How Wireshark dissects packets

```
Wireshark applies protocol dissectors in order:

Frame (raw bytes)
  → Ethernet (or other L2)
    → IPv4 / IPv6
      → TCP / UDP / ICMP
        → HTTP / TLS / DNS / SMTP / ...
          → Sub-protocols (HTTP/2 frames, gRPC, etc.)

Each dissector understands its protocol's header format
and extracts meaningful fields for display and filtering.
```

### Useful dissector features

```
TCP:
  - Relative sequence numbers (easier than absolute)
  - Stream index (groups packets by connection)
  - Window size with scaling applied
  - RTT calculation per ACK

HTTP:
  - Request/response pairing
  - Full method, URI, headers, body
  - Chunked transfer decoding
  - Compressed content (gzip/deflate) display

TLS:
  - Handshake message types
  - Cipher suite names (not just numbers)
  - Certificate chain display
  - SNI extraction
  - Decryption with key file

DNS:
  - Full query and response parsing
  - DNSSEC validation
  - Response code translation (NXDOMAIN, SERVFAIL)
```

---

## Following Streams

### TCP stream reassembly

```
Right-click packet → Follow → TCP Stream

Wireshark reassembles the entire conversation:
  - Red text   = client → server
  - Blue text  = server → client
  - Shows complete request/response exchange
  - For HTTP: full headers and body visible

Example output:
  GET /api/users HTTP/1.1
  Host: api.example.com
  Accept: application/json
  
  HTTP/1.1 200 OK
  Content-Type: application/json
  Content-Length: 234
  
  {"users": [{"id": 1, "name": "John"}, ...]}

Dropdown at bottom: "Stream 0", "Stream 1", ...
  → Navigate between different TCP conversations
```

### HTTP stream

```
Follow → HTTP Stream
  - Decompresses gzip/deflate content
  - Shows request and response paired
  - For HTTP/2: individual streams

Follow → TLS Stream
  - Shows encrypted data (opaque unless decryption key provided)
  - With SSLKEYLOGFILE: shows decrypted content
```

---

## TCP Analysis Features

### Automatic problem detection

Wireshark colors and annotates TCP problems automatically:

```
[TCP Retransmission]:
  Same data sent again. Original was likely lost.
  Dark red shading in default coloring.

[TCP Fast Retransmission]:
  Retransmission triggered by 3 duplicate ACKs (not timeout).
  Faster recovery than timeout-based retransmit.

[TCP Duplicate ACK]:
  Receiver sent the same ACK again → signals gap in received data.
  Triggers fast retransmit after 3 duplicates.

[TCP Out-Of-Order]:
  Segment arrived with lower sequence number than previous.
  Common with multipath or load-balanced links.

[TCP Zero Window]:
  Receiver advertising window size = 0.
  "Stop sending, my buffer is full!"

[TCP Window Update]:
  Receiver advertising increased window.
  "I've drained some buffer, you can send more."

[TCP Previous Segment Not Captured]:
  Gap in sequence numbers — Wireshark missed a packet.
  Either packet was lost or not captured.

[TCP ACKed Unseen Segment]:
  ACK for data Wireshark didn't capture.
  Usually means capture started mid-conversation.

[TCP Keep-Alive]:
  Keepalive probe packet (1 byte, old sequence number).
```

### RTT analysis

```
Wireshark calculates RTT from matching data → ACK:

  Packet 10: seq 1000, data
  Packet 15: ack 2448        → RTT = time(15) - time(10)

View ACK RTT:
  tcp.analysis.ack_rtt > 0.1    (filter for slow ACKs > 100ms)

Statistics → TCP Stream Graphs → Round Trip Time Graph
  Shows RTT over time for a stream
  Spikes = congestion or processing delays
```

### TCP flow graphs

```
Statistics → Flow Graph
  Shows message sequence between endpoints
  Visual representation of SYN, data, ACK, FIN, RST exchanges
  Time flows top to bottom

Statistics → TCP Stream Graphs:
  - Stevens Graph (sequence number over time)
  - Throughput Graph
  - Round Trip Time Graph
  - Window Scaling Graph

These graphs reveal:
  - Slow start and rate increase
  - Loss events (flat spots in Stevens graph)
  - Window size limiting throughput
  - Congestion events (RTT spikes)
```

---

## TLS Decryption

### Setup

```
To decrypt TLS traffic in Wireshark:

Method 1: SSLKEYLOGFILE (recommended)
  1. Before generating traffic:
     export SSLKEYLOGFILE=/tmp/tls_keys.log
  2. Generate traffic:
     curl https://example.com
  3. Capture simultaneously:
     sudo tcpdump -i eth0 -w tls.pcap port 443
  4. In Wireshark:
     Edit → Preferences → Protocols → TLS
     (Pre)-Master-Secret log filename: /tmp/tls_keys.log
  5. Open tls.pcap → decrypted HTTP/2 visible!

Method 2: Server private key (TLS 1.2 only, RSA key exchange)
  Edit → Preferences → Protocols → TLS → RSA Keys List
  Does NOT work with TLS 1.3 or ECDHE (forward secrecy)

After decryption:
  - HTTP/2 frames visible and filterable
  - Can follow HTTP/2 streams
  - See actual request/response data
  - Filter by http2.header.name, http2.header.value
```

### What's visible without decryption

```
Even without decrypting, you can see:
  ✓ Handshake messages (ClientHello, ServerHello)
  ✓ SNI (hostname in ClientHello)
  ✓ Supported TLS versions and cipher suites
  ✓ Server certificate (including chain)
  ✓ Certificate validity dates
  ✓ Selected cipher suite
  ✓ TLS alerts (errors)
  ✗ HTTP request/response content (encrypted)
```

---

## DNS Analysis

```
Filter: dns

Wireshark shows:
  - Query name and type (A, AAAA, CNAME, MX, etc.)
  - Response code (No Error, NXDOMAIN, SERVFAIL)
  - Answer records
  - Authority and additional sections
  - DNSSEC signatures and validation
  - Response time (by pairing query and response)

Useful DNS filters:
  dns.qry.name == "example.com"
  dns.flags.rcode == 3              # NXDOMAIN
  dns.flags.rcode == 2              # SERVFAIL
  dns.resp.type == 5                # CNAME responses
  dns.time > 0.5                    # Slow DNS (> 500ms)

Statistics → DNS
  Shows query/response statistics, response times,
  most queried domains, error rates
```

---

## Statistics and Graphs

### Conversations

```
Statistics → Conversations
  Shows all communication pairs (L2, L3, L4)
  Sortable by: packets, bytes, duration
  Great for finding top talkers

  Columns: Address A, Address B, Packets A→B, Bytes A→B, 
           Packets B→A, Bytes B→A, Duration

  Sort by Bytes → find biggest transfers
  Sort by Duration → find longest connections
  Sort by Packets → find chattiest connections
```

### Endpoints

```
Statistics → Endpoints
  Shows all unique endpoints
  IPv4 tab: all unique IPs and their traffic
  TCP tab: all unique IP:port pairs

  Sort by Bytes → who's generating the most traffic?
  Sort by Packets → who's sending the most packets?
```

### IO Graphs

```
Statistics → I/O Graphs
  X-axis: time
  Y-axis: packets/sec, bytes/sec, bits/sec
  
  Add multiple graphs:
    Graph 1: All traffic (no filter)
    Graph 2: tcp.analysis.retransmission (retransmit rate)
    Graph 3: dns (DNS query rate)
    Graph 4: http.response.code >= 500 (server errors)

  Correlate events: retransmit spike coincides with DNS failure?
```

### Protocol hierarchy

```
Statistics → Protocol Hierarchy
  Tree view of all protocols in capture
  Shows percentage of traffic per protocol

  Example:
  Frame (100%)
  └ Ethernet (100%)
    └ IPv4 (99.5%)     IPv6 (0.5%)
      └ TCP (85%)      UDP (14.5%)
        └ TLS (70%)    HTTP (15%)    DNS (14%)
          └ HTTP/2 (70%)
```

---

## Expert Information

```
Analyze → Expert Information

Wireshark's built-in problem detection:

Severity levels:
  ERROR:    Malformed packets, protocol violations
  WARNING:  Retransmissions, out-of-order, zero window
  NOTE:     TCP keepalives, window updates, connection setup/teardown
  CHAT:     Informational (new connections, etc.)

Example expert messages:
  WARNING  TCP: Fast Retransmission (12 occurrences)
  WARNING  TCP: Out-of-Order (3 occurrences)
  WARNING  TCP: Zero Window (1 occurrence)
  NOTE     TCP: Duplicate ACK (45 occurrences)
  NOTE     TCP: Keep-Alive (100 occurrences)
  ERROR    DNS: Malformed packet (1 occurrence)

Filter by expert level:
  _ws.expert.severity == "Warning"
```

---

## Coloring Rules

```
View → Coloring Rules

Default coloring:
  Light purple:    TCP (general)
  Light green:     HTTP
  Light blue:      UDP/DNS
  Black on red:    RST, errors, retransmissions
  Black on yellow: TCP problems (window issues, etc.)
  Dark red:        TCP retransmission

Customize coloring for your analysis:
  Add rule: tcp.analysis.retransmission → bright red
  Add rule: http.response.code >= 500 → orange
  Add rule: dns.flags.rcode != 0 → dark red

Temporary coloring:
  Right-click packet → Colorize with filter
  Quick way to highlight related packets
```

---

## Practical Analysis Workflows

### Workflow: Debug slow web page

```
1. Capture: tcpdump -i eth0 -w slow_page.pcap port 443
2. Open in Wireshark
3. Statistics → TCP Stream Graphs → Round Trip Time
   → Look for RTT spikes
4. Filter: tcp.analysis.retransmission
   → Count retransmissions (causes delay)
5. Filter: dns.time > 0.1
   → Check for slow DNS lookups
6. Statistics → Conversations → TCP tab
   → Sort by Duration → find long-lived connections
7. Follow → TCP Stream for slowest connection
   → See where data flow stalls
```

### Workflow: Debug connection failures

```
1. Filter: tcp.flags.syn == 1 and tcp.flags.ack == 0
   → All connection attempts
2. Filter: tcp.flags.reset == 1
   → All resets (connection refused)
3. Check: RST in response to SYN = port closed / firewall
4. Check: SYN retransmissions = no response at all
5. Check expert info for context
```

### Workflow: Debug TLS issues

```
1. Filter: tls.handshake.type == 1
   → See all ClientHellos (check supported versions, cipher suites)
2. Filter: tls.handshake.type == 2
   → ServerHello (what was negotiated)
3. Filter: tls.alert_message
   → TLS alerts (handshake_failure, certificate_unknown, etc.)
4. Click on certificate message → examine certificate details
   → Check CN/SAN, expiry, issuer
5. If needed: enable SSLKEYLOGFILE decryption for full visibility
```

---

## Key Takeaways

1. **Capture with tcpdump, analyze with Wireshark** — tcpdump on servers (lightweight), Wireshark on desktop (powerful)
2. **Display filters > capture filters for analysis** — display filters understand protocols deeply
3. **Right-click → Follow → TCP Stream** is the fastest way to see a full conversation
4. **`tcp.analysis.*` filters** automatically detect retransmissions, duplicate ACKs, zero windows — Wireshark does the hard work
5. **Expert Information** summarizes all detected problems — start every analysis here
6. **Statistics → Conversations** reveals top talkers, longest connections, and biggest transfers
7. **IO Graphs** correlate events over time — overlay retransmits with traffic volume
8. **TLS decryption with SSLKEYLOGFILE** unlocks full HTTPS/HTTP2 visibility — essential for debugging encrypted traffic
9. **TCP Stream Graphs** (Stevens, RTT, Throughput) visualize congestion, loss, and flow control
10. **Protocol Hierarchy** gives instant overview of traffic composition in a capture

---

## Next

→ [Reading TCP Streams and Detecting Issues](./03-reading-tcp-streams-detecting-issues.md) — Systematic approach to finding problems in packet captures
