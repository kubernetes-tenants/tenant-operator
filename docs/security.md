# Security Guide

Security best practices for Tenant Operator.

## Credentials Management

### Database Credentials

Always use Kubernetes Secrets for sensitive data:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mysql-credentials
  namespace: default
type: Opaque
stringData:
  password: your-secure-password
```

Reference in TenantRegistry:

```yaml
spec:
  source:
    mysql:
      passwordRef:
        name: mysql-credentials
        key: password
```

**Never hardcode credentials in CRDs!**

### Rotating Credentials

1. Update Secret:
```bash
kubectl create secret generic mysql-credentials \
  --from-literal=password=new-password \
  --dry-run=client -o yaml | kubectl apply -f -
```

2. Operator automatically detects change and reconnects.

## RBAC

### Operator Permissions

The operator requires:

**CRD Management:**
- `tenantregistries`, `tenanttemplates`, `tenants`: All verbs

**Resource Management:**
- Managed resources (Deployments, Services, etc.): All verbs in target namespaces
- `namespaces`: Create, list, watch, get (cluster-scoped)

**Supporting Resources:**
- `events`: Create, patch
- `leases`: Get, create, update (for leader election)
- `secrets`: Get, list, watch (for credentials, namespace-scoped)

### Least Privilege

Scope RBAC to specific namespaces when possible:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role  # Not ClusterRole
metadata:
  name: tenant-operator-role
  namespace: production  # Specific namespace
rules:
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["*"]
```

### Service Account

Default service account: `tenant-operator-controller-manager`

Custom service account:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: custom-sa
  namespace: tenant-operator-system
---
apiVersion: v1
kind: Pod
spec:
  serviceAccountName: custom-sa
```

## Multi-Tenancy Isolation

TODO

## Data Security

### Sensitive Data in Templates

Avoid storing sensitive data in database columns. Instead:

1. Store only references:
```sql
-- Good
api_key_ref = "secret-acme-api-key"

-- Bad
api_key = "sk-abc123..."
```

2. Reference Secrets in templates:
```yaml
env:
- name: API_KEY
  valueFrom:
    secretKeyRef:
      name: "{{ .uid }}-secrets"
      key: api-key
```

## Audit Logging

### Enable Audit Logs

Configure Kubernetes audit policy:

```yaml
# audit-policy.yaml
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
- level: RequestResponse
  resources:
  - group: "operator.kubernetes-tenants.org"
    resources: ["tenantregistries", "tenanttemplates", "tenants"]
```

### Track Changes

Monitor events:

```bash
kubectl get events --all-namespaces | grep Tenant
```

## Compliance

### Data Retention

Configure deletion policies for compliance:

```yaml
persistentVolumeClaims:
  - id: data
    deletionPolicy: Retain  # Keep data after tenant deletion
```

### Immutable Resources

Use `CreationPolicy: Once` for audit resources:

```yaml
configMaps:
  - id: audit-log
    creationPolicy: Once  # Never update
```

## Vulnerability Management

### Container Scanning

Scan operator images:

```bash
# Using Trivy
trivy image ghcr.io/kubernetes-tenants/tenant-operator:latest

# Using Snyk
snyk container test ghcr.io/kubernetes-tenants/tenant-operator:latest
```

### Dependency Updates

Keep dependencies updated:

```bash
# Update Go dependencies
go get -u ./...
go mod tidy

# Check for vulnerabilities
go list -json -m all | nancy sleuth
```

## Best Practices

1. **Never hardcode credentials** - Use Secrets with SecretRef
2. **Enforce least privilege** - Scope RBAC to specific namespaces
3. **Apply security contexts** - Run as non-root, drop capabilities
4. **Enable audit logging** - Track all CRD changes
5. **Scan container images** - Regular vulnerability scanning
6. **Rotate credentials** - Regular password rotation
7. **Apply network policies** - Isolate tenant traffic
8. **Enforce resource quotas** - Prevent resource exhaustion

## See Also

- [Configuration Guide](configuration.md)
- [Installation Guide](installation.md)
- [RBAC Documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
