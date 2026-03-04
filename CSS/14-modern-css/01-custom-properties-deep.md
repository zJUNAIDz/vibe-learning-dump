# Lesson 01 — Custom Properties Deep Dive

## Beyond Simple Variables

You already know the basics (`--color: blue; var(--color)`). This lesson covers advanced patterns that make custom properties a powerful runtime system.

## Scope and Inheritance

Custom properties follow the **cascade** and **inherit** by default:

```css
:root {
  --text: #333;           /* Available everywhere */
}

.card {
  --text: #666;           /* Overrides for .card and descendants */
}

.card .title {
  color: var(--text);     /* Gets #666 from .card */
}

.footer {
  color: var(--text);     /* Gets #333 from :root */
}
```

### Scoped Component Theming

```css
/* Component defines its API via custom properties */
.button {
  --btn-bg: #2563eb;
  --btn-color: white;
  --btn-radius: 4px;
  --btn-padding: 8px 16px;

  background: var(--btn-bg);
  color: var(--btn-color);
  border-radius: var(--btn-radius);
  padding: var(--btn-padding);
  border: none;
  cursor: pointer;
}

/* Variants override the properties */
.button.danger {
  --btn-bg: #dc2626;
}

.button.ghost {
  --btn-bg: transparent;
  --btn-color: #2563eb;
}

/* Context can override too */
.dark-theme .button {
  --btn-bg: #3b82f6;
}
```

This pattern creates a **public API** for component styling without touching internals.

## Fallback Values

```css
.card {
  /* Simple fallback */
  color: var(--card-color, #333);

  /* Nested fallback */
  background: var(--card-bg, var(--surface-bg, white));

  /* Fallback is only used when the property is NOT SET (not invalid) */
}
```

**Important:** A custom property set to `initial` is different from not being set:

```css
:root {
  --color: initial;        /* SET to CSS-wide keyword */
}
.text {
  color: var(--color, red); /* Does NOT use fallback — --color IS set */
  /* Results in: color: initial → inherited or default */
}
```

## The Invalid at Computed Value Time Problem

```css
.box {
  --size: blue;                  /* Valid custom property (any value is valid) */
  width: var(--size);            /* "blue" is invalid for width */
  /* Result: width becomes "initial", NOT the fallback */
}
```

Custom properties accept **any** value during parse time. Validation only happens at computed-value time. If invalid, the property resets to its **inherited or initial value** — not the fallback in `var()`.

This is the **Guaranteed-Invalid Value** problem. `@property` solves it.

## @property — Typed Custom Properties

```css
@property --hue {
  syntax: '<angle>';           /* Type constraint */
  inherits: false;             /* Don't inherit */
  initial-value: 0deg;         /* Required for non-inherited */
}

@property --progress {
  syntax: '<number>';
  inherits: false;
  initial-value: 0;
}

@property --brand-color {
  syntax: '<color>';
  inherits: true;
  initial-value: #2563eb;
}
```

### Supported Syntax Types

| Syntax | Example Values |
|--------|---------------|
| `<number>` | `0`, `3.14`, `-1` |
| `<integer>` | `0`, `42`, `-3` |
| `<length>` | `10px`, `2rem`, `50vh` |
| `<percentage>` | `50%`, `100%` |
| `<length-percentage>` | `10px`, `50%` |
| `<color>` | `red`, `#ff0`, `rgb()` |
| `<angle>` | `45deg`, `0.5turn` |
| `<time>` | `300ms`, `1s` |
| `<custom-ident>` | `my-value` |
| `*` | Any value (default) |

### Why @property Matters: Animating Custom Properties

Without `@property`, custom properties can't be animated — the browser doesn't know the type:

```css
/* ❌ Without @property — jumps, doesn't animate */
.gradient-box {
  --angle: 0deg;
  background: linear-gradient(var(--angle), red, blue);
  transition: --angle 1s;     /* Browser doesn't know it's an angle */
}

/* ✅ With @property — smooth animation */
@property --angle {
  syntax: '<angle>';
  inherits: false;
  initial-value: 0deg;
}

.gradient-box {
  --angle: 0deg;
  background: linear-gradient(var(--angle), red, blue);
  transition: --angle 1s;     /* Browser knows it's an angle — can interpolate */
}
.gradient-box:hover {
  --angle: 180deg;
}
```

