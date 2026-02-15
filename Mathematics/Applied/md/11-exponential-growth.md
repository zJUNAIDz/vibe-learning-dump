# 11. Exponential Growth & Decay

**Phase 4: Math for Finance & Decision Making**  
⏱ ~40 minutes | 📈 Critical Pattern Recognition | 🚀 Scales Everything

---

## What Problem This Solves

You're observing:
- Database size growing "faster than expected"
- AWS costs doubling every 6 months
- GitHub stars going from 10 → 100 → 1000 seemingly overnight
- Technical debt making changes 2x harder each quarter
- A feature request that "can't possibly scale"

**Without exponential intuition**, linear thinking betrays you. You project linearly ("if we grow 10 users/month, we'll have 120 users next year"), but reality is exponential ("we doubled monthly, so we'll have 4,096 users next year"). You're blindsided by runaway costs, viral growth, or compounding decay.

**With exponential understanding**, you spot these patterns early, forecast correctly, and design systems that either leverage exponential growth (viral loops) or defend against exponential blowup (rate limiting, cost caps).

---

## Intuition & Mental Model

### The Folding Paper Analogy

```
Fold a paper in half:
Fold 1:  2 layers
Fold 2:  4 layers
Fold 3:  8 layers
...
Fold 42: 4,398,046,511,104 layers
         → Reaches from Earth to the Moon

Exponential growth = deceptively slow, then suddenly EXPLOSIVE
```

**Mental Model: The Hockey Stick**

```
Linear Growth:              Exponential Growth:
   /                                            /
  /                                           /
 /                                          /
/___________                        _______/
             Time                           Time
                                   ^ "Hockey stick"
```

Early exponential growth looks linear. Then it explodes.

**The Key Insight**:
```
Linear:      Each step adds a CONSTANT
             1, 2, 3, 4, 5, 6...
             
Exponential: Each step MULTIPLIES
             1, 2, 4, 8, 16, 32...
             
Same start, wildly different outcomes.
```

---

## Core Concepts

### 1. The Exponential Function

```javascript
function exponentialGrowth(initial, growthRate, time) {
  return initial * Math.pow(1 + growthRate, time);
}

// Startup users: 100 users, 15% monthly growth
exponentialGrowth(100, 0.15, 12);  // 535 users after 1 year

// Database size: 10GB, 20% monthly growth
exponentialGrowth(10, 0.20, 12);  // 89GB after 1 year

// Viral app: 1000 DAU, 10% daily growth
exponentialGrowth(1000, 0.10, 30);  // 17,449 DAU after 1 month
```

**Decay (Negative Growth)**:
```javascript
function exponentialDecay(initial, decayRate, time) {
  return initial * Math.pow(1 - decayRate, time);
}

// Cache hit rate degrading: 90% → ?
exponentialDecay(0.90, 0.05, 10);  // 59.9% after 10 periods

// Radioactive decay (half-life)
exponentialDecay(1000, 0.5, 5);  // 31.25 remaining after 5 half-lives
```

### 2. Doubling Time & Half-Life

**Doubling Time**: How long until quantity doubles?

```javascript
function doublingTime(growthRate) {
  // Formula: ln(2) / ln(1 + r)
  // Approximation: 70 / (rate × 100)  [Rule of 70]
  return Math.log(2) / Math.log(1 + growthRate);
}

// 10% monthly growth → doubles every:
doublingTime(0.10);  // 7.27 months

// 3% annual inflation → prices double in:
doublingTime(0.03);  // 23.45 years

// 100% daily growth (viral app) → doubles every:
doublingTime(1.0);  // 1 day (obviously)
```

**Rule of 70**: Quick mental math
```
Doubling Time ≈ 70 / (growth rate percentage)

5% → 70/5 = 14 periods
10% → 70/10 = 7 periods
20% → 70/20 = 3.5 periods
```

**Half-Life (Decay)**:
```javascript
function halfLife(decayRate) {
  return Math.log(0.5) / Math.log(1 - decayRate);
}

// Session cookies: 10% expire per day → half-life:
halfLife(0.10);  // 6.58 days

// User churn: 5% monthly → half your cohort gone in:
halfLife(0.05);  // 13.5 months
```

### 3. Continuous Compounding

**Discrete vs Continuous**:

```javascript
// Discrete (annual compounding)
function discreteCompounding(principal, rate, years) {
  return principal * Math.pow(1 + rate, years);
}

// Continuous (every instant)
function continuousCompounding(principal, rate, years) {
  return principal * Math.exp(rate * years);
}

// $1000 at 10% for 5 years:
discreteCompounding(1000, 0.10, 5);    // $1,610.51
continuousCompounding(1000, 0.10, 5);  // $1,648.72

// Continuous is ~2.4% higher (compounding more frequently)
```

**Real-world continuous processes**:
- API request rate (requests per second)
- Population growth (births happen continuously)
- Radioactive decay
- Database writes

```javascript
// e^(rt) form
function continuousGrowth(initial, rate, time) {
  return initial * Math.exp(rate * time);
}

// Requests: 100 req/s, growing 5% per second (viral load spike)
continuousGrowth(100, 0.05, 60);  // 2,008 req/s after 60 seconds
```

### 4. Logistic Growth (S-Curve)

**Problem**: Nothing grows exponentially forever. You hit limits.

```javascript
function logisticGrowth(initial, growthRate, capacity, time) {
  // Logistic equation: P(t) = K / (1 + ((K - P0) / P0) × e^(-rt))
  const ratio = (capacity - initial) / initial;
  return capacity / (1 + ratio * Math.exp(-growthRate * time));
}

// Startup users: 100 initial, 50% growth, market cap 10,000
function simulateStartupGrowth() {
  const times = [0, 2, 4, 6, 8, 10, 12];
  return times.map(t => ({
    month: t,
    users: Math.round(logisticGrowth(100, 0.50, 10000, t))
  }));
}

simulateStartupGrowth();
/* [
  { month: 0, users: 100 },      Slow start
  { month: 2, users: 272 },      Exponential phase
  { month: 4, users: 714 },      Hockey stick!
  { month: 6, users: 1857 },     Still fast
  { month: 8, users: 4533 },     Slowing...
  { month: 10, users: 7769 },    Approaching limit
  { month: 12, users: 9283 }     Plateau (market saturated)
]
// Classic S-curve: slow → fast → plateau */
```

**The Three Phases**:
```
Exponential Phase → Inflection Point → Saturation

      Capacity ─────────────────
                          /‾‾‾‾‾
                        /
                      /
                    /
      ___________/
      
Early adopters → Mainstream → Market saturated
```

### 5. Viral Growth & K-Factor

**Virality**: Each user brings N new users

```javascript
function viralGrowth(initialUsers, kFactor, cycles) {
  // k = average invites × conversion rate
  // k > 1 → exponential growth
  // k = 1 → linear growth
  // k < 1 → decay
  
  return Array.from({ length: cycles + 1 }, (_, cycle) => ({
    cycle,
    users: Math.round(initialUsers * Math.pow(kFactor, cycle))
  }));
}

// Example: Each user invites 5 friends, 30% sign up → k = 1.5
viralGrowth(100, 1.5, 10);
/* [
  { cycle: 0, users: 100 },
  { cycle: 1, users: 150 },
  { cycle: 2, users: 225 },
  { cycle: 3, users: 338 },
  ...
  { cycle: 10, users: 5767 }
]
// 57x growth in 10 cycles! */

// Anti-viral: k = 0.9 (each user refers <1 new user)
viralGrowth(100, 0.9, 10);
// → Decays to 35 users
```

**Critical Threshold**: k > 1 for sustained growth
```javascript
function calculateKFactor(invites, conversionRate, cycleTime) {
  return {
    k: invites * conversionRate,
    verdict: invites * conversionRate > 1 ? '🚀 Viral' : '😴 Not viral',
    doublingCycles: invites * conversionRate > 1 
      ? doublingTime(invites * conversionRate - 1) 
      : 'Never (decay)'
  };
}

calculateKFactor(10, 0.15, 7);  // 10 invites, 15% conversion, 7-day cycle
// { k: 1.5, verdict: '🚀 Viral', doublingCycles: 1.71 cycles }
// Doubles every ~12 days
```

### 6. Compounding Negative Effects (Technical Debt)

**Technical Debt as Exponential Decay**:

```javascript
function technicalDebtImpact(initialVelocity, debtAccumulationRate, sprints) {
  // Each sprint, velocity decreases by X%
  return Array.from({ length: sprints + 1 }, (_, sprint) => ({
    sprint,
    velocity: initialVelocity * Math.pow(1 - debtAccumulationRate, sprint),
    timeToFeature: sprint === 0 ? 1 : 1 / Math.pow(1 - debtAccumulationRate, sprint)
  }));
}

// Team starts at 10 story points/sprint, debt grows 5% per sprint
technicalDebtImpact(10, 0.05, 12);
/* [
  { sprint: 0, velocity: 10.0, timeToFeature: 1.0 },
  { sprint: 1, velocity: 9.5, timeToFeature: 1.05 },
  { sprint: 3, velocity: 8.6, timeToFeature: 1.16 },
  { sprint: 6, velocity: 7.4, timeToFeature: 1.35 },
  { sprint: 12, velocity: 5.4, timeToFeature: 1.85 }
]
// After 1 year: Velocity halved, features take 2x as long */
```

**Compounding Build Times**:
```javascript
function buildTimeGrowth(initialSeconds, codeGrowthRate, weeks) {
  // As codebase grows, build times grow superlinearly
  return Array.from({ length: weeks + 1 }, (_, week) => {
    const buildTime = initialSeconds * Math.pow(1 + codeGrowthRate, week);
    return {
      week,
      buildTimeMinutes: (buildTime / 60).toFixed(1),
      developerTimeWastedHours: ((buildTime * 20 * 5) / 3600).toFixed(1)  // 20 builds/day, 5 devs
    };
  });
}

// Build starts at 5 minutes, grows 3% weekly
buildTimeGrowth(300, 0.03, 52);
/* Week 0:  5.0 min build, 8.3 hrs/week wasted
   Week 26: 10.5 min build, 17.5 hrs/week wasted
   Week 52: 22.2 min build, 37.0 hrs/week wasted
   
   After 1 year: Builds 4x slower, team loses 1 engineer to waiting! */
```

### 7. Moore's Law & Technology Curves

**Moore's Law**: Computing power doubles every ~18 months

```javascript
function mooresLaw(currentPower, months) {
  const doublingPeriod = 18;
  return currentPower * Math.pow(2, months / doublingPeriod);
}

// CPU power in 2020 vs 2030 (10 years = 120 months)
const power2020 = 1;
const power2030 = mooresLaw(power2020, 120);
console.log(power2030.toFixed(1) + 'x');  // 141.4x more powerful

// Practical: iPhone CPU
// 2010: A4 chip (1x)
// 2024: A17 chip → 14 years = 168 months
mooresLaw(1, 168);  // 724x faster
```

**Cloud Cost Curves** (Declining exponentially):
```javascript
function cloudCostDecline(currentCost, years, annualDecline) {
  // AWS pricing drops ~5-10% per year
  return Array.from({ length: years + 1 }, (_, year) => ({
    year,
    cost: (currentCost * Math.pow(1 - annualDecline, year)).toFixed(2)
  }));
}

// $10k/month in 2024, 8% annual price drops
cloudCostDecline(10000, 10, 0.08);
/* Year 0:  $10,000
   Year 5:  $6,591 (34% cheaper)
   Year 10: $4,344 (56% cheaper) */
```

---

## Software Engineering Connections

### 1. **Database Growth**

```javascript
function databaseGrowthProjection(currentSizeGB, monthlyGrowthRate, months) {
  const projection = [];
  let size = currentSizeGB;
  
  for (let month = 0; month <= months; month++) {
    projection.push({
      month,
      sizeGB: Math.round(size),
      monthlyCost: (size * 0.10).toFixed(2),  // $0.10/GB
      warningFlag: size > 1000 ? '⚠️ Scaling needed' : '✅'
    });
    size *= (1 + monthlyGrowthRate);
  }
  
  return projection;
}

// Starting at 50GB, growing 15% per month
databaseGrowthProjection(50, 0.15, 24);
/* Month 0:  50 GB, $5/mo
   Month 6:  113 GB, $11/mo
   Month 12: 255 GB, $26/mo
   Month 18: 576 GB, $58/mo
   Month 24: 1,301 GB, $130/mo ⚠️
   
Linear thinking: "We grow 7GB/month, so in 2 years we'll have 218GB"
Reality: 1,301 GB (6x worse than linear projection!) */
```

### 2. **API Rate Limiting**

