# Real-World Network Failure Scenarios

> Theory is important. But production networks fail in ways textbooks don't cover. This file contains real failure scenarios — the kind you'll encounter at 3 AM when something breaks. Each scenario includes symptoms, investigation steps, root cause, and the fix. Study these patterns; they repeat across every infrastructure.

---

## Table of Contents

1. [How to Read These Scenarios](#how-to-read)
2. [Scenario: The Mystery 502s](#502s)
3. [Scenario: DNS Resolution Takes 5 Seconds](#dns-5s)
4. [Scenario: Connection Resets After 60 Seconds](#conn-reset-60s)
5. [Scenario: The Deployment That Broke Everything](#bad-deploy)
6. [Scenario: Intermittent Connection Failures Between Services](#intermittent)
7. [Scenario: CLOSE_WAIT Leak](#close-wait)
8. [Scenario: TLS Handshake Failure After Certificate Rotation](#tls-fail)
9. [Scenario: Pod Can't Reach External API](#pod-external)
10. [Scenario: The 99th Percentile Latency Spike](#p99)
11. [Scenario: SYN Flood From Your Own Infrastructure](#syn-flood)
12. [Scenario: MTU Black Hole in VPN Tunnel](#mtu-blackhole)
13. [Scenario: Load Balancer Draining Gone Wrong](#lb-drain)
14. [Scenario: The Conntrack Table Is Full](#conntrack-full)
15. [Lessons and Patterns](#lessons)
16. [Key Takeaways](#key-takeaways)

---

## How to Read These Scenarios

```
Each scenario follows this structure:

  ALERT:    What oncall sees (the symptom)
  SYMPTOMS: Observable behavior
  INVESTIGATION: Step-by-step debugging process
  ROOT CAUSE: The actual problem
  FIX: How to resolve it
  PREVENTION: How to prevent it in the future
  
Try to diagnose each one yourself before reading the root cause.
```

---

## Scenario: The Mystery 502s

### Alert

```
"API error rate spiked to 15%. Users seeing intermittent 502 Bad Gateway errors."
```

### Symptoms

```
- 502 errors from nginx reverse proxy
- Not all requests fail — intermittent
- Backend health checks passing
- Started after scaling up the backend from 3 to 6 pods
```

### Investigation

```bash
# Check nginx error logs
tail -f /var/log/nginx/error.log
# "upstream prematurely closed connection while reading response header"
# → nginx opened connection, backend closed it before responding

# Check backend pod logs
kubectl logs <backend-pod> -f
# No errors logged — requests not even reaching application

# Check connection timing
curl -v -o /dev/null http://backend-service:8080/health
# Works 99% of the time

# Check keepalive configuration
# nginx.conf: upstream { keepalive 64; }
# nginx reuses connections to backends

# Check backend: how long does it keep idle connections?
# Backend framework (Express.js): server.keepAliveTimeout = 5000 (5 seconds)
# nginx: keepalive_timeout 60 (60 seconds to client, but what about upstream?)
```

### Root cause

```
RACE CONDITION: nginx keepalive vs backend keepalive timeout.

Timeline:
  0s:   nginx opens connection to backend
  5s:   Backend idle timeout expires → backend sends FIN
  5s:   nginx simultaneously sends request on same connection
  
  Backend: "I'm closing this connection"
  nginx: "I'm sending a request on this connection"
  
  Result: nginx sends request on a closing connection
  → Backend rejects it → nginx returns 502

This is WORSE with more pods because:
  - More pods = more connections in the pool
  - Higher chance of hitting an idle connection
  - Why it started after scaling up
```

### Fix

```
Option 1: Backend keepalive > nginx keepalive
  Backend keepAliveTimeout = 75 seconds
  nginx keepalive_timeout = 60 seconds
  → Backend never closes before nginx does

Option 2: Configure nginx upstream keepalive_timeout
  upstream backend {
    server backend-service:8080;
    keepalive 64;
    keepalive_timeout 60s;
    keepalive_requests 1000;
  }

Rule: Upstream keepalive timeout MUST be longer than 
      the downstream proxy's keepalive timeout.
      Backend > nginx > client (always longer upstream)
```

---

## Scenario: DNS Resolution Takes 5 Seconds

### Alert

```
"API response times spiked to 5+ seconds for external service calls."
```

### Symptoms

```
- p50 latency: 200ms (normal)
- p99 latency: 5,200ms (terrible)
- Only happens on first request to new domains
- Kubernetes pods calling external APIs
```

### Investigation

```bash
# Test from inside the pod
kubectl exec -it <pod> -- sh

# Time DNS resolution
time nslookup api.stripe.com
# real    5.012s  ← 5 SECONDS for DNS!

# Check resolv.conf
cat /etc/resolv.conf
# nameserver 10.96.0.10
# search default.svc.cluster.local svc.cluster.local cluster.local
# options ndots:5

# Test with full FQDN
time nslookup api.stripe.com.
# real    0.015s  ← 15ms! Fast!

# What happened with 5 DNS queries:
dig api.stripe.com.default.svc.cluster.local   # NXDOMAIN (waited)
dig api.stripe.com.svc.cluster.local            # NXDOMAIN (waited)
dig api.stripe.com.cluster.local                # NXDOMAIN (waited)
dig api.stripe.com.                              # SUCCESS (finally!)

# 4 failed queries × ~1s timeout each ≈ 5 seconds
```

### Root cause

```
ndots:5 in Kubernetes resolv.conf.

"api.stripe.com" has 2 dots (< 5).
So resolver tries ALL search domains FIRST:
  1. api.stripe.com.default.svc.cluster.local → NXDOMAIN
  2. api.stripe.com.svc.cluster.local         → NXDOMAIN
  3. api.stripe.com.cluster.local             → NXDOMAIN
  4. api.stripe.com.                          → SUCCESS

DNS server was slow to respond with NXDOMAIN (1-2 seconds each).
Total: 5+ seconds before the real answer.
```

### Fix

```yaml
# Option 1: Use trailing dot in application code
# "api.stripe.com." instead of "api.stripe.com"

# Option 2: Reduce ndots in pod spec
spec:
  dnsConfig:
    options:
    - name: ndots
      value: "1"
  # Trade-off: must use FQDN for cluster services
  # "my-service.default.svc.cluster.local" instead of "my-service"

# Option 3: Use NodeLocal DNSCache
# Caches DNS responses on each node
# Repeated failures resolved instantly from cache
# https://kubernetes.io/docs/tasks/administer-cluster/nodelocaldns/
```

---

## Scenario: Connection Resets After 60 Seconds

### Alert

```
"Long-running gRPC streams disconnecting after exactly 60 seconds."
```

### Symptoms

```
- gRPC bidirectional streams work for exactly 60 seconds
- Then: RST packet received
- Reconnect works, then another disconnect at 60 seconds
- Happens only in production, works fine locally
```

### Investigation

```bash
# Capture packets on the client
sudo tcpdump -i eth0 host <server-ip> -nn -w grpc.pcap

# Analysis in Wireshark:
# ...normal gRPC traffic...
# At 60.001s: RST from 10.0.0.1 (NOT the server IP)
# Source MAC of RST → belongs to AWS NAT Gateway!

# Check: what's between client and server?
traceroute -n <server-ip>
# Hop 2: 10.0.0.1 (NAT gateway)

# Check NAT gateway settings:
# AWS NAT Gateway idle connection timeout: 350 seconds (TCP)
# But AWS ALB idle timeout: 60 seconds (default!)
```

### Root cause

```
AWS Application Load Balancer (ALB) idle timeout = 60 seconds.

gRPC stream has no data flowing for 60 seconds.
ALB considers connection idle → sends RST to both sides.

          Client → ALB → Server
                   │
                   │ 60s idle
                   │
                   └── RST → Client
                   └── RST → Server
```

### Fix

```
Option 1: Increase ALB idle timeout
  aws elbv2 modify-load-balancer-attributes \
    --load-balancer-arn <arn> \
    --attributes Key=idle_timeout.timeout_seconds,Value=3600

Option 2: Enable gRPC keepalive (better solution)
  Client: send keepalive ping every 30 seconds
  Server: allow keepalive pings
  → Connection never goes idle → ALB never resets it

  // Go gRPC client:
  conn, err := grpc.Dial(target,
    grpc.WithKeepaliveParams(keepalive.ClientParameters{
      Time:    30 * time.Second,
      Timeout: 10 * time.Second,
    }),
  )

Option 3: Use NLB instead of ALB for gRPC
  NLB idle timeout: 350 seconds
  NLB passes through TCP (no HTTP parsing overhead)
```

---

## Scenario: The Deployment That Broke Everything

### Alert

```
"After deploying v2.5.0, all API calls to third-party payment 
processor are failing with 'connection reset'."
```

### Symptoms

```
- v2.4.0 worked fine
- v2.5.0: all HTTPS calls to payment API get RST
- curl from the pod works!
- Only the application can't connect
```

### Investigation

```bash
# curl works from inside the pod
kubectl exec -it <pod> -- curl -v https://api.payments.com/health
# HTTP/2 200 OK  ← works fine!

# Application logs show:
# "javax.net.ssl.SSLHandshakeException: Received fatal alert: handshake_failure"

# Check what changed in v2.5.0
git diff v2.4.0..v2.5.0 -- pom.xml Dockerfile
# Found: JDK upgraded from 11 to 17
# Found: TLS config change: enabled TLS 1.3 only

# Check: does payment API support TLS 1.3?
openssl s_client -connect api.payments.com:443 -tls1_3
# Error: no protocols available
# → Payment API does NOT support TLS 1.3!

openssl s_client -connect api.payments.com:443 -tls1_2
# Connected successfully
# → Only supports TLS 1.2
```

### Root cause

```
JDK upgrade changed default TLS configuration.
New JDK 17 defaults to TLS 1.3 only.
Payment processor only supports TLS 1.2.
Handshake fails because no common TLS version.
curl works because curl supports TLS 1.2 fallback.
```

### Fix

```java
// Re-enable TLS 1.2 in Java
SSLContext ctx = SSLContext.getInstance("TLSv1.2");
// Or in application.properties:
server.ssl.enabled-protocols=TLSv1.2,TLSv1.3

// Long-term: ask payment provider to support TLS 1.3
```

---

## Scenario: Intermittent Connection Failures Between Services

### Alert

```
"Service A gets connection refused from Service B about 2% of the time."
```

### Symptoms

```
- 98% of requests: < 50ms, successful
- 2% of requests: immediate "connection refused" (RST)
- Service B pods are healthy
- Happens on all instances of Service A
```

### Investigation

```bash
# Are all Service B endpoints healthy?
kubectl get endpoints service-b
# 10.244.1.5:8080, 10.244.2.8:8080, 10.244.1.12:8080 ← 3 endpoints

# Test each directly
for ip in 10.244.1.5 10.244.2.8 10.244.1.12; do
  echo -n "$ip: "
  curl -s -o /dev/null -w "%{http_code}" http://$ip:8080/health
  echo
done
# 10.244.1.5: 200
# 10.244.2.8: 200
# 10.244.1.12: 200  ← all healthy?

# Check endpoints more carefully
kubectl get endpoints service-b -o yaml
# Noticed: addresses list keeps changing
# 10.244.3.20 appears briefly then disappears

# Check pod events
kubectl get events --field-selector reason=Killing
# Pod service-b-xxx being killed and restarted every 45 seconds!

# Check the crashing pod
kubectl describe pod service-b-xxx
# Restart Count: 147
# Last State: Terminated (OOMKilled)
```

### Root cause

```
One of Service B's pods is OOM-killed repeatedly.

  Timeline:
  1. Pod starts → healthy → added to Service endpoints
  2. Pod handles requests, memory grows
  3. Pod exceeds memory limit → OOMKilled
  4. kube-proxy still has old iptables rules (stale endpoint)
  5. Requests routed to dead pod → connection refused (RST)
  6. Pod restarts → briefly healthy → cycle repeats

  2% failure rate = ~1/3 pods × brief window of being dead
  
  The gap between pod dying and kube-proxy removing 
  the endpoint is the window where requests fail.
```

### Fix

```yaml
# Fix 1: Increase memory limit
resources:
  limits:
    memory: "512Mi"  # was 256Mi

# Fix 2: Add readiness probe (critical!)
readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  
# Readiness probe ensures endpoint is removed BEFORE pod dies
# Without it: pod removed from endpoints only after kube-proxy sync

# Fix 3: Fix the memory leak in Service B
```

---

## Scenario: CLOSE_WAIT Leak

### Alert

```
"Service running out of file descriptors. 'Too many open files' errors."
```

### Symptoms

```
- Application starts rejecting new connections
- Error: "accept: too many open files"
- File descriptor count growing linearly over hours
- Never decreases
```

### Investigation

```bash
# Count open file descriptors
ls /proc/$(pgrep myapp)/fd | wc -l
# 12847  ← way too many

# What kind of FDs?
ls -la /proc/$(pgrep myapp)/fd | head -20
# socket:[12345] → mostly sockets

# Check socket states
ss -tnp | grep myapp | awk '{print $1}' | sort | uniq -c
#      3 ESTAB
#  12800 CLOSE-WAIT      ← 12,800 connections in CLOSE_WAIT!
#      4 TIME-WAIT

# CLOSE_WAIT means:
# Remote side sent FIN (closed their end)
# Local side HASN'T called close() yet
# = Application bug: not closing connections

# Which remote hosts?
ss -tnp state close-wait | awk '{print $4}' | cut -d: -f1 | sort | uniq -c | sort -rn
# 12800 10.96.0.50  ← all to one service (the upstream API)
```

### Root cause

```
Application using HTTP connection pool but not properly 
closing response bodies:

  // Bug:
  resp, err := http.Get("http://api/data")
  if err != nil { return err }
  data := process(resp)  // ← never closed resp.Body!
  
  // The response body must be fully read AND closed
  // Otherwise the connection stays open forever on our side
  // When server times out and sends FIN → we enter CLOSE_WAIT
  // We never close() → CLOSE_WAIT forever → FD leak

  Fix:
  resp, err := http.Get("http://api/data")
  if err != nil { return err }
  defer resp.Body.Close()          // ← always close!
  io.Copy(io.Discard, resp.Body)   // ← drain the body
  data := process(resp)
```

---

## Scenario: TLS Handshake Failure After Certificate Rotation

### Alert

```
"All HTTPS requests to api.example.com failing after certificate 
 renewal. Error: certificate verify failed."
```

### Investigation

```bash
# Check the certificate
openssl s_client -connect api.example.com:443 -servername api.example.com \
  </dev/null 2>/dev/null | openssl x509 -noout -text | head -20
# Issuer: Let's Encrypt R3
# Subject: CN=api.example.com
# Validity: Not Before: today, Not After: 90 days from now
# ← Certificate looks fine!

# Check the certificate CHAIN
openssl s_client -connect api.example.com:443 -servername api.example.com \
  </dev/null 2>&1 | grep -E 'verify|depth|Certificate chain'
# depth=0: CN=api.example.com (OK)
# depth=1: ???
# verify error: unable to get local issuer certificate
# → INCOMPLETE CHAIN!

# The intermediate certificate is missing!
openssl s_client -connect api.example.com:443 </dev/null 2>&1 \
  | grep -c 'BEGIN CERTIFICATE'
# 1  ← only the leaf certificate. Should be 2 (leaf + intermediate)
```

### Root cause

```
Certificate renewal script replaced only the leaf certificate.
Intermediate certificate was not included in the new bundle.

  Clients need:
    Server Certificate (leaf)    ← present ✓
    Intermediate Certificate     ← MISSING ✗
    Root Certificate             ← in client's trust store
    
  Without intermediate:
    Client can't build chain from leaf to root → verification fails
    
  Some clients (like browsers) cache intermediates 
  and may still work → "works in browser, fails in curl"
```

### Fix

```bash
# Concatenate leaf + intermediate into bundle
cat server.crt intermediate.crt > fullchain.crt

# Update nginx
ssl_certificate /etc/nginx/ssl/fullchain.crt;  # full chain
ssl_certificate_key /etc/nginx/ssl/server.key;

# Verify the chain is complete
openssl verify -CAfile /etc/ssl/certs/ca-certificates.crt fullchain.crt
# fullchain.crt: OK

# Test
openssl s_client -connect api.example.com:443 </dev/null 2>&1 \
  | grep -c 'BEGIN CERTIFICATE'
# 2  ← leaf + intermediate ✓
```

---

## Scenario: Pod Can't Reach External API

### Alert

```
"New pods can't reach external APIs. Existing pods work fine."
```

### Investigation

```bash
# Test from failing pod
kubectl exec -it new-pod -- curl -v https://api.external.com
# Connection timed out

# Test from working pod
kubectl exec -it old-pod -- curl -v https://api.external.com
# Works!

# Compare network config
# Failing pod:
kubectl exec new-pod -- ip route
# default via 10.244.0.1 dev eth0  ← OK

kubectl exec new-pod -- cat /etc/resolv.conf
# nameserver 10.96.0.10  ← OK

# Check: what node is the new pod on?
kubectl get pod new-pod -o wide
# Node: node-5 (new node, just added)

# SSH to node-5
# Check iptables NAT rules
sudo iptables -t nat -L POSTROUTING -n | grep MASQUERADE
# (nothing!)  ← No MASQUERADE rule!

# Check on a working node
sudo iptables -t nat -L POSTROUTING -n | grep MASQUERADE
# MASQUERADE all -- 10.244.0.0/16 0.0.0.0/0  ← present!
```

### Root cause

```
New node (node-5) was added to cluster but CNI plugin 
didn't fully initialize:
  - Pod networking works (CNI created veth + bridge)
  - But MASQUERADE iptables rule not created
  - Pods can reach other pods (direct routing works)
  - Pods CAN'T reach external IPs (need NAT to use node's IP)
  
Old pods on old nodes work because those nodes 
have the MASQUERADE rule from initial CNI setup.
```

### Fix

```bash
# Restart CNI on the new node
kubectl delete pod -n kube-system <cni-pod-on-node-5>
# CNI pod recreates, reinitializes iptables rules

# Verify
sudo iptables -t nat -L POSTROUTING -n | grep MASQUERADE
# MASQUERADE present ✓

# Prevention: Add CNI initialization to node bootstrap validation
# Check iptables rules as part of node readiness
```

---

## Scenario: The 99th Percentile Latency Spike

### Alert

```
"p99 latency jumped from 100ms to 2,000ms. p50 is still 50ms."
```

### Investigation

```bash
# Check: is it one pod or all pods?
# Prometheus query: histogram_quantile(0.99, ...) by (pod)
# Result: ALL pods show p99 spike

# Check: is it one endpoint or all?
# Result: Only /api/users is slow

# Check: server processing time
# Application metrics show processing under 50ms for all requests
# → Network is adding the latency!

# Capture on the server
sudo tcpdump -i eth0 port 8080 -w latency.pcap

# Analyze in tshark
tshark -r latency.pcap -Y 'tcp.analysis.retransmission' | wc -l
# 847 retransmissions in 5 minutes!

# When do retransmissions happen?
tshark -r latency.pcap -Y 'tcp.analysis.retransmission' \
  -T fields -e frame.time -e ip.src -e tcp.analysis.rto | head
# Concentrated at :00 and :30 of each minute

# What else happens at those times?
# Checked cronjobs:
kubectl get cronjobs
# backup-job: */1 * * * *  (every minute)
# The backup job runs every 30 seconds (custom schedule)
# and saturates the network interface
```

### Root cause

```
A backup CronJob runs every 30 seconds and transfers
large files, saturating the node's network interface.

During backup: network queue fills → packet drops → retransmissions
  → retransmission adds 200ms-2000ms to affected requests
  → Only impacts p99 because most requests complete between intervals

p50 unaffected: most requests don't coincide with backup
p99 spiked: ~2% of requests hit the backup window
```

### Fix

```
1. Move backup to off-peak hours
2. Rate-limit backup traffic (tc qdisc on backup pod)
3. Use separate network interface for backup traffic
4. Use resource limits on backup pod (network bandwidth)
```

---

## Scenario: SYN Flood From Your Own Infrastructure

### Alert

```
"Database connection rejected — too many connections. 
 Also: nstat shows SyncookiesSent > 0 on database server."
```

### Investigation

```bash
# On database server
ss -tn | awk '{print $1}' | sort | uniq -c
#  5000 ESTABLISHED
#  8000 SYN-RECV
#  2000 TIME-WAIT

# 8000 connections in SYN-RECV?! That's a SYN flood.
# But from where?

ss -tn state syn-recv | awk '{print $4}' | cut -d: -f1 | sort | uniq -c | sort -rn
#  7500 10.244.1.0/24  ← all from Kubernetes pods!
#   300 10.244.2.0/24
#   200 10.244.3.0/24

# Check: what pods are on 10.244.1.x?
# Pods from the API deployment — hundreds of pods reconnecting

# Application logs on API pods:
# "connection pool exhausted, creating new connection"
# "connection to database failed: too many connections"
# "retrying connection in 100ms"
# → Connection storm! Pods retrying connections aggressively
```

### Root cause

```
Database restart triggered a connection storm:

  1. Database restarted (planned maintenance)
  2. All 200 API pods lost connections simultaneously
  3. All 200 pods retry immediately
  4. Each pod opens 10 connections → 2,000 simultaneous SYN
  5. Database SYN queue overwhelmed → SYN cookies activated
  6. Some connections fail → pods retry immediately → more SYN
  7. Thundering herd amplifies the problem
  
  The retry logic had NO backoff:
    while (!connected) {
      connect();  // immediate retry!
    }
```

### Fix

```
1. Add exponential backoff with jitter:
   retryDelay = min(baseDelay * 2^attempt + random, maxDelay)
   
2. Add connection pool limits:
   Max connections: 10 per pod
   Max wait time: 5 seconds
   
3. Implement circuit breaker:
   After 5 failures → stop trying for 30 seconds
   
4. Increase database SYN queue:
   net.ipv4.tcp_max_syn_backlog = 4096
   net.core.somaxconn = 4096
```

---

## Scenario: MTU Black Hole in VPN Tunnel

### Alert

```
"Can ping servers across VPN, but wget/curl hangs."
```

### Symptoms

```
- ping works (small packets: 84 bytes)
- SSH login works (small initial packets)
- SSH: typing commands works
- SSH: `ls` on small directory works
- SSH: `ls` on large directory HANGS
- wget: connects but never downloads data
- curl: prints headers, then hangs
```

### Investigation

```bash
# Test with different packet sizes
ping -s 56 -M do -c 1 <vpn-server>    # 84 total → OK
ping -s 1000 -M do -c 1 <vpn-server>  # 1028 total → OK
ping -s 1400 -M do -c 1 <vpn-server>  # 1428 total → OK
ping -s 1472 -M do -c 1 <vpn-server>  # 1500 total → FAIL!

# What's the VPN interface MTU?
ip link show tun0
# mtu 1500  ← WRONG! VPN adds encapsulation overhead (~100 bytes)

# Effective MTU through VPN should be ~1400
# But TCP negotiated MSS based on interface MTU 1500:
# MSS = 1500 - 40 = 1460

# So TCP sends 1460-byte segments
# VPN encapsulates: 1460 + 40 (TCP/IP) + 100 (VPN) = 1600 bytes
# Doesn't fit in physical MTU (1500)!
# DF bit set → packet dropped, ICMP "Fragmentation Needed" sent
# But: firewall is blocking ICMP! → sender never learns of the problem
# → PMTUD Black Hole
```

### Root cause

```
VPN interface MTU = 1500 (same as physical)
But VPN adds ~100 bytes encapsulation.
Large packets (> ~1400 bytes) can't fit after encapsulation.
ICMP "Fragmentation Needed" blocked by firewall.
TCP never learns the real MTU → keeps sending large segments → dropped.

Small packets work: ping (84 bytes), SSH keystrokes (~100 bytes)
Large packets fail: bulk data, file transfers, large HTTP responses
```

### Fix

```bash
# Fix 1: Set correct MTU on VPN interface
sudo ip link set tun0 mtu 1400

# Fix 2: Clamp MSS in iptables (works even if you can't fix MTU)
sudo iptables -t mangle -A FORWARD -p tcp \
  --tcp-flags SYN,RST SYN -j TCPMSS --clamp-mss-to-pmtu

# Fix 3: STOP BLOCKING ICMP!
# Allow ICMP type 3 code 4 (Fragmentation Needed) through firewall
sudo iptables -A INPUT -p icmp --icmp-type 3/4 -j ACCEPT

# Fix 4: In WireGuard, set MTU explicitly
[Interface]
MTU = 1420
```

---

## Scenario: Load Balancer Draining Gone Wrong

### Alert

```
"During deployment rollout, users see 502 errors for 10-30 seconds."
```

### Symptoms

```
- Rolling deployment: old pods terminate, new pods start
- During termination: 502 errors
- Old pod receives SIGTERM, starts shutting down
- Load balancer still sends traffic to terminating pod
```

### Root cause

```
Pod receives SIGTERM → starts shutting down → stops accepting connections
But: iptables rules still route traffic to this pod
Gap: 1-10 seconds between SIGTERM and endpoint removal

  Timeline:
  0s: kubectl deletes pod
  0s: kubelet sends SIGTERM to pod
  0s: Pod starts shutting down, closes listener
  0-5s: kube-proxy hasn't updated iptables yet
  → During this window, new connections get RST (connection refused)
  5s: kube-proxy removes endpoint from iptables
  → New connections go to other pods ✓
  
  User experience: 502 errors for 1-5 seconds during deployment
```

### Fix

```yaml
# Add preStop hook with sleep
spec:
  containers:
  - name: app
    lifecycle:
      preStop:
        exec:
          command: ["sh", "-c", "sleep 10"]
    # Pod keeps running for 10 seconds after SIGTERM
    # Gives kube-proxy time to remove endpoints
    # Then app receives SIGTERM and shuts down

  # Also: handle SIGTERM gracefully in the app
  # Continue serving in-flight requests
  # Stop accepting NEW connections
  # Drain existing connections
  terminationGracePeriodSeconds: 30
```

---

## Scenario: The Conntrack Table Is Full

### Alert

```
"Random packet drops across the cluster. dmesg shows:
 'nf_conntrack: table full, dropping packet'"
```

### Symptoms

```
- Intermittent connection failures cluster-wide
- Affects TCP, UDP, and ICMP
- dmesg: "nf_conntrack: table full, dropping packet"
- High-traffic node handling many NAT connections
```

### Investigation

```bash
# Check conntrack usage
cat /proc/sys/net/netfilter/nf_conntrack_count
# 262000
cat /proc/sys/net/netfilter/nf_conntrack_max
# 262144
# ← 99.9% full!

# What's filling the table?
conntrack -L | awk '{print $3}' | sort | uniq -c | sort -rn
# 180000 tcp
#  60000 udp
#  22000 icmp

# TCP breakdown by state
conntrack -L -p tcp | awk '{print $4}' | sort | uniq -c | sort -rn
# 150000 TIME_WAIT    ← most entries are TIME_WAIT!
#  20000 ESTABLISHED
#  10000 CLOSE_WAIT

# TIME_WAIT entries staying for:
cat /proc/sys/net/netfilter/nf_conntrack_tcp_timeout_time_wait
# 120  ← 120 seconds! (too long for high-traffic)
```

### Fix

```bash
# Increase conntrack table size
sudo sysctl -w net.netfilter.nf_conntrack_max=1048576

# Reduce TIME_WAIT timeout in conntrack
sudo sysctl -w net.netfilter.nf_conntrack_tcp_timeout_time_wait=30

# Reduce other timeouts
sudo sysctl -w net.netfilter.nf_conntrack_tcp_timeout_close_wait=30
sudo sysctl -w net.netfilter.nf_conntrack_tcp_timeout_fin_wait=30

# Make persistent in /etc/sysctl.d/99-conntrack.conf
net.netfilter.nf_conntrack_max = 1048576
net.netfilter.nf_conntrack_tcp_timeout_time_wait = 30

# Long-term: consider Cilium (eBPF-based, no conntrack for pod traffic)
```

---

## Lessons and Patterns

```
Recurring themes across all failures:

1. TIMEOUTS CAUSE MORE DAMAGE THAN CRASHES
   A crash is instant: detect and failover.
   A timeout wastes seconds/minutes of user time.
   Always set aggressive timeouts.

2. RACE CONDITIONS AT BOUNDARIES
   Proxy keepalive vs backend keepalive (502s)
   kube-proxy sync vs pod termination (deployment 502s)
   DNS search vs actual resolution (5s DNS)
   → Any time two systems have independent timers, there's a race.

3. THUNDERING HERD
   One failure → all clients retry simultaneously → amplified failure.
   ALWAYS use exponential backoff with jitter.

4. MISSING INTERMEDIATE CERTIFICATES
   Most common TLS issue in production.
   Always test with: openssl s_client, not just a browser.
   Browsers cache intermediates; curl/code doesn't.

5. MTU ISSUES ARE INVISIBLE
   Small packets work, large packets silently fail.
   Always test with large payloads after network changes.
   NEVER block ICMP type 3 code 4.

6. RESOURCE LIMITS SAVE CLUSTERS
   No memory limit → OOM kill chain
   No conntrack tuning → table full → cluster-wide drops
   No connection pool limit → thundering herd
```

---

## Key Takeaways

1. **502 Bad Gateway usually means keepalive mismatch** — upstream server must have a LONGER keepalive timeout than the proxy in front of it
2. **Kubernetes ndots:5 causes 4 extra DNS queries for external domains** — use trailing dots or reduce ndots. This is the #1 Kubernetes performance gotcha
3. **Exactly-timed disconnects = timeout, not failure** — 60 second disconnects = load balancer idle timeout, 120 seconds = conntrack timeout, 300 seconds = NAT timeout
4. **CLOSE_WAIT leaks mean the application isn't closing connections** — always close HTTP response bodies, database connections, and file handles
5. **Incomplete certificate chains work in browsers but fail in code** — browsers cache intermediates, but curl, Go, Java, and Node.js don't. Always test with `openssl s_client`
6. **Connection storms happen when all clients retry simultaneously** — exponential backoff with jitter is mandatory for any retry logic
7. **PMTUD black holes happen when ICMP is blocked** — never block ICMP Fragmentation Needed (type 3, code 4). Small packets work, large ones silently die
8. **Kubernetes pod termination has a race condition** — add `preStop: sleep 10` to give kube-proxy time to remove endpoints before the pod stops serving
9. **Conntrack table full = cluster-wide packet drops** — monitor `nf_conntrack_count` vs `nf_conntrack_max`. Tune timeouts for high-traffic nodes
10. **The question isn't IF these failures will happen, but WHEN** — every production system eventually encounters these patterns. Knowing them speeds recovery from hours to minutes

---

## Next Module

→ [Module 19: Capstone — Follow a Packet](../19-capstone-follow-packet/01-follow-a-packet.md) — Trace a packet's journey through every layer of the stack
