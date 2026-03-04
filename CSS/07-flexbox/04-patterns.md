# Lesson 04 — Patterns & Gotchas

## Production Patterns

### 1. Equal-Width Columns

```css
.columns {
  display: flex;
  gap: 20px;
}

.col {
  flex: 1;  /* = flex: 1 1 0% → starts at 0, grows equally */
}
```

Why `flex: 1` and not `width: 33%`? Because `flex: 1` adapts automatically to any number of columns and accounts for gaps.

### 2. Sidebar + Content

```css
.layout {
  display: flex;
  gap: 20px;
}

.sidebar {
  flex: 0 0 250px;  /* fixed width, no grow, no shrink */
}

.content {
  flex: 1;  /* takes remaining space */
}
```

### 3. Sticky Footer

```css
body {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

main {
  flex: 1;  /* grows to push footer down */
}

footer {
  flex-shrink: 0;  /* never collapse */
}
```

### 4. Centering (The Classic)

```css
.container {
  display: flex;
  justify-content: center;
  align-items: center;
}

/* Or: */
.container { display: flex; }
.child { margin: auto; }
```

### 5. Navbar with Spaced Groups

```css
.nav {
  display: flex;
  align-items: center;
  gap: 10px;
}

.nav-right {
  margin-left: auto;  /* pushes right group to the end */
}
```

### 6. Wrapping Card Grid

```css
.cards {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
}

.card {
  flex: 1 1 300px;  /* grow from 300px, wrap when smaller */
}
```

**Warning**: This doesn't produce a true grid — the last row may have fewer cards that grow wider. Use CSS Grid for equal-width wrapped layouts.

## Gotchas

### Gotcha 1: `min-width: auto`

By default, flex items have `min-width: auto` (not `min-width: 0`). This means items **refuse to shrink below their content size**.

```css
/* Problem: long text overflows */
.item {
  flex: 1;
  /* min-width: auto is implicit → text won't wrap, causes overflow */
}

/* Fix: */
.item {
  flex: 1;
  min-width: 0;  /* allow shrinking below content size */
}

/* Or: */
.item {
  flex: 1;
  overflow: hidden;  /* also resets effective min-width */
}
```

### Gotcha 2: `flex-basis` vs `width`

```css
/* These are NOT the same when flex-shrink or flex-grow apply: */
.item { flex-basis: 200px; }
.item { width: 200px; }
```

| | `flex-basis` | `width` |
|-|-------------|---------|
| Participates in flex algorithm | Yes ✓ | Only as fallback for `flex-basis: auto` |
| Applies to main axis | Always | Only when main axis = horizontal |
| With `flex-direction: column` | Acts as height | Still acts as width |

Rule: Always use `flex-basis` or the `flex` shorthand. Avoid `width` on flex items.

### Gotcha 3: Shrink Is Proportional to Size

```css
/* Item A: flex-shrink: 1, flex-basis: 500px
   Item B: flex-shrink: 1, flex-basis: 100px
   Container: 400px
   Overflow: 200px
   
   A's ratio: (1 × 500) / (1×500 + 1×100) = 5/6 → loses 167px → becomes 333px
   B's ratio: (1 × 100) / (1×500 + 1×100) = 1/6 → loses  33px → becomes  67px
   
   Even though both have shrink: 1, A loses MORE because it's bigger.
*/
```

### Gotcha 4: `flex-wrap` Breaks `align-items` Expectations

With `flex-wrap: wrap`, each line has its own cross-axis. `align-items` aligns within each line, not across all lines. Use `align-content` to control inter-line spacing.

### Gotcha 5: Margin Collapse Doesn't Happen

Flex containers **disable margin collapsing** for their children. Adjacent flex items' margins never collapse — they add up.

## Experiment: Gotchas

