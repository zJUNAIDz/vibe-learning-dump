# The TCP/IP Model — The Model the Internet Actually Uses

> The OSI model is for teaching. The TCP/IP model is what runs the internet. Let's understand why it won and how it's structured.

---

## Table of Contents

1. [Origin Story: How TCP/IP Beat OSI](#origin-story-how-tcpip-beat-osi)
2. [The Four (or Five) Layers](#the-four-or-five-layers)
3. [Layer-by-Layer Deep Dive](#layer-by-layer-deep-dive)
4. [TCP/IP vs OSI — The Real Differences](#tcpip-vs-osi--the-real-differences)
5. [The Hourglass Architecture](#the-hourglass-architecture)
6. [Seeing TCP/IP on Your Linux Machine](#seeing-tcpip-on-your-linux-machine)

---

## Origin Story: How TCP/IP Beat OSI

### The ARPANET beginnings

In 1969, the US Defense Advanced Research Projects Agency (DARPA) built ARPANET — the first wide-area packet-switched network. It connected four universities.

The problem: ARPANET used a monolithic protocol called NCP (Network Control Protocol). NCP was designed for one network. As more networks appeared (satellite networks, radio networks, Ethernet LANs), they needed a way to interconnect different types of networks.

### Vint Cerf and Bob Kahn's insight (1974)

Cerf and Kahn published "A Protocol for Packet Network Intercommunication" — the foundational paper for TCP/IP. Their key insight:

> **Don't try to build one universal network. Instead, build many different networks and create a protocol that lets them talk to each other.**

This is the "inter" in "internet" — it's a network OF networks, connected by a common protocol (IP).

### Why TCP/IP won over OSI

The OSI model and its associated protocols were developed by an international standards body (ISO) through committee processes. TCP/IP was developed by engineers who were building and using the ARPANET.

Key differences in how they were developed:

| Aspect | OSI | TCP/IP |
|--------|-----|--------|
| Development | Committee-driven, top-down | Engineer-driven, bottom-up |
| First implementation | After the spec was finalized | While the spec was being written |
| Design goal | Complete, elegant architecture | Working code, then standardize |
| Adoption | Governments mandated it | Universities and researchers chose it |
| Cost | Commercial implementations were expensive | Free implementations in BSD Unix |

The critical factor was **BSD Unix**. UC Berkeley's implementation of TCP/IP was included in the BSD Unix distribution starting in 1983. Since most universities ran BSD, they immediately had TCP/IP networking. By the time OSI protocols were ready, TCP/IP had already achieved critical mass.

The IETF (Internet Engineering Task Force) motto captures the TCP/IP philosophy:

> "We reject kings, presidents, and voting. We believe in rough consensus and running code."

### The lesson

The technically "better" standard (OSI was more comprehensive) lost to the practical one (TCP/IP worked and was free). This happens often in technology. Keep this in mind: **deployed and working beats elegant and theoretical**.

---

## The Four (or Five) Layers

The TCP/IP model is sometimes described as four layers, sometimes as five. Here's the confusion and the resolution:

### The original 4-layer model (RFC 1122)

```
┌─────────────────────┐
│   Application       │  HTTP, DNS, SSH, SMTP
├─────────────────────┤
│   Transport          │  TCP, UDP
├─────────────────────┤
│   Internet           │  IP, ICMP, ARP (debatable)
├─────────────────────┤
│   Link               │  Ethernet, Wi-Fi, PPP
└─────────────────────┘
```

### The practical 5-layer model (commonly used today)

```
┌─────────────────────┐
│   Application       │  HTTP, DNS, SSH, SMTP
├─────────────────────┤
│   Transport          │  TCP, UDP
├─────────────────────┤
│   Network            │  IP, ICMP
├─────────────────────┤
│   Data Link          │  Ethernet framing, MAC, ARP
├─────────────────────┤
│   Physical           │  Electrical signals, fiber optics
└─────────────────────┘
```

The 5-layer model splits the "Link" layer into "Data Link" and "Physical" because they handle genuinely different concerns:
- **Physical**: Sending raw bits over a medium (voltages, light)
- **Data Link**: Organizing bits into frames, addressing local devices (MAC)

In this curriculum, we use the **5-layer model** because the distinction between physical and data link is useful for debugging.

---

## Layer-by-Layer Deep Dive

### Link Layer (Physical + Data Link)

**Job**: Get a frame from one device to another device on the **same local network segment**.

This is the "one hop" layer. When your computer sends a packet to your home router, the link layer handles that single hop. When the router forwards to the next router, a different link layer handles that hop.

**Key concept**: Link-layer addresses (MAC addresses) are used for **local** delivery only. They change at each hop because they only need to identify the next device in the chain.

```
Your PC ──[Ethernet]──> Your Router ──[Fiber]──> ISP Router ──[Fiber]──> ...
          MAC A→B                     MAC B→C                MAC C→D
          IP: src→dst                 IP: src→dst            IP: src→dst
```

Notice: IP addresses stay the same end-to-end. MAC addresses change at every hop. This is fundamental.

**Protocols at this layer**:
- Ethernet (IEEE 802.3) — wired LANs
- Wi-Fi (IEEE 802.11) — wireless LANs
- PPP (Point-to-Point Protocol) — serial links

```bash
# See your link-layer interfaces
ip -d link show

# See the link-layer (MAC) addresses of local devices
ip neigh show

# Capture at the link layer
sudo tcpdump -i eth0 -e -c 10
# The -e flag shows Ethernet headers (MAC addresses)
```

---

### Network Layer (Internet Layer)

**Job**: Get a packet from the source machine to the destination machine, **across multiple networks**.

This is the "many hops" layer. It handles:
- **Addressing**: IP addresses (globally meaningful, hierarchical)
- **Routing**: Determining which next-hop router to send the packet to
- **Fragmentation**: Breaking packets into smaller pieces if needed (though this is increasingly avoided)

**The core protocol**: IP (Internet Protocol), currently version 4 (IPv4) and version 6 (IPv6).

**Key design decision**: IP is **unreliable and connectionless**.
- **Unreliable**: IP does not guarantee delivery. A packet might be dropped and IP won't retry.
- **Connectionless**: Each packet is independent. IP doesn't remember previous packets.

This is deliberate. It keeps routers simple and fast — they just forward packets based on destination address. Reliability is handled by higher layers (TCP).

**Supporting protocols at this layer**:
- **ICMP** (Internet Control Message Protocol): Error reporting and diagnostics. `ping` uses ICMP.
- **ARP** (Address Resolution Protocol): Translates IP addresses to MAC addresses on local networks. (Some place ARP between layers 2 and 3 — it bridges both.)

```bash
# See your IP addresses
ip addr show

# See the routing table
ip route show

# Trace the network-layer path to a destination
traceroute -n 8.8.8.8

# See ICMP in action
ping -c 3 8.8.8.8

# Watch IP-level decisions
ip route get 10.0.0.1
```

---

### Transport Layer

**Job**: Deliver data to the correct **application** on the destination machine, and (optionally) provide **reliability**.

The network layer gets packets to the right machine. The transport layer gets data to the right process on that machine.

**Two main protocols**:

#### TCP (Transmission Control Protocol)
- **Connection-oriented**: Requires a handshake before data flows
- **Reliable**: Guarantees delivery (retransmits lost packets)
- **Ordered**: Data arrives in the order sent
- **Flow-controlled**: Prevents sender from overwhelming receiver
- **Congestion-controlled**: Prevents sender from overwhelming the network

Used for: HTTP, SSH, email, database connections — anything where losing data is unacceptable.

#### UDP (User Datagram Protocol)
- **Connectionless**: No handshake, just send
- **Unreliable**: No delivery guarantees
- **Unordered**: Packets may arrive in any order
- **No flow/congestion control**: Application must manage this

Used for: DNS queries, video streaming, gaming, VoIP — anything where speed matters more than perfect delivery.

**How multiplexing works**:

A machine has one IP address but runs many applications. Port numbers (16-bit, range 0–65535) differentiate them:

```
Source IP: 192.168.1.100, Source Port: 54321 → identifies the sender's application
Dest IP:   93.184.216.34, Dest Port:   80    → identifies the receiver's application
```

The combination (src_ip, src_port, dst_ip, dst_port, protocol) uniquely identifies a connection. This means:
- The same machine can have thousands of simultaneous TCP connections
- Each connection has a unique 5-tuple
- The OS uses this tuple to route incoming data to the right process

```bash
# See all listening ports (which applications are waiting for connections)
ss -tuln

# See established TCP connections
ss -tn state established

# See which process owns which connection
sudo ss -tpn

# See detailed TCP state (RTT, congestion window, retransmits)
ss -ti
```

---

### Application Layer

**Job**: Define the meaning and format of data exchanged between applications.

Everything above the transport layer is "application" in the TCP/IP model. This includes things that OSI splits into Session, Presentation, and Application layers.

Application-layer protocols define:
- **What operations are available** (HTTP: GET, POST, PUT, DELETE)
- **Data formats** (JSON, XML, Protocol Buffers)
- **Error codes** (HTTP: 200 OK, 404 Not Found, 500 Internal Server Error)
- **Session management** (cookies, tokens)
- **Encryption** (TLS — though some argue this is between transport and application)

**Important**: Application-layer protocols run on top of TCP or UDP:

```
HTTP/1.1, HTTP/2 → TCP (port 80/443)
HTTP/3           → QUIC → UDP (port 443)
DNS              → UDP (port 53, sometimes TCP)
SSH              → TCP (port 22)
SMTP             → TCP (port 25/587)
```

```bash
# See HTTP at the application layer
curl -v http://example.com

# See DNS at the application layer
dig +all example.com

# Initiate SSH
ssh -v user@host
# The -v flag shows protocol negotiation details
```

---

## TCP/IP vs OSI — The Real Differences

This comparison is not about "which is right" — it's about understanding two different perspectives on the same reality.

### Structural differences

```
OSI (7 layers)              TCP/IP (5 layers)
────────────────            ─────────────────
7. Application              5. Application
6. Presentation                (merged)
5. Session                     (merged)
4. Transport                4. Transport
3. Network                  3. Network
2. Data Link                2. Data Link
1. Physical                 1. Physical
```

### Philosophical differences

| Aspect | OSI | TCP/IP |
|--------|-----|--------|
| **Design approach** | Top-down: define layers, then build protocols | Bottom-up: build working protocols, then describe layers |
| **Layer boundaries** | Rigid: each layer has a strict, well-defined interface | Flexible: protocols can span or bypass layers |
| **Session/Presentation** | Separate layers with distinct functions | Absorbed into application—developers handle encoding and sessions |
| **Universality** | Designed to be THE universal networking model | Designed to work on the actual internet |
| **Where it's used** | Teaching, certification exams, vendor marketing | Actual internet protocol design and implementation |

### Why it matters

When someone says "Layer 7 load balancer," they mean the OSI numbering. When you look at the Linux kernel networking stack, it follows the TCP/IP model. Both numbering systems are used in industry, often interchangeably. Understanding where they differ prevents confusion.

### The practical overlap

For 90% of daily work, the models agree:
- Application (OSI 7 = TCP/IP 5)
- Transport (OSI 4 = TCP/IP 4)
- Network (OSI 3 = TCP/IP 3)
- Data Link (OSI 2 = TCP/IP 2)
- Physical (OSI 1 = TCP/IP 1)

The disagreement is only about layers 5 and 6, which barely exist as separate entities in practice.

---

## The Hourglass Architecture

The internet's protocol architecture has a very specific shape — an hourglass:

```
            ┌─────────────────────────────┐
            │   HTTP  DNS  SSH  SMTP  ... │  Many application protocols
            ├─────────────────────────────┤
            │       TCP        UDP        │  Few transport protocols
            ├─────────────────────────────┤
            │            IP               │  ONE network protocol ← the "waist"
            ├─────────────────────────────┤
            │   Ethernet  Wi-Fi  PPP ...  │  Many link protocols
            └─────────────────────────────┘
```

### Why IP is "the waist of the internet"

The entire internet works because everything converges on one protocol: **IP**.

- Any application protocol can run on TCP or UDP
- TCP and UDP run on IP
- IP runs on any link-layer protocol

This means:
- A new application protocol (e.g., QUIC) only needs to work with UDP/IP — it automatically works over Ethernet, Wi-Fi, fiber, etc.
- A new link technology (e.g., 5G cellular) only needs to carry IP packets — it automatically supports every application protocol

If we had TWO competing network-layer protocols, every application would need to support both, and every link technology would need to support both. The combinatorial explosion would be unmanageable.

### The cost of the hourglass

The hourglass makes IP incredibly hard to change. IPv6 was designed in 1998 and is STILL not fully deployed (as of 2025+). Why? Because changing the waist of the hourglass requires changing EVERYTHING above and below it simultaneously.

This is called **ossification** — the most critical protocol in the stack is the hardest to evolve. Understanding this helps you understand why:
- HTTP/3 uses QUIC (which runs on UDP) instead of changing TCP
- Many innovations happen at the application layer (easier to change)
- IPv6 adoption has taken decades

---

## Seeing TCP/IP on Your Linux Machine

### The kernel's view of the protocol stack

The Linux kernel implements the TCP/IP stack. You can see its view:

```bash
# See network protocols the kernel understands
cat /proc/net/protocols

# See all TCP sockets
cat /proc/net/tcp

# See all UDP sockets
cat /proc/net/udp

# See IP routing
cat /proc/net/route

# See ARP table
cat /proc/net/arp

# See network statistics per protocol
cat /proc/net/snmp
```

### Watching a packet traverse the layers

Let's see the layers in action with a simple HTTP request:

```bash
# Terminal 1: Start a detailed packet capture
sudo tcpdump -i any -nn -vv -X port 80

# Terminal 2: Make a request
curl http://example.com
```

In the tcpdump output, you'll see the layered structure:

**Packet 1: TCP SYN (Layer 4)**
```
IP (tos 0x0, ttl 64, id 12345, offset 0, flags [DF], proto TCP (6), length 60)
    192.168.1.100.54321 > 93.184.216.34.80: Flags [S], seq 987654321, win 64240, ...
```

Reading this:
- `IP` → Layer 3 header
- `proto TCP (6)` → The Network layer says "the payload is TCP"
- `192.168.1.100.54321 > 93.184.216.34.80` → Layer 3 (IPs) and Layer 4 (ports)
- `Flags [S]` → TCP SYN flag (Layer 4 control)
- `win 64240` → TCP window size (Layer 4 flow control)

**Packet 4-5: HTTP Request (Layer 5/Application)**
After the TCP handshake completes, you'll see:
```
GET / HTTP/1.1
Host: example.com
User-Agent: curl/7.68.0
Accept: */*
```

This is pure Layer 7 — the application data carried inside TCP segments, inside IP packets, inside Ethernet frames.

### The encapsulation reality

Each layer adds its own header to the data:

```
Total packet size breakdown for an HTTP request:

Layer 2 (Ethernet):  14 bytes (6 dst MAC + 6 src MAC + 2 type)
Layer 3 (IP):        20 bytes (minimum IP header)
Layer 4 (TCP):       20 bytes (minimum TCP header, can be 32+ with options)
Layer 7 (HTTP):      variable (the actual request/response)

═══════════════════════════════════════════════════════
Minimum overhead per packet: 54 bytes (before any application data)
```

For a 1-byte application payload, you're sending 55 bytes total. The overhead is 98%. This is why:
- Protocols try to batch data (TCP's Nagle algorithm)
- Jumbo frames exist (increase the allowed payload per frame)
- Protocol efficiency matters for small messages

---

## Key Takeaways

1. **TCP/IP won because of pragmatism**: Working code, free implementation in BSD, engineers building what they needed
2. **The internet is a 5-layer stack**: Physical, Data Link, Network (IP), Transport (TCP/UDP), Application
3. **IP is the waist of the hourglass**: Everything converges on it, which makes the internet work but also makes IP hard to change
4. **MAC addresses are local, IP addresses are global**: MAC changes at each hop, IP stays the same
5. **Ports identify applications**: The transport layer uses port numbers to multiplex connections
6. **Each layer adds overhead**: Headers from every layer consume bandwidth, especially noticeable with small payloads
7. **The Linux kernel implements the TCP/IP stack**: You can see every layer's state through `/proc`, `ip`, `ss`, and `tcpdump`

---

## Next Up

→ [03-encapsulation-and-reality.md](03-encapsulation-and-reality.md) — How data actually flows through layers, and where layering breaks down
