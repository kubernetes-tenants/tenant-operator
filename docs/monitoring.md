# Monitoring & Observability Guide

This guide covers metrics, logging, events, and alerting for Tenant Operator.

[[toc]]

## Getting Started

### Accessing Metrics

::: info Endpoint
Tenant Operator exposes Prometheus metrics at `:8443/metrics` over HTTPS.
:::

**Port-forward for local testing:**

```bash
# Port-forward to metrics endpoint
kubectl port-forward -n tenant-operator-system \
  deployment/tenant-operator-controller-manager 8443:8443

# Access metrics (requires valid TLS client or use --insecure)
curl -k https://localhost:8443/metrics
```

**Check if metrics are enabled:**

```bash
# Check if metrics port is exposed
kubectl get svc -n tenant-operator-system tenant-operator-controller-manager-metrics-service

# Check if ServiceMonitor is deployed (requires prometheus-operator)
kubectl get servicemonitor -n tenant-operator-system
```

### Enabling ServiceMonitor

If using Prometheus Operator, enable ServiceMonitor by uncommenting in `config/default/kustomization.yaml`:

```yaml
# Line 27: Uncomment this
- ../prometheus
```

Then redeploy:

```bash
kubectl apply -k config/default
```

::: tip Verify scrape job
After redeploying, confirm that a `ServiceMonitor` named `tenant-operator-controller-manager` appears and that Prometheus discovers the target.
:::

## Metrics

Tenant Operator exposes the following custom Prometheus metrics at `:8443/metrics`.

### Controller Metrics

#### `tenant_reconcile_duration_seconds`

Histogram of tenant reconciliation duration.

**Labels:**
- `result`: `success` or `error`

**Queries:**
```promql
# 95th percentile reconciliation time
histogram_quantile(0.95, rate(tenant_reconcile_duration_seconds_bucket[5m]))

# Reconciliation rate
rate(tenant_reconcile_duration_seconds_count[5m])

# Error rate
rate(tenant_reconcile_duration_seconds{result="error"}[5m])
```

**Alerts:**
```yaml
- alert: SlowTenantReconciliation
  expr: histogram_quantile(0.95, rate(tenant_reconcile_duration_seconds_bucket[5m])) > 30
  for: 5m
  annotations:
    summary: Tenant reconciliation taking > 30s

- alert: TenantReconciliationErrors
  expr: rate(tenant_reconcile_duration_seconds{result="error"}[5m]) > 0.1
  annotations:
    summary: High tenant reconciliation error rate
```

### Resource Metrics

#### `tenant_resources_desired`

Gauge of desired resources for a tenant.

**Labels:**
- `tenant`: Tenant name
- `namespace`: Tenant namespace

**Queries:**
```promql
# Total desired resources
sum(tenant_resources_desired)

# Per tenant
tenant_resources_desired{tenant="acme-prod-template"}
```

#### `tenant_resources_ready`

Gauge of ready resources for a tenant.

**Labels:**
- `tenant`: Tenant name
- `namespace`: Tenant namespace

**Queries:**
```promql
# Total ready resources
sum(tenant_resources_ready)

# Readiness percentage
sum(tenant_resources_ready) / sum(tenant_resources_desired) * 100
```

#### `tenant_resources_failed`

Gauge of failed resources for a tenant.

**Labels:**
- `tenant`: Tenant name
- `namespace`: Tenant namespace

**Alerts:**
```yaml
- alert: TenantResourcesFailed
  expr: tenant_resources_failed > 0
  for: 5m
  annotations:
    summary: Tenant {{ $labels.tenant }} has {{ $value }} failed resources
```

### Registry Metrics

#### `registry_desired`

Gauge of desired tenant CRs for a registry.

**Labels:**
- `registry`: Registry name
- `namespace`: Registry namespace

**Queries:**
```promql
# Total desired tenants across all registries
sum(registry_desired)

# Per registry
registry_desired{registry="my-saas-registry"}
```

#### `registry_ready`

Gauge of ready tenant CRs for a registry.

**Queries:**
```promql
# Registry health percentage
sum(registry_ready) / sum(registry_desired) * 100
```

#### `registry_failed`

Gauge of failed tenant CRs for a registry.

