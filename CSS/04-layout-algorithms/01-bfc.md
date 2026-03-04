# Lesson 01 — Block Formatting Context (BFC)

## Concept

A **Block Formatting Context** is an isolated layout region where:

1. Boxes stack **vertically**, one after another
2. Each box's left edge touches the left edge of the containing block
3. Vertical margins between adjacent boxes **collapse**
4. The BFC **contains** all its descendants — floats, margins, everything stays inside

A BFC is the "normal" block layout algorithm with one crucial addition: **isolation**. The root `<html>` element establishes the initial BFC.

## What Creates a New BFC?

| Trigger | Example | Notes |
|---------|---------|-------|
| Root element | `<html>` | The initial BFC |
| Float | `float: left / right` | Always creates a BFC |
| Absolute/fixed position | `position: absolute / fixed` | Taken out of flow, own BFC |
| `display: inline-block` | `display: inline-block` | Classic BFC trigger |
| `display: flow-root` | `display: flow-root` | **Modern, purpose-built** BFC creator |
| `overflow` != `visible` | `overflow: hidden / auto / scroll` | Side-effect BFC |
| `display: flex / grid` | Flex/grid container | Children use flex/grid layout, but the container establishes a BFC for its own positioning |
| Table cells | `display: table-cell` | Each cell is a BFC |
| `contain: layout / content / paint` | CSS Containment | Modern containment API |
| Multicol containers | `column-count / column-width` | Each column is a BFC |

## Why BFCs Matter

### Problem 1: Float Containment

Without a BFC, a parent doesn't "see" its floated children — it collapses to zero height:

```html
<!-- 01-float-containment.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>BFC Float Containment</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .container {
      background: #e0e0e0;
      border: 2px solid #999;
      margin-bottom: 30px;
    }
    
    .float-child {
      float: left;
      width: 150px;
      height: 100px;
      background: lightcoral;
      border: 2px solid darkred;
      margin: 10px;
    }
    
    .bfc-container {
      display: flow-root; /* <-- Creates BFC */
      background: #d4edda;
      border: 2px solid #28a745;
      margin-bottom: 30px;
    }
    
    .clearfix::after {
      content: '';
      display: block;
      clear: both;
    }
    
    .label {
      font-family: monospace;
      font-size: 13px;
      padding: 5px;
    }
    
    .after-element {
      background: lightyellow;
      padding: 10px;
      border: 1px solid goldenrod;
    }
  </style>
</head>
<body>
  <h2>BFC Float Containment</h2>
  
  <h3>No BFC — Parent collapses, content leaks</h3>
  <div class="container" id="noBfc">
    <div class="float-child">Float A</div>
    <div class="float-child">Float B</div>
    <div class="label">No clearfix, no BFC</div>
  </div>
  <div class="after-element">This element is affected by the uncontained floats above</div>
  
  <h3 style="margin-top: 60px;">Old fix: clearfix hack</h3>
  <div class="container clearfix" id="clearfixed">
    <div class="float-child">Float A</div>
    <div class="float-child">Float B</div>
    <div class="label">clearfix::after { clear: both }</div>
  </div>
  <div class="after-element">This is properly below</div>
  
  <h3>Modern fix: display: flow-root (creates BFC)</h3>
  <div class="bfc-container" id="bfcFixed">
    <div class="float-child">Float A</div>
    <div class="float-child">Float B</div>
    <div class="label">display: flow-root</div>
  </div>
  <div class="after-element">This is properly below — clean solution</div>
  
  <script>
    ['noBfc', 'clearfixed', 'bfcFixed'].forEach(id => {
      const el = document.getElementById(id);
      console.log(`${id} height:`, el.offsetHeight);
    });
  </script>
</body>
</html>
```

### Problem 2: Margin Containment

A BFC prevents parent-child margin collapsing (covered in Module 03):

```css
.parent {
  display: flow-root; /* BFC prevents child margin from leaking out */
}
```

### Problem 3: Float Exclusion

Content in a BFC won't wrap around adjacent floats — it stays in its own column:

```html
<!-- 02-bfc-exclusion.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>BFC Float Exclusion</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .wrapper {
      width: 500px;
      background: #f0f0f0;
      border: 1px solid #ccc;
      padding: 10px;
      margin-bottom: 30px;
    }
    
    .sidebar {
      float: left;
      width: 150px;
      height: 120px;
      background: lightcoral;
      border: 2px solid darkred;
      margin-right: 10px;
    }
    
    .content-no-bfc {
      background: lightyellow;
      border: 2px solid goldenrod;
      padding: 10px;
    }
    
    .content-bfc {
      display: flow-root; /* BFC: won't wrap around float */
      background: #d4edda;
      border: 2px solid green;
      padding: 10px;
    }
  </style>
</head>
<body>
  <h2>BFC Float Exclusion</h2>
  
  <h3>Without BFC: text wraps around float</h3>
  <div class="wrapper" style="display: flow-root;">
    <div class="sidebar">Float</div>
    <div class="content-no-bfc">
      This content wraps around the float.
      Lorem ipsum dolor sit amet, consectetur adipiscing elit.
      The text flows into the space beside and below the float.
      This is the default inline flow behaviour.
    </div>
  </div>
  
  <h3>With BFC: content stays in its own column</h3>
  <div class="wrapper" style="display: flow-root;">
    <div class="sidebar">Float</div>
    <div class="content-bfc">
      This content has display: flow-root (BFC).
      It creates its own isolated block — no wrapping around the float.
      This is how you create a two-column layout with floats.
      The content sits entirely beside the float.
    </div>
  </div>
</body>
</html>
```

## BFC Mental Model

```
Think of a BFC as a "glass box":
- Everything inside is contained (floats, margins)
- Nothing leaks out
- Nothing from outside leaks in
- Each BFC is its own layout world

display: flow-root === "make this element a glass box"
```

## DevTools Exercise

1. Open Experiment 01
2. In DevTools, temporarily add `display: flow-root` to the broken container
3. Watch it expand to contain its floated children
4. Remove it — watch it collapse again
5. Compare `overflow: hidden` (old hack) vs `flow-root` (modern clean approach)

## Next

→ [Lesson 02: Inline Formatting Context](02-ifc.md)
