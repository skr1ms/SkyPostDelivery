package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/hardware"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/repo/inmemory"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/pkg/logger"
)

type CellManagerUseCase struct {
	cellRepo inmemory.CellMappingInterface
	arduino  hardware.ArduinoInterface
	display  hardware.DisplayInterface
	logger   logger.Interface
}

func NewCellManagerUseCase(
	cellRepo inmemory.CellMappingInterface,
	arduino hardware.ArduinoInterface,
	display hardware.DisplayInterface,
	log logger.Interface,
) *CellManagerUseCase {
	return &CellManagerUseCase{
		cellRepo: cellRepo,
		arduino:  arduino,
		display:  display,
		logger:   log,
	}
}

func (uc *CellManagerUseCase) SyncCells(ctx context.Context, req *entity.SyncCellsRequest) error {
	automatID, err := uuid.Parse(req.ParcelAutomatID)
	if err != nil {
		return fmt.Errorf("CellManagerUseCase - SyncCells - ParseAutomatID: %w", entityError.ErrCellInvalidUUID)
	}

	externalCells := make([]uuid.UUID, 0, len(req.CellsOut))
	for i, cellIDStr := range req.CellsOut {
		cellID, err := uuid.Parse(cellIDStr)
		if err != nil {
			return fmt.Errorf("CellManagerUseCase - SyncCells - ParseExternalCellID[%d]: %w", i, entityError.ErrCellInvalidUUID)
		}
		externalCells = append(externalCells, cellID)
	}

	internalCells := make([]uuid.UUID, 0, len(req.CellsInternal))
	for i, cellIDStr := range req.CellsInternal {
		cellID, err := uuid.Parse(cellIDStr)
		if err != nil {
			return fmt.Errorf("CellManagerUseCase - SyncCells - ParseInternalCellID[%d]: %w", i, entityError.ErrCellInvalidUUID)
		}
		internalCells = append(internalCells, cellID)
	}

	if err := uc.cellRepo.Sync(automatID, externalCells, internalCells); err != nil {
		return fmt.Errorf("CellManagerUseCase - SyncCells - Sync: %w", err)
	}

	uc.logger.Info("Cells synchronized successfully", nil, map[string]any{
		"automat_id":     automatID.String(),
		"external_cells": len(externalCells),
		"internal_cells": len(internalCells),
	})

	return nil
}

func (uc *CellManagerUseCase) GetMapping() *entity.CellMapping {
	return uc.cellRepo.GetMapping()
}

func (uc *CellManagerUseCase) IsInitialized() bool {
	return uc.cellRepo.IsInitialized()
}

func (uc *CellManagerUseCase) OpenCell(ctx context.Context, cellNumber int, orderNumber string) (*entity.OpenCellResponse, error) {
	if !uc.cellRepo.IsInitialized() {
		return nil, entityError.ErrCellNotInitialized
	}

	cellUUID, err := uc.cellRepo.GetCellUUID(cellNumber)
	if err != nil {
		return nil, err
	}

	if uc.display != nil {
		_ = uc.display.ShowCellOpening(cellNumber, orderNumber)
	}

	if err := uc.arduino.OpenCell(cellNumber); err != nil {
		return nil, fmt.Errorf("CellManagerUseCase - OpenCell - OpenCell: %w", err)
	}

	if uc.display != nil {
		_ = uc.display.ShowCellOpened(cellNumber)
	}

	uc.logger.Info("Cell opened successfully", nil, map[string]any{
		"cell_number": cellNumber,
		"cell_uuid":   cellUUID.String(),
	})

	return &entity.OpenCellResponse{
		Success:       true,
		CellNumber:    cellNumber,
		CellUUID:      cellUUID,
		Action:        "opened",
		Type:          "external",
		ArduinoStatus: "OK",
	}, nil
}

func (uc *CellManagerUseCase) OpenInternalDoor(ctx context.Context, doorNumber int) (*entity.OpenCellResponse, error) {
	if !uc.cellRepo.IsInitialized() {
		return nil, entityError.ErrCellNotInitialized
	}

	doorUUID, err := uc.cellRepo.GetInternalCellUUID(doorNumber)
	if err != nil {
		return nil, err
	}

	if err := uc.arduino.OpenInternalDoor(doorNumber); err != nil {
		return nil, fmt.Errorf("CellManagerUseCase - OpenInternalDoor - OpenInternalDoor: %w", err)
	}

	uc.logger.Info("Internal door opened successfully", nil, map[string]any{
		"door_number": doorNumber,
		"door_uuid":   doorUUID.String(),
	})

	return &entity.OpenCellResponse{
		Success:       true,
		CellNumber:    doorNumber,
		CellUUID:      doorUUID,
		Action:        "internal_opened",
		Type:          "internal",
		ArduinoStatus: "OK",
	}, nil
}

