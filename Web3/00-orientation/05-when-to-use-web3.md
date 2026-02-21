# When to Use Web3 (And When Not To)

🟢 **Fundamentals**

---

## Decision Framework

Before using Web3, answer these questions honestly:

### Question 1: Do you need public verifiability?
- Can anyone verify the state without trusting you?
- Example: Transparent voting, provenance tracking

**If NO** → You probably don't need Web3.

### Question 2: Is there a trusted intermediary unavailable?
- Can you trust a bank, government, or platform?
- Do participants distrust each other?

**If NO** → Use a traditional database with authentication.

### Question 3: Is censorship resistance critical?
- Will powerful entities try to block this?
- Do you need permissionless access?

**If NO** → Web2 is simpler.

### Question 4: Can users tolerate the tradeoffs?
- Slow transactions (seconds to minutes)
- High fees ($1-$100+ per transaction)
- Poor UX (wallets, seed phrases, gas fees)
- No customer support

**If NO** → Web3 will frustrate users.

---

## Valid Use Cases

### ✅ 1. Decentralized Finance (DeFi)

**Problem:** Banks are gatekeepers for financial services.

**Web3 solution:**
- Anyone can borrow, lend, or trade without permission
- Transparent rules (smart contracts)
- No KYC required (for better or worse)

**Examples:**
- **Uniswap:** Decentralized token exchange
- **Aave:** Decentralized lending protocol
- **MakerDAO:** Decentralized stablecoin (DAI)

**Tradeoffs accepted:**
- Higher fees than centralized exchanges
- Risk of smart contract bugs
- No customer support if you make a mistake

**When it makes sense:**
- Users are excluded from traditional finance
- Transparency matters more than convenience
- Users are willing to manage their own keys

---

### ✅ 2. Digital Ownership & Provenance

**Problem:** Proving authenticity and ownership of digital goods is hard.

**Web3 solution:**
- NFTs (Non-Fungible Tokens) represent unique assets
- Ownership is public and verifiable
- Transfer history is transparent

**Examples:**
- **Digital art:** Provenance tracking
- **Event tickets:** Prevent counterfeiting
- **Domain names:** Decentralized DNS (ENS)

**Tradeoffs accepted:**
- Art metadata often stored off-chain (centralized)
- NFTs don't prevent copying (only prove ownership)
- High minting costs

**When it makes sense:**
- Provenance matters (collectibles, art)
- Resale royalties need enforcement
- Digital scarcity has value

---

### ✅ 3. Crowdfunding & DAOs (Decentralized Autonomous Organizations)

**Problem:** Crowdfunding platforms can freeze campaigns or take high fees.

**Web3 solution:**
- Smart contracts hold funds
- Rules are transparent and enforceable
- No platform can censor

**Examples:**
- **Gitcoin:** Quadratic funding for public goods
- **MolochDAO:** Coordinating Ethereum development funding
- **Constitution DAO:** Attempt to buy US Constitution (failed, but transparent)

**Tradeoffs accepted:**
- If smart contract has bug, funds can be lost
- Governance is often plutocratic (whales control votes)
- Slow decision-making

**When it makes sense:**
- Censorship is a risk
- Transparency in fund management matters
- Participants don't trust a central organizer

---

### ✅ 4. Cross-Border Payments

**Problem:** International remittances are slow and expensive.

**Web3 solution:**
- Send cryptocurrency globally in minutes
- Lower fees than SWIFT/Western Union
- No bank account required

**Examples:**
- **Stablecoins (USDC, USDT):** Dollar-pegged tokens for remittances
- **Bitcoin:** Peer-to-peer payments

**Tradeoffs accepted:**
- Recipient must convert crypto to local currency (not always easy)
- Price volatility (if not using stablecoins)
- Regulatory uncertainty

**When it makes sense:**
- Traditional systems are slow/expensive
- Parties are in different jurisdictions
- No shared banking infrastructure

---

### ✅ 5. Supply Chain Transparency

**Problem:** Verifying authenticity in supply chains (e.g., conflict-free diamonds, organic food).

**Web3 solution:**
- Each step in supply chain writes to blockchain
- Consumers can verify provenance
- Tampering is detectable

**Examples:**
- **VeChain:** Supply chain tracking
- **IBM Food Trust:** Food safety tracking

