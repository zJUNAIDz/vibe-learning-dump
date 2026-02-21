# Comparing Consensus Mechanisms

🟡 **Intermediate**

---

## Side-by-Side Comparison

| Aspect | Proof of Work (PoW) | Proof of Stake (PoS) |
|--------|---------------------|----------------------|
| **Resource** | Computation (electricity) | Capital (staked crypto) |
| **Participation** | Anyone with hardware | Anyone with minimum stake |
| **Selection** | Competitive race | Pseudo-random (weighted by stake) |
| **Block Time** | Variable (probabilistic) | Fixed slots |
| **Finality** | Probabilistic (need confirmations) | Economic finality (~15 min on Ethereum) |
| **Energy** | ~150 TWh/year (Bitcoin) | ~0.01 TWh/year (Ethereum) |
| **51% Attack Cost** | Rent/buy >50% hash power | Buy >50% of staked supply |
| **Attack Penalty** | None (can sell hardware) | Slashing (lose stake) |
| **Barrier to Entry** | Hardware + electricity costs | Minimum stake (e.g., 32 ETH) |
| **Centralization Risk** | Pools, cheap electricity, ASICs | Exchanges, whales, staking pools |
| **Track Record** | Bitcoin since 2009 (15+ years) | Ethereum PoS since 2022 (~2 years at scale) |

---

## Security Model Comparison

### Proof of Work

**Assumption:** Honest majority of hash power.

**Why it works:**
- Attacking costs electricity (ongoing expense)
- Can't reuse prior computation
- If you have hash power to attack, more profitable to mine honestly

**Attack scenario:**
1. Rent 51% hash power
2. Double-spend
3. Return rented hardware

**Defense:** Hash power not easily rentable at scale. Hardware expensive.

---

### Proof of Stake

**Assumption:** Honest majority of stake.

**Why it works:**
- Attacking loses your capital (slashing)
- Must buy majority of supply (drives price up)
- If you have majority stake, attacking devalues your holdings

**Attack scenario:**
1. Buy 51% of staked supply
2. Attack network
3. Token price crashes
4. Your holdings worthless + slashed

**Defense:** Economic irrationality. Shooting yourself in the foot.

---

## When Each Makes Sense

### Use PoW When:

✅ **Maximum decentralization priority**
- Anyone can mine (no minimum capital)
- Geographic distribution (follow cheap energy)

✅ **Simple, proven security model**
- No slashing conditions to implement
- No weak subjectivity

✅ **Long-term immutability critical**
- Objective consensus (no social layer needed for long-range attacks)

**Examples:** Bitcoin (digital gold, store of value)

---

### Use PoS When:

✅ **Energy efficiency matters**
- 99.95% less energy than PoW
- Regulatory/environmental concerns

✅ **Faster finality needed**
- Minutes vs hours
- Better UX for applications

✅ **Lower issuance desired**
- Don't need to pay electricity costs
- Less inflation

**Examples:** Ethereum (smart contract platform), Cardano, Solana

---

## Hybrid Approaches

### Proof of Authority (PoA)
- Known validators (companies, institutions)
- No mining or staking
- Fast, efficient
- **Tradeoff:** Centralized (permissioned)

**Use case:** Enterprise blockchains, testnets

---

### Proof of Space/Time
- Use hard drive space instead of computation
- **Example:** Chia

**Tradeoff:** Still requires hardware, less tested

---

### Proof of Burn
- "Burn" cryptocurrency to earn mining rights
- Send coins to unspendable address

**Tradeoff:** Wasteful in different way (destroying money vs electricity)

---

## The Scalability Dimension

**PoW:**
- Block time limited by orphan rate (network propagation)
- Faster blocks → more orphans → less security
- Bitcoin: 10 min, Ethereum (pre-merge): ~13 sec

**PoS:**
- Faster slots possible (single proposer, no race)
- Ethereum: 12-second slots
- Solana: 400ms slots (aggressive parameters)

**Key insight:** PoS enables faster consensus, but doesn't solve scalability alone. Still need sharding/layer 2.

---

## Real-World Performance

### Bitcoin (PoW)
- **TPS:** ~7 transactions/second
- **Finality:** 6 confirmations = ~60 minutes
- **Energy:** ~150 TWh/year
- **Security budget:** ~$10B/year (block rewards + fees)

### Ethereum (PoS, post-merge)
- **TPS:** ~15 transactions/second (base layer)
- **Finality:** 2 epochs = ~13 minutes
- **Energy:** ~0.01 TWh/year
- **Security budget:** ~$2B/year (much lower issuance)

### Solana (PoS variant)
- **TPS:** ~3,000 transactions/second (claimed)
- **Finality:** ~1 second (optimistic)
- **Energy:** Low (PoS)
- **Tradeoff:** More centralized (high hardware requirements)

