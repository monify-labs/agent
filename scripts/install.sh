#!/bin/bash
#
# Monify Agent Installation Script
# 
# Usage:
#   curl -sSL https://monify.cloud/install.sh | sudo bash -s -- YOUR_TOKEN
#
# This script:
# 1. Detects system architecture (amd64/arm64)
# 2. Downloads the latest agent binary
# 3. Installs to /usr/local/bin
# 4. Creates systemd service
# 5. Saves token and starts the agent automatically
#

set -e

# Configuration
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/monify"
LOG_DIR="/var/log/monify"
SERVICE_FILE="/etc/systemd/system/monify.service"
BINARY_NAME="monify"
DOWNLOAD_BASE="https://github.com/monify-labs/agent/releases/latest/download"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Show usage
show_usage() {
    echo ""
    echo "Usage:"
    echo "  curl -sSL https://monify.cloud/install.sh | sudo bash -s -- YOUR_TOKEN"
    echo ""
    echo "Get your token from: https://dash.monify.cloud"
    echo ""
}

# Check if running as root
check_root() {
    if [ "$(id -u)" != "0" ]; then
        print_error "This script must be run as root"
        echo "Please run: curl -sSL https://monify.cloud/install.sh | sudo bash -s -- YOUR_TOKEN"
        exit 1
    fi
}

# Detect system architecture
detect_arch() {
    local arch
    arch=$(uname -m)
    
    case $arch in
        x86_64|amd64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        *)
            print_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
}

# Check if systemd is available
check_systemd() {
    if ! command -v systemctl &> /dev/null; then
        print_error "systemd is not available on this system"
        print_info "Monify Agent requires systemd for service management"
        exit 1
    fi
}

# Check Linux distribution
check_os() {
    if [ "$(uname)" != "Linux" ]; then
        print_error "This script only supports Linux"
        exit 1
    fi
    
    # Print distribution info
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        print_info "Detected: $NAME $VERSION_ID"
    fi
}

# Download and install binary
install_binary() {
    local arch=$1
    local download_url="${DOWNLOAD_BASE}/${BINARY_NAME}-linux-${arch}"
    local temp_file="/tmp/${BINARY_NAME}"
    
    print_info "Downloading Monify Agent for linux/${arch}..."
    
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
    
    # Make executable and move to install directory
    chmod +x "$temp_file"
    mv "$temp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    
    print_success "Binary installed to ${INSTALL_DIR}/${BINARY_NAME}"
}

# Create configuration directory and save token
setup_config() {
    local token=$1
    
    print_info "Setting up configuration..."
    
    # Create config directory
    mkdir -p "$CONFIG_DIR"
    chmod 700 "$CONFIG_DIR"
    
    # Create env file with token
    cat > "${CONFIG_DIR}/env" << EOF
# Monify Agent Configuration
MONIFY_TOKEN=${token}
EOF
    chmod 600 "${CONFIG_DIR}/env"
    
    # Create log directory
    mkdir -p "$LOG_DIR"
    chmod 755 "$LOG_DIR"
    
    print_success "Token saved to ${CONFIG_DIR}/env"
}

# Create systemd service
create_service() {
    print_info "Creating systemd service..."
    
    cat > "$SERVICE_FILE" << 'EOF'
[Unit]
Description=Monify Monitoring Agent
Documentation=https://docs.monify.cloud
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/monify run
Restart=always
RestartSec=5
RestartPreventExitStatus=3
StandardOutput=journal
StandardError=journal
SyslogIdentifier=monify

# Security settings
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
ReadWritePaths=/etc/monify /var/log/monify
ProtectKernelTunables=yes
ProtectControlGroups=yes

# Resource limits
MemoryMax=64M
CPUQuota=5%

[Install]
WantedBy=multi-user.target
EOF

    chmod 644 "$SERVICE_FILE"
    
    # Reload systemd
    systemctl daemon-reload
    
    print_success "Systemd service created"
}

# Stop existing service if running
stop_existing() {
    if systemctl is-active --quiet monify 2>/dev/null; then
        print_info "Stopping existing Monify Agent..."
        systemctl stop monify
    fi
}

