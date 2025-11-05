# cert-manager Dependency Guide

## Why is cert-manager Required?

The Tenant Operator uses **Kubernetes Admission Webhooks** to provide:

1. **Validation Webhooks**
   - Prevents invalid TenantRegistry/TenantTemplate configurations
   - Validates template syntax at admission time (before reconciliation)
   - Ensures referential integrity (TenantTemplate references valid TenantRegistry)
   - Validates value mappings (uid, hostOrUrl, activate are required)

2. **Mutating Webhooks (Defaulting)**
   - Automatically sets default values for optional fields
   - Simplifies user experience (less boilerplate)
   - Ensures consistent behavior across resources

## Why Webhooks Need TLS Certificates

Kubernetes API server **requires** all admission webhooks to use TLS (HTTPS) for security:
- Prevents man-in-the-middle attacks
- Ensures authenticity of webhook server
- Required by Kubernetes admission controller architecture

**cert-manager** automates:
- Certificate provisioning (CSR, CA signing)
- Automatic certificate renewal before expiration
- CA bundle injection into webhook configurations

## Installation Options

### Standard Installation (cert-manager Required) ✅

**cert-manager is now REQUIRED for all installations** to ensure consistency and validation across all environments.

```bash
# Step 1: Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Step 2: Wait for cert-manager to be ready
kubectl wait --for=condition=Available --timeout=300s -n cert-manager \
  deployment/cert-manager \
  deployment/cert-manager-webhook \
  deployment/cert-manager-cainjector

# Step 3: Verify
kubectl get pods -n cert-manager

# Step 4: Install tenant-operator
helm install tenant-operator tenant-operator/tenant-operator \
  --namespace tenant-operator-system \
  --create-namespace
```

**Result**: Full validation, defaulting, consistent behavior across all environments

### Local Development

Use the same steps as above, but with `values-local.yaml` for optimized resource settings:

```bash
# Install cert-manager (same as above)
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
kubectl wait --for=condition=Available --timeout=300s -n cert-manager deployment/cert-manager-webhook

# Install with local development values
helm install tenant-operator tenant-operator/tenant-operator \
  -f https://raw.githubusercontent.com/kubernetes-tenants/tenant-operator/main/chart/values-local.yaml \
  --namespace tenant-operator-system \
  --create-namespace
```

**Benefits**:
- ✅ Validation at admission time (catches errors early)
- ✅ Automatic defaulting (less boilerplate)
- ✅ Consistent behavior with production
- ✅ Lower resource requirements for local clusters

---

## What Happens Without cert-manager?

### Scenario: Webhooks Enabled, cert-manager Not Installed ❌

```bash
# This configuration will FAIL
helm install tenant-operator tenant-operator/tenant-operator \
  --set webhook.enabled=true \
  --set certManager.enabled=true
# (but cert-manager is not actually installed in the cluster)
```

**Result**:
1. Certificate resource is created
2. Certificate remains in "Pending" state (no cert-manager to fulfill it)
3. Webhook secret is never created
4. Operator pod fails to start (cannot mount webhook certificate)
5. ValidatingWebhookConfiguration has no CA bundle (admission fails)

**Error logs**:
```
Error: secret "tenant-operator-serving-cert" not found
```

**Fix**:
```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
kubectl wait --for=condition=Available --timeout=300s -n cert-manager deployment/cert-manager-webhook

# Wait for certificate to be issued
kubectl wait --for=condition=Ready --timeout=60s -n tenant-operator-system \
  certificate/tenant-operator-serving-cert

# Restart operator pod to pick up certificate
kubectl rollout restart deployment -n tenant-operator-system
```

---

## Why We Require cert-manager for All Environments

Previously, we allowed disabling webhooks for local development. **We no longer support this** because:

1. **Consistency**: Development should match production behavior
2. **Early Detection**: Validation catches errors at admission time (faster feedback)
3. **Best Practices**: Encourages proper configuration from the start
4. **Simplified Support**: One deployment path reduces confusion
5. **cert-manager is Standard**: Most Kubernetes clusters already have cert-manager installed

### What You Would Lose Without Webhooks

**Example of invalid config that would be accepted without webhooks:**

```yaml
# This INVALID config would be accepted (missing required fields)
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantRegistry
metadata:
  name: broken-registry
spec:
  source:
    type: mysql
  # Missing: host, username, database, valueMappings
  # Operator will fail at reconciliation time (not admission time)
```

