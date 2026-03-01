# Phase 8 — TypeScript Implementation

## Setup

```bash
cd ts/
# kafkajs already installed
```

### File Structure

```
ts/
├── src/
│   ├── lag-monitor.ts
│   ├── throughput-meter.ts
│   ├── health-check.ts
│   └── load-generator.ts
├── package.json
└── tsconfig.json
```

---

## Tool 1: `lag-monitor.ts` — Consumer Lag Alerting

Continuously polls consumer group lag and alerts when it exceeds a threshold.

```typescript
import { Kafka, Admin } from "kafkajs";

const kafka = new Kafka({ clientId: "lag-monitor", brokers: ["localhost:9092"] });
const admin = kafka.admin();

interface PartitionLag {
  topic: string;
  partition: number;
  currentOffset: number;
  logEndOffset: number;
  lag: number;
}

interface GroupLag {
  groupId: string;
  state: string;
  totalLag: number;
  partitions: PartitionLag[];
}

async function getGroupLag(groupId: string): Promise<GroupLag> {
  const description = await admin.describeGroups([groupId]);
  const group = description.groups[0];

  const offsets = await admin.fetchOffsets({ groupId, topics: ["orders"] });
  const topicOffsets = await admin.fetchTopicOffsets("orders");

  const partitions: PartitionLag[] = [];

  for (const topicData of offsets) {
    for (const partData of topicData.partitions) {
      const logEnd = topicOffsets.find((t) => t.partition === partData.partition);
      const currentOffset = parseInt(partData.offset) || 0;
      const logEndOffset = parseInt(logEnd?.offset ?? "0");
      const lag = Math.max(0, logEndOffset - currentOffset);

      partitions.push({
        topic: topicData.topic,
        partition: partData.partition,
        currentOffset,
        logEndOffset,
        lag,
      });
    }
  }

  return {
    groupId,
    state: group.state,
    totalLag: partitions.reduce((sum, p) => sum + p.lag, 0),
    partitions,
  };
}

async function main(): Promise<void> {
  const groupId = process.argv[2] ?? "payment-group";
  const threshold = parseInt(process.argv[3] ?? "100");
  const intervalMs = parseInt(process.argv[4] ?? "5000");

  await admin.connect();

  console.log(`[Lag Monitor] Watching group: ${groupId}`);
  console.log(`[Lag Monitor] Alert threshold: ${threshold}`);
  console.log(`[Lag Monitor] Poll interval: ${intervalMs}ms`);
  console.log();

  let alertActive = false;
  let consecutiveAlerts = 0;

  const poll = async (): Promise<void> => {
    try {
      const lag = await getGroupLag(groupId);
      const now = new Date().toISOString().substring(11, 19);

      // Header line
      let status = "✅";
      if (lag.totalLag > threshold) {
        status = "🚨";
        consecutiveAlerts++;
      } else {
        consecutiveAlerts = 0;
      }

      console.log(
        `[${now}] ${status} ${lag.groupId} | state=${lag.state} | total_lag=${lag.totalLag}`
      );

      // Per-partition detail
      for (const p of lag.partitions) {
        const bar = "█".repeat(Math.min(50, Math.round(p.lag / 10)));
        const lagStr = p.lag > threshold ? `⚠️  ${p.lag}` : `${p.lag}`;
        console.log(
          `  P${p.partition}: ${p.currentOffset}/${p.logEndOffset} lag=${lagStr} ${bar}`
        );
      }

      // Alert logic
      if (lag.totalLag > threshold && !alertActive) {
        alertActive = true;
        console.log(
          `\n  🚨 ALERT: Lag ${lag.totalLag} exceeds threshold ${threshold}!`
        );
        console.log(
          `  🚨 Action: Check consumer health, consider scaling out\n`
        );
      } else if (lag.totalLag <= threshold && alertActive) {
        alertActive = false;
        console.log(`\n  ✅ RECOVERED: Lag back to normal (${lag.totalLag})\n`);
      }

      if (consecutiveAlerts >= 3) {
        console.log(
          `  🔥 CRITICAL: Lag above threshold for ${consecutiveAlerts} consecutive checks`
        );
      }

      console.log();
    } catch (err) {
      console.log(`[${new Date().toISOString().substring(11, 19)}] ❌ Error: ${err}`);
    }
  };

  // Initial poll
  await poll();

  // Continuous polling
  setInterval(poll, intervalMs);
}

main().catch(console.error);
```

**Usage:**

```bash
# Monitor payment-group, alert at lag > 100, poll every 5s
npx ts-node src/lag-monitor.ts payment-group 100 5000
```

---

## Tool 2: `throughput-meter.ts` — Real-time Rates

