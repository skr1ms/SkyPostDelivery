from pydantic import BaseModel, Field
from typing import List
from datetime import datetime


class CellsPayload(BaseModel):
    cells: List[str] = Field(...,
                             description="Список UUID ячеек в порядке их создания")
    parcel_automat_id: str = Field(..., description="UUID постамата")


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
    mapping: dict
    cells_count: int


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
