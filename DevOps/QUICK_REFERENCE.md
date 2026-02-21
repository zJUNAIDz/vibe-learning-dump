# DevOps Curriculum - Quick Reference

> **Fast lookup for commands, concepts, and workflows**

---

## Linux & Systems

### Process Management
```bash
# View processes
ps aux
htop

# Kill process
kill -SIGTERM <PID>  # Graceful
kill -9 <PID>        # Force

# Check process memory
ps aux --sort=-%mem | head -10

# Inspect process
cat /proc/<PID>/status
ls -la /proc/<PID>/fd/
```

### Memory
```bash
# System memory
free -h

# Per-process
pmap <PID>

# Check for OOM kills
dmesg | grep -i oom
```

### Networking
```bash
# Check listening ports
ss -tuln
lsof -i :8080

# Test connectivity
nc -zv google.com 443
curl -v http://example.com

# DNS lookup
dig google.com
nslookup google.com

# Capture packets
sudo tcpdump -i any port 80
```

### systemd
```bash
# Service management
sudo systemctl start myapp
sudo systemctl stop myapp
sudo systemctl restart myapp
sudo systemctl status myapp

# Enable auto-start
sudo systemctl enable myapp

# View logs
sudo journalctl -u myapp -f
sudo journalctl -u myapp -n 100
```

---

## Docker

### Image Management
```bash
# Build
docker build -t myapp:v1.0 .
docker build --build-arg NODE_VERSION=18 -t myapp .

# Push/Pull
docker push myapp:v1.0
docker pull nginx:alpine

# List/Remove
docker images
docker rmi myapp:v1.0
docker image prune  # Remove dangling images

# Inspect
docker history myapp:v1.0
docker inspect myapp:v1.0
```

### Container Management
```bash
# Run
docker run -d --name myapp -p 8080:80 nginx
docker run -it --rm alpine sh

# List
docker ps        # Running
docker ps -a     # All

# Logs
docker logs myapp
docker logs -f myapp

# Exec
docker exec -it myapp /bin/sh

# Stop/Remove
docker stop myapp
docker rm myapp

# Resource limits
docker run --memory=512m --cpus=1 myapp
```

### Cleanup
```bash
docker system df           # Show disk usage
docker system prune        # Remove unused data
docker system prune -a     # Remove all unused images
docker volume prune        # Remove unused volumes
```

---

## Kubernetes

### Cluster Info
```bash
kubectl cluster-info
kubectl get nodes
kubectl get namespaces
kubectl config get-contexts
kubectl config use-context <context>
```

### Working with Resources
```bash
# Get
kubectl get pods
kubectl get pods -o wide
kubectl get pods -n kube-system
kubectl get all

# Describe
kubectl describe pod <name>
kubectl describe service <name>

# Create/Apply
kubectl apply -f deployment.yaml
kubectl apply -f .  # All YAML in directory

# Delete
kubectl delete pod <name>
kubectl delete -f deployment.yaml

# Edit live
kubectl edit deployment <name>
```

### Pods
```bash
# Logs
kubectl logs <pod>
kubectl logs <pod> -c <container>  # Multi-container pod
kubectl logs <pod> -f              # Follow
kubectl logs <pod> --previous      # Previous crashed container

# Exec
kubectl exec -it <pod> -- /bin/sh
kubectl exec <pod> -- ls /app

# Port forward
kubectl port-forward <pod> 8080:80

# Copy files
kubectl cp <pod>:/path/to/file ./local-file
kubectl cp ./local-file <pod>:/path/to/file
```

### Deployments
```bash
# Create
kubectl create deployment nginx --image=nginx:alpine

# Scale
kubectl scale deployment nginx --replicas=5

# Rollout
kubectl rollout status deployment/nginx
kubectl rollout history deployment/nginx
kubectl rollout undo deployment/nginx
kubectl rollout restart deployment/nginx
```

