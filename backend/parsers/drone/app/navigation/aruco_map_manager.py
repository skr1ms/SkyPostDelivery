import logging
from typing import Dict, Tuple, Optional
from pathlib import Path

logger = logging.getLogger(__name__)


class ArucoMapManager:
    def __init__(self, map_file: str):
        self.map_file = map_file
        self.markers: Dict[int, Tuple[float, float, float]] = {}
        self.marker_sizes: Dict[int, float] = {}

    def load_map(self) -> bool:
        try:
            if not Path(self.map_file).exists():
                logger.error(f"Map file not found: {self.map_file}")
                return False

            with open(self.map_file, 'r') as f:
                for line in f:
                    line = line.strip()

                    if not line or line.startswith('#'):
                        continue

                    parts = line.split()
                    if len(parts) < 4:
                        continue

                    marker_id = int(parts[0])
                    x = float(parts[1])
                    y = float(parts[2])
                    z = float(parts[3])

                    self.markers[marker_id] = (x, y, z)

                    if len(parts) >= 5:
                        self.marker_sizes[marker_id] = float(parts[4])

            logger.info(f"Loaded {len(self.markers)} markers from map")
            return True

        except Exception as e:
            logger.error(f"Failed to load map: {e}")
            return False

    def get_marker_position(self, marker_id: int) -> Optional[Tuple[float, float, float]]:
        return self.markers.get(marker_id)

    def get_marker_xy(self, marker_id: int) -> Optional[Tuple[float, float]]:
        pos = self.markers.get(marker_id)
        if pos:
            return (pos[0], pos[1])
        return None

    def is_valid_marker(self, marker_id: int) -> bool:
        return marker_id in self.markers

    def find_nearest_marker(self, x: float, y: float) -> Optional[int]:
        min_dist = float('inf')
        nearest_id = None

        for marker_id, (mx, my, _) in self.markers.items():
            dist = ((x - mx) ** 2 + (y - my) ** 2) ** 0.5
            if dist < min_dist:
                min_dist = dist
                nearest_id = marker_id

        return nearest_id

    def get_distance_between_markers(self, id1: int, id2: int) -> Optional[float]:
        pos1 = self.get_marker_xy(id1)
        pos2 = self.get_marker_xy(id2)

        if pos1 and pos2:
            return ((pos2[0] - pos1[0]) ** 2 + (pos2[1] - pos1[1]) ** 2) ** 0.5

        return None

    def get_all_marker_ids(self):
        return list(self.markers.keys())
