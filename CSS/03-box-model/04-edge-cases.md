# Lesson 04 — Edge Cases

## 1. Percentage Margins

One of CSS's most counter-intuitive rules:

> **All percentage margins (including `margin-top` and `margin-bottom`) resolve against the containing block's WIDTH — not height.**

This applies to padding too. The reason is practical: height is often unknown or auto, so resolving against width provides a stable reference.

```html
<!-- 01-percentage-margins.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Percentage Margins</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .container {
      width: 400px;
      height: 200px;
      background: #e0e0e0;
      border: 2px solid #999;
      position: relative;
    }
    
    .child {
      width: 100px;
      height: 50px;
      background: lightyellow;
      border: 2px solid goldenrod;
      margin: 10%; /* 10% of 400px = 40px — ALL FOUR SIDES */
    }
    
    .measurement {
      font-family: monospace;
      font-size: 13px;
      background: #fff3cd;
      padding: 10px;
      margin: 10px 0;
    }
    
    .container-tall {
      width: 200px;
      height: 600px;
      background: #e0e0e0;
      border: 2px solid #999;
    }
    
    .child-tall {
      width: 100px;
      height: 50px;
      background: #f0fff0;
      border: 2px solid green;
      margin: 10%; /* 10% of 200px = 20px — still uses WIDTH */
    }
  </style>
</head>
<body>
  <h2>Percentage Margins Resolve Against Containing Block WIDTH</h2>
  
  <h3>Container: 400px × 200px, Child: margin: 10%</h3>
  <div class="container">
    <div class="child" id="child1">margin: 10%</div>
  </div>
  <div class="measurement" id="m1"></div>
  
  <h3 style="margin-top: 30px;">Container: 200px × 600px, Child: margin: 10%</h3>
  <p style="font-size: 14px; color: #666;">Even though the container is taller,<br>margin still resolves against width (200px), not height (600px)</p>
  <div class="container-tall">
    <div class="child-tall" id="child2">margin: 10%</div>
  </div>
  <div class="measurement" id="m2"></div>

  <script>
    function measure(childId, measureId) {
      const el = document.getElementById(childId);
      const cs = getComputedStyle(el);
      document.getElementById(measureId).innerHTML = [
        `margin-top: ${cs.marginTop}`,
        `margin-right: ${cs.marginRight}`,
        `margin-bottom: ${cs.marginBottom}`,
        `margin-left: ${cs.marginLeft}`,
        `(All resolve against containing block WIDTH, even top/bottom)`,
      ].join('<br>');
    }
    measure('child1', 'm1');
    measure('child2', 'm2');
  </script>
</body>
</html>
```

### Percentage Padding — Same Rule

`padding: 10%` also resolves against containing block width on **all four sides**. This enables the classic "aspect ratio hack":

```css
/* Maintain 16:9 aspect ratio (before aspect-ratio property) */
.aspect-box {
  width: 100%;
  height: 0;
  padding-bottom: 56.25%; /* 9/16 = 0.5625 */
}
```

## 2. Replaced Elements

Replaced elements (`<img>`, `<video>`, `<iframe>`, `<canvas>`, `<input>`) have **intrinsic dimensions** and behave differently from normal block/inline elements.

