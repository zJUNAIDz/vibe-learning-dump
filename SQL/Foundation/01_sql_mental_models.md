# SQL Mental Models: Think Like the Database

## The Fundamental Shift: SQL is Declarative, Not Procedural

If you've been writing JavaScript, Python, or any imperative language for years, SQL will mess with your head. And that's the point.

### What "Declarative" Actually Means

In procedural code, you write **instructions**:
```javascript
// You tell the computer HOW to do it
let total = 0;
for (let user of users) {
  if (user.age > 18) {
    total += user.balance;
  }
}
```

In SQL, you write **specifications**:
```sql
-- You tell the database WHAT you want
SELECT SUM(balance) 
FROM users 
WHERE age > 18;
```

You don't control the loop. You don't manage memory. You don't decide whether to scan or seek. **The optimizer does**. And it's smarter than you think—when it has good statistics.

### The Cost: You Lose Control of "How"

This is why ORM developers freeze. They're used to **tracing execution**:
- "First this line runs, then this line, then..."
- "I can step through with a debugger"
- "I can add a console.log to see what's happening"

In SQL, you can't do that. You specify a **result set**, and the database figures out the cheapest way to produce it. Sometimes it does something brilliant. Sometimes it does something stupid. But you don't *control* it—you *influence* it.

## Mental Model #1: SQL Operates on Sets, Not Rows

### The Set-Based Mindset

SQL descended from relational algebra. A table is a **set** (or multiset, technically). Operations produce sets.

**Wrong mental model:**
```sql
-- "Loop through orders and for each one, grab the customer"
SELECT * FROM orders, customers WHERE orders.customer_id = customers.id;
```

**Correct mental model:**
```sql
-- "Give me the Cartesian product of orders × customers, 
--  filtered to matching pairs"
SELECT * FROM orders JOIN customers ON orders.customer_id = customers.id;
```

It's not iterating. It's **joining sets**. The database might use nested loops, hash joins, or merge joins—but that's an implementation detail hidden behind the declarative surface.

### Why This Matters

When you write:
```sql
SELECT COUNT(*) FROM users WHERE active = true;
```

You're not saying:
- "Initialize a counter"
- "Loop through users"
- "Increment if active"

You're saying:
- "Count the cardinality of the set {u ∈ users | u.active = true}"

The database might:
- Scan an index on `active` and count entries
- Do a sequential scan with a bitmap filter
- Use table statistics if it has them
- Parallelize the scan across workers

**You don't decide. The optimizer does.**

## Mental Model #2: The Optimizer Is Your Frenemy

### How the Optimizer Thinks

The query optimizer is a **cost-based decision engine**:

1. It parses your SQL into a logical plan
2. It generates multiple physical plans (scan? index? join order?)
3. It estimates the **cost** of each plan (I/O, CPU, memory)
4. It picks the cheapest one

**The catch:** Cost estimates depend on statistics. If your statistics are stale, the optimizer is flying blind.

### Example: When the Optimizer Outsmarts You

You write:
```sql
SELECT * FROM orders WHERE status = 'pending' AND created_at > NOW() - INTERVAL '7 days';
```

You might think: "It should use the index on `created_at` first to narrow it down."

The optimizer might think: "Only 0.01% of orders are pending. I'll scan the `status` index, then filter by date."

**It chose differently because it knows the data distribution.** This is why blindly adding indexes doesn't always help.

### When the Optimizer Lies

The optimizer isn't omniscient. Common failure modes:

1. **Correlated subqueries**: Optimizer often underestimates cost
2. **Complex joins**: Wrong join order can explode intermediate results
3. **Cross-schema queries**: Statistics might not transfer
4. **Out-of-date statistics**: The optimizer thinks your table has 1000 rows when it has 10 million

**You can't control the optimizer, but you can influence it** with indexes, query structure, and hints (use sparingly).

## Mental Model #3: "Order of Writing" ≠ "Order of Execution"

This is where ORM developers get burned.

You write:
```sql
SELECT user_id, COUNT(*) as order_count
FROM orders
WHERE order_count > 5
GROUP BY user_id;
```

**ERROR:** `column "order_count" does not exist`

Wait, what? You just defined it!

### The Logical Execution Order

