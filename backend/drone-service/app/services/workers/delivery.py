import logging
from uuid import UUID
from app.hardware.rabbitmq import RabbitMQClient
from app.services.use_cases import DeliveryUseCase

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)


class DeliveryWorker:
    def __init__(
        self,
        rabbitmq_client: RabbitMQClient,
        delivery_use_case: DeliveryUseCase
    ):
        self.rabbitmq_client = rabbitmq_client
        self.delivery_use_case = delivery_use_case
        self._running = False

    async def start(self):
        self._running = True

        logger.info("Starting delivery worker, setting up consumers...")

        await self.rabbitmq_client.consume(
            "deliveries",
            self.handle_delivery_task
        )
        await self.rabbitmq_client.consume(
            "deliveries.priority",
            self.handle_delivery_task
        )
        await self.rabbitmq_client.consume(
            "delivery.return",
            self.handle_return_task
        )

        logger.info(
            "Delivery worker started successfully (consuming: deliveries, deliveries.priority, delivery.return)")

    async def handle_delivery_task(self, message: dict):
        try:
            logger.info(f"Received raw delivery message: {message}")
            print(f"[WORKER] Received message: {message}")
            drone_id = UUID(message["drone_id"])
            order_id = UUID(message["order_id"])
            good_id = UUID(message["good_id"])
            parcel_automat_id = UUID(message["parcel_automat_id"])
            aruco_id = int(message["aruco_id"])
            coordinates = message.get("coordinates", "")
            weight = float(message["weight"])
            height = float(message["height"])
            length = float(message["length"])
            width = float(message["width"])
            priority = int(message.get("priority", 0))
            internal_cell_id_str = message.get("internal_locker_cell_id")
            internal_cell_id = str(UUID(internal_cell_id_str)) if internal_cell_id_str else None
            drone_ip = message.get("drone_ip", "")

            logger.info(
                f"Processing delivery task: "
                f"drone_id={drone_id}, order_id={order_id}, good_id={good_id}, aruco_id={aruco_id}, "
                f"coordinates={coordinates}, priority={priority}, drone_ip={drone_ip}"
            )

            logger.info(
                "Dispatching task to DeliveryUseCase",
                extra={
                    "drone_id": str(drone_id),
                    "order_id": str(order_id),
                    "aruco_id": aruco_id
                }
            )
            print(
                f"[WORKER] Dispatching task to use case: drone={drone_id}, order={order_id}, aruco={aruco_id}")

            await self.delivery_use_case.execute_delivery(
                drone_id=str(drone_id),
                order_id=str(order_id),
                good_id=str(good_id),
                parcel_automat_id=str(parcel_automat_id),
                aruco_id=aruco_id,
                coordinates=coordinates,
                weight=weight,
                height=height,
                length=length,
                width=width,
                internal_locker_cell_id=internal_cell_id,
            )

            logger.info(
                f"Successfully handed over delivery task for order {order_id}")
            print(f"[WORKER] Task handed over for order {order_id}")

        except Exception as e:
            logger.error(
                f"Failed to process delivery task: {e}", exc_info=True)
            raise

    async def handle_return_task(self, message: dict):
        """
        Handle return task from RabbitMQ.
        Message format:
        {
            "drone_id": uuid,
            "aruco_id": int (base marker, default 131)
        }
        """
        try:
            drone_id = UUID(message["drone_id"])
            base_marker_id = int(message.get("aruco_id", 131))

            logger.info(
                f"Processing return task: drone_id={drone_id}, base_marker={base_marker_id}"
            )

            await self.delivery_use_case.send_return_command(
                drone_id=str(drone_id),
                base_marker_id=base_marker_id
            )

            logger.info(
                f"Successfully sent return command to drone {drone_id}")

        except Exception as e:
            logger.error(f"Failed to process return task: {e}", exc_info=True)
            raise

    async def stop(self):
        self._running = False
        await self.rabbitmq_client.close()
        logger.info("Delivery worker stopped")
