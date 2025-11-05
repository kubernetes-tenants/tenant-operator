# Roadmap

Future plans and feature roadmap for Tenant Operator.

[[toc]]

## v1.0 ✅

::: info Status
Released
:::

### Features
- ✅ MySQL datasource support
- ✅ Template-based resource generation
- ✅ Server-Side Apply (SSA)
- ✅ Dependency management with DAG
- ✅ Policy-based lifecycle (Creation/Deletion/Conflict)
- ✅ Patch strategies (apply/merge/replace)
- ✅ Fast reconciliation (30s requeue)
- ✅ Smart watch predicates
- ✅ Multi-template support
- ✅ Webhook validation
- ✅ Prometheus metrics
- ✅ Comprehensive documentation

### Performance
- ✅ Event-driven architecture
- ✅ Optimized reconciliation
- ✅ Label-based namespace tracking
- ✅ Efficient database querying

## v1.1 (Current) ✅

::: info Focus
Cross-namespace support and operational improvements
:::

### New Features

- ✅ **Helm Chart Distribution**
  - Helm chart published via GitHub Releases
  - Public repo: https://kubernetes-tenants.github.io/tenant-operator
  - Customizable values and upgrade path with `helm upgrade`

- ✅ **Cross-Namespace Resource Provisioning**
  - Support creating tenant resources in different namespaces using `targetNamespace` field
  - Uses label-based tracking (`kubernetes-tenants.org/tenant`, `kubernetes-tenants.org/tenant-namespace`) for cross-namespace resources
  - Automatic detection: same-namespace uses ownerReferences, cross-namespace uses labels
  - Dual watch system: `Owns()` for same-namespace + `Watches()` with label selectors for cross-namespace
  - Enables multi-namespace tenant isolation and organizational boundaries

- ✅ **Orphan Resource Cleanup**
  - Automatic detection and cleanup of resources removed from templates
  - Status-based tracking with `appliedResources` field
  - Respects DeletionPolicy (Delete/Retain)
  - Orphan labels for retained resources for easy identification

### Improvements
- ✅ Fast reconciliation (30s requeue)
- ✅ Smart watch predicates
- ✅ Event-driven architecture optimizations

## v1.2

::: info Focus
Additional datasources and enhanced observability
:::

### New Features

- [ ] **PostgreSQL Datasource**
  - Full PostgreSQL support
  - Connection pooling
  - SSL/TLS support
  - Query optimization

- [ ] **Enhanced Metrics Dashboard**
  - Pre-built Grafana dashboards
  - Comprehensive AlertManager rules
  - Multi-tenant metrics visualization
  - Performance analytics

### Improvements
- [ ] Improved error messages
- [ ] Performance optimizations
- [ ] Extended template functions
- [ ] Better documentation examples

## Contributing to Roadmap

Want to influence the roadmap?

1. **Open a Discussion**: Share your use case
2. **Vote on Features**: Upvote existing requests
3. **Submit PRs**: Implement features yourself
4. **Join Community**: Participate in discussions

## Stability Commitments

### API Stability
- v1 API: Stable, no breaking changes
- Future versions: Migration guides provided
- Deprecation policy: 6 months notice

### Backwards Compatibility
- Database schema changes: Automatic migration
- Template syntax: Backwards compatible
- Metrics: No breaking changes without notice

## Getting Involved

- 💬 Discussions: https://github.com/kubernetes-tenants/tenant-operator/discussions
- 🐛 Issues: https://github.com/kubernetes-tenants/tenant-operator/issues
- 📧 Email: rationlunas@gmail.com
- 🔔 Release notifications: Watch repository

## See Also

- [Contributing Guide](https://github.com/kubernetes-tenants/tenant-operator/blob/main/CONTRIBUTING.md)
- [Development Guide](development.md)
- [GitHub Discussions](https://github.com/kubernetes-tenants/tenant-operator/discussions)