# Start and enable the service
start_service() {
    print_info "Starting Monify Agent..."
    
    systemctl enable monify >/dev/null 2>&1
    systemctl start monify
    
    # Wait for service to stabilize
    sleep 3
    
    if systemctl is-active --quiet monify; then
        print_success "Monify Agent is running!"
    else
        # Check if it was an auth failure
        local exit_status
        exit_status=$(systemctl show monify --property=ExecMainStatus | cut -d= -f2)
        
        if [ "$exit_status" = "3" ]; then
            print_error "Authentication failed - invalid token"
            echo ""
            echo "The token you provided is invalid or expired."
            echo "Please check your token at: https://dash.monify.cloud"
            echo ""
            echo "To fix:"
            echo "  1. Get a valid token from the dashboard"
            echo "  2. Run: sudo monify login YOUR_NEW_TOKEN"
            echo "  3. Run: sudo systemctl start monify"
            exit 1
        else
            print_error "Failed to start Monify Agent"
            echo "Check logs: journalctl -u monify --no-pager -n 20"
            exit 1
        fi
    fi
}

# Print success message
print_complete() {
    local version
    version=$("${INSTALL_DIR}/${BINARY_NAME}" version 2>/dev/null | head -1 || echo "unknown")
    
    echo ""
    echo "======================================"
    echo -e "${GREEN}Monify Agent installed and running!${NC}"
    echo "======================================"
    echo ""
    echo "Version: $version"
    echo "Status:  $(systemctl is-active monify)"
    echo ""
    echo "Useful commands:"
    echo "  View logs:     journalctl -u monify -f"
    echo "  Check status:  systemctl status monify"
    echo "  Restart:       sudo systemctl restart monify"
    echo "  Stop:          sudo systemctl stop monify"
    echo ""
    echo "Your server should appear at: https://dash.monify.cloud"
    echo ""
}

# Main installation flow
main() {
    local token="$1"
    local is_update=false
    local existing_token=""
    
    echo ""
    echo "======================================"
    echo "  Monify Agent Installer"
    echo "======================================"
    echo ""
    
    # Check if already installed
    if [ -f "${CONFIG_DIR}/env" ]; then
        existing_token=$(grep "^MONIFY_TOKEN=" "${CONFIG_DIR}/env" 2>/dev/null | cut -d= -f2)
    fi
    
    # No token provided?
    if [ -z "$token" ]; then
        if [ -n "$existing_token" ] && [ "$existing_token" != "" ]; then
            # Update mode: use existing token
            token="$existing_token"
            is_update=true
            print_info "Update mode: Using existing token"
        else
            # No token anywhere
            print_error "Token is required!"
            show_usage
            exit 1
        fi
    else
        # Token provided - check if different from existing
        if [ -n "$existing_token" ] && [ "$existing_token" != "" ] && [ "$existing_token" != "$token" ]; then
            print_warning "Agent is already installed with a different token!"
            echo ""
            echo "  Current token: ${existing_token:0:8}..."
            echo "  New token:     ${token:0:8}..."
            echo ""
            print_warning "This will REPLACE the existing token."
            print_warning "The old server will lose connection to this agent."
            echo ""
            
            # Ask for confirmation (only if interactive)
            if [ -t 0 ]; then
                read -p "Continue? [y/N] " -n 1 -r
                echo ""
                if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                    print_info "Installation cancelled."
                    exit 0
                fi
            else
                # Non-interactive: require --force flag
                if [ "$2" != "--force" ]; then
                    print_error "To replace existing token in non-interactive mode, use:"
                    echo "  curl -sSL https://monify.cloud/install.sh | sudo bash -s -- NEW_TOKEN --force"
                    exit 1
                fi
                print_info "Force mode: Replacing token"
            fi
        fi
    fi
    
    check_root
    check_os
    check_systemd
    
    local arch
    arch=$(detect_arch)
    print_info "Architecture: $arch"
    
    stop_existing
    install_binary "$arch"
    setup_config "$token"
    create_service
    start_service
    print_complete
}

# Run main with first argument as token
main "$1"
