# Linear Functions

## Why This Matters

Linear functions are the **simplest and most fundamental functions**. They model:
- **Constant change**: Speed, pricing, growth
- **Relationships**: Cause and effect
- **Predictions**: Trends and forecasting

In programming and data science:
- **Linear regression**: Fitting lines to data
- **Time complexity**: O(n) algorithms
- **Interpolation**: Estimating between points
- **Animation**: Linear motion

Linear functions are your mental model for "steady change."

---

## The Big Picture: Constant Rate of Change

**Linear function**: Output changes by the same amount for each unit of input.

```
Every time x increases by 1, y increases by the same amount

x:  0   1   2   3   4
y:  1   3   5   7   9

Change: +2  +2  +2  +2  (constant)
```

**Graphically**: A straight line.

---

## 1. The Slope-Intercept Form

### The Standard Equation

```
y = mx + b

m = slope (rate of change)
b = y-intercept (starting value)
x = input (independent variable)
y = output (dependent variable)
```

This is the most important equation in algebra.

### Visual Breakdown

```
y = 2x + 3
    ↑   ↑
    │   └─ b (y-intercept) = 3
    └───── m (slope) = 2

    y
    │
  7 ├───────●  (2, 7)
  6 ├───────┊
  5 ├─────●    (1, 5)
  4 ├─────┊
  3 ├───●──────  (0, 3) ← y-intercept (b=3)
  2 ├───┊
  1 ├─●        (-1, 1)
  0 ├─┊
────┼─┴──────────► x
   -1 0  1  2

Slope m = 2: "rise 2, run 1"
```

### Programming Analogy

```javascript
// Linear function as code
function linearFunction(x, m, b) {
  return m * x + b;
}

// Or as object
const line = {
  slope: 2,
  intercept: 3,
  evaluate(x) {
    return this.slope * x + this.intercept;
  }
};

line.evaluate(5);  // 2(5) + 3 = 13
```

---

## 2. Understanding Slope (m)

### What Is Slope?

**Slope = Rate of change = Rise over run**

```
       Δy   y₂ - y₁
m = ────── = ───────
       Δx   x₂ - x₁
```

### Interpretation

**m = 2**: For every 1 unit right, go up 2 units

**m = -3**: For every 1 unit right, go down 3 units

**m = 0.5**: For every 1 unit right, go up 0.5 units

**m = 0**: Horizontal line (no change)

**m = undefined**: Vertical line (infinite change)

### Types of Slopes

```
Positive (m > 0):    Negative (m < 0):
      ╱                    ╲
     ╱ (rising)             ╲ (falling)
    ╱                        ╲

Zero (m = 0):        Undefined:
  ─────                    │
  (horizontal)             │ (vertical)
                           │
```

### Calculating Slope from Two Points

**Given**: (1, 3) and (4, 9)

```
m = (y₂ - y₁) / (x₂ - x₁)
  = (9 - 3) / (4 - 1)
  = 6 / 3
  = 2
```

### Real-World Slope Examples

**Speed**:
```
Distance = rate × time
d = rt

If r = 60 mph:
d = 60t

Slope m = 60 (60 miles per hour)
```

**Pricing**:
```
Cost = price_per_unit × quantity + fixed_cost
C = 5q + 20

Slope m = 5 (cost increases $5 per unit)
```

**Temperature conversion**:
```
F = (9/5)C + 32

Slope m = 9/5 = 1.8
(Fahrenheit changes 1.8° for each 1° Celsius)
```

---

## 3. Understanding Y-Intercept (b)

### What Is Y-Intercept?

**The y-value when x = 0** (where the line crosses the y-axis)

```
y = mx + b

When x = 0:
y = m(0) + b = b
```

### Visual

```
    y
    │
  5 ├───
  4 ├───
  3 ├●──────── This is b (y-intercept)
  2 ├─╱
  1 ├╱
  0 ├─────────► x
```

### Interpretation

**b = 3**: Starting value is 3 (when x = 0)

