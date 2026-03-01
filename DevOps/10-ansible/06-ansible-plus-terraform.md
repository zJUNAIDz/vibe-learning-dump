# Ansible + Terraform

> **Terraform creates the infrastructure. Ansible configures what's running on it. They aren't competitors — they're partners. Use the right tool for each job.**

---

## 🟢 Why Both?

```
Terraform says: "Create an EC2 instance with this AMI, in this VPC, with this security group."
  → Infrastructure exists.

Ansible says:  "On that EC2 instance, install nginx, deploy the app, configure monitoring."
  → Infrastructure is configured and running.

                    Terraform                    Ansible
                    ─────────                    ───────
Creates:            VPCs, subnets, VMs,         Nothing (uses existing infra)
                    load balancers, DNS,
                    databases, S3 buckets

Configures:         Basic (user_data scripts)    Full OS configuration,
                                                 packages, services, users,
                                                 files, cron jobs

State:              terraform.tfstate             No persistent state
                    (tracks all resources)        (checks live system)

Language:           HCL                           YAML + Jinja2

Idempotent:         Yes                           Yes

Agent needed:       No (API calls)                No (SSH)
```

---

## 🟢 The Handoff Pattern

```
┌─────────────────────────────────────────────────┐
│                  Workflow                         │
│                                                   │
│  1. terraform apply                               │
│     ├── Create VPC                                │
│     ├── Create subnets                            │
│     ├── Create security groups                    │
│     ├── Create EC2 instances                      │
│     ├── Create RDS database                       │
│     └── Create load balancer                      │
│                                                   │
│  2. Terraform outputs IPs/endpoints               │
│     ├── web_server_ips = ["10.0.1.10", ...]      │
│     ├── db_endpoint = "mydb.abc.rds.amazonaws..." │
│     └── lb_dns = "my-lb-123.elb.amazonaws.com"   │
│                                                   │
│  3. ansible-playbook site.yml                     │
│     ├── Install packages on web servers           │
│     ├── Deploy application                        │
│     ├── Configure nginx with LB settings          │
│     ├── Set up monitoring agents                  │
│     └── Configure log shipping                    │
│                                                   │
└─────────────────────────────────────────────────┘
```

---

## 🟡 Dynamic Inventory from Terraform

### Option 1: Terraform Output → Ansible Inventory

```hcl
# terraform/main.tf
resource "aws_instance" "web" {
  count         = 3
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t3.medium"
  key_name      = var.ssh_key_name
  
  tags = {
    Name  = "web-${count.index + 1}"
    Role  = "webserver"
    Env   = var.environment
  }
}

resource "aws_instance" "db" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "r5.large"
  key_name      = var.ssh_key_name
  
  tags = {
    Name = "db-1"
    Role = "database"
    Env  = var.environment
  }
}

# Output for Ansible
output "web_ips" {
  value = aws_instance.web[*].public_ip
}

output "db_ip" {
  value = aws_instance.db.private_ip
}

output "db_endpoint" {
  value = aws_db_instance.main.endpoint
}
```

### Generate Inventory Script

```bash
#!/bin/bash
# scripts/generate-inventory.sh

cd terraform/

# Get Terraform outputs as JSON
WEB_IPS=$(terraform output -json web_ips | jq -r '.[]')
DB_IP=$(terraform output -raw db_ip)
DB_ENDPOINT=$(terraform output -raw db_endpoint)

# Generate Ansible inventory
cat > ../ansible/inventory/hosts.yml << EOF
all:
  children:
    webservers:
      hosts:
$(for ip in $WEB_IPS; do echo "        $ip:"; done)
      vars:
        db_endpoint: ${DB_ENDPOINT}
    dbservers:
      hosts:
        ${DB_IP}:
EOF

echo "Inventory generated!"
```

### Option 2: AWS Dynamic Inventory Plugin

```yaml
# inventory/aws_ec2.yml
---
plugin: amazon.aws.aws_ec2

regions:
  - us-east-1

filters:
  tag:Env: production
  instance-state-name: running

keyed_groups:
  # Group by Role tag
  - key: tags.Role
    prefix: role
    # Creates groups: role_webserver, role_database

  # Group by environment
  - key: tags.Env
    prefix: env

hostnames:
  - private-ip-address

compose:
  ansible_host: private_ip_address
  ansible_user: "'ubuntu'"
```

```bash
# Test dynamic inventory
ansible-inventory -i inventory/aws_ec2.yml --list
ansible-inventory -i inventory/aws_ec2.yml --graph
```

---

## 🟡 Integrated Workflow (Makefile)

