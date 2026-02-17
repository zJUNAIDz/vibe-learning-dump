# 13 - Common Mistakes

## The Mistakes That Cost You Days

These are the errors that look innocent but destroy hours of debugging time.

---

## Category 1: Offer/Answer Mistakes

### Mistake: Creating Offer Before Adding Tracks

**What you did**:

```javascript
const pc = new RTCPeerConnection(config);

const offer = await pc.createOffer(); // ❌ No tracks yet!

const stream = await navigator.mediaDevices.getUserMedia({ video: true });
stream.getTracks().forEach(track => pc.addTrack(track, stream));
```

**Why it breaks**: The offer SDP contains no media sections (`m=video`). Bob receives an offer with no tracks, his answer also has no tracks, and the connection works but **no media flows**.

**Symptoms**: Connection succeeds (`iceConnectionState = 'connected'`), but no video/audio.

**Fix**:

```javascript
// Add tracks FIRST
const stream = await navigator.mediaDevices.getUserMedia({ video: true });
stream.getTracks().forEach(track => pc.addTrack(track, stream));

// THEN create offer
const offer = await pc.createOffer();
```

**Remember**: Offer describes what you're sending. If nothing added, offer is empty.

---

### Mistake: Not Handling Null ICE Candidate

**What you did**:

```javascript
pc.onicecandidate = (event) => {
  signalingClient.send({
    type: 'ice-candidate',
    candidate: event.candidate // ❌ Can be null!
  });
};
```

**Why it breaks**: The last `icecandidate` event has `event.candidate = null` (signals end of gathering). Bob tries to parse `null` as a candidate, crashes.

**Fix**:

```javascript
pc.onicecandidate = (event) => {
  if (event.candidate) {
    signalingClient.send({
      type: 'ice-candidate',
      candidate: event.candidate
    });
  }
  // null candidate means gathering complete - no need to send
};
```

---

### Mistake: Forgetting setRemoteDescription

**What you did**:

```javascript
// Alice sends offer
const offer = await pc.createOffer();
await pc.setLocalDescription(offer);
signalingClient.send(offer);

// Bob receives offer
signalingClient.on('offer', async (offer) => {
  // ❌ Forgot to set remote description!
  const answer = await pc.createAnswer();
  await pc.setLocalDescription(answer);
  signalingClient.send(answer);
});
```

**Why it breaks**: Bob creates an answer without knowing what Alice offered. `createAnswer()` throws error.

**Fix**:

```javascript
signalingClient.on('offer', async (offer) => {
  await pc.setRemoteDescription(offer); // ✅ Set remote FIRST
  const answer = await pc.createAnswer();
  await pc.setLocalDescription(answer);
  signalingClient.send(answer);
});
```

**Correct order**:
1. Alice: `createOffer` → `setLocalDescription`
2. Bob: `setRemoteDescription` → `createAnswer` → `setLocalDescription`
3. Alice: `setRemoteDescription`

---

### Mistake: Renegotiating Without Polite/Impolite Pattern

**Scenario**: Both Alice and Bob call `pc.createOffer()` simultaneously.

**What happens**: Signaling states collide, connection breaks.

**Fix**: Use **perfect negotiation**:

```javascript
let makingOffer = false;
let ignoreOffer = false;
const polite = isPolite(localUserId, remoteUserId); // Deterministic

pc.onnegotiationneeded = async () => {
  try {
    makingOffer = true;
    await pc.setLocalDescription();
    signalingClient.send({ type: 'offer', sdp: pc.localDescription });
  } finally {
    makingOffer = false;
  }
};

signalingClient.on('offer', async (offer) => {
  const offerCollision = makingOffer || pc.signalingState !== 'stable';
  ignoreOffer = !polite && offerCollision;
  
  if (ignoreOffer) return; // Impolite peer ignores offer
  
  await pc.setRemoteDescription(offer);
  const answer = await pc.createAnswer();
  await pc.setLocalDescription(answer);
  signalingClient.send({ type: 'answer', sdp: pc.localDescription });
});
```

**Result**: One peer always wins collisions.

---

## Category 2: ICE and Connection Mistakes

