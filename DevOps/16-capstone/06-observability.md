# Capstone Phase 06: Observability

> **Your application is deployed. How do you know it's working? Observability turns "I think it's fine" into "I can prove it's fine." This phase wires up metrics, logs, traces, dashboards, and alerts for the task-service.**

---

## 🟢 What You Already Have

From Phase 01 (Application), your task-service already includes:

- **Structured logging** — pino with JSON output, request IDs, redacted secrets
- **Prometheus metrics** — `http_requests_total`, `http_request_duration_seconds`, `http_requests_in_progress`
- **Health endpoints** — `/health/live`, `/health/ready`
- **Metrics endpoint** — `/metrics`

This phase connects those to the observability stack.

---

## 🟢 The Observability Stack

```
┌──────────────────────────────────────────────────────┐
│                     Grafana                           │
│         Dashboards  │  Alerts  │  Explore             │
│              ▲           ▲          ▲                  │
│              │           │          │                  │
│    ┌─────────┴──┐  ┌─────┴────┐  ┌─┴──────────┐      │
│    │ Prometheus  │  │ Loki     │  │ Jaeger     │      │
│    │ (metrics)   │  │ (logs)   │  │ (traces)   │      │
│    └──────▲──────┘  └────▲─────┘  └──────▲─────┘      │
│           │              │               │             │
│     ┌─────┴─────┐  ┌────┴─────┐         │             │
│     │ Service   │  │ Promtail │         │             │
│     │ Monitor   │  │          │         │             │
│     └─────▲─────┘  └────▲─────┘         │             │
│           │              │               │             │
│    ┌──────┴──────────────┴───────────────┴──────┐     │
│    │              task-service                    │     │
│    │     /metrics    stdout logs    OTel traces   │     │
│    └─────────────────────────────────────────────┘     │
└──────────────────────────────────────────────────────┘
```

---

## 🟢 ServiceMonitor for Prometheus

```yaml
# k8s/observability/servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: task-service
  namespace: task-service
  labels:
    release: monitoring  # Must match kube-prometheus-stack selector
spec:
  selector:
    matchLabels:
      app: task-service
  endpoints:
    - port: http
      path: /metrics
      interval: 15s
      # Relabel to add namespace and pod labels
      relabelings:
        - sourceLabels: [__meta_kubernetes_namespace]
          targetLabel: namespace
        - sourceLabels: [__meta_kubernetes_pod_name]
          targetLabel: pod
```

Verify Prometheus is scraping:

```bash
# Port forward to Prometheus
kubectl port-forward svc/monitoring-kube-prometheus-prometheus 9090 -n monitoring

# Check targets in browser: http://localhost:9090/targets
# Look for task-service endpoints — they should show UP

# Query directly
curl -s 'http://localhost:9090/api/v1/query?query=http_requests_total' | jq .
```

---

