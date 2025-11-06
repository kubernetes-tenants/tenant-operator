# Runbook: Registry Sync Issues

## Alert Details

**Alert Name:** `RegistryDesiredCountMismatch`
**Severity:** Warning
**Threshold:** registry_ready != registry_desired AND registry_desired > 0 AND registry_failed == 0 for 20+ minutes

## Description

This alert fires when a registry's ready tenant count doesn't match the desired count, but there are no failed tenants. This indicates synchronization delays or issues with the registry controller's ability to create/update tenants, but not hard failures.

## Symptoms

- Registry ready count doesn't match desired count
- No failed tenants reported
- Tenants may be slowly provisioning
- New tenants from datasource not appearing
- Deleted datasource rows not cleaned up

## Possible Causes

1. **Slow Tenant Provisioning**
   - Many tenants being created simultaneously
   - Individual tenant resources taking time to become ready
   - Cluster under load causing delays

2. **Registry Controller Issues**
   - Registry controller not running or restarting
   - Sync interval too long
   - Rate limiting affecting sync operations
   - Controller goroutines blocked

3. **Datasource Query Issues**
   - Slow database queries
   - Database connection intermittent
   - Query timeouts
   - Large result sets

4. **Tenant Creation Bottleneck**
   - API server throttling requests
   - etcd performance issues
   - Controller concurrency limits
   - Workqueue backed up

5. **Cache Synchronization Issues**
   - Informer cache not syncing
   - Watch connections dropped
   - Missed events
   - Stale cache data

## Diagnosis

### 1. Check Registry Status

```bash
# Get registry counts
kubectl get tenantregistry <registry-name> -n <namespace> \
  -o jsonpath='Desired: {.status.desired}{"\n"}Ready: {.status.ready}{"\n"}Failed: {.status.failed}{"\n"}'

# Check last sync time
kubectl get tenantregistry <registry-name> -n <namespace> \
  -o jsonpath='{.status.lastSyncTime}'

# Check sync interval
kubectl get tenantregistry <registry-name> -n <namespace> \
  -o jsonpath='{.spec.source.syncInterval}'
```

### 2. Identify Missing or Extra Tenants

```bash
# Count actual tenants
kubectl get tenants -n <namespace> \
  -l kubernetes-tenants.org/registry=<registry-name> --no-headers | wc -l

# Check tenant statuses
kubectl get tenants -n <namespace> \
  -l kubernetes-tenants.org/registry=<registry-name> \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.conditions[?(@.type=="Ready")].status}{"\n"}{end}' \
  | awk '{print $2}' | sort | uniq -c

# List tenants with their ready status
kubectl get tenants -n <namespace> \
  -l kubernetes-tenants.org/registry=<registry-name> \
  -o custom-columns=NAME:.metadata.name,READY:.status.conditions[?(@.type==\"Ready\")].status,RESOURCES:.status.readyResources/.status.desiredResources
```

### 3. Check Registry Controller Health

```bash
# Check operator pods
kubectl get pods -n tenant-operator-system -l app=tenant-operator

# Check for recent restarts
kubectl get pods -n tenant-operator-system -l app=tenant-operator \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.containerStatuses[0].restartCount}{"\n"}{end}'

# Check controller logs for registry sync
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=200 \
  | grep "registry=<registry-name>"

# Look for sync completion messages
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 \
  | grep -i "synced.*<registry-name>\|sync.*complete"
```

### 4. Check Datasource Connectivity

```bash
# Get datasource configuration
kubectl get tenantregistry <registry-name> -n <namespace> -o yaml

# Test database connection
PASS_SECRET=$(kubectl get tenantregistry <registry-name> -n <namespace> \
  -o jsonpath='{.spec.source.mysql.passwordRef.name}')

kubectl run mysql-test --rm -it --restart=Never --image=mysql:8 \
  --env="MYSQL_PWD=$(kubectl get secret $PASS_SECRET -n <namespace> -o jsonpath='{.data.password}' | base64 -d)" \
  -- mysql -h <host> -u <user> -D <database> \
  -e "SELECT COUNT(*) as count, SUM(activate=1) as active FROM <table>"
```

### 5. Check for Pending Tenant Creation

```bash
# Check operator workqueue depth
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=100 \
  | grep -i "workqueue\|queue depth"

# Check for rate limiting
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 \
  | grep -i "rate limit\|throttl"

# Check API server response times
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 \
  | grep -i "request took\|duration"
```

### 6. Compare Datasource with Kubernetes State

