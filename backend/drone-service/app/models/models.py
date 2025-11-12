from dataclasses import dataclass
from typing import Optional
from datetime import datetime
from enum import Enum


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


class DeliveryStatus(str, Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"


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
class DroneState:
    drone_id: str
    status: DroneStatus
    battery_level: float
    current_position: Position
    speed: float
    last_updated: datetime
    current_delivery_id: Optional[str] = None
    error_message: Optional[str] = None


@dataclass
class DeliveryTask:
    delivery_id: str
    order_id: str
    good_id: str
    locker_cell_id: str
    parcel_automat_id: str
    dimensions: GoodDimensions
    created_at: datetime
    internal_locker_cell_id: Optional[str] = None
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    drone_id: Optional[str] = None
    error_message: Optional[str] = None
    aruco_id: Optional[int] = None
