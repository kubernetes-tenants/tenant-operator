# Database-per-Tenant with Crossplane

## Overview

Provision isolated cloud databases (RDS, Cloud SQL) automatically for each tenant using Crossplane.

This pattern provides:
- **True Data Isolation**: Each tenant gets a dedicated database instance
- **Compliance**: Meets regulatory requirements for data separation
- **Performance**: No noisy neighbor problems
- **Scalability**: Independent database sizing per tenant
- **Automated Provisioning**: Cloud resources managed as Kubernetes objects

## Prerequisites

This pattern assumes **Crossplane** is already installed in your cluster.

::: tip
For Crossplane installation, see [Crossplane Integration](/integration-crossplane) documentation.
:::

### Install AWS Provider

```bash
# Install AWS provider
kubectl apply -f - <<EOF
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-aws
spec:
  package: xpkg.upbound.io/upbound/provider-aws:v0.40.0
EOF

# Configure AWS credentials
kubectl create secret generic aws-creds -n crossplane-system \
  --from-file=credentials=$HOME/.aws/credentials

# Create ProviderConfig
kubectl apply -f - <<EOF
apiVersion: aws.upbound.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: aws-creds
      key: credentials
EOF
```

## Database Schema

```sql
CREATE TABLE tenants (
  tenant_id VARCHAR(63) PRIMARY KEY,
  domain VARCHAR(255) NOT NULL,
  is_active BOOLEAN DEFAULT TRUE,

  -- Database provisioning
  db_type VARCHAR(20) DEFAULT 'postgres',           -- postgres, mysql
  db_instance_class VARCHAR(30) DEFAULT 'db.t3.micro',
  db_storage_gb INT DEFAULT 20,
  db_multi_az BOOLEAN DEFAULT FALSE,

  -- RDS identifier (populated after provisioning)
  rds_instance_id VARCHAR(255),
  rds_endpoint VARCHAR(255),

  plan_type VARCHAR(20) DEFAULT 'basic'             -- basic, pro, enterprise
);
```

## TenantRegistry

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantRegistry
metadata:
  name: database-per-tenant
  namespace: tenant-operator-system
spec:
  source:
    type: mysql
    syncInterval: 1m
    mysql:
      host: mysql.database.svc.cluster.local
      port: 3306
      database: tenants_db
      username: tenant_reader
      passwordRef:
        name: mysql-credentials
        key: password
      table: tenants

  valueMappings:
    uid: tenant_id
    hostOrUrl: domain
    activate: is_active

  extraValueMappings:
    dbInstanceClass: db_instance_class
    dbStorageGb: db_storage_gb
    dbMultiAz: db_multi_az
    planType: plan_type
