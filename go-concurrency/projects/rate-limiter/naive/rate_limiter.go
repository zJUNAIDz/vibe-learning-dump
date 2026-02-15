package naive

import (
	"sync"
	"time"
)

// RateLimiter implements a naive token bucket rate limiter with intentional issues:
// 1. Global mutex causes severe contention with many clients
// 2. No cleanup leads to memory leak
// 3. Inefficient: calculates tokens on every request
// 4. Poor performance: single lock serializes everything
//
// Expected throughput: ~50k req/sec (plateaus due to lock contention)
type RateLimiter struct {
	mu      sync.Mutex // ❌ PROBLEM: Single global lock for ALL clients
	clients map[string]*bucket
}

// bucket holds token bucket state for one client
type bucket struct {
	rate       float64   // tokens per second
	burst      int64     // maximum tokens
	tokens     float64   // current available tokens
	lastRefill time.Time // last time tokens were added
}

// NewRateLimiter creates a naive rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*bucket),
		// ❌ PROBLEM: No cleanup goroutine, memory leaks!
	}
}

// SetLimit configures rate limit for a client
func (rl *RateLimiter) SetLimit(clientID string, rate float64, burst int64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.clients[clientID] = &bucket{
		rate:       rate,
		burst:      burst,
		tokens:     float64(burst), // Start with full burst
		lastRefill: time.Now(),
	}
}

// Allow checks if request is allowed (returns true) or rate limited (returns false)
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	// ❌ PROBLEM: Lock held during entire operation, even for reads

	b, exists := rl.clients[clientID]
	if !exists {
		// Default: deny unknown clients (should have called SetLimit first)
		return false
	}

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()

	// ❌ PROBLEM: Float arithmetic on every request is inefficient
	// Could precompute or use integer math
	newTokens := elapsed * b.rate
	b.tokens += newTokens

	// Cap at burst limit
	if b.tokens > float64(b.burst) {
		b.tokens = float64(b.burst)
	}

	b.lastRefill = now

	// Check if we have at least 1 token
	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true
	}

	return false
}

// GetClientCount returns number of tracked clients
// ❌ PROBLEM: This number only grows, never shrinks (memory leak)
func (rl *RateLimiter) GetClientCount() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.clients)
}

// GetTokens returns current token count for debugging
func (rl *RateLimiter) GetTokens(clientID string) float64 {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, exists := rl.clients[clientID]
	if !exists {
		return 0
	}
	return b.tokens
}

// ❌ MAJOR PROBLEMS SUMMARY:
//
// 1. GLOBAL LOCK CONTENTION
//    - Single mu.Lock() serializes ALL operations
//    - With N clients making concurrent requests, throughput is O(1/N)
//    - Benchmark will show throughput plateau around 50k req/sec
//
// 2. MEMORY LEAK
//    - clients map grows forever
//    - No cleanup of inactive clients
//    - Memory usage: StartMemory + (200 bytes * UniqueClients)
//    - After 1 million unique clients: ~200MB leaked
//
// 3. INEFFICIENT COMPUTATION
//    - time.Now() called twice per request
//    - Float math (elapsed * rate) on hot path
//    - No lazy evaluation (refills even if tokens available)
//
// 4. POOR SCALABILITY
//    - Does NOT benefit from multiple cores
//    - CPU utilization stays low (~10%) because goroutines wait on lock
//    - Latency increases linearly with concurrency
//
// 5. NO OBSERVABILITY
//    - No metrics (allowed vs denied requests)
//    - No visibility into rate limit status
//    - Hard to debug in production
//
// HOW TO OBSERVE THESE PROBLEMS:
//
// 1. Run stress test with 100 concurrent clients:
//    go test -run=TestStress -v
//    Expected: throughput plateaus, high lock contention
//
// 2. Run with race detector:
//    go test -race
//    Expected: should pass (but slow due to lock overhead)
//
// 3. Profile with pprof:
//    go test -bench=. -cpuprofile=cpu.prof
//    go tool pprof cpu.prof
//    (pprof) top
//    Expected: sync.(*Mutex).Lock shows high CPU time
//
// 4. Memory profile:
//    m1 := GetClientCount() // 1000 clients
//    ... use for 5 minutes ...
//    m2 := GetClientCount() // Still 1000 (should be 0 if inactive)
//    Expected: map never shrinks
//
// FIXES IN improved/rate_limiter.go:
// - Shard clients across 256 locks (reduces contention)
// - Background cleanup goroutine (removes inactive clients)
// - Lazy token refill (only when needed)
// - Integer math instead of float (faster)
