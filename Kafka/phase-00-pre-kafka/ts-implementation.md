# Phase 0 — TypeScript Implementation

## The Naive Synchronous Pipeline

We simulate four services running as HTTP servers. Each service calls the next one synchronously.

### Project Setup

```bash
mkdir -p phase-00-pre-kafka/ts
cd phase-00-pre-kafka/ts
npm init -y
npm install express
npm install -D typescript @types/node @types/express ts-node
npx tsc --init
```

Update `tsconfig.json`:
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "commonjs",
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "resolveJsonModule": true
  }
}
```

### File Structure

```
ts/
├── src/
│   ├── order-service.ts
│   ├── payment-service.ts
│   ├── notification-service.ts
│   └── inventory-service.ts
├── package.json
└── tsconfig.json
```

---

### `src/inventory-service.ts` — The Last in the Chain

```typescript
import express from "express";

const app = express();
app.use(express.json());

// In-memory stock database
const stock: Record<string, number> = {
  "ITEM-001": 100,
  "ITEM-002": 50,
  "ITEM-003": 200,
};

app.post("/inventory/decrement", (req, res) => {
  const { orderId, itemId, quantity } = req.body;

  console.log(`[Inventory] Decrementing ${quantity}x ${itemId} for order ${orderId}`);

  // Simulate slow database (uncomment to see cascading latency)
  // await new Promise(resolve => setTimeout(resolve, 3000));

  if (!stock[itemId] || stock[itemId] < quantity) {
    console.log(`[Inventory] ❌ Insufficient stock for ${itemId}`);
    res.status(400).json({ error: "Insufficient stock" });
    return;
  }

  stock[itemId] -= quantity;
  console.log(`[Inventory] ✅ Stock updated. ${itemId} remaining: ${stock[itemId]}`);

  res.json({ success: true, remaining: stock[itemId] });
});

app.listen(3004, () => {
  console.log("[Inventory Service] listening on :3004");
});
```

---

### `src/notification-service.ts` — Calls Inventory

```typescript
import express from "express";

const app = express();
app.use(express.json());

app.post("/notify", async (req, res) => {
  const { orderId, userId, itemId, quantity, amount } = req.body;

  console.log(`[Notification] Sending confirmation for order ${orderId} to user ${userId}`);

  // Simulate sending email
  console.log(`[Notification] 📧 Email sent: "Your order ${orderId} for $${amount} is confirmed."`);

  // Now call Inventory Service (synchronous chain continues)
  try {
    const inventoryRes = await fetch("http://localhost:3004/inventory/decrement", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ orderId, itemId, quantity }),
    });

    if (!inventoryRes.ok) {
      const error = await inventoryRes.json();
      console.log(`[Notification] ❌ Inventory update failed: ${JSON.stringify(error)}`);
      res.status(500).json({ error: "Inventory update failed" });
      return;
    }

    console.log(`[Notification] ✅ Inventory updated for order ${orderId}`);
    res.json({ success: true });
  } catch (err) {
    console.log(`[Notification] ❌ Inventory Service unreachable: ${(err as Error).message}`);
    res.status(500).json({ error: "Inventory service unreachable" });
  }
});

app.listen(3003, () => {
  console.log("[Notification Service] listening on :3003");
});
```

---

### `src/payment-service.ts` — Calls Notification

```typescript
import express from "express";

const app = express();
app.use(express.json());

// Simulate payment processing
function processPayment(orderId: string, amount: number): boolean {
  // Simulate ~10% failure rate
  if (Math.random() < 0.1) {
    console.log(`[Payment] ❌ Payment declined for order ${orderId}`);
    return false;
  }
  console.log(`[Payment] 💳 Charged $${amount} for order ${orderId}`);
  return true;
}

app.post("/payments", async (req, res) => {
  const { orderId, userId, itemId, quantity, amount } = req.body;

  console.log(`[Payment] Processing payment for order ${orderId}: $${amount}`);

  // Simulate payment processing delay
  await new Promise(resolve => setTimeout(resolve, 500));

  const success = processPayment(orderId, amount);
  if (!success) {
    res.status(400).json({ error: "Payment declined" });
    return;
  }

  // Now call Notification Service (synchronous chain continues)
  try {
    const notifyRes = await fetch("http://localhost:3003/notify", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ orderId, userId, itemId, quantity, amount }),
    });

    if (!notifyRes.ok) {
      // Payment already charged! But notification failed.
      // What do we do? Refund? Ignore? This is the problem.
      console.log(`[Payment] ⚠️ Payment charged but notification failed for order ${orderId}`);
      res.status(500).json({ error: "Post-payment processing failed" });
      return;
    }

    console.log(`[Payment] ✅ Full pipeline complete for order ${orderId}`);
    res.json({ success: true, orderId });
  } catch (err) {
    console.log(`[Payment] ❌ Notification Service unreachable: ${(err as Error).message}`);
    // Payment was charged but downstream failed. Inconsistent state.
    res.status(500).json({ error: "Notification service unreachable" });
  }
});

