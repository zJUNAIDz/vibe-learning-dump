# Start Here: Quick Orientation

**This curriculum is for developers who already write TypeScript daily** and want to go from "I use it" to "I understand it."

---

## What This Curriculum IS

- A deep dive into TypeScript's **type system as a programming language**
- Mental models that explain *why* the compiler behaves the way it does
- The patterns that power Zod, tRPC, Prisma, React Query, and similar libraries
- Practitioner-grade content for senior engineers

## What This Curriculum IS NOT

- A TypeScript tutorial
- Coverage of basics (unions, generics syntax, interfaces)
- Framework-specific guidance (React, Next.js)
- A cookbook of "useful types"

---

## Prerequisites

You must already be comfortable with:

- ✅ Union and intersection types
- ✅ Generic functions and types
- ✅ `keyof`, `typeof`, index signatures
- ✅ Type narrowing and type guards
- ✅ Module systems (ESM, CJS)
- ✅ TypeScript with at least one framework (React, Express, Next.js)
- ✅ `tsconfig.json` basics

**Self-test:** If you can explain what `Record<string, unknown>` does and why it's different from `object`, you're ready.

---

## How to Use This

### The Levels

Each level is a folder (`00-orientation/`, `01-type-system/`, etc.) containing markdown files that build on each other.

**Read them in order.** Each level assumes mastery of the previous one.

### The Code Examples

Every code example is designed to be pasted into a TypeScript playground or `.ts` file. Experiment. Break things. The type system is a REPL — use it like one.

**Recommended:** Open the [TypeScript Playground](https://www.typescriptlang.org/play) alongside this curriculum. Paste examples. Modify them. See what the compiler says.

### The Exercises

Where exercises exist, they are not optional. They are designed to expose gaps in understanding that reading alone won't fill.

---

## Curriculum Map

| Level | Topic | You'll Learn |
|-------|-------|------|
| 00 | [Orientation](00-orientation/) | TypeScript's actual architecture, the two languages |
| 01 | [Type System Foundations](01-type-system/) | Structural typing, assignability, soundness holes |
| 02 | [Inference Mastery](02-inference/) | Bidirectional inference, contextual typing, `as const` |
| 03 | [Advanced Generics](03-advanced-generics/) | Constraints, inference from arguments, API design |
| 04 | [Conditional Types](04-conditional-types/) | Distribution, `infer`, type-level branching |
| 05 | [Mapped Types](05-mapped-types/) | `keyof` pitfalls, template literals, deep transforms |
| 06 | [Type-Level Programming](06-type-level-programming/) | Branded types, nominal hacks, exhaustiveness |
| 07 | [Runtime ↔ Type Bridge](07-runtime-type-bridge/) | Zod mental model, schema-first design |
| 08 | [Library Author Patterns](08-library-patterns/) | Ergonomic APIs, overloads, intentional `any` |
| 09 | [Compiler & Tooling](09-compiler-tooling/) | `tsconfig` flags, `strict` sub-flags, performance |
| 10 | [Debugging Types](10-debugging-types/) | Mental printing, introspection, refactoring impossible types |
| 11 | [Capstone](11-capstone/) | Projects, exercises, wizard checklist |

---

## Start Now

→ [00-orientation/](00-orientation/)
