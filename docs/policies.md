# Policies Guide

Tenant Operator provides fine-grained control over resource lifecycle through various policies. This guide explains each policy type and when to use them.

[[toc]]

## Policy Types Overview

| Policy | Controls | Default | Options |
|--------|----------|---------|---------|
| CreationPolicy | When resources are created | `WhenNeeded` | `Once`, `WhenNeeded` |
| DeletionPolicy | What happens on delete | `Delete` | `Delete`, `Retain` |
| ConflictPolicy | Ownership conflict handling | `Stuck` | `Stuck`, `Force` |
| PatchStrategy | How resources are updated | `apply` | `apply`, `merge`, `replace` |

::: tip New in v1.1.4: Field-Level Control
For fine-grained control over specific fields while using `WhenNeeded`, see [Field-Level Ignore Control](field-ignore.md). This allows you to selectively ignore certain fields during reconciliation (e.g., HPA-managed replicas).
:::

```mermaid
flowchart TD
    Start([Tenant Template])
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
              command: ["sh", "-c", "echo Initializing tenant {{ .uid }}"]
            restartPolicy: Never
```

**Behavior:**
- ✅ Creates resource on first reconciliation
- ❌ Never updates resource, even if template changes
- ✅ Skips if resource already exists with `kubernetes-tenants.org/created-once` annotation
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
    kubernetes-tenants.org/created-once: "true"
```

## DeletionPolicy

Controls what happens to resources when a Tenant CR is deleted.

### `Delete` (Default)

Resources are deleted when the Tenant is deleted.

```yaml
deployments:
  - id: app
    deletionPolicy: Delete  # Default
    nameTemplate: "{{ .uid }}-app"
    spec:
      # ... deployment spec
```

**Behavior:**
- ✅ Removes resource from cluster
- ✅ Cleans up automatically
- ✅ No orphaned resources

**Use when:**
- Resources are tenant-specific and should be removed
- You want complete cleanup
- Resources have no value after tenant deletion

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
- ✅ Resource stays in cluster even when Tenant is deleted
- ✅ Orphan labels added on deletion for identification
- ❌ No automatic cleanup by Kubernetes garbage collector
- ⚠️  Manual deletion required

**Why no ownerReference?**

Setting ownerReference would cause Kubernetes garbage collector to automatically delete the resource when the Tenant CR is deleted, regardless of DeletionPolicy. The operator evaluates DeletionPolicy **at resource creation time** and uses label-based tracking (`kubernetes-tenants.org/tenant`, `kubernetes-tenants.org/tenant-namespace`) instead of ownerReference for Retain resources.

**Use when:**
- Data must survive tenant deletion
- Resources are expensive to recreate
- Regulatory/compliance requirements
- Debugging or forensics needed

**Example:** PersistentVolumeClaims, backup resources, audit logs

::: details Advanced: Retain Lifecycle - Delete and Recreate Scenario

When using `DeletionPolicy: Retain`, understanding the delete-recreate lifecycle is crucial for managing resources correctly.

#### Scenario: Tenant Deletion and Recreation

**What happens when you delete a Tenant and then recreate it with the same UID?**

```mermaid
flowchart TD
    Start([Initial: Tenant Created])
    CreateResource[Create Resource<br/>+ Label tracking<br/>+ NO ownerReference<br/>DeletionPolicy: Retain]
    ResourceActive1[Resource Active<br/>Managed by Tenant]
    TenantDelete[User Deletes Tenant CR]
    FinalizerRun[Finalizer Executes]
    RemoveLabels[Remove Tracking Labels:<br/>- kubernetes-tenants.org/tenant<br/>- kubernetes-tenants.org/tenant-namespace]
    AddOrphanLabels[Add Orphan Markers:<br/>+ orphaned: true<br/>+ orphaned-at: timestamp<br/>+ orphaned-reason: TenantDeleted]
    ResourceOrphaned[Resource Orphaned<br/>Still in Cluster<br/>No Active Management]

    RecreateDecision{User Recreates<br/>Tenant with Same UID}
    CheckExists{Resource Exists?}
    CheckOrphan{Has Orphan Labels?}

    ReAdopt[Re-adoption:<br/>1. Remove Orphan Labels<br/>2. Add Tracking Labels<br/>3. Resume Management]
    ResourceActive2[Resource Active Again<br/>Managed by New Tenant]

    NoOrphanConflict{ConflictPolicy?}
    StuckError[ConflictPolicy: Stuck<br/>Mark Tenant Degraded<br/>Stop Reconciliation]
    ForceAdopt[ConflictPolicy: Force<br/>Force Take Ownership<br/>Add Tracking Labels]

    CreateNew[No Conflict:<br/>Add Tracking Labels<br/>Resume Management]

    Start --> CreateResource
    CreateResource --> ResourceActive1
    ResourceActive1 --> TenantDelete
    TenantDelete --> FinalizerRun
    FinalizerRun --> RemoveLabels
    RemoveLabels --> AddOrphanLabels
    AddOrphanLabels --> ResourceOrphaned

    ResourceOrphaned --> RecreateDecision
    RecreateDecision -->|Yes| CheckExists
    RecreateDecision -->|No| ResourceOrphaned

    CheckExists -->|Yes| CheckOrphan
    CheckExists -->|No| CreateResource

    CheckOrphan -->|Yes| ReAdopt
    CheckOrphan -->|No| NoOrphanConflict

    ReAdopt --> ResourceActive2

    NoOrphanConflict -->|Stuck| StuckError
    NoOrphanConflict -->|Force| ForceAdopt
    NoOrphanConflict -->|No Other Owner| CreateNew

    ForceAdopt --> ResourceActive2
    CreateNew --> ResourceActive2

    classDef createStyle fill:#e8f5e9,stroke:#4caf50,stroke-width:2px;
    classDef orphanStyle fill:#fff3e0,stroke:#ff9800,stroke-width:2px;
    classDef activeStyle fill:#e3f2fd,stroke:#2196f3,stroke-width:2px;
    classDef errorStyle fill:#ffebee,stroke:#f44336,stroke-width:2px;
    classDef decisionStyle fill:#f3e5f5,stroke:#ba68c8,stroke-width:2px;

    class CreateResource,ReAdopt createStyle;
    class ResourceOrphaned,AddOrphanLabels,RemoveLabels orphanStyle;
    class ResourceActive1,ResourceActive2,CreateNew activeStyle;
    class StuckError errorStyle;
    class RecreateDecision,CheckExists,CheckOrphan,NoOrphanConflict decisionStyle;
