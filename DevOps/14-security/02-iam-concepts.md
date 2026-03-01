# IAM Concepts

> **Authentication asks "WHO are you?" Authorization asks "WHAT are you allowed to do?" Get these wrong and either nobody can work, or everyone can destroy everything.**

---

## 🟢 Authentication vs Authorization

```
AUTHENTICATION (AuthN) — Proving identity
  → "I am Alice" (here's my password/certificate/token)
  → Methods: passwords, SSH keys, OAuth tokens, certificates, MFA
  → Result: verified identity

AUTHORIZATION (AuthZ) — Checking permissions
  → "Alice wants to delete the production database"
  → "Is Alice allowed to do that?"
  → Result: allow or deny

Order matters:
  1. First: WHO are you? (authentication)
  2. Then: WHAT can you do? (authorization)

Examples:
  → kubectl: certificate authenticates you → RBAC authorizes actions
  → AWS: access key authenticates → IAM policy authorizes
  → SSH: key authenticates → file permissions authorize
```

---

## 🟢 Principle of Least Privilege

```
Give every user, service, and process ONLY the permissions
they need to do their job. Nothing more.

Why:
  → Limits blast radius of compromises
  → Prevents accidental damage
  → Makes auditing easier
  → Compliance requirement (SOC 2, ISO 27001, etc.)

Examples:

BAD:
  → Every developer has admin access to production
  → CI/CD pipeline runs as root
  → Service account has full database access
  → "We gave it * permissions because it was easier"

GOOD:
  → Developers have read-only access to production
  → CI/CD pipeline has deploy-only permissions
  → Service account has SELECT on specific tables
  → Permissions match exactly what's needed
```

---

## 🟢 RBAC (Role-Based Access Control)

```
Instead of assigning permissions to individual users,
you assign permissions to ROLES, then give users roles.

Without RBAC:
  Alice → can deploy, can read logs, can restart pods
  Bob → can deploy, can read logs, can restart pods
  Charlie → can deploy, can read logs, can restart pods
  (If you add a permission, update EVERY user)

With RBAC:
  Role: "developer" → can deploy, can read logs, can restart pods
  Alice → developer
  Bob → developer
  Charlie → developer
  (Change the role, all users get updated)
```

---

## 🟡 Kubernetes RBAC

### Core Concepts

```
RBAC in Kubernetes has 4 objects:

1. Role (namespace-scoped permissions)
   → "Can read pods in the 'app' namespace"

2. ClusterRole (cluster-wide permissions)
   → "Can read pods in ALL namespaces"
   → "Can manage nodes" (cluster-level resource)

3. RoleBinding (assigns Role to user/group/service account)
   → "Alice gets the 'pod-reader' Role in 'app' namespace"

4. ClusterRoleBinding (assigns ClusterRole cluster-wide)
   → "The 'admin' group gets the 'cluster-admin' ClusterRole"

        User/Group/ServiceAccount
                │
                │ bound via
                ▼
         RoleBinding / ClusterRoleBinding
                │
                │ references
                ▼
          Role / ClusterRole
                │
                │ defines
                ▼
         Permissions (verbs on resources)
```

### Creating Roles

```yaml
# Role: can read pods and logs in 'app' namespace
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: pod-reader
  namespace: app
rules:
  - apiGroups: [""]            # Core API group
    resources: ["pods"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["pods/log"]
    verbs: ["get"]

---
# ClusterRole: can read pods in ALL namespaces
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: pod-reader-global
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list", "watch"]

---
# Role: developer role with deploy capabilities
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: developer
  namespace: app
rules:
  - apiGroups: [""]
    resources: ["pods", "services", "configmaps"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]
  - apiGroups: [""]
    resources: ["pods/log", "pods/exec"]
    verbs: ["get", "create"]
  # Explicitly NO access to:
  #   → secrets (can't read passwords)
  #   → nodes (can't touch infrastructure)
  #   → namespaces (can't create/delete namespaces)
```

### Binding Roles

