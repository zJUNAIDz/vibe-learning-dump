# Lesson 04 — Color & Functions

## Modern Color Spaces

### The Problem with RGB/HSL

`rgb()` and `hsl()` use the **sRGB** color space, which:
- Only covers ~35% of colors modern displays can show
- Has non-perceptual lightness (HSL `lightness: 50%` looks different across hues)
- Makes palette generation inconsistent

### oklch() — Perceptually Uniform Color

```css
/* oklch(lightness chroma hue) */
.button {
  background: oklch(55% 0.2 260);
  /*          L: 0-100%
              C: 0-0.4 (colorfulness)
              H: 0-360 (hue angle) */
}
```

**Why oklch is better:**

| Property | HSL | oklch |
|----------|-----|-------|
| Perceptual lightness | ❌ Different hues look different at same L | ✅ Same L = same perceived brightness |
| Gamut | sRGB only | P3 and beyond |
| Palette generation | Inconsistent | Consistent — vary H, keep L and C |

### Generating a Consistent Palette

```css
:root {
  /* Same lightness and chroma, different hues → equally vibrant colors */
  --red:    oklch(65% 0.2 25);
  --orange: oklch(65% 0.2 55);
  --yellow: oklch(65% 0.2 90);
  --green:  oklch(65% 0.2 145);
  --blue:   oklch(65% 0.2 260);
  --purple: oklch(65% 0.2 300);

  /* Shades — same hue and chroma, vary lightness */
  --blue-100: oklch(95% 0.03 260);
  --blue-200: oklch(85% 0.08 260);
  --blue-300: oklch(75% 0.12 260);
  --blue-400: oklch(65% 0.16 260);
  --blue-500: oklch(55% 0.2  260);
  --blue-600: oklch(45% 0.2  260);
  --blue-700: oklch(35% 0.18 260);
  --blue-800: oklch(25% 0.14 260);
  --blue-900: oklch(15% 0.1  260);
}
```

## color-mix()

Mix two colors in any color space:

```css
.button {
  background: var(--primary);
}
.button:hover {
  /* Mix with black for darkening */
  background: color-mix(in oklch, var(--primary), black 20%);
}
.button:active {
  background: color-mix(in oklch, var(--primary), black 35%);
}

/* Semi-transparent version */
.badge {
  background: color-mix(in srgb, var(--primary) 15%, transparent);
}

/* Mix two brand colors */
.accent {
  color: color-mix(in oklch, var(--blue), var(--purple) 40%);
}
```

### Color Space Matters

```css
/* Mixing in different spaces produces different results */
.srgb   { background: color-mix(in srgb, blue, yellow); }
.oklch  { background: color-mix(in oklch, blue, yellow); }
.oklab  { background: color-mix(in oklab, blue, yellow); }
```

`oklch` and `oklab` produce more natural-looking mixes because they interpolate perceptually.

## Relative Color Syntax

Transform a color by modifying its components:

```css
:root {
  --brand: #2563eb;
}

.light {
  /* Take --brand, modify its lightness in oklch */
  background: oklch(from var(--brand) calc(l + 0.3) c h);
}

.desaturated {
  /* Reduce chroma */
  color: oklch(from var(--brand) l calc(c * 0.5) h);
}

.complement {
  /* Rotate hue by 180° */
  color: oklch(from var(--brand) l c calc(h + 180));
}
```

> **Note:** Relative color syntax is Baseline 2024. Check support for your targets.

## Trigonometric Functions

CSS now has `sin()`, `cos()`, `tan()`, `asin()`, `acos()`, `atan()`, `atan2()`:

```css
/* Arrange items in a circle */
.circle-layout {
  position: relative;
  width: 300px;
  height: 300px;
}

.circle-layout .item {
  --total: 8;
  --radius: 120px;
  --angle: calc(360deg / var(--total) * var(--i));

  position: absolute;
  top: 50%;
  left: 50%;
  translate: 
    calc(cos(var(--angle)) * var(--radius) - 50%)
    calc(sin(var(--angle)) * var(--radius) - 50%);
}
```

