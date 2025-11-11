import pytest
from unittest.mock import Mock, AsyncMock, patch
import httpx
from app.services.qr_scan_service import QRScanService
from app.models.schemas import QRScanResponse


class TestQRScanService:
    @pytest.fixture
    def qr_service(self):
        return QRScanService(orchestrator_url="http://test-orchestrator:8080/api/v1")

    @pytest.mark.asyncio
    async def test_validate_qr_with_go_success(self, qr_service):
        test_qr_data = '{"order_id": "test-123"}'
        test_parcel_automat_id = "automat-uuid-1"
        expected_response = {
            "success": True,
            "message": "QR code validated",
            "cell_ids": ["cell-uuid-1", "cell-uuid-2"]
        }

        with patch.object(qr_service.client, 'post', new_callable=AsyncMock) as mock_post:
            mock_response = Mock()
            mock_response.json.return_value = expected_response
            mock_post.return_value = mock_response

            result = await qr_service.validate_qr_with_go(test_qr_data, test_parcel_automat_id)

            assert isinstance(result, QRScanResponse)
            assert result.success is True
            assert result.message == "QR code validated"
            assert len(result.cell_ids) == 2
            assert result.cell_ids == ["cell-uuid-1", "cell-uuid-2"]

            mock_post.assert_called_once_with(
                "http://test-orchestrator:8080/api/v1/automats/qr-scan",
                json={"qr_data": test_qr_data,
                      "parcel_automat_id": test_parcel_automat_id}
            )

    @pytest.mark.asyncio
    async def test_validate_qr_with_go_failure(self, qr_service):
        test_qr_data = '{"invalid": "data"}'
        test_parcel_automat_id = "automat-uuid-1"
        expected_response = {
            "success": False,
            "message": "Invalid QR code",
            "cell_ids": []
        }

        with patch.object(qr_service.client, 'post', new_callable=AsyncMock) as mock_post:
            mock_response = Mock()
            mock_response.json.return_value = expected_response
            mock_post.return_value = mock_response

            result = await qr_service.validate_qr_with_go(test_qr_data, test_parcel_automat_id)

            assert result.success is False
            assert result.message == "Invalid QR code"
            assert len(result.cell_ids) == 0

    @pytest.mark.asyncio
    async def test_validate_qr_http_error(self, qr_service):
        test_qr_data = '{"order_id": "test-456"}'
        test_parcel_automat_id = "automat-uuid-1"

        with patch.object(qr_service.client, 'post', new_callable=AsyncMock) as mock_post:
            mock_post.side_effect = httpx.HTTPError("Connection failed")

            with pytest.raises(httpx.HTTPError):
                await qr_service.validate_qr_with_go(test_qr_data, test_parcel_automat_id)

    @pytest.mark.asyncio
    async def test_confirm_pickup_success(self, qr_service):
        cell_ids = ["cell-uuid-1", "cell-uuid-2"]
        expected_response = {
            "success": True,
            "message": "Pickup confirmed"
        }

        with patch.object(qr_service.client, 'post', new_callable=AsyncMock) as mock_post:
            mock_response = Mock()
            mock_response.json.return_value = expected_response
            mock_post.return_value = mock_response

            result = await qr_service.confirm_pickup(cell_ids)

            assert result["success"] is True
            assert result["message"] == "Pickup confirmed"

            mock_post.assert_called_once_with(
                "http://test-orchestrator:8080/api/v1/automats/confirm-pickup",
                json={"cell_ids": cell_ids}
            )

    @pytest.mark.asyncio
    async def test_confirm_pickup_http_error(self, qr_service):
        cell_ids = ["cell-uuid-1"]

        with patch.object(qr_service.client, 'post', new_callable=AsyncMock) as mock_post:
            mock_post.side_effect = httpx.HTTPError("Network error")

            with pytest.raises(httpx.HTTPError):
                await qr_service.confirm_pickup(cell_ids)

    @pytest.mark.asyncio
    async def test_confirm_loaded_success(self, qr_service):
        order_id = "order-uuid-123"
        locker_cell_id = "cell-uuid-456"
        expected_response = {
            "success": True,
            "message": "Load confirmed"
        }

        with patch.object(qr_service.client, 'post', new_callable=AsyncMock) as mock_post:
            mock_response = Mock()
            mock_response.json.return_value = expected_response
            mock_post.return_value = mock_response

            result = await qr_service.confirm_loaded(order_id, locker_cell_id)

            assert result["success"] is True
            assert result["message"] == "Load confirmed"

            mock_post.assert_called_once()
            call_args = mock_post.call_args
            assert "http://test-orchestrator:8080/api/v1/automats/confirm-loaded" in call_args[0]
            assert call_args[1]["json"]["order_id"] == order_id
            assert call_args[1]["json"]["locker_cell_id"] == locker_cell_id

    @pytest.mark.asyncio
    async def test_confirm_loaded_http_error(self, qr_service):
        order_id = "order-uuid-123"
        locker_cell_id = "cell-uuid-456"

        with patch.object(qr_service.client, 'post', new_callable=AsyncMock) as mock_post:
            mock_post.side_effect = httpx.HTTPError("Server error")

            with pytest.raises(httpx.HTTPError):
                await qr_service.confirm_loaded(order_id, locker_cell_id)

    @pytest.mark.asyncio
    async def test_close(self, qr_service):
        with patch.object(qr_service.client, 'aclose', new_callable=AsyncMock) as mock_close:
            await qr_service.close()
            mock_close.assert_called_once()

    @pytest.mark.asyncio
    async def test_default_orchestrator_url(self):
        with patch('app.services.qr_scan_service.settings') as mock_settings:
            mock_settings.go_orchestrator_url = "http://default:9000/api"

            service = QRScanService()
            assert service.orchestrator_url == "http://default:9000/api"

    @pytest.mark.asyncio
    async def test_custom_orchestrator_url(self):
        custom_url = "http://custom:7000/api/v2"
        service = QRScanService(orchestrator_url=custom_url)

        assert service.orchestrator_url == custom_url
