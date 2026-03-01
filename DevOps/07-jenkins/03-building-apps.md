# Building TypeScript/Go Apps in Jenkins

> **Your CI pipeline should build exactly what production runs — same tools, same versions, same environment.**

---

## 🟢 Node.js / TypeScript Pipeline

### Complete TypeScript Pipeline

```groovy
pipeline {
    agent {
        docker {
            image 'node:20-alpine'
            args '-v /tmp/.npm:/root/.npm'   // Cache npm between builds
        }
    }
    
    environment {
        CI           = 'true'
        NODE_ENV     = 'test'
        APP_NAME     = 'my-ts-service'
        REGISTRY     = credentials('docker-registry-url')
        DOCKER_CREDS = credentials('docker-hub-creds')
    }
    
    options {
        timeout(time: 20, unit: 'MINUTES')
        disableConcurrentBuilds()
        buildDiscarder(logRotator(numToKeepStr: '10'))
    }
    
    stages {
        stage('Install') {
            steps {
                sh 'npm ci'
                sh 'echo "Installed $(npm ls --depth=0 2>/dev/null | wc -l) packages"'
            }
        }
        
        stage('Quality') {
            parallel {
                stage('Lint') {
                    steps {
                        sh 'npm run lint'
                    }
                }
                stage('Type Check') {
                    steps {
                        sh 'npx tsc --noEmit'
                    }
                }
                stage('Unit Tests') {
                    steps {
                        sh 'npm run test:unit -- --coverage'
                    }
                    post {
                        always {
                            junit 'junit-results/*.xml'
                            publishHTML(target: [
                                reportDir: 'coverage',
                                reportFiles: 'index.html',
                                reportName: 'Coverage Report'
                            ])
                        }
                    }
                }
            }
        }
        
        stage('Build') {
            steps {
                sh 'npm run build'
                sh 'ls -la dist/'
            }
        }
        
        stage('Integration Tests') {
            steps {
                sh 'npm run test:integration'
            }
        }
        
        stage('Docker Build') {
            agent { label 'docker' }
            steps {
                sh """
                    docker build \
                        --build-arg NODE_ENV=production \
                        -t ${APP_NAME}:${BUILD_NUMBER} \
                        -t ${APP_NAME}:latest \
                        .
                """
            }
        }
        
        stage('Docker Push') {
            when { branch 'main' }
            steps {
                sh '''
                    echo "$DOCKER_CREDS_PSW" | docker login -u "$DOCKER_CREDS_USR" --password-stdin
                    docker push ${REGISTRY}/${APP_NAME}:${BUILD_NUMBER}
                    docker push ${REGISTRY}/${APP_NAME}:latest
                '''
            }
        }
        
        stage('Deploy') {
            when { branch 'main' }
            steps {
                sh """
                    kubectl set image deployment/${APP_NAME} \
                        ${APP_NAME}=${REGISTRY}/${APP_NAME}:${BUILD_NUMBER} \
                        --namespace=production
                    kubectl rollout status deployment/${APP_NAME} \
                        --namespace=production --timeout=300s
                """
            }
        }
    }
    
    post {
        always {
            cleanWs()
        }
        failure {
            slackSend(
                channel: '#ci-alerts',
                color: 'danger',
                message: "❌ ${APP_NAME} build #${BUILD_NUMBER} FAILED\n${BUILD_URL}"
            )
        }
    }
}
```

### Key TypeScript Build Details

**Why `npm ci` not `npm install`:**

```
npm install:
  - Reads package.json
  - May update package-lock.json
  - Different runs → potentially different versions
  - NONDETERMINISTIC ❌

npm ci:
  - Reads package-lock.json EXACTLY
  - Deletes node_modules first
  - Fails if lock file doesn't match package.json
  - Same input → same output ALWAYS
  - DETERMINISTIC ✅
```

**Multi-stage Dockerfile for TypeScript:**

```dockerfile
# Stage 1: Build
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY tsconfig.json ./
COPY src/ ./src/
RUN npm run build
RUN npm prune --production    # Remove dev dependencies

# Stage 2: Production
FROM node:20-alpine
WORKDIR /app

# Security: non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./

USER appuser
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=5s CMD wget -qO- http://localhost:3000/health || exit 1
CMD ["node", "dist/index.js"]
```

---

## 🟢 Go Pipeline

### Complete Go Pipeline

