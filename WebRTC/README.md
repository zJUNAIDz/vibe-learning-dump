# WebRTC Lessons

A comprehensive, systems-first guide to building production WebRTC applications.

## What You'll Build

By the end of this course, you'll be able to build:

1. **P2P File Transfer** - Send files directly between browsers using data channels
2. **1-on-1 Audio/Video Calls** - Production-ready video calling with fallbacks
3. **Group Calls with SFU** - Scalable multi-party video using mediasoup

## Who This Is For

- Full-stack developers who want to add real-time communication to their apps
- Engineers preparing for interviews at video-first companies (Zoom, Discord, etc.)
- Anyone tired of academic WebRTC explanations that don't ship code

## What Makes This Different

**Systems-first, not spec-first:**
- Explains *why* things work before *how*
- Focuses on production gotchas, not textbook theory
- Every concept includes failure modes and debugging strategies

**Code-heavy:**
- Complete implementations, not code snippets
- Production patterns (retry logic, error handling, monitoring)
- Real-world tradeoffs (cost, latency, complexity)

## Course Structure

### Foundations (00-03)
- [00 - Mental Model](00-mental-model.md) - WebRTC as "secure UDP pipe over signaling"
- [01 - Networking Primer](01-networking-primer.md) - NAT, firewalls, why P2P is hard
- [02 - Signaling](02-signaling.md) - SDP exchange and WebSocket implementation
- [03 - ICE/STUN/TURN](03-ice-stun-turn.md) - NAT traversal and relay servers

### Core WebRTC (04-06)
- [04 - Peer Connection](04-peer-connection-core.md) - RTCPeerConnection lifecycle
- [05 - Media Flow](05-media-flow.md) - getUserMedia, codecs, RTP/SRTP
- [06 - Data Channels](06-data-channels-file-transfer.md) - Non-media P2P transfer

### Building Applications (07-09)
- [07 - Building 1-on-1 Calls](07-building-1-1-call.md) - Complete call implementation
- [08 - Group Call Architecture](08-group-calls-architecture.md) - Mesh vs MCU vs SFU
- [09 - SFU Deep Dive](09-sfu-deep-dive.md) - mediasoup server implementation

### Production (10-13)
- [10 - Scaling & Production](10-scaling-and-prod-gotchas.md) - TURN costs, mobile issues
- [11 - Debugging](11-debugging-and-observability.md) - chrome://webrtc-internals, stats API
- [12 - Security & Privacy](12-security-and-privacy.md) - DTLS/SRTP, E2EE, threat model
- [13 - Common Mistakes](13-common-mistakes.md) - Errors that waste days

### Final Validation (14)
- [14 - Build Your Own Checklist](14-build-your-own-checklist.md) - Self-assessment + implementation checklists

## How to Use This Course

### Fast Path (8 hours)
Read files 00, 02, 04, 05, 07, 10, 13, 14. Build one project.

### Deep Path (20 hours)
Read all files in order. Build all three projects. Deploy to production.

### Interview Prep (4 hours)
Read files 00, 01, 03, 08, and section "Can You Explain...?" in file 14.

## Prerequisites

- Comfortable with JavaScript (async/await, promises, classes)
- Basic networking (what an IP address is, HTTP vs WebSocket)
- Node.js installed (for SFU server)

**You don't need:**
- Prior WebRTC experience
- Deep networking knowledge (we explain NAT/firewalls)
- C++ or codec expertise

## Learning Approach

Each file follows this structure:

1. **Mental Model** - High-level concept (why it exists)
2. **How It Works** - Technical details (what it does)
3. **Code Examples** - Production implementations (how to use it)
4. **Failure Modes** - What breaks and why
5. **Quick Self-Check** - Validate understanding before moving on

## Key Technologies

- **WebRTC APIs:** RTCPeerConnection, getUserMedia, RTCDataChannel
- **Signaling:** WebSocket (Socket.IO)
- **SFU:** mediasoup (Node.js + C++)
- **Debugging:** chrome://webrtc-internals, Stats API

## Projects Overview

### Project 1: P2P File Transfer
**Difficulty:** Beginner  
**Time:** 2-4 hours  
**Key Concepts:** Data channels, chunking, flow control

Transfer files up to 100 MB directly between browsers. Includes progress bars, error handling, and connection recovery.

