# Integrals

## Why This Matters

**Integrals measure accumulation.** They answer:

> "How much total change happened?"

Integrals are everywhere:
- **Physics**: Distance from velocity, work from force
- **Statistics**: Probability, expected values, distributions
- **Economics**: Total cost, consumer surplus
- **Data science**: Area under ROC curve, cumulative distributions
- **Engineering**: Signal processing, control systems

Understanding integrals means understanding **total quantities from rates**.

---

## The Big Picture: From Rates to Totals

### The Fundamental Question

**Given**: Rate of change (derivative)
**Find**: Total amount (original function)

```
Speed (mph) → Total distance traveled
Flow rate (gal/min) → Total water
Marginal cost → Total cost
```

### Derivative vs Integral

```
         differentiation
f(x) ───────────────────────→ f'(x)
     ←─────────────────────── 
         integration
```

**Integrals "undo" derivatives** (with a twist).

---

## 1. The Accumulation Concept

### Visual: Area Under Curve

**Definite integral from a to b**:
```
  y
  │     ╱╲
  │    ╱  ╲
  │   ╱    ╲
  │  ╱▓▓▓▓▓▓╲
──┴────────────► x
  a          b

∫[a to b] f(x)dx = shaded area
```

### Riemann Sums: Building Intuition

**Approximate area with rectangles**:

```
  │
  │   ││││││││
  │   ││││││││
  │   ││││││││
──┴────────────
  a          b

Area ≈ f(x₁)Δx + f(x₂)Δx + ... + f(xₙ)Δx
```

**As rectangles get thinner** (Δx → 0), approximation becomes exact:

```
           n
∫[a to b] f(x)dx = lim Σ f(xᵢ)Δx
                  n→∞ i=1
```

This limit of sums is the **definite integral**.

---

## 2. Definite vs Indefinite Integrals

### Definite Integral

```
∫[a to b] f(x)dx

- Has limits: a and b
- Gives a NUMBER (the accumulated total)
- Represents area, total change, etc.
```

**Example**:
```
∫[0 to 3] x² dx = 9  (units depend on context)
```

### Indefinite Integral

```
∫ f(x)dx = F(x) + C

- No limits
- Gives a FUNCTION (antiderivative)
- C is an arbitrary constant
```

**Example**:
```
∫ x² dx = x³/3 + C
```

**The "+ C" is crucial** because derivatives of constants are zero:
```
d/dx(x³/3) = x²
d/dx(x³/3 + 5) = x²
d/dx(x³/3 + C) = x²  for any C
```

---

## 3. The Fundamental Theorem of Calculus

### Part 1: Connecting Derivative and Integral

**If F'(x) = f(x), then**:
```
∫[a to b] f(x)dx = F(b) - F(a)
```

**In words**: 
- To find definite integral, find antiderivative F
- Evaluate at endpoints: F(b) - F(a)

**Example**: ∫[1 to 3] x² dx
```
Antiderivative: F(x) = x³/3

∫[1 to 3] x² dx = F(3) - F(1)
                = 27/3 - 1/3
                = 9 - 1/3
                = 26/3
```

### Part 2: Derivative of an Integral

```
d/dx[∫[a to x] f(t)dt] = f(x)
```

**Integration and differentiation are inverse operations.**

---

## 4. Basic Integration Rules

### Power Rule (Reverse of Derivative)

```
∫ xⁿ dx = xⁿ⁺¹/(n+1) + C  (if n ≠ -1)
```

**Examples**:
```
∫ x³ dx = x⁴/4 + C
∫ x dx = x²/2 + C
∫ 1 dx = x + C
```

**Special case** (n = -1):
```
∫ 1/x dx = ln|x| + C
```

### Constant Multiple

```
∫ cf(x)dx = c∫ f(x)dx
```

**Example**:
```
∫ 5x² dx = 5∫ x² dx = 5·x³/3 + C = 5x³/3 + C
```

### Sum/Difference

```
∫ [f(x) + g(x)]dx = ∫ f(x)dx + ∫ g(x)dx
```

**Example**:
```
∫ (x² + 3x - 5)dx = x³/3 + 3x²/2 - 5x + C
```

---

## 5. Common Antiderivatives

### Polynomials

```
∫ xⁿ dx = xⁿ⁺¹/(n+1) + C
```

### Exponential