```

## TenantTemplate with Crossplane Resources

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: database-provisioning
  namespace: tenant-operator-system
spec:
  registryId: database-per-tenant

  # Create tenant namespace
  manifests:
    - id: tenant-namespace
      spec:
        apiVersion: v1
        kind: Namespace
        metadata:
          name: "tenant-{{ .uid }}"
          labels:
            tenant-id: "{{ .uid }}"

  # Provision RDS instance via Crossplane
  manifests:
    - id: rds-instance
      nameTemplate: "{{ .uid }}-postgres"
      targetNamespace: "tenant-{{ .uid }}"
      dependIds: ["tenant-namespace"]
      creationPolicy: Once  # Create database only once
      deletionPolicy: Retain  # Retain database even if tenant deleted (backup first!)
      waitForReady: true
      timeoutSeconds: 1800  # RDS can take 15-30 minutes
      spec:
        apiVersion: database.aws.crossplane.io/v1beta1
        kind: RDSInstance
        metadata:
          name: "{{ .uid | trunc 40 }}-db"  # RDS naming limits
          labels:
            tenant-id: "{{ .uid }}"
        spec:
          forProvider:
            region: us-west-2
            dbInstanceClass: "{{ .dbInstanceClass }}"
            engine: postgres
            engineVersion: "15.3"
            masterUsername: "{{ .uid }}"
            allocatedStorage: {{ .dbStorageGb }}
            storageType: gp3
            storageEncrypted: true
            multiAZ: {{ .dbMultiAz }}
            publiclyAccessible: false
            vpcSecurityGroupIds:
              - sg-0123456789abcdef0  # Your VPC security group
            dbSubnetGroupName: tenant-db-subnet-group
            skipFinalSnapshot: false
            finalDBSnapshotIdentifier: "{{ .uid }}-final-snapshot-{{ now | date \"20060102150405\" }}"
            tags:
              - key: tenant-id
                value: "{{ .uid }}"
              - key: plan-type
                value: "{{ .planType }}"
              - key: managed-by
                value: tenant-operator
          writeConnectionSecretToRef:
            name: "{{ .uid }}-db-conn"
            namespace: "tenant-{{ .uid }}"
          providerConfigRef:
            name: aws-provider-config

  # Application deployment (waits for database)
  deployments:
    - id: app
      nameTemplate: "{{ .uid }}-app"
      targetNamespace: "tenant-{{ .uid }}"
      dependIds: ["rds-instance"]
      waitForReady: true
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
                image: registry.example.com/tenant-app:v1.0.0
                env:
                  - name: TENANT_ID
                    value: "{{ .uid }}"
                  # Crossplane automatically creates connection secret
                  - name: DATABASE_HOST
                    valueFrom:
                      secretKeyRef:
                        name: "{{ .uid }}-db-conn"
                        key: endpoint
                  - name: DATABASE_PORT
                    valueFrom:
                      secretKeyRef:
                        name: "{{ .uid }}-db-conn"
                        key: port
                  - name: DATABASE_NAME
                    value: "{{ .uid }}"
                  - name: DATABASE_USER
                    valueFrom:
                      secretKeyRef:
                        name: "{{ .uid }}-db-conn"
                        key: username
                  - name: DATABASE_PASSWORD
                    valueFrom:
                      secretKeyRef:
                        name: "{{ .uid }}-db-conn"
                        key: password
                ports:
                  - containerPort: 8080
                resources:
                  requests:
                    cpu: "{{ if eq .planType \"enterprise\" }}1000m{{ else }}500m{{ end }}"
                    memory: "{{ if eq .planType \"enterprise\" }}2Gi{{ else }}1Gi{{ end }}"
```

::: tip Connection Secret
Crossplane automatically creates a Secret with connection details (endpoint, port, username, password) that your application can consume.
:::

## Database Connection Secret

Crossplane automatically creates a Secret with connection details:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: acme-corp-db-conn
  namespace: tenant-acme-corp
type: connection.crossplane.io/v1alpha1
data:
  endpoint: <base64-encoded-rds-endpoint>
  port: NTQzMg==  # 5432
  username: <base64-encoded-username>
  password: <base64-encoded-password>
```

## Monitoring Provisioning

```bash
# Check RDS provisioning status
kubectl get rdsinstance -l tenant-id=acme-corp

# Expected output:
# NAME              READY   SYNCED   EXTERNAL-NAME                    AGE
# acme-corp-db      True    True     acme-corp-db-20231105123456      25m

# Check connection secret
kubectl get secret acme-corp-db-conn -n tenant-acme-corp -o yaml

# Verify application can connect
kubectl logs -n tenant-acme-corp deployment/acme-corp-app
```

## Cost Optimization

Define tiered database offerings:

```yaml
# Different instance classes per plan
db.t3.micro   # Basic plan: $15/month
db.t3.small   # Pro plan: $30/month
db.m5.large   # Enterprise plan: $150/month
```

Use database views to filter tenants by plan:

```sql
CREATE VIEW enterprise_tenants AS
SELECT * FROM tenants WHERE plan_type = 'enterprise' AND is_active = TRUE;
```

Then create separate registries and templates per plan tier.

## Benefits

1. **True Isolation**: Each tenant gets dedicated database instance
2. **Cloud-Native**: Leverage managed database services (RDS, Cloud SQL)
3. **Automatic Credentials**: Crossplane manages connection secrets
4. **Declarative**: Database provisioning as code
5. **Retention Policy**: Keep data even after tenant deletion

## Limitations

1. **Cost**: More expensive than shared database
2. **Provisioning Time**: RDS takes 15-30 minutes to provision
3. **Management Overhead**: More databases to backup and maintain
4. **Resource Limits**: AWS account limits on RDS instances

## Related Documentation

- [Crossplane Integration](/integration-crossplane) - Detailed Crossplane setup
- [Policies](/policies) - CreationPolicy and DeletionPolicy for databases
- [Advanced Use Cases](/advanced-use-cases) - Other infrastructure patterns

## Next Steps

- Set up Crossplane provider
- Configure VPC and security groups
- Implement backup strategy
- Set up cost monitoring
