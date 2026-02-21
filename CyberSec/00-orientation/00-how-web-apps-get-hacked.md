# 🎯 Module 00: How Web Apps Actually Get Hacked

**Difficulty:** 🟢 Fundamentals  
**Time:** 30-45 minutes

---

## What You'll Learn

- What cybersecurity actually is (and isn't)
- The boring reality of most security breaches
- Why developers are the primary security boundary
- The attacker's economic model
- How to think about risk realistically

---

## Security vs Privacy vs Safety

These terms are often confused. Let's clarify:

### Security
**Controlling who can do what with your systems.**

Examples:
- Ensuring only authorized users can access admin panels
- Preventing attackers from injecting SQL commands
- Protecting API keys from leaking

### Privacy
**Controlling who knows what about users.**

Examples:
- Not logging personally identifiable information (PII)
- Encrypting data at rest
- Respecting user consent for data collection

### Safety
**Preventing harm to humans.**

Examples:
- Rate limiting to prevent service disruption
- Input validation to prevent system crashes
- Failsafes in critical systems (medical, automotive)

**Why this matters:** Most developers conflate these. "Using HTTPS" is security (preventing MITM attacks) but doesn't guarantee privacy (the server still sees everything) or safety (a bug can still crash the app).

---

## The Boring Truth About Breaches

### Myth: Hackers are Genius Masterminds

Reality: Most breaches are **embarrassingly simple**.

| Breach Type | % of Breaches | Example |
|-------------|---------------|---------|
| Stolen/leaked credentials | ~30% | Password in GitHub repo |
| Phishing | ~25% | "Click here to verify your account" |
| Misconfigured cloud storage | ~15% | Public S3 bucket with PII |
| Unpatched vulnerabilities | ~15% | Running ancient WordPress version |
| Insider threats | ~10% | Disgruntled employee |
| Zero-days and APTs | ~5% | Nation-state attacks, novel exploits |

**Key insight:** The "elite hacker" narrative is mostly fiction. Most attackers are:
- Opportunistic (scanning for easy targets)
- Economically motivated (ransomware, data theft)
- Using automated tools (not hand-crafting exploits)

---

## Why Most Breaches Are Boring

### 1. Default Configurations
```bash
# Elasticsearch exposed to the internet with no auth
curl http://exposed-server:9200/_search?pretty

# Returns entire database
```

**Why it happens:**
- Developer followed quick-start guide
- Intended to "fix it later"
- Shipped to production unchanged

---

### 2. Leaked Credentials

**Scenario:** Developer pushes code to GitHub:

```typescript
// config.ts
export const DATABASE_URL = "postgresql://admin:hunter2@prod-db.example.com:5432/users";
export const JWT_SECRET = "my-super-secret-key";
```

**What happens:**
1. Automated scanners find this within minutes
2. Attackers use credentials to access database
3. All user data exfiltrated

**Why it happens:**
- No `.gitignore` for config files
- No pre-commit hooks checking for secrets
- No secret management system

---

### 3. Unpatched Software

**Scenario:** A critical CVE is announced for a popular library.

```json
{
  "dependencies": {
    "old-library": "1.2.3"  // Vulnerable version
  }
}
```

**What happens:**
1. Attackers scan the internet for vulnerable versions
2. Exploit publicly available (often on GitHub)
3. Remote code execution achieved

**Why it happens:**
- No automated dependency updates
- "If it ain't broke, don't fix it" mentality
- Lack of visibility into dependency tree

---

### 4. Logic Flaws

**Scenario:** Insufficiently validated business logic.

```typescript
app.post('/api/transfer', async (req, res) => {
  const { fromAccount, toAccount, amount } = req.body;
  
  // Check if user owns fromAccount
  if (req.user.id !== fromAccount.userId) {
    return res.status(403).json({ error: 'Unauthorized' });
  }
  
  // Transfer money
  await transferMoney(fromAccount, toAccount, amount);
  
  return res.json({ success: true });
});
```

**What's wrong?**
- No validation that `amount` is positive
- Attacker can send `amount: -1000` to steal money from victim

**Why it happens:**
- Assumptions about input validation
- Not threat-modeling edge cases
- Trusting client-side validation

---

## Why "Just Use HTTPS" Isn't Security

HTTPS is **necessary but insufficient**.

```mermaid
graph LR
    A[Browser] -->|HTTPS| B[Server]
    B -->|Processes Request| C[Logic Flaw?]
    B -->|Queries DB| D[SQL Injection?]
    B -->|Stores Data| E[Insecure Storage?]
    B -->|Returns Response| F[XSS Vulnerability?]
    
    style A fill:#90EE90
    style B fill:#FFD700
    style C fill:#FF6B6B
    style D fill:#FF6B6B
    style E fill:#FF6B6B
    style F fill:#FF6B6B
```

### What HTTPS Does
✅ Protects data **in transit**  
✅ Prevents man-in-the-middle attacks  
✅ Ensures you're talking to the right server  

### What HTTPS Doesn't Do
❌ Validate input  
❌ Check authorization  
❌ Prevent SQL injection  
❌ Stop XSS attacks  
❌ Protect data at rest  
❌ Fix logic bugs  

**Analogy:** HTTPS is like locking your car doors. Important! But it doesn't prevent brake failure.

---

## Developers Are the Main Attack Surface

### The Uncomfortable Truth

Modern security breaches are **mostly caused by application-level vulnerabilities**, not network attacks.

```mermaid
pie
    title "Where Vulnerabilities Exist"
    "Application Code" : 60
    "Configuration" : 25
    "Infrastructure" : 10
    "Zero-Days" : 5
```

### Why Developers?

1. **Vast attack surface:**
   - Every API endpoint
   - Every form field
   - Every file upload
   - Every authentication check
   - Every authorization boundary

2. **Complexity:**
   - Modern apps have hundreds of dependencies
   - Microservices = more trust boundaries
   - Frontend + Backend = double the risk

3. **Velocity:**
   - Ship fast = less time for review
   - Technical debt = shortcuts taken
   - "Fix it later" rarely happens

---

## The Attacker's Economic Model

### Attackers Are Not Artists

They're **economically rational actors**.

#### Cost-Benefit Analysis

| Attack Type | Effort | Reward | Risk |
|-------------|--------|--------|------|
| Automated scanning | Low | Medium | Low |
| Credential stuffing | Low | High | Low |
| Phishing | Low | High | Medium |
| Exploiting known CVEs | Medium | High | Low |
| Finding zero-days | High | Very High | Medium |

**Key insight:** Attackers take the **path of least resistance**.

If your app is harder to exploit than your competitor's, attackers move on.

---

### Attacker Profiles

#### 1. Script Kiddies
- **Goal:** Lulz, reputation
- **Skills:** Low (running pre-built tools)
- **Threat:** Opportunistic, mass scanning
- **Defense:** Basic hygiene (patch, secure defaults)

#### 2. Cybercriminals
- **Goal:** Money (ransomware, theft)
- **Skills:** Medium to High
- **Threat:** Targeted, persistent
- **Defense:** Defense in depth, monitoring, incident response

#### 3. Nation-State Actors (APTs)
- **Goal:** Espionage, sabotage
- **Skills:** Very High
- **Threat:** Highly targeted, patient, well-resourced
- **Defense:** Assume breach, zero-trust, compartmentalization

**Reality check:** Unless you work in defense, finance, or critical infrastructure, you're facing #1 and #2.

---

## Security Is Not a Binary State

### Common Developer Mistakes

#### ❌ "Is my app secure?"
This question is meaningless.

#### ✅ "What's my threat model?"
This question is useful.

Security is about **reducing risk** to acceptable levels, not eliminating it.

```mermaid
graph TD
    A[Threat Modeling] --> B[Identify Assets]
    B --> C[Identify Threats]
    C --> D[Assess Risk]
    D --> E[Mitigate High-Priority Risks]
    E --> F[Monitor and Respond]
    F --> A
    
    style D fill:#FFD700
    style E fill:#90EE90
```

---

## The Security-Usability-Cost Triangle

You can't optimize all three simultaneously.

```
         Security
            △
           / \
          /   \
         /     \
        /       \
       /_________\
 Usability      Cost
```

**Examples:**

| Scenario | Security | Usability | Cost |
|----------|----------|-----------|------|
| SMS-based 2FA | Medium | High | Low |
| Hardware key (YubiKey) | High | Medium | Medium |
| Biometric auth | High | High | High |

**Tradeoffs are inevitable.** The goal is to make informed choices.

---

## What Cybersecurity Actually Is

### It's Not:
- ❌ Running a vulnerability scanner
- ❌ Following a compliance checklist
- ❌ Using fancy tools
- ❌ Hiring a security team and forgetting about it

### It Is:
- ✅ **Reasoning about trust boundaries**
- ✅ **Questioning assumptions**
- ✅ **Designing systems that fail safely**
- ✅ **Making security easy for developers to do right**
- ✅ **Detecting and responding to incidents**

---

## The Defensive Mindset

### Think Like an Attacker (But Act Like a Defender)

**Attacker thinking:**
- "What can I control?"
- "What does the system assume?"
- "What happens if I do something unexpected?"

**Example:**

```typescript
app.get('/user/:id', async (req, res) => {
  const user = await db.users.findById(req.params.id);
  return res.json(user);
});
```

**Attacker questions:**
- Can I access other users by changing `:id`?
- What if `:id` is not a number?
- What if `:id` is negative?
- Does the response leak sensitive data?

**Defender response:**
- Add authorization check (does requesting user own this data?)
- Validate `:id` is a valid UUID/integer
- Only return public fields

---

## Why This Curriculum Exists

### The Problem

Most security training is:
1. **Too academic** (security theory without practical context)
2. **Too tool-focused** (run this scanner, done!)
3. **Too offense-oriented** (become a penetration tester)

### What's Missing

Training that helps developers:
- **Understand root causes** of vulnerabilities
- **Reason about systems** holistically
- **Build security in** from the start
- **Use tools effectively** (not just run them)

### The Goal

After this curriculum, you should:
- ✅ Recognize vulnerabilities in code reviews
- ✅ Threat-model new features
- ✅ Write secure code by default
- ✅ Use security tooling effectively
- ✅ Communicate with security teams
- ✅ Not feel like security is black magic

---

## Summary

1. **Most breaches are boring** — credentials, misconfigurations, unpatched software
2. **HTTPS ≠ Security** — it's one layer of many
3. **Developers are the main attack surface** — application logic is where bugs live
4. **Attackers are economically rational** — they take the path of least resistance
5. **Security is risk management**, not perfection
6. **Defensive mindset** — question assumptions, validate everything

---

## What's Next?

Now that you understand what cybersecurity actually is, let's learn how to think about threats systematically.

→ **Next: [Module 01: Threat Modeling](../01-threat-modeling/01-thinking-like-an-attacker.md)**

---

## Further Reading

- [Krebs on Security](https://krebsonsecurity.com/) — Real-world breach analysis
- [Troy Hunt's Blog](https://www.troyhunt.com/) — Practical web security
- [OWASP Top 10](https://owasp.org/www-project-top-ten/) — Common vulnerabilities

---

## Exercises

1. **Find a recent security breach** (Google "data breach 2025")
   - What was the root cause?
   - Was it sophisticated or boring?
   - How could developers have prevented it?

2. **Review your own code:**
   - Find 3 places where you trust user input
   - What assumptions are you making?
   - What could go wrong?

3. **Threat model your login page:**
   - What assets does it protect?
   - Who would want to attack it?
   - What would they gain?