```typescript
import { Kafka, EachMessagePayload } from "kafkajs";

const kafka = new Kafka({ clientId: "throughput-meter", brokers: ["localhost:9092"] });

interface WindowStats {
  count: number;
  bytes: number;
  startTime: number;
}

class ThroughputMeter {
  private windows: WindowStats[] = [];
  private currentWindow: WindowStats;
  private readonly windowMs: number;

  constructor(windowMs: number = 1000) {
    this.windowMs = windowMs;
    this.currentWindow = { count: 0, bytes: 0, startTime: Date.now() };
  }

  record(bytes: number): void {
    const now = Date.now();

    if (now - this.currentWindow.startTime >= this.windowMs) {
      this.windows.push(this.currentWindow);
      // Keep last 60 windows (1 minute of history)
      if (this.windows.length > 60) {
        this.windows.shift();
      }
      this.currentWindow = { count: 0, bytes: 0, startTime: now };
    }

    this.currentWindow.count++;
    this.currentWindow.bytes += bytes;
  }

  getStats(): { msgPerSec: number; bytesPerSec: number; totalMessages: number } {
    const recentWindows = this.windows.slice(-5);
    if (recentWindows.length === 0) {
      return { msgPerSec: 0, bytesPerSec: 0, totalMessages: 0 };
    }

    const totalCount = recentWindows.reduce((s, w) => s + w.count, 0);
    const totalBytes = recentWindows.reduce((s, w) => s + w.bytes, 0);
    const durationSec = (recentWindows.length * this.windowMs) / 1000;

    const allTime = this.windows.reduce((s, w) => s + w.count, 0) + this.currentWindow.count;

    return {
      msgPerSec: Math.round(totalCount / durationSec),
      bytesPerSec: Math.round(totalBytes / durationSec),
      totalMessages: allTime,
    };
  }
}

async function main(): Promise<void> {
  const topic = process.argv[2] ?? "orders";
  const consumer = kafka.consumer({ groupId: `throughput-meter-${Date.now()}` });
  await consumer.connect();
  await consumer.subscribe({ topic, fromBeginning: false });

  const meter = new ThroughputMeter(1000);

  console.log(`[Throughput Meter] Measuring consume rate on '${topic}'`);
  console.log("[Throughput Meter] Start producing messages to see throughput\n");

  await consumer.run({
    eachMessage: async ({ message }: EachMessagePayload) => {
      const size = (message.value?.length ?? 0) + (message.key?.length ?? 0);
      meter.record(size);
    },
  });

  // Print stats every 2 seconds
  setInterval(() => {
    const stats = meter.getStats();
    const now = new Date().toISOString().substring(11, 19);

    const kbPerSec = (stats.bytesPerSec / 1024).toFixed(1);

    console.log(
      `[${now}] ${stats.msgPerSec} msg/s | ${kbPerSec} KB/s | total: ${stats.totalMessages}`
    );
  }, 2000);
}

main().catch(console.error);
```

---

## Tool 3: `health-check.ts` — Cluster Health Report

```typescript
import { Kafka } from "kafkajs";

const kafka = new Kafka({ clientId: "health-check", brokers: ["localhost:9092"] });
const admin = kafka.admin();

async function main(): Promise<void> {
  await admin.connect();

  console.log("╔══════════════════════════════════════════════════════");
  console.log("║ KAFKA CLUSTER HEALTH CHECK");
  console.log("╠══════════════════════════════════════════════════════");

  // ─── Cluster info ───
  const cluster = await admin.describeCluster();
  console.log("║");
  console.log(`║ Controller: Broker ${cluster.controller}`);
  console.log(`║ Brokers: ${cluster.brokers.length}`);
  for (const broker of cluster.brokers) {
    console.log(`║   ID=${broker.nodeId} ${broker.host}:${broker.port}`);
  }

  // ─── Topics ───
  const topics = await admin.listTopics();
  const topicMetadata = await admin.fetchTopicMetadata({ topics });

  console.log("║");
  console.log(`║ Topics: ${topics.length}`);

  let totalPartitions = 0;
  let underReplicated = 0;
  const issues: string[] = [];

  for (const topic of topicMetadata.topics) {
    totalPartitions += topic.partitions.length;

    for (const partition of topic.partitions) {
      if (partition.isr.length < partition.replicas.length) {
        underReplicated++;
        issues.push(
          `${topic.name} P${partition.partitionId}: ISR=${partition.isr.length}/${partition.replicas.length}`
        );
      }
    }

    console.log(
      `║   ${topic.name}: ${topic.partitions.length} partitions`
    );
  }

  console.log("║");
  console.log(`║ Total partitions: ${totalPartitions}`);
  console.log(`║ Under-replicated: ${underReplicated}`);

  // ─── Consumer groups ───
  const groups = await admin.listGroups();
  console.log("║");
  console.log(`║ Consumer Groups: ${groups.groups.length}`);

  for (const group of groups.groups) {
    const desc = await admin.describeGroups([group.groupId]);
    const g = desc.groups[0];
    const memberCount = g.members.length;

    let stateIcon = "✅";
    if (g.state === "Empty") stateIcon = "⚪";
    if (g.state === "Rebalancing") stateIcon = "⚠️";
    if (g.state === "Dead") stateIcon = "❌";

    console.log(
      `║   ${stateIcon} ${group.groupId}: ${g.state} (${memberCount} members)`
    );

    if (g.state === "Rebalancing") {
      issues.push(`Group ${group.groupId} is rebalancing`);
    }
  }

  // ─── Overall status ───
  console.log("║");
  console.log("╠══════════════════════════════════════════════════════");

  if (issues.length === 0) {
    console.log("║ ✅ HEALTHY — No issues detected");
  } else {
    console.log(`║ ⚠️  ${issues.length} ISSUE(S) FOUND:`);
    for (const issue of issues) {
      console.log(`║   - ${issue}`);
    }
  }

  console.log("╚══════════════════════════════════════════════════════");

  await admin.disconnect();
}

main().catch(console.error);
```

