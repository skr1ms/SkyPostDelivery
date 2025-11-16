from typing import Dict, List, Optional

from pydantic import BaseModel, Field, model_validator
from datetime import datetime


class CellsPayload(BaseModel):
    cells_out: Optional[List[str]] = Field(
        default=None, description="Список UUID внешних ячеек (по порядку)"
    )
    cells_internal: Optional[List[str]] = Field(
        default=None, description="Список UUID внутренних дверей (по порядку)"
    )
    cells: Optional[List[str]] = Field(
        default=None, description="УСТАРЕВШЕЕ поле: список UUID внешних ячеек"
    )
    parcel_automat_id: str = Field(..., description="UUID постамата")

    @model_validator(mode="before")
    @classmethod
    def ensure_cells(cls, values: Dict) -> Dict:
        if "cells_out" not in values or values.get("cells_out") is None:
            legacy = values.get("cells")
            if legacy:
                values["cells_out"] = legacy
        if values.get("cells_out") is None:
            raise ValueError(
                "cells_out field is required (or provide legacy 'cells')")
        if values.get("cells_internal") is None:
            values["cells_internal"] = []
        return values


class QRScanRequest(BaseModel):
    qr_data: str = Field(..., description="JSON строка с данными QR кода")


class QRScanResponse(BaseModel):
    success: bool
    message: str
    cell_ids: List[str]


class ConfirmPickupRequest(BaseModel):
    cell_ids: List[str] = Field(...,
                                description="UUID ячеек, из которых забрали товары")


class ConfirmLoadedRequest(BaseModel):
    order_id: str = Field(..., description="UUID заказа")
    locker_cell_id: str = Field(..., description="UUID ячейки")
    timestamp: datetime = Field(default_factory=datetime.now)


class CellMappingResponse(BaseModel):
    mapping: Dict
    cells_count: int
    internal_mapping: Dict = Field(default_factory=dict)
    internal_cells_count: int = 0


class CellUUIDResponse(BaseModel):
    cell_number: int
    cell_uuid: str


class OpenCellResponse(BaseModel):
    success: bool
    cell_number: int
    cell_uuid: str
    action: str


class PrepareCellRequest(BaseModel):
    cell_id: str = Field(..., description="UUID ячейки для подготовки")


class HealthResponse(BaseModel):
    status: str = "healthy"
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class ServiceInfoResponse(BaseModel):
    service: str
    version: str
    status: str