```html
<!-- 04-gotchas.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Flex Gotchas</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    h3 { margin-top: 30px; }
    
    .flex {
      display: flex;
      width: 400px;
      background: #e0e0e0;
      border: 2px solid #999;
      padding: 10px;
      gap: 10px;
      margin-bottom: 10px;
    }
    
    .item {
      background: lightblue;
      border: 2px solid steelblue;
      padding: 10px;
      font-family: monospace;
      font-size: 12px;
    }
    
    .label {
      font-family: monospace;
      font-size: 13px;
      margin-bottom: 5px;
    }
    
    .bad { background: #ffcccc; border-color: red; }
    .good { background: #ccffcc; border-color: green; }
  </style>
</head>
<body>
  <h2>Flex Gotchas</h2>
  
  <h3>Gotcha 1: min-width: auto</h3>
  
  <div class="label bad">❌ Long text overflows — min-width: auto prevents shrinking</div>
  <div class="flex">
    <div class="item bad" style="flex: 1;">Short</div>
    <div class="item bad" style="flex: 1;">This_is_a_very_long_unbreakable_string_that_will_cause_overflow_problems</div>
    <div class="item bad" style="flex: 1;">Short</div>
  </div>
  
  <div class="label good">✅ Fixed with min-width: 0</div>
  <div class="flex">
    <div class="item good" style="flex: 1; min-width: 0;">Short</div>
    <div class="item good" style="flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">This_is_a_very_long_unbreakable_string_that_will_cause_overflow_problems</div>
    <div class="item good" style="flex: 1; min-width: 0;">Short</div>
  </div>
  
  <h3>Gotcha 2: flex-basis vs width with column direction</h3>
  
  <div class="label">flex-direction: column — flex-basis acts as HEIGHT</div>
  <div class="flex" style="flex-direction: column; width: 200px; height: auto;">
    <div class="item" style="flex: 0 0 80px;">flex-basis: 80px (height!)</div>
    <div class="item" style="width: 200px; height: 80px;">width: 200px, height: 80px</div>
  </div>
  
  <h3>Gotcha 3: flex: 1 vs flex: auto</h3>
  
  <div class="label">flex: 1 (basis: 0) — equal widths regardless of content</div>
  <div class="flex">
    <div class="item" style="flex: 1;">Hi</div>
    <div class="item" style="flex: 1;">Hello World, this is more content here</div>
    <div class="item" style="flex: 1;">Hi</div>
  </div>
  
  <div class="label">flex: auto (basis: auto) — content affects width</div>
  <div class="flex">
    <div class="item" style="flex: auto;">Hi</div>
    <div class="item" style="flex: auto;">Hello World, this is more content here</div>
    <div class="item" style="flex: auto;">Hi</div>
  </div>

  <h3>Pattern: Sticky Footer</h3>
  <div style="display: flex; flex-direction: column; height: 250px; width: 300px; border: 2px solid #333;">
    <div style="background: #ddd; padding: 10px; font-family: monospace; font-size: 12px;">Header</div>
    <div style="flex: 1; background: #f9f9f9; padding: 10px; font-family: monospace; font-size: 12px;">Main (flex: 1)</div>
    <div style="background: #333; color: white; padding: 10px; font-family: monospace; font-size: 12px;">Footer (pushed down)</div>
  </div>
</body>
</html>
```

## When NOT to Use Flexbox

| Use Case | Use Instead |
|----------|-------------|
| 2D grid of equal-sized cards | **CSS Grid** |
| Complex page layout with named areas | **CSS Grid** |
| Items must align across rows | **CSS Grid** |
| Simple horizontal or vertical alignment | **Flexbox** ✓ |
| Dynamic number of items in one direction | **Flexbox** ✓ |
| Component-level layout (navbar, card body) | **Flexbox** ✓ |

Rule of thumb: **Flexbox = 1D**, **Grid = 2D**.

## Summary

- The flex sizing algorithm: flex-basis → grow/shrink → final size
- `flex: 1` for equal columns, `flex: auto` for content-proportional
- `margin: auto` absorbs free space (overrides justify-content)
- Always set `min-width: 0` on flex items that may overflow
- Flexbox is for 1D layout; use Grid for 2D

## Next Module

→ [Module 08: CSS Grid](../08-grid/README.md)
