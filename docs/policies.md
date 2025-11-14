# Policies Guide

Lynq provides fine-grained control over resource lifecycle through four policy types. This guide explains each policy and when to use them.

[[toc]]

## Policy Types Overview

| Policy | Controls | Default | Options |
|--------|----------|---------|---------|
| CreationPolicy | When resources are created | `WhenNeeded` | `Once`, `WhenNeeded` |
| DeletionPolicy | What happens on delete | `Delete` | `Delete`, `Retain` |
| ConflictPolicy | Ownership conflict handling | `Stuck` | `Stuck`, `Force` |
| PatchStrategy | How resources are updated | `apply` | `apply`, `merge`, `replace` |

::: tip Practical Examples
See [Policy Combinations Examples](policies-examples.md) for detailed real-world scenarios with diagrams and step-by-step explanations.
:::

::: tip Field-Level Control (v1.1.4+)
For fine-grained control over specific fields while using `WhenNeeded`, see [Field-Level Ignore Control](field-ignore.md). This allows you to selectively ignore certain fields during reconciliation (e.g., HPA-managed replicas).
:::

```mermaid
flowchart TD
    Start([LynqForm])
    Creation{{CreationPolicy}}
    Deletion{{DeletionPolicy}}
    Conflict{{ConflictPolicy}}
    Patch{{PatchStrategy}}
    Runtime[(Cluster Resources)]

    Start --> Creation --> Conflict --> Patch --> Runtime
    Creation -.->|Once| Runtime
    Creation -.->|WhenNeeded| Runtime

    Start --> Deletion --> Runtime
    Deletion -.->|Delete| Runtime
    Deletion -.->|Retain| Runtime

    Conflict -.->|"Stuck → Alert"| Runtime
    Conflict -.->|"Force → SSA force apply"| Runtime

    Patch -.->|apply| Runtime
    Patch -.->|merge| Runtime
    Patch -.->|replace| Runtime

    classDef decision fill:#fff3e0,stroke:#ffb74d,stroke-width:2px;
    classDef runtime fill:#e3f2fd,stroke:#64b5f6,stroke-width:2px;
    class Creation,Deletion,Conflict,Patch decision;
    class Runtime runtime;
```

## CreationPolicy

Controls when a resource is created or re-applied.

<CreationPolicyVisualizer />

### `WhenNeeded` (Default)

Resource is created and updated whenever the spec changes.

```yaml
deployments:
  - id: app
    creationPolicy: WhenNeeded  # Default
    nameTemplate: "{{ .uid }}-app"
    spec:
      # ... deployment spec
```

**Behavior:**
- ✅ Creates resource if it doesn't exist
- ✅ Updates resource when spec changes
- ✅ Re-applies if manually deleted
- ✅ Maintains desired state continuously

**Use when:**
- Resources should stay synchronized with templates
- You want drift correction
- Resource state should match database

**Example:** Application deployments, services, configmaps

::: tip Alternative: Use ignoreFields
If you need to update most fields but ignore specific ones (e.g., replicas controlled by HPA), consider using `creationPolicy: WhenNeeded` with `ignoreFields` instead of using `Once`. This provides more flexibility while still allowing selective field updates. See [Field-Level Ignore Control](field-ignore.md) for details.
:::

### `Once`

Resource is created only once and never updated, even if spec changes.

```yaml
jobs:
  - id: init-job
    creationPolicy: Once
    nameTemplate: "{{ .uid }}-init"
    spec:
      apiVersion: batch/v1
      kind: Job
      spec:
        template:
          spec:
            containers:
            - name: init
              image: busybox
              command: ["sh", "-c", "echo Initializing node {{ .uid }}"]
            restartPolicy: Never
```

**Behavior:**
- ✅ Creates resource on first reconciliation
- ❌ Never updates resource, even if template changes
- ✅ Skips if resource already exists with `lynq.sh/created-once` annotation
- ✅ Re-creates if manually deleted

**Use when:**
- One-time initialization tasks
- Security resources that shouldn't change
- Database migrations
- Initial setup jobs

**Example:** Init Jobs, security configurations, bootstrap scripts

**Annotation Added:**
```yaml
metadata:
  annotations:
    lynq.sh/created-once: "true"
```

## DeletionPolicy

Controls what happens to resources when a LynqNode CR is deleted.

<DeletionPolicyVisualizer />

### `Delete` (Default)

Resources are deleted when the Node is deleted via ownerReference.

```yaml
deployments:
  - id: app
    deletionPolicy: Delete  # Default
    nameTemplate: "{{ .uid }}-app"
    spec:
      # ... deployment spec
```

