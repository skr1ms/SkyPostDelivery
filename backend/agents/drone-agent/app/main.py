#!/usr/bin/env python3

import asyncio
import signal
import sys
import threading
import rospy
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from app.config.settings import settings
from app.utils.logger import setup_logger
from app.core.websocket_client import WebSocketClient
from app.core.ros_bridge import ROSBridge
from app.core.state_machine import StateMachine
from app.hardware.camera_handler import CameraHandler
from app.hardware.flight_controller import FlightController
from app.services.delivery_service import DeliveryService
from app.services.telemetry_service import TelemetryService
from app.services.video_service import VideoService

logger = setup_logger(__name__, settings.log_level)


class DroneApplication:
    def __init__(self):
        self.is_running = False
        self.ros_thread = None
        self.event_loop = None

        self.ws_client = None
        self.ros_bridge = None
        self.state_machine = None
        self.camera_handler = None
        self.flight_controller = None
        self.delivery_service = None
        self.telemetry_service = None
        self.video_service = None

    def _init_ros(self):
        try:
            logger.info("Initializing ROS node")
            rospy.init_node("drone_agent", anonymous=False, disable_signals=True)

            self.ros_bridge = ROSBridge(self.event_loop)
            self.ros_bridge.start()

            self.camera_handler = CameraHandler()
            self.camera_handler.start()

            logger.info("ROS node initialized successfully")
            rospy.spin()

        except Exception as e:
            logger.error(f"ROS initialization failed: {e}")
            self.is_running = False

    async def start(self):
        logger.info("=" * 60)
        logger.info(f"Starting Drone Agent - IP: {settings.drone_ip}")
        logger.info(f"Connecting to: {settings.websocket_url}")
        logger.info("=" * 60)

        self.is_running = True
        self.event_loop = asyncio.get_event_loop()

        self.ros_thread = threading.Thread(target=self._init_ros, daemon=True)
        self.ros_thread.start()

        await asyncio.sleep(2)

        try:
            self.state_machine = StateMachine()
            self.flight_controller = FlightController(settings.scripts_dir)
            self.ws_client = WebSocketClient(settings)

            self.delivery_service = DeliveryService(
                self.ws_client,
                self.state_machine,
                self.flight_controller,
                self.ros_bridge,
            )

            self.telemetry_service = TelemetryService(
                self.ws_client, self.ros_bridge, settings.heartbeat_interval
            )

            self.video_service = VideoService(
                self.ws_client, self.camera_handler, settings.video_fps
            )

            self.ws_client.on_delivery_task = self.delivery_service.handle_delivery_task

            async def handle_command(payload: dict):
                command = payload.get("command")
                if command == "drop_cargo":
                    await self.delivery_service.handle_drop_command(payload)
                else:
                    logger.warning(f"Unknown command: {command}")

            self.ws_client.on_command = handle_command

            self.ros_bridge.set_arrival_callback(self.delivery_service.on_arrival)
            self.ros_bridge.set_drop_ready_callback(self.delivery_service.on_drop_ready)
            self.ros_bridge.set_home_callback(self.delivery_service.on_home_arrived)

            await self.telemetry_service.start()
            await self.video_service.start()

            await self.ws_client.connect()

        except asyncio.CancelledError:
            logger.info("Application cancelled")
        finally:
            await self.stop()

    async def stop(self):
        logger.info("Stopping Drone Agent")
        self.is_running = False

        if self.video_service:
            await self.video_service.stop()

        if self.telemetry_service:
            await self.telemetry_service.stop()

        if self.ws_client:
            await self.ws_client.close()

        if self.camera_handler:
            self.camera_handler.stop()

        if self.ros_bridge:
            self.ros_bridge.stop()

        if not rospy.is_shutdown():
            rospy.signal_shutdown("Application stopping")

        logger.info("Drone Agent stopped")


async def main():
    app = DroneApplication()

    loop = asyncio.get_event_loop()

    def signal_handler():
        logger.info("Received shutdown signal")
        asyncio.create_task(app.stop())

    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, signal_handler)

    try:
        await app.start()
    except KeyboardInterrupt:
        logger.info("Keyboard interrupt received")
    finally:
        await app.stop()


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("Application terminated")
