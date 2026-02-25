# Wireless and Mobile Networking

> Wi-Fi is the network most users actually touch. Everything you've learned about Ethernet, TCP, and routing still applies — but wireless adds an entirely new set of challenges: shared medium, signal degradation, mobility, and the hidden node problem. Understanding wireless is understanding why "it works on my machine" is so often a Wi-Fi issue.

---

## Table of Contents

1. [Why Wireless is Different](#why-different)
2. [Wi-Fi Fundamentals (802.11)](#wifi-fundamentals)
3. [Radio Frequency Basics](#rf-basics)
4. [Wi-Fi Channels and Bands](#channels)
5. [CSMA/CA — The Wireless Access Method](#csma-ca)
6. [The Hidden Node Problem](#hidden-node)
7. [Wi-Fi Frame Types](#frame-types)
8. [Authentication and Association](#auth)
9. [Wi-Fi Security (WPA2, WPA3)](#security)
10. [Enterprise Wi-Fi (802.1X)](#enterprise)
11. [Roaming and Handoff](#roaming)
12. [Wi-Fi Performance Factors](#performance)
13. [Cellular Networking Basics](#cellular)
14. [How Mobile Data Works (3G → 5G)](#mobile-data)
15. [Mobile IP and Handover](#mobile-ip)
16. [Troubleshooting Wireless Issues](#troubleshooting)
17. [Key Takeaways](#key-takeaways)

---

## Why Wireless is Different

```
Wired Ethernet:                    Wireless (Wi-Fi):
─────────────                      ─────────────────

Dedicated cable                    Shared radio medium
Full duplex                        Half duplex (can't send + receive)
Predictable speed                  Speed varies with distance/interference
Point-to-point link                Broadcast medium (everyone hears)
Cable doesn't change               Environment changes constantly
No signal degradation              Signal weakens with distance
100m max (Cat6)                    ~30m indoor / ~100m outdoor (typical)
CSMA/CD (collision detect)         CSMA/CA (collision avoid — can't detect!)
```

### The fundamental challenge

```
                        ┌─────────────────┐
                        │   Access Point   │
                        │    (AP / Router) │
                        └─────────┬───────┘
                                  │ Radio
                           ╔═════╧═════╗
                           ║  Shared    ║
                           ║  Airspace  ║
                           ╚═════╤═════╝
                        ┌────────┼────────┐
                        │        │        │
                     Laptop   Phone    Tablet
                     
All devices share THE SAME radio channel.
Only ONE device can transmit at a time.
Every device hears every transmission.
→ This is fundamentally different from switched Ethernet
  where each port has dedicated bandwidth.
```

---

## Wi-Fi Fundamentals (802.11)

### The 802.11 family

```
Standard    Year   Band      Max Speed    Notes
────────    ────   ────      ─────────    ─────
802.11b     1999   2.4 GHz   11 Mbps      First widely adopted
802.11a     1999   5 GHz     54 Mbps      Higher frequency, shorter range
802.11g     2003   2.4 GHz   54 Mbps      Backwards compat with b
802.11n     2009   2.4/5     600 Mbps     MIMO (Wi-Fi 4)
802.11ac    2013   5 GHz     6.9 Gbps     MU-MIMO, beamforming (Wi-Fi 5)
802.11ax    2020   2.4/5/6   9.6 Gbps     OFDMA, BSS coloring (Wi-Fi 6/6E)
802.11be    2024   2.4/5/6   46 Gbps      MLO (Wi-Fi 7)
```

### Wi-Fi architecture

```
BSS (Basic Service Set) — Single AP:

  ┌──────────────┐
  │     AP       │ ← Connected to wired network (uplink)
  │  BSSID:      │
  │  AA:BB:CC:.. │
  └──────┬───────┘
         │ Radio
    ┌────┴────┐
    │   BSS   │  ← All clients associated with this AP
    │         │     share this BSS
    └─────────┘

ESS (Extended Service Set) — Multiple APs, same SSID:

  ┌────────┐              ┌────────┐
  │  AP 1  │──── LAN ─────│  AP 2  │
  └───┬────┘              └───┬────┘
      │                       │
  ┌───┴───┐              ┌───┴───┐
  │ BSS 1 │              │ BSS 2 │
  │       │ ← overlap →  │       │
  └───────┘              └───────┘
  
  Same SSID, different channels
  Clients can roam between APs
```

---

## Radio Frequency Basics

### Signal propagation

```
Signal strength decreases with distance (inverse square law):

  AP                                          
  🔊 ═══════════════════════════════════════►  
  │                                           │
  │  -30 dBm   -50 dBm   -67 dBm   -80 dBm  │     
  │  Excellent  Good      OK        Poor      │
  │                                           │
  ├─── 3m ────┼─── 10m ──┼─── 20m ──┼── 30m ─┤


Signal strength (RSSI) interpretation:
  -30 to -50 dBm : Excellent (right next to AP)
  -50 to -60 dBm : Good (same room)
  -60 to -70 dBm : OK (through 1-2 walls)
  -70 to -80 dBm : Weak (edge of range)
  -80 to -90 dBm : Barely usable
  Below -90 dBm  : Disconnection likely
```

### Signal-to-Noise Ratio (SNR)

```
What matters is NOT just signal strength.
It's signal RELATIVE TO noise (SNR):

  SNR = Signal Power - Noise Floor
  
  Good:  Signal = -50 dBm, Noise = -95 dBm → SNR = 45 dB ✓
  Bad:   Signal = -50 dBm, Noise = -60 dBm → SNR = 10 dB ✗ (microwave!)

  SNR > 40 dB : Excellent (can use highest speeds)
  SNR 25-40 dB: Good
  SNR 15-25 dB: Fair (speed will be limited)
  SNR < 15 dB : Poor (retransmissions, disconnects)

Common noise sources:
  - Microwave ovens (2.4 GHz — same frequency as Wi-Fi!)
  - Bluetooth devices (2.4 GHz)
  - Baby monitors
  - Other Wi-Fi networks (co-channel interference)
  - USB 3.0 devices (emit 2.4 GHz interference)
```

---

## Wi-Fi Channels and Bands

### 2.4 GHz band

```
Only 3 non-overlapping channels: 1, 6, 11

  Channel:  1    2    3    4    5    6    7    8    9   10   11
            ├────────────────┤
                 ├────────────────┤
                      ├────────────────┤
                                  ├────────────────┤
                                                   ├────────────────┤

Non-overlapping channels (use these!):
  Channel 1:  [████████████████]
  Channel 6:            [████████████████]
  Channel 11:                      [████████████████]

RULE: Always use 1, 6, or 11 on 2.4 GHz
  Using channel 3 or 4 causes overlapping interference with BOTH 1 and 6
  Overlapping interference is WORSE than co-channel interference
```

### 5 GHz band

```
Many more non-overlapping channels (20 MHz each):

  36, 40, 44, 48       ← UNII-1 (indoor)
  52, 56, 60, 64       ← UNII-2 (DFS required — radar avoidance)
  100-144               ← UNII-2 Extended (DFS required)
  149, 153, 157, 161   ← UNII-3 (indoor/outdoor)

5 GHz advantages:
  + Many more channels → less interference
  + Higher speeds (wider channels: 40/80/160 MHz)
  + Less crowded (fewer devices use it)

5 GHz disadvantages:
  - Shorter range (higher frequency = more absorption)
  - Doesn't penetrate walls as well
  - DFS channels may require radar avoidance
```

### Channel width

```
Wider channels = more speed but more interference:

  20 MHz: Standard, most reliable
  40 MHz: 2x speed, uses 2 channels (only useful on 5 GHz)
  80 MHz: 4x speed (Wi-Fi 5/6, 5 GHz only)
  160 MHz: 8x speed (Wi-Fi 6, requires many free channels)

Trade-off:
  ┌─────────────────────────────────────┐
  │   Narrow channel   Wide channel     │
  │   ├──20──┤         ├──────80──────┤ │
  │   More reliable    More speed       │
  │   Less speed       More interference│
  │   Works at range   Needs strong SNR │
  └─────────────────────────────────────┘
```

---

## CSMA/CA — The Wireless Access Method

### Why not CSMA/CD?

```
Ethernet uses CSMA/CD (Collision Detection):
  1. Listen before sending
  2. If collision detected → stop, wait random time, retry
  3. Works because you CAN detect collisions on a wire

Wi-Fi CANNOT detect collisions:
  - When you're transmitting, your own signal drowns out others
  - You can't hear a collision while sending
  - By the time you realize → whole frame is wasted

Solution: CSMA/CA (Collision Avoidance)
  Avoid collisions instead of detecting them
```

### How CSMA/CA works

```
1. LISTEN (Clear Channel Assessment)
   Is the channel busy?
   
   If busy → wait until clear + random backoff
   If clear → proceed to step 2

2. WAIT (DIFS)
   Wait a mandatory interval (DCF Interframe Spacing)
   Still clear? → proceed
   
3. RANDOM BACKOFF
   Wait random additional time (0 to CW × slot time)
   This prevents all waiting stations from transmitting simultaneously
   
4. TRANSMIT
   Send the frame
   
5. WAIT FOR ACK
   Receiver sends ACK after SIFS (shortest wait)
   
   If ACK received → success!
   If no ACK → collision/error occurred → double backoff window, retry

Timeline:
  Time →
  ├─BUSY─┤─DIFS─┤─Backoff─┤─────FRAME─────┤─SIFS─┤─ACK─┤
                                                     
  The ACK gets priority (SIFS < DIFS) so no one else
  can jump in before the acknowledgment.
```

---

## The Hidden Node Problem

```
The classic Wi-Fi problem that CSMA/CA alone can't solve:

    ┌────┐                 ┌────┐                 ┌────┐
    │ A  │                 │ AP │                 │ B  │
    └────┘                 └────┘                 └────┘
    
    A's range: [=========AP=========]
    B's range:             [=========AP=========]
    
    A can hear AP but NOT B
    B can hear AP but NOT A
    
    A and B are "hidden" from each other!

Problem:
    A does carrier sense → channel clear (can't hear B)
    B does carrier sense → channel clear (can't hear A)
    Both transmit simultaneously → collision at AP!
    Neither A nor B knows about the collision!

Solution: RTS/CTS (Request to Send / Clear to Send)

    A → AP:  RTS (I want to send, need X time)
    AP → All: CTS (A is sending, everyone wait X time)
    B hears CTS → stays quiet for X time
    A transmits → no collision

    Timeline:
    A:  [RTS]─────────────[DATA──────────────]─[ACK]
    AP: ─────[CTS]───────────────────────────[ACK]──
    B:  ──────────QUIET (heard CTS)──────────────────

    Trade-off: RTS/CTS adds overhead
    Only useful when hidden nodes cause problems
```

---

## Wi-Fi Frame Types

```
Wi-Fi uses THREE types of frames, not just data:

1. MANAGEMENT FRAMES
   ├── Beacon         AP broadcasts SSID, capabilities every ~100ms
   ├── Probe Request  Client searching for networks
   ├── Probe Response AP responds to probe
   ├── Authentication Client authenticates with AP
   ├── Association    Client joins the AP's BSS
   ├── Deauthentication  Disconnect (can be spoofed!)
   └── Disassociation    Disconnect from BSS

2. CONTROL FRAMES
   ├── ACK            Acknowledge received frame
   ├── RTS            Request to Send
   ├── CTS            Clear to Send
   └── Block ACK      Acknowledge multiple frames

3. DATA FRAMES
   ├── Data           Regular data
   ├── QoS Data       Priority data (voice/video)
   └── Null Data      Keepalive (no payload)
```

### Wi-Fi frame addresses (4 addresses!)

```
Ethernet frame: Source MAC → Destination MAC (2 addresses)

Wi-Fi frame: Up to 4 addresses!

  Address 1: Receiver Address (immediate recipient)
  Address 2: Transmitter Address (immediate sender)
  Address 3: Destination Address (final destination)
  Address 4: Source Address (only in WDS/mesh)

Why? Because the AP is a relay:

  Client → AP → Wired Network
  
  Address 1 = AP (receiver of radio frame)
  Address 2 = Client (sender of radio frame)
  Address 3 = Gateway MAC (where AP should forward to)
  
  The AP rewrites this into an Ethernet frame:
  Source MAC = Client, Dest MAC = Gateway
```

---

## Authentication and Association

### Connection process

```
Client connecting to Wi-Fi:

  Client                          AP
    │                              │
    │────── Probe Request ────────>│  "What networks are here?"
    │<───── Probe Response ────────│  "I'm NetworkX, WPA2, channel 6"
    │                              │
    │────── Authentication ───────>│  Open System Authentication
    │<───── Authentication ────────│  (just a formality in WPA2)
    │                              │
    │────── Association Req ──────>│  "Can I join? I support X, Y, Z"
    │<───── Association Resp ──────│  "OK, your AID is 3"
    │                              │
    │═══════ 4-Way Handshake ═════>│  WPA2 key exchange
    │<════════════════════════════=│  (derive per-session keys)
    │                              │
    │  ← Connected, can send data →│

Total time: 100-500ms typically
  (significant compared to wired: instant)
```

---

## Wi-Fi Security (WPA2, WPA3)

### Evolution

```
Protocol    Year   Encryption   Key Exchange       Status
────────    ────   ──────────   ────────────       ──────
WEP         1999   RC4          Static key         BROKEN (minutes to crack)
WPA         2003   TKIP/RC4     Pre-shared/802.1X  Deprecated
WPA2        2004   AES-CCMP     Pre-shared/802.1X  Standard (KRACK vuln found)
WPA3        2018   AES-GCMP     SAE/802.1X         Current best
```

### WPA2-Personal (PSK)

```
Pre-Shared Key (the Wi-Fi password everyone knows):

  Password → PBKDF2(password, SSID, 4096) → PSK (256-bit)
  
  4-Way Handshake:
  ┌────────┐                           ┌────────┐
  │ Client │                           │   AP   │
  └───┬────┘                           └───┬────┘
      │                                    │
      │← ANonce (AP's random number)───────│  Message 1
      │                                    │
      │  Client generates PTK from:        │
      │  PSK + ANonce + SNonce + MACs      │
      │                                    │
      │── SNonce + MIC ───────────────────>│  Message 2
      │                                    │
      │  AP generates same PTK             │
      │  Verifies MIC (proves client       │
      │  has the correct PSK)              │
      │                                    │
      │← GTK (encrypted) + MIC ───────────│  Message 3
      │                                    │
      │── ACK + MIC ─────────────────────>│  Message 4
      │                                    │
      │  ← Encrypted communication →       │

  PTK = Pairwise Transient Key (unique per session)
  GTK = Group Temporal Key (for broadcast/multicast)
  
  Every client gets a DIFFERENT PTK (even with same password)
```

### WPA3-Personal (SAE)

```
SAE = Simultaneous Authentication of Equals

Improvement over WPA2-PSK:
  1. Forward secrecy — captured traffic can't be decrypted later
     even if password is compromised
  2. Resistant to offline dictionary attacks
  3. No handshake capture → password crack possible
  
WPA2 vulnerability:
  Attacker captures 4-way handshake → offline brute-force password
  
WPA3 fix:
  SAE (Dragonfly) exchange before key derivation
  Each attempt requires interaction with AP → can't brute-force offline
```

---

## Enterprise Wi-Fi (802.1X)

```
WPA2/WPA3-Enterprise: Each user has unique credentials

  ┌────────┐      ┌────────┐      ┌────────┐
  │ Client │ WiFi │   AP   │ RADIUS│ Server │
  │(Suppl.)│──────│(Auth.) │───────│(RADIUS)│
  └────────┘      └────────┘      └────────┘
  
  1. Client connects to AP
  2. AP creates a restricted port (no network access yet)
  3. Client authenticates via EAP through AP to RADIUS server
  4. RADIUS says "OK" → AP opens the port
  5. Unique encryption keys derived per user

EAP Methods:
  EAP-TLS:   Client certificate + server certificate (most secure)
  PEAP:      Server certificate + username/password (most common)
  EAP-TTLS:  Similar to PEAP (more flexible)

Benefits:
  - Individual credentials (revoke one user, not everyone)
  - Unique encryption keys per user
  - Centralized authentication (LDAP/AD integration)
  - No shared password that everyone knows
```

---

## Roaming and Handoff

### The roaming problem

```
Client moving between APs:

  ┌────┐         ┌────┐         ┌────┐
  │AP 1│         │AP 2│         │AP 3│
  └──┬─┘         └──┬─┘         └──┬─┘
     │               │              │
     ╰───BSS 1──╯    ╰──BSS 2──╯   ╰──BSS 3──╯
     
  Client walks: AP1 ───→ AP2 ───→ AP3

Basic roaming (slow):
  1. Signal from AP1 weakens
  2. Client scans for better AP (100-500ms!)
  3. Deauthenticate from AP1
  4. Authenticate with AP2
  5. Associate with AP2
  6. 4-way handshake with AP2
  → Total: 500ms-2s gap (noticeable for VoIP/video!)

Fast roaming (802.11r/k/v):
  802.11r (Fast Transition): Pre-authenticate with next AP
  802.11k (Radio Resource): AP tells client about neighbors
  802.11v (BSS Transition): AP can suggest "move to AP2 now"
  
  → Total: 50ms or less (acceptable for voice)
```

---

## Wi-Fi Performance Factors

### What determines actual throughput

```
Advertised "300 Mbps" vs reality:

  802.11n theoretical max: 300 Mbps
  
  Subtract:
    - Half duplex overhead:            -50%  (150 Mbps)
    - CSMA/CA overhead (backoff, ACK): -30%  (105 Mbps)
    - Management frames (beacons):     -5%   (100 Mbps)
    - Encryption overhead:             -5%   (95 Mbps)
    - Other clients sharing channel:   -50%  (47 Mbps)
    - Distance/wall attenuation:       -30%  (33 Mbps)
    
  Actual throughput: 30-50 Mbps (10-15% of advertised)
  This is NORMAL for Wi-Fi.
```

### Things that kill Wi-Fi performance

```
1. Co-channel interference (other APs on same channel)
   - In apartments: 20+ networks visible = disaster on 2.4 GHz
   - Fix: Use 5 GHz, or optimize channel selection

2. Too many clients on one AP
   - Wi-Fi is shared medium → 20 clients = 1/20th bandwidth each
   - Enterprise: max 25-30 clients per AP (for good performance)

3. Legacy clients (the slowest device slows EVERYONE)
   - One 802.11b client (11 Mbps) on the network
   - AP must use longer frame durations for compatibility
   - All other clients wait longer
   - Fix: Disable legacy rates on AP

4. Sticky clients (client holds onto weak AP)
   - Client connected to AP1 at -82 dBm
   - AP2 right there at -40 dBm
   - Client won't roam (implementation dependent)
   - Fix: AP-side minimum RSSI settings, 802.11v

5. Channel bonding in crowded environments
   - 80 MHz channel = 4× bandwidth but 4× interference surface
   - In apartments: use 20 MHz channels for reliability
   - 80/160 MHz only makes sense in controlled environments
```

---

## Cellular Networking Basics

### Cell tower architecture

```
         ┌──────────────────────────────────────────┐
         │            Core Network (EPC/5GC)         │
         │                                           │
         │  ┌─────┐  ┌─────┐  ┌─────┐  ┌────────┐  │
         │  │ MME │  │ SGW │  │ PGW │  │ HSS/   │  │
         │  │     │  │     │  │     │  │ UDM    │  │
         │  └──┬──┘  └──┬──┘  └──┬──┘  └────────┘  │
         └─────┼────────┼───────┼──────────────────┘
               │        │       │
         ┌─────┼────────┼───────┼──────┐  
         │     │   Backhaul (fiber)     │
         └─────┼────────┼───────┼──────┘
               │        │       │
          ┌────┴──┐ ┌───┴──┐ ┌─┴─────┐
          │Cell 1 │ │Cell 2│ │Cell 3  │   ← Base Stations (eNodeB/gNB)
          └───┬───┘ └──┬───┘ └───┬────┘
              │        │         │
          ────┴────────┴─────────┴────  Radio (air interface)
              │        │         │
           ┌──┴──┐  ┌──┴──┐  ┌──┴──┐
           │Phone│  │Phone│  │Phone│
           └─────┘  └─────┘  └─────┘

Key components:
  eNodeB/gNB:  Base station (the cell tower radio)
  MME:         Mobility Management Entity (tracks where you are)
  SGW:         Serving Gateway (data plane routing)
  PGW:         PDN Gateway (connection to internet)
  HSS/UDM:     Subscriber database (your SIM info)
```

---

## How Mobile Data Works (3G → 5G)

```
Generation   Technology   Typical Speed    Latency     Key Feature
──────────   ──────────   ─────────────    ───────     ───────────
3G           UMTS/HSPA    1-10 Mbps        100-200ms   Mobile internet
4G/LTE       OFDMA        10-100 Mbps      30-50ms     IP-based, all-data
4G LTE-A     CA + MIMO    100-300 Mbps     20-30ms     Carrier aggregation
5G (sub-6)   OFDMA+       100-900 Mbps     10-20ms     Network slicing
5G (mmWave)  Beamforming  1-10 Gbps        1-5ms       Ultra-low latency

Key technology jumps:
  3G → 4G:  Voice became IP (VoLTE), everything is data
  4G → 5G:  Network slicing (virtual networks for different needs)
            Ultra-low latency for IoT/autonomous vehicles
            Massive device density (1M devices/km²)

Carrier Aggregation (4G/5G):
  Combine multiple frequency bands simultaneously:
  Band 1 (10 MHz) + Band 3 (20 MHz) + Band 7 (15 MHz)
  = 45 MHz total → much higher throughput
```

---

## Mobile IP and Handover

### The mobility problem for IP

```
Problem: IP addresses are tied to location (subnet)

  You're connected to Tower A → IP: 10.1.1.50 (Tower A's subnet)
  You move to Tower B → Tower B is subnet 10.2.1.0/24
  Your IP 10.1.1.50 doesn't belong here!
  
  In wired networking: you'd need a new IP
  In mobile: can't drop every connection when you move!

Solution: GTP Tunnel (GPRS Tunneling Protocol)

  ┌───────┐     ┌─────────┐     ┌────────┐     ┌──────────┐
  │ Phone │────>│ Tower B │────>│  SGW   │────>│   PGW    │──> Internet
  └───────┘     └─────────┘     └────────┘     └──────────┘
                                                 Your IP is
                                                 anchored HERE
  
  PGW assigns you 10.x.x.x → stays same regardless of tower
  GTP tunnel carries your packets between tower and PGW
  When you move: tunnel endpoint changes, IP stays same
  
  This is why mobile connections survive tower changes!
```

### Handover types

```
Hard handover (3G):
  Disconnect from Tower A → Connect to Tower B
  Brief interruption (~100ms)

Soft handover (3G CDMA):
  Connected to Tower A AND B simultaneously
  Seamless transition
  
LTE handover:
  1. Phone measures neighbor cells (always scanning)
  2. Reports measurements to current tower
  3. Current tower decides: "hand off to Tower B"
  4. Tower A tells Tower B: "prepare for this user"
  5. Tower A tells phone: "switch to Tower B now"
  6. Phone switches → Tower B confirms → done
  
  Interruption: 0-50ms (usually imperceptible)
  Phone calls (VoLTE) survive handover seamlessly
```

---

## Troubleshooting Wireless Issues

### Common problems and diagnosis

```bash
# Linux: Check Wi-Fi connection details
iwconfig wlan0
# Look for: Signal level, Bit Rate, Link Quality

# Detailed info
iw dev wlan0 link
# Connected to AA:BB:CC:DD:EE:FF (BSSID)
# signal: -48 dBm   ← good
# tx bitrate: 866.7 MBit/s  ← negotiated speed

# Scan for available networks
sudo iw dev wlan0 scan | grep -E 'SSID|signal|freq'
# Shows all visible APs with signal strength

# Check for interference and channel usage
sudo iw dev wlan0 survey dump
# Shows noise levels and channel busy time

# Monitor mode (for deep wireless debugging)
sudo airmon-ng start wlan0
# Now can capture ALL wireless frames with Wireshark
```

### Diagnosis decision tree

```
Slow Wi-Fi?
├── Check signal strength (iwconfig / Wi-Fi analyzer)
│   ├── Weak signal (< -70 dBm)
│   │   → Move closer to AP, or add AP
│   └── Strong signal (> -50 dBm)
│       → Problem is not range
│
├── Check channel congestion
│   ├── Many APs on same channel
│   │   → Switch to less used channel (5 GHz preferred)
│   └── Few APs
│       → Not interference
│
├── Check negotiated speed
│   ├── Low speed despite good signal
│   │   → Legacy client problem or AP settings
│   └── High speed but slow throughput
│       → Too many clients or backhaul issue
│
├── Check for retransmissions
│   ├── High frame retry rate (> 10%)
│   │   → Hidden nodes, interference, or multipath
│   └── Low retry rate
│       → Problem is elsewhere (internet, server)
│
└── Check if wired connection is fast
    ├── Wired also slow → ISP/server issue
    └── Wired fast, Wi-Fi slow → Wi-Fi specific issue
```

### macOS and Windows diagnostics

```
macOS:
  Option+click Wi-Fi icon → detailed info:
    - PHY Mode, Channel, RSSI, Noise, Tx Rate
  
  Wireless Diagnostics:
    /System/Library/CoreServices/Applications/Wireless Diagnostics.app
    
Windows:
  netsh wlan show interfaces
    - Signal, channel, radio type, receive/transmit rate
  
  WiFi Analyzer app (Microsoft Store)
    - Visual channel map, signal strength over time
```

---

## Key Takeaways

1. **Wi-Fi is half duplex, shared medium** — only one device can transmit at a time, total bandwidth is shared among all clients
2. **CSMA/CA avoids rather than detects collisions** — this overhead, plus ACK requirement, is why Wi-Fi actual throughput is 10-15% of advertised speed
3. **Use 5 GHz when possible** — more channels, less interference. 2.4 GHz is overcrowded and only has 3 usable non-overlapping channels (1, 6, 11)
4. **Hidden node problem is real** — two clients hidden from each other cause collisions the AP has to deal with. RTS/CTS helps
5. **Signal strength isn't everything — SNR matters more** — a strong signal in a noisy environment (many APs, microwaves) is still bad
6. **WPA2-PSK is vulnerable to offline attack** — captured handshake can be brute-forced, WPA3-SAE fixes this with forward secrecy
7. **Enterprise 802.1X gives per-user credentials and keys** — essential for organizations, integrates with LDAP/AD
8. **Roaming is slow without 802.11r/k/v** — basic roaming takes 500ms-2s, fast transition (11r) reduces to under 50ms
9. **Legacy devices slow down everyone** — one old 802.11b client forces the AP to use compatibility mode, impacting all clients
10. **Cellular solves mobility with GTP tunnels** — your IP stays anchored at the PGW regardless of which tower you're connected to, making handover seamless

---

## Next Module

→ [Module 16: Network Virtualization](../16-network-virtualization/01-namespaces-veth.md) — Linux namespaces, veth pairs, bridges, and container networking
