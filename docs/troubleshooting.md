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

**Cause:** Webhook TLS certificates not found. cert-manager is **REQUIRED** for all installations.

::: danger cert-manager Required
cert-manager v1.13.0+ is **REQUIRED** for ALL installations including local development. Webhooks provide validation and defaulting at admission time.
:::

**Diagnosis:**

```bash
# Check if cert-manager is installed
kubectl get pods -n cert-manager

# Check if Certificate resource exists
kubectl get certificate -n tenant-operator-system

# Check Certificate details
kubectl describe certificate -n tenant-operator-system

# Check if secret was created
kubectl get secret -n tenant-operator-system | grep webhook-server-cert
```

**Solutions:**

**A. Install cert-manager** (if not installed):
```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Wait for cert-manager to be ready
kubectl wait --for=condition=Available --timeout=300s -n cert-manager \
  deployment/cert-manager \
  deployment/cert-manager-webhook \
  deployment/cert-manager-cainjector

# Verify cert-manager is running
kubectl get pods -n cert-manager
```

**B. Restart operator** (after cert-manager is ready):
```bash
kubectl rollout restart -n tenant-operator-system deployment/tenant-operator-controller-manager

# Watch rollout status
kubectl rollout status -n tenant-operator-system deployment/tenant-operator-controller-manager
```

**C. Check Certificate issuance**:
```bash
# Check if Certificate is Ready
kubectl get certificate -n tenant-operator-system

# If not ready, check cert-manager logs
kubectl logs -n cert-manager -l app=cert-manager

# Check if Issuer exists
kubectl get issuer -n tenant-operator-system
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
- --tenant-concurrency=20      # Increase Tenant reconciliation concurrency
- --template-concurrency=10    # Increase Template reconciliation concurrency
- --registry-concurrency=5     # Increase Registry reconciliation concurrency
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
- --tenant-concurrency=5       # Reduce Tenant reconciliation concurrency
- --template-concurrency=3     # Reduce Template reconciliation concurrency
- --registry-concurrency=1     # Reduce Registry reconciliation concurrency
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

### 11. Orphaned Resources Not Cleaning Up

**Symptoms:**
- Resources removed from TenantTemplate still exist in cluster
- `appliedResources` status not updating
- Unexpected resources with tenant labels/ownerReferences

**Diagnosis:**

```bash
# Check current applied resources
kubectl get tenant <name> -o jsonpath='{.status.appliedResources}'

# Should show: ["Deployment/default/app@deploy-1", "Service/default/app@svc-1"]

# List resources with tenant labels
kubectl get all -l kubernetes-tenants.org/tenant=<tenant-name>

# Find orphaned resources (retained with DeletionPolicy=Retain)
kubectl get all -A -l kubernetes-tenants.org/orphaned=true

# Find orphaned resources from this tenant
kubectl get all -A -l kubernetes-tenants.org/orphaned=true,kubernetes-tenants.org/tenant=<tenant-name>

# Check resource DeletionPolicy
kubectl get tenanttemplate <name> -o yaml | grep -A2 deletionPolicy
```

**Common Causes:**

1. **DeletionPolicy=Retain**: Resource was intentionally retained and marked with orphan labels
2. **Status not syncing**: AppliedResources field not updated
3. **Manual resource modification**: OwnerReference or labels removed manually
4. **Operator version**: Upgrade from version without orphan cleanup

::: tip Expected Behavior
Resources with `DeletionPolicy=Retain` are **intentionally kept** in the cluster and marked with orphan labels for easy identification. This is not a bug - it's the designed behavior!
:::

**Solutions:**

**A. Verify DeletionPolicy:**
```yaml
# Check template definition
deployments:
  - id: old-deployment
    deletionPolicy: Delete  # Should be Delete, not Retain
```

**B. Force reconciliation:**
```bash
# Trigger reconciliation by updating an annotation
kubectl annotate tenant <name> force-sync="$(date +%s)" --overwrite

# Watch logs
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager -f
```

**C. Manual cleanup (if needed):**
```bash
# Delete orphaned resource manually
kubectl delete deployment <orphaned-resource>

# Or remove owner reference if you want to keep it
kubectl patch deployment <name> --type=json -p='[{"op": "remove", "path": "/metadata/ownerReferences"}]'
```

**D. Check status update:**
```bash
# Verify appliedResources is being updated
kubectl get tenant <name> -o jsonpath='{.status.appliedResources}' | jq

