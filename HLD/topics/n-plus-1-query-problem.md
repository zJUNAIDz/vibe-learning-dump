# N+1 Query Problem (and Real Fixes)

## 1. The Real Problem This Exists to Solve

When loading a collection of entities with related data, executing one query per entity (N+1 queries total) causes massive performance degradation. Instead of efficient bulk loading, the application makes hundreds or thousands of individual database roundtrips.

Real production scenario:
- Blog platform displaying list of 100 posts with author information
- **Without optimization (N+1 queries):**
  - Query 1: `SELECT * FROM posts LIMIT 100` (fetches 100 posts)
  - Query 2: `SELECT * FROM users WHERE id = 1` (author of post 1)
  - Query 3: `SELECT * FROM users WHERE id = 2` (author of post 2)
  - ...
  - Query 101: `SELECT * FROM users WHERE id = 100` (author of post 100)
  - Total: 101 queries
  - Latency: 101 × 5ms = 505ms just for database roundtrips
  - Database load: 101 separate queries
- **With dataloader (batching):**
  - Query 1: `SELECT * FROM posts LIMIT 100`
  - Query 2: `SELECT * FROM users WHERE id IN (1,2,3,...,100)`
  - Total: 2 queries
  - Latency: 2 × 5ms = 10ms
  - Database load: 2 queries
- **Performance improvement: 50× faster, 98% fewer queries**

**The fundamental problem**: ORM lazy loading and naive data fetching create hidden N+1 patterns. Each relationship traversal triggers a new query. With nested relationships (posts → authors → avatars), the problem compounds exponentially.

Without N+1 fixes:
- Loading 100 items = 100+ database queries
- Response time grows linearly with result set size
- Database overwhelmed with trivial queries
- Application doesn't scale
- User experiences slow page loads

With proper fixes:
- Loading 100 items = 2-3 queries (constant)
- Response time independent of result set size
- Database handles reasonable query volume
- Application scales to millions of users
- Fast page loads

## 2. The Naive / Incorrect Approaches (IMPORTANT)

### ❌ Incorrect Approach #1: ORM Lazy Loading (Classic N+1)

```typescript
// Incorrect: Classic ORM lazy loading pattern
// Sequelize example
const posts = await Post.findAll({ limit: 100 });

// Render posts with author names
for (const post of posts) {
  // Each iteration triggers a new query!
  const author = await post.getAuthor();  // SELECT * FROM users WHERE id = ?
  console.log(`${post.title} by ${author.name}`);
}

// Result: 1 + 100 = 101 queries
```

**Why it seems reasonable:**
- Clean, readable code
- ORM handles all SQL
- Works with any relationship
- Looks simple and elegant

**How it breaks:**
- Every `post.getAuthor()` executes a new query
- 100 posts = 100 separate author queries
- Each query has network roundtrip overhead (~5ms)
- Total latency: 100 × 5ms = 500ms minimum
- Database connection pool exhaustion under load
- Doesn't scale beyond toy examples

**Production symptoms:**
- Database monitoring shows thousands of identical queries
- Query log: `SELECT * FROM users WHERE id = 1`, `SELECT * FROM users WHERE id = 2`, ...
- Page load time increases linearly with result count
- Database CPU at 100% from query overhead
- Application slow under normal traffic

### ❌ Incorrect Approach #2: Eager Loading Everything (Over-fetching)

```typescript
// Incorrect: Eager load ALL relationships always
const posts = await Post.findAll({
  limit: 100,
  include: [
    { model: User, as: 'author', include: [
      { model: Avatar },
      { model: Profile },
      { model: Settings },
    ]},
    { model: Comment, include: [
      { model: User, as: 'commenter' },
    ]},
    { model: Tag },
    { model: Category },
    { model: Image },
  ],
});

// Generated SQL: massive JOIN with all tables
// Fetches data we don't need
```

**Why it seems reasonable:**
- Solves N+1 (single query)
- ORM handles complex JOINs
- All data available upfront
- No lazy loading

**How it breaks:**
- Fetches data you don't need (over-fetching)
- Massive JOIN query (slow)
- Cartesian product explosion with multiple has-many relationships
- Query returns 100 posts × 50 comments × 10 tags = 50,000 rows (for 100 posts)
- Network bandwidth wasted
- Memory bloat in application
- Slower than N+1 in many cases

