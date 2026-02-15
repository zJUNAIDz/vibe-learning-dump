package final

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestRateLimiter_Basic tests basic allow functionality.
func TestRateLimiter_Basic(t *testing.T) {
	rl := NewRateLimiter(Config{
		DefaultRate:     10, // 10 req/sec
		DefaultBurst:    10,
		NumShards:       16,
		CleanupInterval: 1 * time.Hour,
		InactivityTTL:   1 * time.Hour,
	})
	defer rl.Close()

	clientID := "test-client"

	// Should allow initial burst
	allowed := 0
	for i := 0; i < 10; i++ {
		if rl.Allow(clientID) {
			allowed++
		}
	}

	if allowed < 9 { // Allow some timing variance
		t.Errorf("Expected ~10 allowed, got %d", allowed)
	}

	// Next request should be denied (burst depleted)
	if rl.Allow(clientID) {
		t.Error("Expected denial after burst depleted")
	}
}

// TestRateLimiter_SetLimit tests custom rate configuration.
func TestRateLimiter_SetLimit(t *testing.T) {
	rl := NewRateLimiter(Config{
		DefaultRate:     10,
		DefaultBurst:    10,
		NumShards:       16,
		CleanupInterval: 1 * time.Hour,
		InactivityTTL:   1 * time.Hour,
	})
	defer rl.Close()

	clientID := "vip-client"

	// Set higher limit for VIP
	rl.SetLimit(clientID, 100, 100)

	// Should allow burst of 100
	allowed := 0
	for i := 0; i < 100; i++ {
		if rl.Allow(clientID) {
			allowed++
		}
	}

	if allowed < 95 { // Allow some timing variance
		t.Errorf("Expected ~100 allowed for VIP, got %d", allowed)
	}
}

// TestRateLimiter_RateAccuracy tests rate limiting accuracy over time.
func TestRateLimiter_RateAccuracy(t *testing.T) {
	rl := NewRateLimiter(Config{
		DefaultRate:     50, // 50 req/sec
		DefaultBurst:    50,
		NumShards:       16,
		CleanupInterval: 1 * time.Hour,
		InactivityTTL:   1 * time.Hour,
	})
	defer rl.Close()

	clientID := "rate-test"

	// Exhaust initial burst
	for i := 0; i < 50; i++ {
		rl.Allow(clientID)
	}

	// Wait 1 second for refill
	time.Sleep(1 * time.Second)

	// Should allow ~50 more requests
	allowed := 0
	for i := 0; i < 60; i++ {
		if rl.Allow(clientID) {
			allowed++
		}
	}

	// Allow ±10% variance
	if allowed < 45 || allowed > 55 {
		t.Errorf("Expected ~50 allowed in 1 second, got %d (rate accuracy: %.0f%%)",
			allowed, float64(allowed)/50.0*100)
	} else {
		t.Logf("Rate accuracy: %.0f%% (%d/50)", float64(allowed)/50.0*100, allowed)
	}
}

// TestRateLimiter_Cleanup tests inactive client cleanup.
func TestRateLimiter_Cleanup(t *testing.T) {
	rl := NewRateLimiter(Config{
		DefaultRate:     10,
		DefaultBurst:    10,
		NumShards:       4,
		CleanupInterval: 100 * time.Millisecond,
		InactivityTTL:   200 * time.Millisecond,
	})
	defer rl.Close()

	// Create some clients
	for i := 0; i < 10; i++ {
		rl.Allow(fmt.Sprintf("client-%d", i))
	}

	initialCount := rl.GetClientCount()
	if initialCount < 10 {
		t.Errorf("Expected 10 clients, got %d", initialCount)
	}

	// Wait for cleanup (TTL + cleanup interval)
	time.Sleep(400 * time.Millisecond)

	// Clients should be cleaned up
	finalCount := rl.GetClientCount()
	if finalCount > 1 {
		t.Errorf("Expected clients cleaned up, still have %d", finalCount)
	} else {
		t.Logf("Cleanup successful: %d → %d clients", initialCount, finalCount)
	}
}

