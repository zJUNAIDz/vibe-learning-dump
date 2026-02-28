# JOINs Demystified: What Your ORM Isn't Telling You

## The Mental Model: Joins Are Set Operations

When you write a JOIN, you're not "looping through one table and looking up the other." You're combining **two sets** according to a condition.

```
users (set A)  ×  orders (set B)  →  combined set
```

How they combine depends on the join type:
- **INNER JOIN:** Intersection (rows that match in both tables)
- **LEFT JOIN:** All of A, matched with B where possible
- **RIGHT JOIN:** All of B, matched with A where possible
- **FULL OUTER JOIN:** Union (all rows from both, matched where possible)
- **CROSS JOIN:** Cartesian product (every combination)

## INNER JOIN: The Default

### What It Does

```sql
SELECT u.name, o.total
FROM users u
INNER JOIN orders o ON u.id = o.user_id;
```

**Result:** Only users who have orders. Only orders that belong to a user.

**Venn diagram:**
```
   users          orders
    ┌─────┐      ┌─────┐
    │     │      │     │
    │  ┌──┼──────┼──┐  │
    │  │  │      │  │  │
    └──┼──┘      └──┼──┘
       │   INNER    │
       └────────────┘
```

**Key insight:** Rows without a match are **excluded** from both tables.

### When to Use

Use INNER JOIN when:
- You only care about **related** rows
- Missing relationships are errors or irrelevant
- Example: "Show orders with their product details"

### Common Mistake: INNER JOIN When You Meant LEFT JOIN

**Problem:** You want all users and their order count:
```sql
SELECT u.name, COUNT(o.id) AS order_count
FROM users u
INNER JOIN orders o ON u.id = o.user_id
GROUP BY u.id, u.name;
```

**Bug:** Users without orders won't appear!

**Fix:** Use LEFT JOIN:
```sql
SELECT u.name, COUNT(o.id) AS order_count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.id, u.name;
```

Now users with 0 orders show up with `order_count = 0`.

## LEFT JOIN (LEFT OUTER JOIN): Keep All From the Left

### What It Does

```sql
SELECT u.name, o.total
FROM users u
LEFT JOIN orders o ON u.id = o.user_id;
```

**Result:**
- All users (even those without orders)
- Orders are NULLed out for users without orders

**Venn diagram:**
```
   users          orders
    ┌─────────┐  ┌─────┐
    │         │  │     │
    │  ┌──────┼──┼──┐  │
    │  │      │  │  │  │
    └──┼──────┘  └──┼──┘
       │ LEFT      │
       └───────────┘
```

### Behavior with NULLs

For users without orders:
```
name     | total
---------|-------
Alice    | 100
Bob      | 200
Charlie  | NULL    ← Charlie has no orders
```

**This is critical for COUNT:**
```sql
COUNT(o.id)      -- Counts non-NULL, so 0 for Charlie
COUNT(*)         -- Counts rows, so 1 for Charlie (the NULL row)
```

### When to Use

Use LEFT JOIN when:
- You want **all rows from the left table**
- Missing relationships should be preserved as NULL
- Example: "All users and their order count (including 0)"

### Common Mistake: Filtering NULLs in WHERE

**Problem:**
```sql
SELECT u.name, o.total
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE o.status = 'completed';
```

**Bug:** This turns into an INNER JOIN! Why?

Because `WHERE o.status = 'completed'` filters out rows where `o.status IS NULL` (users without orders). The LEFT JOIN's purpose is defeated.

**Fix:** Put the filter in the ON clause (if you want to filter the right table before joining):
```sql
SELECT u.name, o.total
FROM users u
LEFT JOIN orders o ON u.id = o.user_id AND o.status = 'completed';
```

Now users without completed orders still appear (with NULL totals).

## ON vs WHERE: The Critical Difference

### For INNER JOIN: No Practical Difference

