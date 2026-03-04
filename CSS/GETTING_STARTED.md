# Getting Started — CSS Mastery Course

## Environment Setup

### Browser
Use **Google Chrome** (latest stable) as your primary development browser.

Chrome DevTools has the best CSS debugging tools:
- Elements panel (Computed tab)
- Rendering tab (paint flashing, layer borders)
- Performance panel (layout/paint profiling)
- Layers panel (compositing visualization)

### Editor
Any text editor works. **VS Code** is recommended with these extensions:
- **Live Server** — instant preview with auto-reload
- **CSS Peek** — jump to CSS definitions
- **HTML CSS Support** — IntelliSense for CSS classes

### Project Setup

Create a working directory for experiments:

```bash
mkdir css-mastery-experiments
cd css-mastery-experiments
```

Create a base template file you'll reuse throughout the course:

```html
<!-- template.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>CSS Experiment</title>
  <style>
    /* Reset */
    *, *::before, *::after {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }

    /* Experiment styles go here */
  </style>
</head>
<body>
  <!-- Experiment markup goes here -->
</body>
</html>
```

### Running Experiments

**Option A — Live Server (recommended):**
1. Open the experiment file in VS Code
2. Right-click → "Open with Live Server"
3. Edit and save — browser refreshes automatically

**Option B — File protocol:**
1. Open the `.html` file directly in Chrome
2. Use `Cmd+Shift+R` / `Ctrl+Shift+R` to hard refresh after changes

**Option C — CodePen/JSFiddle:**
Good for quick experiments, but you lose DevTools depth.

### DevTools Setup

Open Chrome DevTools (`F12` or `Cmd+Option+I`):

1. **Elements panel** → Enable "Show user agent shadow DOM" in Settings
2. **Rendering tab** → Access via `Cmd+Shift+P` → "Show Rendering"
3. **Enable paint flashing** in Rendering tab
4. **Enable layer borders** in Rendering tab
5. **Performance panel** → Enable "Screenshots" checkbox

### Verification

Open `template.html` in Chrome. Open DevTools. You should see:
- The Elements panel showing your DOM
- The Styles panel showing your CSS
- The Computed tab showing resolved values

If all of this works, you're ready.

## Course Navigation

Each module is a folder (`01-browser-rendering/`, `02-cascade/`, etc.).

Each module contains:
- `README.md` — Module overview and lesson index
- Numbered lesson files — Sequential lessons within the module

Work through modules in order. Each builds on concepts from previous modules.

## Next Step

→ [Module 01: Browser Rendering](01-browser-rendering/README.md)
