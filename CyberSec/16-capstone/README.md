# 🎯 Module 16: Capstone Project

**Difficulty:** 🔴 Advanced  
**Time:** 8-12 hours (spread over multiple sessions)

---

## Project Overview

Build, attack, fix, and secure a **deliberately vulnerable web application**.

You'll wear three hats:
1. **Developer** — Build the app
2. **Attacker** — Find and exploit vulnerabilities
3. **Security Engineer** — Fix vulnerabilities and add defenses

---

## Project Specification

### Application: Task Management API

**Features:**
- User registration and authentication
- Create, read, update, delete tasks
- Share tasks with other users
- File upload (task attachments)
- Admin panel

**Tech Stack:**
- **Backend:** Node.js + Express + TypeScript
- **Database:** PostgreSQL (via Prisma)
- **Auth:** JWT
- **Storage:** Local filesystem (simulate S3)

---

## Phase 1: Build (Deliberately Vulnerable)

### Setup

```bash
mkdir security-capstone
cd security-capstone
npm init -y
npm install express prisma @prisma/client jsonwebtoken bcrypt multer cors dotenv
npm install -D typescript @types/node @types/express @types/jsonwebtoken @types/bcrypt @types/multer ts-node nodemon

npx tsc --init
npx prisma init
```

---

### Database Schema

```prisma
// prisma/schema.prisma
datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

generator client {
  provider = "prisma-client-js"
}

model User {
  id        String   @id @default(cuid())
  email     String   @unique
  username  String   @unique
  password  String
  role      String   @default("user")  // "user" or "admin"
  createdAt DateTime @default(now())
  
  tasks     Task[]
}

model Task {
  id          String   @id @default(cuid())
  title       String
  description String
  completed   Boolean  @default(false)
  authorId    String
  author      User     @relation(fields: [authorId], references: [id])
  sharedWith  String[] // Array of user IDs
  attachments String[] // Array of file paths
  createdAt   DateTime @default(now())
}
```

```bash
npx prisma migrate dev --name init
```

---

### Server Code (VULNERABLE)

**`src/server.ts`**

