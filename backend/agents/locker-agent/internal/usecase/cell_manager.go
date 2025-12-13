package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity"
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
		return fmt.Errorf("cell manager usecase - SyncCells - parse automat ID: %w", err)
	}

	externalCells := make([]uuid.UUID, 0, len(req.CellsOut))
	for _, cellIDStr := range req.CellsOut {
		cellID, err := uuid.Parse(cellIDStr)
		if err != nil {
			return fmt.Errorf("cell manager usecase - SyncCells - parse external cell ID: %w", err)
		}
		externalCells = append(externalCells, cellID)
	}

	internalCells := make([]uuid.UUID, 0, len(req.CellsInternal))
	for _, cellIDStr := range req.CellsInternal {
		cellID, err := uuid.Parse(cellIDStr)
		if err != nil {
			return fmt.Errorf("cell manager usecase - SyncCells - parse internal cell ID: %w", err)
		}
		internalCells = append(internalCells, cellID)
	}

	if err := uc.cellRepo.Sync(automatID, externalCells, internalCells); err != nil {
		return fmt.Errorf("cell manager usecase - SyncCells - repo.Sync: %w", err)
	}

	uc.logger.Info("Cells synchronized successfully", nil, map[string]interface{}{
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
		return nil, fmt.Errorf("cell manager usecase - OpenCell: %w", inmemory.ErrNotInitialized)
	}

	cellUUID, err := uc.cellRepo.GetCellUUID(cellNumber)
	if err != nil {
		return nil, fmt.Errorf("cell manager usecase - OpenCell - get cell UUID: %w", err)
	}

	if uc.display != nil {
		_ = uc.display.ShowCellOpening(cellNumber, orderNumber)
	}

	if err := uc.arduino.OpenCell(cellNumber); err != nil {
		return nil, fmt.Errorf("cell manager usecase - OpenCell - arduino: %w", err)
	}

	if uc.display != nil {
		_ = uc.display.ShowCellOpened(cellNumber)
	}

	uc.logger.Info("Cell opened successfully", nil, map[string]interface{}{
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
		return nil, fmt.Errorf("cell manager usecase - OpenInternalDoor: %w", inmemory.ErrNotInitialized)
	}

	doorUUID, err := uc.cellRepo.GetInternalCellUUID(doorNumber)
	if err != nil {
		return nil, fmt.Errorf("cell manager usecase - OpenInternalDoor - get door UUID: %w", err)
	}

	if err := uc.arduino.OpenInternalDoor(doorNumber); err != nil {
		return nil, fmt.Errorf("cell manager usecase - OpenInternalDoor - arduino: %w", err)
	}

	uc.logger.Info("Internal door opened successfully", nil, map[string]interface{}{
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

func (uc *CellManagerUseCase) OpenCellsByUUIDs(ctx context.Context, cellUUIDs []string) []entity.OpenCellResponse {
	results := make([]entity.OpenCellResponse, 0, len(cellUUIDs))

	for _, cellUUIDStr := range cellUUIDs {
		cellUUID, err := uuid.Parse(cellUUIDStr)
		if err != nil {
			uc.logger.Error("Failed to parse cell UUID", err, map[string]interface{}{
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
			uc.logger.Error("Cell UUID not found in mapping", err, map[string]interface{}{
				"cell_uuid": cellUUID.String(),
			})
			results = append(results, entity.OpenCellResponse{
				Success:  false,
				CellUUID: cellUUID,
				Type:     "unknown",
			})
			continue
		}

		var response *entity.OpenCellResponse
		if cellType == entity.CellTypeExternal {
			response, err = uc.OpenCell(ctx, cellNumber, "")
		} else {
			response, err = uc.OpenInternalDoor(ctx, cellNumber)
		}

		if err != nil {
			uc.logger.Error("Failed to open cell", err, map[string]interface{}{
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

		results = append(results, *response)
	}

	return results
}

func (uc *CellManagerUseCase) PrepareCell(ctx context.Context, cellUUIDStr string) (*entity.PrepareCellResponse, error) {
	if !uc.cellRepo.IsInitialized() {
		return nil, fmt.Errorf("cell manager usecase - PrepareCell: %w", inmemory.ErrNotInitialized)
	}

	cellUUID, err := uuid.Parse(cellUUIDStr)
	if err != nil {
		return nil, fmt.Errorf("cell manager usecase - PrepareCell - parse UUID: %w", err)
	}

	cellNumber, cellType, err := uc.cellRepo.GetCellNumber(cellUUID)
	if err != nil {
		return nil, fmt.Errorf("cell manager usecase - PrepareCell - get cell number: %w", err)
	}

	var response *entity.OpenCellResponse
	if cellType == entity.CellTypeExternal {
		response, err = uc.OpenCell(ctx, cellNumber, "")
	} else {
		response, err = uc.OpenInternalDoor(ctx, cellNumber)
	}

	if err != nil {
		return nil, fmt.Errorf("cell manager usecase - PrepareCell - open cell: %w", err)
	}

	return &entity.PrepareCellResponse{
		Success:    response.Success,
		Message:    "Cell opened successfully",
		CellNumber: response.CellNumber,
		CellUUID:   response.CellUUID.String(),
		Type:       response.Type,
	}, nil
}

func (uc *CellManagerUseCase) GetCellsCount() *entity.CellsCountResponse {
	mapping := uc.cellRepo.GetMapping()

	totalExternal, err := uc.arduino.GetCellsCount()
	if err != nil {
		uc.logger.Error("Failed to get cells count from Arduino", err)
		totalExternal = 0
	}

	totalInternal := 0

	return &entity.CellsCountResponse{
		CellsCount:          totalExternal,
		MappedCells:         len(mapping.ExternalCells),
		InternalCellsCount:  totalInternal,
		MappedInternalCells: len(mapping.InternalCells),
	}
}
