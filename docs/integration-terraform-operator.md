# Terraform Operator Integration Guide

This guide shows how to integrate Tenant Operator with Terraform Operator for provisioning external cloud resources (AWS, GCP, Azure) per tenant.

[[toc]]

## Overview

**Terraform Operator** allows you to manage Terraform resources as Kubernetes Custom Resources. When integrated with Tenant Operator, each tenant can automatically provision **any infrastructure resource** that Terraform supports - from cloud services to on-premises systems.

### Key Benefits

**Universal Resource Provisioning**: Terraform supports 3,000+ providers, enabling you to provision virtually any infrastructure:
- ☁️ **Cloud Resources**: AWS, GCP, Azure, DigitalOcean, Alibaba Cloud
- 📦 **Databases**: PostgreSQL, MySQL, MongoDB, Cassandra, DynamoDB
- 📬 **Messaging Systems**: Kafka, RabbitMQ, Pulsar, ActiveMQ, AWS SQS/SNS
- 🔍 **Search & Analytics**: Elasticsearch, OpenSearch, Splunk
- 🗄️ **Caching**: Redis, Memcached, AWS ElastiCache
- 🌐 **DNS & CDN**: Route53, Cloudflare, Akamai, Fastly
- 🔐 **Security**: Vault, Auth0, Keycloak, AWS IAM
- 📊 **Monitoring**: Datadog, New Relic, PagerDuty
- 🏢 **On-Premises**: VMware vSphere, Proxmox, Bare Metal

**Automatic Lifecycle Management**:
- ✅ **Provisioning**: Resources created when tenant is activated (`activate=1`)
- 🔄 **Drift Detection**: Terraform ensures desired state matches actual state
- 🗑️ **Cleanup**: Resources automatically destroyed when tenant is deleted
- 📦 **Consistent State**: All tenant infrastructure managed declaratively

### Use Cases

#### Cloud Services (AWS, GCP, Azure)
- **S3/GCS/Blob Storage**: Isolated storage per tenant
- **RDS/Cloud SQL**: Dedicated databases per tenant
- **CloudFront/Cloud CDN**: Tenant-specific CDN distributions
- **IAM Roles/Policies**: Tenant-specific access control
- **VPCs/Subnets**: Network isolation
- **ElastiCache/Memorystore**: Per-tenant caching layers
- **Lambda/Cloud Functions**: Serverless functions per tenant

#### Messaging & Streaming
- **Kafka Topics**: Dedicated topics and ACLs per tenant
- **RabbitMQ VHosts**: Virtual hosts and users per tenant
- **AWS SQS/SNS**: Queue and topic isolation
- **Pulsar Namespaces**: Tenant-isolated messaging
- **NATS Accounts**: Multi-tenant streaming

#### Databases (Self-Managed & Managed)
- **PostgreSQL Schemas**: Isolated schemas in shared cluster
- **MongoDB Databases**: Dedicated databases with authentication
- **Redis Databases**: Separate database indexes per tenant
- **Elasticsearch Indices**: Tenant-specific indices with ILM policies
- **InfluxDB Organizations**: Time-series data isolation

#### On-Premises & Hybrid
- **VMware VMs**: Provision VMs per tenant
- **Proxmox Containers**: Lightweight tenant isolation
- **F5 Load Balancer**: Per-tenant virtual servers
- **NetBox IPAM**: IP address allocation per tenant

## Prerequisites

::: info Requirements
- Kubernetes cluster v1.16+
- Tenant Operator installed
- Cloud provider account (AWS, GCP, or Azure)
- Terraform ≥ 1.0
- Cloud provider credentials (stored as Secrets)
:::

## Installation

### 1. Install Tofu Controller

We'll use **tofu-controller** (formerly tf-controller), which is the production-ready Flux controller for managing Terraform/OpenTofu resources.

::: info Project evolution
The original Weave tf-controller has evolved into tofu-controller, now maintained by the Flux community: https://github.com/flux-iac/tofu-controller
:::

#### Installation via Helm (Recommended)

```bash
# Install Flux (required)
flux install

# Add tofu-controller Helm repository
helm repo add tofu-controller https://flux-iac.github.io/tofu-controller
helm repo update

# Install tofu-controller
helm install tofu-controller tofu-controller/tofu-controller \
  --namespace flux-system \
  --create-namespace
```

#### Installation via Manifests

```bash
# Install Flux
flux install

# Install tofu-controller CRDs and controller
kubectl apply -f https://raw.githubusercontent.com/flux-iac/tofu-controller/main/config/crd/bases/infra.contrib.fluxcd.io_terraforms.yaml

kubectl apply -f https://raw.githubusercontent.com/flux-iac/tofu-controller/main/config/rbac/role.yaml
kubectl apply -f https://raw.githubusercontent.com/flux-iac/tofu-controller/main/config/rbac/role_binding.yaml
kubectl apply -f https://raw.githubusercontent.com/flux-iac/tofu-controller/main/config/manager/deployment.yaml
```

#### Verify Installation

```bash
# Check tofu-controller pod
kubectl get pods -n flux-system -l app=tofu-controller

# Check CRD
kubectl get crd terraforms.infra.contrib.fluxcd.io

# Check controller logs
kubectl logs -n flux-system -l app=tofu-controller
```

### 2. Create Cloud Provider Credentials

#### AWS Credentials

```bash
# Create AWS credentials secret
kubectl create secret generic aws-credentials \
  --namespace default \
  --from-literal=AWS_ACCESS_KEY_ID=your-access-key \
  --from-literal=AWS_SECRET_ACCESS_KEY=your-secret-key \
  --from-literal=AWS_DEFAULT_REGION=us-east-1
```

#### GCP Credentials

```bash
# Create GCP service account key secret
kubectl create secret generic gcp-credentials \
  --namespace default \
  --from-file=credentials.json=path/to/your-service-account-key.json
```

#### Azure Credentials

