# Runbook: Tenant Resource Conflicts

## Alert Details

**Alert Name:** `TenantResourcesConflicted`
**Severity:** Warning
**Threshold:** tenant_resources_conflicted > 0 for 10+ minutes

## Description

This alert fires when a tenant has resources in conflict state. Conflicts occur when the operator attempts to create or update a resource that already exists with a different owner or manager. This usually happens with `ConflictPolicy: Stuck` (default), which prevents overwriting existing resources.

## Symptoms

- Tenant has resources in conflicted state
- Some resources are not being created or updated
- Tenant may be in degraded state
- Events show ownership or SSA conflicts

## Possible Causes

1. **Naming Collisions**
   - Multiple tenants using same resource names
   - Insufficient uniqueness in `nameTemplate`
   - Shared resources across namespaces

2. **Existing Resources**
   - Resources pre-exist from previous deployments
   - Manual resource creation conflicts with operator
   - Resources from deleted tenants not cleaned up

3. **Multiple Owners**
   - Other controllers managing same resources
   - Resources owned by different operators
   - Helm/Kustomize managed resources

4. **SSA Field Manager Conflicts**
   - Different field managers trying to control same fields
   - Changes to operator field manager name
   - Manual kubectl apply commands

5. **Cross-Registry Conflicts**
   - Multiple registries creating resources in same namespace
   - Different templates with overlapping resource names

## Diagnosis

### 1. Check Conflicted Resources Count

```bash
# Check conflict count
kubectl get tenant <tenant-name> -n <namespace> \
  -o jsonpath='{.status.resourcesConflicted}'

# Check if tenant is degraded due to conflicts
kubectl get tenant <tenant-name> -n <namespace> \
  -o jsonpath='{.status.conditions[?(@.type=="Degraded")].reason}'
```

### 2. Identify Conflicting Resources

```bash
# Get tenant events showing conflicts
kubectl describe tenant <tenant-name> -n <namespace> | grep -i conflict

# Get all events related to conflicts
kubectl get events -n <namespace> --field-selector reason=Conflict

# Check operator logs for conflict details
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=200 \
  | grep -i "conflict\|already exists"
```

### 3. Find Existing Resource Ownership

```bash
# List what the tenant expects to own
kubectl get tenant <tenant-name> -n <namespace> -o yaml | grep "name:"

# Check actual resources in namespace
kubectl get all -n <namespace> | grep <pattern>

# Check resource ownership
kubectl get <resource-kind> <resource-name> -n <namespace> \
  -o jsonpath='{.metadata.ownerReferences[*].name}'

# Check SSA field managers
kubectl get <resource-kind> <resource-name> -n <namespace> \
  -o jsonpath='{.metadata.managedFields[*].manager}' | tr ' ' '\n'
```

### 4. Check Naming Pattern

```bash
# Get tenant template
TEMPLATE=$(kubectl get tenant <tenant-name> -n <namespace> \
  -o jsonpath='{.metadata.labels.kubernetes-tenants\.org/template}')

# Check nameTemplate patterns
kubectl get tenanttemplate $TEMPLATE -n <namespace> \
  -o yaml | grep -A 2 nameTemplate

# Check for overlapping patterns with other templates
kubectl get tenanttemplates -n <namespace> -o yaml | grep -A 2 nameTemplate
```

### 5. Check Conflict Policy

```bash
# Check tenant's conflict policy settings
kubectl get tenant <tenant-name> -n <namespace> \
  -o jsonpath='{.spec.resources[*].conflictPolicy}' | tr ' ' '\n' | sort | uniq -c
```

## Resolution

### For Naming Collisions

1. **Make names more unique:**
   ```bash
   # Edit template to add more uniqueness
   kubectl edit tenanttemplate <template-name> -n <namespace>

   # Examples of better naming patterns:
   # Before: nameTemplate: "myapp"
   # After:  nameTemplate: "{{ .uid }}-myapp"
   # Or:     nameTemplate: "{{ printf \"%s-%s\" .uid .templateRef | trunc63 }}"
   ```

2. **Verify uniqueness:**
   ```bash
   # Check if new names would be unique
   kubectl get all -n <namespace> | grep "<new-pattern>"
   ```

3. **Apply changes:**
   - Template changes will trigger tenant reconciliation
   - Existing resources may need manual cleanup

### For Pre-existing Resources

#### Option A: Remove Conflicting Resources (Safest)

```bash
# Backup existing resource
kubectl get <resource-kind> <resource-name> -n <namespace> -o yaml > backup.yaml

# Delete the conflicting resource
kubectl delete <resource-kind> <resource-name> -n <namespace>

# Operator will recreate it with correct ownership
# Monitor tenant reconciliation
kubectl get tenant <tenant-name> -n <namespace> -w
```

#### Option B: Change Conflict Policy to Force (Use with Caution)

```bash
# Edit tenant template to use Force policy
kubectl edit tenanttemplate <template-name> -n <namespace>

# Add or change conflictPolicy for affected resources:
# spec:
#   deployments:
#     - id: app-deployment
#       conflictPolicy: Force  # Changes from Stuck to Force
#       # ... rest of spec
```

**Warning:** `ConflictPolicy: Force` will take ownership of existing resources, potentially disrupting other systems managing them.

#### Option C: Remove Conflicting Resource's Owner Reference

```bash
# If resource is owned by something else, remove owner reference
kubectl patch <resource-kind> <resource-name> -n <namespace> \
  --type=json \
  -p='[{"op": "remove", "path": "/metadata/ownerReferences"}]'

# Then delete and let operator recreate
kubectl delete <resource-kind> <resource-name> -n <namespace>
```

### For SSA Field Manager Conflicts

