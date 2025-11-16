#!/bin/bash

set -e

echo "=== Drone Parser Service Uninstaller ==="
echo ""

if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (sudo ./uninstall-service.sh)"
    exit 1
fi

SERVICE_NAME="drone-parser"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

echo "Stopping service..."
systemctl stop "$SERVICE_NAME" 2>/dev/null || true

echo "Disabling service..."
systemctl disable "$SERVICE_NAME" 2>/dev/null || true

echo "Removing service file..."
rm -f "$SERVICE_FILE"

echo "Reloading systemd..."
systemctl daemon-reload

echo ""
echo "=== Uninstallation completed! ==="
echo ""

