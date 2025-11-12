# ФАЙЛЫ ИЗ ЭТОЙ ДИРЕКТОРИИ ДОЛЖНЫ НАХОДИТСЯ В /home/user_name на вашей raspberry pi # Drone Flight Scripts

## Overview

This directory contains autonomous flight scripts that run directly on the Raspberry Pi 4 onboard the Clover drone. These scripts use ROS topics and Clover API for navigation.

## Scripts

### delivery_flight.py

Autonomous delivery flight script that:
- Takes off to cruise altitude (1.5m)
- Navigates to target ArUco marker coordinates
- Descends and lands on the marker
- Publishes arrival notification to `/drone/delivery/arrived` topic
- Waits for drop confirmation (60s timeout)
- Publishes drop ready notification to `/drone/delivery/drop_ready`
- Waits 10 seconds for cargo drop mechanism
- Automatically launches return flight script

**Usage:**
```bash
python3 /root/delivery_flight.py <aruco_id> <x_coord> <y_coord>
```

**ROS Topics Published:**
- `/drone/delivery/arrived` - Arrival notification
- `/drone/delivery/drop_ready` - Ready to drop cargo

**ROS Topics Subscribed:**
- `/mavros/battery` - Battery status
- `/mavros/local_position/pose` - Current position
- `/mavros/state` - Flight controller state

### flight_back.py

Return to base flight script that:
- Takes off if needed
- Navigates back to home ArUco marker
- Lands on home marker
- Publishes home arrival notification

**Usage:**
```bash
python3 /root/flight_back.py [home_aruco_id] [home_x] [home_y]
```

Default: ArUco 131 at (0, 0)

**ROS Topics Published:**
- `/drone/delivery/home_arrived` - Home arrival notification

## Installation

Scripts are automatically copied to `/root/` during service installation:

```bash
cd /home/takuya/hiTech/backend/parsers/drone/run_as_service
sudo ./install-service.sh
```

## ROS Integration

### Required Services

- `navigate` - Navigate to coordinates
- `get_telemetry` - Get current telemetry
- `land` - Landing command

### Camera Stream

Video stream available at:
```
http://localhost:8080/stream?topic=/main_camera/image_raw
```

### Battery Monitoring

Battery status from `/mavros/battery`:
- `voltage` - Current voltage
- `percentage` - Battery percentage (0-1)
- `current` - Current draw

## Flight Parameters

Configured in main application:
- Cruise altitude: 1.5m
- Landing altitude: 0.5m
- Navigation speed: 0.8 m/s
- Landing speed: 0.3 m/s
- Position tolerance: 0.2m

## Execution Flow

1. WebSocket service receives delivery task
2. FlightManager launches `delivery_flight.py` with parameters
3. Script navigates autonomously using ROS
4. On arrival, publishes to ROS topic
5. Waits for drop confirmation
6. After cargo drop, launches `flight_back.py`
7. Return flight completes autonomously

## Safety

- 60 second timeout for drop confirmation
- Automatic return to home after delivery
- Battery monitoring via ROS topics
- Position validation before landing

## Troubleshooting

Check ROS topics:
```bash
rostopic list
rostopic echo /drone/delivery/arrived
rostopic echo /mavros/battery
```

Check script logs:
```bash
tail -f /var/log/syslog | grep delivery_flight
```

Monitor service:
```bash
sudo journalctl -u drone-parser -f
```

## ВАЖНО

ФАЙЛЫ delivery_flight.py и flight_back.py ДОЛЖНЫ НАХОДИТЬСЯ В /root/ НА RASPBERRY PI 4