---

## The Impossible Trade-off

No consensus mechanism solves all problems:

```
        Security
           /\
          /  \
         /    \
        /      \
       /________\
  Scalability  Decentralization
  
  (Blockchain Trilemma)
```

**PoW:** Chooses Security + Decentralization, sacrifices Scalability
**PoS:** Improves Scalability slightly, but still trade-offs
**DPoS:** Chooses Scalability, sacrifices Decentralization

---

## Migration Challenges (PoW → PoS)

Ethereum's experience:

### What Went Smoothly:
- Historical data preserved
- Consensus switched seamlessly
- 99.95% energy reduction achieved

### What Was Hard:
- Took ~5 years of R&D
- Required coordinating entire network
- Complexity increased (slashing, withdrawals)
- Staking pools centralized (Lido ~30% of stake)

### What's Still Unknown:
- Long-term security at scale (only ~2 years)
- Centralization trends (exchanges, pools)
- Liquid staking derivatives risks

---

## Developer Implications

### For PoW Chains:
```typescript
// Wait for confirmations
async function waitForConfirmations(
  txHash: string,
  confirmations: number
): Promise<void> {
  const receipt = await provider.getTransactionReceipt(txHash);
  const currentBlock = await provider.getBlockNumber();
  
  const blocksConfirmed = currentBlock - receipt.blockNumber;
  
  if (blocksConfirmed < confirmations) {
    // Not final yet!
    throw new Error(`Only ${blocksConfirmed} confirmations`);
  }
}

// Usage: high-value transactions
await waitForConfirmations(txHash, 6); // ~1 hour on Bitcoin
```

### For PoS Chains:
```typescript
// Finality is faster but still need to wait
async function waitForFinality(txHash: string): Promise<void> {
  const receipt = await provider.getTransactionReceipt(txHash);
  
  // On Ethereum PoS: finality at next epoch boundary
  // Usually ~13 minutes
  
  if (receipt.status === 1) {
    // Transaction included, but wait for finality
    await new Promise(resolve => setTimeout(resolve, 15 * 60 * 1000)); // 15 min
  }
}
```

**Key point:** Even PoS isn't instant. Applications need to handle pending states.

---

## Future Directions

### Research Areas:

**1. Post-quantum security**
- Current signatures vulnerable to quantum computers
- Need quantum-resistant algorithms

**2. MEV (Maximal Extractable Value)**
- PoS makes transaction ordering more explicit
- Need fair ordering protocols

**3. Single Secret Leader Election**
- Hide next proposer identity
- Prevent targeted DoS attacks

**4. Accountable safety**
- Prove who caused consensus failure
- Better than generic slashing

---

## Summary

**Proof of Work:**
- ✅ Simple, proven, decentralized
- ❌ Energy-intensive, slow finality

**Proof of Stake:**
- ✅ Energy-efficient, faster finality
- ❌ More complex, shorter track record

**No free lunch:** All consensus mechanisms trade off security, scalability, and decentralization.

**Practical advice:** 
- High-value, store of value → PoW (Bitcoin)
- Smart contract platform, apps → PoS (Ethereum)
- Enterprise/permissioned → PoA

---

## Exercise

Build a consensus simulator:

```typescript
interface ConsensusMetrics {
  throughput: number; // tx/sec
  finality: number; // seconds
  energyCost: number; // kWh per tx
  decentralization: number; // 0-100
}

class PoWConsensus {
  getMetrics(): ConsensusMetrics {
    return {
      throughput: 7,
      finality: 3600, // 6 confirmations
      energyCost: 700, // ~700 kWh per Bitcoin tx
      decentralization: 70 // Moderate (pools exist)
    };
  }
}

class PoSConsensus {
  getMetrics(): ConsensusMetrics {
    return {
      throughput: 15,
      finality: 780, // ~13 minutes
      energyCost: 0.01, // 99.95% reduction
      decentralization: 60 // Slightly less (staking pools)
    };
  }
}

function compareConsensus(): void {
  const pow = new PoWConsensus();
  const pos = new PoSConsensus();
  
  console.log('Proof of Work:', pow.getMetrics());
  console.log('Proof of Stake:', pos.getMetrics());
  
  // Calculate energy savings
  const energySavings = 
    ((pow.getMetrics().energyCost - pos.getMetrics().energyCost) / 
     pow.getMetrics().energyCost) * 100;
  
  console.log(`Energy savings: ${energySavings.toFixed(2)}%`);
}

compareConsensus();
```

**Questions:**
1. How would you model DPoS (Delegated Proof of Stake)?
2. What metrics matter most for your use case?
3. Can you predict centralization trends over time?

---

## Next Lesson

[→ Finality and Reorgs](05-finality.md)
