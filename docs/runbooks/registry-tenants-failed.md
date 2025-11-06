# Runbook: Registry Tenants Failed

## Alert Details

**Alert Name:** `RegistryTenantsFailure`
**Severity:** Warning
**Threshold:** registry_failed > 0 AND registry_failed <= 5 for 10+ minutes

## Description

This alert fires when a TenantRegistry has a small number of failed tenants (1-5). This indicates isolated tenant failures rather than a systemic issue, typically caused by tenant-specific configuration problems or resource issues.

## Symptoms

- Registry status shows some failed tenants
- Most tenants are healthy, but 1-5 are failing
- Failed tenants may have unique characteristics
- Registry overall is functional

## Possible Causes

1. **Tenant-Specific Data Issues**
   - Invalid data in specific tenant rows
   - Missing required fields for certain tenants
   - Malformed URLs or identifiers
   - NULL or empty required values

2. **Template Variable Issues**
   - Specific tenant data doesn't match template expectations
   - Optional variables missing for some tenants
   - Type mismatches in tenant data

3. **Resource Conflicts**
   - Specific tenant names conflict with existing resources
   - Isolated naming collisions
   - Pre-existing resources blocking creation

4. **Resource Quota Issues**
   - Namespace quota reached affecting new tenants
   - Specific tenant resource requests exceed limits
   - Storage quota exhausted

5. **Tenant-Specific Configuration**
   - Invalid configurations in extraValueMappings
   - Incorrect domain names or URLs
   - Certificate or secret reference issues

## Diagnosis

### 1. Identify Failed Tenants

```bash
# Get registry status
kubectl get tenantregistry <registry-name> -n <namespace> \
  -o jsonpath='Failed: {.status.failed}/{.status.desired}{"\n"}'

# List all tenants from registry
kubectl get tenants -n <namespace> -l kubernetes-tenants.org/registry=<registry-name>

# Filter failed tenants
kubectl get tenants -n <namespace> \
  -l kubernetes-tenants.org/registry=<registry-name> \
  -o json | jq -r '.items[] | select(.status.conditions[]? | select(.type=="Ready" and .status=="False")) | .metadata.name'
```

### 2. Analyze Failed Tenant Pattern

```bash
# Get failure reasons
for tenant in $(kubectl get tenants -n <namespace> \
  -l kubernetes-tenants.org/registry=<registry-name> \
  -o json | jq -r '.items[] | select(.status.conditions[]? | select(.type=="Ready" and .status=="False")) | .metadata.name'); do

  echo "=== $tenant ==="
  kubectl get tenant $tenant -n <namespace> \
    -o jsonpath='{.status.conditions[?(@.type=="Degraded")].reason}{": "}{.status.conditions[?(@.type=="Degraded")].message}{"\n"}'
done
```

### 3. Compare Failed vs Successful Tenants

```bash
# Get a successful tenant for comparison
SUCCESS_TENANT=$(kubectl get tenants -n <namespace> \
  -l kubernetes-tenants.org/registry=<registry-name> \
  -o json | jq -r '.items[] | select(.status.conditions[]? | select(.type=="Ready" and .status=="True")) | .metadata.name' | head -1)

FAILED_TENANT=$(kubectl get tenants -n <namespace> \
  -l kubernetes-tenants.org/registry=<registry-name> \
  -o json | jq -r '.items[] | select(.status.conditions[]? | select(.type=="Ready" and .status=="False")) | .metadata.name' | head -1)

# Compare annotations (data from datasource)
echo "=== Successful Tenant Data ==="
kubectl get tenant $SUCCESS_TENANT -n <namespace> -o jsonpath='{.metadata.annotations}' | jq

echo -e "\n=== Failed Tenant Data ==="
kubectl get tenant $FAILED_TENANT -n <namespace> -o jsonpath='{.metadata.annotations}' | jq
```

### 4. Check Failed Tenant Events

```bash
for tenant in $(kubectl get tenants -n <namespace> \
  -l kubernetes-tenants.org/registry=<registry-name> \
  -o json | jq -r '.items[] | select(.status.conditions[]? | select(.type=="Ready" and .status=="False")) | .metadata.name'); do

  echo "=== Events for $tenant ==="
  kubectl describe tenant $tenant -n <namespace> | grep -A 20 "Events:"
done
```

### 5. Check Datasource for Failed Tenants

