# Proof of Work (PoW)

🟡 **Intermediate**

---

## The Big Idea

**Make lying expensive.**

To propose a block, miners must solve a computationally hard puzzle. Solving requires:
- Real electricity
- Real hardware
- Real time

**Result:** Attacking the network costs more than honest mining profits.

---

## The Puzzle

Find a `nonce` such that:

```
hash(blockHeader + nonce) < target
```

**Example:**
```typescript
// Bitcoin-style PoW
interface BlockHeader {
  previousHash: string;
  timestamp: number;
  transactions: string;
  nonce: number;
}

function mine(block: BlockHeader, difficulty: number): number {
  const target = '0'.repeat(difficulty);
  let nonce = 0;
  
  while (true) {
    const hash = sha256(JSON.stringify({ ...block, nonce }));
    
    if (hash.startsWith(target)) {
      return nonce; // Found!
    }
    
    nonce++;
  }
}

// Usage
const block = {
  previousHash: '00000abc...',
  timestamp: Date.now(),
  transactions: 'tx1, tx2, tx3',
  nonce: 0
};

const solution = mine(block, 4); // Find hash starting with '0000'
console.log(`Solution found: nonce = ${solution}`);
```

---

## Why This Works

### 1. Hard to Solve
Average attempts needed: `2^difficulty`

**Example:** Difficulty = 4 (hash starts with '0000')
- Probability per attempt: `1 / 16^4 = 1 / 65,536`
- Expected attempts: ~65,536 tries

### 2. Easy to Verify
Once found, anyone can verify:
```typescript
function verify(block: BlockHeader, nonce: number, difficulty: number): boolean {
  const hash = sha256(JSON.stringify({ ...block, nonce }));
  return hash.startsWith('0'.repeat(difficulty));
}
```

### 3. Unfakeable
You can't shortcut SHA-256. Only way to find solution: brute force.

---

## Difficulty Adjustment

**Problem:** Mining speed depends on total computing power.
- More miners → blocks found faster
- Fewer miners → blocks found slower

**Solution:** Adjust difficulty to maintain consistent block time.

```typescript
function adjustDifficulty(
  currentDifficulty: number,
  targetBlockTime: number, // seconds
  actualBlockTime: number
): number {
  if (actualBlockTime < targetBlockTime) {
    return currentDifficulty + 1; // Make harder
  } else if (actualBlockTime > targetBlockTime * 2) {
    return currentDifficulty - 1; // Make easier
  }
  return currentDifficulty;
}
```

**Bitcoin:** Adjusts every 2016 blocks (~2 weeks) to maintain 10-minute blocks.

**Ethereum (pre-merge):** Adjusted every block to maintain ~13 seconds.

---

## Economic Security

### Attack Cost

To control the network (51% attack):
- Need > 50% of total hash power
- Must outpace honest miners

**Bitcoin example (2024):**
- Total hash rate: ~400 EH/s (exahashes/second)
- Cost to match: ~$10 billion in hardware + electricity

### Attack Profit

What can attacker do?
1. **Double-spend:** Reverse their own transactions
2. **Censor transactions:** Refuse to include certain transactions

What attacker **cannot** do:
- Steal other people's coins (still need private keys)
- Mint unlimited coins (rules enforced by full nodes)
- Change past blocks (would need to redo all work since that block)

**Result:** Attack costs more than potential profit.

---

## The Longest Chain Rule

When multiple valid blocks exist, follow the chain with most work:

```
        Block 100A
       /
Block 99 
       \
        Block 100B
```

Miners choose which to build on. Eventually one chain grows longer. Short chain is abandoned (orphaned).

**Why this matters:**
- Transactions aren't final immediately
- Must wait for confirmations (more blocks on top)
- More confirmations = exponentially harder to reverse

**Bitcoin convention:**
- 1 confirmation: ~10 minutes
- 6 confirmations (~1 hour): Considered final for large transactions

---

## Energy Consumption

**Reality check:** PoW uses massive energy.

**Bitcoin (2024):**
- Power consumption: ~150 TWh/year
- Comparable to: Argentina's total electricity use
- CO2 emissions: ~65 million tons/year

**Why?**
- Security comes from making attacks expensive
- More hash power = harder to attack = more electricity

**Tradeoff:** Security vs. environmental cost.

---

## Centralization Pressures

