# Powers, Roots, and Exponents

## Why This Matters

Powers and exponents represent **repeated multiplication**, just like multiplication represents repeated addition. They're everywhere:
- Compound interest (money grows exponentially)
- Big-O notation (algorithm complexity: O(n²), O(2ⁿ))
- Data storage (1 KB = 2¹⁰ bytes)
- Squares and cubes (area, volume)
- Scientific notation (3 × 10⁸ m/s)

Understanding exponents unlocks understanding of growth, scaling, and geometric relationships.

---

## The Big Picture: Repeated Operations

```
Addition (repeated counting):
3 + 3 + 3 + 3 = 4 × 3 = 12

Multiplication (repeated addition):
3 × 3 × 3 × 3 = 3⁴ = 81

Exponentiation (repeated multiplication):
Power tower: 3^(3^3) = 3^27 = ...huge
```

Each operation is "one level up" from the previous.

---

## 1. Exponents: Repeated Multiplication

### What They Are

**Exponent notation**: bⁿ

```
b^n = b × b × b × ... × b  (n times)
 ↑   ↑
base exponent
```

Examples:
```
2³ = 2 × 2 × 2 = 8
5² = 5 × 5 = 25
10⁴ = 10 × 10 × 10 × 10 = 10,000
```

### Reading Exponents

- **2³**: "two to the third power" or "two cubed"
- **5²**: "five to the second power" or "five squared"
- **10⁴**: "ten to the fourth power"

### Why They Exist

**Problem**: Writing 2 × 2 × 2 × 2 × 2 × 2 × 2 × 2 is tedious.

**Solution**: 2⁸ (much more compact)

Exponents are **shorthand for repetition**.

### Visual: Geometric Meaning

#### Squaring (x²)

**Area of a square**:
```
Side length: 3

┌───┬───┬───┐
│ 1 │ 2 │ 3 │
├───┼───┼───┤
│ 4 │ 5 │ 6 │
├───┼───┼───┤
│ 7 │ 8 │ 9 │
└───┴───┴───┘

Area = 3² = 9 square units
```

That's why we call it "squared"—it's literally a square.

#### Cubing (x³)

**Volume of a cube**:
```
Side length: 2

    ┌─────┬─────┐
   /     /     /│
  ┌─────┬─────┐ │
 /     /     /│ │
┌─────┬─────┐ │ │
│     │     │ │/│
│     │     │/│ │
├─────┼─────┤ │/
│     │     │/
└─────┴─────┘

Volume = 2³ = 8 cubic units
```

That's why we call it "cubed"—it's literally a cube.

### Programming Analogy

```javascript
// Exponent as loop
function power(base, exponent) {
  let result = 1;
  for (let i = 0; i < exponent; i++) {
    result *= base;
  }
  return result;
}

power(2, 3);  // 8

// Or use built-in
Math.pow(2, 3);  // 8
2 ** 3;          // 8 (ES7 exponentiation operator)
```

---

## 2. Special Cases and Rules

### Zero Exponent

**Any number to the power of zero is 1**:
```
5⁰ = 1
100⁰ = 1
(-3)⁰ = 1
```

**Why?** Pattern recognition:
```
2³ = 8
2² = 4   (divided by 2)
2¹ = 2   (divided by 2)
2⁰ = 1   (divided by 2)
```

Each time you decrease the exponent by 1, you divide by the base.

### One Exponent

**Any number to the power of one is itself**:
```
5¹ = 5
100¹ = 100
```

This makes sense: "multiply 5 by itself once" = 5.

### Negative Exponents

**Negative exponent = reciprocal**:
```
2⁻³ = 1 / 2³ = 1/8

x⁻ⁿ = 1 / xⁿ
```

**Why?** Continue the pattern:
```
2³ = 8
2² = 4    (÷ 2)
2¹ = 2    (÷ 2)
2⁰ = 1    (÷ 2)
2⁻¹ = 1/2 (÷ 2)
2⁻² = 1/4 (÷ 2)
```

**Mental model**: Negative exponent "flips" the number:
```
5² = 25
5⁻² = 1/25
```

### Fractional Exponents (Preview)

**Fractional exponents = roots**:
```
x^(1/2) = √x      (square root)
x^(1/3) = ∛x      (cube root)
x^(2/3) = (∛x)²   (cube root, then squared)
```

We'll explore this more in the roots section.

---

## 3. Laws of Exponents

