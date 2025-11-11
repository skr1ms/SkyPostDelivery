import logging
import cv2
from pyzbar import pyzbar
from typing import Optional, Dict
import asyncio
import time
import json
from datetime import datetime

logger = logging.getLogger(__name__)


class QRScanner:
    def __init__(self, camera_device: int = 0, mock_mode: bool = False,
                 stable_frames: int = 3, scan_interval: float = 0.1):
        self.camera_device = camera_device
        self.camera = None
        self.qr_decoder = None
        self._mock_mode = mock_mode
        self.stable_frames = stable_frames
        self.scan_interval = scan_interval

        self._current_qr = None
        self._current_count = 0
        self._last_scan_time = 0
        self._frame_count = 0

        if self._mock_mode:
            logger.info("Running QR scanner in MOCK mode")
        else:
            self._initialize_camera()
            self._initialize_qr_decoder()

    @property
    def mock_mode(self) -> bool:
        return self._mock_mode

    def _initialize_camera(self):
        try:
            logger.info(
                f"Attempting to open camera device {self.camera_device}")

            camera_paths = [
                f"/dev/video{self.camera_device}",
                self.camera_device,
                f"/dev/video{1 if self.camera_device == 0 else 0}",
            ]

            backends = [
                (cv2.CAP_V4L2, "V4L2"),
                (cv2.CAP_ANY, "ANY"),
            ]

            for cam_path in camera_paths:
                for backend, backend_name in backends:
                    try:
                        logger.info(
                            f"Trying camera '{cam_path}' with backend {backend_name}")
                        self.camera = cv2.VideoCapture(cam_path, backend)

                        if self.camera and self.camera.isOpened():
                            self.camera.set(cv2.CAP_PROP_FRAME_WIDTH, 640)
                            self.camera.set(cv2.CAP_PROP_FRAME_HEIGHT, 480)
                            self.camera.set(cv2.CAP_PROP_FPS, 30)
                            self.camera.set(
                                cv2.CAP_PROP_FOURCC, cv2.VideoWriter_fourcc('M', 'J', 'P', 'G'))

                            time.sleep(0.5)

                            for attempt in range(5):
                                ret, frame = self.camera.read()
                                if ret and frame is not None:
                                    logger.info(
                                        f"Camera successfully initialized on '{cam_path}' "
                                        f"with backend {backend_name}, frame size: {frame.shape}")
                                    return
                                logger.warning(
                                    f"Read attempt {attempt + 1}/5 failed, retrying...")
                                time.sleep(0.2)

                            logger.warning(
                                f"Camera '{cam_path}' opened but cannot read frame after 5 attempts")
                            self.camera.release()
                            self.camera = None
                        else:
                            logger.info(
                                f"Failed to open camera '{cam_path}' with {backend_name}")
                    except Exception as e:
                        logger.warning(
                            f"Error with camera '{cam_path}' backend {backend_name}: {e}")
                        if self.camera:
                            try:
                                self.camera.release()
                            except:
                                pass
                            self.camera = None
                        continue

            raise Exception("Failed to open camera with any backend or path")

        except Exception as e:
            logger.error(f"Failed to initialize camera: {e}")
            logger.warning(
                "Switching to MOCK mode due to camera initialization failure")
            self._mock_mode = True
            self.camera = None

    def _initialize_qr_decoder(self):
        if self._mock_mode:
            return
        try:
            self.qr_decoder = pyzbar
            logger.info("QR decoder initialized")
        except Exception as e:
            logger.error(f"Failed to initialize QR decoder: {e}")
            logger.warning("Install pyzbar for QR scanning")
            self._mock_mode = True

    def read_frame(self):
        if self._mock_mode:
            logger.debug("QR scanner is in mock mode; read_frame returns None")
            return None

        if self.camera is None:
            self._initialize_camera()

        if not self.camera or not self.camera.isOpened():
            logger.error("Camera is not opened when attempting to read frame")
            return None

        ret, frame = self.camera.read()
        if not ret:
            logger.warning("read_frame(): camera.read() returned False")
            return None
        return frame

    def extract_qr_info(self, qr_data: str) -> Optional[Dict]:
        try:
            data = json.loads(qr_data)
            return {
                "user_id": data.get("user_id", "unknown")[:13],
                "email": data.get("email", "unknown"),
                "name": data.get("name", "unknown"),
                "issued_at": data.get("issued_at"),
                "expires_at": data.get("expires_at"),
            }
        except Exception:
            return None

    def _validate_qr_locally(self, qr_data: str) -> tuple[bool, Optional[str]]:
        try:
            try:
                data = json.loads(qr_data)
            except json.JSONDecodeError as e:
                logger.warning(f"Invalid JSON in QR: {e}")
                return False, "Invalid JSON format"

            required_fields = ["user_id", "email", "expires_at", "signature"]
            missing_fields = [
                field for field in required_fields if field not in data]

            if missing_fields:
                logger.warning(
                    f"Missing required fields in QR: {missing_fields}")
                return False, f"Missing fields: {', '.join(missing_fields)}"

            user_id = data.get("user_id", "")
            if not user_id or len(user_id) != 36:
                logger.warning(f"Invalid user_id format: {user_id}")
                return False, "Invalid user_id format"

            if user_id.count('-') != 4:
                logger.warning(f"Invalid UUID pattern: {user_id}")
                return False, "Invalid UUID pattern"

            expires_at_str = data.get("expires_at")
            if expires_at_str:
                try:
                    if 'T' in expires_at_str:
                        expires_at = datetime.fromisoformat(
                            expires_at_str.replace('Z', '+00:00'))
                    else:
                        expires_at = datetime.fromtimestamp(
                            float(expires_at_str))

                    if datetime.now() > expires_at:
                        logger.warning(f"QR code expired: {expires_at}")
                        return False, "QR code expired"
                except (ValueError, TypeError) as e:
                    logger.warning(f"Invalid expires_at format: {e}")
                    return False, "Invalid expiration date format"

            signature = data.get("signature", "")
            if not signature or len(signature) < 20:
                logger.warning("Invalid or missing signature")
                return False, "Invalid signature"

            logger.info(
                f"QR local validation passed for user_id: {user_id[:8]}...")
            return True, None

        except Exception as e:
            logger.error(f"Unexpected error in local QR validation: {e}")
            return False, f"Validation error: {str(e)}"

    def scan_once(self) -> Optional[str]:
        if self._mock_mode:
            return None

        try:
            self._frame_count += 1
            frame = self.read_frame()
            if frame is None:
                return None

            decoded_objects = self.qr_decoder.decode(frame)

            if decoded_objects:
                qr_data = decoded_objects[0].data.decode('utf-8').strip()

                if qr_data == self._current_qr:
                    self._current_count += 1
                else:
                    self._current_qr = qr_data
                    self._current_count = 1
                    logger.info(
                        f"New QR detected: {qr_data[:50]}... (1/{self.stable_frames})")

                if self._current_count >= self.stable_frames:
                    logger.info(f"QR code stabilized: {qr_data[:100]}...")

                    is_valid, error_reason = self._validate_qr_locally(qr_data)
                    if not is_valid:
                        logger.warning(f"QR validation failed: {error_reason}")
                        self._current_qr = None
                        self._current_count = 0
                        return None

                    logger.info("QR validation passed")
                    self._current_qr = None
                    self._current_count = 0
                    return qr_data
                else:
                    logger.info(
                        f"QR stability: {self._current_count}/{self.stable_frames}")
            else:
                if self._current_qr:
                    logger.info("QR lost, resetting counter")
                self._current_qr = None
                self._current_count = 0

                if self._frame_count % 100 == 0:
                    logger.info(
                        f"Scanned {self._frame_count} frames, no QR detected")

            return None

        except Exception as e:
            logger.error(f"Failed to scan QR code: {e}")
            return None

    def scan_once_fast(self) -> Optional[str]:
        if self._mock_mode:
            return None

        try:
            frame = self.read_frame()
            if frame is None:
                return None

            decoded_objects = self.qr_decoder.decode(frame)

            if decoded_objects:
                qr_data = decoded_objects[0].data.decode('utf-8').strip()
                logger.info(f"QR code scanned: {qr_data[:100]}...")

                is_valid, error_reason = self._validate_qr_locally(qr_data)
                if not is_valid:
                    logger.warning(f"QR validation failed: {error_reason}")
                    return None

                logger.info("QR validation passed")
                return qr_data

            return None

        except Exception as e:
            logger.error(f"Failed to scan QR code: {e}")
            return None

    async def scan_continuous(self, interval: float = 0.5, callback=None):
        logger.info(f"Starting continuous QR scanning (interval: {interval}s)")
        while True:
            qr_data = self.scan_once()
            if qr_data and callback:
                try:
                    await callback(qr_data)
                except Exception as e:
                    logger.error(f"Callback error: {e}")
            await asyncio.sleep(interval)

    def close(self):
        if self.camera and not self._mock_mode:
            self.camera.release()
            self.camera = None
            logger.info("Camera closed")
