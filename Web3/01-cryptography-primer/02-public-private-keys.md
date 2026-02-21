# Public-Private Key Cryptography

🟢 **Fundamentals**

---

## The Problem: How Do Strangers Authenticate?

In Web2, you log in with a username and password to a server that verifies your identity.

In Web3, there is no central server. How do you prove you're authorized to send a transaction?

**Answer:** Public-private key cryptography (also called asymmetric cryptography).

---

## What Are Public and Private Keys?

A **key pair** consists of:
1. **Private key:** A secret number (keep this safe!)
2. **Public key:** Derived from the private key (can be shared)

```
Private Key → (math) → Public Key → (hash) → Address
```

### Analogy
- **Private key** = Your house key (keep secret)
- **Public key** = Your house address (share freely)
- Anyone with your address can send you mail, but only you can unlock the door.

---

## How It Works

### 1. Key Generation
A private key is a randomly generated large number.

**Example (simplified):**
```
Private Key: 0x1234567890abcdef... (256 bits)
Public Key: Derived using elliptic curve math
```

In reality, the private key is 256 bits (77 decimal digits).

### 2. Deriving a Public Key
Public key = `G * privateKey` (elliptic curve multiplication)

Where `G` is a generator point on the curve.

**Key insight:** You can derive a public key from a private key, but NOT the reverse.

```
Private Key → Public Key   ✅ Easy
Public Key → Private Key   ❌ Computationally infeasible
```

### 3. Deriving an Address (Ethereum)
```
Private Key → Public Key → Keccak-256 → Last 20 bytes → Address
```

**Example:**
```
Private Key: 0x1234...
Public Key: 0x04a1b2c3d4... (64 bytes, uncompressed)
Hash: Keccak-256(public key)
Address: 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
```

The address is what you share publicly. It's like your username or account number.

---

## Elliptic Curve Cryptography (ECC)

Web3 uses **Elliptic Curve Digital Signature Algorithm (ECDSA)**.

**Why?**
- Small key sizes (256 bits)
- Fast computations
- Secure at current standards

**Curve used in Ethereum and Bitcoin:** `secp256k1`

---

## Code Example (TypeScript)

### Generate a key pair and address using ethers.js:

```typescript
import { ethers } from 'ethers';

// Generate a random wallet
const wallet = ethers.Wallet.createRandom();

console.log("Private Key:", wallet.privateKey);
console.log("Public Key:", wallet.publicKey);
console.log("Address:", wallet.address);
```

**Output (example):**
```
Private Key: 0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef
Public Key: 0x04a1b2c3d4e5f6...
Address: 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
```

---

## Key Properties

### 1. Private Key Controls the Account
Whoever has the private key controls the account (can spend funds, sign transactions).

**No password reset.** If you lose the private key, the account is locked forever.

### 2. Public Key Can Be Shared
Anyone can see your public key (derived from your address). This is safe.

### 3. Address Is Your Identity
Your address (e.g., `0x742d35...`) is how others send you funds or interact with you.

---

## Use Cases in Web3

### 1. Authentication (Signing Transactions)
You sign transactions with your private key.

```
Transaction data → Sign with private key → Signature
```

Anyone can verify:
```
Signature + Transaction data + Public key → Verify ✅
```

### 2. Proving Ownership
If you can sign a message with a private key, you own the corresponding address.

**Example:**
Website: "Sign this message to prove you own 0x742d35..."
You: Sign message with private key
Website: Verifies signature → Grants access

### 3. Receiving Funds
Your address is where others send ETH or tokens.

---

## Security Considerations

### ✅ Good Practices
1. **Never share your private key**
   - If someone gets your private key, they own your account.

2. **Store private keys securely**
   - Use hardware wallets (Ledger, Trezor) for large amounts.

3. **Backup your seed phrase**
   - Write it down on paper. Store it safely.

### ❌ Bad Practices
1. **Storing private keys in plain text**
   - Don't save them in unencrypted files.

2. **Entering private keys on websites**
   - Phishing sites steal keys.

3. **Sharing seed phrases**
   - Scammers often ask for seed phrases. Never share.

---

## Key Differences: Web2 vs Web3

| Aspect | Web2 (Username/Password) | Web3 (Public/Private Keys) |
|--------|--------------------------|----------------------------|
| **Identity** | Username (email, etc.) | Address (0x742d35...) |
| **Authentication** | Password (stored on server) | Private key (you hold it) |
| **Password reset** | Email link | Impossible (no reset) |
| **Account recovery** | Contact support | Impossible (if key is lost) |
| **Control** | Platform controls account | You control account |

---

## The Tradeoff

**Web3:**
- ✅ You control your identity
- ✅ No one can freeze your account
- ❌ If you lose your key, funds are gone forever
- ❌ No customer support

**Web2:**
- ✅ Password reset available
- ✅ Customer support
- ❌ Platform controls your account
- ❌ Account can be frozen/deleted

---

## Common Misconceptions

### ❌ "My wallet holds my crypto"
**Reality:** The blockchain holds your crypto. Your wallet holds your private key.

### ❌ "If I delete my wallet app, I lose my funds"
**Reality:** Your funds are on the blockchain. As long as you have your private key (or seed phrase), you can restore your wallet.

### ❌ "Public keys are like passwords"
**Reality:** Public keys can be shared. They don't need to be secret.

---

## Analogy: Physical Mailbox

- **Private key** = Key to your mailbox (only you have it)
- **Address** = Your mailbox location (everyone can see it)
- Anyone can deposit mail (send you funds)
- Only you can retrieve mail (spend funds)

---

## Math Behind It (Optional, Advanced)

### Elliptic Curve Equation (secp256k1):
```
y² = x³ + 7
```

### Key Generation:
```
Private key: Random number k (256 bits)
Public key: P = k * G (where G is a generator point)
```

### Why it's secure:
Given `P` and `G`, finding `k` is called the **discrete logarithm problem** and is computationally infeasible.

---

## Exercise

### 1. Generate a Key Pair

Use ethers.js to generate a wallet:

```typescript
import { ethers } from 'ethers';

const wallet = ethers.Wallet.createRandom();
console.log("Private Key:", wallet.privateKey);
console.log("Address:", wallet.address);
```

### 2. Derive Address from Private Key

Given a private key, derive its address:

```typescript
const privateKey = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef";
const wallet = new ethers.Wallet(privateKey);
console.log("Address:", wallet.address);
```

### 3. Understand Irreversibility

Try to "reverse" an address back to its private key. Realize it's impossible without brute-forcing 2^256 possibilities.

---

## Summary

**Public-private key cryptography:**
- Private key = secret number (controls your account)
- Public key = derived from private key (can be shared)
- Address = derived from public key (your identity)

**Key insights:**
- Private key → Public key (easy)
- Public key → Private key (impossible)
- Whoever controls the private key controls the account

**In Web3:**
- No passwords
- No centralized authentication
- Full self-custody (and full responsibility)

---

## Next Lesson

[→ Digital Signatures](03-digital-signatures.md)
