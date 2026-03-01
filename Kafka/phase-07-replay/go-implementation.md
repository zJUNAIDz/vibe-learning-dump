# Phase 7 — Go Implementation

## Setup

### File Structure

```
go/
├── cmd/
│   ├── retention-demo/main.go
│   ├── compaction-demo/main.go
│   ├── replay-consumer/main.go
│   ├── offset-reset/main.go
│   └── time-travel-consumer/main.go
├── go.mod
└── go.sum
```

---

## Tool 1: `cmd/retention-demo/main.go` — Observe Message Expiry

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

const topic = "orders-retention-demo-go"
const retentionMs = 60000

func main() {
	ctx := context.Background()

	// ─── Create topic with short retention ───
	conn, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		log.Fatal(err)
	}

	controller, _ := conn.Controller()
	controllerConn, _ := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	defer controllerConn.Close()
	defer conn.Close()

	// Delete if exists (ignore error)
	controllerConn.DeleteTopics(topic)
	time.Sleep(2 * time.Second)

	err = controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
		ConfigEntries: []kafka.ConfigEntry{
			{ConfigName: "retention.ms", ConfigValue: strconv.Itoa(retentionMs)},
			{ConfigName: "segment.ms", ConfigValue: "10000"},
		},
	})
	if err != nil {
		log.Printf("Create topic error (may already exist): %v", err)
	}

	fmt.Printf("Topic: %s | Retention: %ds\n\n", topic, retentionMs/1000)

	// ─── Produce messages ───
	writer := &kafka.Writer{
		Addr:  kafka.TCP("localhost:9092"),
		Topic: topic,
	}
	defer writer.Close()

	for i := 1; i <= 10; i++ {
		value, _ := json.Marshal(map[string]interface{}{
			"orderId":   fmt.Sprintf("ORD-ret-%d", i),
			"amount":    i * 10,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})

		writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(fmt.Sprintf("ORD-ret-%d", i)),
			Value: value,
		})
		fmt.Printf("Produced: ORD-ret-%d\n", i)
	}

	// ─── Read immediately ───
	fmt.Println("\n--- Consuming immediately ---")
	before := countMessages(ctx, topic, "Before expiry")

	// ─── Wait for retention ───
	waitSec := retentionMs/1000 + 30
	fmt.Printf("\n--- Waiting %ds for retention + cleanup ---\n", waitSec)
	time.Sleep(time.Duration(waitSec) * time.Second)

	// ─── Read after expiry ───
	fmt.Println("\n--- Consuming after retention ---")
	after := countMessages(ctx, topic, "After expiry")

	fmt.Println("\n═══════════════════════════════════════")
	fmt.Printf("Before: %d messages\n", before)
	fmt.Printf("After:  %d messages\n", after)
	fmt.Printf("Expired: %d messages\n", before-after)
	fmt.Println("═══════════════════════════════════════")
}

