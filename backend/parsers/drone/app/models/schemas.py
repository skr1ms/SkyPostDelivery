from pydantic import BaseModel, Field
from typing import Optional
from datetime import datetime
from enum import Enum


class MessageType(str, Enum):
    DELIVERY_TASK = "delivery_task"
    STATUS_UPDATE = "status_update"
    DELIVERY_UPDATE = "delivery_update"
    HEARTBEAT = "heartbeat"
    COMMAND = "command"
    ERROR = "error"


class DroneStatus(str, Enum):
    IDLE = "idle"
    TAKING_OFF = "taking_off"
    PICKING_UP = "picking_up"
    IN_TRANSIT = "in_transit"
    DELIVERING = "delivering"
    RETURNING = "returning"
    LANDING = "landing"
    CHARGING = "charging"
    ERROR = "error"
    MAINTENANCE = "maintenance"


class Position(BaseModel):
    latitude: float
    longitude: float
    altitude: float


class GoodDimensions(BaseModel):
    weight: float
    height: float
    length: float
    width: float


class DeliveryTaskPayload(BaseModel):
    delivery_id: str
    order_id: str
    good_id: str
    parcel_automat_id: str
    aruco_id: int
    coordinates: Optional[str] = None
    dimensions: GoodDimensions


class StatusUpdatePayload(BaseModel):
    status: DroneStatus
    battery_level: float
    position: Position
    speed: float
    current_delivery_id: Optional[str] = None
    error_message: Optional[str] = None


class DeliveryUpdatePayload(BaseModel):
    delivery_id: str
    drone_status: str
    message: Optional[str] = None
    position: Optional[Position] = None
    order_id: Optional[str] = None
    parcel_automat_id: Optional[str] = None


class IncomingMessage(BaseModel):
    type: MessageType
    timestamp: str
    payload: dict


class OutgoingMessage(BaseModel):
    type: MessageType
    timestamp: str = Field(default_factory=lambda: datetime.now().isoformat())
    payload: dict
