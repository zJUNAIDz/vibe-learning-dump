# 📡 Module 09: Wireshark for Developers

**Difficulty:** 🟡 Intermediate  
**Time:** 45 minutes

---

## What You'll Learn

- What Wireshark captures (and what it doesn't)
- TCP stream analysis
- Finding unencrypted data
- Understanding TLS handshakes
- What encrypted traffic still leaks

---

## What Is Wireshark?

**Wireshark is a packet capture tool that lets you see network traffic at the packet level.**

**Think of it as:** Reading the raw bytes traveling over the network.

---

## Installing Wireshark

```bash
# Fedora
sudo dnf install wireshark wireshark-cli

# Add user to wireshark group
sudo usermod -aG wireshark $USER

# Log out and back in
```

---

## Basic Capture

### Start Capture

1. Open Wireshark (GUI)
2. Select interface (e.g., `eth0`, `wlan0`)
3. Click **Start**

### What You See

```
No.  Time      Source          Destination     Protocol  Info
1    0.000000  192.168.1.100   93.184.216.34   TCP       52342 → 80 [SYN]
2    0.023415  93.184.216.34   192.168.1.100   TCP       80 → 52342 [SYN, ACK]
3    0.023500  192.168.1.100   93.184.216.34   TCP       52342 → 80 [ACK]
4    0.023600  192.168.1.100   93.184.216.34   HTTP      GET / HTTP/1.1
```

---

## Display Filters

**Filters let you focus on specific traffic.**

### Common Filters

```
# HTTP traffic only
http

# Traffic to/from specific IP
ip.addr == 192.168.1.100

# Traffic on port 80
tcp.port == 80

# HTTP POSTs
http.request.method == "POST"

# Contains specific string
frame contains "password"

# TLS handshake
ssl.handshake
```

---

## Following TCP Streams

**To see full conversation:**

1. Right-click a packet
2. **Follow → TCP Stream**
3. See entire request/response

---

### Example: HTTP Request

```http
GET /api/login HTTP/1.1
Host: example.com
User-Agent: curl/7.68.0
Accept: */*
```

**Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{"token": "abc123xyz"}
```

**Takeaway:** Unencrypted HTTP is completely visible.

---

## Analyzing HTTPS Traffic

### What You CAN'T See

- ❌ Request method (GET/POST)
- ❌ URL path
- ❌ Headers
- ❌ Body

### What You CAN See

- ✅ Destination IP
- ✅ Destination domain (SNI in TLS handshake)
- ✅ Packet sizes
- ✅ Timing

---

### Server Name Indication (SNI)

**During TLS handshake, client sends domain in plaintext.**

**Filter:** `ssl.handshake.extensions_server_name`

**Why it matters:** Even with HTTPS, observers know what sites you visit.

---

## Use Case: Finding Unencrypted Credentials

### Scenario: App uses HTTP for login

**Filter:** `http.request.method == "POST"`

**Follow TCP stream:**
```http
POST /login HTTP/1.1
Content-Type: application/x-www-form-urlencoded

username=alice&password=secret123
```

**Credentials captured in plaintext!**

---

## Use Case: Detecting Data Leaks

### Filter for Specific Strings

```
frame contains "password"
frame contains "ssn"
frame contains "secret"
```

**If these appear in HTTP → data leak.**

---

## Use Case: TLS Version Detection

**Filter:** `ssl.handshake.version`

**Check:**
- TLS 1.0/1.1 → Insecure, deprecated
- TLS 1.2/1.3 → Secure

---

## Command Line: tshark

**tshark = Wireshark CLI (for scripts/automation)**

### Basic Capture

```bash
# Capture on interface, save to file
sudo tshark -i eth0 -w capture.pcap

# Stop with Ctrl+C
```

### Read Capture File

```bash
tshark -r capture.pcap
```

### Filters

```bash
# HTTP traffic only
tshark -r capture.pcap -Y "http"

# Extract HTTP URLs
tshark -r capture.pcap -Y "http.request" -T fields -e http.request.full_uri
```

---

## Common Wireshark Findings

### 1. Unencrypted Traffic

**Symptom:** HTTP instead of HTTPS.

**Risk:** Credentials, tokens, PII visible.

---

### 2. Weak TLS Versions

**Symptom:** TLS 1.0 or 1.1.

**Risk:** Vulnerable to BEAST, POODLE attacks.

---

### 3. Large Data Transfers

**Symptom:** Unusually large packets.

**Risk:** Data exfiltration.

---

### 4. DNS Queries

**Symptom:** `dns.qry.name` contains unusual domains.

**Risk:** Malware calling home, DNS tunneling.

---

## What Wireshark Can't Do

### 1. Decrypt HTTPS (Without Keys)

**TLS encryption is strong.** Wireshark can't break it.

**Exception:** If you have the server's private key (rare, dangerous to export).

---

### 2. Capture on Switched Networks (Easily)

**Modern networks use switches, not hubs.**

**Switches:** Send packets only to intended recipient.

**Hubs:** Broadcast to everyone (easier to sniff).

**To capture others' traffic:**
- Need to be on same WiFi (monitor mode)
- Or use ARP spoofing (MITM attack)

---

## Security Posture Analysis

### What to Look For

```bash
# 1. Are you using HTTPS everywhere?
tshark -r capture.pcap -Y "http" | grep -v "https"

# 2. What TLS versions are in use?
tshark -r capture.pcap -Y "ssl.handshake.version" -T fields -e ssl.handshake.version | sort | uniq -c

# 3. What domains are contacted?
tshark -r capture.pcap -Y "dns" -T fields -e dns.qry.name | sort | uniq
```

---

## Wireshark for Debugging

### 1. Connection Issues

**Symptom:** Connection hangs.

**Check Wireshark:**
- Do you see SYN, SYN-ACK, ACK? (TCP handshake)
- Or just SYN with no response? (firewall blocking)

---

### 2. Slow Requests

**Symptom:** API calls are slow.

**Check Wireshark:**
- Time between request and response
- Are there retransmissions? (packet loss)

---

### 3. Unexpected Requests

**Symptom:** App behaves oddly.

**Check Wireshark:**
- Is app making extra API calls?
- Are third-party scripts loading?

---

## Summary

1. **Wireshark captures packets** at the network layer
2. **HTTP is completely visible** — always use HTTPS
3. **HTTPS hides content** but leaks metadata (domain, timing, size)
4. **SNI reveals domain** even with HTTPS
5. **tshark for automation** and filtering
6. **Use Wireshark to verify security** claims

---

## Exercises

### Exercise 1: Capture Your Traffic
1. Start Wireshark
2. Visit `http://neverssl.com` (intentionally HTTP)
3. Follow TCP stream
4. See your request in plaintext

### Exercise 2: Compare HTTP vs HTTPS
1. Visit `http://example.com`
2. Visit `https://example.com`
3. Compare what's visible in Wireshark

### Exercise 3: Find DNS Queries
```bash
tshark -i eth0 -Y "dns" -T fields -e dns.qry.name
```
What domains does your system contact?

---

## What's Next?

Now let's explore command-line security tools.

→ **Next: [Module 10: Linux CLI Security Tools](../10-linux-cli-tools/01-cli-tools.md)**

---

## Further Reading

- [Wireshark User's Guide](https://www.wireshark.org/docs/wsug_html/)
- [Wireshark Display Filters](https://wiki.wireshark.org/DisplayFilters)
