# Lesson 04 — Offset Properties & Sizing

## How Offsets Interact with Dimensions

For absolutely positioned elements, the relationship between offsets and dimensions follows a constraint equation:

```
left + margin-left + border-left + padding-left + width + padding-right + border-right + margin-right + right
= containing block width

top + margin-top + border-top + padding-top + height + padding-bottom + border-bottom + margin-bottom + bottom
= containing block height
```

When some values are `auto`, the browser resolves them using these priority rules.

## Resolution Rules

### Horizontal (`left`, `width`, `right`)

| What's `auto`? | Result |
|----------------|--------|
| Everything | Element at its static position, shrink-to-fit width |
| Only `width` | Stretch between `left` and `right` |
| Only `left` | Compute from `right` + `width` |
| Only `right` | Compute from `left` + `width` |
| `left` + `right` | Use `width`, place at static position (LTR → left wins) |
| `left` + `width` | `width` = shrink-to-fit, position from `right` |
| `width` + `right` | `width` = shrink-to-fit, position from `left` |

### The `inset` Shorthand

`inset` is the shorthand for `top`, `right`, `bottom`, `left` (same order as margin/padding):

```css
inset: 10px;                 /* all four */
inset: 10px 20px;            /* vertical | horizontal */
inset: 10px 20px 30px;       /* top | horizontal | bottom */
inset: 10px 20px 30px 40px;  /* top | right | bottom | left */
```

## Experiment 01: Offset-Based Sizing

```html
<!-- 01-offset-sizing.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Offset Sizing</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .container {
      position: relative;
      width: 400px;
      height: 250px;
      background: #e0e0e0;
      border: 2px solid #999;
      margin-bottom: 30px;
    }
    
    .abs {
      position: absolute;
      background: rgba(65, 105, 225, 0.2);
      border: 2px solid royalblue;
      display: flex;
      align-items: center;
      justify-content: center;
      font-family: monospace;
      font-size: 11px;
      text-align: center;
    }
    
    .label { font-family: monospace; font-size: 13px; margin-bottom: 5px; }
  </style>
</head>
<body>
  <h2>Offset-Based Sizing</h2>

  <div class="label">left: 20px; right: 20px; (width stretches to fill)</div>
  <div class="container">
    <div class="abs" style="left: 20px; right: 20px; top: 10px; height: 50px;">
      left: 20, right: 20<br>width = 400 - 20 - 20 = 360px
    </div>
  </div>
  
  <div class="label">top: 20px; bottom: 20px; (height stretches to fill)</div>
  <div class="container">
    <div class="abs" style="top: 20px; bottom: 20px; left: 10px; width: 150px;">
      top: 20, bottom: 20<br>height = 250 - 20 - 20 = 210px
    </div>
  </div>
  
  <div class="label">inset: 30px; (all sides, like a frame)</div>
  <div class="container">
    <div class="abs" style="inset: 30px;">
      inset: 30px<br>width = 340, height = 190
    </div>
  </div>
  
  <div class="label">top: 0; right: 0; (auto width/height → shrink-to-fit)</div>
  <div class="container">
    <div class="abs" style="top: 0; right: 0;">
      Shrink<br>to fit
    </div>
  </div>
  
  <div class="label">inset: 0; width: 200px; height: 80px; margin: auto; (CENTERED)</div>
  <div class="container">
    <div class="abs" style="inset: 0; width: 200px; height: 80px; margin: auto;">
      Perfectly centered<br>using inset: 0 + margin: auto
    </div>
  </div>
</body>
</html>
```

## Experiment 02: Static Position

When an absolutely positioned element has `auto` for its offsets, it appears at its **static position** — where it would have been if it were `position: static`:

