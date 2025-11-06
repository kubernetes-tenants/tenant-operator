# Runbook: High Reconciliation Error Rate

## Alert Details

**Alert Name:** `TenantReconciliationErrors`
**Severity:** Warning
**Threshold:** sum(rate(tenant_reconcile_duration_seconds_count{result="error"}[5m])) / sum(rate(tenant_reconcile_duration_seconds_count[5m])) > 0.1 for 10+ minutes

## Description

This alert fires when the tenant operator is experiencing a high reconciliation error rate (>10% of all reconciliations resulting in errors sustained for 10+ minutes). This indicates widespread issues with the reconciliation process, suggesting problems with the controller, API server, or cluster infrastructure rather than isolated tenant issues.

## Symptoms

- Many tenants failing to reconcile successfully
- Operator logs show frequent errors
- Tenants may be stuck in inconsistent states
- Status updates may be delayed or failing
- Performance degradation of operator

## Possible Causes

1. **API Server Issues**
   - API server overloaded or unavailable
   - API server throttling requests
   - Network connectivity problems to API server
   - Certificate or authentication errors

2. **Controller Issues**
   - Bugs in reconciliation logic
   - Panic or crashes in controller code
   - Resource exhaustion (memory, CPU)
   - Deadlocks or goroutine leaks

3. **Cluster Infrastructure**
   - etcd performance degradation
   - Control plane node issues
   - Network partitions or instability
   - DNS resolution failures

4. **RBAC or Permission Issues**
   - Operator service account permissions revoked
   - ClusterRole or RoleBinding modified
   - Cross-namespace permission issues
   - Admission webhooks blocking operations

5. **Resource Constraints**
   - Namespace quotas exceeded across multiple namespaces
   - Cluster resource exhaustion
   - Storage provisioner failures
   - Pod scheduling failures

6. **External Dependencies**
   - Datasource (database) unavailable
   - External APIs timing out
   - DNS failures for external services
   - Network policies blocking traffic

## Diagnosis

### 1. Check Overall Error Rate

```bash
# Check recent reconciliation results
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 \
  | grep -i "reconcil" | grep -c "error"

# Check error vs success ratio
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=1000 \
  | grep -i "reconciliation" | awk '{print $NF}' | sort | uniq -c
```

### 2. Identify Common Error Patterns

```bash
# Get most common error messages
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=2000 \
  | grep -i "error" | awk '{$1=$2=""; print $0}' | sort | uniq -c | sort -rn | head -20

# Check for specific error types
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=1000 \
  | grep -E "timeout|throttl|forbidden|unauthorized|connection refused"
```

### 3. Check Operator Health

```bash
# Check operator pods
kubectl get pods -n tenant-operator-system -l app=tenant-operator

# Check restart count
kubectl get pods -n tenant-operator-system -l app=tenant-operator \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.containerStatuses[0].restartCount}{"\n"}{end}'

# Check resource usage
kubectl top pods -n tenant-operator-system -l app=tenant-operator

# Check for OOMKilled
kubectl get pods -n tenant-operator-system -l app=tenant-operator \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.containerStatuses[*].lastState.terminated.reason}{"\n"}{end}'
```

### 4. Check API Server Health

```bash
# Check API server responsiveness
time kubectl get nodes

# Check API server health endpoints
kubectl get --raw /healthz
kubectl get --raw /readyz

# Check for rate limiting
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 \
  | grep -i "429\|rate limit\|throttl"

# Check API server metrics (if accessible)
kubectl get --raw /metrics | grep apiserver_request_duration_seconds_sum
```

### 5. Check Controller Metrics

```bash
# If metrics endpoint is exposed
OPERATOR_POD=$(kubectl get pods -n tenant-operator-system -l app=tenant-operator -o jsonpath='{.items[0].metadata.name}')

# Port forward to metrics
kubectl port-forward -n tenant-operator-system $OPERATOR_POD 8080:8080 &

# Query metrics
curl -s localhost:8080/metrics | grep tenant_reconcile_duration_seconds
curl -s localhost:8080/metrics | grep workqueue

# Stop port forward
pkill -f "port-forward.*8080:8080"
```

### 6. Check Affected Tenants

```bash
# Count tenants by status
kubectl get tenants --all-namespaces \
  -o jsonpath='{range .items[*]}{.status.conditions[?(@.type=="Ready")].status}{"\n"}{end}' \
  | sort | uniq -c

# List tenants with errors
kubectl get tenants --all-namespaces \
  -o json | jq -r '.items[] | select(.status.conditions[]? | select(.type=="Ready" and .status=="False")) | "\(.metadata.namespace)/\(.metadata.name)"'
```

