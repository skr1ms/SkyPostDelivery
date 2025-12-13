import sys
from unittest.mock import MagicMock

sys.modules['numpy'] = MagicMock()
sys.modules['cv2'] = MagicMock()
sys.modules['rospy'] = MagicMock()
sys.modules['sensor_msgs'] = MagicMock()
sys.modules['sensor_msgs.msg'] = MagicMock()
sys.modules['geometry_msgs'] = MagicMock()
sys.modules['geometry_msgs.msg'] = MagicMock()
sys.modules['mavros_msgs'] = MagicMock()
sys.modules['mavros_msgs.msg'] = MagicMock()
sys.modules['std_msgs'] = MagicMock()
sys.modules['std_msgs.msg'] = MagicMock()
sys.modules['cv_bridge'] = MagicMock()

mock_clover = MagicMock()
mock_clover.srv = MagicMock()
sys.modules['clover'] = mock_clover
sys.modules['clover.srv'] = mock_clover.srv

sys.modules['pigpio'] = MagicMock()