SQL processes queries in this order (we'll cover this deeply in the next file):

```
FROM       → identifies source tables
WHERE      → filters rows BEFORE grouping
GROUP BY   → aggregates rows into groups
HAVING     → filters groups AFTER grouping
SELECT     → projects columns and aliases
ORDER BY   → sorts result set
LIMIT      → restricts result set
```

When you write `WHERE order_count > 5`, the `SELECT` hasn't happened yet. The alias doesn't exist.

**Correct version:**
```sql
SELECT user_id, COUNT(*) as order_count
FROM orders
GROUP BY user_id
HAVING COUNT(*) > 5;
```

### Why This Trips Up ORM Users

ORMs let you write:
```javascript
const users = await prisma.user.findMany({
  where: { active: true },
  orderBy: { created_at: 'desc' },
  take: 10
});
```

You read it top-to-bottom. It feels procedural. But under the hood, Prisma generates:
```sql
SELECT * FROM users WHERE active = true ORDER BY created_at DESC LIMIT 10;
```

And **SQL executes it differently than you read it.** The moment you write raw SQL, that mental model breaks.

## Mental Model #4: ORMs Hide the Cost Model

ORMs are *syntactic* abstractions, not *semantic* ones. They make queries **easier to write**, but they don't make them **cheaper to execute**.

### The Hidden N+1 Problem

Prisma/Drizzle code:
```javascript
const users = await prisma.user.findMany();
for (const user of users) {
  const orders = await prisma.order.findMany({ 
    where: { userId: user.id } 
  });
  console.log(`${user.name}: ${orders.length} orders`);
}
```

Looks innocent. But it executes:
```sql
SELECT * FROM users;               -- 1 query
SELECT * FROM orders WHERE user_id = 1;  -- N queries
SELECT * FROM orders WHERE user_id = 2;
SELECT * FROM orders WHERE user_id = 3;
-- ...
```

**One query becomes N+1 queries.** If you have 1000 users, you just hit the database 1001 times.

The SQL way:
```sql
SELECT 
  u.name,
  COUNT(o.id) as order_count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.id, u.name;
```

**One query. One round trip.** The database does the aggregation.

### Why ORMs Hide This

ORMs optimize for **developer ergonomics**, not **database efficiency**. They assume:
- Network latency is low (connection pooling, co-located DB)
- Queries are simple (they're not always)
- Developers will use eager loading strategically (they don't)

In production, this assumption breaks. Hard.

## Mental Model #5: SQL Has No Concept of "Time" (Within a Query)

This is subtle but critical.

### There's No "Then"

In JavaScript:
```javascript
let x = 5;       // First this
let y = x + 1;   // Then this
```

In SQL:
```sql
SELECT 
  price,
  price * 1.1 AS price_with_tax,
  price_with_tax * 0.9 AS discounted_price  -- ERROR!
FROM products;
```

**You can't reference `price_with_tax` in the same SELECT.** Because SELECT is **evaluated simultaneously**, not sequentially.

**Why?** Because SQL is set-based. The projection (SELECT) happens **in one logical step** after grouping/filtering.

**Workaround:**
```sql
SELECT 
  price,
  price * 1.1 AS price_with_tax,
  price * 1.1 * 0.9 AS discounted_price
FROM products;
```

Or use a subquery/CTE:
```sql
WITH prices AS (
  SELECT price, price * 1.1 AS price_with_tax
  FROM products
)
SELECT price, price_with_tax, price_with_tax * 0.9 AS discounted_price
FROM prices;
```

### The Exception: Window Functions (Kind Of)

Window functions *do* have ordering, but only within their `OVER` clause:
```sql
SELECT 
  name,
  salary,
  AVG(salary) OVER (ORDER BY hire_date ROWS BETWEEN 2 PRECEDING AND CURRENT ROW)
FROM employees;
```

But that's a special case. The general rule holds: **SQL clauses execute logically, not temporally.**

## Common Beginner Traps (and Why They Happen)

### Trap #1: Treating NULL Like a Value

```sql
SELECT * FROM users WHERE email = NULL;  -- Returns nothing!
```

NULL means "unknown." You can't compare to unknown. You check for it:
```sql
SELECT * FROM users WHERE email IS NULL;
```

**Why this happens:** Most languages treat `null` as a value. SQL treats NULL as **the absence of a value**. Different semantics.

### Trap #2: Misunderstanding COUNT

```sql
SELECT COUNT(email) FROM users;  -- Counts non-NULL emails
SELECT COUNT(*) FROM users;      -- Counts rows
```

If `email` is nullable, these return different numbers.

**Why this happens:** Developers assume COUNT counts rows. It counts **non-NULL values** of the specified expression.

### Trap #3: Implicit GROUP BY Disasters

```sql
-- MySQL (with certain modes) allows this:
SELECT user_id, order_id FROM orders GROUP BY user_id;
```

What's `order_id` here? **Undefined.** Postgres and modern MySQL reject this. Old MySQL picks an arbitrary one.

**Why this happens:** SQL Standard says non-aggregated columns in SELECT must be in GROUP BY. Some databases were lenient (and wrong).

## The Brutal Truth: ORMs Make You Lazy

Here's what ORMs do well:
- Type safety (if using TypeScript + Prisma)
- Schema migrations (DrizzleKit, Prisma Migrate)
- Basic CRUD with minimal boilerplate
- Parameterized queries (SQL injection protection)

Here's what ORMs hide:
- **Query cost**: You don't see the EXPLAIN plan
- **Join strategies**: You don't know if it's doing nested loops or hash joins
- **Index usage**: You don't know if your query is scanning 10 rows or 10 million
- **Lock contention**: You don't know what locks you're holding
- **Network overhead**: You don't see N+1 queries happening

### The Uncomfortable Question

Would you ship a JavaScript app without checking the browser's Network tab?

Then why ship database queries without checking `EXPLAIN ANALYZE`?

## How to Think Like the Database

### 1. Start With the Data Model

Before writing a query, visualize:
- What tables are involved?
- What's the cardinality? (1:1, 1:many, many:many)
- What's the join key?

**Draw it out.** Seriously. ASCII art helps:

```
users (1) ----< (N) orders (N) >---- (1) products
           user_id            product_id
```

### 2. Think in Sets, Not Loops

Don't think: "For each user, find their orders."

Think: "Join the set of users with the set of orders."

### 3. Predict the Cost

Before running a query, ask:
- Will this use an index?
- How many rows will it scan?
- What's the join cardinality?

Then run `EXPLAIN` and see if you were right.

### 4. Understand the Optimizer's Constraints

The optimizer is smart, but:
- It only knows what statistics tell it
- It can't predict business logic
- It doesn't know your app's access patterns

**Your job:** Give it good data (statistics, indexes) and don't write queries it can't optimize.

## A Real-World Example

Let's say you need: "All users who placed an order in the last 30 days."

### ORM Way (Prisma)
```javascript
const recentOrders = await prisma.order.findMany({
  where: { createdAt: { gte: new Date(Date.now() - 30 * 86400000) } },
  include: { user: true }
});
const activeUsers = [...new Set(recentOrders.map(o => o.user))];
```

**What this does:**
1. Fetches ALL orders from last 30 days (could be millions)
2. JOINs users table for EACH order
3. Transfers all that data to your app
4. De-duplicates in JavaScript memory

**Cost:** High network transfer, high memory usage, JOIN overhead.

### SQL Way
```sql
SELECT DISTINCT u.*
FROM users u
JOIN orders o ON u.id = o.user_id
WHERE o.created_at >= NOW() - INTERVAL '30 days';
```

**What this does:**
1. Database scans orders (using index on `created_at` if it exists)
2. JOINs users (hash join or nested loop, optimizer decides)
3. De-duplicates on the database side
4. Returns only the user rows

**Cost:** Minimal network transfer, database does heavy lifting.

### Even Better SQL Way
```sql
SELECT u.*
FROM users u
WHERE EXISTS (
  SELECT 1 FROM orders o 
  WHERE o.user_id = u.id 
    AND o.created_at >= NOW() - INTERVAL '30 days'
);
```

**Why it's better:**
- EXISTS short-circuits (stops at first match)
- No need for DISTINCT (users are already unique)
- Optimizer can use semi-join strategies

**This is the kind of thinking ORMs don't teach you.**

## Closing Thoughts

SQL is frustrating *because it's different*. It's not a worse version of JavaScript. It's a different **way of thinking about data**.

The path forward:
1. Accept that you don't control execution
2. Learn to influence the optimizer (indexes, query structure)
3. Understand cost models (we'll cover EXPLAIN in detail)
4. Use ORMs when they're appropriate, raw SQL when they're not

**Next up:** We'll nail down query execution order, because that's where 90% of SQL bugs come from.
