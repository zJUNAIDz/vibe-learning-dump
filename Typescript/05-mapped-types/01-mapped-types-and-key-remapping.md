# Level 05: Mapped Types & Key Remapping

## The Problem

You need to transform types at scale:

- Make all properties of an API response optional for patch endpoints
- Convert snake_case database columns to camelCase frontend types
- Create getter/setter interfaces from a plain data type
- Build deep readonly types for immutable stores

Mapped types are the **`for` loop of the type system** — they iterate over keys and transform each one.

---

## Mapped Types: The Core Mechanism

```typescript
type Mapped<T> = {
  [K in keyof T]: T[K];
};
```

This iterates over every key `K` in `T` and maps it to the same value type `T[K]`. By itself it's an identity transform. Power comes from modifying the key, the value, or both.

### Modifier manipulation

```typescript
// Add readonly
type Readonly<T> = { readonly [K in keyof T]: T[K] };

// Remove readonly
type Mutable<T> = { -readonly [K in keyof T]: T[K] };

// Add optional
type Partial<T> = { [K in keyof T]?: T[K] };

// Remove optional
type Required<T> = { [K in keyof T]-?: T[K] };

// Combine: make mutable and required
type Concrete<T> = { -readonly [K in keyof T]-?: T[K] };
```

The `-` prefix *removes* the modifier. Without `-`, adding `readonly` or `?` adds it.

### Value transformation

```typescript
type Nullable<T> = { [K in keyof T]: T[K] | null };
type Promisified<T> = { [K in keyof T]: Promise<T[K]> };
type Boxed<T> = { [K in keyof T]: { value: T[K] } };
```

---

## `keyof` — Dangerous If Misused

### What `keyof` actually returns

```typescript
type User = { name: string; age: number; email: string };
type UserKeys = keyof User;  // 'name' | 'age' | 'email'
```

### Trap 1: `keyof` with index signatures

```typescript
type Dict = { [key: string]: unknown };
type Keys = keyof Dict;  // string | number  ← NOT just string!
```

Why `number`? Because JavaScript coerces number keys to strings (`obj[0]` is `obj['0']`). TypeScript models this by making `keyof { [key: string]: X }` return `string | number`.

### Trap 2: `keyof` on unions

```typescript
type A = { x: number; y: number };
type B = { y: number; z: number };

type Keys = keyof (A | B);  // 'y' — only COMMON keys!
// NOT 'x' | 'y' | 'z'
```

`keyof` on a union returns the **intersection** of keys (keys present in ALL members). This follows from the set theory: if you have an `A | B`, you can only safely access properties that exist on both.

To get ALL keys:

```typescript
type AllKeys = keyof A | keyof B;  // 'x' | 'y' | 'z'
```

### Trap 3: `keyof` on intersections

```typescript
type Keys = keyof (A & B);  // 'x' | 'y' | 'z' — ALL keys
```

Opposite of unions. Intersection has all properties of both types.

### Using `keyof` safely

```typescript
// Safe: iterate only over known keys
function getProperty<T, K extends keyof T>(obj: T, key: K): T[K] {
  return obj[key];
}

// Dangerous: string index access on objects
function unsafeGet(obj: Record<string, unknown>, key: string) {
  return obj[key];  // Can't know if key exists. Enable noUncheckedIndexedAccess.
}
```

---

## Key Remapping (TypeScript 4.1+)

Key remapping lets you **transform the keys** during a mapped type, not just the values:

```typescript
type Getters<T> = {
  [K in keyof T as `get${Capitalize<string & K>}`]: () => T[K];
};

type User = { name: string; age: number };
type UserGetters = Getters<User>;
// { getName: () => string; getAge: () => number }
```

### The `as` clause in mapped types

```typescript
// General form:
type Mapped<T> = {
  [K in keyof T as NewKey]: NewValue;
};
```

### Filtering keys with `as never`

Remapping a key to `never` removes it:

```typescript
// Keep only string-valued properties
type StringProps<T> = {
  [K in keyof T as T[K] extends string ? K : never]: T[K];
};

type User = { name: string; age: number; email: string };
type StringUser = StringProps<User>;
// { name: string; email: string }

// Remove specific keys
type OmitByKey<T, Keys extends keyof T> = {
  [K in keyof T as K extends Keys ? never : K]: T[K];
};

type WithoutAge = OmitByKey<User, 'age'>;
// { name: string; email: string }
```

### Renaming keys

```typescript
// Snake to camel case for all keys
type CamelCase<S extends string> =
  S extends `${infer Head}_${infer Tail}`
    ? `${Head}${Capitalize<CamelCase<Tail>>}`
    : S;

type CamelCaseKeys<T> = {
  [K in keyof T as CamelCase<string & K>]: T[K];
};

type DBRow = {
  user_id: number;
  first_name: string;
  last_name: string;
  created_at: Date;
};

type FrontendUser = CamelCaseKeys<DBRow>;
// { userId: number; firstName: string; lastName: string; createdAt: Date }
```

---

## Template Literal Types

