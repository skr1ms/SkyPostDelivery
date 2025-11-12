#!/bin/bash

set -e

echo "=== Drone Parser Service Installer ==="
echo ""

if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (sudo ./install-service.sh)"
    exit 1
fi

SCRIPT_DIR=$(dirname "$(realpath "$0")")
SOURCE_DIR=$(realpath "$SCRIPT_DIR/..")

SERVICE_NAME="drone-parser"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
USER_NAME="${SUDO_USER:-$(whoami)}"

echo "Installing for user: $USER_NAME"
echo "Repository path: $SOURCE_DIR"
echo ""

echo "Step 1: Ensure start_drone.sh is executable..."
chmod +x "$SOURCE_DIR/start_drone.sh"

if [ ! -f "$SOURCE_DIR/.env" ]; then
    echo "WARNING: .env not found in $SOURCE_DIR"
    echo "Create one before starting the service (cp .env.example .env)"
fi

if [ ! -f "$SOURCE_DIR/start_drone.sh" ]; then
    echo "ERROR: start_drone.sh not found in $SOURCE_DIR"
    exit 1
fi

echo "Step 2: Preparing systemd service file..."
TMP_SERVICE=$(mktemp)
cp "$SCRIPT_DIR/drone.service" "$TMP_SERVICE"
sed -i "s#__USER__#$USER_NAME#g" "$TMP_SERVICE"
sed -i "s#__WORKING_DIR__#$SOURCE_DIR#g" "$TMP_SERVICE"

echo "Step 3: Installing systemd unit..."
cp "$TMP_SERVICE" "$SERVICE_FILE"
rm "$TMP_SERVICE"

systemctl daemon-reload
systemctl enable "$SERVICE_NAME"

echo ""
echo "=== Installation completed! ==="
echo "Service commands:"
echo "  sudo systemctl start $SERVICE_NAME"
echo "  sudo systemctl stop $SERVICE_NAME"
echo "  sudo systemctl restart $SERVICE_NAME"
echo "  sudo systemctl status $SERVICE_NAME"
echo ""
echo "Logs (journalctl):"
echo "  sudo journalctl -u $SERVICE_NAME -f"
echo ""
echo "Remember to configure ROS (.bashrc or start_drone.sh) and .env"
echo ""

