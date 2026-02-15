# Limits

## Why This Matters

**Limits are the gateway to calculus.** They answer the question:

> "What value does a function approach as the input gets closer and closer to some number?"

Limits let us:
- **Handle infinity**: Understand what happens "at the edge"
- **Deal with discontinuities**: Analyze behavior at problem points
- **Define derivatives**: Instantaneous rate of change
- **Define integrals**: Total accumulation

Without limits, there is no calculus. With limits, you can understand change itself.

---

## The Big Picture: Approaching vs Reaching

### The Core Idea

**A limit is about getting arbitrarily close, not necessarily arriving.**

```
"What does f(x) approach as x approaches 2?"

x values:  1.9   1.99   1.999   1.9999  ...
f(x):      3.8   3.98   3.998   3.9998  ...

The function approaches 4 (even if f(2) ≠ 4 or doesn't exist)
```

**Notation**:
```
lim f(x) = L
x→a

Read as: "The limit of f(x) as x approaches a equals L"
```

### Why Not Just Substitute?

**Sometimes substitution works**:
```
f(x) = 2x + 1

lim f(x) = f(3) = 7
x→3
```

**But sometimes it doesn't**:
```
       x² - 4
f(x) = ──────
        x - 2

At x = 2: 0/0 (undefined!)

But the limit exists: lim f(x) = 4
                      x→2
```

---

## 1. Intuitive Understanding

### Visual: Zooming In

```
Graph of f(x) = x²:

    y
    │
  4 ├───────●  (2, 4)
    │      ╱
  3 ├─────╱
    │    ╱
  2 ├───╱
    │  ╱
  1 ├─●
────┼──────────► x
    0  1  2  3

As x → 2, f(x) → 4
```

Even without computing f(2), we can see where it's heading.

### Table of Values

```
f(x) = (x² - 4) / (x - 2)

Approaching from the left:
x:    1.9      1.99     1.999    1.9999
f(x): 3.9      3.99     3.999    3.9999

Approaching from the right:
x:    2.1      2.01     2.001    2.0001
f(x): 4.1      4.01     4.001    4.0001

Both approach 4
```

### One-Sided Limits

**Left-hand limit**: Approach from values less than a
```
lim f(x)  (x → a from the left)
x→a⁻
```

**Right-hand limit**: Approach from values greater than a
```
lim f(x)  (x → a from the right)
x→a⁺
```

**Two-sided limit exists if**:
```
lim f(x) = lim f(x) = L
x→a⁻     x→a⁺

Then: lim f(x) = L
      x→a
```

---

## 2. Evaluating Limits

### Direct Substitution

**If the function is continuous at a, just plug in**:

```
lim (3x² + 2x - 1) = 3(2)² + 2(2) - 1
x→2                = 12 + 4 - 1
                   = 15
```

### Indeterminate Form 0/0

**When substitution gives 0/0, factor and simplify**:

```
       x² - 9
lim ──────────
x→3    x - 3

Direct: (9-9)/(3-3) = 0/0  ✗

Factor numerator:
       (x+3)(x-3)
lim ────────────
x→3     x - 3

Cancel (valid since x ≠ 3 in the limit):
lim (x + 3) = 6
x→3
```

### Rationalizing

**Multiply by conjugate to eliminate square roots**:

```
      √(x+1) - 1
lim ────────────
x→0      x

Direct: 0/0  ✗

Multiply by conjugate:
      √(x+1) - 1   √(x+1) + 1
    ──────────── × ────────────
           x       √(x+1) + 1

       (x+1) - 1
  = ─────────────────
     x(√(x+1) + 1)

       x
  = ─────────────────
     x(√(x+1) + 1)

  = ─────────────  (cancel x)
     √(x+1) + 1

  = 1/(1 + 1) = 1/2
```

---

## 3. Limits at Infinity

### What It Means

**What happens as x gets arbitrarily large (or negative)?**

```
lim f(x)  (x → ∞)
x→∞

lim f(x)  (x → -∞)
x→-∞
```

### Polynomial Limits at Infinity

**Dominated by highest-degree term**:

