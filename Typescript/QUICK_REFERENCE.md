# TypeScript Deep Dive — Quick Reference

> Quick lookup for advanced type-level patterns. Not a tutorial — a cheat sheet.

---

## Type-Level Operators

```typescript
// Extract / Exclude
type T1 = Extract<'a' | 'b' | 'c', 'a' | 'b'>;  // 'a' | 'b'
type T2 = Exclude<'a' | 'b' | 'c', 'a'>;          // 'b' | 'c'

// NonNullable
type T3 = NonNullable<string | null | undefined>;  // string

// ReturnType / Parameters
type T4 = ReturnType<typeof JSON.parse>;            // any
type T5 = Parameters<typeof setTimeout>;            // [callback: ..., ms?: ...]

// Awaited (unwrap Promise)
type T6 = Awaited<Promise<Promise<string>>>;        // string

// ConstructorParameters / InstanceType
type T7 = InstanceType<typeof Map>;                 // Map<any, any>
```

---

## Inference Patterns

```typescript
// Infer return type
type GetReturn<T> = T extends (...args: any[]) => infer R ? R : never;

// Infer array element
type ElementOf<T> = T extends readonly (infer E)[] ? E : never;

// Infer promise value
type UnwrapPromise<T> = T extends Promise<infer V> ? V : T;

// Infer first argument
type FirstArg<T> = T extends (first: infer A, ...rest: any[]) => any ? A : never;

// Infer key from template literal
type GetParam<S> = S extends `${string}:${infer Param}/${infer Rest}`
  ? Param | GetParam<Rest>
  : S extends `${string}:${infer Param}`
    ? Param
    : never;
```

---

## Mapped Type Patterns

```typescript
// Make all properties optional (deep)
type DeepPartial<T> = {
  [K in keyof T]?: T[K] extends object ? DeepPartial<T[K]> : T[K];
};

// Make specific keys required
type RequireKeys<T, K extends keyof T> = T & Required<Pick<T, K>>;

// Rename keys
type Getters<T> = {
  [K in keyof T as `get${Capitalize<string & K>}`]: () => T[K];
};

// Filter keys by value type
type KeysOfType<T, V> = {
  [K in keyof T]: T[K] extends V ? K : never;
}[keyof T];
```

---

## Conditional Type Patterns

```typescript
// Distributive (operates on each union member)
type ToArray<T> = T extends any ? T[] : never;
// ToArray<string | number> → string[] | number[]

// Non-distributive (wrapping in tuple)
type ToArrayND<T> = [T] extends [any] ? T[] : never;
// ToArrayND<string | number> → (string | number)[]

// Recursive conditional
type Flatten<T> = T extends Array<infer U> ? Flatten<U> : T;
// Flatten<number[][][]> → number
```

---

## Branded Types

```typescript
// Declaration
declare const __brand: unique symbol;
type Brand<T, B> = T & { readonly [__brand]: B };

// Usage
type UserId = Brand<string, 'UserId'>;
type OrderId = Brand<string, 'OrderId'>;

// Constructor
function userId(id: string): UserId { return id as UserId; }
function orderId(id: string): OrderId { return id as OrderId; }
```

---

## Exhaustiveness

```typescript
// The never trick
function assertNever(x: never): never {
  throw new Error(`Unexpected value: ${x}`);
}

// Usage in switch
type Shape = { kind: 'circle'; r: number } | { kind: 'square'; s: number };

function area(shape: Shape): number {
  switch (shape.kind) {
    case 'circle': return Math.PI * shape.r ** 2;
    case 'square': return shape.s ** 2;
    default: return assertNever(shape); // Compile error if a case is missing
  }
}
```

---

## Variance Markers

```typescript
// out = covariant (produces T)
// in = contravariant (consumes T)
// in out = invariant
type Producer<out T> = { get(): T };
type Consumer<in T> = { accept(value: T): void };
type Both<in out T> = { transform(value: T): T };
```

---

## Satisfies Operator

```typescript
// Validates type WITHOUT widening
const routes = {
  home: '/',
  about: '/about',
  user: '/user/:id',
} satisfies Record<string, string>;

// routes.home is '/' (literal), not string
// But TS verified it matches Record<string, string>
```

---

## Const Type Parameters (5.0+)

```typescript
// Without: T inferred as string[]
function old<T extends readonly string[]>(arr: T) { return arr; }

// With: T inferred as readonly ['a', 'b']
function withConst<const T extends readonly string[]>(arr: T) { return arr; }

const result = withConst(['a', 'b']);
// type: readonly ["a", "b"]
```

---

## Discriminated Unions — Advanced

```typescript
// Result pattern
type Result<T, E = Error> =
  | { success: true; data: T }
  | { success: false; error: E };

// With exhaustive matching
function handle<T>(result: Result<T>): T {
  if (result.success) return result.data;
  throw result.error;
}
```

---

## Template Literal Types

```typescript
type EventName<T extends string> = `on${Capitalize<T>}`;
// EventName<'click'> → 'onClick'

type CSSProperty = `${string}-${string}`;
// Matches 'font-size', 'margin-top', etc.

// Parsing
type Split<S extends string, D extends string> =
  S extends `${infer Head}${D}${infer Tail}`
    ? [Head, ...Split<Tail, D>]
    : [S];
// Split<'a.b.c', '.'> → ['a', 'b', 'c']
```

---

## Common "Why Does This Work?" Patterns

```typescript
// Why `& {}` prevents distribution
type NoDistribute<T> = T & {};

// Why `[T] extends [any]` prevents distribution
type IsNever<T> = [T] extends [never] ? true : false;

// Why `T extends T` forces distribution
type Distribute<T> = T extends T ? { value: T }[] : never;

// Why `keyof any` is `string | number | symbol`
type AllKeys = keyof any;

// Why `{} extends T` checks for unconstrained T
type IsAny<T> = 0 extends (1 & T) ? true : false;
```
