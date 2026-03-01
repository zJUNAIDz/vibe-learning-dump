# Observability Failures

> **The irony: the system designed to DETECT failures can itself fail silently. When your monitoring goes down, you're flying blind — and you won't know until a customer tells you.**

---

## 🟢 Alert Fatigue

### How It Kills Reliability

```
Timeline of alert fatigue:

Month 1: Team sets up monitoring
  → 5 alerts configured, all meaningful
  → Engineers respond to every alert within 5 minutes
  → Issues fixed promptly

Month 3: More alerts added
  → 25 alerts configured
  → Some fire frequently ("CPU > 70% for 1 minute")
  → Engineers check, nothing is actually wrong
  → Response time: 15 minutes (still okayish)

Month 6: Alert overload
  → 80 alerts configured
  → 15-20 fire daily
  → Most are noise
  → Engineers mute Slack channel
  → Response time: "when I get around to it"

Month 9: Real incident
  → Database fills up at 2 AM
  → Alert fires
  → Nobody notices (channel muted, phone silenced)
  → Users discover the outage at 8 AM
  → 6 hours of downtime
  → "But we HAVE monitoring!"
```

### The Fix

```
1. Audit every alert quarterly
   For each alert, ask:
   - When was the last time this fired?
   - What did the responder DO?
   - If they did nothing → DELETE the alert
   - If they did the same thing every time → AUTOMATE it

2. Track alert metrics
   - alerts_fired_total per week
   - alerts_acknowledged_total per week
   - alerts_resolved_without_action_total per week
   
   If (without_action / total) > 50% → too much noise

3. Severity discipline
   - P1 (page): Only if users are actively impacted NOW
   - P2 (Slack): Service degraded, needs fix within hours
   - P3 (ticket): Something to look at this week
   - P4 (dashboard): Informational only, never notify

4. Golden rule: If an engineer is paged and the correct
   response is "do nothing," that alert should not page.
```

---

## 🟢 Missing Metrics

### What You Forgot to Monitor

```
Scenario: Everything looks green, but users complain

Dashboard shows:
  ✅ CPU: 30%
  ✅ Memory: 50%
  ✅ Error rate: 0.1%
  ✅ Request rate: 1000 req/s
  
Users report: "The site is incredibly slow"

What you DIDN'T monitor:
  ❌ p99 latency (average was 50ms, p99 was 15 SECONDS)
  ❌ Database connection pool saturation (all connections used)
  ❌ External API latency (Stripe taking 10s per call)
  ❌ DNS resolution time (5s per lookup due to resolver issues)
  ❌ Garbage collection pauses (500ms stop-the-world)
  ❌ Disk I/O wait (database on slow disk)

Average latency hides everything.
CPU/memory alone tells you almost nothing about user experience.
```

### What You MUST Monitor

```
For every service:

USER-FACING:
  □ Request rate (total throughput)
  □ Error rate (% of 4xx and 5xx)
  □ Latency percentiles (p50, p90, p99 — NOT just average)
  □ Availability (% of time returning success)

DEPENDENCIES:
  □ Database query latency (by query type)
  □ Database connection pool utilization
  □ Cache hit rate
  □ External API latency and error rate
  □ Message queue depth and consumer lag

INFRASTRUCTURE:
  □ CPU, memory, disk usage (with TRENDS)
  □ Network I/O
  □ Pod restart count
  □ Node count and health

BUSINESS:
  □ Signups per minute (sudden drop = problem)
  □ Orders per minute (drop during peak = revenue loss)
  □ Payment success rate
  □ Search result count (returning 0 = search broken)
```

---

## 🟡 Log Explosion

### When Logs Become the Problem

```
Scenario: Log storage costs more than the application itself

Before:
  → App logs at DEBUG level
  → 50GB of logs per day
  → Loki/Elasticsearch storage: 1.5TB per month
  → Cost: $500/month just for log storage+indexing
  → Searching logs takes 30+ seconds

Causes:
  1. DEBUG logging in production
     → Every SQL query logged
     → Every HTTP request logged with full headers+body
     → Every function entry/exit logged
  
  2. No log rotation or retention
     → Keeping logs for 1 year "just in case"
     → Nobody looks at logs older than 2 weeks
  
  3. Logging in hot paths
     → Logging inside loops that run 1M times/second
     → Each log line: 200 bytes × 1M/sec = 200 MB/sec
  
  4. Stack traces for expected errors
     → "User not found" logs full stack trace
     → This is expected behavior, not an error
     → Each stack trace: 2KB vs 100 bytes for a message
```

### The Fix

```
1. Log levels by environment:
   → Development: DEBUG
   → Staging: INFO
   → Production: WARN (with INFO for specific services when debugging)

2. Retention policy:
   → Hot (fast search): 7 days
   → Warm (slower search): 30 days
   → Archive (S3/GCS, no search): 90 days
   → Delete: after 90 days

3. Sampling:
   → Log 100% of errors
   → Log 100% of slow requests (> 1s)
   → Log 10% of successful requests
   → Log 0% of health checks

4. Cost tracking:
   → Track log volume per service
   → Alert if a service exceeds expected volume
   → One chatty service shouldn't cost $200/month in logs
```

