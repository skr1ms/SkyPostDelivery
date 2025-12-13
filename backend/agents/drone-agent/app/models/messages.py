from dataclasses import dataclass
from typing import Optional
from enum import Enum


class MessageType(str, Enum):
    REGISTER = "register"
    REGISTERED = "registered"
    DELIVERY_TASK = "delivery_task"
    STATUS_UPDATE = "status_update"
    DELIVERY_UPDATE = "delivery_update"
    HEARTBEAT = "heartbeat"
    COMMAND = "command"
    VIDEO_FRAME = "video_frame"
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


@dataclass
class Position:
    latitude: float
    longitude: float
    altitude: float


@dataclass
class GoodDimensions:
    weight: float
    height: float
    length: float
    width: float


@dataclass
class IncomingMessage:
    type: MessageType
    timestamp: str
    payload: dict


@dataclass
class OutgoingMessage:
    type: MessageType
    payload: dict
    timestamp: Optional[str] = None
