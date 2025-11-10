# Prometheus Query Examples

This document provides ready-to-use PromQL queries for monitoring Tenant Operator.

[[toc]]

## Tenant Health

### Check Ready Tenants

```promql
# All ready tenants
tenant_condition_status{type="Ready"} == 1

# Count ready tenants
count(tenant_condition_status{type="Ready"} == 1)

# Percentage of ready tenants
count(tenant_condition_status{type="Ready"} == 1) / count(tenant_condition_status{type="Ready"}) * 100
```

### Check Not Ready Tenants

```promql
# All not ready tenants
tenant_condition_status{type="Ready"} != 1

# Count not ready tenants
count(tenant_condition_status{type="Ready"} != 1)

# List not ready tenants with details
tenant_condition_status{type="Ready"} != 1
```

### Check Degraded Tenants

```promql
# All degraded tenants
tenant_degraded_status == 1

# Count degraded tenants
count(tenant_degraded_status == 1)

# Degraded tenants by reason
sum(tenant_degraded_status) by (reason)

# Top 10 degraded tenants
topk(10, tenant_degraded_status)

# Degraded tenants with resources not ready (v1.1.4+)
tenant_degraded_status{reason="ResourcesNotReady"} == 1

# Count by specific degraded reason
sum(tenant_degraded_status{reason="ResourcesNotReady"})

# Tenants with resource failures only
tenant_degraded_status{reason="ResourceFailures"} == 1

# Tenants with conflicts only
tenant_degraded_status{reason="ResourceConflicts"} == 1

# Tenants with both failures and conflicts
tenant_degraded_status{reason="ResourceFailuresAndConflicts"} == 1
```

### Resource Health by Tenant

```promql
# Ready resources per tenant
tenant_resources_ready

# Failed resources per tenant
tenant_resources_failed

# Resource readiness percentage per tenant
(tenant_resources_ready / tenant_resources_desired) * 100

# Tenants with 100% resources ready
(tenant_resources_ready / tenant_resources_desired) == 1
```

## Conflict Monitoring

### Current Conflicts

```promql
# Total resources currently in conflict
sum(tenant_resources_conflicted)

# Tenants with conflicts
tenant_resources_conflicted > 0

# Top 10 tenants with most conflicts
topk(10, tenant_resources_conflicted)

# Conflict percentage per tenant
(tenant_resources_conflicted / tenant_resources_desired) * 100
```

### Conflict Rate

```promql
# Conflict rate (conflicts per second)
rate(tenant_conflicts_total[5m])

# Conflict rate per tenant
sum(rate(tenant_conflicts_total[5m])) by (tenant)

# Conflict rate by resource kind
sum(rate(tenant_conflicts_total[5m])) by (resource_kind)

# Conflict rate by policy
sum(rate(tenant_conflicts_total[5m])) by (conflict_policy)
```

### Historical Conflicts

```promql
# Total conflicts in last hour
increase(tenant_conflicts_total[1h])

# Total conflicts in last 24 hours
increase(tenant_conflicts_total[24h])

# Conflicts over time (5m windows)
sum(increase(tenant_conflicts_total[5m])) by (tenant)
```

### Conflict Policy Analysis

```promql
# Conflicts by policy type
sum(tenant_conflicts_total) by (conflict_policy)

# Force policy usage rate
rate(tenant_conflicts_total{conflict_policy="Force"}[5m])

# Stuck policy conflicts
rate(tenant_conflicts_total{conflict_policy="Stuck"}[5m])
```

## Failure Detection

### Failed Resources

```promql
# Total failed resources
sum(tenant_resources_failed)

# Tenants with failed resources
tenant_resources_failed > 0

# Top 10 tenants with most failures
topk(10, tenant_resources_failed)

# Failure rate per tenant
(tenant_resources_failed / tenant_resources_desired) * 100
```

### Failure Trends

```promql
# Failed resources over time
tenant_resources_failed

# Increase in failures (last 1h)
increase(tenant_resources_failed[1h])

# Average failures per tenant
avg(tenant_resources_failed)
```

### Critical Failures

```promql
# Tenants with >50% resources failed
(tenant_resources_failed / tenant_resources_desired) > 0.5

# Tenants with >5 failed resources
tenant_resources_failed > 5

# Tenants that are both degraded and have failures
tenant_degraded_status == 1 and tenant_resources_failed > 0
```

## Performance Monitoring

### Reconciliation Duration