**Behavior:**
- ✅ Removes resource from cluster automatically
- ✅ Uses ownerReference for garbage collection
- ✅ No orphaned resources

**Use when:**
- Resources are node-specific and should be removed
- You want complete cleanup
- Resources have no value after node deletion

**Example:** Deployments, Services, ConfigMaps, Secrets

### `Retain`

Resources are kept in the cluster and **never have ownerReference set** (use label-based tracking instead).

```yaml
persistentVolumeClaims:
  - id: data-pvc
    deletionPolicy: Retain
    nameTemplate: "{{ .uid }}-data"
    spec:
      apiVersion: v1
      kind: PersistentVolumeClaim
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 10Gi
```

**Behavior:**
- ✅ **No ownerReference** (label-based tracking only)
- ✅ Resource stays in cluster even when Node is deleted
- ✅ Orphan labels added on deletion for identification
- ❌ No automatic cleanup by Kubernetes garbage collector
- ⚠️ Manual deletion required

**Why no ownerReference?**

Setting ownerReference would cause Kubernetes garbage collector to automatically delete the resource when the LynqNode CR is deleted, regardless of DeletionPolicy. The operator evaluates DeletionPolicy **at resource creation time** and uses label-based tracking (`lynq.sh/node`, `lynq.sh/node-namespace`) instead of ownerReference for Retain resources.

**Use when:**
- Data must survive node deletion
- Resources are expensive to recreate
- Regulatory/compliance requirements
- Debugging or forensics needed

**Example:** PersistentVolumeClaims, backup resources, audit logs

**Orphan Markers:**

When resources are retained (DeletionPolicy=Retain), they are automatically marked for easy identification:

```yaml
metadata:
  labels:
    lynq.sh/orphaned: "true"  # Label for selector queries
  annotations:
    lynq.sh/orphaned-at: "2025-01-15T10:30:00Z"  # RFC3339 timestamp
    lynq.sh/orphaned-reason: "RemovedFromTemplate"  # or "LynqNodeDeleted"
```

**Finding orphaned resources:**

```bash
# Find all orphaned resources
kubectl get all -A -l lynq.sh/orphaned=true

# Find resources orphaned due to template changes
kubectl get all -A -l lynq.sh/orphaned=true \
  -o jsonpath='{range .items[?(@.metadata.annotations.k8s-lynq\.org/orphaned-reason=="RemovedFromTemplate")]}{.kind}/{.metadata.name}{"\n"}{end}'

# Find resources orphaned due to node deletion
kubectl get all -A -l lynq.sh/orphaned=true \
  -o jsonpath='{range .items[?(@.metadata.annotations.k8s-lynq\.org/orphaned-reason=="LynqNodeDeleted")]}{.kind}/{.metadata.name}{"\n"}{end}'
```

### Orphan Resource Cleanup

::: tip Dynamic Template Evolution
DeletionPolicy applies not only when a LynqNode CR is deleted, but also when resources are **removed from the LynqForm**.
:::

**How it works:**

The operator tracks all applied resources in `status.appliedResources` with keys in format `kind/namespace/name@id`. During each reconciliation:

1. **Detect Orphans**: Compares current template resources with previously applied resources
2. **Respect Policy**: Applies the resource's `deletionPolicy` setting
3. **Update Status**: Tracks the new set of applied resources

**Orphan Lifecycle - Re-adoption:**

If you re-add a previously removed resource to the template, the operator automatically:
1. Removes all orphan markers (label + annotations)
2. Re-applies tracking labels or ownerReferences based on current DeletionPolicy
3. Resumes full management of the resource

This means you can safely experiment with template changes:
- Remove a resource → It becomes orphaned (if Retain policy)
- Re-add the same resource → It's cleanly re-adopted into management
- No manual cleanup or label management needed!

## Protecting LynqNodes from Cascade Deletion

::: danger Cascading deletions are immediate
Deleting a LynqHub or LynqForm cascades to all LynqNode CRs, which in turn deletes managed resources unless retention policies are set.
:::

### The Problem

```mermaid
flowchart TB
    Hub[LynqHub<br/>deleted] --> Nodes[LynqNode CRs<br/>finalizers trigger]
    Nodes --> Resources["Node Resources<br/>(Deployments, PVCs, ...)"]
    style Hub fill:#ffebee,stroke:#ef5350,stroke-width:2px;
    style Nodes fill:#fff3e0,stroke:#ffb74d,stroke-width:2px;
    style Resources fill:#f3e5f5,stroke:#ba68c8,stroke-width:2px;
```

