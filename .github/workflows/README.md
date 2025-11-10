# GitHub Actions Workflows

This directory contains automated workflows for the Tenant Operator project.

## Workflows

### 1. Deploy VitePress Docs (`deploy-docs-v1.yml`)

**Trigger**: Push to `main` branch or manual dispatch

**Purpose**: Builds and deploys VitePress documentation to GitHub Pages

**Key Features**:
- Builds VitePress site from `docs/` directory
- Deploys to gh-pages branch using `peaceiris/actions-gh-pages@v4`
- Uses `keep_files: true` to preserve Helm chart files (index.yaml, *.tgz, artifacthub-repo.yml)
- Shares `gh-pages-deploy` concurrency group to prevent conflicts

### 2. Release Helm Chart (`helm-release.yml`)

**Trigger**: Push of version tags (v*)

**Purpose**: Packages and releases Helm chart to GitHub Pages Helm repository

**Key Features**:
- Validates and packages Helm chart
- Uploads chart package to GitHub Release
- Updates Helm repository index (index.yaml) on gh-pages branch
- Deploys Artifact Hub metadata using `peaceiris/actions-gh-pages@v4` with `keep_files: true`
- Shares `gh-pages-deploy` concurrency group to prevent conflicts

## GitHub Pages Structure

The `gh-pages` branch contains:

```
gh-pages/
├── index.html              # VitePress docs (root)
├── assets/                 # VitePress assets
├── guide/                  # VitePress guide pages
├── api/                    # VitePress API docs
├── index.yaml              # Helm repository index (from helm-release.yml)
├── artifacthub-repo.yml    # Artifact Hub metadata (from helm-release.yml)
├── tenant-operator-*.tgz   # Helm chart packages (from helm-release.yml)
└── CNAME                   # Custom domain: docs.kubernetes-tenants.org
```

## Concurrency Control

Both workflows use the same concurrency group (`gh-pages-deploy`) with `cancel-in-progress: false` to:
- Prevent simultaneous deployments that could conflict
- Queue deployments sequentially
- Ensure all files are preserved using `keep_files: true`

## Best Practices

1. **keep_files: true**: Both workflows use this option to preserve files from other workflows
2. **Concurrency group**: Prevents race conditions when both workflows run simultaneously
3. **Atomic deployments**: Each workflow only updates its own files, others are preserved
4. **Idempotent**: Running workflows multiple times produces consistent results

## Troubleshooting

### Issue: index.yaml disappears after docs deployment
**Solution**: This was resolved by adding `keep_files: true` to the docs deployment workflow

### Issue: artifacthub-repo.yml not updating
**Solution**: Ensure helm-release.yml completes successfully and uses peaceiris/actions-gh-pages with keep_files

### Issue: Race condition between workflows
**Solution**: Both workflows share the `gh-pages-deploy` concurrency group to queue deployments
