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

echo "Step 1: Ensure scripts are executable..."
chmod +x "$SOURCE_DIR/root_scripts/delivery_flight.py"
chmod +x "$SOURCE_DIR/root_scripts/flight_back.py"

if [ ! -f "$SOURCE_DIR/.env" ]; then
    echo "WARNING: .env not found in $SOURCE_DIR"
    echo "Create one before starting the service (cp .env.example .env)"
fi

echo "Step 2: Copy flight scripts to /root..."
if [ -d "/root" ]; then
    cp "$SOURCE_DIR/root_scripts/delivery_flight.py" /root/
    cp "$SOURCE_DIR/root_scripts/flight_back.py" /root/
    chmod +x /root/delivery_flight.py
    chmod +x /root/flight_back.py
    echo "Flight scripts copied to /root/"
else
    echo "WARNING: /root directory not accessible"
fi

echo "Step 3: Preparing systemd service file..."
TMP_SERVICE=$(mktemp)
cp "$SCRIPT_DIR/drone.service" "$TMP_SERVICE"
sed -i "s#__USER__#$USER_NAME#g" "$TMP_SERVICE"
sed -i "s#__WORKING_DIR__#$SOURCE_DIR#g" "$TMP_SERVICE"

echo "Step 4: Installing systemd unit..."
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
echo "Flight scripts installed:"
echo "  /root/delivery_flight.py"
echo "  /root/flight_back.py"
echo ""
echo "Remember to configure ROS environment and .env file"
echo ""
