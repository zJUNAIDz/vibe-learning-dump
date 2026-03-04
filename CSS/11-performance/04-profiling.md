# Lesson 04 — Profiling & Measurement

## Core Web Vitals

Google's metrics for user experience:

| Metric | What It Measures | CSS Impact | Good |
|--------|-----------------|------------|------|
| **LCP** (Largest Contentful Paint) | When the largest visible element renders | Render-blocking CSS, font loading | < 2.5s |
| **CLS** (Cumulative Layout Shift) | Visual stability — how much content shifts | Missing dimensions, font swap, dynamic content | < 0.1 |
| **INP** (Interaction to Next Paint) | Responsiveness to input | Layout thrashing, expensive style recalc | < 200ms |

## Preventing CLS (Cumulative Layout Shift)

```css
/* ❌ CLS-causing: image loads → pushes content down */
img { width: 100%; }

/* ✅ Reserve space with aspect-ratio: */
img {
  width: 100%;
  aspect-ratio: 16 / 9;
  object-fit: cover;
}

/* ✅ Or use width + height attributes (browser calculates ratio): */
/* <img src="photo.jpg" width="800" height="450"> */

/* ❌ CLS-causing: font swap changes line heights */
body { font-family: 'CustomFont', sans-serif; }

/* ✅ Reduce swap impact with size-adjust: */
@font-face {
  font-family: 'CustomFont';
  src: url('custom.woff2') format('woff2');
  font-display: swap;
  size-adjust: 105%;          /* match fallback metrics */
  ascent-override: 90%;
  descent-override: 20%;
  line-gap-override: 0%;
}
```

## DevTools Performance Panel

### Recording a Profile

1. DevTools → **Performance** tab
2. Click ⏺️ Record (or `Ctrl+E`)
3. Interact with the page
4. Stop recording

### Reading the Flame Chart

| Color | Phase | Meaning |
|-------|-------|---------|
| 🟡 Yellow | **Scripting** | JavaScript execution |
| 🟣 Purple | **Rendering** | Style recalc + Layout |
| 🟢 Green | **Painting** | Paint + Composite |
| ⚪ Gray | **System** | Browser internal work |

### What to Look For

```
Frame  ─────────────────────────────────
  ├─ Recalculate Style  (purple)      ← How many elements?
  ├─ Layout              (purple)      ← How often? Forced?
  ├─ Paint               (green)       ← How large an area?
  └─ Composite Layers    (green)       ← How many layers?
```

**Warning triangles** in the profile indicate:
- "Forced reflow" — layout thrashing
- "Long task" — > 50ms, blocks main thread

## DevTools Rendering Tab

Access: DevTools → `⋮` → More tools → **Rendering**

| Option | What It Shows |
|--------|-------------|
| **Paint flashing** | Green flash on repainted areas |
| **Layout shift regions** | Blue flash on CLS events |
| **Layer borders** | Orange borders around compositor layers |
| **FPS meter** | Real-time frame rate + GPU memory |
| **Scrolling performance issues** | Highlights scroll-blocking listeners |

## Performance Checklist

### Layout

- [ ] No layout thrashing (batched reads/writes)
- [ ] `contain: content` on isolated widgets
- [ ] `content-visibility: auto` on off-screen sections
- [ ] Images have `width` and `height` attributes

### Paint

- [ ] Animations use only `transform` and `opacity`
- [ ] No unnecessary `box-shadow` blur on many elements
- [ ] `will-change` applied only when needed, removed after

### Loading

- [ ] Critical CSS inlined (< 14KB)
- [ ] Non-critical CSS loaded async
- [ ] Fonts use `font-display: swap` or `optional`
- [ ] Critical fonts preloaded
- [ ] Unused CSS removed (< 10% waste)

### CLS

- [ ] All images/videos have dimensions
- [ ] Font swap uses `size-adjust` to match fallback
- [ ] Dynamic content reserves space
- [ ] No DOM insertion above visible content

## Experiment: Performance Audit