# Should reflect current template resources only
```

**Prevention:**

1. Use `deletionPolicy: Delete` for resources that should be cleaned up
2. Monitor `appliedResources` status field regularly
3. Test template changes in non-production first
4. Review orphan cleanup behavior in [Policies Guide](policies.md#orphan-resource-cleanup)

### 12. Tenant Showing Degraded with ResourcesNotReady

::: tip New in v1.1.4
The `ResourcesNotReady` degraded condition provides granular visibility into resources that haven't reached ready state yet.
:::

**Symptoms:**
- Tenant condition `Degraded=True` with reason `ResourcesNotReady`
- Not all resources showing as ready even though they exist
- `readyResources < desiredResources` in tenant status
- Ready condition shows `status=False` with reason `NotAllResourcesReady`

**Diagnosis:**

```bash
# Check tenant status
kubectl get tenant <name> -o jsonpath='{.status}' | jq

# Should show:
# "conditions": [
#   {"type": "Ready", "status": "False", "reason": "NotAllResourcesReady"},
#   {"type": "Degraded", "status": "True", "reason": "ResourcesNotReady"}
# ]

# Check resource readiness
kubectl get tenant <name> -o jsonpath='{.status.readyResources} / {.status.desiredResources}'

# Identify which resources are not ready
kubectl describe tenant <name>

# Check recent events
kubectl get events --field-selector involvedObject.name=<tenant-name> --sort-by='.lastTimestamp'
```

**Common Causes:**

1. **Resources still starting up**: Normal during initial provisioning
2. **Resource readiness checks failing**: Container failing health checks
3. **Dependency not satisfied**: Waiting for dependent resources
4. **Timeout exceeded**: Resource taking longer than `timeoutSeconds`
5. **Image pull errors**: Container images not available
6. **Insufficient resources**: Not enough CPU/memory in cluster

**Solutions:**

**A. Check resource status:**
```bash
# For Deployments
kubectl get deployment <name> -o jsonpath='{.status}' | jq

# Check if replicas match
kubectl get deployment <name> -o jsonpath='{.spec.replicas} desired, {.status.availableReplicas} available'

# For StatefulSets
kubectl get statefulset <name> -o jsonpath='{.status.readyReplicas}/{.spec.replicas} ready'

# For Jobs
kubectl get job <name> -o jsonpath='{.status.succeeded}'

# For Services (should be immediate)
kubectl get service <name>

# Check pod status
kubectl get pods -l app=<name>
kubectl describe pod <pod-name>
```

**B. Check resource logs:**
```bash
# Deployment pods
kubectl logs deployment/<name> --tail=50

# Job logs
kubectl logs job/<name>

# Check for errors
kubectl logs deployment/<name> --previous  # If pod crashed
```

**C. Check readiness probes:**
```bash
# See if readiness probes are configured correctly
kubectl get deployment <name> -o jsonpath='{.spec.template.spec.containers[*].readinessProbe}' | jq
```

**D. Adjust timeouts if needed:**
```yaml
# In TenantTemplate
deployments:
  - id: app
    timeoutSeconds: 600  # Increase from default 300s
    waitForReady: true   # Ensure readiness checks are enabled
    spec:
      template:
        spec:
          containers:
          - name: app
            readinessProbe:
              httpGet:
                path: /health
                port: 8080
              initialDelaySeconds: 10  # Allow time to start
              periodSeconds: 5
```

**E. Monitor reconciliation:**
```bash
# Watch tenant status updates (30-second interval in v1.1.4)
watch -n 5 'kubectl get tenant <name> -o jsonpath="{.status.readyResources}/{.status.desiredResources} ready"'

# Watch pod status
watch kubectl get pods -l tenant=<name>
```

**Expected Behavior:**

::: tip v1.1.4+ Fast Status Updates
- **Status updates every 30 seconds** (down from 5 minutes in earlier versions)
- **Event-driven**: Immediate reconciliation when child resources change
- Resources typically become ready within 1-2 minutes (depending on resource type)
- Degraded condition clears automatically when all resources reach ready state
:::

**When to Investigate Further:**

- Resources stuck "not ready" for > 5 minutes
- `readyResources` count not increasing over time
- Events show repeated failures or errors
- Logs indicate application-level issues

**Prevention:**

1. Set realistic `timeoutSeconds` values for resource types (Deployments: 300s, Jobs: 600s)
2. Ensure resource specifications have correct readiness probes
3. Test templates in non-production environments first
4. Monitor `tenant_resources_ready` and `tenant_degraded_status` metrics
5. Use `kubectl wait` for pre-flight checks:
   ```bash
   kubectl wait --for=condition=Ready tenant/<name> --timeout=300s
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
