# The TLS Handshake — Step by Step

> The TLS handshake is the most important security protocol on the internet. Every HTTPS connection, every secure API call, every email submission begins with this handshake. Understanding it step-by-step is essential for debugging connection failures, certificate errors, and performance issues.

---

## Table of Contents

1. [TLS History](#history)
2. [TLS 1.2 Handshake](#tls12)
3. [TLS 1.3 Handshake](#tls13)
4. [Key Derivation](#key-derivation)
5. [Cipher Suites](#cipher-suites)
6. [TLS Record Protocol](#record-protocol)
7. [Session Resumption](#resumption)
8. [0-RTT (Early Data)](#zero-rtt)
9. [TLS Termination](#termination)
10. [Common TLS Failures](#failures)
11. [Linux: TLS Debugging](#linux-debugging)
12. [Key Takeaways](#key-takeaways)

---

## History

```
1995: SSL 2.0 (Netscape) — first public version. Deeply flawed.
1996: SSL 3.0 — significant redesign. Broken by POODLE (2014).
1999: TLS 1.0 (RFC 2246) — SSL 3.0 rename + fixes. Broken by BEAST.
2006: TLS 1.1 (RFC 4346) — fixed BEAST-type attacks. Deprecated 2021.
2008: TLS 1.2 (RFC 5246) — still widely used. Secure with right config.
2018: TLS 1.3 (RFC 8446) — major overhaul: faster, simpler, more secure.

As of 2024:
  TLS 1.3: ~65% of connections (preferred)
  TLS 1.2: ~34% of connections (still supported)
  TLS 1.0/1.1: <1% (deprecated, browsers reject)
  SSL *:  Completely dead
```

---

## TLS 1.2 Handshake

TLS 1.2 takes **2 round trips** (4 messages) before encrypted data can flow:

```
Client                                          Server
  │                                                │
  │──── ClientHello ──────────────────────────→     │  RTT 1
  │     (version, random, cipher suites,           │
  │      extensions, SNI)                          │
  │                                                │
  │     ←───────────────────── ServerHello ─────    │
  │       (version, random, chosen cipher,         │
  │        session ID)                             │
  │     ←───────────────────── Certificate ─────   │
  │       (server's X.509 certificate chain)       │
  │     ←───────────────────── ServerKeyExchange ── │
  │       (DH parameters, signed by server)        │
  │     ←───────────────────── ServerHelloDone ──── │
  │                                                │
  │──── ClientKeyExchange ────────────────────→     │  RTT 2
  │     (client's DH public value)                 │
  │──── ChangeCipherSpec ─────────────────────→     │
  │     ("switching to encrypted mode")            │
  │──── Finished ─────────────────────────────→     │
  │     (HMAC of all handshake messages)           │
  │                                                │
  │     ←───────────────────── ChangeCipherSpec ──  │
  │     ←───────────────────── Finished ────────── │
  │                                                │
  │════ Encrypted application data ═══════════════ │
```

### Step-by-step breakdown

**1. ClientHello**

```
Client → Server:
  Protocol version: TLS 1.2
  Client random: 32 bytes of random data (used in key derivation)
  Session ID: empty (new connection) or cached ID (resumption)
  Cipher suites: [
    TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
    TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
    TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
    ...
  ]
  Compression: null (compression removed due to CRIME attack)
  Extensions:
    server_name (SNI): "example.com"    ← which site (virtual hosting)
    supported_groups: [x25519, secp256r1, secp384r1]
    signature_algorithms: [rsa_pss_rsae_sha256, ecdsa_secp256r1_sha256]
    alpn: [h2, http/1.1]               ← protocol negotiation
```

**2. ServerHello**

```
Server → Client:
  Protocol version: TLS 1.2
  Server random: 32 bytes
  Session ID: new or resumed
  Chosen cipher: TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
  Compression: null
  Extensions:
    alpn: h2    ← "we'll use HTTP/2"
```

**3. Certificate**

```
Server → Client:
  Certificate chain:
    [0] Server cert: CN=example.com (signed by intermediate CA)
    [1] Intermediate CA: CN=Let's Encrypt R3 (signed by root CA)
    [Root CA not sent — client already has it in trust store]
```

**4. ServerKeyExchange (for ECDHE)**

```
Server → Client:
  Curve: x25519
  Server's DH public key: <32 bytes>
  Signature: RSA_PSS_SHA256(client_random + server_random + DH params)
  
  The signature proves the server OWNS the private key matching the certificate.
```

**5. ClientKeyExchange**

```
Client → Server:
  Client's DH public key: <32 bytes>
  
  Now both sides can compute:
  Pre-master secret = ECDH(server_private, client_public)
                    = ECDH(client_private, server_public)  ← same value!
  
  Derive keys:
  Master secret = PRF(pre-master secret, "master secret",
                      client_random + server_random)
  
  Key material = PRF(master secret, "key expansion",
                     server_random + client_random)
  → split into: client_write_key, server_write_key,
                client_write_MAC_key, server_write_MAC_key,
                client_write_IV, server_write_IV
```

**6. Finished messages**

```
Both sides send Finished message containing:
  HMAC(master_secret, "client finished", hash(all_handshake_messages))

This verifies:
  1. Both sides computed the same master secret
  2. No one tampered with handshake messages
  3. Cipher suite negotiation wasn't downgraded
```

---

## TLS 1.3 Handshake

TLS 1.3 reduces the handshake to **1 round trip** (for new connections):

```
Client                                          Server
  │                                                │
  │──── ClientHello ──────────────────────────→     │  RTT 1
  │     (version, random, cipher suites,           │
  │      key_share, SNI, ALPN)                     │
  │                                                │
  │     ←───────────────────── ServerHello ─────    │
  │       (version, random, chosen cipher,         │
  │        key_share)                              │
  │     ←───────────────────── {EncryptedExtensions}│
  │     ←───────────────────── {Certificate} ────── │
  │     ←───────────────────── {CertificateVerify}─ │
  │     ←───────────────────── {Finished} ────────  │
  │                                                │
  │──── {Finished} ───────────────────────────→     │
  │                                                │
  │══════ Encrypted application data ═════════════ │
  
  {} = encrypted (after ServerHello, everything is encrypted!)
```

### What changed from TLS 1.2

```
1. Key share in ClientHello:
   Client GUESSES which key exchange the server will pick
   and sends DH public key upfront (no separate KeyExchange round trip)

2. Everything after ServerHello is encrypted:
   Certificate, extensions — all encrypted
   (TLS 1.2 sent certificate in plaintext!)

3. Removed:
   - RSA key exchange (no forward secrecy → removed)
   - Static DH (no forward secrecy → removed)
   - ChangeCipherSpec messages (unnecessary → removed)
   - Compression (CRIME attack → removed)
   - Renegotiation (complexity → removed, replaced by KeyUpdate)
   - RC4, DES, 3DES, CBC mode ciphers (insecure → removed)
   - SHA-1, MD5 in handshake (insecure → removed)

4. Only AEAD ciphers allowed:
   - TLS_AES_128_GCM_SHA256
   - TLS_AES_256_GCM_SHA384
   - TLS_CHACHA20_POLY1305_SHA256
   (5 cipher suites total, vs 300+ in TLS 1.2)
```

### The 1-RTT improvement explained

```
TLS 1.2 requires 2 RTTs because:
  RTT 1: Exchange hellos, server sends cert+DH params
  RTT 2: Client sends DH params, both compute keys, Finished

TLS 1.3 requires 1 RTT because:
  Client sends DH key_share in ClientHello (guesses curve)
  Server responds with its key_share in ServerHello
  Both can compute keys IMMEDIATELY after RTT 1
  Server sends encrypted cert+Finished in same flight

If client guesses wrong curve:
  Server sends HelloRetryRequest → client resends with correct curve
  → Falls back to 2 RTTs (rare in practice)
```

---

## Key Derivation

TLS 1.3 uses HKDF (HMAC-based Key Derivation Function) instead of the custom PRF:

```
                    0
                    │
                    ▼
        PSK ──→ HKDF-Extract = Early Secret
                    │
                    ▼ Derive-Secret(., "derived", "")
                    │
     ECDHE ──→ HKDF-Extract = Handshake Secret
                    │
                    ├──→ client_handshake_traffic_secret
                    │      → client_handshake_key + client_handshake_iv
                    │
                    ├──→ server_handshake_traffic_secret
                    │      → server_handshake_key + server_handshake_iv
                    │
                    ▼ Derive-Secret(., "derived", "")
                    │
     0 ──────→ HKDF-Extract = Master Secret
                    │
                    ├──→ client_application_traffic_secret
                    │      → client_app_key + client_app_iv
                    │
                    └──→ server_application_traffic_secret
                           → server_app_key + server_app_iv
```

Different keys for different purposes:
- Handshake keys: Encrypt handshake messages (Certificate, Finished)
- Application keys: Encrypt actual data (HTTP requests/responses)
- Client/server use different keys (different directions)

---

## Cipher Suites

### TLS 1.2 format

```
TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
  │      │    │       │    │       │
  │      │    │       │    │       └── Hash for PRF
  │      │    │       │    └────────── AEAD mode
  │      │    │       └─────────────── Symmetric cipher + key size
  │      │    └─────────────────────── Authentication method
  │      └──────────────────────────── Key exchange
  └─────────────────────────────────── Protocol
```

### TLS 1.3 format (simplified)

```
TLS_AES_256_GCM_SHA384
  │    │    │      │
  │    │    │      └── Hash for HKDF
  │    │    └────────── AEAD mode
  │    └─────────────── Symmetric cipher + key size
  └──────────────────── Protocol

Key exchange and authentication negotiated separately via extensions.
Only 5 cipher suites in TLS 1.3:
  TLS_AES_128_GCM_SHA256          (mandatory)
  TLS_AES_256_GCM_SHA384
  TLS_CHACHA20_POLY1305_SHA256
  TLS_AES_128_CCM_SHA256
  TLS_AES_128_CCM_8_SHA256
```

### Choosing ciphers

```bash
# See which ciphers a server supports
nmap --script ssl-enum-ciphers -p 443 example.com

# Or with openssl
openssl s_client -connect example.com:443 -cipher 'ALL' </dev/null 2>/dev/null | grep "Cipher"

# Check TLS 1.3 specifically
openssl s_client -connect example.com:443 -tls1_3 </dev/null 2>/dev/null | grep "Cipher"
```

---

## Record Protocol

After handshake, data is sent in TLS records:

```
+──────────────────────────────────+
│ Content Type      (1 byte)       │  23 = application data
│ Protocol Version  (2 bytes)      │  0x0303 = TLS 1.2 (also used by 1.3!)
│ Length            (2 bytes)       │  Size of encrypted payload
├──────────────────────────────────┤
│ Encrypted payload                │
│ (actual data + AEAD tag)         │
│ Max 2^14 + 256 bytes             │
+──────────────────────────────────+
```

### TLS 1.3 record encryption

```
Plaintext:
  [content type] [padding...0x00] [actual data]

Encrypted with AEAD:
  nonce = per-record IV XOR sequence number
  additional_data = record header (content type + version + length)
  
  AES-GCM(key, nonce, plaintext, additional_data) → ciphertext + tag
  
  The tag ensures:
    - Ciphertext wasn't modified
    - Record header wasn't modified
    - Records can't be reordered (sequence number in nonce)
    - Records can't be replayed (sequence number increments)
```

---

## Session Resumption

Establishing new TLS connections is expensive. Resumption reuses previous session parameters.

### TLS 1.2: Session IDs and Session Tickets

```
Session ID:
  1. Full handshake → server generates session ID
  2. Server stores session state (keys) indexed by ID
  3. Next connection: Client sends session ID in ClientHello
  4. Server looks up cached state → abbreviated handshake (1 RTT)
  Problem: Server must store state for every client

Session Tickets:
  1. Full handshake → server encrypts session state into "ticket"
  2. Server sends ticket to client
  3. Next connection: Client sends ticket in ClientHello
  4. Server decrypts ticket → recovers session state → abbreviated handshake
  Advantage: Server is stateless! Client stores the encrypted state.
```

### TLS 1.3: PSK (Pre-Shared Key)

```
After a TLS 1.3 handshake completes:
  Server sends NewSessionTicket message (encrypted session data)
  Client stores it

Next connection:
  Client includes PSK identity in ClientHello
  Server validates PSK → 1-RTT handshake (or 0-RTT with early data)

PSK + ECDHE: Forward secrecy maintained
  Even with PSK, TLS 1.3 does ECDHE for forward secrecy
  PSK just provides authentication (skip certificate exchange)
```

---

## 0-RTT

TLS 1.3's most aggressive optimization — send data on the FIRST packet:

```
Standard 1-RTT:
  Client: ClientHello ─────────→ Server
  Server: ServerHello + data ──→ Client
  Client: Finished + HTTP GET ─→ Server  ← data flows after 1 RTT

0-RTT:
  Client: ClientHello + HTTP GET ────→ Server  ← data flows IMMEDIATELY!
  Server: ServerHello + response ────→ Client
  
  Saved: 1 full RTT (50-200ms on transatlantic connections)
```

### 0-RTT security trade-offs

```
Problem: 0-RTT data has NO forward secrecy and is REPLAYABLE.

Replay attack:
  1. Client sends 0-RTT: "Transfer $100 to Bob"
  2. Attacker records the encrypted 0-RTT data
  3. Attacker replays it to the server
  4. Server processes it again → $200 transferred!

Mitigations:
  - Only use 0-RTT for idempotent requests (GET, not POST)
  - Server maintains replay cache (reject duplicate tickets)
  - Time-limited: 0-RTT only valid within short window
  - Application-level idempotency keys

Most implementations:
  - Allow 0-RTT for safe HTTP methods (GET, HEAD)
  - Reject 0-RTT for unsafe methods (POST, PUT, DELETE)
```

---

## TLS Termination

Where TLS is terminated affects architecture and security:

### At the origin server

```
Client ────── TLS ────── Server
End-to-end encryption.
Server handles TLS overhead (CPU for crypto).
Simplest architecture.
```

### At the load balancer

```
Client ── TLS ── Load Balancer ── HTTP ── Server
                     ↑
                TLS terminated here

Pros: Offloads CPU from servers, centralized cert management
Cons: Traffic between LB and server is unencrypted!
Fix:  Re-encrypt: LB → internal TLS → Server (but adds latency)
```

### At CDN/reverse proxy

```
Client ── TLS ── CDN edge ── TLS ── Origin Server

CDN terminates TLS at edge (close to user → lower latency)
CDN re-encrypts to origin
CDN has access to plaintext content (can cache, optimize)
```

### SNI — Server Name Indication

```
Problem: One IP hosts multiple HTTPS sites.
  TLS handshake happens BEFORE HTTP headers.
  Server doesn't know which certificate to present!

Solution: SNI extension in ClientHello:
  server_name: "example.com"
  Server reads SNI → selects correct certificate → proceeds

Problem with SNI:
  SNI is sent in PLAINTEXT (before encryption!)
  → ISP/observer KNOWS which site you're connecting to

Solution: Encrypted Client Hello (ECH, draft):
  Encrypt the entire ClientHello using a key from DNS
  → Even SNI is hidden
```

---

## Common TLS Failures

### Certificate errors

```
Error: "certificate has expired"
  → Certificate validity period passed
  → Fix: Renew certificate (Let's Encrypt auto-renews)

Error: "certificate is not yet valid"
  → Certificate's notBefore date is in the future
  → Usually: client's clock is wrong

Error: "unable to verify the first certificate"
  → Server didn't send intermediate certificate
  → Fix: Configure server to send full chain

Error: "self-signed certificate"
  → Certificate not signed by trusted CA
  → Expected in dev, flag in production

Error: "hostname mismatch"
  → Certificate CN/SAN doesn't match the domain
  → Example: cert for example.com, accessed via www.example.com
  → Fix: Include all domains in SAN (Subject Alternative Name)
```

### Protocol/cipher errors

```
Error: "no shared cipher"
  → Client and server have no common cipher suite
  → Usually: server only supports outdated ciphers

Error: "protocol version not supported"
  → Client requires TLS 1.2+, server only supports TLS 1.0
  → Or: server requires TLS 1.3, client is too old

Error: "handshake failure"
  → Generic catch-all for handshake problems
  → Check: cipher mismatch, certificate issues, SNI missing
```

### Network-level TLS failures

```
Error: "connection reset" during handshake
  → Firewall or DPI device blocking TLS
  → Middlebox doesn't understand TLS version

Error: "connection timeout" during handshake
  → Firewall silently dropping TLS packets
  → Server overloaded (TLS is CPU-intensive)

Error: "SSL_ERROR_SYSCALL" / "unexpected EOF"
  → Server closed connection abruptly
  → Often: server crashed or certificate loading error
```

---

## Linux: TLS Debugging

### openssl s_client (most important TLS debugging tool)

```bash
# Basic TLS connection test
openssl s_client -connect example.com:443 </dev/null

# Key output to look for:
#   Certificate chain      → verify chain is complete
#   Protocol               → TLSv1.3
#   Cipher                 → TLS_AES_256_GCM_SHA384
#   Verify return code: 0 (ok)  → certificate valid

# Force specific TLS version
openssl s_client -connect example.com:443 -tls1_2 </dev/null
openssl s_client -connect example.com:443 -tls1_3 </dev/null

# Specify SNI (important for virtual hosting)
openssl s_client -connect 93.184.216.34:443 -servername example.com </dev/null

# Show full certificate
openssl s_client -connect example.com:443 </dev/null 2>/dev/null | \
  openssl x509 -noout -text

# Check certificate dates
openssl s_client -connect example.com:443 </dev/null 2>/dev/null | \
  openssl x509 -noout -dates
# notBefore=Jan  1 00:00:00 2024 GMT
# notAfter=Apr  1 00:00:00 2024 GMT

# Check certificate SANs (Subject Alternative Names)
openssl s_client -connect example.com:443 </dev/null 2>/dev/null | \
  openssl x509 -noout -ext subjectAltName

# Show cipher suites offered
openssl s_client -connect example.com:443 -cipher 'ALL:COMPLEMENTOFALL' </dev/null

# Check OCSP stapling
openssl s_client -connect example.com:443 -status </dev/null 2>/dev/null | \
  grep "OCSP Response"
```

### curl TLS debugging

```bash
# Verbose TLS details
curl -v https://example.com 2>&1 | grep -E "SSL|TLS|certificate|issuer"

# Force TLS version
curl --tls-max 1.2 https://example.com
curl --tlsv1.3 https://example.com

# Show certificate info
curl -vvv https://example.com 2>&1 | grep -A5 "Server certificate"

# Ignore certificate errors (testing only!)
curl -k https://self-signed.example.com

# Use specific CA bundle
curl --cacert /path/to/ca-bundle.crt https://example.com
```

### Capture and decrypt TLS with Wireshark

```bash
# Method 1: SSLKEYLOGFILE (works with any client that supports it)
export SSLKEYLOGFILE=/tmp/tls-keys.log
curl https://example.com
# Open pcap in Wireshark, set TLS → Pre-Master-Secret log → /tmp/tls-keys.log
# Now you can see decrypted TLS application data

# Method 2: With browsers
SSLKEYLOGFILE=/tmp/tls-keys.log firefox
SSLKEYLOGFILE=/tmp/tls-keys.log google-chrome
```

### testssl.sh — comprehensive TLS testing

```bash
# Install
git clone --depth 1 https://github.com/drwetter/testssl.sh.git

# Run full test
./testssl.sh/testssl.sh https://example.com

# Output includes:
#   Protocol support (TLS 1.0-1.3)
#   Cipher suite preference
#   Certificate details
#   Known vulnerabilities (Heartbleed, POODLE, etc.)
#   HTTP security headers
```

---

## Key Takeaways

1. **TLS 1.2** = 2-RTT handshake; **TLS 1.3** = 1-RTT (and 0-RTT for resumed connections)
2. **TLS 1.3 removed** static RSA, CBC mode, SHA-1, compression, and renegotiation — only AEAD ciphers allowed
3. **Everything after ServerHello is encrypted** in TLS 1.3 (certificate too!) — TLS 1.2 sent certificates in plaintext
4. **ECDHE** (ephemeral DH) provides forward secrecy — compromised key doesn't reveal past traffic
5. **Session resumption** avoids full handshake overhead — TLS 1.3 uses PSK (Pre-Shared Keys)
6. **0-RTT is fast but dangerous** — replayable, no forward secrecy. Only use for idempotent operations (GET)
7. **SNI tells server which certificate** to use (unencrypted in TLS 1.2/1.3 — ECH will fix this)
8. **Certificate errors** are the #1 TLS failure: expired, missing intermediate, hostname mismatch
9. **`openssl s_client`** is the Swiss Army knife for TLS debugging
10. **SSLKEYLOGFILE** lets you decrypt TLS traffic in Wireshark — essential for debugging

---

## Next

→ [03-certificates-and-pki.md](03-certificates-and-pki.md) — How trust is established: certificates, CAs, and PKI
