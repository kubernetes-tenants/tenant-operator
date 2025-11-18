<div align="center">

<img src="docs/public/logo.png" alt="Kubernetes Lynq Operator" width="400">

# Lynq Operator

### Kubernetes-Native Database-Driven Automation

**Automate node lifecycle from database to production**

[![Go Report Card](https://goreportcard.com/badge/github.com/k8s-lynq/lynq)](https://goreportcard.com/report/github.com/k8s-lynq/lynq)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/lynq)](https://artifacthub.io/packages/search?repo=lynq)
[![Build Status](https://github.com/k8s-lynq/lynq/actions/workflows/build-push.yml/badge.svg)](https://github.com/k8s-lynq/lynq/actions/workflows/build-push.yml)
[![Container Image](https://img.shields.io/badge/container-ghcr.io-blue)](https://github.com/k8s-lynq/lynq/pkgs/container/lynq)
[![Go Version](https://img.shields.io/github/go-mod/go-version/k8s-lynq/lynq)](go.mod)

[Documentation](https://lynq.sh/) • [Quick Start](https://lynq.sh/quickstart) • [Architecture](https://lynq.sh/architecture) • [Contributing](#-contributing)

</div>

---

## 📖 Overview

**Lynq Operator** is a Kubernetes operator that automates database-driven infrastructure provisioning. It reads node data from external datasources and dynamically creates, updates, and manages Kubernetes resources using declarative templates.

**One database row = One fully provisioned node stack**

🗄️ Database-driven automation • 📝 Go templates + Sprig functions • 🔄 Server-Side Apply • 📊 DAG-based dependencies • ⚙️ Lifecycle policies • 🚀 Production-ready

📚 **[Complete Documentation](https://lynq.sh/)** • 🏗️ **[Architecture Guide](https://lynq.sh/architecture)**

---

## 🚀 Quick Start

Get started with Lynq Operator in 5 minutes:

🎯 **[Quick Start Guide](https://lynq.sh/quickstart)** - Step-by-step tutorial with working examples
📦 **[Installation Guide](https://lynq.sh/installation)** - Helm, Kustomize, and manual installation options
⚙️ **[Configuration Examples](https://lynq.sh/quickstart#configuration)** - LynqHub and LynqForm setup

---

## 📚 Documentation

Complete documentation is available at **[lynq.sh](https://lynq.sh/)**

**Getting Started**: [Quick Start](https://lynq.sh/quickstart) • [Installation](https://lynq.sh/installation)
**Core Concepts**: [Architecture](https://lynq.sh/architecture) • [API Reference](https://lynq.sh/api) • [Templates](https://lynq.sh/templates)
**Operations**: [DataSource Setup](https://lynq.sh/datasource) • [Monitoring](https://lynq.sh/monitoring) • [Troubleshooting](https://lynq.sh/troubleshooting)
**Integrations**: [ExternalDNS](https://lynq.sh/integration-external-dns) • [Flux](https://lynq.sh/integration-flux) • [Argo CD](https://lynq.sh/integration-argocd)

---

## 🛠️ Development

```bash
git clone https://github.com/k8s-lynq/lynq.git
cd lynq
make install  # Install CRDs
make run      # Run locally
make test     # Run tests
```

📖 **[Development Guide](https://lynq.sh/development)** - Building, testing, and contributing to Lynq Operator

---

## 🤝 Contributing

We welcome contributions! Whether you're fixing bugs, adding features, improving documentation, or sharing use cases - all contributions are valued.

🌟 **Want to add a new datasource?** Lynq uses a pluggable adapter pattern - see our [Contributing a Datasource Guide](https://lynq.sh/contributing-datasource)

📋 **[Contributing Guidelines](CONTRIBUTING.md)** - Code standards, commit conventions, and PR process

---

## 📊 Project Status

**Production-ready** • Kubernetes v1.28-v1.33 validated • Apache 2.0 License

- 🗺️ **[Roadmap](docs/roadmap.md)** - Feature plans and versioning

---

## 📝 License

Licensed under the [Apache License 2.0](LICENSE).
Copyright 2025 Lynq Operator Authors.

---

## 📬 Community

🐛 [GitHub Issues](https://github.com/k8s-lynq/lynq/issues) • 💬 [Discussions](https://github.com/k8s-lynq/lynq/discussions)

---

## Star History

<a href="https://www.star-history.com/#k8s-lynq/lynq&type=timeline&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=k8s-lynq/lynq&type=timeline&theme=dark&legend=top-left" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=k8s-lynq/lynq&type=timeline&legend=top-left" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=k8s-lynq/lynq&type=timeline&legend=top-left" />
 </picture>
</a>

---

<div align="center">

**[⬆ Back to Top](#lynq)**

</div>
