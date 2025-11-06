# Runbook: Slow Tenant Reconciliation

## Alert Details

**Alert Name:** `TenantReconciliationSlow`
**Severity:** Warning
**Threshold:** histogram_quantile(0.95, sum(rate(tenant_reconcile_duration_seconds_bucket{result="success"}[5m])) by (le)) > 30 for 15+ minutes

## Description

This alert fires when the 95th percentile of successful tenant reconciliation duration exceeds 30 seconds sustained for 15+ minutes. This indicates performance degradation in the reconciliation process, even though reconciliations are eventually completing successfully.

## Symptoms

- Tenant reconciliations taking much longer than usual
- Delayed status updates for tenants
- Slow propagation of template changes
- Increased time for new tenants to become ready
- High latency in tenant operations

## Possible Causes

1. **Complex Tenant Configurations**
   - Large number of resources per tenant
   - Complex dependency chains
   - Long dependency wait times
   - Many resources with `waitForReady=true`

2. **Resource Provisioning Delays**
   - Slow image pulls
   - Storage provisioning delays
   - LoadBalancer assignment delays
   - DNS propagation delays

3. **Cluster Performance Issues**
   - API server overloaded
   - etcd performance degradation
   - Slow node responses
   - Network latency

4. **Controller Performance**
   - Insufficient controller resources
   - High CPU/memory usage
   - Excessive logging or debugging
   - Inefficient reconciliation logic

5. **External Dependencies**
   - Slow datasource queries
   - External API timeouts
   - Template function delays
   - Custom validation webhooks

6. **Scale Issues**
   - Too many concurrent reconciliations
   - Large cluster with many objects
   - Informer cache performance
   - Watch connection issues

## Diagnosis

### 1. Check Reconciliation Duration Metrics

```bash
# Get operator pod for metrics
OPERATOR_POD=$(kubectl get pods -n tenant-operator-system -l app=tenant-operator -o jsonpath='{.items[0].metadata.name}')

# Port forward to metrics endpoint
kubectl port-forward -n tenant-operator-system $OPERATOR_POD 8080:8080 &

# Query reconciliation duration
curl -s localhost:8080/metrics | grep tenant_reconcile_duration_seconds | grep -v "#"

# Stop port forward
pkill -f "port-forward.*8080:8080"
```

### 2. Analyze Reconciliation Times in Logs

```bash
# Extract reconciliation durations from logs
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=1000 \
  | grep -i "reconcil.*duration\|reconcil.*took" \
  | awk '{print $(NF-1)}' \
  | sort -n \
  | tail -20

# Calculate average from recent logs
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 \
  | grep -i "reconciliation complete" \
  | awk '{sum+=$(NF-1); count++} END {print "Average:", sum/count, "seconds"}'
```

### 3. Identify Slow Tenants

```bash
# Check which tenants are taking longest
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=2000 \
  | grep -E "Reconciling Tenant|reconciliation complete" \
  | grep -B 1 "duration" \
  | grep -A 1 "Reconciling" \
  | awk '/Reconciling/{tenant=$NF} /duration/{print tenant, $(NF-1)}' \
  | sort -k2 -rn \
  | head -10
```

### 4. Check Controller Resource Usage

```bash
# Check CPU and memory usage
kubectl top pods -n tenant-operator-system -l app=tenant-operator --containers

# Check if controller is CPU throttled
kubectl get pods -n tenant-operator-system -l app=tenant-operator \
  -o jsonpath='{range .items[*]}{.spec.containers[*].resources}{"\n"}{end}'
```

### 5. Analyze Tenant Complexity

```bash
# Check resource counts per tenant
kubectl get tenants --all-namespaces \
  -o custom-columns=NAME:.metadata.name,NAMESPACE:.metadata.namespace,DESIRED:.status.desiredResources \
  | sort -k3 -rn | head -20

# Sample a slow tenant's configuration
SLOW_TENANT="<tenant-name>"  # From earlier analysis
kubectl get tenant $SLOW_TENANT -n <namespace> \
  -o jsonpath='{.spec.resources[*].id}' | tr ' ' '\n' | wc -l
echo "Resource count: "

# Check for long dependency chains
kubectl get tenant $SLOW_TENANT -n <namespace> -o yaml | grep -A 5 dependIds
```

### 6. Check API Server Performance

```bash
# Check API server request duration
kubectl get --raw /metrics | grep apiserver_request_duration_seconds_sum

# Check for API server throttling
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 \
  | grep -i "429\|throttl\|rate limit"
```

### 7. Check Resource Readiness Times

