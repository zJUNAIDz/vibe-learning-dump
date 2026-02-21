# Module 06: CI/CD from First Principles

> **Continuous Integration and Continuous Delivery explained without the buzzwords**

---

## What Is CI/CD? (Really)

**CI (Continuous Integration):**
- Merge code frequently (multiple times per day)
- Automated tests run on every commit
- Catch bugs early (when they're cheap to fix)

**CD (Continuous Delivery):**
- Code is **always** in a deployable state
- Deployment to production is a **business decision**, not a technical challenge

**CD (Continuous Deployment):**
- Every commit that passes tests **automatically** goes to production
- No human approval gate

---

## Topics Covered

### 📝 01. Why CI/CD Exists
- The problem: Manual, error-prone releases
- The solution: Automation and fast feedback
- Mental model: Pipeline as assembly line

### 📝 02. Build vs Test vs Deploy
- Build: Compile, bundle, containerize
- Test: Unit, integration, end-to-end
- Deploy: Push to staging, production

### 📝 03. Immutable Artifacts
- Why you build once, deploy many times
- Docker images as artifacts
- Semantic versioning
- Git tags and releases

### 📝 04. Pipeline Stages
- Checkout code
- Install dependencies
- Run tests
- Build Docker image
- Push to registry
- Deploy to Kubernetes

### 📝 05. Deployment Strategies
- Blue/green deployments
- Canary deployments
- Rolling updates (Kubernetes default)
- Rollbacks vs roll-forwards

### 📝 06. CI/CD Best Practices
- Keep builds fast (<10 minutes)
- Fail fast (run fast tests first)
- Idempotent pipelines
- Secrets management
- Branch strategies (trunk-based, GitFlow)

---

## Example Pipeline (Conceptual)

```
1. Commit pushed to GitHub
   ↓
2. CI server (Jenkins, GitHub Actions) triggered
   ↓
3. Checkout code
   ↓
4. Install dependencies
   ↓
5. Run unit tests (fail fast)
   ↓
6. Run integration tests
   ↓
7. Build Docker image
   ↓
8. Tag image: myapp:v1.2.3
   ↓
9. Push to Docker registry
   ↓
10. Deploy to staging (automated)
    ↓
11. Run smoke tests
    ↓
12. Deploy to production (manual approval or automated)
```

---

## Key Concepts

### Trunk-Based Development
- Everyone commits to `main` (or `trunk`)
- Short-lived feature branches (<1 day)
- Feature flags for incomplete features

### GitFlow (Alternative)
- `main` = production
- `develop` = integration
- Feature branches merge to `develop`
- Release branches for production

### Semantic Versioning
```
v1.2.3
│ │ │
│ │ └─ Patch (bug fixes)
│ └─── Minor (new features, backward compatible)
└───── Major (breaking changes)
```

---

## Common Mistakes

1. **No tests** → CI is useless without tests
2. **Slow pipelines** → 1 hour builds kill productivity
3. **Manual steps** → "Click here, then SSH and run this" defeats the purpose
4. **Deploying to prod without staging** → Recipe for disaster
5. **No rollback plan** → Hope is not a strategy

---

**Previous:** [05. K9s](../05-k9s/)  
**Next:** [07. Jenkins Deep Dive](../07-jenkins/)
