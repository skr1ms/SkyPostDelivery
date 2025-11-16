#!/usr/bin/env python3

import asyncio
import signal
import logging
import threading
import rospy
from sensor_msgs.msg import Image, BatteryState, Range
from geometry_msgs.msg import PoseStamped
from mavros_msgs.msg import State
from cv_bridge import CvBridge
import cv2
import base64

from config.config import settings
from app.dependencies import websocket_service, cleanup
from app.hardware.flight_manager import flight_manager
from app.hardware.ros_bridge import ros_bridge as ros_topic_bridge

logging.basicConfig(
    level=settings.log_level,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class ROSBridge:
    def __init__(self):
        self.bridge = CvBridge()
        self.latest_frame = None
        self.latest_battery = None
        self.latest_pose = None
        self.latest_altitude = None
        self.latest_state = None

    def camera_callback(self, msg: Image):
        try:
            cv_image = self.bridge.imgmsg_to_cv2(msg, desired_encoding='bgr8')
            _, buffer = cv2.imencode(
                '.jpg', cv_image, [cv2.IMWRITE_JPEG_QUALITY, 80])
            self.latest_frame = base64.b64encode(buffer).decode('utf-8')
        except Exception as e:
            logger.error(f"Error processing camera frame: {e}")

    def battery_callback(self, msg: BatteryState):
        self.latest_battery = {
            'voltage': msg.voltage,
            'percentage': msg.percentage * 100 if msg.percentage > 0 else None,
            'current': msg.current
        }

    def pose_callback(self, msg: PoseStamped):
        self.latest_pose = {
            'x': msg.pose.position.x,
            'y': msg.pose.position.y,
            'z': msg.pose.position.z,
            'orientation': {
                'x': msg.pose.orientation.x,
                'y': msg.pose.orientation.y,
                'z': msg.pose.orientation.z,
                'w': msg.pose.orientation.w
            }
        }

    def rangefinder_callback(self, msg: Range):
        self.latest_altitude = msg.range

    def state_callback(self, msg: State):
        self.latest_state = {
            'armed': msg.armed,
            'connected': msg.connected,
            'mode': msg.mode
        }

    def get_frame(self):
        return self.latest_frame

    def get_telemetry(self):
        return {
            'battery': self.latest_battery,
            'pose': self.latest_pose,
            'altitude': self.latest_altitude,
            'state': self.latest_state
        }


class DroneApplication:
    def __init__(self):
        self.is_running = False
        self.ros_bridge = None
        self.ros_thread = None

    def init_ros(self):
        try:
            logger.info("Initializing ROS node")
            rospy.init_node('drone_parser', anonymous=False,
                            disable_signals=True)

            self.ros_bridge = ROSBridge()

            rospy.Subscriber('/main_camera/image_raw', Image,
                             self.ros_bridge.camera_callback, queue_size=1)

            rospy.Subscriber('/mavros/battery', BatteryState,
                             self.ros_bridge.battery_callback, queue_size=1)
            rospy.Subscriber('/mavros/local_position/pose', PoseStamped,
                             self.ros_bridge.pose_callback, queue_size=1)
            rospy.Subscriber('/rangefinder/range', Range,
                             self.ros_bridge.rangefinder_callback, queue_size=1)
            rospy.Subscriber('/mavros/state', State,
                             self.ros_bridge.state_callback, queue_size=1)

            logger.info("ROS subscribers initialized")

            rospy.spin()

        except Exception as e:
            logger.error(f"ROS initialization failed: {e}")
            self.is_running = False

    async def start(self):
        logger.info(f"Starting Drone Application - ID: {settings.drone_id}")
        logger.info(f"Drone IP: {settings.drone_ip}")
        logger.info(f"Parcel Automat IP: {settings.parcel_automat_ip}")
        logger.info(f"Connecting to: {settings.websocket_url}")

        self.is_running = True

        self.ros_thread = threading.Thread(target=self.init_ros, daemon=True)
        self.ros_thread.start()

        await asyncio.sleep(2)

        try:
            if self.ros_bridge:
                websocket_service.ros_bridge = self.ros_bridge
                websocket_service.flight_manager = flight_manager

            async def on_arrival():
                logger.info("Drone arrived at delivery location")
                logger.info("Notifying backend via WebSocket...")
                
                from app.models.schemas import DeliveryUpdatePayload
                if websocket_service.is_connected and settings.drone_id:
                    current_task = websocket_service.current_delivery_task
                    if current_task:
                        await websocket_service.send_delivery_update(
                            DeliveryUpdatePayload(
                                delivery_id=current_task.get("delivery_id", ""),
                                drone_status="arrived_at_destination",
                                order_id=current_task.get("order_id"),
                                parcel_automat_id=current_task.get("parcel_automat_id")
                            )
                        )
                        logger.info(f"Sent arrival notification with order_id={current_task.get('order_id')}, parcel_automat_id={current_task.get('parcel_automat_id')}")
                    else:
                        logger.warning("No current delivery task found, cannot send arrival notification")
                        await websocket_service.send_delivery_update(
                            DeliveryUpdatePayload(
                                delivery_id="unknown",
                                drone_status="arrived_at_destination"
                            )
                        )
                
            async def on_drop_ready():
                logger.info("Cargo drop ready - drone waiting for confirmation")
                logger.info("Backend should confirm cell is open and ready for drop")
                
                
            async def on_home():
                logger.info("Drone returned to home base")
                logger.info("Mission completed, drone is now IDLE")
                
                from app.models.schemas import StatusUpdatePayload, DroneStatus, Position
                if websocket_service.is_connected and settings.drone_id:
                    await websocket_service.send_status_update(
                        StatusUpdatePayload(
                            status=DroneStatus.IDLE,
                            battery_level=100.0,  # Will be updated from ROS
                            position=Position(latitude=0.0, longitude=0.0, altitude=0.0),
                            speed=0.0
                        )
                    )
            
            ros_topic_bridge.set_arrival_callback(on_arrival)
            ros_topic_bridge.set_drop_ready_callback(on_drop_ready)
            ros_topic_bridge.set_home_callback(on_home)
            
            await ros_topic_bridge.start_listeners()

            await websocket_service.connect()
        except asyncio.CancelledError:
            logger.info("Application cancelled")
        finally:
            await self.stop()

    async def stop(self):
        logger.info("Stopping Drone Application")
        self.is_running = False
        
        await ros_topic_bridge.stop_listeners()

        if not rospy.is_shutdown():
            rospy.signal_shutdown("Application stopping")

        await cleanup()
        logger.info("Drone Application stopped")


async def main():
    app = DroneApplication()

    loop = asyncio.get_event_loop()

    def signal_handler():
        logger.info("Received shutdown signal")
        asyncio.create_task(app.stop())

    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, signal_handler)

    try:
        await app.start()
    except KeyboardInterrupt:
        logger.info("Keyboard interrupt received")
    finally:
        await app.stop()


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("Application terminated")
