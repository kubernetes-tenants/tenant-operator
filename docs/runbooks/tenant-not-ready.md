# Runbook: Tenant Not Ready

## Alert Details

**Alert Name:** `TenantNotReady`
**Severity:** Critical
**Threshold:** tenant_condition_status{type="Ready"} == 0 for 15+ minutes

## Description

This alert fires when a Tenant CR's `Ready` condition is `False` for an extended period. A Tenant is considered Ready only when ALL of its resources are successfully applied and ready. This indicates the tenant is not functioning properly.

## Symptoms

- Tenant `Ready` condition status is `False`
- Applications or services may be unavailable
- Some resources may be missing or not ready
- End users may experience service disruptions

## Possible Causes

1. **Resource Creation Failures**
   - One or more resources failed to create
   - See `TenantResourcesFailed` alert

2. **Resources Not Becoming Ready**
   - Pods stuck in Pending, CrashLoopBackOff, or ImagePullBackOff
   - Deployments not reaching desired replicas
   - Jobs not completing successfully

3. **Template or Configuration Issues**
   - Invalid template rendering
   - Missing required fields
   - Incorrect resource specifications

4. **Infrastructure Issues**
   - Insufficient cluster resources (CPU, memory, storage)
   - Node scheduling problems
   - Network connectivity issues

5. **Dependency Issues**
   - Dependent resources not ready
   - External services unavailable
   - Required secrets or configmaps missing

6. **Slow Reconciliation**
   - Resources taking longer than expected to become ready
   - High cluster load delaying operations
   - Complex dependency chains

## Diagnosis

### 1. Check Tenant Overall Status

```bash
kubectl get tenant <tenant-name> -n <namespace>

# Detailed status with conditions
kubectl get tenant <tenant-name> -n <namespace> -o yaml | grep -A 30 "status:"
```

### 2. Check Resource Readiness

```bash
# Compare ready vs desired
kubectl get tenant <tenant-name> -n <namespace> -o jsonpath='{.status.readyResources}/{.status.desiredResources}'

# Check for failed resources
kubectl get tenant <tenant-name> -n <namespace> -o jsonpath='{.status.failedResources}'
```

### 3. Identify Not-Ready Resources

```bash
# List all resources owned by tenant
kubectl get all -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name>

# Check pod status specifically
kubectl get pods -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name>

# Check other resources
kubectl get deployments,statefulsets,jobs,services,ingresses -n <namespace> \
  -l kubernetes-tenants.org/tenant=<tenant-name>
```

### 4. Inspect Problematic Resources

```bash
# For pods not running
kubectl describe pod <pod-name> -n <namespace>
kubectl logs <pod-name> -n <namespace>

# For deployments not ready
kubectl describe deployment <deployment-name> -n <namespace>
kubectl rollout status deployment/<deployment-name> -n <namespace>

# For jobs not completed
kubectl describe job <job-name> -n <namespace>
```

### 5. Check Tenant Events

```bash
kubectl describe tenant <tenant-name> -n <namespace>

# All events in namespace
kubectl get events -n <namespace> --sort-by='.lastTimestamp' | tail -50
```

### 6. Review Operator Logs

```bash
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=100 \
  | grep <tenant-name>
```

## Resolution

### For Pod Scheduling Issues

1. **Check node resources:**
   ```bash
   kubectl top nodes
   kubectl describe nodes | grep -A 5 "Allocated resources"
   ```

2. **Check pod events:**
   ```bash
   kubectl describe pod <pod-name> -n <namespace>
   # Look for: "FailedScheduling", "Insufficient cpu", "Insufficient memory"
   ```

3. **Solutions:**
   - Scale cluster or add nodes if resources exhausted
   - Reduce resource requests in tenant template
   - Remove node selectors/affinity rules if too restrictive

### For Image Pull Issues

1. **Check pod status:**
   ```bash
   kubectl describe pod <pod-name> -n <namespace>
   # Look for: "ImagePullBackOff", "ErrImagePull"
   ```

2. **Common fixes:**
   ```bash
   # Verify image exists and tag is correct
   docker pull <image-name>

   # Add imagePullSecrets if needed
   kubectl create secret docker-registry regcred \
     --docker-server=<registry> \
     --docker-username=<username> \
     --docker-password=<password> \
     -n <namespace>

   # Update template to use secret
   kubectl edit tenanttemplate <template-name> -n <namespace>
   ```

### For Application Crashes

1. **Check application logs:**
   ```bash
   kubectl logs <pod-name> -n <namespace>
   kubectl logs <pod-name> -n <namespace> --previous  # Previous crash
   ```

2. **Common issues:**
   - Missing environment variables
   - Incorrect configuration
   - Database connection failures
   - Port conflicts

3. **Fix application configuration:**
   ```bash
   # Update ConfigMap or Secret
   kubectl edit configmap <configmap-name> -n <namespace>

   # Or update template
   kubectl edit tenanttemplate <template-name> -n <namespace>
   ```

### For Readiness Probe Failures

1. **Check probe configuration:**
   ```bash
   kubectl get pod <pod-name> -n <namespace> -o yaml | grep -A 10 readinessProbe
   ```

