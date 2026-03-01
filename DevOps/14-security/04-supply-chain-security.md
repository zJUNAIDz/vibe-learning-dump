# Supply Chain Security

> **Your app is only as secure as its weakest dependency. If one npm package is compromised, every app using it is compromised. Supply chain attacks target the build pipeline, not the running app.**

---

## 🟢 The Problem

```
Your app's dependency tree:

  your-app
    ├─ express (4.18.2) ← you chose this
    │   ├─ body-parser ← express chose this
    │   │   ├─ bytes
    │   │   ├─ content-type
    │   │   ├─ depd
    │   │   └─ raw-body
    │   │       └─ unpipe
    │   ├─ cookie (0.5.0)
    │   ├─ ...47 more packages
    │   └─ send
    │       └─ mime (1.6.0) ← you never heard of this
    ├─ pg (8.11.3)
    │   ├─ pg-connection-string
    │   ├─ pg-pool
    │   └─ pg-protocol
    └─ jsonwebtoken (9.0.0)
        ├─ jws
        │   └─ jwa
        │       └─ ecdsa-sig-formatter
        └─ ...12 more packages

You directly chose 3 packages.
Your node_modules has 200+ packages.
Any ONE of them can be malicious or vulnerable.
```

### Real Supply Chain Attacks

```
event-stream (2018):
  → Popular npm package (2M weekly downloads)
  → Maintainer burned out, handed ownership to stranger
  → New owner added malicious code targeting cryptocurrency wallets
  → Went undetected for 2 months

ua-parser-js (2021):
  → 7M weekly downloads
  → Maintainer's npm account compromised
  → Malicious versions published with crypto miners
  → Affected for ~4 hours before caught

SolarWinds (2020):
  → Build system compromised
  → Malicious code injected during build process
  → Signed and distributed to 18,000+ customers
  → Including US government agencies

Colors.js / Faker.js (2022):
  → Maintainer intentionally sabotaged own packages
  → Added infinite loop printing garbage
  → Affected thousands of projects
```

---

## 🟢 Dependency Vulnerability Scanning

### npm audit

```bash
# Check for known vulnerabilities
npm audit

# Output:
# 3 vulnerabilities (1 low, 1 moderate, 1 high)
#
# xml2js  <0.5.0
# Severity: moderate
# Prototype Pollution - https://github.com/advisories/GHSA-xxx
# fix available via `npm audit fix`

# Auto-fix (safe updates only)
npm audit fix

# Force fix (may include breaking changes)
npm audit fix --force

# CI pipeline: fail on high/critical
npm audit --audit-level=high
```

### Snyk (More Comprehensive)

```bash
# Snyk scans dependencies AND suggests fixes
npx snyk test

# Monitor continuously
npx snyk monitor

# CI integration
# .github/workflows/security.yaml
jobs:
  security:
    steps:
      - uses: actions/checkout@v4
      - name: Run Snyk
        uses: snyk/actions/node@master
        env:
          SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
        with:
          args: --severity-threshold=high
```

### Go Vulnerability Check

```bash
# Built-in Go vulnerability checker
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Output:
# Vulnerability #1: GO-2024-XXXX
#   stdlib: net/http
#   Found in: net/http@go1.21.0
#   Fixed in: net/http@go1.21.8
#   Example trace: main.go:42 → http.ListenAndServe
```

---

## 🟡 SBOM (Software Bill of Materials)

```
An SBOM lists every component in your software:
  → All dependencies (direct and transitive)
  → Their versions
  → Their licenses
  → Their source (npm registry, GitHub, etc.)

Why:
  → When Log4Shell happens, you can instantly check:
    "Do any of our services use Log4j?"
  → Compliance requirements (US Executive Order 14028)
  → License auditing (is any dependency GPL?)

Formats:
  → SPDX (ISO standard)
  → CycloneDX (OWASP standard)
```

### Generating SBOMs

```bash
# Using Syft (popular SBOM generator)
brew install syft

# Generate SBOM from a container image
syft my-app:latest -o cyclonedx-json > sbom.json

# Generate SBOM from source code
syft dir:. -o spdx-json > sbom.json

# Verify SBOM: scan for vulnerabilities
grype sbom:./sbom.json
```

