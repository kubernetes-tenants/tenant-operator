# Quick Start with Minikube

Get Lynq running on Minikube in under 5 minutes using automated scripts.

::: info Multi-Node Example
As the most common use case for Lynq, this guide uses **Multi-Node** infrastructure (SaaS application with multiple customers/nodes) as an example. The pattern shown here can be adapted for any database-driven infrastructure automation scenario.
:::

[[toc]]

## Overview

This guide uses automated scripts to set up a complete local environment:
1. **Minikube cluster** with **cert-manager** (automatically installed)
2. **Lynq** deployed and running with webhooks enabled
3. **MySQL test database** for node data
4. **Sample LynqHub** and **LynqForm**
5. **Live node provisioning** from database

```mermaid
flowchart LR
    Cluster["Minikube Cluster"]
    Operator["Lynq"]
    Database["MySQL Test DB"]
    Templates["Sample Hub & Template"]
    Nodes["LynqNode CRs & Resources"]

    Cluster --> Operator --> Database --> Templates --> Nodes

    classDef stage fill:#e3f2fd,stroke:#64b5f6,stroke-width:2px;
    class Cluster,Operator,Database,Templates,Nodes stage;
```

::: tip Time required
Full setup typically completes in around 5 minutes.
:::

::: info cert-manager Included
cert-manager is **automatically installed** by the setup script. It's required for webhook validation and defaulting in all environments (including local development).
:::

## Prerequisites

### Required Tools

