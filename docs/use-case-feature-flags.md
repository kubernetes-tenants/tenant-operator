# Dynamic Feature Flags

## Overview

Enable/disable features per tenant using environment variables and separate templates for optional components.

This pattern enables:
- **A/B Testing**: Test new features with subset of users
- **Gradual Rollout**: Enable features progressively
- **Plan-Based Features**: Different features for different subscription tiers
- **Cost Control**: Expensive features (GPU, etc.) only for paying customers
- **Rapid Iteration**: Enable/disable features without redeploying

## Patterns

There are two main patterns for implementing feature flags:

### Pattern 1: Application-Level Feature Flags
Pass flags as environment variables, application handles the logic. Best for lightweight features.

### Pattern 2: Infrastructure-Level Feature Flags
Use database views and separate templates for expensive features (GPU, workers, etc.).

## Database Schema

```sql
CREATE TABLE tenants (
  tenant_id VARCHAR(63) PRIMARY KEY,
  domain VARCHAR(255) NOT NULL,
  is_active BOOLEAN DEFAULT TRUE,

  -- Feature flags (boolean features)
  feature_analytics BOOLEAN DEFAULT FALSE,
  feature_ai_assistant BOOLEAN DEFAULT FALSE,
  feature_advanced_reports BOOLEAN DEFAULT FALSE,
  feature_sso BOOLEAN DEFAULT FALSE,
  feature_audit_logs BOOLEAN DEFAULT TRUE,
  feature_webhooks BOOLEAN DEFAULT FALSE,

  -- Feature configuration (JSON for complex settings)
  feature_config JSON,

  plan_type VARCHAR(20) DEFAULT 'basic'
);

-- Example feature_config JSON:
-- {
--   "rate_limits": {"requests_per_minute": 100},
--   "storage_quota_gb": 50,
--   "max_users": 25,
--   "custom_branding": {"logo_url": "https://...", "primary_color": "#ff6600"}
-- }
```

## Pattern 1: Application-Level Feature Flags

Pass feature flags as environment variables and let the application enable/disable features:

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: base-app
  namespace: tenant-operator-system
spec:
  registryId: production-tenants

  deployments:
    - id: main-app
      nameTemplate: "{{ .uid }}-app"
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
                image: registry.example.com/tenant-app:v1.5.0
                env:
                  - name: TENANT_ID
                    value: "{{ .uid }}"

                  # Feature flags as environment variables
                  - name: FEATURE_ANALYTICS
                    value: "{{ .featureAnalytics }}"
                  - name: FEATURE_SSO
                    value: "{{ .featureSso }}"
                  - name: FEATURE_AUDIT_LOGS
                    value: "{{ .featureAuditLogs }}"
                  - name: FEATURE_ADVANCED_REPORTS
                    value: "{{ .featureAdvancedReports }}"

                  # Complex feature config as JSON
                  - name: FEATURE_CONFIG
                    value: "{{ .featureConfig | toJson }}"
                ports:
                  - containerPort: 8080
                resources:
                  requests:
                    cpu: 500m
                    memory: 1Gi

  services:
    - id: app-svc
      nameTemplate: "{{ .uid }}-app"
      dependIds: ["main-app"]
      spec:
        selector:
          app: "{{ .uid }}"
        ports:
          - port: 80
            targetPort: 8080
```

**Benefits:**
- Simple and efficient
- All tenants get same deployment
- Features enabled/disabled at application level
- No infrastructure changes needed

**Use for:** Lightweight features that don't require additional infrastructure

## Pattern 2: Separate Templates for Optional Features

For expensive features (like AI assistants with GPU), use database views and separate templates:

### Database Views

```sql
-- Base view for all active tenants
CREATE OR REPLACE VIEW tenants_active AS
SELECT * FROM tenants WHERE is_active = TRUE;

-- View for tenants with AI assistant enabled
CREATE OR REPLACE VIEW tenants_with_ai AS
SELECT * FROM tenants WHERE is_active = TRUE AND feature_ai_assistant = TRUE;

-- View for tenants with webhook workers
CREATE OR REPLACE VIEW tenants_with_webhooks AS
SELECT * FROM tenants WHERE is_active = TRUE AND feature_webhooks = TRUE;
```

### TenantRegistry for AI Assistant

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantRegistry
metadata:
  name: tenants-with-ai
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
      table: tenants_with_ai  # View that filters AI-enabled tenants

  valueMappings:
    uid: tenant_id
    hostOrUrl: domain
    activate: is_active

  extraValueMappings:
    featureConfig: feature_config
```

### TenantTemplate for AI Assistant

