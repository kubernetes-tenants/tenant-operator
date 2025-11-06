# Runbook: Registry Many Tenants Failed

## Alert Details

**Alert Name:** `RegistryManyTenantsFailure`
**Severity:** Critical
**Threshold:** registry_failed > 5 OR (registry_failed / registry_desired > 0.5 AND registry_desired > 0) for 5+ minutes

## Description

This alert fires when a TenantRegistry has many failed tenants (more than 5 absolute, or more than 50% of desired count). This indicates a systemic issue affecting multiple tenants, likely caused by problems with the registry itself, templates, or infrastructure rather than individual tenant issues.

## Symptoms

- Multiple tenants from the same registry are failing
- Registry status shows high failure count
- New tenants may continue to fail
- Widespread service disruptions for multiple customers

## Possible Causes

1. **Registry Data Source Issues**
   - Database connection failures
   - Credential authentication failures
   - Query syntax errors
   - Network connectivity to data source

2. **Template Problems**
   - Invalid template syntax affecting all tenants
   - Missing required template functions
   - Schema validation errors in template output
   - Incorrect resource specifications

3. **Infrastructure Problems**
   - Cluster resource exhaustion
   - Storage class unavailable
   - Network policy blocking traffic
   - Ingress controller issues

4. **RBAC or Permission Issues**
   - Operator permissions revoked or misconfigured
   - Namespace quota restrictions
   - Security policy violations

5. **Configuration Issues**
   - Invalid value mappings in registry
   - Incorrect extraValueMappings
   - Template references non-existent resources

## Diagnosis

### 1. Check Registry Status

```bash
# Get registry overview
kubectl get tenantregistry <registry-name> -n <namespace>

# Detailed status
kubectl get tenantregistry <registry-name> -n <namespace> -o yaml | grep -A 20 "status:"

# Check failure count
kubectl get tenantregistry <registry-name> -n <namespace> \
  -o jsonpath='Failed: {.status.failed}/{.status.desired}{"\n"}'
```

### 2. Identify Failed Tenants Pattern

```bash
# List all tenants from this registry
kubectl get tenants -n <namespace> -l kubernetes-tenants.org/registry=<registry-name>

# Count tenants by status
kubectl get tenants -n <namespace> -l kubernetes-tenants.org/registry=<registry-name> \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.conditions[?(@.type=="Ready")].status}{"\n"}{end}' \
  | awk '{print $2}' | sort | uniq -c

# Get common failure reasons
kubectl get tenants -n <namespace> -l kubernetes-tenants.org/registry=<registry-name> \
  -o jsonpath='{range .items[*]}{.status.conditions[?(@.type=="Degraded")].reason}{"\n"}{end}' \
  | sort | uniq -c
```

### 3. Check Registry Data Source Connection

```bash
# Check registry spec
kubectl get tenantregistry <registry-name> -n <namespace> -o yaml

# Check credentials secret
PASS_SECRET=$(kubectl get tenantregistry <registry-name> -n <namespace> -o jsonpath='{.spec.source.mysql.passwordRef.name}')
kubectl get secret $PASS_SECRET -n <namespace>

# Test database connection (if MySQL)
kubectl run mysql-test --rm -it --restart=Never \
  --image=mysql:8 \
  --env="MYSQL_PWD=$(kubectl get secret $PASS_SECRET -n <namespace> -o jsonpath='{.data.password}' | base64 -d)" \
  -- mysql -h <host> -u <user> -e "SELECT 1"
```

### 4. Check Registry Controller Logs

```bash
# Filter logs for this registry
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 \
  | grep "registry=<registry-name>"

# Look for datasource errors
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 \
  | grep -i "datasource\|mysql\|connection\|query"
```

### 5. Check Template Issues

```bash
# Get template referenced by registry
kubectl get tenanttemplates -n <namespace> \
  -o json | jq -r '.items[] | select(.spec.registryId=="<registry-name>") | .metadata.name'

# Check template for errors
kubectl get tenanttemplate <template-name> -n <namespace> -o yaml

# Validate template syntax
kubectl get tenanttemplate <template-name> -n <namespace> -o yaml | grep -A 50 "spec:"
```

