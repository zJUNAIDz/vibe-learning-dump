package final

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestBroker_Basic tests basic pub-sub functionality.
func TestBroker_Basic(t *testing.T) {
	broker := NewBroker(Config{BufferSize: 10})
	defer broker.Close()

	ctx := context.Background()
	var received atomic.Int32

	// Subscribe
	id, err := broker.Subscribe(ctx, "test", func(ctx context.Context, msg *Message) error {
		received.Add(1)
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish
	msgID, err := broker.Publish("test", "hello", nil)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	// Wait for delivery
	time.Sleep(50 * time.Millisecond)

	if received.Load() != 1 {
		t.Errorf("Expected 1 message, got %d", received.Load())
	}

	// Ack message
	if err := broker.Ack(id, msgID); err != nil {
		t.Errorf("Failed to ack: %v", err)
	}

	// Check metrics
	metrics := broker.Metrics()
	if metrics.Published != 1 {
		t.Errorf("Expected 1 published, got %d", metrics.Published)
	}
}

// TestBroker_MultipleSubscribers tests fan-out to multiple subscribers.
func TestBroker_MultipleSubscribers(t *testing.T) {
	broker := NewBroker(Config{BufferSize: 10})
	defer broker.Close()

	ctx := context.Background()
	var count1, count2, count3 atomic.Int32

	// Subscribe 3 times to same topic
	broker.Subscribe(ctx, "topic", func(ctx context.Context, msg *Message) error {
		count1.Add(1)
		return nil
	})

	broker.Subscribe(ctx, "topic", func(ctx context.Context, msg *Message) error {
		count2.Add(1)
		return nil
	})

	broker.Subscribe(ctx, "topic", func(ctx context.Context, msg *Message) error {
		count3.Add(1)
		return nil
	})

	// Publish once
	broker.Publish("topic", "message", nil)

	// Wait for delivery
	time.Sleep(50 * time.Millisecond)

	// All 3 should receive
	if count1.Load() != 1 || count2.Load() != 1 || count3.Load() != 1 {
		t.Errorf("Expected all subscribers to receive message: %d, %d, %d",
			count1.Load(), count2.Load(), count3.Load())
	}
}

// TestBroker_PatternMatching tests topic pattern matching.
func TestBroker_PatternMatching(t *testing.T) {
	broker := NewBroker(Config{BufferSize: 10})
	defer broker.Close()

	ctx := context.Background()
	var orders, events, all atomic.Int32

	// Subscribe with patterns
	broker.Subscribe(ctx, "orders.*", func(ctx context.Context, msg *Message) error {
		orders.Add(1)
		return nil
	})

	broker.Subscribe(ctx, "events.*", func(ctx context.Context, msg *Message) error {
		events.Add(1)
		return nil
	})

	broker.Subscribe(ctx, "*", func(ctx context.Context, msg *Message) error {
		all.Add(1)
		return nil
	})

	// Publish to different topics
	broker.Publish("orders.created", "order1", nil)
	broker.Publish("orders.updated", "order2", nil)
	broker.Publish("events.click", "click1", nil)
	broker.Publish("other.topic", "other", nil)

	time.Sleep(100 * time.Millisecond)

	// Check counts
	if orders.Load() != 2 {
		t.Errorf("Expected 2 orders messages, got %d", orders.Load())
	}
	if events.Load() != 1 {
		t.Errorf("Expected 1 events message, got %d", events.Load())
	}
	if all.Load() != 4 {
		t.Errorf("Expected 4 all messages, got %d", all.Load())
	}
}

// TestBroker_Retry tests retry logic on handler failure.
func TestBroker_Retry(t *testing.T) {
	broker := NewBroker(Config{
		BufferSize:   10,
		MaxRetries:   3,
		RetryBackoff: 50 * time.Millisecond,
	})
	defer broker.Close()

	ctx := context.Background()
	var attempts atomic.Int32

	// Handler that fails first 2 times
	broker.Subscribe(ctx, "test", func(ctx context.Context, msg *Message) error {
		count := attempts.Add(1)
		if count < 3 {
			return errors.New("temporary failure")
		}
		return nil
	})

	// Publish
	broker.Publish("test", "message", nil)

	// Wait for retries
	time.Sleep(500 * time.Millisecond)

	// Should have tried 3 times (initial + 2 retries)
	if attempts.Load() != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts.Load())
	}

	// Check metrics
	metrics := broker.Metrics()
	if metrics.Retries != 2 {
		t.Errorf("Expected 2 retries, got %d", metrics.Retries)
	}
}

