# 11 - Debugging and Observability

## Chrome WebRTC Internals: Your Best Friend

**The most powerful WebRTC debugging tool**: `chrome://webrtc-internals`

Open it in a tab, then start your WebRTC app in another tab. You'll see:

- All RTCPeerConnections
- Complete SDP offer/answer
- ICE candidates (all types, states)
- Stats graphs (bitrate, packet loss, RTT)
- Event timeline

---

## Reading webrtc-internals

### Connection Overview

```
RTCPeerConnection(id=123)
  State: connected
  ICE connection state: connected
  Signaling state: stable
  ICE gathering state: complete
```

### SDP Section

```
createOffer:
  v=0
  o=- 123456789 2 IN IP4 127.0.0.1
  ...
  m=video 9 UDP/TLS/RTP/SAVPF 96 97
  a=rtpmap:96 VP8/90000
  
setLocalDescription: success
```

**Check**:
- Is SDP valid?
- Are codecs negotiated?
- Are ICE credentials present?

### ICE Candidates

```
Candidate: host 192.168.1.45:54321 (local)
Candidate: srflx 203.0.113.50:61234 (STUN)
Candidate: relay 198.51.100.75:3478 (TURN)

Selected pair:
  Local: srflx 203.0.113.50:61234
  Remote: srflx 198.51.100.20:61001
  State: succeeded
```

**What to look for**:
- Are TURN candidates present? (Should be if configured)
- Which pair was selected? (Host = direct, srflx = STUN, relay = TURN)
- Are checks failing? (State = failed)

### Stats Graphs

Hover over graphs to see:
- **Bytes sent/received**: Bandwidth usage
- **Packets lost**: Network quality
- **RTT**: Round-trip time

**Red flags**:
- Bitrate drops to zero = connection dead
- Packet loss >5% = poor quality
- RTT >200ms = high latency

---

## Stats API: Programmatic Monitoring

### Getting All Stats

```javascript
const stats = await pc.getStats();

stats.forEach(report => {
  console.log('Report type:', report.type);
  console.log('Report ID:', report.id);
  console.log('Data:', report);
});
```

### Key Report Types

#### 1. inbound-rtp (Receiving)

```javascript
stats.forEach(report => {
  if (report.type === 'inbound-rtp' && report.kind === 'video') {
    console.log('Packets received:', report.packetsReceived);
    console.log('Packets lost:', report.packetsLost);
    console.log('Bytes received:', report.bytesReceived);
    console.log('Jitter (ms):', report.jitter * 1000);
    console.log('Frame rate:', report.framesPerSecond);
    
    const lossRate = report.packetsLost / report.packetsReceived;
    if (lossRate > 0.05) {
      console.warn('High packet loss:', (lossRate * 100).toFixed(1) + '%');
    }
  }
});
```

#### 2. outbound-rtp (Sending)

```javascript
if (report.type === 'outbound-rtp' && report.kind === 'video') {
  console.log('Packets sent:', report.packetsSent);
  console.log('Bytes sent:', report.bytesSent);
  console.log('Frame rate:', report.framesPerSecond);
  console.log('Frames encoded:', report.framesEncoded);
  
  // Quality limiting factor
  console.log('Quality limitation reason:', report.qualityLimitationReason);
  // Possible values: 'none', 'cpu', 'bandwidth', 'other'
}
```

#### 3. candidate-pair (Connection)

```javascript
if (report.type === 'candidate-pair' && report.state === 'succeeded') {
  console.log('RTT (ms):', report.currentRoundTripTime * 1000);
  console.log('Bytes sent:', report.bytesSent);
  console.log('Bytes received:', report.bytesReceived);
  console.log('Priority:', report.priority);
  
  if (report.currentRoundTripTime > 0.2) {
    console.warn('High RTT:', report.currentRoundTripTime * 1000, 'ms');
  }
}
```

#### 4. track (Media Track)

```javascript
if (report.type === 'track' && report.kind === 'video') {
  console.log('Frame width:', report.frameWidth);
  console.log('Frame height:', report.frameHeight);
  console.log('Frames received:', report.framesReceived);
  console.log('Frames dropped:', report.framesDropped);
  console.log('Frames decoded:', report.framesDecoded);
  
  if (report.framesDropped > 0) {
    const dropRate = report.framesDropped / report.framesReceived;
    console.warn('Frame drop rate:', (dropRate * 100).toFixed(1) + '%');
  }
}
```

