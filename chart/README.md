# Tenant Operator Helm Chart

Official Helm chart for deploying the Tenant Operator to Kubernetes clusters.

## Overview

The Tenant Operator is a Kubernetes operator that automates the provisioning, configuration, and lifecycle management of multi-tenant applications. This Helm chart provides a simple and configurable way to deploy the operator to your cluster.

## Prerequisites

- **Kubernetes** 1.23+
- **Helm** 3.8+
- **cert-manager v1.13+** ⚠️ **REQUIRED for all environments** (webhook certificates)
  - Webhooks provide validation and defaulting for CRDs
  - Installation: `kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml`
  - **Must be installed BEFORE installing tenant-operator**

### Why cert-manager is Required

The Tenant Operator uses **admission webhooks** for:
- ✅ **Validation**: Prevents invalid configurations at admission time
- ✅ **Defaulting**: Automatically sets sensible defaults for CRDs
- ✅ **Integrity**: Ensures referential integrity (e.g., TenantTemplate → TenantRegistry)

Webhooks require **TLS certificates** for secure communication with the Kubernetes API server. cert-manager automates certificate provisioning and renewal.

**cert-manager is now REQUIRED in all environments** (including local development) to ensure consistency and prevent invalid configurations from being applied.

**📚 For detailed information**, see [CERT-MANAGER.md](./CERT-MANAGER.md) - Comprehensive guide on cert-manager dependency, troubleshooting, and FAQ.

## Installation

### Add Helm Repository

```bash
helm repo add tenant-operator https://kubernetes-tenants.github.io/tenant-operator
helm repo update
```

### Install the Chart

#### Quick Start (Local Development)

For local development with minikube or kind:

```bash
# Step 1: Install cert-manager (REQUIRED)
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Wait for cert-manager to be ready
kubectl wait --for=condition=Available --timeout=300s -n cert-manager \
  deployment/cert-manager \
  deployment/cert-manager-webhook \
  deployment/cert-manager-cainjector

# Step 2: Install tenant-operator with local values
helm install tenant-operator tenant-operator/tenant-operator \
  -f https://raw.githubusercontent.com/kubernetes-tenants/tenant-operator/main/chart/values-local.yaml \
  --namespace tenant-operator-system \
  --create-namespace
```

**Note**: Local values use lower resource requirements and self-signed certificates, but webhooks remain enabled for consistency with production.

#### Production Installation

For production environments with webhooks and monitoring:

```bash
# ⚠️  STEP 1: Install cert-manager FIRST (REQUIRED)
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Wait for cert-manager to be ready (IMPORTANT: Do not proceed until ready)
kubectl wait --for=condition=Available --timeout=300s deployment/cert-manager -n cert-manager
kubectl wait --for=condition=Available --timeout=300s deployment/cert-manager-webhook -n cert-manager
kubectl wait --for=condition=Available --timeout=300s deployment/cert-manager-cainjector -n cert-manager

# Verify cert-manager is working
kubectl get pods -n cert-manager

# STEP 2: Install tenant-operator
helm install tenant-operator tenant-operator/tenant-operator \
  -f https://raw.githubusercontent.com/kubernetes-tenants/tenant-operator/main/chart/values-prod.yaml \
  --namespace tenant-operator-system \
  --create-namespace
```

**⚠️ CRITICAL**: If you skip Step 1, the installation will fail because webhook certificates cannot be provisioned.

#### Custom Installation

```bash
helm install tenant-operator tenant-operator/tenant-operator \
  --set image.tag=v1.0.0 \
  --set webhook.enabled=true \
  --set monitoring.enabled=true \
  --namespace tenant-operator-system \
  --create-namespace
```

## Uninstallation

```bash
helm uninstall tenant-operator -n tenant-operator-system
```

**Note**: CRDs are not deleted automatically. To remove them:

```bash
kubectl delete crd tenantregistries.operator.kubernetes-tenants.org
kubectl delete crd tenanttemplates.operator.kubernetes-tenants.org
kubectl delete crd tenants.operator.kubernetes-tenants.org
```

## Upgrading

```bash
# Update repository
helm repo update

# Upgrade release
helm upgrade tenant-operator tenant-operator/tenant-operator \
  --namespace tenant-operator-system
```

## Configuration

### Key Configuration Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of operator replicas | `1` |
| `image.registry` | Container image registry | `ghcr.io` |
| `image.repository` | Container image repository | `kubernetes-tenants/tenant-operator` |
| `image.tag` | Container image tag | `""` (uses Chart appVersion) |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `webhook.enabled` | Enable admission webhooks | `true` |
| `certManager.enabled` | Enable cert-manager integration | `true` |
| `monitoring.enabled` | Enable Prometheus ServiceMonitor | `false` |
| `rbac.create` | Create RBAC resources | `true` |
| `serviceAccount.create` | Create ServiceAccount | `true` |

### Resource Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `128Mi` |
| `resources.requests.cpu` | CPU request | `10m` |
| `resources.requests.memory` | Memory request | `64Mi` |

### Webhook Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `webhook.enabled` | Enable webhooks | `true` |
| `webhook.port` | Webhook server port | `9443` |
| `webhook.certificate.issuerName` | Certificate issuer name | `selfsigned-issuer` |
| `webhook.certificate.issuerKind` | Certificate issuer kind | `Issuer` |
| `webhook.certificate.duration` | Certificate duration | `2160h` (90 days) |