```
       3x² + 5x - 1
lim ──────────────────
x→∞   2x² - x + 7

For large x, only highest powers matter:
     3x²
≈ ──────── = 3/2
     2x²

lim = 3/2
x→∞
```

**General rule**:
```
       aₘxᵐ + ...
lim ───────────────
x→∞    bₙxⁿ + ...

If m < n: limit = 0
If m = n: limit = aₘ/bₙ
If m > n: limit = ±∞
```

### Exponential vs Polynomial

**Exponential grows faster**:

```
lim x/eˣ = 0     (denominator grows faster)
x→∞

lim eˣ/x = ∞     (numerator grows faster)
x→∞
```

### Logarithmic Growth

**Logarithms grow slower than polynomials**:

```
lim log(x)/x = 0
x→∞

lim x/log(x) = ∞
x→∞
```

---

## 4. Continuous Functions

### Definition

A function is **continuous at x = a** if:

```
lim f(x) = f(a)
x→a
```

**Three conditions**:
1. f(a) is defined
2. lim f(x) exists
   x→a
3. They're equal

### Visual

```
Continuous:           Discontinuous:
    │                     │
────●────             ────○──●─
  (smooth)              (break)
```

### Types of Discontinuity

**Jump discontinuity**:
```
      │
    ──┤  ●
      │
    ● ├──
      │

Left and right limits exist but differ
```

**Removable discontinuity**:
```
      │
    ──○── ● (hole at one point)
      │

Can be "fixed" by redefining f(a)
```

**Infinite discontinuity**:
```
      │    │
    ──┘    └──
      │ ∞  │

Vertical asymptote (limit is ∞ or -∞)
```

---

## 5. Important Limit Patterns

### Limit of sin(x)/x

```
lim sin(x)/x = 1
x→0
```

**One of the most important limits in calculus.**

**Visual reasoning**:
```
For small angles (in radians):
sin(x) ≈ x

So sin(x)/x ≈ x/x = 1
```

### Limit of (1 + 1/n)ⁿ

```
lim (1 + 1/n)ⁿ = e ≈ 2.71828...
n→∞
```

**Definition of Euler's number e.**

**Generalization**:
```
lim (1 + x/n)ⁿ = eˣ
n→∞
```

### Squeeze Theorem

**If g(x) ≤ f(x) ≤ h(x) and**:
```
lim g(x) = lim h(x) = L
x→a       x→a
```

**Then**:
```
lim f(x) = L
x→a
```

**Example**: lim x² sin(1/x)
            x→0

```
-1 ≤ sin(1/x) ≤ 1  (always true)

Multiply by x²:
-x² ≤ x² sin(1/x) ≤ x²

lim (-x²) = 0,  lim (x²) = 0
x→0           x→0

Therefore: lim x² sin(1/x) = 0
          x→0
```

---

## 6. L'Hôpital's Rule (Preview)

**For indeterminate forms 0/0 or ∞/∞**:

```
       f(x)         f'(x)
lim ──────── = lim ────────
x→a    g(x)   x→a   g'(x)
```

(Take derivatives of numerator and denominator separately)

**Example**:
```
       x² - 1
lim ──────────  = 0/0
x→1    x - 1

Apply L'Hôpital:
       2x
lim ────── = 2/1 = 2
x→1     1
```

**Note**: We haven't learned derivatives yet, but this preview shows where limits lead.

---

## 7. Programming Perspective

### Numerical Limit Estimation

```javascript
function estimateLimit(f, a, epsilon = 1e-6) {
  // Approach from both sides
  const leftValue = f(a - epsilon);
  const rightValue = f(a + epsilon);
  
  // If close, return average
  if (Math.abs(leftValue - rightValue) < epsilon) {
    return (leftValue + rightValue) / 2;
  }
  
  return null;  // Limit doesn't exist or needs smaller epsilon
}

// Example: lim (x² - 4)/(x - 2) as x → 2
const f = x => (x*x - 4) / (x - 2);
estimateLimit(f, 2);  // ≈ 4
```

### Asymptotic Analysis (Big-O)

**Limits describe algorithm behavior as n → ∞**:

```
T(n) = 3n² + 100n + 5000

lim T(n)/n² = 3
n→∞

So T(n) is O(n²) (quadratic)
```

