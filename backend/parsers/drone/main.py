#!/usr/bin/env python3
"""
Drone ROS-WebSocket Bridge
Runs with system Python3 + ROS environment
Bridges ROS topics to drone-service via WebSocket
"""
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

logging.basicConfig(
    level=settings.log_level,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class ROSBridge:
    """Bridge between ROS topics and WebSocket"""
    
    def __init__(self):
        self.bridge = CvBridge()
        self.latest_frame = None
        self.latest_battery = None
        self.latest_pose = None
        self.latest_altitude = None
        self.latest_state = None
        
    def camera_callback(self, msg: Image):
        """Handle /main_camera/image_raw topic"""
        try:
            cv_image = self.bridge.imgmsg_to_cv2(msg, desired_encoding='bgr8')
            _, buffer = cv2.imencode('.jpg', cv_image, [cv2.IMWRITE_JPEG_QUALITY, 80])
            self.latest_frame = base64.b64encode(buffer).decode('utf-8')
        except Exception as e:
            logger.error(f"Error processing camera frame: {e}")
    
    def battery_callback(self, msg: BatteryState):
        """Handle /mavros/battery topic"""
        self.latest_battery = {
            'voltage': msg.voltage,
            'percentage': msg.percentage * 100 if msg.percentage > 0 else None,
            'current': msg.current
        }
    
    def pose_callback(self, msg: PoseStamped):
        """Handle /mavros/local_position/pose topic"""
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
        """Handle /rangefinder/range topic"""
        self.latest_altitude = msg.range
    
    def state_callback(self, msg: State):
        """Handle /mavros/state topic"""
        self.latest_state = {
            'armed': msg.armed,
            'connected': msg.connected,
            'mode': msg.mode
        }
    
    def get_frame(self):
        """Get latest camera frame"""
        return self.latest_frame
    
    def get_telemetry(self):
        """Get latest telemetry data"""
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
        """Initialize ROS node and subscribers in separate thread"""
        try:
            logger.info("Initializing ROS node")
            rospy.init_node('drone_parser', anonymous=False)
            
            self.ros_bridge = ROSBridge()
            
            # Subscribe to camera
            rospy.Subscriber('/main_camera/image_raw', Image, 
                           self.ros_bridge.camera_callback, queue_size=1)
            
            # Subscribe to telemetry
            rospy.Subscriber('/mavros/battery', BatteryState,
                           self.ros_bridge.battery_callback, queue_size=1)
            rospy.Subscriber('/mavros/local_position/pose', PoseStamped,
                           self.ros_bridge.pose_callback, queue_size=1)
            rospy.Subscriber('/rangefinder/range', Range,
                           self.ros_bridge.rangefinder_callback, queue_size=1)
            rospy.Subscriber('/mavros/state', State,
                           self.ros_bridge.state_callback, queue_size=1)
            
            logger.info("ROS subscribers initialized")
            
            # Keep ROS spinning in this thread
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
        
        # Start ROS in separate thread
        self.ros_thread = threading.Thread(target=self.init_ros, daemon=True)
        self.ros_thread.start()
        
        # Wait for ROS to initialize
        await asyncio.sleep(2)

        try:
            # Inject ROS bridge into websocket service
            if self.ros_bridge:
                websocket_service.ros_bridge = self.ros_bridge
            
            await websocket_service.connect()
        except asyncio.CancelledError:
            logger.info("Application cancelled")
        finally:
            await self.stop()

    async def stop(self):
        logger.info("Stopping Drone Application")
        self.is_running = False
        
        # Shutdown ROS
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
