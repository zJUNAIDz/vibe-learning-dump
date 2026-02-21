# Module 11: Makefile for Developers

> **Make is 50 years old and still relevant — here's why**

---

## Why Makefile?

**Problem:** Developers forget commands.

```bash
# How do I build this project again?
docker build -t myapp:latest --build-arg NODE_VERSION=18 .

# How do I run tests?
npm test -- --coverage --verbose

# How do I deploy?
kubectl apply -f k8s/ && kubectl rollout status deployment/myapp
```

**Solution:** Makefile

```bash
make build
make test
make deploy
```

---

## Topics Covered

### 📝 01. Make Fundamentals
- Targets
- Dependencies
- Phony targets
- Variables

### 📝 02. Common Patterns
- build, test, deploy targets
- Clean targets
- Help target

### 📝 03. Advanced Features
- Pattern rules
- Automatic variables
- Conditional execution

### 📝 04. Makefile vs npm scripts vs task runners
- When to use which
- Pros and cons

---

**Previous:** [10. Ansible](../10-ansible/)  
**Next:** [12. Cloud Fundamentals](../12-cloud-fundamentals/)
