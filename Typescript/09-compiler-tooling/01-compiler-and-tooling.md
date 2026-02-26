# Level 09: Compiler & Tooling Knowledge

## The Problem

Most developers copy-paste `tsconfig.json` from starter templates without understanding what each flag does. Then they hit issues:

- "Why is `any` leaking into my codebase?"
- "My build takes 45 seconds"
- "The types work in my editor but break in CI"
- "I enabled `strict` but still get runtime crashes"

This level covers the flags that actually matter, performance, and the compiler behaviors that affect daily work.

---

## `strict` and Its Sub-Flags

`"strict": true` is not one flag — it enables **multiple** sub-flags. Understanding each one matters because some projects enable `strict` but then selectively disable sub-flags, creating false confidence.

### What `strict: true` enables

| Flag | What it does | Why it matters |
|------|-------------|----------------|
| `strictNullChecks` | `null` and `undefined` are distinct types | Without this, EVERY type includes `null`. `string` = `string \| null \| undefined`. |
| `strictFunctionTypes` | Function parameters are contravariant | Without this, function params are bivariant (unsound) |
| `strictBindCallApply` | Type-check `bind`, `call`, `apply` | Without this, `fn.call(null, wrongArg)` doesn't error |
| `strictPropertyInitialization` | Class properties must be initialized | Catches `this.value` being `undefined` before assignment |
| `noImplicitAny` | Error on inferred `any` | Without this, unannotated params are silently `any` |
| `noImplicitThis` | Error on `this` with unclear type | Catches `this` being `any` in standalone functions |
| `useUnknownInCatchVariables` | `catch(e)` types `e` as `unknown`, not `any` | Forces you to check error types before using them |
| `alwaysStrict` | Emits `"use strict"` in every file | JavaScript strict mode |

### The flags `strict` does NOT enable

These are separate and arguably more important than some `strict` sub-flags:

```json
{
  "exactOptionalPropertyTypes": true,
  "noUncheckedIndexedAccess": true,
  "noPropertyAccessFromIndexSignature": true,
  "noFallthroughCasesInSwitch": true
}
```

### `exactOptionalPropertyTypes`

```typescript
type User = {
  name: string;
  nickname?: string;  // Optional
};

// Without exactOptionalPropertyTypes:
const user: User = { name: 'Alice', nickname: undefined };  // ✅ allowed

// With exactOptionalPropertyTypes:
const user: User = { name: 'Alice', nickname: undefined };  // ❌ error!
// undefined is NOT the same as "not present"
```

**Why this matters:** `Object.keys(user)` includes `'nickname'` when it's explicitly `undefined`, but not when it's absent. APIs, databases, and serialization treat these differently. This flag enforces the distinction.

### `noUncheckedIndexedAccess`

```typescript
const map: Record<string, string> = { a: 'hello' };

// Without noUncheckedIndexedAccess:
const val = map['nonexistent'];  // type: string ← lie!

// With noUncheckedIndexedAccess:
const val = map['nonexistent'];  // type: string | undefined ← honest

// Also applies to arrays:
const arr = [1, 2, 3];
const item = arr[10];  // type: number | undefined (with flag)
```

**This is arguably the most important flag not in `strict`.** Without it, every record/dict/array access is a potential runtime crash that TypeScript claims is safe.

### `noPropertyAccessFromIndexSignature`

```typescript
type Config = {
  port: number;
  [key: string]: unknown;
};

const config: Config = { port: 3000 };

// Without flag:
config.whatever;  // type: unknown ✅ (but autocomplete suggests it exists)

// With flag:
config.whatever;  // ❌ Must use config['whatever']
config.port;      // ✅ Known properties still work with dot
```

Forces bracket notation for index signature access, making it visually obvious when you're accessing a potentially-nonexistent property.

---

## `tsconfig` Flags That Affect Type Behavior

### `moduleResolution`

| Value | Behavior | When to use |
|-------|----------|-------------|
| `node` | Node.js CJS resolution | Legacy Node.js CJS projects |
| `node16` / `nodenext` | Node.js ESM + CJS resolution | Modern Node.js projects |
| `bundler` | Like `node16` but no `.js` extension requirement | Vite, webpack, esbuild projects |

**Common mistake:** Using `"moduleResolution": "node"` with ESM. This causes `.d.ts` files not to be found for packages that use `exports` in `package.json`.

### `verbatimModuleSyntax`

```typescript
// Without verbatimModuleSyntax:
import { User } from './types';  // TypeScript might elide this import if User is type-only
// Problem: if types.ts has side effects, they're lost

// With verbatimModuleSyntax:
import type { User } from './types';  // Must be explicit
import { userSchema } from './types'; // This stays because it's a value import
```

Forces you to use `import type` for type-only imports. Prevents accidental side-effect loss and makes the intent clear.

### `isolatedModules`

Requires that every file can be independently compiled (as esbuild/swc/babel do):

```typescript
// ❌ Not allowed with isolatedModules
const enum Direction { Up, Down }  // const enums need cross-file info

// ❌ Not allowed
export { SomeType } from './types';  // Must re-export with `export type`

// ✅ Required
export type { SomeType } from './types';
```

**Always enable this** if using esbuild, swc, babel, or any non-tsc compiler.

### `skipLibCheck`

Skips type checking of `.d.ts` files (including `node_modules` type definitions).

```json
{ "skipLibCheck": true }
```

**Always enable this.** Reasons:
1. Library `.d.ts` files may use different TypeScript versions
2. Checking them slows compilation significantly
3. You can't fix errors in `node_modules` anyway
4. Library authors should test their own types

