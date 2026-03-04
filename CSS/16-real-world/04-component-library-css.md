# Lesson 04 — Component Library CSS

## Building CSS for Reusable Components

When you build a component library (design system, shared UI kit), the CSS architecture decisions affect every consumer. Get them wrong and you create coupling, specificity fights, and maintenance nightmares.

## The Component CSS API

Every component has a **public styling interface**. Define it explicitly:

```css
/* Button component — explicit API via custom properties */
.btn {
  /* === PUBLIC API === */
  --btn-bg: var(--color-primary);
  --btn-color: white;
  --btn-border: none;
  --btn-radius: var(--radius-md);
  --btn-padding-x: var(--spacing-md);
  --btn-padding-y: var(--spacing-sm);
  --btn-font-size: var(--font-size-body);
  --btn-font-weight: var(--font-semibold);

  /* === PRIVATE IMPLEMENTATION === */
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5em;
  background: var(--btn-bg);
  color: var(--btn-color);
  border: var(--btn-border);
  border-radius: var(--btn-radius);
  padding: var(--btn-padding-y) var(--btn-padding-x);
  font-size: var(--btn-font-size);
  font-weight: var(--btn-font-weight);
  cursor: pointer;
  text-decoration: none;
  line-height: 1;
  white-space: nowrap;
  transition: background 0.15s, box-shadow 0.15s;
}

/* Variants override the API, not internals */
.btn--sm {
  --btn-padding-x: var(--spacing-sm);
  --btn-padding-y: var(--spacing-xs);
  --btn-font-size: var(--font-size-body-sm);
}

.btn--lg {
  --btn-padding-x: var(--spacing-lg);
  --btn-padding-y: var(--spacing-md);
  --btn-font-size: var(--text-lg);
}

.btn--danger {
  --btn-bg: var(--color-danger);
}

.btn--ghost {
  --btn-bg: transparent;
  --btn-color: var(--color-primary);
  --btn-border: 1px solid var(--color-border);
}
```

### Benefits of the API Pattern

1. **Consumers customize without touching internals:**
```css
/* App-level override — easy and safe */
.nav .btn {
  --btn-radius: 0;      /* Square buttons in the nav */
  --btn-font-size: 14px;
}
```

2. **Refactoring internals doesn't break consumers:**
```css
/* Library can change from padding to something else internally */
/* API (custom properties) stays the same */
```

3. **Self-documenting** — the custom properties list IS the documentation

## Component Composition

### Slot-Based Components

Components should provide **insertion points**, not be monolithic:

```html
<!-- Flexible card structure — slots for any content -->
<div class="card">
  <div class="card__header">
    <!-- Consumer puts anything here -->
  </div>
  <div class="card__body">
    <!-- Consumer puts anything here -->
  </div>
  <div class="card__footer">
    <!-- Consumer puts anything here -->
  </div>
</div>
```

```css
.card {
  --card-padding: var(--spacing-md);
  --card-radius: var(--radius-lg);
  --card-border: 1px solid var(--color-border);

  background: var(--color-bg-surface);
  border: var(--card-border);
  border-radius: var(--card-radius);
  overflow: hidden;
}

.card__header {
  padding: var(--card-padding);
  border-bottom: var(--card-border);
}

.card__body {
  padding: var(--card-padding);
}

.card__footer {
  padding: var(--card-padding);
  border-top: var(--card-border);
}

/* Layout variants */
.card--horizontal {
  display: grid;
  grid-template-columns: auto 1fr;
  grid-template-rows: 1fr auto;
}

.card--horizontal .card__header {
  grid-row: 1 / -1;
  border-bottom: none;
  border-right: var(--card-border);
}
```

## Avoiding Specificity Conflicts

### Rule: Keep Specificity Flat

```css
/* ❌ LIBRARY CODE — high specificity blocks overrides */
.card .card__title .highlight {
  color: blue;                  /* (0,3,0) — hard to override */
}

/* ✅ LIBRARY CODE — flat specificity */
.card__title-highlight {
  color: blue;                  /* (0,1,0) — easy to override */
}
```

