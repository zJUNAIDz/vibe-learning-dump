# Quick Reference

Fast lookup for key Web3 concepts, terms, and comparisons.

---

## Core Concepts

| Term | Definition | Web2 Analogy |
|------|------------|--------------|
| **Blockchain** | Append-only distributed ledger replicated across nodes | Database with public replication |
| **Block** | Batch of transactions with hash pointer to previous block | Database transaction log entry |
| **Hash** | One-way cryptographic fingerprint of data | Checksum (but cryptographically secure) |
| **Node** | Computer running blockchain software | Database server / API server |
| **Consensus** | Process by which nodes agree on state | Database replication protocol |
| **Smart Contract** | Immutable code deployed on blockchain | Backend API service (but permanent) |
| **Transaction** | State transition request | API POST/PUT request |
| **Gas** | Computational cost of executing transaction | API rate limiting / compute billing |
| **Wallet** | Software managing private keys | Authentication token manager |
| **Private Key** | Secret used to sign transactions | Password / SSH private key |
| **Public Key / Address** | Derived from private key, identifies account | Username / account ID |

---

## Key Tradeoffs

| Feature | Blockchain | Traditional Database |
|---------|-----------|---------------------|
| **Trust Model** | Trustless (adversarial) | Trusted (authenticated) |
| **Write Speed** | Slow (seconds to minutes) | Fast (milliseconds) |
| **Read Speed** | Fast (if indexed) | Fast |
| **Cost per Write** | High ($0.01 - $100+) | Near-zero |
| **Immutability** | Permanent | Mutable |
| **Privacy** | Public by default | Private by default |
| **Scalability** | Low (10-1000 TPS) | High (10K-100K+ TPS) |
| **Code Updates** | Complex/impossible | Easy |
| **Downtime** | Rare | Possible |

---

## Web2 vs Web3 Architecture

### Web2 Stack
```
User → Frontend → Backend API → Database → Cache → Queue
```

### Web3 Stack
```
User → Frontend → Wallet → RPC Node → Blockchain
            ↓
      Backend API (still needed for UX)
            ↓
      Database (indexing blockchain data)
```

**Key insight:** Most Web3 apps still need Web2 infrastructure.

---

## Consensus Mechanisms

| Mechanism | How It Works | Pros | Cons |
|-----------|--------------|------|------|
| **Proof of Work (PoW)** | Miners solve puzzles to propose blocks | Battle-tested, secure | Energy-intensive, slow |
| **Proof of Stake (PoS)** | Validators stake capital to propose blocks | Energy-efficient, faster | Plutocracy concerns |

---

## Ethereum Basics

### Account Types
- **Externally Owned Account (EOA):** Controlled by private key (users)
- **Contract Account:** Controlled by code (smart contracts)

### Transaction Fields
```typescript
{
  from: address,      // Sender
  to: address,        // Recipient or contract
  value: wei,         // Amount of ETH to send
  data: bytes,        // Function call data
  gas: number,        // Max gas willing to spend
  gasPrice: wei,      // Price per unit of gas
  nonce: number       // Transaction counter (prevents replays)
}
```

### Gas
- **Gas Limit:** Max computation you're willing to pay for
- **Gas Price:** Amount you pay per unit of gas
- **Total Cost:** `gas_used * gas_price`

Example:
- Gas limit: 21000
- Gas price: 50 gwei
- **Cost:** 21000 × 50 = 1,050,000 gwei = 0.00105 ETH

---

## Solidity Cheat Sheet

### Basic Contract
```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract SimpleStorage {
    uint256 public value;  // State variable
    
    function setValue(uint256 _value) public {
        value = _value;
    }
    
    function getValue() public view returns (uint256) {
        return value;
    }
}
```

### Common Data Types
- `uint256`, `int256`: Integers
- `address`: Ethereum address (20 bytes)
- `bool`: True/false
- `string`: Text data
- `bytes`: Binary data
- `mapping(address => uint)`: Key-value store

