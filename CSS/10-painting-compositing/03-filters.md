# Lesson 03 — Filters & Effects

## CSS `filter`

Applies visual effects to an element (including its children):

```css
.element {
  filter: blur(5px);                    /* Gaussian blur */
  filter: brightness(1.2);             /* > 1 brighter, < 1 darker */
  filter: contrast(1.5);               /* > 1 more contrast */
  filter: grayscale(100%);             /* 0% = normal, 100% = gray */
  filter: saturate(2);                 /* > 1 more saturated */
  filter: sepia(80%);                  /* vintage effect */
  filter: hue-rotate(90deg);           /* shift hue */
  filter: invert(100%);               /* invert colors */
  filter: opacity(0.5);               /* like opacity but composited differently */
  filter: drop-shadow(4px 4px 8px rgba(0,0,0,0.3)); /* follows shape, not box */
  
  /* Chained: */
  filter: brightness(1.1) contrast(1.2) saturate(1.3);
}
```

### `filter` vs `opacity`

| | `filter: opacity()` | `opacity` |
|-|---------------------|-----------|
| Can be chained with other filters | Yes | No |
| GPU-composited independently | Sometimes | Yes (compositor layer) |
| Creates stacking context | Yes | Yes (when < 1) |

### `drop-shadow` vs `box-shadow`

```css
/* box-shadow: follows the RECTANGULAR box */
.element { box-shadow: 4px 4px 8px rgba(0,0,0,0.3); }

/* drop-shadow: follows the VISUAL shape (including transparency) */
.element { filter: drop-shadow(4px 4px 8px rgba(0,0,0,0.3)); }
```

`drop-shadow` respects `border-radius`, clipped shapes (via `clip-path`), and image transparency. `box-shadow` always shadows the rectangular box.

## `backdrop-filter`

Applies filters to **what's behind** the element (not the element itself):

```css
.frosted-glass {
  background: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(10px) saturate(1.5);
}
```

Common for frosted glass effects, overlays, and sticky headers.

**Gotcha**: `backdrop-filter` requires a **semi-transparent background** to see through. `background: white` would hide the backdrop effect.

## Blend Modes

Control how an element's colors combine with what's behind it:

```css
/* Element blend mode: */
.element {
  mix-blend-mode: multiply;    /* multiply with background */
}

/* Background blend mode: */
.element {
  background-image: url('texture.jpg'), linear-gradient(red, blue);
  background-blend-mode: overlay;  /* blend between background layers */
}
```

| Mode | Effect |
|------|--------|
| `normal` | Default (no blending) |
| `multiply` | Darken (colors multiplied) |
| `screen` | Lighten (inverse multiply) |
| `overlay` | Multiply darks, screen lights |
| `darken` | Keep darker pixel |
| `lighten` | Keep lighter pixel |
| `color-dodge` | Bright, high contrast |
| `color-burn` | Dark, high contrast |
| `difference` | Absolute difference |
| `exclusion` | Softer difference |

### `isolation: isolate`

Prevents `mix-blend-mode` from blending with ancestors beyond the isolation boundary:

```css
.card {
  isolation: isolate;  /* blend modes inside won't affect anything outside */
}
```

## CSS Masks

Masks hide parts of an element based on an image's luminance or alpha:

```css
.element {
  mask-image: linear-gradient(to bottom, black, transparent);  /* fade out */
  mask-image: url('mask.svg');
  mask-size: cover;
  mask-repeat: no-repeat;
  -webkit-mask-image: linear-gradient(to bottom, black, transparent); /* Safari */
}
```

## Side Effects Table

| Property | Creates Stacking Context | Creates Containing Block | Compositor Layer |
|----------|------------------------|-------------------------|-----------------|
| `filter` | ✅ | ✅ (for abs/fixed) | ✅ |
| `backdrop-filter` | ✅ | ✅ | ✅ |
| `mix-blend-mode` | ✅ | ❌ | Sometimes |
| `clip-path` | ✅ | ❌ | Sometimes |
| `mask` | ✅ | ❌ | ✅ |
| `opacity < 1` | ✅ | ❌ | ✅ |

## Experiment: Filters & Effects