```
∫ eˣ dx = eˣ + C

∫ aˣ dx = aˣ/ln(a) + C
```

### Logarithmic

```
∫ 1/x dx = ln|x| + C
```

### Trigonometric

```
∫ sin(x)dx = -cos(x) + C
∫ cos(x)dx = sin(x) + C
∫ sec²(x)dx = tan(x) + C
∫ 1/√(1-x²) dx = sin⁻¹(x) + C
∫ 1/(1+x²) dx = tan⁻¹(x) + C
```

---

## 6. Integration Techniques (Brief Overview)

### Substitution (Chain Rule in Reverse)

**For integrals like** ∫ f(g(x))·g'(x)dx:

```
Let u = g(x), then du = g'(x)dx

∫ f(g(x))·g'(x)dx = ∫ f(u)du
```

**Example**: ∫ 2x·sin(x²)dx
```
Let u = x², du = 2x dx

∫ 2x·sin(x²)dx = ∫ sin(u)du
                = -cos(u) + C
                = -cos(x²) + C
```

### Integration by Parts (Product Rule in Reverse)

```
∫ u dv = uv - ∫ v du
```

**Example**: ∫ x·eˣ dx
```
Let u = x, dv = eˣdx
Then du = dx, v = eˣ

∫ x·eˣ dx = x·eˣ - ∫ eˣ dx
          = x·eˣ - eˣ + C
          = eˣ(x - 1) + C
```

### Partial Fractions (For Rational Functions)

**Break complex fractions into simpler ones.**

```
     1           A        B
────────── = ────── + ──────
(x-1)(x+2)    x-1      x+2
```

Then integrate each separately.

---

## 7. Definite Integrals: Properties

### Basic Properties

```
∫[a to b] cf(x)dx = c∫[a to b] f(x)dx

∫[a to b] [f(x) + g(x)]dx = ∫[a to b] f(x)dx + ∫[a to b] g(x)dx

∫[a to a] f(x)dx = 0

∫[a to b] f(x)dx = -∫[b to a] f(x)dx
```

### Additivity

```
∫[a to b] f(x)dx + ∫[b to c] f(x)dx = ∫[a to c] f(x)dx
```

### Comparison

```
If f(x) ≤ g(x) on [a,b], then:
∫[a to b] f(x)dx ≤ ∫[a to b] g(x)dx
```

---

## 8. Applications of Integrals

### Area Between Curves

**Area between f(x) and g(x) from a to b**:
```
A = ∫[a to b] |f(x) - g(x)|dx
```

If f(x) ≥ g(x):
```
A = ∫[a to b] [f(x) - g(x)]dx
```

**Example**: Area between y = x² and y = x from 0 to 1
```
A = ∫[0 to 1] (x - x²)dx
  = [x²/2 - x³/3][0 to 1]
  = 1/2 - 1/3
  = 1/6
```

### Distance from Velocity

**If v(t) is velocity, total distance is**:
```
distance = ∫[t₁ to t₂] v(t)dt
```

**Example**: v(t) = 3t² from t = 0 to t = 2
```
distance = ∫[0 to 2] 3t² dt
         = [t³][0 to 2]
         = 8 - 0
         = 8 units
```

### Total Cost from Marginal Cost

**If MC(x) is marginal cost**:
```
Total cost = Fixed cost + ∫[0 to x] MC(q)dq
```

**Example**: MC(x) = 2x + 5, fixed cost = $100
```
Total cost = 100 + ∫[0 to x] (2q + 5)dq
           = 100 + [q² + 5q][0 to x]
           = 100 + x² + 5x
```

### Average Value

**Average value of f on [a,b]**:
```
         1      b
f_avg = ─── ∫[a to ] f(x)dx
        b-a
```

**Example**: Average of f(x) = x² on [0, 3]
```
         1   ∫[0 to 3] x² dx
f_avg = ─── 
         3
       
       = 1/3 · [x³/3][0 to 3]
       = 1/3 · 9
       = 3
```

### Probability (Area = 1)

**For probability density function f(x)**:
```
P(a ≤ X ≤ b) = ∫[a to b] f(x)dx

∫[-∞ to ∞] f(x)dx = 1  (total probability)
```

### Work and Energy

**Work = force × distance**

If force varies:
```
W = ∫[a to b] F(x)dx
```

**Example**: Spring with F(x) = kx from 0 to d
```
W = ∫[0 to d] kx dx
  = [kx²/2][0 to d]
  = kd²/2
```

