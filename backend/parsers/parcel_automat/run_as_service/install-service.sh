#!/bin/bash

set -e

echo "=== Parcel Automat Service Installer ==="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "Please run as root (sudo ./install-service.sh)"
    exit 1
fi

# Paths
SCRIPT_DIR=$(dirname "$(realpath "$0")")
SOURCE_DIR=$(realpath "$SCRIPT_DIR/..")

# Configuration
SERVICE_NAME="parcel-automat"
INSTALL_DIR="/opt/parcel-automat"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
LOG_DIR="/var/log/parcel-automat"
USER="${SUDO_USER:-$(whoami)}"

echo "Installing for user: $USER"
echo ""

echo "Step 1: Installing system dependencies..."
apt update -qq
apt install -y \
    libzbar0 \
    python3-venv \
    python3-pip \
    python3-dev \
    build-essential
echo "System dependencies installed."
echo ""

echo "Step 2: Creating directories..."
mkdir -p $INSTALL_DIR
mkdir -p $LOG_DIR
mkdir -p $INSTALL_DIR/data

echo "Step 3: Copying application files..."
cp -r "$SOURCE_DIR"/. "$INSTALL_DIR"/
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
    cat > "$INSTALL_DIR/.env" << EOF
SERVICE_HOST=0.0.0.0
SERVICE_PORT=8000

GO_ORCHESTRATOR_URL=http://localhost:8080/api/v1

CELLS_MAPPING_FILE=/opt/parcel-automat/data/cells_mapping.json

CAMERA_INDEX=0
QR_SCAN_INTERVAL=0.1
USE_MOCK_SCANNER=false
SCANNER_STABLE_FRAMES=3
SCANNER_MISS_FRAMES=5
SCANNER_DEBOUNCE_SECONDS=5

ARDUINO_TIMEOUT=1.0
ARDUINO_PORT=/dev/ttyUSB0
ARDUINO_BAUDRATE=9600
USE_MOCK_ARDUINO=false

DISPLAY_PORT=/dev/ttyUSB1
DISPLAY_BAUDRATE=115200

LOG_LEVEL=INFO
EOF
    echo "Created default .env file. Please edit it: nano $INSTALL_DIR/.env"
fi

echo "Step 6: Setting permissions..."
chown -R $USER:$USER $INSTALL_DIR
chown -R $USER:$USER $LOG_DIR
chmod +x $INSTALL_DIR/*.sh 2>/dev/null || true

echo "Step 7: Installing systemd service..."
cp "$SCRIPT_DIR/parcel-automat.service" $SERVICE_FILE
sed -i "s/User=admin/User=$USER/g" $SERVICE_FILE
sed -i "s/Group=admin/Group=$USER/g" $SERVICE_FILE
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

