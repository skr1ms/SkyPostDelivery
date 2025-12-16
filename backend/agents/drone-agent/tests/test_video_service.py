import pytest
from unittest.mock import AsyncMock, MagicMock
from app.services.video_service import VideoService


@pytest.fixture
def mock_websocket_client():
    client = AsyncMock()
    client.is_connected = True
    client.settings = MagicMock()
    client.settings.drone_id = "drone-123"
    client.send_video_frame = AsyncMock()
    return client


@pytest.fixture
def mock_camera_handler():
    handler = MagicMock()
    handler.get_latest_frame = MagicMock(return_value="base64encodedframe")
    return handler


@pytest.fixture
def video_service(mock_websocket_client, mock_camera_handler):
    return VideoService(mock_websocket_client, mock_camera_handler, fps=10)


@pytest.mark.asyncio
async def test_video_service_start():
    ws_client = AsyncMock()
    camera = MagicMock()
    service = VideoService(ws_client, camera, fps=5)
    
    await service.start()
    
    assert service._is_streaming is True
    assert service._streaming_task is not None
    
    await service.stop()


@pytest.mark.asyncio
async def test_video_service_already_streaming():
    ws_client = AsyncMock()
    camera = MagicMock()
    service = VideoService(ws_client, camera, fps=5)
    
    await service.start()
    await service.start()
    
    assert service._is_streaming is True
    
    await service.stop()


@pytest.mark.asyncio
async def test_video_service_stop():
    ws_client = AsyncMock()
    camera = MagicMock()
    service = VideoService(ws_client, camera, fps=5)
    
    await service.start()
    await service.stop()
    
    assert service._is_streaming is False


@pytest.mark.asyncio
async def test_video_streaming_sends_frames(mock_websocket_client, mock_camera_handler):
    service = VideoService(mock_websocket_client, mock_camera_handler, fps=10)
    
    await service.start()
    await asyncio.sleep(0.3)
    await service.stop()
    
    mock_websocket_client.send_video_frame.assert_called()
    call_args = mock_websocket_client.send_video_frame.call_args
    assert call_args[0][0] == "base64encodedframe"


@pytest.mark.asyncio
async def test_video_streaming_skips_when_not_connected():
    ws_client = AsyncMock()
    ws_client.is_connected = False
    ws_client.settings = MagicMock()
    ws_client.settings.drone_id = "drone-123"
    ws_client.send_video_frame = AsyncMock()
    
    camera = MagicMock()
    camera.get_latest_frame = MagicMock(return_value="frame")
    
    service = VideoService(ws_client, camera, fps=10)
    
    await service.start()
    await asyncio.sleep(0.3)
    await service.stop()
    
    ws_client.send_video_frame.assert_not_called()


@pytest.mark.asyncio
async def test_video_streaming_skips_when_no_drone_id():
    ws_client = AsyncMock()
    ws_client.is_connected = True
    ws_client.settings = MagicMock()
    ws_client.settings.drone_id = None
    ws_client.send_video_frame = AsyncMock()
    
    camera = MagicMock()
    camera.get_latest_frame = MagicMock(return_value="frame")
    
    service = VideoService(ws_client, camera, fps=10)
    
    await service.start()
    await asyncio.sleep(0.3)
    await service.stop()
    
    ws_client.send_video_frame.assert_not_called()


@pytest.mark.asyncio
async def test_video_streaming_skips_when_no_frame():
    ws_client = AsyncMock()
    ws_client.is_connected = True
    ws_client.settings = MagicMock()
    ws_client.settings.drone_id = "drone-123"
    ws_client.send_video_frame = AsyncMock()
    
    camera = MagicMock()
    camera.get_latest_frame = MagicMock(return_value=None)
    
    service = VideoService(ws_client, camera, fps=10)
    
    await service.start()
    await asyncio.sleep(0.3)
    await service.stop()
    
    ws_client.send_video_frame.assert_not_called()


@pytest.mark.asyncio
async def test_video_streaming_handles_errors():
    ws_client = AsyncMock()
    ws_client.is_connected = True
    ws_client.settings = MagicMock()
    ws_client.settings.drone_id = "drone-123"
    ws_client.send_video_frame = AsyncMock(side_effect=Exception("Network error"))
    
    camera = MagicMock()
    camera.get_latest_frame = MagicMock(return_value="frame")
    
    service = VideoService(ws_client, camera, fps=10)
    
    await service.start()
    await asyncio.sleep(0.3)
    await service.stop()
    
    assert service._is_streaming is False


@pytest.mark.asyncio
async def test_video_fps_interval():
    ws_client = AsyncMock()
    ws_client.is_connected = True
    ws_client.settings = MagicMock()
    ws_client.settings.drone_id = "drone-123"
    ws_client.send_video_frame = AsyncMock()
    
    camera = MagicMock()
    camera.get_latest_frame = MagicMock(return_value="frame")
    
    service = VideoService(ws_client, camera, fps=5)
    
    await service.start()
    await asyncio.sleep(0.5)
    await service.stop()
    
    assert ws_client.send_video_frame.call_count >= 2
    assert ws_client.send_video_frame.call_count <= 3


import asyncio
