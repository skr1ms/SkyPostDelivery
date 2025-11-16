import asyncio
import time
import logging
from typing import List, Optional, Dict, TYPE_CHECKING

import httpx
from pyzbar import pyzbar

from ..models.schemas import QRScanResponse, ConfirmLoadedRequest
from config.config import settings

logger = logging.getLogger(__name__)

if TYPE_CHECKING:
    from ..hardware.display_controller import DisplayController
    from ..hardware.qr_scanner import QRScanner
    from .cell_management_service import CellManagementService


class QRScanService:
    def __init__(self, orchestrator_url: str = None):
        self.orchestrator_url = orchestrator_url or settings.go_orchestrator_url
        self.client = httpx.AsyncClient(timeout=10.0)

    async def validate_qr_with_go(self, qr_data: str, parcel_automat_id: str) -> QRScanResponse:
        url = f"{self.orchestrator_url}/automats/qr-scan"
        payload = {
            "qr_data": qr_data,
            "parcel_automat_id": parcel_automat_id
        }
        logger.info(f"Sending QR validation request to GO: {url}")
        try:
            response = await self.client.post(url, json=payload)
            response.raise_for_status()
            data = response.json()
            logger.info(
                f"QR validated successfully. Cells to open: {data.get('cell_ids', [])}")
            return QRScanResponse(**data)
        except httpx.HTTPError as e:
            logger.error(f"Failed to validate QR with GO: {e}")
            raise

    async def confirm_pickup(self, cell_ids: List[str]) -> dict:
        url = f"{self.orchestrator_url}/automats/confirm-pickup"
        payload = {"cell_ids": cell_ids}
        logger.info(f"Sending pickup confirmation to GO for cells: {cell_ids}")
        try:
            response = await self.client.post(url, json=payload, headers={"Content-Type": "application/json"})
            response.raise_for_status()
            data = response.json()
            logger.info("Pickup confirmed successfully")
            return data
        except httpx.HTTPError as e:
            logger.error(f"Failed to confirm pickup with GO: {e}")
            raise

    async def confirm_loaded(self, order_id: str, locker_cell_id: str) -> dict:
        url = f"{self.orchestrator_url}/automats/confirm-loaded"
        payload = ConfirmLoadedRequest(
            order_id=order_id,
            locker_cell_id=locker_cell_id
        ).model_dump(mode='json')
        logger.info(f"Sending loaded confirmation to GO for order: {order_id}")
        try:
            response = await self.client.post(url, json=payload, headers={"Content-Type": "application/json"})
            response.raise_for_status()
            data = response.json()
            logger.info("Load confirmed successfully")
            return data
        except httpx.HTTPError as e:
            logger.error(f"Failed to confirm load with GO: {e}")
            raise

    async def close(self):
        await self.client.aclose()


