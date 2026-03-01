# Jenkinsfile — Pipeline as Code

> **If your pipeline isn't in a Jenkinsfile, it's not reproducible. If it's not reproducible, it's not CI/CD.**

---

## 🟢 What Is a Jenkinsfile?

A **Jenkinsfile** is a text file (checked into your repo) that defines your CI/CD pipeline.

```
Without Jenkinsfile:
1. Log into Jenkins UI
2. Click "New Item" → "Pipeline"
3. Fill in 47 form fields
4. Click "Build Now"
5. "Who set this up? Nobody knows."
6. "It's broken. Nobody knows why."

With Jenkinsfile:
1. Put a Jenkinsfile in your repo root
2. Jenkins reads it automatically
3. Pipeline changes are code-reviewed
4. Pipeline history is in Git
5. Everyone can see and modify the pipeline
```

---

## 🟢 Declarative vs Scripted Pipelines

Jenkins has **two** pipeline syntaxes:

### Declarative Pipeline (Use This)

```groovy
pipeline {
    agent any
    
    environment {
        APP_NAME = 'my-service'
        REGISTRY = 'registry.example.com'
    }
    
    stages {
        stage('Build') {
            steps {
                sh 'npm ci'
                sh 'npm run build'
            }
        }
        
        stage('Test') {
            steps {
                sh 'npm test'
            }
        }
        
        stage('Deploy') {
            steps {
                sh 'kubectl apply -f k8s/'
            }
        }
    }
    
    post {
        success {
            echo 'Pipeline succeeded!'
        }
        failure {
            echo 'Pipeline failed!'
        }
    }
}
```

### Scripted Pipeline (Legacy/Advanced)

```groovy
node {
    stage('Build') {
        checkout scm
        sh 'npm ci'
        sh 'npm run build'
    }
    
    stage('Test') {
        try {
            sh 'npm test'
        } catch (err) {
            currentBuild.result = 'UNSTABLE'
        }
    }
    
    stage('Deploy') {
        if (env.BRANCH_NAME == 'main') {
            sh 'kubectl apply -f k8s/'
        }
    }
}
```

### When to Use Which?

| Feature | Declarative | Scripted |
|---------|-------------|----------|
| Syntax | Structured, opinionated | Free-form Groovy |
| Learning curve | Easy | Hard |
| Error handling | Built-in `post` block | Manual try/catch |
| Validation | Linted before running | Fails at runtime |
| Flexibility | Limited | Unlimited |
| **Recommendation** | **Use this 95% of the time** | Only when declarative can't do it |

---

## 🟢 Pipeline Anatomy

```groovy
pipeline {                              // 1. PIPELINE BLOCK (required)
    agent { label 'linux' }             // 2. WHERE to run

    options {                           // 3. PIPELINE OPTIONS
        timeout(time: 30, unit: 'MINUTES')
        disableConcurrentBuilds()
        buildDiscarder(logRotator(numToKeepStr: '10'))
    }

    environment {                       // 4. ENVIRONMENT VARIABLES
        APP_NAME = 'my-app'
        VERSION  = sh(script: 'git describe --tags', returnStdout: true).trim()
    }

    parameters {                        // 5. BUILD PARAMETERS
        string(name: 'DEPLOY_ENV', defaultValue: 'staging', description: 'Target environment')
        booleanParam(name: 'RUN_INTEGRATION', defaultValue: false, description: 'Run integration tests?')
    }

    stages {                            // 6. STAGES (the meat)
        stage('Build') {
            steps {
                sh 'make build'
            }
        }
        stage('Test') {
            steps {
                sh 'make test'
            }
        }
    }

    post {                              // 7. POST-BUILD ACTIONS
        always   { cleanWs() }
        success  { slackSend message: "Build passed" }
        failure  { slackSend message: "Build FAILED", color: 'danger' }
    }
}
```

---

## 🟢 Stages and Steps

### Sequential Stages

```groovy
stages {
    stage('Checkout') {
        steps {
            checkout scm
        }
    }
    stage('Build') {
        steps {
            sh 'npm ci'
            sh 'npm run build'
        }
    }
    stage('Unit Test') {
        steps {
            sh 'npm run test:unit'
        }
    }
    stage('Integration Test') {
        steps {
            sh 'npm run test:integration'
        }
    }
    stage('Deploy') {
        steps {
            sh 'make deploy'
        }
    }
}
```

### Parallel Stages

```groovy
stage('Test') {
    parallel {
        stage('Unit Tests') {
            agent { label 'linux' }
            steps {
                sh 'npm run test:unit'
            }
        }
        stage('Integration Tests') {
            agent { label 'linux' }
            steps {
                sh 'npm run test:integration'
            }
        }
        stage('Lint') {
            agent { label 'linux' }
            steps {
                sh 'npm run lint'
            }
        }
    }
}
```

**Why parallel?**
```
Sequential: Build (2m) → Unit (3m) → Integration (5m) → Lint (1m) = 11 minutes
Parallel:   Build (2m) → [Unit (3m) | Integration (5m) | Lint (1m)] = 7 minutes
                                                           ↑ Total = 2 + 5 = 7m
```

