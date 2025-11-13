# Performance Tuning Guide

Practical optimization strategies for scaling Lynq to thousands of nodes.

[[toc]]

## Understanding Performance

Lynq uses three reconciliation layers:

1. **Event-Driven (Immediate)**: Reacts to resource changes instantly
2. **Periodic (30 seconds)**: Fast status updates and drift detection
3. **Database Sync (Configurable)**: Syncs node data at defined intervals

This architecture ensures:
- ✅ Immediate drift correction
- ✅ Fast status reflection (30s)
- ✅ Configurable database sync frequency

## Configuration Tuning

### 1. Database Sync Interval

Adjust how frequently the operator checks your database:

```yaml
apiVersion: operator.lynq.sh/v1
kind: LynqHub
metadata:
  name: my-hub
spec:
  source:
    syncInterval: 1m  # Default: 1 minute
```

**Recommendations:**
- **High-frequency changes**: `30s` - Faster node provisioning, higher DB load
- **Normal usage**: `1m` (default) - Balanced performance
- **Stable nodes**: `5m` - Lower DB load, slower updates

### 2. Resource Wait Timeouts

Control how long to wait for resources to become ready:

```yaml
deployments:
  - id: app
    waitForReady: true
    timeoutSeconds: 300  # Default: 5 minutes (max: 3600)
```

**Recommendations:**
- **Fast services**: `60s` - Quick deployments (< 1 min)
- **Normal apps**: `300s` (default) - Standard deployments
- **Heavy apps**: `600s` - Database migrations, complex initialization
- **Skip waiting**: Set `waitForReady: false` for non-critical resources

### 3. Creation Policy Optimization

Reduce unnecessary reconciliations:

```yaml
configMaps:
  - id: init-config
    creationPolicy: Once  # Create once, never reapply
```

**Use Cases:**
- `Once`: Init scripts, immutable configs, security resources
- `WhenNeeded` (default): Normal resources that may need updates

## Template Optimization

### 1. Keep Templates Simple

**✅ Good - Efficient template:**
```yaml
nameTemplate: "{{ .uid }}-app"
```

**❌ Bad - Complex template:**
```yaml
nameTemplate: "{{ .uid }}-{{ .region }}-{{ .planId }}-{{ now | date \"20060102\" }}"
# Avoid: timestamps, random values, complex logic
```

**Tips:**
- Keep templates simple and predictable
- Avoid `now`, `randAlphaNum`, or other non-deterministic functions
- Use consistent naming patterns
- Cache-friendly templates improve performance

### 2. Dependency Graph Optimization

**✅ Good - Shallow dependency tree:**
```yaml
resources:
  - id: namespace      # No dependencies
  - id: deployment     # Depends on: namespace
  - id: service        # Depends on: deployment
# Depth: 3 - Resources can be created in parallel groups
```

**❌ Bad - Deep dependency tree:**
```yaml
resources:
  - id: a              # No dependencies
  - id: b              # Depends on: a
  - id: c              # Depends on: b
  - id: d              # Depends on: c
  - id: e              # Depends on: d
# Depth: 5 - Fully sequential, slow
```

**Impact:**
- Shallow trees enable parallel execution
- Deep trees force sequential execution
- Each level adds wait time

### 3. Minimize Resource Count

**Example:** Create 5 essential resources per node instead of 15

```yaml
# Essential only
spec:
  namespaces: [1]
  deployments: [1]
  services: [1]
  configMaps: [1]
  ingresses: [1]
# Total: 5 resources
```

**Impact:**
- Fewer resources = Faster reconciliation
- Less API server load
- Lower memory usage

## Scaling Considerations

### Resource Limits

Adjust operator resource limits based on node count:

```yaml
# values.yaml for Helm
resources:
  limits:
    cpu: 2000m      # For 1000+ nodes
    memory: 2Gi     # For 1000+ nodes
  requests:
    cpu: 500m       # Minimum for stable operation
    memory: 512Mi   # Minimum for stable operation
```

**Guidelines:**
- **< 100 nodes**: Default limits (500m CPU, 512Mi RAM)
- **100-500 nodes**: 1 CPU, 1Gi RAM
- **500-1000 nodes**: 2 CPU, 2Gi RAM
- **1000+ nodes**: Consider horizontal scaling (coming in v1.3)

### Database Optimization

1. **Add indexes** to node table:
```sql
CREATE INDEX idx_is_active ON node_configs(is_active);
CREATE INDEX idx_node_id ON node_configs(node_id);
```

2. **Use read replicas** for high-frequency syncs

3. **Connection pooling**: Operator uses persistent connections

## Monitoring Performance

### Key Metrics to Watch

Monitor these Prometheus metrics:

```promql
# Reconciliation duration (target: < 5s P95)
histogram_quantile(0.95,
  sum(rate(lynqnode_reconcile_duration_seconds_bucket[5m])) by (le)
)

# Node readiness rate (target: > 95%)
sum(lynqnode_resources_ready) / sum(lynqnode_resources_desired)

# High error rate alert (target: < 5%)
sum(rate(lynqnode_reconcile_duration_seconds_count{result="error"}[5m]))
/ sum(rate(lynqnode_reconcile_duration_seconds_count[5m]))
```

See [Monitoring Guide](monitoring.md) for complete metrics reference.

## Troubleshooting Slow Performance

### Symptom: Slow Node Creation

**Check:**
1. Database query performance
2. `waitForReady` timeouts
3. Dependency chain depth

**Solution:**
```bash
# Check reconciliation times
kubectl logs -n lynq-system -l control-plane=controller-manager | grep "Reconciliation completed"

# Reduce sync interval if database is slow
kubectl patch lynqhub my-hub --type=merge -p '{"spec":{"source":{"syncInterval":"2m"}}}'
```

### Symptom: High CPU Usage

**Check:**
1. Reconciliation frequency
2. Template complexity
3. Total node count

**Solution:**
```bash
# Check CPU usage
kubectl top pods -n lynq-system

# Increase resource limits
kubectl edit deployment -n lynq-system lynq-controller-manager
```

### Symptom: Memory Growth

**Possible causes:**
1. Too many cached resources
2. Large template outputs
3. Memory leak (file an issue)

**Solution:**
```bash
# Restart operator to clear cache
kubectl rollout restart deployment -n lynq-system lynq-controller-manager

# Monitor memory over time
kubectl top pods -n lynq-system --watch
```

## Best Practices Summary

1. **✅ Start with defaults** - Only optimize if you see issues
2. **✅ Keep templates simple** - Avoid complex logic and non-deterministic functions
3. **✅ Use shallow dependency trees** - Enable parallel resource creation
4. **✅ Set appropriate timeouts** - Balance speed vs reliability
5. **✅ Monitor key metrics** - Watch reconciliation duration and error rates
6. **✅ Index your database** - Improve sync query performance
7. **✅ Use `CreationPolicy: Once`** - For immutable resources

## See Also

- [Monitoring Guide](monitoring.md) - Complete metrics reference and dashboards
- [Prometheus Queries](prometheus-queries.md) - Ready-to-use queries
- [Configuration Guide](configuration.md) - All operator settings
- [Troubleshooting Guide](troubleshooting.md) - Common issues and solutions
