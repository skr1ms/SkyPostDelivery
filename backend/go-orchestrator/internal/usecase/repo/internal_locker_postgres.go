package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase/repo/sqlc"
)

type InternalLockerRepo struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func NewInternalLockerRepo(db *pgxpool.Pool) *InternalLockerRepo {
	return &InternalLockerRepo{db: db, q: sqlc.New(db)}
}

func toEntityInternalLockerCell(c sqlc.LockerCellsInternal) *entity.LockerCell {
	return &entity.LockerCell{
		ID:     c.ID,
		PostID: c.PostID,
		Height: c.Height,
		Length: c.Length,
		Width:  c.Width,
		Status: c.Status,
	}
}

func (r *InternalLockerRepo) Create(ctx context.Context, postID uuid.UUID, height, length, width float64) (*entity.LockerCell, error) {
	cell, err := r.q.CreateInternalLockerCell(ctx, sqlc.CreateInternalLockerCellParams{
		PostID:     postID,
		Height:     height,
		Length:     length,
		Width:      width,
		Status:     "available",
		CellNumber: nil,
	})
	if err != nil {
		return nil, fmt.Errorf("InternalLockerRepo - Create - q.CreateInternalLockerCell: %w", err)
	}
	return toEntityInternalLockerCell(cell), nil
}

func (r *InternalLockerRepo) CreateWithNumber(ctx context.Context, postID uuid.UUID, height, length, width float64, cellNumber int) (*entity.LockerCell, error) {
	num := int32(cellNumber)
	cell, err := r.q.CreateInternalLockerCell(ctx, sqlc.CreateInternalLockerCellParams{
		PostID:     postID,
		Height:     height,
		Length:     length,
		Width:      width,
		Status:     "available",
		CellNumber: &num,
	})
	if err != nil {
		return nil, fmt.Errorf("InternalLockerRepo - CreateWithNumber - q.CreateInternalLockerCell: %w", err)
	}
	return toEntityInternalLockerCell(cell), nil
}

func (r *InternalLockerRepo) CreateCell(ctx context.Context, postID uuid.UUID, height, length, width float64, status string) (*entity.LockerCell, error) {
	cell, err := r.q.CreateInternalLockerCell(ctx, sqlc.CreateInternalLockerCellParams{
		PostID:     postID,
		Height:     height,
		Length:     length,
		Width:      width,
		Status:     status,
		CellNumber: nil,
	})
	if err != nil {
		return nil, fmt.Errorf("InternalLockerRepo - CreateCell - q.CreateInternalLockerCell: %w", err)
	}
	return toEntityInternalLockerCell(cell), nil
}

func (r *InternalLockerRepo) GetCellByID(ctx context.Context, id uuid.UUID) (*entity.LockerCell, error) {
	cell, err := r.q.GetInternalLockerCellByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("InternalLockerRepo - GetCellByID - q.GetInternalLockerCellByID: %w", err)
	}
	return toEntityInternalLockerCell(cell), nil
}

func (r *InternalLockerRepo) FindAvailableCell(ctx context.Context, height, length, width float64) (*entity.LockerCell, error) {
	cell, err := r.q.FindAvailableInternalCell(ctx, sqlc.FindAvailableInternalCellParams{
		Height: height,
		Length: length,
		Width:  width,
	})
	if err != nil {
		return nil, fmt.Errorf("InternalLockerRepo - FindAvailableCell - q.FindAvailableInternalCell: %w", err)
	}
	return toEntityInternalLockerCell(cell), nil
}

func (r *InternalLockerRepo) UpdateCellStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.q.UpdateInternalLockerCellStatus(ctx, sqlc.UpdateInternalLockerCellStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return fmt.Errorf("InternalLockerRepo - UpdateCellStatus - q.UpdateInternalLockerCellStatus: %w", err)
	}
	return nil
}

func (r *InternalLockerRepo) UpdateDimensions(ctx context.Context, id uuid.UUID, height, length, width float64) (*entity.LockerCell, error) {
	cell, err := r.q.UpdateInternalLockerCellDimensions(ctx, sqlc.UpdateInternalLockerCellDimensionsParams{
		ID:     id,
		Height: height,
		Length: length,
		Width:  width,
	})
	if err != nil {
		return nil, fmt.Errorf("InternalLockerRepo - UpdateDimensions - q.UpdateInternalLockerCellDimensions: %w", err)
	}
	return toEntityInternalLockerCell(cell), nil
}

func (r *InternalLockerRepo) ListCellsByPostID(ctx context.Context, postID uuid.UUID) ([]*entity.LockerCell, error) {
	rows, err := r.q.ListInternalLockerCellsByPostID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("InternalLockerRepo - ListCellsByPostID - q.ListInternalLockerCellsByPostID: %w", err)
	}

	result := make([]*entity.LockerCell, 0, len(rows))
	for _, row := range rows {
		result = append(result, toEntityInternalLockerCell(row))
	}

	return result, nil
}
