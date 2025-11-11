import cv2
import numpy as np
import logging
from typing import Optional, List, Tuple
from dataclasses import dataclass
from config.config import settings

logger = logging.getLogger(__name__)


@dataclass
class MarkerDetection:
    marker_id: int
    center_x: float
    center_y: float
    distance: float
    angle: float
    corners: np.ndarray
    rvec: np.ndarray
    tvec: np.ndarray


class ArucoDetector:
    def __init__(self, camera_matrix: Optional[np.ndarray] = None,
                 dist_coeffs: Optional[np.ndarray] = None):
        self.aruco_dict = cv2.aruco.getPredefinedDictionary(
            settings.aruco_dict)
        self.parameters = cv2.aruco.DetectorParameters()
        self.marker_size = settings.marker_size_m

        if camera_matrix is None:
            self.camera_matrix = np.array([
                [640, 0, 320],
                [0, 640, 240],
                [0, 0, 1]
            ], dtype=np.float32)
        else:
            self.camera_matrix = camera_matrix

        if dist_coeffs is None:
            self.dist_coeffs = np.zeros(5, dtype=np.float32)
        else:
            self.dist_coeffs = dist_coeffs

        self.last_detection_time = 0
        logger.info(
            f"ArUco detector initialized with dict: {settings.aruco_dict_type}, marker size: {self.marker_size}m")

    def detect_markers(self, frame: np.ndarray) -> List[MarkerDetection]:
        if frame is None or frame.size == 0:
            logger.warning("Empty frame received")
            return []

        gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)

        corners, ids, rejected = cv2.aruco.detectMarkers(
            gray, self.aruco_dict, parameters=self.parameters
        )

        if ids is None or len(ids) == 0:
            return []

        detections = []

        rvecs, tvecs, _ = cv2.aruco.estimatePoseSingleMarkers(
            corners, self.marker_size, self.camera_matrix, self.dist_coeffs
        )

        for i, marker_id in enumerate(ids.flatten()):
            corner = corners[i][0]
            center_x = np.mean(corner[:, 0])
            center_y = np.mean(corner[:, 1])

            tvec = tvecs[i][0]
            rvec = rvecs[i][0]

            distance = np.linalg.norm(tvec)

            rotation_matrix, _ = cv2.Rodrigues(rvec)
            angle = np.arctan2(rotation_matrix[1, 0], rotation_matrix[0, 0])
            angle_deg = np.degrees(angle)

            detection = MarkerDetection(
                marker_id=int(marker_id),
                center_x=float(center_x),
                center_y=float(center_y),
                distance=float(distance),
                angle=float(angle_deg),
                corners=corner,
                rvec=rvec,
                tvec=tvec
            )

            detections.append(detection)

        if detections:
            marker_ids = [d.marker_id for d in detections]
            logger.debug(f"Detected markers: {marker_ids}")

        return detections

    def get_marker_offset(self, detection: MarkerDetection, frame_width: int, frame_height: int) -> Tuple[float, float]:
        center_x_norm = (detection.center_x -
                         frame_width / 2) / (frame_width / 2)
        center_y_norm = (detection.center_y - frame_height /
                         2) / (frame_height / 2)

        offset_x = center_x_norm * detection.distance * 0.5
        offset_y = center_y_norm * detection.distance * 0.5

        return offset_x, offset_y

    def find_marker_by_id(self, detections: List[MarkerDetection], marker_id: int) -> Optional[MarkerDetection]:
        for detection in detections:
            if detection.marker_id == marker_id:
                return detection
        return None

    def draw_detections(self, frame: np.ndarray, detections: List[MarkerDetection]) -> np.ndarray:
        if not detections:
            return frame

        annotated_frame = frame.copy()

        for detection in detections:
            corners = detection.corners.reshape((4, 2)).astype(int)

            for i in range(4):
                cv2.line(annotated_frame, tuple(corners[i]),
                         tuple(corners[(i + 1) % 4]), (0, 255, 0), 2)

            center = (int(detection.center_x), int(detection.center_y))
            cv2.circle(annotated_frame, center, 5, (0, 0, 255), -1)

            text = f"ID:{detection.marker_id} D:{detection.distance:.2f}m"
            cv2.putText(annotated_frame, text,
                        (center[0] - 50, center[1] - 10),
                        cv2.FONT_HERSHEY_SIMPLEX, 0.5, (255, 255, 255), 2)

            cv2.drawFrameAxes(annotated_frame, self.camera_matrix,
                              self.dist_coeffs, detection.rvec,
                              detection.tvec, self.marker_size * 0.5)

        return annotated_frame

    def is_centered(self, detection: MarkerDetection, frame_width: int, frame_height: int) -> bool:
        offset_x, offset_y = self.get_marker_offset(
            detection, frame_width, frame_height)
        offset_magnitude = np.sqrt(offset_x**2 + offset_y**2)
        return offset_magnitude < settings.center_threshold
