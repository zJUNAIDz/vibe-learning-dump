# Reading TCP Streams and Detecting Issues

> Knowing the tools is one thing. Knowing what to LOOK FOR is another. This file teaches you the pattern recognition for TCP analysis: how to read a stream, spot problems, and diagnose the root cause. These are the patterns that experienced network engineers use to solve real production issues in minutes.

---

## Table of Contents

1. [The Approach: Systematic TCP Analysis](#approach)
2. [Normal TCP Behavior — Your Baseline](#baseline)
3. [Pattern: Retransmissions](#retransmissions)
4. [Pattern: Duplicate ACKs and Fast Retransmit](#dup-acks)
5. [Pattern: Zero Window](#zero-window)
6. [Pattern: RST (Reset)](#rst)
7. [Pattern: Connection Timeout](#timeout)
8. [Pattern: Half-Open / Half-Closed Connections](#half-open)
9. [Pattern: Small Packets (Nagle Problem)](#nagle)
10. [Pattern: Delayed ACKs + Nagle Interaction](#delayed-ack)
11. [Pattern: MTU / Fragmentation Issues](#mtu)
12. [Pattern: Connection Storms](#storms)
13. [Putting It All Together — Case Studies](#case-studies)
14. [Analysis Checklist](#checklist)
15. [Key Takeaways](#key-takeaways)

---

## The Approach: Systematic TCP Analysis

```
When you have a packet capture and a problem to diagnose:

1. START WITH EXPERT INFO
   Wireshark: Analyze → Expert Information
   tshark: tshark -r capture.pcap -q -z expert
   → Immediate overview of all anomalies

2. CHECK CONVERSATIONS
   Statistics → Conversations → TCP
   → Find the connection(s) related to the problem

3. FILTER TO THE PROBLEM FLOW
   ip.addr == X.X.X.X and tcp.port == YYYY
   → Isolate the relevant stream

4. LOOK AT THE HANDSHAKE
   Was it normal? How long did it take?
   → SYN → SYN-ACK → ACK timing reveals network latency

5. FOLLOW THE DATA FLOW
   Follow → TCP Stream
   → See the application-level conversation

6. CHECK FOR ANOMALIES
   tcp.analysis.retransmission
   tcp.analysis.zero_window
   tcp.analysis.duplicate_ack
   tcp.analysis.out_of_order
   tcp.flags.reset == 1
   → Find the problem events

7. CORRELATE TIMING
   When did the anomaly happen relative to the application issue?
   IO Graph with multiple overlaid metrics
```

---

## Normal TCP Behavior — Your Baseline

### Know what normal looks like

```
Normal connection lifecycle:

Time     Source → Dest        Flags    Notes
─────    ──────────────────   ─────    ─────────────────
0.000    Client → Server      [S]      SYN (seq=0)
0.012    Server → Client      [S.]     SYN-ACK (seq=0, ack=1)
0.013    Client → Server      [.]      ACK (ack=1)
                                        ← Handshake: 13ms = healthy
0.014    Client → Server      [P.]     Data (GET /api/health)
0.015    Server → Client      [.]      ACK
0.018    Server → Client      [P.]     Data (HTTP 200 OK)
0.019    Client → Server      [.]      ACK
0.020    Client → Server      [F.]     FIN (close)
0.021    Server → Client      [F.]     FIN
0.022    Client → Server      [.]      ACK

Normal characteristics:
  - Handshake completes in ~1 RTT (same-region: 1-20ms)
  - ACKs arrive within 0-40ms of data
  - No retransmissions
  - No zero windows
  - No RSTs (except clean connection closure in some apps)
  - Window sizes adequate (not near zero)
  - Sequence numbers always increasing
```

---

## Pattern: Retransmissions

### What it looks like

```
Time     Source → Dest        Flags    Seq        Notes
─────    ──────────────────   ─────    ───        ─────────────
0.100    Client → Server      [P.]     1000:2448  Original data
0.300    Client → Server      [P.]     1000:2448  ← RETRANSMIT (200ms later)
0.700    Client → Server      [P.]     1000:2448  ← RETRANSMIT (400ms, doubled)
1.500    Client → Server      [P.]     1000:2448  ← RETRANSMIT (800ms, doubled)

Key indicators:
  - Same sequence number appears multiple times
  - Time between retransmits doubles (exponential backoff)
  - First retransmit after ~200ms (initial RTO)
```

### Root causes

```
1. Packet loss (most common)
   - Congested link dropping packets
   - NIC ring buffer overflow
   - Firewall drop
   
   Evidence: Server never ACKs the original
   Verify: Check drop counters (ethtool -S, nstat)

2. ACK loss (response lost on return path)
   - Server sent ACK but it was lost
   - Asymmetric routing issue
   
   Evidence: Server-side capture shows ACK was sent
   Verify: Capture on BOTH sides

3. Delayed ACK + Nagle interaction
   - ACK is delayed (up to 40ms)
   - Combined with Nagle, can cause 200ms+ delay
   - Not actual loss — just slow
   
   Evidence: ACK arrives eventually, but late
   Verify: Check tcp.analysis.ack_rtt values

Impact on application:
  1 retransmit = ~200ms added latency (RTO)
  2 retransmits = ~600ms total
  3 retransmits = ~1400ms total
  6 retransmits = ~63 seconds → connection times out
```

---

## Pattern: Duplicate ACKs and Fast Retransmit

### What it looks like

```
Time     Source → Dest        Flags    Seq/Ack        Notes
─────    ──────────────────   ─────    ──────         ─────────────
0.100    Sender → Receiver    [.]      seq 1000:2448  Segment 1
0.101    Sender → Receiver    [.]      seq 2448:3896  Segment 2
0.102    Sender → Receiver    [.]      seq 3896:5344  Segment 3 ← LOST
0.103    Sender → Receiver    [.]      seq 5344:6792  Segment 4 (arrives)
0.104    Receiver → Sender    [.]      ack 3896       ← ACK for 1+2
0.105    Receiver → Sender    [.]      ack 3896       ← Dup ACK #1 (got 4, missing 3)
0.106    Receiver → Sender    [.]      ack 3896       ← Dup ACK #2
0.107    Receiver → Sender    [.]      ack 3896       ← Dup ACK #3
0.108    Sender → Receiver    [.]      seq 3896:5344  ← FAST RETRANSMIT!
0.120    Receiver → Sender    [.]      ack 6792       ← ACKs everything

Key: 3 duplicate ACKs → fast retransmit (no waiting for RTO timeout)
  Much faster than timeout retransmit (milliseconds vs 200ms+)
```

### Diagnosis

```
Duplicate ACKs mean:
  - Some segments arrived OK (specifically the ones AFTER the gap)
  - One segment in the middle was lost or delayed
  - Receiver is saying "I still need byte 3896!"

Fast retransmit is GOOD:
  - Recovery happens in ~1 RTT instead of ~200ms+
  - TCP's fast recovery algorithm keeps the pipe full
  - SACK makes this even better (retransmit only what's missing)

If you see lots of duplicate ACKs:
  → There IS packet loss on the path
  → But TCP is handling it efficiently
  → Application may still see slight latency
```

---

## Pattern: Zero Window

### What it looks like

```
Time     Source → Dest        Flags    Win    Notes
─────    ──────────────────   ─────    ───    ─────────────
0.100    Server → Client      [P.]     32768  Data (response)
0.102    Client → Server      [.]      16384  ACK (window shrinking)
0.110    Server → Client      [P.]     32768  More data
0.112    Client → Server      [.]      4096   ACK (window still shrinking!)
0.120    Server → Client      [P.]     32768  More data
0.122    Client → Server      [.]      0      ← ZERO WINDOW!
                                               Server MUST stop sending

...pause...

0.500    Client → Server      [.]      16384  ← Window Update
                                               Server can resume
0.501    Server → Client      [P.]     32768  Resume sending
```

### Root cause: Application not reading fast enough

```
The RECEIVER's application is not calling recv() fast enough:

  Network → Kernel receive buffer → Application
  
  If application is SLOW:
    Kernel buffer fills up
    TCP advertises smaller and smaller window
    Eventually: window = 0 → "STOP SENDING"

Common causes:
  1. Application doing CPU-heavy processing before reading next chunk
  2. Application blocked on database/external call
  3. Application GC pause (Java, Go)
  4. Single-threaded application bottleneck
  5. Disk I/O blocking the read loop

Fix: Make the application read faster
  - Read in a separate thread from processing
  - Increase socket buffer size (temporary relief)
  - Profile the application to find the bottleneck
  - Check for GC pauses, lock contention, etc.
  
How to confirm:
  ss -tnm | grep <connection>
  # Check Recv-Q: if large, application is slow
```

---

## Pattern: RST (Reset)

### RST variations and meanings

```
1. RST in response to SYN → Port not open (connection refused)

  Time     Source → Dest        Flags    Notes
  0.000    Client → Server      [S]      SYN to port 8080
  0.001    Server → Client      [R.]     RST-ACK ← nothing listening on 8080

  Cause: No process listening on that port
  Fix: Start the service, check binding address


2. RST during established connection → Abrupt close

  Time     Source → Dest        Flags    Notes
  0.100    Client → Server      [P.]     Data
  0.105    Server → Client      [R]      RST ← connection killed

  Causes:
    - Application crashed
    - Application called close() with SO_LINGER = 0
    - Firewall (conntrack timeout, then RST on next packet)
    - Load balancer health check failure → RST to backend
    - Server received invalid data


3. RST after timeout → Stale connection detected

  Time     Source → Dest        Flags    Notes
  0.000    Client → Server      [.]      Keepalive probe
  0.001    Server → Client      [R]      RST ← "I don't know this connection"

  Cause: Server rebooted, or conntrack entry expired
  The server has no memory of this connection


4. RST with no prior connection → Unexpected/spoofed 

  Time     Source → Dest        Flags    Notes
  0.000    ???.???.???.??? → Me  [R]      RST out of nowhere

  Cause: Response to scan, or port unreachable for non-TCP protocols
```

### Diagnosing RSTs

```bash
# Count RSTs in capture
tshark -r capture.pcap -Y 'tcp.flags.reset == 1' | wc -l

# RSTs by source
tshark -r capture.pcap -Y 'tcp.flags.reset == 1' -T fields -e ip.src | sort | uniq -c | sort -rn

# RST with context (show preceding packets)
# In Wireshark: right-click RST → Follow → TCP Stream
# See what happened RIGHT BEFORE the reset
```

---

## Pattern: Connection Timeout

### SYN timeout (can't establish connection)

```
Time     Source → Dest        Flags    Notes
─────    ──────────────────   ─────    ─────────────
0.000    Client → Server      [S]      SYN
1.000    Client → Server      [S]      SYN retransmit (1s)
3.000    Client → Server      [S]      SYN retransmit (2s)
7.000    Client → Server      [S]      SYN retransmit (4s)
15.000   Client → Server      [S]      SYN retransmit (8s)
31.000   Client → Server      [S]      SYN retransmit (16s)
         ← Application reports: "Connection timed out" (~30-130 seconds)

Key: SYN sent but NO SYN-ACK received
  
Causes:
  - Firewall dropping SYN (silent drop, no RST)
  - Server not listening (but no RST = firewall in front)
  - Network unreachable (but no ICMP unreachable = filtered)
  - Server's SYN queue full (overloaded)
  - Routing issue (packets going to wrong place)

Diagnosis:
  1. Capture on SERVER side — does SYN arrive?
  2. If SYN arrives but no SYN-ACK → server issue (listen, backlog)
  3. If SYN doesn't arrive → routing or firewall issue
```

### Data timeout (connection established but data lost)

```
Time     Source → Dest        Flags    Notes
─────    ──────────────────   ─────    ─────────────
0.000    Handshake                      (OK, fast)
0.100    Client → Server      [P.]     Request sent
0.300    Client → Server      [P.]     Retransmit (200ms)
0.700    Client → Server      [P.]     Retransmit (400ms)
1.500    Client → Server      [P.]     Retransmit (800ms)
...

Server never ACKs the data.
Connection eventually times out.

Possible causes:
  - Server application is hung (received data but can't process)
  - Firewall closing connection silently (idle timeout)
  - NAT table entry expired (return path broken)
  - Server crashed after handshake
```

---

## Pattern: Half-Open / Half-Closed Connections

### Half-open (one side doesn't know)

```
Scenario: Server crashes, client doesn't know.

Before crash:
  Client ←→ Server: ESTABLISHED

After server crash:
  Client thinks: ESTABLISHED (still has state)
  Server: no state (crashed, restarted)

Next time client sends data:
  Client → Server: [P.] data
  Server → Client: [R]  RST (server: "what connection?")

This is why TCP keepalive exists:
  - Periodically probe idle connections
  - Detect dead peers
  - Default: 2 hours (too long for most apps!)
  
Application-level keepalive is better:
  - HTTP/2 PING frames
  - gRPC keepalive
  - WebSocket ping/pong
```

### Half-closed (one side sent FIN)

```
Time     Source → Dest        Flags    Notes
─────    ──────────────────   ─────    ─────────────
0.100    Client → Server      [F.]     Client is done sending
0.101    Server → Client      [.]      ACK
                                        Server can STILL send data
0.200    Server → Client      [P.]     Server sends more data
0.201    Client → Server      [.]      Client ACKs (can still receive)
0.300    Server → Client      [F.]     Server is done too
0.301    Client → Server      [.]      ACK → connection fully closed

Half-closed is NORMAL and useful:
  - Client sends request, FINs (done sending)
  - Server sends response, then FINs
  
  Used by: HTTP/1.0, some batch protocols
  
Problem: If server sends FIN but client doesn't:
  → Server enters FIN_WAIT_2 (waiting for client's FIN)
  → Client is in CLOSE_WAIT (should close but isn't)
  → CLOSE_WAIT accumulation = application bug (not calling close())
```

---

## Pattern: Small Packets (Nagle Problem)

### What Nagle's algorithm does

```
Without Nagle:
  Application writes 10 bytes → immediately sent as 1 segment
  Application writes 10 bytes → immediately sent as 1 segment
  Application writes 10 bytes → immediately sent as 1 segment
  → 3 packets with 10 bytes each (54 bytes overhead each)

With Nagle (default):
  Application writes 10 bytes → sent immediately (first write)
  Application writes 10 bytes → buffered (waiting for ACK of first)
  Application writes 10 bytes → buffered (still waiting)
  ACK arrives → send buffered 20 bytes as one segment
  → 2 packets instead of 3 (more efficient)
```

### When Nagle is a problem

```
Nagle + Delayed ACKs = 40-200ms latency spike

  Client writes 100 bytes (small request)
  Server receives, processes, writes small response
  
  Server's Nagle: "I wrote 50 bytes but there's unACKed data, buffer it"
  Client's delayed ACK: "I'll wait 40ms before ACKing to batch"
  
  Result: 40ms+ delay waiting for either:
    - Client's delayed ACK timer fires (40ms)
    - Server's buffer fills enough to send

Fix: TCP_NODELAY on both sides
  setsockopt(fd, IPPROTO_TCP, TCP_NODELAY, &one, sizeof(one));
  
ALWAYS enable TCP_NODELAY for:
  - Request-response protocols (gRPC, HTTP APIs)
  - Interactive applications (SSH, terminals)
  - Real-time systems (games, trading)
  - Any low-latency requirement
```

### What it looks like in tcpdump

```
Time     Source → Dest        Flags    Len    Notes
─────    ──────────────────   ─────    ───    ─────────
0.000    Client → Server      [P.]     50     Small request
0.040    Server → Client      [.]      0      ← Delayed ACK (40ms!)
0.041    Server → Client      [P.]     50     Response (sent after ACK freed Nagle)

vs with TCP_NODELAY:
0.000    Client → Server      [P.]     50     Small request
0.001    Server → Client      [P.]     50     Immediate response
0.001    Client → Server      [.]      0      ACK

40ms difference per request-response! At 1000 req/sec = catastrophic
```

---

## Pattern: MTU / Fragmentation Issues

### Symptoms

```
1. Small packets work, large ones fail:
   ping -s 56 works (84 bytes total)
   ping -s 1472 works (1500 bytes total)
   ping -s 1473 fails (needs fragmentation)

2. TCP connection establishes but data transfer fails:
   Handshake OK (small packets: SYN/ACK = 40-60 bytes)
   Data transfer stalls (large packets: 1500 bytes)
   
   This happens when:
   - DF (Don't Fragment) bit is set
   - Path has MTU < 1500 somewhere
   - ICMP "Fragmentation Needed" is blocked by firewall
   
   Called: PMTUD Black Hole

3. In tcpdump:
   Client → Server: [P.] seq 1:1449, length 1448      ← sent
   (no ACK, no ICMP error — just silence)
   Client → Server: [P.] seq 1:1449, length 1448      ← retransmit
   (still nothing)
```

### Diagnosis

```bash
# Test Path MTU
ping -s 1472 -M do <destination>
# If fails: "Frag needed and DF set (mtu = 1400)"
# Try smaller sizes until it works

# Binary search for MTU
ping -s 1400 -M do -c 1 <dest>  # works?
ping -s 1450 -M do -c 1 <dest>  # works?
ping -s 1472 -M do -c 1 <dest>  # fails?
# MTU is between 1450+28=1478 and 1472+28=1500

# Check interface MTU
ip link show eth0 | grep mtu

# Common MTU values:
# 1500:  Standard Ethernet
# 1480:  PPPoE (DSL)
# 1400:  VPN tunnels (WireGuard, IPsec)
# 1360:  Double encapsulation
# 9000:  Jumbo frames (datacenter)

# TCP MSS (Maximum Segment Size) should = MTU - 40
# MSS 1460 = MTU 1500 (standard)
# MSS 1360 = MTU 1400 (tunnel)

# Check MSS in SYN packets:
tcpdump -r capture.pcap -nn 'tcp[13] == 2' -v | grep mss
# options [mss 1460,...]   ← expected for standard ethernet
# options [mss 1360,...]   ← tunnel or VPN in path
```

---

## Pattern: Connection Storms

### SYN flood

```
Massive number of SYN packets from many source IPs:

0.000  203.0.113.1.12345 > 10.0.0.1.80: [S]
0.000  198.51.100.2.23456 > 10.0.0.1.80: [S]
0.000  192.0.2.3.34567 > 10.0.0.1.80: [S]
0.001  203.0.113.4.45678 > 10.0.0.1.80: [S]
... thousands per second, no handshake completion

Detection:
  - Many SYN, few SYN-ACK, very few ACK
  - SYN/ACK ratio >> 3:1 (should be ~1:1)
  - Many different source IPs (spoofed)

Check SYN cookies:
  nstat -az | grep SyncookiesSent
  # SyncookiesSent > 0 means kernel activated SYN cookies (defense)
```

### Reconnection storm

```
Application keeps reconnecting rapidly:

0.000  connect → establish → RST
0.010  connect → establish → RST
0.020  connect → establish → RST
...

Cause: 
  - Backend failing, client retrying immediately
  - Misconfigured health check + retry logic
  - Database connection pool exhaustion → reconnect loop

Evidence:
  - Short-lived connections (< 1 second each)
  - High rate of SYN/FIN/RST
  - Connection count oscillating

Fix: Implement exponential backoff, circuit breaker
```

---

## Putting It All Together — Case Studies

### Case 1: API calls taking 200ms longer than expected

```
Capture shows:
  - Handshake: 5ms (normal)
  - Request sent: immediately after handshake
  - 200ms gap before response
  - No retransmissions

Analysis:
  - Not a network issue (no retransmissions, low RTT)
  - 200ms gap = server processing time
  
Further investigation:
  - Check server-side logs for slow queries
  - Profile server application
  
Verdict: Application-level slowness, not network
```

### Case 2: Intermittent 5-second delays

```
Capture shows:
  - Some requests: 50ms response (normal)
  - Some requests: 5000ms response (intermittent)
  - Failed requests show: data sent → 200ms retransmit → success

Analysis:
  - 200ms retransmit = packet loss + recovery
  - But 5000ms total? More investigation needed
  
  Looking deeper: DNS queries before some requests
  - Failing DNS: query sent, no response, 5s timeout, retry to backup DNS
  
Verdict: Intermittent DNS server failure / unreachability
Fix: Reduce DNS timeout, add local DNS cache
```

### Case 3: Connections working then suddenly dying

```
Capture shows:
  - Connection works fine for 5 minutes
  - At exactly 5:00, RST from firewall
  - Pattern repeats on all long-lived connections

Analysis:
  - Clockwork timing = timeout (not random failure)
  - RST source IP = NAT gateway / firewall
  - 5-minute idle timeout on firewall/NAT
  
Verdict: Firewall idle connection timeout = 300 seconds
Fix: Enable TCP keepalive with interval < 300s
     Or application-level keepalive (HTTP/2 PING)
```

---

## Analysis Checklist

```
□ HANDSHAKE
  □ SYN → SYN-ACK → ACK timing (should be ~1 RTT)
  □ MSS negotiated (should match MTU - 40)
  □ Window scaling enabled
  □ SACK enabled
  □ No SYN retransmissions

□ DATA TRANSFER
  □ No retransmissions (or minimal)
  □ No zero windows
  □ Window sizes adequate
  □ No out-of-order segments
  □ Sequence numbers always advancing

□ CLOSING
  □ Clean FIN exchange (not RST)
  □ No unexpected RSTs
  □ No CLOSE_WAIT accumulation on server

□ TIMING
  □ RTT reasonable for the distance
  □ ACKs arriving within expected time
  □ No unexplained delays
  □ Consistent timing (low jitter)

□ APPLICATION LAYER
  □ Valid request/response exchange
  □ Expected HTTP status codes
  □ TLS handshake succeeds
  □ DNS resolution succeeds and is fast
```

---

## Key Takeaways

1. **Know normal** — you can't spot anomalies without understanding what healthy TCP looks like
2. **Same sequence number twice = retransmission** — the most common TCP problem, indicates packet loss
3. **3 duplicate ACKs → fast retransmit** — faster recovery than timeout, means TCP is working as designed
4. **Zero window = slow receiver** — application isn't calling `recv()` fast enough. Check the application, not the network
5. **RST in response to SYN = port closed** — RST during connection = crash/firewall. RST after idle = timeout
6. **SYN retransmissions = can't connect** — firewall silently dropping, or server overloaded
7. **40ms delay between small packets = Nagle + delayed ACK** — fix with TCP_NODELAY
8. **Data works for small packets, fails for large = MTU issue** — PMTUD black hole, test with ping -s
9. **200ms extra latency = one retransmission** — each retry doubles the wait. 6 retries = timeout
10. **Always capture on BOTH sides** — one-sided capture can't distinguish "packet lost" from "response lost"

---

## Next Module

→ [Module 15: Wireless and Mobile Networking](../15-wireless-mobile/01-wireless-and-mobile.md) — How Wi-Fi and cellular networks work