Template literal types let you construct string types from other types:

```typescript
type EventName = `on${Capitalize<'click' | 'focus' | 'blur'>}`;
// 'onClick' | 'onFocus' | 'onBlur'
```

### Combination explosion

Template literals distribute over unions in each position:

```typescript
type Size = 'sm' | 'md' | 'lg';
type Color = 'red' | 'blue';
type ClassName = `${Size}-${Color}`;
// 'sm-red' | 'sm-blue' | 'md-red' | 'md-blue' | 'lg-red' | 'lg-blue'
```

### Built-in string manipulation types

```typescript
type T1 = Uppercase<'hello'>;     // 'HELLO'
type T2 = Lowercase<'HELLO'>;     // 'hello'
type T3 = Capitalize<'hello'>;    // 'Hello'
type T4 = Uncapitalize<'Hello'>;  // 'hello'
```

### Parsing with template literals

```typescript
// Type-safe route parser
type ParseRoute<T extends string> =
  T extends `${string}:${infer Param}/${infer Rest}`
    ? { [K in Param | keyof ParseRoute<Rest>]: string }
    : T extends `${string}:${infer Param}`
      ? { [K in Param]: string }
      : {};

type Params = ParseRoute<'/users/:id/posts/:postId'>;
// { id: string; postId: string }
```

### Real-world: CSS-in-JS property types

```typescript
type CSSLength = `${number}${'px' | 'em' | 'rem' | '%' | 'vh' | 'vw'}`;

function setWidth(width: CSSLength) { /* ... */ }

setWidth('100px');   // ✅
setWidth('2.5em');   // ✅
setWidth('100');     // ❌ Missing unit
setWidth('100km');   // ❌ Invalid unit
```

---

## Deep vs Shallow Transformations

### Shallow (built-in utilities)

```typescript
// Partial<T> only affects top-level properties
type User = {
  name: string;
  address: {
    street: string;
    city: string;
  };
};

type PartialUser = Partial<User>;
// { name?: string; address?: { street: string; city: string } }
// address is optional, but if provided, street and city are REQUIRED
```

### Deep transformations

```typescript
type DeepPartial<T> = T extends object
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : T;

type DeepPartialUser = DeepPartial<User>;
// { name?: string; address?: { street?: string; city?: string } }
```

### Deep readonly

```typescript
type DeepReadonly<T> =
  T extends Function ? T :
  T extends Map<infer K, infer V> ? ReadonlyMap<DeepReadonly<K>, DeepReadonly<V>> :
  T extends Set<infer V> ? ReadonlySet<DeepReadonly<V>> :
  T extends Array<infer V> ? ReadonlyArray<DeepReadonly<V>> :
  T extends object ? { readonly [K in keyof T]: DeepReadonly<T[K]> } :
  T;
```

Note the special handling for `Function`, `Map`, `Set`, and `Array`. Without these guards, the recursion would break on built-in types.

### Deep Required

```typescript
type DeepRequired<T> = T extends object
  ? { [K in keyof T]-?: DeepRequired<NonNullable<T[K]>> }
  : T;
```

**Failure mode:** Forgetting to handle built-in types. Recursing into `Date`, `RegExp`, `Function`, `Map`, etc. produces garbage types. Always guard with `T extends Function ? T :` or similar.

---

## Practical Patterns

### Pattern: Pick-and-transform

```typescript
// Make only certain keys optional
type PartialBy<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>;

type UserUpdate = PartialBy<User, 'email' | 'age'>;
// { name: string } & { email?: string; age?: number }
```

### Pattern: Create API response types

```typescript
type APIResponse<T> = {
  [K in keyof T as `${string & K}`]: T[K] extends Date
    ? string  // Dates become strings in JSON
    : T[K] extends object
      ? APIResponse<T[K]>
      : T[K];
};
```

### Pattern: Event handler map

```typescript
type EventHandlers<Events extends Record<string, any>> = {
  [K in keyof Events as `on${Capitalize<string & K>}`]: (event: Events[K]) => void;
};

type DOMEvents = {
  click: MouseEvent;
  keydown: KeyboardEvent;
  scroll: Event;
};

type Handlers = EventHandlers<DOMEvents>;
// { onClick: (event: MouseEvent) => void; onKeydown: ... ; onScroll: ... }
```

---

## Exercises

1. **Build `Merge<A, B>`**: Properties in B override properties in A. nested objects are merged recursively.

2. **Build `Paths<T>`**: Generate all valid dot-notation paths for a nested object, with correct value types:

```typescript
type Get<T, P extends string> = /* your implementation */;

type User = {
  name: string;
  address: { street: string; zip: number };
};

type T = Get<User, 'address.street'>;  // string
type T2 = Get<User, 'name'>;           // string
type T3 = Get<User, 'address.zip'>;    // number
```

3. **Build `SnakeCaseKeys<T>`**: Recursively convert all keys from camelCase to snake_case.

---

## Next

→ [Level 06: Type-Level Programming Patterns](../06-type-level-programming/)
