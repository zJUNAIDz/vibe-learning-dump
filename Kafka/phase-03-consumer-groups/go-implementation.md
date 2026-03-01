# Phase 3 — Go Implementation

## Setup

### File Structure

```
go/
├── cmd/
│   ├── group-consumer/main.go
│   ├── manual-commit-consumer/main.go
│   ├── producer-load/main.go
│   └── group-inspector/main.go
├── go.mod
└── go.sum
```

---

## `cmd/group-consumer/main.go` — Consumer with Explicit Offset Handling

In `kafka-go`, `ReadMessage` both reads and commits. To separate read from commit, use `FetchMessage` + `CommitMessages`.

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
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

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"localhost:9092"},
		Topic:          "orders",
		GroupID:        "payment-group-v3-go",
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0, // Disable auto-commit — we commit manually
	})
	defer reader.Close()

	log.Printf("[%s] Subscribed to 'orders' (group: payment-group-v3-go)", consumerID)
	log.Printf("[%s] Using MANUAL commit (FetchMessage + CommitMessages)", consumerID)
	fmt.Println()

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Printf("\n[%s] Shutting down...", consumerID)
		cancel()
	}()

	for {
		// FetchMessage reads but does NOT commit
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("[%s] Fetch error: %v", consumerID, err)
			continue
		}

		var order OrderEvent
		json.Unmarshal(msg.Value, &order)

		log.Printf("[%s] P%d:%d | %s | %s | $%.2f",
			consumerID, msg.Partition, msg.Offset,
			order.EventType, order.OrderID, order.Amount)

		// Simulate processing (200-800ms)
		processingTime := time.Duration(200+rand.Intn(600)) * time.Millisecond
		time.Sleep(processingTime)

		// NOW commit — after successful processing
		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("[%s] ❌ Commit failed: %v", consumerID, err)
			// In production: retry commit or handle gracefully
			continue
		}

		log.Printf("[%s] ✅ Processed + committed %s (%v)", consumerID, order.OrderID, processingTime)
	}

	log.Printf("[%s] Shutdown complete", consumerID)
}
```

### `FetchMessage` vs `ReadMessage` — The Critical Difference

```
ReadMessage  = Fetch + Auto-Commit (convenient but dangerous)
FetchMessage = Fetch only (you must call CommitMessages yourself)
```

In production, always use `FetchMessage` + `CommitMessages`. This gives you control over when the offset is committed — after you've actually processed the message.

---

## `cmd/manual-commit-consumer/main.go` — Batch Commits (Optimization)

Committing after every single message is safe but has overhead. In high-throughput scenarios, you batch commits.

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
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	consumerID := "batch-consumer"
	if len(os.Args) > 1 {
		consumerID = os.Args[1]
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"localhost:9092"},
		Topic:          "orders",
		GroupID:        "payment-batch-group",
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0,
	})
	defer reader.Close()

	log.Printf("[%s] Using BATCH commits (every 10 messages or 5 seconds)", consumerID)
	fmt.Println()

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	const batchSize = 10
	var pending []kafka.Message
	lastCommit := time.Now()

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("[%s] Error: %v", consumerID, err)
			continue
		}

		var event map[string]interface{}
		json.Unmarshal(msg.Value, &event)

		log.Printf("[%s] P%d:%d | %s",
			consumerID, msg.Partition, msg.Offset, event["orderId"])

		// Process message...
		pending = append(pending, msg)

		// Commit when batch is full OR 5 seconds have passed
		if len(pending) >= batchSize || time.Since(lastCommit) > 5*time.Second {
			if err := reader.CommitMessages(ctx, pending...); err != nil {
				log.Printf("[%s] ❌ Batch commit failed: %v", consumerID, err)
			} else {
				log.Printf("[%s] ✅ Committed batch of %d messages", consumerID, len(pending))
			}
			pending = pending[:0] // Clear the batch
			lastCommit = time.Now()
		}
	}

	// Commit any remaining on shutdown
	if len(pending) > 0 {
		reader.CommitMessages(context.Background(), pending...)
		log.Printf("[%s] Committed final %d messages", consumerID, len(pending))
	}

	log.Printf("[%s] Shutdown complete", consumerID)
}
```

### Tradeoff: Batch Size

| Batch Size | Throughput | Risk on Crash |
|-----------|-----------|---------------|
| 1 (commit every message) | Lower (more Kafka RPCs) | Re-process at most 1 message |
| 10 | Higher | Re-process up to 10 messages |
| 100 | Highest | Re-process up to 100 messages |

Choose based on how expensive reprocessing is. For idempotent processing (Phase 4), larger batches are fine.

---

