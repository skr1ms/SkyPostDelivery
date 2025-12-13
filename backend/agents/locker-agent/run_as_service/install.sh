#!/bin/bash

set -e

SERVICE_NAME="locker-agent.service"
SERVICE_FILE="$(dirname "$0")/${SERVICE_NAME}"
INSTALL_DIR="/opt/locker-agent"
SYSTEMD_DIR="/etc/systemd/system"
BINARY_NAME="locker-agent"

echo "=========================================="
echo "  SkyPost Locker Agent Service Installer "
echo "=========================================="
echo ""

if [ "$EUID" -ne 0 ]; then 
    echo "Error: This script must be run as root"
    echo "Usage: sudo $0"
    exit 1
fi

if [ ! -f "$SERVICE_FILE" ]; then
    echo "Error: Service file not found: $SERVICE_FILE"
    exit 1
fi

SCRIPT_DIR="$(dirname "$0")/.."
BINARY_PATH="$SCRIPT_DIR/$BINARY_NAME"

if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: Binary not found: $BINARY_PATH"
    echo "Please build the binary first: make build"
    exit 1
fi

if [ ! -d "$INSTALL_DIR" ]; then
    echo "Creating installation directory: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR"
fi

echo "Copying locker-agent binary to $INSTALL_DIR"
cp "$BINARY_PATH" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

if [ -f "$SCRIPT_DIR/config/config.yaml" ]; then
    echo "Copying configuration files"
    cp -r "$SCRIPT_DIR/config" "$INSTALL_DIR/" 2>/dev/null || true
fi

if [ ! -f "$INSTALL_DIR/.env" ]; then
    if [ -f "$SCRIPT_DIR/.env.example" ]; then
        echo "Creating .env from .env.example"
        cp "$SCRIPT_DIR/.env.example" "$INSTALL_DIR/.env"
        echo "Please edit $INSTALL_DIR/.env with your configuration"
    fi
fi

if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "Stopping existing service"
    systemctl stop "$SERVICE_NAME"
fi

echo "Installing systemd service file"
cp "$SERVICE_FILE" "$SYSTEMD_DIR/$SERVICE_NAME"
chmod 644 "$SYSTEMD_DIR/$SERVICE_NAME"

echo "Reloading systemd daemon"
systemctl daemon-reload

echo "Enabling service to start on boot"
systemctl enable "$SERVICE_NAME"

echo ""
echo "=========================================="
echo "  Installation Complete"
echo "=========================================="
echo ""
echo "Service installed: $SERVICE_NAME"
echo "Installation directory: $INSTALL_DIR"
echo ""
echo "Next steps:"
echo "  1. Edit configuration: sudo nano $INSTALL_DIR/.env"
echo "  2. Start service: sudo systemctl start $SERVICE_NAME"
echo "  3. Check status: sudo systemctl status $SERVICE_NAME"
echo "  4. View logs: sudo journalctl -u $SERVICE_NAME -f"
echo ""
echo "Service will automatically start on system boot"
echo ""
