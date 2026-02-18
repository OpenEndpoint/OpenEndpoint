# OpenEndpoint Commands Quick Reference

## Server Commands

```bash
# Start server
openep server

# With custom config
openep server -c /path/to/config.yaml
```

## Bucket Commands

```bash
# Create bucket
openep bucket create my-bucket

# List buckets
openep bucket ls

# Bucket info
openep bucket info my-bucket

# Delete bucket
openep bucket rm my-bucket
# Or with force
openep bucket rm my-bucket -f
```

## Object Commands

```bash
# Upload file
openep object put local-file.txt s3://bucket/key

# Download file
openep object get s3://bucket/key local-file.txt

# List objects
openep object ls s3://bucket
openep object ls s3://bucket/prefix
openep object ls s3://bucket -r  # recursive

# Delete object
openep object rm s3://bucket/key
# Or with force
openep object rm s3://bucket/key -f

# Copy object
openep object cp s3://src/bucket/key s3://dst/bucket/key
```

## Admin Commands

```bash
# Server info
openep admin info

# Server stats
openep admin stats
```

## Configuration

```bash
# Get config value
openep config get server.port
```

## Docker

```bash
# Start with Docker Compose
cd deploy/docker
docker-compose up -d

# Stop
docker-compose down

# Build image
docker build -t openendpoint/openendpoint:latest .
```

## Make Commands

```bash
make build        # Build binary
make test        # Run tests
make run         # Run server
make docker-up   # Start Docker
make clean       # Clean artifacts
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| OPENEP_SERVER_HOST | Server host | 0.0.0.0 |
| OPENEP_SERVER_PORT | Server port | 9000 |
| OPENEP_STORAGE_DATA_DIR | Data directory | /data |
| OPENEP_AUTH_SECRET_KEY | Secret key | (required) |
| OPENEP_AUTH_ACCESS_KEY | Access key | (required) |
| OPENEP_LOG_LEVEL | Log level | info |

## S3 API Endpoints

| Endpoint | Description |
|---------|-------------|
| http://localhost:9000/s3/ | S3 API |
| http://localhost:9000/_mgmt/ | Management API |
| http://localhost:9000/health | Health check |
| http://localhost:9000/ready | Readiness check |
| http://localhost:9000/metrics | Prometheus metrics |

## AWS CLI Examples

```bash
# Configure
aws configure
# Enter access key and secret key

# List buckets
aws --endpoint-url=http://localhost:9000 s3 ls

# Create bucket
aws --endpoint-url=http://localhost:9000 s3 mb s3://my-bucket

# Upload file
aws --endpoint-url=http://localhost:9000 s3 cp file.txt s3://my-bucket/

# Download file
aws --endpoint-url=http://localhost:9000 s3 cp s3://my-bucket/file.txt ./

# List objects
aws --endpoint-url=http://localhost:9000 s3 ls s3://my-bucket/
```
