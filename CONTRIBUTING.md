# Contributing to Synapse

Thank you for your interest in contributing to Synapse! This document provides guidelines for contributing to the project.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/synapse.git`
3. Add upstream remote: `git remote add upstream https://github.com/princetheprogrammer/synapse.git`
4. Create a feature branch: `git checkout -b feature/your-feature-name`

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional, for build automation)

### Building

```bash
# Using Make
make build

# Or directly with Go
go build -o bin/synapse ./cmd/synapse
```

### Running Tests

```bash
# Using Make
make test

# With verbose output
make test-v

# With coverage report
make coverage

# Or directly with Go
go test ./...
```

## Code Standards

### Style Guide

- Follow the guidelines in `docs/DEVELOPMENT.md`
- Use `gofmt` for formatting (run `make fmt`)
- Run `golangci-lint` before committing (run `make lint`)
- Write meaningful commit messages following Conventional Commits

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `build`: Build system changes
- `perf`: Performance improvements

**Examples:**
```
feat(p2p): add peer discovery mechanism

Implement mDNS-based peer discovery for local network.
Allows nodes to find each other without manual configuration.

Closes #42
```

### Testing Requirements

- Write unit tests for new features
- Maintain minimum 80% code coverage
- All tests must pass before submitting PR
- Use table-driven tests for multiple scenarios
- Include edge cases and error conditions

### Code Review Process

1. Ensure all tests pass: `make test`
2. Format code: `make fmt`
3. Run linter: `make lint` (if available)
4. Commit changes with descriptive messages
5. Push to your fork
6. Create a Pull Request

### Pull Request Guidelines

- Provide a clear description of changes
- Reference related issues (e.g., "Fixes #123")
- Ensure CI/CD pipeline passes
- Keep PRs focused (one feature/fix per PR)
- Update documentation if needed
- Add tests for new functionality

## Project Structure

```
synapse/
├── cmd/synapse/       # Main application entry point
├── pkg/               # Public packages
│   ├── node/          # Core node implementation
│   ├── p2p/           # P2P networking
│   ├── storage/       # Data persistence
│   └── ...
├── internal/          # Private packages
│   ├── config/        # Configuration
│   └── logger/        # Logging
├── docs/              # Documentation
├── scripts/           # Build and utility scripts
└── tests/             # Integration tests
```

## Development Phases

The project is developed in 10 phases. Check `docs/PHASE_*.md` for details on each phase and current progress.

## Reporting Issues

### Bug Reports

When reporting bugs, please include:
- Go version (`go version`)
- Operating system and architecture
- Steps to reproduce
- Expected behavior
- Actual behavior
- Relevant logs or error messages

### Feature Requests

When requesting features:
- Describe the problem you're trying to solve
- Explain why existing features don't address it
- Provide examples or use cases
- Consider if it aligns with project goals

## Communication

- **Issues**: For bug reports and feature requests
- **Pull Requests**: For code contributions
- **Discussions**: For questions and general discussion (if enabled)

## License

By contributing to Synapse, you agree that your contributions will be licensed under the MIT License.

## Recognition

Contributors will be recognized in the project's README and release notes.

Thank you for contributing to Synapse! 🚀