```typescript
import express from 'express';
import cors from 'cors';
import jwt from 'jsonwebtoken';
import bcrypt from 'bcrypt';
import { PrismaClient } from '@prisma/client';
import multer from 'multer';
import path from 'path';
import fs from 'fs';

const app = express();
const prisma = new PrismaClient();
const SECRET = 'supersecret123';  // ❌ Hardcoded secret!

app.use(cors());
app.use(express.json());

// ❌ VULNERABILITY: No rate limiting!

// ================ Authentication ================

// Register
app.post('/api/register', async (req, res) => {
  const { email, username, password } = req.body;
  
  // ❌ VULNERABILITY: No input validation!
  
  const hashedPassword = await bcrypt.hash(password, 10);
  
  const user = await prisma.user.create({
    data: {
      email,
      username,
      password: hashedPassword
    }
  });
  
  const token = jwt.sign({ userId: user.id }, SECRET);
  res.json({ token });
});

// Login
app.post('/api/login', async (req, res) => {
  const { username, password } = req.body;
  
  // ❌ VULNERABILITY: SQL injection possible if using raw query
  const user = await prisma.user.findUnique({
    where: { username }
  });
  
  if (!user || !await bcrypt.compare(password, user.password)) {
    // ❌ VULNERABILITY: No brute force protection!
    return res.status(401).json({ error: 'Invalid credentials' });
  }
  
  const token = jwt.sign({ userId: user.id }, SECRET);
  res.json({ token });
});

// Middleware
function authenticate(req: any, res: any, next: any) {
  const token = req.headers.authorization?.split(' ')[1];
  
  if (!token) {
    return res.status(401).json({ error: 'Unauthorized' });
  }
  
  try {
    const decoded = jwt.verify(token, SECRET) as any;
    req.userId = decoded.userId;
    next();
  } catch {
    res.status(401).json({ error: 'Invalid token' });
  }
}

// ================ Tasks ================

// Create task
app.post('/api/tasks', authenticate, async (req: any, res) => {
  const { title, description } = req.body;
  
  // ❌ VULNERABILITY: No XSS protection (if rendered on frontend)
  
  const task = await prisma.task.create({
    data: {
      title,
      description,
      authorId: req.userId
    }
  });
  
  res.json(task);
});

// Get all tasks
app.get('/api/tasks', authenticate, async (req: any, res) => {
  const tasks = await prisma.task.findMany({
    where: {
      OR: [
        { authorId: req.userId },
        { sharedWith: { has: req.userId } }
      ]
    }
  });
  
  res.json(tasks);
});

// Update task
app.put('/api/tasks/:id', authenticate, async (req: any, res) => {
  const { id } = req.params;
  
  // ❌ VULNERABILITY: No authorization check! (IDOR)
  // Any user can update any task
  
  const task = await prisma.task.update({
    where: { id },
    data: req.body  // ❌ VULNERABILITY: Mass assignment!
  });
  
  res.json(task);
});

// Delete task
app.delete('/api/tasks/:id', authenticate, async (req: any, res) => {
  const { id } = req.params;
  
  // ❌ VULNERABILITY: No authorization check! (IDOR)
  
  await prisma.task.delete({ where: { id } });
  res.status(204).send();
});

// ================ File Upload ================

const upload = multer({ dest: 'uploads/' });

app.post('/api/tasks/:id/upload', authenticate, upload.single('file'), async (req: any, res) => {
  const { id } = req.params;
  
  // ❌ VULNERABILITY: No file type validation!
  // ❌ VULNERABILITY: No file size limit!
  // ❌ VULNERABILITY: Path traversal possible!
  
  const task = await prisma.task.update({
    where: { id },
    data: {
      attachments: { push: req.file.path }
    }
  });
  
  res.json(task);
});

// Download file
app.get('/api/files/:filename', (req, res) => {
  const { filename } = req.params;
  
  // ❌ VULNERABILITY: Path traversal!
  // Example: /api/files/../../etc/passwd
  
  const filePath = path.join(__dirname, '../uploads', filename);
  res.sendFile(filePath);
});

// ================ Admin Panel ================

app.get('/api/admin/users', authenticate, async (req: any, res) => {
  // ❌ VULNERABILITY: No role check!
  // Any authenticated user can access admin endpoint
  
  const users = await prisma.user.findMany({
    select: {
      id: true,
      email: true,
      username: true,
      role: true,
      password: true  // ❌ VULNERABILITY: Exposing password hashes!
    }
  });
  
  res.json(users);
});

app.post('/api/admin/users/:id/role', authenticate, async (req: any, res) => {
  const { id } = req.params;
  const { role } = req.body;
  
  // ❌ VULNERABILITY: No role check!
  // Any user can make themselves admin
  
  await prisma.user.update({
    where: { id },
    data: { role }
  });
  
  res.json({ success: true });
});

// ================ Start Server ================

app.listen(3000, () => {
  console.log('Server running on http://localhost:3000');
});
```

---

## Phase 2: Attack (Find Vulnerabilities)

### Your Task: Find and Exploit

**Document each vulnerability:**
1. Vulnerability name
2. Where it exists (endpoint, line of code)
3. How to exploit it
4. Potential impact

---

### Vulnerabilities to Find

#### 🟢 Easy (10 vulnerabilities)

1. **Hardcoded secret** — Where is it? What's the risk?
2. **No rate limiting** — Test with 100 login attempts
3. **No input validation** — Register with empty username
4. **IDOR: Update task** — Can you update someone else's task?
5. **IDOR: Delete task** — Can you delete someone else's task?
6. **Mass assignment** — Can you set yourself as author of another user's task?
7. **No role check (admin users)** — Can non-admin access `/api/admin/users`?
8. **No role check (change role)** — Can you make yourself admin?
9. **Password hashes exposed** — Are hashes returned in admin endpoint?
10. **No file type validation** — Can you upload a .exe or .sh file?

---

#### 🟡 Medium (5 vulnerabilities)

11. **Path traversal (file download)** — Can you access `/api/files/../../etc/passwd`?
12. **No file size limit** — Can you upload a 1 GB file and DoS the server?
13. **No CSRF protection** — Can attacker trick user into making requests?
14. **JWT secret hardcoded** — Can attacker forge tokens?
15. **No logging** — If breached, how would you know?

---

#### 🔴 Hard (3 vulnerabilities)

