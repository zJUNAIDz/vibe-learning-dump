# Lesson 03 — Utility-First CSS

## The Utility-First Idea

Instead of writing component classes, compose styles from single-purpose utility classes:

```html
<!-- BEM approach -->
<div class="card card--featured">
  <h3 class="card__title">Hello</h3>
</div>

<!-- Utility-first approach -->
<div class="border rounded-lg shadow-md p-4 bg-yellow-50">
  <h3 class="text-lg font-bold text-gray-900">Hello</h3>
</div>
```

Each utility does **one thing**:

```css
.p-4       { padding: 1rem; }
.text-lg   { font-size: 1.125rem; }
.font-bold { font-weight: 700; }
.rounded-lg { border-radius: 0.5rem; }
.shadow-md { box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1); }
.flex      { display: flex; }
.gap-4     { gap: 1rem; }
```

## Why It Works

### 1. CSS Stops Growing

With BEM, every new component adds new CSS. With utilities, new components reuse existing classes:

```
                BEM                     Utility-First
Components:  10 → 50 → 200         10 → 50 → 200
CSS size:    5KB → 25KB → 100KB    15KB → 18KB → 20KB
```

### 2. No Naming Fatigue

You never struggle to name `.card-wrapper-inner-container`. You describe what it looks like.

### 3. Refactoring Is Local

Changing a component means editing HTML, not hunting through CSS files wondering what else might break.

### 4. Design Constraints

A finite set of utilities enforces consistency:

```css
/* Only these spacings exist: */
.p-1 { padding: 0.25rem; }  /* 4px */
.p-2 { padding: 0.5rem; }   /* 8px */
.p-3 { padding: 0.75rem; }  /* 12px */
.p-4 { padding: 1rem; }     /* 16px */
.p-6 { padding: 1.5rem; }   /* 24px */
.p-8 { padding: 2rem; }     /* 32px */
/* No arbitrary padding: 13px — forced onto the scale */
```

## Tailwind CSS

The most popular utility-first framework. Key features:

### Design Tokens as Classes

```html
<!-- Colors from a curated palette -->
<p class="text-blue-600 bg-blue-50 border-blue-200">Info</p>

<!-- Responsive variants (mobile-first breakpoints) -->
<div class="flex flex-col md:flex-row lg:gap-8">...</div>

<!-- State variants -->
<button class="bg-blue-500 hover:bg-blue-600 active:bg-blue-700 
               focus:ring-2 focus:ring-blue-300 
               disabled:opacity-50 disabled:cursor-not-allowed">
  Submit
</button>

<!-- Dark mode -->
<div class="bg-white dark:bg-gray-900 text-black dark:text-white">
```

### Extracting Components

When utility strings repeat, extract them — but into **template components**, not CSS:

```jsx
// React component (preferred extraction)
function Badge({ children, color = "blue" }) {
  return (
    <span className={`
      inline-flex items-center px-2 py-1 
      rounded-full text-xs font-medium
      bg-${color}-100 text-${color}-800
    `}>
      {children}
    </span>
  );
}
```

Only use `@apply` when you can't use component extraction (e.g., third-party HTML):

```css
/* Only when necessary */
.prose h2 {
  @apply text-2xl font-bold mt-8 mb-4 text-gray-900;
}
```

## Trade-Offs

| Aspect | Pro | Con |
|--------|-----|-----|
| HTML readability | Styles are visible inline | Long class strings |
| CSS file size | Grows slowly (reuse) | Initial framework size (purged in prod) |
| Consistency | Design token enforcement | Must configure tokens properly |
| Learning curve | Low barrier per-utility | Must learn naming conventions |
| Refactoring | Local changes only | Find-replace across templates |
| Specificity | All utilities are (0,1,0) | Ordering matters for conflicts |
| Caching | CSS rarely changes | HTML changes more often |

## When Utility-First Works Best

1. **Component-based frameworks** (React, Vue, Svelte) — templates are the reuse boundary
2. **Rapid prototyping** — iterate visually without switching files
3. **Design system enforcement** — tokens constrain choices
4. **Small teams** — less CSS architecture to maintain

## When It Doesn't

1. **Content-heavy sites** — can't add classes to CMS-generated HTML
2. **Email templates** — need inline styles anyway
3. **CSS art / complex animations** — utilities can't express arbitrary values
4. **Teams that don't use component frameworks** — duplication across raw HTML pages

## Building Your Own Utility Layer

You don't need Tailwind. A small set of utilities augments any methodology:

```css
/* spacing.css */
:root {
  --space-1: 0.25rem;
  --space-2: 0.5rem;
  --space-3: 0.75rem;
  --space-4: 1rem;
  --space-6: 1.5rem;
  --space-8: 2rem;
}

.mt-1 { margin-top: var(--space-1); }
.mt-2 { margin-top: var(--space-2); }
/* ... generate with a preprocessor or postcss */

/* display.css */
.flex { display: flex; }
.grid { display: grid; }
.hidden { display: none; }
.block { display: block; }

/* text.css */
.text-center { text-align: center; }
.text-sm { font-size: 0.875rem; }
.text-lg { font-size: 1.125rem; }
.font-bold { font-weight: 700; }
.truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Utilities should come LAST in stylesheet order */
```

## Next

→ [Lesson 04: Component-Scoped CSS](04-component-scoped.md)
