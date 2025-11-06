# Runbook: Tenant Resources Mismatch

## Alert Details

**Alert Name:** `TenantResourcesMismatch`
**Severity:** Warning
**Threshold:** tenant_resources_ready != tenant_resources_desired AND tenant_resources_desired > 0 AND tenant_resources_failed == 0 for 15+ minutes

## Description

This alert fires when a tenant's ready resource count doesn't match the desired count, but there are no failed resources. This indicates that reconciliation is progressing slowly, resources are stuck in pending state, or there's a synchronization issue - but not an outright failure.

## Symptoms

- Tenant status shows ready count doesn't match desired count
- No failed resources reported
- Tenant may still be partially functional
- Some resources may be in pending or creating state

## Possible Causes

1. **Slow Resource Provisioning**
   - Pods taking long time to become ready
   - Persistent volumes taking time to provision
   - External dependencies slow to respond
   - Image pulls taking extended time

2. **Resource Waiting on Dependencies**
   - Resources waiting for `dependIds` to be ready
   - Circular or long dependency chains
   - Dependent resources slow to become ready

3. **Partial Resource Readiness**
   - Deployments with some replicas ready
   - StatefulSets progressing through pods sequentially
   - Jobs in progress but not completed

4. **Synchronization Issues**
   - Operator cache not synchronized
   - Watch events delayed or missed
   - Status updates not propagating

5. **Resource Conditions Not Met**
   - Readiness probes not passing yet
   - Resources created but `waitForReady=true` still pending
   - Timeout not reached but resources not ready

## Diagnosis

### 1. Check Resource Counts

```bash
# Get current vs desired counts
kubectl get tenant <tenant-name> -n <namespace> \
  -o jsonpath='Ready: {.status.readyResources}/{.status.desiredResources}{"\n"}Failed: {.status.failedResources}{"\n"}'

# Check conditions
kubectl get tenant <tenant-name> -n <namespace> \
  -o jsonpath='{.status.conditions[*].type}{"\n"}{.status.conditions[*].status}'
```

### 2. Identify Non-Ready Resources

```bash
# List all resources owned by tenant
kubectl get all -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name>

# Check pod readiness specifically
kubectl get pods -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name> \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.phase}{"\t"}{.status.conditions[?(@.type=="Ready")].status}{"\n"}{end}'

# Check deployments
kubectl get deployments -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name> \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.readyReplicas}/{.spec.replicas}{"\n"}{end}'

# Check statefulsets
kubectl get statefulsets -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name> \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.readyReplicas}/{.spec.replicas}{"\n"}{end}'

# Check jobs
kubectl get jobs -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name> \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.succeeded}/{.spec.completions}{"\n"}{end}'
```

### 3. Check Resource Events

```bash
# Get all events for tenant resources
kubectl get events -n <namespace> --sort-by='.lastTimestamp' \
  | grep -i "$(kubectl get all -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name> -o name | cut -d/ -f2 | head -1)"

# Or describe specific non-ready resources
kubectl describe pod <pod-name> -n <namespace>
kubectl describe deployment <deployment-name> -n <namespace>
```

### 4. Check for Pending Pods

```bash
# Find pods not running
kubectl get pods -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name> \
  --field-selector=status.phase!=Running

# Describe pending pods
for pod in $(kubectl get pods -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name> \
  --field-selector=status.phase=Pending -o name); do
  echo "=== $pod ==="
  kubectl describe $pod -n <namespace> | grep -A 10 "Events:"
done
```

### 5. Check Dependency Chain

```bash
# Check if resources are waiting on dependencies
kubectl get tenant <tenant-name> -n <namespace> -o yaml | grep -A 5 dependIds

# Check operator logs for dependency waits
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=200 \
  | grep -i "waiting\|depend"
```

### 6. Check Resource Timeouts

