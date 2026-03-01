# Phase 4 — Go Implementation

## Setup

### File Structure

```
go/
├── cmd/
│   ├── idempotent-consumer/main.go
│   └── idempotent-producer/main.go
├── internal/
│   ├── payment/simulator.go
│   └── retry/retry.go
├── go.mod
└── go.sum
```

---

## `internal/retry/retry.go` — Retry with Exponential Backoff

```go
package retry

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// TransientError indicates the operation can be retried
type TransientError struct {
	Message string
}

func (e *TransientError) Error() string { return e.Message }

// PermanentError indicates the operation should NOT be retried
type PermanentError struct {
	Message string
}

func (e *PermanentError) Error() string { return e.Message }

type Options struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

type Result struct {
	Attempts int
	Err      error
}

// Do executes fn with exponential backoff retries.
// It only retries on TransientError. PermanentError stops immediately.
func Do(fn func() error, opts Options) Result {
	for attempt := 1; attempt <= opts.MaxRetries; attempt++ {
		err := fn()
		if err == nil {
			return Result{Attempts: attempt, Err: nil}
		}

		// Permanent error — stop immediately
		var permErr *PermanentError
		if errors.As(err, &permErr) {
			fmt.Printf("  [Retry] Permanent error on attempt %d: %s\n", attempt, err)
			return Result{Attempts: attempt, Err: err}
		}

		// Transient error — retry with backoff
		if attempt < opts.MaxRetries {
			delay := time.Duration(
				math.Min(
					float64(opts.BaseDelay)*math.Pow(2, float64(attempt-1))+
						float64(rand.Intn(100))*float64(time.Millisecond),
					float64(opts.MaxDelay),
				),
			)
			fmt.Printf("  [Retry] Attempt %d/%d failed: %s. Retrying in %v...\n",
				attempt, opts.MaxRetries, err, delay)
			time.Sleep(delay)
		} else {
			fmt.Printf("  [Retry] All %d attempts exhausted: %s\n", opts.MaxRetries, err)
			return Result{Attempts: attempt, Err: err}
		}
	}

	return Result{Attempts: opts.MaxRetries, Err: fmt.Errorf("unreachable")}
}
```

---

## `internal/payment/simulator.go` — Flaky Payment Gateway

```go
package payment

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"order-pipeline/internal/retry"
)

type PaymentResult struct {
	OrderID       string `json:"orderId"`
	Status        string `json:"status"`
	TransactionID string `json:"transactionId"`
}

var (
	chargedOrders = make(map[string]bool)
	mu            sync.Mutex
)

// Charge simulates a payment gateway with transient and permanent failures.
func Charge(orderID string, amount float64) (*PaymentResult, error) {
	// Simulate network latency
	time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

	// Idempotency check
	mu.Lock()
	alreadyCharged := chargedOrders[orderID]
	mu.Unlock()

	if alreadyCharged {
		fmt.Printf("  [Payment] Order %s already charged — skipping (idempotent)\n", orderID)
		return &PaymentResult{
			OrderID:       orderID,
			Status:        "already_charged",
			TransactionID: fmt.Sprintf("TXN-%s-DUP", orderID),
		}, nil
	}

	// ~30% transient failure
	if rand.Float64() < 0.3 {
		return nil, &retry.TransientError{
			Message: fmt.Sprintf("Payment gateway timeout for order %s", orderID),
		}
	}

	// ~5% permanent failure
	if rand.Float64() < 0.05 {
		return nil, &retry.PermanentError{
			Message: fmt.Sprintf("Card declined for order %s: insufficient funds", orderID),
		}
	}

	// Success
	mu.Lock()
	chargedOrders[orderID] = true
	mu.Unlock()

	return &PaymentResult{
		OrderID:       orderID,
		Status:        "charged",
		TransactionID: fmt.Sprintf("TXN-%s-%d", orderID, time.Now().UnixMilli()),
	}, nil
}
```

---

## `cmd/idempotent-consumer/main.go` — Consumer with Retry + Idempotency

