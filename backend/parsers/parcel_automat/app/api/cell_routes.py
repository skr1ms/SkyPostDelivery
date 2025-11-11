from fastapi import APIRouter, HTTPException, Depends, status
from typing import Dict
import logging

from ..models.schemas import (
    CellsPayload,
    CellMappingResponse,
    CellUUIDResponse,
    OpenCellResponse,
    PrepareCellRequest
)
from ..services.cell_management_service import CellManagementService

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/cells", tags=["cells"])


def get_cell_service() -> CellManagementService:
    from ..dependencies import cell_service
    return cell_service


@router.post("/sync", status_code=status.HTTP_200_OK)
async def sync_cells(payload: CellsPayload, service: CellManagementService = Depends(get_cell_service)) -> Dict:
    try:
        if not payload.parcel_automat_id:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="parcel_automat_id is required in request body"
            )

        mapping = service.sync_cells(payload.cells, payload.parcel_automat_id)
        logger.info(f"Cells synced for automat {payload.parcel_automat_id}")
        return {
            "message": "Cells synchronized successfully",
            "cells_count": len(mapping),
            "mapping": mapping
        }
    except Exception as e:
        logger.error(f"Failed to sync cells: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to sync cells: {str(e)}"
        )


@router.get("/mapping", response_model=CellMappingResponse)
async def get_cells_mapping(service: CellManagementService = Depends(get_cell_service)) -> CellMappingResponse:
    try:
        mapping = service.get_mapping()
        return CellMappingResponse(
            mapping=mapping,
            cells_count=len(mapping)
        )
    except Exception as e:
        logger.error(f"Failed to get cells mapping: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to get cells mapping: {str(e)}"
        )


@router.get("/{cell_number}/uuid", response_model=CellUUIDResponse)
async def get_cell_uuid(cell_number: int, service: CellManagementService = Depends(get_cell_service)) -> CellUUIDResponse:
    try:
        cell_uuid = service.get_cell_uuid(cell_number)
        return CellUUIDResponse(
            cell_number=cell_number,
            cell_uuid=cell_uuid
        )
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e)
        )
    except Exception as e:
        logger.error(f"Failed to get cell UUID: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to get cell UUID: {str(e)}"
        )


@router.post("/{cell_number}/open", response_model=OpenCellResponse)
async def open_cell(cell_number: int, service: CellManagementService = Depends(get_cell_service)) -> OpenCellResponse:
    try:
        result = service.open_cell(cell_number)
        return OpenCellResponse(**result)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e)
        )
    except Exception as e:
        logger.error(f"Failed to open cell: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to open cell: {str(e)}"
        )


@router.post("/{cell_number}/close", response_model=OpenCellResponse)
async def close_cell(cell_number: int, service: CellManagementService = Depends(get_cell_service)) -> OpenCellResponse:
    try:
        result = service.close_cell(cell_number)
        return OpenCellResponse(**result)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e)
        )
    except Exception as e:
        logger.error(f"Failed to close cell: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to close cell: {str(e)}"
        )


@router.get("/{cell_number}/status")
async def get_cell_status(cell_number: int, service: CellManagementService = Depends(get_cell_service)) -> Dict:
    try:
        status_str = service.get_cell_status(cell_number)
        cell_uuid = service.get_cell_uuid(cell_number)
        return {
            "cell_number": cell_number,
            "cell_uuid": cell_uuid,
            "status": status_str
        }
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e)
        )
    except Exception as e:
        logger.error(f"Failed to get cell status: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to get cell status: {str(e)}"
        )


@router.get("/count")
async def get_cells_count(service: CellManagementService = Depends(get_cell_service)) -> Dict:
    try:
        count = service.arduino.get_cells_count()
        return {
            "cells_count": count,
            "mapped_cells": len(service.get_mapping())
        }
    except Exception as e:
        logger.error(f"Failed to get cells count: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to get cells count: {str(e)}"
        )


@router.post("/prepare")
async def prepare_cell(request: PrepareCellRequest, service: CellManagementService = Depends(get_cell_service)) -> Dict:
    try:
        cell_num = service.get_cell_number(request.cell_id)
        if cell_num is None:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Cell {request.cell_id} not found"
            )

        result = await service.open_cells_by_uuids([request.cell_id])

        if result["success"]:
            return {
                "success": True,
                "message": "Cell opened successfully",
                "cell_number": cell_num
            }
        else:
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"Failed to open cell: {result.get('message', 'Unknown error')}"
            )
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to prepare cell: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=str(e)
        )
