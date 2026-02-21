# What Problem Web3 Tries to Solve

🟢 **Fundamentals**

---

## The Coordination Problem

Imagine you want to build a system where:
- Multiple parties need to agree on shared state
- No single party should control the system
- Participants may not trust each other
- The rules should be transparent and enforceable

**Traditional solution:** Use a trusted intermediary.

**Web3 solution:** Use a blockchain.

---

## Example: A Simple Ledger

### Scenario: Tracking IOUs Among Friends

You and three friends want to track who owes whom money.

#### Option 1: Alice Keeps the Ledger (Centralized)
```
Alice's Notebook:
- Bob owes Carol $20
- Dave owes Alice $15
- Carol owes Bob $10
```

**Problems:**
- What if Alice loses the notebook?
- What if Alice lies about the entries?
- What if Alice is unavailable?
- Everyone must trust Alice

#### Option 2: Everyone Keeps a Copy (Naive Replication)
```
Alice's Copy:    Bob's Copy:     Carol's Copy:   Dave's Copy:
Bob → Carol: $20  Bob → Carol: $20  Bob → Carol: $20  Bob → Carol: $20
Dave → Alice: $15 Dave → Alice: $15 Dave → Alice: $15 Dave → Alice: $15
```

**Problems:**
- What if copies diverge?
- How do you add new entries?
- What if someone lies about their copy?
- Who resolves conflicts?

#### Option 3: Blockchain Approach
- Everyone has a copy
- New entries are proposed as "transactions"
- Everyone must agree on the order of transactions
- Once agreed, entries are permanent
- Anyone can verify the history

This is what blockchains attempt to solve.

---

## The Trust Problem

### Traditional Systems Require Trust

In traditional systems, you trust:
- Banks (to manage your money)
- Governments (to enforce rules)
- Companies (to store your data)
- Platforms (to not censor you)

This works fine **when the intermediary is trustworthy**.

### When Trust Fails

Trust breaks down when:
- Intermediaries can be censored or coerced
- Intermediaries have conflicting incentives
- Intermediaries can be hacked or fail
- Parties are in different jurisdictions
- No neutral third party exists

---

## Web3's Core Promise

**"Don't trust, verify."**

Instead of trusting an intermediary:
- Trust math and cryptography
- Trust transparent, auditable code
- Trust economic incentives
- Trust consensus mechanisms

**BUT:** You still have to trust something (more on this later).

---

## The Double-Spend Problem

A core problem blockchains solve: **preventing double-spending in digital assets**.

### The Problem with Digital Money

Digital files can be copied infinitely:
```
Alice has coin.txt
Alice sends coin.txt to Bob
Alice still has coin.txt!
Alice sends coin.txt to Carol (double-spend)
```

### Traditional Solution: Trusted Ledger
Banks prevent double-spending by maintaining a centralized ledger:
```
Bank ledger:
Alice: $100
Alice sends $50 to Bob → Bank verifies Alice has $50
Bank updates ledger:
Alice: $50, Bob: $50
```

**Problem:** You must trust the bank.

### Blockchain Solution: Public Ledger + Consensus
- Everyone has a copy of the ledger
- Transactions are ordered via consensus
- Once a transaction is confirmed, it's permanent
- Anyone can verify no double-spending occurred

---

## What Web3 Claims to Enable

1. **Censorship resistance**
   - No single entity can block transactions
   - Example: Sending money to controversial causes

2. **Permissionless participation**
   - Anyone can join without approval
   - Example: Creating a financial service without a banking license

3. **Transparent execution**
   - Rules are public and verifiable
   - Example: Smart contracts execute deterministically

4. **No single point of failure**
   - System continues even if some nodes fail
   - Example: No "server downtime"

5. **Programmable trust**
   - Encode agreements in code
   - Example: Automatic payments based on conditions

---

## Real-World Scenarios

### Scenario 1: International Remittances
**Traditional:** Alice in the US sends money to Bob in Kenya.
- Bank fees: $20-50
- Time: 3-7 days
- Requires bank accounts
- Subject to restrictions

**Web3:** Alice sends cryptocurrency to Bob.
- Network fees: $0.01-5 (depends on network)
- Time: Minutes
- No bank account needed
- No intermediary approval

**Tradeoff:** Bob needs to convert crypto to local currency (not always easy).

### Scenario 2: Crowdfunding for Censored Cause
**Traditional:** Platform (e.g., GoFundMe) can freeze or block campaigns.

**Web3:** Smart contract holds funds, releases them based on code logic. No platform can intervene.

**Tradeoff:** If the smart contract has a bug, funds can be lost permanently.

### Scenario 3: Supply Chain Tracking
**Traditional:** Each party maintains their own records. Hard to verify authenticity.

**Web3:** All parties write to a shared blockchain. Transparent and auditable.

**Tradeoff:** Privacy is reduced. Competitors can see each other's data.

---

## The Core Insight

**Blockchains replace trusted intermediaries with cryptography and consensus.**

This is useful when:
- No trusted intermediary exists
- Intermediaries are corruptible or censorable
- Coordination across adversarial parties is needed
- Transparency is more important than privacy

This is NOT useful when:
- A trusted intermediary exists and works fine
- Privacy is critical
- Performance matters
- Flexibility is needed

---

## Questions to Ask Yourself

Before using Web3, ask:
1. **Who am I trying to not trust?**
   - If the answer is "no one," you don't need Web3.

2. **What am I giving up?**
   - Performance? Privacy? Flexibility?

3. **Could a database + authentication solve this?**
   - If yes, use a database.

4. **Is transparency worth the cost?**
   - Every transaction costs money and is public.

---

## Summary

**Web3 tries to solve:**
- Coordination without trusted intermediaries
- Double-spending in digital assets
- Censorship and single points of failure

**Web3 introduces:**
- Slower performance
- Higher costs
- Less privacy
- Immutable mistakes

**Web3 is useful when:**
- Trust is impossible or expensive
- Censorship resistance matters
- Transparent rule execution is critical

**Web3 is NOT useful when:**
- Performance, privacy, or flexibility matter more
- A trusted intermediary works fine

---

## Exercise

For each scenario, decide: **Should you use Web3? Why or why not?**

1. A company wants to track employee expenses internally.
2. A journalist wants to receive anonymous donations from a hostile government.
3. A multiplayer game wants to track in-game item ownership.
4. A hospital wants to store patient medical records.
5. A group of strangers wants to pool money for a shared investment.

**Answers:**
1. **No.** Internal system. Database with auth is simpler.
2. **Maybe.** Censorship resistance matters, but usability is poor.
3. **Maybe.** Depends on whether "true ownership" matters to players.
4. **No.** Privacy is critical. Blockchains are public.
5. **Yes.** No trusted party. Smart contract can enforce rules.

---

## Next Lesson

[→ What Web3 Actually Is](02-what-web3-actually-is.md)
