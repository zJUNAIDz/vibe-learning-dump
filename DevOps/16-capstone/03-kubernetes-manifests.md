# Capstone Phase 03: Kubernetes Manifests

> **Your application runs in Docker. Now make it run in Kubernetes — properly. With resource limits, health checks, security context, and everything you'd need in production.**

---

## 🟢 Full Manifest Set

```
k8s/
├── namespace.yaml
├── deployment.yaml
├── service.yaml
├── ingress.yaml
├── configmap.yaml
├── secret.yaml
├── hpa.yaml
└── pdb.yaml
```

---

## 🟢 Namespace

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: task-service
  labels:
    app: task-service
    # Pod Security Standard — enforce restricted
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/warn: restricted
```

---

## 🟢 Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: task-service
  namespace: task-service
  labels:
    app: task-service
    version: v1
spec:
  replicas: 3
  
  # Rolling update strategy
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0      # Never fewer pods than desired
      maxSurge: 1             # One extra pod during rollout
  minReadySeconds: 10         # Wait 10s after ready before continuing
  
  selector:
    matchLabels:
      app: task-service
  
  template:
    metadata:
      labels:
        app: task-service
        version: v1
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "3000"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: task-service
      automountServiceAccountToken: false
      
      # Security context (pod level)
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        runAsGroup: 1000
        fsGroup: 1000
        seccompProfile:
          type: RuntimeDefault
      
      # Spread across nodes
      topologySpreadConstraints:
        - maxSkew: 1
          topologyKey: kubernetes.io/hostname
          whenUnsatisfiable: DoNotSchedule
          labelSelector:
            matchLabels:
              app: task-service
      
      containers:
        - name: task-service
          image: ghcr.io/myorg/task-service:REPLACE_ME
          ports:
            - name: http
              containerPort: 3000
              protocol: TCP
          
          # Environment from ConfigMap and Secret
          envFrom:
            - configMapRef:
                name: task-service-config
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: task-service-secrets
                  key: database-url
          
          # Security context (container level)
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop: ["ALL"]
          
          # Resource limits
          resources:
            requests:
              memory: "128Mi"
              cpu: "50m"
            limits:
              memory: "256Mi"
              cpu: "200m"
          
          # Readiness probe — can this pod handle traffic?
          readinessProbe:
            httpGet:
              path: /health/ready
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
            failureThreshold: 3
            successThreshold: 1
          
          # Liveness probe — is this pod alive?
          livenessProbe:
            httpGet:
              path: /health/live
              port: http
            initialDelaySeconds: 15
            periodSeconds: 20
            failureThreshold: 3
          
          # Startup probe — has this pod finished starting?
          startupProbe:
            httpGet:
              path: /health/live
              port: http
            periodSeconds: 5
            failureThreshold: 12  # 12 × 5s = 60s max startup
          
          # Tmp volume for writable directory
          volumeMounts:
            - name: tmp
              mountPath: /tmp
      
      volumes:
        - name: tmp
          emptyDir: {}
```

---

## 🟢 Service

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: task-service
  namespace: task-service
  labels:
    app: task-service
spec:
  type: ClusterIP
  selector:
    app: task-service
  ports:
    - name: http
      port: 80
      targetPort: http
      protocol: TCP
```

---

## 🟢 ConfigMap and Secret

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: task-service-config
  namespace: task-service
data:
  NODE_ENV: "production"
  PORT: "3000"
  LOG_LEVEL: "info"

---
# k8s/secret.yaml
# For demonstration only — in real projects use Sealed Secrets
# or External Secrets Operator
apiVersion: v1
kind: Secret
metadata:
  name: task-service-secrets
  namespace: task-service
type: Opaque
stringData:
  database-url: "postgres://user:password@db:5432/tasks"
```

---

## 🟢 ServiceAccount

```yaml
# k8s/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: task-service
  namespace: task-service
automountServiceAccountToken: false
```

---

## 🟡 Ingress

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: task-service
  namespace: task-service
  annotations:
    # nginx ingress controller
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    # cert-manager for automatic TLS
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - tasks.example.com
      secretName: task-service-tls
  rules:
    - host: tasks.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: task-service
                port:
                  number: 80
```

---

## 🟡 Horizontal Pod Autoscaler

```yaml
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: task-service
  namespace: task-service
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: task-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60  # Wait before scaling up
      policies:
        - type: Pods
          value: 2
          periodSeconds: 60           # Add max 2 pods per minute
    scaleDown:
      stabilizationWindowSeconds: 300 # Wait 5 min before scaling down
      policies:
        - type: Pods
          value: 1
          periodSeconds: 120          # Remove max 1 pod per 2 min
```

---

## 🟡 Pod Disruption Budget

```yaml
# k8s/pdb.yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: task-service
  namespace: task-service
spec:
  minAvailable: 2  # Always keep at least 2 pods running
  selector:
    matchLabels:
      app: task-service
```

---

## 🟡 ServiceMonitor (Prometheus)

```yaml
# k8s/servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: task-service
  namespace: task-service
  labels:
    release: monitoring  # Must match Prometheus operator selector
spec:
  selector:
    matchLabels:
      app: task-service
  endpoints:
    - port: http
      path: /metrics
      interval: 15s
```

---

## 🔴 Deploying and Verifying

```bash
# Apply all manifests
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/

# Verify deployment
kubectl get all -n task-service

# Check pods are running
kubectl get pods -n task-service
# NAME                           READY   STATUS    RESTARTS
# task-service-abc123-xyz1       1/1     Running   0
# task-service-abc123-xyz2       1/1     Running   0
# task-service-abc123-xyz3       1/1     Running   0

# Check health
kubectl exec -n task-service deploy/task-service -- \
  wget -qO- http://localhost:3000/health/ready
# {"status":"ready"}

# Port forward for local testing
kubectl port-forward -n task-service svc/task-service 3000:80

# Test API
curl http://localhost:3000/api/tasks
curl http://localhost:3000/metrics

# Test rollback
kubectl rollout undo deployment/task-service -n task-service
kubectl rollout status deployment/task-service -n task-service
```

---

## 🔴 Checklist

```
□ Namespace created with pod security labels
□ Deployment with 3+ replicas
□ Rolling update with maxUnavailable: 0
□ Non-root security context (runAsNonRoot, drop ALL capabilities)
□ Read-only root filesystem
□ Resource requests AND limits
□ Readiness + liveness + startup probes
□ ServiceAccount (not default)
□ ConfigMap for non-sensitive config
□ Secret for sensitive data (or External Secrets)
□ Service (ClusterIP)
□ Ingress with TLS
□ HPA for autoscaling
□ PDB for availability during disruptions
□ ServiceMonitor for Prometheus scraping
□ topologySpreadConstraints for node distribution
```

---

**Previous:** [02. Containerization](./02-containerization.md)  
**Next:** [04. CI/CD Pipeline](./04-cicd-pipeline.md)
