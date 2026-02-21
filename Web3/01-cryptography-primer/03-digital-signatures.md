# Digital Signatures

🟢 **Fundamentals**

---

## The Problem: How Do You Prove You Authorized a Transaction?

In Web2:
- You log into a server
- Server verifies your session
- Server processes your request

In Web3:
- No central server
- How do nodes know you authorized a transaction?

**Answer:** Digital signatures.

---

## What Is a Digital Signature?

A **digital signature** is cryptographic proof that:
1. You authorized a message/transaction
2. The message hasn't been tampered with

**Process:**
```
1. Create transaction
2. Sign transaction with your private key → Signature
3. Broadcast transaction + signature
4. Anyone can verify signature using your public key
```

---

## How Signing Works

### Step 1: Create a Transaction
```typescript
const transaction = {
  from: "0x742d35...",
  to: "0xABCDEF...",
  value: "1 ETH",
  nonce: 5,
  gas: 21000
};
```

### Step 2: Hash the Transaction
```
hash = Keccak-256(transaction)
```

### Step 3: Sign the Hash with Your Private Key
```
signature = sign(hash, privateKey)
```

The signature consists of three values: `(r, s, v)`
- `r` and `s`: Cryptographic proof
- `v`: Recovery ID (helps derive public key from signature)

### Step 4: Broadcast Transaction + Signature
```
{
  transaction data,
  signature: (r, s, v)
}
```

---

## How Verification Works

Anyone (including blockchain nodes) can verify:

```
1. Hash the transaction data
2. Use the signature (r, s, v) to recover the public key
3. Derive the address from the public key
4. Check: Does the address match transaction.from?
   - If yes → signature is valid ✅
   - If no → signature is invalid ❌
```

**Key insight:** You never reveal your private key. The signature proves you have it.

---

## Code Example (TypeScript)

### Signing a Message

```typescript
import { ethers } from 'ethers';

// Create a wallet (or load from private key)
const wallet = ethers.Wallet.createRandom();

// Message to sign
const message = "I authorize this action";

// Sign the message
const signature = await wallet.signMessage(message);

console.log("Signature:", signature);
// Signature: 0x1234abcd... (130 characters)
```

### Verifying a Signature

```typescript
import { ethers } from 'ethers';

const message = "I authorize this action";
const signature = "0x1234abcd...";

// Recover the address that signed this message
const recoveredAddress = ethers.verifyMessage(message, signature);

console.log("Signer address:", recoveredAddress);

// Check if it matches the expected address
if (recoveredAddress === wallet.address) {
  console.log("Signature valid ✅");
} else {
  console.log("Signature invalid ❌");
}
```

---

## Signing Transactions

### Example: Sign an Ethereum Transaction

```typescript
import { ethers } from 'ethers';

const wallet = new ethers.Wallet("0xYOUR_PRIVATE_KEY");

// Create a transaction
const tx = {
  to: "0xRecipientAddress",
  value: ethers.parseEther("1.0"), // 1 ETH
  gasLimit: 21000,
  gasPrice: ethers.parseUnits("50", "gwei"),
  nonce: 5
};

// Sign the transaction
const signedTx = await wallet.signTransaction(tx);

console.log("Signed Transaction:", signedTx);
// This can now be broadcast to the network
```

---

## Properties of Digital Signatures

### 1. Authentication
Proves the signer has the private key (without revealing it).

### 2. Non-Repudiation
The signer can't deny signing the message later.

### 3. Integrity
If the message is altered, the signature becomes invalid.

### 4. Public Verifiability
Anyone with the public key can verify the signature.

---

## How Digital Signatures Secure Web3

### 1. Transaction Authorization
Every transaction must be signed with your private key.

**Without this:**
- Anyone could send transactions from your address
- No way to prove you authorized an action

### 2. Message Signing (Off-Chain Authentication)
Websites can ask you to sign a message to prove you own an address.

**Example:**
```
Website: "Sign this message to log in"
You: Sign message with MetaMask
Website: Verifies signature → Grants access
```

This replaces username/password authentication.

### 3. Multi-Signature Wallets
Require multiple signatures to authorize a transaction.

**Example:**
- 3-of-5 multisig wallet
- Need 3 out of 5 keyholders to sign a transaction

---

## Signature Components (r, s, v)

An Ethereum signature consists of three values:

### r and s
Cryptographic proof derived from:
- The private key
- The hash of the message
- A random nonce

### v (Recovery ID)
A small value (27 or 28) that helps recover the public key from the signature.

**Why v matters:**
Ethereum signatures allow you to recover the signer's public key (and thus address) without needing to store it separately.

---

## Common Signature Standards

