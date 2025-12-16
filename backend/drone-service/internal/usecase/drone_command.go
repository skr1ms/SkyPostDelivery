package usecase

import (
	"context"
	"fmt"

	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/logger"
)

type DroneCommandUseCase struct {
	notifier DroneNotifier
	logger   logger.Interface
}

func NewDroneCommandUseCase(
	notifier DroneNotifier,
	logger logger.Interface,
) *DroneCommandUseCase {
	return &DroneCommandUseCase{
		notifier: notifier,
		logger:   logger,
	}
}

func (uc *DroneCommandUseCase) SendCommand(ctx context.Context, droneID string, command map[string]any) error {
	if uc.notifier == nil {
		uc.logger.Warn("DroneCommandUseCase - SendCommand - drone notifier not available", nil, map[string]any{
			"droneID": droneID,
		})
		return fmt.Errorf("DroneCommandUseCase - SendCommand - drone notifier not available")
	}

	if err := uc.notifier.SendToDrone(ctx, droneID, command); err != nil {
		uc.logger.Error("DroneCommandUseCase - SendCommand - SendToDrone", err, map[string]any{
			"droneID": droneID,
		})
		return fmt.Errorf("DroneCommandUseCase - SendCommand - SendToDrone: %w", err)
	}

	return nil
}