```go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
	"order-pipeline/internal/payment"
	"order-pipeline/internal/retry"
)

type OrderEvent struct {
	EventType string  `json:"eventType"`
	OrderID   string  `json:"orderId"`
	UserID    string  `json:"userId"`
	ItemID    string  `json:"itemId"`
	Quantity  int     `json:"quantity"`
	Amount    float64 `json:"amount"`
	Timestamp string  `json:"timestamp"`
}

type ProcessingRecord struct {
	Status      string
	ProcessedAt time.Time
}

var (
	processedOrders = make(map[string]ProcessingRecord)
	mu              sync.Mutex
)

func main() {
	consumerID := "consumer-go"
	if len(os.Args) > 1 {
		consumerID = os.Args[1]
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"localhost:9092"},
		Topic:          "orders",
		GroupID:        "payment-retry-group-go",
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0, // Manual commit
	})
	defer reader.Close()

	log.Printf("[%s] Payment consumer with retry + idempotency", consumerID)
	log.Printf("[%s] Max retries: 3, Backoff: 200ms-2000ms", consumerID)
	fmt.Println()

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Printf("\n[%s] Shutting down...", consumerID)
		printSummary(consumerID)
		cancel()
	}()

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("[%s] Fetch error: %v", consumerID, err)
			continue
		}

		var order OrderEvent
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			log.Printf("[%s] Bad message at P%d:%d — skipping", consumerID, msg.Partition, msg.Offset)
			reader.CommitMessages(ctx, msg)
			continue
		}

		log.Printf("\n[%s] P%d:%d | %s | %s | $%.2f",
			consumerID, msg.Partition, msg.Offset,
			order.EventType, order.OrderID, order.Amount)

		// IDEMPOTENCY CHECK
		mu.Lock()
		record, alreadyProcessed := processedOrders[order.OrderID]
		mu.Unlock()

		if alreadyProcessed {
			log.Printf("[%s] ⏭️ Skipping %s — already processed (%s at %s)",
				consumerID, order.OrderID, record.Status, record.ProcessedAt.Format(time.RFC3339))
			reader.CommitMessages(ctx, msg)
			continue
		}

		// RETRY with exponential backoff
		var paymentResult *payment.PaymentResult
		retryResult := retry.Do(
			func() error {
				result, err := payment.Charge(order.OrderID, order.Amount)
				if err != nil {
					return err
				}
				paymentResult = result
				return nil
			},
			retry.Options{
				MaxRetries: 3,
				BaseDelay:  200 * time.Millisecond,
				MaxDelay:   2 * time.Second,
			},
		)

		if retryResult.Err != nil {
			var permErr *retry.PermanentError
			status := "retry_exhausted"
			if errors.As(retryResult.Err, &permErr) {
				status = "permanently_failed"
				log.Printf("[%s] ❌ PERMANENT FAILURE for %s: %s",
					consumerID, order.OrderID, retryResult.Err)
				log.Printf("[%s] 📤 Should send to dead-letter topic (Phase 5)",
					consumerID)
			} else {
				log.Printf("[%s] ❌ TRANSIENT FAILURE for %s after %d attempts",
					consumerID, order.OrderID, retryResult.Attempts)
				log.Printf("[%s] 📤 Should send to dead-letter topic for later retry",
					consumerID)
			}

			mu.Lock()
			processedOrders[order.OrderID] = ProcessingRecord{
				Status:      status,
				ProcessedAt: time.Now(),
			}
			mu.Unlock()
		} else {
			log.Printf("[%s] ✅ %s charged (%s) in %d attempt(s) — TXN: %s",
				consumerID, order.OrderID, paymentResult.Status,
				retryResult.Attempts, paymentResult.TransactionID)

			mu.Lock()
			processedOrders[order.OrderID] = ProcessingRecord{
				Status:      "success",
				ProcessedAt: time.Now(),
			}
			mu.Unlock()
		}

		// Commit offset — message is handled (success or dead-lettered)
		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("[%s] ❌ Commit failed: %v", consumerID, err)
		}
	}

	log.Printf("[%s] Shutdown complete", consumerID)
}

func printSummary(consumerID string) {
	mu.Lock()
	defer mu.Unlock()

	fmt.Printf("\n%s\n", "═══════════════════════════════════════════════════")
	fmt.Println("Processing Summary:")
	success, failed := 0, 0
	for orderID, rec := range processedOrders {
		if rec.Status == "success" {
			success++
		} else {
			failed++
		}
		fmt.Printf("  %s: %s\n", orderID, rec.Status)
	}
	fmt.Printf("\nTotal: %d (✅ %d | ❌ %d)\n", len(processedOrders), success, failed)
	fmt.Printf("%s\n\n", "═══════════════════════════════════════════════════")
}
```

