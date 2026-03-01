# Phase 7 — TypeScript Implementation

## Setup

```bash
cd ts/
# kafkajs already installed
```

### File Structure

```
ts/
├── src/
│   ├── retention-demo.ts
│   ├── compaction-demo.ts
│   ├── replay-consumer.ts
│   ├── offset-reset.ts
│   └── time-travel-consumer.ts
├── package.json
└── tsconfig.json
```

---

## Tool 1: `retention-demo.ts` — Observe Message Expiry

Creates a topic with 60-second retention, produces messages, and shows them disappearing.

```typescript
import { Kafka, Admin } from "kafkajs";

const kafka = new Kafka({ clientId: "retention-demo", brokers: ["localhost:9092"] });
const admin = kafka.admin();
const producer = kafka.producer();

const TOPIC = "orders-retention-demo";
const RETENTION_MS = "60000"; // 60 seconds

async function createTopicWithRetention(): Promise<void> {
  await admin.connect();

  // Delete if exists
  const topics = await admin.listTopics();
  if (topics.includes(TOPIC)) {
    await admin.deleteTopics({ topics: [TOPIC] });
    console.log(`Deleted existing topic: ${TOPIC}`);
    // Wait for deletion to propagate
    await new Promise((r) => setTimeout(r, 2000));
  }

  await admin.createTopics({
    topics: [
      {
        topic: TOPIC,
        numPartitions: 1,
        replicationFactor: 1,
        configEntries: [
          { name: "retention.ms", value: RETENTION_MS },
          { name: "segment.ms", value: "10000" },     // Roll segments every 10s
          { name: "segment.bytes", value: "1048576" }, // 1MB segments
        ],
      },
    ],
  });

  console.log(`Created topic: ${TOPIC}`);
  console.log(`Retention: ${RETENTION_MS}ms (${parseInt(RETENTION_MS) / 1000}s)`);
}

async function produceMessages(): Promise<void> {
  await producer.connect();

  for (let i = 1; i <= 10; i++) {
    await producer.send({
      topic: TOPIC,
      messages: [
        {
          key: `ORD-ret-${i}`,
          value: JSON.stringify({
            orderId: `ORD-ret-${i}`,
            amount: i * 10,
            timestamp: new Date().toISOString(),
          }),
        },
      ],
    });
    console.log(`Produced: ORD-ret-${i}`);
  }
}

async function consumeAndCount(label: string): Promise<number> {
  const consumer = kafka.consumer({ groupId: `retention-check-${Date.now()}` });
  await consumer.connect();
  await consumer.subscribe({ topic: TOPIC, fromBeginning: true });

  let count = 0;
  const messages: string[] = [];

  await new Promise<void>((resolve) => {
    const timeout = setTimeout(() => resolve(), 3000);

    consumer.run({
      eachMessage: async ({ message }) => {
        count++;
        messages.push(message.key?.toString() ?? "?");
      },
    });
  });

  console.log(`[${label}] Found ${count} messages: [${messages.join(", ")}]`);
  await consumer.disconnect();
  return count;
}

async function main(): Promise<void> {
  await createTopicWithRetention();
  await produceMessages();

  console.log("\n--- Consuming immediately ---");
  const before = await consumeAndCount("Before expiry");

  console.log(`\n--- Waiting ${parseInt(RETENTION_MS) / 1000 + 30}s for retention + cleanup ---`);
  console.log("(Kafka checks retention periodically, not instantly)");

  await new Promise((r) => setTimeout(r, parseInt(RETENTION_MS) + 30000));

  console.log("\n--- Consuming after retention period ---");
  const after = await consumeAndCount("After expiry");

  console.log("\n═══════════════════════════════════════");
  console.log(`Before: ${before} messages`);
  console.log(`After:  ${after} messages`);
  console.log(`Expired: ${before - after} messages`);
  console.log("═══════════════════════════════════════");

  await admin.disconnect();
  await producer.disconnect();
}

main().catch(console.error);
```

---

## Tool 2: `compaction-demo.ts` — Log Compaction in Action

Writes multiple values for the same keys, then reads back to observe compaction.

