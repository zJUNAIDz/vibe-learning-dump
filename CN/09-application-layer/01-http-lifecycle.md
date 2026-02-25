# HTTP Lifecycle — From URL to Rendered Page

> Every click, every API call, every web request follows the same lifecycle: DNS → TCP → TLS → HTTP request → HTTP response. Understanding this end-to-end flow is essential for debugging latency, failures, and performance.

---

## Table of Contents

1. [The Complete Lifecycle](#complete-lifecycle)
2. [HTTP Fundamentals](#fundamentals)
3. [HTTP Request Format](#request-format)
4. [HTTP Response Format](#response-format)
5. [HTTP Methods (Verbs)](#methods)
6. [Status Codes](#status-codes)
7. [Headers That Matter](#headers)
8. [HTTP/1.0 vs HTTP/1.1](#http10-vs-11)
9. [Connection Management](#connection-management)
10. [Cookies and State](#cookies)
11. [Linux: HTTP from the Command Line](#linux-http)
12. [Key Takeaways](#key-takeaways)

---

## Complete Lifecycle

What happens when you type `https://example.com/page` and press Enter:

```
1. URL parsing
   Browser extracts: scheme=https, host=example.com, path=/page, port=443

2. DNS resolution (Module 08)
   example.com → 93.184.216.34  (~50ms uncached, <1ms cached)

3. TCP handshake (Module 06)
   SYN → SYN-ACK → ACK  (~15ms per RTT)

4. TLS handshake (Module 10)
   ClientHello → ServerHello → keys → Finished  (~30ms, 1-2 RTTs)

5. HTTP request
   GET /page HTTP/1.1
   Host: example.com

6. Server processing
   Server receives request → application logic → generates response

7. HTTP response
   HTTP/1.1 200 OK
   Content-Type: text/html
   [HTML body]

8. Rendering
   Browser parses HTML → discovers CSS/JS/images → makes more requests
   → renders page

Total time for first byte: DNS + TCP + TLS + server processing
                         ≈ 50 + 15 + 30 + 50 = 145ms (typical)
```

### Time breakdown (real-world)

```bash
# Measure each phase with curl
curl -o /dev/null -w "\
DNS:        %{time_namelookup}s\n\
TCP:        %{time_connect}s\n\
TLS:        %{time_appconnect}s\n\
First byte: %{time_starttransfer}s\n\
Total:      %{time_total}s\n" \
https://example.com

# Example output:
# DNS:        0.012s
# TCP:        0.028s     (DNS + TCP handshake)
# TLS:        0.058s     (DNS + TCP + TLS handshake)
# First byte: 0.145s     (DNS + TCP + TLS + server processing)
# Total:      0.152s     (until last byte received)
```

---

## Fundamentals

HTTP (Hypertext Transfer Protocol) is a **request-response** protocol:

```
Client sends:  REQUEST  (method + URL + headers + optional body)
Server sends:  RESPONSE (status code + headers + optional body)
```

### Key properties

| Property | Description |
|----------|-------------|
| **Text-based** | Headers and request line are human-readable ASCII text |
| **Stateless** | Each request is independent — server doesn't remember previous requests |
| **Client-initiated** | Client always sends the request first (server can't push unsolicited) |
| **Application-layer** | Sits on top of TCP (or QUIC for HTTP/3) |

### HTTP is stateless — why?

Statelessness simplifies server design. The server doesn't need to track which client is on which page. Each request contains everything needed to process it. This enables:
- Easy horizontal scaling (any server can handle any request)
- Simple recovery from server crashes
- Caching at any intermediate point

State is added back via cookies, tokens, and sessions when needed.

---

## Request Format

```
METHOD /path HTTP/version\r\n
Header1: value1\r\n
Header2: value2\r\n
\r\n
[Optional body]
```

### Real example

```
GET /api/users?page=2 HTTP/1.1\r\n
Host: api.example.com\r\n
Accept: application/json\r\n
Authorization: Bearer eyJhbGc...\r\n
User-Agent: curl/7.88.1\r\n
\r\n
```

### Components

| Part | Example | Purpose |
|------|---------|---------|
| **Method** | GET | What action to perform |
| **Request URI** | /api/users?page=2 | Which resource |
| **HTTP Version** | HTTP/1.1 | Protocol version |
| **Headers** | Host: api.example.com | Metadata about the request |
| **Body** | {"name":"John"} | Data payload (for POST/PUT/PATCH) |

### The Host header is mandatory in HTTP/1.1

Multiple websites share the same IP address (virtual hosting). The Host header tells the server which site you want:

```
Same IP (93.184.216.34) hosts:
  Host: example.com     → serves example.com
  Host: example.org     → serves example.org
  Host: example.net     → serves example.net
```

---

## Response Format

```
HTTP/version StatusCode ReasonPhrase\r\n
Header1: value1\r\n
Header2: value2\r\n
\r\n
[Optional body]
```

### Real example

```
HTTP/1.1 200 OK\r\n
Content-Type: application/json\r\n
Content-Length: 245\r\n
Cache-Control: max-age=3600\r\n
Date: Mon, 01 Jan 2024 12:00:00 GMT\r\n
\r\n
{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]}
```

---

## Methods

| Method | Purpose | Body? | Idempotent? | Safe? |
|--------|---------|-------|-------------|-------|
| **GET** | Retrieve resource | No | Yes | Yes |
| **POST** | Create resource / submit data | Yes | No | No |
| **PUT** | Replace resource entirely | Yes | Yes | No |
| **PATCH** | Partially update resource | Yes | No | No |
| **DELETE** | Remove resource | Optional | Yes | No |
| **HEAD** | GET without body (headers only) | No | Yes | Yes |
| **OPTIONS** | Describe available methods (CORS preflight) | No | Yes | Yes |

### Idempotent vs Safe

**Safe**: No side effects. Calling it doesn't change anything on the server. GET, HEAD, OPTIONS.

**Idempotent**: Calling it once = calling it 10 times (same result). GET, PUT, DELETE, HEAD, OPTIONS.

```
GET /user/42       ← safe + idempotent (just reading)
DELETE /user/42    ← idempotent (deleting twice → same as deleting once)
POST /users        ← NOT idempotent (posting twice → two new users)
```

Why this matters: If a network error occurs during a PUT or DELETE, the client can retry safely. For POST, retrying might create duplicates → use idempotency keys.

---

## Status Codes

### Categories

| Range | Category | Meaning |
|-------|----------|---------|
| **1xx** | Informational | Request received, continue |
| **2xx** | Success | Request processed successfully |
| **3xx** | Redirection | Further action needed (follow redirect) |
| **4xx** | Client Error | Problem with the request |
| **5xx** | Server Error | Server failed to process valid request |

### Must-know status codes

```
200 OK                 ← Success
201 Created            ← Resource created (POST success)
204 No Content         ← Success, no body (DELETE success)

301 Moved Permanently  ← Resource moved, update bookmarks/cache
302 Found              ← Temporary redirect
304 Not Modified       ← Cached version is still valid (conditional GET)

400 Bad Request        ← Malformed request (wrong JSON, missing fields)
401 Unauthorized       ← Not authenticated (no/invalid credentials)
403 Forbidden          ← Authenticated but not authorized
404 Not Found          ← Resource doesn't exist
405 Method Not Allowed ← Wrong HTTP method (POST to a GET-only endpoint)
408 Request Timeout    ← Server waited too long for request
409 Conflict           ← Conflicts with current state (concurrent update)
429 Too Many Requests  ← Rate limited
431 Request Header Fields Too Large ← Headers too big

500 Internal Server Error ← Generic server failure
502 Bad Gateway        ← Proxy/load balancer can't reach backend
503 Service Unavailable ← Server overloaded or maintenance
504 Gateway Timeout    ← Proxy/load balancer: backend didn't respond in time
```

### Status codes for debugging

```
502 vs 504:
  502: LB received invalid response from backend (app crashed, sent garbage)
  504: LB didn't receive ANY response from backend (app hung, unreachable)

401 vs 403:
  401: "Who are you?" (no authentication or invalid token)
  403: "I know who you are, but you can't do this" (lacks permission)

301 vs 302:
  301: Cache the redirect (permanent — browser won't check again)
  302: Don't cache (temporary — check again next time)
```

---

## Headers

### Request headers

```
Host: api.example.com              # REQUIRED — which virtual host
Accept: application/json           # What content types client accepts
Content-Type: application/json     # What format the body is in
Authorization: Bearer <token>      # Authentication credentials
User-Agent: Mozilla/5.0            # Client identification
Cookie: session=abc123             # Stored cookies
Cache-Control: no-cache            # Caching directives
Accept-Encoding: gzip, br          # Compression algorithms supported
Connection: keep-alive             # TCP connection management
If-None-Match: "etag123"           # Conditional request (304 if unchanged)
If-Modified-Since: Mon, 01 Jan...  # Conditional request (time-based)
```

### Response headers

```
Content-Type: application/json     # Body format
Content-Length: 245                 # Body size in bytes
Content-Encoding: gzip             # Compression used
Cache-Control: max-age=3600        # Caching rules
ETag: "abc123"                     # Resource version (for conditional requests)
Last-Modified: Mon, 01 Jan...      # When resource last changed
Set-Cookie: session=xyz; HttpOnly  # Set a cookie on client
Location: /new-url                 # Redirect destination (with 3xx)
Retry-After: 60                    # When to retry (with 429/503)
Access-Control-Allow-Origin: *     # CORS — which origins allowed
```

### Content negotiation

```
Client: Accept: application/json, text/html;q=0.9, */*;q=0.8
         → "I prefer JSON, then HTML, then anything"

Server: Content-Type: application/json
         → "Here you go, JSON it is"
```

---

## HTTP/1.0 vs HTTP/1.1

### HTTP/1.0 (1996)

- One request per TCP connection
- After response, connection closes
- Each request = DNS + TCP + TLS + request-response

```
Request 1: TCP connect → GET /page → response → TCP close
Request 2: TCP connect → GET /style.css → response → TCP close
Request 3: TCP connect → GET /script.js → response → TCP close
(3 TCP handshakes = expensive!)
```

### HTTP/1.1 (1997, still widely used)

- **Persistent connections** (keep-alive): Multiple requests on one TCP connection
- **Pipelining** (rarely used): Send multiple requests without waiting for responses
- **Host header required**: Virtual hosting support
- **Chunked transfer encoding**: Send body in chunks (unknown size upfront)

```
TCP connect → GET /page → response
            → GET /style.css → response
            → GET /script.js → response
            → TCP close (after idle timeout)
(1 TCP handshake for all requests!)
```

---

## Connection Management

### Keep-alive

HTTP/1.1 connections are persistent by default:

```
Connection: keep-alive    (default in HTTP/1.1)
Connection: close         (close after this request)
```

```bash
# Demonstrate keep-alive
curl -v --http1.1 https://example.com https://example.com
# Second request reuses the TCP connection
# Look for: "Re-using existing connection"
```

### Head-of-line blocking (HTTP/1.1's problem)

Even with persistent connections, HTTP/1.1 requires responses IN ORDER:

```
Client sends:  Request A (large file), Request B (small icon)
Server returns: Response A first (slow!), THEN Response B

Even though B is ready, it waits behind A → head-of-line blocking

Browser workaround: Open 6 parallel TCP connections per hostname
(But each connection = TCP + TLS overhead)
```

This is the main problem HTTP/2 solves (multiplexing).

### Chunked transfer encoding

When the server doesn't know the response size upfront:

```
HTTP/1.1 200 OK
Transfer-Encoding: chunked

1a\r\n                          (26 bytes chunk)
This is the first chunk.\n\r\n
1c\r\n                          (28 bytes chunk)
And this is the second one.\n\r\n
0\r\n                           (0 = last chunk)
\r\n
```

Used for: Streaming responses, server-sent events, dynamically generated content.

---

## Cookies

HTTP is stateless, but applications need state (login sessions, shopping carts, preferences). Cookies add state to stateless HTTP.

### How cookies work

```
1. Server sets cookie in response:
   Set-Cookie: session=abc123; Path=/; HttpOnly; Secure; SameSite=Lax

2. Browser stores the cookie.

3. Browser sends cookie in subsequent requests:
   Cookie: session=abc123

4. Server reads cookie → identifies user → serves personalized response
```

### Cookie attributes

| Attribute | Purpose |
|-----------|---------|
| `HttpOnly` | JavaScript can't access (prevents XSS cookie theft) |
| `Secure` | Only sent over HTTPS |
| `SameSite=Strict` | Not sent in cross-site requests (CSRF protection) |
| `SameSite=Lax` | Sent for top-level navigation, not embedded requests |
| `SameSite=None` | Sent in all contexts (requires Secure) |
| `Domain=.example.com` | Sent to all subdomains |
| `Path=/api` | Only sent for /api paths |
| `Max-Age=3600` | Expires in 3600 seconds |
| `Expires=<date>` | Expires at specific date |

```bash
# View cookies for a site in curl
curl -v https://example.com 2>&1 | grep "Set-Cookie"

# Send cookies
curl -b "session=abc123" https://example.com

# Store and replay cookies
curl -c cookies.txt https://example.com/login -d "user=alice&pass=secret"
curl -b cookies.txt https://example.com/dashboard
```

---

## Linux: HTTP on Command Line

### curl — the Swiss Army knife

```bash
# Simple GET
curl https://api.example.com/users

# With headers
curl -H "Authorization: Bearer token123" https://api.example.com/users

# POST with JSON
curl -X POST -H "Content-Type: application/json" \
     -d '{"name":"Alice","email":"alice@example.com"}' \
     https://api.example.com/users

# PUT
curl -X PUT -H "Content-Type: application/json" \
     -d '{"name":"Alice Updated"}' \
     https://api.example.com/users/42

# DELETE
curl -X DELETE https://api.example.com/users/42

# Verbose output (see headers, TLS)
curl -v https://example.com

# Only headers
curl -I https://example.com

# Follow redirects
curl -L https://example.com

# Download file
curl -O https://example.com/file.tar.gz

# Upload file
curl -X POST -F "file=@local-file.txt" https://example.com/upload

# Timing
curl -o /dev/null -s -w "Total: %{time_total}s\n" https://example.com
```

### httpie — human-friendly HTTP

```bash
# Install
sudo apt install httpie

# GET (auto-detects JSON)
http https://api.example.com/users

# POST with JSON (implicit with key=value syntax)
http POST https://api.example.com/users name=Alice email=alice@example.com

# Headers
http https://api.example.com/users "Authorization: Bearer token123"
```

### nc (netcat) — raw HTTP

```bash
# Manually construct an HTTP request
echo -e "GET / HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n" | nc example.com 80
```

---

## Key Takeaways

1. **HTTP lifecycle**: DNS → TCP → TLS → Request → Response → Render
2. **Stateless protocol**: Each request is independent; state added via cookies/tokens
3. **Request = Method + URL + Headers + Body**; **Response = Status + Headers + Body**
4. **Status codes**: 2xx success, 3xx redirect, 4xx client error, 5xx server error
5. **HTTP/1.1 adds persistent connections** (keep-alive) — reuse TCP connections
6. **Head-of-line blocking** in HTTP/1.1: responses must be in order (HTTP/2 fixes this)
7. **Host header enables virtual hosting** — multiple sites on one IP
8. **Cookies add state** — always use HttpOnly, Secure, SameSite for security
9. **`curl -v`** is your best friend for HTTP debugging
10. **Measure timing** with `curl -w` to identify which phase is slow

---

## Next

→ [02-http2-http3.md](02-http2-http3.md) — Modern HTTP: multiplexing, streams, and QUIC
