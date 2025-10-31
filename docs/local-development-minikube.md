# Local Development with Minikube

Development workflow guide for contributing to and modifying Tenant Operator.

[[toc]]

::: tip Quick taste
If you only want to experience the operator, follow the [Quick Start guide](quickstart.md) instead.
:::

## Overview

This guide covers the **development workflow** for making code changes to Tenant Operator and testing them locally on Minikube.

**Use this guide when you want to:**
- ✅ Modify Tenant Operator source code
- ✅ Add new features or fix bugs
- ✅ Test code changes locally before committing
- ✅ Debug the operator with breakpoints
- ✅ Iterate quickly on code changes

**For initial setup:** Follow the [Quick Start guide](quickstart.md) first to set up your Minikube environment.

## Prerequisites

::: info Prerequisites
Complete the [Quick Start guide](quickstart.md) first. You should have:
- ✅ Minikube cluster running (`tenant-operator` profile)
- ✅ **cert-manager installed** (automatically by setup script)
- ✅ Tenant Operator deployed with webhooks enabled
- ✅ MySQL test database (optional, for full testing)

::: warning cert-manager Required
cert-manager is **REQUIRED** for all installations. The automated setup scripts install it automatically. If setting up manually, install cert-manager before deploying the operator.
:::

Additional development tools:
- **Go** 1.22+
- **make**
- **golangci-lint** (optional, for linting)
- **delve** (optional, for debugging)

## Development Workflow

### Typical Development Cycle

```bash
# 1. Make code changes
vim internal/controller/tenant_controller.go

# 2. Run unit tests
make test

# 3. Run linter
make lint

# 4. Build and deploy to Minikube
./scripts/deploy-to-minikube.sh

# 5. View operator logs
kubectl logs -n tenant-operator-system -l control-plane=controller-manager -f

# 6. Test changes
kubectl apply -f config/samples/

# 7. Verify results
kubectl get tenants
kubectl get all -n tenant-<uid>

# 8. Repeat steps 1-7 as needed
```

::: tip Iteration speed
Expect roughly 1–2 minutes per build + deploy cycle when using the provided scripts.
:::

## Code Changes & Rebuilding

### Quick Rebuild and Deploy

After making code changes:

```bash
# Rebuild and redeploy operator
./scripts/deploy-to-minikube.sh
```

This script:
1. Builds new Docker image with timestamp tag
2. Loads image into Minikube's internal registry
3. Updates operator deployment
4. Waits for readiness

**Why timestamp tags?** Each deployment gets a unique tag, preventing Kubernetes from using cached old images.

### Custom Image Tag

Use a custom tag for easier identification:

```bash
IMG=tenant-operator:my-feature ./scripts/deploy-to-minikube.sh
```

### Manual Build (if needed)

```bash
# Build binary locally
make build

# Build Docker image
make docker-build IMG=tenant-operator:dev

# Load into Minikube
minikube -p tenant-operator image load tenant-operator:dev
```

## Running Operator Locally (Outside Cluster)

For fastest iteration, run the operator locally on your machine while connecting to the Minikube cluster:

```bash
# Ensure CRDs are installed
make install

# Run operator locally
make run
```

::: tip Benefits
- ✅ Instant restarts (no image build/load)
- ✅ Direct Go debugging with breakpoints
- ✅ Real-time logs in your terminal
- ✅ Fast feedback loop (~5 seconds)
:::

::: warning Limitations
- ⚠️ **Webhooks unavailable** (TLS certificates require in-cluster deployment with cert-manager)
- ⚠️ **No validation at admission time** (changes are only validated at reconciliation)
- ⚠️ **No automatic defaulting** (must specify all fields manually)
- ⚠️ Runtime differs from production environment
:::

**When to use:**
- Controller logic changes
- Quick iteration on reconciliation loops
- Debugging with delve

**When NOT to use:**
- Testing webhooks (requires full deployment with cert-manager)
- Testing validation/defaulting behavior
- Verifying in-cluster networking
- Final testing before PR

::: tip Testing with Webhooks
For complete testing including webhooks, always deploy to cluster:
```bash
./scripts/deploy-to-minikube.sh  # Includes cert-manager and webhooks
```
:::

**Testing with local run:**
```bash
# Terminal 1: Run operator
make run

# Terminal 2: Apply resources
kubectl apply -f config/samples/
kubectl get tenants --watch

# Terminal 3: View database changes
kubectl exec -it deployment/mysql -n tenant-operator-test -- \
  mysql -u tenant_reader -p tenants -e "SELECT * FROM tenant_configs;"
```

## Debugging

### Debug with Delve

Run operator with debugger:

```bash
# Install delve if needed
go install github.com/go-delve/delve/cmd/dlv@latest

# Run with delve
dlv debug ./cmd/main.go -- --zap-devel=true
```

Then in delve:
```
(dlv) break internal/controller/tenant_controller.go:123
(dlv) continue
```

