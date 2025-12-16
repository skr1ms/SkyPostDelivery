package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/logger"
)

type DroneTelemetryUseCase struct {
	droneRepo repo.DroneRepo
	logger    logger.Interface
}

func NewDroneTelemetryUseCase(
	droneRepo repo.DroneRepo,
	logger logger.Interface,
) *DroneTelemetryUseCase {
	return &DroneTelemetryUseCase{
		droneRepo: droneRepo,
		logger:    logger,
	}
}

func (uc *DroneTelemetryUseCase) parseHeartbeatPayload(droneID string, payload map[string]any) (*entity.DroneState, error) {
	var hb entity.HeartbeatPayload
	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &hb,
	})
	if err := decoder.Decode(payload); err != nil {
		return nil, fmt.Errorf("DroneTelemetryUseCase - parseHeartbeatPayload - Decode: %w", err)
	}

	if hb.Status == "" {
		hb.Status = "idle"
	}

	if hb.BatteryLevel == 0 {
		hb.BatteryLevel = 100.0
	}

	if hb.BatteryLevel < 0 || hb.BatteryLevel > 100 {
		return nil, fmt.Errorf("DroneTelemetryUseCase - parseHeartbeatPayload - ValidateBatteryLevel[%f]: %w", hb.BatteryLevel, entityError.ErrInvalidValue)
	}

	state := &entity.DroneState{
		DroneID:         droneID,
		Status:          entity.DroneStatus(hb.Status),
		BatteryLevel:    hb.BatteryLevel,
		CurrentPosition: hb.Position,
		Speed:           hb.Speed,
		LastUpdated:     time.Now(),
	}

	if hb.CurrentDeliveryID != "" {
		state.CurrentDeliveryID = &hb.CurrentDeliveryID
	}

	if hb.ErrorMessage != "" {
		state.ErrorMessage = &hb.ErrorMessage
	}

	return state, nil
}

func (uc *DroneTelemetryUseCase) parseStatusUpdatePayload(droneID string, payload map[string]any) (*entity.DroneState, error) {
	var su entity.StatusUpdatePayload
	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &su,
	})
	if err := decoder.Decode(payload); err != nil {
		return nil, fmt.Errorf("DroneTelemetryUseCase - parseStatusUpdatePayload - Decode: %w", err)
	}

	state := &entity.DroneState{
		DroneID:         droneID,
		Status:          entity.DroneStatus(su.Status),
		BatteryLevel:    su.BatteryLevel,
		CurrentPosition: su.Position,
		Speed:           su.Speed,
		LastUpdated:     time.Now(),
	}

	if su.CurrentDeliveryID != "" {
		state.CurrentDeliveryID = &su.CurrentDeliveryID
	}

	if su.ErrorMessage != "" {
		state.ErrorMessage = &su.ErrorMessage
	}

	return state, nil
}

func (uc *DroneTelemetryUseCase) ProcessHeartbeat(ctx context.Context, droneID string, payload map[string]any) error {
	state, err := uc.parseHeartbeatPayload(droneID, payload)
	if err != nil {
		if errors.Is(err, entityError.ErrInvalidPayload) || errors.Is(err, entityError.ErrInvalidValue) {
			return fmt.Errorf("DroneTelemetryUseCase - ProcessHeartbeat: %w", err)
		}
		uc.logger.Error("DroneTelemetryUseCase - ProcessHeartbeat - parseHeartbeatPayload", err, map[string]any{
			"droneID": droneID,
		})
		return fmt.Errorf("DroneTelemetryUseCase - ProcessHeartbeat - parseHeartbeatPayload: %w", err)
	}

	if state.BatteryLevel > 0 {
		if err := uc.droneRepo.UpdateDroneBattery(ctx, droneID, state.BatteryLevel); err != nil {
			uc.logger.Error("DroneTelemetryUseCase - ProcessHeartbeat - UpdateDroneBattery", err, map[string]any{
				"droneID":      droneID,
				"batteryLevel": state.BatteryLevel,
			})
			return fmt.Errorf("DroneTelemetryUseCase - ProcessHeartbeat - UpdateDroneBattery: %w", err)
		}
	}

	if err := uc.droneRepo.SaveDroneState(ctx, state); err != nil {
		uc.logger.Error("DroneTelemetryUseCase - ProcessHeartbeat - SaveDroneState", err, map[string]any{
			"droneID": droneID,
		})
		return fmt.Errorf("DroneTelemetryUseCase - ProcessHeartbeat - SaveDroneState: %w", err)
	}

	return nil
}

func (uc *DroneTelemetryUseCase) ProcessStatusUpdate(ctx context.Context, droneID string, payload map[string]any) error {
	state, err := uc.parseStatusUpdatePayload(droneID, payload)
	if err != nil {
		if errors.Is(err, entityError.ErrInvalidPayload) {
			return fmt.Errorf("DroneTelemetryUseCase - ProcessStatusUpdate: %w", err)
		}
		uc.logger.Error("DroneTelemetryUseCase - ProcessStatusUpdate - parseStatusUpdatePayload", err, map[string]any{
			"droneID": droneID,
		})
		return fmt.Errorf("DroneTelemetryUseCase - ProcessStatusUpdate - parseStatusUpdatePayload: %w", err)
	}

	if err := uc.droneRepo.SaveDroneState(ctx, state); err != nil {
		uc.logger.Error("DroneTelemetryUseCase - ProcessStatusUpdate - SaveDroneState", err, map[string]any{
			"droneID": droneID,
		})
		return fmt.Errorf("DroneTelemetryUseCase - ProcessStatusUpdate - SaveDroneState: %w", err)
	}

	return nil
}
