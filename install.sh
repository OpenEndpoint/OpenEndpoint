#!/bin/bash
# OpenEndpoint One-Click Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/OpenEndpoint/OpenEndpoint/main/install.sh | bash

set -e

REPO="OpenEndpoint/OpenEndpoint"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.openendpoint"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect OS and Architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac

    case "$OS" in
        linux|darwin)
            PLATFORM="${OS}-${ARCH}"
            ;;
        mingw*|cygwin*|msys*)
            PLATFORM="windows-${ARCH}"
            ;;
        *)
            echo -e "${RED}Unsupported OS: $OS${NC}"
            exit 1
            ;;
    esac
}

# Get latest release version
get_latest_version() {
    curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
}

# Download and install
download_and_install() {
    local version=$1
    local platform=$2

    if [[ "$platform" == windows-* ]]; then
        BINARY="openep-${platform}.exe"
        OUTPUT_NAME="openep.exe"
    else
        BINARY="openep-${platform}"
        OUTPUT_NAME="openep"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${version}/${BINARY}.tar.gz"

    if [[ "$platform" == windows-* ]]; then
        DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${version}/openep-${version}-windows-amd64.zip"
    fi

    echo -e "${YELLOW}Downloading OpenEndpoint ${version} for ${platform}...${NC}"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR"

    if [[ "$platform" == windows-* ]]; then
        curl -fsSL -o openep.zip "$DOWNLOAD_URL"
        unzip -q openep.zip
        mv openep-windows-* "$OUTPUT_NAME"
    else
        curl -fsSL -o openep.tar.gz "$DOWNLOAD_URL"
        tar -xzf openep.tar.gz
        mv openep-* "$OUTPUT_NAME"
    fi

    # Install binary
    if [[ "$platform" == windows-* ]]; then
        mkdir -p "$HOME/bin"
        mv "$OUTPUT_NAME" "$HOME/bin/"
        echo -e "${GREEN}OpenEndpoint installed to $HOME/bin/$OUTPUT_NAME${NC}"
        echo -e "${YELLOW}Add $HOME/bin to your PATH${NC}"
    else
        echo -e "${YELLOW}Installing to ${INSTALL_DIR} (may require sudo)...${NC}"
        sudo mv "$OUTPUT_NAME" "${INSTALL_DIR}/"
        sudo chmod +x "${INSTALL_DIR}/${OUTPUT_NAME}"
        echo -e "${GREEN}OpenEndpoint installed successfully!${NC}"
    fi

    # Cleanup
    cd -
    rm -rf "$TMP_DIR"
}

# Create sample configuration
setup_sample() {
    echo -e "${YELLOW}Setting up sample configuration...${NC}"

    mkdir -p "$CONFIG_DIR"

    cat > "$CONFIG_DIR/config.yaml" << 'EOF'
server:
  host: "0.0.0.0"
  port: 8080

storage:
  type: "flatfile"
  path: "./data"

logging:
  level: "info"
  format: "json"

# Sample buckets
buckets:
  - name: "my-bucket"
    region: "us-east-1"
  - name: "uploads"
    region: "us-east-1"

# Sample user
credentials:
  - access_key: "demo-access-key"
    secret_key: "demo-secret-key"
    buckets:
      - "my-bucket"
      - "uploads"
EOF

    echo -e "${GREEN}Sample configuration created at ${CONFIG_DIR}/config.yaml${NC}"
}

# Create sample data
create_sample_data() {
    echo -e "${YELLOW}Creating sample data...${NC}"

    SAMPLE_DIR="$CONFIG_DIR/sample-data"
    mkdir -p "$SAMPLE_DIR"

    # Create sample text file
    echo "Hello from OpenEndpoint!" > "$SAMPLE_DIR/hello.txt"

    # Create sample JSON
    cat > "$SAMPLE_DIR/sample.json" << 'EOF'
{
  "message": "Welcome to OpenEndpoint",
  "version": "1.0.0",
  "features": [
    "S3-compatible API",
    "Multi-platform support",
    "Easy deployment"
  ]
}
EOF

    echo -e "${GREEN}Sample data created in ${SAMPLE_DIR}${NC}"
}

# Main installation flow
main() {
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  OpenEndpoint One-Click Installer${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo

    detect_platform
    echo -e "Detected platform: ${YELLOW}${PLATFORM}${NC}"

    VERSION=$(get_latest_version)
    if [ -z "$VERSION" ]; then
        echo -e "${RED}Failed to get latest version${NC}"
        exit 1
    fi
    echo -e "Latest version: ${YELLOW}${VERSION}${NC}"

    download_and_install "$VERSION" "$PLATFORM"
    setup_sample
    create_sample_data

    echo
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  Installation Complete!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo
    echo -e "To start OpenEndpoint:"
    echo -e "  ${YELLOW}openep --config ${CONFIG_DIR}/config.yaml${NC}"
    echo
    echo -e "Sample data location:"
    echo -e "  ${YELLOW}${CONFIG_DIR}/sample-data/${NC}"
    echo
    echo -e "API Endpoint:"
    echo -e "  ${YELLOW}http://localhost:8080${NC}"
    echo
    echo -e "Default credentials:"
    echo -e "  Access Key: ${YELLOW}demo-access-key${NC}"
    echo -e "  Secret Key: ${YELLOW}demo-secret-key${NC}"
    echo
}

main "$@"
