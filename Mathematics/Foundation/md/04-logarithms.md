# Logarithms

## Why This Matters

Logarithms are the **inverse of exponentials**. They answer the question: "What power do I need to raise this base to get this number?"

As a developer, logarithms are absolutely everywhere:
- **Big-O notation**: O(log n) is super fast
- **Binary search**: Halving search space repeatedly
- **Compression**: JPEG, MP3 use logarithmic scales
- **Sound**: Decibels are logarithmic
- **Chemistry**: pH scale
- **Data structures**: Tree height is log(n)
- **Cryptography**: Key sizes (2048-bit = 2^2048)

Logarithms turn multiplication into addition, division into subtraction, and exponentials into multiplication. They're one of the most powerful tools in mathematics.

---

## The Big Picture: Undoing Exponents

### The Relationship

```
If: 2³ = 8
Then: log₂(8) = 3

Read as: "log base 2 of 8 equals 3"
Meaning: "2 to what power gives 8? Answer: 3"
```

**General form**:
```
If: bˣ = y
Then: log_b(y) = x

b = base
x = exponent
y = result
```

### Mental Model: Inverse Operations

```
Exponent asks:     "What is 2³?"           → 8
Logarithm asks:    "2 to what power is 8?" → 3

Square asks:       "What is 5²?"           → 25
Square root asks:  "What squared is 25?"   → 5

Multiply asks:     "What is 3 × 4?"        → 12
Divide asks:       "What times 4 is 12?"   → 3
```

Logarithms are as fundamental as square roots or division—they're just the inverse of a different operation.

---

## 1. Basic Logarithm Notation

### Standard Form

```
log_b(x) = y

b = base (subscript)
x = argument (what you're taking the log of)
y = result (the exponent)
```

**Examples**:
```
log₂(8) = 3      because 2³ = 8
log₁₀(100) = 2   because 10² = 100
log₃(27) = 3     because 3³ = 27
log₅(25) = 2     because 5² = 25
```

### Common Bases

**Base 10** (common logarithm):
```
log₁₀(x)  often written as  log(x)

log(100) = 2    because 10² = 100
log(1000) = 3   because 10³ = 1000
```

**Base 2** (binary logarithm):
```
log₂(x)  often written as  lg(x)  in CS

lg(8) = 3       because 2³ = 8
lg(1024) = 10   because 2¹⁰ = 1024
```

**Base e** (natural logarithm):
```
logₑ(x)  written as  ln(x)

e ≈ 2.71828...  (Euler's number)

ln(e) = 1       because e¹ = e
ln(e²) = 2      because e² = e²
```

### Programming

```javascript
Math.log10(100);   // 2 (base 10)
Math.log2(8);      // 3 (base 2)
Math.log(Math.E);  // 1 (natural log, base e)

// General base
function logBase(x, base) {
  return Math.log(x) / Math.log(base);
}

logBase(8, 2);  // 3
```

---

## 2. Why Logarithms Exist: The Huge Numbers Problem

### The Problem

**Question**: 2 to what power equals 1024?

You could try:
```
2¹ = 2
2² = 4
2³ = 8
2⁴ = 16
2⁵ = 32
...keep going...
2¹⁰ = 1024  ✓
```

But this is tedious. **Logarithms give you the answer directly**:
```
log₂(1024) = 10
```

### Logarithms Make Hard Problems Easy

**Without logs**: "What power of 2 gives 4,294,967,296?"
- You'd have to compute powers forever

**With logs**: log₂(4,294,967,296) = 32
- Instant answer

### The Scale Problem

Some quantities span huge ranges:
- Sound: From whisper (10⁻¹² W/m²) to jet engine (1 W/m²)
- Earthquakes: From 2.0 (barely felt) to 9.0 (catastrophic)
- Chemical acidity: pH 1 (strong acid) to pH 14 (strong base)

**Logarithmic scales compress these ranges** so they're manageable:
```
Linear scale:  1, 10, 100, 1000, 10000, ...
Log scale:     0,  1,   2,    3,     4, ...
```

