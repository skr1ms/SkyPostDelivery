from .repositories.cell_mapping_repository import CellMappingRepository
from .services.cell_management_service import CellManagementService
from .services.qr_scan_service import QRScanService, QRScannerWorker
from .hardware.arduino_controller import ArduinoController
from .hardware.qr_scanner import QRScanner
from .hardware.display_controller import DisplayController
from config.config import settings
import logging

logger = logging.getLogger(__name__)

cell_repo = CellMappingRepository(mapping_file=settings.cells_mapping_file)

arduino_controller = ArduinoController(
    port=settings.arduino_port,
    baudrate=settings.arduino_baudrate,
    timeout=settings.arduino_timeout,
    mock_mode=settings.use_mock_arduino
)

qr_scanner = QRScanner(
    camera_device=settings.camera_device,
    mock_mode=settings.use_mock_scanner
)

display = DisplayController(
    port=settings.display_port,
    baudrate=settings.display_baudrate
)

cell_service = CellManagementService(
    cell_repo=cell_repo,
    arduino_controller=arduino_controller,
    display=display
)

qr_service = QRScanService(orchestrator_url=settings.go_orchestrator_url)

scanner_worker = QRScannerWorker(
    qr_scanner=qr_scanner,
    qr_service=qr_service,
    cell_service=cell_service,
    display=display,
    scan_interval=settings.qr_scan_interval,
    stable_frames=settings.scanner_stable_frames,
    miss_frames=settings.scanner_miss_frames,
    debounce_seconds=settings.scanner_debounce_seconds,
)


def get_cell_service():
    return cell_service


def get_scanner_worker():
    return scanner_worker


async def cleanup():
    await scanner_worker.stop()
    await qr_service.close()
    arduino_controller.close()
    qr_scanner.close()
    display.close()
