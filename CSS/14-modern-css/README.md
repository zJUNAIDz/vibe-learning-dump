# Module 14 — Modern CSS Features

## Overview

CSS is evolving rapidly. This module covers features that have reached baseline browser support (or near-baseline) and fundamentally change how you write CSS. These aren't experimental — they're production-ready tools that eliminate workarounds and solve real problems.

```mermaid
graph TD
    A[Modern CSS] --> B[Custom Properties]
    A --> C[Cascade Layers]
    A --> D[New Selectors]
    A --> E[Advanced Functions]
    
    B --> B1[Theming]
    B --> B2[Dynamic values]
    C --> C1[Third-party CSS control]
    C --> C2[Specificity management]
    D --> D1[:has parent selector]
    D --> D2[:is / :where]
    E --> E1[color-mix / oklch]
    E --> E2[@property typed variables]
    
    style A fill:#f0f0ff,stroke:#333
```

## Lessons

| # | File | Topic |
|---|------|-------|
| 01 | [01-custom-properties-deep.md](01-custom-properties-deep.md) | Custom properties — advanced patterns, animation, @property |
| 02 | [02-cascade-layers.md](02-cascade-layers.md) | @layer — controlling cascade order explicitly |
| 03 | [03-selectors.md](03-selectors.md) | :has(), :is(), :where(), :not() — modern selector combinators |
| 04 | [04-color-and-functions.md](04-color-and-functions.md) | color-mix(), oklch, trigonometric functions, new units |

## Prerequisites

- Completed Module 02 (Cascade) — understanding origins, specificity
- Completed Module 12 (Architecture) — knowing why these features matter

## Next

→ [Lesson 01: Custom Properties Deep Dive](01-custom-properties-deep.md)
