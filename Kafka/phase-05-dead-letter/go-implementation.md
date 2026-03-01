# Phase 5 — Go Implementation

## Setup

### File Structure

```
go/
├── cmd/
│   ├── consumer-with-dlt/main.go
│   ├── dlt-consumer/main.go
│   ├── dlt-republisher/main.go
│   └── poison-producer/main.go
├── internal/
│   ├── deadletter/deadletter.go
│   ├── payment/simulator.go
│   └── retry/retry.go
├── go.mod
└── go.sum
```

---

## `internal/deadletter/deadletter.go` — Dead Letter Producer

```go
package deadletter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type DeadLetterMessage struct {
	OriginalTopic     string `json:"originalTopic"`
	OriginalPartition int    `json:"originalPartition"`
	OriginalOffset    int64  `json:"originalOffset"`
	OriginalKey       string `json:"originalKey"`
	OriginalValue     string `json:"originalValue"`
	Error             string `json:"error"`
	ErrorType         string `json:"errorType"`
	Attempts          int    `json:"attempts"`
	ConsumerGroup     string `json:"consumerGroup"`
	ConsumerID        string `json:"consumerId"`
	FailedAt          string `json:"failedAt"`
}

type Producer struct {
	writer     *kafka.Writer
	groupID    string
	consumerID string
}

func NewProducer(brokers []string, dlTopic, groupID, consumerID string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:  kafka.TCP(brokers...),
			Topic: dlTopic,
		},
		groupID:    groupID,
		consumerID: consumerID,
	}
}

func (p *Producer) Send(ctx context.Context, msg kafka.Message, err error, errorType string, attempts int) error {
	dlMsg := DeadLetterMessage{
		OriginalTopic:     msg.Topic,
		OriginalPartition: msg.Partition,
		OriginalOffset:    msg.Offset,
		OriginalKey:       string(msg.Key),
		OriginalValue:     string(msg.Value),
		Error:             err.Error(),
		ErrorType:         errorType,
		Attempts:          attempts,
		ConsumerGroup:     p.groupID,
		ConsumerID:        p.consumerID,
		FailedAt:          time.Now().UTC().Format(time.RFC3339),
	}

	value, _ := json.Marshal(dlMsg)

	writeErr := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   msg.Key, // Preserve original key
		Value: value,
		Headers: []kafka.Header{
			{Key: "x-original-topic", Value: []byte(msg.Topic)},
			{Key: "x-error-type", Value: []byte(errorType)},
			{Key: "x-failed-at", Value: []byte(dlMsg.FailedAt)},
		},
	})

	if writeErr != nil {
		return fmt.Errorf("failed to send to dead letter: %w", writeErr)
	}

	fmt.Printf("[DLT] 📤 Sent to dead-letter: %s (%s)\n", string(msg.Key), errorType)
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
```

---

## `cmd/consumer-with-dlt/main.go` — Consumer with Dead Letter Routing

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
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
	"order-pipeline/internal/deadletter"
	"order-pipeline/internal/payment"
	"order-pipeline/internal/retry"
)

type OrderEvent struct {
	EventType string  `json:"eventType"`
	OrderID   string  `json:"orderId"`
	UserID    string  `json:"userId"`
	Amount    float64 `json:"amount"`
}

