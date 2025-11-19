package usecase

import (
	"context"
	"fmt"
	"sync"

	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/repo"
)

type DroneManagerUseCase struct {
	droneRepo        repo.DroneRepo
	registeredDrones map[string]bool
	mu               sync.RWMutex
}

func NewDroneManagerUseCase(droneRepo repo.DroneRepo) *DroneManagerUseCase {
	return &DroneManagerUseCase{
		droneRepo:        droneRepo,
		registeredDrones: make(map[string]bool),
	}
}

func (dm *DroneManagerUseCase) RegisterDrone(ctx context.Context, droneID string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.registeredDrones[droneID] = true
	return nil
}

func (dm *DroneManagerUseCase) GetFreeDrone(ctx context.Context) (string, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	for droneID := range dm.registeredDrones {
		state, err := dm.droneRepo.GetDroneState(ctx, droneID)
		if err != nil {
			continue
		}

		if state != nil && state.Status == entity.DroneStatusIdle && state.CurrentDeliveryID == nil {
			if state.BatteryLevel > 30.0 {
				return droneID, nil
			}
		}
	}

	return "", nil
}

func (dm *DroneManagerUseCase) AssignDeliveryToDrone(ctx context.Context, droneID string, deliveryID string) error {
	state, err := dm.droneRepo.GetDroneState(ctx, droneID)
	if err != nil {
		return fmt.Errorf("drone manager usecase - AssignDeliveryToDrone - droneStateRepo.GetDroneState: %w", err)
	}

	if state != nil {
		state.CurrentDeliveryID = &deliveryID
		state.Status = entity.DroneStatusTakingOff
		if err := dm.droneRepo.SaveDroneState(ctx, state); err != nil {
			return fmt.Errorf("drone manager usecase - AssignDeliveryToDrone - droneStateRepo.SaveDroneState: %w", err)
		}
	}

	return nil
}

func (dm *DroneManagerUseCase) ReleaseDrone(ctx context.Context, droneID string) error {
	state, err := dm.droneRepo.GetDroneState(ctx, droneID)
	if err != nil {
		return fmt.Errorf("drone manager usecase - ReleaseDrone - droneStateRepo.GetDroneState: %w", err)
	}

	if state != nil {
		state.CurrentDeliveryID = nil
		state.Status = entity.DroneStatusIdle
		if err := dm.droneRepo.SaveDroneState(ctx, state); err != nil {
			return fmt.Errorf("drone manager usecase - ReleaseDrone - droneStateRepo.SaveDroneState: %w", err)
		}
	}

	return nil
}

func (dm *DroneManagerUseCase) UnregisterDrone(ctx context.Context, droneID string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	delete(dm.registeredDrones, droneID)
	return nil
}

func (dm *DroneManagerUseCase) GetDroneState(ctx context.Context, droneID string) (*entity.DroneState, error) {
	state, err := dm.droneRepo.GetDroneState(ctx, droneID)
	if err != nil {
		return nil, fmt.Errorf("drone manager usecase - GetDroneState - droneRepo.GetDroneState: %w", err)
	}
	return state, nil
}

func (dm *DroneManagerUseCase) GetAllDrones() []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	drones := make([]string, 0, len(dm.registeredDrones))
	for droneID := range dm.registeredDrones {
		drones = append(drones, droneID)
	}
	return drones
}

func (dm *DroneManagerUseCase) GetRegisteredDrones() []string {
	return dm.GetAllDrones()
}