```bash
# Get active tenant UIDs from datasource
# Example for MySQL:
kubectl run mysql-client --rm -it --restart=Never --image=mysql:8 -- \
  mysql -h <host> -u <user> -p<password> <database> \
  -e "SELECT uid FROM <table> WHERE activate=1" \
  | tail -n +2 | sort > /tmp/datasource-uids.txt

# Get tenant UIDs from Kubernetes
kubectl get tenants -n <namespace> \
  -l kubernetes-tenants.org/registry=<registry-name> \
  -o jsonpath='{range .items[*]}{.metadata.annotations.tenant\.operator\.kubernetes-tenants\.org/uid}{"\n"}{end}' \
  | sort > /tmp/k8s-uids.txt

# Find differences
echo "=== In datasource but not in K8s ==="
comm -23 /tmp/datasource-uids.txt /tmp/k8s-uids.txt

echo -e "\n=== In K8s but not in datasource ==="
comm -13 /tmp/datasource-uids.txt /tmp/k8s-uids.txt
```

## Resolution

### For Slow Tenant Provisioning

1. **Wait for provisioning to complete:**
   ```bash
   # Monitor tenant readiness progress
   watch 'kubectl get tenants -n <namespace> \
     -l kubernetes-tenants.org/registry=<registry-name> \
     --no-headers | awk "{print \$2}" | sort | uniq -c'
   ```

2. **Check individual tenant progress:**
   ```bash
   # Find tenants not ready
   kubectl get tenants -n <namespace> \
     -l kubernetes-tenants.org/registry=<registry-name> \
     -o json | jq -r '.items[] | select(.status.conditions[]? | select(.type=="Ready" and .status!="True")) | .metadata.name'

   # Check their resource status
   for tenant in $(kubectl get tenants -n <namespace> \
     -l kubernetes-tenants.org/registry=<registry-name> \
     -o json | jq -r '.items[] | select(.status.conditions[]? | select(.type=="Ready" and .status!="True")) | .metadata.name' | head -5); do
     echo "=== $tenant ==="
     kubectl get tenant $tenant -n <namespace> \
       -o jsonpath='{.status.readyResources}/{.status.desiredResources}{"\n"}'
   done
   ```

3. **If consistently slow, optimize:**
   - Increase controller concurrency
   - Optimize tenant templates
   - Add more controller replicas

### For Registry Controller Issues

1. **Restart controller if unhealthy:**
   ```bash
   kubectl delete pod -n tenant-operator-system -l app=tenant-operator

   # Wait for new pod
   kubectl wait --for=condition=ready pod -n tenant-operator-system \
     -l app=tenant-operator --timeout=60s
   ```

2. **Check controller resources:**
   ```bash
   kubectl top pods -n tenant-operator-system -l app=tenant-operator

   # Check if OOMKilled
   kubectl get pods -n tenant-operator-system -l app=tenant-operator \
     -o jsonpath='{.items[*].status.containerStatuses[*].lastState.terminated.reason}'
   ```

3. **Increase controller resources if needed:**
   ```bash
   kubectl edit deployment -n tenant-operator-system tenant-operator-controller-manager
   # Increase memory and CPU limits
   ```

### For Datasource Issues

1. **Verify connectivity:**
   ```bash
   # Test from operator pod
   OPERATOR_POD=$(kubectl get pods -n tenant-operator-system \
     -l app=tenant-operator -o jsonpath='{.items[0].metadata.name}')

   kubectl exec -n tenant-operator-system $OPERATOR_POD -- \
     sh -c "nc -zv <database-host> <port>"
   ```

2. **Check query performance:**
   ```bash
   # Time the query
   time mysql -h <host> -u <user> -p<password> <database> \
     -e "SELECT * FROM <table> WHERE activate=1"

   # Check for slow queries
   mysql -h <host> -u <user> -p<password> \
     -e "SHOW PROCESSLIST"
   ```

3. **Optimize if slow:**
   - Add database indexes
   - Reduce result set size
   - Optimize query
   - Increase database resources

### For Cache Synchronization Issues

1. **Force registry sync:**
   ```bash
   # Annotate to trigger immediate sync
   kubectl annotate tenantregistry <registry-name> -n <namespace> \
     tenant.operator.kubernetes-tenants.org/sync="$(date +%s)" --overwrite
   ```

2. **Restart controller to reset caches:**
   ```bash
   kubectl delete pod -n tenant-operator-system -l app=tenant-operator
   ```

### For Missing Tenants

```bash
# Identify missing tenant UIDs
comm -23 /tmp/datasource-uids.txt /tmp/k8s-uids.txt > /tmp/missing-uids.txt

# Force registry sync to create them
kubectl annotate tenantregistry <registry-name> -n <namespace> \
  tenant.operator.kubernetes-tenants.org/sync="$(date +%s)" --overwrite

# Monitor creation
watch 'kubectl get tenants -n <namespace> -l kubernetes-tenants.org/registry=<registry-name> | wc -l'
```