---

## `cmd/idempotent-producer/main.go` — Idempotent Producer

```go
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func main() {
	// kafka-go doesn't have a built-in "idempotent" flag like kafkajs.
	// In practice, you'd configure this at the broker level or use
	// the confluent-kafka-go library for full idempotent producer support.
	//
	// For kafka-go, the important things are:
	// 1. Set RequiredAcks to all replicas (kafka.RequireAll)
	// 2. Use a single writer (don't create multiple writers for same topic)
	// 3. Let the writer handle retries internally

	writer := &kafka.Writer{
		Addr:         kafka.TCP("localhost:9092"),
		Topic:        "orders",
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll, // Wait for all in-sync replicas
		MaxAttempts:  3,                // Internal retry on transient failures
	}
	defer writer.Close()

	log.Println("[Producer] Connected (RequiredAcks=All, MaxAttempts=3)")
	log.Println("[Producer] Type: userId itemId quantity amount")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		parts := strings.Fields(strings.TrimSpace(scanner.Text()))
		if len(parts) != 4 {
			log.Println("Usage: userId itemId quantity amount")
			continue
		}

		quantity, _ := strconv.Atoi(parts[2])
		amount, _ := strconv.ParseFloat(parts[3], 64)
		orderID := fmt.Sprintf("ORD-%s", uuid.New().String()[:8])

		order := map[string]interface{}{
			"eventType": "ORDER_CREATED",
			"orderId":   orderID,
			"userId":    parts[0],
			"itemId":    parts[1],
			"quantity":  quantity,
			"amount":    amount,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		value, _ := json.Marshal(order)

		err := writer.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(orderID),
			Value: value,
		})
		if err != nil {
			log.Printf("[Producer] ❌ Failed: %v", err)
			continue
		}

		log.Printf("[Producer] ✅ %s sent", orderID)
	}
}
```

---

## Idiomatic Differences: TypeScript vs Go

| Aspect | TypeScript | Go |
|--------|-----------|-----|
| **Idempotent producer** | `idempotent: true` — one flag | `kafka-go` doesn't expose this. Use `RequiredAcks: All`. For true idempotent producer, use `confluent-kafka-go`. |
| **Error types** | Custom error classes + `instanceof` | Custom error types + `errors.As()` |
| **Retry** | `async`/`await` with `setTimeout` | Blocking `time.Sleep` in same goroutine |
| **State store** | `Map<string, object>` | `map[string]struct` + `sync.Mutex` |
| **Signal handling** | `process.on("SIGINT", ...)` | `signal.Notify` + channel |

The key difference is error handling. TypeScript uses `instanceof` for error type checking — clean but fragile (errors across module boundaries can fail `instanceof`). Go uses `errors.As()` which works across package boundaries reliably.

---

## Running the Demo

```bash
# Terminal 1: Consumer
go run cmd/idempotent-consumer/main.go consumer-A

# Terminal 2: Producer
go run cmd/idempotent-producer/main.go
# Type orders and watch retries + idempotency in action
```

→ Next: [Phase 5 — Dead Letters & Poison Messages](../phase-05-dead-letter/README.md)
