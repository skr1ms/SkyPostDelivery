import os
from pathlib import Path
from dotenv import load_dotenv

load_dotenv()


class Settings:
    api_host: str = os.getenv("SERVICE_HOST", "0.0.0.0")
    api_port: int = int(os.getenv("SERVICE_PORT", "8000"))
    api_title: str = "OrangePI Parcel Automat Service"
    api_version: str = "1.0.0"

    go_orchestrator_url: str = os.getenv(
        "GO_ORCHESTRATOR_URL", "http://localhost:8080/api/v1")

    cells_mapping_file: Path = Path(
        os.getenv("CELLS_MAPPING_FILE", "data/cells_mapping.json"))

    camera_device: int = int(os.getenv("CAMERA_INDEX", "0"))
    qr_scan_interval: float = float(os.getenv("QR_SCAN_INTERVAL", "0.1"))
    use_mock_scanner: bool = os.getenv(
        "USE_MOCK_SCANNER", "false").lower() == "true"

    scanner_stable_frames: int = int(os.getenv("SCANNER_STABLE_FRAMES", "3"))
    scanner_miss_frames: int = int(os.getenv("SCANNER_MISS_FRAMES", "5"))
    scanner_debounce_seconds: float = float(
        os.getenv("SCANNER_DEBOUNCE_SECONDS", "5"))

    arduino_port: str = os.getenv("ARDUINO_PORT", "/dev/ttyUSB0")
    arduino_baudrate: int = int(os.getenv("ARDUINO_BAUDRATE", "9600"))
    arduino_timeout: float = 1.0

    use_mock_arduino: bool = os.getenv(
        "USE_MOCK_ARDUINO", "false").lower() == "true"

    display_port: str = os.getenv("DISPLAY_PORT", "/dev/ttyUSB1")
    display_baudrate: int = int(os.getenv("DISPLAY_BAUDRATE", "115200"))

    log_level: str = os.getenv("LOG_LEVEL", "INFO")


settings = Settings()