---

## 3. Evaluating Logarithms

### Easy Cases (Powers You Know)

```
log₂(8) = ?
Think: 2 to what power is 8?
2³ = 8
Answer: 3

log₁₀(1000) = ?
Think: 10 to what power is 1000?
10³ = 1000
Answer: 3

log₅(1) = ?
Think: 5 to what power is 1?
5⁰ = 1
Answer: 0
```

### Special Values

```
log_b(1) = 0     because b⁰ = 1
log_b(b) = 1     because b¹ = b
log_b(bⁿ) = n    because bⁿ = bⁿ
```

**Examples**:
```
log₁₀(1) = 0
log₂(2) = 1
log₅(5³) = 3
```

### Fractional Results

Not all logs are whole numbers:
```
log₂(5) ≈ 2.32   because 2^2.32 ≈ 5
log₁₀(50) ≈ 1.70 because 10^1.70 ≈ 50
```

You need a calculator for most of these.

### Negative Results

```
log₂(1/2) = -1   because 2⁻¹ = 1/2
log₁₀(0.01) = -2 because 10⁻² = 0.01
```

**Pattern**: Fractions (values less than 1) give negative logs.

### Undefined Cases

```
log_b(0) = undefined  (no power of b gives 0)
log_b(negative) = undefined  (in real numbers)
```

---

## 4. Logarithm Rules (The Magic)

These rules make logarithms incredibly powerful. They turn multiplication/division into addition/subtraction.

### Rule 1: Log of a Product

**Multiplication becomes addition**:
```
log_b(xy) = log_b(x) + log_b(y)
```

**Example**:
```
log₂(8 × 4) = log₂(8) + log₂(4)
log₂(32) = 3 + 2
5 = 5 ✓
```

**Why it works**:
```
If x = b^m and y = b^n, then:
xy = b^m × b^n = b^(m+n)

So: log_b(xy) = m + n = log_b(x) + log_b(y)
```

**Programming analogy**:
```javascript
// Multiplying in linear space
const result = x * y;

// Adding in log space
const logResult = Math.log(x) + Math.log(y);
const result = Math.exp(logResult);  // convert back
```

### Rule 2: Log of a Quotient

**Division becomes subtraction**:
```
log_b(x/y) = log_b(x) - log_b(y)
```

**Example**:
```
log₂(8/2) = log₂(8) - log₂(2)
log₂(4) = 3 - 1
2 = 2 ✓
```

### Rule 3: Log of a Power

**Exponents become multiplication**:
```
log_b(xⁿ) = n × log_b(x)
```

**Example**:
```
log₂(8³) = 3 × log₂(8)
log₂(512) = 3 × 3
9 = 9 ✓
```

**This is huge**: It means you can pull exponents out front.
```
log₁₀(x¹⁰⁰) = 100 × log₁₀(x)
```

### Rule 4: Change of Base

**Convert between bases**:
```
log_b(x) = log_a(x) / log_a(b)
```

**Example**: Convert log₂(8) to base 10
```
log₂(8) = log₁₀(8) / log₁₀(2)
        = 0.903 / 0.301
        ≈ 3 ✓
```

**Why it's useful**: Calculators only have log₁₀ and ln, so you use this to compute other bases.

### Summary of Rules

| Operation | Logarithm Rule | Example |
|-----------|----------------|---------|
| Multiply | log(xy) = log(x) + log(y) | log(6) = log(2) + log(3) |
| Divide | log(x/y) = log(x) - log(y) | log(4) = log(8) - log(2) |
| Power | log(xⁿ) = n·log(x) | log(8) = 3·log(2) |
| Change Base | log_b(x) = log(x)/log(b) | log₂(8) = log(8)/log(2) |

---

## 5. Visual Intuition: Logarithmic Scales

### Linear vs Logarithmic

**Linear scale**: Even spacing
```
0──1──2──3──4──5──6──7──8──9──10
```