```groovy
pipeline {
    agent {
        docker {
            image 'golang:1.22-alpine'
            args '-v /tmp/go-cache:/root/.cache/go-build'
        }
    }
    
    environment {
        GO111MODULE = 'on'
        CGO_ENABLED = '0'
        GOOS        = 'linux'
        GOARCH      = 'amd64'
        APP_NAME    = 'my-go-service'
        REGISTRY    = credentials('docker-registry-url')
    }
    
    options {
        timeout(time: 15, unit: 'MINUTES')
        buildDiscarder(logRotator(numToKeepStr: '10'))
    }
    
    stages {
        stage('Dependencies') {
            steps {
                sh 'go mod download'
                sh 'go mod verify'
            }
        }
        
        stage('Quality') {
            parallel {
                stage('Vet') {
                    steps {
                        sh 'go vet ./...'
                    }
                }
                stage('Lint') {
                    steps {
                        sh '''
                            # Install golangci-lint
                            wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
                            golangci-lint run ./...
                        '''
                    }
                }
                stage('Unit Tests') {
                    steps {
                        sh 'go test -v -race -coverprofile=coverage.out ./...'
                        sh 'go tool cover -func=coverage.out'
                    }
                    post {
                        always {
                            archiveArtifacts artifacts: 'coverage.out', fingerprint: true
                        }
                    }
                }
            }
        }
        
        stage('Build') {
            steps {
                sh '''
                    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
                    COMMIT=$(git rev-parse --short HEAD)
                    BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
                    
                    go build \
                        -ldflags="-s -w \
                            -X main.version=${VERSION} \
                            -X main.commit=${COMMIT} \
                            -X main.buildTime=${BUILD_TIME}" \
                        -o bin/${APP_NAME} \
                        ./cmd/server
                    
                    ls -la bin/
                    file bin/${APP_NAME}
                '''
            }
        }
        
        stage('Integration Tests') {
            steps {
                sh 'go test -v -tags=integration ./...'
            }
        }
        
        stage('Docker Build') {
            agent { label 'docker' }
            steps {
                sh """
                    docker build \
                        -t ${REGISTRY}/${APP_NAME}:${BUILD_NUMBER} \
                        -t ${REGISTRY}/${APP_NAME}:latest \
                        .
                """
                sh "docker images | grep ${APP_NAME}"
            }
        }
        
        stage('Security Scan') {
            steps {
                sh """
                    # Scan Docker image for vulnerabilities
                    trivy image --severity HIGH,CRITICAL \
                        --exit-code 1 \
                        ${REGISTRY}/${APP_NAME}:${BUILD_NUMBER}
                """
            }
        }
        
        stage('Deploy') {
            when { branch 'main' }
            steps {
                input message: 'Deploy to production?', ok: 'Deploy'
                sh """
                    kubectl set image deployment/${APP_NAME} \
                        ${APP_NAME}=${REGISTRY}/${APP_NAME}:${BUILD_NUMBER}
                    kubectl rollout status deployment/${APP_NAME} --timeout=300s
                """
            }
        }
    }
    
    post {
        always { cleanWs() }
    }
}
```

### Go Binary Optimization

```bash
# Development build (with debug info)
go build -o myapp ./cmd/server
# Size: ~15MB

# Production build (stripped)
CGO_ENABLED=0 go build -ldflags="-s -w" -o myapp ./cmd/server
# Size: ~8MB  (46% smaller)

# What -ldflags flags do:
# -s  → Strip symbol table (no debugger support)
# -w  → Strip DWARF debug info
# Both are fine for production — you don't debug production binaries
```

### Multi-stage Dockerfile for Go

```dockerfile
# Stage 1: Build
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/bin/server ./cmd/server

# Stage 2: Minimal production image
FROM scratch                          # Empty image — nothing but our binary
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bin/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
```

```
Image sizes:
  golang:1.22         → ~800MB (full Go toolchain)
  golang:1.22-alpine  → ~250MB (Alpine-based Go)
  alpine:3.19         → ~7MB   (minimal OS)
  scratch             → ~0MB   (literally nothing)
  
  Our final image: scratch + binary ≈ 8MB total
```

---

## 🟡 Parallel Build Stages

### Running Independent Tasks in Parallel

```groovy
stage('Quality Gates') {
    parallel {
        stage('Unit Tests') {
            agent { label 'linux' }
            steps {
                sh 'go test -v ./...'
            }
        }
        stage('Lint') {
            agent { label 'linux' }
            steps {
                sh 'golangci-lint run'
            }
        }
        stage('Security Scan') {
            agent { label 'linux' }
            steps {
                sh 'gosec ./...'
            }
        }
        stage('Benchmark') {
            agent { label 'linux' }
            steps {
                sh 'go test -bench=. -benchmem ./...'
            }
        }
    }
}
```

```
Without parallel:
  Unit Tests (3m) → Lint (1m) → Security (2m) → Benchmark (2m) = 8 minutes

With parallel:
  Unit Tests (3m) ─┐
  Lint (1m)       ─┤
  Security (2m)   ─┤ = 3 minutes (time of slowest stage)
  Benchmark (2m)  ─┘
```

