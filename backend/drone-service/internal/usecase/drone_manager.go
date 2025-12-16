package usecase

import (
	"context"
	"fmt"
	"sync"

	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/logger"
)

type DroneManagerUseCase struct {
	droneRepo        repo.DroneRepo
	registeredDrones map[string]bool
	mu               sync.RWMutex
	logger           logger.Interface
}

func NewDroneManagerUseCase(droneRepo repo.DroneRepo, logger logger.Interface) *DroneManagerUseCase {
	return &DroneManagerUseCase{
		droneRepo:        droneRepo,
		registeredDrones: make(map[string]bool),
		logger:           logger,
	}
}

func (uc *DroneManagerUseCase) RegisterDrone(ctx context.Context, droneID string) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	uc.registeredDrones[droneID] = true
	return nil
}

func (uc *DroneManagerUseCase) GetFreeDrone(ctx context.Context) (string, error) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	for droneID := range uc.registeredDrones {
		state, err := uc.droneRepo.GetDroneState(ctx, droneID)
		if err != nil {
			uc.logger.Warn("DroneManagerUseCase - GetFreeDrone - GetDroneState", err, map[string]any{
				"droneID": droneID,
			})
			continue
		}

		if state != nil && state.Status == entity.DroneStatusIdle && state.CurrentDeliveryID == nil {
			if state.BatteryLevel > 30.0 {
				return droneID, nil
			}
		}
	}

	return "", entityError.ErrDroneNotAvailable
}

func (uc *DroneManagerUseCase) AssignDeliveryToDrone(ctx context.Context, droneID string, deliveryID string) error {
	state, err := uc.droneRepo.GetDroneState(ctx, droneID)
	if err != nil {
		uc.logger.Error("DroneManagerUseCase - AssignDeliveryToDrone - GetDroneState", err, map[string]any{
			"droneID":    droneID,
			"deliveryID": deliveryID,
		})
		return fmt.Errorf("DroneManagerUseCase - AssignDeliveryToDrone - GetDroneState: %w", err)
	}

	if state == nil {
		uc.logger.Warn("DroneManagerUseCase - AssignDeliveryToDrone - drone state not found", entityError.ErrDroneStateNotFound, map[string]any{
			"droneID": droneID,
		})
		return entityError.ErrDroneStateNotFound
	}

	state.CurrentDeliveryID = &deliveryID
	state.Status = entity.DroneStatusTakingOff
	if err := uc.droneRepo.SaveDroneState(ctx, state); err != nil {
		uc.logger.Error("DroneManagerUseCase - AssignDeliveryToDrone - SaveDroneState", err, map[string]any{
			"droneID":    droneID,
			"deliveryID": deliveryID,
		})
		return fmt.Errorf("DroneManagerUseCase - AssignDeliveryToDrone - SaveDroneState: %w", err)
	}

	uc.logger.Info("Delivery assigned to drone", nil, map[string]any{
		"droneID":    droneID,
		"deliveryID": deliveryID,
	})

	return nil
}

func (uc *DroneManagerUseCase) ReleaseDrone(ctx context.Context, droneID string) error {
	state, err := uc.droneRepo.GetDroneState(ctx, droneID)
	if err != nil {
		uc.logger.Error("DroneManagerUseCase - ReleaseDrone - GetDroneState", err, map[string]any{
			"droneID": droneID,
		})
		return fmt.Errorf("DroneManagerUseCase - ReleaseDrone - GetDroneState: %w", err)
	}

	if state == nil {
		uc.logger.Warn("DroneManagerUseCase - ReleaseDrone - drone state not found", entityError.ErrDroneStateNotFound, map[string]any{
			"droneID": droneID,
		})
		return entityError.ErrDroneStateNotFound
	}

	state.CurrentDeliveryID = nil
	state.Status = entity.DroneStatusIdle
	if err := uc.droneRepo.SaveDroneState(ctx, state); err != nil {
		uc.logger.Error("DroneManagerUseCase - ReleaseDrone - SaveDroneState", err, map[string]any{
			"droneID": droneID,
		})
		return fmt.Errorf("DroneManagerUseCase - ReleaseDrone - SaveDroneState: %w", err)
	}

	uc.logger.Info("Drone released", nil, map[string]any{
		"droneID": droneID,
	})

	return nil
}

func (uc *DroneManagerUseCase) UnregisterDrone(ctx context.Context, droneID string) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	delete(uc.registeredDrones, droneID)
	return nil
}

func (uc *DroneManagerUseCase) GetDroneState(ctx context.Context, droneID string) (*entity.DroneState, error) {
	state, err := uc.droneRepo.GetDroneState(ctx, droneID)
	if err != nil {
		uc.logger.Error("DroneManagerUseCase - GetDroneState - GetDroneState", err, map[string]any{
			"droneID": droneID,
		})
		return nil, fmt.Errorf("DroneManagerUseCase - GetDroneState - GetDroneState: %w", err)
	}
	return state, nil
}

func (uc *DroneManagerUseCase) GetAllDrones() []string {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	drones := make([]string, 0, len(uc.registeredDrones))
	for droneID := range uc.registeredDrones {
		drones = append(drones, droneID)
	}
	return drones
}

func (uc *DroneManagerUseCase) GetRegisteredDrones() []string {
	return uc.GetAllDrones()
}