**Logarithmic scale**: Each step is a multiplication
```
1──10──100──1000──10000
↑   ×10  ×10  ×10   ×10
```

### Graph of y = log(x)

```
y
│
3├─────────────────●  (1000, 3)
2├────────────●       (100, 2)
1├───────●            (10, 1)
0├───●────────────    (1, 0)
-1├●                  (0.1, -1)
-2├                   (0.01, -2)
──┼────────────────────────► x
  0     1    10   100  1000
```

Key features:
- **Passes through (1, 0)**: log(1) = 0
- **Increases slowly**: Takes large x changes for small y changes
- **Negative for 0 < x < 1**: Fractions have negative logs
- **Undefined at x = 0**: Vertical asymptote
- **Never horizontal**: Always increasing (for base > 1)

### Comparison to y = eˣ

They're **inverses** (mirror images across y = x):

```
y=eˣ (exponential)      y=ln(x) (logarithm)
     │                      │
   5 ├───────────●        5 ├
   4 ├────────●           4 ├
   3 ├─────●              3 ├──────────────●
   2 ├───●                2 ├────────────●
   1 ├─●                  1 ├───────●
   0 ├●                   0 ├───●
  -1 ├                   -1 ├●
     └──────────► x          └──────────► x
     0  1  2  3               0  1  2  3
```

Notice:
- Exponential grows fast, log grows slow
- They're reflections across the line y = x

---

## 6. Real-World Applications

### Big-O Notation: O(log n)

**Binary search** is O(log n):
```
Array of 1,000,000 elements
Linear search: Up to 1,000,000 comparisons
Binary search: log₂(1,000,000) ≈ 20 comparisons

That's 50,000× faster!
```

**Why?** Each comparison cuts the problem in half:
```
1,000,000 → 500,000 → 250,000 → 125,000 → ... → 1

Number of halvings = log₂(1,000,000)
```

**Code**:
```javascript
function binarySearch(arr, target) {
  let left = 0, right = arr.length - 1;
  
  while (left <= right) {
    const mid = Math.floor((left + right) / 2);
    
    if (arr[mid] === target) return mid;
    if (arr[mid] < target) left = mid + 1;
    else right = mid - 1;
  }
  
  return -1;
}

// Time complexity: O(log n)
```

### Tree Height

**Balanced binary tree with n nodes has height log₂(n)**:
```
1 node:  height 0  (2⁰ = 1)
3 nodes: height 1  (2¹ - 1 = 1, plus 2 children)
7 nodes: height 2  (2² - 1 = 3, plus 4 grandchildren)

Height = ⌈log₂(n+1)⌉ - 1
```

**Why it matters**: Operations take O(height) time.

### Decibels (Sound)

```
Decibels (dB) = 10 × log₁₀(I / I₀)

I = intensity
I₀ = reference intensity (threshold of hearing)
```

**Examples**:
```
Whisper:     30 dB   (1,000× reference)
Conversation: 60 dB   (1,000,000× reference)
Jet engine:  140 dB   (10¹⁴× reference)
```

Each 10 dB increase = 10× louder.

### pH Scale (Chemistry)

```
pH = -log₁₀([H⁺])

[H⁺] = hydrogen ion concentration
```

**Examples**:
```
pH 7 (neutral):   [H⁺] = 10⁻⁷
pH 1 (strong acid): [H⁺] = 10⁻¹ (1,000,000× more acidic)
pH 14 (strong base): [H⁺] = 10⁻¹⁴
```

### Richter Scale (Earthquakes)

```
Magnitude = log₁₀(A / A₀)

A = amplitude
A₀ = reference amplitude
```

Each whole number increase = 10× more energy released.
```
Magnitude 5: Notable
Magnitude 6: 10× stronger
Magnitude 7: 100× stronger (destructive)
Magnitude 8: 1,000× stronger (major earthquake)
```

### Information Theory

**Bits needed to represent n items**:
```
bits = log₂(n)

8 items: log₂(8) = 3 bits
256 items: log₂(256) = 8 bits (1 byte)
```