```bash
# Create Azure credentials secret
kubectl create secret generic azure-credentials \
  --namespace default \
  --from-literal=ARM_CLIENT_ID=your-client-id \
  --from-literal=ARM_CLIENT_SECRET=your-client-secret \
  --from-literal=ARM_TENANT_ID=your-tenant-id \
  --from-literal=ARM_SUBSCRIPTION_ID=your-subscription-id
```

### 3. Verify Installation

```bash
# Check tf-controller pod
kubectl get pods -n flux-system -l app=tf-controller

# Check CRD
kubectl get crd terraforms.infra.contrib.fluxcd.io
```

## Basic Integration

### Example 1: S3 Bucket per Tenant

**TenantTemplate with Terraform manifest:**

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: tenant-with-s3
  namespace: default
spec:
  registryId: my-registry

  # Terraform resource for S3 bucket
  manifests:
  - id: s3-bucket
    nameTemplate: "{{ .uid }}-s3"
    spec:
      apiVersion: infra.contrib.fluxcd.io/v1alpha2
      kind: Terraform
      metadata:
        annotations:
          tenant-operator.kubernetes-tenants.org/tenant-id: "{{ .uid }}"
      spec:
        interval: 5m
        retryInterval: 30s

        # Terraform source (inline or from Git)
        sourceRef:
          kind: GitRepository
          name: terraform-modules
          namespace: default

        # Or use inline Terraform code
        path: ""

        # Inline Terraform HCL
        values:
          hcl: |
            terraform {
              required_providers {
                aws = {
                  source  = "hashicorp/aws"
                  version = "~> 5.0"
                }
              }
              backend "kubernetes" {
                secret_suffix = "{{ .uid }}-s3"
                namespace     = "default"
              }
            }

            provider "aws" {
              region = var.aws_region
            }

            variable "tenant_id" {
              type = string
            }

            variable "aws_region" {
              type    = string
              default = "us-east-1"
            }

            resource "aws_s3_bucket" "tenant_bucket" {
              bucket = "tenant-${var.tenant_id}-bucket"

              tags = {
                Name        = "Tenant ${var.tenant_id} Bucket"
                TenantId    = var.tenant_id
                ManagedBy   = "tenant-operator"
              }
            }

            resource "aws_s3_bucket_versioning" "tenant_bucket_versioning" {
              bucket = aws_s3_bucket.tenant_bucket.id

              versioning_configuration {
                status = "Enabled"
              }
            }

            resource "aws_s3_bucket_server_side_encryption_configuration" "tenant_bucket_encryption" {
              bucket = aws_s3_bucket.tenant_bucket.id

              rule {
                apply_server_side_encryption_by_default {
                  sse_algorithm = "AES256"
                }
              }
            }

            output "bucket_name" {
              value = aws_s3_bucket.tenant_bucket.id
            }

            output "bucket_arn" {
              value = aws_s3_bucket.tenant_bucket.arn
            }

            output "bucket_region" {
              value = aws_s3_bucket.tenant_bucket.region
            }

        # Variables passed to Terraform
        vars:
        - name: tenant_id
          value: "{{ .uid }}"
        - name: aws_region
          value: "us-east-1"

        # Use AWS credentials from secret
        varsFrom:
        - kind: Secret
          name: aws-credentials

        # Write Terraform outputs to ConfigMap
        writeOutputsToSecret:
          name: "{{ .uid }}-s3-outputs"

  # ConfigMap referencing Terraform outputs
  configMaps:
  - id: app-config
    nameTemplate: "{{ .uid }}-config"
    dependIds: ["s3-bucket"]
    spec:
      apiVersion: v1
      kind: ConfigMap
      data:
        tenant_id: "{{ .uid }}"
        # Note: Outputs will be in the secret created by Terraform
        s3_outputs_secret: "{{ .uid }}-s3-outputs"

  # Application using S3 bucket
  deployments:
  - id: app-deploy
    nameTemplate: "{{ .uid }}-app"
    dependIds: ["s3-bucket", "app-config"]
    waitForReady: true
    timeoutSeconds: 600  # Wait up to 10 minutes for Terraform
    spec:
      apiVersion: apps/v1
      kind: Deployment
      spec:
        replicas: 1
        selector:
          matchLabels:
            app: "{{ .uid }}"
        template:
          metadata:
            labels:
              app: "{{ .uid }}"
          spec:
            containers:
            - name: app
              image: mycompany/app:latest
              env:
              - name: TENANT_ID
                value: "{{ .uid }}"
              # S3 bucket name from Terraform output
              - name: S3_BUCKET_NAME
                valueFrom:
                  secretKeyRef:
                    name: "{{ .uid }}-s3-outputs"
                    key: bucket_name
              - name: AWS_REGION
                valueFrom:
                  secretKeyRef:
                    name: aws-credentials
                    key: AWS_DEFAULT_REGION
              envFrom:
              - secretRef:
                  name: aws-credentials
