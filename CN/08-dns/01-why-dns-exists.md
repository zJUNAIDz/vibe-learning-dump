# Why DNS Exists — The Internet's Phone Book

> Every time you type "google.com" into a browser, your computer needs to find the IP address 142.250.x.x. DNS (Domain Name System) provides this mapping. It's the largest distributed database in the world, handling trillions of queries per day, and its failure takes down everything.

---

## Table of Contents

1. [The Problem: Humans vs Machines](#the-problem)
2. [Before DNS: The hosts File](#before-dns)
3. [DNS Design Principles](#design-principles)
4. [The DNS Hierarchy](#hierarchy)
5. [DNS Resolution: Step by Step](#resolution)
6. [Recursive vs Iterative Queries](#recursive-vs-iterative)
7. [DNS Caching and TTL](#caching)
8. [Record Types](#record-types)
9. [CNAME vs A vs AAAA](#cname-vs-a)
10. [The Root Servers](#root-servers)
11. [Linux: DNS Resolution Chain](#linux-resolution)
12. [Key Takeaways](#key-takeaways)

---

## The Problem

IP addresses are for machines: `142.250.80.46`
Domain names are for humans: `google.com`

You wouldn't memorize phone numbers for 100 contacts. You'd use a phone book. DNS is the Internet's phone book — it maps human-readable names to machine-routable IP addresses.

```
Without DNS:
  browser → enter 142.250.80.46 → connect

With DNS:
  browser → enter google.com
          → ask DNS: "What IP is google.com?"
          → DNS: "142.250.80.46"
          → connect to 142.250.80.46
```

DNS also does much more: mail routing (MX records), service discovery (SRV records), security (DNSSEC, CAA), content distribution (geography-based answers), load balancing.

---

## Before DNS

### /etc/hosts

Before DNS (pre-1983), every computer on the Internet (then ARPANET) had a file called `hosts.txt` that listed every hostname and IP address.

```
# /etc/hosts (1979 style)
10.0.0.1    sri-nic
10.0.0.73   bbn-tenex
10.1.0.13   xerox-parc
```

This file was maintained centrally at SRI-NIC (Stanford Research Institute Network Information Center). When a new host was added, everyone had to download the updated file.

### Why this broke

1. **Scale**: By 1983, the ARPANET had hundreds of hosts. The file was growing too fast.
2. **Update delay**: Changes took days to propagate
3. **Naming conflicts**: No structure — anyone could claim any name
4. **Single point of failure**: If SRI-NIC was down, no updates
5. **Bandwidth**: Everyone downloading the full file regularly consumed significant bandwidth

Paul Mockapetris designed DNS (RFC 882/883, 1983) to solve all of these.

### /etc/hosts still works

```bash
# /etc/hosts is checked BEFORE DNS (usually)
cat /etc/hosts
# 127.0.0.1   localhost
# 192.168.1.50 myserver.local

# Check resolution order
cat /etc/nsswitch.conf | grep hosts
# hosts: files dns
# "files" (hosts file) is checked first, then "dns"

# Add a custom mapping
echo "192.168.1.100 myapp.test" | sudo tee -a /etc/hosts
# Now "ping myapp.test" resolves to 192.168.1.100
```

---

## Design Principles

DNS was designed with several key architectural choices:

### 1. Hierarchical namespace

Instead of flat names (anyone can claim "mail"), DNS uses a tree structure:

```
                    . (root)
                   ╱|╲
                 ╱  |  ╲
              com  org   net
             ╱ |          |
          google  wikipedia  cloudflare
          ╱   |
       www  mail
```

`www.google.com.` reads right-to-left: root → com → google → www

### 2. Distributed database

No single server stores everything. Each level of the hierarchy is managed by different organizations:
- Root servers: managed by ICANN and 12 organizations
- `.com`: managed by Verisign
- `google.com`: managed by Google
- `www.google.com`: decided by Google's DNS infrastructure

### 3. Caching

DNS responses are cached for a time (TTL — Time To Live). Most queries never reach the authoritative server because the answer is cached closer to the client.

### 4. UDP (primarily)

DNS uses UDP port 53 for queries (small, stateless, fast). TCP port 53 is used for:
- Responses larger than 512 bytes (or ~1232 bytes with EDNS0)
- Zone transfers (replication between DNS servers)
- DNS-over-TLS (port 853) and DNS-over-HTTPS (port 443)

---

## Hierarchy

### Domain name structure

```
        www.mail.google.com.
        │    │     │      │ │
        │    │     │      │ └── Root domain (the trailing dot, usually implicit)
        │    │     │      └──── Top-Level Domain (TLD)
        │    │     └─────────── Second-Level Domain (SLD)
        │    └───────────────── Subdomain
        └────────────────────── Subdomain (hostname)
```

### Types of TLDs

| Type | Examples | Managed By |
|------|----------|-----------|
| Generic (gTLD) | .com, .org, .net, .info | Various registries |
| Country-code (ccTLD) | .uk, .de, .jp, .pk | Country organizations |
| Sponsored (sTLD) | .edu, .gov, .mil | Specific organizations |
| New gTLD | .dev, .app, .xyz, .cloud | Various (since 2013) |
| Infrastructure | .arpa | IANA (reverse DNS) |

### Full Qualified Domain Name (FQDN)

An FQDN includes the trailing dot: `www.google.com.`

The trailing dot means "start from root." Without it, the resolver might append a search domain:

```bash
# If search domain is "mycompany.com":
ping mail         → resolves mail.mycompany.com
ping mail.        → resolves mail. (just "mail" at the root — would fail)
ping mail.google.com   → resolves mail.google.com.mycompany.com FIRST, then tries mail.google.com
ping mail.google.com.  → resolves mail.google.com. (FQDN, no search domain appended)

# Check your search domain
cat /etc/resolv.conf | grep search
# search mycompany.com
```

---

## Resolution

When you type `www.example.com` in a browser, here's what happens:

### Step 1: Check local caches

```
Browser cache → OS DNS cache → /etc/hosts → resolver
```

### Step 2: Ask the recursive resolver

Your computer asks a **recursive resolver** (usually your ISP's DNS or 8.8.8.8 / 1.1.1.1). This server does the hard work of walking the hierarchy.

### Step 3: The resolver walks the tree

```
Client                 Recursive Resolver
  |                         |
  | "What's www.example.com?" |
  |------------------------>|
  |                         |
  |    Resolver to Root Server:
  |    "Who handles .com?"
  |                 Root: "Ask a.gtld-servers.net (Verisign)"
  |                         |
  |    Resolver to .com TLD:
  |    "Who handles example.com?"
  |                 TLD: "Ask ns1.example.com (198.51.100.1)"
  |                         |
  |    Resolver to Authoritative:
  |    "What's www.example.com?"
  |                 Auth: "93.184.216.34, TTL=3600"
  |                         |
  |    "93.184.216.34"      |
  |<------------------------|
```

### The complete resolution involves up to 4 server types:

| Server | Role | Example |
|--------|------|---------|
| **Stub resolver** | Your computer's DNS client | systemd-resolved, glibc |
| **Recursive resolver** | Does the walking, caches results | 8.8.8.8, 1.1.1.1, ISP DNS |
| **Root server** | Knows TLD servers | a.root-servers.net |
| **TLD server** | Knows authoritative servers for domains | a.gtld-servers.net |
| **Authoritative server** | Has the actual records | ns1.example.com |

---

## Recursive vs Iterative

### Recursive query

"Give me the final answer. I don't care how you get it."

```
Client → Recursive Resolver: "Resolve www.example.com"
Recursive Resolver: does all the work, returns final answer
```

The client sends ONE query and gets the final answer. The recursive resolver handles all intermediate queries.

### Iterative query

"Tell me who to ask next."

```
Client → Root: "Resolve www.example.com"
Root → Client: "I don't know, but .com TLD is at 192.5.6.30"
Client → TLD: "Resolve www.example.com"
TLD → Client: "I don't know, but example.com's NS is at 198.51.100.1"
Client → Auth: "Resolve www.example.com"
Auth → Client: "93.184.216.34"
```

The client asks each server and follows the referrals. Root and TLD servers only do iterative (they'd be overwhelmed doing recursive for everyone).

---

## Caching

### TTL (Time To Live)

Every DNS response includes a TTL — how many seconds the answer can be cached.

```bash
# Check TTL of a record
dig google.com
# google.com.    300    IN    A    142.250.80.46
#                ^^^
#                TTL = 300 seconds (5 minutes)

# After 300 seconds, the cached answer expires
# Next query goes back to the authoritative server
```

### TTL trade-offs

| Low TTL (60s) | High TTL (86400s) |
|---|---|
| Changes propagate fast | Changes propagate slow |
| More DNS traffic | Less DNS traffic |
| Good for: failover, blue-green deploy | Good for: stable services |
| Risk: DNS flood if cache too short | Risk: stale records after changes |

### Caching hierarchy

```
Browser cache (Chrome: chrome://net-internals/#dns)
  ↓ miss
OS cache (Linux: systemd-resolved, nscd)
  ↓ miss
Local recursive resolver (ISP, 8.8.8.8, corporate DNS)
  ↓ miss
Authoritative server
```

```bash
# Clear DNS cache on Linux
sudo systemd-resolve --flush-caches
# or
sudo resolvectl flush-caches

# View cache statistics
systemd-resolve --statistics
# or
resolvectl statistics
```

---

## Record Types

| Type | Purpose | Example |
|------|---------|---------|
| **A** | Maps name to IPv4 address | `google.com → 142.250.80.46` |
| **AAAA** | Maps name to IPv6 address | `google.com → 2607:f8b0::200e` |
| **CNAME** | Alias to another name | `www.google.com → google.com` |
| **MX** | Mail server for domain | `gmail.com → alt1.gmail-smtp-in.l.google.com` |
| **NS** | Authoritative nameservers | `google.com → ns1.google.com` |
| **TXT** | Arbitrary text (SPF, DKIM, verification) | `google.com → "v=spf1 include:..."` |
| **SOA** | Zone authority (serial, refresh, retry) | Start of Authority for a zone |
| **SRV** | Service location (port, priority, weight) | `_sip._tcp.example.com → server:5060` |
| **PTR** | Reverse lookup (IP → name) | `46.80.250.142.in-addr.arpa → google.com` |
| **CAA** | Certificate authority authorization | `google.com CAA 0 issue "pki.goog"` |

```bash
# Query specific record types
dig google.com A          # IPv4 address
dig google.com AAAA       # IPv6 address
dig google.com MX         # Mail servers
dig google.com NS         # Nameservers
dig google.com TXT        # Text records (SPF, DKIM)
dig google.com ANY        # All records (many servers block this)

# Query with specific resolver
dig @8.8.8.8 google.com
dig @1.1.1.1 google.com
```

---

## CNAME vs A

### A record

Direct mapping: name → IP address.

```
google.com.    A    142.250.80.46
```

### AAAA record

Direct mapping: name → IPv6 address.

```
google.com.    AAAA    2607:f8b0:4004:800::200e
```

### CNAME record

Alias: name → another name (which eventually resolves to an A/AAAA).

```
www.google.com.    CNAME    google.com.
google.com.        A        142.250.80.46

Resolution: www.google.com → CNAME → google.com → A → 142.250.80.46
```

### CNAME restrictions

A CNAME cannot coexist with other records at the same name:

```
# INVALID:
example.com.    CNAME    other.com.
example.com.    MX       mail.example.com.    ← CONFLICT!

# You CANNOT put a CNAME at the zone apex (example.com)
# because the apex always has NS and SOA records

# Solutions:
# 1. Use A/AAAA at apex
# 2. Many DNS providers offer "ALIAS" or "ANAME" (non-standard pseudo-record)
#    that acts like CNAME at apex
```

---

## Root Servers

There are 13 root server identities (a.root-servers.net through m.root-servers.net), but **not** 13 physical servers. Through Anycast, there are over 1,500 instances worldwide.

```bash
# See all root servers
dig . NS

# Query a root server directly
dig @a.root-servers.net com. NS
# Returns the NS records for .com TLD
```

### Why 13?

The original DNS response had to fit in a 512-byte UDP packet. 13 NS records with their glue (A) records is the maximum that fits.

### Anycast

Each root server IP is announced from multiple locations worldwide. Packets go to the nearest instance:

```
a.root-servers.net (198.41.0.4)
  → Instance in Amsterdam
  → Instance in Miami
  → Instance in Singapore
  → Instance in São Paulo
  → ... dozens more

Your DNS query to 198.41.0.4 goes to whichever instance is closest (BGP determines this)
```

---

## Linux Resolution

### The resolution chain

```
Application
    │
    ▼
glibc (getaddrinfo)
    │
    ▼
/etc/nsswitch.conf → determines order
    │
    ├──► /etc/hosts (files)
    │
    ├──► systemd-resolved or similar (dns)
    │        │
    │        ▼
    │    /etc/resolv.conf → recursive resolver IP
    │        │
    │        ▼
    │    Recursive resolver (8.8.8.8, ISP, etc.)
    │
    └──► mDNS, LLMNR, etc.
```

### /etc/resolv.conf

```bash
cat /etc/resolv.conf
# nameserver 8.8.8.8
# nameserver 1.1.1.1
# search mycompany.com
# options ndots:5 timeout:2 attempts:3

# nameserver: recursive resolver to use (up to 3)
# search: domains to append for short names
# ndots:  if name has fewer than N dots, append search domain first
# timeout: seconds to wait for response
# attempts: number of retries
```

### systemd-resolved

Modern Linux uses systemd-resolved as a local DNS stub resolver and cache:

```bash
# Status
resolvectl status
# Shows: DNS servers, search domains, DNSSEC mode, per-interface settings

# Query (like dig but uses systemd-resolved)
resolvectl query google.com

# Statistics (cache hits, misses)
resolvectl statistics

# DNS used per interface
resolvectl dns

# Flush cache
resolvectl flush-caches
```

### dig, nslookup, host — DNS debugging tools

```bash
# dig (most powerful)
dig google.com              # standard query
dig +short google.com       # just the answer
dig +trace google.com       # follow the full resolution path
dig +norecurse @a.root-servers.net google.com  # iterative query

# nslookup (simpler, cross-platform)
nslookup google.com
nslookup -type=MX gmail.com
nslookup google.com 8.8.8.8  # use specific server

# host (simplest)
host google.com
host -t MX gmail.com
```

---

## Key Takeaways

1. **DNS maps names to IP addresses** — the most critical infrastructure service on the Internet
2. **Hierarchical**: root → TLD → authoritative, each level managed by different organizations
3. **Distributed**: No single server stores everything; queries walk the tree
4. **Recursive resolvers** do the hard work; your computer just asks and gets answers
5. **Caching with TTL** reduces load; most queries are answered from cache
6. **A = IPv4**, **AAAA = IPv6**, **CNAME = alias**, **MX = mail**, **NS = nameservers**
7. **CNAME can't coexist** with other records at the same name (important for zone apex)
8. **13 root server identities**, 1,500+ instances via Anycast
9. **Linux resolution**: nsswitch.conf → /etc/hosts → systemd-resolved → recursive resolver
10. **`dig`** is the essential DNS debugging tool — learn `dig +trace` and `dig +short`

---

## Next

→ [02-dns-deep-dive.md](02-dns-deep-dive.md) — DNS protocol internals, DNSSEC, and advanced topics