**Entropy** (information content):
```
H = -Σ p(x) log₂(p(x))
```

### Compound Interest (Doubling Time)

```
A = P(1 + r)ᵗ

To find when money doubles:
2P = P(1 + r)ᵗ
2 = (1 + r)ᵗ
log(2) = t × log(1 + r)
t = log(2) / log(1 + r)

For 5% interest:
t = log(2) / log(1.05) ≈ 14 years
```

---

## 7. Solving Logarithmic and Exponential Equations

### Type 1: Solve for x in log(x) = n

```
log₂(x) = 5

Convert to exponential form:
x = 2⁵ = 32
```

**General**: If log_b(x) = n, then x = bⁿ

### Type 2: Solve for x in bˣ = n

```
2ˣ = 32

Take log of both sides:
log₂(2ˣ) = log₂(32)

Use log rule (power comes out):
x × log₂(2) = log₂(32)
x × 1 = 5
x = 5
```

**Or**: Recognize that 32 = 2⁵, so x = 5

### Type 3: Different Bases

```
3ˣ = 100

Take log of both sides (any base, use base 10):
log(3ˣ) = log(100)
x × log(3) = log(100)
x × log(3) = 2
x = 2 / log(3)
x ≈ 2 / 0.477 ≈ 4.19
```

### Type 4: Multiple Logs

```
log(x) + log(x-3) = 1

Use product rule:
log(x(x-3)) = 1

Convert to exponential (assuming base 10):
x(x-3) = 10¹
x² - 3x = 10
x² - 3x - 10 = 0

Factor:
(x-5)(x+2) = 0
x = 5 or x = -2

Check: x must be positive (can't log negative)
x = 5 ✓
```

---

## 8. Natural Logarithm (ln) and e

### Euler's Number (e)

```
e ≈ 2.71828...
```

**Definition**: The base of natural logarithm, defined by:
```
e = lim(n→∞) (1 + 1/n)ⁿ
```

Or:
```
e = 1 + 1/1! + 1/2! + 1/3! + 1/4! + ...
```

### Why e Is Special

**Natural growth/decay** uses e:
```
Continuous compound interest: A = Pe^(rt)
Population growth: P(t) = P₀e^(kt)
Radioactive decay: N(t) = N₀e^(-λt)
```

**Derivative property**: The derivative of eˣ is eˣ (unchanged!)

### Natural Logarithm

```
ln(x) = log_e(x)

ln(e) = 1
ln(e²) = 2
ln(1) = 0
```

**Connection**:
```
If y = eˣ, then x = ln(y)

They're inverses:
e^(ln(x)) = x
ln(e^x) = x
```

### Programming

```javascript
Math.E;           // 2.718281828459045
Math.log(Math.E); // 1 (natural log)
Math.exp(1);      // e (same as e¹)

// e^x
Math.exp(2);      // e² ≈ 7.389

// ln(x)
Math.log(10);     // ln(10) ≈ 2.303
```

---

## 9. Logarithmic Thinking

### Halving Problems

"How many times can you divide 1000 by 2 before reaching 1?"
```
1000 → 500 → 250 → 125 → 62.5 → 31.25 → ...

Answer: log₂(1000) ≈ 10 times
```

### Doubling Problems

"How many times must you double 1 to reach 1000?"
```
1 → 2 → 4 → 8 → 16 → 32 → 64 → 128 → 256 → 512 → 1024

Answer: log₂(1000) ≈ 10 times
```

### Order of Magnitude

**How big is this number?**
```
log₁₀(5000) ≈ 3.7

Interpretation: Between 10³ and 10⁴ (thousands)
```

### Scaling Intuition

If you double the input to a log function, how much does the output change?
```
log₂(10) ≈ 3.32
log₂(20) ≈ 4.32

Difference: 1 (always, regardless of starting value)

Doubling adds 1 to the log.
```

---

## Common Mistakes & Misconceptions

