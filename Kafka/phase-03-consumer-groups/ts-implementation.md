# Phase 3 — TypeScript Implementation

## Setup

Same project, new consumers. We'll run multiple instances of the same consumer to see consumer groups in action.

### File Structure

```
ts/
├── src/
│   ├── group-consumer.ts       ← Consumer with manual commit
│   ├── auto-commit-consumer.ts ← Consumer with auto commit (to contrast)
│   ├── producer-load.ts        ← Produces a stream of orders for testing
│   └── group-inspector.ts      ← Shows group status via code
├── package.json
└── tsconfig.json
```

---

## `src/group-consumer.ts` — Consumer with Manual Offset Commit

This is the "correct" way to consume in production. You control when offsets are committed.

```typescript
import { Kafka, EachMessagePayload } from "kafkajs";

const consumerId = process.argv[2] || `consumer-${process.pid}`;

const kafka = new Kafka({
  clientId: `payment-service-${consumerId}`,
  brokers: ["localhost:9092"],
});

const consumer = kafka.consumer({
  groupId: "payment-group-v3",
  // Disable auto-commit — we commit manually
  // This is the key change from Phase 1
});

async function processMessage(payload: EachMessagePayload): Promise<void> {
  const { topic, partition, message, heartbeat } = payload;
  const order = JSON.parse(message.value!.toString());

  console.log(
    `[${consumerId}] P${partition}:${message.offset} | ` +
    `${order.eventType} | ${order.orderId} | $${order.amount}`
  );

  // Simulate processing work (200-800ms)
  const processingTime = 200 + Math.random() * 600;
  await new Promise((resolve) => setTimeout(resolve, processingTime));

  // If processing takes a long time, send heartbeat to avoid rebalance
  // Kafka expects a heartbeat every session.timeout.ms (default 30s)
  await heartbeat();

  console.log(
    `[${consumerId}] ✅ Processed ${order.orderId} in ${Math.round(processingTime)}ms`
  );

  // Offset is committed automatically AFTER eachMessage returns successfully
  // If you throw an error here, the offset is NOT committed
  // This gives us at-least-once semantics
}

async function main(): Promise<void> {
  await consumer.connect();

  // Log partition assignment changes (rebalances)
  consumer.on("consumer.group_join", (event) => {
    console.log(`\n[${consumerId}] 🔄 Joined group. Assignment:`,
      JSON.stringify(event.payload.memberAssignment));
  });

  consumer.on("consumer.rebalancing", () => {
    console.log(`\n[${consumerId}] ⚠️ Rebalancing in progress...`);
  });

  await consumer.subscribe({ topic: "orders", fromBeginning: false });

  console.log(`[${consumerId}] Subscribed to 'orders' (group: payment-group-v3)`);
  console.log(`[${consumerId}] Waiting for messages...\n`);

  await consumer.run({
    // autoCommit is true by default in kafkajs, but commits happen
    // AFTER eachMessage returns successfully — so it's effectively
    // "commit after processing" which is what we want
    autoCommitInterval: 5000,
    autoCommitThreshold: 1, // Commit after every message
    eachMessage: processMessage,
  });
}

// Graceful shutdown
const shutdown = async () => {
  console.log(`\n[${consumerId}] Shutting down...`);
  await consumer.disconnect();
  process.exit(0);
};

process.on("SIGINT", shutdown);
process.on("SIGTERM", shutdown);

main().catch(console.error);
```

### How kafkajs Handles Commits

`kafkajs` is slightly different from other clients:
- When `eachMessage` is used, the library automatically commits the offset *after* your handler returns
- If your handler throws, the offset is NOT committed — the message will be redelivered
- `autoCommitThreshold: 1` means it commits after every single message
- This gives you **at-least-once** semantics without writing explicit commit calls

---

## `src/auto-commit-consumer.ts` — The Dangerous Way (For Contrast)

