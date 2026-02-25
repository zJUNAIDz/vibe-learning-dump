# Start Here — How to Use This Curriculum

---

## What This Is

This is a **deep, complete Computer Networks curriculum** — not a cheat sheet, not a textbook summary.

Every module is written as **lecture notes + lab manual combined**. You will:
- Understand the *why* behind every protocol
- See the *how* on a real Linux system
- Learn to *debug* like an engineer, not memorize like a student

---

## Prerequisites

You should already:
- Know what IP addresses, TCP, and DNS are (at a surface level)
- Be comfortable using a Linux terminal
- Have used `ping`, `curl`, or `ssh` at least once
- Understand basic programming (any language)

You do NOT need to:
- Know the OSI model by heart
- Have read any networking textbook cover-to-cover
- Know how routers work internally

---

## How to Navigate

### Go in order
The modules are numbered `00` through `20`. They build on each other.

- Module 00 resets your mental model
- Modules 01–05 cover the "infrastructure" layers
- Modules 06–07 are the most critical (transport + congestion)
- Modules 08–09 cover what you interact with daily
- Modules 10–14 are about security, tools, and analysis
- Modules 15–18 are advanced/practical
- Module 19 ties everything together
- Module 20 maps it all to cloud

### Run every command
If a module says "run this on your Linux machine" — do it. Reading about `tcpdump` is not the same as seeing packets scroll by.

### Don't skip the "why"
If you find yourself skipping paragraphs to get to the "answer" — slow down. The explanations ARE the answer.

---

## Time Estimate

| Modules | Time (focused study) |
|---------|---------------------|
| 00–05   | ~15–20 hours        |
| 06–07   | ~10–12 hours        |
| 08–10   | ~8–10 hours         |
| 11–14   | ~10–12 hours        |
| 15–18   | ~8–10 hours         |
| 19–20   | ~6–8 hours          |
| **Total** | **~60–70 hours**  |

---

## What You'll Be Able to Do After

- Explain how a packet travels from your browser to a server and back
- Debug DNS failures, connection timeouts, and packet loss
- Read `tcpdump` output and understand TCP state
- Reason about why a service is slow (latency vs loss vs congestion)
- Understand cloud networking as "normal networking with APIs"
- Configure Linux networking (namespaces, bridges, routes)
- Explain TLS, certificates, and trust chains