**b = -2**: Starting value is -2 (below origin)

**b = 0**: Line passes through origin

### Real-World Y-Intercept Examples

**Fixed cost**:
```
Total = variable_cost × units + fixed_cost
C = 5q + 100

b = 100 (fixed cost even if q = 0)
```

**Initial position**:
```
position = velocity × time + starting_position
s = 3t + 10

b = 10 (started at position 10)
```

**Base salary**:
```
Income = commission × sales + base_salary
I = 0.1s + 30000

b = 30,000 (salary even with zero sales)
```

---

## 4. Writing Linear Equations

### From Slope and Y-Intercept

**Given**: m = 3, b = -2

```
y = 3x - 2
```

Done! Just plug into y = mx + b.

### From Slope and a Point

**Given**: m = 2, point (3, 7)

**Method**: Use point-slope form
```
y - y₁ = m(x - x₁)

y - 7 = 2(x - 3)
y - 7 = 2x - 6
y = 2x + 1
```

### From Two Points

**Given**: (1, 3) and (4, 9)

**Step 1**: Find slope
```
m = (9 - 3) / (4 - 1) = 6/3 = 2
```

**Step 2**: Use point-slope form with either point
```
y - 3 = 2(x - 1)
y - 3 = 2x - 2
y = 2x + 1
```

### From a Graph

**Read directly**:
1. Find y-intercept (where line crosses y-axis): b
2. Count rise/run from any two points: m
3. Write y = mx + b

---

## 5. Special Cases

### Horizontal Lines

```
y = 5    (or y = b where m = 0)

    y
    │
  5 ├───────────── y = 5 (constant)
  4 ├─────────────
  3 ├─────────────
    │
────┴─────────────► x
```

**Characteristics**:
- Slope m = 0
- y never changes
- Form: y = b

**Example**: Temperature stays at 70°F all day.

### Vertical Lines

```
x = 3

    y
    │
    │   │
    │   │ x = 3
    │   │
────┴───│──► x
    0   3
```

**Characteristics**:
- Slope undefined (division by zero)
- x never changes
- **NOT a function** (fails vertical line test)
- Form: x = a

### Lines Through Origin

```
y = mx    (b = 0)

    y
    │  ╱
    │ ╱
    │╱
────●─────► x
   origin
```

**Examples**:
- Direct proportionality
- y = 2x (doubling)
- Distance = speed × time (starting from origin)

---

## 6. Parallel and Perpendicular Lines

### Parallel Lines

**Same slope, different y-intercepts**

```
y = 2x + 1
y = 2x - 3

Both have m = 2 (parallel)
```

**Visual**:
```
    y
    │   ╱
    │  ╱  ← y = 2x + 1
    │ ╱
    │╱   ← y = 2x - 3
────┴─────► x
```

**Never intersect** (same direction).

### Perpendicular Lines

**Slopes are negative reciprocals**:
```
m₁ × m₂ = -1

If m₁ = 2, then m₂ = -1/2
If m₁ = 3/4, then m₂ = -4/3
```

**Example**:
```
y = 2x + 1     (m = 2)
y = -½x + 3    (m = -1/2)

These are perpendicular (meet at 90°)
```

**Visual**:
```
    y
    │    ╱
    │   ╱ ← m = 2
────┼──╱────
    │╲
    │ ╲ ← m = -1/2
```

**Special case**: Horizontal and vertical lines are perpendicular
```
y = 5  (horizontal, m = 0)
x = 3  (vertical, m = undefined)
```

---

## 7. Applications and Examples

### Motion at Constant Speed

```
Distance = speed × time
d = 60t

m = 60 mph (speed)
b = 0 (starts at origin)

After 3 hours: d = 60(3) = 180 miles
```

### Linear Pricing

```
Cost = price_per_item × quantity + fixed_cost
C = 15q + 200

m = 15 (variable cost per unit)
b = 200 (fixed costs)

For 50 items: C = 15(50) + 200 = $950
```