```

## Advanced Examples

### Example 2: RDS PostgreSQL Database per Tenant

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: tenant-with-rds
  namespace: default
spec:
  registryId: my-registry

  manifests:
  - id: rds-database
    nameTemplate: "{{ .uid }}-rds"
    creationPolicy: Once  # Create once, don't modify
    deletionPolicy: Retain  # Keep database when tenant deleted
    spec:
      apiVersion: infra.contrib.fluxcd.io/v1alpha2
      kind: Terraform
      spec:
        interval: 10m
        retryInterval: 1m

        values:
          hcl: |
            terraform {
              required_providers {
                aws = {
                  source  = "hashicorp/aws"
                  version = "~> 5.0"
                }
                random = {
                  source  = "hashicorp/random"
                  version = "~> 3.5"
                }
              }
              backend "kubernetes" {
                secret_suffix = "{{ .uid }}-rds"
                namespace     = "default"
              }
            }

            provider "aws" {
              region = var.aws_region
            }

            variable "tenant_id" {
              type = string
            }

            variable "aws_region" {
              type    = string
              default = "us-east-1"
            }

            variable "db_instance_class" {
              type    = string
              default = "db.t3.micro"
            }

            variable "db_allocated_storage" {
              type    = number
              default = 20
            }

            # Generate random password
            resource "random_password" "db_password" {
              length  = 32
              special = true
            }

            # Security group for RDS
            resource "aws_security_group" "rds_sg" {
              name_prefix = "tenant-${var.tenant_id}-rds-"
              description = "Security group for tenant ${var.tenant_id} RDS"

              ingress {
                from_port   = 5432
                to_port     = 5432
                protocol    = "tcp"
                cidr_blocks = ["10.0.0.0/8"]  # Adjust to your VPC CIDR
              }

              egress {
                from_port   = 0
                to_port     = 0
                protocol    = "-1"
                cidr_blocks = ["0.0.0.0/0"]
              }

              tags = {
                Name      = "tenant-${var.tenant_id}-rds-sg"
                TenantId  = var.tenant_id
                ManagedBy = "tenant-operator"
              }
            }

            # RDS instance
            resource "aws_db_instance" "tenant_db" {
              identifier     = "tenant-${var.tenant_id}-db"
              engine         = "postgres"
              engine_version = "15.4"

              instance_class    = var.db_instance_class
              allocated_storage = var.db_allocated_storage
              storage_type      = "gp3"
              storage_encrypted = true

              db_name  = "tenant_${replace(var.tenant_id, "-", "_")}"
              username = "dbadmin"
              password = random_password.db_password.result

              vpc_security_group_ids = [aws_security_group.rds_sg.id]

              backup_retention_period = 7
              backup_window          = "03:00-04:00"
              maintenance_window     = "mon:04:00-mon:05:00"

              skip_final_snapshot = false
              final_snapshot_identifier = "tenant-${var.tenant_id}-final-snapshot"

              tags = {
                Name      = "tenant-${var.tenant_id}-db"
                TenantId  = var.tenant_id
                ManagedBy = "tenant-operator"
              }
            }

            output "db_endpoint" {
              value     = aws_db_instance.tenant_db.endpoint
              sensitive = false
            }

            output "db_name" {
              value = aws_db_instance.tenant_db.db_name
            }

            output "db_username" {
              value = aws_db_instance.tenant_db.username
            }

            output "db_password" {
              value     = random_password.db_password.result
              sensitive = true
            }

            output "db_port" {
              value = aws_db_instance.tenant_db.port
            }

        vars:
        - name: tenant_id
          value: "{{ .uid }}"

        varsFrom:
        - kind: Secret
          name: aws-credentials

        writeOutputsToSecret:
          name: "{{ .uid }}-db-credentials"

  # Application using RDS
  deployments:
  - id: app-deploy
    nameTemplate: "{{ .uid }}-app"
    dependIds: ["rds-database"]
    timeoutSeconds: 900  # 15 minutes for RDS provisioning
    spec:
      apiVersion: apps/v1
      kind: Deployment
      spec:
        replicas: 2
        selector:
          matchLabels:
            app: "{{ .uid }}"
        template:
          metadata:
            labels:
              app: "{{ .uid }}"
          spec:
            containers:
            - name: app
              image: mycompany/app:latest
              env:
              - name: TENANT_ID
                value: "{{ .uid }}"
              - name: DB_HOST
                valueFrom:
                  secretKeyRef:
                    name: "{{ .uid }}-db-credentials"
                    key: db_endpoint
              - name: DB_NAME
                valueFrom:
                  secretKeyRef:
                    name: "{{ .uid }}-db-credentials"
                    key: db_name
              - name: DB_USER
                valueFrom:
                  secretKeyRef:
                    name: "{{ .uid }}-db-credentials"
                    key: db_username
              - name: DB_PASSWORD
                valueFrom:
                  secretKeyRef:
                    name: "{{ .uid }}-db-credentials"
                    key: db_password
              - name: DB_PORT
                valueFrom:
                  secretKeyRef:
                    name: "{{ .uid }}-db-credentials"
                    key: db_port
```

### Example 3: CloudFront CDN Distribution

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: tenant-with-cdn
  namespace: default
spec:
  registryId: my-registry

  manifests:
  - id: cloudfront-cdn
    nameTemplate: "{{ .uid }}-cdn"
    spec:
      apiVersion: infra.contrib.fluxcd.io/v1alpha2
      kind: Terraform
      spec:
        interval: 5m

        values:
          hcl: |
            terraform {
              required_providers {
                aws = {
                  source  = "hashicorp/aws"
                  version = "~> 5.0"
                }
              }
              backend "kubernetes" {
                secret_suffix = "{{ .uid }}-cdn"
                namespace     = "default"
              }
            }

            provider "aws" {
              region = "us-east-1"  # CloudFront is global
            }

            variable "tenant_id" {
              type = string
            }

            variable "origin_domain" {
              type = string
            }

            # S3 bucket for CDN logs
            resource "aws_s3_bucket" "cdn_logs" {
              bucket = "tenant-${var.tenant_id}-cdn-logs"

              tags = {
                Name      = "tenant-${var.tenant_id}-cdn-logs"
                TenantId  = var.tenant_id
                ManagedBy = "tenant-operator"
              }
            }

            # CloudFront distribution
            resource "aws_cloudfront_distribution" "cdn" {
              enabled             = true
              is_ipv6_enabled     = true
              comment             = "CDN for tenant ${var.tenant_id}"
              default_root_object = "index.html"

              origin {
                domain_name = var.origin_domain
                origin_id   = "tenant-${var.tenant_id}-origin"

                custom_origin_config {
                  http_port              = 80
                  https_port             = 443
                  origin_protocol_policy = "https-only"
                  origin_ssl_protocols   = ["TLSv1.2"]
                }
              }

              default_cache_behavior {
                allowed_methods  = ["GET", "HEAD", "OPTIONS"]
                cached_methods   = ["GET", "HEAD"]
                target_origin_id = "tenant-${var.tenant_id}-origin"

                forwarded_values {
                  query_string = true
                  cookies {
                    forward = "none"
                  }
                }

                viewer_protocol_policy = "redirect-to-https"
                min_ttl                = 0
                default_ttl            = 3600
                max_ttl                = 86400
                compress               = true
              }

              restrictions {
                geo_restriction {
                  restriction_type = "none"
                }
              }

              viewer_certificate {
                cloudfront_default_certificate = true
              }

              logging_config {
                include_cookies = false
                bucket          = aws_s3_bucket.cdn_logs.bucket_domain_name
                prefix          = "cdn-logs/"
              }

              tags = {
                Name      = "tenant-${var.tenant_id}-cdn"
                TenantId  = var.tenant_id
                ManagedBy = "tenant-operator"
              }
            }

            output "cdn_domain_name" {
              value = aws_cloudfront_distribution.cdn.domain_name
            }

            output "cdn_distribution_id" {
              value = aws_cloudfront_distribution.cdn.id
            }

            output "cdn_arn" {
              value = aws_cloudfront_distribution.cdn.arn
            }

        vars:
        - name: tenant_id
          value: "{{ .uid }}"
        - name: origin_domain
          value: "{{ .host }}"

        varsFrom:
        - kind: Secret
          name: aws-credentials

        writeOutputsToSecret:
          name: "{{ .uid }}-cdn-outputs"