**Tradeoffs accepted:**
- Garbage in, garbage out (blockchain doesn't verify physical world)
- Competitors can see your supply chain
- High cost per write

**When it makes sense:**
- Transparency matters (ethical sourcing)
- Multiple untrusted parties involved
- Consumers value verifiability

---

## Invalid Use Cases

### ❌ 1. User Data Storage (e.g., Social Media)

**Why it doesn't work:**
- Blockchains are public (no privacy)
- Storage is expensive ($1+ per photo)
- Can't delete content (GDPR violation)

**Better solution:**
- Traditional databases with encryption
- Federated systems (e.g., Mastodon)

---

### ❌ 2. High-Frequency Applications (Gaming, Trading)

**Why it doesn't work:**
- Blockchains are slow (15-30 TPS)
- High latency (12+ seconds per confirmation)

**Better solution:**
- Centralized servers with WebSockets
- If decentralization is needed: Layer 2 solutions or sidechains

**Exception:** Turn-based games or low-frequency state changes can work.

---

### ❌ 3. Internal Company Tools

**Why it doesn't work:**
- Employees are trusted (no adversaries)
- Privacy is needed (blockchain is public)
- Performance matters

**Better solution:**
- Traditional database with authentication

---

### ❌ 4. Healthcare Records

**Why it doesn't work:**
- Health data must be private (HIPAA, GDPR)
- Patients need to be able to delete data
- Immutability is harmful (what if doctor makes error?)

**Better solution:**
- Encrypted databases with access control
- Audit logs (don't need blockchain)

---

### ❌ 5. Real-Time Payments (Point of Sale)

**Why it doesn't work:**
- Confirmations take seconds to minutes
- Fees are unpredictable ($1-$50+)
- Poor UX (wallet setup, gas fees)

**Better solution:**
- Credit cards, Apple Pay, etc.

**Exception:** Lightning Network (Bitcoin Layer 2) enables faster payments, but adds complexity.

---

## The Checklist

Use this checklist before deciding on Web3:

### ✅ Web3 Might Be Appropriate If:
- [ ] Multiple parties don't trust each other
- [ ] No neutral intermediary exists
- [ ] Transparency is more important than privacy
- [ ] Censorship resistance is critical
- [ ] Users are technical enough to manage keys
- [ ] Transaction cost ($1-$100) is acceptable
- [ ] Transaction latency (12+ seconds) is acceptable
- [ ] Immutability is a feature, not a bug

### ❌ Web3 Is Probably NOT Appropriate If:
- [ ] A trusted intermediary exists and works fine
- [ ] Privacy is critical
- [ ] Performance matters (high TPS, low latency)
- [ ] Users need customer support
- [ ] Costs must be predictable and low
- [ ] Code needs to be updated frequently
- [ ] You can't explain why decentralization is necessary

---

## Case Studies

### Case Study 1: Wikipedia
**Should Wikipedia use Web3?**

**Analysis:**
- Centralized (Wikimedia Foundation is trusted)
- Needs to update/delete content (vandalism, errors)
- Free to use (gas fees would hurt accessibility)
- Doesn't need censorship resistance in most countries

**Decision: NO.** Wikipedia works fine as-is.

---

### Case Study 2: Prediction Markets (Pre-Election Betting)
**Should prediction markets use Web3?**

**Analysis:**
- Gambling is often illegal/restricted
- Platforms can be shut down
- Transparency in odds/settlement is valuable
- Users are sophisticated enough for Web3

**Decision: YES.** Censorship resistance justifies tradeoffs.

**Example:** Polymarket (blockchain-based prediction market)

---

### Case Study 3: Uber-Like Rideshare
**Should a rideshare app use Web3?**

**Analysis:**
- Real-time matching required (low latency)
- Disputes need resolution (customer support)
- Users want convenience, not decentralization
- Privacy matters (who you ride with)

**Decision: NO.** Performance and UX requirements make Web3 impractical.

**Exception:** A DAO could govern a cooperative rideshare company (like a union), but the app itself should be Web2.

---

### Case Study 4: Whistleblower Tip Platform
**Should a whistleblower platform use Web3?**

**Analysis:**
- Anonymity is critical
- Censorship resistance is critical
- Platform can be shut down by governments
- Users are motivated to learn Web3

**Decision: MAYBE.**
- Use blockchain for immutability and censorship resistance
- Use privacy-preserving tech (Tor, Zcash) for anonymity
- But acknowledge complexity may deter whistleblowers

---

### Case Study 5: Corporate Expense Tracking
**Should a company use Web3 for expense tracking?**

**Analysis:**
- Employees are trusted (not adversaries)
- Privacy is needed (competitors shouldn't see expenses)
- Needs to be fast and cheap
- Flexibility required (policy changes)

**Decision: NO.** Use a traditional database.

---

## The Honest Rubric

| Use Case | Trust Needed? | Needs Decentralization? | Can Tolerate Tradeoffs? | Use Web3? |
|----------|---------------|-------------------------|-------------------------|-----------|
| DeFi lending | Low | Yes (permissionless access) | Yes | ✅ |
| Digital collectibles | Medium | Yes (provenance) | Yes | ✅ |
| Crowdfunding | Low | Yes (censorship resistance) | Yes | ✅ |
| Social media | High | No (users trust the platform) | No (privacy, UX) | ❌ |
| Healthcare | High | No (privacy critical) | No (immutability harmful) | ❌ |
| Internal tools | High | No (trust exists) | No (performance needed) | ❌ |

---

## When in Doubt, Ask:

**"Could a Postgres database with proper authentication do this?"**

If the answer is yes, use Postgres.

---

## Summary

### Use Web3 When:
- Decentralization solves a real problem
- Transparency matters more than privacy
- Censorship resistance is critical
- Users can tolerate the tradeoffs (cost, UX, speed)

### Don't Use Web3 When:
- A trusted intermediary works fine
- Privacy, performance, or flexibility matter
- Users expect Web2 UX
- You can't justify why decentralization is necessary

---

## Exercise

For each scenario, decide: **Web3 or Web2?**

1. **A voting system for a public election**
   - Web3 or Web2?

2. **A ride-sharing app**
   - Web3 or Web2?

3. **A decentralized exchange for trading tokens**
   - Web3 or Web2?

4. **A hospital's patient records system**
   - Web3 or Web2?

5. **A fundraiser for a controversial political cause**
   - Web3 or Web2?

**Answers:**
1. **Depends.** Transparency is good, but voter privacy is critical. Hybrid approach?
2. **Web2.** Performance and UX are critical.
3. **Web3.** Censorship resistance and permissionless access matter.
4. **Web2.** Privacy is mandatory.
5. **Web3.** Censorship resistance justifies tradeoffs.

---

## Next Lesson

[→ Why Web3 Is Controversial](06-why-controversial.md)
