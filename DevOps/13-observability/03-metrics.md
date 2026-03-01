# Metrics

> **Metrics are numbers over time. CPU usage, request count, error rate, latency percentiles. They answer "how is the system doing RIGHT NOW?" and "what changed?"**

---

## 🟢 What Are Metrics?

```
A metric is a numeric measurement at a point in time:

  http_requests_total = 15,234        (at 14:00:00)
  http_requests_total = 15,289        (at 14:00:15)
  → 55 requests in 15 seconds = 3.67 requests/sec

  cpu_usage_percent = 72.3            (at 14:00:00)
  memory_used_bytes = 3,421,172,736   (at 14:00:00)
  http_request_duration_seconds = 0.045 (for this request)

Metrics are:
  → Small (a number + labels + timestamp)
  → Cheap to store (compared to logs)
  → Fast to query (time-series databases)
  → Easy to aggregate (sum, avg, percentiles)
  → Great for dashboards and alerting
```

---

## 🟢 Metric Types

### Counter (Only Goes Up)

```
A counter is a cumulative value that only increases (or resets to 0).

http_requests_total{method="GET", path="/api/users", status="200"} = 15234
http_requests_total{method="POST", path="/api/orders", status="201"} = 892
http_requests_total{method="GET", path="/api/users", status="500"} = 3

Use for:
  → Total requests served
  → Total errors
  → Total bytes transferred
  → Total tasks completed

Query the RATE to get useful info:
  rate(http_requests_total[5m]) → requests per second (averaged over 5 min)
```

### Gauge (Goes Up and Down)

```
A gauge is a current value that can increase or decrease.

temperature_celsius = 42.5
cpu_usage_percent = 72.3
active_connections = 1234
queue_depth = 89
memory_used_bytes = 3421172736

Use for:
  → CPU/memory/disk usage
  → Active connections
  → Queue depth
  → Temperature
  → Number of running pods
```

### Histogram (Distribution of Values)

```
A histogram counts observations in configurable "buckets."

http_request_duration_seconds_bucket{le="0.01"} = 2341    → ≤10ms
http_request_duration_seconds_bucket{le="0.05"} = 8924    → ≤50ms
http_request_duration_seconds_bucket{le="0.1"}  = 11203   → ≤100ms
http_request_duration_seconds_bucket{le="0.5"}  = 15002   → ≤500ms
http_request_duration_seconds_bucket{le="1.0"}  = 15189   → ≤1s
http_request_duration_seconds_bucket{le="+Inf"} = 15234   → all
http_request_duration_seconds_sum = 1523.4                 → total seconds
http_request_duration_seconds_count = 15234                → total count

Use for:
  → Request latency
  → Response sizes
  → Queue wait times
  → Any value where distribution matters

Why histograms > averages:
  Average latency: 50ms (looks fine!)
  p99 latency: 3000ms (1% of users wait 3 SECONDS)
  
  Averages HIDE tail latency. Use percentiles.
```

### Summary (Pre-calculated Percentiles)

```
Similar to histogram but calculates percentiles on the client.

http_request_duration_seconds{quantile="0.5"}  = 0.042    → p50 (median)
http_request_duration_seconds{quantile="0.9"}  = 0.089    → p90
http_request_duration_seconds{quantile="0.99"} = 0.341    → p99

Histogram vs Summary:
  Histogram → Server-side calculation, can aggregate across instances
  Summary   → Client-side calculation, can't aggregate (pre-computed)
  
  → Use HISTOGRAM in most cases (more flexible)
```

---

## 🟢 The Four Golden Signals

Google's SRE book defines four key metrics for any service:

```
1. LATENCY
   → How long do requests take?
   → Track: p50, p90, p99
   → Alert: p99 > 500ms
   
2. TRAFFIC  
   → How many requests per second?
   → Track: requests/sec by endpoint
   → Alert: sudden drop (service might be down)
   
3. ERRORS
   → What percentage of requests fail?
   → Track: 5xx rate, error rate by type
   → Alert: error rate > 1%
   
4. SATURATION
   → How full is the system?
   → Track: CPU, memory, disk, queue depth
   → Alert: CPU > 80%, disk > 90%

If you only track 4 things, track these.
```

---

## 🟢 Prometheus (Industry Standard)

### How Prometheus Works

```
┌─────────────────────────────────────────────┐
│                Prometheus                    │
│                                             │
│  1. DISCOVER targets (Kubernetes, DNS, etc) │
│  2. SCRAPE /metrics endpoint every 15s      │
│  3. STORE in time-series database           │
│  4. EVALUATE alert rules                    │
│  5. SEND alerts to AlertManager             │
│                                             │
│  ┌─────┐  ┌─────┐  ┌─────┐  ┌─────────┐   │
│  │Scrape│→ │Store│→ │Query│→ │Alerting │   │
│  └─────┘  └─────┘  └─────┘  └─────────┘   │
└─────────────────┬───────────────────────────┘
                  │ Pull model: Prometheus
                  │ scrapes YOUR app
         ┌────────┼────────┐
         ▼        ▼        ▼
    ┌────────┐ ┌───────┐ ┌────────┐
    │ App 1  │ │ App 2 │ │ App 3  │
    │/metrics│ │/metric│ │/metric │
    └────────┘ └───────┘ └────────┘
```

