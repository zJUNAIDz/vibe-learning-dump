# Getting Started

## Environment Setup

### 1. Docker & Docker Compose

You need Docker to run Kafka locally. Install Docker Desktop or Docker Engine.

```bash
docker --version
docker compose version
```

### 2. Node.js & TypeScript

```bash
node --version   # 18+
npm install -g typescript ts-node
```

### 3. Go

```bash
go version   # 1.21+
```

### 4. Kafka CLI Tools

Kafka ships with CLI tools inside its Docker image. We'll alias them:

```bash
# Add to your .bashrc / .zshrc
alias kafka-topics='docker exec -it kafka kafka-topics'
alias kafka-console-producer='docker exec -it kafka kafka-console-producer'
alias kafka-console-consumer='docker exec -it kafka kafka-console-consumer'
alias kafka-consumer-groups='docker exec -it kafka kafka-consumer-groups'
```

These aliases assume a container named `kafka` — which our Docker Compose file creates.

### 5. Verify Everything

After Phase 1 sets up Docker Compose:

```bash
# Start Kafka
docker compose up -d

# Create a test topic
kafka-topics --bootstrap-server localhost:9092 --create --topic test --partitions 1

# Produce
echo "hello" | kafka-console-producer --bootstrap-server localhost:9092 --topic test

# Consume
kafka-console-consumer --bootstrap-server localhost:9092 --topic test --from-beginning
```

If you see `hello`, you're ready.

## Project Structure

```
Kafka/
├── README.md                    ← You are here (go read it first)
├── GETTING_STARTED.md           ← This file
├── QUICK_REFERENCE.md           ← Kafka CLI cheat sheet
├── START_HERE.md                ← Narrative entry point
├── docker-compose.yml           ← Shared Kafka cluster config
├── phase-00-pre-kafka/          ← Start here
├── phase-01-log-basics/
├── phase-02-partitions/
├── phase-03-consumer-groups/
├── phase-04-failure-retries/
├── phase-05-dead-letter/
├── phase-06-schemas/
├── phase-07-replay/
└── phase-08-ops/
```

## Language Libraries

### TypeScript: `kafkajs`

```bash
npm init -y
npm install kafkajs
npm install -D typescript @types/node ts-node
```

### Go: `segmentio/kafka-go`

```bash
go mod init order-pipeline
go get github.com/segmentio/kafka-go
```

Both are lightweight, well-maintained, and have no hidden dependencies.
