# The Chain: Linking Blocks

🟢 **Fundamentals**

---

## The Chain Structure

A blockchain is a **linked list** where each block points to the previous block via a hash pointer.

```
Block 0 (Genesis)
  hash: 0xabc123...
       ↓
Block 1
  parentHash: 0xabc123...
  hash: 0xdef456...
       ↓
Block 2
  parentHash: 0xdef456...
  hash: 0x789ghi...
       ↓
Block 3 ...
```

---

## Hash Pointers

A **hash pointer** is a hash of the previous block's data.

```typescript
Block N-1: {
  data: "some transactions",
  hash: SHA-256(data + header) = 0xabc123...
}

Block N: {
  parentHash: 0xabc123...,
  data: "more transactions",
  hash: SHA-256(data + header + parentHash) = 0xdef456...
}
```

**Key property:** If Block N-1 changes, its hash changes, invalidating Block N.

---

## Why This Creates Immutability

### Scenario: Attacker Tries to Change History

Suppose an attacker wants to change a transaction in Block 1000:

```
Original Block 1000:
  Transaction: "Alice sends 10 ETH to Bob"
  hash: 0xabc123...

Tampered Block 1000:
  Transaction: "Alice sends 10 ETH to Eve"  (changed!)
  hash: 0xXYZ999...  (different!)
```

Now Block 1001's `parentHash` doesn't match:
```
Block 1001:
  parentHash: 0xabc123...  (expects original)
  but Block 1000's hash is now: 0xXYZ999...
  
Chain is broken! ❌
```

To fix this, the attacker must:
1. Recompute Block 1001's hash
2. Which changes Block 1002's parent hash
3. Which requires recomputing Block 1002
4. And so on, for every block up to the current block

**This is computationally expensive** (especially in Proof of Work).

---

## The Genesis Block

The **genesis block** is the first block (Block 0). It has no parent.

```typescript
{
  parentHash: "0x0000000000000000000000000000000000000000000000000000000000000000",
  number: 0,
  // ...other fields
}
```

**Special properties:**
- Hardcoded in client software
- No parent hash (or hash of zeros)
- Contains initial state (e.g., initial token distribution)

---

## Chain Forks

A **fork** occurs when two miners/validators produce blocks at the same time.

```
Block N
  ↓
Block N+1 (Miner A)
Block N+1 (Miner B)  (conflict!)
```

**Types of forks:**

### 1. Temporary Fork (Common)
- Two miners find blocks simultaneously
- Network temporarily has two chains
- Next block resolves the fork (longest chain wins)

```
Block N → Block N+1a → Block N+2  (longer, wins)
       → Block N+1b  (orphaned)
```

### 2. Hard Fork (Protocol Change)
- Protocol rules change incompatibly
- Old nodes and new nodes disagree
- Permanent chain split

**Example:** Ethereum vs Ethereum Classic (2016, after The DAO hack)

---

## Code Example: Building a Chain

```typescript
import crypto from 'crypto';

interface Block {
  index: number;
  timestamp: number;
  data: string;
  parentHash: string;
  hash: string;
}

function sha256(input: string): string {
  return crypto.createHash('sha256').update(input).digest('hex');
}

function createBlock(index: number, data: string, parentHash: string): Block {
  const timestamp = Date.now();
  const hash = sha256(index + timestamp + data + parentHash);
  
  return { index, timestamp, data, parentHash, hash };
}

// Create blockchain
const genesisBlock = createBlock(0, "Genesis Block", "0");
const block1 = createBlock(1, "Alice → Bob: 10 ETH", genesisBlock.hash);
const block2 = createBlock(2, "Bob → Carol: 5 ETH", block1.hash);

console.log("Genesis:", genesisBlock);
console.log("Block 1:", block1);
console.log("Block 2:", block2);

// Verify chain integrity
function verifyChain(blocks: Block[]): boolean {
  for (let i = 1; i < blocks.length; i++) {
    if (blocks[i].parentHash !== blocks[i-1].hash) {
      console.log(`Chain broken at block ${i}`);
      return false;
    }
  }
  return true;
}

const blockchain = [genesisBlock, block1, block2];
console.log("Chain valid?", verifyChain(blockchain)); // true

// Tamper with block1
block1.data = "Alice → Eve: 10 ETH"; // Changed!
console.log("Chain valid after tampering?", verifyChain(blockchain)); // false
```

---

## Chain Reorganization (Reorg)

A **reorg** happens when a longer chain replaces the current chain.

```
Original chain:
Block N → Block N+1a → Block N+2a

Competing chain:
Block N → Block N+1b → Block N+2b → Block N+3b  (longer!)

Network switches to the longer chain (reorg).
```

**Why this matters:**
- Transactions in Block N+1a might not be in Block N+1b
- "Confirmed" transactions can be reversed

**Solution:** Wait for multiple confirmations (e.g., 6 blocks).

---

## Chain Length vs Chain Weight

### Bitcoin/Early PoW: Longest Chain Wins
The chain with the most blocks is considered valid.

### Ethereum PoW (pre-merge): Heaviest Chain Wins
The chain with the most accumulated difficulty wins (not just length).

### Ethereum PoS (post-merge): Finality Gadget
Uses Casper FFG for finality. Once finalized, blocks cannot be reverted.

---

## Why Immutability Matters

### Positive: Trust in History
- You can verify past transactions
- No one can rewrite history (easily)

### Negative: Bugs Are Permanent
- Smart contract bugs can't be patched
- Bad data stays forever

---

## Attack: 51% Attack

If an attacker controls >50% of mining power (PoW) or stake (PoS), they can:
1. Mine a private chain faster than the honest chain
2. Release it to cause a reorg
3. Double-spend (send coins, receive goods, then revert transaction)

**Why it's expensive:**
- Need >50% of hash power (Bitcoin: billions in hardware)
- Or >50% of stake (Ethereum: billions in ETH)

**Why it's rare:**
- Economically irrational (you'd devalue your investment)
- Only feasible on small chains

---

## Exercise

### 1. Simulate a Chain

Build a 5-block chain:

```typescript
const blocks = [];
blocks.push(createBlock(0, "Genesis", "0"));

for (let i = 1; i <= 4; i++) {
  const prevBlock = blocks[i-1];
  blocks.push(createBlock(i, `Transaction ${i}`, prevBlock.hash));
}

console.log(blocks);
```

### 2. Tamper and Detect

Modify Block 2's data and verify the chain breaks:

```typescript
blocks[2].data = "Tampered!";
console.log("Valid?", verifyChain(blocks)); // false
```

### 3. Fix the Chain After Tampering

Recompute hashes for all blocks after the tampered block:

```typescript
function fixChain(blocks: Block[], startIndex: number) {
  for (let i = startIndex; i < blocks.length; i++) {
    const block = blocks[i];
    block.hash = sha256(
      block.index + block.timestamp + block.data + block.parentHash
    );
    if (i + 1 < blocks.length) {
      blocks[i + 1].parentHash = block.hash;
    }
  }
}

blocks[2].data = "Tampered!";
fixChain(blocks, 2);
console.log("Valid after fix?", verifyChain(blocks)); // true
```

This demonstrates why rewriting history requires recomputing all subsequent blocks.

---

## Summary

**The chain:**
- Linked list of blocks
- Each block points to previous via hash pointer
- Tampering breaks the chain
- Fixing requires recomputing all subsequent blocks

**Key insights:**
- Immutability comes from computational cost of rewriting history
- Genesis block has no parent
- Forks can occur (temporary or permanent)
- Longer/heavier chain wins (depends on protocol)

---

## Next Lesson

[→ Merkle Trees](03-merkle-trees.md)