### Temperature Conversion

**Celsius to Fahrenheit**:
```
F = (9/5)C + 32

m = 9/5 = 1.8
b = 32

At 0°C: F = 32°F
At 100°C: F = (9/5)(100) + 32 = 212°F
```

**Fahrenheit to Celsius** (inverse):
```
C = (5/9)(F - 32)
C = (5/9)F - 160/9

m = 5/9 ≈ 0.556
b = -160/9 ≈ -17.78
```

### Depreciation

```
Value = initial_value - depreciation_rate × years
V = 20000 - 2000t

m = -2000 (loses $2000/year)
b = 20000 (initial value)

After 5 years: V = 20000 - 2000(5) = $10,000
```

### Salary with Commission

```
Income = commission_rate × sales + base
I = 0.08s + 35000

m = 0.08 (8% commission)
b = 35000 (base salary)

With $100k sales: I = 0.08(100000) + 35000 = $43,000
```

---

## 8. Finding Intersections

### Where Two Lines Meet

**Solve the system of equations**:

```
Line 1: y = 2x + 1
Line 2: y = -x + 4

Set equal:
2x + 1 = -x + 4
3x = 3
x = 1

Substitute back:
y = 2(1) + 1 = 3

Intersection: (1, 3)
```

### Visual

```
    y
    │
  4 ├────●     Line 2
  3 ├───●●     Intersection (1, 3)
  2 ├──●─●
  1 ├─●──●     Line 1
────┼────────► x
    0  1  2
```

### Break-Even Analysis

**When do costs equal revenue?**

```
Cost: C = 5q + 1000    (production cost)
Revenue: R = 10q       (sales income)

Break-even: C = R
5q + 1000 = 10q
1000 = 5q
q = 200 units

At 200 units, cost = revenue = $2000
```

---

## 9. Linear Regression (Data Fitting)

### The Problem

Given data points, find the "best fit" line.

**Data**: (1, 2), (2, 4), (3, 5), (4, 7)

**Goal**: Find y = mx + b that best approximates the points.

### Least Squares Method (Intuition)

**Minimize the total squared error** between predicted and actual values.

**Result**: Formulas for m and b:
```
m = (n·Σxy - Σx·Σy) / (n·Σx² - (Σx)²)
b = (Σy - m·Σx) / n

where n = number of points
```

(The math is complex, but computers do it instantly.)

### Programming

```javascript
function linearRegression(points) {
  const n = points.length;
  const sumX = points.reduce((sum, p) => sum + p.x, 0);
  const sumY = points.reduce((sum, p) => sum + p.y, 0);
  const sumXY = points.reduce((sum, p) => sum + p.x * p.y, 0);
  const sumX2 = points.reduce((sum, p) => sum + p.x * p.x, 0);
  
  const m = (n * sumXY - sumX * sumY) / (n * sumX2 - sumX * sumX);
  const b = (sumY - m * sumX) / n;
  
  return { slope: m, intercept: b };
}

const data = [
  {x: 1, y: 2},
  {x: 2, y: 4},
  {x: 3, y: 5},
  {x: 4, y: 7}
];

const line = linearRegression(data);
// { slope: 1.7, intercept: 0.5 }
// Best fit: y = 1.7x + 0.5
```

### Use Cases

- **Trend analysis**: Sales over time
- **Predictions**: Forecasting
- **Correlation**: Relationship between variables
- **Machine learning**: Linear models

---

## 10. Inequalities with Lines

### Linear Inequalities

**Instead of y = mx + b, use <, >, ≤, ≥**

```
y > 2x + 1    (above the line)
y ≤ -x + 3    (on or below the line)
```

### Graphing Inequalities

**y > 2x + 1**:

```
    y
    │ ▓▓▓▓▓▓▓
    │▓▓▓▓╱▓▓    ← Shaded region (solutions)
    │▓▓╱▓▓▓
    │▓╱▓▓▓      ← Dashed line (not included)
────┼╱──────► x
```

