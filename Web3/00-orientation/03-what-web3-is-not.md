# What Web3 Is NOT

🟢 **Fundamentals**

---

## Debunking the Hype

Web3 has been marketed with grandiose claims. Let's address them honestly.

---

## ❌ Web3 Is NOT "The Future of the Internet"

### The Claim
"Web3 will replace Web2. Everything will be decentralized."

### The Reality
- Most "Web3 apps" still run on AWS
- Frontends are centralized (hosted on Vercel, Netlify, etc.)
- RPC nodes are often centralized (Infura, Alchemy)
- Users still interact via traditional browsers

**Web3 complements Web2, it doesn't replace it.**

**Why?**
- Decentralization is expensive
- Most apps don't need it
- Users care about UX, not decentralization

---

## ❌ Web3 Is NOT Faster Than Traditional Systems

### The Claim
"Blockchain transactions settle instantly."

### The Reality
| System | Transaction Time |
|--------|------------------|
| Credit card | ~1-3 seconds (appears instant) |
| Database write | Milliseconds |
| Ethereum | 12-15 seconds (1 block) |
| Bitcoin | ~10 minutes (1 block) |
| "Finality" on Ethereum | ~15 minutes (multiple blocks) |

**Web3 is intentionally slow** to ensure consensus.

**Why?**
- All nodes must agree
- Consensus mechanisms (PoW, PoS) take time
- Security requires multiple confirmations

---

## ❌ Web3 Is NOT Private

### The Claim
"Web3 gives you control over your data."

### The Reality
- **Every transaction is public**
- Your wallet balance is public
- Your transaction history is public
- Your smart contract interactions are public

**Pseudonymity ≠ Anonymity**

Your address (`0x1234...`) doesn't show your name, but:
- Addresses can be linked to identities
- Transaction patterns are traceable
- Exchanges (Coinbase, Binance) know your identity

**Example:**
If Alice sends ETH from Coinbase to address `0xABC...`, then `0xABC...` sends ETH to a controversial organization, Alice can be traced.

**Privacy options exist (ZK-proofs, Tornado Cash), but:**
- They're complex
- Often targeted by regulators
- Not default behavior

---

## ❌ Web3 Is NOT Cheap

### The Claim
"Web3 eliminates middlemen, so it's cheaper."

### The Reality
Transaction fees (gas) can be:
- **Ethereum L1:** $1-$50+ per transaction (during congestion: $100+)
- **Polygon:** $0.01-$0.50
- **Optimistic Rollups (L2s):** $0.10-$2

**Example costs:**
- Deploying a smart contract: $50-$500+
- Minting an NFT: $10-$100+
- Swapping tokens on Uniswap: $5-$50+

Compare to:
- Database write: ~$0.0000001
- API call: ~$0.00001
- Credit card transaction fee: 2-3% (predictable)

**Why so expensive?**
- Every node processes every transaction
- Miners/validators need compensation
- Demand exceeds supply during congestion

---

## ❌ Web3 Is NOT Truly Decentralized (In Practice)

### The Claim
"No one controls Web3."

### The Reality
Centralization vectors:
- **Mining/Validator pools:** 3-5 entities control most hash power (Bitcoin) or stake (Ethereum)
- **RPC providers:** Infura and Alchemy handle most Ethereum requests
- **Frontends:** Hosted on centralized servers
- **Development teams:** Small groups control protocol upgrades
- **Wealth distribution:** Whales and VCs dominate governance

**Example: Infura Outage (2020)**
Infura (a centralized RPC provider) went down. Most "decentralized" apps stopped working.

**Key insight:** The protocol may be decentralized, but the infrastructure around it often isn't.

---

## ❌ Web3 Is NOT Secure by Default

### The Claim
"Smart contracts execute exactly as programmed. No bugs."

### The Reality
Smart contracts have bugs, and they're exploited constantly.

**Notable hacks:**
- **The DAO (2016):** $60M stolen (reentrancy attack)
- **Parity Wallet (2017):** $150M accidentally frozen
- **Poly Network (2021):** $600M stolen
- **Ronin Bridge (2022):** $625M stolen
- **Euler Finance (2023):** $197M stolen

**Why?**
- Smart contracts are immutable (bugs are permanent)
- Audits help, but aren't foolproof
- New attack vectors emerge constantly
- Economic incentives attract attackers

---

## ❌ Web3 Is NOT Uncensorable (Always)

### The Claim
"No one can censor Web3 transactions."

### The Reality
Censorship can occur at multiple levels:

1. **Validator/miner level:**
   - Miners can refuse to include transactions
   - Post-merge Ethereum: validators can be pressured to comply with sanctions

2. **RPC provider level:**
   - Infura/Alchemy can block certain addresses
   - Frontends can hide certain transactions

3. **Smart contract level:**
   - Contracts can have admin keys (centralization)
   - Upgradeable contracts can freeze accounts

4. **Social layer:**
   - Protocol forks can reverse transactions (Ethereum Classic split)

