# Tools — Prometheus, Grafana, Loki, Jaeger

> **You don't need 20 tools. You need 4 that work together: Prometheus for metrics, Grafana for visualization, Loki for logs, and Jaeger (or Tempo) for traces.**

---

## 🟢 The Observability Stack

```
┌───────────────────────────────────────────────────────────┐
│                    GRAFANA (Visualization)                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │ Metrics   │  │ Logs     │  │ Traces   │  │ Alerts   │ │
│  │ Dashboard │  │ Explorer │  │ Explorer │  │ Panel    │ │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘ │
│       │              │              │              │       │
│  ┌────▼────┐   ┌─────▼───┐   ┌─────▼───┐   ┌─────▼───┐ │
│  │Prometheus│   │  Loki   │   │  Jaeger  │   │ Alert   │ │
│  │ (TSDB)  │   │ (logs)  │   │ (traces) │   │ Manager │ │
│  └────┬────┘   └────┬────┘   └────┬─────┘   └─────────┘ │
└───────┼──────────────┼──────────────┼─────────────────────┘
        │              │              │
   ┌────▼────────────▼──────────────▼─────┐
   │          Your Applications            │
   │  /metrics  │  stdout logs  │  traces  │
   └──────────────────────────────────────┘
```

---

## 🟢 Prometheus

### What It Does

```
Prometheus is a time-series database that:
  → SCRAPES metrics from your apps (pull model)
  → STORES them efficiently (local TSDB)
  → QUERIES with PromQL
  → EVALUATES alert rules
  → SENDS alerts to AlertManager

What it's NOT:
  → Not for logs (use Loki)
  → Not for traces (use Jaeger/Tempo)
  → Not for long-term storage by default (use Thanos/Mimir)
```

### Kubernetes Deployment

```yaml
# Using kube-prometheus-stack Helm chart (the standard)
# Includes: Prometheus, AlertManager, Grafana, node-exporter, kube-state-metrics

# values.yaml
prometheus:
  prometheusSpec:
    retention: 15d
    resources:
      requests:
        memory: 2Gi
        cpu: 500m
      limits:
        memory: 4Gi
    storageSpec:
      volumeClaimTemplate:
        spec:
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: 50Gi
    
    # Service discovery — auto-find pods with annotations
    serviceMonitorSelector: {}
    podMonitorSelector: {}

alertmanager:
  alertmanagerSpec:
    resources:
      requests:
        memory: 256Mi

grafana:
  adminPassword: "change-me-immediately"
  persistence:
    enabled: true
    size: 10Gi
```

```bash
# Install the full stack
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install monitoring prometheus-community/kube-prometheus-stack \
  -n monitoring --create-namespace \
  -f values.yaml
```

### Service Discovery

```yaml
# Tell Prometheus to scrape your app
# Option 1: ServiceMonitor (preferred in Kubernetes)
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: order-service
  labels:
    release: monitoring  # Must match Prometheus selector
spec:
  selector:
    matchLabels:
      app: order-service
  endpoints:
    - port: http
      path: /metrics
      interval: 15s

# Option 2: Pod annotations (simpler but less flexible)
# In your Deployment spec:
template:
  metadata:
    annotations:
      prometheus.io/scrape: "true"
      prometheus.io/port: "3000"
      prometheus.io/path: "/metrics"
```

---

## 🟢 Grafana

### What It Does

```
Grafana is a visualization platform that:
  → Connects to multiple data sources (Prometheus, Loki, Jaeger, etc.)
  → Creates dashboards with panels (graphs, tables, stats, etc.)
  → Supports alerting (can also alert directly, not just via Prometheus)
  → Allows drill-down from metrics → logs → traces

Key concept: Grafana doesn't store data. It QUERIES backends.
```

### Essential Dashboard Panels

```
For every service, create a dashboard with:

Row 1: Overview
  → Request Rate (rate(http_requests_total[5m]))
  → Error Rate (% 5xx)
  → Active requests (gauge)

Row 2: Latency
  → p50 latency (histogram_quantile(0.5, ...))
  → p95 latency
  → p99 latency
  → Latency heatmap

Row 3: Infrastructure
  → CPU usage
  → Memory usage
  → Pod count
  → Pod restarts

Row 4: Dependencies
  → Database latency
  → Database connection pool
  → Cache hit rate
  → External API latency
```

### Dashboard as Code (JSON provisioning)

```yaml
# Grafana dashboard provisioning
# grafana/provisioning/dashboards/dashboard.yaml
apiVersion: 1
providers:
  - name: 'default'
    folder: 'Services'
    type: file
    options:
      path: /var/lib/grafana/dashboards
      foldersFromFilesStructure: true
```