### Debug Operator Logs

View operator logs with different verbosity:

```bash
# Default logs
kubectl logs -n tenant-operator-system -l control-plane=controller-manager -f

# Filter for specific tenant
kubectl logs -n tenant-operator-system -l control-plane=controller-manager | grep acme-corp

# Follow logs for errors only
kubectl logs -n tenant-operator-system -l control-plane=controller-manager -f | grep -i error

# View logs from previous crash
kubectl logs -n tenant-operator-system -l control-plane=controller-manager --previous
```

### Debug Test Resources

View what the operator sees:

```bash
# Check Tenant CR status
kubectl get tenant acme-corp-test-template -o yaml

# Check registry sync status
kubectl get tenantregistry test-registry -o yaml | yq '.status'

# Check template
kubectl get tenanttemplate test-template -o yaml

# View events
kubectl get events --sort-by='.lastTimestamp' -n tenant-operator-system

# Describe resource for events
kubectl describe tenant acme-corp-test-template
```

## Testing

### Unit Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test ./internal/controller/... -v

# Run specific test
go test ./internal/controller/ -run TestTenantController_Reconcile -v
```

### Integration Tests

```bash
# Requires running Minikube cluster
make test-integration
```

### E2E Testing

Test complete workflow:

```bash
# 1. Deploy fresh operator
./scripts/deploy-to-minikube.sh

# 2. Deploy test database
./scripts/deploy-mysql.sh

# 3. Deploy test registry and template
./scripts/deploy-tenantregistry.sh
./scripts/deploy-tenanttemplate.sh

# 4. Verify tenants created
kubectl get tenants
kubectl get deployments,services -l kubernetes-tenants.org/tenant

# 5. Test lifecycle: Add tenant
kubectl exec -it deployment/mysql -n tenant-operator-test -- \
  mysql -u root -p tenants -e \
  "INSERT INTO tenant_configs VALUES ('delta-co', 'https://delta.example.com', 1, 'enterprise');"

# Wait 30s, then verify
kubectl get tenant delta-co-test-template

# 6. Test lifecycle: Deactivate tenant
kubectl exec -it deployment/mysql -n tenant-operator-test -- \
  mysql -u root -p tenants -e \
  "UPDATE tenant_configs SET is_active = 0 WHERE tenant_id = 'acme-corp';"

# Wait 30s, then verify deletion
kubectl get tenant acme-corp-test-template
# Should be NotFound
```

## Tips for Fast Iteration

### 1. Skip Image Build for Controller Changes

If only changing controller logic (not CRDs, RBAC, etc.):

```bash
# Run locally instead of deploying
make run
```

**~10x faster** than full build/deploy cycle.

### 2. Keep Logs Open

```bash
# In a dedicated terminal
kubectl logs -n tenant-operator-system -l control-plane=controller-manager -f
```

### 3. Use Watch Commands

```bash
# Watch tenants
watch kubectl get tenants

# Watch specific tenant
watch kubectl get tenant acme-corp-test-template -o yaml
```

### 4. Quick MySQL Queries

Create aliases:

```bash
alias mysql-test='kubectl exec -it deployment/mysql -n tenant-operator-test -- mysql -u tenant_reader -p$(kubectl get secret mysql-credentials -n tenant-operator-test -o jsonpath="{.data.password}" | base64 -d) tenants'

# Then use:
mysql-test -e "SELECT * FROM tenant_configs;"
```

### 5. Fast Context Switching

```bash
# Add to ~/.zshrc or ~/.bashrc
alias kto='kubectl config use-context tenant-operator'
alias ktos='kubectl -n tenant-operator-system'
alias ktot='kubectl -n tenant-operator-test'

# Usage:
kto  # Switch to tenant-operator context
ktos get pods  # Get pods in operator namespace
```

## Common Development Scenarios

### Scenario 1: Testing Template Changes

```bash
# 1. Modify template logic in tenant_controller.go
vim internal/controller/tenant_controller.go

# 2. Run locally for quick feedback
make run

# 3. In another terminal, apply test template
kubectl apply -f config/samples/operator_v1_tenanttemplate.yaml

# 4. Watch logs and verify rendered resources
kubectl logs -n tenant-operator-system -l control-plane=controller-manager -f
kubectl get tenant -o yaml | grep -A 10 "spec:"
```

### Scenario 2: Testing Database Sync

```bash
# 1. Modify registry controller
vim internal/controller/tenantregistry_controller.go

# 2. Deploy to test in-cluster
./scripts/deploy-to-minikube.sh

# 3. Change database and watch sync
mysql-test -e "UPDATE tenant_configs SET subscription_plan = 'premium' WHERE tenant_id = 'acme-corp';"

# 4. Verify Tenant CR updated
kubectl get tenant acme-corp-test-template -o yaml | grep planId
```

### Scenario 3: Testing CRD Changes

```bash
# 1. Modify CRD in api/v1/
vim api/v1/tenant_types.go

