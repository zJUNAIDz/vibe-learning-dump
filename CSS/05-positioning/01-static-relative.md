# Lesson 01 — Static & Relative Positioning

## Static (Default)

`position: static` is the default for every element. It means:

- Element is in **normal flow**
- `top`, `right`, `bottom`, `left` are **ignored**
- `z-index` is **ignored**
- The element does **not** create a containing block for positioned children

```css
.element {
  position: static; /* This is the default — you never need to write this */
  top: 50px;        /* Completely ignored */
  z-index: 10;      /* Completely ignored */
}
```

The only reason to ever write `position: static` explicitly is to **reset** a previously applied position.

## Relative Positioning

`position: relative` does two things:

1. **Visually offsets** the element from its normal position (using `top`/`right`/`bottom`/`left`)
2. **Creates a containing block** for absolutely positioned descendants

The key insight: **the element's original space in the flow is preserved**. Other elements don't move. The offset is purely visual.

```
Normal flow:         With position: relative; top: 20px; left: 30px:

┌─────────┐          ┌─────────┐
│    A    │          │    A    │
└─────────┘          └─────────┘
┌─────────┐  gap     ┌ ─ ─ ─ ─ ┐  ← original space preserved (ghost box)
│    B    │  still   │          │
└─────────┘  same    └ ─ ─ ─ ─ ┘
┌─────────┐          ┌─────────┐     ┌─────────┐
│    C    │          │    C    │     │    B    │ ← visually shifted
└─────────┘          └─────────┘     └─────────┘
```

## Experiment 01: Relative Positioning Behaviour

```html
<!-- 01-relative-positioning.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Relative Positioning</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .flow-container {
      width: 400px;
      background: #f0f0f0;
      border: 2px solid #999;
      padding: 10px;
    }
    
    .box {
      width: 200px;
      padding: 20px;
      margin-bottom: 10px;
      border: 2px solid;
      font-family: monospace;
      font-size: 13px;
    }
    
    .box-a { background: lightblue; border-color: steelblue; }
    .box-b { background: lightyellow; border-color: goldenrod; }
    .box-c { background: #f0fff0; border-color: green; }
    
    /* Relative offset on B */
    .box-b-relative {
      position: relative;
      top: 30px;
      left: 50px;
      /* This box moves visually but its space stays */
    }
    
    /* Overlay to show original position */
    .ghost {
      width: 200px;
      padding: 20px;
      margin-bottom: 10px;
      border: 2px dashed goldenrod;
      opacity: 0.3;
      pointer-events: none;
    }
  </style>
</head>
<body>
  <h2>position: relative — Visual Offset, Space Preserved</h2>
  
  <h3>Before (all static):</h3>
  <div class="flow-container">
    <div class="box box-a">A (static)</div>
    <div class="box box-b">B (static)</div>
    <div class="box box-c">C (static)</div>
  </div>
  
  <h3 style="margin-top: 30px;">After (B: position: relative; top: 30px; left: 50px):</h3>
  <div class="flow-container">
    <div class="box box-a">A (static)</div>
    <div class="box box-b box-b-relative">B (relative)<br>top: 30px, left: 50px</div>
    <div class="box box-c">C (static) — NOT MOVED. B's ghost space is still here.</div>
  </div>
  
  <div style="background: #fff3cd; padding: 15px; margin-top: 20px; border: 1px solid #ffc107; border-radius: 4px;">
    <strong>Key points:</strong>
    <ul>
      <li>C doesn't move — it still respects B's original position</li>
      <li>B overlaps content below it — this can cause visual overlap</li>
      <li>The offset doesn't affect layout of surrounding elements at all</li>
      <li>Use case: fine-tuning position without disrupting flow</li>
    </ul>
  </div>
</body>
</html>
```

## Experiment 02: Relative as Containing Block

```html
<!-- 02-relative-containing-block.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Relative as Containing Block</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .card {
      width: 300px;
      height: 200px;
      background: #f0f0f0;
      border: 2px solid #999;
      margin-bottom: 30px;
      padding: 15px;
    }
    
    .card-positioned {
      position: relative; /* This is the primary use case */
    }
    
    .badge {
      position: absolute;
      top: -10px;
      right: -10px;
      background: red;
      color: white;
      width: 24px;
      height: 24px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 12px;
      font-weight: bold;
    }
    
    .label {
      font-family: monospace;
      font-size: 12px;
    }
  </style>
</head>
<body>
  <h2>Relative Positioning: The Container Use Case</h2>
  
  <h3>Without position: relative (badge escapes to viewport)</h3>
  <div class="card">
    <div class="label">Card (position: static)</div>
    <div class="badge">3</div>
  </div>
  
  <h3>With position: relative (badge anchored to card)</h3>
  <div class="card card-positioned">
    <div class="label">Card (position: relative)</div>
    <div class="badge">3</div>
  </div>
  
  <div style="background: #d4edda; padding: 15px; border: 1px solid #28a745; border-radius: 4px;">
    <strong>This is the #1 use of position: relative:</strong><br>
    Creating a containing block for absolute children without moving the element itself.
    When you see <code>position: relative</code> without any <code>top/left/right/bottom</code>,
    that's what it's for.
  </div>
</body>
</html>
```

## Conflicting Offsets

When both `top` and `bottom` (or `left` and `right`) are set on a relatively positioned element:

- **LTR direction**: `left` wins over `right`
- **RTL direction**: `right` wins over `left`
- `top` always wins over `bottom`

```css
.element {
  position: relative;
  top: 10px;
  bottom: 20px;    /* Ignored — top wins */
  left: 10px;
  right: 20px;     /* Ignored in LTR — left wins */
}
```

## Next

→ [Lesson 02: Absolute & Fixed](02-absolute-fixed.md)
