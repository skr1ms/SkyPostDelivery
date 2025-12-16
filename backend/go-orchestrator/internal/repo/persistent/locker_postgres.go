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

type LockerRepo struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func NewLockerRepo(db *pgxpool.Pool) *LockerRepo {
	return &LockerRepo{db: db, q: sqlc.New(db)}
}

func toEntityLockerCell(c sqlc.LockerCellsOut) *entity.LockerCell {
	return &entity.LockerCell{
		ID:     c.ID,
		PostID: c.PostID,
		Height: c.Height,
		Length: c.Length,
		Width:  c.Width,
		Status: c.Status,
	}
}

func (r *LockerRepo) Create(ctx context.Context, cell *entity.LockerCell) (*entity.LockerCell, error) {
	c, err := r.q.CreateLockerCell(ctx, sqlc.CreateLockerCellParams{
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
		return nil, fmt.Errorf("LockerRepo - Create: %w", err)
	}
	return toEntityLockerCell(c), nil
}

func (r *LockerRepo) CreateWithNumber(ctx context.Context, cell *entity.LockerCell, cellNumber int) (*entity.LockerCell, error) {
	num := int32(cellNumber)
	c, err := r.q.CreateLockerCell(ctx, sqlc.CreateLockerCellParams{
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
		return nil, fmt.Errorf("LockerRepo - CreateWithNumber: %w", err)
	}
	return toEntityLockerCell(c), nil
}

func (r *LockerRepo) CreateCell(ctx context.Context, cell *entity.LockerCell) (*entity.LockerCell, error) {
	c, err := r.q.CreateLockerCell(ctx, sqlc.CreateLockerCellParams{
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
		return nil, fmt.Errorf("LockerRepo - CreateCell: %w", err)
	}
	return toEntityLockerCell(c), nil
}

func (r *LockerRepo) GetCellByID(ctx context.Context, id uuid.UUID) (*entity.LockerCell, error) {
	c, err := r.q.GetLockerCellByID(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrLockerCellNotFound
		}
		return nil, fmt.Errorf("LockerRepo - GetCellByID: %w", err)
	}
	return toEntityLockerCell(c), nil
}

func (r *LockerRepo) FindAvailableCell(ctx context.Context, height, length, width float64) (*entity.LockerCell, error) {
	c, err := r.q.FindAvailableCell(ctx, sqlc.FindAvailableCellParams{
		Height: height,
		Length: length,
		Width:  width,
	})
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrLockerCellNotFound
		}
		return nil, fmt.Errorf("LockerRepo - FindAvailableCell: %w", err)
	}
	return toEntityLockerCell(c), nil
}

func (r *LockerRepo) UpdateCellStatus(ctx context.Context, cell *entity.LockerCell) error {
	_, err := r.q.UpdateLockerCellStatus(ctx, sqlc.UpdateLockerCellStatusParams{
		ID:     cell.ID,
		Status: cell.Status,
	})
	if err != nil {
		if isNoRows(err) {
			return entityError.ErrLockerCellNotFound
		}
		return fmt.Errorf("LockerRepo - UpdateCellStatus: %w", err)
	}
	return nil
}

func (r *LockerRepo) UpdateDimensions(ctx context.Context, cell *entity.LockerCell) (*entity.LockerCell, error) {
	c, err := r.q.UpdateLockerCellDimensions(ctx, sqlc.UpdateLockerCellDimensionsParams{
		ID:     cell.ID,
		Height: cell.Height,
		Length: cell.Length,
		Width:  cell.Width,
	})
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrLockerCellNotFound
		}
		return nil, fmt.Errorf("LockerRepo - UpdateDimensions: %w", err)
	}
	return toEntityLockerCell(c), nil
}

func (r *LockerRepo) ListCellsByPostID(ctx context.Context, postID uuid.UUID) ([]*entity.LockerCell, error) {
	rows, err := r.q.ListLockerCellsByPostID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("LockerRepo - ListCellsByPostID: %w", err)
	}
	cells := make([]*entity.LockerCell, 0, len(rows))
	for _, c := range rows {
		cells = append(cells, toEntityLockerCell(c))
	}
	return cells, nil
}