### Conditional Stages

```groovy
stage('Deploy to Production') {
    when {
        branch 'main'               // Only on main branch
        not { changeRequest() }     // Not on pull requests
    }
    steps {
        sh 'make deploy-prod'
    }
}

stage('Deploy to Staging') {
    when {
        anyOf {
            branch 'develop'
            branch 'staging'
        }
    }
    steps {
        sh 'make deploy-staging'
    }
}

stage('Run Expensive Tests') {
    when {
        expression { params.RUN_INTEGRATION == true }
    }
    steps {
        sh 'make test-integration'
    }
}
```

---

## 🟢 Environment Variables

### Built-in Variables

```groovy
pipeline {
    agent any
    stages {
        stage('Info') {
            steps {
                echo "Build Number: ${env.BUILD_NUMBER}"
                echo "Job Name:     ${env.JOB_NAME}"
                echo "Branch:       ${env.BRANCH_NAME}"
                echo "Workspace:    ${env.WORKSPACE}"
                echo "Jenkins URL:  ${env.JENKINS_URL}"
                echo "Build URL:    ${env.BUILD_URL}"
                echo "Node Name:    ${env.NODE_NAME}"
            }
        }
    }
}
```

### Custom Variables

```groovy
environment {
    // Static values
    APP_NAME = 'my-service'
    REGION   = 'us-east-1'
    
    // From shell commands
    GIT_HASH  = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
    VERSION   = sh(script: 'cat VERSION', returnStdout: true).trim()
    
    // From credentials (see next section)
    DOCKER_CREDS = credentials('docker-hub-creds')
}
```

### Stage-level Variables

```groovy
stages {
    stage('Build') {
        environment {
            NODE_ENV = 'production'     // Only available in this stage
        }
        steps {
            sh 'npm run build'          // NODE_ENV=production
        }
    }
    stage('Test') {
        environment {
            NODE_ENV = 'test'           // Different value for this stage
        }
        steps {
            sh 'npm test'               // NODE_ENV=test
        }
    }
}
```

---

## 🟢 Credentials Management

### Never Do This

```groovy
// ❌ NEVER NEVER NEVER
pipeline {
    environment {
        DB_PASSWORD = 'supersecret123'     // In plain text, in Git, visible to everyone
        AWS_SECRET  = 'AKIAIOSFODNN7EXAMPLE' // Leaked to logs, build history
    }
}
```

### Use Jenkins Credentials

**Step 1: Store credentials in Jenkins**
```
Jenkins → Manage Jenkins → Credentials → Add Credentials
  Kind: Username with password
  ID:   docker-hub-creds
  Username: myuser
  Password: ****
```

**Step 2: Use in Jenkinsfile**

```groovy
pipeline {
    agent any
    
    environment {
        // For username/password credentials
        DOCKER_CREDS = credentials('docker-hub-creds')
        // Creates: DOCKER_CREDS_USR and DOCKER_CREDS_PSW
    }
    
    stages {
        stage('Docker Push') {
            steps {
                sh '''
                    echo "$DOCKER_CREDS_PSW" | docker login -u "$DOCKER_CREDS_USR" --password-stdin
                    docker push myapp:latest
                '''
            }
        }
    }
}
```

```groovy
// Using withCredentials block (more explicit)
stage('Deploy') {
    steps {
        withCredentials([
            usernamePassword(
                credentialsId: 'docker-hub-creds',
                usernameVariable: 'DOCKER_USER',
                passwordVariable: 'DOCKER_PASS'
            ),
            string(
                credentialsId: 'slack-webhook',
                variable: 'SLACK_URL'
            ),
            file(
                credentialsId: 'kubeconfig',
                variable: 'KUBECONFIG'
            )
        ]) {
            sh '''
                echo "$DOCKER_PASS" | docker login -u "$DOCKER_USER" --password-stdin
                docker push myapp:latest
                kubectl --kubeconfig="$KUBECONFIG" apply -f k8s/
            '''
        }
    }
}
```

### Credential Types

| Type | Use For | Variables Created |
|------|---------|-------------------|
| Username/Password | Docker registries, APIs | `_USR`, `_PSW` |
| Secret text | API tokens, webhook URLs | Single variable |
| Secret file | kubeconfig, certificates | File path |
| SSH key | Git clone, server access | Key file path |
| Certificate | TLS, client certificates | Keystore path |

---

## 🟡 Post-Build Actions