```

### Example 4: Using Git Repository for Terraform Modules

**Create GitRepository:**

```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: terraform-modules
  namespace: default
spec:
  interval: 5m
  url: https://github.com/your-org/terraform-modules
  ref:
    branch: main
  # Optional: Use SSH key for private repos
  # secretRef:
  #   name: git-credentials
```

**TenantTemplate using Git modules:**

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: tenant-with-git-modules
  namespace: default
spec:
  registryId: my-registry

  manifests:
  - id: tenant-infrastructure
    nameTemplate: "{{ .uid }}-infra"
    spec:
      apiVersion: infra.contrib.fluxcd.io/v1alpha2
      kind: Terraform
      spec:
        interval: 10m

        # Reference Git repository
        sourceRef:
          kind: GitRepository
          name: terraform-modules
          namespace: default

        # Path to module in repository
        path: ./modules/tenant-stack

        # Pass variables to module
        vars:
        - name: tenant_id
          value: "{{ .uid }}"
        - name: tenant_host
          value: "{{ .host }}"
        - name: environment
          value: "production"

        varsFrom:
        - kind: Secret
          name: aws-credentials

        writeOutputsToSecret:
          name: "{{ .uid }}-infra-outputs"
```

**Example Terraform module structure in Git:**

```
terraform-modules/
├── modules/
│   ├── tenant-stack/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   ├── s3.tf
│   │   ├── rds.tf
│   │   └── cloudfront.tf
│   ├── networking/
│   └── security/
└── README.md
```

### Example 5: Kafka Topics and ACLs per Tenant

Provision dedicated Kafka topics and access controls for each tenant:

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: tenant-with-kafka
  namespace: default
spec:
  registryId: my-registry

  manifests:
  - id: kafka-resources
    nameTemplate: "{{ .uid }}-kafka"
    spec:
      apiVersion: infra.contrib.fluxcd.io/v1alpha2
      kind: Terraform
      spec:
        interval: 5m

        values:
          hcl: |
            terraform {
              required_providers {
                kafka = {
                  source  = "Mongey/kafka"
                  version = "~> 0.7"
                }
              }
              backend "kubernetes" {
                secret_suffix = "{{ .uid }}-kafka"
                namespace     = "default"
              }
            }

            provider "kafka" {
              bootstrap_servers = var.kafka_bootstrap_servers
              tls_enabled       = true
              sasl_username     = var.kafka_username
              sasl_password     = var.kafka_password
              sasl_mechanism    = "plain"
            }

            variable "tenant_id" { type = string }
            variable "kafka_bootstrap_servers" { type = list(string) }
            variable "kafka_username" { type = string }
            variable "kafka_password" { type = string sensitive = true }

            # Topics for tenant
            resource "kafka_topic" "events" {
              name               = "tenant-${var.tenant_id}-events"
              replication_factor = 3
              partitions         = 6

              config = {
                "cleanup.policy" = "delete"
                "retention.ms"   = "604800000"  # 7 days
                "segment.ms"     = "86400000"   # 1 day
              }
            }

            resource "kafka_topic" "commands" {
              name               = "tenant-${var.tenant_id}-commands"
              replication_factor = 3
              partitions         = 3

              config = {
                "cleanup.policy" = "delete"
                "retention.ms"   = "259200000"  # 3 days
              }
            }

            resource "kafka_topic" "dlq" {
              name               = "tenant-${var.tenant_id}-dlq"
              replication_factor = 3
              partitions         = 1

              config = {
                "cleanup.policy" = "delete"
                "retention.ms"   = "2592000000"  # 30 days
              }
            }

            # ACLs for tenant
            resource "kafka_acl" "tenant_producer" {
              resource_name             = "tenant-${var.tenant_id}-*"
              resource_type             = "Topic"
              acl_principal             = "User:tenant-${var.tenant_id}"
              acl_host                  = "*"
              acl_operation             = "Write"
              acl_permission_type       = "Allow"
              resource_pattern_type_filter = "Prefixed"
            }

            resource "kafka_acl" "tenant_consumer" {
              resource_name             = "tenant-${var.tenant_id}-*"
              resource_type             = "Topic"
              acl_principal             = "User:tenant-${var.tenant_id}"
              acl_host                  = "*"
              acl_operation             = "Read"
              acl_permission_type       = "Allow"
              resource_pattern_type_filter = "Prefixed"
            }

            resource "kafka_acl" "tenant_consumer_group" {
              resource_name             = "tenant-${var.tenant_id}-*"
              resource_type             = "Group"
              acl_principal             = "User:tenant-${var.tenant_id}"
              acl_host                  = "*"
              acl_operation             = "Read"
              acl_permission_type       = "Allow"
              resource_pattern_type_filter = "Prefixed"
            }

            output "events_topic" { value = kafka_topic.events.name }
            output "commands_topic" { value = kafka_topic.commands.name }
            output "dlq_topic" { value = kafka_topic.dlq.name }

        vars:
        - name: tenant_id
          value: "{{ .uid }}"
        - name: kafka_bootstrap_servers
          value: '["kafka-broker-1:9092","kafka-broker-2:9092","kafka-broker-3:9092"]'

        varsFrom:
        - kind: Secret
          name: kafka-credentials

        writeOutputsToSecret:
          name: "{{ .uid }}-kafka-outputs"