```html
<div class="circle-layout">
  <div class="item" style="--i: 0">1</div>
  <div class="item" style="--i: 1">2</div>
  <div class="item" style="--i: 2">3</div>
  <!-- ... -->
</div>
```

## round(), mod(), rem()

```css
/* Round to nearest grid increment */
.container {
  width: round(100vw, 8px);        /* Round viewport width to nearest 8px */
}

/* Snap font size to pixel grid */
.text {
  font-size: round(1.333rem, 1px); /* Avoid sub-pixel rendering issues */
}

/* Cycle through values */
.item {
  --hue: mod(calc(var(--i) * 45), 360);
  background: oklch(65% 0.2 var(--hue));
}
```

## New Units

### Container Query Units

```css
.card {
  container-type: inline-size;
}

.card .title {
  /* Size relative to container, not viewport */
  font-size: clamp(14px, 4cqi, 24px);
  /* cqi = 1% of container's inline size */
}
```

| Unit | Relative To |
|------|------------|
| `cqw` | Container width |
| `cqh` | Container height |
| `cqi` | Container inline size |
| `cqb` | Container block size |
| `cqmin` | Smaller of cqi/cqb |
| `cqmax` | Larger of cqi/cqb |

### Dynamic Viewport Units

```css
/* Standard vh doesn't account for mobile browser chrome */
.hero {
  height: 100vh;    /* May be taller than visible area on mobile */
}

/* Dynamic viewport units adapt to browser chrome */
.hero {
  height: 100dvh;   /* Always exactly the visible height */
}
```

| Unit | Behavior |
|------|----------|
| `svh` / `svw` | Small viewport — browser chrome fully visible (smallest) |
| `lvh` / `lvw` | Large viewport — browser chrome retracted (largest) |
| `dvh` / `dvw` | Dynamic — adjusts as chrome shows/hides |

**Recommendation:** Use `dvh` for full-viewport elements on mobile.

## Experiment — oklch Palette Generator