# 2. Regenerate manifests
make manifests

# 3. Install updated CRDs
make install

# 4. Rebuild and deploy operator
./scripts/deploy-to-minikube.sh

# 5. Test with updated CRD
kubectl apply -f config/samples/
```

### Scenario 4: Testing Webhook Validation

```bash
# 1. Modify webhook in api/v1/*_webhook.go
vim api/v1/tenanttemplate_webhook.go

# 2. Must deploy to cluster (webhooks need TLS)
./scripts/deploy-to-minikube.sh

# 3. Test invalid resource
kubectl apply -f - <<EOF
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: invalid-template
spec:
  registryId: non-existent-registry  # Should fail validation
EOF

# 4. Should see validation error
```

## Cleanup

### Partial Cleanup (Keep Cluster)

```bash
# Delete test resources
kubectl delete tenants --all
kubectl delete tenanttemplate test-template
kubectl delete tenantregistry test-registry

# Delete MySQL
kubectl delete deployment,service,pvc mysql -n tenant-operator-test

# Delete operator
kubectl delete deployment tenant-operator-controller-manager -n tenant-operator-system
```

### Full Cleanup

```bash
# Delete everything including cluster
./scripts/cleanup-minikube.sh

# Answer 'y' to all prompts for complete cleanup
```

### Fresh Start

```bash
# Complete reset
./scripts/cleanup-minikube.sh  # Delete everything
./scripts/setup-minikube.sh    # Recreate cluster
./scripts/deploy-to-minikube.sh  # Deploy operator
```

## Troubleshooting

### Operator Won't Start

```bash
# Check pod status
kubectl get pods -n tenant-operator-system

# Check logs
kubectl logs -n tenant-operator-system -l control-plane=controller-manager

# Common issues:

# 1. cert-manager not ready
kubectl get pods -n cert-manager
# If pods are not running, wait or check:
kubectl describe pods -n cert-manager

# 2. Webhook certificates not ready
kubectl get certificate -n tenant-operator-system
# Should show "Ready=True"

# 3. Image not loaded
minikube -p tenant-operator image ls | grep tenant-operator

# 4. CRDs not installed
kubectl get crd | grep tenant
```

::: danger cert-manager is Critical
If the operator pod fails to start with webhook certificate errors, cert-manager is likely not installed or not ready. Check:

```bash
# Verify cert-manager installation
kubectl get pods -n cert-manager

# Check certificate status
kubectl get certificate -n tenant-operator-system
kubectl describe certificate -n tenant-operator-system

# If missing, install cert-manager:
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
kubectl wait --for=condition=Available --timeout=300s -n cert-manager deployment/cert-manager-webhook
```
:::

### Tests Failing

```bash
# Ensure test cluster is accessible
kubectl cluster-info

# Check if CRDs are installed
kubectl get crd | grep tenant

# Run tests with verbose output
go test ./... -v -count=1
```

### Image Not Updating

```bash
# Force rebuild without cache
docker build --no-cache -t tenant-operator:dev .

# Reload into Minikube
minikube -p tenant-operator image load tenant-operator:dev

# Restart operator pod
kubectl rollout restart deployment -n tenant-operator-system tenant-operator-controller-manager
```

## Advanced Workflows

### Multiple Minikube Profiles

Work on multiple features simultaneously:

```bash
# Feature A cluster
MINIKUBE_PROFILE=feature-a ./scripts/setup-minikube.sh
MINIKUBE_PROFILE=feature-a ./scripts/deploy-to-minikube.sh

# Feature B cluster
MINIKUBE_PROFILE=feature-b ./scripts/setup-minikube.sh
MINIKUBE_PROFILE=feature-b ./scripts/deploy-to-minikube.sh

# Switch between them
kubectl config use-context feature-a
kubectl config use-context feature-b
```

### Custom Resource Allocations

```bash
# More powerful cluster for load testing
MINIKUBE_CPUS=8 \
MINIKUBE_MEMORY=16384 \
./scripts/setup-minikube.sh
```

## See Also

- [Quick Start](quickstart.md) - Initial setup guide
- [Development Guide](development.md) - General development practices
- [Contributing](https://github.com/kubernetes-tenants/tenant-operator/blob/main/CONTRIBUTING.md) - Contribution guidelines
- [Troubleshooting](troubleshooting.md) - Common issues

## Summary

**Fast iteration workflow:**
1. Make code changes
2. Run `make run` for controller changes (5s feedback)
3. Or run `./scripts/deploy-to-minikube.sh` for full testing (2min)
4. Test with `kubectl apply -f config/samples/`
5. Iterate

**Key takeaways:**
- Use `make run` for fastest iteration (no webhooks)
- Use `./scripts/deploy-to-minikube.sh` for full testing (with webhooks)
- Keep logs open in a separate terminal
- Test E2E with real MySQL database
- Clean up and reset when needed

Happy coding! 🚀