```

#### Key Points

**1. Initial Deletion (Tenant → Resource)**
- Tenant is deleted
- Finalizer removes tracking labels
- Orphan labels added:
  ```yaml
  labels:
    kubernetes-tenants.org/orphaned: "true"
  annotations:
    kubernetes-tenants.org/orphaned-at: "2025-01-15T10:30:00Z"
    kubernetes-tenants.org/orphaned-reason: "TenantDeleted"
  ```
- Resource stays in cluster (NO ownerReference means no GC deletion)

**2. Recreate with Same UID (Depends on CreationPolicy)**

::: tip CreationPolicy Matters!
Re-adoption behavior differs significantly between `Once` and `WhenNeeded`:
:::

**With `CreationPolicy: WhenNeeded` (automatic re-adoption):**
- Operator detects existing resource
- Finds orphan labels → automatic re-adoption
- Removes orphan markers
- Adds tracking labels back
- Resumes normal management
- ✅ **No data loss, seamless recovery**

**With `CreationPolicy: Once` (NO re-adoption):**
- Operator detects existing resource
- Finds `created-once` annotation → **SKIP** (continue)
- `ApplyResource` is **not called**
- Orphan markers **remain**
- Tracking labels **not re-added**
- ⚠️ **Resource counted as Ready but orphan markers persist**
- Manual intervention needed to remove orphan markers (see Example 1 details)

**3. Recreate with Different UID (Edge Case)**
- Operator tries to create resource with different name
- If nameTemplate uses UID, creates new resource
- Old orphaned resource remains
- Manual cleanup needed

**4. Resource Already Exists Without Orphan Labels (Conflict)**
- Another controller or user may have modified the resource
- Behavior depends on `ConflictPolicy`:
  - **Stuck**: Marks Tenant as Degraded, stops reconciliation
  - **Force**: Takes ownership with SSA force=true
  - **No owner**: Adds tracking labels, resumes management

#### Example Scenario

```bash
# 1. Initial tenant with PVC
kubectl apply -f - <<EOF
apiVersion: operator.kubernetes-tenants.org/v1
kind: Tenant
metadata:
  name: acme-web
spec:
  uid: acme
  persistentVolumeClaims:
    - id: data
      deletionPolicy: Retain
      nameTemplate: "{{ .uid }}-data"
      # ... spec
EOF

# PVC created: acme-data
# Labels: kubernetes-tenants.org/tenant: acme-web

# 2. Delete tenant
kubectl delete tenant acme-web

# PVC still exists: acme-data
# Labels changed to: kubernetes-tenants.org/orphaned: "true"
# Annotations: orphaned-at, orphaned-reason

# 3. Recreate tenant with same UID
kubectl apply -f - <<EOF
apiVersion: operator.kubernetes-tenants.org/v1
kind: Tenant
metadata:
  name: acme-web-v2
spec:
  uid: acme  # Same UID!
  persistentVolumeClaims:
    - id: data
      deletionPolicy: Retain
      nameTemplate: "{{ .uid }}-data"
      # ... spec
EOF

# Result: PVC acme-data is re-adopted
# - Orphan labels removed
# - Tracking labels restored
# - Data preserved
# - Management resumed
```

#### Benefits of Re-adoption

- ✅ **Zero data loss**: Existing data in PVC preserved
- ✅ **No downtime**: Pods can keep using the same PVC
- ✅ **Cost efficient**: No need to restore from backup
- ✅ **Automatic recovery**: No manual intervention needed
- ✅ **Audit trail**: Orphan timestamps show deletion history

#### When Re-adoption Fails

Re-adoption may fail if:

1. **Resource modified externally**:
   - Labels removed manually
   - Resource managed by another controller
   - Solution: Use `ConflictPolicy: Force` or clean up manually

2. **Different resource name**:
   - Changed nameTemplate
   - Different UID
   - Solution: Manually add tracking labels to old resource

3. **Resource type mismatch**:
   - Changed resource kind (ConfigMap → Secret)
   - Solution: Manual migration required

#### Best Practices

1. **Consistent UIDs**: Use stable UIDs for tenants to enable re-adoption
2. **Document orphaned resources**: Track why resources are orphaned
3. **Regular cleanup**: Periodically review and clean orphaned resources
4. **Test recovery**: Verify delete/recreate workflow in staging
5. **Monitor orphans**: Alert on orphaned resources older than N days

:::

### Orphan Resource Cleanup

::: tip Dynamic Template Evolution
DeletionPolicy applies not only when a Tenant CR is deleted, but also when resources are **removed from the TenantTemplate**.
:::

**How it works:**

The operator tracks all applied resources in `status.appliedResources` with keys in format `kind/namespace/name@id`. During each reconciliation:

1. **Detect Orphans**: Compares current template resources with previously applied resources
2. **Respect Policy**: Applies the resource's `deletionPolicy` setting:
   - `Delete`: Removes the orphaned resource from cluster (ownerReference enables automatic deletion)
   - `Retain`: Removes tracking labels and adds orphan labels, but keeps the resource
3. **Update Status**: Tracks the new set of applied resources

**Example scenario:**

```yaml
# Initial template
deployments:
  - id: web
    nameTemplate: "{{ .uid }}-web"
    deletionPolicy: Delete  # Will be removed when deleted from template
  - id: worker
    nameTemplate: "{{ .uid }}-worker"
    deletionPolicy: Retain  # Will be kept when deleted from template
