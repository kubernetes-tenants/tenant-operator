# Roadmap

Future plans and feature roadmap for Tenant Operator.

## v1.0 (Current) ✅

**Status:** Released

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

## v1.1

**Focus:** Additional datasources and operational improvements

### New Features
- [ ] **PostgreSQL Datasource**
  - Full PostgreSQL support
  - Connection pooling
  - SSL/TLS support

- [ ] **Helm Chart Distribution**
  - Official Helm chart
  - Artifact Hub listing
  - Customizable values
  - Easy upgrades

- [ ] **Enhanced Metrics Dashboard**
  - Pre-built Grafana dashboards
  - AlertManager rules

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

- [Contributing Guide](../CONTRIBUTING.md)
- [Development Guide](development.md)
- [GitHub Discussions](https://github.com/kubernetes-tenants/tenant-operator/discussions)
