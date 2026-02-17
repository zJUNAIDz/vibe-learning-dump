# Tooling: Drizzle and go-migrate

## The Philosophy of Migration Tools

Two approaches:

### 1. ORM-Managed (Type-Safe, Abstracted)
Examples: Drizzle, Prisma, TypeORM, Sequelize

**Philosophy**: Define schema in code, generate migrations automatically.

```typescript
// Define schema
export const users = pgTable('users', {
  id: serial('id').primaryKey(),
  email: text('email').notNull(),
});

// Generate migration automatically
$ drizzle-kit generate:pg
```

**Pros**:
- Type-safe
- DRY (schema is single source of truth)
- Fast development
- Auto-generated migrations

**Cons**:
- Abstraction leaks
- Generated SQL might not be optimal
- Less control
- Harder to review

### 2. SQL-First (Explicit, Controlled)
Examples: go-migrate, migrate, Atlas, Flyway, Liquibase

**Philosophy**: Write raw SQL, tool just tracks what's applied.

```sql
-- 20240215_add_email.up.sql
ALTER TABLE users ADD COLUMN email TEXT;

-- 20240215_add_email.down.sql
ALTER TABLE users DROP COLUMN email;
```

**Pros**:
- Full control
- Easy to review
- Portable
- No surprises

**Cons**:
- Manual SQL writing
- No type-safety
- Schema drift risk (code vs DB)

## Drizzle: TypeScript-First Migrations

### Setup

```bash
npm install drizzle-orm pg
npm install -D drizzle-kit
```

```typescript
// drizzle.config.ts
import { defineConfig } from 'drizzle-kit';

export default defineConfig({
  schema: './src/db/schema.ts',
  out: './drizzle',
  driver: 'pg',
  dbCredentials: {
    connectionString: process.env.DATABASE_URL!,
  },
});
```

### Defining Schema

```typescript
// src/db/schema.ts
import { pgTable, serial, text, timestamp, boolean, index } from 'drizzle-orm/pg-core';

export const users = pgTable('users', {
  id: serial('id').primaryKey(),
  email: text('email').notNull().unique(),
  username: text('username').notNull(),
  emailVerified: boolean('email_verified').default(false),
  createdAt: timestamp('created_at').defaultNow(),
}, (table) => ({
  emailIdx: index('email_idx').on(table.email),
}));

export const posts = pgTable('posts', {
  id: serial('id').primaryKey(),
  userId: integer('user_id').references(() => users.id),
  title: text('title').notNull(),
  content: text('content'),
  createdAt: timestamp('created_at').defaultNow(),
});
```

### Generating Migrations

```bash
# Generate migration from schema changes
$ npm run drizzle-kit generate:pg

# Output:
# ✅ Generated migration: drizzle/0000_add_users_table.sql
```

Generated SQL:
```sql
-- drizzle/0000_add_users_table.sql
CREATE TABLE IF NOT EXISTS "users" (
  "id" serial PRIMARY KEY NOT NULL,
  "email" text NOT NULL UNIQUE,
  "username" text NOT NULL,
  "email_verified" boolean DEFAULT false,
  "created_at" timestamp DEFAULT now()
);

CREATE INDEX IF NOT EXISTS "email_idx" ON "users" ("email");
```

### Running Migrations

```typescript
// src/db/migrate.ts
import { drizzle } from 'drizzle-orm/postgres-js';
import { migrate } from 'drizzle-orm/postgres-js/migrator';
import postgres from 'postgres';

const sql = postgres(process.env.DATABASE_URL!, { max: 1 });
const db = drizzle(sql);

async function main() {
  console.log('Running migrations...');
  await migrate(db, { migrationsFolder: './drizzle' });
  console.log('Migrations complete!');
  process.exit(0);
}

main().catch((err) => {
  console.error('Migration failed!', err);
  process.exit(1);
});
```

```bash
$ tsx src/db/migrate.ts
```

### Drizzle's Approach to Safety

**What Drizzle generates**:
```sql
-- Good: Uses IF NOT EXISTS
CREATE TABLE IF NOT EXISTS "users" (...);

-- Good: Uses IF NOT EXISTS
CREATE INDEX IF NOT EXISTS "email_idx" ON "users" ("email");
```

**What Drizzle doesn't handle well**:
- `CREATE INDEX CONCURRENTLY` (you must edit SQL manually)
- NOT NULL constraint phasing
- Complex data transformations

### When to Edit Generated Migrations

```sql
-- Generated migration
CREATE INDEX IF NOT EXISTS "users_email_idx" ON "users" ("email");
```

**Edit for production**:
```sql
-- Manually edited for safety
CREATE INDEX CONCURRENTLY IF NOT EXISTS "users_email_idx" ON "users" ("email");
```

