# Level 00: Orientation — TypeScript's Actual Architecture

## The Two Languages

TypeScript is not "JavaScript with types." It is **two separate languages** sharing one syntax:

1. **The runtime language** — JavaScript. This is what executes. TypeScript adds zero runtime behavior (with minor exceptions like `enum` and `namespace`).

2. **The compile-time language** — The type system. This is a separate programming language that runs during compilation, produces no output, and is erased entirely before execution.

This distinction matters because the type-level language has its own:
- **Variables** — generics (`T`, `U`, `K`)
- **Conditionals** — `T extends U ? X : Y`
- **Loops** — mapped types (`{ [K in keyof T]: ... }`)
- **Functions** — generic type aliases and generic functions
- **Recursion** — recursive conditional types
- **Pattern matching** — `infer` keyword

Most TypeScript developers learn the runtime language and treat types as annotations. This curriculum teaches you the compile-time language as a first-class skill.

---

## How the Compiler Actually Works

### Compilation Pipeline

```
Source code (.ts)
       ↓
   [Parser]          → AST (Abstract Syntax Tree)
       ↓
   [Binder]          → Symbol table (connects declarations to usages)
       ↓
   [Type Checker]    → Type resolution, inference, error reporting
       ↓
   [Emitter]         → JavaScript output (types completely erased)
```

The **type checker** is where 90% of your daily interaction happens. Understanding its behavior is the purpose of this curriculum.

### What "Type Erasure" Really Means

```typescript
// What you write
function greet(name: string): string {
  return `Hello, ${name}`;
}

// What executes (types are gone)
function greet(name) {
  return `Hello, ${name}`;
}
```

Implications that trip up experienced developers:

1. **You cannot branch on types at runtime.** `if (T extends string)` is not a thing. The type system and runtime are completely separate evaluation phases.

2. **Generic type parameters don't exist at runtime.** You cannot write `new T()` or `typeof T` where `T` is a generic parameter.

3. **Type assertions (`as`) do nothing at runtime.** `value as string` doesn't convert, cast, or validate. It just tells the compiler to shut up.

4. **`interface` and `type` produce zero JavaScript.** They are compile-time-only constructs.

---

## The Type System's Evaluation Model

### Types Resolve Eagerly (Mostly)

When you write:

```typescript
type A = string | number;
type B = A extends string ? 'yes' : 'no';
```

`B` is resolved **immediately** to `'no'`. There is no lazy evaluation here.

But when generics are involved:

```typescript
type B<T> = T extends string ? 'yes' : 'no';
```

`B` is **deferred** — it cannot resolve until `T` is provided. This is the fundamental mechanism behind conditional types, and it's why generic types feel like functions.

### Types Are Sets

The single most important mental model:

**Every type in TypeScript is a set of possible values.**

| Type | Set |
|------|-----|
| `never` | Empty set ∅ |
| `'hello'` | { "hello" } |
| `string` | All strings |
| `string \| number` | All strings ∪ all numbers |
| `string & number` | All strings ∩ all numbers = ∅ = `never` |
| `unknown` | Universal set (everything) |
| `any` | Magic — both universal and empty simultaneously |

Union (`|`) is set union. Intersection (`&`) is set intersection. `extends` is the subset check (⊆).

```typescript
// "Is A a subset of B?"
type Check = 'hello' extends string ? true : false;  // true
// Because { "hello" } ⊆ { all strings }
```

This model explains most "surprising" compiler behavior:

```typescript
// Why does this work?
type T1 = string & ('hello' | 'world');  // 'hello' | 'world'
// Because { all strings } ∩ { "hello", "world" } = { "hello", "world" }

// Why is this never?
type T2 = string & number;  // never
// Because { all strings } ∩ { all numbers } = ∅
```

---

## `any` vs `unknown` vs `never` — The Real Story

These three types occupy special positions in the type universe:

### `unknown` — the top type

Every type is assignable to `unknown`. It is the universal set.

```typescript
const a: unknown = 'hello';  // ✅
const b: unknown = 42;       // ✅
const c: unknown = null;     // ✅

// But you can't USE it without narrowing
const d: string = a;         // ❌ Type 'unknown' is not assignable to type 'string'
```

**Use case:** Function parameters where you accept anything but force the consumer to validate.

### `never` — the bottom type

No value is assignable to `never`. It is the empty set.

```typescript
const a: never = 'hello';  // ❌
const b: never = 42;       // ❌

// But never is assignable to EVERYTHING
function crash(): never { throw new Error(); }
const c: string = crash();  // ✅ (function never returns)
```

**Use case:** Representing impossible states, exhaustiveness checking, and as the identity element for unions (`T | never = T`).

### `any` — the escape hatch

`any` breaks the type system's rules. It is simultaneously assignable to everything AND everything is assignable to it:

```typescript
const a: any = 'hello';     // ✅ (everything → any)
const b: string = a;        // ✅ (any → everything)
const c: number = a;        // ✅ (any → everything)
```

This is **logically incoherent** — it says `any` is both the universal set and the empty set. TypeScript does this deliberately because `any` is a pragmatic escape hatch, not a logical type.

**Key insight:** `any` is not "I don't know the type." That's `unknown`. `any` is "I want the compiler to stop checking this."

---

## Why This Matters

Everything in this curriculum builds on three ideas:

1. **Types are sets** — union is ∪, intersection is ∩, extends is ⊆
2. **The type system is a separate language** — with its own variables, conditionals, loops, functions
3. **Erasure is total** — types produce zero runtime behavior

If you internalize these, the rest of the curriculum is connecting patterns to this foundation. If you skip this, everything else will feel like memorizing syntax.

---

## Next

→ [Level 01: The Type System (Not Syntax Sugar)](../01-type-system/)
