# Jenkins + Kubernetes Integration

> **Dynamic build agents in Kubernetes: spin up a pod, run the build, destroy the pod. Clean, scalable, beautiful.**

---

## 🟢 Why Run Jenkins Agents in Kubernetes?

```
Traditional Jenkins:
  3 permanent agents × 24 hours = 72 agent-hours/day
  Actual build time: 8 hours/day
  Wasted: 64 agent-hours/day (89% idle!)
  Cost: $$$

Kubernetes Jenkins:
  Pods created on demand
  Build runs → pod exists for 5 minutes
  Build done → pod deleted
  Wasted: 0 agent-hours/day
  Cost: $ (pay only for what you use)
```

---

## 🟢 Kubernetes Plugin for Jenkins

### How It Works

```
1. Developer pushes code → Jenkins pipeline triggered
2. Jenkins controller sees: "I need an agent with label 'linux'"
3. Kubernetes plugin creates a Pod in the cluster
4. Pod starts, connects to Jenkins controller
5. Pipeline runs inside the pod
6. Pipeline finishes → pod is deleted
7. Clean, no leftover state
```

```
┌────────────────────┐
│  Jenkins Controller│
│  (running in K8s   │──────────────────────┐
│   or external)     │                      │
└────────────────────┘                      │
         │                                  │
         │ Creates pods via K8s API         │ Connects back
         ▼                                  │ (JNLP)
┌────────────────────────────────────┐      │
│        Kubernetes Cluster          │      │
│                                    │      │
│  ┌──────────┐  ┌──────────┐       │      │
│  │ Build    │  │ Build    │       │      │
│  │ Pod #1   │  │ Pod #2   │───────┘      │
│  │          │  │          │              │
│  │ jnlp     │  │ jnlp     │              │
│  │ container│  │ container│              │
│  │          │  │          │              │
│  │ build    │  │ build    │              │
│  │ container│  │ container│              │
│  └──────────┘  └──────────┘              │
│                                    │      │
└────────────────────────────────────┘      │
```

### Plugin Configuration

```yaml
# Jenkins running in Kubernetes — Helm values
jenkins:
  controller:
    image: jenkins/jenkins
    tag: "2.440-lts"
    resources:
      limits:
        cpu: "2"
        memory: "4Gi"
    
  agent:
    enabled: true
    image: jenkins/inbound-agent
    tag: "3206.vb_15dcf73f6a_9-1"
    resources:
      limits:
        cpu: "1"
        memory: "2Gi"
    
    # Pod template for dynamic agents
    podTemplates:
      - name: "default"
        label: "jenkins-agent"
        containers:
          - name: "jnlp"
            image: "jenkins/inbound-agent:latest"
            resourceRequestCpu: "500m"
            resourceRequestMemory: "512Mi"
```

---

## 🟢 Pod Templates in Jenkinsfile

### Basic Pod Template

```groovy
pipeline {
    agent {
        kubernetes {
            yaml '''
                apiVersion: v1
                kind: Pod
                metadata:
                  labels:
                    app: jenkins-agent
                spec:
                  containers:
                  - name: golang
                    image: golang:1.22-alpine
                    command: ["cat"]
                    tty: true
                    resources:
                      requests:
                        memory: "512Mi"
                        cpu: "500m"
                      limits:
                        memory: "1Gi"
                        cpu: "1"
            '''
        }
    }
    
    stages {
        stage('Build') {
            steps {
                container('golang') {
                    sh 'go version'
                    sh 'go build -o myapp ./cmd/server'
                }
            }
        }
    }
}
```

### Multi-Container Pod Template

