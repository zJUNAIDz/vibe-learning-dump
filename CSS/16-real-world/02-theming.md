# Lesson 02 — Theming

## Dark Mode Implementation

### Strategy: Override Semantic Tokens

The three-tier system makes dark mode a single-layer swap:

```css
/* Light (default) */
:root {
  --color-bg: #ffffff;
  --color-bg-surface: #ffffff;
  --color-bg-muted: #f1f5f9;
  --color-text: #0f172a;
  --color-text-secondary: #475569;
  --color-text-muted: #94a3b8;
  --color-border: #e2e8f0;
  --color-primary: #2563eb;
  --color-primary-hover: #1d4ed8;
  --shadow-md: 0 4px 6px -1px rgba(0,0,0,0.1);
}

/* Dark — override semantic tokens only */
[data-theme="dark"] {
  --color-bg: #0f172a;
  --color-bg-surface: #1e293b;
  --color-bg-muted: #334155;
  --color-text: #e2e8f0;
  --color-text-secondary: #94a3b8;
  --color-text-muted: #64748b;
  --color-border: #334155;
  --color-primary: #3b82f6;
  --color-primary-hover: #60a5fa;
  --shadow-md: 0 4px 6px -1px rgba(0,0,0,0.4);
}
```

**Zero component changes needed.** Every component that uses semantic tokens automatically adapts.

### System Preference + Manual Override

```css
/* System preference detection */
@media (prefers-color-scheme: dark) {
  :root:not([data-theme="light"]) {
    --color-bg: #0f172a;
    /* ... dark overrides */
  }
}

/* Manual override always wins (higher specificity via attribute) */
[data-theme="dark"] {
  --color-bg: #0f172a;
  /* ... dark overrides */
}

[data-theme="light"] {
  --color-bg: #ffffff;
  /* ... light values */
}
```

### Theme Toggle JavaScript

```javascript
function getThemePreference() {
  // Check saved preference
  const saved = localStorage.getItem('theme');
  if (saved) return saved;
  
  // Fall back to system preference
  return window.matchMedia('(prefers-color-scheme: dark)').matches 
    ? 'dark' 
    : 'light';
}

function applyTheme(theme) {
  document.documentElement.setAttribute('data-theme', theme);
  localStorage.setItem('theme', theme);
}

// Apply on load (run before body renders to prevent flash)
applyTheme(getThemePreference());

// Listen for system changes
window.matchMedia('(prefers-color-scheme: dark)')
  .addEventListener('change', (e) => {
    if (!localStorage.getItem('theme')) {
      applyTheme(e.matches ? 'dark' : 'light');
    }
  });
```

### Preventing Flash of Wrong Theme (FOWT)

Place the theme detection script in `<head>` **before** any CSS loads:

```html
<head>
  <script>
    // Runs synchronously before paint
    (function() {
      const theme = localStorage.getItem('theme') || 
        (matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
      document.documentElement.setAttribute('data-theme', theme);
    })();
  </script>
  <link rel="stylesheet" href="styles.css">
</head>
```

## Multi-Brand Theming

When one codebase serves multiple brands (white-label, SaaS with custom branding):

```css
/* Brand tokens override semantic tokens */
[data-brand="acme"] {
  --color-primary: #e63946;
  --color-primary-hover: #c1121f;
  --font-family-heading: 'Playfair Display', serif;
  --radius-md: 0;           /* Square corners for Acme brand */
}

[data-brand="nova"] {
  --color-primary: #06d6a0;
  --color-primary-hover: #05b389;
  --font-family-heading: 'Inter', sans-serif;
  --radius-md: 12px;        /* Rounded for Nova brand */
}

/* Components use tokens — brand-agnostic */
.button {
  background: var(--color-primary);
  border-radius: var(--radius-md);
  font-family: var(--font-family-heading);
}
```

### Combining Brand + Theme

```html
<html data-brand="acme" data-theme="dark">
```

```css
/* Base tokens */
:root { /* primitives */ }

/* Brand tokens (override primitives) */
[data-brand="acme"] { --color-primary: #e63946; }

/* Theme tokens (override semantic) */
[data-theme="dark"] { --color-bg: #0f172a; }

/* Brand + theme specific (rare, for edge cases) */
[data-brand="acme"][data-theme="dark"] {
  --color-primary: #ff6b6b;  /* Lighter red for dark backgrounds */
}
```

## User Preferences Beyond Color Scheme

### Reduced Motion

```css
/* Default: animations enabled */
.card {
  transition: transform 0.3s ease;
}
.card:hover {
  transform: translateY(-4px);
}

/* Respect user preference */
@media (prefers-reduced-motion: reduce) {
  .card {
    transition: none;
  }
  .card:hover {
    transform: none;
  }

  /* Or reduce rather than remove */
  * {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

### High Contrast / Forced Colors

```css
/* Windows High Contrast Mode */
@media (forced-colors: active) {
  .button {
    /* System colors replace custom colors */
    border: 2px solid ButtonText;
    /* Custom colors are ignored — use system colors */
  }

  .icon {
    /* Force SVG icons to use system color */
    forced-color-adjust: auto;
  }

  .decorative-border {
    /* Preserve custom styling */
    forced-color-adjust: none;
  }
}
```

### Contrast Preference

```css
@media (prefers-contrast: more) {
  :root {
    --color-text: #000000;
    --color-bg: #ffffff;
    --color-border: #000000;
    --color-text-secondary: #333333;
  }
}

