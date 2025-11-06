# Runbook: High Conflict Rate

## Alert Details

**Alert Name:** `TenantHighConflictRate`
**Severity:** Warning
**Threshold:** rate(tenant_conflicts_total[5m]) > 0.1 for 10+ minutes

## Description

This alert fires when a tenant is experiencing a high rate of conflicts (>0.1 conflicts per second sustained for 10+ minutes). This indicates repeated attempts to create or update resources that conflict with existing ones, suggesting a persistent configuration or naming issue.

## Symptoms

- Tenant experiencing continuous conflict errors
- Reconciliation repeatedly attempting same operations
- Resource creation/updates failing repeatedly
- Events showing recurring conflict messages
- Tenant may be degraded or stuck

## Possible Causes

1. **Reconciliation Loop**
   - Operator repeatedly trying to apply same conflicting resource
   - Conflict not resolving automatically
   - ConflictPolicy set to Stuck (default) preventing resolution

2. **Naming Pattern Issues**
   - Insufficient uniqueness in name generation
   - Template variables not providing uniqueness
   - Multiple tenants generating same names

3. **External Changes**
   - External system continuously modifying resources
   - Another controller competing for same resources
   - Manual interventions fighting operator changes

4. **Rapid Tenant Churn**
   - Tenants being created and deleted rapidly
   - Resources not cleaned up before recreation
   - Timing issues with resource lifecycle

5. **Template Changes**
   - Recent template update causing naming conflicts
   - Changes to nameTemplate creating collisions
   - Multiple versions of template active

## Diagnosis

### 1. Check Conflict Rate and Pattern

```bash
# Check current conflict count for tenant
kubectl get tenant <tenant-name> -n <namespace> \
  -o jsonpath='{.status.resourcesConflicted}'

# Get conflict events
kubectl describe tenant <tenant-name> -n <namespace> | grep -i conflict

# Check operator logs for conflict pattern
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 \
  | grep -i "conflict.*<tenant-name>" | head -20
```

### 2. Identify Conflicting Resources

```bash
# Get detailed conflict information from logs
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=1000 \
  | grep -i "conflict" \
  | grep "<tenant-name>" \
  | awk '{print $NF}' | sort | uniq -c

# Check resource kinds involved
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 \
  | grep "<tenant-name>.*conflict" \
  | grep -oP 'kind=\K\w+' | sort | uniq -c
```

### 3. Check Conflict Policy

```bash
# Check tenant's conflict policies
kubectl get tenant <tenant-name> -n <namespace> -o yaml \
  | grep -A 2 conflictPolicy

# Check if using Stuck policy (default)
kubectl get tenant <tenant-name> -n <namespace> \
  -o jsonpath='{.spec.resources[*].conflictPolicy}' | tr ' ' '\n' | sort | uniq -c
```

### 4. Check Resource Ownership

```bash
# Find the conflicting resource name from logs
RESOURCE_NAME="<resource-name>"  # Get from logs
RESOURCE_KIND="<resource-kind>"  # Get from logs

# Check who owns the resource
kubectl get $RESOURCE_KIND $RESOURCE_NAME -n <namespace> \
  -o jsonpath='{.metadata.ownerReferences[*].name}'

# Check SSA field managers
kubectl get $RESOURCE_KIND $RESOURCE_NAME -n <namespace> \
  -o jsonpath='{.metadata.managedFields[*].manager}' | tr ' ' '\n'
```

### 5. Check for Reconciliation Loop

```bash
# Monitor reconciliation attempts
kubectl logs -n tenant-operator-system -l app=tenant-operator --follow \
  | grep "Reconciling Tenant.*<tenant-name>"

# Check reconciliation frequency (should not be constant)
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=100 \
  | grep "<tenant-name>" | grep -c "Reconciling"
```

### 6. Check Recent Template Changes

```bash
# Get template history
kubectl get tenanttemplate <template-name> -n <namespace> -o yaml

# Check for recent updates
kubectl get events -n <namespace> --sort-by='.lastTimestamp' \
  | grep TenantTemplate

# Compare with previous version if using GitOps
git log -p -- path/to/template.yaml
```

