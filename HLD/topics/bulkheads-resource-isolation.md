# Bulkheads & Resource Isolation

## 1. The Real Problem This Exists to Solve

When multiple features or services share the same resource pool (threads, connections, memory), a failure or overload in one area cascades to affect all other areas. Bulkheads prevent cascading failures by isolating resources.

Real production scenario:
- E-commerce application with shared database connection pool (100 connections)
- Normal operation:
  - Product browsing: 40 connections
  - Checkout: 30 connections
  - Analytics: 20 connections
  - Admin panel: 10 connections
- Analytics team deploys slow query (takes 60s instead of 0.1s)
- Analytics queries start piling up, consuming all 100 connections
- Within 10 seconds: All DB connections held by slow analytics queries
- Checkout cannot get DB connection
- Product browsing fails
- Admin panel unresponsive
- **Revenue impact: $50,000/minute in failed checkouts**
- **Root cause: One slow analytics query took down entire site**

**The fundamental problem**: Shared resources create tight coupling between independent components. Failure in one component consumes all resources, starving other components.

Without bulkheads:
- All components share same resource pool
- Misbehaving component can consume all resources
- Failures cascade across unrelated features
- Single point of failure
- No isolation between critical and non-critical operations

With bulkheads:
- Each component gets dedicated resource pool
- Failure isolated to one bulkhead
- Critical operations protected from non-critical failures
- Graceful degradation (analytics fails, but checkout works)
- Clear resource boundaries and accountability

## 2. The Naive / Incorrect Approaches (IMPORTANT)

### ❌ Incorrect Approach #1: Single Shared Connection Pool

```typescript
// Incorrect: All services share one connection pool
const pool = new Pool({
  max: 100,  // 100 connections shared by everyone
  host: 'db.example.com',
});

// Checkout service
app.post('/checkout', async (req, res) => {
  const client = await pool.connect();  // Competes with everyone
  try {
    await client.query('INSERT INTO orders...');
    res.json({ success: true });
  } finally {
    client.release();
  }
});

// Analytics service (slow queries)
app.get('/analytics/report', async (req, res) => {
  const client = await pool.connect();  // Takes from same pool
  try {
    // Slow query: 60 seconds
    await client.query('SELECT * FROM orders JOIN users JOIN products...');
    res.json(results);
  } finally {
    client.release();
  }
});

// Admin service
app.get('/admin/users', async (req, res) => {
  const client = await pool.connect();  // Starved when pool exhausted
  // ...
});
```

**Why it seems reasonable:**
- Simple configuration (one pool)
- Efficient resource utilization (all connections available to everyone)
- No need to predict resource allocation
- Standard database connection pattern

**How it breaks:**
- Analytics deploys slow query (60s execution time)
- Within seconds, analytics consumes all 100 connections
- Checkout service cannot get connection → customers can't checkout
- Admin panel cannot get connection → can't investigate issue
- Product browsing fails → entire site appears down
- All services blocked waiting for connection
- One bad query takes down entire application