**Alerts:**
```yaml
- alert: RegistryUnhealthy
  expr: registry_failed > 0
  for: 10m
  annotations:
    summary: Registry {{ $labels.registry }} has {{ $value }} failed tenants
```

### Apply Metrics

#### `apply_attempts_total`

Counter of resource apply attempts.

**Labels:**
- `kind`: Resource kind (Deployment, Service, etc.)
- `result`: `success` or `error`
- `conflict_policy`: `Stuck` or `Force`

**Queries:**
```promql
# Apply success rate
rate(apply_attempts_total{result="success"}[5m]) / rate(apply_attempts_total[5m])

# Applies per kind
sum(rate(apply_attempts_total[5m])) by (kind)

# Conflict policy usage
sum(rate(apply_attempts_total[5m])) by (conflict_policy)
```

**Alerts:**
```yaml
- alert: HighApplyFailureRate
  expr: rate(apply_attempts_total{result="error"}[5m]) / rate(apply_attempts_total[5m]) > 0.1
  annotations:
    summary: > 10% of apply attempts failing
```

### Conflict and Failure Metrics

#### `tenant_condition_status`

Gauge tracking the status of tenant conditions.

**Labels:**
- `tenant`: Tenant name
- `namespace`: Tenant namespace
- `type`: Condition type (Ready, Progressing, Conflicted, Degraded)

**Values:**
- `0`: False
- `1`: True
- `2`: Unknown

**Queries:**
```promql
# Check if tenants are ready
tenant_condition_status{type="Ready"} == 1

# Count tenants not ready
count(tenant_condition_status{type="Ready"} != 1)

# Check degraded tenants (v1.1.4+)
tenant_condition_status{type="Degraded"} == 1

# Count degraded tenants
count(tenant_condition_status{type="Degraded"} == 1)
```

**Alerts:**
```yaml
- alert: TenantNotReady
  expr: tenant_condition_status{type="Ready"} != 1
  for: 10m
  annotations:
    summary: Tenant {{ $labels.tenant }} is not ready
```

#### `tenant_conflicts_total`

Counter tracking the total number of resource conflicts encountered.

**Labels:**
- `tenant`: Tenant name
- `namespace`: Tenant namespace
- `resource_kind`: Kind of resource in conflict (Deployment, Service, etc.)
- `conflict_policy`: Applied policy (Stuck or Force)

**Queries:**
```promql
# Total conflicts
sum(tenant_conflicts_total)

# Conflicts per tenant
sum(rate(tenant_conflicts_total[5m])) by (tenant)

# Conflicts by resource kind
sum(rate(tenant_conflicts_total[5m])) by (resource_kind)

# Conflicts by policy
sum(rate(tenant_conflicts_total[5m])) by (conflict_policy)
```

**Alerts:**
```yaml
- alert: HighConflictRate
  expr: rate(tenant_conflicts_total[5m]) > 0.1
  for: 10m
  annotations:
    summary: High conflict rate for tenant {{ $labels.tenant }}

- alert: NewConflictsDetected
  expr: increase(tenant_conflicts_total[5m]) > 0
  for: 1m
  annotations:
    summary: New conflicts detected for tenant {{ $labels.tenant }}
```

#### `tenant_resources_conflicted`

Gauge tracking the current number of resources in conflict state.

**Labels:**
- `tenant`: Tenant name
- `namespace`: Tenant namespace

**Queries:**
```promql
# Total resources in conflict
sum(tenant_resources_conflicted)

# Tenants with conflicts
tenant_resources_conflicted > 0

# Conflict percentage
sum(tenant_resources_conflicted) / sum(tenant_resources_desired) * 100
```

**Alerts:**
```yaml
- alert: TenantResourcesConflicted
  expr: tenant_resources_conflicted > 0
  for: 10m
  annotations:
    summary: Tenant {{ $labels.tenant }} has {{ $value }} resources in conflict
```

#### `tenant_degraded_status`

Gauge indicating if a tenant is in degraded state.

**Labels:**
- `tenant`: Tenant name
- `namespace`: Tenant namespace
- `reason`: Reason for degradation (TemplateRenderError, ConflictDetected, DependencyCycle, ResourceFailuresAndConflicts, ResourceFailures, ResourceConflicts, ResourcesNotReady)

