# Advanced Make Features

> **Pattern rules, automatic variables, and conditional execution make Makefiles powerful. But use them sparingly — readability matters more than cleverness.**

---

## 🟡 Pattern Rules

Pattern rules use `%` as a wildcard to create generic rules.

```makefile
# Build any Go binary from its cmd/ directory
bin/%: cmd/%/main.go
	go build -o $@ ./cmd/$*

# Usage:
# make bin/api       → go build -o bin/api ./cmd/api
# make bin/worker    → go build -o bin/worker ./cmd/worker
# make bin/cli       → go build -o bin/cli ./cmd/cli

# Compile any .ts to .js
dist/%.js: src/%.ts
	npx tsc $< --outDir dist/

# Build any Docker image from its Dockerfile
docker-%:
	docker build -t $(REGISTRY)/$*:$(VERSION) -f docker/$*/Dockerfile .
# make docker-api    → builds from docker/api/Dockerfile
# make docker-worker → builds from docker/worker/Dockerfile
```

---

## 🟡 Automatic Variables

```makefile
# $@ — Target name
# $< — First prerequisite
# $^ — All prerequisites
# $* — The "stem" matched by %
# $? — Prerequisites newer than target
# $(@D) — Directory of target
# $(@F) — Filename of target

# Examples
bin/myapp: src/main.go src/handler.go src/db.go
	go build -o $@ $^
	# $@ = bin/myapp
	# $< = src/main.go
	# $^ = src/main.go src/handler.go src/db.go

dist/%.min.css: src/%.css
	csso $< -o $@
	# For input: make dist/app.min.css
	# $@ = dist/app.min.css
	# $< = src/app.css
	# $* = app
```

---

## 🟡 Conditional Execution

### ifeq / ifneq

```makefile
# Check environment
ifeq ($(ENVIRONMENT),production)
  REPLICAS := 3
  LOG_LEVEL := warn
else
  REPLICAS := 1
  LOG_LEVEL := debug
endif

deploy:
	kubectl scale deployment/myapp --replicas=$(REPLICAS)

# Check if variable is set
ifdef VERSION
  TAG := $(VERSION)
else
  TAG := latest
endif

# Check OS
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
  OS := linux
endif
ifeq ($(UNAME_S),Darwin)
  OS := darwin
endif
```

### Conditional commands

```makefile
# Check if a command exists
lint:
	@command -v golangci-lint >/dev/null 2>&1 || \
		(echo "Installing golangci-lint..." && \
		 go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

# Guard against missing variables  
deploy:
	@test -n "$(VERSION)" || (echo "VERSION is required. Usage: make deploy VERSION=1.2.3" && exit 1)
	@test -n "$(ENVIRONMENT)" || (echo "ENVIRONMENT is required" && exit 1)
	kubectl set image deployment/myapp myapp=$(IMAGE)
```

---

## 🟡 Functions

```makefile
# $(wildcard pattern) — expand file globs
SRC_FILES := $(wildcard src/*.ts)
TEST_FILES := $(wildcard src/**/*_test.go)

# $(patsubst pattern,replacement,text) — pattern substitution
JS_FILES := $(patsubst src/%.ts,dist/%.js,$(SRC_FILES))
# src/main.ts src/utils.ts → dist/main.js dist/utils.js

# $(filter pattern,text) — keep matching
GO_FILES := $(filter %.go,$(ALL_FILES))

# $(filter-out pattern,text) — remove matching
NON_TEST := $(filter-out %_test.go,$(GO_FILES))

# $(foreach var,list,text) — loop
SERVICES := api worker cron
IMAGES := $(foreach svc,$(SERVICES),$(REGISTRY)/$(svc):$(VERSION))
# registry.example.com/api:v1 registry.example.com/worker:v1 ...

# $(shell command) — run shell command
GIT_SHA := $(shell git rev-parse --short HEAD)
TIMESTAMP := $(shell date -u +%Y%m%dT%H%M%SZ)
FILE_COUNT := $(shell find src -name '*.ts' | wc -l)

# $(word n,text) — nth word
FIRST_SERVICE := $(word 1,$(SERVICES))  # api

# $(sort list) — sort and remove duplicates
UNIQUE := $(sort $(LIST))
```

