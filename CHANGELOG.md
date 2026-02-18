# Changelog

All notable changes to this project will be documented in this file.

## [1.0.0] - 2026-02-18

### Added
- Production-ready S3-compatible object storage platform
- Full AWS Signature V2/V4 authentication support
- S3-compatible REST API with complete bucket and object operations
- Multipart upload support with initiate, upload part, complete, abort operations
- Object versioning with full version history
- Object locking with GOVERNANCE and COMPLIANCE modes
- Lifecycle policies for automated transitions and expiration
- Replication configuration for bucket-level disaster recovery
- Object tagging for categorization and filtering
- Server-side AES-256-GCM encryption
- Presigned URLs for time-limited access
- CORS configuration for web applications
- Bucket policies for fine-grained access control

### Storage Backends
- Flat file backend (default) - Simple and direct file storage
- Packed volume backend - High density storage with volume packing

### Metadata Stores
- Pebble backend (default) - High-performance Go key-value store
- BBolt backend - Embedded key-value database

### Performance & Reliability
- Connection pooling for optimized concurrent access
- Request batching for improved throughput
- LRU caching for frequently accessed data
- Rate limiting with token bucket algorithm
- Structured JSON logging with rotation
- Prometheus-compatible metrics endpoint

### Operations
- CLI tool for complete bucket and object management
- Web dashboard for visual interface
- Metrics dashboard for performance monitoring
- Health and readiness probes for Kubernetes
- Docker Compose and Helm chart deployment

## [0.1.0] - 2026-01-15

### Added
- Initial release
- Basic S3 API operations (Put, Get, Delete, List)
- Bucket management
- Simple authentication
- Docker Compose setup
- Helm chart for Kubernetes