**Values:**
- `0`: Not degraded
- `1`: Degraded

**Queries:**
```promql
# Count degraded tenants
count(tenant_degraded_status == 1)

# List degraded tenants with reasons
tenant_degraded_status{reason!=""} == 1

# Degraded tenants by reason
sum(tenant_degraded_status) by (reason)

# Tenants with resources not ready (v1.1.4+)
tenant_degraded_status{reason="ResourcesNotReady"} == 1

# Count by specific degraded reason
sum(tenant_degraded_status{reason="ResourcesNotReady"})
```

**Alerts:**
```yaml
- alert: TenantDegraded
  expr: tenant_degraded_status > 0
  for: 5m
  annotations:
    summary: Tenant {{ $labels.tenant }} is degraded
    description: "Reason: {{ $labels.reason }}"

- alert: TenantNotAllResourcesReady
  expr: tenant_degraded_status{reason="ResourcesNotReady"} == 1
  for: 5m
  annotations:
    summary: Tenant {{ $labels.tenant }} has resources that are not ready
    description: Check tenant status for readiness details
```

### Smart Reconciliation Metrics

::: tip New in v1.1.4
v1.1.4 introduces enhanced status tracking with a 30-second requeue interval for fast status reflection.
:::

**Key Changes:**
- **Fast Status Updates**: Child resource status changes reflected in Tenant status within 30 seconds (down from 5 minutes)
- **Event-Driven**: Immediate reconciliation on watched resource changes
- **Smart Predicates**: Only reconcile on Generation/Annotation changes, not status-only updates

**Impact on Metrics:**

The 30-second requeue interval means you'll see:
- **Higher reconciliation frequency**: ~2 reconciles per minute per tenant
- **Lower latency**: Status changes propagate faster
- **Optimized overhead**: Smart predicates filter unnecessary reconciliations

**Monitoring Reconciliation Patterns:**

```promql
# Reconciliation frequency (should show ~2 per minute per tenant in v1.1.4+)
rate(tenant_reconcile_duration_seconds_count[5m])

# P50 latency (should remain low despite faster requeue)
histogram_quantile(0.50, rate(tenant_reconcile_duration_seconds_bucket[5m]))

# P95 latency (watch for spikes > 30s)
histogram_quantile(0.95, rate(tenant_reconcile_duration_seconds_bucket[5m]))

# Status-only reconciliations (fast path)
rate(tenant_reconcile_duration_seconds_count{result="status_only"}[5m])
```

**Alert for Reconciliation Bottlenecks:**

```yaml
- alert: SlowReconciliation
  expr: histogram_quantile(0.95, rate(tenant_reconcile_duration_seconds_bucket[5m])) > 30
  for: 5m
  annotations:
    summary: Tenant reconciliation P95 latency > 30s
    description: May indicate controller performance issues or resource readiness delays

- alert: HighReconciliationRate
  expr: rate(tenant_reconcile_duration_seconds_count[5m]) > 5
  for: 10m
  annotations:
    summary: Unusually high reconciliation rate
    description: May indicate reconciliation loop or frequent resource changes
```

**Best Practices:**
1. **Capacity Planning**: Monitor reconciliation rate for horizontal scaling decisions
2. **Latency Tracking**: P95 latency should stay under 10s for healthy systems
3. **Event-Driven Behavior**: Most reconciliations should be triggered by resource changes, not periodic requeues
4. **Watch Predicates**: Verify that status-only updates don't trigger full reconciliations

### Controller-Runtime Metrics

Standard controller-runtime metrics:

```promql
# Work queue depth
workqueue_depth{name="tenant"}

# Work queue add rate
rate(workqueue_adds_total{name="tenant"}[5m])

# Work queue latency
workqueue_queue_duration_seconds{name="tenant"}
```

## Metrics Collection

### Prometheus ServiceMonitor

To enable ServiceMonitor, uncomment the prometheus section in `config/default/kustomization.yaml`:

```yaml
# Uncomment this line:
#- ../prometheus
```

The ServiceMonitor configuration (already available in `config/prometheus/monitor.yaml`):

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    control-plane: controller-manager
    app.kubernetes.io/name: tenant-operator
    app.kubernetes.io/managed-by: kustomize
  name: controller-manager-metrics-monitor
  namespace: tenant-operator-system
