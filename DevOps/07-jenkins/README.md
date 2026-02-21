# Module 07: Jenkins Deep Dive

> **Jenkins is not just a CI tool — it's a programmable automation server**

---

## What Is Jenkins?

Jenkins is an **automation server** that runs pipelines (sequences of steps).

**Common uses:**
- Build and test code
- Deploy applications
- Run scheduled tasks
- Orchestrate complex workflows

---

## Topics Covered

### 📝 01. Jenkins Architecture
- Master/controller
- Agents/nodes
- Executors
- Plugins (the Jenkins ecosystem)

### 📝 02. Jenkinsfile (Pipeline as Code)
- Declarative vs scripted pipelines
- Stages and steps
- Environment variables
- Credentials management

### 📝 03. Building TypeScript/Go Apps
- Node.js pipeline
- Go pipeline
- Docker builds
- Parallel stages

### 📝 04. Integration with Kubernetes
- Kubernetes plugin
- Dynamic agents
- Deploying to K8s from Jenkins

### 📝 05. Common Jenkins Disasters
- Plugin hell
- Disk space exhaustion
- Credential leaks
- Zombie builds

---

## Declarative Pipeline Example

```groovy
pipeline {
    agent any
    
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
        
        stage('Docker') {
            steps {
                sh 'docker build -t myapp:${BUILD_NUMBER} .'
            }
        }
        
        stage('Deploy') {
            steps {
                sh 'kubectl set image deployment/myapp myapp=myapp:${BUILD_NUMBER}'
            }
        }
    }
}
```

---

**Previous:** [06. CI/CD Fundamentals](../06-ci-cd-fundamentals/)  
**Next:** [08. Infrastructure as Code](../08-iac-fundamentals/)
