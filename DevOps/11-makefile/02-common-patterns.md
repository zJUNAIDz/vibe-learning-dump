# Common Makefile Patterns

> **Every project needs these targets: build, test, clean, deploy, help. Here are production-ready patterns you can copy.**

---

## 🟢 The Standard Target Set

```makefile
.PHONY: all build test clean deploy lint fmt help

# Default target
all: lint test build

build:
	@echo "Building..."
	docker build -t $(IMAGE) .

test:
	@echo "Running tests..."
	npm test

lint:
	@echo "Linting..."
	npm run lint

fmt:
	@echo "Formatting..."
	npx prettier --write .

clean:
	@echo "Cleaning..."
	rm -rf dist/ node_modules/ coverage/

deploy: build
	@echo "Deploying..."
	kubectl apply -f k8s/

help:
	@echo "Available targets:"
	@echo "  build   - Build Docker image"
	@echo "  test    - Run tests"
	@echo "  lint    - Run linter"
	@echo "  fmt     - Format code"
	@echo "  clean   - Remove build artifacts"
	@echo "  deploy  - Deploy to Kubernetes"
```

---

## 🟢 TypeScript/Node.js Project

```makefile
.PHONY: all build test clean dev lint fmt install ci deploy help

APP_NAME := my-api
VERSION ?= $(shell git rev-parse --short HEAD)
REGISTRY := ghcr.io/myorg
IMAGE := $(REGISTRY)/$(APP_NAME):$(VERSION)

# ──────────────────────────────────────────────
# Development
# ──────────────────────────────────────────────

install:
	npm ci

dev:
	npm run dev

# ──────────────────────────────────────────────
# Quality
# ──────────────────────────────────────────────

lint:
	npx eslint src/ --ext .ts
	npx tsc --noEmit

fmt:
	npx prettier --write "src/**/*.ts"

fmt-check:
	npx prettier --check "src/**/*.ts"

test:
	npm test -- --coverage

test-watch:
	npm test -- --watch

# ──────────────────────────────────────────────
# Build
# ──────────────────────────────────────────────

build:
	npx tsc
	@echo "Build complete: dist/"

docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		-t $(IMAGE) \
		-t $(REGISTRY)/$(APP_NAME):latest .
	@echo "Built: $(IMAGE)"

docker-push: docker-build
	docker push $(IMAGE)
	docker push $(REGISTRY)/$(APP_NAME):latest

# ──────────────────────────────────────────────
# Deploy
# ──────────────────────────────────────────────

deploy:
	kubectl set image deployment/$(APP_NAME) \
		$(APP_NAME)=$(IMAGE) -n production
	kubectl rollout status deployment/$(APP_NAME) -n production

deploy-staging:
	kubectl set image deployment/$(APP_NAME) \
		$(APP_NAME)=$(IMAGE) -n staging
	kubectl rollout status deployment/$(APP_NAME) -n staging

rollback:
	kubectl rollout undo deployment/$(APP_NAME) -n production

# ──────────────────────────────────────────────
# CI
# ──────────────────────────────────────────────

ci: install lint fmt-check test docker-build
	@echo "CI passed ✅"

# ──────────────────────────────────────────────
# Cleanup
# ──────────────────────────────────────────────

clean:
	rm -rf dist/ coverage/ node_modules/.cache
	-docker rmi $(IMAGE) 2>/dev/null

clean-all: clean
	rm -rf node_modules/

# ──────────────────────────────────────────────
# Help
# ──────────────────────────────────────────────

help:
	@echo "$(APP_NAME) — Development Commands"
	@echo ""
	@echo "Development:"
	@echo "  install      Install dependencies"
	@echo "  dev          Start dev server"
	@echo ""
	@echo "Quality:"
	@echo "  lint         Run ESLint + TypeScript check"
	@echo "  fmt          Format code"
	@echo "  test         Run tests with coverage"
	@echo ""
	@echo "Build & Deploy:"
	@echo "  docker-build Build Docker image"
	@echo "  deploy       Deploy to production"
	@echo "  rollback     Rollback production"
	@echo ""
	@echo "  VERSION=$(VERSION)"
	@echo "  IMAGE=$(IMAGE)"
```

---

## 🟢 Go Project

```makefile
.PHONY: all build test clean lint run help

APP_NAME := myservice
VERSION ?= $(shell git describe --tags --always --dirty)
GIT_SHA := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(GIT_SHA) -X main.buildTime=$(BUILD_TIME)"

# Go settings
GOBIN := $(shell go env GOPATH)/bin
GOOS ?= linux
GOARCH ?= amd64

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/$(APP_NAME)
	@echo "Built: bin/$(APP_NAME) ($(VERSION))"

run:
	go run $(LDFLAGS) ./cmd/$(APP_NAME)

test:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

test-integration:
	go test -race -tags=integration -count=1 ./...

bench:
	go test -bench=. -benchmem ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	goimports -w .

vet:
	go vet ./...

# Generate mocks, protobuf, etc.
generate:
	go generate ./...

# Dependencies
tidy:
	go mod tidy
	go mod verify

# Cross-compile
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-linux-amd64 ./cmd/$(APP_NAME)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin-arm64 ./cmd/$(APP_NAME)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-windows-amd64.exe ./cmd/$(APP_NAME)

clean:
	rm -rf bin/ coverage.out

ci: lint vet test build
	@echo "CI passed"
```

---

## 🟡 Terraform + Ansible Project

