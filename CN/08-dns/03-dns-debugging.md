# DNS Debugging — Practical Troubleshooting

> "It's always DNS." This joke exists because DNS is involved in almost every network operation. When DNS breaks, everything breaks — and the symptoms are often misleading. This file teaches you to systematically diagnose and fix DNS issues.

---

## Table of Contents

1. [DNS Debugging Mindset](#mindset)
2. [The Essential Tool: dig](#dig)
3. [dig +trace: Follow the Resolution Path](#dig-trace)
4. [Common DNS Failures](#common-failures)
5. [Debugging Resolution Order](#resolution-order)
6. [TTL and Caching Issues](#ttl-issues)
7. [Propagation Delays](#propagation)
8. [DNS and Docker/Kubernetes](#containers)
9. [Performance Diagnosis](#performance)
10. [Quick DNS Cheat Sheet](#cheat-sheet)
11. [Key Takeaways](#key-takeaways)

---

## Mindset

When something "isn't resolving," follow this systematic approach:

```
1. What EXACTLY is the error?
   - NXDOMAIN (name doesn't exist)?
   - SERVFAIL (server error)?
   - Timeout (no response)?
   - Wrong IP (stale cache)?

2. WHERE does resolution fail?
   - Local machine (/etc/hosts, local cache)?
   - Recursive resolver (8.8.8.8)?
   - Authoritative server?

3. Is it affecting everyone or just me?
   - Try multiple resolvers (8.8.8.8, 1.1.1.1, ISP)
   - Ask someone on a different network

4. Did it recently change?
   - Check TTL — maybe old cached entry
   - Check recent DNS record changes
```

---

## dig

`dig` (Domain Information Groper) is THE DNS debugging tool. Learn it deeply.

### Basic query

```bash
dig example.com

# Output explained:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 12345
;; flags: qr rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1

;; QUESTION SECTION:
;example.com.                   IN      A

;; ANSWER SECTION:
example.com.            3600    IN      A       93.184.216.34

;; Query time: 25 msec
;; SERVER: 127.0.0.53#53(127.0.0.53)
;; WHEN: Mon Jan 01 12:00:00 UTC 2024
;; MSG SIZE  rcvd: 56
```

### Key dig flags

```bash
# Short output (just the answer)
dig +short example.com
# 93.184.216.34

# Specific record type
dig example.com AAAA     # IPv6
dig example.com MX       # Mail
dig example.com NS       # Nameservers
dig example.com TXT      # Text records
dig example.com SOA      # Authority info
dig example.com ANY      # All records (often blocked)

# Specific resolver
dig @8.8.8.8 example.com
dig @1.1.1.1 example.com
dig @ns1.example.com example.com  # ask authoritative directly

# Reverse lookup (IP → name)
dig -x 8.8.8.8
# 8.8.8.8.in-addr.arpa.  IN  PTR  dns.google.

# No recursion (ask THIS server only, don't follow referrals)
dig +norecurse @a.root-servers.net com. NS

# Show DNSSEC records
dig +dnssec example.com

# Disable EDNS0 (for compatibility testing)
dig +noedns example.com

# TCP instead of UDP
dig +tcp example.com

# Set timeout and retries
dig +timeout=2 +tries=3 example.com
```

---

## dig +trace

`dig +trace` follows the entire resolution path from root to authoritative, showing every step:

```bash
dig +trace www.google.com

# Output (simplified):
; <<>> DiG 9.18 <<>> +trace www.google.com
.                       518400  IN  NS  a.root-servers.net.
.                       518400  IN  NS  b.root-servers.net.
...
;; Received from 127.0.0.53#53 in 0 ms

com.                    172800  IN  NS  a.gtld-servers.net.
com.                    172800  IN  NS  b.gtld-servers.net.
...
;; Received from 198.41.0.4#53(a.root-servers.net) in 24 ms

google.com.             172800  IN  NS  ns1.google.com.
google.com.             172800  IN  NS  ns2.google.com.
...
;; Received from 192.5.6.30#53(a.gtld-servers.net) in 15 ms

www.google.com.         300     IN  A   142.250.80.100
;; Received from 216.239.34.10#53(ns2.google.com) in 8 ms
```

### Interpreting +trace

Each section shows a referral to the next level:
1. Root servers → referral to `.com` TLD servers
2. `.com` TLD → referral to `google.com` authoritative
3. `google.com` authoritative → final answer

If resolution breaks at a specific level, you know exactly where the problem is.

---

## Common Failures

### 1. NXDOMAIN — Name does not exist

```bash
dig nonexistent.example.com
# status: NXDOMAIN

# Common causes:
# - Typo in domain name
# - Record hasn't been created yet
# - Record was deleted
# - Wrong search domain appended

# Debugging:
# Check the authoritative server directly
dig @ns1.example.com nonexistent.example.com
# If authoritative says NXDOMAIN → record truly doesn't exist

# Check if resolver cache has stale negative cache
dig @8.8.8.8 nonexistent.example.com  # try different resolver
```

### 2. SERVFAIL — Server failure

```bash
dig broken.example.com
# status: SERVFAIL

# Common causes:
# - DNSSEC validation failure (signature expired/invalid)
# - Authoritative server unreachable
# - Authoritative server misconfigured
# - Lame delegation (NS records point to server that doesn't serve the zone)

# Debugging:
# Disable DNSSEC validation to check if that's the issue
dig +cd broken.example.com
# +cd = Checking Disabled (skip DNSSEC validation)
# If this returns an answer but without +cd it's SERVFAIL → DNSSEC problem

# Query authoritative directly
dig @ns1.broken.example.com broken.example.com
# Timeout → authoritative server is down
# Answer → issue is between resolver and authoritative
```

### 3. Timeout — No response

```bash
dig example.com
# ;; connection timed out; no servers could be reached

# Common causes:
# - DNS server is down
# - Firewall blocking port 53
# - Network connectivity issue

# Debugging:
# Can you reach the DNS server at all?
ping 8.8.8.8
nc -zvu 8.8.8.8 53   # test UDP port 53
nc -zv 8.8.8.8 53    # test TCP port 53

# Try a different resolver
dig @1.1.1.1 example.com

# Check if local resolver is running
systemctl status systemd-resolved
ss -ulnp | grep 53
```

### 4. Wrong IP — Stale cache

```bash
# You changed DNS records but still getting old IP
dig example.com
# Returns old IP

# Check TTL — might be cached
dig @8.8.8.8 example.com
# Check the TTL value — if it's close to what you set, it was recently refreshed
# If it's counting down from a large number, it's a cached old entry

# Verify at authoritative
dig @ns1.example.com example.com
# This always returns the current record (no caching)

# Solution: Wait for TTL to expire, or flush caches
sudo resolvectl flush-caches     # local
# Can't flush 8.8.8.8's cache — must wait for TTL
```

### 5. Different results from different resolvers

```bash
dig @8.8.8.8 example.com +short    # 1.2.3.4
dig @1.1.1.1 example.com +short    # 5.6.7.8
dig @ns1.example.com example.com +short  # 5.6.7.8

# 8.8.8.8 has a stale cache (old TTL not expired yet)
# 1.1.1.1 has the new record
# Authoritative confirms the new record is correct

# This is normal during DNS propagation
```

---

## Resolution Order

### Linux name resolution is complex

```bash
# Check resolution order
cat /etc/nsswitch.conf | grep hosts
# hosts: files mymachines myhostname resolve [!UNAVAIL=return] dns

# This means:
# 1. files      → /etc/hosts
# 2. mymachines → systemd-machined (containers)
# 3. myhostname → systemd hostname resolution
# 4. resolve    → systemd-resolved
# 5. dns        → traditional DNS (/etc/resolv.conf)
```

### getent vs dig vs nslookup

These tools use DIFFERENT resolution paths:

```bash
# getent uses glibc → follows /etc/nsswitch.conf → checks /etc/hosts FIRST
getent hosts myserver
# Returns: whatever /etc/hosts says (if entry exists)

# dig bypasses glibc entirely → queries DNS directly
dig myserver
# Returns: whatever DNS says (ignores /etc/hosts)

# nslookup uses its own DNS implementation (not glibc)
nslookup myserver
# Returns: whatever DNS says (ignores /etc/hosts)

# This is why "dig works but the app doesn't" or vice versa happens!
# Apps use getent/glibc → check /etc/hosts
# dig queries DNS directly → bypasses /etc/hosts
```

### Debugging which path an application uses

```bash
# Use strace to see which resolution method an app uses
strace -e network curl https://example.com 2>&1 | head -30
# Look for connect() calls to port 53 (DNS) or /etc/hosts reads

# Or with ltrace
ltrace -e getaddrinfo curl https://example.com 2>&1 | head
```

---

## TTL Issues

### "I changed the record but it's not updating"

```bash
# Step 1: Check what the authoritative says (ground truth)
dig @ns1.example.com example.com +short

# Step 2: Check what a public resolver says
dig @8.8.8.8 example.com

# Step 3: Look at the TTL in the response
dig example.com
# example.com.  1800  IN  A  1.2.3.4
#               ^^^^
#               1800 seconds remaining in cache

# If the old TTL was 86400 (24 hours), worst case you wait 24 hours

# Before changing important records:
# 1. Lower TTL to 60 seconds
# 2. Wait for old TTL to expire (e.g., 24 hours if old TTL was 86400)
# 3. Make the change
# 4. Now everyone gets the new record within 60 seconds
# 5. After propagation, optionally increase TTL again
```

### DNS propagation timeline

```
T-24h:  Set TTL to 60s on old record
T-0:    Change IP in DNS record (TTL already 60s everywhere)
T+60s:  Most resolvers have new record
T+5m:   Virtually all resolvers have new record
T+1h:   Raise TTL back to 3600 if desired
```

---

## Propagation

"DNS propagation" is a misleading term. DNS doesn't "push" changes. Instead, cached records expire and resolvers re-query.

### Why "propagation" is slow

Each resolver independently caches records and respects TTL. A resolver that cached your record moments before you changed it will serve the old record until its TTL expires.

```
Resolver A: Cached at T-10s, TTL=3600 → serves old record until T+3590
Resolver B: Cached at T-3000s, TTL=3600 → serves old record until T+600
Resolver C: Cache expired, re-queries → gets new record immediately
```

### Check propagation status

```bash
# Query from multiple global locations
for server in 8.8.8.8 1.1.1.1 9.9.9.9 208.67.222.222; do
    echo "$server: $(dig @$server +short example.com)"
done

# Or use online tools:
# https://www.whatsmydns.net
# https://dnschecker.org
```

---

## Containers

### Docker DNS

Docker containers use Docker's embedded DNS (127.0.0.11):

```bash
# Inside a Docker container
cat /etc/resolv.conf
# nameserver 127.0.0.11
# ndots:0

# Docker resolves:
# - Container names (within same network)
# - Service names (in Docker Compose)
# - External names (forwards to host DNS)

# Common issue: "can't resolve host"
# Check: Is the container on the right Docker network?
docker network inspect bridge

# Debug DNS inside container
docker exec -it mycontainer dig google.com
# No dig? Try:
docker exec -it mycontainer nslookup google.com
# No nslookup? Try:
docker exec -it mycontainer cat /etc/resolv.conf
docker exec -it mycontainer getent hosts google.com
```

### Kubernetes DNS

Kubernetes runs CoreDNS as the cluster DNS:

```bash
# Pods get DNS configured automatically
kubectl exec -it mypod -- cat /etc/resolv.conf
# nameserver 10.96.0.10         ← CoreDNS ClusterIP
# search default.svc.cluster.local svc.cluster.local cluster.local
# ndots:5

# Service discovery via DNS:
# my-service                  → my-service.default.svc.cluster.local
# my-service.other-namespace  → my-service.other-namespace.svc.cluster.local

# Debug DNS in Kubernetes
kubectl run dnstest --image=busybox --rm -it -- nslookup my-service
kubectl run dnstest --image=busybox --rm -it -- nslookup kubernetes.default

# Check CoreDNS status
kubectl get pods -n kube-system -l k8s-app=kube-dns
kubectl logs -n kube-system -l k8s-app=kube-dns

# ndots:5 gotcha:
# With ndots:5, any name with <5 dots is tried with search domains FIRST
# "api.external.com" has 2 dots (< 5)
# Kubernetes tries:
#   api.external.com.default.svc.cluster.local (NXDOMAIN)
#   api.external.com.svc.cluster.local (NXDOMAIN)
#   api.external.com.cluster.local (NXDOMAIN)
#   api.external.com (finally!)
# = 4 unnecessary queries before the real one
# Fix: Use FQDN with trailing dot: "api.external.com."
```

---

## Performance

### Measuring DNS latency

```bash
# Simple timing
time dig google.com > /dev/null
# real    0m0.025s

# dig shows query time
dig google.com | grep "Query time"
# ;; Query time: 12 msec

# Test multiple queries
for i in $(seq 1 10); do
    dig google.com | grep "Query time" | awk '{print $4}'
done

# Measure uncached vs cached
resolvectl flush-caches
dig example.com | grep "Query time"    # uncached: ~50ms
dig example.com | grep "Query time"    # cached: ~0ms
```

### DNS performance optimization

```bash
# 1. Use a local caching resolver (systemd-resolved, unbound)
# Most queries hit cache → <1ms response

# 2. Pre-fetch expiring records (Unbound)
# In unbound.conf:
# prefetch: yes
# prefetch-key: yes

# 3. Connection reuse (for TCP/DoT/DoH)
# Persistent connections to resolver avoid handshake overhead

# 4. Minimize queries (in applications)
# Cache DNS results in your application
# Use connection pools (reuse connections instead of re-resolving)

# 5. Check for query storms
tcpdump -n port 53 -c 100 | awk '{print $5}' | sort | uniq -c | sort -rn | head
# Shows which domains are queried most
# If one domain appears thousands of times → misconfiguration
```

---

## Cheat Sheet

```bash
# ─── Query Commands ──────────────────────────────────────
dig example.com                  # Standard A query
dig example.com AAAA             # IPv6
dig example.com MX               # Mail servers
dig example.com NS               # Nameservers
dig +short example.com           # Just the answer
dig @8.8.8.8 example.com        # Use specific resolver
dig +trace example.com           # Full resolution trace
dig -x 8.8.8.8                   # Reverse lookup

# ─── Debugging Commands ─────────────────────────────────
dig +norecurse @ns1.example.com example.com   # Non-recursive
dig +cd example.com              # Skip DNSSEC validation
dig +tcp example.com             # Force TCP
dig example.com SOA              # Zone authority info

# ─── Local Resolution ───────────────────────────────────
getent hosts example.com         # Uses system resolution (incl. /etc/hosts)
resolvectl query example.com     # Uses systemd-resolved
resolvectl status                # DNS configuration
resolvectl flush-caches          # Clear local cache

# ─── Monitoring ─────────────────────────────────────────
tcpdump -n port 53               # Watch DNS traffic
ss -ulnp | grep 53               # What's listening on port 53
cat /etc/resolv.conf             # Current DNS configuration

# ─── Record format reminder ─────────────────────────────
# NAME    TTL    CLASS   TYPE    DATA
# example.com.  300    IN      A       93.184.216.34
```

---

## Key Takeaways

1. **`dig` is the essential tool** — learn `+short`, `+trace`, `@server`, record types
2. **`dig` vs `getent`**: dig queries DNS directly; getent follows system resolution (incl. /etc/hosts)
3. **SERVFAIL** is often DNSSEC — try `+cd` to test
4. **"DNS propagation"** is really "cache expiration" — lower TTL BEFORE making changes
5. **Kubernetes ndots:5** causes 4+ unnecessary queries per external lookup — use trailing dots for FQDNs
6. **resolution order matters**: /etc/nsswitch.conf controls whether /etc/hosts or DNS wins
7. **Check authoritative directly** (`dig @ns1.example.com`) to see ground truth vs cached data
8. **Pre-flight checklist for DNS changes**: Lower TTL → Wait for old TTL → Make change → Verify → Restore TTL

---

## Next

→ [../09-application-layer/01-http-lifecycle.md](../09-application-layer/01-http-lifecycle.md) — HTTP: the protocol that powers the web
