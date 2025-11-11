<div align="center">

<img src="docs/public/logo.png" alt="Kubernetes Tenant Operator" width="400">

# Tenant Operator

### Kubernetes-Native Multi-Tenant Application Provisioning

**Automate tenant lifecycle from database to production**

[![Go Report Card](https://goreportcard.com/badge/github.com/kubernetes-tenants/tenant-operator)](https://goreportcard.com/report/github.com/kubernetes-tenants/tenant-operator)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/tenant-operator)](https://artifacthub.io/packages/search?repo=tenant-operator)
[![Build Status](https://github.com/kubernetes-tenants/tenant-operator/actions/workflows/build-push.yml/badge.svg)](https://github.com/kubernetes-tenants/tenant-operator/actions/workflows/build-push.yml)
[![Container Image](https://img.shields.io/badge/container-ghcr.io-blue)](https://github.com/kubernetes-tenants/tenant-operator/pkgs/container/tenant-operator)
[![Go Version](https://img.shields.io/github/go-mod/go-version/kubernetes-tenants/tenant-operator)](go.mod)

[Overview](#-overview) • [Features](#-key-features) • [Quick Start](#-quick-start) • [Documentation](#-documentation) • [Contributing](#-contributing)

</div>

---

## 📖 Overview

**Tenant Operator** is a Kubernetes operator that automates multi-tenant application provisioning from database records. It reads tenant data from external datasources (MySQL, PostgreSQL planned for v1.2) and dynamically creates, updates, and manages Kubernetes resources using declarative templates.

**One database row = One fully provisioned tenant stack**

### Why Tenant Operator?

**Traditional multi-tenant approaches are limited:**
- ❌ Helm: Manual per-tenant releases and values files
- ❌ GitOps: Static manifests don't scale to thousands of tenants
- ❌ Custom scripts: Fragile, hard to maintain, no drift correction

**Tenant Operator provides:**
- ✅ **Database-driven automation**: Your existing database becomes the source of truth
- ✅ **Real-time sync**: Changes propagate automatically (30s status reflection)
- ✅ **Production-grade**: SSA, webhooks, finalizers, metrics, and drift detection built-in

📚 **[Complete Documentation](https://docs.kubernetes-tenants.org/)** • 🏗️ **[Architecture Details](https://docs.kubernetes-tenants.org/architecture)**

---

## ✨ Key Features

| Feature | Description |
|---------|-------------|
| **🗄️ Database-Driven** | MySQL support (PostgreSQL in v1.2) |
| **📝 Template Engine** | Go templates + 200+ Sprig functions |
| **🔄 Server-Side Apply** | Conflict-free resource management |
| **📊 Dependencies** | DAG-based resource ordering |
| **⚙️ Policies** | Fine-grained lifecycle control |
| **🚀 Production-Ready** | Webhooks, metrics, drift detection |

**Advanced capabilities:** Multi-template support, cross-namespace provisioning, orphan cleanup, smart watch predicates, custom template functions (`sha1sum`, `fromJson`, `toHost`), and more.

📖 **Full feature list:** [Documentation](https://docs.kubernetes-tenants.org/)

### Integrations

- [**ExternalDNS**](https://docs.kubernetes-tenants.org/integration-external-dns) - Automatic DNS records (Route53, Cloudflare, etc.)
- [**Terraform Operator**](https://docs.kubernetes-tenants.org/integration-terraform-operator) - Cloud resource provisioning (S3, RDS, CDN)
- **cert-manager** - Automatic TLS certificates
- **Prometheus/Grafana** - Complete monitoring with dashboards

---

## 🏗️ Architecture

Tenant Operator uses a three-controller design for database-to-Kubernetes synchronization:

1. **TenantRegistry Controller** - Syncs database → Creates Tenant CRs
2. **TenantTemplate Controller** - Validates templates and linkage
3. **Tenant Controller** - Renders templates → Applies resources via SSA

**Multi-template support:** One registry can be referenced by multiple templates (prod, staging, etc.). Desired count = `referencingTemplates × activeRows`.

📊 **Detailed architecture diagrams:** [Architecture Guide](https://docs.kubernetes-tenants.org/architecture)

## Supported Kubernetes Versions

| Version | Status |
|---------|--------|
| v1.28 - v1.33 | ✅ Validated |
| Other GA releases | ⚠️ Expected |

The operator targets GA/stable Kubernetes APIs and is decoupled from specific cluster releases. Validate in staging before production rollout.

---

## 🚀 Quick Start

Get started in 5 minutes with a working example:

🎯 **[Quick Start Guide with Minikube](https://docs.kubernetes-tenants.org/quickstart)** - Automated setup scripts included

### Installation (Helm - Recommended)

```bash
# 1. Install cert-manager (required for webhooks)
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# 2. Install Tenant Operator
helm repo add tenant-operator https://kubernetes-tenants.github.io/tenant-operator
helm repo update

helm install tenant-operator tenant-operator/tenant-operator \
  --namespace tenant-operator-system \
  --create-namespace
```

📖 **More installation options:** [Installation Guide](https://docs.kubernetes-tenants.org/installation)

### Configuration Example

**1. Connect to your database (TenantRegistry):**

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantRegistry
metadata:
  name: my-saas-registry
spec:
  source:
    type: mysql
    mysql:
      host: mysql.database.svc.cluster.local
      database: tenants
      table: tenant_configs
      passwordRef:
        name: mysql-secret
        key: password
    syncInterval: 1m
  valueMappings:
    uid: tenant_id
    hostOrUrl: tenant_url
    activate: is_active  # Must be "1", "true", or "yes"
```

📖 **Database setup guide:** [DataSource Configuration](https://docs.kubernetes-tenants.org/datasource)

**2. Define resource template (TenantTemplate):**

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: saas-app-template
spec:
  registryId: my-saas-registry
  deployments:
    - id: app
      nameTemplate: "{{ .uid }}-app"
      spec:
        apiVersion: apps/v1
        kind: Deployment
        spec:
          replicas: 2
          template:
            spec:
              containers:
              - name: app
                image: myapp:latest
                env:
                - name: TENANT_ID
                  value: "{{ .uid }}"
```

📖 **Template syntax and functions:** [Template Guide](https://docs.kubernetes-tenants.org/templates)

**3. Verify:**

```bash
kubectl get tenants --watch
kubectl get tenant <tenant-name> -o yaml
```

**Result:** Each active database row automatically provisions a complete tenant stack!

---

## 📚 Documentation

Complete documentation is available at **[docs.kubernetes-tenants.org](https://docs.kubernetes-tenants.org/)**

### Quick Links

| Category | Pages |
|----------|-------|
| **Getting Started** | [Quick Start](https://docs.kubernetes-tenants.org/quickstart) • [Installation](https://docs.kubernetes-tenants.org/installation) |
| **Core Concepts** | [Architecture](https://docs.kubernetes-tenants.org/architecture) • [API Reference](https://docs.kubernetes-tenants.org/api) • [Templates](https://docs.kubernetes-tenants.org/templates) • [Policies](https://docs.kubernetes-tenants.org/policies) • [Dependencies](https://docs.kubernetes-tenants.org/dependencies) |
| **Operations** | [DataSource Setup](https://docs.kubernetes-tenants.org/datasource) • [Monitoring](https://docs.kubernetes-tenants.org/monitoring) • [Alert Runbooks](https://docs.kubernetes-tenants.org/alert-runbooks) • [Troubleshooting](https://docs.kubernetes-tenants.org/troubleshooting) • [Performance](https://docs.kubernetes-tenants.org/performance) |
| **Integrations** | [ExternalDNS](https://docs.kubernetes-tenants.org/integration-external-dns) • [Terraform Operator](https://docs.kubernetes-tenants.org/integration-terraform-operator) • [Argo CD](https://docs.kubernetes-tenants.org/integration-argocd) |
| **Advanced** | [Security](https://docs.kubernetes-tenants.org/security) • [Development](https://docs.kubernetes-tenants.org/development) • [Contributing](https://docs.kubernetes-tenants.org/contributing-datasource) |

### Examples

**Simple SaaS Application:**
```sql
INSERT INTO tenants VALUES ('acme-corp', 'https://acme.myapp.io', 1, 'enterprise');
```
→ Automatically creates Deployment, Service, Ingress, ConfigMaps, and Secrets

**Multi-region with custom variables:**
```yaml
extraValueMappings:
  region: deployment_region
  dbHost: database_host
```

**Template functions:**
```yaml
nameTemplate: "{{ .uid | sha1sum | trunc63 }}"  # Unique names
value: "{{ (.config | fromJson).apiKey }}"      # Parse JSON
value: "{{ .tenantUrl | toHost }}"              # Extract host
```

📖 **More examples:** [Quick Start Guide](https://docs.kubernetes-tenants.org/quickstart)

---

## 🛠️ Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/kubernetes-tenants/tenant-operator.git
cd tenant-operator

# Install dependencies
go mod download

# Run tests
make test

# Build binary
make build

# Build and push container
make docker-build docker-push IMG=<your-registry>/tenant-operator:tag
```

### Running Locally

```bash
# Install CRDs
make install

# Run controller locally (uses ~/.kube/config)
make run

# Run with debug logging
LOG_LEVEL=debug make run
```

### Running Tests

```bash
# Unit tests
make test

# Integration tests
make test-integration

# E2E tests (requires kind)
make test-e2e

# Coverage report
make test-coverage
```

---

## 🤝 Contributing

We welcome contributions from anyone interested in multi-tenant Kubernetes automation.

**Our vision:** We're building this project with open governance and community-first principles. As the project grows, we aspire to join cloud-native foundations like CNCF to foster broader collaboration and adoption.

**Ways to contribute:**
- 🐛 Bug reports and feature requests
- 📝 Documentation improvements
- 💻 Code contributions (features, bug fixes, optimizations)
- 🌍 Translations and internationalization
- 🎨 UX/UI improvements for tooling
- 📊 Use case sharing and feedback

### 🌟 Want to Add a New Datasource?

Tenant Operator uses a **pluggable adapter pattern** that makes it easy to add support for new datasources (PostgreSQL, MongoDB, REST APIs, etc.).

**Why contribute a datasource?**
- ✅ Only 2 methods to implement
- ✅ MySQL reference implementation to follow
- ✅ Complete step-by-step guide provided
- ✅ Recognition in release notes

📚 **Full Guide**: [Contributing a New Datasource](docs/contributing-datasource.md)

### How to Contribute

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'feat: add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Contribution Guidelines

- Follow [Conventional Commits](https://www.conventionalcommits.org/)
- Add tests for new features
- Update documentation
- Run `make lint` before submitting
- Ensure all CI checks pass

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

## 🗺️ Roadmap

See [full roadmap](docs/roadmap.md) for details.

---

## 📊 Status

| Component | Status | Description |
|-----------|--------|-------------|
| **Core Controllers** | ✅ Production | Registry, Template, and Tenant controllers with SSA |
| **MySQL Datasource** | ✅ Production | Sync from MySQL with column mapping and VIEWs |
| **Template Engine** | ✅ Production | Go templates with 200+ Sprig functions + custom functions |
| **Webhooks** | ✅ Production | Validation and defaulting for all CRDs |
| **Performance** | ✅ Production | Event-driven reconciliation with smart watch predicates |
| **Multi-Template** | ✅ Production | One registry, multiple templates (prod, staging, etc.) |
| **Cross-Namespace** | ✅ Production | Provision resources across namespaces with label tracking |
| **Orphan Cleanup** | ✅ Production | Automatic detection and cleanup of removed resources |
| **PostgreSQL** | 🚧 Planned | PostgreSQL datasource adapter (v1.2) |

**Kubernetes:** v1.28-v1.33 validated • **Status:** Production-ready

---

## ❓ FAQ

<details>
<summary><b>How is this different from Helm or GitOps?</b></summary>

Tenant Operator is **database-driven** and designed for SaaS platforms where tenant data lives in databases, not git repositories. It automates provisioning from database rows with real-time sync, whereas Helm and GitOps require manual per-tenant configuration.

📖 [Architecture Guide](https://docs.kubernetes-tenants.org/architecture)
</details>

<details>
<summary><b>Can I use my existing database?</b></summary>

Yes! You just need read-only access and column mappings. If your schema doesn't match, create a MySQL VIEW to transform the data.

📖 [DataSource Configuration Guide](https://docs.kubernetes-tenants.org/datasource)
</details>

<details>
<summary><b>What values are valid for the `activate` column?</b></summary>

Must be **exactly** one of: `"1"`, `"true"`, `"TRUE"`, `"True"`, `"yes"`, `"YES"`, `"Yes"`

All other values (including `"active"`, `"0"`, empty, NULL) are considered inactive. Use a VIEW to transform your data if needed.

📖 [DataSource Guide - Activate Column](https://docs.kubernetes-tenants.org/datasource#activate-column-requirements)
</details>

<details>
<summary><b>⚠️ What happens if I delete TenantRegistry or TenantTemplate?</b></summary>

**Warning:** Causes cascade deletion of all Tenant CRs and their resources!

**Protection:** Set `deletionPolicy: Retain` on resources BEFORE deleting, or update in-place instead of delete/recreate.

📖 [Policies Guide - Protecting Tenants](https://docs.kubernetes-tenants.org/policies#protecting-tenants-from-cascade-deletion)
</details>

<details>
<summary><b>How does it scale to thousands of tenants?</b></summary>

Production deployments handle 1000+ tenants with concurrent reconciliation, SSA efficiency, resource caching, and optional sharding.

📖 [Performance Guide](https://docs.kubernetes-tenants.org/performance)
</details>

<details>
<summary><b>How fast does it react to changes?</b></summary>

- **Immediate**: Event-driven drift correction
- **30 seconds**: Periodic status reflection
- **Configurable**: Database sync interval (e.g., 1 minute)

📖 [Architecture - Reconciliation Flow](https://docs.kubernetes-tenants.org/architecture#reconciliation-flow)
</details>

<details>
<summary><b>Can one registry support multiple environments?</b></summary>

Yes! One registry can be referenced by multiple templates (prod, staging, dev). Each database row creates multiple Tenant CRs.

📖 [Configuration Guide - Multi-Template](https://docs.kubernetes-tenants.org/configuration#multi-template-support)
</details>

<details>
<summary><b>How do I set up webhook certificates?</b></summary>

Install **cert-manager** first - it automatically manages TLS certificates for webhook communication. See installation guide for details.

📖 [Installation Guide](https://docs.kubernetes-tenants.org/installation)
</details>

---

## 📝 License

Copyright 2025 Tenant Operator Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

---

## 🌟 Acknowledgments

Built with:
- [Kubebuilder](https://kubebuilder.io/) - Kubernetes operator framework
- [Operator SDK](https://sdk.operatorframework.io/) - Operator development toolkit
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) - Kubernetes controller library
- [Sprig](https://masterminds.github.io/sprig/) - Template function library

Inspired by the cloud-native ecosystem and CNCF projects.

---

## 📬 Contact & Community

- 🐛 **Issues**: [GitHub Issues](https://github.com/kubernetes-tenants/tenant-operator/issues) - Report bugs or request features
- 💬 **Discussions**: [GitHub Discussions](https://github.com/kubernetes-tenants/tenant-operator/discussions) - Ask questions or share ideas
- 📧 **Email**: rationlunas@gmail.com - Direct contact for partnership inquiries

**We're looking for:**
- Users sharing their experiences and use cases
- Contributors interested in multi-tenant Kubernetes
- Organizations interested in collaboration

---

## Star History

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=kubernetes-tenants/tenant-operator&type=date&theme=dark&legend=top-left" />
  <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=kubernetes-tenants/tenant-operator&type=date&legend=top-left" />
  <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=kubernetes-tenants/tenant-operator&type=date&legend=top-left" />
</picture>

---

<div align="center">

**[⬆ Back to Top](#tenant-operator)**

</div>
