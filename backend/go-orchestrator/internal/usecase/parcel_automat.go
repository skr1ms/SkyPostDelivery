package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/v1/request"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/logger"
)

type ParcelAutomatUseCase struct {
	parcelAutomatRepo  repo.ParcelAutomatRepo
	lockerRepo         repo.LockerRepo
	internalLockerRepo repo.InternalLockerRepo
	orderRepo          repo.OrderRepo
	deliveryRepo       repo.DeliveryRepo
	qrUseCase          *QRUseCase
	orangePIWebAPI     repo.OrangePIWebAPI
	logger             logger.Interface
}

func NewParcelAutomatUseCase(
	parcelAutomatRepo repo.ParcelAutomatRepo,
	lockerRepo repo.LockerRepo,
	internalLockerRepo repo.InternalLockerRepo,
	orderRepo repo.OrderRepo,
	deliveryRepo repo.DeliveryRepo,
	qrUseCase *QRUseCase,
	orangePIWebAPI repo.OrangePIWebAPI,
	logger logger.Interface,
) *ParcelAutomatUseCase {
	return &ParcelAutomatUseCase{
		parcelAutomatRepo:  parcelAutomatRepo,
		lockerRepo:         lockerRepo,
		internalLockerRepo: internalLockerRepo,
		orderRepo:          orderRepo,
		deliveryRepo:       deliveryRepo,
		qrUseCase:          qrUseCase,
		orangePIWebAPI:     orangePIWebAPI,
		logger:             logger,
	}
}

const defaultInternalDoorCount = 3

func (uc *ParcelAutomatUseCase) Create(ctx context.Context, automat *entity.ParcelAutomat, cells []request.CellDimensions) (*entity.ParcelAutomat, error) {
	automat.IsWorking = true
	createdAutomat, err := uc.parcelAutomatRepo.Create(ctx, automat)
	if err != nil {
		return nil, err
	}

	cleanup := func() {
		if err := uc.parcelAutomatRepo.Delete(ctx, createdAutomat.ID); err != nil {
			uc.logger.Error("ParcelAutomatUseCase - Create - Rollback failed", err, map[string]any{
				"automatID": createdAutomat.ID,
			})
		}
	}

	cellUUIDs := make([]uuid.UUID, 0, len(cells))
	for i, cell := range cells {
		cellEntity := &entity.LockerCell{
			PostID: createdAutomat.ID,
			Height: cell.Height,
			Length: cell.Length,
			Width:  cell.Width,
		}
		createdCell, err := uc.lockerRepo.CreateWithNumber(ctx, cellEntity, i+1)
		if err != nil {
			cleanup()
			return nil, fmt.Errorf("ParcelAutomatUseCase - Create - CreateLockerCell: %w", err)
		}
		cellUUIDs = append(cellUUIDs, createdCell.ID)
	}

	internalCellUUIDs := make([]uuid.UUID, 0, defaultInternalDoorCount)
	for i := 0; i < defaultInternalDoorCount; i++ {
		cellEntity := &entity.LockerCell{
			PostID: createdAutomat.ID,
			Height: 0,
			Length: 0,
			Width:  0,
		}
		createdCell, err := uc.internalLockerRepo.CreateWithNumber(ctx, cellEntity, i+1)
		if err != nil {
			cleanup()
			return nil, fmt.Errorf("ParcelAutomatUseCase - Create - CreateInternalLockerCell: %w", err)
		}
		internalCellUUIDs = append(internalCellUUIDs, createdCell.ID)
	}

	if createdAutomat.IPAddress != "" {
		if err := uc.orangePIWebAPI.SendCellUUIDs(ctx, createdAutomat.IPAddress, createdAutomat.ID, cellUUIDs, internalCellUUIDs); err != nil {
			uc.logger.Warn("ParcelAutomatUseCase - Create - SendCellUUIDs", err, map[string]any{"ipAddress": createdAutomat.IPAddress})
		}
	}

	return createdAutomat, nil
}

func (uc *ParcelAutomatUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.ParcelAutomat, error) {
	automat, err := uc.parcelAutomatRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return automat, nil
}

func (uc *ParcelAutomatUseCase) List(ctx context.Context) ([]*entity.ParcelAutomat, error) {
	automats, err := uc.parcelAutomatRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("ParcelAutomatUseCase - List: %w", err)
	}
	return automats, nil
}

func (uc *ParcelAutomatUseCase) ListWorking(ctx context.Context) ([]*entity.ParcelAutomat, error) {
	automats, err := uc.parcelAutomatRepo.ListWorking(ctx)
	if err != nil {
		return nil, fmt.Errorf("ParcelAutomatUseCase - ListWorking: %w", err)
	}
	return automats, nil
}

func (uc *ParcelAutomatUseCase) UpdateStatus(ctx context.Context, automat *entity.ParcelAutomat) (*entity.ParcelAutomat, error) {
	updatedAutomat, err := uc.parcelAutomatRepo.UpdateStatus(ctx, automat)
	if err != nil {
		return nil, err
	}

	return updatedAutomat, nil
}

