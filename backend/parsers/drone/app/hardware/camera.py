import cv2
import base64
import logging
import numpy as np
from typing import Optional
from config.config import settings

logger = logging.getLogger(__name__)


class CameraController:
    def __init__(self):
        self.camera = None
        self.is_active = False
        self.use_mock = settings.use_mock_hardware
        self.use_picamera = False

    def initialize(self) -> bool:
        try:
            if self.use_mock:
                logger.info("Camera initialized in MOCK mode")
                self.is_active = True
                return True

            try:
                from picamera2 import Picamera2
                self.use_picamera = True
                self.camera = Picamera2()

                config = self.camera.create_video_configuration(
                    main={"size": (640, 480), "format": "RGB888"}
                )
                self.camera.configure(config)
                self.camera.start()

                self.is_active = True
                logger.info("Camera initialized successfully using picamera2")
                return True

            except ImportError as e:
                logger.warning(
                    f"picamera2 not available: {e}, falling back to OpenCV")
                self.use_picamera = False
            except Exception as e:
                logger.warning(
                    f"Failed to initialize picamera2: {e}, falling back to OpenCV")
                self.use_picamera = False

                camera_index = getattr(settings, 'camera_index', 0)
                self.camera = cv2.VideoCapture(camera_index)

                if not self.camera.isOpened():
                    logger.error(
                        f"Failed to open camera at index {camera_index}")
                    return False

                self.camera.set(cv2.CAP_PROP_FRAME_WIDTH, 640)
                self.camera.set(cv2.CAP_PROP_FRAME_HEIGHT, 480)
                self.camera.set(cv2.CAP_PROP_FPS, 30)

                self.is_active = True
                logger.info(
                    f"Camera initialized successfully on index {camera_index}")
                return True

        except Exception as e:
            logger.error(f"Error initializing camera: {e}")
            return False

    def capture_frame_raw(self) -> Optional['np.ndarray']:
        try:
            if self.use_mock:
                import random
                mock_frame = np.zeros((480, 640, 3), dtype=np.uint8)
                mock_frame[:, :] = [random.randint(0, 100), random.randint(
                    0, 100), random.randint(100, 255)]
                cv2.putText(mock_frame, "MOCK FRAME", (50, 240),
                            cv2.FONT_HERSHEY_SIMPLEX, 2, (255, 255, 255), 3)
                return mock_frame

            if not self.is_active or not self.camera:
                logger.warning("Camera not active")
                return None

            if self.use_picamera:
                frame = self.camera.capture_array()

                if frame is None:
                    logger.warning("Failed to capture frame from picamera2")
                    return None

                frame_bgr = cv2.cvtColor(frame, cv2.COLOR_RGB2BGR)
                return frame_bgr

            else:
                ret, frame = self.camera.read()

                if not ret or frame is None:
                    logger.warning("Failed to capture frame from camera")
                    return None

                return frame

        except Exception as e:
            logger.error(f"Error capturing frame: {e}")
            return None

    def capture_frame(self) -> Optional[str]:
        try:
            frame = self.capture_frame_raw()

            if frame is None:
                return None

            _, buffer = cv2.imencode(
                '.jpg', frame, [cv2.IMWRITE_JPEG_QUALITY, 80])
            frame_base64 = base64.b64encode(buffer).decode('utf-8')

            return frame_base64

        except Exception as e:
            logger.error(f"Error capturing frame: {e}")
            return None

    def close(self):
        try:
            self.is_active = False
            if self.camera:
                if self.use_picamera:
                    self.camera.stop()
                    self.camera.close()
                else:
                    self.camera.release()
                logger.info("Camera released")
        except Exception as e:
            logger.error(f"Error releasing camera: {e}")
