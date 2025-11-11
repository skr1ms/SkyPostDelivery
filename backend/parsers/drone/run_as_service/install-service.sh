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
INSTALL_DIR="/opt/drone-parser"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
LOG_DIR="/var/log/drone-parser"
USER="${SUDO_USER:-$(whoami)}"

echo "Installing for user: $USER"
echo ""

echo "Step 1: Installing system dependencies..."
apt update -qq
apt install -y \
    python3-venv \
    python3-pip \
    python3-dev \
    build-essential \
    libatlas-base-dev \
    libgl1
echo "System dependencies installed."
echo ""

echo "Step 2: Creating directories..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$LOG_DIR"

echo "Step 3: Copying application files..."
rsync -a --delete --exclude 'run_as_service' "$SOURCE_DIR/" "$INSTALL_DIR/"
rsync -a "$SCRIPT_DIR/" "$INSTALL_DIR/run_as_service/"
cd "$INSTALL_DIR"

echo "Step 4: Creating virtual environment..."
if [ ! -d "venv" ]; then
    python3 -m venv venv
    source venv/bin/activate
    pip install --upgrade pip
    pip install -r requirements.txt
    deactivate
else
    echo "Virtual environment already exists, skipping..."
fi

echo "Step 5: Creating .env file if not exists..."
if [ ! -f "$INSTALL_DIR/.env" ]; then
    cat > "$INSTALL_DIR/.env" << 'ENVEOF'
DRONE_IP=192.168.10.3
DRONE_SERVICE_HOST=localhost
DRONE_SERVICE_PORT=8001
PARCEL_AUTOMAT_IP=192.168.10.2

RECONNECT_INTERVAL=5
HEARTBEAT_INTERVAL=30

CAMERA_INDEX=0
VIDEO_FPS=5

LOG_LEVEL=INFO
USE_MOCK_HARDWARE=false

ARUCO_DICT_TYPE=DICT_6X6_250
MARKER_SIZE_CM=12.0

CRUISE_ALTITUDE=1.5
CRUISE_SPEED=0.3
LANDING_ALTITUDE=0.3
CENTER_THRESHOLD=0.05
DISTANCE_BETWEEN_MARKERS=2.0

FC_PORT=/dev/ttyAMA0
FC_BAUDRATE=115200
FC_TIMEOUT=1.0

MARKER_DETECTION_INTERVAL=0.2
TELEMETRY_INTERVAL=0.5

TARGET_MARKER_ID=52
ENVEOF
    echo "Created default .env file. Please edit it: nano $INSTALL_DIR/.env"
fi

echo "Step 6: Setting permissions..."
chown -R $USER:$USER "$INSTALL_DIR"
chown -R $USER:$USER "$LOG_DIR"
chmod +x "$INSTALL_DIR"/run_as_service/*.sh 2>/dev/null || true

echo "Step 7: Installing systemd service..."
cp "$INSTALL_DIR/run_as_service/drone.service" "$SERVICE_FILE"
sed -i "s/User=admin/User=$USER/g" "$SERVICE_FILE"
sed -i "s/Group=admin/Group=$USER/g" "$SERVICE_FILE"
systemctl daemon-reload

echo "Step 8: Enabling service..."
systemctl enable $SERVICE_NAME

echo ""
echo "=== Installation completed! ==="
echo ""
echo "Service commands:"
echo "  Start:   sudo systemctl start $SERVICE_NAME"
echo "  Stop:    sudo systemctl stop $SERVICE_NAME"
echo "  Restart: sudo systemctl restart $SERVICE_NAME"
echo "  Status:  sudo systemctl status $SERVICE_NAME"
echo "  Logs:    sudo journalctl -u $SERVICE_NAME -f"
echo ""
echo "Log files:"
echo "  Service: $LOG_DIR/service.log"
echo "  Errors:  $LOG_DIR/error.log"
echo ""
echo "Configuration:"
echo "  Edit:    sudo nano $INSTALL_DIR/.env"
echo ""
echo "To start the service now, run:"
echo "  sudo systemctl start $SERVICE_NAME"