**With webhooks (properly rejected):**
```
Error from server (Forbidden): error when creating "broken-registry.yaml":
admission webhook "vtenantregistry.kb.io" denied the request:
spec.source.mysql: Required value
spec.valueMappings: Required value
```

---

## FAQ

### Q: Can I use my own certificate instead of cert-manager?

**A: Technically yes, but strongly discouraged**. You would need to:
1. Manually create TLS secret with your certificate
2. Manually inject CA bundle into webhook configurations
3. Manually renew certificates before expiration
4. Set `certManager.enabled=false` to skip Certificate resource creation

```bash
# Create secret manually
kubectl create secret tls tenant-operator-serving-cert \
  --cert=path/to/cert.pem \
  --key=path/to/key.pem \
  -n tenant-operator-system

# Install without cert-manager resources
helm install tenant-operator tenant-operator/tenant-operator \
  --set webhook.enabled=true \
  --set certManager.enabled=false
```

**⚠️ Not recommended**:
- Manual certificate renewal is error-prone
- cert-manager is the Kubernetes standard
- Adds unnecessary operational overhead

---

### Q: Can I disable webhooks for local development?

**A: Not supported anymore**. We require webhooks in all environments for consistency.

cert-manager is lightweight and easy to install in local clusters (minikube, kind, k3d):
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

**Benefits of always using webhooks:**
- Catches configuration errors immediately
- Consistent behavior across all environments
- Validates template syntax before reconciliation
- Enforces best practices from day one

---

### Q: What if cert-manager is already installed in my cluster?

**A: Perfect!** Just install the operator normally:

```bash
# Verify cert-manager is present
kubectl get pods -n cert-manager

# Install tenant-operator (will use existing cert-manager)
helm install tenant-operator tenant-operator/tenant-operator
```

The chart **does not install cert-manager** as a dependency (by design). It assumes cert-manager is pre-installed.

---

### Q: Can I use a different cert-manager version?

**A: Yes**, but ensure compatibility:
- cert-manager v1.0+ (API version `cert-manager.io/v1`)
- Recommended: v1.13+ (latest stable)

The operator's Certificate resource uses `cert-manager.io/v1` API.

---

### Q: What about clusters without internet access (air-gapped)?

**A: You need to**:
1. Pre-load cert-manager images into your private registry
2. Install cert-manager from your registry
3. Install tenant-operator

```bash
# Example for air-gapped environments
helm install cert-manager jetstack/cert-manager \
  --set image.repository=my-registry.com/cert-manager-controller \
  --set webhook.image.repository=my-registry.com/cert-manager-webhook \
  --set cainjector.image.repository=my-registry.com/cert-manager-cainjector

helm install tenant-operator tenant-operator/tenant-operator \
  --set image.registry=my-registry.com
```

---

## Verification Checklist

After installation, verify everything is working:

```bash
# 1. Check cert-manager pods
kubectl get pods -n cert-manager
# Expected: cert-manager, cert-manager-webhook, cert-manager-cainjector (all Running)

# 2. Check certificate is issued
kubectl get certificate -n tenant-operator-system
# Expected: tenant-operator-serving-cert (Ready=True)

# 3. Check secret is created
kubectl get secret tenant-operator-serving-cert -n tenant-operator-system
# Expected: Type=kubernetes.io/tls

# 4. Check webhook configurations
kubectl get validatingwebhookconfiguration | grep tenant-operator
kubectl get mutatingwebhookconfiguration | grep tenant-operator
# Expected: CA bundle is populated (not empty)

# 5. Test webhook validation
kubectl apply -f - <<EOF
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantRegistry
metadata:
  name: test-invalid
spec:
  # Missing required fields - should be rejected
EOF
# Expected: Error from server (admission webhook denied the request)
```

If any step fails, see [Troubleshooting](#troubleshooting) in the main README.

---

## Summary

| Scenario | cert-manager | webhook.enabled | Result | Use Case |
|----------|--------------|-----------------|--------|----------|
| **Standard (All Environments)** ✅ | Installed | true | Full validation & defaulting | All deployments (prod, dev, test) |
| **Broken Configuration** ❌ | Not installed | true | Installation fails | Misconfiguration - must install cert-manager first |

**Policy**: cert-manager is **required** for all installations to ensure consistency and validation across all environments.
