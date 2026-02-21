# What Smart Contracts Really Are

🟢 **Fundamentals**

---

## The Name is Misleading

**"Smart Contract" suggests:**
- Artificial intelligence ❌
- Legal contracts ❌
- Flexibility ❌

**Reality:** Deterministic programs that execute on-chain.

---

## What They Actually Are

```typescript
// This is closer to reality
class SmartContract {
  // Immutable code
  private code: string; // Can't be changed after deployment
  
  // Mutable state (expensive to change)
  private storage: Map<string, any>;
  
  // Must be deterministic
  execute(input: string): string {
    // Same input → Same output
    // No randomness, no external API calls, no current time
    return deterministicFunction(input);
  }
}
```

**Better name:** "On-chain programs" or "Replicated state machines"

---

## Key Properties

### 1. Immutability

Once deployed, code cannot be changed.

**Analogy:** Publishing a book. Typos are permanent.

```solidity
// Deployed with a bug
contract MyContract {
  function transfer(address to, uint amount) public {
    // Bug: no balance check!
    balances[to] += amount;
    balances[msg.sender] -= amount; // Can go negative
  }
}

// Can't fix it. Bug is permanent.
// Must deploy new contract, convince users to migrate
```

**Consequence:** Bugs are permanent. Funds can be lost forever.

---

### 2. Transparency

All code and state is public.

**Good:** Anyone can audit code  
**Bad:** No secrets. No private business logic.

```solidity
contract Auction {
  uint256 private reservePrice = 1000; // Marked "private"
  
  // But "private" only prevents other contracts from reading
  // Anyone can read this value from blockchain storage
}

// Read private variable
const reservePrice = await provider.getStorageAt(
  contractAddress, 
  0 // Storage slot
);
```

**Zero privacy by default.**

---

### 3. Determinism

Must produce same result given same input.

**Allowed:**
- Math operations
- Reading blockchain state
- Cryptographic functions

**Not allowed:**
- `Math.random()` (no randomness)
- API calls (no external data)
- `Date.now()` (no current time, use block.timestamp)
- File system (no persistence outside blockchain)

```solidity
// ❌ Won't compile - non-deterministic
contract Bad {
  function getWeather() public returns (string memory) {
    // Can't call external API
    return fetch("https://api.weather.com");
  }
}

// ✅ Deterministic
contract Good {
  function add(uint a, uint b) public pure returns (uint) {
    return a + b; // Same inputs → same output
  }
}
```

---

### 4. Gas Costs

Every operation costs gas (paid in ETH).

**Consequence:** Computation is expensive.

```typescript
// Traditional backend
function filterUsers(users: User[]): User[] {
  return users.filter(u => u.age > 18); // Free
}

// Smart contract
// Iterating over array costs gas for EVERY item
// Large arrays can exceed block gas limit
// Result: Can't do it
```

**Pattern:** Do complex computation off-chain, verify on-chain.

---

## What You Can and Can't Do

### ✅ Can Do

**1. Value Transfer**
```solidity
function sendEther(address payable recipient) public payable {
  recipient.transfer(msg.value);
}
```

**2. Conditional Logic**
```solidity
function withdraw() public {
  require(balances[msg.sender] > 0, "No balance");
  // ...
}
```

**3. State Storage**
```solidity
mapping(address => uint) public balances;
```

**4. Emit Events (for off-chain listening)**
```solidity
event Transfer(address from, address to, uint amount);
```

---

### ❌ Can't Do

**1. Call External APIs**
```solidity
// ❌ Impossible
function getStockPrice() public returns (uint) {
  return fetch("https://api.stocks.com/AAPL");
}
```

**Workaround:** Oracles (Chainlink) - but they introduce trust assumptions.

---

**2. Generate Randomness**
```solidity
// ❌ Predictable (miners can manipulate)
function random() public view returns (uint) {
  return uint(keccak256(abi.encodePacked(block.timestamp)));
}
```