func (uc *ParcelAutomatUseCase) Update(ctx context.Context, automat *entity.ParcelAutomat) (*entity.ParcelAutomat, error) {
	updatedAutomat, err := uc.parcelAutomatRepo.Update(ctx, automat)
	if err != nil {
		return nil, err
	}

	if updatedAutomat.IPAddress != "" {
		outCells, err := uc.lockerRepo.ListCellsByPostID(ctx, updatedAutomat.ID)
		if err != nil {
			uc.logger.Warn("ParcelAutomatUseCase - Update - ListCells", err, map[string]any{"automatID": updatedAutomat.ID})
		} else {
			internalCells, err := uc.internalLockerRepo.ListCellsByPostID(ctx, updatedAutomat.ID)
			if err != nil {
				uc.logger.Warn("ParcelAutomatUseCase - Update - ListInternalCells", err, map[string]any{"automatID": updatedAutomat.ID})
			} else {
				outCellUUIDs := make([]uuid.UUID, 0, len(outCells))
				for _, cell := range outCells {
					outCellUUIDs = append(outCellUUIDs, cell.ID)
				}

				internalCellUUIDs := make([]uuid.UUID, 0, len(internalCells))
				for _, cell := range internalCells {
					internalCellUUIDs = append(internalCellUUIDs, cell.ID)
				}

				if err := uc.orangePIWebAPI.SendCellUUIDs(ctx, updatedAutomat.IPAddress, updatedAutomat.ID, outCellUUIDs, internalCellUUIDs); err != nil {
					uc.logger.Warn("ParcelAutomatUseCase - Update - SendCellUUIDs", err, map[string]any{"ipAddress": updatedAutomat.IPAddress})
				}
			}
		}
	}

	return updatedAutomat, nil
}

func (uc *ParcelAutomatUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	if err := uc.parcelAutomatRepo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

func (uc *ParcelAutomatUseCase) GetAutomatCells(ctx context.Context, automatID uuid.UUID) ([]*entity.LockerCell, error) {
	cells, err := uc.lockerRepo.ListCellsByPostID(ctx, automatID)
	if err != nil {
		return nil, fmt.Errorf("ParcelAutomatUseCase - GetAutomatCells: %w", err)
	}
	return cells, nil
}

func (uc *ParcelAutomatUseCase) ProcessQRScan(ctx context.Context, qrDataJSON string, automatID uuid.UUID) ([]uuid.UUID, error) {
	user, err := uc.qrUseCase.ValidateQR(ctx, qrDataJSON)
	if err != nil {
		return nil, fmt.Errorf("ParcelAutomatUseCase - ProcessQRScan - ValidateQR: %w", err)
	}

	orders, err := uc.orderRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("ParcelAutomatUseCase - ProcessQRScan - ListOrders: %w", err)
	}

	cellIDs := make([]uuid.UUID, 0)

	for _, order := range orders {
		if order.Status != "delivered" || order.LockerCellID == nil {
			continue
		}

		delivery, err := uc.deliveryRepo.GetByOrderID(ctx, order.ID)
		if err != nil {
			uc.logger.Warn("ParcelAutomatUseCase - ProcessQRScan - GetDelivery", err, map[string]any{
				"orderID": order.ID,
			})
			continue
		}

		if delivery.ParcelAutomatID != automatID {
			continue
		}

		cell, err := uc.lockerRepo.GetCellByID(ctx, *order.LockerCellID)
		if err != nil {
			uc.logger.Warn("ParcelAutomatUseCase - ProcessQRScan - GetCell", err, map[string]any{
				"cellID": *order.LockerCellID,
			})
			continue
		}

		if cell.Status == "occupied" {
			cellIDs = append(cellIDs, cell.ID)
		}
	}

	if len(cellIDs) == 0 {
		return nil, entityError.ErrQRNoOrdersForPickup
	}

	for _, cellID := range cellIDs {
		cell, err := uc.lockerRepo.GetCellByID(ctx, cellID)
		if err != nil {
			uc.logger.Error("ParcelAutomatUseCase - ProcessQRScan - GetCell", err, map[string]any{
				"cellID": cellID,
			})
			continue
		}
		cell.Status = "opened"
		if err := uc.lockerRepo.UpdateCellStatus(ctx, cell); err != nil {
			uc.logger.Error("ParcelAutomatUseCase - ProcessQRScan - UpdateCellStatus", err, map[string]any{
				"cellID": cellID,
			})
		}
	}

	return cellIDs, nil
}

func (uc *ParcelAutomatUseCase) UpdateCell(ctx context.Context, cellID uuid.UUID, height, length, width float64) (*entity.LockerCell, error) {
	cell, err := uc.lockerRepo.GetCellByID(ctx, cellID)
	if err != nil {
		return nil, err
	}

	if cell.Status != "available" {
		return nil, entityError.ErrLockerCellCannotUpdate
	}

	cell.Height = height
	cell.Length = length
	cell.Width = width

	updatedCell, err := uc.lockerRepo.UpdateDimensions(ctx, cell)
	if err != nil {
		return nil, fmt.Errorf("ParcelAutomatUseCase - UpdateCell - UpdateDimensions: %w", err)
	}

	return updatedCell, nil
}

