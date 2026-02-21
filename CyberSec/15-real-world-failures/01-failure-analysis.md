# 💥 Module 15: Real-World Security Failures

**Difficulty:** 🔴 Advanced  
**Time:** 90 minutes

---

## What You'll Learn

- Real vulnerability scenarios
- Root cause analysis
- Defense-in-depth failures
- Post-incident fixes
- Lessons learned

---

## Scenario 1: GitHub OAuth Token Leak

### The Incident

**Company:** Tech startup  
**Date:** 2023  
**Impact:** 50,000 user records exposed

---

### What Happened

**Developer committed `.env` file to public GitHub repo:**

```bash
# .env
DATABASE_URL=postgres://admin:MySecurePassword123@db.example.com:5432/prod
GITHUB_CLIENT_SECRET=abc123def456ghi789
STRIPE_SECRET_KEY=sk_live_1234567890abcdef
```

**Timeline:**
- **Day 1, 2:34 PM:** Developer commits `.env` file
- **Day 1, 2:35 PM:** Automated bot scrapes GitHub for secrets
- **Day 1, 2:40 PM:** Attacker uses Stripe key to view customer data
- **Day 1, 3:15 PM:** Attacker downloads database using leaked credentials
- **Day 2, 9:00 AM:** Company discovers breach via unusual API usage alert

---

### Root Causes

1. **No .gitignore:** `.env` file not excluded
2. **No pre-commit hooks:** No secret scanning before push
3. **No secret rotation:** Keys exposed for 18 hours
4. **No monitoring:** No alerts on secret exposure
5. **No principle of least privilege:** Database credential had full access

---

### How Attacker Found It

```bash
# GitHub dorks (automated search)
filename:.env "STRIPE_SECRET_KEY"
filename:.env "DATABASE_URL" password
extension:json "api_key"
```

Automated bots constantly scan GitHub for these patterns.

---

### The Fix

**Immediate:**
```bash
# Rotate ALL secrets
- Revoke Stripe API key
- Reset database password
- Rotate GitHub OAuth credentials
- Force password reset for affected users
```

**Long-term:**
```bash
# 1. Add .gitignore
echo ".env" >> .gitignore
echo ".env.*" >> .gitignore

# 2. Install git-secrets (pre-commit hook)
brew install git-secrets
git secrets --install
git secrets --register-aws

# 3. Scan entire history
git secrets --scan-history

# 4. Enable secret scanning (GitHub)
# Repository Settings → Security → Secret scanning → Enable

# 5. Use secret manager (AWS Secrets Manager, HashiCorp Vault)
```

---

### Cost of Breach

- **Incident response:** 2 engineers × 3 days = 6 days
- **Customer notification:** Email to 50,000 users
- **Regulatory fine:** $25,000 (GDPR)
- **Reputation damage:** Lost customers
- **Total estimated cost:** $200,000+

---

## Scenario 2: Admin Panel IDOR

### The Incident

**Company:** SaaS platform  
**Date:** 2022  
**Impact:** Any user could become admin

---

### The Vulnerability

**Admin panel endpoint:**

```typescript
// ❌ VULNERABLE CODE
app.post('/api/users/:userId/role', authenticate, async (req, res) => {
  const { userId } = req.params;
  const { role } = req.body;  // "admin", "user", "moderator"
  
  // ❌ No authorization check!
  await db.user.update({
    where: { id: userId },
    data: { role }
  });
  
  res.json({ success: true });
});
```

---

### The Attack

**Normal user discovers endpoint in browser DevTools:**

```bash
# Attacker's request
POST /api/users/attacker-user-id/role
Content-Type: application/json
Authorization: Bearer <attacker-token>

{"role": "admin"}
```

**Result:** Attacker is now admin.

---

### Root Causes

1. **No authorization check:** Only authentication, no authorization
2. **Frontend security:** Button hidden in UI, but endpoint accessible
3. **No audit logging:** Role change not logged
4. **No anomaly detection:** Admin privilege escalation went unnoticed

---

### The Fix