```groovy
post {
    always {
        // Always runs — cleanup
        cleanWs()
        junit '**/test-results/*.xml'
    }
    
    success {
        // Only on success
        slackSend(
            channel: '#deploys',
            color: 'good',
            message: "✅ ${env.JOB_NAME} #${env.BUILD_NUMBER} succeeded"
        )
        archiveArtifacts artifacts: 'dist/**', fingerprint: true
    }
    
    failure {
        // Only on failure
        slackSend(
            channel: '#deploys',
            color: 'danger',
            message: "❌ ${env.JOB_NAME} #${env.BUILD_NUMBER} FAILED\n${env.BUILD_URL}"
        )
        mail to: 'team@example.com',
             subject: "Build Failed: ${env.JOB_NAME}",
             body: "Check: ${env.BUILD_URL}"
    }
    
    unstable {
        // Tests failed but build succeeded
        slackSend(
            channel: '#deploys',
            color: 'warning',
            message: "⚠️ ${env.JOB_NAME} #${env.BUILD_NUMBER} has test failures"
        )
    }
    
    changed {
        // Status changed from previous build (e.g., failure → success)
        echo "Build status changed!"
    }
}
```

---

## 🟡 Shared Libraries

When multiple repos share pipeline code, use **Shared Libraries**.

```
# Shared library repo structure
jenkins-shared-library/
├── vars/
│   ├── buildDocker.groovy      # Global function: buildDocker()
│   ├── deployToK8s.groovy      # Global function: deployToK8s()
│   └── notifySlack.groovy      # Global function: notifySlack()
├── src/
│   └── com/
│       └── mycompany/
│           └── Pipeline.groovy # Full Groovy classes
└── resources/
    └── templates/
        └── Dockerfile.template
```

```groovy
// vars/buildDocker.groovy
def call(Map config) {
    def image = config.image ?: error("image is required")
    def tag   = config.tag   ?: env.BUILD_NUMBER
    
    sh """
        docker build -t ${image}:${tag} .
        docker push ${image}:${tag}
    """
    
    return "${image}:${tag}"
}
```

```groovy
// Jenkinsfile using shared library
@Library('my-shared-lib') _

pipeline {
    agent { label 'linux' }
    
    stages {
        stage('Build') {
            steps {
                script {
                    def imageTag = buildDocker(
                        image: 'registry.example.com/my-app'
                    )
                    deployToK8s(
                        image: imageTag,
                        namespace: 'staging'
                    )
                }
            }
        }
    }
    
    post {
        success { notifySlack(status: 'success') }
        failure { notifySlack(status: 'failure') }
    }
}
```

---

## 🟡 Multi-Branch Pipelines

A **multibranch pipeline** automatically discovers branches and creates pipelines for each.

```
Repository:
├── main        → Jenkins pipeline (deploy to prod)
├── develop     → Jenkins pipeline (deploy to staging)
├── feature/x   → Jenkins pipeline (run tests only)
└── feature/y   → Jenkins pipeline (run tests only)
```

```groovy
// Jenkinsfile that behaves differently per branch
pipeline {
    agent { label 'linux' }
    
    stages {
        stage('Build & Test') {
            steps {
                sh 'make build test'
            }
        }
        
        stage('Deploy to Staging') {
            when { branch 'develop' }
            steps {
                sh 'make deploy-staging'
            }
        }
        
        stage('Deploy to Production') {
            when { branch 'main' }
            steps {
                input message: 'Deploy to production?'    // Manual approval gate
                sh 'make deploy-prod'
            }
        }
    }
}
```

---

## 🔴 Jenkinsfile Anti-Patterns

### Anti-Pattern 1: God Pipeline

```groovy
// ❌ Everything in one stage
stage('Do Everything') {
    steps {
        sh 'npm ci && npm run build && npm test && npm run lint && docker build . && docker push && kubectl apply -f k8s/'
    }
}

// ✅ Separate stages — clear, debuggable
stage('Install')  { steps { sh 'npm ci' } }
stage('Build')    { steps { sh 'npm run build' } }
stage('Test')     { steps { sh 'npm test' } }
stage('Lint')     { steps { sh 'npm run lint' } }
stage('Docker')   { steps { sh 'docker build .' } }
stage('Deploy')   { steps { sh 'kubectl apply -f k8s/' } }
```

### Anti-Pattern 2: Hardcoded Values

```groovy
// ❌ Hardcoded everywhere
sh 'docker build -t registry.example.com/my-app:1.2.3 .'
sh 'kubectl -n production set image deployment/my-app my-app=registry.example.com/my-app:1.2.3'

// ✅ Use variables
environment {
    REGISTRY = 'registry.example.com'
    APP_NAME = 'my-app'
    VERSION  = "${env.BUILD_NUMBER}"
}
// ...
sh "docker build -t ${REGISTRY}/${APP_NAME}:${VERSION} ."
sh "kubectl -n production set image deployment/${APP_NAME} ${APP_NAME}=${REGISTRY}/${APP_NAME}:${VERSION}"
```

### Anti-Pattern 3: No Timeout

```groovy
// ❌ Build hangs forever, nobody notices until disk is full
pipeline {
    agent any
    stages { ... }
}

// ✅ Pipeline times out and fails
pipeline {
    agent any
    options {
        timeout(time: 30, unit: 'MINUTES')
    }
    stages { ... }
}
```

---

**Previous:** [01. Jenkins Architecture](./01-jenkins-architecture.md)  
**Next:** [03. Building TypeScript/Go Apps](./03-building-apps.md)