### Recommended Solution: Use `Retain` DeletionPolicy

**Before deleting LynqHub or LynqForm**, ensure all resources in your templates use `deletionPolicy: Retain`:

```yaml
apiVersion: operator.lynq.sh/v1
kind: LynqForm
metadata:
  name: my-template
spec:
  hubId: my-hub

  # Set Retain for ALL resources
  deployments:
    - id: app
      deletionPolicy: Retain  # ✅ Keeps deployment
      nameTemplate: "{{ .uid }}-app"
      spec:
        # ... deployment spec

  services:
    - id: svc
      deletionPolicy: Retain  # ✅ Keeps service
      nameTemplate: "{{ .uid }}-svc"
      spec:
        # ... service spec

  persistentVolumeClaims:
    - id: data
      deletionPolicy: Retain  # ✅ Keeps PVC and data
      nameTemplate: "{{ .uid }}-data"
      spec:
        # ... PVC spec
```

### Why This Works

With `deletionPolicy: Retain`:
1. **At creation time**: Resources are created with label-based tracking only (NO ownerReference)
2. Even if LynqHub/LynqForm is deleted → LynqNode CRs are deleted
3. When LynqNode CRs are deleted → Resources stay in cluster (no ownerReference = no automatic deletion)
4. Finalizer adds orphan labels for easy identification
5. **Resources stay in the cluster** because Kubernetes garbage collector never marks them for deletion

**Key insight**: DeletionPolicy is evaluated when creating resources, not when deleting them. This prevents the Kubernetes garbage collector from auto-deleting Retain resources.

### When to Use This Strategy

✅ **Use `Retain` when:**
- You need to delete/recreate LynqHub for migration
- You're updating LynqForm with breaking changes
- You're testing hub configuration changes
- You have production LynqNodes that must not be interrupted
- You're performing maintenance on operator components

❌ **Don't use `Retain` when:**
- You actually want to clean up all node resources
- Testing in development environments
- You have backup/restore procedures in place

### Alternative: Update Instead of Delete

Instead of deleting and recreating, consider:

```bash
# ❌ DON'T: Delete and recreate (causes cascade deletion)
kubectl delete lynqhub my-hub
kubectl apply -f updated-hub.yaml

# ✅ DO: Update in place
kubectl apply -f updated-hub.yaml
```

## ConflictPolicy

Controls what happens when a resource already exists with a different owner or field manager.

<ConflictPolicyVisualizer />

### `Stuck` (Default)

Reconciliation stops if ownership conflict is detected.

```yaml
services:
  - id: app-svc
    conflictPolicy: Stuck  # Default
    nameTemplate: "{{ .uid }}-app"
    spec:
      # ... service spec
```

**Behavior:**
- ✅ Fails safe - doesn't overwrite existing resources
- ❌ Stops reconciliation on conflict
- 📢 Emits `ResourceConflict` event
- ⚠️ Marks Node as Degraded

**Use when:**
- Safety is paramount
- You want to investigate conflicts manually
- Resources might be managed by other controllers
- Default case (most conservative)

**Example:** Any resource where safety > availability

### `Force`

Attempts to take ownership using Server-Side Apply with `force=true`.

```yaml
deployments:
  - id: app
    conflictPolicy: Force
    nameTemplate: "{{ .uid }}-app"
    spec:
      # ... deployment spec
```

**Behavior:**
- ✅ Takes ownership forcefully
- ⚠️ May overwrite other controllers' changes
- ✅ Reconciliation continues
- 📢 Emits events on success/failure

**Use when:**
- Lynq should be the source of truth
- Conflicts are expected and acceptable
- You're migrating from another management system
- Availability > safety

**Example:** Resources exclusively managed by Lynq

**Warning:** This can override changes from other controllers or users!

## PatchStrategy

Controls how resources are updated.

### `apply` (Default - Server-Side Apply)

Uses Kubernetes Server-Side Apply for declarative updates.

```yaml
deployments:
  - id: app
    patchStrategy: apply  # Default
    nameTemplate: "{{ .uid }}-app"
    spec:
      # ... deployment spec
```

**Behavior:**
- ✅ Declarative updates
- ✅ Conflict detection
- ✅ Preserves fields managed by other controllers
- ✅ Field-level ownership tracking
- ✅ Most efficient

**Use when:**
- Multiple controllers manage the same resource
- You want Kubernetes-native updates
- Default case (best practice)

**Field Manager:** `lynq`

### `merge` (Strategic Merge Patch)

