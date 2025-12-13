#!/bin/bash

set -e

SERVICE_NAME="locker-agent.service"
INSTALL_DIR="/opt/locker-agent"
SYSTEMD_DIR="/etc/systemd/system"

echo "=========================================="
echo " SkyPost Locker Agent Service Uninstaller"
echo "=========================================="
echo ""

if [ "$EUID" -ne 0 ]; then 
    echo "Error: This script must be run as root"
    echo "Usage: sudo $0"
    exit 1
fi

if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "Stopping service"
    systemctl stop "$SERVICE_NAME"
fi

if systemctl is-enabled --quiet "$SERVICE_NAME" 2>/dev/null; then
    echo "Disabling service"
    systemctl disable "$SERVICE_NAME"
fi

if [ -f "$SYSTEMD_DIR/$SERVICE_NAME" ]; then
    echo "Removing service file"
    rm "$SYSTEMD_DIR/$SERVICE_NAME"
fi

echo "Reloading systemd daemon"
systemctl daemon-reload
systemctl reset-failed

echo ""
echo "=========================================="
echo "  Uninstallation Complete"
echo "=========================================="
echo ""
echo "Service removed: $SERVICE_NAME"
echo ""
echo "Note: Installation directory not removed: $INSTALL_DIR"
echo "To remove it manually: sudo rm -rf $INSTALL_DIR"
echo ""
