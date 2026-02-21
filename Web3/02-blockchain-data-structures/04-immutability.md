# Immutability

🟢 **Fundamentals**

---

## What Immutability Really Means

**Immutability:** Data, once written, cannot be modified or deleted.

In blockchains:
- Transactions are permanent
- Smart contracts (usually) can't be changed
- History can't be rewritten (without massive cost)

**Key insight:** Immutability is a spectrum, not absolute.

---

## How Blockchains Achieve Immutability

### 1. Hash Chains
Changing old data breaks the hash chain.

```
Block N (modified) → hash changes
Block N+1 → parentHash no longer matches → chain breaks
```

### 2. Proof of Work (PoW)
Recomputing blocks requires redoing expensive computations.

**Attack cost:** Must redo all work from the modified block to present.

### 3. Replication
Thousands of nodes store the blockchain. Consensus determines the valid chain.

**Attack requirement:** Control >50% of nodes/hash power.

---

## Types of Immutability

### 1. Blockchain Data (Transactions)
**Always immutable:** Can't delete or modify past transactions.

### 2. Smart Contract Code
**Depends on design:**
- **Non-upgradeable:** Code is permanent (default)
- **Upgradeable:** Admin can change implementation (via proxy pattern)

### 3. Off-Chain Data (IPFS, etc.)
**Depends on pinning:** If no one hosts the data, it "disappears."

---

## Benefits of Immutability

### 1. Auditability
Full history is preserved and verifiable.

**Example:** Tax audits, provenance tracking.

### 2. Trust in Execution
Code executes as written. No one can change rules retroactively.

**Example:** DeFi protocols with transparent rules.

### 3. Censorship Resistance
No one can delete transactions or rewrite history (easily).

**Example:** Wikileaks donations, activist funding.

---

## Costs of Immutability

### 1. Bugs Are Permanent
If a smart contract has a bug, it can't be patched.

**Example: Parity Wallet Bug (2017)**
- Bug froze $150M in ETH
- Code was immutable → funds locked forever

### 2. Bad Data Stays Forever
Illegal content, mistakes, spam—all permanent.

**Example:** Someone posted illegal content to Ethereum. It's still there.

### 3. No "Undo" Button
Sent funds to wrong address? Too bad.

**Example:** $300M sent to wrong address (various incidents).

### 4. Storage Bloat
Blockchain size grows forever. Ethereum is >1 TB.

**Result:** Fewer people can run full nodes (centralization pressure).

---

## Immutability vs GDPR (Right to Be Forgotten)

**GDPR:** EU law requiring companies to delete user data on request.

**Blockchains:** Data can't be deleted.

**Conflict:**
- If personal data is on-chain, it can't be removed
- Violates GDPR's "right to be forgotten"

**Solutions:**
- Don't store personal data on-chain
- Store only hashes (data lives off-chain)
- Use private/permissioned blockchains

---

## When Immutability Is Harmful

### 1. Medical Records
What if a doctor makes a data entry error?
- In a database: Correction is trivial
- On a blockchain: Error is permanent

### 2. Software Development
Code needs to evolve, bugs need fixing.
- Traditional apps: Deploy updates easily
- Smart contracts: Immutability makes updates complex or impossible

### 3. Content Moderation
Illegal/harmful content on-chain can't be removed.
- Web2: Platforms can delete content
- Web3: Content is permanent

---

## Breaking Immutability: Hard Forks

A **hard fork** changes protocol rules, creating a new chain.

**Example: Ethereum vs Ethereum Classic (2016)**

**Background:**
- The DAO smart contract was hacked ($60M stolen)
- Community debated: Should we reverse the hack?

**Outcome:**
- **Ethereum:** Hard forked to reverse the hack (immutability violated)
- **Ethereum Classic:** Kept original chain ("code is law")

**Result:** Two separate blockchains with different philosophies.

**Lesson:** "Immutability" isn't absolute. Social consensus can override it.

---

## Practical Immutability: How Long Is Forever?

### Bitcoin
- Been running since 2009 (~15 years)
- No major history rewrites
- Strong immutability in practice

### Ethereum
- Been running since 2015 (~9 years)
- One major hard fork (The DAO)
- Otherwise immutable

### Small Chains
- Can be 51% attacked
- History can be rewritten
- Immutability is weaker

**Key insight:** Immutability depends on security and economic cost to rewrite.

---

