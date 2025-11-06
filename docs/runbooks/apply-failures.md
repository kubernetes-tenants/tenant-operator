# Runbook: High Apply Failure Rate

## Alert Details

**Alert Name:** `HighApplyFailureRate`
**Severity:** Warning
**Threshold:** sum(rate(apply_attempts_total{result="failure"}[5m])) by (kind) / sum(rate(apply_attempts_total[5m])) by (kind) > 0.2 for 10+ minutes

## Description

This alert fires when a specific resource kind is experiencing a high apply failure rate (>20% of apply attempts failing sustained for 10+ minutes). This indicates systemic issues with applying resources of that particular kind, suggesting RBAC issues, API server problems, or invalid resource specifications.

## Symptoms

- High failure rate for specific resource kind (Deployment, Service, etc.)
- Resources of specific kind not being created/updated
- Events showing repeated apply failures
- Tenants failing with consistent pattern for one resource type

## Possible Causes

1. **RBAC Permission Issues**
   - Operator lacks permissions for specific resource kind
   - Cross-namespace permission issues
   - ClusterRole or RoleBinding misconfigured
   - ServiceAccount permissions revoked

2. **Resource Validation Failures**
   - Invalid resource specifications in templates
   - API schema violations
   - Required fields missing
   - Type mismatches in resource specs

3. **Admission Webhook Rejections**
   - ValidatingWebhook rejecting resources
   - MutatingWebhook causing conflicts
   - Policy violations (OPA, Kyverno, etc.)
   - Custom admission controllers blocking

4. **API Server Issues**
   - API server rejecting specific resource kind
   - CRD not installed or misconfigured
   - API version deprecation or removal
   - Resource type disabled in API server

5. **Quota or Limit Issues**
   - Namespace resource quotas for specific kind
   - Cluster-wide limits on resource count
   - Object count limits per namespace
   - API server object limits

6. **Template Issues**
   - Invalid template rendering for specific resource kind
   - Missing required template variables
   - Type conversion errors in templates
   - Malformed YAML in spec

## Diagnosis

### 1. Identify Failing Resource Kind

```bash
# Check apply failure rate by kind
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=2000 \
  | grep -i "apply failed\|apply error" \
  | awk '{for(i=1;i<=NF;i++) if($i~/kind=/) print $i}' \
  | sort | uniq -c | sort -rn

# Get specific error messages for failing kind
FAILING_KIND="Deployment"  # Replace with actual failing kind
kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=1000 \
  | grep "kind=$FAILING_KIND" | grep -i "error\|failed" | head -10
```

### 2. Check RBAC Permissions

```bash
# Check if operator can create the failing resource kind
RESOURCE_KIND="deployments"  # Use lowercase plural
kubectl auth can-i create $RESOURCE_KIND \
  --as=system:serviceaccount:tenant-operator-system:tenant-operator-controller-manager

# Check all verbs for this resource
for verb in get list watch create update patch delete; do
  echo -n "$verb $RESOURCE_KIND: "
  kubectl auth can-i $verb $RESOURCE_KIND \
    --as=system:serviceaccount:tenant-operator-system:tenant-operator-controller-manager
done

# Check cross-namespace permissions
kubectl auth can-i create $RESOURCE_KIND \
  --as=system:serviceaccount:tenant-operator-system:tenant-operator-controller-manager \
  --namespace=<target-namespace>
```

### 3. Check Resource Specifications

```bash
# Get sample failed resource spec from tenant
kubectl get tenants --all-namespaces -o json \
  | jq -r ".items[] | select(.status.failedResources > 0) | .spec.${FAILING_KIND}s[0]" \
  | head -1 > /tmp/failed-resource.yaml

# Validate against API server (dry-run)
kubectl create -f /tmp/failed-resource.yaml --dry-run=server
```

### 4. Check for Admission Webhooks

```bash
# List validating webhooks
kubectl get validatingwebhookconfigurations

# List mutating webhooks
kubectl get mutatingwebhookconfigurations

# Check if any webhook is blocking the resource kind
kubectl get validatingwebhookconfigurations -o json \
  | jq -r '.items[] | select(.webhooks[].rules[].resources[]? == "'$RESOURCE_KIND'"s") | .metadata.name'

# Check webhook logs if available
kubectl logs -n <webhook-namespace> -l <webhook-selector> --tail=200 \
  | grep -i "$RESOURCE_KIND\|denied\|rejected"
```

### 5. Check Quota and Limits

```bash
# Check resource quotas
kubectl get resourcequotas --all-namespaces \
  -o json | jq -r '.items[] | select(.spec.hard | keys[] | contains("'$RESOURCE_KIND'")) | "\(.metadata.namespace) \(.metadata.name)"'

# Check specific namespace quota
kubectl describe resourcequota -n <namespace>

# Check API server limits
kubectl get --raw /api/v1 | jq -r '.resources[] | select(.name=="'$RESOURCE_KIND's") | .namespaced'
```

