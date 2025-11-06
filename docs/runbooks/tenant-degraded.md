# Runbook: Tenant Degraded

## Alert Details

**Alert Name:** `TenantDegraded`
**Severity:** Critical
**Threshold:** Tenant degraded status > 0 for 5+ minutes

## Description

This alert fires when a Tenant CR enters a degraded state, indicating that the operator cannot successfully reconcile the tenant's resources. This is a critical condition that prevents the tenant from functioning properly.

## Symptoms

- Tenant's `Ready` condition is `False`
- Tenant status shows degraded state with a specific reason
- Resources may be partially applied or stuck in failed states
- Events show errors related to template rendering, conflicts, or dependencies

## Possible Causes

1. **Template Rendering Errors**
   - Missing or invalid template variables
   - Syntax errors in Go templates
   - Type conversion failures

2. **Dependency Cycle Detected**
   - Circular dependencies between resources (A depends on B, B depends on A)
   - Invalid `dependIds` configuration

3. **Resource Conflicts**
   - Resources exist with different owners (when using `ConflictPolicy: Stuck`)
   - Naming collisions with existing resources

4. **External Dependencies Unavailable**
   - Referenced ConfigMaps or Secrets don't exist
   - Required namespaces not created yet

5. **RBAC Permission Issues**
   - Operator lacks permissions to create certain resource types
   - Cross-namespace resource creation blocked

## Diagnosis

### 1. Check Tenant Status

```bash
kubectl get tenant <tenant-name> -n <namespace> -o yaml
```

Look for:
- `status.conditions` - Check `Degraded` condition and its `reason`
- `status.message` - Detailed error message
- `status.failedResources` - Count of failed resources

### 2. Review Tenant Events

```bash
kubectl describe tenant <tenant-name> -n <namespace>
```

Look for recent events indicating errors or conflicts.

### 3. Check Operator Logs

```bash
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=100 | grep <tenant-name>
```

Search for:
- Template rendering errors
- Dependency cycle detection
- Conflict detection messages
- Apply failures

### 4. Validate Template Variables

```bash
kubectl get tenant <tenant-name> -n <namespace> -o jsonpath='{.metadata.annotations}'
```

Ensure all required variables (`.uid`, `.hostOrUrl`, `.activate`) are present and valid.

### 5. Check Resource Dependencies

```bash
kubectl get tenant <tenant-name> -n <namespace> -o jsonpath='{.spec.resources[*].dependIds}'
```

Verify no circular dependencies exist.

## Resolution

### For Template Rendering Errors

1. **Fix missing variables:**
   ```bash
   # Check TenantRegistry extraValueMappings
   kubectl get tenantregistry <registry-name> -n <namespace> -o yaml

   # Ensure all required columns are mapped
   ```

2. **Update TenantTemplate:**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>
   # Fix template syntax or add default values using sprig functions
   # Example: {{ .variable | default "fallback-value" }}
   ```

### For Dependency Cycles

1. **Identify the cycle:**
   ```bash
   # Review dependIds in tenant spec
   kubectl get tenant <tenant-name> -n <namespace> -o yaml | grep -A 5 dependIds
   ```

2. **Fix TenantTemplate dependencies:**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>
   # Remove or reorder dependIds to break the cycle
   ```

### For Resource Conflicts

1. **Identify conflicting resources:**
   ```bash
   # Check events for conflict messages
   kubectl describe tenant <tenant-name> -n <namespace>
   ```

2. **Option A - Change naming template:**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>
   # Update nameTemplate to ensure uniqueness
   # Example: {{ .uid }}-{{ .templateRef }}-myapp
   ```

3. **Option B - Force takeover (use with caution):**
   ```yaml
   # Update resource ConflictPolicy to Force
   spec:
     resources:
       - id: conflicted-resource
         conflictPolicy: Force  # Changes from Stuck to Force
   ```

### For RBAC Issues

1. **Check operator permissions:**
   ```bash
   kubectl auth can-i create <resource-kind> --as=system:serviceaccount:tenant-operator-system:tenant-operator-controller-manager -n <target-namespace>
   ```

2. **Update RBAC if needed:**
   ```bash
   kubectl edit clusterrole tenant-operator-manager-role
   # Add missing resource permissions
   ```

## Prevention

1. **Validate templates before deployment:**
   - Test templates with sample data
   - Use default values for optional fields:
     ```
     {{ .variable | default "value" }}
     ```
   - Avoid circular dependencies

2. **Use unique naming patterns:**
   - Include tenant UID in all resource names
   - Use `trunc63` function for K8s name length limits
   - Example:
     ```
     {{ printf "%s-%s" .uid .resourceType | trunc63 }}
     ```

3. **Monitor template changes:**
   - Review TenantTemplate updates carefully
   - Test changes in non-production environments first
   - Use GitOps for version-controlled template management

4. **Set appropriate conflict policies:**
   - Use `ConflictPolicy: Stuck` (default) for production safety
   - Only use `Force` for non-critical resources or when takeover is intentional

## Metrics to Monitor

```promql
# Degraded tenants by reason
tenant_degraded_status{reason!=""}

# Tenant condition status
tenant_condition_status{type="Ready"}

# Failed resources per tenant
tenant_resources_failed > 0
```

## Related Alerts

- `TenantResourcesFailed` - May fire concurrently
- `TenantNotReady` - Will likely fire if degraded persists
- `TenantResourcesConflicted` - May indicate conflict-related degradation

## Escalation

If the issue persists after following this runbook:

1. Collect diagnostic bundle:
   ```bash
   kubectl get tenant <tenant-name> -n <namespace> -o yaml > tenant.yaml
   kubectl describe tenant <tenant-name> -n <namespace> > tenant-describe.txt
   kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 > operator-logs.txt
   ```

2. Check for known issues in operator repository
3. Contact platform team or open GitHub issue with diagnostic bundle
