import json
import asyncio
from fastapi import WebSocket, WebSocketDisconnect
from datetime import datetime
from typing import Set
from app.core.ports import StateRepositoryPort
from app.core.drone_manager import DroneManager
from config.config import settings


class AdminWebSocketHandler:
    def __init__(
        self,
        state_repository: StateRepositoryPort,
        drone_manager: DroneManager
    ):
        self.state_repository = state_repository
        self.drone_manager = drone_manager
        self.admin_connections: Set[WebSocket] = set()
        self.broadcast_task = None

    async def handle_admin_connection(self, websocket: WebSocket):
        await websocket.accept()
        self.admin_connections.add(websocket)
        print(
            f"Admin panel connected. Total connections: {len(self.admin_connections)}")

        if len(self.admin_connections) == 1:
            self.broadcast_task = asyncio.create_task(self._broadcast_loop())

        try:
            while True:
                message = await websocket.receive_text()
                data = json.loads(message)
                if data.get("type") == "ping":
                    await websocket.send_json({"type": "pong", "timestamp": datetime.now().isoformat()})
        except WebSocketDisconnect:
            self.admin_connections.discard(websocket)
            print(
                f"Admin panel disconnected. Total connections: {len(self.admin_connections)}")

            if len(self.admin_connections) == 0 and self.broadcast_task:
                self.broadcast_task.cancel()
                self.broadcast_task = None

    async def _broadcast_loop(self):

        while True:
            try:
                await asyncio.sleep(settings.WEBSOCKET_BROADCAST_INTERVAL)

                if not self.admin_connections:
                    continue

                drones_status = await self._get_all_drones_status()

                message = {
                    "type": "drones_status",
                    "timestamp": datetime.now().isoformat(),
                    "drones": drones_status
                }

                await self._broadcast(message)

            except asyncio.CancelledError:
                print("Broadcast loop cancelled")
                break
            except Exception as e:
                print(f"Error in broadcast loop: {e}")
                await asyncio.sleep(1)

    async def _get_all_drones_status(self) -> list:
        drones_status = []

        all_drones = self.drone_manager.get_all_drones()

        for drone_id in all_drones:
            try:
                state = await self.state_repository.get_drone_state(drone_id)
                if state:
                    drones_status.append({
                        "drone_id": state.drone_id,
                        "status": state.status.value,
                        "battery_level": state.battery_level,
                        "position": {
                            "latitude": state.current_position.latitude,
                            "longitude": state.current_position.longitude,
                            "altitude": state.current_position.altitude
                        },
                        "speed": state.speed,
                        "current_delivery_id": state.current_delivery_id,
                        "error_message": state.error_message,
                        "last_updated": state.last_updated.isoformat() if state.last_updated else None
                    })
                else:
                    drones_status.append({
                        "drone_id": drone_id,
                        "status": "offline",
                        "battery_level": 0.0,
                        "position": {"latitude": 0.0, "longitude": 0.0, "altitude": 0.0},
                        "speed": 0.0,
                        "current_delivery_id": None,
                        "error_message": None,
                        "last_updated": None
                    })
            except Exception as e:
                print(f"Error getting status for drone {drone_id}: {e}")

        return drones_status

    async def _broadcast(self, message: dict):
        disconnected = []

        for connection in self.admin_connections:
            try:
                await connection.send_json(message)
            except WebSocketDisconnect:
                disconnected.append(connection)
            except Exception as e:
                print(f"Error broadcasting to admin connection: {e}")
                disconnected.append(connection)

        for connection in disconnected:
            self.admin_connections.discard(connection)

        if disconnected:
            print(
                f"Removed {len(disconnected)} disconnected admin connections")