```

After removing the `worker` deployment from template:
- `web` deployment: Still managed normally
- `worker` deployment: **Retained in cluster** (ownerReference removed, resource kept)

After removing the `web` deployment from template:
- `web` deployment: **Deleted from cluster** automatically

**Orphan Markers:**

When resources are retained (DeletionPolicy=Retain), they are automatically marked for easy identification:

```yaml
metadata:
  labels:
    kubernetes-tenants.org/orphaned: "true"  # Label for selector queries
  annotations:
    kubernetes-tenants.org/orphaned-at: "2025-01-15T10:30:00Z"  # RFC3339 timestamp
    kubernetes-tenants.org/orphaned-reason: "RemovedFromTemplate"
```

**Why split label/annotation?**
- **Label** `orphaned: "true"`: Simple boolean for selector queries (Kubernetes labels have strict RFC 1123 format requirements - no colons allowed in values)
- **Annotations**: Detailed metadata like timestamps (no format restrictions)

**Orphan Lifecycle - Re-adoption:**

If you re-add a previously removed resource to the template, the operator automatically:
1. Removes all orphan markers (label + annotations)
2. Re-applies tracking labels or ownerReferences based on current DeletionPolicy
3. Resumes full management of the resource

This means you can safely experiment with template changes:
- Remove a resource → It becomes orphaned (if Retain policy)
- Re-add the same resource → It's cleanly re-adopted into management
- No manual cleanup or label management needed!

**Finding orphaned resources:**

```bash
# Find all orphaned resources (using label selector)
kubectl get all -A -l kubernetes-tenants.org/orphaned=true

# Find resources orphaned due to template changes (filter by annotation)
kubectl get all -A -l kubernetes-tenants.org/orphaned=true -o jsonpath='{range .items[?(@.metadata.annotations.kubernetes-tenants\.org/orphaned-reason=="RemovedFromTemplate")]}{.kind}/{.metadata.name}{"\n"}{end}'

# Find resources orphaned due to tenant deletion (filter by annotation)
kubectl get all -A -l kubernetes-tenants.org/orphaned=true -o jsonpath='{range .items[?(@.metadata.annotations.kubernetes-tenants\.org/orphaned-reason=="TenantDeleted")]}{.kind}/{.metadata.name}{"\n"}{end}'
```

**Benefits:**
- ✅ Safe template evolution without manual cleanup
- ✅ No orphaned resources accumulation (Delete policy)
- ✅ Easy identification of retained orphans (Retain policy)
- ✅ DeletionPolicy consistency across all deletion scenarios
- ✅ Automatic detection during normal reconciliation
- ✅ Tracking of orphan timestamp and reason

## Protecting Tenants from Cascade Deletion

::: danger Cascading deletions are immediate
Deleting a TenantRegistry or TenantTemplate cascades to all Tenant CRs, which in turn deletes managed resources unless retention policies are set.
:::

### The Problem

```mermaid
flowchart TB
    Registry[TenantRegistry<br/>deleted] --> Tenants[Tenant CRs<br/>finalizers trigger]
    Tenants --> Resources["Tenant Resources<br/>(Deployments, PVCs, ...)"]
    style Registry fill:#ffebee,stroke:#ef5350,stroke-width:2px;
    style Tenants fill:#fff3e0,stroke:#ffb74d,stroke-width:2px;
    style Resources fill:#f3e5f5,stroke:#ba68c8,stroke-width:2px;
```

```mermaid
flowchart TB
    Template[TenantTemplate<br/>deleted] --> Tenants2[Tenant CRs<br/>cascade removed]
    Tenants2 --> Resources2["Tenant Resources<br/>(deleted unless retained)"]
    style Template fill:#ffebee,stroke:#ef5350,stroke-width:2px;
    style Tenants2 fill:#fff3e0,stroke:#ffb74d,stroke-width:2px;
    style Resources2 fill:#f3e5f5,stroke:#ba68c8,stroke-width:2px;
```

### Recommended Solution: Use `Retain` DeletionPolicy

**Before deleting TenantRegistry or TenantTemplate**, ensure all resources in your templates use `deletionPolicy: Retain`:

```mermaid
flowchart TB
    Delete["Delete TenantRegistry / Template"]
    TenantsRetain["Tenant CRs finalize"]
    Retained["Resources with Retain<br/>deletionPolicy"]
    Cleanup["Manual review / cleanup"]

    Delete --> TenantsRetain --> Retained --> Cleanup

    classDef safe fill:#e8f5e9,stroke:#81c784,stroke-width:2px;
    class Retained safe;
```

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: my-template
spec:
  registryId: my-registry

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
2. Even if TenantRegistry/TenantTemplate is deleted → Tenant CRs are deleted
3. When Tenant CRs are deleted → Resources stay in cluster (no ownerReference = no automatic deletion)
4. Finalizer adds orphan labels for easy identification
5. **Resources stay in the cluster** because Kubernetes garbage collector never marks them for deletion

**Key insight**: DeletionPolicy is evaluated when creating resources, not when deleting them. This prevents the Kubernetes garbage collector from auto-deleting Retain resources.

### When to Use This Strategy

✅ **Use `Retain` when:**
- You need to delete/recreate TenantRegistry for migration
- You're updating TenantTemplate with breaking changes
- You're testing registry configuration changes
- You have production tenants that must not be interrupted
- You're performing maintenance on operator components

❌ **Don't use `Retain` when:**
- You actually want to clean up all tenant resources
- Testing in development environments
- You have backup/restore procedures in place

### Alternative: Update Instead of Delete

Instead of deleting and recreating, consider:

```bash
# ❌ DON'T: Delete and recreate (causes cascade deletion)
kubectl delete tenantregistry my-registry
kubectl apply -f updated-registry.yaml

