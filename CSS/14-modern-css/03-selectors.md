# Lesson 03 — Modern Selectors

## :is() — Forgiving Selector Matching

Matches any selector in the list. Takes the **highest specificity** of its arguments.

```css
/* Without :is() */
header a:hover,
nav a:hover,
footer a:hover {
  color: blue;
}

/* With :is() */
:is(header, nav, footer) a:hover {
  color: blue;
}
```

### Specificity of :is()

`:is()` takes the specificity of its **most specific argument**:

```css
:is(.card, #sidebar) a {
  /* Specificity = (1,0,1) — takes #sidebar's (1,0,0) */
  color: blue;
}
```

This can be surprising. Even though `.card a` would normally be (0,1,1), wrapping it with `#sidebar` in the same `:is()` elevates the entire selector.

### Forgiving Parsing

If one selector in the list is invalid, others still work:

```css
/* :is() is forgiving — unknown selectors don't break the rule */
:is(.valid, :unsupported-pseudo) p {
  color: blue;  /* Still applies to .valid p */
}

/* Regular selector lists are NOT forgiving */
.valid p, :unsupported-pseudo p {
  color: blue;  /* Entire rule is dropped */
}
```

## :where() — Zero-Specificity :is()

Same matching behavior as `:is()`, but contributes **zero specificity**:

```css
/* Specificity: (0,0,0) for the :where() part */
:where(header, nav, footer) a {
  color: blue;          /* (0,0,1) — only the `a` counts */
}

/* Easy to override */
nav a {
  color: red;           /* (0,0,2) — wins */
}
```

### Use Case: Overridable Defaults

```css
/* Framework provides defaults with :where() */
:where(.btn) {
  padding: 8px 16px;
  border-radius: 4px;
  border: 1px solid #ddd;
}

/* User overrides with normal specificity — always wins */
.btn {
  border-radius: 0;     /* (0,1,0) beats (0,0,0) */
}
```

This is why modern CSS resets use `:where()`:

```css
/* Modern reset — zero specificity, trivially overridable */
:where(ul, ol) { list-style: none; }
:where(img) { max-width: 100%; display: block; }
:where(a) { text-decoration: none; color: inherit; }
```

## :has() — The "Parent Selector"

Selects an element **based on its descendants, siblings, or state**:

```css
/* Card that contains an image */
.card:has(img) {
  grid-template-rows: auto 1fr;
}

/* Card without an image */
.card:not(:has(img)) {
  padding: 2rem;
}

/* Form group with an invalid input */
.form-group:has(input:invalid) {
  --border-color: red;
  --label-color: red;
}

/* Label when its associated input is focused */
label:has(+ input:focus) {
  color: blue;
  font-weight: bold;
}
```

### :has() as a Sibling Selector

```css
/* Select an element followed by a specific sibling */
h2:has(+ p) {
  margin-bottom: 0.5rem;    /* h2 followed by p gets tighter spacing */
}

h2:has(+ ul) {
  margin-bottom: 1rem;      /* h2 followed by ul gets more space */
}
```

### :has() for Page-Level Conditions

```css
/* Body styling based on page content */
body:has(.modal.open) {
  overflow: hidden;           /* Lock scroll when modal is open */
}

body:has(#sidebar.collapsed) {
  --sidebar-width: 60px;
}

/* Apply dark mode based on a toggle checkbox */
html:has(#dark-mode:checked) {
  --bg: #0f172a;
  --text: #e2e8f0;
}
```

### Performance Consideration

`:has()` forces the browser to evaluate **upward** in the DOM tree. Complex `:has()` selectors with deep descendants can be expensive:

```css
/* ❌ Potentially expensive — deep descendant check */
.page:has(.deeply .nested .element) { }

/* ✅ Better — direct child or adjacent sibling */
.card:has(> img) { }
label:has(+ input:focus) { }
```

## :not() — Enhanced Negation

Modern `:not()` accepts **selector lists**:

```css
/* Old :not() — single simple selector */
.item:not(.active) { opacity: 0.5; }

/* Modern :not() — selector list */
.item:not(.active, .pending, .loading) {
  opacity: 0.5;
}

/* Combine with :has() */
.card:not(:has(img)):not(:has(video)) {
  /* Text-only cards */
  padding: 2rem;
}
```

### Specificity of :not()

Like `:is()`, `:not()` takes the specificity of its most specific argument:

```css
:not(#sidebar) p {
  /* (1,0,1) — #sidebar's specificity counts */
}
```

## Combining Modern Selectors

