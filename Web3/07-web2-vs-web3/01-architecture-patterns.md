# Architecture Patterns: Web2 vs Web3

🟢 **Fundamentals**

---

## Traditional Web2 Architecture

You already know this:

```
┌─────────┐      HTTPS       ┌─────────┐      SQL      ┌──────────┐
│ Browser │ ────────────────▶│  API    │ ────────────▶ │ Database │
│         │◀──────────────── │ Server  │◀──────────────│          │
└─────────┘      JSON        └─────────┘    Rows       └──────────┘
```

**Characteristics:**
- **Fast:** Millisecond response times
- **Cheap:** Pennies per million requests
- **Flexible:** Update code anytime
- **Private:** You control who sees what
- **Centralized:** You control everything

---

## Naive Web3 Architecture (Don't Do This)

```
┌─────────┐      JSON-RPC     ┌──────────┐
│ Browser │ ────────────────▶ │ Ethereum │
│         │◀──────────────────│   Node   │
└─────────┘      Slow/Expensive└──────────┘
```

**Problems:**
- ❌ **Slow:** 12-second block time minimum
- ❌ **Expensive:** $1-100 per write
- ❌ **Can't query:** No SQL, no joins, no full-text search
- ❌ **Poor UX:** Users pay for every action

**This is why early dApps sucked.**

---

## Real Web3 Architecture (Hybrid)

```
┌─────────┐   Read (Free)    ┌─────────────┐
│ Browser │─────────────────▶│  Your API   │
│         │                   │  + Database │
│         │   Write (Paid)    └─────────────┘
│         │─────────────────▶        │
│         │                          │ Index events
│         │                          ▼
│         │                   ┌──────────┐
│         │◀──────────────────│ Ethereum │
└─────────┘    Via Wallet     │   Node   │
                               └──────────┘
```

**Key insight:** Blockchain for **writes** (security), traditional DB for **reads** (performance).

---

## The Hybrid Pattern (Most Common)

### Components

#### 1. Smart Contract (Source of Truth)
```solidity
contract TokenRegistry {
  mapping(uint => string) public tokens;
  
  event TokenRegistered(uint indexed id, string name);
  
  function register(string memory name) public {
    uint id = nextId++;
    tokens[id] = name;
    emit TokenRegistered(id, name);
  }
}
```

**Role:** Canonical data storage, business logic.

---

#### 2. Indexer (Read Layer)
```typescript
// Listen to blockchain events
contract.on('TokenRegistered', async (id, name) => {
  // Store in PostgreSQL for fast queries
  await db.query(
    'INSERT INTO tokens (id, name, registered_at) VALUES ($1, $2, $3)',
    [id, name, new Date()]
  );
});
```

**Role:** Make blockchain data queryable.

---

#### 3. API Server (User-Facing)
```typescript
// Fast queries against database
app.get('/tokens', async (req, res) => {
  const tokens = await db.query(
    'SELECT * FROM tokens WHERE name LIKE $1 ORDER BY registered_at DESC',
    [`%${req.query.search}%`]
  );
  res.json(tokens);
});
```

**Role:** Traditional REST/GraphQL API.

---

#### 4. Frontend (Browser)
```typescript
// Read from your API (fast, free)
const tokens = await fetch('/api/tokens').then(r => r.json());

// Write to blockchain (slow, costs gas)
const tx = await contract.register('MyToken');
await tx.wait(); // Wait for confirmation
```

**Role:** User interface, wallet integration.

---

## Decision Matrix: What Goes Where?

### Store On-Chain ⛓️