```typescript
import { Kafka } from "kafkajs";

const consumerId = process.argv[2] || `auto-consumer-${process.pid}`;

const kafka = new Kafka({
  clientId: `auto-consumer-${consumerId}`,
  brokers: ["localhost:9092"],
});

const consumer = kafka.consumer({
  groupId: "payment-auto-group",
});

async function main(): Promise<void> {
  await consumer.connect();
  await consumer.subscribe({ topic: "orders", fromBeginning: false });

  console.log(`[${consumerId}] ⚠️ Running with AUTO-COMMIT (dangerous in production)`);
  console.log(`[${consumerId}] This consumer commits offsets on a timer, not after processing\n`);

  await consumer.run({
    autoCommitInterval: 3000, // Commit every 3 seconds regardless of processing state
    eachMessage: async ({ topic, partition, message }) => {
      const order = JSON.parse(message.value!.toString());
      console.log(
        `[${consumerId}] P${partition}:${message.offset} | ${order.orderId}`
      );

      // Simulate SLOW processing
      console.log(`[${consumerId}] Processing... (takes 5 seconds)`);
      await new Promise((resolve) => setTimeout(resolve, 5000));

      console.log(`[${consumerId}] Done processing ${order.orderId}`);

      // Problem: auto-commit may have already committed THIS offset
      // during the 5-second processing window.
      // If we crash now, this message was "committed" but never fully processed.
    },
  });
}

process.on("SIGINT", async () => {
  await consumer.disconnect();
  process.exit(0);
});

main().catch(console.error);
```

### The Problem with Auto-Commit

1. Message arrives at offset 10
2. Processing starts (takes 5 seconds)
3. Auto-commit fires at 3 seconds → commits offset 11
4. Consumer crashes at 4 seconds
5. Restart → consumer resumes at offset 11
6. **Offset 10 was never fully processed but is skipped**

This is why auto-commit is dangerous for any processing that has side effects (payments, inventory, notifications).

---

## `src/producer-load.ts` — Load Producer

Produces a steady stream of orders so you can observe consumer group behavior.

```typescript
import { Kafka } from "kafkajs";
import crypto from "crypto";

const kafka = new Kafka({
  clientId: "load-producer",
  brokers: ["localhost:9092"],
});

const producer = kafka.producer();

const items = ["ITEM-001", "ITEM-002", "ITEM-003", "ITEM-004", "ITEM-005"];
const users = ["user-1", "user-2", "user-3", "user-4", "user-5"];

async function main(): Promise<void> {
  await producer.connect();

  const rate = parseInt(process.argv[2] || "2", 10); // orders per second
  const total = parseInt(process.argv[3] || "50", 10);

  console.log(`[Load] Producing ${total} orders at ${rate}/sec\n`);

  let produced = 0;

  const interval = setInterval(async () => {
    if (produced >= total) {
      clearInterval(interval);
      console.log(`\n[Load] Done. Produced ${produced} orders.`);
      await producer.disconnect();
      return;
    }

    const orderId = `ORD-${crypto.randomUUID().slice(0, 8)}`;
    const order = {
      eventType: "ORDER_CREATED",
      orderId,
      userId: users[Math.floor(Math.random() * users.length)],
      itemId: items[Math.floor(Math.random() * items.length)],
      quantity: 1 + Math.floor(Math.random() * 5),
      amount: parseFloat((Math.random() * 100 + 10).toFixed(2)),
      timestamp: new Date().toISOString(),
    };

    const result = await producer.send({
      topic: "orders",
      messages: [{ key: orderId, value: JSON.stringify(order) }],
    });

    produced++;
    console.log(
      `[Load] #${produced} ${orderId} → P${result[0].partition} (${order.userId}, $${order.amount})`
    );
  }, 1000 / rate);
}

main().catch(console.error);
```

---

## `src/group-inspector.ts` — Inspect Consumer Group State

```typescript
import { Kafka } from "kafkajs";

const kafka = new Kafka({
  clientId: "group-inspector",
  brokers: ["localhost:9092"],
});

const admin = kafka.admin();

