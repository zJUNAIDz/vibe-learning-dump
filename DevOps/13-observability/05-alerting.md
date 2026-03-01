# Alerting

> **An alert is a notification that something needs human attention. Good alerting wakes you up for real problems. Bad alerting wakes you up for noise — and eventually you ignore everything.**

---

## 🟢 What Makes a Good Alert

```
A good alert has ALL of these properties:

1. ACTIONABLE
   → Someone can DO something about it right now
   → If nobody can act, it shouldn't be an alert
   
2. URGENT  
   → It needs attention NOW (or soon)
   → If it can wait until Monday, it's not an alert
   
3. REAL
   → It indicates actual user impact or imminent failure
   → If it fires and nothing is actually wrong, it's noise
   
4. UNIQUE
   → One problem = one alert
   → If a DB goes down, you get 1 alert for "DB down"
   → NOT 50 alerts for every service that can't reach the DB

Bad alert: "CPU > 80% for 1 minute"
  → Not actionable (maybe it's just a deploy)
  → Not necessarily urgent
  → Not necessarily a problem

Good alert: "Error rate > 5% for 5 minutes"
  → Actionable (check what's failing)
  → Urgent (users are affected)
  → Real (errors = actual failures)
```

---

## 🟢 Severity Levels

```
P1 — CRITICAL (page immediately, wake people up)
  → Service is DOWN
  → Data loss occurring
  → Security breach detected
  Response: Acknowledge in 5 min, resolve or mitigate in 30 min
  Examples:
    → 0 successful responses for 2+ minutes
    → Database unreachable
    → All pods in CrashLoopBackOff

P2 — HIGH (page during business hours, Slack urgently)
  → Service degraded but not down
  → Error rate significantly elevated
  → Approaching resource limits
  Response: Acknowledge in 15 min, resolve in 4 hours
  Examples:
    → Error rate > 5% for 5 minutes
    → p99 latency > 5 seconds for 10 minutes
    → Disk > 90%

P3 — WARNING (Slack notification, ticket)
  → Something is trending badly
  → Non-critical component failed
  → Approaching threshold
  Response: Address within 1-2 business days
  Examples:
    → Disk growing faster than expected
    → SSL certificate expires in 14 days
    → Non-critical job failed

P4 — INFO (dashboard, weekly review)
  → Interesting but not urgent
  → Trend to watch
  Response: Review in next planning session
  Examples:
    → Traffic pattern changed
    → New error type appeared (low volume)
```

---

## 🟢 Alert Rules in Prometheus

### Prometheus Alerting Rules

```yaml
# alert-rules.yaml
groups:
  - name: service-health
    rules:
      # High error rate
      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m]))
          / sum(rate(http_requests_total[5m])) > 0.05
        for: 5m  # Must be true for 5 minutes before firing
        labels:
          severity: critical
          team: backend
        annotations:
          summary: "High error rate: {{ $value | humanizePercentage }}"
          description: "Error rate is above 5% for the last 5 minutes"
          runbook: "https://wiki.internal/runbooks/high-error-rate"
          dashboard: "https://grafana.internal/d/service-health"

      # High latency
      - alert: HighLatency
        expr: |
          histogram_quantile(0.99, 
            sum by (le) (rate(http_request_duration_seconds_bucket[5m]))
          ) > 2
        for: 10m
        labels:
          severity: warning
          team: backend
        annotations:
          summary: "p99 latency above 2 seconds"
          description: "99th percentile latency: {{ $value }}s"
          runbook: "https://wiki.internal/runbooks/high-latency"

      # Service down (zero requests)
      - alert: ServiceDown
        expr: |
          absent(up{job="order-service"} == 1)
        for: 2m
        labels:
          severity: critical
          team: platform
        annotations:
          summary: "Service order-service is down"
          runbook: "https://wiki.internal/runbooks/service-down"

  - name: infrastructure
    rules:
      # High CPU
      - alert: HighCPU
        expr: |
          100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 85
        for: 15m  # Sustained high CPU, not just a spike
        labels:
          severity: warning
          team: platform
        annotations:
          summary: "CPU above 85% on {{ $labels.instance }}"

      # Disk almost full
      - alert: DiskAlmostFull
        expr: |
          (node_filesystem_avail_bytes / node_filesystem_size_bytes) < 0.10
        for: 10m
        labels:
          severity: critical
          team: platform
        annotations:
          summary: "Disk < 10% free on {{ $labels.instance }}"
          runbook: "https://wiki.internal/runbooks/disk-full"

      # Pod CrashLooping
      - alert: PodCrashLooping
        expr: |
          increase(kube_pod_container_status_restarts_total[1h]) > 5
        for: 5m
        labels:
          severity: critical
          team: platform
        annotations:
          summary: "Pod {{ $labels.pod }} is crash-looping"
          description: "{{ $labels.pod }} restarted {{ $value }} times in 1 hour"
```

### AlertManager Configuration