---

## 9. Numerical Integration (Programming)

### Trapezoidal Rule

**Approximate area using trapezoids**:
```
∫[a to b] f(x)dx ≈ (b-a)/2 · [f(a) + f(b)]
```

**Better with multiple intervals**:
```javascript
function trapezoidalRule(f, a, b, n) {
  const h = (b - a) / n;
  let sum = (f(a) + f(b)) / 2;
  
  for (let i = 1; i < n; i++) {
    sum += f(a + i*h);
  }
  
  return h * sum;
}

// Example: ∫[0 to 1] x² dx (exact: 1/3)
const f = x => x**2;
trapezoidalRule(f, 0, 1, 100);  // ≈ 0.33335
```

### Simpson's Rule

**More accurate (uses parabolas)**:
```javascript
function simpsonsRule(f, a, b, n) {
  // n must be even
  const h = (b - a) / n;
  let sum = f(a) + f(b);
  
  for (let i = 1; i < n; i++) {
    const coeff = (i % 2 === 0) ? 2 : 4;
    sum += coeff * f(a + i*h);
  }
  
  return (h / 3) * sum;
}

simpsonsRule(f, 0, 1, 100);  // ≈ 0.333333333
```

### Monte Carlo Integration

**Use random sampling**:
```javascript
function monteCarloIntegrate(f, a, b, numSamples) {
  let sum = 0;
  
  for (let i = 0; i < numSamples; i++) {
    const x = a + Math.random() * (b - a);
    sum += f(x);
  }
  
  return (b - a) * sum / numSamples;
}

monteCarloIntegrate(f, 0, 1, 10000);  // ≈ 0.333
```

**Useful for high-dimensional integrals** (where grid methods fail).

---

## 10. Integrals as Array Reduction

### The reduce() Analogy

**JavaScript reduce is like discrete integration**:

```javascript
// Sum array elements (discrete integral)
const values = [1, 2, 3, 4, 5];
const total = values.reduce((acc, val) => acc + val, 0);
// total = 15

// This is like: ∫ values dx ≈ Σ values[i]
```

### Cumulative Sum (Running Integral)

```javascript
function cumulativeSum(arr) {
  let cumsum = [0];
  for (let i = 0; i < arr.length; i++) {
    cumsum.push(cumsum[i] + arr[i]);
  }
  return cumsum;
}

cumulativeSum([1, 2, 3, 4]);  // [0, 1, 3, 6, 10]

// Like: F(x) = ∫[0 to x] f(t)dt
```

### From Rates to Totals

```javascript
// Velocities at each second
const velocities = [10, 15, 20, 25, 30];  // m/s

// Total distance (trapezoidal approximation)
let distance = 0;
for (let i = 0; i < velocities.length - 1; i++) {
  distance += (velocities[i] + velocities[i+1]) / 2 * 1;  // 1 sec intervals
}
// distance ≈ ∫ v(t) dt
```

---

## 11. Improper Integrals (Infinite Limits)

### Infinite Upper Limit

```
∫[a to ∞] f(x)dx = lim[b→∞] ∫[a to b] f(x)dx
```

**Example**: ∫[1 to ∞] 1/x² dx
```
= lim[b→∞] [-1/x][1 to b]
= lim[b→∞] (-1/b + 1)
= 0 + 1
= 1  (converges)
```

**But**: ∫[1 to ∞] 1/x dx diverges (goes to ∞)

### When They Converge

```
∫[1 to ∞] 1/xᵖ dx converges if p > 1
                  diverges if p ≤ 1
```

---

## 12. Connection to Other Concepts

### Integration and Probability

**Cumulative Distribution Function (CDF)**:
```
F(x) = ∫[-∞ to x] f(t)dt

where f(t) is probability density function (PDF)
```

### Integration and Machine Learning

**Loss over dataset**:
```
Total loss = ∫ L(f(x), y)·p(x,y) dx dy

In practice: Average over samples
```

**Area Under ROC Curve (AUC)**:
```
AUC = ∫[0 to 1] TPR(FPR) d(FPR)
```

### Differential Equations

**Many solutions involve integrals**:
```
dy/dx = f(x)  →  y = ∫ f(x)dx
```

---

## Common Mistakes & Misconceptions

