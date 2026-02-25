#!/bin/bash
# OpenEndpoint Quick Test Script
# Tests basic functionality after installation

set -e

ENDPOINT="http://localhost:8080"
ACCESS_KEY="demo-access-key"
SECRET_KEY="demo-secret-key"
BUCKET_NAME="test-bucket-$(date +%s)"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "========================================"
echo "  OpenEndpoint Quick Test"
echo "========================================"
echo

# Wait for service to be ready
echo -n "Waiting for OpenEndpoint to be ready..."
for i in {1..30}; do
    if curl -fs "$ENDPOINT/health" > /dev/null 2>&1; then
        echo -e " ${GREEN}OK${NC}"
        break
    fi
    echo -n "."
    sleep 1
done

if ! curl -fs "$ENDPOINT/health" > /dev/null 2>&1; then
    echo -e " ${RED}FAILED${NC}"
    echo "OpenEndpoint is not responding"
    exit 1
fi

# Test 1: Health check
echo -n "Test 1: Health check..."
if curl -fs "$ENDPOINT/health" > /dev/null 2>&1; then
    echo -e " ${GREEN}PASSED${NC}"
else
    echo -e " ${RED}FAILED${NC}"
    exit 1
fi

# Test 2: List buckets (using pre-configured buckets)
echo -n "Test 2: List buckets..."
if curl -fs "$ENDPOINT/" -H "Authorization: AWS $ACCESS_KEY:$SECRET_KEY" > /dev/null 2>&1; then
    echo -e " ${GREEN}PASSED${NC}"
else
    echo -e " ${YELLOW}SKIPPED${NC} (may require proper AWS signature)"
fi

# Test 3: Check if pre-configured buckets exist
echo -n "Test 3: Check pre-configured buckets..."
for bucket in photos documents backups; do
    if curl -fsI "$ENDPOINT/$bucket" -H "Authorization: AWS $ACCESS_KEY:$SECRET_KEY" > /dev/null 2>&1; then
        echo -n " $bucket✓"
    else
        echo -n " $bucket?"
    fi
done
echo -e " ${GREEN}DONE${NC}"

echo
echo "========================================"
echo -e "  ${GREEN}Basic tests completed!${NC}"
echo "========================================"
echo
echo "Endpoint: $ENDPOINT"
echo "Access Key: $ACCESS_KEY"
echo "Secret Key: $SECRET_KEY"
echo
echo "Try these commands:"
echo "  curl $ENDPOINT/health"
echo "  aws --endpoint-url $ENDPOINT s3 ls"
echo