```html
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>oklch Palette Generator</title>
<style>
* { margin: 0; box-sizing: border-box; }
body {
  font-family: system-ui;
  padding: 2rem;
  background: #0f172a;
  color: #e2e8f0;
  min-height: 100vh;
}

h1 { margin-bottom: 0.5rem; }
.subtitle { color: #94a3b8; margin-bottom: 2rem; }

.controls {
  display: flex;
  gap: 2rem;
  flex-wrap: wrap;
  margin-bottom: 2rem;
}

.control {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.control label {
  font-size: 13px;
  color: #94a3b8;
}

.control input[type="range"] {
  width: 200px;
}

.control span {
  font-size: 12px;
  font-family: monospace;
  color: #64748b;
}

.palette {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 8px;
  margin-bottom: 2rem;
}

.swatch {
  aspect-ratio: 1;
  border-radius: 12px;
  display: grid;
  place-items: center;
  font-size: 12px;
  font-family: monospace;
  font-weight: bold;
  transition: transform 0.2s;
  cursor: pointer;
}
.swatch:hover { transform: scale(1.05); }

.comparison {
  margin-top: 2rem;
  padding: 1.5rem;
  background: #1e293b;
  border-radius: 12px;
}
.comparison h3 { margin-bottom: 1rem; font-size: 16px; }
.comparison .row {
  display: flex;
  gap: 4px;
  margin-bottom: 8px;
}
.comparison .row .swatch {
  flex: 1;
  aspect-ratio: 2;
  border-radius: 6px;
  font-size: 10px;
}
.comparison .label {
  font-size: 12px;
  color: #94a3b8;
  margin-bottom: 4px;
}
</style>
</head>
<body>
  <h1>oklch Palette Generator</h1>
  <p class="subtitle">Perceptually uniform — same lightness looks equally bright across all hues</p>

  <div class="controls">
    <div class="control">
      <label>Lightness (L)</label>
      <input type="range" id="lightness" min="20" max="90" value="65">
      <span id="l-val">65%</span>
    </div>
    <div class="control">
      <label>Chroma (C)</label>
      <input type="range" id="chroma" min="0" max="35" value="20">
      <span id="c-val">0.20</span>
    </div>
    <div class="control">
      <label>Steps</label>
      <input type="range" id="steps" min="4" max="16" value="8">
      <span id="s-val">8</span>
    </div>
  </div>

  <div class="palette" id="palette"></div>

  <div class="comparison">
    <h3>HSL vs oklch — Same Lightness Comparison</h3>
    <p class="label">HSL (50% lightness across hues — notice brightness inconsistency):</p>
    <div class="row" id="hsl-row"></div>
    <p class="label">oklch (65% lightness across hues — perceptually uniform):</p>
    <div class="row" id="oklch-row"></div>
  </div>

  <script>
    const lightnessEl = document.getElementById('lightness');
    const chromaEl = document.getElementById('chroma');
    const stepsEl = document.getElementById('steps');

    function render() {
      const L = parseInt(lightnessEl.value);
      const C = parseInt(chromaEl.value) / 100;
      const steps = parseInt(stepsEl.value);

      document.getElementById('l-val').textContent = L + '%';
      document.getElementById('c-val').textContent = C.toFixed(2);
      document.getElementById('s-val').textContent = steps;

      // Generate palette
      const palette = document.getElementById('palette');
      palette.innerHTML = '';
      for (let i = 0; i < steps; i++) {
        const hue = Math.round((360 / steps) * i);
        const swatch = document.createElement('div');
        swatch.className = 'swatch';
        swatch.style.background = `oklch(${L}% ${C} ${hue})`;
        swatch.style.color = L > 55 ? 'black' : 'white';
        swatch.textContent = `${hue}°`;
        swatch.title = `oklch(${L}% ${C} ${hue})`;
        palette.appendChild(swatch);
      }

      // Generate comparison rows
      const hslRow = document.getElementById('hsl-row');
      const oklchRow = document.getElementById('oklch-row');
      hslRow.innerHTML = '';
      oklchRow.innerHTML = '';

      for (let h = 0; h < 360; h += 30) {
        const hslSwatch = document.createElement('div');
        hslSwatch.className = 'swatch';
        hslSwatch.style.background = `hsl(${h} 100% 50%)`;
        hslSwatch.textContent = h + '°';
        hslSwatch.style.color = 'white';
        hslSwatch.style.textShadow = '0 1px 2px rgba(0,0,0,0.5)';
        hslRow.appendChild(hslSwatch);

        const oklchSwatch = document.createElement('div');
        oklchSwatch.className = 'swatch';
        oklchSwatch.style.background = `oklch(65% 0.2 ${h})`;
        oklchSwatch.textContent = h + '°';
        oklchSwatch.style.color = 'white';
        oklchSwatch.style.textShadow = '0 1px 2px rgba(0,0,0,0.5)';
        oklchRow.appendChild(oklchSwatch);
      }
    }

    lightnessEl.addEventListener('input', render);
    chromaEl.addEventListener('input', render);
    stepsEl.addEventListener('input', render);
    render();
  </script>
</body>
</html>
```

### What to Observe

1. **oklch palette:** All swatches at the same L/C look equally bright regardless of hue
2. **HSL comparison:** Yellow (60°) looks much brighter than blue (240°) at HSL 50% lightness
3. **oklch comparison:** All hues at oklch 65% look perceptually similar in brightness
4. Try reducing chroma to 0 — all swatches become the same gray (proof that L is hue-independent)

---

## Module 14 Summary

You learned:
- **Custom properties** — scoped theming, `@property` for typed animation, performance implications
- **@layer** — explicit cascade ordering, layer precedence > specificity, taming third-party CSS
- **:has()** — parent/context selection without JavaScript
- **:is() / :where()** — compact selectors with controlled specificity
- **oklch** — perceptually uniform color, consistent palette generation
- **color-mix()** — runtime color manipulation without preprocessors
- **New functions** — trig functions, `round()`, `mod()`
- **New units** — container query units, dynamic viewport units

## Next

→ [Module 15: CSS Debugging](../15-debugging/README.md)
