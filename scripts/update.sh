#!/bin/bash
#
# Monify Agent Update Script
# 
# Usage:
#   curl -sSL https://monify.cloud/update.sh | sudo bash
#
# This script:
# 1. Detects system architecture (amd64/arm64)
# 2. Downloads the latest agent binary
# 3. Stops existing agent service
# 4. Replaces the binary in /usr/local/bin
# 5. Restarts the agent service
#

set -e

# Configuration
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/monify"
SERVICE_NAME="monify"
BINARY_NAME="monify"
DOWNLOAD_BASE="https://github.com/monify-labs/agent/releases/latest/download"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Print functions
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_root() {
    if [ "$(id -u)" != "0" ]; then
        print_error "This script must be run as root"
        echo "Please run: curl -sSL https://monify.cloud/update.sh | sudo bash"
        exit 1
    fi
}

# Detect system architecture
detect_arch() {
    local arch
    arch=$(uname -m)
    case $arch in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *)
            print_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
}

# Check if already installed
check_installed() {
    if [ ! -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        print_error "Monify Agent is not installed in ${INSTALL_DIR}"
        print_info "Please run the installation script first:"
        print_info "curl -sSL https://monify.cloud/install.sh | sudo bash -s -- YOUR_TOKEN"
        exit 1
    fi
}

# Download and replace binary
update_binary() {
    local arch=$1
    local download_url="${DOWNLOAD_BASE}/${BINARY_NAME}-linux-${arch}"
    local temp_file="/tmp/${BINARY_NAME}_new"
    
    print_info "Downloading latest Monify Agent for linux/${arch}..."
    
    if command -v curl &> /dev/null; then
        curl -sSL -o "$temp_file" "$download_url" || {
            print_error "Failed to download from $download_url"
            exit 1
        }
    elif command -v wget &> /dev/null; then
        wget -q -O "$temp_file" "$download_url" || {
            print_error "Failed to download from $download_url"
            exit 1
        }
    else
        print_error "curl or wget is required"
        exit 1
    fi
    
    # Check if we got a valid binary (or at least something that exists)
    if [ ! -s "$temp_file" ]; then
        print_error "Downloaded binary is empty"
        exit 1
    fi

    # Make executable
    chmod +x "$temp_file"
    
    # Stop service before replacing
    if systemctl is-active --quiet $SERVICE_NAME; then
        print_info "Stopping $SERVICE_NAME service..."
        systemctl stop $SERVICE_NAME
    fi
    
    # Replace the binary
    mv "$temp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    
    print_success "Binary updated to latest version"
}

# Start the service
start_service() {
    print_info "Starting Monify Agent..."
    systemctl restart $SERVICE_NAME
    
    # Wait for service to stabilize
    sleep 2
    
    if systemctl is-active --quiet $SERVICE_NAME; then
        print_success "Monify Agent is running successfully!"
    else
        print_error "Failed to start Monify Agent after update"
        echo "Check logs: journalctl -u $SERVICE_NAME --no-pager -n 20"
        exit 1
    fi
}

# Main update flow
main() {
    echo ""
    echo "======================================"
    echo "  Monify Agent Updater"
    echo "======================================"
    echo ""
    
    check_root
    check_installed
    
    local arch
    arch=$(detect_arch)
    print_info "Architecture: $arch"
    
    # Get current version for display
    local current_version
    current_version=$("${INSTALL_DIR}/${BINARY_NAME}" --version 2>/dev/null | head -1 || echo "unknown")
    print_info "Current version: $current_version"
    
    update_binary "$arch"
    start_service
    
    local new_version
    new_version=$("${INSTALL_DIR}/${BINARY_NAME}" --version 2>/dev/null | head -1 || echo "unknown")
    
    echo ""
    echo "======================================"
    echo -e "${GREEN}Update completed successfully!${NC}"
    echo "======================================"
    echo ""
    echo "Old Version: $current_version"
    echo "New Version: $new_version"
    echo "Status:      $(systemctl is-active $SERVICE_NAME)"
    echo ""
}

main "$@"