### 7. Check RBAC Permissions

```bash
# Verify operator can list tenants
kubectl auth can-i list tenants.operator.kubernetes-tenants.org \
  --as=system:serviceaccount:tenant-operator-system:tenant-operator-controller-manager

# Verify operator can update tenant status
kubectl auth can-i update tenants/status \
  --as=system:serviceaccount:tenant-operator-system:tenant-operator-controller-manager

# Check for critical resource permissions
for resource in deployments services configmaps secrets; do
  echo -n "$resource: "
  kubectl auth can-i create $resource \
    --as=system:serviceaccount:tenant-operator-system:tenant-operator-controller-manager
done
```

## Resolution

### For API Server Issues

1. **Check API server logs:**
   ```bash
   # On control plane node
   sudo journalctl -u kube-apiserver --since "10 minutes ago" | tail -100
   ```

2. **If API server is throttling:**
   ```bash
   # Increase operator QPS limits
   kubectl edit deployment -n tenant-operator-system tenant-operator-controller-manager

   # Add or update args:
   # --kube-api-qps=50
   # --kube-api-burst=100
   ```

3. **If API server is overloaded:**
   - Scale API server (if possible)
   - Reduce operator concurrency temporarily
   - Identify other heavy API consumers

### For Controller Issues

1. **Check controller logs for panics:**
   ```bash
   kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 \
     | grep -i "panic\|fatal\|stack trace"
   ```

2. **Restart controller:**
   ```bash
   kubectl delete pod -n tenant-operator-system -l app=tenant-operator

   # Wait for new pod
   kubectl wait --for=condition=ready pod -n tenant-operator-system \
     -l app=tenant-operator --timeout=60s
   ```

3. **Increase controller resources:**
   ```bash
   kubectl edit deployment -n tenant-operator-system tenant-operator-controller-manager

   # Increase resources:
   # resources:
   #   requests:
   #     memory: "512Mi"
   #     cpu: "1000m"
   #   limits:
   #     memory: "2Gi"
   #     cpu: "4000m"
   ```

4. **Reduce concurrency temporarily:**
   ```bash
   kubectl set env deployment/tenant-operator-controller-manager \
     -n tenant-operator-system \
     --containers=manager \
     TENANT_CONCURRENCY=5  # Reduce from default 10
   ```

### For RBAC Issues

1. **Reapply RBAC manifests:**
   ```bash
   kubectl apply -f config/rbac/
   ```

2. **Verify service account:**
   ```bash
   kubectl get serviceaccount -n tenant-operator-system tenant-operator-controller-manager

   # Check if token is valid
   kubectl get secrets -n tenant-operator-system \
     | grep tenant-operator-controller-manager
   ```

### For External Dependency Issues

1. **Check datasource connectivity:**
   ```bash
   # Test from operator pod
   OPERATOR_POD=$(kubectl get pods -n tenant-operator-system \
     -l app=tenant-operator -o jsonpath='{.items[0].metadata.name}')

   # Test database connection
   kubectl exec -n tenant-operator-system $OPERATOR_POD -- \
     sh -c "nc -zv <database-host> 3306"

   # Test DNS resolution
   kubectl exec -n tenant-operator-system $OPERATOR_POD -- \
     nslookup <database-host>
   ```

2. **Check network policies:**
   ```bash
   kubectl get networkpolicies -n tenant-operator-system
   kubectl describe networkpolicy -n tenant-operator-system
   ```

### For Cluster Resource Issues

1. **Check node status:**
   ```bash
   kubectl get nodes
   kubectl describe nodes | grep -A 5 "Conditions:"
   ```

2. **Check cluster capacity:**
   ```bash
   kubectl top nodes
   kubectl describe nodes | grep -A 5 "Allocated resources"
   ```

3. **Check namespace quotas:**
   ```bash
   # Check quotas across namespaces
   kubectl get resourcequotas --all-namespaces
   ```

## Emergency Mitigation

### Reduce Operator Load

```bash
# Scale down to single replica
kubectl scale deployment -n tenant-operator-system \
  tenant-operator-controller-manager --replicas=1

# Reduce concurrency
kubectl set env deployment/tenant-operator-controller-manager \
  -n tenant-operator-system \
  --containers=manager \
  TENANT_CONCURRENCY=3 \
  REGISTRY_CONCURRENCY=1

# Monitor error rate
watch 'kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=100 | grep -c error'
```

### Pause Non-Critical Registries

