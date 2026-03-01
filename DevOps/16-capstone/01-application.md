# Capstone Phase 01: Application

> **Build a simple but production-grade REST API. Not a toy — a service with health checks, structured logging, and metrics from day one.**

---

## 🟢 What We're Building

```
A Task Management API:
  POST   /api/tasks       → Create a task
  GET    /api/tasks       → List all tasks
  GET    /api/tasks/:id   → Get a task
  PUT    /api/tasks/:id   → Update a task
  DELETE /api/tasks/:id   → Delete a task

Plus production endpoints:
  GET /health/live    → Liveness probe (am I alive?)
  GET /health/ready   → Readiness probe (can I serve traffic?)
  GET /metrics        → Prometheus metrics

This is deliberately simple. The VALUE is in how we
build, deploy, and operate it — not the business logic.
```

---

## 🟢 TypeScript Implementation

### Project Structure

```
task-service/
├── src/
│   ├── server.ts          # Entry point
│   ├── app.ts             # Express app setup
│   ├── routes/
│   │   ├── tasks.ts       # Task CRUD routes
│   │   └── health.ts      # Health check routes
│   ├── middleware/
│   │   ├── requestLogger.ts
│   │   └── errorHandler.ts
│   ├── metrics.ts         # Prometheus metrics
│   ├── logger.ts          # Structured logger
│   └── types.ts           # Type definitions
├── tests/
│   └── tasks.test.ts
├── Dockerfile
├── package.json
├── tsconfig.json
└── .env.example
```

### Entry Point

```typescript
// src/server.ts
import { app } from './app';
import { logger } from './logger';

const PORT = parseInt(process.env.PORT || '3000', 10);

const server = app.listen(PORT, () => {
  logger.info({ port: PORT }, 'Server started');
});

// Graceful shutdown
const shutdown = async (signal: string) => {
  logger.info({ signal }, 'Shutdown signal received');
  
  server.close(() => {
    logger.info('HTTP server closed');
    process.exit(0);
  });

  // Force shutdown after 10 seconds
  setTimeout(() => {
    logger.error('Forced shutdown after timeout');
    process.exit(1);
  }, 10000);
};

process.on('SIGTERM', () => shutdown('SIGTERM'));
process.on('SIGINT', () => shutdown('SIGINT'));
```

### Application Setup

```typescript
// src/app.ts
import express from 'express';
import { taskRoutes } from './routes/tasks';
import { healthRoutes } from './routes/health';
import { requestLogger } from './middleware/requestLogger';
import { errorHandler } from './middleware/errorHandler';
import { metricsMiddleware, metricsEndpoint } from './metrics';

export const app = express();

// Middleware
app.use(express.json());
app.use(requestLogger);
app.use(metricsMiddleware);

// Routes
app.use('/api/tasks', taskRoutes);
app.use('/health', healthRoutes);
app.get('/metrics', metricsEndpoint);

// Error handler (must be last)
app.use(errorHandler);
```

### Structured Logger

```typescript
// src/logger.ts
import pino from 'pino';

export const logger = pino({
  level: process.env.LOG_LEVEL || 'info',
  formatters: {
    level: (label) => ({ level: label }),
  },
  redact: {
    paths: ['req.headers.authorization', 'req.headers.cookie'],
    censor: '[REDACTED]',
  },
  // Pretty print in dev, JSON in production
  ...(process.env.NODE_ENV === 'development' && {
    transport: {
      target: 'pino-pretty',
      options: { colorize: true },
    },
  }),
});
```

### Prometheus Metrics

```typescript
// src/metrics.ts
import { Registry, Counter, Histogram, collectDefaultMetrics } from 'prom-client';
import { Request, Response, NextFunction } from 'express';

export const register = new Registry();

collectDefaultMetrics({ register });

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
  buckets: [0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5],
  registers: [register],
});

export const metricsMiddleware = (req: Request, res: Response, next: NextFunction) => {
  if (req.path === '/metrics' || req.path === '/health/live') {
    return next();
  }

  const start = Date.now();
  res.on('finish', () => {
    const duration = (Date.now() - start) / 1000;
    const path = req.route?.path || req.path;
    httpRequestsTotal.inc({ method: req.method, path, status: res.statusCode });
    httpRequestDuration.observe({ method: req.method, path }, duration);
  });
  next();
};

export const metricsEndpoint = async (_req: Request, res: Response) => {
  res.set('Content-Type', register.contentType);
  res.end(await register.metrics());
};
```

### Task Routes

