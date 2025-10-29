# Contributing to Tenant Operator

Thank you for your interest in contributing to Tenant Operator! We welcome contributions from the community.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
- [Development Workflow](#development-workflow)
- [Coding Guidelines](#coding-guidelines)
- [Commit Message Guidelines](#commit-message-guidelines)
- [Pull Request Process](#pull-request-process)
- [Review Process](#review-process)

---

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please be respectful and constructive in all interactions.

**Expected Behavior:**
- Be respectful and inclusive
- Welcome newcomers
- Focus on what is best for the community
- Show empathy towards other community members

**Unacceptable Behavior:**
- Harassment, trolling, or derogatory comments
- Personal or political attacks
- Publishing others' private information
- Other conduct which could reasonably be considered inappropriate

---

## Getting Started

### Prerequisites

- Go 1.24 or later
- Docker 17.03+
- kubectl v1.11.3+
- Access to a Kubernetes cluster (kind, minikube, or remote)
- Git

### Setting Up Your Development Environment

1. **Fork the repository** on GitHub

2. **Clone your fork:**
   ```bash
   git clone https://github.com/<your-username>/tenant-operator.git
   cd tenant-operator
   ```

3. **Add upstream remote:**
   ```bash
   git remote add upstream https://github.com/kubernetes-tenants/tenant-operator.git
   ```

4. **Install dependencies:**
   ```bash
   go mod download
   ```

5. **Install CRDs:**
   ```bash
   make install
   ```

6. **Run tests:**
   ```bash
   make test
   ```

---

## How to Contribute

### Reporting Bugs

Before creating a bug report:
- Check the [existing issues](https://github.com/kubernetes-tenants/tenant-operator/issues)
- Ensure you're using the latest version

**Bug Report Template:**
```markdown
**Description:**
A clear description of what the bug is.

**Steps to Reproduce:**
1. Deploy with configuration X
2. Apply resource Y
3. Observe error Z

**Expected Behavior:**
What you expected to happen.

**Actual Behavior:**
What actually happened.

**Environment:**
- Tenant Operator version:
- Kubernetes version:
- MySQL version (if applicable):
- OS:

**Logs:**
```
paste relevant logs here
```

**Additional Context:**
Any other information that might be helpful.
```

### Suggesting Features

We welcome feature suggestions! Please:
- Search existing issues for similar requests
- Create a detailed issue describing:
  - The problem you're trying to solve
  - Your proposed solution
  - Alternative solutions you've considered
  - How this benefits the community

### Documentation Improvements

Documentation improvements are always welcome! This includes:
- README improvements
- Code comments
- API documentation
- Tutorials and guides
- Fixing typos

---

## Development Workflow

### Creating a Feature Branch

```bash
# Sync with upstream
git fetch upstream
git checkout main
git merge upstream/main

# Create feature branch
git checkout -b feature/your-feature-name
```

### Making Changes

1. Make your changes
2. Add tests for new features
3. Update documentation
4. Run tests locally:
   ```bash
   make test
   make test-integration
   ```

5. Lint your code:
   ```bash
   make lint
   ```

### Running the Operator Locally

```bash
# Run against your Kubernetes cluster
make run

# Or with debug logging
LOG_LEVEL=debug make run
```

### Testing Your Changes

```bash
# Unit tests
make test

# Integration tests (requires cluster)
make test-integration

# E2E tests (requires kind)
make test-e2e

# Check coverage
make test-coverage
```

---

## Coding Guidelines

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting (run `make fmt`)
- Use `golangci-lint` for linting (run `make lint`)
- Write tests for new functionality
- Keep functions small and focused

### Code Organization

```
├── api/v1/              # CRD definitions
├── cmd/                 # Main application entry point
├── config/              # Kubernetes manifests
├── internal/
│   ├── controller/      # Controller implementations
│   ├── database/        # Database clients
│   ├── template/        # Template engine
│   ├── apply/           # SSA apply logic
│   ├── graph/           # Dependency graph
│   └── metrics/         # Prometheus metrics
└── test/                # E2E and integration tests
```

### Testing Guidelines

- **Unit Tests**: Test individual functions in isolation
  ```go
  func TestMyFunction(t *testing.T) {
      result := MyFunction("input")
      if result != "expected" {
          t.Errorf("got %s, want %s", result, "expected")
      }
  }
  ```

- **Integration Tests**: Test controller behavior with envtest
  ```go
  func TestControllerReconciliation(t *testing.T) {
      // Use testenv from controller-runtime
  }
  ```

- **Table-Driven Tests**: Use for multiple test cases
  ```go
  tests := []struct {
      name string
      input string
      want string
  }{
      {"case1", "input1", "output1"},
      {"case2", "input2", "output2"},
  }
  ```

### Error Handling

- Always return errors, don't panic
- Wrap errors with context:
  ```go
  if err != nil {
      return fmt.Errorf("failed to apply resource %s: %w", name, err)
  }
  ```
- Log errors appropriately:
  ```go
  logger.Error(err, "Failed to reconcile", "tenant", tenant.Name)
  ```

### Logging

Use structured logging:
```go
logger.Info("Reconciling tenant",
    "tenant", tenant.Name,
    "namespace", tenant.Namespace,
    "resources", resourceCount)
```

**Log Levels:**
- `Error`: Errors that need attention
- `Info`: Important informational messages
- `V(1).Info`: Debug information

---

## Commit Message Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, no logic change)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### Examples

```
feat(controller): add drift detection with periodic reconciliation

Implements automatic drift detection by watching owned resources
and requeuing reconciliation every 5 minutes.

Closes #123
```

```
fix(template): handle nil pointer in fromJson function

The fromJson function crashed when receiving empty strings.
Now returns empty map gracefully.

Fixes #456
```

```
docs(readme): add FAQ section with common questions

Added 5 frequently asked questions covering:
- Comparison with other solutions
- Database integration
- Scaling capabilities
```

---

## Pull Request Process

### Before Submitting

1. ✅ Tests pass: `make test`
2. ✅ Linting passes: `make lint`
3. ✅ Code is formatted: `make fmt`
4. ✅ Documentation is updated
5. ✅ Commit messages follow conventions
6. ✅ Branch is up to date with main

### Creating a Pull Request

1. **Push your branch:**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create PR on GitHub** with:
   - Clear title following commit conventions
   - Detailed description of changes
   - Reference to related issues
   - Screenshots (if UI changes)

### PR Template

```markdown
## Description
Brief description of changes

## Related Issues
Closes #123
Fixes #456

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests pass
- [ ] Manually tested in cluster

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] Tests added for new features
- [ ] All tests pass
```

### PR Size Guidelines

- **Small PRs** (< 200 lines): Preferred, reviewed quickly
- **Medium PRs** (200-500 lines): Acceptable with good description
- **Large PRs** (> 500 lines): Split if possible, requires detailed explanation

---

## Review Process

### What Reviewers Look For

1. **Correctness**: Does it solve the problem?
2. **Tests**: Are there adequate tests?
3. **Code Quality**: Is it readable and maintainable?
4. **Documentation**: Is it documented?
5. **Backwards Compatibility**: Does it break existing functionality?

### Responding to Reviews

- Address all comments
- Ask questions if feedback is unclear
- Update your PR and request re-review
- Be patient and respectful

### Approval and Merge

- PRs require **at least 1 approval** from maintainers
- All CI checks must pass
- Maintainers will merge when ready
- Squash merging is preferred for cleaner history

---

## Development Tips

### Running Specific Tests

```bash
# Run single test
go test ./internal/controller -run TestTenantController

# Run with verbose output
go test -v ./internal/template

# Run with coverage
go test -coverprofile=coverage.out ./internal/controller
go tool cover -html=coverage.out
```

### Debugging

```bash
# Run with delve debugger
dlv debug cmd/main.go

# Or use your IDE's debugger (VSCode, GoLand)
```

### Useful Make Targets

```bash
make help              # Show all available targets
make generate          # Generate code (deepcopy, CRDs, etc.)
make manifests         # Generate Kubernetes manifests
make docker-build      # Build container image
make deploy            # Deploy to cluster
make undeploy          # Remove from cluster
```

---

## Getting Help

- 💬 [GitHub Discussions](https://github.com/kubernetes-tenants/tenant-operator/discussions) - Ask questions
- 🐛 [GitHub Issues](https://github.com/kubernetes-tenants/tenant-operator/issues) - Report bugs
- 📧 maintainers@kubernetes-tenants.org - Direct contact

---

## Recognition

Contributors will be:
- Listed in release notes
- Acknowledged in the README
- Invited to join the maintainers (for sustained contributions)

Thank you for contributing to Tenant Operator! 🎉
