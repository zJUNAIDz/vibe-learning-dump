# Phase 6 — TypeScript Implementation

## Setup

```bash
cd ts/
# kafkajs already installed from Phase 1
```

### File Structure

```
ts/
├── src/
│   ├── schema-v1-producer.ts
│   ├── schema-v2-producer.ts
│   ├── multi-version-consumer.ts
│   └── schema-validator.ts
├── package.json
└── tsconfig.json
```

---

## Schema Definitions

Before writing producers/consumers, define the schemas:

### `src/schemas.ts`

```typescript
// ─── Envelope (stable across all versions) ───
export interface EventEnvelope<T = unknown> {
  version: number;
  eventType: string;
  timestamp: string;
  source: string;
  payload: T;
}

// ─── V1 Payload ───
export interface OrderPayloadV1 {
  orderId: string;
  userId: string;
  amount: number;
}

// ─── V2 Payload (adds currency, items) ───
export interface OrderPayloadV2 {
  orderId: string;
  userId: string;
  amount: number;
  currency: string;      // new in v2
  items?: LineItem[];     // new in v2, optional
}

export interface LineItem {
  sku: string;
  name: string;
  quantity: number;
  price: number;
}

// ─── Normalized internal representation ───
// The consumer converts ANY version into this shape
export interface NormalizedOrder {
  orderId: string;
  userId: string;
  amount: number;
  currency: string;      // default: "USD"
  items: LineItem[];      // default: []
  _originalVersion: number;
}

// ─── Schema validation ───
export function validateV1(payload: unknown): payload is OrderPayloadV1 {
  const p = payload as Record<string, unknown>;
  return (
    typeof p.orderId === "string" &&
    typeof p.userId === "string" &&
    typeof p.amount === "number" &&
    p.amount > 0
  );
}

export function validateV2(payload: unknown): payload is OrderPayloadV2 {
  const p = payload as Record<string, unknown>;
  return (
    validateV1(payload) &&
    typeof p.currency === "string" &&
    p.currency.length === 3
  );
}

// ─── Normalization ───
export function normalizeOrder(envelope: EventEnvelope): NormalizedOrder {
  switch (envelope.version) {
    case 1: {
      const p = envelope.payload as OrderPayloadV1;
      return {
        orderId: p.orderId,
        userId: p.userId,
        amount: p.amount,
        currency: "USD",   // default for v1
        items: [],          // not available in v1
        _originalVersion: 1,
      };
    }
    case 2: {
      const p = envelope.payload as OrderPayloadV2;
      return {
        orderId: p.orderId,
        userId: p.userId,
        amount: p.amount,
        currency: p.currency,
        items: p.items ?? [],
        _originalVersion: 2,
      };
    }
    default:
      throw new Error(`Unknown schema version: ${envelope.version}`);
  }
}
```

---

## Tool 1: `schema-v1-producer.ts` — Legacy Producer

Sends events in v1 format (no envelope version, no currency). Simulates what the producer looked like before the schema evolution.

```typescript
import { Kafka } from "kafkajs";
import { EventEnvelope, OrderPayloadV1 } from "./schemas";

const kafka = new Kafka({ clientId: "schema-v1-producer", brokers: ["localhost:9092"] });
const producer = kafka.producer();

async function main(): Promise<void> {
  await producer.connect();

  const orders: EventEnvelope<OrderPayloadV1>[] = [
    {
      version: 1,
      eventType: "ORDER_CREATED",
      timestamp: new Date().toISOString(),
      source: "order-service-v1",
      payload: {
        orderId: "ORD-v1-001",
        userId: "user-10",
        amount: 49.99,
      },
    },
    {
      version: 1,
      eventType: "ORDER_CREATED",
      timestamp: new Date().toISOString(),
      source: "order-service-v1",
      payload: {
        orderId: "ORD-v1-002",
        userId: "user-11",
        amount: 129.50,
      },
    },
    {
      version: 1,
      eventType: "ORDER_CREATED",
      timestamp: new Date().toISOString(),
      source: "order-service-v1",
      payload: {
        orderId: "ORD-v1-003",
        userId: "user-12",
        amount: 9.99,
      },
    },
  ];

  for (const order of orders) {
    await producer.send({
      topic: "orders",
      messages: [
        {
          key: order.payload.orderId,
          value: JSON.stringify(order),
          headers: { "schema-version": "1" },
        },
      ],
    });

    console.log(`Sent v1: ${order.payload.orderId} | $${order.payload.amount}`);
  }

  console.log(`\n[v1 Producer] ${orders.length} messages sent (no currency field)`);
  await producer.disconnect();
}

main().catch(console.error);
```

---

## Tool 2: `schema-v2-producer.ts` — Evolved Producer

Sends events with the v2 envelope: adds `currency` and optional `items`.