```bash
# Using Trivy (scanner + SBOM)
trivy image --format cyclonedx --output sbom.json my-app:latest

# In CI pipeline
# .github/workflows/build.yaml
jobs:
  build:
    steps:
      - name: Build image
        run: docker build -t my-app:${{ github.sha }} .
      
      - name: Generate SBOM
        run: syft my-app:${{ github.sha }} -o cyclonedx-json > sbom.json
      
      - name: Upload SBOM as artifact
        uses: actions/upload-artifact@v4
        with:
          name: sbom
          path: sbom.json
```

---

## 🟡 Signed Images

```
Problem: How do you know the image you're deploying 
is the one you built? Not tampered with?

Solution: Sign your container images.

Flow:
  1. CI builds image
  2. CI signs image with private key
  3. Registry stores image + signature
  4. Kubernetes verifies signature before running
  5. If signature doesn't match → pod REJECTED
```

### Signing with Cosign

```bash
# Install cosign
brew install cosign

# Generate key pair (one-time)
cosign generate-key-pair
# Creates cosign.key (private) and cosign.pub (public)

# Sign an image
cosign sign --key cosign.key my-registry/my-app:v1.0.0

# Verify an image
cosign verify --key cosign.pub my-registry/my-app:v1.0.0

# Keyless signing (uses OIDC identity — recommended)
# No key management required
cosign sign my-registry/my-app:v1.0.0
# Signs using your GitHub/Google/MSFT identity via Sigstore
```

### CI Pipeline with Signing

```yaml
# .github/workflows/build.yaml
jobs:
  build-and-sign:
    permissions:
      contents: read
      packages: write
      id-token: write  # Required for keyless signing
    steps:
      - uses: actions/checkout@v4
      
      - name: Build and push image
        run: |
          docker build -t ghcr.io/${{ github.repository }}:${{ github.sha }} .
          docker push ghcr.io/${{ github.repository }}:${{ github.sha }}
      
      - name: Install cosign
        uses: sigstore/cosign-installer@v3
      
      - name: Sign image (keyless)
        run: |
          cosign sign ghcr.io/${{ github.repository }}:${{ github.sha }}
        env:
          COSIGN_EXPERIMENTAL: "true"
```

### Enforcing Signed Images in Kubernetes

```yaml
# Using Kyverno policy engine
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: verify-image-signatures
spec:
  validationFailureAction: Enforce  # Block unsigned images
  rules:
    - name: verify-cosign-signature
      match:
        any:
          - resources:
              kinds: ["Pod"]
      verifyImages:
        - imageReferences: ["ghcr.io/myorg/*"]
          attestors:
            - entries:
                - keyless:
                    subject: "https://github.com/myorg/*"
                    issuer: "https://token.actions.githubusercontent.com"
```

---

## 🔴 Lockfile Integrity

```
Lock files (package-lock.json, go.sum) pin exact versions.
If someone modifies the lockfile, they can inject different code.

Always:
  → Commit lockfiles to Git
  → Use npm ci (not npm install) in CI — respects lockfile exactly
  → Review lockfile changes in PRs
  → Verify integrity hashes match
```

```bash
# npm ci vs npm install
npm install  → May update lockfile, install different versions
npm ci       → Installs EXACTLY what's in lockfile, fails if mismatch

# In CI, ALWAYS use:
npm ci

# Go: verify checksums
go mod verify
# All modules verified

# If tampered:
# verifying github.com/lib/pq@v1.10.9: checksum mismatch
```

---

## 🔴 Anti-Patterns

```
❌ No dependency scanning in CI
   → "We'll check manually"
   → You won't. Automate it.

❌ Running npm install in production Dockerfile
   → Use npm ci to ensure reproducible builds
   → Lock exact versions

❌ Not pinning base image versions
   → FROM node:latest ← changes without warning
   → FROM node:20.11.1-slim ← predictable

❌ Using unmaintained packages
   → Check: last commit date, open issues, maintainer count
   → If abandoned for 2+ years, find an alternative

❌ Ignoring npm audit warnings
   → "It says moderate, probably fine"
   → Moderate today, exploit tomorrow

❌ No SBOM for deployed services
   → When a zero-day hits, you need to know in minutes
     which services are affected
   → Generate and store SBOMs for every release
```

---

**Previous:** [03. Container Security](./03-container-security.md)  
**Next:** [05. CI/CD Security](./05-cicd-security.md)
