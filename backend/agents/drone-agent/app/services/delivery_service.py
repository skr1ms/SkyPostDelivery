import logging
from ..core.websocket_client import WebSocketClient
from ..core.state_machine import StateMachine
from ..hardware.flight_controller import FlightController
from ..models.task import DeliveryTask, DeliveryState

logger = logging.getLogger(__name__)


class DeliveryService:
    def __init__(
        self,
        websocket_client: WebSocketClient,
        state_machine: StateMachine,
        flight_controller: FlightController,
        ros_bridge,
    ):
        self.ws_client = websocket_client
        self.state_machine = state_machine
        self.flight_controller = flight_controller
        self.ros_bridge = ros_bridge

    async def handle_delivery_task(self, payload: dict):
        try:
            task = DeliveryTask(
                delivery_id=payload.get("delivery_id"),
                order_id=payload.get("order_id"),
                good_id=payload.get("good_id"),
                parcel_automat_id=payload.get("parcel_automat_id"),
                target_aruco_id=payload.get(
                    "target_aruco_id", payload.get("aruco_id", 135)
                ),
                home_aruco_id=payload.get("home_aruco_id", 131),
                coordinates=payload.get("coordinates"),
                internal_cell_id=payload.get("internal_cell_id"),
                dimensions=payload.get("dimensions"),
            )

            logger.info("=" * 60)
            logger.info("DELIVERY TASK RECEIVED")
            logger.info(f"  Order ID: {task.order_id}")
            logger.info(f"  Delivery ID: {task.delivery_id}")
            logger.info(f"  Parcel Automat ID: {task.parcel_automat_id}")
            logger.info(f"  Target ArUco: {task.target_aruco_id}")
            logger.info(f"  Home ArUco: {task.home_aruco_id}")
            logger.info(f"  Internal Cell ID: {task.internal_cell_id}")
            logger.info("=" * 60)

            self.state_machine.set_task(task)

            success = self.flight_controller.launch_delivery_flight(
                target_aruco_id=task.target_aruco_id, home_aruco_id=task.home_aruco_id
            )

            if success:
                self.state_machine.transition_to(DeliveryState.TAKING_OFF)
                logger.info("Delivery flight script launched successfully")
            else:
                self.state_machine.transition_to(DeliveryState.FAILED)
                logger.error("Failed to launch delivery flight script")

        except Exception as e:
            logger.error(f"Error handling delivery task: {e}")
            if self.state_machine.current_task:
                self.state_machine.transition_to(DeliveryState.FAILED)

    async def on_arrival(self):
        if not self.state_machine.current_task:
            logger.warning("Arrival event received but no active task")
            return

        task = self.state_machine.current_task
        event_id = f"{task.delivery_id}_arrived"

        if self.state_machine.is_event_sent(event_id):
            logger.debug("Arrival event already sent, skipping")
            return

        self.state_machine.transition_to(DeliveryState.ARRIVED)

        logger.info("Notifying backend of arrival")
        await self.ws_client.send_delivery_update(
            delivery_id=task.delivery_id,
            drone_status="arrived_at_destination",
            order_id=task.order_id,
            parcel_automat_id=task.parcel_automat_id,
        )

        self.state_machine.mark_event_sent(event_id)
        logger.info("Arrival notification sent")

    async def on_drop_ready(self):
        if not self.state_machine.current_task:
            logger.warning("Drop ready event received but no active task")
            return

        self.state_machine.transition_to(DeliveryState.WAITING_CONFIRMATION)
        logger.info("Cargo drop ready, waiting for backend confirmation")

    async def on_home_arrived(self):
        if not self.state_machine.current_task:
            logger.warning("Home arrival event received but no active task")
            return

        self.state_machine.transition_to(DeliveryState.COMPLETED)

        logger.info("Drone returned to home base, mission completed")
        await self.ws_client.send_status_update(
            status="idle",
            battery_level=100.0,
            position={"latitude": 0.0, "longitude": 0.0, "altitude": 0.0},
            speed=0.0,
        )

        self.state_machine.clear_task()

    async def handle_drop_command(self, payload: dict):
        if not self.state_machine.current_task:
            logger.warning("Drop command received but no active task")
            return

        order_id = payload.get("order_id")
        cell_id = payload.get("cell_id")
        internal_cell_id = payload.get("internal_cell_id")

        logger.info("=" * 60)
        logger.info("DROP CARGO COMMAND RECEIVED")
        logger.info(f"  Order ID: {order_id}")
        logger.info(f"  Cell ID: {cell_id}")
        logger.info(f"  Internal Cell ID: {internal_cell_id}")
        logger.info("=" * 60)

        self.state_machine.transition_to(DeliveryState.DROPPING)

        success = self.ros_bridge.send_drop_confirmation()
        if success:
            logger.info("Drop confirmation sent to flight script")
        else:
            logger.error("Failed to send drop confirmation")
