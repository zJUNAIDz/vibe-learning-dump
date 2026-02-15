# How Numbers Actually Work

## Why This Matters

Numbers are the atoms of mathematics. Before you can understand equations, functions, or calculus, you need to *really* understand what numbers are, why they exist, and how they behave.

If you've ever wondered:
- Why can't you divide by zero?
- What's the difference between -5 and 5 besides the sign?
- Why do fractions and decimals represent the same thing?
- What does √2 even mean?

...then you're asking the right questions. Let's answer them from scratch.

## The Big Picture: What Problem Do Numbers Solve?

**Numbers exist to measure and compare things.**

- "How many apples?" → Natural numbers (1, 2, 3...)
- "How much water?" → Fractions (1/2 cup, 3/4 liter)
- "What's the temperature?" → Negative numbers (-5°C)
- "How far exactly?" → Irrational numbers (√2 meters)

Each type of number was invented to solve a specific problem humans faced.

---

## 1. Natural Numbers (Counting Numbers)

### What They Are
**1, 2, 3, 4, 5, 6, ...**

These are the first numbers humans invented. You can count them on your fingers. They answer "how many?"

### Mental Model
Think of natural numbers as **discrete items in an array**:
```javascript
const apples = [🍎, 🍎, 🍎];
console.log(apples.length); // 3
```

You can't have 2.5 apples in this array—either you have 2 or 3.

### What You Can Do
- **Add**: 3 + 2 = 5 (combine two groups)
- **Multiply**: 3 × 4 = 12 (repeated addition: 3 + 3 + 3 + 3)
- **Compare**: 5 > 3 (one is bigger)

### What You Can't Do
- **Subtract freely**: 3 - 5 = ? (You can't have -2 apples... yet)
- **Divide freely**: 5 ÷ 2 = ? (Not always a natural number)

This is where we hit the limits of natural numbers.

---

## 2. Integers (Whole Numbers Including Negatives)

### What They Are
**..., -3, -2, -1, 0, 1, 2, 3, ...**

Integers extend natural numbers to include:
- **Zero**: "nothing" or "the starting point"
- **Negatives**: "opposite direction" or "debt"

### Mental Model: The Number Line

```
        ←─────────────┼─────────────→
    -3  -2  -1   0   1   2   3
 (left/down/debt)   (right/up/assets)
```

The number line is your most important mental tool. It turns numbers into **positions** or **distances**.

### Why Negatives Exist

**Problem**: You have $10, then spend $15. How much do you have?

Natural numbers can't answer this. But integers can: **-$5** (you're in debt).

Think of negatives as:
- **Direction**: Moving left instead of right
- **Opposite**: The reverse of something
- **Debt vs. Assets**: Below zero

### Programming Analogy
```javascript
let balance = 10;
balance -= 15;
console.log(balance); // -5 (totally valid)
```

In code, negative numbers are just numbers. In early math education, they're treated as scary. They're not.

### Zero: The Starting Point

Zero is special:
- It means "nothing" (0 apples)
- It's the **origin** on the number line
- It's the boundary between positive and negative
- **0 + x = x** (adding zero does nothing — the identity element)

### Integer Operations

| Operation | Example | Intuition |
|-----------|---------|-----------|
| Add | 3 + (-5) = -2 | Move 3 right, then 5 left |
| Subtract | 3 - 5 = -2 | Same as 3 + (-5) |
| Multiply | 3 × (-2) = -6 | Flip direction (negative) |
| Divide | -6 ÷ 3 = -2 | Reverse of multiply |

### Rules for Negative Multiplication

- **Positive × Positive = Positive** (normal)
- **Positive × Negative = Negative** (flip direction)
- **Negative × Negative = Positive** (flip twice = back to original)

**Why does negative × negative = positive?**

Think of it as reversing a reversal:
- Facing forward (positive)
- Turn around (negative)
- Turn around again (negative again) → You're facing forward (positive)

Or programmatically:
```javascript
let direction = 1;     // forward
direction *= -1;       // -1 (backward)
direction *= -1;       // 1 (forward again)
```

