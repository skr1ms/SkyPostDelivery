#!/bin/bash
# Startup script for Clover drone parser with ROS
# Must be run with system Python3, not in venv

set -e

echo "========================================="
echo "Starting Clover Drone Parser"
echo "========================================="

# Source ROS environment
if [ -f "/opt/ros/noetic/setup.bash" ]; then
    echo "Sourcing ROS Noetic..."
    source /opt/ros/noetic/setup.bash
elif [ -f "/opt/ros/melodic/setup.bash" ]; then
    echo "Sourcing ROS Melodic..."
    source /opt/ros/melodic/setup.bash
else
    echo "ERROR: ROS environment not found!"
    echo "Please install ROS or adjust the path in this script."
    exit 1
fi

# Source catkin workspace if exists
if [ -f "$HOME/catkin_ws/devel/setup.bash" ]; then
    echo "Sourcing catkin workspace..."
    source $HOME/catkin_ws/devel/setup.bash
fi

# Check if we're in the correct directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo "Working directory: $(pwd)"

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "WARNING: .env file not found!"
    echo "Please create .env based on .env.example"
    exit 1
fi

# Load environment variables
echo "Loading environment variables from .env"
export $(grep -v '^#' .env | xargs)

# Check Python dependencies
echo "Checking Python dependencies..."
python3 -c "import rospy" 2>/dev/null || {
    echo "ERROR: rospy not found!"
    echo "Make sure ROS Python packages are installed:"
    echo "  sudo apt install ros-noetic-rospy ros-noetic-geometry-msgs ros-noetic-sensor-msgs"
    exit 1
}

python3 -c "import cv_bridge" 2>/dev/null || {
    echo "ERROR: cv_bridge not found!"
    echo "Install cv_bridge:"
    echo "  sudo apt install ros-noetic-cv-bridge"
    exit 1
}

# Check if other dependencies are installed
python3 -c "import websockets, cv2, numpy" 2>/dev/null || {
    echo "ERROR: Some Python dependencies missing!"
    echo "Install with: pip3 install websockets opencv-python numpy pydantic"
    exit 1
}

echo ""
echo "========================================="
echo "✅ All checks passed!"
echo "========================================="
echo "Drone ID: ${DRONE_IP:-not set}"
echo "Connecting to: ws://${DRONE_SERVICE_HOST:-localhost}:${DRONE_SERVICE_PORT:-8081}/ws/drone"
echo "========================================="
echo ""

# Run the drone parser
exec python3 main.py

