FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /openep ./cmd/openep

# Final stage
FROM alpine:3.19

# Install CA certificates and runas
RUN apk add --no-cache ca-certificates shadow && \
    usermod -u 1000 nobody && \
    groupmod -g 1000 nobody

# Set working directory
WORKDIR /data

# Create volume for data
VOLUME ["/data"]

# Copy binary from builder
COPY --from=builder /openep /usr/local/bin/openep

# Expose ports
# S3 API
EXPOSE 9000
# Management API
EXPOSE 9001
# Metrics
EXPOSE 9090

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/openep"]
CMD ["server"]