---

## 3. Rational Numbers (Fractions)

### What They Are
Numbers that can be written as **one integer divided by another**: a/b (where b ≠ 0)

Examples: 1/2, 3/4, -5/6, 7/1 (which is just 7)

### Mental Model: Fractions as Division

**Don't think of fractions as weird symbols. Think of them as division operations that haven't been completed yet.**

```
5/2 = 5 ÷ 2 = 2.5
```

The fraction bar is just a division sign:
```
  5
  ─   means   5 ÷ 2
  2
```

### Why Fractions Exist

**Problem**: You have 1 pizza and 4 people. How much does each person get?

1 ÷ 4 = 1/4

Fractions let you represent **parts of a whole** or the result of division.

### Visual: The Pizza Model

```
Original Pizza:
┌─────────┐
│  Whole  │
│    1    │
└─────────┘

Divided among 4 people:
┌───┬───┐
│1/4│1/4│
├───┼───┤
│1/4│1/4│
└───┴───┘
```

Each person gets 1/4.

### Numerator and Denominator

```
  3  ← numerator (how many parts you have)
  ─
  4  ← denominator (how many parts make a whole)
```

**Programming Analogy**:
```javascript
const fraction = {
  numerator: 3,
  denominator: 4,
  toDecimal() {
    return this.numerator / this.denominator; // 0.75
  }
};
```

### Equivalent Fractions

**1/2 = 2/4 = 3/6 = 4/8 = ...**

They all represent the same **value**, just written differently.

Why? Because you can multiply both top and bottom by the same number:
```
1     1×2     2
─  =  ───  =  ─
2     2×2     4
```

Think of it like this:
```javascript
const a = 1/2;      // 0.5
const b = 2/4;      // 0.5
console.log(a === b); // true
```

### Simplifying Fractions

Find the **greatest common divisor (GCD)** and divide both parts:

```
6     6÷2     3
──  = ───  =  ─
8     8÷2     4
```

**Why simplify?** Smaller numbers are easier to work with.

### Operations on Fractions

#### Addition (Same Denominator)
```
1     2     1+2     3
─  +  ─  =  ───  =  ─
4     4      4      4
```

Easy: just add the numerators.

#### Addition (Different Denominators)
```
1     1     ?
─  +  ─  =  
2     3
```

Find a **common denominator** (lowest common multiple):
```
1     1     3     2     5
─  +  ─  =  ─  +  ─  =  ─
2     3     6     6     6
```

#### Multiplication
```
1     2     1×2     2     1
─  ×  ─  =  ───  =  ─  =  ─
3     4     3×4    12     6
```

Multiply tops, multiply bottoms.

**Intuition**: "2/4 of 1/3" means "take 1/3, then take half of *that*"

#### Division
```
1     2     1     4     1×4     4
─  ÷  ─  =  ─  ×  ─  =  ───  =  ─
3     4     3     2     3×2     6
```

Flip the second fraction and multiply. This is called **multiplying by the reciprocal**.

**Why?** Division is the inverse of multiplication. If you multiply by 2/4, you divide by 4/2.

---

## 4. Decimals (Another Way to Write Fractions)

### What They Are
Numbers with a decimal point: 0.5, 3.14, -2.75

**Key insight: Decimals are just fractions in disguise.**

```
0.5   =  5/10   =  1/2
0.75  =  75/100 =  3/4
0.333 =  333/1000 ≈ 1/3
```

### Place Value

```
  3 2 1 . 5 6 7
  │ │ │   │ │ │
  │ │ │   │ │ └─ 7/1000 (thousandths)
  │ │ │   │ └─── 6/100  (hundredths)
  │ │ │   └───── 5/10   (tenths)
  │ │ └───────── 1      (ones)
  │ └─────────── 2×10   (tens)
  └───────────── 3×100  (hundreds)
```

321.567 = 300 + 20 + 1 + 0.5 + 0.06 + 0.007

### Why Decimals?

Decimals are easier to:
- Type (0.5 vs 1/2)
- Compare (0.7 vs 0.65 — just compare digit by digit)
- Use in calculators

