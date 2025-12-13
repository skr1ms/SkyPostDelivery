#!/usr/bin/env python3

import rospy
import sys
from clover import srv
from std_srvs.srv import Trigger
from sensor_msgs.msg import BatteryState
from geometry_msgs.msg import PoseStamped
from mavros_msgs.msg import State

rospy.init_node("flight_back")

get_telemetry = rospy.ServiceProxy("get_telemetry", srv.GetTelemetry)
navigate = rospy.ServiceProxy("navigate", srv.Navigate)
land_srv = rospy.ServiceProxy("land", Trigger)

home_pub = rospy.Publisher("/drone/delivery/home_arrived", PoseStamped, queue_size=10)

battery_voltage = 12.0
battery_percentage = 100.0
current_pose = None
mavros_state = None


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


battery_sub = rospy.Subscriber(
    "/mavros/battery", BatteryState, battery_callback, queue_size=10
)
pose_sub = rospy.Subscriber(
    "/mavros/local_position/pose", PoseStamped, pose_callback, queue_size=10
)
state_sub = rospy.Subscriber("/mavros/state", State, state_callback, queue_size=10)


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


home_aruco_id = 131
home_x = 0.0
home_y = 0.0

if len(sys.argv) >= 2:
    home_aruco_id = int(sys.argv[1])
if len(sys.argv) >= 4:
    home_x = float(sys.argv[2])
    home_y = float(sys.argv[3])

rospy.loginfo("=" * 60)
rospy.loginfo("RETURN TO BASE FLIGHT STARTED")
rospy.loginfo(f"Home: ArUco {home_aruco_id} at ({home_x}, {home_y})")
rospy.loginfo(f"Battery: {battery_percentage:.1f}%")
rospy.loginfo("=" * 60)

telem = get_telemetry(frame_id="aruco_map")
if telem.z < 0.5:
    rospy.loginfo("Step 1: Taking off to cruise altitude")
    navigate_wait(z=1.5, frame_id="body", auto_arm=True)
    rospy.sleep(2)
else:
    rospy.loginfo("Step 1: Already airborne, proceeding to home")

rospy.loginfo(f"Step 2: Flying to home position ({home_x}, {home_y})")
navigate_wait(x=home_x, y=home_y, z=1.5, frame_id="aruco_map", speed=0.8)
rospy.sleep(1)

rospy.loginfo(f"Step 3: Descending above home ArUco {home_aruco_id}")
navigate_wait(z=0.5, frame_id=f"aruco_{home_aruco_id}", speed=0.3)
rospy.sleep(1)

rospy.loginfo("Step 4: Landing at home base")
land_wait()
rospy.sleep(2)

rospy.loginfo("Step 5: Publishing home arrival notification")
if current_pose:
    home_msg = PoseStamped()
    home_msg.header.stamp = rospy.Time.now()
    home_msg.header.frame_id = "aruco_map"
    home_msg.pose = current_pose.pose
    home_pub.publish(home_msg)
    rospy.loginfo("Home arrival notification published")
else:
    rospy.logwarn("No pose data available")

rospy.loginfo("=" * 60)
rospy.loginfo("RETURN TO BASE COMPLETED")
rospy.loginfo(f"Battery remaining: {battery_percentage:.1f}%")
rospy.loginfo("=" * 60)
