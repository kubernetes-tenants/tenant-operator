# ExternalDNS Integration Guide

This guide shows how to integrate Tenant Operator with ExternalDNS for automatic DNS record management.

## Overview

**ExternalDNS** synchronizes exposed Kubernetes Services and Ingresses with DNS providers like AWS Route53, Google Cloud DNS, Cloudflare, and more. When integrated with Tenant Operator, each tenant's DNS records are automatically created and deleted as tenants are provisioned.

### Use Cases

- **Multi-tenant SaaS**: Automatic subdomain creation per tenant (e.g., `tenant-a.example.com`, `tenant-b.example.com`)
- **Dynamic environments**: DNS records follow tenant lifecycle (created/deleted with tenant)
- **Multiple domains**: Different tenants on different domains or subdomains
- **SSL/TLS automation**: Combined with cert-manager for automatic certificate provisioning

---

## Prerequisites

- Kubernetes cluster v1.11+
- Tenant Operator installed
- DNS provider account (AWS Route53, Cloudflare, etc.)
- DNS zone created in your provider

---

## Installation

### 1. Install ExternalDNS

#### AWS Route53

```bash
# Create IAM policy for ExternalDNS
cat <<EOF > external-dns-policy.json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "route53:ChangeResourceRecordSets"
      ],
      "Resource": [
        "arn:aws:route53:::hostedzone/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "route53:ListHostedZones",
        "route53:ListResourceRecordSets"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF

aws iam create-policy --policy-name ExternalDNSPolicy --policy-document file://external-dns-policy.json

# Create service account with IRSA (IAM Roles for Service Accounts)
eksctl create iamserviceaccount \
  --name external-dns \
  --namespace kube-system \
  --cluster my-cluster \
  --attach-policy-arn arn:aws:iam::ACCOUNT_ID:policy/ExternalDNSPolicy \
  --approve
```

**Deploy ExternalDNS:**

```yaml
# external-dns-route53.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: external-dns
  namespace: kube-system
  # annotations:
  #   eks.amazonaws.com/role-arn: arn:aws:iam::ACCOUNT_ID:role/external-dns
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: external-dns
rules:
- apiGroups: [""]
  resources: ["services", "endpoints", "pods"]
  verbs: ["get", "watch", "list"]
- apiGroups: ["extensions", "networking.k8s.io"]
  resources: ["ingresses"]
  verbs: ["get", "watch", "list"]
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: external-dns-viewer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: external-dns
subjects:
- kind: ServiceAccount
  name: external-dns
  namespace: kube-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: external-dns
  namespace: kube-system
spec:
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: external-dns
  template:
    metadata:
      labels:
        app: external-dns
    spec:
      serviceAccountName: external-dns
      containers:
      - name: external-dns
        image: registry.k8s.io/external-dns/external-dns:v0.14.0
        args:
        - --source=service
        - --source=ingress
        - --domain-filter=example.com  # Limit to specific domain
        - --provider=aws
        - --policy=upsert-only  # Prevent ExternalDNS from deleting existing records
        - --aws-zone-type=public
        - --registry=txt
        - --txt-owner-id=my-cluster-id
        - --log-level=info
        resources:
          limits:
            memory: 50Mi
          requests:
            memory: 50Mi
            cpu: 10m
```

```bash
kubectl apply -f external-dns-route53.yaml
```

#### Cloudflare

```yaml
# external-dns-cloudflare.yaml
apiVersion: v1
kind: Secret
metadata:
  name: cloudflare-api-token
  namespace: kube-system
type: Opaque
stringData:
  apiToken: "YOUR_CLOUDFLARE_API_TOKEN"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: external-dns
  namespace: kube-system
spec:
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: external-dns
  template:
    metadata:
      labels:
        app: external-dns
    spec:
      serviceAccountName: external-dns
      containers:
      - name: external-dns
        image: registry.k8s.io/external-dns/external-dns:v0.14.0
        args:
        - --source=service
        - --source=ingress
        - --domain-filter=example.com
        - --provider=cloudflare
        - --cloudflare-proxied  # Enable Cloudflare proxy
        env:
        - name: CF_API_TOKEN
          valueFrom:
            secretKeyRef:
              name: cloudflare-api-token
              key: apiToken
```

