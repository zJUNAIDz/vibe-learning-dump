# Building a Simple Blockchain

🟢 **Fundamentals**

---

## Overview

Let's build a minimal blockchain from scratch to solidify your understanding.

**What we'll build:**
- Blocks with transactions
- Hash-based linking
- Chain validation
- Simple Proof of Work

**Goal:** Understand the mechanics, not build production code.

---

## Step 1: Define the Block Structure

```typescript
import crypto from 'crypto';

interface Transaction {
  from: string;
  to: string;
  amount: number;
}

interface Block {
  index: number;
  timestamp: number;
  transactions: Transaction[];
  parentHash: string;
  hash: string;
  nonce: number; // For Proof of Work
}

function sha256(input: string): string {
  return crypto.createHash('sha256').update(input).digest('hex');
}
```

---

## Step 2: Calculate Block Hash

```typescript
function calculateHash(block: Pick<Block, 'index' | 'timestamp' | 'transactions' | 'parentHash' | 'nonce'>): string {
  const data = block.index +
               block.timestamp +
               JSON.stringify(block.transactions) +
               block.parentHash +
               block.nonce;
  
  return sha256(data);
}
```

---

## Step 3: Create Genesis Block

```typescript
function createGenesisBlock(): Block {
  const genesisBlock: Block = {
    index: 0,
    timestamp: Date.now(),
    transactions: [],
    parentHash: "0",
    hash: "",
    nonce: 0
  };
  
  genesisBlock.hash = calculateHash(genesisBlock);
  return genesisBlock;
}
```

---

## Step 4: Mine a Block (Proof of Work)

```typescript
function mineBlock(block: Block, difficulty: number): Block {
  const target = "0".repeat(difficulty); // e.g., "0000" for difficulty 4
  
  while (!block.hash.startsWith(target)) {
    block.nonce++;
    block.hash = calculateHash(block);
  }
  
  console.log(`Block mined! Hash: ${block.hash}, Nonce: ${block.nonce}`);
  return block;
}
```

**Proof of Work:**
- Find a nonce such that the hash starts with N zeros
- Higher difficulty = more computation required

---

## Step 5: Add New Block

```typescript
function createBlock(
  index: number,
  transactions: Transaction[],
  parentHash: string,
  difficulty: number
): Block {
  const block: Block = {
    index,
    timestamp: Date.now(),
    transactions,
    parentHash,
    hash: "",
    nonce: 0
  };
  
  return mineBlock(block, difficulty);
}
```

---

## Step 6: Blockchain Class

```typescript
class Blockchain {
  chain: Block[];
  difficulty: number;
  pendingTransactions: Transaction[];
  
  constructor(difficulty: number = 2) {
    this.chain = [createGenesisBlock()];
    this.difficulty = difficulty;
    this.pendingTransactions = [];
  }
  
  getLatestBlock(): Block {
    return this.chain[this.chain.length - 1];
  }
  
  addTransaction(transaction: Transaction): void {
    this.pendingTransactions.push(transaction);
  }
  
  minePendingTransactions(): void {
    const block = createBlock(
      this.chain.length,
      this.pendingTransactions,
      this.getLatestBlock().hash,
      this.difficulty
    );
    
    this.chain.push(block);
    this.pendingTransactions = [];
  }
  
  isValid(): boolean {
    for (let i = 1; i < this.chain.length; i++) {
      const currentBlock = this.chain[i];
      const previousBlock = this.chain[i - 1];
      
      // Verify hash
      if (currentBlock.hash !== calculateHash(currentBlock)) {
        console.log(`❌ Block ${i} has invalid hash`);
        return false;
      }
      
      // Verify chain linkage
      if (currentBlock.parentHash !== previousBlock.hash) {
        console.log(`❌ Block ${i} has invalid parent hash`);
        return false;
      }
      
      // Verify Proof of Work
      const target = "0".repeat(this.difficulty);
      if (!currentBlock.hash.startsWith(target)) {
        console.log(`❌ Block ${i} doesn't satisfy PoW`);
        return false;
      }
    }
    
    return true;
  }
  
  printChain(): void {
    console.log(JSON.stringify(this.chain, null, 2));
  }
}
```

---

## Step 7: Use the Blockchain

```typescript
// Create blockchain with difficulty 3
const blockchain = new Blockchain(3);

