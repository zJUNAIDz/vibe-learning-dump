# Level 08: Library Author Patterns

## The Problem

You can write TypeScript application code well. But when you try to write a reusable library, SDK, or framework, your types become:

- Hard to use (consumers need to manually annotate everything)
- Hard to read (error messages are walls of text)
- Hard to maintain (changing internals breaks public types)
- Accidentally exposing implementation details

This level teaches the patterns used by authors of Zod, tRPC, Prisma, React Query, and similar libraries.

---

## Designing Ergonomic Public Types

### Rule 1: The consumer should never write a type annotation

```typescript
// ❌ Bad API — consumer must annotate
function createStore<TState>(config: StoreConfig<TState>): Store<TState> { /* ... */ }

// Consumer:
const store = createStore<{ count: number; name: string }>({
  initialState: { count: 0, name: '' },
  //             ^^ the type info is RIGHT HERE but they had to write it twice
});

// ✅ Good API — infer from arguments
function createStore<TState>(initialState: TState) {
  return {
    getState: (): TState => initialState,
    setState: (updater: (prev: TState) => TState) => { /* ... */ },
  };
}

// Consumer:
const store = createStore({ count: 0, name: '' });
// TState inferred as { count: number; name: string }
// Zero annotations needed
```

### Rule 2: Error messages should point to the consumer's mistake

```typescript
// ❌ Bad — error appears deep in library types
type Config<T> = {
  fields: { [K in keyof T]: FieldDef<T[K]> };
  validate: (data: { [K in keyof T]: T[K] extends string ? string : T[K] extends number ? number : never }) => boolean;
};
// Consumer sees: "Type 'string' is not assignable to type 'T[K] extends string ? string : ...'"

// ✅ Good — error at the site of the mistake
function defineFields<const T extends Record<string, 'string' | 'number'>>(schema: T) {
  return schema;
}
// Consumer sees: "Type '"boolean"' is not assignable to type '"string" | "number"'"
// Clear and actionable
```

### Rule 3: Hover types should be readable

```typescript
// ❌ Hover shows: Prettify<Omit<ExtractFields<InternalSchema<T, Opts>>, '__meta'> & Required<Pick<T, RequiredKeys<T>>>>
// Unreadable

// ✅ Use a "Prettify" helper to flatten intersections and mapped types
type Prettify<T> = { [K in keyof T]: T[K] } & {};

type Ugly = { a: string } & { b: number } & Omit<{ c: boolean; d: Date }, 'd'>;
type Clean = Prettify<Ugly>;
// Hover shows: { a: string; b: number; c: boolean }
```

---

## Hiding Internal Complexity

### The public/internal type split

```typescript
// INTERNAL: Complex, optimized, ugly
type _InternalRouterState<
  TRoutes extends Record<string, _InternalRouteDefinition<any, any, any>>,
  TMiddleware extends _InternalMiddlewareChain<any>,
  TContext extends _InternalContextShape<TMiddleware>
> = {
  __routes: TRoutes;
  __middleware: TMiddleware;
  __context: TContext;
};

// PUBLIC: Clean, simple, what consumers see
type Router<TRoutes extends Record<string, RouteHandler>> = {
  handle: (req: Request) => Promise<Response>;
  routes: TRoutes;
};

// The implementation bridges internal to public
function createRouter(): Router<{}> {
  const internal: _InternalRouterState<{}, [], {}> = { /* ... */ };
  
  return {
    handle: async (req) => { /* uses internal state */ },
    routes: {} as any,
  } as Router<{}>;
}
```

### Convention: prefix internal types with `_`

```typescript
// Consumers should never import or reference these
export type _Brand<T, B> = T & { __brand: B };
export type _InferInput<T> = T extends _Schema<infer I, any> ? I : never;

// These are the public API
export type Brand<T, B extends string> = _Brand<T, B>;
export type InferInput<T extends Schema<any>> = _InferInput<T>;
```

### Using `@internal` JSDoc

```typescript
/**
 * @internal
 * Do not use directly — use `createRouter()` instead.
 */
export class _RouterImpl<T> {
  // ...
}
```

