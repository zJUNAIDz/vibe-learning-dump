# Level 01: Assignability, Variance, and Widening

## Assignability in Depth

### The Assignability Algorithm

When TypeScript checks `A extends B` (or equivalently, `const b: B = aValue`), it's asking:

**"Does every value that satisfies `A` also satisfy `B`?"**

This is checked structurally:

```typescript
type A = { x: number; y: number };
type B = { x: number };

// A extends B?
// Does every {x: number, y: number} also satisfy {x: number}?
// Yes → A is assignable to B

type Check = A extends B ? true : false;  // true
```

### Assignability with unions

```typescript
// Is 'a' | 'b' assignable to string?
// Is every value in {'a', 'b'} also in {all strings}? Yes.
type T1 = ('a' | 'b') extends string ? true : false;  // true

// Is string assignable to 'a' | 'b'?
// Is every value in {all strings} also in {'a', 'b'}? No.
type T2 = string extends ('a' | 'b') ? true : false;  // false
```

### The empty object type `{}`

```typescript
// {} means "any value with no required properties"
// Almost everything is assignable to {} — except null and undefined
type T1 = string extends {} ? true : false;     // true
type T2 = number extends {} ? true : false;     // true
type T3 = { x: 1 } extends {} ? true : false;  // true
type T4 = null extends {} ? true : false;       // false
type T5 = undefined extends {} ? true : false;  // false
```

**Failure mode:** Using `{}` when you mean "any object." Use `Record<string, unknown>` or `object` instead.

---

## Variance: How Container Types Behave

Variance describes how subtyping of **contained types** affects subtyping of **container types**.

### Four kinds of variance

| Variance | Meaning | Example |
|----------|---------|---------|
| **Covariant** | Container preserves subtype direction | `Promise<Dog>` → `Promise<Animal>` ✅ |
| **Contravariant** | Container reverses subtype direction | `(animal: Animal) => void` → `(dog: Dog) => void` ✅ |
| **Invariant** | Container allows neither direction | Only exact type match |
| **Bivariant** | Container allows both directions | Methods (unsound) |

### Covariance (output/producer position)

Return types are covariant:

```typescript
type Animal = { name: string };
type Dog = Animal & { breed: string };

type AnimalFactory = () => Animal;
type DogFactory = () => Dog;

// Dog factory can substitute for Animal factory?
const makeDog: DogFactory = () => ({ name: 'Rex', breed: 'Lab' });
const makeAnimal: AnimalFactory = makeDog;  // ✅ Covariant: Dog → Animal
```

**Why safe:** If you expect an `Animal` back, getting a `Dog` is fine (it has everything `Animal` has, plus more).

### Contravariance (input/consumer position)

Parameter types are contravariant (with `--strictFunctionTypes`):

```typescript
type AnimalHandler = (a: Animal) => void;
type DogHandler = (d: Dog) => void;

const handleAnimal: AnimalHandler = (a) => console.log(a.name);
const handleDog: DogHandler = handleAnimal;  // ✅ Contravariant: Animal → Dog

// But NOT the reverse:
const handleDog2: DogHandler = (d) => console.log(d.breed);
const handleAnimal2: AnimalHandler = handleDog2;  // ❌ Would crash if called with a Cat
```

**Why safe:** If a function accepts `Animal`, it only uses `Animal` properties. So it can safely handle any `Dog` (which has all `Animal` properties). The reverse is not true.

### Invariance

When a type parameter appears in both input and output positions, it becomes invariant:

```typescript
interface MutableBox<T> {
  get(): T;       // Output position → covariant
  set(val: T): void;  // Input position → contravariant
  // Both → invariant: neither direction allowed
}

declare let animalBox: MutableBox<Animal>;
declare let dogBox: MutableBox<Dog>;

animalBox = dogBox;  // ❌ Not safe — could set a Cat via animalBox.set()
dogBox = animalBox;  // ❌ Not safe — could get a non-Dog via dogBox.get()
```

### Explicit variance annotations (TypeScript 4.7+)

You can mark variance explicitly for better performance and documentation:

```typescript
type Producer<out T> = () => T;        // Covariant
type Consumer<in T> = (val: T) => void; // Contravariant
type Processor<in out T> = {            // Invariant
  process(val: T): T;
};
```

These annotations are **checked by the compiler** — if you mark `out` but use T in input position, it errors.