### Monitoring Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `monitoring.enabled` | Enable Prometheus ServiceMonitor | `false` |
| `monitoring.interval` | Scrape interval | `30s` |
| `monitoring.scrapeTimeout` | Scrape timeout | `10s` |

### Full Configuration

See [values.yaml](./values.yaml) for all available configuration options.

## Examples

### Example 1: Custom Resource Limits

```bash
# Install cert-manager first (if not already installed)
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Install with custom resource limits
helm install tenant-operator tenant-operator/tenant-operator \
  --set resources.limits.cpu=1000m \
  --set resources.limits.memory=256Mi \
  --namespace tenant-operator-system \
  --create-namespace
```

### Example 2: Production with Monitoring

```bash
helm install tenant-operator tenant-operator/tenant-operator \
  --set webhook.enabled=true \
  --set certManager.enabled=true \
  --set monitoring.enabled=true \
  --set monitoring.labels.prometheus=kube-prometheus \
  --set resources.limits.cpu=1000m \
  --set resources.limits.memory=512Mi \
  --namespace tenant-operator-system \
  --create-namespace
```

### Example 3: High Availability

```bash
helm install tenant-operator tenant-operator/tenant-operator \
  --set replicaCount=3 \
  --set affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].weight=100 \
  --set affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.topologyKey=kubernetes.io/hostname \
  --namespace tenant-operator-system \
  --create-namespace
```

### Example 4: Custom Image

```bash
helm install tenant-operator tenant-operator/tenant-operator \
  --set image.registry=my-registry.io \
  --set image.repository=my-org/tenant-operator \
  --set image.tag=custom-v1.0.0 \
  --namespace tenant-operator-system \
  --create-namespace
```

## Environment-Specific Values

### Local Development (`values-local.yaml`)

Optimized for minikube, kind, or k3d:
- ✅ Webhooks enabled (validation & defaulting)
- ✅ cert-manager required (must be pre-installed)
- Lower resource requirements
- Faster probe settings
- Self-signed certificates

```bash
# Install cert-manager first (REQUIRED)
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
kubectl wait --for=condition=Available --timeout=300s -n cert-manager deployment/cert-manager-webhook

# Install tenant-operator
helm install tenant-operator tenant-operator/tenant-operator \
  -f values-local.yaml \
  --namespace tenant-operator-system \
  --create-namespace
```

### Production (`values-prod.yaml`)

Production-ready configuration:
- ✅ Webhooks enabled
- ✅ cert-manager required (must be pre-installed)
- ✅ Prometheus monitoring enabled
- ✅ Network policies enabled
- Higher resource limits
- Pod anti-affinity for HA

```bash
# Install cert-manager first
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
kubectl wait --for=condition=Available --timeout=300s -n cert-manager deployment/cert-manager-webhook

# Install tenant-operator
helm install tenant-operator tenant-operator/tenant-operator \
  -f values-prod.yaml \
  --namespace tenant-operator-system \
  --create-namespace
```

## Troubleshooting

### Operator Pod Not Starting

Check logs:
```bash
kubectl logs -n tenant-operator-system -l control-plane=controller-manager
```

### Webhook Certificate Issues

**Symptom**: Operator pod fails to start with webhook certificate errors, or ValidatingWebhookConfiguration shows no CA bundle.

**Diagnosis**:
```bash
# Check if cert-manager is installed
kubectl get pods -n cert-manager

# Check if certificate is created
kubectl get certificate -n tenant-operator-system

# Check certificate details
kubectl describe certificate -n tenant-operator-system

# Check if secret is created
kubectl get secret -n tenant-operator-system | grep cert
```

**Solution**:
1. If cert-manager is not installed:
   ```bash
   kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
   kubectl wait --for=condition=Available --timeout=300s -n cert-manager deployment/cert-manager-webhook
   ```

2. If certificate is not ready, check cert-manager logs:
   ```bash
   kubectl logs -n cert-manager -l app=cert-manager
   ```

3. If you don't want to use cert-manager, disable webhooks:
   ```bash
   helm upgrade tenant-operator tenant-operator/tenant-operator \
     --set webhook.enabled=false \
     --set certManager.enabled=false \
     --namespace tenant-operator-system
   ```

### CRDs Not Installed

Check if CRDs exist:
```bash
kubectl get crd | grep operator.kubernetes-tenants.org
```

If missing, manually install:
```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes-tenants/tenant-operator/main/config/crd/bases/
```

## Development

### Testing Locally

```bash
# Lint the chart
helm lint ./chart

# Dry-run installation
helm install tenant-operator ./chart \
  --dry-run --debug \
  --namespace tenant-operator-system

# Template rendering
helm template tenant-operator ./chart \
  -f ./chart/values-local.yaml
```

## Support

- **Documentation**: https://kubernetes-tenants.github.io/tenant-operator
- **GitHub**: https://github.com/kubernetes-tenants/tenant-operator
- **Issues**: https://github.com/kubernetes-tenants/tenant-operator/issues
- **Discussions**: https://github.com/kubernetes-tenants/tenant-operator/discussions

## License

Apache License 2.0