```

### Example 6: RabbitMQ Virtual Host and User per Tenant

Provision isolated RabbitMQ resources for each tenant:

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: tenant-with-rabbitmq
  namespace: default
spec:
  registryId: my-registry

  manifests:
  - id: rabbitmq-resources
    nameTemplate: "{{ .uid }}-rabbitmq"
    spec:
      apiVersion: infra.contrib.fluxcd.io/v1alpha2
      kind: Terraform
      spec:
        interval: 5m

        values:
          hcl: |
            terraform {
              required_providers {
                rabbitmq = {
                  source  = "cyrilgdn/rabbitmq"
                  version = "~> 1.8"
                }
                random = {
                  source  = "hashicorp/random"
                  version = "~> 3.5"
                }
              }
              backend "kubernetes" {
                secret_suffix = "{{ .uid }}-rabbitmq"
                namespace     = "default"
              }
            }

            provider "rabbitmq" {
              endpoint = var.rabbitmq_endpoint
              username = var.rabbitmq_admin_user
              password = var.rabbitmq_admin_password
            }

            variable "tenant_id" { type = string }
            variable "rabbitmq_endpoint" { type = string }
            variable "rabbitmq_admin_user" { type = string }
            variable "rabbitmq_admin_password" { type = string sensitive = true }

            # Generate password for tenant user
            resource "random_password" "tenant_password" {
              length  = 32
              special = true
            }

            # Virtual host for tenant
            resource "rabbitmq_vhost" "tenant_vhost" {
              name = "tenant-${var.tenant_id}"
            }

            # User for tenant
            resource "rabbitmq_user" "tenant_user" {
              name     = "tenant-${var.tenant_id}"
              password = random_password.tenant_password.result
              tags     = []
            }

            # Permissions for tenant user on their vhost
            resource "rabbitmq_permissions" "tenant_permissions" {
              user  = rabbitmq_user.tenant_user.name
              vhost = rabbitmq_vhost.tenant_vhost.name

              permissions {
                configure = ".*"
                write     = ".*"
                read      = ".*"
              }
            }

            # Default exchanges and queues
            resource "rabbitmq_exchange" "tenant_events" {
              name  = "events"
              vhost = rabbitmq_vhost.tenant_vhost.name

              settings {
                type        = "topic"
                durable     = true
                auto_delete = false
              }
            }

            resource "rabbitmq_queue" "tenant_tasks" {
              name  = "tasks"
              vhost = rabbitmq_vhost.tenant_vhost.name

              settings {
                durable     = true
                auto_delete = false
                arguments = {
                  "x-message-ttl"          = 86400000  # 24 hours
                  "x-max-length"           = 10000
                  "x-queue-type"           = "quorum"
                }
              }
            }

            output "vhost" { value = rabbitmq_vhost.tenant_vhost.name }
            output "username" { value = rabbitmq_user.tenant_user.name }
            output "password" { value = random_password.tenant_password.result sensitive = true }
            output "connection_string" {
              value = "amqp://${rabbitmq_user.tenant_user.name}:${random_password.tenant_password.result}@${var.rabbitmq_endpoint}/${rabbitmq_vhost.tenant_vhost.name}"
              sensitive = true
            }

        vars:
        - name: tenant_id
          value: "{{ .uid }}"
        - name: rabbitmq_endpoint
          value: "rabbitmq.default.svc.cluster.local:5672"

        varsFrom:
        - kind: Secret
          name: rabbitmq-admin-credentials

        writeOutputsToSecret:
          name: "{{ .uid }}-rabbitmq-credentials"
```

### Example 7: PostgreSQL Schema and User per Tenant

Provision isolated PostgreSQL schemas in a shared database:

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: tenant-with-pg-schema
  namespace: default
