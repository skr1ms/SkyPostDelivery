import json
from pathlib import Path
from typing import Dict, Optional
import logging

logger = logging.getLogger(__name__)


class CellMappingRepository:
    def __init__(self, mapping_file: Path):
        self.mapping_file = mapping_file
        self._ensure_data_directory()

    def _ensure_data_directory(self):
        self.mapping_file.parent.mkdir(parents=True, exist_ok=True)

    def load_mapping(self) -> Dict[str, Dict[str, str]]:
        if not self.mapping_file.exists():
            logger.info(f"Mapping file not found: {self.mapping_file}")
            return {}
        try:
            with open(self.mapping_file, 'r', encoding='utf-8') as f:
                mapping = json.load(f)
                logger.info(f"Loaded mapping with {len(mapping)} cells")
                return mapping
        except Exception as e:
            logger.error(f"Failed to load mapping: {e}")
            return {}

    def save_mapping(self, mapping: Dict[str, Dict[str, str]]) -> None:
        try:
            with open(self.mapping_file, 'w', encoding='utf-8') as f:
                json.dump(mapping, f, indent=2, ensure_ascii=False)
            logger.info(f"Saved mapping with {len(mapping)} cells")
        except Exception as e:
            logger.error(f"Failed to save mapping: {e}")
            raise

    def get_cell_data(self, cell_number: int) -> Optional[Dict[str, str]]:
        mapping = self.load_mapping()
        return mapping.get(str(cell_number))

    def get_cell_uuid(self, cell_number: int) -> Optional[str]:
        cell_data = self.get_cell_data(cell_number)
        return cell_data.get("cell_uuid") if cell_data else None

    def get_parcel_automat_id(self) -> Optional[str]:
        mapping = self.load_mapping()
        if mapping:
            first_cell = next(iter(mapping.values()), None)
            return first_cell.get("parcel_automat_id") if first_cell else None
        return None

    def get_cell_number(self, cell_uuid: str) -> Optional[int]:
        mapping = self.load_mapping()
        for cell_num, cell_data in mapping.items():
            if cell_data.get("cell_uuid") == cell_uuid:
                return int(cell_num)
        return None

    def clear_mapping(self) -> None:
        if self.mapping_file.exists():
            self.mapping_file.unlink()
            logger.info("Mapping file deleted")
