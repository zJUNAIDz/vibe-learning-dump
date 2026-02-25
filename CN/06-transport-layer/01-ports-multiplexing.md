# Ports & Multiplexing — How Multiple Applications Share One Network Connection

> Your laptop has one IP address but runs dozens of network-connected applications simultaneously — a browser with 20 tabs, Spotify, Slack, SSH sessions, Docker containers. Ports make this possible.

---

## Table of Contents

1. [The Problem: One IP, Many Applications](#the-problem)
2. [What Ports Are](#what-ports-are)
3. [Port Number Ranges](#port-number-ranges)
4. [Sockets: The Complete Address](#sockets)
5. [The 5-Tuple: Identifying a Connection](#the-5-tuple)
6. [Port Exhaustion](#port-exhaustion)
7. [Linux: Working with Ports](#linux-working-with-ports)

---

## The Problem

IP delivers packets to a **host** (identified by IP address). But the host runs many applications. When a packet arrives for 10.0.0.5, which application should receive it?

Without ports, you'd need a separate IP address for every application. With 30 browser tabs, 3 SSH sessions, and 5 background services, you'd need 38 IP addresses on one machine. Clearly unworkable.

**Ports solve this**: They add a 16-bit number to the address, creating a sub-address for each application. IP delivers to the host. The port delivers to the specific application.

```
Without ports: IP packet → Host → ???
With ports:    IP packet → Host → Port 443 → nginx
                                  Port 22  → sshd
                                  Port 5432 → postgres
```

---

## What Ports Are

A port is a **16-bit unsigned integer** (0–65535) that identifies a specific process/service on a host.

Both TCP and UDP use ports independently. TCP port 80 and UDP port 80 are different.

### How ports appear in headers

```
TCP Header:
  Source Port:      49152     (client's ephemeral port)
  Destination Port: 443       (server's well-known port)

UDP Header:
  Source Port:      54321
  Destination Port: 53        (DNS)
```

---

## Port Number Ranges

| Range | Name | Description |
|-------|------|-------------|
| 0–1023 | **Well-known / System** | Assigned to standard services. Require root on Linux. |
| 1024–49151 | **Registered** | Assigned by IANA for specific applications. No root needed. |
| 49152–65535 | **Ephemeral / Dynamic** | Used temporarily by clients for outgoing connections |

### Common well-known ports (MEMORIZE these)

| Port | Protocol | Service |
|------|----------|---------|
| 20/21 | TCP | FTP (data/control) |
| 22 | TCP | SSH |
| 23 | TCP | Telnet (insecure, deprecated) |
| 25 | TCP | SMTP (email sending) |
| 53 | TCP/UDP | DNS |
| 67/68 | UDP | DHCP (server/client) |
| 80 | TCP | HTTP |
| 110 | TCP | POP3 (email retrieval) |
| 143 | TCP | IMAP (email retrieval) |
| 443 | TCP | HTTPS |
| 465/587 | TCP | SMTP with TLS |
| 993 | TCP | IMAP with TLS |
| 3306 | TCP | MySQL |
| 5432 | TCP | PostgreSQL |
| 6379 | TCP | Redis |
| 8080 | TCP | HTTP alternative |
| 8443 | TCP | HTTPS alternative |
| 27017 | TCP | MongoDB |

### Why root is needed for ports < 1024

Historical Unix convention: only root (or processes with `CAP_NET_BIND_SERVICE` capability) can bind to ports below 1024. The idea was that if you connect to port 22, you can trust that a system administrator set up that service (not a random user).

This is largely a convention today, not a real security mechanism. But it persists.

```bash
# Fails (without root)
python3 -m http.server 80
# PermissionError: [Errno 13] Permission denied

# Works (with root)
sudo python3 -m http.server 80

# Or grant capability (better than running as root)
sudo setcap 'cap_net_bind_service=+ep' /usr/bin/python3
```

---

## Sockets

A **socket** is the combination of an IP address and a port:

$$\text{Socket} = \text{IP Address} : \text{Port}$$

Example: `10.0.0.5:443`

A socket uniquely identifies an endpoint in a network conversation.

### Server sockets vs client sockets

**Server socket**: Opens a socket on a well-known port and **listens** for connections:
```
nginx: 0.0.0.0:443 (listening on all interfaces, port 443)
sshd:  0.0.0.0:22  (listening on all interfaces, port 22)
```

**Client socket**: Opens a socket on an ephemeral port (OS-assigned) and **connects**:
```
curl: 192.168.1.100:52341 → google.com:443
ssh:  192.168.1.100:52342 → server.com:22
```

---

## The 5-Tuple

A network connection is uniquely identified by 5 values:

```
┌─────────────────────────────────────────────────────────────┐
│  Source IP  │ Source Port │ Protocol │ Dest IP  │ Dest Port │
│ 10.0.0.5   │ 49152      │ TCP      │ 1.2.3.4  │ 443       │
└─────────────────────────────────────────────────────────────┘
```

**Every combination is a unique connection**. This means:

- The SAME client can have MULTIPLE connections to the SAME server (different source ports)
- The SAME server port can handle MILLIONS of simultaneous connections (each from a different source IP:port)

```bash
# Open 3 connections to google.com:443 — each gets a different source port
curl -sI https://google.com &
curl -sI https://google.com &
curl -sI https://google.com &

ss -tn | grep 443
# ESTAB  0  0  192.168.1.100:49152  142.250.x.x:443
# ESTAB  0  0  192.168.1.100:49153  142.250.x.x:443
# ESTAB  0  0  192.168.1.100:49154  142.250.x.x:443
```

### How nginx handles millions of connections on one port

nginx listens on port 443. When a client connects, the OS creates a new socket for that specific 5-tuple. nginx doesn't need a separate port per client:

```
Client A: 1.2.3.4:50000 → nginx:443    → Socket A
Client B: 5.6.7.8:50000 → nginx:443    → Socket B
Client C: 1.2.3.4:50001 → nginx:443    → Socket C

All three are different 5-tuples → different sockets → no conflict
```

---

## Port Exhaustion

### The problem

Ephemeral ports are 16-bit. The typical range is 32768–60999 on Linux (about 28,000 ports). If a single client opens more than 28,000 connections to the same destination IP:port, it runs out of source ports.

### When this happens

- **High-traffic proxies/load balancers**: A reverse proxy making requests to a backend server can exhaust ports
- **NAT gateways**: All traffic from an internal network shares the NAT's IP, sharing the ephemeral port space
- **Microservices**: A service making frequent short HTTP calls to another service

### Detection

```bash
# Check ephemeral port range
cat /proc/sys/net/ipv4/ip_local_port_range
# 32768   60999

# Count connections in TIME_WAIT (these hold ports)
ss -s
# Shows TCP socket summary including TIME_WAIT count

ss -tn state time-wait | wc -l
# If this is approaching your port range size, you have a problem
```

### Mitigation

```bash
# Expand ephemeral port range
sudo sysctl -w net.ipv4.ip_local_port_range="1024 65535"
# Now ~64,000 ports available

# Enable TIME_WAIT reuse (allows reusing ports in TIME_WAIT for new connections to the same destination)
sudo sysctl -w net.ipv4.tcp_tw_reuse=1

# Use connection pooling in your applications (don't open/close connections rapidly)
```

---

## Linux: Working with Ports

### Viewing listening ports

```bash
# Show all listening TCP and UDP ports
ss -tulnp
# -t: TCP
# -u: UDP  
# -l: listening (server) sockets
# -n: numeric (don't resolve names)
# -p: show process name

# Example output:
# State  Recv-Q Send-Q Local Address:Port  Peer Address:Port Process
# LISTEN 0      128    0.0.0.0:22          0.0.0.0:*         users:(("sshd",pid=1234))
# LISTEN 0      511    0.0.0.0:80          0.0.0.0:*         users:(("nginx",pid=5678))

# With netstat (older)
netstat -tulnp
```

### Viewing established connections

```bash
# All established TCP connections
ss -tn
# Shows source:port → dest:port for each connection

# Filter by port
ss -tn sport = :22
# All SSH connections (source port 22 = outgoing)
ss -tn dport = :443
# All HTTPS connections (destination port 443)

# Count connections per destination port
ss -tn | awk '{print $5}' | cut -d: -f2 | sort | uniq -c | sort -rn | head
```

### Checking if a port is in use

```bash
# Check if port 8080 is in use
ss -tln | grep 8080

# Or more directly
lsof -i :8080
# Shows which process has the port open

# Find what process is on port 80
sudo fuser 80/tcp
# Returns PID
```

### Testing port connectivity

```bash
# Can I reach port 443 on google.com?
nc -zv google.com 443
# Connection to google.com 443 port [tcp/https] succeeded!

# With timeout
nc -zv -w 3 google.com 443

# Using curl (for HTTP/S)
curl -v --connect-timeout 5 https://google.com 2>&1 | head -20
```

---

## Key Takeaways

1. **Ports multiplex applications** — one IP address supports 65,536 ports per protocol
2. **The 5-tuple uniquely identifies a connection** — (src IP, src port, protocol, dst IP, dst port)
3. **Servers listen on well-known ports** (0-1023). **Clients use ephemeral ports** (49152-65535)
4. **A server on one port can handle millions of connections** — each client has a unique 5-tuple
5. **Port exhaustion** is a real problem for high-connection systems — mitigate with pooling and TIME_WAIT reuse
6. **`ss` is the modern tool** for viewing sockets and ports on Linux

---

## Next

→ [02-udp.md](02-udp.md) — The simplest transport protocol