---

## 🟡 Parallel Execution

```makefile
# Run targets in parallel with -j flag
# make -j4 build-api build-worker build-cron

# Mark targets for parallel execution
build-all: build-api build-worker build-cron

build-api:
	docker build -t $(REGISTRY)/api:$(VERSION) -f docker/api/Dockerfile .

build-worker:
	docker build -t $(REGISTRY)/worker:$(VERSION) -f docker/worker/Dockerfile .

build-cron:
	docker build -t $(REGISTRY)/cron:$(VERSION) -f docker/cron/Dockerfile .

# Usage:
# make -j3 build-all    ← Builds all 3 images in parallel
# make build-all        ← Builds sequentially (default)
```

### Order-Only Prerequisites

```makefile
# | means "ensure this exists but don't rebuild when it changes"
bin/myapp: src/main.go | bin/
	go build -o $@ .

bin/:
	mkdir -p bin/

# bin/ is created if missing, but changing it doesn't trigger rebuild
```

---

## 🟡 Recursive Make (Multi-Directory Projects)

```makefile
# Root Makefile
SUBDIRS := services/api services/worker services/cron

.PHONY: all clean test $(SUBDIRS)

all: $(SUBDIRS)

$(SUBDIRS):
	$(MAKE) -C $@ build

clean:
	$(foreach dir,$(SUBDIRS),$(MAKE) -C $(dir) clean;)

test:
	$(foreach dir,$(SUBDIRS),$(MAKE) -C $(dir) test;)
```

```makefile
# services/api/Makefile
.PHONY: build test clean

build:
	go build -o bin/api ./cmd/api

test:
	go test ./...

clean:
	rm -rf bin/
```

---

## 🟡 .env File Loading

```makefile
# Load .env file if it exists
-include .env
export

# Or explicitly
ifneq (,$(wildcard .env))
  include .env
  export
endif

# Now all variables from .env are available
build:
	docker build \
		--build-arg DATABASE_URL=$(DATABASE_URL) \
		-t $(IMAGE) .
```

---

## 🔴 Advanced: Target-Specific Variables

```makefile
# Variables that only apply to specific targets

deploy-staging: ENVIRONMENT := staging
deploy-staging: REPLICAS := 1
deploy-staging: deploy

deploy-production: ENVIRONMENT := production
deploy-production: REPLICAS := 3
deploy-production: deploy

deploy:
	@echo "Deploying to $(ENVIRONMENT) with $(REPLICAS) replicas"
	kubectl -n $(ENVIRONMENT) set image deployment/myapp myapp=$(IMAGE)
	kubectl -n $(ENVIRONMENT) scale deployment/myapp --replicas=$(REPLICAS)
```

---

## 🔴 Advanced: Grouped Targets (GNU Make 4.3+)

```makefile
# Multiple outputs from one command
dist/main.js dist/main.css &: src/main.ts src/main.css
	npx webpack --config webpack.config.js

# The & means "this ONE command produces ALL listed targets"
# Without &, Make runs the command once per target
```

---

## 🔴 Performance: Only Rebuild What Changed

```makefile
# Timestamp-based rebuilds (Make's original superpower)

GO_SOURCES := $(shell find . -name '*.go' -not -path './vendor/*')
BINARY := bin/myapp

# Only rebuild if Go source files changed
$(BINARY): $(GO_SOURCES)
	go build -o $@ ./cmd/myapp

# Only rebuild Docker image if binary or Dockerfile changed  
.docker-build: $(BINARY) Dockerfile
	docker build -t $(IMAGE) .
	touch .docker-build

# .docker-build is a marker file
# If it's newer than binary + Dockerfile, skip docker build
```

---

**Previous:** [02. Common Patterns](./02-common-patterns.md)  
**Next:** [04. Makefile vs npm Scripts](./04-makefile-vs-npm-scripts.md)
