package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/hitech-ekb/internal/entity"
	"github.com/skr1ms/hitech-ekb/internal/usecase/repo/sqlc"
)

type LockerRepo struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func NewLockerRepo(db *pgxpool.Pool) *LockerRepo {
	return &LockerRepo{db: db, q: sqlc.New(db)}
}

func toEntityLockerCell(c sqlc.LockerCell) *entity.LockerCell {
	return &entity.LockerCell{
		ID:     c.ID,
		PostID: c.PostID,
		Height: c.Height,
		Length: c.Length,
		Width:  c.Width,
		Status: c.Status,
	}
}

func (r *LockerRepo) Create(ctx context.Context, postID uuid.UUID, height, length, width float64) (*entity.LockerCell, error) {
	c, err := r.q.CreateLockerCell(ctx, sqlc.CreateLockerCellParams{
		PostID: postID,
		Height: height,
		Length: length,
		Width:  width,
		Status: "available",
	})
	if err != nil {
		return nil, fmt.Errorf("LockerRepo - Create - q.CreateLockerCell: %w", err)
	}
	return toEntityLockerCell(c), nil
}

func (r *LockerRepo) CreateCell(ctx context.Context, postID uuid.UUID, height, length, width float64, status string) (*entity.LockerCell, error) {
	c, err := r.q.CreateLockerCell(ctx, sqlc.CreateLockerCellParams{
		PostID: postID,
		Height: height,
		Length: length,
		Width:  width,
		Status: status,
	})
	if err != nil {
		return nil, fmt.Errorf("LockerRepo - CreateCell - q.CreateLockerCell: %w", err)
	}
	return toEntityLockerCell(c), nil
}

func (r *LockerRepo) GetCellByID(ctx context.Context, id uuid.UUID) (*entity.LockerCell, error) {
	c, err := r.q.GetLockerCellByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("LockerRepo - GetCellByID - q.GetLockerCellByID: %w", err)
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
		return nil, fmt.Errorf("LockerRepo - FindAvailableCell - q.FindAvailableCell: %w", err)
	}
	return toEntityLockerCell(c), nil
}

func (r *LockerRepo) UpdateCellStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.q.UpdateLockerCellStatus(ctx, sqlc.UpdateLockerCellStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return fmt.Errorf("LockerRepo - UpdateCellStatus - q.UpdateLockerCellStatus: %w", err)
	}
	return nil
}

func (r *LockerRepo) UpdateDimensions(ctx context.Context, id uuid.UUID, height, length, width float64) (*entity.LockerCell, error) {
	c, err := r.q.UpdateLockerCellDimensions(ctx, sqlc.UpdateLockerCellDimensionsParams{
		ID:     id,
		Height: height,
		Length: length,
		Width:  width,
	})
	if err != nil {
		return nil, fmt.Errorf("LockerRepo - UpdateDimensions - q.UpdateLockerCellDimensions: %w", err)
	}
	return toEntityLockerCell(c), nil
}

func (r *LockerRepo) ListCellsByPostID(ctx context.Context, postID uuid.UUID) ([]*entity.LockerCell, error) {
	rows, err := r.q.ListLockerCellsByPostID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("LockerRepo - ListCellsByPostID - q.ListLockerCellsByPostID: %w", err)
	}
	cells := make([]*entity.LockerCell, 0, len(rows))
	for _, c := range rows {
		cells = append(cells, toEntityLockerCell(c))
	}
	return cells, nil
}
