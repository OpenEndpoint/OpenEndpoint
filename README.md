# OpenEndpoint

**A production-ready, S3-compatible object storage platform built with Go.**

## Status: v1.0 - Production Ready

OpenEndpoint is a developer-first, self-hosted object storage platform compatible with the S3 API. It provides high performance, scalability, and enterprise features while remaining simple to deploy and operate.

## Features

### Core Storage
- **S3 Compatible API** - Full compatibility with AWS S3 API and SDKs
- **Multiple Storage Backends** - Flat file and packed volume storage
- **Metadata Stores** - Pebble and BBolt support
- **Object Locking** - GOVERNANCE and COMPLIANCE modes
- **Versioning** - Full object versioning support
- **Multipart Uploads** - Large file uploads with parallel parts

### Data Management
- **Lifecycle Policies** - Automated transitions and expiration
- **Replication Configuration** - Bucket-level replication rules
- **Object Tagging** - Categorize and filter objects
- **Quota Management** - Per-bucket storage limits

### Security
- **AWS Signature V2/V4** - Industry-standard authentication
- **Presigned URLs** - Time-limited access
- **Encryption** - AES-256-GCM server-side encryption
- **CORS** - Cross-origin resource sharing
- **Bucket Policies** - Fine-grained access control

### Performance & Reliability
- **Connection Pooling** - Optimized concurrent access
- **Request Batching** - Improved throughput
- **LRU Caching** - Frequently accessed data
- **Rate Limiting** - Token bucket algorithm

### Operations
- **CLI Tools** - Full command-line management
- **Metrics Dashboard** - Prometheus-compatible metrics
- **Web Dashboard** - Visual interface
- **Health Checks** - Readiness and liveness probes
- **Logging** - Structured JSON logging with rotation

## Supported S3 Operations

| Category | Operations |
|----------|------------|
| Buckets | Create, Delete, List, Head |
| Objects | Put, Get, Delete, Head, Copy |
| Listing | ListObjectsV2, ListMultipartUploads, ListParts |
| Multipart | Initiate, UploadPart, Complete, Abort |
| Advanced | GetObjectAttributes, SelectObjectContent |

## Quick Start

### Using Docker

```bash
# Start with Docker Compose
cd deploy/docker
cp .env.example .env
docker-compose up -d
```

### Using Binary

```bash
# Download the latest release
wget https://github.com/OpenEndpoint/openendpoint/releases/latest/download/openep-linux-amd64

# Make executable
chmod +x openep-linux-amd64

# Start the server
./openep server
```

## Configuration

OpenEndpoint can be configured via YAML file, environment variables, or CLI flags.

### Configuration File

```bash
# Copy example config
cp config.example.yaml config.yaml

# Edit configuration
nano config.yaml
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENEP_SERVER_HOST` | Server bind address | `0.0.0.0` |
| `OPENEP_SERVER_PORT` | S3 API port | `9000` |
| `OPENEP_STORAGE_DATA_DIR` | Data directory | `/var/lib/openendpoint` |
| `OPENEP_STORAGE_BACKEND` | Storage backend | `flatfile` |
| `OPENEP_AUTH_SECRET_KEY` | Secret key | (required) |
| `OPENEP_AUTH_ACCESS_KEY` | Access key | `minioadmin` |
| `OPENEP_LOG_LEVEL` | Log level | `info` |

### CLI Options

```bash
# Start with custom config
openep server --config /path/to/config.yaml

# Start with custom port
openep server --port 9000 --host 0.0.0.0

# Start with custom data directory
openep server --data-dir /data/openendpoint
```

## Deployment Options

### Docker

```bash
# Start with Docker Compose
cd deploy/docker
cp .env.example .env
docker-compose up -d
```

### Kubernetes (Helm)

```bash
# Add Helm repository
helm repo add openendpoint https://openendpoint.github.io/helm-charts
helm install openendpoint openendpoint/openendpoint
```

### Binary

```bash
# Download the latest release
wget https://github.com/OpenEndpoint/openendpoint/releases/latest/download/openep-linux-amd64

# Make executable
chmod +x openep-linux-amd64

# Start the server
./openep server
```

## CLI Usage

```bash
# Start server
openep server

# Create a bucket
openep bucket create my-bucket

# List buckets
openep bucket ls

# Upload a file
openep object put local-file.txt my-bucket/remote-file.txt

# Download a file
openep object get my-bucket/remote-file.txt local-file.txt

# List objects
openep object ls my-bucket

# Delete an object
openep object rm my-bucket/remote-file.txt
```

## S3 API

The S3 API is available at `http://localhost:9000`.

### API Endpoints

| Endpoint | Description |
|----------|-------------|
| `/s3/` | S3 API (REST) |
| `/` | Web Dashboard |
| `/_dashboard/metrics` | Metrics Dashboard |
| `/health` | Health Check |
| `/ready` | Readiness Check |
| `/metrics` | Prometheus Metrics |
| `/_mgmt/` | Management API |

### AWS CLI Example

```bash
# Configure AWS CLI
aws configure
# Enter your access key and secret key
# Set default region to us-east-1

# Create a bucket
aws --endpoint-url=http://localhost:9000 s3 mb s3://my-bucket

# Upload a file
aws --endpoint-url=http://localhost:9000 s3 cp local-file.txt s3://my-bucket/

# List objects
aws --endpoint-url=http://localhost:9000 s3 ls s3://my-bucket/

# Download a file
aws --endpoint-url=http://localhost:9000 s3 cp s3://my-bucket/file.txt ./
```

## Development

### Prerequisites

- Go 1.22+
- Docker (for building)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/OpenEndpoint/openendpoint.git
cd openendpoint

# Install dependencies
go mod tidy

# Build
make build

# Run tests
make test

# Run with custom config
make run CONFIG=/path/to/config.yaml
```

### Project Structure

```
.
├── cmd/openep/           # CLI and server entry points
├── internal/
│   ├── api/              # S3 HTTP handlers
│   ├── auth/             # AWS Signature authentication
│   ├── cache/            # LRU caching
│   ├── config/           # Configuration
│   ├── dashboard/         # Web UI
│   ├── encryption/       # AES-256-GCM encryption
│   ├── engine/           # Core object service
│   ├── events/           # Event notifications
│   ├── health/           # Health checks
│   ├── iam/              # Policies and ACL
│   ├── lifecycle/        # Lifecycle processor
│   ├── locking/          # Object lock
│   ├── logging/          # Logging
│   ├── metadata/          # Pebble/BBolt stores
│   ├── middleware/        # HTTP middleware
│   ├── mgmt/             # Management API
│   ├── quota/            # Quota management
│   ├── ratelimit/        # Rate limiting
│   ├── storage/          # Storage backends
│   ├── tags/             # Object tagging
│   ├── telemetry/        # Metrics
│   ├── websocket/        # Real-time events
│   └── website/          # Static website hosting
├── pkg/
│   ├── client/           # Go SDK
│   ├── s3types/         # S3 types
│   └── util/            # Utilities
├── deploy/
│   ├── docker/           # Docker files
│   └── helm/             # Kubernetes Helm chart
├── docs/                 # Documentation
└── test/                 # Tests
```

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please read our [contributing guidelines](CONTRIBUTING.md) first.

## Documentation

For more information, see:
- [Error Codes](docs/ERROR_CODES.md)
- [TASKS.md](docs/TASKS.md) - Development progress
- [QUICKREF.md](docs/QUICKREF.md) - Quick reference guide

---

**Maintained by [ersinkoc](https://github.com/ersinkoc)**

**Repository**: [github.com/openendpoint/openendpoint](https://github.com/openendpoint/openendpoint)
