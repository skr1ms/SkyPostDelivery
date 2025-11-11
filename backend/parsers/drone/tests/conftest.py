import pytest
import sys
from pathlib import Path

drone_dir = Path(__file__).parent.parent
sys.path.insert(0, str(drone_dir))


@pytest.fixture
def anyio_backend():
    return 'asyncio'
