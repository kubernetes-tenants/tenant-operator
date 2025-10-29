# GitHub Actions Workflows

This directory contains automated CI/CD workflows for the Tenant Operator project.

## Workflows

### 🚀 Build and Push (`build-push.yml`)

Automatically builds and publishes container images to GitHub Container Registry (GHCR).

**Triggers:**
- Push to `main` branch
- Version tags (`v*`)
- Pull requests (build only, no push)

**Features:**
- ✅ Multi-platform builds (`linux/amd64`, `linux/arm64`)
- ✅ Automatic semantic versioning from git tags
- ✅ Build caching for faster builds
- ✅ SLSA provenance attestation for supply chain security
- ✅ PR validation without publishing

**Image Registry:** `ghcr.io/kubernetes-tenants/tenant-operator`

**Permissions Required:**
- `contents: read` - Read repository code
- `packages: write` - Push to GHCR
- `id-token: write` - Generate attestations

### 🧪 Test (`test.yml`)

Runs unit tests on every push and pull request.

**Tests:**
- Unit tests with coverage
- Multiple Go versions

### 🧹 Lint (`lint.yml`)

Runs code quality checks.

**Checks:**
- `go fmt` formatting
- `go vet` static analysis
- golangci-lint

### 🔬 E2E Tests (`test-e2e.yml`)

Runs end-to-end tests against a real Kubernetes cluster.

**Environment:**
- Kind cluster
- Full operator deployment
- Integration with MySQL

## Image Tagging Strategy

| Event | Generated Tags | Example |
|-------|---------------|---------|
| Push to `main` | `main`, `latest`, `main-<sha>` | `main`, `latest`, `main-abc123` |
| Tag `v1.2.3` | `v1.2.3`, `v1.2`, `v1`, `latest` | All semver variants |
| PR #42 | `pr-42` | Build only, not pushed |
| Commit SHA | `<branch>-<sha>` | `main-abc123` |

## Using Pre-built Images

### Pull the latest image
```bash
docker pull ghcr.io/kubernetes-tenants/tenant-operator:latest
```

### Deploy to Kubernetes
```bash
make deploy IMG=ghcr.io/kubernetes-tenants/tenant-operator:v1.0.0
```

### Verify image provenance
```bash
cosign verify-attestation \
  --type slsaprovenance \
  ghcr.io/kubernetes-tenants/tenant-operator:v1.0.0
```

## Workflow Configuration

### Secrets Required

No additional secrets needed! The workflows use the built-in `GITHUB_TOKEN` which is automatically provided by GitHub Actions.

### Repository Settings

To enable GHCR publishing:

1. **Go to:** Repository Settings → Actions → General
2. **Workflow permissions:** Set to "Read and write permissions"
3. **Allow GitHub Actions to create and approve pull requests:** ✅ (optional)

### Package Visibility

After the first push, set package visibility:

1. **Go to:** Repository → Packages → tenant-operator
2. **Package settings → Change visibility**
3. **Choose:** Public (recommended) or Private

## Development

### Testing Workflows Locally

Use [act](https://github.com/nektos/act) to test workflows locally:

```bash
# Install act
brew install act

# Test the build workflow
act push -j build-and-push --secret GITHUB_TOKEN=your_token

# Test on pull_request event
act pull_request
```

### Modifying Workflows

When updating workflows:

1. Test locally with `act` if possible
2. Create a PR to trigger validation
3. Check workflow runs in GitHub Actions tab
4. Merge after all checks pass

### Adding New Workflows

1. Create `.github/workflows/new-workflow.yml`
2. Follow GitHub Actions syntax
3. Test with `act` or in a PR
4. Document in this README

## Troubleshooting

### Build fails with "permission denied"

**Solution:** Check repository workflow permissions:
- Settings → Actions → General → Workflow permissions
- Select "Read and write permissions"

### Image not found after push

**Solution:** Check package visibility:
- Go to repository packages
- Ensure tenant-operator package is set to Public
- Verify the image exists at: https://github.com/orgs/kubernetes-tenants/packages

### Multi-platform build timeout

**Solution:**
- Multi-arch builds take longer (especially arm64)
- Increase timeout in workflow if needed
- Use build caching (already enabled)

### Provenance attestation fails

**Solution:**
- Requires `id-token: write` permission (already set)
- Only runs on actual pushes (not PRs)
- Check GitHub Actions logs for details

## Monitoring

### View Workflow Status

**GitHub UI:**
- Repository → Actions tab
- See all workflow runs and their status

**Badge in README:**
```markdown
[![Build Status](https://github.com/kubernetes-tenants/tenant-operator/actions/workflows/build-push.yml/badge.svg)](https://github.com/kubernetes-tenants/tenant-operator/actions/workflows/build-push.yml)
```

### View Published Images

**GHCR Package Page:**
https://github.com/kubernetes-tenants/tenant-operator/pkgs/container/tenant-operator

**Docker Hub Alternative:**
```bash
# List all tags
curl -s https://ghcr.io/v2/kubernetes-tenants/tenant-operator/tags/list
```

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [SLSA Provenance](https://slsa.dev/provenance/)

## Related Documentation

- [DEPLOYMENT.md](DEPLOYMENT.md) - Detailed deployment guide
- [../../README.md](../../README.md) - Main project README