```html
<!-- 02-static-position.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Static Position</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .container {
      position: relative;
      width: 400px;
      background: #e0e0e0;
      border: 2px solid #999;
      padding: 15px;
    }
    
    .normal-a {
      background: lightblue;
      padding: 15px;
      border: 2px solid steelblue;
      margin-bottom: 10px;
    }
    
    .abs-no-offsets {
      position: absolute;
      /* No top, right, bottom, left — uses static position */
      background: lightyellow;
      border: 2px solid goldenrod;
      padding: 15px;
    }
    
    .normal-b {
      background: #f0fff0;
      padding: 15px;
      border: 2px solid green;
    }
    
    .label { font-family: monospace; font-size: 12px; }
  </style>
</head>
<body>
  <h2>Absolute with No Offsets → Static Position</h2>
  
  <div class="container">
    <div class="normal-a"><div class="label">Normal A</div></div>
    <div class="abs-no-offsets">
      <div class="label">position: absolute (no offsets)</div>
      Appears where it WOULD have been in normal flow
    </div>
    <div class="normal-b">
      <div class="label">Normal B — ignores the absolute element above</div>
      This element is positioned as if the absolute element doesn't exist.
    </div>
  </div>
  
  <div style="background: #fff3cd; padding: 15px; margin-top: 20px; border: 1px solid #ffc107; border-radius: 4px;">
    <strong>Static position:</strong> The absolute element appears at its normal flow position 
    but is removed from flow. Normal B slides up to take its space, causing overlap.
  </div>
</body>
</html>
```

## Experiment 03: Centering Techniques Compared

```html
<!-- 03-centering-compared.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Centering Techniques</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; max-width: 800px; }
    
    .container {
      position: relative;
      width: 100%;
      height: 200px;
      background: #e0e0e0;
      border: 2px solid #999;
    }
    
    .centered {
      width: 120px;
      height: 60px;
      background: lightblue;
      border: 2px solid steelblue;
      display: flex;
      align-items: center;
      justify-content: center;
      font-family: monospace;
      font-size: 10px;
      text-align: center;
    }
    
    .method-label {
      font-family: monospace;
      font-size: 11px;
      background: #333;
      color: white;
      padding: 4px 8px;
    }
  </style>
</head>
<body>
  <h2>Absolute Centering Techniques Compared</h2>
  
  <div class="grid">
    <div>
      <div class="method-label">1. translate(-50%, -50%)</div>
      <div class="container">
        <div class="centered" style="position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%);">
          translate<br>hack
        </div>
      </div>
    </div>
    
    <div>
      <div class="method-label">2. inset: 0; margin: auto; ✅ BEST</div>
      <div class="container">
        <div class="centered" style="position: absolute; inset: 0; margin: auto;">
          inset + auto<br>margin
        </div>
      </div>
    </div>
    
    <div>
      <div class="method-label">3. Flex on parent (no positioning needed)</div>
      <div class="container" style="display: flex; align-items: center; justify-content: center;">
        <div class="centered" style="position: static;">
          flex on<br>parent
        </div>
      </div>
    </div>
    
    <div>
      <div class="method-label">4. Grid + place-items (no positioning needed)</div>
      <div class="container" style="display: grid; place-items: center;">
        <div class="centered" style="position: static;">
          grid<br>place-items
        </div>
      </div>
    </div>
  </div>
  
  <div style="background: #d4edda; padding: 15px; margin-top: 20px; border: 1px solid #28a745; border-radius: 4px;">
    <strong>Recommendation hierarchy:</strong>
    <ol>
      <li><strong>Grid/Flex</strong>: Best for layout-level centering (avoids positioning entirely)</li>
      <li><strong>inset: 0; margin: auto</strong>: Best for absolute centering (no sub-pixel issues)</li>
      <li><strong>translate</strong>: Works when you don't know the element's size, but can cause blurry text at half-pixels</li>
    </ol>
  </div>
</body>
</html>
```

## Summary Table

| Technique | Use When | Drawback |
|-----------|---------|----------|
| Explicit `width` + `left` | Known size, known position | Rigid |
| `left` + `right` (no `width`) | Stretch to fill with margins | Elements must be absolute |
| `inset: 0` + `margin: auto` + size | Centering in containing block | Needs explicit width/height |
| `top: 50%; left: 50%; transform` | Centering when size is unknown | Sub-pixel blur possible |
| Flex/Grid centering | Component-level centering | May change parent's layout mode |

## Next Module

→ [Module 06: Stacking Contexts](../06-stacking-contexts/README.md)
