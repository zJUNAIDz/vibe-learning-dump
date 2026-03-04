# Module 05 — Networking Internals

## Overview

Networking is the reason Node.js exists. Ryan Dahl created Node specifically because existing servers (Apache, Ruby) handled concurrent connections poorly — blocking a thread per connection. Node's event-driven, non-blocking architecture makes it handle thousands of concurrent connections on a single thread.

This module dissects how TCP and HTTP actually work inside Node.js, from JavaScript down to kernel syscalls.

---

## Architecture

```mermaid
graph TB
    subgraph "Your Code"
        HTTP[http.createServer]
        NET[net.createServer]
    end
    
    subgraph "Node.js Internals"
        PARSER[llhttp Parser]
        TCP[TCP Wrap C++]
        PIPE[Pipe Wrap C++]
    end
    
    subgraph "libuv"
        UVTCP[uv_tcp_t handle]
        UVSTREAM[uv_stream_t]
        UVPOLL[Poll Phase]
    end
    
    subgraph "Kernel"
        SOCKET[socket()]
        BIND[bind()]
        LISTEN[listen()]
        ACCEPT[accept()]
        EPOLL[epoll_wait()]
    end
    
    HTTP --> PARSER --> TCP
    NET --> TCP
    TCP --> UVTCP --> UVSTREAM --> UVPOLL
    UVPOLL --> EPOLL
    EPOLL --> SOCKET
    SOCKET --> BIND --> LISTEN --> ACCEPT
```

---

## Lessons

| # | Lesson | What You'll Learn |
|---|--------|-------------------|
| 01 | [TCP Lifecycle](01-tcp-lifecycle.md) | How sockets work from syscall to JavaScript |
| 02 | [Socket Internals](02-socket-internals.md) | File descriptors, Nagle's algorithm, TCP_NODELAY |
| 03 | [HTTP Parsing](03-http-parsing.md) | How llhttp parses HTTP, keep-alive, pipelining |
| 04 | [Raw TCP Server](04-raw-tcp-server.md) | Build a TCP server from scratch |
| 05 | [Minimal HTTP Server](05-minimal-http-server.md) | Build an HTTP server on raw TCP |

---

## Key Takeaways

- Every Node.js connection is a libuv handle wrapping a kernel file descriptor
- `net.Server` is the foundation — `http.Server` inherits from it
- Connections are accepted in the poll phase via `epoll_wait()`
- llhttp (written in C) parses HTTP at ~1.5 GB/s
- TCP_NODELAY disables Nagle's algorithm for low-latency responses
- The kernel backlog queue determines how many pending connections are buffered
