#!/bin/bash
set -e

# Monify Agent Uninstallation Script

INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/monify"
SYSTEMD_DIR="/etc/systemd/system"
LOG_DIR="/var/log/monify"

BINARY_NAME="monify"
SERVICE_NAME="monify"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if running as root
check_root() {
    if [ "$(id -u)" != "0" ]; then
        echo -e "${RED}✗ This script must be run as root${NC}"
        echo "Please run: sudo $0"
        exit 1
    fi
}

# Stop and disable service
stop_service() {
    if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
        echo -e "${BLUE}Stopping ${SERVICE_NAME} service...${NC}"
        systemctl stop "$SERVICE_NAME"
        echo -e "${GREEN}✓ Service stopped${NC}"
    fi
    
    if systemctl is-enabled --quiet "$SERVICE_NAME" 2>/dev/null; then
        echo -e "${BLUE}Disabling ${SERVICE_NAME} service...${NC}"
        systemctl disable "$SERVICE_NAME"
        echo -e "${GREEN}✓ Service disabled${NC}"
    fi
}

# Remove binary
remove_binary() {
    if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        echo -e "${BLUE}Removing binary...${NC}"
        rm -f "${INSTALL_DIR}/${BINARY_NAME}"
        echo -e "${GREEN}✓ Binary removed${NC}"
    fi
}

# Remove systemd service
remove_service() {
    if [ -f "${SYSTEMD_DIR}/${SERVICE_NAME}.service" ]; then
        echo -e "${BLUE}Removing systemd service...${NC}"
        rm -f "${SYSTEMD_DIR}/${SERVICE_NAME}.service"
        systemctl daemon-reload
        echo -e "${GREEN}✓ Service removed${NC}"
    fi
}

# Remove config and logs
remove_data() {
    echo ""
    echo -e "${YELLOW}Do you want to remove configuration and logs?${NC}"
    echo -e "  Config: ${CONFIG_DIR}"
    echo -e "  Logs:   ${LOG_DIR}"
    echo ""
    read -p "Remove data? [y/N]: " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        if [ -d "$CONFIG_DIR" ]; then
            rm -rf "$CONFIG_DIR"
            echo -e "${GREEN}✓ Removed config directory${NC}"
        fi
        
        if [ -d "$LOG_DIR" ]; then
            rm -rf "$LOG_DIR"
            echo -e "${GREEN}✓ Removed log directory${NC}"
        fi
    else
        echo -e "${BLUE}ℹ Kept configuration and logs${NC}"
    fi
}

# Main uninstallation
main() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  Monify Agent Uninstaller${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    check_root
    stop_service
    remove_service
    remove_binary
    remove_data
    
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  Uninstallation Complete!${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${BLUE}Thank you for using Monify Agent!${NC}"
    echo ""
}

# Run main
main "$@"
