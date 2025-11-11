from typing import Dict, Optional
from app.core.entities import DroneState, DeliveryTask, DeliveryStatus
from app.core.ports import StateRepositoryPort


class InMemoryStateRepository(StateRepositoryPort):
    def __init__(self):
        self.drones: Dict[str, DroneState] = {}
        self.deliveries: Dict[str, DeliveryTask] = {}
        self.drone_ip_map: Dict[str, str] = {}
    
    async def save_drone_state(self, state: DroneState) -> bool:
        try:
            self.drones[state.drone_id] = state
            return True
        except Exception as e:
            print(f"Failed to save drone state: {e}")
            return False
    
    async def get_drone_state(self, drone_id: str) -> Optional[DroneState]:
        return self.drones.get(drone_id)
    
    async def get_drone_id_by_ip(self, ip_address: str) -> Optional[str]:
        return self.drone_ip_map.get(ip_address)
    
    async def update_drone_battery(self, drone_id: str, battery_level: float) -> bool:
        try:
            state = self.drones.get(drone_id)
            if state:
                state.battery_level = battery_level
                return True
            return False
        except Exception as e:
            print(f"Failed to update battery: {e}")
            return False
    
    async def save_delivery_task(self, task: DeliveryTask) -> bool:
        try:
            self.deliveries[task.delivery_id] = task
            return True
        except Exception as e:
            print(f"Failed to save delivery task: {e}")
            return False
    
    async def get_delivery_task(self, delivery_id: str) -> Optional[DeliveryTask]:
        return self.deliveries.get(delivery_id)
    
    async def update_delivery_status(
        self, 
        delivery_id: str, 
        status: DeliveryStatus,
        error_message: Optional[str] = None
    ) -> bool:
        try:
            task = self.deliveries.get(delivery_id)
            if task:
                task.status = status
                if error_message:
                    task.error_message = error_message
                return True
            return False
        except Exception as e:
            print(f"Failed to update delivery status: {e}")
            return False

