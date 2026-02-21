# Module 15: Capstone Project

**Focus:** Build something real to solidify your learning.

---

## Overview

Time to build. This module guides you through creating a complete Web3 application from scratch, making all the architectural decisions yourself.

---

## Project Options

Choose based on your interests and skill level:

### Option 1: DeFi Protocol (Advanced)
Build a simplified lending protocol.

**Requirements:**
- Smart contracts for deposits/borrows
- Interest rate calculations
- Liquidation mechanism
- Frontend for interaction
- Backend indexer for data

**Skills:** Solidity, security, financial logic

---

### Option 2: NFT Marketplace (Intermediate)
Build a platform for minting and trading NFTs.

**Requirements:**
- ERC-721 contract
- Marketplace contract (listing, buying)
- IPFS integration for metadata
- Frontend with wallet connection
- Event indexing

**Skills:** Smart contracts, IPFS, frontend

---

### Option 3: DAO Governance Tool (Intermediate)
Build a tool for creating and managing DAOs.

**Requirements:**
- Governance token (ERC-20)
- Proposal and voting contracts
- Treasury management
- Frontend for proposals
- Vote tracking

**Skills:** Token standards, governance logic

---

### Option 4: Blockchain Analytics Dashboard (Backend-focused)
Build a real-time analytics platform.

**Requirements:**
- Node connection and event listening
- PostgreSQL for indexed data
- REST API for queries
- Caching layer (Redis)
- Dashboard frontend

**Skills:** Backend architecture, indexing, SQL

---

## Capstone Structure

### Week 1: Planning
- [Project scoping](01-scoping.md)
- Architecture design
- Technology selection
- Security considerations

### Week 2: Smart Contracts
- [Contract development](02-contracts.md)
- Testing strategy
- Gas optimization
- Security review

### Week 3: Backend/Infrastructure
- [Backend setup](03-backend.md)
- Event indexing
- API design
- Caching

### Week 4: Frontend
- [Frontend development](04-frontend.md)
- Wallet integration
- Transaction handling
- Error states

### Week 5: Integration & Testing
- [End-to-end testing](05-testing.md)
- Testnet deployment
- User testing
- Bug fixes

### Week 6: Production & Documentation
- [Mainnet deployment](06-deployment.md)
- Contract verification
- Documentation
- Post-mortem

---

## Evaluation Criteria

Rate yourself honestly on:

### Technical Implementation (40%)
- ✅ Smart contracts are secure and well-tested
- ✅ Gas optimization where it matters
- ✅ Proper error handling
- ✅ Code quality and documentation

### Architecture (30%)
- ✅ Appropriate use of blockchain (vs traditional DB)
- ✅ Handles reorgs and finality correctly
- ✅ Scalable design
- ✅ Security best practices

### UX (20%)
- ✅ Clear transaction states
- ✅ Helpful error messages
- ✅ Wallet connection works smoothly
- ✅ Responsive design

### Critical Thinking (10%)
- ✅ Can articulate tradeoffs made
- ✅ Understands limitations
- ✅ Knows what wouldn't work at scale

---

## Resources

All modules reference relevant code examples and patterns. Key references:

- [OpenZeppelin Contracts](https://docs.openzeppelin.com/contracts) - Battle-tested implementations
- [Solidity by Example](https://solidity-by-example.org) - Pattern reference
- [Ethereum Stack Exchange](https://ethereum.stackexchange.com) - Community Q&A

---

## After the Capstone

### You should now be able to:

1. **Evaluate blockchain appropriateness** for any use case
2. **Design** Web3 architectures with clear tradeoffs
3. **Implement** smart contracts securely
4. **Build** full-stack Web3 applications
5. **Debug** blockchain-specific issues
6. **Explain** Web3 concepts accurately (no hype)

### What's next?

- **Contribute to open source:** Many protocols need developers
- **Specialize:** Pick an area (security, DeFi, infrastructure)
- **Stay current:** Space evolves rapidly (follow EIPs, protocol upgrades)
- **Keep learning:** Zero-knowledge proofs, MEV, account abstraction

---

## Final Thoughts

Blockchain technology is powerful for specific use cases but oversold for most. You now have the knowledge to:

- Identify where it adds genuine value
- Build production-quality applications
- Avoid common security pitfalls
- Understand and explain the tradeoffs

**Most importantly:** You can make informed decisions, not hype-driven ones.

---

**[← Previous Module](../14-backend-developer-pov/README.md)** | **[↑ Back to Curriculum](../README.md)**