# ✅ DO: Update in place
kubectl apply -f updated-registry.yaml
```

### Recovery After Accidental Deletion

If you accidentally deleted TenantRegistry/TenantTemplate without `Retain`:

1. **Check if resources still exist:**
   ```bash
   kubectl get all -n <tenant-namespace>
   ```

2. **If deleted:** Resources are gone. You need to:
   - Restore from backups
   - Recreate TenantRegistry/TenantTemplate
   - Operator will recreate resources based on database

3. **If retained:** Manually clean up ownerless resources:
   ```bash
   # Find resources without owners
   kubectl get pods,svc,deploy -A -o json | \
     jq '.items[] | select(.metadata.ownerReferences == null) | .metadata.name'
   ```

### Best Practice Checklist

When planning to modify TenantRegistry or TenantTemplate:

- [ ] Review current `deletionPolicy` settings in all templates
- [ ] Set `deletionPolicy: Retain` for critical resources
- [ ] Test changes in non-production environment first
- [ ] Create backups of TenantRegistry/TenantTemplate YAML
- [ ] Document the change and expected impact
- [ ] Prefer `kubectl apply` (update) over delete/recreate
- [ ] Monitor Tenant CR status after changes

### Example: Safe Template Update

```bash
# 1. Backup current configuration
kubectl get tenanttemplate my-template -o yaml > backup-template.yaml
kubectl get tenantregistry my-registry -o yaml > backup-registry.yaml

# 2. Update template with Retain policies
cat <<EOF | kubectl apply -f -
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: my-template
spec:
  registryId: my-registry
  # Add deletionPolicy: Retain to all resources
  deployments:
    - id: app
      deletionPolicy: Retain
      # ... rest of config
EOF

# 3. Verify templates are updated
kubectl get tenanttemplate my-template -o yaml | grep deletionPolicy

# 4. Now safe to update registry if needed
kubectl apply -f updated-registry.yaml
```

## ConflictPolicy

Controls what happens when a resource already exists with a different owner or field manager.

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
- ⚠️  Marks Tenant as Degraded

**Use when:**
- Safety is paramount
- You want to investigate conflicts manually
- Resources might be managed by other controllers
- Default case (most conservative)

**Example:** Any resource where safety > availability

**Error Event:**
```
ResourceConflict: Resource conflict detected for default/acme-app (Kind: Deployment, Policy: Stuck).
Another controller or user may be managing this resource.
Consider using ConflictPolicy=Force to take ownership or resolve the conflict manually.
```

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
- ⚠️  May overwrite other controllers' changes
- ✅ Reconciliation continues
- 📢 Emits events on success/failure

**Use when:**
- Tenant Operator should be the source of truth
- Conflicts are expected and acceptable
- You're migrating from another management system
- Availability > safety

**Example:** Resources exclusively managed by Tenant Operator

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

**Field Manager:** `tenant-operator`

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
- ⚠️  Less precise conflict detection
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
- ⚠️  Replaces entire resource
- ❌ Loses fields not in template
- ✅ Guarantees exact match
- ✅ Handles resourceVersion conflicts

**Use when:**
- Exact resource state required
- No other controllers manage the resource
- Complete replacement is intentional

**Warning:** This removes any fields not in your template!

## Policy Combinations

Understanding how policies work together is crucial for designing reliable tenant resources. This section provides real-world examples with visual diagrams.

### Example 1: Stateful Data (PVC)

**Use Case:** Persistent storage that must survive tenant lifecycle changes and never lose data.

```yaml
persistentVolumeClaims:
  - id: data
    creationPolicy: Once        # Create only once
    deletionPolicy: Retain      # Keep data after tenant deletion
    conflictPolicy: Stuck       # Don't overwrite existing PVCs
    patchStrategy: apply        # Standard SSA
    nameTemplate: "{{ .uid }}-data"
    spec:
      # ... PVC spec
```

```mermaid
flowchart TD
    Start([Tenant Created])
    CheckExists{PVC Exists?}
    HasAnnotation{Has created-once<br/>annotation?}
    CreatePVC[Create PVC<br/>+ Add annotation<br/>+ Label tracking only<br/>NO ownerReference]
    SkipCreate[Skip Creation<br/>Count as Ready]
    TemplateChange[Template Updated]
    SkipUpdate[Skip Update<br/>CreationPolicy=Once]
    TenantDelete[Tenant Deleted]
    RemoveLabels[Remove Tracking Labels<br/>Add Orphan Labels]
    PVCRetained[PVC Retained in Cluster<br/>Data Preserved]
    ConflictDetect{PVC owned by<br/>another controller?}
    MarkDegraded[Mark Tenant Degraded<br/>Emit ResourceConflict Event]

    Start --> CheckExists
    CheckExists -->|No| ConflictDetect
    CheckExists -->|Yes| HasAnnotation

    ConflictDetect -->|No| CreatePVC
    ConflictDetect -->|Yes| MarkDegraded

    HasAnnotation -->|Yes| SkipCreate
    HasAnnotation -->|No| CreatePVC

    CreatePVC --> TemplateChange
    SkipCreate --> TemplateChange
    TemplateChange --> SkipUpdate

    SkipUpdate --> TenantDelete
    TenantDelete --> RemoveLabels
    RemoveLabels --> PVCRetained

    classDef createStyle fill:#e8f5e9,stroke:#4caf50,stroke-width:2px;
    classDef skipStyle fill:#fff3e0,stroke:#ff9800,stroke-width:2px;
    classDef retainStyle fill:#e3f2fd,stroke:#2196f3,stroke-width:2px;
    classDef errorStyle fill:#ffebee,stroke:#f44336,stroke-width:2px;

    class CreatePVC createStyle;
    class SkipCreate,SkipUpdate skipStyle;
    class PVCRetained,RemoveLabels retainStyle;
    class MarkDegraded errorStyle;
