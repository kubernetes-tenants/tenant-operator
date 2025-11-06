# Runbook: Tenant Resources Failed

## Alert Details

**Alert Name:** `TenantResourcesFailed`
**Severity:** Critical
**Threshold:** tenant_resources_failed > 0 for 5+ minutes

## Description

This alert fires when one or more resources belonging to a Tenant CR have failed to be created or updated successfully. This indicates a critical provisioning failure that affects the tenant's functionality.

## Symptoms

- Tenant status shows `failedResources > 0`
- Some tenant resources are missing or not ready
- Application components may be unavailable
- Events show resource creation or update failures

## Possible Causes

1. **API Server Rejection**
   - Invalid resource specifications
   - Schema validation failures
   - Admission webhook rejections

2. **Resource Quota Exceeded**
   - Namespace resource quotas reached
   - Cluster capacity limits hit
   - LimitRanges preventing resource creation

3. **Insufficient Permissions**
   - Operator lacks RBAC permissions
   - ServiceAccount permissions issues
   - Cross-namespace access denied

4. **Resource Dependencies Not Met**
   - Required ConfigMaps or Secrets missing
   - Dependent resources not ready yet
   - Namespace not created

5. **Readiness Timeout**
   - Resources created but not becoming ready
   - Exceeded `timeoutSeconds` (default: 300s)
   - Pods crashing or ImagePullBackOff

6. **Network or Storage Issues**
   - PVC creation failures
   - LoadBalancer provisioning failures
   - Ingress controller issues

## Diagnosis

### 1. Identify Failed Resources

```bash
# Check tenant status
kubectl get tenant <tenant-name> -n <namespace> -o jsonpath='{.status.failedResources}'

# Get detailed status
kubectl get tenant <tenant-name> -n <namespace> -o yaml | grep -A 20 "status:"
```

### 2. Review Resource Events

```bash
# Check tenant events
kubectl describe tenant <tenant-name> -n <namespace>

# Check events for specific resource types
kubectl get events -n <namespace> --sort-by='.lastTimestamp' | grep -i error
```

### 3. Check Operator Logs

```bash
# Filter logs for this tenant
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=200 | grep <tenant-name>

# Look for apply failures
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=200 | grep -i "apply failed\|error applying"
```

### 4. Inspect Individual Resources

```bash
# List resources owned by tenant
kubectl get all -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name>

# Check specific resource type
kubectl describe deployment <deployment-name> -n <namespace>
kubectl describe service <service-name> -n <namespace>
```

### 5. Check Resource Quotas

```bash
# View namespace resource quotas
kubectl describe resourcequota -n <namespace>

# Check limit ranges
kubectl describe limitrange -n <namespace>
```

### 6. Verify RBAC Permissions

```bash
# Check if operator can create resources
kubectl auth can-i create deployments --as=system:serviceaccount:tenant-operator-system:tenant-operator-controller-manager -n <namespace>

# Check for specific resource types
for kind in deployment service configmap secret ingress; do
  echo -n "$kind: "
  kubectl auth can-i create $kind --as=system:serviceaccount:tenant-operator-system:tenant-operator-controller-manager -n <namespace>
done
```

## Resolution

### For API Server Rejections

1. **Review resource specifications:**
   ```bash
   # Get resource spec from tenant
   kubectl get tenant <tenant-name> -n <namespace> -o yaml

   # Check for invalid fields or values
   ```

2. **Fix template if needed:**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>
   # Correct invalid field values or schema violations
   ```

### For Resource Quota Issues

1. **Check current usage:**
   ```bash
   kubectl describe resourcequota -n <namespace>
   ```

2. **Increase quota if appropriate:**
   ```bash
   kubectl edit resourcequota <quota-name> -n <namespace>
   # Increase limits for CPU, memory, or object counts
   ```

3. **Or reduce tenant resource requests:**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>
   # Adjust resource requests/limits in Deployment specs
   ```

### For Permission Issues

1. **Grant missing permissions:**
   ```bash
   kubectl edit clusterrole tenant-operator-manager-role
   # Add required resource types and verbs
   ```

2. **For cross-namespace resources:**
   ```yaml
   # Ensure operator has cluster-wide permissions for cross-namespace resources
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRole
   metadata:
     name: tenant-operator-manager-role
   rules:
   - apiGroups: [""]
     resources: ["namespaces", "services", "configmaps"]
     verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
   ```

