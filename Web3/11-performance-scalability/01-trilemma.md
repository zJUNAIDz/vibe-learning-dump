# The Scalability Trilemma

🟢 **Fundamentals**

---

## The Core Problem

**Claim:** You can only have 2 of 3:

```
        Security
           /\
          /  \
         /    \
        /      \
       /________\
  Scalability  Decentralization
```

**Choose 2, sacrifice 1.**

---

## Defining the Terms

### Security
Can the system resist attacks?

**Measured by:**
- Cost to attack (51% attack)
- History of successful attacks
- Economic incentives

---

### Decentralization
How many independent parties control the system?

**Measured by:**
- Number of nodes
- Geographic distribution
- Validator/miner concentration
- Client diversity

---

### Scalability
How many transactions per second (TPS)?

**Measured by:**
- Throughput (TPS)
- Latency (time to finality)
- Cost per transaction

---

## Why You Can't Have All Three

### The Fundamental Constraint

**Every node processes every transaction.**

```
Transaction → Node 1 ✓
           → Node 2 ✓
           → Node 3 ✓
           → Node 4 ✓
           → Node N ✓

All nodes must:
1. Receive transaction
2. Validate it
3. Execute it
4. Store the result
```

**Consequence:** Network speed = slowest node's speed.

---

### The Math

**If:**
- 10,000 nodes (decentralized)
- Each can handle 1,000 TPS
- Every node processes every transaction

**Then:**
- Network throughput: ~10-15 TPS (bottlenecked by propagation)

**Why not 1,000 TPS?**
- Network latency (nodes worldwide)
- Storage growth (every node stores full history)
- Consensus overhead (coordinating 10,000 nodes)

---

## Real-World Examples

### Bitcoin: Security + Decentralization → Sacrifices Scalability

**Specs:**
- **TPS:** ~7
- **Block time:** 10 minutes
- **Node count:** ~15,000

**Tradeoffs:**
- ✅ Very secure ($10B+ to attack)
- ✅ Very decentralized (anyone can run node)
- ❌ Very slow (7 TPS)

**Cost:** Users wait hours, pay high fees during congestion.

---

### Ethereum: Security + Moderate Decentralization → Limited Scalability

**Specs:**
- **TPS:** ~15
- **Block time:** 12 seconds
- **Node count:** ~8,000

**Tradeoffs:**
- ✅ Secure (billions locked in DeFi proves it)
- ✅ Reasonably decentralized
- ❌ Expensive ($1-50 per transaction during spikes)

---

### Solana: Scalability + Security → Sacrifices Decentralization

**Specs:**
- **TPS:** ~3,000 (claimed)
- **Block time:** 400ms
- **Node count:** ~3,000 (but high hardware requirements)

**Hardware requirements:**
- 12-core CPU
- 256 GB RAM
- High-bandwidth internet

**Tradeoffs:**
- ✅ Fast (sub-second finality)
- ✅ Cheap ($0.00025 per transaction)
- ❌ Centralization (few can afford hardware)
- ❌ Network outages (downtime in 2021, 2022)

---

### Binance Smart Chain (BSC): Scalability + Moderate Decentralization → Security Questions

**Specs:**
- **TPS:** ~160
- **Block time:** 3 seconds
- **Validators:** 21 (very centralized)

**Tradeoffs:**
- ✅ Fast and cheap
- ❌ Only 21 validators (controlled by Binance ecosystem)
- ❌ Multiple hacks (Ronin Bridge: $600M, BNB Bridge: $100M)

---

## Breaking Down the Tradeoffs

### Increasing Scalability → Decreases Decentralization

**Method 1: Bigger blocks**
```
Bigger blocks → More data per block
             → Harder to sync
             → Fewer people can run nodes
             → Centralization
```

**Bitcoin block size war (2017):** Keep 1 MB blocks or increase?
- Increase → Faster TPS → Fewer nodes
- Community split: Bitcoin vs Bitcoin Cash

---

**Method 2: Faster blocks**
```
Faster blocks → Less time for propagation
             → More orphaned blocks
             → Advantage to well-connected nodes
             → Centralization toward data centers
```

---

**Method 3: Higher requirements**
```
High CPU/RAM/bandwidth → Expensive to run node
                       → Only data centers can participate
                       → Centralization
```

This is Solana's approach.

---

### Increasing Decentralization → Decreases Scalability

**More nodes:**
```
More nodes → More propagation delay
          → Must wait for slowest node
          → Lower TPS
```

**More geographic distribution:**
```
Global nodes → Higher latency
            → Slower consensus
            → Lower TPS
```