```

**Rationale:**
- `Once`: PVC spec shouldn't change (size immutable in many storage classes)
- `Retain`: Data survives tenant deletion - **NO ownerReference** set to prevent automatic deletion
- `Stuck`: Safety - don't overwrite someone else's PVC **on initial creation only**
- `apply`: Standard SSA for declarative management

**Key Behavior:**
- ✅ Created once, never updated (even if template changes)
- ✅ Survives tenant deletion (label-based tracking)
- ✅ Safe conflict detection **on initial creation** (skipped on subsequent reconciliations)
- 📊 Data persists indefinitely
- ⚠️ **Important**: Once `created-once` annotation is set, `ApplyResource` is never called again
  - No conflict checks on re-reconciliation
  - No orphan marker cleanup on Tenant recreation
  - Resource is "fire-and-forget" (see detailed explanation below)

::: details What happens if I delete and recreate the Tenant?

::: warning CreationPolicy=Once Limitation
With `CreationPolicy: Once`, the operator **SKIPS** resources that have the `created-once` annotation. This means:
- **NO re-adoption** occurs on Tenant recreation
- **Orphan markers remain** on the resource
- **NO conflict detection** (ApplyResource is not called)
- Resource is **counted as Ready** but not actively managed

This is the trade-off of using `Once` - resources are truly immutable and "fire-and-forget".
:::

When you delete a Tenant with `DeletionPolicy: Retain` + `CreationPolicy: Once` and recreate it, the behavior is different from `WhenNeeded`:

**Scenario Timeline:**

```mermaid
sequenceDiagram
    participant User
    participant Tenant as Tenant CR
    participant Operator
    participant PVC as PVC (acme-data)

    Note over Tenant,PVC: Phase 1: Initial State
    User->>Tenant: Create Tenant (uid: acme)
    Tenant->>Operator: Reconcile
    Operator->>PVC: Create PVC<br/>+ created-once: true<br/>+ NO ownerReference
    Note over PVC: Active & Managed Once

    Note over Tenant,PVC: Phase 2: Deletion
    User->>Tenant: Delete Tenant
    Tenant->>Operator: Finalizer runs
    Operator->>PVC: Add orphan labels<br/>(created-once: true REMAINS)
    Note over PVC: Orphaned but exists<br/>orphaned: true<br/>created-once: true

    Note over Tenant,PVC: Phase 3: Recreation (Same UID)
    User->>Tenant: Create Tenant (uid: acme)
    Tenant->>Operator: Reconcile
    Operator->>PVC: Check exists & has created-once?
    PVC-->>Operator: Yes, has created-once
    Operator->>Operator: SKIP (continue)<br/>Count as Ready<br/>NO ApplyResource call
    Note over PVC: STILL ORPHANED<br/>orphaned: true remains<br/>created-once: true remains
```

**Step-by-Step:**

1. **Initial Creation**
   ```yaml
   # PVC created: acme-data
   metadata:
     labels:
       kubernetes-tenants.org/tenant: acme-web
       kubernetes-tenants.org/tenant-namespace: default
     annotations:
       kubernetes-tenants.org/created-once: "true"  # ← CreationPolicy marker
   ```

2. **After Tenant Deletion**
   ```yaml
   # PVC still exists: acme-data
   metadata:
     labels:
       kubernetes-tenants.org/orphaned: "true"
     annotations:
       kubernetes-tenants.org/created-once: "true"  # ← STILL PRESENT
       kubernetes-tenants.org/orphaned-at: "2025-01-15T10:30:00Z"
       kubernetes-tenants.org/orphaned-reason: "TenantDeleted"
   ```

3. **After Tenant Recreation (same UID)**
   ```yaml
   # PVC exists but NOT re-adopted due to created-once annotation
   metadata:
     labels:
       kubernetes-tenants.org/orphaned: "true"  # ← REMAINS (not removed!)
     annotations:
       kubernetes-tenants.org/created-once: "true"  # ← Causes skip
       kubernetes-tenants.org/orphaned-at: "2025-01-15T10:30:00Z"
       kubernetes-tenants.org/orphaned-reason: "TenantDeleted"  # ← REMAINS
   ```

**Result:**
- ✅ PVC data preserved
- ✅ Zero downtime
- ⚠️ **NO automatic recovery** - orphan markers remain
- 📊 Tenant considers PVC as "Ready" but doesn't actively manage it

**Why This Happens:**

`CreationPolicy: Once` behavior (from code `tenant_controller.go:332-336`):
```go
if exists && hasAnnotation {  // hasAnnotation = has "created-once"
    // Resource already created with Once policy, skip
    readyCount++ // Count as ready since it exists
    continue     // ← ApplyResource NOT CALLED
}
```

Since `ApplyResource` is not called:
- Orphan markers are **not removed** (removal happens in `ApplyResource`)
- Tracking labels are **not re-added**
- Conflict checks **don't occur**

**Manual Recovery:**

If you need to "adopt" the orphaned resource:

```bash
# Remove the created-once annotation to allow re-adoption
kubectl annotate pvc acme-data kubernetes-tenants.org/created-once-

# Next reconciliation will call ApplyResource and remove orphan markers
```

**Comparison: Once vs WhenNeeded:**

| Aspect | Once + Retain | WhenNeeded + Retain |
|--------|---------------|---------------------|
| Re-adoption | ❌ No (skipped) | ✅ Yes (automatic) |
| Orphan markers removed | ❌ No | ✅ Yes |
| Tracking labels re-added | ❌ No | ✅ Yes |
| Conflict check | ❌ No (not reached) | ✅ Yes |
| Active management | ❌ No | ✅ Yes |

**When to use Once + Retain:**
- PVCs with immutable size
- One-time initialization that must never change
- Resources that should truly be "fire-and-forget"
- You accept manual cleanup of orphan markers

**When NOT to use Once + Retain:**
- Resources that need re-adoption on recreate
- Resources that require active management
- Use `WhenNeeded + Retain` instead for automatic recovery

See [Example 4](#example-4-shared-infrastructure) for `WhenNeeded + Retain` with automatic re-adoption.

### Example 2: Init Job

**Use Case:** One-time initialization task that runs once per tenant and cleans up after tenant deletion.

```yaml
jobs:
  - id: init
    creationPolicy: Once        # Run only once
    deletionPolicy: Delete      # Clean up after tenant deletion
    conflictPolicy: Force       # Re-create if needed
    patchStrategy: replace      # Exact job spec
    nameTemplate: "{{ .uid }}-init"
    spec:
      # ... job spec
