# Structured Logging

> **Unstructured logs are for humans. Structured logs are for machines. In production, machines process your logs — so write for them.**

---

## 🟢 Unstructured vs Structured

```
Unstructured (human-readable, machine-hostile):
  [2024-01-15 14:32:01] INFO User john@example.com logged in from 192.168.1.1
  [2024-01-15 14:32:02] ERROR Failed to process payment for order #12345: timeout
  [2024-01-15 14:32:03] WARN Disk usage at 87% on /dev/sda1

Problems:
  - How do you search for all logins from a specific IP? Regex 😱
  - How do you count errors per order? Parse every line
  - How do you filter by user in a dashboard? Good luck
  - Different services log differently = no consistency

Structured (machine-readable, JSON):
  {"timestamp":"2024-01-15T14:32:01Z","level":"info","msg":"user logged in","user":"john@example.com","ip":"192.168.1.1","service":"auth"}
  {"timestamp":"2024-01-15T14:32:02Z","level":"error","msg":"payment failed","order_id":"12345","error":"timeout","service":"payment"}
  {"timestamp":"2024-01-15T14:32:03Z","level":"warn","msg":"disk usage high","percent":87,"mount":"/dev/sda1","service":"monitoring"}

Benefits:
  - Query: level="error" AND service="payment" → instant
  - Count: count by order_id → aggregation
  - Dashboard: filter by any field
  - Consistent across all services
```

---

## 🟢 Log Levels

```
TRACE  → Extremely detailed (loop iterations, variable values)
         Almost NEVER in production

DEBUG  → Development details (SQL queries, function args)
         Disabled in production (too noisy)

INFO   → Normal operations (user actions, requests, deploys)
         The default production level

WARN   → Something unexpected but handled (retry succeeded,
         fallback used, approaching limit)

ERROR  → Something failed, user was affected
         (failed request, unhandled exception)

FATAL  → Application is crashing, cannot continue
         (missing database, out of memory)

Production: INFO and above (INFO, WARN, ERROR, FATAL)
Debugging:  DEBUG and above (temporarily)
NEVER:      TRACE in production
```

---

## 🟢 What to Log

```yaml
# ALWAYS log:
- Request received (method, path, user, request_id)
- Request completed (status, duration, request_id)
- Errors with stack traces
- Authentication events (login, logout, failed attempts)
- Authorization failures (403s)
- External API calls (service, duration, status)
- Database query performance (slow queries > 100ms)
- Business events (order created, payment processed)
- Deployment events (version, deployer, timestamp)

# NEVER log:
- Passwords or credentials
- Session tokens or API keys
- Credit card numbers or SSNs
- Personal health information
- Full request bodies (may contain secrets)
- Anything covered by GDPR/HIPAA without consent
```

---

## 🟢 Structured Logging in Code

### TypeScript/Node.js (pino)

```typescript
import pino from 'pino';

const logger = pino({
  level: process.env.LOG_LEVEL || 'info',
  timestamp: pino.stdTimeFunctions.isoTime,
  // In production, output JSON (for log aggregators)
  // In development, pretty print for humans
  transport: process.env.NODE_ENV === 'development' 
    ? { target: 'pino-pretty' }
    : undefined,
});

// Basic usage
logger.info('Server started');
// {"level":30,"time":"2024-01-15T14:32:01.000Z","msg":"Server started"}

// With context
logger.info({ port: 3000, env: 'production' }, 'Server started');
// {"level":30,"time":"...","port":3000,"env":"production","msg":"Server started"}

// Error logging
try {
  await processPayment(orderId);
} catch (error) {
  logger.error({ 
    err: error,
    orderId,
    userId: req.user.id 
  }, 'Payment processing failed');
}

// Child logger (adds context to all messages)
const requestLogger = logger.child({ 
  requestId: req.id,
  userId: req.user?.id,
  method: req.method,
  path: req.url 
});

requestLogger.info('Request received');
requestLogger.info({ statusCode: 200, duration: 45 }, 'Request completed');
// All log lines include requestId, userId, method, path
```

### Go (zerolog)

```go
package main

import (
    "os"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func main() {
    // Production: JSON output
    // Development: pretty console output
    if os.Getenv("ENV") == "development" {
        log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
    }

    // Basic
    log.Info().Msg("Server started")
    // {"level":"info","time":"2024-01-15T14:32:01Z","message":"Server started"}

    // With fields
    log.Info().
        Str("service", "api").
        Int("port", 3000).
        Msg("Server started")

    // Error
    log.Error().
        Err(err).
        Str("order_id", orderID).
        Str("user_id", userID).
        Msg("Payment processing failed")

    // Sub-logger with context
    reqLogger := log.With().
        Str("request_id", requestID).
        Str("method", r.Method).
        Str("path", r.URL.Path).
        Logger()

    reqLogger.Info().Msg("Request received")
    reqLogger.Info().
        Int("status", 200).
        Dur("duration", elapsed).
        Msg("Request completed")
}
```