## 🟡 Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Task Service",
    "uid": "task-service",
    "tags": ["task-service", "production"],
    "timezone": "browser",
    "refresh": "30s",
    "panels": [
      {
        "title": "Request Rate (req/s)",
        "type": "timeseries",
        "gridPos": { "h": 8, "w": 12, "x": 0, "y": 0 },
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{job=\"task-service\"}[5m])) by (method, route)",
            "legendFormat": "{{ method }} {{ route }}"
          }
        ]
      },
      {
        "title": "Error Rate (%)",
        "type": "stat",
        "gridPos": { "h": 8, "w": 6, "x": 12, "y": 0 },
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{job=\"task-service\", status_code=~\"5..\"}[5m])) / sum(rate(http_requests_total{job=\"task-service\"}[5m])) * 100"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "percent",
            "thresholds": {
              "steps": [
                { "value": 0, "color": "green" },
                { "value": 1, "color": "yellow" },
                { "value": 5, "color": "red" }
              ]
            }
          }
        }
      },
      {
        "title": "P99 Latency",
        "type": "stat",
        "gridPos": { "h": 8, "w": 6, "x": 18, "y": 0 },
        "targets": [
          {
            "expr": "histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{job=\"task-service\"}[5m])) by (le))"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "s",
            "thresholds": {
              "steps": [
                { "value": 0, "color": "green" },
                { "value": 0.5, "color": "yellow" },
                { "value": 1, "color": "red" }
              ]
            }
          }
        }
      },
      {
        "title": "In-Flight Requests",
        "type": "timeseries",
        "gridPos": { "h": 8, "w": 12, "x": 0, "y": 8 },
        "targets": [
          {
            "expr": "sum(http_requests_in_progress{job=\"task-service\"})",
            "legendFormat": "In Progress"
          }
        ]
      },
      {
        "title": "Latency Distribution",
        "type": "heatmap",
        "gridPos": { "h": 8, "w": 12, "x": 12, "y": 8 },
        "targets": [
          {
            "expr": "sum(increase(http_request_duration_seconds_bucket{job=\"task-service\"}[5m])) by (le)",
            "format": "heatmap"
          }
        ]
      },
      {
        "title": "Pod Memory Usage",
        "type": "timeseries",
        "gridPos": { "h": 8, "w": 12, "x": 0, "y": 16 },
        "targets": [
          {
            "expr": "container_memory_working_set_bytes{namespace=\"task-service\", container=\"task-service\"}",
            "legendFormat": "{{ pod }}"
          }
        ],
        "fieldConfig": { "defaults": { "unit": "bytes" } }
      },
      {
        "title": "Pod CPU Usage",
        "type": "timeseries",
        "gridPos": { "h": 8, "w": 12, "x": 12, "y": 16 },
        "targets": [
          {
            "expr": "rate(container_cpu_usage_seconds_total{namespace=\"task-service\", container=\"task-service\"}[5m])",
            "legendFormat": "{{ pod }}"
          }
        ],
        "fieldConfig": { "defaults": { "unit": "short" } }
      },
      {
        "title": "Pod Restarts",
        "type": "stat",
        "gridPos": { "h": 4, "w": 6, "x": 0, "y": 24 },
        "targets": [
          {
            "expr": "sum(kube_pod_container_status_restarts_total{namespace=\"task-service\", container=\"task-service\"})"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "thresholds": {
              "steps": [
                { "value": 0, "color": "green" },
                { "value": 1, "color": "red" }
              ]
            }
          }
        }
      },
      {
        "title": "Logs",
        "type": "logs",
        "gridPos": { "h": 8, "w": 24, "x": 0, "y": 28 },
        "targets": [
          {
            "expr": "{namespace=\"task-service\", container=\"task-service\"} |= ``",
            "datasource": "Loki"
          }
        ]
      }
    ]
  }
}
```

Deploy as a ConfigMap for auto-provisioning:

```yaml
# k8s/observability/grafana-dashboard.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: task-service-dashboard
  namespace: monitoring
  labels:
    grafana_dashboard: "1"  # Auto-discovered by Grafana sidecar
data:
  task-service.json: |
    { ... the JSON above ... }
```

---

## 🟡 Logs with Loki

Promtail (deployed as part of the monitoring stack) automatically collects stdout/stderr from all pods. Your structured logs from Phase 01 are already flowing.

### Useful LogQL Queries

```
# All task-service logs
{namespace="task-service", container="task-service"}

# Error logs only
{namespace="task-service"} | json | level = "error"

# Slow requests (> 500ms)
{namespace="task-service"} | json | responseTime > 500

# Specific request ID
{namespace="task-service"} | json | requestId = "abc-123"

# Failed requests
{namespace="task-service"} | json | statusCode >= 500

# Rate of errors per minute
sum(rate({namespace="task-service"} | json | level = "error" [1m]))
```

---

## 🟡 Alerting Rules

```yaml
# k8s/observability/prometheusrule.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: task-service-alerts
  namespace: task-service
  labels:
    release: monitoring
spec:
  groups:
    - name: task-service
      rules:
        # High error rate
        - alert: TaskServiceHighErrorRate
          expr: |
            sum(rate(http_requests_total{job="task-service", status_code=~"5.."}[5m]))
            /
            sum(rate(http_requests_total{job="task-service"}[5m]))
            > 0.05
          for: 2m
          labels:
            severity: critical
          annotations:
            summary: "Task service error rate > 5%"
            description: "{{ $value | humanizePercentage }} of requests are failing"
            runbook: "https://runbooks.example.com/task-service/high-error-rate"

        # High latency
        - alert: TaskServiceHighLatency
          expr: |
            histogram_quantile(0.99,
              sum(rate(http_request_duration_seconds_bucket{job="task-service"}[5m])) by (le)
            ) > 1
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "Task service P99 latency > 1s"
            description: "P99 latency is {{ $value }}s"

        # Pod restarts
        - alert: TaskServicePodRestarting
          expr: |
            increase(kube_pod_container_status_restarts_total{
              namespace="task-service", container="task-service"
            }[1h]) > 3
          for: 0m
          labels:
            severity: warning
          annotations:
            summary: "Task service pod restarting frequently"
            description: "{{ $labels.pod }} has restarted {{ $value }} times in the last hour"

        # No traffic (possible outage)
        - alert: TaskServiceNoTraffic
          expr: |
            sum(rate(http_requests_total{job="task-service"}[5m])) == 0
          for: 5m
          labels:
            severity: critical
          annotations:
            summary: "Task service receiving zero traffic"
            description: "No requests in the last 5 minutes — possible outage"

        # High memory usage
        - alert: TaskServiceHighMemory
          expr: |
            container_memory_working_set_bytes{namespace="task-service", container="task-service"}
            /
            container_spec_memory_limit_bytes{namespace="task-service", container="task-service"}
            > 0.9
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "Task service using > 90% of memory limit"
            description: "{{ $labels.pod }} is at {{ $value | humanizePercentage }} memory"
