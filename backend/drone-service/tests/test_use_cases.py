import pytest
import uuid
from datetime import datetime
from unittest.mock import AsyncMock
from app.services.use_cases import DeliveryUseCase
from app.models.models import DeliveryTask, GoodDimensions, DeliveryStatus


@pytest.fixture
def mock_state_repository():
    repo = AsyncMock()
    repo.save_delivery_task = AsyncMock(return_value=True)
    repo.update_delivery_status = AsyncMock(return_value=True)
    repo.get_delivery_task = AsyncMock(return_value=None)
    return repo


@pytest.fixture
def mock_drone_manager():
    manager = AsyncMock()
    manager.registered_drones = {}
    manager.register_drone = AsyncMock()
    manager.assign_delivery_to_drone = AsyncMock()
    manager.release_drone = AsyncMock()
    return manager


@pytest.fixture
def mock_drone_ws_handler():
    handler = AsyncMock()
    handler.send_task_to_drone = AsyncMock(return_value=True)
    handler.send_command_to_drone = AsyncMock()
    return handler


@pytest.fixture
def mock_orchestrator_grpc():
    client = AsyncMock()
    client.request_cell_open = AsyncMock(return_value={"success": True, "cell_id": "cell_123", "internal_cell_id": "internal_456"})
    return client


@pytest.fixture
def mock_rabbitmq_client():
    client = AsyncMock()
    client.publish = AsyncMock()
    return client


@pytest.fixture
def delivery_use_case(
    mock_state_repository,
    mock_drone_manager,
    mock_drone_ws_handler,
    mock_orchestrator_grpc,
    mock_rabbitmq_client
):
    return DeliveryUseCase(
        state_repository=mock_state_repository,
        drone_manager=mock_drone_manager,
        drone_ws_handler=mock_drone_ws_handler,
        orchestrator_grpc_client=mock_orchestrator_grpc,
        rabbitmq_client=mock_rabbitmq_client
    )


@pytest.mark.asyncio
async def test_execute_delivery_sends_correct_task_data(delivery_use_case, mock_drone_ws_handler):
    drone_id = str(uuid.uuid4())
    good_id = str(uuid.uuid4())
    parcel_automat_id = str(uuid.uuid4())
    
    await delivery_use_case.execute_delivery(
        drone_id=drone_id,
        good_id=good_id,
        parcel_automat_id=parcel_automat_id,
        aruco_id=42,
        weight=2.5,
        height=10.0,
        length=20.0,
        width=15.0
    )
    
    mock_drone_ws_handler.send_task_to_drone.assert_called_once()
    call_args = mock_drone_ws_handler.send_task_to_drone.call_args
    assert call_args[0][0] == drone_id
    
    task_data = call_args[0][1]
    assert task_data["good_id"] == good_id
    assert task_data["parcel_automat_id"] == parcel_automat_id
    assert task_data["aruco_id"] == 42
    assert task_data["dimensions"]["weight"] == 2.5
    assert task_data["dimensions"]["height"] == 10.0
    assert task_data["dimensions"]["length"] == 20.0
    assert task_data["dimensions"]["width"] == 15.0


@pytest.mark.asyncio
async def test_execute_delivery_updates_status_to_in_progress(delivery_use_case, mock_state_repository):
    await delivery_use_case.execute_delivery(
        drone_id=str(uuid.uuid4()),
        good_id=str(uuid.uuid4()),
        parcel_automat_id=str(uuid.uuid4()),
        aruco_id=42,
        weight=2.5,
        height=10.0,
        length=20.0,
        width=15.0
    )
    
    mock_state_repository.update_delivery_status.assert_called_once()
    call_args = mock_state_repository.update_delivery_status.call_args[0]
    assert call_args[1] == DeliveryStatus.IN_PROGRESS


@pytest.mark.asyncio
async def test_execute_delivery_handles_send_failure(
    delivery_use_case,
    mock_drone_ws_handler,
    mock_state_repository,
    mock_drone_manager
):
    mock_drone_ws_handler.send_task_to_drone.return_value = False
    drone_id = str(uuid.uuid4())
    
    await delivery_use_case.execute_delivery(
        drone_id=drone_id,
        good_id=str(uuid.uuid4()),
        parcel_automat_id=str(uuid.uuid4()),
        aruco_id=42,
        weight=2.5,
        height=10.0,
        length=20.0,
        width=15.0
    )
    
    status_calls = [call[0][1] for call in mock_state_repository.update_delivery_status.call_args_list]
    assert DeliveryStatus.FAILED in status_calls
    mock_drone_manager.release_drone.assert_called_once_with(drone_id)


@pytest.mark.asyncio
async def test_cancel_delivery_sends_command_and_updates_status(
    delivery_use_case,
    mock_state_repository,
    mock_drone_ws_handler,
    mock_drone_manager
):
    delivery_id = str(uuid.uuid4())
    drone_id = str(uuid.uuid4())
    
    task = DeliveryTask(
        delivery_id=delivery_id,
        order_id=delivery_id,
        good_id=str(uuid.uuid4()),
        locker_cell_id="cell_123",
        parcel_automat_id=str(uuid.uuid4()),
        dimensions=GoodDimensions(weight=2.5, height=10.0, length=20.0, width=15.0),
        created_at=datetime.now(),
        drone_id=drone_id
    )
    mock_state_repository.get_delivery_task.return_value = task
    
    result = await delivery_use_case.cancel_delivery(delivery_id)
    
    assert result["success"] is True
    mock_drone_ws_handler.send_command_to_drone.assert_called_once_with(
        drone_id,
        {"command": "cancel_delivery"}
    )
    mock_state_repository.update_delivery_status.assert_called_once_with(
        delivery_id,
        DeliveryStatus.CANCELLED
    )
    mock_drone_manager.release_drone.assert_called_once_with(drone_id)