async function main(): Promise<void> {
  await admin.connect();

  const groupId = process.argv[2] || "payment-group-v3";

  // Describe the group
  const groups = await admin.describeGroups([groupId]);
  const group = groups.groups[0];

  console.log(`\nConsumer Group: ${groupId}`);
  console.log(`State: ${group.state}`);
  console.log(`Protocol: ${group.protocol}`);
  console.log(`Members: ${group.members.length}`);
  console.log();

  for (const member of group.members) {
    console.log(`  Member: ${member.memberId.slice(0, 40)}...`);
    console.log(`  Client: ${member.clientId}`);

    // Parse member assignment to see partition assignments
    const assignment = member.memberAssignment;
    if (assignment && assignment.length > 0) {
      console.log(`  Assignment: ${assignment.toString("hex").slice(0, 60)}...`);
    }
    console.log();
  }

  // Get offsets
  const offsets = await admin.fetchOffsets({ groupId, topics: ["orders"] });
  console.log("Committed Offsets:");
  for (const topicOffset of offsets) {
    for (const partition of topicOffset.partitions) {
      console.log(`  ${topicOffset.topic} P${partition.partition}: offset ${partition.offset}`);
    }
  }

  // Get topic offsets for lag calculation
  const topicOffsets = await admin.fetchTopicOffsets("orders");
  console.log("\nTopic End Offsets:");
  for (const to of topicOffsets) {
    console.log(`  P${to.partition}: ${to.offset}`);
  }

  console.log("\nConsumer Lag:");
  for (const topicOffset of offsets) {
    for (const partition of topicOffset.partitions) {
      const endOffset = topicOffsets.find(
        (to) => to.partition === partition.partition
      );
      if (endOffset) {
        const lag = parseInt(endOffset.offset) - parseInt(partition.offset);
        const indicator = lag > 0 ? "⚠️" : "✅";
        console.log(
          `  ${indicator} P${partition.partition}: lag = ${lag} messages`
        );
      }
    }
  }

  await admin.disconnect();
}

main().catch(console.error);
```

---

## Running the Demo

### Experiment 1: Watch Rebalancing

```bash
# Terminal 1: Start first consumer
npx ts-node src/group-consumer.ts consumer-A

# Terminal 2: Start producing
npx ts-node src/producer-load.ts 2 100

# Watch consumer-A get all 3 partitions

# Terminal 3: Start second consumer
npx ts-node src/group-consumer.ts consumer-B

# Watch: rebalance happens!
# Partitions are redistributed between consumer-A and consumer-B

# Terminal 4: Start third consumer
npx ts-node src/group-consumer.ts consumer-C

# Each consumer now has exactly 1 partition
```

### Experiment 2: Consumer Crash Recovery

```bash
# With 3 consumers running, kill consumer-B (Ctrl+C)
# Watch: rebalance happens, consumer-A or consumer-C picks up B's partition
# No messages are lost
```

### Experiment 3: Compare Auto-Commit

```bash
# Start the auto-commit consumer
npx ts-node src/auto-commit-consumer.ts

# Produce a message
# While it's "processing" (5 second delay), kill it with Ctrl+C
# Restart it — the message may be skipped!
```

### Experiment 4: Inspect Group State

```bash
# While consumers are running
npx ts-node src/group-inspector.ts payment-group-v3

# Or via CLI
docker exec -it kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --describe --group payment-group-v3
```

---

## What to Observe

1. **Partition assignment is automatic.** You don't decide which consumer gets which partition — Kafka does.
2. **Rebalances cause a pause.** ALL consumers stop briefly when any consumer joins or leaves.
3. **Consumer lag shows real-time health.** If lag grows, your consumers can't keep up.
4. **Manual commit = at-least-once.** You might process a message twice, but you won't miss one.
5. **Auto commit = at-most-once (with slow processing).** You might miss messages if you crash mid-processing.

→ Next: [Phase 3 — Go Implementation](go-implementation.md)