func main() {
	consumerID := "consumer-go"
	if len(os.Args) > 1 {
		consumerID = os.Args[1]
	}

	groupID := "payment-dlt-group-go"

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"localhost:9092"},
		Topic:          "orders",
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0,
	})
	defer reader.Close()

	dlProducer := deadletter.NewProducer(
		[]string{"localhost:9092"},
		"orders.dead-letter",
		groupID,
		consumerID,
	)
	defer dlProducer.Close()

	log.Printf("[%s] Payment consumer with Dead Letter Topic", consumerID)
	log.Printf("[%s] Failed messages → orders.dead-letter", consumerID)
	fmt.Println()

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
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

		log.Printf("\n[%s] P%d:%d | key=%s",
			consumerID, msg.Partition, msg.Offset, string(msg.Key))

		// Step 1: Parse
		var order OrderEvent
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			log.Printf("[%s] ❌ Malformed message — sending to DLT", consumerID)
			dlProducer.Send(ctx, msg, err, "ParseError", 0)
			reader.CommitMessages(ctx, msg)
			continue
		}

		// Step 2: Validate
		if order.OrderID == "" || order.Amount <= 0 {
			validErr := fmt.Errorf("invalid order: orderId=%s, amount=%.2f", order.OrderID, order.Amount)
			log.Printf("[%s] ❌ Validation failed — sending to DLT", consumerID)
			dlProducer.Send(ctx, msg, validErr, "ValidationError", 0)
			reader.CommitMessages(ctx, msg)
			continue
		}

		log.Printf("[%s] Processing %s | $%.2f", consumerID, order.OrderID, order.Amount)

		// Step 3: Process with retries
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
			errorType := "TransientError"
			var permErr *retry.PermanentError
			if errors.As(retryResult.Err, &permErr) {
				errorType = "PermanentError"
			}

			dlProducer.Send(ctx, msg, retryResult.Err, errorType, retryResult.Attempts)
			reader.CommitMessages(ctx, msg)
			continue
		}

		log.Printf("[%s] ✅ %s charged (%s) — %d attempt(s)",
			consumerID, order.OrderID, paymentResult.Status, retryResult.Attempts)

		reader.CommitMessages(ctx, msg)
	}

	log.Printf("[%s] Shutdown complete", consumerID)
}
```

---

## `cmd/dlt-consumer/main.go` — Dead Letter Inspector

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"
	"order-pipeline/internal/deadletter"
)

func main() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "orders.dead-letter",
		GroupID:  "dlt-inspector-group-go",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	log.Println("[DLT Inspector] Monitoring orders.dead-letter...")
	fmt.Println()

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			continue
		}

		var dlMsg deadletter.DeadLetterMessage
		json.Unmarshal(msg.Value, &dlMsg)

		fmt.Println("╔══════════════════════════════════════════════")
		fmt.Println("║ DEAD LETTER MESSAGE")
		fmt.Println("╠──────────────────────────────────────────────")
		fmt.Printf("║ Key:            %s\n", string(msg.Key))
		fmt.Printf("║ Error Type:     %s\n", dlMsg.ErrorType)
		fmt.Printf("║ Error:          %s\n", dlMsg.Error)
		fmt.Printf("║ Attempts:       %d\n", dlMsg.Attempts)
		fmt.Printf("║ Consumer:       %s\n", dlMsg.ConsumerID)
		fmt.Printf("║ Failed At:      %s\n", dlMsg.FailedAt)
		fmt.Printf("║ Original Topic: %s\n", dlMsg.OriginalTopic)
		fmt.Printf("║ Original P:O:   %d:%d\n", dlMsg.OriginalPartition, dlMsg.OriginalOffset)

		if dlMsg.OriginalValue != "" {
			if len(dlMsg.OriginalValue) > 120 {
				fmt.Printf("║ Original Msg:   %s...\n", dlMsg.OriginalValue[:120])
			} else {
				fmt.Printf("║ Original Msg:   %s\n", dlMsg.OriginalValue)
			}
		}

		fmt.Println("╚══════════════════════════════════════════════")
		fmt.Println()
	}
}
```

---

## `cmd/poison-producer/main.go` — Intentionally Bad Messages

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	writer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "orders",
		Balancer: &kafka.Hash{},
	}
	defer writer.Close()

	type testMessage struct {
		key   string
		value string
		label string
	}

	goodOrder := func(id string, amount float64) string {
		v, _ := json.Marshal(map[string]interface{}{
			"eventType": "ORDER_CREATED",
			"orderId":   id,
			"userId":    "user-1",
			"amount":    amount,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return string(v)
	}

	messages := []testMessage{
		{"ORD-good-001", goodOrder("ORD-good-001", 49.99), "✅ Valid order"},
		{"ORD-good-002", goodOrder("ORD-good-002", 29.99), "✅ Valid order"},
		{"ORD-poison-001", "{this is not valid json!!!", "💀 Malformed JSON"},
		{"ORD-poison-002", `{"eventType":"ORDER_CREATED"}`, "💀 Missing orderId and amount"},
		{"ORD-poison-003", goodOrder("ORD-poison-003", -50), "💀 Negative amount"},
		{"ORD-good-003", goodOrder("ORD-good-003", 99.99), "✅ Valid order (after poison)"},
		{"ORD-poison-004", "", "💀 Empty body"},
	}

	fmt.Println("[Poison Producer] Sending mix of valid and invalid messages:")
	fmt.Println()

	for _, m := range messages {
		err := writer.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(m.key),
			Value: []byte(m.value),
		})
		if err != nil {
			log.Printf("  ❌ Failed to send %s: %v", m.key, err)
			continue
		}
		fmt.Printf("  Sent: %s (key=%s)\n", m.label, m.key)
	}

	fmt.Printf("\n[Poison Producer] Done. %d messages sent.\n", len(messages))
	fmt.Println("[Poison Producer] Watch the consumer route bad ones to dead-letter")
}
```

---

## Idiomatic Differences: TypeScript vs Go

| Aspect | TypeScript | Go |
|--------|-----------|-----|
| **DLT producer** | Same `producer.send()` API | Separate `kafka.Writer` for DLT topic |
| **Error routing** | `instanceof PermanentError` | `errors.As(&permErr)` |
| **Structured dead letter** | Object → `JSON.stringify` | Struct → `json.Marshal` |
| **Headers** | `headers: { "x-error-type": value }` | `Headers: []kafka.Header{{Key: ..., Value: []byte(...)}}` |
| **Cleanup** | `process.on("SIGINT")` + disconnect | `defer writer.Close()` + signal channel |

In Go, the dead letter producer is a clean separate package (`internal/deadletter/`). In TypeScript, it's typically a function in the same file or a utility module. Go's package-based structure naturally encourages separation.

---

## Running the Demo

```bash
# Terminal 1: Dead Letter Inspector
go run cmd/dlt-consumer/main.go

# Terminal 2: Main Consumer
go run cmd/consumer-with-dlt/main.go consumer-A

# Terminal 3: Send Poison Messages
go run cmd/poison-producer/main.go
```

Watch the main consumer skip bad messages and route them to the DLT. The inspector shows exactly why each message failed.

→ Next: [Phase 6 — Schema & Evolution](../phase-06-schemas/README.md)
