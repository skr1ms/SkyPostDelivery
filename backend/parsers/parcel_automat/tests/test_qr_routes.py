import pytest
from unittest.mock import Mock, AsyncMock, patch
from fastapi.testclient import TestClient
from httpx import HTTPError

from app.models.schemas import QRScanResponse


@pytest.fixture
def mock_qr_scanner():
    scanner = Mock()
    scanner.scan_once = Mock()
    return scanner


@pytest.fixture
def mock_qr_service():
    service = Mock()
    service.validate_qr_with_go = AsyncMock()
    service.confirm_pickup = AsyncMock()
    service.confirm_loaded = AsyncMock()
    return service


@pytest.fixture
def mock_cell_service():
    service = Mock()
    service.open_cells_by_uuids = Mock()
    return service


@pytest.fixture
def client(mock_qr_scanner, mock_qr_service, mock_cell_service):
    from main import app
    from app.api import qr_routes

    def override_get_qr_scanner():
        return mock_qr_scanner

    def override_get_qr_service():
        return mock_qr_service

    def override_get_cell_service():
        return mock_cell_service

    app.dependency_overrides[qr_routes.get_qr_scanner] = override_get_qr_scanner
    app.dependency_overrides[qr_routes.get_qr_service] = override_get_qr_service
    app.dependency_overrides[qr_routes.get_cell_service] = override_get_cell_service

    with TestClient(app) as test_client:
        yield test_client

    app.dependency_overrides.clear()


class TestQRScanEndpoint:

    def test_scan_qr_code_success(self, client, mock_qr_service, mock_cell_service):
        mock_qr_service.validate_qr_with_go.return_value = QRScanResponse(
            success=True,
            message="Valid QR code",
            cell_ids=["cell-1", "cell-2"]
        )
        mock_cell_service.open_cells_by_uuids.return_value = [
            {"cell_number": 1, "cell_uuid": "cell-1"},
            {"cell_number": 2, "cell_uuid": "cell-2"}
        ]

        response = client.post(
            "/api/qr/scan",
            json={"qr_data": '{"order_id": "test-123"}'}
        )

        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True
        assert data["message"] == "QR code validated and cells opened successfully"
        assert data["cell_count"] == 2
        assert len(data["cells_opened"]) == 2

    def test_scan_qr_code_validation_failure(self, client, mock_qr_service, mock_cell_service):
        mock_qr_service.validate_qr_with_go.return_value = QRScanResponse(
            success=False,
            message="Invalid QR code",
            cell_ids=[]
        )

        response = client.post(
            "/api/qr/scan",
            json={"qr_data": '{"invalid": "data"}'}
        )

        assert response.status_code == 200
        data = response.json()
        assert data["success"] is False
        assert data["message"] == "Invalid QR code"
        assert len(data["cells_opened"]) == 0

        mock_cell_service.open_cells_by_uuids.assert_not_called()

    def test_scan_qr_code_orchestrator_error(self, client, mock_qr_service):
        mock_qr_service.validate_qr_with_go.side_effect = HTTPError(
            "Connection failed")

        response = client.post(
            "/api/qr/scan",
            json={"qr_data": '{"order_id": "test-123"}'}
        )

        assert response.status_code == 500
        assert "Failed to process QR scan" in response.json()["detail"]


class TestConfirmPickupEndpoint:

    def test_confirm_pickup_success(self, client, mock_qr_service):
        mock_qr_service.confirm_pickup.return_value = {
            "status": "confirmed",
            "timestamp": "2025-11-01T12:00:00Z"
        }

        response = client.post(
            "/api/qr/confirm-pickup",
            json={"cell_ids": ["cell-1", "cell-2"]}
        )

        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True
        assert data["message"] == "Pickup confirmed successfully"
        assert "data" in data

    def test_confirm_pickup_error(self, client, mock_qr_service):
        mock_qr_service.confirm_pickup.side_effect = HTTPError("Network error")

        response = client.post(
            "/api/qr/confirm-pickup",
            json={"cell_ids": ["cell-1"]}
        )

        assert response.status_code == 500
        assert "Failed to confirm pickup" in response.json()["detail"]


class TestConfirmLoadedEndpoint:

    def test_confirm_loaded_success(self, client, mock_qr_service):
        mock_qr_service.confirm_loaded.return_value = {
            "status": "loaded",
            "delivery_id": "delivery-123"
        }

        response = client.post(
            "/api/qr/confirm-loaded",
            json={
                "order_id": "order-123",
                "locker_cell_id": "cell-456"
            }
        )

        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True
        assert data["message"] == "Load confirmed successfully"
        assert "data" in data

    def test_confirm_loaded_error(self, client, mock_qr_service):
        mock_qr_service.confirm_loaded.side_effect = HTTPError("Server error")

        response = client.post(
            "/api/qr/confirm-loaded",
            json={
                "order_id": "order-123",
                "locker_cell_id": "cell-456"
            }
        )

        assert response.status_code == 500
        assert "Failed to confirm load" in response.json()["detail"]


class TestScanFromCameraEndpoint:

    def test_scan_from_camera_success(self, client, mock_qr_scanner, mock_qr_service, mock_cell_service):
        test_qr_data = '{"order_id": "camera-test-123"}'
        mock_qr_scanner.scan_once.return_value = test_qr_data
        mock_qr_service.validate_qr_with_go.return_value = QRScanResponse(
            success=True,
            message="Valid QR from camera",
            cell_ids=["cell-1"]
        )
        mock_cell_service.open_cells_by_uuids.return_value = [
            {"cell_number": 1, "cell_uuid": "cell-1"}
        ]

        response = client.post("/api/qr/scan-from-camera")

        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True
        assert data["message"] == "QR code scanned and cells opened successfully"
        assert data["cell_count"] == 1
        assert "qr_data" in data

        mock_qr_scanner.scan_once.assert_called_once()

    def test_scan_from_camera_no_qr_detected(self, client, mock_qr_scanner):
        mock_qr_scanner.scan_once.return_value = None

        response = client.post("/api/qr/scan-from-camera")

        assert response.status_code == 200
        data = response.json()
        assert data["success"] is False
        assert "No QR code detected" in data["message"]
        assert len(data["cells_opened"]) == 0

    def test_scan_from_camera_validation_failure(self, client, mock_qr_scanner, mock_qr_service):
        test_qr_data = '{"invalid": "qr"}'
        mock_qr_scanner.scan_once.return_value = test_qr_data
        mock_qr_service.validate_qr_with_go.return_value = QRScanResponse(
            success=False,
            message="Invalid QR code format",
            cell_ids=[]
        )

        response = client.post("/api/qr/scan-from-camera")

        assert response.status_code == 200
        data = response.json()
        assert data["success"] is False
        assert data["message"] == "Invalid QR code format"
        assert "qr_data" in data

    def test_scan_from_camera_scanner_error(self, client, mock_qr_scanner):
        mock_qr_scanner.scan_once.side_effect = Exception("Camera malfunction")

        response = client.post("/api/qr/scan-from-camera")

        assert response.status_code == 500
        assert "Failed to scan QR from camera" in response.json()["detail"]
