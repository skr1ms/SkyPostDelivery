from typing import List, Dict, Optional, Tuple
import logging
from ..repositories.cell_mapping_repository import CellMappingRepository
from ..hardware.arduino_controller import ArduinoController
from ..hardware.display_controller import DisplayController

logger = logging.getLogger(__name__)


class CellManagementService:
    def __init__(
        self,
        cell_repo: CellMappingRepository,
        internal_repo: Optional[CellMappingRepository],
        arduino_controller: ArduinoController,
        display: Optional[DisplayController] = None,
    ):
        self.cell_repo = cell_repo
        self.internal_repo = internal_repo
        self.arduino = arduino_controller
        self.display = display

    @staticmethod
    def _build_mapping(cell_uuids: List[str], parcel_automat_id: str) -> Dict[str, Dict[str, str]]:
        mapping: Dict[str, Dict[str, str]] = {}
        for idx, cell_uuid in enumerate(cell_uuids, start=1):
            mapping[str(idx)] = {
                "cell_uuid": cell_uuid,
                "parcel_automat_id": parcel_automat_id,
            }
        return mapping

    def sync_cells(
        self,
        cells_out: List[str],
        cells_internal: Optional[List[str]],
        parcel_automat_id: str,
    ) -> Dict[str, Dict[str, Dict[str, str]]]:
        cells_internal = cells_internal or []

        external_mapping = self._build_mapping(cells_out, parcel_automat_id)
        self.cell_repo.save_mapping(external_mapping)
        logger.info(
            f"Synchronized {len(cells_out)} external cells for automat {parcel_automat_id}:"
        )
        for cell_num, cell_data in external_mapping.items():
            logger.info(f"   EXT Cell {cell_num} -> {cell_data['cell_uuid']}")

        internal_mapping: Dict[str, Dict[str, str]] = {}
        if self.internal_repo is not None:
            internal_mapping = self._build_mapping(
                cells_internal, parcel_automat_id)
            self.internal_repo.save_mapping(internal_mapping)
            logger.info(
                f"Synchronized {len(cells_internal)} internal cells for automat {parcel_automat_id}:"
            )
            for cell_num, cell_data in internal_mapping.items():
                logger.info(
                    f"   INT Door {cell_num} -> {cell_data['cell_uuid']}")

        return {
            "external": external_mapping,
            "internal": internal_mapping,
        }

    def get_mapping(self) -> Dict[str, Dict[str, str]]:
        return self.cell_repo.load_mapping()

    def get_internal_mapping(self) -> Dict[str, Dict[str, str]]:
        if not self.internal_repo:
            return {}
        return self.internal_repo.load_mapping()

    def get_cell_uuid(self, cell_number: int) -> str:
        cell_uuid = self.cell_repo.get_cell_uuid(cell_number)
        if not cell_uuid:
            raise ValueError(f"Cell {cell_number} not found in mapping")
        return cell_uuid

    def get_internal_uuid(self, door_number: int) -> str:
        if not self.internal_repo:
            raise ValueError(
                "Internal door mapping repository is not configured")
        door_uuid = self.internal_repo.get_cell_uuid(door_number)
        if not door_uuid:
            raise ValueError(
                f"Internal door {door_number} not found in mapping")
        return door_uuid

    def get_parcel_automat_id(self) -> str:
        automat_id = self.cell_repo.get_parcel_automat_id()
        if not automat_id and self.internal_repo:
            automat_id = self.internal_repo.get_parcel_automat_id()
        if not automat_id:
            raise ValueError(
                "Parcel automat ID not found. Please sync cells from orchestrator first (POST /api/cells/sync)"
            )
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
            "arduino_response": arduino_response,
            "type": "external",
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

    def open_internal_door(self, door_number: int) -> Dict:
        door_uuid = self.get_internal_uuid(door_number)

        arduino_response = self.arduino.open_internal_door(door_number)
        logger.info(f"Opened internal door {door_number} (UUID: {door_uuid})")

        return {
            "success": True,
            "door_number": door_number,
            "cell_number": door_number,
            "cell_uuid": door_uuid,
            "action": "internal_opened",
            "arduino_response": arduino_response,
            "type": "internal",
        }

    def _find_cell_number_by_uuid(self, cell_uuid: str) -> Tuple[Optional[int], str]:
        number = self.cell_repo.get_cell_number(cell_uuid)
        if number:
            return number, "external"
        if self.internal_repo:
            internal_number = self.internal_repo.get_cell_number(cell_uuid)
            if internal_number:
                return internal_number, "internal"
        return None, "unknown"

    def open_cells_by_uuids(self, cell_uuids: List[str]) -> List[Dict]:
        results = []
        for cell_uuid in cell_uuids:
            cell_number, cell_type = self._find_cell_number_by_uuid(cell_uuid)
            if cell_number:
                try:
                    if cell_type == "external":
                        result = self.open_cell(cell_number)
                    else:
                        result = self.open_internal_door(cell_number)
                    results.append(result)
                except Exception as e:
                    logger.error(
                        f"Failed to open {cell_type} cell {cell_number}: {e}")
                    results.append(
                        {
                            "success": False,
                            "cell_uuid": cell_uuid,
                            "error": str(e),
                            "type": cell_type,
                        }
                    )
            else:
                logger.warning(f"Cell UUID not found in mapping: {cell_uuid}")
                results.append(
                    {
                        "success": False,
                        "cell_uuid": cell_uuid,
                        "error": "cell_uuid_not_found",
                        "type": "unknown",
                    }
                )
        return results

    def clear_mapping(self) -> None:
        self.cell_repo.clear_mapping()
        if self.internal_repo:
            self.internal_repo.clear_mapping()
        logger.info("Cells mapping cleared")

    def get_cell_number(self, cell_uuid: str) -> int:
        cell_number = self.cell_repo.get_cell_number(cell_uuid)
        if not cell_number:
            raise ValueError(f"Cell UUID {cell_uuid} not found in mapping")
        return cell_number

    def get_internal_door_number(self, door_uuid: str) -> int:
        if not self.internal_repo:
            raise ValueError(
                "Internal door mapping repository is not configured")
        door_number = self.internal_repo.get_cell_number(door_uuid)
        if not door_number:
            raise ValueError(
                f"Internal door UUID {door_uuid} not found in mapping")
        return door_number