✅ **Value transfers** (send ETH, tokens)  
✅ **Ownership records** (who owns what)  
✅ **Critical business logic** (can't be tampered)  
✅ **Censorship-resistant data** (must be permanent)

**Examples:**
- Token balances
- NFT ownership
- DAO votes
- Financial transactions

---

### Store Off-Chain 🗄️

✅ **Metadata** (descriptions, images)  
✅ **User profiles** (names, avatars)  
✅ **Computed/aggregated data** (leaderboards, statistics)  
✅ **Temporary data** (pending transactions)  
✅ **Private data** (anything confidential)

**Storage options:**
- **Your database:** Full control, centralized
- **IPFS:** Decentralized, content-addressed
- **Arweave:** Permanent storage (pay once)

---

## Common Patterns

### Pattern 1: Blockchain as Audit Log

Traditional database for operations, blockchain for immutable history.

```
User Action
    ↓
  Your DB (fast operations)
    ↓
Blockchain (immutable record)
```

**Example:** Financial transactions
- Process in your system (fast)
- Hash and commit to blockchain (provable)

---

### Pattern 2: Blockchain as Source of Truth

Blockchain has canonical data, database is read replica.

```
Write → Blockchain (source of truth)
         ↓
      Events
         ↓
      Indexer
         ↓
    Database (read replica)
         ↓
      API (fast reads)
```

**Example:** NFT marketplace
- Ownership on blockchain
- Search/filter via database

---

### Pattern 3: Optimistic UI

Show immediate feedback, confirm on-chain later.

```typescript
// User clicks "Buy NFT"
setLoading(true);

// Optimistically show success
showNotification('Purchase initiated');
updateUIAsOwned(nftId);

// Send transaction
const tx = await contract.buyNFT(nftId);

try {
  await tx.wait(); // Wait for confirmation
  showNotification('Purchase confirmed!');
} catch (error) {
  // Revert optimistic update
  revertUIUpdate(nftId);
  showError('Transaction failed');
}
```

**Tradeoff:** Better UX, but must handle reversals.

---

## Authentication Patterns

### Web2: Session-Based
```
1. User logs in with password
2. Server creates session
3. Cookie sent with each request
```

### Web3: Signature-Based
```
1. User connects wallet
2. Signs challenge message (proves ownership of address)
3. Server verifies signature, creates JWT
4. JWT sent with requests
```

```typescript
// Backend: Generate challenge
app.get('/auth/challenge', (req, res) => {
  const challenge = `Sign this message to log in: ${randomBytes(32)}`;
  // Store challenge temporarily
  res.json({ challenge });
});

// Backend: Verify signature
app.post('/auth/verify', async (req, res) => {
  const { address, signature, challenge } = req.body;
  
  const recovered = ethers.verifyMessage(challenge, signature);
  
  if (recovered.toLowerCase() === address.toLowerCase()) {
    const jwt = generateJWT(address);
    res.json({ token: jwt });
  } else {
    res.status(401).json({ error: 'Invalid signature' });
  }
});
```

---

## Data Flow Example: NFT Marketplace

### Writing (Selling NFT)

```
1. User clicks "List NFT"
   ↓
2. Frontend calls contract.listNFT(tokenId, price)
   ↓
3. Wallet pops up for approval
   ↓
4. Transaction sent to blockchain
   ↓
5. Miners/validators include in block
   ↓
6. Block confirmed (12 seconds on Ethereum)
   ↓
7. Contract emits NFTListed event
   ↓
8. Your indexer detects event
   ↓
9. Updates database: INSERT INTO listings ...
```

**User sees: 12+ second delay**

---

### Reading (Browsing NFTs)

```
1. User visits marketplace
   ↓
2. Frontend fetches /api/listings
   ↓
3. Your API queries PostgreSQL
   ↓
4. Returns JSON (100ms)
   ↓
5. Frontend renders
```

**User sees: Instant**

---

## Cost Comparison

### All On-Chain (Naive)
```
Write: $10/transaction
Read:  $0 (anyone can read blockchain)
Total: $10/user action
```

### Hybrid (Smart)
```
Write: $10/transaction (rare actions like buying)
Read:  $0.0001/query (your server costs)
Total: ~$0.01/user session
```

**100x-1000x cheaper.**

---

## Handling Blockchain Delays

### Problem
Blockchain confirms slowly. Users expect instant feedback.

### Solutions

**1. Optimistic UI** (show success immediately, revert if fails)

**2. Pending states** (show "confirming..." UI)

**3. Webhooks/websockets** (notify when confirmed)

```typescript
// Listen for transaction confirmation
provider.once(tx.hash, (receipt) => {
  if (receipt.status === 1) {
    notifyUser('Transaction confirmed!');
  } else {
    notifyUser('Transaction failed');
  }
});
```

---

## Summary

**Pure Web2:**
- Fast, cheap, flexible
- Centralized, can be censored

**Pure Web3:**
- Slow, expensive, immutable
- Decentralized, censorship-resistant

**Hybrid (Best Practice):**
- Blockchain for writes (security, ownership)
- Traditional DB for reads (speed, queries)
- Best of both worlds

**Architecture decision framework:**
1. What data needs censorship resistance? → On-chain
2. What needs fast queries? → Off-chain (indexed)
3. What needs privacy? → Off-chain or encrypted
4. What changes frequently? → Off-chain

**Key insight:** You don't build "on blockchain." You build systems that use blockchain strategically for specific properties (immutability, decentralization).

---

**[Next Lesson →](02-reading-data.md)**
