# 🔐 Module 11: Application Security in TypeScript

**Difficulty:** 🔴 Advanced  
**Time:** 90 minutes

---

## What You'll Learn

- Input validation and sanitization
- Output encoding (preventing XSS)
- ORM security (Prisma, TypeORM)
- Regular expression DoS (ReDoS)
- JSON parsing pitfalls
- Prototype pollution
- Dependency vulnerabilities

---

## Input Validation

### ❌ Never Trust User Input

```typescript
// ❌ DANGEROUS: No validation
app.post('/api/user', async (req, res) => {
  const user = await db.user.create({
    data: req.body  // ← Attacker controls this!
  });
  res.json(user);
});
```

**Attack:**
```bash
curl -X POST /api/user \
  -d '{"username":"admin","isAdmin":true}'
```

---

### ✅ Use Validation Libraries

**Zod Example:**

```typescript
import { z } from 'zod';

const UserSchema = z.object({
  username: z.string().min(3).max(30).regex(/^[a-zA-Z0-9_]+$/),
  email: z.string().email(),
  age: z.number().int().min(13).max(120)
});

app.post('/api/user', async (req, res) => {
  try {
    const validated = UserSchema.parse(req.body);
    
    const user = await db.user.create({
      data: {
        username: validated.username,
        email: validated.email,
        age: validated.age
        // isAdmin NOT allowed from user input
      }
    });
    
    res.json(user);
  } catch (error) {
    if (error instanceof z.ZodError) {
      return res.status(400).json({ errors: error.errors });
    }
    throw error;
  }
});
```

---

### Other Validation Libraries

**joi:**
```typescript
import Joi from 'joi';

const schema = Joi.object({
  username: Joi.string().alphanum().min(3).max(30).required(),
  email: Joi.string().email().required()
});

const { error, value } = schema.validate(req.body);
```

**class-validator:**
```typescript
import { IsString, IsEmail, Min, Max, Length } from 'class-validator';

class CreateUserDto {
  @IsString()
  @Length(3, 30)
  username: string;

  @IsEmail()
  email: string;

  @Min(13)
  @Max(120)
  age: number;
}
```

---

## Output Encoding (XSS Prevention)

### ❌ DANGEROUS: Direct HTML Rendering

```typescript
// ❌ Server-side rendering without escaping
app.get('/profile/:username', async (req, res) => {
  const user = await db.user.findUnique({
    where: { username: req.params.username }
  });
  
  const html = `
    <h1>Profile: ${user.bio}</h1>
  `;
  // ↑ If bio contains <script>alert('XSS')</script>, it executes!
  
  res.send(html);
});
```

---

### ✅ Use Template Engines with Auto-Escaping

**EJS (auto-escaping):**
```typescript
// views/profile.ejs
<h1>Profile: <%= user.bio %></h1>
<!-- <%= auto-escapes HTML -->

app.get('/profile/:username', async (req, res) => {
  const user = await db.user.findUnique({
    where: { username: req.params.username }
  });
  res.render('profile', { user });
});
```

**Handlebars:**
```typescript
// {{user.bio}} is auto-escaped
<h1>Profile: {{user.bio}}</h1>

// {{{user.bio}}} is NOT escaped (dangerous!)
```

---

### ✅ Frontend: Use React/Vue (Auto-Escaping)

```typescript
// React auto-escapes text content
function Profile({ user }) {
  return (
    <div>
      <h1>{user.bio}</h1> {/* Safe */}
      <div dangerouslySetInnerHTML={{ __html: user.bio }} /> {/* Dangerous! */}
    </div>
  );
}
```

---

### Manual Escaping Function

```typescript
function escapeHtml(text: string): string {
  const map: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#039;'
  };
  return text.replace(/[&<>"']/g, (char) => map[char]);
}

const html = `<h1>Profile: ${escapeHtml(user.bio)}</h1>`;
```

---

## ORM Security

### Prisma (Generally Safe)

```typescript
// ✅ Parameterized by default
const user = await prisma.user.findMany({
  where: {
    username: req.query.username  // Safe: parameterized
  }
});
```

**But raw queries are dangerous:**

```typescript
// ❌ DANGEROUS: SQL Injection
const username = req.query.username;
const users = await prisma.$queryRawUnsafe(
  `SELECT * FROM users WHERE username = '${username}'`
);
// Attack: ?username=' OR '1'='1

// ✅ SAFE: Use parameters
const users = await prisma.$queryRaw`
  SELECT * FROM users WHERE username = ${username}
`;
```

---

### TypeORM

```typescript
// ✅ Safe: Query builder
const user = await userRepository.findOne({
  where: { username: req.query.username }
});

// ❌ DANGEROUS: Raw query
const users = await userRepository.query(
  `SELECT * FROM users WHERE username = '${req.query.username}'`
);

// ✅ SAFE: Parameters
const users = await userRepository.query(
  `SELECT * FROM users WHERE username = $1`,
  [req.query.username]
);
```

---

## Regular Expression Denial of Service (ReDoS)

### What Is ReDoS?

**Certain regex patterns have exponential time complexity.**

---

### ❌ DANGEROUS Regex

```typescript
// ❌ Catastrophic backtracking
const emailRegex = /^([a-zA-Z0-9]+)*@[a-zA-Z0-9]+\.[a-z]+$/;

// Attack:
const attack = 'a'.repeat(30) + '!';
emailRegex.test(attack);  // ← Hangs for seconds/minutes
```

**The pattern `([a-zA-Z0-9]+)*` causes exponential backtracking.**

---

### ✅ Safe Alternatives

