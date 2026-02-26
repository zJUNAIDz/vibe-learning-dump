# Level 10: Debugging the Type System

## The Problem

You stare at a type error that spans 30 lines. The error message references types you didn't write, from files you've never opened. The hover tooltip shows a type so complex it wraps three times. You randomly add `as any` until it compiles.

This level teaches you to debug types the way you debug runtime code: systematically.

---

## How to "Print" Types Mentally

### Tool 1: Hover (the type REPL)

The most basic but underused debugging tool. In VS Code, hover over any expression to see its resolved type.

```typescript
const x = createRouter(config);
//    ^ hover here to see what createRouter returned
```

**Pro tip:** Use `// ^?` comments with the Twoslash VS Code extension:

```typescript
const x = ['hello', 42] as const;
//    ^? readonly ["hello", 42]
```

### Tool 2: Intermediate type variables

When a complex expression produces an unexpected type, break it into pieces:

```typescript
// ❌ Can't debug this
type Result = Extract<keyof Omit<User, 'id'>, `${string}Name`>;

// ✅ Break it apart
type Step1 = Omit<User, 'id'>;           // { name: string; firstName: string; age: number }
type Step2 = keyof Step1;                 // 'name' | 'firstName' | 'age'
type Step3 = Extract<Step2, `${string}Name`>;  // 'firstName'
// Now hover each step
```

### Tool 3: The `Resolve` / `Prettify` helper

```typescript
// Force TypeScript to expand a type for hover
type Resolve<T> = T extends Function ? T : { [K in keyof T]: T[K] };

// Instead of seeing: Pick<User, 'name'> & Omit<Options, 'debug'>
// You see: { name: string; verbose: boolean; timeout: number }

type Ugly = Pick<User, 'name'> & Partial<Pick<User, 'age'>>;
type Readable = Resolve<Ugly>;
// Hover: { name: string; age?: number }
```

### Tool 4: The `Expect` / `IsEqual` assertion

```typescript
// Type-level assertion — compile error if types don't match
type IsEqual<A, B> =
  (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2)
    ? true
    : false;

type Assert<T extends true> = T;

// Now use it to verify your understanding:
type _test1 = Assert<IsEqual<ReturnType<typeof myFn>, string>>;     // ✅ compiles
type _test2 = Assert<IsEqual<ReturnType<typeof myFn>, number>>;     // ❌ compile error
```

This technique is used by `type-challenges` and testing libraries like `tsd`.

---

## Reading Type Error Messages

### Anatomy of a type error

```
error TS2322: Type '{ name: string; age: string; }' is not assignable to type 'User'.
  Types of property 'age' are incompatible.
    Type 'string' is not assignable to type 'number'.
```

Read bottom-up:
1. **Root cause** (last line): `string` is not `number`
2. **Location** (middle): It's the `age` property
3. **Context** (first line): You tried to assign `{ name: string; age: string }` to `User`

### Complex errors with generics

```
error TS2345: Argument of type '(input: { id: string; }) => { name: string; }'
is not assignable to parameter of type '(input: ExtractInput<TSchema>) =>
ExtractOutput<TSchema>'.
  Types of parameters 'input' and 'input' are incompatible.
    Type 'ExtractInput<TSchema>' is not assignable to type '{ id: string; }'.
```

Translation:
- You passed a callback to a generic function
- The callback expects `{ id: string }` as input
- But the generic function resolved the input type to `ExtractInput<TSchema>`
- Those two types don't match

**Fix strategy:** Hover over `TSchema` to see what it resolved to. Then check `ExtractInput<TSchema>` by breaking it into a separate `type Debug = ExtractInput<typeof yourSchema>`.

### The "excessively deep" error

```
error TS2589: Type instantiation is excessively deep and possibly infinite.
```

This means a recursive type hit the depth or instantiation limit. Common causes:

1. Recursive conditional type without proper base case
2. Template literal type with too many permutations
3. Mapped type recursing into itself through circular references

**Debug strategy:**

```typescript
// Add a depth counter
type DeepPartial<T, Depth extends any[] = []> =
  Depth['length'] extends 10  // Max depth
    ? T
    : T extends object
      ? { [K in keyof T]?: DeepPartial<T[K], [...Depth, any]> }
      : T;
```

---

## Debugging Conditional Types

### The distribution trap

```typescript
type Debug<T> = T extends string ? 'string' : 'other';
type Result = Debug<string | number>;  // 'string' | 'other'  — not 'other'!
```

**Debug tool:** Wrap in tuple to see non-distributed behavior:

```typescript
type Debug<T> = [T] extends [string] ? 'string' : 'other';
type Result = Debug<string | number>;  // 'other'
```

### The `never` input trap

```typescript
type Debug<T> = T extends string ? 'yes' : 'no';
type Result = Debug<never>;  // never — not 'yes' or 'no'
```

**Debug tool:** Always check with `never` when writing conditional types:

```typescript
type IsNever<T> = [T] extends [never] ? true : false;
type Check = IsNever<never>;  // true
```

### Stepping through `infer`

When `infer` doesn't extract what you expect:

```typescript
type GetReturn<T> = T extends (...args: any[]) => infer R ? R : never;

// Debug: what does T actually look like?
type Debug = typeof myFunction;
// Hover: (x: string, y: number) => Promise<User>

type Result = GetReturn<typeof myFunction>;
// Hover: Promise<User>  ← correct

// What if it's never?
type Result2 = GetReturn<string>;
// Hover: never ← string doesn't match function pattern
```

---

## Introspection Helper Types

### Type identity checker

