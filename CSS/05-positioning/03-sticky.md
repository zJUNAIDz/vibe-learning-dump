# Lesson 03 — Sticky Positioning

## Concept

`position: sticky` is a hybrid — it acts like `relative` until the element reaches a scroll threshold, then it acts like `fixed` (within its containing block).

```
┌─────────────────────────────┐
│ Viewport                     │
│                              │
│  ┌────────────────────────┐  │
│  │ Scroll container        │  │
│  │                         │  │
│  │  ┌──────────────────┐   │  │
│  │  │ Sticky element    │   │  │ ← Scrolls normally
│  │  │ (top: 0)          │   │  │
│  │  └──────────────────┘   │  │
│  │                         │  │
│  │  More content...        │  │
│  └────────────────────────┘  │
│                              │
└─────────────────────────────┘

After scrolling past threshold:

┌─────────────────────────────┐
│ Viewport                     │
│  ┌────────────────────────┐  │
│  │ ┌──────────────────┐   │  │
│  │ │ Sticky element    │   │  │ ← STUCK at top: 0
│  │ │ (stuck!)          │   │  │
│  │ └──────────────────┘   │  │
│  │                         │  │
│  │  Content scrolling...   │  │
│  │                         │  │
│  └────────────────────────┘  │
└─────────────────────────────┘
```

## Requirements for Sticky to Work

All of these must be true or sticky will silently fail:

| Requirement | Why |
|-------------|-----|
| At least one offset (`top`, `bottom`, `left`, `right`) | Browser needs to know WHERE to stick |
| Scroll container ancestor must exist | Sticky attaches to the nearest scrolling ancestor |
| Parent must be taller than the sticky element | Element can only be sticky within its parent's boundary |
| No `overflow: hidden` on ancestors (between sticky and scroll container) | This clips the sticky behaviour |

## Experiment 01: Basic Sticky

```html
<!-- 01-basic-sticky.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Sticky Positioning</title>
  <style>
    body { font-family: system-ui; padding: 0; margin: 0; }
    
    .header {
      background: navy;
      color: white;
      padding: 15px 30px;
    }
    
    .sticky-nav {
      position: sticky;
      top: 0;
      background: white;
      border-bottom: 2px solid navy;
      padding: 10px 30px;
      font-weight: bold;
      z-index: 10;
    }
    
    .content {
      padding: 30px;
      max-width: 600px;
    }
    
    .content p {
      margin-bottom: 15px;
      line-height: 1.6;
    }
    
    .section-header {
      position: sticky;
      top: 44px; /* Below the nav */
      background: #f0f0f0;
      border-bottom: 1px solid #ccc;
      padding: 10px 30px;
      margin: 0 -30px;
      font-weight: bold;
      font-size: 14px;
      z-index: 5;
    }
  </style>
</head>
<body>
  <div class="header">
    <h1 style="margin: 0; font-size: 18px;">Page Title</h1>
  </div>
  
  <div class="sticky-nav">
    Navigation (position: sticky, top: 0)
  </div>
  
  <div class="content">
    <p>Scroll down — the navigation bar sticks to the top of the viewport.</p>
    
    <div class="section-header">Section A (sticky, top: 44px)</div>
    <p>Content for section A. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</p>
    <p>More content here to make the page scrollable. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.</p>
    <p>And even more content. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.</p>
    
    <div class="section-header">Section B (sticky, top: 44px)</div>
    <p>Content for section B. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.</p>
    <p>More content. Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium.</p>
    <p>And more. Totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo.</p>
    
    <div class="section-header">Section C (sticky, top: 44px)</div>
    <p>Content for section C. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit.</p>
    <p>More content to fill the page. Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet.</p>
    <p>Final padding content. Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit laboriosam.</p>
    <p>More filler. Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse quam nihil molestiae consequatur.</p>
    <p>And more. At vero eos et accusamus et iusto odio dignissimos ducimus qui blanditiis praesentium voluptatum.</p>
  </div>
</body>
</html>
```

## Experiment 02: Why Sticky Fails

