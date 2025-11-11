from fastapi import APIRouter, HTTPException, Depends, status
import logging

from ..models.schemas import (
    QRScanRequest,
    ConfirmPickupRequest,
    ConfirmLoadedRequest
)
from ..services.qr_scan_service import QRScanService
from ..services.cell_management_service import CellManagementService
from ..hardware.qr_scanner import QRScanner

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/qr", tags=["qr-scanning"])


def get_qr_service() -> QRScanService:
    from ..dependencies import qr_service
    return qr_service


def get_cell_service() -> CellManagementService:
    from ..dependencies import cell_service
    return cell_service


def get_qr_scanner() -> QRScanner:
    from ..dependencies import qr_scanner
    return qr_scanner


def get_display():
    from ..dependencies import display
    return display


@router.post("/scan", status_code=status.HTTP_200_OK)
async def scan_qr_code(request: QRScanRequest, qr_service: QRScanService = Depends(get_qr_service), cell_service: CellManagementService = Depends(get_cell_service)):
    try:
        parcel_automat_id = cell_service.get_parcel_automat_id()
        validation_result = await qr_service.validate_qr_with_go(request.qr_data, parcel_automat_id)
        if not validation_result.success:
            return {
                "success": False,
                "message": validation_result.message,
                "cells_opened": []
            }
        opened_cells = cell_service.open_cells_by_uuids(
            validation_result.cell_ids)
        return {
            "success": True,
            "message": "QR code validated and cells opened successfully",
            "cells_opened": opened_cells,
            "cell_count": len(opened_cells)
        }
    except Exception as e:
        logger.error(f"Failed to process QR scan: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to process QR scan: {str(e)}"
        )


@router.post("/confirm-pickup", status_code=status.HTTP_200_OK)
async def confirm_pickup(request: ConfirmPickupRequest, qr_service: QRScanService = Depends(get_qr_service)):
    try:
        result = await qr_service.confirm_pickup(request.cell_ids)
        return {
            "success": True,
            "message": "Pickup confirmed successfully",
            "data": result
        }
    except Exception as e:
        logger.error(f"Failed to confirm pickup: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to confirm pickup: {str(e)}"
        )


@router.post("/confirm-loaded", status_code=status.HTTP_200_OK)
async def confirm_loaded(request: ConfirmLoadedRequest, qr_service: QRScanService = Depends(get_qr_service)):
    try:
        result = await qr_service.confirm_loaded(
            request.order_id,
            request.locker_cell_id
        )
        return {
            "success": True,
            "message": "Load confirmed successfully",
            "data": result
        }
    except Exception as e:
        logger.error(f"Failed to confirm load: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to confirm load: {str(e)}"
        )


@router.post("/scan-from-camera", status_code=status.HTTP_200_OK)
async def scan_qr_from_camera(qr_scanner: QRScanner = Depends(get_qr_scanner), qr_service: QRScanService = Depends(get_qr_service),
                              cell_service: CellManagementService = Depends(get_cell_service), display=Depends(get_display)):
    import asyncio

    try:
        if display:
            display.show_scanning()

        max_attempts = 100
        qr_data = None

        for attempt in range(max_attempts):
            qr_data = qr_scanner.scan_once()
            if qr_data:
                break
            await asyncio.sleep(0.1)

        if not qr_data:
            if display:
                display.show_qr_invalid()
            return {
                "success": False,
                "message": "No QR code detected after 10 seconds",
                "cells_opened": []
            }

        parcel_automat_id = cell_service.get_parcel_automat_id()
        validation_result = await qr_service.validate_qr_with_go(qr_data, parcel_automat_id)

        if not validation_result.success:
            if display:
                display.show_qr_invalid()
            return {
                "success": False,
                "message": validation_result.message,
                "cells_opened": [],
                "qr_data": qr_data[:100]
            }

        if display:
            display.show_qr_success()

        opened_cells = cell_service.open_cells_by_uuids(
            validation_result.cell_ids)

        return {
            "success": True,
            "message": "QR code scanned and cells opened",
            "cells_opened": opened_cells,
            "cell_count": len(opened_cells),
            "qr_data": qr_data[:100]
        }

    except Exception as e:
        logger.error(f"Failed to scan QR from camera: {e}")
        if display:
            display.show_error()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to scan QR from camera: {str(e)}"
        )
