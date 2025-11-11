import pytest
from unittest.mock import AsyncMock, MagicMock, patch
from app.services.delivery_service import DeliveryService
from app.models.schemas import (
    DeliveryTaskPayload,
    GoodDimensions,
    DroneStatus,
)


@pytest.fixture
def delivery_service():
    service = DeliveryService()
    service._status_callback = AsyncMock()
    service._delivery_update_callback = AsyncMock()
    return service


@pytest.mark.asyncio
async def test_initial_status(delivery_service):
    status = delivery_service.get_current_status()
    assert status.status == DroneStatus.IDLE
    assert status.battery_level == 100.0
    assert status.current_delivery_id is None


@pytest.mark.asyncio
async def test_initialize_without_map_file(delivery_service):
    with patch('app.services.delivery_service.Path') as mock_path:
        mock_path.return_value.exists.return_value = False
        result = await delivery_service.initialize()
        assert result == False


@pytest.mark.asyncio
async def test_execute_delivery_with_mock_nav():
    service = DeliveryService()
    service._status_callback = AsyncMock()
    service._delivery_update_callback = AsyncMock()
    
    mock_nav = MagicMock()
    mock_nav.execute_delivery.return_value = True
    mock_nav.api.get_battery.return_value = 95.0
    service.nav_controller = mock_nav
    service.is_initialized = True
    
    task = DeliveryTaskPayload(
        delivery_id="test_001",
        good_id="good_123",
        parcel_automat_id="automat_456",
        aruco_id=52,
        dimensions=GoodDimensions(weight=1.5, height=10.0, length=20.0, width=15.0)
    )
    
    await service.execute_delivery(task)
    
    assert service._status_callback.call_count >= 1
    assert service._delivery_update_callback.call_count >= 1
    mock_nav.execute_delivery.assert_called_once_with(52)


@pytest.mark.asyncio
async def test_execute_delivery_failure():
    service = DeliveryService()
    service._status_callback = AsyncMock()
    service._delivery_update_callback = AsyncMock()
    
    mock_nav = MagicMock()
    mock_nav.execute_delivery.return_value = False
    mock_nav.api.get_battery.return_value = 90.0
    service.nav_controller = mock_nav
    service.is_initialized = True
    
    task = DeliveryTaskPayload(
        delivery_id="test_002",
        good_id="good_123",
        parcel_automat_id="automat_456",
        aruco_id=52,
        dimensions=GoodDimensions(weight=1.5, height=10.0, length=20.0, width=15.0)
    )
    
    await service.execute_delivery(task)
    
    assert service.current_delivery_id == "test_002"
