# API Reference

Complete API reference for Lynq CRDs.

[[toc]]

## LynqHub

Defines external data source and sync configuration.

::: info Resource metadata
- **Kind:** `LynqHub`
- **API Version:** `operator.lynq.sh/v1`
:::

### Spec

```yaml
apiVersion: operator.lynq.sh/v1
kind: LynqHub
metadata:
  name: my-registry
spec:
  # Data source configuration
  source:
    type: mysql                      # mysql (postgresql planned for v1.2)
    mysql:
      host: string                   # Database host (required)
      port: int                      # Database port (default: 3306)
      username: string               # Database username (required)
      passwordRef:                   # Password secret reference (required)
        name: string                 # Secret name
        key: string                  # Secret key
      database: string               # Database name (required)
      table: string                  # Table name (required)
    syncInterval: duration           # Sync interval (required, e.g., "1m")
  
  # Required column mappings
  valueMappings:
    uid: string                      # Tenant ID column (required)
    hostOrUrl: string                # Tenant URL column (required)
    activate: string                 # Activation flag column (required)
  
  # Optional column mappings
  extraValueMappings:
    key: value                       # Additional column mappings (optional)
```

### Status

```yaml
status:
  observedGeneration: int64          # Last observed generation
  referencingTemplates: int32        # Number of templates referencing this registry
  desired: int32                     # Desired LynqNode CRs (templates × rows)
  ready: int32                       # Ready LynqNode CRs
  failed: int32                      # Failed LynqNode CRs
  conditions:                        # Status conditions
  - type: Ready
    status: "True"
    reason: SyncSucceeded
    message: "Successfully synced N nodes"
    lastTransitionTime: timestamp
```

## LynqForm

Defines resource blueprint for nodes.

::: info Resource metadata
- **Kind:** `LynqForm`
- **API Version:** `operator.lynq.sh/v1`
:::

### Spec

```yaml
apiVersion: operator.lynq.sh/v1
kind: LynqForm
metadata:
  name: my-template
spec:
  registryId: string                 # LynqHub name (required)

  # Resource arrays
  serviceAccounts: []TResource
  deployments: []TResource
  statefulSets: []TResource
  daemonSets: []TResource
  services: []TResource
  configMaps: []TResource
  secrets: []TResource
  persistentVolumeClaims: []TResource
  jobs: []TResource
  cronJobs: []TResource
  ingresses: []TResource
  podDisruptionBudgets: []TResource  # PodDisruptionBudget resources
  networkPolicies: []TResource       # NetworkPolicy resources
  horizontalPodAutoscalers: []TResource  # HorizontalPodAutoscaler resources
  manifests: []TResource             # Raw unstructured resources
```

### TResource Structure

```yaml
id: string                           # Unique resource ID (required)
nameTemplate: string                 # Go template for resource name (required)
labelsTemplate:                      # Template-enabled labels (optional)
  key: value
annotationsTemplate:                 # Template-enabled annotations (optional)
  key: value
dependIds: []string                  # Dependency IDs (optional)
creationPolicy: string               # Once | WhenNeeded (default: WhenNeeded)
deletionPolicy: string               # Delete | Retain (default: Delete)
conflictPolicy: string               # Stuck | Force (default: Stuck)
patchStrategy: string                # apply | merge | replace (default: apply)
waitForReady: bool                   # Wait for resource ready (default: true)
timeoutSeconds: int32                # Readiness timeout (default: 300, max: 3600)
spec: object                         # Kubernetes resource spec (required)
```

### Status

```yaml
status:
  observedGeneration: int64
  validationErrors: []string         # Template validation errors
  totalNodes: int32                # Total nodes using this template
  readyNodes: int32                # Ready nodes
  conditions:
  - type: Valid
    status: "True"
    reason: ValidationSucceeded
```

## LynqNode

Represents a single tenant instance.

::: info Resource metadata
- **Kind:** `LynqNode`
- **API Version:** `operator.lynq.sh/v1`
:::

::: warning Managed resource
LynqNode objects are typically managed by the operator and rarely created manually.
:::

### Spec

