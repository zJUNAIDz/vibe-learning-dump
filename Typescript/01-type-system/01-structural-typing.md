# Level 01: Structural Typing — What TypeScript Actually Checks

## The Problem

You've seen this and been confused:

```typescript
interface User {
  name: string;
  age: number;
}

function greet(user: User) {
  return `Hello, ${user.name}`;
}

// This works — why?
const obj = { name: 'Alice', age: 30, email: 'alice@test.com' };
greet(obj);  // ✅ No error

// But THIS fails — why?
greet({ name: 'Alice', age: 30, email: 'alice@test.com' });  // ❌ Error!
```

Most developers learn to work around this. Few understand why.

---

## Mental Model: Structural Typing

TypeScript uses **structural typing** (also called "duck typing for types"). A value matches a type if it has *at least* the required properties with compatible types. Extra properties are fine.

This is the opposite of **nominal typing** (used by Java, C#, Rust), where types must be explicitly declared as the same type to be compatible.

```typescript
// In Java: these are DIFFERENT types, incompatible
class UserId { value: string }
class OrderId { value: string }

// In TypeScript: these are THE SAME type, fully compatible
type UserId = { value: string };
type OrderId = { value: string };

const userId: UserId = { value: '123' };
const orderId: OrderId = userId;  // ✅ No error. Same structure = same type.
```

**Why this matters at scale:** If two types have the same shape, TypeScript considers them interchangeable. This is a feature (flexibility) and a hazard (accidental compatibility). Level 06 shows how to work around this with branded types.

---

## Excess Property Checking: The Exception to the Rule

The object literal example above reveals TypeScript's one major deviation from structural typing:

### Rule: Direct object literals get **excess property checking**

```typescript
interface Point {
  x: number;
  y: number;
}

// Direct assignment of object literal → excess property check kicks in
const p1: Point = { x: 1, y: 2, z: 3 };  // ❌ 'z' does not exist in type 'Point'

// Indirect assignment → no excess property check
const temp = { x: 1, y: 2, z: 3 };
const p2: Point = temp;  // ✅ Fine. temp has x and y, that's enough.

// Function argument (same as direct assignment)
function draw(point: Point) {}
draw({ x: 1, y: 2, z: 3 });  // ❌ Excess property check

// Indirect argument
const coords = { x: 1, y: 2, z: 3 };
draw(coords);  // ✅ Fine
```

### Why does this exception exist?

Excess property checking is a **usability feature**, not a type-safety feature. When you write an object literal directly at the assignment site, you probably made a typo or confused two interfaces. TypeScript catches that.

When the value already exists in a variable, the extra properties are genuinely irrelevant to the type — structural typing takes over.

### Failure Mode: Thinking excess property checking is always active

```typescript
type Config = {
  port: number;
  host: string;
};

function startServer(config: Config) { /* ... */ }

// You expect this to error, but it doesn't:
const myConfig = { port: 3000, host: 'localhost', debug: true };
startServer(myConfig);  // ✅ No error — debug is silently ignored
```

**At scale:** This means your function can receive data it doesn't know about. If you destructure in the function body, the extra properties vanish silently. If you pass the whole object to another system (like a database), extra data might leak through.

---

## Assignability: The Subset Rule

Assignability is the core question the type checker answers: **"Can I use value A where type B is expected?"**

### The Rule

`A` is assignable to `B` if `A` is a **subtype** of `B` — meaning A's set of possible values is a **subset** of B's set.

```typescript
// Literal is assignable to its base type
const a: 'hello' = 'hello';
const b: string = a;  // ✅  {'hello'} ⊆ {all strings}

// Base type is NOT assignable to literal
const c: string = 'world';
const d: 'hello' = c;  // ❌  {all strings} ⊄ {'hello'}
```

### Object subtyping: More properties = more specific = subtype

This is counterintuitive:

```typescript
type Animal = { name: string };
type Dog = { name: string; breed: string };

// Dog is a SUBTYPE of Animal (Dog has MORE properties)
const dog: Dog = { name: 'Rex', breed: 'Labrador' };
const animal: Animal = dog;  // ✅ Dog ⊆ Animal in the type sense

// Why? Because every Dog value is also a valid Animal value.
// The set of all {name, breed} objects is a subset of all {name} objects.
```

**Mental model:** More properties = smaller set (more constraints = fewer valid values) = subtype.

### Function assignability: Contra-variance of parameters

Functions flip the subtyping rule for their parameters:

```typescript
type Handler = (event: MouseEvent) => void;

// Can I use a more general handler where a specific one is expected?
const generalHandler = (event: Event) => { console.log(event.type); };
const mouseHandler: Handler = generalHandler;  // ✅

// Can I use a more specific handler where a general one is expected?
const specificHandler = (event: MouseEvent & { custom: string }) => {};
const handler2: Handler = specificHandler;  // ❌
```

**Why:** A function that accepts `Event` can safely handle any `MouseEvent` (because `MouseEvent` is a subtype of `Event`). But a function that requires `MouseEvent & { custom: string }` can't safely receive a plain `MouseEvent`.

---

## Soundness vs Usability: Where TypeScript Lies

TypeScript is **intentionally unsound** in specific places. "Sound" means the compiler never lets an invalid operation through. TypeScript sacrifices soundness for usability in several cases:

### 1. Covariant arrays

```typescript
const dogs: Dog[] = [{ name: 'Rex', breed: 'Lab' }];
const animals: Animal[] = dogs;  // ✅ TypeScript allows this

animals.push({ name: 'Cat' });  // ✅ No error! But now dogs[1] has no breed!
console.log(dogs[1].breed);     // Runtime: undefined. TypeScript: string.
```

The compiler just lied. `dogs[1].breed` is `undefined` at runtime but TypeScript says it's `string`. This is an **intentional** unsoundness because making arrays invariant would make TypeScript extremely painful to use.

### 2. Bivariant function parameters (in methods)

```typescript
interface Animal {
  makeSound(): void;
}

interface Dog extends Animal {
  fetch(): void;
}

// With --strictFunctionTypes, this is checked correctly for standalone functions
// But methods in interfaces/classes are still bivariant
```

### 3. Type assertions (`as`)

```typescript
const x: unknown = 'hello';
const y: number = x as number;  // ✅ No error. Complete lie.
console.log(y.toFixed(2));      // Runtime crash: "hello".toFixed is not a function
```

### 4. `any` propagation

```typescript
function dangerous(): any {
  return 'not a number';
}

const result: number = dangerous();  // ✅ No error
result.toFixed(2);                   // Runtime crash
```

### Why You Need to Know This

These unsoundnesses are not bugs — they are design decisions. Knowing where the compiler lies means you know where to add runtime validation, where to avoid patterns, and where you need branded types or schema validation (Levels 06 and 07).

---

## Exercises

1. **Predict the behavior:** For each code block, predict whether TypeScript errors, and if so, why. Check your answers in the playground.

```typescript
// Exercise 1
type A = { x: number };
type B = { x: number; y: number };
const b: B = { x: 1, y: 2 };
const a: A = b;
const b2: B = a;  // ?

// Exercise 2
type Fn1 = (x: string | number) => void;
type Fn2 = (x: string) => void;
const fn1: Fn1 = (x) => {};
const fn2: Fn2 = fn1;  // ?
const fn3: Fn1 = ((x: string) => {}) as Fn2;  // ?

// Exercise 3
const arr: readonly number[] = [1, 2, 3];
const arr2: number[] = arr;  // ?
```

2. **Find the unsoundness:** Write a program that compiles with zero errors but crashes at runtime, using only the features described in this file (no `as`, no `any`). Hint: covariant arrays.

---

## Next

→ [02-assignability-and-variance.md](02-assignability-and-variance.md)
