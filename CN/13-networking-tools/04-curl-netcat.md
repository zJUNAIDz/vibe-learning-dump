# curl and netcat — HTTP Testing and Raw Socket Communication

> `curl` is your Swiss Army knife for HTTP/HTTPS testing. `netcat` (nc/ncat) is the raw socket tool for anything else — TCP, UDP, port scanning, file transfer, proxying. Together they let you test, debug, and interact with any network service from the command line.

---

## Table of Contents

1. [curl — HTTP from the Command Line](#curl)
2. [curl Basics](#curl-basics)
3. [HTTP Methods with curl](#methods)
4. [Headers and Authentication](#headers)
5. [TLS/SSL with curl](#tls)
6. [curl Performance and Timing](#timing)
7. [Advanced curl](#advanced-curl)
8. [netcat (nc/ncat) — Raw Sockets](#netcat)
9. [netcat Use Cases](#nc-uses)
10. [socat — netcat on Steroids](#socat)
11. [wget — Quick File Downloads](#wget)
12. [Practical Scenarios](#scenarios)
13. [Key Takeaways](#key-takeaways)

---

## curl — HTTP from the Command Line

```
curl = Client URL
  - Transfer data from/to servers
  - Supports HTTP, HTTPS, FTP, SMTP, IMAP, and many more
  - Follows redirects, handles cookies, supports auth
  - Available on virtually every Linux system
  - Used in scripts, CI/CD, monitoring, debugging
```

---

## curl Basics

```bash
# Simple GET request
curl https://example.com

# Show response headers
curl -I https://example.com            # HEAD request (headers only)
curl -i https://example.com            # Include headers in output

# Verbose (see full request/response including TLS handshake)
curl -v https://example.com
# > GET / HTTP/2
# > Host: example.com
# > User-Agent: curl/8.1.2
# > Accept: */*
# >
# < HTTP/2 200
# < content-type: text/html; charset=UTF-8
# < content-length: 1256

# Save to file
curl -o page.html https://example.com           # specified filename
curl -O https://example.com/file.tar.gz          # original filename

# Follow redirects
curl -L https://example.com            # follows 301/302 redirects
curl -L --max-redirs 5 https://example.com

# Silent mode (no progress bar)
curl -s https://example.com            # silent
curl -sS https://example.com           # silent but show errors

# Show only HTTP status code
curl -s -o /dev/null -w "%{http_code}" https://example.com
# 200
```

---

## HTTP Methods with curl

### GET

```bash
# Default method
curl https://api.example.com/users

# With query parameters
curl "https://api.example.com/users?page=2&limit=10"
# Quote the URL! & is special in shell
```

### POST

```bash
# Form data (application/x-www-form-urlencoded)
curl -X POST -d "username=john&password=secret" https://api.example.com/login

# JSON data
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"username": "john", "email": "john@example.com"}' \
  https://api.example.com/users

# Read body from file
curl -X POST \
  -H "Content-Type: application/json" \
  -d @data.json \
  https://api.example.com/users

# Read from stdin
echo '{"key": "value"}' | curl -X POST -H "Content-Type: application/json" -d @- https://api.example.com
```

### PUT, PATCH, DELETE

```bash
# PUT (replace resource)
curl -X PUT \
  -H "Content-Type: application/json" \
  -d '{"name": "John Updated"}' \
  https://api.example.com/users/123

# PATCH (partial update)
curl -X PATCH \
  -H "Content-Type: application/json" \
  -d '{"email": "new@example.com"}' \
  https://api.example.com/users/123

# DELETE
curl -X DELETE https://api.example.com/users/123
```

### File uploads

```bash
# Upload file (multipart/form-data)
curl -X POST -F "file=@photo.jpg" https://api.example.com/upload
curl -X POST -F "file=@photo.jpg" -F "description=My photo" https://api.example.com/upload

# Upload with specific content type
curl -X POST -F "file=@data.csv;type=text/csv" https://api.example.com/upload
```

---

## Headers and Authentication

### Custom headers

```bash
# Set headers
curl -H "Authorization: Bearer eyJhbG..." https://api.example.com/me
curl -H "X-Custom-Header: value" https://api.example.com/

# Multiple headers
curl \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN" \
  -H "Accept: application/json" \
  https://api.example.com/users
```

### Authentication

```bash
# Basic auth
curl -u username:password https://api.example.com/
# Equivalent to: -H "Authorization: Basic dXNlcm5hbWU6cGFzc3dvcmQ="

# Bearer token
curl -H "Authorization: Bearer eyJhbG..." https://api.example.com/

# Digest auth
curl --digest -u username:password https://api.example.com/
```

### Cookies

```bash
# Send cookie
curl -b "session=abc123" https://example.com/

# Save cookies to file (from Set-Cookie headers)
curl -c cookies.txt https://example.com/login -d "user=john&pass=secret"

# Use saved cookies
curl -b cookies.txt https://example.com/dashboard

# Save and use in one session
curl -c cookies.txt -b cookies.txt https://example.com/
```

---

## TLS/SSL with curl

```bash
# Show TLS negotiation details
curl -v https://example.com 2>&1 | grep -E '\*|<|>'
# * TLSv1.3 (OUT), TLS handshake, Client hello (1):
# * TLSv1.3 (IN), TLS handshake, Server hello (2):
# * TLSv1.3 (IN), TLS handshake, Encrypted Extensions (8):
# * TLSv1.3 (IN), TLS handshake, Certificate (11):
# * SSL certificate verify ok.
# * using HTTP/2

# Skip certificate verification (for testing ONLY)
curl -k https://self-signed.example.com
# NEVER use -k in production scripts!

# Use specific CA certificate
curl --cacert /path/to/ca.pem https://example.com

# Client certificate (mTLS)
curl --cert client.pem --key client-key.pem https://example.com

# Force TLS version
curl --tlsv1.2 https://example.com        # minimum TLS 1.2
curl --tlsv1.3 https://example.com        # minimum TLS 1.3
curl --tls-max 1.2 https://example.com    # maximum TLS 1.2

# Show certificate info
curl -v --silent https://example.com 2>&1 | grep -A 5 'Server certificate'
# * Server certificate:
# *  subject: C=US; ST=California; O=Internet Corporation for Assigned Names and Numbers; CN=www.example.org
# *  start date: Jan 30 00:00:00 2024 GMT
# *  expire date: Mar  1 23:59:59 2025 GMT
# *  issuer: C=US; O=DigiCert Inc; CN=DigiCert Global G2 TLS RSA SHA256 2020 CA1
```

---

## curl Performance and Timing

### Timing breakdown

```bash
# Detailed timing info
curl -w "\
    DNS:        %{time_namelookup}s\n\
    Connect:    %{time_connect}s\n\
    TLS:        %{time_appconnect}s\n\
    TTFB:       %{time_starttransfer}s\n\
    Total:      %{time_total}s\n\
    Size:       %{size_download} bytes\n\
    Speed:      %{speed_download} bytes/sec\n\
    HTTP Code:  %{http_code}\n" \
  -o /dev/null -s https://example.com

# Output:
#     DNS:        0.012s         ← DNS resolution time
#     Connect:    0.045s         ← TCP handshake complete
#     TLS:        0.098s         ← TLS handshake complete
#     TTFB:       0.156s         ← Time To First Byte
#     Total:      0.234s         ← Total request time
#     Size:       1256 bytes     ← Response body size
#     Speed:      5366 bytes/sec ← Download speed
#     HTTP Code:  200

# Breakdown:
#   DNS time:        12 ms
#   TCP handshake:   45 - 12 = 33 ms
#   TLS handshake:   98 - 45 = 53 ms
#   Server think:    156 - 98 = 58 ms
#   Data transfer:   234 - 156 = 78 ms
```

### HTTP/2 and HTTP/3

```bash
# Force HTTP/1.1
curl --http1.1 https://example.com

# Use HTTP/2
curl --http2 https://example.com

# Use HTTP/3 (QUIC) — requires curl built with HTTP/3 support
curl --http3 https://example.com

# Check which protocol was used
curl -w "Protocol: %{http_version}\n" -o /dev/null -s https://example.com
# Protocol: 2
```

---

## Advanced curl

### Resolve without DNS (test specific server)

```bash
# Force resolution to specific IP (bypass DNS)
curl --resolve example.com:443:93.184.216.34 https://example.com

# Test a staging server with production hostname
curl --resolve api.myapp.com:443:10.0.0.50 https://api.myapp.com/health

# Connect to specific host (similar but changes Host header behavior)
curl --connect-to example.com:443:backend.internal:8443 https://example.com
```

### Using a proxy

```bash
# HTTP proxy
curl -x http://proxy:8080 https://example.com

# SOCKS5 proxy
curl --socks5-hostname proxy:1080 https://example.com

# No proxy for specific hosts
curl --noproxy "localhost,127.0.0.1,.internal" https://example.com
```

### Rate limiting and retries

```bash
# Limit download speed
curl --limit-rate 1M https://example.com/largefile.tar.gz

# Retry on failure
curl --retry 3 --retry-delay 5 https://api.example.com/health

# Timeout
curl --connect-timeout 5 --max-time 30 https://example.com
# connect-timeout: TCP connection must be established in 5s
# max-time: entire operation must complete in 30s
```

---

## netcat (nc/ncat) — Raw Sockets

### What netcat is

```
netcat = the "network Swiss Army knife"
  - Create raw TCP/UDP connections
  - Listen for connections
  - Port scanning
  - File transfer
  - Chat / debugging
  - Proxy traffic

Variants:
  nc:     Original netcat (limited features)
  ncat:   Nmap project's netcat (recommended, SSL support)
  socat:  Advanced bidirectional relay (most powerful)
```

---

## netcat Use Cases

### Test if a port is open

```bash
# TCP port check
nc -zv 10.0.0.1 80
# Connection to 10.0.0.1 80 port [tcp/http] succeeded!
# -z = scan mode (don't send data)
# -v = verbose

# Multiple ports
nc -zv 10.0.0.1 80 443 8080

# Port range
nc -zv 10.0.0.1 1-1024

# UDP port check (less reliable — UDP doesn't confirm)
nc -zuv 10.0.0.1 53

# With timeout
nc -zv -w 3 10.0.0.1 80    # 3 second timeout
```

### Raw HTTP request

```bash
# Send HTTP request manually
echo -e "GET / HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n" | nc example.com 80

# Interactive HTTP
nc example.com 80
# Type:
# GET / HTTP/1.1
# Host: example.com
# Connection: close
#
# (blank line to send)
# → See raw HTTP response
```

### TCP listener (simple server)

```bash
# Listen on port 8080
nc -l -p 8080
# Waits for connection, prints received data to stdout

# Listen and respond
echo "Hello from server" | nc -l -p 8080

# Keep listening (restart after each connection)
while true; do echo "OK" | nc -l -p 8080; done
```

### File transfer

```bash
# On receiver:
nc -l -p 9999 > received_file.tar.gz

# On sender:
nc 10.0.0.2 9999 < file.tar.gz

# With progress (using pv)
pv file.tar.gz | nc 10.0.0.2 9999

# Compressed transfer
tar czf - /data | nc 10.0.0.2 9999          # sender
nc -l -p 9999 | tar xzf -                    # receiver
```

### Port forwarding / proxy

```bash
# Simple TCP proxy (forward local 8080 to remote 80)
ncat -l -p 8080 --sh-exec "ncat example.com 80"

# With socat (more reliable)
socat TCP-LISTEN:8080,fork TCP:example.com:80
```

### Chat between two machines

```bash
# Machine A (listener):
nc -l -p 4444

# Machine B (connector):
nc 10.0.0.1 4444

# Type in either terminal → appears in the other
# Ctrl+C to exit
```

### Banner grabbing

```bash
# See what service is running
nc -v 10.0.0.1 22
# SSH-2.0-OpenSSH_8.9p1 Ubuntu-3

nc -v 10.0.0.1 25
# 220 mail.example.com ESMTP Postfix

nc -v 10.0.0.1 3306
# MySQL binary protocol header...
```

---

## socat — netcat on Steroids

### What socat adds

```
socat = SOcket CAT
  - Bidirectional data relay between two data channels
  - SSL/TLS support
  - Unix domain sockets
  - IPv6
  - Process forking
  - Address rewriting
  - More reliable than nc for persistent connections
```

### Common socat commands

```bash
# TCP connection (like nc/telnet)
socat - TCP:example.com:80
# Type HTTP request manually

# TCP listener
socat TCP-LISTEN:8080,reuseaddr,fork STDOUT

# TCP proxy / port forward
socat TCP-LISTEN:8080,reuseaddr,fork TCP:backend:80
# Every connection to localhost:8080 → forwarded to backend:80

# TLS connection
socat - OPENSSL:example.com:443

# Connect to Unix domain socket
socat - UNIX-CONNECT:/var/run/docker.sock
# Then type: GET /version HTTP/1.1\r\nHost: localhost\r\n\r\n

# UDP relay
socat UDP-LISTEN:5000,fork UDP:dns-server:53

# File descriptor magic: TCP to Unix socket
socat TCP-LISTEN:8080,reuseaddr,fork UNIX-CONNECT:/var/run/app.sock
# Expose Unix socket as TCP (e.g., expose Docker API)
```

---

## wget — Quick File Downloads

```bash
# Download file
wget https://example.com/file.tar.gz

# Download to specific location
wget -O output.txt https://example.com/data

# Continue interrupted download
wget -c https://example.com/large-file.iso

# Recursive download (mirror website)
wget -r -l 2 https://example.com/docs/
# -r = recursive | -l 2 = depth limit 2

# Quiet mode
wget -q https://example.com/file.tar.gz

# Download with authentication
wget --user=admin --password=secret https://example.com/protected/

# Limit rate
wget --limit-rate=1m https://example.com/large-file.iso

# Spider mode (check links without downloading)
wget --spider https://example.com/page.html
```

### curl vs wget

```
curl:                           wget:
  HTTP tool (debug, API test)     Download tool
  Outputs to stdout               Saves to file
  Many protocols                  HTTP/HTTPS/FTP
  No recursive download           Recursive download
  Better for scripting             Better for mirroring
  Follows redirects with -L        Follows redirects by default
```

---

## Practical Scenarios

### Scenario: Test API endpoint

```bash
# Health check
curl -s https://api.example.com/health | jq .
# {"status": "ok", "version": "1.2.3"}

# Timed health check
curl -s -o /dev/null -w "%{http_code} %{time_total}s\n" https://api.example.com/health
# 200 0.156s

# POST with authentication
curl -s -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "test"}' \
  https://api.example.com/items | jq .
```

### Scenario: Verify TLS certificate

```bash
# Check certificate expiry
curl -vI https://example.com 2>&1 | grep "expire date"
# *  expire date: Mar  1 23:59:59 2025 GMT

# Test with specific TLS version
curl --tlsv1.2 --tls-max 1.2 https://example.com  # force TLS 1.2

# Verify certificate chain
curl -v https://example.com 2>&1 | grep -E "SSL|subject|issuer|expire"
```

### Scenario: Debug connectivity

```bash
# 1. Quick port check
nc -zv server.example.com 443 -w 3
# or
curl -v --connect-timeout 5 https://server.example.com 2>&1 | head -20

# 2. DNS check
curl -w "DNS: %{time_namelookup}s\n" -o /dev/null -s https://example.com

# 3. Full timing breakdown
curl -w "DNS: %{time_namelookup}\nTCP: %{time_connect}\nTLS: %{time_appconnect}\nTTFB: %{time_starttransfer}\nTotal: %{time_total}\n" \
  -o /dev/null -s https://example.com
```

---

## Key Takeaways

1. **curl is your HTTP debugging tool** — use `-v` for verbose, `-i` for headers, `-w` for timing
2. **curl timing breakdown** tells you exactly where time is spent: DNS, TCP, TLS, server processing, transfer
3. **Always use `jq`** with JSON APIs: `curl -s https://api.example.com/data | jq .`
4. **`curl --resolve`** lets you test a specific server without changing DNS — essential for migration testing
5. **`curl -k` skips TLS verification** — useful for testing but NEVER in production
6. **netcat tests raw TCP/UDP connectivity** — `nc -zv host port` answers "is the port open?"
7. **netcat creates raw connections** — send manual HTTP, grab banners, transfer files
8. **socat** is more powerful than netcat — TLS support, Unix sockets, bidirectional relays
9. **wget for downloads**, curl for API testing — they're complementary, not competing
10. **Combine tools**: `curl` for HTTP, `nc` for raw TCP, `tcpdump` for packet capture, `ss` for socket state

---

## Next Module

→ [Module 14: Packet Analysis](../14-packet-analysis/01-tcpdump-deep-dive.md) — Deep dive into packet capture and analysis
