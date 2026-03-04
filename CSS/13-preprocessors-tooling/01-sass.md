# Lesson 01 — Sass/SCSS

## What Sass Adds to CSS

Sass (Syntactically Awesome Style Sheets) compiles to standard CSS. **SCSS** is the CSS-compatible syntax (uses `{}`). The indented syntax (`.sass`) uses whitespace instead of braces.

This lesson covers SCSS, the more widely used syntax.

## Variables

```scss
// _variables.scss
$color-primary: #2563eb;
$color-danger: #dc2626;
$spacing-unit: 8px;
$font-stack: 'Inter', system-ui, sans-serif;
$breakpoint-md: 768px;

// Usage
.button {
  background: $color-primary;
  padding: $spacing-unit ($spacing-unit * 2);
  font-family: $font-stack;
}
```

**Sass variables are compile-time.** They produce static values in the output — no runtime behavior, no cascade interaction.

```scss
// This does NOT work like CSS custom properties
$color: blue;
.parent { $color: red; }  // Creates a new scoped variable
.child { color: $color; }  // Still blue — Sass scoping, not cascade
```

## Nesting

```scss
.nav {
  display: flex;
  gap: 16px;

  &__item {            // & = parent selector → .nav__item
    padding: 8px 12px;

    &:hover {           // .nav__item:hover
      background: #f5f5f5;
    }

    &--active {         // .nav__item--active
      font-weight: bold;
      border-bottom: 2px solid blue;
    }
  }

  &__link {
    text-decoration: none;
    color: inherit;
  }
}
```

**Compiles to:**

```css
.nav { display: flex; gap: 16px; }
.nav__item { padding: 8px 12px; }
.nav__item:hover { background: #f5f5f5; }
.nav__item--active { font-weight: bold; border-bottom: 2px solid blue; }
.nav__link { text-decoration: none; color: inherit; }
```

### Nesting Rule

**Never nest deeper than 3 levels.** Each nesting level increases specificity and coupling:

```scss
// ❌ Too deep — produces .page .sidebar .nav .item .link
.page {
  .sidebar {
    .nav {
      .item {
        .link { color: blue; }
      }
    }
  }
}

// ✅ Flat BEM with nesting only for &
.sidebar-nav {
  &__item { padding: 8px; }
  &__link { color: blue; }
}
```

## Mixins

Reusable blocks of declarations, optionally with parameters:

```scss
@mixin respond-to($breakpoint) {
  @if $breakpoint == 'md' {
    @media (min-width: 768px) { @content; }
  } @else if $breakpoint == 'lg' {
    @media (min-width: 1024px) { @content; }
  } @else if $breakpoint == 'xl' {
    @media (min-width: 1280px) { @content; }
  }
}

@mixin truncate($lines: 1) {
  @if $lines == 1 {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  } @else {
    display: -webkit-box;
    -webkit-line-clamp: $lines;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
}

// Usage
.card__title {
  @include truncate(2);

  @include respond-to('md') {
    font-size: 24px;
  }
}
```

### Common Mixin Patterns

```scss
// Visually hidden (accessible)
@mixin sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

// Focus ring
@mixin focus-ring($color: $color-primary, $offset: 2px) {
  outline: 2px solid $color;
  outline-offset: $offset;
}

// Container
@mixin container($max-width: 1200px) {
  width: 100%;
  max-width: $max-width;
  margin-inline: auto;
  padding-inline: 16px;
}
```

## @extend vs @mixin

`@extend` shares selectors. `@mixin` duplicates declarations.

```scss
// --- @extend ---
%button-base {
  padding: 8px 16px;
  border-radius: 4px;
  border: none;
  cursor: pointer;
}

.button-primary {
  @extend %button-base;          // Shares the selector
  background: blue;
}

.button-danger {
  @extend %button-base;
  background: red;
}
```

**Compiled output of @extend:**

```css
.button-primary, .button-danger {
  padding: 8px 16px;
  border-radius: 4px;
  border: none;
  cursor: pointer;
}
.button-primary { background: blue; }
.button-danger { background: red; }
```

**The `@extend` problem:** Selector explosion in large codebases. Selectors get merged across distant parts of the stylesheet, making output unpredictable.

**Rule: Prefer `@mixin` over `@extend`.** The small duplication is worth the predictability.

## Functions

```scss
// Strip unit
@function strip-unit($value) {
  @return $value / ($value * 0 + 1);
}

// Convert px to rem
@function to-rem($px, $base: 16px) {
  @return ($px / $base) * 1rem;
}

// Usage
.card {
  padding: to-rem(24px);     // 1.5rem
  font-size: to-rem(14px);   // 0.875rem
}
```

### Built-in Functions

```scss
// Color manipulation
$base: #2563eb;

.button {
  background: $base;
  &:hover { background: darken($base, 10%); }
  &:active { background: darken($base, 20%); }
}

.badge {
  background: lighten($base, 40%);
  color: darken($base, 20%);
  border: 1px solid lighten($base, 20%);
}

// Math
$columns: 12;
.col-4 { width: percentage(4 / $columns); }   // 33.3333%

// String
$icon: 'arrow';
.icon-#{$icon} { /* ... */ }   // .icon-arrow

// List
$breakpoints: (sm: 640px, md: 768px, lg: 1024px);
@each $name, $value in $breakpoints {
  .container-#{$name} { max-width: $value; }
}
```

## Partials and @use

```scss
// File structure
// styles/
//   _variables.scss    (underscore = partial, not compiled directly)
//   _mixins.scss
//   _reset.scss
//   components/
//     _button.scss
//     _card.scss
//   main.scss

// main.scss
@use 'variables' as vars;     // Namespaced import
@use 'mixins' as mix;
@use 'reset';                  // Default namespace: reset
@use 'components/button';
@use 'components/card';

.container {
  max-width: vars.$max-width;  // Access with namespace
  @include mix.respond-to('md') { padding: 24px; }
}
```

**`@use` vs `@import`:**

| Feature | `@import` (deprecated) | `@use` |
|---------|----------------------|--------|
| Scope | Global (pollutes) | Namespaced |
| Execution | Every `@import` re-evaluates | Loaded once, cached |
| Variables | All global | Accessed via namespace |
| Status | Being deprecated | Current standard |

## When to Still Use Sass

Despite native CSS improvements, Sass remains valuable for:

1. **Loops and generation** — `@each`, `@for` to generate utility classes
2. **Complex math** — functions with conditionals
3. **Map data structures** — design token management
4. **Mixins** — especially media query wrappers
5. **Legacy codebases** — gradual migration

## Next

→ [Lesson 02: PostCSS](02-postcss.md)
