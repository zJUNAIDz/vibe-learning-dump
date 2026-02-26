# Level 04: Conditional Types — The Brain Bender Zone

## The Problem

You've seen types like this in library code and couldn't parse them:

```typescript
type UnwrapPromise<T> = T extends Promise<infer U> ? UnwrapPromise<U> : T;
type IsNever<T> = [T] extends [never] ? true : false;
type DeepReadonly<T> = T extends Function ? T : { readonly [K in keyof T]: DeepReadonly<T[K]> };
```

Conditional types are the **`if/else` of the type system**. Combined with `infer`, they become pattern matching. Combined with mapped types, they become type-level programs.

---

## Conditional Types: Syntax and Semantics

```typescript
type Check<T> = T extends string ? 'yes' : 'no';
```

This reads: "If `T` is assignable to `string`, resolve to `'yes'`. Otherwise, resolve to `'no'`."

### With concrete types: immediate resolution

```typescript
type A = Check<string>;    // 'yes'
type B = Check<number>;    // 'no'
type C = Check<'hello'>;   // 'yes' — 'hello' extends string
type D = Check<never>;     // never — !!! (explained below)
```

### With generic types: deferred resolution

```typescript
function example<T>(val: T): Check<T> {
  // Inside the function, Check<T> is NOT resolved
  // TypeScript keeps it as Check<T> because T is unknown
  // This means you can't safely return 'yes' or 'no' here
  return 'yes' as any;  // Must assert
}
```

**Failure mode:** Expecting the compiler to narrow conditional types inside generic function bodies. It won't. The conditional type can only resolve when `T` is concrete.

---

## Distributive Conditional Types

This is the single most confusing behavior in TypeScript's type system.

### The Rule

When a conditional type acts on a **naked type parameter** with a **union type**, it **distributes** over each member of the union:

```typescript
type ToArray<T> = T extends any ? T[] : never;

// With a union:
type Result = ToArray<string | number>;
// Distributes: ToArray<string> | ToArray<number>
// = string[] | number[]
// NOT: (string | number)[]
```

### "Naked type parameter" means:

```typescript
type Naked<T> = T extends string ? 'yes' : 'no';     // T is naked → distributes
type Wrapped<T> = [T] extends [string] ? 'yes' : 'no'; // T is wrapped → does NOT distribute
```

### Why distribution matters

```typescript
type Exclude<T, U> = T extends U ? never : T;

// Exclude<'a' | 'b' | 'c', 'a'>
// Distributes to:
// ('a' extends 'a' ? never : 'a') | ('b' extends 'a' ? never : 'b') | ('c' extends 'a' ? never : 'c')
// = never | 'b' | 'c'
// = 'b' | 'c'
```

Without distribution, `Exclude` wouldn't work — the whole union `'a' | 'b' | 'c'` extends `'a'` is false, so it would just return the whole union unchanged.

### Preventing distribution

Sometimes you DON'T want distribution:

```typescript
// Distributive (wrong for this use case)
type IsString<T> = T extends string ? true : false;
type Test1 = IsString<string | number>;  // boolean (= true | false) — WRONG

// Non-distributive (correct)
type IsString2<T> = [T] extends [string] ? true : false;
type Test2 = IsString2<string | number>;  // false — correct!
```

### The `never` trap

`never` is the empty union. Distributing over an empty union produces `never`:

```typescript
type Check<T> = T extends string ? 'yes' : 'no';
type Result = Check<never>;  // never (not 'yes' or 'no')

// Because: distribute over nothing → nothing

// To handle never correctly:
type Check2<T> = [T] extends [never] ? 'empty' : T extends string ? 'yes' : 'no';
type Result2 = Check2<never>;  // 'empty'
```

---

## The `infer` Keyword: Type-Level Pattern Matching

`infer` declares a type variable that TypeScript "solves" by pattern matching against the structure:

### Basic patterns

```typescript
// Extract return type
type GetReturn<T> = T extends (...args: any[]) => infer R ? R : never;
type T1 = GetReturn<() => string>;          // string
type T2 = GetReturn<(x: number) => void>;   // void

// Extract array element
type ElementOf<T> = T extends (infer E)[] ? E : never;
type T3 = ElementOf<string[]>;   // string
type T4 = ElementOf<[1, 2, 3]>;  // 1 | 2 | 3

// Extract promise value
type Unwrap<T> = T extends Promise<infer V> ? V : T;
type T5 = Unwrap<Promise<string>>;     // string
type T6 = Unwrap<Promise<number[]>>;   // number[]
type T7 = Unwrap<string>;              // string (passthrough)
```

### `infer` with constraints (TypeScript 4.7+)

```typescript
// infer R extends string → R is constrained to string
type GetStringReturn<T> = T extends (...args: any[]) => infer R extends string ? R : never;

type T1 = GetStringReturn<() => 'hello'>;   // 'hello'
type T2 = GetStringReturn<() => number>;    // never (return isn't string)
```

### `infer` in template literals

