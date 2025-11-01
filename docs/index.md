---
layout: home

hero:
  name: "Tenant Operator"
  text: "Multi-Tenant Kubernetes Automation"
  tagline: Declarative, template-based resource provisioning with Server-Side Apply
  image:
    src: /logo.png
    alt: Tenant Operator
  actions:
    - theme: brand
      text: Get Started
      link: /quickstart
    - theme: alt
      text: View on GitHub
      link: https://github.com/kubernetes-tenants/tenant-operator

features:
  - icon: 🚀
    title: Multi-Tenant Auto-Provisioning
    details: Read tenant data from external datasources (MySQL) and automatically create/sync Kubernetes resources using templates

  - icon: ⚙️
    title: Policy-Based Lifecycle
    details: CreationPolicy (Once/WhenNeeded), DeletionPolicy (Delete/Retain), ConflictPolicy (Stuck/Force), PatchStrategy (apply/merge/replace)

  - icon: 🔄
    title: Server-Side Apply (SSA)
    details: Kubernetes-native declarative resource management with conflict-free updates using SSA field manager

  - icon: 📊
    title: Dependency Management
    details: DAG-based resource ordering with automatic cycle detection and topological sorting

  - icon: 📝
    title: Powerful Template System
    details: Go text/template with 200+ Sprig functions and custom helpers (toHost, trunc63, sha1sum, fromJson)

  - icon: 🔍
    title: Comprehensive Observability
    details: 12 Prometheus metrics, structured logging, Kubernetes events, and Grafana dashboards

  - icon: ✅
    title: Strong Consistency
    details: Ensures desired count = (referencing templates × active rows) with automatic garbage collection

  - icon: 🛡️
    title: Production Ready
    details: Webhooks, finalizers, drift detection, auto-correction, RBAC, and comprehensive validation

  - icon: 🔌
    title: Extensible Integration
    details: External datasource support (MySQL), External-DNS, Terraform Operator, and custom resources
---

## Quick Example

### 1. Define a Registry (External Datasource)

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantRegistry
metadata:
  name: my-saas-registry
spec:
  source:
    type: mysql
    mysql:
      host: mysql.default.svc.cluster.local
      port: 3306
      database: tenants
      table: tenant_data
      username: tenant_reader
      passwordRef:
        name: mysql-secret
        key: password
    syncInterval: 30s
  valueMappings:
    uid: tenant_id
    hostOrUrl: domain
    activate: is_active
```

### 2. Create a Template

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: web-app
spec:
  registryId: my-saas-registry
  deployments:
    - id: app-deployment
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
                  image: "nginx:latest"
                  env:
                    - name: TENANT_ID
                      value: "{{ .uid }}"
                    - name: DOMAIN
                      value: "{{ .host }}"
```

### 3. Automatic Tenant Provisioning

The operator automatically creates Tenant CRs for each active row:

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: Tenant
metadata:
  name: acme-web-app
spec:
  uid: acme
  templateRef: web-app
  registryId: my-saas-registry
  # ... auto-populated resources
status:
  desiredResources: 10
  readyResources: 10
  failedResources: 0
  conditions:
    - type: Ready
      status: "True"
```

## Architecture

### System Overview

```mermaid
flowchart TB
    subgraph External["External Data Source"]
        DB[(MySQL / PostgreSQL)]
    end

    subgraph Cluster["Kubernetes Cluster"]
        direction TB

        subgraph Controllers["Operator Controllers"]
            RC[TenantRegistry Controller]
            TC[TenantTemplate Controller]
            TNC[Tenant Controller]
        end

        subgraph CRDs["Custom Resources"]
            TR[TenantRegistry]
            TT[TenantTemplate]
            T[Tenant CRs]
        end

        subgraph Engine["Apply Engine"]
            SSA["SSA Apply Engine<br/>(fieldManager: tenant-operator)"]
        end

        subgraph Resources["Kubernetes Resources"]
            DEP[Deployments]
            SVC[Services]
            ING[Ingresses]
            etc[ConfigMaps, Secrets, ...]
        end

        API[(etcd / K8s API Server)]
    end

    DB -->|"syncInterval<br/>(e.g., 1m)"| RC
    RC -->|"Creates/Updates/Deletes<br/>Tenant CRs"| API
    API -->|"Stores"| TR
    API -->|"Stores"| TT
    API -->|"Stores"| T

    TC -->|"Validates<br/>template-registry linkage"| API
    TNC -->|"Reconciles<br/>each Tenant"| SSA
    SSA -->|"Server-Side Apply"| API

    API -->|"Creates"| DEP
    API -->|"Creates"| SVC
    API -->|"Creates"| ING
    API -->|"Creates"| etc

    style RC fill:#e3f2fd,stroke:#64b5f6,stroke-width:2px
    style TC fill:#e8f5e9,stroke:#81c784,stroke-width:2px
    style TNC fill:#fff3e0,stroke:#ffb74d,stroke-width:2px
    style SSA fill:#fce4ec,stroke:#f06292,stroke-width:2px
    style DB fill:#f3e5f5,stroke:#ba68c8,stroke-width:2px
