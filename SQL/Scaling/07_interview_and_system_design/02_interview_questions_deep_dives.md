# Interview Questions: Deep Dives

If a candidate nails the foundational questions, it's time to go deeper. These questions probe their understanding of trade-offs, failure modes, and the practical realities of running systems at scale. A good answer here isn't just about knowing the definition; it's about demonstrating experience and good judgment.

---

### 1. You need to choose a shard key for your `users` table. What are the properties of a good shard key? What are some bad choices?

*   **What I'm Listening For:**
    *   **High Cardinality:** The key should have many, many possible values. This allows the data to be spread across a large number of shards.
    *   **Even Distribution:** The key should distribute the workload evenly. A key that leads to hot partitions is a bad key.
    *   **Locality:** The key should group data together that is likely to be accessed together. This is key for minimizing cross-shard queries.
    *   **Bad Choices & Why:**
        *   `country_code`: Low cardinality, terrible distribution (a few countries have billions of users, many have very few). A classic hot partition problem.
        *   `is_active` (boolean): The worst possible cardinality (only two values).
        *   A free-text field like `city_name`: Prone to typos and variations ("SF", "San Francisco") leading to data fragmentation.
    *   **Good Choices & Why:**
        *   `user_id`: High cardinality, perfectly distributed (assuming UUIDs or a good sequence generator). Excellent choice.
        *   `tenant_id` or `organization_id`: This is a fantastic choice for B2B SaaS apps. It provides perfect data locality for each customer. All of a single customer's data lives on one shard, making queries fast and simple. The trade-off is that you might have a "hot tenant" if one customer is much larger than others.

*   **Red Flag Answer:**
    *   "Just use `user_id`." (While often a good choice, it shows they aren't thinking about the trade-offs, especially data locality for specific query patterns).
    *   Not being able to explain *why* a key like `country_code` is a bad idea.

---

### 2. Describe the circuit breaker pattern. Why is it critical in a microservices architecture?

*   **What I'm Listening For:**
    *   A clear explanation of the three states: **Closed, Open, and Half-Open**.
    *   They must explain that when the circuit is **Open**, requests fail *immediately* without hitting the network. This is the most important part.
    *   They must explain the purpose: to prevent a **cascading failure**.
    *   The "why it's critical" part: In a microservices architecture, you have a deep graph of service dependencies. A single slow service can cause all upstream services to block, time out, and fail. The circuit breaker contains the failure to the immediate caller, protecting the rest of the system and giving the slow service time to recover.

*   **Red Flag Answer:**
    *   "It's for retrying requests." (This is a fundamental misunderstanding. The circuit breaker is what *stops* requests).
    *   Describing it as just a simple timeout. It's a state machine that wraps the timeout logic.

---

### 3. What is a "thundering herd" problem, and how do you prevent it?

*   **What I'm Listening For:**
    *   A clear definition: A large number of clients all trying to connect or request a resource at the exact same moment, overwhelming the server.
    *   The classic scenario: A server or database crashes and restarts. All the clients that were waiting immediately try to reconnect simultaneously.
    *   The solution: **Exponential backoff with jitter**. They should be able to explain both parts.
        *   **Exponential Backoff:** Increase the wait time between retries (1s, 2s, 4s, 8s...).
        *   **Jitter:** Add a small, random amount of time to each wait. This is crucial for spreading out the retries so they don't all happen at the same time.

*   **Red Flag Answer:**
    *   "You just add a `sleep()` before you retry." (Misses the "exponential" and "jitter" parts, which are the most important).
    *   Confusing it with a general traffic spike. The thundering herd is specifically about a *synchronized* rush of clients, often after a failure event.

---

### 4. You're designing a social media feed. What are the two main strategies for generating the feed, and what are the trade-offs?

*   **What I'm Listening For:**
    *   The two strategies:
        1.  **Fan-out on Read (Pull Model):** When a user loads their feed, you query the database for all the people they follow, find all their recent posts, and assemble the feed on the fly.
        2.  **Fan-out on Write (Push Model):** When a user makes a post, you find all of their followers and inject the new post directly into each follower's feed (e.g., into a Redis list that represents their timeline).
    *   **Trade-offs:**
        *   **Fan-out on Read:**
            *   Pro: Simple to implement. No work is done until a user requests their feed. Good for users who don't log in often.
            *   Con: Very slow for users who follow many people (the "celebrity problem"). The read load can be immense.
        *   **Fan-out on Write:**
            *   Pro: The feed load is incredibly fast. It's just reading a pre-computed list.
            *   Con: The write load can be massive. If a celebrity with 100 million followers posts, you have to do 100 million writes. This can be slow and resource-intensive.
    *   **The Hybrid Approach:** The best candidates will mention that real-world systems use a hybrid. Normal users use fan-out on write. For celebrities, you don't push to all their followers. Instead, you merge the celebrity's posts into the user's feed at read time.

*   **Red Flag Answer:**
    *   Only being able to describe one of the two methods.
    *   "You just do a `SELECT` with a `JOIN`." (Shows a complete lack of understanding of the performance implications at scale).
    *   Not considering the "celebrity problem" as a key factor in the design.