### For Readiness Timeouts

1. **Check pod status:**
   ```bash
   kubectl get pods -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name>
   kubectl describe pod <pod-name> -n <namespace>
   ```

2. **Common pod issues:**
   - **ImagePullBackOff:** Fix image name or add imagePullSecrets
   - **CrashLoopBackOff:** Check pod logs for application errors
   - **Pending:** Check node resources and scheduling constraints

3. **Increase timeout if needed:**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>
   # Increase timeoutSeconds for slow-starting resources
   ```

4. **Disable readiness check for non-critical resources:**
   ```yaml
   spec:
     resources:
       - id: optional-job
         waitForReady: false  # Don't block on this resource
   ```

### For Network/Storage Issues

1. **Check PVC status:**
   ```bash
   kubectl get pvc -n <namespace>
   kubectl describe pvc <pvc-name> -n <namespace>
   ```

2. **Check storage class:**
   ```bash
   kubectl get storageclass
   # Ensure referenced storage class exists and is default
   ```

3. **Check LoadBalancer services:**
   ```bash
   kubectl get svc -n <namespace> -o wide
   # Verify LoadBalancer has EXTERNAL-IP assigned
   ```

4. **Check Ingress:**
   ```bash
   kubectl get ingress -n <namespace>
   kubectl describe ingress <ingress-name> -n <namespace>
   ```

## Quick Fixes

### Force Reconciliation

```bash
# Add annotation to trigger immediate reconciliation
kubectl annotate tenant <tenant-name> -n <namespace> \
  tenant.operator.kubernetes-tenants.org/reconcile="$(date +%s)" --overwrite
```

### Temporarily Skip Failed Resource

```bash
# Edit tenant template to exclude problematic resource temporarily
kubectl edit tenanttemplate <template-name> -n <namespace>
# Comment out or remove the failing resource
# Once root cause is fixed, add it back
```

### Reset Resource State

```bash
# For resources with ConflictPolicy=Stuck, manually delete conflicting resource
kubectl delete <resource-kind> <resource-name> -n <namespace>
# Operator will recreate it on next reconciliation
```

## Prevention

1. **Validate templates thoroughly:**
   - Test in development environment first
   - Use dry-run mode if available
   - Validate against API server schema

2. **Set appropriate resource limits:**
   - Define reasonable CPU/memory requests
   - Ensure quotas accommodate expected tenant count
   - Use LimitRanges for default values

3. **Configure adequate timeouts:**
   - Set `timeoutSeconds` based on actual resource readiness time
   - Use `waitForReady: false` for optional resources
   - Consider startup time for complex applications

4. **Monitor resource capacity:**
   - Track cluster resource utilization
   - Alert on quota usage approaching limits
   - Plan capacity for tenant growth

5. **Use dependency ordering:**
   - Define `dependIds` to ensure proper sequencing
   - Create namespaces and secrets first
   - Deploy applications after infrastructure resources

## Metrics to Monitor

```promql
# Failed resources per tenant
tenant_resources_failed{tenant="<tenant-name>"}

# Resource vs desired count
tenant_resources_ready / tenant_resources_desired < 1

# Apply failure rate by resource kind
rate(apply_attempts_total{result="failure"}[5m])

# Reconciliation errors
rate(tenant_reconcile_duration_seconds_count{result="error"}[5m])
```

## Related Alerts

- `TenantDegraded` - Likely fires concurrently
- `TenantNotReady` - Tenant won't be ready with failed resources
- `TenantResourcesMismatch` - May indicate partial failure
- `HighApplyFailureRate` - May indicate systemic issue

## Escalation

If issue persists after troubleshooting:

1. Collect diagnostic information:
   ```bash
   # Tenant details
   kubectl get tenant <tenant-name> -n <namespace> -o yaml > tenant.yaml

   # All related resources
   kubectl get all,ingress,configmap,secret -n <namespace> \
     -l kubernetes-tenants.org/tenant=<tenant-name> -o yaml > resources.yaml

   # Events
   kubectl get events -n <namespace> --sort-by='.lastTimestamp' > events.txt

   # Operator logs
   kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 > operator.log
   ```

2. Review operator GitHub issues for similar problems
3. Contact platform engineering team with diagnostic bundle
