# Advanced TypeScript Curriculum

> **Deep, practitioner-grade curriculum** for experienced developers who use TypeScript daily but want to *think* in it.

## Overview

This curriculum treats TypeScript's type system as a **programming language in its own right**. It is not about syntax or features — it's about the mental models that separate people who *use* TypeScript from people who *understand* it.

**This is NOT for beginners.** There is zero coverage of `string | number`, `interface` vs `type`, or how generics work at a surface level. If you need that, read the handbook first.

---

## Who This Is For

✅ **You should take this if you:**
- Have 2+ years of daily TypeScript usage
- Work with React, Node.js, Next.js, backend APIs
- Look at Zod/tRPC/Prisma types and wonder "how did they do that?"
- Hit compiler errors you can't explain
- Want to write library-quality types
- Want to stop fighting the compiler and start leveraging it

❌ **Skip this if you:**
- Are learning TypeScript for the first time
- Need help with basic generics or union types
- Want framework tutorials (React, Next.js)

---

## Philosophy

### This Curriculum Believes:

1. **TypeScript is two languages** — a runtime language (JavaScript) and a compile-time language (the type system). You must learn both separately.
2. **Inference is the real feature** — Good TypeScript code barely has type annotations. If you're annotating everything, you're doing it wrong.
3. **The type system is Turing-complete** — You can write loops, conditionals, recursion, and pattern matching at the type level. This is not trivia — it's how every serious library works.
4. **Soundness is a spectrum** — TypeScript intentionally has unsound escape hatches. Knowing where and why they exist is essential.
5. **The compiler is your pair programmer** — but only if you know how to speak its language.

---

## Curriculum Structure

### Learning Path

```
┌─────────────────────────────────────────────────────────────┐
│  Level 00: Orientation                                       │
│  TypeScript's actual architecture. The two languages.        │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 01: The Type System (Not Syntax Sugar)                │
│  Structural typing, assignability, soundness holes           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 02: Inference Mastery                                 │
│  Bidirectional inference, contextual typing, as const        │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 03: Advanced Generics                                 │
│  Constraints, inference from args, partial inference         │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 04: Conditional Types                                 │
│  Distribution, infer, type-level branching                   │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 05: Mapped Types & Key Remapping                      │
│  keyof pitfalls, template literals, deep transforms          │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 06: Type-Level Programming                            │
│  Branded types, nominal hacks, exhaustiveness, invariants    │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 07: Runtime ↔ Type System Bridge                      │
│  Zod/Valibot mental model, schema-first vs type-first        │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 08: Library Author Patterns                           │
│  Ergonomic APIs, hiding complexity, overloads vs generics    │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 09: Compiler & Tooling Knowledge                      │
│  tsconfig that matters, strict flags, performance            │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 10: Debugging the Type System                         │
│  Mental printing, introspection helpers, refactoring types   │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Level 11: Capstone & Wizard Checklist                       │
│  Projects, exercises, mastery verification                   │
└─────────────────────────────────────────────────────────────┘
```

---

## How Each Level Works

Every topic follows this structure:

| Section | Purpose |
|---------|---------|
| **The Problem** | What real-world gap this feature fills |
| **Mental Model** | How to think about it, not just the syntax |
| **Real-World Use** | Where this appears in APIs, DTOs, libraries |
| **Failure Modes** | Common wrong assumptions and why they break |
| **Code Examples** | Small but non-trivial, copy-paste-and-experiment |
| **Why It Matters at Scale** | DX, refactors, safety implications |

---

## Estimated Time

| Level | Topic | Hours |
|-------|-------|-------|
| 00 | Orientation | 1-2 |
| 01 | Type System Foundations | 3-4 |
| 02 | Inference Mastery | 3-4 |
| 03 | Advanced Generics | 4-5 |
| 04 | Conditional Types | 4-5 |
| 05 | Mapped Types & Key Remapping | 3-4 |
| 06 | Type-Level Programming | 4-5 |
| 07 | Runtime ↔ Type Bridge | 3-4 |
| 08 | Library Author Patterns | 3-4 |
| 09 | Compiler & Tooling | 2-3 |
| 10 | Debugging the Type System | 3-4 |
| 11 | Capstone & Wizard Checklist | 5-8 |
| **Total** | | **~40-50 hours** |

---

## Quick Start

1. Read [START_HERE.md](START_HERE.md)
2. Check [GETTING_STARTED.md](GETTING_STARTED.md) for environment setup
3. Begin with [00-orientation/](00-orientation/)
4. Use [QUICK_REFERENCE.md](QUICK_REFERENCE.md) for lookup during exercises

---

## The Goal

The goal is not "knowing TypeScript features."

**The goal is thinking in TypeScript.**

When you complete this curriculum, you should be able to:

- Read any library's `.d.ts` files and understand the design decisions
- Build type-safe APIs that guide consumers with zero documentation
- Debug compiler errors by reasoning about the type system, not guessing
- Write types that make invalid states unrepresentable
- Understand *why* TypeScript behaves the way it does, including when it lies
