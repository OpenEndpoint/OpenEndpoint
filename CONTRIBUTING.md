# Contributing to OpenEndpoint

Thank you for your interest in contributing to OpenEndpoint!

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/openendpoint.git`
3. Create a feature branch: `git checkout -b feature/my-feature`

## Development Setup

```bash
# Install Go 1.22+
go version

# Download dependencies
go mod download

# Run tests
make test

# Build
make build

# Run locally
make run
```

## Code Style

- Follow Go standard formatting (`go fmt`)
- Use meaningful variable names
- Add comments for public APIs
- Write tests for new features

## Pull Request Process

1. Update documentation for any changed functionality
2. Add tests for new features
3. Ensure all tests pass: `go test ./...`
4. Update the CHANGELOG.md
5. Submit a pull request

## Reporting Issues

Use GitHub issues to report bugs or request features. Include:
- Clear description
- Steps to reproduce
- Environment details
- Any relevant logs

## License

By contributing, you agree that your contributions will be licensed under Apache 2.0.
