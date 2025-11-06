# Runbook: Tenant Status Unknown

## Alert Details

**Alert Name:** `TenantStatusUnknown`
**Severity:** Critical
**Threshold:** tenant_condition_status{type="Ready"} == 2 for 10+ minutes

## Description

This alert fires when a Tenant CR's `Ready` condition is in `Unknown` state for an extended period. An Unknown status typically indicates a problem with the operator controller itself, API server communication issues, or cluster-wide problems rather than tenant-specific issues.

## Symptoms

- Tenant `Ready` condition status is `Unknown`
- Tenant status may not be updating
- LastTransitionTime for conditions is stale
- Operator may not be reconciling the tenant

## Possible Causes

1. **Controller Issues**
   - Operator pod crashed or restarting
   - Operator pod not running
   - Controller goroutines stuck or panicking
   - Resource exhaustion in operator pod

2. **API Server Communication Problems**
   - Network connectivity issues to API server
   - API server overloaded or unavailable
   - Client-side rate limiting or throttling
   - Certificate or authentication issues

3. **Etcd or Control Plane Issues**
   - Etcd cluster unhealthy
   - Control plane nodes under pressure
   - Distributed consensus problems

4. **Namespace or RBAC Issues**
   - Namespace in terminating state
   - Operator service account deleted or misconfigured
   - RBAC permissions revoked

5. **Resource Watch Failures**
   - Watch connection broken
   - Cache not syncing properly
   - Informer errors

## Diagnosis

### 1. Check Operator Pod Health

```bash
# Check operator pods are running
kubectl get pods -n tenant-operator-system -l app=tenant-operator

# Check pod readiness and restart count
kubectl get pods -n tenant-operator-system -l app=tenant-operator -o wide

# Check for recent restarts
kubectl get pods -n tenant-operator-system -l app=tenant-operator \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.containerStatuses[0].restartCount}{"\n"}{end}'
```

### 2. Review Operator Logs

```bash
# Check recent logs
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=200

# Look for errors
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=500 | grep -i "error\|panic\|fatal"

# Check if controller is reconciling
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=100 | grep "Reconciling Tenant"
```

### 3. Check Operator Pod Resources

```bash
# Check resource usage
kubectl top pods -n tenant-operator-system -l app=tenant-operator

# Check resource limits
kubectl get pods -n tenant-operator-system -l app=tenant-operator -o yaml | grep -A 5 resources:

# Check for OOMKilled
kubectl get pods -n tenant-operator-system -l app=tenant-operator -o yaml | grep -i oomkilled
```

### 4. Check API Server Health

```bash
# Check API server responsiveness
kubectl get --raw /healthz

# Check API server metrics
kubectl get --raw /metrics | grep apiserver_request_duration_seconds

# Check for rate limiting
kubectl get --raw /metrics | grep rest_client_requests_total
```

### 5. Check Control Plane Health

```bash
# Check control plane components
kubectl get componentstatuses  # Deprecated but still useful

# Check control plane pods
kubectl get pods -n kube-system

# Check etcd health (if accessible)
kubectl exec -n kube-system etcd-<node> -- etcdctl endpoint health
```

### 6. Check Tenant-Specific Status

```bash
# Check tenant status
kubectl get tenant <tenant-name> -n <namespace> -o yaml | grep -A 20 "status:"

# Check last reconciliation time
kubectl get tenant <tenant-name> -n <namespace> \
  -o jsonpath='{.status.conditions[?(@.type=="Ready")].lastTransitionTime}'

# Check for finalizers
kubectl get tenant <tenant-name> -n <namespace> -o jsonpath='{.metadata.finalizers}'
```

### 7. Check Namespace Health

```bash
# Check namespace status
kubectl get namespace <namespace> -o yaml

# Check if namespace is terminating
kubectl get namespace <namespace> -o jsonpath='{.status.phase}'
```

## Resolution

### For Operator Pod Issues

1. **Restart operator pod:**
   ```bash
   kubectl delete pod -n tenant-operator-system -l app=tenant-operator

   # Wait for new pod to be ready
   kubectl wait --for=condition=ready pod -n tenant-operator-system -l app=tenant-operator --timeout=60s
   ```

2. **Check if restart helped:**
   ```bash
   # Check tenant status after restart
   kubectl get tenant <tenant-name> -n <namespace> -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'
   ```

3. **If pod is CrashLooping:**
   ```bash
   # Check crash logs
   kubectl logs -n tenant-operator-system -l app=tenant-operator --previous

   # Check pod events
   kubectl describe pod -n tenant-operator-system -l app=tenant-operator
   ```

### For Resource Exhaustion

1. **Increase operator resources:**
   ```bash
   kubectl edit deployment -n tenant-operator-system tenant-operator-controller-manager

   # Increase resources:
   # resources:
   #   requests:
   #     memory: "128Mi"
   #     cpu: "500m"
   #   limits:
   #     memory: "512Mi"
   #     cpu: "2000m"
   ```

2. **Monitor resource usage:**
   ```bash
   kubectl top pods -n tenant-operator-system -l app=tenant-operator --containers
   ```

### For API Server Communication Issues

1. **Check network connectivity:**
   ```bash
   # From operator pod
   kubectl exec -n tenant-operator-system <operator-pod> -- wget -qO- https://kubernetes.default.svc/healthz
   ```