```css
/* Select all interactive elements (forms) in a specific context */
:is(form, fieldset):has(:is(input, select, textarea):invalid) {
  border-color: red;
}

/* Apply styles to siblings of the hovered item */
.nav-item:has(~ .nav-item:hover) {
  /* Items BEFORE the hovered item */
  opacity: 0.7;
  transform: translateX(-4px);
}

.nav-item:hover ~ .nav-item {
  /* Items AFTER the hovered item */
  opacity: 0.7;
  transform: translateX(4px);
}
```

## Experiment — :has() Form Validation

```html
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>:has() Form Validation</title>
<style>
* { margin: 0; box-sizing: border-box; }
body {
  font-family: system-ui;
  padding: 2rem;
  background: #f8fafc;
  min-height: 100vh;
}

.form {
  max-width: 400px;
  margin: 0 auto;
  background: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

h2 { margin-bottom: 1.5rem; }

.field {
  margin-bottom: 1.5rem;
  --label-color: #374151;
  --border-color: #d1d5db;
  --message-display: none;
  --message-color: #666;
}

/* Valid state */
.field:has(input:valid:not(:placeholder-shown)) {
  --label-color: #059669;
  --border-color: #059669;
}

/* Invalid state (only when user has typed something) */
.field:has(input:invalid:not(:placeholder-shown)) {
  --label-color: #dc2626;
  --border-color: #dc2626;
  --message-display: block;
  --message-color: #dc2626;
}

/* Focused state */
.field:has(input:focus) {
  --border-color: #2563eb;
  --label-color: #2563eb;
}

.field label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 600;
  font-size: 14px;
  color: var(--label-color);
  transition: color 0.2s;
}

.field input {
  width: 100%;
  padding: 10px 12px;
  border: 2px solid var(--border-color);
  border-radius: 6px;
  font-size: 16px;
  outline: none;
  transition: border-color 0.2s;
}

.field .message {
  display: var(--message-display);
  font-size: 12px;
  margin-top: 4px;
  color: var(--message-color);
}

/* Submit button state based on form validity */
.form:has(input:invalid) .submit-btn {
  opacity: 0.5;
  cursor: not-allowed;
}

.form:not(:has(input:invalid)) .submit-btn {
  opacity: 1;
  cursor: pointer;
}

.submit-btn {
  width: 100%;
  padding: 12px;
  background: #2563eb;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 16px;
  font-weight: 600;
  transition: opacity 0.2s;
}

.info {
  text-align: center;
  margin-top: 1.5rem;
  font-size: 13px;
  color: #94a3b8;
  max-width: 400px;
  margin-inline: auto;
}
</style>
</head>
<body>
  <form class="form" onsubmit="event.preventDefault()">
    <h2>:has() Validation</h2>

    <div class="field">
      <label>Email</label>
      <input type="email" placeholder="you@example.com" required>
      <span class="message">Please enter a valid email</span>
    </div>

    <div class="field">
      <label>Password</label>
      <input type="password" placeholder="Min 8 characters" 
             required minlength="8" pattern=".{8,}">
      <span class="message">Must be at least 8 characters</span>
    </div>

    <div class="field">
      <label>Username</label>
      <input type="text" placeholder="letters and numbers only" 
             required pattern="[a-zA-Z0-9]{3,20}">
      <span class="message">3-20 alphanumeric characters</span>
    </div>

    <button type="submit" class="submit-btn">Create Account</button>
  </form>
  <p class="info">
    No JavaScript used for validation styling.<br>
    All states driven by <code>:has()</code> + <code>:valid</code> / <code>:invalid</code>.
  </p>
</body>
</html>
```

### What to Observe

1. Labels change color based on input validity — driven entirely by `:has()`
2. Error messages appear only for invalid, non-empty fields
3. Submit button dims when any input is invalid (`.form:has(input:invalid)`)
4. **Zero JavaScript** for all visual validation state

## Selector Reference Table

| Selector | Specificity | Purpose |
|----------|-------------|---------|
| `:is(A, B)` | Highest of A, B | Compact matching, forgiving |
| `:where(A, B)` | (0,0,0) | Overridable defaults |
| `:has(A)` | Normal (of A) | Parent/context selection |
| `:not(A, B)` | Highest of A, B | Exclusion |

## Browser Support

- `:is()` / `:where()` — Baseline 2021
- `:has()` — Baseline 2023 (Chrome 105+, Firefox 121+, Safari 15.4+)
- `:not()` selector list — Baseline 2021

## Next

→ [Lesson 04: Color & Functions](04-color-and-functions.md)
