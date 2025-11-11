import asyncio
import signal
import logging

from config.config import settings
from app.dependencies import websocket_service, cleanup

logging.basicConfig(
    level=settings.log_level,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class DroneApplication:
    def __init__(self):
        self.is_running = False

    async def start(self):
        logger.info(f"Starting Drone Application - ID: {settings.drone_id}")
        logger.info(f"Drone IP: {settings.drone_ip}")
        logger.info(f"Parcel Automat IP: {settings.parcel_automat_ip}")
        logger.info(f"Connecting to: {settings.websocket_url}")

        self.is_running = True

        try:
            await websocket_service.connect()
        except asyncio.CancelledError:
            logger.info("Application cancelled")
        finally:
            await self.stop()

    async def stop(self):
        logger.info("Stopping Drone Application")
        self.is_running = False
        await cleanup()
        logger.info("Drone Application stopped")


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