But fractions are better for:
- Exact values (1/3 is exact, 0.333... is approximate)
- Showing relationships

### Converting Between Fractions and Decimals

**Fraction → Decimal**: Just divide
```
3/4 = 3 ÷ 4 = 0.75
```

**Decimal → Fraction**: Use place value
```
0.75 = 75/100 = 3/4 (simplified)
```

### Terminating vs Repeating Decimals

**Terminating**: Ends after a certain point
```
1/2 = 0.5
1/4 = 0.25
```

**Repeating**: Goes on forever
```
1/3 = 0.333333...
1/7 = 0.142857142857...
```

We write repeating decimals with a bar:
```
1/3 = 0.3̄  (the 3 repeats)
```

---

## 5. Irrational Numbers (Numbers That Can't Be Fractions)

### What They Are
Numbers that **cannot** be written as a fraction of two integers.

Famous examples:
- **π** (pi) ≈ 3.14159...
- **√2** (square root of 2) ≈ 1.41421...
- **e** (Euler's number) ≈ 2.71828...

### Why They Exist

**Problem**: What's the diagonal of a 1×1 square?

Using the Pythagorean theorem: diagonal² = 1² + 1² = 2

So: diagonal = √2

But √2 cannot be written as a fraction. It's been proven mathematically.

### Visual: Why √2 is Irrational

```
┌─────────┐
│        /│
│  1   /  │
│    /    │ diagonal = √2 ≈ 1.414...
│  /      │
│/    1   │
└─────────┘
```

No matter how you try to express it as a ratio of whole numbers, you can't. The decimal goes on forever *without repeating*.

### Decimal Expansion

**Rational**: Eventually repeats
```
1/3  = 0.333333...  (repeats)
1/7  = 0.142857142857...  (repeats)
```

**Irrational**: Never repeats, goes on forever
```
√2 = 1.41421356237309504880168872420969807856967187537694...
π  = 3.14159265358979323846264338327950288419716939937510...
```

### Programming Note
```javascript
console.log(Math.sqrt(2));  // 1.4142135623730951
console.log(Math.PI);       // 3.141592653589793
```

Computers approximate irrationals with floating-point numbers. They can't store infinite decimals.

### Why This Matters

Irrational numbers show up everywhere:
- Circles (π)
- Right triangles (√2)
- Natural growth (e)

You can't avoid them, so embrace them as "numbers that can't be written as simple fractions."

---

## 6. The Complete Number Line

```
  Irrational   Rational    Integers    Natural
      ↓           ↓           ↓           ↓
  ────●─────────●─────────────●───────────●─────→
     -π   -2   -1   0  1/2  1   √2    2   3
           ↑
        All these are REAL NUMBERS
```

### Real Numbers

**Real numbers** = All the numbers on the number line

This includes:
- Natural numbers (1, 2, 3, ...)
- Zero (0)
- Negative integers (-1, -2, -3, ...)
- Fractions (1/2, 3/4, ...)
- Irrational numbers (π, √2, ...)

If you can point to it on the number line, it's a real number.

---

## 7. Absolute Value (Distance, Not Direction)

### What It Means

The **absolute value** of a number is its distance from zero, ignoring direction.

**Notation**: |x| (read as "absolute value of x")

```
|5|  = 5   (distance from 0 is 5)
|-5| = 5   (distance from 0 is 5)
|0|  = 0   (distance from 0 is 0)
```

### Visual

```
        5 units        5 units
   ←──────────┼──────────→
           -5   0   5
```

Both -5 and 5 are 5 units away from zero.

### Mental Model: Distance

Think of absolute value as `Math.abs()` in code:
```javascript
Math.abs(5);   // 5
Math.abs(-5);  // 5
Math.abs(0);   // 0
```

It strips away the sign and gives you the magnitude.

### When You Use It

- **Distance**: "How far apart are -3 and 2?"
  ```
  |-3 - 2| = |-5| = 5
  ```
- **Error/Difference**: "How much did I miss by?"
  ```
  |actual - expected|
  ```
- **Magnitude**: "How big is this, regardless of direction?"

### Rules

```
|-x| = |x|           (sign doesn't matter)
|x × y| = |x| × |y|  (distribute through multiplication)
|x| ≥ 0              (always non-negative)
|x| = 0  ⟺  x = 0   (only zero has distance zero)
```

---

## Common Mistakes & Misconceptions

### ❌ "Negative numbers are smaller than zero"
Not in terms of magnitude. -1000 is "more negative" than -1, but both are negative.

### ❌ "Fractions are different from decimals"
They're the same thing in different notation: 0.5 = 1/2

### ❌ "You can't divide by zero because it's zero"
You can't divide by zero because **it's undefined**. Division by zero would break all of mathematics.

Think about it:
```
10 ÷ 0 = ?
```
This would mean: "What number, times 0, gives 10?"
But **any number × 0 = 0**, so there's no answer.

### ❌ "Irrational numbers are rare"
Most numbers are irrational! Rationals are actually the rare ones.

### ❌ "Numbers are just symbols"
Numbers represent quantities, distances, and comparisons. They have meaning.

---

## Real-World Examples

### Money (Rationals and Negatives)
```
Balance: $50.25  (decimal/rational)
Withdrawal: -$75 (negative)
New balance: -$24.75 (debt)
```

### Temperature (Integers and Negatives)
```
Freezing: 0°C
Cold: -10°C
Hot: 35°C
```

### Distances (Rationals and Irrationals)
```
Diagonal of square: √2 meters (irrational)
Half the distance: 0.5 km (rational decimal)
```

### Programming (All Types)
```javascript
const items = 5;              // natural number
let offset = -10;             // negative integer
const ratio = 0.75;           // decimal/rational
const pi = Math.PI;           // irrational (approximated)
const distance = Math.abs(x); // absolute value
```

---

## Tiny Practice

Try these to test your understanding:

1. **Place on the number line**: -2, 0.5, 3, -1.5, 2
2. **Absolute value**: |-7|, |3|, |-0.5|
3. **Convert**: 3/4 to decimal, 0.2 to fraction
4. **True or false**: Is -5 an integer? Is 0.5 an integer?
5. **Simplify**: 6/8, 10/15
6. **Compute**: |-3 - 5|, |4 - 1|

<details>
<summary>Answers</summary>

1. Number line: -2 — -1.5 — 0 — 0.5 — 2 — 3
2. Absolute values: 7, 3, 0.5
3. 3/4 = 0.75, 0.2 = 2/10 = 1/5
4. True (yes, -5 is an integer), False (0.5 is not an integer)
5. 6/8 = 3/4, 10/15 = 2/3
6. |-8| = 8, |3| = 3

</details>

---

## Summary Cheat Sheet

| Type | Examples | What It Solves |
|------|----------|----------------|
| **Natural** | 1, 2, 3, ... | Counting |
| **Integers** | ..., -2, -1, 0, 1, 2, ... | Direction, debt |
| **Rational** | 1/2, 3/4, 0.75 | Parts of a whole |
| **Irrational** | π, √2, e | Exact geometric values |
| **Real** | All of the above | Everything on the number line |

### Key Concepts

- **Number line**: Visual representation of numbers as positions
- **Negative**: Opposite direction, below zero
- **Fraction = Division**: 3/4 means 3 ÷ 4
- **Decimal = Fraction**: 0.5 means 5/10 = 1/2
- **Absolute value**: Distance from zero (|x|)
- **Zero**: The origin, boundary, identity element

### Mental Models

- Natural numbers = items in an array
- Integers = positions on a line
- Fractions = slicing a whole into parts
- Decimals = place value system
- Irrationals = infinite non-repeating decimals
- Absolute value = `Math.abs()`

---

## Next Steps

Now that you understand what numbers *are*, you're ready to learn what you can **do** with them.

In the next section, we'll explore **Algebra Foundations**—how to use numbers as variables and solve for unknowns.

**Continue to**: [01-algebra-foundations.md](01-algebra-foundations.md)