```typescript
// Parse route parameters
type ExtractParams<T extends string> =
  T extends `${string}:${infer Param}/${infer Rest}`
    ? Param | ExtractParams<Rest>
    : T extends `${string}:${infer Param}`
      ? Param
      : never;

type Params = ExtractParams<'/users/:userId/posts/:postId'>;
// type: 'userId' | 'postId'
```

This is how Express-like routers (e.g., Hono) extract route parameter names at the type level.

### `infer` in multiple positions

```typescript
// Extract first and rest of a tuple
type Head<T extends any[]> = T extends [infer H, ...any[]] ? H : never;
type Tail<T extends any[]> = T extends [any, ...infer R] ? R : never;

type T1 = Head<[1, 2, 3]>;  // 1
type T2 = Tail<[1, 2, 3]>;  // [2, 3]

// Extract first and last
type FirstAndLast<T extends any[]> =
  T extends [infer First, ...any[], infer Last] ? [First, Last] : never;

type T3 = FirstAndLast<[1, 2, 3, 4]>;  // [1, 4]
```

### Co-variant vs contra-variant `infer` positions

When `infer` appears in multiple positions for the same variable, TypeScript resolves differently based on variance:

```typescript
// Covariant positions → union
type Foo<T> = T extends { a: infer U; b: infer U } ? U : never;
type T1 = Foo<{ a: string; b: number }>;  // string | number

// Contravariant positions → intersection
type Bar<T> = T extends { a: (x: infer U) => void; b: (x: infer U) => void } ? U : never;
type T2 = Bar<{ a: (x: string) => void; b: (x: number) => void }>;  // string & number = never
```

**Why this matters:** This is how you can build `UnionToIntersection<T>`:

```typescript
type UnionToIntersection<U> =
  (U extends any ? (k: U) => void : never) extends (k: infer I) => void ? I : never;

type Result = UnionToIntersection<{ a: 1 } | { b: 2 }>;
// type: { a: 1 } & { b: 2 }
```

---

## Building Type-Level Functions

Conditional types with generics are effectively functions at the type level.

### Recursive conditional types

```typescript
// Deep flatten
type Flatten<T> = T extends Array<infer U> ? Flatten<U> : T;

type T1 = Flatten<number[][][]>;        // number
type T2 = Flatten<string[]>;            // string
type T3 = Flatten<Array<Array<boolean>>>;  // boolean

// Deep Awaited (built into TS as Awaited<T>)
type DeepAwaited<T> =
  T extends Promise<infer U> ? DeepAwaited<U> : T;

type T4 = DeepAwaited<Promise<Promise<Promise<string>>>>;  // string
```

### Type-level string manipulation

```typescript
// Snake to camel case
type SnakeToCamel<S extends string> =
  S extends `${infer Head}_${infer Tail}`
    ? `${Head}${Capitalize<SnakeToCamel<Tail>>}`
    : S;

type T1 = SnakeToCamel<'user_first_name'>;  // 'userFirstName'
type T2 = SnakeToCamel<'id'>;               // 'id'
```

### Tuple manipulation

```typescript
// Reverse a tuple
type Reverse<T extends any[]> =
  T extends [infer Head, ...infer Tail]
    ? [...Reverse<Tail>, Head]
    : [];

type T1 = Reverse<[1, 2, 3]>;  // [3, 2, 1]

// Remove element from tuple
type Without<T extends any[], U> =
  T extends [infer Head, ...infer Tail]
    ? Head extends U
      ? Without<Tail, U>
      : [Head, ...Without<Tail, U>]
    : [];

type T2 = Without<[1, 2, 3, 2, 1], 2>;  // [1, 3, 1]
```

---

## Common Pitfalls

### 1. Forgetting distribution

```typescript
// "Why does this return boolean instead of true?"
type IsString<T> = T extends string ? true : false;
type Test = IsString<string | number>;  // boolean (true | false)
```

### 2. Infinite recursion

```typescript
// ❌ No base case for non-array types
type Bad<T> = T extends Array<infer U> ? Bad<U> : Bad<T>;  // Infinite

// ✅ Base case handles non-matching types
type Good<T> = T extends Array<infer U> ? Good<U> : T;
```

### 3. Using `infer` without enough structure

```typescript
// ❌ Ambiguous — what is T matching against?
type Bad<T> = T extends infer U ? U : never;  // U = T always. Useless.

// ✅ Meaningful pattern match
type Good<T> = T extends { data: infer D } ? D : never;
```

---

## Exercises

1. **Build `DeepRequired<T>`:** Make all properties required recursively, including nested objects.

2. **Build `PathKeys<T>`:** Given a nested object type, produce a union of dot-path strings:

```typescript
type User = {
  name: string;
  address: {
    street: string;
    city: string;
    geo: { lat: number; lng: number };
  };
};

type Keys = PathKeys<User>;
// 'name' | 'address' | 'address.street' | 'address.city' | 'address.geo' | 'address.geo.lat' | 'address.geo.lng'
```

3. **Build `ParseQueryString<T>`:** Parse a URL query string into a type:

```typescript
type Result = ParseQueryString<'name=alice&age=30&role=admin'>;
// { name: string; age: string; role: string }
```

---

## Next

→ [Level 05: Mapped Types & Key Remapping](../05-mapped-types/)
