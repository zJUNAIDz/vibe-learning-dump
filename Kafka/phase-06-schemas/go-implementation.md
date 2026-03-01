# Phase 6 — Go Implementation

## Setup

### File Structure

```
go/
├── cmd/
│   ├── schema-v1-producer/main.go
│   ├── schema-v2-producer/main.go
│   ├── multi-version-consumer/main.go
│   └── schema-validator/main.go
├── internal/
│   ├── deadletter/deadletter.go
│   └── schema/schema.go
├── go.mod
└── go.sum
```

---

## `internal/schema/schema.go` — Schema Definitions & Normalization

```go
package schema

import (
	"encoding/json"
	"fmt"
)

// ─── Envelope (stable across all versions) ───

type EventEnvelope struct {
	Version   int             `json:"version"`
	EventType string          `json:"eventType"`
	Timestamp string          `json:"timestamp"`
	Source    string           `json:"source"`
	Payload   json.RawMessage `json:"payload"`
}

// ─── V1 Payload ───

type OrderPayloadV1 struct {
	OrderID string  `json:"orderId"`
	UserID  string  `json:"userId"`
	Amount  float64 `json:"amount"`
}

// ─── V2 Payload ───

type OrderPayloadV2 struct {
	OrderID  string     `json:"orderId"`
	UserID   string     `json:"userId"`
	Amount   float64    `json:"amount"`
	Currency string     `json:"currency"`
	Items    []LineItem `json:"items,omitempty"`
}

type LineItem struct {
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

// ─── Normalized internal representation ───

type NormalizedOrder struct {
	OrderID         string
	UserID          string
	Amount          float64
	Currency        string
	Items           []LineItem
	OriginalVersion int
}

// ─── Validation ───

func ValidateV1(payload json.RawMessage) (*OrderPayloadV1, error) {
	var p OrderPayloadV1
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("v1 parse error: %w", err)
	}
	if p.OrderID == "" {
		return nil, fmt.Errorf("v1: orderId is required")
	}
	if p.UserID == "" {
		return nil, fmt.Errorf("v1: userId is required")
	}
	if p.Amount <= 0 {
		return nil, fmt.Errorf("v1: amount must be positive, got %.2f", p.Amount)
	}
	return &p, nil
}

func ValidateV2(payload json.RawMessage) (*OrderPayloadV2, error) {
	var p OrderPayloadV2
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("v2 parse error: %w", err)
	}
	if p.OrderID == "" {
		return nil, fmt.Errorf("v2: orderId is required")
	}
	if p.UserID == "" {
		return nil, fmt.Errorf("v2: userId is required")
	}
	if p.Amount <= 0 {
		return nil, fmt.Errorf("v2: amount must be positive, got %.2f", p.Amount)
	}
	if len(p.Currency) != 3 {
		return nil, fmt.Errorf("v2: currency must be 3-letter code, got %q", p.Currency)
	}
	return &p, nil
}

// ─── Normalization ───

func Normalize(envelope EventEnvelope) (*NormalizedOrder, error) {
	switch envelope.Version {
	case 1:
		p, err := ValidateV1(envelope.Payload)
		if err != nil {
			return nil, err
		}
		return &NormalizedOrder{
			OrderID:         p.OrderID,
			UserID:          p.UserID,
			Amount:          p.Amount,
			Currency:        "USD", // default for v1
			Items:           nil,   // not available in v1
			OriginalVersion: 1,
		}, nil

	case 2:
		p, err := ValidateV2(envelope.Payload)
		if err != nil {
			return nil, err
		}
		return &NormalizedOrder{
			OrderID:         p.OrderID,
			UserID:          p.UserID,
			Amount:          p.Amount,
			Currency:        p.Currency,
			Items:           p.Items,
			OriginalVersion: 2,
		}, nil

	default:
		return nil, fmt.Errorf("unknown schema version: %d", envelope.Version)
	}
}
```

---