```typescript
// src/routes/tasks.ts
import { Router, Request, Response } from 'express';
import { v4 as uuidv4 } from 'uuid';
import { logger } from '../logger';
import { Task } from '../types';

const router = Router();

// In-memory storage (replace with DB in real app)
const tasks = new Map<string, Task>();

router.post('/', (req: Request, res: Response) => {
  const { title, description } = req.body;
  
  if (!title) {
    return res.status(400).json({ error: 'Title is required' });
  }

  const task: Task = {
    id: uuidv4(),
    title,
    description: description || '',
    completed: false,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };

  tasks.set(task.id, task);
  logger.info({ taskId: task.id }, 'Task created');
  res.status(201).json(task);
});

router.get('/', (_req: Request, res: Response) => {
  const allTasks = Array.from(tasks.values());
  res.json({ tasks: allTasks, count: allTasks.length });
});

router.get('/:id', (req: Request, res: Response) => {
  const task = tasks.get(req.params.id);
  if (!task) {
    return res.status(404).json({ error: 'Task not found' });
  }
  res.json(task);
});

router.put('/:id', (req: Request, res: Response) => {
  const task = tasks.get(req.params.id);
  if (!task) {
    return res.status(404).json({ error: 'Task not found' });
  }

  const updated: Task = {
    ...task,
    ...req.body,
    id: task.id,
    createdAt: task.createdAt,
    updatedAt: new Date().toISOString(),
  };

  tasks.set(updated.id, updated);
  logger.info({ taskId: updated.id }, 'Task updated');
  res.json(updated);
});

router.delete('/:id', (req: Request, res: Response) => {
  if (!tasks.has(req.params.id)) {
    return res.status(404).json({ error: 'Task not found' });
  }
  tasks.delete(req.params.id);
  logger.info({ taskId: req.params.id }, 'Task deleted');
  res.status(204).send();
});

export const taskRoutes = router;
```

### Health Checks

```typescript
// src/routes/health.ts
import { Router, Request, Response } from 'express';

const router = Router();

router.get('/live', (_req: Request, res: Response) => {
  // Liveness: am I alive? (don't check dependencies)
  res.json({ status: 'alive', timestamp: new Date().toISOString() });
});

router.get('/ready', (_req: Request, res: Response) => {
  // Readiness: can I serve traffic?
  // In real app: check DB connection, cache, etc.
  res.json({ status: 'ready', timestamp: new Date().toISOString() });
});

export const healthRoutes = router;
```

### Request Logger Middleware

```typescript
// src/middleware/requestLogger.ts
import { Request, Response, NextFunction } from 'express';
import { logger } from '../logger';

export const requestLogger = (req: Request, res: Response, next: NextFunction) => {
  if (req.path === '/health/live' || req.path === '/metrics') {
    return next();
  }

  const start = Date.now();
  res.on('finish', () => {
    logger.info({
      method: req.method,
      path: req.path,
      statusCode: res.statusCode,
      durationMs: Date.now() - start,
      userAgent: req.get('user-agent'),
    }, 'HTTP request');
  });
  next();
};
```

### Error Handler

```typescript
// src/middleware/errorHandler.ts
import { Request, Response, NextFunction } from 'express';
import { logger } from '../logger';

export const errorHandler = (
  err: Error, 
  req: Request, 
  res: Response, 
  _next: NextFunction
) => {
  logger.error({
    error: err.message,
    stack: err.stack,
    method: req.method,
    path: req.path,
  }, 'Unhandled error');

  res.status(500).json({ error: 'Internal server error' });
};
```

### Types

```typescript
// src/types.ts
export interface Task {
  id: string;
  title: string;
  description: string;
  completed: boolean;
  createdAt: string;
  updatedAt: string;
}
```

---

## 🟡 Go Alternative

```go
// main.go — minimal Go version
package main

import (
    "encoding/json"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "context"
    "time"

    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    slog.SetDefault(logger)

    mux := http.NewServeMux()
    mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
    })
    mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
    })
    mux.Handle("GET /metrics", promhttp.Handler())
    // ... task routes

    server := &http.Server{Addr: ":3000", Handler: mux}
    
    go func() {
        slog.Info("Server started", "port", 3000)
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            slog.Error("Server error", "error", err)
            os.Exit(1)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    server.Shutdown(ctx)
    slog.Info("Server stopped")
}
```

---

## 🟢 Testing

```typescript
// tests/tasks.test.ts
import request from 'supertest';
import { app } from '../src/app';

describe('Task API', () => {
  let taskId: string;

  test('POST /api/tasks creates a task', async () => {
    const res = await request(app)
      .post('/api/tasks')
      .send({ title: 'Test task', description: 'A test' })
      .expect(201);
    
    expect(res.body.title).toBe('Test task');
    expect(res.body.id).toBeDefined();
    taskId = res.body.id;
  });

  test('GET /api/tasks lists tasks', async () => {
    const res = await request(app).get('/api/tasks').expect(200);
    expect(res.body.count).toBeGreaterThan(0);
  });

  test('GET /health/live returns alive', async () => {
    const res = await request(app).get('/health/live').expect(200);
    expect(res.body.status).toBe('alive');
  });

  test('GET /metrics returns prometheus metrics', async () => {
    const res = await request(app).get('/metrics').expect(200);
    expect(res.text).toContain('http_requests_total');
  });

  test('POST /api/tasks without title returns 400', async () => {
    await request(app)
      .post('/api/tasks')
      .send({ description: 'No title' })
      .expect(400);
  });
});
```

---

**Next:** [02. Containerization](./02-containerization.md)  
**Up:** [README](./README.md)
