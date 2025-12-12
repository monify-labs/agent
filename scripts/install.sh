#!/bin/bash
set -e

# Monify Agent Installation Script
# Usage: bash install.sh [TOKEN] [VERSION]
# Example: bash install.sh your_token_here latest

INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/monify"
SYSTEMD_DIR="/etc/systemd/system"
LOCK_DIR="/var/run"
LOG_DIR="/var/log/monify"

BINARY_NAME="monify"
SERVICE_NAME="monify"
GITHUB_REPO="monify-labs/agent"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_root() {
    if [ "$(id -u)" != "0" ]; then
        log_error "This script must be run as root"
        echo "Please run: sudo $0 $*"
        exit 1
    fi
}

# Detect platform
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            log_error "Unsupported architecture: $ARCH"
            log_info "Supported: x86_64 (amd64), aarch64/arm64"
            exit 1
            ;;
    esac
    
    if [ "$OS" != "linux" ]; then
        log_error "This script only supports Linux (detected: $OS)"
        exit 1
    fi
    
    log_info "Platform: ${OS}/${ARCH}"
}

# Check dependencies
check_dependencies() {
    if ! command -v systemctl &> /dev/null; then
        log_error "systemd is required but not found"
        exit 1
    fi
    
    if ! command -v curl &> /dev/null && ! command -v wget &> /dev/null; then
        log_error "Either curl or wget is required"
        exit 1
    fi
}

# Download binary
download_binary() {
    local version="${1:-latest}"
    
    # Resolve latest version if requested
    if [ "$version" = "latest" ]; then
        log_info "Resolving latest version..."
        if command -v curl &> /dev/null; then
            # Follow redirect to get tag name
            local release_url=$(curl -sL -o /dev/null -w %{url_effective} "https://github.com/${GITHUB_REPO}/releases/latest")
            version=${release_url##*/}
        else
            # Fallback for wget
            local release_url=$(wget -S -O /dev/null "https://github.com/${GITHUB_REPO}/releases/latest" 2>&1 | grep "Location:" | tail -n 1 | awk '{print $2}')
            version=${release_url##*/}
        fi
        
        if [ -z "$version" ] || [ "$version" = "latest" ]; then
            log_warn "Could not resolve latest version tag, using 'latest' generic url"
            version="latest"
        else
            log_info "Resolved latest version to: $version"
        fi
    fi

    local download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${BINARY_NAME}-linux-${ARCH}"
    
    # If using 'latest' generic tag, use the special upload path
    if [ "$version" = "latest" ]; then
        download_url="https://github.com/${GITHUB_REPO}/releases/latest/download/${BINARY_NAME}-linux-${ARCH}"
    fi
    
    log_info "Downloading ${BINARY_NAME} (${version})..."
    
    if command -v curl &> /dev/null; then
        if ! curl -fsSL "$download_url" -o "/tmp/${BINARY_NAME}"; then
            log_error "Failed to download binary from: $download_url"
            exit 1
        fi
    else
        if ! wget -q "$download_url" -O "/tmp/${BINARY_NAME}"; then
            log_error "Failed to download binary from: $download_url"
            exit 1
        fi
    fi
    
    chmod +x "/tmp/${BINARY_NAME}"
    log_success "Binary downloaded"
}

# Stop existing service
stop_service() {
    if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
        log_info "Stopping existing service..."
        systemctl stop "$SERVICE_NAME"
        log_success "Service stopped"
    fi
}

# Install binary
install_binary() {
    log_info "Installing binary to ${INSTALL_DIR}..."
    install -m 755 "/tmp/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    log_success "Binary installed"
}

# Create directories
create_directories() {
    log_info "Creating directories..."
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$LOCK_DIR"
    mkdir -p "$LOG_DIR"
    log_success "Directories created"
}

# Create configuration
create_config() {
    local token="$1"
    
    # Save token if provided
    if [ -n "$token" ]; then
        log_info "Configuring authentication token..."
        echo "$token" > "${CONFIG_DIR}/token"
        chmod 600 "${CONFIG_DIR}/token"
        log_success "Token configured"
    else
        log_warn "No token provided, you'll need to configure it manually"
        log_info "Run: sudo ${BINARY_NAME} login"
    fi
}

# Install systemd service
install_systemd_service() {
    log_info "Installing systemd service..."
    
    cat > "${SYSTEMD_DIR}/${SERVICE_NAME}.service" <<EOF
[Unit]
Description=Monify Monitoring Agent
Documentation=https://github.com/${GITHUB_REPO}
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
Group=root
ExecStart=${INSTALL_DIR}/${BINARY_NAME} start
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=10s
StartLimitInterval=300s
StartLimitBurst=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=${SERVICE_NAME}

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${CONFIG_DIR} ${LOCK_DIR} ${LOG_DIR}

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096
TasksMax=4096

[Install]
WantedBy=multi-user.target
EOF
    
    chmod 644 "${SYSTEMD_DIR}/${SERVICE_NAME}.service"
    systemctl daemon-reload
    log_success "Systemd service installed"
}

# Enable and start service
enable_service() {
    log_info "Enabling service..."
    systemctl enable "$SERVICE_NAME"
    log_success "Service enabled (will start on boot)"
}

start_service() {
    log_info "Starting service..."
    systemctl start "$SERVICE_NAME"
    log_success "Service started"
}

# Print installation summary
print_summary() {
    local token="$1"
    
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  Monify Agent Installation Complete!${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    
    if [ -n "$token" ]; then
        log_success "Agent is installed and running"
        echo ""
        echo "Check status:"
        echo "  ${BINARY_NAME} status"
        echo ""
        echo "View logs:"
        echo "  journalctl -u ${SERVICE_NAME} -f"
    else
        log_warn "Token not configured"
        echo ""
        echo "Configure your token:"
        echo "  sudo ${BINARY_NAME} login"
        echo ""
        echo "Then check status:"
        echo "  ${BINARY_NAME} status"
    fi
    
    echo ""
    echo "Useful commands:"
    echo "  Status:  ${BINARY_NAME} status"
    echo "  Login:   sudo ${BINARY_NAME} login"
    echo "  Start:   sudo systemctl start ${SERVICE_NAME}"
    echo "  Stop:    sudo systemctl stop ${SERVICE_NAME}"
    echo "  Restart: sudo systemctl restart ${SERVICE_NAME}"
    echo "  Logs:    sudo journalctl -u ${SERVICE_NAME} -f"
    echo ""
    echo "Dashboard: https://dash.monify.cloud"
    echo "Documentation: https://github.com/${GITHUB_REPO}"
    echo ""
}

# Cleanup temporary files
cleanup() {
    rm -f "/tmp/${BINARY_NAME}"
}

# Main installation flow
main() {
    local token="$1"
    local version="${2:-latest}"
    
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  Monify Agent Installer${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    
    check_root "$@"
    detect_platform
    check_dependencies
    
    # Check if agent is already running
    if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
        echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${YELLOW}  ⚠️  Monify Agent is already running!${NC}"
        echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        log_warn "An active instance of Monify Agent was detected."
        echo ""
        echo "To reinstall, you must first uninstall the existing agent."
        echo "Run the following command to uninstall:"
        echo ""
        echo -e "  ${GREEN}curl -fsSL https://monify.cloud/uninstall.sh | sudo bash${NC}"
        echo ""
        echo "After uninstalling, you can run the install command again."
        echo ""
        exit 1
    fi

    download_binary "$version"
    stop_service
    install_binary
    create_directories
    create_config "$token"
    install_systemd_service
    enable_service
    start_service
    cleanup
    print_summary "$token"
}

# Run main function
main "$@"
