#!/bin/bash
#
# Monify Agent Uninstallation Script
# Usage: curl -sSL https://monify.cloud/uninstall.sh | sudo bash
#
# This script:
# 1. Stops the agent service
# 2. Disables the systemd service
# 3. Removes binary, configuration, and logs
#

set -e

# Configuration
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/monify"
LOG_DIR="/var/log/monify"
SERVICE_FILE="/etc/systemd/system/monify.service"
BINARY_NAME="monify"

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

# Check if running as root
check_root() {
    if [ "$(id -u)" != "0" ]; then
        print_error "This script must be run as root"
        echo "Please run: curl -sSL https://monify.cloud/uninstall.sh | sudo bash"
        exit 1
    fi
}

# Stop and disable service
stop_service() {
    print_info "Stopping Monify Agent service..."
    
    if systemctl is-active --quiet monify 2>/dev/null; then
        systemctl stop monify
        print_success "Service stopped"
    else
        print_info "Service is not running"
    fi
    
    if systemctl is-enabled --quiet monify 2>/dev/null; then
        systemctl disable monify
        print_success "Service disabled"
    fi
}

# Remove systemd service file
remove_service() {
    print_info "Removing systemd service..."
    
    if [ -f "$SERVICE_FILE" ]; then
        rm -f "$SERVICE_FILE"
        systemctl daemon-reload
        print_success "Service file removed"
    else
        print_info "Service file not found"
    fi
}

# Remove binary
remove_binary() {
    print_info "Removing binary..."
    
    if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        rm -f "${INSTALL_DIR}/${BINARY_NAME}"
        print_success "Binary removed"
    else
        print_info "Binary not found"
    fi
}

# Remove configuration
remove_config() {
    print_info "Removing configuration..."
    
    if [ -d "$CONFIG_DIR" ]; then
        # Backup token if exists
        if [ -f "${CONFIG_DIR}/env" ]; then
            print_warning "Removing saved configuration (including token)"
        fi
        rm -rf "$CONFIG_DIR"
        print_success "Configuration removed"
    else
        print_info "Configuration directory not found"
    fi
}

# Remove logs
remove_logs() {
    print_info "Removing logs..."
    
    if [ -d "$LOG_DIR" ]; then
        rm -rf "$LOG_DIR"
        print_success "Logs removed"
    else
        print_info "Log directory not found"
    fi
}

# Clean up any remaining files
cleanup() {
    print_info "Cleaning up..."
    
    # Remove PID file if exists
    rm -f /var/run/monify.pid 2>/dev/null || true
    
    # Remove any lock files
    rm -f /var/run/monify.lock 2>/dev/null || true
    
    print_success "Cleanup complete"
}

# Print completion message
print_complete() {
    echo ""
    echo "======================================"
    echo -e "${GREEN}Monify Agent uninstalled successfully!${NC}"
    echo "======================================"
    echo ""
    echo "All files have been removed:"
    echo "  - Binary: ${INSTALL_DIR}/${BINARY_NAME}"
    echo "  - Config: ${CONFIG_DIR}"
    echo "  - Logs: ${LOG_DIR}"
    echo "  - Service: ${SERVICE_FILE}"
    echo ""
    echo "Thank you for using Monify!"
    echo "If you have any feedback, visit: https://monify.cloud/feedback"
    echo ""
}

# Confirm uninstallation
confirm_uninstall() {
    echo ""
    echo "======================================"
    echo "  Monify Agent Uninstaller"
    echo "======================================"
    echo ""
    print_warning "This will completely remove Monify Agent from your system."
    echo ""
    
    # Check if running interactively or via pipe
    if [ -t 0 ]; then
        read -p "Are you sure you want to continue? [y/N] " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Uninstallation cancelled"
            exit 0
        fi
    else
        # Running via pipe, proceed without confirmation
        print_info "Running in non-interactive mode, proceeding..."
    fi
}

# Main uninstallation flow
main() {
    check_root
    confirm_uninstall
    
    stop_service
    remove_service
    remove_binary
    remove_config
    remove_logs
    cleanup
    print_complete
}

# Run main
main "$@"