```

---

## 🔴 AlertManager Routing

```yaml
# In your kube-prometheus-stack Helm values:
alertmanager:
  config:
    route:
      receiver: "default"
      group_by: ["alertname", "namespace"]
      group_wait: 30s
      group_interval: 5m
      repeat_interval: 4h
      routes:
        - match:
            severity: critical
          receiver: "pagerduty"
          continue: true
        - match:
            severity: critical
          receiver: "slack-critical"
        - match:
            severity: warning
          receiver: "slack-warnings"

    receivers:
      - name: "default"
        slack_configs:
          - channel: "#alerts-default"
            api_url: "<SLACK_WEBHOOK>"

      - name: "pagerduty"
        pagerduty_configs:
          - service_key: "<PD_SERVICE_KEY>"

      - name: "slack-critical"
        slack_configs:
          - channel: "#alerts-critical"
            api_url: "<SLACK_WEBHOOK>"
            title: "🔴 CRITICAL: {{ .GroupLabels.alertname }}"
            text: "{{ range .Alerts }}{{ .Annotations.description }}\n{{ end }}"

      - name: "slack-warnings"
        slack_configs:
          - channel: "#alerts-warnings"
            api_url: "<SLACK_WEBHOOK>"
```

---

## 🔴 Debugging with Observability

When something goes wrong, use the three pillars together:

```
1. ALERT fires → TaskServiceHighErrorRate (> 5% errors)

2. DASHBOARD → Grafana shows spike in 5xx at 14:32
   - Error rate panel: jumped from 0.1% to 8%
   - Latency panel: P99 spiked to 3s
   - Memory panel: pod-xyz at 95% memory limit

3. LOGS → Loki query:
   {namespace="task-service"} | json | level = "error"
   
   Shows: "Cannot read property 'id' of undefined"
   with requestId: "req-abc-123"
   at route: POST /api/tasks

4. TRACES → (if OTel configured) Search by request ID
   Shows: POST /api/tasks → DB query → null result → crash

5. ROOT CAUSE: Missing null check when database returns empty result

6. FIX → Deploy patch → Watch error rate drop on dashboard
```

---

## 🔴 Complete Checklist

```
□ ServiceMonitor deployed — Prometheus scraping /metrics
□ Grafana dashboard with request rate, error rate, P99, memory, CPU
□ Dashboard deployed as ConfigMap (dashboard-as-code)
□ Logs flowing to Loki (structured JSON from pino)
□ LogQL queries working for errors, slow requests, request IDs
□ PrometheusRule with alerts: error rate, latency, restarts, no traffic, memory
□ AlertManager routing: critical → PagerDuty + Slack, warning → Slack
□ Can correlate: alert → dashboard → logs → root cause
□ Runbook linked in alert annotations
```

---

## 🔴 What You've Built (Full Capstone)

```
Phase 01: Application      → TypeScript REST API with health, metrics, logging
Phase 02: Containerization  → Multi-stage Docker, scanning, optimization
Phase 03: K8s Manifests     → Deployment, Service, Ingress, HPA, PDB, security
Phase 04: CI/CD Pipeline    → Test → Build → Scan → Stage → Smoke → Prod
Phase 05: IaC               → Terraform modules (VPC, EKS, RDS) per environment
Phase 06: Observability     → Metrics, logs, dashboards, alerts, correlation
```

From code push to production with full observability. That's DevOps.

---

**Previous:** [05. Infrastructure as Code](./05-infrastructure-as-code.md)  
**Back to README:** [Capstone Overview](./README.md)
