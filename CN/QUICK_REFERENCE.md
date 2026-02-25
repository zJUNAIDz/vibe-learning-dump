# Quick Reference — Networking Commands & Concepts

> Keep this open while working through the modules.

---

## Layer-by-Layer Quick Tools

| Layer | What to Check | Tool | Example |
|-------|--------------|------|---------|
| Physical | Interface up/down | `ip link show` | `ip link show eth0` |
| Data Link | MAC, ARP | `ip neigh`, `arp` | `ip neigh show` |
| Network | IP, routing | `ip addr`, `ip route` | `ip route get 8.8.8.8` |
| Transport | Ports, connections | `ss` | `ss -tuln` |
| Application | HTTP, DNS | `curl`, `dig` | `curl -v example.com` |

---

## Essential Commands

### See your interfaces
```bash
ip -br addr show           # Brief view of all interfaces
ip addr show eth0           # Detailed view of one interface
```

### See your routes
```bash
ip route show               # All routes
ip route get 10.0.0.1       # Which route would be used for this IP?
```

### See open ports / connections
```bash
ss -tuln                    # Listening TCP/UDP ports
ss -tp                      # Established TCP connections with process names
ss -ti                      # TCP connections with internal info (cwnd, rtt)
```

### DNS
```bash
dig example.com             # Full DNS query
dig +short example.com      # Just the answer
dig @8.8.8.8 example.com    # Query specific DNS server
nslookup example.com        # Simpler DNS lookup
```

### Packet capture
```bash
sudo tcpdump -i eth0 -nn                        # All traffic on eth0
sudo tcpdump -i any port 80 -nn                  # HTTP traffic
sudo tcpdump -i any host 10.0.0.1 -nn            # Traffic to/from specific host
sudo tcpdump -i any -w capture.pcap              # Save to file for Wireshark
```

### Connectivity
```bash
ping -c 4 8.8.8.8           # Basic reachability (ICMP)
traceroute 8.8.8.8           # Path to destination
mtr 8.8.8.8                  # Combined ping + traceroute (live)
```

### HTTP
```bash
curl -v https://example.com  # Verbose HTTP request (shows TLS, headers)
curl -o /dev/null -s -w '%{time_total}\n' https://example.com  # Just timing
```

### Network namespaces
```bash
sudo ip netns list                              # List namespaces
sudo ip netns exec <ns> ip addr show            # Run command in namespace
```

---

## TCP State Reference

| State | Meaning |
|-------|---------|
| LISTEN | Waiting for incoming connections |
| SYN-SENT | Client sent SYN, waiting for SYN-ACK |
| SYN-RECEIVED | Server received SYN, sent SYN-ACK |
| ESTABLISHED | Connection open, data flowing |
| FIN-WAIT-1 | Sent FIN, waiting for ACK |
| FIN-WAIT-2 | Received ACK of FIN, waiting for peer's FIN |
| CLOSE-WAIT | Received FIN, waiting for application to close |
| TIME-WAIT | Waiting to ensure peer received final ACK (2*MSL) |
| CLOSING | Both sides sent FIN simultaneously |
| LAST-ACK | Sent FIN after receiving FIN, waiting for ACK |
| CLOSED | Connection fully closed |

---

## Common Ports

| Port | Protocol | Service |
|------|----------|---------|
| 22 | TCP | SSH |
| 53 | TCP/UDP | DNS |
| 80 | TCP | HTTP |
| 443 | TCP | HTTPS |
| 3306 | TCP | MySQL |
| 5432 | TCP | PostgreSQL |
| 6379 | TCP | Redis |
| 8080 | TCP | HTTP (alt) |
| 8443 | TCP | HTTPS (alt) |

---

## Subnet Cheat Sheet

| CIDR | Subnet Mask | Hosts | Notes |
|------|-------------|-------|-------|
| /32 | 255.255.255.255 | 1 | Single host |
| /31 | 255.255.255.254 | 2 | Point-to-point link |
| /30 | 255.255.255.252 | 2 | Point-to-point (traditional) |
| /28 | 255.255.255.240 | 14 | Small subnet |
| /24 | 255.255.255.0 | 254 | Standard "Class C" |
| /16 | 255.255.0.0 | 65,534 | Large network |
| /8 | 255.0.0.0 | 16M+ | Huge block |

---

## RFC Private Address Ranges

| Range | CIDR | Typical Use |
|-------|------|-------------|
| 10.0.0.0 – 10.255.255.255 | 10.0.0.0/8 | Large enterprise, cloud VPCs |
| 172.16.0.0 – 172.31.255.255 | 172.16.0.0/12 | Docker default, medium orgs |
| 192.168.0.0 – 192.168.255.255 | 192.168.0.0/16 | Home networks |
