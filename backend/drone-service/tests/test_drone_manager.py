import pytest
import uuid
from unittest.mock import AsyncMock
from datetime import datetime
from app.services.drone_manager import DroneManager
from app.models.models import DroneState, DroneStatus, Position


@pytest.fixture
def mock_state_repository():
    repo = AsyncMock()
    repo.get_drone_state = AsyncMock(return_value=None)
    repo.save_drone_state = AsyncMock(return_value=True)
    return repo


@pytest.fixture
def drone_manager(mock_state_repository):
    return DroneManager(mock_state_repository)


@pytest.mark.asyncio
async def test_register_drone(drone_manager):
    drone_id = str(uuid.uuid4())
    
    await drone_manager.register_drone(drone_id)
    
    assert drone_id in drone_manager.registered_drones


@pytest.mark.asyncio
async def test_unregister_drone(drone_manager):
    drone_id = str(uuid.uuid4())
    await drone_manager.register_drone(drone_id)
    
    await drone_manager.unregister_drone(drone_id)
    
    assert drone_id not in drone_manager.registered_drones


@pytest.mark.asyncio
async def test_assign_delivery_to_drone(drone_manager, mock_state_repository):
    drone_id = str(uuid.uuid4())
    delivery_id = str(uuid.uuid4())
    
    state = DroneState(
        drone_id=drone_id,
        status=DroneStatus.IDLE,
        battery_level=80.0,
        current_position=Position(latitude=0.0, longitude=0.0, altitude=0.0),
        speed=0.0,
        last_updated=datetime.now()
    )
    mock_state_repository.get_drone_state.return_value = state
    mock_state_repository.save_drone_state = AsyncMock(return_value=True)
    
    result = await drone_manager.assign_delivery_to_drone(drone_id, delivery_id)
    
    assert result is True
    assert state.current_delivery_id == delivery_id
    assert state.status == DroneStatus.TAKING_OFF


@pytest.mark.asyncio
async def test_release_drone(drone_manager, mock_state_repository):
    drone_id = str(uuid.uuid4())
    
    state = DroneState(
        drone_id=drone_id,
        status=DroneStatus.IN_TRANSIT,
        battery_level=80.0,
        current_position=Position(latitude=0.0, longitude=0.0, altitude=0.0),
        speed=10.0,
        last_updated=datetime.now(),
        current_delivery_id="delivery_123"
    )
    mock_state_repository.get_drone_state.return_value = state
    mock_state_repository.save_drone_state = AsyncMock(return_value=True)
    
    result = await drone_manager.release_drone(drone_id)
    
    assert result is True
    assert state.current_delivery_id is None
    assert state.status == DroneStatus.IDLE


@pytest.mark.asyncio
async def test_get_free_drone_with_sufficient_battery(drone_manager, mock_state_repository):
    drone_id = str(uuid.uuid4())
    await drone_manager.register_drone(drone_id)
    
    state = DroneState(
        drone_id=drone_id,
        status=DroneStatus.IDLE,
        battery_level=80.0,
        current_position=Position(latitude=0.0, longitude=0.0, altitude=0.0),
        speed=0.0,
        last_updated=datetime.now()
    )
    mock_state_repository.get_drone_state.return_value = state
    
    available = await drone_manager.get_free_drone()
    
    assert available == drone_id


@pytest.mark.asyncio
async def test_get_free_drone_low_battery(drone_manager, mock_state_repository):
    drone_id = str(uuid.uuid4())
    await drone_manager.register_drone(drone_id)
    
    state = DroneState(
        drone_id=drone_id,
        status=DroneStatus.IDLE,
        battery_level=20.0,
        current_position=Position(latitude=0.0, longitude=0.0, altitude=0.0),
        speed=0.0,
        last_updated=datetime.now()
    )
    mock_state_repository.get_drone_state.return_value = state
    
    available = await drone_manager.get_free_drone()
    
    assert available is None


@pytest.mark.asyncio
async def test_get_free_drone_with_active_delivery(drone_manager, mock_state_repository):
    drone_id = str(uuid.uuid4())
    await drone_manager.register_drone(drone_id)
    
    state = DroneState(
        drone_id=drone_id,
        status=DroneStatus.IN_TRANSIT,
        battery_level=80.0,
        current_position=Position(latitude=0.0, longitude=0.0, altitude=0.0),
        speed=10.0,
        last_updated=datetime.now(),
        current_delivery_id="delivery_123"
    )
    mock_state_repository.get_drone_state.return_value = state
    
    available = await drone_manager.get_free_drone()
    
    assert available is None


@pytest.mark.asyncio
async def test_get_free_drone_none_registered(drone_manager):
    available = await drone_manager.get_free_drone()
    
    assert available is None
