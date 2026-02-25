# The Linux Kernel Networking Stack — How Packets Flow

> When a packet arrives at your NIC, it doesn't magically appear in your application. It traverses an intricate pipeline inside the Linux kernel — from hardware interrupt to device driver to protocol stack to socket buffer. Understanding this pipeline explains WHY things like packet drops happen, WHERE latency comes from, and HOW tools like tcpdump and iptables intercept packets.

---

## Table of Contents

1. [The Big Picture](#big-picture)
2. [NIC and Device Drivers](#nic-drivers)
3. [Interrupt Handling — NAPI](#napi)
4. [Network Stack Layers](#layers)
5. [Receive Path (Ingress)](#receive-path)
6. [Send Path (Egress)](#send-path)
7. [Netfilter Hook Points](#netfilter)
8. [Queuing Disciplines (qdisc)](#qdisc)
9. [Kernel Bypass and High Performance](#kernel-bypass)
10. [Tuning the Kernel Stack](#tuning)
11. [Debugging the Kernel Stack](#debugging)
12. [Key Takeaways](#key-takeaways)

---

## The Big Picture

```
Packet arrives at NIC
        │
        ▼
┌────────────────┐
│  NIC Hardware   │  DMA copies packet into ring buffer (RAM)
└───────┬────────┘
        │ hardware interrupt
        ▼
┌────────────────┐
│  Device Driver  │  NAPI: schedule softirq, disable further interrupts
└───────┬────────┘
        │ softirq (NET_RX_SOFTIRQ)
        ▼
┌────────────────┐
│  Network Core   │  __netif_receive_skb() → allocate sk_buff
│  (GRO, RPS)     │  Generic Receive Offload, Receive Packet Steering
└───────┬────────┘
        │
        ▼
┌────────────────┐
│  Netfilter      │  PREROUTING chain (iptables/nftables)
│  (PREROUTING)   │  conntrack, DNAT
└───────┬────────┘
        │
        ▼
┌────────────────┐
│  Routing        │  Is this packet for us? → INPUT chain
│  Decision       │  Is it for someone else? → FORWARD chain
└───────┬────────┘
        │
   ┌────┴────┐
   │ INPUT   │  Netfilter INPUT chain
   └────┬────┘
        │
        ▼
┌────────────────┐
│  Transport      │  TCP: sequence check, ACK, reassembly
│  (TCP/UDP)      │  UDP: checksum verify, deliver
└───────┬────────┘
        │
        ▼
┌────────────────┐
│  Socket Buffer  │  Data lands in socket's receive buffer
│  (sk_buff)      │  Application calls recv() → data copied to userspace
└────────────────┘
```

---

## NIC and Device Drivers

### Ring buffers

```
The NIC and kernel communicate through ring buffers (circular queues) in RAM:

        ┌─────────────────────────────────────┐
        │         Ring Buffer (RX)             │
        │  ┌───┬───┬───┬───┬───┬───┬───┬───┐  │
        │  │pkt│pkt│   │   │   │   │pkt│pkt│  │
        │  └───┴───┴───┴───┴───┴───┴───┴───┘  │
        │    ▲                           ▲     │
        │    │                           │     │
        │  Head (NIC writes here)   Tail       │
        │                    (kernel reads)    │
        └─────────────────────────────────────┘

1. NIC receives frame from wire
2. NIC DMAs (Direct Memory Access) frame into next ring buffer slot
3. NIC raises hardware interrupt
4. Kernel reads frame from ring buffer, advances tail pointer

If ring buffer is FULL (kernel too slow):
  → NIC drops packet SILENTLY
  → "rx_missed_errors" or "rx_no_buffer_count" in ethtool stats
```

### Checking and tuning ring buffers

```bash
# View ring buffer sizes
ethtool -g eth0
# Ring parameters for eth0:
# Pre-set maximums:
# RX:      4096
# TX:      4096
# Current hardware settings:
# RX:      256    ← might be too small!
# TX:      256

# Increase ring buffer
ethtool -G eth0 rx 4096 tx 4096

# View NIC statistics (packet drops, errors)
ethtool -S eth0 | grep -i "drop\|error\|miss\|fifo"
# rx_missed_errors: 0
# rx_no_buffer_count: 0     ← ring buffer overflow drops
# tx_dropped: 0
```

---

## Interrupt Handling — NAPI

### The problem with interrupts

```
Old approach: Interrupt per packet
  1 Gbps ≈ ~1.5 million packets/sec (small packets)
  = 1.5 million interrupts/sec
  = CPU does nothing but handle interrupts (livelock!)

Solution: NAPI (New API) — interrupt + polling hybrid
```

### How NAPI works

```
1. First packet arrives → NIC raises hardware interrupt
2. Driver interrupt handler:
   a. DISABLES further NIC interrupts
   b. Schedules a softirq (NET_RX_SOFTIRQ)
3. Softirq handler runs (in ksoftirqd thread):
   a. POLLS ring buffer — grabs up to budget packets (default 64)
   b. Processes each packet
   c. If more packets remain → keep polling
   d. If ring buffer empty → RE-ENABLE interrupts
   
Result:
  - Low load: interrupt-driven (low latency)
  - High load: polling mode (high throughput, no interrupt storm)
```

### Softirqs and ksoftirqd

```bash
# See softirq counts
cat /proc/softirqs
#                  CPU0       CPU1       CPU2       CPU3
# NET_RX:      1234567    2345678    3456789    4567890
# NET_TX:       123456     234567     345678     456789

# See ksoftirqd threads
ps aux | grep ksoftirqd
# root  3  ... [ksoftirqd/0]
# root  8  ... [ksoftirqd/1]
# One per CPU

# Monitor softirq processing
watch -d cat /proc/softirqs
```

---

## Network Stack Layers

### The sk_buff structure

```
sk_buff (socket buffer) is THE fundamental data structure.
Every packet in the kernel is an sk_buff:

┌──────────────────────────────────────────┐
│                sk_buff                    │
├──────────────────────────────────────────┤
│  Metadata:                                │
│    - Timestamp                            │
│    - Network device (which NIC)           │
│    - Protocol (ETH_P_IP, ETH_P_ARP, ...)  │
│    - Conntrack state                      │
│    - Marks, priority                      │
│    - Socket owner (if relevant)           │
├──────────────────────────────────────────┤
│  Pointers into packet data:               │
│    head ──→ ┌──────────────┐              │
│    data ──→ │ L2 header    │ (Ethernet)   │
│             │ L3 header    │ (IP)         │
│             │ L4 header    │ (TCP/UDP)    │
│    tail ──→ │ Payload      │              │
│    end ───→ └──────────────┘              │
│                                           │
│  As packet moves UP the stack:            │
│    data pointer advances past each header │
│  As packet moves DOWN the stack:          │
│    headers are prepended (data ptr moves back) │
└──────────────────────────────────────────┘
```

### Layer processing

```
Each layer does specific work:

L2 (Ethernet):
  - Check destination MAC (for us? broadcast? multicast?)
  - Strip Ethernet header
  - Determine upper protocol (EtherType: 0x0800=IPv4, 0x0806=ARP)

L3 (IP):
  - Validate header checksum
  - Check destination IP (for us? forward?)
  - Defragment if fragmented
  - Make routing decision
  - Decrement TTL (if forwarding)

L4 (TCP):
  - Validate checksum
  - Match to socket (by 5-tuple)
  - Sequence number processing
  - ACK processing
  - Reassemble data stream
  - Deliver to socket receive buffer

L4 (UDP):
  - Validate checksum (optional in IPv4, mandatory IPv6)
  - Match to socket (by 4-tuple: proto, dst_ip, dst_port, or connected src)
  - Deliver datagram to socket receive buffer
```

---

## Receive Path (Ingress)

### Step by step

```
1. NIC → DMA → Ring Buffer
   Packet lands in pre-allocated ring buffer slot

2. Hardware Interrupt → NAPI
   Driver schedules NET_RX_SOFTIRQ

3. NAPI poll → __netif_receive_skb()
   sk_buff allocated, metadata populated

4. GRO (Generic Receive Offload)
   Merges multiple small packets into one large sk_buff
   Reduces per-packet processing overhead
   E.g., 10 TCP segments → 1 large segment for stack processing

5. RPS (Receive Packet Steering)
   Distributes packets across CPUs (if NIC has only 1 RX queue)
   Hash of packet headers → pick CPU → queue to that CPU's backlog

6. tc (traffic control) ingress
   If configured: classify, police, filter

7. Netfilter PREROUTING
   conntrack, DNAT, raw table

8. Routing Decision
   ip_route_input() → local delivery (INPUT) or forwarding (FORWARD)

9. Netfilter INPUT
   Filter rules for locally-destined packets

10. Transport Protocol Handler
    tcp_v4_rcv() or udp_rcv()
    Socket lookup → deliver to socket receive buffer

11. Socket Receive Buffer
    Data available for application recv()/read()

12. Application
    Wakes up from epoll_wait()/select() or block on recv()
    Data copied from kernel buffer to userspace buffer
```

### Where packets get dropped (ingress)

```
Drop point                  How to detect
─────────────────────────────────────────────────
NIC ring buffer full        ethtool -S ethX | grep rx_missed
NIC firmware/hardware       ethtool -S ethX | grep rx_errors
Driver softirq budget       /proc/net/softnet_stat (2nd column)
Netfilter (iptables DROP)   iptables -L -v -n (check counters)
Routing (no route)          ip route get <dst>
TCP: socket buffer full     ss -tnm (Recv-Q)
Socket: backlog full        ss -tnl (Recv-Q vs Send-Q on LISTEN)
Application too slow        Socket buffer fills → drops
```

---

## Send Path (Egress)

### Step by step

```
1. Application calls send()/write()
   Data copied from userspace to socket send buffer

2. Transport Protocol
   TCP: segment data, compute checksum, manage window
   UDP: create datagram, compute checksum

3. IP Layer
   Add IP header, select source IP, make routing decision
   Fragment if needed (avoid this — use PMTUD)

4. Netfilter OUTPUT
   Filter rules for locally-generated packets

5. Routing
   Determine output device and next-hop

6. Netfilter POSTROUTING
   SNAT, MASQUERADE

7. Neighbor Subsystem (ARP)
   Resolve next-hop MAC address
   If not in ARP cache → ARP request → queue packet

8. tc (traffic control) egress / qdisc
   Queue packet in output queue
   Apply shaping, scheduling, dropping policy

9. Device Driver
   DMA packet to NIC's TX ring buffer

10. NIC Hardware
    Transmit frame on wire
    Raise TX completion interrupt
    Driver frees sk_buff
```

### TSO and GSO

```
TSO (TCP Segmentation Offload):
  Application sends large chunk (64KB)
  Kernel creates ONE large sk_buff
  NIC hardware splits into MTU-sized segments
  → CPU doesn't do segmentation → huge performance gain

GSO (Generic Segmentation Offload):
  Same as TSO but done in kernel (software) if NIC doesn't support TSO
  Still faster because segmentation happens at the last moment

Check offload status:
  ethtool -k eth0 | grep offload
  # tcp-segmentation-offload: on
  # generic-segmentation-offload: on
  # generic-receive-offload: on
```

---

## Netfilter Hook Points

```
Netfilter hooks in the packet path:

                        Incoming packet
                              │
                              ▼
                    ┌─────────────────┐
                    │   PREROUTING    │  (raw, conntrack, mangle, nat)
                    └────────┬────────┘
                             │
                        Routing decision
                        ┌────┴────┐
                  For us │        │ Forward
                        ▼         ▼
                ┌────────────┐ ┌────────────┐
                │   INPUT    │ │  FORWARD   │
                │            │ │            │
                └────┬───────┘ └────┬───────┘
                     │              │
               Local process        │
                     │              │
                     ▼              │
                ┌────────────┐     │
                │   OUTPUT   │     │
                └────┬───────┘     │
                     │              │
                     └──────┬───────┘
                            ▼
                    ┌─────────────────┐
                    │  POSTROUTING    │  (mangle, nat: SNAT/MASQUERADE)
                    └────────┬────────┘
                             │
                             ▼
                       Outgoing packet

Each hook can have multiple tables (filter, nat, mangle, raw)
registered at different priorities.
```

---

## Queuing Disciplines (qdisc)

### What qdiscs do

```
Every network interface has a qdisc (queuing discipline):
  - Sits between IP layer and device driver
  - Controls HOW packets are queued for transmission
  - Implements scheduling, shaping, policing

Default: pfifo_fast (simple priority FIFO)
Modern default: fq_codel (Fair Queuing with Controlled Delay)
```

### Common qdiscs

```
pfifo_fast:
  3 priority bands (TOS-based)
  FIFO within each band
  Simple, no shaping

fq_codel (recommended):
  Fair queuing: separate queue per flow
  CoDel: intelligent drop to fight bufferbloat
  Best general-purpose qdisc

htb (Hierarchical Token Bucket):
  Rate limiting and bandwidth sharing
  Used for traffic shaping
  Example: limit upload to 10 Mbps

tbf (Token Bucket Filter):
  Simple rate limiting
  Burst + sustained rate

netem (Network Emulator):
  Add delay, loss, jitter, corruption
  Great for testing!
```

### Viewing and configuring qdiscs

```bash
# View current qdisc
tc qdisc show dev eth0
# qdisc fq_codel 0: root refcnt 2 limit 10240p flows 1024 ...

# Add delay (testing)
tc qdisc add dev eth0 root netem delay 100ms 20ms
# 100ms delay ± 20ms jitter

# Rate limit
tc qdisc add dev eth0 root tbf rate 10mbit burst 32kbit latency 400ms

# Remove
tc qdisc del dev eth0 root

# Statistics
tc -s qdisc show dev eth0
# Sent 1234567 bytes 12345 pkt (dropped 0, overlimits 0 requeues 0)
```

---

## Kernel Bypass and High Performance

### Why bypass the kernel?

```
Kernel networking overhead per packet:
  - Interrupt handling
  - sk_buff allocation
  - Multiple pointer dereferences
  - Netfilter traversal
  - Protocol processing
  - Data copy (kernel → userspace)

For 10/25/100 Gbps:
  - 14.88 Mpps at 10 Gbps (64-byte packets)
  - ~67 ns per packet budget
  - Kernel stack takes ~5-10 μs per packet
  - IMPOSSIBLE to keep up!
```

### Kernel bypass technologies

```
DPDK (Data Plane Development Kit):
  - NIC mapped directly to userspace
  - Application polls NIC (no interrupts)
  - Zero-copy packet processing
  - Used by: high-perf firewalls, routers, load balancers

XDP (eXpress Data Path):
  - eBPF programs run AT the NIC driver (before sk_buff allocation)
  - Can DROP, REDIRECT, or PASS packets
  - ~100× faster than iptables for DDoS mitigation
  - Used by: Cloudflare, Facebook, Cilium (Kubernetes)

AF_XDP:
  - Socket type that receives XDP-redirected packets
  - Userspace gets raw packets with minimal overhead
  - Compromise between DPDK (full bypass) and kernel stack

io_uring:
  - General async I/O framework (not networking-specific)
  - Eliminates system call overhead (shared ring buffers)
  - Growing networking support
```

---

## Tuning the Kernel Stack

### Key sysctls

```bash
# Receive buffer sizes (TCP autotuning)
net.ipv4.tcp_rmem = 4096 131072 6291456    # min default max
net.ipv4.tcp_wmem = 4096 16384 4194304     # min default max
net.core.rmem_max = 16777216               # max receive buffer
net.core.wmem_max = 16777216               # max send buffer

# Backlog queues
net.core.netdev_max_backlog = 5000         # queue before processing (default 1000)
net.core.somaxconn = 65535                 # max listen() backlog (default 4096)

# TCP tuning
net.ipv4.tcp_max_syn_backlog = 65535       # SYN queue size
net.ipv4.tcp_fin_timeout = 30              # TIME_WAIT duration (default 60)
net.ipv4.tcp_tw_reuse = 1                  # reuse TIME_WAIT for outbound
net.ipv4.tcp_slow_start_after_idle = 0     # don't reset cwnd on idle

# Conntrack (if used)
net.netfilter.nf_conntrack_max = 1048576
```

### Multi-queue and CPU affinity

```bash
# Modern NICs have multiple RX/TX queues
# Each queue → one CPU → parallel processing

# See number of queues
ethtool -l eth0
# Combined: 8 (i.e., 8 RX+TX queue pairs)

# See interrupt → CPU mapping
cat /proc/interrupts | grep eth0
# 45:  12345  0  0  0  IR-PCI-MSI  eth0-TxRx-0
# 46:  0  23456  0  0  IR-PCI-MSI  eth0-TxRx-1
# ...

# Set CPU affinity for interrupt
echo 1 > /proc/irq/45/smp_affinity  # pin to CPU 0
echo 2 > /proc/irq/46/smp_affinity  # pin to CPU 1

# Or use irqbalance (automatic)
systemctl status irqbalance
```

---

## Debugging the Kernel Stack

### Where to look

```bash
# Overall network statistics
cat /proc/net/snmp
# Shows IP, TCP, UDP error counters

# Detailed protocol stats
cat /proc/net/netstat
# TcpExt: SyncookiesSent SyncookiesRecv SyncookiesFailed
#   ListenOverflows ListenDrops ...

# Useful one-liner
nstat -az | grep -i "drop\|overflow\|error\|retrans"
# TcpExtListenOverflows    42    ← accept queue overflow!
# TcpExtListenDrops        42    ← packets dropped due to accept overflow
# TcpExtTCPTimeouts        10    ← retransmission timeouts
# TcpExtSyncookiesSent     100   ← SYN queue overflow → SYN cookies
# UdpInErrors              5     ← UDP receive buffer overflow

# NIC-level drops
ethtool -S eth0 | grep -i drop

# Softirq processing
cat /proc/net/softnet_stat
# Each row = one CPU: [processed] [dropped] [time_squeeze]
# Column 2 (dropped) > 0 → netdev_max_backlog too small
# Column 3 (time_squeeze) > 0 → softirq budget exhausted

# Interface drops
ip -s link show eth0
# RX:  bytes packets errors dropped missed mcast
# TX:  bytes packets errors dropped carrier collisions
```

### Packet tracing

```bash
# Trace a packet through Netfilter
# (see 02-iptables-nftables.md for TRACE details)
iptables -t raw -A PREROUTING -p tcp --dport 80 -j TRACE
dmesg | grep TRACE

# Drop monitoring (kernel 5.14+)
# dropwatch shows WHERE in the kernel packets are dropped
dropwatch -l kas
# start
# ... 
# drop at: nf_hook_slow+0xa3/0x140 (packet dropped by netfilter)
# drop at: tcp_v4_rcv+0x47/0x9e0 (packet dropped by TCP)

# perf tracing
perf trace -e 'skb:*' -a --duration 5
```

---

## Key Takeaways

1. **Packets traverse a long pipeline** — NIC → DMA → interrupt → NAPI → GRO/RPS → Netfilter → routing → transport → socket → application
2. **Ring buffers are the first bottleneck** — if the kernel can't process fast enough, NIC drops packets silently. Check with `ethtool -S`
3. **NAPI prevents interrupt livelock** — switches from interrupt-driven to polling under load
4. **sk_buff is the universal packet structure** — every packet in the kernel lives in an sk_buff
5. **GRO reduces per-packet overhead** — merges small packets before stack processing
6. **Netfilter hooks at 5 points** — PREROUTING, INPUT, FORWARD, OUTPUT, POSTROUTING
7. **qdiscs control egress queuing** — use fq_codel to fight bufferbloat, netem for testing
8. **For 10+ Gbps, kernel is too slow** — use DPDK, XDP, or AF_XDP to bypass the stack
9. **Key tuning points**: ring buffer size, backlog queue, buffer sizes, CPU affinity, conntrack
10. **Debug with**: `ethtool -S` (NIC drops), `nstat` (protocol errors), `/proc/net/softnet_stat` (CPU drops), `dropwatch` (kernel drop location)

---

## Next

→ [Buffers, epoll, and select](./03-buffers-epoll-select.md) — I/O multiplexing and the event-driven model that powers modern servers