```promql
# P50 reconciliation duration
histogram_quantile(0.50, rate(tenant_reconcile_duration_seconds_bucket[5m]))

# P95 reconciliation duration
histogram_quantile(0.95, rate(tenant_reconcile_duration_seconds_bucket[5m]))

# P99 reconciliation duration
histogram_quantile(0.99, rate(tenant_reconcile_duration_seconds_bucket[5m]))

# Max reconciliation duration
histogram_quantile(1.0, rate(tenant_reconcile_duration_seconds_bucket[5m]))
```

### Reconciliation Rate

```promql
# Total reconciliation rate
rate(tenant_reconcile_duration_seconds_count[5m])

# Success rate
rate(tenant_reconcile_duration_seconds_count{result="success"}[5m])

# Error rate
rate(tenant_reconcile_duration_seconds_count{result="error"}[5m])

# Success percentage
(rate(tenant_reconcile_duration_seconds_count{result="success"}[5m]) / rate(tenant_reconcile_duration_seconds_count[5m])) * 100
```

### Apply Performance

```promql
# Apply rate by result
sum(rate(apply_attempts_total[5m])) by (result)

# Apply rate by resource kind
sum(rate(apply_attempts_total[5m])) by (kind)

# Apply success rate
rate(apply_attempts_total{result="success"}[5m]) / rate(apply_attempts_total[5m])

# Failed applies by kind
sum(rate(apply_attempts_total{result="error"}[5m])) by (kind)
```

## Registry Health

### Registry Status

```promql
# Desired tenants per registry
registry_desired

# Ready tenants per registry
registry_ready

# Failed tenants per registry
registry_failed

# Registry health percentage
(registry_ready / registry_desired) * 100
```

### Registry Capacity

```promql
# Total desired tenants across all registries
sum(registry_desired)

# Total ready tenants across all registries
sum(registry_ready)

# Total failed tenants across all registries
sum(registry_failed)

# Overall health percentage
(sum(registry_ready) / sum(registry_desired)) * 100
```

### Registry Trends

```promql
# Registry health over time
(registry_ready / registry_desired) * 100

# Registries with >90% health
(registry_ready / registry_desired) > 0.9

# Unhealthy registries (<80% ready)
(registry_ready / registry_desired) < 0.8
```

## Capacity Planning

### Resource Counts

```promql
# Total desired resources across all tenants
sum(tenant_resources_desired)

# Total ready resources
sum(tenant_resources_ready)

# Total failed resources
sum(tenant_resources_failed)

# Total conflicted resources
sum(tenant_resources_conflicted)
```

### Growth Trends

```promql
# Desired tenant growth rate
rate(registry_desired[24h])

# Resource growth per tenant
rate(tenant_resources_desired[24h])

# Average resources per tenant
avg(tenant_resources_desired)
```

### Load Distribution

```promql
# Top 10 tenants by resource count
topk(10, tenant_resources_desired)

# Bottom 10 tenants by resource count
bottomk(10, tenant_resources_desired)

# Tenants with >100 resources
tenant_resources_desired > 100

# Distribution of resources per tenant
histogram_quantile(0.50, tenant_resources_desired)
```

## Combined Queries

### Overall Health Dashboard

```promql
# Total tenants
count(tenant_condition_status{type="Ready"})

# Ready percentage
count(tenant_condition_status{type="Ready"} == 1) / count(tenant_condition_status{type="Ready"}) * 100

# Total resources
sum(tenant_resources_desired)

# Ready resources percentage
sum(tenant_resources_ready) / sum(tenant_resources_desired) * 100

# Active conflicts
sum(tenant_resources_conflicted)

# Total failures
sum(tenant_resources_failed)
```

### Problem Detection

```promql
# Tenants with issues (not ready OR degraded OR conflicts OR failures)
(tenant_condition_status{type="Ready"} != 1)
or (tenant_degraded_status == 1)
or (tenant_resources_conflicted > 0)
or (tenant_resources_failed > 0)

# Count problematic tenants
count(
  (tenant_condition_status{type="Ready"} != 1)
  or (tenant_degraded_status == 1)
  or (tenant_resources_conflicted > 0)
  or (tenant_resources_failed > 0)
)
```

### Performance Summary

```promql
# P95 latency, error rate, and throughput
{
  p95_latency: histogram_quantile(0.95, rate(tenant_reconcile_duration_seconds_bucket[5m])),
  error_rate: rate(tenant_reconcile_duration_seconds_count{result="error"}[5m]),
  throughput: rate(tenant_reconcile_duration_seconds_count[5m])
}
```

## Alert Conditions

These queries are used in the alert rules (`config/prometheus/alerts.yaml`):

### Critical Conditions

