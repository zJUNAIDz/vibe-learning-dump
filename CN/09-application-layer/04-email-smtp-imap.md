# Email Protocols — SMTP, IMAP, and POP3

> Email is one of the oldest internet applications (predating the web by 20 years) and remains critical infrastructure. Understanding SMTP, IMAP, and POP3 reveals fundamental concepts about store-and-forward, relay systems, and protocol design that apply across networking.

---

## Table of Contents

1. [Email Architecture](#architecture)
2. [SMTP — Sending Email](#smtp)
3. [SMTP Transaction Walk-Through](#smtp-transaction)
4. [SMTP Authentication and Relay](#smtp-auth)
5. [IMAP — Reading Email](#imap)
6. [POP3 — The Simple Alternative](#pop3)
7. [IMAP vs POP3](#imap-vs-pop3)
8. [Email Security](#security)
9. [DNS Records for Email](#dns-records)
10. [Email Authentication — SPF, DKIM, DMARC](#email-auth)
11. [Linux: Email from the Command Line](#linux-email)
12. [Key Takeaways](#key-takeaways)

---

## Architecture

Email uses a **store-and-forward** architecture with multiple components:

```
Sender (Alice)                                      Recipient (Bob)
    │                                                    │
    ▼                                                    ▼
┌─────────┐     ┌──────────┐     ┌──────────┐     ┌─────────┐
│  MUA    │────→│  MSA/MTA │────→│  MTA     │────→│  MDA    │
│ (Mail   │SMTP │ (Alice's │SMTP │ (Bob's   │     │ (Mail   │
│  Client)│     │  server) │     │  server) │     │ Delivery│
└─────────┘     └──────────┘     └──────────┘     │  Agent) │
                                                   └────┬────┘
                                                        │
                                                   ┌────▼────┐
                                                   │ Mailbox │
                                                   │ Storage │
                                                   └────┬────┘
                                                        │
                                                   ┌────▼────┐
                                                   │  MUA    │
                                                   │ (Bob's  │◄── IMAP/POP3
                                                   │  Client)│
                                                   └─────────┘
```

| Component | Full Name | Role |
|-----------|-----------|------|
| **MUA** | Mail User Agent | Email client (Thunderbird, Gmail web, Outlook) |
| **MSA** | Mail Submission Agent | Receives email from MUA (port 587) |
| **MTA** | Mail Transfer Agent | Routes email between servers (port 25) |
| **MDA** | Mail Delivery Agent | Delivers to user's mailbox |

### The flow

```
1. Alice composes email in MUA (Thunderbird)
2. MUA sends to Alice's mail server via SMTP (port 587, authenticated)
3. Alice's MTA looks up MX record for bob@example.com
   → dig MX example.com → mail.example.com
4. Alice's MTA connects to Bob's MTA via SMTP (port 25)
5. Bob's MTA receives email, MDA delivers to Bob's mailbox
6. Bob's MUA retrieves email via IMAP (port 993) or POP3 (port 995)
```

---

## SMTP

SMTP (Simple Mail Transfer Protocol, RFC 5321) handles sending and relaying email.

### Ports

| Port | Protocol | Purpose |
|------|----------|---------|
| 25 | SMTP | Server-to-server relay (MTA to MTA) |
| 465 | SMTPS | SMTP over implicit TLS (deprecated, then un-deprecated) |
| 587 | Submission | Client-to-server (MUA to MSA, requires auth) |

### SMTP commands

```
EHLO     – Extended hello (identify client, list capabilities)
MAIL FROM: <sender@example.com>  – Envelope sender
RCPT TO: <recipient@example.com> – Envelope recipient
DATA     – Begin message body (ends with \r\n.\r\n)
QUIT     – End session
RSET     – Reset current transaction
VRFY     – Verify if user exists (usually disabled)
AUTH     – Authenticate (LOGIN, PLAIN, CRAM-MD5, etc.)
STARTTLS – Upgrade to encrypted connection
```

### SMTP reply codes

```
2xx – Success:
  220: Server ready
  250: OK (command successful)
  235: Authentication successful
  354: Start mail input (after DATA command)

3xx – Intermediate:
  334: Server challenge (during AUTH)

4xx – Temporary failure (retry later):
  421: Service not available (try later)
  450: Mailbox unavailable (temporary)
  451: Local error (try later)

5xx – Permanent failure (don't retry):
  500: Syntax error
  503: Bad sequence of commands
  550: Mailbox not found
  553: Invalid address
  554: Transaction failed
```

---

## SMTP Transaction

### Complete session

```
# Client connects to server on port 587

S: 220 mail.example.com ESMTP ready
C: EHLO client.example.com
S: 250-mail.example.com Hello
S: 250-SIZE 35882577
S: 250-8BITMIME
S: 250-STARTTLS
S: 250-AUTH PLAIN LOGIN
S: 250 HELP

# Upgrade to TLS
C: STARTTLS
S: 220 Ready to start TLS
   (TLS handshake occurs)

# Authenticate
C: AUTH PLAIN AHVzZXIAcGFzc3dvcmQ=
S: 235 Authentication successful

# Send email
C: MAIL FROM:<alice@example.com>
S: 250 OK

C: RCPT TO:<bob@example.com>
S: 250 OK

C: DATA
S: 354 End data with <CR><LF>.<CR><LF>

C: From: Alice <alice@example.com>
C: To: Bob <bob@example.com>
C: Subject: Hello Bob
C: Date: Mon, 01 Jan 2024 12:00:00 +0000
C: MIME-Version: 1.0
C: Content-Type: text/plain; charset=utf-8
C:
C: Hi Bob,
C: How are you doing?
C: .
S: 250 OK: message queued as ABC123

C: QUIT
S: 221 Bye
```

### Envelope vs headers

```
Envelope (SMTP commands):           Headers (inside DATA):
  MAIL FROM: alice@example.com       From: alice@example.com
  RCPT TO: bob@example.com          To: bob@example.com

These CAN differ!
  Envelope: who actually sent/receives
  Headers: what the user sees in their email client

Legitimate uses:
  - Mailing lists: envelope FROM is list bounce address
  - BCC: envelope has recipient, headers don't show them

Malicious uses:
  - Spoofing: envelope says attacker, headers say "Your Bank"
  - This is why we need SPF/DKIM/DMARC (covered later)
```

---

## SMTP Authentication and Relay

### Open relay — the original sin

Early SMTP had no authentication. Any server would relay email to anyone:

```
Spammer → any open relay → recipient
           No questions asked!

This is why email spam exists. Modern SMTP requires:
  Port 587: Authentication required (your mail server)
  Port 25:  Only accepts email FOR its own domains (no relaying)
```

### STARTTLS vs implicit TLS

```
STARTTLS (port 587):
  1. Connect unencrypted
  2. Client sends STARTTLS
  3. TLS handshake
  4. Everything after is encrypted
  
  Problem: Vulnerable to downgrade attack
    MITM removes "250-STARTTLS" from server capabilities
    Client doesn't know TLS is available
    Emails sent in plaintext!
    
  Fix: MTA-STS (RFC 8461) — publish policy saying "always use TLS"

Implicit TLS (port 465):
  1. Connect directly with TLS (like HTTPS)
  2. No unencrypted phase
  3. No downgrade possible
```

---

## IMAP

IMAP (Internet Message Access Protocol, RFC 9051) is the protocol for reading email stored on a server.

### Key design: server-side storage

```
IMAP philosophy: Email lives ON THE SERVER.
  - Multiple devices see the same mailbox
  - Folders, read/unread status synced across devices
  - Search happens on server
  - Only download messages when needed
```

### IMAP ports

| Port | Protocol |
|------|----------|
| 143 | IMAP (STARTTLS) |
| 993 | IMAPS (implicit TLS) |

### IMAP concepts

```
Mailbox:  A folder (INBOX, Sent, Drafts, custom folders)
Message:  An individual email
UID:      Unique ID for each message (persistent across sessions)
Sequence: Temporary message number (can change between sessions)
Flags:    \Seen, \Answered, \Flagged, \Deleted, \Draft
```

### IMAP session example

```
S: * OK IMAP4rev2 server ready
C: a001 LOGIN alice password123
S: a001 OK LOGIN completed

C: a002 LIST "" "*"
S: * LIST (\HasNoChildren) "/" "INBOX"
S: * LIST (\HasNoChildren) "/" "Sent"
S: * LIST (\HasNoChildren) "/" "Drafts"
S: * LIST (\HasNoChildren) "/" "Trash"
S: a002 OK LIST completed

C: a003 SELECT INBOX
S: * 17 EXISTS           (17 messages in INBOX)
S: * 2 RECENT            (2 new since last check)
S: * OK [UIDVALIDITY 3857529045]
S: * OK [UIDNEXT 4392]
S: * FLAGS (\Answered \Flagged \Deleted \Seen \Draft)
S: a003 OK [READ-WRITE] SELECT completed

# Fetch headers of last 5 messages
C: a004 FETCH 13:17 (FLAGS ENVELOPE)
S: * 13 FETCH (FLAGS (\Seen) ENVELOPE ("Mon, 01 Jan..." ...))
S: * 14 FETCH (FLAGS () ENVELOPE ("Tue, 02 Jan..." ...))
...

# Fetch full body of message 14
C: a005 FETCH 14 BODY[]
S: * 14 FETCH (BODY[] {2345}
S: (full email content including headers and body)
S: )
S: a005 OK FETCH completed

# Mark as read
C: a006 STORE 14 +FLAGS (\Seen)
S: * 14 FETCH (FLAGS (\Seen))
S: a006 OK STORE completed

# Move to trash
C: a007 MOVE 14 "Trash"
S: a007 OK MOVE completed

C: a008 LOGOUT
S: * BYE IMAP4rev2 server terminating
S: a008 OK LOGOUT completed
```

### IMAP IDLE — push notifications

```
# Instead of polling for new messages:
C: a009 IDLE
S: + idling
   (connection held open...)
   (5 minutes later, new email arrives)
S: * 18 EXISTS            (server pushes notification!)
C: DONE                   (client exits IDLE to fetch)

# This is how email clients show instant notifications
# Without IDLE: client polls every 1-5 minutes
```

---

## POP3

POP3 (Post Office Protocol v3, RFC 1939) is the simpler, older alternative to IMAP.

### Key design: client-side storage

```
POP3 philosophy: Download email, delete from server.
  - Email lives on YOUR DEVICE
  - Once downloaded, server copy optionally deleted
  - No folder sync between devices
  - Simple protocol, limited features
```

### POP3 ports

| Port | Protocol |
|------|----------|
| 110 | POP3 (STARTTLS) |
| 995 | POP3S (implicit TLS) |

### POP3 session

```
S: +OK POP3 server ready
C: USER alice
S: +OK
C: PASS password123
S: +OK Logged in

C: STAT
S: +OK 3 45678            (3 messages, 45678 bytes total)

C: LIST
S: +OK 3 messages
S: 1 12345
S: 2 23456
S: 3 9877
S: .

C: RETR 1                 (download message 1)
S: +OK 12345 octets
S: (full email content)
S: .

C: DELE 1                 (delete message 1 from server)
S: +OK Deleted

C: QUIT
S: +OK Bye
```

---

## IMAP vs POP3

| Feature | IMAP | POP3 |
|---------|------|------|
| Email storage | Server | Client (device) |
| Multi-device access | Yes (synced) | No (download to one device) |
| Folders | Yes (server-side) | No |
| Search | Server-side | Client-only |
| Offline access | Partial (download on demand) | Full (all downloaded) |
| Disk usage (server) | High (all email stored) | Low (deleted after download) |
| Disk usage (client) | Low (headers only initially) | High (all email downloaded) |
| Push notifications | IDLE command | No (polling only) |
| Bandwidth | Lower (fetch on demand) | Higher (download everything) |
| Complexity | High | Simple |

### When to use each

```
IMAP:
  ✓ Multiple devices (phone, laptop, web)
  ✓ Large mailboxes
  ✓ Keep email accessible from anywhere
  ✓ Server-side search
  
POP3:
  ✓ Single device
  ✓ Offline-first usage
  ✓ Privacy (email not stored on server)
  ✓ Limited server storage
```

---

## Security

### The email security problem

```
Original SMTP (1982): No encryption, no authentication, no integrity
  → Anyone can read email in transit (plaintext)
  → Anyone can forge the sender (spoofing)
  → Anyone can modify email in transit (tampering)
  → Anyone can relay through any server (spam)

Modern fixes:
  Transport: TLS (STARTTLS / implicit TLS)
  Authentication: SPF, DKIM, DMARC
  End-to-end: S/MIME, PGP (rarely used)
```

### Transport encryption (TLS)

```
MUA → MSA: Port 587/465 with TLS (user's mail client to server)
MTA → MTA: Port 25 with STARTTLS (server to server)

Problem: STARTTLS is opportunistic
  If TLS fails or is stripped by MITM → fallback to plaintext
  
Solutions:
  MTA-STS: Server publishes policy: "always require TLS"
  DANE/TLSA: DNS records specify expected TLS certificate
  SMTP TLS Reporting: Servers report TLS failures
```

---

## DNS Records for Email

Email delivery depends heavily on DNS:

### MX records — where to deliver

```bash
dig MX gmail.com
# gmail-smtp-in.l.google.com.  5
# alt1.gmail-smtp-in.l.google.com.  10
# alt2.gmail-smtp-in.l.google.com.  20

# Priority: lower number = preferred
# If priority 5 server is down → try priority 10 → try priority 20
```

### The A/AAAA record fallback

```
If NO MX record exists:
  Fall back to A/AAAA record of the domain itself
  (Not recommended — always set MX records)
```

### Null MX — "I don't accept email"

```
example.com. MX 0 .

A single dot as the MX host means:
  "This domain does NOT accept email"
  Prevents backscatter spam to domains without mail servers
```

---

## Email Authentication

### SPF (Sender Policy Framework)

```
Problem: Anyone can claim to send email from your domain.
Solution: Publish which servers ARE allowed to send.

DNS TXT record:
  example.com. TXT "v=spf1 ip4:203.0.113.0/24 include:_spf.google.com -all"

Meaning:
  v=spf1              → SPF version 1
  ip4:203.0.113.0/24  → These IPs can send as example.com
  include:_spf.google.com → Also allow Google's servers
  -all                → Reject all others (hard fail)

Qualifiers:
  +  Pass (allow)     — default
  -  Hard fail (reject)
  ~  Soft fail (accept but mark suspicious)
  ?  Neutral (no policy)
```

```bash
# Check SPF record
dig TXT example.com | grep spf

# Check SPF for Gmail
dig TXT _spf.google.com
```

### DKIM (DomainKeys Identified Mail)

```
Problem: SPF only checks the sending IP, not the message content.
         Email can be modified in transit.
Solution: Cryptographically sign the email headers and body.

How it works:
1. Sending server signs email with private key
2. Signature added as DKIM-Signature header
3. Receiving server gets public key from DNS
4. Receiving server verifies signature
5. If valid → email wasn't modified in transit

DNS record:
  selector._domainkey.example.com. TXT "v=DKIM1; k=rsa; p=MIGfMA0G..."

Email header:
  DKIM-Signature: v=1; a=rsa-sha256; d=example.com; s=selector;
    h=from:to:subject:date; bh=<body hash>; b=<signature>
```

```bash
# Check DKIM record
dig TXT google._domainkey.example.com

# Verify: Look at email headers for "DKIM-Signature"
# and "Authentication-Results: dkim=pass"
```

### DMARC (Domain-based Message Authentication, Reporting & Conformance)

```
Problem: SPF and DKIM exist, but what should receivers DO when they fail?
Solution: DMARC tells receivers the policy.

DNS TXT record:
  _dmarc.example.com. TXT "v=DMARC1; p=reject; rua=mailto:dmarc@example.com"

Meaning:
  v=DMARC1           → DMARC version 1
  p=reject           → Reject emails that fail SPF AND DKIM
  rua=mailto:...     → Send aggregate reports to this address

Policies:
  p=none     → Monitor only (collect reports, don't act)
  p=quarantine → Mark as spam
  p=reject   → Reject outright

DMARC alignment: The "From" header domain must match SPF/DKIM domain
  SPF alignment:  MAIL FROM domain matches From header domain
  DKIM alignment: DKIM d= domain matches From header domain
```

```bash
# Check DMARC record
dig TXT _dmarc.gmail.com
# "v=DMARC1; p=none; sp=quarantine; rua=mailto:..."
```

### How SPF + DKIM + DMARC work together

```
Email arrives claiming to be from alice@example.com:

1. SPF check: Is sending IP authorized by example.com's SPF record?
   → Pass or fail

2. DKIM check: Does the DKIM signature verify with example.com's public key?
   → Pass or fail

3. DMARC check: 
   a) Does at least one (SPF or DKIM) pass AND align with From domain?
   b) What policy does example.com's DMARC record specify?
   → Apply policy (none/quarantine/reject)

Result:
  SPF pass + DKIM pass + aligned = ✅ Deliver
  SPF fail + DKIM pass + aligned = ✅ Deliver (one pass sufficient)
  SPF fail + DKIM fail           = Apply DMARC policy (reject/quarantine)
```

---

## Linux: Email on Command Line

### Send email with telnet (raw SMTP for learning)

```bash
# Connect to a mail server (port 25, unencrypted — for testing only!)
telnet mail.example.com 25

# Or with openssl for TLS:
openssl s_client -starttls smtp -connect mail.example.com:587
openssl s_client -connect mail.example.com:465    # implicit TLS

# Then type SMTP commands manually (as shown in transaction section)
```

### Send email with swaks (Swiss Army Knife for SMTP)

```bash
# Install
sudo apt install swaks

# Send test email
swaks --to bob@example.com \
      --from alice@example.com \
      --server smtp.example.com:587 \
      --tls \
      --auth LOGIN \
      --auth-user alice@example.com \
      --auth-password "secret"

# Test with specific EHLO
swaks --to admin@example.com \
      --server mail.example.com:25 \
      --ehlo mytest.example.com

# Attach a file
swaks --to bob@example.com \
      --attach /path/to/file.pdf \
      --server smtp.example.com:587 --tls
```

### Check email headers

```bash
# View raw email headers (reveals authentication results)
# In Gmail: Open email → Three dots → "Show original"
# The Authentication-Results header shows SPF/DKIM/DMARC results:

Authentication-Results: mx.google.com;
       dkim=pass header.d=example.com;
       spf=pass smtp.mailfrom=example.com;
       dmarc=pass (p=REJECT sp=REJECT) header.from=example.com
```

### DNS lookups for email debugging

```bash
# MX records — where does email go?
dig MX example.com +short

# SPF record
dig TXT example.com +short | grep spf

# DKIM record (need to know the selector)
dig TXT selector._domainkey.example.com +short

# DMARC policy
dig TXT _dmarc.example.com +short

# Reverse DNS (PTR) — does the sending IP have valid rDNS?
dig -x 203.0.113.42 +short
# Missing PTR record is a spam indicator!

# Full email deliverability check
dig MX example.com && dig TXT example.com && dig TXT _dmarc.example.com
```

---

## Key Takeaways

1. **Email is store-and-forward**: Messages hop from MUA → MSA → MTA → MTA → MDA → mailbox → MUA
2. **SMTP (port 25/587)** sends email; **IMAP (port 993)** and **POP3 (port 995)** retrieve it
3. **Envelope vs headers**: SMTP envelope (MAIL FROM/RCPT TO) can differ from email headers — basis of spoofing
4. **IMAP keeps email on server** (multi-device sync); **POP3 downloads to device** (single device, offline)
5. **IMAP IDLE** enables push notifications — server tells client when new email arrives
6. **STARTTLS is vulnerable** to downgrade attacks — use MTA-STS to enforce TLS
7. **SPF** = "which IPs can send for my domain" (DNS TXT record)
8. **DKIM** = "this email was signed by my domain" (cryptographic signature)
9. **DMARC** = "what to do when SPF/DKIM fail" (policy: none/quarantine/reject)
10. **All three (SPF + DKIM + DMARC)** are needed for proper email authentication — one alone is insufficient

---

## Next Module

→ [Module 10: TLS/SSL](../10-tls-ssl/01-why-encryption.md) — How encryption protects everything we've discussed
