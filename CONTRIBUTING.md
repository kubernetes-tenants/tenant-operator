# Contributing to Tenant Operator

Thank you for your interest in contributing to Tenant Operator! This guide will help you get started.

## Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please be respectful and constructive in all interactions.

## How to Contribute

### Reporting Bugs

Before creating a bug report:
1. Check [existing issues](https://github.com/kubernetes-tenants/tenant-operator/issues)
2. Ensure you're using the latest version
3. Collect relevant information (logs, YAML files, etc.)

**Bug Report Template:**
```markdown
**Description:**
Brief description of the bug

**Steps to Reproduce:**
1. Step 1
2. Step 2
3. ...

**Expected Behavior:**
What should happen

**Actual Behavior:**
What actually happens

**Environment:**
- Operator version: v1.0.0
- Kubernetes version: v1.28.0
- Platform: GKE / EKS / kind
- Database: MySQL 8.0

**Logs:**
\```
Paste relevant logs here
\```

**YAML:**
\```yaml
# Paste relevant CRD YAML
\```
```

### Requesting Features

Feature requests are welcome! Please:
1. Check [existing requests](https://github.com/kubernetes-tenants/tenant-operator/discussions/categories/feature-requests)
2. Describe your use case
3. Explain why existing features don't work
4. Propose a solution if possible

**Feature Request Template:**
```markdown
**Feature Description:**
Brief description of the feature

**Use Case:**
Why you need this feature

**Proposed Solution:**
How you envision it working

**Alternatives Considered:**
Other approaches you've considered
```

### Contributing Code

#### Prerequisites

- Go 1.22+
- kubectl
- kind or minikube
- Docker
- make
- golangci-lint

#### Workflow

1. **Fork the Repository**
   ```bash
   # Click "Fork" on GitHub
   git clone https://github.com/YOUR_USERNAME/tenant-operator.git
   cd tenant-operator
   ```

2. **Create a Branch**
   ```bash
   git checkout -b feature/my-feature
   # or
   git checkout -b fix/my-bugfix
   ```

3. **Make Changes**
   - Write code
   - Add tests
   - Update documentation
   - Follow code style guidelines

4. **Test Your Changes**
   ```bash
   # Run tests
   make test

   # Run linter
   make lint

   # Test locally
   make install
   make run
   ```

5. **Commit Your Changes**
   ```bash
   # Use conventional commits
   git commit -m "feat: add new feature"
   git commit -m "fix: resolve bug"
   git commit -m "docs: update documentation"
   ```

6. **Push to Your Fork**
   ```bash
   git push origin feature/my-feature
   ```

7. **Open a Pull Request**
   - Go to https://github.com/kubernetes-tenants/tenant-operator
   - Click "Compare & pull request"
   - Fill out the PR template
   - Link related issues

#### Conventional Commits

We use [Conventional Commits](https://www.conventionalcommits.org/) for clear commit history:

```
feat: add new feature
fix: fix bug
docs: update documentation
test: add or update tests
refactor: refactor code without changing behavior
perf: improve performance
chore: maintenance tasks
ci: CI/CD changes
style: code style changes (formatting, etc.)
```

**Examples:**
```bash
git commit -m "feat: add PostgreSQL datasource support"
git commit -m "fix: resolve template rendering error for missing variables"
git commit -m "docs: add examples for multi-template setup"
git commit -m "test: add unit tests for dependency graph"
```

### Code Style Guidelines

#### Go Code

Follow standard Go conventions:

```go
// Good
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // Get Tenant
    tenant := &tenantsv1.Tenant{}
    if err := r.Get(ctx, req.NamespacedName, tenant); err != nil {
        if errors.IsNotFound(err) {
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }

    // Business logic...

    return ctrl.Result{}, nil
}

// Bad - no error handling
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    tenant := &tenantsv1.Tenant{}
    r.Get(ctx, req.NamespacedName, tenant)
    return ctrl.Result{}, nil
}
```

**Guidelines:**
- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Add comments for exported functions
- Keep functions small and focused
- Handle all errors explicitly

#### Testing

Write tests for all new code:

```go
func TestTenantController_Reconcile(t *testing.T) {
    tests := []struct {
        name    string
        tenant  *tenantsv1.Tenant
        want    ctrl.Result
        wantErr bool
    }{
        {
            name: "successful reconciliation",
            tenant: &tenantsv1.Tenant{
                // Test setup
            },
            want:    ctrl.Result{RequeueAfter: 30 * time.Second},
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

**Guidelines:**
- Write table-driven tests
- Test happy path and error cases
- Use meaningful test names
- Mock external dependencies
- Aim for > 80% coverage

### Documentation

Update documentation for:
- New features
- API changes
- Configuration options
- Examples

**Documentation files:**
- `README.md` - Overview and quick start
- `docs/*.md` - Detailed guides
- `CLAUDE.md` - Development guidelines
- Code comments - Function documentation

### Pull Request Guidelines

#### PR Title

Use conventional commit format:
```
feat: add PostgreSQL datasource support
fix: resolve template rendering error
docs: add multi-template examples
```

#### PR Description

Use this template:

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change fixing an issue)
- [ ] New feature (non-breaking change adding functionality)
- [ ] Breaking change (fix or feature causing existing functionality to not work)
- [ ] Documentation update

## Related Issues
Fixes #123
Related to #456

## How Has This Been Tested?
- [ ] Unit tests
- [ ] Integration tests
- [ ] Manual testing on kind cluster
- [ ] Tested with MySQL 8.0

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] All tests passing
- [ ] Linter passing
- [ ] No breaking changes (or documented if necessary)

## Screenshots (if applicable)
```

#### Code Review Process

1. Automated checks run (tests, linter)
2. Maintainers review code
3. Address feedback
4. Approval required from 1+ maintainers
5. Merge to main

**Review Timeline:**
- Initial response: 1-3 days
- Full review: 3-7 days
- Large PRs may take longer

## Development Setup

See [Development Guide](docs/development.md) for detailed setup instructions.

**Quick Start:**
```bash
# Clone and setup
git clone https://github.com/YOUR_USERNAME/tenant-operator.git
cd tenant-operator
go mod download

# Create test cluster
kind create cluster --name tenant-dev

# Install and run
make install
make run
```

## Community

### Communication Channels

- 💬 **Discussions**: https://github.com/kubernetes-tenants/tenant-operator/discussions
- 🐛 **Issues**: https://github.com/kubernetes-tenants/tenant-operator/issues
- 📧 **Email**: rationlunas@gmail.com

### Getting Help

- Check [documentation](docs/)
- Search [existing issues](https://github.com/kubernetes-tenants/tenant-operator/issues)
- Ask in [discussions](https://github.com/kubernetes-tenants/tenant-operator/discussions)

### Regular Meetings

- **Community Call**: First Tuesday of each month, 3PM UTC
- **Agenda**: https://github.com/kubernetes-tenants/tenant-operator/discussions

## Recognition

Contributors are recognized in:
- Release notes
- CONTRIBUTORS file
- Project README

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.

## Questions?

Feel free to reach out:
- Open a [discussion](https://github.com/kubernetes-tenants/tenant-operator/discussions)

Thank you for contributing to Tenant Operator! 🎉
