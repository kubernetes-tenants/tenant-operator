# Container Image Deployment

This project automatically builds and publishes container images to GitHub Container Registry (GHCR).

## Image Location

```
ghcr.io/kubernetes-tenants/tenant-operator
```

## Automated Builds

The `build-push.yml` workflow automatically builds and publishes images on:

### 1. **Main Branch Push**
- Triggered on every push to `main` branch
- Tags: `main`, `latest`, `main-<git-sha>`
- Platforms: `linux/amd64`, `linux/arm64`

### 2. **Version Tags**
- Triggered on version tags (e.g., `v1.0.0`, `v1.2.3`)
- Tags: `v1.0.0`, `v1.0`, `v1`, `latest`
- Platforms: `linux/amd64`, `linux/arm64`

### 3. **Pull Requests**
- Build-only (no push to registry)
- Validates that the image builds successfully

## Using the Image

### Pull the latest image
```bash
docker pull ghcr.io/kubernetes-tenants/tenant-operator:latest
```

### Pull a specific version
```bash
docker pull ghcr.io/kubernetes-tenants/tenant-operator:v1.0.0
```

### Deploy to Kubernetes
```bash
# Using kubectl
kubectl set image deployment/tenant-operator \
  manager=ghcr.io/kubernetes-tenants/tenant-operator:v1.0.0

# Or update your manifests
# config/manager/manager.yaml
spec:
  template:
    spec:
      containers:
      - name: manager
        image: ghcr.io/kubernetes-tenants/tenant-operator:v1.0.0
```

## Authentication

### Public Access
Images are publicly accessible for pulling. No authentication required.

## Image Tagging Strategy

| Event | Tags Generated | Example |
|-------|---------------|---------|
| Push to `main` | `main`, `latest`, `main-<sha>` | `main`, `latest`, `main-a1b2c3d` |
| Tag `v1.2.3` | `v1.2.3`, `v1.2`, `v1`, `latest` | `v1.2.3`, `v1.2`, `v1` |
| PR #123 | `pr-123` | `pr-123` (build only, not pushed) |

## Multi-Platform Support

All images are built for multiple platforms:
- `linux/amd64` (Intel/AMD 64-bit)
- `linux/arm64` (ARM 64-bit, including Apple Silicon)

Docker will automatically pull the correct image for your platform.

## Build Provenance

Images include SLSA build provenance attestations for supply chain security.

You can verify the attestation:
```bash
# Install cosign
# https://docs.sigstore.dev/cosign/installation/

# Verify provenance
cosign verify-attestation \
  --type slsaprovenance \
  ghcr.io/kubernetes-tenants/tenant-operator:v1.0.0
```

## Manual Build (for testing)

```bash
# Build locally
make docker-build IMG=ghcr.io/kubernetes-tenants/tenant-operator:dev

# Push manually (requires authentication)
make docker-push IMG=ghcr.io/kubernetes-tenants/tenant-operator:dev
```

## Troubleshooting

### Image not found
- Check if the repository is public: https://github.com/kubernetes-tenants/tenant-operator/pkgs/container/tenant-operator
- Ensure you're using the correct image name: `ghcr.io/kubernetes-tenants/tenant-operator`

### Authentication errors
- Verify your GitHub token has `read:packages` permission
- Check token expiration
- Ensure the token belongs to a user with access to the repository

### Platform mismatch
- Docker will automatically select the correct platform
- To force a specific platform: `docker pull --platform linux/amd64 ghcr.io/kubernetes-tenants/tenant-operator:latest`

## CI/CD Integration

The workflow includes:
- ✅ Multi-platform builds (amd64, arm64)
- ✅ Automatic tagging based on git refs
- ✅ Build caching for faster builds
- ✅ SLSA provenance attestation
- ✅ PR validation (build without push)

## Permissions

The workflow requires these permissions (automatically granted):
- `contents: read` - Read repository contents
- `packages: write` - Push to GHCR
- `id-token: write` - Generate provenance attestations
