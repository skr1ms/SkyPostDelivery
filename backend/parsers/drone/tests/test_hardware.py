import pytest
from app.navigation.aruco_map_manager import ArucoMapManager
from pathlib import Path
import tempfile


@pytest.fixture
def temp_map_file():
    content = """# Test map
0 0.0 0.0 0 0.33
1 1.0 0.0 0 0.33
52 5.0 6.0 0 0.33
"""
    with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.txt') as f:
        f.write(content)
        f.flush()
        yield f.name
    Path(f.name).unlink(missing_ok=True)


def test_aruco_map_manager_load(temp_map_file):
    manager = ArucoMapManager(temp_map_file)
    assert manager.load_map() == True
    assert len(manager.markers) == 3
    assert 52 in manager.markers


def test_aruco_map_manager_get_marker_position(temp_map_file):
    manager = ArucoMapManager(temp_map_file)
    manager.load_map()
    
    pos = manager.get_marker_position(52)
    assert pos is not None
    assert pos[0] == 5.0
    assert pos[1] == 6.0


def test_aruco_map_manager_get_marker_xy(temp_map_file):
    manager = ArucoMapManager(temp_map_file)
    manager.load_map()
    
    xy = manager.get_marker_xy(1)
    assert xy == (1.0, 0.0)


def test_aruco_map_manager_is_valid_marker(temp_map_file):
    manager = ArucoMapManager(temp_map_file)
    manager.load_map()
    
    assert manager.is_valid_marker(0) == True
    assert manager.is_valid_marker(999) == False


def test_aruco_map_manager_find_nearest(temp_map_file):
    manager = ArucoMapManager(temp_map_file)
    manager.load_map()
    
    nearest = manager.find_nearest_marker(0.5, 0.0)
    assert nearest in [0, 1]


def test_aruco_map_manager_distance_between(temp_map_file):
    manager = ArucoMapManager(temp_map_file)
    manager.load_map()
    
    dist = manager.get_distance_between_markers(0, 1)
    assert dist == pytest.approx(1.0, rel=0.1)


def test_aruco_map_manager_invalid_file():
    manager = ArucoMapManager("nonexistent.txt")
    assert manager.load_map() == False
