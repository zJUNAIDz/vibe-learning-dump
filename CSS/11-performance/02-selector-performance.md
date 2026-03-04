# Lesson 02 — Selector & Style Performance

## How Browsers Match Selectors

Browsers match selectors **right to left**:

```css
.sidebar .nav ul li a { color: blue; }
/*  5     4  3  2  1  ← matching order */
```

1. Find all `<a>` elements in the document
2. For each, check if parent is `<li>`
3. Check if ancestor is `<ul>`
4. Check if ancestor is `.nav`
5. Check if ancestor is `.sidebar`

**Right-to-left matching** is fast because most selectors fail early — most `<a>` elements aren't inside `.sidebar .nav ul li`.

## Which Selectors Are Expensive?

| Selector | Cost | Why |
|----------|------|-----|
| `.class` | Fast | Simple hash lookup |
| `#id` | Fast | Single hash lookup |
| `element` | Fast | Tag-indexed |
| `.parent .child` | Medium | Ancestry walk |
| `.parent > .child` | Fast | One-level check |
| `*` | Slow | Matches everything |
| `[attr]` | Medium | Attribute scan |
| `:nth-child(n)` | Medium | Sibling counting |
| `.a .b .c .d .e` | Slow | Deep ancestry walk |
| `:has(.x)` | Variable | Forward-looking (browser-optimized) |

### Practical Truth

In modern browsers, **selector matching is rarely the bottleneck**. With 1000+ rules, the difference between fast and slow selectors is microseconds. The real performance costs are:

1. **Style recalculation scope** — how many elements need restyling
2. **Layout thrashing** — forced synchronous layout
3. **Paint complexity** — expensive paint operations

## Style Recalculation

When a class changes, the browser must determine which elements are affected:

```javascript
// Changes class on one element — browser restyres subtree
document.body.classList.add('dark-mode');
// Every rule with .dark-mode is re-evaluated against the ENTIRE DOM
```

### Reducing Scope

```css
/* ❌ Every element re-evaluated: */
.dark-mode * { color: white; }

/* ✅ Scoped — only affected elements re-styled: */
.dark-mode .text { color: white; }
.dark-mode .heading { color: #eee; }
```

Containment helps:

```css
.widget {
  contain: style;  /* counters/quotes scoped */
  contain: content; /* layout + paint contained */
}
```

## CSS Variables & Performance

CSS custom properties are **live** — changing one triggers restyles on all elements that use it.

```css
:root {
  --primary: blue;  /* changing this restyles everything using --primary */
}
```

Scope variables as narrowly as possible:

```css
/* ❌ Global → all elements re-evaluated: */
:root { --card-bg: white; }

/* ✅ Scoped → only .card descendants re-evaluated: */
.card { --card-bg: white; }
```

## Expensive Style Patterns

| Pattern | Why It's Expensive |
|---------|-------------------|
| Very large `box-shadow` | Painted per frame, large blur radius = slow |
| `filter: blur()` on large elements | Pixel-by-pixel computation |
| `backdrop-filter` | Composites everything behind the element |
| Complex gradients with many stops | More computation per paint |
| `border-radius` on many elements | Clip computation |
| `text-shadow` with blur | Per-glyph computation |

## Experiment: Style Recalculation

```html
<!-- 02-style-recalc.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Style Recalculation</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .grid {
      display: grid;
      grid-template-columns: repeat(20, 1fr);
      gap: 2px;
      margin-bottom: 20px;
    }
    
    .cell {
      height: 20px;
      background: steelblue;
      border-radius: 2px;
      transition: background-color 0.15s;
    }
    
    /* Broad selector — triggers massive restyle */
    .highlight-all .cell { background: tomato; }
    
    /* Scoped selector — minimal restyle */
    .cell.highlighted { background: limegreen; }
    
    button {
      padding: 10px 20px;
      font-family: monospace;
      font-size: 13px;
      margin-right: 10px;
      margin-bottom: 10px;
      cursor: pointer;
      border: 2px solid;
      border-radius: 6px;
    }
    
    .result {
      font-family: monospace;
      font-size: 13px;
      padding: 10px;
      margin-top: 10px;
      background: #f5f5f5;
      border-radius: 4px;
    }
  </style>
</head>
<body>
  <h2>Style Recalculation Scope</h2>
  
  <div class="grid" id="grid"></div>
  
  <button style="background: #ffcccc; border-color: red;" onclick="broadRestyle()">
    ❌ Broad: toggle class on parent (restyles all)
  </button>
  <button style="background: #ccffcc; border-color: green;" onclick="scopedRestyle()">
    ✅ Scoped: toggle class on individual cells
  </button>
  <button onclick="reset()">Reset</button>
  
  <div class="result" id="result">
    Open DevTools → Performance → Record while clicking.
    <br>Look for "Recalculate Style" events and their duration.
  </div>

  <script>
    const grid = document.getElementById('grid');
    for (let i = 0; i < 400; i++) {
      const cell = document.createElement('div');
      cell.className = 'cell';
      grid.appendChild(cell);
    }
    
    function broadRestyle() {
      const start = performance.now();
      grid.classList.toggle('highlight-all');
      // Force style recalc measurement:
      getComputedStyle(grid.firstElementChild).backgroundColor;
      const time = (performance.now() - start).toFixed(2);
      document.getElementById('result').textContent = `Broad restyle: ${time}ms (all 400 cells re-evaluated)`;
    }
    
    function scopedRestyle() {
      const start = performance.now();
      const cells = grid.querySelectorAll('.cell');
      const quarter = Math.floor(cells.length / 4);
      for (let i = 0; i < quarter; i++) {
        cells[i].classList.toggle('highlighted');
      }
      getComputedStyle(cells[0]).backgroundColor;
      const time = (performance.now() - start).toFixed(2);
      document.getElementById('result').textContent = `Scoped restyle: ${time}ms (${quarter} cells toggled)`;
    }
    
    function reset() {
      grid.classList.remove('highlight-all');
      grid.querySelectorAll('.cell').forEach(c => c.classList.remove('highlighted'));
      document.getElementById('result').textContent = 'Reset.';
    }
  </script>
</body>
</html>
```

## Summary

- Selector matching is right-to-left and rarely a bottleneck in modern browsers
- **Style recalculation scope** matters more than selector complexity
- Scope CSS variables as narrowly as possible
- Use `contain` to isolate subtrees from restyle propagation
- Avoid expensive paint patterns (`blur`, large `box-shadow`)

## Next

→ [Lesson 03: Critical CSS & Loading](03-critical-css.md)
