# Module 09: Terraform

> **Terraform is the lingua franca of infrastructure as code**

---

## What Is Terraform?

Terraform is a **declarative IaC tool** that provisions infrastructure across multiple cloud providers.

---

## Topics Covered

### 📝 01. Terraform Fundamentals
- Providers (AWS, GCP, Azure, Kubernetes, etc.)
- Resources
- Data sources
- Variables and outputs

### 📝 02. State Deep Dive
- What's in the state file?
- Local vs remote state
- State locking (S3 + DynamoDB)
- State security

### 📝 03. Modules
- Creating reusable modules
- Module inputs and outputs
- Public module registry

### 📝 04. Workspaces and Environments
- Dev, staging, production
- Workspace strategies
- Variable files (terraform.tfvars)

### 📝 05. Best Practices
- Directory structure
- Naming conventions
- When NOT to use Terraform
- Terraform + Kubernetes (when it makes sense)

### 📝 06. Common Pitfalls
- State file corruption
- Drift detection
- Resource dependencies
- Blast radius (destroying everything accidentally)

---

**Previous:** [08. IaC Fundamentals](../08-iac-fundamentals/)  
**Next:** [10. Ansible](../10-ansible/)