func (uc *ParcelAutomatUseCase) ConfirmPickup(ctx context.Context, cellIDs []uuid.UUID) error {
	var hasErrors bool

	for _, cellID := range cellIDs {
		cell, err := uc.lockerRepo.GetCellByID(ctx, cellID)
		if err != nil {
			uc.logger.Error("ParcelAutomatUseCase - ConfirmPickup - GetCell", err, map[string]any{
				"cellID": cellID,
			})
			hasErrors = true
			continue
		}

		if cell.Status != "opened" {
			uc.logger.Warn("ParcelAutomatUseCase - ConfirmPickup - InvalidStatus", nil, map[string]any{
				"cellID": cellID,
				"status": cell.Status,
			})
			continue
		}

		cell.Status = "available"
		if err := uc.lockerRepo.UpdateCellStatus(ctx, cell); err != nil {
			uc.logger.Error("ParcelAutomatUseCase - ConfirmPickup - UpdateCellStatus", err, map[string]any{
				"cellID": cellID,
			})
			hasErrors = true
			continue
		}

		order, err := uc.orderRepo.GetByLockerCellID(ctx, cellID)
		if err != nil {
			uc.logger.Warn("ParcelAutomatUseCase - ConfirmPickup - GetOrder", err, map[string]any{
				"cellID": cellID,
			})
			continue
		}

		delivery, err := uc.deliveryRepo.GetByOrderID(ctx, order.ID)
		if err == nil && delivery.InternalLockerCellID != nil {
			if uc.internalLockerRepo != nil {
				internalCell, err := uc.internalLockerRepo.GetCellByID(ctx, *delivery.InternalLockerCellID)
				if err != nil {
					uc.logger.Warn("ParcelAutomatUseCase - ConfirmPickup - GetInternalCell", err, map[string]any{
						"cellID": *delivery.InternalLockerCellID,
					})
				} else {
					internalCell.Status = "available"
					if err := uc.internalLockerRepo.UpdateCellStatus(ctx, internalCell); err != nil {
						uc.logger.Warn("ParcelAutomatUseCase - ConfirmPickup - ReleaseInternalCell", err, map[string]any{
							"cellID": *delivery.InternalLockerCellID,
						})
					}
				}
			}
		}

		order.Status = "completed"
		if _, err := uc.orderRepo.UpdateStatus(ctx, order); err != nil {
			uc.logger.Error("ParcelAutomatUseCase - ConfirmPickup - UpdateOrderStatus", err, map[string]any{
				"orderID": order.ID,
			})
			hasErrors = true
		}
	}

	if hasErrors {
		return entityError.ErrParcelAutomatPartialPickupFailure
	}

	return nil
}

func (uc *ParcelAutomatUseCase) PrepareCell(ctx context.Context, orderID, parcelAutomatID uuid.UUID) (uuid.UUID, *uuid.UUID, error) {
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return uuid.Nil, nil, err
	}

	if order.LockerCellID == nil {
		return uuid.Nil, nil, entityError.ErrOrderHasNoCellAssigned
	}

	cell, err := uc.lockerRepo.GetCellByID(ctx, *order.LockerCellID)
	if err != nil {
		return uuid.Nil, nil, err
	}

	if cell.Status != "occupied" && cell.Status != "reserved" {
		return uuid.Nil, nil, entityError.ErrLockerCellInvalidStatus
	}

	automat, err := uc.parcelAutomatRepo.GetByID(ctx, parcelAutomatID)
	if err != nil {
		return uuid.Nil, nil, err
	}

	var internalDoorID *uuid.UUID
	delivery, err := uc.deliveryRepo.GetByOrderID(ctx, order.ID)
	if err != nil {
		uc.logger.Warn("ParcelAutomatUseCase - PrepareCell - GetByOrderID", err, map[string]any{
			"orderID": order.ID,
		})
	} else if delivery.InternalLockerCellID != nil {
		internalDoorID = delivery.InternalLockerCellID
		if err := uc.orangePIWebAPI.OpenCell(ctx, automat.IPAddress, *internalDoorID); err != nil {
			return uuid.Nil, nil, fmt.Errorf("ParcelAutomatUseCase - PrepareCell - OpenInternalCell: %w", err)
		}
		if uc.internalLockerRepo != nil {
			internalCell, err := uc.internalLockerRepo.GetCellByID(ctx, *internalDoorID)
			if err != nil {
				uc.logger.Warn("ParcelAutomatUseCase - PrepareCell - GetInternalCell", err, map[string]any{"cellID": *internalDoorID})
			} else {
				internalCell.Status = "opened"
				if err := uc.internalLockerRepo.UpdateCellStatus(ctx, internalCell); err != nil {
					uc.logger.Warn("ParcelAutomatUseCase - PrepareCell - UpdateInternalCellStatus", err, map[string]any{"cellID": *internalDoorID})
				}
			}
		}
	}
	return cell.ID, internalDoorID, nil
}