spec:
  endpoints:
  - path: /metrics
    port: https
    scheme: https
    bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    tlsConfig:
      insecureSkipVerify: true
  selector:
    matchLabels:
      control-plane: controller-manager
      app.kubernetes.io/name: tenant-operator
```

**Note:** For production, use cert-manager for metrics TLS by enabling the cert patch in `config/default/kustomization.yaml`.

### Manual Scrape Configuration

```yaml
# prometheus.yml
scrape_configs:
- job_name: 'tenant-operator'
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names:
      - tenant-operator-system
  relabel_configs:
  - source_labels: [__meta_kubernetes_pod_label_control_plane]
    action: keep
    regex: controller-manager
  - source_labels: [__meta_kubernetes_pod_container_port_name]
    action: keep
    regex: https
```

## Logging

### Log Levels

Configure via `--zap-log-level`:

```yaml
args:
- --zap-log-level=info  # Options: debug, info, error
```

**Levels:**
- `debug`: Verbose logging (template values, API calls)
- `info`: Standard logging (reconciliation events)
- `error`: Errors only

### Structured Logging

All logs are structured JSON:

```json
{
  "level": "info",
  "ts": "2025-01-15T10:30:00.000Z",
  "msg": "Reconciliation completed",
  "tenant": "acme-prod-template",
  "ready": 10,
  "failed": 0,
  "changed": 2
}
```

### Key Log Messages

#### Reconciliation Events

```
"msg": "Reconciliation completed"
"msg": "Reconciliation completed with changes"
"msg": "Failed to reconcile tenant"
```

#### Resource Events

```
"msg": "Failed to render resource"
"msg": "Failed to apply resource"
"msg": "Resource not ready within timeout"
```

#### Registry Events

```
"msg": "Deleting Tenant (no longer in desired set)"
"msg": "Successfully deleted Tenant"
```

### Querying Logs

```bash
# All logs
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager

# Follow logs
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager -f

# Errors only
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager | grep '"level":"error"'

# Specific tenant
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager | grep 'acme-prod'

# Reconciliation events
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager | grep "Reconciliation completed"
```

## Events

Kubernetes events are emitted for key operations.

### Viewing Events

```bash
# All Tenant events
kubectl get events --all-namespaces --field-selector involvedObject.kind=Tenant

# Specific Tenant
kubectl describe tenant <name>

# Recent events
kubectl get events --sort-by='.lastTimestamp'
```

### Event Types

#### Normal Events

- `TemplateApplied`: Template successfully applied
- `TemplateAppliedComplete`: All resources applied
- `TenantDeleting`: Tenant deletion started
- `TenantDeleted`: Tenant deletion completed

#### Warning Events

- `TemplateRenderError`: Template rendering failed
- `ApplyFailed`: Resource apply failed
- `ResourceConflict`: Ownership conflict detected
- `ReadinessTimeout`: Resource not ready within timeout
- `DependencyError`: Dependency cycle detected
- `TenantDeletionFailed`: Tenant deletion failed

### Event Examples

```bash
# Success
TemplateAppliedComplete: Applied 10 resources (10 ready, 0 failed, 2 changed)

# Conflict
ResourceConflict: Resource conflict detected for default/acme-app (Kind: Deployment, Policy: Stuck).
Another controller or user may be managing this resource.