Uses strategic merge patch for updates.

```yaml
services:
  - id: app-svc
    patchStrategy: merge
    nameTemplate: "{{ .uid }}-app"
    spec:
      # ... service spec
```

**Behavior:**
- ✅ Merges changes with existing resource
- ✅ Preserves unspecified fields
- ⚠️ Less precise conflict detection
- ✅ Works with older Kubernetes versions

**Use when:**
- Partial updates needed
- Compatibility with legacy systems
- Strategic merge semantics preferred

### `replace` (Full Replacement)

Completely replaces the resource.

```yaml
configMaps:
  - id: config
    patchStrategy: replace
    nameTemplate: "{{ .uid }}-config"
    spec:
      # ... configmap spec
```

**Behavior:**
- ⚠️ Replaces entire resource
- ❌ Loses fields not in template
- ✅ Guarantees exact match
- ✅ Handles resourceVersion conflicts

**Use when:**
- Exact resource state required
- No other controllers manage the resource
- Complete replacement is intentional

**Warning:** This removes any fields not in your template!

## Default Values

If policies are not specified, these defaults apply:

```yaml
resources:
  - id: example
    creationPolicy: WhenNeeded  # ✅ Default
    deletionPolicy: Delete      # ✅ Default
    conflictPolicy: Stuck       # ✅ Default
    patchStrategy: apply        # ✅ Default
```

## Policy Decision Matrix

Recommended policy combinations by resource type:

| Resource Type | CreationPolicy | DeletionPolicy | ConflictPolicy | PatchStrategy |
|---------------|----------------|----------------|----------------|---------------|
| Deployment | WhenNeeded | Delete | Stuck | apply |
| Service | WhenNeeded | Delete | Stuck | apply |
| ConfigMap | WhenNeeded | Delete | Stuck | apply |
| Secret | WhenNeeded | Delete | Force | apply |
| PVC | Once | Retain | Stuck | apply |
| Init Job | Once | Delete | Force | replace |
| Namespace | WhenNeeded | Retain | Force | apply |
| Ingress | WhenNeeded | Delete | Stuck | apply |

::: tip See Detailed Examples
For in-depth explanations with diagrams and scenarios, see [Policy Combinations Examples](policies-examples.md).
:::

## Observability

### Events

Policies trigger various events:

```bash
# View Node events
kubectl describe lynqnode <lynqnode-name>
```

**Conflict Events:**
```
ResourceConflict: Resource conflict detected for default/acme-app (Kind: Deployment, Policy: Stuck)
```

**Deletion Events:**
```
LynqNodeDeleting: Deleting LynqNode 'acme-prod-template' (template: prod-template, uid: acme)
LynqNodeDeleted: Successfully deleted LynqNode 'acme-prod-template'
```

### Metrics

```promql
# Count apply attempts by policy
apply_attempts_total{kind="Deployment",result="success",conflict_policy="Stuck"}

# Track conflicts
lynqnode_conflicts_total{lynqnode="acme-web",conflict_policy="Stuck"}

# Failed reconciliations
lynqnode_reconcile_duration_seconds{result="error"}
```

See [Monitoring Guide](monitoring.md) for complete metrics reference.

## Troubleshooting

### Conflict Stuck

**Symptom:** LynqNode shows `Degraded` condition

**Cause:** Resource exists with different owner

**Solution:**
1. Check who owns the resource:
   ```bash
   kubectl get <resource-type> <resource-name> -o yaml | grep -A5 managedFields
   ```

2. Either:
   - Delete the conflicting resource
   - Change to `conflictPolicy: Force`
   - Use unique `nameTemplate`

### Resource Not Updating

**Symptom:** Changes to template don't apply

**Cause:** `creationPolicy: Once` is set

**Solution:**
- Change to `creationPolicy: WhenNeeded`, or
- Delete the resource to force recreation, or
- This is expected behavior for `Once` policy

### Resource Not Deleted

**Symptom:** Resource remains after LynqNode deletion

**Cause:** `deletionPolicy: Retain` is set

**Solution:**
- Manually delete: `kubectl delete <resource-type> <resource-name>`
- This is expected behavior for `Retain` policy

## See Also

- **[Policy Combinations Examples](policies-examples.md)** - Detailed real-world scenarios with diagrams
- [Field-Level Ignore Control](field-ignore.md) - Fine-grained field management
- [Template Guide](templates.md) - Template syntax and functions
- [Dependencies Guide](dependencies.md) - Resource ordering
- [Troubleshooting](troubleshooting.md) - Common issues
