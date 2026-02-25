# OpenEndpoint QuickStart

One-command setup to get OpenEndpoint running locally.

## Prerequisites

- Docker and Docker Compose installed
- Or: OpenEndpoint binary installed

## Quick Start with Docker (Recommended)

```bash
# Clone or download the quickstart files
git clone https://github.com/OpenEndpoint/OpenEndpoint.git
cd OpenEndpoint/examples/quickstart

# Start OpenEndpoint
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f openendpoint
```

OpenEndpoint will be available at:
- **S3 API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health

## Default Credentials

- **Access Key**: `demo-access-key`
- **Secret Key**: `demo-secret-key`

## Test with AWS CLI

```bash
# Configure AWS CLI
aws configure --profile openendpoint
# AWS Access Key ID: demo-access-key
# AWS Secret Access Key: demo-secret-key
# Default region: us-east-1
# Default output: json

# List buckets
aws --profile openendpoint --endpoint-url http://localhost:8080 s3 ls

# Create a bucket
aws --profile openendpoint --endpoint-url http://localhost:8080 s3 mb s3://my-test-bucket

# Upload a file
aws --profile openendpoint --endpoint-url http://localhost:8080 s3 cp hello.txt s3://my-test-bucket/

# List objects
aws --profile openendpoint --endpoint-url http://localhost:8080 s3 ls s3://my-test-bucket/
```

## Test with cURL

```bash
# Health check
curl http://localhost:8080/health

# List buckets
curl -X GET \
  -H "Authorization: AWS demo-access-key:demo-secret-key" \
  http://localhost:8080/
```

## Stop and Clean Up

```bash
# Stop services
docker-compose down

# Stop and remove data
docker-compose down -v
```

## Quick Start with Binary

If you prefer running the binary directly:

```bash
# Install OpenEndpoint
curl -fsSL https://raw.githubusercontent.com/OpenEndpoint/OpenEndpoint/main/install.sh | bash

# Start with sample config
openep --config config.yaml
```
