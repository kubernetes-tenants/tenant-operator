import{_ as n,c as a,o as l,a2 as p}from"./chunks/framework.9Uv4PgnO.js";const E=JSON.parse('{"title":"Terraform Operator Integration Guide","description":"","frontmatter":{},"headers":[],"relativePath":"integration-terraform-operator.md","filePath":"integration-terraform-operator.md","lastUpdated":1761887194000}'),e={name:"integration-terraform-operator.md"};function r(o,s,c,t,i,u){return l(),a("div",null,[...s[0]||(s[0]=[p(`<h1 id="terraform-operator-integration-guide" tabindex="-1">Terraform Operator Integration Guide <a class="header-anchor" href="#terraform-operator-integration-guide" aria-label="Permalink to &quot;Terraform Operator Integration Guide&quot;">​</a></h1><p>This guide shows how to integrate Tenant Operator with Terraform Operator for provisioning external cloud resources (AWS, GCP, Azure) per tenant.</p><nav class="table-of-contents"><ul><li><a href="#overview">Overview</a><ul><li><a href="#key-benefits">Key Benefits</a></li><li><a href="#use-cases">Use Cases</a></li></ul></li><li><a href="#prerequisites">Prerequisites</a></li><li><a href="#installation">Installation</a><ul><li><a href="#_1-install-tofu-controller">1. Install Tofu Controller</a></li><li><a href="#_2-create-cloud-provider-credentials">2. Create Cloud Provider Credentials</a></li><li><a href="#_3-verify-installation">3. Verify Installation</a></li></ul></li><li><a href="#basic-integration">Basic Integration</a><ul><li><a href="#example-1-s3-bucket-per-tenant">Example 1: S3 Bucket per Tenant</a></li></ul></li><li><a href="#advanced-examples">Advanced Examples</a><ul><li><a href="#example-2-rds-postgresql-database-per-tenant">Example 2: RDS PostgreSQL Database per Tenant</a></li><li><a href="#example-3-cloudfront-cdn-distribution">Example 3: CloudFront CDN Distribution</a></li><li><a href="#example-4-using-git-repository-for-terraform-modules">Example 4: Using Git Repository for Terraform Modules</a></li><li><a href="#example-5-kafka-topics-and-acls-per-tenant">Example 5: Kafka Topics and ACLs per Tenant</a></li><li><a href="#example-6-rabbitmq-virtual-host-and-user-per-tenant">Example 6: RabbitMQ Virtual Host and User per Tenant</a></li><li><a href="#example-7-postgresql-schema-and-user-per-tenant">Example 7: PostgreSQL Schema and User per Tenant</a></li><li><a href="#example-8-redis-database-per-tenant">Example 8: Redis Database per Tenant</a></li></ul></li><li><a href="#complete-multi-resource-example">Complete Multi-Resource Example</a></li><li><a href="#how-it-works">How It Works</a><ul><li><a href="#workflow">Workflow</a></li><li><a href="#state-management">State Management</a></li></ul></li><li><a href="#best-practices">Best Practices</a><ul><li><a href="#_1-use-creationpolicy-once-for-immutable-infrastructure">1. Use CreationPolicy: Once for Immutable Infrastructure</a></li><li><a href="#_2-set-appropriate-timeouts">2. Set Appropriate Timeouts</a></li><li><a href="#_3-use-remote-state-backend-production">3. Use Remote State Backend (Production)</a></li><li><a href="#_4-secure-sensitive-outputs">4. Secure Sensitive Outputs</a></li><li><a href="#_5-use-dependency-ordering">5. Use Dependency Ordering</a></li><li><a href="#_6-monitor-terraform-resources">6. Monitor Terraform Resources</a></li></ul></li><li><a href="#troubleshooting">Troubleshooting</a><ul><li><a href="#terraform-apply-fails">Terraform Apply Fails</a></li><li><a href="#state-lock-issues">State Lock Issues</a></li><li><a href="#outputs-not-available">Outputs Not Available</a></li><li><a href="#resource-already-exists">Resource Already Exists</a></li></ul></li><li><a href="#cost-optimization">Cost Optimization</a><ul><li><a href="#_1-use-appropriate-instance-sizes">1. Use Appropriate Instance Sizes</a></li><li><a href="#_2-enable-auto-scaling">2. Enable Auto-Scaling</a></li><li><a href="#_3-use-lifecycle-policies">3. Use Lifecycle Policies</a></li></ul></li><li><a href="#see-also">See Also</a></li></ul></nav><h2 id="overview" tabindex="-1">Overview <a class="header-anchor" href="#overview" aria-label="Permalink to &quot;Overview&quot;">​</a></h2><p><strong>Terraform Operator</strong> allows you to manage Terraform resources as Kubernetes Custom Resources. When integrated with Tenant Operator, each tenant can automatically provision <strong>any infrastructure resource</strong> that Terraform supports - from cloud services to on-premises systems.</p><h3 id="key-benefits" tabindex="-1">Key Benefits <a class="header-anchor" href="#key-benefits" aria-label="Permalink to &quot;Key Benefits&quot;">​</a></h3><p><strong>Universal Resource Provisioning</strong>: Terraform supports 3,000+ providers, enabling you to provision virtually any infrastructure:</p><ul><li>☁️ <strong>Cloud Resources</strong>: AWS, GCP, Azure, DigitalOcean, Alibaba Cloud</li><li>📦 <strong>Databases</strong>: PostgreSQL, MySQL, MongoDB, Cassandra, DynamoDB</li><li>📬 <strong>Messaging Systems</strong>: Kafka, RabbitMQ, Pulsar, ActiveMQ, AWS SQS/SNS</li><li>🔍 <strong>Search &amp; Analytics</strong>: Elasticsearch, OpenSearch, Splunk</li><li>🗄️ <strong>Caching</strong>: Redis, Memcached, AWS ElastiCache</li><li>🌐 <strong>DNS &amp; CDN</strong>: Route53, Cloudflare, Akamai, Fastly</li><li>🔐 <strong>Security</strong>: Vault, Auth0, Keycloak, AWS IAM</li><li>📊 <strong>Monitoring</strong>: Datadog, New Relic, PagerDuty</li><li>🏢 <strong>On-Premises</strong>: VMware vSphere, Proxmox, Bare Metal</li></ul><p><strong>Automatic Lifecycle Management</strong>:</p><ul><li>✅ <strong>Provisioning</strong>: Resources created when tenant is activated (<code>activate=1</code>)</li><li>🔄 <strong>Drift Detection</strong>: Terraform ensures desired state matches actual state</li><li>🗑️ <strong>Cleanup</strong>: Resources automatically destroyed when tenant is deleted</li><li>📦 <strong>Consistent State</strong>: All tenant infrastructure managed declaratively</li></ul><h3 id="use-cases" tabindex="-1">Use Cases <a class="header-anchor" href="#use-cases" aria-label="Permalink to &quot;Use Cases&quot;">​</a></h3><h4 id="cloud-services-aws-gcp-azure" tabindex="-1">Cloud Services (AWS, GCP, Azure) <a class="header-anchor" href="#cloud-services-aws-gcp-azure" aria-label="Permalink to &quot;Cloud Services (AWS, GCP, Azure)&quot;">​</a></h4><ul><li><strong>S3/GCS/Blob Storage</strong>: Isolated storage per tenant</li><li><strong>RDS/Cloud SQL</strong>: Dedicated databases per tenant</li><li><strong>CloudFront/Cloud CDN</strong>: Tenant-specific CDN distributions</li><li><strong>IAM Roles/Policies</strong>: Tenant-specific access control</li><li><strong>VPCs/Subnets</strong>: Network isolation</li><li><strong>ElastiCache/Memorystore</strong>: Per-tenant caching layers</li><li><strong>Lambda/Cloud Functions</strong>: Serverless functions per tenant</li></ul><h4 id="messaging-streaming" tabindex="-1">Messaging &amp; Streaming <a class="header-anchor" href="#messaging-streaming" aria-label="Permalink to &quot;Messaging &amp; Streaming&quot;">​</a></h4><ul><li><strong>Kafka Topics</strong>: Dedicated topics and ACLs per tenant</li><li><strong>RabbitMQ VHosts</strong>: Virtual hosts and users per tenant</li><li><strong>AWS SQS/SNS</strong>: Queue and topic isolation</li><li><strong>Pulsar Namespaces</strong>: Tenant-isolated messaging</li><li><strong>NATS Accounts</strong>: Multi-tenant streaming</li></ul><h4 id="databases-self-managed-managed" tabindex="-1">Databases (Self-Managed &amp; Managed) <a class="header-anchor" href="#databases-self-managed-managed" aria-label="Permalink to &quot;Databases (Self-Managed &amp; Managed)&quot;">​</a></h4><ul><li><strong>PostgreSQL Schemas</strong>: Isolated schemas in shared cluster</li><li><strong>MongoDB Databases</strong>: Dedicated databases with authentication</li><li><strong>Redis Databases</strong>: Separate database indexes per tenant</li><li><strong>Elasticsearch Indices</strong>: Tenant-specific indices with ILM policies</li><li><strong>InfluxDB Organizations</strong>: Time-series data isolation</li></ul><h4 id="on-premises-hybrid" tabindex="-1">On-Premises &amp; Hybrid <a class="header-anchor" href="#on-premises-hybrid" aria-label="Permalink to &quot;On-Premises &amp; Hybrid&quot;">​</a></h4><ul><li><strong>VMware VMs</strong>: Provision VMs per tenant</li><li><strong>Proxmox Containers</strong>: Lightweight tenant isolation</li><li><strong>F5 Load Balancer</strong>: Per-tenant virtual servers</li><li><strong>NetBox IPAM</strong>: IP address allocation per tenant</li></ul><h2 id="prerequisites" tabindex="-1">Prerequisites <a class="header-anchor" href="#prerequisites" aria-label="Permalink to &quot;Prerequisites&quot;">​</a></h2><div class="info custom-block"><p class="custom-block-title">Requirements</p><ul><li>Kubernetes cluster v1.16+</li><li>Tenant Operator installed</li><li>Cloud provider account (AWS, GCP, or Azure)</li><li>Terraform ≥ 1.0</li><li>Cloud provider credentials (stored as Secrets)</li></ul></div><h2 id="installation" tabindex="-1">Installation <a class="header-anchor" href="#installation" aria-label="Permalink to &quot;Installation&quot;">​</a></h2><h3 id="_1-install-tofu-controller" tabindex="-1">1. Install Tofu Controller <a class="header-anchor" href="#_1-install-tofu-controller" aria-label="Permalink to &quot;1. Install Tofu Controller&quot;">​</a></h3><p>We&#39;ll use <strong>tofu-controller</strong> (formerly tf-controller), which is the production-ready Flux controller for managing Terraform/OpenTofu resources.</p><div class="info custom-block"><p class="custom-block-title">Project evolution</p><p>The original Weave tf-controller has evolved into tofu-controller, now maintained by the Flux community: <a href="https://github.com/flux-iac/tofu-controller" target="_blank" rel="noreferrer">https://github.com/flux-iac/tofu-controller</a></p></div><h4 id="installation-via-helm-recommended" tabindex="-1">Installation via Helm (Recommended) <a class="header-anchor" href="#installation-via-helm-recommended" aria-label="Permalink to &quot;Installation via Helm (Recommended)&quot;">​</a></h4><div class="language-bash line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#6A737D;"># Install Flux (required)</span></span>
<span class="line"><span style="color:#B392F0;">flux</span><span style="color:#9ECBFF;"> install</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;"># Add tofu-controller Helm repository</span></span>
<span class="line"><span style="color:#B392F0;">helm</span><span style="color:#9ECBFF;"> repo</span><span style="color:#9ECBFF;"> add</span><span style="color:#9ECBFF;"> tofu-controller</span><span style="color:#9ECBFF;"> https://flux-iac.github.io/tofu-controller</span></span>
<span class="line"><span style="color:#B392F0;">helm</span><span style="color:#9ECBFF;"> repo</span><span style="color:#9ECBFF;"> update</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;"># Install tofu-controller</span></span>
<span class="line"><span style="color:#B392F0;">helm</span><span style="color:#9ECBFF;"> install</span><span style="color:#9ECBFF;"> tofu-controller</span><span style="color:#9ECBFF;"> tofu-controller/tofu-controller</span><span style="color:#79B8FF;"> \\</span></span>
<span class="line"><span style="color:#79B8FF;">  --namespace</span><span style="color:#9ECBFF;"> flux-system</span><span style="color:#79B8FF;"> \\</span></span>
<span class="line"><span style="color:#79B8FF;">  --create-namespace</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br></div></div><h4 id="installation-via-manifests" tabindex="-1">Installation via Manifests <a class="header-anchor" href="#installation-via-manifests" aria-label="Permalink to &quot;Installation via Manifests&quot;">​</a></h4><div class="language-bash line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#6A737D;"># Install Flux</span></span>
<span class="line"><span style="color:#B392F0;">flux</span><span style="color:#9ECBFF;"> install</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;"># Install tofu-controller CRDs and controller</span></span>
<span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> apply</span><span style="color:#79B8FF;"> -f</span><span style="color:#9ECBFF;"> https://raw.githubusercontent.com/flux-iac/tofu-controller/main/config/crd/bases/infra.contrib.fluxcd.io_terraforms.yaml</span></span>
<span class="line"></span>
<span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> apply</span><span style="color:#79B8FF;"> -f</span><span style="color:#9ECBFF;"> https://raw.githubusercontent.com/flux-iac/tofu-controller/main/config/rbac/role.yaml</span></span>
<span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> apply</span><span style="color:#79B8FF;"> -f</span><span style="color:#9ECBFF;"> https://raw.githubusercontent.com/flux-iac/tofu-controller/main/config/rbac/role_binding.yaml</span></span>
<span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> apply</span><span style="color:#79B8FF;"> -f</span><span style="color:#9ECBFF;"> https://raw.githubusercontent.com/flux-iac/tofu-controller/main/config/manager/deployment.yaml</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br></div></div><h4 id="verify-installation" tabindex="-1">Verify Installation <a class="header-anchor" href="#verify-installation" aria-label="Permalink to &quot;Verify Installation&quot;">​</a></h4><div class="language-bash line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#6A737D;"># Check tofu-controller pod</span></span>
<span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> get</span><span style="color:#9ECBFF;"> pods</span><span style="color:#79B8FF;"> -n</span><span style="color:#9ECBFF;"> flux-system</span><span style="color:#79B8FF;"> -l</span><span style="color:#9ECBFF;"> app=tofu-controller</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;"># Check CRD</span></span>
<span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> get</span><span style="color:#9ECBFF;"> crd</span><span style="color:#9ECBFF;"> terraforms.infra.contrib.fluxcd.io</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;"># Check controller logs</span></span>
<span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> logs</span><span style="color:#79B8FF;"> -n</span><span style="color:#9ECBFF;"> flux-system</span><span style="color:#79B8FF;"> -l</span><span style="color:#9ECBFF;"> app=tofu-controller</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br></div></div><h3 id="_2-create-cloud-provider-credentials" tabindex="-1">2. Create Cloud Provider Credentials <a class="header-anchor" href="#_2-create-cloud-provider-credentials" aria-label="Permalink to &quot;2. Create Cloud Provider Credentials&quot;">​</a></h3><h4 id="aws-credentials" tabindex="-1">AWS Credentials <a class="header-anchor" href="#aws-credentials" aria-label="Permalink to &quot;AWS Credentials&quot;">​</a></h4><div class="language-bash line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#6A737D;"># Create AWS credentials secret</span></span>
<span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> create</span><span style="color:#9ECBFF;"> secret</span><span style="color:#9ECBFF;"> generic</span><span style="color:#9ECBFF;"> aws-credentials</span><span style="color:#79B8FF;"> \\</span></span>
<span class="line"><span style="color:#79B8FF;">  --namespace</span><span style="color:#9ECBFF;"> default</span><span style="color:#79B8FF;"> \\</span></span>
<span class="line"><span style="color:#79B8FF;">  --from-literal=AWS_ACCESS_KEY_ID=your-access-key</span><span style="color:#79B8FF;"> \\</span></span>
<span class="line"><span style="color:#79B8FF;">  --from-literal=AWS_SECRET_ACCESS_KEY=your-secret-key</span><span style="color:#79B8FF;"> \\</span></span>
<span class="line"><span style="color:#79B8FF;">  --from-literal=AWS_DEFAULT_REGION=us-east-1</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br></div></div><h4 id="gcp-credentials" tabindex="-1">GCP Credentials <a class="header-anchor" href="#gcp-credentials" aria-label="Permalink to &quot;GCP Credentials&quot;">​</a></h4><div class="language-bash line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#6A737D;"># Create GCP service account key secret</span></span>
<span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> create</span><span style="color:#9ECBFF;"> secret</span><span style="color:#9ECBFF;"> generic</span><span style="color:#9ECBFF;"> gcp-credentials</span><span style="color:#79B8FF;"> \\</span></span>
<span class="line"><span style="color:#79B8FF;">  --namespace</span><span style="color:#9ECBFF;"> default</span><span style="color:#79B8FF;"> \\</span></span>
<span class="line"><span style="color:#79B8FF;">  --from-file=credentials.json=path/to/your-service-account-key.json</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br></div></div><h4 id="azure-credentials" tabindex="-1">Azure Credentials <a class="header-anchor" href="#azure-credentials" aria-label="Permalink to &quot;Azure Credentials&quot;">​</a></h4><div class="language-bash line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#6A737D;"># Create Azure credentials secret</span></span>
<span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> create</span><span style="color:#9ECBFF;"> secret</span><span style="color:#9ECBFF;"> generic</span><span style="color:#9ECBFF;"> azure-credentials</span><span style="color:#79B8FF;"> \\</span></span>
<span class="line"><span style="color:#79B8FF;">  --namespace</span><span style="color:#9ECBFF;"> default</span><span style="color:#79B8FF;"> \\</span></span>
<span class="line"><span style="color:#79B8FF;">  --from-literal=ARM_CLIENT_ID=your-client-id</span><span style="color:#79B8FF;"> \\</span></span>
<span class="line"><span style="color:#79B8FF;">  --from-literal=ARM_CLIENT_SECRET=your-client-secret</span><span style="color:#79B8FF;"> \\</span></span>
<span class="line"><span style="color:#79B8FF;">  --from-literal=ARM_TENANT_ID=your-tenant-id</span><span style="color:#79B8FF;"> \\</span></span>
<span class="line"><span style="color:#79B8FF;">  --from-literal=ARM_SUBSCRIPTION_ID=your-subscription-id</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br></div></div><h3 id="_3-verify-installation" tabindex="-1">3. Verify Installation <a class="header-anchor" href="#_3-verify-installation" aria-label="Permalink to &quot;3. Verify Installation&quot;">​</a></h3><div class="language-bash line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#6A737D;"># Check tf-controller pod</span></span>
<span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> get</span><span style="color:#9ECBFF;"> pods</span><span style="color:#79B8FF;"> -n</span><span style="color:#9ECBFF;"> flux-system</span><span style="color:#79B8FF;"> -l</span><span style="color:#9ECBFF;"> app=tf-controller</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;"># Check CRD</span></span>
<span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> get</span><span style="color:#9ECBFF;"> crd</span><span style="color:#9ECBFF;"> terraforms.infra.contrib.fluxcd.io</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br></div></div><h2 id="basic-integration" tabindex="-1">Basic Integration <a class="header-anchor" href="#basic-integration" aria-label="Permalink to &quot;Basic Integration&quot;">​</a></h2><h3 id="example-1-s3-bucket-per-tenant" tabindex="-1">Example 1: S3 Bucket per Tenant <a class="header-anchor" href="#example-1-s3-bucket-per-tenant" aria-label="Permalink to &quot;Example 1: S3 Bucket per Tenant&quot;">​</a></h3><p><strong>TenantTemplate with Terraform manifest:</strong></p><div class="language-yaml line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">yaml</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#85E89D;">apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">operator.kubernetes-tenants.org/v1</span></span>
<span class="line"><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">TenantTemplate</span></span>
<span class="line"><span style="color:#85E89D;">metadata</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant-with-s3</span></span>
<span class="line"><span style="color:#85E89D;">  namespace</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">default</span></span>
<span class="line"><span style="color:#85E89D;">spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  registryId</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">my-registry</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">  # Terraform resource for S3 bucket</span></span>
<span class="line"><span style="color:#85E89D;">  manifests</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">  - </span><span style="color:#85E89D;">id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">s3-bucket</span></span>
<span class="line"><span style="color:#85E89D;">    nameTemplate</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-s3&quot;</span></span>
<span class="line"><span style="color:#85E89D;">    spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">      apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">infra.contrib.fluxcd.io/v1alpha2</span></span>
<span class="line"><span style="color:#85E89D;">      kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Terraform</span></span>
<span class="line"><span style="color:#85E89D;">      metadata</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">        annotations</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          tenant-operator.kubernetes-tenants.org/tenant-id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#85E89D;">      spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">        interval</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">5m</span></span>
<span class="line"><span style="color:#85E89D;">        retryInterval</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">30s</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">        # Terraform source (inline or from Git)</span></span>
<span class="line"><span style="color:#85E89D;">        sourceRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">GitRepository</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">terraform-modules</span></span>
<span class="line"><span style="color:#85E89D;">          namespace</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">default</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">        # Or use inline Terraform code</span></span>
<span class="line"><span style="color:#85E89D;">        path</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">        # Inline Terraform HCL</span></span>
<span class="line"><span style="color:#85E89D;">        values</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          hcl</span><span style="color:#E1E4E8;">: </span><span style="color:#F97583;">|</span></span>
<span class="line"><span style="color:#9ECBFF;">            terraform {</span></span>
<span class="line"><span style="color:#9ECBFF;">              required_providers {</span></span>
<span class="line"><span style="color:#9ECBFF;">                aws = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  source  = &quot;hashicorp/aws&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                  version = &quot;~&gt; 5.0&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">              backend &quot;kubernetes&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">                secret_suffix = &quot;{{ .uid }}-s3&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                namespace     = &quot;default&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            provider &quot;aws&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              region = var.aws_region</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;tenant_id&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              type = string</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;aws_region&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              type    = string</span></span>
<span class="line"><span style="color:#9ECBFF;">              default = &quot;us-east-1&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;aws_s3_bucket&quot; &quot;tenant_bucket&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              bucket = &quot;tenant-\${var.tenant_id}-bucket&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              tags = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                Name        = &quot;Tenant \${var.tenant_id} Bucket&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                TenantId    = var.tenant_id</span></span>
<span class="line"><span style="color:#9ECBFF;">                ManagedBy   = &quot;tenant-operator&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;aws_s3_bucket_versioning&quot; &quot;tenant_bucket_versioning&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              bucket = aws_s3_bucket.tenant_bucket.id</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              versioning_configuration {</span></span>
<span class="line"><span style="color:#9ECBFF;">                status = &quot;Enabled&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;aws_s3_bucket_server_side_encryption_configuration&quot; &quot;tenant_bucket_encryption&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              bucket = aws_s3_bucket.tenant_bucket.id</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              rule {</span></span>
<span class="line"><span style="color:#9ECBFF;">                apply_server_side_encryption_by_default {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  sse_algorithm = &quot;AES256&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;bucket_name&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              value = aws_s3_bucket.tenant_bucket.id</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;bucket_arn&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              value = aws_s3_bucket.tenant_bucket.arn</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;bucket_region&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              value = aws_s3_bucket.tenant_bucket.region</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">        # Variables passed to Terraform</span></span>
<span class="line"><span style="color:#85E89D;">        vars</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant_id</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">aws_region</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;us-east-1&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">        # Use AWS credentials from secret</span></span>
<span class="line"><span style="color:#85E89D;">        varsFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Secret</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">aws-credentials</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">        # Write Terraform outputs to ConfigMap</span></span>
<span class="line"><span style="color:#85E89D;">        writeOutputsToSecret</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-s3-outputs&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">  # ConfigMap referencing Terraform outputs</span></span>
<span class="line"><span style="color:#85E89D;">  configMaps</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">  - </span><span style="color:#85E89D;">id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">app-config</span></span>
<span class="line"><span style="color:#85E89D;">    nameTemplate</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-config&quot;</span></span>
<span class="line"><span style="color:#85E89D;">    dependIds</span><span style="color:#E1E4E8;">: [</span><span style="color:#9ECBFF;">&quot;s3-bucket&quot;</span><span style="color:#E1E4E8;">]</span></span>
<span class="line"><span style="color:#85E89D;">    spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">      apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">v1</span></span>
<span class="line"><span style="color:#85E89D;">      kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">ConfigMap</span></span>
<span class="line"><span style="color:#85E89D;">      data</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">        tenant_id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#6A737D;">        # Note: Outputs will be in the secret created by Terraform</span></span>
<span class="line"><span style="color:#85E89D;">        s3_outputs_secret</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-s3-outputs&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">  # Application using S3 bucket</span></span>
<span class="line"><span style="color:#85E89D;">  deployments</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">  - </span><span style="color:#85E89D;">id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">app-deploy</span></span>
<span class="line"><span style="color:#85E89D;">    nameTemplate</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-app&quot;</span></span>
<span class="line"><span style="color:#85E89D;">    dependIds</span><span style="color:#E1E4E8;">: [</span><span style="color:#9ECBFF;">&quot;s3-bucket&quot;</span><span style="color:#E1E4E8;">, </span><span style="color:#9ECBFF;">&quot;app-config&quot;</span><span style="color:#E1E4E8;">]</span></span>
<span class="line"><span style="color:#85E89D;">    waitForReady</span><span style="color:#E1E4E8;">: </span><span style="color:#79B8FF;">true</span></span>
<span class="line"><span style="color:#85E89D;">    timeoutSeconds</span><span style="color:#E1E4E8;">: </span><span style="color:#79B8FF;">600</span><span style="color:#6A737D;">  # Wait up to 10 minutes for Terraform</span></span>
<span class="line"><span style="color:#85E89D;">    spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">      apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">apps/v1</span></span>
<span class="line"><span style="color:#85E89D;">      kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Deployment</span></span>
<span class="line"><span style="color:#85E89D;">      spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">        replicas</span><span style="color:#E1E4E8;">: </span><span style="color:#79B8FF;">1</span></span>
<span class="line"><span style="color:#85E89D;">        selector</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          matchLabels</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">            app</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#85E89D;">        template</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          metadata</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">            labels</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">              app</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#85E89D;">          spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">            containers</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">            - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">app</span></span>
<span class="line"><span style="color:#85E89D;">              image</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">mycompany/app:latest</span></span>
<span class="line"><span style="color:#85E89D;">              env</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">TENANT_ID</span></span>
<span class="line"><span style="color:#85E89D;">                value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#6A737D;">              # S3 bucket name from Terraform output</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">S3_BUCKET_NAME</span></span>
<span class="line"><span style="color:#85E89D;">                valueFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                  secretKeyRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                    name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-s3-outputs&quot;</span></span>
<span class="line"><span style="color:#85E89D;">                    key</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">bucket_name</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">AWS_REGION</span></span>
<span class="line"><span style="color:#85E89D;">                valueFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                  secretKeyRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                    name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">aws-credentials</span></span>
<span class="line"><span style="color:#85E89D;">                    key</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">AWS_DEFAULT_REGION</span></span>
<span class="line"><span style="color:#85E89D;">              envFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">secretRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                  name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">aws-credentials</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br><span class="line-number">12</span><br><span class="line-number">13</span><br><span class="line-number">14</span><br><span class="line-number">15</span><br><span class="line-number">16</span><br><span class="line-number">17</span><br><span class="line-number">18</span><br><span class="line-number">19</span><br><span class="line-number">20</span><br><span class="line-number">21</span><br><span class="line-number">22</span><br><span class="line-number">23</span><br><span class="line-number">24</span><br><span class="line-number">25</span><br><span class="line-number">26</span><br><span class="line-number">27</span><br><span class="line-number">28</span><br><span class="line-number">29</span><br><span class="line-number">30</span><br><span class="line-number">31</span><br><span class="line-number">32</span><br><span class="line-number">33</span><br><span class="line-number">34</span><br><span class="line-number">35</span><br><span class="line-number">36</span><br><span class="line-number">37</span><br><span class="line-number">38</span><br><span class="line-number">39</span><br><span class="line-number">40</span><br><span class="line-number">41</span><br><span class="line-number">42</span><br><span class="line-number">43</span><br><span class="line-number">44</span><br><span class="line-number">45</span><br><span class="line-number">46</span><br><span class="line-number">47</span><br><span class="line-number">48</span><br><span class="line-number">49</span><br><span class="line-number">50</span><br><span class="line-number">51</span><br><span class="line-number">52</span><br><span class="line-number">53</span><br><span class="line-number">54</span><br><span class="line-number">55</span><br><span class="line-number">56</span><br><span class="line-number">57</span><br><span class="line-number">58</span><br><span class="line-number">59</span><br><span class="line-number">60</span><br><span class="line-number">61</span><br><span class="line-number">62</span><br><span class="line-number">63</span><br><span class="line-number">64</span><br><span class="line-number">65</span><br><span class="line-number">66</span><br><span class="line-number">67</span><br><span class="line-number">68</span><br><span class="line-number">69</span><br><span class="line-number">70</span><br><span class="line-number">71</span><br><span class="line-number">72</span><br><span class="line-number">73</span><br><span class="line-number">74</span><br><span class="line-number">75</span><br><span class="line-number">76</span><br><span class="line-number">77</span><br><span class="line-number">78</span><br><span class="line-number">79</span><br><span class="line-number">80</span><br><span class="line-number">81</span><br><span class="line-number">82</span><br><span class="line-number">83</span><br><span class="line-number">84</span><br><span class="line-number">85</span><br><span class="line-number">86</span><br><span class="line-number">87</span><br><span class="line-number">88</span><br><span class="line-number">89</span><br><span class="line-number">90</span><br><span class="line-number">91</span><br><span class="line-number">92</span><br><span class="line-number">93</span><br><span class="line-number">94</span><br><span class="line-number">95</span><br><span class="line-number">96</span><br><span class="line-number">97</span><br><span class="line-number">98</span><br><span class="line-number">99</span><br><span class="line-number">100</span><br><span class="line-number">101</span><br><span class="line-number">102</span><br><span class="line-number">103</span><br><span class="line-number">104</span><br><span class="line-number">105</span><br><span class="line-number">106</span><br><span class="line-number">107</span><br><span class="line-number">108</span><br><span class="line-number">109</span><br><span class="line-number">110</span><br><span class="line-number">111</span><br><span class="line-number">112</span><br><span class="line-number">113</span><br><span class="line-number">114</span><br><span class="line-number">115</span><br><span class="line-number">116</span><br><span class="line-number">117</span><br><span class="line-number">118</span><br><span class="line-number">119</span><br><span class="line-number">120</span><br><span class="line-number">121</span><br><span class="line-number">122</span><br><span class="line-number">123</span><br><span class="line-number">124</span><br><span class="line-number">125</span><br><span class="line-number">126</span><br><span class="line-number">127</span><br><span class="line-number">128</span><br><span class="line-number">129</span><br><span class="line-number">130</span><br><span class="line-number">131</span><br><span class="line-number">132</span><br><span class="line-number">133</span><br><span class="line-number">134</span><br><span class="line-number">135</span><br><span class="line-number">136</span><br><span class="line-number">137</span><br><span class="line-number">138</span><br><span class="line-number">139</span><br><span class="line-number">140</span><br><span class="line-number">141</span><br><span class="line-number">142</span><br><span class="line-number">143</span><br><span class="line-number">144</span><br><span class="line-number">145</span><br><span class="line-number">146</span><br><span class="line-number">147</span><br><span class="line-number">148</span><br><span class="line-number">149</span><br><span class="line-number">150</span><br><span class="line-number">151</span><br><span class="line-number">152</span><br><span class="line-number">153</span><br><span class="line-number">154</span><br><span class="line-number">155</span><br><span class="line-number">156</span><br><span class="line-number">157</span><br><span class="line-number">158</span><br><span class="line-number">159</span><br><span class="line-number">160</span><br><span class="line-number">161</span><br><span class="line-number">162</span><br><span class="line-number">163</span><br><span class="line-number">164</span><br><span class="line-number">165</span><br><span class="line-number">166</span><br><span class="line-number">167</span><br><span class="line-number">168</span><br><span class="line-number">169</span><br></div></div><h2 id="advanced-examples" tabindex="-1">Advanced Examples <a class="header-anchor" href="#advanced-examples" aria-label="Permalink to &quot;Advanced Examples&quot;">​</a></h2><h3 id="example-2-rds-postgresql-database-per-tenant" tabindex="-1">Example 2: RDS PostgreSQL Database per Tenant <a class="header-anchor" href="#example-2-rds-postgresql-database-per-tenant" aria-label="Permalink to &quot;Example 2: RDS PostgreSQL Database per Tenant&quot;">​</a></h3><div class="language-yaml line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">yaml</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#85E89D;">apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">operator.kubernetes-tenants.org/v1</span></span>
<span class="line"><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">TenantTemplate</span></span>
<span class="line"><span style="color:#85E89D;">metadata</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant-with-rds</span></span>
<span class="line"><span style="color:#85E89D;">  namespace</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">default</span></span>
<span class="line"><span style="color:#85E89D;">spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  registryId</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">my-registry</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">  manifests</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">  - </span><span style="color:#85E89D;">id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">rds-database</span></span>
<span class="line"><span style="color:#85E89D;">    nameTemplate</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-rds&quot;</span></span>
<span class="line"><span style="color:#85E89D;">    creationPolicy</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Once</span><span style="color:#6A737D;">  # Create once, don&#39;t modify</span></span>
<span class="line"><span style="color:#85E89D;">    deletionPolicy</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Retain</span><span style="color:#6A737D;">  # Keep database when tenant deleted</span></span>
<span class="line"><span style="color:#85E89D;">    spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">      apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">infra.contrib.fluxcd.io/v1alpha2</span></span>
<span class="line"><span style="color:#85E89D;">      kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Terraform</span></span>
<span class="line"><span style="color:#85E89D;">      spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">        interval</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">10m</span></span>
<span class="line"><span style="color:#85E89D;">        retryInterval</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">1m</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        values</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          hcl</span><span style="color:#E1E4E8;">: </span><span style="color:#F97583;">|</span></span>
<span class="line"><span style="color:#9ECBFF;">            terraform {</span></span>
<span class="line"><span style="color:#9ECBFF;">              required_providers {</span></span>
<span class="line"><span style="color:#9ECBFF;">                aws = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  source  = &quot;hashicorp/aws&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                  version = &quot;~&gt; 5.0&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">                random = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  source  = &quot;hashicorp/random&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                  version = &quot;~&gt; 3.5&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">              backend &quot;kubernetes&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">                secret_suffix = &quot;{{ .uid }}-rds&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                namespace     = &quot;default&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            provider &quot;aws&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              region = var.aws_region</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;tenant_id&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              type = string</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;aws_region&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              type    = string</span></span>
<span class="line"><span style="color:#9ECBFF;">              default = &quot;us-east-1&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;db_instance_class&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              type    = string</span></span>
<span class="line"><span style="color:#9ECBFF;">              default = &quot;db.t3.micro&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;db_allocated_storage&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              type    = number</span></span>
<span class="line"><span style="color:#9ECBFF;">              default = 20</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # Generate random password</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;random_password&quot; &quot;db_password&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              length  = 32</span></span>
<span class="line"><span style="color:#9ECBFF;">              special = true</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # Security group for RDS</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;aws_security_group&quot; &quot;rds_sg&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              name_prefix = &quot;tenant-\${var.tenant_id}-rds-&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              description = &quot;Security group for tenant \${var.tenant_id} RDS&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              ingress {</span></span>
<span class="line"><span style="color:#9ECBFF;">                from_port   = 5432</span></span>
<span class="line"><span style="color:#9ECBFF;">                to_port     = 5432</span></span>
<span class="line"><span style="color:#9ECBFF;">                protocol    = &quot;tcp&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                cidr_blocks = [&quot;10.0.0.0/8&quot;]  # Adjust to your VPC CIDR</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              egress {</span></span>
<span class="line"><span style="color:#9ECBFF;">                from_port   = 0</span></span>
<span class="line"><span style="color:#9ECBFF;">                to_port     = 0</span></span>
<span class="line"><span style="color:#9ECBFF;">                protocol    = &quot;-1&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                cidr_blocks = [&quot;0.0.0.0/0&quot;]</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              tags = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                Name      = &quot;tenant-\${var.tenant_id}-rds-sg&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                TenantId  = var.tenant_id</span></span>
<span class="line"><span style="color:#9ECBFF;">                ManagedBy = &quot;tenant-operator&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # RDS instance</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;aws_db_instance&quot; &quot;tenant_db&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              identifier     = &quot;tenant-\${var.tenant_id}-db&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              engine         = &quot;postgres&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              engine_version = &quot;15.4&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              instance_class    = var.db_instance_class</span></span>
<span class="line"><span style="color:#9ECBFF;">              allocated_storage = var.db_allocated_storage</span></span>
<span class="line"><span style="color:#9ECBFF;">              storage_type      = &quot;gp3&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              storage_encrypted = true</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              db_name  = &quot;tenant_\${replace(var.tenant_id, &quot;-&quot;, &quot;_&quot;)}&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              username = &quot;dbadmin&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              password = random_password.db_password.result</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              vpc_security_group_ids = [aws_security_group.rds_sg.id]</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              backup_retention_period = 7</span></span>
<span class="line"><span style="color:#9ECBFF;">              backup_window          = &quot;03:00-04:00&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              maintenance_window     = &quot;mon:04:00-mon:05:00&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              skip_final_snapshot = false</span></span>
<span class="line"><span style="color:#9ECBFF;">              final_snapshot_identifier = &quot;tenant-\${var.tenant_id}-final-snapshot&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              tags = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                Name      = &quot;tenant-\${var.tenant_id}-db&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                TenantId  = var.tenant_id</span></span>
<span class="line"><span style="color:#9ECBFF;">                ManagedBy = &quot;tenant-operator&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;db_endpoint&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              value     = aws_db_instance.tenant_db.endpoint</span></span>
<span class="line"><span style="color:#9ECBFF;">              sensitive = false</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;db_name&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              value = aws_db_instance.tenant_db.db_name</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;db_username&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              value = aws_db_instance.tenant_db.username</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;db_password&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              value     = random_password.db_password.result</span></span>
<span class="line"><span style="color:#9ECBFF;">              sensitive = true</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;db_port&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              value = aws_db_instance.tenant_db.port</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        vars</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant_id</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        varsFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Secret</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">aws-credentials</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        writeOutputsToSecret</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-db-credentials&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">  # Application using RDS</span></span>
<span class="line"><span style="color:#85E89D;">  deployments</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">  - </span><span style="color:#85E89D;">id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">app-deploy</span></span>
<span class="line"><span style="color:#85E89D;">    nameTemplate</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-app&quot;</span></span>
<span class="line"><span style="color:#85E89D;">    dependIds</span><span style="color:#E1E4E8;">: [</span><span style="color:#9ECBFF;">&quot;rds-database&quot;</span><span style="color:#E1E4E8;">]</span></span>
<span class="line"><span style="color:#85E89D;">    timeoutSeconds</span><span style="color:#E1E4E8;">: </span><span style="color:#79B8FF;">900</span><span style="color:#6A737D;">  # 15 minutes for RDS provisioning</span></span>
<span class="line"><span style="color:#85E89D;">    spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">      apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">apps/v1</span></span>
<span class="line"><span style="color:#85E89D;">      kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Deployment</span></span>
<span class="line"><span style="color:#85E89D;">      spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">        replicas</span><span style="color:#E1E4E8;">: </span><span style="color:#79B8FF;">2</span></span>
<span class="line"><span style="color:#85E89D;">        selector</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          matchLabels</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">            app</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#85E89D;">        template</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          metadata</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">            labels</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">              app</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#85E89D;">          spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">            containers</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">            - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">app</span></span>
<span class="line"><span style="color:#85E89D;">              image</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">mycompany/app:latest</span></span>
<span class="line"><span style="color:#85E89D;">              env</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">TENANT_ID</span></span>
<span class="line"><span style="color:#85E89D;">                value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">DB_HOST</span></span>
<span class="line"><span style="color:#85E89D;">                valueFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                  secretKeyRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                    name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-db-credentials&quot;</span></span>
<span class="line"><span style="color:#85E89D;">                    key</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">db_endpoint</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">DB_NAME</span></span>
<span class="line"><span style="color:#85E89D;">                valueFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                  secretKeyRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                    name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-db-credentials&quot;</span></span>
<span class="line"><span style="color:#85E89D;">                    key</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">db_name</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">DB_USER</span></span>
<span class="line"><span style="color:#85E89D;">                valueFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                  secretKeyRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                    name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-db-credentials&quot;</span></span>
<span class="line"><span style="color:#85E89D;">                    key</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">db_username</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">DB_PASSWORD</span></span>
<span class="line"><span style="color:#85E89D;">                valueFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                  secretKeyRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                    name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-db-credentials&quot;</span></span>
<span class="line"><span style="color:#85E89D;">                    key</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">db_password</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">DB_PORT</span></span>
<span class="line"><span style="color:#85E89D;">                valueFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                  secretKeyRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                    name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-db-credentials&quot;</span></span>
<span class="line"><span style="color:#85E89D;">                    key</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">db_port</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br><span class="line-number">12</span><br><span class="line-number">13</span><br><span class="line-number">14</span><br><span class="line-number">15</span><br><span class="line-number">16</span><br><span class="line-number">17</span><br><span class="line-number">18</span><br><span class="line-number">19</span><br><span class="line-number">20</span><br><span class="line-number">21</span><br><span class="line-number">22</span><br><span class="line-number">23</span><br><span class="line-number">24</span><br><span class="line-number">25</span><br><span class="line-number">26</span><br><span class="line-number">27</span><br><span class="line-number">28</span><br><span class="line-number">29</span><br><span class="line-number">30</span><br><span class="line-number">31</span><br><span class="line-number">32</span><br><span class="line-number">33</span><br><span class="line-number">34</span><br><span class="line-number">35</span><br><span class="line-number">36</span><br><span class="line-number">37</span><br><span class="line-number">38</span><br><span class="line-number">39</span><br><span class="line-number">40</span><br><span class="line-number">41</span><br><span class="line-number">42</span><br><span class="line-number">43</span><br><span class="line-number">44</span><br><span class="line-number">45</span><br><span class="line-number">46</span><br><span class="line-number">47</span><br><span class="line-number">48</span><br><span class="line-number">49</span><br><span class="line-number">50</span><br><span class="line-number">51</span><br><span class="line-number">52</span><br><span class="line-number">53</span><br><span class="line-number">54</span><br><span class="line-number">55</span><br><span class="line-number">56</span><br><span class="line-number">57</span><br><span class="line-number">58</span><br><span class="line-number">59</span><br><span class="line-number">60</span><br><span class="line-number">61</span><br><span class="line-number">62</span><br><span class="line-number">63</span><br><span class="line-number">64</span><br><span class="line-number">65</span><br><span class="line-number">66</span><br><span class="line-number">67</span><br><span class="line-number">68</span><br><span class="line-number">69</span><br><span class="line-number">70</span><br><span class="line-number">71</span><br><span class="line-number">72</span><br><span class="line-number">73</span><br><span class="line-number">74</span><br><span class="line-number">75</span><br><span class="line-number">76</span><br><span class="line-number">77</span><br><span class="line-number">78</span><br><span class="line-number">79</span><br><span class="line-number">80</span><br><span class="line-number">81</span><br><span class="line-number">82</span><br><span class="line-number">83</span><br><span class="line-number">84</span><br><span class="line-number">85</span><br><span class="line-number">86</span><br><span class="line-number">87</span><br><span class="line-number">88</span><br><span class="line-number">89</span><br><span class="line-number">90</span><br><span class="line-number">91</span><br><span class="line-number">92</span><br><span class="line-number">93</span><br><span class="line-number">94</span><br><span class="line-number">95</span><br><span class="line-number">96</span><br><span class="line-number">97</span><br><span class="line-number">98</span><br><span class="line-number">99</span><br><span class="line-number">100</span><br><span class="line-number">101</span><br><span class="line-number">102</span><br><span class="line-number">103</span><br><span class="line-number">104</span><br><span class="line-number">105</span><br><span class="line-number">106</span><br><span class="line-number">107</span><br><span class="line-number">108</span><br><span class="line-number">109</span><br><span class="line-number">110</span><br><span class="line-number">111</span><br><span class="line-number">112</span><br><span class="line-number">113</span><br><span class="line-number">114</span><br><span class="line-number">115</span><br><span class="line-number">116</span><br><span class="line-number">117</span><br><span class="line-number">118</span><br><span class="line-number">119</span><br><span class="line-number">120</span><br><span class="line-number">121</span><br><span class="line-number">122</span><br><span class="line-number">123</span><br><span class="line-number">124</span><br><span class="line-number">125</span><br><span class="line-number">126</span><br><span class="line-number">127</span><br><span class="line-number">128</span><br><span class="line-number">129</span><br><span class="line-number">130</span><br><span class="line-number">131</span><br><span class="line-number">132</span><br><span class="line-number">133</span><br><span class="line-number">134</span><br><span class="line-number">135</span><br><span class="line-number">136</span><br><span class="line-number">137</span><br><span class="line-number">138</span><br><span class="line-number">139</span><br><span class="line-number">140</span><br><span class="line-number">141</span><br><span class="line-number">142</span><br><span class="line-number">143</span><br><span class="line-number">144</span><br><span class="line-number">145</span><br><span class="line-number">146</span><br><span class="line-number">147</span><br><span class="line-number">148</span><br><span class="line-number">149</span><br><span class="line-number">150</span><br><span class="line-number">151</span><br><span class="line-number">152</span><br><span class="line-number">153</span><br><span class="line-number">154</span><br><span class="line-number">155</span><br><span class="line-number">156</span><br><span class="line-number">157</span><br><span class="line-number">158</span><br><span class="line-number">159</span><br><span class="line-number">160</span><br><span class="line-number">161</span><br><span class="line-number">162</span><br><span class="line-number">163</span><br><span class="line-number">164</span><br><span class="line-number">165</span><br><span class="line-number">166</span><br><span class="line-number">167</span><br><span class="line-number">168</span><br><span class="line-number">169</span><br><span class="line-number">170</span><br><span class="line-number">171</span><br><span class="line-number">172</span><br><span class="line-number">173</span><br><span class="line-number">174</span><br><span class="line-number">175</span><br><span class="line-number">176</span><br><span class="line-number">177</span><br><span class="line-number">178</span><br><span class="line-number">179</span><br><span class="line-number">180</span><br><span class="line-number">181</span><br><span class="line-number">182</span><br><span class="line-number">183</span><br><span class="line-number">184</span><br><span class="line-number">185</span><br><span class="line-number">186</span><br><span class="line-number">187</span><br><span class="line-number">188</span><br><span class="line-number">189</span><br><span class="line-number">190</span><br><span class="line-number">191</span><br><span class="line-number">192</span><br><span class="line-number">193</span><br><span class="line-number">194</span><br><span class="line-number">195</span><br><span class="line-number">196</span><br><span class="line-number">197</span><br><span class="line-number">198</span><br><span class="line-number">199</span><br><span class="line-number">200</span><br><span class="line-number">201</span><br><span class="line-number">202</span><br><span class="line-number">203</span><br><span class="line-number">204</span><br><span class="line-number">205</span><br><span class="line-number">206</span><br><span class="line-number">207</span><br><span class="line-number">208</span><br></div></div><h3 id="example-3-cloudfront-cdn-distribution" tabindex="-1">Example 3: CloudFront CDN Distribution <a class="header-anchor" href="#example-3-cloudfront-cdn-distribution" aria-label="Permalink to &quot;Example 3: CloudFront CDN Distribution&quot;">​</a></h3><div class="language-yaml line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">yaml</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#85E89D;">apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">operator.kubernetes-tenants.org/v1</span></span>
<span class="line"><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">TenantTemplate</span></span>
<span class="line"><span style="color:#85E89D;">metadata</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant-with-cdn</span></span>
<span class="line"><span style="color:#85E89D;">  namespace</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">default</span></span>
<span class="line"><span style="color:#85E89D;">spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  registryId</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">my-registry</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">  manifests</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">  - </span><span style="color:#85E89D;">id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">cloudfront-cdn</span></span>
<span class="line"><span style="color:#85E89D;">    nameTemplate</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-cdn&quot;</span></span>
<span class="line"><span style="color:#85E89D;">    spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">      apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">infra.contrib.fluxcd.io/v1alpha2</span></span>
<span class="line"><span style="color:#85E89D;">      kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Terraform</span></span>
<span class="line"><span style="color:#85E89D;">      spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">        interval</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">5m</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        values</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          hcl</span><span style="color:#E1E4E8;">: </span><span style="color:#F97583;">|</span></span>
<span class="line"><span style="color:#9ECBFF;">            terraform {</span></span>
<span class="line"><span style="color:#9ECBFF;">              required_providers {</span></span>
<span class="line"><span style="color:#9ECBFF;">                aws = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  source  = &quot;hashicorp/aws&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                  version = &quot;~&gt; 5.0&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">              backend &quot;kubernetes&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">                secret_suffix = &quot;{{ .uid }}-cdn&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                namespace     = &quot;default&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            provider &quot;aws&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              region = &quot;us-east-1&quot;  # CloudFront is global</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;tenant_id&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              type = string</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;origin_domain&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              type = string</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # S3 bucket for CDN logs</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;aws_s3_bucket&quot; &quot;cdn_logs&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              bucket = &quot;tenant-\${var.tenant_id}-cdn-logs&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              tags = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                Name      = &quot;tenant-\${var.tenant_id}-cdn-logs&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                TenantId  = var.tenant_id</span></span>
<span class="line"><span style="color:#9ECBFF;">                ManagedBy = &quot;tenant-operator&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # CloudFront distribution</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;aws_cloudfront_distribution&quot; &quot;cdn&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              enabled             = true</span></span>
<span class="line"><span style="color:#9ECBFF;">              is_ipv6_enabled     = true</span></span>
<span class="line"><span style="color:#9ECBFF;">              comment             = &quot;CDN for tenant \${var.tenant_id}&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              default_root_object = &quot;index.html&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              origin {</span></span>
<span class="line"><span style="color:#9ECBFF;">                domain_name = var.origin_domain</span></span>
<span class="line"><span style="color:#9ECBFF;">                origin_id   = &quot;tenant-\${var.tenant_id}-origin&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">                custom_origin_config {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  http_port              = 80</span></span>
<span class="line"><span style="color:#9ECBFF;">                  https_port             = 443</span></span>
<span class="line"><span style="color:#9ECBFF;">                  origin_protocol_policy = &quot;https-only&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                  origin_ssl_protocols   = [&quot;TLSv1.2&quot;]</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              default_cache_behavior {</span></span>
<span class="line"><span style="color:#9ECBFF;">                allowed_methods  = [&quot;GET&quot;, &quot;HEAD&quot;, &quot;OPTIONS&quot;]</span></span>
<span class="line"><span style="color:#9ECBFF;">                cached_methods   = [&quot;GET&quot;, &quot;HEAD&quot;]</span></span>
<span class="line"><span style="color:#9ECBFF;">                target_origin_id = &quot;tenant-\${var.tenant_id}-origin&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">                forwarded_values {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  query_string = true</span></span>
<span class="line"><span style="color:#9ECBFF;">                  cookies {</span></span>
<span class="line"><span style="color:#9ECBFF;">                    forward = &quot;none&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                  }</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">                viewer_protocol_policy = &quot;redirect-to-https&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                min_ttl                = 0</span></span>
<span class="line"><span style="color:#9ECBFF;">                default_ttl            = 3600</span></span>
<span class="line"><span style="color:#9ECBFF;">                max_ttl                = 86400</span></span>
<span class="line"><span style="color:#9ECBFF;">                compress               = true</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              restrictions {</span></span>
<span class="line"><span style="color:#9ECBFF;">                geo_restriction {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  restriction_type = &quot;none&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              viewer_certificate {</span></span>
<span class="line"><span style="color:#9ECBFF;">                cloudfront_default_certificate = true</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              logging_config {</span></span>
<span class="line"><span style="color:#9ECBFF;">                include_cookies = false</span></span>
<span class="line"><span style="color:#9ECBFF;">                bucket          = aws_s3_bucket.cdn_logs.bucket_domain_name</span></span>
<span class="line"><span style="color:#9ECBFF;">                prefix          = &quot;cdn-logs/&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              tags = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                Name      = &quot;tenant-\${var.tenant_id}-cdn&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                TenantId  = var.tenant_id</span></span>
<span class="line"><span style="color:#9ECBFF;">                ManagedBy = &quot;tenant-operator&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;cdn_domain_name&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              value = aws_cloudfront_distribution.cdn.domain_name</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;cdn_distribution_id&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              value = aws_cloudfront_distribution.cdn.id</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;cdn_arn&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              value = aws_cloudfront_distribution.cdn.arn</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        vars</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant_id</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">origin_domain</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .host }}&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        varsFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Secret</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">aws-credentials</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        writeOutputsToSecret</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-cdn-outputs&quot;</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br><span class="line-number">12</span><br><span class="line-number">13</span><br><span class="line-number">14</span><br><span class="line-number">15</span><br><span class="line-number">16</span><br><span class="line-number">17</span><br><span class="line-number">18</span><br><span class="line-number">19</span><br><span class="line-number">20</span><br><span class="line-number">21</span><br><span class="line-number">22</span><br><span class="line-number">23</span><br><span class="line-number">24</span><br><span class="line-number">25</span><br><span class="line-number">26</span><br><span class="line-number">27</span><br><span class="line-number">28</span><br><span class="line-number">29</span><br><span class="line-number">30</span><br><span class="line-number">31</span><br><span class="line-number">32</span><br><span class="line-number">33</span><br><span class="line-number">34</span><br><span class="line-number">35</span><br><span class="line-number">36</span><br><span class="line-number">37</span><br><span class="line-number">38</span><br><span class="line-number">39</span><br><span class="line-number">40</span><br><span class="line-number">41</span><br><span class="line-number">42</span><br><span class="line-number">43</span><br><span class="line-number">44</span><br><span class="line-number">45</span><br><span class="line-number">46</span><br><span class="line-number">47</span><br><span class="line-number">48</span><br><span class="line-number">49</span><br><span class="line-number">50</span><br><span class="line-number">51</span><br><span class="line-number">52</span><br><span class="line-number">53</span><br><span class="line-number">54</span><br><span class="line-number">55</span><br><span class="line-number">56</span><br><span class="line-number">57</span><br><span class="line-number">58</span><br><span class="line-number">59</span><br><span class="line-number">60</span><br><span class="line-number">61</span><br><span class="line-number">62</span><br><span class="line-number">63</span><br><span class="line-number">64</span><br><span class="line-number">65</span><br><span class="line-number">66</span><br><span class="line-number">67</span><br><span class="line-number">68</span><br><span class="line-number">69</span><br><span class="line-number">70</span><br><span class="line-number">71</span><br><span class="line-number">72</span><br><span class="line-number">73</span><br><span class="line-number">74</span><br><span class="line-number">75</span><br><span class="line-number">76</span><br><span class="line-number">77</span><br><span class="line-number">78</span><br><span class="line-number">79</span><br><span class="line-number">80</span><br><span class="line-number">81</span><br><span class="line-number">82</span><br><span class="line-number">83</span><br><span class="line-number">84</span><br><span class="line-number">85</span><br><span class="line-number">86</span><br><span class="line-number">87</span><br><span class="line-number">88</span><br><span class="line-number">89</span><br><span class="line-number">90</span><br><span class="line-number">91</span><br><span class="line-number">92</span><br><span class="line-number">93</span><br><span class="line-number">94</span><br><span class="line-number">95</span><br><span class="line-number">96</span><br><span class="line-number">97</span><br><span class="line-number">98</span><br><span class="line-number">99</span><br><span class="line-number">100</span><br><span class="line-number">101</span><br><span class="line-number">102</span><br><span class="line-number">103</span><br><span class="line-number">104</span><br><span class="line-number">105</span><br><span class="line-number">106</span><br><span class="line-number">107</span><br><span class="line-number">108</span><br><span class="line-number">109</span><br><span class="line-number">110</span><br><span class="line-number">111</span><br><span class="line-number">112</span><br><span class="line-number">113</span><br><span class="line-number">114</span><br><span class="line-number">115</span><br><span class="line-number">116</span><br><span class="line-number">117</span><br><span class="line-number">118</span><br><span class="line-number">119</span><br><span class="line-number">120</span><br><span class="line-number">121</span><br><span class="line-number">122</span><br><span class="line-number">123</span><br><span class="line-number">124</span><br><span class="line-number">125</span><br><span class="line-number">126</span><br><span class="line-number">127</span><br><span class="line-number">128</span><br><span class="line-number">129</span><br><span class="line-number">130</span><br><span class="line-number">131</span><br><span class="line-number">132</span><br><span class="line-number">133</span><br><span class="line-number">134</span><br><span class="line-number">135</span><br><span class="line-number">136</span><br><span class="line-number">137</span><br><span class="line-number">138</span><br><span class="line-number">139</span><br><span class="line-number">140</span><br></div></div><h3 id="example-4-using-git-repository-for-terraform-modules" tabindex="-1">Example 4: Using Git Repository for Terraform Modules <a class="header-anchor" href="#example-4-using-git-repository-for-terraform-modules" aria-label="Permalink to &quot;Example 4: Using Git Repository for Terraform Modules&quot;">​</a></h3><p><strong>Create GitRepository:</strong></p><div class="language-yaml line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">yaml</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#85E89D;">apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">source.toolkit.fluxcd.io/v1</span></span>
<span class="line"><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">GitRepository</span></span>
<span class="line"><span style="color:#85E89D;">metadata</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">terraform-modules</span></span>
<span class="line"><span style="color:#85E89D;">  namespace</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">default</span></span>
<span class="line"><span style="color:#85E89D;">spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  interval</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">5m</span></span>
<span class="line"><span style="color:#85E89D;">  url</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">https://github.com/your-org/terraform-modules</span></span>
<span class="line"><span style="color:#85E89D;">  ref</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">    branch</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">main</span></span>
<span class="line"><span style="color:#6A737D;">  # Optional: Use SSH key for private repos</span></span>
<span class="line"><span style="color:#6A737D;">  # secretRef:</span></span>
<span class="line"><span style="color:#6A737D;">  #   name: git-credentials</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br><span class="line-number">12</span><br><span class="line-number">13</span><br></div></div><p><strong>TenantTemplate using Git modules:</strong></p><div class="language-yaml line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">yaml</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#85E89D;">apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">operator.kubernetes-tenants.org/v1</span></span>
<span class="line"><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">TenantTemplate</span></span>
<span class="line"><span style="color:#85E89D;">metadata</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant-with-git-modules</span></span>
<span class="line"><span style="color:#85E89D;">  namespace</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">default</span></span>
<span class="line"><span style="color:#85E89D;">spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  registryId</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">my-registry</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">  manifests</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">  - </span><span style="color:#85E89D;">id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant-infrastructure</span></span>
<span class="line"><span style="color:#85E89D;">    nameTemplate</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-infra&quot;</span></span>
<span class="line"><span style="color:#85E89D;">    spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">      apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">infra.contrib.fluxcd.io/v1alpha2</span></span>
<span class="line"><span style="color:#85E89D;">      kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Terraform</span></span>
<span class="line"><span style="color:#85E89D;">      spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">        interval</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">10m</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">        # Reference Git repository</span></span>
<span class="line"><span style="color:#85E89D;">        sourceRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">GitRepository</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">terraform-modules</span></span>
<span class="line"><span style="color:#85E89D;">          namespace</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">default</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">        # Path to module in repository</span></span>
<span class="line"><span style="color:#85E89D;">        path</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">./modules/tenant-stack</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">        # Pass variables to module</span></span>
<span class="line"><span style="color:#85E89D;">        vars</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant_id</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant_host</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .host }}&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">environment</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;production&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        varsFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Secret</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">aws-credentials</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        writeOutputsToSecret</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-infra-outputs&quot;</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br><span class="line-number">12</span><br><span class="line-number">13</span><br><span class="line-number">14</span><br><span class="line-number">15</span><br><span class="line-number">16</span><br><span class="line-number">17</span><br><span class="line-number">18</span><br><span class="line-number">19</span><br><span class="line-number">20</span><br><span class="line-number">21</span><br><span class="line-number">22</span><br><span class="line-number">23</span><br><span class="line-number">24</span><br><span class="line-number">25</span><br><span class="line-number">26</span><br><span class="line-number">27</span><br><span class="line-number">28</span><br><span class="line-number">29</span><br><span class="line-number">30</span><br><span class="line-number">31</span><br><span class="line-number">32</span><br><span class="line-number">33</span><br><span class="line-number">34</span><br><span class="line-number">35</span><br><span class="line-number">36</span><br><span class="line-number">37</span><br><span class="line-number">38</span><br><span class="line-number">39</span><br><span class="line-number">40</span><br><span class="line-number">41</span><br></div></div><p><strong>Example Terraform module structure in Git:</strong></p><div class="language- line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang"></span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span>terraform-modules/</span></span>
<span class="line"><span>├── modules/</span></span>
<span class="line"><span>│   ├── tenant-stack/</span></span>
<span class="line"><span>│   │   ├── main.tf</span></span>
<span class="line"><span>│   │   ├── variables.tf</span></span>
<span class="line"><span>│   │   ├── outputs.tf</span></span>
<span class="line"><span>│   │   ├── s3.tf</span></span>
<span class="line"><span>│   │   ├── rds.tf</span></span>
<span class="line"><span>│   │   └── cloudfront.tf</span></span>
<span class="line"><span>│   ├── networking/</span></span>
<span class="line"><span>│   └── security/</span></span>
<span class="line"><span>└── README.md</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br><span class="line-number">12</span><br></div></div><h3 id="example-5-kafka-topics-and-acls-per-tenant" tabindex="-1">Example 5: Kafka Topics and ACLs per Tenant <a class="header-anchor" href="#example-5-kafka-topics-and-acls-per-tenant" aria-label="Permalink to &quot;Example 5: Kafka Topics and ACLs per Tenant&quot;">​</a></h3><p>Provision dedicated Kafka topics and access controls for each tenant:</p><div class="language-yaml line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">yaml</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#85E89D;">apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">operator.kubernetes-tenants.org/v1</span></span>
<span class="line"><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">TenantTemplate</span></span>
<span class="line"><span style="color:#85E89D;">metadata</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant-with-kafka</span></span>
<span class="line"><span style="color:#85E89D;">  namespace</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">default</span></span>
<span class="line"><span style="color:#85E89D;">spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  registryId</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">my-registry</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">  manifests</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">  - </span><span style="color:#85E89D;">id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">kafka-resources</span></span>
<span class="line"><span style="color:#85E89D;">    nameTemplate</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-kafka&quot;</span></span>
<span class="line"><span style="color:#85E89D;">    spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">      apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">infra.contrib.fluxcd.io/v1alpha2</span></span>
<span class="line"><span style="color:#85E89D;">      kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Terraform</span></span>
<span class="line"><span style="color:#85E89D;">      spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">        interval</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">5m</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        values</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          hcl</span><span style="color:#E1E4E8;">: </span><span style="color:#F97583;">|</span></span>
<span class="line"><span style="color:#9ECBFF;">            terraform {</span></span>
<span class="line"><span style="color:#9ECBFF;">              required_providers {</span></span>
<span class="line"><span style="color:#9ECBFF;">                kafka = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  source  = &quot;Mongey/kafka&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                  version = &quot;~&gt; 0.7&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">              backend &quot;kubernetes&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">                secret_suffix = &quot;{{ .uid }}-kafka&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                namespace     = &quot;default&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            provider &quot;kafka&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              bootstrap_servers = var.kafka_bootstrap_servers</span></span>
<span class="line"><span style="color:#9ECBFF;">              tls_enabled       = true</span></span>
<span class="line"><span style="color:#9ECBFF;">              sasl_username     = var.kafka_username</span></span>
<span class="line"><span style="color:#9ECBFF;">              sasl_password     = var.kafka_password</span></span>
<span class="line"><span style="color:#9ECBFF;">              sasl_mechanism    = &quot;plain&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;tenant_id&quot; { type = string }</span></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;kafka_bootstrap_servers&quot; { type = list(string) }</span></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;kafka_username&quot; { type = string }</span></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;kafka_password&quot; { type = string sensitive = true }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # Topics for tenant</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;kafka_topic&quot; &quot;events&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              name               = &quot;tenant-\${var.tenant_id}-events&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              replication_factor = 3</span></span>
<span class="line"><span style="color:#9ECBFF;">              partitions         = 6</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              config = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                &quot;cleanup.policy&quot; = &quot;delete&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                &quot;retention.ms&quot;   = &quot;604800000&quot;  # 7 days</span></span>
<span class="line"><span style="color:#9ECBFF;">                &quot;segment.ms&quot;     = &quot;86400000&quot;   # 1 day</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;kafka_topic&quot; &quot;commands&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              name               = &quot;tenant-\${var.tenant_id}-commands&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              replication_factor = 3</span></span>
<span class="line"><span style="color:#9ECBFF;">              partitions         = 3</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              config = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                &quot;cleanup.policy&quot; = &quot;delete&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                &quot;retention.ms&quot;   = &quot;259200000&quot;  # 3 days</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;kafka_topic&quot; &quot;dlq&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              name               = &quot;tenant-\${var.tenant_id}-dlq&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              replication_factor = 3</span></span>
<span class="line"><span style="color:#9ECBFF;">              partitions         = 1</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              config = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                &quot;cleanup.policy&quot; = &quot;delete&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                &quot;retention.ms&quot;   = &quot;2592000000&quot;  # 30 days</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # ACLs for tenant</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;kafka_acl&quot; &quot;tenant_producer&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              resource_name             = &quot;tenant-\${var.tenant_id}-*&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              resource_type             = &quot;Topic&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              acl_principal             = &quot;User:tenant-\${var.tenant_id}&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              acl_host                  = &quot;*&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              acl_operation             = &quot;Write&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              acl_permission_type       = &quot;Allow&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              resource_pattern_type_filter = &quot;Prefixed&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;kafka_acl&quot; &quot;tenant_consumer&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              resource_name             = &quot;tenant-\${var.tenant_id}-*&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              resource_type             = &quot;Topic&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              acl_principal             = &quot;User:tenant-\${var.tenant_id}&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              acl_host                  = &quot;*&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              acl_operation             = &quot;Read&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              acl_permission_type       = &quot;Allow&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              resource_pattern_type_filter = &quot;Prefixed&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;kafka_acl&quot; &quot;tenant_consumer_group&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              resource_name             = &quot;tenant-\${var.tenant_id}-*&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              resource_type             = &quot;Group&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              acl_principal             = &quot;User:tenant-\${var.tenant_id}&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              acl_host                  = &quot;*&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              acl_operation             = &quot;Read&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              acl_permission_type       = &quot;Allow&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              resource_pattern_type_filter = &quot;Prefixed&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;events_topic&quot; { value = kafka_topic.events.name }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;commands_topic&quot; { value = kafka_topic.commands.name }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;dlq_topic&quot; { value = kafka_topic.dlq.name }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        vars</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant_id</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">kafka_bootstrap_servers</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&#39;[&quot;kafka-broker-1:9092&quot;,&quot;kafka-broker-2:9092&quot;,&quot;kafka-broker-3:9092&quot;]&#39;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        varsFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Secret</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">kafka-credentials</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        writeOutputsToSecret</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-kafka-outputs&quot;</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br><span class="line-number">12</span><br><span class="line-number">13</span><br><span class="line-number">14</span><br><span class="line-number">15</span><br><span class="line-number">16</span><br><span class="line-number">17</span><br><span class="line-number">18</span><br><span class="line-number">19</span><br><span class="line-number">20</span><br><span class="line-number">21</span><br><span class="line-number">22</span><br><span class="line-number">23</span><br><span class="line-number">24</span><br><span class="line-number">25</span><br><span class="line-number">26</span><br><span class="line-number">27</span><br><span class="line-number">28</span><br><span class="line-number">29</span><br><span class="line-number">30</span><br><span class="line-number">31</span><br><span class="line-number">32</span><br><span class="line-number">33</span><br><span class="line-number">34</span><br><span class="line-number">35</span><br><span class="line-number">36</span><br><span class="line-number">37</span><br><span class="line-number">38</span><br><span class="line-number">39</span><br><span class="line-number">40</span><br><span class="line-number">41</span><br><span class="line-number">42</span><br><span class="line-number">43</span><br><span class="line-number">44</span><br><span class="line-number">45</span><br><span class="line-number">46</span><br><span class="line-number">47</span><br><span class="line-number">48</span><br><span class="line-number">49</span><br><span class="line-number">50</span><br><span class="line-number">51</span><br><span class="line-number">52</span><br><span class="line-number">53</span><br><span class="line-number">54</span><br><span class="line-number">55</span><br><span class="line-number">56</span><br><span class="line-number">57</span><br><span class="line-number">58</span><br><span class="line-number">59</span><br><span class="line-number">60</span><br><span class="line-number">61</span><br><span class="line-number">62</span><br><span class="line-number">63</span><br><span class="line-number">64</span><br><span class="line-number">65</span><br><span class="line-number">66</span><br><span class="line-number">67</span><br><span class="line-number">68</span><br><span class="line-number">69</span><br><span class="line-number">70</span><br><span class="line-number">71</span><br><span class="line-number">72</span><br><span class="line-number">73</span><br><span class="line-number">74</span><br><span class="line-number">75</span><br><span class="line-number">76</span><br><span class="line-number">77</span><br><span class="line-number">78</span><br><span class="line-number">79</span><br><span class="line-number">80</span><br><span class="line-number">81</span><br><span class="line-number">82</span><br><span class="line-number">83</span><br><span class="line-number">84</span><br><span class="line-number">85</span><br><span class="line-number">86</span><br><span class="line-number">87</span><br><span class="line-number">88</span><br><span class="line-number">89</span><br><span class="line-number">90</span><br><span class="line-number">91</span><br><span class="line-number">92</span><br><span class="line-number">93</span><br><span class="line-number">94</span><br><span class="line-number">95</span><br><span class="line-number">96</span><br><span class="line-number">97</span><br><span class="line-number">98</span><br><span class="line-number">99</span><br><span class="line-number">100</span><br><span class="line-number">101</span><br><span class="line-number">102</span><br><span class="line-number">103</span><br><span class="line-number">104</span><br><span class="line-number">105</span><br><span class="line-number">106</span><br><span class="line-number">107</span><br><span class="line-number">108</span><br><span class="line-number">109</span><br><span class="line-number">110</span><br><span class="line-number">111</span><br><span class="line-number">112</span><br><span class="line-number">113</span><br><span class="line-number">114</span><br><span class="line-number">115</span><br><span class="line-number">116</span><br><span class="line-number">117</span><br><span class="line-number">118</span><br><span class="line-number">119</span><br><span class="line-number">120</span><br><span class="line-number">121</span><br><span class="line-number">122</span><br><span class="line-number">123</span><br><span class="line-number">124</span><br><span class="line-number">125</span><br><span class="line-number">126</span><br><span class="line-number">127</span><br></div></div><h3 id="example-6-rabbitmq-virtual-host-and-user-per-tenant" tabindex="-1">Example 6: RabbitMQ Virtual Host and User per Tenant <a class="header-anchor" href="#example-6-rabbitmq-virtual-host-and-user-per-tenant" aria-label="Permalink to &quot;Example 6: RabbitMQ Virtual Host and User per Tenant&quot;">​</a></h3><p>Provision isolated RabbitMQ resources for each tenant:</p><div class="language-yaml line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">yaml</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#85E89D;">apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">operator.kubernetes-tenants.org/v1</span></span>
<span class="line"><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">TenantTemplate</span></span>
<span class="line"><span style="color:#85E89D;">metadata</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant-with-rabbitmq</span></span>
<span class="line"><span style="color:#85E89D;">  namespace</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">default</span></span>
<span class="line"><span style="color:#85E89D;">spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  registryId</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">my-registry</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">  manifests</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">  - </span><span style="color:#85E89D;">id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">rabbitmq-resources</span></span>
<span class="line"><span style="color:#85E89D;">    nameTemplate</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-rabbitmq&quot;</span></span>
<span class="line"><span style="color:#85E89D;">    spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">      apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">infra.contrib.fluxcd.io/v1alpha2</span></span>
<span class="line"><span style="color:#85E89D;">      kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Terraform</span></span>
<span class="line"><span style="color:#85E89D;">      spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">        interval</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">5m</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        values</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          hcl</span><span style="color:#E1E4E8;">: </span><span style="color:#F97583;">|</span></span>
<span class="line"><span style="color:#9ECBFF;">            terraform {</span></span>
<span class="line"><span style="color:#9ECBFF;">              required_providers {</span></span>
<span class="line"><span style="color:#9ECBFF;">                rabbitmq = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  source  = &quot;cyrilgdn/rabbitmq&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                  version = &quot;~&gt; 1.8&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">                random = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  source  = &quot;hashicorp/random&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                  version = &quot;~&gt; 3.5&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">              backend &quot;kubernetes&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">                secret_suffix = &quot;{{ .uid }}-rabbitmq&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                namespace     = &quot;default&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            provider &quot;rabbitmq&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              endpoint = var.rabbitmq_endpoint</span></span>
<span class="line"><span style="color:#9ECBFF;">              username = var.rabbitmq_admin_user</span></span>
<span class="line"><span style="color:#9ECBFF;">              password = var.rabbitmq_admin_password</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;tenant_id&quot; { type = string }</span></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;rabbitmq_endpoint&quot; { type = string }</span></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;rabbitmq_admin_user&quot; { type = string }</span></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;rabbitmq_admin_password&quot; { type = string sensitive = true }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # Generate password for tenant user</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;random_password&quot; &quot;tenant_password&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              length  = 32</span></span>
<span class="line"><span style="color:#9ECBFF;">              special = true</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # Virtual host for tenant</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;rabbitmq_vhost&quot; &quot;tenant_vhost&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              name = &quot;tenant-\${var.tenant_id}&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # User for tenant</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;rabbitmq_user&quot; &quot;tenant_user&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              name     = &quot;tenant-\${var.tenant_id}&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              password = random_password.tenant_password.result</span></span>
<span class="line"><span style="color:#9ECBFF;">              tags     = []</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # Permissions for tenant user on their vhost</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;rabbitmq_permissions&quot; &quot;tenant_permissions&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              user  = rabbitmq_user.tenant_user.name</span></span>
<span class="line"><span style="color:#9ECBFF;">              vhost = rabbitmq_vhost.tenant_vhost.name</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              permissions {</span></span>
<span class="line"><span style="color:#9ECBFF;">                configure = &quot;.*&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                write     = &quot;.*&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                read      = &quot;.*&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # Default exchanges and queues</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;rabbitmq_exchange&quot; &quot;tenant_events&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              name  = &quot;events&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              vhost = rabbitmq_vhost.tenant_vhost.name</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              settings {</span></span>
<span class="line"><span style="color:#9ECBFF;">                type        = &quot;topic&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                durable     = true</span></span>
<span class="line"><span style="color:#9ECBFF;">                auto_delete = false</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;rabbitmq_queue&quot; &quot;tenant_tasks&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              name  = &quot;tasks&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              vhost = rabbitmq_vhost.tenant_vhost.name</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              settings {</span></span>
<span class="line"><span style="color:#9ECBFF;">                durable     = true</span></span>
<span class="line"><span style="color:#9ECBFF;">                auto_delete = false</span></span>
<span class="line"><span style="color:#9ECBFF;">                arguments = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  &quot;x-message-ttl&quot;          = 86400000  # 24 hours</span></span>
<span class="line"><span style="color:#9ECBFF;">                  &quot;x-max-length&quot;           = 10000</span></span>
<span class="line"><span style="color:#9ECBFF;">                  &quot;x-queue-type&quot;           = &quot;quorum&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;vhost&quot; { value = rabbitmq_vhost.tenant_vhost.name }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;username&quot; { value = rabbitmq_user.tenant_user.name }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;password&quot; { value = random_password.tenant_password.result sensitive = true }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;connection_string&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              value = &quot;amqp://\${rabbitmq_user.tenant_user.name}:\${random_password.tenant_password.result}@\${var.rabbitmq_endpoint}/\${rabbitmq_vhost.tenant_vhost.name}&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              sensitive = true</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        vars</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant_id</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">rabbitmq_endpoint</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;rabbitmq.default.svc.cluster.local:5672&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        varsFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Secret</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">rabbitmq-admin-credentials</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        writeOutputsToSecret</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-rabbitmq-credentials&quot;</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br><span class="line-number">12</span><br><span class="line-number">13</span><br><span class="line-number">14</span><br><span class="line-number">15</span><br><span class="line-number">16</span><br><span class="line-number">17</span><br><span class="line-number">18</span><br><span class="line-number">19</span><br><span class="line-number">20</span><br><span class="line-number">21</span><br><span class="line-number">22</span><br><span class="line-number">23</span><br><span class="line-number">24</span><br><span class="line-number">25</span><br><span class="line-number">26</span><br><span class="line-number">27</span><br><span class="line-number">28</span><br><span class="line-number">29</span><br><span class="line-number">30</span><br><span class="line-number">31</span><br><span class="line-number">32</span><br><span class="line-number">33</span><br><span class="line-number">34</span><br><span class="line-number">35</span><br><span class="line-number">36</span><br><span class="line-number">37</span><br><span class="line-number">38</span><br><span class="line-number">39</span><br><span class="line-number">40</span><br><span class="line-number">41</span><br><span class="line-number">42</span><br><span class="line-number">43</span><br><span class="line-number">44</span><br><span class="line-number">45</span><br><span class="line-number">46</span><br><span class="line-number">47</span><br><span class="line-number">48</span><br><span class="line-number">49</span><br><span class="line-number">50</span><br><span class="line-number">51</span><br><span class="line-number">52</span><br><span class="line-number">53</span><br><span class="line-number">54</span><br><span class="line-number">55</span><br><span class="line-number">56</span><br><span class="line-number">57</span><br><span class="line-number">58</span><br><span class="line-number">59</span><br><span class="line-number">60</span><br><span class="line-number">61</span><br><span class="line-number">62</span><br><span class="line-number">63</span><br><span class="line-number">64</span><br><span class="line-number">65</span><br><span class="line-number">66</span><br><span class="line-number">67</span><br><span class="line-number">68</span><br><span class="line-number">69</span><br><span class="line-number">70</span><br><span class="line-number">71</span><br><span class="line-number">72</span><br><span class="line-number">73</span><br><span class="line-number">74</span><br><span class="line-number">75</span><br><span class="line-number">76</span><br><span class="line-number">77</span><br><span class="line-number">78</span><br><span class="line-number">79</span><br><span class="line-number">80</span><br><span class="line-number">81</span><br><span class="line-number">82</span><br><span class="line-number">83</span><br><span class="line-number">84</span><br><span class="line-number">85</span><br><span class="line-number">86</span><br><span class="line-number">87</span><br><span class="line-number">88</span><br><span class="line-number">89</span><br><span class="line-number">90</span><br><span class="line-number">91</span><br><span class="line-number">92</span><br><span class="line-number">93</span><br><span class="line-number">94</span><br><span class="line-number">95</span><br><span class="line-number">96</span><br><span class="line-number">97</span><br><span class="line-number">98</span><br><span class="line-number">99</span><br><span class="line-number">100</span><br><span class="line-number">101</span><br><span class="line-number">102</span><br><span class="line-number">103</span><br><span class="line-number">104</span><br><span class="line-number">105</span><br><span class="line-number">106</span><br><span class="line-number">107</span><br><span class="line-number">108</span><br><span class="line-number">109</span><br><span class="line-number">110</span><br><span class="line-number">111</span><br><span class="line-number">112</span><br><span class="line-number">113</span><br><span class="line-number">114</span><br><span class="line-number">115</span><br><span class="line-number">116</span><br><span class="line-number">117</span><br><span class="line-number">118</span><br><span class="line-number">119</span><br><span class="line-number">120</span><br><span class="line-number">121</span><br><span class="line-number">122</span><br><span class="line-number">123</span><br><span class="line-number">124</span><br></div></div><h3 id="example-7-postgresql-schema-and-user-per-tenant" tabindex="-1">Example 7: PostgreSQL Schema and User per Tenant <a class="header-anchor" href="#example-7-postgresql-schema-and-user-per-tenant" aria-label="Permalink to &quot;Example 7: PostgreSQL Schema and User per Tenant&quot;">​</a></h3><p>Provision isolated PostgreSQL schemas in a shared database:</p><div class="language-yaml line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">yaml</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#85E89D;">apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">operator.kubernetes-tenants.org/v1</span></span>
<span class="line"><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">TenantTemplate</span></span>
<span class="line"><span style="color:#85E89D;">metadata</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant-with-pg-schema</span></span>
<span class="line"><span style="color:#85E89D;">  namespace</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">default</span></span>
<span class="line"><span style="color:#85E89D;">spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  registryId</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">my-registry</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">  manifests</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">  - </span><span style="color:#85E89D;">id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">postgresql-schema</span></span>
<span class="line"><span style="color:#85E89D;">    nameTemplate</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-pg-schema&quot;</span></span>
<span class="line"><span style="color:#85E89D;">    creationPolicy</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Once</span></span>
<span class="line"><span style="color:#85E89D;">    deletionPolicy</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Retain</span></span>
<span class="line"><span style="color:#85E89D;">    spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">      apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">infra.contrib.fluxcd.io/v1alpha2</span></span>
<span class="line"><span style="color:#85E89D;">      kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Terraform</span></span>
<span class="line"><span style="color:#85E89D;">      spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">        interval</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">5m</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        values</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          hcl</span><span style="color:#E1E4E8;">: </span><span style="color:#F97583;">|</span></span>
<span class="line"><span style="color:#9ECBFF;">            terraform {</span></span>
<span class="line"><span style="color:#9ECBFF;">              required_providers {</span></span>
<span class="line"><span style="color:#9ECBFF;">                postgresql = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  source  = &quot;cyrilgdn/postgresql&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                  version = &quot;~&gt; 1.21&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">                random = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  source  = &quot;hashicorp/random&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                  version = &quot;~&gt; 3.5&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">              backend &quot;kubernetes&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">                secret_suffix = &quot;{{ .uid }}-pg&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                namespace     = &quot;default&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            provider &quot;postgresql&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              host            = var.pg_host</span></span>
<span class="line"><span style="color:#9ECBFF;">              port            = var.pg_port</span></span>
<span class="line"><span style="color:#9ECBFF;">              database        = var.pg_database</span></span>
<span class="line"><span style="color:#9ECBFF;">              username        = var.pg_admin_user</span></span>
<span class="line"><span style="color:#9ECBFF;">              password        = var.pg_admin_password</span></span>
<span class="line"><span style="color:#9ECBFF;">              sslmode         = &quot;require&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              connect_timeout = 15</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;tenant_id&quot; { type = string }</span></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;pg_host&quot; { type = string }</span></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;pg_port&quot; { type = number default = 5432 }</span></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;pg_database&quot; { type = string }</span></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;pg_admin_user&quot; { type = string }</span></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;pg_admin_password&quot; { type = string sensitive = true }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # Generate password for tenant</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;random_password&quot; &quot;tenant_password&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              length  = 32</span></span>
<span class="line"><span style="color:#9ECBFF;">              special = true</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # Schema for tenant</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;postgresql_schema&quot; &quot;tenant_schema&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              name  = &quot;tenant_\${replace(var.tenant_id, &quot;-&quot;, &quot;_&quot;)}&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              owner = postgresql_role.tenant_user.name</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # User/Role for tenant</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;postgresql_role&quot; &quot;tenant_user&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              name     = &quot;tenant_\${replace(var.tenant_id, &quot;-&quot;, &quot;_&quot;)}&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              login    = true</span></span>
<span class="line"><span style="color:#9ECBFF;">              password = random_password.tenant_password.result</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # Grant schema usage to tenant user</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;postgresql_grant&quot; &quot;schema_usage&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              database    = var.pg_database</span></span>
<span class="line"><span style="color:#9ECBFF;">              role        = postgresql_role.tenant_user.name</span></span>
<span class="line"><span style="color:#9ECBFF;">              schema      = postgresql_schema.tenant_schema.name</span></span>
<span class="line"><span style="color:#9ECBFF;">              object_type = &quot;schema&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              privileges  = [&quot;USAGE&quot;, &quot;CREATE&quot;]</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # Grant all privileges on tables in schema</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;postgresql_grant&quot; &quot;tables&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              database    = var.pg_database</span></span>
<span class="line"><span style="color:#9ECBFF;">              role        = postgresql_role.tenant_user.name</span></span>
<span class="line"><span style="color:#9ECBFF;">              schema      = postgresql_schema.tenant_schema.name</span></span>
<span class="line"><span style="color:#9ECBFF;">              object_type = &quot;table&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              privileges  = [&quot;SELECT&quot;, &quot;INSERT&quot;, &quot;UPDATE&quot;, &quot;DELETE&quot;]</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;schema_name&quot; { value = postgresql_schema.tenant_schema.name }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;db_user&quot; { value = postgresql_role.tenant_user.name }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;db_password&quot; { value = random_password.tenant_password.result sensitive = true }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;connection_string&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              value = &quot;postgresql://\${postgresql_role.tenant_user.name}:\${random_password.tenant_password.result}@\${var.pg_host}:\${var.pg_port}/\${var.pg_database}?options=-c%20search_path%3D\${postgresql_schema.tenant_schema.name}&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              sensitive = true</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        vars</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant_id</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">pg_host</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;postgres.default.svc.cluster.local&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">pg_database</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;tenants&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        varsFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Secret</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">postgres-admin-credentials</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        writeOutputsToSecret</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-postgres-credentials&quot;</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br><span class="line-number">12</span><br><span class="line-number">13</span><br><span class="line-number">14</span><br><span class="line-number">15</span><br><span class="line-number">16</span><br><span class="line-number">17</span><br><span class="line-number">18</span><br><span class="line-number">19</span><br><span class="line-number">20</span><br><span class="line-number">21</span><br><span class="line-number">22</span><br><span class="line-number">23</span><br><span class="line-number">24</span><br><span class="line-number">25</span><br><span class="line-number">26</span><br><span class="line-number">27</span><br><span class="line-number">28</span><br><span class="line-number">29</span><br><span class="line-number">30</span><br><span class="line-number">31</span><br><span class="line-number">32</span><br><span class="line-number">33</span><br><span class="line-number">34</span><br><span class="line-number">35</span><br><span class="line-number">36</span><br><span class="line-number">37</span><br><span class="line-number">38</span><br><span class="line-number">39</span><br><span class="line-number">40</span><br><span class="line-number">41</span><br><span class="line-number">42</span><br><span class="line-number">43</span><br><span class="line-number">44</span><br><span class="line-number">45</span><br><span class="line-number">46</span><br><span class="line-number">47</span><br><span class="line-number">48</span><br><span class="line-number">49</span><br><span class="line-number">50</span><br><span class="line-number">51</span><br><span class="line-number">52</span><br><span class="line-number">53</span><br><span class="line-number">54</span><br><span class="line-number">55</span><br><span class="line-number">56</span><br><span class="line-number">57</span><br><span class="line-number">58</span><br><span class="line-number">59</span><br><span class="line-number">60</span><br><span class="line-number">61</span><br><span class="line-number">62</span><br><span class="line-number">63</span><br><span class="line-number">64</span><br><span class="line-number">65</span><br><span class="line-number">66</span><br><span class="line-number">67</span><br><span class="line-number">68</span><br><span class="line-number">69</span><br><span class="line-number">70</span><br><span class="line-number">71</span><br><span class="line-number">72</span><br><span class="line-number">73</span><br><span class="line-number">74</span><br><span class="line-number">75</span><br><span class="line-number">76</span><br><span class="line-number">77</span><br><span class="line-number">78</span><br><span class="line-number">79</span><br><span class="line-number">80</span><br><span class="line-number">81</span><br><span class="line-number">82</span><br><span class="line-number">83</span><br><span class="line-number">84</span><br><span class="line-number">85</span><br><span class="line-number">86</span><br><span class="line-number">87</span><br><span class="line-number">88</span><br><span class="line-number">89</span><br><span class="line-number">90</span><br><span class="line-number">91</span><br><span class="line-number">92</span><br><span class="line-number">93</span><br><span class="line-number">94</span><br><span class="line-number">95</span><br><span class="line-number">96</span><br><span class="line-number">97</span><br><span class="line-number">98</span><br><span class="line-number">99</span><br><span class="line-number">100</span><br><span class="line-number">101</span><br><span class="line-number">102</span><br><span class="line-number">103</span><br><span class="line-number">104</span><br><span class="line-number">105</span><br><span class="line-number">106</span><br><span class="line-number">107</span><br><span class="line-number">108</span><br><span class="line-number">109</span><br><span class="line-number">110</span><br><span class="line-number">111</span><br><span class="line-number">112</span><br><span class="line-number">113</span><br><span class="line-number">114</span><br></div></div><h3 id="example-8-redis-database-per-tenant" tabindex="-1">Example 8: Redis Database per Tenant <a class="header-anchor" href="#example-8-redis-database-per-tenant" aria-label="Permalink to &quot;Example 8: Redis Database per Tenant&quot;">​</a></h3><p>Provision dedicated Redis database numbers for each tenant:</p><div class="language-yaml line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">yaml</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#85E89D;">apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">operator.kubernetes-tenants.org/v1</span></span>
<span class="line"><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">TenantTemplate</span></span>
<span class="line"><span style="color:#85E89D;">metadata</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant-with-redis</span></span>
<span class="line"><span style="color:#85E89D;">  namespace</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">default</span></span>
<span class="line"><span style="color:#85E89D;">spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  registryId</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">my-registry</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">  manifests</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">  - </span><span style="color:#85E89D;">id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">redis-database</span></span>
<span class="line"><span style="color:#85E89D;">    nameTemplate</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-redis&quot;</span></span>
<span class="line"><span style="color:#85E89D;">    spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">      apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">infra.contrib.fluxcd.io/v1alpha2</span></span>
<span class="line"><span style="color:#85E89D;">      kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Terraform</span></span>
<span class="line"><span style="color:#85E89D;">      spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">        interval</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">5m</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        values</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          hcl</span><span style="color:#E1E4E8;">: </span><span style="color:#F97583;">|</span></span>
<span class="line"><span style="color:#9ECBFF;">            terraform {</span></span>
<span class="line"><span style="color:#9ECBFF;">              required_providers {</span></span>
<span class="line"><span style="color:#9ECBFF;">                redis = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  source  = &quot;redis/redis&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                  version = &quot;~&gt; 1.3&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">              backend &quot;kubernetes&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">                secret_suffix = &quot;{{ .uid }}-redis&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                namespace     = &quot;default&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            provider &quot;redis&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              address = var.redis_address</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;tenant_id&quot; { type = string }</span></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;redis_address&quot; { type = string }</span></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;redis_db_number&quot; { type = number }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # Note: Redis doesn&#39;t have native ACLs for DB numbers in older versions</span></span>
<span class="line"><span style="color:#9ECBFF;">            # This example shows configuration; actual implementation may vary</span></span>
<span class="line"><span style="color:#9ECBFF;">            # For Redis 6+, use ACLs instead</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            locals {</span></span>
<span class="line"><span style="color:#9ECBFF;">              db_number = var.redis_db_number</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;redis_host&quot; { value = split(&quot;:&quot;, var.redis_address)[0] }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;redis_port&quot; { value = split(&quot;:&quot;, var.redis_address)[1] }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;redis_db&quot; { value = local.db_number }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;connection_string&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              value = &quot;redis://\${var.redis_address}/\${local.db_number}&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        vars</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant_id</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">redis_address</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;redis.default.svc.cluster.local:6379&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">redis_db_number</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid | sha1sum | trunc 2 }}&quot;</span><span style="color:#6A737D;">  # Generate DB number from tenant ID</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        writeOutputsToSecret</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-redis-config&quot;</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br><span class="line-number">12</span><br><span class="line-number">13</span><br><span class="line-number">14</span><br><span class="line-number">15</span><br><span class="line-number">16</span><br><span class="line-number">17</span><br><span class="line-number">18</span><br><span class="line-number">19</span><br><span class="line-number">20</span><br><span class="line-number">21</span><br><span class="line-number">22</span><br><span class="line-number">23</span><br><span class="line-number">24</span><br><span class="line-number">25</span><br><span class="line-number">26</span><br><span class="line-number">27</span><br><span class="line-number">28</span><br><span class="line-number">29</span><br><span class="line-number">30</span><br><span class="line-number">31</span><br><span class="line-number">32</span><br><span class="line-number">33</span><br><span class="line-number">34</span><br><span class="line-number">35</span><br><span class="line-number">36</span><br><span class="line-number">37</span><br><span class="line-number">38</span><br><span class="line-number">39</span><br><span class="line-number">40</span><br><span class="line-number">41</span><br><span class="line-number">42</span><br><span class="line-number">43</span><br><span class="line-number">44</span><br><span class="line-number">45</span><br><span class="line-number">46</span><br><span class="line-number">47</span><br><span class="line-number">48</span><br><span class="line-number">49</span><br><span class="line-number">50</span><br><span class="line-number">51</span><br><span class="line-number">52</span><br><span class="line-number">53</span><br><span class="line-number">54</span><br><span class="line-number">55</span><br><span class="line-number">56</span><br><span class="line-number">57</span><br><span class="line-number">58</span><br><span class="line-number">59</span><br><span class="line-number">60</span><br><span class="line-number">61</span><br><span class="line-number">62</span><br><span class="line-number">63</span><br><span class="line-number">64</span><br><span class="line-number">65</span><br></div></div><h2 id="complete-multi-resource-example" tabindex="-1">Complete Multi-Resource Example <a class="header-anchor" href="#complete-multi-resource-example" aria-label="Permalink to &quot;Complete Multi-Resource Example&quot;">​</a></h2><p>Full example provisioning S3, RDS, and CloudFront for each tenant:</p><div class="language-yaml line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">yaml</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#85E89D;">apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">operator.kubernetes-tenants.org/v1</span></span>
<span class="line"><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">TenantTemplate</span></span>
<span class="line"><span style="color:#85E89D;">metadata</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">enterprise-tenant-stack</span></span>
<span class="line"><span style="color:#85E89D;">  namespace</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">default</span></span>
<span class="line"><span style="color:#85E89D;">spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  registryId</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">enterprise-registry</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">  # Terraform for complete infrastructure stack</span></span>
<span class="line"><span style="color:#85E89D;">  manifests</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">  - </span><span style="color:#85E89D;">id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant-infrastructure</span></span>
<span class="line"><span style="color:#85E89D;">    nameTemplate</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-infrastructure&quot;</span></span>
<span class="line"><span style="color:#85E89D;">    creationPolicy</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Once</span></span>
<span class="line"><span style="color:#85E89D;">    deletionPolicy</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Retain</span></span>
<span class="line"><span style="color:#85E89D;">    timeoutSeconds</span><span style="color:#E1E4E8;">: </span><span style="color:#79B8FF;">1800</span><span style="color:#6A737D;">  # 30 minutes</span></span>
<span class="line"><span style="color:#85E89D;">    spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">      apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">infra.contrib.fluxcd.io/v1alpha2</span></span>
<span class="line"><span style="color:#85E89D;">      kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Terraform</span></span>
<span class="line"><span style="color:#85E89D;">      spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">        interval</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">15m</span></span>
<span class="line"><span style="color:#85E89D;">        retryInterval</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">2m</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        values</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          hcl</span><span style="color:#E1E4E8;">: </span><span style="color:#F97583;">|</span></span>
<span class="line"><span style="color:#9ECBFF;">            terraform {</span></span>
<span class="line"><span style="color:#9ECBFF;">              required_providers {</span></span>
<span class="line"><span style="color:#9ECBFF;">                aws = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  source  = &quot;hashicorp/aws&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                  version = &quot;~&gt; 5.0&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">                random = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  source  = &quot;hashicorp/random&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                  version = &quot;~&gt; 3.5&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">              backend &quot;kubernetes&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">                secret_suffix = &quot;{{ .uid }}-infra&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                namespace     = &quot;default&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            provider &quot;aws&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              region = var.aws_region</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;tenant_id&quot; { type = string }</span></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;tenant_host&quot; { type = string }</span></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;aws_region&quot; { type = string }</span></span>
<span class="line"><span style="color:#9ECBFF;">            variable &quot;db_instance_class&quot; { type = string }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # Random password for database</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;random_password&quot; &quot;db_password&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              length  = 32</span></span>
<span class="line"><span style="color:#9ECBFF;">              special = true</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # S3 bucket for tenant data</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;aws_s3_bucket&quot; &quot;tenant_data&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              bucket = &quot;tenant-\${var.tenant_id}-data&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              tags = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                TenantId = var.tenant_id</span></span>
<span class="line"><span style="color:#9ECBFF;">                Purpose  = &quot;tenant-data&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;aws_s3_bucket_versioning&quot; &quot;tenant_data_versioning&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              bucket = aws_s3_bucket.tenant_data.id</span></span>
<span class="line"><span style="color:#9ECBFF;">              versioning_configuration {</span></span>
<span class="line"><span style="color:#9ECBFF;">                status = &quot;Enabled&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # S3 bucket for static assets</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;aws_s3_bucket&quot; &quot;tenant_static&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              bucket = &quot;tenant-\${var.tenant_id}-static&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              tags = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                TenantId = var.tenant_id</span></span>
<span class="line"><span style="color:#9ECBFF;">                Purpose  = &quot;static-assets&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;aws_s3_bucket_public_access_block&quot; &quot;tenant_static_pab&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              bucket = aws_s3_bucket.tenant_static.id</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              block_public_acls       = false</span></span>
<span class="line"><span style="color:#9ECBFF;">              block_public_policy     = false</span></span>
<span class="line"><span style="color:#9ECBFF;">              ignore_public_acls      = false</span></span>
<span class="line"><span style="color:#9ECBFF;">              restrict_public_buckets = false</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # RDS PostgreSQL</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;aws_db_instance&quot; &quot;tenant_db&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              identifier            = &quot;tenant-\${var.tenant_id}-db&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              engine                = &quot;postgres&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              engine_version        = &quot;15.4&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              instance_class        = var.db_instance_class</span></span>
<span class="line"><span style="color:#9ECBFF;">              allocated_storage     = 20</span></span>
<span class="line"><span style="color:#9ECBFF;">              storage_encrypted     = true</span></span>
<span class="line"><span style="color:#9ECBFF;">              db_name               = &quot;tenant_\${replace(var.tenant_id, &quot;-&quot;, &quot;_&quot;)}&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              username              = &quot;dbadmin&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              password              = random_password.db_password.result</span></span>
<span class="line"><span style="color:#9ECBFF;">              skip_final_snapshot   = false</span></span>
<span class="line"><span style="color:#9ECBFF;">              final_snapshot_identifier = &quot;tenant-\${var.tenant_id}-final&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              tags = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                TenantId = var.tenant_id</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # CloudFront distribution</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;aws_cloudfront_distribution&quot; &quot;tenant_cdn&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              enabled         = true</span></span>
<span class="line"><span style="color:#9ECBFF;">              is_ipv6_enabled = true</span></span>
<span class="line"><span style="color:#9ECBFF;">              comment         = &quot;CDN for \${var.tenant_id}&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              origin {</span></span>
<span class="line"><span style="color:#9ECBFF;">                domain_name = aws_s3_bucket.tenant_static.bucket_regional_domain_name</span></span>
<span class="line"><span style="color:#9ECBFF;">                origin_id   = &quot;S3-\${var.tenant_id}&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              default_cache_behavior {</span></span>
<span class="line"><span style="color:#9ECBFF;">                allowed_methods        = [&quot;GET&quot;, &quot;HEAD&quot;]</span></span>
<span class="line"><span style="color:#9ECBFF;">                cached_methods         = [&quot;GET&quot;, &quot;HEAD&quot;]</span></span>
<span class="line"><span style="color:#9ECBFF;">                target_origin_id       = &quot;S3-\${var.tenant_id}&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                viewer_protocol_policy = &quot;redirect-to-https&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">                forwarded_values {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  query_string = false</span></span>
<span class="line"><span style="color:#9ECBFF;">                  cookies {</span></span>
<span class="line"><span style="color:#9ECBFF;">                    forward = &quot;none&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                  }</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              restrictions {</span></span>
<span class="line"><span style="color:#9ECBFF;">                geo_restriction {</span></span>
<span class="line"><span style="color:#9ECBFF;">                  restriction_type = &quot;none&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                }</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              viewer_certificate {</span></span>
<span class="line"><span style="color:#9ECBFF;">                cloudfront_default_certificate = true</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              tags = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                TenantId = var.tenant_id</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # IAM user for tenant access</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;aws_iam_user&quot; &quot;tenant_user&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              name = &quot;tenant-\${var.tenant_id}-user&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              tags = {</span></span>
<span class="line"><span style="color:#9ECBFF;">                TenantId = var.tenant_id</span></span>
<span class="line"><span style="color:#9ECBFF;">              }</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;aws_iam_access_key&quot; &quot;tenant_access_key&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              user = aws_iam_user.tenant_user.name</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # IAM policy for tenant S3 access</span></span>
<span class="line"><span style="color:#9ECBFF;">            resource &quot;aws_iam_user_policy&quot; &quot;tenant_s3_policy&quot; {</span></span>
<span class="line"><span style="color:#9ECBFF;">              name = &quot;tenant-\${var.tenant_id}-s3-policy&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">              user = aws_iam_user.tenant_user.name</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">              policy = jsonencode({</span></span>
<span class="line"><span style="color:#9ECBFF;">                Version = &quot;2012-10-17&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                Statement = [</span></span>
<span class="line"><span style="color:#9ECBFF;">                  {</span></span>
<span class="line"><span style="color:#9ECBFF;">                    Effect = &quot;Allow&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                    Action = [</span></span>
<span class="line"><span style="color:#9ECBFF;">                      &quot;s3:GetObject&quot;,</span></span>
<span class="line"><span style="color:#9ECBFF;">                      &quot;s3:PutObject&quot;,</span></span>
<span class="line"><span style="color:#9ECBFF;">                      &quot;s3:DeleteObject&quot;,</span></span>
<span class="line"><span style="color:#9ECBFF;">                      &quot;s3:ListBucket&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                    ]</span></span>
<span class="line"><span style="color:#9ECBFF;">                    Resource = [</span></span>
<span class="line"><span style="color:#9ECBFF;">                      aws_s3_bucket.tenant_data.arn,</span></span>
<span class="line"><span style="color:#9ECBFF;">                      &quot;\${aws_s3_bucket.tenant_data.arn}/*&quot;</span></span>
<span class="line"><span style="color:#9ECBFF;">                    ]</span></span>
<span class="line"><span style="color:#9ECBFF;">                  }</span></span>
<span class="line"><span style="color:#9ECBFF;">                ]</span></span>
<span class="line"><span style="color:#9ECBFF;">              })</span></span>
<span class="line"><span style="color:#9ECBFF;">            }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9ECBFF;">            # Outputs</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;s3_data_bucket&quot; { value = aws_s3_bucket.tenant_data.id }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;s3_static_bucket&quot; { value = aws_s3_bucket.tenant_static.id }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;db_endpoint&quot; { value = aws_db_instance.tenant_db.endpoint }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;db_name&quot; { value = aws_db_instance.tenant_db.db_name }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;db_username&quot; { value = aws_db_instance.tenant_db.username }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;db_password&quot; { value = random_password.db_password.result sensitive = true }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;cdn_domain&quot; { value = aws_cloudfront_distribution.tenant_cdn.domain_name }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;cdn_distribution_id&quot; { value = aws_cloudfront_distribution.tenant_cdn.id }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;iam_access_key_id&quot; { value = aws_iam_access_key.tenant_access_key.id }</span></span>
<span class="line"><span style="color:#9ECBFF;">            output &quot;iam_secret_access_key&quot; { value = aws_iam_access_key.tenant_access_key.secret sensitive = true }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        vars</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant_id</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">tenant_host</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .host }}&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">aws_region</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;us-east-1&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">db_instance_class</span></span>
<span class="line"><span style="color:#85E89D;">          value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;db.t3.micro&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        varsFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">        - </span><span style="color:#85E89D;">kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Secret</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">aws-credentials</span></span>
<span class="line"></span>
<span class="line"><span style="color:#85E89D;">        writeOutputsToSecret</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-infrastructure&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">  # ConfigMap with infrastructure info</span></span>
<span class="line"><span style="color:#85E89D;">  configMaps</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">  - </span><span style="color:#85E89D;">id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">infra-config</span></span>
<span class="line"><span style="color:#85E89D;">    nameTemplate</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-infra-config&quot;</span></span>
<span class="line"><span style="color:#85E89D;">    dependIds</span><span style="color:#E1E4E8;">: [</span><span style="color:#9ECBFF;">&quot;tenant-infrastructure&quot;</span><span style="color:#E1E4E8;">]</span></span>
<span class="line"><span style="color:#85E89D;">    spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">      apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">v1</span></span>
<span class="line"><span style="color:#85E89D;">      kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">ConfigMap</span></span>
<span class="line"><span style="color:#85E89D;">      data</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">        tenant_id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#85E89D;">        terraform_outputs_secret</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-infrastructure&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">  # Application deployment</span></span>
<span class="line"><span style="color:#85E89D;">  deployments</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">  - </span><span style="color:#85E89D;">id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">app-deploy</span></span>
<span class="line"><span style="color:#85E89D;">    nameTemplate</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-app&quot;</span></span>
<span class="line"><span style="color:#85E89D;">    dependIds</span><span style="color:#E1E4E8;">: [</span><span style="color:#9ECBFF;">&quot;tenant-infrastructure&quot;</span><span style="color:#E1E4E8;">]</span></span>
<span class="line"><span style="color:#85E89D;">    spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">      apiVersion</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">apps/v1</span></span>
<span class="line"><span style="color:#85E89D;">      kind</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Deployment</span></span>
<span class="line"><span style="color:#85E89D;">      spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">        replicas</span><span style="color:#E1E4E8;">: </span><span style="color:#79B8FF;">2</span></span>
<span class="line"><span style="color:#85E89D;">        selector</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          matchLabels</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">            app</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#85E89D;">        template</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">          metadata</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">            labels</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">              app</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span>
<span class="line"><span style="color:#85E89D;">          spec</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">            containers</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">            - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">app</span></span>
<span class="line"><span style="color:#85E89D;">              image</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">mycompany/enterprise-app:latest</span></span>
<span class="line"><span style="color:#85E89D;">              env</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#6A737D;">              # Database connection</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">DB_HOST</span></span>
<span class="line"><span style="color:#85E89D;">                valueFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                  secretKeyRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                    name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-infrastructure&quot;</span></span>
<span class="line"><span style="color:#85E89D;">                    key</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">db_endpoint</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">DB_NAME</span></span>
<span class="line"><span style="color:#85E89D;">                valueFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                  secretKeyRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                    name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-infrastructure&quot;</span></span>
<span class="line"><span style="color:#85E89D;">                    key</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">db_name</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">DB_USER</span></span>
<span class="line"><span style="color:#85E89D;">                valueFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                  secretKeyRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                    name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-infrastructure&quot;</span></span>
<span class="line"><span style="color:#85E89D;">                    key</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">db_username</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">DB_PASSWORD</span></span>
<span class="line"><span style="color:#85E89D;">                valueFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                  secretKeyRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                    name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-infrastructure&quot;</span></span>
<span class="line"><span style="color:#85E89D;">                    key</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">db_password</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">              # S3 buckets</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">S3_DATA_BUCKET</span></span>
<span class="line"><span style="color:#85E89D;">                valueFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                  secretKeyRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                    name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-infrastructure&quot;</span></span>
<span class="line"><span style="color:#85E89D;">                    key</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">s3_data_bucket</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">S3_STATIC_BUCKET</span></span>
<span class="line"><span style="color:#85E89D;">                valueFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                  secretKeyRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                    name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-infrastructure&quot;</span></span>
<span class="line"><span style="color:#85E89D;">                    key</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">s3_static_bucket</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">              # CloudFront CDN</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">CDN_DOMAIN</span></span>
<span class="line"><span style="color:#85E89D;">                valueFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                  secretKeyRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                    name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-infrastructure&quot;</span></span>
<span class="line"><span style="color:#85E89D;">                    key</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">cdn_domain</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;">              # IAM credentials</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">AWS_ACCESS_KEY_ID</span></span>
<span class="line"><span style="color:#85E89D;">                valueFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                  secretKeyRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                    name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-infrastructure&quot;</span></span>
<span class="line"><span style="color:#85E89D;">                    key</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">iam_access_key_id</span></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">AWS_SECRET_ACCESS_KEY</span></span>
<span class="line"><span style="color:#85E89D;">                valueFrom</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                  secretKeyRef</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">                    name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-infrastructure&quot;</span></span>
<span class="line"><span style="color:#85E89D;">                    key</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">iam_secret_access_key</span></span>
<span class="line"></span>
<span class="line"><span style="color:#E1E4E8;">              - </span><span style="color:#85E89D;">name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">TENANT_ID</span></span>
<span class="line"><span style="color:#85E89D;">                value</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}&quot;</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br><span class="line-number">12</span><br><span class="line-number">13</span><br><span class="line-number">14</span><br><span class="line-number">15</span><br><span class="line-number">16</span><br><span class="line-number">17</span><br><span class="line-number">18</span><br><span class="line-number">19</span><br><span class="line-number">20</span><br><span class="line-number">21</span><br><span class="line-number">22</span><br><span class="line-number">23</span><br><span class="line-number">24</span><br><span class="line-number">25</span><br><span class="line-number">26</span><br><span class="line-number">27</span><br><span class="line-number">28</span><br><span class="line-number">29</span><br><span class="line-number">30</span><br><span class="line-number">31</span><br><span class="line-number">32</span><br><span class="line-number">33</span><br><span class="line-number">34</span><br><span class="line-number">35</span><br><span class="line-number">36</span><br><span class="line-number">37</span><br><span class="line-number">38</span><br><span class="line-number">39</span><br><span class="line-number">40</span><br><span class="line-number">41</span><br><span class="line-number">42</span><br><span class="line-number">43</span><br><span class="line-number">44</span><br><span class="line-number">45</span><br><span class="line-number">46</span><br><span class="line-number">47</span><br><span class="line-number">48</span><br><span class="line-number">49</span><br><span class="line-number">50</span><br><span class="line-number">51</span><br><span class="line-number">52</span><br><span class="line-number">53</span><br><span class="line-number">54</span><br><span class="line-number">55</span><br><span class="line-number">56</span><br><span class="line-number">57</span><br><span class="line-number">58</span><br><span class="line-number">59</span><br><span class="line-number">60</span><br><span class="line-number">61</span><br><span class="line-number">62</span><br><span class="line-number">63</span><br><span class="line-number">64</span><br><span class="line-number">65</span><br><span class="line-number">66</span><br><span class="line-number">67</span><br><span class="line-number">68</span><br><span class="line-number">69</span><br><span class="line-number">70</span><br><span class="line-number">71</span><br><span class="line-number">72</span><br><span class="line-number">73</span><br><span class="line-number">74</span><br><span class="line-number">75</span><br><span class="line-number">76</span><br><span class="line-number">77</span><br><span class="line-number">78</span><br><span class="line-number">79</span><br><span class="line-number">80</span><br><span class="line-number">81</span><br><span class="line-number">82</span><br><span class="line-number">83</span><br><span class="line-number">84</span><br><span class="line-number">85</span><br><span class="line-number">86</span><br><span class="line-number">87</span><br><span class="line-number">88</span><br><span class="line-number">89</span><br><span class="line-number">90</span><br><span class="line-number">91</span><br><span class="line-number">92</span><br><span class="line-number">93</span><br><span class="line-number">94</span><br><span class="line-number">95</span><br><span class="line-number">96</span><br><span class="line-number">97</span><br><span class="line-number">98</span><br><span class="line-number">99</span><br><span class="line-number">100</span><br><span class="line-number">101</span><br><span class="line-number">102</span><br><span class="line-number">103</span><br><span class="line-number">104</span><br><span class="line-number">105</span><br><span class="line-number">106</span><br><span class="line-number">107</span><br><span class="line-number">108</span><br><span class="line-number">109</span><br><span class="line-number">110</span><br><span class="line-number">111</span><br><span class="line-number">112</span><br><span class="line-number">113</span><br><span class="line-number">114</span><br><span class="line-number">115</span><br><span class="line-number">116</span><br><span class="line-number">117</span><br><span class="line-number">118</span><br><span class="line-number">119</span><br><span class="line-number">120</span><br><span class="line-number">121</span><br><span class="line-number">122</span><br><span class="line-number">123</span><br><span class="line-number">124</span><br><span class="line-number">125</span><br><span class="line-number">126</span><br><span class="line-number">127</span><br><span class="line-number">128</span><br><span class="line-number">129</span><br><span class="line-number">130</span><br><span class="line-number">131</span><br><span class="line-number">132</span><br><span class="line-number">133</span><br><span class="line-number">134</span><br><span class="line-number">135</span><br><span class="line-number">136</span><br><span class="line-number">137</span><br><span class="line-number">138</span><br><span class="line-number">139</span><br><span class="line-number">140</span><br><span class="line-number">141</span><br><span class="line-number">142</span><br><span class="line-number">143</span><br><span class="line-number">144</span><br><span class="line-number">145</span><br><span class="line-number">146</span><br><span class="line-number">147</span><br><span class="line-number">148</span><br><span class="line-number">149</span><br><span class="line-number">150</span><br><span class="line-number">151</span><br><span class="line-number">152</span><br><span class="line-number">153</span><br><span class="line-number">154</span><br><span class="line-number">155</span><br><span class="line-number">156</span><br><span class="line-number">157</span><br><span class="line-number">158</span><br><span class="line-number">159</span><br><span class="line-number">160</span><br><span class="line-number">161</span><br><span class="line-number">162</span><br><span class="line-number">163</span><br><span class="line-number">164</span><br><span class="line-number">165</span><br><span class="line-number">166</span><br><span class="line-number">167</span><br><span class="line-number">168</span><br><span class="line-number">169</span><br><span class="line-number">170</span><br><span class="line-number">171</span><br><span class="line-number">172</span><br><span class="line-number">173</span><br><span class="line-number">174</span><br><span class="line-number">175</span><br><span class="line-number">176</span><br><span class="line-number">177</span><br><span class="line-number">178</span><br><span class="line-number">179</span><br><span class="line-number">180</span><br><span class="line-number">181</span><br><span class="line-number">182</span><br><span class="line-number">183</span><br><span class="line-number">184</span><br><span class="line-number">185</span><br><span class="line-number">186</span><br><span class="line-number">187</span><br><span class="line-number">188</span><br><span class="line-number">189</span><br><span class="line-number">190</span><br><span class="line-number">191</span><br><span class="line-number">192</span><br><span class="line-number">193</span><br><span class="line-number">194</span><br><span class="line-number">195</span><br><span class="line-number">196</span><br><span class="line-number">197</span><br><span class="line-number">198</span><br><span class="line-number">199</span><br><span class="line-number">200</span><br><span class="line-number">201</span><br><span class="line-number">202</span><br><span class="line-number">203</span><br><span class="line-number">204</span><br><span class="line-number">205</span><br><span class="line-number">206</span><br><span class="line-number">207</span><br><span class="line-number">208</span><br><span class="line-number">209</span><br><span class="line-number">210</span><br><span class="line-number">211</span><br><span class="line-number">212</span><br><span class="line-number">213</span><br><span class="line-number">214</span><br><span class="line-number">215</span><br><span class="line-number">216</span><br><span class="line-number">217</span><br><span class="line-number">218</span><br><span class="line-number">219</span><br><span class="line-number">220</span><br><span class="line-number">221</span><br><span class="line-number">222</span><br><span class="line-number">223</span><br><span class="line-number">224</span><br><span class="line-number">225</span><br><span class="line-number">226</span><br><span class="line-number">227</span><br><span class="line-number">228</span><br><span class="line-number">229</span><br><span class="line-number">230</span><br><span class="line-number">231</span><br><span class="line-number">232</span><br><span class="line-number">233</span><br><span class="line-number">234</span><br><span class="line-number">235</span><br><span class="line-number">236</span><br><span class="line-number">237</span><br><span class="line-number">238</span><br><span class="line-number">239</span><br><span class="line-number">240</span><br><span class="line-number">241</span><br><span class="line-number">242</span><br><span class="line-number">243</span><br><span class="line-number">244</span><br><span class="line-number">245</span><br><span class="line-number">246</span><br><span class="line-number">247</span><br><span class="line-number">248</span><br><span class="line-number">249</span><br><span class="line-number">250</span><br><span class="line-number">251</span><br><span class="line-number">252</span><br><span class="line-number">253</span><br><span class="line-number">254</span><br><span class="line-number">255</span><br><span class="line-number">256</span><br><span class="line-number">257</span><br><span class="line-number">258</span><br><span class="line-number">259</span><br><span class="line-number">260</span><br><span class="line-number">261</span><br><span class="line-number">262</span><br><span class="line-number">263</span><br><span class="line-number">264</span><br><span class="line-number">265</span><br><span class="line-number">266</span><br><span class="line-number">267</span><br><span class="line-number">268</span><br><span class="line-number">269</span><br><span class="line-number">270</span><br><span class="line-number">271</span><br><span class="line-number">272</span><br><span class="line-number">273</span><br><span class="line-number">274</span><br><span class="line-number">275</span><br><span class="line-number">276</span><br><span class="line-number">277</span><br><span class="line-number">278</span><br><span class="line-number">279</span><br><span class="line-number">280</span><br><span class="line-number">281</span><br><span class="line-number">282</span><br><span class="line-number">283</span><br><span class="line-number">284</span><br><span class="line-number">285</span><br><span class="line-number">286</span><br><span class="line-number">287</span><br><span class="line-number">288</span><br><span class="line-number">289</span><br><span class="line-number">290</span><br><span class="line-number">291</span><br><span class="line-number">292</span><br><span class="line-number">293</span><br><span class="line-number">294</span><br><span class="line-number">295</span><br><span class="line-number">296</span><br><span class="line-number">297</span><br><span class="line-number">298</span><br><span class="line-number">299</span><br><span class="line-number">300</span><br><span class="line-number">301</span><br><span class="line-number">302</span><br><span class="line-number">303</span><br><span class="line-number">304</span><br></div></div><h2 id="how-it-works" tabindex="-1">How It Works <a class="header-anchor" href="#how-it-works" aria-label="Permalink to &quot;How It Works&quot;">​</a></h2><h3 id="workflow" tabindex="-1">Workflow <a class="header-anchor" href="#workflow" aria-label="Permalink to &quot;Workflow&quot;">​</a></h3><ol><li><strong>Tenant Created</strong>: TenantRegistry creates Tenant CR from database</li><li><strong>Terraform Applied</strong>: Tenant controller creates Terraform CR</li><li><strong>tf-controller Processes</strong>: Runs terraform init/plan/apply</li><li><strong>Resources Provisioned</strong>: Cloud resources created (S3, RDS, etc.)</li><li><strong>Outputs Saved</strong>: Terraform outputs written to Kubernetes Secret</li><li><strong>App Deployed</strong>: Application uses infrastructure via Secret references</li><li><strong>Tenant Deleted</strong>: Terraform runs destroy (if deletionPolicy=Delete)</li></ol><h3 id="state-management" tabindex="-1">State Management <a class="header-anchor" href="#state-management" aria-label="Permalink to &quot;State Management&quot;">​</a></h3><p>Terraform state is stored in Kubernetes Secrets by default:</p><div class="language- line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang"></span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span>Secret: tfstate-default-{tenant-id}-{resource-name}</span></span>
<span class="line"><span>Namespace: default</span></span>
<span class="line"><span>Data: tfstate (gzipped)</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br></div></div><h2 id="best-practices" tabindex="-1">Best Practices <a class="header-anchor" href="#best-practices" aria-label="Permalink to &quot;Best Practices&quot;">​</a></h2><h3 id="_1-use-creationpolicy-once-for-immutable-infrastructure" tabindex="-1">1. Use CreationPolicy: Once for Immutable Infrastructure <a class="header-anchor" href="#_1-use-creationpolicy-once-for-immutable-infrastructure" aria-label="Permalink to &quot;1. Use CreationPolicy: Once for Immutable Infrastructure&quot;">​</a></h3><div class="language-yaml line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">yaml</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#85E89D;">manifests</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">- </span><span style="color:#85E89D;">id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">rds-database</span></span>
<span class="line"><span style="color:#85E89D;">  creationPolicy</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Once</span><span style="color:#6A737D;">  # Create once, never update</span></span>
<span class="line"><span style="color:#85E89D;">  deletionPolicy</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">Retain</span><span style="color:#6A737D;">  # Keep on tenant deletion</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br></div></div><h3 id="_2-set-appropriate-timeouts" tabindex="-1">2. Set Appropriate Timeouts <a class="header-anchor" href="#_2-set-appropriate-timeouts" aria-label="Permalink to &quot;2. Set Appropriate Timeouts&quot;">​</a></h3><p>Terraform provisioning can take 10-30 minutes:</p><div class="language-yaml line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">yaml</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#85E89D;">deployments</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">- </span><span style="color:#85E89D;">id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">app</span></span>
<span class="line"><span style="color:#85E89D;">  dependIds</span><span style="color:#E1E4E8;">: [</span><span style="color:#9ECBFF;">&quot;terraform-resources&quot;</span><span style="color:#E1E4E8;">]</span></span>
<span class="line"><span style="color:#85E89D;">  timeoutSeconds</span><span style="color:#E1E4E8;">: </span><span style="color:#79B8FF;">1800</span><span style="color:#6A737D;">  # 30 minutes</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br></div></div><h3 id="_3-use-remote-state-backend-production" tabindex="-1">3. Use Remote State Backend (Production) <a class="header-anchor" href="#_3-use-remote-state-backend-production" aria-label="Permalink to &quot;3. Use Remote State Backend (Production)&quot;">​</a></h3><p>For production, use S3 backend instead of Kubernetes:</p><div class="language-hcl line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">hcl</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#B392F0;">terraform</span><span style="color:#E1E4E8;"> {</span></span>
<span class="line"><span style="color:#B392F0;">  backend</span><span style="color:#79B8FF;"> &quot;s3&quot;</span><span style="color:#E1E4E8;"> {</span></span>
<span class="line"><span style="color:#E1E4E8;">    bucket</span><span style="color:#F97583;"> =</span><span style="color:#9ECBFF;"> &quot;my-terraform-state&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">    key</span><span style="color:#F97583;">    =</span><span style="color:#9ECBFF;"> &quot;tenants/</span><span style="color:#F97583;">\${</span><span style="color:#E1E4E8;">var</span><span style="color:#F97583;">.</span><span style="color:#E1E4E8;">tenant_id</span><span style="color:#F97583;">}</span><span style="color:#9ECBFF;">/terraform.tfstate&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">    region</span><span style="color:#F97583;"> =</span><span style="color:#9ECBFF;"> &quot;us-east-1&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">    encrypt</span><span style="color:#F97583;"> =</span><span style="color:#79B8FF;"> true</span></span>
<span class="line"><span style="color:#E1E4E8;">    dynamodb_table</span><span style="color:#F97583;"> =</span><span style="color:#9ECBFF;"> &quot;terraform-locks&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">  }</span></span>
<span class="line"><span style="color:#E1E4E8;">}</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br></div></div><h3 id="_4-secure-sensitive-outputs" tabindex="-1">4. Secure Sensitive Outputs <a class="header-anchor" href="#_4-secure-sensitive-outputs" aria-label="Permalink to &quot;4. Secure Sensitive Outputs&quot;">​</a></h3><p>Mark sensitive outputs:</p><div class="language-hcl line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">hcl</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#B392F0;">output</span><span style="color:#79B8FF;"> &quot;db_password&quot;</span><span style="color:#E1E4E8;"> {</span></span>
<span class="line"><span style="color:#E1E4E8;">  value</span><span style="color:#F97583;">     =</span><span style="color:#E1E4E8;"> random_password</span><span style="color:#F97583;">.</span><span style="color:#E1E4E8;">db_password</span><span style="color:#F97583;">.</span><span style="color:#E1E4E8;">result</span></span>
<span class="line"><span style="color:#E1E4E8;">  sensitive</span><span style="color:#F97583;"> =</span><span style="color:#79B8FF;"> true</span></span>
<span class="line"><span style="color:#E1E4E8;">}</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br></div></div><h3 id="_5-use-dependency-ordering" tabindex="-1">5. Use Dependency Ordering <a class="header-anchor" href="#_5-use-dependency-ordering" aria-label="Permalink to &quot;5. Use Dependency Ordering&quot;">​</a></h3><p>Ensure proper resource creation order:</p><div class="language-yaml line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">yaml</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#85E89D;">deployments</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#E1E4E8;">- </span><span style="color:#85E89D;">id</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">app</span></span>
<span class="line"><span style="color:#85E89D;">  dependIds</span><span style="color:#E1E4E8;">: [</span><span style="color:#9ECBFF;">&quot;tenant-infrastructure&quot;</span><span style="color:#E1E4E8;">]  </span><span style="color:#6A737D;"># Wait for Terraform</span></span>
<span class="line"><span style="color:#85E89D;">  waitForReady</span><span style="color:#E1E4E8;">: </span><span style="color:#79B8FF;">true</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br></div></div><h3 id="_6-monitor-terraform-resources" tabindex="-1">6. Monitor Terraform Resources <a class="header-anchor" href="#_6-monitor-terraform-resources" aria-label="Permalink to &quot;6. Monitor Terraform Resources&quot;">​</a></h3><div class="language-bash line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#6A737D;"># Check Terraform resources</span></span>
<span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> get</span><span style="color:#9ECBFF;"> terraform</span><span style="color:#79B8FF;"> -n</span><span style="color:#9ECBFF;"> default</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;"># Check specific tenant&#39;s Terraform</span></span>
<span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> get</span><span style="color:#9ECBFF;"> terraform</span><span style="color:#79B8FF;"> -n</span><span style="color:#9ECBFF;"> default</span><span style="color:#79B8FF;"> -l</span><span style="color:#9ECBFF;"> tenant-operator.kubernetes-tenants.org/tenant-id=tenant-alpha</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;"># View Terraform plan</span></span>
<span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> describe</span><span style="color:#9ECBFF;"> terraform</span><span style="color:#9ECBFF;"> tenant-alpha-infrastructure</span></span>
<span class="line"></span>
<span class="line"><span style="color:#6A737D;"># View Terraform outputs</span></span>
<span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> get</span><span style="color:#9ECBFF;"> secret</span><span style="color:#9ECBFF;"> tenant-alpha-infrastructure</span><span style="color:#79B8FF;"> -o</span><span style="color:#9ECBFF;"> yaml</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br></div></div><h2 id="troubleshooting" tabindex="-1">Troubleshooting <a class="header-anchor" href="#troubleshooting" aria-label="Permalink to &quot;Troubleshooting&quot;">​</a></h2><h3 id="terraform-apply-fails" tabindex="-1">Terraform Apply Fails <a class="header-anchor" href="#terraform-apply-fails" aria-label="Permalink to &quot;Terraform Apply Fails&quot;">​</a></h3><p><strong>Problem:</strong> Terraform fails to apply resources.</p><p><strong>Solution:</strong></p><ol><li><p><strong>Check Terraform logs:</strong></p><div class="language-bash line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> logs</span><span style="color:#79B8FF;"> -n</span><span style="color:#9ECBFF;"> flux-system</span><span style="color:#79B8FF;"> -l</span><span style="color:#9ECBFF;"> app=tf-controller</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br></div></div></li><li><p><strong>Check Terraform CR status:</strong></p><div class="language-bash line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> describe</span><span style="color:#9ECBFF;"> terraform</span><span style="color:#9ECBFF;"> tenant-alpha-infrastructure</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br></div></div></li><li><p><strong>View Terraform plan output:</strong></p><div class="language-bash line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> get</span><span style="color:#9ECBFF;"> terraform</span><span style="color:#9ECBFF;"> tenant-alpha-infrastructure</span><span style="color:#79B8FF;"> -o</span><span style="color:#9ECBFF;"> jsonpath=&#39;{.status.plan.pending}&#39;</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br></div></div></li><li><p><strong>Check credentials:</strong></p><div class="language-bash line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> get</span><span style="color:#9ECBFF;"> secret</span><span style="color:#9ECBFF;"> aws-credentials</span><span style="color:#79B8FF;"> -o</span><span style="color:#9ECBFF;"> yaml</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br></div></div></li></ol><h3 id="state-lock-issues" tabindex="-1">State Lock Issues <a class="header-anchor" href="#state-lock-issues" aria-label="Permalink to &quot;State Lock Issues&quot;">​</a></h3><p><strong>Problem:</strong> Terraform state locked.</p><p><strong>Solution:</strong></p><div class="language-bash line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#6A737D;"># Force unlock (use with caution!)</span></span>
<span class="line"><span style="color:#6A737D;"># This requires accessing the Terraform pod</span></span>
<span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> exec</span><span style="color:#79B8FF;"> -it</span><span style="color:#79B8FF;"> -n</span><span style="color:#9ECBFF;"> flux-system</span><span style="color:#9ECBFF;"> tf-controller-xxx</span><span style="color:#79B8FF;"> --</span><span style="color:#9ECBFF;"> sh</span></span>
<span class="line"><span style="color:#B392F0;">terraform</span><span style="color:#9ECBFF;"> force-unlock</span><span style="color:#F97583;"> &lt;</span><span style="color:#9ECBFF;">lock-i</span><span style="color:#E1E4E8;">d</span><span style="color:#F97583;">&gt;</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br></div></div><h3 id="outputs-not-available" tabindex="-1">Outputs Not Available <a class="header-anchor" href="#outputs-not-available" aria-label="Permalink to &quot;Outputs Not Available&quot;">​</a></h3><p><strong>Problem:</strong> Terraform outputs not written to secret.</p><p><strong>Solution:</strong></p><ol><li><p><strong>Verify writeOutputsToSecret is set:</strong></p><div class="language-yaml line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">yaml</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#85E89D;">writeOutputsToSecret</span><span style="color:#E1E4E8;">:</span></span>
<span class="line"><span style="color:#85E89D;">  name</span><span style="color:#E1E4E8;">: </span><span style="color:#9ECBFF;">&quot;{{ .uid }}-outputs&quot;</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br></div></div></li><li><p><strong>Check if Terraform apply completed:</strong></p><div class="language-bash line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> get</span><span style="color:#9ECBFF;"> terraform</span><span style="color:#9ECBFF;"> tenant-alpha-infra</span><span style="color:#79B8FF;"> -o</span><span style="color:#9ECBFF;"> jsonpath=&#39;{.status.conditions[?(@.type==&quot;Ready&quot;)].status}&#39;</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br></div></div></li><li><p><strong>Check secret exists:</strong></p><div class="language-bash line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#B392F0;">kubectl</span><span style="color:#9ECBFF;"> get</span><span style="color:#9ECBFF;"> secret</span><span style="color:#9ECBFF;"> tenant-alpha-outputs</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br></div></div></li></ol><h3 id="resource-already-exists" tabindex="-1">Resource Already Exists <a class="header-anchor" href="#resource-already-exists" aria-label="Permalink to &quot;Resource Already Exists&quot;">​</a></h3><p><strong>Problem:</strong> Terraform fails because resource already exists.</p><p><strong>Solution:</strong></p><p>Use <code>terraform import</code> or recreate with different name:</p><div class="language-hcl line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">hcl</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#B392F0;">resource</span><span style="color:#79B8FF;"> &quot;aws_s3_bucket&quot;</span><span style="color:#79B8FF;"> &quot;tenant_bucket&quot;</span><span style="color:#E1E4E8;"> {</span></span>
<span class="line"><span style="color:#E1E4E8;">  bucket</span><span style="color:#F97583;"> =</span><span style="color:#9ECBFF;"> &quot;tenant-</span><span style="color:#F97583;">\${</span><span style="color:#E1E4E8;">var</span><span style="color:#F97583;">.</span><span style="color:#E1E4E8;">tenant_id</span><span style="color:#F97583;">}</span><span style="color:#9ECBFF;">-bucket-v2&quot;</span><span style="color:#6A737D;">  # Add suffix</span></span>
<span class="line"><span style="color:#E1E4E8;">}</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br></div></div><h2 id="cost-optimization" tabindex="-1">Cost Optimization <a class="header-anchor" href="#cost-optimization" aria-label="Permalink to &quot;Cost Optimization&quot;">​</a></h2><h3 id="_1-use-appropriate-instance-sizes" tabindex="-1">1. Use Appropriate Instance Sizes <a class="header-anchor" href="#_1-use-appropriate-instance-sizes" aria-label="Permalink to &quot;1. Use Appropriate Instance Sizes&quot;">​</a></h3><div class="language-hcl line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">hcl</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#B392F0;">variable</span><span style="color:#79B8FF;"> &quot;db_instance_class&quot;</span><span style="color:#E1E4E8;"> {</span></span>
<span class="line"><span style="color:#E1E4E8;">  type</span><span style="color:#F97583;"> =</span><span style="color:#F97583;"> string</span></span>
<span class="line"><span style="color:#E1E4E8;">  default</span><span style="color:#F97583;"> =</span><span style="color:#9ECBFF;"> &quot;db.t3.micro&quot;</span><span style="color:#6A737D;">  # ~$15/month</span></span>
<span class="line"><span style="color:#E1E4E8;">}</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br></div></div><h3 id="_2-enable-auto-scaling" tabindex="-1">2. Enable Auto-Scaling <a class="header-anchor" href="#_2-enable-auto-scaling" aria-label="Permalink to &quot;2. Enable Auto-Scaling&quot;">​</a></h3><div class="language-hcl line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">hcl</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#B392F0;">resource</span><span style="color:#79B8FF;"> &quot;aws_appautoscaling_target&quot;</span><span style="color:#79B8FF;"> &quot;rds_target&quot;</span><span style="color:#E1E4E8;"> {</span></span>
<span class="line"><span style="color:#E1E4E8;">  max_capacity</span><span style="color:#F97583;">       =</span><span style="color:#79B8FF;"> 10</span></span>
<span class="line"><span style="color:#E1E4E8;">  min_capacity</span><span style="color:#F97583;">       =</span><span style="color:#79B8FF;"> 1</span></span>
<span class="line"><span style="color:#E1E4E8;">  resource_id</span><span style="color:#F97583;">        =</span><span style="color:#9ECBFF;"> &quot;cluster:</span><span style="color:#F97583;">\${</span><span style="color:#E1E4E8;">aws_rds_cluster</span><span style="color:#F97583;">.</span><span style="color:#E1E4E8;">tenant_db</span><span style="color:#F97583;">.</span><span style="color:#E1E4E8;">cluster_identifier</span><span style="color:#F97583;">}</span><span style="color:#9ECBFF;">&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">  scalable_dimension</span><span style="color:#F97583;"> =</span><span style="color:#9ECBFF;"> &quot;rds:cluster:ReadReplicaCount&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">  service_namespace</span><span style="color:#F97583;">  =</span><span style="color:#9ECBFF;"> &quot;rds&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">}</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br></div></div><h3 id="_3-use-lifecycle-policies" tabindex="-1">3. Use Lifecycle Policies <a class="header-anchor" href="#_3-use-lifecycle-policies" aria-label="Permalink to &quot;3. Use Lifecycle Policies&quot;">​</a></h3><div class="language-hcl line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">hcl</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span style="color:#B392F0;">resource</span><span style="color:#79B8FF;"> &quot;aws_s3_bucket_lifecycle_configuration&quot;</span><span style="color:#79B8FF;"> &quot;tenant_bucket_lifecycle&quot;</span><span style="color:#E1E4E8;"> {</span></span>
<span class="line"><span style="color:#E1E4E8;">  bucket</span><span style="color:#F97583;"> =</span><span style="color:#E1E4E8;"> aws_s3_bucket</span><span style="color:#F97583;">.</span><span style="color:#E1E4E8;">tenant_bucket</span><span style="color:#F97583;">.</span><span style="color:#E1E4E8;">id</span></span>
<span class="line"></span>
<span class="line"><span style="color:#B392F0;">  rule</span><span style="color:#E1E4E8;"> {</span></span>
<span class="line"><span style="color:#E1E4E8;">    id</span><span style="color:#F97583;">     =</span><span style="color:#9ECBFF;"> &quot;archive-old-data&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">    status</span><span style="color:#F97583;"> =</span><span style="color:#9ECBFF;"> &quot;Enabled&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="color:#B392F0;">    transition</span><span style="color:#E1E4E8;"> {</span></span>
<span class="line"><span style="color:#E1E4E8;">      days</span><span style="color:#F97583;">          =</span><span style="color:#79B8FF;"> 90</span></span>
<span class="line"><span style="color:#E1E4E8;">      storage_class</span><span style="color:#F97583;"> =</span><span style="color:#9ECBFF;"> &quot;GLACIER&quot;</span></span>
<span class="line"><span style="color:#E1E4E8;">    }</span></span>
<span class="line"></span>
<span class="line"><span style="color:#B392F0;">    expiration</span><span style="color:#E1E4E8;"> {</span></span>
<span class="line"><span style="color:#E1E4E8;">      days</span><span style="color:#F97583;"> =</span><span style="color:#79B8FF;"> 365</span></span>
<span class="line"><span style="color:#E1E4E8;">    }</span></span>
<span class="line"><span style="color:#E1E4E8;">  }</span></span>
<span class="line"><span style="color:#E1E4E8;">}</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br><span class="line-number">12</span><br><span class="line-number">13</span><br><span class="line-number">14</span><br><span class="line-number">15</span><br><span class="line-number">16</span><br><span class="line-number">17</span><br></div></div><h2 id="see-also" tabindex="-1">See Also <a class="header-anchor" href="#see-also" aria-label="Permalink to &quot;See Also&quot;">​</a></h2><ul><li><a href="https://github.com/flux-iac/tofu-controller" target="_blank" rel="noreferrer">Tofu Controller (OpenTofu/Terraform)</a></li><li><a href="https://fluxcd.io/docs/" target="_blank" rel="noreferrer">Flux Documentation</a></li><li><a href="https://registry.terraform.io/browse/providers" target="_blank" rel="noreferrer">Terraform Registry - All Providers</a></li><li><a href="./integration-external-dns.html">ExternalDNS Integration</a></li><li><a href="./templates.html">Tenant Operator Templates Guide</a></li><li><a href="https://registry.terraform.io/providers/hashicorp/aws/latest/docs" target="_blank" rel="noreferrer">AWS Terraform Provider</a></li><li><a href="https://registry.terraform.io/providers/Mongey/kafka/latest/docs" target="_blank" rel="noreferrer">Kafka Terraform Provider</a></li><li><a href="https://registry.terraform.io/providers/cyrilgdn/rabbitmq/latest/docs" target="_blank" rel="noreferrer">RabbitMQ Terraform Provider</a></li><li><a href="https://registry.terraform.io/providers/cyrilgdn/postgresql/latest/docs" target="_blank" rel="noreferrer">PostgreSQL Terraform Provider</a></li><li><a href="https://registry.terraform.io/providers/phillbaker/elasticsearch/latest/docs" target="_blank" rel="noreferrer">Elasticsearch Terraform Provider</a></li></ul>`,121)])])}const y=n(e,[["render",r]]);export{E as __pageData,y as default};