```json
{
  "dashboard": {
    "title": "Order Service",
    "panels": [
      {
        "title": "Request Rate",
        "type": "timeseries",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{service=\"order-service\"}[5m]))",
            "legendFormat": "requests/sec"
          }
        ],
        "gridPos": { "h": 8, "w": 12, "x": 0, "y": 0 }
      },
      {
        "title": "Error Rate",
        "type": "stat",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{service=\"order-service\",status=~\"5..\"}[5m])) / sum(rate(http_requests_total{service=\"order-service\"}[5m])) * 100"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "thresholds": {
              "steps": [
                { "color": "green", "value": 0 },
                { "color": "yellow", "value": 1 },
                { "color": "red", "value": 5 }
              ]
            },
            "unit": "percent"
          }
        },
        "gridPos": { "h": 8, "w": 12, "x": 12, "y": 0 }
      }
    ]
  }
}
```

---

## 🟡 Loki (Log Aggregation)

### What It Does

```
Loki is a log aggregation system designed for Grafana:
  → Indexes LABELS only (not full text) → much cheaper than Elasticsearch
  → Uses same label model as Prometheus
  → Queries with LogQL (similar to PromQL)
  → Integrates seamlessly in Grafana (metrics → logs in one click)

Loki vs Elasticsearch:
  Elasticsearch: Indexes every word → expensive, powerful full-text search
  Loki: Indexes only labels → cheap, grep-like search on log content
  
  For most teams: Loki is enough and 10x cheaper.
```

### Deployment with Promtail

```yaml
# Promtail runs as a DaemonSet — collects logs from pods on each node
# It reads container logs from /var/log/pods/ and sends to Loki

# Loki (simple scalable mode)
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: loki
spec:
  replicas: 1
  template:
    spec:
      containers:
        - name: loki
          image: grafana/loki:latest
          args:
            - -config.file=/etc/loki/config.yaml
          ports:
            - containerPort: 3100
          volumeMounts:
            - name: config
              mountPath: /etc/loki
            - name: storage
              mountPath: /loki
```

```yaml
# Loki config
# loki-config.yaml
auth_enabled: false

server:
  http_listen_port: 3100

common:
  path_prefix: /loki
  storage:
    filesystem:
      chunks_directory: /loki/chunks
      rules_directory: /loki/rules
  replication_factor: 1
  ring:
    kvstore:
      store: inmemory

schema_config:
  configs:
    - from: 2024-01-01
      store: tsdb
      object_store: filesystem
      schema: v13
      index:
        prefix: index_
        period: 24h

limits_config:
  retention_period: 720h  # 30 days
```

### LogQL (Loki Query Language)

```logql
# Find all error logs from order-service
{app="order-service"} |= "error"

# JSON parsed logs — filter by level
{app="order-service"} | json | level="error"

# Search for a specific trace ID
{app=~".*"} |= "trace_id=abc123"

# Count errors per minute
count_over_time({app="order-service"} |= "error" [1m])

# Error rate per service
sum by (app) (count_over_time({app=~".+"} | json | level="error" [5m]))

# Logs with latency > 1 second
{app="order-service"} | json | duration > 1s

# Top 10 error messages
{app="order-service"} | json | level="error" 
  | line_format "{{.message}}" 
  | topk(10, count_over_time({app="order-service"} |= "error" [1h]))
```

---

## 🟡 Jaeger (Distributed Tracing)

### What It Does

```
Jaeger is a distributed tracing backend:
  → Receives trace data from apps (via OpenTelemetry)
  → Stores traces
  → Provides UI to search and visualize traces
  → Helps find latency bottlenecks and error paths

Alternative: Grafana Tempo (same purpose, integrates better with Grafana)
```

### Kubernetes Deployment

```yaml
# Simple all-in-one deployment (dev/staging)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: jaeger
spec:
  replicas: 1
  selector:
    matchLabels:
      app: jaeger
  template:
    metadata:
      labels:
        app: jaeger
    spec:
      containers:
        - name: jaeger
          image: jaegertracing/all-in-one:latest
          ports:
            - containerPort: 16686  # UI
            - containerPort: 4317   # OTLP gRPC
            - containerPort: 4318   # OTLP HTTP
          env:
            - name: COLLECTOR_OTLP_ENABLED
              value: "true"
            - name: SPAN_STORAGE_TYPE
              value: "badger"  # Embedded storage
            - name: BADGER_DIRECTORY_VALUE
              value: "/badger/data"
            - name: BADGER_DIRECTORY_KEY
              value: "/badger/key"
          volumeMounts:
            - name: badger
              mountPath: /badger
      volumes:
        - name: badger
          persistentVolumeClaim:
            claimName: jaeger-storage
---
apiVersion: v1
kind: Service
metadata:
  name: jaeger
spec:
  ports:
    - name: ui
      port: 16686
    - name: otlp-grpc
      port: 4317
    - name: otlp-http
      port: 4318
  selector:
    app: jaeger
```