```html
<!-- 03-filters.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Filters & Effects</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .filter-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
      gap: 15px;
      margin-bottom: 30px;
    }
    
    .filter-item {
      text-align: center;
    }
    
    .filter-box {
      width: 100%;
      height: 100px;
      background: linear-gradient(135deg, #e74c3c, #3498db, #2ecc71);
      border-radius: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      color: white;
      font-size: 24px;
      font-weight: bold;
      text-shadow: 0 1px 3px rgba(0,0,0,0.5);
    }
    
    .filter-caption {
      font-family: monospace;
      font-size: 11px;
      margin-top: 5px;
    }
    
    .glass-demo {
      background: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='200' height='200'%3E%3Cdefs%3E%3ClinearGradient id='g' x1='0' y1='0' x2='1' y2='1'%3E%3Cstop offset='0%25' stop-color='%23e74c3c'/%3E%3Cstop offset='50%25' stop-color='%233498db'/%3E%3Cstop offset='100%25' stop-color='%232ecc71'/%3E%3C/linearGradient%3E%3C/defs%3E%3Crect width='200' height='200' fill='url(%23g)'/%3E%3Ccircle cx='60' cy='60' r='30' fill='%23f39c12'/%3E%3Ccircle cx='140' cy='100' r='40' fill='%238e44ad'/%3E%3Ccircle cx='80' cy='150' r='25' fill='%231abc9c'/%3E%3C/svg%3E");
      background-size: cover;
      padding: 40px;
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-bottom: 30px;
      height: 200px;
    }
    
    .glass-card {
      background: rgba(255, 255, 255, 0.15);
      backdrop-filter: blur(12px) saturate(1.5);
      -webkit-backdrop-filter: blur(12px) saturate(1.5);
      border: 1px solid rgba(255, 255, 255, 0.3);
      border-radius: 16px;
      padding: 25px 40px;
      color: white;
      text-align: center;
      font-family: monospace;
    }
    
    .shadow-demo { display: flex; gap: 40px; margin-bottom: 30px; align-items: center; }
    .shadow-box {
      width: 0; height: 0;
      border-left: 50px solid transparent;
      border-right: 50px solid transparent;
      border-bottom: 80px solid steelblue;
    }
    
    .label { font-family: monospace; font-size: 13px; margin-bottom: 8px; margin-top: 20px; }
  </style>
</head>
<body>
  <h2>CSS Filters</h2>
  <div class="filter-grid">
    <div class="filter-item">
      <div class="filter-box">CSS</div>
      <div class="filter-caption">none (original)</div>
    </div>
    <div class="filter-item">
      <div class="filter-box" style="filter: blur(3px);">CSS</div>
      <div class="filter-caption">blur(3px)</div>
    </div>
    <div class="filter-item">
      <div class="filter-box" style="filter: grayscale(100%);">CSS</div>
      <div class="filter-caption">grayscale(100%)</div>
    </div>
    <div class="filter-item">
      <div class="filter-box" style="filter: sepia(100%);">CSS</div>
      <div class="filter-caption">sepia(100%)</div>
    </div>
    <div class="filter-item">
      <div class="filter-box" style="filter: brightness(1.5);">CSS</div>
      <div class="filter-caption">brightness(1.5)</div>
    </div>
    <div class="filter-item">
      <div class="filter-box" style="filter: contrast(2);">CSS</div>
      <div class="filter-caption">contrast(2)</div>
    </div>
    <div class="filter-item">
      <div class="filter-box" style="filter: hue-rotate(90deg);">CSS</div>
      <div class="filter-caption">hue-rotate(90deg)</div>
    </div>
    <div class="filter-item">
      <div class="filter-box" style="filter: invert(100%);">CSS</div>
      <div class="filter-caption">invert(100%)</div>
    </div>
    <div class="filter-item">
      <div class="filter-box" style="filter: saturate(3);">CSS</div>
      <div class="filter-caption">saturate(3)</div>
    </div>
  </div>
  
  <h2>backdrop-filter: Frosted Glass</h2>
  <div class="glass-demo">
    <div class="glass-card">
      backdrop-filter: blur(12px)<br>
      + saturate(1.5)
    </div>
  </div>
  
  <h2>drop-shadow vs box-shadow (on a triangle)</h2>
  <div class="shadow-demo">
    <div>
      <div class="shadow-box" style="box-shadow: 5px 5px 10px rgba(0,0,0,0.4);"></div>
      <div class="label">box-shadow (boxes rectangle)</div>
    </div>
    <div>
      <div class="shadow-box" style="filter: drop-shadow(5px 5px 10px rgba(0,0,0,0.4));"></div>
      <div class="label">drop-shadow (follows shape)</div>
    </div>
  </div>
</body>
</html>
```

## Next

→ [Lesson 04: Layer Promotion & will-change](04-layers.md)