```typescript
// Sampling implementation
const shouldLog = (level: string, req: Request): boolean => {
  if (level === 'error' || level === 'fatal') return true;
  if (req.path === '/health' || req.path === '/metrics') return false;
  if (level === 'warn') return true;
  
  // Sample 10% of info-level logs
  return Math.random() < 0.1;
};
```

---

## 🟡 Dashboard Rot

```
Scenario: 50 dashboards, nobody uses any of them

Symptoms:
  → "We have a lot of dashboards" (pride)
  → During incident: "Which dashboard should I look at?" (panic)
  → Dashboards show metrics from services that no longer exist
  → Dashboards with broken queries (underlying metric renamed)
  → Only 3 people know which dashboard matters

The fix:

1. One "golden" dashboard per service
   → Shows the 4 golden signals (latency, traffic, errors, saturation)
   → Linked from runbooks and alerts
   → Tested monthly: can you diagnose a problem with THIS dashboard?

2. Delete unused dashboards
   → Grafana tracks dashboard views
   → Dashboard not viewed in 30 days → delete it
   → Nobody will miss it. If they do, recreate it.

3. Dashboard naming convention
   → [Service] - Overview
   → [Service] - Dependencies
   → [Service] - Debugging
   → NOT: "John's dashboard" or "temp-metrics-2"

4. Dashboard as code
   → Store dashboard JSON in Git
   → Review changes in PRs
   → Auto-provision from Git (Grafana provisioning)
```

---

## 🔴 Monitoring the Monitoring

### When Prometheus Goes Down

```
Scenario:
  1. Prometheus pod runs out of disk (state fills up)
  2. Prometheus crashes
  3. No metrics collected
  4. Alert rules can't evaluate (Prometheus is down!)
  5. AlertManager has nothing to send
  6. Team gets ZERO alerts
  7. Meanwhile: production service fails
  8. No alerts for the production failure either
  9. User reports problem 2 hours later

Your monitoring system has NO monitoring.
```

### The Fix: Meta-Monitoring

```yaml
# 1. Dead man's switch (heartbeat alert)
# This alert fires CONSTANTLY. If it STOPS firing,
# something is wrong with Prometheus/AlertManager.

# Prometheus alert rule:
- alert: Watchdog
  expr: vector(1)
  labels:
    severity: none
  annotations:
    summary: "Heartbeat"

# AlertManager sends this to a dead man's switch service:
# → Healthchecks.io (free tier available)
# → PagerDuty heartbeat
# → OpsGenie heartbeat

# If Prometheus goes down → heartbeat stops →
# Dead man's switch alerts via DIFFERENT channel

# 2. Monitor Prometheus itself from OUTSIDE
# Simple uptime check:
# → Uptime Robot / Better Uptime / Pingdom
# → Checks https://prometheus.internal/-/healthy every 30s
# → Alerts via email/SMS (not through Prometheus!)

# 3. Prometheus disk space alert (before it fills)
- alert: PrometheusStorageFilling
  expr: |
    predict_linear(prometheus_tsdb_storage_size_bytes[6h], 24 * 60 * 60) 
    > prometheus_tsdb_retention_limit_bytes
  for: 1h
  labels:
    severity: warning
  annotations:
    summary: "Prometheus storage will be full in ~24 hours"
```

---

## 🔴 Correlation Failure

```
You have metrics, logs, and traces — but they're disconnected.

Scenario:
  Dashboard: Error rate spiked at 14:23
  Logs: 50,000 error logs around 14:23 (which ones matter?)
  Traces: Hundreds of error traces (which request caused it?)

  Without correlation:
    → Manually search logs by timestamp
    → Guess which log lines are related
    → Can't link a specific error to a specific trace
    → Take 45 minutes to find root cause

  With correlation:
    → Dashboard shows error rate spike
    → Click → filtered logs for that time window
    → Log shows: {"level":"error","trace_id":"abc123","message":"timeout"}
    → Click trace_id → full trace in Jaeger
    → Trace shows: payment-service → stripe-api (timeout after 30s)
    → Root cause found in 3 minutes

How to enable:
  1. Include trace_id in EVERY log line (structured logging)
  2. Use same labels in metrics and logs (app=order-service)
  3. Configure Grafana derived fields (log → trace link)
  4. Include request_id that users can give you for debugging
```

---

## 🔴 Anti-Pattern Summary

```
❌ Monitoring without alerting
   → Dashboard exists but nobody watches it 24/7
   → Fix: alerts for conditions that need human response

❌ Alerting without runbooks
   → "Error rate high" — what should I do?
   → Fix: Every alert MUST link to a runbook

❌ Watching averages instead of percentiles
   → Average latency 50ms (looks great!)
   → p99 latency 10 seconds (1% of users furious)
   → Fix: Always use p50, p90, p99

❌ Not monitoring business metrics
   → All technical metrics green
   → But signups dropped 80% (payment form broken)
   → Fix: Monitor business KPIs, alert on anomalies

❌ Not testing alerts
   → "I think the alerts work"
   → Fix: Run game days — intentionally break things
   → Verify: alert fires, notification arrives, runbook works

❌ Log everything forever
   → 10TB of logs, $2000/month, nobody searches past 7 days
   → Fix: Retention policies, sampling, cost tracking
```

---

**Previous:** [05. Terraform State Corruption](./05-terraform-state-corruption.md)  
**Up:** [README](./README.md)