---

## Why TypeScript Gets Slow

### Common performance problems

| Problem | Impact | Fix |
|---------|--------|-----|
| **Large union types** (100+ members) | Type comparison is O(n²) | Split into smaller unions or use maps |
| **Deep conditional type recursion** | Exponential blowup | Add tail-call optimization, limit depth |
| **Complex mapped types on large objects** | Slow property resolution | Simplify or cache intermediate types |
| **`include` too broad** | Compiles files you don't need | Narrow glob patterns |
| **No `skipLibCheck`** | Checks all `.d.ts` files | Enable `skipLibCheck` |
| **Project references absent** | Monorepo compiles everything | Use `references` for incremental builds |
| **Circular type references** | Can cause infinite expansion | Break cycles with interfaces |

### Measuring TypeScript performance

```bash
# Generate a trace
npx tsc --generateTrace ./trace-output

# Analyze with TypeScript's trace analyzer
# https://github.com/nicolo-ribaudo/tc39-proposal-type-annotations

# Quick check: how long does type checking take?
time npx tsc --noEmit

# See which files are slow
npx tsc --extendedDiagnostics --noEmit
```

### The `--extendedDiagnostics` output

```
Files:           1547
Lines:           189432
Identifiers:    67891
Symbols:        45123
Types:          23456
Instantiations: 12345
Memory used:    287654K
Parse time:     1.23s
Bind time:      0.45s
Check time:     8.92s    ← This is where slowness usually is
Emit time:      0.05s
Total time:     10.65s
```

`Check time` dominates. Reduce it by:
- Fewer type instantiations (simpler generics)
- Project references for incremental builds
- `skipLibCheck: true`

### Type instantiation limits

TypeScript has a built-in limit of ~50 levels of type recursion:

```typescript
// ❌ Will hit recursion limit on deep objects
type DeepPartial<T> = T extends object
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : T;

// For objects nested 50+ levels deep, TypeScript gives up
```

The instantiation limit (2,500,000 by default) catches runaway type computations:

```typescript
// ❌ Causes "Type instantiation is excessively deep and possibly infinite"
type HugeUnion = '0' | '1' | '2' | ... | '999';
type CartesianProduct = `${HugeUnion}-${HugeUnion}`;
// 1,000,000 type members — too many
```

---

## Project References (Monorepo Performance)

### The problem

In a monorepo, `tsc` compiles everything on every change:

```
packages/
  shared/     ← unchanged
  backend/    ← unchanged
  frontend/   ← changed one file → entire monorepo recompiled
```

### Project references

```json
// packages/frontend/tsconfig.json
{
  "compilerOptions": { /* ... */ },
  "references": [
    { "path": "../shared" }
  ]
}

// packages/shared/tsconfig.json
{
  "compilerOptions": {
    "composite": true,        // Required for referenced projects
    "declaration": true,      // Required for consumers
    "declarationMap": true    // Optional: enables "go to source"
  }
}
```

```bash
# Build with references
npx tsc --build

# Only recompiles changed projects
npx tsc --build --watch
```

**Impact:** 10-100x faster incremental builds in large monorepos.

---

## `declaration` and `.d.ts` Files

### What `declaration: true` produces

```typescript
// src/utils.ts
export function add(a: number, b: number): number {
  return a + b;
}

export type Result<T> = { ok: true; value: T } | { ok: false; error: Error };
```

```typescript
// dist/utils.d.ts (generated)
export declare function add(a: number, b: number): number;
export type Result<T> = { ok: true; value: T } | { ok: false; error: Error };
```

### `declarationMap`

Enables "Go to Definition" to jump to the `.ts` source instead of the `.d.ts` file. Enable this in libraries during development.

### `emitDeclarationOnly`

When using esbuild/swc for JavaScript output:

```json
{
  "compilerOptions": {
    "emitDeclarationOnly": true,  // tsc only generates .d.ts
    "declaration": true
  }
}
```

Your build pipeline: esbuild/swc compiles TS→JS. tsc generates only `.d.ts` files. Fastest of both worlds.

---

## Practical `tsconfig.json` Templates

### Application (Next.js / Node.js)

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "strict": true,
    "exactOptionalPropertyTypes": true,
    "noUncheckedIndexedAccess": true,
    "noPropertyAccessFromIndexSignature": true,
    "verbatimModuleSyntax": true,
    "isolatedModules": true,
    "skipLibCheck": true,
    "esModuleInterop": true,
    "resolveJsonModule": true,
    "outDir": "./dist",
    "rootDir": "./src"
  },
  "include": ["src"]
}
```

### Library (npm package)

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "strict": true,
    "exactOptionalPropertyTypes": true,
    "noUncheckedIndexedAccess": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "skipLibCheck": true,
    "isolatedModules": true
  },
  "include": ["src"],
  "exclude": ["**/*.test.ts"]
}
```

---

## Exercises

1. **Audit your `tsconfig.json`:** Check which of these flags you have enabled. Add `noUncheckedIndexedAccess` and `exactOptionalPropertyTypes` to an existing project. Fix the resulting errors — each one is a real bug that was hidden.

2. **Measure your build:** Run `tsc --extendedDiagnostics --noEmit` on your project. What's your check time? How many type instantiations? If over 1M, investigate which types are expensive.

3. **Set up project references:** If you work in a monorepo, add `composite: true` and `references` to your `tsconfig` files. Measure the incremental build time difference.

---

## Next

→ [Level 10: Debugging the Type System](../10-debugging-types/)
