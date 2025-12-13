import pytest
from unittest.mock import AsyncMock, MagicMock
from app.services.delivery_service import DeliveryService
from app.models.task import DeliveryTask


@pytest.fixture
def mock_websocket_client():
    client = AsyncMock()
    client.send_delivery_update = AsyncMock()
    client.send_status_update = AsyncMock()
    return client


@pytest.fixture
def mock_state_machine():
    return MagicMock()


@pytest.fixture
def mock_flight_controller():
    controller = MagicMock()
    controller.launch_delivery_flight = MagicMock(return_value=True)
    return controller


@pytest.fixture
def mock_ros_bridge():
    return MagicMock()


@pytest.fixture
def delivery_service(
    mock_websocket_client, mock_state_machine, mock_flight_controller, mock_ros_bridge
):
    return DeliveryService(
        mock_websocket_client,
        mock_state_machine,
        mock_flight_controller,
        mock_ros_bridge,
    )


@pytest.mark.asyncio
async def test_handle_delivery_task(
    delivery_service, mock_flight_controller, mock_state_machine
):
    payload = {
        "delivery_id": "test-123",
        "order_id": "order-456",
        "good_id": "good-789",
        "parcel_automat_id": "automat-1",
        "target_aruco_id": 135,
        "home_aruco_id": 131,
    }

    await delivery_service.handle_delivery_task(payload)

    mock_state_machine.set_task.assert_called_once()
    mock_flight_controller.launch_delivery_flight.assert_called_once_with(
        target_aruco_id=135, home_aruco_id=131
    )


@pytest.mark.asyncio
async def test_on_arrival_sends_update(
    delivery_service, mock_websocket_client, mock_state_machine
):
    task = DeliveryTask(
        delivery_id="test-123",
        order_id="order-456",
        good_id="good-789",
        parcel_automat_id="automat-1",
        target_aruco_id=135,
        home_aruco_id=131,
    )
    mock_state_machine.current_task = task
    mock_state_machine.is_event_sent = MagicMock(return_value=False)

    await delivery_service.on_arrival()

    mock_websocket_client.send_delivery_update.assert_called_once()
    mock_state_machine.mark_event_sent.assert_called_once()


@pytest.mark.asyncio
async def test_on_arrival_idempotency(
    delivery_service, mock_websocket_client, mock_state_machine
):
    task = DeliveryTask(
        delivery_id="test-123",
        order_id="order-456",
        good_id="good-789",
        parcel_automat_id="automat-1",
        target_aruco_id=135,
        home_aruco_id=131,
    )
    mock_state_machine.current_task = task
    mock_state_machine.is_event_sent = MagicMock(return_value=True)

    await delivery_service.on_arrival()

    mock_websocket_client.send_delivery_update.assert_not_called()


@pytest.mark.asyncio
async def test_on_arrival_no_task(delivery_service, mock_state_machine):
    mock_state_machine.current_task = None

    await delivery_service.on_arrival()

    mock_state_machine.transition_to.assert_not_called()


@pytest.mark.asyncio
async def test_on_drop_ready(delivery_service, mock_state_machine):
    task = DeliveryTask(
        delivery_id="test-123",
        order_id="order-456",
        good_id="good-789",
        parcel_automat_id="automat-1",
        target_aruco_id=135,
        home_aruco_id=131,
    )
    mock_state_machine.current_task = task

    await delivery_service.on_drop_ready()

    mock_state_machine.transition_to.assert_called_once()


@pytest.mark.asyncio
async def test_on_home_arrived(
    delivery_service, mock_websocket_client, mock_state_machine
):
    task = DeliveryTask(
        delivery_id="test-123",
        order_id="order-456",
        good_id="good-789",
        parcel_automat_id="automat-1",
        target_aruco_id=135,
        home_aruco_id=131,
    )
    mock_state_machine.current_task = task

    await delivery_service.on_home_arrived()

    mock_websocket_client.send_status_update.assert_called_once()
    mock_state_machine.clear_task.assert_called_once()


@pytest.mark.asyncio
async def test_handle_drop_command(delivery_service, mock_state_machine, mock_ros_bridge):
    task = DeliveryTask(
        delivery_id="test-123",
        order_id="order-456",
        good_id="good-789",
        parcel_automat_id="automat-1",
        target_aruco_id=135,
        home_aruco_id=131,
    )
    mock_state_machine.current_task = task
    mock_ros_bridge.send_drop_confirmation = MagicMock(return_value=True)

    payload = {
        "order_id": "order-456",
        "cell_id": "cell-1",
        "internal_cell_id": "internal-1",
    }

    await delivery_service.handle_drop_command(payload)

    mock_state_machine.transition_to.assert_called_once()
    mock_ros_bridge.send_drop_confirmation.assert_called_once()


@pytest.mark.asyncio
async def test_handle_drop_command_no_task(delivery_service, mock_state_machine):
    mock_state_machine.current_task = None

    payload = {
        "order_id": "order-456",
        "cell_id": "cell-1",
    }

    await delivery_service.handle_drop_command(payload)

    mock_state_machine.transition_to.assert_not_called()


@pytest.mark.asyncio
async def test_handle_delivery_task_with_default_aruco():
    mock_ws = AsyncMock()
    mock_sm = MagicMock()
    mock_fc = MagicMock()
    mock_fc.launch_delivery_flight = MagicMock(return_value=True)
    mock_rb = MagicMock()

    service = DeliveryService(mock_ws, mock_sm, mock_fc, mock_rb)

    payload = {
        "delivery_id": "test-123",
        "order_id": "order-456",
        "good_id": "good-789",
        "parcel_automat_id": "automat-1",
    }

    await service.handle_delivery_task(payload)

    task_arg = mock_sm.set_task.call_args[0][0]
    assert task_arg.target_aruco_id == 135
    assert task_arg.home_aruco_id == 131


@pytest.mark.asyncio
async def test_handle_delivery_task_flight_failure():
    mock_ws = AsyncMock()
    mock_sm = MagicMock()
    mock_fc = MagicMock()
    mock_fc.launch_delivery_flight = MagicMock(return_value=False)
    mock_rb = MagicMock()

    service = DeliveryService(mock_ws, mock_sm, mock_fc, mock_rb)

    payload = {
        "delivery_id": "test-123",
        "order_id": "order-456",
        "good_id": "good-789",
        "parcel_automat_id": "automat-1",
        "target_aruco_id": 136,
        "home_aruco_id": 131,
    }

    await service.handle_delivery_task(payload)

    mock_sm.transition_to.assert_called()
    from app.models.task import DeliveryState
    assert mock_sm.transition_to.call_args[0][0] == DeliveryState.FAILED
