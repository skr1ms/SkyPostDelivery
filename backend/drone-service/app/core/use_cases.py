import asyncio
import uuid
from datetime import datetime
from typing import Dict, Any
from app.core.entities import (
    DeliveryTask, GoodDimensions, DeliveryStatus
)
from app.core.ports import StateRepositoryPort
from app.core.drone_manager import DroneManager


class DeliveryUseCase:
    def __init__(
        self,
        state_repository: StateRepositoryPort,
        drone_manager: DroneManager,
        drone_ws_handler=None,
        orchestrator_grpc_client=None,
        rabbitmq_client=None
    ):
        self.state_repository = state_repository
        self.drone_manager = drone_manager
        self.drone_ws_handler = drone_ws_handler
        self.orchestrator_grpc_client = orchestrator_grpc_client
        self.rabbitmq_client = rabbitmq_client

    async def start_delivery(
        self,
        drone_id: str,
        order_id: str,
        good_id: str,
        locker_cell_id: str,
        parcel_automat_id: str,
        aruco_id: int,
        dimensions: Dict[str, float]
    ) -> Dict[str, Any]:
        try:
            delivery_id = str(uuid.uuid4())

            task = DeliveryTask(
                delivery_id=delivery_id,
                order_id=order_id,
                good_id=good_id,
                locker_cell_id=locker_cell_id,
                parcel_automat_id=parcel_automat_id,
                dimensions=GoodDimensions(**dimensions),
                created_at=datetime.now(),
                drone_id=drone_id
            )
            task.aruco_id = aruco_id

            await self.state_repository.save_delivery_task(task)

            if drone_id not in self.drone_manager.registered_drones:
                await self.drone_manager.register_drone(drone_id)

            await self.drone_manager.assign_delivery_to_drone(drone_id, delivery_id)

            asyncio.create_task(self._execute_delivery(task))

            return {
                "success": True,
                "message": f"Delivery initiated with drone {drone_id}",
                "delivery_id": delivery_id
            }
        except Exception as e:
            return {
                "success": False,
                "message": f"Failed to start delivery: {str(e)}",
                "delivery_id": ""
            }

    async def execute_delivery(
        self,
        drone_id: str,
        order_id: str,
        good_id: str,
        parcel_automat_id: str,
        aruco_id: int,
        coordinates: str = "",
        weight: float = 0,
        height: float = 0,
        length: float = 0,
        width: float = 0
    ):
        task = DeliveryTask(
            delivery_id=order_id,
            order_id=order_id,
            good_id=good_id,
            drone_id=drone_id,
            locker_cell_id="",
            parcel_automat_id=parcel_automat_id,
            dimensions=GoodDimensions(
                weight=weight,
                height=height,
                length=length,
                width=width
            )
        )
        task.aruco_id = aruco_id
        task.coordinates = coordinates
        await self._execute_delivery(task)

    async def _execute_delivery(self, task: DeliveryTask):
        try:
            task.started_at = datetime.now()
            await self.state_repository.update_delivery_status(
                task.delivery_id,
                DeliveryStatus.IN_PROGRESS
            )

            if self.drone_ws_handler:
                task_data = {
                    "delivery_id": task.delivery_id,
                    "order_id": task.order_id,
                    "good_id": task.good_id,
                    "parcel_automat_id": task.parcel_automat_id,
                    "aruco_id": task.aruco_id,
                    "dimensions": {
                        "weight": task.dimensions.weight,
                        "height": task.dimensions.height,
                        "length": task.dimensions.length,
                        "width": task.dimensions.width
                    }
                }
                success = await self.drone_ws_handler.send_task_to_drone(task.drone_id, task_data)
            else:
                success = False

            if not success:
                raise Exception("Failed to send task to drone")

        except Exception as e:
            task.error_message = str(e)
            await self.state_repository.update_delivery_status(
                task.delivery_id,
                DeliveryStatus.FAILED,
                str(e)
            )
            if task.drone_id:
                await self.drone_manager.release_drone(task.drone_id)

    async def cancel_delivery(self, delivery_id: str) -> Dict[str, Any]:
        try:
            task = await self.state_repository.get_delivery_task(delivery_id)
            if not task:
                return {"success": False, "message": "Delivery not found"}

            if self.drone_ws_handler and task.drone_id:
                await self.drone_ws_handler.send_command_to_drone(
                    task.drone_id,
                    {"command": "cancel_delivery"}
                )

            await self.state_repository.update_delivery_status(
                delivery_id,
                DeliveryStatus.CANCELLED
            )
            if task.drone_id:
                await self.drone_manager.release_drone(task.drone_id)

            return {"success": True, "message": "Delivery cancelled"}
        except Exception as e:
            return {"success": False, "message": f"Failed to cancel: {str(e)}"}

    async def get_delivery_status(self, delivery_id: str) -> Dict[str, Any]:
        task = await self.state_repository.get_delivery_task(delivery_id)
        if not task:
            return {"success": False, "message": "Delivery not found"}

        return {
            "success": True,
            "delivery_id": task.delivery_id,
            "status": task.status.value if task.status else "unknown",
            "drone_id": task.drone_id
        }

    async def handle_drone_arrived(self, drone_id: str, order_id: str, parcel_automat_id: str) -> Dict[str, Any]:
        try:
            if not self.orchestrator_grpc_client:
                return {"success": False, "message": "Orchestrator client not configured"}

            response = await self.orchestrator_grpc_client.request_cell_open(
                order_id=order_id,
                parcel_automat_id=parcel_automat_id
            )

            if response["success"]:
                if self.drone_ws_handler:
                    await self.drone_ws_handler.send_command_to_drone(drone_id, {
                        "command": "drop_cargo",
                        "order_id": order_id,
                        "cell_id": response["cell_id"]
                    })
                return response
            else:
                return {"success": False, "message": f"Failed to open cell: {response['message']}"}
        except Exception as e:
            return {"success": False, "message": f"Failed to handle drone arrival: {str(e)}"}

    async def handle_cargo_dropped(self, order_id: str, locker_cell_id: str) -> Dict[str, Any]:
        try:
            task = await self.state_repository.get_delivery_task(order_id)
            if not task:
                return {"success": False, "message": "Delivery task not found"}

            await self.state_repository.update_delivery_status(
                order_id,
                DeliveryStatus.COMPLETED
            )

            if task.drone_id:
                await self.drone_manager.release_drone(task.drone_id)

            if self.rabbitmq_client:
                confirmation_message = {
                    "order_id": order_id,
                    "locker_cell_id": locker_cell_id or task.locker_cell_id
                }
                await self.rabbitmq_client.publish("confirmations", confirmation_message)
                print(
                    f"Sent delivery confirmation to RabbitMQ for order {order_id}")

            return {"success": True, "message": "Cargo dropped successfully"}
        except Exception as e:
            return {"success": False, "message": f"Failed to handle cargo drop: {str(e)}"}
    
    async def send_return_command(self, drone_id: str, base_marker_id: int = 131) -> Dict[str, Any]:
        """Отправляет команду дрону вернуться на базовую метку"""
        try:
            if not self.drone_ws_handler:
                return {"success": False, "message": "WebSocket handler not available"}
            
            # Отправляем команду через WebSocket
            command_data = {
                "command": "return_to_base",
                "base_marker_id": base_marker_id
            }
            
            success = await self.drone_ws_handler.send_command_to_drone(drone_id, command_data)
            
            if success:
                # Обновляем статус дрона в менеджере
                if drone_id in self.drone_manager.registered_drones:
                    await self.drone_manager.release_drone(drone_id)
                
                return {"success": True, "message": f"Return command sent to drone {drone_id}"}
            else:
                return {"success": False, "message": f"Failed to send return command to drone {drone_id}"}
                
        except Exception as e:
            return {"success": False, "message": f"Failed to send return command: {str(e)}"}
