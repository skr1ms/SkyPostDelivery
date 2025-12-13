from dataclasses import dataclass
from typing import Optional
from enum import Enum


class DeliveryState(str, Enum):
    IDLE = "idle"
    PENDING = "pending"
    TAKING_OFF = "taking_off"
    NAVIGATING = "navigating"
    LANDING = "landing"
    ARRIVED = "arrived"
    WAITING_CONFIRMATION = "waiting_confirmation"
    DROPPING = "dropping"
    RETURNING = "returning"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"


@dataclass
class DeliveryTask:
    delivery_id: str
    order_id: str
    good_id: str
    parcel_automat_id: str
    target_aruco_id: int
    home_aruco_id: int = 131
    coordinates: Optional[str] = None
    internal_cell_id: Optional[str] = None
    dimensions: Optional[dict] = None
    state: DeliveryState = DeliveryState.PENDING
