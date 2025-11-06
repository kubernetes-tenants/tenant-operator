# API Reference

Complete API reference for Tenant Operator CRDs.

[[toc]]

## TenantRegistry

Defines external data source and sync configuration.

::: info Resource metadata
- **Kind:** `TenantRegistry`
- **API Version:** `operator.kubernetes-tenants.org/v1`
:::

### Spec

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantRegistry
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
      maxOpenConns: int              # Max open connections (optional, default: 10)
      maxIdleConns: int              # Max idle connections (optional, default: 5)
      connMaxLifetime: duration      # Connection lifetime (optional, default: 5m)
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
  desired: int32                     # Desired Tenant CRs (templates × rows)
  ready: int32                       # Ready Tenant CRs
  failed: int32                      # Failed Tenant CRs
  conditions:                        # Status conditions
  - type: Ready
    status: "True"
    reason: SyncSucceeded
    message: "Successfully synced N tenants"
    lastTransitionTime: timestamp
```

## TenantTemplate

Defines resource blueprint for tenants.

::: info Resource metadata
- **Kind:** `TenantTemplate`
- **API Version:** `operator.kubernetes-tenants.org/v1`
:::

### Spec

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: my-template
spec:
  registryId: string                 # TenantRegistry name (required)

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
  totalTenants: int32                # Total tenants using this template
  readyTenants: int32                # Ready tenants
  conditions:
  - type: Valid
    status: "True"
    reason: ValidationSucceeded
```

## Tenant

Represents a single tenant instance.

::: info Resource metadata
- **Kind:** `Tenant`
- **API Version:** `operator.kubernetes-tenants.org/v1`
:::

::: warning Managed resource
Tenant objects are typically managed by the operator and rarely created manually.
:::

### Spec

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: Tenant
metadata:
  name: acme-prod-template
  annotations:
    # Template variables (set by Registry controller)
    kubernetes-tenants.org/uid: "acme-corp"
    kubernetes-tenants.org/host: "acme.example.com"
    kubernetes-tenants.org/hostOrUrl: "https://acme.example.com"
    kubernetes-tenants.org/activate: "true"
    # Extra variables from extraValueMappings
    kubernetes-tenants.org/planId: "enterprise"
spec:
  registryId: string                 # Registry name
  templateRef: string                # Template name

  # Rendered resources (already evaluated)
  deployments: []TResource
  # ... (same structure as TenantTemplate)
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
kubernetes-tenants.org/uid: string
kubernetes-tenants.org/host: string
kubernetes-tenants.org/hostOrUrl: string
kubernetes-tenants.org/activate: string

# Extra variables from extraValueMappings
kubernetes-tenants.org/<key>: value

# CreationPolicy tracking
kubernetes-tenants.org/created-once: "true"
```

### Resource Tracking Labels

**Label-based tracking** is used instead of ownerReferences for:
- Cross-namespace resources (ownerReferences don't work across namespaces)
- Namespace resources (cannot have ownerReferences)
- **DeletionPolicy=Retain resources** (to prevent automatic garbage collection)

```yaml
# Tracking labels (set at resource creation)
kubernetes-tenants.org/tenant: tenant-name
kubernetes-tenants.org/tenant-namespace: tenant-namespace

# Orphan label (added when resource becomes orphaned - for selector queries)
kubernetes-tenants.org/orphaned: "true"
```

**Orphan Markers:**

When resources are retained (not deleted) due to `DeletionPolicy=Retain`, the operator adds:

- **Label** `orphaned: "true"` - For easy filtering with label selectors
- **Annotation** `orphaned-at` - RFC3339 timestamp when the resource became orphaned
- **Annotation** `orphaned-reason` - Reason for becoming orphaned:
  - `RemovedFromTemplate`: Resource was removed from TenantTemplate
  - `TenantDeleted`: Tenant CR was deleted

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
kubectl get all -A -l kubernetes-tenants.org/orphaned=true

# Find orphaned resources by reason (using annotation)
kubectl get all -A -l kubernetes-tenants.org/orphaned=true -o jsonpath='{range .items[?(@.metadata.annotations.kubernetes-tenants\.org/orphaned-reason=="RemovedFromTemplate")]}{.kind}/{.metadata.name}{"\n"}{end}'

# Find orphaned resources from a specific tenant (label still available)
kubectl get all -A -l kubernetes-tenants.org/orphaned=true,kubernetes-tenants.org/tenant=my-tenant
```

### Resource Annotations

**DeletionPolicy Annotation:**

The operator automatically adds a `deletion-policy` annotation to all created resources:

```yaml
metadata:
  annotations:
    kubernetes-tenants.org/deletion-policy: "Retain"  # or "Delete"
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
kubectl get all -A -o jsonpath='{range .items[?(@.metadata.annotations.kubernetes-tenants\.org/deletion-policy=="Retain")]}{.kind}/{.metadata.name}{"\n"}{end}'
```

## Examples

See [Templates Guide](templates.md) and [Quick Start Guide](quickstart.md) for complete examples.

## Validation Rules

### TenantRegistry

- `spec.valueMappings` must include: `uid`, `hostOrUrl`, `activate`
- `spec.source.syncInterval` must match pattern: `^\d+(s|m|h)$`
- `spec.source.mysql.host` required when `type=mysql`

### TenantTemplate

- `spec.registryId` must reference existing TenantRegistry
- Each `TResource.id` must be unique within template
- `dependIds` must not form cycles
- Templates must be valid Go templates

### Tenant

- Typically validated by operator, not manually created
- All referenced resources must exist

## See Also

- [Template Guide](templates.md) - Template syntax and functions
- [Policies Guide](policies.md) - Policy options
- [Dependencies Guide](dependencies.md) - Resource ordering
