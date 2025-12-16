import asyncio
import json
import logging
import websockets
from typing import Optional
from datetime import datetime
from ..config.settings import Settings
from ..models.messages import MessageType
from ..utils.retry import async_retry

logger = logging.getLogger(__name__)


class WebSocketClient:
    def __init__(self, settings: Settings):
        self.settings = settings
        self.websocket: Optional[websockets.WebSocketClientProtocol] = None
        self.is_connected = False
        self.on_delivery_task = None
        self.on_command = None
        
    async def connect(self):
        while True:
            try:
                logger.info(f"Connecting to {self.settings.websocket_url}")
                self.websocket = await websockets.connect(
                    self.settings.websocket_url,
                    ping_interval=20,
                    ping_timeout=10
                )
                self.is_connected = True
                logger.info("WebSocket connected")
                
                await self._register()
                
                registration_response = await asyncio.wait_for(
                    self.websocket.recv(),
                    timeout=10.0
                )
                
                data = json.loads(registration_response)
                if data.get("type") == "registered":
                    self.settings.drone_id = data.get("drone_id")
                    logger.info(f"Drone registered with ID: {self.settings.drone_id}")
                elif data.get("type") == "error":
                    logger.error(f"Registration failed: {data.get('message')}")
                    raise Exception("Registration failed")
                
                await self._listen()
                
            except websockets.exceptions.ConnectionClosed as e:
                logger.warning(f"Connection closed: {e}")
                self.is_connected = False
                self.settings.drone_id = None
                await self._reconnect()
            except Exception as e:
                logger.error(f"Connection error: {e}")
                self.is_connected = False
                self.settings.drone_id = None
                await self._reconnect()
    
    async def _register(self):
        registration = {
            "type": "register",
            "ip_address": self.settings.drone_ip
        }
        await self.websocket.send(json.dumps(registration))
        logger.info(f"Registration sent with IP: {self.settings.drone_ip}")
    
    async def _listen(self):
        try:
            async for message in self.websocket:
                await self._handle_message(message)
        except websockets.exceptions.ConnectionClosed:
            logger.warning("Connection closed while listening")
            raise
    
    async def _handle_message(self, message: str):
        try:
            data = json.loads(message)
            msg_type = data.get("type")
            payload = data.get("payload", {})
            
            logger.debug(f"Received message type: {msg_type}")
            
            if msg_type == MessageType.DELIVERY_TASK and self.on_delivery_task:
                await self.on_delivery_task(payload)
            elif msg_type == MessageType.COMMAND and self.on_command:
                await self.on_command(payload)
            
        except Exception as e:
            logger.error(f"Error handling message: {e}")
    
    @async_retry(max_attempts=3, delay=0.5)
    async def send_message(self, msg_type: MessageType, payload: dict):
        if not self.is_connected or not self.websocket:
            raise Exception("WebSocket not connected")
        
        message = {
            "type": msg_type,
            "timestamp": datetime.now().isoformat(),
            "payload": payload
        }
        await self.websocket.send(json.dumps(message))
        logger.debug(f"Sent message: {msg_type}")
    
    async def send_heartbeat(self, battery_level: float, position: dict, status: str, speed: float):
        heartbeat = {
            "type": "heartbeat",
            "battery_level": battery_level,
            "position": position,
            "status": status,
            "speed": speed
        }
        await self.websocket.send(json.dumps(heartbeat))
    
    async def send_status_update(self, status: str, battery_level: float, position: dict, speed: float):
        await self.send_message(MessageType.STATUS_UPDATE, {
            "status": status,
            "battery_level": battery_level,
            "position": position,
            "speed": speed
        })
    
    async def send_delivery_update(self, delivery_id: str, drone_status: str, order_id: str = None, parcel_automat_id: str = None):
        await self.send_message(MessageType.DELIVERY_UPDATE, {
            "delivery_id": delivery_id,
            "drone_status": drone_status,
            "order_id": order_id,
            "parcel_automat_id": parcel_automat_id
        })
    
    async def send_video_frame(self, frame_base64: str, delivery_id: str = None):
        video_message = {
            "type": "video_frame",
            "payload": {
                "frame": frame_base64,
                "delivery_id": delivery_id
            }
        }
        await self.websocket.send(json.dumps(video_message))
    
    async def _reconnect(self):
        logger.info(f"Reconnecting in {self.settings.reconnect_interval} seconds...")
        await asyncio.sleep(self.settings.reconnect_interval)
    
    async def close(self):
        self.is_connected = False
        if self.websocket:
            await self.websocket.close()
            logger.info("WebSocket connection closed")