16. **XSS in task description** — Store `<script>alert('XSS')</script>` in description
17. **Race condition** — Can you delete a task and update it simultaneously?
18. **JWT algorithm confusion** — Change `alg` to `none` in JWT header

---

### Attack Lab

**Test IDOR:**

```bash
# Create task as user A
curl -X POST http://localhost:3000/api/register \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","username":"alice","password":"password"}'

TOKEN_A=<token>

curl -X POST http://localhost:3000/api/tasks \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d '{"title":"Alice'\''s task","description":"Secret"}'

# Task ID: task_abc123

# Create user B
curl -X POST http://localhost:3000/api/register \
  -H "Content-Type: application/json" \
  -d '{"email":"bob@example.com","username":"bob","password":"password"}'

TOKEN_B=<token>

# Try to update Alice's task as Bob
curl -X PUT http://localhost:3000/api/tasks/task_abc123 \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -d '{"description":"Bob hacked this!"}'

# Did it work? (It should — IDOR vulnerability!)
```

---

**Test privilege escalation:**

```bash
# Get your user ID
curl http://localhost:3000/api/admin/users \
  -H "Authorization: Bearer $TOKEN_B"

# Make yourself admin
curl -X POST http://localhost:3000/api/admin/users/<your-user-id>/role \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -d '{"role":"admin"}'

# Verify
curl http://localhost:3000/api/admin/users \
  -H "Authorization: Bearer $TOKEN_B"
```

---

**Test path traversal:**

```bash
# Try to read /etc/passwd
curl http://localhost:3000/api/files/..%2F..%2Fetc%2Fpasswd

# Or
curl http://localhost:3000/api/files/../../../../etc/passwd
```

---

## Phase 3: Fix (Secure the Application)

### Your Task: Fix ALL Vulnerabilities

For each vulnerability:
1. Identify the fix
2. Implement it
3. Test that exploit no longer works

---

### Example Fixes

**1. Move secret to environment variable:**

```typescript
// .env
JWT_SECRET=<generate-random-64-char-string>

// server.ts
import dotenv from 'dotenv';
dotenv.config();

const SECRET = process.env.JWT_SECRET;
if (!SECRET) {
  throw new Error('JWT_SECRET not set');
}
```

---

**2. Add input validation:**

```typescript
import { z } from 'zod';

const RegisterSchema = z.object({
  email: z.string().email(),
  username: z.string().min(3).max(30).regex(/^[a-zA-Z0-9_]+$/),
  password: z.string().min(8)
});

app.post('/api/register', async (req, res) => {
  try {
    const validated = RegisterSchema.parse(req.body);
    // ...
  } catch (error) {
    return res.status(400).json({ error: 'Validation failed' });
  }
});
```

---

**3. Fix IDOR (authorization check):**

```typescript
app.put('/api/tasks/:id', authenticate, async (req: any, res) => {
  const { id } = req.params;
  
  // ✅ Check ownership
  const task = await prisma.task.findUnique({ where: { id } });
  
  if (!task) {
    return res.status(404).json({ error: 'Task not found' });
  }
  
  if (task.authorId !== req.userId) {
    return res.status(403).json({ error: 'Forbidden' });
  }
  
  // ✅ Whitelist fields
  const { title, description, completed } = req.body;
  
  const updated = await prisma.task.update({
    where: { id },
    data: { title, description, completed }
  });
  
  res.json(updated);
});
```

---

**4. Add role-based authorization:**

```typescript
function requireAdmin(req: any, res: any, next: any) {
  const user = await prisma.user.findUnique({
    where: { id: req.userId }
  });
  
  if (user?.role !== 'admin') {
    return res.status(403).json({ error: 'Admin access required' });
  }
  
  next();
}

app.get('/api/admin/users', authenticate, requireAdmin, async (req, res) => {
  const users = await prisma.user.findMany({
    select: {
      id: true,
      email: true,
      username: true,
      role: true
      // ✅ Don't expose password hash
    }
  });
  
  res.json(users);
});
```

---

**5. Fix path traversal:**