**Understanding**:
```
For large n, only highest-order term matters
Just like polynomial limits at infinity
```

---

## 8. Real-World Applications

### Instantaneous Velocity

**Average velocity**:
```
        distance
v_avg = ────────
         time
```

**Instantaneous velocity**: Average velocity over an infinitely small time interval

```
       s(t + h) - s(t)
v(t) = lim ──────────────
       h→0        h
```

This is the **derivative** (next chapter).

### Marginal Cost

**Average cost of next unit**:
```
                C(x + 1) - C(x)
Marginal cost = lim ───────────────
                x→0        1
```

### Tangent Lines

**Slope of tangent** = limit of slopes of secant lines:

```
       f(x + h) - f(x)
m = lim ───────────────
    h→0        h
```

---

## Common Mistakes & Misconceptions

### ❌ "The limit equals the function value"
Not always:
```
       x² - 1
f(x) = ──────
       x - 1

f(1) is undefined, but lim f(x) = 2
                         x→1
```

### ❌ "If f(a) doesn't exist, the limit doesn't exist"
The limit can exist even if f(a) doesn't:
```
f(x) = (x-2)/(x-2)  with hole at x=2

lim f(x) = 1, but f(2) undefined
x→2
```

### ❌ "Limits always give a number"
Limits can be ∞, -∞, or not exist at all.

### ❌ "lim (f+g) = lim f + lim g always"
Only if both limits exist individually.

### ❌ "Approaching from one side is enough"
Need both sides to match for the limit to exist.

---

## Tiny Practice

**Evaluate**:
1. lim (3x + 2)
   x→4

2. lim (x² - 9)/(x - 3)
   x→3

3. lim (x³ + 2x)
   x→0

4. lim 5
   x→100

**Determine if continuous at the given point**:
5. f(x) = x² at x = 2
6. f(x) = 1/x at x = 0

**Limits at infinity**:
7. lim (3x + 1)/(x - 2)
   x→∞

8. lim (x² + 5)/(x + 1)
   x→∞

<details>
<summary>Answers</summary>

1. 14 (direct substitution)
2. 6 (factor: (x+3)(x-3)/(x-3) = x+3)
3. 0 (direct substitution)
4. 5 (constant function)
5. Yes (polynomial, continuous everywhere)
6. No (undefined at x=0, vertical asymptote)
7. 3 (highest powers: 3x/x = 3)
8. ∞ (degree of numerator > denominator)

</details>

---

## Summary Cheat Sheet

### Definition

```
lim f(x) = L
x→a

"f(x) approaches L as x approaches a"
```

### One-Sided Limits

```
lim f(x)  (from left)
x→a⁻

lim f(x)  (from right)
x→a⁺

Limit exists if both equal
```

### Continuity

```
f is continuous at a if:
lim f(x) = f(a)
x→a
```

### Evaluation Techniques

| Case | Method |
|------|--------|
| Continuous | Direct substitution |
| 0/0 form | Factor and simplify |
| ∞/∞ form | Divide by highest power |
| Square roots | Rationalize |

### Limits at Infinity

```
       aₘxᵐ + ...
lim ───────────────
x→∞    bₙxⁿ + ...

m < n: → 0
m = n: → aₘ/bₙ
m > n: → ±∞
```

### Key Limits

```
lim sin(x)/x = 1
x→0

lim (1 + 1/n)ⁿ = e
n→∞
```

### Programming

```javascript
// Numerical estimation
function limit(f, a, h = 1e-6) {
  return (f(a + h) - f(a - h)) / (2*h);
}

// Check continuity
function isContinuous(f, a, epsilon = 1e-6) {
  try {
    const limit = (f(a + epsilon) + f(a - epsilon)) / 2;
    return Math.abs(f(a) - limit) < epsilon;
  } catch {
    return false;
  }
}
```

---

## Next Steps

Limits are the foundation of calculus. You now understand:
- Approaching vs reaching
- Evaluating limits
- Continuity
- Behavior at infinity

Next, we'll use limits to define **Derivatives**—the mathematics of instantaneous change.

**Continue to**: [11-derivatives.md](11-derivatives.md)