### Mistake: No TURN Server Configured

**What you did**:

```javascript
const pc = new RTCPeerConnection({
  iceServers: [
    { urls: 'stun:stun.l.google.com:19302' }
  ]
});
```

**Why it breaks**: ~10% of connections fail (symmetric NAT, corporate firewalls). Users report "connection stuck".

**Fix**:

```javascript
const pc = new RTCPeerConnection({
  iceServers: [
    { urls: 'stun:stun.l.google.com:19302' },
    { 
      urls: 'turn:turn.example.com:3478',
      username: 'user',
      credential: 'pass'
    }
  ]
});
```

**When to see impact**:
- Corporate networks
- Mobile CGNAT
- Symmetric NAT routers

---

### Mistake: Hardcoding TURN Credentials

**What you did**:

```javascript
const pc = new RTCPeerConnection({
  iceServers: [{
    urls: 'turn:turn.example.com:3478',
    username: 'admin',
    credential: 'supersecret' // ❌ In client code!
  }]
});
```

**Why it breaks**:
- Anyone can inspect source, steal credentials
- Attackers use your TURN server for free
- Your AWS bill explodes

**Fix**: Generate time-limited credentials on backend (see [12-security-and-privacy.md](12-security-and-privacy.md)).

---

### Mistake: Not Handling ICE Restart

**Scenario**: User switches from Wi-Fi to cellular.

**What you did**: Nothing.

**Result**: Connection dies, user sees frozen video.

**Fix**:

```javascript
// Detect network change
window.addEventListener('online', restartICE);

pc.oniceconnectionstatechange = () => {
  if (pc.iceConnectionState === 'failed') {
    restartICE();
  }
};

async function restartICE() {
  const offer = await pc.createOffer({ iceRestart: true });
  await pc.setLocalDescription(offer);
  signalingClient.send({ type: 'offer', sdp: offer });
}
```

---

## Category 3: Media Mistakes

### Mistake: Not Checking Track State

**What you did**:

```javascript
const sender = pc.getSenders()[0];
await sender.replaceTrack(newTrack); // ❌ What if sender is null?
```

**Why it breaks**: No senders exist if tracks weren't added. Throws error.

**Fix**:

```javascript
const videoSender = pc.getSenders().find(s => s.track?.kind === 'video');

if (videoSender) {
  await videoSender.replaceTrack(newTrack);
} else {
  console.error('No video sender found');
}
```

---

### Mistake: Closing Track Instead of Replacing

**What you did**:

```javascript
// Switch camera
const oldTrack = stream.getVideoTracks()[0];
oldTrack.stop(); // ❌ Stops sending video!

const newStream = await navigator.mediaDevices.getUserMedia({ 
  video: { deviceId: newCameraId }
});
```

**Why it breaks**: `stop()` removes the track from peer connection. Remote peer sees black screen.

**Fix**:

```javascript
const newStream = await navigator.mediaDevices.getUserMedia({ 
  video: { deviceId: newCameraId }
});
const newTrack = newStream.getVideoTracks()[0];

const sender = pc.getSenders().find(s => s.track?.kind === 'video');
await sender.replaceTrack(newTrack); // ✅ Swap track

// Now safe to stop old track
oldTrack.stop();
```

---

### Mistake: Not Handling getUserMedia Failure

**What you did**:

```javascript
const stream = await navigator.mediaDevices.getUserMedia({ 
  video: true, 
  audio: true 
});
// ❌ What if user denies permission?
```

**Why it breaks**: Throws `NotAllowedError`. App crashes.

**Fix**:

```javascript
try {
  const stream = await navigator.mediaDevices.getUserMedia({ 
    video: true, 
    audio: true 
  });
} catch (err) {
  if (err.name === 'NotAllowedError') {
    alert('Camera/mic permission denied');
  } else if (err.name === 'NotFoundError') {
    alert('No camera/mic found');
  } else {
    console.error('getUserMedia error:', err);
  }
}
```

---

## Category 4: SDP Mistakes

### Mistake: Modifying SDP Without Understanding

**What you did**:

```javascript
let offer = await pc.createOffer();

// "I read online to do this"
offer.sdp = offer.sdp.replace('useinbandfec=1', 'useinbandfec=0');

await pc.setLocalDescription(offer);
```

**Why it breaks**: Might work for you, breaks for others. SDP is fragile.

**When to modify**:
- Prefer codecs (use `setCodecPreferences`)
- Force bandwidth limits (use sender parameters)

**Last resort**: Understand what you're changing.

---

### Mistake: Not Waiting for setRemoteDescription

**What you did**:

```javascript
pc.setRemoteDescription(offer); // ❌ Not awaited!

const answer = await pc.createAnswer(); // Might use stale state
```

**Why it breaks**: `setRemoteDescription` is async. `createAnswer` might run before it completes.

**Fix**:

```javascript
await pc.setRemoteDescription(offer); // ✅ Wait

const answer = await pc.createAnswer();
```

---

## Category 5: Data Channel Mistakes

### Mistake: Not Checking bufferedAmount

**What you did**:

```javascript
// Send 100 MB file
for (const chunk of fileChunks) {
  dc.send(chunk); // ❌ Floods buffer
}
```

**Why it breaks**: `bufferedAmount` exceeds threshold (16 MB), WebRTC drops packets.

**Fix**:

```javascript
async function sendFile(dc, chunks) {
  for (const chunk of chunks) {
    while (dc.bufferedAmount > 16 * 1024 * 1024) {
      await new Promise(resolve => setTimeout(resolve, 100));
    }
    dc.send(chunk);
  }
}
```

---

### Mistake: Not Handling Data Channel Events

**What you did**:

```javascript
const dc = pc.createDataChannel('chat');
dc.send('Hello'); // ❌ Channel not open yet!
```

**Why it breaks**: Data channel must be `open` before sending.

**Fix**:

```javascript
const dc = pc.createDataChannel('chat');

dc.onopen = () => {
  dc.send('Hello'); // ✅ Wait for open
};

dc.onerror = (err) => {
  console.error('Data channel error:', err);
};
```

---

## Category 6: State Management Mistakes

### Mistake: Not Listening to Connection States

**What you did**:

```javascript
const pc = new RTCPeerConnection(config);
// ❌ No state listeners!
```

**Why it breaks**: Connection fails, user stares at loading spinner forever.

**Fix**:

```javascript
pc.oniceconnectionstatechange = () => {
  console.log('ICE state:', pc.iceConnectionState);
  
  if (pc.iceConnectionState === 'failed') {
    handleConnectionFailure();
  }
};

pc.onconnectionstatechange = () => {
  console.log('Connection state:', pc.connectionState);
  
  if (pc.connectionState === 'failed') {
    handleConnectionFailure();
  }
};
```

---

### Mistake: Leaking Peer Connections

**What you did**:

```javascript
function startCall() {
  const pc = new RTCPeerConnection(config);
  // Call ends, pc never closed
}
```

**Why it breaks**: Memory leak. After 10 calls, browser slows down.

**Fix**:

```javascript
let currentPC = null;

function startCall() {
  if (currentPC) {
    currentPC.close();
  }
  
  currentPC = new RTCPeerConnection(config);
}

function endCall() {
  if (currentPC) {
    currentPC.close();
    currentPC = null;
  }
}
```

---

## Category 7: SFU Client Mistakes

### Mistake: Not Handling Simulcast Downgrade

**Scenario**: Network slows down, SFU switches from high to low quality.

**What you did**: Assume high quality always available.

**Fix**:

```javascript
// SFU tells you current layer
socket.on('quality-changed', ({ layer }) => {
  if (layer === 'low') {
    showNotification('Quality reduced due to network');
  }
});
```

---

### Mistake: Not Requesting Key Frames

**Scenario**: New participant joins, receives corrupted video (missing I-frame).

**Fix**:

```javascript
// When new participant joins
await fetch('/sfu/request-keyframe', {
  method: 'POST',
  body: JSON.stringify({ producerId: remoteProducerId })
});
```

---

## Category 8: Architecture Mistakes

### Mistake: Using Database for ICE Candidates

**What you did**:

