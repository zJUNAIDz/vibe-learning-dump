# Phase 0 — Go Implementation

## The Naive Synchronous Pipeline

Same architecture as the TypeScript version: four services, synchronous HTTP calls, cascading failures.

### Project Setup

```bash
mkdir -p phase-00-pre-kafka/go
cd phase-00-pre-kafka/go
go mod init order-pipeline-naive
```

### File Structure

```
go/
├── cmd/
│   ├── order-service/main.go
│   ├── payment-service/main.go
│   ├── notification-service/main.go
│   └── inventory-service/main.go
├── go.mod
└── go.sum
```

Go's standard library is sufficient for this. No frameworks needed.

---

### `cmd/inventory-service/main.go`

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type DecrementRequest struct {
	OrderID  string `json:"orderId"`
	ItemID   string `json:"itemId"`
	Quantity int    `json:"quantity"`
}

type DecrementResponse struct {
	Success   bool `json:"success"`
	Remaining int  `json:"remaining"`
}

var (
	stock = map[string]int{
		"ITEM-001": 100,
		"ITEM-002": 50,
		"ITEM-003": 200,
	}
	mu sync.Mutex
)

func handleDecrement(w http.ResponseWriter, r *http.Request) {
	var req DecrementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	log.Printf("[Inventory] Decrementing %dx %s for order %s", req.Quantity, req.ItemID, req.OrderID)

	// Simulate slow database (uncomment to see cascading latency)
	// time.Sleep(3 * time.Second)

	mu.Lock()
	defer mu.Unlock()

	remaining, exists := stock[req.ItemID]
	if !exists || remaining < req.Quantity {
		log.Printf("[Inventory] ❌ Insufficient stock for %s", req.ItemID)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Insufficient stock"})
		return
	}

	stock[req.ItemID] -= req.Quantity
	remaining = stock[req.ItemID]
	log.Printf("[Inventory] ✅ Stock updated. %s remaining: %d", req.ItemID, remaining)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DecrementResponse{Success: true, Remaining: remaining})
}

func main() {
	http.HandleFunc("/inventory/decrement", handleDecrement)
	log.Println("[Inventory Service] listening on :3004")
	log.Fatal(http.ListenAndServe(":3004", nil))
}
```

---

### `cmd/notification-service/main.go`

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type NotifyRequest struct {
	OrderID  string  `json:"orderId"`
	UserID   string  `json:"userId"`
	ItemID   string  `json:"itemId"`
	Quantity int     `json:"quantity"`
	Amount   float64 `json:"amount"`
}

func handleNotify(w http.ResponseWriter, r *http.Request) {
	var req NotifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	log.Printf("[Notification] Sending confirmation for order %s to user %s", req.OrderID, req.UserID)
	log.Printf("[Notification] 📧 Email sent: \"Your order %s for $%.2f is confirmed.\"", req.OrderID, req.Amount)

	// Call Inventory Service
	inventoryPayload, _ := json.Marshal(map[string]interface{}{
		"orderId":  req.OrderID,
		"itemId":   req.ItemID,
		"quantity": req.Quantity,
	})

	resp, err := http.Post(
		"http://localhost:3004/inventory/decrement",
		"application/json",
		bytes.NewReader(inventoryPayload),
	)
	if err != nil {
		log.Printf("[Notification] ❌ Inventory Service unreachable: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Inventory service unreachable"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[Notification] ❌ Inventory update failed: %s", string(body))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Inventory update failed"})
		return
	}

	log.Printf("[Notification] ✅ Inventory updated for order %s", req.OrderID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func main() {
	http.HandleFunc("/notify", handleNotify)
	log.Println("[Notification Service] listening on :3003")
	log.Fatal(http.ListenAndServe(":3003", nil))
}
```

---

### `cmd/payment-service/main.go`

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type PaymentRequest struct {
	OrderID  string  `json:"orderId"`
	UserID   string  `json:"userId"`
	ItemID   string  `json:"itemId"`
	Quantity int     `json:"quantity"`
	Amount   float64 `json:"amount"`
}

func handlePayment(w http.ResponseWriter, r *http.Request) {
	var req PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	log.Printf("[Payment] Processing payment for order %s: $%.2f", req.OrderID, req.Amount)

	// Simulate processing delay
	time.Sleep(500 * time.Millisecond)

	// Simulate ~10% failure rate
	if rand.Float64() < 0.1 {
		log.Printf("[Payment] ❌ Payment declined for order %s", req.OrderID)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Payment declined"})
		return
	}

	log.Printf("[Payment] 💳 Charged $%.2f for order %s", req.Amount, req.OrderID)

	// Call Notification Service
	notifyPayload, _ := json.Marshal(req)

	resp, err := http.Post(
		"http://localhost:3003/notify",
		"application/json",
		bytes.NewReader(notifyPayload),
	)
	if err != nil {
		log.Printf("[Payment] ❌ Notification Service unreachable: %v", err)
		log.Printf("[Payment] ⚠️ Payment charged but notification failed for order %s", req.OrderID)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Notification service unreachable"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[Payment] ⚠️ Payment charged but post-processing failed: %s", string(body))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Post-payment processing failed"})
		return
	}

	log.Printf("[Payment] ✅ Full pipeline complete for order %s", req.OrderID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "orderId": req.OrderID})
}