@pytest.mark.asyncio
async def test_cancel_delivery_returns_error_when_not_found(delivery_use_case, mock_state_repository):
    mock_state_repository.get_delivery_task.return_value = None
    
    result = await delivery_use_case.cancel_delivery(str(uuid.uuid4()))
    
    assert result["success"] is False
    assert "not found" in result["message"].lower()


@pytest.mark.asyncio
async def test_get_delivery_status_returns_correct_data(delivery_use_case, mock_state_repository):
    delivery_id = str(uuid.uuid4())
    drone_id = str(uuid.uuid4())
    
    task = DeliveryTask(
        delivery_id=delivery_id,
        order_id=delivery_id,
        good_id=str(uuid.uuid4()),
        locker_cell_id="cell_123",
        parcel_automat_id=str(uuid.uuid4()),
        dimensions=GoodDimensions(weight=2.5, height=10.0, length=20.0, width=15.0),
        created_at=datetime.now(),
        drone_id=drone_id
    )
    task.status = DeliveryStatus.IN_PROGRESS
    mock_state_repository.get_delivery_task.return_value = task
    
    result = await delivery_use_case.get_delivery_status(delivery_id)
    
    assert result["success"] is True
    assert result["delivery_id"] == delivery_id
    assert result["status"] == DeliveryStatus.IN_PROGRESS.value
    assert result["drone_id"] == drone_id


@pytest.mark.asyncio
async def test_get_delivery_status_returns_error_when_not_found(delivery_use_case, mock_state_repository):
    mock_state_repository.get_delivery_task.return_value = None
    
    result = await delivery_use_case.get_delivery_status(str(uuid.uuid4()))
    
    assert result["success"] is False
    assert "not found" in result["message"].lower()


@pytest.mark.asyncio
async def test_handle_cargo_dropped_completes_delivery_and_publishes_confirmation(
    delivery_use_case,
    mock_state_repository,
    mock_drone_manager,
    mock_rabbitmq_client
):
    order_id = str(uuid.uuid4())
    drone_id = str(uuid.uuid4())
    locker_cell_id = "cell_123"
    
    task = DeliveryTask(
        delivery_id=order_id,
        order_id=order_id,
        good_id=str(uuid.uuid4()),
        locker_cell_id=locker_cell_id,
        parcel_automat_id=str(uuid.uuid4()),
        dimensions=GoodDimensions(weight=2.5, height=10.0, length=20.0, width=15.0),
        created_at=datetime.now(),
        drone_id=drone_id
    )
    mock_state_repository.get_delivery_task.return_value = task
    
    result = await delivery_use_case.handle_cargo_dropped(order_id, locker_cell_id)
    
    if not result["success"]:
        print(f"Error: {result.get('message')}")
    
    assert result["success"] is True
    mock_state_repository.update_delivery_status.assert_called_once_with(
        order_id,
        DeliveryStatus.COMPLETED
    )
    assert mock_rabbitmq_client.publish.called


@pytest.mark.asyncio
async def test_handle_drone_arrived_requests_cell_and_sends_drop_command(
    delivery_use_case,
    mock_orchestrator_grpc,
    mock_drone_ws_handler
):
    drone_id = str(uuid.uuid4())
    order_id = str(uuid.uuid4())
    parcel_automat_id = str(uuid.uuid4())
    
    result = await delivery_use_case.handle_drone_arrived(
        drone_id,
        order_id,
        parcel_automat_id
    )
    
    assert result["success"] is True
    assert result["cell_id"] == "cell_123"
    mock_orchestrator_grpc.request_cell_open.assert_called_once_with(
        order_id=order_id,
        parcel_automat_id=parcel_automat_id
    )
    mock_drone_ws_handler.send_command_to_drone.assert_called_once_with(
        drone_id,
        {"command": "drop_cargo", "order_id": order_id, "cell_id": "cell_123", "internal_cell_id": "internal_456"}
    )


@pytest.mark.asyncio
async def test_handle_drone_arrived_returns_error_when_no_orchestrator(delivery_use_case):
    delivery_use_case.orchestrator_grpc_client = None
    
    result = await delivery_use_case.handle_drone_arrived(
        str(uuid.uuid4()),
        str(uuid.uuid4()),
        str(uuid.uuid4())
    )
    
    assert result["success"] is False
    assert "not configured" in result["message"].lower()


@pytest.mark.asyncio
async def test_handle_drone_arrived_returns_error_when_cell_open_fails(
    delivery_use_case,
    mock_orchestrator_grpc
):
    mock_orchestrator_grpc.request_cell_open.return_value = {
        "success": False,
        "message": "Cell unavailable"
    }
    
    result = await delivery_use_case.handle_drone_arrived(
        str(uuid.uuid4()),
        str(uuid.uuid4()),
        str(uuid.uuid4())
    )
    
    assert result["success"] is False
    assert "failed to open cell" in result["message"].lower()
