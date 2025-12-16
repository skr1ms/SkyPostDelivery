import base64
import cv2
import rospy
from sensor_msgs.msg import Image
from cv_bridge import CvBridge
from typing import Optional
import logging

logger = logging.getLogger(__name__)


class CameraHandler:
    def __init__(self):
        self.bridge = CvBridge()
        self.latest_frame_base64: Optional[str] = None
        self._subscriber = None

    def start(self):
        self._subscriber = rospy.Subscriber(
            "/main_camera/image_raw", Image, self._on_frame, queue_size=1
        )
        logger.info("Camera handler started")

    def _on_frame(self, msg: Image):
        try:
            cv_image = self.bridge.imgmsg_to_cv2(msg, desired_encoding="bgr8")
            encode_param = [cv2.IMWRITE_JPEG_QUALITY, 80]
            _, buffer = cv2.imencode(".jpg", cv_image, encode_param)
            self.latest_frame_base64 = base64.b64encode(buffer).decode("utf-8")
        except Exception as e:
            logger.error(f"Error processing camera frame: {e}")

    def get_latest_frame(self) -> Optional[str]:
        return self.latest_frame_base64

    def stop(self):
        if self._subscriber:
            self._subscriber.unregister()
            logger.info("Camera handler stopped")