These rules make working with exponents much easier. They're not arbitrary—they come from the definition.

### Law 1: Multiplying Same Base

**When multiplying, add the exponents**:
```
xᵃ × xᵇ = x^(a+b)
```

**Example**:
```
2³ × 2² = (2×2×2) × (2×2) = 2⁵ = 32
```

Count the 2's: 3 + 2 = 5

**Why it works**:
```
x³ × x² = (x×x×x) × (x×x) = x×x×x×x×x = x⁵
```

### Law 2: Dividing Same Base

**When dividing, subtract the exponents**:
```
xᵃ / xᵇ = x^(a-b)
```

**Example**:
```
2⁵ / 2² = 32 / 4 = 8 = 2³
```

**Why it works**:
```
x⁵ / x² = (x×x×x×x×x) / (x×x) = x×x×x = x³
```

Cancel out pairs from top and bottom.

### Law 3: Power of a Power

**When raising a power to a power, multiply the exponents**:
```
(xᵃ)ᵇ = x^(a×b)
```

**Example**:
```
(2³)² = (8)² = 64 = 2⁶
```

**Why it works**:
```
(x³)² = x³ × x³ = x⁶
```

### Law 4: Power of a Product

**Distribute the exponent to each factor**:
```
(xy)ⁿ = xⁿ × yⁿ
```

**Example**:
```
(2 × 3)² = 6² = 36
2² × 3² = 4 × 9 = 36
```

**Why it works**:
```
(xy)³ = (xy) × (xy) × (xy) = x×x×x × y×y×y = x³y³
```

### Law 5: Power of a Quotient

**Distribute the exponent to numerator and denominator**:
```
(x/y)ⁿ = xⁿ / yⁿ
```

**Example**:
```
(2/3)² = 4/9

Check: 2²/3² = 4/9 ✓
```

### Summary Table

| Rule | Formula | Example |
|------|---------|---------|
| Multiply | xᵃ × xᵇ = x^(a+b) | 2³ × 2² = 2⁵ |
| Divide | xᵃ / xᵇ = x^(a-b) | 2⁵ / 2² = 2³ |
| Power of Power | (xᵃ)ᵇ = x^(a×b) | (2³)² = 2⁶ |
| Power of Product | (xy)ⁿ = xⁿyⁿ | (2×3)² = 2²×3² |
| Power of Quotient | (x/y)ⁿ = xⁿ/yⁿ | (2/3)² = 4/9 |

### Programming Application

```javascript
// These laws apply in code too
Math.pow(2, 3) * Math.pow(2, 2) === Math.pow(2, 5);  // true

// Bit shifting uses powers of 2
1 << 3  // 2³ = 8
1 << 5  // 2⁵ = 32
```

---

## 4. Roots: Undoing Powers

### What They Are

A **root** is the inverse operation of a power.

**Square root** (√): What number, when squared, gives you this?
```
√25 = 5    because 5² = 25
√9 = 3     because 3² = 9
√2 ≈ 1.414 because 1.414² ≈ 2
```

**Notation**:
```
√x   = square root (most common)
∛x   = cube root
∜x   = fourth root
ⁿ√x  = nth root
```

### Visual: Square Root as Side Length

```
Area = 25 square units
Side length = √25 = 5

┌─────┐
│  5  │ 5
│     │
└─────┘
   5
```

If you know the area, the square root gives you the side length.

### Cube Root

**What number, when cubed, gives you this?**
```
∛8 = 2     because 2³ = 8
∛27 = 3    because 3³ = 27
∛64 = 4    because 4³ = 64
```

### Fractional Exponent Notation

Roots can be written as fractional exponents:
```
√x = x^(1/2)
∛x = x^(1/3)
∜x = x^(1/4)
```

**Why?** It follows the power rules:
```
(x^(1/2))² = x^(1/2 × 2) = x¹ = x ✓
```

### Combining Roots and Powers

```
x^(2/3) means:
1. Take the cube root: ∛x
2. Then square it: (∛x)²

Or equivalently:
1. Square it first: x²
2. Then take cube root: ∛(x²)
```

**Example**:
```
8^(2/3) = (∛8)² = 2² = 4

Or: 8^(2/3) = ∛(8²) = ∛64 = 4
```

### Programming

