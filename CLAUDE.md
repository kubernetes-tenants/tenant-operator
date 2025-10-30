# Tenant Operator - Development Guidelines

This document provides essential context and guidelines for Claude when working on the Tenant Operator project.

## Project Overview

**Tenant Operator** is a Kubernetes operator that automates multi-tenant resource provisioning using a template-based approach. It reads tenant data from external datasources (initially MySQL) and creates/synchronizes Kubernetes resources declaratively using Server-Side Apply (SSA).

### Core Objectives

1. **Multi-tenant Auto-provisioning**: Read tenant lists from external datasources and create/sync Kubernetes resources using templates
2. **K8s-native Operations**: 3 CRDs (TenantRegistry, TenantTemplate, Tenant) with SSA-centric declarative synchronization
3. **Strong Consistency**: Ensure `desired count = (referencing templates * active rows)`, supporting multiple templates per registry
4. **Policy-based Lifecycle**: Creation/deletion/conflict policies, dependency graph-based apply ordering, failure isolation

---

## Architecture Overview

### Three-Controller Design

```
TenantRegistry Controller -> Reads external DB -> Emits Tenant CRs
TenantTemplate Controller -> Ensures template-registry linkage
Tenant Controller -> Reconciles each Tenant -> SSA applies resources
```

### Key Components

- **TenantRegistry**: Defines external datasource (MySQL), sync interval, and value mappings
- **TenantTemplate**: Blueprint for resources to create per tenant (Deployments, Services, Ingresses, etc.)
- **Tenant**: Instance representing a single tenant row with status tracking
- **SSA Engine**: Server-Side Apply with fieldManager: `tenant-operator`

---

## CRD Responsibilities

### TenantRegistry

**Purpose**: Periodically read rows from external datasource and determine the active tenant set, then create/sync Tenant CRs for each template-row combination.

**Multi-Template Support** ✅:
- **One registry can be referenced by multiple TenantTemplates**
- For each active row, a separate Tenant CR is created for each referencing template
- Tenant naming: `{uid}-{template-name}` format
- Example: Registry "mysql-prod" with 3 rows and 2 templates creates 6 Tenant CRs

**Key Points**:
- Syncs at `spec.source.syncInterval`
- Only rows where `activate` field is truthy are considered active
- `status.referencingTemplates` = count of TenantTemplates referencing this registry
- `status.desired` = `referencingTemplates * activeRows` (total Tenant CRs that should exist)
- `status.ready` = count of Ready Tenants across all templates
- `status.failed` = count of failed Tenants across all templates
- **Garbage Collection** ✅: Automatic Tenant CR cleanup when:
  - Database row is deleted
  - Row's `activate` field becomes false
  - Template no longer references this registry
  - Detailed events emitted: `TenantDeleting`, `TenantDeleted`, `TenantDeletionFailed`
  - Tenant finalizers ensure proper resource cleanup before deletion

**Required Value Mappings**:
- `uid`: Tenant identifier (required)
- `hostOrUrl`: Tenant URL/host (required, auto-extracts `.host`)
- `activate`: Activation flag (required)

**Extra Value Mappings**: Arbitrary column-to-variable mappings for templates

### TenantTemplate

**Purpose**: Defines the resource blueprint for a specific Registry. Each active tenant row uses this template to generate Kubernetes resources.

**Key Points**:
- References a `registryId`
- Contains arrays of resource types: `serviceAccounts`, `deployments`, `services`, `ingresses`, `configMaps`, `secrets`, `jobs`, `cronJobs`, `manifests` (raw)
- Each resource follows `TResource` structure (see api/v1/common_types.go:58)
- All `*Template` fields support Go `text/template` + sprig functions

**Template Variables Available**:
- Required: `.uid`, `.hostOrUrl` (-> `.host` auto-extracted), `.activate`
- Extra: Any keys from `extraValueMappings`
- Context: `.registryId`, `.templateRef`, etc.

