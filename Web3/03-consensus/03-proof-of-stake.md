# Proof of Stake (PoS)

🟡 **Intermediate**

---

## The Big Idea

**Replace computation with capital.**

Instead of miners spending electricity, validators lock up cryptocurrency as collateral.

**Security model:**
- Attacking requires controlling majority of staked capital
- Attackers lose their stake (slashing)
- Attacking costs more than potential profit

---

## How It Works

### 1. Staking
Validators lock up capital (e.g., 32 ETH on Ethereum).

```typescript
interface Validator {
  address: string;
  stakedAmount: number; // ETH
  publicKey: string;
}

function becomeValidator(amount: number): Validator {
  if (amount < 32) {
    throw new Error('Minimum stake: 32 ETH');
  }
  
  // Lock funds in staking contract
  return {
    address: generateAddress(),
    stakedAmount: amount,
    publicKey: generatePublicKey()
  };
}
```

### 2. Random Selection
Protocol randomly selects validators to propose blocks.

**Key difference from PoW:** No race. One validator chosen per slot.

```typescript
function selectValidator(
  validators: Validator[],
  randomSeed: string,
  slot: number
): Validator {
  // Pseudo-random selection weighted by stake
  const totalStake = validators.reduce((sum, v) => sum + v.stakedAmount, 0);
  const random = hashToNumber(randomSeed + slot) % totalStake;
  
  let cumulative = 0;
  for (const validator of validators) {
    cumulative += validator.stakedAmount;
    if (random < cumulative) {
      return validator;
    }
  }
  
  return validators[0]; // Shouldn't reach here
}

function hashToNumber(input: string): number {
  const hash = sha256(input);
  return parseInt(hash.slice(0, 8), 16);
}
```

### 3. Attestation
Other validators attest (vote) that block is valid.

```typescript
interface Attestation {
  validatorIndex: number;
  blockHash: string;
  signature: string; // Sign with validator's private key
}

function attest(validator: Validator, block: Block): Attestation {
  return {
    validatorIndex: validator.address,
    blockHash: block.hash,
    signature: sign(block.hash, validator.privateKey)
  };
}
```

### 4. Finality
Block becomes final when 2/3 of validators attest.

---

## Slashing

**Problem:** What if validators misbehave?

**Solution:** Destroy part of their stake.

### Slashable Offenses

**1. Double-signing:** Proposing two different blocks for same slot
```typescript
function detectDoubleSign(
  block1: Block,
  block2: Block,
  validator: Validator
): boolean {
  return (
    block1.slot === block2.slot &&
    block1.hash !== block2.hash &&
    block1.proposer === validator.address &&
    block2.proposer === validator.address
  );
}
```

**2. Surround voting:** Attesting to conflicting checkpoints
```typescript
// Validator votes for:
//   Block A at epoch 10 → Block C at epoch 20
// Then votes for:
//   Block B at epoch 12 → Block D at epoch 18
// "Surrounds" previous vote
```

**Penalty:** Validator loses 0.5 - 100% of stake (depends on severity and how many validators slashed simultaneously).

---

## Nothing at Stake Problem

**Theory:** In PoS, validators can vote for multiple forks simultaneously (no cost like electricity in PoW).

```
        Block 100A
       /
Block 99 
       \
        Block 100B
```

Validator votes for **both** to maximize rewards.

**Solution:** Slashing.
- Vote for multiple forks → detected → lose stake
- One honest vote expected per slot

---

## Long-Range Attacks

**Attack:** Rewrite history from very old block.

In PoW: Impossible (need to redo all computational work).

In PoS: Attacker could buy keys of validators who staked long ago (and already withdrew).

**Example:**
1. Block 1,000,000 is current
2. Attacker buys old validator keys from block 1,000
3. Attacker creates alternative chain from block 1,000
4. New nodes see both chains, can't tell which is real

**Solution: Weak Subjectivity**
- New nodes ask trusted peers for recent checkpoint
- Only accept chains that include that checkpoint
- Assumption: Node must sync within ~4 months (Ethereum)

**Tradeoff:** No longer purely objective like PoW.

---

## Economic Security

### Attack Cost (Ethereum)

To control network (>50% stake):
- Total staked: ~30 million ETH
- Need: >15 million ETH
- At ETH = $2,000: **$30 billion**

Plus penalty: **Stake gets slashed**. Unlike PoW (can sell mining hardware), PoS attackers lose capital.

### Validator Revenue

**Sources:**
- Block rewards: New issuance (~0.5% APR)
- Transaction fees: Variable
- MEV (Maximal Extractable Value): Transaction ordering profits

