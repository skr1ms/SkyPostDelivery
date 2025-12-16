from dataclasses import dataclass
from typing import Optional


@dataclass
class Battery:
    voltage: float
    percentage: Optional[float] = None
    current: Optional[float] = None


@dataclass
class Pose:
    x: float
    y: float
    z: float
    orientation: Optional[dict] = None


@dataclass
class MavrosState:
    armed: bool
    connected: bool
    mode: str


@dataclass
class Telemetry:
    battery: Optional[Battery] = None
    pose: Optional[Pose] = None
    altitude: Optional[float] = None
    state: Optional[MavrosState] = None
