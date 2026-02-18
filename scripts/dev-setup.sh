#!/bin/bash
set -e

# OpenEndpoint Development Setup

echo "Setting up OpenEndpoint development environment..."

# Check Go version
GO_VERSION=$(go version | grep -oP 'go\d+\.\d+')
GO_MAJOR=$(echo $GO_VERSION | grep -oP '\d+')
GO_MINOR=$(echo $GO_VERSION | grep -oP '\.\d+' | tr -d '.')

if [ "$GO_MAJOR" -lt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 22 ]); then
    echo "Error: Go 1.22+ required. Current version: $GO_VERSION"
    exit 1
fi

echo "Go version OK: $GO_VERSION"

# Download dependencies
echo "Downloading dependencies..."
go mod download

# Install development tools
echo "Installing development tools..."
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Create necessary directories
echo "Creating data directories..."
mkdir -p /tmp/openendpoint/data

# Create default config
echo "Creating default configuration..."
cat > /tmp/openendpoint/config.yaml << EOF
server:
  host: "0.0.0.0"
  port: 9000

storage:
  data_dir: "/tmp/openendpoint/data"
  storage_backend: "flatfile"

auth:
  secret_key: "minioadmin"
  access_key: "minioadmin"

log_level: "debug"
EOF

echo "Development environment ready!"
echo ""
echo "To start the server:"
echo "  go run ./cmd/openep server -c /tmp/openendpoint/config.yaml"
echo ""
echo "Or use the Makefile:"
echo "  make run"
