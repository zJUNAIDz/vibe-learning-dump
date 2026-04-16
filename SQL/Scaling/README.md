# Database Scaling — From Zero to Production-Ready

This repository is a complete, production-grade learning guide to database scaling. It's designed to take you from the absolute basics of why we need to scale to the advanced, gnarly problems you'll face in real-world distributed systems.

This isn't a dry, academic textbook. It's a brutally honest, practical guide written in the tone of a senior engineer mentoring you at 2 AM. We'll cover the good, the bad, and the ugly of making databases handle massive load.

## The Philosophy

*   **Mental Models First:** Understand the *why* before the *how*.
*   **Explain Why Things Break:** Failure is the best teacher. We'll focus on failure modes, footguns, and production horror stories.
*   **The Machine's Perspective:** What is the hardware *actually* doing? No black boxes.
*   **Practical & Raw:** No fluffy corporate language. Just the stuff that matters in production.

## Repository Structure

```text
.
├── 00_foundations/
│   ├── 01_why_scaling_exists.md
│   ├── 02_vertical_vs_horizontal.md
│   └── 03_latency_bandwidth_mental_models.md
├── 01_sql_scaling/
│   ├── 01_why_sql_is_harder_to_scale.md
│   ├── 02_distributed_joins.md
│   ├── 03_distributed_transactions.md
│   └── 04_acid_tradeoffs.md
├── 02_replication/
│   ├── 01_primary_replica_architecture.md
│   ├── 02_replication_lag_and_stale_reads.md
│   └── 03_failover_and_split_brain.md
├── 03_sharding/
│   ├── 01_sharding_basics.md
│   ├── 02_shard_keys_and_hot_partitions.md
│   ├── 03_consistent_hashing_and_rebalancing.md
│   └── 04_sharding_migration_strategies.md
├── 04_nosql_comparison/
│   ├── 01_why_nosql_shards_easily.md
│   ├── 02_cap_theorem_in_practice.md
│   └── 03_consistency_models.md
├── 05_real_system_design/
│   ├── 01_designing_a_chat_app.md
│   ├── 02_designing_an_e-commerce_platform.md
│   └── 03_designing_a_social_media_feed.md
├── 06_failure_modes/
│   ├── 01_common_db_failure_patterns.md
│   └── 02_cascading_failures_and_retry_storms.md
├── 07_interview_and_system_design/
│   ├── 01_interview_questions_foundations.md
│   └── 02_system_design_prompts.md
└── README.md
```

## How to Use This Repository

1.  **Start from the beginning.** The concepts build on each other. Don't jump straight to sharding if you don't understand replication lag.
2.  **Read for intuition.** Focus on the mental models and analogies.
3.  **Study the diagrams.** They are not just for show; they are core to the explanations.
4.  **Engage with the examples.** Think about how the SQL queries would change, what would break, and what the error messages would look like.
5.  **Treat the "Production Gotcha" and "Interview Note" sections as gold.** This is the distilled experience from years of building and breaking things.

Let's begin.