## Experiment — Animated Gradient Border

```html
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>@property Animation</title>
<style>
@property --gradient-angle {
  syntax: '<angle>';
  inherits: false;
  initial-value: 0deg;
}

@property --glow-opacity {
  syntax: '<number>';
  inherits: false;
  initial-value: 0.5;
}

* { margin: 0; box-sizing: border-box; }

body {
  min-height: 100vh;
  display: grid;
  place-items: center;
  background: #0f172a;
  font-family: system-ui;
}

.card {
  --gradient-angle: 0deg;
  --glow-opacity: 0.5;

  position: relative;
  width: 300px;
  padding: 2rem;
  background: #1e293b;
  border-radius: 12px;
  color: white;

  /* Animated gradient border using pseudo-element */
  border: 2px solid transparent;
  background-clip: padding-box;

  transition: --glow-opacity 0.3s;
  animation: rotate-gradient 3s linear infinite;
}

.card::before {
  content: '';
  position: absolute;
  inset: -2px;
  border-radius: inherit;
  background: conic-gradient(
    from var(--gradient-angle),
    #06b6d4,
    #8b5cf6,
    #ec4899,
    #06b6d4
  );
  z-index: -1;
  opacity: var(--glow-opacity);
  transition: opacity 0.3s;
}

.card:hover {
  --glow-opacity: 1;
}

@keyframes rotate-gradient {
  to { --gradient-angle: 360deg; }
}

.card h2 { margin-bottom: 0.5rem; }
.card p { color: #94a3b8; line-height: 1.6; }

.info {
  margin-top: 2rem;
  color: #64748b;
  font-size: 14px;
  text-align: center;
}
</style>
</head>
<body>
  <div>
    <div class="card">
      <h2>@property Magic</h2>
      <p>This gradient border is animated using a typed custom property.
         Without @property, the angle would jump instead of rotating smoothly.
         Hover to increase glow.</p>
    </div>
    <p class="info">
      @property declares --gradient-angle as &lt;angle&gt;,<br>
      enabling smooth keyframe interpolation.
    </p>
  </div>
</body>
</html>
```

### What to Observe

1. The gradient border **rotates smoothly** — only possible because `@property` tells the browser `--gradient-angle` is an `<angle>`
2. Hover increases opacity smoothly — `--glow-opacity` is typed as `<number>`
3. Remove the `@property` declarations and watch the animation break — it'll jump between start and end

## Dark Mode Pattern with Custom Properties

```css
:root {
  /* Light mode (default) */
  --bg: #ffffff;
  --text: #1a1a1a;
  --text-muted: #666666;
  --surface: #f5f5f5;
  --border: #e0e0e0;
  --primary: #2563eb;
  --primary-hover: #1d4ed8;
}

@media (prefers-color-scheme: dark) {
  :root {
    --bg: #0f172a;
    --text: #e2e8f0;
    --text-muted: #94a3b8;
    --surface: #1e293b;
    --border: #334155;
    --primary: #3b82f6;
    --primary-hover: #60a5fa;
  }
}

/* Manual toggle override */
[data-theme="dark"] {
  --bg: #0f172a;
  --text: #e2e8f0;
  /* ... same as above */
}

/* Components use variables — no theme-specific selectors needed */
body { background: var(--bg); color: var(--text); }
.card { background: var(--surface); border: 1px solid var(--border); }
.button { background: var(--primary); }
.button:hover { background: var(--primary-hover); }
```

## Performance Considerations

1. **Custom properties inherit** — setting on `:root` affects every element (recalculation on change)
2. **Minimize `:root` changes** — changing one `:root` property triggers style recalc for the entire page
3. **Scope narrowly** — set properties on the component that needs them, not globally
4. **`@property` with `inherits: false`** — limits recalculation scope

```css
/* ❌ Every element recalculates when this changes */
:root { --accent: blue; }

/* ✅ Only .toolbar descendants recalculate */
.toolbar { --accent: blue; }
```

## Next

→ [Lesson 02: Cascade Layers](02-cascade-layers.md)
