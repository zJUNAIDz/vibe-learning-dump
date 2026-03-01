# Distributed Tracing

> **A trace follows a single request as it travels through multiple services. It shows you EXACTLY where time is spent and where failures happen in a distributed system.**

---

## 🟢 Why Tracing Exists

```
Without tracing:
  User: "The checkout page is slow"
  You: *checks API gateway logs* — 200 OK, 2.3s response time
  You: *checks order service logs* — processed in 50ms
  You: *checks payment service logs* — processed in 80ms
  You: *checks inventory service logs* — processed in 30ms
  You: Where did 2.3 seconds go?! No idea.

With tracing:
  User: "The checkout page is slow"
  You: *opens trace*
  
  ├─ API Gateway          ──────────────────── 2300ms
  │  ├─ Auth middleware    ──── 200ms
  │  ├─ Order service      ──── 50ms
  │  ├─ Payment service    ─────────── 800ms
  │  │  └─ Stripe API call ────────── 780ms  ← HERE!
  │  ├─ Inventory service  ── 30ms
  │  └─ Email service      ────────────────── 1200ms  ← AND HERE
  │     └─ SMTP connection ────────────────── 1180ms
  
  Root cause: Stripe API slow (780ms) + SMTP connection slow (1180ms)
  Fix: Make email async (don't block checkout on sending receipt)
```

---

## 🟢 Core Concepts

### Trace, Span, and Context

```
TRACE = The full journey of one request through all services
  → Has a unique trace_id (e.g., abc-123-def-456)
  → Contains multiple spans

SPAN = One operation within the trace
  → Has a span_id
  → Has a parent_span_id (except root span)
  → Records: start time, duration, status, attributes, events

CONTEXT = The trace_id + span_id passed between services
  → Propagated via HTTP headers, gRPC metadata, message headers

Example headers:
  traceparent: 00-abc123def456-span789-01
  │              │   │            │      │
  │              │   │            │      └─ Flags (sampled)
  │              │   │            └─ Parent span ID
  │              │   └─ Trace ID
  │              └─ Version
  └─ W3C Trace Context standard
```

### Span Relationships

```
Trace: abc123

Root Span (API Gateway) ─────────────────── 2300ms
  │
  ├─ Child Span (auth) ─── 200ms
  │
  ├─ Child Span (order-service) ─── 50ms
  │     │
  │     └─ Child Span (postgres query) ── 15ms
  │
  ├─ Child Span (payment-service) ── 800ms
  │     │
  │     └─ Child Span (stripe-api) ── 780ms
  │
  └─ Child Span (email-service) ──── 1200ms
        │
        └─ Child Span (smtp) ──── 1180ms

Each span knows:
  → Its trace_id (which request)
  → Its span_id (which operation)
  → Its parent_span_id (who called it)
  → Start time and duration
  → Status (OK / ERROR)
  → Attributes (key-value metadata)
  → Events (logs attached to the span)
```

---

## 🟢 OpenTelemetry (OTel)

```
OpenTelemetry is THE standard for distributed tracing.
(Merged from OpenTracing and OpenCensus)

Components:
  → API: Interfaces for creating spans
  → SDK: Implementation of the API
  → Exporters: Send data to backends (Jaeger, Zipkin, etc.)
  → Collector: Receives, processes, and exports telemetry

         Your App                  Collector              Backend
  ┌─────────────────┐     ┌────────────────────┐    ┌──────────┐
  │ OTel SDK         │     │                    │    │          │
  │  → Creates spans │────→│ Receive → Process  │───→│  Jaeger  │
  │  → Propagates    │OTLP │  → Filter          │    │  Tempo   │
  │    context        │     │  → Sample          │    │  Zipkin  │
  │  → Exports        │     │  → Transform       │    │          │
  └─────────────────┘     └────────────────────┘    └──────────┘
```

### TypeScript Instrumentation

```typescript
import { NodeSDK } from '@opentelemetry/sdk-node';
import { getNodeAutoInstrumentations } from '@opentelemetry/auto-instrumentations-node';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { Resource } from '@opentelemetry/resources';
import { SemanticResourceAttributes } from '@opentelemetry/semantic-conventions';

// Initialize BEFORE importing your app code
const sdk = new NodeSDK({
  resource: new Resource({
    [SemanticResourceAttributes.SERVICE_NAME]: 'order-service',
    [SemanticResourceAttributes.SERVICE_VERSION]: '1.2.3',
    [SemanticResourceAttributes.DEPLOYMENT_ENVIRONMENT]: 'production',
  }),
  traceExporter: new OTLPTraceExporter({
    url: 'http://otel-collector:4318/v1/traces',
  }),
  instrumentations: [
    getNodeAutoInstrumentations({
      // Auto-instruments: HTTP, Express, pg, mysql, redis, etc.
      '@opentelemetry/instrumentation-http': {
        ignoreIncomingPaths: ['/health', '/metrics'],
      },
    }),
  ],
});

sdk.start();
```