### 2. Verify ExternalDNS Installation

```bash
# Check ExternalDNS pod
kubectl get pods -n kube-system -l app=external-dns

# Check logs
kubectl logs -n kube-system -l app=external-dns
```

---

## Integration with Tenant Operator

### Basic Example: Ingress with ExternalDNS

**TenantTemplate with ExternalDNS annotations:**

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: web-app-with-dns
  namespace: default
spec:
  registryId: my-registry

  # Ingress with ExternalDNS annotation
  ingresses:
  - id: web-ingress
    nameTemplate: "{{ .uid }}-ingress"
    spec:
      apiVersion: networking.k8s.io/v1
      kind: Ingress
      metadata:
        annotations:
          # ExternalDNS will create DNS record automatically
          external-dns.alpha.kubernetes.io/hostname: "{{ .host }}"
          # Optional: Set TTL
          external-dns.alpha.kubernetes.io/ttl: "300"
          # Optional: Cloudflare proxy
          # external-dns.alpha.kubernetes.io/cloudflare-proxied: "true"
      spec:
        ingressClassName: nginx
        rules:
        - host: "{{ .host }}"
          http:
            paths:
            - path: /
              pathType: Prefix
              backend:
                service:
                  name: "{{ .uid }}-service"
                  port:
                    number: 80
        tls:
        - hosts:
          - "{{ .host }}"
          secretName: "{{ .uid }}-tls"

  # Service
  services:
  - id: web-service
    nameTemplate: "{{ .uid }}-service"
    spec:
      apiVersion: v1
      kind: Service
      spec:
        selector:
          app: "{{ .uid }}"
        ports:
        - port: 80
          targetPort: 8080

  # Deployment
  deployments:
  - id: web-deploy
    nameTemplate: "{{ .uid }}-deploy"
    dependIds: ["web-service"]
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
              image: nginx:stable
              ports:
              - containerPort: 8080
              env:
              - name: TENANT_ID
                value: "{{ .uid }}"
              - name: TENANT_HOST
                value: "{{ .host }}"
```

### Advanced Example: LoadBalancer Service with ExternalDNS

For direct LoadBalancer services (without Ingress):

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: api-service-with-dns
  namespace: default
spec:
  registryId: my-registry

  services:
  - id: api-service
    nameTemplate: "{{ .uid }}-api-service"
    spec:
      apiVersion: v1
      kind: Service
      metadata:
        annotations:
          # ExternalDNS will use the LoadBalancer IP/hostname
          external-dns.alpha.kubernetes.io/hostname: "api-{{ .uid }}.example.com"
          external-dns.alpha.kubernetes.io/ttl: "60"
      spec:
        type: LoadBalancer
        selector:
          app: "{{ .uid }}-api"
        ports:
        - port: 443
          targetPort: 8443
          protocol: TCP
```

### Multi-Domain Example

Different tenants on different domains:

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantRegistry
metadata:
  name: multi-domain-registry
  namespace: default
spec:
  source:
    type: mysql
    syncInterval: 1m
    mysql:
      host: mysql.default.svc.cluster.local
      port: 3306
      database: tenants
      table: tenant_configs
      usernameRef:
        name: mysql-credentials
        key: username
      passwordRef:
        name: mysql-credentials
        key: password
  valueMappings:
    uid: tenant_id
    hostOrUrl: tenant_domain  # Full domain per tenant
    activate: is_active
  extraValueMappings:
    subdomain: tenant_subdomain
    rootDomain: root_domain
---
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: multi-domain-template
  namespace: default
spec:
  registryId: multi-domain-registry

  ingresses:
  - id: tenant-ingress
    nameTemplate: "{{ .uid }}-ingress"
    spec:
      apiVersion: networking.k8s.io/v1
      kind: Ingress
      metadata:
        annotations:
          # Use custom domain from database
          external-dns.alpha.kubernetes.io/hostname: "{{ .subdomain }}.{{ .rootDomain }}"
          cert-manager.io/cluster-issuer: "letsencrypt-prod"
      spec:
        ingressClassName: nginx
        rules:
        - host: "{{ .subdomain }}.{{ .rootDomain }}"
          http:
            paths:
            - path: /
              pathType: Prefix
              backend:
                service:
                  name: "{{ .uid }}-service"
                  port:
                    number: 80
        tls:
        - hosts:
          - "{{ .subdomain }}.{{ .rootDomain }}"
          secretName: "{{ .uid }}-tls"