spec:
  registryId: my-registry

  manifests:
  - id: postgresql-schema
    nameTemplate: "{{ .uid }}-pg-schema"
    creationPolicy: Once
    deletionPolicy: Retain
    spec:
      apiVersion: infra.contrib.fluxcd.io/v1alpha2
      kind: Terraform
      spec:
        interval: 5m

        values:
          hcl: |
            terraform {
              required_providers {
                postgresql = {
                  source  = "cyrilgdn/postgresql"
                  version = "~> 1.21"
                }
                random = {
                  source  = "hashicorp/random"
                  version = "~> 3.5"
                }
              }
              backend "kubernetes" {
                secret_suffix = "{{ .uid }}-pg"
                namespace     = "default"
              }
            }

            provider "postgresql" {
              host            = var.pg_host
              port            = var.pg_port
              database        = var.pg_database
              username        = var.pg_admin_user
              password        = var.pg_admin_password
              sslmode         = "require"
              connect_timeout = 15
            }

            variable "tenant_id" { type = string }
            variable "pg_host" { type = string }
            variable "pg_port" { type = number default = 5432 }
            variable "pg_database" { type = string }
            variable "pg_admin_user" { type = string }
            variable "pg_admin_password" { type = string sensitive = true }

            # Generate password for tenant
            resource "random_password" "tenant_password" {
              length  = 32
              special = true
            }

            # Schema for tenant
            resource "postgresql_schema" "tenant_schema" {
              name  = "tenant_${replace(var.tenant_id, "-", "_")}"
              owner = postgresql_role.tenant_user.name
            }

            # User/Role for tenant
            resource "postgresql_role" "tenant_user" {
              name     = "tenant_${replace(var.tenant_id, "-", "_")}"
              login    = true
              password = random_password.tenant_password.result
            }

            # Grant schema usage to tenant user
            resource "postgresql_grant" "schema_usage" {
              database    = var.pg_database
              role        = postgresql_role.tenant_user.name
              schema      = postgresql_schema.tenant_schema.name
              object_type = "schema"
              privileges  = ["USAGE", "CREATE"]
            }

            # Grant all privileges on tables in schema
            resource "postgresql_grant" "tables" {
              database    = var.pg_database
              role        = postgresql_role.tenant_user.name
              schema      = postgresql_schema.tenant_schema.name
              object_type = "table"
              privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE"]
            }

            output "schema_name" { value = postgresql_schema.tenant_schema.name }
            output "db_user" { value = postgresql_role.tenant_user.name }
            output "db_password" { value = random_password.tenant_password.result sensitive = true }
            output "connection_string" {
              value = "postgresql://${postgresql_role.tenant_user.name}:${random_password.tenant_password.result}@${var.pg_host}:${var.pg_port}/${var.pg_database}?options=-c%20search_path%3D${postgresql_schema.tenant_schema.name}"
              sensitive = true
            }

        vars:
        - name: tenant_id
          value: "{{ .uid }}"
        - name: pg_host
          value: "postgres.default.svc.cluster.local"
        - name: pg_database
          value: "tenants"

        varsFrom:
        - kind: Secret
          name: postgres-admin-credentials

        writeOutputsToSecret:
          name: "{{ .uid }}-postgres-credentials"
```

### Example 8: Redis Database per Tenant

Provision dedicated Redis database numbers for each tenant:

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: tenant-with-redis
  namespace: default
spec:
  registryId: my-registry

  manifests:
  - id: redis-database
    nameTemplate: "{{ .uid }}-redis"
    spec:
      apiVersion: infra.contrib.fluxcd.io/v1alpha2
      kind: Terraform
      spec:
        interval: 5m

        values:
          hcl: |
            terraform {
              required_providers {
                redis = {
                  source  = "redis/redis"
                  version = "~> 1.3"
                }
              }
              backend "kubernetes" {
                secret_suffix = "{{ .uid }}-redis"
                namespace     = "default"
              }
            }

            provider "redis" {
              address = var.redis_address
            }

            variable "tenant_id" { type = string }
            variable "redis_address" { type = string }
            variable "redis_db_number" { type = number }

            # Note: Redis doesn't have native ACLs for DB numbers in older versions
            # This example shows configuration; actual implementation may vary
            # For Redis 6+, use ACLs instead

            locals {
              db_number = var.redis_db_number
            }

            output "redis_host" { value = split(":", var.redis_address)[0] }
            output "redis_port" { value = split(":", var.redis_address)[1] }
            output "redis_db" { value = local.db_number }
            output "connection_string" {
              value = "redis://${var.redis_address}/${local.db_number}"
            }

        vars:
        - name: tenant_id
          value: "{{ .uid }}"
        - name: redis_address
          value: "redis.default.svc.cluster.local:6379"
        - name: redis_db_number
          value: "{{ .uid | sha1sum | trunc 2 }}"  # Generate DB number from tenant ID

        writeOutputsToSecret:
          name: "{{ .uid }}-redis-config"
```

## Complete Multi-Resource Example

Full example provisioning S3, RDS, and CloudFront for each tenant:

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: enterprise-tenant-stack
  namespace: default