**Steps**:
1. Graph the line y = 2x + 1
2. Dashed line if < or > (not included)
3. Solid line if ≤ or ≥ (included)
4. Shade above for >, ≥
5. Shade below for <, ≤

### Testing Points

**Check if a point satisfies the inequality**:

Is (2, 6) a solution to y > 2x + 1?
```
6 > 2(2) + 1
6 > 5
True ✓
```

Is (0, 0) a solution?
```
0 > 2(0) + 1
0 > 1
False ✗
```

---

## Common Mistakes & Misconceptions

### ❌ "Slope is always positive"
Slope can be negative, zero, or undefined.

### ❌ "Steep lines have small slopes"
**Opposite!** Steep lines have large absolute slopes:
```
m = 10: Very steep
m = 0.1: Very gradual
```

### ❌ "Parallel lines have the same equation"
Same slope, but different y-intercepts:
```
y = 2x + 1
y = 2x + 5
(parallel but different)
```

### ❌ "b is always positive"
Y-intercept can be negative:
```
y = 2x - 5  (b = -5)
```

### ❌ "All linear equations look like y = mx + b"
Other forms exist:
```
Standard form: Ax + By = C
Point-slope: y - y₁ = m(x - x₁)
```

### ❌ "Lines always have exactly one intersection"
- **Same line**: Infinite intersections
- **Parallel lines**: Zero intersections
- **Different non-parallel**: Exactly one

---

## Tiny Practice

**Write equations**:
1. Slope 3, y-intercept -2
2. Slope -1/2, passes through (0, 4)
3. Passes through (1, 5) and (3, 11)

**Find slope and y-intercept**:
4. y = -2x + 7
5. 3x + 2y = 6

**Determine**:
6. Are y = 3x + 1 and y = 3x - 2 parallel?
7. Are y = 2x + 1 and y = -½x + 3 perpendicular?

**Solve**:
8. Find intersection of y = 2x + 1 and y = x + 3
9. Convert 25°C to Fahrenheit using F = (9/5)C + 32

<details>
<summary>Answers</summary>

1. y = 3x - 2
2. y = -½x + 4
3. m = (11-5)/(3-1) = 3, then y = 3x + 2
4. m = -2, b = 7
5. Solve for y: y = -3/2 x + 3, so m = -3/2, b = 3
6. Yes (both have m = 3)
7. Yes (2 × -½ = -1)
8. x = 2, y = 5, intersection at (2, 5)
9. F = (9/5)(25) + 32 = 77°F

</details>

---

## Summary Cheat Sheet

### Standard Form

```
y = mx + b

m = slope (rate of change)
b = y-intercept (starting value)
```

### Finding Slope

```
       rise   Δy   y₂-y₁
m = ────── = ── = ─────
       run    Δx   x₂-x₁

Positive: ╱ rising
Negative: ╲ falling
Zero: ─ horizontal
Undefined: │ vertical
```

### Writing Equations

| Given | Method |
|-------|--------|
| m and b | y = mx + b |
| m and point | y - y₁ = m(x - x₁) |
| Two points | Find m, then use point-slope |

### Special Lines

```
Horizontal: y = b     (m = 0)
Vertical:   x = a     (m = undefined)
Origin:     y = mx    (b = 0)
```

### Relationships

```
Parallel: Same slope (m₁ = m₂)
Perpendicular: m₁ × m₂ = -1
```

### Programming

```javascript
// Function
const f = (x, m, b) => m * x + b;

// Object
const line = {
  slope: 2,
  intercept: 3,
  eval(x) { return this.slope * x + this.intercept; }
};

// Regression
const {slope, intercept} = linearRegression(data);
```

---

## Next Steps

Linear functions are your foundation for understanding all functions. You now grasp:
- Constant rate of change
- Slope and intercept
- Graphing and applications

Next, we'll explore **Trigonometry**—functions based on angles, rotation, and waves.

**Continue to**: [08-trigonometry.md](08-trigonometry.md)
