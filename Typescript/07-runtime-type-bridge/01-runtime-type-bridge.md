# Level 07: Runtime ↔ Type System Bridge

## The Problem

TypeScript types are erased at runtime. This creates a fundamental gap:

```typescript
type User = { name: string; age: number };

// This function accepts unknown data (from an API, database, user input)
function processUser(data: unknown) {
  // How do you go from `unknown` to `User`?
  
  // Option 1: Assert (UNSAFE — the compiler trusts you, runtime doesn't check)
  const user = data as User;
  
  // Option 2: Type predicate (manual, error-prone)
  if (isUser(data)) { /* ... */ }
  
  // Option 3: Schema validation (Zod, Valibot, io-ts)
  const user = UserSchema.parse(data);  // Validates AND types
}
```

The fundamental question: **How do you get runtime guarantees that match your compile-time types?**

---

## Why TypeScript Cannot Infer Runtime Truth

### Types are erased

```typescript
type User = { name: string; age: number };

// At runtime, this is just:
function processUser(data) {
  // There is no User type to check against
  // typeof data === 'object' tells you nothing about shape
}
```

### `typeof` is useless for complex types

```typescript
typeof null === 'object';         // JavaScript's original sin
typeof [] === 'object';           // Can't distinguish arrays
typeof new Date() === 'object';   // Can't identify Date
typeof { name: 'Alice' } === 'object';  // Same as everything else
```

### Type assertions are lies

```typescript
interface Config {
  port: number;
  host: string;
  database: {
    url: string;
    pool: number;
  };
}

// This compiles but crashes if the JSON is wrong
const config: Config = JSON.parse(rawData);
// or
const config = JSON.parse(rawData) as Config;

// Both are equally dangerous — no runtime validation occurs
```

**At scale:** Every trust boundary in your application — API responses, environment variables, database queries, user input, file reads, message queues — is a place where `as` casts can introduce type lies.

---

## Schema-First vs Type-First Design

### Type-first (traditional approach)

1. Define a TypeScript type
2. Write validation logic manually
3. Hope they stay in sync

```typescript
// Step 1: Type
type User = {
  name: string;
  age: number;
  email: string;
};

// Step 2: Validator (must match type manually)
function isUser(data: unknown): data is User {
  return (
    typeof data === 'object' &&
    data !== null &&
    typeof (data as any).name === 'string' &&
    typeof (data as any).age === 'number' &&
    typeof (data as any).email === 'string'
  );
}

// Problem: If you add a field to User, the validator silently becomes stale
```

### Schema-first (Zod / Valibot / io-ts approach)

1. Define a schema (runtime object)
2. Derive the TypeScript type FROM the schema
3. They are always in sync because the type IS the schema

```typescript
import { z } from 'zod';

// Step 1: Schema (this is a runtime value)
const UserSchema = z.object({
  name: z.string(),
  age: z.number().int().positive(),
  email: z.string().email(),
});

// Step 2: Type derived from schema (compile-time only)
type User = z.infer<typeof UserSchema>;
// { name: string; age: number; email: string }

// Step 3: Validate at trust boundaries
const user = UserSchema.parse(untrustedData);
// Throws if invalid. If it returns, `user` is guaranteed to be User.
```

---

## How Zod's Type System Works (Mental Model)

### The core insight: Schema objects carry type information in generics

```typescript
// Simplified version of what ZodType looks like internally
class ZodType<Output, Input = Output> {
  _output!: Output;  // Phantom type — never actually set
  _input!: Input;
  
  parse(data: unknown): Output {
    // Runtime validation here
    // If valid, return data as Output
    // If invalid, throw
  }
}

class ZodString extends ZodType<string> {
  // parse() validates typeof data === 'string'
}

class ZodNumber extends ZodType<number> {
  // parse() validates typeof data === 'number'
}

class ZodObject<T extends Record<string, ZodType<any>>> extends ZodType<{
  [K in keyof T]: T[K]['_output'];
}> {
  // parse() validates each property against its schema
}
```

### How `z.infer` works

```typescript
// z.infer is just this:
type infer<T extends ZodType<any>> = T['_output'];

// When you write:
const schema = z.object({
  name: z.string(),  // ZodString extends ZodType<string>
  age: z.number(),   // ZodNumber extends ZodType<number>
});
// schema is ZodObject<{ name: ZodString; age: ZodNumber }>
// which extends ZodType<{ name: string; age: number }>

type User = z.infer<typeof schema>;
// = typeof schema['_output']
// = { name: string; age: number }
```

### The `typeof` + `z.infer` pattern

This is the key idiom:

```typescript
// typeof schema → the ZodObject instance type (with all generic info)
// z.infer<...> → extracts the _output type from the generic

// So z.infer<typeof schema> goes from:
// runtime schema object → its TypeScript type → the type it validates
```

**Why this is powerful:** The schema is a runtime value you can pass around, compose, transform. The type is derived from it automatically. You get both runtime validation AND compile-time types from a single source of truth.

---

## Schema Patterns in Practice

### API endpoint validation

```typescript
import { z } from 'zod';

// Request schema
const CreateUserRequest = z.object({
  name: z.string().min(1).max(100),
  email: z.string().email(),
  role: z.enum(['admin', 'user', 'viewer']),
});

// Response schema (what you return)
const UserResponse = z.object({
  id: z.string().uuid(),
  name: z.string(),
  email: z.string(),
  role: z.enum(['admin', 'user', 'viewer']),
  createdAt: z.string().datetime(),
});

// Types derived from schemas
type CreateUserRequest = z.infer<typeof CreateUserRequest>;
type UserResponse = z.infer<typeof UserResponse>;

// In your handler:
async function createUser(rawBody: unknown): Promise<UserResponse> {
  const body = CreateUserRequest.parse(rawBody);  // Validated
  // body.name is string, body.role is 'admin' | 'user' | 'viewer'
  
  const user = await db.users.create(body);
  return UserResponse.parse(user);  // Validate outgoing data too
}
```