```typescript
type TypeName<T> =
  T extends string ? 'string' :
  T extends number ? 'number' :
  T extends boolean ? 'boolean' :
  T extends undefined ? 'undefined' :
  T extends null ? 'null' :
  T extends symbol ? 'symbol' :
  T extends bigint ? 'bigint' :
  T extends (...args: any[]) => any ? 'function' :
  T extends any[] ? 'array' :
  T extends object ? 'object' :
  'unknown';

type T1 = TypeName<string>;           // 'string'
type T2 = TypeName<() => void>;       // 'function'
type T3 = TypeName<{ a: 1 }>;         // 'object'
type T4 = TypeName<string | number>;  // 'string' | 'number' (distributive!)
```

### Check if type is `any`

```typescript
type IsAny<T> = 0 extends (1 & T) ? true : false;

type T1 = IsAny<any>;      // true
type T2 = IsAny<unknown>;  // false
type T3 = IsAny<string>;   // false
```

Why this works: `1 & T` for most types `T` is just `1`. But `1 & any` is `any`. And `0 extends any` is `true`.

### Check if type is a union

```typescript
type IsUnion<T, C = T> =
  T extends C
    ? [C] extends [T]
      ? false
      : true
    : never;

type T1 = IsUnion<string>;           // false
type T2 = IsUnion<string | number>;  // true
type T3 = IsUnion<never>;            // never (edge case)
```

### Check if two types are exactly equal

```typescript
type IsExact<A, B> =
  (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2)
    ? true
    : false;

type T1 = IsExact<string, string>;     // true
type T2 = IsExact<string, number>;     // false
type T3 = IsExact<any, string>;        // false (any ≠ string)
type T4 = IsExact<never, never>;       // true
type T5 = IsExact<1 | 2, 2 | 1>;      // true (unions are order-independent)
```

---

## Refactoring Impossible Types

### Strategy 1: Replace conditional types with mapped types

```typescript
// ❌ Hard to read, fragile
type FormValue<T> =
  T extends string ? string :
  T extends number ? number :
  T extends boolean ? boolean :
  T extends Date ? string :
  T extends Array<infer U> ? FormValue<U>[] :
  T extends object ? { [K in keyof T]: FormValue<T[K]> } :
  never;

// ✅ Clearer with a lookup table
type FormValueMap = {
  string: string;
  number: number;
  boolean: boolean;
};

type FormValue<T> =
  T extends Date ? string :
  T extends Array<infer U> ? FormValue<U>[] :
  T extends object ? { [K in keyof T]: FormValue<T[K]> } :
  T;
```

### Strategy 2: Break complex types into named compositions

```typescript
// ❌ Unreadable
type APIHandler<T extends RouteConfig> = (
  req: Omit<Request, 'body'> & { body: ExtractInput<T['schema']> },
  ctx: T['middleware'] extends Middleware<infer C> ? C : {}
) => Promise<ExtractOutput<T['schema']>>;

// ✅ Name each piece
type WithBody<T> = Omit<Request, 'body'> & { body: T };
type MiddlewareContext<T> = T extends Middleware<infer C> ? C : {};

type APIHandler<T extends RouteConfig> = (
  req: WithBody<ExtractInput<T['schema']>>,
  ctx: MiddlewareContext<T['middleware']>
) => Promise<ExtractOutput<T['schema']>>;
```

### Strategy 3: Use interfaces for recursive types

```typescript
// ❌ Type alias recursion can hit limits
type JSONValue = string | number | boolean | null | JSONValue[] | { [key: string]: JSONValue };
// This works, but only because TS special-cases it. More complex recursion fails.

// ✅ Interface recursion is better supported
interface JSONObject { [key: string]: JSONValue }
interface JSONArray extends Array<JSONValue> {}
type JSONValue = string | number | boolean | null | JSONArray | JSONObject;
```

### Strategy 4: Replace type-level computation with codegen

Sometimes the right answer is "don't do this at the type level":

```typescript
// ❌ 500-line recursive conditional type to parse SQL
type ParseSQL<T extends string> = /* absurdly complex type */ ;

// ✅ Use a code generator (like sqlc, Prisma, or custom script)
// Generate the types from the SQL schema at build time
// The types are simple, readable, and fast
```

---

## The Debugging Workflow

When you hit a type error you can't explain:

1. **Read the error bottom-up.** The last line is the root cause.
2. **Hover over the expression.** See what type TypeScript actually inferred.
3. **Break complex types into intermediate variables.** Hover each one.
4. **Use `Resolve<T>` to flatten intersections and mapped types** for readability.
5. **Check edge cases:** What happens with `never`? `any`? `unknown`? Empty objects? Unions?
6. **Minimize the reproduction.** Remove code until only the error remains. This often reveals the cause.
7. **Check `extends` direction.** "A is not assignable to B" means A is not a subset of B. Think about which set is bigger.

---

## Exercises

1. **Debug this error** without using `as any`:

```typescript
type EventMap = {
  click: { x: number; y: number };
  keydown: { key: string; code: string };
};

function on<K extends keyof EventMap>(
  event: K,
  handler: (data: EventMap[K]) => void
) { /* ... */ }

const handlers = {
  click: (data: { x: number; y: number }) => console.log(data.x),
  keydown: (data: { key: string; code: string }) => console.log(data.key),
};

function registerAll() {
  for (const [event, handler] of Object.entries(handlers)) {
    on(event, handler);  // ❌ Error. Why? Fix it.
  }
}
```

2. **Build a type test suite:** Write `IsExact` assertions for at least 10 types in your current project. How many surprise you?

3. **Find the performance bottleneck:** Create a type that causes `"Type instantiation is excessively deep"` and then fix it by adding depth limiting.

---

## Next

→ [Level 11: Capstone & Wizard Checklist](../11-capstone/)
