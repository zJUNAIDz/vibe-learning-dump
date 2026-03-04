# Lesson 03 — Native CSS vs Preprocessors

## The Convergence

Modern CSS has adopted many features that previously required preprocessors. This lesson maps each preprocessor feature to its native equivalent.

## Variables

### Sass Variables (Compile-Time)

```scss
$primary: #2563eb;
$spacing: 8px;

.button {
  background: $primary;
  padding: $spacing * 2;          // Math at compile time
}
```

**Output:** Static values. No runtime behavior.

### CSS Custom Properties (Runtime)

```css
:root {
  --primary: #2563eb;
  --spacing: 8px;
}

.button {
  background: var(--primary);
  padding: calc(var(--spacing) * 2);   /* Math at runtime */
}

/* Dynamically overridden at any level */
.dark-theme {
  --primary: #60a5fa;
}

/* Changed via JavaScript */
/* document.documentElement.style.setProperty('--primary', '#ff0000'); */
```

### Key Differences

| Aspect | Sass `$var` | CSS `var()` |
|--------|-------------|-------------|
| When resolved | Build time | Runtime |
| Cascade-aware | ❌ | ✅ |
| Changeable at runtime | ❌ | ✅ (JS or class toggle) |
| Scoped to selectors | ❌ (lexical scope) | ✅ (cascade scope) |
| Fallback values | ❌ | ✅ `var(--x, fallback)` |
| Media query values | ❌ | ❌ (can't use in MQ conditions) |
| Math | `$a * 2` | `calc(var(--a) * 2)` |

**Recommendation:** Use CSS custom properties for theming, dynamic values, and anything that changes. Use Sass variables only for static build-time constants (breakpoints, config).

## Nesting

### Sass Nesting

```scss
.card {
  border: 1px solid #ddd;

  .title {
    font-size: 18px;
  }

  &:hover {
    border-color: blue;
  }

  @media (min-width: 768px) {
    padding: 24px;
  }
}
```

### Native CSS Nesting (Baseline 2023)

```css
.card {
  border: 1px solid #ddd;

  .title {         /* Equivalent to .card .title */
    font-size: 18px;
  }

  &:hover {        /* .card:hover */
    border-color: blue;
  }

  @media (min-width: 768px) {
    padding: 24px;
  }
}
```

Native nesting works in all modern browsers (Chrome 120+, Firefox 117+, Safari 17.2+).

**Difference:** In native CSS, nested selectors that start with a letter require `&`:

```css
.card {
  /* ✅ Works — starts with . */
  .title { }

  /* ✅ Works — uses & */
  & h3 { }

  /* ❌ Error — bare element without & */
  /* h3 { } */
}
```

## Color Functions

### Sass

```scss
$base: #2563eb;
.button {
  background: $base;
  &:hover { background: darken($base, 10%); }
  &:active { background: darken($base, 20%); }
}
.badge {
  background: rgba($base, 0.1);
}
```

### Native CSS

```css
:root { --base: #2563eb; }

.button {
  background: var(--base);
}
.button:hover {
  background: color-mix(in srgb, var(--base), black 20%);
}
.button:active {
  background: color-mix(in srgb, var(--base), black 35%);
}
.badge {
  background: color-mix(in srgb, var(--base) 10%, transparent);
}
```

`color-mix()` is supported in all modern browsers (Baseline 2023).

## Conditionals

### Sass `@if`

```scss
@mixin button-variant($type) {
  @if $type == 'primary' {
    background: blue;
    color: white;
  } @else if $type == 'danger' {
    background: red;
    color: white;
  } @else {
    background: gray;
    color: black;
  }
}
```

### Native CSS — No Direct Equivalent

Use custom properties + selectors instead:

```css
.button {
  --bg: gray;
  --color: black;
  background: var(--bg);
  color: var(--color);
}

.button[data-variant="primary"] {
  --bg: blue;
  --color: white;
}

.button[data-variant="danger"] {
  --bg: red;
  --color: white;
}
```

## Loops / Generation

### Sass

```scss
$sizes: (sm: 14px, md: 16px, lg: 20px, xl: 24px);

@each $name, $size in $sizes {
  .text-#{$name} {
    font-size: $size;
  }
}
```

### Native CSS — No Equivalent

This is where Sass still wins. Native CSS has no loops or generation. For utility class generation, you still need:
- Sass `@each` / `@for`
- PostCSS plugin
- Build-time code generation

## Feature Migration Guide

```mermaid
graph TD
    A[Sass Feature] --> B{Native CSS<br>equivalent?}
    B -->|Yes| C[Migrate]
    B -->|No| D[Keep Sass]
    
    C --> C1[Variables → Custom Properties]
    C --> C2[Nesting → Native Nesting]
    C --> C3[darken/lighten → color-mix]
    C --> C4[@import → @layer + @import]
    
    D --> D1[Loops @each @for]
    D --> D2[Mixins with logic]
    D --> D3[Maps / data structures]
    D --> D4[String interpolation]
    
    style C fill:#efe,stroke:#0a0
    style D fill:#fff3e0,stroke:#e65100
```

## Summary Table

| Feature | Sass | Native CSS | Status |
|---------|------|-----------|--------|
| Variables | `$var` | `var(--prop)` | ✅ Native is superior (runtime) |
| Nesting | Built-in | Native nesting | ✅ Feature parity |
| Color manipulation | `darken()`, `lighten()` | `color-mix()` | ✅ Native available |
| `@import` organization | `@import` partial | `@layer` | ✅ Native (different model) |
| Mixins | `@mixin` | ❌ | Sass only |
| `@extend` | `%placeholder` | ❌ | Sass only (avoid anyway) |
| Loops | `@each`, `@for` | ❌ | Sass only |
| Conditionals | `@if` / `@else` | ❌ | Sass only |
| Functions | `@function` | ❌ | Sass only |
| Maps | `$map: (k: v)` | ❌ | Sass only |

## Practical Migration Strategy

1. **Stop using** `@import` → switch to `@use` (Sass) or native `@layer`
2. **Replace** Sass variables with CSS custom properties for any value that's used in the cascade or needs to change
3. **Keep** `$breakpoint` variables in Sass (can't use custom properties in media query conditions)
4. **Replace** `darken()` / `lighten()` with `color-mix()`
5. **Adopt** native nesting, drop Sass nesting
6. **Keep** Sass for loops, complex mixins, and utility generation
7. **Goal:** Sass handles generation; CSS handles runtime behavior

## Next

→ [Lesson 04: Build Pipelines](04-build-pipelines.md)