```javascript
pc.onicecandidate = async (event) => {
  if (event.candidate) {
    await fetch('/api/ice-candidate', {
      method: 'POST',
      body: JSON.stringify(event.candidate)
    });
  }
};
```

**Why it breaks**:
- ICE candidates arrive in milliseconds
- Database writes are slow (50-100ms)
- Candidates arrive late, connection fails

**Fix**: Use WebSocket, not HTTP.

---

### Mistake: Single TURN Server for Global App

**What you did**: TURN server in US-East-1 only.

**Why it breaks**:
- User in Australia → 200ms RTT to TURN
- Noticeable lag

**Fix**: Regional TURN servers, route by user location.

---

### Mistake: No Retry Logic for Signaling

**What you did**: WebSocket disconnects, app freezes.

**Fix**:

```javascript
class SignalingClient {
  constructor(url) {
    this.url = url;
    this.reconnectAttempts = 0;
    this.connect();
  }
  
  connect() {
    this.ws = new WebSocket(this.url);
    
    this.ws.onclose = () => {
      const delay = Math.min(1000 * 2 ** this.reconnectAttempts, 30000);
      setTimeout(() => {
        this.reconnectAttempts++;
        this.connect();
      }, delay);
    };
    
    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
    };
  }
}
```

---

## The "I Didn't Know That Could Happen" Mistakes

### Apple Safari: Requires User Gesture for getUserMedia

**Breaks**: `getUserMedia` called on page load.

**Fix**: Call inside button click handler.

```javascript
document.getElementById('join-call').addEventListener('click', async () => {
  const stream = await navigator.mediaDevices.getUserMedia({ video: true });
});
```

---

### iOS: Doesn't Release Tracks on PeerConnection.close()

**Breaks**: Camera stays on (green indicator) after call ends.

**Fix**:

```javascript
function endCall() {
  // Stop tracks BEFORE closing peer connection
  pc.getSenders().forEach(sender => {
    if (sender.track) {
      sender.track.stop();
    }
  });
  
  pc.close();
}
```

---

### Firefox: Doesn't Support Unified Plan by Default (Old Versions)

**Breaks**: Renegotiation fails with cryptic errors.

**Fix**: Enforce `sdpSemantics` (though deprecated):

```javascript
const pc = new RTCPeerConnection({
  sdpSemantics: 'unified-plan'
});
```

**Better**: Require modern browser versions.

---

## What You Must Understand

| Mistake | Impact | Fix Complexity |
|---------|--------|----------------|
| **Offer before tracks** | No media flows | Easy |
| **Null ICE candidate** | Crashes | Easy |
| **No TURN server** | 10% connection failures | Medium |
| **No ICE restart** | Breaks on network switch | Medium |
| **Not awaiting setRemoteDescription** | Rare race condition | Easy |
| **Data channel floods** | Packet loss | Medium |
| **Leaking peer connections** | Memory leak | Easy |
| **Database for ICE candidates** | Slow connection | Hard |

---

## Debug Checklist

When something breaks, check:

1. [ ] Are tracks added before `createOffer`?
2. [ ] Are you checking for `null` in `onicecandidate`?
3. [ ] Are you awaiting `setRemoteDescription`?
4. [ ] Do you have a TURN server configured?
5. [ ] Are you handling getUserMedia errors?
6. [ ] Are you listening to connection state changes?
7. [ ] Are you closing peer connections when done?
8. [ ] Are you using WebSocket (not HTTP) for signaling?

---

## Next Steps

You now know the landmines. Avoid them.

**Next**: [14-build-your-own-checklist.md](14-build-your-own-checklist.md) - Validate your understanding.

Mistakes are inevitable. Knowing the common ones saves time.

---

## Quick Self-Check

- [ ] Create offer AFTER adding tracks
- [ ] Handle null ICE candidate
- [ ] Add TURN server to ICE config
- [ ] Implement ICE restart on network change
- [ ] Replace tracks instead of stopping them
- [ ] Close peer connections to prevent leaks
- [ ] Use WebSocket for signaling (not HTTP)
- [ ] Handle browser-specific quirks (Safari, iOS)

If you've made at least 5 of these mistakes, you're learning. If you haven't, you will.