### Project 2: 1-on-1 Call
**Difficulty:** Intermediate  
**Time:** 4-6 hours  
**Key Concepts:** Signaling, media tracks, ICE restart

Production-ready video calling with mute/unmute, camera switching, and network failure recovery.

### Project 3: Group Call with SFU
**Difficulty:** Advanced  
**Time:** 8-12 hours  
**Key Concepts:** SFU architecture, simulcast, active speaker detection

Scalable group calls for 5-10 participants using mediasoup. Includes bandwidth adaptation and quality selection.

## Cost Estimates (Production)

**Signaling Server:**
- $10-20/month (DigitalOcean/AWS t3.small)

**STUN Server:**
- Free (use Google's or self-host)

**TURN Server:**
- ~$0.08/GB for bandwidth
- ~10% of connections need TURN
- 200 concurrent users at 2 Mbps = $11,000/month
- (See file 10 for optimization strategies)

**SFU Server:**
- $50-200/month per instance (handles 10-20 concurrent rooms)

## Browser Support

| Browser | Version | Notes |
|---------|---------|-------|
| Chrome | 72+ | Full support |
| Firefox | 68+ | Full support |
| Safari | 14+ | getUserMedia requires user gesture |
| Edge | 79+ | Chromium-based, full support |
| iOS Safari | 14+ | Track cleanup issues (see file 10) |

## Common Gotchas

1. **Add tracks before createOffer** - Or SDP will be empty
2. **Handle null ICE candidate** - Last candidate is null
3. **Use TURN servers** - 10% of connections need relay
4. **Close peer connections** - Prevents memory leaks
5. **Use WebSocket for signaling** - HTTP too slow

(See file 13 for complete list)

## Performance Benchmarks

**P2P File Transfer:**
- 100 MB file in ~8 seconds on gigabit connection
- Limited by slower peer's bandwidth

**1-on-1 Call:**
- Connection time: 1-3 seconds (host/srflx), 3-5 seconds (relay)
- Latency: 50-100ms (direct), 100-200ms (TURN)

**SFU Group Call:**
- 5 participants: ~10 Mbps upload per user
- 10 participants with simulcast: ~5 Mbps upload per user
- Latency: 100-150ms (one hop to SFU)

## Troubleshooting

**Connection fails:**
1. Check ice servers in RTCPeerConnection config
2. Verify TURN credentials are valid
3. Check chrome://webrtc-internals for failed ICE checks

**No audio/video:**
1. Verify tracks added before createOffer
2. Check remote video element has srcObject set
3. Ensure tracks are not muted

**Poor quality:**
1. Check packet loss in stats API (should be <5%)
2. Verify bitrate adapting to available bandwidth
3. Consider enabling simulcast for group calls

(See file 11 for comprehensive debugging guide)

## Resources

**Official Specs:**
- [W3C WebRTC API](https://www.w3.org/TR/webrtc/)
- [RFC 8825: WebRTC Overview](https://datatracker.ietf.org/doc/html/rfc8825)

**Further Reading:**
- [WebRTC for the Curious](https://webrtcforthecurious.com/) - Free book, protocol-level
- [mediasoup Documentation](https://mediasoup.org/documentation/v3/)

**Libraries:**
- [simple-peer](https://github.com/feross/simple-peer) - Simplified P2P wrapper
- [mediasoup](https://github.com/versatica/mediasoup) - Production SFU
- [adapter.js](https://github.com/webrtcHacks/adapter) - Browser polyfills

## FAQ

**Q: Can I use this for production?**  
A: Yes. This course teaches production patterns (error handling, retry logic, monitoring). But test thoroughly in your network environment.

**Q: Do I need to understand the RFCs?**  
A: No. This course is practical, not academic. Read RFCs only if you're implementing a browser or debugging protocol issues.

**Q: Which SFU should I use?**  
A: mediasoup (Node.js) if you control the server. Otherwise, use a hosted service (Agora, Twilio, Livekit).

**Q: How do I handle 100+ participants?**  
A: Limit rendered videos to 16 tiles, rest audio-only. Use active speaker switching. Consider MCU for very large meetings (webinar mode).

**Q: Is WebRTC secure?**  
A: Yes - media is encrypted (SRTP). But signaling server can see metadata and become MITM. See file 12 for threat model.

**Q: Can I use this on mobile?**  
A: Yes. WebRTC works in mobile browsers. For native apps, use platform APIs (iOS: WebRTC framework, Android: WebRTC library).