func main() {
	http.HandleFunc("/payments", handlePayment)
	log.Println("[Payment Service] listening on :3002")
	log.Fatal(http.ListenAndServe(":3002", nil))
}
```

---

### `cmd/order-service/main.go`

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Order struct {
	OrderID   string  `json:"orderId"`
	UserID    string  `json:"userId"`
	ItemID    string  `json:"itemId"`
	Quantity  int     `json:"quantity"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"createdAt"`
}

type OrderRequest struct {
	UserID   string  `json:"userId"`
	ItemID   string  `json:"itemId"`
	Quantity int     `json:"quantity"`
	Amount   float64 `json:"amount"`
}

var (
	orders []Order
	mu     sync.Mutex
)

func handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var req OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	orderID := fmt.Sprintf("ORD-%s", uuid.New().String()[:8])

	log.Println()
	log.Println("============================================================")
	log.Printf("[Order] New order %s from user %s", orderID, req.UserID)
	log.Printf("[Order] Item: %s, Qty: %d, Amount: $%.2f", req.ItemID, req.Quantity, req.Amount)

	order := Order{
		OrderID:   orderID,
		UserID:    req.UserID,
		ItemID:    req.ItemID,
		Quantity:  req.Quantity,
		Amount:    req.Amount,
		Status:    "pending",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	mu.Lock()
	orders = append(orders, order)
	mu.Unlock()

	log.Println("[Order] Saved to DB. Calling Payment Service...")

	// Call Payment Service synchronously
	start := time.Now()

	paymentPayload, _ := json.Marshal(map[string]interface{}{
		"orderId":  orderID,
		"userId":   req.UserID,
		"itemId":   req.ItemID,
		"quantity": req.Quantity,
		"amount":   req.Amount,
	})

	resp, err := http.Post(
		"http://localhost:3002/payments",
		"application/json",
		bytes.NewReader(paymentPayload),
	)
	elapsed := time.Since(start)

	if err != nil {
		order.Status = "failed"
		log.Printf("[Order] ❌ Payment Service unreachable after %v: %v", elapsed, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Payment service unreachable",
			"orderId": orderID,
			"elapsed": elapsed.String(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		order.Status = "failed"
		log.Printf("[Order] ❌ Order %s failed after %v: %s", orderID, elapsed, string(body))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Order processing failed",
			"orderId": orderID,
			"elapsed": elapsed.String(),
		})
		return
	}

	order.Status = "completed"
	log.Printf("[Order] ✅ Order %s completed in %v", orderID, elapsed)
	log.Println("============================================================")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"orderId": orderID,
		"status":  "completed",
		"elapsed": elapsed.String(),
	})
}

func handleListOrders(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func main() {
	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreateOrder(w, r)
		case http.MethodGet:
			handleListOrders(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("[Order Service] listening on :3001")
	log.Println()
	log.Println("Try: curl -X POST http://localhost:3001/orders \\")
	log.Println("  -H \"Content-Type: application/json\" \\")
	log.Println("  -d '{\"userId\":\"user-1\",\"itemId\":\"ITEM-001\",\"quantity\":2,\"amount\":49.99}'")
	log.Fatal(http.ListenAndServe(":3001", nil))
}
```

> **Note:** `go get github.com/google/uuid` for the UUID generation — or replace with `fmt.Sprintf("ORD-%d", time.Now().UnixNano())` if you want zero dependencies.

---

## Running It

```bash
# Terminal 1
go run cmd/inventory-service/main.go

# Terminal 2
go run cmd/notification-service/main.go

# Terminal 3
go run cmd/payment-service/main.go

# Terminal 4
go run cmd/order-service/main.go
```

### Place an Order

```bash
curl -X POST http://localhost:3001/orders \
  -H "Content-Type: application/json" \
  -d '{"userId":"user-1","itemId":"ITEM-001","quantity":2,"amount":49.99}'
```

---

## Idiomatic Differences: TypeScript vs Go

| Aspect | TypeScript | Go |
|--------|-----------|-----|
| **HTTP server** | Express (third-party) | `net/http` (stdlib) |
| **JSON parsing** | `req.body` automatic with middleware | Manual `json.NewDecoder` |
| **Async model** | `async/await` — single-threaded event loop | Goroutines — true concurrency with OS threads |
| **Error handling** | `try/catch` | Explicit `if err != nil` |
| **Concurrency safety** | Not needed (single-threaded) | `sync.Mutex` required for shared state |
| **Type definitions** | Inline or interfaces | Explicit structs with JSON tags |

The Go version is more verbose but more explicit. Every error is handled. Concurrency is explicit. There's no hidden middleware magic.

In a real system, the Go version's explicitness is an advantage — you can trace exactly what happens on every code path.

---

## What's The Same in Both

Both versions have the same fundamental problem: **temporal coupling**.

The user's HTTP request blocks until the entire chain completes. If any service is slow or down, the order fails — even if some services already did their work (like charging the card).

This is the problem Kafka solves. Not by being fast. By decoupling services in **time**.

→ Next: [Phase 1 — Kafka as a Log](../phase-01-log-basics/README.md)
