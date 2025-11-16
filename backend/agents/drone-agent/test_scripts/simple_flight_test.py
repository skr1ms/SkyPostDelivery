#!/usr/bin/env python3
"""
ArUco map navigation test
Takeoff at (0,0) and fly to (1,1) using aruco_map frame
"""
import rospy
from clover import srv
from std_srvs.srv import Trigger
import tf2_ros

print("=" * 60)
print("ARUCO MAP NAVIGATION TEST")
print("=" * 60)

rospy.init_node('aruco_navigation_test')

get_telemetry = rospy.ServiceProxy('get_telemetry', srv.GetTelemetry)
navigate = rospy.ServiceProxy('navigate', srv.Navigate)
land = rospy.ServiceProxy('land', Trigger)

TAKEOFF_HEIGHT = 2.0
TARGET_X = 1.0  # Координата X маркера 135 в map.txt
TARGET_Y = 1.0  # Координата Y маркера 135 в map.txt

def wait_for_frame(frame_id, timeout=30):
    """Ждём появления системы координат в TF"""
    print(f"Waiting for frame '{frame_id}'...")
    tf_buffer = tf2_ros.Buffer()
    tf_listener = tf2_ros.TransformListener(tf_buffer)
    
    start_time = rospy.Time.now()
    while (rospy.Time.now() - start_time).to_sec() < timeout:
        try:
            # Проверяем, есть ли трансформация от map к искомому фрейму
            tf_buffer.lookup_transform('map', frame_id, rospy.Time(0), rospy.Duration(1.0))
            print(f"Frame '{frame_id}' found!")
            return True
        except (tf2_ros.LookupException, tf2_ros.ConnectivityException, tf2_ros.ExtrapolationException):
            rospy.sleep(0.5)
    
    print(f"WARNING: Frame '{frame_id}' not found after {timeout}s")
    return False

# ===== STAGE 1: TAKEOFF =====
print("\n[Stage 1] Takeoff in body frame")
print(f"Target height: {TAKEOFF_HEIGHT} m")

navigate(x=0, y=0, z=TAKEOFF_HEIGHT, frame_id='body', auto_arm=True)
print("Takeoff command sent, waiting 7 seconds...")
rospy.sleep(7)

telem = get_telemetry(frame_id='body')
print(f"Current altitude: {telem.z:.2f} m")
print(f"Armed: {telem.armed}, Mode: {telem.mode}")

# ===== STAGE 2: WAIT FOR ARUCO_MAP =====
print("\n[Stage 2] Waiting for aruco_map frame...")
print("Camera must see at least one marker from map.txt")

# Проверяем, какой фрейм создан: aruco_map или aruco_map_detected
map_frame = None
if wait_for_frame('aruco_map', timeout=15):
    map_frame = 'aruco_map'
elif wait_for_frame('aruco_map_detected', timeout=5):
    map_frame = 'aruco_map_detected'
    print("Using 'aruco_map_detected' (aruco_vpe is enabled)")
else:
    print("\nERROR: No aruco map frame found!")
    print("Check:")
    print("  1. Camera sees markers")
    print("  2. map.txt is loaded")
    print("  3. aruco_map=true in aruco.launch")
    print("\nAborting, landing...")
    land()
    rospy.sleep(5)
    exit(1)

# ===== STAGE 3: NAVIGATE TO TARGET =====
print(f"\n[Stage 3] Navigate to ({TARGET_X}, {TARGET_Y}) in {map_frame}")

navigate(x=TARGET_X, y=TARGET_Y, z=TAKEOFF_HEIGHT, frame_id=map_frame, speed=0.5)
print("Navigation command sent, flying...")

# ===== STAGE 4: HOVER AND CHECK =====
print("\n[Stage 4] Hovering at target (5 seconds)")
telem = get_telemetry(frame_id=map_frame)
print(f"Final position: x={telem.x:.2f}, y={telem.y:.2f}, z={telem.z:.2f}")

for i in range(5):
    rospy.sleep(1)
    telem = get_telemetry(frame_id=map_frame)
    print(f"  t={i+1}s | pos=({telem.x:.2f}, {telem.y:.2f}, {telem.z:.2f}) | battery={telem.voltage:.2f}V")

# ===== STAGE 5: LANDING =====
print("\n[Stage 5] Landing")
land()
print("Landing command sent, descending...")

for i in range(15):
    rospy.sleep(1)
    telem = get_telemetry(frame_id='body')
    print(f"  t={i+1:02d}s | altitude={telem.z:.2f}m | armed={telem.armed}")
    if telem.z < 0.15 and not telem.armed:
        print("Landed successfully!")
        break

print("\n" + "=" * 60)
print("TEST COMPLETED")
print("=" * 60)