```

**Database schema example:**

```sql
CREATE TABLE tenant_configs (
  tenant_id VARCHAR(255) PRIMARY KEY,
  tenant_subdomain VARCHAR(255),
  root_domain VARCHAR(255),
  tenant_domain VARCHAR(255) GENERATED ALWAYS AS (CONCAT(tenant_subdomain, '.', root_domain)) STORED,
  is_active VARCHAR(10)
);

INSERT INTO tenant_configs (tenant_id, tenant_subdomain, root_domain, is_active) VALUES
('tenant-alpha', 'alpha', 'saas.example.com', '1'),
('tenant-beta', 'beta', 'saas.example.com', '1'),
('tenant-gamma', 'gamma', 'enterprise.example.io', '1');
```

---

## How It Works

### Workflow

1. **Tenant Created**: TenantRegistry controller creates Tenant CR from database row
2. **Resources Applied**: Tenant controller creates Ingress/Service with ExternalDNS annotations
3. **DNS Record Created**: ExternalDNS detects new Ingress/Service and creates DNS record in provider
4. **Traffic Routed**: DNS resolves to LoadBalancer/Ingress IP
5. **Tenant Deleted**: When tenant is deactivated, Ingress/Service is deleted
6. **DNS Record Deleted**: ExternalDNS removes DNS record automatically

### DNS Record Lifecycle

```
Database: activate=1
    ↓
Tenant CR Created
    ↓
Ingress Applied (with annotation)
    ↓
ExternalDNS detects Ingress
    ↓
DNS Record Created (A/CNAME)
    ↓
Traffic flows to tenant

---

Database: activate=0
    ↓
Tenant CR Deleted (via finalizer)
    ↓
Ingress Deleted
    ↓
ExternalDNS detects deletion
    ↓
DNS Record Removed
```

---

## ExternalDNS Annotations Reference

### Common Annotations

| Annotation | Description | Example |
|------------|-------------|---------|
| `external-dns.alpha.kubernetes.io/hostname` | DNS hostname to create | `tenant-a.example.com` |
| `external-dns.alpha.kubernetes.io/ttl` | DNS record TTL | `"300"` (seconds) |
| `external-dns.alpha.kubernetes.io/target` | Override target IP/hostname | `"custom-lb.example.com"` |
| `external-dns.alpha.kubernetes.io/alias` | AWS Route53 alias record | `"true"` |

### Provider-Specific Annotations

#### AWS Route53

```yaml
metadata:
  annotations:
    external-dns.alpha.kubernetes.io/hostname: "tenant.example.com"
    external-dns.alpha.kubernetes.io/alias: "true"  # Use Route53 Alias
    external-dns.alpha.kubernetes.io/set-identifier: "tenant-a"  # For weighted routing
    external-dns.alpha.kubernetes.io/aws-weight: "100"
```

#### Cloudflare

```yaml
metadata:
  annotations:
    external-dns.alpha.kubernetes.io/hostname: "tenant.example.com"
    external-dns.alpha.kubernetes.io/cloudflare-proxied: "true"  # Enable CF proxy
```

#### Google Cloud DNS

```yaml
metadata:
  annotations:
    external-dns.alpha.kubernetes.io/hostname: "tenant.example.com"
    external-dns.alpha.kubernetes.io/ttl: "300"