### ❌ "Forgetting the + C"
Indefinite integrals always have an arbitrary constant.
```
∫ x dx = x²/2 + C  (not just x²/2)
```

### ❌ "∫ f·g = (∫f)·(∫g)"
**No!** Integration doesn't distribute over multiplication.

### ❌ "∫ f/g = (∫f)/(∫g)"
**No!** Use substitution or other techniques.

### ❌ "Area is always positive"
```
∫[-1 to 1] x dx = 0  (areas cancel)

For geometric area, use: ∫ |f(x)|dx
```

### ❌ "Definite integral needs + C"
No! Definite integrals give numbers, not functions.

---

## Tiny Practice

**Find antiderivatives**:
1. ∫ (3x² - 2x + 1)dx
2. ∫ (x³ + 1/x)dx
3. ∫ eˣ dx
4. ∫ sin(x)dx
5. ∫ (2x + 1)³ · 2 dx  (hint: substitution)

**Evaluate definite integrals**:
6. ∫[0 to 2] x² dx
7. ∫[1 to e] (1/x) dx
8. ∫[-π to π] sin(x)dx

**Applications**:
9. Find area under y = x² from x = 0 to x = 3
10. If v(t) = 2t + 1, find distance from t = 0 to t = 3

<details>
<summary>Answers</summary>

1. x³ - x² + x + C
2. x⁴/4 + ln|x| + C
3. eˣ + C
4. -cos(x) + C
5. u = 2x+1, du = 2dx → ∫u³du = u⁴/4 + C = (2x+1)⁴/4 + C
6. [x³/3][0 to 2] = 8/3
7. [ln(x)][1 to e] = ln(e) - ln(1) = 1 - 0 = 1
8. [-cos(x)][-π to π] = -cos(π) + cos(-π) = 1 - 1 = 0
9. ∫[0 to 3] x² dx = [x³/3][0 to 3] = 9
10. ∫[0 to 3] (2t+1)dt = [t² + t][0 to 3] = 9 + 3 = 12 units

</details>

---

## Summary Cheat Sheet

### Definitions

```
Definite:   ∫[a to b] f(x)dx = F(b) - F(a)  (number)

Indefinite: ∫ f(x)dx = F(x) + C  (function)
```

### Fundamental Theorem

```
∫[a to b] f(x)dx = F(b) - F(a)

where F'(x) = f(x)
```

### Key Rules

| Integral | Result |
|----------|--------|
| ∫ xⁿ dx | xⁿ⁺¹/(n+1) + C |
| ∫ 1/x dx | ln\|x\| + C |
| ∫ eˣ dx | eˣ + C |
| ∫ sin(x)dx | -cos(x) + C |
| ∫ cos(x)dx | sin(x) + C |
| ∫ cf(x)dx | c∫f(x)dx |
| ∫ [f+g]dx | ∫f dx + ∫g dx |

### Applications

```
Area: ∫[a to b] f(x)dx

Distance: ∫[t₁ to t₂] v(t)dt

Average: (1/(b-a))∫[a to b] f(x)dx

Work: ∫[a to b] F(x)dx
```

### Programming

```javascript
// Trapezoidal
const integrate = (f, a, b, n) => {
  const h = (b-a)/n;
  let sum = (f(a) + f(b))/2;
  for (let i = 1; i < n; i++) sum += f(a + i*h);
  return h * sum;
};

// Reduce analogy
const total = arr.reduce((sum, x) => sum + x, 0);
```

---

## Congratulations! 🎉

You've completed the entire mathematics curriculum from **numbers to calculus**!

You now understand:
- ✓ Number systems and arithmetic
- ✓ Algebraic manipulation
- ✓ Functions and their properties
- ✓ Coordinate geometry
- ✓ Trigonometry
- ✓ Limits and continuity
- ✓ Derivatives (rates of change)
- ✓ Integrals (accumulation)

### What's Next?

**Keep practicing**:
- Work through problems in each chapter
- Apply concepts to programming projects
- Explore Khan Academy, 3Blue1Brown, or Brilliant

**Advanced topics** (when ready):
- Multivariable calculus (functions of x, y, z)
- Differential equations (modeling change)
- Linear algebra (vectors, matrices, transformations)
- Real analysis (rigorous foundations)
- Probability and statistics

**Return to**: [README.md](README.md) to review any topics

---

**You've built a solid mathematical foundation. Now go apply it!** 🚀