```yaml
apiVersion: operator.lynq.sh/v1
kind: LynqNode
metadata:
  name: acme-prod-template
  annotations:
    # Template variables (set by Registry controller)
    lynq.sh/uid: "acme-corp"
    lynq.sh/host: "acme.example.com"
    lynq.sh/hostOrUrl: "https://acme.example.com"
    lynq.sh/activate: "true"
    # Extra variables from extraValueMappings
    lynq.sh/planId: "enterprise"
spec:
  registryId: string                 # Registry name
  templateRef: string                # Template name

  # Rendered resources (already evaluated)
  deployments: []TResource
  # ... (same structure as LynqForm)
```

### Status

```yaml
status:
  observedGeneration: int64
  desiredResources: int32            # Total resources
  readyResources: int32              # Ready resources
  failedResources: int32             # Failed resources
  appliedResources: []string         # Tracked resource keys for orphan detection
                                     # Format: "kind/namespace/name@id"
                                     # Example: ["Deployment/default/app@deploy-1", "Service/default/app@svc-1"]
  conditions:
  - type: Ready
    status: "True"
    reason: Reconciled
    message: "Successfully reconciled all resources"
    lastTransitionTime: timestamp
  - type: Progressing
    status: "False"
    reason: ReconcileComplete
  - type: Conflicted
    status: "False"
    reason: NoConflicts
  - type: Degraded
    status: "False"
    reason: Healthy
    message: "All resources are healthy"
```

#### Condition Types

**Ready Condition**

Indicates whether the tenant is fully reconciled and all resources are ready.

**Status Values:**
- `True`: All resources successfully reconciled and ready
- `False`: Not all resources are ready or some have failed

**Possible Reasons** (when `status=False`):
- `ResourcesFailedAndConflicted`: Both failed and conflicted resources exist (highest priority)
- `ResourcesConflicted`: One or more resources in conflict state
- `ResourcesFailed`: One or more resources failed during reconciliation
- `NotAllResourcesReady`: Resources exist but haven't reached ready state yet

::: tip New in v1.1.4
The Ready condition now provides granular reasons to help quickly identify the root cause of failures. Conflict-related reasons are prioritized for better visibility.
:::

**Progressing Condition**

Indicates whether reconciliation is currently in progress.

**Status Values:**
- `True`: Reconciliation is actively applying changes
- `False`: Reconciliation completed

**Degraded Condition**

::: tip New in v1.1.4
The Degraded condition provides visibility into tenant health issues separate from the Ready condition.
:::

Indicates when a tenant is not functioning optimally, even if reconciliation has completed.

**Status Values:**
- `True`: Tenant has health issues
- `False`: Tenant is healthy

**Possible Reasons** (when `status=True`):
- `ResourceFailuresAndConflicts`: Tenant has both failed and conflicted resources
- `ResourceFailures`: Tenant has failed resources
- `ResourceConflicts`: Tenant has conflicted resources
- `ResourcesNotReady`: Not all resources have reached ready state (new in v1.1.4)

**Conflicted Condition**

Indicates whether any resources have ownership conflicts.

**Status Values:**
- `True`: One or more resources are in conflict
- `False`: No conflicts detected
```

## Field Types

### Duration

String with unit suffix:
- `s`: seconds
- `m`: minutes
- `h`: hours

Examples: `30s`, `1m`, `2h`

### CreationPolicy

- `Once`: Create once, never update
- `WhenNeeded`: Create and update as needed (default)

### DeletionPolicy

- `Delete`: Delete resource on Tenant deletion (default) - uses ownerReference for automatic cleanup
- `Retain`: Keep resource on deletion - uses label-based tracking only (NO ownerReference set at creation)

### ConflictPolicy

- `Stuck`: Stop on ownership conflict (default)
- `Force`: Take ownership forcefully

### PatchStrategy

- `apply`: Server-Side Apply (default)
- `merge`: Strategic Merge Patch
- `replace`: Full replacement

## Annotations

### Tenant Annotations (auto-generated)

```yaml
# Template variables
lynq.sh/uid: string
lynq.sh/host: string
lynq.sh/hostOrUrl: string
lynq.sh/activate: string