```makefile
.PHONY: help plan apply destroy configure deploy

ENVIRONMENT ?= staging
TF_DIR := terraform/environments/$(ENVIRONMENT)
ANSIBLE_DIR := ansible

# ──────────────────────────────────────────────
# Infrastructure (Terraform)
# ──────────────────────────────────────────────

tf-init:
	cd $(TF_DIR) && terraform init

tf-plan: tf-init
	cd $(TF_DIR) && terraform plan -out=tfplan

tf-apply: tf-plan
	cd $(TF_DIR) && terraform apply tfplan
	$(MAKE) generate-inventory

tf-destroy:
	cd $(TF_DIR) && terraform destroy

tf-output:
	cd $(TF_DIR) && terraform output

# ──────────────────────────────────────────────
# Configuration (Ansible)
# ──────────────────────────────────────────────

generate-inventory:
	./scripts/generate-inventory.sh $(ENVIRONMENT)

configure:
	cd $(ANSIBLE_DIR) && \
		ansible-playbook -i inventory/$(ENVIRONMENT)/hosts.yml \
			playbooks/site.yml --diff

deploy:
	@test -n "$(VERSION)" || (echo "Usage: make deploy VERSION=1.2.3"; exit 1)
	cd $(ANSIBLE_DIR) && \
		ansible-playbook -i inventory/$(ENVIRONMENT)/hosts.yml \
			playbooks/deploy.yml \
			-e "app_version=$(VERSION)"

# ──────────────────────────────────────────────
# Full Lifecycle
# ──────────────────────────────────────────────

up: tf-apply configure
	@echo "Infrastructure up and configured for $(ENVIRONMENT)"

down: tf-destroy
	@echo "Infrastructure destroyed for $(ENVIRONMENT)"
```

---

## 🟡 Docker Compose Project

```makefile
.PHONY: up down logs restart build clean

COMPOSE := docker compose
COMPOSE_FILE := docker-compose.yml
COMPOSE_DEV := docker-compose.dev.yml

# Development
up:
	$(COMPOSE) -f $(COMPOSE_FILE) -f $(COMPOSE_DEV) up -d
	@echo "Services started. API: http://localhost:3000"

down:
	$(COMPOSE) down

down-clean:
	$(COMPOSE) down -v --remove-orphans

logs:
	$(COMPOSE) logs -f

logs-api:
	$(COMPOSE) logs -f api

restart:
	$(COMPOSE) restart $(SERVICE)

build:
	$(COMPOSE) build --no-cache

ps:
	$(COMPOSE) ps

# Database
db-shell:
	$(COMPOSE) exec db psql -U postgres -d myapp

db-migrate:
	$(COMPOSE) exec api npm run migrate

db-seed:
	$(COMPOSE) exec api npm run seed

# Testing
test:
	$(COMPOSE) exec api npm test

test-e2e:
	$(COMPOSE) -f docker-compose.test.yml up --abort-on-container-exit
```

---

## 🟡 Self-Documenting Help Target

```makefile
# The best help target — extracts comments from the Makefile

.DEFAULT_GOAL := help

## Build the Docker image
build:
	docker build -t $(IMAGE) .

## Run tests with coverage
test:
	npm test -- --coverage

## Deploy to production (requires VERSION)
deploy:
	kubectl set image deployment/myapp myapp=$(IMAGE)

## Show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## //'
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*' $(MAKEFILE_LIST) | \
		awk -F: '/^##/{desc=$$0; next} {if(desc){print "  " $$1 "\t" desc; desc=""}}' 
```

### Better version (extracts ## comments above targets):

```makefile
.DEFAULT_GOAL := help

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n\n"} \
		/^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' \
		$(MAKEFILE_LIST)

build: ## Build Docker image
	docker build -t $(IMAGE) .

test: ## Run tests
	npm test

deploy: ## Deploy to production
	kubectl apply -f k8s/

clean: ## Remove build artifacts
	rm -rf dist/
```

Output:
```
Usage: make <target>

  help             Show this help
  build            Build Docker image
  test             Run tests
  deploy           Deploy to production
  clean            Remove build artifacts
```

---

## 🔴 Anti-Patterns

### ❌ Spaces instead of tabs

```makefile
# This will FAIL — spaces are not tabs
build:
    docker build -t myapp .    # ← SPACES (invisible error!)
# Makefile:2: *** missing separator.  Stop.

# This works — must be a real TAB character
build:
	docker build -t myapp .    # ← TAB
```

### ❌ Not using .PHONY

```makefile
# If a file named "test" exists, this won't run:
test:
	npm test
# make: 'test' is up to date.

# Fix:
.PHONY: test
test:
	npm test
```

### ❌ Complex logic in Makefiles

```makefile
# BAD — Makefiles aren't shell scripts
deploy:
	if [ "$(ENVIRONMENT)" = "production" ]; then \
		if [ -z "$(VERSION)" ]; then \
			echo "VERSION required for production"; \
			exit 1; \
		fi; \
		if [ "$(BRANCH)" != "main" ]; then \
			echo "Must deploy from main"; \
			exit 1; \
		fi; \
	fi; \
	# 30 more lines of conditionals...

# GOOD — move complex logic to a script
deploy:
	./scripts/deploy.sh $(ENVIRONMENT) $(VERSION)
```

---

**Previous:** [01. Make Fundamentals](./01-make-fundamentals.md)  
**Next:** [03. Advanced Features](./03-advanced-features.md)