```typescript
import path from 'path';

app.get('/api/files/:filename', authenticate, async (req: any, res) => {
  const { filename } = req.params;
  
  // ✅ Sanitize filename
  const safeFilename = path.basename(filename);
  
  // ✅ Verify file belongs to user's tasks
  const task = await prisma.task.findFirst({
    where: {
      authorId: req.userId,
      attachments: { has: `uploads/${safeFilename}` }
    }
  });
  
  if (!task) {
    return res.status(404).json({ error: 'File not found' });
  }
  
  const filePath = path.join(__dirname, '../uploads', safeFilename);
  res.sendFile(filePath);
});
```

---

**6. Add file upload validation:**

```typescript
const upload = multer({
  dest: 'uploads/',
  limits: { fileSize: 5 * 1024 * 1024 },  // 5 MB
  fileFilter: (req, file, cb) => {
    const allowedTypes = ['image/jpeg', 'image/png', 'application/pdf'];
    
    if (!allowedTypes.includes(file.mimetype)) {
      return cb(new Error('Invalid file type'));
    }
    
    cb(null, true);
  }
});
```

---

**7. Add rate limiting:**

```typescript
import rateLimit from 'express-rate-limit';

const loginLimiter = rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 5,
  message: 'Too many login attempts'
});

app.post('/api/login', loginLimiter, async (req, res) => {
  // ...
});
```

---

**8. Add security logging:**

```typescript
import winston from 'winston';

const logger = winston.createLogger({
  format: winston.format.json(),
  transports: [
    new winston.transports.File({ filename: 'security.log' })
  ]
});

// Log failed login
logger.warn('LOGIN_FAILED', {
  username: req.body.username,
  ipAddress: req.ip
});

// Log privilege escalation
logger.alert('ROLE_CHANGED', {
  targetUser: id,
  newRole: role,
  performedBy: req.userId
});
```

---

## Phase 4: Verify

### Security Checklist

- [ ] Secrets in environment variables (not hardcoded)
- [ ] Input validation on all endpoints
- [ ] Authorization checks (IDOR fixed)
- [ ] Role-based access control (admin endpoints protected)
- [ ] Rate limiting (login, registration)
- [ ] File upload validation (type, size)
- [ ] Path traversal fixed
- [ ] XSS protection (output encoding)
- [ ] CSRF protection (tokens or SameSite cookies)
- [ ] Security logging (auth events, access denials)
- [ ] Password hashes not exposed
- [ ] Mass assignment prevented (field whitelisting)
- [ ] Error messages don't leak info
- [ ] Dependencies audited (`npm audit`)

---

### Re-Test All Exploits

**Document:**
- Vulnerability
- Original exploit
- Fix implemented
- Verification that exploit no longer works

---

## Bonus Challenges

### 🌟 Challenge 1: Add MFA

Implement TOTP-based multi-factor authentication.

---

### 🌟 Challenge 2: Add CSRF Protection

Use `csurf` middleware or SameSite cookies.

---

### 🌟 Challenge 3: Add Monitoring

Set up alerts for:
- 10+ failed logins in 5 minutes
- Admin role changes
- File uploads from unusual IP

---

### 🌟 Challenge 4: Penetration Test

Use Burp Suite to test your "fixed" app. Can you find any remaining vulnerabilities?

---

## Deliverables

1. **Vulnerability Report** (markdown)
   - List all 18 vulnerabilities found
   - Exploitation steps
   - Impact assessment

2. **Fixed Codebase**
   - All vulnerabilities patched
   - Security features added

3. **Security Documentation**
   - How to run the app securely
   - Environment variables needed
   - Security best practices for deployment

4. **Reflection** (500 words)
   - Hardest vulnerability to fix
   - Most surprising vulnerability
   - What you learned

---

## What's Next?

**Congratulations!** 🎉

You've completed the Developer Cybersecurity Curriculum.

### Continue Learning:
- Bug bounty programs (HackerOne, Bugcrowd)
- Capture The Flag (CTF) competitions
- OWASP projects
- Security certifications (CEH, OSCP)

### Keep Building:
- Apply security principles to every project
- Stay updated (security newsletters, CVE databases)
- Share knowledge (blog, talks, mentorship)

---

## Further Reading

- [OWASP Web Security Testing Guide](https://owasp.org/www-project-web-security-testing-guide/)
- [PortSwigger Web Security Academy](https://portswigger.net/web-security)
- [HackTheBox](https://www.hackthebox.com/)
- [OWASP Juice Shop](https://owasp.org/www-project-juice-shop/)
