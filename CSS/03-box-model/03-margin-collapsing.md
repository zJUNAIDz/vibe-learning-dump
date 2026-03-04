# Lesson 03 — Margin Collapsing

## Concept

Margin collapsing is CSS's most misunderstood behaviour. It is **not a bug** — it was intentionally designed for typographic documents where paragraphs should have consistent spacing regardless of nesting.

**The rule**: When two **vertical** margins are adjacent (no border, padding, or content between them), they **collapse** into a single margin equal to the **larger** of the two.

```
     ┌──────────────────────┐
     │ element A             │
     │ margin-bottom: 30px   │
     └──────────────────────┘
              │
        30px  │  ← NOT 50px (30 + 20)
              │     only the larger margin survives
              │
     ┌──────────────────────┐
     │ element B             │
     │ margin-top: 20px      │
     └──────────────────────┘
```

## Three Types of Margin Collapsing

### 1. Adjacent Siblings

Consecutive block-level siblings collapse their adjoining vertical margins.

```
<p style="margin-bottom: 30px">A</p>
<p style="margin-top: 20px">B</p>

Result: 30px gap (not 50px)
```

### 2. Parent-Child (First/Last Child)

A parent's margin collapses with its first child's `margin-top` (or last child's `margin-bottom`) if there is **nothing** between them (no border, no padding, no inline content, no `height`/`min-height` on the parent, no BFC boundary).

```
Parent margin-top: 0
Child margin-top: 40px

Result: The 40px "leaks" out of the parent.
The parent appears to have margin-top: 40px.
```

### 3. Empty Blocks

An element with no height, no border, no padding, and no content collapses its own `margin-top` and `margin-bottom` into a single margin.

## Collapsing Rules Summary

| Condition | Collapses? | Why |
|-----------|-----------|-----|
| Two adjacent siblings (vertical margins) | ✅ Yes | Adjoining margins |
| Parent + first-child margin-top | ✅ Yes | No separation between them |
| Parent + last-child margin-bottom | ✅ Yes | No separation between them |
| Empty block's own margins | ✅ Yes | Top and bottom are adjacent |
| Horizontal margins | ❌ Never | Only vertical margins collapse |
| Margins on floated elements | ❌ Never | Floats establish BFC |
| Margins on flex/grid items | ❌ Never | Flex/grid formatting contexts |
| Margins on absolutely positioned elements | ❌ Never | Out of normal flow |
| Margins separated by border/padding | ❌ No | Border/padding breaks adjacency |
| Margins on elements with `overflow` != `visible` | ❌ No | Creates BFC |

## Negative Margin Collapsing

When negative margins are involved:

- **Both positive**: max(A, B)
- **Both negative**: min(A, B) (most negative wins)
- **One positive, one negative**: A + B (they add together)

## Experiment 01: Collapsing Visualized

```html
<!-- 01-collapsing-visualized.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Margin Collapsing</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .demo-section {
      margin-bottom: 40px;
    }
    
    /* === Adjacent Siblings === */
    .siblings-demo {
      background: #f0f0f0;
      padding: 1px 10px; /* padding prevents parent-child collapse */
    }
    
    .box-a {
      background: lightblue;
      padding: 15px;
      margin-bottom: 30px;
      border: 2px solid steelblue;
    }
    
    .box-b {
      background: lightyellow;
      padding: 15px;
      margin-top: 20px;
      border: 2px solid goldenrod;
    }
    
    /* === Parent-Child === */
    .parent-demo {
      background: #e0e0e0;
      margin-top: 0;
      /* no padding, no border → child's margin will leak */
    }
    
    .child-leaky {
      background: lightcoral;
      padding: 15px;
      margin-top: 40px;
    }
    
    .parent-demo-fixed {
      background: #e0e0e0;
      padding-top: 1px; /* This single pixel prevents the collapse */
    }
    
    .child-fixed {
      background: lightgreen;
      padding: 15px;
      margin-top: 40px;
    }
    
    /* === Empty Blocks === */
    .empty-demo-container { background: #f0f0f0; padding: 1px 10px; }
    .before-empty { background: lightblue; padding: 15px; margin-bottom: 20px; }
    .empty-block { margin-top: 30px; margin-bottom: 40px; /* own margins collapse to 40px */ }
    .after-empty { background: lightyellow; padding: 15px; margin-top: 10px; }
    
    .label {
      font-family: monospace;
      font-size: 12px;
      color: #666;
    }
    
    .measurement {
      background: #fff3cd;
      padding: 10px;
      margin: 5px 0;
      font-family: monospace;
      font-size: 13px;
    }
  </style>
</head>
<body>
  <h2>Margin Collapsing — All Three Types</h2>
  
  <!-- Adjacent Siblings -->
  <div class="demo-section">
    <h3>1. Adjacent Siblings</h3>
    <p class="label">Box A: margin-bottom: 30px, Box B: margin-top: 20px → gap = 30px (not 50px)</p>
    <div class="siblings-demo">
      <div class="box-a" id="boxA">Box A (margin-bottom: 30px)</div>
      <div class="box-b" id="boxB">Box B (margin-top: 20px)</div>
    </div>
    <div class="measurement" id="siblingMeasure"></div>
  </div>
  
  <!-- Parent-Child -->
  <div class="demo-section">
    <h3>2. Parent-Child Collapse (margin leak)</h3>
    <p class="label">Child margin-top: 40px leaks out of parent (no padding/border):</p>
    <div style="border: 2px dashed red; margin-bottom: 20px;">
      <div class="parent-demo" id="parentLeaky">
        <div class="child-leaky" id="childLeaky">Child (margin-top: 40px)</div>
      </div>
    </div>
    
    <p class="label">Fixed: 1px padding on parent prevents collapse:</p>
    <div style="border: 2px dashed green;">
      <div class="parent-demo-fixed" id="parentFixed">
        <div class="child-fixed" id="childFixed">Child (margin-top: 40px, contained)</div>
      </div>
    </div>
  </div>
  
  <!-- Empty Block -->
  <div class="demo-section">
    <h3>3. Empty Block Self-Collapse</h3>
    <p class="label">Empty block between: margin-top: 30px, margin-bottom: 40px → collapses to 40px</p>
    <div class="empty-demo-container">
      <div class="before-empty" id="beforeEmpty">Before (margin-bottom: 20px)</div>
      <div class="empty-block" id="emptyBlock"></div>
      <div class="after-empty" id="afterEmpty">After (margin-top: 10px)</div>
    </div>
    <div class="measurement" id="emptyMeasure"></div>
  </div>

  <script>
    // Measure actual gaps
    const boxA = document.getElementById('boxA');
    const boxB = document.getElementById('boxB');
    const rectA = boxA.getBoundingClientRect();
    const rectB = boxB.getBoundingClientRect();
    const gap = rectB.top - rectA.bottom;
    document.getElementById('siblingMeasure').textContent =
      `Actual gap between siblings: ${gap}px (30 wins over 20)`;
    
    const before = document.getElementById('beforeEmpty');
    const after = document.getElementById('afterEmpty');
    const gap2 = after.getBoundingClientRect().top - before.getBoundingClientRect().bottom;
    document.getElementById('emptyMeasure').textContent =
      `Actual gap with empty block: ${gap2}px (empty block collapses 30/40 → 40, then collapses with siblings)`;
  </script>
</body>
</html>
```

## Experiment 02: Preventing Collapse

```html
<!-- 02-preventing-collapse.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Preventing Margin Collapse</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .demo-row {
      display: flex;
      gap: 20px;
      margin-bottom: 30px;
    }
    
    .demo-col {
      flex: 1;
    }
    
    .wrapper {
      background: #e0e0e0;
      min-height: 80px;
    }
    
    .child {
      background: lightblue;
      padding: 15px;
      margin: 20px 0;
    }
    
    /* Prevention methods */
    .prevent-padding { padding-top: 1px; padding-bottom: 1px; }
    .prevent-border { border-top: 1px solid transparent; border-bottom: 1px solid transparent; }
    .prevent-overflow { overflow: hidden; }
    .prevent-display { display: flow-root; }
    .prevent-flex { display: flex; flex-direction: column; }
    
    .method-label {
      font-family: monospace;
      font-size: 11px;
      background: #333;
      color: white;
      padding: 4px 8px;
      margin-bottom: 4px;
    }
    
    h3 { margin-top: 0; }
  </style>
</head>
<body>
  <h2>5 Ways to Prevent Parent-Child Margin Collapse</h2>
  
  <div class="demo-row">
    <div class="demo-col">
      <div class="method-label">DEFAULT (margin leaks)</div>
      <div class="wrapper">
        <div class="child">Child (margin: 20px 0)</div>
      </div>
    </div>
    
    <div class="demo-col">
      <div class="method-label">padding: 1px</div>
      <div class="wrapper prevent-padding">
        <div class="child">Child (margin: 20px 0)</div>
      </div>
    </div>
    
    <div class="demo-col">
      <div class="method-label">border: 1px solid transparent</div>
      <div class="wrapper prevent-border">
        <div class="child">Child (margin: 20px 0)</div>
      </div>
    </div>
  </div>
  
  <div class="demo-row">
    <div class="demo-col">
      <div class="method-label">overflow: hidden (BFC)</div>
      <div class="wrapper prevent-overflow">
        <div class="child">Child (margin: 20px 0)</div>
      </div>
    </div>
    
    <div class="demo-col">
      <div class="method-label">display: flow-root (BFC) ✅ BEST</div>
      <div class="wrapper prevent-display">
        <div class="child">Child (margin: 20px 0)</div>
      </div>
    </div>
    
    <div class="demo-col">
      <div class="method-label">display: flex (different FC)</div>
      <div class="wrapper prevent-flex">
        <div class="child">Child (margin: 20px 0)</div>
      </div>
    </div>
  </div>
  
  <div style="background: #d4edda; border: 1px solid #28a745; padding: 15px; margin-top: 20px; border-radius: 4px;">
    <h3>Recommendation</h3>
    <p><code>display: flow-root</code> is the cleanest solution. It was specifically designed to create a new Block Formatting Context without side effects (unlike <code>overflow: hidden</code> which clips content, or padding/border hacks that add unwanted space).</p>
    <p>However, in modern layouts using flexbox or grid, margin collapsing simply doesn't apply — it's a block-flow-only behaviour.</p>
  </div>
</body>
</html>
```

## Experiment 03: Negative Margins + Collapsing

```html
<!-- 03-negative-margin-collapsing.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Negative Margin Collapsing</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .example {
      background: #f0f0f0;
      padding: 1px 10px;
      margin-bottom: 30px;
    }
    
    .block {
      padding: 15px;
      border: 2px solid;
    }
    .block-a { background: lightblue; border-color: steelblue; }
    .block-b { background: lightyellow; border-color: goldenrod; }
    
    .result {
      font-family: monospace;
      font-size: 13px;
      background: #fff3cd;
      padding: 10px;
      margin: 5px 0 15px;
    }
  </style>
</head>
<body>
  <h2>Negative Margin Collapsing Rules</h2>
  
  <h3>Both positive: max(30, 20) = 30px</h3>
  <div class="example" id="ex1">
    <div class="block block-a" style="margin-bottom: 30px;">A (margin-bottom: 30px)</div>
    <div class="block block-b" style="margin-top: 20px;">B (margin-top: 20px)</div>
  </div>
  
  <h3>Both negative: min(-30, -20) = -30px</h3>
  <div class="example" id="ex2">
    <div class="block block-a" style="margin-bottom: -30px;">A (margin-bottom: -30px)</div>
    <div class="block block-b" style="margin-top: -20px;">B (margin-top: -20px)</div>
  </div>
  
  <h3>Mixed: 30 + (-20) = 10px</h3>
  <div class="example" id="ex3">
    <div class="block block-a" style="margin-bottom: 30px;">A (margin-bottom: 30px)</div>
    <div class="block block-b" style="margin-top: -20px;">B (margin-top: -20px)</div>
  </div>
  
  <h3>Mixed (reverse): -30 + 20 = -10px (they overlap!)</h3>
  <div class="example" id="ex4">
    <div class="block block-a" style="margin-bottom: -30px;">A (margin-bottom: -30px)</div>
    <div class="block block-b" style="margin-top: 20px;">B (margin-top: 20px)</div>
  </div>
  
  <script>
    function measureGap(containerId, label) {
      const container = document.getElementById(containerId);
      const blocks = container.querySelectorAll('.block');
      const rectA = blocks[0].getBoundingClientRect();
      const rectB = blocks[1].getBoundingClientRect();
      const gap = rectB.top - rectA.bottom;
      
      const result = document.createElement('div');
      result.className = 'result';
      result.textContent = `Measured gap: ${gap}px`;
      container.after(result);
    }
    
    measureGap('ex1', 'Both positive');
    measureGap('ex2', 'Both negative');
    measureGap('ex3', 'Mixed positive-negative');
    measureGap('ex4', 'Mixed negative-positive');
  </script>
</body>
</html>
```

## DevTools Exercise

1. Open any of the experiments above
2. Select an element → **Computed** tab
3. Look at the box model diagram — the margin area shows the **declared** margin, not the collapsed result
4. To see the actual collapsed margin:
   - Hover over elements in the **Elements** panel
   - The orange margin overlay shows what the browser actually used
   - Compare with adjacent elements
5. `display: flow-root` on a parent is the modern, clean way to prevent all parent-child collapse

## Key Mental Model

```
Margin collapsing only happens in BLOCK flow layout.
Flex items, grid items, floated elements, absolutely positioned elements → NEVER collapse.
If you use modern layout, you rarely encounter it.
But when you do, it's usually the parent-child leak that's confusing.
Fix: display: flow-root on the parent.
```

## Next

→ [Lesson 04: Edge Cases](04-edge-cases.md)