**Production symptoms:**
- Queries taking 5+ seconds
- EXPLAIN shows 8-way JOIN
- Network traffic spike (transferring megabytes)
- OOM errors from loading too much data
- Database query timeout

### ❌ Incorrect Approach #3: Manual Loop with IN Clause (Race Conditions)

```typescript
// Incorrect: Manual batching without deduplication
const posts = await db.query('SELECT * FROM posts LIMIT 100');

// Extract author IDs
const authorIds = posts.map(p => p.author_id);

// Fetch all authors (good!)
const authors = await db.query(
  'SELECT * FROM users WHERE id IN (?)',
  [authorIds]
);

// Build map
const authorMap = new Map();
for (const author of authors) {
  authorMap.set(author.id, author);
}

// Attach authors to posts
for (const post of posts) {
  post.author = authorMap.get(post.author_id);
}
```

**Why it seems reasonable:**
- Solves N+1 (2 queries instead of 101)
- Simple to understand
- Manual control over queries

**How it breaks:**
- Code duplication (must repeat for every relationship)
- Doesn't handle nested relationships
- No caching across requests
- No automatic batching
- Race conditions if multiple requests need same data
- Doesn't compose well (hard to maintain)

**Production symptoms:**
- Copy-pasted code everywhere
- Bug fixes need to be applied in 50 places
- Nested relationships still have N+1 (comments → authors)
- Hard to reason about performance

### ❌ Incorrect Approach #4: Caching Without Batch Loading

```typescript
// Incorrect: Cache individual fetches (still N+1)
const userCache = new Map();

async function getUserWithCache(userId: number): Promise<User> {
  if (userCache.has(userId)) {
    return userCache.get(userId);
  }
  
  // Still makes N individual queries on cache miss
  const user = await db.query('SELECT * FROM users WHERE id = ?', [userId]);
  userCache.set(userId, user);
  return user;
}

// Usage
const posts = await Post.findAll({ limit: 100 });
for (const post of posts) {
  post.author = await getUserWithCache(post.author_id);  // Still N queries on first load
}
```

**Why it seems reasonable:**
- Caching reduces repeated queries
- Subsequent requests fast
- Simple caching layer

**How it breaks:**
- First request still N+1 (cache miss)
- Each userId queried individually
- No batching of cache misses
- Doesn't solve fundamental problem
- Only helps with duplicate IDs in same result set

**Production symptoms:**
- First page load after cache clear: slow (500ms)
- Subsequent loads: fast (50ms)
- High cache hit rate but still periodic slowdowns
- Doesn't scale with fresh data

## 3. Correct Mental Model (How It Actually Works)

The N+1 problem occurs when code loops over results and makes a query per iteration. The solution is batching: collect all needed IDs, fetch in one query, then distribute results.

### The Batching Pattern

```
N+1 pattern (bad):
  for each post:
    query user by post.author_id
  Total: N queries

Batching pattern (good):
  collect all author_ids from posts
  query users WHERE id IN (author_ids)
  map users to posts
  Total: 1 query
```

### The DataLoader Pattern

```
DataLoader = Batching + Caching + Request-scoped

Key features:
  1. Batching: Collects keys from multiple calls
  2. Waits one tick: Aggregates requests made in same event loop tick
  3. Single query: Fetches all keys in one batch
  4. Caching: Caches results per request
  5. Returns promises: Each caller gets their specific result
```

### The Execution Flow

```
// Code calls loader multiple times
dataLoader.load(1)
dataLoader.load(2)
dataLoader.load(1)  // duplicate
dataLoader.load(3)

// DataLoader batches:
  1. Collects keys: [1, 2, 1, 3]
  2. Deduplicates: [1, 2, 3]
  3. Calls batch function: batchLoadUsers([1, 2, 3])
  4. Executes: SELECT * FROM users WHERE id IN (1, 2, 3)
  5. Returns: [user1, user2, user3]
  6. Maps results to callers:
     - First load(1) gets user1
     - load(2) gets user2
     - Second load(1) gets user1 (from cache)
     - load(3) gets user3
```

### The Relationship Loading Strategy

```
Query 1: Load primary entities
  SELECT * FROM posts LIMIT 100

Query 2: Load first-level relationships
  SELECT * FROM users WHERE id IN (author_ids from posts)

Query 3: Load second-level relationships
  SELECT * FROM avatars WHERE user_id IN (user_ids from users)

Total: 3 queries regardless of N
```