```

---

## Complete Example: SaaS Application

Full example combining Tenant Operator, ExternalDNS, cert-manager, and NGINX Ingress:

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: saas-platform
  namespace: default
spec:
  registryId: saas-registry

  # Namespace for tenant isolation
  namespaces:
  - id: tenant-ns
    nameTemplate: "tenant-{{ .uid }}"
    spec:
      apiVersion: v1
      kind: Namespace

  # ConfigMap with tenant config
  configMaps:
  - id: tenant-config
    nameTemplate: "{{ .uid }}-config"
    namespaceTemplate: "tenant-{{ .uid }}"
    dependIds: ["tenant-ns"]
    spec:
      apiVersion: v1
      kind: ConfigMap
      data:
        tenant_id: "{{ .uid }}"
        tenant_host: "{{ .host }}"
        api_endpoint: "https://{{ .host }}/api"

  # Application deployment
  deployments:
  - id: app-deploy
    nameTemplate: "{{ .uid }}-app"
    namespaceTemplate: "tenant-{{ .uid }}"
    dependIds: ["tenant-ns", "tenant-config"]
    spec:
      apiVersion: apps/v1
      kind: Deployment
      spec:
        replicas: 2
        selector:
          matchLabels:
            app: "{{ .uid }}"
            tier: frontend
        template:
          metadata:
            labels:
              app: "{{ .uid }}"
              tier: frontend
          spec:
            containers:
            - name: app
              image: mycompany/saas-app:v1.2.3
              ports:
              - containerPort: 8080
              env:
              - name: TENANT_ID
                value: "{{ .uid }}"
              - name: TENANT_HOST
                value: "{{ .host }}"
              envFrom:
              - configMapRef:
                  name: "{{ .uid }}-config"
              livenessProbe:
                httpGet:
                  path: /healthz
                  port: 8080
                initialDelaySeconds: 30
                periodSeconds: 10
              readinessProbe:
                httpGet:
                  path: /ready
                  port: 8080
                initialDelaySeconds: 5
                periodSeconds: 5
              resources:
                requests:
                  cpu: 100m
                  memory: 128Mi
                limits:
                  cpu: 500m
                  memory: 512Mi

  # Service
  services:
  - id: app-service
    nameTemplate: "{{ .uid }}-service"
    namespaceTemplate: "tenant-{{ .uid }}"
    dependIds: ["tenant-ns"]
    spec:
      apiVersion: v1
      kind: Service
      spec:
        type: ClusterIP
        selector:
          app: "{{ .uid }}"
          tier: frontend
        ports:
        - port: 80
          targetPort: 8080
          protocol: TCP

  # Ingress with ExternalDNS + cert-manager
  ingresses:
  - id: app-ingress
    nameTemplate: "{{ .uid }}-ingress"
    namespaceTemplate: "tenant-{{ .uid }}"
    dependIds: ["tenant-ns", "app-service"]
    waitForReady: false  # Don't wait for Ingress (DNS propagation takes time)
    spec:
      apiVersion: networking.k8s.io/v1
      kind: Ingress
      metadata:
        annotations:
          # NGINX Ingress
          nginx.ingress.kubernetes.io/ssl-redirect: "true"
          nginx.ingress.kubernetes.io/force-ssl-redirect: "true"

          # ExternalDNS
          external-dns.alpha.kubernetes.io/hostname: "{{ .host }}"
          external-dns.alpha.kubernetes.io/ttl: "300"

          # cert-manager for automatic TLS
          cert-manager.io/cluster-issuer: "letsencrypt-prod"
          cert-manager.io/acme-challenge-type: "http01"
      spec:
        ingressClassName: nginx
        rules:
        - host: "{{ .host }}"
          http:
            paths:
            - path: /
              pathType: Prefix
              backend:
                service:
                  name: "{{ .uid }}-service"
                  port:
                    number: 80
        tls:
        - hosts:
          - "{{ .host }}"
          secretName: "{{ .uid }}-tls"  # cert-manager will populate this
```

**Deploy:**

```bash
# Apply template
kubectl apply -f saas-platform-template.yaml

# Wait for tenants to be created
kubectl get tenant -n default -w

# Check DNS records (example with Route53)
aws route53 list-resource-record-sets \
  --hosted-zone-id Z1234567890ABC \
  --query "ResourceRecordSets[?contains(Name, 'tenant')]"

# Test DNS resolution
nslookup tenant-alpha.saas.example.com
nslookup tenant-beta.saas.example.com

# Test HTTPS
curl https://tenant-alpha.saas.example.com
curl https://tenant-beta.saas.example.com
```

---

## Troubleshooting

### DNS Records Not Created

**Problem:** ExternalDNS not creating DNS records.

**Solution:**