```bash
# Check pod startup times
kubectl get pods -n <namespace> -l kubernetes-tenants.org/tenant=<slow-tenant> \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.creationTimestamp}{"\t"}{.status.conditions[?(@.type=="Ready")].lastTransitionTime}{"\n"}{end}'

# Calculate time to ready
kubectl get pods -n <namespace> -l kubernetes-tenants.org/tenant=<slow-tenant> \
  -o json | jq -r '.items[] | "\(.metadata.name)\t\(.metadata.creationTimestamp)\t\(.status.conditions[]? | select(.type=="Ready") | .lastTransitionTime)"'
```

## Resolution

### For Complex Tenant Configurations

1. **Optimize dependency chains:**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>

   # Reduce dependency depth:
   # - Remove unnecessary dependIds
   # - Parallelize independent resources
   # - Only wait for truly critical dependencies
   ```

2. **Disable waiting for non-critical resources:**
   ```yaml
   spec:
     jobs:
       - id: optional-init-job
         waitForReady: false  # Don't block on this
     configMaps:
       - id: optional-config
         waitForReady: false
   ```

3. **Increase timeouts for slow resources:**
   ```yaml
   spec:
     deployments:
       - id: slow-app
         timeoutSeconds: 600  # 10 minutes instead of default 300
   ```

### For Controller Performance

1. **Increase controller resources:**
   ```bash
   kubectl edit deployment -n tenant-operator-system tenant-operator-controller-manager

   # Increase resources:
   resources:
     requests:
       memory: "512Mi"
       cpu: "1000m"
     limits:
       memory: "2Gi"
       cpu: "4000m"
   ```

2. **Adjust concurrency settings:**
   ```bash
   kubectl set env deployment/tenant-operator-controller-manager \
     -n tenant-operator-system \
     --containers=manager \
     TENANT_CONCURRENCY=15  # Increase from default 10

   # Monitor impact on reconciliation times
   ```

3. **Reduce logging verbosity if excessive:**
   ```bash
   kubectl edit deployment -n tenant-operator-system tenant-operator-controller-manager
   # Change --zap-log-level=debug to --zap-log-level=info
   ```

### For API Server Performance

1. **Increase API client rate limits:**
   ```bash
   kubectl edit deployment -n tenant-operator-system tenant-operator-controller-manager

   # Add or update args:
   args:
     - --kube-api-qps=50  # Increase from default
     - --kube-api-burst=100
   ```

2. **Check API server health:**
   ```bash
   # On control plane node
   kubectl top nodes
   kubectl get --raw /metrics | grep apiserver
   ```

### For Resource Provisioning Delays

1. **Optimize image pulls:**
   ```yaml
   # In tenant template
   spec:
     deployments:
       - spec:
           template:
             spec:
               containers:
                 - imagePullPolicy: IfNotPresent  # Avoid always pull
   ```

2. **Pre-pull images on nodes:**
   ```bash
   # Use DaemonSet or pre-pull script
   for node in $(kubectl get nodes -o name); do
     kubectl debug $node -it --image=<your-image> -- /bin/true
   done
   ```

3. **Use local registry or cache:**
   - Set up pull-through cache
   - Use local registry mirror
   - Deploy images to local registry

### For Storage Provisioning

1. **Use faster storage class:**
   ```bash
   # Check available storage classes
   kubectl get storageclass

   # Update template to use faster class
   kubectl edit tenanttemplate <template-name> -n <namespace>
   ```

2. **Pre-provision PVs if possible:**
   ```bash
   # For StatefulSets, consider static provisioning
   # Or use faster dynamic provisioner
   ```

### For Datasource Query Performance

1. **Optimize database queries:**
   ```sql
   -- Add indexes
   CREATE INDEX idx_uid ON tenants(uid);
   CREATE INDEX idx_activate ON tenants(activate);

   -- Analyze query performance
   EXPLAIN SELECT * FROM tenants WHERE activate=1;
   ```

2. **Use query caching:**
   - Enable query cache in database
   - Cache results in operator (if supported)
   - Reduce query frequency

### Temporary Mitigation

If slow reconciliation is causing operational issues:

```bash
# Reduce concurrency to lower load
kubectl set env deployment/tenant-operator-controller-manager \
  -n tenant-operator-system \
  --containers=manager \
  TENANT_CONCURRENCY=5

# Increase sync intervals for registries
kubectl patch tenantregistry <registry-name> -n <namespace> --type=merge \
  -p '{"spec":{"source":{"syncInterval":"15m"}}}'

# Monitor improvement
watch 'kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=20 | grep duration'
```

## Performance Tuning

### Benchmark Current Performance

```bash
# Measure average reconciliation time
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=1000 \
  | grep "reconciliation complete" \
  | awk '{print $(NF-1)}' \
  | awk '{sum+=$1; count++} END {print "Avg:", sum/count, "seconds"}'

