import pytest
from unittest.mock import MagicMock, patch
from app.navigation.clover_navigation_controller import CloverNavigationController, FlightState
import tempfile
from pathlib import Path


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


@patch('app.navigation.clover_navigation_controller.CloverAPI')
def test_navigation_controller_init(mock_api_class, temp_map_file):
    controller = CloverNavigationController(temp_map_file)
    assert controller.state == FlightState.IDLE
    assert controller.cruise_altitude == 1.5
    assert controller.cruise_speed == 0.5


@patch('app.navigation.clover_navigation_controller.CloverAPI')
def test_navigation_controller_initialize(mock_api_class, temp_map_file):
    mock_api = MagicMock()
    mock_api.initialize.return_value = True
    mock_api_class.return_value = mock_api
    
    controller = CloverNavigationController(temp_map_file)
    result = controller.initialize()
    
    assert result == True
    assert controller.state == FlightState.IDLE


@patch('app.navigation.clover_navigation_controller.CloverAPI')
def test_navigation_target_marker_validation(mock_api_class, temp_map_file):
    mock_api = MagicMock()
    mock_api.initialize.return_value = True
    mock_api_class.return_value = mock_api
    
    controller = CloverNavigationController(temp_map_file)
    controller.initialize()
    
    assert controller.map_manager.is_valid_marker(52) == True
    assert controller.map_manager.is_valid_marker(999) == False

