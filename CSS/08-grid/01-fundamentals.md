# Lesson 01 — Grid Fundamentals

## Creating a Grid

```css
.container {
  display: grid;  /* block-level grid container */
}

/* or */
.container {
  display: inline-grid;  /* inline-level grid container */
}
```

`display: grid` creates a **Grid Formatting Context** (GFC). All direct children become **grid items**.

Like flex items, grid items:
- Are blockified (inline items become block-level)
- Don't have margin collapsing
- `float`, `clear`, `vertical-align` are ignored

## Defining Tracks

```css
.container {
  display: grid;
  grid-template-columns: 200px 1fr 1fr;    /* 3 columns */
  grid-template-rows: 80px auto 50px;       /* 3 rows */
}
```

| Unit | Meaning |
|------|---------|
| `px`, `em`, `rem` | Fixed size |
| `%` | Percentage of grid container |
| `auto` | Fits content (min-content to max-content, stretches if space available) |
| `fr` | Fraction of **remaining space** after fixed tracks are placed |
| `min-content` | Smallest size without overflow |
| `max-content` | Ideal size without wrapping |

## The `fr` Unit

`fr` distributes **leftover space** after fixed tracks, gaps, and padding:

```css
grid-template-columns: 200px 1fr 2fr;
/* Step 1: 200px is reserved
   Step 2: Remaining space split into 3 parts (1fr + 2fr)
   Step 3: Column 2 gets 1/3, Column 3 gets 2/3 */
```

**Key**: `fr` is NOT a percentage. It distributes **remaining** space, not total space.

## `repeat()`

```css
/* These are equivalent: */
grid-template-columns: 1fr 1fr 1fr 1fr 1fr 1fr;
grid-template-columns: repeat(6, 1fr);

/* Mixed: */
grid-template-columns: 200px repeat(4, 1fr) 200px;

/* Auto-fill: creates as many tracks as fit */
grid-template-columns: repeat(auto-fill, 200px);

/* Auto-fill with minmax: responsive grid without media queries */
grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
```

## `minmax()`

```css
grid-template-columns: minmax(200px, 1fr) 3fr;
/* Column 1: at least 200px, but can grow up to 1fr of remaining */

grid-template-rows: minmax(100px, auto);
/* Row: at least 100px, grows with content */
```

## Line-Based Placement

Grid lines are numbered starting at **1**:

```
  1    2      3     4
  |----|----- |-----|
  | C1 |  C2  | C3  |   ← Row 1 (between line 1 and 2)
  |----|----- |-----|
                         ← Row 2 (between line 2 and 3)
  |----|----- |-----|
```

```css
.item {
  grid-column: 1 / 3;     /* Start at line 1, end at line 3 (spans 2 columns) */
  grid-row: 1 / 2;        /* Start at line 1, end at line 2 (spans 1 row) */
}

/* Shorthand: */
.item {
  grid-column: 1 / span 2;  /* Start at 1, span 2 tracks */
}

/* Negative lines count from the end: */
.item {
  grid-column: 1 / -1;    /* Full width (line 1 to last line) */
}
```

## Named Lines

```css
.container {
  display: grid;
  grid-template-columns: [sidebar-start] 250px [sidebar-end content-start] 1fr [content-end];
  grid-template-rows: [header-start] 80px [header-end main-start] 1fr [main-end footer-start] 60px [footer-end];
}

.sidebar {
  grid-column: sidebar-start / sidebar-end;
  grid-row: main-start / main-end;
}
```

## Experiment: Building a Grid

```html
<!-- 01-grid-fundamentals.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Grid Fundamentals</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .grid {
      display: grid;
      gap: 5px;
      background: #e0e0e0;
      padding: 5px;
      margin-bottom: 30px;
    }
    
    .cell {
      background: lightblue;
      border: 2px solid steelblue;
      padding: 15px;
      font-family: monospace;
      font-size: 12px;
      text-align: center;
    }
    
    .cell:nth-child(even) { background: lightyellow; border-color: goldenrod; }
    .label { font-family: monospace; font-size: 13px; margin-bottom: 5px; }

    /* Grid 1: Basic 3-column */
    #grid1 { grid-template-columns: 200px 1fr 1fr; grid-template-rows: 80px 80px; }
    
    /* Grid 2: repeat + fr */
    #grid2 { grid-template-columns: repeat(4, 1fr); }
    
    /* Grid 3: minmax + auto-fill (responsive!) */
    #grid3 { grid-template-columns: repeat(auto-fill, minmax(150px, 1fr)); }
    
    /* Grid 4: Line-based placement */
    #grid4 { grid-template-columns: repeat(3, 1fr); grid-template-rows: repeat(3, 80px); }
    #grid4 .span-full { grid-column: 1 / -1; background: lightcoral; border-color: darkred; }
    #grid4 .span-two { grid-column: span 2; background: lightgreen; border-color: darkgreen; }
  </style>
</head>
<body>
  <h2>Grid Fundamentals</h2>
  
  <div class="label">200px | 1fr | 1fr (resize browser to see fr adapt)</div>
  <div class="grid" id="grid1">
    <div class="cell">200px (fixed)</div>
    <div class="cell">1fr</div>
    <div class="cell">1fr</div>
    <div class="cell">200px (fixed)</div>
    <div class="cell">1fr</div>
    <div class="cell">1fr</div>
  </div>

  <div class="label">repeat(4, 1fr) — four equal columns</div>
  <div class="grid" id="grid2">
    <div class="cell">1</div>
    <div class="cell">2</div>
    <div class="cell">3</div>
    <div class="cell">4</div>
    <div class="cell">5</div>
    <div class="cell">6</div>
    <div class="cell">7</div>
    <div class="cell">8</div>
  </div>
  
  <div class="label">repeat(auto-fill, minmax(150px, 1fr)) — responsive (resize!)</div>
  <div class="grid" id="grid3">
    <div class="cell">A</div>
    <div class="cell">B</div>
    <div class="cell">C</div>
    <div class="cell">D</div>
    <div class="cell">E</div>
    <div class="cell">F</div>
  </div>
  
  <div class="label">Line-based placement: spanning columns</div>
  <div class="grid" id="grid4">
    <div class="cell span-full">grid-column: 1 / -1 (full width)</div>
    <div class="cell span-two">grid-column: span 2</div>
    <div class="cell">Normal</div>
    <div class="cell">Normal</div>
    <div class="cell">Normal</div>
    <div class="cell">Normal</div>
  </div>
</body>
</html>
```

## DevTools Exercise

1. Open the experiment in Chrome
2. Elements panel → click the **grid** badge on the container
3. In the **Layout** panel, enable:
   - Show track sizes
   - Show line numbers
   - Show line names (if using named lines)
   - Show area names
4. Hover over grid items → see which tracks they occupy
5. Observe how `auto-fill` + `minmax` creates/removes columns as you resize

## Next

→ [Lesson 02: The Track Sizing Algorithm](02-track-sizing.md)