### Matrix Builds (Multiple Environments)

```groovy
stage('Test Matrix') {
    matrix {
        axes {
            axis {
                name 'GO_VERSION'
                values '1.21', '1.22'
            }
            axis {
                name 'OS'
                values 'linux', 'darwin'
            }
        }
        stages {
            stage('Test') {
                agent {
                    docker {
                        image "golang:${GO_VERSION}-alpine"
                    }
                }
                steps {
                    sh "echo Testing Go ${GO_VERSION} on ${OS}"
                    sh "GOOS=${OS} go test ./..."
                }
            }
        }
    }
}
```

---

## 🟡 Docker-in-Docker vs Docker Socket

### Two Ways to Build Docker Images in Jenkins

**Option 1: Docker Socket Mount (Simpler, Less Isolated)**

```groovy
pipeline {
    agent {
        docker {
            image 'docker:24-cli'
            args '-v /var/run/docker.sock:/var/run/docker.sock'
            // Uses the HOST's Docker daemon
        }
    }
    stages {
        stage('Build') {
            steps {
                sh 'docker build -t myapp .'
            }
        }
    }
}
```

```
⚠️ Security risk:
  - Agent container can access ALL containers on the host
  - Can mount host filesystem
  - Essentially root access to the host
  
Use when: Trusted builds, internal CI
```

**Option 2: Docker-in-Docker (More Isolated)**

```groovy
pipeline {
    agent {
        docker {
            image 'docker:24-dind'
            args '--privileged'     // Required for DinD
        }
    }
    stages {
        stage('Build') {
            steps {
                sh 'docker build -t myapp .'
            }
        }
    }
}
```

```
⚠️ Also has issues:
  - --privileged is a security risk
  - Performance overhead (nested filesystem layers)
  - Caching is harder
  
Use when: Strict isolation requirements
```

**Option 3: Kaniko (No Docker Daemon)**

```groovy
stage('Build with Kaniko') {
    agent {
        kubernetes {
            yaml '''
                apiVersion: v1
                kind: Pod
                spec:
                  containers:
                  - name: kaniko
                    image: gcr.io/kaniko-project/executor:debug
                    command: ["tail", "-f", "/dev/null"]
                    volumeMounts:
                    - name: docker-config
                      mountPath: /kaniko/.docker
                  volumes:
                  - name: docker-config
                    secret:
                      secretName: docker-credentials
            '''
        }
    }
    steps {
        container('kaniko') {
            sh """
                /kaniko/executor \
                    --context=dir://. \
                    --destination=registry.example.com/myapp:${BUILD_NUMBER}
            """
        }
    }
}
```

```
✅ No Docker daemon needed
✅ No privileged mode
✅ Works in Kubernetes pods
✅ Most secure option for container builds in CI
```

---

## 🔴 Common Build Pipeline Problems

### Problem 1: Flaky Tests

```groovy
// ❌ Pipeline fails randomly because of flaky test
stage('Test') {
    steps {
        sh 'npm test'     // Fails 1 in 10 runs
    }
}

// ✅ Retry (temporary fix while you fix the test)
stage('Test') {
    steps {
        retry(3) {
            sh 'npm test'
        }
    }
}

// ✅✅ Better: quarantine flaky tests and fix them
stage('Test') {
    steps {
        sh 'npm test -- --exclude=flaky'
    }
}
```

### Problem 2: Slow Builds

```
Common culprits:
1. npm install vs npm ci (install resolves every time)
2. No dependency caching
3. Sequential stages that could be parallel
4. Building from scratch instead of using Docker layer cache
5. Running ALL tests when only one file changed
```

```groovy
// Cache Go modules between builds
pipeline {
    agent {
        docker {
            image 'golang:1.22-alpine'
            args '-v go-mod-cache:/go/pkg/mod -v go-build-cache:/root/.cache/go-build'
        }
    }
}
```

### Problem 3: "Works in Jenkins, Fails Locally" (or Vice Versa)

```
Why this happens:
1. Different tool versions (Node 18 vs Node 20)
2. Different OS (alpine vs ubuntu)
3. Missing environment variables
4. Different file permissions
5. Network access differences

Solution: Use Docker for local builds too!
```

```bash
# Developer runs the same Docker image locally
docker run --rm -v $(pwd):/app -w /app node:20-alpine sh -c "npm ci && npm test"

# Same image, same tools, same result — locally and in Jenkins
```

---

**Previous:** [02. Jenkinsfile (Pipeline as Code)](./02-jenkinsfile-pipeline-as-code.md)  
**Next:** [04. Integration with Kubernetes](./04-jenkins-kubernetes-integration.md)
