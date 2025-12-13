#!/usr/bin/env python3

import rospy
import sys
import subprocess
import time
from clover import srv
from std_srvs.srv import Trigger
from std_msgs.msg import Bool
from sensor_msgs.msg import BatteryState
from geometry_msgs.msg import PoseStamped
from mavros_msgs.msg import State

try:
    import pigpio

    PIGPIO_AVAILABLE = True
except ImportError:
    PIGPIO_AVAILABLE = False
    rospy.logwarn("pigpio not available, servo control will be simulated")

rospy.init_node("delivery_flight")

get_telemetry = rospy.ServiceProxy("get_telemetry", srv.GetTelemetry)
navigate = rospy.ServiceProxy("navigate", srv.Navigate)
navigate_global = rospy.ServiceProxy("navigate_global", srv.NavigateGlobal)
set_position = rospy.ServiceProxy("set_position", srv.SetPosition)
set_velocity = rospy.ServiceProxy("set_velocity", srv.SetVelocity)
set_attitude = rospy.ServiceProxy("set_attitude", srv.SetAttitude)
set_rates = rospy.ServiceProxy("set_rates", srv.SetRates)
land_srv = rospy.ServiceProxy("land", Trigger)

arrival_pub = rospy.Publisher("/drone/delivery/arrived", PoseStamped, queue_size=10)
drop_ready_pub = rospy.Publisher(
    "/drone/delivery/drop_ready", PoseStamped, queue_size=10
)

battery_voltage = 12.0
battery_percentage = 100.0
current_pose = None
mavros_state = None
drop_confirmed = False


def battery_callback(data):
    global battery_voltage, battery_percentage
    battery_voltage = data.voltage
    battery_percentage = data.percentage * 100.0 if data.percentage > 0 else 100.0


def pose_callback(data):
    global current_pose
    current_pose = data


def state_callback(data):
    global mavros_state
    mavros_state = data


def drop_confirm_callback(data):
    global drop_confirmed
    if data.data:
        drop_confirmed = True
        rospy.loginfo("Drop confirmation received!")


battery_sub = rospy.Subscriber(
    "/mavros/battery", BatteryState, battery_callback, queue_size=10
)
pose_sub = rospy.Subscriber(
    "/mavros/local_position/pose", PoseStamped, pose_callback, queue_size=10
)
state_sub = rospy.Subscriber("/mavros/state", State, state_callback, queue_size=10)
drop_confirm_sub = rospy.Subscriber(
    "/drone/delivery/drop_confirm", Bool, drop_confirm_callback, queue_size=10
)


def navigate_wait(
    x=0,
    y=0,
    z=0,
    yaw=float("nan"),
    speed=0.5,
    frame_id="body",
    auto_arm=False,
    tolerance=0.2,
):
    navigate(x=x, y=y, z=z, yaw=yaw, speed=speed, frame_id=frame_id, auto_arm=auto_arm)
    while not rospy.is_shutdown():
        telem = get_telemetry(frame_id="navigate_target")
        if (
            abs(telem.x) < tolerance
            and abs(telem.y) < tolerance
            and abs(telem.z) < tolerance
        ):
            break
        rospy.sleep(0.2)


def land_wait():
    land_srv()
    while not rospy.is_shutdown():
        telem = get_telemetry(frame_id="aruco_map")
        if telem.armed:
            rospy.sleep(0.5)
        else:
            break


def drop_cargo():
    SERVO_PIN = 13

    if not PIGPIO_AVAILABLE:
        rospy.logwarn("pigpio not available, simulating cargo drop")
        rospy.loginfo("Opening cargo bay (SIMULATED)")
        rospy.sleep(2)
        rospy.loginfo("Cargo bay opened and package released (SIMULATED)")
        return

    try:
        rospy.loginfo("Initializing servo controller (pigpio)")
        pi = pigpio.pi()

        if not pi.connected:
            rospy.logerr("Failed to connect to pigpiod daemon")
            rospy.logwarn("Make sure pigpiod is running: sudo systemctl start pigpiod")
            rospy.loginfo("Simulating cargo drop instead")
            rospy.sleep(2)
            return

        pi.set_mode(SERVO_PIN, pigpio.OUTPUT)

        def set_angle(angle):
            if angle < 0:
                angle = 0
            elif angle > 180:
                angle = 180
            pulse_width = 500 + (angle / 180.0) * 2000
            pi.set_servo_pulsewidth(SERVO_PIN, pulse_width)

        rospy.loginfo("Closing cargo bay door (0°)")
        set_angle(0)
        rospy.sleep(1)

        rospy.loginfo("Opening cargo bay door (180°)")
        set_angle(180)
        rospy.sleep(2)

        rospy.loginfo("Cargo bay opened, package released!")

        rospy.sleep(1)

        pi.set_servo_pulsewidth(SERVO_PIN, 0)
        pi.stop()

    except Exception as e:
        rospy.logerr(f"Error controlling servo: {e}")
        rospy.loginfo("Simulating cargo drop instead")
        rospy.sleep(2)