2. **Test probe manually:**
   ```bash
   kubectl exec <pod-name> -n <namespace> -- wget -qO- http://localhost:8080/health
   # Or whatever the readiness probe endpoint is
   ```

3. **Adjust probe settings:**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>
   # Increase initialDelaySeconds, periodSeconds, or failureThreshold
   ```

### For Job Completion Issues

1. **Check job status:**
   ```bash
   kubectl describe job <job-name> -n <namespace>
   kubectl logs job/<job-name> -n <namespace>
   ```

2. **Fix job if needed:**
   ```bash
   # Delete failed job if using CreationPolicy: Once
   kubectl delete job <job-name> -n <namespace>

   # Or update template and recreate tenant
   kubectl edit tenanttemplate <template-name> -n <namespace>
   ```

### For Dependency Chain Issues

1. **Check resource dependencies:**
   ```bash
   kubectl get tenant <tenant-name> -n <namespace> -o yaml | grep -B 5 dependIds
   ```

2. **Verify dependent resources are ready:**
   ```bash
   # Check each resource in dependency chain
   kubectl get <resource-kind> <resource-name> -n <namespace>
   ```

3. **Fix dependency issues:**
   - Ensure dependent resources exist
   - Remove unnecessary dependencies
   - Adjust timeouts if resources are slow to start

### Force Reconciliation

```bash
# Trigger immediate reconciliation
kubectl annotate tenant <tenant-name> -n <namespace> \
  tenant.operator.kubernetes-tenants.org/reconcile="$(date +%s)" --overwrite

# Watch tenant status
kubectl get tenant <tenant-name> -n <namespace> -w
```

## Quick Health Check Commands

```bash
# One-liner to check tenant health
kubectl get tenant <tenant-name> -n <namespace> \
  -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'

# Check all pods status
kubectl get pods -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name> \
  --no-headers | awk '{print $3}' | sort | uniq -c

# Count ready vs total resources
echo "Ready: $(kubectl get tenant <tenant-name> -n <namespace> -o jsonpath='{.status.readyResources}')"
echo "Desired: $(kubectl get tenant <tenant-name> -n <namespace> -o jsonpath='{.status.desiredResources}')"
echo "Failed: $(kubectl get tenant <tenant-name> -n <namespace> -o jsonpath='{.status.failedResources}')"
```

## Prevention

1. **Set realistic resource requests:**
   - Profile application resource usage
   - Set appropriate CPU and memory requests
   - Don't over-request resources

2. **Configure appropriate timeouts:**
   - Set `timeoutSeconds` based on actual startup time
   - Allow extra time for slow-starting applications
   - Use `waitForReady: false` for optional resources

3. **Use health checks properly:**
   - Configure readiness probes correctly
   - Set appropriate `initialDelaySeconds`
   - Ensure health endpoints are reliable

4. **Test templates thoroughly:**
   - Validate in development environment
   - Test with realistic configurations
   - Verify all dependencies are included

5. **Monitor resource capacity:**
   - Track cluster resource utilization
   - Ensure sufficient capacity for tenant count
   - Plan for growth and spikes

## Metrics to Monitor

```promql
# Tenants not ready
tenant_condition_status{type="Ready"} == 0

# Resource readiness ratio
tenant_resources_ready / tenant_resources_desired

# Time tenant has been not ready
time() - tenant_condition_status{type="Ready"} < 1

# Pod restart rate
rate(kube_pod_container_status_restarts_total[5m]) > 0
```

## Related Alerts

- `TenantResourcesFailed` - Check for failed resources first
- `TenantDegraded` - May indicate root cause
- `TenantResourcesMismatch` - Partial readiness issue
- `TenantReconciliationSlow` - May be contributing factor

## Escalation

If tenant remains not ready after following this runbook:

1. Collect comprehensive diagnostics:
   ```bash
   # Create diagnostic bundle
   mkdir -p tenant-diagnostics

   # Tenant details
   kubectl get tenant <tenant-name> -n <namespace> -o yaml > tenant-diagnostics/tenant.yaml
   kubectl describe tenant <tenant-name> -n <namespace> > tenant-diagnostics/tenant-describe.txt

   # All resources
   kubectl get all,ingress,configmap,secret,pvc -n <namespace> \
     -l kubernetes-tenants.org/tenant=<tenant-name> -o yaml > tenant-diagnostics/resources.yaml

   # Pod logs
   for pod in $(kubectl get pods -n <namespace> -l kubernetes-tenants.org/tenant=<tenant-name> -o name); do
     kubectl logs -n <namespace> $pod > tenant-diagnostics/${pod}.log 2>&1
   done

   # Events
   kubectl get events -n <namespace> --sort-by='.lastTimestamp' > tenant-diagnostics/events.txt

   # Operator logs
   kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=1000 > tenant-diagnostics/operator.log

   # Create tarball
   tar czf tenant-diagnostics-$(date +%Y%m%d-%H%M%S).tar.gz tenant-diagnostics/
   ```

2. Review GitHub issues for similar cases
3. Contact platform team or open issue with diagnostic bundle