### Creating Custom Spans

```typescript
import { trace, SpanStatusCode } from '@opentelemetry/api';

const tracer = trace.getTracer('order-service');

async function processOrder(orderId: string): Promise<Order> {
  // Create a span for this operation
  return tracer.startActiveSpan('processOrder', async (span) => {
    try {
      // Add attributes (metadata) to the span
      span.setAttribute('order.id', orderId);
      
      // Validate order (child span created automatically)
      const order = await validateOrder(orderId);
      span.setAttribute('order.total', order.total);
      span.setAttribute('order.items_count', order.items.length);
      
      // Process payment (this creates another child span)
      const payment = await processPayment(order);
      
      // Add event (point-in-time log within span)
      span.addEvent('payment_processed', {
        'payment.id': payment.id,
        'payment.method': payment.method,
      });
      
      span.setStatus({ code: SpanStatusCode.OK });
      return order;
    } catch (error) {
      // Record error in span
      span.setStatus({ 
        code: SpanStatusCode.ERROR, 
        message: error.message 
      });
      span.recordException(error);
      throw error;
    } finally {
      span.end(); // Always end the span
    }
  });
}

async function processPayment(order: Order): Promise<Payment> {
  // This span is automatically a child of processOrder
  return tracer.startActiveSpan('processPayment', async (span) => {
    try {
      span.setAttribute('payment.provider', 'stripe');
      span.setAttribute('payment.amount', order.total);
      
      const result = await stripe.charges.create({
        amount: order.total,
        currency: 'usd',
      });
      
      span.setStatus({ code: SpanStatusCode.OK });
      return result;
    } catch (error) {
      span.setStatus({ code: SpanStatusCode.ERROR });
      span.recordException(error);
      throw error;
    } finally {
      span.end();
    }
  });
}
```

### Go Instrumentation

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("order-service")

func ProcessOrder(ctx context.Context, orderID string) (*Order, error) {
    ctx, span := tracer.Start(ctx, "ProcessOrder",
        trace.WithAttributes(
            attribute.String("order.id", orderID),
        ),
    )
    defer span.End()

    order, err := validateOrder(ctx, orderID)
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        span.RecordError(err)
        return nil, err
    }

    span.SetAttributes(
        attribute.Float64("order.total", order.Total),
        attribute.Int("order.items_count", len(order.Items)),
    )

    // Context carries the trace — child spans link automatically
    payment, err := processPayment(ctx, order)
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        span.RecordError(err)
        return nil, err
    }

    span.AddEvent("payment_processed", trace.WithAttributes(
        attribute.String("payment.id", payment.ID),
    ))

    span.SetStatus(codes.Ok, "")
    return order, nil
}
```

---

## 🟡 Context Propagation

```
The MOST important concept in tracing.

When Service A calls Service B:
  1. Service A adds trace headers to the outgoing request
  2. Service B extracts trace headers from the incoming request
  3. Service B creates child spans under the same trace

Without propagation:
  Service A: trace=abc123 (spans only in A)
  Service B: trace=xyz789 (different trace! disconnected!)

With propagation:
  Service A: trace=abc123 → HTTP header → Service B
  Service B: trace=abc123 (same trace! connected!)

HTTP propagation (W3C Trace Context):
  traceparent: 00-abc123-span456-01

gRPC propagation:
  metadata: traceparent = 00-abc123-span456-01

Kafka propagation:
  message header: traceparent = 00-abc123-span456-01
```

### Propagation Across Async Boundaries

```typescript
import { propagation, context } from '@opentelemetry/api';

// When PUBLISHING to a message queue
function publishMessage(queue: string, payload: any) {
  const carrier: Record<string, string> = {};
  
  // Inject current trace context into carrier
  propagation.inject(context.active(), carrier);
  
  // Send carrier as message headers
  queue.publish({
    body: payload,
    headers: carrier,  // { traceparent: '00-abc123-span456-01' }
  });
}

