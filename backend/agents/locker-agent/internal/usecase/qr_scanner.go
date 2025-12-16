package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/hardware"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/repo/inmemory"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/pkg/logger"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/pkg/orchestrator"
)

type QRScannerUseCase struct {
	cellManager        *CellManagerUseCase
	orchestratorClient orchestrator.ClientInterface
	qrCamera           hardware.QRCameraInterface
	display            hardware.DisplayInterface
	cellRepo           inmemory.CellMappingInterface
	logger             logger.Interface
	stopChan           chan struct{}
	doneChan           chan struct{}
}

func NewQRScannerUseCase(
	cellManager *CellManagerUseCase,
	orchestratorClient orchestrator.ClientInterface,
	qrCamera hardware.QRCameraInterface,
	display hardware.DisplayInterface,
	cellRepo inmemory.CellMappingInterface,
	log logger.Interface,
) *QRScannerUseCase {
	return &QRScannerUseCase{
		cellManager:        cellManager,
		orchestratorClient: orchestratorClient,
		qrCamera:           qrCamera,
		display:            display,
		cellRepo:           cellRepo,
		logger:             log,
	}
}

func (uc *QRScannerUseCase) ProcessQRScan(ctx context.Context, qrData string) (*entity.QRScanResponse, error) {
	if !uc.cellManager.IsInitialized() {
		return nil, entityError.ErrCellNotInitialized
	}

	mapping := uc.cellManager.GetMapping()
	automatID := mapping.ParcelAutomatID

	uc.logger.Debug("Validating QR with orchestrator", nil, map[string]any{
		"automat_id": automatID.String(),
	})

	if uc.display != nil {
		_ = uc.display.ShowScanning()
	}

	validationResp, err := uc.orchestratorClient.ValidateQR(ctx, qrData, automatID.String())
	if err != nil {
		uc.logger.Error("Failed to validate QR with orchestrator", err)
		if uc.display != nil {
			_ = uc.display.ShowError("Connection error")
		}
		return nil, fmt.Errorf("QRScannerUseCase - ProcessQRScan - ValidateQR: %w", err)
	}

	if !validationResp.Success {
		uc.logger.Warn("QR validation failed", nil, map[string]any{
			"message": validationResp.Message,
		})
		if uc.display != nil {
			_ = uc.display.ShowInvalid()
		}
		return &entity.QRScanResponse{
			Success: false,
			Message: validationResp.Message,
		}, nil
	}

	if uc.display != nil {
		_ = uc.display.ShowSuccess("")
	}

	openedCells := uc.cellManager.OpenCellsByUUIDs(ctx, validationResp.CellIDs)

	successCount := 0
	for _, cell := range openedCells {
		if cell.Success {
			successCount++
		}
	}

	if uc.display != nil && successCount > 0 {
		_ = uc.display.ShowPleaseClose()
	}

	return &entity.QRScanResponse{
		Success:     true,
		Message:     "QR code validated and cells opened successfully",
		CellsOpened: openedCells,
		CellCount:   successCount,
	}, nil
}

func (uc *QRScannerUseCase) ConfirmPickup(ctx context.Context, cellIDs []string) error {
	cellUUIDs := make([]uuid.UUID, 0, len(cellIDs))
	for i, cellIDStr := range cellIDs {
		cellUUID, err := uuid.Parse(cellIDStr)
		if err != nil {
			return fmt.Errorf("QRScannerUseCase - ConfirmPickup - ParseCellID[%d]: %w", i, entityError.ErrCellInvalidUUID)
		}
		cellUUIDs = append(cellUUIDs, cellUUID)
	}

	if err := uc.orchestratorClient.ConfirmPickup(ctx, cellUUIDs); err != nil {
		return fmt.Errorf("QRScannerUseCase - ConfirmPickup - ConfirmPickup: %w", err)
	}

	uc.logger.Info("Pickup confirmed successfully", nil, map[string]any{
		"cell_count": len(cellUUIDs),
	})

	if uc.display != nil {
		_ = uc.display.ShowThankYou()
	}

	return nil
}

func (uc *QRScannerUseCase) ConfirmLoaded(ctx context.Context, orderID, lockerCellID string) error {
	orderUUID, err := uuid.Parse(orderID)
	if err != nil {
		return fmt.Errorf("QRScannerUseCase - ConfirmLoaded - ParseOrderID: %w", entityError.ErrCellInvalidUUID)
	}

	cellUUID, err := uuid.Parse(lockerCellID)
	if err != nil {
		return fmt.Errorf("QRScannerUseCase - ConfirmLoaded - ParseLockerCellID: %w", entityError.ErrCellInvalidUUID)
	}

	if err := uc.orchestratorClient.ConfirmLoaded(ctx, orderUUID, cellUUID); err != nil {
		return fmt.Errorf("QRScannerUseCase - ConfirmLoaded - ConfirmLoaded: %w", err)
	}

	uc.logger.Info("Load confirmed successfully", nil, map[string]any{
		"order_id":       orderID,
		"locker_cell_id": lockerCellID,
	})

	return nil
}

func (uc *QRScannerUseCase) StartScanner(ctx context.Context) {
	uc.stopChan = make(chan struct{})
	uc.doneChan = make(chan struct{})

	uc.qrCamera.Start(ctx)

	go uc.handleQRResults(ctx)

	uc.logger.Info("QR scanner background worker started", nil)
}

func (uc *QRScannerUseCase) handleQRResults(ctx context.Context) {
	defer close(uc.doneChan)

	resultChan := uc.qrCamera.GetResultChannel()

	for {
		select {
		case <-ctx.Done():
			uc.logger.Info("QR scanner background worker stopped (context)", nil)
			return
		case <-uc.stopChan:
			uc.logger.Info("QR scanner background worker stopped", nil)
			return
		case qrData := <-resultChan:
			if qrData == "" {
				continue
			}

			if !uc.isValidQRData(qrData) {
				uc.logger.Warn("Invalid QR data format", nil, map[string]any{
					"data_length": len(qrData),
				})
				if uc.display != nil {
					_ = uc.display.ShowError("Invalid QR code")
				}
				continue
			}

			uc.logger.Info("QR code detected, processing", nil, map[string]any{
				"data_length": len(qrData),
			})

			if _, err := uc.ProcessQRScan(ctx, qrData); err != nil {
				uc.logger.Error("Failed to process QR scan", err)
			}
		}
	}
}

func (uc *QRScannerUseCase) isValidQRData(data string) bool {
	var qr struct {
		UserID string `json:"user_id"`
		Type   string `json:"type"`
	}

	if err := json.Unmarshal([]byte(data), &qr); err != nil {
		return false
	}

	return qr.UserID != "" && qr.Type != ""
}

func (uc *QRScannerUseCase) StopScanner() {
	if uc.stopChan != nil {
		close(uc.stopChan)
	}
	if uc.doneChan != nil {
		<-uc.doneChan
	}
	uc.qrCamera.Stop()
	uc.logger.Info("QR scanner stopped", nil)
}