func (uc *CellManagerUseCase) openCellByType(ctx context.Context, cellNumber int, cellType entity.CellType, orderNumber string) (*entity.OpenCellResponse, error) {
	if cellType == entity.CellTypeExternal {
		return uc.OpenCell(ctx, cellNumber, orderNumber)
	}
	return uc.OpenInternalDoor(ctx, cellNumber)
}

func (uc *CellManagerUseCase) OpenCellsByUUIDs(ctx context.Context, cellUUIDs []string) []entity.OpenCellResponse {
	results := make([]entity.OpenCellResponse, 0, len(cellUUIDs))
	successCount := 0
	failedCount := 0

	for _, cellUUIDStr := range cellUUIDs {
		cellUUID, err := uuid.Parse(cellUUIDStr)
		if err != nil {
			failedCount++
			uc.logger.Error("Failed to parse cell UUID", err, map[string]any{
				"cell_uuid": cellUUIDStr,
			})
			results = append(results, entity.OpenCellResponse{
				Success: false,
				Type:    "unknown",
			})
			continue
		}

		cellNumber, cellType, err := uc.cellRepo.GetCellNumber(cellUUID)
		if err != nil {
			failedCount++
			uc.logger.Error("Cell UUID not found in mapping", err, map[string]any{
				"cell_uuid": cellUUID.String(),
			})
			results = append(results, entity.OpenCellResponse{
				Success:  false,
				CellUUID: cellUUID,
				Type:     "unknown",
			})
			continue
		}

		response, err := uc.openCellByType(ctx, cellNumber, cellType, "")
		if err != nil {
			failedCount++
			uc.logger.Error("Failed to open cell", err, map[string]any{
				"cell_number": cellNumber,
				"cell_type":   cellType,
			})
			results = append(results, entity.OpenCellResponse{
				Success:    false,
				CellNumber: cellNumber,
				CellUUID:   cellUUID,
				Type:       string(cellType),
			})
			continue
		}

		successCount++
		results = append(results, *response)
	}

	uc.logger.Info("Batch cell opening completed", nil, map[string]any{
		"total":   len(cellUUIDs),
		"success": successCount,
		"failed":  failedCount,
	})

	return results
}

func (uc *CellManagerUseCase) PrepareCell(ctx context.Context, cellUUIDStr string) (*entity.PrepareCellResponse, error) {
	if !uc.cellRepo.IsInitialized() {
		return nil, entityError.ErrCellNotInitialized
	}

	cellUUID, err := uuid.Parse(cellUUIDStr)
	if err != nil {
		return nil, fmt.Errorf("CellManagerUseCase - PrepareCell - ParseUUID: %w", entityError.ErrCellInvalidUUID)
	}

	cellNumber, cellType, err := uc.cellRepo.GetCellNumber(cellUUID)
	if err != nil {
		return nil, fmt.Errorf("CellManagerUseCase - PrepareCell - GetCellNumber: %w", err)
	}

	response, err := uc.openCellByType(ctx, cellNumber, cellType, "")
	if err != nil {
		return nil, fmt.Errorf("CellManagerUseCase - PrepareCell - openCellByType: %w", err)
	}

	return &entity.PrepareCellResponse{
		Success:    response.Success,
		Message:    "Cell opened successfully",
		CellNumber: response.CellNumber,
		CellUUID:   response.CellUUID.String(),
		Type:       response.Type,
	}, nil
}

func (uc *CellManagerUseCase) GetCellsCount() (*entity.CellsCountResponse, error) {
	mapping := uc.cellRepo.GetMapping()

	totalExternal, err := uc.arduino.GetCellsCount()
	if err != nil {
		return nil, fmt.Errorf("CellManagerUseCase - GetCellsCount - GetCellsCount: %w", err)
	}

	totalInternal := 0

	return &entity.CellsCountResponse{
		CellsCount:          totalExternal,
		MappedCells:         len(mapping.ExternalCells),
		InternalCellsCount:  totalInternal,
		MappedInternalCells: len(mapping.InternalCells),
	}, nil
}