```bash
# Get tenant UIDs from failed tenants
FAILED_UIDS=$(kubectl get tenants -n <namespace> \
  -l kubernetes-tenants.org/registry=<registry-name> \
  -o json | jq -r '.items[] | select(.status.conditions[]? | select(.type=="Ready" and .status=="False")) | .metadata.annotations."tenant.operator.kubernetes-tenants.org/uid"')

echo "Failed tenant UIDs: $FAILED_UIDS"

# Query datasource directly to check data
# For MySQL:
kubectl run mysql-client --rm -it --restart=Never --image=mysql:8 -- \
  mysql -h <host> -u <user> -p<password> <database> \
  -e "SELECT * FROM <table> WHERE uid IN ($FAILED_UIDS)"
```

### 6. Check Operator Logs

```bash
# Get logs for failed tenant reconciliation
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 \
  | grep -E "$(kubectl get tenants -n <namespace> \
      -l kubernetes-tenants.org/registry=<registry-name> \
      -o json | jq -r '.items[] | select(.status.conditions[]? | select(.type=="Ready" and .status=="False")) | .metadata.name' | tr '\n' '|' | sed 's/|$//')"
```

## Resolution

### For Data Quality Issues

1. **Validate tenant data:**
   ```bash
   # Check for NULL or empty required fields
   # Query datasource directly
   mysql -h <host> -u <user> -p<password> <database> -e "
   SELECT uid, hostOrUrl, activate
   FROM <table>
   WHERE uid IS NULL
      OR hostOrUrl IS NULL
      OR activate IS NULL
      OR hostOrUrl = ''
      OR TRIM(hostOrUrl) = ''
   "
   ```

2. **Fix data in datasource:**
   ```sql
   -- Fix specific tenant data
   UPDATE <table>
   SET hostOrUrl = 'https://tenant.example.com'
   WHERE uid = '<failed-tenant-uid>';

   -- Or deactivate invalid tenants temporarily
   UPDATE <table>
   SET activate = 0
   WHERE uid = '<failed-tenant-uid>';
   ```

3. **Wait for next sync or trigger manually:**
   ```bash
   # Registry controller will sync on next interval
   # Or delete and recreate tenant to force immediate sync
   kubectl delete tenant <failed-tenant-name> -n <namespace>
   # It will be recreated on next registry sync
   ```

### For Template Variable Issues

1. **Check what variables failed tenant needs:**
   ```bash
   kubectl get tenant <failed-tenant-name> -n <namespace> \
     -o yaml | grep -A 50 "spec:"
   ```

2. **Add default values to template:**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>

   # Use default filters for optional fields:
   # image: "{{ .deployImage | default \"nginx:stable\" }}"
   # replicas: {{ .replicas | default 1 }}
   # value: "{{ .optionalField | default \"default-value\" }}"
   ```

3. **Or add missing extraValueMappings:**
   ```bash
   kubectl edit tenantregistry <registry-name> -n <namespace>

   # Add missing column mappings:
   # spec:
   #   source:
   #     extraValueMappings:
   #       deployImage: deploy_image_column
   #       replicas: replica_count_column
   ```

### For Resource Conflicts

1. **Identify conflicting resources:**
   ```bash
   kubectl describe tenant <failed-tenant-name> -n <namespace> | grep -i conflict
   ```

2. **Remove conflicting resources:**
   ```bash
   # Backup first
   kubectl get <resource-kind> <resource-name> -n <namespace> -o yaml > backup.yaml

   # Delete conflicting resource
   kubectl delete <resource-kind> <resource-name> -n <namespace>
   ```

3. **Or adjust naming pattern:**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>
   # Ensure nameTemplate includes tenant UID for uniqueness
   ```

### For Quota Issues

1. **Check namespace quotas:**
   ```bash
   kubectl describe resourcequota -n <namespace>
   kubectl describe limitrange -n <namespace>
   ```

2. **Increase quota if appropriate:**
   ```bash
   kubectl edit resourcequota <quota-name> -n <namespace>
   # Increase limits
   ```

3. **Or reduce tenant resource requests:**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>
   # Reduce CPU/memory requests
   ```

### Force Single Tenant Reconciliation

```bash
# Delete and recreate specific failed tenant
kubectl delete tenant <failed-tenant-name> -n <namespace>

# It will be recreated on next registry sync (every syncInterval)
# Or annotate registry to force immediate sync
kubectl annotate tenantregistry <registry-name> -n <namespace> \
  tenant.operator.kubernetes-tenants.org/sync="$(date +%s)" --overwrite
```

## Batch Fix for Multiple Failed Tenants

If all failed tenants have same root cause:

```bash
# Fix root cause (template, registry, quotas, etc.)
kubectl edit tenanttemplate <template-name> -n <namespace>

