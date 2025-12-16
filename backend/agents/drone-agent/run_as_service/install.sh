#!/bin/bash

set -e

SERVICE_NAME="drone-agent.service"
SERVICE_FILE="$(dirname "$0")/${SERVICE_NAME}"
INSTALL_DIR="/home/pi/drone-agent"
SYSTEMD_DIR="/etc/systemd/system"

echo "=========================================="
echo "  SkyPost Drone Agent Service Installer  "
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

if [ ! -f "/opt/ros/noetic/setup.bash" ]; then
    echo "Warning: ROS Noetic not found at /opt/ros/noetic/"
    echo "Make sure ROS is installed before running the service"
fi

if [ ! -d "$INSTALL_DIR" ]; then
    echo "Creating installation directory: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR"
    chown pi:pi "$INSTALL_DIR"
fi

echo "Copying drone-agent files to $INSTALL_DIR"
SCRIPT_DIR="$(dirname "$0")/.."
cp -r "$SCRIPT_DIR"/{app,scripts,config,.env.example} "$INSTALL_DIR/" 2>/dev/null || true
chown -R pi:pi "$INSTALL_DIR"

if [ ! -f "$INSTALL_DIR/.env" ]; then
    if [ -f "$INSTALL_DIR/.env.example" ]; then
        echo "Creating .env from .env.example"
        cp "$INSTALL_DIR/.env.example" "$INSTALL_DIR/.env"
        chown pi:pi "$INSTALL_DIR/.env"
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