```bash
# List registries by tenant count
kubectl get tenantregistries --all-namespaces \
  -o custom-columns=NAMESPACE:.metadata.namespace,NAME:.metadata.name,DESIRED:.status.desired

# Temporarily pause large registries
kubectl annotate tenantregistry <registry-name> -n <namespace> \
  tenant.operator.kubernetes-tenants.org/pause=true

# Resume after issues resolved
kubectl annotate tenantregistry <registry-name> -n <namespace> \
  tenant.operator.kubernetes-tenants.org/pause-
```

## Prevention

1. **Proper resource allocation:**
   ```yaml
   resources:
     requests:
       memory: "512Mi"
       cpu: "500m"
     limits:
       memory: "2Gi"
       cpu: "2000m"
   ```

2. **Appropriate concurrency settings:**
   ```bash
   --tenant-concurrency=10
   --registry-concurrency=3
   --template-concurrency=5
   ```

3. **API client configuration:**
   ```bash
   --kube-api-qps=50
   --kube-api-burst=100
   ```

4. **Implement retries with backoff:**
   - Use exponential backoff for retries
   - Implement circuit breakers
   - Add jitter to avoid thundering herd

5. **Monitor operator health:**
   - Set up alerts for high error rates
   - Monitor resource usage
   - Track reconciliation duration
   - Alert on pod restarts

6. **Regular testing:**
   - Load testing in staging
   - Chaos engineering practices
   - Disaster recovery drills

## Monitoring

```promql
# Error rate
sum(rate(tenant_reconcile_duration_seconds_count{result="error"}[5m]))
  / sum(rate(tenant_reconcile_duration_seconds_count[5m]))

# Total errors per minute
rate(tenant_reconcile_duration_seconds_count{result="error"}[1m]) * 60

# Errors by controller
sum(rate(tenant_reconcile_duration_seconds_count{result="error"}[5m])) by (controller)

# Operator pod restarts
rate(kube_pod_container_status_restarts_total{namespace="tenant-operator-system"}[5m])

# API server request duration
histogram_quantile(0.95, rate(apiserver_request_duration_seconds_bucket[5m]))
```

## Related Alerts

- `TenantReconciliationSlow` - May indicate performance issues
- `TenantNotReady` - Many tenants may be not ready
- `TenantDegraded` - Many tenants may be degraded
- Cluster-level alerts for control plane health

## Quick Health Check

```bash
#!/bin/bash
echo "=== Operator Pods ==="
kubectl get pods -n tenant-operator-system -l app=tenant-operator

echo -e "\n=== Recent Errors ==="
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=100 \
  | grep -i error | tail -10

echo -e "\n=== Error Rate (last 100 reconciliations) ==="
TOTAL=$(kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 \
  | grep -c "Reconciling Tenant")
ERRORS=$(kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 \
  | grep "Reconciling Tenant" | grep -c error)
echo "Errors: $ERRORS / $TOTAL = $(awk "BEGIN {printf \"%.1f%%\", $ERRORS/$TOTAL*100}")"

echo -e "\n=== API Server Health ==="
kubectl get --raw /healthz

echo -e "\n=== Operator Resources ==="
kubectl top pods -n tenant-operator-system -l app=tenant-operator
```

## Escalation

If high error rate persists:

1. Collect comprehensive diagnostics:
   ```bash
   # Operator logs
   kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=5000 > operator.log

   # Operator pod describe
   kubectl describe pods -n tenant-operator-system -l app=tenant-operator > operator-pods.txt

   # Operator metrics (if available)
   kubectl port-forward -n tenant-operator-system <pod> 8080:8080 &
   curl -s localhost:8080/metrics > operator-metrics.txt
   pkill -f "port-forward.*8080:8080"

   # API server health
   kubectl get --raw /healthz > apiserver-health.txt
   kubectl get --raw /metrics | grep apiserver > apiserver-metrics.txt

   # Cluster state
   kubectl get nodes > nodes.txt
   kubectl top nodes > nodes-resources.txt
   kubectl get events --all-namespaces --sort-by='.lastTimestamp' | tail -100 > events.txt

   # Sample tenants
   kubectl get tenants --all-namespaces -o yaml > tenants-sample.yaml
   ```

2. Identify timing:
   - When did errors start?
   - Recent operator updates?
   - Recent cluster changes?
   - Correlation with cluster events?

3. If unable to resolve:
   - Escalate to platform engineering
   - Review recent changes (operator, cluster, infrastructure)
   - Check for known issues in operator repository
   - Open GitHub issue with full diagnostics
