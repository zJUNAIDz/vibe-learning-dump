# Level 02: Inference — TypeScript's Most Powerful Feature

## The Problem

Most developers over-annotate their TypeScript code:

```typescript
// Over-annotated (common mistake for experienced devs)
const users: Array<User> = data.map((item: RawUser): User => ({
  id: item.id as string,
  name: item.name as string,
}));

// Let inference work
const users = data.map((item) => ({
  id: item.id,
  name: item.name,
}));
```

The second version is shorter AND more type-safe — `as` casts can lie, inference cannot.

**Rule of thumb:** Only annotate when inference can't or shouldn't do the job: function parameters, public API return types, and places where you want to widen intentionally.

---

## How Inference Actually Works

### Bottom-Up Inference

TypeScript infers from **values up to types**:

```typescript
const x = 42;           // inferred: 42 (literal)
let y = 42;             // inferred: number (widened)
const z = [1, 2, 3];    // inferred: number[]
const w = { a: 1 };     // inferred: { a: number }
```

This is the simplest form. The compiler looks at the value and determines the narrowest useful type.

### Top-Down / Contextual Typing

TypeScript also infers from **types down to values**. This is called **contextual typing**:

```typescript
// The parameter type of addEventListener is known
// So TypeScript infers `e` as MouseEvent
document.addEventListener('click', (e) => {
  e.clientX;  // ✅ TypeScript knows this is MouseEvent
});

// Array.map knows the callback receives the array element type
const nums = [1, 2, 3];
nums.map((n) => n.toFixed(2));  // n is inferred as number

// Object satisfying an interface
const handler: React.ChangeEventHandler<HTMLInputElement> = (e) => {
  e.target.value;  // ✅ e is fully typed from the annotation
};
```

**Key insight:** Contextual typing flows type information *backwards* — from the expected type to the value being checked. This is what makes callbacks ergonomic in TypeScript.

### Bidirectional Inference

The real power emerges when both directions work simultaneously:

```typescript
// Generic function
function identity<T>(value: T): T {
  return value;
}

// Bottom-up: TypeScript sees 'hello', infers T = 'hello'
const result = identity('hello');  // type: 'hello'

// But if the return is assigned to a typed variable:
const result2: string = identity('hello');  // T still inferred as 'hello'
// The return type is 'hello', which is assignable to string ✅
```

The compiler runs inference in both directions and finds a solution that satisfies all constraints.

---

## Inference in Generic Functions

### How TypeScript Infers Generic Parameters

```typescript
function first<T>(arr: T[]): T | undefined {
  return arr[0];
}

// T is inferred from the argument
first([1, 2, 3]);      // T = number, return: number | undefined
first(['a', 'b']);      // T = string, return: string | undefined
first([1, 'a', true]); // T = string | number | boolean
```

### Inference sites

TypeScript can infer `T` from multiple positions:

```typescript
function merge<T>(a: T, b: T): T {
  return { ...a, ...b };
}

// Both arguments provide inference for T
merge({ x: 1 }, { y: 2 });  // T = { x: number } | { y: number }
// Wait — that might not be what you want...
```

When multiple arguments constrain the same type parameter, TypeScript **unions** the candidates:

```typescript
function pair<T>(a: T, b: T): [T, T] {
  return [a, b];
}

pair(1, 'hello');  // T = string | number, return: [string | number, string | number]
// You might have wanted T = 1 | 'hello', or an error. You got neither.
```

**Failure mode:** Assuming multiple inference sites for the same parameter will cause an error. They don't — TypeScript unions them.

### Inference from return position

```typescript
function createState<T>(initial: T): {
  get: () => T;
  set: (value: T) => void;
} {
  let state = initial;
  return {
    get: () => state,
    set: (value) => { state = value; },
  };
}

const counter = createState(0);  // T = number
counter.set(42);    // ✅
counter.set('hi');  // ❌ Argument of type 'string' is not assignable to 'number'
```

`T` is locked at call time. The returned object's methods enforce that type forever.

---

## `as const` — Preserving Literal Types

### The Problem

```typescript
const config = {
  endpoint: '/api/users',
  method: 'GET',
};
// type: { endpoint: string; method: string }
// We've lost the literal information! method could be any string.

fetch(config.endpoint, { method: config.method });
// ❌ Type 'string' is not assignable to type '"GET" | "POST" | ...'
```

### The Solution

```typescript
const config = {
  endpoint: '/api/users',
  method: 'GET',
} as const;
// type: { readonly endpoint: '/api/users'; readonly method: 'GET' }

fetch(config.endpoint, { method: config.method });  // ✅
```