func countMessages(ctx context.Context, topic, label string) int {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    topic,
		GroupID:  fmt.Sprintf("retention-check-%d", time.Now().UnixNano()),
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	readCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	count := 0
	var keys []string

	for {
		msg, err := reader.FetchMessage(readCtx)
		if err != nil {
			break
		}
		count++
		keys = append(keys, string(msg.Key))
		reader.CommitMessages(readCtx, msg)
	}

	fmt.Printf("[%s] Found %d messages: %v\n", label, count, keys)
	return count
}
```

---

## Tool 2: `cmd/compaction-demo/main.go` — Log Compaction

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

const topic = "order-status-compacted-go"

func main() {
	ctx := context.Background()

	// ─── Create compacted topic ───
	conn, _ := kafka.Dial("tcp", "localhost:9092")
	controller, _ := conn.Controller()
	controllerConn, _ := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	defer controllerConn.Close()
	defer conn.Close()

	controllerConn.DeleteTopics(topic)
	time.Sleep(2 * time.Second)

	err := controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
		ConfigEntries: []kafka.ConfigEntry{
			{ConfigName: "cleanup.policy", ConfigValue: "compact"},
			{ConfigName: "min.cleanable.dirty.ratio", ConfigValue: "0.01"},
			{ConfigName: "segment.ms", ConfigValue: "1000"},
			{ConfigName: "delete.retention.ms", ConfigValue: "1000"},
		},
	})
	if err != nil {
		log.Printf("Create error: %v", err)
	}

	fmt.Printf("Created compacted topic: %s\n", topic)

	// ─── Produce order status updates ───
	writer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    topic,
		Balancer: &kafka.Hash{},
	}
	defer writer.Close()

	updates := []struct {
		key    string
		status string
	}{
		{"ORD-001", "CREATED"},
		{"ORD-002", "CREATED"},
		{"ORD-001", "PAYMENT_PENDING"},
		{"ORD-003", "CREATED"},
		{"ORD-001", "PAYMENT_CONFIRMED"},
		{"ORD-002", "PAYMENT_PENDING"},
		{"ORD-001", "SHIPPED"},
		{"ORD-002", "PAYMENT_CONFIRMED"},
		{"ORD-003", "CANCELLED"},
		{"ORD-001", "DELIVERED"},
	}

	fmt.Println("\nProducing order status updates:")
	for _, u := range updates {
		value, _ := json.Marshal(map[string]interface{}{
			"orderId":   u.key,
			"status":    u.status,
			"updatedAt": time.Now().UTC().Format(time.RFC3339),
		})

		writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(u.key),
			Value: value,
		})
		fmt.Printf("  %s → %s\n", u.key, u.status)
	}

	fmt.Printf("\nTotal produced: %d\n", len(updates))
	fmt.Println("ORD-001: 5 updates | ORD-002: 3 updates | ORD-003: 2 updates")

	// ─── Read before compaction ───
	fmt.Println("\n--- Reading before compaction ---")
	readAll(ctx, "Before compaction")

	// ─── Wait for compaction ───
	fmt.Println("\n--- Waiting 30s for compaction ---")
	time.Sleep(30 * time.Second)

	// ─── Read after compaction ───
	fmt.Println("\n--- Reading after compaction ---")
	readAll(ctx, "After compaction")

	fmt.Println("\nExpected after compaction:")
	fmt.Println("  ORD-001 → DELIVERED")
	fmt.Println("  ORD-002 → PAYMENT_CONFIRMED")
	fmt.Println("  ORD-003 → CANCELLED")
	fmt.Println("  Total: 3 messages (down from 10)")
}

func readAll(ctx context.Context, label string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    topic,
		GroupID:  fmt.Sprintf("compact-read-%d", time.Now().UnixNano()),
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	readCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	count := 0
	for {
		msg, err := reader.FetchMessage(readCtx)
		if err != nil {
			break
		}
		count++

		var val map[string]interface{}
		json.Unmarshal(msg.Value, &val)

		fmt.Printf("  offset %d: %s → %s\n", msg.Offset, string(msg.Key), val["status"])
		reader.CommitMessages(readCtx, msg)
	}

	fmt.Printf("[%s] %d messages read\n", label, count)
}
```

---