## 4. Correct Design & Algorithm

### Strategy 1: DataLoader (Most Common)

```typescript
import DataLoader from 'dataloader';

const userLoader = new DataLoader(async (userIds: number[]) => {
  const users = await db.query(
    'SELECT * FROM users WHERE id IN (?)',
    [userIds]
  );
  
  // Return users in same order as userIds
  const userMap = new Map(users.map(u => [u.id, u]));
  return userIds.map(id => userMap.get(id) || null);
});

// Usage
const posts = await db.query('SELECT * FROM posts LIMIT 100');
for (const post of posts) {
  post.author = await userLoader.load(post.author_id);
}
// Result: 2 queries total
```

### Strategy 2: JOIN with Aggregation

```sql
-- Single query with JSON aggregation
SELECT 
  posts.*,
  json_agg(comments.*) as comments
FROM posts
LEFT JOIN comments ON comments.post_id = posts.id
WHERE posts.id IN (1, 2, 3)
GROUP BY posts.id;
```

### Strategy 3: Manual Batching

```typescript
async function loadPostsWithAuthors(postIds: number[]) {
  const [posts, authors] = await Promise.all([
    db.query('SELECT * FROM posts WHERE id IN (?)', [postIds]),
    db.query(`
      SELECT users.* FROM users
      JOIN posts ON posts.author_id = users.id
      WHERE posts.id IN (?)
    `, [postIds])
  ]);
  
  const authorMap = new Map(authors.map(a => [a.id, a]));
  return posts.map(p => ({
    ...p,
    author: authorMap.get(p.author_id)
  }));
}
```

## 5. Full Production-Grade Implementation