// TestBroker_DLQ tests dead letter queue for failed messages.
func TestBroker_DLQ(t *testing.T) {
	broker := NewBroker(Config{
		BufferSize:   10,
		MaxRetries:   2,
		RetryBackoff: 10 * time.Millisecond,
	})
	defer broker.Close()

	ctx := context.Background()

	// Handler that always fails
	broker.Subscribe(ctx, "test", func(ctx context.Context, msg *Message) error {
		return errors.New("always fails")
	})

	// Publish
	broker.Publish("test", "bad-message", nil)

	// Wait for max retries
	time.Sleep(200 * time.Millisecond)

	// Check DLQ
	dlq := broker.GetDLQ()
	if len(dlq) != 1 {
		t.Errorf("Expected 1 message in DLQ, got %d", len(dlq))
	}

	if len(dlq) > 0 {
		if dlq[0].Payload != "bad-message" {
			t.Errorf("Wrong message in DLQ: %v", dlq[0].Payload)
		}
		if dlq[0].Metadata["dlq_reason"] == "" {
			t.Error("DLQ message missing reason")
		}
	}

	// Check metrics
	metrics := broker.Metrics()
	if metrics.Failed != 1 {
		t.Errorf("Expected 1 failed, got %d", metrics.Failed)
	}
	if metrics.DLQSize != 1 {
		t.Errorf("Expected DLQ size 1, got %d", metrics.DLQSize)
	}
}

// TestBroker_CircuitBreaker tests circuit breaker for failing subscribers.
func TestBroker_CircuitBreaker(t *testing.T) {
	broker := NewBroker(Config{BufferSize: 5})
	defer broker.Close()

	ctx := context.Background()
	var attempts atomic.Int32

	// Slow handler that will cause buffer to fill
	broker.Subscribe(ctx, "test", func(ctx context.Context, msg *Message) error {
		attempts.Add(1)
		time.Sleep(200 * time.Millisecond)
		return nil
	})

	// Publish many messages quickly (will overflow buffer)
	for i := 0; i < 20; i++ {
		broker.Publish("test", fmt.Sprintf("msg-%d", i), nil)
	}

	// Wait
	time.Sleep(100 * time.Millisecond)

	// Circuit should have tripped
	metrics := broker.Metrics()
	if metrics.CircuitTrips == 0 {
		t.Error("Expected circuit breaker to trip")
	}

	t.Logf("Circuit breaker trips: %d", metrics.CircuitTrips)
}

// TestBroker_RateLimit tests per-subscriber rate limiting.
func TestBroker_RateLimit(t *testing.T) {
	broker := NewBroker(Config{BufferSize: 100})
	defer broker.Close()

	ctx := context.Background()
	var received atomic.Int32
	var timestamps []time.Time
	var mu sync.Mutex

	// Subscribe with rate limit
	broker.Subscribe(ctx, "test", func(ctx context.Context, msg *Message) error {
		mu.Lock()
		timestamps = append(timestamps, time.Now())
		mu.Unlock()
		received.Add(1)
		return nil
	}, WithRateLimit(100*time.Millisecond))

	// Publish 5 messages quickly
	start := time.Now()
	for i := 0; i < 5; i++ {
		broker.Publish("test", i, nil)
	}

	// Wait for processing
	time.Sleep(600 * time.Millisecond)

	// Should have rate limited to ~5 msgs in 500ms
	elapsed := time.Since(start)
	count := received.Load()

	if count < 3 || count > 6 {
		t.Errorf("Expected 3-6 messages with rate limit, got %d in %v", count, elapsed)
	}

	// Check timing between messages
	mu.Lock()
	for i := 1; i < len(timestamps); i++ {
		gap := timestamps[i].Sub(timestamps[i-1])
		if gap < 80*time.Millisecond { // Allow some slack
			t.Errorf("Messages too close: %v (expected ≥100ms)", gap)
		}
	}
	mu.Unlock()

	t.Logf("Received %d messages with rate limiting", count)
}

// TestBroker_Concurrent tests concurrent publishing and subscribing.
func TestBroker_Concurrent(t *testing.T) {
	broker := NewBroker(Config{BufferSize: 1000})
	defer broker.Close()

	ctx := context.Background()
	var received atomic.Int64

	// Start multiple subscribers
	for i := 0; i < 10; i++ {
		broker.Subscribe(ctx, "test", func(ctx context.Context, msg *Message) error {
			received.Add(1)
			return nil
		})
	}

	// Concurrent publishers
	var wg sync.WaitGroup
	publishCount := 100

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < publishCount; j++ {
				broker.Publish("test", fmt.Sprintf("msg-%d-%d", id, j), nil)
			}
		}(i)
	}

	wg.Wait()

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	// Should have received ~10,000 messages (10 publishers × 100 msgs × 10 subscribers)
	expectedMin := int64(publishCount * 10 * 10 * 8 / 10) // Allow 20% loss
	if received.Load() < expectedMin {
		t.Errorf("Expected at least %d messages, got %d", expectedMin, received.Load())
	}

	t.Logf("Received %d messages", received.Load())
}