## Resolution

### For Persistent Naming Conflicts

1. **Identify the naming issue:**
   ```bash
   # Check nameTemplate pattern
   kubectl get tenanttemplate <template-name> -n <namespace> \
     -o jsonpath='{.spec.deployments[*].nameTemplate}' | tr ' ' '\n'

   # Check if tenant variables provide uniqueness
   kubectl get tenant <tenant-name> -n <namespace> \
     -o jsonpath='{.metadata.annotations}'
   ```

2. **Fix naming template:**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>

   # Ensure uniqueness:
   # nameTemplate: "{{ .uid }}-myapp"  # Add tenant UID
   # OR
   # nameTemplate: "{{ printf \"%s-%s\" .uid .templateRef | trunc63 }}"
   ```

3. **Clean up and retry:**
   ```bash
   # Delete conflicting resource
   kubectl delete $RESOURCE_KIND $RESOURCE_NAME -n <namespace>

   # Trigger reconciliation
   kubectl annotate tenant <tenant-name> -n <namespace> \
     tenant.operator.kubernetes-tenants.org/reconcile="$(date +%s)" --overwrite
   ```

### For ConflictPolicy Issues

If conflicts are expected and takeover is safe:

```bash
kubectl edit tenanttemplate <template-name> -n <namespace>

# Change policy for affected resources:
# spec:
#   deployments:
#     - id: app
#       conflictPolicy: Force  # Take ownership forcefully
#       # ... rest of spec
```

**Warning:** Only use `Force` if:
- You own the conflicting resources
- Takeover won't disrupt other systems
- You understand the implications

### For External Controller Conflicts

1. **Identify competing controller:**
   ```bash
   # Check field managers
   kubectl get $RESOURCE_KIND $RESOURCE_NAME -n <namespace> \
     -o json | jq '.metadata.managedFields[]'
   ```

2. **Options:**
   - **Option A:** Disable competing controller
   - **Option B:** Use different namespaces
   - **Option C:** Coordinate field management via SSA
   - **Option D:** Use Force policy (with caution)

3. **Implement chosen solution:**
   ```bash
   # Example: Add label to exclude from other controller
   kubectl label $RESOURCE_KIND $RESOURCE_NAME -n <namespace> \
     <competing-controller-label>=false
   ```

### For Reconciliation Loop

1. **Temporarily pause reconciliation:**
   ```bash
   # This capability may not exist - check operator docs
   # If available:
   kubectl annotate tenant <tenant-name> -n <namespace> \
     tenant.operator.kubernetes-tenants.org/pause=true
   ```

2. **Fix root cause** (naming, conflict policy, etc.)

3. **Resume reconciliation:**
   ```bash
   kubectl annotate tenant <tenant-name> -n <namespace> \
     tenant.operator.kubernetes-tenants.org/pause-
   ```

### Emergency Mitigation

If high conflict rate is causing operator performance issues:

1. **Identify affected tenants:**
   ```bash
   # Find tenants with high conflict rate
   kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=5000 \
     | grep -i conflict | awk '{print $(NF-1)}' | sort | uniq -c | sort -rn | head -10
   ```

2. **Temporarily remove problem tenant:**
   ```bash
   # Backup tenant
   kubectl get tenant <tenant-name> -n <namespace> -o yaml > tenant-backup.yaml

   # Delete tenant
   kubectl delete tenant <tenant-name> -n <namespace>

   # Fix template issue
   kubectl edit tenanttemplate <template-name> -n <namespace>

   # Recreate tenant (will happen automatically from registry sync)
   ```

## Prevention

1. **Design unique naming patterns:**
   ```yaml
   # Always include tenant UID
   nameTemplate: "{{ .uid }}-{{ .resourceType }}"

   # For shared namespaces, add more context
   nameTemplate: "{{ .registryId }}-{{ .uid }}-{{ .resourceType }}"

   # Ensure K8s compliance
   nameTemplate: "{{ printf \"%s-%s\" .uid .resourceType | trunc63 }}"
   ```

2. **Set appropriate conflict policies:**
   ```yaml
   # Default for most resources (safe)
   conflictPolicy: Stuck

   # Only for resources you control
   conflictPolicy: Force
   ```

3. **Namespace isolation:**
   - Use dedicated namespaces per registry
   - Use per-tenant namespaces where appropriate
   - Avoid mixing manual and operator-managed resources

4. **Test template changes:**
   - Deploy to test environment first
   - Verify name uniqueness
   - Check for conflicts with existing resources
   - Use canary deployments for template updates

5. **Monitor conflict metrics:**
   ```promql
   # Set alerts for sustained conflict rates
   rate(tenant_conflicts_total[5m]) > 0.05

   # Monitor by resource kind
   rate(tenant_conflicts_total[5m]) by (resource_kind)
   ```

## Monitoring

```promql
# Conflict rate per tenant
rate(tenant_conflicts_total[5m]) by (tenant, namespace)