---

## 🔴 Full Stack Docker Compose (Local Dev)

```yaml
# docker-compose.observability.yaml
# Complete local observability stack
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./prometheus/alert-rules.yml:/etc/prometheus/alert-rules.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.retention.time=15d'
      - '--web.enable-lifecycle'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning
      - ./grafana/dashboards:/var/lib/grafana/dashboards
    depends_on:
      - prometheus
      - loki
      - jaeger

  loki:
    image: grafana/loki:latest
    ports:
      - "3100:3100"
    volumes:
      - ./loki/config.yaml:/etc/loki/config.yaml
      - loki_data:/loki
    command: -config.file=/etc/loki/config.yaml

  promtail:
    image: grafana/promtail:latest
    volumes:
      - ./promtail/config.yaml:/etc/promtail/config.yaml
      - /var/log:/var/log
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
    command: -config.file=/etc/promtail/config.yaml
    depends_on:
      - loki

  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"  # UI
      - "4317:4317"    # OTLP gRPC
      - "4318:4318"    # OTLP HTTP
    environment:
      - COLLECTOR_OTLP_ENABLED=true

  alertmanager:
    image: prom/alertmanager:latest
    ports:
      - "9093:9093"
    volumes:
      - ./alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml

  # OTel Collector — receives traces/metrics from apps
  otel-collector:
    image: otel/opentelemetry-collector-contrib:latest
    ports:
      - "4317"   # OTLP gRPC (for apps)
      - "4318"   # OTLP HTTP (for apps)
    volumes:
      - ./otel-collector/config.yaml:/etc/otelcol-contrib/config.yaml
    depends_on:
      - jaeger
      - prometheus

volumes:
  prometheus_data:
  grafana_data:
  loki_data:
```

### OTel Collector Config

```yaml
# otel-collector/config.yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 5s
    send_batch_size: 1000

exporters:
  # Send traces to Jaeger
  otlp/jaeger:
    endpoint: jaeger:4317
    tls:
      insecure: true

  # Send metrics to Prometheus
  prometheusremotewrite:
    endpoint: http://prometheus:9090/api/v1/write

  # Debug output (dev only)
  debug:
    verbosity: detailed

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp/jaeger]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [prometheusremotewrite]
```

### Grafana Data Sources (Provisioning)

```yaml
# grafana/provisioning/datasources/datasources.yaml
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true

  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    jsonData:
      derivedFields:
        # Click trace_id in logs → opens Jaeger trace
        - name: TraceID
          matcherRegex: '"trace_id":"(\w+)"'
          url: '$${__value.raw}'
          datasourceUid: jaeger-uid
          urlDisplayLabel: View Trace

  - name: Jaeger
    type: jaeger
    uid: jaeger-uid
    access: proxy
    url: http://jaeger:16686
```

---

## 🔴 Connecting the Dots (Metrics → Logs → Traces)

```
The real power is connecting all three:

1. Dashboard shows error rate spike (METRICS)
   → Click on the spike
   
2. Split view shows error logs at that time (LOGS)
   → See: {"level":"error","message":"payment failed","trace_id":"abc123"}
   → Click trace_id

3. Trace view shows the full request journey (TRACES)
   → Root span → auth → order → payment (ERROR: timeout)
   → See EXACTLY which call failed and why

How to enable this:
  1. Add trace_id to every log line (structured logging)
  2. Configure Grafana's derived fields (LogQL → Jaeger link)
  3. Use consistent labels across metrics and logs 
     (app="order-service" in both Prometheus and Loki)
```

---

## 🔴 Anti-Patterns

```
❌ Using different labels in Prometheus vs Loki
   → Prometheus: service="order-svc"
   → Loki: app="order-service"
   → Can't correlate! Pick ONE naming convention.

❌ Not including trace_id in logs
   → Without trace_id, logs and traces are disconnected
   → You can't jump from a log line to its trace

❌ Running Prometheus without persistent storage
   → Pod restart = all metrics lost
   → Always use PVCs for Prometheus data

❌ No retention policy
   → Prometheus fills disk → crashes
   → Set --storage.tsdb.retention.time=15d
   → Set Loki retention_period

❌ Grafana dashboards only for "looking at"
   → Dashboards should link to runbooks
   → Dashboards should be used during incident response
   → If nobody uses the dashboard, delete it
```

---

**Previous:** [05. Alerting](./05-alerting.md)  
**Up:** [README](./README.md)
