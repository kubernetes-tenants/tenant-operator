# Troubleshooting Guide

Common issues and solutions for Tenant Operator.

[[toc]]

## General Debugging

### Check Operator Status

```bash
# Check if operator is running
kubectl get pods -n tenant-operator-system

# View operator logs
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager -f

# Check operator events
kubectl get events -n tenant-operator-system --sort-by='.lastTimestamp'
```

### Check CRD Status

```bash
# List all Tenant CRs
kubectl get tenants --all-namespaces

# Describe a specific Tenant
kubectl describe tenant <tenant-name>

# Get Tenant status
kubectl get tenant <tenant-name> -o jsonpath='{.status}'
```

## Common Issues

### 1. Webhook TLS Certificate Errors

**Error:**
```
open /tmp/k8s-webhook-server/serving-certs/tls.crt: no such file or directory
```

**Cause:** Webhook TLS certificates not found

**Solutions:**

Install cert-manager:
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

### 2. Tenant Not Creating Resources

**Symptoms:**
- Tenant CR exists
- Status shows `desiredResources > 0`
- But `readyResources = 0`

**Diagnosis:**
```bash
# Check Tenant status
kubectl get tenant <name> -o yaml

# Check events
kubectl describe tenant <name>

# Check operator logs
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager | grep <tenant-name>
```

**Common Causes:**

**A. Template Rendering Error**
```bash
# Look for: "Failed to render resource"
kubectl describe tenant <name> | grep -A5 "TemplateRenderError"
```

Solution: Fix template syntax in TenantTemplate

**B. Missing Variable**
```bash
# Look for: "map has no entry for key"
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager | grep "missing"
```

Solution: Add missing variable to `extraValueMappings`

**C. Resource Conflict**
```bash
# Look for: "ResourceConflict"
kubectl describe tenant <name> | grep "ResourceConflict"
```

Solution: Delete conflicting resource or use `conflictPolicy: Force`

### 3. Database Connection Failures

**Error:**
```
Failed to query database: dial tcp: connect: connection refused
```

**Diagnosis:**
```bash
# Check secret exists
kubectl get secret <mysql-secret> -o yaml

# Check Registry status
kubectl get tenantregistry <name> -o yaml

# Test database connection from a pod
kubectl run -it --rm mysql-test --image=mysql:8 --restart=Never -- \
  mysql -h <host> -u <user> -p<password> -e "SELECT 1"
```

**Solutions:**

A. Verify credentials:
```bash
kubectl get secret <mysql-secret> -o jsonpath='{.data.password}' | base64 -d
```

B. Check network connectivity:
```bash
kubectl exec -n tenant-operator-system deployment/tenant-operator-controller-manager -- \
  nc -zv <mysql-host> 3306
```

C. Verify TenantRegistry configuration:
```yaml
spec:
  source:
    mysql:
      host: mysql.default.svc.cluster.local  # Correct FQDN
      port: 3306
      database: tenants
```

### 4. Tenant Status Not Updating

**Symptoms:**
- Resources are ready in cluster
- Tenant status shows `readyResources = 0`

**Causes:**
- Reconciliation not triggered
- Readiness check failing

**Solutions:**

A. Force reconciliation:
```bash
# Add annotation to trigger reconciliation
kubectl annotate tenant <name> force-sync="$(date +%s)" --overwrite
```

B. Check readiness logic:
```bash
# For Deployments
kubectl get deployment <name> -o jsonpath='{.status}'

# Check if replicas match
kubectl get deployment <name> -o jsonpath='{.spec.replicas} {.status.availableReplicas}'
```

C. Wait longer (resources take time to become ready):
- Deployments: 30s - 2min
- Jobs: Variable
- Ingresses: 10s - 1min

### 5. Template Variables Not Substituting

::: v-pre
**Symptoms:**
- Template shows `{{ .uid }}` literally in resources
- Variables not replaced
:::

**Cause:** Templates not rendered correctly

**Diagnosis:**
```bash
# Check rendered Tenant spec
kubectl get tenant <name> -o jsonpath='{.spec.deployments[0].nameTemplate}'
```

**Solution:**
- Ensure Registry has correct `valueMappings`
- Check database column names match mappings
- Verify tenant row has non-empty values

### 6. Slow Tenant Provisioning

**Symptoms:**
- Tenants taking > 5 minutes to provision
- High operator CPU usage