```typescript
import { Kafka } from "kafkajs";
import { EventEnvelope, OrderPayloadV2 } from "./schemas";

const kafka = new Kafka({ clientId: "schema-v2-producer", brokers: ["localhost:9092"] });
const producer = kafka.producer();

async function main(): Promise<void> {
  await producer.connect();

  const orders: EventEnvelope<OrderPayloadV2>[] = [
    {
      version: 2,
      eventType: "ORDER_CREATED",
      timestamp: new Date().toISOString(),
      source: "order-service-v2",
      payload: {
        orderId: "ORD-v2-001",
        userId: "user-20",
        amount: 149.99,
        currency: "EUR",
        items: [
          { sku: "LAPTOP-001", name: "USB-C Dock", quantity: 1, price: 149.99 },
        ],
      },
    },
    {
      version: 2,
      eventType: "ORDER_CREATED",
      timestamp: new Date().toISOString(),
      source: "order-service-v2",
      payload: {
        orderId: "ORD-v2-002",
        userId: "user-21",
        amount: 59.98,
        currency: "GBP",
        items: [
          { sku: "CABLE-001", name: "HDMI Cable", quantity: 2, price: 29.99 },
        ],
      },
    },
    {
      version: 2,
      eventType: "ORDER_CREATED",
      timestamp: new Date().toISOString(),
      source: "order-service-v2",
      payload: {
        orderId: "ORD-v2-003",
        userId: "user-22",
        amount: 299.97,
        currency: "USD",
      },
    },
  ];

  for (const order of orders) {
    await producer.send({
      topic: "orders",
      messages: [
        {
          key: order.payload.orderId,
          value: JSON.stringify(order),
          headers: { "schema-version": "2" },
        },
      ],
    });

    const itemCount = order.payload.items?.length ?? 0;
    console.log(
      `Sent v2: ${order.payload.orderId} | $${order.payload.amount} ${order.payload.currency} | ${itemCount} items`
    );
  }

  console.log(`\n[v2 Producer] ${orders.length} messages sent (with currency + items)`);
  await producer.disconnect();
}

main().catch(console.error);
```

---

## Tool 3: `multi-version-consumer.ts` — Handles v1 and v2

The key tool. One consumer that transparently handles both schema versions by normalizing into a common internal representation.

```typescript
import { Kafka, EachMessagePayload } from "kafkajs";
import {
  EventEnvelope,
  normalizeOrder,
  NormalizedOrder,
  validateV1,
  validateV2,
} from "./schemas";

const kafka = new Kafka({ clientId: "multi-version-consumer", brokers: ["localhost:9092"] });
const consumer = kafka.consumer({ groupId: "payment-schema-group" });

// ─── Dead letter producer for unknowns ───
const dlProducer = kafka.producer();

async function sendToDLT(topic: string, key: string, value: string, error: string): Promise<void> {
  await dlProducer.send({
    topic: "orders.dead-letter",
    messages: [
      {
        key,
        value: JSON.stringify({
          originalTopic: topic,
          originalKey: key,
          originalValue: value,
          error,
          errorType: "SchemaError",
          failedAt: new Date().toISOString(),
        }),
      },
    ],
  });
}

async function processOrder(order: NormalizedOrder): Promise<void> {
  // Simulate payment processing on the normalized shape
  console.log(`    💳 Charging $${order.amount} ${order.currency}`);
  if (order.items.length > 0) {
    console.log(`    📦 Items: ${order.items.map((i) => `${i.name} x${i.quantity}`).join(", ")}`);
  }
}

async function handleMessage({ topic, partition, message }: EachMessagePayload): Promise<void> {
  const key = message.key?.toString() ?? "unknown";
  const raw = message.value?.toString() ?? "";
  const versionHeader = message.headers?.["schema-version"]?.toString();

  console.log(`\nP${partition}:${message.offset} | key=${key} | schema-version=${versionHeader ?? "?"}`);

  // Step 1: Parse envelope
  let envelope: EventEnvelope;
  try {
    envelope = JSON.parse(raw) as EventEnvelope;
  } catch {
    console.log(`  ❌ Malformed JSON → DLT`);
    await sendToDLT(topic, key, raw, "Malformed JSON");
    return;
  }

  // Step 2: Check version
  if (typeof envelope.version !== "number") {
    console.log(`  ❌ Missing version field → DLT`);
    await sendToDLT(topic, key, raw, "Missing version field");
    return;
  }

  // Step 3: Validate payload for known versions
  if (envelope.version === 1 && !validateV1(envelope.payload)) {
    console.log(`  ❌ v1 validation failed → DLT`);
    await sendToDLT(topic, key, raw, "v1 payload validation failed");
    return;
  }

  if (envelope.version === 2 && !validateV2(envelope.payload)) {
    console.log(`  ❌ v2 validation failed → DLT`);
    await sendToDLT(topic, key, raw, "v2 payload validation failed");
    return;
  }

  // Step 4: Normalize
  let normalized: NormalizedOrder;
  try {
    normalized = normalizeOrder(envelope);
  } catch (err) {
    console.log(`  ⚠️ Unknown version ${envelope.version} → DLT`);
    await sendToDLT(topic, key, raw, `Unknown schema version: ${envelope.version}`);
    return;
  }

  // Step 5: Process the normalized order
  console.log(
    `  ✅ v${normalized._originalVersion} → normalized | ` +
      `${normalized.orderId} | $${normalized.amount} ${normalized.currency}`
  );

  await processOrder(normalized);
}

async function main(): Promise<void> {
  await dlProducer.connect();
  await consumer.connect();
  await consumer.subscribe({ topic: "orders", fromBeginning: false });

  console.log("[Multi-Version Consumer] Handles v1 + v2 transparently");
  console.log("[Multi-Version Consumer] Unknown versions → DLT\n");

  await consumer.run({ eachMessage: handleMessage });
}

main().catch(console.error);

process.on("SIGINT", async () => {
  await consumer.disconnect();
  await dlProducer.disconnect();
  process.exit(0);
});
```