```typescript
import DataLoader from 'dataloader';

/**
 * DataLoader for users
 */
function createUserLoader(db: Database): DataLoader<number, User> {
  return new DataLoader(
    async (userIds: readonly number[]) => {
      const users = await db.query(
        'SELECT id, name, email, avatar_url FROM users WHERE id = ANY($1)',
        [Array.from(userIds)]
      );
      
      // Create map for O(1) lookup
      const userMap = new Map<number, User>(
        users.rows.map(user => [user.id, user])
      );
      
      // Return in same order as input (DataLoader requirement)
      return userIds.map(id => userMap.get(id) || null);
    },
    {
      // Cache results per request
      cache: true,
      // Batch size limit
      maxBatchSize: 100,
    }
  );
}

/**
 * DataLoader for posts by author
 */
function createPostsByAuthorLoader(db: Database): DataLoader<number, Post[]> {
  return new DataLoader(
    async (authorIds: readonly number[]) => {
      const posts = await db.query(
        'SELECT * FROM posts WHERE author_id = ANY($1) ORDER BY created_at DESC',
        [Array.from(authorIds)]
      );
      
      // Group posts by author_id
      const postsByAuthor = new Map<number, Post[]>();
      for (const post of posts.rows) {
        if (!postsByAuthor.has(post.author_id)) {
          postsByAuthor.set(post.author_id, []);
        }
        postsByAuthor.get(post.author_id)!.push(post);
      }
      
      // Return array of posts for each author (in order)
      return authorIds.map(id => postsByAuthor.get(id) || []);
    }
  );
}

/**
 * DataLoader for comments by post
 */
function createCommentsByPostLoader(db: Database): DataLoader<number, Comment[]> {
  return new DataLoader(
    async (postIds: readonly number[]) => {
      const comments = await db.query(
        'SELECT * FROM comments WHERE post_id = ANY($1) ORDER BY created_at ASC',
        [Array.from(postIds)]
      );
      
      const commentsByPost = new Map<number, Comment[]>();
      for (const comment of comments.rows) {
        if (!commentsByPost.has(comment.post_id)) {
          commentsByPost.set(comment.post_id, []);
        }
        commentsByPost.get(comment.post_id)!.push(comment);
      }
      
      return postIds.map(id => commentsByPost.get(id) || []);
    }
  );
}

/**
 * Request context with all loaders
 */
interface RequestContext {
  loaders: {
    user: DataLoader<number, User>;
    postsByAuthor: DataLoader<number, Post[]>;
    commentsByPost: DataLoader<number, Comment[]>;
  };
}

/**
 * Create fresh loaders for each request
 */
function createLoaders(db: Database): RequestContext['loaders'] {
  return {
    user: createUserLoader(db),
    postsByAuthor: createPostsByAuthorLoader(db),
    commentsByPost: createCommentsByPostLoader(db),
  };
}

// Express middleware to attach loaders to request
app.use((req, res, next) => {
  req.context = {
    loaders: createLoaders(db),
  };
  next();
});

/**
 * Resolve post with nested relationships
 */
async function resolvePost(post: Post, context: RequestContext): Promise<PostWithRelations> {
  const [author, comments] = await Promise.all([
    context.loaders.user.load(post.author_id),
    context.loaders.commentsByPost.load(post.id),
  ]);
  
  // Resolve comment authors
  const commentsWithAuthors = await Promise.all(
    comments.map(async (comment) => ({
      ...comment,
      author: await context.loaders.user.load(comment.author_id),
    }))
  );
  
  return {
    ...post,
    author,
    comments: commentsWithAuthors,
  };
}

// API endpoint
app.get('/api/posts', async (req, res) => {
  // Fetch posts
  const posts = await db.query('SELECT * FROM posts ORDER BY created_at DESC LIMIT 20');
  
  // Resolve all relationships in parallel
  const resolvedPosts = await Promise.all(
    posts.rows.map(post => resolvePost(post, req.context))
  );
  
  res.json(resolvedPosts);
});

// GraphQL resolver example
const resolvers = {
  Query: {
    posts: async (_, __, context: RequestContext) => {
      const result = await db.query('SELECT * FROM posts LIMIT 20');
      return result.rows;
    },
  },
  Post: {
    author: async (post: Post, _, context: RequestContext) => {
      return context.loaders.user.load(post.author_id);
    },
    comments: async (post: Post, _, context: RequestContext) => {
      return context.loaders.commentsByPost.load(post.id);
    },
  },
  Comment: {
    author: async (comment: Comment, _, context: RequestContext) => {
      return context.loaders.user.load(comment.author_id);
    },
  },
};

/**
 * Alternative: Prisma with include (handles N+1 automatically)
 */
async function getPostsWithPrisma() {
  const posts = await prisma.post.findMany({
    take: 20,
    include: {
      author: true,  // Single JOIN
      comments: {
        include: {
          author: true,  // Additional JOIN
        },
      },
    },
  });
  return posts;
}

/**
 * Manual batching for custom queries
 */
class ManualBatcher {
  private pending: Array<{
    id: number;
    resolve: (value: any) => void;
    reject: (error: Error) => void;
  }> = [];
  
  private timer: NodeJS.Timeout | null = null;

  async load(id: number): Promise<User> {
    return new Promise((resolve, reject) => {
      this.pending.push({ id, resolve, reject });
      
      if (!this.timer) {
        // Batch on next tick
        this.timer = setTimeout(() => this.flush(), 0);
      }
    });
  }

  private async flush(): Promise<void> {
    const batch = this.pending.splice(0);
    this.timer = null;

    if (batch.length === 0) return;

    try {
      const ids = batch.map(item => item.id);
      const users = await db.query(
        'SELECT * FROM users WHERE id IN (?)',
        [ids]
      );

      const userMap = new Map(users.map(u => [u.id, u]));

      for (const item of batch) {
        const user = userMap.get(item.id);
        if (user) {
          item.resolve(user);
        } else {
          item.reject(new Error(`User ${item.id} not found`));
        }
      }
    } catch (error) {
      for (const item of batch) {
        item.reject(error);
      }
    }
  }
}

/**
 * Monitoring N+1 queries in development
 */
let queryCount = 0;

function trackQuery(sql: string) {
  queryCount++;
  if (queryCount > 10) {
    console.warn(`⚠️  N+1 detected: ${queryCount} queries in single request`);
    console.warn(`   Last query: ${sql}`);
  }
}

app.use((req, res, next) => {
  queryCount = 0;
  res.on('finish', () => {
    if (queryCount > 10) {
      console.error(`N+1 ALERT: ${req.path} executed ${queryCount} queries`);
    }
  });
  next();
});

interface User {
  id: number;
  name: string;
  email: string;
  avatar_url?: string;
}

interface Post {
  id: number;
  title: string;
  content: string;
  author_id: number;
  created_at: Date;
}

interface Comment {
  id: number;
  content: string;
  post_id: number;
  author_id: number;
  created_at: Date;
}

interface PostWithRelations extends Post {
  author: User;
  comments: Array<Comment & { author: User }>;
}
```