### 6. Analyze Template for Failing Resource

```bash
# Get template that's generating failing resources
TEMPLATE_NAME="<template-name>"
kubectl get tenanttemplate $TEMPLATE_NAME -n <namespace> \
  -o yaml | grep -A 50 "${FAILING_KIND}s:"

# Check for template syntax errors
kubectl get tenanttemplate $TEMPLATE_NAME -n <namespace> \
  -o jsonpath="{.spec.${FAILING_KIND}s[*].spec}" | jq .
```

### 7. Check API Server Events

```bash
# Get events related to the failing resource kind
kubectl get events --all-namespaces --sort-by='.lastTimestamp' \
  | grep -i "$FAILING_KIND" | tail -20

# Check for API server rejections
kubectl get events --all-namespaces \
  -o json | jq -r '.items[] | select(.reason=="FailedCreate" or .reason=="FailedUpdate") | select(.involvedObject.kind=="'$FAILING_KIND'") | "\(.lastTimestamp) \(.message)"' \
  | tail -10
```

## Resolution

### For RBAC Permission Issues

1. **Verify ClusterRole has required permissions:**
   ```bash
   kubectl get clusterrole tenant-operator-manager-role -o yaml
   ```

2. **Add missing permissions:**
   ```bash
   kubectl edit clusterrole tenant-operator-manager-role

   # Add rule for failing resource kind:
   - apiGroups:
       - "apps"  # or appropriate API group
     resources:
       - deployments  # failing resource kind
     verbs:
       - get
       - list
       - watch
       - create
       - update
       - patch
       - delete
   ```

3. **For cross-namespace resources:**
   ```bash
   # Ensure ClusterRole (not Role) is used
   kubectl get clusterrolebinding tenant-operator-manager-rolebinding -o yaml

   # Verify it binds ClusterRole to service account
   ```

4. **Reapply RBAC if needed:**
   ```bash
   kubectl apply -f config/rbac/
   ```

### For Resource Validation Failures

1. **Fix template specification:**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>

   # Common fixes:
   # - Add required fields (e.g., spec.selector for Deployments)
   # - Fix field types (e.g., replicas should be integer not string)
   # - Correct API version
   # - Fix nested object structures
   ```

2. **Validate template changes:**
   ```bash
   # Extract resource spec
   kubectl get tenanttemplate <template-name> -n <namespace> \
     -o jsonpath='{.spec.deployments[0].spec}' > /tmp/test-resource.yaml

   # Test with dry-run
   kubectl create -f /tmp/test-resource.yaml --dry-run=server
   ```

3. **Add default values for optional fields:**
   ```yaml
   spec:
     deployments:
       - id: app
         spec:
           replicas: {{ .replicas | default 1 }}  # Provide default
           selector:
             matchLabels:
               app: "{{ .uid }}-app"
           template:
             metadata:
               labels:
                 app: "{{ .uid }}-app"  # Must match selector
   ```

### For Admission Webhook Rejections

1. **Identify blocking webhook:**
   ```bash
   # Check webhook configurations
   kubectl get validatingwebhookconfigurations -o yaml \
     | grep -A 10 "$RESOURCE_KIND"
   ```

2. **Review webhook logs:**
   ```bash
   # Find webhook pod
   kubectl get pods -A -l <webhook-selector>

   # Check logs
   kubectl logs -n <namespace> <webhook-pod> --tail=200 \
     | grep -i "denied\|rejected"
   ```

3. **Options to resolve:**

   **Option A: Fix resource to comply with policy**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>
   # Adjust spec to meet policy requirements
   ```

   **Option B: Exclude operator from webhook (if appropriate)**
   ```bash
   kubectl edit validatingwebhookconfiguration <webhook-name>

   # Add namespace selector to exclude operator:
   namespaceSelector:
     matchExpressions:
       - key: kubernetes.io/metadata.name
         operator: NotIn
         values:
           - tenant-operator-system
   ```

   **Option C: Temporarily disable webhook (emergency only)**
   ```bash
   kubectl delete validatingwebhookconfiguration <webhook-name>
   # Remember to restore after fixing root cause
   ```

### For API Version Issues

1. **Check deprecated API versions:**
   ```bash
   # Check API version in template
   kubectl get tenanttemplate <template-name> -n <namespace> \
     -o jsonpath='{.spec.deployments[0].spec.apiVersion}'

   # Check available API versions
   kubectl api-resources | grep -i $RESOURCE_KIND
   ```

2. **Update to current API version:**
   ```bash
   kubectl edit tenanttemplate <template-name> -n <namespace>

   # Examples:
   # Old: extensions/v1beta1 -> New: apps/v1 (Deployments)
   # Old: networking.k8s.io/v1beta1 -> New: networking.k8s.io/v1 (Ingresses)
   ```

### For Quota Issues

1. **Increase quota:**
   ```bash
   kubectl edit resourcequota <quota-name> -n <namespace>

   # Increase limit for resource kind:
   # spec:
   #   hard:
   #     deployments.apps: "100"  # Increase from current value
   ```