```groovy
pipeline {
    agent {
        kubernetes {
            yaml '''
                apiVersion: v1
                kind: Pod
                spec:
                  containers:
                  - name: golang
                    image: golang:1.22-alpine
                    command: ["cat"]
                    tty: true
                  - name: docker
                    image: docker:24-cli
                    command: ["cat"]
                    tty: true
                    volumeMounts:
                    - name: docker-sock
                      mountPath: /var/run/docker.sock
                  - name: kubectl
                    image: bitnami/kubectl:1.29
                    command: ["cat"]
                    tty: true
                  volumes:
                  - name: docker-sock
                    hostPath:
                      path: /var/run/docker.sock
            '''
        }
    }
    
    stages {
        stage('Build') {
            steps {
                container('golang') {
                    sh 'go build -ldflags="-s -w" -o myapp ./cmd/server'
                }
            }
        }
        
        stage('Docker') {
            steps {
                container('docker') {
                    sh 'docker build -t myapp:${BUILD_NUMBER} .'
                    sh 'docker push registry.example.com/myapp:${BUILD_NUMBER}'
                }
            }
        }
        
        stage('Deploy') {
            steps {
                container('kubectl') {
                    sh '''
                        kubectl set image deployment/myapp \
                            myapp=registry.example.com/myapp:${BUILD_NUMBER}
                        kubectl rollout status deployment/myapp --timeout=300s
                    '''
                }
            }
        }
    }
}
```

**Why multi-container?** Each container has exactly the tools it needs. No bloated "do-everything" image.

---

## 🟡 Deploying TO Kubernetes FROM Jenkins

### Using kubectl

```groovy
stage('Deploy') {
    steps {
        withCredentials([file(credentialsId: 'kubeconfig', variable: 'KUBECONFIG')]) {
            sh '''
                # Update the image tag in deployment
                kubectl set image deployment/myapp \
                    myapp=registry.example.com/myapp:${BUILD_NUMBER} \
                    --namespace=production

                # Wait for rollout to complete
                kubectl rollout status deployment/myapp \
                    --namespace=production \
                    --timeout=300s

                # Verify pods are running
                kubectl get pods -n production -l app=myapp
            '''
        }
    }
}
```

### Using Kustomize

```groovy
stage('Deploy with Kustomize') {
    steps {
        container('kubectl') {
            sh '''
                cd k8s/overlays/production
                
                # Update image tag
                kustomize edit set image myapp=registry.example.com/myapp:${BUILD_NUMBER}
                
                # Apply
                kubectl apply -k .
                
                # Wait
                kubectl rollout status deployment/myapp -n production --timeout=300s
            '''
        }
    }
}
```

### Using Helm

```groovy
stage('Deploy with Helm') {
    steps {
        container('helm') {
            sh """
                helm upgrade --install myapp ./charts/myapp \
                    --namespace production \
                    --set image.tag=${BUILD_NUMBER} \
                    --set image.repository=registry.example.com/myapp \
                    --wait \
                    --timeout 5m
            """
        }
    }
}
```

---

## 🟡 Rolling Deployments from Jenkins

### Zero-Downtime Deploy Pattern

```groovy
stage('Deploy') {
    when { branch 'main' }
    steps {
        script {
            def previousVersion = sh(
                script: "kubectl get deployment myapp -n production -o jsonpath='{.spec.template.spec.containers[0].image}'",
                returnStdout: true
            ).trim()
            
            echo "Current version: ${previousVersion}"
            echo "Deploying version: ${BUILD_NUMBER}"
            
            try {
                sh """
                    kubectl set image deployment/myapp \
                        myapp=registry.example.com/myapp:${BUILD_NUMBER} \
                        -n production
                    
                    kubectl rollout status deployment/myapp \
                        -n production --timeout=300s
                """
                echo "✅ Deployment successful"
            } catch (err) {
                echo "❌ Deployment failed, rolling back to ${previousVersion}"
                sh """
                    kubectl rollout undo deployment/myapp -n production
                    kubectl rollout status deployment/myapp -n production --timeout=300s
                """
                error("Deployment failed and was rolled back")
            }
        }
    }
}
```

### Canary Deploy from Jenkins

