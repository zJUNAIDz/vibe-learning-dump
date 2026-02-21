# Wallets and Key Management

🟢 **Fundamentals**

---

## What Is a Wallet? (The Truth)

**Common misconception:** "A wallet holds my cryptocurrency."

**Reality:** Your cryptocurrency lives on the blockchain. A wallet holds your private key.

```
Wallet = Key manager, not a container
```

Think of it like this:
- The blockchain is a bank vault
- Your funds are in the vault
- Your wallet holds the key to your section of the vault

---

## Wallet Software vs Wallet Hardware

### Software Wallets
Programs that manage your private keys.

**Examples:**
- **Browser extensions:** MetaMask, Rabby, Coinbase Wallet
- **Mobile apps:** Trust Wallet, Rainbow, Argent
- **Desktop apps:** Exodus, Atomic Wallet

**Pros:**
- Convenient
- Free
- Easy to use

**Cons:**
- Private keys stored on your device (vulnerable to malware)
- Less secure for large amounts

---

### Hardware Wallets
Physical devices that store private keys offline.

**Examples:**
- Ledger Nano S/X
- Trezor One/Model T

**Pros:**
- Private keys never leave the device
- More secure (even if your computer is compromised)
- Best for large amounts

**Cons:**
- Cost money ($50-$200)
- Less convenient (need to connect device)

---

### Custodial vs Non-Custodial

#### Non-Custodial (You Control the Keys)
- You hold the private key
- Full control
- Full responsibility

**Examples:** MetaMask, Ledger, Trezor

**Tradeoff:**
- ✅ No one can freeze your account
- ❌ If you lose your key, funds are gone forever

---

#### Custodial (Someone Else Controls the Keys)
- Exchange or service holds your keys
- They manage security
- You trust them

**Examples:** Coinbase, Binance, Kraken

**Tradeoff:**
- ✅ If you forget your password, customer support can help
- ❌ Platform can freeze your account
- ❌ "Not your keys, not your coins"

---

## Seed Phrases (Mnemonic Phrases)

A **seed phrase** (also called a mnemonic or recovery phrase) is a human-readable representation of your private key.

### Example Seed Phrase:
```
witch collapse practice feed shame open despair creek road again ice least
```

**12 or 24 words** (randomly selected from a list of 2048 words).

---

### How It Works

```
Seed phrase → Binary seed → Master private key → Multiple accounts
```

**Key insight:** From a single seed phrase, you can derive unlimited accounts (addresses).

---

### BIP-39 Standard
Seed phrases follow the BIP-39 (Bitcoin Improvement Proposal 39) standard.

**Word list:**
- 2048 words
- Carefully chosen to avoid confusion (e.g., no similar-sounding words)

**Example:**
```
abandon, ability, able, about, above, absent, absorb, abstract, absurd...
```

---

### Why Seed Phrases?

**Problem:** Private keys are hard to write down or remember.

```
Private Key: 0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef
```

**Solution:** Convert to words.

```
Seed Phrase: witch collapse practice feed shame open despair creek road again ice least
```

Much easier to write down on paper.

---

## Hierarchical Deterministic Wallets (HD Wallets)

Modern wallets use **BIP-32** and **BIP-44** standards to derive multiple accounts from one seed.

### Derivation Path
```
m / purpose' / coin_type' / account' / change / address_index
```

**Example (Ethereum):**
```
m / 44' / 60' / 0' / 0 / 0  → First Ethereum account
m / 44' / 60' / 0' / 0 / 1  → Second Ethereum account
```

**Key insight:** One seed phrase → unlimited accounts.

---

### Code Example: Deriving Multiple Accounts

