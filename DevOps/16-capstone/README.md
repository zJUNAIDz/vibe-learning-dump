# Module 16: Capstone - From Laptop to Production

> **Put it all together: A real service, from code to production**

---

## The Project

Build and deploy a production-ready microservice:
- TypeScript or Go
- Containerized (Docker)
- CI pipeline (Jenkins or GitHub Actions)
- Deployed to Kubernetes
- Infrastructure as Code (Terraform)
- Observability (metrics, logs)
- Secure (secrets, RBAC, image scanning)

---

## Phases

### 📝 01. Application
- Build a simple REST API (TypeScript or Go)
- Health endpoints
- Structured logging
- Prometheus metrics endpoint

### 📝 02. Containerization
- Multi-stage Dockerfile
- Optimized image (<100 MB)
- Non-root user
- Security scanning

### 📝 03. Kubernetes Manifests
- Deployment with resource limits
- Service (ClusterIP)
- Ingress (with TLS)
- ConfigMap and Secret
- Health checks

### 📝 04. CI/CD Pipeline
- Automated tests
- Docker build and push
- Deploy to staging
- Deploy to production (with approval)

### 📝 05. Infrastructure as Code
- Terraform for cloud resources
- Version controlled
- Multiple environments

### 📝 06. Observability
- Prometheus metrics
- Grafana dashboard
- Structured logs
- Alerts

---

## Success Criteria

- ✅ Service runs in Kubernetes
- ✅ CI pipeline builds and deploys automatically
- ✅ Zero-downtime deployments
- ✅ Rollback works
- ✅ Metrics and logs available
- ✅ Alerts fire on errors
- ✅ All configuration in Git
- ✅ Secrets managed properly

---

**You've completed the DevOps curriculum! 🎉**

---

**Previous:** [15. Failure Modes](../15-failure-modes/)
