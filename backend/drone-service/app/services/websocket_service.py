import json
from fastapi import WebSocket, WebSocketDisconnect
from datetime import datetime
from app.models.schemas import DroneState, DroneStatus, Position, DeliveryStatus
from app.models.ports import StateRepositoryPort
from app.services.drone_manager_service import DroneManager


class DroneWebSocketHandler:
    def __init__(self, state_repository: StateRepositoryPort, drone_manager: DroneManager, delivery_use_case=None):
        self.state_repository = state_repository
        self.drone_manager = drone_manager
        self.connected_drones: dict[str, WebSocket] = {}
        self.delivery_use_case = delivery_use_case
    
    async def handle_drone_connection(self, websocket: WebSocket, drone_id: str):
        await websocket.accept()
        self.connected_drones[drone_id] = websocket
        await self.drone_manager.register_drone(drone_id)
        
        try:
            while True:
                data = await websocket.receive_text()
                await self._process_drone_message(drone_id, data)
        except WebSocketDisconnect:
            del self.connected_drones[drone_id]
            await self.drone_manager.unregister_drone(drone_id)
    
    async def _process_drone_message(self, drone_id: str, message: str):
        try:
            data = json.loads(message)
            message_type = data.get("type")
            
            if message_type == "status_update":
                await self._handle_status_update(drone_id, data)
            elif message_type == "delivery_update":
                await self._handle_delivery_update(drone_id, data)
            elif message_type == "arrived_at_destination":
                await self._handle_arrived_at_destination(drone_id, data)
            elif message_type == "cargo_dropped":
                await self._handle_cargo_dropped(drone_id, data)
        except Exception:
            pass
    
    async def _handle_status_update(self, drone_id: str, data: dict):
        payload = data.get("payload", {})
        
        state = DroneState(
            drone_id=drone_id,
            status=DroneStatus(payload.get("status", "idle")),
            battery_level=payload.get("battery_level", 0.0),
            current_position=Position(
                latitude=payload.get("position", {}).get("latitude", 0.0),
                longitude=payload.get("position", {}).get("longitude", 0.0),
                altitude=payload.get("position", {}).get("altitude", 0.0)
            ),
            speed=payload.get("speed", 0.0),
            last_updated=datetime.now(),
            current_delivery_id=payload.get("current_delivery_id"),
            error_message=payload.get("error_message")
        )
        
        await self.state_repository.save_drone_state(state)
    
    async def _handle_delivery_update(self, drone_id: str, data: dict):
        payload = data.get("payload", {})
        delivery_id = payload.get("delivery_id")
        drone_status = payload.get("drone_status")
        
        if not delivery_id:
            return
        
        if drone_status == "arrived_at_locker":
            await self.state_repository.update_delivery_status(delivery_id, DeliveryStatus.IN_PROGRESS, "Drone arrived at locker, waiting for confirmation")
            
        elif drone_status == "returning":
            await self.drone_manager.release_drone(drone_id)
    
    async def send_task_to_drone(self, drone_id: str, task: dict) -> bool:
        if drone_id not in self.connected_drones:
            return False
        
        try:
            websocket = self.connected_drones[drone_id]
            await websocket.send_json({
                "type": "delivery_task",
                "timestamp": datetime.now().isoformat(),
                "payload": task
            })
            return True
        except Exception:
            return False

    async def send_command_to_drone(self, drone_id: str, command: dict) -> bool:
        if drone_id not in self.connected_drones:
            return False
        
        try:
            websocket = self.connected_drones[drone_id]
            await websocket.send_json({
                "type": "command",
                "timestamp": datetime.now().isoformat(),
                "payload": command
            })
            return True
        except Exception:
            return False

    async def _handle_arrived_at_destination(self, drone_id: str, data: dict):
        payload = data.get("payload", {})
        order_id = payload.get("order_id")
        parcel_automat_id = payload.get("parcel_automat_id")
        
        if not order_id or not parcel_automat_id:
            return
        
        if self.delivery_use_case:
            await self.delivery_use_case.handle_drone_arrived(
                drone_id, order_id, parcel_automat_id
            )
    
    async def _handle_cargo_dropped(self, drone_id: str, data: dict):
        payload = data.get("payload", {})
        order_id = payload.get("order_id")
        locker_cell_id = payload.get("locker_cell_id")
        
        if not order_id:
            return
        
        if self.delivery_use_case:
            await self.delivery_use_case.handle_cargo_dropped(order_id, locker_cell_id)
    
    def is_drone_connected(self, drone_id: str) -> bool:
        return drone_id in self.connected_drones