```html
<!-- 04-audit.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Performance Audit Exercise</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .checklist {
      max-width: 600px;
      font-size: 14px;
    }
    
    .check-item {
      padding: 12px 15px;
      border: 1px solid #ddd;
      margin-bottom: -1px;
      display: flex;
      align-items: center;
      gap: 10px;
      cursor: pointer;
    }
    
    .check-item:hover { background: #f9f9f9; }
    
    .check-item input[type="checkbox"] {
      width: 18px;
      height: 18px;
      accent-color: green;
    }
    
    .check-item label { cursor: pointer; flex: 1; }
    
    .category {
      background: #333;
      color: white;
      padding: 10px 15px;
      font-weight: bold;
      font-size: 13px;
      text-transform: uppercase;
      letter-spacing: 1px;
    }
    
    .score {
      font-family: monospace;
      font-size: 24px;
      margin-top: 20px;
      padding: 15px;
      border-radius: 8px;
      text-align: center;
    }
  </style>
</head>
<body>
  <h2>CSS Performance Audit Checklist</h2>
  <p style="color: #666; font-size: 14px;">
    Run this against your own project. Check each item you've verified.
  </p>
  
  <div class="checklist" id="checklist">
    <div class="category">Loading</div>
    <div class="check-item"><input type="checkbox" id="c1"><label for="c1">Critical CSS is inlined or under 14KB</label></div>
    <div class="check-item"><input type="checkbox" id="c2"><label for="c2">Non-critical CSS loaded async</label></div>
    <div class="check-item"><input type="checkbox" id="c3"><label for="c3">Unused CSS < 10% (Coverage tool)</label></div>
    <div class="check-item"><input type="checkbox" id="c4"><label for="c4">Fonts use font-display: swap/optional</label></div>
    <div class="check-item"><input type="checkbox" id="c5"><label for="c5">Critical fonts preloaded</label></div>
    
    <div class="category">Layout</div>
    <div class="check-item"><input type="checkbox" id="c6"><label for="c6">No layout thrashing in JS</label></div>
    <div class="check-item"><input type="checkbox" id="c7"><label for="c7">contain: content on isolated components</label></div>
    <div class="check-item"><input type="checkbox" id="c8"><label for="c8">content-visibility: auto on long pages</label></div>
    
    <div class="category">Paint & Composite</div>
    <div class="check-item"><input type="checkbox" id="c9"><label for="c9">Animations use transform/opacity only</label></div>
    <div class="check-item"><input type="checkbox" id="c10"><label for="c10">will-change used sparingly, removed after</label></div>
    <div class="check-item"><input type="checkbox" id="c11"><label for="c11">No excessive compositor layers (< 30)</label></div>
    
    <div class="category">CLS</div>
    <div class="check-item"><input type="checkbox" id="c12"><label for="c12">All images have width/height or aspect-ratio</label></div>
    <div class="check-item"><input type="checkbox" id="c13"><label for="c13">Font swap uses size-adjust</label></div>
    <div class="check-item"><input type="checkbox" id="c14"><label for="c14">No DOM insertion above fold</label></div>
  </div>
  
  <div class="score" id="score"></div>

  <script>
    const checks = document.querySelectorAll('input[type="checkbox"]');
    const scoreEl = document.getElementById('score');
    
    function updateScore() {
      const checked = document.querySelectorAll('input:checked').length;
      const total = checks.length;
      const percent = Math.round((checked / total) * 100);
      const color = percent >= 80 ? '#ccffcc' : percent >= 50 ? '#fff3cd' : '#ffcccc';
      scoreEl.style.background = color;
      scoreEl.textContent = `${checked}/${total} (${percent}%)`;
    }
    
    checks.forEach(c => c.addEventListener('change', updateScore));
    updateScore();
  </script>
</body>
</html>
```

## Summary

- Core Web Vitals: LCP (loading), CLS (stability), INP (responsiveness)
- CLS prevention: reserve space for images, control font swap
- Performance panel: look for forced layouts, long tasks, excessive paints
- Rendering tab: paint flashing, layer borders, FPS meter
- Apply the performance checklist to every production project

## Next Module

→ [Module 12: Architecture](../12-architecture/README.md)
