# Module 08: Infrastructure as Code (IaC)

> **Infrastructure should be versioned, tested, and repeatable — just like application code**

---

## What Is IaC?

**Before IaC:**
```
1. Log into AWS console
2. Click "Launch Instance"
3. Choose AMI, instance type, network, security group
4. Click "Launch"
5. Repeat for 10 servers
6. Forget what you clicked
7. Can't reproduce environment
```

**With IaC:**
```bash
terraform apply
# Creates 10 identical servers
# Config is in Git
# Reproducible, auditable, testable
```

---

## Topics Covered

### 📝 01. Why IaC Exists
- Manual provisioning is error-prone
- Infrastructure drift
- Disaster recovery
- Multi-environment management

### 📝 02. Declarative vs Imperative
- Declarative: "I want 3 servers"
- Imperative: "Create 3 servers"
- Idempotency

### 📝 03. State Management
- What is state?
- Why state is scary
- State locking
- Remote state

### 📝 04. IaC Tools Comparison
- Terraform (multi-cloud)
- CloudFormation (AWS-only)
- Pulumi (code-based)
- CDK (code-based)

---

**Previous:** [07. Jenkins](../07-jenkins/)  
**Next:** [09. Terraform](../09-terraform/)
