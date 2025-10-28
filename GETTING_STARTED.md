# Getting Started with Tenant Operator v1.0 Development

This guide helps you get started with developing and testing the Tenant Operator v1.0.

## Prerequisites

- Go 1.20 or later
- Docker Desktop or equivalent
- kubectl
- Kind or Minikube (for local testing)
- operator-sdk v1.41.1 or later

## Project Structure

```
tenant-operator/
├── api/v1/                          # API type definitions
│   ├── common_types.go             # Common types (TResource, policies)
│   ├── tenantregistry_types.go     # TenantRegistry CRD
│   ├── tenanttemplate_types.go     # TenantTemplate CRD
│   ├── tenant_types.go             # Tenant CRD
│   └── zz_generated.deepcopy.go    # Generated code
├── cmd/
│   └── main.go                     # Operator entry point
├── internal/controller/            # Controller implementations
│   ├── tenantregistry_controller.go
│   ├── tenanttemplate_controller.go
│   ├── tenant_controller.go
│   └── suite_test.go
├── config/                         # Kubernetes manifests
│   ├── crd/bases/                  # Generated CRD YAML files
│   ├── samples/                    # Sample custom resources
│   ├── manager/                    # Operator deployment
│   └── rbac/                       # RBAC configurations
├── MIGRATION_GUIDE.md              # Migration from v0.x
└── README.md                       # Project documentation
```

## Quick Start

### 1. Clone and Build

```bash
# Navigate to the project directory
cd /Users/tim/Projects/personals/tenant-operator

# Install dependencies
go mod download

# Generate code and manifests
make generate
make manifests

# Build the operator binary
make build
```

### 2. Run Tests

```bash
# Run unit tests
make test

# Run with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 3. Run Locally

```bash
# Install CRDs to your cluster
make install

# Run the operator locally (requires KUBECONFIG)
make run

# The operator will run on your machine and connect to your k8s cluster
```

### 4. Deploy to Cluster

```bash
# Build and push Docker image
export IMG=your-registry/tenant-operator:v1.0
make docker-build docker-push IMG=$IMG

# Deploy to cluster
make deploy IMG=$IMG