```typescript
import { ethers } from 'ethers';

// Create a wallet from a mnemonic
const mnemonic = "witch collapse practice feed shame open despair creek road again ice least";
const hdNode = ethers.HDNodeWallet.fromPhrase(mnemonic);

// Derive multiple accounts
for (let i = 0; i < 3; i++) {
  const path = `m/44'/60'/0'/0/${i}`;
  const wallet = hdNode.derivePath(path);
  console.log(`Account ${i}: ${wallet.address}`);
}
```

**Output:**
```
Account 0: 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
Account 1: 0x5A0b54D5dc17e0AadC383d2db43B0a0D3E029c4c
Account 2: 0x6B175474E89094C44Da98b954EedeAC495271d0F
```

---

## Storing Your Seed Phrase

### ✅ Good Practices
1. **Write it down on paper**
   - Never store digitally (no screenshots, no cloud storage)

2. **Store in a safe place**
   - Fireproof safe, safety deposit box

3. **Consider splitting it**
   - Shamir's Secret Sharing (split into multiple parts)

4. **Never share it**
   - Legitimate services NEVER ask for seed phrases

---

### ❌ Bad Practices
1. **Storing digitally**
   - Screenshot → easy to steal
   - Cloud storage → can be hacked

2. **Emailing or messaging it**
   - Plaintext = insecure

3. **Sharing with "support"**
   - Scammers impersonate support and ask for seed phrases

---

## Key Management Best Practices

### For Small Amounts ($10-$1000)
- Software wallet (MetaMask) is fine
- Backup seed phrase on paper

### For Medium Amounts ($1000-$10,000)
- Consider hardware wallet (Ledger, Trezor)
- Store seed phrase in a safe

### For Large Amounts ($10,000+)
- Hardware wallet (mandatory)
- Consider multisig wallet (require multiple keys)
- Consider Shamir's Secret Sharing (split seed phrase)

---

## Common Wallet Features

### 1. Account Management
- Create new accounts
- Import/export accounts
- Manage multiple chains (Ethereum, Polygon, etc.)

### 2. Transaction Signing
- Sign transactions
- Review transaction details
- Adjust gas fees

### 3. Token Management
- View token balances
- Add custom tokens
- Send/receive tokens

### 4. DApp Connection
- Connect to decentralized apps
- Approve spending limits
- Manage permissions

---

## MetaMask Walkthrough (Most Common Wallet)

### Install MetaMask
1. Browser extension (Chrome, Firefox, Brave)
2. Click "Create a Wallet"
3. Write down seed phrase (12 words)
4. Confirm seed phrase
5. Wallet created ✅

### Add a Network (e.g., Polygon)
1. Open MetaMask
2. Click network dropdown
3. Select "Add Network"
4. Enter network details:
   - Network Name: Polygon
   - RPC URL: https://polygon-rpc.com
   - Chain ID: 137
   - Currency Symbol: MATIC

### Import an Account
1. Click account icon
2. Select "Import Account"
3. Paste private key
4. Account imported ✅

---

## Security Risks

### 1. Phishing Attacks
**Attack:** Fake website looks like MetaMask, steals your seed phrase.

**Protection:**
- Always check the URL
- Never enter seed phrase on a website

---

### 2. Malicious DApps
**Attack:** DApp tricks you into signing a transaction that drains your wallet.

**Protection:**
- Read what you're signing (use EIP-712)
- Use a separate wallet for testing

---

### 3. Malware
**Attack:** Keylogger or malware steals your private key from software wallet.

**Protection:**
- Use hardware wallet for large amounts
- Keep OS and software updated

---

### 4. Social Engineering
**Attack:** Scammer impersonates support, asks for seed phrase.

**Protection:**
- **Never share your seed phrase**
- Support will NEVER ask for it

---

## Advanced: Multisig Wallets

A **multisignature wallet** requires multiple signatures to authorize transactions.

**Example:**
- 2-of-3 multisig
- Requires 2 out of 3 keyholders to sign

**Use cases:**
- Company treasury (prevent single point of failure)
- Shared custody (couples, partnerships)
- Security (even if one key is compromised, funds are safe)

**Popular multisig wallets:**
- Gnosis Safe (most common)
- Multi-sig wallets on Ethereum

---

## Advanced: Social Recovery

Some wallets (e.g., Argent) use **social recovery**:
- Select trusted "guardians" (friends, family)
- If you lose your key, guardians can help you recover

**Tradeoff:**
- Easier recovery
- But you must trust your guardians

---

## Web3 Auth (Sign In With Ethereum)

Wallets are increasingly used for authentication (replacing usernames/passwords).

### Flow:
1. Website: "Connect Wallet"
2. User: Connects MetaMask
3. Website: "Sign this message"
4. User: Signs message
5. Website: Verifies signature → User authenticated

**Benefits:**
- No password to remember
- No password database to hack

**Tradeoffs:**
- Requires wallet setup
- Unfamiliar UX for non-crypto users

---

## Exercise

### 1. Install MetaMask (On a Test Account)
- Create a new wallet
- Write down seed phrase (for testing only)
- Generate a few accounts

### 2. Derive Multiple Accounts from a Seed Phrase

```typescript
import { ethers } from 'ethers';

const mnemonic = "test test test test test test test test test test test junk";
const hdNode = ethers.HDNodeWallet.fromPhrase(mnemonic);

for (let i = 0; i < 5; i++) {
  const wallet = hdNode.derivePath(`m/44'/60'/0'/0/${i}`);
  console.log(`Account ${i}:`, wallet.address);
}
```

### 3. Practice Wallet Security
- Write down a seed phrase on paper
- Store it somewhere safe
- Practice explaining why you should never share it

---

## Summary

**Wallets:**
- Manage private keys (not funds)
- Software wallets: convenient, less secure
- Hardware wallets: secure, less convenient

**Seed phrases:**
- 12-24 words
- Derive unlimited accounts
- Must be backed up securely

**Key management:**
- Never share seed phrases
- Use hardware wallets for large amounts
- Consider multisig for critical accounts

**Key insight:** "Not your keys, not your coins."

---

## Next Lesson

[→ What Crypto Does NOT Protect](05-what-crypto-does-not-protect.md)