# Measure P95
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=1000 \
  | grep "reconciliation complete" \
  | awk '{print $(NF-1)}' \
  | sort -n \
  | awk '{all[NR]=$1} END {print "P95:", all[int(NR*0.95)]}'
```

### Optimize Resource Limits

```yaml
# Start conservative
resources:
  requests:
    memory: "256Mi"
    cpu: "500m"
  limits:
    memory: "1Gi"
    cpu: "2000m"

# Monitor and adjust based on usage
# Goal: No CPU throttling, no OOM kills
```

### Optimize Concurrency

```bash
# Test different concurrency levels
for concurrency in 5 10 15 20; do
  echo "Testing concurrency: $concurrency"
  kubectl set env deployment/tenant-operator-controller-manager \
    -n tenant-operator-system \
    TENANT_CONCURRENCY=$concurrency

  sleep 120  # Let it stabilize

  # Measure reconciliation time
  kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=100 \
    | grep "reconciliation complete" \
    | awk '{sum+=$(NF-1); count++} END {print "Concurrency", '$concurrency', "Avg:", sum/count}'
done
```

## Prevention

1. **Template optimization:**
   - Minimize resource count per tenant
   - Avoid unnecessary dependencies
   - Use `waitForReady: false` for optional resources
   - Set realistic timeouts

2. **Controller sizing:**
   - Adequate CPU and memory
   - Appropriate concurrency settings
   - No CPU throttling
   - Monitor and adjust based on load

3. **Cluster capacity:**
   - Ensure cluster has headroom
   - Fast storage provisioners
   - Low-latency network
   - Healthy API server

4. **Regular optimization:**
   - Profile reconciliation bottlenecks
   - Optimize slow code paths
   - Update to latest operator versions
   - Review and optimize templates

5. **Monitoring:**
   ```promql
   # Track P95 reconciliation time
   histogram_quantile(0.95, rate(tenant_reconcile_duration_seconds_bucket[5m]))

   # Track P99
   histogram_quantile(0.99, rate(tenant_reconcile_duration_seconds_bucket[5m]))

   # Track by tenant
   histogram_quantile(0.95, rate(tenant_reconcile_duration_seconds_bucket[5m])) by (tenant)
   ```

## Monitoring

```promql
# P95 reconciliation duration
histogram_quantile(0.95,
  sum(rate(tenant_reconcile_duration_seconds_bucket{result="success"}[5m])) by (le)
) > 30

# P99 reconciliation duration
histogram_quantile(0.99,
  sum(rate(tenant_reconcile_duration_seconds_bucket{result="success"}[5m])) by (le)
)

# Average reconciliation duration
rate(tenant_reconcile_duration_seconds_sum{result="success"}[5m])
  / rate(tenant_reconcile_duration_seconds_count{result="success"}[5m])

# Controller CPU usage
rate(container_cpu_usage_seconds_total{pod=~"tenant-operator.*"}[5m])

# Controller memory usage
container_memory_working_set_bytes{pod=~"tenant-operator.*"}
```

## Related Alerts

- `TenantReconciliationErrors` - May accompany slow reconciliation
- `TenantNotReady` - Slow reconciliation delays readiness
- `TenantResourcesMismatch` - May be caused by slow reconciliation

## Performance Targets

| Metric | Target | Alert Threshold |
|--------|--------|-----------------|
| P50 reconciliation | < 5s | - |
| P95 reconciliation | < 15s | 30s |
| P99 reconciliation | < 30s | 60s |
| Average reconciliation | < 10s | 20s |

## Escalation

If slow reconciliation persists after optimization:

1. Collect performance diagnostics:
   ```bash
   # Reconciliation metrics
   kubectl port-forward -n tenant-operator-system <pod> 8080:8080 &
   curl -s localhost:8080/metrics | grep tenant_reconcile > reconcile-metrics.txt
   pkill -f "port-forward.*8080:8080"

   # Controller resource usage over time
   kubectl top pods -n tenant-operator-system -l app=tenant-operator --containers > resources-snapshot.txt

   # Sample slow tenant configurations
   kubectl get tenants --all-namespaces -o yaml > tenants-sample.yaml

   # Operator logs with timing info
   kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=5000 > operator-timing.log
   ```

2. Profile reconciliation:
   - Identify slowest operations
   - Measure time in each phase
   - Check for inefficient loops
   - Look for blocking operations

3. If unable to optimize further:
   - Review with platform engineering
   - Consider operator code optimization
   - Open GitHub issue with profiling data
   - Request performance improvements in operator