---

## Proposed Solutions (Spoiler: They Still Make Tradeoffs)

### Layer 2 Rollups

**Idea:** Do computation off-chain, post proofs on-chain.

```
1000 transactions off-chain
    ↓
Process in rollup
    ↓
Post 1 proof on-chain (1 transaction)
```

**Effective TPS:** 1000x improvement

**Tradeoff:**
- ✅ Inherits Ethereum security
- ❌ Adds complexity (bridging, withdrawals)
- ❌ Fragmented liquidity (each rollup is separate)

---

### Sharding

**Idea:** Split blockchain into parallel chains (shards).

```
Shard 1: Transactions 1-1000
Shard 2: Transactions 1001-2000
Shard 3: Transactions 2001-3000
```

**Effective TPS:** N shards = N × base TPS

**Tradeoff:**
- ✅ Parallelization
- ❌ Cross-shard communication is slow
- ❌ Security per shard is lower (need 51% of 1 shard, not whole network)

---

### Alternative Consensus (PoS vs PoW)

**PoS advantages:**
- Faster finality (no need to wait for many blocks)
- Less energy waste

**But doesn't solve trilemma:**
- Still need all validators to process all transactions
- Faster consensus helps, but not a 100x improvement

---

### Modular Blockchains

**Idea:** Separate layers for different functions.

```
Execution Layer (Rollups)
      ↓
Consensus Layer (Ethereum)
      ↓
Data Availability Layer (Celestia)
```

**Each layer optimized differently.**

**Tradeoff:**
- ✅ Specialization improves each component
- ❌ More complex architecture
- ❌ More trust assumptions

---

## Why Traditional Databases Don't Have This Problem

**Centralized database:**
- 1 server (or small cluster)
- No consensus needed
- Can do millions of TPS

**Blockchain:**
- Thousands of servers
- Must reach consensus
- Each node processes everything

**The entire point of blockchain is decentralization. That's why it's slow.**

---

## Comparing Performance

| System | TPS | Latency | Decentralization | Cost/tx |
|--------|-----|---------|------------------|---------|
| **Visa** | 24,000 | Seconds | Centralized | $0.03 |
| **PostgreSQL** | 10,000+ | Milliseconds | Centralized (1 server) | Free |
| **Bitcoin** | 7 | 60 minutes | Very high | $1-50 |
| **Ethereum** | 15 | 13 minutes | High | $1-100 |
| **Solana** | 3,000 | 1 second | Medium | $0.00025 |
| **BSC** | 160 | 3 seconds | Low (21 validators) | $0.10 |

**Observation:** Decentralization and performance are inversely correlated.

---

## Don't Trust TPS Claims

**Marketing trick:** Quote theoretical max, ignore real-world usage.

```
Claimed: 100,000 TPS!
Reality: 
- For simple transfers only
- At 100% usage (never happens)
- Ignoring consensus overhead
- Assuming perfect network conditions
```

**Always check:**
1. Actual sustained TPS in production
2. Cost per transaction at that TPS
3. Number of validators/nodes
4. Hardware requirements

---

## Practical Implications for Developers

### Don't Build Like It's a Database

❌ **Don't:** Put every action on-chain  
✅ **Do:** Only critical state on-chain

❌ **Don't:** Query blockchain directly for listings  
✅ **Do:** Index events into your database

❌ **Don't:** Expect sub-second finality on L1  
✅ **Do:** Design for 12+ second delays

---

### Choose Your Blockchain Based on Needs

**Need maximum security and decentralization?**
→ Bitcoin or Ethereum L1

**Need cheaper transactions, okay with some centralization?**
→ L2 rollups, Polygon, BSC

**Building game with high throughput, low value?**
→ Solana or application-specific chain

**There's no "best" blockchain. Only tradeoffs.**

---

## Summary

**Scalability Trilemma:**
- Can't have Security + Decentralization + Scalability
- Pick 2, sacrifice 1

**Why it exists:**
- Decentralization = many nodes process everything
- Scalability = high throughput
- These goals conflict

**Solutions (all make tradeoffs):**
- **L2 Rollups:** Off-chain computation, on-chain proofs
- **Sharding:** Parallel chains, cross-shard complexity
- **Faster consensus:** PoS helps, but not 100x
- **Centralization:** Simply use fewer nodes (many chains take this shortcut)

**Key insight:** Blockchain is slow by design. Decentralization is the feature, slowness is the cost.

**As a developer:** Choose the right tool for the job. Don't use blockchain for problems that need databases.

---

**[Next Lesson →](02-layer-2.md)**