---

## Overloads vs Generics

### When to use overloads

Use overloads when the relationship between input and output types is **not expressible as a generic**:

```typescript
// The behavior changes based on input type — not a generic mapping
function createElement(tag: 'div'): HTMLDivElement;
function createElement(tag: 'span'): HTMLSpanElement;
function createElement(tag: 'input'): HTMLInputElement;
function createElement(tag: string): HTMLElement;
function createElement(tag: string): HTMLElement {
  return document.createElement(tag);
}

const div = createElement('div');    // HTMLDivElement
const span = createElement('span');  // HTMLSpanElement
const custom = createElement('x');   // HTMLElement
```

### When to use generics

Use generics when the output type is **structurally derived** from the input:

```typescript
function pick<T, K extends keyof T>(obj: T, keys: K[]): Pick<T, K> {
  const result = {} as Pick<T, K>;
  for (const key of keys) {
    result[key] = obj[key];
  }
  return result;
}

const user = { name: 'Alice', age: 30, email: 'a@b.com' };
const picked = pick(user, ['name', 'email']);
// type: Pick<{ name: string; age: number; email: string }, 'name' | 'email'>
// = { name: string; email: string }
```

### Overloads for consumer ergonomics, generic for implementation

```typescript
// Public API: clean overloads
function fetch(url: string): Promise<Response>;
function fetch(url: string, init: RequestInit): Promise<Response>;
function fetch<T>(url: string, init: RequestInit & { parseAs: 'json' }): Promise<T>;

// Implementation: single generic
function fetch<T = Response>(
  url: string,
  init?: RequestInit & { parseAs?: 'json' }
): Promise<T | Response> {
  // ...
}
```

### Overload ordering matters

```typescript
// ❌ Wrong order — generic overload catches everything
function parse(input: string): string;   // Never reached
function parse<T>(input: T): T;          // Catches all calls
function parse(input: any) { return input; }

parse('hello');  // type: 'hello' (matched generic, not string overload)

// ✅ Correct order — most specific first
function parse<T>(input: T): T;
function parse(input: string): string;
function parse(input: any) { return input; }
```

Wait — that's also wrong. TypeScript checks overloads **top to bottom** and uses the **first match**.

```typescript
// ✅ Actually correct
function parse(input: string): string;     // Try this first
function parse<T>(input: T): T;            // Fallback
function parse(input: any) { return input; }

parse('hello');  // string (matched first overload)
parse(42);       // 42 (matched generic fallback)
```

---

## When to Use `any` Intentionally

### Inside library implementation (not in the public API)

```typescript
// The public type is perfectly safe
function merge<A extends object, B extends object>(a: A, b: B): Prettify<A & B> {
  // Inside the implementation, the compiler can't verify the spread
  // matches A & B because it doesn't understand structural spread.
  // This is a SAFE use of `as any` — the public type is correct.
  return { ...a, ...b } as any;
}
```

### Type-level "escape hatch" in conditional types

```typescript
function process<T extends string | number>(val: T): T extends string ? string : number {
  // TypeScript can't narrow T inside the body (Level 03)
  // `as any` is the standard pattern
  if (typeof val === 'string') return val.toUpperCase() as any;
  return (val as number * 2) as any;
}
```

### Bridge between runtime and type system

```typescript
// Builder pattern where types accumulate
class Builder<T extends Record<string, any> = {}> {
  private data: any = {};

  add<K extends string, V>(key: K, value: V): Builder<T & Record<K, V>> {
    this.data[key] = value;
    return this as any;  // Safe — the public generic tracks the actual shape
  }

  build(): T {
    return this.data;
  }
}

const result = new Builder()
  .add('name', 'Alice')
  .add('age', 30)
  .build();
// type: { name: string } & { age: number }
```

### Rules for safe `any` usage

1. **Never in public API signatures** — consumers should never see `any`
2. **Only when the public type is correct** — `any` inside, safe type outside
3. **Comment why** — `as any // safe: public type tracks actual shape`
4. **Consider `as unknown as TargetType` instead** — more explicit about the cast

---