| Tool | Version | Install on macOS | Install on Linux |
| --- | --- | --- | --- |
| Minikube | v1.28.0+ | `brew install minikube` | [Installation Guide](https://minikube.sigs.k8s.io/docs/start/) |
| kubectl | v1.28.0+ | `brew install kubectl` | [Installation Guide](https://kubernetes.io/docs/tasks/tools/) |
| Docker (driver) | Latest | Docker Desktop | Docker Engine |

### System Requirements

- **CPU**: 1+ core
- **Memory**: 1+ GB RAM
- **Disk**: 5+ GB free space

## Step-by-Step Setup

### Step 1: Setup Minikube Cluster

Create a Minikube cluster with all prerequisites:

```bash
cd /path/to/lynq
./scripts/setup-minikube.sh
```

**What this does:**
- ✅ Creates Minikube cluster (2 CPUs, 2GB RAM)
- ✅ **Installs cert-manager v1.13.2** (required for webhooks)
- ✅ Installs Lynq CRDs
- ✅ Creates namespaces: `lynq-system`, `lynq-test`

::: warning cert-manager is Essential
cert-manager provides webhook TLS certificates for validation and defaulting. It's no longer optional, even for local development, to ensure consistency with production environments.
:::

**Time:** ~2 minutes

---

### Step 2: Deploy Lynq

Build and deploy the operator to Minikube:

```bash
./scripts/deploy-to-minikube.sh
```

**What this does:**
- ✅ Builds operator Docker image with timestamp tag
- ✅ Loads image into Minikube's internal registry
- ✅ Deploys operator to `lynq-system` namespace
- ✅ Waits for operator to be ready

**Time:** ~2 minutes

---

### Step 3: Deploy MySQL Test Database

Deploy a MySQL database with sample node data:

```bash
./scripts/deploy-mysql.sh
```

**What this does:**
- ✅ Deploys MySQL 8.0 to `lynq-test` namespace
- ✅ Creates `nodes` database and `node_configs` table
- ✅ Inserts 3 sample node rows
- ✅ Creates read-only user `node_reader`
- ✅ Creates Kubernetes Secret with credentials

**Sample data inserted:**
```sql
node_id         node_url                        is_active   subscription_plan
-----------     -----------------------------   ---------   -----------------
acme-corp       https://acme.example.com        1           enterprise
beta-inc        https://beta.example.com        1           startup
gamma-llc       https://gamma.example.com       0           trial
```

**Time:** ~1 minute

---

### Step 4: Deploy LynqHub

Create a LynqHub that connects to the MySQL database:

```bash
./scripts/deploy-lynqhub.sh
```

**What this does:**
- ✅ Creates LynqHub CR named `test-hub`
- ✅ Configures MySQL connection to test database
- ✅ Sets up column mappings (uid, hostOrUrl, activate)
- ✅ Starts syncing every 30 seconds

**Time:** ~30 seconds

---

### Step 5: Deploy LynqForm

Create a LynqForm that provisions resources for each node:

```bash
./scripts/deploy-lynqform.sh
```

**What this does:**
- ✅ Creates LynqForm CR named `test-template`
- ✅ Defines resource blueprints (Deployment, Service)
- ✅ Links to `test-hub`
- ✅ Triggers automatic node provisioning

**Time:** ~30 seconds

## 🎉 Success! You're Running Lynq

You now have:
- ✅ **Minikube cluster** with **cert-manager** (for webhook TLS)
- ✅ **Lynq** managing nodes with **webhooks enabled**
- ✅ **MySQL database** with 3 node rows
- ✅ **2 Active Nodes** (acme-corp, beta-inc) fully provisioned
- ✅ **Live sync** between database and Kubernetes
- ✅ **Admission validation** catching errors at apply time

### What Was Created?

For each active node (acme-corp, beta-inc):
```
LynqNode CR: acme-corp-test-template
├── Deployment: acme-corp-app
└── Service: acme-corp-app
```

**Verify your setup:**
```bash
# Check LynqNode CRs
kubectl get lynqnodes

# Check node resources
kubectl get deployments,services -l lynq.sh/node

# View operator logs
kubectl logs -n lynq-system -l control-plane=controller-manager -f
```

## Real-World Example

Let's see the complete lifecycle of a node from database to Kubernetes.

### Adding a Node

Insert a new row into the database:

```sql
INSERT INTO node_configs (node_id, node_url, is_active, subscription_plan)
VALUES ('acme-corp', 'https://acme.example.com', 1, 'enterprise');
```

**What happens automatically:**

Within 30 seconds (syncInterval), the operator creates:

```bash
# 1. LynqNode CR
kubectl get lynqnode acme-corp-test-template

# 2. Namespace (if configured)
kubectl get namespace acme-corp-namespace

# 3. Deployment
kubectl get deployment acme-corp-app

# 4. Service
kubectl get service acme-corp-app

# 5. Ingress (if configured)
kubectl get ingress acme-corp-ingress
```

All without writing any YAML files. The template defines the blueprint, the database row provides the variables.

### Deactivating a Node

Update the database:

```sql
UPDATE node_configs SET is_active = 0 WHERE node_id = 'acme-corp';
```

**What happens automatically:**

Within 30 seconds:
- LynqNode CR is deleted
- All associated resources are cleaned up (based on `DeletionPolicy`)
- Namespace is removed (if created)

No manual `kubectl delete` commands needed. The database is your source of truth.

---

## Explore the System

### Test Node Lifecycle

#### 1. Add a New Node

Add a row to the database:

```bash
# Connect to MySQL
kubectl exec -it deployment/mysql -n lynq-test -- \
  mysql -u root -p$(kubectl get secret mysql-root-password -n lynq-test -o jsonpath='{.data.password}' | base64 -d) nodes

# Insert new node
INSERT INTO node_configs (node_id, node_url, is_active, subscription_plan)
VALUES ('delta-co', 'https://delta.example.com', 1, 'enterprise');

exit
```

**Wait 30 seconds** (syncInterval), then verify:

```bash
kubectl get lynqnode delta-co-test-template
kubectl get deployment delta-co-app
```

#### 2. Deactivate a Node

```bash
# Update database
kubectl exec -it deployment/mysql -n lynq-test -- \
  mysql -u root -p$(kubectl get secret mysql-root-password -n lynq-test -o jsonpath='{.data.password}' | base64 -d) -e \
  "UPDATE nodes.node_configs SET is_active = 0 WHERE node_id = 'acme-corp';"
```

**Wait 30 seconds**, then verify resources are cleaned up:

```bash
kubectl get lynqnode acme-corp-test-template  # Not found
kubectl get deployment acme-corp-app        # Not found
```

#### 3. Modify Template

Edit the template to add more resources:

```bash
kubectl edit lynqform test-template
```

Changes automatically apply to all nodes. Monitor reconciliation:

```bash
kubectl get lynqnodes --watch
```

### View Metrics

```bash
# Port-forward metrics endpoint
kubectl port-forward -n lynq-system deployment/lynq-controller-manager 8080:8080

# View metrics
curl http://localhost:8080/metrics | grep lynqnode_
```

## Cleanup

### Option 1: Clean Resources Only

Keep the cluster, remove operator and nodes:

```bash
kubectl delete lynqnodes --all
kubectl delete lynqform test-template
kubectl delete lynqhub test-hub
kubectl delete deployment,service,pvc mysql -n lynq-test
kubectl delete deployment lynq-controller-manager -n lynq-system
```

### Option 2: Full Cleanup

Delete everything including Minikube cluster:

```bash
./scripts/cleanup-minikube.sh
```

This script interactively prompts for MySQL, operator, cluster, context, and image cache cleanup. Answer 'y' to all prompts for complete cleanup.

## Troubleshooting

### Quick Diagnostics

```bash
# Check operator status
kubectl get pods -n lynq-system
kubectl logs -n lynq-system -l control-plane=controller-manager

# Check hub sync
kubectl get lynqhub test-hub -o yaml

# Check node status
kubectl get lynqnode <lynqnode-name> -o yaml

# Check database connection
kubectl exec -it deployment/mysql -n lynq-test -- \
  mysql -u node_reader -p$(kubectl get secret mysql-credentials -n lynq-test -o jsonpath='{.data.password}' | base64 -d) \
  -e "SELECT * FROM nodes.node_configs;"
```

**Common issues:**
- **Operator not starting**: Check cert-manager is ready (`kubectl get pods -n cert-manager`)
- **Nodes not created**: Verify MySQL is ready and `is_active = 1` in database
- **Resources missing**: Check LynqNode CR status and operator logs

::: tip Detailed Troubleshooting
For comprehensive troubleshooting, see [Troubleshooting Guide](troubleshooting.md).
:::

## Customizing Scripts

All scripts support environment variables for customization:

```bash
# Example: Custom cluster configuration
MINIKUBE_CPUS=8 MINIKUBE_MEMORY=16384 ./scripts/setup-minikube.sh

# Example: Custom image tag
IMG=lynq:my-tag ./scripts/deploy-to-minikube.sh

# Example: Custom namespace
MYSQL_NAMESPACE=my-test-ns ./scripts/deploy-mysql.sh
```

Run any script with `--help` for full options.

## What's Next?

Now that you have Lynq running, explore these topics:

### Concepts & Configuration

- [**Templates Guide**](templates.md) - Template syntax and 200+ functions
- [**Policies Guide**](policies.md) - CreationPolicy, DeletionPolicy, ConflictPolicy, PatchStrategy
- [**DataSource Guide**](datasource.md) - MySQL configuration, VIEWs, and extraValueMappings
- [**Dependencies**](dependencies.md) - Resource ordering with dependency graphs

### Operations

- [**Installation Guide**](installation.md) - Deploy to production clusters
- [**Security Guide**](security.md) - RBAC and secrets management
- [**Performance Guide**](performance.md) - Scaling and optimization
- [**Monitoring Guide**](monitoring.md) - Prometheus metrics, alerts, and Grafana dashboards

### Advanced Topics

- [**Local Development**](local-development-minikube.md) - Development workflow and debugging
- [**Integration with External DNS**](integration-external-dns.md) - Automatic DNS per node
- [**Integration with Terraform Operator**](integration-terraform-operator.md) - Cloud resource provisioning

## Summary

You've successfully:
- ✅ Set up Minikube with Lynq in ~5 minutes
- ✅ Deployed MySQL with sample node data
- ✅ Created LynqHub and LynqForm
- ✅ Provisioned nodes automatically from database
- ✅ Tested node lifecycle (create, update, delete)

**Next:** Experiment with templates, policies, and template functions to build your database-driven platform!

## Need Help?

- 📖 **Documentation**: See [documentation site](./) for detailed guides
- 🐛 **Issues**: [GitHub Issues](https://github.com/k8s-lynq/lynq/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/k8s-lynq/lynq/discussions)