```groovy
stage('Canary Deploy') {
    steps {
        script {
            // Deploy canary (10% of pods)
            sh """
                kubectl set image deployment/myapp-canary \
                    myapp=registry.example.com/myapp:${BUILD_NUMBER} \
                    -n production
                kubectl rollout status deployment/myapp-canary -n production
            """
            
            // Wait and check metrics
            sleep(time: 5, unit: 'MINUTES')
            
            def errorRate = sh(
                script: '''
                    curl -s "http://prometheus:9090/api/v1/query?query=rate(http_requests_total{status=~'5..'}[5m])/rate(http_requests_total[5m])" \
                    | jq -r '.data.result[0].value[1]'
                ''',
                returnStdout: true
            ).trim().toFloat()
            
            if (errorRate > 0.01) {
                echo "❌ Error rate ${errorRate} > 1%, rolling back canary"
                sh 'kubectl rollout undo deployment/myapp-canary -n production'
                error("Canary failed — error rate too high")
            }
            
            echo "✅ Canary healthy, promoting to full deployment"
            sh """
                kubectl set image deployment/myapp \
                    myapp=registry.example.com/myapp:${BUILD_NUMBER} \
                    -n production
                kubectl rollout status deployment/myapp -n production
            """
        }
    }
}
```

---

## 🟡 Running Jenkins Itself in Kubernetes

### Jenkins Helm Chart

```bash
# Install Jenkins in Kubernetes
helm repo add jenkins https://charts.jenkins.io
helm repo update

helm install jenkins jenkins/jenkins \
    --namespace jenkins \
    --create-namespace \
    --values jenkins-values.yaml
```

```yaml
# jenkins-values.yaml
controller:
  image: jenkins/jenkins
  tag: "2.440-lts"
  
  resources:
    requests:
      cpu: "1"
      memory: "2Gi"
    limits:
      cpu: "2"
      memory: "4Gi"
  
  # Persistent storage for Jenkins home
  persistence:
    enabled: true
    size: "50Gi"
    storageClass: "gp3"
  
  # Plugins to install
  installPlugins:
    - kubernetes:4029.v5712230ccb_f8
    - workflow-aggregator:596.v8c21c963d92d
    - git:5.2.0
    - configuration-as-code:1775.v810dc950b_514
  
  # JCasC configuration
  JCasC:
    configScripts:
      welcome-message: |
        jenkins:
          systemMessage: "Jenkins on Kubernetes"
          numExecutors: 0

agent:
  enabled: true
  podTemplates:
    default: |
      - name: default
        label: "jenkins-agent"
        serviceAccount: jenkins
        containers:
          - name: jnlp
            image: jenkins/inbound-agent:latest
            resourceRequestCpu: "500m"
            resourceRequestMemory: "512Mi"

serviceAccount:
  create: true
  name: jenkins
```

---

## 🔴 Common Jenkins + K8s Problems

### Problem 1: Pod Stuck in Pending

```
Build is stuck in queue: "Waiting for next available executor"
Pod is Pending: "0/5 nodes available: 5 Insufficient memory"

Cause: Cluster doesn't have enough resources for the build pod
Fix:
  1. Reduce resource requests in pod template
  2. Scale up the cluster (add nodes)
  3. Use cluster autoscaler
  4. Set proper resource requests (don't request 4Gi if you need 512Mi)
```

### Problem 2: JNLP Connection Timeout

```
Agent pod starts but can't connect to Jenkins controller

Causes:
  1. Network policy blocking JNLP port (50000)
  2. Controller URL not reachable from pod
  3. DNS resolution failure

Fix:
  - Ensure port 50000 is accessible
  - Use Kubernetes service name: jenkins.jenkins.svc.cluster.local
  - Check NetworkPolicies
```

### Problem 3: Workspace Persistence

```
Build #1: git clone, npm install (5 minutes)
Pod deleted.
Build #2: git clone, npm install (5 minutes) ← Same work repeated!

Solution: Use PersistentVolumeClaims for workspace caching

pipeline {
    agent {
        kubernetes {
            yaml '''
                spec:
                  containers:
                  - name: node
                    image: node:20-alpine
                    volumeMounts:
                    - name: npm-cache
                      mountPath: /root/.npm
                  volumes:
                  - name: npm-cache
                    persistentVolumeClaim:
                      claimName: jenkins-npm-cache
            '''
        }
    }
}
```

---

**Previous:** [03. Building TypeScript/Go Apps](./03-building-apps.md)  
**Next:** [05. Common Jenkins Disasters](./05-common-jenkins-disasters.md)