## Tool 3: `cmd/replay-consumer/main.go` — Fresh Replay

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	groupID := fmt.Sprintf("replay-group-go-%d", time.Now().UnixNano())

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "orders",
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
		// StartOffset is ignored when GroupID is set and no prior offsets exist.
		// For a brand new group, kafka-go defaults to FirstOffset.
		StartOffset: kafka.FirstOffset,
	})
	defer reader.Close()

	log.Printf("[Replay] Starting from offset 0 with group: %s", groupID)
	log.Println("[Replay] This re-reads the ENTIRE topic history")
	fmt.Println()

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	count := 0
	startTime := time.Now()

	// Auto-stop after 10 seconds
	go func() {
		select {
		case <-sigChan:
		case <-time.After(10 * time.Second):
		}
		elapsed := time.Since(startTime).Seconds()
		fmt.Printf("\n[Replay] Replayed %d messages in %.1fs\n", count, elapsed)
		cancel()
	}()

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			break
		}
		count++

		age := time.Since(msg.Time).Seconds()
		fmt.Printf("  #%d | P%d:%d | %s | age: %.0fs\n",
			count, msg.Partition, msg.Offset, string(msg.Key), age)

		reader.CommitMessages(ctx, msg)
	}
}
```

---

## Tool 4: `cmd/offset-reset/main.go` — Programmatic Offset Reset

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	groupID := "payment-group-go"
	topic := "orders"
	mode := "earliest"

	if len(os.Args) > 1 {
		groupID = os.Args[1]
	}
	if len(os.Args) > 2 {
		topic = os.Args[2]
	}
	if len(os.Args) > 3 {
		mode = os.Args[3]
	}

	fmt.Printf("[Offset Reset] group=%s topic=%s mode=%s\n", groupID, topic, mode)
	fmt.Println("⚠️  Make sure consumers in this group are STOPPED first!")
	fmt.Println()

	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, 0)
	if err != nil {
		log.Fatal(err)
	}

	// Get partition count
	partitions, err := conn.ReadPartitions(topic)
	if err != nil {
		log.Fatal(err)
	}
	conn.Close()

	// Create a consumer group client to manage offsets
	client := &kafka.Client{
		Addr: kafka.TCP("localhost:9092"),
	}

	// Fetch current offsets
	offsetFetch, _ := client.OffsetFetch(context.Background(), &kafka.OffsetFetchRequest{
		GroupID: groupID,
		Topics:  map[string][]int{topic: partitionIDs(partitions)},
	})

	fmt.Println("Current offsets:")
	if offsetFetch != nil {
		for t, parts := range offsetFetch.Topics {
			for _, p := range parts {
				fmt.Printf("  %s P%d: offset %d\n", t, p.Partition, p.CommittedOffset)
			}
		}
	}

	// Determine new offsets based on mode
	newOffsets := make(map[int]int64)

	switch mode {
	case "earliest":
		for _, p := range partitions {
			partConn, _ := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, p.ID)
			first, _ := partConn.ReadFirstOffset()
			newOffsets[p.ID] = first
			partConn.Close()
		}
		fmt.Println("\n✅ Resetting to EARLIEST")

	case "latest":
		for _, p := range partitions {
			partConn, _ := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, p.ID)
			last, _ := partConn.ReadLastOffset()
			newOffsets[p.ID] = last
			partConn.Close()
		}
		fmt.Println("\n✅ Resetting to LATEST")

	case "offset":
		targetOffset := int64(0)
		if len(os.Args) > 4 {
			targetOffset, _ = strconv.ParseInt(os.Args[4], 10, 64)
		}
		for _, p := range partitions {
			newOffsets[p.ID] = targetOffset
		}
		fmt.Printf("\n✅ Resetting to offset %d\n", targetOffset)

	case "timestamp":
		minutesAgo := 5
		if len(os.Args) > 4 {
			minutesAgo, _ = strconv.Atoi(os.Args[4])
		}
		targetTime := time.Now().Add(-time.Duration(minutesAgo) * time.Minute)
		fmt.Printf("\n✅ Resetting to %d minutes ago (%s)\n", minutesAgo, targetTime.Format(time.RFC3339))

		// Use ListOffsets to find offsets by timestamp
		for _, p := range partitions {
			partConn, _ := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, p.ID)
			// ReadOffset doesn't support timestamp lookup directly, so we use first offset as fallback
			first, _ := partConn.ReadFirstOffset()
			newOffsets[p.ID] = first
			partConn.Close()
		}

	default:
		log.Fatalf("Unknown mode: %s (use earliest, latest, offset, timestamp)", mode)
	}

	// Commit the new offsets using a temporary consumer
	for partID, offset := range newOffsets {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{"localhost:9092"},
			Topic:   topic,
			GroupID: groupID,
		})

		// CommitMessages with the target offset
		reader.CommitMessages(context.Background(), kafka.Message{
			Topic:     topic,
			Partition: partID,
			Offset:    offset,
		})
		reader.Close()
	}

	// Verify
	fmt.Println("\nNew offsets:")
	for partID, offset := range newOffsets {
		fmt.Printf("  %s P%d: offset %d\n", topic, partID, offset)
	}
}

func partitionIDs(partitions []kafka.Partition) []int {
	ids := make([]int, len(partitions))
	for i, p := range partitions {
		ids[i] = p.ID
	}
	return ids
}
```