```javascript
Math.sqrt(25);        // 5 (square root)
Math.pow(25, 0.5);    // 5 (same thing)
Math.cbrt(8);         // 2 (cube root)
Math.pow(8, 1/3);     // 2 (same thing)

// Fourth root
Math.pow(16, 1/4);    // 2 (∜16 = 2)

// General: nth root of x
Math.pow(x, 1/n);
```

---

## 5. Principal vs Multiple Roots

### The Square Root Issue

**Every positive number has TWO square roots**:
```
√25 = ±5

Because:
5² = 25   ✓
(-5)² = 25 ✓
```

**Convention**: The radical symbol √ means the **positive root** (principal root).
```
√25 = 5    (principal root)
```

If you want both, write:
```
x² = 25
x = ±5  (plus or minus 5)
```

### Odd vs Even Roots

**Even roots** (√, ∜, etc.):
- Only defined for non-negative numbers (in real numbers)
- √(-4) is not a real number (involves imaginary numbers)
- Always give positive results (principal root)

**Odd roots** (∛, ∜∜∜, etc.):
- Defined for all real numbers
- ∛(-8) = -2 (because (-2)³ = -8)
- Preserve the sign

---

## 6. Simplifying Radicals

### Perfect Squares

Some numbers are **perfect squares**:
```
1² = 1
2² = 4
3² = 9
4² = 16
5² = 25
6² = 36
...
10² = 100
```

These are easy to take the square root of.

### Simplifying Non-Perfect Squares

**Factor out perfect squares**:

```
√12 = √(4 × 3) = √4 × √3 = 2√3
√18 = √(9 × 2) = √9 × √2 = 3√2
√50 = √(25 × 2) = √25 × √2 = 5√2
```

**Method**:
1. Find the largest perfect square factor
2. Split the radical
3. Simplify

**Why it works**:
```
√(a × b) = √a × √b
```

### Rationalizing the Denominator

**Don't leave radicals in the denominator**:

```
1/√2  →  multiply top and bottom by √2

1/√2 × √2/√2 = √2/2
```

This is preferred because it's easier to approximate:
```
√2/2 ≈ 1.414/2 ≈ 0.707
```

---

## 7. Exponential Growth vs Polynomial Growth

### Polynomial Growth (Powers)

```
Linear:     y = x     (doubles when x doubles)
Quadratic:  y = x²    (quadruples when x doubles)
Cubic:      y = x³    (8× when x doubles)
```

**Graph intuition**:
```
x:  1   2   3   4   5
x²: 1   4   9  16  25  (getting steeper)
x³: 1   8  27  64 125  (even steeper)
```

### Exponential Growth (Base)

```
y = 2ˣ

x:  1   2   3   4   5
2ˣ: 2   4   8  16  32  (doubling each time)
```

**Key difference**:
- **Polynomial**: x increases, y increases by power
- **Exponential**: x increases, y multiplies by base

**Exponential grows much faster**:
```
At x = 10:
x² = 100
2ˣ = 1024

At x = 20:
x² = 400
2ˣ = 1,048,576  (exponential explodes)
```

### Big-O Notation

```
O(n)    = linear    (fast)
O(n²)   = quadratic (slower)
O(2ⁿ)   = exponential (very slow)
O(log n) = logarithmic (very fast - next chapter!)
```

**Why it matters**: Algorithm efficiency

```javascript
// O(n²) - nested loops
for (let i = 0; i < n; i++) {
  for (let j = 0; j < n; j++) {
    // n × n operations
  }
}

// O(2ⁿ) - exponential (bad!)
function fibonacci(n) {
  if (n <= 1) return n;
  return fibonacci(n-1) + fibonacci(n-2);  // doubles work each level
}
```

---

## 8. Real-World Applications

### Data Storage (Powers of 2)

```
1 KB = 2¹⁰ bytes = 1,024 bytes
1 MB = 2²⁰ bytes = 1,048,576 bytes
1 GB = 2³⁰ bytes = 1,073,741,824 bytes
```

Why powers of 2? Binary system (computers use base-2).

### Compound Interest (Exponential Growth)

```
A = P(1 + r)ⁿ

P = principal ($1000)
r = interest rate (5% = 0.05)
n = years
A = final amount

After 10 years:
A = 1000(1.05)¹⁰ ≈ $1,629
```

### Half-Life (Exponential Decay)

```
Remaining = Initial × (1/2)^(t / half-life)

If half-life is 5 years and t = 10:
Remaining = Initial × (1/2)² = Initial / 4

After 10 years, only 1/4 remains.
```