```yaml
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: ai-assistant
  namespace: tenant-operator-system
spec:
  registryId: tenants-with-ai  # References filtered registry

  deployments:
    - id: ai-assistant
      nameTemplate: "{{ .uid }}-ai"
      waitForReady: true
      spec:
        replicas: 1
        selector:
          matchLabels:
            app: "{{ .uid }}-ai"
            component: ai-assistant
        template:
          metadata:
            labels:
              app: "{{ .uid }}-ai"
              component: ai-assistant
          spec:
            containers:
              - name: ai-assistant
                image: registry.example.com/ai-assistant:v2.0.0
                env:
                  - name: TENANT_ID
                    value: "{{ .uid }}"
                  - name: OPENAI_API_KEY
                    valueFrom:
                      secretKeyRef:
                        name: openai-credentials
                        key: api-key
                ports:
                  - containerPort: 8080
                    name: http
                resources:
                  requests:
                    cpu: 1000m
                    memory: 2Gi
                    nvidia.com/gpu: "1"
                  limits:
                    cpu: 2000m
                    memory: 4Gi
                    nvidia.com/gpu: "1"

  services:
    - id: ai-service
      nameTemplate: "{{ .uid }}-ai"
      dependIds: ["ai-assistant"]
      spec:
        selector:
          app: "{{ .uid }}-ai"
        ports:
          - port: 80
            targetPort: http
```

**Benefits:**
- Resource efficiency: Only tenants with enabled features consume resources
- Cost optimization: GPU/expensive resources only allocated when needed
- Automatic cleanup: Disabling feature in DB automatically removes infrastructure
- Independent scaling: Feature-specific resources scaled separately

**Use for:** Expensive features requiring dedicated infrastructure (GPU, workers, etc.)

## Feature Rollout Workflow

### Enable Application-Level Feature (Pattern 1)

```sql
-- Enable SSO for premium customer
UPDATE tenants
SET feature_sso = TRUE,
    feature_config = JSON_SET(feature_config, '$.sso_provider', 'okta')
WHERE tenant_id = 'acme-corp';
```

Tenant Operator:
1. Updates main-app Deployment with new environment variables
2. Kubernetes rolls out the updated pods automatically
3. Application detects `FEATURE_SSO=true` and enables SSO

### Enable Infrastructure-Level Feature (Pattern 2)

```sql
-- Enable AI assistant for premium customer
UPDATE tenants
SET feature_ai_assistant = TRUE,
    feature_config = JSON_SET(feature_config, '$.ai_model', 'gpt-4')
WHERE tenant_id = 'acme-corp';
```

Since the database view `tenants_with_ai` filters on `feature_ai_assistant = TRUE`:
1. Registry `tenants-with-ai` syncs and detects new tenant
2. Creates Tenant CR `acme-corp-ai-assistant`
3. Deploys AI assistant Deployment + Service with GPU
4. Marks Tenant as Ready once all resources are up

### Gradual Rollout by Plan

```sql
-- Enable advanced reports for all pro/enterprise customers
UPDATE tenants
SET feature_advanced_reports = TRUE
WHERE plan_type IN ('pro', 'enterprise');
```

### Feature Flag A/B Testing

```sql
-- Enable webhooks for 10% of users (random sampling)
UPDATE tenants
SET feature_webhooks = TRUE
WHERE RAND() < 0.1 AND plan_type = 'pro';
```

This triggers creation of webhook worker deployments only for those 10% of tenants.

## Automatic Cleanup on Feature Disable

### Pattern 1 (Application-Level)

```sql
UPDATE tenants SET feature_sso = FALSE WHERE tenant_id = 'acme-corp';
```

- Environment variable updated in Deployment
- Kubernetes rolls out updated pods
- No resource deletion needed

### Pattern 2 (Infrastructure-Level)

```sql
UPDATE tenants SET feature_ai_assistant = FALSE WHERE tenant_id = 'acme-corp';
```

- Tenant `acme-corp` no longer appears in `tenants_with_ai` view
- Registry `tenants-with-ai` syncs and detects tenant removal
- Tenant Operator deletes Tenant CR `acme-corp-ai-assistant`
- AI assistant Deployment + Service automatically garbage collected
- GPU resources freed

## Benefits

1. **Flexibility**: Enable/disable features without code deployment
2. **A/B Testing**: Easy to test features with subset of users
3. **Cost Control**: Expensive features only for paying customers
4. **Gradual Rollout**: Roll out features progressively
5. **Plan-Based Features**: Different features for different tiers

## Best Practices

1. **Default Values**: Set sensible defaults for feature flags
2. **Documentation**: Document what each feature flag controls
3. **Monitoring**: Track feature usage and performance
4. **Testing**: Test feature combinations thoroughly
5. **Cleanup**: Implement proper cleanup when features are disabled

## Related Documentation

- [Templates Guide](/templates) - Template functions and conditionals
- [Policies](/policies) - Resource lifecycle management
- [Monitoring](/monitoring) - Feature usage tracking
- [Advanced Use Cases](/advanced-use-cases) - Other patterns

## Next Steps

- Design feature flag schema
- Implement application-level feature detection
- Set up monitoring for feature usage
- Create feature rollout playbook