console.log("Mining block 1...");
blockchain.addTransaction({ from: "Alice", to: "Bob", amount: 100 });
blockchain.addTransaction({ from: "Bob", to: "Carol", amount: 50 });
blockchain.minePendingTransactions();

console.log("\nMining block 2...");
blockchain.addTransaction({ from: "Carol", to: "Dave", amount: 25 });
blockchain.minePendingTransactions();

console.log("\nBlockchain:");
blockchain.printChain();

console.log("\nChain valid?", blockchain.isValid());

// Tamper with block
console.log("\n--- Tampering with Block 1 ---");
blockchain.chain[1].transactions[0].amount = 1000; // Change amount

console.log("Chain valid after tampering?", blockchain.isValid());
```

---

## Example Output

```
Mining block 1...
Block mined! Hash: 000a1b2c3d4e5f6789..., Nonce: 12543

Mining block 2...
Block mined! Hash: 0007f6e5d4c3b2a1..., Nonce: 8721

Blockchain:
[
  {
    "index": 0,
    "timestamp": 1705843200000,
    "transactions": [],
    "parentHash": "0",
    "hash": "abc123...",
    "nonce": 0
  },
  {
    "index": 1,
    "timestamp": 1705843215000,
    "transactions": [
      { "from": "Alice", "to": "Bob", "amount": 100 },
      { "from": "Bob", "to": "Carol", "amount": 50 }
    ],
    "parentHash": "abc123...",
    "hash": "000a1b2c3d4e5f6789...",
    "nonce": 12543
  },
  ...
]

Chain valid? true

--- Tampering with Block 1 ---
❌ Block 1 has invalid hash
Chain valid after tampering? false
```

---

## Understanding Proof of Work

### Difficulty 1
```
Hash must start with "0"
Example: 0a1b2c3d...
Average attempts: 16
```

### Difficulty 2
```
Hash must start with "00"
Example: 00a1b2c3...
Average attempts: 256
```

### Difficulty 4
```
Hash must start with "0000"
Example: 0000a1b2...
Average attempts: 65,536
```

**Key insight:** Each additional zero multiplies attempts by ~16.

**Bitcoin difficulty:** ~19 leading zeros (~$10^{22}$ attempts).

---

## What's Missing from Our Implementation?

### 1. Networking
Real blockchains distribute blocks across nodes.

### 2. Consensus
Nodes must agree on which chain is valid (longest chain, heaviest chain, finality, etc.).

### 3. Transaction Signatures
Transactions should be signed with private keys.

### 4. UTXO or Account Model
Track balances properly (our example doesn't validate balances).

### 5. Mempool
Pending transactions wait in a pool before being mined.

### 6. Incentives
Miners/validators need rewards (block rewards + transaction fees).

---

## Exercise

### 1. Adjust Difficulty

Try different difficulties and measure mining time:

```typescript
const start = Date.now();
const blockchain = new Blockchain(4); // difficulty 4
blockchain.addTransaction({ from: "Alice", to: "Bob", amount: 100 });
blockchain.minePendingTransactions();
const end = Date.now();
console.log(`Mining took ${end - start}ms`);
```

### 2. Implement Transaction Validation

Add logic to verify:
- Sender has sufficient balance
- Transaction is signed (simplified: check sender isn't empty)

```typescript
addTransaction(transaction: Transaction): void {
  if (!transaction.from || !transaction.to) {
    throw new Error("Transaction must have from and to");
  }
  
  // In real blockchain, verify signature here
  
  this.pendingTransactions.push(transaction);
}
```

### 3. Calculate Total Supply

Write a function to calculate total coins in circulation:

```typescript
getTotalSupply(): number {
  let supply = 0;
  for (const block of this.chain) {
    for (const tx of block.transactions) {
      supply += tx.amount;
    }
  }
  return supply;
}
```

(Note: This is oversimplified. Real blockchains track balances more carefully.)

---

## Full Code

Here's the complete implementation:

```typescript
import crypto from 'crypto';

interface Transaction {
  from: string;
  to: string;
  amount: number;
}