```javascript
function simulateTrafficSpike(normalRPS, spikeGrowthRate, seconds, capacity) {
  // Without rate limiting, exponential spike crashes system
  const timeline = [];
  
  for (let sec = 0; sec <= seconds; sec++) {
    const requests = normalRPS * Math.pow(1 + spikeGrowthRate, sec);
    const status = requests > capacity ? '❌ OVERLOAD' : '✅ OK';
    
    timeline.push({ sec, requests: Math.round(requests), status });
  }
  
  return timeline;
}

// Normal: 100 req/s, spike grows 20% per second, capacity 1000 req/s
simulateTrafficSpike(100, 0.20, 10, 1000);
/* Sec 0:  100 req/s ✅
   Sec 5:  249 req/s ✅
   Sec 10: 619 req/s ✅
   Sec 13: 1,121 req/s ❌ OVERLOAD
   
   Only 13 seconds from OK to crash! */
```

### 3. **Viral Loop Design**

```javascript
function designViralLoop(targetUsers, timeframeDays, initialUsers) {
  // Work backwards: What k-factor do we need?
  const cycles = timeframeDays / 7;  // Assume weekly cycles
  const requiredMultiplier = targetUsers / initialUsers;
  const kFactor = Math.pow(requiredMultiplier, 1 / cycles);
  
  // What does this require?
  const scenarios = [
    { invites: 5, conversion: (kFactor / 5).toFixed(3) },
    { invites: 10, conversion: (kFactor / 10).toFixed(3) },
    { invites: 20, conversion: (kFactor / 20).toFixed(3) }
  ];
  
  return {
    targetUsers,
    currentUsers: initialUsers,
    requiredK: kFactor.toFixed(2),
    cyclesNeeded: cycles,
    scenarios
  };
}

// Goal: 10,000 users in 90 days, starting with 100
designViralLoop(10000, 90, 100);
/* {
  requiredK: 1.43,
  cyclesNeeded: 12.86,
  scenarios: [
    { invites: 5, conversion: '0.286' },   // 5 invites × 28.6% = k=1.43
    { invites: 10, conversion: '0.143' },  // 10 invites × 14.3% = k=1.43
    { invites: 20, conversion: '0.072' }   // 20 invites × 7.2% = k=1.43
  ]
} */
```

### 4. **Caching & Exponential Backoff**

```javascript
function exponentialBackoff(attempt, baseDelayMs = 1000, maxDelayMs = 32000) {
  // Retry with exponentially increasing delays
  const delay = Math.min(baseDelayMs * Math.pow(2, attempt), maxDelayMs);
  return delay;
}

// Retry sequence
Array.from({ length: 8 }, (_, i) => ({
  attempt: i,
  delayMs: exponentialBackoff(i),
  delaySeconds: (exponentialBackoff(i) / 1000).toFixed(1)
}));
/* [
  { attempt: 0, delayMs: 1000, delaySeconds: '1.0' },
  { attempt: 1, delayMs: 2000, delaySeconds: '2.0' },
  { attempt: 2, delayMs: 4000, delaySeconds: '4.0' },
  { attempt: 3, delayMs: 8000, delaySeconds: '8.0' },
  { attempt: 4, delayMs: 16000, delaySeconds: '16.0' },
  { attempt: 5, delayMs: 32000, delaySeconds: '32.0' },  // Hit max
  { attempt: 6, delayMs: 32000, delaySeconds: '32.0' },
  { attempt: 7, delayMs: 32000, delaySeconds: '32.0' }
]
// Prevents overwhelming a failing service */
```

### 5. **Infrastructure Cost Explosion**

```javascript
function projectInfrastructureCosts(currentMonthlyCost, monthlyGrowth, months) {
  let total = 0;
  const breakdown = [];
  
  for (let month = 1; month <= months; month++) {
    const monthlyCost = currentMonthlyCost * Math.pow(1 + monthlyGrowth, month - 1);
    total += monthlyCost;
    
    breakdown.push({
      month,
      monthlyCost: Math.round(monthlyCost),
      cumulativeTotal: Math.round(total)
    });
  }
  
  return { breakdown, totalCost: Math.round(total) };
}

// $5k/month AWS, growing 10% monthly
const costs = projectInfrastructureCosts(5000, 0.10, 24);
console.log(`Total 2-year cost: $${costs.totalCost.toLocaleString()}`);
// $316,681 total (not $120k as linear thinking suggests!)

// Month 1: $5k
// Month 12: $14k
// Month 24: $42k 😱
```

### 6. **User Churn Modeling**