### For Orphaned Tenants

```bash
# Identify orphaned tenant UIDs (in K8s but not in datasource)
comm -13 /tmp/datasource-uids.txt /tmp/k8s-uids.txt > /tmp/orphan-uids.txt

# These should be garbage collected on next sync
# If not, manually delete:
for uid in $(cat /tmp/orphan-uids.txt); do
  TENANT_NAME=$(kubectl get tenants -n <namespace> \
    -l kubernetes-tenants.org/registry=<registry-name> \
    -o json | jq -r ".items[] | select(.metadata.annotations.\"tenant.operator.kubernetes-tenants.org/uid\"==\"$uid\") | .metadata.name")

  if [ -n "$TENANT_NAME" ]; then
    echo "Deleting orphaned tenant: $TENANT_NAME (uid: $uid)"
    kubectl delete tenant $TENANT_NAME -n <namespace>
  fi
done
```

### Adjust Sync Interval

If sync is too infrequent:

```bash
kubectl patch tenantregistry <registry-name> -n <namespace> --type=merge \
  -p '{"spec":{"source":{"syncInterval":"5m"}}}'  # Increase frequency

# Or decrease if too frequent and causing load
kubectl patch tenantregistry <registry-name> -n <namespace> --type=merge \
  -p '{"spec":{"source":{"syncInterval":"15m"}}}'  # Decrease frequency
```

## Prevention

1. **Appropriate sync interval:**
   ```yaml
   spec:
     source:
       syncInterval: "5m"  # Balance between freshness and load
   ```

2. **Datasource optimization:**
   - Add indexes on uid and activate columns
   - Optimize query performance
   - Use connection pooling
   - Monitor database health

3. **Controller sizing:**
   ```yaml
   # Adjust controller concurrency
   --registry-concurrency=5  # Increase if needed
   --tenant-concurrency=20

   # Adequate resources
   resources:
     requests:
       memory: "256Mi"
       cpu: "500m"
     limits:
       memory: "1Gi"
       cpu: "2000m"
   ```

4. **Monitoring:**
   - Track sync completion time
   - Monitor desired vs ready gap
   - Alert on sync delays
   - Track datasource query performance

5. **Capacity planning:**
   - Plan for peak tenant count
   - Ensure cluster has capacity
   - Test with expected load
   - Monitor growth trends

## Monitoring

```promql
# Desired vs ready gap
registry_desired - registry_ready != 0

# Gap percentage
((registry_desired - registry_ready) / registry_desired) * 100

# Time since last successful sync
time() - registry_last_sync_time > 600  # 10 minutes

# Datasource query duration
datasource_query_duration_seconds{registry="<registry-name>"}

# Tenant creation rate
rate(registry_desired[5m])
```

## Related Alerts

- `RegistryTenantsFailure` - Check if failures develop
- `RegistryManyTenantsFailure` - Escalates if many fail
- `TenantResourcesMismatch` - Individual tenant sync issues

## Performance Tuning

```bash
# Check current controller settings
kubectl get deployment -n tenant-operator-system \
  tenant-operator-controller-manager -o yaml | grep -A 10 args:

# Increase concurrency
kubectl set env deployment/tenant-operator-controller-manager \
  -n tenant-operator-system \
  --containers=manager \
  REGISTRY_CONCURRENCY=5 \
  TENANT_CONCURRENCY=20

# Monitor impact
kubectl logs -n tenant-operator-system -l app=tenant-operator --follow \
  | grep -i "reconcile\|sync"
```

## Escalation

If sync issues persist:

1. Collect diagnostics:
   ```bash
   # Registry details
   kubectl get tenantregistry <registry-name> -n <namespace> -o yaml > registry.yaml

   # Tenant counts
   kubectl get tenants -n <namespace> -l kubernetes-tenants.org/registry=<registry-name> > tenants-list.txt

   # Datasource query results
   mysql -h <host> -u <user> -p<password> <database> \
     -e "SELECT uid, activate FROM <table>" > datasource-data.txt

   # Operator logs
   kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=2000 > operator.log

   # Controller metrics
   kubectl top pods -n tenant-operator-system -l app=tenant-operator > controller-resources.txt
   ```

2. Check timing:
   - When did mismatch start?
   - Recent datasource changes?
   - Recent operator updates?
   - Cluster capacity changes?

3. If unable to resolve:
   - Review with platform team
   - Check datasource health and performance
   - Verify network connectivity
   - Open GitHub issue with diagnostics