1. **Check field managers:**
   ```bash
   kubectl get <resource-kind> <resource-name> -n <namespace> \
     -o json | jq '.metadata.managedFields[] | {manager, operation}'
   ```

2. **Force apply if safe:**
   ```bash
   # Change conflict policy to Force for this resource
   kubectl edit tenanttemplate <template-name> -n <namespace>
   # Set conflictPolicy: Force
   ```

3. **Or manually clear field management:**
   ```bash
   # Remove conflicting field manager (risky!)
   kubectl patch <resource-kind> <resource-name> -n <namespace> \
     --type=json \
     -p='[{"op": "remove", "path": "/metadata/managedFields/0"}]'
   ```

### For Cross-Registry Conflicts

1. **Identify conflicting registries:**
   ```bash
   # Find all tenants with same resource names
   kubectl get tenants -n <namespace> -o yaml | grep -B 5 <resource-name>
   ```

2. **Resolve at template level:**
   ```bash
   # Update templates to include registry or template name in resource names
   kubectl edit tenanttemplate <template-name> -n <namespace>

   # Use unique prefix:
   # nameTemplate: "{{ .registryId }}-{{ .uid }}-myapp"
   # Or: nameTemplate: "{{ .templateRef }}-{{ .uid }}-myapp"
   ```

## Bulk Conflict Resolution

If many tenants have conflicts:

```bash
# List all conflicted tenants
kubectl get tenants -n <namespace> -o json | \
  jq -r '.items[] | select(.status.resourcesConflicted > 0) | .metadata.name'

# For each, get conflict details
for tenant in $(kubectl get tenants -n <namespace> -o json | \
  jq -r '.items[] | select(.status.resourcesConflicted > 0) | .metadata.name'); do

  echo "=== $tenant ==="
  kubectl describe tenant $tenant -n <namespace> | grep -i conflict
  echo ""
done

# If same root cause, fix template once
kubectl edit tenanttemplate <template-name> -n <namespace>

# Force reconciliation of all affected tenants
for tenant in $(kubectl get tenants -n <namespace> \
  -l kubernetes-tenants.org/template=<template-name> -o name); do
  kubectl annotate $tenant -n <namespace> \
    tenant.operator.kubernetes-tenants.org/reconcile="$(date +%s)" --overwrite
done
```

## Prevention

1. **Use unique naming patterns:**
   ```yaml
   # Always include tenant UID in names
   nameTemplate: "{{ .uid }}-{{ .resourceType }}"

   # For multi-template setups, include template name
   nameTemplate: "{{ .templateRef }}-{{ .uid }}-{{ .resourceType }}"

   # Ensure K8s name length compliance
   nameTemplate: "{{ printf \"%s-%s\" .uid .resourceType | trunc63 }}"
   ```

2. **Namespace isolation:**
   - Use dedicated namespaces per registry
   - Or use per-tenant namespaces
   - Avoid sharing namespaces across registries

3. **Set appropriate conflict policies:**
   ```yaml
   # Default (safe): Stuck - prevents accidental overwrites
   conflictPolicy: Stuck

   # Use Force only when necessary and safe:
   conflictPolicy: Force  # Use for: ConfigMaps, non-critical resources
   ```

4. **Clean up before redeployment:**
   - Delete old tenant resources before applying new ones
   - Use proper deletion policies
   - Ensure finalizers clean up resources

5. **Test templates before production:**
   - Deploy to test namespace first
   - Verify name uniqueness
   - Check for naming collisions

## Monitoring

```promql
# Current conflicted resources
tenant_resources_conflicted > 0

# Conflict rate over time
rate(tenant_conflicts_total[5m])

# Conflicts by resource kind
sum(tenant_conflicts_total) by (resource_kind)

# Conflicts by conflict policy
sum(tenant_conflicts_total) by (conflict_policy)

# Tenants with conflicts
count(tenant_resources_conflicted > 0)
```

## Related Alerts

- `TenantHighConflictRate` - High rate of conflicts detected
- `TenantDegraded` - May fire if conflicts cause degradation
- `TenantResourcesFailed` - May fire alongside conflicts
- `TenantNewConflictsDetected` - Info alert for new conflicts

## Best Practices

1. **Naming conventions:**
   - Always prefix with tenant UID
   - Include template or registry name for multi-template scenarios
   - Use consistent patterns across templates
   - Respect 63-character limit with `trunc63`

2. **Conflict policy strategy:**
   - Start with `Stuck` (default) in production
   - Use `Force` only for:
     - ConfigMaps that can be safely overwritten
     - Resources you fully control
     - Non-production environments
   - Document why `Force` is used

3. **Resource ownership:**
   - Avoid manual resource creation in operator-managed namespaces
   - Use separate namespaces for manual and operator-managed resources
   - Document resource ownership clearly

4. **Testing:**
   - Always test template changes in non-production first
   - Verify name uniqueness before deploying
   - Check for conflicts with existing resources

## Escalation

If conflicts persist after following this runbook:

1. Collect diagnostics:
   ```bash
   # Tenant details
   kubectl get tenant <tenant-name> -n <namespace> -o yaml > tenant.yaml

   # Conflicting resources
   kubectl get all,ingress,configmap,secret -n <namespace> -o yaml > resources.yaml

   # Template
   kubectl get tenanttemplate <template-name> -n <namespace> -o yaml > template.yaml

   # Events
   kubectl get events -n <namespace> --sort-by='.lastTimestamp' > events.txt

   # Operator logs
   kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 \
     | grep -i conflict > conflicts.log
   ```

2. Review with platform team:
   - Check naming patterns across all templates
   - Verify namespace allocation strategy
   - Review conflict policy choices

3. If unable to resolve:
   - Open GitHub issue with diagnostics
   - Include template definitions and conflict patterns
   - Describe expected vs actual behavior