**Workaround:** VRF (Verifiable Random Function) like Chainlink VRF.

---

**3. Schedule Future Execution**
```solidity
// ❌ No built-in cron jobs
function executeIn1Hour() public {
  // Can't automatically run code later
}
```

**Workaround:** Off-chain bots calling contract functions.

---

**4. Process Large Datasets**
```solidity
// ❌ Will exceed gas limit
function sumAll(uint[] memory numbers) public returns (uint) {
  uint sum = 0;
  for (uint i = 0; i < numbers.length; i++) {
    sum += numbers[i];
  }
  return sum; // If array has 10,000 items, transaction fails
}
```

**Workaround:** Pagination, Merkle proofs, or zkSNARKs.

---

## Common Misconceptions

### Misconception 1: Smart Contracts are AI
**Reality:** Dumb programs. No learning, no adaptation.

### Misconception 2: Smart Contracts are legally binding
**Reality:** Code, not law. Courts don't recognize them (yet).

### Misconception 3: Smart Contracts are free/cheap
**Reality:** Every operation costs gas. Can cost $100+ for complex transactions.

### Misconception 4: Smart Contracts are bug-free
**Reality:** Written by humans. Bugs common and permanent.

### Misconception 5: Smart Contracts are private
**Reality:** All code and data is public.

---

## Comparison to Traditional Programs

| Aspect | Traditional Program | Smart Contract |
|--------|---------------------|----------------|
| **Deployment** | Can be updated | Immutable |
| **Execution** | Runs on your servers | Runs on thousands of nodes |
| **Cost** | Server costs | Gas fees per execution |
| **Data privacy** | Private by default | Public by default |
| **Randomness** | `Math.random()` works | No true randomness |
| **External data** | API calls work | Need oracles |
| **Bug fixes** | Deploy patch | Deploy new contract, migrate users |
| **Performance** | Fast (milliseconds) | Slow (seconds to minutes) |

---

## Why Use Smart Contracts Then?

Despite limitations, they enable things traditional systems can't:

### 1. No Trusted Intermediary
```
Traditional: You → Bank → Recipient (bank can freeze funds)
Smart Contract: You → Contract → Recipient (no one can stop it)
```

### 2. Composability
Contracts can call other contracts. Building blocks.

```solidity
// Use Uniswap to swap, then deposit to Aave, then stake
// All in one transaction, no intermediaries
```

### 3. Transparency
Anyone can verify the code. No hidden backdoors.

### 4. Censorship Resistance
If blockchain is running, contract works. Can't be shut down.

---

## Real Example: ERC-20 Token

```solidity
// Simplified ERC-20 token
contract Token {
  mapping(address => uint256) public balances;
  
  function transfer(address to, uint256 amount) public {
    require(balances[msg.sender] >= amount, "Insufficient balance");
    
    balances[msg.sender] -= amount;
    balances[to] += amount;
    
    emit Transfer(msg.sender, to, amount);
  }
}
```

**What this enables:**
- Anyone can verify the supply (read `balances`)
- No central authority can freeze accounts
- Token works as long as Ethereum works
- Anyone can build on top (DEXs, lending)

**What this costs:**
- Every transfer costs gas (~$1-10)
- Bugs are permanent
- No privacy (all transfers public)

---

## Summary

**Smart contracts are:**
- Immutable programs on blockchain
- Transparent (all code/data public)
- Deterministic (same input → same output)
- Expensive (gas costs)

**Not:**
- Artificial intelligence
- Legal contracts
- Free to execute
- Private
- Flexible after deployment

**Use when:**
- Need censorship resistance
- Want to eliminate trusted intermediaries
- Benefit from transparency
- Value composability

**Don't use when:**
- Need privacy
- Need flexibility (frequent updates)
- Need performance (high throughput)
- Need external data (APIs)

**Next:** Understanding the EVM - the machine that executes these contracts.

---

**[Next Lesson →](02-evm-basics.md)**
