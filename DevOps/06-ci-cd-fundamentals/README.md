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

```mermaid
graph TD
    A["1. Commit pushed to GitHub"] --> B["2. CI server triggered<br/>(Jenkins, GitHub Actions)"]
    B --> C["3. Checkout code"]
    C --> D["4. Install dependencies"]
    D --> E["5. Run unit tests<br/>(fail fast)"]
    E --> F["6. Run integration tests"]
    F --> G["7. Build Docker image"]
    G --> H["8. Tag image<br/>myapp:v1.2.3"]
    H --> I["9. Push to Docker registry"]
    I --> J["10. Deploy to staging<br/>(automated)"]
    J --> K["11. Run smoke tests"]
    K --> L["12. Deploy to production<br/>(manual approval or automated)"]
    
    style A fill:#bfb,stroke:#333,stroke-width:2px
    style B fill:#ffd,stroke:#333,stroke-width:2px
    style C fill:#ddf,stroke:#333,stroke-width:2px
    style D fill:#ddf,stroke:#333,stroke-width:2px
    style E fill:#bbf,stroke:#333,stroke-width:2px
    style F fill:#bbf,stroke:#333,stroke-width:2px
    style G fill:#fda,stroke:#333,stroke-width:2px
    style H fill:#fda,stroke:#333,stroke-width:2px
    style I fill:#fda,stroke:#333,stroke-width:2px
    style J fill:#fcf,stroke:#333,stroke-width:2px
    style K fill:#bbf,stroke:#333,stroke-width:2px
    style L fill:#f9f,stroke:#333,stroke-width:2px
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
```mermaid
graph TD
    A["v1.2.3"] --> B["Major (1)<br/>Breaking changes"]
    A --> C["Minor (2)<br/>New features,<br/>backward compatible"]
    A --> D["Patch (3)<br/>Bug fixes"]
    
    style A fill:#ddf,stroke:#333,stroke-width:2px,color:#000
    style B fill:#fbb,stroke:#333,stroke-width:2px,color:#000
    style C fill:#bfb,stroke:#333,stroke-width:2px,color:#000
    style D fill:#ffd,stroke:#333,stroke-width:2px,color:#000
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