```html
<!-- 02-replaced-elements.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Replaced Elements</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .demo { margin-bottom: 30px; }
    
    .container {
      width: 300px;
      background: #f0f0f0;
      border: 2px solid #ccc;
      padding: 10px;
    }
    
    /* Images are inline-replaced by default */
    img {
      border: 2px solid navy;
    }
    
    /* Unlike normal inline elements, images CAN have width/height */
    .sized-img {
      width: 200px;
      height: 100px;
    }
    
    /* Unlike normal inline elements, images CAN have vertical margin/padding */
    .margined-img {
      margin-top: 30px;
      margin-bottom: 30px;
      padding: 20px;
      background: lightyellow;
    }
    
    .note {
      font-size: 13px;
      font-family: monospace;
      color: #666;
    }
  </style>
</head>
<body>
  <h2>Replaced Elements — Special Box Model Rules</h2>
  
  <div class="demo">
    <h3>1. Images are inline by default, but accept width/height</h3>
    <p class="note">Normal inline elements (&lt;span&gt;) IGNORE width/height.</p>
    <div class="container">
      <p>Text before
        <img src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='150' height='80'%3E%3Crect fill='%23ddd' width='150' height='80'/%3E%3Ctext x='50%25' y='50%25' text-anchor='middle' dy='.3em' font-size='14'%3Edefault%3C/text%3E%3C/svg%3E" class="default-img" alt="default">
        text after
      </p>
      <p>Text before
        <img src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='150' height='80'%3E%3Crect fill='%23ddd' width='150' height='80'/%3E%3Ctext x='50%25' y='50%25' text-anchor='middle' dy='.3em' font-size='14'%3Esized%3C/text%3E%3C/svg%3E" class="sized-img" alt="sized">
        text after
      </p>
    </div>
  </div>
  
  <div class="demo">
    <h3>2. Images accept vertical margin and padding</h3>
    <p class="note">Normal inline elements IGNORE margin-top/bottom. Replaced elements don't.</p>
    <div class="container">
      <p>Text before
        <img src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='100' height='50'%3E%3Crect fill='%23ddd' width='100' height='50'/%3E%3Ctext x='50%25' y='50%25' text-anchor='middle' dy='.3em' font-size='12'%3Emargined%3C/text%3E%3C/svg%3E" class="margined-img" alt="margined">
        text after
      </p>
    </div>
  </div>
  
  <div class="demo">
    <h3>3. object-fit (controls how content fills the box)</h3>
    <div style="display: flex; gap: 10px;">
      <div>
        <div class="note">object-fit: fill (default, distorts)</div>
        <img src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='200' height='100'%3E%3Crect fill='%23c8e6c9' width='200' height='100'/%3E%3Ctext x='50%25' y='50%25' text-anchor='middle' dy='.3em'%3E200×100%3C/text%3E%3C/svg%3E"
             style="width: 100px; height: 100px; object-fit: fill; border: 2px solid green;">
      </div>
      <div>
        <div class="note">object-fit: contain (letterbox)</div>
        <img src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='200' height='100'%3E%3Crect fill='%23c8e6c9' width='200' height='100'/%3E%3Ctext x='50%25' y='50%25' text-anchor='middle' dy='.3em'%3E200×100%3C/text%3E%3C/svg%3E"
             style="width: 100px; height: 100px; object-fit: contain; border: 2px solid green;">
      </div>
      <div>
        <div class="note">object-fit: cover (crop)</div>
        <img src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='200' height='100'%3E%3Crect fill='%23c8e6c9' width='200' height='100'/%3E%3Ctext x='50%25' y='50%25' text-anchor='middle' dy='.3em'%3E200×100%3C/text%3E%3C/svg%3E"
             style="width: 100px; height: 100px; object-fit: cover; border: 2px solid green;">
      </div>
    </div>
  </div>
</body>
</html>
```

## 3. Inline Box Model

Inline elements (`<span>`, `<a>`, `<em>`, `<strong>`) have a fundamentally different box model:

| Property | Block | Inline |
|----------|-------|--------|
| `width` / `height` | ✅ Applies | ❌ Ignored |
| `margin-left` / `margin-right` | ✅ | ✅ |
| `margin-top` / `margin-bottom` | ✅ | ❌ Ignored |
| `padding` (all sides) | ✅ | ✅ Applies but doesn't push surrounding lines |
| `border` (all sides) | ✅ | ✅ Applies but doesn't push surrounding lines |
| Line-height | N/A | Determines vertical space |