@media (prefers-contrast: less) {
  :root {
    --color-text: #444444;
    --color-border: #e0e0e0;
  }
}
```

## Experiment — Theme Switcher

```html
<!DOCTYPE html>
<html lang="en" data-theme="light">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Theme System</title>
<script>
  // Prevent FOWT
  (function() {
    const theme = localStorage.getItem('theme') ||
      (matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
    document.documentElement.setAttribute('data-theme', theme);
  })();
</script>
<style>
/* === PRIMITIVE TOKENS === */
:root {
  --white: #ffffff;
  --gray-50: #f8fafc; --gray-100: #f1f5f9;
  --gray-200: #e2e8f0; --gray-400: #94a3b8;
  --gray-600: #475569; --gray-800: #1e293b;
  --gray-900: #0f172a; --gray-950: #020617;
  --blue-500: #3b82f6; --blue-600: #2563eb;
  --blue-700: #1d4ed8;
}

/* === SEMANTIC TOKENS (light) === */
:root, [data-theme="light"] {
  --bg: var(--gray-50);
  --surface: var(--white);
  --text: var(--gray-900);
  --text-secondary: var(--gray-600);
  --text-muted: var(--gray-400);
  --border: var(--gray-200);
  --primary: var(--blue-600);
  --primary-hover: var(--blue-700);
  --shadow: 0 1px 3px rgba(0,0,0,0.08);
  color-scheme: light;
}

/* === SEMANTIC TOKENS (dark) === */
[data-theme="dark"] {
  --bg: var(--gray-950);
  --surface: var(--gray-900);
  --text: var(--gray-100);
  --text-secondary: var(--gray-400);
  --text-muted: var(--gray-600);
  --border: var(--gray-800);
  --primary: var(--blue-500);
  --primary-hover: var(--blue-600);
  --shadow: 0 1px 3px rgba(0,0,0,0.3);
  color-scheme: dark;
}

/* === COMPONENTS (use semantic tokens only) === */
* { margin: 0; box-sizing: border-box; }

body {
  font-family: system-ui;
  background: var(--bg);
  color: var(--text);
  min-height: 100vh;
  padding: 2rem;
  transition: background 0.3s, color 0.3s;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

h1 { font-size: 1.5rem; }

.theme-toggle {
  padding: 8px 16px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--text);
  cursor: pointer;
  font-size: 14px;
  transition: background 0.2s;
}
.theme-toggle:hover { background: var(--bg); }

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}

.card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: var(--shadow);
  transition: background 0.3s, border-color 0.3s, box-shadow 0.3s;
}

.card h3 {
  margin-bottom: 0.5rem;
  color: var(--text);
}

.card p {
  color: var(--text-secondary);
  line-height: 1.6;
  font-size: 14px;
}

.badge {
  display: inline-block;
  margin-top: 1rem;
  padding: 4px 10px;
  background: var(--primary);
  color: white;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
}

.token-display {
  margin-top: 2rem;
  padding: 1.5rem;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  font-family: monospace;
  font-size: 13px;
}

.token-row {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
  border-bottom: 1px solid var(--border);
}

.token-name { color: var(--text-muted); }
.token-value { color: var(--text); }
</style>
</head>
<body>
  <div class="header">
    <h1>Theme System Demo</h1>
    <button class="theme-toggle" id="toggle">Switch Theme</button>
  </div>

  <div class="grid">
    <div class="card">
      <h3>Design Tokens</h3>
      <p>Semantic tokens abstract raw values. The same component works
         in any theme without modification.</p>
      <span class="badge">Tokens</span>
    </div>
    <div class="card">
      <h3>Zero Component Changes</h3>
      <p>Toggle the theme — every component adapts because they 
         reference semantic variables, not hardcoded colors.</p>
      <span class="badge">Theming</span>
    </div>
    <div class="card">
      <h3>No Flash</h3>
      <p>The theme script runs in &lt;head&gt; before CSS loads,
         preventing flash of wrong theme on page load.</p>
      <span class="badge">UX</span>
    </div>
  </div>

  <div class="token-display" id="tokens"></div>

  <script>
    const toggle = document.getElementById('toggle');
    const tokensEl = document.getElementById('tokens');
    const root = document.documentElement;

    function updateTokenDisplay() {
      const style = getComputedStyle(root);
      const tokens = ['--bg', '--surface', '--text', '--text-secondary', 
                       '--border', '--primary'];
      tokensEl.innerHTML = '<strong>Active Token Values:</strong><br><br>' +
        tokens.map(t => {
          const val = style.getPropertyValue(t).trim();
          return `<div class="token-row">
            <span class="token-name">${t}</span>
            <span class="token-value">${val}</span>
          </div>`;
        }).join('');
    }

    toggle.addEventListener('click', () => {
      const current = root.getAttribute('data-theme');
      const next = current === 'dark' ? 'light' : 'dark';
      root.setAttribute('data-theme', next);
      localStorage.setItem('theme', next);
      toggle.textContent = next === 'dark' ? '☀️ Light' : '🌙 Dark';
      updateTokenDisplay();
    });

    // Initial state
    const theme = root.getAttribute('data-theme');
    toggle.textContent = theme === 'dark' ? '☀️ Light' : '🌙 Dark';
    updateTokenDisplay();
  </script>
</body>
</html>
```

## Next

→ [Lesson 03: Responsive Systems](03-responsive-systems.md)
