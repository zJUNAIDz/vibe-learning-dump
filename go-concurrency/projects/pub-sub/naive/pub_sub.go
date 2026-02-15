// Package naive implements a pub-sub system with INTENTIONAL PROBLEMS.
//
// This implementation demonstrates common mistakes when building message brokers:
// 1. Blocking Publish - Slow subscriber blocks publisher
// 2. Memory Leak - Subscribers never cleaned up
// 3. No Buffering - Messages dropped if subscriber busy
// 4. Race Condition - Concurrent map access without mutex
// 5. No Delivery Guarantee - Messages lost on subscriber crash
// 6. No Backpressure - Fast publisher overwhelms slow consumer
//
// Expected behavior:
// - Fast subscribers: Works (lucky!)
// - Slow subscriber: Blocks entire system
// - High load: Messages dropped, race conditions
//
// HOW TO OBSERVE THESE PROBLEMS:
//
//  1. Blocking Publish (slow consumer blocks everyone):
//     broker := NewBroker()
//     broker.Subscribe("topic", slowConsumer) // Takes 1s per message
//     broker.Publish("topic", msg)            // Blocks for 1s!
//     // All publishers blocked by slowest consumer
//
//  2. Memory Leak (subscribers never removed):
//     for i := 0; i < 1000; i++ {
//     broker.Subscribe("topic", consumer)
//     }
//     // 1000 subscribers registered forever
//     // Even if goroutines exit, subscribers stay in map
//
//  3. No Buffering (messages dropped):
//     broker.Subscribe("topic", func(msg) {
//     time.Sleep(100ms) // Slow consumer
//     })
//     for i := 0; i < 100; i++ {
//     broker.Publish("topic", i) // Most messages lost
//     }
//
//  4. Race Condition (concurrent map panic):
//     go broker.Subscribe("topic", consumer1)
//     go broker.Subscribe("topic", consumer2)
//     go broker.Publish("topic", msg)
//     // Result: "fatal error: concurrent map writes"
//
//  5. No Delivery Guarantee (messages lost):
//     broker.Publish("topic", criticalMsg)
//     // Consumer crashes before processing
//     // Message lost forever, no retry
//
//  6. No Backpressure (memory explosion):
//     // Fast publisher
//     for i := 0; i < 1000000; i++ {
//     broker.Publish("topic", i)
//     }
//     // Messages queued in channels
//     // Memory grows unbounded, eventual OOM
package naive

import (
	"fmt"
)

// Message represents a pub-sub message.
type Message struct {
	Topic   string
	Payload interface{}
}

// Handler processes messages.
type Handler func(msg Message)

// Broker represents a naive pub-sub broker.
//
// PROBLEMS:
// - No mutex (race conditions)
// - Blocking publishes
// - No subscriber cleanup
type Broker struct {
	subscribers map[string][]Handler // RACE: no mutex
}

// NewBroker creates a new naive broker.
func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[string][]Handler),
	}
}

// Subscribe registers a handler for a topic.
//
// PROBLEMS:
// - No subscription ID (can't unsubscribe)
// - Race condition on map
// - No buffering
// - Handler called synchronously (blocks publish)
func (b *Broker) Subscribe(topic string, handler Handler) {
	// PROBLEM: Concurrent map write - WILL PANIC!
	b.subscribers[topic] = append(b.subscribers[topic], handler)

	fmt.Printf("Subscribed to topic: %s (total: %d)\n", topic, len(b.subscribers[topic]))
}

// Publish sends a message to all subscribers.
//
// PROBLEMS:
// - Calls handlers synchronously (blocks on slow consumer)
// - No buffering (can't handle burst traffic)
// - Race condition on map read
// - No delivery guarantee (handler panic loses message)
func (b *Broker) Publish(topic string, payload interface{}) error {
	msg := Message{
		Topic:   topic,
		Payload: payload,
	}

	// PROBLEM: Concurrent map read without mutex
	handlers := b.subscribers[topic]

	// PROBLEM: Synchronous calls block publisher
	// Slowest handler determines publish latency!
	for _, handler := range handlers {
		// PROBLEM: If handler panics, message lost
		// No recovery, no retry
		handler(msg)
	}

	return nil
}