## 6. Correct Usage Patterns (Where This Shines)

### Pattern 1: GraphQL Resolvers

```typescript
const resolvers = {
  Post: {
    author: (post, _, { loaders }) => loaders.user.load(post.author_id),
    comments: (post, _, { loaders }) => loaders.commentsByPost.load(post.id),
  },
  Comment: {
    author: (comment, _, { loaders }) => loaders.user.load(comment.author_id),
  },
};
// DataLoader automatically batches all user.load() calls
```

### Pattern 2: REST API with Nested Resources

```typescript
app.get('/api/posts/:id', async (req, res) => {
  const post = await db.query('SELECT * FROM posts WHERE id = $1', [req.params.id]);
  post.author = await req.context.loaders.user.load(post.author_id);
  post.comments = await req.context.loaders.commentsByPost.load(post.id);
  res.json(post);
});
```

### Pattern 3: Batch Loading for Reports

```typescript
async function generateReport(userIds: number[]) {
  const users = await userLoader.loadMany(userIds);
  const posts = await Promise.all(
    userIds.map(id => postsByAuthorLoader.load(id))
  );
  // Efficiently loads all data with minimal queries
}
```

## 7. Failure Modes & Edge Cases

### DataLoader Not Created Per-Request

**Problem:** Sharing DataLoader across requests causes cache leaks.

**Fix:** Create new loaders for each request.

### Ordering Mismatch

**Problem:** Batch function returns results in wrong order.

**Fix:** Ensure batch function returns array matching input order exactly.

### Null Handling

**Problem:** Missing entities cause undefined instead of null.

**Fix:** Return null for missing items, not undefined.

## 8. Performance Characteristics & Tradeoffs

### Query Count

- N+1: 1 + N queries
- DataLoader: 1 + D queries (D = depth of nesting)
- For 100 posts with authors: 101 → 2 queries (50× reduction)

### Memory

- DataLoader caches results per request (~KB per request)
- Negligible compared to N+1 savings

### Latency

- N+1: N × roundtrip_time
- DataLoader: D × roundtrip_time (D usually 2-4)

## 9. Foot-Guns & Common Mistakes (DO NOT SKIP)

### Mistake 1: Forgetting to Return Array in Same Order

**Fix:** Batch function MUST return results in same order as input keys.

### Mistake 2: Reusing DataLoader Across Requests

**Fix:** Create new loaders per request (avoid stale cache).

### Mistake 3: Not Handling Null/Missing Entities

**Fix:** Return null for missing entities, not undefined.

### Mistake 4: Batch Function That Doesn't Scale

**Fix:** Limit batch size with maxBatchSize option.

### Mistake 5: Not Monitoring Query Count

**Fix:** Track queries per request, alert on >10 queries.

## 10. When NOT to Use This (Anti-Patterns)

### Single Entity Fetch

If loading one user, don't use DataLoader (overhead not worth it).

### Write Operations

DataLoader is for reads. Don't batch INSERT/UPDATE/DELETE.

### Real-Time Consistency Required

DataLoader caches may serve stale data within request.

## 11. Related Concepts (With Contrast)

### Eager Loading

**Difference:** Eager loading uses JOINs. DataLoader batches separate queries.

### Caching

**Difference:** Cache persists across requests. DataLoader is per-request only.

### Query Optimization

**Related:** N+1 fixes are form of query optimization.

## 12. Production Readiness Checklist

### DataLoader Setup

- [ ] Create loaders per request (not global)
- [ ] Batch functions return correct order
- [ ] Handle null/missing entities
- [ ] Set maxBatchSize (100-1000)

### Monitoring

- [ ] Track queries per request
- [ ] Alert on >10 queries
- [ ] Dashboard showing query count distribution
- [ ] APM showing database time

### Code Review

- [ ] Grep for `await` in loops
- [ ] Check ORM lazy loading disabled
- [ ] Verify relationships use loaders
- [ ] Test with N=1000 to catch N+1

### Testing

- [ ] Load test with large N
- [ ] Verify query count stays constant
- [ ] Check database connection pool
- [ ] Profile memory usage