```

```mermaid
flowchart TD
    Start([Tenant Created])
    CheckExists{Job Exists?}
    HasAnnotation{Has created-once<br/>annotation?}
    CheckConflict{Job owned by<br/>another controller?}
    ForceApply[Force Take Ownership<br/>SSA with force=true]
    CreateJob[Create Job<br/>+ Add annotation<br/>+ ownerReference]
    SkipCreate[Skip Creation<br/>Job Already Completed]
    RunJob[Job Executes Once]
    TemplateChange[Template Updated]
    SkipUpdate[Skip Update<br/>CreationPolicy=Once<br/>Job keeps running]
    ManualDelete[User Manually<br/>Deletes Job]
    RecreateJob[Recreate Job<br/>on Next Reconcile]
    TenantDelete[Tenant Deleted]
    AutoDelete[Kubernetes GC<br/>Deletes Job<br/>via ownerReference]
    Cleanup[Cleanup Complete]

    Start --> CheckExists
    CheckExists -->|No| CreateJob
    CheckExists -->|Yes| HasAnnotation

    HasAnnotation -->|Yes| SkipCreate
    HasAnnotation -->|No| CheckConflict

    CheckConflict -->|Yes| ForceApply
    CheckConflict -->|No| CreateJob

    CreateJob --> RunJob
    ForceApply --> RunJob
    SkipCreate --> TemplateChange

    RunJob --> TemplateChange
    TemplateChange --> SkipUpdate
    SkipUpdate --> ManualDelete
    ManualDelete --> RecreateJob
    RecreateJob --> TenantDelete

    SkipUpdate --> TenantDelete
    TenantDelete --> AutoDelete
    AutoDelete --> Cleanup

    classDef createStyle fill:#e8f5e9,stroke:#4caf50,stroke-width:2px;
    classDef skipStyle fill:#fff3e0,stroke:#ff9800,stroke-width:2px;
    classDef deleteStyle fill:#ffebee,stroke:#f44336,stroke-width:2px;
    classDef forceStyle fill:#fce4ec,stroke:#e91e63,stroke-width:2px;

    class CreateJob,RecreateJob createStyle;
    class SkipCreate,SkipUpdate skipStyle;
    class AutoDelete,Cleanup deleteStyle;
    class ForceApply forceStyle;
```

**Rationale:**
- `Once`: Initialization runs only once - even if template changes, job won't re-run
- `Delete`: No need to keep job history after tenant deletion
- `Force`: Operator owns this resource exclusively - takes ownership if conflict
- `replace`: Ensures exact job spec match

**Key Behavior:**
- ✅ Runs once per tenant lifetime
- ✅ Automatically cleaned up on tenant deletion
- ✅ Force takes ownership from conflicts
- 🔄 Re-creates if manually deleted (but still runs only once)

### Example 3: Application Deployment

**Use Case:** Main application that should stay synchronized with template changes and clean up completely on deletion.

```yaml
deployments:
  - id: app
    creationPolicy: WhenNeeded  # Keep updated
    deletionPolicy: Delete      # Clean up on deletion
    conflictPolicy: Stuck       # Safe default
    patchStrategy: apply        # Kubernetes best practice
    nameTemplate: "{{ .uid }}-app"
    spec:
      # ... deployment spec
```

```mermaid
flowchart TD
    Start([Tenant Created])
    CheckExists{Deployment<br/>Exists?}
    CheckConflict{Owned by another<br/>controller?}
    MarkDegraded[Mark Tenant Degraded<br/>Stop Reconciliation<br/>Emit ResourceConflict]
    CreateDeploy[Create Deployment<br/>SSA with fieldManager<br/>+ ownerReference]
    DeployRunning[Deployment Running]
    TemplateChange[Template Updated<br/>DB Data Changed]
    ApplyUpdate[Apply Changes<br/>SSA updates only<br/>managed fields]
    DriftDetect[Manual Change<br/>Detected]
    AutoCorrect[Auto-Correct Drift<br/>Revert to desired state]
    TenantDelete[Tenant Deleted]
    AutoDelete[Kubernetes GC<br/>Deletes Deployment<br/>+ Pods + ReplicaSets]
    Cleanup[Complete Cleanup]

    Start --> CheckExists
    CheckExists -->|No| CreateDeploy
    CheckExists -->|Yes| CheckConflict

    CheckConflict -->|Yes| MarkDegraded
    CheckConflict -->|No| DeployRunning

    CreateDeploy --> DeployRunning
    DeployRunning --> TemplateChange
    TemplateChange --> ApplyUpdate
    ApplyUpdate --> DeployRunning

    DeployRunning --> DriftDetect
    DriftDetect --> AutoCorrect
    AutoCorrect --> DeployRunning

    DeployRunning --> TenantDelete
    TenantDelete --> AutoDelete
    AutoDelete --> Cleanup

    classDef createStyle fill:#e8f5e9,stroke:#4caf50,stroke-width:2px;
    classDef updateStyle fill:#e3f2fd,stroke:#2196f3,stroke-width:2px;
    classDef deleteStyle fill:#ffebee,stroke:#f44336,stroke-width:2px;
    classDef errorStyle fill:#fce4ec,stroke:#e91e63,stroke-width:2px;

    class CreateDeploy createStyle;
    class ApplyUpdate,AutoCorrect updateStyle;
    class AutoDelete,Cleanup deleteStyle;
    class MarkDegraded errorStyle;
```

**Rationale:**
- `WhenNeeded`: Always keep deployment in sync with template and database
- `Delete`: Standard cleanup via ownerReference
- `Stuck`: Safe default - investigate conflicts rather than force override
- `apply`: SSA best practice - preserves fields from other controllers (e.g., HPA)

**Key Behavior:**
- ✅ Continuously synchronized with template
- ✅ Auto-corrects manual drift
- ✅ Plays well with other controllers (HPA, VPA)
- ✅ Complete cleanup on deletion
- ⚠️  Stops on conflicts for safety

### Example 4: Shared Infrastructure

**Use Case:** Configuration data that should stay updated but survive tenant deletion for debugging or shared resource references.

```yaml
configMaps:
  - id: shared-config
    creationPolicy: WhenNeeded  # Maintain config
    deletionPolicy: Retain      # Keep config for investigation
    conflictPolicy: Force       # Operator manages configs
    patchStrategy: apply        # SSA
    nameTemplate: "{{ .uid }}-shared-config"
    spec:
      # ... configmap spec