```typescript
import { Kafka } from "kafkajs";

const kafka = new Kafka({ clientId: "compaction-demo", brokers: ["localhost:9092"] });
const admin = kafka.admin();
const producer = kafka.producer();

const TOPIC = "order-status-compacted";

async function setup(): Promise<void> {
  await admin.connect();

  const topics = await admin.listTopics();
  if (topics.includes(TOPIC)) {
    await admin.deleteTopics({ topics: [TOPIC] });
    await new Promise((r) => setTimeout(r, 2000));
  }

  await admin.createTopics({
    topics: [
      {
        topic: TOPIC,
        numPartitions: 1,
        replicationFactor: 1,
        configEntries: [
          { name: "cleanup.policy", value: "compact" },
          { name: "min.cleanable.dirty.ratio", value: "0.01" },
          { name: "segment.ms", value: "1000" },
          { name: "delete.retention.ms", value: "1000" },
        ],
      },
    ],
  });

  console.log(`Created compacted topic: ${TOPIC}`);
}

async function produceOrderUpdates(): Promise<void> {
  await producer.connect();

  // Simulate order lifecycle: multiple status updates per order
  const updates = [
    { key: "ORD-001", status: "CREATED", step: 1 },
    { key: "ORD-002", status: "CREATED", step: 1 },
    { key: "ORD-001", status: "PAYMENT_PENDING", step: 2 },
    { key: "ORD-003", status: "CREATED", step: 1 },
    { key: "ORD-001", status: "PAYMENT_CONFIRMED", step: 3 },
    { key: "ORD-002", status: "PAYMENT_PENDING", step: 2 },
    { key: "ORD-001", status: "SHIPPED", step: 4 },
    { key: "ORD-002", status: "PAYMENT_CONFIRMED", step: 3 },
    { key: "ORD-003", status: "CANCELLED", step: 2 },
    { key: "ORD-001", status: "DELIVERED", step: 5 },
  ];

  console.log("\nProducing order status updates:");
  for (const update of updates) {
    await producer.send({
      topic: TOPIC,
      messages: [
        {
          key: update.key,
          value: JSON.stringify({
            orderId: update.key,
            status: update.status,
            updatedAt: new Date().toISOString(),
          }),
        },
      ],
    });
    console.log(`  ${update.key} → ${update.status}`);
  }

  console.log(`\nTotal messages produced: ${updates.length}`);
  console.log("ORD-001 has 5 updates (only latest should survive compaction)");
  console.log("ORD-002 has 3 updates");
  console.log("ORD-003 has 2 updates");
}

async function readAll(label: string): Promise<void> {
  const consumer = kafka.consumer({ groupId: `compact-read-${Date.now()}` });
  await consumer.connect();
  await consumer.subscribe({ topic: TOPIC, fromBeginning: true });

  const messages: { key: string; status: string; offset: string }[] = [];

  await new Promise<void>((resolve) => {
    setTimeout(resolve, 3000);
    consumer.run({
      eachMessage: async ({ message }) => {
        const val = JSON.parse(message.value?.toString() ?? "{}");
        messages.push({
          key: message.key?.toString() ?? "",
          status: val.status,
          offset: message.offset,
        });
      },
    });
  });

  console.log(`\n[${label}] ${messages.length} messages read:`);
  for (const m of messages) {
    console.log(`  offset ${m.offset}: ${m.key} → ${m.status}`);
  }

  await consumer.disconnect();
}

async function main(): Promise<void> {
  await setup();
  await produceOrderUpdates();

  console.log("\n--- Reading immediately (before compaction) ---");
  await readAll("Before compaction");

  console.log("\n--- Waiting 30s for log compaction to run ---");
  console.log("(Compaction runs in the background, timing varies)");
  await new Promise((r) => setTimeout(r, 30000));

  console.log("\n--- Reading after compaction ---");
  await readAll("After compaction");

  console.log("\nExpected after compaction:");
  console.log("  ORD-001 → DELIVERED (latest of 5 updates)");
  console.log("  ORD-002 → PAYMENT_CONFIRMED (latest of 3 updates)");
  console.log("  ORD-003 → CANCELLED (latest of 2 updates)");
  console.log("  Total: 3 messages (down from 10)");

  await admin.disconnect();
  await producer.disconnect();
}

main().catch(console.error);
```

