# Contributing to Go Orchestrator

Thank you for your interest in contributing to Go Orchestrator! This document provides guidelines and information for contributors.

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/AnasImloul/go-orchestrator.git`
3. Create a feature branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `go test ./...`
6. Run linting: `golangci-lint run`
7. Commit your changes: `git commit -m "Add your feature"`
8. Push to your fork: `git push origin feature/your-feature-name`
9. Create a Pull Request

## Development Guidelines

### Code Style

- Follow standard Go formatting: `gofmt -s -w .`
- Use `golint` and `golangci-lint` for code quality
- Write clear, self-documenting code
- Add comments for exported functions and types

### Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/) for consistent commit messages:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `ci`: CI/CD changes
- `build`: Build system changes

**Examples:**
```
feat(di): add support for scoped services
fix(lifecycle): resolve deadlock in component disposal
docs(api): update orchestrator interface documentation
```

### Release Process

This project uses automated semantic versioning:

- **Patch releases** (v1.0.1): Bug fixes, documentation updates
- **Minor releases** (v1.1.0): New features, non-breaking changes
- **Major releases** (v2.0.0): Breaking changes

Releases are automatically created when:
- Code is pushed to the `main` branch
- Commit messages follow conventional commit format
- All tests pass

**Breaking Changes:**
To trigger a major version bump, include `BREAKING CHANGE:` in your commit message or use `!` after the type (e.g., `feat!: remove deprecated API`).

### Testing

- Write unit tests for new functionality
- Maintain or improve test coverage
- Use table-driven tests where appropriate
- Test error conditions and edge cases

### Documentation

- Update README.md for user-facing changes
- Add or update API documentation
- Include examples for new features
- Update CHANGELOG.md for significant changes

## Project Structure

```
go-orchestrator/
├── cmd/              # Example applications
├── internal/         # Private implementation (not importable)
│   ├── di/          # Dependency injection
│   ├── lifecycle/   # Lifecycle management
│   └── logger/      # Logging interface
├── pkg/             # Public API (importable)
│   └── orchestrator/
├── examples/        # Usage examples
├── docs/           # Documentation
└── tests/          # Integration tests
```

## Pull Request Process

1. Ensure your code follows the project's coding standards
2. Add tests for new functionality
3. Update documentation as needed
4. Ensure all tests pass
5. Request review from maintainers

## Reporting Issues

When reporting issues, please include:

- Go version
- Operating system
- Steps to reproduce
- Expected behavior
- Actual behavior
- Any relevant error messages

## License

By contributing to this project, you agree that your contributions will be licensed under the MIT License.
