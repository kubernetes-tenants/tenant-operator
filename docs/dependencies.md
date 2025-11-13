# Dependency Management Guide

Resource ordering and dependency graphs in Lynq.

[[toc]]

## Overview

Lynq uses a DAG (Directed Acyclic Graph) to order resource creation and ensure dependencies are satisfied before applying resources.

## Dependency Visualizer

The Lynq includes an interactive dependency graph visualizer tool that helps you:

- **Visualize Dependencies**: See the complete dependency graph of your LynqForm
- **Detect Cycles**: Automatically identify circular dependencies that would cause failures
- **Understand Execution Order**: View numbered badges showing the order resources will be applied
- **Test Your Templates**: Paste your YAML and analyze dependencies before deployment

::: tip Interactive Tool Available
Visit the **[🔍 Dependency Visualizer](./dependency-visualizer.md)** page to analyze your LynqForm dependencies interactively. Load preset examples or paste your own YAML to visualize the dependency graph in real-time.
:::

## Defining Dependencies

Use the `dependIds` field to specify dependencies:

::: info Syntax
Set `dependIds` to an array of resource IDs. The controller ensures all referenced resources finish reconciling before applying the dependent resource.
:::

```yaml
deployments:
  - id: app
    dependIds: ["secret"]  # Wait for secret first
    nameTemplate: "{{ .uid }}-app"
    spec:
      # ... deployment spec
```

## Dependency Resolution

### Topological Sort

Resources are applied in dependency order:

```
secret (no dependencies)
  ↓
app (depends on: secret)
  ↓
svc (depends on: app)
```

### Cycle Detection

Circular dependencies are rejected:

::: warning Why it fails
Dependency resolution uses a DAG. Any cycle blocks reconciliation and surfaces as `DependencyError`.
:::

```yaml
# ❌ This will fail
- id: a
  dependIds: ["b"]
- id: b
  dependIds: ["a"]
```

Error: `DependencyError: Dependency cycle detected: a -> b -> a`

## Common Patterns

### Pattern 1: Secrets Before Apps

```yaml
secrets:
  - id: api-secret
    nameTemplate: "{{ .uid }}-secret"
    # No dependencies

deployments:
  - id: app
    dependIds: ["api-secret"]  # Wait for secret
```

### Pattern 2: ConfigMap Before Deployment

```yaml
configMaps:
  - id: app-config
    nameTemplate: "{{ .uid }}-config"

deployments:
  - id: app
    dependIds: ["app-config"]  # Wait for configmap
```

### Pattern 3: App Before Service

```yaml
deployments:
  - id: app
    # No dependencies

services:
  - id: svc
    dependIds: ["app"]  # Wait for deployment first
```

### Pattern 4: PVC Before StatefulSet

```yaml
persistentVolumeClaims:
  - id: data-pvc
    # No dependencies

statefulSets:
  - id: stateful-app
    dependIds: ["data-pvc"]  # Wait for PVC
```

## Readiness Gates

Use `waitForReady` to wait for resource readiness:

::: tip Combine readiness and dependencies
`dependIds` only guarantees creation order. Enable `waitForReady` to ensure *ready* status before dependent workloads roll out.
:::

```yaml
deployments:
  - id: db
    waitForReady: true
    timeoutSeconds: 300

deployments:
  - id: app
    dependIds: ["db"]  # Wait for db to exist AND be ready
    waitForReady: true
```

## Best Practices

### 1. Shallow Dependencies

Prefer shallow dependency trees:

**Good (depth: 2):**
```
secret
  ├─ app
  │   └─ svc
  └─ db
      └─ db-svc
```

**Bad (depth: 5):**
```
secret → config → pvc → db → db-svc → app
```

### 2. Parallel Execution

Independent resources execute in parallel:

```yaml
deployments:
  - id: app-a
    dependIds: ["secret"]  # Both execute in parallel

  - id: app-b
    dependIds: ["secret"]  # Both execute in parallel
```

### 3. Minimal Dependencies

Only specify necessary dependencies:

**Good:**
```yaml
- id: app
  dependIds: ["secret"]  # Only essential dependency
```

**Bad:**
```yaml
- id: app
  dependIds: ["secret", "unrelated-resource"]  # Unnecessary wait
```

## Debugging Dependencies

### Common Errors

::: danger Cycle detected
```
DependencyError: Dependency cycle detected: a -> b -> c -> a
```
**Fix:** Remove or refactor at least one edge so the graph becomes acyclic.
:::

::: warning Missing dependency
```
DependencyError: Resource 'app' depends on 'missing-id' which does not exist
```
**Fix:** Ensure every `dependIds` entry references a defined resource ID.
:::

::: warning Readiness timeout
```
ReadinessTimeout: Resource db not ready within 300s
```
**Fix:** Increase `timeoutSeconds` or set `waitForReady: false` when readiness should not block dependent resources.
:::

## See Also

- [🔍 Dependency Visualizer](dependency-visualizer.md) - Interactive tool for analyzing dependencies
- [Template Guide](templates.md)
- [Policies Guide](policies.md)
- [Troubleshooting Guide](troubleshooting.md)