**Example: Tornado Cash (2022)**
- US Treasury sanctioned Tornado Cash
- Infura/Alchemy blocked access
- GitHub removed the code repository
- Developers were arrested

**Key insight:** Technical censorship resistance ≠ real-world censorship resistance.

---

## ❌ Web3 Is NOT Environmentally Friendly (Historically)

### The Claim (2017-2022)
"Bitcoin/Ethereum are just computers running code."

### The Reality (Proof of Work era)
- **Bitcoin:** ~150 TWh/year (comparable to Argentina)
- **Ethereum (pre-merge):** ~100 TWh/year

**Post-merge (Ethereum switched to Proof of Stake, 2022):**
- Energy usage dropped ~99.95%
- Bitcoin still uses Proof of Work

**Key insight:** Proof of Work is energy-intensive by design. Proof of Stake is better, but still not free.

---

## ❌ Web3 Is NOT Easy to Use

### The Claim
"Web3 empowers users."

### The Reality
**Onboarding a new user requires:**
1. Explaining what a blockchain is
2. Explaining what a wallet is
3. Installing MetaMask (or similar)
4. Writing down a 12-24 word seed phrase (lose it = lose everything)
5. Buying cryptocurrency (KYC on exchange)
6. Transferring crypto to wallet
7. Understanding gas fees
8. Approving token contracts
9. Signing transactions

**Compare to Web2:**
1. Click "Sign up with Google"
2. Done.

**Common user errors:**
- Sending tokens to the wrong address (lost forever)
- Losing seed phrase (wallet locked forever)
- Approving malicious contracts (funds stolen)
- Paying too little gas (transaction fails, fee still charged)
- Falling for phishing sites (send money to scammer)

**Web3 UX is hostile by design** (self-custody = self-responsibility).

---

## ❌ Web3 Is NOT Immune to Bugs

### The Claim
"Code is law. Smart contracts are deterministic."

### The Reality
- Code can have bugs
- Protocols can have vulnerabilities
- Attacks happen constantly

**Categories of bugs:**
1. **Smart contract bugs:**
   - Reentrancy, overflow, access control flaws

2. **Protocol vulnerabilities:**
   - Consensus failures, 51% attacks

3. **Bridge exploits:**
   - Cross-chain bridges are notoriously vulnerable

4. **Oracle manipulation:**
   - Smart contracts relying on bad data

**Key insight:** "Code is law" means bugs are also law.

---

## ❌ Web3 Does NOT Eliminate Scams

### The Claim
"Blockchain transparency prevents fraud."

### The Reality
Web3 is full of scams:
- **Rug pulls:** Developers abandon project, steal funds
- **Ponzi schemes:** "Yield farming" that collapses
- **Phishing:** Fake websites steal private keys
- **Fake tokens:** Scammers create tokens that look real
- **Pump and dumps:** Coordinated price manipulation

**Why?**
- Irreversibility means stolen funds can't be recovered
- Pseudonymity makes scammers hard to track
- Lack of regulation = no consumer protection

---

## What Web3 Is NOT: Summary

| Claim | Reality |
|-------|---------|
| "Future of the internet" | Complements Web2, doesn't replace it |
| "Fast" | Intentionally slow (consensus takes time) |
| "Private" | Public by default (pseudonymous, not anonymous) |
| "Cheap" | Expensive (gas fees) |
| "Decentralized" | Often centralized in practice |
| "Secure" | Bugs and hacks are common |
| "Uncensorable" | Censorship is possible |
| "Environmentally friendly" | PoW is energy-intensive, PoS is better |
| "Easy to use" | UX is hostile and error-prone |
| "Scam-proof" | Full of scams |

---

## Why These Myths Persist

1. **Financial incentives:**
   - Early adopters profit from hype
   - VCs need user adoption for returns

2. **Ideological motivation:**
   - Genuine belief in decentralization
   - Libertarian/anti-establishment values

3. **Marketing:**
   - "Revolutionary" sells better than "useful for niche cases"

4. **Lack of technical literacy:**
   - Non-technical people repeat claims uncritically

---

## Does This Mean Web3 Is Useless?

**No.**

Web3 has real use cases (covered in future modules). But understanding its limitations is critical.

**Web3 is useful when:**
- Decentralization matters more than performance
- Transparency matters more than privacy
- Censorship resistance justifies the cost
- No trusted intermediary exists

**Web3 is NOT useful when:**
- You need speed, privacy, or low cost
- A database would suffice
- User experience matters

---

## Exercise

Identify the false claim and explain why:

1. "Ethereum transactions are instant and free."
   - **False.** Transactions take 12+ seconds and cost $1-$50+.

2. "All Web3 apps are fully decentralized."
   - **False.** Most rely on centralized frontends, RPC nodes, and infrastructure.

3. "Blockchains are public, so your transactions are visible."
   - **True.** This is a real limitation.

4. "Smart contracts never have bugs because code is law."
   - **False.** Bugs are common and permanent.

5. "Web3 eliminates the need for trust."
   - **False.** Trust is distributed, not eliminated.

---

## Next Lesson

[→ The Tradeoffs](04-tradeoffs.md)