## Advanced Patterns from Real Libraries

### Pattern: Plugin/middleware chaining (tRPC-style)

```typescript
type Middleware<TContext, TNewContext> = {
  (ctx: TContext): TNewContext;
};

class ProcedureBuilder<TContext> {
  use<TNewContext>(
    middleware: Middleware<TContext, TNewContext>
  ): ProcedureBuilder<TContext & TNewContext> {
    // Store middleware internally
    return this as any;
  }

  handler<TOutput>(
    fn: (ctx: TContext) => TOutput
  ): { (ctx: TContext): TOutput } {
    return fn;
  }
}

const procedure = new ProcedureBuilder<{ req: Request }>()
  .use((ctx) => ({ ...ctx, user: { id: '123' } }))
  .use((ctx) => ({ ...ctx, db: createDB() }))
  .handler((ctx) => {
    // ctx: { req: Request } & { user: { id: string } } & { db: DB }
    ctx.user.id;  // ✅
    ctx.db.query;  // ✅
  });
```

### Pattern: Fluent builder with discriminated result

```typescript
type QueryState = 'unfiltered' | 'filtered' | 'sorted';

class Query<T, State extends QueryState = 'unfiltered'> {
  where(predicate: (item: T) => boolean): Query<T, 'filtered'> {
    return this as any;
  }

  // orderBy is only available after filtering
  orderBy(
    this: Query<T, 'filtered'>,
    key: keyof T
  ): Query<T, 'sorted'> {
    return this as any;
  }

  // execute is available in any state
  execute(): T[] {
    return [];
  }
}

const query = new Query<{ name: string; age: number }>();
query.orderBy('name');           // ❌ Can't sort unfiltered query
query.where(() => true).orderBy('name');  // ✅
```

### Pattern: `declare function` for type-only exports

When you're writing `.d.ts` files or type utilities:

```typescript
// No implementation needed — this is a type-level function
declare function defineConfig<const T extends Record<string, unknown>>(config: T): T;

// Consumers get perfect inference
const config = defineConfig({
  port: 3000,
  features: ['auth', 'logging'] as const,
});
// type: { readonly port: 3000; readonly features: readonly ['auth', 'logging'] }
```

---

## Publishing Type-Safe Libraries

### `package.json` type export fields

```json
{
  "name": "my-library",
  "types": "./dist/index.d.ts",
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": "./dist/index.mjs",
      "require": "./dist/index.cjs"
    }
  }
}
```

**The `types` field must come FIRST** in the exports condition — TypeScript uses the first matching condition.

### What to export

```typescript
// index.ts — public API

// Export types that consumers need
export type { User, Config, Result } from './types';

// Export runtime values
export { createRouter, defineHandler } from './router';

// Do NOT export internal types
// export type { _InternalState } from './internal';  ← no
```

### Testing your public types

```typescript
// __tests__/types.test.ts
import { expectType, expectError } from 'tsd';
import { createRouter } from '../src';

// Verify correct types
const router = createRouter({ home: '/' });
expectType<string>(router.routes.home);

// Verify errors for wrong usage
expectError(createRouter({ home: 123 }));
```

---

## Exercises

1. **Build a type-safe event bus:**

Design a publish/subscribe API where:
- Events are defined as a schema: `{ userCreated: { id: string }; orderPlaced: { total: number } }`
- `subscribe('userCreated', (data) => ...)` infers `data` type
- `publish('userCreated', { id: '123' })` enforces correct payload
- No type annotations needed by the consumer

2. **Build a type-safe SQL query builder:**

```typescript
const query = createQuery<User>()
  .select('name', 'email')
  .where('age', '>', 18)
  .orderBy('name', 'asc')
  .limit(10);

// query.execute() returns Pick<User, 'name' | 'email'>[]
// .where() only accepts valid keys and type-appropriate comparisons
```

3. **Audit a real library's types:** Pick a library you use (e.g., React Query). Open its `.d.ts` files and identify: overloads, conditional types, generic inference patterns, and internal vs public types.

---

## Next

→ [Level 09: Compiler & Tooling Knowledge](../09-compiler-tooling/)
