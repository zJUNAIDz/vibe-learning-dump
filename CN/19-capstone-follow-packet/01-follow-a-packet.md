# Capstone: Follow a Packet

> This is the synthesis of everything in this course. We're going to follow a single HTTP request — from the moment a user types a URL in their browser to the moment the response appears on screen. Every layer, every hop, every transformation. If you understand this journey, you understand computer networking.

---

## Table of Contents

1. [The Setup](#setup)
2. [Phase 1: The User Types the URL](#phase-1)
3. [Phase 2: DNS Resolution](#phase-2)
4. [Phase 3: TCP Three-Way Handshake](#phase-3)
5. [Phase 4: TLS Handshake](#phase-4)
6. [Phase 5: HTTP Request](#phase-5)
7. [Phase 6: Server-Side Processing](#phase-6)
8. [Phase 7: The Response Returns](#phase-7)
9. [Phase 8: Browser Renders](#phase-8)
10. [The Complete Timeline](#timeline)
11. [Every Transformation Summarized](#transformations)
12. [What Can Go Wrong at Each Phase](#failures)
13. [Key Takeaways](#key-takeaways)

---

## The Setup

```
The scenario:

  User in New York opens their laptop and types:
  https://api.example.com/users/42

  The server runs in AWS us-east-1 (Virginia).
  It's a Kubernetes cluster behind an ALB.

  Let's follow every single step.

The path:

  User's Laptop
    → Wi-Fi Access Point
      → Home Router (NAT)
        → ISP
          → Internet backbone
            → AWS network
              → ALB (Load Balancer)
                → Kubernetes Node
                  → Pod (nginx → app)
                    → Database query
                  → Response back
                → Through ALB
              → AWS network
            → Internet backbone
          → ISP
        → Home Router
      → Wi-Fi AP
    → User's Laptop
  → Browser renders the response
```

---

## Phase 1: The User Types the URL

### The browser parses the URL

```
https://api.example.com/users/42

  Scheme:    https (→ TLS required, port 443)
  Host:      api.example.com (→ need DNS lookup)
  Path:      /users/42 (→ part of HTTP request)
  Port:      443 (implicit from https)

Browser checks:
  1. HSTS list: Is example.com HSTS? → Yes, force HTTPS ✓
  2. Cache: Do I have a cached response? → No
  3. DNS cache: Do I know the IP? → No (first visit)
  4. Connection pool: Open connection to this host? → No

→ Must perform DNS lookup, then connect.
```

---

## Phase 2: DNS Resolution

### Step by step

```
Browser → OS resolver → DNS hierarchy

1. Browser DNS cache → MISS
2. OS DNS cache → MISS
3. OS reads /etc/resolv.conf → nameserver 192.168.1.1 (home router)

4. Laptop sends DNS query:
   ┌──────────────────────────────────────────────────┐
   │ DNS Query                                        │
   │                                                  │
   │ IP Header:                                       │
   │   Src: 192.168.1.100 (laptop LAN IP)             │
   │   Dst: 192.168.1.1 (home router)                 │
   │   Protocol: UDP                                  │
   │                                                  │
   │ UDP Header:                                      │
   │   Src Port: 54321 (random ephemeral)             │
   │   Dst Port: 53 (DNS)                             │
   │                                                  │
   │ DNS Payload:                                     │
   │   Query: A record for api.example.com            │
   │   Type: A (IPv4)                                 │
   │   Class: IN (Internet)                           │
   └──────────────────────────────────────────────────┘

5. This packet goes through the Wi-Fi stack:
   - Application creates UDP socket, sends data
   - Kernel builds UDP datagram
   - Kernel builds IP packet (src, dst, TTL=64)
   - Kernel does ARP lookup for gateway MAC (192.168.1.1)
   - Kernel builds Ethernet frame (laptop MAC → router MAC)
   - Frame handed to Wi-Fi driver
   - Wi-Fi: CSMA/CA, wait for clear channel, transmit

6. Home router receives, NATs the DNS query:
   - Changes Src IP: 192.168.1.100 → 73.45.22.100 (public IP)
   - Changes Src Port: 54321 → 61234 (NAT mapping)
   - Records NAT translation in conntrack table
   - Forwards to ISP's recursive resolver (e.g., 8.8.8.8)

7. Recursive resolver walks the DNS tree:
   . (root) → "Ask .com servers"
   .com TLD → "Ask ns1.example.com"
   ns1.example.com (authoritative) → "api.example.com = 52.20.30.40"
   
   (These steps may be cached at the resolver)

8. DNS response returns:
   52.20.30.40 for api.example.com, TTL=300

9. Response traverses back:
   Resolver → ISP → Home router (de-NAT) → Laptop
   
   Home router: Dst 73.45.22.100:61234 → 192.168.1.100:54321
   (conntrack entry maps it back)

10. Laptop receives DNS response:
    api.example.com → 52.20.30.40
    Cache for 300 seconds (TTL)

Time: ~20-100ms (depending on cache hits at resolver)
```

---

## Phase 3: TCP Three-Way Handshake

### SYN

```
Browser needs a TCP connection to 52.20.30.40:443.

Application: connect(fd, {52.20.30.40, 443})
Kernel: Creates TCP socket, starts handshake.

SYN Packet:
  ┌──────────────────────────────────────────────────┐
  │ Ethernet Header:                                 │
  │   Src MAC: aa:bb:cc:dd:ee:01 (laptop)            │
  │   Dst MAC: ff:ff:ff:ff:ff:ff → ARP → router MAC │
  │   EtherType: 0x0800 (IPv4)                       │
  │                                                  │
  │ IP Header:                                       │
  │   Src: 192.168.1.100                             │
  │   Dst: 52.20.30.40                               │
  │   TTL: 64                                        │
  │   Protocol: 6 (TCP)                              │
  │   Total Length: 60 bytes                          │
  │                                                  │
  │ TCP Header:                                      │
  │   Src Port: 49876 (ephemeral)                    │
  │   Dst Port: 443                                  │
  │   Seq: 1000 (ISN - random)                       │
  │   Flags: SYN                                     │
  │   Window: 65535                                  │
  │   Options:                                       │
  │     MSS: 1460                                    │
  │     Window Scale: 7 (multiply window by 128)     │
  │     SACK Permitted                               │
  │     Timestamps                                   │
  └──────────────────────────────────────────────────┘

The SYN travels:
  Laptop → Wi-Fi → Router → NAT → ISP → Internet → AWS → ALB
  
  At each hop:
    - Routers decrement TTL
    - Routers rewrite Src/Dst MAC (new Ethernet frame per hop)
    - NAT rewrites Src IP and Src Port at home router
    - IP header stays the same (except TTL) across the internet
    - TCP header untouched by intermediate routers
```

### SYN-ACK

```
AWS ALB receives SYN on port 443.
ALB sends SYN-ACK:

  ┌──────────────────────────────────────────────────┐
  │ TCP Header:                                      │
  │   Src Port: 443                                  │
  │   Dst Port: 49876 (NAT'd: really 61235 at NAT)  │
  │   Seq: 5000 (ALB's ISN)                          │
  │   Ack: 1001 (SYN seq + 1)                        │
  │   Flags: SYN, ACK                                │
  │   Window: 65535                                  │
  │   Options: MSS 1460, Window Scale 8, SACK        │
  └──────────────────────────────────────────────────┘

Returns through: AWS → Internet → ISP → Home Router (de-NAT) → Laptop
```

### ACK

```
Laptop sends ACK:
  Seq: 1001, Ack: 5001, Flags: ACK

TCP connection is now ESTABLISHED.

Time for handshake: ~15ms (same region) to ~80ms (cross-country)
  = 1 RTT (Round Trip Time)
```

---

## Phase 4: TLS Handshake

### TLS 1.3 (modern)

```
TLS 1.3 adds 1 RTT to connection setup:

Client                                          Server (ALB)
  │                                               │
  │──── ClientHello ─────────────────────────────>│
  │     Supported cipher suites                    │
  │     Key share (Curve25519 public key)          │
  │     SNI: api.example.com                       │
  │                                               │
  │<─── ServerHello + EncryptedExtensions ────────│
  │     Selected cipher: TLS_AES_256_GCM_SHA384    │
  │     Key share (server's public key)            │
  │     Certificate (api.example.com)              │
  │     Certificate Verify (signature)             │
  │     Finished (MAC)                             │
  │                                               │
  │──── Finished ────────────────────────────────>│
  │     (Client's MAC)                             │
  │                                               │
  │  ← Encrypted communication begins →           │

What happens during TLS:
  1. Client sends supported algorithms + key material
  2. Server selects algorithm, sends certificate
  3. Client verifies certificate:
     - Is it signed by a trusted CA?
     - Does CN/SAN match api.example.com?
     - Is it not expired?
     - OCSP: is it not revoked?
  4. Both derive shared secret using ECDHE
  5. Both derive session keys from shared secret
  6. All further traffic encrypted with AES-256-GCM

TLS 1.3: 1 RTT (ClientHello → ServerHello+data)
TLS 1.2: 2 RTTs (extra round trip for key exchange)

Total so far: DNS + TCP + TLS = ~200ms
```

---

## Phase 5: HTTP Request

### Building the request

```
Encrypted inside the TLS tunnel, the browser sends:

  GET /users/42 HTTP/2
  Host: api.example.com
  User-Agent: Mozilla/5.0 ...
  Accept: application/json
  Accept-Encoding: gzip, br
  Authorization: Bearer eyJhbG...
  Connection: keep-alive

HTTP/2 framing:
  This is not sent as plain text.
  HTTP/2 uses binary frames:
  
  ┌──────────────────────────┐
  │ HEADERS frame            │
  │   Stream ID: 1           │
  │   :method = GET          │
  │   :path = /users/42      │
  │   :authority = api.ex... │
  │   (HPACK compressed)     │
  └──────────────────────────┘
  
  HTTP/2 multiplexes: multiple requests on one TCP connection.
  This is why we didn't need a new TCP+TLS handshake for each request.

The HTTP request is:
  1. Serialized into HTTP/2 binary frame
  2. Handed to TLS layer → encrypted
  3. Handed to TCP layer → segmented (if needed)
  4. Handed to IP layer → IP header added
  5. Handed to Ethernet → framed with MACs
  6. Transmitted over Wi-Fi
```

### Packet on the wire

```
What the actual bytes look like (after encryption):

  ┌──────────┬──────────┬──────────┬──────────┬───────────────────┐
  │ Ethernet │ IP       │ TCP      │ TLS      │ Encrypted HTTP/2  │
  │ 14 bytes │ 20 bytes │ 32 bytes │ 5 bytes  │ ~200 bytes        │
  │ MACs +   │ Src/Dst  │ Ports +  │ Record   │ (opaque to anyone │
  │ EtherType│ IPs +TTL │ Seq/Ack  │ header   │  without the key) │
  └──────────┴──────────┴──────────┴──────────┴───────────────────┘
  
  Visible to intermediate routers:
    ✓ Ethernet MACs (changes per hop)
    ✓ IP addresses (src/dst)
    ✓ TCP ports (443)
  
  NOT visible (encrypted):
    ✗ The URL (/users/42)
    ✗ The HTTP headers
    ✗ The Authorization token
    ✗ The response data
  
  SNI (Server Name Indication) in TLS ClientHello:
    ✓ api.example.com IS visible during TLS handshake
    (this is why your ISP/firewall can see which domains you visit
     even with HTTPS — ECH/ESNI aims to fix this)
```

---

## Phase 6: Server-Side Processing

### Through the AWS infrastructure

```
Request arrives at AWS:

  Internet → AWS Internet Gateway → VPC
  
  1. ALB (Application Load Balancer)
     - Terminates TLS (decrypts)
     - Reads HTTP headers
     - Applies routing rules (host + path)
     - Selects target: Kubernetes NodePort on node-3
     - Opens NEW TCP connection to backend
     - Re-encrypts if backend requires TLS
     
  2. Kubernetes Node (node-3)
     - Packet arrives on NodePort (e.g., 30080)
     - kube-proxy iptables DNAT:
       Dst: node-3:30080 → 10.244.2.15:8080 (pod IP)
     - Packet enters VXLAN tunnel (if using overlay CNI)
       Encapsulated: node-3 → node-5 (where pod actually runs)
     
  3. Kubernetes Pod (on node-5)
     - VXLAN decapsulated
     - Arrives at pod's veth interface
     - Through cni0 bridge → pod's eth0
     - Pod's network namespace:
       nginx sidecar receives on :8080
       Proxies to app container on localhost:3000
     
  4. Application processes request
     - Reads /users/42
     - Queries database (another pod or RDS)
     - Database query: SELECT * FROM users WHERE id = 42
     - Database returns result
     - App serializes response as JSON
     - Returns HTTP response to nginx → ALB → client
```

### Database query sub-journey

```
The database query is ANOTHER full network journey:

  App pod (10.244.2.15) → Database pod (10.244.3.20:5432)
  
  1. App creates TCP connection to database service
     DNS: postgres.default.svc.cluster.local → 10.96.0.100 (ClusterIP)
  
  2. kube-proxy DNAT: 10.96.0.100:5432 → 10.244.3.20:5432 (pod IP)
  
  3. If on different node: VXLAN tunnel between nodes
  
  4. TCP+TLS handshake to database
  
  5. PostgreSQL wire protocol:
     Query → Parse → Bind → Execute → Response
  
  6. Result flows back through the same path
  
  Time: 1-5ms for simple queries within the cluster
```

---

## Phase 7: The Response Returns

### HTTP Response

```
The server sends back:

  HTTP/2 200 OK
  Content-Type: application/json
  Content-Length: 256
  Cache-Control: no-cache
  X-Request-Id: abc-123-def-456
  
  {
    "id": 42,
    "name": "Alice",
    "email": "alice@example.com",
    ...
  }
```

### The return journey

```
Response travels back through every layer:

  1. App container → nginx sidecar (localhost)
  2. Pod eth0 → cni0 bridge → veth → node network
  3. Node → VXLAN tunnel to node where ALB target is
  4. Node → ALB (NodePort reverse path)
  5. ALB:
     - Receives HTTP response from backend
     - Applies response headers (X-Forwarded-For, etc.)
     - TLS encrypts for the client
     - Sends on the client's original TCP connection
  6. ALB → AWS Internet Gateway → Internet
  7. Internet → ISP → Home Router
  8. Home Router: De-NAT (conntrack reverse mapping)
     Dst: 73.45.22.100:61235 → 192.168.1.100:49876
  9. Home Router → Wi-Fi AP → Laptop
  10. Laptop:
      - Wi-Fi driver receives frame
      - Kernel: Ethernet → IP → TCP → TLS → HTTP
      - TCP: ACK the received data
      - TLS: decrypt the payload
      - HTTP/2: demultiplex the stream
      - Browser: receives JSON response
```

### TCP during data transfer

```
The response may be larger than one MSS (1460 bytes):

  Server sends multiple segments:
  
  Seq 5001:6461  [1460 bytes]  ← first segment
  Seq 6461:7921  [1460 bytes]  ← second segment
  Seq 7921:8200  [279 bytes]   ← final segment (PSH flag set)
  
  Client ACKs (could be delayed ACK or immediate):
  ACK 8200       ← acknowledges all 3 segments
  
  TCP flow control ensures:
  - Sender doesn't overwhelm receiver (window)
  - Congestion control prevents network overload
  - Lost segments are retransmitted efficiently
```

---

## Phase 8: Browser Renders

```
Browser receives the HTTP response:

  1. Check status code: 200 OK ✓
  2. Parse Content-Type: application/json
  3. Decompress if needed (gzip/brotli)
  4. Parse JSON
  5. JavaScript callback receives data
  6. React/Vue/Angular updates the DOM
  7. User sees the result on screen

Connection stays open (HTTP/2 keep-alive):
  Next request uses the SAME TCP+TLS connection.
  No handshake overhead for subsequent requests.
  
  Connection can carry hundreds of requests over its lifetime.
  Browser typically closes after 30-60 seconds of idle.
```

---

## The Complete Timeline

```
Time (ms)   Event                              Layer
────────    ─────                              ─────
0           User presses Enter                 Application
1           Browser parses URL                 Application
2           DNS query sent                     L7/L4/L3
2-10        DNS: local cache check             L7
10-80       DNS: recursive resolution          L7
80          DNS response received              L7
81          TCP SYN sent                       L4
81-85       SYN through NAT, routing           L3/L2
85-115      SYN traverses Internet             L3
115         SYN arrives at ALB                 L4
116         SYN-ACK sent by ALB                L4
116-150     SYN-ACK returns to client          L3
150         ACK sent → TCP ESTABLISHED         L4
151         TLS ClientHello sent               L6 (TLS)
151-185     ClientHello traverses to ALB        L3
185         ALB processes TLS handshake        L6
186-220     ServerHello returns                L6/L3
220         TLS ESTABLISHED                    L6
221         HTTP/2 GET request sent            L7
221-255     Request traverses to ALB           L3
255         ALB receives, routes to K8s        L7
256-260     Request through K8s networking     L3/L2
260         App container receives request     L7
260-265     Database query                     L7/L4
265         App generates response             L7
266-300     Response traverses to ALB          L3
300-335     Response traverses Internet        L3
335-345     Response through NAT, Wi-Fi        L3/L2
345         Browser receives response          L7
346         JSON parsed, DOM updated           Application
350         User sees result                   Application

TOTAL: ~350ms for a full page load, first visit
  - DNS: ~80ms
  - TCP: ~70ms (1 RTT)
  - TLS: ~70ms (1 RTT for TLS 1.3)
  - HTTP request/response: ~120ms (1 RTT + server processing)
  - Rendering: ~5ms

Subsequent requests (same connection):
  - DNS: 0ms (cached)
  - TCP: 0ms (reused connection)
  - TLS: 0ms (reused session)
  - HTTP: ~120ms (1 RTT + processing)
  TOTAL: ~120ms ← much faster!
```

---

## Every Transformation Summarized

```
As the packet moves through the network, it is transformed:

Location              What changes
────────              ────────────
Application           URL → HTTP request → HTTP/2 binary frame
TLS Layer             Plaintext → AES-encrypted ciphertext
TCP Layer             Stream → Segments (with seq/ack numbers)
IP Layer              Segment → Packet (with src/dst IP)
Ethernet/WiFi         Packet → Frame (with src/dst MAC)

At each router hop:
  - TTL decremented (IP)
  - Source/Dest MAC rewritten (Ethernet)
  - IP addresses unchanged* (except NAT)

At NAT (home router):
  - Source IP: private → public
  - Source port: remapped

At ALB (load balancer):
  - TLS terminated (decrypted)
  - New TCP connection to backend
  - HTTP forwarded (with added headers)

At kube-proxy (DNAT):
  - Destination IP: Service ClusterIP → Pod IP
  - Destination port: may change

At VXLAN (overlay):
  - Original frame encapsulated in outer UDP/IP
  - New outer Ethernet/IP/UDP headers added
  - At destination: outer headers stripped

The HTTP payload (GET /users/42) is untouched through ALL of this.
Only source/destination addresses change.
Only encryption/decryption transforms the data.
```

---

## What Can Go Wrong at Each Phase

```
Phase           What can fail                 How to detect
─────           ──────────────                ──────────────
DNS             NXDOMAIN, timeout, wrong IP   dig, nslookup
TCP handshake   Timeout (firewall), RST        tcpdump for SYN without SYN-ACK
TLS             Cert expired, name mismatch   openssl s_client
                Protocol mismatch             curl -v (shows TLS errors)
HTTP            4xx/5xx, timeout, wrong path  curl -v, check status code
Server          Crash, OOM, slow query        kubectl logs, server metrics
Database        Connection refused, slow      DB logs, query profiler
k8s Service     No endpoints, wrong port      kubectl get endpoints
k8s CNI         Pod unreachable across nodes  ping pod IP from other node
NAT             Conntrack full, timeout       dmesg, nf_conntrack_count
Wi-Fi           Signal weak, interference     iwconfig, channel analysis
MTU             Large packets dropped          ping -s 1472 -M do
Load Balancer   Health check fail, draining   ALB logs, target health
```

---

## Key Takeaways

1. **A single HTTPS request touches 7+ layers of the networking stack** — from Wi-Fi radio transmission to HTTP/2 binary framing, each layer adds headers and transformations
2. **DNS adds 20-100ms to the first request** — after caching, it's free. This is why DNS performance and caching matter enormously
3. **TCP + TLS handshake costs 2-3 RTTs for the first connection** — HTTP/2 + TLS 1.3 reduces this to 2 RTTs. Connection reuse eliminates it for subsequent requests
4. **The HTTP payload is untouched across the entire journey** — it's encrypted by TLS, but routers only change Ethernet MACs and IP TTLs (plus NAT)
5. **NAT creates state that must persist for the life of the connection** — home router's conntrack table maps private:port to public:port. If this entry expires, the connection breaks
6. **The ALB is a full proxy — it terminates and re-creates connections** — the client's TCP connection ends at the ALB. A completely new connection goes to the backend
7. **Kubernetes adds 3 extra hops: kube-proxy DNAT, CNI bridge, and possibly VXLAN overlay** — each is a potential failure point, each can be debugged with standard Linux tools
8. **Subsequent requests on the same connection skip DNS, TCP, and TLS** — going from 350ms (first request) to 120ms (subsequent). Connection reuse is the biggest performance optimization
9. **Encrypted payload is opaque to everyone except endpoints** — routers, ISPs, and firewalls can see IP addresses and port 443, but not the URL, headers, or data. SNI leaks the domain name
10. **Understanding this full journey is the difference between guessing and diagnosing** — when something breaks, you know exactly which phase to investigate based on the symptoms

---

## Next Module

→ [Module 20: Cloud Networking](../20-cloud-networking/01-sdn-and-vpc-concepts.md) — SDN, VPCs, and how cloud providers build their networks
