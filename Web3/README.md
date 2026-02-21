# Web3 for Web Developers

A skeptical, first-principles curriculum for web developers learning blockchain technology.

---

## What This Is

This is a **developer-first Web3 curriculum** designed for:
- Fullstack/backend web developers
- Engineers comfortable with TypeScript, databases, REST APIs
- People who understand Linux, networking, and distributed systems basics
- **Anyone with ZERO prior Web3 or crypto experience**

This curriculum teaches Web3 as:
- A distributed systems problem
- A database with adversaries
- A set of tradeoffs, not magic

---

## What This Is NOT

- ❌ Crypto hype or "get rich" material
- ❌ Token trading or NFT flipping tutorials
- ❌ "Web3 will replace everything" propaganda
- ❌ Influencer-style content

---

## Teaching Philosophy

1. **First principles over products**
   - Teach problems before solutions
   - Example: "How do strangers agree on state?" → THEN blockchain

2. **Web3 is a database with enemies**
   - Blockchains are append-only logs, replicated state machines, and public databases with adversaries

3. **Tradeoffs are the point**
   - Every benefit has costs: latency, throughput, expense, complexity

4. **No evangelism**
   - Be skeptical, precise, and honest about failures

5. **Incremental depth**
   - 🟢 Fundamentals
   - 🟡 Intermediate
   - 🔴 Advanced

---

## Curriculum Structure

```
Web3/
├── 00-orientation/             What Web3 actually is (and isn't)
├── 01-cryptography-primer/     Only the crypto you need
├── 02-blockchain-data-structures/  Blocks, chains, and Merkle trees
├── 03-consensus/               Agreeing without trust
├── 04-accounts-transactions/   State transitions and gas
├── 05-smart-contracts/         What they really are
├── 06-solidity/                Writing contracts (carefully)
├── 07-web2-vs-web3/            Architecture comparisons
├── 08-wallets-ux/              Why UX is painful
├── 09-security/                How things go catastrophically wrong
├── 10-tooling/                 Local development
├── 11-performance-scalability/ Why blockchains don't scale
├── 12-decentralization-reality/ Who really controls protocols
├── 13-use-cases/               When to use (and not use) Web3
├── 14-backend-pov/             Web3 from a backend engineer's view
└── 15-capstone/                Build a minimal Web3 app
```

---

## How to Use This Curriculum

1. **Read `START_HERE.md`** for setup and prerequisites
2. **Follow modules in order** (00 → 15)
3. **Do not skip fundamentals** (especially 02-03)
4. **Challenge assumptions** as you read
5. **Build the capstone project** (module 15) only after understanding modules 00-14

---

## Prerequisites

### Required
- Comfortable with TypeScript/JavaScript
- Understand REST APIs and HTTP
- Familiar with databases (SQL or NoSQL)
- Basic Linux command line usage

### Helpful (Not Required)
- Distributed systems concepts (consensus, replication)
- Computer networking (TCP/IP, DNS)
- Operating systems fundamentals
- Experience debugging production systems

---

## Key Themes

Throughout this curriculum, you'll encounter these recurring themes:

### Immutability vs Flexibility
- Code is permanent
- Bugs are permanent
- Upgrades are complex

### Decentralization vs Performance
- More nodes = slower consensus
- Public verification = higher cost
- Trustlessness = complexity

### Transparency vs Privacy
- Everything is public by default
- Pseudonymity ≠ anonymity
- On-chain data is permanent and readable

### Security as Necessity
- Attackers are assumed
- Economic incentives drive behavior
- Code vulnerabilities = permanent theft

---

## Why This Curriculum Exists

Most Web3 content is:
- Hype-driven
- Product-focused
- Lacking in computer science fundamentals
- Dismissive of valid criticisms

This curriculum:
- Treats Web3 as a technical tool, not a revolution
- Explains what problems blockchains solve (and don't solve)
- Acknowledges failures and limitations
- Prepares you to make informed engineering decisions

---

## When to Use Web3

Use Web3 when you need:
- Public verifiability of state
- Coordination without trusted intermediaries
- Censorship resistance
- Transparent rule execution

**Do NOT use Web3 when:**
- A database would work fine
- You need privacy
- You need performance
- You need flexibility
- You can afford downtime for upgrades

---

## Contribute or Provide Feedback

This curriculum is opinionated but aims to be accurate and helpful.

If you find:
- Technical errors
- Outdated information
- Unclear explanations
- Missing critical topics

Please open an issue or submit a pull request.

---

## License

This curriculum is provided as educational material. Use it freely.

---

**Start learning:** [START_HERE.md](START_HERE.md)
