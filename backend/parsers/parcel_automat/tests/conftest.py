import pytest
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))


@pytest.fixture(scope="session")
def test_orchestrator_url():
    return "http://test-orchestrator:8080/api/v1"


@pytest.fixture(scope="session")
def test_qr_data():
    return '{"order_id": "test-order-123", "type": "pickup"}'


@pytest.fixture(scope="session")
def test_cell_ids():
    return ["cell-uuid-1", "cell-uuid-2", "cell-uuid-3"]


@pytest.fixture
def sample_qr_response():
    from app.models.schemas import QRScanResponse
    return QRScanResponse(
        success=True,
        message="QR code validated successfully",
        cell_ids=["cell-1", "cell-2"]
    )