# Deletion
TenantDeleting: Deleting Tenant 'acme-prod-template' (template: prod-template, uid: acme) -
no longer in active dataset. This could be due to: row deletion, activate=false, or template change.
```

## Dashboards

### Grafana Dashboard

A comprehensive Grafana dashboard is available at: `config/monitoring/grafana-dashboard.json`

**How to import:**

1. Open Grafana UI
2. Go to Dashboards → Import
3. Upload `config/monitoring/grafana-dashboard.json`
4. Select your Prometheus datasource

**Dashboard includes 10 panels:**
1. **Reconciliation Duration (Percentiles)** - P50, P95, P99 latency
2. **Reconciliation Rate** - Success vs Error rate
3. **Error Rate** - Gauge showing current error percentage
4. **Total Desired Tenants** - Sum across all registries
5. **Total Ready Tenants** - Healthy tenant count
6. **Total Failed Tenants** - Failed tenant count
7. **Resource Counts by Tenant** - Stacked area chart per tenant
8. **Registry Health** - Table showing health percentage per registry
9. **Apply Rate by Kind** - Apply attempts by resource type
10. **Work Queue Depth** - Controller queue depths

### Sample Queries

**Reconciliation Performance:**
```promql
# P50, P95, P99 latency
histogram_quantile(0.50, rate(tenant_reconcile_duration_seconds_bucket[5m]))
histogram_quantile(0.95, rate(tenant_reconcile_duration_seconds_bucket[5m]))
histogram_quantile(0.99, rate(tenant_reconcile_duration_seconds_bucket[5m]))
```

**Resource Health:**
```promql
# % of resources ready
sum(tenant_resources_ready) / sum(tenant_resources_desired) * 100
```

**Top Slow Tenants:**
```promql
# Tenants with most failed resources
topk(10, tenant_resources_failed)
```

## Alerting

### Prometheus Alert Rules

A comprehensive set of Prometheus alert rules is available at `config/prometheus/alerts.yaml`.

**To deploy the alerts:**

```bash
# Apply the PrometheusRule resource
kubectl apply -f config/prometheus/alerts.yaml

# Or use kustomize
kubectl apply -k config/prometheus
```

**Alert Categories:**

1. **Critical Alerts**
   - `TenantResourcesFailed` - Tenant has failed resources
   - `TenantDegraded` - Tenant is in degraded state
   - `TenantNotReady` - Tenant not ready for extended period
   - `RegistryManyTenantsFailure` - Many tenants failing in a registry

2. **Warning Alerts**
   - `TenantResourcesConflicted` - Resources in conflict state
   - `TenantHighConflictRate` - High rate of conflicts
   - `TenantResourcesMismatch` - Ready count doesn't match desired
   - `RegistryTenantsFailure` - Some tenants failing
   - `TenantReconciliationErrors` - High error rate
   - `TenantReconciliationSlow` - Slow reconciliation performance

3. **Info Alerts**
   - `TenantNewConflictsDetected` - New conflicts detected

### Sample Alert Rules

#### Critical Alerts

```yaml
# Tenant has failed resources
- alert: TenantResourcesFailed
  expr: tenant_resources_failed > 0
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Tenant {{ $labels.tenant }} has failed resources"

# Tenant is degraded
- alert: TenantDegraded
  expr: tenant_degraded_status > 0
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Tenant {{ $labels.tenant }} is in degraded state"
    description: "Reason: {{ $labels.reason }}"

# Registry has many failed tenants
- alert: RegistryManyTenantsFailure
  expr: registry_failed > 5 or (registry_failed / registry_desired > 0.5 and registry_desired > 0)
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Registry {{ $labels.registry }} has many failed tenants"
```

#### Warning Alerts

```yaml
# Resources in conflict
- alert: TenantResourcesConflicted
  expr: tenant_resources_conflicted > 0
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Tenant {{ $labels.tenant }} has resources in conflict"

# High conflict rate
- alert: TenantHighConflictRate
  expr: rate(tenant_conflicts_total[5m]) > 0.1
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "High conflict rate for tenant {{ $labels.tenant }}"
```

#### Performance Alerts

```yaml
# Slow reconciliation
- alert: TenantReconciliationSlow
  expr: histogram_quantile(0.95, rate(tenant_reconcile_duration_seconds_bucket[5m])) > 30
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Slow tenant reconciliation"

# High error rate
- alert: TenantReconciliationErrors
  expr: rate(tenant_reconcile_duration_seconds_count{result="error"}[5m]) > 0.1
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "High tenant reconciliation error rate"
```

### Alert Routing (AlertManager)

Configure AlertManager to route alerts based on severity:

```yaml
# alertmanager.yml
route:
  group_by: ['alertname', 'tenant', 'namespace']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'default'
  routes:
  # Critical alerts to PagerDuty
  - match:
      severity: critical
    receiver: 'pagerduty'
    continue: true

  # Warning alerts to Slack
  - match:
      severity: warning
    receiver: 'slack'
    continue: true

  # Info alerts to email
  - match:
      severity: info
    receiver: 'email'

receivers:
- name: 'default'
  webhook_configs:
  - url: 'http://example.com/webhook'

