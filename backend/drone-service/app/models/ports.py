from abc import ABC, abstractmethod
from typing import Optional
from app.models.models import DroneState, DeliveryTask, DeliveryStatus


class DroneHardwarePort(ABC):
    @abstractmethod
    async def connect(self) -> bool:
        pass

    @abstractmethod
    async def disconnect(self) -> bool:
        pass

    @abstractmethod
    async def send_delivery_command(self, task: DeliveryTask) -> bool:
        pass

    @abstractmethod
    async def get_current_state(self) -> DroneState:
        pass

    @abstractmethod
    async def emergency_stop(self) -> bool:
        pass


class StateRepositoryPort(ABC):
    @abstractmethod
    async def save_drone_state(self, state: DroneState) -> bool:
        pass

    @abstractmethod
    async def get_drone_state(self, drone_id: str) -> Optional[DroneState]:
        pass

    @abstractmethod
    async def get_drone_id_by_ip(self, ip_address: str) -> Optional[str]:
        pass

    @abstractmethod
    async def update_drone_battery(self, drone_id: str, battery_level: float) -> bool:
        pass

    @abstractmethod
    async def save_delivery_task(self, task: DeliveryTask) -> bool:
        pass

    @abstractmethod
    async def get_delivery_task(self, delivery_id: str) -> Optional[DeliveryTask]:
        pass

    @abstractmethod
    async def update_delivery_status(
        self,
        delivery_id: str,
        status: DeliveryStatus,
        error_message: Optional[str] = None
    ) -> bool:
        pass