// TestRateLimiter_Concurrent tests concurrent access with race detector.
func TestRateLimiter_Concurrent(t *testing.T) {
	rl := NewRateLimiter(Config{
		DefaultRate:     100,
		DefaultBurst:    100,
		NumShards:       64,
		CleanupInterval: 1 * time.Hour,
		InactivityTTL:   1 * time.Hour,
	})
	defer rl.Close()

	var wg sync.WaitGroup
	var totalAllowed atomic.Uint64

	// 100 goroutines × 1000 ops = 100,000 total operations
	numGoroutines := 100
	opsPerGoroutine := 1000

	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			clientID := fmt.Sprintf("client-%d", id)

			for j := 0; j < opsPerGoroutine; j++ {
				if rl.Allow(clientID) {
					totalAllowed.Add(1)
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	allowed := totalAllowed.Load()
	throughput := float64(numGoroutines*opsPerGoroutine) / elapsed.Seconds()

	t.Logf("Concurrent test: %d goroutines × %d ops = %d total",
		numGoroutines, opsPerGoroutine, numGoroutines*opsPerGoroutine)
	t.Logf("Allowed: %d, Denied: %d", allowed, numGoroutines*opsPerGoroutine-int(allowed))
	t.Logf("Elapsed: %v, Throughput: %.0f req/sec", elapsed, throughput)

	// Should achieve high throughput (>100k req/sec typical)
	if throughput < 50000 {
		t.Errorf("Throughput too low: %.0f req/sec (expected >50k)", throughput)
	}
}

// TestRateLimiter_Metrics tests metrics collection.
func TestRateLimiter_Metrics(t *testing.T) {
	rl := NewRateLimiter(Config{
		DefaultRate:     10,
		DefaultBurst:    10,
		NumShards:       16,
		CleanupInterval: 1 * time.Hour,
		InactivityTTL:   1 * time.Hour,
	})
	defer rl.Close()

	clientID := "metrics-test"

	// Make some requests
	allowed := 0
	denied := 0
	for i := 0; i < 20; i++ {
		if rl.Allow(clientID) {
			allowed++
		} else {
			denied++
		}
	}

	// Check metrics
	m := rl.Metrics()
	if m.Allowed != uint64(allowed) {
		t.Errorf("Metrics: allowed mismatch, expected %d, got %d", allowed, m.Allowed)
	}
	if m.Denied != uint64(denied) {
		t.Errorf("Metrics: denied mismatch, expected %d, got %d", denied, m.Denied)
	}

	t.Logf("Metrics: Allowed=%d, Denied=%d, ActiveClients=%d",
		m.Allowed, m.Denied, m.ActiveClients)
}

// TestRateLimiter_Sharding tests that sharding reduces contention.
func TestRateLimiter_Sharding(t *testing.T) {
	// Test with 1 shard (high contention)
	rl1 := NewRateLimiter(Config{
		DefaultRate:     1000,
		DefaultBurst:    1000,
		NumShards:       1, // Single shard
		CleanupInterval: 1 * time.Hour,
		InactivityTTL:   1 * time.Hour,
	})
	defer rl1.Close()

	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10000; j++ {
				rl1.Allow(fmt.Sprintf("client-%d", id))
			}
		}(i)
	}
	wg.Wait()
	throughput1 := float64(100000) / time.Since(start).Seconds()

	// Test with 256 shards (low contention)
	rl256 := NewRateLimiter(Config{
		DefaultRate:     1000,
		DefaultBurst:    1000,
		NumShards:       256, // Many shards
		CleanupInterval: 1 * time.Hour,
		InactivityTTL:   1 * time.Hour,
	})
	defer rl256.Close()

	start = time.Now()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10000; j++ {
				rl256.Allow(fmt.Sprintf("client-%d", id))
			}
		}(i)
	}
	wg.Wait()
	throughput256 := float64(100000) / time.Since(start).Seconds()

	improvement := throughput256 / throughput1
	t.Logf("1 shard:   %.0f req/sec", throughput1)
	t.Logf("256 shards: %.0f req/sec", throughput256)
	t.Logf("Improvement: %.1fx", improvement)

	// Sharding should provide significant improvement
	if improvement < 2.0 {
		t.Errorf("Expected >2x improvement from sharding, got %.1fx", improvement)
	}
}

// TestRateLimiter_GetTokens tests token balance inspection.
func TestRateLimiter_GetTokens(t *testing.T) {
	rl := NewRateLimiter(Config{
		DefaultRate:     10,
		DefaultBurst:    10,
		NumShards:       16,
		CleanupInterval: 1 * time.Hour,
		InactivityTTL:   1 * time.Hour,
	})
	defer rl.Close()

	clientID := "token-test"

	// Initial tokens should be equal to burst
	tokens := rl.GetTokens(clientID)
	if tokens < 9.0 || tokens > 11.0 { // Allow timing variance
		t.Errorf("Expected ~10 initial tokens, got %.2f", tokens)
	}

	// Consume some tokens
	for i := 0; i < 5; i++ {
		rl.Allow(clientID)
	}

	// Should have ~5 tokens left
	tokens = rl.GetTokens(clientID)
	if tokens < 4.0 || tokens > 6.0 {
		t.Errorf("Expected ~5 tokens after consuming 5, got %.2f", tokens)
	}

	t.Logf("Token balance: %.2f", tokens)
}

// BenchmarkRateLimiter_Allow benchmarks Allow() throughput.
func BenchmarkRateLimiter_Allow(b *testing.B) {
	rl := NewRateLimiter(DefaultConfig())
	defer rl.Close()

	clientID := "bench-client"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow(clientID)
	}
}

// BenchmarkRateLimiter_AllowParallel benchmarks concurrent Allow().
func BenchmarkRateLimiter_AllowParallel(b *testing.B) {
	rl := NewRateLimiter(DefaultConfig())
	defer rl.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		id := 0
		for pb.Next() {
			clientID := fmt.Sprintf("client-%d", id%100) // 100 different clients
			rl.Allow(clientID)
			id++
		}
	})
}

// BenchmarkRateLimiter_Sharding compares different shard counts.
func BenchmarkRateLimiter_Sharding(b *testing.B) {
	shardCounts := []int{1, 16, 64, 256}

	for _, numShards := range shardCounts {
		b.Run(fmt.Sprintf("shards=%d", numShards), func(b *testing.B) {
			rl := NewRateLimiter(Config{
				DefaultRate:     1000,
				DefaultBurst:    1000,
				NumShards:       numShards,
				CleanupInterval: 1 * time.Hour,
				InactivityTTL:   1 * time.Hour,
			})
			defer rl.Close()

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				id := 0
				for pb.Next() {
					clientID := fmt.Sprintf("client-%d", id%10)
					rl.Allow(clientID)
					id++
				}
			})
		})
	}
}

// Example demonstrating basic usage.
func ExampleRateLimiter() {
	rl := NewRateLimiter(DefaultConfig())
	defer rl.Close()

	clientID := "user-123"

	// Set custom limit
	rl.SetLimit(clientID, 100, 200) // 100 req/sec, burst 200

	// Check rate limit
	if rl.Allow(clientID) {
		fmt.Println("Request allowed")
	} else {
		fmt.Println("Rate limit exceeded")
	}

	// Check metrics
	m := rl.Metrics()
	fmt.Printf("Total: %d allowed, %d denied\n", m.Allowed, m.Denied)

	// Output:
	// Request allowed
	// Total: 1 allowed, 0 denied
}