class QRScannerWorker:
    QR_FORGET_TIME_SECONDS = 120

    def __init__(
        self,
        qr_scanner: "QRScanner",
        qr_service: QRScanService,
        cell_service: "CellManagementService",
        display: Optional["DisplayController"] = None,
        scan_interval: float = 0.1,
        stable_frames: int = 3,
        miss_frames: int = 5,
        debounce_seconds: float = 5.0,
    ):
        self.qr_scanner = qr_scanner
        self.qr_service = qr_service
        self.cell_service = cell_service
        self.display = display

        self.scan_interval = scan_interval
        self.stable_frames = stable_frames
        self.miss_frames = miss_frames
        self.debounce_seconds = debounce_seconds

        self.qr_visible_frames: Dict[str, int] = {}
        self.qr_missing_frames: Dict[str, int] = {}
        self.qr_to_id: Dict[str, int] = {}
        self.last_seen: Dict[str, float] = {}
        self.active_qrs: set[str] = set()

        self._no_frame_count: int = 0
        self._frame_seen_logged: bool = False
        self._frame_counter: int = 0

        self.qr_index_counter = 0
        self.running = False
        self._task: Optional[asyncio.Task] = None
        self._mock_queue: Optional[asyncio.Queue] = None

    async def start(self):
        if self.running:
            return

        self.running = True
        logger.info(
            "Starting QRScannerWorker (mock_mode=%s, interval=%.2fs)",
            getattr(self.qr_scanner, "mock_mode", False),
            self.scan_interval,
        )

        if getattr(self.qr_scanner, "mock_mode", False) and self._mock_queue is None:
            self._mock_queue = asyncio.Queue()

        if self.display:
            self.display.show_welcome()

        if not getattr(self.qr_scanner, "mock_mode", False):
            frame = self.qr_scanner.read_frame()
            if frame is None:
                logger.warning(
                    "Initial camera frame could not be read; scanner will keep retrying")

        self._task = asyncio.create_task(self._scan_loop())

    async def stop(self):
        if not self.running:
            return

        self.running = False
        if self._task:
            self._task.cancel()
            try:
                await self._task
            except asyncio.CancelledError:
                pass
        self._task = None
        logger.info("QRScannerWorker stopped")

    def enqueue_mock_qr(self, qr_data: str):
        if not self._mock_queue:
            raise RuntimeError(
                "Mock mode is disabled; cannot enqueue mock QR data")
        self._mock_queue.put_nowait(qr_data)

    async def _scan_loop(self):
        logger.info("QR scanner loop started")
        while self.running:
            try:
                if getattr(self.qr_scanner, "mock_mode", False):
                    qr_text = await self._get_next_mock_qr()
                    if qr_text:
                        await self._handle_confirmed_qr(qr_text)
                    await asyncio.sleep(self.scan_interval)
                    continue

                frame = self.qr_scanner.read_frame()
                if frame is None:
                    self._no_frame_count += 1
                    if self._no_frame_count == 1 or self._no_frame_count % 10 == 0:
                        logger.warning(
                            "Camera returned no frame (count=%d). "
                            "Ensure the camera is connected and not used by another process.",
                            self._no_frame_count,
                        )
                    await asyncio.sleep(1.0)
                    continue
                else:
                    if not self._frame_seen_logged:
                        logger.info(
                            "Camera frame captured successfully for QR scanning.")
                        self._frame_seen_logged = True
                    self._no_frame_count = 0
                    self._frame_counter += 1

                decoded_qrs = [
                    obj.data.decode("utf-8").strip()
                    for obj in pyzbar.decode(frame)
                    if obj.data
                ]

                if not decoded_qrs and self._frame_counter % 30 == 0:
                    logger.info(
                        "Processed %d frames without detecting a QR code yet. "
                        "Ensure QR is well-lit and within focus.",
                        self._frame_counter,
                    )

                now = time.time()
                if decoded_qrs:
                    for qr_text in decoded_qrs:
                        self.qr_visible_frames[qr_text] = self.qr_visible_frames.get(
                            qr_text, 0) + 1
                        self.qr_missing_frames[qr_text] = 0

                        if qr_text not in self.qr_to_id:
                            self.qr_index_counter += 1
                            self.qr_to_id[qr_text] = self.qr_index_counter
                            logger.info(
                                f"[NEW] QR#{self.qr_index_counter} detected")

                        qr_id = self.qr_to_id[qr_text]
                        if self.qr_visible_frames[qr_text] == self.stable_frames:
                            if self._should_process_qr(qr_text, now):
                                await self._handle_confirmed_qr(qr_text, qr_id)
                                self.last_seen[qr_text] = now
                                self.active_qrs.add(qr_text)
                                self.qr_visible_frames[qr_text] = 0
                                self.qr_missing_frames[qr_text] = 0

                for tracked_qr in list(self.qr_visible_frames.keys()):
                    if tracked_qr not in decoded_qrs:
                        self.qr_missing_frames[tracked_qr] = self.qr_missing_frames.get(
                            tracked_qr, 0) + 1
                        if self.qr_missing_frames[tracked_qr] >= self.miss_frames:
                            self._forget_qr(tracked_qr, reason="lost")

                self._cleanup_stale_qrs(now)
                await asyncio.sleep(self.scan_interval)

            except asyncio.CancelledError:
                logger.info("QR scanner loop cancelled")
                break
            except Exception as exc:
                logger.exception(f"Error in QR scanner loop: {exc}")
                await asyncio.sleep(1.0)

        logger.info("QR scanner loop finished")

    async def _get_next_mock_qr(self) -> Optional[str]:
        if not self._mock_queue:
            return None
        try:
            return self._mock_queue.get_nowait()
        except asyncio.QueueEmpty:
            return None

    def _should_process_qr(self, qr_text: str, now: float) -> bool:
        if qr_text not in self.active_qrs:
            return True
        last_time = self.last_seen.get(qr_text, 0)
        return (now - last_time) >= self.debounce_seconds

    async def _handle_confirmed_qr(self, qr_text: str, qr_id: int | None = None):
        qr_id = qr_id or self.qr_to_id.get(qr_text, 0)
        logger.info(f"[SCAN] QR#{qr_id} confirmed, sending to orchestrator")

        if self.display:
            self.display.show_scanning()

        try:
            parcel_automat_id = self.cell_service.get_parcel_automat_id()
        except Exception as exc:
            logger.error(f"Parcel automat ID missing: {exc}")
            if self.display:
                self.display.show_error("No automat ID")
                await asyncio.sleep(2)
                self.display.show_welcome()
            return

        try:
            validation = await self.qr_service.validate_qr_with_go(qr_text, parcel_automat_id)
        except Exception as exc:
            logger.error(f"QR validation failed: {exc}")
            if self.display:
                self.display.show_error("Validation error")
                await asyncio.sleep(2)
                self.display.show_welcome()
            return

        if not validation.success:
            logger.warning(
                f"QR rejected by orchestrator: {validation.message}")
            if self.display:
                self.display.show_qr_invalid()
                await asyncio.sleep(2)
                self.display.show_welcome()
            return

        logger.info(f"QR accepted. Cells to open: {validation.cell_ids}")

        if self.display:
            self.display.show_qr_success()

        opened_cells = self.cell_service.open_cells_by_uuids(
            validation.cell_ids)
        success_count = sum(1 for cell in opened_cells if cell.get("success"))
        logger.info(f"Opened {success_count}/{len(validation.cell_ids)} cells")

        await asyncio.sleep(2)
        if self.display:
            self.display.show_welcome()

    def _cleanup_stale_qrs(self, now: float):
        for qr_text, last_time in list(self.last_seen.items()):
            if (now - last_time) > self.QR_FORGET_TIME_SECONDS:
                self._forget_qr(qr_text, reason="expired")

    def _forget_qr(self, qr_text: str, reason: str):
        logger.info(
            f"[CLEANUP] Forgetting QR (reason={reason}): {qr_text[:30]}...")
        self.qr_visible_frames.pop(qr_text, None)
        self.qr_missing_frames.pop(qr_text, None)
        self.qr_to_id.pop(qr_text, None)
        self.last_seen.pop(qr_text, None)
        self.active_qrs.discard(qr_text)
