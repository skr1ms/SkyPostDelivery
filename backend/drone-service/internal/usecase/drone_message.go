package usecase

import (
	"context"
	"fmt"

	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/repo"
)

type VideoHandler interface {
	HandleVideoFrame(ctx context.Context, droneID string, frameData []byte, deliveryID string) error
}

type DroneMessageUseCase struct {
	droneRepo       repo.DroneRepo
	deliveryRepo    repo.DeliveryRepo
	droneManager    *DroneManagerUseCase
	deliveryUseCase *DeliveryUseCase
	droneNotifier   DroneNotifier
	videoHandler    VideoHandler
}

func NewDroneMessageUseCase(
	droneRepo repo.DroneRepo,
	deliveryRepo repo.DeliveryRepo,
	droneManager *DroneManagerUseCase,
	deliveryUseCase *DeliveryUseCase,
	droneNotifier DroneNotifier,
	videoHandler VideoHandler,
) *DroneMessageUseCase {
	return &DroneMessageUseCase{
		droneRepo:       droneRepo,
		deliveryRepo:    deliveryRepo,
		droneManager:    droneManager,
		deliveryUseCase: deliveryUseCase,
		droneNotifier:   droneNotifier,
		videoHandler:    videoHandler,
	}
}

func (uc *DroneMessageUseCase) RegisterDrone(ctx context.Context, ipAddress string) (string, error) {
	droneID, err := uc.droneRepo.GetDroneIDByIP(ctx, ipAddress)
	if err != nil || droneID == "" {
		return "", fmt.Errorf("no drone found with IP %s: %w", ipAddress, err)
	}

	if err := uc.droneManager.RegisterDrone(ctx, droneID); err != nil {
		return "", fmt.Errorf("failed to register drone: %w", err)
	}

	return droneID, nil
}

func (uc *DroneMessageUseCase) UnregisterDrone(ctx context.Context, droneID string) error {
	return uc.droneManager.UnregisterDrone(ctx, droneID)
}

func (uc *DroneMessageUseCase) ProcessHeartbeat(ctx context.Context, droneID string, payload map[string]interface{}) error {
	status, _ := payload["status"].(string)
	if status == "" {
		status = "idle"
	}

	batteryLevel, _ := payload["battery_level"].(float64)
	if batteryLevel == 0 {
		batteryLevel = 100.0
	}

	currentDeliveryID, _ := payload["current_delivery_id"].(string)
	errorMessage, _ := payload["error_message"].(string)

	position := entity.Position{}
	if pos, ok := payload["position"].(map[string]interface{}); ok {
		position.Latitude, _ = pos["latitude"].(float64)
		position.Longitude, _ = pos["longitude"].(float64)
		position.Altitude, _ = pos["altitude"].(float64)
	}

	speed, _ := payload["speed"].(float64)

	if batteryLevel > 0 {
		if err := uc.droneRepo.UpdateDroneBattery(ctx, droneID, batteryLevel); err != nil {
			return fmt.Errorf("failed to update battery: %w", err)
		}
	}

	state := &entity.DroneState{
		DroneID:         droneID,
		Status:          entity.DroneStatus(status),
		BatteryLevel:    batteryLevel,
		CurrentPosition: position,
		Speed:           speed,
		LastUpdated:     entity.DroneState{}.LastUpdated,
	}

	if currentDeliveryID != "" {
		state.CurrentDeliveryID = &currentDeliveryID
	}

	if errorMessage != "" {
		state.ErrorMessage = &errorMessage
	}

	return uc.droneRepo.SaveDroneState(ctx, state)
}

func (uc *DroneMessageUseCase) ProcessStatusUpdate(ctx context.Context, droneID string, payload map[string]interface{}) error {
	status, _ := payload["status"].(string)
	batteryLevel, _ := payload["battery_level"].(float64)
	currentDeliveryID, _ := payload["current_delivery_id"].(string)
	errorMessage, _ := payload["error_message"].(string)

	position := entity.Position{}
	if pos, ok := payload["position"].(map[string]interface{}); ok {
		position.Latitude, _ = pos["latitude"].(float64)
		position.Longitude, _ = pos["longitude"].(float64)
		position.Altitude, _ = pos["altitude"].(float64)
	}

	speed, _ := payload["speed"].(float64)

	state := &entity.DroneState{
		DroneID:         droneID,
		Status:          entity.DroneStatus(status),
		BatteryLevel:    batteryLevel,
		CurrentPosition: position,
		Speed:           speed,
	}

	if currentDeliveryID != "" {
		state.CurrentDeliveryID = &currentDeliveryID
	}

	if errorMessage != "" {
		state.ErrorMessage = &errorMessage
	}

	return uc.droneRepo.SaveDroneState(ctx, state)
}

func (uc *DroneMessageUseCase) ProcessDeliveryUpdate(ctx context.Context, droneID string, payload map[string]interface{}) error {
	droneStatus, _ := payload["drone_status"].(string)
	orderID, _ := payload["order_id"].(string)
	parcelAutomatID, _ := payload["parcel_automat_id"].(string)

	if droneStatus == "arrived_at_destination" {
		if orderID != "" && parcelAutomatID != "" && uc.deliveryUseCase != nil {
			_, err := uc.deliveryUseCase.HandleDroneArrived(ctx, droneID, orderID, parcelAutomatID)
			return err
		}
		return nil
	}

	deliveryID, ok := payload["delivery_id"].(string)
	if !ok || deliveryID == "" {
		return nil
	}

	switch droneStatus {
	case "arrived_at_locker":
		return uc.deliveryRepo.UpdateDeliveryStatus(ctx, deliveryID, entity.DeliveryStatusInProgress, nil)
	case "returning":
		return uc.droneManager.ReleaseDrone(ctx, droneID)
	}

	return nil
}

func (uc *DroneMessageUseCase) ProcessArrivedAtDestination(ctx context.Context, droneID string, payload map[string]interface{}) error {
	orderID, _ := payload["order_id"].(string)
	parcelAutomatID, _ := payload["parcel_automat_id"].(string)

	if orderID == "" || parcelAutomatID == "" {
		return nil
	}

	if uc.deliveryUseCase != nil {
		_, err := uc.deliveryUseCase.HandleDroneArrived(ctx, droneID, orderID, parcelAutomatID)
		return err
	}

	return nil
}

func (uc *DroneMessageUseCase) ProcessCargoDropped(ctx context.Context, payload map[string]interface{}) error {
	orderID, _ := payload["order_id"].(string)
	lockerCellID, _ := payload["locker_cell_id"].(string)

	if orderID == "" {
		return nil
	}

	if uc.deliveryUseCase != nil {
		_, err := uc.deliveryUseCase.HandleCargoDropped(ctx, orderID, lockerCellID)
		return err
	}

	return nil
}

func (uc *DroneMessageUseCase) ProcessVideoFrame(ctx context.Context, droneID string, frameData []byte, deliveryID string) error {
	if uc.videoHandler != nil {
		return uc.videoHandler.HandleVideoFrame(ctx, droneID, frameData, deliveryID)
	}
	return nil
}

func (uc *DroneMessageUseCase) GetDroneIDByIP(ctx context.Context, ipAddress string) (string, error) {
	return uc.droneRepo.GetDroneIDByIP(ctx, ipAddress)
}

func (uc *DroneMessageUseCase) SendDeliveryCommand(ctx context.Context, droneID string, command map[string]interface{}) error {
	if uc.droneNotifier != nil {
		return uc.droneNotifier.SendToDrone(ctx, droneID, command)
	}
	return fmt.Errorf("drone notifier not available")
}
