from fastapi import APIRouter, HTTPException, Depends, status
from typing import Dict
import logging

from ..models.schemas import (
    CellsPayload,
    CellMappingResponse,
    CellUUIDResponse,
    OpenCellResponse,
    PrepareCellRequest,
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

        sync_result = service.sync_cells(
            cells_out=payload.cells_out,
            cells_internal=payload.cells_internal,
            parcel_automat_id=payload.parcel_automat_id,
        )
        external_mapping = sync_result.get("external", {})
        internal_mapping = sync_result.get("internal", {})

        logger.info(f"Cells synced for automat {payload.parcel_automat_id}")
        return {
            "message": "Cells synchronized successfully",
            "cells_count": len(external_mapping),
            "mapping": external_mapping,
            "internal_cells_count": len(internal_mapping),
            "internal_mapping": internal_mapping,
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
        internal_mapping = service.get_internal_mapping()
        return CellMappingResponse(
            mapping=mapping,
            cells_count=len(mapping),
            internal_mapping=internal_mapping,
            internal_cells_count=len(internal_mapping),
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
        internal_count = service.arduino.get_internal_cells_count()
        return {
            "cells_count": count,
            "mapped_cells": len(service.get_mapping()),
            "internal_cells_count": internal_count,
            "mapped_internal_cells": len(service.get_internal_mapping()),
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
        results = service.open_cells_by_uuids([request.cell_id])
        if not results:
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail="No response from cell service"
            )

        result = results[0]

        if result.get("success"):
            return {
                "success": True,
                "message": "Cell opened successfully",
                "cell_number": result.get("cell_number") or result.get("door_number"),
                "cell_uuid": result.get("cell_uuid"),
                "type": result.get("type", "external")
            }
        else:
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"Failed to open cell: {result.get('error', 'Unknown error')}"
            )
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to prepare cell: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=str(e)
        )
