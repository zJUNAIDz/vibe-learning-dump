# Finality and Reorganizations

🔴 **Advanced**

---

## What is Finality?

**Finality:** The guarantee that a transaction cannot be reversed.

In traditional databases: Instant (once committed, it's permanent).

In blockchains: **It's complicated.**

---

## Types of Finality

### 1. Probabilistic Finality (PoW)

Confidence increases with each confirmation, but never reaches 100%.

```
Confirmations | Probability of Reversal
--------------|-------------------------
0             | ~50% (mempool)
1             | ~10%
2             | ~1%
3             | ~0.1%
6             | ~0.0001% (Bitcoin standard)
```

**Mathematics:**
```
P(reversal) ≈ (q/p)^n

where:
- q = attacker's hash power (e.g., 0.4 = 40%)
- p = honest hash power (e.g., 0.6 = 60%)
- n = number of confirmations
```

**Example:** Attacker with 30% hash power
```typescript
function reversalProbability(
  attackerHashPower: number,
  confirmations: number
): number {
  const honestHashPower = 1 - attackerHashPower;
  return Math.pow(attackerHashPower / honestHashPower, confirmations);
}

console.log('1 conf:', reversalProbability(0.3, 1)); // ~42%
console.log('3 conf:', reversalProbability(0.3, 3)); // ~6.4%
console.log('6 conf:', reversalProbability(0.3, 6)); // ~0.4%
```

---

### 2. Economic Finality (PoS)

Transaction is final when reversing it would cost more than any possible gain.

**Ethereum example:**
- After 2 epochs (~13 minutes), block is "finalized"
- Reversing requires 1/3 of validators to get slashed
- At 15M ETH staked → 5M ETH destroyed (~$10 billion)

**Greater than any double-spend profit.**

```typescript
interface Checkpoint {
  epoch: number;
  blockHash: string;
  attestations: number; // Number of validators attesting
}

function isFinalized(
  checkpoint: Checkpoint,
  totalValidators: number
): boolean {
  const twoThirdsMajority = (totalValidators * 2) / 3;
  return checkpoint.attestations >= twoThirdsMajority;
}
```

---

### 3. Absolute Finality (BFT)

Some consensus protocols guarantee finality immediately.

**Examples:** Tendermint (Cosmos), Algorand

**How:** Classic BFT consensus (requires 2/3+ votes).

**Tradeoff:** 
- ✅ Instant finality
- ❌ Liveness risk (network halt if <2/3 validators online)
- ❌ Often more centralized (smaller validator sets)

---

## Chain Reorganizations (Reorgs)

**Reorg:** When the chain's history is rewritten.

### How Reorgs Happen

**Normal case:**
```
Block 99 → Block 100 → Block 101
```

**Fork appears:**
```
                Block 100A → Block 101A
               /
Block 99 → Block 100B → Block 101B → Block 102B
```

If chain B grows longer, nodes switch to it. **Block 100A and 101A are abandoned.**

Transactions in abandoned blocks:
- Return to mempool
- May be included in new chain
- Or may become invalid (if conflicting tx in Block 100B)

---

### Reorg Example

```typescript
interface Block {
  number: number;
  hash: string;
  parentHash: string;
  transactions: string[];
}

class BlockchainNode {
  chains: Map<string, Block[]> = new Map();
  canonicalChain: string = 'main';

  addBlock(chainId: string, block: Block): void {
    if (!this.chains.has(chainId)) {
      this.chains.set(chainId, []);
    }
    this.chains.get(chainId)!.push(block);

    // Check if we need to reorg
    this.checkReorg();
  }

  checkReorg(): void {
    let longestLength = 0;
    let longestChain = '';

    for (const [id, chain] of this.chains) {
      if (chain.length > longestLength) {
        longestLength = chain.length;
        longestChain = id;
      }
    }

    if (longestChain !== this.canonicalChain) {
      console.log(`REORG: Switching from ${this.canonicalChain} to ${longestChain}`);
      this.canonicalChain = longestChain;
    }
  }

  getOrphanedTransactions(oldChain: string, newChain: string): string[] {
    const oldBlocks = this.chains.get(oldChain) || [];
    const newBlocks = this.chains.get(newChain) || [];

    const newTxs = new Set(
      newBlocks.flatMap(b => b.transactions)
    );

    return oldBlocks
      .flatMap(b => b.transactions)
      .filter(tx => !newTxs.has(tx));
  }
}
```

---

## Real-World Reorg Events

### Bitcoin (2013): 24-block reorg
- Two mining pools on different versions
- 24 blocks reorganized (~4 hours)
- No double-spends, but significant disruption

### Ethereum Classic (2020): 3,693-block reorg
- 51% attack
- Attacker mined private chain, then released
- Double-spent $5.6 million

### Ethereum (2021): 7-block reorg
- Two proposers in same slot (network timing issues)
- Reorganized 7 blocks (~84 seconds)
- No malicious intent, just bad luck

---

## Finality from Application Perspective

### How many confirmations do I need?

**Depends on transaction value:**

```typescript
function requiredConfirmations(
  amountUSD: number,
  consensusType: 'pow' | 'pos'
): number {
  if (consensusType === 'pow') {
    if (amountUSD < 100) return 1;      // ~10 min
    if (amountUSD < 10000) return 3;    // ~30 min
    return 6;                            // ~60 min (standard)
  } else {
    if (amountUSD < 100) return 1;      // ~12 sec
    if (amountUSD < 10000) return 32;   // ~6.4 min (1 epoch)
    return 64;                           // ~13 min (2 epochs, finalized)
  }
}
```

### Implementing Confirmation Tracking

```typescript
import { ethers } from 'ethers';

async function monitorTransaction(
  provider: ethers.Provider,
  txHash: string,
  requiredConfirmations: number
): Promise<void> {
  console.log(`Monitoring ${txHash}...`);

  // Wait for transaction to be mined
  const receipt = await provider.waitForTransaction(txHash);
  console.log(`Mined in block ${receipt.blockNumber}`);

  // Wait for confirmations
  await provider.waitForTransaction(txHash, requiredConfirmations);
  console.log(`${requiredConfirmations} confirmations reached`);

  // Even "final" transactions can be reorged in PoW if < 6 confirmations
  // In PoS, after finality (~13 min), extremely unlikely
}

// Listen for reorgs
provider.on('block', async (blockNumber) => {
  const block = await provider.getBlock(blockNumber);
  console.log(`New block: ${blockNumber}, hash: ${block.hash}`);
  
  // Check if parent matches expected
  // If not, possible reorg detected
});
```

---

## Handling Reorgs in Applications

### Strategy 1: Wait for Finality

```typescript
async function processSafePayment(
  provider: ethers.Provider,
  txHash: string
): Promise<void> {
  // Wait for economic finality (Ethereum PoS)
  await provider.waitForTransaction(txHash, 64); // 2 epochs
  
  // Now safe to fulfill order, ship product, etc.
  await fulfillOrder();
}
```

**Tradeoff:** Slow UX (13+ minutes).

---

### Strategy 2: Optimistic with Revert Capability

```typescript
class PaymentProcessor {
  pendingPayments: Map<string, Payment> = new Map();

  async processOptimistically(txHash: string): Promise<void> {
    // Wait for 1 confirmation
    await provider.waitForTransaction(txHash, 1);
    
    // Optimistically fulfill
    const payment = await this.fulfillOrder(txHash);
    this.pendingPayments.set(txHash, payment);
    
    // Monitor for finality
    this.monitorForReorg(txHash);
  }

  async monitorForReorg(txHash: string): Promise<void> {
    try {
      await provider.waitForTransaction(txHash, 64);
      // Finalized, remove from pending
      this.pendingPayments.delete(txHash);
    } catch (error) {
      // Transaction disappeared (reorged)
      console.error(`Reorg detected for ${txHash}`);
      const payment = this.pendingPayments.get(txHash);
      await this.revertOrder(payment);
    }
  }
}
```

**Tradeoff:** Better UX, but must handle reversals.

---

### Strategy 3: Probabilistic Risk Management

```typescript
function shouldFulfillImmediately(
  amountUSD: number,
  confirmations: number,
  chainType: 'pow' | 'pos'
): boolean {
  const reorgRisk = calculateReorgRisk(confirmations, chainType);
  
  // Risk-adjusted decision
  const expectedLoss = amountUSD * reorgRisk;
  const acceptableRisk = 1; // $1 loss acceptable
  
  return expectedLoss < acceptableRisk;
}

// Example
// $10 payment, 1 confirmation, PoW
// Reorg risk: ~10%
// Expected loss: $1
// → Fulfill immediately

// $1000 payment, 1 confirmation, PoW  
// Expected loss: $100
// → Wait for more confirmations
```

---

## MEV and Reorgs

**MEV (Maximal Extractable Value):** Profit from transaction ordering.

**Reorg for profit:**
1. See profitable transaction in mempool
2. Mine competing block with your transaction first
3. Reorg to make your block canonical

**Example:** Arbitrage opportunities, liquidations

**Mitigation:**
- Ethereum PoS: Single proposer per slot (harder to reorg)
- Flashbots: Off-chain order flow
- Threshold encryption: Hide tx contents until inclusion

---

## Summary

**Finality Types:**
- **Probabilistic (PoW):** Never 100%, increases exponentially with confirmations
- **Economic (PoS):** Reversal costs more than gain after finality threshold
- **Absolute (BFT):** Immediate, but liveness risks

**Reorgs:**
- Normal in PoW (short reorgs common)
- Rare in PoS (but still possible before finality)
- Can cause tx reversals, orphaned blocks

**Practical Advice:**
- **Low value:** 1-2 confirmations (~1-2 minutes)
- **Medium value:** 1 epoch (~6 minutes on Ethereum PoS)
- **High value:** Economic finality (~13 minutes on Ethereum PoS, 60 minutes on Bitcoin)
- **Critical value:** Wait longer or use application-layer escrow

**Key takeaway:** "Confirmed" doesn't mean "final". Build applications that handle uncertainty gracefully.

---

## Exercise

Build a finality monitor:

```typescript
import { ethers } from 'ethers';

class FinalityMonitor {
  provider: ethers.Provider;
  
  constructor(providerUrl: string) {
    this.provider = new ethers.JsonRpcProvider(providerUrl);
  }

  async trackFinality(txHash: string): Promise<void> {
    console.log(`Tracking finality for ${txHash}\n`);

    const receipt = await this.provider.getTransactionReceipt(txHash);
    if (!receipt) {
      console.log('Transaction not yet mined');
      return;
    }

    const minedBlock = receipt.blockNumber;
    console.log(`Mined in block ${minedBlock}`);

    // Monitor confirmations
    this.provider.on('block', async (blockNumber) => {
      const confirmations = blockNumber - minedBlock;
      const probability = this.reversalProbability(confirmations);
      
      console.log(
        `Block ${blockNumber}: ${confirmations} confirmations, ` +
        `reversal probability: ${(probability * 100).toFixed(4)}%`
      );

      if (confirmations >= 6) {
        console.log('\n✅ Transaction finalized (6+ confirmations)');
        this.provider.removeAllListeners('block');
      }
    });
  }

  reversalProbability(confirmations: number): number {
    // Assume attacker has 30% hash power
    const q = 0.3;
    const p = 0.7;
    return Math.pow(q / p, confirmations);
  }
}

// Usage
const monitor = new FinalityMonitor('https://eth.llamarpc.com');
monitor.trackFinality('0x...');
```

**Extension ideas:**
1. Add reorg detection (check if block hash changes)
2. Implement different finality rules for PoW vs PoS
3. Alert when transaction disappears from chain (reorged)

---

**[← Previous: Comparing Consensus](04-comparing-consensus.md)** | **[↑ Back to Module](README.md)** | **[Next Module: Accounts and Transactions →](../04-accounts-transactions/README.md)**
