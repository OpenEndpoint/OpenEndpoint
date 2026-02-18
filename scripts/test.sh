#!/bin/bash
set -e

# OpenEndpoint Test Script

echo "Running OpenEndpoint tests..."

# Unit tests
echo "Running unit tests..."
go test -v -short ./...

# Integration tests (if available)
if [ -d "test/integration" ]; then
    echo "Running integration tests..."
    go test -v ./test/integration/...
fi

# Code coverage
echo "Checking code coverage..."
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

echo "Tests complete!"
