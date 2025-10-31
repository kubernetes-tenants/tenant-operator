# Glossary

Comprehensive reference of terms and concepts used in Tenant Operator.

[[toc]]

## Core Concepts

### Tenant Operator

A Kubernetes operator that automates multi-tenant application provisioning by synchronizing tenant data from external data sources (MySQL, PostgreSQL) and creating Kubernetes resources using template-based declarative configuration.

### Multi-Tenancy

An architectural pattern where a single instance of software serves multiple customers (tenants), each with isolated resources and data.

### Operator Pattern

A Kubernetes design pattern that uses custom controllers to extend Kubernetes functionality by automating application-specific operational knowledge.

## Custom Resource Definitions (CRDs)

### TenantRegistry

A Custom Resource that defines:
- External data source connection (MySQL, PostgreSQL)
- Synchronization interval
- Column mappings between database schema and operator variables
- The source of truth for active tenant list

**Example:**
```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantRegistry
metadata:
  name: my-registry
spec:
  source:
    type: mysql
    syncInterval: 1m
```

### TenantTemplate

A Custom Resource that defines:
- Resource blueprints for tenant provisioning
- Template definitions for Kubernetes resources
- Lifecycle policies (creation, deletion, conflict)
- Dependency relationships between resources

**Referenced by:** Tenant CRs
**References:** TenantRegistry (via `registryId`)

### Tenant

A Custom Resource representing a single tenant instance. Automatically created by TenantRegistry controller based on active database rows.

