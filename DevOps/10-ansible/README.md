# Module 10: Ansible

> **Configuration management for servers — agentless, idempotent, and (mostly) simple**

---

## Ansible vs Terraform

| Terraform | Ansible |
|-----------|----------|
| Provisions infrastructure | Configures servers |
| "Create a server" | "Install nginx on the server" |
| Immutable infrastructure | Mutable infrastructure |
| State-based | Push-based |

**Often used together:**
1. Terraform creates servers
2. Ansible configures them

---

## Topics Covered

### 📝 01. Ansible Basics
- Inventory files
- Playbooks
- Tasks
- Handlers

### 📝 02. Modules
- Package management (dnf, apt, yum)
- File operations
- Service management
- Command execution

### 📝 03. Roles
- Organizing playbooks
- Role structure
- Galaxy (role repository)

### 📝 04. Variables and Templates
- Host variables
- Group variables
- Jinja2 templates

### 📝 05. Idempotency
- What it means
- Why it matters
- Writing idempotent tasks

### 📝 06. Ansible + Terraform
- Hybrid workflows
- When to use which
- Common patterns

---

**Previous:** [09. Terraform](../09-terraform/)  
**Next:** [11. Makefile for Developers](../11-makefile/)