### What `as const` actually does

1. Makes all properties `readonly`
2. Preserves all literal types (no widening)
3. Arrays become `readonly` tuples

```typescript
const arr = [1, 2, 3] as const;
// type: readonly [1, 2, 3]  — not number[]

const obj = { a: { b: { c: 42 } } } as const;
// type: { readonly a: { readonly b: { readonly c: 42 } } }  — deep readonly, all literals
```

### `const` type parameters (TypeScript 5.0+)

```typescript
// Problem: generic function widens
function routes<T extends Record<string, string>>(config: T): T {
  return config;
}
const r = routes({ home: '/', about: '/about' });
// type: { home: string; about: string }  ← widened!

// Solution: const type parameter
function routes2<const T extends Record<string, string>>(config: T): T {
  return config;
}
const r2 = routes2({ home: '/', about: '/about' });
// type: { readonly home: '/'; readonly about: '/about' }  ← literals preserved!
```

This is how libraries like tRPC and Hono preserve route paths as literal types.

---

## `satisfies` — Validate Without Widening

### The Problem

```typescript
// Option 1: Annotate the type → lose literal types
const palette: Record<string, string | [number, number, number]> = {
  red: [255, 0, 0],
  green: '#00ff00',
  blue: [0, 0, 255],
};
palette.red;  // type: string | [number, number, number]  ← can't tell it's a tuple

// Option 2: No annotation → no validation
const palette2 = {
  red: [255, 0, 0],
  green: '#00ff00',
  blu: [0, 0, 255],  // Typo! No error because there's no schema to check against
};
```

### The Solution

```typescript
const palette = {
  red: [255, 0, 0],
  green: '#00ff00',
  blue: [0, 0, 255],
} satisfies Record<string, string | [number, number, number]>;

palette.red;    // type: [number, number, number]  ← precise!
palette.green;  // type: string
// palette.blu  → ❌ 'blu' typo would be caught during satisfies check
```

`satisfies` says: "Verify this matches the schema, but keep the inferred type."

### Real-world use case: route configuration

```typescript
type RouteConfig = {
  path: string;
  method: 'GET' | 'POST' | 'PUT' | 'DELETE';
  auth?: boolean;
};

const routes = {
  getUsers: { path: '/users', method: 'GET' },
  createUser: { path: '/users', method: 'POST', auth: true },
  deleteUser: { path: '/users/:id', method: 'DELETE', auth: true },
} satisfies Record<string, RouteConfig>;

// routes.getUsers.method is 'GET' (literal), not 'GET' | 'POST' | 'PUT' | 'DELETE'
// But the structure was validated against RouteConfig
```

---

## Inference Loss: When and Why It Happens

### Common causes

1. **Assigning to a wider type annotation:**

```typescript
const x: string = 'hello';  // 'hello' → string (lost)
```

2. **Returning from function without annotation or as const:**

```typescript
function getDirection() {
  return 'north';  // inferred return: string, not 'north'
}
```

3. **Spreading into a new object:**

```typescript
const base = { type: 'user' as const, name: 'Alice' };
const extended = { ...base, role: 'admin' };
// extended.type is still 'user' ✅ (preserved through spread)
// But be careful with functions that take the spread result
```

4. **Generic parameter defaulting to constraint:**

```typescript
function wrap<T extends string>(val: T) { return { val }; }
// If T can't be inferred, it defaults to string, not the literal
```

### Prevention strategies

| Strategy | When to use |
|----------|-------------|
| `as const` | Object/array literals where you need literal types |
| `satisfies` | Validation without widening |
| `const` type parameter | Generic functions that should preserve literals |
| Return type annotation | Public API boundaries |
| Store in `const` variable | Before passing to function |

---

## Exercises

1. **Fix the inference loss:**

```typescript
// This function should return a type where status is 'loading' | 'success' | 'error'
// (literal union, not string)
function createMachine() {
  const states = ['loading', 'success', 'error'];
  return {
    states,
    current: states[0],
    transition(to: typeof states[number]) { /* ... */ }
  };
}
// Fix it so that `states` preserves literal types and `current` is typed correctly
```

2. **Explain why this fails and fix it:**

```typescript
const handlers = {
  click: (e: MouseEvent) => e.clientX,
  keydown: (e: KeyboardEvent) => e.key,
};

function dispatch<K extends keyof typeof handlers>(
  event: K,
  payload: Parameters<typeof handlers[K]>[0]
) {
  return handlers[event](payload);  // ❌ Why does this error?
}
```

---

## Next

→ [Level 03: Advanced Generics](../03-advanced-generics/)