### Services
```bash
# Expose deployment
kubectl expose deployment nginx --port=80 --type=LoadBalancer

# Get endpoints
kubectl get endpoints <service>
```

### Debugging
```bash
# Events
kubectl get events --sort-by='.lastTimestamp'

# Resource usage
kubectl top nodes
kubectl top pods

# Troubleshoot networking
kubectl run debug --rm -it --image=nicolaka/netshoot -- /bin/bash
```

### ConfigMaps & Secrets
```bash
# Create ConfigMap
kubectl create configmap myconfig --from-file=config.txt
kubectl create configmap myconfig --from-literal=key=value

# Create Secret
kubectl create secret generic mysecret --from-literal=password=secret123
kubectl create secret docker-registry regcred \
  --docker-server=ghcr.io \
  --docker-username=user \
  --docker-password=token

# View
kubectl get configmap myconfig -o yaml
kubectl get secret mysecret -o yaml
```

---

## CI/CD

### Jenkins Pipeline (Declarative)
```groovy
pipeline {
    agent any
    
    environment {
        DOCKER_REGISTRY = 'ghcr.io'
        IMAGE_NAME = "${DOCKER_REGISTRY}/myorg/myapp"
    }
    
    stages {
        stage('Build') {
            steps {
                sh 'npm install'
                sh 'npm run build'
            }
        }
        
        stage('Test') {
            steps {
                sh 'npm test'
            }
        }
        
        stage('Docker Build') {
            steps {
                sh "docker build -t ${IMAGE_NAME}:${BUILD_NUMBER} ."
            }
        }
        
        stage('Push') {
            steps {
                withCredentials([usernamePassword(credentialsId: 'docker-creds', usernameVariable: 'USER', passwordVariable: 'PASS')]) {
                    sh "echo $PASS | docker login ${DOCKER_REGISTRY} -u $USER --password-stdin"
                    sh "docker push ${IMAGE_NAME}:${BUILD_NUMBER}"
                }
            }
        }
        
        stage('Deploy') {
            steps {
                sh "kubectl set image deployment/myapp myapp=${IMAGE_NAME}:${BUILD_NUMBER}"
            }
        }
    }
    
    post {
        always {
            cleanWs()
        }
    }
}
```

---

## Terraform

### Basic Workflow
```bash
# Initialize
terraform init

# Plan
terraform plan
terraform plan -out=plan.tfplan

# Apply
terraform apply
terraform apply plan.tfplan

# Destroy
terraform destroy

# Format
terraform fmt

# Validate
terraform validate

# Show
terraform show
terraform state list
terraform state show <resource>
```

### Example: AWS EC2
```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

resource "aws_instance" "web" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t3.micro"
  
  tags = {
    Name = "WebServer"
  }
}

output "public_ip" {
  value = aws_instance.web.public_ip
}
```

---

## Ansible

### Basic Commands
```bash
# Ping hosts
ansible all -i inventory.ini -m ping

# Run command
ansible all -i inventory.ini -a "uptime"

# Run playbook
ansible-playbook -i inventory.ini playbook.yml

# Check syntax
ansible-playbook playbook.yml --syntax-check

# Dry run
ansible-playbook playbook.yml --check

# Verbose
ansible-playbook playbook.yml -vvv
```

### Example Playbook
```yaml
---
- name: Deploy web application
  hosts: webservers
  become: yes
  
  tasks:
    - name: Install nginx
      package:
        name: nginx
        state: present
    
    - name: Copy config
      copy:
        src: nginx.conf
        dest: /etc/nginx/nginx.conf
      notify: Restart nginx
    
    - name: Start nginx
      service:
        name: nginx
        state: started
        enabled: yes
  
  handlers:
    - name: Restart nginx
      service:
        name: nginx
        state: restarted
```

---

## Makefile