```promql
# Tenant has failed resources
tenant_resources_failed > 0

# Tenant is degraded
tenant_degraded_status > 0

# Tenant not ready
tenant_condition_status{type="Ready"} != 1

# Registry has many failures
registry_failed > 5 or (registry_failed / registry_desired > 0.5 and registry_desired > 0)
```

### Warning Conditions

```promql
# Resources in conflict
tenant_resources_conflicted > 0

# High conflict rate
rate(tenant_conflicts_total[5m]) > 0.1

# Resources mismatch
tenant_resources_ready != tenant_resources_desired and tenant_resources_desired > 0

# Slow reconciliation
histogram_quantile(0.95, rate(tenant_reconcile_duration_seconds_bucket[5m])) > 30

# Resources not ready (v1.1.4+)
tenant_degraded_status{reason="ResourcesNotReady"} > 0

# Tenants with both failures and conflicts (v1.1.4+)
tenant_degraded_status{reason="ResourceFailuresAndConflicts"} > 0
```

## v1.1.4 Enhanced Status Queries

::: tip New in v1.1.4
v1.1.4 introduces more granular degraded condition reasons and smart reconciliation with 30-second requeue interval.
:::

### New Degraded Reasons

```promql
# All tenants degraded due to resources not ready
tenant_degraded_status{reason="ResourcesNotReady"} == 1

# Tenants with resource failures only
tenant_degraded_status{reason="ResourceFailures"} == 1

# Tenants with conflicts only
tenant_degraded_status{reason="ResourceConflicts"} == 1

# Tenants with both failures and conflicts
tenant_degraded_status{reason="ResourceFailuresAndConflicts"} == 1

# Count tenants by degraded reason
sum(tenant_degraded_status == 1) by (reason)
```

### Ready Condition Granularity

Check why tenants are not ready:

```promql
# All not-ready tenants with detailed reason
tenant_condition_status{type="Ready"} != 1

# Degraded tenants with Ready=False
tenant_condition_status{type="Ready"} != 1 and tenant_condition_status{type="Degraded"} == 1

# Query tenant annotations for reason details (requires external tooling)
# Reasons: ResourcesFailedAndConflicted, ResourcesConflicted,
#          ResourcesFailed, NotAllResourcesReady
```

### Smart Reconciliation Monitoring

Monitor the 30-second requeue behavior:

```promql
# Reconciliation frequency (should show ~2 per minute per tenant in v1.1.4+)
rate(tenant_reconcile_duration_seconds_count[5m])

# P50 latency (should remain low due to fast requeue)
histogram_quantile(0.50, rate(tenant_reconcile_duration_seconds_bucket[5m]))

# P95 latency (watch for spikes > 30s)
histogram_quantile(0.95, rate(tenant_reconcile_duration_seconds_bucket[5m]))

# Detect reconciliation bottlenecks
histogram_quantile(0.95, rate(tenant_reconcile_duration_seconds_bucket[5m])) > 10

# Status-only reconciliations (fast path)
rate(tenant_reconcile_duration_seconds_count{result="status_only"}[5m])
```

### Readiness Tracking

Track how quickly resources become ready:

```promql
# Percentage of resources ready
(sum(tenant_resources_ready) / sum(tenant_resources_desired)) * 100

# Tenants with incomplete readiness
tenant_resources_ready < tenant_resources_desired

# Average time to readiness (approximation via reconciliation duration)
avg(rate(tenant_reconcile_duration_seconds_sum{result="success"}[5m]) /
    rate(tenant_reconcile_duration_seconds_count{result="success"}[5m]))
```

## Tips for Using These Queries

1. **Adjust Time Windows**: Change `[5m]`, `[1h]`, `[24h]` based on your needs
2. **Filter by Namespace**: Add `{namespace="default"}` to filter
3. **Filter by Tenant**: Add `{tenant="my-tenant"}` to focus on specific tenants
4. **Combine Queries**: Use `and`, `or`, `unless` for complex conditions
5. **Aggregation**: Use `sum`, `avg`, `max`, `min` for aggregations
6. **Top/Bottom N**: Use `topk(N, ...)` or `bottomk(N, ...)`

### Example: Filter by Namespace and Time

```promql
# Failed resources in default namespace, last 1 hour
tenant_resources_failed{namespace="default"}[1h]

# Conflicts for specific tenant in last 5 minutes
rate(tenant_conflicts_total{tenant="acme-prod-template", namespace="default"}[5m])
```

## See Also

- [Monitoring Guide](monitoring.md) - Complete monitoring documentation
- [Alert Rules](../config/prometheus/alerts.yaml) - Prometheus alert rules
- [Troubleshooting](troubleshooting.md) - Common issues and solutions