### ❌ "log(a + b) = log(a) + log(b)"
**No!** Logs don't distribute over addition.
```
log(a + b) ≠ log(a) + log(b)

log(10 + 10) = log(20) ≈ 1.30
log(10) + log(10) = 1 + 1 = 2 ≠ 1.30
```

**Correct**:
```
log(a × b) = log(a) + log(b)  (multiplication → addition)
```

### ❌ "log(x)/log(y) = log(x/y)"
**No!**
```
log(x)/log(y) = log_y(x)  (change of base)
log(x/y) = log(x) - log(y) (quotient rule)
```

### ❌ "log₂(8) = log₁₀(8)"
**No!** Different bases give different results:
```
log₂(8) = 3
log₁₀(8) ≈ 0.903
```

### ❌ "Logs can take negative inputs"
Not in real numbers:
```
log(-5) = undefined (no real answer)
```

### ❌ "log(0) = 0"
**No!**
```
log(1) = 0
log(0) = undefined (negative infinity)
```

---

## Tiny Practice

Evaluate:

1. log₂(16)
2. log₁₀(1000)
3. log₅(25)
4. log₃(1)
5. log₂(1/4)

Simplify using log rules:

6. log(5) + log(2)
7. log(100) - log(10)
8. 3 × log(2)
9. log(x³)
10. log₂(8) + log₂(4) - log₂(2)

Solve:

11. log₂(x) = 4
12. 2ˣ = 64
13. If log(x) = 3, what is x? (assume base 10)
14. How many times can you divide 512 by 2 before reaching 1?

<details>
<summary>Answers</summary>

1. 4 (2⁴ = 16)
2. 3 (10³ = 1000)
3. 2 (5² = 25)
4. 0 (3⁰ = 1)
5. -2 (2⁻² = 1/4)
6. log(10) = 1 (product rule)
7. log(10) = 1 (quotient rule)
8. log(2³) = log(8) ≈ 0.903
9. 3log(x) (power rule)
10. log₂(16) = 4
11. x = 16 (2⁴ = 16)
12. x = 6 (2⁶ = 64)
13. x = 1000 (10³ = 1000)
14. 9 times (log₂(512) = 9)

</details>

---

## Summary Cheat Sheet

### Definition

```
If bˣ = y, then log_b(y) = x

log_b(y) asks: "b to what power gives y?"
```

### Common Bases

```
log(x)    = log₁₀(x)  (common logarithm)
lg(x)     = log₂(x)   (binary logarithm)
ln(x)     = log_e(x)  (natural logarithm)
```

### Key Values

```
log_b(1) = 0
log_b(b) = 1
log_b(bⁿ) = n
```

### Logarithm Rules

| Rule | Formula | Intuition |
|------|---------|-----------|
| Product | log(xy) = log(x) + log(y) | Multiply → Add |
| Quotient | log(x/y) = log(x) - log(y) | Divide → Subtract |
| Power | log(xⁿ) = n·log(x) | Exponent → Multiply |
| Change Base | log_b(x) = log(x)/log(b) | Convert bases |

### Inverse Relationship

```
b^(log_b(x)) = x
log_b(bˣ) = x

e^(ln(x)) = x
ln(eˣ) = x
```

### Applications

- **Algorithm analysis**: O(log n)
- **Data structures**: Tree height
- **Sound**: Decibels
- **Chemistry**: pH
- **Earthquakes**: Richter scale
- **Information**: Bits needed

### Programming

```javascript
Math.log10(x);     // base 10
Math.log2(x);      // base 2
Math.log(x);       // natural (base e)

// Change of base
Math.log(x) / Math.log(base);
```

---

## Next Steps

You now understand logarithms—one of the most powerful tools in mathematics. They let you work with massive ranges of values, analyze algorithms, and understand growth/decay.

Next, we'll explore **Coordinate Geometry**—how to represent points, lines, and shapes on a grid.

**Continue to**: [05-coordinate-geometry.md](05-coordinate-geometry.md)