```typescript
// ✅ Use built-in validation or safe regex
import validator from 'validator';

if (validator.isEmail(userInput)) {
  // ...
}

// ✅ Simpler regex
const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
```

---

### Detecting ReDoS

**Use tools:**

```bash
npm install -g redos-detector

redos-detector "^([a-zA-Z0-9]+)*@"
# Output: Vulnerable to ReDoS
```

---

## JSON Parsing Pitfalls

### Large Payload DoS

```typescript
// ❌ No size limit
app.use(express.json());

// Attack: Send 1 GB JSON payload → OOM crash
```

**✅ Set limits:**

```typescript
app.use(express.json({ limit: '1mb' }));
```

---

### Prototype Pollution

```typescript
// ❌ DANGEROUS: Merging user input
function merge(target: any, source: any) {
  for (const key in source) {
    target[key] = source[key];
  }
}

const userSettings = {};
merge(userSettings, req.body);

// Attack:
// POST with: {"__proto__": {"isAdmin": true}}
// Now ALL objects have isAdmin = true!
```

---

### ✅ Prevent Prototype Pollution

**Option 1: Object.create(null)**

```typescript
const userSettings = Object.create(null);
// No prototype chain
```

**Option 2: Validate keys**

```typescript
function safeMerge(target: any, source: any) {
  for (const key in source) {
    if (key === '__proto__' || key === 'constructor' || key === 'prototype') {
      continue;  // Skip dangerous keys
    }
    target[key] = source[key];
  }
}
```

**Option 3: Use safe libraries**

```typescript
import merge from 'lodash/merge';  // Patched against prototype pollution
```

---

## Dependency Vulnerabilities

### Audit Dependencies Regularly

```bash
# Check for known vulnerabilities
npm audit

# Fix automatically (if patches available)
npm audit fix

# Force fixes (may break things)
npm audit fix --force
```

---

### Use Lockfiles

```bash
# ✅ Always commit package-lock.json or yarn.lock
git add package-lock.json
git commit -m "Lock dependencies"

# ❌ Don't ignore lockfiles in .gitignore
```

**Why?**
- Ensures reproducible builds
- Prevents supply chain attacks (new malicious version)

---

### Monitor Dependencies

**Snyk:**

```bash
# Install
npm install -g snyk

# Test project
snyk test

# Monitor continuously
snyk monitor
```

**Dependabot (GitHub):**
- Automatically creates PRs for vulnerable dependencies
- Enable in repository settings

---

## Secret Management

### ❌ Hardcoded Secrets

```typescript
// ❌ NEVER do this
const API_KEY = 'sk-1234567890abcdef';
const DB_PASSWORD = 'MyPassword123';
```

---

### ✅ Environment Variables

```typescript
// .env (NOT committed to git)
API_KEY=sk-1234567890abcdef
DB_URL=postgres://user:pass@localhost:5432/db

// .gitignore
.env

// app.ts
import dotenv from 'dotenv';
dotenv.config();

const apiKey = process.env.API_KEY;
if (!apiKey) {
  throw new Error('API_KEY not set');
}
```

---

### ✅ Secret Managers (Production)

```typescript
// AWS Secrets Manager
import { SecretsManagerClient, GetSecretValueCommand } 
  from '@aws-sdk/client-secrets-manager';

const client = new SecretsManagerClient({ region: 'us-east-1' });
const response = await client.send(
  new GetSecretValueCommand({ SecretId: 'prod/api/key' })
);
const apiKey = response.SecretString;
```

---

## Type Safety ≠ Security

### TypeScript Does NOT Prevent:

```typescript
// TypeScript compiles fine, but SQL injection possible
function getUser(username: string) {
  return db.query(`SELECT * FROM users WHERE username = '${username}'`);
}

// TypeScript compiles fine, but XSS possible
function renderProfile(bio: string) {
  return `<div>${bio}</div>`;
}
```

**Types help, but you still need validation and encoding.**

---

## Summary Table

| Vulnerability | Risk | Prevention |
|---------------|------|------------|
| **Mass Assignment** | Privilege escalation | Validate input, whitelist fields |
| **XSS** | Account takeover | Auto-escaping templates, CSP |
| **SQL Injection** | Data breach | Use ORM, parameterized queries |
| **ReDoS** | DoS | Safe regex, timeouts |
| **Prototype Pollution** | RCE, privilege escalation | Validate keys, Object.create(null) |
| **Dependency Vuln** | Various | npm audit, lockfiles, Snyk |
| **Hardcoded Secrets** | Credential theft | .env files, secret managers |

---

## Exercises

### Exercise 1: Fix Mass Assignment

```typescript
// Given this vulnerable code:
app.post('/api/user', (req, res) => {
  const user = db.user.create({ data: req.body });
  res.json(user);
});
```

**Task:** Use Zod to allow only `username` and `email`.

---

### Exercise 2: Prevent ReDoS

Test this regex with a long string:

```typescript
const regex = /^(a+)+$/;
regex.test('a'.repeat(30) + '!');
```

Is it vulnerable? How long does it take?

---

### Exercise 3: Audit Your Project

```bash
cd your-project
npm audit
```

What vulnerabilities exist? Can you fix them?

---

## What's Next?

Now let's design secure APIs.

→ **Next: [Module 12: Secure API Design](../12-secure-api-design/01-api-design.md)**

---

## Further Reading

- [OWASP Input Validation Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Input_Validation_Cheat_Sheet.html)
- [Prototype Pollution Explained](https://portswigger.net/daily-swig/prototype-pollution)
- [ReDoS Attack Examples](https://owasp.org/www-community/attacks/Regular_expression_Denial_of_Service_-_ReDoS)
