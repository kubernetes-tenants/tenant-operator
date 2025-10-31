# Helm Chart Packaging

This document describes the structure, configuration surface, and workflow for packaging the Tenant Operator with Helm.

## Layout

- `charts/tenant-operator/Chart.yaml`: chart metadata (name, version, sources).
- `charts/tenant-operator/values.yaml`: default configuration used by the templates.
- `charts/tenant-operator/templates/`: Kubernetes objects rendered by Helm (manager `Deployment`, RBAC, metrics service, webhook resources, and optional cert-manager assets).
- `charts/tenant-operator/crds/`: CustomResourceDefinition manifests shipped with the chart. Helm installs these before the rest of the resources.

When CRDs change (e.g., after `make manifests`), copy the regenerated files from `config/crd/bases` into the `crds/` directory.

## Configuration Highlights

Key values surfaced in `values.yaml`:

- `image.repository`, `image.tag`, `image.pullPolicy`: container image settings for the controller manager.
- `replicaCount`: number of controller manager replicas.
- `serviceAccount.*`: toggle between creating a managed ServiceAccount or reusing an existing one.
- `rbac.create`: master switch for controller RBAC (ClusterRoles/Bindings and leader election Role/Binding).
- `rbac.additionalClusterRoles`: installs helper ClusterRoles (`tenant-*/tenantregistry-*/tenanttemplate-*`) for delegating access to application teams.
- `metrics.enabled`: adds the HTTPS metrics endpoint argument, RBAC, and the metrics `Service`. `metrics.service.port` controls the port exposed by both the Service and the container.
- `webhook.enabled`: controls webhook Service, configurations, and the webhook port on the controller manager.
- `webhook.certManager.enabled`: when true, the chart provisions an optional self-signed Issuer (if `issuer.create` is true) and a `Certificate` resource that targets the webhook Service. The webhook configurations receive `cert-manager.io/inject-ca-from` annotations.
- `webhook.existingSecretName`: reuse a pre-provisioned TLS secret instead of creating one with cert-manager. When set, populate `webhook.caBundle` with the PEM-encoded CA (the chart base64 encodes it).
- `controllerManager.*`: pod-level knobs (pod annotations, node scheduling, security contexts, resources, additional args/env/volumes).

## Build and Test Workflow

From the project root (requires Helm 3):

```sh
# Validate that the chart renders cleanly
helm lint charts/tenant-operator

# Render the manifests with default values (no cluster access required)
helm template tenant charts/tenant-operator --namespace tenant-operator-system > /tmp/tenant-operator.yaml
```

To build a distributable archive locally, prefer the automation described below.

## Automation

### Local packaging script

Run `scripts/package-helm-chart.sh` to sync CRDs, lint the chart, and emit `dist/tenant-operator-<version>.tgz`:

```sh
./scripts/package-helm-chart.sh
```

The script expects Helm 3 to be available in `PATH`. It overwrites existing chart packages in `dist/` so the directory always reflects the latest run.
Set `CHART_VERSION` (and optionally `APP_VERSION`) to override the version embedded in `Chart.yaml` without editing the file manually:

```sh
CHART_VERSION=0.2.0 ./scripts/package-helm-chart.sh
```

Set `PACKAGE_CHART=false` to skip the `helm package` step when another tool (e.g., the GitHub Actions workflow) will perform packaging:

```sh
PACKAGE_CHART=false CHART_VERSION=0.2.0 ./scripts/package-helm-chart.sh
```

### GitHub Actions release

The workflow `.github/workflows/release-helm-chart.yml` packages and publishes the chart whenever a tag matching `v*` or `chart-v*` is pushed. It derives the version from the tag (e.g., `v0.2.0`, `chart-v0.2.0` → `0.2.0`) and then:

1. Checks out the repository (including the `gh-pages` branch) and installs Helm.
2. Executes `scripts/package-helm-chart.sh` with `CHART_VERSION` set to the derived version (packaging disabled) to sync CRDs and lint the chart.
3. Runs `helm/chart-releaser-action`, which packages the chart, creates (or updates) a GitHub Release, and regenerates the `index.yaml` hosted on the `gh-pages` branch so the chart can be consumed via `helm repo add`.
4. Uploads the generated package from `.cr-release-packages/` as an Actions artifact for quick inspection.

To publish a new chart:

1. Ensure `make manifests` has been run if CRDs changed; the packaging script copies the results automatically.
2. Optionally run `CHART_VERSION=<target> ./scripts/package-helm-chart.sh` (with or without packaging) to verify the release locally.
3. Tag the commit (e.g., `git tag v0.1.0 && git push origin v0.1.0` or `git tag chart-v0.1.0`).
4. The workflow will package the chart with the derived version, update the GitHub Pages chart repository, and publish a GitHub release once the tag reaches the repository.

> **Prerequisite:** Ensure a `gh-pages` branch exists (can be empty) and GitHub Pages is configured to serve that branch. After the first release, users can install via `helm repo add tenant-operator https://<github-username>.github.io/tenant-operator` (replace with your organization/repo path).

## Release Strategy

1. Update CRDs (if needed) via `make manifests`; `scripts/package-helm-chart.sh` copies them into the chart.
2. Adjust values defaults or documentation in the chart as needed. The release automation sets `version`/`appVersion` based on the tag, but update them in git if you prefer the repository to reflect the last published release.
3. Run `helm lint` and `helm template` (or the packaging script) against representative configurations.
4. Package the chart manually if distributing outside GitHub Releases, or push a `v*`/`chart-v*` tag to trigger the workflow.
5. Update any external documentation or Helm repository index that references the new release.

## Usage Notes

- Install into the desired namespace (for example `tenant-operator-system`) using `helm install tenant charts/tenant-operator --namespace tenant-operator-system --create-namespace`.
- When disabling cert-manager integration, supply a valid `webhook.caBundle` and ensure `webhook.existingSecretName` points to a TLS secret containing `tls.crt` and `tls.key`.
- The metric and webhook Services rely on the `control-plane=controller-manager` label, which is applied consistently across the Deployment and Services.
- Additional helper ClusterRoles are optional; disable them via `--set rbac.additionalClusterRoles=false` if your platform ships opinionated role templates.