**Immediate:**
```typescript
// ✅ FIXED CODE
app.post('/api/users/:userId/role', authenticate, async (req, res) => {
  const { userId } = req.params;
  const { role } = req.body;
  
  // ✅ Check if requester is admin
  if (req.user.role !== 'admin') {
    logger.warn('UNAUTHORIZED_ROLE_CHANGE_ATTEMPT', {
      requesterId: req.user.id,
      targetUserId: userId,
      attemptedRole: role
    });
    return res.status(403).json({ error: 'Forbidden' });
  }
  
  // ✅ Additional check: Can't change your own role
  if (req.user.id === userId) {
    return res.status(400).json({ error: 'Cannot change your own role' });
  }
  
  // ✅ Validate role
  const validRoles = ['user', 'moderator', 'admin'];
  if (!validRoles.includes(role)) {
    return res.status(400).json({ error: 'Invalid role' });
  }
  
  // ✅ Log the change
  logger.alert('ROLE_CHANGED', {
    performedBy: req.user.id,
    targetUser: userId,
    oldRole: user.role,
    newRole: role,
    timestamp: new Date()
  });
  
  await db.user.update({
    where: { id: userId },
    data: { role }
  });
  
  res.json({ success: true });
});
```

**Long-term:**
```typescript
// Centralized authorization middleware
function requireAdmin(req: Request, res: Response, next: NextFunction) {
  if (req.user?.role !== 'admin') {
    logger.warn('ADMIN_ENDPOINT_ACCESS_DENIED', { userId: req.user?.id });
    return res.status(403).json({ error: 'Admin access required' });
  }
  next();
}

app.post('/api/users/:userId/role', authenticate, requireAdmin, async (req, res) => {
  // Now we know user is admin
});
```

---

## Scenario 3: SSRF via Image Upload

### The Incident

**Company:** Social media platform  
**Date:** 2023  
**Impact:** Internal network exposed

---

### The Vulnerability

**Avatar upload endpoint:**

```typescript
// ❌ VULNERABLE CODE
app.post('/api/profile/avatar', authenticate, async (req, res) => {
  const { imageUrl } = req.body;
  
  // Fetch image from URL and save
  const response = await fetch(imageUrl);  // ❌ No validation!
  const buffer = await response.buffer();
  
  // Save to S3
  await s3.upload({
    Bucket: 'avatars',
    Key: `${req.user.id}.jpg`,
    Body: buffer
  });
  
  res.json({ success: true });
});
```

---

### The Attack

**Attacker probes internal network:**

```bash
POST /api/profile/avatar
Content-Type: application/json
Authorization: Bearer <token>

{
  "imageUrl": "http://169.254.169.254/latest/meta-data/iam/security-credentials/"
}
```

**Result:** AWS metadata service credentials leaked!

**Other attacks:**
```bash
# Read internal files
http://localhost/etc/passwd

# Scan internal network
http://10.0.0.1:22   (SSH open?)
http://10.0.0.1:3306 (MySQL open?)
http://10.0.0.2:6379 (Redis open?)

# Access internal services
http://internal-admin.company.local
```

---

### Root Causes

1. **No URL validation:** Any URL accepted
2. **SSRF vulnerability:** Server fetches arbitrary URLs
3. **No network segmentation:** App server can access internal services
4. **IMDSv1 enabled:** AWS metadata accessible without token

---

### The Fix

**Immediate:**

```typescript
// ✅ FIXED CODE
import { URL } from 'url';

const BLOCKED_HOSTS = [
  'localhost',
  '127.0.0.1',
  '169.254.169.254',  // AWS metadata
  '10.0.0.0/8',       // Private networks
  '172.16.0.0/12',
  '192.168.0.0/16'
];

function isUrlSafe(urlString: string): boolean {
  try {
    const url = new URL(urlString);
    
    // Only allow HTTP/HTTPS
    if (!['http:', 'https:'].includes(url.protocol)) {
      return false;
    }
    
    // Block private IPs and domains
    for (const blocked of BLOCKED_HOSTS) {
      if (url.hostname.includes(blocked)) {
        return false;
      }
    }
    
    return true;
  } catch {
    return false;
  }
}

app.post('/api/profile/avatar', authenticate, async (req, res) => {
  const { imageUrl } = req.body;
  
  // ✅ Validate URL
  if (!isUrlSafe(imageUrl)) {
    logger.warn('SSRF_ATTEMPT', { userId: req.user.id, url: imageUrl });
    return res.status(400).json({ error: 'Invalid image URL' });
  }
  
  // ✅ Set timeout and size limit
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 5000);
  
  try {
    const response = await fetch(imageUrl, {
      signal: controller.signal,
      headers: { 'User-Agent': 'MyApp/1.0' }
    });
    
    clearTimeout(timeout);
    
    // ✅ Validate content type
    const contentType = response.headers.get('content-type');
    if (!contentType?.startsWith('image/')) {
      return res.status(400).json({ error: 'URL must point to an image' });
    }
    
    // ✅ Limit size
    const MAX_SIZE = 5 * 1024 * 1024; // 5 MB
    const size = parseInt(response.headers.get('content-length') || '0');
    if (size > MAX_SIZE) {
      return res.status(400).json({ error: 'Image too large' });
    }
    
    const buffer = await response.buffer();
    
    // Save to S3
    await s3.upload({
      Bucket: 'avatars',
      Key: `${req.user.id}.jpg`,
      Body: buffer
    });
    
    res.json({ success: true });
  } catch (error) {
    logger.error('IMAGE_UPLOAD_FAILED', { error, userId: req.user.id });
    res.status(500).json({ error: 'Failed to upload image' });
  }
});
```

