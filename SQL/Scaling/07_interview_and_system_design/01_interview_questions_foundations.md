# Interview Questions: Foundations

This is where we separate the pretenders from the contenders. A candidate who can't answer these questions clearly and concisely does not have a solid foundation in distributed systems. These aren't "gotcha" questions; they are the absolute basics.

The format here is simple: Question, What I'm Listening For, and a Red Flag answer.

---

### 1. What's the difference between vertical and horizontal scaling?

*   **What I'm Listening For:**
    *   A clear, simple definition. Vertical scaling = bigger machine (more CPU, RAM). Horizontal scaling = more machines.
    *   Mention of the limitations of vertical scaling (there's a physical limit to how big one machine can get, and it gets exponentially more expensive).
    *   Mention of the challenges of horizontal scaling (it introduces complexity; the software must be designed to run on multiple machines).
    *   Bonus points for mentioning that you often do both.

*   **Red Flag Answer:**
    *   "Umm, one is like, up, and the other is like, across?" (Too vague, shows they memorized a shape, not a concept).
    *   "Horizontal scaling is always better." (Dogmatic, lacks nuance. The right answer is almost always "it depends").

---

### 2. Explain the CAP Theorem.

*   **What I'm Listening For:**
    *   They can correctly identify the three components: **C**onsistency, **A**vailability, and **P**artition Tolerance.
    *   They can explain that in a distributed system, you *must* choose Partition Tolerance (because networks are unreliable).
    *   Therefore, the real trade-off is between Consistency and Availability.
    *   They can give a simple example. "If the network partitions, do you want to return an error (choosing Consistency) or potentially stale data (choosing Availability)?"
    *   Bonus points for mentioning that the CAP theorem is a simplification and that different systems offer different trade-offs (e.g., eventual consistency).

*   **Red Flag Answer:**
    *   "You can only have two of the three." (This is the classic, but incomplete, answer. It misses the crucial point that Partition Tolerance is non-negotiable in most real-world distributed systems).
    *   Getting the definitions wrong, e.g., confusing "Consistency" in CAP with "Consistency" in ACID.

---

### 3. What is database replication, and why would you use it?

*   **What I'm Listening For:**
    *   A clear definition: "Replication is the process of copying data from a primary (or master) database to one or more replica (or slave) databases."
    *   The two primary reasons for using it:
        1.  **High Availability / Failover:** If the primary fails, you can promote a replica to be the new primary.
        2.  **Read Scaling:** You can direct read queries to the replicas to take the load off the primary.
    *   Mention of the main drawback: **replication lag**.

*   **Red Flag Answer:**
    *   "It's for backups." (While replicas can be used as part of a backup strategy, that's not their primary purpose. This answer misses the point about availability and scaling).
    *   Not being able to explain *why* you'd want to scale reads separately from writes.

---

### 4. What is database sharding?

*   **What I'm Listening For:**
    *   A clear definition: "Sharding is the process of splitting a large database into smaller, more manageable pieces called shards. It's a form of horizontal scaling."
    *   The key concept of a **shard key**, the piece of data used to decide which shard a row of data belongs to.
    *   The primary benefit: It allows you to scale both reads and writes beyond the limits of a single machine.
    *   Mention of the main challenges: shard key selection, hot partitions, and the inability to do joins across shards easily.

*   **Red Flag Answer:**
    *   Confusing sharding with partitioning (partitioning is a feature within a single database instance; sharding is across multiple instances).
    *   "You just split the data randomly." (Shows a lack of understanding of the importance of the shard key and data locality).

---

### 5. A user updates their profile picture, but when the page reloads, they see the old one. What's the most likely cause?

*   **What I'm Listening For:**
    *   The immediate, confident answer: **"Replication lag."**
    *   A clear explanation: "The write to update the profile picture URL went to the primary database. The subsequent read request to fetch the user's profile was served by a read replica, which hadn't yet received the update from the primary."
    *   Bonus points for suggesting a solution: "To fix this, you need to ensure **read-your-writes consistency**. A common strategy is to route the user's reads to the primary for a short period after they've made a write."

*   **Red Flag Answer:**
    *   "A caching issue." (While it *could* be a caching issue at the CDN or application layer, replication lag is a far more fundamental and common cause for this specific problem in a scaled database architecture).
    *   A long, rambling answer that doesn't quickly identify the most probable cause. This is a test of their diagnostic instincts.