spec:
  registryId: enterprise-registry

  # Terraform for complete infrastructure stack
  manifests:
  - id: tenant-infrastructure
    nameTemplate: "{{ .uid }}-infrastructure"
    creationPolicy: Once
    deletionPolicy: Retain
    timeoutSeconds: 1800  # 30 minutes
    spec:
      apiVersion: infra.contrib.fluxcd.io/v1alpha2
      kind: Terraform
      spec:
        interval: 15m
        retryInterval: 2m

        values:
          hcl: |
            terraform {
              required_providers {
                aws = {
                  source  = "hashicorp/aws"
                  version = "~> 5.0"
                }
                random = {
                  source  = "hashicorp/random"
                  version = "~> 3.5"
                }
              }
              backend "kubernetes" {
                secret_suffix = "{{ .uid }}-infra"
                namespace     = "default"
              }
            }

            provider "aws" {
              region = var.aws_region
            }

            variable "tenant_id" { type = string }
            variable "tenant_host" { type = string }
            variable "aws_region" { type = string }
            variable "db_instance_class" { type = string }

            # Random password for database
            resource "random_password" "db_password" {
              length  = 32
              special = true
            }

            # S3 bucket for tenant data
            resource "aws_s3_bucket" "tenant_data" {
              bucket = "tenant-${var.tenant_id}-data"
              tags = {
                TenantId = var.tenant_id
                Purpose  = "tenant-data"
              }
            }

            resource "aws_s3_bucket_versioning" "tenant_data_versioning" {
              bucket = aws_s3_bucket.tenant_data.id
              versioning_configuration {
                status = "Enabled"
              }
            }

            # S3 bucket for static assets
            resource "aws_s3_bucket" "tenant_static" {
              bucket = "tenant-${var.tenant_id}-static"
              tags = {
                TenantId = var.tenant_id
                Purpose  = "static-assets"
              }
            }

            resource "aws_s3_bucket_public_access_block" "tenant_static_pab" {
              bucket = aws_s3_bucket.tenant_static.id

              block_public_acls       = false
              block_public_policy     = false
              ignore_public_acls      = false
              restrict_public_buckets = false
            }

            # RDS PostgreSQL
            resource "aws_db_instance" "tenant_db" {
              identifier            = "tenant-${var.tenant_id}-db"
              engine                = "postgres"
              engine_version        = "15.4"
              instance_class        = var.db_instance_class
              allocated_storage     = 20
              storage_encrypted     = true
              db_name               = "tenant_${replace(var.tenant_id, "-", "_")}"
              username              = "dbadmin"
              password              = random_password.db_password.result
              skip_final_snapshot   = false
              final_snapshot_identifier = "tenant-${var.tenant_id}-final"

              tags = {
                TenantId = var.tenant_id
              }
            }

            # CloudFront distribution
            resource "aws_cloudfront_distribution" "tenant_cdn" {
              enabled         = true
              is_ipv6_enabled = true
              comment         = "CDN for ${var.tenant_id}"

              origin {
                domain_name = aws_s3_bucket.tenant_static.bucket_regional_domain_name
                origin_id   = "S3-${var.tenant_id}"
              }

              default_cache_behavior {
                allowed_methods        = ["GET", "HEAD"]
                cached_methods         = ["GET", "HEAD"]
                target_origin_id       = "S3-${var.tenant_id}"
                viewer_protocol_policy = "redirect-to-https"

                forwarded_values {
                  query_string = false
                  cookies {
                    forward = "none"
                  }
                }
              }

              restrictions {
                geo_restriction {
                  restriction_type = "none"
                }
              }

              viewer_certificate {
                cloudfront_default_certificate = true
              }

              tags = {
                TenantId = var.tenant_id
              }
            }

            # IAM user for tenant access
            resource "aws_iam_user" "tenant_user" {
              name = "tenant-${var.tenant_id}-user"
              tags = {
                TenantId = var.tenant_id
              }
            }

            resource "aws_iam_access_key" "tenant_access_key" {
              user = aws_iam_user.tenant_user.name
            }

            # IAM policy for tenant S3 access
            resource "aws_iam_user_policy" "tenant_s3_policy" {
              name = "tenant-${var.tenant_id}-s3-policy"
              user = aws_iam_user.tenant_user.name

              policy = jsonencode({
                Version = "2012-10-17"
                Statement = [
                  {
                    Effect = "Allow"
                    Action = [
                      "s3:GetObject",
                      "s3:PutObject",
                      "s3:DeleteObject",
                      "s3:ListBucket"
                    ]
                    Resource = [
                      aws_s3_bucket.tenant_data.arn,
                      "${aws_s3_bucket.tenant_data.arn}/*"
                    ]
                  }
                ]
              })
            }

            # Outputs
            output "s3_data_bucket" { value = aws_s3_bucket.tenant_data.id }
            output "s3_static_bucket" { value = aws_s3_bucket.tenant_static.id }
            output "db_endpoint" { value = aws_db_instance.tenant_db.endpoint }
            output "db_name" { value = aws_db_instance.tenant_db.db_name }
            output "db_username" { value = aws_db_instance.tenant_db.username }
            output "db_password" { value = random_password.db_password.result sensitive = true }
            output "cdn_domain" { value = aws_cloudfront_distribution.tenant_cdn.domain_name }
            output "cdn_distribution_id" { value = aws_cloudfront_distribution.tenant_cdn.id }
            output "iam_access_key_id" { value = aws_iam_access_key.tenant_access_key.id }
            output "iam_secret_access_key" { value = aws_iam_access_key.tenant_access_key.secret sensitive = true }

        vars:
        - name: tenant_id
          value: "{{ .uid }}"
        - name: tenant_host
          value: "{{ .host }}"
        - name: aws_region
          value: "us-east-1"
        - name: db_instance_class
          value: "db.t3.micro"

        varsFrom:
        - kind: Secret
          name: aws-credentials

        writeOutputsToSecret:
          name: "{{ .uid }}-infrastructure"

  # ConfigMap with infrastructure info
  configMaps:
  - id: infra-config
    nameTemplate: "{{ .uid }}-infra-config"
    dependIds: ["tenant-infrastructure"]
    spec:
      apiVersion: v1
      kind: ConfigMap
      data:
        tenant_id: "{{ .uid }}"
        terraform_outputs_secret: "{{ .uid }}-infrastructure"

  # Application deployment
  deployments:
  - id: app-deploy
    nameTemplate: "{{ .uid }}-app"
    dependIds: ["tenant-infrastructure"]
    spec:
      apiVersion: apps/v1
      kind: Deployment
      spec:
        replicas: 2
        selector:
          matchLabels:
            app: "{{ .uid }}"
        template:
          metadata:
            labels:
              app: "{{ .uid }}"
          spec:
            containers:
            - name: app
              image: mycompany/enterprise-app:latest
              env:
              # Database connection
              - name: DB_HOST
                valueFrom:
                  secretKeyRef:
                    name: "{{ .uid }}-infrastructure"
                    key: db_endpoint
              - name: DB_NAME
                valueFrom:
                  secretKeyRef:
                    name: "{{ .uid }}-infrastructure"
                    key: db_name
              - name: DB_USER
                valueFrom:
                  secretKeyRef:
                    name: "{{ .uid }}-infrastructure"
                    key: db_username
              - name: DB_PASSWORD
                valueFrom:
                  secretKeyRef:
                    name: "{{ .uid }}-infrastructure"
                    key: db_password

              # S3 buckets
              - name: S3_DATA_BUCKET
                valueFrom:
                  secretKeyRef:
                    name: "{{ .uid }}-infrastructure"
                    key: s3_data_bucket
              - name: S3_STATIC_BUCKET
                valueFrom:
                  secretKeyRef:
                    name: "{{ .uid }}-infrastructure"
                    key: s3_static_bucket

              # CloudFront CDN
              - name: CDN_DOMAIN
                valueFrom:
                  secretKeyRef:
                    name: "{{ .uid }}-infrastructure"
                    key: cdn_domain

              # IAM credentials
              - name: AWS_ACCESS_KEY_ID
                valueFrom:
                  secretKeyRef:
                    name: "{{ .uid }}-infrastructure"
                    key: iam_access_key_id
              - name: AWS_SECRET_ACCESS_KEY
                valueFrom:
                  secretKeyRef:
                    name: "{{ .uid }}-infrastructure"
                    key: iam_secret_access_key

              - name: TENANT_ID
                value: "{{ .uid }}"
