# Lesson 01 — Flex Container & Axes

## Creating a Flex Container

```css
.container {
  display: flex;       /* block-level flex container */
  /* or */
  display: inline-flex; /* inline-level flex container */
}
```

When you set `display: flex`, the element becomes a **flex container** and all its **direct children** become **flex items**. The children's `display` value (block, inline, etc.) is largely irrelevant — they now follow flex layout rules.

## The Two Axes

Everything in flexbox revolves around two perpendicular axes:

| Property | Main Axis | Cross Axis |
|----------|-----------|------------|
| `flex-direction: row` | → (left-to-right in LTR) | ↓ (top-to-bottom) |
| `flex-direction: row-reverse` | ← (right-to-left in LTR) | ↓ |
| `flex-direction: column` | ↓ (top-to-bottom) | → (left-to-right in LTR) |
| `flex-direction: column-reverse` | ↑ (bottom-to-top) | → |

```
flex-direction: row (default)
┌───────────────────────────────────────────┐
│ ┌───────┐  ┌───────┐  ┌───────┐          │
│ │ Item1 │  │ Item2 │  │ Item3 │  ← Main  │
│ └───────┘  └───────┘  └───────┘    axis   │
│ ↕ Cross axis                              │
└───────────────────────────────────────────┘

flex-direction: column
┌──────────────┐
│ ┌──────────┐ │ ↕ Main axis
│ │  Item 1  │ │
│ └──────────┘ │
│ ┌──────────┐ │
│ │  Item 2  │ │
│ └──────────┘ │
│ ┌──────────┐ │
│ │  Item 3  │ │
│ └──────────┘ │
│ ← Cross →   │
└──────────────┘
```

## Container Properties

| Property | Values | Default | What It Does |
|----------|--------|---------|-------------|
| `flex-direction` | `row` / `row-reverse` / `column` / `column-reverse` | `row` | Sets main axis direction |
| `flex-wrap` | `nowrap` / `wrap` / `wrap-reverse` | `nowrap` | Allow items to wrap to new lines |
| `flex-flow` | `<direction> <wrap>` | `row nowrap` | Shorthand for direction + wrap |
| `justify-content` | Various | `flex-start` | Distribute on main axis |
| `align-items` | Various | `stretch` | Align on cross axis |
| `align-content` | Various | `stretch` | Distribute wrapped lines on cross axis |
| `gap` | `<length>` | `0` | Space between items |

## Experiment 01: Axes and Direction

```html
<!-- 01-axes-and-direction.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Flex Axes</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .demos { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; max-width: 800px; }
    
    .flex-container {
      display: flex;
      background: #e0e0e0;
      border: 2px solid #999;
      padding: 10px;
      gap: 10px;
      min-height: 150px;
    }
    
    .flex-item {
      background: lightblue;
      border: 2px solid steelblue;
      padding: 15px 20px;
      font-family: monospace;
      font-size: 13px;
      text-align: center;
    }
    
    .method-label {
      font-family: monospace;
      font-size: 12px;
      background: #333;
      color: white;
      padding: 4px 8px;
      margin-bottom: 5px;
    }
    
    h3 { margin-top: 0; margin-bottom: 5px; }
  </style>
</head>
<body>
  <h2>flex-direction: Controlling the Axes</h2>
  
  <div class="demos">
    <div>
      <div class="method-label">flex-direction: row (default)</div>
      <div class="flex-container" style="flex-direction: row;">
        <div class="flex-item">1</div>
        <div class="flex-item">2</div>
        <div class="flex-item">3</div>
      </div>
    </div>
    
    <div>
      <div class="method-label">flex-direction: row-reverse</div>
      <div class="flex-container" style="flex-direction: row-reverse;">
        <div class="flex-item">1</div>
        <div class="flex-item">2</div>
        <div class="flex-item">3</div>
      </div>
    </div>
    
    <div>
      <div class="method-label">flex-direction: column</div>
      <div class="flex-container" style="flex-direction: column; min-height: 200px;">
        <div class="flex-item">1</div>
        <div class="flex-item">2</div>
        <div class="flex-item">3</div>
      </div>
    </div>
    
    <div>
      <div class="method-label">flex-direction: column-reverse</div>
      <div class="flex-container" style="flex-direction: column-reverse; min-height: 200px;">
        <div class="flex-item">1</div>
        <div class="flex-item">2</div>
        <div class="flex-item">3</div>
      </div>
    </div>
  </div>
</body>
</html>
```

## Experiment 02: Wrapping

```html
<!-- 02-wrapping.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Flex Wrapping</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .flex-container {
      display: flex;
      background: #e0e0e0;
      border: 2px solid #999;
      padding: 10px;
      gap: 10px;
      width: 400px;
      margin-bottom: 30px;
    }
    
    .flex-item {
      background: lightblue;
      border: 2px solid steelblue;
      padding: 15px 20px;
      font-family: monospace;
      font-size: 13px;
      min-width: 100px;
      text-align: center;
    }
    
    .label { font-family: monospace; font-size: 13px; margin-bottom: 5px; }
  </style>
</head>
<body>
  <h2>flex-wrap Behaviour</h2>
  
  <div class="label">flex-wrap: nowrap (default) — items SHRINK to fit</div>
  <div class="flex-container" style="flex-wrap: nowrap;">
    <div class="flex-item">One</div>
    <div class="flex-item">Two</div>
    <div class="flex-item">Three</div>
    <div class="flex-item">Four</div>
    <div class="flex-item">Five</div>
  </div>
  
  <div class="label">flex-wrap: wrap — items wrap to new lines</div>
  <div class="flex-container" style="flex-wrap: wrap;">
    <div class="flex-item">One</div>
    <div class="flex-item">Two</div>
    <div class="flex-item">Three</div>
    <div class="flex-item">Four</div>
    <div class="flex-item">Five</div>
  </div>
  
  <div class="label">flex-wrap: wrap-reverse — new lines added ABOVE</div>
  <div class="flex-container" style="flex-wrap: wrap-reverse;">
    <div class="flex-item">One</div>
    <div class="flex-item">Two</div>
    <div class="flex-item">Three</div>
    <div class="flex-item">Four</div>
    <div class="flex-item">Five</div>
  </div>
  
  <div style="background: #fff3cd; padding: 15px; border: 1px solid #ffc107; border-radius: 4px;">
    <strong>Critical:</strong> With <code>flex-wrap: nowrap</code> (default), items will shrink 
    below their min-content size rather than overflow (unless you set <code>min-width: 0</code>
    or <code>overflow: hidden</code>). The <code>flex-shrink</code> algorithm kicks in.
  </div>
</body>
</html>
```

## What Happens to Children

When an element becomes a flex item:
- Its `display` is "blockified" (inline-level becomes block-level)
- `float` and `clear` are ignored
- `vertical-align` is ignored
- It establishes a new formatting context for its own children
- Its margins don't collapse (flex margins never collapse)

## Next

→ [Lesson 02: The Flex Sizing Algorithm](02-sizing-algorithm.md)
