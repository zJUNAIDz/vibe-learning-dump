# 14 - Build Your Own: Validation Checklist

## You've Read Everything. Now What?

**This is your final exam** - not from me, from yourself.

Answer these questions without looking back. If you can't, review the relevant section.

---

## Foundational Understanding

### Can You Explain...?

**To a non-technical PM:**

- [ ] What WebRTC is in one sentence  
  _"A browser API that lets two users send audio/video/data directly to each other without going through your server"_

- [ ] Why you need a signaling server  
  _"Browsers can't discover each other automatically - signaling exchanges initial connection info (SDP, ICE candidates)"_

- [ ] What TURN servers do and when they're needed  
  _"Relay traffic when direct connection fails (~10% of cases due to firewalls/NAT)"_

**To another engineer:**

- [ ] The three-layer WebRTC model  
  _Application (signaling) → DTLS/SCTP (encryption/transport) → RTP/SRTP (media)_

- [ ] Offer/answer flow in correct order  
  _Alice: createOffer → setLocal → send → Bob: setRemote → createAnswer → setLocal → send → Alice: setRemote_

- [ ] Why symmetric NAT requires TURN  
  _Symmetric NAT assigns different external ports per destination - hole punching fails_

- [ ] How simulcast works  
  _Browser encodes video at 3 resolutions (low/mid/high), SFU forwards appropriate layer per subscriber_

- [ ] Why mesh doesn't scale beyond 4 participants  
  _Bandwidth grows O(N²) - each peer sends to N-1 others, receives from N-1 others_

---

## Technical Deep Dives

### ICE / STUN / TURN

- [ ] Explain candidate types (host, srflx, relay)  
  _Host = local IP, srflx = public IP via STUN, relay = TURN server IP_

- [ ] Calculate TURN bandwidth for 200 concurrent users at 2 Mbps each  
  _(200 users × 2 Mbps × 2 directions × 3600 seconds) / 8 = 175 GB/hour_

- [ ] When would you use `iceTransportPolicy: 'relay'`?  
  _Privacy-focused apps (hides real IP), or debugging TURN issues_

### Signaling

- [ ] Why you can't use HTTP polling for signaling  
  _Too slow - ICE candidates arrive within milliseconds, need real-time delivery_

- [ ] What trickle ICE is  
  _Send ICE candidates as they're discovered (incremental) instead of waiting for all_

- [ ] What perfect negotiation solves  
  _Prevents deadlock when both peers call createOffer simultaneously_

### Media

- [ ] What codec you'd use for audio in production  
  _Opus - best quality, adaptive bitrate, handles packet loss well_

- [ ] Difference between VP8, VP9, H.264  
  _VP8: royalty-free, older. VP9: better compression, slower encode. H.264: hardware-accelerated, patent-encumbered_

- [ ] What RTP does  
  _Transports media packets with timing info for synchronization_

- [ ] What causes jitter and how browsers handle it  
  _Network delay variance - jitter buffer holds packets to smooth playback_

### SFU Architecture

- [ ] Why SFU is better than mesh for 10+ participants  
  _Bandwidth: SFU is O(N), mesh is O(N²). CPU: SFU doesn't decode, just forwards_

- [ ] What SSRC is and why SFU rewrites it  
  _Synchronization Source ID - SFU assigns unique IDs to avoid conflicts_

- [ ] When to use active speaker detection  
  _Prioritize high-quality layer for active speaker, lower for others - saves bandwidth_

---

## Implementation Validation

### Project 1: P2P File Transfer

**Can you build this in 2 hours?**

**Requirements**:
- Connect two browsers (Alice, Bob)
- Alice selects file (up to 100 MB)
- Transfer via data channel
- Show progress bar on both sides
- Handle connection failure gracefully

**Checklist**:

- [ ] Signaling server (WebSocket)
- [ ] RTCPeerConnection setup with ICE servers
- [ ] Data channel creation (`ordered: true`, `maxRetransmits: null`)
- [ ] File chunking (16 KB chunks)
- [ ] Flow control (`bufferedAmount` monitoring)
- [ ] Progress calculation (bytes sent / total size)
- [ ] Reassembly on receiver side
- [ ] Error handling (connection failed, data channel closed)

**Key code snippets you should remember**:

```javascript
// Sender
const dc = pc.createDataChannel('file-transfer', {
  ordered: true
});

dc.onopen = async () => {
  const CHUNK_SIZE = 16 * 1024;
  for (let offset = 0; offset < file.size; offset += CHUNK_SIZE) {
    const chunk = file.slice(offset, offset + CHUNK_SIZE);
    const buffer = await chunk.arrayBuffer();
    
    while (dc.bufferedAmount > 16 * 1024 * 1024) {
      await new Promise(resolve => setTimeout(resolve, 100));
    }
    
    dc.send(buffer);
  }
};

// Receiver
pc.ondatachannel = (event) => {
  const dc = event.channel;
  const chunks = [];
  
  dc.onmessage = (event) => {
    chunks.push(event.data);
    // Update progress
  };
  
  dc.onclose = () => {
    const blob = new Blob(chunks);
    // Download file
  };
};
```

**Did you remember**:
- `bufferedAmount` check?
- Chunking at 16 KB?
- `ordered: true` for files?

---

### Project 2: 1-on-1 Audio/Video Call

**Can you build this in 4 hours?**

**Requirements**:
- Join room (enter room ID)
- Camera/mic preview before joining
- Connect to peer in same room
- Mute/unmute audio/video
- Switch camera (front/back)
- End call, cleanup resources

**Checklist**:

- [ ] SignalingClient class (offer/answer/ICE exchange)
- [ ] CallManager class (peer connection lifecycle)
- [ ] MediaDeviceManager class (getUserMedia, device enumeration)
- [ ] UI state management (joining, connected, disconnected)
- [ ] Track management (add before offer, replace for camera switch)
- [ ] ICE restart on failure
- [ ] Cleanup on hangup (stop tracks, close peer connection)

**Key code snippets you should remember**:

```javascript
// CallManager
class CallManager {
  constructor(signalingClient, localStream) {
    this.pc = new RTCPeerConnection({
      iceServers: [
        { urls: 'stun:stun.l.google.com:19302' },
        { urls: 'turn:...', username: '...', credential: '...' }
      ]
    });
    
    localStream.getTracks().forEach(track => {
      this.pc.addTrack(track, localStream);
    });
    
    this.pc.ontrack = (event) => {
      remoteVideo.srcObject = event.streams[0];
    };
    
    this.pc.onicecandidate = (event) => {
      if (event.candidate) {
        signalingClient.send({ type: 'ice-candidate', candidate: event.candidate });
      }
    };
    
    this.pc.oniceconnectionstatechange = () => {
      if (this.pc.iceConnectionState === 'failed') {
        this.restartICE();
      }
    };
  }
  
  async call() {
    const offer = await this.pc.createOffer();
    await this.pc.setLocalDescription(offer);
    signalingClient.send({ type: 'offer', sdp: offer });
  }
  
  async handleOffer(offer) {
    await this.pc.setRemoteDescription(offer);
    const answer = await this.pc.createAnswer();
    await this.pc.setLocalDescription(answer);
    signalingClient.send({ type: 'answer', sdp: answer });
  }
  
  async handleAnswer(answer) {
    await this.pc.setRemoteDescription(answer);
  }
  
  async handleICECandidate(candidate) {
    await this.pc.addIceCandidate(candidate);
  }
  
  async restartICE() {
    const offer = await this.pc.createOffer({ iceRestart: true });
    await this.pc.setLocalDescription(offer);
    signalingClient.send({ type: 'offer', sdp: offer });
  }
  
  hangup() {
    this.pc.getSenders().forEach(sender => {
      if (sender.track) sender.track.stop();
    });
    this.pc.close();
  }
}
```

**Did you remember**:
- Add tracks before `createOffer`?
- Handle ICE restart?
- Stop tracks before closing peer connection?
- TURN server in ICE config?

---

### Project 3: Group Call with SFU

**Can you build this in 8 hours?**

**Requirements**:
- Up to 10 participants
- Simulcast (3 layers)
- Active speaker detection
- Join/leave notifications
- Mute indicators
- Bandwidth-adaptive quality selection

**Checklist**:

**SFU Server (mediasoup)**:

- [ ] Room management (create/join/leave)
- [ ] WebRTC transport creation (send/receive per client)
- [ ] Producer/consumer model
- [ ] Simulcast enablement in router codec
- [ ] Active speaker detection (audio levels)
- [ ] Layer selection per consumer

**Client**:

- [ ] Signaling protocol (join-room, new-producer, etc.)
- [ ] Publish local stream (createProducer)
- [ ] Subscribe to remote streams (createConsumer)
- [ ] Simulcast parameters in send encoding
- [ ] UI grid layout (up to 10 participants)
- [ ] Cleanup on disconnect

**Key code snippets you should remember**:

**Server**:

```javascript
// Room.js
class Room {
  constructor(mediasoup.Router router, roomId) {
    this.router = router;
    this.roomId = roomId;
    this.peers = new Map(); // peerId -> { transports, producers, consumers }
  }
  
  async createWebRtcTransport(peerId, direction) {
    const transport = await this.router.createWebRtcTransport({
      listenIps: [{ ip: '0.0.0.0', announcedIp: 'YOUR_PUBLIC_IP' }],
      enableUdp: true,
      enableTcp: true,
      preferUdp: true
    });
    
    this.peers.get(peerId)[direction + 'Transport'] = transport;
    
    return {
      id: transport.id,
      iceParameters: transport.iceParameters,
      iceCandidates: transport.iceCandidates,
      dtlsParameters: transport.dtlsParameters
    };
  }
  
  async produce(peerId, transportId, kind, rtpParameters) {
    const transport = this.peers.get(peerId).sendTransport;
    const producer = await transport.produce({ kind, rtpParameters });
    
    this.peers.get(peerId).producers.push(producer);
    
    // Notify others
    this.broadcast({ type: 'new-producer', producerId: producer.id, peerId }, peerId);
    
    return producer.id;
  }
  
  async consume(consumerPeerId, producerId) {
    const producer = this.findProducer(producerId);
    const transport = this.peers.get(consumerPeerId).receiveTransport;
    
    const consumer = await transport.consume({
      producerId: producer.id,
      rtpCapabilities: this.peers.get(consumerPeerId).rtpCapabilities,
      paused: true // Start paused, resume after client ready
    });
    
    // Prefer spatial layer 2 (high quality) initially
    if (consumer.type === 'simulcast') {
      await consumer.setPreferredLayers({ spatialLayer: 2, temporalLayer: 2 });
    }
    
    return {
      id: consumer.id,
      producerId: producer.id,
      kind: consumer.kind,
      rtpParameters: consumer.rtpParameters
    };
  }
}
```

**Client**:

```javascript
// SFUClient.js
class SFUClient {
  async publish(track) {
    const transport = this.sendTransport;
    
    const encodings = track.kind === 'video' ? [
      { rid: 'r0', maxBitrate: 100000, scaleResolutionDownBy: 4 },
      { rid: 'r1', maxBitrate: 300000, scaleResolutionDownBy: 2 },
      { rid: 'r2', maxBitrate: 900000 }
    ] : undefined;
    
    const producer = await transport.produce({
      track,
      encodings,
      codecOptions: {
        videoGoogleStartBitrate: 1000
      }
    });
    
    return producer;
  }
  
  async subscribe(producerId) {
    const response = await this.signaling.request('consume', {
      producerId
    });
    
    const consumer = await this.receiveTransport.consume({
      id: response.id,
      producerId: response.producerId,
      kind: response.kind,
      rtpParameters: response.rtpParameters
    });
    
    await this.signaling.request('resume-consumer', { consumerId: consumer.id });
    
    return consumer.track;
  }
}
```

**Did you remember**:
- Simulcast encoding parameters (3 layers)?
- Router codec configuration (enable simulcast)?
- Start consumer paused, resume after ready?
- Active speaker → switch to high layer?

---

## Production Readiness Checklist

Before deploying to production:

### Reliability

- [ ] TURN server configured with time-limited credentials
- [ ] ICE restart on connection failure
- [ ] Exponential backoff on signaling reconnection
- [ ] Circuit breaker pattern for repeated failures
- [ ] Connection timeout (10-15 seconds max)
- [ ] Fallback to audio-only on repeated video failures

### Performance

- [ ] Simulcast enabled for 3+ participants
- [ ] Active speaker detection (switch to high layer)
- [ ] Limit max concurrent videos (e.g., 9 tiles)
- [ ] Lazy-load remote streams (subscribe on demand)
- [ ] Monitor CPU via `qualityLimitationReason`
- [ ] Adaptive bitrate based on available bandwidth

### Observability

- [ ] Log all state changes (ICE, connection, signaling)
- [ ] Send metrics to backend (packet loss, RTT, bitrate)
- [ ] Capture diagnostics on error (stats, connection state)
- [ ] "Export Diagnostics" button for users
- [ ] Alert on high packet loss (>5%) or RTT (>200ms)

### Security

- [ ] HTTPS for signaling server
- [ ] Time-limited TURN credentials (generated on backend)
- [ ] Content Security Policy headers
- [ ] DTLS/SRTP verification (check cipher in stats)
- [ ] No hardcoded credentials in client code
- [ ] Optional: Fingerprint verification for high-security

### User Experience

- [ ] Camera/mic preview before joining call
- [ ] "Connecting..." indicator during ICE checks
- [ ] Mute/unmute buttons (with visual feedback)
- [ ] Network quality indicator (green/yellow/red)
- [ ] Graceful degradation (audio-only fallback)
- [ ] Clear error messages ("Connection failed - check your network")

