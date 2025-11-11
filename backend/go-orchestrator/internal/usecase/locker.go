package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/skr1ms/hitech-ekb/internal/entity"
)

type LockerUseCase struct {
	lockerRepo LockerRepo
}

func NewLockerUseCase(
	lockerRepo LockerRepo,
) *LockerUseCase {
	return &LockerUseCase{
		lockerRepo: lockerRepo,
	}
}

func (uc *LockerUseCase) GetCell(ctx context.Context, id uuid.UUID) (*entity.LockerCell, error) {
	cell, err := uc.lockerRepo.GetCellByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("locker usecase - GetCell - lockerRepo.GetCellByID: %w", err)
	}
	return cell, nil
}

func (uc *LockerUseCase) ListCells(ctx context.Context, postID uuid.UUID) ([]*entity.LockerCell, error) {
	cells, err := uc.lockerRepo.ListCellsByPostID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("locker usecase - ListCells - lockerRepo.ListCellsByPostID: %w", err)
	}
	return cells, nil
}

func (uc *LockerUseCase) UpdateCellStatus(ctx context.Context, cellID uuid.UUID, status string) error {
	if err := uc.lockerRepo.UpdateCellStatus(ctx, cellID, status); err != nil {
		return fmt.Errorf("locker usecase - UpdateCellStatus - lockerRepo.UpdateCellStatus: %w", err)
	}

	return nil
}
