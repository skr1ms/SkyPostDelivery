package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
)

type GoodUseCase struct {
	goodRepo GoodRepo
}

func NewGoodUseCase(goodRepo GoodRepo) *GoodUseCase {
	return &GoodUseCase{
		goodRepo: goodRepo,
	}
}

func (uc *GoodUseCase) ListGoods(ctx context.Context) ([]*entity.Good, error) {
	goods, err := uc.goodRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("GoodUseCase - ListGoods - goodRepo.List: %w", err)
	}
	return goods, nil
}

func (uc *GoodUseCase) GetGood(ctx context.Context, id uuid.UUID) (*entity.Good, error) {
	good, err := uc.goodRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GoodUseCase - GetGood - goodRepo.GetByID: %w", err)
	}
	return good, nil
}

func (uc *GoodUseCase) CreateGood(ctx context.Context, name string, weight, height, length, width float64) (*entity.Good, error) {
	good, err := uc.goodRepo.Create(ctx, name, weight, height, length, width, 1)
	if err != nil {
		return nil, fmt.Errorf("GoodUseCase - CreateGood - goodRepo.Create: %w", err)
	}

	return good, nil
}

func (uc *GoodUseCase) CreateGoods(ctx context.Context, name string, weight, height, length, width float64, quantity int) ([]*entity.Good, error) {
	good, err := uc.goodRepo.Create(ctx, name, weight, height, length, width, quantity)
	if err != nil {
		return nil, fmt.Errorf("GoodUseCase - CreateGoods - goodRepo.Create: %w", err)
	}

	return []*entity.Good{good}, nil
}

func (uc *GoodUseCase) UpdateGood(ctx context.Context, id uuid.UUID, name string, weight, height, length, width float64) (*entity.Good, error) {
	good, err := uc.goodRepo.Update(ctx, id, name, weight, height, length, width)
	if err != nil {
		return nil, fmt.Errorf("GoodUseCase - UpdateGood - goodRepo.Update: %w", err)
	}
	return good, nil
}

func (uc *GoodUseCase) DeleteGood(ctx context.Context, id uuid.UUID) error {
	if err := uc.goodRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("GoodUseCase - DeleteGood - goodRepo.Delete: %w", err)
	}
	return nil
}
