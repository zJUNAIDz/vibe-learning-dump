# Kubernetes Outages

> **Kubernetes is reliable — until it isn't. When a K8s cluster fails, it fails in ways that are hard to diagnose, hard to fix, and affect every service running on it.**

---

## 🟢 Cascading Pod Failures

### The Domino Effect

```
Scenario: One service starts failing, takes down everything.

Timeline:
  09:00 — Payment service deploys with memory leak
  09:15 — Payment pods using 2x normal memory
  09:30 — First payment pod OOM killed
  09:31 — Remaining payment pods handle more traffic → leak faster
  09:33 — All payment pods OOM killed → CrashLoopBackOff
  09:34 — Order service calls payment service → timeout
  09:35 — Order service goroutine/thread pool fills up waiting
  09:36 — Order service stops responding to health checks
  09:37 — Order service pods killed (liveness probe failure)
  09:38 — API gateway has no backends → 502 to all users
  09:39 — Every service that depends on orders starts failing
  09:45 — Complete cluster meltdown

Root cause: No circuit breaker. No resource isolation.
```

### Prevention

```yaml
# 1. Resource limits on EVERY pod
resources:
  limits:
    memory: "512Mi"  # OOM kill only THIS pod, not the node
    cpu: "500m"       # Throttle, don't starve neighbors

# 2. PodDisruptionBudget
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: payment-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: payment-service

# 3. Anti-affinity (spread pods across nodes)
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchLabels:
              app: payment-service
          topologyKey: kubernetes.io/hostname
```

```typescript
// 4. Circuit breaker in application code
import CircuitBreaker from 'opossum';

const paymentBreaker = new CircuitBreaker(callPaymentService, {
  timeout: 3000,        // 3 second timeout
  errorThresholdPercentage: 50,  // Open circuit at 50% errors
  resetTimeout: 10000,  // Try again after 10 seconds
});

paymentBreaker.fallback(() => {
  return { status: 'pending', message: 'Payment queued for retry' };
});
```

---

## 🟡 DNS Failures

### CoreDNS Goes Down

```
Kubernetes uses CoreDNS (or kube-dns) for all service discovery.
If DNS fails, NOTHING can communicate.

Symptoms:
  → Pods can't resolve service names
  → "dial tcp: lookup payment-service.app.svc.cluster.local: no such host"
  → Health checks fail (if they resolve DNS)
  → Cascading failures across all services

Common causes:
  → CoreDNS pods OOM killed (too many queries)
  → CoreDNS ConfigMap misconfigured
  → Node running CoreDNS pods goes down (if no affinity spread)
  → Network policy blocking DNS traffic
```

### Prevention

```yaml
# 1. Multiple CoreDNS replicas across nodes
apiVersion: apps/v1
kind: Deployment
metadata:
  name: coredns
  namespace: kube-system
spec:
  replicas: 3  # At least 2, ideally 3
  template:
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            - topologyKey: kubernetes.io/hostname
              labelSelector:
                matchLabels:
                  k8s-app: kube-dns

# 2. Resource limits for CoreDNS
resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
  limits:
    memory: "256Mi"
    cpu: "200m"

# 3. Monitor CoreDNS
# Alert: coredns_dns_requests_total rate drops to 0
# Alert: coredns_dns_responses_total{rcode="SERVFAIL"} increases
```

### Debugging DNS

```bash
# Test DNS from inside a pod
kubectl run dns-test --image=busybox:1.28 --rm -it -- nslookup payment-service.app.svc.cluster.local

# Check CoreDNS pods
kubectl get pods -n kube-system -l k8s-app=kube-dns

# Check CoreDNS logs
kubectl logs -n kube-system -l k8s-app=kube-dns --tail=50

# Check CoreDNS config
kubectl get configmap coredns -n kube-system -o yaml

# Verify DNS policy on pod
kubectl get pod my-pod -o yaml | grep -A5 dnsPolicy
```

---

## 🟡 Network Policy Lockout

### Locking Yourself Out

```
Scenario: Security team applies strict NetworkPolicy

apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all
  namespace: app
spec:
  podSelector: {}  # Applies to ALL pods
  policyTypes:
    - Ingress
    - Egress
  # No ingress/egress rules = DENY EVERYTHING

Result:
  → Pods can't talk to each other
  → Pods can't reach DNS (port 53 blocked!)
  → Pods can't reach external APIs
  → Health checks fail (kubelet can't reach pod)
  → ALL pods marked as unhealthy
  → Complete namespace outage
```

### Correct NetworkPolicy

```yaml
# Start with deny-all, then open specific paths
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all
  namespace: app
spec:
  podSelector: {}
  policyTypes:
    - Ingress
    - Egress

---
# Allow DNS (CRITICAL — always allow this)
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-dns
  namespace: app
spec:
  podSelector: {}
  policyTypes:
    - Egress
  egress:
    - to: []
      ports:
        - port: 53
          protocol: UDP
        - port: 53
          protocol: TCP

---
# Allow specific service communication
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-order-to-payment
  namespace: app
spec:
  podSelector:
    matchLabels:
      app: payment-service
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: order-service
      ports:
        - port: 3000
```

