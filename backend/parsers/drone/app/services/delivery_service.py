import asyncio
import logging
from typing import Optional, Callable
from pathlib import Path

from ..models.schemas import (
    DeliveryTaskPayload,
    DroneStatus,
    Position,
    StatusUpdatePayload,
    DeliveryUpdatePayload
)
from ..navigation.clover_navigation_controller import CloverNavigationController
from config.config import settings

logger = logging.getLogger(__name__)


class DeliveryService:
    def __init__(self):
        self._status_callback: Optional[Callable] = None
        self._delivery_update_callback: Optional[Callable] = None
        self.nav_controller: Optional[CloverNavigationController] = None
        self.is_initialized = False
        self.current_delivery_id: Optional[str] = None
        self.current_task: Optional[DeliveryTaskPayload] = None
        self.cargo_drop_approved: bool = False
        self.target_cell_id: Optional[str] = None

    def set_status_callback(self, callback: Callable):
        self._status_callback = callback

    def set_delivery_update_callback(self, callback: Callable):
        self._delivery_update_callback = callback

    async def initialize(self):
        if self.is_initialized:
            return True

        logger.info("Initializing Clover navigation system")

        map_file = Path(settings.aruco_map_file)
        if not map_file.exists():
            map_file = Path(__file__).parent.parent.parent / \
                settings.aruco_map_file

        self.nav_controller = CloverNavigationController(str(map_file))
        success = self.nav_controller.initialize()

        if success:
            self.is_initialized = True
            logger.info("Clover navigation initialized")
        else:
            logger.error("Failed to initialize navigation")

        return success

    async def return_to_base(self, base_marker_id: int = 131):
        try:
            if not self.is_initialized:
                logger.info("Initializing for return to base")
                await self.initialize()

            if not self.is_initialized:
                logger.error("Cannot return to base: not initialized")
                return False

            logger.info(f"Returning to base (marker {base_marker_id})")

            if self._status_callback:
                await self._status_callback(StatusUpdatePayload(
                    status=DroneStatus.RETURNING,
                    battery_level=self.nav_controller.api.get_battery() if self.nav_controller else 100.0,
                    position=Position(
                        latitude=0.0, longitude=0.0, altitude=0.0),
                    speed=0.0,
                    current_delivery_id=None
                ))

            current_alt = self.nav_controller.api.get_altitude()
            if current_alt < 0.5:
                logger.info("Taking off for return to base")
                if not self.nav_controller.takeoff():
                    logger.error("Takeoff failed during return")
                    return False

            logger.info(f"Navigating to base marker {base_marker_id}")
            if not self.nav_controller.navigate_to_marker(base_marker_id):
                logger.error("Navigation to base failed")
                return False

            logger.info(f"Landing on base marker {base_marker_id}")
            if not self.nav_controller.land_on_marker(base_marker_id):
                logger.error("Landing on base failed")
                return False

            if self._status_callback:
                await self._status_callback(StatusUpdatePayload(
                    status=DroneStatus.IDLE,
                    battery_level=self.nav_controller.api.get_battery(),
                    position=Position(
                        latitude=0.0, longitude=0.0, altitude=0.0),
                    speed=0.0,
                    current_delivery_id=None
                ))

            logger.info("Successfully returned to base")
            self.current_delivery_id = None
            self.current_task = None
            return True

        except Exception as e:
            logger.error(f"Error during return to base: {e}")
            return False

    async def execute_delivery(self, task: DeliveryTaskPayload):
        try:
            if not self.is_initialized:
                logger.info("Initializing on first delivery")
                await self.initialize()

            if not self.is_initialized:
                logger.error("Cannot execute: not initialized")
                return

            self.current_delivery_id = task.delivery_id
            self.current_task = task
            logger.info(
                f"Delivery {task.delivery_id} (order {task.order_id}) to marker {task.aruco_id}")

            if self._status_callback:
                await self._status_callback(StatusUpdatePayload(
                    status=DroneStatus.TAKING_OFF,
                    battery_level=self.nav_controller.api.get_battery() if self.nav_controller else 100.0,
                    position=Position(
                        latitude=0.0, longitude=0.0, altitude=0.0),
                    speed=0.0,
                    current_delivery_id=task.delivery_id
                ))

            logger.info("Executing takeoff")
            if not self.nav_controller.takeoff():
                logger.error("Takeoff failed")
                if self._delivery_update_callback:
                    await self._delivery_update_callback(DeliveryUpdatePayload(
                        delivery_id=task.delivery_id,
                        drone_status="failed",
                        message="Takeoff failed"
                    ))
                return

            if self._status_callback:
                await self._status_callback(StatusUpdatePayload(
                    status=DroneStatus.IN_TRANSIT,
                    battery_level=self.nav_controller.api.get_battery(),
                    position=Position(latitude=0.0, longitude=0.0,
                                      altitude=self.nav_controller.cruise_altitude),
                    speed=self.nav_controller.cruise_speed,
                    current_delivery_id=task.delivery_id
                ))

            logger.info(f"Navigating to marker {task.aruco_id}")
            if not self.nav_controller.navigate_to_marker(task.aruco_id):
                logger.error("Navigation failed")
                if self._delivery_update_callback:
                    await self._delivery_update_callback(DeliveryUpdatePayload(
                        delivery_id=task.delivery_id,
                        drone_status="failed",
                        message="Navigation to marker failed"
                    ))
                return

            if self._status_callback:
                await self._status_callback(StatusUpdatePayload(
                    status=DroneStatus.DELIVERING,
                    battery_level=self.nav_controller.api.get_battery(),
                    position=Position(
                        latitude=0.0, longitude=0.0, altitude=self.nav_controller.landing_altitude),
                    speed=0.3,
                    current_delivery_id=task.delivery_id
                ))

            logger.info(f"Landing on marker {task.aruco_id}")
            if not self.nav_controller.land_on_marker(task.aruco_id):
                logger.error("Landing failed")
                if self._delivery_update_callback:
                    await self._delivery_update_callback(DeliveryUpdatePayload(
                        delivery_id=task.delivery_id,
                        drone_status="failed",
                        message="Landing failed"
                    ))
                return

            logger.info(
                f"Arrived at parcel automat {task.parcel_automat_id}, waiting for cell opening...")
            self.cargo_drop_approved = False
            self.target_cell_id = None

            if self._delivery_update_callback:
                await self._delivery_update_callback(DeliveryUpdatePayload(
                    delivery_id=task.delivery_id,
                    drone_status="arrived_at_destination",
                    message=f"Arrived at automat {task.parcel_automat_id}",
                    order_id=task.order_id,
                    parcel_automat_id=task.parcel_automat_id
                ))

            logger.info("Waiting for drop_cargo command from server...")
            timeout = 60
            wait_time = 0
            while not self.cargo_drop_approved and wait_time < timeout:
                await asyncio.sleep(1)
                wait_time += 1
                if wait_time % 5 == 0:
                    logger.info(
                        f"Still waiting for drop_cargo command... ({wait_time}s)")

            if not self.cargo_drop_approved:
                logger.error("Timeout waiting for drop_cargo command")
                if self._delivery_update_callback:
                    await self._delivery_update_callback(DeliveryUpdatePayload(
                        delivery_id=task.delivery_id,
                        drone_status="failed",
                        message="Timeout waiting for cell opening"
                    ))
                return

            logger.info(
                f"Drop cargo approved! Target cell: {self.target_cell_id}")

            logger.info("Dropping cargo...")
            await asyncio.sleep(3)

            if self._delivery_update_callback:
                await self._delivery_update_callback(DeliveryUpdatePayload(
                    delivery_id=task.delivery_id,
                    drone_status="cargo_dropped",
                    message=f"Cargo dropped into cell {self.target_cell_id}"
                ))

            self.cargo_drop_approved = False
            self.target_cell_id = None

            logger.info(f"Delivery {task.delivery_id} completed")
            if self._delivery_update_callback:
                await self._delivery_update_callback(DeliveryUpdatePayload(
                    delivery_id=task.delivery_id,
                    drone_status="completed",
                    message="Package delivered"
                ))

            if self._status_callback:
                await self._status_callback(StatusUpdatePayload(
                    status=DroneStatus.IDLE,
                    battery_level=self.nav_controller.api.get_battery(),
                    position=Position(
                        latitude=0.0, longitude=0.0, altitude=0.0),
                    speed=0.0,
                    current_delivery_id=None
                ))

        except Exception as e:
            logger.error(f"Error during delivery: {e}")
            if self._delivery_update_callback:
                await self._delivery_update_callback(DeliveryUpdatePayload(
                    delivery_id=task.delivery_id,
                    drone_status="failed",
                    message=f"Error: {str(e)}"
                ))

    async def shutdown(self):
        self.is_initialized = False
        logger.info("Delivery service shutdown")

    def get_current_status(self) -> StatusUpdatePayload:
        if self.nav_controller and self.nav_controller.api:
            try:
                pos = self.nav_controller.api.get_position('map')
                battery = self.nav_controller.api.get_battery()

                return StatusUpdatePayload(
                    status=DroneStatus.IDLE,
                    battery_level=battery,
                    position=Position(
                        latitude=pos[0] if pos else 0.0,
                        longitude=pos[1] if pos else 0.0,
                        altitude=pos[2] if pos else 0.0
                    ),
                    speed=0.0,
                    current_delivery_id=self.current_delivery_id
                )
            except:
                pass

        return StatusUpdatePayload(
            status=DroneStatus.IDLE,
            battery_level=100.0,
            position=Position(latitude=0.0, longitude=0.0, altitude=0.0),
            speed=0.0,
            current_delivery_id=None
        )