```bash
# Check timeout settings
kubectl get tenant <tenant-name> -n <namespace> \
  -o jsonpath='{.spec.resources[*].timeoutSeconds}' | tr ' ' '\n' | sort | uniq

# Check how long resources have been pending
kubectl get pods -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name> \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.creationTimestamp}{"\n"}{end}'
```

## Resolution

### For Slow Resource Provisioning

1. **Check what's taking time:**
   ```bash
   # Check pod initialization
   kubectl describe pod <pod-name> -n <namespace>

   # Check container status
   kubectl get pod <pod-name> -n <namespace> \
     -o jsonpath='{.status.containerStatuses[*].state}'
   ```

2. **Common slow provisioning issues:**

   **Image pulls:**
   ```bash
   # Check image pull status
   kubectl get pod <pod-name> -n <namespace> \
     -o jsonpath='{.status.containerStatuses[*].state.waiting.reason}'

   # If pulling large images, wait or optimize:
   # - Use smaller base images
   # - Pre-pull images on nodes
   # - Use local registry
   ```

   **Storage provisioning:**
   ```bash
   # Check PVC status
   kubectl get pvc -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name>

   # Check storage class provisioner
   kubectl get storageclass
   kubectl describe pvc <pvc-name> -n <namespace>
   ```

3. **If genuinely slow but progressing:**
   - Wait for resources to become ready
   - Alert will auto-resolve once resources are ready
   - Consider increasing timeouts if consistent

### For Readiness Probe Issues

1. **Check probe configuration:**
   ```bash
   kubectl get pod <pod-name> -n <namespace> -o yaml | grep -A 15 readinessProbe
   ```

2. **Test probe manually:**
   ```bash
   # HTTP probe
   kubectl exec <pod-name> -n <namespace> -- wget -qO- http://localhost:8080/health

   # Command probe
   kubectl exec <pod-name> -n <namespace> -- /bin/sh -c '<probe-command>'
   ```

3. **Adjust if needed:**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>
   # Increase initialDelaySeconds, periodSeconds, or failureThreshold
   ```

### For Dependency Issues

1. **Check dependency readiness:**
   ```bash
   # Get resources with dependencies
   kubectl get tenant <tenant-name> -n <namespace> -o yaml \
     | grep -B 3 -A 2 dependIds
   ```

2. **Verify dependent resources are ready:**
   ```bash
   # Check each dependent resource
   kubectl get <resource-kind> <resource-name> -n <namespace>
   ```

3. **Fix dependency chain if stuck:**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>
   # Options:
   # - Remove unnecessary dependencies
   # - Increase timeoutSeconds
   # - Set waitForReady: false for non-critical resources
   ```

### For StatefulSet Rollout

```bash
# Check statefulset rollout progress
kubectl rollout status statefulset/<statefulset-name> -n <namespace>

# StatefulSets provision pods sequentially
# If one pod is slow, it blocks others
kubectl get pods -n <namespace> -l app=<statefulset-name> --sort-by=.metadata.name

# Check individual pod issues
kubectl describe pod <statefulset-name>-0 -n <namespace>
```

### For Job Completion

```bash
# Check job status
kubectl get jobs -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name>

# Check job pod logs
kubectl logs job/<job-name> -n <namespace>

# If job is hanging:
# - Check if it's a long-running initialization job
# - Verify job has correct completion criteria
# - Consider setting waitForReady: false if non-critical
```

### Force Status Update

```bash
# Trigger reconciliation
kubectl annotate tenant <tenant-name> -n <namespace> \
  tenant.operator.kubernetes-tenants.org/reconcile="$(date +%s)" --overwrite

# Watch status changes
kubectl get tenant <tenant-name> -n <namespace> -w
```

### Temporary Workaround

If specific resource is perpetually slow but not critical:

```bash
kubectl edit tenanttemplate <template-name> -n <namespace>

# Set waitForReady: false for slow resource:
# spec:
#   jobs:
#     - id: slow-init-job
#       waitForReady: false  # Don't block reconciliation
#       # ... rest of spec
```

## Prevention

