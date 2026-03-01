# The Three Pillars of Observability

> **Logs tell you WHAT happened. Metrics tell you HOW MUCH. Traces tell you WHERE in the system. You need all three to debug production.**

---

## 🟢 Why Observability?

```
Monitoring: "Is the system up?"
  → Dashboard shows green/red
  → Alerts when things break
  → Answers KNOWN questions

Observability: "WHY is the system slow for users in Europe?"
  → Traces show request flow
  → Metrics show latency spike at 14:00
  → Logs reveal database timeout in the payment service
  → You didn't predict this question. You can still answer it.

Monitoring is a subset of observability.
Observability lets you ask QUESTIONS YOU DIDN'T ANTICIPATE.
```

---

## 🟢 The Three Pillars

```
┌─────────────────────────────────────────────────────────┐
│                   OBSERVABILITY                          │
│                                                         │
│  ┌──────────┐   ┌──────────────┐   ┌────────────────┐  │
│  │   LOGS   │   │   METRICS    │   │    TRACES      │  │
│  │          │   │              │   │                │  │
│  │ What     │   │ How much     │   │ Where the      │  │
│  │ happened │   │ How fast     │   │ request went   │  │
│  │          │   │ How many     │   │                │  │
│  │ Text     │   │ Numbers      │   │ Request flow   │  │
│  │ events   │   │ over time    │   │ across services│  │
│  │          │   │              │   │                │  │
│  │ "User    │   │ CPU: 82%     │   │ API→Auth→DB   │  │
│  │  login   │   │ Latency:     │   │ 200ms  50ms   │  │
│  │  failed" │   │  p99 = 340ms │   │   150ms       │  │
│  └──────────┘   └──────────────┘   └────────────────┘  │
│                                                         │
│  Tools:          Tools:            Tools:               │
│  Loki            Prometheus        Jaeger               │
│  ELK Stack       Grafana           Zipkin               │
│  CloudWatch      DataDog           Tempo                │
│  Fluentd         InfluxDB          AWS X-Ray            │
└─────────────────────────────────────────────────────────┘
```

---

## 🟢 How They Work Together

Scenario: User reports "checkout is slow"

```
Step 1: METRICS (detect the problem)
  Dashboard shows: p99 latency for /checkout spiked from 200ms to 3s
  Started at 14:00
  
Step 2: TRACES (locate the problem)
  Pick a slow /checkout trace:
    → API Gateway:     5ms
    → Auth Service:    10ms
    → Cart Service:    15ms
    → Payment Service: 2800ms  ← HERE!
      → Stripe API:   100ms
      → Database query: 2700ms ← THE BOTTLENECK
    → Email Service:   queued

Step 3: LOGS (understand the problem)
  Payment service logs at 14:00:
  [WARN] Slow query: SELECT * FROM transactions WHERE user_id = ? 
         AND status IN ('pending','completed','refunded')
         Duration: 2700ms
         Missing index on (user_id, status)
  
  Root cause: New feature added status filter, no index.
  Fix: CREATE INDEX idx_transactions_user_status ON transactions(user_id, status);
```

---

## 🟡 Choosing Between Pillars

```
"Something is wrong but I don't know what"
  → Start with METRICS (wide view of system health)

"I know which endpoint is slow but not why"
  → Use TRACES (follow the request through services)

"I know which service is broken but not the root cause"
  → Use LOGS (detailed error messages and stack traces)

"How do I prove it's fixed?"
  → METRICS (latency returned to normal)

Common mistake:
  ❌ Only having logs → grep through millions of lines
  ❌ Only having metrics → know WHAT is slow, not WHY
  ❌ Only having traces → missing the big picture
  
  You need all three. They complement each other.
```

---

## 🟡 Data Volume Reality

```
For a system handling 1000 requests/second:

Metrics:
  → ~100 data points per second (counters, gauges, histograms)
  → ~8 GB/month
  → Cheap to store, query, and retain for months

Logs:
  → ~5,000 log lines per second (multiple per request)
  → ~500 GB/month (uncompressed)
  → Expensive to store, must filter aggressively

Traces:
  → ~1,000 traces per second (one per request)
  → With sampling at 10%: ~100 traces/sec stored
  → ~50 GB/month
  → Sampling is essential at scale

Cost implications:
  Metrics: keep everything, retain for years
  Logs: filter in production, retain for 30-90 days
  Traces: SAMPLE (don't store every trace), retain for 7-30 days
```

---

## 🟡 The Correlation Problem

The real power comes from linking pillars together:

```
Trace ID: abc-123-def-456
  ↓
Metrics: request_duration{trace_id="abc-123"} = 3.2s
  ↓
Logs: [abc-123-def-456] [ERROR] Connection pool exhausted
  ↓
All three tell the same story, from different angles.

How to correlate:
  1. Generate a unique trace ID at the entry point
  2. Pass it through every service (via HTTP headers)
  3. Include it in every log line
  4. Record it as a metric label (careful: high cardinality!)
  
  Trace ID is the GLUE between all three pillars.
```

---

## 🔴 Anti-Patterns

```
❌ "We have Grafana, we're observable"
   → Grafana is a DASHBOARD tool. You need data sources.
   
❌ "We log everything at DEBUG level in production"
   → 10 TB/month of logs, $$$, most is noise.
   
❌ "We'll add observability later"
   → Day 1 decision. Retrofitting is 10x harder.
   
❌ "Just use DataDog/New Relic for everything"
   → Works great until the bill arrives ($100K+/year).
   → Know the open-source alternatives.
   
❌ "We trace 100% of requests"
   → At scale, tracing everything is prohibitively expensive.
   → Sample intelligently (always trace errors, sample successes).
```

---

**Previous:** [README](./README.md)  
**Next:** [02. Structured Logging](./02-structured-logging.md)
