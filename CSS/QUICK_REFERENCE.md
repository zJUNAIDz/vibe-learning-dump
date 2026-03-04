# CSS Quick Reference

> Fast lookup for specification-grounded CSS concepts. Use this when you need to quickly recall a rule or behavior.

---

## The Rendering Pipeline

```
Bytes → Characters → Tokens → Nodes → DOM/CSSOM → Render Tree → Layout → Paint → Composite → Pixels
```

---

## Cascade Order (Highest to Lowest Priority)

1. User-agent `!important`
2. User `!important`  
3. Author `!important` (unlayered)
4. Author `!important` (layered — first declared layer wins)
5. CSS animations `@keyframes`
6. Author normal (unlayered)
7. Author normal (layered — last declared layer wins)
8. User normal
9. User-agent normal

---

## Specificity (A, B, C)

| Component | Counts |
|---|---|
| **A** — ID selectors | `#id` |
| **B** — Class, attribute, pseudo-class | `.class`, `[attr]`, `:hover` |
| **C** — Type, pseudo-element | `div`, `::before` |

- Inline styles: specificity = (1, 0, 0, 0) — beats all selectors
- `!important`: overrides specificity entirely (uses cascade origin rules)
- `:where()`: zero specificity
- `:is()`, `:not()`, `:has()`: specificity of most specific argument

---

## Box Model

```
┌─────────────── margin ───────────────┐
│ ┌─────────── border ───────────────┐ │
│ │ ┌───────── padding ───────────┐  │ │
│ │ │ ┌─────── content ────────┐  │  │ │
│ │ │ │                        │  │  │ │
│ │ │ └────────────────────────┘  │  │ │
│ │ └─────────────────────────────┘  │ │
│ └──────────────────────────────────┘ │
└──────────────────────────────────────┘
```

- `box-sizing: content-box` (default): width/height = content only
- `box-sizing: border-box`: width/height = content + padding + border

---

## Margin Collapsing Rules

Margins collapse when:
1. Adjacent siblings (vertical margins)
2. Parent and first/last child (no border/padding/content between)
3. Empty blocks

Margins do **NOT** collapse when:
- Floats
- Absolutely positioned elements
- Inline-block elements
- Elements with `overflow` other than `visible`
- Flex/grid items
- Horizontal margins (never collapse)

---

## Formatting Contexts

| Context | Triggered By |
|---|---|
| **Block FC (BFC)** | Root, float, `position: absolute/fixed`, `display: flow-root`, `overflow != visible`, flex/grid items, table cells |
| **Inline FC** | Default for inline content |
| **Flex FC** | `display: flex/inline-flex` |
| **Grid FC** | `display: grid/inline-grid` |

---

## Containing Block Rules

| Position Value | Containing Block |
|---|---|
| `static`, `relative`, `sticky` | Nearest **block-level** ancestor's content box |
| `absolute` | Nearest **positioned** ancestor's padding box |
| `fixed` | The viewport (ICB) |
| `fixed` (with transform ancestor) | The ancestor with transform |

---

## Stacking Context Triggers

A new stacking context is created by:
- Root element (`<html>`)
- `position: absolute/relative` + `z-index` other than `auto`
- `position: fixed/sticky`
- `opacity` < 1
- `transform` other than `none`
- `filter` other than `none`
- `backdrop-filter` other than `none`
- `perspective` other than `none`
- `clip-path` other than `none`
- `mask` / `mask-image` / `mask-border`
- `mix-blend-mode` other than `normal`
- `isolation: isolate`
- `will-change` (specifying a stacking-context-creating property)
- `contain: layout/paint/strict/content`
- Flex/grid child with `z-index` other than `auto`

---

## Stacking Order (Back to Front)

1. Stacking context background/border
2. Negative z-index children
3. Block-level non-positioned children (in DOM order)
4. Float children
5. Inline-level non-positioned children
6. `z-index: 0` and positioned children (DOM order)
7. Positive z-index children (ascending)

---

## Flexbox Algorithm Summary

1. Determine available main size
2. Collect flex items, compute hypothetical main size (`flex-basis`)
3. Determine if items overflow or underflow
4. Distribute free space using `flex-grow` or shrink using `flex-shrink`
5. Resolve cross sizes
6. Align items on cross axis (`align-items`, `align-self`)
7. Align content on main axis (`justify-content`)
8. Handle wrapping if `flex-wrap` is set

---

## Grid Track Sizing Algorithm Summary

1. Initialize track sizes (from explicit definitions)
2. Resolve intrinsic sizes of content-sized tracks
3. Maximize tracks within growth limits
4. Expand flexible tracks (`fr` units)
5. Expand auto tracks to fill remaining space

---

## Common `display` Values

| Value | Outer | Inner |
|---|---|---|
| `block` | block | flow |
| `inline` | inline | flow |
| `inline-block` | inline | flow-root |
| `flex` | block | flex |
| `inline-flex` | inline | flex |
| `grid` | block | grid |
| `inline-grid` | inline | grid |
| `flow-root` | block | flow-root |
| `none` | removed from rendering | — |
| `contents` | removed box, children promoted | — |

---

## Value Resolution Order

```
Declared → Cascaded → Specified → Computed → Used → Actual
```

1. **Declared**: All declarations that apply to the element
2. **Cascaded**: Winner after cascade resolution
3. **Specified**: Cascaded value, or inherited/initial value
4. **Computed**: Absolute values resolved (e.g., `em` → `px`, relative URLs resolved)
5. **Used**: Final values after layout (e.g., percentages resolved)
6. **Actual**: Value after rounding/clamping for the device

---

## Performance: What Triggers What

| Change | Triggers |
|---|---|
| Layout property (`width`, `height`, `margin`, `padding`, `top`, `left`, `display`, `position`, `float`) | Layout → Paint → Composite |
| Paint-only property (`color`, `background`, `box-shadow`, `border-radius`, `visibility`) | Paint → Composite |
| Composite-only property (`transform`, `opacity`) | Composite only |

---

## Logical Properties Mapping (LTR horizontal-tb)

| Physical | Logical |
|---|---|
| `width` | `inline-size` |
| `height` | `block-size` |
| `margin-top` | `margin-block-start` |
| `margin-bottom` | `margin-block-end` |
| `margin-left` | `margin-inline-start` |
| `margin-right` | `margin-inline-end` |
| `padding-top` | `padding-block-start` |
| `top` | `inset-block-start` |
| `left` | `inset-inline-start` |
