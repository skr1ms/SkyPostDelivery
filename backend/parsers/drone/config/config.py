import os
import cv2
from typing import Optional
from dotenv import load_dotenv

load_dotenv()


class Settings:
    drone_ip: str = os.getenv("DRONE_IP", "192.168.10.3")

    drone_id: Optional[str] = None

    drone_service_host: str = os.getenv("DRONE_SERVICE_HOST", "localhost")
    drone_service_port: int = int(os.getenv("DRONE_SERVICE_PORT", "8001"))

    parcel_automat_ip: str = os.getenv("PARCEL_AUTOMAT_IP", "192.168.10.2")

    reconnect_interval: int = int(os.getenv("RECONNECT_INTERVAL", "5"))
    heartbeat_interval: int = int(os.getenv("HEARTBEAT_INTERVAL", "30"))

    camera_index: int = int(os.getenv("CAMERA_INDEX", "0"))
    video_fps: int = int(os.getenv("VIDEO_FPS", "5"))

    log_level: str = os.getenv("LOG_LEVEL", "INFO")

    use_mock_hardware: bool = os.getenv(
        "USE_MOCK_HARDWARE", "false").lower() == "true"

    aruco_dict_type: str = os.getenv("ARUCO_DICT_TYPE", "DICT_6X6_250")
    marker_size_cm: float = float(os.getenv("MARKER_SIZE_CM", "33.0"))

    cruise_altitude: float = float(os.getenv("CRUISE_ALTITUDE", "1.5"))
    cruise_speed: float = float(os.getenv("CRUISE_SPEED", "0.5"))
    landing_altitude: float = float(os.getenv("LANDING_ALTITUDE", "0.5"))
    center_threshold: float = float(os.getenv("CENTER_THRESHOLD", "0.05"))
    distance_between_markers: float = float(
        os.getenv("DISTANCE_BETWEEN_MARKERS", "0.85"))

    field_width: float = float(os.getenv("FIELD_WIDTH", "10.0"))
    field_height: float = float(os.getenv("FIELD_HEIGHT", "12.0"))
    grid_width: int = int(os.getenv("GRID_WIDTH", "10"))
    grid_height: int = int(os.getenv("GRID_HEIGHT", "10"))

    aruco_map_file: str = os.getenv("ARUCO_MAP_FILE", "config/aruco_map.txt")
    use_clover_api: bool = os.getenv(
        "USE_CLOVER_API", "true").lower() == "true"

    fc_port: str = os.getenv("FC_PORT", "/dev/ttyAMA0")
    fc_baudrate: int = int(os.getenv("FC_BAUDRATE", "115200"))
    fc_timeout: float = float(os.getenv("FC_TIMEOUT", "1.0"))

    marker_detection_interval: float = float(
        os.getenv("MARKER_DETECTION_INTERVAL", "0.2"))
    telemetry_interval: float = float(os.getenv("TELEMETRY_INTERVAL", "0.5"))

    target_marker_id: int = int(os.getenv("TARGET_MARKER_ID", "52"))

    @property
    def websocket_url(self) -> str:
        return f"ws://{self.drone_service_host}:{self.drone_service_port}/ws/drone"

    @property
    def aruco_dict(self):
        return getattr(cv2.aruco, self.aruco_dict_type)

    @property
    def marker_size_m(self) -> float:
        return self.marker_size_cm / 100.0


settings = Settings()
