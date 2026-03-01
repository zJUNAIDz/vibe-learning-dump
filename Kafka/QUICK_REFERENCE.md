# Quick Reference — Kafka CLI Cheat Sheet

All commands assume `--bootstrap-server localhost:9092`. Adjust if your broker is elsewhere.

## Topics

```bash
# List all topics
kafka-topics --bootstrap-server localhost:9092 --list

# Create a topic
kafka-topics --bootstrap-server localhost:9092 \
  --create --topic orders \
  --partitions 3 --replication-factor 1

# Describe a topic
kafka-topics --bootstrap-server localhost:9092 \
  --describe --topic orders

# Delete a topic
kafka-topics --bootstrap-server localhost:9092 \
  --delete --topic orders

# Alter partition count (can only increase)
kafka-topics --bootstrap-server localhost:9092 \
  --alter --topic orders --partitions 6
```

## Producing

```bash
# Interactive producer
kafka-console-producer --bootstrap-server localhost:9092 \
  --topic orders

# Producer with keys (key:value)
kafka-console-producer --bootstrap-server localhost:9092 \
  --topic orders \
  --property parse.key=true \
  --property key.separator=:
```

## Consuming

```bash
# Consume from beginning
kafka-console-consumer --bootstrap-server localhost:9092 \
  --topic orders --from-beginning

# Consume with keys, timestamps, partitions
kafka-console-consumer --bootstrap-server localhost:9092 \
  --topic orders --from-beginning \
  --property print.key=true \
  --property print.timestamp=true \
  --property print.partition=true

# Consume from specific partition
kafka-console-consumer --bootstrap-server localhost:9092 \
  --topic orders --partition 0 --offset earliest

# Consume with a consumer group
kafka-console-consumer --bootstrap-server localhost:9092 \
  --topic orders --group my-group
```

## Consumer Groups

```bash
# List consumer groups
kafka-consumer-groups --bootstrap-server localhost:9092 --list

# Describe a group (shows lag, offsets, assignments)
kafka-consumer-groups --bootstrap-server localhost:9092 \
  --describe --group my-group

# Reset offsets to earliest (group must be stopped)
kafka-consumer-groups --bootstrap-server localhost:9092 \
  --group my-group --topic orders \
  --reset-offsets --to-earliest --execute

# Reset offsets to a specific timestamp
kafka-consumer-groups --bootstrap-server localhost:9092 \
  --group my-group --topic orders \
  --reset-offsets --to-datetime 2025-01-01T00:00:00.000 --execute

# Reset to specific offset
kafka-consumer-groups --bootstrap-server localhost:9092 \
  --group my-group --topic orders \
  --reset-offsets --to-offset 42 --execute
```

## Metadata & Debugging

```bash
# Broker metadata
kafka-metadata --bootstrap-server localhost:9092 --describe

# Get topic config
kafka-configs --bootstrap-server localhost:9092 \
  --entity-type topics --entity-name orders --describe

# Set retention to 1 hour
kafka-configs --bootstrap-server localhost:9092 \
  --entity-type topics --entity-name orders \
  --alter --add-config retention.ms=3600000

# Get earliest/latest offsets
kafka-get-offsets --bootstrap-server localhost:9092 \
  --topic orders
```

## Docker Compose Shortcuts

```bash
# Start cluster
docker compose up -d

# Stop cluster
docker compose down

# Stop and delete all data
docker compose down -v

# View Kafka logs
docker compose logs -f kafka

# Exec into Kafka container
docker exec -it kafka bash
```
