# Lesson 02 — Transforms & Animations

## CSS Transforms

`transform` applies geometric transformations **without affecting layout** — the element's original space is preserved.

### 2D Transforms

```css
.element {
  transform: translate(50px, 20px);   /* move */
  transform: scale(1.5);              /* resize (1 = normal) */
  transform: rotate(45deg);           /* rotate */
  transform: skew(10deg, 5deg);       /* skew */
  
  /* Chained (applied right to left): */
  transform: translate(100px, 0) rotate(45deg) scale(0.8);
}
```

### 3D Transforms

```css
.parent {
  perspective: 800px;    /* 3D depth — lower = more dramatic */
}

.child {
  transform: rotateY(30deg);            /* rotate around Y axis */
  transform: translateZ(-100px);        /* move toward/away from viewer */
  transform: rotate3d(1, 1, 0, 45deg); /* arbitrary axis */
  transform-style: preserve-3d;         /* children participate in 3D */
  backface-visibility: hidden;          /* hide when facing away */
}
```

### `transform-origin`

The point around which transforms apply (default: `center center`):

```css
.element {
  transform-origin: top left;     /* rotate/scale from top-left */
  transform-origin: 0% 100%;     /* bottom-left */
  transform-origin: 50% 50% -50px; /* 3D origin */
}
```

### What Transform Does to the Element

| Effect | Detail |
|--------|--------|
| Creates a **stacking context** | Always, even `transform: translateX(0)` |
| Creates a **containing block** | For absolute/fixed descendants |
| **Breaks** `position: fixed` | Fixed children become absolute to the transform ancestor |
| Does NOT cause **layout** | Original space is preserved; only composite phase |

### Individual Transform Properties (Modern)

```css
/* Instead of: */
.old { transform: translate(10px, 20px) rotate(45deg) scale(1.2); }

/* Modern individual properties: */
.new {
  translate: 10px 20px;
  rotate: 45deg;
  scale: 1.2;
}
```

These can be animated independently — no need to repeat the whole `transform` chain.

## CSS Transitions

Transitions animate property changes **between two states**:

```css
.button {
  background: blue;
  transform: scale(1);
  transition: background 0.3s ease, transform 0.2s ease;
}

.button:hover {
  background: darkblue;
  transform: scale(1.05);
}
```

### Transition Properties

```css
transition: <property> <duration> <timing-function> <delay>;

/* Shorthand: */
transition: all 0.3s ease;              /* animate everything */
transition: transform 0.3s ease-out;    /* specific property */
transition: opacity 0.2s linear 0.1s;   /* with delay */

/* Multiple: */
transition: transform 0.3s ease, opacity 0.2s ease;
```

### Timing Functions

| Function | Behavior |
|----------|----------|
| `ease` | Slow start/end, fast middle (default) |
| `linear` | Constant speed |
| `ease-in` | Slow start |
| `ease-out` | Slow end |
| `ease-in-out` | Slow start and end |
| `cubic-bezier(x1,y1,x2,y2)` | Custom curve |
| `steps(n, jump-start)` | Stepped animation |

### What Can Be Transitioned?

Only properties with **interpolable values**:
- ✅ `opacity`, `transform`, `color`, `background-color`, `width`, `height`, `margin`, `padding`, `font-size`, `border-width`, `box-shadow`
- ❌ `display`, `font-family`, `position`, `grid-template-*` (discrete values)

## CSS Animations

Animations provide **multi-step** sequences with `@keyframes`:

```css
@keyframes slide-in {
  from {
    transform: translateX(-100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

/* Or with percentages: */
@keyframes bounce {
  0%   { transform: translateY(0); }
  40%  { transform: translateY(-30px); }
  60%  { transform: translateY(-15px); }
  80%  { transform: translateY(-5px); }
  100% { transform: translateY(0); }
}

.element {
  animation: slide-in 0.5s ease-out forwards;
}
```

### Animation Properties

```css
animation: <name> <duration> <timing> <delay> <iteration> <direction> <fill> <play>;

animation: bounce 1s ease-in-out 0s infinite alternate both running;
```

| Property | Values |
|----------|--------|
| `animation-iteration-count` | `1` (default), `2`, `infinite` |
| `animation-direction` | `normal`, `reverse`, `alternate`, `alternate-reverse` |
| `animation-fill-mode` | `none`, `forwards` (keep end state), `backwards`, `both` |
| `animation-play-state` | `running`, `paused` |