### Example Makefile
```makefile
.PHONY: help build test deploy clean

# Default target
help:
	@echo "Available targets:"
	@echo "  build   - Build Docker image"
	@echo "  test    - Run tests"
	@echo "  deploy  - Deploy to Kubernetes"
	@echo "  clean   - Clean up"

# Variables
IMAGE := myapp:latest
REGISTRY := ghcr.io/myorg

build:
	docker build -t $(IMAGE) .

test:
	npm test

push: build
	docker tag $(IMAGE) $(REGISTRY)/$(IMAGE)
	docker push $(REGISTRY)/$(IMAGE)

deploy: push
	kubectl set image deployment/myapp myapp=$(REGISTRY)/$(IMAGE)

clean:
	docker rmi $(IMAGE) || true
	rm -rf dist/ node_modules/

dev:
	npm run dev

# Run tests on file change
watch:
	npm run test:watch
```

---

## Git

### Common Workflows
```bash
# Clone
git clone https://github.com/user/repo.git

# Branches
git checkout -b feature/new-feature
git branch -a
git branch -d feature/old-feature

# Commit
git add .
git commit -m "feat: add new feature"
git push origin feature/new-feature

# Merge
git checkout main
git pull origin main
git merge feature/new-feature
git push origin main

# Rebase
git checkout feature/my-branch
git rebase main

# Stash
git stash
git stash pop
git stash list

# Tags
git tag v1.0.0
git push origin v1.0.0
```

---

## Observability

### Prometheus Queries
```promql
# CPU usage
rate(container_cpu_usage_seconds_total[5m])

# Memory usage
container_memory_usage_bytes

# HTTP request rate
rate(http_requests_total[5m])

# Error rate
rate(http_requests_total{status=~"5.."}[5m])

# P95 latency
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

### Logging
```bash
# Follow logs (Docker)
docker logs -f myapp

# Follow logs (Kubernetes)
kubectl logs -f deployment/myapp

# Grep logs
kubectl logs deployment/myapp | grep ERROR

# Multiple containers
kubectl logs -f deployment/myapp -c sidecar
```

---

## Security

### Image Scanning
```bash
# Trivy
trivy image myapp:latest
trivy image --severity HIGH,CRITICAL myapp:latest

# Docker Scout
docker scout cves myapp:latest
docker scout recommendations myapp:latest
```

### Secrets Management
```bash
# Create secret from file
kubectl create secret generic db-creds --from-file=./credentials.json

# Create secret from literal
kubectl create secret generic api-key --from-literal=key=abc123

# Encode/decode (base64)
echo -n 'mypassword' | base64
echo 'bXlwYXNzd29yZA==' | base64 -d
```

---

## Common Issues

### Container Won't Start
```bash
# Check logs
docker logs <container>

# Check events
kubectl describe pod <pod>

# Common causes:
# - Image pull error (wrong image name, auth issue)
# - CrashLoopBackOff (app exits immediately)
# - Resource limits (OOMKilled)
```

### Networking Issues
```bash
# Test DNS
kubectl run debug --rm -it --image=alpine -- sh
nslookup kubernetes.default

# Test connectivity
kubectl run debug --rm -it --image=nicolaka/netshoot -- sh
curl http://service-name

# Check service endpoints
kubectl get endpoints <service>
```

### Performance Issues
```bash
# Check resource usage
kubectl top nodes
kubectl top pods

# Check limits
kubectl describe pod <pod> | grep -A 5 Limits

# Check metrics
kubectl get --raw /apis/metrics.k8s.io/v1beta1/nodes
```

---

## Keyboard Shortcuts (kubectl)

```bash
# Aliases (add to ~/.bashrc or ~/.zshrc)
alias k=kubectl
alias kgp='kubectl get pods'
alias kgd='kubectl get deployments'
alias kgs='kubectl get services'
alias kdp='kubectl describe pod'
alias kl='kubectl logs'
alias kex='kubectl exec -it'
```

---

**This is your cheat sheet. Bookmark it, print it, tattoo it on your arm.**