### Visibility Modifiers
- `public`: Anyone can call
- `external`: Only callable from outside
- `internal`: Only within contract and derived contracts
- `private`: Only within contract

### Function Modifiers
- `view`: Reads state, doesn't modify
- `pure`: Doesn't read or modify state
- `payable`: Can receive ETH

---

## Common Vulnerabilities

| Vulnerability | Description | Prevention |
|---------------|-------------|------------|
| **Reentrancy** | Calling external contract before updating state | Use checks-effects-interactions pattern |
| **Integer Overflow** | Math operations exceed type limits | Use Solidity 0.8+ (built-in checks) |
| **Access Control** | Missing permission checks | Use modifiers, OpenZeppelin's Ownable |
| **Front-running** | Attackers see txs before execution | Design protocols to be front-run resistant |
| **Oracle Manipulation** | Bad data from external sources | Use decentralized oracles (Chainlink) |

---

## Useful Commands

### Node / npm
```bash
node --version              # Check Node version
npm init -y                 # Initialize project
npm install <package>       # Install dependency
npx hardhat compile         # Compile contracts (Hardhat)
npx hardhat test            # Run tests
npx hardhat node            # Start local blockchain
```

### Foundry
```bash
forge init                  # Create new project
forge build                 # Compile contracts
forge test                  # Run tests
forge test -vvv             # Verbose test output
anvil                       # Start local blockchain
```

### Git
```bash
git clone <url>             # Clone repository
git status                  # Check status
git add .                   # Stage changes
git commit -m "message"     # Commit changes
git push                    # Push to remote
```

---

## Units

### Ether Units
```
1 wei = 1
1 gwei = 1,000,000,000 wei (10^9)
1 ether = 1,000,000,000,000,000,000 wei (10^18)
```

### Time Units (Solidity)
```
1 seconds = 1
1 minutes = 60 seconds
1 hours = 60 minutes
1 days = 24 hours
1 weeks = 7 days
```

---

## Important Addresses

### Ethereum Mainnet
- **Chain ID:** 1
- **Currency:** ETH
- **Block Time:** ~12 seconds

### Test Networks
- **Sepolia:** Chain ID 11155111 (recommended testnet)
- **Goerli:** Chain ID 5 (being deprecated)

### Special Addresses
- **Zero Address:** `0x0000000000000000000000000000000000000000` (burn address)

---

## Key Resources

### Documentation
- Ethereum docs: https://ethereum.org/developers
- Solidity docs: https://docs.soliditylang.org
- Hardhat docs: https://hardhat.org/docs
- Foundry book: https://book.getfoundry.sh

### Security
- OpenZeppelin: https://openzeppelin.com/contracts
- Consensys best practices: https://consensys.github.io/smart-contract-best-practices

### Block Explorers
- Etherscan: https://etherscan.io
- Tenderly: https://tenderly.co

---

## Common Acronyms

- **DApp:** Decentralized Application
- **DEX:** Decentralized Exchange
- **DeFi:** Decentralized Finance
- **DAO:** Decentralized Autonomous Organization
- **EVM:** Ethereum Virtual Machine
- **EOA:** Externally Owned Account
- **MEV:** Maximal Extractable Value (front-running, etc.)
- **NFT:** Non-Fungible Token
- **RPC:** Remote Procedure Call (how apps talk to nodes)
- **TPS:** Transactions Per Second
- **TVL:** Total Value Locked
- **UTXO:** Unspent Transaction Output (Bitcoin model)

---

## When to Use Web3

✅ **Use Web3 when:**
- You need public verifiability
- No trusted intermediary exists
- Censorship resistance is critical
- Transparent rule execution matters

❌ **Do NOT use Web3 when:**
- A database would work
- You need privacy
- You need performance (high TPS)
- You need to update code easily
- Cost per transaction matters

---

## Learn More

- [README.md](README.md) - Curriculum overview
- [START_HERE.md](START_HERE.md) - How to begin
- [00-orientation/](00-orientation/) - Start module 00

---

**Bookmark this page** for quick reference while learning.
