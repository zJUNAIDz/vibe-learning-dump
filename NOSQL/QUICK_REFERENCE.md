# Quick Reference — NoSQL Key Concepts

> This is a cheat sheet, not a learning resource. Read the curriculum first.

---

## CAP Theorem Summary

| Property | Meaning | You give up |
|----------|---------|-------------|
| **Consistency** | Every read sees the latest write | Availability during partition |
| **Availability** | Every request gets a response | Consistency during partition |
| **Partition Tolerance** | System works despite network splits | *(Cannot be dropped)* |

**Reality:** You always have P. You choose between C and A during failures.

---

## NoSQL Family Cheat Sheet

| Family | Optimized For | Bad At | Example |
|--------|--------------|--------|---------|
| Document | Flexible reads on nested data | Cross-document joins | MongoDB |
| Wide-Column | High-throughput writes, time-series | Ad-hoc queries | Cassandra |
| Key-Value | Simple lookups by key | Complex queries | Redis, DynamoDB |
| Graph | Relationship traversals | Aggregations, bulk scans | Neo4j |
| Time-Series | Append-heavy temporal data | Random updates | InfluxDB |

---

## MongoDB Quick Reference

### Consistency Controls
| Setting | Meaning |
|---------|---------|
| `w: 1` | Acknowledged by primary only |
| `w: "majority"` | Acknowledged by majority of replicas |
| `readConcern: "majority"` | Read only data committed on majority |
| `readConcern: "linearizable"` | Strongest read guarantee |
| `readPreference: "secondary"` | Read from secondaries (eventual) |

### Index Types
| Type | Use Case |
|------|----------|
| Single field | Simple equality/range queries |
| Compound | Multi-field queries (order matters!) |
| Multikey | Array fields |
| Text | Full-text search |
| TTL | Auto-expire documents |
| Partial | Index only matching documents |
| Wildcard | Dynamic/unknown field names |

### Aggregation Pipeline Stages (Most Used)
```
$match → $group → $project → $sort → $limit → $lookup → $unwind → $addFields
```

---

## Cassandra Quick Reference

### Key Concepts
| Term | Meaning |
|------|---------|
| **Partition Key** | Determines which node stores the data |
| **Clustering Key** | Determines sort order within a partition |
| **Primary Key** | Partition key + clustering key(s) |
| **Token** | Hash of partition key → ring position |

### Consistency Levels
| Level | Reads/Writes | Guarantee |
|-------|-------------|-----------|
| `ONE` | 1 replica | Fastest, weakest |
| `QUORUM` | ⌊N/2⌋ + 1 replicas | Strong if R + W > N |
| `ALL` | All replicas | Strongest, least available |
| `LOCAL_QUORUM` | Quorum in local DC | Multi-DC friendly |

### Anti-Patterns
- ❌ `SELECT *` with no partition key (full cluster scan)
- ❌ `ALLOW FILTERING` in production
- ❌ Large partitions (> 100MB)
- ❌ Too many tombstones
- ❌ Using Cassandra for ad-hoc analytics

---

## Consistency Formulas

```
Strong Consistency: R + W > N

Where:
  R = read replicas consulted
  W = write replicas acknowledged
  N = total replicas
```

| R | W | N | Strong? |
|---|---|---|---------|
| 2 | 2 | 3 | ✅ Yes (2+2 > 3) |
| 1 | 1 | 3 | ❌ No (1+1 < 3) |
| 3 | 1 | 3 | ✅ Yes (3+1 > 3) |
| 1 | 3 | 3 | ✅ Yes (1+3 > 3) |

---

## Data Modeling Decision Tree

```
1. What are my queries?
   ├── Single document lookup? → Document store or KV
   ├── Time-range scans? → Wide-column or time-series
   ├── Relationship traversals? → Graph
   └── Complex joins & aggregations? → Stay with SQL

2. What consistency do I need?
   ├── Strong (financial, inventory)? → SQL or tunable consistency
   └── Eventual (social feeds, logs)? → NoSQL native

3. What's my write pattern?
   ├── Write-heavy, append-only? → Cassandra, time-series
   ├── Read-heavy, complex reads? → MongoDB, SQL
   └── Simple get/set? → Redis, DynamoDB

4. Can I predict my access patterns?
   ├── Yes → NoSQL (model for those patterns)
   └── No → SQL (ad-hoc queries welcome)
```

---

## Common Embedding vs. Referencing Decision

```
Embed when:
  - Data is read together
  - Child has no independent identity
  - Child count is bounded (< ~100)
  - Updates are infrequent

Reference when:
  - Data is large (> 16MB document limit)
  - Many independent consumers
  - Child count is unbounded
  - Frequent independent updates
```
