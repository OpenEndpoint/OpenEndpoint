# OpenEndpoint Makefile

.PHONY: all build test clean install run deps

BINARY_NAME=openep
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

# Default target
all: deps build test

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build the binary
build:
	CGO_ENABLED=0 go build ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd/openep

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-amd64 ./cmd/openep
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-arm64 ./cmd/openep
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-amd64 ./cmd/openep
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-arm64 ./cmd/openep
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ${LDFLAGS} -o bin/${BINARY_NAME}-windows-amd64.exe ./cmd/openep

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Install binary to GOPATH
install:
	go install ${LDFLAGS} ./cmd/openep

# Run the server
run:
	go run ./cmd/openep

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...
	gofmt -s -w .

# Run docker build
docker-build:
	docker build -t openendpoint/openendpoint:${VERSION} -f deploy/docker/Dockerfile .

# Run docker compose
docker-up:
	docker-compose -f deploy/docker/docker-compose.yml up -d

# Generate code
generate:
	go generate ./...

# Show help
help:
	@echo "OpenEndpoint Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  all           - Build and test (default)"
	@echo "  deps          - Download dependencies"
	@echo "  build         - Build binary"
	@echo "  build-all     - Build for all platforms"
	@echo "  test          - Run tests"
	@echo "  test-unit     - Run unit tests only"
	@echo "  bench         - Run benchmarks"
	@echo "  clean         - Clean artifacts"
	@echo "  install       - Install binary"
	@echo "  run           - Run server"
	@echo "  run-dev       - Run server with dev config"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-up     - Start with Docker Compose"
	@echo "  docker-down   - Stop Docker Compose"
	@echo "  setup-dev     - Setup development environment"
	@echo "  release       - Create release builds"
