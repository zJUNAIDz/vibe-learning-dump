# DNS Deep Dive — Protocol, DNSSEC, and Advanced Topics

> DNS is deceptively simple on the surface. Under the hood, there are zone transfers, DNSSEC signature chains, negative caching, DNS amplification attacks, and split-horizon configurations. This file goes deep.

---

## Table of Contents

1. [DNS Message Format](#message-format)
2. [Zones and Zone Files](#zones)
3. [Zone Transfers (AXFR/IXFR)](#zone-transfers)
4. [Negative Caching (NXDOMAIN)](#negative-caching)
5. [EDNS0: Extension Mechanisms](#edns0)
6. [DNSSEC: Signing the Phone Book](#dnssec)
7. [DNS over TLS and HTTPS](#dns-encryption)
8. [DNS-Based Attacks](#attacks)
9. [Split-Horizon DNS](#split-horizon)
10. [DNS Load Balancing](#load-balancing)
11. [Linux: Running Your Own Resolver](#own-resolver)
12. [Key Takeaways](#key-takeaways)

---

## Message Format

Every DNS query and response follows the same binary format (RFC 1035):

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                      ID                         |  16 bits: transaction ID
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|QR| Opcode  |AA|TC|RD|RA| Z  |AD|CD|   RCODE    |  16 bits: flags
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    QDCOUNT                      |  # of questions
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    ANCOUNT                      |  # of answers
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    NSCOUNT                      |  # of authority records
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    ARCOUNT                      |  # of additional records
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                  Questions                      |  Variable
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                  Answers                        |  Variable
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                  Authority                      |  Variable
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                  Additional                     |  Variable
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

### Key flags

| Flag | Meaning |
|------|---------|
| QR | 0 = query, 1 = response |
| AA | Authoritative Answer (server is authoritative for this zone) |
| TC | Truncated (response too large for UDP, retry with TCP) |
| RD | Recursion Desired (client wants recursive resolution) |
| RA | Recursion Available (server supports recursion) |
| AD | Authenticated Data (DNSSEC validated) |
| CD | Checking Disabled (don't validate DNSSEC) |
| RCODE | 0=NOERROR, 3=NXDOMAIN, 2=SERVFAIL, 5=REFUSED |

```bash
# See the full DNS message with dig
dig +noall +comments google.com
# ;; flags: qr rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1
#             ^  ^  ^
#             |  |  └─ Recursion Available
#             |  └──── Recursion Desired
#             └─────── Query Response
```

---

## Zones

A **zone** is a portion of the DNS namespace managed by a specific organization. A zone file contains all DNS records for that zone.

### Zone file format

```
; Zone file for example.com
$ORIGIN example.com.
$TTL 3600       ; Default TTL: 1 hour

; SOA Record (Start of Authority)
@   IN  SOA  ns1.example.com. admin.example.com. (
                2024010101 ; Serial (YYYYMMDDNN)
                3600       ; Refresh (1 hour)
                900        ; Retry (15 minutes)
                604800     ; Expire (1 week)
                86400      ; Negative cache TTL (1 day)
            )

; NS Records (Nameservers)
@       IN  NS   ns1.example.com.
@       IN  NS   ns2.example.com.

; A Records
@       IN  A    93.184.216.34
www     IN  A    93.184.216.34
mail    IN  A    93.184.216.35
api     IN  A    93.184.216.36

; AAAA Records
@       IN  AAAA 2606:2800:220:1:248:1893:25c8:1946

; CNAME Records
blog    IN  CNAME  www.example.com.
docs    IN  CNAME  readthedocs.io.

; MX Records (mail routing)
@       IN  MX   10  mail.example.com.
@       IN  MX   20  backup-mail.example.com.

; TXT Records
@       IN  TXT  "v=spf1 include:_spf.google.com ~all"

; Glue Records (NS hosts within this zone need A records)
ns1     IN  A    198.51.100.1
ns2     IN  A    198.51.100.2
```

### SOA record details

The SOA (Start of Authority) controls zone behavior:

| Field | Purpose |
|-------|---------|
| Serial | Version number. MUST increase on every change. Convention: YYYYMMDDNN |
| Refresh | How often secondaries check for updates |
| Retry | How soon to retry if refresh fails |
| Expire | When secondaries stop serving data if they can't refresh |
| Negative TTL | How long to cache NXDOMAIN (name does not exist) responses |

---

## Zone Transfers

### AXFR (Full zone transfer)

When a secondary DNS server needs all records from the primary:

```
Secondary → Primary: "Send me the entire zone for example.com"
Primary → Secondary: [streams all records over TCP]
```

```bash
# Attempt a zone transfer (many servers restrict this)
dig @ns1.example.com example.com AXFR

# zone transfer uses TCP because the response is typically large
```

### IXFR (Incremental zone transfer)

Only transfer records that changed since a given serial number:

```
Secondary: "I have serial 2024010101. What changed?"
Primary: "Records added: [...], Records removed: [...]"
```

### Security concern

Zone transfers expose ALL records in a domain. This is an information leak:

```bash
# If a server allows AXFR to anyone:
dig @ns1.target.com target.com AXFR
# Could reveal: internal hostnames, IP ranges, server roles

# Prevention: restrict AXFR to specific IPs (secondary servers only)
# In BIND: allow-transfer { 198.51.100.2; };
```

---

## Negative Caching

When you query a name that doesn't exist, you get an NXDOMAIN response. This answer is ALSO cached.

```bash
dig nonexistent.example.com
# status: NXDOMAIN
# SOA in authority section shows the negative cache TTL

# Why cache negatives?
# Without negative caching:
#   Typo in config → DNS query every time → floods authoritative servers
#   E.g., app configured with "databaes.internal" (typo) → queries every second forever

# With negative caching:
#   First query → NXDOMAIN, cached for SOA minimum TTL
#   Subsequent queries → answered from cache → no flood
```

### NODATA vs NXDOMAIN

- **NXDOMAIN**: The name doesn't exist at all
- **NODATA** (RCODE=NOERROR, empty answer): The name exists but doesn't have the requested record type

```bash
dig example.com AAAA
# If example.com has A records but no AAAA:
# status: NOERROR
# ANSWER SECTION: (empty)
# This is NODATA — the name exists, but no AAAA record
```

---

## EDNS0

The original DNS message was limited to 512 bytes (UDP). EDNS0 (Extension Mechanisms for DNS, RFC 6891) extends this.

### What EDNS0 provides

```
Original DNS:  512 byte UDP limit, no extensions
EDNS0:         Up to 4096 bytes UDP, extension flags

EDNS0 adds an OPT pseudo-record in the Additional section:
  - UDP payload size (typically 1232 or 4096)
  - DNSSEC OK (DO) flag
  - Extended RCODE bits
  - Additional options (client subnet, cookie, etc.)
```

### Why 1232 bytes?

The recommended EDNS0 buffer size is 1232 bytes (RFC 8020 recommendation). This is the largest size that avoids IP fragmentation on virtually all network paths (1280 bytes IPv6 minimum MTU - 48 bytes headers).

```bash
# dig uses EDNS0 by default
dig google.com
# Look for: ;; OPT PSEUDOSECTION: EDNS: version: 0, flags: do; udp: 512

# Disable EDNS0
dig +noedns google.com

# Set EDNS0 buffer size
dig +bufsize=1232 google.com
```

---

## DNSSEC

DNS was designed without security. Any response could be spoofed. DNSSEC adds cryptographic signatures to DNS responses.

### The problem DNSSEC solves

```
Without DNSSEC:
  You ask: "What's the IP of mybank.com?"
  Attacker intercepts and responds: "Evil IP"
  Your browser connects to attacker's server
  You enter your password → stolen

With DNSSEC:
  You ask: "What's the IP of mybank.com?"
  Response includes cryptographic signature
  Your resolver validates the signature using the parent zone's key
  Invalid signature → response rejected → you're safe
```

### How DNSSEC works

Each zone has a key pair:
- **ZSK (Zone Signing Key)**: Signs individual records
- **KSK (Key Signing Key)**: Signs the ZSK (allows key rotation without updating parent)

```
Trust chain:
  Root zone (IANA holds the root KSK)
    → Signs .com zone's DS record (hash of .com's KSK)
      → .com signs example.com's DS record
        → example.com signs its own records with ZSK

Validation:
  Resolver has root trust anchor (built-in)
  → Validates .com's key using root's signature
    → Validates example.com's key using .com's signature
      → Validates www.example.com's A record using example.com's key
```

### DNSSEC record types

| Record | Purpose |
|--------|---------|
| **RRSIG** | Signature over a record set |
| **DNSKEY** | Zone's public key (ZSK and KSK) |
| **DS** | Delegation Signer — hash of child's KSK, stored in parent zone |
| **NSEC/NSEC3** | Proves a name does NOT exist (authenticated denial) |

```bash
# Query with DNSSEC validation
dig +dnssec google.com

# Check DNSSEC chain
dig +sigchase +trusted-key=/etc/trusted-key.key google.com

# Verify a zone is signed
dig google.com DNSKEY +short
# Returns public keys if zone is signed

# Check DS record in parent
dig google.com DS +short
```

### DNSSEC limitations

1. **Doesn't encrypt**: DNSSEC proves authenticity, not confidentiality. Everyone can still see your queries.
2. **Complex key management**: Key rotation, algorithm changes, DS record updates at registrar
3. **NSEC zone walking**: NSEC records can be used to enumerate all names in a zone (NSEC3 partially mitigates this)
4. **Incomplete deployment**: Many domains are not signed

---

## DNS Encryption

### DNS-over-TLS (DoT)

Encrypts DNS queries using TLS on port 853:

```
Traditional:  Client →→→ [plaintext query] →→→ Resolver
DoT:          Client →→→ [TLS encrypted query] →→→ Resolver

Benefits: ISP/attacker can't see your queries
Drawback: Traffic on port 853 is obviously DNS (can be blocked)
```

### DNS-over-HTTPS (DoH)

Encrypts DNS queries inside HTTPS on port 443:

```
DoH:  Client →→→ [HTTPS POST to resolver] →→→ Resolver

URL: https://dns.google/dns-query
     https://cloudflare-dns.com/dns-query

Benefits: Indistinguishable from normal HTTPS traffic (harder to block)
Drawback: Bypasses enterprise DNS policies, complicates network debugging
```

### Configuring encrypted DNS on Linux

```bash
# systemd-resolved with DoT
sudo vim /etc/systemd/resolved.conf
# [Resolve]
# DNS=1.1.1.1#cloudflare-dns.com
# DNSOverTLS=yes

sudo systemctl restart systemd-resolved

# Verify
resolvectl status
# Should show: DNSOverTLS: yes
```

---

## Attacks

### DNS cache poisoning

Inject false records into a resolver's cache:

```
Scenario:
  Attacker sends forged response for mybank.com → evil_ip
  If it arrives before the real response, resolver caches the fake answer
  Users get evil_ip when they look up mybank.com for the next TTL period

Defense:
  - DNSSEC (validates signatures)
  - Source port randomization (RFC 5452)
  - Transaction ID randomization
  - 0x20 encoding (randomize case in query: gOoGLe.CoM)
```

### DNS amplification attack (DDoS)

DNS responses can be much larger than queries. Attackers exploit this:

```
1. Attacker sends DNS query with SPOOFED source IP = victim's IP
2. Query: small (60 bytes), type=ANY
3. Response: large (3000 bytes), sent to victim (spoofed source)
4. Amplification factor: 50x

Mitigation:
  - Rate limiting DNS responses
  - Blocking ANY queries
  - BCP38 (ingress filtering to prevent IP spoofing)
  - Response Rate Limiting (RRL) on authoritative servers
```

### DNS tunneling

Encoding data inside DNS queries to bypass firewalls:

```
Client sends: aGVsbG8gd29ybGQ.tunnel.attacker.com
              ^^^^^^^^^^^^^^^^ base64-encoded data in subdomain

Attacker's DNS server: decodes subdomain, processes data,
                       encodes response in TXT record

Works because: DNS is almost never blocked (even strict firewalls allow port 53)
Detection:
  - Anomalously long subdomain labels
  - High query volume to unusual domains
  - TXT record queries with encoded-looking data
```

### DNS rebinding

Trick a browser into connecting to internal network resources:

```
1. Attacker's domain resolves to attacker's server (first query)
2. Attacker's page loads JavaScript
3. DNS TTL expires, second query returns 192.168.1.1 (internal IP)
4. JavaScript makes request to same domain → goes to internal IP
5. Browser thinks it's same-origin (same domain) → allows request

Defense: DNS pinning (browser caches DNS independently), private IP filtering
```

---

## Split-Horizon DNS

The same domain name resolves to different IPs depending on who's asking.

```
External query:   api.company.com → 203.0.113.50 (public load balancer)
Internal query:   api.company.com → 10.0.5.100   (private backend server)
```

### Why

- Internal clients should connect directly to internal servers (faster, no hairpin through public IP)
- Different records for VPN users vs. external users
- Cloud services that have different internal/external endpoints

```bash
# Test by querying different resolvers
dig @8.8.8.8 api.company.com       # external view
dig @10.0.0.53 api.company.com     # internal view
# Different answers = split-horizon in effect
```

---

## Load Balancing

### Round-robin DNS

Return multiple A records. Clients pick the first one (or rotate):

```bash
dig google.com A +short
# 142.250.80.46
# 142.250.80.78
# 142.250.80.110

# Each client gets a different order → primitive load balancing
```

**Limitations**: No health checking (if one IP is down, clients still get it). Uneven distribution. Client-side caching means imbalance.

### Geographic DNS (GeoDNS)

Return different IPs based on the client's location:

```
Client in Tokyo → cdn.example.com → 203.0.113.1 (Asia server)
Client in London → cdn.example.com → 198.51.100.1 (Europe server)
Client in NYC → cdn.example.com → 192.0.2.1 (US East server)
```

Used by CDNs (CloudFlare, Akamai, AWS Route53). Based on the resolver's IP or EDNS Client Subnet.

### Weighted DNS

Return different IPs with different probabilities:

```
90% of responses → new-server.example.com (canary deployment)
10% of responses → old-server.example.com
```

### Health-check aware DNS

Cloud DNS services (Route53, Cloud DNS) can monitor backend health and remove unhealthy IPs from responses automatically.

---

## Own Resolver

### Why run your own?

1. **Privacy**: Don't send queries to ISP/Google/Cloudflare
2. **Speed**: Local cache = sub-millisecond responses
3. **Custom records**: Override public DNS for internal names
4. **Ad blocking**: Block ad/tracker domains at DNS level

### Unbound (recommended recursive resolver)

```bash
# Install
sudo apt install unbound

# Basic config: /etc/unbound/unbound.conf
server:
    interface: 127.0.0.1
    access-control: 127.0.0.0/8 allow
    do-ip6: no
    prefetch: yes             # pre-fetch expiring records
    cache-min-ttl: 60
    cache-max-ttl: 86400
    
    # DNSSEC validation
    auto-trust-anchor-file: "/var/lib/unbound/root.key"
    
    # Block ads (optional)
    # include: /etc/unbound/blocklist.conf

# Start
sudo systemctl enable --now unbound

# Point system DNS to Unbound
echo "nameserver 127.0.0.1" | sudo tee /etc/resolv.conf

# Test
dig @127.0.0.1 google.com
```

### dnsmasq (simpler, for home/small networks)

```bash
# Install
sudo apt install dnsmasq

# Config: /etc/dnsmasq.conf
listen-address=127.0.0.1
cache-size=1000
server=1.1.1.1
server=8.8.8.8

# Add local records
address=/myapp.local/192.168.1.100

sudo systemctl restart dnsmasq
```

---

## Key Takeaways

1. **DNS message format**: Header (12 bytes) + Question + Answer + Authority + Additional sections
2. **Zone files** define all records for a domain; SOA controls zone behavior and caching
3. **AXFR/IXFR** replicate zones to secondary servers — restrict access to prevent info leaks
4. **Negative caching** caches NXDOMAIN responses — prevents query floods for nonexistent names
5. **DNSSEC** adds cryptographic signatures with a chain of trust from root — prevents spoofing
6. **DoT/DoH** encrypt DNS queries — protect privacy but complicate network debugging
7. **DNS attacks**: Cache poisoning, amplification DDoS, tunneling, rebinding
8. **Split-horizon** returns different answers for internal vs external clients
9. **DNS load balancing**: Round-robin, geographic, weighted, health-checked
10. **Run your own resolver** (Unbound) for privacy, speed, custom records, and ad blocking

---

## Next

→ [03-dns-debugging.md](03-dns-debugging.md) — Practical DNS troubleshooting