## `cmd/producer-load/main.go` — Load Producer

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func main() {
	rate := 2 // orders per second
	total := 50

	if len(os.Args) > 1 {
		rate, _ = strconv.Atoi(os.Args[1])
	}
	if len(os.Args) > 2 {
		total, _ = strconv.Atoi(os.Args[2])
	}

	writer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "orders",
		Balancer: &kafka.Hash{},
	}
	defer writer.Close()

	items := []string{"ITEM-001", "ITEM-002", "ITEM-003", "ITEM-004", "ITEM-005"}
	users := []string{"user-1", "user-2", "user-3", "user-4", "user-5"}

	log.Printf("[Load] Producing %d orders at %d/sec", total, rate)
	fmt.Println()

	ticker := time.NewTicker(time.Second / time.Duration(rate))
	defer ticker.Stop()

	produced := 0
	for range ticker.C {
		if produced >= total {
			break
		}

		orderID := fmt.Sprintf("ORD-%s", uuid.New().String()[:8])
		order := map[string]interface{}{
			"eventType": "ORDER_CREATED",
			"orderId":   orderID,
			"userId":    users[rand.Intn(len(users))],
			"itemId":    items[rand.Intn(len(items))],
			"quantity":  1 + rand.Intn(5),
			"amount":    float64(10+rand.Intn(90)) + rand.Float64(),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		value, _ := json.Marshal(order)

		err := writer.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(orderID),
			Value: value,
		})
		if err != nil {
			log.Printf("[Load] ❌ %v", err)
			continue
		}

		produced++
		log.Printf("[Load] #%d %s (%s, $%.2f)",
			produced, orderID, order["userId"], order["amount"])
	}

	log.Printf("[Load] Done. Produced %d orders.", produced)
}
```

---

## `cmd/group-inspector/main.go` — Inspect Consumer Group

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
)

func main() {
	groupID := "payment-group-v3-go"
	if len(os.Args) > 1 {
		groupID = os.Args[1]
	}

	// Use the low-level connection to inspect groups
	conn, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Get partition info
	partitions, err := conn.ReadPartitions("orders")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nTopic: orders (%d partitions)\n", len(partitions))
	fmt.Println("─────────────────────────────")

	for _, p := range partitions {
		// Connect to partition leader to get offsets
		leaderAddr := fmt.Sprintf("%s:%d", p.Leader.Host, p.Leader.Port)
		pConn, err := kafka.DialLeader(context.Background(), "tcp", leaderAddr, "orders", p.ID)
		if err != nil {
			log.Printf("Failed to connect to partition %d: %v", p.ID, err)
			continue
		}

		first, last, _ := pConn.ReadOffsets()
		fmt.Printf("  Partition %d: offsets [%d, %d) (%d messages)\n",
			p.ID, first, last, last-first)

		pConn.Close()
	}

	fmt.Printf("\nUse kafka CLI for detailed group info:\n")
	fmt.Printf("docker exec -it kafka kafka-consumer-groups \\\n")
	fmt.Printf("  --bootstrap-server localhost:9092 \\\n")
	fmt.Printf("  --describe --group %s\n\n", groupID)
}
```

---

## Idiomatic Differences: TypeScript vs Go

| Aspect | TypeScript (kafkajs) | Go (kafka-go) |
|--------|---------------------|---------------|
| **Auto vs Manual commit** | `autoCommitThreshold`, `autoCommitInterval` settings | `CommitInterval: 0` disables auto-commit; use `FetchMessage` + `CommitMessages` |
| **Rebalance events** | `consumer.on('consumer.rebalancing')` callback | No built-in event. Reader handles internally. Log and monitor externally. |
| **Batch commit** | `consumer.commitOffsets([...])` | `reader.CommitMessages(ctx, ...msgs)` |
| **Group inspection** | `admin.describeGroups()` API | Direct TCP inspection or CLI tools |
| **Concurrency model** | Single event loop, callback-based | Goroutine per consumer, blocking read loop |
| **Heartbeat** | `heartbeat()` function in handler | Handled by Reader internally on separate goroutine |

The biggest practical difference is rebalance visibility. `kafkajs` emits events you can listen to. `kafka-go` handles rebalances internally — you monitor via Kafka CLI or metrics.

In production Go services, you'd pair `kafka-go` with Prometheus metrics to observe rebalance frequency and partition assignments.

---

## Running the Demo

```bash
# Terminal 1: First consumer
go run cmd/group-consumer/main.go consumer-A

# Terminal 2: Load producer
go run cmd/producer-load/main.go 2 100

# Terminal 3: Second consumer (watch the rebalance)
go run cmd/group-consumer/main.go consumer-B

# Terminal 4: Third consumer
go run cmd/group-consumer/main.go consumer-C

# Kill consumer-B → watch rebalance redistribute partitions

# Inspect the group
docker exec -it kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --describe --group payment-group-v3-go
```

→ Next: [Phase 4 — Failure, Retries & Idempotency](../phase-04-failure-retries/README.md)
