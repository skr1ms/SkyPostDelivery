#!/bin/bash

set -e

echo "=== Drone Parser Service Uninstaller ==="
echo ""

if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (sudo ./uninstall-service.sh)"
    exit 1
fi

SERVICE_NAME="drone-parser"
INSTALL_DIR="/opt/drone-parser"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
LOG_DIR="/var/log/drone-parser"

echo "Stopping service..."
systemctl stop $SERVICE_NAME 2>/dev/null || true

echo "Disabling service..."
systemctl disable $SERVICE_NAME 2>/dev/null || true

echo "Removing service file..."
rm -f "$SERVICE_FILE"

echo "Reloading systemd..."
systemctl daemon-reload

echo ""
read -p "Remove installation directory ($INSTALL_DIR)? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Removing installation directory..."
    rm -rf "$INSTALL_DIR"
fi

echo ""
read -p "Remove log directory ($LOG_DIR)? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Removing log directory..."
    rm -rf "$LOG_DIR"
fi

echo ""
echo "=== Uninstallation completed! ==="

