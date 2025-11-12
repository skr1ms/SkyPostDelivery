import asyncio
import json
import logging
import websockets
from typing import Optional, Callable
from websockets.exceptions import ConnectionClosed

from ..models.schemas import (
    IncomingMessage,
    OutgoingMessage,
    MessageType,
    StatusUpdatePayload,
    DeliveryUpdatePayload,
    DroneStatus,
    Position
)
from config.config import settings

logger = logging.getLogger(__name__)


class WebSocketService:
    def __init__(self, on_delivery_task: Callable):
        self.websocket: Optional[websockets.WebSocketClientProtocol] = None
        self.on_delivery_task = on_delivery_task
        self.is_connected = False
        self.heartbeat_task: Optional[asyncio.Task] = None
        self.video_task: Optional[asyncio.Task] = None
        self.camera = None
        self.delivery_service = None
        self.ros_bridge = None

    async def connect(self):
        while True:
            try:
                logger.info(f"Connecting to {settings.websocket_url}")
                self.websocket = await websockets.connect(
                    settings.websocket_url,
                    ping_interval=20,
                    ping_timeout=10
                )
                self.is_connected = True
                logger.info("WebSocket connected, registering drone...")

                await self._register_drone()

                registration_response = await asyncio.wait_for(
                    self.websocket.recv(),
                    timeout=10.0
                )

                data = json.loads(registration_response)
                if data.get("type") == "registered":
                    settings.drone_id = data.get("drone_id")
                    logger.info(
                        f"Drone registered with ID: {settings.drone_id}")
                elif data.get("type") == "error":
                    logger.error(f"Registration failed: {data.get('message')}")
                    raise Exception("Registration failed")

                await self._send_initial_status()

                if self.ros_bridge:
                    logger.info(
                        "Using ROS camera stream, skipping local camera initialization")
                    self.camera = None
                else:
                try:
                    from ..hardware.camera import CameraController
                    self.camera = CameraController()
                    if self.camera.initialize():
                        logger.info(
                            "Camera initialized, starting video stream")
                    else:
                        logger.warning(
                            "Camera initialization failed, video disabled")
                        self.camera = None
                except Exception as e:
                    logger.error(f"Error initializing camera: {e}")
                    self.camera = None

                if self.heartbeat_task:
                    self.heartbeat_task.cancel()
                self.heartbeat_task = asyncio.create_task(
                    self._heartbeat_loop())

                if self.video_task:
                    self.video_task.cancel()
                if self.ros_bridge or self.camera:
                    self.video_task = asyncio.create_task(
                        self._video_stream_loop())

                await self._listen()

            except ConnectionClosed as e:
                logger.warning(f"Connection closed: {e}")
                self.is_connected = False
                settings.drone_id = None
                await self._reconnect()

            except Exception as e:
                logger.error(f"Connection error: {e}")
                self.is_connected = False
                settings.drone_id = None
                await self._reconnect()

    async def _listen(self):
        try:
            async for message in self.websocket:
                await self._handle_message(message)
        except ConnectionClosed:
            logger.warning("Connection closed while listening")
            raise

    async def _handle_message(self, message: str):
        try:
            data = json.loads(message)
            incoming = IncomingMessage(**data)

            logger.info(f"Received message type: {incoming.type}")

            if incoming.type == MessageType.DELIVERY_TASK:
                await self.on_delivery_task(incoming.payload)
            elif incoming.type == MessageType.COMMAND:
                await self._handle_command(incoming.payload)
            else:
                logger.debug(f"Unhandled message type: {incoming.type}")

        except Exception as e:
            logger.error(f"Error handling message: {e}")

    async def _handle_command(self, payload: dict):
        command = payload.get("command")
        logger.info(f"Received command: {command}")

        if command == "drop_cargo":
            order_id = payload.get("order_id")
            cell_id = payload.get("cell_id")
            logger.info(
                f"Drop cargo command for order {order_id}, cell {cell_id}")

            if self.delivery_service:
                self.delivery_service.cargo_drop_approved = True
                self.delivery_service.target_cell_id = cell_id

        elif command == "return_to_base":
            base_marker_id = payload.get("base_marker_id", 131)
            logger.info(
                f"Return to base command received, target marker: {base_marker_id}")

            if self.delivery_service:
                asyncio.create_task(
                    self.delivery_service.return_to_base(base_marker_id))
            else:
                logger.error(
                    "Cannot return to base: delivery_service not available")

        else:
            logger.warning(f"Unknown command: {command}")

    async def send_status_update(self, payload: StatusUpdatePayload):
        if not self.is_connected or not self.websocket:
            logger.warning("Cannot send status update: not connected")
            return

        try:
            message = OutgoingMessage(
                type=MessageType.STATUS_UPDATE,
                payload=payload.model_dump()
            )
            await self.websocket.send(message.model_dump_json())
            logger.debug(f"Status update sent: {payload.status}")

        except Exception as e:
            logger.error(f"Error sending status update: {e}")

    async def send_delivery_update(self, payload: DeliveryUpdatePayload):
        if not self.is_connected or not self.websocket:
            logger.warning("Cannot send delivery update: not connected")
            return

        try:
            message = OutgoingMessage(
                type=MessageType.DELIVERY_UPDATE,
                payload=payload.model_dump()
            )
            await self.websocket.send(message.model_dump_json())
            logger.info(
                f"Delivery update sent: {payload.drone_status} for {payload.delivery_id}")

        except Exception as e:
            logger.error(f"Error sending delivery update: {e}")

    async def _send_initial_status(self):
        initial_status = StatusUpdatePayload(
            status=DroneStatus.IDLE,
            battery_level=100.0,
            position=Position(latitude=0.0, longitude=0.0, altitude=0.0),
            speed=0.0
        )
        await self.send_status_update(initial_status)

    async def _register_drone(self):
        registration = {
            "type": "register",
            "ip_address": settings.drone_ip
        }
        await self.websocket.send(json.dumps(registration))
        logger.info(f"Registration sent with IP: {settings.drone_ip}")

    async def _heartbeat_loop(self):
        while self.is_connected:
            try:
                await asyncio.sleep(settings.heartbeat_interval)

                if self.websocket and self.is_connected and settings.drone_id:
                    battery_level = 85.5
                    latitude = 0.0
                    longitude = 0.0
                    altitude = 0.0
                    status = "idle"

                    if self.ros_bridge:
                        try:
                            telemetry = self.ros_bridge.get_telemetry()
                            if telemetry.get('battery'):
                                battery_level = telemetry['battery'].get(
                                    'voltage', 85.5)
                            if telemetry.get('pose'):
                                pose = telemetry['pose']
                                latitude = pose.get('x', 0.0)
                                longitude = pose.get('y', 0.0)
                                altitude = pose.get('z', 0.0)
                            if telemetry.get('state'):
                                state = telemetry['state']
                                if state.get('armed'):
                                    status = "flying"
                                elif state.get('connected'):
                                    status = "idle"
                        except Exception as e:
                            logger.warning(
                                f"Failed to get telemetry from ROS: {e}")
                    elif self.delivery_service and self.delivery_service.nav_controller and self.delivery_service.nav_controller.api:
                        try:
                            battery_level = self.delivery_service.nav_controller.api.get_battery()
                            pos = self.delivery_service.nav_controller.api.get_position(
                                'map')
                            if pos:
                                latitude = pos[0]
                                longitude = pos[1]
                                altitude = pos[2]
                            status = self.delivery_service.nav_controller.state.value
                        except Exception as e:
                            logger.warning(
                                f"Failed to get telemetry for heartbeat: {e}")

                    heartbeat = {
                        "type": "heartbeat",
                        "battery_level": battery_level,
                        "position": {
                            "latitude": latitude,
                            "longitude": longitude,
                            "altitude": altitude
                        },
                        "status": status,
                        "speed": 0.0
                    }
                    await self.websocket.send(json.dumps(heartbeat))
                    logger.debug("Heartbeat sent")

            except Exception as e:
                logger.error(f"Error in heartbeat loop: {e}")
                break

    async def _video_stream_loop(self):
        frame_interval = 1.0 / settings.video_fps if settings.video_fps > 0 else 0.2
        frame_counter = 0

        logger.info(
            f"Starting video stream loop with {settings.video_fps} FPS (interval: {frame_interval}s)")

        while self.is_connected and (self.ros_bridge or self.camera):
            try:
                await asyncio.sleep(frame_interval)

                if not self.websocket or not self.is_connected or not settings.drone_id:
                    logger.warning(
                        "Video loop: Not ready to send (ws/connected/drone_id missing)")
                    continue

                frame_base64 = None

                if self.ros_bridge:
                    frame_base64 = self.ros_bridge.get_frame()
                elif self.camera:
                frame_base64 = self.camera.capture_frame()

                if frame_base64:
                    frame_counter += 1
                    video_message = {
                        "type": "video_frame",
                        "payload": {
                            "frame": frame_base64,
                            "delivery_id": None
                        }
                    }
                    await self.websocket.send(json.dumps(video_message))

                    if frame_counter % 10 == 0:
                        logger.info(
                            f"Video frame #{frame_counter} sent, size: {len(frame_base64)} bytes")
                    else:
                        logger.debug(f"Video frame #{frame_counter} sent")
                else:
                    logger.warning("No frame available from ROS or camera!")

            except Exception as e:
                logger.error(f"Error in video stream loop: {e}")
                break

    async def _reconnect(self):
        logger.info(
            f"Reconnecting in {settings.reconnect_interval} seconds...")
        await asyncio.sleep(settings.reconnect_interval)

    async def close(self):
        self.is_connected = False

        if self.heartbeat_task:
            self.heartbeat_task.cancel()

        if self.video_task:
            self.video_task.cancel()

        if self.camera:
            self.camera.close()

        if self.websocket:
            await self.websocket.close()
            logger.info("WebSocket connection closed")
