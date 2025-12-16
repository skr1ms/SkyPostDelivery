package usecase

import (
	"context"
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/logger"
)

type DroneUseCase struct {
	droneRepo repo.DroneRepo
	logger    logger.Interface
}

func NewDroneUseCase(
	droneRepo repo.DroneRepo,
	logger logger.Interface,
) *DroneUseCase {
	return &DroneUseCase{
		droneRepo: droneRepo,
		logger:    logger,
	}
}

var validDroneStatuses = map[string]bool{
	"idle":      true,
	"busy":      true,
	"returning": true,
	"charging":  true,
	"offline":   true,
}

func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func (uc *DroneUseCase) GetByID(ctx context.Context, droneID uuid.UUID) (*entity.Drone, error) {
	drone, err := uc.droneRepo.GetByID(ctx, droneID)
	if err != nil {
		return nil, fmt.Errorf("DroneUseCase - GetByID: %w", err)
	}
	return drone, nil
}

func (uc *DroneUseCase) GetStatus(ctx context.Context, droneID uuid.UUID) (*entity.DroneStatus, error) {
	drone, err := uc.droneRepo.GetByID(ctx, droneID)
	if err != nil {
		return nil, fmt.Errorf("DroneUseCase - GetStatus: %w", err)
	}

	return &entity.DroneStatus{
		DroneID: drone.ID,
		Status:  drone.Status,
	}, nil
}

func (uc *DroneUseCase) List(ctx context.Context) ([]*entity.Drone, error) {
	drones, err := uc.droneRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("DroneUseCase - List: %w", err)
	}
	return drones, nil
}

func (uc *DroneUseCase) Create(ctx context.Context, model, ipAddress string) (*entity.Drone, error) {
	if model == "" {
		return nil, entityError.ErrDroneInvalidModel
	}

	if ipAddress == "" {
		return nil, entityError.ErrDroneInvalidIP
	}

	if !isValidIP(ipAddress) {
		return nil, entityError.ErrDroneInvalidIP
	}

	drone := &entity.Drone{
		Model:     model,
		Status:    "idle",
		IPAddress: ipAddress,
	}

	createdDrone, err := uc.droneRepo.Create(ctx, drone)
	if err != nil {
		return nil, fmt.Errorf("DroneUseCase - Create: %w", err)
	}

	uc.logger.Info("Drone created", nil, map[string]any{
		"droneID":   createdDrone.ID,
		"model":     createdDrone.Model,
		"ipAddress": createdDrone.IPAddress,
	})

	return createdDrone, nil
}

func (uc *DroneUseCase) Update(ctx context.Context, droneID uuid.UUID, model, ipAddress, status string) (*entity.Drone, error) {
	if model == "" && ipAddress == "" && status == "" {
		return nil, entityError.ErrDroneNothingToUpdate
	}

	if ipAddress != "" && !isValidIP(ipAddress) {
		return nil, entityError.ErrDroneInvalidIP
	}

	if status != "" && !validDroneStatuses[status] {
		return nil, entityError.ErrDroneInvalidStatus
	}

	drone, err := uc.droneRepo.GetByID(ctx, droneID)
	if err != nil {
		return nil, fmt.Errorf("DroneUseCase - Update: %w", err)
	}

	if model != "" {
		drone.Model = model
	}
	if ipAddress != "" {
		drone.IPAddress = ipAddress
	}
	if status != "" {
		drone.Status = status
	}

	updatedDrone, err := uc.droneRepo.Update(ctx, drone)
	if err != nil {
		return nil, fmt.Errorf("DroneUseCase - Update: %w", err)
	}

	uc.logger.Info("Drone updated", nil, map[string]any{
		"droneID": droneID,
		"model":   updatedDrone.Model,
		"status":  updatedDrone.Status,
	})

	return updatedDrone, nil
}

func (uc *DroneUseCase) UpdateStatus(ctx context.Context, droneID uuid.UUID, status string) error {
	if !validDroneStatuses[status] {
		return entityError.ErrDroneInvalidStatus
	}

	drone, err := uc.droneRepo.GetByID(ctx, droneID)
	if err != nil {
		return fmt.Errorf("DroneUseCase - UpdateStatus: %w", err)
	}

	if drone.Status == status {
		uc.logger.Debug("Drone status already set, skipping update", nil, map[string]any{
			"droneID": droneID,
			"status":  status,
		})
		return nil
	}

	oldStatus := drone.Status
	drone.Status = status
	if err := uc.droneRepo.UpdateStatus(ctx, drone); err != nil {
		return fmt.Errorf("DroneUseCase - UpdateStatus: %w", err)
	}

	uc.logger.Info("Drone status updated", nil, map[string]any{
		"droneID":   droneID,
		"oldStatus": oldStatus,
		"newStatus": status,
	})

	return nil
}

func (uc *DroneUseCase) Delete(ctx context.Context, droneID uuid.UUID) error {
	drone, err := uc.droneRepo.GetByID(ctx, droneID)
	if err != nil {
		return fmt.Errorf("DroneUseCase - Delete: %w", err)
	}

	if drone.Status == "busy" {
		return entityError.ErrDroneCannotDelete
	}

	if err := uc.droneRepo.Delete(ctx, droneID); err != nil {
		return fmt.Errorf("DroneUseCase - Delete: %w", err)
	}

	uc.logger.Info("Drone deleted", nil, map[string]any{
		"droneID": droneID,
	})

	return nil
}