// PublishAsync sends message asynchronously.
//
// PROBLEMS:
// - Spawns unlimited goroutines (memory leak)
// - No limit on concurrent handlers
// - Still has race condition on map
func (b *Broker) PublishAsync(topic string, payload interface{}) error {
	msg := Message{
		Topic:   topic,
		Payload: payload,
	}

	// PROBLEM: Concurrent map read - RACE!
	handlers := b.subscribers[topic]

	// PROBLEM: Spawns goroutine for EACH subscriber
	// On popular topic with 1000 subscribers + 100 msgs/sec
	// = 100,000 goroutines per second created!
	for _, handler := range handlers {
		h := handler
		// UNBOUNDED GOROUTINE CREATION
		go func() {
			// PROBLEM: No recovery from panic
			h(msg)
		}()
	}

	return nil
}

// GetStats returns subscriber counts.
//
// PROBLEM: Race condition on map access
func (b *Broker) GetStats() map[string]int {
	stats := make(map[string]int)

	// RACE: concurrent map access without mutex
	for topic, handlers := range b.subscribers {
		stats[topic] = len(handlers)
	}

	return stats
}

// Close closes the broker.
//
// PROBLEM: Doesn't actually clean up anything
func (b *Broker) Close() error {
	// PROBLEM: Doesn't:
	// - Unregister subscribers
	// - Wait for in-flight messages
	// - Close any channels
	// - Stop any goroutines

	fmt.Println("Broker closed (nothing actually cleaned up)")
	return nil
}

// Example usage (demonstrates problems):
//
//	func main() {
//		broker := naive.NewBroker()
//
//		// PROBLEM 1: Blocking publish
//		broker.Subscribe("orders", func(msg Message) {
//			time.Sleep(time.Second) // Slow handler
//			fmt.Println("Order:", msg.Payload)
//		})
//
//		start := time.Now()
//		broker.Publish("orders", "order-1") // Blocks 1 second!
//		fmt.Printf("Publish took: %v\n", time.Since(start))
//
//		// PROBLEM 2: Race condition
//		go broker.Subscribe("topic", handler1)
//		go broker.Subscribe("topic", handler2)
//		go broker.Publish("topic", "msg")
//		// Race detector: WARNING: DATA RACE
//
//		// PROBLEM 3: Memory leak
//		for i := 0; i < 10000; i++ {
//			broker.Subscribe("spam", func(msg Message) {
//				// These accumulate forever!
//			})
//		}
//
//		// PROBLEM 4: Unbounded goroutines
//		broker.PublishAsync("popular", "msg")
//		// Spawns goroutine per subscriber
//		// With 1000 subscribers = 1000 goroutines per publish
//
//		// PROBLEM 5: No delivery guarantee
//		broker.Subscribe("critical", func(msg Message) {
//			// Process...
//			panic("crash!") // Message lost forever
//		})
//		broker.Publish("critical", "important-data")
//	}
//
// WHY THIS IS BAD:
//
// 1. Slow subscriber blocks all publishers (latency spike)
// 2. Memory leak from never-removed subscribers
// 3. Race conditions cause panics
// 4. Messages dropped on slow consumers
// 5. No delivery guarantee (lost on crash)
// 6. Unbounded goroutine creation (OOM)
// 7. No observability (can't monitor)
//
// REAL-WORLD IMPACT:
// - Publisher blocked: 1 slow consumer = entire system slow
// - Memory leak: Eventually OOM
// - Lost messages: Critical data disappeared
// - Crashes: Race detector panics in production
//
// PERFORMANCE:
// - With fast consumers: ~10k msgs/sec
// - With 1 slow consumer: Limited by slowest (serial processing)
// - Memory: Unbounded growth
// - Goroutines: Unlimited creation
//
// FIX STRATEGY (see improved/):
// 1. Add sync.RWMutex for subscribers map
// 2. Use buffered channels per subscriber
// 3. Non-blocking publish with select/default
// 4. Return subscription ID for cleanup
// 5. Add context for cancellation
// 6. Implement proper error handling