**Best practice with Drizzle**:
1. Generate migration
2. Review generated SQL
3. Edit for production safety
4. Commit to git

### Pros and Cons

**Pros**:
- ✓ Fast development
- ✓ Type-safe queries
- ✓ Schema and types in sync
- ✓ Good for small-medium projects

**Cons**:
- ✗ Generated SQL not always optimal
- ✗ Need to manually edit for CONCURRENTLY
- ✗ Abstraction hides complexity
- ✗ Schema drift if you manually edit DB

**Best for**: TypeScript projects, small-medium scale, teams comfortable with ORMs

## go-migrate: SQL-First Approach

### Setup

```bash
# Install
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Or via Homebrew (Mac)
brew install golang-migrate

# Or download binary
curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz
```

### Creating Migrations

```bash
# Create new migration
$ migrate create -ext sql -dir migrations -seq add_users_table

# Creates:
# migrations/000001_add_users_table.up.sql
# migrations/000001_add_users_table.down.sql
```

Write migrations manually:

```sql
-- migrations/000001_add_users_table.up.sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  username VARCHAR(100) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX CONCURRENTLY idx_users_email ON users(email);
```

```sql
-- migrations/000001_add_users_table.down.sql
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

### Running Migrations

```bash
# Run all pending migrations
$ migrate -path ./migrations -database "postgres://user:pass@localhost:5432/mydb?sslmode=disable" up

# Run specific number of migrations
$ migrate -path ./migrations -database "$DATABASE_URL" up 2

# Rollback one migration
$ migrate -path ./migrations -database "$DATABASE_URL" down 1

# Check version
$ migrate -path ./migrations -database "$DATABASE_URL" version

# Force version (after manual fix)
$ migrate -path ./migrations -database "$DATABASE_URL" force 3
```

### In Application Code

```go
package main

import (
    "database/sql"
    "log"
    
    _ "github.com/lib/pq"
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
    db, err := sql.Open("postgres", "postgres://user:pass@localhost:5432/mydb?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    
    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        log.Fatal(err)
    }
    
    m, err := migrate.NewWithDatabaseInstance(
        "file://migrations",
        "postgres", 
        driver,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        log.Fatal(err)
    }
    
    log.Println("Migrations applied successfully!")
}
```

### Handling CONCURRENTLY

go-migrate doesn't support transactions for `CONCURRENTLY`:

```sql
-- migrations/000002_add_index.up.sql
-- Note: No BEGIN/COMMIT because CONCURRENTLY can't be in transaction

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_username 
  ON users(username);
```

go-migrate handles this automatically - it doesn't wrap in transaction if it detects `CONCURRENTLY`.

### Pros and Cons

**Pros**:
- ✓ Full control over SQL
- ✓ Easy to review
- ✓ No magic
- ✓ Portable (Go binary runs anywhere)
- ✓ Handles CONCURRENTLY correctly

**Cons**:
- ✗ Manual SQL writing
- ✗ No type-safety
- ✗ Schema drift possible
- ✗ Need to maintain up/down migrations manually

**Best for**: Go projects, large-scale systems, teams that want full control

## Comparing Workflows

### Drizzle Workflow

```typescript
// 1. Update schema
export const users = pgTable('users', {
  // Add new field
  phone: text('phone'),
});

// 2. Generate migration
$ npm run drizzle-kit generate:pg

// 3. Review generated SQL
# Edit drizzle/0001_add_phone.sql if needed

// 4. Run migration
$ tsx src/db/migrate.ts

// 5. Use in code (types automatically updated)
const user = await db.select().from(users);
user.phone  // TypeScript knows about this field
```

### go-migrate Workflow

```bash
# 1. Create migration
$ migrate create -ext sql -dir migrations add_phone_to_users

# 2. Write SQL manually
# migrations/000002_add_phone_to_users.up.sql

# 3. Run migration
$ migrate -database "$DATABASE_URL" -path ./migrations up