---

## Tool 3: `replay-consumer.ts` — Fresh Replay from Beginning

```typescript
import { Kafka, EachMessagePayload } from "kafkajs";

const kafka = new Kafka({ clientId: "replay-consumer", brokers: ["localhost:9092"] });

// Use a unique group ID to replay from scratch
const GROUP_ID = `replay-group-${Date.now()}`;

async function main(): Promise<void> {
  const consumer = kafka.consumer({ groupId: GROUP_ID });
  await consumer.connect();
  await consumer.subscribe({ topic: "orders", fromBeginning: true });

  let count = 0;
  const startTime = Date.now();

  console.log(`[Replay] Starting from offset 0 with group: ${GROUP_ID}`);
  console.log("[Replay] This re-reads the ENTIRE topic history\n");

  await consumer.run({
    eachMessage: async ({ partition, message }: EachMessagePayload) => {
      count++;
      const key = message.key?.toString() ?? "";
      const ts = message.timestamp;
      const age = Date.now() - parseInt(ts);

      console.log(
        `  #${count} | P${partition}:${message.offset} | ${key} | age: ${Math.round(age / 1000)}s`
      );
    },
  });

  // Print summary after 10 seconds
  setTimeout(async () => {
    const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
    console.log(`\n[Replay] Replayed ${count} messages in ${elapsed}s`);
    console.log("[Replay] This is a full re-read — no messages were skipped");
    await consumer.disconnect();
    process.exit(0);
  }, 10000);
}

main().catch(console.error);
```

---

## Tool 4: `offset-reset.ts` — Programmatic Offset Reset

```typescript
import { Kafka } from "kafkajs";

const kafka = new Kafka({ clientId: "offset-reset", brokers: ["localhost:9092"] });
const admin = kafka.admin();

type ResetMode = "earliest" | "latest" | "timestamp" | "offset";

async function resetOffsets(
  groupId: string,
  topic: string,
  mode: ResetMode,
  value?: number
): Promise<void> {
  await admin.connect();

  // Fetch current offsets
  const currentOffsets = await admin.fetchOffsets({ groupId, topics: [topic] });
  console.log("\nCurrent offsets:");
  for (const t of currentOffsets) {
    for (const p of t.partitions) {
      console.log(`  ${t.topic} P${p.partition}: offset ${p.offset}`);
    }
  }

  // Get topic metadata for partition info
  const topicOffsets = await admin.fetchTopicOffsets(topic);

  switch (mode) {
    case "earliest": {
      const earliest = await admin.fetchTopicOffsetsByTimestamp(topic, -2);
      await admin.setOffsets({
        groupId,
        topic,
        partitions: earliest.map((p) => ({
          partition: p.partition,
          offset: p.offset,
        })),
      });
      console.log("\n✅ Reset to EARLIEST");
      break;
    }

    case "latest": {
      await admin.setOffsets({
        groupId,
        topic,
        partitions: topicOffsets.map((p) => ({
          partition: p.partition,
          offset: p.offset, // high watermark
        })),
      });
      console.log("\n✅ Reset to LATEST");
      break;
    }

    case "timestamp": {
      if (!value) throw new Error("Timestamp required");
      const byTime = await admin.fetchTopicOffsetsByTimestamp(topic, value);
      await admin.setOffsets({
        groupId,
        topic,
        partitions: byTime.map((p) => ({
          partition: p.partition,
          offset: p.offset,
        })),
      });
      console.log(`\n✅ Reset to timestamp ${new Date(value).toISOString()}`);
      break;
    }

    case "offset": {
      if (value === undefined) throw new Error("Offset required");
      await admin.setOffsets({
        groupId,
        topic,
        partitions: topicOffsets.map((p) => ({
          partition: p.partition,
          offset: String(value),
        })),
      });
      console.log(`\n✅ Reset all partitions to offset ${value}`);
      break;
    }
  }

  // Verify
  const newOffsets = await admin.fetchOffsets({ groupId, topics: [topic] });
  console.log("\nNew offsets:");
  for (const t of newOffsets) {
    for (const p of t.partitions) {
      console.log(`  ${t.topic} P${p.partition}: offset ${p.offset}`);
    }
  }

  await admin.disconnect();
}

