# Why Encryption — Threat Models and Cryptographic Foundations

> Before diving into TLS, you must understand WHY encryption matters. What exactly are we protecting against? What happens without it? This module builds the threat model first, then introduces the cryptographic primitives that TLS relies on.

---

## Table of Contents

1. [The Threat Model](#threat-model)
2. [What Happens Without Encryption](#without-encryption)
3. [Security Goals — The CIA Triad](#cia-triad)
4. [Symmetric Encryption](#symmetric)
5. [Asymmetric Encryption (Public Key)](#asymmetric)
6. [Hash Functions](#hashes)
7. [Message Authentication Codes (MAC)](#mac)
8. [Digital Signatures](#signatures)
9. [Diffie-Hellman Key Exchange](#diffie-hellman)
10. [Forward Secrecy](#forward-secrecy)
11. [Putting It All Together](#together)
12. [Linux: Crypto on the Command Line](#linux-crypto)
13. [Key Takeaways](#key-takeaways)

---

## The Threat Model

### Who can see your traffic?

Every packet you send crosses multiple networks:

```
Your laptop
  → WiFi access point (coffee shop, hotel, airport)
    → ISP router
      → ISP backbone
        → Internet exchange point (IXP)
          → Destination ISP
            → Destination server

Entities that can monitor/modify traffic:
  1. Anyone on the same WiFi (eavesdropping, ARP spoofing)
  2. WiFi access point operator (sees all unencrypted traffic)
  3. Your ISP (legally required to cooperate with law enforcement)
  4. Government agencies (mass surveillance programs)
  5. IXPs and backbone operators
  6. CDNs and reverse proxies (if they terminate TLS)
  7. Destination server operator
```

### The attack types

| Attack | What it means | Example |
|--------|--------------|---------|
| **Eavesdropping** | Reading data in transit | Sniffing WiFi for passwords |
| **Tampering** | Modifying data in transit | Injecting ads into HTTP pages |
| **Spoofing** | Pretending to be someone else | Fake DNS response, fake server |
| **Replay** | Resending captured valid data | Replaying a payment request |
| **Man-in-the-Middle** | Intercept + modify + forward | MITM proxy between you and bank |

---

## Without Encryption

### HTTP (no encryption)

```bash
# On public WiFi, anyone can capture your traffic:
sudo tcpdump -i wlan0 -A 'port 80' | grep -i "password\|cookie\|session"

# What they see:
POST /login HTTP/1.1
Host: example.com
Content-Type: application/x-www-form-urlencoded

username=alice&password=MyS3cretP@ss!    ← plaintext password!

# Also visible:
Cookie: session=abc123def456             ← steal session
Authorization: Basic YWxpY2U6c2VjcmV0   ← base64 = not encryption!
```

### ISP injection

```
Real-world: ISPs have been caught injecting:
  - Ads into HTTP pages
  - Tracking headers (Verizon's X-UIDH "supercookie")
  - JavaScript for analytics

Without encryption, anyone in the path can:
  → Read everything you send/receive
  → Modify web pages in transit
  → Inject malicious content
  → Track every site you visit
```

---

## The CIA Triad

Encryption provides three fundamental security properties:

### Confidentiality

```
Only intended parties can read the data.

Without: Anyone on the network reads your traffic
With:    Traffic looks like random bytes to observers

Tool: Encryption (symmetric or asymmetric)
```

### Integrity

```
Data hasn't been modified in transit.

Without: ISP inserts ad into webpage, you don't know
With:    Any modification detected → connection rejected

Tool: Hash functions, MACs, digital signatures
```

### Authentication

```
You're talking to who you think you're talking to.

Without: DNS poisoning sends you to fake bank site
With:    Server proves identity with certificate

Tool: Digital signatures, certificates, PKI
```

Plus two more properties TLS provides:

### Non-repudiation

```
Sender can't deny sending a message.
Tool: Digital signatures
```

### Forward secrecy

```
Compromising today's keys doesn't reveal yesterday's traffic.
Tool: Ephemeral Diffie-Hellman
```

---

## Symmetric Encryption

Same key encrypts and decrypts:

```
Alice                                     Bob
  │                                        │
  │  Key: "s3cr3tK3y"                      │  Key: "s3cr3tK3y"
  │                                        │
  │  Encrypt("Hello", key)                 │
  │  → 0x7a3f2b1c...                       │
  │  ─────────────────────────────────→     │
  │                                        │  Decrypt(0x7a3f2b1c..., key)
  │                                        │  → "Hello"
```

### Algorithms

| Algorithm | Key Size | Block Size | Status |
|-----------|----------|------------|--------|
| AES-128 | 128 bit | 128 bit | ✅ Secure |
| AES-256 | 256 bit | 128 bit | ✅ Secure (used for classified data) |
| ChaCha20 | 256 bit | N/A (stream) | ✅ Secure (modern alternative to AES) |
| 3DES | 168 bit | 64 bit | ⚠️ Deprecated |
| DES | 56 bit | 64 bit | ❌ Broken (brute-forceable) |
| RC4 | 40-2048 bit | N/A (stream) | ❌ Broken |

### Block cipher modes

AES is a block cipher (encrypts 128-bit blocks). Modes determine how to encrypt data longer than one block:

```
ECB (Electronic Codebook) — NEVER USE:
  Same plaintext block → same ciphertext block
  Patterns visible! (famous "ECB penguin" image)

CBC (Cipher Block Chaining):
  Each block XORed with previous ciphertext block
  Requires IV (initialization vector)
  Vulnerable to padding oracle attacks (POODLE)
  
GCM (Galois/Counter Mode) — RECOMMENDED:
  Counter mode encryption + authentication tag (AEAD)
  Parallelizable, hardware-accelerated
  TLS 1.3 uses AES-128-GCM and AES-256-GCM

CCM (Counter with CBC-MAC):
  Counter mode + CBC-MAC for authentication
  Used in WiFi (WPA2/WPA3)
```

### The key distribution problem

```
Symmetric encryption is fast, but has a fatal problem:

How do Alice and Bob share the secret key?
  → Can't send it over the network (anyone can intercept)
  → Meet in person? Doesn't scale to millions of websites
  → Ship USB drives? Impractical

Solution: Use asymmetric encryption to exchange symmetric keys!
(This is exactly what TLS does)
```

---

## Asymmetric Encryption

Two different keys: public key (share with everyone) and private key (keep secret):

```
Alice                                     Bob
  │                                        │
  │  Bob's public key (known to everyone)  │  Bob's private key (secret)
  │                                        │  Bob's public key (published)
  │  Encrypt("Hello", Bob's public key)    │
  │  → 0x8f2e1a...                         │
  │  ───────────────────────────────────→   │
  │                                        │  Decrypt(0x8f2e1a..., Bob's private key)
  │                                        │  → "Hello"
  │                                        │
  │  ONLY Bob can decrypt (only he has     │
  │  the private key)                      │
```

### Algorithms

| Algorithm | Key Size | Speed | Status |
|-----------|----------|-------|--------|
| RSA-2048 | 2048 bit | Slow | ✅ Secure (minimum recommended) |
| RSA-4096 | 4096 bit | Very slow | ✅ Secure (long-term) |
| ECDSA (P-256) | 256 bit | Fast | ✅ Secure |
| EdDSA (Ed25519) | 256 bit | Fast | ✅ Secure (modern preference) |

### Why both symmetric and asymmetric?

```
Asymmetric encryption is 100-1000× slower than symmetric.

TLS solution: Use both!
  1. Asymmetric: Exchange a shared secret (key exchange) — slow but secure
  2. Symmetric: Encrypt actual data with shared secret — fast

This is called "hybrid encryption" and is how TLS works.
```

---

## Hash Functions

A hash function converts any input to a fixed-size output (digest):

```
Input: "Hello" → SHA-256 → 185f8db32271fe25f561a6fc938b2e264306ec304eda518007d1764826381969
Input: "Hello!" → SHA-256 → 334d016f755cd6dc58c53a86e183882f8ec14f52fb05345887c8a5edd42c87b7
                              ↑ completely different output for tiny change!
```

### Properties

```
1. Deterministic: Same input → same hash (always)
2. Fixed output: Any input size → same hash size (SHA-256 = 256 bits)
3. One-way: Can't reverse hash → original data (no decryption)
4. Collision resistant: Can't find two inputs with same hash
5. Avalanche: 1-bit input change → ~50% of output bits change
```

### Algorithms

| Algorithm | Output Size | Status |
|-----------|-------------|--------|
| MD5 | 128 bit | ❌ Broken (collisions found) |
| SHA-1 | 160 bit | ❌ Broken (collision demonstrated 2017) |
| SHA-256 | 256 bit | ✅ Secure |
| SHA-384 | 384 bit | ✅ Secure |
| SHA-512 | 512 bit | ✅ Secure |
| SHA-3 | Variable | ✅ Secure (different design from SHA-2) |
| BLAKE2/3 | Variable | ✅ Secure (faster than SHA-2) |

### Uses in networking

```
1. Data integrity: Hash of file verifies it wasn't corrupted
2. TLS: Verify handshake messages weren't tampered with
3. HMAC: Create authentication codes (next section)
4. Certificate fingerprints: Identify certificates
5. Password storage: Store hash, not password (actually use bcrypt/argon2)
```

---

## MAC

A MAC (Message Authentication Code) combines a hash with a secret key:

```
HMAC = Hash(key || message)  (simplified)

Alice → Bob:
  Message: "Transfer $100 to Bob"
  HMAC: HMAC-SHA256("Transfer $100 to Bob", shared_key)
  Send: message + HMAC

Bob verifies:
  Compute: HMAC-SHA256("Transfer $100 to Bob", shared_key)
  Compare: computed HMAC == received HMAC?
  If match → message is authentic and unmodified
```

### Why MAC instead of just hash?

```
Hash alone (no key):
  Attacker intercepts: "Transfer $100" + SHA256("Transfer $100")
  Attacker modifies:   "Transfer $10000" + SHA256("Transfer $10000")
  → Attacker can recompute hash (no secret needed!)

HMAC (with shared key):
  Attacker intercepts: "Transfer $100" + HMAC(key, "Transfer $100")
  Attacker modifies:   "Transfer $10000" + ???
  → Can't compute valid HMAC without the key!
```

### AEAD — Authenticated Encryption with Associated Data

Modern ciphers combine encryption + MAC in one operation:

```
AES-256-GCM = AES encryption + GHASH-based authentication tag

Input:  plaintext + key + nonce + associated data (AAD)
Output: ciphertext + authentication tag

Tag verifies:
  1. Ciphertext wasn't modified (integrity)
  2. Associated data wasn't modified (e.g., packet headers)
  3. Sender knows the key (authentication)

TLS 1.3 exclusively uses AEAD ciphers (AES-GCM, ChaCha20-Poly1305)
```

---

## Digital Signatures

Asymmetric version of a MAC — prove who created a message:

```
Signing (private key):
  Alice signs: Sign("document", Alice's_private_key) → signature

Verifying (public key):
  Anyone verifies: Verify("document", signature, Alice's_public_key) → true/false
  
  Only Alice could have created this signature (only she has private key)
  Anyone can verify (public key is public)
```

### How digital signatures work in practice

```
1. Hash the message: digest = SHA-256("document")
2. Sign the hash: signature = RSA_Sign(digest, private_key)
3. Send: message + signature

Verification:
1. Hash received message: digest' = SHA-256("document")
2. Verify: RSA_Verify(digest', signature, public_key) → true
```

### Use in TLS

```
1. Certificate signing: CA signs server's certificate with CA's private key
2. Handshake authentication: Server signs handshake parameters with server's private key
3. Client can verify: Using CA's and server's public keys
```

---

## Diffie-Hellman Key Exchange

DH solves the key distribution problem — two parties agree on a shared secret over an insecure channel:

```
Public knowledge: p (large prime), g (generator)

Alice:                                Bob:
  Pick secret a                         Pick secret b
  Compute A = g^a mod p                 Compute B = g^b mod p
  
  ──── Send A ────────────────→
  ←──── Send B ────────────────
  
  Compute: B^a mod p                    Compute: A^b mod p
  = (g^b)^a mod p                       = (g^a)^b mod p
  = g^(ab) mod p                        = g^(ab) mod p
  
  SAME VALUE! This is the shared secret.
```

### Why eavesdroppers can't derive the secret

```
Eavesdropper sees: p, g, A = g^a mod p, B = g^b mod p

To find shared secret g^(ab) mod p, they need a or b.
Computing a from A = g^a mod p is the "discrete logarithm problem"
→ computationally infeasible for large p (2048+ bits)
```

### Elliptic Curve Diffie-Hellman (ECDHE)

```
Same concept, but using elliptic curve math instead of modular exponentiation.

Advantages:
  - Smaller keys (256-bit ECDH ≈ 3072-bit DH security)
  - Faster computation
  - Less bandwidth

TLS 1.3 uses: X25519 (Curve25519) — most common
               P-256 (NIST curve) — also supported
```

---

## Forward Secrecy

### The problem without forward secrecy

```
Static RSA key exchange (old TLS):
  1. Server has long-lived RSA key pair
  2. Client generates random "pre-master secret"
  3. Client encrypts with server's public key
  4. Server decrypts with private key
  5. Both derive session keys from pre-master secret

If server's private key is EVER compromised:
  Attacker recorded ALL past encrypted traffic
  Attacker decrypts ALL past traffic with stolen key
  → All historical sessions compromised!
```

### Forward secrecy with ephemeral DH

```
Ephemeral Diffie-Hellman (DHE/ECDHE):
  1. Each session: generate NEW DH key pair
  2. Exchange DH public values
  3. Derive shared secret
  4. After session: DELETE DH private key

If server's long-term key is compromised:
  Attacker can impersonate server (future sessions)
  BUT past sessions used DIFFERENT DH keys (deleted!)
  → Past traffic remains encrypted!

TLS 1.3 REQUIRES forward secrecy (only ECDHE/DHE allowed)
TLS 1.2: Optional (depends on cipher suite selection)
```

---

## Putting It All Together

How TLS combines all these primitives:

```
TLS needs:
  Confidentiality  → Symmetric encryption (AES-GCM/ChaCha20)
  Key exchange     → ECDHE (forward-secret shared secret)
  Authentication   → Digital signatures (server proves identity)
  Integrity        → AEAD tag (built into AES-GCM/ChaCha20-Poly1305)
  Trust            → Certificates + PKI (trust chain)

TLS 1.3 cipher suite example:
  TLS_AES_256_GCM_SHA384
  
  AES_256_GCM = symmetric encryption (AEAD — encryption + integrity)
  SHA384      = hash for key derivation (HKDF)
  
  Key exchange (ECDHE) and authentication (ECDSA/RSA) are 
  negotiated separately in TLS 1.3.
```

---

## Linux: Crypto on the Command Line

### Hashing

```bash
# SHA-256 hash of a file
sha256sum file.txt
# a1b2c3d4... file.txt

# SHA-256 hash of a string
echo -n "Hello World" | sha256sum
# a591a6d40... -

# Compare: MD5 (don't use for security)
echo -n "Hello World" | md5sum
```

### Symmetric encryption with OpenSSL

```bash
# Encrypt a file with AES-256-CBC
openssl enc -aes-256-cbc -salt -in plaintext.txt -out encrypted.bin -pass pass:mypassword

# Decrypt
openssl enc -aes-256-cbc -d -in encrypted.bin -out decrypted.txt -pass pass:mypassword

# AES-256-GCM (AEAD — preferred)
openssl enc -aes-256-gcm -in plaintext.txt -out encrypted.bin -pass pass:mypassword
```

### Asymmetric keys

```bash
# Generate RSA key pair
openssl genrsa -out private.pem 4096
openssl rsa -in private.pem -pubout -out public.pem

# Generate Ed25519 key pair (modern)
openssl genpkey -algorithm Ed25519 -out private.pem
openssl pkey -in private.pem -pubout -out public.pem

# Encrypt with public key
openssl pkeyutl -encrypt -pubin -inkey public.pem -in secret.txt -out encrypted.bin

# Decrypt with private key
openssl pkeyutl -decrypt -inkey private.pem -in encrypted.bin -out decrypted.txt
```

### Digital signatures

```bash
# Sign a file
openssl dgst -sha256 -sign private.pem -out signature.bin document.txt

# Verify signature
openssl dgst -sha256 -verify public.pem -signature signature.bin document.txt
# Verified OK
```

### HMAC

```bash
# Compute HMAC
echo -n "message" | openssl dgst -sha256 -hmac "secret-key"
# (stdin)= a4331...(hex digest)
```

### Check TLS connection

```bash
# See what cipher suite a server uses
openssl s_client -connect example.com:443 </dev/null 2>/dev/null | grep "Cipher"
# New, TLSv1.3, Cipher is TLS_AES_256_GCM_SHA384

# Full TLS information
openssl s_client -connect example.com:443 </dev/null 2>/dev/null | \
  grep -E "Protocol|Cipher|Server public key"
```

---

## Key Takeaways

1. **Without encryption**: Anyone on the path (WiFi, ISP, backbone) can read and modify your traffic
2. **CIA triad**: Confidentiality (encryption), Integrity (hashes/MACs), Authentication (signatures/certs)
3. **Symmetric encryption** (AES) is fast but requires shared key — the key distribution problem
4. **Asymmetric encryption** (RSA/ECDSA) solves key distribution but is 100-1000× slower
5. **TLS uses both**: Asymmetric for key exchange, symmetric for data — hybrid encryption
6. **Hash functions** (SHA-256) provide integrity — any change produces completely different output
7. **HMAC** = hash + key — proves both integrity AND authenticity (need key to generate)
8. **AEAD** (AES-GCM) combines encryption + authentication in one operation — what TLS 1.3 uses
9. **Diffie-Hellman** allows agreeing on shared secret over insecure channel (eavesdropper can't derive it)
10. **Forward secrecy** (ephemeral DH): Compromising today's key doesn't reveal yesterday's traffic — TLS 1.3 requires it

---

## Next

→ [02-tls-handshake.md](02-tls-handshake.md) — The TLS handshake step by step
