# Testing Guide

This guide explains how to run tests for the Jyogi Member Authentication System.

## Types of Tests

The project includes two main types of tests:

1. **Unit Tests**: Test individual logic within `internal` and `pkg` packages.
2. **Integration Tests**: Located in the `tests/integration` directory, these test the entire flow by calling actual API endpoints.

## Running Tests

### Running in Docker (Recommended)

Using Docker allows you to run tests in an isolated environment, avoiding issues caused by local environment differences.

```bash
make test
```

This command runs all tests inside a test container using `docker-compose run`.

### Running Locally

If you have a Go environment set up locally, you can run tests directly.

```bash
make test-local
```

### Testing Specific Packages

You can test specific packages using standard Go commands.

```bash
go test ./pkg/discord/...
```

## Test Coverage

The `make test` command automatically generates a coverage report (`coverage.txt`).
To view the coverage, use the following command:

```bash
go tool cover -html=coverage.txt
```

## CI/CD

GitHub Actions automatically runs tests when a Pull Request is created.