# Force reconciliation of all failed tenants
for tenant in $(kubectl get tenants -n <namespace> \
  -l kubernetes-tenants.org/registry=<registry-name> \
  -o json | jq -r '.items[] | select(.status.conditions[]? | select(.type=="Ready" and .status=="False")) | .metadata.name'); do

  echo "Reconciling $tenant"
  kubectl annotate tenant $tenant -n <namespace> \
    tenant.operator.kubernetes-tenants.org/reconcile="$(date +%s)" --overwrite

  sleep 2  # Avoid overwhelming operator
done

# Monitor recovery
watch 'kubectl get tenants -n <namespace> -l kubernetes-tenants.org/registry=<registry-name> --no-headers | awk "{print \$2}" | sort | uniq -c'
```

## Prevention

1. **Data validation at source:**
   - Implement database constraints
   - Validate data before insertion
   - Use NOT NULL constraints on required columns
   - Validate URL formats

   ```sql
   ALTER TABLE tenants
   ADD CONSTRAINT check_uid CHECK (uid IS NOT NULL AND uid != '');

   ALTER TABLE tenants
   ADD CONSTRAINT check_url CHECK (hostOrUrl LIKE 'http%://%');
   ```

2. **Template robustness:**
   ```yaml
   # Always use default values for optional fields
   image: "{{ .deployImage | default \"nginx:stable\" }}"
   replicas: {{ .replicas | default 1 }}
   value: "{{ .optionalValue | default \"\" }}"

   # Validate and sanitize inputs
   name: "{{ .uid | trunc63 | lower }}"
   ```

3. **Comprehensive testing:**
   - Test templates with edge cases
   - Include tenants with minimal data
   - Test with NULL or empty values
   - Validate all code paths

4. **Monitoring and alerting:**
   - Alert on any failed tenants
   - Track failure patterns
   - Monitor datasource data quality

5. **Gradual rollout:**
   - Start with subset of tenants
   - Validate before enabling all
   - Use canary deployments

## Monitoring

```promql
# Failed tenant count
registry_failed{registry="<registry-name>"}

# Failed tenant percentage
(registry_failed / registry_desired) * 100

# Tenants becoming failed
increase(registry_failed{registry="<registry-name>"}[10m]) > 0

# Specific tenant status
tenant_condition_status{tenant="<tenant-name>", type="Ready"}
```

## Differentiation from Other Alerts

| Alert | Condition | Severity | Meaning |
|-------|-----------|----------|---------|
| RegistryTenantsFailure | 1-5 failed | Warning | Isolated failures - tenant-specific |
| RegistryManyTenantsFailure | >5 or >50% | Critical | Systemic issue - registry-wide |

## Related Alerts

- `RegistryManyTenantsFailure` - Escalates to if more tenants fail
- `TenantResourcesFailed` - Individual tenant resource failures
- `TenantDegraded` - Failed tenants will be degraded
- `TenantNotReady` - Failed tenants won't be ready

## Investigation Checklist

- [ ] Identify which specific tenants are failing
- [ ] Check failure reasons from tenant conditions
- [ ] Compare failed tenant data with successful ones
- [ ] Validate tenant data in datasource
- [ ] Check for common patterns in failed tenants
- [ ] Review recent template or registry changes
- [ ] Check namespace quotas and limits
- [ ] Verify no resource naming conflicts
- [ ] Check operator logs for specific errors
- [ ] Test fix with one tenant before batch apply

## Escalation

If failures persist or pattern is unclear:

1. Collect diagnostics:
   ```bash
   # Registry details
   kubectl get tenantregistry <registry-name> -n <namespace> -o yaml > registry.yaml

   # All tenants status
   kubectl get tenants -n <namespace> -l kubernetes-tenants.org/registry=<registry-name> -o yaml > tenants.yaml

   # Failed tenants details
   for tenant in $(kubectl get tenants -n <namespace> \
     -l kubernetes-tenants.org/registry=<registry-name> \
     -o json | jq -r '.items[] | select(.status.conditions[]? | select(.type=="Ready" and .status=="False")) | .metadata.name'); do
     kubectl describe tenant $tenant -n <namespace> > $tenant-describe.txt
   done

   # Operator logs filtered for failed tenants
   kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=1000 > operator.log

   # Datasource query results for failed tenants
   # Export data for failed tenant UIDs
   ```

2. Analyze for patterns:
   - Are failures correlated with specific data patterns?
   - Did failures start after a specific change?
   - Is there a common error message?

3. If unable to resolve:
   - Escalate to platform team
   - Review datasource data quality
   - Check for operator bugs
   - Open GitHub issue with diagnostics