1. **Set realistic timeouts:**
   ```yaml
   # Account for actual resource startup time
   spec:
     deployments:
       - id: my-app
         timeoutSeconds: 600  # 10 minutes for slow-starting app
         waitForReady: true
   ```

2. **Optimize resource startup:**
   - Use smaller container images
   - Implement efficient readiness probes
   - Pre-pull images on nodes
   - Use fast storage classes

3. **Use appropriate waitForReady settings:**
   ```yaml
   # For init jobs or non-critical resources
   - id: optional-job
     waitForReady: false

   # For critical path resources
   - id: database
     waitForReady: true
     timeoutSeconds: 900  # 15 minutes
   ```

4. **Optimize dependency chains:**
   - Keep dependency chains short
   - Only define necessary dependencies
   - Parallelize where possible

5. **Configure readiness probes properly:**
   ```yaml
   readinessProbe:
     httpGet:
       path: /health
       port: 8080
     initialDelaySeconds: 30  # Allow time for app startup
     periodSeconds: 10
     failureThreshold: 3
     timeoutSeconds: 5
   ```

## Monitoring

```promql
# Resource count mismatch
(tenant_resources_ready - tenant_resources_desired) != 0

# Percentage ready
(tenant_resources_ready / tenant_resources_desired) * 100 < 100

# Time in mismatched state
time() - tenant_condition_status{type="Ready"} < 1

# Pod not ready duration
kube_pod_status_ready{condition="false"} * on(pod) group_left(label_kubernetes_tenants_org_tenant)
  kube_pod_labels{label_kubernetes_tenants_org_tenant="<tenant-name>"}
```

## Differentiation from Other Alerts

| Alert | Condition | Meaning |
|-------|-----------|---------|
| TenantResourcesFailed | `failed > 0` | Hard failures - resources can't be created |
| TenantResourcesMismatch | `ready != desired AND failed == 0` | Resources pending/slow - no hard failures |
| TenantNotReady | `Ready condition == False` | Overall tenant not ready (could be either) |

## Related Alerts

- `TenantNotReady` - Will fire if mismatch persists
- `TenantReconciliationSlow` - May indicate slow reconciliation
- `TenantResourcesFailed` - Check if this fires instead

## Quick Check Script

```bash
#!/bin/bash
TENANT=$1
NAMESPACE=$2

echo "=== Tenant Status ==="
kubectl get tenant $TENANT -n $NAMESPACE \
  -o jsonpath='Ready: {.status.readyResources}/{.status.desiredResources}{"\n"}Failed: {.status.failedResources}{"\n"}'

echo -e "\n=== Pod Status ==="
kubectl get pods -n $NAMESPACE -l kubernetes-tenants.org/tenant=$TENANT

echo -e "\n=== Non-Running Pods ==="
kubectl get pods -n $NAMESPACE -l kubernetes-tenants.org/tenant=$TENANT \
  --field-selector=status.phase!=Running

echo -e "\n=== Deployment Status ==="
kubectl get deployments -n $NAMESPACE -l kubernetes-tenants.org/tenant=$TENANT \
  -o custom-columns=NAME:.metadata.name,READY:.status.readyReplicas,DESIRED:.spec.replicas

echo -e "\n=== Recent Events ==="
kubectl get events -n $NAMESPACE --sort-by='.lastTimestamp' | tail -10
```

## Escalation

If resources remain in mismatch state despite troubleshooting:

1. Collect diagnostics:
   ```bash
   kubectl get tenant <tenant-name> -n <namespace> -o yaml > tenant.yaml
   kubectl get all -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name> -o yaml > resources.yaml
   kubectl describe pod -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name> > pods-describe.txt
   kubectl get events -n <namespace> --sort-by='.lastTimestamp' > events.txt
   ```

2. Check for cluster-wide issues:
   - Node resource capacity
   - Storage provisioner health
   - Image registry availability

3. If issue persists:
   - Escalate to platform team
   - Check if it's a known issue
   - Open GitHub issue with diagnostics if necessary