# Verify deployment
kubectl get pods -n tenant-operator-system
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager
```

## Next Steps: Implementing Controllers

The current project has scaffolded controllers, but the reconciliation logic needs to be implemented. Here's the recommended implementation order:

### Phase 1: TenantRegistry Controller

**Goal**: Read from MySQL and create/update Tenant CRs.

**Implementation checklist**:
- [ ] Database connection pooling
- [ ] Periodic sync loop (syncInterval)
- [ ] Query execution with valueMappings
- [ ] Tenant CR creation/update logic
- [ ] Garbage collection (remove inactive tenants)
- [ ] Status updates (desired/ready/failed counts)
- [ ] Error handling and retries

**File**: `internal/controller/tenantregistry_controller.go`

**Sample code structure**:
```go
func (r *TenantRegistryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. Fetch TenantRegistry
    // 2. Connect to MySQL
    // 3. Query active tenants (WHERE activate = true)
    // 4. Compare with existing Tenant CRs
    // 5. Create/update/delete Tenant CRs
    // 6. Update TenantRegistry status
    // 7. Requeue after syncInterval
}
```

### Phase 2: TenantTemplate Controller

**Goal**: Validate template and ensure TenantRegistry reference exists.

**Implementation checklist**:
- [ ] Validate registryId exists
- [ ] Check for ID conflicts within template
- [ ] Validate dependency graph (no cycles)
- [ ] Validate template syntax (optional pre-check)
- [ ] Status updates with validation errors

**File**: `internal/controller/tenanttemplate_controller.go`

### Phase 3: Tenant Controller

**Goal**: Apply all resources defined in template with SSA.

**Implementation checklist**:
- [ ] Fetch associated TenantTemplate
- [ ] Resolve template variables (.uid, .host, etc.)
- [ ] Build dependency graph (topological sort)
- [ ] For each resource in order:
  - [ ] Render templates (Go text/template + Sprig)
  - [ ] Apply with SSA (kubectl apply --server-side)
  - [ ] Handle conflict policies (Stuck vs Force)
  - [ ] Wait for readiness (if waitForReady=true)
  - [ ] Apply timeouts
- [ ] Update Tenant status (readyResources, failedResources)
- [ ] Handle deletion policies

**File**: `internal/controller/tenant_controller.go`

### Phase 4: Template Engine

**Goal**: Centralize template rendering logic.

**Implementation checklist**:
- [ ] Create `internal/template/` package
- [ ] Implement Go text/template rendering
- [ ] Add Sprig function library
- [ ] Add custom functions:
  - [ ] `toHost(url)`: Extract hostname
  - [ ] `trunc63(str)`: Truncate to K8s limits
- [ ] Variable resolution (.uid, .host, .extraMappings)
- [ ] Error handling for template syntax errors

**File**: `internal/template/engine.go`

### Phase 5: SSA Apply Engine

**Goal**: Centralize Server-Side Apply logic.

**Implementation checklist**:
- [ ] Create `internal/apply/` package
- [ ] Implement SSA apply with client-go
- [ ] Handle conflict policies
- [ ] Set fieldManager: "tenant-operator"
- [ ] Set ownerReferences to Tenant CR
- [ ] Handle deletion policies (retain vs delete)

**File**: `internal/apply/ssa.go`

### Phase 6: Readiness Checks

**Goal**: Determine when resources are "ready".

**Implementation checklist**:
- [ ] Create `internal/readiness/` package
- [ ] Implement per-resource-type readiness logic:
  - [ ] Deployment: observedGeneration, availableReplicas
  - [ ] StatefulSet: readyReplicas
  - [ ] Job: succeeded
  - [ ] Service: (instant ready)
  - [ ] Ingress: loadBalancer.ingress exists
- [ ] Timeout handling

**File**: `internal/readiness/checker.go`

## Development Workflow

### Making Changes

1. **Edit API types**: Modify `api/v1/*_types.go`
2. **Regenerate**: Run `make generate && make manifests`
3. **Update controller**: Implement logic in `internal/controller/`
4. **Test**: Run `make test`
5. **Run locally**: `make run`
6. **Apply samples**: `kubectl apply -f config/samples/`

### Common Commands

```bash
# Regenerate code after changing types
make generate

# Regenerate CRDs after changing kubebuilder markers
make manifests

# Run all tests
make test

# Run specific test
go test -v ./internal/controller -run TestTenantRegistryReconcile

# Format code
make fmt

# Lint code
make vet

# Build Docker image
make docker-build IMG=your-registry/tenant-operator:dev

# Deploy to cluster
make deploy IMG=your-registry/tenant-operator:dev

# Uninstall from cluster
make undeploy

# Clean up
make clean
```

## Testing Strategy

### Unit Tests

Test controller logic in isolation:

```go
// internal/controller/tenantregistry_controller_test.go
func TestTenantRegistryReconcile(t *testing.T) {
    // Use envtest to create fake Kubernetes API
    // Create TenantRegistry
    // Verify Tenant CRs are created
}
```

### Integration Tests

Test with real MySQL database:

```bash
# Start MySQL in Docker
docker run --name mysql-test -e MYSQL_ROOT_PASSWORD=test -p 3306:3306 -d mysql:8

# Populate test data
mysql -h 127.0.0.1 -u root -ptest -e "CREATE DATABASE tenants; USE tenants; CREATE TABLE tenants (id INT, url VARCHAR(255), isActive BOOLEAN);"

# Run operator
make run
```

### E2E Tests

Test full workflow in Kind cluster:

```bash
# Create Kind cluster
kind create cluster --name tenant-operator-test

# Deploy operator
make deploy IMG=your-registry/tenant-operator:dev

# Apply samples
kubectl apply -f config/samples/

# Verify
kubectl get tenants -A
kubectl get deploy -A | grep tenant-
```

## Debugging

### Enable Debug Logging

```bash
# When running locally
ZAP_LOG_LEVEL=debug make run

# When deployed to cluster
kubectl set env deploy/tenant-operator-controller-manager -n tenant-operator-system ZAP_LOG_LEVEL=debug
```

### View Metrics

```bash
# Port-forward metrics service
kubectl port-forward -n tenant-operator-system svc/tenant-operator-controller-manager-metrics-service 8080:8443

# Query metrics
curl -k https://localhost:8080/metrics
```

### View Events

```bash
# Watch events
kubectl get events -n tenant-operator-system --sort-by='.lastTimestamp' -w

# Events for specific Tenant
kubectl describe tenant my-tenant
```

## Migration from Legacy Operator

See [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md) for detailed instructions on migrating from the v0.x tenant-operator.

## Resources

- **Kubebuilder Book**: https://book.kubebuilder.io/
- **Operator SDK**: https://sdk.operatorframework.io/
- **Controller Runtime**: https://pkg.go.dev/sigs.k8s.io/controller-runtime
- **Sprig Functions**: https://masterminds.github.io/sprig/
- **Server-Side Apply**: https://kubernetes.io/docs/reference/using-api/server-side-apply/

## Common Issues

### Issue: CRD validation errors

**Solution**: Regenerate CRDs after changing types:
```bash
make manifests
kubectl apply -f config/crd/bases/
```

### Issue: Template rendering fails

**Solution**: Check template syntax and available variables:
```bash
kubectl describe tenant <name>
# Look for condition: type=TemplateFailed
```

### Issue: MySQL connection refused

**Solution**: Verify database connectivity:
```bash
kubectl run -it --rm debug --image=mysql:8 --restart=Never -- mysql -h mysql.default.svc.cluster.local -u root -p
```

## Next Steps

1. Implement TenantRegistry controller to read from MySQL
2. Implement Tenant controller with template rendering and SSA
3. Add comprehensive unit tests
4. Set up CI/CD pipeline
5. Write validation webhooks (optional)
6. Add Prometheus metrics
7. Test with production-like workload (100+ tenants)

## Support

For questions or issues:
- Open an issue: https://github.com/kubernetes-tenants/tenant-operator/issues
- Review design doc: See README.md or provided design document