1. **Check ExternalDNS logs:**
   ```bash
   kubectl logs -n kube-system -l app=external-dns --tail=100
   ```

2. **Verify annotation exists:**
   ```bash
   kubectl get ingress -n tenant-alpha tenant-alpha-ingress -o yaml | grep external-dns
   ```

3. **Check ExternalDNS permissions:**
   ```bash
   # AWS: Verify IAM role
   kubectl describe sa -n kube-system external-dns

   # Check if role has Route53 permissions
   aws iam get-role-policy --role-name external-dns-role --policy-name external-dns-policy
   ```

4. **Verify domain filter:**
   ```bash
   kubectl get deployment -n kube-system external-dns -o yaml | grep domain-filter
   ```

### DNS Records Not Deleted

**Problem:** DNS records remain after tenant deletion.

**Solution:**

1. **Check ExternalDNS policy:**
   ```bash
   kubectl get deployment -n kube-system external-dns -o yaml | grep policy
   ```

   Should be `--policy=sync` or `--policy=upsert-only`

2. **Check TXT records:**
   ExternalDNS uses TXT records for ownership. Verify:
   ```bash
   aws route53 list-resource-record-sets --hosted-zone-id Z123 | grep TXT
   ```

3. **Manually delete if needed:**
   ```bash
   aws route53 change-resource-record-sets \
     --hosted-zone-id Z1234567890ABC \
     --change-batch file://delete-record.json
   ```

### DNS Propagation Delays

**Problem:** DNS records created but not resolving.

**Solution:**

1. **Check DNS propagation:**
   ```bash
   # Query authoritative nameserver directly
   dig @ns-1234.awsdns-12.org tenant-alpha.example.com

   # Check from multiple locations
   dig +trace tenant-alpha.example.com
   ```

2. **Reduce TTL** for faster updates:
   ```yaml
   annotations:
     external-dns.alpha.kubernetes.io/ttl: "60"  # 1 minute
   ```

3. **Wait for propagation** (typically 1-5 minutes for low TTL)

### Ingress Not Getting IP

**Problem:** Ingress created but has no IP/hostname.

**Solution:**

1. **Check Ingress controller:**
   ```bash
   kubectl get pods -n ingress-nginx
   kubectl get svc -n ingress-nginx
   ```

2. **Check Ingress status:**
   ```bash
   kubectl describe ingress -n tenant-alpha tenant-alpha-ingress
   ```

3. **Verify IngressClass:**
   ```bash
   kubectl get ingressclass
   ```

---

## Best Practices

### 1. Use Separate Hosted Zones

For multi-tenant SaaS, create separate hosted zones per environment:

```
- production.example.com (Zone ID: Z111)
  - tenant-a.production.example.com
  - tenant-b.production.example.com

- staging.example.com (Zone ID: Z222)
  - tenant-a.staging.example.com
  - tenant-b.staging.example.com
```

### 2. Set Appropriate TTLs

- **Development**: TTL 60s (fast iteration)
- **Staging**: TTL 300s (5 minutes)
- **Production**: TTL 3600s (1 hour)

### 3. Use DNS Policy

```yaml
# Recommended: Sync mode for full control
args:
- --policy=sync

# Alternative: Upsert-only (safer, but won't delete)
args:
- --policy=upsert-only
```

### 4. Monitor DNS Changes

Set up CloudWatch alarms or monitoring:

```bash
# AWS CloudWatch alarm for Route53 changes
aws cloudwatch put-metric-alarm \
  --alarm-name route53-high-change-rate \
  --metric-name ChangeCount \
  --namespace AWS/Route53 \
  --statistic Sum \
  --period 300 \
  --evaluation-periods 1 \
  --threshold 100 \
  --comparison-operator GreaterThanThreshold
```

### 5. Use TXT Record Registry

Always enable TXT registry for ownership tracking:

```yaml
args:
- --registry=txt
- --txt-owner-id=my-cluster-id
- --txt-prefix=external-dns-
```

---

## See Also

- [ExternalDNS Documentation](https://github.com/kubernetes-sigs/external-dns)
- [cert-manager Integration](https://cert-manager.io/docs/)
- [Tenant Operator Templates Guide](templates.md)
- [Multi-Tenant Architecture](../README.md#architecture)
