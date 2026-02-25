# Computer Networks — Deep Curriculum

> A complete, deep, A→Z Computer Networks curriculum.
> Written as lecture notes + lab manual — not a textbook summary.

---

## Who This Is For

- CS graduate who studied networks in college
- Has basic–intermediate familiarity
- Wants full revision, deep intuition, practical debugging skills
- Uses Linux regularly
- NOT a beginner. NOT a PhD researcher.

---

## Directory Structure

```
CN/
├── README.md                  ← You are here
├── START_HERE.md               ← Orientation & how to use this curriculum
├── GETTING_STARTED.md          ← Environment setup
├── QUICK_REFERENCE.md          ← Cheat sheet for tools & commands
│
├── 00-orientation/             ← What a network really is
│   └── 01-what-is-a-network.md
│
├── 01-network-models/          ← OSI, TCP/IP, encapsulation, reality
│   ├── 01-osi-model.md
│   ├── 02-tcpip-model.md
│   └── 03-encapsulation-and-reality.md
│
├── 02-metrics/                 ← Latency, bandwidth, jitter, loss
│   ├── 01-latency-deep-dive.md
│   ├── 02-bandwidth-throughput.md
│   └── 03-jitter-loss-measurement.md
│
├── 03-physical-datalink/       ← Ethernet, MAC, switching, ARP
│   ├── 01-signals-and-encoding.md
│   ├── 02-ethernet-framing.md
│   ├── 03-mac-switching-flooding.md
│   └── 04-arp-deep-dive.md
│
├── 04-ip-layer/                ← IPv4, CIDR, fragmentation, TTL
│   ├── 01-ipv4-addressing.md
│   ├── 02-cidr-subnetting.md
│   ├── 03-ip-header-fragmentation.md
│   └── 04-ttl-and-unreliability.md
│
├── 05-routing-forwarding/      ← Tables, BGP, OSPF, longest prefix match
│   ├── 01-routing-tables-forwarding.md
│   ├── 02-static-dynamic-routing.md
│   ├── 03-ospf-conceptual.md
│   └── 04-bgp-conceptual.md
│
├── 06-transport-layer/         ← TCP, UDP, handshake, flow control
│   ├── 01-ports-multiplexing.md
│   ├── 02-udp.md
│   ├── 03-tcp-handshake-teardown.md
│   ├── 04-tcp-reliability-retransmission.md
│   └── 05-tcp-flow-control.md
│
├── 07-congestion-control/      ← AIMD, slow start, cwnd, bufferbloat
│   ├── 01-congestion-collapse.md
│   ├── 02-aimd-slow-start-cwnd.md
│   └── 03-bufferbloat-and-beyond.md
│
├── 08-dns/                     ← Resolution, caching, failures, debugging
│   ├── 01-dns-resolution-deep-dive.md
│   ├── 02-dns-caching-ttl.md
│   └── 03-dns-failures-debugging.md
│
├── 09-application-layer/       ← HTTP/1.1, HTTP/2, HTTP/3, WebSockets
│   ├── 01-http1-deep-dive.md
│   ├── 02-http2.md
│   ├── 03-http3-quic.md
│   └── 04-websockets.md
│
├── 10-tls-security/            ← TLS handshake, certs, trust, limits
│   ├── 01-why-encryption.md
│   ├── 02-tls-handshake.md
│   └── 03-certificates-trust-limits.md
│
├── 11-nat-firewalls/           ← NAT, firewalls, proxies, load balancers
│   ├── 01-nat-types-deep-dive.md
│   ├── 02-firewalls-stateful-stateless.md
│   └── 03-proxies-load-balancers-middleboxes.md
│
├── 12-linux-internals/         ← Sockets, FDs, kernel stack, epoll
│   ├── 01-sockets-and-fds.md
│   ├── 02-kernel-networking-stack.md
│   └── 03-buffers-epoll-select.md
│
├── 13-networking-tools/        ← ip, ss, ping, traceroute, tcpdump, curl, nc
│   ├── 01-ip-and-ss.md
│   ├── 02-ping-traceroute.md
│   ├── 03-tcpdump.md
│   └── 04-curl-netcat.md
│
├── 14-packet-analysis/         ← tcpdump, Wireshark, reading TCP streams
│   ├── 01-tcpdump-deep-dive.md
│   ├── 02-wireshark.md
│   └── 03-reading-tcp-streams-detecting-issues.md
│
├── 15-wireless-mobile/         ← Wi-Fi, reliability, cellular
│   └── 01-wireless-and-mobile.md
│
├── 16-network-virtualization/  ← Namespaces, veth, bridges, overlays
│   ├── 01-namespaces-veth.md
│   ├── 02-bridges-overlays.md
│   └── 03-docker-kubernetes-networking.md
│
├── 17-debugging-methodology/   ← Systematic debugging, reasoning
│   └── 01-systematic-network-debugging.md
│
├── 18-failure-scenarios/       ← DNS outages, NAT exhaustion, MTU, etc.
│   └── 01-real-world-failures.md
│
├── 19-capstone-follow-packet/  ← Follow a packet end-to-end
│   └── 01-follow-a-packet.md
│
└── 20-cloud-networking/        ← VPCs, subnets, IGW, SGs, LBs
    ├── 01-sdn-and-vpc-concepts.md
    ├── 02-subnets-routing-gateways.md
    └── 03-security-groups-nacls-load-balancers.md
```

---

## How to Use

1. Read [START_HERE.md](START_HERE.md) for orientation
2. Set up your environment with [GETTING_STARTED.md](GETTING_STARTED.md)
3. Go through modules **in order** — they build on each other
4. Every module has Linux commands — **run them**
5. Keep [QUICK_REFERENCE.md](QUICK_REFERENCE.md) open as a cheat sheet

---

## Philosophy

- **Depth over breadth** — every concept is explained slowly
- **Why before what** — understand the problem before the solution
- **Theory + practice** — every concept has Linux commands
- **No jargon without explanation** — every term is defined
- **Debug-first thinking** — learn to reason about failures
