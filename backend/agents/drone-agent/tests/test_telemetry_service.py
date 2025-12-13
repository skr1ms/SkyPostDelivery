import pytest
from unittest.mock import AsyncMock, MagicMock, patch
from app.services.telemetry_service import TelemetryService
from app.models.telemetry import Telemetry, Battery, Pose, MavrosState


@pytest.fixture
def mock_websocket_client():
    client = AsyncMock()
    client.is_connected = True
    client.settings = MagicMock()
    client.settings.drone_id = "drone-123"
    client.send_heartbeat = AsyncMock()
    return client


@pytest.fixture
def mock_ros_bridge():
    bridge = MagicMock()
    telemetry = Telemetry()
    telemetry.battery = Battery(voltage=12.5, percentage=85.0, current=2.5)
    telemetry.pose = Pose(x=1.0, y=2.0, z=3.0, orientation={})
    telemetry.state = MavrosState(armed=True, connected=True, mode="OFFBOARD")
    bridge.get_telemetry = MagicMock(return_value=telemetry)
    return bridge


@pytest.fixture
def telemetry_service(mock_websocket_client, mock_ros_bridge):
    return TelemetryService(mock_websocket_client, mock_ros_bridge, interval=1)


@pytest.mark.asyncio
async def test_telemetry_service_start():
    ws_client = AsyncMock()
    ros_bridge = MagicMock()
    service = TelemetryService(ws_client, ros_bridge, interval=1)
    
    await service.start()
    
    assert service._running is True
    assert service._task is not None
    
    await service.stop()


@pytest.mark.asyncio
async def test_telemetry_service_already_running():
    ws_client = AsyncMock()
    ros_bridge = MagicMock()
    service = TelemetryService(ws_client, ros_bridge, interval=1)
    
    await service.start()
    await service.start()
    
    assert service._running is True
    
    await service.stop()


@pytest.mark.asyncio
async def test_telemetry_service_stop():
    ws_client = AsyncMock()
    ros_bridge = MagicMock()
    service = TelemetryService(ws_client, ros_bridge, interval=1)
    
    await service.start()
    await service.stop()
    
    assert service._running is False


@pytest.mark.asyncio
async def test_heartbeat_sends_telemetry(mock_websocket_client, mock_ros_bridge):
    service = TelemetryService(mock_websocket_client, mock_ros_bridge, interval=0.1)
    
    await service.start()
    await asyncio.sleep(0.2)
    await service.stop()
    
    mock_websocket_client.send_heartbeat.assert_called()
    call_args = mock_websocket_client.send_heartbeat.call_args
    assert call_args.kwargs["battery_level"] == 12.5
    assert call_args.kwargs["status"] == "flying"


@pytest.mark.asyncio
async def test_heartbeat_skips_when_not_connected():
    ws_client = AsyncMock()
    ws_client.is_connected = False
    ws_client.settings = MagicMock()
    ws_client.settings.drone_id = "drone-123"
    ws_client.send_heartbeat = AsyncMock()
    
    ros_bridge = MagicMock()
    telemetry = Telemetry()
    ros_bridge.get_telemetry = MagicMock(return_value=telemetry)
    
    service = TelemetryService(ws_client, ros_bridge, interval=0.1)
    
    await service.start()
    await asyncio.sleep(0.2)
    await service.stop()
    
    ws_client.send_heartbeat.assert_not_called()


@pytest.mark.asyncio
async def test_heartbeat_skips_when_no_drone_id():
    ws_client = AsyncMock()
    ws_client.is_connected = True
    ws_client.settings = MagicMock()
    ws_client.settings.drone_id = None
    ws_client.send_heartbeat = AsyncMock()
    
    ros_bridge = MagicMock()
    telemetry = Telemetry()
    ros_bridge.get_telemetry = MagicMock(return_value=telemetry)
    
    service = TelemetryService(ws_client, ros_bridge, interval=0.1)
    
    await service.start()
    await asyncio.sleep(0.2)
    await service.stop()
    
    ws_client.send_heartbeat.assert_not_called()


@pytest.mark.asyncio
async def test_heartbeat_handles_errors():
    ws_client = AsyncMock()
    ws_client.is_connected = True
    ws_client.settings = MagicMock()
    ws_client.settings.drone_id = "drone-123"
    ws_client.send_heartbeat = AsyncMock(side_effect=Exception("Network error"))
    
    ros_bridge = MagicMock()
    telemetry = Telemetry()
    ros_bridge.get_telemetry = MagicMock(return_value=telemetry)
    
    service = TelemetryService(ws_client, ros_bridge, interval=0.1)
    
    await service.start()
    await asyncio.sleep(0.2)
    await service.stop()
    
    assert service._running is False


@pytest.mark.asyncio
async def test_heartbeat_with_no_battery():
    ws_client = AsyncMock()
    ws_client.is_connected = True
    ws_client.settings = MagicMock()
    ws_client.settings.drone_id = "drone-123"
    ws_client.send_heartbeat = AsyncMock()
    
    ros_bridge = MagicMock()
    telemetry = Telemetry()
    telemetry.battery = None
    ros_bridge.get_telemetry = MagicMock(return_value=telemetry)
    
    service = TelemetryService(ws_client, ros_bridge, interval=0.1)
    
    await service.start()
    await asyncio.sleep(0.2)
    await service.stop()
    
    call_args = ws_client.send_heartbeat.call_args
    assert call_args.kwargs["battery_level"] == 100.0


import asyncio