---

## Building a Real-Time Dashboard

```javascript
class ConnectionMonitor {
  constructor(pc) {
    this.pc = pc;
    this.metrics = {
      videoInbound: {},
      videoOutbound: {},
      audioInbound: {},
      audioOutbound: {},
      connection: {}
    };
  }
  
  async update() {
    const stats = await this.pc.getStats();
    
    stats.forEach(report => {
      switch(report.type) {
        case 'inbound-rtp':
          if (report.kind === 'video') {
            this.metrics.videoInbound = {
              packetsLost: report.packetsLost,
              packetsReceived: report.packetsReceived,
              lossRate: report.packetsLost / (report.packetsReceived || 1),
              bitrate: this.calculateBitrate(report),
              jitter: report.jitter,
              fps: report.framesPerSecond
            };
          } else {
            this.metrics.audioInbound = {
              packetsLost: report.packetsLost,
              packetsReceived: report.packetsReceived,
              jitter: report.jitter
            };
          }
          break;
          
        case 'outbound-rtp':
          if (report.kind === 'video') {
            this.metrics.videoOutbound = {
              packetsSent: report.packetsSent,
              bitrate: this.calculateBitrate(report),
              fps: report.framesPerSecond,
              qualityLimitation: report.qualityLimitationReason
            };
          }
          break;
          
        case 'candidate-pair':
          if (report.state === 'succeeded') {
            this.metrics.connection = {
              rtt: report.currentRoundTripTime * 1000,
              bytesSent: report.bytesSent,
              bytesReceived: report.bytesReceived,
              localType: report.localCandidateType,
              remoteType: report.remoteCandidateType
            };
          }
          break;
      }
    });
    
    return this.metrics;
  }
  
  calculateBitrate(report) {
    const now = Date.now();
    const bytes = report.bytesReceived || report.bytesSent || 0;
    
    if (this.lastReport) {
      const timeDelta = (now - this.lastTimestamp) / 1000;
      const bytesDelta = bytes - this.lastBytes;
      const bitrate = (bytesDelta * 8) / timeDelta / 1000; // kbps
      
      this.lastReport = report;
      this.lastTimestamp = now;
      this.lastBytes = bytes;
      
      return bitrate.toFixed(0);
    }
    
    this.lastReport = report;
    this.lastTimestamp = now;
    this.lastBytes = bytes;
    
    return 0;
  }
  
  startMonitoring(callback, interval = 1000) {
    setInterval(async () => {
      const metrics = await this.update();
      callback(metrics);
    }, interval);
  }
}

// Usage
const monitor = new ConnectionMonitor(pc);

monitor.startMonitoring((metrics) => {
  console.log('Video loss rate:', (metrics.videoInbound.lossRate * 100).toFixed(1) + '%');
  console.log('Video bitrate:', metrics.videoInbound.bitrate, 'kbps');
  console.log('RTT:', metrics.connection.rtt, 'ms');
  
  // Update UI
  document.getElementById('stats').innerHTML = `
    Bitrate: ${metrics.videoInbound.bitrate} kbps<br>
    Loss: ${(metrics.videoInbound.lossRate * 100).toFixed(1)}%<br>
    RTT: ${metrics.connection.rtt.toFixed(0)} ms
  `;
});
```

---

## Common Issues and How to Diagnose

### Issue: Video Freezes

**Symptoms**: Video stops updating, audio continues

**Diagnosis**:
1. Check `framesReceived` in stats → Is it increasing?
2. Check `framesDecoded` → Decoder stuck?
3. Check `packetsLost` → Network issue?

**Common causes**:
- Key frame lost → Request PLI (Picture Loss Indication)
- Decoder crash → Browser bug, restart stream
- CPU overload → Check `qualityLimitationReason: 'cpu'`

**Fix**:
```javascript
// Request key frame
const receiver = pc.getReceivers().find(r => r.track.kind === 'video');
// Browser requests keyframe automatically via RTCP PLI

// Or replace track to force refresh
await sender.replaceTrack(newTrack);
```

### Issue: One-Way Audio

**Symptoms**: Alice hears Bob, Bob doesn't hear Alice

**Diagnosis**:
1. Check `chrome://webrtc-internals` → Are tracks added?
2. Check stats → Is `outbound-rtp` present for audio?
3. Check signaling → Did Bob receive Alice's SDP?

