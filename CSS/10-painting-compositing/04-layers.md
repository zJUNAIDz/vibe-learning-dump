# Lesson 04 — Layer Promotion & will-change

## Compositor Layers Recap

The browser organizes painted content into **compositor layers**. Each layer:
- Is rasterized to a texture on the GPU
- Can be moved, rotated, scaled, faded **without repainting**
- Has GPU memory cost

## `will-change`

Hints to the browser that a property will change, prompting layer promotion:

```css
.element {
  will-change: transform;   /* promote to own compositor layer */
  will-change: opacity;
  will-change: transform, opacity;  /* multiple properties */
}
```

### Rules for `will-change`

1. **Don't apply to everything**: `* { will-change: transform; }` creates hundreds of layers → crashes GPU memory
2. **Apply before animation starts**: Set via JS before animation, remove after
3. **Don't set if not needed**: Only use when actual jank is measured
4. **Remove when done**:

```javascript
element.addEventListener('mouseenter', () => {
  element.style.willChange = 'transform';
});

element.addEventListener('transitionend', () => {
  element.style.willChange = 'auto';
});
```

### `will-change` Side Effects

| `will-change` value | Creates |
|--------------------|---------|
| `transform` | Stacking context + containing block + compositor layer |
| `opacity` | Stacking context + compositor layer |
| `filter` | Stacking context + containing block + compositor layer |
| `position` | No layer; hint only |
| `top`, `left`, etc. | No layer; hint only |

## Layer Explosion

Each compositor layer costs GPU memory (~width × height × 4 bytes). Too many layers can:
- Exceed GPU memory → crash on mobile devices
- Slow down composite phase → worse than no layers
- Cause "layer squashing" — browser merges overlapping layers

### Common Causes

```css
/* BAD: 1000 list items each get their own layer */
.list-item {
  will-change: transform;  /* ← 1000 layers! */
}

/* BAD: translateZ(0) hack on many elements */
.card {
  transform: translateZ(0);  /* ← forced layer promotion */
}
```

### How to Check Layer Count

DevTools → More tools → **Layers** panel:
- See all compositor layers
- See memory cost per layer
- Rotate the 3D view to understand stacking
- Click a layer to see why it was promoted

## The `translateZ(0)` and `translate3d(0,0,0)` Hack

Before `will-change` existed, developers forced layer promotion with:

```css
/* Old hack — avoid in modern code: */
.element { transform: translateZ(0); }
.element { transform: translate3d(0, 0, 0); }
```

These force a compositor layer because any 3D transform triggers promotion. **Prefer `will-change` instead** — it's semantically clear and can be removed when not needed.

## `contain` Property

`contain` tells the browser what can be isolated for optimization:

```css
.widget {
  contain: layout;   /* layout of this element is independent */
  contain: paint;    /* nothing paints outside this box */
  contain: size;     /* size doesn't depend on children */
  contain: style;    /* counters/quotes scoped to this subtree */
  contain: content;  /* = layout + paint */
  contain: strict;   /* = layout + paint + size */
}
```

| Value | What It Does | Use Case |
|-------|-------------|----------|
| `layout` | Layout changes inside don't affect outside | Isolated widgets |
| `paint` | Children clipped to element's box, creates stacking context | Off-screen content |
| `size` | Element ignores children for sizing (needs explicit size) | Fixed-size containers |
| `content` | `layout` + `paint` (safest combination) | Most components |
| `strict` | All containment types | Performance-critical fixed-size containers |

### `content-visibility`

Skips rendering of off-screen content entirely:

```css
.section {
  content-visibility: auto;           /* skip rendering if off-screen */
  contain-intrinsic-size: auto 500px; /* estimated height when skipped */
}
```

This can dramatically improve initial render performance for long pages — off-screen sections skip layout, paint, and composite until scrolled into view.

## Experiment: Layer Debugging

```html
<!-- 04-layers.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Layer Promotion</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .container {
      position: relative;
      width: 500px;
      height: 400px;
      background: #f5f5f5;
      border: 2px solid #ccc;
      overflow: hidden;
    }
    
    .layer-box {
      position: absolute;
      width: 150px;
      height: 100px;
      border-radius: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-family: monospace;
      font-size: 11px;
      color: white;
      text-align: center;
    }
    
    .no-layer {
      background: steelblue;
      top: 30px; left: 30px;
    }
    
    .will-change-layer {
      background: #e74c3c;
      top: 80px; left: 120px;
      will-change: transform;
    }
    
    .transform-layer {
      background: #27ae60;
      top: 150px; left: 60px;
      transform: translateZ(0);
    }
    
    .animated-layer {
      background: #8e44ad;
      top: 220px; left: 180px;
      animation: float 2s ease-in-out infinite alternate;
    }
    
    .opacity-layer {
      background: #f39c12;
      top: 100px; left: 300px;
      opacity: 0.8;
      transition: opacity 0.3s;
    }
    .opacity-layer:hover { opacity: 0.3; }
    
    @keyframes float {
      from { transform: translateY(0); }
      to { transform: translateY(-30px); }
    }
    
    .info {
      font-family: monospace;
      font-size: 12px;
      margin-top: 15px;
      padding: 10px;
      background: #fff3cd;
      border-radius: 4px;
      max-width: 500px;
    }
    
    .label { font-family: monospace; font-size: 13px; margin-bottom: 8px; }
  </style>
</head>
<body>
  <h2>Compositor Layer Promotion</h2>
  <div class="label">Open DevTools → More tools → Layers to see which boxes are promoted</div>
  
  <div class="container">
    <div class="layer-box no-layer">No promotion<br>(main layer)</div>
    <div class="layer-box will-change-layer">will-change:<br>transform</div>
    <div class="layer-box transform-layer">translateZ(0)<br>(hack)</div>
    <div class="layer-box animated-layer">Animated<br>(auto-promoted)</div>
    <div class="layer-box opacity-layer">opacity: 0.8<br>(hover me)</div>
  </div>
  
  <div class="info">
    <strong>DevTools checklist:</strong><br>
    1. Layers panel → count layers<br>
    2. Click each layer → see "Compositing Reasons"<br>
    3. Check paint count (each layer shows repaint count)<br>
    4. Hover the opacity box → watch for paint flashing<br>
    5. Memory: each layer shows its size in bytes
  </div>
  
  <h2 style="margin-top: 30px;">contain: content</h2>
  <div style="display: flex; gap: 20px;">
    <div style="width: 200px; padding: 15px; border: 2px solid #ccc; border-radius: 8px;">
      <div class="label" style="color: red;">Without contain</div>
      <div style="background: #ffcccc; padding: 10px; font-size: 13px;">
        Changing content here could affect layout of siblings.
      </div>
    </div>
    <div style="width: 200px; padding: 15px; border: 2px solid #ccc; border-radius: 8px; contain: content;">
      <div class="label" style="color: green;">contain: content</div>
      <div style="background: #ccffcc; padding: 10px; font-size: 13px;">
        Layout changes here are isolated. Browser can skip re-layout of siblings.
      </div>
    </div>
  </div>
</body>
</html>
```

## Summary

| Concept | Key Point |
|---------|-----------|
| Compositor layers | GPU-cached textures that can be transformed/faded cheaply |
| `will-change` | Explicit promotion hint — use sparingly, remove when done |
| `translateZ(0)` | Old hack for forced promotion — prefer `will-change` |
| Layer explosion | Too many layers → GPU memory pressure → jank |
| `contain` | Tells browser what's independent for optimization |
| `content-visibility: auto` | Skip rendering off-screen content entirely |

## Next Module

→ [Module 11: Performance](../11-performance/README.md)
