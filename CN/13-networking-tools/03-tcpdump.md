# tcpdump — Packet Capture on the Command Line

> tcpdump is the most essential packet capture tool in the networking world. It lets you see exactly what's on the wire — every SYN, every ACK, every DNS query, every HTTP request. When logs lie and metrics mislead, packets tell the truth. Master tcpdump and you can debug anything.

---

## Table of Contents

1. [Why tcpdump?](#why)
2. [Basic Usage](#basic)
3. [Capture Filters (BPF)](#filters)
4. [Reading tcpdump Output](#reading)
5. [TCP Analysis with tcpdump](#tcp-analysis)
6. [Protocol-Specific Captures](#protocols)
7. [Writing and Reading Capture Files](#files)
8. [Advanced Techniques](#advanced)
9. [Performance and Production Use](#production)
10. [Common Recipes](#recipes)
11. [Key Takeaways](#key-takeaways)

---

## Why tcpdump?

```
When to use tcpdump:
  - "The server says it responded but the client says it didn't"
  - "Is the packet even reaching the server?"
  - "What's the actual TLS version being negotiated?"
  - "Is DNS returning the wrong IP?"
  - "Why is the connection resetting?"
  - "Is there a retransmission problem?"

tcpdump vs alternatives:
  tcpdump:    CLI, fast, scriptable, available everywhere, lightweight
  Wireshark:  GUI, deep protocol analysis, not available on servers
  tshark:     Wireshark CLI (richer parsing than tcpdump, heavier)

Typical workflow:
  1. Capture with tcpdump on server → save to .pcap file
  2. Transfer .pcap to local machine
  3. Analyze in Wireshark (if GUI needed) or tshark
```

---

## Basic Usage

```bash
# Capture all traffic on an interface (requires root)
sudo tcpdump -i eth0

# Capture with more detail
sudo tcpdump -i eth0 -nn -v
# -nn: don't resolve hostnames or port names (MUCH faster)
# -v:  verbose (-vv more verbose, -vvv maximum)

# Capture on all interfaces
sudo tcpdump -i any

# Limit capture count
sudo tcpdump -i eth0 -c 100    # stop after 100 packets

# Capture and save to file
sudo tcpdump -i eth0 -w capture.pcap

# Read from file
tcpdump -r capture.pcap

# Show hex + ASCII dump
sudo tcpdump -i eth0 -XX

# Show only ASCII payload
sudo tcpdump -i eth0 -A
```

### Essential flags

```
-i eth0          Capture on specific interface (-i any for all)
-nn              No name resolution (ALWAYS use this)
-v/-vv/-vvv      Verbose levels
-c N             Stop after N packets
-w file.pcap     Write raw packets to file
-r file.pcap     Read from file
-A               Print packet payload as ASCII
-X               Print hex and ASCII
-XX              Print hex and ASCII including link-layer header
-e               Show link-layer (Ethernet) header
-s N             Snap length (capture N bytes per packet, default 262144)
-S               Print absolute TCP sequence numbers (not relative)
-t               Don't print timestamp
-tttt            Print human-readable timestamp with date
-q               Quiet (less protocol info)
-l               Line-buffered output (for piping to grep)
```

---

## Capture Filters (BPF)

### Filter syntax

Capture filters use Berkeley Packet Filter (BPF) syntax. They run in the kernel — efficient because non-matching packets are never copied to userspace.

```bash
# Filter by host
sudo tcpdump -i eth0 host 10.0.0.5
sudo tcpdump -i eth0 src host 10.0.0.5    # source only
sudo tcpdump -i eth0 dst host 10.0.0.5    # destination only

# Filter by network
sudo tcpdump -i eth0 net 10.0.0.0/24

# Filter by port
sudo tcpdump -i eth0 port 80
sudo tcpdump -i eth0 src port 80
sudo tcpdump -i eth0 dst port 443
sudo tcpdump -i eth0 portrange 8000-9000

# Filter by protocol
sudo tcpdump -i eth0 tcp
sudo tcpdump -i eth0 udp
sudo tcpdump -i eth0 icmp
sudo tcpdump -i eth0 arp

# Combine with and/or/not
sudo tcpdump -i eth0 'host 10.0.0.5 and port 443'
sudo tcpdump -i eth0 'src 10.0.0.5 and dst port 80'
sudo tcpdump -i eth0 'port 80 or port 443'
sudo tcpdump -i eth0 'not port 22'              # exclude SSH
sudo tcpdump -i eth0 'not (port 22 or port 53)' # exclude SSH and DNS

# Filter by TCP flags
sudo tcpdump -i eth0 'tcp[tcpflags] & tcp-syn != 0'      # SYN packets
sudo tcpdump -i eth0 'tcp[tcpflags] & tcp-rst != 0'      # RST packets
sudo tcpdump -i eth0 'tcp[tcpflags] == tcp-syn'           # SYN only (no ACK)
sudo tcpdump -i eth0 'tcp[tcpflags] & (tcp-syn|tcp-fin) != 0'  # SYN or FIN

# Filter by packet size
sudo tcpdump -i eth0 'greater 1000'     # packets > 1000 bytes
sudo tcpdump -i eth0 'less 100'          # packets < 100 bytes
```

---

## Reading tcpdump Output

### TCP packet output

```bash
sudo tcpdump -i eth0 -nn port 80
# 14:23:45.123456 IP 10.0.0.2.54321 > 10.0.0.1.80: Flags [S], seq 1234567890, win 64240, options [mss 1460,sackOK,TS val 1234 ecr 0,nop,wscale 7], length 0
#  │               │  │              │             │        │          │        │                                                                   │
#  │               │  │              │             │        │          │        │                                                                   └── payload
#  │               │  │              │             │        │          │        └── TCP options
#  │               │  │              │             │        │          └── window size
#  │               │  │              │             │        └── sequence number
#  │               │  │              │             └── TCP flags
#  │               │  │              └── destination IP:port
#  │               │  └── source IP:port
#  │               └── IP version
#  └── timestamp
```

### TCP flags

```
Flags field:
  [S]     SYN         (connection initiation)
  [S.]    SYN-ACK     (connection acceptance)
  [.]     ACK         (acknowledgment, or "no flags of interest")
  [P.]    PSH-ACK     (push data immediately)
  [F.]    FIN-ACK     (connection close)
  [R]     RST         (connection reset)
  [R.]    RST-ACK     (reset with acknowledgment)
  [F]     FIN         (close)

Normal TCP connection:
  Client → Server: [S]    SYN
  Server → Client: [S.]   SYN-ACK
  Client → Server: [.]    ACK
  Client → Server: [P.]   data
  Server → Client: [.]    ACK
  Server → Client: [P.]   response data  
  Client → Server: [F.]   FIN (close)
  Server → Client: [F.]   FIN (close)
  Client → Server: [.]    ACK
```

### Three-way handshake in tcpdump

```bash
sudo tcpdump -i eth0 -nn 'host 10.0.0.2 and port 80'

# 14:23:45.001 IP 10.0.0.2.54321 > 10.0.0.1.80: Flags [S], seq 100, win 64240, length 0
# 14:23:45.002 IP 10.0.0.1.80 > 10.0.0.2.54321: Flags [S.], seq 200, ack 101, win 65535, length 0
# 14:23:45.003 IP 10.0.0.2.54321 > 10.0.0.1.80: Flags [.], ack 201, win 502, length 0

# SYN:      seq=100
# SYN-ACK:  seq=200, ack=101 (100+1)
# ACK:      ack=201 (200+1)
# Connection established!
```

---

## TCP Analysis with tcpdump

### Detecting retransmissions

```bash
# Look for same sequence number appearing multiple times
sudo tcpdump -i eth0 -nn 'tcp and port 80' -S  # -S = absolute seq numbers

# 14:23:45.100  10.0.0.2.54321 > 10.0.0.1.80: seq 1000:2448, ack 500, length 1448
# 14:23:45.350  10.0.0.2.54321 > 10.0.0.1.80: seq 1000:2448, ack 500, length 1448  ← RETRANSMIT!
#                                                   ^^^^^^^^
# Same sequence number = retransmission (250ms later = RTO)

# With verbose flag, tcpdump may show [TCP Retransmission] if using newer versions
```

### Detecting RSTs (connection resets)

```bash
# Find RST packets
sudo tcpdump -i eth0 -nn 'tcp[tcpflags] & tcp-rst != 0'

# Common RST causes:
# - Connection to closed port
# - Firewall rejection (RST vs ICMP unreachable)
# - Application crash / ungraceful close
# - Stale connection (conntrack timeout)
# - TCP keepalive failure
```

### Detecting zero window

```bash
# Zero window = receiver buffer full → sender must stop
sudo tcpdump -i eth0 -nn 'tcp and port 80' -v | grep 'win 0'
# ... win 0 ...    ← receiver telling sender to STOP

# Later:
# ... win 502 ...  ← window update → sender can resume
```

### Detecting connection refused

```bash
sudo tcpdump -i eth0 -nn port 8080
# 14:23:45.001 10.0.0.2.54321 > 10.0.0.1.8080: Flags [S], seq 100
# 14:23:45.001 10.0.0.1.8080 > 10.0.0.2.54321: Flags [R.], ack 101
#                                                        ^^
# RST-ACK in response to SYN = port not open (connection refused)
```

---

## Protocol-Specific Captures

### DNS

```bash
# Capture DNS traffic
sudo tcpdump -i eth0 -nn port 53

# Verbose DNS decoding
sudo tcpdump -i eth0 -nn -v port 53

# 14:23:45.001 IP 10.0.0.2.54321 > 8.8.8.8.53: 12345+ A? example.com. (29)
#                                                 │     │  │
#                                                 │     │  └── query domain
#                                                 │     └── query type (A record)
#                                                 └── transaction ID

# 14:23:45.015 IP 8.8.8.8.53 > 10.0.0.2.54321: 12345 1/0/0 A 93.184.216.34 (45)
#                                                │     │       │
#                                                │     │       └── answer
#                                                │     └── 1 answer, 0 authority, 0 additional
#                                                └── matching transaction ID
```

### HTTP (unencrypted)

```bash
# Capture HTTP requests
sudo tcpdump -i eth0 -nn -A 'tcp port 80 and (((ip[2:2] - ((ip[0]&0xf)<<2)) - ((tcp[12]&0xf0)>>2)) != 0)'
# This complex filter matches TCP packets with payload (not just ACKs)

# Simpler: capture ASCII on port 80
sudo tcpdump -i eth0 -nn -A port 80 | grep -E 'GET|POST|HTTP|Host:'
# GET /api/users HTTP/1.1
# Host: api.example.com
# HTTP/1.1 200 OK
```

### TLS

```bash
# Can't see encrypted content, but CAN see:
# - ClientHello (TLS version, cipher suites, SNI)
# - ServerHello (selected cipher)
# - Certificate exchange

# Capture TLS handshakes
sudo tcpdump -i eth0 -nn port 443

# See TLS version in ClientHello (byte offset 9-10 of TLS record)
sudo tcpdump -i eth0 -nn -x port 443 | head -50
```

### ICMP

```bash
# All ICMP
sudo tcpdump -i eth0 -nn icmp

# Specific ICMP types
sudo tcpdump -i eth0 -nn 'icmp[icmptype] == icmp-echo'          # ping requests
sudo tcpdump -i eth0 -nn 'icmp[icmptype] == icmp-echoreply'     # ping replies
sudo tcpdump -i eth0 -nn 'icmp[icmptype] == icmp-unreach'       # unreachable
sudo tcpdump -i eth0 -nn 'icmp[icmptype] == icmp-timxceed'      # traceroute

# Decode ICMP unreachable subtypes
sudo tcpdump -i eth0 -nn -v icmp
# ICMP host 10.0.0.5 unreachable - admin prohibited     ← firewall
# ICMP 10.0.0.5 udp port 53 unreachable                 ← port closed
# ICMP 10.0.0.5 unreachable - need to frag (mtu 1400)   ← MTU issue
```

### ARP

```bash
# Capture ARP
sudo tcpdump -i eth0 -nn arp
# ARP, Request who-has 10.0.0.1 tell 10.0.0.2, length 28
# ARP, Reply 10.0.0.1 is-at 52:54:00:ab:cd:ef, length 28

# Useful for:
# - Detecting ARP storms
# - Finding duplicate IPs (Gratuitous ARP)
# - Detecting ARP spoofing
```

---

## Writing and Reading Capture Files

### Saving captures

```bash
# Save to pcap file (full packet data)
sudo tcpdump -i eth0 -w capture.pcap

# Save with filter
sudo tcpdump -i eth0 -w dns.pcap port 53

# Limit file size
sudo tcpdump -i eth0 -w capture.pcap -c 10000       # 10K packets
sudo tcpdump -i eth0 -w capture.pcap -C 100          # rotate at 100 MB

# Rotate files (ring buffer)
sudo tcpdump -i eth0 -w capture.pcap -C 100 -W 5
# Creates: capture.pcap0, capture.pcap1, ..., capture.pcap4
# Keeps only last 5 files (500 MB total)

# Save with timestamp in filename
sudo tcpdump -i eth0 -w "capture_$(date +%Y%m%d_%H%M%S).pcap"
```

### Reading captures

```bash
# Read pcap file
tcpdump -r capture.pcap

# Apply filter to saved capture
tcpdump -r capture.pcap 'host 10.0.0.5 and port 443'
tcpdump -r capture.pcap -nn -v 'tcp[tcpflags] & tcp-rst != 0'

# Count packets matching filter
tcpdump -r capture.pcap 'port 80' | wc -l

# Extract timestamps and calculate timing
tcpdump -r capture.pcap -nn -ttt port 443 | head
# -ttt = time delta between packets
```

---

## Advanced Techniques

### Capture on remote machine, analyze locally

```bash
# Method 1: Capture, transfer, analyze
ssh server "sudo tcpdump -i eth0 -w - -c 1000 port 443" > remote.pcap
# Then open remote.pcap in Wireshark

# Method 2: Live stream to local Wireshark
ssh server "sudo tcpdump -i eth0 -U -w - port 443" | wireshark -k -i -
# -U = packet-buffered output (flush after each packet)
```

### Following TCP streams

```bash
# See full conversation
tcpdump -r capture.pcap -nn -A 'tcp port 80 and host 10.0.0.2' | less
# Shows request and response interleaved

# Better: use tshark for stream reassembly
tshark -r capture.pcap -q -z follow,tcp,ascii,0
# Stream index 0, fully reassembled
```

### Capturing on specific conditions

```bash
# Only packets > 1000 bytes (find large transfers)
sudo tcpdump -i eth0 -nn 'greater 1000'

# Only SYN packets (new connections)
sudo tcpdump -i eth0 -nn 'tcp[tcpflags] == tcp-syn'

# Only packets with specific TCP options (e.g., MSS)
sudo tcpdump -i eth0 -nn 'tcp[tcpflags] & tcp-syn != 0' -v | grep mss

# Capture VLAN-tagged traffic
sudo tcpdump -i eth0 -nn -e vlan

# Capture traffic in a network namespace
sudo ip netns exec my-ns tcpdump -i eth0 -nn
```

---

## Performance and Production Use

### Minimizing impact

```bash
# ALWAYS use -nn (no DNS lookups — this is critical for performance)
sudo tcpdump -i eth0 -nn

# Limit snap length if you only need headers
sudo tcpdump -i eth0 -nn -s 96   # 96 bytes captures all headers
# Default is 262144 bytes (full packet)

# Use kernel filters (BPF) — filtered in kernel, not userspace
sudo tcpdump -i eth0 -nn port 443  # efficient
# vs
sudo tcpdump -i eth0 -nn | grep 443  # TERRIBLE (captures everything)

# Write to file (faster than terminal output)
sudo tcpdump -i eth0 -nn -w capture.pcap port 443

# Rotate files in production
sudo tcpdump -i eth0 -nn -w /var/log/pcap/capture.pcap -C 100 -W 10 -Z nobody
# -C 100: rotate at 100 MB | -W 10: keep 10 files | -Z nobody: drop privileges
```

### Impact assessment

```
tcpdump performance impact:
  - Light filter, low traffic: negligible
  - No filter, high traffic: can drop packets (misses) and use CPU
  - Writing to disk: I/O intensive on high-traffic interfaces
  - DNS lookups (without -nn): HUGE impact (avoid!)

On production servers:
  ✓ Always use specific BPF filters
  ✓ Always use -nn
  ✓ Write to file, analyze later
  ✓ Set -c (packet count limit) or time limit
  ✓ Use -s 96 if you only need headers
  ✗ Never capture everything without filters
  ✗ Never pipe through grep without -l flag
```

---

## Common Recipes

### Debug DNS issues

```bash
# See all DNS queries and responses
sudo tcpdump -i eth0 -nn port 53 -v
# Check: Is the query being sent? Is a response coming back?
# Check: Is the response correct? NXDOMAIN? SERVFAIL?
```

### Find who's connecting to a port

```bash
# All new connections to port 443
sudo tcpdump -i eth0 -nn 'tcp[tcpflags] == tcp-syn and dst port 443'
# Shows source IPs of all new connections
```

### Debug TLS/SSL issues

```bash
# Capture TLS handshake
sudo tcpdump -i eth0 -nn -w tls_debug.pcap port 443 -c 50
# Open in Wireshark → filter: tls.handshake
# Check: ClientHello (supported versions, ciphers, SNI)
# Check: ServerHello (selected cipher)
# Check: Certificate (correct cert chain?)
```

### Detect retransmissions

```bash
# Capture with absolute sequence numbers
sudo tcpdump -i eth0 -nn -S port 80 -w retrans_check.pcap
# Analyze in Wireshark: filter tcp.analysis.retransmission
# Or with tshark:
tshark -r retrans_check.pcap -Y 'tcp.analysis.retransmission' | wc -l
```

### Monitor bandwidth per host

```bash
# Quick byte count per source IP
sudo tcpdump -i eth0 -nn -q -c 10000 | \
  awk '{print $3}' | cut -d. -f1-4 | sort | uniq -c | sort -rn | head
#   3456 10.0.0.2
#   2345 10.0.0.3
#    789 10.0.0.4
```

---

## Key Takeaways

1. **tcpdump shows truth** — when logs and metrics disagree, packets are the ground truth
2. **ALWAYS use `-nn`** — DNS lookups during capture drastically slow it down and can change behavior
3. **BPF filters run in kernel** — use `tcpdump -i eth0 port 443`, never `tcpdump | grep 443`
4. **Save to file, analyze later** — `tcpdump -w capture.pcap` then `tcpdump -r` or open in Wireshark
5. **TCP flags decode**: `[S]` = SYN, `[S.]` = SYN-ACK, `[.]` = ACK, `[R]` = RST, `[F.]` = FIN
6. **Same seq number twice = retransmission** — use `-S` flag for absolute sequence numbers
7. **RST in response to SYN = port closed** — connection refused at the TCP level
8. **Rotate files in production** — `-C 100 -W 10` keeps 1 GB of rolling captures
9. **Limit snap length** — `-s 96` captures headers only, saves disk and reduces impact
10. **Combine with Wireshark** — capture with tcpdump on server, analyze with Wireshark on desktop

---

## Next

→ [curl and netcat](./04-curl-netcat.md) — HTTP testing and raw socket communication