```

```mermaid
flowchart TD
    Start([Tenant Created])
    CheckExists{ConfigMap<br/>Exists?}
    CheckConflict{Owned by another<br/>controller?}
    ForceTake[Force Take Ownership<br/>SSA with force=true<br/>+ Label tracking only<br/>NO ownerReference]
    CreateCM[Create ConfigMap<br/>SSA apply<br/>+ Label tracking only<br/>NO ownerReference]
    CMActive[ConfigMap Active]
    TemplateChange[Template Updated<br/>DB Data Changed]
    ApplyUpdate[Apply Changes<br/>SSA updates config data<br/>Force if conflict]
    OtherPodRef[Other Pods/Services<br/>Reference ConfigMap]
    TenantDelete[Tenant Deleted]
    RemoveLabels[Remove Tracking Labels<br/>Add Orphan Labels<br/>+ Timestamp + Reason]
    CMRetained[ConfigMap Retained<br/>Available for Investigation<br/>or Shared Use]

    Start --> CheckExists
    CheckExists -->|No| CreateCM
    CheckExists -->|Yes| CheckConflict

    CheckConflict -->|Yes| ForceTake
    CheckConflict -->|No| CMActive

    CreateCM --> CMActive
    ForceTake --> CMActive

    CMActive --> TemplateChange
    TemplateChange --> ApplyUpdate
    ApplyUpdate --> CMActive

    CMActive --> OtherPodRef
    OtherPodRef --> CMActive

    CMActive --> TenantDelete
    TenantDelete --> RemoveLabels
    RemoveLabels --> CMRetained

    classDef createStyle fill:#e8f5e9,stroke:#4caf50,stroke-width:2px;
    classDef updateStyle fill:#e3f2fd,stroke:#2196f3,stroke-width:2px;
    classDef retainStyle fill:#fff3e0,stroke:#ff9800,stroke-width:2px;
    classDef forceStyle fill:#fce4ec,stroke:#e91e63,stroke-width:2px;

    class CreateCM createStyle;
    class ApplyUpdate updateStyle;
    class RemoveLabels,CMRetained retainStyle;
    class ForceTake forceStyle;
```

**Rationale:**
- `WhenNeeded`: Keep configmap data updated as template/database changes
- `Retain`: ConfigMap might be referenced by other resources or needed for debugging - **NO ownerReference** to prevent deletion
- `Force`: Operator is authoritative for this config - takes ownership if conflict exists
- `apply`: SSA for declarative configuration management

**Key Behavior:**
- ✅ Continuously synchronized with changes
- ✅ Force takes ownership from conflicts
- ✅ Survives tenant deletion (label-based tracking)
- 📊 Available for investigation post-deletion
- 🔗 Can be referenced by non-tenant resources

**Common Scenarios:**
- Debugging tenant issues after deletion
- Shared configuration referenced by multiple systems
- Compliance/audit requirements
- Migration to new tenant system

::: details What happens with WhenNeeded + Retain on delete/recreate?

Unlike `CreationPolicy: Once`, resources with `WhenNeeded` continue to update but can still be retained and re-adopted:

**Key Difference from Example 1 (PVC):**

| Aspect | Example 1 (PVC)<br/>Once + Retain | Example 4 (ConfigMap)<br/>WhenNeeded + Retain |
|--------|-----------------------------------|-----------------------------------------------|
| **Updates** | 🚫 Never (frozen after creation) | ✅ Always (syncs with template) |
| **Retention** | ✅ Yes (orphaned on delete) | ✅ Yes (orphaned on delete) |
| **Re-adoption** | ✅ Yes (if same UID) | ✅ Yes (if same UID) |
| **Force Ownership** | ❌ No (Stuck policy) | ✅ Yes (Force policy) |

**Scenario Timeline:**

```mermaid
sequenceDiagram
    participant User
    participant Tenant as Tenant CR
    participant Operator
    participant ConfigMap as ConfigMap<br/>(acme-shared-config)

    Note over Tenant,ConfigMap: Phase 1: Active Updates
    User->>Tenant: Create Tenant (uid: acme)
    Tenant->>Operator: Reconcile
    Operator->>ConfigMap: Create ConfigMap<br/>Labels: tenant=acme-web<br/>NO ownerReference
    Note over ConfigMap: Active & Managed<br/>Syncs with template

    User->>Tenant: Update Template<br/>(change config data)
    Tenant->>Operator: Reconcile
    Operator->>ConfigMap: Apply Updates<br/>Force if conflict
    Note over ConfigMap: Updated with new data

    Note over Tenant,ConfigMap: Phase 2: Deletion & Retention
    User->>Tenant: Delete Tenant
    Tenant->>Operator: Finalizer runs
    Operator->>ConfigMap: Remove tracking labels<br/>Add orphan labels
    Note over ConfigMap: Orphaned but exists<br/>Last data preserved

    Note over Tenant,ConfigMap: Phase 3: Re-adoption & Resume Updates
    User->>Tenant: Create Tenant (uid: acme)
    Tenant->>Operator: Reconcile
    Operator->>ConfigMap: Check exists & orphan?
    ConfigMap-->>Operator: Yes, found orphan
    Operator->>ConfigMap: Re-adopt + Apply Updates
    Note over ConfigMap: Active & Managed again<br/>Updates resume

    User->>Tenant: Update Template<br/>(more changes)
    Tenant->>Operator: Reconcile
    Operator->>ConfigMap: Apply Updates
    Note over ConfigMap: Syncs continuously