```

## How It Works

### Workflow

1. **Tenant Created**: TenantRegistry creates Tenant CR from database
2. **Terraform Applied**: Tenant controller creates Terraform CR
3. **tf-controller Processes**: Runs terraform init/plan/apply
4. **Resources Provisioned**: Cloud resources created (S3, RDS, etc.)
5. **Outputs Saved**: Terraform outputs written to Kubernetes Secret
6. **App Deployed**: Application uses infrastructure via Secret references
7. **Tenant Deleted**: Terraform runs destroy (if deletionPolicy=Delete)

### State Management

Terraform state is stored in Kubernetes Secrets by default:

```
Secret: tfstate-default-{tenant-id}-{resource-name}
Namespace: default
Data: tfstate (gzipped)
```

## Best Practices

### 1. Use CreationPolicy: Once for Immutable Infrastructure

```yaml
manifests:
- id: rds-database
  creationPolicy: Once  # Create once, never update
  deletionPolicy: Retain  # Keep on tenant deletion
```

### 2. Set Appropriate Timeouts

Terraform provisioning can take 10-30 minutes:

```yaml
deployments:
- id: app
  dependIds: ["terraform-resources"]
  timeoutSeconds: 1800  # 30 minutes
```

### 3. Use Remote State Backend (Production)

For production, use S3 backend instead of Kubernetes:

```hcl
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "tenants/${var.tenant_id}/terraform.tfstate"
    region = "us-east-1"
    encrypt = true
    dynamodb_table = "terraform-locks"
  }
}
```

### 4. Secure Sensitive Outputs

Mark sensitive outputs:

```hcl
output "db_password" {
  value     = random_password.db_password.result
  sensitive = true
}
```

### 5. Use Dependency Ordering

Ensure proper resource creation order:

```yaml
deployments:
- id: app
  dependIds: ["tenant-infrastructure"]  # Wait for Terraform
  waitForReady: true
```

### 6. Monitor Terraform Resources

```bash
# Check Terraform resources
kubectl get terraform -n default

# Check specific tenant's Terraform
kubectl get terraform -n default -l tenant-operator.kubernetes-tenants.org/tenant-id=tenant-alpha

# View Terraform plan
kubectl describe terraform tenant-alpha-infrastructure

# View Terraform outputs
kubectl get secret tenant-alpha-infrastructure -o yaml
```

## Troubleshooting

### Terraform Apply Fails

**Problem:** Terraform fails to apply resources.

**Solution:**

1. **Check Terraform logs:**
   ```bash
   kubectl logs -n flux-system -l app=tf-controller
   ```

2. **Check Terraform CR status:**
   ```bash
   kubectl describe terraform tenant-alpha-infrastructure
   ```

3. **View Terraform plan output:**
   ```bash
   kubectl get terraform tenant-alpha-infrastructure -o jsonpath='{.status.plan.pending}'
   ```

4. **Check credentials:**
   ```bash
   kubectl get secret aws-credentials -o yaml
   ```

### State Lock Issues

**Problem:** Terraform state locked.

**Solution:**

```bash
# Force unlock (use with caution!)
# This requires accessing the Terraform pod
kubectl exec -it -n flux-system tf-controller-xxx -- sh
terraform force-unlock <lock-id>
```

### Outputs Not Available

**Problem:** Terraform outputs not written to secret.

**Solution:**

1. **Verify writeOutputsToSecret is set:**
   ```yaml
   writeOutputsToSecret:
     name: "{{ .uid }}-outputs"
   ```

2. **Check if Terraform apply completed:**
   ```bash
   kubectl get terraform tenant-alpha-infra -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'
   ```

3. **Check secret exists:**
   ```bash
   kubectl get secret tenant-alpha-outputs
   ```

### Resource Already Exists

**Problem:** Terraform fails because resource already exists.

**Solution:**

Use `terraform import` or recreate with different name:

```hcl
resource "aws_s3_bucket" "tenant_bucket" {
  bucket = "tenant-${var.tenant_id}-bucket-v2"  # Add suffix
}
```

## Cost Optimization

### 1. Use Appropriate Instance Sizes

```hcl
variable "db_instance_class" {
  type = string
  default = "db.t3.micro"  # ~$15/month
}
```

### 2. Enable Auto-Scaling

```hcl
resource "aws_appautoscaling_target" "rds_target" {
  max_capacity       = 10
  min_capacity       = 1
  resource_id        = "cluster:${aws_rds_cluster.tenant_db.cluster_identifier}"
  scalable_dimension = "rds:cluster:ReadReplicaCount"
  service_namespace  = "rds"
}
```

### 3. Use Lifecycle Policies

```hcl
resource "aws_s3_bucket_lifecycle_configuration" "tenant_bucket_lifecycle" {
  bucket = aws_s3_bucket.tenant_bucket.id

  rule {
    id     = "archive-old-data"
    status = "Enabled"

    transition {
      days          = 90
      storage_class = "GLACIER"
    }

    expiration {
      days = 365
    }
  }
}
```

## See Also

- [Tofu Controller (OpenTofu/Terraform)](https://github.com/flux-iac/tofu-controller)
- [Flux Documentation](https://fluxcd.io/docs/)
- [Terraform Registry - All Providers](https://registry.terraform.io/browse/providers)
- [ExternalDNS Integration](integration-external-dns.md)
- [Tenant Operator Templates Guide](templates.md)
- [AWS Terraform Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [Kafka Terraform Provider](https://registry.terraform.io/providers/Mongey/kafka/latest/docs)
- [RabbitMQ Terraform Provider](https://registry.terraform.io/providers/cyrilgdn/rabbitmq/latest/docs)
- [PostgreSQL Terraform Provider](https://registry.terraform.io/providers/cyrilgdn/postgresql/latest/docs)
- [Elasticsearch Terraform Provider](https://registry.terraform.io/providers/phillbaker/elasticsearch/latest/docs)