**Production symptoms:**
- Monitoring shows: DB connection pool at 100/100
- All connections held by analytics queries
- Checkout endpoint timing out (can't get DB connection)
- Error logs: "TimeoutError: Unable to acquire connection within 5000ms"
- Revenue drops to $0 (checkout broken)
- Meanwhile analytics team unaware they caused outage
- Database itself healthy (low CPU, just slow queries)

### ❌ Incorrect Approach #2: Priority Queuing Without Resource Isolation

```typescript
// Incorrect: Priority queue but still shared resource pool
class PriorityConnectionPool {
  private pool: Pool;
  private queue: PriorityQueue<ConnectionRequest>;

  async getConnection(priority: number): Promise<Client> {
    return new Promise((resolve) => {
      this.queue.enqueue({ priority, resolve });
      this.tryAssignConnection();
    });
  }

  private tryAssignConnection() {
    if (this.pool.availableConnections > 0) {
      const request = this.queue.dequeue();  // Highest priority first
      const conn = this.pool.acquire();
      request.resolve(conn);
    }
  }
}

// Usage
app.post('/checkout', async (req, res) => {
  const conn = await priorityPool.getConnection(1);  // High priority
  // ...
});

app.get('/analytics', async (req, res) => {
  const conn = await priorityPool.getConnection(10);  // Low priority
  // ...
});
```

**Why it seems reasonable:**
- High-priority requests (checkout) get connections first
- Low-priority requests (analytics) wait
- Protects critical operations

**How it breaks:**
- If analytics queries are slow, they still hold connections
- High-priority checkout gets connection, but pool still depleted
- Once analytics has acquired connections, priority doesn't matter
- Priority only helps for acquiring, not for releasing
- If 100 slow analytics queries already running, checkout still starved
- No limit on how many connections analytics can acquire over time

**Production symptoms:**
- Priority queue shows 50 pending high-priority requests
- But all 100 connections held by low-priority analytics queries
- Priority queue can't help (connections already acquired)
- Checkout still failing

### ❌ Incorrect Approach #3: Rate Limiting Without Isolation

```typescript
// Incorrect: Limit request rate but share resources
const analyticsLimiter = rateLimit({ max: 10 });

app.get('/analytics', analyticsLimiter, async (req, res) => {
  const conn = await sharedPool.connect();
  await conn.query('SLOW QUERY...');
});
```

**Why it seems reasonable:**
- Limits number of analytics requests
- Prevents analytics from overwhelming system

**How it breaks:**
- Rate limits requests, not resource consumption
- 10 slow queries (60s each) can still consume 100 connections
- If queries are slow enough, rate limit doesn't prevent exhaustion
- Example: 10 req/s × 60s duration = 600 concurrent connections needed
- Still exhausts pool despite rate limit

**Production symptoms:**
- Rate limiter shows: 10 req/s (under limit)
- But connection pool exhausted
- Rate limiting wrong metric (requests vs connections)

### ❌ Incorrect Approach #4: Timeouts Without Isolation

```typescript
// Incorrect: Add timeouts but still share pool
const pool = new Pool({
  max: 100,
  connectionTimeoutMillis: 5000,  // 5s timeout
  idleTimeoutMillis: 30000,
});

app.get('/analytics', async (req, res) => {
  try {
    const client = await pool.connect();
    await client.query('SLOW QUERY...');  // Still takes 60s
  } catch (err) {
    res.status(500).json({ error: 'Timeout' });
  }
});
```

**Why it seems reasonable:**
- Timeout prevents indefinite waiting
- Clients not blocked forever

**How it breaks:**
- `connectionTimeoutMillis` is timeout to acquire connection, not query timeout
- Once connection acquired, query still runs for 60s
- Timeout only helps clients waiting for connection (returns error faster)
- Doesn't prevent pool exhaustion
- Analytics queries still hold all connections

**Production symptoms:**
- Clients get timeout errors quickly (good UX)
- But pool still exhausted (queries still running)
- Checkout fails with "connection timeout" error
- Underlying problem (pool exhaustion) not solved

## 3. Correct Mental Model (How It Actually Works)

Bulkheads partition resources into isolated pools, preventing failure in one area from affecting others. Named after ship bulkheads that prevent one compartment flooding from sinking entire ship.

### The Ship Analogy

```
Ship without bulkheads:
  [===================================]
  One hole → entire ship floods → sinks

Ship with bulkheads:
  [====|====|====|====|====]
  One hole → only one compartment floods → ship stays afloat
```

### Resource Pool Partitioning

```
Without bulkheads (100 connections shared):
  Pool: [100 connections]
  Analytics bug → consumes all 100 → checkout fails

With bulkheads:
  Checkout:    [50 connections] ← guaranteed, isolated
  Browse:      [30 connections] ← guaranteed, isolated
  Analytics:   [15 connections] ← isolated, can't steal from others
  Admin:       [5 connections]  ← guaranteed, even during incident

  Analytics bug → consumes 15 connections → other services unaffected
```

### The Isolation Guarantee

```
Key property: Failure in one bulkhead CANNOT affect other bulkheads

Checkout bulkhead:
  - 50 dedicated connections
  - Even if analytics exhausts its pool
  - Even if admin exhausts its pool
  - Checkout always has 50 connections available

This is STRONG ISOLATION
```

### Resource Allocation Strategy

```
Priority-based allocation:
  Critical operations:     50% of resources
  High priority:           30% of resources
  Medium priority:         15% of resources
  Low priority/Background: 5% of resources

Over-provisioning allowed for efficiency:
  If checkout using 20/50 connections
  Analytics can temporarily borrow unused 30
  But checkout can reclaim at any time
```

## 4. Correct Design & Algorithm

### Strategy 1: Dedicated Resource Pools

```
For each service/feature:
  Create isolated pool with fixed size
  
checkout_pool = new Pool({ max: 50 })
analytics_pool = new Pool({ max: 15 })
admin_pool = new Pool({ max: 5 })

Each service only uses its own pool
Failure isolated to that bulkhead
```

### Strategy 2: Thread Pool Isolation

```
For each service:
  Dedicated thread pool
  
checkout_threads = ThreadPoolExecutor(max_workers=50)
analytics_threads = ThreadPoolExecutor(max_workers=10)

Slow analytics task cannot block checkout threads
```

### Strategy 3: Semaphore-Based Isolation

```
For each resource type:
  Semaphore limiting concurrent access
  
checkout_semaphore = Semaphore(50)
analytics_semaphore = Semaphore(15)

Before using resource:
  acquire semaphore
  if failed: reject request
  else: proceed
```

### Strategy 4: Circuit Breaker per Bulkhead

```
Each bulkhead has circuit breaker:
  
If analytics bulkhead failing:
  Open circuit breaker
  Fast-fail analytics requests
  Protect analytics resources
  Other bulkheads unaffected
```

## 5. Full Production-Grade Implementation

```typescript
interface BulkheadConfig {
  name: string;
  maxConcurrency: number;
  queueSize: number;
  timeout: number;
  fallback?: () => Promise<any>;
}

interface BulkheadMetrics {
  activeRequests: number;
  queuedRequests: number;
  totalRequests: number;
  rejectedRequests: number;
  timeoutRequests: number;
  avgExecutionTime: number;
}

class Bulkhead {
  private config: BulkheadConfig;
  private semaphore: Semaphore;
  private queue: QueuedRequest[] = [];
  private activeRequests = 0;
  private metrics: BulkheadMetrics;
  private executionTimes: number[] = [];

  constructor(config: BulkheadConfig) {
    this.config = config;
    this.semaphore = new Semaphore(config.maxConcurrency);
    this.metrics = {
      activeRequests: 0,
      queuedRequests: 0,
      totalRequests: 0,
      rejectedRequests: 0,
      timeoutRequests: 0,
      avgExecutionTime: 0,
    };
  }

  /**
   * Execute function within bulkhead isolation
   */
  async execute<T>(fn: () => Promise<T>): Promise<T> {
    this.metrics.totalRequests++;

    // Check if we can acquire permit
    const acquired = this.semaphore.tryAcquire();
    
    if (!acquired) {
      // Check if we can queue
      if (this.queue.length >= this.config.queueSize) {
        this.metrics.rejectedRequests++;
        throw new BulkheadRejectedException(
          `Bulkhead ${this.config.name} at capacity: ` +
          `${this.activeRequests} active, ${this.queue.length} queued`
        );
      }

      // Queue the request
      return this.enqueue(fn);
    }

    // Execute immediately
    return this.doExecute(fn);
  }

  /**
   * Execute function with timeout and cleanup
   */
  private async doExecute<T>(fn: () => Promise<T>): Promise<T> {
    this.activeRequests++;
    this.metrics.activeRequests = this.activeRequests;
    
    const startTime = Date.now();
    
    try {
      // Execute with timeout
      const result = await this.withTimeout(fn(), this.config.timeout);
      
      const executionTime = Date.now() - startTime;
      this.recordExecutionTime(executionTime);
      
      return result;
    } catch (error) {
      if (error instanceof TimeoutError) {
        this.metrics.timeoutRequests++;
      }
      
      // Try fallback if available
      if (this.config.fallback) {
        return this.config.fallback();
      }
      
      throw error;
    } finally {
      this.activeRequests--;
      this.metrics.activeRequests = this.activeRequests;
      this.semaphore.release();
      
      // Process queued request if any
      this.processQueue();
    }
  }

  /**
   * Enqueue request to wait for available slot
   */
  private enqueue<T>(fn: () => Promise<T>): Promise<T> {
    return new Promise((resolve, reject) => {
      this.queue.push({
        fn,
        resolve,
        reject,
        enqueuedAt: Date.now(),
      });
      this.metrics.queuedRequests = this.queue.length;
    });
  }

  /**
   * Process next queued request
   */
  private processQueue(): void {
    if (this.queue.length === 0) return;
    
    const item = this.queue.shift()!;
    this.metrics.queuedRequests = this.queue.length;
    
    // Check if request timed out while queued
    const queueTime = Date.now() - item.enqueuedAt;
    if (queueTime > this.config.timeout) {
      this.metrics.timeoutRequests++;
      item.reject(new TimeoutError('Request timed out in queue'));
      return;
    }
    
    // Execute the request
    this.doExecute(item.fn)
      .then(item.resolve)
      .catch(item.reject);
  }

  /**
   * Execute promise with timeout
   */
  private withTimeout<T>(promise: Promise<T>, timeoutMs: number): Promise<T> {
    return Promise.race([
      promise,
      new Promise<T>((_, reject) =>
        setTimeout(() => reject(new TimeoutError('Execution timeout')), timeoutMs)
      ),
    ]);
  }

  /**
   * Record execution time for metrics
   */
  private recordExecutionTime(time: number): void {
    this.executionTimes.push(time);
    if (this.executionTimes.length > 100) {
      this.executionTimes.shift();
    }
    
    const sum = this.executionTimes.reduce((a, b) => a + b, 0);
    this.metrics.avgExecutionTime = sum / this.executionTimes.length;
  }

  /**
   * Get current metrics
   */
  getMetrics(): BulkheadMetrics {
    return { ...this.metrics };
  }

  /**
   * Reset metrics
   */
  resetMetrics(): void {
    this.metrics.totalRequests = 0;
    this.metrics.rejectedRequests = 0;
    this.metrics.timeoutRequests = 0;
    this.executionTimes = [];
  }
}

class Semaphore {
  private permits: number;
  private maxPermits: number;

  constructor(permits: number) {
    this.permits = permits;
    this.maxPermits = permits;
  }

  tryAcquire(): boolean {
    if (this.permits > 0) {
      this.permits--;
      return true;
    }
    return false;
  }

  release(): void {
    if (this.permits < this.maxPermits) {
      this.permits++;
    }
  }

  available(): number {
    return this.permits;
  }
}

class BulkheadRejectedException extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'BulkheadRejectedException';
  }
}

class TimeoutError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'TimeoutError';
  }
}

interface QueuedRequest {
  fn: () => Promise<any>;
  resolve: (value: any) => void;
  reject: (error: Error) => void;
  enqueuedAt: number;
}

// Database connection pool bulkheads
class DatabaseBulkheads {
  private pools: Map<string, Pool>;
  private bulkheads: Map<string, Bulkhead>;

  constructor() {
    this.pools = new Map();
    this.bulkheads = new Map();
    this.initializeBulkheads();
  }

  private initializeBulkheads(): void {
    // Checkout bulkhead - critical, gets most resources
    this.pools.set('checkout', new Pool({
      max: 50,
      host: 'db.example.com',
    }));
    this.bulkheads.set('checkout', new Bulkhead({
      name: 'checkout',
      maxConcurrency: 50,
      queueSize: 100,
      timeout: 5000,
    }));

    // Browse bulkhead - high traffic, medium resources
    this.pools.set('browse', new Pool({
      max: 30,
      host: 'db.example.com',
    }));
    this.bulkheads.set('browse', new Bulkhead({
      name: 'browse',
      maxConcurrency: 30,
      queueSize: 50,
      timeout: 3000,
      fallback: async () => ({ cached: true, products: [] }),
    }));

    // Analytics bulkhead - can fail, gets fewer resources
    this.pools.set('analytics', new Pool({
      max: 15,
      host: 'db.example.com',
    }));
    this.bulkheads.set('analytics', new Bulkhead({
      name: 'analytics',
      maxConcurrency: 15,
      queueSize: 10,
      timeout: 30000,  // Analytics queries can be slow
      fallback: async () => ({ error: 'Analytics temporarily unavailable' }),
    }));

    // Admin bulkhead - low traffic but must always work
    this.pools.set('admin', new Pool({
      max: 5,
      host: 'db.example.com',
    }));
    this.bulkheads.set('admin', new Bulkhead({
      name: 'admin',
      maxConcurrency: 5,
      queueSize: 5,
      timeout: 10000,
    }));
  }

  /**
   * Execute query within bulkhead
   */
  async query<T>(
    bulkheadName: string,
    queryFn: (client: PoolClient) => Promise<T>
  ): Promise<T> {
    const pool = this.pools.get(bulkheadName);
    const bulkhead = this.bulkheads.get(bulkheadName);

    if (!pool || !bulkhead) {
      throw new Error(`Unknown bulkhead: ${bulkheadName}`);
    }

    return bulkhead.execute(async () => {
      const client = await pool.connect();
      try {
        return await queryFn(client);
      } finally {
        client.release();
      }
    });
  }

  /**
   * Get metrics for all bulkheads
   */
  getAllMetrics(): Record<string, BulkheadMetrics> {
    const metrics: Record<string, BulkheadMetrics> = {};
    
    for (const [name, bulkhead] of this.bulkheads) {
      metrics[name] = bulkhead.getMetrics();
    }
    
    return metrics;
  }
}

// Global instance
const dbBulkheads = new DatabaseBulkheads();

// Express routes using bulkheads
app.post('/checkout', async (req, res) => {
  try {
    const result = await dbBulkheads.query('checkout', async (client) => {
      return await client.query(
        'INSERT INTO orders (user_id, total) VALUES ($1, $2) RETURNING id',
        [req.user.id, req.body.total]
      );
    });
    
    res.json({ orderId: result.rows[0].id });
  } catch (error) {
    if (error instanceof BulkheadRejectedException) {
      res.status(503).json({ error: 'Checkout temporarily unavailable' });
    } else if (error instanceof TimeoutError) {
      res.status(504).json({ error: 'Checkout timeout' });
    } else {
      res.status(500).json({ error: 'Internal error' });
    }
  }
});

app.get('/products', async (req, res) => {
  try {
    const result = await dbBulkheads.query('browse', async (client) => {
      return await client.query('SELECT * FROM products LIMIT 50');
    });
    
    res.json(result.rows);
  } catch (error) {
    if (error instanceof BulkheadRejectedException) {
      // Fallback to cache
      res.json({ cached: true, products: await getFromCache('products') });
    } else {
      res.status(500).json({ error: 'Error loading products' });
    }
  }
});

app.get('/analytics/report', async (req, res) => {
  try {
    const result = await dbBulkheads.query('analytics', async (client) => {
      // Potentially slow query
      return await client.query(`
        SELECT 
          date_trunc('day', created_at) as day,
          COUNT(*) as orders,
          SUM(total) as revenue
        FROM orders
        WHERE created_at > NOW() - INTERVAL '90 days'
        GROUP BY day
        ORDER BY day
      `);
    });
    
    res.json(result.rows);
  } catch (error) {
    if (error instanceof BulkheadRejectedException) {
      res.status(503).json({ 
        error: 'Analytics overloaded',
        message: 'Please try again later'
      });
    } else if (error instanceof TimeoutError) {
      res.status(504).json({ error: 'Query timeout' });
    } else {
      res.status(500).json({ error: 'Analytics error' });
    }
  }
});

app.get('/admin/users', async (req, res) => {
  try {
    const result = await dbBulkheads.query('admin', async (client) => {
      return await client.query('SELECT * FROM users LIMIT 100');
    });
    
    res.json(result.rows);
  } catch (error) {
    if (error instanceof BulkheadRejectedException) {
      res.status(503).json({ error: 'Admin panel overloaded' });
    } else {
      res.status(500).json({ error: 'Error loading users' });
    }
  }
});

// Metrics endpoint
app.get('/metrics/bulkheads', (req, res) => {
  const metrics = dbBulkheads.getAllMetrics();
  
  res.json({
    bulkheads: Object.entries(metrics).map(([name, m]) => ({
      name,
      active: m.activeRequests,
      queued: m.queuedRequests,
      total: m.totalRequests,
      rejected: m.rejectedRequests,
      rejectionRate: (m.rejectedRequests / Math.max(m.totalRequests, 1) * 100).toFixed(2) + '%',
      avgExecutionMs: m.avgExecutionTime.toFixed(0),
    })),
  });
});
```

## 6. Correct Usage Patterns (Where This Shines)

### Pattern 1: Multi-Tenant SaaS with Customer Isolation

```typescript
// Each customer gets isolated bulkhead
class CustomerBulkheads {
  private bulkheads = new Map<string, Bulkhead>();

  getBulkhead(customerId: string): Bulkhead {
    if (!this.bulkheads.has(customerId)) {
      this.bulkheads.set(customerId, new Bulkhead({
        name: `customer-${customerId}`,
        maxConcurrency: 10,
        queueSize: 20,
        timeout: 5000,
      }));
    }
    return this.bulkheads.get(customerId)!;
  }
}

app.use(async (req, res, next) => {
  const bulkhead = customerBulkheads.getBulkhead(req.customer.id);
  
  try {
    await bulkhead.execute(async () => {
      await next();
    });
  } catch (error) {
    if (error instanceof BulkheadRejectedException) {
      res.status(429).json({ error: 'Too many concurrent requests' });
    }
  }
});
```

**Why this works:**
- Customer A's traffic spike doesn't affect Customer B
- Noisy neighbor problem solved
- Fair resource allocation per customer

### Pattern 2: Microservices with Per-Dependency Bulkheads

```typescript
// Isolate each downstream service
const paymentBulkhead = new Bulkhead({
  name: 'payment-service',
  maxConcurrency: 20,
  timeout: 5000,
});

const inventoryBulkhead = new Bulkhead({
  name: 'inventory-service',
  maxConcurrency: 50,
  timeout: 3000,
});

const shippingBulkhead = new Bulkhead({
  name: 'shipping-service',
  maxConcurrency: 30,
  timeout: 10000,
});

async function createOrder(orderData: any) {
  // Each external call isolated
  const payment = await paymentBulkhead.execute(() =>
    paymentService.charge(orderData.payment)
  );
  
  const inventory = await inventoryBulkhead.execute(() =>
    inventoryService.reserve(orderData.items)
  );
  
  const shipping = await shippingBulkhead.execute(() =>
    shippingService.schedule(orderData.address)
  );
  
  return { payment, inventory, shipping };
}
```

**Why this works:**
- If payment service slow, only payment bulkhead affected
- Inventory and shipping calls still work
- Partial degradation instead of complete failure

### Pattern 3: Thread Pool Isolation for CPU-Intensive Tasks

```typescript
// Separate thread pools for different task types
const imageBulkhead = new Bulkhead({
  name: 'image-processing',
  maxConcurrency: 4,  // CPU-bound, limit to core count
  queueSize: 100,
  timeout: 30000,
});

const videoBulkhead = new Bulkhead({
  name: 'video-processing',
  maxConcurrency: 2,  // Very CPU intensive
  queueSize: 50,
  timeout: 300000,
});

const apiCallsBulkhead = new Bulkhead({
  name: 'api-calls',
  maxConcurrency: 100,  // I/O bound, can be high
  queueSize: 500,
  timeout: 10000,
});
```

**Why this works:**
- CPU-intensive video processing can't starve I/O-bound API calls
- Different resource characteristics handled appropriately
- System remains responsive even when video processing backlogged

## 7. Failure Modes & Edge Cases

### Resource Starvation from Conservative Sizing

**Problem:** Bulkhead sized too small, rejects requests despite available system resources.

**Symptoms:**
- Checkout bulkhead: 50 concurrent → rejecting requests
- System CPU: 20% (plenty of capacity)
- Database: 50/1000 connections used
- Artificial limit causing unnecessary rejections

**Mitigation:**
- Monitor rejection rates and system utilization
- Increase bulkhead size if rejections high but resources available
- Dynamic resizing based on observed capacity

### Cascade Failure from Shared Dependencies

**Problem:** Bulkheads isolated but share downstream dependency.

**Symptoms:**
- Checkout bulkhead: 50 connections
- Analytics bulkhead: 15 connections
- Both hit same database
- Database overloaded from combined load
- Both bulkheads fail

**Mitigation:**
- Bulkheads at every layer (database also needs bulkheads)
- Separate read replicas for analytics
- Circuit breakers on shared dependencies

### Inefficient Resource Utilization

**Problem:** Strict partitioning wastes resources during uneven load.

**Symptoms:**
- Checkout bulkhead: 10/50 used (40 wasted)
- Analytics bulkhead: 15/15 used (rejecting requests)
- Total resources available but not utilized

**Mitigation:**
- Allow bulkheads to borrow unused capacity
- Implement guarantees (reserved) + limits (max burst)
- Elastic bulkheads that can grow into shared pool

### Queue Buildup During Overload

**Problem:** Queue fills up, requests wait long time before rejection.

**Symptoms:**
- Request queued for 30 seconds
- Finally gets rejected (timeout)
- Poor user experience (long wait then failure)

**Mitigation:**
- Set reasonable queue sizes
- Fast-fail if queue too deep
- Monitor queue depth and alert

## 8. Performance Characteristics & Tradeoffs

### Fault Isolation

**Without bulkheads:**
- One bad component → entire system fails
- Mean time between failures (MTBF): Low

**With bulkheads:**
- One bad component → only that bulkhead fails
- Other components continue working
- MTBF: High (for unaffected components)

### Resource Efficiency

**Shared pool (100 connections):**
- Dynamic allocation, efficient
- But no isolation

**Bulkheads (50+30+15+5 = 100 connections):**
- Fixed allocation, potentially less efficient
- But strong isolation guarantees
- May waste resources during uneven load

**Trade-off:** Safety vs efficiency

### Latency Impact

**Without bulkheads:**
- Normal: 50ms
- During incident: 30 seconds (queuing behind slow queries)

**With bulkheads:**
- Normal: 50ms
- During incident in other bulkhead: 50ms (unaffected)
- During incident in own bulkhead: Fast-fail (immediate rejection)

## 9. Foot-Guns & Common Mistakes (DO NOT SKIP)

### Mistake 1: Too Many Bulkheads

**Why engineers do it:** Want perfect isolation for every endpoint.

**What breaks:** 50 endpoints × 10 connections = 500 connections (too many).

**Fix:** Group related endpoints into bulkheads (3-5 bulkheads total).

### Mistake 2: Forgetting to Isolate Thread Pools

**Why engineers do it:** Only isolate connection pools, forget thread pools.

**What breaks:** Slow analytics query blocks all threads despite isolated connections.

**Fix:** Isolate both connection pools AND thread pools.

### Mistake 3: Not Monitoring Bulkhead Health

**Why engineers do it:** Set up bulkheads and forget about them.

**What breaks:** Don't notice when bulkhead rejection rate high.

**Fix:** Alert on rejection rate >5% for critical bulkheads.

### Mistake 4: Symmetric Bulkheads

**Why engineers do it:** Give all bulkheads equal size.

**What breaks:** Critical service (checkout) gets same resources as non-critical (analytics).

**Fix:** Size bulkheads based on criticality and expected load.

### Mistake 5: No Fallback Strategy

**Why engineers do it:** Assume bulkhead rejection won't happen.

**What breaks:** User sees error when bulkhead full.

**Fix:** Implement fallbacks (cached data, degraded experience).

## 10. When NOT to Use This (Anti-Patterns)

### Anti-Pattern 1: Homogeneous Load

If all traffic has same characteristics, bulkheads add complexity without benefit.

### Anti-Pattern 2: Monolithic Application

Small single-server application doesn't need bulkheads (not enough resources to partition).

### Anti-Pattern 3: Already Using Separate Services

If each component is already a separate microservice with its own resources, bulkheads redundant.

## 11. Related Concepts (With Contrast)

### Circuit Breaker

**Difference:** Circuit breaker stops calling failing service. Bulkhead prevents one service from starving resources from another.

**Together:** Use both. Bulkhead isolates resources, circuit breaker detects failures.

### Rate Limiting

**Difference:** Rate limiting controls incoming request rate. Bulkhead limits concurrent resource usage.

### Load Shedding

**Difference:** Load shedding drops low-priority requests. Bulkhead isolates resources for different components.

## 12. Production Readiness Checklist

### Resource Planning

- [ ] Identify components needing isolation
- [ ] Calculate resource requirements per component
- [ ] Size bulkheads based on priority and load
- [ ] Total resources = sum of bulkheads + buffer

### Configuration

- [ ] Set maxConcurrency per bulkhead
- [ ] Set queueSize (should be small, fast-fail better)
- [ ] Set timeout appropriate for operation
- [ ] Define fallback strategies

### Monitoring

- [ ] Track active requests per bulkhead
- [ ] Track rejection rate per bulkhead
- [ ] Alert on rejection rate >5% for critical bulkheads
- [ ] Dashboard showing bulkhead utilization

### Testing

- [ ] Load test each bulkhead independently
- [ ] Verify isolation (overload one, others unaffected)
- [ ] Test fallback behavior
- [ ] Test with combined load on all bulkheads

### Incident Response

- [ ] Document which features in which bulkhead
- [ ] Runbook for bulkhead exhaustion
- [ ] Ability to increase bulkhead size dynamically
- [ ] Circuit breaker integration