**Long-term:**
```bash
# Enable AWS IMDSv2 (requires token)
aws ec2 modify-instance-metadata-options \
  --instance-id i-1234567890abcdef0 \
  --http-tokens required

# Network segmentation
- App servers in public subnet
- Databases in private subnet
- No direct access between app and internal services
```

---

## Scenario 4: Race Condition in Wallet

### The Incident

**Company:** Fintech app  
**Date:** 2022  
**Impact:** $100,000+ stolen

---

### The Vulnerability

**Withdrawal endpoint:**

```typescript
// ❌ VULNERABLE CODE
app.post('/api/wallet/withdraw', authenticate, async (req, res) => {
  const { amount } = req.body;
  
  // 1. Check balance
  const wallet = await db.wallet.findUnique({
    where: { userId: req.user.id }
  });
  
  if (wallet.balance < amount) {
    return res.status(400).json({ error: 'Insufficient funds' });
  }
  
  // ❌ RACE CONDITION: Another request can execute here!
  
  // 2. Deduct balance
  await db.wallet.update({
    where: { userId: req.user.id },
    data: { balance: wallet.balance - amount }
  });
  
  // 3. Process withdrawal
  await processWithdrawal(req.user.id, amount);
  
  res.json({ success: true });
});
```

---

### The Attack

**Attacker sends multiple simultaneous requests:**

```bash
# Balance: $1000
# Request 1: Withdraw $1000 (t=0ms)
# Request 2: Withdraw $1000 (t=5ms)
# Request 3: Withdraw $1000 (t=10ms)

# All three pass the balance check before any deduction!
# Result: $3000 withdrawn from $1000 balance
```

---

### Root Causes

1. **No transaction locking:** Race condition possible
2. **Check-then-use:** Balance checked, then used (not atomic)
3. **No idempotency:** Same withdrawal can be processed multiple times

---

### The Fix

**Option 1: Database transaction with locking:**

```typescript
// ✅ FIXED CODE
app.post('/api/wallet/withdraw', authenticate, async (req, res) => {
  const { amount } = req.body;
  
  try {
    await db.$transaction(async (tx) => {
      // ✅ Lock the row (SELECT ... FOR UPDATE)
      const wallet = await tx.wallet.findUnique({
        where: { userId: req.user.id },
        // In raw SQL: SELECT * FROM wallets WHERE user_id = ? FOR UPDATE
      });
      
      if (wallet.balance < amount) {
        throw new Error('Insufficient funds');
      }
      
      // ✅ Atomic update
      await tx.wallet.update({
        where: { userId: req.user.id },
        data: { balance: wallet.balance - amount }
      });
      
      await processWithdrawal(req.user.id, amount);
    });
    
    res.json({ success: true });
  } catch (error) {
    res.status(400).json({ error: error.message });
  }
});
```

**Option 2: Atomic database operation:**

```typescript
// ✅ FIXED CODE (even better)
app.post('/api/wallet/withdraw', authenticate, async (req, res) => {
  const { amount } = req.body;
  
  try {
    // ✅ Atomic decrement
    const wallet = await db.wallet.update({
      where: {
        userId: req.user.id,
        balance: { gte: amount }  // ✅ Ensure sufficient funds
      },
      data: {
        balance: { decrement: amount }
      }
    });
    
    await processWithdrawal(req.user.id, amount);
    res.json({ success: true });
  } catch (error) {
    // If balance < amount, update fails
    res.status(400).json({ error: 'Insufficient funds' });
  }
});
```

---

## Scenario 5: Business Logic Flaw

### The Incident

