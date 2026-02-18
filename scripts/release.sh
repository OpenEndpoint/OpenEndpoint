#!/bin/bash
set -e

# OpenEndpoint Release Script

VERSION=${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

echo "Building OpenEndpoint v${VERSION}"

# Create release directory
mkdir -p release

# Build for different platforms
echo "Building for linux/amd64..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
    -ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -s -w" \
    -o release/openendpoint-linux-amd64 \
    ./cmd/openep

echo "Building for linux/arm64..."
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build \
    -ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -s -w" \
    -o release/openendpoint-linux-arm64 \
    ./cmd/openep

echo "Building for darwin/amd64..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build \
    -ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -s -w" \
    -o release/openendpoint-darwin-amd64 \
    ./cmd/openep

echo "Building for darwin/arm64..."
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build \
    -ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -s -w" \
    -o release/openendpoint-darwin-arm64 \
    ./cmd/openep

echo "Building for windows/amd64..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build \
    -ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -s -w" \
    -o release/openendpoint-windows-amd64.exe \
    ./cmd/openep

# Create checksums
echo "Creating checksums..."
cd release
sha256sum openendpoint-* > SHA256SUMS
cd ..

# Build Docker image
echo "Building Docker image..."
docker build -t openendpoint/openendpoint:${VERSION} -t openendpoint/openendpoint:latest deploy/docker/

echo "Release complete!"
echo ""
echo "Artifacts in release/:"
ls -la release/
