import asyncio
import logging
from ..core.websocket_client import WebSocketClient
from ..core.ros_bridge import ROSBridge

logger = logging.getLogger(__name__)


class TelemetryService:
    def __init__(
        self,
        websocket_client: WebSocketClient,
        ros_bridge: ROSBridge,
        interval: int = 30,
    ):
        self.ws_client = websocket_client
        self.ros_bridge = ros_bridge
        self.interval = interval
        self._task = None
        self._running = False

    async def start(self):
        if self._running:
            logger.warning("Telemetry service already running")
            return

        self._running = True
        self._task = asyncio.create_task(self._heartbeat_loop())
        logger.info(f"Telemetry service started (interval: {self.interval}s)")

    async def _heartbeat_loop(self):
        while self._running:
            try:
                await asyncio.sleep(self.interval)

                if (
                    not self.ws_client.is_connected
                    or not self.ws_client.settings.drone_id
                ):
                    continue

                telemetry = self.ros_bridge.get_telemetry()

                battery_level = 100.0
                if telemetry.battery:
                    battery_level = telemetry.battery.voltage

                position = {"latitude": 0.0, "longitude": 0.0, "altitude": 0.0}
                if telemetry.pose:
                    position = {
                        "latitude": telemetry.pose.x,
                        "longitude": telemetry.pose.y,
                        "altitude": telemetry.pose.z,
                    }

                status = "idle"
                if telemetry.state:
                    status = "flying" if telemetry.state.armed else "idle"

                await self.ws_client.send_heartbeat(
                    battery_level=battery_level,
                    position=position,
                    status=status,
                    speed=0.0,
                )

                logger.debug("Heartbeat sent")

            except Exception as e:
                logger.error(f"Error in heartbeat loop: {e}")

    async def stop(self):
        self._running = False
        if self._task:
            self._task.cancel()
            try:
                await self._task
            except asyncio.CancelledError:
                pass
        logger.info("Telemetry service stopped")
