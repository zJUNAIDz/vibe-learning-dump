# Level 03: Advanced Generics — The Real Stuff

## The Problem

You use generics daily. But can you explain why this works:

```typescript
// tRPC-style router
const router = createRouter({
  getUser: procedure.input(z.object({ id: z.string() })).query(({ input }) => {
    return db.user.findUnique({ where: { id: input.id } });
    //                                      ^^ TypeScript knows input.id is string
  }),
});
```

The entire chain — from `z.object({ id: z.string() })` down to `input.id` inside the callback — is type-safe with **zero manual annotations**. This is generics doing the heavy lifting.

This level teaches you to build APIs like this.

---

## Generic Constraints: Beyond `extends`

### The basic constraint

```typescript
function getLength<T extends { length: number }>(val: T): number {
  return val.length;
}

getLength('hello');     // ✅ string has .length
getLength([1, 2, 3]);   // ✅ array has .length
getLength(42);           // ❌ number has no .length
```

### Constraints as inference anchors

The constraint doesn't just restrict — it **guides inference**:

```typescript
function create<T extends Record<string, unknown>>(obj: T): T {
  return obj;
}

const result = create({ name: 'Alice', age: 30 });
// T = { name: string; age: number }
// NOT Record<string, unknown> — the constraint is a minimum, not the inferred type
```

**Key insight:** The constraint sets a lower bound. TypeScript still infers the most specific type that satisfies the constraint.

### Constraint narrowing

```typescript
function process<T extends string | number>(val: T): T extends string ? string : number {
  if (typeof val === 'string') {
    return val.toUpperCase() as any;  // See why this needs 'as any' below
  }
  return (val * 2) as any;
}
```

**Failure mode:** You cannot narrow a generic parameter with runtime checks. Inside the function body, `T` is still `string | number`. The `extends` check only works at the type level. The function body must use type assertions or overloads. This is one of TypeScript's sharpest edges.

---

## Inference from Function Arguments

### Single inference site

```typescript
function wrap<T>(value: T): { value: T } {
  return { value };
}

wrap(42);        // T = 42 (literal inferred from const-like position? No — T = number)
wrap('hello');   // T = string

// Wait, why number and not 42?
// Because the parameter is typed as T, not as a const position.
// The value 42 in a mutable position widens to number.
```

### Preserving literals with `const` type parameters

```typescript
function wrap<const T>(value: T): { value: T } {
  return { value };
}

wrap(42);      // T = 42
wrap('hello'); // T = 'hello'
```

### Multiple inference sites for the same parameter

```typescript
function assertEqual<T>(a: T, b: T): boolean {
  return a === b;
}

assertEqual(1, 2);         // T = number ✅
assertEqual(1, 'hello');   // T = string | number — no error!
```

When TypeScript has multiple candidates for `T`, it **unions** them. If you want an error, you need a different pattern:

```typescript
// Force exact match using a "helper" parameter
function assertEqual<T>(a: T, b: NoInfer<T>): boolean {
  return a === b;
}

assertEqual(1, 'hello');  // ❌ Type 'string' is not assignable to type 'number'
```

`NoInfer<T>` (TypeScript 5.4+) prevents the second parameter from contributing to inference of `T`. The first argument alone determines `T`, and the second must match.

### Inference from callback arguments

This is how libraries create "magical" type experiences:

```typescript
function defineHandler<TInput, TOutput>(config: {
  input: TInput;
  handler: (input: TInput) => TOutput;
}): TOutput {
  return config.handler(config.input);
}

const result = defineHandler({
  input: { name: 'Alice', age: 30 },
  handler: (input) => {
    // input is { name: string; age: number } — inferred from the `input` field!
    return input.name.toUpperCase();
  },
});
// result is string
```

**How this works:** TypeScript infers `TInput` from `config.input`, then uses that inference to type the `input` parameter of `config.handler`. Finally, it infers `TOutput` from the handler's return type.

This is the pattern that powers tRPC, React Query's `queryFn`, and similar APIs.

---

## Partial Inference Failures

### The problem

```typescript
function createStore<TState, TActions>(config: {
  state: TState;
  actions: TActions;
}): { state: TState; actions: TActions } {
  return config;
}

// You want TState inferred from `state`, but you want to constrain TActions
// TypeScript can't partially specify generics — it's all or nothing
createStore<{ count: number }>({  // ❌ Expected 2 type arguments, got 1
  state: { count: 0 },
  actions: { increment: () => {} },
});
```

### Solution 1: Curried function (the builder pattern)

```typescript
function createStore<TState>(state: TState) {
  return {
    actions<TActions extends Record<string, (state: TState) => TState>>(
      actions: TActions
    ) {
      return { state, actions };
    },
  };
}

const store = createStore({ count: 0 })
  .actions({
    increment: (state) => ({ count: state.count + 1 }),
    //          ^^ state is { count: number } — inferred from first call
  });
```

