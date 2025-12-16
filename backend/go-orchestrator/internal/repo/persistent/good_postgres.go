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

func (r *GoodRepo) Create(ctx context.Context, good *entity.Good) (*entity.Good, error) {
	g, err := r.q.CreateGood(ctx, sqlc.CreateGoodParams{
		Name:              good.Name,
		Weight:            good.Weight,
		Height:            good.Height,
		Length:            good.Length,
		Width:             good.Width,
		QuantityAvailable: int32(good.QuantityAvailable),
	})
	if err != nil {
		if isPgUniqueViolation(err) {
			return nil, entityError.ErrGoodCreateFailed
		}
		return nil, fmt.Errorf("GoodRepo - Create: %w", err)
	}
	return toEntityGood(g), nil
}

func (r *GoodRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Good, error) {
	g, err := r.q.GetGoodByID(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrGoodNotFound
		}
		return nil, fmt.Errorf("GoodRepo - GetByID: %w", err)
	}
	return toEntityGood(g), nil
}

func (r *GoodRepo) List(ctx context.Context) ([]*entity.Good, error) {
	rows, err := r.q.ListGoods(ctx)
	if err != nil {
		return nil, fmt.Errorf("GoodRepo - List: %w", err)
	}
	goods := make([]*entity.Good, 0, len(rows))
	for _, g := range rows {
		goods = append(goods, toEntityGood(g))
	}
	return goods, nil
}

func (r *GoodRepo) Update(ctx context.Context, good *entity.Good) (*entity.Good, error) {
	g, err := r.q.UpdateGood(ctx, sqlc.UpdateGoodParams{
		ID:     good.ID,
		Name:   good.Name,
		Weight: good.Weight,
		Height: good.Height,
		Length: good.Length,
		Width:  good.Width,
	})
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrGoodNotFound
		}
		return nil, fmt.Errorf("GoodRepo - Update: %w", err)
	}
	return toEntityGood(g), nil
}

func (r *GoodRepo) ListAvailable(ctx context.Context) ([]*entity.Good, error) {
	rows, err := r.q.ListAvailableGoods(ctx)
	if err != nil {
		return nil, fmt.Errorf("GoodRepo - ListAvailable: %w", err)
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
		if isNoRows(err) {
			return nil, entityError.ErrGoodNotFound
		}
		return nil, fmt.Errorf("GoodRepo - UpdateQuantity: %w", err)
	}
	return toEntityGood(g), nil
}

func (r *GoodRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeleteGood(ctx, id); err != nil {
		if isNoRows(err) {
			return entityError.ErrGoodNotFound
		}
		return fmt.Errorf("GoodRepo - Delete: %w", err)
	}
	return nil
}
