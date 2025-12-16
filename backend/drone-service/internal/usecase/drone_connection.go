package usecase

import (
	"context"
	"fmt"

	entityError "github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/logger"
)

type DroneConnectionUseCase struct {
	droneRepo repo.DroneRepo
	manager   *DroneManagerUseCase
	logger    logger.Interface
}

func NewDroneConnectionUseCase(
	droneRepo repo.DroneRepo,
	manager *DroneManagerUseCase,
	logger logger.Interface,
) *DroneConnectionUseCase {
	return &DroneConnectionUseCase{
		droneRepo: droneRepo,
		manager:   manager,
		logger:    logger,
	}
}

func (uc *DroneConnectionUseCase) RegisterDrone(ctx context.Context, ipAddress string) (string, error) {
	droneID, err := uc.droneRepo.GetDroneIDByIP(ctx, ipAddress)
	if err != nil {
		uc.logger.Error("DroneConnectionUseCase - RegisterDrone - GetDroneIDByIP", err, map[string]any{
			"ipAddress": ipAddress,
		})
		return "", fmt.Errorf("DroneConnectionUseCase - RegisterDrone - GetDroneIDByIP: %w", err)
	}

	if droneID == "" {
		uc.logger.Warn("DroneConnectionUseCase - RegisterDrone - drone not found", entityError.ErrDroneNotFound, map[string]any{
			"ipAddress": ipAddress,
		})
		return "", entityError.ErrDroneNotFound
	}

	if err := uc.manager.RegisterDrone(ctx, droneID); err != nil {
		uc.logger.Error("DroneConnectionUseCase - RegisterDrone - RegisterDrone", err, map[string]any{
			"droneID": droneID,
		})
		return "", fmt.Errorf("DroneConnectionUseCase - RegisterDrone - RegisterDrone: %w", err)
	}

	uc.logger.Info("Drone registered", nil, map[string]any{
		"droneID":   droneID,
		"ipAddress": ipAddress,
	})

	return droneID, nil
}

func (uc *DroneConnectionUseCase) UnregisterDrone(ctx context.Context, droneID string) error {
	if err := uc.manager.UnregisterDrone(ctx, droneID); err != nil {
		uc.logger.Error("DroneConnectionUseCase - UnregisterDrone - UnregisterDrone", err, map[string]any{
			"droneID": droneID,
		})
		return fmt.Errorf("DroneConnectionUseCase - UnregisterDrone - UnregisterDrone: %w", err)
	}

	uc.logger.Info("Drone unregistered", nil, map[string]any{
		"droneID": droneID,
	})

	return nil
}

func (uc *DroneConnectionUseCase) GetDroneIDByIP(ctx context.Context, ipAddress string) (string, error) {
	droneID, err := uc.droneRepo.GetDroneIDByIP(ctx, ipAddress)
	if err != nil {
		uc.logger.Error("DroneConnectionUseCase - GetDroneIDByIP - GetDroneIDByIP", err, map[string]any{
			"ipAddress": ipAddress,
		})
		return "", fmt.Errorf("DroneConnectionUseCase - GetDroneIDByIP: %w", err)
	}
	return droneID, nil
}
