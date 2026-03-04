# Lesson 02 — BEM & Methodologies

## BEM (Block, Element, Modifier)

The most widely used CSS naming convention. It encodes component relationships into class names:

```
Block:    .card
Element:  .card__title     (part of card, uses __)
Modifier: .card--featured  (variant of card, uses --)
```

### Rules

1. **Block**: Standalone component name (`.card`, `.nav`, `.form`)
2. **Element**: Part of a block, prefixed with `__` (`.card__title`, `.card__image`)
3. **Modifier**: Variant or state, suffixed with `--` (`.card--large`, `.card__title--muted`)

```css
/* Block */
.card {
  border: 1px solid #ddd;
  border-radius: 8px;
  overflow: hidden;
}

/* Elements */
.card__image {
  width: 100%;
  aspect-ratio: 16 / 9;
  object-fit: cover;
}

.card__title {
  font-size: 18px;
  font-weight: bold;
  padding: 16px 16px 0;
}

.card__body {
  padding: 8px 16px 16px;
  color: #666;
}

/* Modifiers */
.card--featured {
  border-color: gold;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.card__title--muted {
  color: #999;
}
```

```html
<div class="card card--featured">
  <img class="card__image" src="photo.jpg" alt="">
  <h3 class="card__title">Featured Article</h3>
  <p class="card__body">Description text...</p>
</div>
```

### BEM Benefits

| Benefit | How |
|---------|-----|
| No specificity wars | All selectors are single class (0,1,0) |
| Self-documenting | `.card__title` clearly belongs to `.card` |
| No nesting needed | `.card__title` instead of `.card .title` |
| Greppable | Search for `.card__` to find all card parts |
| Refactorable | Rename the block, rename all elements |

### BEM Pitfalls

```css
/* ❌ DON'T: nest elements more than one level */
.card__header__title__icon { }  /* This is wrong */

/* ✅ DO: flatten to block__element */
.card__header-icon { }

/* ❌ DON'T: use element selectors */
.card h3 { }

/* ✅ DO: use BEM class */
.card__title { }
```

## Other Methodologies

### OOCSS (Object-Oriented CSS)

Separate **structure** from **skin**, and **container** from **content**:

```css
/* Structure */
.media { display: flex; gap: 16px; }
.media__body { flex: 1; }

/* Skin */
.theme-primary { color: blue; }
.theme-danger { color: red; }
```

### SMACSS (Scalable & Modular Architecture)

Categorizes CSS into five types:

| Category | Prefix | Example |
|----------|--------|---------|
| Base | (none) | `body`, `a`, `h1` |
| Layout | `l-` | `.l-sidebar`, `.l-grid` |
| Module | (none) | `.card`, `.nav` |
| State | `is-` | `.is-active`, `.is-hidden` |
| Theme | `theme-` | `.theme-dark` |

### ITCSS (Inverted Triangle CSS)

Organizes CSS from **most generic to most specific**:

```
Settings    → $variables (no CSS output)
Tools       → @mixins, @functions (no CSS output)
Generic     → reset, normalize
Elements    → bare HTML (h1, a, input)
Objects     → layout patterns (.container, .grid)
Components  → UI components (.card, .nav)
Utilities   → overrides (.u-hidden, .u-text-center)
```

Each layer has **increasing specificity**, so utilities always win.

## Choosing a Methodology

| Methodology | Best For |
|-------------|----------|
| **BEM** | Teams, large codebases, clear component boundaries |
| **OOCSS** | Design systems with reusable patterns |
| **SMACSS** | Server-rendered apps with distinct page sections |
| **ITCSS** | Large-scale projects needing strict ordering |

In practice, most teams use **BEM naming** with **ITCSS organization**.

## Next

→ [Lesson 03: Utility-First CSS](03-utility-first.md)
