# Tenant Operator v1.0 Migration Guide

This document provides a comprehensive guide for migrating from the legacy tenant-operator (v0.x) to the new v1.0 architecture.

## Table of Contents

1. [Overview](#overview)
2. [Key Architectural Changes](#key-architectural-changes)
3. [CRD Mapping](#crd-mapping)
4. [Migration Steps](#migration-steps)
5. [Template Conversion](#template-conversion)
6. [Breaking Changes](#breaking-changes)
7. [New Features](#new-features)
8. [Rollback Strategy](#rollback-strategy)

---

## Overview

The v1.0 release represents a complete architectural redesign focused on:

- **Template-driven provisioning** with Go templates + Sprig functions
- **Server-Side Apply (SSA)** for declarative resource management
- **Policy-based lifecycle management** (creation, deletion, conflict policies)
- **Dependency graph** with resource ordering and readiness checks
- **Stronger consistency guarantees** between database and Kubernetes state

### Repository Information

- **Legacy**: `github.com/Ecube-Labs/tenant-operator`
- **New**: `github.com/kubernetes-tenants/tenant-operator`
- **Domain**: Changed from `ecubelabs.com` to `tenants.ecube.dev`

---

## Key Architectural Changes

### 1. CRD Changes

| Aspect | v0.x | v1.0 |
|--------|------|------|
| **Registry** | TenantPool | TenantRegistry |
| **Template** | (Inline in TenantPool) | TenantTemplate (separate CRD) |
| **Instance** | Tenant | Tenant |
| **Provisioning** | ProvisioningRequest | (Integrated into Tenant) |

### 2. Template System

| Feature | v0.x | v1.0 |
|---------|------|------|
| **Syntax** | String substitution (`${VAR}`) | Go text/template + Sprig |
| **Variables** | `TO_TENANT_ID`, `TO_TENANT_APP_HOST` | `.uid`, `.host`, `.hostOrUrl` + custom |
| **Functions** | None | 200+ Sprig functions + custom |

### 3. Resource Management

| Aspect | v0.x | v1.0 |
|--------|------|------|
| **Apply Method** | Direct creation | Server-Side Apply (SSA) |
| **Ownership** | ProvisioningRequest owns all | Tenant owns all |
| **Lifecycle** | Job-based provisioning | Direct reconciliation |
| **Dependencies** | None | DAG-based ordering |

### 4. Policies (New in v1.0)

- **CreationPolicy**: `Once` (one-time) vs `WhenNeeded` (continuous)
- **DeletionPolicy**: `Delete` vs `Retain` (preserve resources)
- **ConflictPolicy**: `Stuck` (fail) vs `Force` (takeover)

---

## CRD Mapping

### TenantPool → TenantRegistry

**v0.x (TenantPool):**
```yaml
apiVersion: tenant.ecubelabs.com/v1alpha1
kind: TenantPool
metadata:
  name: haulla-hauler
spec:
  host: mysql.default.svc.cluster.local
  port: 3306
  auth:
    username: root
    password: "1234"  # ⚠️ Plain text
  database: tenant-pool
  table: hauler
  tenantIdColumn: id
  tenantHostColumn: url
  tenantActivateColumn: isActive
  tenants:
    - id: haulla-api
      image: my-image:latest
      # ... inline resource specs
```

**v1.0 (TenantRegistry):**
```yaml
apiVersion: tenants.ecube.dev/v1
kind: TenantRegistry
metadata:
  name: sample-registry
spec:
  source:
    type: mysql
    syncInterval: 30s
    mysql:
      host: mysql.default.svc.cluster.local
      port: 3306
      username: root
      passwordRef:  # ✅ Secret reference
        name: mysql-cred
        key: password
      database: tenant-pool
      table: hauler
  valueMappings:
    uid: id
    hostOrUrl: url
    activate: isActive
  extraValueMappings:
    deployImage: image_column
    planId: plan_column
```

### TenantPool.tenants → TenantTemplate

**v0.x (Inline in TenantPool):**
```yaml
spec:
  tenants:
    - id: api
      image: "my-api:${TO_TENANT_ID}"
      host: "api-${TO_TENANT_APP_HOST}"
      replicas: 2
      env:
        - name: TENANT_ID
          value: "${TO_TENANT_ID}"
```

**v1.0 (TenantTemplate):**
```yaml
apiVersion: tenants.ecube.dev/v1
kind: TenantTemplate
metadata:
  name: webapp-template
spec:
  registryId: sample-registry
  deployments:
    - id: api
      nameTemplate: "{{ .uid }}-api"
      namespaceTemplate: "tenant-{{ .uid }}"
      spec:
        apiVersion: apps/v1
        kind: Deployment
        metadata:
          labels:
            tenant: "{{ .uid }}"
        spec:
          replicas: 2
          template:
            spec:
              containers:
                - name: api
                  image: "my-api:{{ .uid }}"
                  env:
                    - name: TENANT_ID
                      value: "{{ .uid }}"
                    - name: TENANT_HOST
                      value: "{{ .host }}"
```

### ProvisioningRequest → Tenant (Integrated)

**v0.x:**
- ProvisioningRequest runs a Job
- Job creates Deployments, Services, Ingresses
- Separate lifecycle

**v1.0:**
- Tenant directly manages all resources
- No intermediate Job (unless explicitly defined as a resource)
- Unified lifecycle with SSA

---

## Migration Steps

### Step 1: Backup Existing State

```bash
# Export existing resources
kubectl get tenantpool -A -o yaml > tenantpool-backup.yaml
kubectl get tenant -A -o yaml > tenant-backup.yaml
kubectl get provisioningrequest -A -o yaml > provreq-backup.yaml

# Export all deployments, services, ingresses created by old operator
kubectl get deploy,svc,ing -l tenant.ecubelabs.com/managed=true -A -o yaml > resources-backup.yaml
```

### Step 2: Install v1.0 CRDs (Parallel)

```bash
# Clone new repository
git clone https://github.com/kubernetes-tenants/tenant-operator.git
cd tenant-operator

# Install new CRDs (different group, no conflict)
make install

# Verify CRDs
kubectl get crd | grep tenants.ecube.dev
```

### Step 3: Create TenantRegistry

Convert each `TenantPool` to a `TenantRegistry`:

```bash
# Example conversion script
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: mysql-cred
  namespace: tenant-operator-system
type: Opaque
data:
  password: $(echo -n "1234" | base64)
---
apiVersion: tenants.ecube.dev/v1
kind: TenantRegistry
metadata:
  name: haulla-registry
  namespace: tenant-operator-system
spec:
  source:
    type: mysql
    syncInterval: 30s
    mysql:
      host: mysql.default.svc.cluster.local
      port: 3306
      username: root
      passwordRef:
        name: mysql-cred
        key: password
      database: tenant-pool
      table: hauler
  valueMappings:
    uid: id
    hostOrUrl: url
    activate: isActive
  extraValueMappings:
    deployImage: image
    planId: plan_id
EOF
```

### Step 4: Create TenantTemplate

Convert inline tenant specs to `TenantTemplate`:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: tenants.ecube.dev/v1
kind: TenantTemplate
metadata:
  name: haulla-template
  namespace: tenant-operator-system
spec:
  registryId: haulla-registry
  namespaces:
    - id: ns
      nameTemplate: "tenant-{{ .uid }}"
      spec:
        apiVersion: v1
        kind: Namespace
        metadata:
          labels:
            tenant: "{{ .uid }}"
  deployments:
    - id: api
      dependIds: [ns]
      nameTemplate: "{{ .uid }}-api"
      namespaceTemplate: "tenant-{{ .uid }}"
      waitForReady: true
      spec:
        apiVersion: apps/v1
        kind: Deployment
        # ... full Deployment spec with templates
  services:
    - id: svc
      dependIds: [api]
      # ... Service spec
  ingresses:
    - id: ing
      dependIds: [svc]
      # ... Ingress spec
EOF
```

### Step 5: Deploy v1.0 Operator

```bash
# Build and deploy
make docker-build IMG=your-registry/tenant-operator:v1.0
make docker-push IMG=your-registry/tenant-operator:v1.0
make deploy IMG=your-registry/tenant-operator:v1.0

# Verify deployment
kubectl get pods -n tenant-operator-system
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager
```

### Step 6: Verify Tenant Creation

The new operator will:
1. Read from TenantRegistry (MySQL)
2. Create Tenant CRs automatically
3. Apply resources via SSA

```bash
# Check Tenant CRs
kubectl get tenants.tenants.ecube.dev -A

# Check TenantRegistry status
kubectl get tenantregistry haulla-registry -o yaml

# Verify resources
kubectl get deploy,svc,ing -A | grep tenant-
```

### Step 7: Transfer Resource Ownership (Critical)

Existing resources created by v0.x operator need ownership transfer:

```bash
# For each existing Deployment/Service/Ingress, patch ownerReferences
# This requires a migration script (see below)
./scripts/transfer-ownership.sh
```

**Sample transfer-ownership.sh:**
```bash
#!/bin/bash
# Transfer ownership from ProvisioningRequest to new Tenant CR

OLD_NAMESPACE="tenant-operator-system"
NEW_NAMESPACE="tenant-operator-system"

# Get all Tenants in new system
for tenant_uid in $(kubectl get tenants.tenants.ecube.dev -n $NEW_NAMESPACE -o jsonpath='{.items[*].spec.uid}'); do
  echo "Processing tenant UID: $tenant_uid"

  # Find matching Deployment
  DEPLOY_NAME="${tenant_uid}-api"
  DEPLOY_NS="tenant-${tenant_uid}"

  if kubectl get deploy $DEPLOY_NAME -n $DEPLOY_NS &>/dev/null; then
    # Get Tenant CR metadata
    TENANT_NAME=$(kubectl get tenants.tenants.ecube.dev -n $NEW_NAMESPACE -o jsonpath="{.items[?(@.spec.uid=='$tenant_uid')].metadata.name}")
    TENANT_UID=$(kubectl get tenants.tenants.ecube.dev $TENANT_NAME -n $NEW_NAMESPACE -o jsonpath='{.metadata.uid}')

    # Patch ownerReferences
    kubectl patch deploy $DEPLOY_NAME -n $DEPLOY_NS --type=json -p="[
      {\"op\":\"replace\",\"path\":\"/metadata/ownerReferences\",\"value\":[{
        \"apiVersion\":\"tenants.ecube.dev/v1\",
        \"kind\":\"Tenant\",
        \"name\":\"$TENANT_NAME\",
        \"uid\":\"$TENANT_UID\",
        \"controller\":true,
        \"blockOwnerDeletion\":true
      }]}
    ]"

    echo "  ✓ Transferred ownership of $DEPLOY_NAME"
  fi
done
```

### Step 8: Scale Down Old Operator

```bash
# Scale down v0.x operator to prevent conflicts
kubectl scale deploy tenant-operator-controller-manager -n tenant-operator-system --replicas=0

# Verify no reconciliation loops
kubectl get events -n tenant-operator-system --sort-by='.lastTimestamp'
```

### Step 9: Cleanup (After Verification)

```bash
# Remove old CRDs (this will delete old CRs!)
kubectl delete crd tenantpools.tenant.ecubelabs.com
kubectl delete crd tenants.tenant.ecubelabs.com
kubectl delete crd provisioningrequests.tenant.ecubelabs.com

# Uninstall old operator
kubectl delete deploy tenant-operator-controller-manager -n tenant-operator-system
```

---

## Template Conversion

### Variable Mapping

| v0.x | v1.0 | Notes |
|------|------|-------|
| `${TO_TENANT_ID}` | `{{ .uid }}` | Tenant unique ID |
| `${TO_TENANT_APP_HOST}` | `{{ .host }}` | Auto-extracted from `.hostOrUrl` |
| `${TO_APP_HOST}` | `{{ .host }}` | Same as above |
| N/A | `{{ .hostOrUrl }}` | Raw URL/host value |
| (Custom columns) | `{{ .customKey }}` | Via `extraValueMappings` |

### Function Examples

**v0.x:**
```yaml
env:
  - name: DB_NAME
    value: "db-${TO_TENANT_ID}"
```

**v1.0 (simple):**
```yaml
env:
  - name: DB_NAME
    value: "db-{{ .uid }}"
```

**v1.0 (advanced):**
```yaml
env:
  - name: DB_NAME
    value: "{{ printf \"db-%s\" .uid | trunc 63 }}"
  - name: DB_PASSWORD
    value: "{{ .uid | sha1sum | trunc 16 }}"
  - name: HOSTNAME
    value: "{{ .hostOrUrl | toHost }}"
  - name: CONFIG_JSON
    value: '{{ dict "uid" .uid "host" .host | toJson }}'
```

### Namespace/Name Templates

**v0.x:**
- Names were constructed programmatically in controller code
- Pattern: `{tenantName}-{tenantId}-{resourceType}`

**v1.0:**
- Explicit templates in resource definitions
- Validation ensures uniqueness

```yaml
namespaceTemplate: "tenant-{{ .uid }}"
nameTemplate: "{{ .uid }}-{{ default \"default\" .planId }}-api"
```

---

## Breaking Changes

### 1. API Group Change

- **Old**: `tenant.ecubelabs.com/v1alpha1`
- **New**: `tenants.ecube.dev/v1`

**Impact**: All client code, RBAC, admission webhooks must update.

### 2. Spec Structure

- **TenantPool** split into **TenantRegistry** + **TenantTemplate**
- **ProvisioningRequest** removed (functionality integrated)

### 3. Template Syntax

- String substitution (`${}`) → Go templates (`{{ }}`)
- All strings must be valid template syntax

### 4. Password Security

- Plain-text passwords no longer supported
- Must use `passwordRef` → Secret

### 5. Resource Ownership

- Old: `ProvisioningRequest` → resources
- New: `Tenant` → resources
- Requires ownership transfer script

### 6. Reconciliation Model

- Old: Job-based provisioning (one-time)
- New: Continuous reconciliation with SSA

---

## New Features

### 1. Dependency Management

Define resource creation order:

```yaml
deployments:
  - id: db
    # ...
  - id: api
    dependIds: [db]  # Wait for db to be ready
    # ...
```

### 2. Creation Policies

```yaml
jobs:
  - id: init-db
    creationPolicy: Once  # Only run on first creation
    # ...
```

### 3. Deletion Policies

```yaml
persistentVolumeClaims:
  - id: data
    deletionPolicy: Retain  # Keep PVC when Tenant is deleted
    # ...
```

### 4. Conflict Handling

```yaml
services:
  - id: external-lb
    conflictPolicy: Force  # Takeover existing Service
    # ...
```

### 5. Readiness Checks

```yaml
statefulSets:
  - id: db
    waitForReady: true
    timeoutSeconds: 600
    # ...
```

### 6. Custom Template Functions

- `toHost(url)`: Extract host from URL
- `trunc63(str)`: Truncate to 63 chars (K8s label limit)
- `sha1sum(str)`: Generate SHA1 hash
- All Sprig functions: `default`, `b64enc`, `toJson`, etc.

---

## Rollback Strategy

### Immediate Rollback (During Migration)

If issues arise during migration:

```bash
# 1. Scale down v1.0 operator
kubectl scale deploy tenant-operator-controller-manager -n tenant-operator-system --replicas=0

# 2. Restore v0.x operator
kubectl scale deploy tenant-operator-controller-manager -n tenant-operator-system --replicas=1

# 3. Restore old CRs (if deleted)
kubectl apply -f tenantpool-backup.yaml
kubectl apply -f tenant-backup.yaml
```

### Long-term Rollback (After Migration)

If rollback is needed after full migration:

```bash
# 1. Export current resource state
kubectl get deploy,svc,ing -A -o yaml > current-resources.yaml

# 2. Remove v1.0 operator
kubectl delete -f config/manager/manager.yaml

# 3. Remove v1.0 CRDs
make uninstall

# 4. Reinstall v0.x
kubectl apply -f https://github.com/Ecube-Labs/tenant-operator/releases/latest/install.yaml

# 5. Restore v0.x CRs
kubectl apply -f tenantpool-backup.yaml

# 6. Manually reconcile resource ownership if needed
```

---

## Testing Checklist

Before production migration:

- [ ] Test TenantRegistry connection to MySQL
- [ ] Verify template rendering with sample data
- [ ] Test dependency ordering (create resources in sequence)
- [ ] Verify SSA behavior (no resource thrashing)
- [ ] Test deletion policies (Delete vs Retain)
- [ ] Test conflict policies (Stuck vs Force)
- [ ] Validate readiness checks for all resource types
- [ ] Load test with 100+ tenants
- [ ] Test database failover/reconnection
- [ ] Verify metrics and observability
- [ ] Test upgrades (change template, observe rolling update)
- [ ] Test rollback procedure

---

## Support and Troubleshooting

### Common Issues

**1. Template Rendering Errors**

```bash
# Check Tenant status for template errors
kubectl describe tenant <name>

# Look for condition: type=TemplateFailed
```

**2. Resource Ownership Conflicts**

```bash
# Check ownerReferences
kubectl get deploy <name> -o jsonpath='{.metadata.ownerReferences}'

# Expected: Tenant CR (tenants.ecube.dev/v1)
```

**3. Database Connection Issues**

```bash
# Check TenantRegistry conditions
kubectl get tenantregistry <name> -o yaml

# Verify Secret exists
kubectl get secret <secret-name> -o yaml
```

### Debug Commands

```bash
# Enable verbose logging
kubectl set env deploy/tenant-operator-controller-manager -n tenant-operator-system ZAP_LOG_LEVEL=debug

# Check reconciliation metrics
kubectl port-forward -n tenant-operator-system svc/tenant-operator-controller-manager-metrics-service 8080:8080
curl http://localhost:8080/metrics | grep tenant_
```

---

## Next Steps

After successful migration:

1. **Update CI/CD pipelines** to use new CRD schemas
2. **Migrate monitoring dashboards** to new metrics
3. **Update documentation** and runbooks
4. **Train team** on new template system and policies
5. **Plan for validation webhooks** (optional, future enhancement)
6. **Consider GitOps integration** (Flux/ArgoCD with new CRDs)

---

## Resources

- **New Repository**: https://github.com/kubernetes-tenants/tenant-operator
- **Design Document**: See root README or `DESIGN.md`
- **API Reference**: `make manifests` → `config/crd/bases/`
- **Sample Templates**: `config/samples/`

---

## Contact

For migration assistance:
- Open an issue: https://github.com/kubernetes-tenants/tenant-operator/issues
- Discussion forum: (TBD)
