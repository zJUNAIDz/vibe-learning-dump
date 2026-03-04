# Lesson 03 — Aspect Ratio & Intrinsic Sizing

## `aspect-ratio`

Maintains a width-to-height ratio on an element:

```css
.video {
  aspect-ratio: 16 / 9;   /* width:height */
  width: 100%;             /* height auto-calculated */
}

.square {
  aspect-ratio: 1;         /* same as 1 / 1 */
}
```

### How It Interacts with Width/Height

| Set | Set | Result |
|-----|-----|--------|
| `width` | `aspect-ratio` | Height is calculated |
| `height` | `aspect-ratio` | Width is calculated |
| `width` + `height` | `aspect-ratio` | **Both win** — `aspect-ratio` ignored |
| nothing | `aspect-ratio` | Depends on context (flex/grid may stretch) |

`min-height` / `max-height` can override the calculated dimension from `aspect-ratio`.

### Replacing the Padding-Top Hack

```css
/* Old hack for 16:9 responsive container: */
.video-wrapper {
  position: relative;
  padding-top: 56.25%;  /* 9 / 16 = 0.5625 */
}
.video-wrapper > * {
  position: absolute;
  inset: 0;
}

/* Modern: */
.video-wrapper {
  aspect-ratio: 16 / 9;
}
```

## `object-fit` and `object-position`

Controls how **replaced elements** (images, videos) fill their box:

```css
img {
  width: 300px;
  height: 200px;
  object-fit: cover;           /* fill box, crop excess */
  object-position: center top; /* anchor point */
}
```

| `object-fit` | Behavior |
|-------------|----------|
| `fill` (default) | Stretches to fill box (distorts) |
| `contain` | Fits entirely, letterboxes |
| `cover` | Fills entirely, crops |
| `none` | Natural size, clips |
| `scale-down` | Like `contain` if larger, `none` if smaller |

## Intrinsic Sizing Keywords

These keywords can be used as values for `width`, `height`, `min-width`, `max-width`, etc.:

| Keyword | Meaning |
|---------|---------|
| `min-content` | Smallest size without overflow (wraps text aggressively) |
| `max-content` | Ideal size if given infinite space (no wrapping) |
| `fit-content` | `min(max-content, max(min-content, available-space))` — wraps if needed, no wider than content |
| `fit-content(<length>)` | Clamps max-content at the given length |

```css
.tag {
  width: fit-content;  /* Shrink-wraps to content, won't overflow */
  padding: 4px 12px;
  background: lightblue;
}
```

## Experiment: Aspect Ratio & Object-Fit

```html
<!-- 03-aspect-ratio.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Aspect Ratio & Object-Fit</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin-bottom: 30px; }
    
    .box {
      background: linear-gradient(135deg, steelblue, lightblue);
      display: flex;
      align-items: center;
      justify-content: center;
      font-family: monospace;
      font-size: 12px;
      color: white;
      text-shadow: 0 1px 2px rgba(0,0,0,0.5);
    }

    .img-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: 15px; }
    
    .img-box {
      border: 2px solid #ccc;
      border-radius: 8px;
      overflow: hidden;
      text-align: center;
    }
    
    .img-box img {
      width: 100%;
      height: 150px;
      display: block;
      /* Use a gradient as placeholder — works without real image */
      background: linear-gradient(135deg, #e74c3c, #f39c12, #2ecc71, #3498db);
    }
    
    .img-box .caption {
      padding: 8px;
      font-family: monospace;
      font-size: 11px;
      background: #f5f5f5;
    }
    
    .label { font-family: monospace; font-size: 13px; margin-bottom: 8px; margin-top: 20px; }
  </style>
</head>
<body>
  <h2>Aspect Ratio</h2>
  
  <div class="label">Different aspect ratios (width: 100%, colored boxes)</div>
  <div class="grid">
    <div class="box" style="aspect-ratio: 1 / 1;">1:1 (square)</div>
    <div class="box" style="aspect-ratio: 16 / 9;">16:9</div>
    <div class="box" style="aspect-ratio: 4 / 3;">4:3</div>
    <div class="box" style="aspect-ratio: 21 / 9;">21:9 (ultra-wide)</div>
  </div>
  
  <h2>object-fit on Images</h2>
  <div class="label">Same image, 100% width × 150px height, different object-fit</div>
  
  <div class="img-grid">
    <div class="img-box">
      <img src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='200'%3E%3Crect width='400' height='200' fill='%234a90d9'/%3E%3Ccircle cx='200' cy='100' r='70' fill='%23f5a623'/%3E%3Ctext x='200' y='108' text-anchor='middle' fill='white' font-size='24'%3E400×200%3C/text%3E%3C/svg%3E" style="object-fit: fill;" alt="">
      <div class="caption">object-fit: fill (stretches)</div>
    </div>
    <div class="img-box">
      <img src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='200'%3E%3Crect width='400' height='200' fill='%234a90d9'/%3E%3Ccircle cx='200' cy='100' r='70' fill='%23f5a623'/%3E%3Ctext x='200' y='108' text-anchor='middle' fill='white' font-size='24'%3E400×200%3C/text%3E%3C/svg%3E" style="object-fit: contain;" alt="">
      <div class="caption">object-fit: contain (letterbox)</div>
    </div>
    <div class="img-box">
      <img src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='200'%3E%3Crect width='400' height='200' fill='%234a90d9'/%3E%3Ccircle cx='200' cy='100' r='70' fill='%23f5a623'/%3E%3Ctext x='200' y='108' text-anchor='middle' fill='white' font-size='24'%3E400×200%3C/text%3E%3C/svg%3E" style="object-fit: cover;" alt="">
      <div class="caption">object-fit: cover (crops)</div>
    </div>
    <div class="img-box">
      <img src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='200'%3E%3Crect width='400' height='200' fill='%234a90d9'/%3E%3Ccircle cx='200' cy='100' r='70' fill='%23f5a623'/%3E%3Ctext x='200' y='108' text-anchor='middle' fill='white' font-size='24'%3E400×200%3C/text%3E%3C/svg%3E" style="object-fit: none;" alt="">
      <div class="caption">object-fit: none (natural size)</div>
    </div>
    <div class="img-box">
      <img src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='200'%3E%3Crect width='400' height='200' fill='%234a90d9'/%3E%3Ccircle cx='200' cy='100' r='70' fill='%23f5a623'/%3E%3Ctext x='200' y='108' text-anchor='middle' fill='white' font-size='24'%3E400×200%3C/text%3E%3C/svg%3E" style="object-fit: scale-down;" alt="">
      <div class="caption">object-fit: scale-down</div>
    </div>
  </div>

  <h2>Intrinsic Sizing Keywords</h2>
  <div style="max-width: 500px; border: 2px solid #ccc; padding: 15px; margin-top: 10px;">
    <div style="width: min-content; background: #ffcccc; padding: 8px; margin-bottom: 8px; font-family: monospace; font-size: 12px; border: 1px solid red;">
      width: min-content — wraps aggressively
    </div>
    <div style="width: max-content; background: #ccffcc; padding: 8px; margin-bottom: 8px; font-family: monospace; font-size: 12px; border: 1px solid green;">
      width: max-content — never wraps (may overflow container)
    </div>
    <div style="width: fit-content; background: #ccccff; padding: 8px; font-family: monospace; font-size: 12px; border: 1px solid blue;">
      width: fit-content — shrink-wraps, won't exceed available space
    </div>
  </div>
</body>
</html>
```

## Next

→ [Lesson 04: Multi-Column & Other Layout](04-multi-column.md)
