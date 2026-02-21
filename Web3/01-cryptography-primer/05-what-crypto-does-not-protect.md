# What Crypto Does NOT Protect

🟢 **Fundamentals**

---

## The Limits of Cryptography

Cryptography is powerful, but it's not magic. Understanding its limitations is critical.

---

## ❌ Crypto Does NOT Guarantee Confidentiality (on Public Blockchains)

### The Problem
Public blockchains are transparent by default.

**What's public:**
- Your address
- Your balance
- Your transaction history
- Smart contract interactions
- Token holdings

**Example:**
```
Address: 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
Balance: 10 ETH
Recent transactions:
- Sent 1 ETH to 0xABCD...
- Bought 100 UNI tokens
- Interacted with Uniswap
```

Anyone can see this on a block explorer (e.g., Etherscan).

### What This Means
- Financial privacy is lost
- Your spending habits are public
- Addresses can be linked to identities (if you've used an exchange)

---

## ❌ Crypto Does NOT Prevent Phishing

### The Problem
If you sign a malicious transaction, crypto can't save you.

**Scenario:**
1. You visit fake-uniswap.com (looks like real Uniswap)
2. Website asks you to "approve token spending"
3. You sign the transaction
4. The malicious contract drains your wallet

**Why crypto doesn't help:**
- Your private key signed the transaction
- The signature is valid
- The blockchain executes it as programmed

---

### Phishing Example

**Real site:** uniswap.org  
**Fake site:** un1swap.org (notice the "1" instead of "i")

Fake site asks you to:
```
Approve unlimited USDC spending to contract 0xMALICIOUS
```

If you sign this, the attacker can steal all your USDC.

---

### Protection
- Always verify URLs
- Use bookmarks for important sites
- Read transaction details carefully
- Use hardware wallets (they show transaction details on the device)

---

## ❌ Crypto Does NOT Protect Against Smart Contract Bugs

### The Problem
Smart contracts are code. Code has bugs.

**If a contract has a bug:**
- Crypto can't prevent exploitation
- Funds can be stolen or locked forever
- Immutability means the bug is permanent

**Example: The DAO Hack (2016)**
- Smart contract had a reentrancy bug
- Attacker drained $60M
- Bug was in the contract logic, not the cryptography

---

### Why This Happens
- Smart contracts are complex
- Audits help, but aren't perfect
- New attack vectors are discovered constantly

---

### Protection
- Use audited contracts (OpenZeppelin, etc.)
- Understand what a contract does before interacting
- Never invest more than you can afford to lose

---

## ❌ Crypto Does NOT Prevent Social Engineering

### The Problem
If an attacker tricks you into revealing your seed phrase, crypto can't protect you.

**Common scams:**
1. **Fake support:** "We're from MetaMask support. Send us your seed phrase to recover your account."
2. **Fake giveaways:** "Send 1 ETH, get 2 ETH back!"
3. **Impersonation:** Scammer pretends to be a friend asking for help.

**Why crypto doesn't help:**
- Your seed phrase = full access
- No password reset
- No chargeback

---

### Protection
- **Never share your seed phrase**
- Support will NEVER ask for it
- Be skeptical of "too good to be true" offers

---

## ❌ Crypto Does NOT Protect Against Key Loss

### The Problem
If you lose your private key or seed phrase, your funds are gone forever.

**No customer support. No password reset.**

**Statistics:**
- ~20% of all Bitcoin is estimated to be permanently lost (lost keys)

---

### Common Ways Keys Are Lost
1. **Device failure:** Laptop crashes, seed phrase was only stored there
2. **Forgetting:** Seed phrase written down, but lost
3. **Death:** User dies, heirs don't have access
4. **Theft:** Physical seed phrase stolen
5. **Catastrophe:** House fire destroys paper backup

---

### Protection
- Backup seed phrase on paper
- Store in multiple secure locations
- Consider metal backups (fireproof)
- Share access plan with trusted person (for inheritance)

---

## ❌ Crypto Does NOT Prevent Front-Running

### The Problem
In public blockchains, transactions sit in a mempool before being included in a block.

**Attackers can see your transaction and submit a competing transaction with higher gas to get processed first.**

**Example:**
1. You submit a transaction to buy Token X for $100
2. Bot sees your transaction in mempool
3. Bot submits a transaction with higher gas to buy Token X first
4. Price goes up
5. Your transaction executes at worse price
6. Bot sells Token X for profit

This is called **MEV (Maximal Extractable Value)**.

---

### Protection
- Use private transaction services (Flashbots)
- Accept that some front-running is unavoidable
- Design protocols to be front-run resistant

---

## ❌ Crypto Does NOT Guarantee Anonymity

### The Problem
Blockchains are pseudonymous, not anonymous.

**Pseudonymous:**
- Your identity is an address (0x742d35...)
- Not directly linked to your real name

**But:**
- If you've used a centralized exchange (Coinbase, Binance), they know your identity
- Blockchain analysis firms (Chainalysis) can link addresses
- Transaction patterns can be traced

---

### How You Can Be Traced
1. **Exchange KYC:**
   - You buy ETH on Coinbase (KYC verified)
   - Coinbase knows your address
   - Any transaction from that address is linked to you

2. **IP address:**
   - Your IP is visible when broadcasting transactions
   - Correlate addresses with IPs

3. **On-chain patterns:**
   - If you interact with known services, patterns emerge

---

### Privacy Tools (with caveats)
- **Mixers (e.g., Tornado Cash):** Mix your coins with others to break traceability
  - **Risk:** Heavily regulated. Tornado Cash was sanctioned by US Treasury.
  
- **Privacy coins (Monero, Zcash):** Built-in privacy
  - **Risk:** Harder to use, less liquidity, regulatory scrutiny

---

## ❌ Crypto Does NOT Protect Against Malware

### The Problem
If your device is compromised, malware can:
- Steal your private key
- Replace addresses you're sending to (clipboard hijacking)
- Sign transactions on your behalf

---

### Example: Clipboard Hijacking
1. You copy an Ethereum address: `0x742d35...`
2. Malware detects the copy
3. Malware replaces it with attacker's address: `0xATTACKER...`
4. You paste and send funds to the attacker

---

### Protection
- Use hardware wallets (private keys never touch your computer)
- Keep OS and software updated
- Use antivirus
- Verify addresses carefully (check first and last characters)

---

## ❌ Crypto Does NOT Prevent Mistakes

### The Problem
Irreversibility means mistakes are permanent.

**Common mistakes:**
- Sending to the wrong address (funds lost forever)
- Sending to a contract address (if it doesn't support receiving, funds locked)
- Sending on wrong network (e.g., sending ETH on BSC instead of Ethereum)
- Paying too little gas (transaction fails, fee still charged)

---

### Example: Sending to Wrong Address
```
You want to send 10 ETH to:
  0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb

But you accidentally send to:
  0x742d35Cc6634C0532925a3b844Bc9e7595f0bEa  (last character wrong)

Funds are gone forever.
```

---

### Protection
- Double-check addresses (compare first and last characters)
- Send a small test transaction first
- Use ENS (Ethereum Name Service) for human-readable names

---

## ❌ Crypto Does NOT Stop Bad Actors

### The Problem
Permissionless systems allow anyone to participate, including scammers.

**Common scams:**
- Rug pulls (developers abandon project, steal funds)
- Ponzi schemes (unsustainable "yield farming")
- Fake tokens (scammers create fake versions of popular tokens)
- Honeypot contracts (you can buy but can't sell)

---

### Why This Happens
- No KYC (anyone can deploy contracts)
- Irreversibility (stolen funds can't be recovered)
- Lack of regulation

---

### Protection
- Do your own research (DYOR)
- Use trusted platforms
- Be skeptical of high yields ("If it sounds too good to be true...")
- Check contract verification on Etherscan

---

## ❌ Crypto Does NOT Replace Legal Systems

### The Problem
"Code is law" sounds nice, but the real world has actual laws.

**If a contract violates laws:**
- Regulators can shut down frontends
- Developers can be arrested (see Tornado Cash)
- Your funds can be seized (if held on a centralized exchange)

**If you're scammed:**
- No regulatory recourse
- No insurance (unless you use a custodial service)
- You can't sue a smart contract

---

## Summary: What Crypto DOES and DOES NOT Do

| Crypto DOES | Crypto DOES NOT |
|-------------|-----------------|
| ✅ Prove authorship (signatures) | ❌ Prevent phishing |
| ✅ Ensure data integrity (hashes) | ❌ Protect against bugs |
| ✅ Enable trustless verification | ❌ Prevent social engineering |
| ✅ Secure key ownership | ❌ Prevent key loss |
| ✅ Authenticate transactions | ❌ Prevent front-running |
| ✅ Create pseudonymous identities | ❌ Guarantee anonymity |
| ✅ Resist censorship (sometimes) | ❌ Prevent malware |
| | ❌ Prevent user mistakes |
| | ❌ Stop bad actors |
| | ❌ Replace legal systems |

---

## The Key Insight

**Cryptography solves technical problems, not human problems.**

- If you sign a malicious transaction → crypto executes it faithfully
- If you lose your seed phrase → crypto can't recover it
- If you're tricked by a scammer → crypto doesn't care

**Your responsibility:**
- Verify what you're signing
- Secure your keys
- Be skeptical
- Understand what you're doing

---

## Exercise

For each scenario, identify the limitation:

### 1. Alice sends ETH to a wrong address by mistake.
**Limitation:** Crypto does not prevent user mistakes.

### 2. Bob is tricked into sharing his seed phrase with a fake support agent.
**Limitation:** Crypto does not prevent social engineering.

### 3. Carol interacts with a smart contract that has a bug, and her funds are locked.
**Limitation:** Crypto does not protect against smart contract bugs.

### 4. Dave's transactions are publicly visible, and someone tracks his spending habits.
**Limitation:** Crypto does not guarantee confidentiality (on public blockchains).

### 5. Eve's laptop is infected with malware that steals her private key.
**Limitation:** Crypto does not protect against malware.

---

## Module Complete!

You've finished **Module 01: Cryptography Primer**.

**You should now understand:**
- ✅ Hash functions (integrity, immutability)
- ✅ Public/private key cryptography (identity, authentication)
- ✅ Digital signatures (authorization, verification)
- ✅ Wallets (key management, not fund storage)
- ✅ Limitations of cryptography

---

## Next Module

[→ Module 02: Blockchain Data Structures](../02-blockchain-data-structures/)

Learn how blockchains are structured (blocks, chains, Merkle trees) and why immutability matters.
