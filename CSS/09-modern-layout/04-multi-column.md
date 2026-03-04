# Lesson 04 — Multi-Column & Other Layout

## Multi-Column Layout

CSS multi-column creates **newspaper-style** columns where content flows from one column to the next:

```css
.article {
  columns: 3;             /* 3 columns, auto width */
  column-gap: 2em;        /* gap between columns */
  column-rule: 1px solid #ccc;  /* vertical line between columns */
}

/* Or specify column width (browser decides how many fit): */
.article {
  columns: 20em;          /* each column ~20em, number is auto */
}

/* Explicit control: */
.article {
  column-count: 3;
  column-width: 15em;     /* minimum width — fewer columns on narrow screens */
}
```

### Spanning Columns

```css
h2 {
  column-span: all;  /* break out and span all columns */
}
```

### Preventing Column Breaks

```css
.card {
  break-inside: avoid;    /* don't split this element across columns */
}

h3 {
  break-after: avoid;     /* don't put a column break right after this heading */
}
```

## Float Shapes

`shape-outside` lets text flow **around non-rectangular shapes**:

```css
.circle-float {
  float: left;
  width: 200px;
  height: 200px;
  border-radius: 50%;
  shape-outside: circle(50%);   /* text wraps around the circle */
  shape-margin: 15px;           /* padding around the shape */
}

/* Polygon: */
.triangle-float {
  float: left;
  width: 200px;
  height: 200px;
  shape-outside: polygon(0 0, 100% 0, 50% 100%);
  clip-path: polygon(0 0, 100% 0, 50% 100%);  /* clip element to match */
}

/* From image alpha channel: */
.shaped {
  float: left;
  shape-outside: url('shape.png');
  shape-image-threshold: 0.5;  /* alpha threshold */
}
```

**Important**: `shape-outside` only works on **floated elements**. The element itself isn't clipped — use `clip-path` to visually match the shape.

## `clip-path`

Clips an element to a geometric shape:

```css
.element {
  clip-path: circle(50%);
  clip-path: ellipse(50% 30% at center);
  clip-path: polygon(50% 0%, 100% 100%, 0% 100%);  /* triangle */
  clip-path: inset(10px 20px 30px 40px round 8px);
  clip-path: path('M 0 0 L 100 0 L 50 100 Z');      /* SVG path */
}
```

`clip-path` creates a **stacking context** (like `transform`).

## Experiment: Multi-Column

```html
<!-- 04-multi-column.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Multi-Column Layout</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; max-width: 900px; }
    
    .multicol {
      columns: 3;
      column-gap: 2em;
      column-rule: 1px solid #ddd;
      margin-bottom: 30px;
    }
    
    .multicol p { margin: 0 0 1em; font-size: 14px; line-height: 1.6; }
    .multicol h3 { column-span: all; text-align: center; border-bottom: 2px solid #333; padding-bottom: 8px; }
    
    .multicol .no-break {
      break-inside: avoid;
      background: #f5f5f5;
      border: 1px solid #ddd;
      padding: 10px;
      border-radius: 6px;
      margin-bottom: 1em;
    }
    
    .shape-demo {
      max-width: 500px;
      margin-bottom: 30px;
    }
    
    .circle-float {
      float: left;
      width: 150px;
      height: 150px;
      border-radius: 50%;
      background: linear-gradient(135deg, steelblue, lightblue);
      shape-outside: circle(50%);
      shape-margin: 15px;
      margin-right: 0;
    }
    
    .label { font-family: monospace; font-size: 13px; margin-bottom: 8px; margin-top: 20px; }
  </style>
</head>
<body>
  <h2>Multi-Column Layout</h2>
  
  <div class="label">columns: 3 with column-span and break-inside: avoid</div>
  <div class="multicol">
    <p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam.</p>
    
    <h3>Section Title (column-span: all)</h3>
    
    <p>Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident.</p>
    
    <div class="no-break">
      <strong>Card (break-inside: avoid)</strong><br>
      This card will not be split across columns.
    </div>
    
    <p>Sunt in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis unde omnis iste natus error sit voluptatem.</p>
    
    <p>Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione.</p>
  </div>
  
  <h2>Float Shape</h2>
  <div class="label">shape-outside: circle(50%) — text wraps around the circle</div>
  <div class="shape-demo">
    <div class="circle-float"></div>
    <p style="font-size: 14px; line-height: 1.6;">
      This text flows around the circular shape. The shape-outside property tells the browser to wrap text along the circle boundary, not the rectangular box. This creates a much more natural and magazine-like text flow. Without shape-outside, text would align to the square bounding box of the floated element. Shape-margin adds extra spacing between the text and the shape boundary.
    </p>
    <div style="clear: both;"></div>
  </div>
  
  <h2>clip-path Shapes</h2>
  <div style="display: flex; gap: 20px; flex-wrap: wrap;">
    <div style="text-align: center;">
      <div style="width: 120px; height: 120px; background: steelblue; clip-path: circle(50%);"></div>
      <code style="font-size: 11px;">circle(50%)</code>
    </div>
    <div style="text-align: center;">
      <div style="width: 120px; height: 120px; background: goldenrod; clip-path: polygon(50% 0%, 100% 100%, 0% 100%);"></div>
      <code style="font-size: 11px;">polygon(triangle)</code>
    </div>
    <div style="text-align: center;">
      <div style="width: 120px; height: 120px; background: tomato; clip-path: polygon(50% 0%, 100% 38%, 82% 100%, 18% 100%, 0% 38%);"></div>
      <code style="font-size: 11px;">polygon(pentagon)</code>
    </div>
    <div style="text-align: center;">
      <div style="width: 120px; height: 120px; background: mediumseagreen; clip-path: inset(10% round 20px);"></div>
      <code style="font-size: 11px;">inset(10% round 20px)</code>
    </div>
  </div>
</body>
</html>
```

## Summary

| Feature | Use Case | Key Gotcha |
|---------|----------|-----------|
| Multi-column | Flowing text content | Use `break-inside: avoid` for cards |
| `shape-outside` | Text wrapping around shapes | Only works on floated elements |
| `clip-path` | Visual clipping to shapes | Creates a stacking context |
| `aspect-ratio` | Maintaining proportions | Overridden if both width AND height are set |
| Logical properties | Internationalization | Always prefer over physical properties |
| Container queries | Component-responsive design | `container-type` creates containment |

## Next Module

→ [Module 10: Painting & Compositing](../10-painting-compositing/README.md)
