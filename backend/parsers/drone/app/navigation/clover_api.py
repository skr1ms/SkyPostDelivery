import rospy
import logging
from typing import Optional, Tuple
from std_srvs.srv import Trigger
from sensor_msgs.msg import Range

logger = logging.getLogger(__name__)


class CloverAPI:
    def __init__(self):
        self.node_initialized = False
        self.navigate = None
        self.get_telemetry_service = None
        self.land_service = None
        self.rangefinder_distance = 0.0
        self.rangefinder_sub = None

    def initialize(self) -> bool:
        try:
            if not rospy.core.is_initialized():
                rospy.init_node('drone_delivery',
                                anonymous=True, disable_signals=True)

            if not self.node_initialized:
                from clover import srv

                rospy.wait_for_service('navigate')
                rospy.wait_for_service('get_telemetry')
                rospy.wait_for_service('land')

                self.navigate = rospy.ServiceProxy('navigate', srv.Navigate)
                self.get_telemetry_service = rospy.ServiceProxy(
                    'get_telemetry', srv.GetTelemetry)
                self.land_service = rospy.ServiceProxy('land', Trigger)

                if self.rangefinder_sub is None:
                    self.rangefinder_sub = rospy.Subscriber(
                        '/rangefinder/range',
                        Range,
                        self._rangefinder_callback
                    )

                self.node_initialized = True

            logger.info("Clover API initialized with rangefinder")
            return True

        except Exception as e:
            logger.error(f"Failed to initialize Clover API: {e}")
            return False

    def _rangefinder_callback(self, msg: Range):
        self.rangefinder_distance = msg.range

    def get_rangefinder_distance(self) -> float:
        return self.rangefinder_distance

    def navigate_to(self, x: float, y: float, z: float,
                    frame_id: str = 'body',
                    speed: float = 0.5,
                    auto_arm: bool = False) -> bool:
        try:
            result = self.navigate(
                x=x, y=y, z=z,
                yaw=float('nan'),
                speed=speed,
                frame_id=frame_id,
                auto_arm=auto_arm
            )
            return result.success
        except Exception as e:
            logger.error(f"Navigate failed: {e}")
            return False

    def get_telemetry(self, frame_id: str = ''):
        try:
            if frame_id:
                return self.get_telemetry_service(frame_id=frame_id)
            return self.get_telemetry_service()
        except Exception as e:
            logger.error(f"Get telemetry failed: {e}")
            return None

    def get_position(self, frame_id: str = 'map') -> Optional[Tuple[float, float, float]]:
        try:
            telem = self.get_telemetry_service(frame_id=frame_id)
            return (telem.x, telem.y, telem.z)
        except Exception as e:
            logger.error(f"Get position failed: {e}")
            return None

    def get_altitude(self) -> float:
        try:
            telem = self.get_telemetry_service(frame_id='body')
            return telem.z
        except Exception as e:
            logger.error(f"Get altitude failed: {e}")
            return 0.0

    def get_battery(self) -> float:
        try:
            telem = self.get_telemetry_service()
            return telem.voltage
        except Exception as e:
            logger.error(f"Get battery failed: {e}")
            return 0.0

    def reached_target(self, tolerance: float = 0.3) -> bool:
        try:
            telem = self.get_telemetry_service(frame_id='navigate_target')
            distance = (telem.x ** 2 + telem.y ** 2 + telem.z ** 2) ** 0.5
            return distance < tolerance
        except Exception as e:
            logger.error(f"Check target failed: {e}")
            return False

    def land(self) -> bool:
        try:
            result = self.land_service()
            logger.info("Landing initiated")
            return result.success
        except Exception as e:
            logger.error(f"Land failed: {e}")
            return False

    def wait_arrival(self, tolerance: float = 0.3, timeout: float = 60.0):
        rate = rospy.Rate(10)
        start_time = rospy.get_time()

        while not rospy.is_shutdown():
            if self.reached_target(tolerance):
                logger.info("Target reached")
                return True

            if rospy.get_time() - start_time > timeout:
                logger.warning("Wait arrival timeout")
                return False

            rate.sleep()

        return False
