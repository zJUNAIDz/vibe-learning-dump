# tcpdump Deep Dive — Advanced Packet Analysis

> Module 13 covered tcpdump basics. This file goes deeper: dissecting packet headers byte by byte, writing complex BPF filters, analyzing multi-packet flows, and extracting meaning from raw captures. This is the skill that separates "I know tcpdump" from "I can debug production network issues."

---

## Table of Contents

1. [Recap: Where We Left Off](#recap)
2. [Dissecting Packet Headers](#dissecting)
3. [BPF Filter Deep Dive](#bpf)
4. [TCP Connection Analysis](#tcp-analysis)
5. [Analyzing Retransmissions and Loss](#retransmissions)
6. [Analyzing TLS Traffic](#tls)
7. [Analyzing DNS Problems](#dns)
8. [Multi-Flow Analysis](#multi-flow)
9. [tshark — Wireshark for the CLI](#tshark)
10. [Building Analysis Pipelines](#pipelines)
11. [Key Takeaways](#key-takeaways)

---

## Recap: Where We Left Off

```
From Module 13's tcpdump file, you know:
  - Basic capture: tcpdump -i eth0 -nn
  - Write/read files: -w capture.pcap / -r capture.pcap
  - Simple filters: host, port, tcp, udp
  - Reading TCP flags: [S], [S.], [.], [P.], [F.], [R]
  - Essential flags: -nn, -v, -c, -A, -X

This file goes deeper into ANALYSIS — extracting meaning from captures.
```

---

## Dissecting Packet Headers

### Hex dump interpretation

```bash
sudo tcpdump -i eth0 -nn -XX -c 1 port 80

# 14:23:45.123456 IP 10.0.0.2.54321 > 10.0.0.1.80: Flags [S], seq 12345, win 64240,
#   options [mss 1460,sackOK,TS val 123 ecr 0,nop,wscale 7], length 0
#
# 0x0000:  5254 0012 3456 5254 00ab cdef 0800 4500  RT..4VRT......E.
# 0x0010:  003c 1a2b 4000 4006 b1c2 0a00 0002 0a00  .<.+@.@.........
# 0x0020:  0001 d431 0050 0000 3039 0000 0000 a002  ...1.P..09......
# 0x0030:  faf0 fe30 0000 0204 05b4 0402 080a 0000  ...0............
# 0x0040:  007b 0000 0000 0103 0307                  .{........
```

### Ethernet header (14 bytes)

```
Offset  Bytes   Field
0x0000  6       Destination MAC: 52:54:00:12:34:56
0x0006  6       Source MAC:      52:54:00:ab:cd:ef
0x000c  2       EtherType:       0x0800 = IPv4
                                 0x0806 = ARP
                                 0x86DD = IPv6
```

### IPv4 header (20+ bytes, starts at 0x000e)

```
Offset  Bits    Field
0x000e  4       Version: 4 (IPv4)
        4       IHL: 5 (header length = 5×4 = 20 bytes)
0x000f  8       DSCP/ECN: 0x00
0x0010  16      Total Length: 0x003c = 60 bytes
0x0012  16      Identification: 0x1a2b
0x0014  3       Flags: 0x4000 → DF bit set (Don't Fragment)
        13      Fragment Offset: 0
0x0016  8       TTL: 0x40 = 64
0x0017  8       Protocol: 0x06 = TCP (0x11 = UDP, 0x01 = ICMP)
0x0018  16      Header Checksum: 0xb1c2
0x001a  32      Source IP: 0x0a000002 = 10.0.0.2
0x001e  32      Dest IP:   0x0a000001 = 10.0.0.1
```

### TCP header (20+ bytes, starts at 0x0022)

```
Offset  Bits    Field
0x0022  16      Source Port: 0xd431 = 54321
0x0024  16      Dest Port:   0x0050 = 80
0x0026  32      Sequence Number: 0x00003039 = 12345
0x002a  32      Acknowledgment:  0x00000000
0x002e  4       Data Offset: 0xa = 10 (header = 10×4 = 40 bytes, so 20 bytes of options)
        6       Reserved
        6       Flags: 0x002 = SYN
0x0030  16      Window Size: 0xfaf0 = 64240
0x0032  16      Checksum: 0xfe30
0x0034  16      Urgent Pointer: 0x0000
0x0036  var     Options (20 bytes in this case)

TCP Flags byte (0x002):
  FIN = 0x001
  SYN = 0x002
  RST = 0x004
  PSH = 0x008
  ACK = 0x010
  URG = 0x020
  ECE = 0x040
  CWR = 0x080
```

---

## BPF Filter Deep Dive

### Byte offset filters

BPF lets you match specific bytes at specific offsets — extremely powerful.

```bash
# Syntax: proto[offset:size]
# proto = ip, tcp, udp, icmp, ether, arp
# offset = byte offset from protocol header start
# size = 1, 2, or 4 bytes

# TCP flags byte is at offset 13 in TCP header
# tcp[13] = flags byte
sudo tcpdump -i eth0 'tcp[13] & 2 != 0'     # SYN flag set
sudo tcpdump -i eth0 'tcp[13] & 4 != 0'     # RST flag set
sudo tcpdump -i eth0 'tcp[13] == 2'          # SYN only (no ACK)
sudo tcpdump -i eth0 'tcp[13] == 18'         # SYN-ACK (SYN=2 + ACK=16 = 18)

# IP TTL (byte 8 in IP header)
sudo tcpdump -i eth0 'ip[8] < 10'           # TTL < 10 (dying packets)
sudo tcpdump -i eth0 'ip[8] == 1'           # TTL = 1 (traceroute probes)

# IP protocol (byte 9)
sudo tcpdump -i eth0 'ip[9] == 6'           # TCP
sudo tcpdump -i eth0 'ip[9] == 17'          # UDP
sudo tcpdump -i eth0 'ip[9] == 1'           # ICMP

# DSCP / TOS (byte 1 in IP header)
sudo tcpdump -i eth0 'ip[1] & 0xfc == 0xb8' # DSCP EF (Expedited Forwarding)

# ICMP type (byte 0 in ICMP header)
sudo tcpdump -i eth0 'icmp[0] == 3'         # Destination unreachable
sudo tcpdump -i eth0 'icmp[0] == 3 and icmp[1] == 4'  # Fragmentation needed
```

### Matching TCP payload

```bash
# TCP payload starts at: ip header length + tcp header length
# Complex filter for packets WITH payload (not just ACKs):

# IP header length = (ip[0] & 0x0f) × 4
# TCP header length = (tcp[12] >> 4) × 4
# Payload exists if: ip total length - ip header - tcp header > 0

sudo tcpdump -i eth0 'tcp port 80 and (((ip[2:2] - ((ip[0]&0xf)<<2)) - ((tcp[12]&0xf0)>>2)) != 0)'
# This is THE standard filter for "TCP packets with data"

# Match HTTP GET in payload
sudo tcpdump -i eth0 'tcp port 80 and tcp[((tcp[12:1] & 0xf0) >> 2):4] = 0x47455420'
# 0x47455420 = "GET " in hex
```

### Combining complex filters

```bash
# SYN packets to port 443 from a specific subnet
sudo tcpdump -i eth0 'tcp[13] == 2 and dst port 443 and src net 10.0.0.0/8'

# RST or FIN packets (connection disruptions)  
sudo tcpdump -i eth0 'tcp[13] & 5 != 0'  # RST(4) or FIN(1)

# Large packets (possible MTU issues)
sudo tcpdump -i eth0 'ip[2:2] > 1400 and tcp'

# Fragmented packets
sudo tcpdump -i eth0 '((ip[6:2] & 0x1fff) != 0) or (ip[6] & 0x20 != 0)'
```

---

## TCP Connection Analysis

### Following a complete connection

```bash
# Capture a curl request
sudo tcpdump -i eth0 -nn -S 'host example.com and port 80' -c 20

# Typical output (simplified):

# 1. Three-way handshake
# .001  10.0.0.2.54321 > 93.184.216.34.80: Flags [S], seq 1000, win 64240
# .002  93.184.216.34.80 > 10.0.0.2.54321: Flags [S.], seq 2000, ack 1001, win 65535
# .003  10.0.0.2.54321 > 93.184.216.34.80: Flags [.], ack 2001, win 502

# 2. HTTP request
# .004  10.0.0.2.54321 > 93.184.216.34.80: Flags [P.], seq 1001:1090, ack 2001, win 502

# 3. HTTP response
# .005  93.184.216.34.80 > 10.0.0.2.54321: Flags [.], ack 1090, win 509
# .006  93.184.216.34.80 > 10.0.0.2.54321: Flags [P.], seq 2001:3457, ack 1090, win 509

# 4. Client ACKs response
# .007  10.0.0.2.54321 > 93.184.216.34.80: Flags [.], ack 3457, win 496

# 5. Connection close
# .008  10.0.0.2.54321 > 93.184.216.34.80: Flags [F.], seq 1090, ack 3457, win 496
# .009  93.184.216.34.80 > 10.0.0.2.54321: Flags [F.], seq 3457, ack 1091, win 509
# .010  10.0.0.2.54321 > 93.184.216.34.80: Flags [.], ack 3458, win 496
```

### Sequence number analysis

```bash
# Use -S for absolute sequence numbers (easier to track)
# seq 1001:1090 means sent bytes 1001 through 1089 (89 bytes of data)
# ack 2001 means "I've received everything up to byte 2000, send me 2001 next"

# Retransmission detection:
# If you see seq 1001:1090 TWICE → retransmission
# The time between them = RTO (retransmission timeout)
```

### Window size tracking

```bash
# Window changes show flow control in action
# .001  ... win 64240    ← client has 64KB buffer
# .005  ... win 509      ← server has 509 × scale factor space
# .007  ... win 496      ← client buffer filling up
# ...
# .015  ... win 0        ← ZERO WINDOW! Client can't receive more!
# .020  ... win 502      ← Window update: client drained buffer, ready again
```

---

## Analyzing Retransmissions and Loss

### Detecting retransmissions

```bash
# Method 1: Look for duplicate sequence numbers
sudo tcpdump -i eth0 -nn -S port 80 | sort -t: -k1 | uniq -d
# Crude but shows repeated seq numbers

# Method 2: Use tshark (better)
tshark -r capture.pcap -Y 'tcp.analysis.retransmission'

# Method 3: Use tcpdump timing
sudo tcpdump -i eth0 -nn -ttt -S port 80
# -ttt shows time DELTA between packets
# Retransmissions show as: gap → then same seq number

# Retransmission timeline:
# 0.000000  seq 5000:6448   ← original
# 0.200000  seq 5000:6448   ← retransmit after 200ms (RTO)
# 0.600000  seq 5000:6448   ← second retransmit (exponential backoff)
```

### Detecting out-of-order delivery

```bash
# Out-of-order packets: sequence numbers arrive non-sequentially
# .001  seq 1000:2448   ← normal
# .002  seq 3896:5344   ← skipped 2448:3896!
# .003  seq 2448:3896   ← arrived late (out of order)
# Receiver sends duplicate ACKs for the gap

# Detect duplicate ACKs (signal of loss or reorder)
sudo tcpdump -i eth0 -nn -S 'tcp port 80' | grep "ack" | sort -k10 | uniq -c | sort -rn
# Shows ACK values that appear multiple times (duplicate ACKs)
```

---

## Analyzing TLS Traffic

### TLS handshake in tcpdump

```bash
sudo tcpdump -i eth0 -nn -v port 443 -c 10

# You can't see encrypted content, but you CAN see:
# 1. TCP handshake (unencrypted, as always)
# 2. ClientHello: TLS versions, cipher suites, SNI
# 3. ServerHello: selected cipher, certificate
# 4. Encrypted application data (opaque)

# TLS record types (first byte after TCP payload):
# 0x16 = Handshake
# 0x17 = Application Data (encrypted)
# 0x15 = Alert
# 0x14 = Change Cipher Spec (TLS 1.2)

# Filter TLS handshakes only:
sudo tcpdump -i eth0 -nn 'tcp port 443 and (tcp[((tcp[12:1] & 0xf0) >> 2)] == 0x16)'
```

### Extracting SNI (Server Name Indication)

```bash
# SNI is in the ClientHello — sent in plaintext!
# Use tshark for clean extraction:
tshark -r capture.pcap -Y 'tls.handshake.type == 1' -T fields -e tls.handshake.extensions_server_name
# example.com
# api.example.com
# cdn.example.com

# This reveals which websites a user visited (even with HTTPS)
# It's why ECH (Encrypted Client Hello) exists
```

### Using SSLKEYLOGFILE for full TLS decryption

```bash
# For debugging YOUR OWN connections:
# 1. Set SSLKEYLOGFILE to capture TLS session keys
export SSLKEYLOGFILE=/tmp/tls_keys.log

# 2. Generate traffic
curl https://example.com

# 3. Capture simultaneously
sudo tcpdump -i eth0 -nn -w tls_capture.pcap port 443

# 4. Open in Wireshark:
#    Edit → Preferences → Protocols → TLS → 
#    (Pre)-Master-Secret log filename: /tmp/tls_keys.log
#    
#    Now you can see DECRYPTED HTTP/2 traffic!
```

---

## Analyzing DNS Problems

### DNS query/response analysis

```bash
sudo tcpdump -i eth0 -nn -v port 53

# Normal query:
# 14:23:45.001 IP 10.0.0.2.54321 > 8.8.8.8.53: 
#   12345+ A? example.com. (29)

# Normal response:
# 14:23:45.015 IP 8.8.8.8.53 > 10.0.0.2.54321: 
#   12345 1/0/0 A 93.184.216.34 (45)

# NXDOMAIN (domain doesn't exist):
# 14:23:45.015 IP 8.8.8.8.53 > 10.0.0.2.54321: 
#   12345 NXDomain 0/1/0 (105)

# SERVFAIL (server error):
# 14:23:45.015 IP 8.8.8.8.53 > 10.0.0.2.54321: 
#   12345 ServFail 0/0/0 (29)

# Truncated (TC flag → retry with TCP):
# 14:23:45.015 IP 8.8.8.8.53 > 10.0.0.2.54321: 
#   12345 |TC| 0/0/0 (29)
```

### Detecting DNS issues

```bash
# Query sent but no response (timeout)
sudo tcpdump -i eth0 -nn port 53 -c 50
# If you see queries but no matching responses → DNS server unreachable

# Slow DNS (high latency)
sudo tcpdump -i eth0 -nn -ttt port 53
# 0.000000  query A? slow-domain.com
# 2.345000  response A  ← 2.3 seconds! Way too slow

# DNS over TCP (unusual, indicates large responses or zone transfers)
sudo tcpdump -i eth0 -nn 'tcp port 53'
```

---

## Multi-Flow Analysis

### Separating flows in a busy capture

```bash
# Count flows by 5-tuple
tcpdump -r capture.pcap -nn -q | \
  awk '{print $3, $5}' | sort | uniq -c | sort -rn | head
#   5432 10.0.0.2.54321 93.184.216.34.443:
#   3210 10.0.0.2.54322 93.184.216.34.443:
#    789 10.0.0.3.55555 10.0.0.1.80:

# Extract specific flow
tcpdump -r capture.pcap -nn 'host 10.0.0.2 and port 54321 and host 93.184.216.34'

# Count packets per protocol
tcpdump -r capture.pcap -nn -q | awk '{print $2}' | sort | uniq -c | sort -rn
#  12345 IP
#    567 ARP
#     23 IP6
```

### Timeline analysis

```bash
# Connection timing for all TCP streams
tshark -r capture.pcap -q -z conv,tcp
# Displays: endpoints, packets each direction, bytes, duration
# Great for finding the highest-volume or longest-lived connections

# IO graph data (packets per interval)
tshark -r capture.pcap -q -z io,stat,1
# Interval | Frames | Bytes
# 0-1      |   234  | 156789
# 1-2      |   456  | 345678   ← traffic spike!
# 2-3      |   123  | 78901
```

---

## tshark — Wireshark for the CLI

### Why tshark

```
tshark = Wireshark's CLI. It understands ALL the protocols Wireshark understands.
  - Deep protocol dissection (hundreds of protocols)
  - Display filters (more expressive than BPF capture filters)
  - Statistics and conversations
  - Stream reassembly
  - Available on servers (no GUI needed)
```

### Basic tshark usage

```bash
# Read pcap with protocol dissection
tshark -r capture.pcap

# Apply display filter (different from BPF capture filter!)
tshark -r capture.pcap -Y 'http.request'
tshark -r capture.pcap -Y 'tcp.analysis.retransmission'
tshark -r capture.pcap -Y 'dns.flags.rcode != 0'     # DNS errors
tshark -r capture.pcap -Y 'tls.handshake.type == 1'   # ClientHello

# Extract specific fields
tshark -r capture.pcap -Y 'http.request' -T fields -e ip.src -e http.host -e http.request.uri
# 10.0.0.2   example.com   /api/users
# 10.0.0.2   example.com   /api/users/123

# Statistics
tshark -r capture.pcap -q -z io,stat,1                    # packets per second
tshark -r capture.pcap -q -z conv,tcp                      # TCP conversations
tshark -r capture.pcap -q -z endpoints,ip                  # IP endpoints
tshark -r capture.pcap -q -z http,tree                     # HTTP stats
tshark -r capture.pcap -q -z expert                        # Expert analysis

# Follow TCP stream
tshark -r capture.pcap -q -z follow,tcp,ascii,0
```

### Key display filters

```bash
# TCP analysis (automatic detection)
tcp.analysis.retransmission          # Retransmissions
tcp.analysis.fast_retransmission     # Fast retransmit (3 dup ACKs)
tcp.analysis.duplicate_ack           # Duplicate ACKs
tcp.analysis.zero_window             # Zero window advertisements
tcp.analysis.window_update           # Window size updates
tcp.analysis.out_of_order            # Out-of-order segments
tcp.analysis.lost_segment            # Lost segments

# HTTP
http.request.method == "POST"
http.response.code >= 400
http.host contains "api"

# TLS
tls.handshake.type == 1              # ClientHello
tls.handshake.type == 2              # ServerHello
tls.record.content_type == 21        # Alert

# DNS
dns.qry.name == "example.com"
dns.flags.rcode == 3                 # NXDOMAIN
```

---

## Building Analysis Pipelines

### Automated capture analysis

```bash
# Quick health check of a pcap
analyze_pcap() {
    local file="$1"
    echo "=== Capture Summary ==="
    tshark -r "$file" -q -z io,stat,0 2>/dev/null
    
    echo -e "\n=== Retransmissions ==="
    tshark -r "$file" -Y 'tcp.analysis.retransmission' 2>/dev/null | wc -l
    
    echo -e "\n=== RST Packets ==="
    tshark -r "$file" -Y 'tcp.flags.reset == 1' 2>/dev/null | wc -l
    
    echo -e "\n=== DNS Errors ==="
    tshark -r "$file" -Y 'dns.flags.rcode != 0' 2>/dev/null | wc -l
    
    echo -e "\n=== Top Talkers ==="
    tshark -r "$file" -q -z endpoints,ip 2>/dev/null | head -15
    
    echo -e "\n=== Expert Warnings ==="
    tshark -r "$file" -q -z expert 2>/dev/null | grep -E "Warn|Error" | head -10
}

analyze_pcap capture.pcap
```

### Continuous monitoring

```bash
# Live retransmission rate
sudo tshark -i eth0 -q -z io,stat,5,"tcp.analysis.retransmission"
# Prints retransmit count every 5 seconds

# Live DNS query monitoring
sudo tshark -i eth0 -f 'udp port 53' -T fields -e dns.qry.name -e dns.flags.rcode
```

---

## Key Takeaways

1. **Understand packet structure** — knowing byte offsets in IP/TCP headers lets you write powerful BPF filters
2. **TCP flags at tcp[13]**: SYN=0x02, ACK=0x10, RST=0x04, FIN=0x01, SYN-ACK=0x12
3. **Same sequence number twice = retransmission** — use `-S` for absolute seq numbers
4. **Zero window = receiver buffer full** — application can't read fast enough
5. **TLS SNI is in plaintext** — `tshark -Y 'tls.handshake.type==1' -T fields -e tls.handshake.extensions_server_name`
6. **SSLKEYLOGFILE enables TLS decryption** in Wireshark — essential for debugging your own HTTPS traffic
7. **tshark is more powerful than tcpdump** for analysis — it understands hundreds of protocols
8. **Display filters ≠ capture filters** — BPF (tcpdump) for capture, Wireshark display filters for analysis
9. **`tcp.analysis.*` display filters** automatically detect retransmissions, duplicate ACKs, zero windows
10. **Build analysis scripts** — automate pcap analysis for quick triage in production incidents

---

## Next

→ [Wireshark](./02-wireshark.md) — Visual packet analysis with the world's most popular protocol analyzer
