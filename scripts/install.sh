#!/bin/bash
set -e

# Monify Agent Installation Script
# Usage: bash install.sh [TOKEN] [VERSION]
# Example: bash install.sh your_token_here latest

INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/monify"
SYSTEMD_DIR="/etc/systemd/system"
LOCK_DIR="/var/run"

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
    local download_url="https://github.com/${GITHUB_REPO}/releases/${version}/download/${BINARY_NAME}-linux-${ARCH}"
    
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
    log_success "Directories created"
}

# Create configuration
create_config() {
    local token="$1"
    
    if [ -f "${CONFIG_DIR}/config.yaml" ]; then
        log_warn "Configuration already exists, keeping existing config"
    else
        log_info "Creating default configuration..."
        cat > "${CONFIG_DIR}/config.yaml" <<'EOF'
server:
  url: "https://api.monify.cloud/v1/agent/metrics"
  timeout: 10s
  tls:
    enabled: true
    insecure_skip_verify: false

agent:
  hostname: ""
  version: "1.0.0"

collection:
  interval: 30s

metrics:
  cpu: true
  memory: true
  disk: true
  network: true
  system: true

port_scanner:
  enabled: true
  timeout: 5s
  max_workers: 100

logging:
  level: "info"
  format: "text"
  file: ""
EOF
        chmod 644 "${CONFIG_DIR}/config.yaml"
        log_success "Configuration created"
    fi
    
    # Create environment file with token
    if [ -n "$token" ]; then
        log_info "Configuring authentication token..."
        cat > "${CONFIG_DIR}/.env" <<EOF
# Monify Agent Environment Configuration
TOKEN_DEVICE=${token}
MONIFY_SERVER_URL=https://api.monify.cloud/v1/agent/metrics
MONIFY_COLLECTION_INTERVAL=30s
MONIFY_LOG_LEVEL=info
EOF
        chmod 600 "${CONFIG_DIR}/.env"
        log_success "Token configured"
    else
        log_warn "No token provided, you'll need to configure it manually"
        log_info "Edit ${CONFIG_DIR}/.env and add: TOKEN_DEVICE=your_token_here"
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
EnvironmentFile=-${CONFIG_DIR}/.env
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
ReadWritePaths=${CONFIG_DIR} ${LOCK_DIR}

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
        echo "  systemctl status ${SERVICE_NAME}"
        echo ""
        echo "View logs:"
        echo "  journalctl -u ${SERVICE_NAME} -f"
    else
        log_warn "Token not configured"
        echo ""
        echo "Configure your token:"
        echo "  1. Edit: sudo nano ${CONFIG_DIR}/.env"
        echo "  2. Add: TOKEN_DEVICE=your_token_here"
        echo "  3. Restart: sudo systemctl restart ${SERVICE_NAME}"
    fi
    
    echo ""
    echo "Useful commands:"
    echo "  Start:   systemctl start ${SERVICE_NAME}"
    echo "  Stop:    systemctl stop ${SERVICE_NAME}"
    echo "  Restart: systemctl restart ${SERVICE_NAME}"
    echo "  Status:  systemctl status ${SERVICE_NAME}"
    echo "  Logs:    journalctl -u ${SERVICE_NAME} -f"
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
