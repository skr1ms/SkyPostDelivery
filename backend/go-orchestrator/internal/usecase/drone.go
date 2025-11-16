package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
)

type DroneUseCase struct {
	droneRepo DroneRepo
}

func NewDroneUseCase(
	droneRepo DroneRepo,
) *DroneUseCase {
	return &DroneUseCase{
		droneRepo: droneRepo,
	}
}

func (uc *DroneUseCase) GetStatus(ctx context.Context, droneID uuid.UUID) (*entity.DroneStatus, error) {
	drone, err := uc.droneRepo.GetByID(ctx, droneID)
	if err != nil {
		return nil, fmt.Errorf("drone usecase - GetStatus - droneRepo.GetByID: %w", err)
	}

	return &entity.DroneStatus{
		DroneID: drone.ID,
		Status:  drone.Status,
	}, nil
}

func (uc DroneUseCase) GetDroneByID(ctx context.Context, droneID uuid.UUID) (*entity.Drone, error) {
	drone, err := uc.droneRepo.GetByID(ctx, droneID)
	if err != nil {
		return nil, fmt.Errorf("drone usecase - GetDroneByID - droneRepo.GetByID: %w", err)
	}
	return drone, nil
}

func (uc *DroneUseCase) ListDrones(ctx context.Context) ([]*entity.Drone, error) {
	drones, err := uc.droneRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("drone usecase - ListDrones - droneRepo.List: %w", err)
	}
	return drones, nil
}

func (uc *DroneUseCase) UpdateDroneStatus(ctx context.Context, droneID uuid.UUID, status string) error {
	if err := uc.droneRepo.UpdateStatus(ctx, droneID, status); err != nil {
		return fmt.Errorf("drone usecase - UpdateDroneStatus - droneRepo.UpdateStatus: %w", err)
	}

	return nil
}

func (uc *DroneUseCase) CreateDrone(ctx context.Context, model string, ipAddress string) (*entity.Drone, error) {
	drone, err := uc.droneRepo.Create(ctx, model, "idle", ipAddress)
	if err != nil {
		return nil, fmt.Errorf("drone usecase - CreateDrone - droneRepo.Create: %w", err)
	}

	return drone, nil
}

func (uc *DroneUseCase) UpdateDrone(ctx context.Context, droneID uuid.UUID, status string) (*entity.Drone, error) {
	if err := uc.droneRepo.UpdateStatus(ctx, droneID, status); err != nil {
		return nil, fmt.Errorf("drone usecase - UpdateDrone - droneRepo.UpdateStatus: %w", err)
	}

	drone, err := uc.droneRepo.GetByID(ctx, droneID)
	if err != nil {
		return nil, fmt.Errorf("drone usecase - UpdateDrone - droneRepo.GetByID: %w", err)
	}

	return drone, nil
}

func (uc *DroneUseCase) Update(ctx context.Context, droneID uuid.UUID, model string) (*entity.Drone, error) {
	drone, err := uc.droneRepo.Update(ctx, droneID, model, "")
	if err != nil {
		return nil, fmt.Errorf("drone usecase - Update - droneRepo.Update: %w", err)
	}

	return drone, nil
}

func (uc *DroneUseCase) UpdateDroneInfo(ctx context.Context, droneID uuid.UUID, model string, ipAddress string) (*entity.Drone, error) {
	drone, err := uc.droneRepo.Update(ctx, droneID, model, ipAddress)
	if err != nil {
		return nil, fmt.Errorf("drone usecase - UpdateDroneInfo - droneRepo.Update: %w", err)
	}

	return drone, nil
}

func (uc *DroneUseCase) DeleteDrone(ctx context.Context, droneID uuid.UUID) error {
	if err := uc.droneRepo.Delete(ctx, droneID); err != nil {
		return fmt.Errorf("drone usecase - DeleteDrone - droneRepo.Delete: %w", err)
	}

	return nil
}