```html
<!-- 02-sticky-failures.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Sticky Failure Cases</title>
  <style>
    body { font-family: system-ui; padding: 30px; margin: 0; }
    
    .example {
      width: 400px;
      height: 200px;
      overflow-y: auto;
      border: 2px solid #999;
      background: #f0f0f0;
      margin-bottom: 30px;
    }
    
    .sticky-item {
      position: sticky;
      top: 0;
      background: lightcoral;
      border: 2px solid darkred;
      padding: 10px;
      font-family: monospace;
      font-size: 12px;
    }
    
    .content-filler {
      padding: 10px;
      line-height: 1.6;
    }
    
    .fail-label {
      font-family: monospace;
      font-size: 13px;
      background: #f8d7da;
      border: 1px solid #dc3545;
      padding: 10px;
      margin-bottom: 5px;
      border-radius: 4px;
    }
    
    .pass-label {
      font-family: monospace;
      font-size: 13px;
      background: #d4edda;
      border: 1px solid #28a745;
      padding: 10px;
      margin-bottom: 5px;
      border-radius: 4px;
    }
  </style>
</head>
<body>
  <h2>Why position: sticky Fails</h2>
  
  <!-- Failure 1: No top/bottom/left/right -->
  <div class="fail-label">❌ FAIL: No offset specified (no top/bottom)</div>
  <div class="example">
    <div style="position: sticky; background: lightcoral; padding: 10px; border: 2px solid darkred;">
      Sticky with NO top value — just scrolls normally
    </div>
    <div class="content-filler">
      Lorem ipsum dolor sit amet, consectetur adipiscing elit.
      Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
      Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.
      Duis aute irure dolor in reprehenderit in voluptate velit esse.
    </div>
  </div>
  
  <!-- Failure 2: overflow: hidden on ancestor -->
  <div class="fail-label">❌ FAIL: overflow: hidden on ancestor</div>
  <div class="example">
    <div style="overflow: hidden;">
      <div class="sticky-item">
        Sticky with top: 0, but parent has overflow: hidden
      </div>
      <div class="content-filler">
        Lorem ipsum dolor sit amet, consectetur adipiscing elit.
        Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
        Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.
        Duis aute irure dolor in reprehenderit in voluptate velit esse.
      </div>
    </div>
  </div>
  
  <!-- Failure 3: Parent not tall enough -->
  <div class="fail-label">❌ FAIL: Parent same height as sticky element</div>
  <div class="example">
    <div style="background: lightyellow; border: 1px solid goldenrod;">
      <div class="sticky-item">
        Parent is only as tall as me — nowhere to stick!
      </div>
    </div>
    <div class="content-filler">
      Other content below the parent.
      Lorem ipsum dolor sit amet, consectetur adipiscing elit.
      Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
    </div>
  </div>
  
  <!-- Success -->
  <div class="pass-label">✅ WORKS: top: 0, parent is tall, no overflow: hidden</div>
  <div class="example">
    <div class="sticky-item" style="background: #d4edda; border-color: green;">
      Sticky with top: 0 ✅
    </div>
    <div class="content-filler">
      Lorem ipsum dolor sit amet, consectetur adipiscing elit.
      Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
      Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.
      Duis aute irure dolor in reprehenderit in voluptate velit esse.
      Excepteur sint occaecat cupidatat non proident, sunt in culpa.
    </div>
  </div>
</body>
</html>
```

## Sticky Boundary: Parent Constraint

A sticky element can only stick **within its parent**. Once the parent scrolls out of view, the sticky element goes with it:

```
Scroll down:

┌─────────────────────┐
│  ┌────────────────┐  │ ← parent boundary
│  │ Sticky element  │  │ ← STUCK at top
│  │                 │  │
│  │ parent content  │  │
│  │                 │  │
│  └────────────────┘  │
│                      │
└─────────────────────┘

Scroll more (parent leaving):

│  │ parent content  │  │ ← parent almost gone
│  │ Sticky element  │  │ ← pushed up WITH parent
│  └────────────────┘  │
│                      │
│  Next content...     │
```

## DevTools Exercise

1. Open Experiment 02
2. For each failing sticky element:
   - Select it in DevTools → Computed tab
   - Check `position: sticky` is applied
   - Check if `top` has a value
   - Check ancestors for `overflow` values
   - Use `getComputedStyle(el).position` in console to verify

## Next

→ [Lesson 04: Offset Properties & Sizing](04-offsets.md)