```javascript
function cohortRetention(initialUsers, monthlyChurnRate, months) {
  // Exponential decay of user cohort
  return Array.from({ length: months + 1 }, (_, month) => ({
    month,
    activeUsers: Math.round(initialUsers * Math.pow(1 - monthlyChurnRate, month)),
    percentRetained: ((Math.pow(1 - monthlyChurnRate, month)) * 100).toFixed(1) + '%'
  }));
}

// January cohort: 1000 users, 5% monthly churn
cohortRetention(1000, 0.05, 12);
/* [
  { month: 0, activeUsers: 1000, percentRetained: '100.0%' },
  { month: 1, activeUsers: 950, percentRetained: '95.0%' },
  { month: 3, activeUsers: 857, percentRetained: '85.7%' },
  { month: 6, activeUsers: 735, percentRetained: '73.5%' },
  { month: 12, activeUsers: 540, percentRetained: '54.0%' }
]
// After 1 year: Lost 46% of cohort */

// High churn (10% monthly) = catastrophic
cohortRetention(1000, 0.10, 12);  // Down to 282 users (72% lost!)
```

---

## Common Misconceptions

### ❌ "If we grow 10% per month, we'll have 120% growth in a year"

**Wrong**: That's additive (linear) thinking.
```javascript
// Linear (wrong)
const linearGrowth = 1.0 + (0.10 * 12);  // 2.2x

// Exponential (correct)
const exponentialGrowth = Math.pow(1.10, 12);  // 3.14x

console.log(`Linear: ${linearGrowth}x, Exponential: ${exponentialGrowth.toFixed(2)}x`);
// You'll grow 43% MORE than linear projection!
```

### ❌ "Exponential growth is always good"

**Not if it's costs, not revenue**:
```javascript
// Good: Revenue growing 20%/month
const revenue = exponentialGrowth(10000, 0.20, 12);  // $89,161

// Bad: Costs growing 20%/month
const costs = exponentialGrowth(8000, 0.20, 12);  // $71,329

// Net: $17,832 profit
// But if costs grow 25% while revenue grows 20%:
const fasterCosts = exponentialGrowth(8000, 0.25, 12);  // $112,449
// Net: -$23,288 loss! 5% difference = disaster
```

### ❌ "We can sustain this growth rate"

**Reality check**:
```javascript
function realityCheck(initialUsers, monthlyGrowth, worldPopulation = 8000000000) {
  let users = initialUsers;
  let month = 0;
  
  while (users < worldPopulation) {
    month++;
    users *= (1 + monthlyGrowth);
  }
  
  return { monthsToSaturate: month, finalUsers: Math.round(users) };
}

// Startup: 1000 users, 30% monthly growth
realityCheck(1000, 0.30);
// { monthsToSaturate: 41, finalUsers: 8,200,000,000 }
// In 3.5 years, you'd have 8 BILLION users? Obviously impossible.

// Exponential growth ALWAYS hits limits (logistic curve)
```

### ❌ "Small growth rates don't matter"

```javascript
function compoundingSmallRates(initial, rate, years) {
  const final = initial * Math.pow(1 + rate, years);
  return {
    rate: (rate * 100).toFixed(1) + '%',
    final: final.toFixed(0),
    multiplier: (final / initial).toFixed(1) + 'x'
  };
}

// 3% annual AWS price increases (seems small)
compoundingSmallRates(10000, 0.03, 10);
// { rate: '3.0%', final: '13439', multiplier: '1.3x' }
// 34% more expensive in 10 years!

// 5% monthly technical debt (seems small)
compoundingSmallRates(10, 0.05, 24);  // 10 hours/week waste
// { final: '32', multiplier: '3.2x' }
// 3x slower in 2 years!
```

---

## Practical Mini-Exercises

<details>
<summary><strong>Exercise 1: Viral Growth</strong> (Click to expand)</summary>

**Problem**: Your app has 500 users. Each user invites 8 friends per week, 12% sign up. Is this viral? When will you hit 100,000 users?

```javascript
function analyzeVirality() {
  const k = 8 * 0.12;  // k-factor
  const initialUsers = 500;
  const targetUsers = 100000;
  
  if (k <= 1) return 'Not viral';
  
  // How many cycles to reach target?
  const cycles = Math.log(targetUsers / initialUsers) / Math.log(k);
  const weeks = cycles;  // 1 week per cycle
  
  return {
    kFactor: k.toFixed(2),
    isViral: k > 1,
    weeksToTarget: weeks.toFixed(1),
    monthsToTarget: (weeks / 4.33).toFixed(1)
  };
}

console.log(analyzeVirality());
// { kFactor: '0.96', isViral: false, weeksToTarget: ..., monthsToTarget: ... }
// k < 1 → Not viral! You'll actually decay, not grow.
// Need k > 1: Increase invites OR conversion rate
```

