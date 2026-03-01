# NoSQL Databases — Deep Curriculum

> A complete, opinionated, teaching-first curriculum for NoSQL databases.
> Written for engineers who already know SQL and want to understand NoSQL idiomatically — not superficially.

---

## Who This Is For

- You already understand SQL well (joins, normalization, ACID, query planning)
- You've used MongoDB for basic CRUD
- You want to understand **why** NoSQL exists, **when** it's the right tool, and **what you lose**
- You care about tradeoffs, data modeling, performance, and correctness
- You are NOT looking for a command reference or a marketing pitch

---

## What This Is NOT

- A MongoDB tutorial
- A "NoSQL is better than SQL" argument
- A product comparison chart
- A syntax reference

This curriculum teaches you **how databases think** — what they optimize for, what they sacrifice, and why those decisions matter for the systems you build.

---

## Directory Structure

```
NOSQL/
├── README.md                              ← You are here
├── START_HERE.md                          ← How to use this curriculum
├── GETTING_STARTED.md                     ← Environment setup
├── QUICK_REFERENCE.md                     ← Key concepts cheat sheet
│
├── 00-why-nosql-exists/                   ← The historical pressure that created NoSQL
│   ├── 01-limits-of-relational-at-scale.md
│   ├── 02-cap-theorem-without-handwaving.md
│   └── 03-eventual-consistency-as-feature.md
│
├── 01-nosql-taxonomy/                     ← Mental map of NoSQL families
│   ├── 01-document-stores.md
│   ├── 02-wide-column-stores.md
│   ├── 03-key-value-stores.md
│   ├── 04-graph-databases.md
│   └── 05-time-series-databases.md
│
├── 02-mongodb-deep-dive/                  ← Document thinking (beyond CRUD)
│   ├── 01-document-vs-relational-modeling.md
│   ├── 02-embedding-vs-referencing.md
│   ├── 03-schema-design-for-reads.md
│   ├── 04-schema-versioning.md
│   ├── 05-index-design.md
│   ├── 06-aggregation-pipeline.md
│   ├── 07-joins-and-lookups.md
│   ├── 08-transactions-when-they-matter.md
│   ├── 09-write-read-concerns.md
│   └── 10-performance-debugging.md
│
├── 03-cassandra-deep-dive/                ← Modeling around queries
│   ├── 01-partition-and-clustering-keys.md
│   ├── 02-no-adhoc-queries.md
│   ├── 03-table-per-query.md
│   ├── 04-denormalization-as-requirement.md
│   ├── 05-tunable-consistency.md
│   ├── 06-read-write-paths.md
│   ├── 07-compaction-strategies.md
│   └── 08-when-cassandra-is-wrong.md
│
├── 04-consistency-replication-failure/    ← The uncomfortable truths
│   ├── 01-eventual-vs-strong-consistency.md
│   ├── 02-quorum-reads-writes.md
│   ├── 03-vector-clocks-and-conflict.md
│   ├── 04-read-repair-and-anti-entropy.md
│   └── 05-split-brain-and-node-failure.md
│
├── 05-data-modeling-patterns/             ← Patterns, not tables
│   ├── 01-embedding-patterns.md
│   ├── 02-fan-out-patterns.md
│   ├── 03-bucket-pattern.md
│   ├── 04-time-series-modeling.md
│   ├── 05-materialized-views.md
│   └── 06-outbox-change-streams-events.md
│
├── 06-performance-and-scale/              ← Reality of production
│   ├── 01-hot-partitions.md
│   ├── 02-write-amplification.md
│   ├── 03-index-cost.md
│   ├── 04-disk-memory-tradeoffs.md
│   └── 05-why-benchmarks-lie.md
│
├── 07-when-not-to-use-nosql/              ← The mandatory warning
│   ├── 01-accidental-nosql.md
│   ├── 02-over-denormalization.md
│   ├── 03-query-regret.md
│   └── 04-rebuilding-sql-manually.md
│
└── 08-choosing-the-right-database/        ← Decision frameworks
    ├── 01-access-patterns-first.md
    ├── 02-consistency-requirements.md
    └── 03-decision-framework.md
```

---

## Phases Overview

| Phase | Focus | Key Shift |
|-------|-------|-----------|
| 0 | Why NoSQL Exists | From "databases store data" to "databases make tradeoffs" |
| 1 | NoSQL Taxonomy | From "NoSQL = MongoDB" to "NoSQL = families with different DNA" |
| 2 | MongoDB Deep Dive | From "documents are flexible" to "documents encode access patterns" |
| 3 | Cassandra Deep Dive | From "I'll just query it" to "I must design for the query" |
| 4 | Consistency & Failure | From "data is safe" to "data is a negotiation" |
| 5 | Data Modeling Patterns | From "how do I store this?" to "how will I read this?" |
| 6 | Performance & Scale | From "it's fast" to "it's fast *for this workload*" |
| 7 | When NOT to Use NoSQL | From "NoSQL solves scale" to "NoSQL trades problems" |
| 8 | Choosing the Right DB | From "which is best?" to "what guarantees do I need?" |

---

## Time Estimate

| Phase | Time (focused study) |
|-------|---------------------|
| 0     | ~4–5 hours          |
| 1     | ~4–5 hours          |
| 2     | ~12–15 hours        |
| 3     | ~10–12 hours        |
| 4     | ~6–8 hours          |
| 5     | ~6–8 hours          |
| 6     | ~5–6 hours          |
| 7     | ~3–4 hours          |
| 8     | ~3–4 hours          |
| **Total** | **~55–65 hours** |

---

## Philosophy

1. **Teach by contrast** — Every NoSQL decision is contrasted with what SQL would do
2. **Start from pain** — Every concept begins with a real problem, not a definition
3. **No vendor marketing** — No "flexible schemas" or "web-scale" fluff
4. **Depth over syntax** — Commands serve decisions, not the other way around
5. **Mistakes are teachers** — Bad designs are shown, analyzed, and corrected

---

## What You'll Be Able to Do After

- Stop asking "Which NoSQL DB is best?"
- Start asking "What guarantees do I need?"
- Model data intentionally for your access patterns
- Understand why NoSQL feels powerful *and* dangerous
- Explain tradeoffs confidently in system design interviews
- Know when to use SQL instead — and defend that choice
