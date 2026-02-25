# Certificates and PKI — The Trust Infrastructure

> TLS encryption means nothing if you can't verify WHO you're talking to. Certificates and Public Key Infrastructure (PKI) are the trust system that makes HTTPS work. Understanding PKI is critical — certificate failures are the most common TLS issue you'll encounter.

---

## Table of Contents

1. [The Trust Problem](#trust-problem)
2. [X.509 Certificates](#x509)
3. [Certificate Chains](#chains)
4. [Certificate Authorities (CAs)](#cas)
5. [Certificate Issuance — How CAs Verify You](#issuance)
6. [Let's Encrypt and ACME](#lets-encrypt)
7. [Certificate Revocation](#revocation)
8. [Certificate Transparency](#transparency)
9. [Certificate Pinning](#pinning)
10. [Mutual TLS (mTLS)](#mtls)
11. [Private PKI](#private-pki)
12. [Linux: Certificate Operations](#linux-certs)
13. [Key Takeaways](#key-takeaways)

---

## The Trust Problem

Encryption alone doesn't provide security:

```
Without authentication:
  You: "I want to connect securely to bank.com"
  Attacker (MITM): "Sure! Here's my public key. Let's encrypt."
  You: (encrypts all banking data with attacker's key)
  Attacker: (decrypts, reads, re-encrypts to real bank)
  
  You have encryption... with the wrong party!
```

The fundamental question: **How do you know the public key belongs to bank.com and not an attacker?**

Answer: A trusted third party (Certificate Authority) vouches for it. "I, DigiCert, verify that this public key belongs to bank.com."

---

## X.509 Certificates

A certificate binds a public key to an identity. The X.509 format is the standard:

```
Certificate:
    Data:
        Version: 3
        Serial Number: 04:56:ab:cd:...
        
        Signature Algorithm: sha256WithRSAEncryption
        
        Issuer: C=US, O=Let's Encrypt, CN=R3
        
        Validity:
            Not Before: Jan  1 00:00:00 2024 GMT
            Not After : Apr  1 00:00:00 2024 GMT
        
        Subject: CN=example.com
        
        Subject Public Key Info:
            Public Key Algorithm: id-ecPublicKey
                Public-Key: (256 bit)
                ASN1 OID: prime256v1
                pub: 04:3b:a5:...
        
        X509v3 Extensions:
            Subject Alternative Name (SAN):
                DNS:example.com
                DNS:www.example.com
                DNS:*.example.com
            
            Basic Constraints: critical
                CA:FALSE
            
            Key Usage: critical
                Digital Signature
            
            Extended Key Usage:
                TLS Web Server Authentication
            
            Authority Information Access:
                OCSP - URI:http://r3.o.lencr.org
                CA Issuers - URI:http://r3.i.lencr.org/
            
            Certificate Policies:
                2.23.140.1.2.1 (DV)
    
    Signature Algorithm: sha256WithRSAEncryption
    Signature Value: 8f:3a:2b:...
```

### Key fields explained

| Field | Purpose |
|-------|---------|
| **Serial Number** | Unique ID for this certificate (per CA) |
| **Issuer** | Who signed this certificate (the CA) |
| **Subject** | Who this certificate identifies |
| **Validity** | When the certificate is valid (notBefore → notAfter) |
| **Public Key** | The server's public key |
| **SAN** | Subject Alternative Name — all valid domain names |
| **Basic Constraints** | CA:TRUE = can sign other certs; CA:FALSE = leaf cert |
| **Key Usage** | What the key can be used for |
| **Signature** | CA's digital signature over all the above |

### SAN vs CN (Common Name)

```
Old way (deprecated): Subject CN = example.com
New way (required):   SAN = DNS:example.com, DNS:www.example.com

Browsers check SAN first. Chrome ignores CN entirely since 2017.
Always put all domains in SAN.

Wildcard: *.example.com matches www.example.com, api.example.com
          Does NOT match: example.com (bare domain)
          Does NOT match: sub.www.example.com (multiple levels)
```

---

## Certificate Chains

Certificates form a chain of trust:

```
Root CA (self-signed, in browser/OS trust store)
  │
  │ Signs ↓
  │
Intermediate CA (signed by root)
  │
  │ Signs ↓
  │
Leaf Certificate (your server's cert, signed by intermediate)
```

### Why intermediates?

```
Root CA private keys are EXTREMELY valuable.
  - Stored in Hardware Security Modules (HSMs) in secure vaults
  - Air-gapped (not connected to any network)
  - Used rarely (only to sign intermediates)

If a root key is compromised → every certificate under it is compromised
  → Millions of websites affected
  → Root must be rotated across all browsers/OS (takes years)

Intermediate CAs:
  - Private key in online HSMs (accessible for signing)
  - If compromised → revoke only this intermediate
  - Root remains safe → create new intermediate
  - Limits blast radius
```

### Chain verification

```
Client receives: [leaf cert] [intermediate cert]

Verification steps:
1. Is leaf cert's signature valid? 
   → Verify with intermediate's public key ✓
2. Is intermediate cert's signature valid?
   → Verify with root's public key ✓
3. Is root in my trust store?
   → Yes ✓ (root CA is pre-installed in OS/browser)
4. Are all certs within validity period?
   → Check notBefore/notAfter ✓
5. Is the leaf cert's SAN matching the hostname?
   → example.com matches DNS:example.com ✓
6. Are any certs revoked?
   → Check CRL/OCSP ✓

Result: Certificate chain is valid → trust this server's public key
```

### Common chain errors

```
Error: "unable to get local issuer certificate"
  → Missing intermediate certificate!
  → Server MUST send: leaf + all intermediates
  → Server should NOT send: root (client already has it)

Test:
  curl https://example.com           → may work (curl has intermediates cached)
  openssl s_client -connect example.com:443  → shows if chain is incomplete

Fix: Configure server to send full chain
  Nginx: ssl_certificate includes leaf + intermediates
  Apache: SSLCertificateChainFile for intermediates
```

---

## Certificate Authorities

### The trust store

Your OS/browser comes with ~100-150 trusted root CAs pre-installed:

```bash
# Linux: System trust store
ls /etc/ssl/certs/
# Or: /usr/share/ca-certificates/

# Count trusted CAs
ls /etc/ssl/certs/*.pem | wc -l

# Firefox: Has its own trust store (separate from OS)
# Chrome: Uses OS trust store
# Java: Uses its own cacerts keystore
```

### Major CAs

```
Let's Encrypt:  Free, automated, most popular (~300M certs)
DigiCert:       Enterprise, EV certs
Sectigo:        High volume, affordable
Google Trust Services: Google's own CA
Amazon Trust Services: For AWS services

Historically problematic:
  Symantec/VeriSign: Distrusted by Chrome (2018) — mismanaged certs
  WoSign/StartCom:   Distrusted — backdated SHA-1 certs
  DigiNotar:         Catastrophic breach (2011) — removed from all trust stores
```

### CA compromise

```
DigiNotar (2011):
  Hackers compromised DigiNotar CA
  Issued fake certificate for *.google.com
  Iranian government used it for MITM on Gmail users
  
  Impact: DigiNotar removed from ALL trust stores
  Company bankrupted within months
  
  Lesson: A single compromised CA breaks trust for everything
  Solution: Certificate Transparency (covered later)
```

---

## Certificate Issuance

### Validation levels

| Level | Verifies | Time | Cost | Visual |
|-------|----------|------|------|--------|
| **DV** (Domain Validation) | Domain ownership only | Minutes | Free-$10 | Padlock |
| **OV** (Organization Validation) | Domain + organization identity | Days | $50-200 | Padlock |
| **EV** (Extended Validation) | Domain + org + legal existence | Weeks | $200-1000 | Padlock (green bar removed) |

### DV — How domain ownership is verified

```
Method 1: HTTP challenge
  CA: "Put this token at http://example.com/.well-known/acme-challenge/TOKEN"
  You: (create the file)
  CA: (fetches the URL, sees the token) → "You control this domain" → issue cert

Method 2: DNS challenge
  CA: "Create TXT record: _acme-challenge.example.com = TOKEN"
  You: (create DNS record)
  CA: (queries DNS, sees the token) → "You control this domain" → issue cert

Method 3: Email challenge
  CA: "We sent email to admin@example.com with a code"
  You: (click verification link)
  CA: → "You control this domain's email" → issue cert
```

---

## Let's Encrypt and ACME

Let's Encrypt revolutionized TLS:
- Free certificates
- Automated issuance and renewal
- 90-day validity (forces automation)
- ~300 million active certificates

### ACME protocol (Automated Certificate Management Environment)

```
1. Create account (one-time):
   Client → ACME server: Register with public key
   
2. Order certificate:
   Client → ACME server: "I want cert for example.com"
   Server → Client: "Prove you control it — complete these challenges"

3. Complete challenge:
   HTTP-01: Place file at /.well-known/acme-challenge/TOKEN
   DNS-01:  Create _acme-challenge TXT record
   TLS-ALPN-01: Respond on port 443 with special certificate

4. Finalize:
   Client → Server: "Challenge complete"
   Server validates → issues certificate
   Client downloads certificate + chain

5. Renew (every 60 days):
   Same process, automated via cron/systemd timer
```

### Certbot (ACME client)

```bash
# Install
sudo apt install certbot

# Get certificate (standalone — certbot runs temporary web server)
sudo certbot certonly --standalone -d example.com -d www.example.com

# Get certificate (webroot — use existing web server)
sudo certbot certonly --webroot -w /var/www/html -d example.com

# Get certificate (DNS challenge — for wildcards)
sudo certbot certonly --manual --preferred-challenges dns -d "*.example.com"

# Auto-renewal (certbot installs systemd timer)
sudo certbot renew --dry-run

# Check certificate
sudo certbot certificates

# Files created:
# /etc/letsencrypt/live/example.com/
#   fullchain.pem  → certificate + intermediates (server config uses this)
#   privkey.pem    → private key (never share!)
#   cert.pem       → certificate only
#   chain.pem      → intermediate certificates only
```

---

## Certificate Revocation

What happens when a private key is compromised? The certificate must be revoked.

### CRL (Certificate Revocation List)

```
CA publishes a list of revoked certificate serial numbers:

CRL:
  Revoked Certificates:
    Serial Number: 04:56:ab:cd
      Revocation Date: Jan 15 00:00:00 2024 GMT
      Reason: Key Compromise
    Serial Number: 07:89:ef:01
      Revocation Date: Jan 20 00:00:00 2024 GMT
      Reason: Superseded

Problems:
  - CRL files grow large (millions of entries)
  - Client must download ENTIRE CRL to check one cert
  - If CRL download fails → soft-fail (accept certificate) or hard-fail (reject)?
  - Most browsers soft-fail → revocation check is useless!
```

### OCSP (Online Certificate Status Protocol)

```
Client → OCSP responder: "Is serial 04:56:ab:cd revoked?"
OCSP responder → Client: "Good" / "Revoked" / "Unknown"

Better than CRL (single query), but:
  - Privacy: OCSP responder knows which sites you visit
  - Availability: If OCSP responder is down → soft-fail (accept?)
  - Latency: Extra HTTP request during TLS handshake
```

### OCSP Stapling (the solution)

```
Server periodically queries OCSP responder for its OWN certificate status.
Server includes timestamped, signed OCSP response in TLS handshake.

Advantages:
  - No client → OCSP responder connection needed
  - No privacy leak (client doesn't contact OCSP)
  - Server caches OCSP response (no per-client latency)
  - Signed by CA → can't be forged

OCSP Must-Staple (extension in certificate):
  Certificate says: "Server MUST staple OCSP response"
  If staple missing → client rejects certificate
  Solves the soft-fail problem!
```

```bash
# Check OCSP stapling
openssl s_client -connect example.com:443 -status </dev/null 2>/dev/null | \
  grep -A 15 "OCSP response"

# Manual OCSP query
openssl s_client -connect example.com:443 </dev/null 2>/dev/null | \
  openssl x509 -noout -ocsp_uri
# http://r3.o.lencr.org

openssl ocsp -issuer chain.pem -cert cert.pem \
  -url http://r3.o.lencr.org -resp_text
```

---

## Certificate Transparency

CT solves the "rogue CA" problem — if a CA issues a certificate for your domain without your knowledge, how would you know?

### How CT works

```
1. CA issues certificate for example.com
2. CA submits certificate to public CT logs (append-only, auditable)
3. CT log returns SCT (Signed Certificate Timestamp) — proof of inclusion
4. Server includes SCT in TLS handshake
5. Anyone can search CT logs for certificates issued for their domain

CT logs are append-only:
  - Can't remove or modify entries
  - Publicly searchable
  - Multiple independent logs (Google, Cloudflare, DigiCert, etc.)
```

### Monitoring your domains

```
Search CT logs for certificates issued for your domain:
  https://crt.sh/?q=example.com
  https://transparencyreport.google.com/https/certificates

Set up alerts:
  - Monitor for unexpected certificates
  - Detect: CA compromise, certificate mis-issuance
  - Tools: certspotter, Facebook CT monitor
```

```bash
# Search CT logs from command line
curl -s "https://crt.sh/?q=%25.example.com&output=json" | jq '.[].common_name' | sort -u
```

---

## Certificate Pinning

### What is pinning?

Pinning = hardcoding expected certificate or public key in the client:

```
Normal TLS: Accept any certificate signed by any trusted CA
Pinned TLS: Accept ONLY certificates with THIS specific public key hash

If CA is compromised and issues fake cert → pinning rejects it
  (because fake cert has different public key hash)
```

### Types of pinning

```
Certificate pinning: Pin the exact certificate
  Problem: Must update when certificate is renewed (every 90 days!)

Public key pinning: Pin the public key (survives cert renewal)
  Better: Same key can be in multiple certificates

Backup pins: Always pin at least 2 keys
  (primary + backup, in case primary key is compromised)
```

### HPKP (HTTP Public Key Pinning) — DEPRECATED

```
HTTP header:
  Public-Key-Pins: pin-sha256="base64hash1"; pin-sha256="base64hash2"; 
                   max-age=2592000; includeSubDomains

DEPRECATED because:
  - Easy to brick your site (wrong pin = site unreachable for max-age)
  - Can be used for extortion ("ransom pin" attack)
  - Certificate Transparency is a better solution
  
Replaced by: Expect-CT header (also deprecated → CT now mandatory)
```

### When pinning is still used

```
Mobile apps: Pin server's public key in app code
  - App only connects to YOUR servers
  - You control both client and server
  - Certificate transparency doesn't help (no browser involved)
  
  Example: Banking app pins bank's server public key
  Even if attacker has valid CA cert → app rejects it
```

---

## Mutual TLS (mTLS)

Standard TLS: Only server presents certificate (server authentication).
Mutual TLS: BOTH sides present certificates (mutual authentication).

```
Standard TLS:
  Client ─── "Who are you?" ──→ Server presents cert
  Client verifies server cert ✓
  Server has NO IDEA who client is (uses username/password later)

Mutual TLS:
  Client ─── "Who are you?" ──→ Server presents cert
  Server ─── "Who are you?" ──→ Client presents cert
  Both sides verified ✓ ✓
```

### Where mTLS is used

```
1. Microservice-to-microservice communication:
   Service A ──── mTLS ──── Service B
   Both services prove identity with certificates
   No passwords, tokens, or API keys needed
   
2. Kubernetes pod-to-pod (via service mesh):
   Istio/Linkerd inject sidecar proxies
   Sidecars handle mTLS automatically
   Every pod has a certificate (auto-rotated)
   
3. VPN and zero-trust networks:
   Client certificate proves device identity
   Not just user authentication — device authentication

4. IoT device authentication:
   Each device has unique certificate
   Server knows which specific device is connecting
```

### mTLS in the handshake

```
Standard TLS handshake + these additions:

Server → Client: CertificateRequest
  "I need to see YOUR certificate"
  Accepted CAs: [list of trusted CAs for client certs]

Client → Server: Certificate
  Client's X.509 certificate

Client → Server: CertificateVerify
  Signature proving client has the private key
```

---

## Private PKI

For internal infrastructure, organizations run their own CA:

```
Public PKI:                              Private PKI:
  Let's Encrypt / DigiCert signs certs     Your internal CA signs certs
  World trusts these CAs                   Only YOUR systems trust this CA
  For public-facing websites               For internal services
  Domain validation                        Any identity validation you want
  90-day / 1-year validity                 Any validity period

Use cases:
  - mTLS between microservices
  - Internal dashboards/tools
  - Database connections (PostgreSQL, MySQL)
  - Development/staging environments
  - VPN/device certificates
```

### Tools for private PKI

```
step-ca (Smallstep):
  Modern CA, ACME support, short-lived certificates
  Best for Kubernetes/microservice environments

HashiCorp Vault PKI:
  Secret management + certificate issuance
  Tight integration with infrastructure

cfssl (Cloudflare):
  Simple CA toolkit, good for scripting

OpenSSL:
  Manual but full control
  Good for learning, not for production rotation
```

---

## Linux: Certificate Operations

### View certificate details

```bash
# View a PEM certificate file
openssl x509 -in cert.pem -noout -text

# View specific fields
openssl x509 -in cert.pem -noout -subject -issuer -dates -serial

# View SANs
openssl x509 -in cert.pem -noout -ext subjectAltName

# View from a running server
echo | openssl s_client -connect example.com:443 2>/dev/null | \
  openssl x509 -noout -text

# Check expiration date
echo | openssl s_client -connect example.com:443 2>/dev/null | \
  openssl x509 -noout -enddate

# Get certificate fingerprint
openssl x509 -in cert.pem -noout -fingerprint -sha256
```

### Generate self-signed certificate

```bash
# One-liner: key + cert
openssl req -x509 -newkey rsa:4096 -sha256 -days 365 \
  -nodes -keyout key.pem -out cert.pem \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

# Or with ECDSA (modern, smaller, faster)
openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 \
  -sha256 -days 365 -nodes -keyout key.pem -out cert.pem \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"
```

### Create a simple private CA

```bash
# 1. Generate CA key and certificate
openssl req -x509 -newkey rsa:4096 -sha256 -days 3650 \
  -nodes -keyout ca-key.pem -out ca-cert.pem \
  -subj "/CN=My Internal CA" \
  -addext "basicConstraints=critical,CA:TRUE" \
  -addext "keyUsage=critical,keyCertSign,cRLSign"

# 2. Generate server key and CSR
openssl req -newkey rsa:2048 -nodes -keyout server-key.pem \
  -out server.csr -subj "/CN=myservice.internal"

# 3. Sign server certificate with CA
openssl x509 -req -in server.csr -CA ca-cert.pem -CAkey ca-key.pem \
  -CAcreateserial -out server-cert.pem -days 365 -sha256 \
  -extfile <(echo "subjectAltName=DNS:myservice.internal,DNS:*.myservice.internal")

# 4. Trust the CA on this system
sudo cp ca-cert.pem /usr/local/share/ca-certificates/my-internal-ca.crt
sudo update-ca-certificates

# 5. Verify chain
openssl verify -CAfile ca-cert.pem server-cert.pem
# server-cert.pem: OK
```

### Convert certificate formats

```bash
# PEM to DER
openssl x509 -in cert.pem -outform DER -out cert.der

# DER to PEM
openssl x509 -in cert.der -inform DER -outform PEM -out cert.pem

# PEM to PKCS#12 (for importing to Windows/Java)
openssl pkcs12 -export -in cert.pem -inkey key.pem -out cert.p12

# PKCS#12 to PEM
openssl pkcs12 -in cert.p12 -out cert.pem -nodes

# PEM to Java Keystore
keytool -import -file cert.pem -alias myserver -keystore keystore.jks
```

### Verify and debug

```bash
# Verify certificate chain
openssl verify -CAfile ca-chain.pem server-cert.pem

# Check if key matches certificate
openssl x509 -noout -modulus -in cert.pem | openssl md5
openssl rsa -noout -modulus -in key.pem | openssl md5
# If MD5 hashes match → key and cert are a pair

# Check certificate chain completeness of a server
openssl s_client -connect example.com:443 -showcerts </dev/null 2>/dev/null | \
  grep -E "s:|i:" 
# s: = subject, i: = issuer
# Each cert's issuer should match next cert's subject
```

---

## Key Takeaways

1. **Certificates bind public keys to identities** — a trusted CA vouches "this key belongs to this domain"
2. **Chain of trust**: Root CA → Intermediate CA → Leaf certificate. Server MUST send leaf + intermediates
3. **Root CAs** are pre-installed in OS/browser (~150 roots). Compromise of one root = catastrophic
4. **DV certificates** only verify domain ownership (sufficient for most use cases). Free via Let's Encrypt
5. **Let's Encrypt + ACME**: Automated, free TLS certificates. Use 90-day certs + auto-renewal
6. **Revocation is broken** in practice — CRL too large, OCSP has soft-fail. OCSP Stapling is the best option
7. **Certificate Transparency** logs all issued certificates publicly — monitor for unauthorized issuance
8. **mTLS** = both sides authenticate with certificates — used for microservice-to-microservice communication
9. **Private PKI** for internal infrastructure — run your own CA for internal services, mTLS, dev environments
10. **`openssl x509`** and **`openssl s_client`** are essential tools for inspecting and debugging certificates

---

## Next Module

→ [Module 11: NAT & Firewalls](../11-nat-firewalls/01-nat-deep-dive.md)
