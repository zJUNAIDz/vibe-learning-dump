# Makefile vs npm Scripts vs Task Runners

> **Use the right tool for the job. npm scripts for Node-specific tasks. Make for project-wide orchestration. Don't overthink it.**

---

## 🟢 Side-by-Side Comparison

### npm scripts (package.json)

```json
{
  "scripts": {
    "dev": "tsx watch src/index.ts",
    "build": "tsc && esbuild src/index.ts --bundle --outdir=dist",
    "test": "vitest run",
    "test:watch": "vitest",
    "lint": "eslint src/ --ext .ts",
    "fmt": "prettier --write 'src/**/*.ts'",
    "db:migrate": "prisma migrate deploy",
    "db:seed": "tsx prisma/seed.ts",
    "docker:build": "docker build -t myapp .",
    "deploy": "kubectl apply -f k8s/"
  }
}
```

### Makefile

```makefile
.PHONY: dev build test lint fmt db-migrate db-seed docker-build deploy

dev:
	npx tsx watch src/index.ts

build:
	npx tsc && npx esbuild src/index.ts --bundle --outdir=dist

test:
	npx vitest run

lint:
	npx eslint src/ --ext .ts

fmt:
	npx prettier --write 'src/**/*.ts'

db-migrate:
	npx prisma migrate deploy

docker-build:
	docker build -t myapp:$(VERSION) .

deploy: docker-build
	kubectl apply -f k8s/
	kubectl rollout status deployment/myapp
```

---

## 🟢 When to Use Which

| Criteria | npm scripts | Makefile |
|----------|------------|---------|
| **Node.js tasks** | **Best** — native integration | Good — calls npm/npx |
| **Multi-language** | Poor — Node.js focused | **Best** — language agnostic |
| **Dependencies** | None — runs tasks independently | **Best** — `deploy: build test` |
| **Conditional logic** | Awkward — needs shell tricks | Good — ifeq, ifdef |
| **CI/CD** | Good | **Best** — standard on all runners |
| **Docker/K8s** | Verbose — everything in strings | **Best** — native shell |
| **Variables** | Limited — env vars only | **Best** — dynamic, overridable |
| **Team familiarity** | High — every JS dev knows it | Medium — some devs avoid Make |
| **Windows support** | **Good** — cross-platform Node | Poor — needs WSL or make port |
| **Discovery** | `npm run` (lists all) | `make help` (must implement) |

---

## 🟡 The Hybrid Pattern (Best of Both)

```json
// package.json — Node-specific tasks
{
  "scripts": {
    "dev": "tsx watch src/index.ts",
    "build": "tsc",
    "test": "vitest run",
    "test:watch": "vitest",
    "lint": "eslint src/ --ext .ts",
    "fmt": "prettier --write 'src/**/*.ts'"
  }
}
```

```makefile
# Makefile — orchestration, Docker, K8s, infra
.PHONY: build test lint deploy ci help

VERSION ?= $(shell git rev-parse --short HEAD)
IMAGE := registry.example.com/myapp:$(VERSION)

# Use npm for Node tasks
build:
	npm run build

test:
	npm test

lint:
	npm run lint

# Use Make for everything else
docker-build: build
	docker build -t $(IMAGE) .

docker-push: docker-build
	docker push $(IMAGE)

deploy: docker-push
	kubectl set image deployment/myapp myapp=$(IMAGE)
	kubectl rollout status deployment/myapp

ci: lint test docker-build
	@echo "CI passed"

# Terraform
tf-plan:
	cd terraform/ && terraform plan

tf-apply:
	cd terraform/ && terraform apply

help:
	@echo "Dev:     make build | test | lint"
	@echo "Docker:  make docker-build | docker-push"
	@echo "Deploy:  make deploy VERSION=abc123"
	@echo "Infra:   make tf-plan | tf-apply"
```