### Environment variable validation

```typescript
const EnvSchema = z.object({
  NODE_ENV: z.enum(['development', 'production', 'test']),
  PORT: z.coerce.number().int().positive(),
  DATABASE_URL: z.string().url(),
  JWT_SECRET: z.string().min(32),
  REDIS_URL: z.string().url().optional(),
});

// Parse once at startup — fail fast
const env = EnvSchema.parse(process.env);

// Now env.PORT is number (not string), env.NODE_ENV is a literal union
// Access env throughout your app with full type safety
```

### Discriminated union validation

```typescript
const NotificationSchema = z.discriminatedUnion('type', [
  z.object({ type: z.literal('email'), address: z.string().email(), subject: z.string() }),
  z.object({ type: z.literal('sms'), phone: z.string() }),
  z.object({ type: z.literal('push'), token: z.string(), badge: z.number().optional() }),
]);

type Notification = z.infer<typeof NotificationSchema>;
// Same discriminated union type as if written manually
```

---

## Comparing Schema Libraries

| Feature | Zod | Valibot | io-ts | ArkType |
|---------|-----|---------|-------|---------|
| **Bundle size** | ~13KB | ~1KB (tree-shakeable) | ~7KB | ~27KB |
| **Approach** | Method chaining | Function composition | Functional / fp-ts | Type-first syntax |
| **Inference** | `z.infer<T>` | `InferOutput<T>` | `TypeOf<T>` | Direct type extraction |
| **Ecosystem** | Largest (tRPC, React Hook Form) | Growing | Mature but niche | New |
| **Error format** | Structured `ZodError` | Structured issues | fp-ts Either | Structured |

### Valibot comparison

```typescript
import * as v from 'valibot';

const UserSchema = v.object({
  name: v.pipe(v.string(), v.minLength(1)),
  age: v.pipe(v.number(), v.integer(), v.minValue(0)),
  email: v.pipe(v.string(), v.email()),
});

type User = v.InferOutput<typeof UserSchema>;

const result = v.safeParse(UserSchema, data);
if (result.success) {
  result.output;  // User
} else {
  result.issues;  // Validation errors
}
```

---

## Avoiding "Type Lies"

A type lie is any place where the static type doesn't match the runtime value.

### Common sources of type lies

| Source | Example | Fix |
|--------|---------|-----|
| **`as` assertions** | `data as User` | Schema validation |
| **`JSON.parse`** | Returns `any` | Validate the result |
| **`any` parameters** | Third-party libs returning `any` | Wrap with schema |
| **API responses** | `fetch().then(r => r.json())` | Validate response |
| **Environment variables** | `process.env.PORT` is `string \| undefined` | Schema at startup |
| **Type predicates** | `function isUser(x): x is User` — buggy guard | Use schema `.safeParse()` |
| **Database queries** | ORM returns `any` or wrong type | Validate or use Prisma/Drizzle |

### The trust boundary pattern

```typescript
// RULE: Validate at every trust boundary. Trust within boundaries.

// Trust boundary 1: HTTP input
app.post('/users', (req, res) => {
  const body = CreateUserSchema.parse(req.body);  // Validate once
  const user = userService.create(body);           // body is trusted inside
  res.json(user);
});

// Trust boundary 2: Database output
async function getUserById(id: string): Promise<User> {
  const row = await db.query('SELECT * FROM users WHERE id = $1', [id]);
  return UserSchema.parse(row);  // Don't trust DB output blindly
}

// Trust boundary 3: Third-party API
async function fetchWeather(city: string): Promise<Weather> {
  const response = await fetch(`https://api.weather.com/${city}`);
  const data = await response.json();
  return WeatherSchema.parse(data);  // External API could change
}
```

---

## Schema + TypeScript Integration Patterns

### Pattern: Shared schemas between client and server

```typescript
// packages/shared/schemas.ts
export const UserSchema = z.object({ /* ... */ });
export type User = z.infer<typeof UserSchema>;

// packages/server/handlers.ts
import { UserSchema, User } from '@myapp/shared';
function createUser(data: unknown): User {
  return UserSchema.parse(data);
}

// packages/client/api.ts
import { UserSchema, User } from '@myapp/shared';
async function fetchUser(id: string): Promise<User> {
  const data = await fetch(`/api/users/${id}`).then(r => r.json());
  return UserSchema.parse(data);
}
```

### Pattern: Schema transforms (input → output)

```typescript
const DateFromString = z.string().datetime().transform((s) => new Date(s));
// Input type: string
// Output type: Date

const UserSchema = z.object({
  name: z.string(),
  birthDate: DateFromString,
});

type UserInput = z.input<typeof UserSchema>;   // { name: string; birthDate: string }
type UserOutput = z.output<typeof UserSchema>;  // { name: string; birthDate: Date }
```

---

## Exercises

1. **Replace all `as` casts in a module:** Take a module that uses `JSON.parse(data) as Config` and refactor it to use schema validation. Ensure the types are derived from the schema.

2. **Build a type-safe environment loader:** Create a function `loadEnv()` that reads `process.env`, validates against a schema, and returns a fully typed config object. Include coercion (string `"3000"` → number `3000`).

3. **Design a schema-first API layer:** Create shared request/response schemas for a simple CRUD API (users). Derive all types. Write client and server code that both use the same schemas.

---

## Next

→ [Level 08: Library Author Patterns](../08-library-patterns/)