**Diagnosis:**
```bash
# Check reconciliation times
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager | \
  grep "Reconciliation completed" | tail -20

# Check resource counts
kubectl get tenants -o json | jq '.items[] | {name: .metadata.name, desired: .status.desiredResources}'
```

**Solutions:**

A. Disable readiness waits:
```yaml
waitForReady: false
```

B. Increase concurrency:
```yaml
args:
- --tenant-concurrency=20
```

C. Optimize templates (see [Performance Guide](performance.md))

### 7. Memory/CPU Issues

**Symptoms:**
- Operator pod OOMKilled
- High CPU usage

**Diagnosis:**
```bash
# Check resource usage
kubectl top pod -n tenant-operator-system

# Check for memory leaks
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager --previous
```

**Solutions:**

A. Increase resource limits:
```yaml
resources:
  limits:
    cpu: 2000m
    memory: 2Gi
```

B. Reduce concurrency:
```yaml
args:
- --tenant-concurrency=5
```

C. Increase requeue interval:
```yaml
args:
- --requeue-interval=1m
```

### 8. Finalizer Stuck

**Symptoms:**
- Tenant CR stuck in `Terminating` state
- Can't delete Tenant

**Diagnosis:**
```bash
# Check finalizers
kubectl get tenant <name> -o jsonpath='{.metadata.finalizers}'

# Check deletion timestamp
kubectl get tenant <name> -o jsonpath='{.metadata.deletionTimestamp}'
```

**Solutions:**

A. Check operator logs for deletion errors:
```bash
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager | \
  grep "Failed to delete"
```

B. Force remove finalizer (last resort):
```bash
kubectl patch tenant <name> -p '{"metadata":{"finalizers":[]}}' --type=merge
```

**Warning:** This may leave orphaned resources!

### 9. Registry Not Syncing

**Symptoms:**
- Database has active rows
- No Tenant CRs created

**Diagnosis:**
```bash
# Check Registry status
kubectl get tenantregistry <name> -o yaml

# Check operator logs
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager | \
  grep "Registry"
```

**Common Causes:**

A. Incorrect `valueMappings`:
```yaml
# Must match database columns exactly
valueMappings:
  uid: tenant_id          # Column must exist
  hostOrUrl: tenant_url   # Column must exist
  activate: is_active     # Column must exist
```

B. No active rows:
```sql
-- Check for active tenants
SELECT COUNT(*) FROM tenants WHERE is_active = TRUE;
```

C. Database query error:
```bash
# Check logs for SQL errors
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager | \
  grep "Failed to query"
```

### 10. Multi-Template Issues

**Symptoms:**
- Expected 2× tenants, only seeing 1×
- Wrong desired count

**Diagnosis:**
```bash
# Check Registry status
kubectl get tenantregistry <name> -o jsonpath='{.status}'

# Should show:
# referencingTemplates: 2
# desired: <templates> × <rows>

# Check templates reference same registry
kubectl get tenanttemplates -o jsonpath='{.items[*].spec.registryId}'
```

**Solution:**
Ensure all templates correctly reference the registry:
```yaml
spec:
  registryId: my-registry  # Must match exactly
```

## Debugging Workflows

### Debug Template Rendering

1. Create test Tenant manually:
```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: Tenant
metadata:
  name: test-tenant
  annotations:
    kubernetes-tenants.org/uid: "test-123"
    kubernetes-tenants.org/host: "test.example.com"
spec:
  # ... copy from template
```

2. Check rendered resources:
```bash
kubectl get tenant test-tenant -o yaml
```

3. Check operator logs:
```bash
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager -f
```

### Debug Database Connection

1. Create test pod:
```bash
kubectl run -it --rm mysql-test --image=mysql:8 --restart=Never -- bash
```

2. Inside pod:
```bash
mysql -h <host> -u <user> -p<password> <database> -e "SELECT * FROM tenants LIMIT 5"
```

### Debug Reconciliation

1. Enable debug logging:
```yaml
# config/manager/manager.yaml
args:
- --zap-log-level=debug
```

2. Watch reconciliation:
```bash
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager -f | \
  grep "Reconciling"
```

## Getting Help

1. Check operator logs
2. Check Tenant events: `kubectl describe tenant <name>`
3. Check Registry status: `kubectl get tenantregistry <name> -o yaml`
4. Review [Performance Guide](performance.md)
5. Open issue: https://github.com/kubernetes-tenants/tenant-operator/issues

Include in bug reports:
- Operator version
- Kubernetes version
- Operator logs
- Tenant/Registry/Template YAML
- Steps to reproduce
