# Module 03 — libuv Internals

## Overview

Module 02 showed you the event loop from JavaScript's perspective. This module goes **beneath** — into libuv's C implementation, the thread pool, kernel async I/O mechanisms, and how libuv bridges the gap between your single-threaded JavaScript and the multi-threaded operating system.

## Lessons

| # | Lesson | Topics |
|---|--------|--------|
| 1 | [libuv Event Loop Implementation](./01-libuv-event-loop.md) | uv_run internals, handle/request lifecycle, loop alive semantics |
| 2 | [Thread Pool Deep Dive](./02-thread-pool.md) | Work queue, thread management, saturation, io_uring |
| 3 | [Kernel Async I/O](./03-kernel-async-io.md) | epoll, kqueue, IOCP — how the kernel does async |
| 4 | [libuv Handles and Watchers](./04-handles-watchers.md) | TCP handles, timer handles, signal watchers, ref/unref |