2. **Or remove quota temporarily:**
   ```bash
   kubectl delete resourcequota <quota-name> -n <namespace>
   # Remember to restore with higher limits
   ```

## Bulk Fix for Affected Tenants

After fixing root cause:

```bash
# Force reconciliation of all tenants with failed resources
kubectl get tenants --all-namespaces -o json \
  | jq -r '.items[] | select(.status.failedResources > 0) | "\(.metadata.namespace) \(.metadata.name)"' \
  | while read ns name; do
      echo "Reconciling $ns/$name"
      kubectl annotate tenant $name -n $ns \
        tenant.operator.kubernetes-tenants.org/reconcile="$(date +%s)" --overwrite
      sleep 1
    done

# Monitor recovery
watch 'kubectl get tenants --all-namespaces -o json | jq -r ".items[] | select(.status.failedResources > 0) | \"\(.metadata.namespace)/\(.metadata.name)\"" | wc -l'
```

## Prevention

1. **Validate templates thoroughly:**
   - Test all resource kinds in templates
   - Use `kubectl create --dry-run=server` to validate
   - Check API version compatibility
   - Ensure all required fields are present

2. **Maintain RBAC properly:**
   - Document required permissions
   - Test RBAC changes in staging
   - Monitor for permission denials
   - Regular RBAC audits

3. **Template best practices:**
   ```yaml
   # Always specify API version
   apiVersion: apps/v1
   kind: Deployment

   # Provide defaults for optional fields
   replicas: {{ .replicas | default 1 }}

   # Validate required fields are present
   {{- if not .uid }}
   {{- fail "uid is required" }}
   {{- end }}

   # Use appropriate types
   replicas: {{ .replicas | int }}  # Ensure integer
   ```

4. **Monitor apply metrics:**
   ```promql
   # Track apply failure rate by kind
   rate(apply_attempts_total{result="failure"}[5m]) by (kind)

   # Alert on any failures for critical resource kinds
   rate(apply_attempts_total{result="failure", kind="Deployment"}[5m]) > 0
   ```

5. **Test admission webhooks:**
   - Test webhook policies with operator-generated resources
   - Ensure operator namespaces are excluded if appropriate
   - Document webhook requirements

## Monitoring

```promql
# Apply failure rate by kind
sum(rate(apply_attempts_total{result="failure"}[5m])) by (kind)
  / sum(rate(apply_attempts_total[5m])) by (kind)

# Total apply failures by kind
sum(increase(apply_attempts_total{result="failure"}[1h])) by (kind)

# Apply attempts by conflict policy
sum(rate(apply_attempts_total[5m])) by (kind, conflict_policy, result)

# Success rate by kind
sum(rate(apply_attempts_total{result="success"}[5m])) by (kind)
  / sum(rate(apply_attempts_total[5m])) by (kind)
```

## Related Alerts

- `TenantResourcesFailed` - Individual tenant failures
- `TenantDegraded` - Tenants degraded due to apply failures
- `TenantReconciliationErrors` - Overall reconciliation errors

## Investigation Checklist

- [ ] Identify which resource kind is failing
- [ ] Check RBAC permissions for that kind
- [ ] Verify resource specifications in templates
- [ ] Check for admission webhooks blocking the kind
- [ ] Review quota limits for the resource kind
- [ ] Check API version compatibility
- [ ] Review operator logs for specific errors
- [ ] Test resource creation with dry-run
- [ ] Check for recent cluster or operator changes

## Escalation

If apply failures persist after troubleshooting:

1. Collect diagnostics:
   ```bash
   # Operator logs filtered for failing kind
   kubectl logs -n tenant-operator-system -l app=tenant-operator --tail=5000 \
     | grep "kind=$FAILING_KIND" > apply-failures-$FAILING_KIND.log

   # RBAC for failing kind
   kubectl get clusterrole tenant-operator-manager-role -o yaml > rbac.yaml

   # Template definitions
   kubectl get tenanttemplates --all-namespaces -o yaml > templates.yaml

   # Sample failing resource
   kubectl get tenants --all-namespaces -o json \
     | jq -r ".items[0].spec.${FAILING_KIND}s[0]" > sample-resource.yaml

   # Admission webhooks
   kubectl get validatingwebhookconfigurations -o yaml > webhooks-validating.yaml
   kubectl get mutatingwebhookconfigurations -o yaml > webhooks-mutating.yaml

   # Resource quotas
   kubectl get resourcequotas --all-namespaces -o yaml > quotas.yaml
   ```

2. Analyze patterns:
   - Is it affecting all tenants or specific ones?
   - Did it start after a specific change?
   - Is it isolated to one resource kind?
   - Are there common error messages?

3. If unable to resolve:
   - Review with platform engineering
   - Check for operator bugs
   - Review Kubernetes API changes
   - Open GitHub issue with full diagnostics
