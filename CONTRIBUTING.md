# Contributing to go-alexa

Thank you for your interest in contributing to go-alexa! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project adheres to a code of conduct that all contributors are expected to follow. Please be respectful and constructive in all interactions.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/portpowered/go-alexa.git`
3. Create a branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `go test ./...`
6. Ensure code passes linting: `golangci-lint run`
7. Commit your changes: `git commit -am 'Add some feature'`
8. Push to your fork: `git push origin feature/your-feature-name`
9. Submit a pull request

## Development Guidelines

### Code Style

- Follow Go's standard formatting: `gofmt` or `go fmt`
- Follow Go naming conventions
- Write clear, concise code
- Add comments for exported functions and types
- Keep functions focused and small

### Testing

- Write tests for all new features
- Maintain or improve test coverage
- Use table-driven tests where appropriate
- Mock external dependencies in unit tests

### Commit Messages

- Use clear, descriptive commit messages
- Reference issue numbers if applicable
- Use imperative mood: "Add feature" not "Added feature"

## Pull Request Process

1. Ensure all tests pass
2. Ensure code passes linting
3. Update documentation as needed
4. Add tests for new features
5. Ensure your branch is up to date with main
6. Submit the pull request with a clear description

## Reporting Issues

When reporting issues, please include:

- Description of the issue
- Steps to reproduce
- Expected behavior
- Actual behavior
- Go version
- Any relevant error messages or logs

## Feature Requests

Feature requests are welcome! Please open an issue describing:

- The feature you'd like to see
- Use case or motivation
- Any implementation ideas (optional)

## Questions

If you have questions, please open an issue with the "question" label.

Thank you for contributing to go-alexa!

