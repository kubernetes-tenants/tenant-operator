# Development Guide

Guide for developing and contributing to Tenant Operator.

**New to Tenant Operator?** Start with the [Quick Start guide](quickstart.md) to get familiar with the system before developing.

## Setup

### Prerequisites

- Go 1.22+
- kubectl
- kind or minikube (for local cluster)
- Docker
- make

### Clone Repository

```bash
git clone https://github.com/kubernetes-tenants/tenant-operator.git
cd tenant-operator
```

### Install Dependencies

```bash
go mod download
```

## Local Development

### Running Locally

```bash
# Install CRDs
make install

# Run controller locally (uses ~/.kube/config)
make run

# Run with debug logging
LOG_LEVEL=debug make run
```

The `make run` command automatically disables webhook TLS for local development.

### Testing Against Local Cluster

```bash
# Create kind cluster
kind create cluster --name tenant-operator-dev

# Install CRDs
make install

# Run operator
make run
```

## Building

### Build Binary

```bash
# Build for current platform
make build

# Binary output: bin/manager
./bin/manager --help
```

### Build Container Image

```bash
# Build image
make docker-build IMG=myregistry/tenant-operator:dev

# Push image
make docker-push IMG=myregistry/tenant-operator:dev

# Build multi-platform
docker buildx build --platform linux/amd64,linux/arm64 \
  -t myregistry/tenant-operator:dev \
  --push .
```

## Testing

### Unit Tests

```bash
# Run all unit tests
make test

# Run with coverage
make test-coverage

# View coverage report
go tool cover -html=cover.out
```

### Integration Tests

```bash
# Run integration tests (requires cluster)
make test-integration
```

### E2E Tests

```bash
# Create test cluster
kind create cluster --name e2e-test

# Run E2E tests
make test-e2e

# Cleanup
kind delete cluster --name e2e-test
```

## Code Quality

### Linting

```bash
# Run linter
make lint

# Auto-fix issues
golangci-lint run --fix
```

### Formatting

```bash
# Format code
go fmt ./...

# Or use goimports
goimports -w .
```

### Generate Code

```bash
# Generate CRD manifests, RBAC, etc.
make generate

# Generate DeepCopy methods
make manifests
```

## Project Structure

```
tenant-operator/
├── api/v1/                    # CRD types
│   ├── tenant_types.go
│   ├── tenantregistry_types.go
│   ├── tenanttemplate_types.go
│   └── common_types.go
├── internal/controller/       # Controllers
│   ├── tenant_controller.go
│   ├── tenantregistry_controller.go
│   └── tenanttemplate_controller.go
├── internal/apply/            # SSA apply engine
├── internal/database/         # Database connectors
├── internal/graph/            # Dependency graph
├── internal/readiness/        # Readiness checks
├── internal/template/         # Template engine
├── internal/metrics/          # Prometheus metrics
├── config/                    # Kustomize configs
│   ├── crd/                   # CRD manifests
│   ├── rbac/                  # RBAC configs
│   ├── manager/               # Deployment configs
│   └── samples/               # Example CRs
├── test/                      # Tests
│   ├── e2e/                   # E2E tests
│   └── utils/                 # Test utilities
├── docs/                      # Documentation
└── cmd/                       # Entry point
```

## Adding Features

### New CRD Field

1. Update API types:
```go
// api/v1/tenant_types.go
type TenantSpec struct {
    NewField string `json:"newField,omitempty"`
}
```

2. Generate code:
```bash
make generate
make manifests
```

3. Update controller logic

4. Add tests

5. Update documentation

### New Controller

1. Create controller file:
```go
// internal/controller/myresource_controller.go
package controller

type MyResourceReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

func (r *MyResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Implementation
}
```

2. Register controller:
```go
// cmd/main.go
if err = (&controller.MyResourceReconciler{
    Client: mgr.GetClient(),
    Scheme: mgr.GetScheme(),
}).SetupWithManager(mgr); err != nil {
    // Handle error
}
```

3. Add tests

## Contributing

### Workflow

1. Fork repository
2. Create feature branch
3. Make changes
4. Add tests
5. Run linter: `make lint`
6. Run tests: `make test`
7. Commit with conventional commits
8. Open Pull Request

### Conventional Commits

```
feat: add new feature
fix: fix bug
docs: update documentation
test: add tests
refactor: refactor code
chore: maintenance tasks
```

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests passing
```

## Release Process

### Version Bump

1. Update version in:
   - `README.md`
   - `config/manager/kustomization.yaml`

2. Generate changelog

3. Create git tag:
```bash
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0
```

4. GitHub Actions builds and publishes release

## Useful Commands

```bash
# Install CRDs
make install

# Uninstall CRDs
make uninstall

# Deploy operator
make deploy IMG=<image>

# Undeploy operator
make undeploy

# Run locally
make run

# Build binary
make build

# Build container
make docker-build IMG=<image>

# Run tests
make test

# Run linter
make lint

# Generate code
make generate manifests
```

## See Also

- [Contributing Guide](../CONTRIBUTING.md)
- [API Reference](api.md)
- [Architecture](../README.md#architecture)