**Total APR:** ~3-5% (Ethereum, varies)

---

## Centralization Risks

### 1. Minimum Stake Requirements
Ethereum: 32 ETH (~$64,000 at $2,000/ETH).

**Problem:** Most users can't afford solo staking.

**Solution:** Staking pools (e.g., Lido, Rocket Pool).

**New problem:** Pools centralize control.

### 2. Rich Get Richer
Validators earn rewards on stake. More stake → more rewards → more stake.

**Counter-argument:** Same as PoW (big miners buy more hardware).

### 3. Large Exchanges
Coinbase, Binance hold millions in customer ETH, can stake it.

**Risk:** A few entities control large percentage of stake.

---

## Ethereum's Transition (The Merge)

**September 2022:** Ethereum switched from PoW to PoS.

### Why?

**1. Energy:** PoW used ~110 TWh/year. PoS uses ~0.01 TWh/year (99.95% reduction).

**2. Issuance:** Lower reward needed (no electricity cost). Reduced inflation.

**3. Security:** More expensive to attack (can't rent hash power temporarily).

### How?

Not a simple upgrade. Took years:
- **Beacon Chain** (Dec 2020): New PoS chain ran in parallel
- **The Merge** (Sep 2022): Ethereum execution layer "merged" with Beacon Chain
- Entire history preserved, no data lost

---

## Variants of PoS

### Delegated Proof of Stake (DPoS)
Users vote for validators. Top N validators produce blocks.

**Examples:** EOS, Tron

**Tradeoff:** Faster, but more centralized (often 21-100 validators).

### Bonded Proof of Stake
Validators must lock stake for fixed period.

**Examples:** Cosmos, Polkadot

**Tradeoff:** Unbonding period (weeks) to prevent long-range attacks.

### Liquid Staking
Users stake via protocol, get tradeable token (e.g., stETH from Lido).

**Benefit:** Staking rewards + liquidity.

**Risk:** Derivatives introduce new systemic risks (see Terra/Luna collapse).

---

## Summary

**Proof of Stake:**
- Replace energy with locked capital
- Validators chosen pseudo-randomly (weighted by stake)
- Misbehavior punished via slashing
- 2/3 attestations → finality

**Tradeoffs:**
- ✅ **Energy efficient:** 99.95% less than PoW
- ✅ **Faster finality:** Minutes vs hours (PoW)
- ✅ **Lower issuance:** Don't need to pay for electricity
- ❌ **More complex:** Slashing conditions, weak subjectivity
- ❌ **Rich get richer:** Rewards proportional to stake
- ❌ **Unproven long-term:** PoW battle-tested since 2009, PoS at scale since 2022

**Key insight:** Security doesn't come from energy anymore, it comes from locked capital at risk.

---

## Exercise

Implement simplified validator selection:

```typescript
import crypto from 'crypto';

interface Validator {
  id: string;
  stake: number;
}

class ProofOfStake {
  validators: Validator[] = [];

  addValidator(id: string, stake: number): void {
    this.validators.push({ id, stake });
  }

  selectProposer(slot: number): Validator {
    const totalStake = this.validators.reduce((sum, v) => sum + v.stake, 0);
    
    // Use slot as randomness source (simplified)
    const seed = crypto.createHash('sha256')
      .update(slot.toString())
      .digest('hex');
    
    const random = parseInt(seed.slice(0, 8), 16) % totalStake;
    
    let cumulative = 0;
    for (const validator of this.validators) {
      cumulative += validator.stake;
      if (random < cumulative) {
        return validator;
      }
    }
    
    return this.validators[0];
  }

  simulateElection(slots: number): Record<string, number> {
    const selections: Record<string, number> = {};
    
    for (let slot = 0; slot < slots; slot++) {
      const proposer = this.selectProposer(slot);
      selections[proposer.id] = (selections[proposer.id] || 0) + 1;
    }
    
    return selections;
  }
}

// Test
const pos = new ProofOfStake();
pos.addValidator('Alice', 32);
pos.addValidator('Bob', 64);
pos.addValidator('Charlie', 32);

const results = pos.simulateElection(1000);

console.log('Selections over 1000 slots:');
for (const [id, count] of Object.entries(results)) {
  const validator = pos.validators.find(v => v.id === id)!;
  const expected = (validator.stake / 128) * 1000;
  console.log(`${id}: ${count} (expected: ${expected.toFixed(0)})`);
}
```

**Questions:**
1. Is selection proportional to stake?
2. What happens if one validator has 51% stake?
3. How would you implement slashing?

---

## Next Lesson

[→ Comparing Consensus Mechanisms](04-comparing-consensus.md)