# 4. Update application code manually
type User struct {
    Phone string `db:"phone"`
}
```

## Other Tools Worth Knowing

### Atlas (SQL-First, Schema-as-Code)

```hcl
// schema.hcl
table "users" {
  schema = "public"
  column "id" {
    type = serial
  }
  column "email" {
    type = text
  }
}
```

```bash
# Generate migration
$ atlas migrate diff --env dev
```

Combines benefits of both approaches.

### Prisma (ORM-Managed, Similar to Drizzle)

```prisma
// schema.prisma
model User {
  id    Int    @id @default(autoincrement())
  email String @unique
  phone String?
}
```

```bash
$ npx prisma migrate dev --name add_phone
```

Similar philosophy to Drizzle, more opinionated.

### Flyway (Java/SQL-First)

```sql
-- V1__create_users_table.sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email VARCHAR(255)
);
```

```bash
$ flyway migrate
```

Industry standard in Java world.

## Choosing the Right Tool

### Choose Drizzle/TypeORM/Prisma if:
- ✓ Working in TypeScript/JavaScript
- ✓ Type-safety is critical
- ✓ Team comfortable with ORMs
- ✓ Small-medium scale
- ✓ Fast iteration

### Choose go-migrate/Atlas/Flyway if:
- ✓ Need full control
- ✓ Large-scale production systems
- ✓ Complex migrations
- ✓ Team has SQL expertise
- ✓ Migrations are heavily reviewed

### Choose Both (Hybrid) if:
- Use ORM for development speed
- Generate migrations
- Review and edit SQL manually
- Treat generated SQL as starting point

## CI/CD Integration

### Drizzle in CI/CD

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      
      - name: Install dependencies
        run: npm ci
      
      - name: Run migrations
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
        run: tsx src/db/migrate.ts
      
      - name: Deploy application
        run: # ... deployment steps
```

### go-migrate in CI/CD

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install migrate
        run: |
          curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz
          chmod +x migrate
          sudo mv migrate /usr/local/bin/
      
      - name: Run migrations
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
        run: migrate -path ./migrations -database "$DATABASE_URL" up
      
      - name: Deploy application
        run: # ... deployment steps
```

## Migration Versioning in Teams

### Sequential (go-migrate default)

```
migrations/
  000001_initial_schema.sql
  000002_add_users.sql
  000003_add_posts.sql
```

**Problem**: Merge conflicts when multiple people create migrations.

**Solution**: Use timestamps.

```bash
migrate create -ext sql -dir migrations -seq add_feature
# Better:
migrate create -ext sql -dir migrations add_feature
# Creates: 20240215143022_add_feature.sql
```

### Timestamp-Based (Drizzle default)

```
drizzle/
  0000_neat_vision.sql
  0001_amazing_hyperion.sql  # Auto-generated names
```

or

```
migrations/
  20240215143022_add_users.sql
  20240215150134_add_posts.sql
```

**Benefit**: No merge conflicts.

## Migration Testing

### Test Locally Before Production

```bash
# Restore production snapshot
pg_restore -d test_db prod_backup.dump

# Test migration
migrate -database "postgres://localhost/test_db" up

# Verify
psql test_db -c "\d users"

# Test rollback
migrate -database "postgres://localhost/test_db" down 1

# Verify rollback worked
psql test_db -c "\d users"
```

### Automated Testing

```typescript
// test/migrations.test.ts
import { describe, it, beforeEach } from 'vitest';
import { migrate } from 'drizzle-orm/postgres-js/migrator';

describe('Migrations', () => {
  beforeEach(async () => {
    // Reset database
    await resetTestDb();
  });

  it('should apply all migrations', async () => {
    await migrate(db, { migrationsFolder: './drizzle' });
    
    // Verify schema
    const tables = await db.execute(`
      SELECT tablename FROM pg_tables WHERE schemaname = 'public'
    `);
    
    expect(tables).toContainEqual({ tablename: 'users' });
  });

  it('should be idempotent', async () => {
    await migrate(db, { migrationsFolder: './drizzle' });
    
    // Run again - should not fail
    await migrate(db, { migrationsFolder: './drizzle' });
  });
});
```

## Summary: Tool Selection Matrix

| Aspect | Drizzle | go-migrate | Prisma | Atlas |
|--------|---------|------------|--------|-------|
| **Language** | TypeScript | Any (CLI) | TypeScript | Any (CLI) |
| **Philosophy** | ORM-managed | SQL-first | ORM-managed | Schema-as-code |
| **Type Safety** | ✓ | ✗ | ✓ | ~ |
| **Control** | Medium | High | Low | High |
| **Generated SQL** | ✓ | ✗ | ✓ | ✓ |
| **Manual SQL** | ✓ (edit) | ✓ | ✗ | ~ |
| **Concurrently** | Manual | Auto | Manual | Auto |
| **Transactions** | ✓ | ✓ | ✓ | ✓ |
| **Learning Curve** | Medium | Low | Medium | Medium |
| **Production Safety** | Review needed | Good | Review needed | Good |

**My recommendation**:
- **Small project**: Drizzle or Prisma (speed)
- **Medium project**: Drizzle + manual review
- **Large project**: go-migrate or Atlas (control)
- **Team with SQL experts**: go-migrate
- **Team new to SQL**: Drizzle

**The truth**: No tool is perfect. Review generated SQL. Test on production-sized data. Have rollback plans.

---

**Next**: [Testing and Reviewing Migrations](./09_testing_and_reviewing_migrations.md) - Catching issues before production