---

## Tool 4: `load-generator.ts` — Sustained Load Producer

```typescript
import { Kafka, CompressionTypes } from "kafkajs";

const kafka = new Kafka({ clientId: "load-generator", brokers: ["localhost:9092"] });
const producer = kafka.producer();

async function main(): Promise<void> {
  const targetRate = parseInt(process.argv[2] ?? "10"); // messages per second
  const durationSec = parseInt(process.argv[3] ?? "60"); // how long to run
  const topic = process.argv[4] ?? "orders";

  await producer.connect();

  console.log(`[Load Generator] Target: ${targetRate} msg/s for ${durationSec}s`);
  console.log(`[Load Generator] Topic: ${topic}`);
  console.log(`[Load Generator] Total expected: ${targetRate * durationSec} messages\n`);

  const intervalMs = 1000 / targetRate;
  let sent = 0;
  let errors = 0;
  const startTime = Date.now();

  const interval = setInterval(async () => {
    const elapsed = (Date.now() - startTime) / 1000;

    if (elapsed >= durationSec) {
      clearInterval(interval);
      const actualRate = (sent / elapsed).toFixed(1);
      console.log(`\n[Load Generator] Done.`);
      console.log(`  Sent: ${sent} messages`);
      console.log(`  Errors: ${errors}`);
      console.log(`  Duration: ${elapsed.toFixed(1)}s`);
      console.log(`  Actual rate: ${actualRate} msg/s`);
      await producer.disconnect();
      process.exit(0);
    }

    const orderId = `ORD-load-${sent + 1}`;

    try {
      await producer.send({
        topic,
        messages: [
          {
            key: orderId,
            value: JSON.stringify({
              version: 2,
              eventType: "ORDER_CREATED",
              timestamp: new Date().toISOString(),
              source: "load-generator",
              payload: {
                orderId,
                userId: `user-${(sent % 100) + 1}`,
                amount: Math.round(Math.random() * 500 * 100) / 100,
                currency: "USD",
              },
            }),
          },
        ],
      });
      sent++;

      // Progress every 100 messages
      if (sent % 100 === 0) {
        const currentRate = (sent / elapsed).toFixed(1);
        console.log(`  [${elapsed.toFixed(0)}s] Sent: ${sent} | Rate: ${currentRate} msg/s`);
      }
    } catch (err) {
      errors++;
    }
  }, intervalMs);
}

main().catch(console.error);
```

**Usage:**

```bash
# 10 messages/second for 60 seconds
npx ts-node src/load-generator.ts 10 60

# 100 messages/second for 30 seconds
npx ts-node src/load-generator.ts 100 30

# Stress test: 500 msg/s for 10 seconds
npx ts-node src/load-generator.ts 500 10
```

---

## Running the Full Demo

### Scenario: Watch Lag Grow and Recover

```bash
# Terminal 1: Start lag monitor
npx ts-node src/lag-monitor.ts payment-group 50 3000

# Terminal 2: Start throughput meter
npx ts-node src/throughput-meter.ts orders

# Terminal 3: Start ONE slow consumer (simulates bottleneck)
# (Use the consumer from Phase 4 with added delay)
npx ts-node src/consumer-with-dlt.ts slow-consumer

# Terminal 4: Start load generator at 50 msg/s
npx ts-node src/load-generator.ts 50 120

# Watch: Lag monitor shows lag growing because ONE consumer can't keep up
# Action: Start a SECOND consumer in same group (Terminal 5)
# Watch: Lag stabilizes and starts recovering
```

### Scenario: Health Check

```bash
# Single comprehensive check
npx ts-node src/health-check.ts
```

---

## Operational Playbook

```
WHEN lag is growing:
  1. Check consumer health (health-check.ts)
  2. Check consumer logs for errors
  3. Scale out consumers (up to partition count)
  4. If at partition limit, consider repartitioning

WHEN rebalancing won't stop:
  1. Check for consumers dying/restarting
  2. Increase session.timeout.ms
  3. Check max.poll.interval.ms vs actual processing time
  4. Look for long GC pauses or network issues

WHEN throughput is lower than expected:
  1. Check broker disk I/O
  2. Check network between producer/consumer and broker
  3. Try batch producing (linger.ms > 0)
  4. Consider compression (lz4 or zstd)
  5. Verify partition count matches parallelism needs
```

→ Next: [Go Implementation](./go-implementation.md)