**Custom Template Functions** (All Implemented ✅):
- Standard: All sprig functions (200+ from Sprig library)
- Custom (Implemented):
  - `toHost(url)` ✅ - Extract host from URL
  - `trunc63(s)` ✅ - Truncate to 63 chars (K8s name limit)
  - `sha1sum(s)` ✅ - SHA1 hash (Priority 1 implementation)
  - `fromJson(s)` ✅ - Parse JSON string to object (Priority 1 implementation)
  - Plus all sprig functions: `default`, `b64enc`, `b64dec`, `toJson`, `sha256sum`, etc.

### Tenant

**Purpose**: Represents a single tenant instance. The operator creates/syncs all Kubernetes resources for this tenant.

**Key Points**:
- Created automatically by TenantRegistry controller
- Contains resolved resource arrays (templates already evaluated)
- Status tracks `readyResources`, `desiredResources`, `failedResources`
- `Ready` condition requires ALL resources to be ready
- Users typically don't edit Tenant specs directly (managed by operator)

---

## Common Types (api/v1/common_types.go)

### TResource Structure

Every resource in a template is a `TResource` with:

```go
type TResource struct {
    ID                  string              // Unique within template (for dependencies)
    Spec                runtime.RawExtension // K8s resource spec
    DependIds           []string            // IDs of resources to wait for
    CreationPolicy      CreationPolicy      // Once | WhenNeeded (default)
    DeletionPolicy      DeletionPolicy      // Delete (default) | Retain
    ConflictPolicy      ConflictPolicy      // Stuck (default) | Force
    NameTemplate        string              // Go template for name
    LabelsTemplate      map[string]string   // Template-enabled labels
    AnnotationsTemplate map[string]string   // Template-enabled annotations
    WaitForReady        *bool               // Default: true
    TimeoutSeconds      int32               // Default: 300, max: 3600
    PatchStrategy       PatchStrategy       // apply (default) | merge | replace
}
```

### Policies

**CreationPolicy**:
- `Once`: Create only once, never reapply even if spec changes (for init Jobs, security resources)
- `WhenNeeded` (default): Reapply when spec changes or state requires it

**DeletionPolicy**:
- `Delete` (default): Remove resource when Tenant is deleted
- `Retain`: Remove ownerReference but leave resource in cluster

**ConflictPolicy**:
- `Stuck` (default): Stop reconciliation if resource exists with different owner, mark Tenant as Degraded
- `Force`: Use SSA with `force=true` to take ownership, fail gracefully if unsuccessful

**PatchStrategy** ✅ (All Implemented):
- `apply` (default): SSA (Server-Side Apply) with conflict management
- `merge` ✅: Strategic merge patch (Priority 2 implementation)
- `replace` ✅: Full replacement via Update (Priority 2 implementation)
  - Handles create-if-not-exists
  - Preserves resourceVersion for conflict-free updates

---

## Reconciliation Logic

### Registry Controller Flow

1. Get all TenantTemplates that reference this registry
2. Query datasource at `syncInterval`
3. Filter rows where `activate=true`
4. Calculate desired Tenant set (all template-row combinations)
   - For each template and each row, create key: `{template-name}-{uid}`
   - Desired count = `len(templates) * len(activeRows)`
5. Create missing Tenants (naming: `{uid}-{template-name}`), update existing, delete excess
6. Update `status.{referencingTemplates, desired, ready, failed}`

### Tenant Controller Flow ✅

1. **Handle Deletion with Finalizer** ✅ (Implemented):
   - Check if `DeletionTimestamp` is set
   - If finalizer present: Run `cleanupTenantResources()`
     - Respect `DeletionPolicy` per resource:
       - `Delete`: Remove resource from cluster
       - `Retain`: Remove ownerReference, keep resource
   - Remove finalizer after cleanup
   - Return (allow Kubernetes to delete Tenant CR)

2. **Add Finalizer if Missing** ✅ (Implemented):
   - Check if finalizer `tenant.operator.kubernetes-tenants.org/finalizer` exists
   - Add finalizer if missing
   - Update Tenant CR
   - Requeue for reconciliation

3. **Build Template Variables**:
   - Extract variables from Tenant annotations
   - Merge with registry data

4. **Resolve Dependencies**:
   - Build DAG from `dependIds`
   - Detect cycles (fail fast if found)

5. **Topological Sort**:
   - Determine apply order based on dependency graph