**Common causes**:
- Track not added before createOffer
- Track muted (`track.enabled = false`)
- Firewall blocking one direction

**Fix**:
```javascript
// Verify tracks added
console.log('Senders:', pc.getSenders().map(s => ({
  kind: s.track?.kind,
  enabled: s.track?.enabled
})));
```

### Issue: High Latency (>1s)

**Symptoms**: Noticeable delay between speech and hearing

**Diagnosis**:
1. Check RTT in stats → Should be <200ms
2. Check jitter → Should be <30ms
3. Check if using TURN → Adds 20-100ms

**Common causes**:
- Jitter buffer too large (browser adapts automatically)
- TURN server geographically distant
- Network congestion

**Fix**:
```javascript
// Use nearest TURN server
const closestTURN = getClosestTURNServer(userLocation);

// Monitor RTT
if (rtt > 200) {
  showLatencyWarning();
  // Consider degrading quality
}
```

### Issue: ICE Failed

**Symptoms**: `iceConnectionState = 'failed'`

**Diagnosis**:
1. Check `chrome://webrtc-internals` → Are candidates gathered?
2. Are TURN candidates present?
3. Are all ICE checks failing?

**Common causes**:
- No TURN server configured
- TURN credentials invalid
- Firewall blocking all connection attempts

**Fix**:
```javascript
// Verify TURN working
pc.onicecandidate = (event) => {
  if (event.candidate) {
    console.log('Candidate type:', event.candidate.type);
    // Should see: host, srflx, relay
  }
};

// If no relay candidates, TURN broken
```

---

## Logging Best Practices

### Structured Logging

```javascript
class WebRTCLogger {
  constructor(sessionId) {
    this.sessionId = sessionId;
  }
  
  log(level, event, data = {}) {
    const logEntry = {
      timestamp: new Date().toISOString(),
      sessionId: this.sessionId,
      level,
      event,
      ...data
    };
    
    console.log(JSON.stringify(logEntry));
    
    // Send to logging service
    this.sendToServer(logEntry);
  }
  
  iceCandidate(candidate) {
    this.log('debug', 'ice_candidate', {
      type: candidate.type,
      protocol: candidate.protocol,
      address: candidate.address,
      port: candidate.port
    });
  }
  
  connectionStateChange(state) {
    this.log('info', 'connection_state_change', {
      iceConnectionState: pc.iceConnectionState,
      connectionState: pc.connectionState,
      signalingState: pc.signalingState
    });
  }
  
  statsSnapshot(metrics) {
    this.log('info', 'stats_snapshot', metrics);
  }
}

// Usage
const logger = new WebRTCLogger(generateSessionId());

pc.onicecandidate = (e) => {
  if (e.candidate) logger.iceCandidate(e.candidate);
};

pc.onconnectionstatechange = () => {
  logger.connectionStateChange();
};
```

### What to Log

**DO log**:
- ICE state changes
- Connection state changes
- Errors
- Stats snapshots (every 10s)
- User actions (mute, camera flip, hangup)

**DON'T log**:
- SDP (too large, contains IPs)
- RTP packets (too frequent, privacy concern)
- ICE passwords (security)

---

## Remote Debugging

### Collecting Diagnostics from Users

```javascript
async function collectDiagnostics() {
  const diagnostics = {
    sessionId: currentSessionId,
    timestamp: new Date().toISOString(),
    browser: navigator.userAgent,
    
    // Connection states
    iceConnectionState: pc.iceConnectionState,
    connectionState: pc.connectionState,
    signalingState: pc.signalingState,
    
    // Candidates
    localCandidates: [],
    remoteCandidates: [],
    
    // Stats
    stats: await pc.getStats().then(stats => {
      const result = [];
      stats.forEach(report => result.push(report));
      return result;
    }),
    
    // Tracks
    senders: pc.getSenders().map(s => ({
      kind: s.track?.kind,
      enabled: s.track?.enabled,
      readyState: s.track?.readyState
    })),
    
    receivers: pc.getReceivers().map(r => ({
      kind: r.track.kind,
      enabled: r.track.enabled,
      readyState: r.track.readyState
    }))
  };
  
  // Download as JSON
  const blob = new Blob([JSON.stringify(diagnostics, null, 2)], {
    type: 'application/json'
  });
  const url = URL.createObjectURL(blob);
  
  const a = document.createElement('a');
  a.href = url;
  a.download = `webrtc-diagnostics-${Date.now()}.json`;
  a.click();
  
  URL.revokeObjectURL(url);
}

// Add button to UI
document.getElementById('export-diagnostics').onclick = collectDiagnostics;
```

