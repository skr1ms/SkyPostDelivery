package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/logger"
)

type GoodUseCase struct {
	goodRepo repo.GoodRepo
	logger   logger.Interface
}

func NewGoodUseCase(goodRepo repo.GoodRepo, logger logger.Interface) *GoodUseCase {
	return &GoodUseCase{
		goodRepo: goodRepo,
		logger:   logger,
	}
}

func (uc *GoodUseCase) validateGoodParams(name string, weight, height, length, width float64) error {
	if name == "" {
		return entityError.ErrGoodInvalidName
	}

	if weight <= 0 {
		return entityError.ErrGoodInvalidDimensions
	}

	if height <= 0 || length <= 0 || width <= 0 {
		return entityError.ErrGoodInvalidDimensions
	}

	const maxDimension = 1000.0
	const maxWeight = 100.0

	if height > maxDimension || length > maxDimension || width > maxDimension {
		return entityError.ErrGoodInvalidDimensions
	}

	if weight > maxWeight {
		return entityError.ErrGoodInvalidDimensions
	}

	return nil
}

func (uc *GoodUseCase) List(ctx context.Context) ([]*entity.Good, error) {
	goods, err := uc.goodRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("GoodUseCase - List: %w", err)
	}
	return goods, nil
}

func (uc *GoodUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Good, error) {
	good, err := uc.goodRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GoodUseCase - GetByID: %w", err)
	}
	return good, nil
}

func (uc *GoodUseCase) Create(ctx context.Context, name string, weight, height, length, width float64) (*entity.Good, error) {
	if err := uc.validateGoodParams(name, weight, height, length, width); err != nil {
		return nil, err
	}

	good := &entity.Good{
		Name:              name,
		Weight:            weight,
		Height:            height,
		Length:            length,
		Width:             width,
		QuantityAvailable: 1,
	}

	createdGood, err := uc.goodRepo.Create(ctx, good)
	if err != nil {
		return nil, fmt.Errorf("GoodUseCase - Create: %w", err)
	}

	uc.logger.Info("Good created", nil, map[string]any{
		"goodID": createdGood.ID,
		"name":   createdGood.Name,
	})

	return createdGood, nil
}

func (uc *GoodUseCase) CreateWithQuantity(ctx context.Context, name string, weight, height, length, width float64, quantity int) (*entity.Good, error) {
	if err := uc.validateGoodParams(name, weight, height, length, width); err != nil {
		return nil, err
	}

	if quantity <= 0 {
		return nil, entityError.ErrGoodInvalidQuantity
	}

	good := &entity.Good{
		Name:              name,
		Weight:            weight,
		Height:            height,
		Length:            length,
		Width:             width,
		QuantityAvailable: quantity,
	}

	createdGood, err := uc.goodRepo.Create(ctx, good)
	if err != nil {
		return nil, fmt.Errorf("GoodUseCase - CreateWithQuantity: %w", err)
	}

	uc.logger.Info("Good created with quantity", nil, map[string]any{
		"goodID":   createdGood.ID,
		"name":     createdGood.Name,
		"quantity": quantity,
	})

	return createdGood, nil
}

func (uc *GoodUseCase) Update(ctx context.Context, id uuid.UUID, name string, weight, height, length, width float64) (*entity.Good, error) {
	if err := uc.validateGoodParams(name, weight, height, length, width); err != nil {
		return nil, err
	}

	good := &entity.Good{
		ID:     id,
		Name:   name,
		Weight: weight,
		Height: height,
		Length: length,
		Width:  width,
	}

	updatedGood, err := uc.goodRepo.Update(ctx, good)
	if err != nil {
		return nil, fmt.Errorf("GoodUseCase - Update: %w", err)
	}

	uc.logger.Info("Good updated", nil, map[string]any{
		"goodID": updatedGood.ID,
		"name":   updatedGood.Name,
	})

	return updatedGood, nil
}

func (uc *GoodUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	if err := uc.goodRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("GoodUseCase - Delete: %w", err)
	}

	uc.logger.Info("Good deleted", nil, map[string]any{
		"goodID": id,
	})

	return nil
}