### Area and Volume

```
Square area: s²
Circle area: πr²
Cube volume: s³
Sphere volume: (4/3)πr³
```

### Distance Formula (Pythagorean Theorem)

```
c² = a² + b²
c = √(a² + b²)

Distance between points (x₁,y₁) and (x₂,y₂):
d = √((x₂-x₁)² + (y₂-y₁)²)
```

### Scientific Notation

```
Speed of light: 3 × 10⁸ m/s
Electron mass: 9.1 × 10⁻³¹ kg
```

Much easier than writing all the zeros.

---

## Common Mistakes & Misconceptions

### ❌ "(x + y)² = x² + y²"
**No!** You must expand:
```
(x + y)² = (x + y)(x + y) = x² + 2xy + y²
```

### ❌ "√(x² + y²) = x + y"
**No!** Roots don't distribute over addition:
```
√(3² + 4²) = √(9 + 16) = √25 = 5
But 3 + 4 = 7 ≠ 5
```

### ❌ "x⁰ = 0"
**No!** x⁰ = 1 (for x ≠ 0)

### ❌ "√4 = ±2"
**No!** √4 = 2 (principal root only)
The equation x² = 4 has solutions x = ±2, but √4 means +2.

### ❌ "2³ × 3³ = 6³"
**No!** Different bases don't combine:
```
2³ × 3³ = 8 × 27 = 216
6³ = 216 ✓ (happens to equal, but not by the rule)

But: 2³ × 3³ = (2×3)³ = 6³ (power of product rule)
```

---

## Tiny Practice

Simplify:

1. 2³ × 2⁴
2. 5⁶ / 5²
3. (3²)³
4. (2 × 5)³
5. 10⁰
6. 2⁻³
7. √36
8. ∛27
9. √18 (simplify)
10. 8^(2/3)

Evaluate:

11. What is the area of a square with side 7?
12. What is the side length of a square with area 64?
13. If 2ˣ = 32, what is x?
14. If x² = 49, what are the possible values of x?

<details>
<summary>Answers</summary>

1. 2⁷ = 128
2. 5⁴ = 625
3. 3⁶ = 729
4. 10³ = 1000
5. 1
6. 1/8
7. 6
8. 3
9. 3√2
10. 4 (∛8² = 2² = 4)
11. 49
12. 8
13. x = 5 (2⁵ = 32)
14. x = ±7

</details>

---

## Summary Cheat Sheet

### Exponent Basics

```
xⁿ = x × x × ... × x  (n times)

x⁰ = 1
x¹ = x
x⁻ⁿ = 1/xⁿ
x^(1/n) = ⁿ√x
```

### Exponent Laws

| Operation | Rule | Example |
|-----------|------|---------|
| Multiply | xᵃ · xᵇ = x^(a+b) | 2³ · 2² = 2⁵ |
| Divide | xᵃ / xᵇ = x^(a-b) | 2⁵ / 2² = 2³ |
| Power of Power | (xᵃ)ᵇ = x^(ab) | (2³)² = 2⁶ |
| Power of Product | (xy)ⁿ = xⁿyⁿ | (2·3)² = 4·9 |
| Power of Quotient | (x/y)ⁿ = xⁿ/yⁿ | (2/3)² = 4/9 |

### Roots

```
√x  = x^(1/2)  (square root)
∛x  = x^(1/3)  (cube root)
ⁿ√x = x^(1/n)  (nth root)

√(x²) = |x|    (absolute value for real numbers)
(√x)² = x      (when x ≥ 0)
```

### Growth Comparison

```
Polynomial:   y = xⁿ   (faster as n increases)
Exponential:  y = aˣ   (much faster than polynomial)

O(n)    < O(n²)   < O(n³)   < O(2ⁿ)
linear    quadratic  cubic     exponential
```

### Perfect Squares to Memorize

```
1²=1   2²=4   3²=9    4²=16   5²=25
6²=36  7²=49  8²=64   9²=81   10²=100
11²=121  12²=144  13²=169  14²=196  15²=225
```

---

## Next Steps

You now understand powers, roots, and exponents—how repeated multiplication works and how to undo it. This foundation is critical for the next topic.

Next, we'll explore **Logarithms**—the inverse of exponentials, and one of the most powerful tools in mathematics and computer science.

**Continue to**: [04-logarithms.md](04-logarithms.md)