```

**Step-by-Step:**

1. **Initial Creation + Updates**
   ```yaml
   # ConfigMap updates continuously
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: acme-shared-config
     labels:
       kubernetes-tenants.org/tenant: acme-web
   data:
     config.json: '{"version": "1.0"}'  # Syncs with template
   ```

2. **After Tenant Deletion**
   ```yaml
   # ConfigMap retained with last state
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: acme-shared-config
     labels:
       kubernetes-tenants.org/orphaned: "true"
     annotations:
       kubernetes-tenants.org/orphaned-at: "2025-01-15T10:30:00Z"
       kubernetes-tenants.org/orphaned-reason: "TenantDeleted"
   data:
     config.json: '{"version": "1.0"}'  # Frozen at deletion time
   ```

3. **After Tenant Recreation**
   ```yaml
   # ConfigMap re-adopted and updates resume
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: acme-shared-config
     labels:
       kubernetes-tenants.org/tenant: acme-web-v2  # New tenant
       # orphaned label removed
   data:
     config.json: '{"version": "2.0"}'  # Updated to latest template
   ```

**Benefits:**

- ✅ Continuous synchronization while tenant exists
- ✅ Data preserved during tenant absence
- ✅ Automatic catch-up on re-adoption (applies latest template)
- ✅ Force policy ensures successful re-adoption even with conflicts
- 📊 Great for debugging (can see last known good config)

**Use Cases:**

1. **Temporary Tenant Removal**: Delete tenant for maintenance, recreate later with same config
2. **Blue-Green Deployments**: Switch between tenant versions while preserving config
3. **Disaster Recovery**: Recreate tenant after failure with preserved configuration
4. **Migration Testing**: Delete test tenant, verify config remains, recreate

**Contrast with WhenNeeded + Delete:**

```mermaid
flowchart LR
    subgraph Retain["WhenNeeded + Retain"]
        R1[Template Change] -->|Updates| R2[ConfigMap]
        R2 -->|Tenant Delete| R3[Orphaned ConfigMap]
        R3 -->|Tenant Recreate| R2
    end

    subgraph Delete["WhenNeeded + Delete"]
        D1[Template Change] -->|Updates| D2[ConfigMap]
        D2 -->|Tenant Delete| D3[ConfigMap Deleted]
        D3 -->|Tenant Recreate| D4[New ConfigMap]
    end

    style R3 fill:#fff3e0,stroke:#ff9800
    style D3 fill:#ffebee,stroke:#f44336
```

**Decision Guide:**

- Use `WhenNeeded + Retain` when: Config should survive tenant lifecycle
- Use `WhenNeeded + Delete` when: Config is tenant-specific and disposable

:::

### Policy Combinations Summary

Quick reference comparing all four examples:

| Aspect | PVC (Stateful) | Init Job | App Deployment | Shared Config |
|--------|----------------|----------|----------------|---------------|
| **CreationPolicy** | `Once` | `Once` | `WhenNeeded` | `WhenNeeded` |
| **DeletionPolicy** | `Retain` | `Delete` | `Delete` | `Retain` |
| **ConflictPolicy** | `Stuck` | `Force` | `Stuck` | `Force` |
| **PatchStrategy** | `apply` | `replace` | `apply` | `apply` |
| **ownerReference** | ❌ No | ✅ Yes | ✅ Yes | ❌ No |
| **Updates** | 🚫 Never | 🚫 Never | ✅ Always | ✅ Always |
| **Survives Deletion** | ✅ Yes | ❌ No | ❌ No | ✅ Yes |
| **Auto-Cleanup** | ❌ Manual | ✅ Auto (GC) | ✅ Auto (GC) | ❌ Manual |
| **Drift Correction** | N/A (Once) | N/A (Once) | ✅ Yes | ✅ Yes |
| **Conflict Handling** | ⚠️ Stop | 💪 Force | ⚠️ Stop | 💪 Force |

**Legend:**
- ✅ Enabled / Yes
- ❌ Disabled / No
- 🚫 Never updates
- ⚠️ Safe mode (stops on conflict)
- 💪 Aggressive (forces ownership)
- N/A: Not applicable

**Choosing the Right Combination:**

```mermaid
flowchart TD
    Start([Choose Policy Combination])
    Q1{Resource holds<br/>persistent data?}
    Q2{Needs continuous<br/>updates?}
    Q3{Runs only once?}
    Q4{Should survive<br/>tenant deletion?}
    Q5{Conflict<br/>tolerance?}

    Result1[Example 1: PVC<br/>Once + Retain + Stuck]
    Result2[Example 2: Init Job<br/>Once + Delete + Force]
    Result3[Example 3: App Deployment<br/>WhenNeeded + Delete + Stuck]
    Result4[Example 4: Shared Config<br/>WhenNeeded + Retain + Force]

    Start --> Q1
    Q1 -->|Yes| Q4
    Q1 -->|No| Q2

    Q4 -->|Yes| Result1
    Q4 -->|No| Q3

    Q3 -->|Yes| Q5
    Q3 -->|No| Q2

    Q5 -->|Force| Result2
    Q5 -->|Stuck| Result1

    Q2 -->|Yes| Q4
    Q2 -->|No| Q3

    Q4 -->|Yes| Result4
    Q4 -->|No| Result3

    classDef decisionStyle fill:#fff3e0,stroke:#ff9800,stroke-width:2px;
    classDef resultStyle fill:#e8f5e9,stroke:#4caf50,stroke-width:2px;

    class Q1,Q2,Q3,Q4,Q5 decisionStyle;
    class Result1,Result2,Result3,Result4 resultStyle;
```

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

## Observability

### Events

Policies trigger various events:

```bash
# View Tenant events
kubectl describe tenant <tenant-name>
```

**Conflict Events:**
```
ResourceConflict: Resource conflict detected for default/acme-app (Kind: Deployment, Policy: Stuck)
```

**Deletion Events:**
```
TenantDeleting: Deleting Tenant 'acme-prod-template' (template: prod-template, uid: acme)
TenantDeleted: Successfully deleted Tenant 'acme-prod-template'
```

### Metrics

```promql
# Count apply attempts by policy
apply_attempts_total{kind="Deployment",result="success",conflict_policy="Stuck"}

# Failed reconciliations
tenant_reconcile_duration_seconds{result="error"}
```

## Troubleshooting

### Conflict Stuck

**Symptom:** Tenant shows `Degraded` condition

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

**Symptom:** Resource remains after Tenant deletion

**Cause:** `deletionPolicy: Retain` is set

**Solution:**
- Manually delete: `kubectl delete <resource-type> <resource-name>`
- This is expected behavior for `Retain` policy

## See Also

- [Template Guide](templates.md) - Template syntax and functions
- [Dependencies Guide](dependencies.md) - Resource ordering
- [Troubleshooting](troubleshooting.md) - Common issues
