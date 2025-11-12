#!/usr/bin/env python3

import rospy
import sys
import subprocess
import time
from clover import srv
from std_srvs.srv import Trigger
from sensor_msgs.msg import BatteryState
from geometry_msgs.msg import PoseStamped
from mavros_msgs.msg import State

rospy.init_node('delivery_flight')

get_telemetry = rospy.ServiceProxy('get_telemetry', srv.GetTelemetry)
navigate = rospy.ServiceProxy('navigate', srv.Navigate)
navigate_global = rospy.ServiceProxy('navigate_global', srv.NavigateGlobal)
set_position = rospy.ServiceProxy('set_position', srv.SetPosition)
set_velocity = rospy.ServiceProxy('set_velocity', srv.SetVelocity)
set_attitude = rospy.ServiceProxy('set_attitude', srv.SetAttitude)
set_rates = rospy.ServiceProxy('set_rates', srv.SetRates)
land_srv = rospy.ServiceProxy('land', Trigger)

arrival_pub = rospy.Publisher(
    '/drone/delivery/arrived', PoseStamped, queue_size=10)
drop_ready_pub = rospy.Publisher(
    '/drone/delivery/drop_ready', PoseStamped, queue_size=10)

battery_voltage = 12.0
battery_percentage = 100.0
current_pose = None
mavros_state = None


def battery_callback(data):
    global battery_voltage, battery_percentage
    battery_voltage = data.voltage
    battery_percentage = data.percentage * 100.0


def pose_callback(data):
    global current_pose
    current_pose = data


def state_callback(data):
    global mavros_state
    mavros_state = data


battery_sub = rospy.Subscriber(
    '/mavros/battery', BatteryState, battery_callback, queue_size=10)
pose_sub = rospy.Subscriber(
    '/mavros/local_position/pose', PoseStamped, pose_callback, queue_size=10)
state_sub = rospy.Subscriber(
    '/mavros/state', State, state_callback, queue_size=10)


def navigate_wait(x=0, y=0, z=0, yaw=float('nan'), speed=0.5, frame_id='body', auto_arm=False, tolerance=0.2):
    navigate(x=x, y=y, z=z, yaw=yaw, speed=speed,
             frame_id=frame_id, auto_arm=auto_arm)

    while not rospy.is_shutdown():
        telem = get_telemetry(frame_id='navigate_target')
        if abs(telem.x) < tolerance and abs(telem.y) < tolerance and abs(telem.z) < tolerance:
            break
        rospy.sleep(0.2)


def land_wait():
    land_srv()
    while not rospy.is_shutdown():
        telem = get_telemetry(frame_id='aruco_map')
        if telem.armed:
            rospy.sleep(0.5)
        else:
            break


if len(sys.argv) < 4:
    rospy.logerr("Usage: delivery_flight.py <aruco_id> <x_coord> <y_coord>")
    sys.exit(1)

aruco_id = int(sys.argv[1])
target_x = float(sys.argv[2])
target_y = float(sys.argv[3])

rospy.loginfo(
    f"Starting delivery to ArUco {aruco_id} at ({target_x}, {target_y})")

navigate_wait(z=1.5, frame_id='body', auto_arm=True)
rospy.sleep(2)

navigate_wait(x=target_x, y=target_y, z=1.5, frame_id='aruco_map', speed=0.8)
rospy.sleep(1)

navigate_wait(z=0.5, frame_id=f'aruco_{aruco_id}', speed=0.3)
rospy.sleep(1)

land_wait()
rospy.sleep(2)

if current_pose:
    arrival_msg = PoseStamped()
    arrival_msg.header.stamp = rospy.Time.now()
    arrival_msg.header.frame_id = 'aruco_map'
    arrival_msg.pose = current_pose.pose
    arrival_pub.publish(arrival_msg)
    rospy.loginfo("Published arrival notification")

rospy.loginfo("Waiting for drop confirmation...")

drop_confirmed = False
start_wait = time.time()
timeout = 10

while not rospy.is_shutdown() and not drop_confirmed:
    if time.time() - start_wait > timeout:
        rospy.logerr("Timeout waiting for drop confirmation")
        break

    rospy.sleep(1)

rospy.loginfo("Drop cargo command received (simulated)")

if current_pose:
    drop_msg = PoseStamped()
    drop_msg.header.stamp = rospy.Time.now()
    drop_msg.header.frame_id = 'aruco_map'
    drop_msg.pose = current_pose.pose
    drop_ready_pub.publish(drop_msg)
    rospy.loginfo("Published drop ready notification")

rospy.sleep(10)

rospy.loginfo("Cargo dropped, launching return flight...")

try:
    subprocess.Popen(['python3', '/root/flight_back.py'],
                     stdout=subprocess.PIPE,
                     stderr=subprocess.PIPE)
    rospy.loginfo("Return flight script launched")
except Exception as e:
    rospy.logerr(f"Failed to launch return flight: {e}")

rospy.loginfo("Delivery flight script completed")