```sql
-- These are equivalent for INNER JOIN:
SELECT * FROM users u
INNER JOIN orders o ON u.id = o.user_id AND o.status = 'completed';

SELECT * FROM users u
INNER JOIN orders o ON u.id = o.user_id
WHERE o.status = 'completed';
```

The optimizer will rewrite them identically.

### For LEFT JOIN: Totally Different

```sql
-- Filters orders BEFORE joining (users without completed orders show up)
SELECT * FROM users u
LEFT JOIN orders o ON u.id = o.user_id AND o.status = 'completed';

-- Filters AFTER joining (users without completed orders are excluded)
SELECT * FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE o.status = 'completed';
```

**Rule of thumb:**
- **ON:** Conditions that define how tables relate
- **WHERE:** Conditions that filter the final result

### Visual Example

**Schema:**
```
users: (1, 'Alice'), (2, 'Bob')
orders: (101, 1, 'completed'), (102, 1, 'pending')
```

**Query 1:** ON with condition
```sql
SELECT u.name, o.id
FROM users u
LEFT JOIN orders o ON u.id = o.user_id AND o.status = 'completed';
```

**Result:**
```
name  | order_id
------|----------
Alice | 101
Bob   | NULL       ← Bob has no completed orders
```

**Query 2:** WHERE with condition
```sql
SELECT u.name, o.id
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE o.status = 'completed';
```

**Result:**
```
name  | order_id
------|----------
Alice | 101
```

Bob disappears! WHERE filters out NULL `o.status`.

## RIGHT JOIN: LEFT JOIN in Reverse

```sql
SELECT u.name, o.total
FROM users u
RIGHT JOIN orders o ON u.id = o.user_id;
```

**Result:** All orders (even those without a user).

**In practice:** Almost never used. People prefer LEFT JOIN and flip the table order:
```sql
-- Instead of RIGHT JOIN users, do LEFT JOIN orders:
SELECT u.name, o.total
FROM orders o
LEFT JOIN users u ON o.user_id = u.id;
```

Same result, clearer intent.

## FULL OUTER JOIN: Keep All From Both

```sql
SELECT u.name, o.total
FROM users u
FULL OUTER JOIN orders o ON u.id = o.user_id;
```

**Result:**
- All users (even those without orders)
- All orders (even those without users)

**Rare in practice.** Usually indicates a data quality issue (orphaned records).

**Venn diagram:**
```
   users          orders
    ┌─────────┐  ┌─────────┐
    │         │  │         │
    │  ┌──────┼──┼──────┐  │
    │  │      │  │      │  │
    └──┼──────┘  └──────┼──┘
       └─────────────────┘
```

## CROSS JOIN: Cartesian Product

```sql
SELECT u.name, p.name
FROM users u
CROSS JOIN products p;
```

**Result:** Every user paired with every product.

If you have 1000 users and 500 products, you get **500,000 rows**.

**When to use:**
- Generating combinations (e.g., all user-product pairs for a recommendation baseline)
- Rarely in production queries

**Accidental CROSS JOINs are a common performance disaster:**
```sql
-- OOPS! Forgot the ON clause
SELECT * FROM users u, orders o;  -- Cartesian product!
```

## Join Cardinality: Predicting Result Size

Understanding cardinality prevents join explosions.

### 1:1 Join (One-to-One)

```
users (id PK) ↔ user_profiles (user_id UNIQUE FK)
```

**Cardinality:** Each user has at most one profile.

**Result size:** At most the size of the smaller table (with INNER JOIN).

### 1:N Join (One-to-Many)

```
users (id PK) → orders (user_id FK)
```

**Cardinality:** Each user can have many orders.

**Result size:** Number of orders (with INNER JOIN).

