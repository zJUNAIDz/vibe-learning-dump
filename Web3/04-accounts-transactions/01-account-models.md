# Account Models

🟢 **Fundamentals**

---

## Two Paradigms

Blockchains track value using two different models:

1. **UTXO (Unspent Transaction Output)** - Bitcoin
2. **Account-Based** - Ethereum

---

## UTXO Model (Bitcoin)

**Concept:** No "accounts" with balances. Only unspent transaction outputs.

### How It Works

Your "balance" = sum of all UTXOs you can spend (have private keys for).

**Example:**
```
Alice receives 5 BTC from Bob
Alice receives 3 BTC from Charlie

Alice's balance: 5 + 3 = 8 BTC

These are TWO separate UTXOs, not one account.
```

### Spending UTXOs

To send 6 BTC, Alice must:
1. Consume entire UTXOs (can't spend partial)
2. Create new UTXO for recipient (6 BTC)
3. Create change UTXO for herself (2 BTC)

```
Input: 5 BTC UTXO + 3 BTC UTXO = 8 BTC
Output: 6 BTC to David + 2 BTC back to Alice
```

**Analogy:** Like cash. You can't spend "half" a $20 bill. Give whole bill, get change.

---

### UTXO Advantages

✅ **Privacy:** Each transaction can use new address  
✅ **Parallel processing:** UTXOs independent, can validate in parallel  
✅ **Simpler model:** No account state to track

### UTXO Disadvantages

❌ **Complex for apps:** Smart contracts awkward (no persistent state)  
❌ **UTXO bloat:** Must track all unspent outputs  
❌ **Poor UX:** "Change" outputs confusing

---

## Account-Based Model (Ethereum)

**Concept:** Explicit accounts with balances (like bank accounts).

### Account Structure

```typescript
interface Account {
  nonce: number;        // Transaction count (prevents replays)
  balance: bigint;      // Wei (1 ETH = 10^18 wei)
  storageRoot: string;  // Merkle root of storage (contracts only)
  codeHash: string;     // Hash of contract code (contracts only)
}
```

### Two Account Types

#### 1. Externally Owned Account (EOA)
Controlled by private key (humans/wallets).

```typescript
const EOA: Account = {
  nonce: 5,           // Sent 5 transactions
  balance: 2000000000000000000n, // 2 ETH in wei
  storageRoot: EMPTY_HASH,       // No storage
  codeHash: EMPTY_HASH           // No code
};
```

#### 2. Contract Account
Controlled by code (smart contracts).

```typescript
const Contract: Account = {
  nonce: 1,           // Created 1 contract
  balance: 500000000000000000n,  // 0.5 ETH
  storageRoot: '0x123...',       // Persistent storage
  codeHash: '0xabc...'           // Contract bytecode hash
};
```

---

### Sending ETH (Account Model)

Simple: Subtract from sender, add to receiver.

```typescript
function transfer(from: Account, to: Account, amount: bigint): void {
  if (from.balance < amount) {
    throw new Error('Insufficient balance');
  }
  
  from.balance -= amount;
  to.balance += amount;
  from.nonce += 1; // Increment transaction count
}
```

**Much simpler than UTXO's input/output model.**

---

## Comparison

| Aspect | UTXO (Bitcoin) | Account (Ethereum) |
|--------|----------------|-------------------|
| **Mental Model** | Cash/coins | Bank account |
| **Balance** | Sum of UTXOs | Single number |
| **Privacy** | Better (new address per tx) | Worse (reuse address) |
| **Smart Contracts** | Difficult | Natural |
| **Parallelization** | Easier | Harder (shared state) |
| **Storage** | All UTXOs | Account state |

---

## Ethereum Address Derivation

**Address = last 20 bytes of Keccak256(publicKey)**

```typescript
import { ethers } from 'ethers';
import crypto from 'crypto';

function addressFromPrivateKey(privateKeyHex: string): string {
  // 1. Private key → Public key (ECDSA on secp256k1)
  const wallet = new ethers.Wallet(privateKeyHex);
  const publicKey = wallet.publicKey; // 65 bytes (uncompressed)
  
  // 2. Public key → Keccak256 hash
  const publicKeyBytes = Buffer.from(publicKey.slice(2), 'hex'); // Remove '0x'
  const hash = ethers.keccak256(publicKeyBytes);
  
  // 3. Take last 20 bytes
  const address = '0x' + hash.slice(-40);
  
  return ethers.getAddress(address); // Checksum format
}

// Example
const privKey = '0x' + crypto.randomBytes(32).toString('hex');
const address = addressFromPrivateKey(privKey);

console.log('Private Key:', privKey);
console.log('Address:', address);
```

---

## EIP-55: Checksummed Addresses

**Problem:** Typos in addresses lose funds permanently.

**Solution:** Mixed-case checksum.

```typescript
function toChecksumAddress(address: string): string {
  const addr = address.toLowerCase().replace('0x', '');
  const hash = ethers.keccak256(Buffer.from(addr, 'utf-8'));
  
  let checksummed = '0x';
  for (let i = 0; i < addr.length; i++) {
    // If hash byte > 7, capitalize
    if (parseInt(hash[i], 16) > 7) {
      checksummed += addr[i].toUpperCase();
    } else {
      checksummed += addr[i];
    }
  }
  
  return checksummed;
}

// Example
console.log(toChecksumAddress('0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed'));
// Output: 0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed
//         ↑ ↑   ↑    ↑  ↑      ↑     ↑  ↑   ↑    ↑
//         Capitalized based on hash (error detection)
```

**If you mistype one character, checksum won't match → wallet rejects.**

---

## Account State Lookup

All account data stored in **State Trie** (Patricia Merkle Tree).

```typescript
// Simplified state lookup
class StateTree {
  state: Map<string, Account> = new Map();
  
  getAccount(address: string): Account {
    return this.state.get(address) || {
      nonce: 0,
      balance: 0n,
      storageRoot: EMPTY_ROOT,
      codeHash: EMPTY_CODE
    };
  }
  
  setAccount(address: string, account: Account): void {
    this.state.set(address, account);
  }
  
  // State root = Merkle root of all accounts
  getStateRoot(): string {
    // Simplified: just hash all accounts
    const allAccounts = JSON.stringify([...this.state.entries()]);
    return ethers.keccak256(Buffer.from(allAccounts));
  }
}
```

**Every block header contains `stateRoot`** - commits to entire world state.

---

## Summary

**UTXO Model (Bitcoin):**
- No accounts, only transaction outputs
- More privacy, simpler validation
- Awkward for smart contracts

**Account Model (Ethereum):**
- Explicit accounts with balances
- Two types: EOA (user) and Contract (code)
- Better for apps, shared state

**Key insight:** The account model you choose affects everything - smart contracts, privacy, scalability.

**Next:** How do transactions actually work in the account model?

---

**[Next Lesson →](02-transaction-anatomy.md)**