```html
<!-- 03-inline-box-model.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Inline Box Model</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; line-height: 2.5; }
    
    .inline-demo {
      background: #f0f0f0;
      padding: 20px;
      width: 500px;
      border: 1px solid #ccc;
      margin-bottom: 30px;
    }
    
    .highlight {
      background: lightyellow;
      border: 2px solid goldenrod;
      padding: 10px 15px;        /* Padding APPLIES but doesn't affect line spacing */
      margin: 50px 20px;          /* Horizontal: yes. Vertical: IGNORED */
    }
    
    .compared-block {
      display: inline-block;      /* inline-block gets FULL box model */
      background: #f0fff0;
      border: 2px solid green;
      padding: 10px 15px;
      margin: 50px 20px;          /* ALL margins apply */
      vertical-align: middle;
    }
    
    .label {
      font-family: monospace;
      font-size: 12px;
      color: #666;
    }
  </style>
</head>
<body>
  <h2>Inline vs inline-block Box Model</h2>
  
  <h3>Inline (span) — vertical margin/padding don't affect layout</h3>
  <div class="inline-demo">
    Lorem ipsum dolor sit amet, 
    <span class="highlight">this span has padding: 10px, margin: 50px, border: 2px</span>
    but notice how the vertical padding and border OVERLAP adjacent lines.
    The vertical margin is completely ignored. Only horizontal margin works.
    Additional text to show how lines aren't pushed apart by the span's padding.
  </div>
  
  <h3>inline-block — full box model applies</h3>
  <div class="inline-demo">
    Lorem ipsum dolor sit amet,
    <span class="compared-block">this inline-block has the same styles</span>
    and notice how it properly pushes content away in ALL directions.
    Vertical margins, padding, and borders all affect layout.
    Additional text to show the difference.
  </div>
  
  <div style="background: #fff3cd; padding: 15px; border: 1px solid #ffc107; border-radius: 4px;">
    <h3 style="margin-top: 0;">Key Takeaway</h3>
    <ul>
      <li><code>inline</code>: vertical padding/border render visually but DON'T affect layout (they overlap)</li>
      <li><code>inline-block</code>: full box model, participates in inline flow but respects all dimensions</li>
      <li><code>block</code>: full box model, takes full width, stacks vertically</li>
    </ul>
  </div>
</body>
</html>
```

## 4. `min-width` / `max-width` Resolution Order

The spec defines a strict resolution order:

```
1. Compute tentative width from the width property
2. If tentative > max-width, use max-width
3. If result < min-width, use min-width

In other words: min-width ALWAYS wins over max-width.
```

```css
/* min-width beats max-width beats width */
.box {
  width: 500px;
  max-width: 300px;   /* constrains to 300px */
  min-width: 400px;   /* overrides max-width → result is 400px */
}
```

## 5. The `auto` Margin Trick

When margins are set to `auto` on a block element with a fixed width, the browser distributes remaining space equally:

```css
/* Horizontal centering */
.centered {
  width: 600px;
  margin-left: auto;
  margin-right: auto;
  /* Browser: remaining = container - 600, each side gets half */
}

/* Right-align */
.right-aligned {
  width: 600px;
  margin-left: auto;
  margin-right: 0;
}
```

For flex items, `margin: auto` absorbs **all** remaining space in both axes:

```css
.flex-container { display: flex; }
.flex-child { margin-left: auto; } /* pushes to right (like float: right) */
.flex-child { margin: auto; }       /* centers both horizontally AND vertically */
```

## 6. `display: contents`

Removes the element's box entirely — as if the element didn't exist and its children were direct children of the parent.

```css
.wrapper {
  display: contents;
  /* This element's box, margin, padding, border are all gone.
     Its children participate in the parent's layout directly.
     Useful for semantic wrappers that shouldn't affect layout. */
}
```

⚠️ **Accessibility warning**: `display: contents` on certain elements (like `<button>`, `<table>`) can remove their semantic role from the accessibility tree in some browsers.

## DevTools Exercise

1. Create an `<img>` and a `<span>` with the same `padding: 20px; margin: 20px; border: 2px solid red`
2. Inspect both in DevTools → **Computed** tab → Box model diagram
3. Notice: the img shows all four margins active; the span shows vertical margins as 0
4. Toggle `display: inline-block` on the span and watch the box model update live

## Summary

| Edge Case | Key Insight |
|-----------|------------|
| Percentage margins | ALL sides resolve against containing block **width** |
| Replaced elements | Inline but accept width, height, vertical margin |
| Inline box model | Vertical padding/border render but don't affect layout |
| min/max resolution | `min-width` > `max-width` > `width` |
| `margin: auto` | Distributes remaining space (centering trick) |
| `display: contents` | Box disappears, children promote to parent's layout |

## Next Module

→ [Module 04: Layout Algorithms](../04-layout-algorithms/README.md)