// TestBroker_Unsubscribe tests subscription cleanup.
func TestBroker_Unsubscribe(t *testing.T) {
	broker := NewBroker(Config{BufferSize: 10})
	defer broker.Close()

	ctx := context.Background()
	var count atomic.Int32

	// Subscribe
	id, _ := broker.Subscribe(ctx, "test", func(ctx context.Context, msg *Message) error {
		count.Add(1)
		return nil
	})

	// Publish first message
	broker.Publish("test", "msg1", nil)
	time.Sleep(50 * time.Millisecond)

	if count.Load() != 1 {
		t.Errorf("Expected 1 message, got %d", count.Load())
	}

	// Unsubscribe
	if err := broker.Unsubscribe(id); err != nil {
		t.Fatalf("Failed to unsubscribe: %v", err)
	}

	// Publish second message
	broker.Publish("test", "msg2", nil)
	time.Sleep(50 * time.Millisecond)

	// Should still be 1 (not received msg2)
	if count.Load() != 1 {
		t.Errorf("Expected 1 message after unsubscribe, got %d", count.Load())
	}
}

// TestBroker_GracefulShutdown tests graceful broker shutdown.
func TestBroker_GracefulShutdown(t *testing.T) {
	broker := NewBroker(Config{BufferSize: 100})

	ctx := context.Background()
	var processed atomic.Int32

	// Subscribe with slow handler
	broker.Subscribe(ctx, "test", func(ctx context.Context, msg *Message) error {
		time.Sleep(50 * time.Millisecond)
		processed.Add(1)
		return nil
	})

	// Publish messages
	for i := 0; i < 10; i++ {
		broker.Publish("test", i, nil)
	}

	// Allow some processing
	time.Sleep(100 * time.Millisecond)

	// Close (should wait for in-flight)
	start := time.Now()
	broker.Close()
	elapsed := time.Since(start)

	count := processed.Load()
	t.Logf("Processed %d messages before shutdown in %v", count, elapsed)

	// Should have processed at least some messages
	if count == 0 {
		t.Error("Expected some messages to be processed before shutdown")
	}
}

// BenchmarkBroker_Publish benchmarks publishing performance.
func BenchmarkBroker_Publish(b *testing.B) {
	broker := NewBroker(Config{BufferSize: 1000})
	defer broker.Close()

	ctx := context.Background()
	broker.Subscribe(ctx, "test", func(ctx context.Context, msg *Message) error {
		return nil
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		broker.Publish("test", i, nil)
	}
}

// BenchmarkBroker_Parallel benchmarks parallel publishing.
func BenchmarkBroker_Parallel(b *testing.B) {
	broker := NewBroker(Config{BufferSize: 10000})
	defer broker.Close()

	ctx := context.Background()

	// Multiple subscribers
	for i := 0; i < 5; i++ {
		broker.Subscribe(ctx, "*", func(ctx context.Context, msg *Message) error {
			return nil
		})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			topic := fmt.Sprintf("topic-%d", i%10)
			broker.Publish(topic, i, nil)
			i++
		}
	})
}

// Example demonstrates full broker usage.
func ExampleBroker() {
	// Create broker
	broker := NewBroker(Config{
		BufferSize:   1000,
		MaxRetries:   3,
		RetryBackoff: time.Second,
		AckTimeout:   30 * time.Second,
	})
	defer broker.Close()

	ctx := context.Background()

	// Subscribe to orders
	orderID, _ := broker.Subscribe(ctx, "orders.*", func(ctx context.Context, msg *Message) error {
		fmt.Printf("Order: %v\n", msg.Payload)
		return nil
	})

	// Subscribe to all events
	broker.Subscribe(ctx, "*", func(ctx context.Context, msg *Message) error {
		fmt.Printf("Event: %s - %v\n", msg.Topic, msg.Payload)
		return nil
	})

	// Publish messages
	msgID, _ := broker.Publish("orders.created", map[string]interface{}{
		"id":     "order-123",
		"amount": 99.99,
	}, map[string]string{
		"user_id": "user-456",
	})

	broker.Publish("events.click", "button-clicked", nil)

	// Ack message
	broker.Ack(orderID, msgID)

	// Check metrics
	time.Sleep(100 * time.Millisecond)
	metrics := broker.Metrics()
	fmt.Printf("Published: %d, Delivered: %d, Acked: %d\n",
		metrics.Published, metrics.Delivered, metrics.Acked)
}