### Instrumenting Your App

#### TypeScript (prom-client)

```typescript
import { Registry, Counter, Histogram, Gauge, collectDefaultMetrics } from 'prom-client';

const register = new Registry();

// Collect default Node.js metrics (CPU, memory, event loop)
collectDefaultMetrics({ register });

// Custom metrics
const httpRequestsTotal = new Counter({
  name: 'http_requests_total',
  help: 'Total HTTP requests',
  labelNames: ['method', 'path', 'status'],
  registers: [register],
});

const httpRequestDuration = new Histogram({
  name: 'http_request_duration_seconds',
  help: 'HTTP request duration in seconds',
  labelNames: ['method', 'path'],
  buckets: [0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10],
  registers: [register],
});

const activeConnections = new Gauge({
  name: 'active_connections',
  help: 'Number of active connections',
  registers: [register],
});

// Middleware to track requests
app.use((req, res, next) => {
  const start = Date.now();
  activeConnections.inc();
  
  res.on('finish', () => {
    const duration = (Date.now() - start) / 1000;
    httpRequestsTotal.inc({ 
      method: req.method, 
      path: req.route?.path || 'unknown', 
      status: res.statusCode 
    });
    httpRequestDuration.observe(
      { method: req.method, path: req.route?.path || 'unknown' }, 
      duration
    );
    activeConnections.dec();
  });
  
  next();
});

// Expose /metrics endpoint for Prometheus to scrape
app.get('/metrics', async (req, res) => {
  res.set('Content-Type', register.contentType);
  res.end(await register.metrics());
});
```

#### Go (prometheus/client_golang)

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    httpRequests = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    httpDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
)

func init() {
    prometheus.MustRegister(httpRequests, httpDuration)
}

// Expose metrics endpoint
http.Handle("/metrics", promhttp.Handler())
```

---

## 🟡 PromQL (Prometheus Query Language)

```promql
# Request rate (per second, averaged over 5 minutes)
rate(http_requests_total[5m])

# Request rate by status code
sum by (status) (rate(http_requests_total[5m]))

# Error rate percentage
sum(rate(http_requests_total{status=~"5.."}[5m])) 
  / sum(rate(http_requests_total[5m])) * 100

# p99 latency
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))

# p95 latency by endpoint
histogram_quantile(0.95, 
  sum by (path, le) (rate(http_request_duration_seconds_bucket[5m]))
)

# CPU usage percentage
100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)

# Memory usage percentage
(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) 
  / node_memory_MemTotal_bytes * 100

# Disk usage
(node_filesystem_size_bytes - node_filesystem_avail_bytes) 
  / node_filesystem_size_bytes * 100

# Top 5 endpoints by request count
topk(5, sum by (path) (rate(http_requests_total[5m])))
```

---

## 🟡 RED and USE Methods

### RED Method (for Services)

```
R = Rate    → Requests per second
E = Errors  → Errors per second (or error %)
D = Duration → Latency (p50, p90, p99)

Every service dashboard should show RED metrics.
```

### USE Method (for Resources)

```
U = Utilization → How busy? (CPU 72%, Memory 85%)
S = Saturation  → How queued? (Thread pool full, disk I/O wait)
E = Errors      → Hardware errors? (disk failures, network drops)

Every infrastructure dashboard should show USE metrics.
```

---

## 🔴 Common Mistakes

### ❌ High cardinality labels

```typescript
// BAD — user_id has millions of unique values
httpRequests.inc({ method: 'GET', user_id: userId });
// Prometheus creates a NEW time series for EACH user
// 1 million users = 1 million time series = Prometheus OOM crash

// GOOD — low cardinality labels only
httpRequests.inc({ method: 'GET', path: '/api/users', status: '200' });
// ~50 unique combinations = 50 time series = fine

// Rule: A label should have at most a few hundred unique values
// Never use: user_id, request_id, order_id, IP address as labels
```

### ❌ Wrong metric type

```typescript
// BAD — using gauge for request count
const requestCount = new Gauge({ name: 'requests' });
requestCount.set(totalRequests);
// If app restarts, count resets. Rate calculation breaks.

// GOOD — counter for things that only go up
const requestCount = new Counter({ name: 'http_requests_total' });
requestCount.inc();
// Prometheus handles rate() correctly even across restarts
```

---

**Previous:** [02. Structured Logging](./02-structured-logging.md)  
**Next:** [04. Distributed Tracing](./04-distributed-tracing.md)