**Key characteristics:**
- Created/deleted automatically (users typically don't create manually)
- Contains resolved resource specifications (templates already evaluated)
- Tracks status of all provisioned resources
- Owns all created Kubernetes resources via ownerReferences

## Data Source Concepts

### Data Source

An external system that stores tenant configuration data. Currently supported:
- **MySQL** (v5.7+, v8.0+)
- **PostgreSQL** (planned for v1.1)

### syncInterval

Duration between data source polling cycles. Defines how frequently the operator checks for tenant data changes.

**Format:** Go duration string (e.g., `30s`, `1m`, `5m`)
**Default:** 1 minute
**Tradeoff:** Lower interval = faster synchronization, higher database load

### valueMappings

Required column mappings from database to operator variables:

| Mapping | Database Column | Operator Variable | Description |
|---------|-----------------|-------------------|-------------|
| `uid` | Custom | `.uid` | Unique tenant identifier |
| `hostOrUrl` | Custom | `.hostOrUrl`, `.host` | Tenant URL (auto-extracts host) |
| `activate` | Custom | `.activate` | Activation flag (truthy/falsy) |

**Example:**
```yaml
valueMappings:
  uid: tenant_id
  hostOrUrl: tenant_url
  activate: is_active
```

### extraValueMappings

Optional custom column mappings for template variables. Allows passing arbitrary database columns to templates.

**Example:**

::: v-pre
```yaml
extraValueMappings:
  planId: subscription_plan  # Available as {{ .planId }}
  region: deployment_region  # Available as {{ .region }}
```
:::

### activate Column

A special database column that determines tenant activation status.

::: warning Truthy values
| Accepted values (case-sensitive) | Result |
| --- | --- |
| `"1"` | Active |
| `"true"`, `"TRUE"`, `"True"` | Active |
| `"yes"`, `"YES"`, `"Yes"` | Active |

Any other value—such as `"0"`, `"false"`, `"active"`, empty strings, or `NULL`—is treated as **inactive**.
:::

::: tip Normalize upstream data
Use MySQL VIEWs to transform incompatible values before the operator consumes them.
:::

### MySQL VIEW

A virtual table created from a SQL query. Used to transform database schemas that don't match operator requirements.

**Use cases:**
- Transform `"active"`/`"inactive"` to `"1"`/`"0"`
- Combine multiple columns
- Add computed fields
- Filter sensitive data

**Example:**
```sql
CREATE VIEW tenant_configs AS
SELECT id, url, CASE WHEN status='active' THEN '1' ELSE '0' END AS is_active FROM tenants;
```

See [DataSource Guide](datasource.md) for detailed examples.

## Template System

### Template

Go `text/template` syntax with Sprig function library. Used to dynamically generate Kubernetes resource specifications.

::: v-pre
**Syntax:**
- Variables: `{{ .uid }}`, `{{ .host }}`
- Functions: `{{ .uid | trunc63 }}`, `{{ .config | fromJson }}`
- Conditionals: `{{ if .planId }}...{{ end }}`
:::

### Template Variables

Data available in template rendering context:

**Required (from valueMappings):**
- `.uid` - Tenant unique identifier
- `.hostOrUrl` - Original URL from database
- `.host` - Auto-extracted hostname from `.hostOrUrl`
- `.activate` - Activation status (truthy/falsy)

**Metadata:**
- `.registryId` - TenantRegistry name
- `.templateRef` - TenantTemplate name

**Custom (from extraValueMappings):**
- Any additional database columns (e.g., `.planId`, `.region`)

### Template Functions

Built-in and custom functions available in templates:

**Custom Functions:**
- `toHost(url)` - Extract hostname from URL
- `trunc63(s)` - Truncate to 63 characters (Kubernetes name limit)
- `sha1sum(s)` - SHA1 hash of string
- `fromJson(s)` - Parse JSON string to object

**Sprig Functions (200+):**
- String: `trim`, `upper`, `lower`, `replace`, `split`
- Encoding: `b64enc`, `b64dec`, `sha256sum`
- Defaults: `default`, `coalesce`, `ternary`
- Collections: `list`, `dict`, `merge`
- And many more: See [Sprig documentation](https://masterminds.github.io/sprig/)

### nameTemplate

A template string that generates the `metadata.name` for a Kubernetes resource.

**Requirements:**
- Must result in valid Kubernetes name (lowercase, alphanumeric, `-`)
- Maximum 63 characters (use `trunc63` function)
- Must be unique within namespace

**Example:**

::: v-pre
```yaml
nameTemplate: "{{ .uid }}-app"
nameTemplate: "{{ .uid | trunc63 }}"
```
:::

## Resource Management

### Server-Side Apply (SSA)

A Kubernetes API mechanism for declarative resource management. The operator uses SSA as the default apply strategy.

**Benefits:**
- Conflict-free updates (multiple controllers can manage same resource)
- Field-level ownership tracking
- Automatic drift correction

**Field Manager:** `tenant-operator`

**Reference:** [Kubernetes SSA Documentation](https://kubernetes.io/docs/reference/using-api/server-side-apply/)

### fieldManager

The identifier used in Server-Side Apply to track which controller owns which fields.

**Value:** `tenant-operator`

All resources applied by Tenant Operator are marked with this field manager.

### TResource

The base structure for all resources in TenantTemplate. Contains:
- `id` - Unique identifier within template
- `spec` - Kubernetes resource specification
- `dependIds` - Dependency list (topological ordering)
- Policies: `creationPolicy`, `deletionPolicy`, `conflictPolicy`, `patchStrategy`
- Templates: `nameTemplate`, `labelsTemplate`, `annotationsTemplate`
- Readiness: `waitForReady`, `timeoutSeconds`

**Resource types:**
- `serviceAccounts`
- `deployments`, `statefulSets`, `daemonSets`
- `services`
- `configMaps`, `secrets`
- `persistentVolumeClaims`
- `jobs`, `cronJobs`
- `ingresses`
- `manifests` (raw YAML for custom resources)

### ownerReference

A Kubernetes metadata field that establishes parent-child relationships between resources.

**In Tenant Operator:**
- All created resources have `ownerReference` pointing to their Tenant CR
- Enables automatic garbage collection when Tenant is deleted
- Exception: Resources with `deletionPolicy: Retain`

### Drift Detection

The process of detecting and correcting manual changes to operator-managed resources.

**Tenant Operator uses dual-layer detection:**

**Event-Driven (Immediate):**
- Watches 11 resource types (Deployments, Services, ConfigMaps, etc.)
- Triggers reconciliation on generation/annotation changes
- Smart predicates filter status-only updates

**Periodic (30 seconds):**
- Regular reconciliation ensures eventual consistency
- Detects child resource status changes
- Balances responsiveness with cluster resource usage

**Result:** Manual changes are automatically reverted to template-defined state.

## Policies

Policies control resource lifecycle behavior in TenantTemplate.

### CreationPolicy

Controls when resources are created/updated.

**Values:**
- `WhenNeeded` (default) - Reapply when spec changes or state requires it
- `Once` - Create only once, never reapply (for init Jobs, immutable resources)

**Use cases for `Once`:**
- Initialization Jobs
- Secret generation
- Certificate creation
- Database migrations

### DeletionPolicy

Controls what happens to resources when Tenant CR is deleted.

**Values:**
- `Delete` (default) - Remove resource from cluster
- `Retain` - Remove ownerReference, keep resource in cluster

**Use cases for `Retain`:**
- Persistent data (PVCs, Databases)
- Shared infrastructure
- Resources needing manual cleanup
- Protection from accidental deletion

### ConflictPolicy

Controls behavior when resource already exists with different owner.

**Values:**
- `Stuck` (default) - Stop reconciliation, mark Tenant as Degraded, emit event
- `Force` - Use SSA with `force=true` to take ownership

**Use `Force` when:**
- Migrating from manual to operator management
- Multiple operators need to share resources
- Recovering from operator failures

**Warning:** `Force` can cause conflicts with other controllers.

### PatchStrategy

The method used to update Kubernetes resources.

**Values:**
- `apply` (default) - Server-Side Apply (SSA) with conflict management
- `merge` - Strategic Merge Patch (preserves fields not in patch)
- `replace` - Full replacement via Update (replace entire resource)

**Recommendation:** Use `apply` (default) unless you have specific requirements.

## Dependency Management

### dependIds

An array of resource IDs that must be created before the current resource.

**Example:**

::: v-pre
```yaml
deployments:
  - id: app-deploy
    dependIds: ["app-config"]  # Wait for configmap
```
:::

**Use cases:**
- PersistentVolumeClaim must exist before StatefulSet
- ConfigMap/Secret must exist before Deployment
- Service must exist before Ingress

### Dependency Graph (DAG)

A Directed Acyclic Graph representing resource creation order based on `dependIds`.

**Properties:**
- Nodes: Resources
- Edges: Dependencies
- Must be acyclic (no circular dependencies)

**Validation:** Operator detects cycles at admission time and rejects invalid templates.

### Topological Sort

An algorithm that determines the correct order to create resources based on their dependencies.

**Process:**
1. Build dependency graph from `dependIds`
2. Detect cycles (fail if found)
3. Sort resources in dependency order
4. Apply resources sequentially

**Result:** Resources are created in the correct order, respecting dependencies.

## Resource Readiness

### waitForReady

A boolean flag (per resource) that determines if the operator should wait for resource readiness before proceeding.

**Default:** `true`

**When false:** Operator applies resource and immediately moves to next one (fire-and-forget).

**Use cases for `false`:**
- Services (typically ready immediately)
- ConfigMaps/Secrets (no readiness concept)
- Background Jobs
- Resources where readiness doesn't matter

### timeoutSeconds

Maximum duration (in seconds) to wait for resource readiness.

**Default:** 300 (5 minutes)
**Maximum:** 3600 (1 hour)

**Behavior on timeout:** Resource marked as failed, Tenant marked as Degraded.

### Ready Condition

A Kubernetes status condition indicating resource health. Different resource types have different readiness criteria:

| Resource Type | Ready When |
|---------------|------------|
| Deployment | `availableReplicas >= replicas` AND `observedGeneration == generation` |
| StatefulSet | `readyReplicas == replicas` |
| Job | `succeeded >= 1` |
| Service | Immediate (or `waitForReady=false` recommended) |
| Namespace | Immediate after creation |
| Custom Resources | `status.conditions[type=Ready].status=True` |

## Status and Observability

### Tenant Status

The `status` field in Tenant CR tracks resource provisioning state:

**Fields:**
- `conditions` - Array of status conditions
  - `Ready` - All resources are ready
  - `Degraded` - Some resources failed
- `desiredResources` - Total count of resources to create
- `readyResources` - Count of ready resources
- `failedResources` - Count of failed resources
- `lastSyncTime` - Timestamp of last successful reconciliation

**Ready calculation:** `Ready = (readyResources == desiredResources) AND (failedResources == 0)`

### Registry Status

The `status` field in TenantRegistry CR tracks tenant provisioning across all templates:

**Fields:**
- `desired` - Expected Tenant count: `referencingTemplates × activeRows`
- `ready` - Count of Ready Tenants across all templates
- `failed` - Count of failed Tenants across all templates
- `lastSyncTime` - Timestamp of last database sync
- `referencingTemplates` - List of templates using this registry

**Example:**
- Active database rows: 10
- Referencing templates: 2
- Desired: 20 (10 rows × 2 templates)

### Reconciliation

The control loop process where the operator compares desired state (from templates) with actual state (in cluster) and takes action to converge them.

**Tenant Operator reconciliation layers:**

**1. TenantRegistry Reconciliation:**
- Query database at `syncInterval`
- Calculate desired Tenant set
- Create/update/delete Tenant CRs

**2. TenantTemplate Reconciliation:**
- Validate template-registry linkage
- Ensure template consistency

**3. Tenant Reconciliation:**
- Render templates with tenant data
- Build dependency graph
- Apply resources in topological order
- Wait for readiness
- Update status

**Triggers:**
- Database sync timer (syncInterval)
- CRD changes (create/update/delete)
- Child resource changes (event-driven)
- Periodic requeue (30 seconds)

### Requeue

The act of scheduling a resource for re-reconciliation after a delay.

**Tenant Operator requeue patterns:**
- **Immediate:** Error conditions, conflicts
- **30 seconds:** Normal reconciliation cycle (fast status reflection)
- **syncInterval:** Database polling (e.g., 1 minute)

## Multi-Template Support

### Multi-Template

The capability for one TenantRegistry to be referenced by multiple TenantTemplates.

**Use cases:**
- Multi-environment deployments (prod, staging, dev)
- A/B testing configurations
- Regional variations
- Different service tiers

**Example:**

::: v-pre
```yaml
# Registry: my-registry (5 active rows)
# Template 1: prod-template (registryId: my-registry)
# Template 2: staging-template (registryId: my-registry)
# Result: 10 Tenant CRs (5 rows × 2 templates)
```
:::

### Referencing Templates

TenantTemplates that reference a specific TenantRegistry via `spec.registryId`.

**Tracked in:** `TenantRegistry.status.referencingTemplates`

**Used for:** Calculating desired Tenant count:
```
desired = len(referencingTemplates) × activeRows
```

### Desired Count Calculation

Formula for determining expected number of Tenant CRs:

```
TenantRegistry.status.desired = referencingTemplates × activeRows
```

**Example:**
- Active database rows: 10
- Referencing templates: 3
- Desired: 30

**Purpose:** Ensures strong consistency between data source and cluster state.

## Performance Optimization

### Smart Watch Predicates

Filtering logic that prevents unnecessary reconciliations by ignoring irrelevant resource changes.

**Filters out:**
- Status-only updates
- Metadata changes (except annotations)
- ResourceVersion changes

**Triggers reconciliation only on:**
- Generation changes (spec updates)
- Annotation changes
- Resource deletion

**Impact:** Significantly reduces reconciliation overhead.

### Namespace Tracking

Label-based mechanism for tracking namespace ownership without ownerReferences (not allowed on cluster-scoped resources).

**Implementation:**
- Namespaces labeled with: `tenant.operator.kubernetes-tenants.org/tenant: <tenant-name>`
- Watch predicate filters by label
- Enables efficient namespace-specific reconciliation

### Concurrency

The number of parallel reconciliation workers.

**Configuration:** (Not exposed as flags in current implementation)
- Controlled by controller-runtime defaults
- Future: `--tenant-concurrency`, `--registry-concurrency` flags

**Tradeoff:** Higher concurrency = faster processing, more resource usage.

## Webhooks and Validation

### Webhook

An HTTP callback mechanism for intercepting Kubernetes API requests before they're persisted.

**Tenant Operator webhooks:**
- **ValidatingWebhook:** Reject invalid TenantRegistry/TenantTemplate CRs
- **MutatingWebhook:** Set default values (policies, timeouts)

**Requires:** TLS certificates (managed by cert-manager)

### cert-manager

A Kubernetes add-on that automates TLS certificate management.

**Used for:** Webhook server TLS certificates

**Installation:**
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

**Benefits:**
- Automatic certificate issuance
- Automatic renewal before expiration
- CA bundle injection into webhook configurations

**Required:** Tenant Operator cannot run without cert-manager.

### Validation Rules

Checks performed by ValidatingWebhook:

**TenantRegistry:**
- `valueMappings` must include required keys: `uid`, `hostOrUrl`, `activate`
- `syncInterval` must be valid duration (e.g., `1m`, not `1 minute`)
- Database connection details must be valid

**TenantTemplate:**
- `spec.registryId` must reference existing TenantRegistry
- `TResource.id` must be unique within template
- `dependIds` must not form cycles
- Templates must be valid Go template syntax
- Policies must be valid enum values

## Garbage Collection and Finalizers

### Finalizer

A Kubernetes mechanism that prevents resource deletion until cleanup tasks complete.

**Tenant Operator finalizer:** `tenant.operator.kubernetes-tenants.org/finalizer`

**Added to:** Tenant CRs

**Purpose:** Ensures proper cleanup of resources when Tenant is deleted (respecting `deletionPolicy`).

### Cascade Deletion

The automatic deletion of child resources when parent resource is deleted.

**In Tenant Operator:**

**Normal flow:**
```
Tenant CR deleted → Finalizer runs → Resources deleted (per deletionPolicy) → Finalizer removed → Tenant CR removed
```

**Warning:**
```
TenantRegistry deleted → All Tenant CRs deleted (ownerReference) → All tenant resources deleted
```

**Protection:** Set `deletionPolicy: Retain` on critical resources BEFORE deleting Registry/Template.

See [Policies Guide - Cascade Deletion](policies.md#️-important-protecting-tenants-from-cascade-deletion) for details.

## Kubernetes Concepts

### Custom Resource Definition (CRD)

A Kubernetes extension mechanism that allows defining custom resource types.

**Tenant Operator CRDs:**
- `tenantregistries.operator.kubernetes-tenants.org`
- `tenanttemplates.operator.kubernetes-tenants.org`
- `tenants.operator.kubernetes-tenants.org`

### Controller

A control loop that watches Kubernetes resources and takes actions to move current state toward desired state.

**Tenant Operator controllers:**
1. TenantRegistry Controller
2. TenantTemplate Controller
3. Tenant Controller

### Reconciliation Loop

See [Reconciliation](#reconciliation) above.

### Leader Election

A mechanism ensuring only one instance of the operator is actively reconciling resources (for high availability deployments).

**Configuration:** `--leader-elect` flag (default: enabled)

### Kubernetes API Server

The central management entity in Kubernetes that exposes the Kubernetes API. All operator interactions go through the API server.

### etcd

The distributed key-value store used by Kubernetes to persist cluster state. The operator reads/writes all resources to etcd via the API server.

## Development and Operations

### Minikube

A tool for running a single-node Kubernetes cluster locally for development and testing.

**Tenant Operator Minikube setup:**
- See [Quick Start](quickstart.md) for automated scripts
- See [Local Development](local-development-minikube.md) for development workflow

### Server-Side Apply (SSA) Engine

The internal component in Tenant Operator that applies Kubernetes resources using Server-Side Apply.

**Features:**
- Conflict detection and resolution
- Field-level ownership tracking
- Automatic drift correction
- Support for multiple patch strategies

**Location:** `internal/apply/`

### Apply Engine

See [Server-Side Apply (SSA) Engine](#server-side-apply-ssa-engine) above.

### Template Engine

The internal component that renders Go templates with tenant data.

**Features:**
- Variable substitution
- Function execution (Sprig + custom)
- Error handling and validation
- Template caching

**Location:** `internal/template/`

## Metrics and Monitoring

### Prometheus Metrics

Time-series metrics exposed by the operator for monitoring.

**Key metrics:**
- `tenant_reconcile_duration_seconds{result}` - Reconciliation duration
- `tenant_resources_ready{tenant}` - Ready resource count per tenant
- `registry_desired` - Expected Tenant count
- `registry_ready` - Ready Tenant count
- `registry_failed` - Failed Tenant count
- `apply_attempts_total{kind, result, conflict_policy}` - Apply attempts

**Endpoint:** `http://operator:8080/metrics`

### Health Checks

HTTP endpoints for liveness and readiness probes.

**Endpoints:**
- `/healthz` - Liveness probe (is operator running?)
- `/readyz` - Readiness probe (is operator ready to reconcile?)

## Common Abbreviations

| Abbreviation | Full Term |
|--------------|-----------|
| SSA | Server-Side Apply |
| CRD | Custom Resource Definition |
| CR | Custom Resource |
| DAG | Directed Acyclic Graph |
| RBAC | Role-Based Access Control |
| TLS | Transport Layer Security |
| API | Application Programming Interface |
| K8s | Kubernetes (8 letters between K and s) |
| PVC | PersistentVolumeClaim |
| SA | ServiceAccount |

## See Also

- [API Reference](api.md) - Complete CRD specification
- [Template Guide](templates.md) - Template syntax and examples
- [Policies Guide](policies.md) - Lifecycle policies in detail
- [DataSource Guide](datasource.md) - MySQL configuration and VIEWs
- [Quick Start](quickstart.md) - Get started in 5 minutes
