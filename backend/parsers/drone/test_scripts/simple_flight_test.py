#!/usr/bin/env python3
"""
Test flight script for a Clover drone navigating by ArUco markers.
Takes off from marker 131 (0,0) and flies to marker 135 (1,1).

Usage:
    rosrun clover simple_flight_test.py
    or
    python3 simple_flight_test.py
"""

import rospy
from clover import srv
from std_srvs.srv import Trigger

print("=" * 60)
print("CLOVER ARUCO NAVIGATION TEST")
print("Marker 131 (0,0) -> Marker 135 (1,1)")
print("=" * 60)

rospy.init_node('aruco_navigation_test')

get_telemetry = rospy.ServiceProxy('get_telemetry', srv.GetTelemetry)
navigate = rospy.ServiceProxy('navigate', srv.Navigate)
land = rospy.ServiceProxy('land', Trigger)

print("\nROS node initialised")
print("Service proxies connected")

TAKEOFF_HEIGHT = 2
CRUISE_HEIGHT = 2
LANDING_HEIGHT = 0.8
START_MARKER = 131
TARGET_MARKER = 135

print(f"\nStage 1: takeoff over marker {START_MARKER} (0,0)")
print(f"Climb to {TAKEOFF_HEIGHT} m")
navigate(x=0, y=0, z=TAKEOFF_HEIGHT, frame_id=f'aruco_{START_MARKER}', auto_arm=True)
print("Takeoff command sent")

print("Waiting for altitude (7 seconds)...")
rospy.sleep(7)

telem = get_telemetry(frame_id=f'aruco_{START_MARKER}')
print(f"Position relative to marker {START_MARKER}:")
print(f"   x={telem.x:.2f} m, y={telem.y:.2f} m, z={telem.z:.2f} m")
print(f"   Battery={telem.voltage:.2f} V")

print("\nStabilising (3 seconds)...")
rospy.sleep(3)

print(f"\nStage 2: fly to marker {TARGET_MARKER} (1,1)")
print(f"Cruise altitude {CRUISE_HEIGHT} m")
navigate(x=0, y=0, z=CRUISE_HEIGHT, frame_id=f'aruco_{TARGET_MARKER}', speed=0.5)
print("Navigation command sent")

print("Transit to target (15 seconds)...")
for i in range(15):
    rospy.sleep(1)
    telem = get_telemetry(frame_id=f'aruco_{TARGET_MARKER}')
    print(f"   {i + 1:02d}/15 s | x={telem.x:.2f} m, y={telem.y:.2f} m, z={telem.z:.2f} m")

telem = get_telemetry(frame_id=f'aruco_{TARGET_MARKER}')
print(f"\nFinal position over marker {TARGET_MARKER}:")
print(f"   x={telem.x:.2f} m, y={telem.y:.2f} m, z={telem.z:.2f} m")
print(f"   Battery={telem.voltage:.2f} V")

print(f"\nStage 3: hover over marker {TARGET_MARKER} (5 seconds)")
for i in range(5):
    rospy.sleep(1)
    telem = get_telemetry(frame_id=f'aruco_{TARGET_MARKER}')
    print(f"   {i + 1:02d}/05 s | x={telem.x:.2f} m, y={telem.y:.2f} m, z={telem.z:.2f} m | Battery={telem.voltage:.2f} V")

print("Hover complete")

print(f"\nStage 4: precision landing on marker {TARGET_MARKER}")
print(f"Descend to {LANDING_HEIGHT} m")
navigate(x=0, y=0, z=LANDING_HEIGHT, frame_id=f'aruco_{TARGET_MARKER}', speed=0.3)
rospy.sleep(5)

telem = get_telemetry(frame_id=f'aruco_{TARGET_MARKER}')
print(f"Height before landing: {telem.z:.2f} m")

print("Smooth descent...")
for height in [0.6, 0.4, 0.3]:
    navigate(x=0, y=0, z=height, frame_id=f'aruco_{TARGET_MARKER}', speed=0.2)
    rospy.sleep(3)
    telem = get_telemetry(frame_id=f'aruco_{TARGET_MARKER}')
    print(f"   Commanded {height:.1f} m | Current {telem.z:.2f} m")

print("Initiating final landing...")
land()
print("Landing command sent")

print("Monitoring descent...")
for i in range(15):
    rospy.sleep(1)
    telem = get_telemetry(frame_id='body')
    print(f"   {i + 1:02d}/15 s | Height={telem.z:.2f} m")
    if telem.z < 0.2:
        print("Drone landed on target marker")
        break

print("\n" + "=" * 60)
print("Final statistics")
print("=" * 60)
final_telem = get_telemetry()
print(f"Height: {final_telem.z:.2f} m")
print(f"Battery: {final_telem.voltage:.2f} V")
print(f"Armed: {final_telem.armed}")
print(f"Route: marker {START_MARKER} (0,0) -> marker {TARGET_MARKER} (1,1)")

print("\n" + "=" * 60)
print("Navigation test finished successfully")
print("=" * 60)
