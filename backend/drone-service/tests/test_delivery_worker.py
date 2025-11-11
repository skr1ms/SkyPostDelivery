import pytest
import uuid
from unittest.mock import AsyncMock
from app.services.workers.delivery import DeliveryWorker


@pytest.fixture
def mock_rabbitmq_client():
    client = AsyncMock()
    client.connect = AsyncMock()
    client.consume = AsyncMock()
    return client


@pytest.fixture
def mock_delivery_use_case():
    use_case = AsyncMock()
    use_case.execute_delivery = AsyncMock()
    return use_case


@pytest.fixture
def delivery_worker(mock_rabbitmq_client, mock_delivery_use_case):
    return DeliveryWorker(mock_rabbitmq_client, mock_delivery_use_case)


@pytest.mark.asyncio
async def test_start_worker(delivery_worker, mock_rabbitmq_client):
    await delivery_worker.start()
    
    assert delivery_worker._running is True
    mock_rabbitmq_client.connect.assert_called_once()
    assert mock_rabbitmq_client.consume.call_count == 2


@pytest.mark.asyncio
async def test_handle_delivery_task_success(delivery_worker, mock_delivery_use_case):
    message = {
        "drone_id": str(uuid.uuid4()),
        "drone_ip": "192.168.1.100",
        "good_id": str(uuid.uuid4()),
        "parcel_automat_id": str(uuid.uuid4()),
        "aruco_id": 42,
        "weight": 2.5,
        "height": 10.0,
        "length": 20.0,
        "width": 15.0,
        "priority": 5,
        "created_at": 1234567890
    }
    
    await delivery_worker.handle_delivery_task(message)
    
    mock_delivery_use_case.execute_delivery.assert_called_once()
    call_kwargs = mock_delivery_use_case.execute_delivery.call_args[1]
    assert "drone_id" in call_kwargs
    assert "good_id" in call_kwargs
    assert "aruco_id" in call_kwargs
    assert call_kwargs["aruco_id"] == 42


@pytest.mark.asyncio
async def test_handle_delivery_task_missing_priority(delivery_worker, mock_delivery_use_case):
    message = {
        "drone_id": str(uuid.uuid4()),
        "good_id": str(uuid.uuid4()),
        "parcel_automat_id": str(uuid.uuid4()),
        "aruco_id": 42,
        "weight": 2.5,
        "height": 10.0,
        "length": 20.0,
        "width": 15.0
    }
    
    await delivery_worker.handle_delivery_task(message)
    
    mock_delivery_use_case.execute_delivery.assert_called_once()


@pytest.mark.asyncio
async def test_handle_delivery_task_invalid_uuid(delivery_worker, mock_delivery_use_case):
    message = {
        "drone_id": "invalid-uuid",
        "good_id": str(uuid.uuid4()),
        "parcel_automat_id": str(uuid.uuid4()),
        "aruco_id": 42,
        "weight": 2.5,
        "height": 10.0,
        "length": 20.0,
        "width": 15.0
    }
    
    await delivery_worker.handle_delivery_task(message)
    
    mock_delivery_use_case.execute_delivery.assert_not_called()


@pytest.mark.asyncio
async def test_stop_worker(delivery_worker):
    delivery_worker._running = True
    
    await delivery_worker.stop()
    
    assert delivery_worker._running is False
