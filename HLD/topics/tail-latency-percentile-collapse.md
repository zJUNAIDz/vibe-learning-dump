# Tail Latency & Percentile Collapse

## 1. The Real Problem This Exists to Solve

In a distributed system with fan-out patterns (one request spawning multiple backend calls), even low per-service failure rates compound catastrophically. When you aggregate responses from N independent services, the overall latency is dominated by the slowest response.

Real production scenario:
- A user request hits your API gateway
- The gateway fans out to 10 backend services in parallel
- Each service has p99 latency of 100ms (meaning 99% of requests complete in ≤100ms)
- You assume your gateway's p99 will be ~100ms
- **Reality: Your gateway's p99 is actually 650ms+**

This is percentile collapse. The actual behavior is:
- If each service has a 1% chance of being slow (p99)
- The probability that ALL 10 services are fast is: (0.99)^10 = 90.4%
- This means 9.6% of requests hit at least one slow service
- Your gateway's p90 latency is now determined by individual services' p99 latency
- Your gateway's p99 is determined by individual services' p99.9+ latency

Without understanding this, engineers build systems that look healthy in per-service metrics but provide terrible user experience. A checkout flow hitting 20 services will have 18% of requests hitting at least one slow backend.

## 2. The Naive / Incorrect Approaches (IMPORTANT)

### ❌ Incorrect Approach #1: Ignoring Fan-Out Multiplier

```typescript
// Incorrect: Just fan out and wait for all responses
async function getUserDashboard(userId: string): Promise<Dashboard> {
  const [profile, orders, recommendations, notifications, credits] = 
    await Promise.all([
      userService.getProfile(userId),
      orderService.getOrders(userId),
      recommendationService.getRecommendations(userId),
      notificationService.getNotifications(userId),
      creditService.getCredits(userId),
    ]);

  return { profile, orders, recommendations, notifications, credits };
}
```

**Why it seems reasonable:**
- Promise.all runs requests in parallel, maximizing throughput
- Each individual service has "good" p99 latency
- Code is simple and readable

**How it breaks:**
- With 5 services at p99=100ms (1% slow), 4.9% of dashboard requests will hit ≥1 slow service
- The dashboard's p95 latency is now ~100ms (individual service p99)
- The dashboard's p99 latency is ~200-300ms (individual service p99.9)
- As you add more services to the dashboard, tail latency degrades exponentially

**Production symptoms:**
- Individual services show p99 < 100ms in metrics
- Dashboard endpoint shows p99 > 500ms
- Engineers blame "network overhead" or "infrastructure issues"
- Users complain about slow page loads despite good per-service metrics
- Load increases cause dramatic p99 degradation even though p50 stays flat

### ❌ Incorrect Approach #2: Setting Uniform Timeouts

```typescript
// Incorrect: Set the same timeout for all calls
const TIMEOUT = 5000; // 5 seconds

async function getUserDashboard(userId: string): Promise<Dashboard> {
  const [profile, orders, recommendations, notifications, credits] = 
    await Promise.all([
      fetchWithTimeout(userService.getProfile(userId), TIMEOUT),
      fetchWithTimeout(orderService.getOrders(userId), TIMEOUT),
      fetchWithTimeout(recommendationService.getRecommendations(userId), TIMEOUT),
      fetchWithTimeout(notificationService.getNotifications(userId), TIMEOUT),
      fetchWithTimeout(creditService.getCredits(userId), TIMEOUT),
    ]);

  return { profile, orders, recommendations, notifications, credits };
}

function fetchWithTimeout<T>(promise: Promise<T>, timeout: number): Promise<T> {
  return Promise.race([
    promise,
    new Promise<never>((_, reject) => 
      setTimeout(() => reject(new Error('Timeout')), timeout)
    ),
  ]);
}
```

**Why it seems reasonable:**
- Prevents requests from hanging forever
- Simple to implement uniformly across all services
- Seems to limit worst-case latency