## `cmd/schema-v1-producer/main.go` — Legacy Producer

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

	type v1Envelope struct {
		Version   int         `json:"version"`
		EventType string      `json:"eventType"`
		Timestamp string      `json:"timestamp"`
		Source    string       `json:"source"`
		Payload   interface{} `json:"payload"`
	}

	type v1Payload struct {
		OrderID string  `json:"orderId"`
		UserID  string  `json:"userId"`
		Amount  float64 `json:"amount"`
	}

	orders := []v1Envelope{
		{
			Version: 1, EventType: "ORDER_CREATED",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Source: "order-service-v1",
			Payload: v1Payload{OrderID: "ORD-v1-001", UserID: "user-10", Amount: 49.99},
		},
		{
			Version: 1, EventType: "ORDER_CREATED",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Source: "order-service-v1",
			Payload: v1Payload{OrderID: "ORD-v1-002", UserID: "user-11", Amount: 129.50},
		},
		{
			Version: 1, EventType: "ORDER_CREATED",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Source: "order-service-v1",
			Payload: v1Payload{OrderID: "ORD-v1-003", UserID: "user-12", Amount: 9.99},
		},
	}

	for _, order := range orders {
		value, _ := json.Marshal(order)
		p := order.Payload.(v1Payload)

		err := writer.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(p.OrderID),
			Value: value,
			Headers: []kafka.Header{
				{Key: "schema-version", Value: []byte("1")},
			},
		})
		if err != nil {
			log.Printf("❌ Failed: %v", err)
			continue
		}
		fmt.Printf("Sent v1: %s | $%.2f\n", p.OrderID, p.Amount)
	}

	fmt.Printf("\n[v1 Producer] %d messages sent (no currency field)\n", len(orders))
}
```

---

## `cmd/schema-v2-producer/main.go` — Evolved Producer

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

type lineItem struct {
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type v2Payload struct {
	OrderID  string     `json:"orderId"`
	UserID   string     `json:"userId"`
	Amount   float64    `json:"amount"`
	Currency string     `json:"currency"`
	Items    []lineItem `json:"items,omitempty"`
}

type v2Envelope struct {
	Version   int         `json:"version"`
	EventType string      `json:"eventType"`
	Timestamp string      `json:"timestamp"`
	Source    string       `json:"source"`
	Payload   interface{} `json:"payload"`
}

func main() {
	writer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "orders",
		Balancer: &kafka.Hash{},
	}
	defer writer.Close()

	orders := []v2Envelope{
		{
			Version: 2, EventType: "ORDER_CREATED",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Source: "order-service-v2",
			Payload: v2Payload{
				OrderID: "ORD-v2-001", UserID: "user-20", Amount: 149.99,
				Currency: "EUR",
				Items: []lineItem{
					{SKU: "LAPTOP-001", Name: "USB-C Dock", Quantity: 1, Price: 149.99},
				},
			},
		},
		{
			Version: 2, EventType: "ORDER_CREATED",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Source: "order-service-v2",
			Payload: v2Payload{
				OrderID: "ORD-v2-002", UserID: "user-21", Amount: 59.98,
				Currency: "GBP",
				Items: []lineItem{
					{SKU: "CABLE-001", Name: "HDMI Cable", Quantity: 2, Price: 29.99},
				},
			},
		},
		{
			Version: 2, EventType: "ORDER_CREATED",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Source: "order-service-v2",
			Payload: v2Payload{
				OrderID: "ORD-v2-003", UserID: "user-22", Amount: 299.97,
				Currency: "USD",
			},
		},
	}

	for _, order := range orders {
		value, _ := json.Marshal(order)
		p := order.Payload.(v2Payload)

		err := writer.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(p.OrderID),
			Value: value,
			Headers: []kafka.Header{
				{Key: "schema-version", Value: []byte("2")},
			},
		})
		if err != nil {
			log.Printf("❌ Failed: %v", err)
			continue
		}
		fmt.Printf("Sent v2: %s | $%.2f %s | %d items\n",
			p.OrderID, p.Amount, p.Currency, len(p.Items))
	}

	fmt.Printf("\n[v2 Producer] %d messages sent (with currency + items)\n", len(orders))
}
```

---

## `cmd/multi-version-consumer/main.go` — Handles v1 and v2

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
	"order-pipeline/internal/schema"
)

func main() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"localhost:9092"},
		Topic:          "orders",
		GroupID:        "payment-schema-group-go",
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0,
	})
	defer reader.Close()

	// DLT writer for unknown versions / validation failures
	dltWriter := &kafka.Writer{
		Addr:  kafka.TCP("localhost:9092"),
		Topic: "orders.dead-letter",
	}
	defer dltWriter.Close()

	log.Println("[Multi-Version Consumer] Handles v1 + v2 transparently")
	log.Println("[Multi-Version Consumer] Unknown versions → DLT")
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
			continue
		}

		key := string(msg.Key)
		fmt.Printf("\nP%d:%d | key=%s\n", msg.Partition, msg.Offset, key)

		// Step 1: Parse envelope
		var envelope schema.EventEnvelope
		if err := json.Unmarshal(msg.Value, &envelope); err != nil {
			fmt.Printf("  ❌ Malformed JSON → DLT\n")
			sendToDLT(ctx, dltWriter, msg, "Malformed JSON")
			reader.CommitMessages(ctx, msg)
			continue
		}

		// Step 2: Check version exists
		if envelope.Version == 0 {
			fmt.Printf("  ❌ Missing version field → DLT\n")
			sendToDLT(ctx, dltWriter, msg, "Missing version field")
			reader.CommitMessages(ctx, msg)
			continue
		}

		// Step 3: Normalize (validates internally per version)
		normalized, err := schema.Normalize(envelope)
		if err != nil {
			fmt.Printf("  ⚠️ %v → DLT\n", err)
			sendToDLT(ctx, dltWriter, msg, err.Error())
			reader.CommitMessages(ctx, msg)
			continue
		}

		// Step 4: Process
		fmt.Printf("  ✅ v%d → normalized | %s | $%.2f %s\n",
			normalized.OriginalVersion, normalized.OrderID,
			normalized.Amount, normalized.Currency)

		fmt.Printf("    💳 Charging $%.2f %s\n", normalized.Amount, normalized.Currency)
		if len(normalized.Items) > 0 {
			for _, item := range normalized.Items {
				fmt.Printf("    📦 %s x%d\n", item.Name, item.Quantity)
			}
		}

		reader.CommitMessages(ctx, msg)
	}

	log.Println("Shutdown complete")
}

