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
        self.ros_bridge = None
        self.flight_manager = None
        self.current_delivery_task: Optional[dict] = None

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
                logger.info("Received delivery task from backend")
                await self._handle_delivery_task(incoming.payload)
            elif incoming.type == MessageType.COMMAND:
                await self._handle_command(incoming.payload)
            else:
                logger.debug(f"Unhandled message type: {incoming.type}")

        except Exception as e:
            logger.error(f"Error handling message: {e}")

    async def _handle_delivery_task(self, payload: dict):
        try:
            order_id = payload.get("order_id")
            delivery_id = payload.get("delivery_id")
            parcel_automat_id = payload.get("parcel_automat_id")
            target_marker_id = payload.get("target_aruco_id", 135)
            home_marker_id = payload.get("home_aruco_id", 131)
            
            # Сохраняем текущую задачу доставки для использования при прибытии
            self.current_delivery_task = {
                "order_id": order_id,
                "delivery_id": delivery_id,
                "parcel_automat_id": parcel_automat_id
            }
            
            logger.info("="*60)
            logger.info("DELIVERY TASK RECEIVED")
            logger.info(f"  Order ID: {order_id}")
            logger.info(f"  Delivery ID: {delivery_id}")
            logger.info(f"  Parcel Automat ID: {parcel_automat_id}")
            logger.info(f"  Target ArUco: {target_marker_id}")
            logger.info(f"  Home ArUco: {home_marker_id}")
            logger.info("="*60)
            
            if not self.flight_manager:
                logger.error("Flight manager not available!")
                return
            
            # Launch delivery flight script
            success = await self.flight_manager.launch_delivery_flight(
                aruco_id=target_marker_id,
                home_aruco_id=home_marker_id
            )
            
            if success:
                logger.info("Delivery flight script launched successfully")
                logger.info("Drone will:")
                logger.info("  1. Take off to cruise altitude")
                logger.info(f"  2. Navigate to ArUco marker {target_marker_id}")
                logger.info("  3. Land on parcel automat")
                logger.info("  4. Publish arrival notification")
                logger.info("  5. Wait for drop confirmation")
                logger.info("  6. Drop cargo")
                logger.info(f"  7. Return to base (ArUco {home_marker_id}) after 10 seconds")
            else:
                logger.error("Failed to launch delivery flight script")
                
        except Exception as e:
            logger.error(f"Error handling delivery task: {e}")
            import traceback
            traceback.print_exc()

    async def _handle_command(self, payload: dict):
        command = payload.get("command")
        logger.info(f"Received command: {command}")

        if command == "drop_cargo":
            order_id = payload.get("order_id")
            cell_id = payload.get("cell_id")
            internal_cell_id = payload.get("internal_cell_id")
            logger.info("="*60)
            logger.info("DROP CARGO COMMAND RECEIVED")
            logger.info(f"  Order ID: {order_id}")
            logger.info(f"  Cell ID: {cell_id}")
            logger.info(f"  Internal Cell ID: {internal_cell_id}")
            logger.info("="*60)

            if self.ros_bridge:
                logger.info("Sending drop confirmation via ROS topic /drone/delivery/drop_confirm")
                success = await self.ros_bridge.send_drop_confirmation()
                if success:
                    logger.info("Drop confirmation sent to drone")
                    logger.info("Drone will now drop cargo and return to base")
                else:
                    logger.error("Failed to send drop confirmation")
            else:
                logger.error("ROS bridge not available, cannot send drop confirmation")

        elif command == "return_to_base":
            base_marker_id = payload.get("base_marker_id", 131)
            logger.info("="*60)
            logger.info("RETURN TO BASE COMMAND")
            logger.info(f"  Target: ArUco {base_marker_id}")
            logger.info("="*60)

            if self.flight_manager:
                success = await self.flight_manager.launch_return_flight(
                    home_aruco_id=base_marker_id
                )
                if success:
                    logger.info("Return flight script launched")
                else:
                    logger.error("Failed to launch return flight")
            else:
                logger.error("Flight manager not available")

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
