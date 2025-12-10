# Contributing to Monify Agent

Thank you for your interest in contributing to Monify Agent! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)

## Code of Conduct

This project adheres to a Code of Conduct. By participating, you are expected to uphold this code. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/agent.git
   cd agent
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/monify-labs/agent.git
   ```

## Development Setup

### Prerequisites

- **Go 1.21+**: [Install Go](https://golang.org/doc/install)
- **Linux environment**: Required for development (native or VM)
- **Make**: Build automation tool
- **Git**: Version control

### Setup Development Environment

1. **Install dependencies**:
   ```bash
   make mod-download
   ```

2. **Copy example configuration**:
   ```bash
   cp .env.example .env
   cp config.yaml.example config.yaml
   ```

3. **Set your development token** in `.env`:
   ```bash
   TOKEN_DEVICE=your_dev_token_here
   ```

4. **Build the agent**:
   ```bash
   make build
   ```

5. **Run tests**:
   ```bash
   make test
   ```

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check existing issues. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce** the issue
- **Expected vs actual behavior**
- **Environment details** (OS, Go version, agent version)
- **Logs** if applicable

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Clear title and description**
- **Use case** and motivation
- **Proposed solution** (if any)
- **Alternative solutions** considered

### Code Contributions

1. **Check existing issues** or create a new one
2. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes** following our coding standards
4. **Write or update tests** for your changes
5. **Run tests and linting**:
   ```bash
   make test
   make lint
   make vet
   ```
6. **Commit your changes** with clear messages:
   ```bash
   git commit -m "feat: add new feature"
   ```
7. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```
8. **Create a Pull Request**

## Coding Standards

### Go Style Guide

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting: `make fmt`
- Run `go vet`: `make vet`
- Use `golangci-lint`: `make lint`

### Code Organization

```
agent/
├── cmd/agent/          # Main application entry point
├── internal/           # Private application code
│   ├── agent/         # Core agent logic
│   ├── collector/     # Metric collectors
│   ├── sender/        # Data transmission
│   └── scanner/       # Port scanner
├── pkg/               # Public libraries
│   ├── config/        # Configuration management
│   ├── lock/          # Single-instance lock
│   └── models/        # Data models
└── scripts/           # Installation and service scripts
```

### Commit Message Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes (formatting, etc.)
- `refactor:` Code refactoring
- `perf:` Performance improvements
- `test:` Adding or updating tests
- `chore:` Maintenance tasks

Examples:
```
feat: add CPU temperature monitoring
fix: resolve memory leak in collector
docs: update installation instructions
refactor: simplify configuration loading
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test -v ./internal/collector/...
```

### Writing Tests

- Write unit tests for all new code
- Aim for >80% code coverage
- Use table-driven tests where appropriate
- Mock external dependencies

Example:
```go
func TestCollectCPU(t *testing.T) {
    tests := []struct {
        name    string
        want    float64
        wantErr bool
    }{
        {"valid", 50.0, false},
        {"error", 0, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Pull Request Process

1. **Update documentation** if needed
2. **Update CHANGELOG.md** under `[Unreleased]` section
3. **Ensure all tests pass** and code is linted
4. **Request review** from maintainers
5. **Address review comments** promptly
6. **Squash commits** if requested
7. **Wait for approval** from at least one maintainer

### PR Checklist

- [ ] Code follows project style guidelines
- [ ] Tests added/updated and passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Commit messages follow convention
- [ ] No merge conflicts
- [ ] CI/CD pipeline passes

## Release Process

Releases are managed by maintainers:

1. Update version in relevant files
2. Update CHANGELOG.md with release date
3. Create and push a version tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
4. GitHub Actions automatically builds and creates release
5. Announce release in relevant channels

## Questions?

- **Documentation**: Check [README.md](README.md) and [API.md](API.md)
- **Issues**: [GitHub Issues](https://github.com/monify-labs/agent/issues)
- **Discussions**: [GitHub Discussions](https://github.com/monify-labs/agent/discussions)
- **Email**: dev@monify.cloud

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Monify Agent! 🎉