### 1. ECDSA (Web3 Standard)
- Used in Bitcoin, Ethereum
- Curve: secp256k1

### 2. EdDSA (Alternative)
- Used in some newer blockchains (Solana, Algorand)
- Faster, simpler

### 3. EIP-712 (Typed Structured Data Signing)
- Human-readable signature requests
- Shows users what they're signing

**Example (EIP-712):**
Instead of signing a hash:
```
Sign: 0x1234abcd...
```

You see:
```
Sign this message:
  From: Alice
  To: Bob
  Amount: 10 ETH
```

This prevents phishing attacks.

---

## Attack Vectors and Protections

### Attack 1: Signature Replay
**Problem:** Attacker reuses a valid signature on a different chain.

**Solution:** Include chain ID in signature (EIP-155).

```typescript
const tx = {
  to: "0xRecipient",
  value: ethers.parseEther("1.0"),
  chainId: 1 // Ethereum Mainnet
};
```

Signature is only valid on the specified chain.

---

### Attack 2: Phishing (Sign Malicious Transactions)
**Problem:** User signs a transaction they don't understand.

**Solution:**
- Use EIP-712 (human-readable signatures)
- Wallets warn about suspicious transactions

**Example:**
```
MetaMask warning: "This transaction will give unlimited approval to spend your tokens."
```

---

### Attack 3: Key Compromise
**Problem:** Attacker steals your private key.

**No solution.** If your key is compromised, your account is compromised.

**Mitigation:**
- Use hardware wallets (Ledger, Trezor)
- Use multisig wallets

---

## Differences: Encryption vs Signing

| Feature | Encryption | Signing |
|---------|-----------|---------|
| **Purpose** | Confidentiality (hide data) | Authentication (prove authorship) |
| **Keys** | Encrypt with public key, decrypt with private key | Sign with private key, verify with public key |
| **Output** | Ciphertext (encrypted data) | Signature (proof) |

**Key insight:** Signing doesn't hide data. Blockchain transactions are public.

---

## Common Mistakes

### Mistake 1: Confusing Signing with Encryption
❌ Signing doesn't encrypt data. Everything signed is public.

### Mistake 2: Signing Without Reading
❌ Always understand what you're signing. Malicious apps can trick you.

### Mistake 3: Reusing Nonces (Advanced)
❌ In ECDSA, if you reuse a nonce when signing two different messages, an attacker can derive your private key.

(This happened to Sony PlayStation 3 in 2010.)

---

## Practical Example: "Sign In With Ethereum"

Many Web3 apps replace traditional login with message signing.

### Flow:
1. **Website generates a challenge:**
   ```
   "Sign this message to prove you own 0x742d35...
    Nonce: 123456
    Timestamp: 2024-01-15T10:00:00Z"
   ```

2. **User signs the message with MetaMask.**

3. **Website verifies the signature:**
   ```typescript
   const recoveredAddress = ethers.verifyMessage(message, signature);
   if (recoveredAddress === "0x742d35...") {
     // User authenticated ✅
   }
   ```

**Benefits:**
- No password to remember
- No password to store (on server)
- No password database to hack

**Tradeoffs:**
- Requires wallet setup
- UX is unfamiliar to Web2 users

---

## Exercise

### 1. Sign a Message

Use ethers.js to sign a message:

```typescript
import { ethers } from 'ethers';

const wallet = ethers.Wallet.createRandom();
const message = "Hello Web3!";
const signature = await wallet.signMessage(message);

console.log("Message:", message);
console.log("Signature:", signature);
console.log("Signer address:", wallet.address);
```

### 2. Verify a Signature

Verify the signature you just created:

```typescript
const recoveredAddress = ethers.verifyMessage(message, signature);
console.log("Recovered address:", recoveredAddress);
console.log("Match?", recoveredAddress === wallet.address);
```

### 3. Tamper with the Message

Change the message slightly and try to verify the signature again. It should fail:

```typescript
const tamperedMessage = "Hello Web3!!"; // One extra character
const recoveredAddress2 = ethers.verifyMessage(tamperedMessage, signature);
console.log("Match after tampering?", recoveredAddress2 === wallet.address); // false
```

---

## Summary

**Digital signatures:**
- Prove you authorized a transaction
- Created by signing with your private key
- Verified using your public key
- Cannot be forged (without the private key)

**In Web3:**
- Every transaction must be signed
- Signatures replace passwords
- No central authority validates signatures (anyone can verify)

**Key insights:**
- You never reveal your private key
- Signatures prove you have the key
- Tampering invalidates signatures

---

## Next Lesson

[→ Wallets and Key Management](04-wallets-and-keys.md)