```makefile
# Makefile — single entry point for infra + config

ENVIRONMENT ?= staging
ANSIBLE_DIR = ansible
TERRAFORM_DIR = terraform/$(ENVIRONMENT)

.PHONY: all infra configure deploy destroy

# Full setup: infrastructure + configuration + deployment
all: infra configure deploy

# Step 1: Create infrastructure
infra:
	cd $(TERRAFORM_DIR) && \
		terraform init && \
		terraform plan -out=tfplan && \
		terraform apply tfplan
	$(MAKE) generate-inventory

# Step 2: Configure servers
configure:
	cd $(ANSIBLE_DIR) && \
		ansible-playbook -i inventory/$(ENVIRONMENT)/hosts.yml \
			playbooks/site.yml

# Step 3: Deploy application
deploy:
	cd $(ANSIBLE_DIR) && \
		ansible-playbook -i inventory/$(ENVIRONMENT)/hosts.yml \
			playbooks/deploy.yml \
			-e "app_version=$(VERSION)"

# Generate Ansible inventory from Terraform outputs
generate-inventory:
	./scripts/generate-inventory.sh $(ENVIRONMENT)

# Destroy everything
destroy:
	cd $(TERRAFORM_DIR) && terraform destroy -auto-approve

# Configuration only (servers already exist)
reconfigure:
	cd $(ANSIBLE_DIR) && \
		ansible-playbook -i inventory/$(ENVIRONMENT)/hosts.yml \
			playbooks/site.yml \
			--diff

# Deploy specific version
release:
	@test -n "$(VERSION)" || (echo "VERSION required: make release VERSION=1.2.3" && exit 1)
	$(MAKE) deploy VERSION=$(VERSION)
```

### Usage

```bash
# Full setup
make all ENVIRONMENT=production

# Just deploy new version
make release VERSION=1.2.3 ENVIRONMENT=production

# Reconfigure without touching infrastructure
make reconfigure ENVIRONMENT=staging

# Teardown
make destroy ENVIRONMENT=staging
```

---

## 🟡 When to Use Which

```
Use TERRAFORM for:                    Use ANSIBLE for:
─────────────────                     ────────────────
✅ Cloud resources (AWS/GCP/Azure)    ✅ OS-level configuration
✅ VPCs, subnets, security groups     ✅ Package installation
✅ Managed services (RDS, S3, SQS)    ✅ Service configuration
✅ DNS records                        ✅ User management
✅ Load balancers                     ✅ File deployment
✅ IAM roles and policies             ✅ Application deployment
✅ Kubernetes clusters (EKS/GKE)      ✅ Monitoring agent setup
                                      ✅ Cron jobs
                                      ✅ Security hardening

Don't use:                            Don't use:
❌ Terraform for installing nginx     ❌ Ansible to create VPCs
❌ Terraform for deploying code       ❌ Ansible to manage S3 buckets
❌ Terraform remote-exec (limited)    ❌ Ansible for cloud resource lifecycle
```

### Decision Flow

```
"I need to create a cloud resource"
  → Terraform

"I need to configure what runs ON a resource"
  → Ansible

"I need to deploy my application code"
  → Ansible (or CI/CD tool directly)

"I need to manage Kubernetes resources"
  → kubectl / Helm (not Terraform or Ansible)
  
"I need a one-time script to fix something"
  → Ansible ad-hoc command
```

---

## 🔴 Common Mistakes

### ❌ Using Terraform for everything

```hcl
# BAD — Terraform is not a config management tool
resource "null_resource" "configure_server" {
  provisioner "remote-exec" {
    inline = [
      "apt-get update",
      "apt-get install -y nginx",
      "systemctl start nginx",
      # 100 more lines...
    ]
  }
}
# Problems: Not idempotent, hard to test, limited error handling
```

### ❌ Using Ansible for everything

```yaml
# BAD — Ansible managing cloud resources
- name: Create VPC
  amazon.aws.ec2_vpc_net:
    name: my-vpc
    cidr_block: 10.0.0.0/16
    
- name: Create subnet
  amazon.aws.ec2_vpc_subnet:
    vpc_id: "{{ vpc.id }}"
    cidr: 10.0.1.0/24

# Problems: No state file, can create duplicates, hard to destroy cleanly
```

### ❌ Terraform user_data for complex setup

```hcl
# BAD — complex config in user_data
resource "aws_instance" "web" {
  user_data = <<-EOF
    #!/bin/bash
    apt-get update
    apt-get install -y nginx nodejs postgresql-client
    # Configure nginx (50 lines)
    # Deploy app (30 lines)
    # Setup monitoring (40 lines)
    # No error handling, no idempotency
  EOF
}

# GOOD — minimal user_data, Ansible handles the rest
resource "aws_instance" "web" {
  user_data = <<-EOF
    #!/bin/bash
    # Just ensure SSH is ready for Ansible
    apt-get update && apt-get install -y python3
  EOF
}
# Then: ansible-playbook -i inventory/ site.yml
```

---

## 🔴 Real-World Project Structure

```
project/
├── terraform/
│   ├── modules/
│   │   ├── vpc/
│   │   ├── compute/
│   │   ├── database/
│   │   └── networking/
│   ├── environments/
│   │   ├── staging/
│   │   │   ├── main.tf
│   │   │   ├── variables.tf
│   │   │   └── terraform.tfvars
│   │   └── production/
│   │       ├── main.tf
│   │       ├── variables.tf
│   │       └── terraform.tfvars
│   └── outputs.tf
│
├── ansible/
│   ├── roles/
│   │   ├── common/
│   │   ├── nginx/
│   │   ├── myapp/
│   │   └── monitoring/
│   ├── inventory/
│   │   ├── staging/
│   │   └── production/
│   ├── playbooks/
│   │   ├── site.yml
│   │   └── deploy.yml
│   ├── group_vars/
│   └── ansible.cfg
│
├── scripts/
│   ├── generate-inventory.sh
│   └── bootstrap.sh
│
├── Makefile
└── README.md
```

---

**Previous:** [05. Idempotency](./05-idempotency.md)  
**Next:** [Module 11: Makefile](../11-makefile/README.md)