2. **Check for rate limiting:**
   ```bash
   # Review operator logs for 429 responses
   kubectl logs -n tenant-operator-system -l app=tenant-operator | grep "429"

   # Check QPS and burst settings in operator deployment
   kubectl get deployment -n tenant-operator-system tenant-operator-controller-manager -o yaml | grep -i qps
   ```

3. **Increase client rate limits if needed:**
   ```bash
   # Add/update controller flags
   kubectl edit deployment -n tenant-operator-system tenant-operator-controller-manager
   # Add flags: --kube-api-qps=50 --kube-api-burst=100
   ```

### For RBAC Issues

1. **Verify service account exists:**
   ```bash
   kubectl get serviceaccount -n tenant-operator-system tenant-operator-controller-manager
   ```

2. **Check RBAC permissions:**
   ```bash
   # Check if operator can list tenants
   kubectl auth can-i list tenants.operator.kubernetes-tenants.org \
     --as=system:serviceaccount:tenant-operator-system:tenant-operator-controller-manager

   # Check if operator can update tenant status
   kubectl auth can-i update tenants/status \
     --as=system:serviceaccount:tenant-operator-system:tenant-operator-controller-manager
   ```

3. **Restore RBAC if needed:**
   ```bash
   # Reapply operator RBAC
   kubectl apply -f config/rbac/
   ```

### For Namespace Terminating Issues

1. **Check namespace status:**
   ```bash
   kubectl get namespace <namespace> -o yaml
   ```

2. **If namespace is stuck terminating:**
   ```bash
   # This is a cluster-wide issue, not operator-specific
   # Check for finalizers blocking deletion
   kubectl get namespace <namespace> -o json | jq '.spec.finalizers'

   # Escalate to cluster admin
   ```

### Force Status Update

```bash
# Delete and recreate tenant to force re-reconciliation
# WARNING: Only if absolutely necessary and you understand the impact

# Backup first
kubectl get tenant <tenant-name> -n <namespace> -o yaml > tenant-backup.yaml

# Remove finalizer if stuck
kubectl patch tenant <tenant-name> -n <namespace> \
  -p '{"metadata":{"finalizers":[]}}' --type=merge

# Delete
kubectl delete tenant <tenant-name> -n <namespace> --wait=false

# Recreate from registry sync
# (Registry controller will recreate it on next sync)
```

## Quick Health Check

```bash
# Check operator and tenant status in one command
echo "=== Operator Pods ==="
kubectl get pods -n tenant-operator-system -l app=tenant-operator

echo -e "\n=== Operator Recent Logs ==="
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=20 | grep -i "error\|reconciling"

echo -e "\n=== Tenant Status ==="
kubectl get tenant <tenant-name> -n <namespace> -o jsonpath='Ready: {.status.conditions[?(@.type=="Ready")].status}{"\n"}Last Update: {.status.conditions[?(@.type=="Ready")].lastTransitionTime}{"\n"}'

echo -e "\n=== API Server Health ==="
kubectl get --raw /healthz
```

## Prevention

1. **Monitor operator health:**
   - Set up alerts for operator pod restarts
   - Monitor operator resource usage
   - Alert on operator reconciliation errors

2. **Resource sizing:**
   - Allocate sufficient resources to operator
   - Monitor and adjust based on tenant count
   - Use HPA if operator load varies significantly

3. **API rate limiting:**
   - Configure appropriate QPS and burst settings
   - Monitor for rate limit errors
   - Adjust limits based on cluster size

4. **High availability:**
   - Run multiple operator replicas if supported
   - Use leader election
   - Ensure proper pod disruption budgets

5. **Regular health checks:**
   - Automated operator health monitoring
   - Synthetic tests for tenant reconciliation
   - Alert on stale tenant status updates

## Metrics to Monitor

```promql
# Operator pod status
up{job="tenant-operator-metrics"}

# Operator restarts
rate(kube_pod_container_status_restarts_total{namespace="tenant-operator-system"}[5m])

# Unknown condition status
tenant_condition_status{type="Ready"} == 2

# Reconciliation errors
rate(tenant_reconcile_duration_seconds_count{result="error"}[5m])

# API client rate limiting
rest_client_requests_total{code="429"}

# Workqueue depth
workqueue_depth{name="tenant"}
```

## Related Alerts

- `TenantNotReady` - May also fire if status can't be determined
- `TenantDegraded` - Previous state before going Unknown
- Cluster-level alerts for control plane health

## Escalation

This is a critical alert that often indicates cluster or operator-level issues:

1. **Immediate actions:**
   - Check operator pod health
   - Restart operator if necessary
   - Verify API server and control plane health

2. **If issue persists:**
   - Escalate to cluster administrators
   - Check for cluster-wide issues (etcd, networking, control plane)
   - Collect operator diagnostics:
     ```bash
     kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=1000 > operator-full.log
     kubectl get events -n tenant-operator-system --sort-by='.lastTimestamp' > operator-events.txt
     kubectl describe pod -n tenant-operator-system -l app=tenant-operator > operator-pods.txt
     ```

3. **Open GitHub issue** with:
   - Operator version
   - Kubernetes version
   - Operator logs and diagnostics
   - Description of cluster state when issue occurred