---

## Widening: How TypeScript Forgets Specificity

Widening is the process by which TypeScript **broadens** a literal type to its base type:

```typescript
// let → widened
let x = 'hello';   // type: string  (widened from 'hello')
let y = 42;         // type: number  (widened from 42)
let z = true;       // type: boolean (widened from true)

// const → not widened
const a = 'hello';  // type: 'hello' (literal)
const b = 42;       // type: 42
const c = true;     // type: true
```

### Why widening exists

`let` can be reassigned, so TypeScript assumes you'll assign other values of the same base type. `const` can't be reassigned, so the literal type is safe to keep.

### Object widening

```typescript
// Object properties are widened even with const
const config = {
  port: 3000,
  host: 'localhost',
};
// type: { port: number; host: string }  ← literals are lost!

// To preserve literals:
const config2 = {
  port: 3000,
  host: 'localhost',
} as const;
// type: { readonly port: 3000; readonly host: 'localhost' }
```

### Array widening

```typescript
const arr = [1, 2, 3];        // type: number[]
const arr2 = [1, 2, 3] as const;  // type: readonly [1, 2, 3]

const mixed = ['hello', 42];  // type: (string | number)[]
const mixed2 = ['hello', 42] as const;  // type: readonly ['hello', 42]
```

### Widening in function returns

```typescript
function getStatus() {
  return 'active';  // Return type: string (widened!)
}

function getStatus2() {
  return 'active' as const;  // Return type: 'active'
}

// Or annotate the return type
function getStatus3(): 'active' | 'inactive' {
  return 'active';  // Return type: 'active' | 'inactive'
}
```

**At scale:** Widening is the #1 cause of "inference loss" (a value has a more specific type than TypeScript tracks). Level 02 covers this in depth.

---

## Narrowing: How TypeScript Recovers Specificity

Narrowing is the inverse of widening — TypeScript tracks control flow to make types more specific:

```typescript
function process(input: string | number) {
  // Here: input is string | number

  if (typeof input === 'string') {
    // Here: input is string (narrowed)
    input.toUpperCase();
  } else {
    // Here: input is number (narrowed by elimination)
    input.toFixed(2);
  }
}
```

### Narrowing mechanisms

| Mechanism | Example |
|-----------|---------|
| `typeof` | `typeof x === 'string'` |
| `instanceof` | `x instanceof Date` |
| `in` operator | `'name' in x` |
| Truthiness | `if (x)` (removes `null`, `undefined`, `0`, `''`, etc.) |
| Equality | `x === null`, `x !== undefined` |
| Discriminated unions | `if (result.type === 'success')` |
| Type predicates | `function isString(x: unknown): x is string` |
| `asserts` | `function assert(x: unknown): asserts x is string` |

### Narrowing failure: Closures (pre-5.4)

```typescript
function example(x: string | number) {
  if (typeof x === 'string') {
    // x is string here
    setTimeout(() => {
      // Pre-5.4: x is string | number again (narrowing lost in closure)
      // 5.4+: x is string (narrowing preserved if x is not reassigned)
      x.toUpperCase();
    }, 100);
  }
}
```

### Narrowing failure: Index signatures

```typescript
const map: Record<string, string> = { a: 'hello' };

if (map['a']) {
  // map['a'] is string (narrowed from string | undefined with noUncheckedIndexedAccess)
  // BUT: accessing map['a'] again doesn't benefit from the narrowing
  const val = map['a'];  // still string | undefined without a variable
}

// Fix: bind to a variable
const val = map['a'];
if (val) {
  val.toUpperCase();  // ✅ val is string
}
```

---

## Exercises

1. Determine the variance of `T` in each position:

```typescript
type A<T> = { produce: () => T; consume: (val: T) => void };
type B<T> = { items: T[] };
type C<T> = { callback: (fn: (val: T) => void) => void };
```

2. Predict the widened types:

```typescript
const a = [{ status: 'active' }, { status: 'inactive' }];
const b = [{ status: 'active' } as const, { status: 'inactive' }];
const c = [{ status: 'active' }, { status: 'inactive' }] as const;

function getConfig(env: string) {
  if (env === 'prod') return { port: 443, secure: true };
  return { port: 80, secure: false };
}
```

---

## Next

→ [Level 02: Inference Mastery](../02-inference/)