# Extra variables from extraValueMappings
lynq.sh/<key>: value

# CreationPolicy tracking
lynq.sh/created-once: "true"
```

### Resource Tracking Labels

**Label-based tracking** is used instead of ownerReferences for:
- Cross-namespace resources (ownerReferences don't work across namespaces)
- Namespace resources (cannot have ownerReferences)
- **DeletionPolicy=Retain resources** (to prevent automatic garbage collection)

```yaml
# Tracking labels (set at resource creation)
lynq.sh/node: tenant-name
lynq.sh/node-namespace: tenant-namespace

# Orphan label (added when resource becomes orphaned - for selector queries)
lynq.sh/orphaned: "true"
```

**Orphan Markers:**

When resources are retained (not deleted) due to `DeletionPolicy=Retain`, the operator adds:

- **Label** `orphaned: "true"` - For easy filtering with label selectors
- **Annotation** `orphaned-at` - RFC3339 timestamp when the resource became orphaned
- **Annotation** `orphaned-reason` - Reason for becoming orphaned:
  - `RemovedFromTemplate`: Resource was removed from LynqForm
  - `LynqNodeDeleted`: LynqNode CR was deleted

**Why split label/annotation?**
- **Label**: Simple value for selector queries (Kubernetes label values must be RFC 1123 compliant)
- **Annotations**: Detailed metadata like timestamps (no value restrictions)

**Orphan Marker Lifecycle:**

1. **Resource removed from template** → Orphan markers added (label + annotations)
2. **Resource re-added to template** → Orphan markers automatically removed during apply
3. **No manual cleanup needed** → Operator manages the full lifecycle

This enables safe template evolution: you can freely add/remove resources from templates, and previously orphaned resources will be cleanly re-adopted if you add them back.

**Finding orphaned resources:**

```bash
# Find all orphaned resources
kubectl get all -A -l lynq.sh/orphaned=true

# Find orphaned resources by reason (using annotation)
kubectl get all -A -l lynq.sh/orphaned=true -o jsonpath='{range .items[?(@.metadata.annotations.k8s-lynq\.org/orphaned-reason=="RemovedFromTemplate")]}{.kind}/{.metadata.name}{"\n"}{end}'

# Find orphaned resources from a specific tenant (label still available)
kubectl get all -A -l lynq.sh/orphaned=true,lynq.sh/lynqnode=my-node
```

### Resource Annotations

**DeletionPolicy Annotation:**

The operator automatically adds a `deletion-policy` annotation to all created resources:

```yaml
metadata:
  annotations:
    lynq.sh/deletion-policy: "Retain"  # or "Delete"
```

**Purpose:**
- **Critical for orphan cleanup**: When resources are removed from templates, they no longer exist in the template spec
- The annotation is the **only source of truth** for determining the correct cleanup behavior
- Without this annotation, all orphaned resources would default to `Delete` policy

**Behavior:**
- Set automatically during resource creation by `renderResource` function
- Read during orphan cleanup in `deleteOrphanedResource` function
- Falls back to `Delete` if annotation is missing (defensive default)

**Example query:**

```bash
# Find all Retain resources
kubectl get all -A -o jsonpath='{range .items[?(@.metadata.annotations.k8s-lynq\.org/deletion-policy=="Retain")]}{.kind}/{.metadata.name}{"\n"}{end}'
```

## Examples

See [Templates Guide](templates.md) and [Quick Start Guide](quickstart.md) for complete examples.

## Validation Rules

### LynqHub

- `spec.valueMappings` must include: `uid`, `hostOrUrl`, `activate`
- `spec.source.syncInterval` must match pattern: `^\d+(s|m|h)$`
- `spec.source.mysql.host` required when `type=mysql`

### LynqForm

- `spec.registryId` must reference existing LynqHub
- Each `TResource.id` must be unique within template
- `dependIds` must not form cycles
- Templates must be valid Go templates

### LynqNode

- Typically validated by operator, not manually created
- All referenced resources must exist

## See Also

- [Template Guide](templates.md) - Template syntax and functions
- [Policies Guide](policies.md) - Policy options
- [Dependencies Guide](dependencies.md) - Resource ordering