// When CONSUMING from a message queue
function consumeMessage(message: QueueMessage) {
  // Extract trace context from message headers
  const parentContext = propagation.extract(
    context.active(), 
    message.headers
  );
  
  // Create span under the original trace
  context.with(parentContext, () => {
    tracer.startActiveSpan('processMessage', (span) => {
      // This span is now part of the ORIGINAL trace
      handleMessage(message);
      span.end();
    });
  });
}
```

---

## 🟡 Sampling

```
Tracing EVERY request is expensive at scale:
  → 10,000 req/s × 10 spans/req × 1KB/span = 100 MB/sec = 8.6 TB/day

Sampling strategies:

1. HEAD SAMPLING (decide at the start)
   → Sample 10% of traces randomly
   → Pros: Simple, predictable cost
   → Cons: Might miss rare errors

2. TAIL SAMPLING (decide at the end)
   → Keep 100% of error traces
   → Keep 100% of slow traces (>1s)
   → Sample 5% of successful fast traces
   → Pros: Never miss interesting traces
   → Cons: Requires collector buffering

3. RATE-BASED
   → Keep first 10 traces per second per service
   → Drop the rest
   → Predictable cost

Recommendation:
  → Development: 100% sampling
  → Staging: 100% sampling
  → Production: Tail sampling via OTel Collector
```

### OTel Collector Tail Sampling Config

```yaml
# otel-collector-config.yaml
processors:
  tail_sampling:
    decision_wait: 10s
    num_traces: 100000
    policies:
      # Keep all error traces
      - name: errors
        type: status_code
        status_code:
          status_codes: [ERROR]
      
      # Keep all slow traces (> 1 second)
      - name: slow-traces
        type: latency
        latency:
          threshold_ms: 1000
      
      # Sample 10% of everything else
      - name: probabilistic
        type: probabilistic
        probabilistic:
          sampling_percentage: 10
```

---

## 🔴 Common Mistakes

### ❌ Missing context propagation

```typescript
// BAD — HTTP call without propagating context
async function callPaymentService(order: Order) {
  // fetch() doesn't automatically propagate trace context!
  const res = await fetch('http://payment-service/charge', {
    method: 'POST',
    body: JSON.stringify(order),
  });
  // Payment service creates a DISCONNECTED trace
}

// GOOD — use instrumented HTTP client or propagate manually
// Option 1: Auto-instrumentation handles it (if configured)
// Option 2: Manual propagation
async function callPaymentService(order: Order) {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  propagation.inject(context.active(), headers);
  
  const res = await fetch('http://payment-service/charge', {
    method: 'POST',
    headers,
    body: JSON.stringify(order),
  });
}
```

### ❌ Too many spans / too few spans

```
Too many spans:
  → Creating a span for every function call
  → Creating spans inside tight loops
  → Result: traces with 10,000 spans = unusable

Too few spans:
  → Only root span, no children
  → Result: "request took 2.3s" with no breakdown

Right balance:
  → Span per network call (HTTP, DB, cache, queue)
  → Span per significant business operation
  → Span per async boundary
  → NO span for simple function calls or loops
```

### ❌ Not recording errors

```typescript
// BAD — catch error but don't record it in span
try {
  await processPayment(order);
} catch (error) {
  logger.error('Payment failed', error);
  // Span shows SUCCESS even though it failed!
}

// GOOD — always update span status on error
try {
  await processPayment(order);
} catch (error) {
  span.setStatus({ code: SpanStatusCode.ERROR, message: error.message });
  span.recordException(error);
  logger.error('Payment failed', error);
  throw error;
}
```

---

## 🔴 Kubernetes Deployment

```yaml
# OTel Collector as DaemonSet (one per node)
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: otel-collector
spec:
  selector:
    matchLabels:
      app: otel-collector
  template:
    metadata:
      labels:
        app: otel-collector
    spec:
      containers:
        - name: collector
          image: otel/opentelemetry-collector-contrib:latest
          ports:
            - containerPort: 4317  # gRPC OTLP
            - containerPort: 4318  # HTTP OTLP
          volumeMounts:
            - name: config
              mountPath: /etc/otelcol-contrib
      volumes:
        - name: config
          configMap:
            name: otel-collector-config
---
# App environment variables
env:
  - name: OTEL_SERVICE_NAME
    value: "order-service"
  - name: OTEL_EXPORTER_OTLP_ENDPOINT
    value: "http://otel-collector:4318"
  - name: OTEL_TRACES_SAMPLER
    value: "parentbased_traceidratio"
  - name: OTEL_TRACES_SAMPLER_ARG
    value: "0.1"  # 10% sampling
```

---

**Previous:** [03. Metrics](./03-metrics.md)  
**Next:** [05. Alerting](./05-alerting.md)