interface Block {
  index: number;
  timestamp: number;
  transactions: Transaction[];
  parentHash: string;
  hash: string;
  nonce: number;
}

function sha256(input: string): string {
  return crypto.createHash('sha256').update(input).digest('hex');
}

function calculateHash(block: Pick<Block, 'index' | 'timestamp' | 'transactions' | 'parentHash' | 'nonce'>): string {
  const data = block.index +
               block.timestamp +
               JSON.stringify(block.transactions) +
               block.parentHash +
               block.nonce;
  return sha256(data);
}

function createGenesisBlock(): Block {
  const genesisBlock: Block = {
    index: 0,
    timestamp: Date.now(),
    transactions: [],
    parentHash: "0",
    hash: "",
    nonce: 0
  };
  genesisBlock.hash = calculateHash(genesisBlock);
  return genesisBlock;
}

function mineBlock(block: Block, difficulty: number): Block {
  const target = "0".repeat(difficulty);
  while (!block.hash.startsWith(target)) {
    block.nonce++;
    block.hash = calculateHash(block);
  }
  console.log(`Block mined! Hash: ${block.hash.substring(0, 20)}..., Nonce: ${block.nonce}`);
  return block;
}

function createBlock(
  index: number,
  transactions: Transaction[],
  parentHash: string,
  difficulty: number
): Block {
  const block: Block = {
    index,
    timestamp: Date.now(),
    transactions,
    parentHash,
    hash: "",
    nonce: 0
  };
  return mineBlock(block, difficulty);
}

class Blockchain {
  chain: Block[];
  difficulty: number;
  pendingTransactions: Transaction[];
  
  constructor(difficulty: number = 2) {
    this.chain = [createGenesisBlock()];
    this.difficulty = difficulty;
    this.pendingTransactions = [];
  }
  
  getLatestBlock(): Block {
    return this.chain[this.chain.length - 1];
  }
  
  addTransaction(transaction: Transaction): void {
    this.pendingTransactions.push(transaction);
  }
  
  minePendingTransactions(): void {
    console.log(`Mining block ${this.chain.length}...`);
    const block = createBlock(
      this.chain.length,
      this.pendingTransactions,
      this.getLatestBlock().hash,
      this.difficulty
    );
    this.chain.push(block);
    this.pendingTransactions = [];
  }
  
  isValid(): boolean {
    for (let i = 1; i < this.chain.length; i++) {
      const currentBlock = this.chain[i];
      const previousBlock = this.chain[i - 1];
      
      if (currentBlock.hash !== calculateHash(currentBlock)) {
        console.log(`❌ Block ${i} has invalid hash`);
        return false;
      }
      
      if (currentBlock.parentHash !== previousBlock.hash) {
        console.log(`❌ Block ${i} has invalid parent hash`);
        return false;
      }
      
      const target = "0".repeat(this.difficulty);
      if (!currentBlock.hash.startsWith(target)) {
        console.log(`❌ Block ${i} doesn't satisfy PoW`);
        return false;
      }
    }
    return true;
  }
}

// Demo
const blockchain = new Blockchain(3);
blockchain.addTransaction({ from: "Alice", to: "Bob", amount: 100 });
blockchain.addTransaction({ from: "Bob", to: "Carol", amount: 50 });
blockchain.minePendingTransactions();

blockchain.addTransaction({ from: "Carol", to: "Dave", amount: 25 });
blockchain.minePendingTransactions();

console.log("\nChain valid?", blockchain.isValid());
```

---

## Summary

**You've built:**
- A block structure with transactions
- Hash-based chain linking
- Proof of Work mining
- Chain validation

**Key insights:**
- Blockchains are append-only linked lists
- PoW makes rewriting history expensive
- Tampering breaks the chain
- Real blockchains add networking, consensus, incentives

---

## Module Complete!

You've finished **Module 02: Blockchain Data Structures**.

**You should now understand:**
- ✅ Block structure (header + body)
- ✅ Hash pointers creating immutability
- ✅ Merkle trees for efficient verification
- ✅ Immutability (benefits and costs)
- ✅ How to build a simple blockchain

---

## Next Module

[→ Module 03: Consensus](../03-consensus/)

Learn how nodes agree on state without trust (Byzantine Generals Problem, Proof of Work, Proof of Stake).