6. **For Each Resource in Order**:
   - **Check CreationPolicy** ✅:
     - `Once`: Skip if already created (check annotation `kubernetes-tenants.org/created-once`)
     - `WhenNeeded` (default): Proceed with apply
   - Evaluate `nameTemplate` (namespace is automatically set to Tenant CR's namespace)
   - Render resource `spec` with template variables
   - **Apply ConflictPolicy**:
     - `Stuck`: Check ownership, fail if conflict
     - `Force`: SSA with `force=true`
   - **Apply using Selected PatchStrategy** ✅:
     - `apply` (default): Server-Side Apply
     - `merge`: Strategic Merge Patch
     - `replace`: Get → Update with resourceVersion
   - If `waitForReady=true`: Wait for resource Ready condition (with timeout)
   - Track success/failure with metrics

7. **Update Tenant Status**:
   - Aggregate resource states
   - Update `readyResources`, `failedResources`, `desiredResources`
   - Set `Ready` condition

8. **Requeue for Fast Status Reflection** ✅ (Implemented - Optimized):
   - Return with `RequeueAfter: 30 * time.Second` (changed from 5 minutes)
   - Ensures rapid detection of child resource status changes
   - Combined with event-driven watches for immediate reaction to changes

**Location**: `internal/controller/tenant_controller.go`

### Synchronization Rules

1. **Desired Set Calculation**: Only `activate=true` rows from Registry, multiplied by referencing templates
   - Key: `{template-name}-{uid}` for each template-row combination
   - Total desired = `referencingTemplates * activeRows`
2. **Creation/Deletion**:
   - `desired \ current` -> Create new Tenants (naming: `{uid}-{template-name}`)
   - `current \ desired` -> Delete Tenants (respect `deletionPolicy`)
3. **Drift Detection** ✅ (Implemented & Optimized):
   - **Event-driven**: `Owns()` watches on 11 resource types + `Watches()` on Namespaces
     - ServiceAccounts, Deployments, StatefulSets, DaemonSets
     - Services, ConfigMaps, Secrets, PersistentVolumeClaims
     - Jobs, CronJobs, Ingresses
     - **Namespaces**: Tracked via labels (cannot use ownerReferences)
   - **Watch Predicates**: Only trigger on Generation/Annotation changes (not status-only updates)
     - Reduces unnecessary reconciliation overhead
     - Filters out noisy status updates from watched resources
   - **Fast Requeue**: 30-second periodic requeue (changed from 5 minutes)
     - Ensures child resource status changes are reflected quickly in Tenant status
     - Balances responsiveness with cluster load
   - **Auto-correction**: Automatically reverts manual changes to tenant resources
   - **Location**: `internal/controller/tenant_controller.go` (SetupWithManager, line ~930)
4. **Naming/Scope**: Use `nameTemplate` (63-char limit via `trunc63`). All resources are created in the same namespace as the Tenant CR.
5. **Ordering**: Topological sort by `dependIds`, `waitForReady` enforces readiness gates

---

## Resource Readiness Rules

**When to mark a resource as Ready**:

- **Deployment**: `status.observedGeneration == metadata.generation` AND `availableReplicas >= spec.replicas`
- **StatefulSet**: `readyReplicas == spec.replicas`
- **Service**: Immediate (or `waitForReady=false` recommended)
- **Ingress**: `status.loadBalancer.ingress` exists OR controller-specific Ready condition
- **Job**: `status.succeeded >= 1`
- **ServiceAccount**: Immediate after creation
- **ConfigMap/Secret**: Immediate after creation
- **Custom Resources**: `status.conditions[type=Ready].status=True` or custom health checks

**Timeout**: Each resource has `timeoutSeconds` (default 300s), exceeding marks resource as failed.

---

## Template System

### Template Language

- **Engine**: Go `text/template` + sprig library
- **Syntax**: `{{ .variable }}`, `{{ function arg }}`, `{{ if .condition }}...{{ end }}`

### Available Variables

**Always available**:
- `.uid`: Tenant unique identifier
- `.hostOrUrl`: Original URL/host from registry
- `.host`: Auto-extracted host (from `.hostOrUrl`)
- `.activate`: Activation status
- `.registryId`: Registry name
- `.templateRef`: Template name

**From extraValueMappings**: Any custom mappings (e.g., `.deployImage`, `.planId`)

### Template Functions

**Sprig functions**: `default`, `trim`, `upper`, `lower`, `b64enc`, `b64dec`, `sha256sum`, etc.

**Custom functions**:
- `toHost(url)`: Extract host from URL
- `trunc63(s)`: Truncate to 63 chars (K8s name limit)
- `sha1sum(s)`: SHA1 hash
- `fromJson(s)`, `toJson(obj)`: JSON serialization

### Template Examples

```yaml
nameTemplate: "{{ .uid }}-api"
nameTemplate: "{{ .uid | trunc63 }}"
nameTemplate: "{{ .uid }}-{{ .planId | default \"basic\" }}"

# In Deployment spec:
image: "{{ default \"nginx:stable\" .deployImage }}"
value: "{{ .host }}"
value: "{{ .uid }}"
```

---

## Security & RBAC

### Credentials

- **SecretRef pattern**: All sensitive data (passwords, tokens) use `SecretRef`
- Example: `spec.source.mysql.passwordRef.{name, key}`
- Never hardcode credentials in CRDs

### OwnerReferences

- All created resources have `ownerReference=Tenant`
- Exception: `deletionPolicy=Retain` removes ownerReference on deletion
- Enables automatic garbage collection

### RBAC Requirements

**For operator ServiceAccount**:
- CRDs: `tenantregistries`, `tenanttemplates`, `tenants` (all verbs)
- Workload resources: Required native resources (Deployments, Services, etc.) - namespace-scoped permissions
- `events`, `leases` for leader election
- `secrets` (read-only) for registry credentials

**Principle**: Least privilege, namespace-scoped permissions. All tenant resources are created in the same namespace as the Tenant CR.

---

## Observability

### Events

Emit events for:
- Conflict detected (resource exists with different owner)
- Force apply attempted
- Resource timeout
- Template rendering error
- Tenant Ready transition
- Dependency cycle detected

### Metrics (Prometheus)

```
tenant_reconcile_duration_seconds{result}
tenant_resources_ready{tenant}
registry_desired
registry_ready
registry_failed
apply_attempts_total{kind, result, conflict_policy}
```

### Logging

- Template rendering input snapshots (mask sensitive data)
- SSA diff summaries
- Reconciliation start/end with duration
- Error details with context

---

## Validation & Webhooks

### ValidatingWebhook

- `TenantTemplate.spec.registryId` must reference existing Registry
- `valueMappings` must include required keys: `uid`, `hostOrUrl`, `activate`
- `TResource.id` must be unique within template
- `dependIds` must not form cycles
- Templates must be valid Go templates

### DefaultingWebhook

Set defaults:
- `creationPolicy=WhenNeeded`
- `deletionPolicy=Delete`
- `conflictPolicy=Stuck`
- `waitForReady=true`
- `timeoutSeconds=300`
- `patchStrategy=apply`

### OpenAPI Schema

All CRDs have comprehensive OpenAPI v3 schemas with:
- Required fields marked
- Enums for policy types
- Patterns for intervals (e.g., `^\d+(s|m|h)$`)
- Min/max constraints

---

## Failure Handling

### Template Rendering Failure

- Mark Tenant as Degraded
- Emit event with error details (missing variable, type error)
- Do not apply any resources

### Conflict (Stuck Policy)

- Stop reconciliation for that resource
- Emit event: "Resource {name} exists with different owner"
- Mark Tenant as Degraded
- Provide hints: Check namespace uniqueness, review naming templates

### Ready Timeout

- Retry with exponential backoff
- If cumulative failures exceed threshold, mark Tenant as Degraded
- Emit event with resource status details

### Dependency Cycle

- Detect during DAG construction
- Mark Tenant as Degraded immediately
- Emit event: "Dependency cycle detected: A -> B -> A"

---

## Performance & Scalability

### Controller Design

- Separate workqueues for Registry/Template/Tenant controllers
- Rate-limited retries with exponential backoff
- Concurrent reconciliation: `--concurrency.tenant=N` flag

### Reconciliation Optimization ✅

- **Fast Status Reflection**: 30-second requeue interval (optimized from 5 minutes)
  - Child resource status changes reflected in Tenant status within 30 seconds
  - Balances responsiveness with cluster resource usage
- **Smart Watch Predicates**: Only reconcile on meaningful changes
  - Generation changes (spec updates)
  - Annotation changes
  - Filters out status-only updates to reduce reconciliation overhead
- **Namespace Tracking**: Label-based tracking for Namespaces
  - Labels: `kubernetes-tenants.org/tenant`, `kubernetes-tenants.org/tenant-namespace`
  - Enables immediate reconciliation when Namespaces are modified
- **Event-Driven Architecture**: Immediate reconciliation on watched resource changes

### Large-scale Optimization

- SSA batching for bulk applies
- Resource-type worker pools
- Optional sharding: `--shard=N/M` or namespace partitioning
- Cache frequently accessed resources (registries, templates)

---

## Development Guidelines

### Code Structure

```
api/v1/               # CRD types (already includes common_types.go)
internal/controller/  # Controller implementations
pkg/template/         # Template rendering engine
pkg/apply/            # SSA apply engine
pkg/health/           # Resource readiness checks
pkg/datasource/       # External datasource integrations
```

### Testing Strategy

- Unit tests: Template rendering, policy logic, dependency graph
- Integration tests: Controller reconciliation against real API server
- E2E tests: Full workflow with MySQL datasource

### Important Invariants

1. `Registry.status.desired` MUST equal active row count
2. `Tenant` is Ready IFF all resources are Ready
3. SSA fieldManager MUST be `tenant-operator`
4. Dependency cycles MUST be rejected
5. Naming MUST respect 63-char K8s limit

### Common Pitfalls

- Don't forget to auto-extract `.host` from `.hostOrUrl`
- Template rendering errors should never panic
- Always validate `dependIds` before topological sort
- SSA requires `fieldManager` and correct content-type
- `waitForReady=true` blocks the reconciliation pipeline

---

## Example Workflow

### Single Template Scenario
1. User creates `TenantRegistry` pointing to MySQL in a specific namespace (e.g., `default`)
2. User creates `TenantTemplate` (e.g., "web-app") referencing the registry in the same namespace
3. Registry controller syncs every `syncInterval`:
   - Finds 1 template referencing this registry
   - Queries MySQL for active rows (`activate=true`) - e.g., 3 rows (uid: a, b, c)
   - Creates 3 `Tenant` CRs: `a-web-app`, `b-web-app`, `c-web-app`
   - Updates `status.referencingTemplates=1`, `status.desired=3`

### Multi-Template Scenario (✅ NEW)
1. User creates `TenantRegistry` pointing to MySQL
2. User creates multiple `TenantTemplate` CRs referencing the same registry:
   - "web-app" template (web tier resources)
   - "worker" template (background job resources)
3. Registry controller syncs every `syncInterval`:
   - Finds 2 templates referencing this registry
   - Queries MySQL for 3 active rows (uid: a, b, c)
   - Creates 6 `Tenant` CRs (2 templates × 3 rows):
     - `a-web-app`, `b-web-app`, `c-web-app`
     - `a-worker`, `b-worker`, `c-worker`
   - Updates `status.referencingTemplates=2`, `status.desired=6`

### Reconciliation Process
4. For each `Tenant`:
   - Template controller ensures linkage
   - Tenant controller reconciles:
     - Renders templates with row data
     - Builds dependency graph
     - Applies all resources in the same namespace as the Tenant CR
     - Waits for readiness
     - Updates status
5. Users observe:
   - `Registry.status.{referencingTemplates, desired, ready, failed}`
   - `Tenant.status.conditions[Ready]`
   - Events and metrics
   - All tenant resources in the same namespace

---

## References

- **K8s SSA**: https://kubernetes.io/docs/reference/using-api/server-side-apply/
- **Sprig Functions**: https://masterminds.github.io/sprig/
- **Kubebuilder**: https://book.kubebuilder.io/
- **Controller Runtime**: https://pkg.go.dev/sigs.k8s.io/controller-runtime
