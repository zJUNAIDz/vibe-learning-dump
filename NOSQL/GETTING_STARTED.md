# Getting Started — Environment Setup

---

## What You'll Need

This curriculum uses **MongoDB** and **Cassandra** as primary teaching databases. You'll also briefly encounter Redis and conceptual examples for other databases.

### Required

| Tool | Purpose | Install |
|------|---------|---------|
| **Docker** | Run databases locally | [docs.docker.com](https://docs.docker.com/get-docker/) |
| **Docker Compose** | Multi-container setups | Included with Docker Desktop |
| **mongosh** | MongoDB shell | `brew install mongosh` or [mongosh docs](https://www.mongodb.com/docs/mongodb-shell/) |
| **Node.js 20+** | TypeScript examples | [nodejs.org](https://nodejs.org/) |
| **Go 1.21+** | Go examples | [go.dev](https://go.dev/dl/) |

### Optional but Recommended

| Tool | Purpose |
|------|---------|
| **MongoDB Compass** | Visual query/index explorer |
| **cqlsh** | Cassandra query shell (comes with Docker image) |
| **Redis CLI** | For key-value store examples |

---

## Quick Docker Setup

### MongoDB (Replica Set — required for transactions and change streams)

```yaml
# docker-compose-mongo.yml
version: '3.8'
services:
  mongo1:
    image: mongo:7
    container_name: mongo1
    command: mongod --replSet rs0 --bind_ip_all
    ports:
      - "27017:27017"
    volumes:
      - mongo1_data:/data/db

  mongo-init:
    image: mongo:7
    depends_on:
      - mongo1
    entrypoint: >
      mongosh --host mongo1:27017 --eval '
        rs.initiate({
          _id: "rs0",
          members: [{ _id: 0, host: "mongo1:27017" }]
        })
      '

volumes:
  mongo1_data:
```

```bash
docker compose -f docker-compose-mongo.yml up -d
# Wait 10 seconds, then:
mongosh "mongodb://localhost:27017/?replicaSet=rs0"
```

### Cassandra (Single Node)

```yaml
# docker-compose-cassandra.yml
version: '3.8'
services:
  cassandra:
    image: cassandra:4.1
    container_name: cassandra
    ports:
      - "9042:9042"
    environment:
      - CASSANDRA_CLUSTER_NAME=learning
      - CASSANDRA_DC=dc1
    volumes:
      - cassandra_data:/var/lib/cassandra

volumes:
  cassandra_data:
```

```bash
docker compose -f docker-compose-cassandra.yml up -d
# Cassandra takes ~60 seconds to start. Then:
docker exec -it cassandra cqlsh
```

### Redis (for Phase 1 examples)

```bash
docker run -d --name redis -p 6379:6379 redis:7-alpine
docker exec -it redis redis-cli
```

---

## TypeScript Project Setup

```bash
mkdir nosql-curriculum && cd nosql-curriculum
npm init -y
npm install typescript tsx @types/node -D
npm install mongodb cassandra-driver redis
npx tsc --init --target ES2022 --module NodeNext --moduleResolution NodeNext
```

**tsconfig.json** adjustments:
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "strict": true,
    "outDir": "./dist",
    "rootDir": "./src"
  }
}
```

Run TypeScript files with:
```bash
npx tsx src/your-file.ts
```

---

## Go Project Setup

```bash
mkdir nosql-curriculum-go && cd nosql-curriculum-go
go mod init nosql-curriculum
go get go.mongodb.org/mongo-driver/mongo
go get github.com/gocql/gocql
go get github.com/redis/go-redis/v9
```

---

## Verify Everything Works

### MongoDB
```bash
mongosh "mongodb://localhost:27017/?replicaSet=rs0" --eval 'db.test.insertOne({hello: "nosql"}); db.test.findOne()'
```

### Cassandra
```bash
docker exec -it cassandra cqlsh -e "CREATE KEYSPACE IF NOT EXISTS test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}; USE test; CREATE TABLE IF NOT EXISTS hello (id int PRIMARY KEY, msg text); INSERT INTO hello (id, msg) VALUES (1, 'nosql'); SELECT * FROM hello;"
```

### Redis
```bash
docker exec -it redis redis-cli SET hello "nosql" && docker exec -it redis redis-cli GET hello
```

If all three return data, you're ready. Move to **Phase 0**.

---

## A Note on "Production Setup"

Everything here runs on a single machine with Docker. This is intentional — you're learning data modeling and tradeoffs, not operations.

When we discuss replication, partitioning, and failure, we'll explain the concepts with diagrams and narratives. You don't need a 5-node cluster to understand why partition keys matter.