---

## Tool 4: `schema-validator.ts` — Standalone Validation

A utility that reads messages from a topic and checks their schema validity without processing them. Useful for auditing.

```typescript
import { Kafka } from "kafkajs";
import { EventEnvelope, validateV1, validateV2 } from "./schemas";

const kafka = new Kafka({ clientId: "schema-validator", brokers: ["localhost:9092"] });
const consumer = kafka.consumer({ groupId: "schema-validator-group" });

interface ValidationResult {
  offset: number;
  partition: number;
  key: string;
  version: number | null;
  valid: boolean;
  errors: string[];
}

const results: ValidationResult[] = [];

async function main(): Promise<void> {
  await consumer.connect();
  await consumer.subscribe({ topic: "orders", fromBeginning: true });

  console.log("[Schema Validator] Auditing all messages on 'orders' topic\n");

  // Run for a few seconds then summarize
  const timeout = setTimeout(async () => {
    await printReport();
    await consumer.disconnect();
    process.exit(0);
  }, 5000);

  await consumer.run({
    eachMessage: async ({ partition, message }) => {
      const key = message.key?.toString() ?? "";
      const raw = message.value?.toString() ?? "";
      const result: ValidationResult = {
        offset: Number(message.offset),
        partition,
        key,
        version: null,
        valid: false,
        errors: [],
      };

      // Parse
      let envelope: EventEnvelope;
      try {
        envelope = JSON.parse(raw) as EventEnvelope;
      } catch {
        result.errors.push("Malformed JSON");
        results.push(result);
        return;
      }

      // Version check
      if (typeof envelope.version !== "number") {
        result.errors.push("Missing version field");
        results.push(result);
        return;
      }

      result.version = envelope.version;

      // Validate by version
      switch (envelope.version) {
        case 1:
          if (validateV1(envelope.payload)) {
            result.valid = true;
          } else {
            result.errors.push("v1 payload validation failed");
          }
          break;
        case 2:
          if (validateV2(envelope.payload)) {
            result.valid = true;
          } else {
            result.errors.push("v2 payload validation failed");
          }
          break;
        default:
          result.errors.push(`Unknown version: ${envelope.version}`);
      }

      results.push(result);
    },
  });
}

async function printReport(): Promise<void> {
  console.log("\n" + "═".repeat(60));
  console.log("SCHEMA VALIDATION REPORT");
  console.log("═".repeat(60));

  const valid = results.filter((r) => r.valid);
  const invalid = results.filter((r) => !r.valid);

  console.log(`Total messages:  ${results.length}`);
  console.log(`Valid:           ${valid.length}`);
  console.log(`Invalid:         ${invalid.length}`);
  console.log();

  // Version distribution
  const versionCounts = new Map<number | null, number>();
  for (const r of results) {
    versionCounts.set(r.version, (versionCounts.get(r.version) ?? 0) + 1);
  }

  console.log("Version Distribution:");
  for (const [ver, count] of versionCounts) {
    console.log(`  v${ver ?? "??"}:  ${count} messages`);
  }

  if (invalid.length > 0) {
    console.log("\nInvalid Messages:");
    for (const r of invalid) {
      console.log(`  P${r.partition}:${r.offset} | ${r.key} | ${r.errors.join(", ")}`);
    }
  }

  console.log("═".repeat(60));
}

main().catch(console.error);
```

---

## Running the Demo

```bash
# Terminal 1: Multi-version consumer
npx ts-node src/multi-version-consumer.ts

# Terminal 2: Send v1 events
npx ts-node src/schema-v1-producer.ts

# Terminal 3: Send v2 events
npx ts-node src/schema-v2-producer.ts

# Terminal 4: Run audit report
npx ts-node src/schema-validator.ts
```

### What to Watch

1. Consumer seamlessly handles both v1 and v2 messages
2. v1 messages get default `currency: "USD"` and empty `items`
3. v2 messages use their actual currency and items
4. The schema validator shows distribution across versions
5. Any unknown versions route to the dead letter topic

---

## Key Code Patterns

```
PATTERN: Multi-Version Consumer

  parse envelope
    → extract version
    → validate against version-specific schema
    → normalize to internal representation
    → process normalized data
    → unknown versions go to DLT

This means you can deploy producer v2 BEFORE updating all consumers.
Consumers on v1 schema won't break — they just apply defaults.
```

→ Next: [Go Implementation](./go-implementation.md)
