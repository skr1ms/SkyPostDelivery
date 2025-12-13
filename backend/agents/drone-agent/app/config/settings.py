import os
from dataclasses import dataclass, field
from typing import Optional
from dotenv import load_dotenv

load_dotenv()


@dataclass
class Settings:
    drone_ip: str = field(default_factory=lambda: os.getenv("DRONE_IP", "192.168.10.3"))
    drone_id: Optional[str] = None

    drone_service_host: str = field(
        default_factory=lambda: os.getenv("DRONE_SERVICE_HOST", "localhost")
    )
    drone_service_port: int = field(
        default_factory=lambda: int(os.getenv("DRONE_SERVICE_PORT", "8001"))
    )

    parcel_automat_ip: str = field(
        default_factory=lambda: os.getenv("PARCEL_AUTOMAT_IP", "192.168.10.2")
    )

    reconnect_interval: int = field(
        default_factory=lambda: int(os.getenv("RECONNECT_INTERVAL", "5"))
    )
    heartbeat_interval: int = field(
        default_factory=lambda: int(os.getenv("HEARTBEAT_INTERVAL", "30"))
    )

    video_fps: int = field(default_factory=lambda: int(os.getenv("VIDEO_FPS", "5")))

    log_level: str = field(default_factory=lambda: os.getenv("LOG_LEVEL", "INFO"))

    use_mock_hardware: bool = field(
        default_factory=lambda: os.getenv("USE_MOCK_HARDWARE", "false").lower()
        == "true"
    )

    scripts_dir: str = field(default_factory=lambda: os.getenv("SCRIPTS_DIR", "/root"))

    @property
    def websocket_url(self) -> str:
        return f"ws://{self.drone_service_host}:{self.drone_service_port}/ws/drone"


settings = Settings()
