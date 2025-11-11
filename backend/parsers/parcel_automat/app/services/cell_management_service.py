from typing import List, Dict, Optional
import logging
from ..repositories.cell_mapping_repository import CellMappingRepository
from ..hardware.arduino_controller import ArduinoController
from ..hardware.display_controller import DisplayController

logger = logging.getLogger(__name__)


class CellManagementService:
    def __init__(self, cell_repo: CellMappingRepository, arduino_controller: ArduinoController, display: Optional[DisplayController] = None):
        self.cell_repo = cell_repo
        self.arduino = arduino_controller
        self.display = display

    def sync_cells(self, cell_uuids: List[str], parcel_automat_id: str) -> Dict[str, Dict[str, str]]:
        mapping = {}
        for idx, cell_uuid in enumerate(cell_uuids, start=1):
            mapping[str(idx)] = {
                "cell_uuid": cell_uuid,
                "parcel_automat_id": parcel_automat_id
            }
        self.cell_repo.save_mapping(mapping)
        logger.info(
            f"Synchronized {len(cell_uuids)} cells for automat {parcel_automat_id}:")
        for cell_num, cell_data in mapping.items():
            logger.info(f"   Cell {cell_num} -> {cell_data['cell_uuid']}")
        return mapping

    def get_mapping(self) -> Dict[str, Dict[str, str]]:
        return self.cell_repo.load_mapping()

    def get_cell_uuid(self, cell_number: int) -> str:
        cell_uuid = self.cell_repo.get_cell_uuid(cell_number)
        if not cell_uuid:
            raise ValueError(f"Cell {cell_number} not found in mapping")
        return cell_uuid

    def get_parcel_automat_id(self) -> str:
        automat_id = self.cell_repo.get_parcel_automat_id()
        if not automat_id:
            raise ValueError(
                "Parcel automat ID not found. Please sync cells from orchestrator first (POST /api/cells/sync)")
        return automat_id

    def open_cell(self, cell_number: int, order_number: str = None) -> Dict:
        cell_uuid = self.get_cell_uuid(cell_number)

        if self.display:
            self.display.show_cell_opening(cell_number, order_number)

        arduino_response = self.arduino.open_cell(cell_number)
        logger.info(f"Opened cell {cell_number} (UUID: {cell_uuid})")

        if self.display:
            self.display.show_cell_opened(cell_number)

        return {
            "success": True,
            "cell_number": cell_number,
            "cell_uuid": cell_uuid,
            "action": "opened",
            "arduino_response": arduino_response
        }

    def close_cell(self, cell_number: int) -> Dict:
        cell_uuid = self.get_cell_uuid(cell_number)

        if self.display:
            self.display.show_please_close()

        arduino_response = self.arduino.close_cell(cell_number)
        logger.info(f"Closed cell {cell_number} (UUID: {cell_uuid})")

        if self.display:
            self.display.show_cell_closed()

        return {
            "success": True,
            "cell_number": cell_number,
            "cell_uuid": cell_uuid,
            "action": "closed",
            "arduino_response": arduino_response
        }

    def get_cell_status(self, cell_number: int) -> str:
        status = self.arduino.get_cell_status(cell_number)
        logger.info(f"Cell {cell_number} status: {status}")
        return status

    def open_cells_by_uuids(self, cell_uuids: List[str]) -> List[Dict]:
        results = []
        for cell_uuid in cell_uuids:
            cell_number = self.cell_repo.get_cell_number(cell_uuid)
            if cell_number:
                try:
                    result = self.open_cell(cell_number)
                    results.append(result)
                except Exception as e:
                    logger.error(f"Failed to open cell {cell_number}: {e}")
                    results.append({
                        "success": False,
                        "cell_uuid": cell_uuid,
                        "error": str(e)
                    })
            else:
                logger.warning(f"Cell UUID not found in mapping: {cell_uuid}")
        return results

    def clear_mapping(self) -> None:
        self.cell_repo.clear_mapping()
        logger.info("Cells mapping cleared")

    def get_cell_number(self, cell_uuid: str) -> int:
        cell_number = self.cell_repo.get_cell_number(cell_uuid)
        if not cell_number:
            raise ValueError(f"Cell UUID {cell_uuid} not found in mapping")
        return cell_number