```yaml
# Bind Role to a user
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: alice-pod-reader
  namespace: app
subjects:
  - kind: User
    name: alice
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: pod-reader
  apiGroup: rbac.authorization.k8s.io

---
# Bind Role to a group (from OIDC/LDAP)
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: developers-binding
  namespace: app
subjects:
  - kind: Group
    name: developers
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: developer
  apiGroup: rbac.authorization.k8s.io

---
# Bind to a ServiceAccount (for pods/CI/CD)
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: ci-deployer
  namespace: app
subjects:
  - kind: ServiceAccount
    name: github-actions
    namespace: ci
roleRef:
  kind: Role
  name: developer
  apiGroup: rbac.authorization.k8s.io
```

### Service Accounts for Applications

```yaml
# Every pod should have its own ServiceAccount
# Don't use the default ServiceAccount

apiVersion: v1
kind: ServiceAccount
metadata:
  name: order-service
  namespace: app
automountServiceAccountToken: false  # Don't mount unless needed

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: order-service
spec:
  template:
    spec:
      serviceAccountName: order-service
      automountServiceAccountToken: false  # Extra safety
      containers:
        - name: app
          image: order-service:latest
```

### Testing RBAC

```bash
# Check if a user can perform an action
kubectl auth can-i get pods --namespace app --as alice
# yes

kubectl auth can-i delete deployments --namespace app --as alice
# no

# Check what a ServiceAccount can do
kubectl auth can-i --list --as system:serviceaccount:app:order-service -n app

# Impersonate a user to test
kubectl get pods -n app --as alice
kubectl get secrets -n app --as alice
# Error: secrets is forbidden
```

---

## 🟡 AWS IAM (Comparison)

```
AWS IAM follows the same concepts, different implementation:

Kubernetes                  AWS
─────────                   ───
ServiceAccount    →         IAM Role
Role/ClusterRole  →         IAM Policy
RoleBinding       →         Role attachment
Namespace         →         Account / Resource ARN
```

### IAM Policy Example

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowS3ReadOnly",
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::my-bucket",
        "arn:aws:s3:::my-bucket/*"
      ]
    },
    {
      "Sid": "DenyDeleteAnything",
      "Effect": "Deny",
      "Action": [
        "s3:DeleteObject",
        "s3:DeleteBucket"
      ],
      "Resource": "*"
    }
  ]
}
```

### IRSA (IAM Roles for Service Accounts)

```
Connect Kubernetes ServiceAccounts to AWS IAM Roles.
Pods get AWS permissions without static credentials.

Pod (ServiceAccount: order-service)
  → Kubernetes maps to IAM Role
  → Pod gets temporary AWS credentials
  → No access keys in environment variables!
```

```yaml
# Terraform — create IAM role for service account
module "order_service_irsa" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  
  role_name = "order-service"
  
  role_policy_arns = {
    s3 = aws_iam_policy.order_service_s3.arn
  }
  
  oidc_providers = {
    main = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["app:order-service"]
    }
  }
}

# Kubernetes — annotate ServiceAccount
apiVersion: v1
kind: ServiceAccount
metadata:
  name: order-service
  namespace: app
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::123456789:role/order-service
```

---

## 🔴 Common Mistakes

```
❌ cluster-admin for everyone
   → "It works" — until someone runs kubectl delete namespace production
   → Only cluster operators should have cluster-admin

❌ Default ServiceAccount
   → Every namespace has a "default" ServiceAccount
   → If it has permissions, EVERY pod in that namespace gets them
   → Always create specific ServiceAccounts per application

❌ No RBAC audit
   → Who has access to what? "We're not sure"
   → Run: kubectl auth can-i --list --as <user> regularly
   → Review bindings quarterly

❌ Long-lived static credentials
   → AWS access keys that never expire
   → Kubeconfig tokens that never rotate
   → Use short-lived tokens, OIDC, IRSA

❌ Over-permissive policies
   → Action: "*", Resource: "*"
   → Start with zero permissions, add what's needed
   → Use AWS Access Analyzer to find unused permissions
```

---

**Previous:** [01. Secrets Management](./01-secrets-management.md)  
**Next:** [03. Container Security](./03-container-security.md)