### 6. Check Sample Failed Tenant

```bash
# Pick one failed tenant to investigate
FAILED_TENANT=$(kubectl get tenants -n <namespace> \
  -l kubernetes-tenants.org/registry=<registry-name> \
  -o jsonpath='{.items[?(@.status.conditions[0].status=="False")].metadata.name}' | head -1)

echo "Investigating: $FAILED_TENANT"

# Get tenant details
kubectl describe tenant $FAILED_TENANT -n <namespace>

# Check tenant events
kubectl get events -n <namespace> --field-selector involvedObject.name=$FAILED_TENANT
```

## Resolution

### For Data Source Connection Issues

1. **Verify credentials:**
   ```bash
   # Check secret exists and has correct keys
   kubectl get secret <password-secret> -n <namespace> -o yaml

   # Update password if needed
   kubectl create secret generic <password-secret> \
     --from-literal=password='<new-password>' \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

2. **Test connection manually:**
   ```bash
   # For MySQL
   kubectl run mysql-client --rm -it --restart=Never --image=mysql:8 -- \
     mysql -h <host> -u <user> -p<password> -e "SELECT COUNT(*) FROM <table>"
   ```

3. **Check network policies:**
   ```bash
   # Check if network policies block operator
   kubectl get networkpolicies -n tenant-operator-system
   kubectl get networkpolicies -n <registry-namespace>
   ```

4. **Update registry configuration:**
   ```bash
   kubectl edit tenantregistry <registry-name> -n <namespace>
   # Fix host, port, database name, or credentials reference
   ```

### For Template Issues

1. **Validate template syntax:**
   ```bash
   # Get template
   kubectl get tenanttemplate <template-name> -n <namespace> -o yaml > template.yaml

   # Review for common issues:
   # - Missing {{ }} around template variables
   # - Incorrect field names in resource specs
   # - Invalid YAML indentation
   ```

2. **Fix template errors:**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>
   # Common fixes:
   # - Add default values: {{ .field | default "value" }}
   # - Fix field references: .uid instead of .id
   # - Correct resource specifications
   ```

3. **Test template with sample data:**
   ```bash
   # Create a test tenant manually
   kubectl create -f - <<EOF
   apiVersion: operator.kubernetes-tenants.org/v1
   kind: Tenant
   metadata:
     name: test-tenant
     namespace: <namespace>
     annotations:
       tenant.operator.kubernetes-tenants.org/uid: "test-123"
       tenant.operator.kubernetes-tenants.org/host: "test.example.com"
   spec:
     # ... copy from failed tenant
   EOF

   # Watch reconciliation
   kubectl describe tenant test-tenant -n <namespace>
   ```

### For Infrastructure Issues

1. **Check cluster capacity:**
   ```bash
   kubectl top nodes
   kubectl describe nodes | grep -A 5 "Allocated resources"

   # Check if any nodes are under pressure
   kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.conditions[?(@.type=="Ready")].status}{"\n"}{end}'
   ```

2. **Check namespace quotas:**
   ```bash
   kubectl describe resourcequota -n <namespace>

   # Increase quota if needed
   kubectl edit resourcequota <quota-name> -n <namespace>
   ```

3. **Check storage class:**
   ```bash
   kubectl get storageclass
   kubectl describe storageclass <class-name>

   # Set default storage class if missing
   kubectl patch storageclass <class-name> -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
   ```

### For RBAC Issues

1. **Verify operator permissions:**
   ```bash
   # Check critical permissions
   for resource in deployments services configmaps secrets ingresses; do
     echo -n "$resource: "
     kubectl auth can-i create $resource \
       --as=system:serviceaccount:tenant-operator-system:tenant-operator-controller-manager \
       -n <namespace>
   done
   ```

2. **Restore RBAC if needed:**
   ```bash
   # Reapply operator RBAC manifests
   kubectl apply -f config/rbac/
   ```

### Emergency Mitigation