// ─── CLI ───
const args = process.argv.slice(2);
const groupId = args[0] ?? "payment-group";
const topic = args[1] ?? "orders";
const mode = (args[2] ?? "earliest") as ResetMode;
const value = args[3] ? parseInt(args[3]) : undefined;

console.log(`[Offset Reset] group=${groupId} topic=${topic} mode=${mode}`);
console.log("⚠️  Make sure consumers in this group are STOPPED first!\n");

resetOffsets(groupId, topic, mode, value).catch(console.error);
```

**Usage:**

```bash
# Reset to beginning
npx ts-node src/offset-reset.ts payment-group orders earliest

# Reset to latest
npx ts-node src/offset-reset.ts payment-group orders latest

# Reset to timestamp (epoch ms)
npx ts-node src/offset-reset.ts payment-group orders timestamp 1705312800000

# Reset to specific offset
npx ts-node src/offset-reset.ts payment-group orders offset 50
```

---

## Tool 5: `time-travel-consumer.ts` — Start from a Specific Time

```typescript
import { Kafka, EachMessagePayload } from "kafkajs";

const kafka = new Kafka({ clientId: "time-travel", brokers: ["localhost:9092"] });
const admin = kafka.admin();

async function main(): Promise<void> {
  const topic = "orders";
  const minutesAgo = parseInt(process.argv[2] ?? "5");
  const targetTime = Date.now() - minutesAgo * 60 * 1000;
  const groupId = `time-travel-${Date.now()}`;

  console.log(`[Time Travel] Consuming from ${minutesAgo} minutes ago`);
  console.log(`[Time Travel] Target: ${new Date(targetTime).toISOString()}`);
  console.log(`[Time Travel] Group: ${groupId}\n`);

  // Find offsets for the target timestamp
  await admin.connect();
  const offsetsByTime = await admin.fetchTopicOffsetsByTimestamp(topic, targetTime);

  console.log("Resolved offsets:");
  for (const p of offsetsByTime) {
    console.log(`  P${p.partition}: offset ${p.offset}`);
  }
  console.log();

  // Create consumer and seek to those offsets
  const consumer = kafka.consumer({ groupId });
  await consumer.connect();
  await consumer.subscribe({ topic, fromBeginning: false });

  // Set the starting offsets
  await admin.setOffsets({
    groupId,
    topic,
    partitions: offsetsByTime.map((p) => ({
      partition: p.partition,
      offset: p.offset,
    })),
  });

  let count = 0;

  await consumer.run({
    eachMessage: async ({ partition, message }: EachMessagePayload) => {
      count++;
      const key = message.key?.toString() ?? "";
      const msgTime = new Date(parseInt(message.timestamp));
      console.log(
        `  #${count} | P${partition}:${message.offset} | ${key} | ${msgTime.toISOString()}`
      );
    },
  });

  setTimeout(async () => {
    console.log(`\n[Time Travel] Read ${count} messages from the last ${minutesAgo} minutes`);
    await consumer.disconnect();
    await admin.disconnect();
    process.exit(0);
  }, 10000);
}

main().catch(console.error);
```

**Usage:**

```bash
# Read all messages from the last 5 minutes
npx ts-node src/time-travel-consumer.ts 5

# Read from the last 60 minutes
npx ts-node src/time-travel-consumer.ts 60
```

---

## Running the Full Demo

```bash
# 1: Retention demo (takes ~90 seconds)
npx ts-node src/retention-demo.ts

# 2: Compaction demo (takes ~30 seconds)
npx ts-node src/compaction-demo.ts

# 3: Replay everything
npx ts-node src/replay-consumer.ts

# 4: Reset a group's offsets
npx ts-node src/offset-reset.ts payment-group orders earliest

# 5: Time travel
npx ts-node src/time-travel-consumer.ts 10
```

→ Next: [Go Implementation](./go-implementation.md)