**How it breaks:**
- The timeout is either too aggressive (causes false positives) or too permissive (doesn't help tail latency)
- If timeout is 5s and p99 is 100ms, you're allowing requests to be 50x slower than typical
- The timeout fires only after the damage is done (user waited 5s)
- Doesn't distinguish between critical and non-critical dependencies
- Timeout errors still fail the entire request (no partial degradation)

**Production symptoms:**
- Either: Frequent timeout errors when services have brief latency spikes
- Or: Timeouts don't prevent bad user experience (users still wait 3-5 seconds)
- Debugging is hard because timeouts hide the root cause (which service was actually slow?)
- Retry logic interacts badly with timeouts (retry after 5s timeout makes total latency 10s)

### ❌ Incorrect Approach #3: Averaging Across Requests

```typescript
// Incorrect: Optimize for average latency
class MetricsCollector {
  private latencies: number[] = [];

  recordLatency(latency: number) {
    this.latencies.push(latency);
  }

  getAverageLatency(): number {
    return this.latencies.reduce((a, b) => a + b, 0) / this.latencies.length;
  }

  isHealthy(): boolean {
    return this.getAverageLatency() < 200; // Average under 200ms = healthy
  }
}
```

**Why it seems reasonable:**
- Average (mean) latency is easy to calculate and understand
- Shows overall system performance
- Single number to track in dashboards

**How it breaks:**
- Average hides the distribution completely
- A service with average 50ms but p99 of 10s looks "healthy"
- 99% of users having a good experience means 1 in 100 users suffer
- At scale (millions of requests), 1% is tens of thousands of bad experiences
- Engineering decisions based on averages optimize the wrong thing

**Production symptoms:**
- Metrics show "everything is fine" (avg latency 80ms)
- User complaints pour in about slow requests
- Customer support sees 1-5% of users reporting issues
- No correlation between metric improvements and user satisfaction
- SLA breaches despite "healthy" average latency

## 3. Correct Mental Model (How It Actually Works)

Tail latency arises from the multiplicative nature of distributed calls and the heavy-tailed distribution of response times.

### The Math

For N independent services, each with probability p of being fast:
- Probability all are fast: p^N
- Probability ≥1 is slow: 1 - p^N

Example with 10 services at p=0.99 (p99 = 1% slow):
- All fast: 0.99^10 = 90.4%
- At least one slow: 9.6%

This means your aggregate p90 = individual p99. This is percentile collapse.

### The Distribution

Real service latencies follow a heavy-tailed distribution (not normal):

```
Most requests:  ████████████████████████████████ (10-50ms)
p90-p99:        ████ (50-200ms)
p99-p99.9:      ██ (200-1000ms)
p99.9+:         █ (1s-30s) ← caused by GC, network retries, DB locks, cold caches
```

The tail is caused by:
- Garbage collection pauses (10-500ms)
- Packet loss requiring TCP retransmission (200ms-2s)
- Query hitting a cold cache or slow replica (100ms-5s)
- Database lock contention (100ms-30s)
- Background compaction/migration (1-10s)

These are **not rare edge cases**. At millions of requests per day, p99.9 events happen thousands of times.

### Fan-Out Amplification

```
User Request
     │
     ├─▶ Service A (p99=100ms) ──┐
     ├─▶ Service B (p99=100ms) ──┤
     ├─▶ Service C (p99=100ms) ──┼─▶ Wait for ALL
     ├─▶ Service D (p99=100ms) ──┤
     └─▶ Service E (p99=100ms) ──┘
          
Result: Aggregate p95 ≈ 100ms (individual p99)
        Aggregate p99 ≈ 500ms (individual p99.9+)
```

Each additional service in the fan-out:
- Increases probability of hitting a slow request
- Pushes aggregate percentiles toward higher individual percentiles
- Makes system sensitivity to tail latency exponential

### Retry Amplification

Retries interact multiplicatively with tail latency:

```
Request → Try 1 (slow/timeout) → Retry → Try 2 (fast) → Total latency = timeout + fast

If timeout=1s and p99=100ms:
- Original p99 latency: 100ms
- With 1 retry on timeout: p99 includes (1s timeout + 100ms retry) = 1.1s
- User experiences 10x worse latency
```

## 4. Correct Design & Algorithm

The correct approach requires multiple strategies working together:

### Strategy 1: Aggressive Timeouts on Non-Critical Paths

Set timeouts based on percentile targets, not worst-case behavior:
- Timeout should be ~2x p99 latency (not 50x)
- Critical dependencies: no timeout or very high timeout
- Non-critical dependencies: aggressive timeout (2-3x p99)

```typescript
const timeouts = {
  critical: 5000,    // Profile, auth (required for page)
  important: 500,    // Orders, primary content
  optional: 200,     // Recommendations, analytics
};
```

### Strategy 2: Graceful Degradation with Partial Responses

Return partial results when non-critical services timeout:

```typescript
async function getDashboard(userId: string): Promise<Dashboard> {
  const [profile, ordersResult, recsResult] = await Promise.allSettled([
    fetchWithTimeout(getProfile(userId), timeouts.critical),
    fetchWithTimeout(getOrders(userId), timeouts.important),
    fetchWithTimeout(getRecommendations(userId), timeouts.optional),
  ]);

  return {
    profile: profile.status === 'fulfilled' ? profile.value : null,
    orders: ordersResult.status === 'fulfilled' ? ordersResult.value : [],
    recommendations: recsResult.status === 'fulfilled' ? recsResult.value : [],
  };
}
```

### Strategy 3: Hedged Requests

For critical paths, issue duplicate requests and take the first response:

```typescript
async function hedgedRequest<T>(
  fn: () => Promise<T>,
  hedgeDelay: number
): Promise<T> {
  return Promise.race([
    fn(),
    sleep(hedgeDelay).then(() => fn()),
  ]);
}
```

### Strategy 4: Percentile-Aware Load Balancing

Route requests away from backends showing high tail latency:
- Track p99 per backend instance
- Reduce traffic to instances with p99 > threshold
- Avoid making tail latency worse through retries

## 5. Full Production-Grade Implementation

```typescript
import { performance } from 'perf_hooks';

interface TimeoutConfig {
  critical: number;
  important: number;
  optional: number;
}

interface ServiceCall<T> {
  name: string;
  fn: () => Promise<T>;
  timeout: number;
  required: boolean;
  hedgeDelay?: number;
}

interface CallResult<T> {
  success: boolean;
  value?: T;
  error?: Error;
  latency: number;
  hedged: boolean;
}

class TailLatencyOptimizer {
  private metrics: Map<string, number[]> = new Map();
  private readonly maxMetricsSize = 1000;

  /**
   * Execute multiple service calls with tail latency optimizations:
   * - Aggressive timeouts based on criticality
   * - Hedged requests for critical paths
   * - Graceful degradation for optional calls
   * - Latency tracking for monitoring
   */
  async executeFanOut<T extends Record<string, any>>(
    calls: ServiceCall<any>[]
  ): Promise<{ data: Partial<T>; results: Map<string, CallResult<any>> }> {
    const results = new Map<string, CallResult<any>>();

    const promises = calls.map(async (call) => {
      const startTime = performance.now();
      let hedged = false;

      try {
        let result: any;

        if (call.hedgeDelay !== undefined) {
          // Critical path: use hedged request
          result = await this.hedgedRequest(call.fn, call.hedgeDelay);
          hedged = true;
        } else {
          // Standard request with timeout
          result = await this.withTimeout(call.fn(), call.timeout, call.name);
        }

        const latency = performance.now() - startTime;
        this.recordLatency(call.name, latency);

        results.set(call.name, {
          success: true,
          value: result,
          latency,
          hedged,
        });

        return { name: call.name, value: result };
      } catch (error) {
        const latency = performance.now() - startTime;
        this.recordLatency(call.name, latency);

        results.set(call.name, {
          success: false,
          error: error as Error,
          latency,
          hedged,
        });

        if (call.required) {
          throw error; // Re-throw for required dependencies
        }

        return { name: call.name, value: undefined };
      }
    });

    // Wait for all calls, even if some fail
    const settled = await Promise.allSettled(promises);

    const data: Partial<T> = {};
    settled.forEach((result, index) => {
      if (result.status === 'fulfilled' && result.value.value !== undefined) {
        data[calls[index].name as keyof T] = result.value.value;
      }
    });

    return { data, results };
  }

  /**
   * Hedged request: Issue duplicate request after delay, take first response
   */
  private async hedgedRequest<T>(
    fn: () => Promise<T>,
    hedgeDelay: number
  ): Promise<T> {
    const controllers: AbortController[] = [];
    const createCall = () => {
      const controller = new AbortController();
      controllers.push(controller);
      return fn();
    };

    const primaryPromise = createCall();
    const hedgedPromise = this.sleep(hedgeDelay).then(() => createCall());

    try {
      const result = await Promise.race([primaryPromise, hedgedPromise]);
      // Cancel any pending requests
      controllers.forEach((c) => c.abort());
      return result;
    } catch (error) {
      controllers.forEach((c) => c.abort());
      throw error;
    }
  }

  /**
   * Execute promise with timeout
   */
  private async withTimeout<T>(
    promise: Promise<T>,
    timeout: number,
    serviceName: string
  ): Promise<T> {
    const timeoutPromise = this.sleep(timeout).then(() => {
      throw new Error(`Timeout after ${timeout}ms for ${serviceName}`);
    });

    return Promise.race([promise, timeoutPromise]);
  }

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  /**
   * Record latency for percentile tracking
   */
  private recordLatency(serviceName: string, latency: number): void {
    if (!this.metrics.has(serviceName)) {
      this.metrics.set(serviceName, []);
    }

    const latencies = this.metrics.get(serviceName)!;
    latencies.push(latency);

    // Keep only recent samples
    if (latencies.length > this.maxMetricsSize) {
      latencies.shift();
    }
  }

  /**
   * Get percentile statistics for a service
   */
  getPercentiles(serviceName: string): {
    p50: number;
    p95: number;
    p99: number;
    p999: number;
  } | null {
    const latencies = this.metrics.get(serviceName);
    if (!latencies || latencies.length === 0) {
      return null;
    }

    const sorted = [...latencies].sort((a, b) => a - b);
    const getPercentile = (p: number) => {
      const index = Math.ceil((p / 100) * sorted.length) - 1;
      return sorted[Math.max(0, index)];
    };

    return {
      p50: getPercentile(50),
      p95: getPercentile(95),
      p99: getPercentile(99),
      p999: getPercentile(99.9),
    };
  }
}

// Example usage
const optimizer = new TailLatencyOptimizer();

interface Dashboard {
  profile?: UserProfile;
  orders?: Order[];
  recommendations?: Recommendation[];
  notifications?: Notification[];
  credits?: Credits;
}

async function getDashboard(userId: string): Promise<Dashboard> {
  const { data, results } = await optimizer.executeFanOut<Dashboard>([
    {
      name: 'profile',
      fn: () => userService.getProfile(userId),
      timeout: 5000,
      required: true,
      hedgeDelay: 50, // Issue hedge after 50ms
    },
    {
      name: 'orders',
      fn: () => orderService.getOrders(userId),
      timeout: 1000,
      required: false,
    },
    {
      name: 'recommendations',
      fn: () => recommendationService.get(userId),
      timeout: 300,
      required: false,
    },
    {
      name: 'notifications',
      fn: () => notificationService.get(userId),
      timeout: 300,
      required: false,
    },
    {
      name: 'credits',
      fn: () => creditService.getCredits(userId),
      timeout: 500,
      required: false,
    },
  ]);

  // Log tail latency for monitoring
  results.forEach((result, name) => {
    if (!result.success) {
      console.warn(`Service ${name} failed:`, result.error?.message);
    }
    console.log(`Service ${name} latency: ${result.latency.toFixed(2)}ms`);
  });

  return data;
}

// Periodically log percentiles
setInterval(() => {
  ['profile', 'orders', 'recommendations'].forEach((service) => {
    const percentiles = optimizer.getPercentiles(service);
    if (percentiles) {
      console.log(`${service} percentiles:`, percentiles);
    }
  });
}, 60000); // Every minute
```

## 6. Correct Usage Patterns (Where This Shines)

### Pattern 1: User-Facing API Gateway

When building a BFF (Backend-for-Frontend) or API gateway that fans out to multiple microservices:

```typescript
app.get('/api/homepage', async (req, res) => {
  const userId = req.user.id;
  
  const { data } = await optimizer.executeFanOut([
    { name: 'hero', fn: () => getHeroContent(), timeout: 200, required: true },
    { name: 'feed', fn: () => getFeed(userId), timeout: 500, required: true },
    { name: 'ads', fn: () => getAds(userId), timeout: 150, required: false },
    { name: 'recommendations', fn: () => getRecs(userId), timeout: 200, required: false },
  ]);

  res.json({
    hero: data.hero,
    feed: data.feed,
    ads: data.ads || null,
    recommendations: data.recommendations || [],
  });
});
```

**Why this works:**
- Hero and feed are required (page unusable without them)
- Ads and recommendations degrade gracefully
- Aggressive timeouts (200-500ms) prevent tail latency from dominating
- User sees page load in <500ms even if ads service is slow

### Pattern 2: Batch Processing with Partial Failures

When processing a batch of items where individual failures shouldn't block the entire batch:

```typescript
async function processBatch(userIds: string[]): Promise<void> {
  const batchSize = 100;
  
  for (let i = 0; i < userIds.length; i += batchSize) {
    const batch = userIds.slice(i, i + batchSize);
    
    const { results } = await optimizer.executeFanOut(
      batch.map(userId => ({
        name: userId,
        fn: () => processUser(userId),
        timeout: 2000,
        required: false,
      }))
    );

    // Log failures for retry queue
    results.forEach((result, userId) => {
      if (!result.success) {
        failureQueue.push({ userId, error: result.error });
      }
    });
  }
}
```

### Pattern 3: Critical Path with Hedging

For authentication or payment processing where high reliability is critical:

```typescript
async function processPayment(paymentData: Payment): Promise<PaymentResult> {
  const { data } = await optimizer.executeFanOut([
    {
      name: 'payment',
      fn: () => paymentGateway.charge(paymentData),
      timeout: 10000,
      required: true,
      hedgeDelay: 100, // Issue duplicate after 100ms
    },
  ]);

  return data.payment!;
}
```

**Why this works:**
- Payment is critical (cannot degrade)
- Hedging reduces p99 latency by issuing duplicate request
- If primary request is slow (> 100ms), hedge request likely succeeds faster
- First successful response wins, second is cancelled

## 7. Failure Modes & Edge Cases

### Race Condition: Duplicate Hedged Requests

**Problem:** Both hedged requests succeed, causing duplicate side effects (double charge, double insert).

**Mitigation:**
- Use idempotency tokens for all hedged requests
- Ensure backend can handle duplicate requests safely
- Cancel losing request immediately upon first success

```typescript
async function hedgedRequestWithIdempotency<T>(
  fn: (idempotencyKey: string) => Promise<T>,
  hedgeDelay: number
): Promise<T> {
  const idempotencyKey = crypto.randomUUID();
  
  return this.hedgedRequest(
    () => fn(idempotencyKey),
    hedgeDelay
  );
}
```

### Timeout Cascade

**Problem:** Service A times out calling Service B, retries, causing Service B to receive double traffic.

**Symptoms:**
- Service B sees traffic spike after latency increase
- Retry storm amplifies the original problem
- System oscillates between slow and fast states

**Mitigation:**
- Use exponential backoff for retries
- Implement circuit breakers (trip after N% failures)
- Add jitter to retries to avoid synchronized retry storms
- Track retry count in headers to detect loops

### Memory Leak from Incomplete Promises

**Problem:** Cancelled hedged requests may not clean up resources (DB connections, file handles).

**Symptoms:**
- File descriptor exhaustion
- Connection pool exhaustion
- Memory growth over time

**Mitigation:**
- Implement proper cancellation using AbortController
- Ensure all promises clean up in finally blocks
- Monitor open connections/handles

```typescript
async function withCleanup<T>(
  fn: (signal: AbortSignal) => Promise<T>,
  controller: AbortController
): Promise<T> {
  try {
    return await fn(controller.signal);
  } finally {
    controller.abort(); // Ensure cleanup even on success
  }
}
```

### Percentile Collapse Amplification

**Problem:** Adding more services to fan-out exponentially degrades tail latency.

**When it happens:**
- System grows from 5 services to 20 services
- Each new service adds 1% probability of slow request
- Aggregate p95 becomes individual p99.9

**Mitigation:**
- Limit fan-out depth (avoid calling 50 services for one user request)
- Batch requests where possible (GraphQL data loader pattern)
- Cache aggressively at the gateway layer
- Make more calls optional with degraded UX

### Timeout Too Aggressive

**Problem:** Timeout shorter than p99 causes false positive failures.

**Symptoms:**
- 1-5% of requests fail with timeout errors
- Underlying service actually succeeded (write committed)
- Users see errors but data was saved

**Mitigation:**
- Set timeout at 2-3x p99, not 1x p99
- Use hedged requests instead of aggressive timeouts for critical paths
- Implement at-least-once retry with idempotency

## 8. Performance Characteristics & Tradeoffs

### Latency Impact

Without tail latency optimization:
- p50: 50ms (primary service latency)
- p95: 200ms (individual service p99)
- p99: 1000ms (individual service p99.9)

With optimization (hedging + timeouts):
- p50: 50ms (unchanged)
- p95: 100ms (hedged requests cut tail)
- p99: 300ms (aggressive timeouts prevent long waits)

### Throughput Impact

**Cost:** Hedged requests double traffic for critical paths
- 10% of requests issue hedge = 10% more backend calls
- Backend must handle 110% normal load
- Increases infrastructure cost by ~10%

**Benefit:** Reduces user-perceived latency by 50-80% at p99
- Higher user satisfaction
- More completed transactions
- Better conversion rates

### Memory Impact

**Tracking percentiles:**
- Storing 1000 latency samples per service
- 5 services × 1000 samples × 8 bytes = 40KB
- Negligible memory overhead

**In-flight hedged requests:**
- Each hedged request doubles in-flight count temporarily
- If 10% are hedged, in-flight count increases by ~5%
- Requires slightly larger connection pools

### CPU Impact

**Minimal:**
- Percentile calculation: O(n log n) once per minute
- For 1000 samples: ~10μs per service
- Timeout management: O(1) per request

## 9. Foot-Guns & Common Mistakes (DO NOT SKIP)

### Mistake 1: Hedging All Requests

**Why engineers do it:** "If hedging helps tail latency, let's hedge everything!"

**What breaks:** Backend receives 2x traffic, overloads, latency increases, more hedging, death spiral.

**Detection:** Backend traffic is 2x frontend traffic; both requests succeed frequently.

**Fix:** Hedge only the slowest 5-10% of requests (those exceeding p95 latency).

```typescript
// Correct: Hedge only if primary is slow
const PRIMARY_TIMEOUT = 50; // p95 latency

async function smartHedge<T>(fn: () => Promise<T>): Promise<T> {
  const primary = fn();
  const hedge = sleep(PRIMARY_TIMEOUT).then(() => fn());
  return Promise.race([primary, hedge]);
}
```

### Mistake 2: Not Cancelling Losing Request

**Why engineers do it:** "Promise.race automatically cancels the loser."

**What breaks:** It doesn't. Both requests continue, doubling backend load and holding connections.

**Detection:** Connection pool exhaustion; backend sees twice as many completed requests as frontend.

**Fix:** Explicitly cancel using AbortController.

```typescript
async function properHedge<T>(fn: (signal: AbortSignal) => Promise<T>): Promise<T> {
  const controller1 = new AbortController();
  const controller2 = new AbortController();

  const promise1 = fn(controller1.signal).finally(() => controller2.abort());
  const promise2 = sleep(50).then(() => fn(controller2.signal).finally(() => controller1.abort()));

  return Promise.race([promise1, promise2]);
}
```

### Mistake 3: Using Same Timeout for All Services

**Why engineers do it:** Simplicity, uniformity.

**What breaks:** Critical services fail unnecessarily (timeout too aggressive) or non-critical services block too long (timeout too permissive).

**Detection:** Either high timeout error rates or slow tail latency despite timeouts.

**Fix:** Per-service timeout based on percentile data.

```typescript
const timeouts = {
  'user-service': 2000,      // Critical, DB-heavy
  'recommendation-service': 300,  // Optional, ML inference
  'analytics-service': 150,  // Optional, fire-and-forget
};
```

### Mistake 4: Ignoring Retry Amplification

**Why engineers do it:** "We need retries for reliability."

**What breaks:** Retries double traffic during slowdown, making it worse. One timeout + one retry = 2x load during incident.

**Detection:** Backend traffic spikes when latency increases; p99 latency includes (timeout + retry latency).

**Fix:** Circuit breaker to stop retries during overload; exponential backoff.

```typescript
class CircuitBreaker {
  private failures = 0;
  private readonly threshold = 5;
  private state: 'closed' | 'open' = 'closed';

  async execute<T>(fn: () => Promise<T>): Promise<T> {
    if (this.state === 'open') {
      throw new Error('Circuit breaker open');
    }

    try {
      const result = await fn();
      this.failures = 0;
      return result;
    } catch (error) {
      this.failures++;
      if (this.failures >= this.threshold) {
        this.state = 'open';
        setTimeout(() => { this.state = 'closed'; this.failures = 0; }, 30000);
      }
      throw error;
    }
  }
}
```

### Mistake 5: Optimizing Average Instead of Percentiles

**Why engineers do it:** Averages are easier to understand and calculate.

**What breaks:** System looks healthy but 1-5% of users have terrible experience.

**Detection:** User complaints don't correlate with average latency metrics.

**Fix:** Track and alert on p95, p99, p99.9 latency.

```typescript
// Wrong
const avgLatency = latencies.reduce((a, b) => a + b) / latencies.length;
if (avgLatency > 100) alert();

// Correct
const sorted = latencies.sort((a, b) => a - b);
const p99 = sorted[Math.floor(sorted.length * 0.99)];
if (p99 > 500) alert();
```

## 10. When NOT to Use This (Anti-Patterns)

### Anti-Pattern 1: Single Service Call (No Fan-Out)

If you're calling one service, tail latency optimization adds complexity without benefit:

```typescript
// Don't do this for single calls
const result = await optimizer.executeFanOut([
  { name: 'user', fn: () => getUser(id), timeout: 1000, required: true },
]);

// Just use a plain timeout
const result = await withTimeout(getUser(id), 1000);
```

### Anti-Pattern 2: Background Jobs

For async background processing where latency doesn't matter:

```typescript
// Don't use hedging or aggressive timeouts here
async function processEmailQueue() {
  while (true) {
    const email = await queue.pop();
    await sendEmail(email); // No timeout needed, retry is fine
  }
}
```

### Anti-Pattern 3: Batch APIs

When client explicitly batches requests (GraphQL, gRPC streaming), let the client handle parallelism:

```typescript
// Don't optimize tail latency in batch endpoints
app.post('/api/batch', async (req, res) => {
  const results = await Promise.all(
    req.body.requests.map(r => processRequest(r))
  );
  res.json(results);
});
```

The client controls parallelism here; server should not add hedging.

### Anti-Pattern 4: Write Operations

Never hedge write operations (creates duplicates):

```typescript
// NEVER do this
const result = await hedgedRequest(() => 
  database.insert({ userId, amount: 100 })
);
```

Hedging is for idempotent reads only.

## 11. Related Concepts (With Contrast)

### Request Coalescing (Singleflight)

**Difference:** Request coalescing deduplicates identical concurrent requests. Tail latency optimization handles fan-out to multiple different services.

**When to combine:** Use both. Coalesce identical requests to reduce backend load, then apply tail latency optimization to the fan-out of unique requests.

### Bulkheads & Resource Isolation

**Difference:** Bulkheads prevent one slow service from consuming all resources (connection pool, threads). Tail latency optimization reduces the probability of hitting slow services.

**When to combine:** Use both. Bulkheads limit blast radius; tail latency optimization reduces likelihood of hitting the problem.

### Circuit Breaker

**Difference:** Circuit breaker stops calling failing services. Tail latency optimization assumes services are working but occasionally slow.

**When to combine:** Use circuit breakers for persistent failures (service down). Use tail latency optimization for transient slowness (GC pause, slow query).

### Load Shedding

**Difference:** Load shedding drops requests when backend is overloaded. Tail latency optimization handles slowness even when backend is not overloaded.

**When to combine:** Use both. Load shedding prevents collapse under extreme load; tail latency optimization improves user experience during normal operation.

## 12. Production Readiness Checklist

Before deploying tail latency optimizations to production:

### Metrics to Monitor

- [ ] Per-service p50, p95, p99, p99.9 latency
- [ ] Aggregate endpoint p50, p95, p99, p99.9 latency
- [ ] Hedged request rate (should be <10%)
- [ ] Timeout error rate per service
- [ ] Required vs optional dependency success rates
- [ ] Backend traffic increase from hedging (should be <10%)

### Logging Requirements

- [ ] Log all timeout events with service name and latency
- [ ] Log when hedged request wins vs primary
- [ ] Log partial failures (optional dependencies timing out)
- [ ] Include correlation IDs to track fan-out across services

### Timeouts and Limits

- [ ] Set per-service timeouts based on percentile data (2-3x p99)
- [ ] Configure hedge delay as ~p95 latency
- [ ] Implement circuit breakers with 5-10 failure threshold
- [ ] Set connection pool size to handle hedging overhead (110% normal)

### Load Testing Considerations

- [ ] Test with realistic fan-out (10-20 services)
- [ ] Inject artificial latency at p99 level into 1-2 services
- [ ] Verify aggregate latency improves with hedging
- [ ] Verify backend doesn't overload from hedging
- [ ] Test circuit breaker trips under sustained failure
- [ ] Validate partial degradation works (UI handles missing optional data)

### Rollout and Rollback Strategy

- [ ] Deploy to 1% of traffic, verify metrics improve
- [ ] Gradually increase to 10%, 50%, 100%
- [ ] Monitor backend traffic increase (should be <10%)
- [ ] Have feature flag to disable hedging instantly
- [ ] Have feature flag to relax timeouts if too aggressive
- [ ] Document expected p99 improvement (baseline vs optimized)

### Alerting

- [ ] Alert if aggregate p99 > 2x baseline
- [ ] Alert if hedged request rate > 20% (indicates systemic slowness)
- [ ] Alert if timeout error rate > 5%
- [ ] Alert if required dependency fails > 0.1%
- [ ] Alert if backend traffic increases > 15% (hedge leak)