---

## 🔴 etcd Corruption

### What Is etcd

```
etcd is Kubernetes' brain.
It stores ALL cluster state:
  → Pod specs, deployments, services
  → ConfigMaps, Secrets
  → RBAC rules
  → Custom resources

If etcd is corrupted or lost:
  → Cluster loses ALL configuration
  → Running pods keep running (kubelet cache)
  → But NOTHING can be created, updated, or deleted
  → No new pods, no deployments, no scaling
  → Effectively: cluster is brain-dead
```

### How Corruption Happens

```
1. Disk full on etcd node
   → etcd can't write → leader election fails
   → Split brain between etcd members
   → Data diverges → corruption

2. Force killing etcd process
   → kill -9 during write → partial write
   → WAL (write-ahead log) corrupted

3. Clock skew between etcd members
   → Different members disagree on time
   → Raft consensus breaks

4. Running etcd on slow/unreliable disk
   → etcd needs < 10ms fsync
   → Network-attached storage too slow
   → Missed heartbeats → leader election storm
```

### Prevention and Recovery

```bash
# 1. Backup etcd regularly
ETCDCTL_API=3 etcdctl snapshot save /backup/etcd-$(date +%Y%m%d).db \
  --endpoints=https://127.0.0.1:2379 \
  --cacert=/etc/kubernetes/pki/etcd/ca.crt \
  --cert=/etc/kubernetes/pki/etcd/server.crt \
  --key=/etc/kubernetes/pki/etcd/server.key

# Verify backup
ETCDCTL_API=3 etcdctl snapshot status /backup/etcd-20240115.db

# 2. Monitor etcd health
etcdctl endpoint health
etcdctl endpoint status --write-out=table

# 3. Restore from backup (emergency)
ETCDCTL_API=3 etcdctl snapshot restore /backup/etcd-20240115.db \
  --data-dir=/var/lib/etcd-restored

# 4. Automate backups with a CronJob
# Run every 6 hours, keep last 7 days
```

```yaml
# etcd backup CronJob
apiVersion: batch/v1
kind: CronJob
metadata:
  name: etcd-backup
  namespace: kube-system
spec:
  schedule: "0 */6 * * *"  # Every 6 hours
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: backup
              image: bitnami/etcd:latest
              command:
                - /bin/sh
                - -c
                - |
                  etcdctl snapshot save /backup/etcd-$(date +%Y%m%d-%H%M).db
                  # Upload to S3
                  aws s3 cp /backup/ s3://my-backups/etcd/ --recursive
                  # Clean old local backups
                  find /backup -mtime +7 -delete
              volumeMounts:
                - name: backup-volume
                  mountPath: /backup
          restartPolicy: OnFailure
          volumes:
            - name: backup-volume
              persistentVolumeClaim:
                claimName: etcd-backup-pvc
```

---

## 🔴 Node Failures

```
Scenario: A node dies

What happens:
  1. Node stops responding to API server
  2. API server waits (node-monitor-grace-period: 40s default)
  3. Node marked NotReady
  4. Controller waits (pod-eviction-timeout: 5m default)
  5. Pods on dead node marked for eviction
  6. Scheduler creates replacement pods on healthy nodes
  7. BUT: if pods have PVCs, they might wait for volume detach
  8. Volume detach timeout: 6 minutes
  
  Total recovery time: ~12 minutes for pod rescheduling

  If you only had pods on ONE node:
    → 12 minutes of complete downtime

Prevention:
  → Run at least 3 nodes
  → Spread pods across nodes (pod anti-affinity)
  → PodDisruptionBudget to prevent all pods on one node
  → Monitor node health
```

---

## 🔴 Debugging Checklist

```bash
# When a K8s cluster is misbehaving:

# 1. Check node health
kubectl get nodes
kubectl describe node <problem-node>
kubectl top nodes

# 2. Check system pods
kubectl get pods -n kube-system
kubectl logs -n kube-system <coredns-pod>
kubectl logs -n kube-system <kube-proxy-pod>

# 3. Check events (recent cluster events)
kubectl get events --sort-by='.lastTimestamp' -A | tail -50

# 4. Check specific workload
kubectl get pods -n <namespace>
kubectl describe pod <pod> -n <namespace>
kubectl logs <pod> -n <namespace> --previous  # Previous crash

# 5. Check resources
kubectl top pods -n <namespace>
kubectl describe node | grep -A5 "Allocated resources"

# 6. Check networking
kubectl run test --image=busybox --rm -it -- wget -qO- http://service-name:port
kubectl get endpoints <service>
kubectl get networkpolicies -n <namespace>

# 7. Check etcd (control plane issue)
kubectl get --raw /healthz
kubectl get componentstatuses  # deprecated but sometimes useful
```

---

**Previous:** [03. CI Pipeline Disasters](./03-ci-pipeline-disasters.md)  
**Next:** [05. Terraform State Corruption](./05-terraform-state-corruption.md)
