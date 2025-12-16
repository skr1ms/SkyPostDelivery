package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo/persistent/sqlc"
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

func (r *InternalLockerRepo) Create(ctx context.Context, cell *entity.LockerCell) (*entity.LockerCell, error) {
	createdCell, err := r.q.CreateInternalLockerCell(ctx, sqlc.CreateInternalLockerCellParams{
		PostID:     cell.PostID,
		Height:     cell.Height,
		Length:     cell.Length,
		Width:      cell.Width,
		Status:     "available",
		CellNumber: nil,
	})
	if err != nil {
		if isPgForeignKeyViolation(err) {
			return nil, entityError.ErrLockerCellCreateFailed
		}
		return nil, fmt.Errorf("InternalLockerRepo - Create: %w", err)
	}
	return toEntityInternalLockerCell(createdCell), nil
}

func (r *InternalLockerRepo) CreateWithNumber(ctx context.Context, cell *entity.LockerCell, cellNumber int) (*entity.LockerCell, error) {
	num := int32(cellNumber)
	createdCell, err := r.q.CreateInternalLockerCell(ctx, sqlc.CreateInternalLockerCellParams{
		PostID:     cell.PostID,
		Height:     cell.Height,
		Length:     cell.Length,
		Width:      cell.Width,
		Status:     "available",
		CellNumber: &num,
	})
	if err != nil {
		if isPgForeignKeyViolation(err) {
			return nil, entityError.ErrLockerCellCreateFailed
		}
		if isPgUniqueViolation(err) {
			return nil, entityError.ErrLockerCellAlreadyExists
		}
		return nil, fmt.Errorf("InternalLockerRepo - CreateWithNumber: %w", err)
	}
	return toEntityInternalLockerCell(createdCell), nil
}

func (r *InternalLockerRepo) CreateCell(ctx context.Context, cell *entity.LockerCell) (*entity.LockerCell, error) {
	createdCell, err := r.q.CreateInternalLockerCell(ctx, sqlc.CreateInternalLockerCellParams{
		PostID:     cell.PostID,
		Height:     cell.Height,
		Length:     cell.Length,
		Width:      cell.Width,
		Status:     cell.Status,
		CellNumber: nil,
	})
	if err != nil {
		if isPgForeignKeyViolation(err) {
			return nil, entityError.ErrLockerCellCreateFailed
		}
		return nil, fmt.Errorf("InternalLockerRepo - CreateCell: %w", err)
	}
	return toEntityInternalLockerCell(createdCell), nil
}

func (r *InternalLockerRepo) GetCellByID(ctx context.Context, id uuid.UUID) (*entity.LockerCell, error) {
	cell, err := r.q.GetInternalLockerCellByID(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrLockerCellNotFound
		}
		return nil, fmt.Errorf("InternalLockerRepo - GetCellByID: %w", err)
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
		if isNoRows(err) {
			return nil, entityError.ErrLockerCellNotFound
		}
		return nil, fmt.Errorf("InternalLockerRepo - FindAvailableCell: %w", err)
	}
	return toEntityInternalLockerCell(cell), nil
}

func (r *InternalLockerRepo) UpdateCellStatus(ctx context.Context, cell *entity.LockerCell) error {
	_, err := r.q.UpdateInternalLockerCellStatus(ctx, sqlc.UpdateInternalLockerCellStatusParams{
		ID:     cell.ID,
		Status: cell.Status,
	})
	if err != nil {
		if isNoRows(err) {
			return entityError.ErrLockerCellNotFound
		}
		return fmt.Errorf("InternalLockerRepo - UpdateCellStatus: %w", err)
	}
	return nil
}

func (r *InternalLockerRepo) UpdateDimensions(ctx context.Context, cell *entity.LockerCell) (*entity.LockerCell, error) {
	updatedCell, err := r.q.UpdateInternalLockerCellDimensions(ctx, sqlc.UpdateInternalLockerCellDimensionsParams{
		ID:     cell.ID,
		Height: cell.Height,
		Length: cell.Length,
		Width:  cell.Width,
	})
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrLockerCellNotFound
		}
		return nil, fmt.Errorf("InternalLockerRepo - UpdateDimensions: %w", err)
	}
	return toEntityInternalLockerCell(updatedCell), nil
}

func (r *InternalLockerRepo) ListCellsByPostID(ctx context.Context, postID uuid.UUID) ([]*entity.LockerCell, error) {
	rows, err := r.q.ListInternalLockerCellsByPostID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("InternalLockerRepo - ListCellsByPostID: %w", err)
	}

	result := make([]*entity.LockerCell, 0, len(rows))
	for _, row := range rows {
		result = append(result, toEntityInternalLockerCell(row))
	}

	return result, nil
}
