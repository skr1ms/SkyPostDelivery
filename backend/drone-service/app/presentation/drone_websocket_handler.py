import json
from fastapi import WebSocket, WebSocketDisconnect
from datetime import datetime
from app.core.entities import DroneState, DroneStatus, Position, DeliveryStatus
from app.core.ports import StateRepositoryPort
from app.core.drone_manager import DroneManager


class DroneWebSocketHandler:
    def __init__(self, state_repository: StateRepositoryPort, drone_manager: DroneManager, delivery_use_case=None):
        self.state_repository = state_repository
        self.drone_manager = drone_manager
        self.connected_drones: dict[str, WebSocket] = {}
        self.ip_to_id: dict[str, str] = {}
        self.delivery_use_case = delivery_use_case

    async def handle_drone_connection(self, websocket: WebSocket):
        await websocket.accept()

        drone_id: str | None = None

        try:
            registration_message = await websocket.receive_text()
            data = json.loads(registration_message)

            if data.get("type") != "register":
                await websocket.send_json({
                    "type": "error",
                    "message": "First message must be registration with ip_address"
                })
                await websocket.close()
                return

            ip_address = data.get("ip_address")
            if not ip_address:
                await websocket.send_json({
                    "type": "error",
                    "message": "ip_address is required for registration"
                })
                await websocket.close()
                return

            drone_id = await self.state_repository.get_drone_id_by_ip(ip_address)

            if not drone_id:
                await websocket.send_json({
                    "type": "error",
                    "message": f"No drone found with IP {ip_address}"
                })
                await websocket.close()
                return

            self.connected_drones[drone_id] = websocket
            self.ip_to_id[ip_address] = drone_id
            await self.drone_manager.register_drone(drone_id)

            await websocket.send_json({
                "type": "registered",
                "drone_id": drone_id,
                "timestamp": datetime.now().isoformat()
            })

            print(f"Drone registered: IP={ip_address}, ID={drone_id}")

            while True:
                data = await websocket.receive_text()
                await self._process_drone_message(drone_id, data)

        except WebSocketDisconnect:
            if drone_id:
                self.connected_drones.pop(drone_id, None)
                for ip, did in list(self.ip_to_id.items()):
                    if did == drone_id:
                        del self.ip_to_id[ip]
                        break
                await self.drone_manager.unregister_drone(drone_id)
                print(f"Drone {drone_id} disconnected")
        except Exception as e:
            print(f"Error in drone connection: {e}")
            if drone_id:
                self.connected_drones.pop(drone_id, None)
                await self.drone_manager.unregister_drone(drone_id)

    async def _process_drone_message(self, drone_id: str, message: str):
        try:
            data = json.loads(message)
            message_type = data.get("type")

            if message_type == "heartbeat":
                await self._handle_heartbeat(drone_id, data)
            elif message_type == "status_update":
                await self._handle_status_update(drone_id, data)
            elif message_type == "delivery_update":
                await self._handle_delivery_update(drone_id, data)
            elif message_type == "arrived_at_destination":
                await self._handle_arrived_at_destination(drone_id, data)
            elif message_type == "cargo_dropped":
                await self._handle_cargo_dropped(drone_id, data)
            else:
                print(
                    f"Unknown message type from drone {drone_id}: {message_type}")
        except Exception as e:
            print(f"Error processing message from drone {drone_id}: {e}")

    async def _handle_heartbeat(self, drone_id: str, data: dict):
        payload = data.get("payload", data)

        battery_level = payload.get("battery_level")
        position = payload.get("position", {})
        status = payload.get("status", "idle")
        
        if battery_level is not None:
            await self.state_repository.update_drone_battery(drone_id, battery_level)
        
        state = DroneState(
            drone_id=drone_id,
            status=DroneStatus(status),
            battery_level=battery_level or 100.0,
            current_position=Position(
                latitude=position.get("latitude", 0.0),
                longitude=position.get("longitude", 0.0),
                altitude=position.get("altitude", 0.0)
            ),
            speed=payload.get("speed", 0.0),
            last_updated=datetime.now(),
            current_delivery_id=payload.get("current_delivery_id"),
            error_message=payload.get("error_message")
        )

        await self.state_repository.save_drone_state(state)
        print(
            f"Heartbeat from {drone_id}: battery={battery_level}%, status={status}")

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
            print(
                f"Invalid delivery update from drone {drone_id}: missing delivery_id")
            return

        if drone_status == "arrived_at_locker":
            await self.state_repository.update_delivery_status(delivery_id, DeliveryStatus.IN_PROGRESS, "Drone arrived at locker, waiting for confirmation")
            print(
                f"Drone {drone_id} arrived at locker for delivery {delivery_id}")

        elif drone_status == "returning":
            await self.drone_manager.release_drone(drone_id)
            print(
                f"Drone {drone_id} returning to base, released from delivery {delivery_id}")

        else:
            print(
                f"Drone {drone_id} status: {drone_status} for delivery {delivery_id}")

    async def send_task_to_drone(self, drone_id: str, task: dict) -> bool:
        if drone_id not in self.connected_drones:
            print(f"Drone {drone_id} not connected")
            return False

        try:
            websocket = self.connected_drones[drone_id]
            await websocket.send_json({
                "type": "delivery_task",
                "timestamp": datetime.now().isoformat(),
                "payload": task
            })
            return True
        except Exception as e:
            print(f"Failed to send task to drone {drone_id}: {e}")
            return False

    async def send_command_to_drone(self, drone_id: str, command: dict) -> bool:
        if drone_id not in self.connected_drones:
            print(f"Drone {drone_id} not connected")
            return False

        try:
            websocket = self.connected_drones[drone_id]
            await websocket.send_json({
                "type": "command",
                "timestamp": datetime.now().isoformat(),
                "payload": command
            })
            return True
        except Exception as e:
            print(f"Failed to send command to drone {drone_id}: {e}")
            return False

    async def _handle_arrived_at_destination(self, drone_id: str, data: dict):
        payload = data.get("payload", {})
        order_id = payload.get("order_id")
        parcel_automat_id = payload.get("parcel_automat_id")

        if not order_id or not parcel_automat_id:
            print(f"Invalid arrived notification from drone {drone_id}")
            return

        print(f"Drone {drone_id} arrived at destination for order {order_id}")

        if self.delivery_use_case:
            result = await self.delivery_use_case.handle_drone_arrived(
                drone_id, order_id, parcel_automat_id
            )
            if not result["success"]:
                print(f"Failed to open cell for drone {drone_id}: {result['message']}")

    async def _handle_cargo_dropped(self, drone_id: str, data: dict):
        payload = data.get("payload", {})
        order_id = payload.get("order_id")
        locker_cell_id = payload.get("locker_cell_id")

        if not order_id:
            print(f"Invalid cargo dropped notification from drone {drone_id}")
            return

        print(f"Drone {drone_id} dropped cargo for order {order_id}")

        if self.delivery_use_case:
            result = await self.delivery_use_case.handle_cargo_dropped(order_id, locker_cell_id)
            if not result["success"]:
                print(f"Failed to confirm cargo drop for order {order_id}: {result['message']}")

    def is_drone_connected(self, drone_id: str) -> bool:
        return drone_id in self.connected_drones
    
    async def send_command_to_drone(self, drone_id: str, command_data: dict) -> bool:
        if drone_id not in self.connected_drones:
            print(f"Cannot send command: drone {drone_id} is not connected")
            return False
        
        try:
            websocket = self.connected_drones[drone_id]
            message = {
                "type": "command",
                "payload": command_data,
                "timestamp": datetime.now().isoformat()
            }
            await websocket.send_json(message)
            print(f"Command sent to drone {drone_id}: {command_data.get('command')}")
            return True
        except Exception as e:
            print(f"Error sending command to drone {drone_id}: {e}")
            return False