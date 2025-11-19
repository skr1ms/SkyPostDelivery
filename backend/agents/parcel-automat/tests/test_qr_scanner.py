import asyncio
import json
import time
from datetime import datetime, timedelta
from types import SimpleNamespace
from unittest.mock import AsyncMock, Mock, patch

import pytest

import app.hardware.qr_scanner as qr_module
from app.hardware.qr_scanner import QRScanner
from app.services.qr_scan_service import QRScanService, QRScannerWorker


def _build_valid_qr_payload():
    return json.dumps(
        {
            "user_id": "00000000-0000-0000-0000-000000000001",
            "email": "user@example.com",
            "expires_at": (datetime.utcnow() + timedelta(minutes=5)).isoformat() + "Z",
            "signature": "valid_signature_value",
        }
    )


@pytest.fixture
def mock_camera():
    cam = Mock()
    cam.isOpened.return_value = True
    cam.read.return_value = (True, SimpleNamespace(shape=(480, 640, 3)))
    return cam


class TestQRScanner:
    def test_initialization_mock_mode_when_no_camera(self):
        failing_camera = Mock()
        failing_camera.isOpened.return_value = False

        with patch("app.hardware.qr_scanner.cv2.VideoCapture", return_value=failing_camera):
            scanner = QRScanner(camera_device=0)

        assert scanner._mock_mode is True
        assert scanner.camera is None

    def test_initialization_success(self, mock_camera):
        with patch("app.hardware.qr_scanner.cv2.VideoCapture", return_value=mock_camera) as mock_capture:
            scanner = QRScanner(camera_device=1)

        mock_capture.assert_called_once_with(1, qr_module.cv2.CAP_V4L2)
        assert scanner._mock_mode is False
        assert scanner.camera is mock_camera

    def test_scan_once_in_mock_mode(self):
        failing_camera = Mock()
        failing_camera.isOpened.return_value = False

        with patch("app.hardware.qr_scanner.cv2.VideoCapture", return_value=failing_camera):
            scanner = QRScanner(camera_device=0)

        assert scanner._mock_mode is True
        assert scanner.scan_once() is None

    def test_scan_once_no_qr_detected(self, mock_camera):
        with patch("app.hardware.qr_scanner.cv2.VideoCapture", return_value=mock_camera):
            scanner = QRScanner(camera_device=0)

        with patch.object(scanner, "read_frame", return_value=Mock()), patch(
            "app.hardware.qr_scanner.pyzbar.decode", return_value=[]
        ):
            assert scanner.scan_once() is None

    def test_scan_once_qr_detected(self, mock_camera):
        valid_qr = _build_valid_qr_payload()
        decoded = SimpleNamespace(data=valid_qr.encode("utf-8"))

        with patch("app.hardware.qr_scanner.cv2.VideoCapture", return_value=mock_camera):
            scanner = QRScanner(camera_device=0, stable_frames=1)

        with patch.object(scanner, "read_frame", return_value=Mock()), patch(
            "app.hardware.qr_scanner.pyzbar.decode", return_value=[decoded]
        ):
            assert scanner.scan_once() == valid_qr

    def test_scan_once_camera_read_fails(self, mock_camera):
        with patch("app.hardware.qr_scanner.cv2.VideoCapture", return_value=mock_camera):
            scanner = QRScanner(camera_device=0)

        with patch.object(scanner, "read_frame", return_value=None):
            assert scanner.scan_once() is None

    def test_scan_once_exception_handling(self, mock_camera):
        with patch("app.hardware.qr_scanner.cv2.VideoCapture", return_value=mock_camera):
            scanner = QRScanner(camera_device=0)

        with patch.object(scanner, "read_frame", side_effect=Exception("Camera error")):
            assert scanner.scan_once() is None

    @pytest.mark.asyncio
    async def test_scan_continuous(self, mock_camera):
        callback_called = asyncio.Event()
        received = {}

        async def callback(qr_data):
            received["data"] = qr_data
            callback_called.set()

        with patch("app.hardware.qr_scanner.cv2.VideoCapture", return_value=mock_camera):
            scanner = QRScanner(camera_device=0)

        with patch.object(scanner, "scan_once", side_effect=["data", None, None]), patch(
            "app.hardware.qr_scanner.asyncio.sleep", new=AsyncMock()
        ):
            task = asyncio.create_task(scanner.scan_continuous(interval=0.01, callback=callback))
            await asyncio.wait_for(callback_called.wait(), timeout=0.1)
            task.cancel()
            with pytest.raises(asyncio.CancelledError):
                await task

        assert received["data"] == "data"

    def test_close_in_normal_mode(self, mock_camera):
        with patch("app.hardware.qr_scanner.cv2.VideoCapture", return_value=mock_camera):
            scanner = QRScanner(camera_device=0)

        scanner.close()
        mock_camera.release.assert_called_once()

    def test_close_in_mock_mode(self):
        failing_camera = Mock()
        failing_camera.isOpened.return_value = False

        with patch("app.hardware.qr_scanner.cv2.VideoCapture", return_value=failing_camera):
            scanner = QRScanner(camera_device=0)

        scanner.close()  # Should not raise


class TestQRScannerWorker:
    @pytest.mark.asyncio
    async def test_enqueue_and_get_mock_qr(self):
        scanner = Mock()
        scanner.mock_mode = True

        worker = QRScannerWorker(
            qr_scanner=scanner,
            qr_service=Mock(spec=QRScanService),
            cell_service=Mock(),
            display=None,
        )
        worker._mock_queue = asyncio.Queue()

        worker.enqueue_mock_qr("mock-data")
        assert await worker._get_next_mock_qr() == "mock-data"

    def test_should_process_qr_debounce(self):
        scanner = Mock()
        scanner.mock_mode = False
        worker = QRScannerWorker(
            qr_scanner=scanner,
            qr_service=Mock(spec=QRScanService),
            cell_service=Mock(),
            display=None,
            debounce_seconds=5,
        )

        now = time.time()
        worker.active_qrs.add("qr")
        worker.last_seen["qr"] = now - 10

        assert worker._should_process_qr("qr", now) is True
        worker.last_seen["qr"] = now
        assert worker._should_process_qr("qr", now) is False

    @pytest.mark.asyncio
    async def test_handle_confirmed_qr_success(self):
        scanner = Mock()
        scanner.mock_mode = False

        qr_service = Mock(spec=QRScanService)
        qr_service.validate_qr_with_go = AsyncMock(
            return_value=SimpleNamespace(success=True, cell_ids=["id-1"])
        )

        cell_service = Mock()
        cell_service.get_parcel_automat_id.return_value = "automat"
        cell_service.open_cells_by_uuids.return_value = [{"success": True}]

        display = Mock()

        worker = QRScannerWorker(
            qr_scanner=scanner,
            qr_service=qr_service,
            cell_service=cell_service,
            display=display,
        )
        worker.qr_to_id["payload"] = 1

        with patch("app.services.qr_scan_service.asyncio.sleep", new=AsyncMock()):
            await worker._handle_confirmed_qr("payload", 1)

        display.show_scanning.assert_called_once()
        display.show_qr_success.assert_called_once()
        display.show_welcome.assert_called()