### Browser Support

- [ ] Test on Chrome, Firefox, Safari, Edge
- [ ] Handle Safari quirks (user gesture for getUserMedia)
- [ ] Handle iOS quirks (stop tracks explicitly)
- [ ] Detect unsupported browsers, show message
- [ ] Polyfill for older browsers (adapter.js)

### Monitoring (Backend)

- [ ] Call success rate (connected / attempted)
- [ ] TURN usage percentage
- [ ] Average connection time (offer → connected)
- [ ] P95 packet loss
- [ ] P95 RTT
- [ ] Regional performance breakdown

---

## Architecture Decision Matrix

### When to Use What

| Use Case | Architecture | Why |
|----------|-------------|-----|
| **2-person call** | Mesh (P2P) | Simple, low latency, no server cost |
| **3-4 people, short call** | Mesh | Acceptable bandwidth (<10 Mbps) |
| **5-10 people** | SFU | Mesh bandwidth too high (24+ Mbps) |
| **10-50 people** | SFU with simulcast | Adaptive quality per subscriber |
| **50+ people (webinar)** | SFU + active speaker switching | Max 9-16 tiles, rest audio-only |
| **Screen share only** | Mesh (if <5) or SFU | Lower bandwidth than video |
| **File transfer** | Data channel (P2P) | No server sees data, E2EE |

---

## Final Reality Check

### Can you answer a PM's questions?

**"Why can't we just use HTTP for everything?"**  
_"HTTP is request/response, not real-time. WebRTC uses UDP for low-latency media. If a video packet is lost, skip it - don't wait for retransmit like TCP does."_

**"How much will this cost for 10,000 users?"**  
_"Depends on TURN usage. If 10% use TURN at 2 Mbps for 30 min calls, that's ~175 GB/hour. At $0.08/GB, ~$14/hour or $10k/month. Most calls (<90%) use direct connection = free."_

**"Can we support 100 people in a call?"**  
_"SFU can handle it, but clients can't render 100 videos. Limit to 16 tiles, rest audio-only. Use active speaker switching to show whoever's talking."_

**"Do we need to worry about security?"**  
_"Media is encrypted (SRTP). But signaling server can see metadata and become MITM. For high-security, implement fingerprint verification or E2EE SDP."_

**"What if user's network is bad?"**  
_"WebRTC adapts bitrate automatically. Show quality indicator. Offer audio-only fallback. Use TURN if direct connection fails."_

---

## Where to Go From Here

### You've completed the masterclass. What's next?

**Build one of the three projects** (file transfer, 1-1 call, group call).

**Read the specs** (if you're curious):
- [RFC 8825: WebRTC Overview](https://datatracker.ietf.org/doc/html/rfc8825)
- [RFC 8829: JSEP (Offer/Answer)](https://datatracker.ietf.org/doc/html/rfc8829)
- [ICE (RFC 8445)](https://datatracker.ietf.org/doc/html/rfc8445)

**Deep dives**:
- [WebRTC for the Curious](https://webrtcforthecurious.com/) - Excellent free book
- [mediasoup documentation](https://mediasoup.org/documentation/v3/) - Best SFU library
- [Pion WebRTC](https://github.com/pion/webrtc) - Golang implementation (learn by reading)

**Stay updated**:
- [W3C WebRTC Working Group](https://www.w3.org/groups/wg/webrtc/)
- [discuss-webrtc Google Group](https://groups.google.com/g/discuss-webrtc)

---

## Final Thought

**You now know more about WebRTC than 95% of developers.**

The last 5% comes from building, breaking, and debugging in production.

Go build something. Good luck.

---

## Completion Self-Check

Answer these honestly:

- [ ] I can explain WebRTC to a PM without using jargon
- [ ] I can build a P2P file transfer in <2 hours
- [ ] I can build a 1-1 call in <4 hours
- [ ] I can integrate with an SFU (mediasoup) in <8 hours
- [ ] I know when to use mesh vs SFU
- [ ] I can calculate TURN bandwidth costs
- [ ] I can debug connection failures using chrome://webrtc-internals
- [ ] I can implement ICE restart on network change
- [ ] I understand DTLS, SRTP, and the security model
- [ ] I know the top 10 common mistakes and how to avoid them

**If you checked 8+: You're ready for production.**  
**If you checked 5-7: Build one of the projects, then reassess.**  
**If you checked <5: Re-read the sections you're weakest on.**

---

## You're Done

This is the end of the masterclass.

You have everything you need to build production-ready WebRTC applications.

Now go build.
