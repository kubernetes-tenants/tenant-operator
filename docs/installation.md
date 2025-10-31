# Installation Guide

This guide covers various installation methods for Tenant Operator.

[[toc]]

::: tip Trying it locally?
Use the [Quick Start with Minikube](quickstart.md) guide for an automated setup tailored to first-time users.
:::

## Prerequisites

### Required

| Component | Minimum version | Notes |
| --- | --- | --- |
| Kubernetes cluster | v1.11.3+ | API compatibility tested with recent releases |
| `kubectl` | Matches cluster | Must target the cluster where you deploy |

### Production

::: warning Don't skip cert-manager
Install **cert-manager v1.13.0+** to provision webhook TLS certificates. It issues and renews certificates automatically and injects CA bundles into webhook configurations. See [Method 1](#method-1-production-install-with-cert-manager).
:::

### Optional

- **MySQL database** for tenant data source (PostgreSQL support planned for v1.1)

## Installation Methods

### Method 1: Production Install with cert-manager

**cert-manager is required** for webhook TLS certificate management.

```bash
# Step 1: Install cert-manager (if not already installed)
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Step 2: Wait for cert-manager to be ready
kubectl wait --for=condition=Available --timeout=300s -n cert-manager deployment/cert-manager
kubectl wait --for=condition=Available --timeout=300s -n cert-manager deployment/cert-manager-webhook

# Step 3: Install Tenant Operator
# cert-manager will automatically issue and manage webhook TLS certificates
kubectl apply -k https://github.com/kubernetes-tenants/tenant-operator/config/default
```

::: info What cert-manager handles
- Issues TLS certificates for the webhook server
- Renews certificates before expiration
- Injects the CA bundle into webhook configurations
- Provides battle-tested certificate automation for Kubernetes clusters
:::

### Method 2: Install from Source

```bash
# Clone repository
git clone https://github.com/kubernetes-tenants/tenant-operator.git
cd tenant-operator

# Install CRDs
make install

# Install cert-manager first if not already installed
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Deploy operator
make deploy IMG=ghcr.io/kubernetes-tenants/tenant-operator:latest
```

::: warning Remember TLS
Even when deploying from source, install cert-manager before applying the operator manifests, otherwise webhooks will fail to start.
:::

### Method 3: Local Development with Minikube

For local development, use Minikube with automated setup scripts.

See [Local Development with Minikube](local-development-minikube.md) for detailed instructions.

```bash
# Quick setup
./scripts/setup-minikube.sh      # Create cluster with cert-manager
./scripts/deploy-to-minikube.sh  # Build and deploy operator
```

## Verification

Check that the operator is running:

```bash
# Check operator deployment
kubectl get deployment -n tenant-operator-system tenant-operator-controller-manager

# Check operator logs
kubectl logs -n tenant-operator-system deployment/tenant-operator-controller-manager -f

# Verify CRDs are installed
kubectl get crd | grep operator.kubernetes-tenants.org
```

Expected output:
```
tenantregistries.operator.kubernetes-tenants.org    2025-01-15T10:00:00Z
tenants.operator.kubernetes-tenants.org             2025-01-15T10:00:00Z
tenanttemplates.operator.kubernetes-tenants.org     2025-01-15T10:00:00Z
```

::: tip Troubleshooting
If the deployment is not ready, inspect `kubectl describe deployment/tenant-operator-controller-manager` for webhook, RBAC, or image issues.
:::

## Configuration Options

### Webhook TLS Configuration

Webhook TLS is managed automatically by cert-manager. The default configuration includes:

```yaml
# config/default/kustomization.yaml
# Webhook patches are enabled by default
patches:
- path: manager_webhook_patch.yaml
- path: webhookcainjection_patch.yaml
```

::: info cert-manager responsibilities
- Issue TLS certificates for the webhook server
- Inject CA bundles into webhook configurations
- Renew certificates before expiration
:::

### Resource Limits

Adjust operator resource limits based on your cluster size:

```yaml
# config/manager/manager.yaml
resources:
  limits:
    cpu: 500m      # Increase for large clusters
    memory: 512Mi  # Increase for many tenants
  requests:
    cpu: 100m
    memory: 128Mi
```

### Concurrency Settings

Configure concurrent reconciliation workers:

```yaml
spec:
  template:
    spec:
      containers:
      - name: manager
        args:
        - --tenant-concurrency=10        # Concurrent Tenant reconciliations
        - --registry-concurrency=5       # Concurrent Registry syncs
        - --leader-elect                 # Enable leader election
```

## Multi-Platform Support

The operator supports multiple architectures:

- `linux/amd64` (Intel/AMD 64-bit)
- `linux/arm64` (ARM 64-bit, Apple Silicon)

Container images are automatically pulled for your platform.

## Namespace Isolation

By default, the operator is installed in `tenant-operator-system` namespace:

```bash
# Check operator namespace
kubectl get all -n tenant-operator-system

# View RBAC
kubectl get clusterrole | grep tenant-operator
kubectl get clusterrolebinding | grep tenant-operator
```

## Upgrading

### Upgrade CRDs First

```bash
# Upgrade CRDs (safe, preserves existing data)
make install

# Or with kubectl
kubectl apply -f config/crd/bases/
```

### Upgrade Operator

```bash
# Update operator deployment
kubectl set image -n tenant-operator-system \
  deployment/tenant-operator-controller-manager \
  manager=ghcr.io/kubernetes-tenants/tenant-operator:v1.1.0

# Or use make
make deploy IMG=ghcr.io/kubernetes-tenants/tenant-operator:v1.1.0
```

### Rolling Back

```bash
# Rollback to previous version
kubectl rollout undo -n tenant-operator-system \
  deployment/tenant-operator-controller-manager

# Check rollout status
kubectl rollout status -n tenant-operator-system \
  deployment/tenant-operator-controller-manager
```

## Uninstallation

```bash
# Delete operator deployment
kubectl delete -k config/default

# Or with make
make undeploy

# Delete CRDs (WARNING: This deletes all Tenant data!)
make uninstall

# Or with kubectl
kubectl delete crd tenantregistries.operator.kubernetes-tenants.org
kubectl delete crd tenanttemplates.operator.kubernetes-tenants.org
kubectl delete crd tenants.operator.kubernetes-tenants.org
```

**Warning:** Deleting CRDs will delete all TenantRegistry, TenantTemplate, and Tenant resources. Ensure you have backups if needed.

## Troubleshooting Installation

### Webhook TLS Errors

**Error:** `open /tmp/k8s-webhook-server/serving-certs/tls.crt: no such file or directory`

**Solution:** Install cert-manager to automatically manage webhook TLS certificates.

```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Wait for cert-manager to be ready
kubectl wait --for=condition=Available --timeout=300s -n cert-manager deployment/cert-manager

# Restart operator to pick up certificates
kubectl rollout restart -n tenant-operator-system deployment/tenant-operator-controller-manager
```

### CRD Already Exists

**Error:** `Error from server (AlreadyExists): customresourcedefinitions.apiextensions.k8s.io "tenants.operator.kubernetes-tenants.org" already exists`

**Solution:** This is normal during upgrades. CRD updates are applied automatically.

### Image Pull Errors

**Error:** `Failed to pull image "ghcr.io/kubernetes-tenants/tenant-operator:latest"`

**Solution:** Ensure your cluster can access GitHub Container Registry (ghcr.io). Check network policies and image pull secrets if needed.

### Permission Denied

**Error:** `Error from server (Forbidden): User "system:serviceaccount:tenant-operator-system:tenant-operator-controller-manager" cannot create resource`

**Solution:** Ensure RBAC resources are installed:
```bash
kubectl apply -f config/rbac/
```

## Next Steps

- [Create your first TenantRegistry](quickstart.md#step-4-deploy-tenantregistry)
- [Learn about Templates](templates.md)
- [Configure Monitoring](monitoring.md)
- [Set up Security](security.md)
