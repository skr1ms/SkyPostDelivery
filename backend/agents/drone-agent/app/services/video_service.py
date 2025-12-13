import asyncio
import logging
from typing import Optional
from ..core.websocket_client import WebSocketClient
from ..hardware.camera_handler import CameraHandler

logger = logging.getLogger(__name__)


class VideoService:
    def __init__(
        self,
        websocket_client: WebSocketClient,
        camera_handler: CameraHandler,
        fps: int = 5,
    ):
        self.ws_client = websocket_client
        self.camera_handler = camera_handler
        self.fps = fps
        self._streaming_task: Optional[asyncio.Task] = None
        self._is_streaming = False

    async def start(self):
        if self._is_streaming:
            logger.warning("Video streaming already active")
            return

        self._is_streaming = True
        self._streaming_task = asyncio.create_task(self._stream_loop())
        logger.info(f"Video streaming started at {self.fps} FPS")

    async def _stream_loop(self):
        interval = 1.0 / self.fps
        frame_counter = 0

        while self._is_streaming:
            try:
                if (
                    not self.ws_client.is_connected
                    or not self.ws_client.settings.drone_id
                ):
                    await asyncio.sleep(interval)
                    continue

                frame_base64 = self.camera_handler.get_latest_frame()

                if frame_base64:
                    await self.ws_client.send_video_frame(frame_base64)
                    frame_counter += 1

                    if frame_counter % 30 == 0:
                        logger.info(f"Video frame #{frame_counter} sent")
                    else:
                        logger.debug(f"Video frame #{frame_counter} sent")

                await asyncio.sleep(interval)

            except Exception as e:
                logger.error(f"Error in video stream loop: {e}")
                await asyncio.sleep(1)

    async def stop(self):
        self._is_streaming = False
        if self._streaming_task:
            self._streaming_task.cancel()
            try:
                await self._streaming_task
            except asyncio.CancelledError:
                pass
        logger.info("Video streaming stopped")