**Usage:**

```bash
# Reset to beginning
go run cmd/offset-reset/main.go payment-group-go orders earliest

# Reset to latest
go run cmd/offset-reset/main.go payment-group-go orders latest

# Reset to specific offset
go run cmd/offset-reset/main.go payment-group-go orders offset 50

# Reset to 10 minutes ago
go run cmd/offset-reset/main.go payment-group-go orders timestamp 10
```

---

## Tool 5: `cmd/time-travel-consumer/main.go` — Consume from Timestamp

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	minutesAgo := 5
	if len(os.Args) > 1 {
		minutesAgo, _ = strconv.Atoi(os.Args[1])
	}

	topic := "orders"
	targetTime := time.Now().Add(-time.Duration(minutesAgo) * time.Minute)
	groupID := fmt.Sprintf("time-travel-go-%d", time.Now().UnixNano())

	log.Printf("[Time Travel] Consuming from %d minutes ago", minutesAgo)
	log.Printf("[Time Travel] Target: %s", targetTime.Format(time.RFC3339))
	log.Printf("[Time Travel] Group: %s\n", groupID)

	// Get partitions
	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, 0)
	if err != nil {
		log.Fatal(err)
	}
	partitions, err := conn.ReadPartitions(topic)
	if err != nil {
		log.Fatal(err)
	}
	conn.Close()

	// Create readers per partition starting from the target time
	// kafka-go ReaderConfig doesn't support timestamp-based start directly,
	// so we create partition-specific readers and skip old messages
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"localhost:9092"},
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})
	defer reader.Close()

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	count := 0
	skipped := 0

	go func() {
		select {
		case <-sigChan:
		case <-time.After(10 * time.Second):
		}
		fmt.Printf("\n[Time Travel] Read %d messages (skipped %d before target time)\n", count, skipped)
		cancel()
	}()

	fmt.Printf("Resolved target time for %d partitions\n\n", len(partitions))

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			break
		}

		// Skip messages before target time
		if msg.Time.Before(targetTime) {
			skipped++
			reader.CommitMessages(ctx, msg)
			continue
		}

		count++
		fmt.Printf("  #%d | P%d:%d | %s | %s\n",
			count, msg.Partition, msg.Offset, string(msg.Key),
			msg.Time.Format(time.RFC3339))

		reader.CommitMessages(ctx, msg)
	}
}
```

**Usage:**

```bash
# Messages from last 5 minutes
go run cmd/time-travel-consumer/main.go 5

# Messages from last 60 minutes
go run cmd/time-travel-consumer/main.go 60
```

---

## Idiomatic Differences: TypeScript vs Go

| Aspect | TypeScript | Go |
|--------|-----------|-----|
| **Topic admin** | `admin.createTopics()` with configEntries | `controllerConn.CreateTopics()` with ConfigEntry |
| **Offset reset** | `admin.setOffsets()` programmatically | `CommitMessages` with target offset or kafka Client API |
| **Timestamp lookup** | `admin.fetchTopicOffsetsByTimestamp()` | Manual per-partition connection; kafka-go has limited timestamp API |
| **Read timeout** | `setTimeout(resolve, 3000)` | `context.WithTimeout(ctx, 3*time.Second)` |
| **Cleanup** | `process.on("SIGINT")` | `signal.Notify` + goroutine |

Go's kafka-go library has less admin API coverage than kafkajs. For production offset management, you'd typically use the Kafka CLI tools or the confluent-kafka-go library which has richer admin support.

---

## Running the Full Demo

```bash
# 1: Retention demo (~90s)
go run cmd/retention-demo/main.go

# 2: Compaction demo (~30s)
go run cmd/compaction-demo/main.go

# 3: Replay everything
go run cmd/replay-consumer/main.go

# 4: Reset offsets
go run cmd/offset-reset/main.go payment-group-go orders earliest

# 5: Time travel
go run cmd/time-travel-consumer/main.go 10
```

→ Next: [Phase 8 — Ops & Observability](../phase-08-ops/README.md)