## Upgradeable Smart Contracts

Some contracts use **proxy patterns** to allow upgrades:

```
User → Proxy Contract → Implementation Contract (logic)
```

Admin can change which implementation the proxy points to.

**Tradeoff:**
- ✅ Can fix bugs
- ❌ Requires trust in admin
- ❌ Not truly immutable

**Examples:** USDC, many DeFi protocols.

---

## Code Example: Demonstrating Immutability

```typescript
import crypto from 'crypto';

interface Block {
  index: number;
  data: string;
  parentHash: string;
  hash: string;
}

function sha256(input: string): string {
  return crypto.createHash('sha256').update(input).digest('hex');
}

function createBlock(index: number, data: string, parentHash: string): Block {
  const hash = sha256(index + data + parentHash);
  return { index, data, parentHash, hash };
}

// Build chain
const blocks: Block[] = [];
blocks.push(createBlock(0, "Genesis", "0"));
blocks.push(createBlock(1, "Alice → Bob: 100 ETH", blocks[0].hash));
blocks.push(createBlock(2, "Bob → Carol: 50 ETH", blocks[1].hash));

console.log("Original chain:", blocks);

// Try to "modify" Block 1 (e.g., steal funds)
console.log("\n--- Attempting to modify Block 1 ---");
blocks[1].data = "Alice → Eve: 100 ETH"; // Attacker changes recipient

// Check chain integrity
function verifyChain(chain: Block[]): boolean {
  for (let i = 1; i < chain.length; i++) {
    const expectedParentHash = chain[i-1].hash;
    const actualParentHash = chain[i].parentHash;
    
    if (expectedParentHash !== actualParentHash) {
      console.log(`❌ Chain broken at block ${i}`);
      console.log(`Expected parent: ${expectedParentHash}`);
      console.log(`Actual parent: ${actualParentHash}`);
      return false;
    }
  }
  return true;
}

console.log("Chain valid after modification?", verifyChain(blocks)); // false

// Attacker must recompute block hash
console.log("\n--- Attacker recomputes Block 1 hash ---");
blocks[1].hash = sha256(blocks[1].index + blocks[1].data + blocks[1].parentHash);

console.log("Chain valid now?", verifyChain(blocks)); // Still false! Block 2 points to old hash

// Attacker must also update Block 2's parentHash
console.log("\n--- Attacker updates Block 2 parentHash ---");
blocks[2].parentHash = blocks[1].hash;
blocks[2].hash = sha256(blocks[2].index + blocks[2].data + blocks[2].parentHash);

console.log("Chain valid after full rewrite?", verifyChain(blocks)); // true

console.log("\nConclusion: Attacker had to recompute ALL blocks after the modification.");
console.log("In real blockchains, this requires massive computational power (PoW) or stake (PoS).");
```

---

## Exercise

### 1. Calculate Rewrite Cost

If Ethereum has 19,000,000 blocks and you want to modify Block 1, how many blocks must you recompute?

**Answer:** 18,999,999 blocks (all blocks after Block 1).

### 2. Estimate Attack Feasibility

Research the cost of a 51% attack on:
- Bitcoin
- Ethereum
- A small chain (e.g., Ethereum Classic)

Use https://www.crypto51.app for estimates.

### 3. Design Decision

For each scenario, decide if immutability is helpful or harmful:

**a) Financial transaction ledger**
- **Helpful.** Prevents fraud and retroactive changes.

**b) Medical records system**
- **Harmful.** Doctors need to correct errors.

**c) Supply chain provenance**
- **Helpful.** Prevents tampering with product history.

**d) Social media posts**
- **Harmful.** Users should be able to delete posts.

---

## Summary

**Immutability:**
- Data can't be modified or deleted (in practice)
- Achieved via hash chains, PoW/PoS, replication
- Has benefits (auditability, trust) and costs (no bug fixes, storage bloat)

**Key insights:**
- Immutability isn't absolute (hard forks can override it)
- Depends on security and economic cost to rewrite
- Sometimes harmful (medical records, content moderation)
- Smart contracts can be designed as upgradeable (tradeoff)

**When to use:**
- Financial records
- Provenance tracking
- Legal contracts

**When NOT to use:**
- Systems requiring flexibility
- Privacy-sensitive data
- Content that may need moderation

---

## Next Lesson

[→ Building a Simple Blockchain](05-building-simple-blockchain.md)