---

## 🟡 Request Logging Middleware

### Express.js

```typescript
import { Request, Response, NextFunction } from 'express';
import { randomUUID } from 'crypto';
import pino from 'pino';

const logger = pino({ level: 'info' });

function requestLogger(req: Request, res: Response, next: NextFunction) {
  const requestId = req.headers['x-request-id'] as string || randomUUID();
  const start = Date.now();
  
  // Attach logger and requestId to request
  req.log = logger.child({ requestId, method: req.method, path: req.path });
  req.requestId = requestId;
  
  // Set request ID in response header (for debugging)
  res.setHeader('x-request-id', requestId);
  
  req.log.info({ 
    userAgent: req.get('user-agent'),
    ip: req.ip 
  }, 'Request received');
  
  // Log when response finishes
  res.on('finish', () => {
    const duration = Date.now() - start;
    const logData = { 
      statusCode: res.statusCode, 
      duration,
      contentLength: res.get('content-length')
    };
    
    if (res.statusCode >= 500) {
      req.log.error(logData, 'Request completed with server error');
    } else if (res.statusCode >= 400) {
      req.log.warn(logData, 'Request completed with client error');
    } else {
      req.log.info(logData, 'Request completed');
    }
  });
  
  next();
}

app.use(requestLogger);
```

---

## 🟡 Log Aggregation Architecture

```
Application Pods → stdout/stderr
    │
    ▼
Log Collector (runs on every node)
    │  DaemonSet: Fluentd / Fluent Bit / Promtail / Vector
    │
    ▼
Log Storage / Query Engine
    │  Loki (lightweight, label-based)
    │  Elasticsearch (powerful, full-text search)
    │  CloudWatch Logs (AWS managed)
    │
    ▼
Visualization
    │  Grafana (with Loki/Elasticsearch)
    │  Kibana (with Elasticsearch)
    │  CloudWatch Console
    │
    ▼
Alerts
    AlertManager / PagerDuty / Slack

Kubernetes logging pipeline:
  1. App writes to stdout (not files!)
  2. Container runtime captures stdout
  3. Node-level agent (DaemonSet) ships logs
  4. Central storage indexes and stores
  5. Grafana/Kibana visualizes
```

### Why stdout, Not Files?

```
# ❌ BAD — writing to files in containers
app.log("/var/log/myapp/app.log", logData)
# Problems:
#  - Container restarts = logs lost
#  - Need volume mounts
#  - Need log rotation (logrotate)
#  - Log collector must read from files

# ✅ GOOD — writing to stdout/stderr
console.log(JSON.stringify(logData))
# Benefits:
#  - Kubernetes captures automatically
#  - kubectl logs works
#  - No disk management
#  - Container runtime handles rotation
#  - Standard 12-factor app practice
```

---

## 🟡 Log Filtering and Sampling

At scale, you can't keep every log line:

```
Development:   Log everything (DEBUG level)
Staging:       Log INFO and above
Production:    Log INFO and above, with rules:

  # High-volume, low-value → sample
  Health check logs:    Log 1% (1 of 100)
  Successful requests:  Log 100% but minimal fields
  
  # Low-volume, high-value → log everything
  Errors:               Log 100% with full context
  Auth failures:        Log 100% (security)
  Slow queries (>100ms): Log 100%
  Business events:      Log 100%
```

---

## 🔴 Anti-Patterns

### ❌ Logging sensitive data

```typescript
// BAD — password in logs
logger.info({ email, password }, 'Login attempt');

// BAD — full request body (may contain secrets)
logger.info({ body: req.body }, 'Request received');

// GOOD — safe fields only
logger.info({ email, ip: req.ip }, 'Login attempt');
```

### ❌ String concatenation instead of structured fields

```typescript
// BAD — unstructured, can't query by orderId
logger.info(`Order ${orderId} created by user ${userId} for $${amount}`);

// GOOD — structured, every field is queryable
logger.info({ orderId, userId, amount }, 'Order created');
```

### ❌ Logging in hot loops

```typescript
// BAD — 1 million log lines per second
for (const item of items) {
  logger.debug({ item }, 'Processing item');  // MILLION TIMES
  process(item);
}

// GOOD — log summary
logger.info({ count: items.length }, 'Processing batch');
const results = await processAll(items);
logger.info({ processed: results.success, failed: results.failed }, 'Batch complete');
```

---

**Previous:** [01. Three Pillars](./01-three-pillars.md)  
**Next:** [03. Metrics](./03-metrics.md)