**Pattern:**
```
npm scripts → dev, build, test, lint, fmt (Node toolchain)
Makefile    → docker, deploy, infra, ci (system-level orchestration)

Developers run: npm run dev, npm test
CI runs:        make ci
Deploy runs:    make deploy VERSION=1.2.3
```

---

## 🟡 Other Task Runners

### Just (justfile)

```just
# justfile — modern alternative to Make

# Variables
version := `git rev-parse --short HEAD`
image := "registry.example.com/myapp:" + version

# Recipes (no tab requirement!)
build:
    npm run build

test:
    npm test

docker-build: build
    docker build -t {{image}} .

deploy env="staging": docker-build
    kubectl -n {{env}} set image deployment/myapp myapp={{image}}

# Conditional
ci: lint test docker-build
    echo "CI passed"
```

**Just vs Make:**
```
Just pros:
  ✅ No tab requirement (spaces work)
  ✅ Better error messages
  ✅ Built-in argument support: just deploy production
  ✅ Cross-platform
  
Just cons:
  ❌ Not pre-installed (must install)
  ❌ Less ubiquitous
  ❌ No file-based dependency tracking
```

### Task (Taskfile.yml)

```yaml
# Taskfile.yml
version: '3'

vars:
  VERSION:
    sh: git rev-parse --short HEAD
  IMAGE: "registry.example.com/myapp:{{.VERSION}}"

tasks:
  build:
    cmds:
      - npm run build

  test:
    cmds:
      - npm test

  docker-build:
    deps: [build]
    cmds:
      - docker build -t {{.IMAGE}} .

  deploy:
    deps: [docker-build]
    cmds:
      - kubectl set image deployment/myapp myapp={{.IMAGE}}
    requires:
      vars: [ENVIRONMENT]
```

### Comparison Table

| Feature | Make | npm scripts | Just | Task |
|---------|------|-------------|------|------|
| Pre-installed | Linux/Mac | With Node.js | No | No |
| Config format | Makefile | package.json | justfile | Taskfile.yml |
| Tab-sensitive | Yes | No | No | No |
| Dependencies | Yes | No | Yes | Yes |
| Variables | Yes | Limited | Yes | Yes |
| Parallel | Yes (-j) | Limited | No | Yes |
| File tracking | Yes | No | No | Yes |
| Cross-platform | Linux/Mac | Yes | Yes | Yes |

---

## 🟡 Decision Framework

```
Starting a new project?

  Is it Node.js only?
    YES → npm scripts for everything. Simple.
    
  Is it multi-language (Go + TS, or Python + TS)?
    YES → Makefile for orchestration + npm scripts for Node tasks.
    
  Does the team hate Make?
    YES → Use Just or Task. Same concept, friendlier syntax.
    
  Is it a DevOps/infra project?
    YES → Makefile. Every CI runner has it. Period.
    
  Complex CI/CD pipeline?
    YES → Makefile as the local entry point.
          CI calls `make ci` instead of duplicating commands.

Rule of thumb:
  - 1 language, simple project → npm/go scripts are fine
  - Multi-language, Docker, K8s, Terraform → Makefile
  - Team prefers YAML → Taskfile
  - Team wants modern Make → Just
```

---

## 🔴 The Makefile-CI Connection

The best pattern: CI configuration calls Make targets.

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  ci:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
      - run: make ci    # ← One command!
```

```makefile
# Makefile
ci: install lint test docker-build
	@echo "CI passed"
```

**Why this is powerful:**
```
1. Developer runs `make ci` locally — same checks as CI
2. CI config is simple — just `make ci`
3. Adding a step? Update Makefile, not CI config
4. Works with ANY CI system (GitHub Actions, Jenkins, GitLab CI)
5. No vendor lock-in to CI-specific features
```

---

**Previous:** [03. Advanced Features](./03-advanced-features.md)  
**Next:** [Module 12: Cloud Fundamentals](../12-cloud-fundamentals/README.md)
