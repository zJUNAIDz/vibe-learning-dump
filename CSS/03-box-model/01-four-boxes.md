# Lesson 01 — The Four Boxes

## Concept

Every CSS element generates a rectangular box with four distinct areas. The `width` and `height` properties reference one of these areas depending on the `box-sizing` value.

```
┌─────────────────────── margin-box ────────────────────────┐
│ (margin — transparent, collapses)                         │
│  ┌────────────────── border-box ───────────────────────┐  │
│  │ (border — visible, has width)                       │  │
│  │  ┌────────────── padding-box ────────────────────┐  │  │
│  │  │ (padding — uses element's background)         │  │  │
│  │  │  ┌────────── content-box ──────────────────┐  │  │  │
│  │  │  │ (content — text, images, children)      │  │  │  │
│  │  │  │                                         │  │  │  │
│  │  │  └─────────────────────────────────────────┘  │  │  │
│  │  └───────────────────────────────────────────────┘  │  │
│  └─────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────┘
```

### Key Facts

| Area | Background? | Can be negative? | Collapses? |
|---|---|---|---|
| Content | Yes (element's background) | No | No |
| Padding | Yes (element's background) | No (clamped to 0) | No |
| Border | Has own color/style | No | No |
| Margin | Always transparent | **Yes** | **Yes** (vertical only, block elements) |

## Experiment 01: Visualizing the Box Model

```html
<!-- 01-box-model-visual.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Box Model Visual</title>
  <style>
    * { margin: 0; padding: 0; }
    body { font-family: system-ui; padding: 40px; background: #f0f0f0; }
    
    .box-demo {
      /* Content dimensions */
      width: 300px;
      height: 150px;
      
      /* Padding — extends background area */
      padding: 20px 30px 20px 30px;
      
      /* Border — visible edge */
      border: 5px solid navy;
      
      /* Margin — transparent space outside */
      margin: 40px;
      
      /* Background shows through content + padding */
      background: lightyellow;
      
      /* Default: box-sizing: content-box */
      box-sizing: content-box;
      
      color: #333;
      font-size: 14px;
    }
    
    .container {
      background: #e0e0e0;
      border: 1px dashed #999;
      display: inline-block;
    }
    
    .measurements {
      font-family: monospace;
      font-size: 13px;
      margin-top: 20px;
      background: white;
      padding: 15px;
      border-radius: 4px;
    }
  </style>
</head>
<body>
  <h2>Box Model Dimensions (box-sizing: content-box)</h2>
  
  <div class="container">
    <div class="box-demo" id="box">
      Content area: 300×150px<br>
      Padding: 20px top/bottom, 30px left/right<br>
      Border: 5px solid<br>
      Margin: 40px all sides
    </div>
  </div>
  
  <div class="measurements" id="measurements"></div>

  <script>
    const box = document.getElementById('box');
    const m = document.getElementById('measurements');
    const cs = getComputedStyle(box);
    const rect = box.getBoundingClientRect();
    
    m.innerHTML = `
<strong>CSS Properties:</strong>
width: ${cs.width}  (content-box width)
height: ${cs.height}  (content-box height)

<strong>Computed Sizes:</strong>
content:    ${box.clientWidth - parseInt(cs.paddingLeft) - parseInt(cs.paddingRight)} × ${box.clientHeight - parseInt(cs.paddingTop) - parseInt(cs.paddingBottom)}
+ padding:  ${cs.paddingTop} / ${cs.paddingRight} / ${cs.paddingBottom} / ${cs.paddingLeft}
= clientWidth × clientHeight: ${box.clientWidth} × ${box.clientHeight}
+ border:   ${cs.borderTopWidth} / ${cs.borderRightWidth} / ${cs.borderBottomWidth} / ${cs.borderLeftWidth}
= offsetWidth × offsetHeight: ${box.offsetWidth} × ${box.offsetHeight}
+ margin:   ${cs.marginTop} / ${cs.marginRight} / ${cs.marginBottom} / ${cs.marginLeft}
= Total space: ${box.offsetWidth + parseInt(cs.marginLeft) + parseInt(cs.marginRight)} × ${box.offsetHeight + parseInt(cs.marginTop) + parseInt(cs.marginBottom)}

<strong>Key JavaScript Properties:</strong>
clientWidth/Height — content + padding (NO border, NO scrollbar)
offsetWidth/Height — content + padding + border (NO margin)
getBoundingClientRect() — same as offset (includes border)
    `.trim();
  </script>
</body>
</html>
```

### DevTools Exercise

1. Select the `.box-demo` element in DevTools
2. Look at the **box model diagram** at the bottom of the Styles/Computed panel
3. Hover over each area (content, padding, border, margin) — Chrome highlights that area on screen
4. Note the colors: blue = content, green = padding, orange = margin, dark = border

## Experiment 02: Padding Background Behavior

```html
<!-- 02-padding-background.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Padding and Background</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: system-ui; padding: 30px; }
    
    .demo { margin: 20px 0; }
    
    /* Background extends through padding by default */
    .bg-default {
      width: 300px;
      padding: 40px;
      background: cornflowerblue;
      border: 3px solid navy;
      color: white;
    }
    
    /* background-clip changes this behavior */
    .bg-content-box {
      width: 300px;
      padding: 40px;
      background: coral;
      background-clip: content-box;  /* Background stops at content edge */
      border: 3px solid darkred;
      color: white;
    }
    
    .bg-padding-box {
      width: 300px;
      padding: 40px;
      background: mediumseagreen;
      background-clip: padding-box;  /* Default behavior, explicitly */
      border: 10px dashed darkgreen; /* Dashed to see through border */
      color: white;
    }
    
    .bg-border-box {
      width: 300px;
      padding: 40px;
      background: mediumpurple;
      background-clip: border-box;  /* Background extends under border */
      border: 10px dashed rgba(0,0,0,0.3);
      color: white;
    }
    
    h3 { margin: 15px 0 5px; }
  </style>
</head>
<body>
  <h2>Where Does Background Paint?</h2>
  
  <h3>background-clip: padding-box (default)</h3>
  <div class="demo">
    <div class="bg-default">Background fills content + padding</div>
  </div>
  
  <h3>background-clip: content-box</h3>
  <div class="demo">
    <div class="bg-content-box">Background stops at content edge. Padding is transparent.</div>
  </div>
  
  <h3>background-clip: padding-box (explicit, dashed border)</h3>
  <div class="demo">
    <div class="bg-padding-box">Background stops at padding edge. Look through the dashed border.</div>
  </div>
  
  <h3>background-clip: border-box (dashed border)</h3>
  <div class="demo">
    <div class="bg-border-box">Background extends under the border. See purple through the dashes.</div>
  </div>
</body>
</html>
```

## Experiment 03: Negative Margins

```html
<!-- 03-negative-margins.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Negative Margins</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: system-ui; padding: 40px; }
    
    .container {
      width: 400px;
      background: #f0f0f0;
      border: 2px solid #ccc;
      padding: 20px;
      margin-bottom: 30px;
    }
    
    .box {
      padding: 15px;
      background: cornflowerblue;
      color: white;
      margin: 10px 0;
    }
    
    /* Negative margin-top: pulls element UP */
    .pull-up { margin-top: -20px; background: coral; }
    
    /* Negative margin-left: pulls element LEFT */
    .pull-left { margin-left: -30px; background: mediumseagreen; }
    
    /* Negative margin-right: pulls NEXT element closer */
    .pull-next { margin-right: -40px; background: mediumpurple; }
    
    /* Negative margin to "break out" of container */
    .breakout {
      margin-left: -20px;   /* equals container padding */
      margin-right: -20px;  /* equals container padding */
      padding-left: 20px;
      background: gold;
      color: #333;
    }
    
    h3 { margin: 20px 0 10px; color: #333; }
    .note { font-size: 13px; color: #666; margin: 5px 0; font-style: italic; }
  </style>
</head>
<body>
  <h2>Negative Margins</h2>
  
  <h3>1. Negative margin-top (pulls element up)</h3>
  <div class="container">
    <div class="box">Normal box</div>
    <div class="box pull-up">margin-top: -20px (overlaps above)</div>
  </div>
  
  <h3>2. Negative margin-left (shifts element left)</h3>
  <div class="container">
    <div class="box pull-left">margin-left: -30px (extends outside container)</div>
  </div>
  
  <h3>3. Negative margin-right (pulls next element closer)</h3>
  <div class="container">
    <div class="box pull-next" style="display: inline-block; width: 50%;">margin-right: -40px</div>
    <div class="box" style="display: inline-block; width: 50%;">I'm pulled left</div>
  </div>
  
  <h3>4. Full-width breakout from padded container</h3>
  <div class="container">
    <div class="box">Normal width</div>
    <div class="box breakout">Negative margins cancel container padding — I'm full width!</div>
    <div class="box">Normal width</div>
  </div>
  <p class="note">This is a common pattern for full-bleed images inside constrained content areas.</p>
</body>
</html>
```

## Summary

| Concept | Key Point |
|---|---|
| Content box | Where text/children render |
| Padding | Extends background, cannot be negative |
| Border | Visible edge, has own color/style |
| Margin | Transparent, can be negative, can collapse |
| `background-clip` | Controls which box the background paints to |
| `clientWidth` | content + padding (no border) |
| `offsetWidth` | content + padding + border |
| Negative margins | Pull elements toward the negative direction |

## Next

→ [Lesson 02: box-sizing](02-box-sizing.md) — content-box vs border-box
