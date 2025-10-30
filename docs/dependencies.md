# Dependency Management Guide

Resource ordering and dependency graphs in Tenant Operator.

## Overview

Tenant Operator uses a DAG (Directed Acyclic Graph) to order resource creation and ensure dependencies are satisfied before applying resources.

## Defining Dependencies

Use the `dependIds` field to specify dependencies:

```yaml
deployments:
  - id: app
    dependIds: ["ns", "secret"]  # Wait for namespace and secret first
    nameTemplate: "{{ .uid }}-app"
    spec:
      # ... deployment spec
```

## Dependency Resolution

### Topological Sort

Resources are applied in dependency order:

```
ns (no dependencies)
  ↓
secret (depends on: ns)
  ↓
app (depends on: ns, secret)
  ↓
svc (depends on: app)
```

### Cycle Detection

Circular dependencies are rejected:

```yaml
# ❌ This will fail
- id: a
  dependIds: ["b"]
- id: b
  dependIds: ["a"]
```

Error: `DependencyError: Dependency cycle detected: a -> b -> a`

## Common Patterns

### Pattern 1: Namespace First

```yaml
namespaces:
  - id: ns
    nameTemplate: "tenant-{{ .uid }}"
    # No dependencies

deployments:
  - id: app
    dependIds: ["ns"]  # Wait for namespace
```

### Pattern 2: Secrets Before Apps

```yaml
secrets:
  - id: api-secret
    dependIds: ["ns"]
    nameTemplate: "{{ .uid }}-secret"

deployments:
  - id: app
    dependIds: ["ns", "api-secret"]  # Wait for namespace and secret
```

### Pattern 3: App Before Service

```yaml
deployments:
  - id: app
    dependIds: ["ns"]

services:
  - id: svc
    dependIds: ["app"]  # Wait for deployment first
```

### Pattern 4: PVC Before StatefulSet

```yaml
persistentVolumeClaims:
  - id: data-pvc
    dependIds: ["ns"]

statefulSets:
  - id: stateful-app
    dependIds: ["data-pvc"]  # Wait for PVC
```

## Readiness Gates

Use `waitForReady` to wait for resource readiness:

```yaml
deployments:
  - id: db
    dependIds: ["ns"]
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

**Good (depth: 3):**
```
ns
  ├─ app
  │   └─ svc
  └─ db
      └─ db-svc
```

**Bad (depth: 6):**
```
ns → secret → config → pvc → db → db-svc → app
```

### 2. Parallel Execution

Independent resources execute in parallel:

```yaml
deployments:
  - id: app-a
    dependIds: ["ns"]  # Both execute in parallel

  - id: app-b
    dependIds: ["ns"]  # Both execute in parallel
```

### 3. Minimal Dependencies

Only specify necessary dependencies:

**Good:**
```yaml
- id: app
  dependIds: ["ns"]  # Only essential dependency
```

**Bad:**
```yaml
- id: app
  dependIds: ["ns", "unrelated-resource"]  # Unnecessary wait
```

## Debugging Dependencies

### Common Errors

**Cycle Detected:**
```
DependencyError: Dependency cycle detected: a -> b -> c -> a
```

**Solution:** Remove circular dependency.

**Missing Dependency:**
```
DependencyError: Resource 'app' depends on 'missing-id' which does not exist
```

**Solution:** Ensure all `dependIds` reference valid resource IDs.

**Timeout:**
```
ReadinessTimeout: Resource db not ready within 300s
```

**Solution:** Increase `timeoutSeconds` or set `waitForReady: false`.

## See Also

- [Template Guide](templates.md)
- [Policies Guide](policies.md)
- [Troubleshooting Guide](troubleshooting.md)