**Example:**
- 10 users
- 100 orders
- Result: 100 rows (each order gets its user's data repeated)

### N:1 Join (Many-to-One)

Same as 1:N but from the other direction:
```
orders (user_id FK) ← users (id PK)
```

### M:N Join (Many-to-Many)

```
users ←→ user_roles ←→ roles
```

**Cardinality:** Each user can have many roles; each role can belong to many users.

**Implementation:** Requires a junction table:
```sql
SELECT u.name, r.role_name
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id;
```

**Result size:** Number of relationships in `user_roles`.

### Join Explosion: The Silent Killer

**Problem:** Joining multiple 1:N tables multiplies rows.

```sql
SELECT u.name, o.id AS order_id, a.street AS address
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
LEFT JOIN addresses a ON u.id = a.user_id;
```

If a user has 3 orders and 2 addresses, they appear **6 times** (3 × 2).

**Why?** You're creating a Cartesian product of orders × addresses for each user.

**Symptoms:**
- Unexpectedly large result set
- Aggregates (COUNT, SUM) are wrong
- Performance tanks

**Fix #1:** Aggregate separately with subqueries
```sql
SELECT 
  u.name,
  (SELECT COUNT(*) FROM orders WHERE user_id = u.id) AS order_count,
  (SELECT COUNT(*) FROM addresses WHERE user_id = u.id) AS address_count
FROM users u;
```

**Fix #2:** Use window functions (we'll cover this later)

**Fix #3:** Join to aggregated subqueries
```sql
SELECT u.name, o.order_count, a.address_count
FROM users u
LEFT JOIN (
  SELECT user_id, COUNT(*) AS order_count 
  FROM orders 
  GROUP BY user_id
) o ON u.id = o.user_id
LEFT JOIN (
  SELECT user_id, COUNT(*) AS address_count 
  FROM addresses 
  GROUP BY user_id
) a ON u.id = a.user_id;
```

## Self-Joins: When a Table Joins Itself

**Use case:** Hierarchical data (employees and managers).

**Schema:**
```sql
CREATE TABLE employees (
  id INT PRIMARY KEY,
  name TEXT,
  manager_id INT REFERENCES employees(id)
);
```

**Query:** "Show employees and their manager's name"
```sql
SELECT 
  e.name AS employee,
  m.name AS manager
FROM employees e
LEFT JOIN employees m ON e.manager_id = m.id;
```

**Note:** LEFT JOIN because the CEO has no manager (NULL).

**Common use cases:**
- Org charts (employee → manager)
- Bill of materials (part → subpart)
- Social graphs (user → friend)

## Semi-Joins: EXISTS and IN

### Semi-Join: "Existence Check"

**Goal:** Find users who have at least one order.

**Wrong way (with JOIN and DISTINCT):**
```sql
SELECT DISTINCT u.name
FROM users u
INNER JOIN orders o ON u.id = o.user_id;
```

**Problem:**
- If a user has 100 orders, they're joined 100 times
- Then DISTINCT de-duplicates them
- Wasteful!

**Right way (with EXISTS):**
```sql
SELECT u.name
FROM users u
WHERE EXISTS (
  SELECT 1 FROM orders o WHERE o.user_id = u.id
);
```

**Why it's better:**
- EXISTS short-circuits (stops at first match)
- No need for DISTINCT
- Database can use semi-join algorithms

**Alternatively, IN:**
```sql
SELECT u.name
FROM users u
WHERE u.id IN (SELECT user_id FROM orders);
```

**EXISTS vs IN:**
- **EXISTS:** Short-circuits, better for correlated subqueries
- **IN:** Simpler syntax, can hit issues with NULLs

### Anti-Join: NOT EXISTS and NOT IN

**Goal:** Find users who have never placed an order.

**Wrong way:**
```sql
SELECT u.name
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE o.id IS NULL;
```

Works, but verbose.

**Right way (with NOT EXISTS):**
```sql
SELECT u.name
FROM users u
WHERE NOT EXISTS (
  SELECT 1 FROM orders o WHERE o.user_id = u.id
);
```

**Why it's better:**
- Clearer intent
- Database can optimize with anti-join algorithms

**NOT IN has a NULL trap:**
```sql
SELECT u.name
FROM users u
WHERE u.id NOT IN (SELECT user_id FROM orders);
```

**Bug:** If `orders.user_id` contains NULL, this returns **zero rows**!

**Why?** Three-valued logic (we'll cover this in the NULL chapter). 

**Safe version:** Use NOT EXISTS, or:
```sql
WHERE u.id NOT IN (SELECT user_id FROM orders WHERE user_id IS NOT NULL)
```

## Join Strategy: How the Database Does It

This is implementation detail, but it helps you understand EXPLAIN output.

### 1. Nested Loop Join

**How it works:**
```
for each row in table A:
    for each row in table B:
        if join condition matches:
            emit row
```

**When it's used:**
- Small outer table
- Inner table has an index on the join key
- Good for 1:N joins with small N

**Performance:**
- Cost: O(rows_A * rows_B), but optimized with indexes

### 2. Hash Join

**How it works:**
1. Build a hash table from the smaller table
2. Probe the hash table with rows from the larger table

**When it's used:**
- Large tables
- Equijoin (=) condition
- No suitable indexes

**Performance:**
- Cost: O(rows_A + rows_B)
- Memory-intensive (hash table must fit in work_mem)

### 3. Merge Join

**How it works:**
1. Sort both tables by join key
2. Merge-scan through both (like mergesort)

**When it's used:**
- Both tables are already sorted (or have indexes)
- Equijoin condition

**Performance:**
- Cost: O(rows_A + rows_B) if pre-sorted
- Otherwise O(rows_A log rows_A + rows_B log rows_B) for sorting

### Which Is Fastest?

**It depends on:**
- Table size
- Index availability
- Data distribution
- Memory (work_mem)

**The optimizer chooses.** Your job: Give it good indexes and statistics.

## ORM Join Pitfalls

### Pitfall #1: N+1 Queries

**ORM code (Prisma):**
```javascript
const users = await prisma.user.findMany();
for (const user of users) {
  const orders = await prisma.order.findMany({ where: { userId: user.id } });
  console.log(`${user.name}: ${orders.length}`);
}
```

**SQL generated:**
```sql
SELECT * FROM users;
SELECT * FROM orders WHERE user_id = 1;
SELECT * FROM orders WHERE user_id = 2;
-- ... N queries
```

**Fix (eager loading):**
```javascript
const users = await prisma.user.findMany({
  include: { orders: true }
});
```

**SQL generated:**
```sql
SELECT * FROM users u LEFT JOIN orders o ON u.id = o.user_id;
```

**One query.** Much faster.

### Pitfall #2: Over-fetching with Includes

**ORM code:**
```javascript
const users = await prisma.user.findMany({
  include: {
    orders: true,
    addresses: true,
    profile: true
  }
});
```

**SQL generated:** Massive LEFT JOINs, Cartesian blowup.

If a user has 10 orders and 2 addresses, they appear 20 times.

**Fix:** Use multiple queries (batch fetch) or aggregations:
```javascript
const users = await prisma.user.findMany();
const userIds = users.map(u => u.id);
const orders = await prisma.order.groupBy({
  by: ['userId'],
  _count: true,
  where: { userId: { in: userIds } }
});
```

### Pitfall #3: Implicit INNER JOINs

**ORM code (Prisma with required relation):**
```javascript
const usersWithOrders = await prisma.user.findMany({
  where: { orders: { some: {} } }
});
```

**SQL generated:** INNER JOIN (or EXISTS).

**Problem:** If you meant LEFT JOIN, this excludes users without orders.

**Solution:** Be explicit about what you need:
```javascript
// All users, even those without orders
const users = await prisma.user.findMany({
  include: { orders: true }
});

// Only users with orders
const usersWithOrders = await prisma.user.findMany({
  where: { orders: { some: {} } }
});
```

## Real-World Join Scenarios

### Scenario 1: "Users Who Never Ordered"

**Goal:** Find inactive users.

**Bad approach:**
```sql
SELECT u.name
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE o.id IS NULL;
```

**Better approach:**
```sql
SELECT u.name
FROM users u
WHERE NOT EXISTS (
  SELECT 1 FROM orders WHERE user_id = u.id
);
```

Clearer intent, likely better plan.

### Scenario 2: "Top Customers by Revenue"

**Goal:** Users with highest total order value.

**Naive JOIN:**
```sql
SELECT u.name, SUM(o.total) as revenue
FROM users u
INNER JOIN orders o ON u.id = o.user_id
GROUP BY u.id, u.name
ORDER BY revenue DESC
LIMIT 10;
```

**Problem:** What if users have no orders? They're excluded (INNER JOIN).

**Better (if you want all users):**
```sql
SELECT u.name, COALESCE(SUM(o.total), 0) as revenue
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.id, u.name
ORDER BY revenue DESC
LIMIT 10;
```

Now users with 0 revenue appear at the bottom.

### Scenario 3: "Orders With Product and Category"

**Schema:**
```
orders → order_items → products → categories
```

**Query:**
```sql
SELECT 
  o.id AS order_id,
  p.name AS product,
  c.name AS category
FROM orders o
JOIN order_items oi ON o.id = oi.order_id
JOIN products p ON oi.product_id = p.id
JOIN categories c ON p.category_id = c.id;
```

**Cardinality:**
- Each order can have multiple items (1:N)
- Each item has one product (N:1)
- Each product has one category (N:1)

**Result size:** Number of order items.

**Watch for:** If you aggregate (`SUM(oi.quantity)`), group by order to avoid explosions.

### Scenario 4: "Latest Order Per User"

**Goal:** For each user, show their most recent order.

**Naive approach (slow):**
```sql
SELECT u.name, o.created_at
FROM users u
JOIN orders o ON u.id = o.user_id
WHERE o.created_at = (
  SELECT MAX(created_at) FROM orders WHERE user_id = u.id
);
```

**Better approach (with window functions):**
```sql
SELECT name, created_at
FROM (
  SELECT 
    u.name,
    o.created_at,
    ROW_NUMBER() OVER (PARTITION BY u.id ORDER BY o.created_at DESC) AS rn
  FROM users u
  JOIN orders o ON u.id = o.user_id
) sub
WHERE rn = 1;
```

We'll cover window functions in the next chapter.

## Join Optimization Checklist

✅ **Index foreign keys:** `CREATE INDEX idx_orders_user_id ON orders(user_id);`

✅ **Use appropriate join type:** Don't use INNER when you need LEFT.

✅ **Filter early:** Put filters in WHERE (for INNER) or ON (for LEFT) appropriately.

✅ **Avoid Cartesian products:** Always specify join conditions.

✅ **Watch for join explosions:** Joining multiple 1:N tables multiplies rows.

✅ **Use EXISTS for existence checks:** Faster than JOIN + DISTINCT.

✅ **Check EXPLAIN:** Verify the database is using indexes and good join strategies.

## Key Takeaways

1. **Joins are set operations, not loops.** Think in terms of combining sets.

2. **ON defines relationships; WHERE filters results.** This matters for LEFT JOIN.

3. **LEFT JOIN preserves left rows; INNER JOIN requires matches.** Choose correctly.

4. **EXISTS/NOT EXISTS are better than JOIN + DISTINCT** for existence checks.

5. **Join explosions happen when you join multiple 1:N tables.** Aggregate separately.

6. **ORMs cause N+1 queries and over-fetching.** Use eager loading wisely.

7. **Index your foreign keys.** Always.

**Next up:** Subqueries, CTEs, and window functions—the tools to write complex queries elegantly.
