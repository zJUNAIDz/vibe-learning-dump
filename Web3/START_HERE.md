# Start Here

Welcome to **Web3 for Web Developers**—a curriculum designed to teach blockchain technology from first principles, without the hype.

---

## Before You Begin

### Who This Is For

You should start this curriculum if you:
- Are a web developer (fullstack or backend)
- Understand TypeScript, REST APIs, and databases
- Have Linux/command-line experience
- Want to understand Web3 deeply, not superficially
- Are skeptical and want honest tradeoff analysis

### Who This Is NOT For

This curriculum may not be for you if you:
- Want "quick wins" or trading advice
- Expect to "get rich" from crypto
- Believe Web3 will replace all existing systems
- Don't have foundational programming experience

---

## Prerequisites

### Required Skills
- **Programming:** Comfortable with TypeScript/JavaScript
- **Web Development:** REST APIs, HTTP, JSON
- **Databases:** SQL or NoSQL experience
- **Linux:** Can navigate the command line

### Recommended (But Not Required)
- Distributed systems concepts
- Computer networking basics
- Operating systems fundamentals
- Experience with production debugging

---

## Setup Your Environment

### 1. Install Node.js
You'll need Node.js 18+ for most Web3 tooling.

```bash
# Check if Node is installed
node --version

# On Fedora/RHEL
sudo dnf install nodejs

# Or use nvm for version management
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
nvm install 18
nvm use 18
```

### 2. Install a Code Editor
Use VS Code or your preferred editor with TypeScript support.

### 3. (Optional) Install Docker
For running local blockchain nodes later in the curriculum:

```bash
# Fedora/RHEL
sudo dnf install docker
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER
```

Log out and back in for group changes to take effect.

### 4. Create a Learning Directory
```bash
mkdir -p ~/web3-learning
cd ~/web3-learning
```

---

## How to Use This Curriculum

### Read in Order
Modules build on each other. Do not skip fundamentals.

**Mandatory sequence:**
```
00 → 01 → 02 → 03 → 04 → 05 → 06 → ... → 15
```

### Track Your Progress
Each module includes:
- 🟢 **Fundamentals:** Core concepts you must understand
- 🟡 **Intermediate:** Deeper technical details
- 🔴 **Advanced:** Optional deep dives for the curious

If you're new to Web3, focus on 🟢 and 🟡 sections first.

### Take Notes
Web3 has lots of jargon. Keep a personal glossary as you learn:
- New terms
- Acronyms
- Concepts that confuse you

### Challenge Everything
This curriculum encourages skepticism. Ask:
- "Why is this necessary?"
- "What are the tradeoffs?"
- "Could a database do this?"
- "What could go wrong?"

### Do the Exercises
Most modules include exercises. Do them.

### Build the Capstone
Module 15 is a small Web3 app that ties everything together. Complete it to validate your understanding.

---

## Curriculum Overview

### Phase 1: Fundamentals (Modules 00-03)
**Goal:** Understand what blockchains are and why they exist.

- **00-orientation:** What Web3 actually is
- **01-cryptography-primer:** Hashes, keys, signatures
- **02-blockchain-data-structures:** Blocks, chains, Merkle trees
- **03-consensus:** How strangers agree on state

**Time estimate:** 1-2 weeks

---

### Phase 2: Execution Layer (Modules 04-06)
**Goal:** Understand how state is managed and code is executed.

- **04-accounts-transactions:** State transitions, gas, nonces
- **05-smart-contracts:** What they really are
- **06-solidity:** Writing contracts (carefully)

**Time estimate:** 2-3 weeks

---

### Phase 3: Practical Development (Modules 07-10)
**Goal:** Understand how to build Web3 apps and what's different from Web2.

- **07-web2-vs-web3:** Architecture comparisons
- **08-wallets-ux:** User experience challenges
- **09-security:** How things go catastrophically wrong
- **10-tooling:** Local development environments

**Time estimate:** 2-3 weeks

---

### Phase 4: Reality Check (Modules 11-14)
**Goal:** Understand limitations, tradeoffs, and real-world considerations.

- **11-performance-scalability:** Why blockchains don't scale
- **12-decentralization-reality:** Who really controls protocols
- **13-use-cases:** When to use (and not use) Web3
- **14-backend-pov:** Web3 from a backend engineer's view

**Time estimate:** 1-2 weeks

---

### Phase 5: Capstone (Module 15)
**Goal:** Build a complete (but minimal) Web3 application.

- **15-capstone:** Build, deploy, and threat-model a Web3 app

**Time estimate:** 1-2 weeks

---

## Expected Timeline

**Full curriculum:** 7-12 weeks (depending on depth and pace)

**Accelerated path (fundamentals only):** 3-4 weeks

---

## Learning Tips

### 1. Don't Memorize, Understand
Web3 changes fast. Focus on concepts, not syntax.

### 2. Compare to What You Know
Constantly map Web3 concepts to Web2 equivalents:
- Smart contracts ↔ Backend APIs
- Blockchains ↔ Databases
- Wallets ↔ Authentication

### 3. Read the Skeptics
Web3 has valid criticisms. Read them:
- Moxie Marlinspike's "My First Impressions of Web3"
- Molly White's "Web3 is Going Just Great"
- Nicholas Weaver's talks on blockchain limitations

### 4. Build Small Things
Don't wait until module 15 to code. Experiment as you learn.

### 5. Join Technical Communities
Avoid hype communities. Look for:
- r/ethdev (Reddit)
- Ethereum Stack Exchange
- Developer-focused Discord servers (e.g., Hardhat, Foundry)

---

## Common Pitfalls to Avoid

### ❌ Skipping Fundamentals
If you don't understand consensus (module 03), you won't understand why Web3 is slow and expensive.

### ❌ Trusting Tutorials Blindly
Many Web3 tutorials teach insecure patterns. Always question what you read.

### ❌ Ignoring Costs
Every transaction costs money. Every storage write costs money. Every computation costs money.

### ❌ Assuming Decentralization = Good
Decentralization is a tool, not a goal. It has costs.

### ❌ Treating Web3 as "The Future"
Web3 solves specific problems. It's not a replacement for existing systems.

---

## How to Get Help

### If You're Stuck
1. Reread the previous module
2. Check the quick reference: `QUICK_REFERENCE.md`
3. Search Ethereum Stack Exchange
4. Ask in developer communities (be specific)

### If You Find Errors
- Open an issue
- Submit a pull request
- Suggest improvements

---

## Ready?

Start with **Module 00: Orientation**

[→ Go to 00-orientation/](00-orientation/)

---

**Remember:** Web3 is a tool, not magic. Approach it with curiosity and skepticism.