app.listen(3002, () => {
  console.log("[Payment Service] listening on :3002");
});
```

---

### `src/order-service.ts` — The Entry Point

```typescript
import express from "express";
import crypto from "crypto";

const app = express();
app.use(express.json());

// In-memory order store
const orders: Array<{
  orderId: string;
  userId: string;
  itemId: string;
  quantity: number;
  amount: number;
  status: string;
  createdAt: string;
}> = [];

app.post("/orders", async (req, res) => {
  const { userId, itemId, quantity, amount } = req.body;
  const orderId = `ORD-${crypto.randomUUID().slice(0, 8)}`;

  console.log(`\n${"=".repeat(60)}`);
  console.log(`[Order] New order ${orderId} from user ${userId}`);
  console.log(`[Order] Item: ${itemId}, Qty: ${quantity}, Amount: $${amount}`);

  // Save to "database"
  const order = {
    orderId,
    userId,
    itemId,
    quantity,
    amount,
    status: "pending",
    createdAt: new Date().toISOString(),
  };
  orders.push(order);

  console.log(`[Order] Saved to DB. Calling Payment Service...`);

  // Call Payment Service synchronously
  const start = Date.now();
  try {
    const paymentRes = await fetch("http://localhost:3002/payments", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ orderId, userId, itemId, quantity, amount }),
    });

    const elapsed = Date.now() - start;

    if (!paymentRes.ok) {
      const error = await paymentRes.json();
      order.status = "failed";
      console.log(`[Order] ❌ Order ${orderId} failed after ${elapsed}ms: ${JSON.stringify(error)}`);
      res.status(500).json({ error: "Order processing failed", orderId, elapsed });
      return;
    }

    order.status = "completed";
    console.log(`[Order] ✅ Order ${orderId} completed in ${elapsed}ms`);
    console.log(`${"=".repeat(60)}\n`);

    res.status(201).json({ orderId, status: "completed", elapsed });
  } catch (err) {
    const elapsed = Date.now() - start;
    order.status = "failed";
    console.log(`[Order] ❌ Payment Service unreachable after ${elapsed}ms: ${(err as Error).message}`);
    res.status(500).json({ error: "Payment service unreachable", orderId, elapsed });
  }
});

// Check all orders
app.get("/orders", (_req, res) => {
  res.json(orders);
});

app.listen(3001, () => {
  console.log("[Order Service] listening on :3001");
  console.log("\nTry: curl -X POST http://localhost:3001/orders \\");
  console.log('  -H "Content-Type: application/json" \\');
  console.log('  -d \'{"userId":"user-1","itemId":"ITEM-001","quantity":2,"amount":49.99}\'');
});
```

---

## Running It

Open **four terminals**:

```bash
# Terminal 1
npx ts-node src/inventory-service.ts

# Terminal 2
npx ts-node src/notification-service.ts

# Terminal 3
npx ts-node src/payment-service.ts

# Terminal 4
npx ts-node src/order-service.ts
```

### Place an Order

```bash
curl -X POST http://localhost:3001/orders \
  -H "Content-Type: application/json" \
  -d '{"userId":"user-1","itemId":"ITEM-001","quantity":2,"amount":49.99}'
```

Watch the logs cascade through all four terminals. The user's request blocks until all four services finish.

---

## Break It

### Failure 1: Kill Notification Service

1. Stop the notification service (Ctrl+C in Terminal 2)
2. Place another order
3. Watch: payment succeeds, but order returns 500
4. The user's card was charged. No email was sent. No inventory update.

### Failure 2: Simulate Slow Inventory

1. Uncomment the `setTimeout` in `inventory-service.ts`
2. Restart the inventory service
3. Place an order
4. Watch the response time jump from ~500ms to ~3500ms
5. Every service in the chain waits

### Failure 3: Spam Orders

```bash
# Send 20 orders rapidly
for i in $(seq 1 20); do
  curl -s -X POST http://localhost:3001/orders \
    -H "Content-Type: application/json" \
    -d "{\"userId\":\"user-$i\",\"itemId\":\"ITEM-001\",\"quantity\":1,\"amount\":9.99}" &
done
wait
```

Watch the chaos in your terminal logs. Some succeed. Some fail. No ordering guarantees. No retry logic.

---

## What You Should Feel

After running this:

- **Latency compounds** — The user waits for every service in the chain
- **One failure kills everything** — Even if payment succeeds, a downstream failure means the user sees an error
- **Inconsistent state** — Payment charged but inventory not updated? That's real.
- **No replay** — If notification was down, those events are gone. You have to manually reconcile.
- **No independent scaling** — You can't scale the notification service independently of the payment service

This is the pain. Phase 1 introduces the solution.

→ Next: [Phase 0 — Go Implementation](go-implementation.md)
→ Skip to: [Phase 1 — Kafka as a Log](../phase-01-log-basics/README.md)