### Solution 2: Defaults with inference

```typescript
function createStore<
  TState,
  TActions extends Record<string, (state: TState) => TState> = {}
>(config: {
  state: TState;
  actions?: TActions;
}) {
  return config;
}

// TState inferred, TActions defaults to {}
const store = createStore({ state: { count: 0 } });
```

### Solution 3: The `satisfies` trick for partial constraints

```typescript
function defineRoutes<T extends Record<string, { path: string; method: string }>>(
  routes: T
): T {
  return routes;
}

// T is fully inferred but constrained
const routes = defineRoutes({
  getUser: { path: '/users/:id', method: 'GET' },
  createUser: { path: '/users', method: 'POST' },
});
// routes.getUser.method is string, not 'GET'

// Better: use const type parameter
function defineRoutes2<const T extends Record<string, { path: string; method: string }>>(
  routes: T
): T {
  return routes;
}
// routes.getUser.method is 'GET' ← literal preserved
```

---

## Designing "Magical" Generic APIs

### Pattern: Infer from configuration, propagate through callbacks

```typescript
type FieldDef = {
  type: 'string' | 'number' | 'boolean';
  required?: boolean;
};

type InferFieldType<F extends FieldDef> =
  F['type'] extends 'string' ? string :
  F['type'] extends 'number' ? number :
  F['type'] extends 'boolean' ? boolean :
  never;

type InferSchema<S extends Record<string, FieldDef>> = {
  [K in keyof S]: InferFieldType<S[K]>;
};

function defineModel<const S extends Record<string, FieldDef>>(schema: S) {
  return {
    create(data: InferSchema<S>): InferSchema<S> {
      return data;
    },
  };
}

const UserModel = defineModel({
  name: { type: 'string', required: true },
  age: { type: 'number' },
  active: { type: 'boolean' },
});

UserModel.create({
  name: 'Alice',   // must be string
  age: 30,         // must be number
  active: true,    // must be boolean
});
```

### Pattern: Builder chain with accumulating types

```typescript
class QueryBuilder<TSelected extends string = never> {
  private columns: string[] = [];

  select<K extends string>(column: K): QueryBuilder<TSelected | K> {
    this.columns.push(column);
    return this as any;
  }

  execute(): Record<TSelected, unknown> {
    return {} as any;  // Implementation detail
  }
}

const result = new QueryBuilder()
  .select('name')
  .select('age')
  .select('email')
  .execute();

// type: Record<'name' | 'age' | 'email', unknown>
// Each .select() call adds to the TSelected union
```

---

## Common Generic Anti-Patterns

### 1. Unnecessary generics

```typescript
// ❌ Generic adds no value
function bad<T extends string>(val: T): T {
  return val;
}

// ✅ Just use the type directly
function good(val: string): string {
  return val;
}
```

Only use a generic when you need to **relate** input types to output types or **preserve** specific type information.

### 2. Constraining then ignoring the constraint

```typescript
// ❌ The generic T is constrained but used as any object
function bad<T extends object>(obj: T): string {
  return JSON.stringify(obj);
}

// ✅ No generic needed
function good(obj: object): string {
  return JSON.stringify(obj);
}
```

### 3. Using generics where conditional types are needed

```typescript
// ❌ Trying to use generics for type-level conditional logic
function process<T extends string | number>(val: T): T {
  // Can't narrow T inside the body
}

// ✅ Use overloads for different behavior per type
function process(val: string): string;
function process(val: number): number;
function process(val: string | number): string | number {
  return typeof val === 'string' ? val.toUpperCase() : val * 2;
}
```

---

## Exercises

1. **Build a type-safe event emitter:**

```typescript
// Design a generic EventEmitter where:
// - Events are defined as a type map: { click: MouseEvent, keydown: KeyboardEvent }
// - .on() infers the event type from the event name
// - .emit() requires the correct payload type
// The consumer should never need to annotate types manually
```

2. **Fix the partial inference:**

```typescript
// This API requires the user to specify ALL generics manually.
// Refactor it so that at least some are inferred.
function createEndpoint<TInput, TOutput, TError>(config: {
  validate: (input: unknown) => TInput;
  handler: (input: TInput) => TOutput;
  onError: (err: unknown) => TError;
}) { return config; }
```

3. **Explain why this compiles but shouldn't:**

```typescript
function merge<T>(a: T, b: T): T {
  return { ...a, ...b };
}

const result = merge({ x: 1 }, { y: 2 });
// result.x exists? result.y exists? What is result's actual type?
```

---

## Next

→ [Level 04: Conditional Types](../04-conditional-types/)