**Company:** E-commerce platform  
**Date:** 2023  
**Impact:** $50,000 in free products

---

### The Vulnerability

**Discount code application:**

```typescript
// ❌ VULNERABLE CODE
app.post('/api/cart/apply-discount', authenticate, async (req, res) => {
  const { code } = req.body;
  
  const discount = await db.discountCode.findUnique({
    where: { code }
  });
  
  if (!discount || discount.expiresAt < new Date()) {
    return res.status(400).json({ error: 'Invalid discount code' });
  }
  
  // ❌ No check if discount already applied!
  await db.cart.update({
    where: { userId: req.user.id },
    data: {
      discountCodes: {
        push: code  // ❌ Can add same code multiple times!
      }
    }
  });
  
  res.json({ success: true });
});
```

---

### The Attack

**Attacker applies same 50% discount code 10 times:**

```bash
POST /api/cart/apply-discount
{"code": "SAVE50"}

POST /api/cart/apply-discount
{"code": "SAVE50"}

# ... 8 more times

# Cart total: $100
# After 10× 50% discount: $100 × 0.5^10 = $0.10
```

---

### Root Causes

1. **No duplicate check:** Same discount can be applied multiple times
2. **No business rule validation:** Total discount can exceed 100%
3. **No limit on discount codes per cart:** Unlimited stacking
4. **No testing with edge cases:** QA didn't test duplicate codes

---

### The Fix

```typescript
// ✅ FIXED CODE
app.post('/api/cart/apply-discount', authenticate, async (req, res) => {
  const { code } = req.body;
  
  const discount = await db.discountCode.findUnique({
    where: { code }
  });
  
  if (!discount || discount.expiresAt < new Date()) {
    return res.status(400).json({ error: 'Invalid discount code' });
  }
  
  const cart = await db.cart.findUnique({
    where: { userId: req.user.id }
  });
  
  // ✅ Check if already applied
  if (cart.discountCodes.includes(code)) {
    return res.status(400).json({ error: 'Discount code already applied' });
  }
  
  // ✅ Check maximum discounts
  if (cart.discountCodes.length >= 3) {
    return res.status(400).json({ error: 'Maximum 3 discount codes allowed' });
  }
  
  // ✅ Calculate total discount
  const totalDiscount = cart.discountCodes.reduce((sum, c) => {
    const d = db.discountCode.findUnique({ where: { code: c } });
    return sum + (d?.percentage || 0);
  }, 0) + discount.percentage;
  
  // ✅ Ensure total discount doesn't exceed 100%
  if (totalDiscount > 100) {
    return res.status(400).json({ 
      error: 'Total discount cannot exceed 100%' 
    });
  }
  
  await db.cart.update({
    where: { userId: req.user.id },
    data: {
      discountCodes: {
        push: code
      }
    }
  });
  
  res.json({ success: true });
});
```

---

## Summary: Common Patterns

| Vulnerability Type | Root Cause | Prevention |
|--------------------|------------|------------|
| **Secret leakage** | No .gitignore, no scanning | git-secrets, secret rotation |
| **IDOR** | No authorization | Resource-based authz checks |
| **SSRF** | No URL validation | Whitelist domains, block private IPs |
| **Race condition** | Non-atomic operations | Database locking, atomic updates |
| **Business logic** | Missing validation | Test edge cases, enforce limits |

---

## Lessons Learned

1. **Defense in depth:** Multiple layers of security
2. **Assume breach:** Monitor, log, alert
3. **Test negative cases:** Not just happy path
4. **Principle of least privilege:** Minimal permissions
5. **Automate security:** Pre-commit hooks, CI scanning
6. **Security is everyone's job:** Not just security team

---

## Exercises

### Exercise 1: Find the IDOR

Review yourcode:
- Are there endpoints that modify resources?
- Do they check if user owns the resource?
- Can user modify other users' data?

### Exercise 2: Check for Secret Leakage

```bash
cd your-project
git log -p | grep -i "password\|secret\|key"
```

Did you accidentally commit secrets?

### Exercise 3: Test Race Condition

If you have withdrawal/payment logic:
- Send 10 simultaneous requests
- Does balance go negative?

---

## What's Next?

Now let's apply everything in a capstone project.

→ **Next: [Module 16: Capstone Project](../16-capstone/README.md)**

---

## Further Reading

- [HackerOne Disclosed Reports](https://hackerone.com/hacktivity)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Real-World Bug Bounty Reports](https://pentester.land/)
