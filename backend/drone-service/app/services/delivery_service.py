import asyncio
import uuid
from datetime import datetime
from typing import Dict, Any
from app.models.schemas import (
    DeliveryTask, GoodDimensions, DeliveryStatus, DroneState
)
from app.models.ports import DroneHardwarePort, StateRepositoryPort
from app.services.drone_manager_service import DroneManager

class DeliveryUseCase:
    def __init__(
        self,
        drone_hardware: DroneHardwarePort,
        state_repository: StateRepositoryPort,
        drone_manager: DroneManager,
        drone_ws_handler=None
    ):
        self.drone_hardware = drone_hardware
        self.state_repository = state_repository
        self.drone_manager = drone_manager
        self.drone_ws_handler = drone_ws_handler
        
    async def start_delivery(
        self,
        drone_id: str,
        good_id: str,
        parcel_automat_id: str,
        dimensions: Dict[str, float]
    ) -> Dict[str, Any]:
        try:
            delivery_id = str(uuid.uuid4())
            
            task = DeliveryTask(
                delivery_id=delivery_id,
                good_id=good_id,
                parcel_automat_id=parcel_automat_id,
                dimensions=GoodDimensions(**dimensions),
                created_at=datetime.now(),
                drone_id=drone_id
            )
            
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
        good_id: str,
        parcel_automat_id: str,
        weight: float,
        height: float,
        length: float,
        width: float
    ):
        delivery_id = str(uuid.uuid4())
        
        task = DeliveryTask(
            delivery_id=delivery_id,
            good_id=good_id,
            drone_id=drone_id,
            parcel_automat_id=parcel_automat_id,
            dimensions=GoodDimensions(
                weight=weight,
                height=height,
                length=length,
                width=width
            ),
            created_at=datetime.now()
        )
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
                    "good_id": task.good_id,
                    "parcel_automat_id": task.parcel_automat_id,
                    "dimensions": {
                        "weight": task.dimensions.weight,
                        "height": task.dimensions.height,
                        "length": task.dimensions.length,
                        "width": task.dimensions.width
                    }
                }
                success = await self.drone_ws_handler.send_task_to_drone(task.drone_id, task_data)
            else:
                success = await self.drone_hardware.send_delivery_command(task)
            
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
            
            await self.drone_hardware.emergency_stop()
            
            await self.state_repository.update_delivery_status(
                delivery_id,
                DeliveryStatus.CANCELLED
            )
            if task.drone_id:
                await self.drone_manager.release_drone(task.drone_id)
            
            return {"success": True, "message": "Delivery cancelled"}
        except Exception as e:
            return {"success": False, "message": f"Failed to cancel: {str(e)}"}
    
    async def get_drone_status(self, drone_id: str) -> DroneState:
        state = await self.drone_hardware.get_current_state()
        await self.state_repository.save_drone_state(state)
        return state

