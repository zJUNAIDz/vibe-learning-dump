# Module 04 — Filesystem Internals

## Overview

When you call `fs.readFile()`, a complex chain of events unfolds: JavaScript → C++ binding → libuv thread pool → kernel syscall → disk controller → and back. This module traces that entire path and teaches you how to work with the filesystem efficiently.

## Lessons

| # | Lesson | Topics |
|---|--------|--------|
| 1 | [File Descriptors and Syscalls](./01-file-descriptors.md) | open/read/write/close, fd management, /proc/self/fd |
| 2 | [Buffers: The Memory Bridge](./02-buffers.md) | Buffer internals, allocation strategies, ArrayBuffer, zero-copy |
| 3 | [Streams for File I/O](./03-file-streams.md) | ReadStream, WriteStream, memory-efficient processing |
| 4 | [Labs: Build Real File Tools](./04-labs.md) | Implement tail -f, log processor, custom file stream |

## Key Architecture

```mermaid
graph TD
    subgraph "Your Code"
        API["fs.readFile('data.json')"]
    end
    
    subgraph "Node.js Internals"
        JS_LAYER["lib/fs.js<br/>Argument validation"]
        CPP["src/node_file.cc<br/>C++ binding"]
    end
    
    subgraph "libuv"
        UV_REQ["uv_fs_read()"]
        THREAD["Thread pool thread"]
    end
    
    subgraph "Kernel"
        VFS["Virtual File System"]
        CACHE["Page Cache"]
        DRIVER["Block Device Driver"]
    end
    
    subgraph "Hardware"
        DISK["SSD / HDD"]
    end
    
    API --> JS_LAYER --> CPP --> UV_REQ --> THREAD
    THREAD --> VFS --> CACHE
    CACHE --> |"Cache miss"| DRIVER --> DISK
    CACHE --> |"Cache hit"| THREAD
    
    style API fill:#3178c6,color:#fff
    style THREAD fill:#9c27b0,color:#fff
    style CACHE fill:#4caf50,color:#fff
```