if len(sys.argv) < 2:
    rospy.logerr("Usage: delivery_flight.py <target_aruco_id> [home_aruco_id]")
    rospy.logerr("Example: delivery_flight.py 135 131")
    sys.exit(1)

target_aruco_id = int(sys.argv[1])
home_aruco_id = int(sys.argv[2]) if len(sys.argv) > 2 else 131

rospy.loginfo("=" * 60)
rospy.loginfo("DELIVERY FLIGHT MISSION STARTED")
rospy.loginfo(f"Target: ArUco {target_aruco_id}")
rospy.loginfo(f"Home: ArUco {home_aruco_id}")
rospy.loginfo(f"Battery: {battery_percentage:.1f}%")
rospy.loginfo("=" * 60)

rospy.loginfo("Step 1: Taking off to cruise altitude (1.5m)")
navigate_wait(z=1.5, frame_id="body", auto_arm=True)
rospy.sleep(2)

rospy.loginfo(f"Step 2: Flying to ArUco marker {target_aruco_id}")
navigate_wait(
    x=0, y=0, z=1.5, frame_id=f"aruco_{target_aruco_id}", speed=0.8, tolerance=0.3
)
rospy.sleep(1)

rospy.loginfo(f"Step 3: Descending above ArUco {target_aruco_id}")
navigate_wait(z=0.5, frame_id=f"aruco_{target_aruco_id}", speed=0.3)
rospy.sleep(1)

rospy.loginfo("Step 4: Landing on target parcel automat")
land_wait()
rospy.sleep(2)

rospy.loginfo("Step 5: Publishing arrival notification to backend")
if current_pose:
    arrival_msg = PoseStamped()
    arrival_msg.header.stamp = rospy.Time.now()
    arrival_msg.header.frame_id = "aruco_map"
    arrival_msg.pose = current_pose.pose
    arrival_pub.publish(arrival_msg)
    rospy.loginfo("Arrival notification published")
else:
    rospy.logwarn("No pose data available, skipping arrival notification")

rospy.loginfo("Step 6: Waiting for drop confirmation from backend...")

start_wait = time.time()
timeout = 60

while not rospy.is_shutdown() and not drop_confirmed:
    elapsed = time.time() - start_wait

    if elapsed > timeout:
        rospy.logerr("Timeout waiting for drop confirmation")
        rospy.loginfo("Proceeding with emergency return to base")
        try:
            subprocess.Popen(
                ["python3", "/root/flight_back.py"],
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
            )
            rospy.loginfo("Emergency return flight launched")
        except Exception as e:
            rospy.logerr(f"Failed to launch return flight: {e}")
        sys.exit(1)

    if int(elapsed) % 10 == 0 and int(elapsed) > 0:
        rospy.loginfo(f"Still waiting... ({int(elapsed)}s elapsed)")

    rospy.sleep(0.5)

rospy.loginfo("Drop confirmation received!")

rospy.loginfo("Step 7: Publishing drop ready notification")
if current_pose:
    drop_msg = PoseStamped()
    drop_msg.header.stamp = rospy.Time.now()
    drop_msg.header.frame_id = "aruco_map"
    drop_msg.pose = current_pose.pose
    drop_ready_pub.publish(drop_msg)
    rospy.loginfo("Drop ready notification published")

rospy.loginfo("Step 8: Dropping cargo")
drop_cargo()

rospy.loginfo("Step 9: Waiting 10 seconds post-drop for stability")
rospy.sleep(10)

rospy.loginfo("Step 10: Launching return flight to base")
try:
    subprocess.Popen(
        ["python3", "/root/flight_back.py"],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    rospy.loginfo("Return flight script launched successfully")
except Exception as e:
    rospy.logerr(f"Failed to launch return flight: {e}")
    sys.exit(1)

rospy.loginfo("=" * 60)
rospy.loginfo("DELIVERY FLIGHT MISSION COMPLETED")
rospy.loginfo(f"Battery remaining: {battery_percentage:.1f}%")
rospy.loginfo("=" * 60)