</details>

<details>
<summary><strong>Exercise 2: Database Planning</strong> (Click to expand)</summary>

**Problem**: Your database is 200GB, growing 12% per month. When will you hit 1TB (need to migrate)?

```javascript
function timeToMigration(currentGB, growthRate, thresholdGB) {
  let size = currentGB;
  let months = 0;
  
  while (size < thresholdGB) {
    months++;
    size *= (1 + growthRate);
  }
  
  return {
    months,
    finalSize: Math.round(size),
    warningMonths: months - 3  // Start planning 3 months early
  };
}

console.log(timeToMigration(200, 0.12, 1000));
// { months: 14, finalSize: 1003, warningMonths: 11 }
// 14 months to hit 1TB, start migration planning in 11 months
```

</details>

<details>
<summary><strong>Exercise 3: Cost Explosion</strong> (Click to expand)</summary>

**Problem**: AWS costs are $3k/month, growing 8% monthly. What's your total spend over the next year?

```javascript
function annualCloudSpend(monthlyStart, growthRate) {
  let total = 0;
  let monthly = monthlyStart;
  
  for (let month = 0; month < 12; month++) {
    total += monthly;
    monthly *= (1 + growthRate);
  }
  
  return {
    startMonthly: monthlyStart,
    endMonthly: Math.round(monthly),
    totalAnnual: Math.round(total),
    averageMonthly: Math.round(total / 12)
  };
}

console.log(annualCloudSpend(3000, 0.08));
/* {
  startMonthly: 3000,
  endMonthly: 7550,
  totalAnnual: 50,389,
  averageMonthly: 4199
}
// Linear assumption: $3k × 12 = $36k
// Reality: $50k (40% more!) */
```

</details>

<details>
<summary><strong>Exercise 4: Technical Debt</strong> (Click to expand)</summary>

**Problem**: Technical debt slows velocity by 4% per sprint. Starting at 20 points/sprint, when do you drop below 10 points/sprint?

```javascript
function debtTipping point() {
  let velocity = 20;
  let sprint = 0;
  const decayRate = 0.04;
  
  while (velocity >= 10) {
    sprint++;
    velocity *= (1 - decayRate);
  }
  
  return {
    sprints: sprint,
    months: (sprint * 2 / 4.33).toFixed(1),  // 2-week sprints
    finalVelocity: velocity.toFixed(1)
  };
}

console.log(debtTippingPoint());
// { sprints: 18, months: '8.3', finalVelocity: '9.8' }
// In 8 months, velocity halved! Technical bankruptcy.
```

</details>

---

## Summary Cheat Sheet

```javascript
// EXPONENTIAL GROWTH FORMULA
y = a × (1 + r)^t
  a = initial value
  r = growth rate
  t = time

// DOUBLING TIME (Rule of 70)
Doubling Time ≈ 70 / (rate × 100)

// VIRAL K-FACTOR
k = (invites per user) × (conversion rate)
k > 1 → Exponential growth (viral)
k = 1 → Linear growth
k < 1 → Decay (death spiral)

// CONTINUOUS COMPOUNDING
y = a × e^(rt)

// HALF-LIFE (Decay)
t_half = ln(0.5) / ln(1 - r)
```

**Key Patterns**:
```
Linear:      Adds constant per step    (1, 2, 3, 4, 5...)
Exponential: Multiplies per step       (1, 2, 4, 8, 16...)
Logistic:    Exponential → S-curve → Plateau
```

**Warning Signs**:
- "Faster than expected" growth → Likely exponential
- Costs "suddenly" exploding → Compounding you missed
- "It was fine last month" → Hockey stick inflection point

---

## Next Steps

✅ **You've completed**: Exponential growth & decay patterns  
➡️ **Up next**: [12. Expected Value & Decision Making](12-expected-value.md) - Weighted outcomes, decisions under uncertainty

**Before moving on**, calculate:
```javascript
// Your current codebase has 100k LOC, growing 5% per month
// Each 10k LOC adds 10 seconds to build time
// How long until builds take 10 minutes?

function buildTimeExplosion() {
  // Your solution here
}
```

---

*Applied Math for Software Engineers • Phase 4 • [Previous: Financial Math](10-financial-math.md) | [Next: Expected Value](12-expected-value.md)*