```

### Reconciliation Flow

```mermaid
sequenceDiagram
    participant DB as MySQL Database
    participant RC as Registry Controller
    participant API as K8s API Server
    participant TC as Template Controller
    participant TNC as Tenant Controller
    participant SSA as SSA Engine

    Note over DB,SSA: Registry Sync Cycle (e.g., every 1 minute)

    RC->>DB: SELECT * FROM tenants WHERE activate=TRUE
    DB-->>RC: Active tenant rows

    RC->>API: Create/Update Tenant CRs (desired set)
    RC->>API: Delete Tenants not in desired set

    TC->>API: Validate Template-Registry linkage
    TC->>API: Ensure consistency

    loop For Each Tenant
        TNC->>API: Get Tenant spec
        TNC->>TNC: Build dependency graph (dependIds)
        TNC->>TNC: Topological sort resources

        loop For Each Resource (in order)
            TNC->>TNC: Render templates (name, namespace, spec)
            TNC->>SSA: Apply resource with conflict policy
            SSA->>API: Server-Side Apply (force or not)

            alt waitForReady = true
                TNC->>API: Wait for resource Ready condition
                API-->>TNC: Ready (or timeout)
            end
        end

        TNC->>API: Update Tenant status (ready/failed counts)
    end

    RC->>API: Update Registry status (desired/ready/failed)
```

## Key Features

### 🎯 Three-Controller Design

1. **TenantRegistry Controller**: Syncs database (e.g., 1m interval) → Creates/Updates/Deletes Tenant CRs
2. **TenantTemplate Controller**: Validates template-registry linkage and invariants
3. **Tenant Controller**: Renders templates → Resolves dependencies → Applies resources via SSA

### 📦 CRD Architecture

- **TenantRegistry**: Defines external datasource, sync interval, value mappings
- **TenantTemplate**: Blueprint for resources (Deployments, Services, Ingresses, etc.)
- **Tenant**: Instance representing a single tenant with status tracking

### 🔧 Advanced Capabilities

- **Multi-template support**: One registry → multiple templates
- **Garbage collection**: Auto-delete when rows removed or activate=false
- **Drift detection**: Event-driven watches with auto-correction
- **Smart requeue**: 30-second intervals for fast status reflection
- **Resource readiness**: 11+ resource types with custom checks
- **Finalizers**: Safe cleanup respecting deletion policies

## Documentation

<div class="vp-doc">
  <div class="custom-block tip">
    <p class="custom-block-title">Getting Started</p>
    <ul>
      <li><a href="/installation">Installation Guide</a> - Deploy to your cluster</li>
      <li><a href="/quickstart">Quick Start</a> - Get up and running in 5 minutes</li>
      <li><a href="/local-development-minikube">Local Development</a> - Minikube setup</li>
    </ul>
  </div>

  <div class="custom-block info">
    <p class="custom-block-title">Core Concepts</p>
    <ul>
      <li><a href="/api">API Reference</a> - Complete CRD documentation</li>
      <li><a href="/datasource">Datasources</a> - External data integration</li>
      <li><a href="/templates">Templates</a> - Go template system</li>
      <li><a href="/policies">Policies</a> - Lifecycle management</li>
    </ul>
  </div>

  <div class="custom-block warning">
    <p class="custom-block-title">Operations</p>
    <ul>
      <li><a href="/monitoring">Monitoring</a> - Prometheus metrics & alerts</li>
      <li><a href="/security">Security</a> - RBAC & best practices</li>
      <li><a href="/troubleshooting">Troubleshooting</a> - Common issues</li>
    </ul>
  </div>

  <div class="custom-block note">
    <p class="custom-block-title">Integrations</p>
    <ul>
      <li><a href="/integration-external-dns">ExternalDNS</a> - Automate DNS lifecycle per tenant</li>
      <li><a href="/integration-terraform-operator">Terraform Operator</a> - Provision cloud services via GitOps</li>
      <li><a href="/integration-argocd">Argo CD</a> - 1:1 Tenant ↔ Application GitOps delivery</li>
    </ul>
  </div>
</div>

## Community

- **GitHub**: [kubernetes-tenants/tenant-operator](https://github.com/kubernetes-tenants/tenant-operator)
- **Issues**: [Report bugs or request features](https://github.com/kubernetes-tenants/tenant-operator/issues)