### Rule: Use @layer for Library CSS

```css
/* Library ships its styles in a layer */
@layer components {
  .btn { /* ... */ }
  .card { /* ... */ }
  .modal { /* ... */ }
}

/* Consumer styles are unlayered — automatically win */
.btn { /* Consumer override — always wins over library */ }

/* Or consumer uses their own layers */
@layer vendor, app;
@import 'ui-library.css' layer(vendor);

@layer app {
  .btn { /* Always wins over vendor layer */ }
}
```

## State Management in CSS

### Data Attributes for State

```css
/* Use data attributes for states — cleaner than modifier classes */
.dialog { ... }
.dialog[data-state="open"] {
  display: block;
  animation: fadeIn 0.2s;
}
.dialog[data-state="closed"] {
  display: none;
}

.accordion__panel[data-expanded="true"] {
  grid-template-rows: 1fr;
}
.accordion__panel[data-expanded="false"] {
  grid-template-rows: 0fr;
}
```

### Animation for State Transitions

```css
/* Smooth expand/collapse with grid rows */
.accordion__panel {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows 0.3s ease;
}

.accordion__panel[data-expanded="true"] {
  grid-template-rows: 1fr;
}

.accordion__panel > .inner {
  overflow: hidden;
}
```

## Documentation Pattern

Document every component with:

```markdown
## Button (.btn)

### Custom Properties (API)

| Property | Default | Description |
|----------|---------|-------------|
| --btn-bg | var(--color-primary) | Background color |
| --btn-color | white | Text color |
| --btn-radius | var(--radius-md) | Border radius |
| --btn-padding-x | var(--spacing-md) | Horizontal padding |
| --btn-padding-y | var(--spacing-sm) | Vertical padding |

### Variants

| Class | Effect |
|-------|--------|
| .btn--sm | Smaller padding and font |
| .btn--lg | Larger padding and font |
| .btn--danger | Red background |
| .btn--ghost | Transparent with border |

### States

| Attribute | Values | Effect |
|-----------|--------|--------|
| disabled | (boolean) | Reduced opacity, no interaction |
| data-loading | true/false | Shows spinner, disables |

### Usage

​```html
<button class="btn btn--sm btn--danger">Delete</button>
​```
```

## Architecture Checklist for Component Libraries

```
□ All component styles are in a @layer
□ All selectors are single-class (0,1,0) specificity
□ Customization is via custom properties, not overriding internals
□ Components don't set external layout (no margin on root element)
□ Components use semantic tokens, not primitive values
□ States use data attributes, not complex class toggling
□ Dark mode works automatically (semantic token overrides)
□ Each component has a documented API
□ No !important in component code
□ Reset/normalize is separate from component styles
```

## The Component Layout Rule

**Components should not set their own external spacing.** A component doesn't know where it will be placed:

```css
/* ❌ Component sets its own margin */
.card {
  margin-bottom: 24px;  /* Assumes vertical stack context */
}

/* ✅ Layout context sets spacing */
.card-grid {
  display: grid;
  gap: 24px;  /* Parent controls spacing between cards */
}

.card-stack > * + * {
  margin-top: 24px;  /* Stack context controls spacing */
}
```

---

## Module 16 Summary

You learned:
- **Design tokens** — three-tier system (primitive → semantic → component), spacing/type scales
- **Theming** — dark mode via semantic token override, multi-brand, user preferences, FOWT prevention
- **Responsive systems** — fluid typography with `clamp()`, intrinsic layouts, container queries, minimal breakpoints
- **Component library CSS** — custom property API, flat specificity, `@layer` for libraries, state management, documentation

---

## Course Complete

You've covered the entire CSS rendering pipeline from first principles:

```
Parsing → Cascade → Computed Values → Layout → Paint → Composite
```

And built the skills to:
- Understand why CSS behaves the way it does
- Debug any visual issue systematically
- Architect CSS for large-scale production systems
- Use modern CSS features effectively
- Build and maintain component libraries

→ Return to [Course Home](../README.md)
