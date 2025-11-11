from fastapi import WebSocket, WebSocketDisconnect
from typing import Dict, Set
import asyncio
import base64
from app.hardware.minio import MinIOClient


class DroneVideoProxyHandler:
    def __init__(self, drone_ws_handler):
        self.drone_ws_handler = drone_ws_handler
        self.admin_video_connections: Dict[str, Set[WebSocket]] = {}
        self.minio_client = MinIOClient()
        self.frame_counters: Dict[str, int] = {}
        
    async def handle_admin_video_connection(self, websocket: WebSocket, drone_id: str):
        await websocket.accept()
        
        if drone_id not in self.admin_video_connections:
            self.admin_video_connections[drone_id] = set()
        
        self.admin_video_connections[drone_id].add(websocket)
        
        try:
            while True:
                await websocket.receive_text()
        except WebSocketDisconnect:
            self.admin_video_connections[drone_id].discard(websocket)
            if len(self.admin_video_connections[drone_id]) == 0:
                del self.admin_video_connections[drone_id]
    
    async def broadcast_frame_to_admins(self, drone_id: str, frame_data: str, delivery_id: str = None):
        if drone_id not in self.admin_video_connections:
            return
        
        disconnected = []
        for connection in self.admin_video_connections[drone_id]:
            try:
                await connection.send_text(frame_data)
            except Exception:
                disconnected.append(connection)
        
        for connection in disconnected:
            self.admin_video_connections[drone_id].discard(connection)
        
        if delivery_id:
            asyncio.create_task(self._save_frame_to_minio(drone_id, delivery_id, frame_data))
    
    async def _save_frame_to_minio(self, drone_id: str, delivery_id: str, frame_data: str):
        try:
            if drone_id not in self.frame_counters:
                self.frame_counters[drone_id] = 0
            
            self.frame_counters[drone_id] += 1
            frame_number = self.frame_counters[drone_id]
            
            frame_bytes = base64.b64decode(frame_data)
            
            await self.minio_client.upload_frame(drone_id, delivery_id, frame_bytes, frame_number)
        except Exception:
            pass