## Performance: Compositor-Only Animations

```css
/* ✅ GOOD — runs on compositor thread (60fps guaranteed): */
.smooth {
  transition: transform 0.3s, opacity 0.3s;
}

/* ❌ BAD — triggers layout on every frame: */
.janky {
  transition: left 0.3s, width 0.3s, margin 0.3s;
}
```

To animate position, always use `transform: translate()` instead of `top`/`left`. To animate size visually, use `transform: scale()` instead of `width`/`height`.

## Experiment: Animation Playground

```html
<!-- 02-animations.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Transforms & Animations</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .demo-row {
      display: flex;
      gap: 30px;
      align-items: center;
      margin-bottom: 30px;
    }
    
    .box {
      width: 100px;
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
    
    /* Transitions */
    .t-translate { background: steelblue; transition: transform 0.3s ease; }
    .t-translate:hover { transform: translate(30px, -20px); }
    
    .t-rotate { background: #e74c3c; transition: transform 0.5s ease; }
    .t-rotate:hover { transform: rotate(180deg); }
    
    .t-scale { background: #27ae60; transition: transform 0.3s ease; }
    .t-scale:hover { transform: scale(1.4); }
    
    .t-combo { background: #8e44ad; transition: transform 0.4s ease, opacity 0.4s ease; }
    .t-combo:hover { transform: translateY(-15px) rotate(5deg) scale(1.1); opacity: 0.8; }
    
    /* Keyframe animations */
    @keyframes pulse {
      0%, 100% { transform: scale(1); }
      50% { transform: scale(1.15); }
    }
    
    @keyframes spin {
      from { transform: rotate(0); }
      to { transform: rotate(360deg); }
    }
    
    @keyframes slide-bounce {
      0% { transform: translateX(0); }
      50% { transform: translateX(150px); }
      70% { transform: translateX(130px); }
      85% { transform: translateX(145px); }
      100% { transform: translateX(140px); }
    }
    
    .a-pulse { background: #e74c3c; animation: pulse 1.5s ease-in-out infinite; }
    .a-spin { background: steelblue; animation: spin 2s linear infinite; }
    .a-bounce { background: #f39c12; animation: slide-bounce 1.5s ease-out infinite alternate; }
    
    /* 3D transform */
    .perspective-container {
      perspective: 600px;
      display: inline-block;
    }
    
    .card-3d {
      width: 150px;
      height: 200px;
      background: linear-gradient(135deg, #667eea, #764ba2);
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      color: white;
      font-family: monospace;
      font-size: 12px;
      transition: transform 0.5s ease;
      transform-style: preserve-3d;
    }
    
    .card-3d:hover {
      transform: rotateY(25deg) rotateX(10deg);
    }
    
    .label { font-family: monospace; font-size: 12px; margin-bottom: 5px; color: #666; }
  </style>
</head>
<body>
  <h2>Transitions (hover each box)</h2>
  <div class="demo-row">
    <div>
      <div class="label">translate</div>
      <div class="box t-translate">translate</div>
    </div>
    <div>
      <div class="label">rotate</div>
      <div class="box t-rotate">rotate</div>
    </div>
    <div>
      <div class="label">scale</div>
      <div class="box t-scale">scale</div>
    </div>
    <div>
      <div class="label">combo</div>
      <div class="box t-combo">translate+<br>rotate+scale<br>+opacity</div>
    </div>
  </div>
  
  <h2>Keyframe Animations (continuous)</h2>
  <div class="demo-row">
    <div>
      <div class="label">pulse (infinite)</div>
      <div class="box a-pulse">pulse</div>
    </div>
    <div>
      <div class="label">spin (linear)</div>
      <div class="box a-spin">spin</div>
    </div>
    <div>
      <div class="label">slide-bounce</div>
      <div class="box a-bounce">bounce</div>
    </div>
  </div>
  
  <h2>3D Transform (hover the card)</h2>
  <div class="perspective-container">
    <div class="card-3d">perspective: 600px<br>rotateY + rotateX</div>
  </div>
</body>
</html>
```

## Next

→ [Lesson 03: Filters & Effects](03-filters.md)