func sendToDLT(ctx context.Context, writer *kafka.Writer, msg kafka.Message, errMsg string) {
	dlMsg, _ := json.Marshal(map[string]interface{}{
		"originalTopic": msg.Topic,
		"originalKey":   string(msg.Key),
		"originalValue": string(msg.Value),
		"error":         errMsg,
		"errorType":     "SchemaError",
	})

	writer.WriteMessages(ctx, kafka.Message{
		Key:   msg.Key,
		Value: dlMsg,
	})
}
```

---

## `cmd/schema-validator/main.go` — Audit Tool

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"order-pipeline/internal/schema"
)

type validationResult struct {
	Offset    int64
	Partition int
	Key       string
	Version   int
	Valid     bool
	Errors    []string
}

func main() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "orders",
		GroupID:  "schema-validator-group-go",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	log.Println("[Schema Validator] Auditing all messages on 'orders' topic")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var results []validationResult

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			break // timeout or error
		}

		key := string(msg.Key)
		result := validationResult{
			Offset:    msg.Offset,
			Partition: msg.Partition,
			Key:       key,
		}

		// Parse envelope
		var envelope schema.EventEnvelope
		if err := json.Unmarshal(msg.Value, &envelope); err != nil {
			result.Errors = append(result.Errors, "Malformed JSON")
			results = append(results, result)
			reader.CommitMessages(ctx, msg)
			continue
		}

		if envelope.Version == 0 {
			result.Errors = append(result.Errors, "Missing version")
			results = append(results, result)
			reader.CommitMessages(ctx, msg)
			continue
		}

		result.Version = envelope.Version

		// Validate
		_, err = schema.Normalize(envelope)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
		} else {
			result.Valid = true
		}

		results = append(results, result)
		reader.CommitMessages(ctx, msg)
	}

	// Print report
	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("SCHEMA VALIDATION REPORT")
	fmt.Println("════════════════════════════════════════════════════════════")

	valid := 0
	invalid := 0
	versionCounts := make(map[int]int)

	for _, r := range results {
		if r.Valid {
			valid++
		} else {
			invalid++
		}
		versionCounts[r.Version]++
	}

	fmt.Printf("Total messages:  %d\n", len(results))
	fmt.Printf("Valid:           %d\n", valid)
	fmt.Printf("Invalid:         %d\n", invalid)
	fmt.Println()

	fmt.Println("Version Distribution:")
	for ver, count := range versionCounts {
		if ver == 0 {
			fmt.Printf("  v??: %d messages\n", count)
		} else {
			fmt.Printf("  v%d:  %d messages\n", ver, count)
		}
	}

	if invalid > 0 {
		fmt.Println("\nInvalid Messages:")
		for _, r := range results {
			if !r.Valid {
				fmt.Printf("  P%d:%d | %s | %v\n", r.Partition, r.Offset, r.Key, r.Errors)
			}
		}
	}

	fmt.Println("════════════════════════════════════════════════════════════")
}
```

---

## Idiomatic Differences: TypeScript vs Go

| Aspect | TypeScript | Go |
|--------|-----------|-----|
| **Schema types** | Interfaces + type guards | Structs + explicit validation functions |
| **Raw payload** | `as EventEnvelope` cast | `json.RawMessage` for deferred parsing |
| **Normalization** | Switch on version → cast | Switch on version → unmarshal into specific struct |
| **Validation** | `typeof` + `instanceof` checks | Field-by-field validation returning errors |
| **Unknown version** | `throw new Error()` | `return nil, fmt.Errorf()` |

Go's `json.RawMessage` is key: the envelope's `Payload` field stays as raw bytes until you know the version and can unmarshal it into the correct struct. TypeScript uses `unknown` with type guards for the same purpose.

---

## Running the Demo

```bash
# Terminal 1: Multi-version consumer
go run cmd/multi-version-consumer/main.go

# Terminal 2: Send v1 events
go run cmd/schema-v1-producer/main.go

# Terminal 3: Send v2 events
go run cmd/schema-v2-producer/main.go

# Terminal 4: Run audit
go run cmd/schema-validator/main.go
```

### What to Watch

1. Consumer processes v1 and v2 messages identically through normalization
2. v1 orders get `currency: "USD"` and empty items by default
3. v2 orders carry their explicit currency and item list
4. Schema validator shows version distribution across the topic

→ Next: [Phase 7 — Replay & Retention](../phase-07-replay/README.md)