1. **Pause registry reconciliation temporarily:**
   ```bash
   # Annotate registry to pause sync
   kubectl annotate tenantregistry <registry-name> -n <namespace> \
     tenant.operator.kubernetes-tenants.org/pause=true
   ```

2. **Scale down new tenant creation:**
   ```bash
   # Reduce sync frequency temporarily
   kubectl patch tenantregistry <registry-name> -n <namespace> --type=merge \
     -p '{"spec":{"source":{"syncInterval":"1h"}}}'
   ```

3. **Fix root cause, then resume:**
   ```bash
   # After fixing issues, resume sync
   kubectl annotate tenantregistry <registry-name> -n <namespace> \
     tenant.operator.kubernetes-tenants.org/pause-

   # Restore sync interval
   kubectl patch tenantregistry <registry-name> -n <namespace> --type=merge \
     -p '{"spec":{"source":{"syncInterval":"5m"}}}'
   ```

## Batch Recovery

After fixing root cause, recover failed tenants:

```bash
# Force reconciliation of all failed tenants
for tenant in $(kubectl get tenants -n <namespace> \
  -l kubernetes-tenants.org/registry=<registry-name> \
  -o jsonpath='{.items[?(@.status.conditions[0].status=="False")].metadata.name}'); do

  echo "Reconciling $tenant"
  kubectl annotate tenant $tenant -n <namespace> \
    tenant.operator.kubernetes-tenants.org/reconcile="$(date +%s)" --overwrite

  sleep 2  # Avoid overwhelming the operator
done

# Monitor recovery progress
watch 'kubectl get tenants -n <namespace> -l kubernetes-tenants.org/registry=<registry-name> | grep -c "True.*Ready"'
```

## Prevention

1. **Monitor data source health:**
   - Set up database monitoring and alerts
   - Configure connection pooling and timeouts
   - Use database high availability setups

2. **Validate templates before deployment:**
   - Test templates in staging environment
   - Use dry-run validation if available
   - Maintain template version control

3. **Capacity planning:**
   - Monitor cluster resource utilization
   - Set up auto-scaling for nodes
   - Reserve capacity for tenant growth

4. **Gradual rollout:**
   - Start with small tenant batches
   - Monitor initial tenants before scaling
   - Use canary deployments for template changes

5. **Implement circuit breakers:**
   - Limit concurrent tenant creation
   - Throttle registry sync on repeated failures
   - Auto-pause on high error rates

## Metrics to Monitor

```promql
# Failed tenant count
registry_failed{registry="<registry-name>"}

# Failure rate
registry_failed / registry_desired > 0.5

# Recent tenant failures
increase(registry_failed{registry="<registry-name>"}[10m]) > 3

# Datasource query duration
datasource_query_duration_seconds{registry="<registry-name>"}

# Template rendering errors
rate(template_render_errors_total[5m]) > 0
```

## Related Alerts

- `RegistryTenantsFailure` - May fire first for smaller counts
- `TenantResourcesFailed` - Individual tenant failures
- `TenantDegraded` - Degraded state for specific tenants
- `TenantReconciliationErrors` - High error rate may precede this

## Escalation

This is a critical systemic issue requiring immediate attention:

1. **Immediate actions:**
   - Investigate registry data source connectivity
   - Check template validity
   - Review recent changes to registry or templates

2. **Collect diagnostics:**
   ```bash
   # Registry details
   kubectl get tenantregistry <registry-name> -n <namespace> -o yaml > registry.yaml

   # All failed tenants
   kubectl get tenants -n <namespace> -l kubernetes-tenants.org/registry=<registry-name> \
     -o yaml > failed-tenants.yaml

   # Template details
   kubectl get tenanttemplates -n <namespace> -o yaml > templates.yaml

   # Operator logs
   kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=1000 > operator.log
   ```

3. **If issue persists after troubleshooting:**
   - Escalate to platform engineering team
   - Review recent changes (GitOps commits, database migrations)
   - Consider rollback if caused by recent changes
   - Open GitHub issue with full diagnostics