**Theory:** Anyone can mine.

**Reality:** Mining centralizes toward:

### 1. Cheap Electricity
Miners relocate to areas with cheapest power (Iceland, Kazakhstan, Texas).

### 2. Specialized Hardware (ASICs)
- **ASIC:** Application-Specific Integrated Circuit
- Chips designed only for mining
- Orders of magnitude more efficient than GPUs
- Expensive, requires manufacturing scale

### 3. Mining Pools
Individual miners join pools to reduce variance:
- Pool mines collectively
- Rewards split proportionally
- Top pools control significant hash power

**Result:** Mining more centralized than originally envisioned.

---

## Selfish Mining

**Attack:** Miner withholds found blocks, releases strategically.

**How it works:**
1. Attacker mines block, keeps it secret
2. Honest network mines on old chain
3. When honest network finds block, attacker releases their secret chain
4. Attacker's chain has equal length, network splits
5. Attacker keeps mining on their chain, has head start

**Result:** Attacker gets more than fair share of rewards.

**Mitigation:** Requires >25% hash power. Risky. Detected quickly.

---

## Summary

**Proof of Work:**
- Make block creation expensive (computation)
- Difficult to solve, easy to verify
- Difficulty adjusts to maintain block time
- Economic security: attacking costs more than profit

**Tradeoffs:**
- ✅ Simple, well-tested (Bitcoin since 2009)
- ✅ Permissionless (anyone can mine)
- ❌ Massive energy consumption
- ❌ Slow finality (need multiple confirmations)
- ❌ Centralizing toward cheap electricity and ASICs

**Key insight:** "Slow" isn't a bug, it's the cost of security without trust.

---

## Exercise

Implement a blockchain with PoW:

```typescript
import crypto from 'crypto';

class Block {
  constructor(
    public index: number,
    public timestamp: number,
    public data: string,
    public previousHash: string,
    public nonce: number = 0,
    public hash: string = ''
  ) {}
}

class Blockchain {
  chain: Block[] = [];
  difficulty: number = 4;

  constructor() {
    this.chain.push(this.createGenesisBlock());
  }

  createGenesisBlock(): Block {
    const block = new Block(0, Date.now(), 'Genesis Block', '0');
    block.hash = this.calculateHash(block);
    return block;
  }

  calculateHash(block: Block): string {
    return crypto
      .createHash('sha256')
      .update(
        block.index +
        block.timestamp +
        block.data +
        block.previousHash +
        block.nonce
      )
      .digest('hex');
  }

  mineBlock(block: Block): Block {
    const target = '0'.repeat(this.difficulty);
    
    while (!block.hash.startsWith(target)) {
      block.nonce++;
      block.hash = this.calculateHash(block);
    }
    
    console.log(`Block mined: ${block.hash}`);
    return block;
  }

  addBlock(data: string): void {
    const previousBlock = this.chain[this.chain.length - 1];
    const newBlock = new Block(
      previousBlock.index + 1,
      Date.now(),
      data,
      previousBlock.hash
    );
    
    this.mineBlock(newBlock);
    this.chain.push(newBlock);
  }

  isValid(): boolean {
    for (let i = 1; i < this.chain.length; i++) {
      const current = this.chain[i];
      const previous = this.chain[i - 1];

      // Verify hash
      if (current.hash !== this.calculateHash(current)) {
        return false;
      }

      // Verify link
      if (current.previousHash !== previous.hash) {
        return false;
      }

      // Verify PoW
      if (!current.hash.startsWith('0'.repeat(this.difficulty))) {
        return false;
      }
    }
    return true;
  }
}

// Test it
const blockchain = new Blockchain();

console.time('Mining Block 1');
blockchain.addBlock('First transaction');
console.timeEnd('Mining Block 1');

console.time('Mining Block 2');
blockchain.addBlock('Second transaction');
console.timeEnd('Mining Block 2');

console.log('Is blockchain valid?', blockchain.isValid());

// Try tampering
blockchain.chain[1].data = 'Tampered data';
console.log('After tampering:', blockchain.isValid());
```

**Questions:**
1. How long does mining take at difficulty 4 vs 5?
2. What happens if you change data in a middle block?
3. How would you implement the longest chain rule?

---

## Next Lesson

[→ Proof of Stake](03-proof-of-stake.md)
