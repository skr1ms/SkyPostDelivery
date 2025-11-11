from typing import Optional, Dict
from app.models.models import DroneState, DroneStatus
from app.models.ports import StateRepositoryPort


class DroneManager:
    def __init__(self, state_repository: StateRepositoryPort):
        self.state_repository = state_repository
        self.registered_drones: Dict[str, DroneState] = {}

    async def register_drone(self, drone_id: str):
        self.registered_drones[drone_id] = None

    async def get_free_drone(self) -> Optional[str]:
        for drone_id in self.registered_drones.keys():
            state = await self.state_repository.get_drone_state(drone_id)

            if state and state.status == DroneStatus.IDLE and not state.current_delivery_id:
                if state.battery_level > 30.0:
                    return drone_id

        return None

    async def assign_delivery_to_drone(self, drone_id: str, delivery_id: str) -> bool:
        state = await self.state_repository.get_drone_state(drone_id)
        if state:
            state.current_delivery_id = delivery_id
            state.status = DroneStatus.TAKING_OFF
            await self.state_repository.save_drone_state(state)
            return True
        return False

    async def release_drone(self, drone_id: str) -> bool:
        state = await self.state_repository.get_drone_state(drone_id)
        if state:
            state.current_delivery_id = None
            state.status = DroneStatus.IDLE
            await self.state_repository.save_drone_state(state)
            return True
        return False

    async def unregister_drone(self, drone_id: str):
        self.registered_drones.pop(drone_id, None)

    async def get_drone_state(self, drone_id: str) -> Optional[DroneState]:
        return await self.state_repository.get_drone_state(drone_id)

    def get_all_drones(self) -> list:
        return list(self.registered_drones.keys())

    def get_registered_drones(self) -> list[str]:
        return list(self.registered_drones.keys())