**When user reports issue**: "Click Export Diagnostics and send us the file"

---

## Alerts and Thresholds

```javascript
class QualityMonitor {
  constructor(pc, callbacks) {
    this.pc = pc;
    this.callbacks = callbacks;
    this.thresholds = {
      packetLoss: 0.05,      // 5%
      rtt: 200,              // ms
      jitter: 30,            // ms
      frameDropRate: 0.05    // 5%
    };
  }
  
  async checkQuality() {
    const stats = await this.pc.getStats();
    
    stats.forEach(report => {
      if (report.type === 'inbound-rtp' && report.kind === 'video') {
        const lossRate = report.packetsLost / report.packetsReceived;
        
        if (lossRate > this.thresholds.packetLoss) {
          this.callbacks.onHighPacketLoss(lossRate);
        }
        
        if (report.jitter * 1000 > this.thresholds.jitter) {
          this.callbacks.onHighJitter(report.jitter * 1000);
        }
      }
      
      if (report.type === 'candidate-pair' && report.state === 'succeeded') {
        const rtt = report.currentRoundTripTime * 1000;
        
        if (rtt > this.thresholds.rtt) {
          this.callbacks.onHighRTT(rtt);
        }
      }
    });
  }
  
  startMonitoring(interval = 5000) {
    setInterval(() => this.checkQuality(), interval);
  }
}

// Usage
const monitor = new QualityMonitor(pc, {
  onHighPacketLoss: (rate) => {
    console.warn('High packet loss:', rate);
    showWarning('Network quality degraded');
  },
  
  onHighRTT: (rtt) => {
    console.warn('High RTT:', rtt);
    showWarning('High latency detected');
  },
  
  onHighJitter: (jitter) => {
    console.warn('High jitter:', jitter);
  }
});

monitor.startMonitoring();
```

---

## Production Monitoring Stack

```javascript
// Send metrics to your monitoring service
class MetricsReporter {
  constructor(sessionId) {
    this.sessionId = sessionId;
  }
  
  async reportMetrics(metrics) {
    await fetch('/api/metrics', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        sessionId: this.sessionId,
        timestamp: Date.now(),
        metrics: {
          // Connection quality
          packet_loss: metrics.videoInbound.lossRate,
          rtt: metrics.connection.rtt,
          jitter: metrics.videoInbound.jitter,
          
          // Bandwidth
          video_bitrate_in: metrics.videoInbound.bitrate,
          video_bitrate_out: metrics.videoOutbound.bitrate,
          
          // Performance
          video_fps: metrics.videoInbound.fps,
          quality_limitation: metrics.videoOutbound.qualityLimitation,
          
          // Connection type
          connection_type: `${metrics.connection.localType}-${metrics.connection.remoteType}`
        }
      })
    });
  }
}
```

**Aggregate in backend** (DataDog, Grafana, etc.):
- Average packet loss across all calls
- P95 RTT
- TURN usage percentage
- Call failure rate

---

## What You Must Understand

| Concept | Why It Matters |
|---------|----------------|
| **chrome://webrtc-internals** | First stop for debugging |
| **Stats API** | Programmatic health monitoring |
| **Packet loss >5% = bad** | Visible quality degradation |
| **RTT >200ms = noticeable** | Lag in conversation |
| **Log structured data** | Essential for debugging at scale |

---

## Next Steps

You now have the tools to debug and monitor WebRTC in production.

**Next**: [12-security-and-privacy.md](12-security-and-privacy.md) - Security considerations and encryption.

Debugging gets you working. Security keeps you safe.

---

## Quick Self-Check

- [ ] Navigate chrome://webrtc-internals
- [ ] Extract key metrics from Stats API
- [ ] Build real-time metrics dashboard
- [ ] Diagnose common issues (freeze, one-way audio, ICE failed)
- [ ] Implement quality monitoring with alerts
- [ ] Collect diagnostics from remote users
- [ ] Send metrics to backend monitoring

If you can debug a call failure from logs alone, you're ready for production.
