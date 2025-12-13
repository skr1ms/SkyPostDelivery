import rospy
from sensor_msgs.msg import BatteryState
from geometry_msgs.msg import PoseStamped
from sensor_msgs.msg import Range
from mavros_msgs.msg import State
from std_msgs.msg import Bool
import asyncio
import logging
from typing import Optional, Callable
from ..models.telemetry import Telemetry, Battery, Pose, MavrosState

logger = logging.getLogger(__name__)


class ROSBridge:
    def __init__(self, event_loop: asyncio.AbstractEventLoop):
        self.event_loop = event_loop
        self.telemetry = Telemetry()

        self.arrival_callback: Optional[Callable] = None
        self.drop_ready_callback: Optional[Callable] = None
        self.home_callback: Optional[Callable] = None

        self._subscribers = []
        self._last_arrival_time = 0
        self._last_drop_ready_time = 0
        self._last_home_time = 0

    def start(self):
        self._subscribers.append(
            rospy.Subscriber(
                "/mavros/battery", BatteryState, self._battery_callback, queue_size=1
            )
        )
        self._subscribers.append(
            rospy.Subscriber(
                "/mavros/local_position/pose",
                PoseStamped,
                self._pose_callback,
                queue_size=1,
            )
        )
        self._subscribers.append(
            rospy.Subscriber(
                "/rangefinder/range", Range, self._rangefinder_callback, queue_size=1
            )
        )
        self._subscribers.append(
            rospy.Subscriber("/mavros/state", State, self._state_callback, queue_size=1)
        )
        self._subscribers.append(
            rospy.Subscriber(
                "/drone/delivery/arrived",
                PoseStamped,
                self._arrival_callback,
                queue_size=1,
            )
        )
        self._subscribers.append(
            rospy.Subscriber(
                "/drone/delivery/drop_ready",
                PoseStamped,
                self._drop_ready_callback_handler,
                queue_size=1,
            )
        )
        self._subscribers.append(
            rospy.Subscriber(
                "/drone/delivery/home_arrived",
                PoseStamped,
                self._home_callback_handler,
                queue_size=1,
            )
        )

        logger.info("ROS Bridge started, all subscribers initialized")

    def _battery_callback(self, msg: BatteryState):
        self.telemetry.battery = Battery(
            voltage=msg.voltage,
            percentage=msg.percentage * 100.0 if msg.percentage > 0 else None,
            current=msg.current,
        )

    def _pose_callback(self, msg: PoseStamped):
        self.telemetry.pose = Pose(
            x=msg.pose.position.x,
            y=msg.pose.position.y,
            z=msg.pose.position.z,
            orientation={
                "x": msg.pose.orientation.x,
                "y": msg.pose.orientation.y,
                "z": msg.pose.orientation.z,
                "w": msg.pose.orientation.w,
            },
        )

    def _rangefinder_callback(self, msg: Range):
        self.telemetry.altitude = msg.range

    def _state_callback(self, msg: State):
        self.telemetry.state = MavrosState(
            armed=msg.armed, connected=msg.connected, mode=msg.mode
        )

    def _arrival_callback(self, msg: PoseStamped):
        current_time = rospy.Time.now().to_sec()
        if current_time - self._last_arrival_time > 5:
            self._last_arrival_time = current_time
            logger.info("Drone arrival detected")
            if self.arrival_callback:
                asyncio.run_coroutine_threadsafe(
                    self.arrival_callback(), self.event_loop
                )

    def _drop_ready_callback_handler(self, msg: PoseStamped):
        current_time = rospy.Time.now().to_sec()
        if current_time - self._last_drop_ready_time > 5:
            self._last_drop_ready_time = current_time
            logger.info("Cargo drop ready detected")
            if self.drop_ready_callback:
                asyncio.run_coroutine_threadsafe(
                    self.drop_ready_callback(), self.event_loop
                )

    def _home_callback_handler(self, msg: PoseStamped):
        current_time = rospy.Time.now().to_sec()
        if current_time - self._last_home_time > 5:
            self._last_home_time = current_time
            logger.info("Drone home arrival detected")
            if self.home_callback:
                asyncio.run_coroutine_threadsafe(self.home_callback(), self.event_loop)

    def send_drop_confirmation(self) -> bool:
        try:
            pub = rospy.Publisher("/drone/delivery/drop_confirm", Bool, queue_size=1)
            rospy.sleep(0.5)
            pub.publish(Bool(data=True))
            logger.info("Drop confirmation sent")
            return True
        except Exception as e:
            logger.error(f"Failed to send drop confirmation: {e}")
            return False

    def set_arrival_callback(self, callback: Callable):
        self.arrival_callback = callback

    def set_drop_ready_callback(self, callback: Callable):
        self.drop_ready_callback = callback

    def set_home_callback(self, callback: Callable):
        self.home_callback = callback

    def get_telemetry(self) -> Telemetry:
        return self.telemetry

    def stop(self):
        for sub in self._subscribers:
            sub.unregister()
        logger.info("ROS Bridge stopped")