# Conflicts by resource kind
sum(rate(tenant_conflicts_total[5m])) by (resource_kind)

# Conflicts by policy
sum(rate(tenant_conflicts_total[5m])) by (conflict_policy)

# Total conflicts in timeframe
increase(tenant_conflicts_total[1h])

# Tenants with conflicts
count(rate(tenant_conflicts_total[5m]) > 0)
```

## Analysis Queries

```bash
# Find top conflicting tenants
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=10000 \
  | grep -i conflict \
  | awk '{for(i=1;i<=NF;i++) if($i~/tenant=/) print $i}' \
  | sort | uniq -c | sort -rn | head -10

# Find top conflicting resource kinds
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=10000 \
  | grep -i conflict \
  | awk '{for(i=1;i<=NF;i++) if($i~/kind=/) print $i}' \
  | sort | uniq -c | sort -rn

# Check conflict time distribution
kubectl logs -n tenant-operator-system -l app=tenant-operator --since=1h \
  | grep -i conflict \
  | awk '{print $1" "$2}' \
  | cut -d: -f1 \
  | uniq -c
```

## Related Alerts

- `TenantResourcesConflicted` - Shows current conflicted state
- `TenantNewConflictsDetected` - Info alert for new conflicts
- `TenantDegraded` - May fire if conflicts cause degradation
- `TenantReconciliationErrors` - High error rate may correlate

## Best Practices

1. **Naming strategy:**
   - Always include unique tenant identifier
   - Use consistent patterns across templates
   - Test naming uniqueness before deployment
   - Document naming conventions

2. **Conflict policy strategy:**
   - Start with `Stuck` (safe default)
   - Only use `Force` when necessary and documented
   - Review and justify all `Force` usages
   - Test policy changes in non-production

3. **Resource management:**
   - Avoid manual resource creation in operator namespaces
   - Use labels to identify resource ownership
   - Implement proper cleanup on tenant deletion
   - Coordinate with other controllers

4. **Monitoring and alerting:**
   - Monitor conflict rates continuously
   - Set alerts for sustained high rates
   - Track conflicts by resource kind
   - Review conflict patterns regularly

## Escalation

If high conflict rate persists:

1. Collect detailed diagnostics:
   ```bash
   # Tenant details
   kubectl get tenant <tenant-name> -n <namespace> -o yaml > tenant.yaml

   # Template
   kubectl get tenanttemplate <template-name> -n <namespace> -o yaml > template.yaml

   # Conflicting resources
   kubectl get all -n <namespace> -o yaml > resources.yaml

   # Conflict history from logs
   kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=5000 \
     | grep -i "conflict.*<tenant-name>" > conflicts.log

   # Events
   kubectl get events -n <namespace> --sort-by='.lastTimestamp' > events.txt
   ```

2. Analyze patterns:
   - Is it isolated to one tenant or systemic?
   - Is it specific to one resource kind?
   - Did it start after a recent change?

3. If unable to resolve:
   - Review with platform engineering team
   - Check for known issues in operator repository
   - Open GitHub issue with diagnostics and analysis
