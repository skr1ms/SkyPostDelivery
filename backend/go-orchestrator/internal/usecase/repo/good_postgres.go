package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase/repo/sqlc"
)

type GoodRepo struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func NewGoodRepo(db *pgxpool.Pool) *GoodRepo {
	return &GoodRepo{db: db, q: sqlc.New(db)}
}

func toEntityGood(g sqlc.Good) *entity.Good {
	return &entity.Good{
		ID:                g.ID,
		Name:              g.Name,
		Weight:            g.Weight,
		Height:            g.Height,
		Length:            g.Length,
		Width:             g.Width,
		QuantityAvailable: int(g.QuantityAvailable),
	}
}

func (r *GoodRepo) Create(ctx context.Context, name string, weight, height, length, width float64, quantity int) (*entity.Good, error) {
	g, err := r.q.CreateGood(ctx, sqlc.CreateGoodParams{
		Name:              name,
		Weight:            weight,
		Height:            height,
		Length:            length,
		Width:             width,
		QuantityAvailable: int32(quantity),
	})
	if err != nil {
		return nil, fmt.Errorf("GoodRepo - Create - q.CreateGood: %w", err)
	}
	return toEntityGood(g), nil
}

func (r *GoodRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Good, error) {
	g, err := r.q.GetGoodByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GoodRepo - GetByID - q.GetGoodByID: %w", err)
	}
	return toEntityGood(g), nil
}

func (r *GoodRepo) List(ctx context.Context) ([]*entity.Good, error) {
	rows, err := r.q.ListGoods(ctx)
	if err != nil {
		return nil, fmt.Errorf("GoodRepo - List - q.ListGoods: %w", err)
	}
	goods := make([]*entity.Good, 0, len(rows))
	for _, g := range rows {
		goods = append(goods, toEntityGood(g))
	}
	return goods, nil
}

func (r *GoodRepo) Update(ctx context.Context, id uuid.UUID, name string, weight, height, length, width float64) (*entity.Good, error) {
	g, err := r.q.UpdateGood(ctx, sqlc.UpdateGoodParams{
		ID:     id,
		Name:   name,
		Weight: weight,
		Height: height,
		Length: length,
		Width:  width,
	})
	if err != nil {
		return nil, fmt.Errorf("GoodRepo - Update - q.UpdateGood: %w", err)
	}
	return toEntityGood(g), nil
}

func (r *GoodRepo) ListAvailable(ctx context.Context) ([]*entity.Good, error) {
	rows, err := r.q.ListAvailableGoods(ctx)
	if err != nil {
		return nil, fmt.Errorf("GoodRepo - ListAvailable - q.ListAvailableGoods: %w", err)
	}
	goods := make([]*entity.Good, 0, len(rows))
	for _, g := range rows {
		goods = append(goods, toEntityGood(g))
	}
	return goods, nil
}

func (r *GoodRepo) UpdateQuantity(ctx context.Context, id uuid.UUID, delta int) (*entity.Good, error) {
	g, err := r.q.UpdateGoodQuantity(ctx, sqlc.UpdateGoodQuantityParams{
		ID:                id,
		QuantityAvailable: int32(delta),
	})
	if err != nil {
		return nil, fmt.Errorf("GoodRepo - UpdateQuantity - q.UpdateGoodQuantity: %w", err)
	}
	return toEntityGood(g), nil
}

func (r *GoodRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeleteGood(ctx, id); err != nil {
		return fmt.Errorf("GoodRepo - Delete - q.DeleteGood: %w", err)
	}
	return nil
}