- name: 'pagerduty'
  pagerduty_configs:
  - service_key: '<pagerduty-key>'

- name: 'slack'
  slack_configs:
  - api_url: '<slack-webhook>'
    channel: '#tenant-operator-alerts'
    text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'

- name: 'email'
  email_configs:
  - to: 'team@example.com'
    from: 'alertmanager@example.com'
```

## Tracing

### Distributed Tracing (Future)

Planned for v1.2:
- OpenTelemetry integration
- Trace reconciliation across controllers
- Span for each resource apply
- Database query tracing

## Best Practices

### 1. Monitor Key Metrics

Essential metrics to track:
- Reconciliation duration (P95)
- Error rate
- Resource ready/failed counts
- Registry desired vs ready

### 2. Set Up Alerts

Minimum recommended alerts:
- Operator down
- High error rate (> 10%)
- Slow reconciliation (P95 > 30s)
- Resources failed (> 0 for 5min)

### 3. Retain Logs

Recommended log retention:
- **Debug logs:** 1-3 days
- **Info logs:** 7-14 days
- **Error logs:** 30+ days

### 4. Dashboard Review

Weekly review:
- Reconciliation performance trends
- Error patterns
- Resource health
- Capacity planning

### 5. Event Monitoring

Monitor events for:
- Conflicts (investigate ownership)
- Timeouts (adjust readiness settings)
- Template errors (fix templates)

## Troubleshooting Metrics

### Metrics Not Available

**Problem:** `curl https://localhost:8443/metrics` returns connection refused.

**Solution:**

1. **Check if metrics port is configured:**
   ```bash
   kubectl get deployment -n tenant-operator-system tenant-operator-controller-manager -o yaml | grep metrics-bind-address
   ```

   Should see: `--metrics-bind-address=:8443`

2. **Check if port is exposed:**
   ```bash
   kubectl get deployment -n tenant-operator-system tenant-operator-controller-manager -o yaml | grep -A 5 "ports:"
   ```

   Should see containerPort 8443.

3. **Check if service exists:**
   ```bash
   kubectl get svc -n tenant-operator-system tenant-operator-controller-manager-metrics-service
   ```

4. **Check operator logs:**
   ```bash
   kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager | grep metrics
   ```

### No Metrics Data

**Problem:** Metrics endpoint works but returns no custom metrics.

**Solution:**

1. **Verify metrics are registered:**
   ```bash
   curl -k https://localhost:8443/metrics | grep tenant_
   ```

   Should see: `tenant_reconcile_duration_seconds`, `tenant_resources_ready`, etc.

2. **Trigger reconciliation:**
   ```bash
   # Apply a test resource
   kubectl apply -f config/samples/operator_v1_tenantregistry.yaml

   # Wait 30s and check metrics again
   curl -k https://localhost:8443/metrics | grep tenant_reconcile_duration_seconds_count
   ```

3. **Check if controllers are running:**
   ```bash
   kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager | grep "Starting Controller"
   ```

### ServiceMonitor Not Working

**Problem:** Prometheus not scraping metrics.

**Solution:**

1. **Check if Prometheus Operator is installed:**
   ```bash
   kubectl get crd servicemonitors.monitoring.coreos.com
   ```

2. **Check if ServiceMonitor is created:**
   ```bash
   kubectl get servicemonitor -n tenant-operator-system
   ```

3. **Check ServiceMonitor labels match Prometheus selector:**
   ```bash
   kubectl get servicemonitor -n tenant-operator-system tenant-operator-controller-manager-metrics-monitor -o yaml
   ```

4. **Check Prometheus logs:**
   ```bash
   kubectl logs -n monitoring prometheus-xyz
   ```

### TLS Certificate Errors

**Problem:** `x509: certificate signed by unknown authority`

**Solution:**

For development, use `--insecure` or `-k`:
```bash
curl -k https://localhost:8443/metrics
```

For production, use cert-manager by enabling the cert patch in `config/default/kustomization.yaml`:
```yaml
# Uncomment this line:
#- path: cert_metrics_manager_patch.yaml
```

## See Also

- [Performance Guide](performance.md) - Performance tuning
- [Troubleshooting Guide](troubleshooting.md) - Common issues
