import rospy
import logging
from typing import Optional
from enum import Enum
from .clover_api import CloverAPI
from .aruco_map_manager import ArucoMapManager

logger = logging.getLogger(__name__)


class FlightState(Enum):
    IDLE = "idle"
    INITIALIZING = "initializing"
    TAKING_OFF = "taking_off"
    NAVIGATING = "navigating"
    APPROACHING_TARGET = "approaching_target"
    LANDING = "landing"
    LANDED = "landed"
    ERROR = "error"


class CloverNavigationController:
    def __init__(self, map_file: str):
        self.api = CloverAPI()
        self.map_manager = ArucoMapManager(map_file)
        self.state = FlightState.IDLE

        self.cruise_altitude = 1.5
        self.cruise_speed = 0.5
        self.landing_altitude = 0.5
        self.position_tolerance = 0.3

    def initialize(self) -> bool:
        try:
            self.state = FlightState.INITIALIZING

            if not self.api.initialize():
                logger.error("Failed to initialize Clover API")
                return False

            if not self.map_manager.load_map():
                logger.error("Failed to load ArUco map")
                return False

            logger.info("Navigation controller initialized")
            self.state = FlightState.IDLE
            return True

        except Exception as e:
            logger.error(f"Initialization failed: {e}")
            self.state = FlightState.ERROR
            return False

    def takeoff(self, altitude: Optional[float] = None) -> bool:
        try:
            self.state = FlightState.TAKING_OFF

            if altitude is None:
                altitude = self.cruise_altitude

            logger.info(f"Taking off to {altitude}m")

            success = self.api.navigate_to(
                x=0, y=0, z=altitude,
                frame_id='body',
                auto_arm=True
            )

            if not success:
                logger.error("Takeoff command failed")
                return False

            rospy.sleep(5)

            current_alt = self.api.get_altitude()
            logger.info(f"Current altitude: {current_alt}m")

            return True

        except Exception as e:
            logger.error(f"Takeoff failed: {e}")
            self.state = FlightState.ERROR
            return False

    def navigate_to_marker(self, target_marker_id: int) -> bool:
        try:
            self.state = FlightState.NAVIGATING

            if not self.map_manager.is_valid_marker(target_marker_id):
                logger.error(f"Invalid marker ID: {target_marker_id}")
                return False

            target_pos = self.map_manager.get_marker_xy(target_marker_id)
            if not target_pos:
                logger.error(f"Marker {target_marker_id} not found in map")
                return False

            logger.info(
                f"Navigating to marker {target_marker_id} at ({target_pos[0]:.2f}, {target_pos[1]:.2f})")

            success = self.api.navigate_to(
                x=target_pos[0],
                y=target_pos[1],
                z=self.cruise_altitude,
                frame_id='aruco_map',
                speed=self.cruise_speed
            )

            if not success:
                logger.error("Navigate command failed")
                return False

            logger.info("Waiting for arrival...")
            arrived = self.api.wait_arrival(
                tolerance=self.position_tolerance,
                timeout=60.0
            )

            if not arrived:
                logger.warning("Did not reach target position")
                return False

            logger.info(f"Arrived at marker {target_marker_id}")

            logger.info("Verifying aruco_map coordinate system is active")
            map_telem = self.api.get_telemetry(frame_id='aruco_map')
            if map_telem and hasattr(map_telem, 'x'):
                logger.info(
                    f"Position in aruco_map: x={map_telem.x:.2f}, y={map_telem.y:.2f}, z={map_telem.z:.2f}")
            else:
                logger.warning(
                    "aruco_map coordinate system not available, map may not be detected")

            return True

        except Exception as e:
            logger.error(f"Navigation failed: {e}")
            self.state = FlightState.ERROR
            return False

    def land_on_marker(self, marker_id: int) -> bool:
        try:
            self.state = FlightState.APPROACHING_TARGET

            target_pos = self.map_manager.get_marker_xy(marker_id)
            if not target_pos:
                logger.error(f"Marker {marker_id} not found")
                return False

            logger.info(
                f"Descending to {self.landing_altitude}m above marker {marker_id}")

            success = self.api.navigate_to(
                x=target_pos[0],
                y=target_pos[1],
                z=self.landing_altitude,
                frame_id='aruco_map',
                speed=0.3
            )

            if not success:
                logger.error("Descent failed")
                return False

            self.api.wait_arrival(tolerance=0.2, timeout=30.0)

            logger.info(
                f"Switching to visual navigation on marker {marker_id}")

            marker_frame = f'aruco_{marker_id}'

            rospy.sleep(1)

            telem = self.api.get_telemetry(frame_id=marker_frame)

            if telem and hasattr(telem, 'z') and telem.z > 0:
                logger.info(f"Marker {marker_id} detected, centering above it")

                success = self.api.navigate_to(
                    x=0,
                    y=0,
                    z=self.landing_altitude,
                    frame_id=marker_frame,
                    speed=0.2
                )

                if success:
                    self.api.wait_arrival(tolerance=0.15, timeout=20.0)
                    logger.info(f"Centered above marker {marker_id}")
                else:
                    logger.warning(
                        "Visual centering failed, proceeding with landing")
            else:
                logger.warning(
                    f"Marker {marker_id} not visible, landing at map coordinates")

            logger.info("Landing with rangefinder control...")
            self.state = FlightState.LANDING

            rate = rospy.Rate(10)
            landing_height = 0.15

            while self.api.get_rangefinder_distance() > landing_height:
                height = self.api.get_rangefinder_distance()
                logger.info(f"Height: {height:.2f}m")

                if height < landing_height + 0.05:
                    break

                rate.sleep()

            success = self.api.land()

            if success:
                rospy.sleep(3)
                self.state = FlightState.LANDED
                logger.info("Landed successfully")

            return success

        except Exception as e:
            logger.error(f"Landing failed: {e}")
            self.state = FlightState.ERROR
            return False

    def execute_delivery(self, target_marker_id: int = 52) -> bool:
        try:
            logger.info(
                f"Starting delivery mission to marker {target_marker_id}")

            if not self.takeoff():
                logger.error("Takeoff failed")
                return False

            if not self.navigate_to_marker(target_marker_id):
                logger.error("Navigation failed")
                return False

            if not self.land_on_marker(target_marker_id):
                logger.error("Landing failed")
                return False

            logger.info("Delivery mission completed successfully")
            return True

        except Exception as e:
            logger.error(f"Delivery mission failed: {e}")
            self.state = FlightState.ERROR
            return False
