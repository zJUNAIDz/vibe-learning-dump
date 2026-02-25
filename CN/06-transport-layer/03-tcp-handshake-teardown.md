# TCP Handshake & Teardown — The Connection Lifecycle

> TCP is a connection-oriented protocol. Before any data flows, both sides must agree to communicate (handshake). When done, both sides must agree to stop (teardown). Understanding this dance — including every state transition — is essential for debugging real-world TCP issues.

---

## Table of Contents

1. [Why Connections Matter](#why-connections)
2. [TCP Header Flags](#tcp-flags)
3. [The 3-Way Handshake](#3-way-handshake)
4. [Data Transfer](#data-transfer)
5. [Connection Teardown (4-Way)](#4-way-teardown)
6. [The TCP State Machine](#tcp-state-machine)
7. [TIME_WAIT: The Controversial State](#time-wait)
8. [Simultaneous Open and Close](#simultaneous)
9. [Half-Open and Half-Closed Connections](#half-connections)
10. [RST: The Emergency Stop](#rst)
11. [SYN Flood Attacks](#syn-flood)
12. [Linux: Observing TCP States](#linux-tcp-states)
13. [Key Takeaways](#key-takeaways)

---

## Why Connections

UDP sends packets into the void and hopes for the best. TCP needs a conversation:

1. **Agree on initial parameters**: Sequence numbers, window sizes, MSS, options
2. **Confirm reachability**: Both sides can send AND receive
3. **Allocate resources**: Buffers, timers, state tracking

Without a handshake, the server wouldn't know a client wants to talk. Without teardown, resources would leak forever.

---

## TCP Flags

TCP has 6 original flags (plus 3 newer ones) in the header:

| Flag | Name | Purpose |
|------|------|---------|
| **SYN** | Synchronize | Initiates connection, synchronizes sequence numbers |
| **ACK** | Acknowledge | Confirms receipt of data/signals |
| **FIN** | Finish | Initiates graceful connection close |
| **RST** | Reset | Abruptly terminates connection |
| **PSH** | Push | Deliver data to application immediately (don't buffer) |
| **URG** | Urgent | Out-of-band data (rarely used today) |
| CWR | Congestion Window Reduced | ECN-related |
| ECE | ECN-Echo | ECN-related |
| NS | Nonce Sum | ECN protection |

Flags can be combined. A `SYN-ACK` packet has both the SYN and ACK flags set.

---

## 3-Way Handshake

```
    Client                                  Server
      |                                       |
      |  1. SYN (seq=x)                      |
      |-------------------------------------->|
      |                                       |
      |  2. SYN-ACK (seq=y, ack=x+1)         |
      |<--------------------------------------|
      |                                       |
      |  3. ACK (seq=x+1, ack=y+1)           |
      |-------------------------------------->|
      |                                       |
      |  <-- Connection ESTABLISHED -->       |
      |                                       |
```

### Step by step

**Step 1: Client sends SYN**
```
Client → Server:
  SYN flag = 1
  Sequence Number = x     (random Initial Sequence Number — ISN)
  ACK flag = 0
  
Client state: SYN_SENT
Server state: LISTEN
```

The client picks a random ISN (not 0, not sequential — randomized for security). This tells the server: "I want to talk. My byte numbering starts at x."

**Step 2: Server sends SYN-ACK**
```
Server → Client:
  SYN flag = 1
  ACK flag = 1
  Sequence Number = y     (server's random ISN)
  Acknowledgment Number = x+1   ("I received your SYN, expecting byte x+1 next")
  
Client state: SYN_SENT
Server state: SYN_RCVD
```

The server acknowledges the client's SYN (by ACKing x+1) AND sends its own SYN (sequence y). Two things in one packet.

**Step 3: Client sends ACK**
```
Client → Server:
  SYN flag = 0
  ACK flag = 1
  Sequence Number = x+1
  Acknowledgment Number = y+1   ("I received your SYN, expecting byte y+1 next")
  
Client state: ESTABLISHED
Server state: ESTABLISHED (upon receiving this ACK)
```

### Why 3 packets? Why not 2?

Two packets would be enough for the server to know the client can send. But the client wouldn't know the server can send. The third packet (ACK) confirms bidirectional communication:

```
After 1 packet:  Server knows: Client can send ✓
After 2 packets: Client knows: Server can send ✓, Server can receive ✓
After 3 packets: Server knows: Client can receive ✓

All four directional capabilities confirmed in exactly 3 packets.
```

### The cost of the handshake

The handshake takes **1.5 RTTs** (round-trip times) before data can flow:

```
Time 0:         Client sends SYN
Time 0.5 RTT:   Server receives SYN, sends SYN-ACK
Time 1.0 RTT:   Client receives SYN-ACK, sends ACK + optional data
Time 1.5 RTT:   Server receives ACK, connection fully established

For a 100ms RTT link: 150ms just to establish the connection
```

### Options negotiated during handshake

Both sides exchange capabilities in the SYN and SYN-ACK:

| Option | Description |
|--------|-------------|
| MSS (Maximum Segment Size) | Largest data chunk (typically 1460 bytes for ethernet) |
| Window Scale | Multiplier for receive window (allows windows > 65535 bytes) |
| SACK Permitted | Whether Selective Acknowledgment is supported |
| Timestamps | For RTT measurement and PAWS (Protection Against Wrapped Sequences) |

```
SYN options example:
  MSS: 1460
  Window Scale: 7 (multiply window by 2^7 = 128)
  SACK Permitted: Yes
  Timestamps: TSval=12345, TSecr=0
```

---

## Data Transfer

After the handshake, data flows using sequence numbers and acknowledgments:

```
Client                              Server
  |                                   |
  |  seq=1001, len=500 ("Hello...")   |
  |---------------------------------->|
  |                                   |
  |  ack=1501 ("Got it, send 1501")   |
  |<----------------------------------|
  |                                   |
  |  seq=1501, len=300 ("More...")    |
  |---------------------------------->|
  |                                   |
```

Key rules:
- **Sequence number** = byte position of the first byte in this segment
- **ACK number** = "I've received all bytes up to this number, send this byte next"
- **ACKs are cumulative**: ACK 1501 means "I have all bytes through 1500"

---

## 4-Way Teardown

Either side can initiate connection close. The closing process has 4 steps because each direction is closed independently.

```
    Client                                  Server
      |                                       |
      |  1. FIN (seq=u)                       |
      |-------------------------------------->|
      |                                       |
      |  2. ACK (ack=u+1)                    |
      |<--------------------------------------|
      |                                       |
      |  (Server may still send data)         |
      |                                       |
      |  3. FIN (seq=v)                       |
      |<--------------------------------------|
      |                                       |
      |  4. ACK (ack=v+1)                    |
      |-------------------------------------->|
      |                                       |
```

### Step by step

**Step 1: Client sends FIN**
- "I'm done sending data." Client enters `FIN_WAIT_1`.

**Step 2: Server sends ACK**
- "I acknowledge your FIN." Client enters `FIN_WAIT_2`. Server enters `CLOSE_WAIT`.
- Server may still send data — only the client-to-server direction is closed.

**Step 3: Server sends FIN**
- "I'm also done sending data." Server enters `LAST_ACK`.

**Step 4: Client sends ACK**
- "I acknowledge your FIN." Client enters `TIME_WAIT`. Server enters `CLOSED`.
- Client stays in TIME_WAIT for 2×MSL (Maximum Segment Lifetime) before closing.

### Why 4 packets, not 3?

Because TCP is **full-duplex** — data flows in both directions independently. Closing one direction doesn't close the other. The server might have more data to send after acknowledging the client's FIN.

In practice, the server often sends its FIN immediately with the ACK (3 packets total). The kernel does this automatically when the application closes the socket immediately.

---

## The TCP State Machine

This is the complete TCP state machine. **Study this diagram carefully** — it explains every TCP behavior.

```
                              ┌──────────┐
                              │  CLOSED  │
                              └────┬─────┘
                   passive open /  │  \ active open
                   create TCB     │   \ send SYN
                              ┌───▼──┐  ┌────────┐
                              │LISTEN│  │SYN_SENT│
                              └───┬──┘  └────┬───┘
                   rcv SYN /      │          │  rcv SYN-ACK /
                   send SYN-ACK   │          │  send ACK
                              ┌───▼──────────▼───┐
                              │     SYN_RCVD     │
                              └────────┬─────────┘
                                       │ rcv ACK
                              ┌────────▼─────────┐
                              │   ESTABLISHED    │◄────── Data transfer
                              └────────┬─────────┘       happens here
                        close /        │
                        send FIN       │ rcv FIN /
                              ┌────────▼───┐ send ACK
                              │ FIN_WAIT_1 │─────────►┌────────────┐
                              └────────┬───┘          │ CLOSE_WAIT │
                 rcv ACK /             │              └─────┬──────┘
                                ┌──────▼───┐   close /      │
                                │FIN_WAIT_2│   send FIN     │
                                └──────┬───┘         ┌──────▼──┐
                        rcv FIN /      │             │LAST_ACK │
                        send ACK       │             └──────┬──┘
                              ┌────────▼───┐  rcv ACK /     │
                              │ TIME_WAIT  │         ┌──────▼──┐
                              └────────┬───┘         │ CLOSED  │
                       timeout 2MSL /  │             └─────────┘
                              ┌────────▼──┐
                              │  CLOSED   │
                              └───────────┘
```

### State descriptions

| State | Who | Meaning |
|-------|-----|---------|
| **CLOSED** | Both | No connection exists |
| **LISTEN** | Server | Waiting for incoming SYN |
| **SYN_SENT** | Client | SYN sent, waiting for SYN-ACK |
| **SYN_RCVD** | Server | SYN received, SYN-ACK sent, waiting for ACK |
| **ESTABLISHED** | Both | Connection open, data can flow |
| **FIN_WAIT_1** | Closer | FIN sent, waiting for ACK |
| **FIN_WAIT_2** | Closer | FIN ACKed, waiting for peer's FIN |
| **CLOSE_WAIT** | Receiver | Peer sent FIN, waiting for application to close |
| **LAST_ACK** | Receiver | FIN sent, waiting for final ACK |
| **TIME_WAIT** | Closer | Both FINs exchanged, waiting 2×MSL before full close |
| **CLOSING** | Both | Rare: Both sides sent FIN simultaneously |

---

## TIME_WAIT

The most misunderstood TCP state. After the final ACK, the connection enters TIME_WAIT and stays there for **2×MSL** (Maximum Segment Lifetime, typically 60 seconds on Linux, so TIME_WAIT = 60 seconds total).

### Why TIME_WAIT exists

**Reason 1: Ensure the final ACK arrives**

If the final ACK is lost, the peer will retransmit its FIN. The TIME_WAIT state ensures the socket still exists to retransmit the ACK.

```
Without TIME_WAIT:
  Client sends final ACK → immediately closes
  ACK is lost
  Server retransmits FIN → Client responds with RST (no such connection)
  Server gets RST → abnormal termination → error in server logs

With TIME_WAIT:
  Client sends final ACK → enters TIME_WAIT (60s)
  ACK is lost
  Server retransmits FIN → Client (still in TIME_WAIT) resends ACK
  Server receives ACK → clean close
```

**Reason 2: Prevent old segments from being accepted by new connections**

If a new connection immediately reuses the same 5-tuple, delayed packets from the old connection could be accepted by the new one.

```
Old connection: 10.0.0.1:5000 ↔ 10.0.0.2:80 (closed)
Delayed packet from old connection still in transit...
New connection: 10.0.0.1:5000 ↔ 10.0.0.2:80 (reusing same ports)
Delayed packet arrives → new connection accepts it → DATA CORRUPTION
```

TIME_WAIT prevents this by blocking the 5-tuple from being reused for 2×MSL.

### TIME_WAIT problems in production

On busy servers, thousands of connections in TIME_WAIT can exhaust:
- **Port numbers** (especially on proxies/load balancers)
- **Memory** (each TIME_WAIT socket uses kernel memory)

```bash
# Count TIME_WAIT sockets
ss -s
# TCP:   1200 (estab 45, closed 8, orphaned 0, timewait 1147)

# Detailed TIME_WAIT view
ss -tn state time-wait | wc -l
ss -tn state time-wait | head -20
```

### Mitigation

```bash
# Allow TIME_WAIT sockets to be reused for new connections
# to the same destination (safe to enable)
sudo sysctl -w net.ipv4.tcp_tw_reuse=1

# Reduce FIN timeout (reduces TIME_WAIT duration)
sudo sysctl -w net.ipv4.tcp_fin_timeout=30   # default 60

# DO NOT use tcp_tw_recycle — removed from Linux kernel (was dangerous with NAT)
```

---

## Simultaneous

Rarely, both sides send SYN simultaneously (simultaneous open) or both send FIN simultaneously (simultaneous close).

### Simultaneous open

```
Client                          Server
  |                               |
  |--- SYN (seq=x) ------>       |
  |       <------ SYN (seq=y) ---|
  |                               |
  |--- SYN-ACK (seq=x, ack=y+1)->|
  |<- SYN-ACK (seq=y, ack=x+1)--|
  |                               |
  Both enter SYN_RCVD, then ESTABLISHED
```

Both sides go through SYN_SENT → SYN_RCVD → ESTABLISHED. This is valid TCP but extremely rare.

### Simultaneous close

```
Client                          Server
  |--- FIN --->   <--- FIN ---|
  |                             |
  Both enter CLOSING state
  Both send ACK
  Both enter TIME_WAIT
```

Both sides enter FIN_WAIT_1 → CLOSING → TIME_WAIT → CLOSED.

---

## Half-Open and Half-Closed Connections

### Half-open (broken)

A connection where one side thinks it's established, but the other side has no knowledge of it.

**How it happens**:
- Server crashes without sending FIN (power failure, kernel panic)
- Client still thinks the connection is alive
- Client sends data → no response → eventually times out

```bash
# Detect with keepalive probes
# Linux default: keepalive after 7200s (2 hours!) of idle
cat /proc/sys/net/ipv4/tcp_keepalive_time
# 7200

# Make keepalive more aggressive
sudo sysctl -w net.ipv4.tcp_keepalive_time=60     # probe after 60s idle
sudo sysctl -w net.ipv4.tcp_keepalive_intvl=10    # probe every 10s
sudo sysctl -w net.ipv4.tcp_keepalive_probes=6    # give up after 6 failures
```

### Half-closed (intentional)

A connection where one direction is closed (FIN sent) but the other remains open. This is a feature, not a bug.

```
Client sends FIN → "I'm done sending"
Server ACKs FIN
Server continues sending data...
Server sends FIN → "Now I'm done too"
Client ACKs FIN
```

The HTTP/1.0 pattern uses this: the client sends a request, then closes its write side. The server sends the full response, then closes.

```bash
# shutdown() vs close() in C
shutdown(fd, SHUT_WR);   # half-close: stop writing, still reading
close(fd);               # full close: stop both directions
```

---

## RST: The Emergency Stop

RST (Reset) immediately tears down a connection without the graceful FIN dance.

### When RST is sent

1. **Connection to closed port**: Server has nothing listening on that port
2. **Aborting a connection**: Application specifically requests abort
3. **Firewall/middlebox interference**: Device in the path rejects the connection
4. **Half-open detection**: Data arrives for a connection the receiver doesn't know about

```
Normal close:  FIN → ACK → FIN → ACK  (graceful, 4 packets)
RST close:     RST                     (immediate, 1 packet, data may be lost)
```

### Detecting RST in practice

```bash
# Capture RST packets
sudo tcpdump -nn 'tcp[tcpflags] & tcp-rst != 0'

# Common causes when you see RST:
# 1. Firewall dropping connections (iptables -j REJECT sends RST)
# 2. Application crash (kernel sends RST for orphaned connections)
# 3. Load balancer timeout (LB sends RST after idle timeout)
# 4. Port not open (no process listening)
```

---

## SYN Flood

A SYN flood sends thousands of SYN packets without completing the handshake.

### How it works

```
Attacker → Server: SYN (from spoofed IP 1.1.1.1)
Attacker → Server: SYN (from spoofed IP 2.2.2.2)
Attacker → Server: SYN (from spoofed IP 3.3.3.3)
... thousands more

Server allocates memory for each SYN_RCVD state,
sends SYN-ACK to spoofed IPs (no response),
fills up its SYN backlog queue,
can't accept legitimate connections.
```

### Defense: SYN Cookies

Instead of allocating state on SYN, the server encodes the connection parameters into the SYN-ACK sequence number itself. If the client completes the handshake with the correct ACK, the server can reconstruct the state.

```bash
# Check if SYN cookies are enabled
cat /proc/sys/net/ipv4/tcp_syncookies
# 1 = enabled (default on most Linux systems)

# Enable SYN cookies
sudo sysctl -w net.ipv4.tcp_syncookies=1

# Tune SYN backlog (queue for half-open connections)
cat /proc/sys/net/ipv4/tcp_max_syn_backlog
# 1024 (default)
sudo sysctl -w net.ipv4.tcp_max_syn_backlog=4096
```

### Monitoring SYN floods

```bash
# Watch SYN_RCVD connections
watch -n 1 'ss -tn state syn-recv | wc -l'

# If this number is consistently near tcp_max_syn_backlog, you're under attack

# Check for SYN cookie activations in dmesg
dmesg | grep "SYN flooding"
# "TCP: request_sock_TCP: Possible SYN flooding on port 80. Sending cookies."
```

---

## Linux: Observing TCP States

### View all TCP states at once

```bash
# Summary of all TCP states
ss -s
# TCP:   234 (estab 45, closed 12, orphaned 3, timewait 174)

# Count by state
ss -tan | awk '{print $1}' | sort | uniq -c | sort -rn
#  156 TIME-WAIT
#   45 ESTAB
#   12 LISTEN
#    8 FIN-WAIT-2
#    5 CLOSE-WAIT
#    3 SYN-SENT
#    2 SYN-RECV
```

### Filter by specific state

```bash
# All ESTABLISHED connections
ss -tn state established

# All TIME_WAIT connections
ss -tn state time-wait

# All CLOSE_WAIT connections (potential socket leak in your application!)
ss -tn state close-wait
# If these keep growing, your app isn't calling close() on sockets

# All LISTEN sockets
ss -tln

# All non-LISTEN, non-TIME_WAIT (active connections)
ss -tn state connected
```

### Watching a handshake in real time

```bash
# Terminal 1: capture the handshake
sudo tcpdump -nn -c 10 port 8080

# Terminal 2: start a server
nc -l 8080

# Terminal 3: connect
nc localhost 8080

# tcpdump output:
# 10:00:00.000 IP 127.0.0.1.45678 > 127.0.0.1.8080: Flags [S], seq 123456789
# 10:00:00.000 IP 127.0.0.1.8080 > 127.0.0.1.45678: Flags [S.], seq 987654321, ack 123456790
# 10:00:00.000 IP 127.0.0.1.45678 > 127.0.0.1.8080: Flags [.], ack 987654322
#
# [S]  = SYN
# [S.] = SYN-ACK (. = ACK flag)
# [.]  = ACK
# [F]  = FIN
# [R]  = RST
# [P.] = PSH-ACK (push data, acknowledge)
```

### Key TCP sysctl parameters

```bash
# Show all TCP parameters
sysctl net.ipv4.tcp | head -40

# Critical tuning parameters:
cat /proc/sys/net/ipv4/tcp_fin_timeout          # TIME_WAIT duration (default: 60)
cat /proc/sys/net/ipv4/tcp_tw_reuse             # Reuse TIME_WAIT sockets (0 or 1)
cat /proc/sys/net/ipv4/tcp_syncookies           # SYN flood protection (default: 1)
cat /proc/sys/net/ipv4/tcp_max_syn_backlog      # SYN queue size (default: 1024)
cat /proc/sys/net/core/somaxconn                # Listen backlog (default: 4096)
cat /proc/sys/net/ipv4/tcp_keepalive_time       # Keepalive idle time (default: 7200)
cat /proc/sys/net/ipv4/tcp_keepalive_intvl      # Keepalive probe interval (default: 75)
cat /proc/sys/net/ipv4/tcp_keepalive_probes     # Keepalive probe count (default: 9)
```

---

## Key Takeaways

1. **3-way handshake** (SYN → SYN-ACK → ACK) establishes a connection in 1.5 RTTs
2. **4-way teardown** (FIN → ACK → FIN → ACK) closes each direction independently
3. **TIME_WAIT lasts 2×MSL** — prevents old packets from contaminating new connections
4. **CLOSE_WAIT growing = socket leak** — your application isn't closing connections
5. **RST is an emergency stop** — no graceful cleanup, data may be lost
6. **SYN cookies protect against SYN floods** — encode state in the sequence number
7. **`ss` with state filters** is the primary debugging tool for TCP state issues
8. **Half-closed connections are a feature** — one direction can close while the other stays open

---

## Next

→ [04-tcp-reliability-retransmission.md](04-tcp-reliability-retransmission.md) — How TCP guarantees delivery
