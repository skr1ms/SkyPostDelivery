package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/logger"
)

type LockerUseCase struct {
	lockerRepo repo.LockerRepo
	logger     logger.Interface
}

func NewLockerUseCase(
	lockerRepo repo.LockerRepo,
	logger logger.Interface,
) *LockerUseCase {
	return &LockerUseCase{
		lockerRepo: lockerRepo,
		logger:     logger,
	}
}

var validCellStatuses = map[string]bool{
	"available": true,
	"reserved":  true,
	"occupied":  true,
	"opened":    true,
}

func (uc *LockerUseCase) GetCell(ctx context.Context, id uuid.UUID) (*entity.LockerCell, error) {
	cell, err := uc.lockerRepo.GetCellByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return cell, nil
}

func (uc *LockerUseCase) ListCells(ctx context.Context, postID uuid.UUID) ([]*entity.LockerCell, error) {
	cells, err := uc.lockerRepo.ListCellsByPostID(ctx, postID)
	if err != nil {
		return nil, err
	}
	return cells, nil
}

func (uc *LockerUseCase) UpdateCellStatus(ctx context.Context, cellID uuid.UUID, status string) error {
	if !validCellStatuses[status] {
		return entityError.ErrLockerInvalidStatus
	}

	cell, err := uc.lockerRepo.GetCellByID(ctx, cellID)
	if err != nil {
		return err
	}

	oldStatus := cell.Status
	cell.Status = status

	if err := uc.lockerRepo.UpdateCellStatus(ctx, cell); err != nil {
		return err
	}

	uc.logger.Info("Cell status updated", nil, map[string]any{
		"cellID":    cellID,
		"oldStatus": oldStatus,
		"newStatus": status,
	})

	return nil
}
