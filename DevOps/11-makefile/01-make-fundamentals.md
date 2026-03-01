# Make Fundamentals

> **Make is a 50-year-old build tool that's still everywhere. It's on every Linux box, every CI runner, every dev machine. One `make build` replaces a 200-character Docker command nobody remembers.**

---

## 🟢 What Is Make?

```
Make reads a Makefile and executes commands to build targets.

Traditional C usage:
  Makefile says: "To build program, compile main.c and utils.c"
  make → compiles only what changed → fast

Modern DevOps usage:
  Makefile says: "Here are all the commands you'll ever need"
  make build → docker build ...
  make test  → npm test ...
  make deploy → kubectl apply ...
  
  It's a project-level command runner.
```

---

## 🟢 Basic Syntax

```makefile
# Makefile

# target: dependencies
#     command (MUST be a tab, not spaces!)

build:
	docker build -t myapp:latest .

test:
	npm test

clean:
	rm -rf dist/ node_modules/
```

### The Rules

```
1. Indentation MUST be tabs (not spaces)
   - This is the #1 Makefile error
   - If you get "*** missing separator", you used spaces

2. Each line runs in a NEW shell
   - cd dir && command    ✅ (same shell)
   - cd dir               ❌ (cd is lost)
     command                   (runs in original dir)

3. First target is the default
   - Running `make` without arguments runs the first target
```

---

## 🟢 Targets and Dependencies

```makefile
# Target with no dependencies — always runs
build:
	docker build -t myapp:latest .

# Target with dependencies — runs dependencies first
deploy: build test
	kubectl apply -f k8s/

# Chain: make deploy → runs build, then test, then deploy

test:
	npm test

# File target — only runs if file doesn't exist or source is newer
dist/bundle.js: src/index.ts
	npx tsc src/index.ts --outDir dist/
# make dist/bundle.js
#   → If dist/bundle.js doesn't exist: compiles
#   → If src/index.ts is newer: recompiles
#   → If dist/bundle.js is up to date: "nothing to do"
```

---

## 🟢 .PHONY Targets

By default, Make checks if a file named after the target exists:

```makefile
# Problem: if a file called "test" exists, `make test` says "up to date"
test:
	npm test

# Solution: declare it as phony (not a real file)
.PHONY: test build clean deploy

test:
	npm test

build:
	docker build -t myapp:latest .
```

**Rule: Always make targets `.PHONY` unless they create actual files.**

```makefile
# Clean way to declare phony targets
.PHONY: all build test clean deploy lint fmt help

all: build test
```

---

## 🟢 Variables

```makefile
# Define variables
APP_NAME := myapp
VERSION := 1.2.3
REGISTRY := registry.example.com
IMAGE := $(REGISTRY)/$(APP_NAME):$(VERSION)
GO_FLAGS := -ldflags "-X main.version=$(VERSION)"

# Use variables
build:
	docker build -t $(IMAGE) .

push:
	docker push $(IMAGE)

# Override from command line
# make build VERSION=2.0.0
```

### Variable Types

```makefile
# := (simple assignment) — evaluated once when defined
NOW := $(shell date +%Y%m%d)

# = (recursive assignment) — evaluated each time used
NOW = $(shell date +%Y%m%d)  # Re-evaluated on each use

# ?= (conditional assignment) — only set if not already defined
VERSION ?= latest
# Can override: make build VERSION=1.2.3
# Or: VERSION=1.2.3 make build

# += (append)
CFLAGS := -Wall
CFLAGS += -Werror   # Now: -Wall -Werror
```

### Environment Variables

```makefile
# Environment variables are available
build:
	echo "User: $(USER)"
	echo "Home: $(HOME)"

# Override env vars
export APP_ENV := production

deploy:
	echo "Deploying to $(APP_ENV)"
```

### Shell Commands in Variables

```makefile
GIT_SHA := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
TIMESTAMP := $(shell date -u +%Y%m%dT%H%M%S)

build:
	docker build \
		--build-arg GIT_SHA=$(GIT_SHA) \
		--build-arg TIMESTAMP=$(TIMESTAMP) \
		-t $(APP_NAME):$(GIT_SHA) .
```

---

## 🟢 Multi-Line Commands

```makefile
# Each line is a separate shell! Variables don't persist across lines.

# ❌ WRONG — cd is lost on next line
deploy:
	cd k8s/
	kubectl apply -f .      # Runs in ORIGINAL directory

# ✅ RIGHT — chain with &&
deploy:
	cd k8s/ && kubectl apply -f .

# ✅ RIGHT — use backslash for continuation (same shell)
deploy:
	cd k8s/ && \
		kubectl apply -f deployment.yml && \
		kubectl apply -f service.yml && \
		kubectl rollout status deployment/myapp

# ✅ RIGHT — use .ONESHELL (GNU Make 3.82+)
.ONESHELL:
deploy:
	cd k8s/
	kubectl apply -f deployment.yml
	kubectl apply -f service.yml
```

---

## 🟡 Error Handling

```makefile
# By default, Make stops on first error

# Ignore errors with - prefix
clean:
	-rm -rf dist/           # Don't fail if dist/ doesn't exist
	-docker rmi myapp:latest

# Continue on error for specific commands
test:
	npm test || true         # Always "succeed" (dangerous!)
	
# Better: capture exit code
test:
	npm test; \
	EXIT_CODE=$$?; \
	echo "Tests exited with: $$EXIT_CODE"; \
	exit $$EXIT_CODE
```

### Silent Commands

```makefile
# @ prefix suppresses command echo
version:
	@echo $(VERSION)

# Without @: Make prints the command, then the output
# version:
#     echo 1.2.3    ← printed by Make
#     1.2.3         ← printed by echo

# With @: Only output shown
# version:
#     1.2.3         ← printed by echo only
```

---

## 🟡 Special Variables

```makefile
# $@ — the target name
# $< — first dependency
# $^ — all dependencies
# $? — dependencies newer than target

# Example
dist/app.js: src/main.ts src/utils.ts
	npx tsc $^ --outDir $(dir $@)
	# $@ = dist/app.js
	# $< = src/main.ts
	# $^ = src/main.ts src/utils.ts

# Escaping $ in shell commands
# Make uses $ for variables, shell also uses $
# Use $$ for literal $ in shell commands

list-pods:
	kubectl get pods | awk '{print $$1}'
	# $$ becomes $ for the shell
	# Without $$, Make tries to expand $1 as a variable
```

---

## 🟡 Including Other Makefiles

```makefile
# Include shared Makefile
include common.mk

# Include if it exists (don't error if missing)
-include .env.mk

# Use for environment-specific config
-include config/$(ENVIRONMENT).mk
```

---

**Previous:** [README](./README.md)  
**Next:** [02. Common Patterns](./02-common-patterns.md)