```yaml
# alertmanager.yaml
global:
  resolve_timeout: 5m
  slack_api_url: 'https://hooks.slack.com/services/XXX'

route:
  receiver: 'default-slack'
  group_by: ['alertname', 'severity']
  group_wait: 30s       # Wait 30s to batch related alerts
  group_interval: 5m    # Wait 5m between sending groups
  repeat_interval: 4h   # Re-send unresolved alerts every 4h
  
  routes:
    # Critical → PagerDuty (wakes people up)
    - match:
        severity: critical
      receiver: 'pagerduty-critical'
      repeat_interval: 1h

    # Warning → Slack channel
    - match:
        severity: warning
      receiver: 'team-slack'
      repeat_interval: 4h

receivers:
  - name: 'default-slack'
    slack_configs:
      - channel: '#alerts-default'
        title: '{{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'

  - name: 'pagerduty-critical'
    pagerduty_configs:
      - service_key: '<pagerduty-service-key>'
        severity: critical

  - name: 'team-slack'
    slack_configs:
      - channel: '#alerts-backend'
        title: '[{{ .GroupLabels.severity | toUpper }}] {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'

# Silence noisy alerts during maintenance
# Use: amtool silence add alertname="HighCPU" --duration=2h --comment="Deploy in progress"
```

---

## 🟡 On-Call Best Practices

```
1. ROTATION
   → Weekly or bi-weekly rotation
   → At least 2 people on-call (primary + secondary)
   → Never same person two weeks in a row
   → Compensate on-call time

2. ESCALATION
   → Primary doesn't respond in 5 min → page secondary
   → Secondary doesn't respond in 5 min → page team lead
   → After 15 min → page engineering manager

3. HANDOFF
   → End of shift: review active incidents
   → Document anything in progress
   → Share context with next person

4. POST-INCIDENT
   → Every P1/P2 gets a blameless post-mortem
   → Identify: What happened? Why? How to prevent?
   → Action items with owners and deadlines
   → Share learnings with team
```

---

## 🟡 Runbooks

```markdown
Every alert MUST have a linked runbook.

## Runbook: HighErrorRate

### What This Means
Error rate has exceeded 5% for 5+ minutes. Users are experiencing failures.

### Impact
Users cannot complete [checkout / login / search]. Revenue impact.

### Diagnosis Steps
1. Open Grafana dashboard: [link]
2. Check which endpoint has errors:
   `sum by (path) (rate(http_requests_total{status=~"5.."}[5m]))`
3. Check recent deployments: `kubectl rollout history deployment/api`
4. Check upstream dependencies:
   - Database: [dashboard link]
   - Payment service: [dashboard link]
   - Redis: [dashboard link]

### Common Causes & Fixes

| Cause | How to Verify | Fix |
|-------|--------------|-----|
| Bad deploy | Errors started with deploy | `kubectl rollout undo deployment/api` |
| DB overloaded | DB CPU > 90% | Scale DB read replicas |
| Upstream down | Payment service errors | Check payment service runbook |
| Memory leak | OOM kills in logs | Restart pods: `kubectl rollout restart` |

### Escalation
If not resolved in 30 minutes, escalate to team lead.
```

---

## 🔴 Alert Fatigue

```
Alert fatigue kills on-call engineers and kills reliability.

Symptoms:
  → Engineers mute alerts
  → "Oh, that alert fires all the time, ignore it"
  → Nobody responds because they assume it's noise
  → A REAL incident gets ignored because of alert fatigue

How to prevent:

1. EVERY alert must be actionable
   → If the response is "do nothing," delete the alert
   
2. Track alert frequency
   → If an alert fires > 5 times/week without action → fix or remove
   
3. Tune thresholds regularly
   → Review alerts monthly
   → Adjust for actual baseline, not theoretical ideal

4. Use "for" duration
   → Don't alert on 1-second spikes
   → Require sustained issue (5-15 minutes)
   
5. Group related alerts
   → DB goes down → 1 alert
   → NOT: 50 "can't connect to DB" alerts from every service

6. Different channels for different urgencies
   → Critical → PagerDuty (wake up)
   → Warning → Slack channel
   → Info → Dashboard only (never notify)
```

---

## 🔴 SLOs and Error Budgets

```
SLI (Service Level Indicator)
  → A metric that measures service quality
  → Example: % of requests that complete in < 500ms

SLO (Service Level Objective)
  → A target for the SLI
  → Example: 99.9% of requests complete in < 500ms

Error Budget
  → How much failure you're allowed
  → 99.9% SLO = 0.1% error budget
  → In 30 days: 0.1% of 43,200 minutes = 43.2 minutes of downtime allowed

If error budget is exhausted:
  → Freeze feature releases
  → Focus on reliability
  → Only ship reliability improvements
```

### SLO-Based Alerting

```yaml
# Alert when burning error budget too fast
# Using multiwindow multi-burn-rate alerts

# Fast burn (2% budget in 1 hour) → Page immediately
- alert: SLOHighBurnRate
  expr: |
    (
      sum(rate(http_requests_total{status=~"5.."}[1h]))
      / sum(rate(http_requests_total[1h]))
    ) > (14.4 * 0.001)
    and
    (
      sum(rate(http_requests_total{status=~"5.."}[5m]))
      / sum(rate(http_requests_total[5m]))
    ) > (14.4 * 0.001)
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "Burning error budget 14.4x faster than allowed"

# Slow burn (10% budget in 3 days) → Ticket
- alert: SLOSlowBurnRate
  expr: |
    (
      sum(rate(http_requests_total{status=~"5.."}[3d]))
      / sum(rate(http_requests_total[3d]))
    ) > (1 * 0.001)
  for: 1h
  labels:
    severity: warning
  annotations:
    summary: "Slowly burning error budget"
```

---

**Previous:** [04. Distributed Tracing](./04-distributed-tracing.md)  
**Next:** [06. Tools](./06-tools.md)
